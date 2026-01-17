package database

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// ConnectionConfig holds database connection parameters
type ConnectionConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// TableInfo holds table structure information
type TableInfo struct {
	Name       string       `json:"name"`
	CreateSQL  string       `json:"createSql"`
	Columns    []ColumnInfo `json:"columns"`
	Indexes    []IndexInfo  `json:"indexes"`
}

// ColumnInfo holds column details
type ColumnInfo struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Nullable     string  `json:"nullable"`
	Key          string  `json:"key"`
	Default      *string `json:"default"`
	Extra        string  `json:"extra"`
	Position     int     `json:"position"`
}

// IndexInfo holds index details
type IndexInfo struct {
	Name      string `json:"name"`
	NonUnique int    `json:"nonUnique"`
	Column    string `json:"column"`
	SeqInIdx  int    `json:"seqInIndex"`
}

// SchemaInfo holds complete database schema
type SchemaInfo struct {
	Database string               `json:"database"`
	Tables   map[string]TableInfo `json:"tables"`
}

// DiffResult holds comparison result
type DiffResult struct {
	Type      string `json:"type"`      // "added", "removed", "modified"
	TableName string `json:"tableName"`
	Detail    string `json:"detail"`
	SQL       string `json:"sql"`
}

// Connect creates a database connection
func Connect(config ConnectionConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
		config.User, config.Password, config.Host, config.Port, config.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// TestConnection tests if the connection works
func TestConnection(config ConnectionConfig) error {
	db, err := Connect(config)
	if err != nil {
		return err
	}
	defer db.Close()
	return nil
}

// GetDatabases returns list of databases
func GetDatabases(config ConnectionConfig) ([]string, error) {
	cfg := config
	cfg.Database = ""

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		cfg.User, cfg.Password, cfg.Host, cfg.Port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		// Skip system databases
		if name != "information_schema" && name != "mysql" &&
		   name != "performance_schema" && name != "sys" {
			databases = append(databases, name)
		}
	}

	return databases, nil
}

// GetSchema retrieves complete schema information
func GetSchema(config ConnectionConfig) (*SchemaInfo, error) {
	db, err := Connect(config)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	schema := &SchemaInfo{
		Database: config.Database,
		Tables:   make(map[string]TableInfo),
	}

	// Get all tables
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tableNames = append(tableNames, name)
	}

	// Get details for each table
	for _, tableName := range tableNames {
		tableInfo, err := getTableInfo(db, tableName)
		if err != nil {
			return nil, err
		}
		schema.Tables[tableName] = *tableInfo
	}

	return schema, nil
}

func getTableInfo(db *sql.DB, tableName string) (*TableInfo, error) {
	info := &TableInfo{
		Name: tableName,
	}

	// Get CREATE TABLE statement
	var tbl, createSQL string
	err := db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName)).Scan(&tbl, &createSQL)
	if err != nil {
		return nil, err
	}
	info.CreateSQL = createSQL

	// Get columns
	colRows, err := db.Query(`
		SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY, COLUMN_DEFAULT, EXTRA, ORDINAL_POSITION
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`, tableName)
	if err != nil {
		return nil, err
	}
	defer colRows.Close()

	for colRows.Next() {
		var col ColumnInfo
		if err := colRows.Scan(&col.Name, &col.Type, &col.Nullable, &col.Key, &col.Default, &col.Extra, &col.Position); err != nil {
			return nil, err
		}
		info.Columns = append(info.Columns, col)
	}

	// Get indexes
	idxRows, err := db.Query(fmt.Sprintf("SHOW INDEX FROM `%s`", tableName))
	if err != nil {
		return nil, err
	}
	defer idxRows.Close()

	cols, _ := idxRows.Columns()
	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for idxRows.Next() {
		if err := idxRows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		idx := IndexInfo{}
		for i, col := range cols {
			val := values[i]
			switch col {
			case "Key_name":
				if v, ok := val.([]byte); ok {
					idx.Name = string(v)
				}
			case "Non_unique":
				if v, ok := val.(int64); ok {
					idx.NonUnique = int(v)
				}
			case "Column_name":
				if v, ok := val.([]byte); ok {
					idx.Column = string(v)
				}
			case "Seq_in_index":
				if v, ok := val.(int64); ok {
					idx.SeqInIdx = int(v)
				}
			}
		}
		info.Indexes = append(info.Indexes, idx)
	}

	return info, nil
}

// CompareSchemas compares two schemas and returns differences
func CompareSchemas(source, target *SchemaInfo) []DiffResult {
	var results []DiffResult

	// Find tables only in source (need to add to target)
	for tableName, sourceTable := range source.Tables {
		if _, exists := target.Tables[tableName]; !exists {
			results = append(results, DiffResult{
				Type:      "added",
				TableName: tableName,
				Detail:    "Table exists in source but not in target",
				SQL:       sourceTable.CreateSQL + ";",
			})
		}
	}

	// Find tables only in target (need to remove from target)
	for tableName := range target.Tables {
		if _, exists := source.Tables[tableName]; !exists {
			results = append(results, DiffResult{
				Type:      "removed",
				TableName: tableName,
				Detail:    "Table exists in target but not in source",
				SQL:       fmt.Sprintf("DROP TABLE `%s`;", tableName),
			})
		}
	}

	// Compare existing tables
	for tableName, sourceTable := range source.Tables {
		if targetTable, exists := target.Tables[tableName]; exists {
			tableDiffs := compareTableStructure(tableName, sourceTable, targetTable)
			results = append(results, tableDiffs...)
		}
	}

	// Sort results by type and table name
	sort.Slice(results, func(i, j int) bool {
		if results[i].Type != results[j].Type {
			order := map[string]int{"added": 0, "modified": 1, "removed": 2}
			return order[results[i].Type] < order[results[j].Type]
		}
		return results[i].TableName < results[j].TableName
	})

	return results
}

