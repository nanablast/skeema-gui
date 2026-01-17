package database

import (
	"database/sql"
	"fmt"
	"strings"
)

// DataSyncConfig holds sync configuration
type DataSyncConfig struct {
	SourceConfig ConnectionConfig `json:"sourceConfig"`
	TargetConfig ConnectionConfig `json:"targetConfig"`
	TableName    string           `json:"tableName"`
	SyncInsert   bool             `json:"syncInsert"`
	SyncUpdate   bool             `json:"syncUpdate"`
	SyncDelete   bool             `json:"syncDelete"`
}

// TableDataInfo holds table data comparison info
type TableDataInfo struct {
	TableName    string   `json:"tableName"`
	PrimaryKeys  []string `json:"primaryKeys"`
	Columns      []string `json:"columns"`
	SourceCount  int      `json:"sourceCount"`
	TargetCount  int      `json:"targetCount"`
	InsertCount  int      `json:"insertCount"`
	UpdateCount  int      `json:"updateCount"`
	DeleteCount  int      `json:"deleteCount"`
}

// DataDiffResult holds data difference details
type DataDiffResult struct {
	Type       string                 `json:"type"` // "insert", "update", "delete"
	TableName  string                 `json:"tableName"`
	PrimaryKey map[string]interface{} `json:"primaryKey"`
	OldValues  map[string]interface{} `json:"oldValues,omitempty"`
	NewValues  map[string]interface{} `json:"newValues,omitempty"`
	SQL        string                 `json:"sql"`
}

// GetTablesForSync returns list of tables available for data sync
func GetTablesForSync(config ConnectionConfig) ([]TableDataInfo, error) {
	db, err := Connect(config)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableDataInfo
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}

		info := TableDataInfo{TableName: tableName}

		// Get primary keys
		pkRows, err := db.Query(`
			SELECT COLUMN_NAME
			FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
			WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
			AND CONSTRAINT_NAME = 'PRIMARY'
			ORDER BY ORDINAL_POSITION`, tableName)
		if err != nil {
			return nil, err
		}

		for pkRows.Next() {
			var pk string
			if err := pkRows.Scan(&pk); err != nil {
				pkRows.Close()
				return nil, err
			}
			info.PrimaryKeys = append(info.PrimaryKeys, pk)
		}
		pkRows.Close()

		// Get columns
		colRows, err := db.Query(`
			SELECT COLUMN_NAME
			FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
			ORDER BY ORDINAL_POSITION`, tableName)
		if err != nil {
			return nil, err
		}

		for colRows.Next() {
			var col string
			if err := colRows.Scan(&col); err != nil {
				colRows.Close()
				return nil, err
			}
			info.Columns = append(info.Columns, col)
		}
		colRows.Close()

		// Get row count
		var count int
		err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)).Scan(&count)
		if err != nil {
			return nil, err
		}
		info.SourceCount = count

		tables = append(tables, info)
	}

	return tables, nil
}

// CompareTableData compares data between source and target tables
func CompareTableData(sourceConfig, targetConfig ConnectionConfig, tableName string) ([]DataDiffResult, error) {
	sourceDB, err := Connect(sourceConfig)
	if err != nil {
		return nil, fmt.Errorf("source connection failed: %v", err)
	}
	defer sourceDB.Close()

	targetDB, err := Connect(targetConfig)
	if err != nil {
		return nil, fmt.Errorf("target connection failed: %v", err)
	}
	defer targetDB.Close()

	// Get primary keys
	primaryKeys, err := getPrimaryKeys(sourceDB, tableName)
	if err != nil {
		return nil, err
	}
	if len(primaryKeys) == 0 {
		return nil, fmt.Errorf("table %s has no primary key", tableName)
	}

	// Get columns
	columns, err := getColumns(sourceDB, tableName)
	if err != nil {
		return nil, err
	}

	var results []DataDiffResult

	// Get source data
	sourceData, err := getTableData(sourceDB, tableName, columns, primaryKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to get source data: %v", err)
	}

	// Get target data
	targetData, err := getTableData(targetDB, tableName, columns, primaryKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to get target data: %v", err)
	}

	// Find inserts and updates
	for pkKey, sourceRow := range sourceData {
		if targetRow, exists := targetData[pkKey]; exists {
			// Check for updates
			if !rowsEqual(sourceRow, targetRow) {
				pk := extractPrimaryKey(sourceRow, primaryKeys)
				results = append(results, DataDiffResult{
					Type:       "update",
					TableName:  tableName,
					PrimaryKey: pk,
					OldValues:  targetRow,
					NewValues:  sourceRow,
					SQL:        generateUpdateSQL(tableName, sourceRow, primaryKeys),
				})
			}
		} else {
			// Insert
			pk := extractPrimaryKey(sourceRow, primaryKeys)
			results = append(results, DataDiffResult{
				Type:       "insert",
				TableName:  tableName,
				PrimaryKey: pk,
				NewValues:  sourceRow,
				SQL:        generateInsertSQL(tableName, sourceRow, columns),
			})
		}
	}

	// Find deletes
	for pkKey, targetRow := range targetData {
		if _, exists := sourceData[pkKey]; !exists {
			pk := extractPrimaryKey(targetRow, primaryKeys)
			results = append(results, DataDiffResult{
				Type:       "delete",
				TableName:  tableName,
				PrimaryKey: pk,
				OldValues:  targetRow,
				SQL:        generateDeleteSQL(tableName, primaryKeys, pk),
			})
		}
	}

	return results, nil
}