func compareTableStructure(tableName string, source, target TableInfo) []DiffResult {
	var results []DiffResult

	sourceColMap := make(map[string]ColumnInfo)
	targetColMap := make(map[string]ColumnInfo)

	for _, col := range source.Columns {
		sourceColMap[col.Name] = col
	}
	for _, col := range target.Columns {
		targetColMap[col.Name] = col
	}

	// Find added columns
	for colName, sourceCol := range sourceColMap {
		if _, exists := targetColMap[colName]; !exists {
			afterClause := ""
			if sourceCol.Position > 1 {
				for _, c := range source.Columns {
					if c.Position == sourceCol.Position-1 {
						afterClause = fmt.Sprintf(" AFTER `%s`", c.Name)
						break
					}
				}
			} else {
				afterClause = " FIRST"
			}

			results = append(results, DiffResult{
				Type:      "modified",
				TableName: tableName,
				Detail:    fmt.Sprintf("Add column: %s", colName),
				SQL:       fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s%s;", tableName, colName, buildColumnDef(sourceCol), afterClause),
			})
		}
	}

	// Find removed columns
	for colName := range targetColMap {
		if _, exists := sourceColMap[colName]; !exists {
			results = append(results, DiffResult{
				Type:      "modified",
				TableName: tableName,
				Detail:    fmt.Sprintf("Drop column: %s", colName),
				SQL:       fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`;", tableName, colName),
			})
		}
	}

	// Find modified columns
	for colName, sourceCol := range sourceColMap {
		if targetCol, exists := targetColMap[colName]; exists {
			if !columnsEqual(sourceCol, targetCol) {
				results = append(results, DiffResult{
					Type:      "modified",
					TableName: tableName,
					Detail:    fmt.Sprintf("Modify column: %s (%s -> %s)", colName, targetCol.Type, sourceCol.Type),
					SQL:       fmt.Sprintf("ALTER TABLE `%s` MODIFY COLUMN `%s` %s;", tableName, colName, buildColumnDef(sourceCol)),
				})
			}
		}
	}

	// Compare indexes
	sourceIdxMap := buildIndexMap(source.Indexes)
	targetIdxMap := buildIndexMap(target.Indexes)

	for idxName, sourceCols := range sourceIdxMap {
		if idxName == "PRIMARY" {
			continue // Skip primary key for now
		}
		if targetCols, exists := targetIdxMap[idxName]; !exists {
			results = append(results, DiffResult{
				Type:      "modified",
				TableName: tableName,
				Detail:    fmt.Sprintf("Add index: %s", idxName),
				SQL:       fmt.Sprintf("ALTER TABLE `%s` ADD INDEX `%s` (%s);", tableName, idxName, strings.Join(sourceCols, ", ")),
			})
		} else if !stringSlicesEqual(sourceCols, targetCols) {
			results = append(results, DiffResult{
				Type:      "modified",
				TableName: tableName,
				Detail:    fmt.Sprintf("Recreate index: %s", idxName),
				SQL:       fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`, ADD INDEX `%s` (%s);", tableName, idxName, idxName, strings.Join(sourceCols, ", ")),
			})
		}
	}

	for idxName := range targetIdxMap {
		if idxName == "PRIMARY" {
			continue
		}
		if _, exists := sourceIdxMap[idxName]; !exists {
			results = append(results, DiffResult{
				Type:      "modified",
				TableName: tableName,
				Detail:    fmt.Sprintf("Drop index: %s", idxName),
				SQL:       fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`;", tableName, idxName),
			})
		}
	}

	return results
}

func buildColumnDef(col ColumnInfo) string {
	def := col.Type
	if col.Nullable == "NO" {
		def += " NOT NULL"
	}
	if col.Default != nil {
		def += fmt.Sprintf(" DEFAULT '%s'", *col.Default)
	}
	if col.Extra != "" {
		def += " " + col.Extra
	}
	return def
}

func columnsEqual(a, b ColumnInfo) bool {
	return a.Type == b.Type && a.Nullable == b.Nullable &&
		   a.Extra == b.Extra && defaultsEqual(a.Default, b.Default)
}

func defaultsEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func buildIndexMap(indexes []IndexInfo) map[string][]string {
	result := make(map[string][]string)
	for _, idx := range indexes {
		result[idx.Name] = append(result[idx.Name], fmt.Sprintf("`%s`", idx.Column))
	}
	return result
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// CreateDatabase creates a new database
func CreateDatabase(config ConnectionConfig, dbName, charset, collation string) error {
	// Connect without specifying a database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		config.User, config.Password, config.Host, config.Port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// Build CREATE DATABASE statement
	sql := fmt.Sprintf("CREATE DATABASE `%s`", dbName)
	if charset != "" {
		sql += fmt.Sprintf(" CHARACTER SET %s", charset)
	}
	if collation != "" {
		sql += fmt.Sprintf(" COLLATE %s", collation)
	}

	_, err = db.Exec(sql)
	return err
}

// DropDatabase drops a database
func DropDatabase(config ConnectionConfig, dbName string) error {
	// Connect without specifying a database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		config.User, config.Password, config.Host, config.Port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE `%s`", dbName))
	return err
}