// GetDataSyncSummary returns a summary of data differences for a table
func GetDataSyncSummary(sourceConfig, targetConfig ConnectionConfig, tableName string) (*TableDataInfo, error) {
	diffs, err := CompareTableData(sourceConfig, targetConfig, tableName)
	if err != nil {
		return nil, err
	}

	sourceDB, err := Connect(sourceConfig)
	if err != nil {
		return nil, err
	}
	defer sourceDB.Close()

	targetDB, err := Connect(targetConfig)
	if err != nil {
		return nil, err
	}
	defer targetDB.Close()

	info := &TableDataInfo{TableName: tableName}

	// Get primary keys
	info.PrimaryKeys, _ = getPrimaryKeys(sourceDB, tableName)
	info.Columns, _ = getColumns(sourceDB, tableName)

	// Get counts
	sourceDB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)).Scan(&info.SourceCount)
	targetDB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)).Scan(&info.TargetCount)

	for _, diff := range diffs {
		switch diff.Type {
		case "insert":
			info.InsertCount++
		case "update":
			info.UpdateCount++
		case "delete":
			info.DeleteCount++
		}
	}

	return info, nil
}

func getPrimaryKeys(db *sql.DB, tableName string) ([]string, error) {
	rows, err := db.Query(`
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = ?
		AND CONSTRAINT_NAME = 'PRIMARY'
		ORDER BY ORDINAL_POSITION`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pks []string
	for rows.Next() {
		var pk string
		if err := rows.Scan(&pk); err != nil {
			return nil, err
		}
		pks = append(pks, pk)
	}
	return pks, nil
}

func getColumns(db *sql.DB, tableName string) ([]string, error) {
	rows, err := db.Query(`
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, err
		}
		cols = append(cols, col)
	}
	return cols, nil
}

func getTableData(db *sql.DB, tableName string, columns, primaryKeys []string) (map[string]map[string]interface{}, error) {
	quotedCols := make([]string, len(columns))
	for i, col := range columns {
		quotedCols[i] = fmt.Sprintf("`%s`", col)
	}

	query := fmt.Sprintf("SELECT %s FROM `%s`", strings.Join(quotedCols, ", "), tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make(map[string]map[string]interface{})

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		var pkParts []string
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}

		// Build primary key string
		for _, pk := range primaryKeys {
			pkParts = append(pkParts, fmt.Sprintf("%v", row[pk]))
		}
		pkKey := strings.Join(pkParts, "|")
		data[pkKey] = row
	}

	return data, nil
}

func rowsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if fmt.Sprintf("%v", v) != fmt.Sprintf("%v", b[k]) {
			return false
		}
	}
	return true
}

func extractPrimaryKey(row map[string]interface{}, primaryKeys []string) map[string]interface{} {
	pk := make(map[string]interface{})
	for _, key := range primaryKeys {
		pk[key] = row[key]
	}
	return pk
}

func generateInsertSQL(tableName string, row map[string]interface{}, columns []string) string {
	var cols []string
	var vals []string

	for _, col := range columns {
		if val, ok := row[col]; ok {
			cols = append(cols, fmt.Sprintf("`%s`", col))
			vals = append(vals, escapeValue(val))
		}
	}

	return fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s);",
		tableName,
		strings.Join(cols, ", "),
		strings.Join(vals, ", "))
}

func generateUpdateSQL(tableName string, row map[string]interface{}, primaryKeys []string) string {
	var sets []string
	var wheres []string

	for col, val := range row {
		isPK := false
		for _, pk := range primaryKeys {
			if col == pk {
				isPK = true
				break
			}
		}
		if !isPK {
			sets = append(sets, fmt.Sprintf("`%s` = %s", col, escapeValue(val)))
		}
	}

	for _, pk := range primaryKeys {
		wheres = append(wheres, fmt.Sprintf("`%s` = %s", pk, escapeValue(row[pk])))
	}

	return fmt.Sprintf("UPDATE `%s` SET %s WHERE %s;",
		tableName,
		strings.Join(sets, ", "),
		strings.Join(wheres, " AND "))
}

func generateDeleteSQL(tableName string, primaryKeys []string, pk map[string]interface{}) string {
	var wheres []string
	for _, key := range primaryKeys {
		wheres = append(wheres, fmt.Sprintf("`%s` = %s", key, escapeValue(pk[key])))
	}
	return fmt.Sprintf("DELETE FROM `%s` WHERE %s;", tableName, strings.Join(wheres, " AND "))
}

func escapeValue(val interface{}) string {
	if val == nil {
		return "NULL"
	}
	switch v := val.(type) {
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "1"
		}
		return "0"
	default:
		s := fmt.Sprintf("%v", v)
		s = strings.ReplaceAll(s, "'", "''")
		return fmt.Sprintf("'%s'", s)
	}
}
