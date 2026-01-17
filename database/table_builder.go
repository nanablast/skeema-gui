package database

import (
	"fmt"
	"strings"
)

// ColumnDefinition holds column definition for table creation
type ColumnDefinition struct {
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Length        int     `json:"length,omitempty"`
	Nullable      bool    `json:"nullable"`
	DefaultValue  *string `json:"defaultValue,omitempty"`
	AutoIncrement bool    `json:"autoIncrement"`
	PrimaryKey    bool    `json:"primaryKey"`
	Unique        bool    `json:"unique"`
	Comment       string  `json:"comment,omitempty"`
}

// IndexDefinition holds index definition
type IndexDefinition struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
}

// TableDefinition holds complete table definition
type TableDefinition struct {
	Name       string             `json:"name"`
	Columns    []ColumnDefinition `json:"columns"`
	Indexes    []IndexDefinition  `json:"indexes"`
	Engine     string             `json:"engine"`
	Charset    string             `json:"charset"`
	Collation  string             `json:"collation"`
	Comment    string             `json:"comment,omitempty"`
}

// CommonDataTypes returns commonly used MySQL data types
func GetCommonDataTypes() []string {
	return []string{
		"INT",
		"BIGINT",
		"TINYINT",
		"SMALLINT",
		"DECIMAL",
		"FLOAT",
		"DOUBLE",
		"VARCHAR",
		"CHAR",
		"TEXT",
		"MEDIUMTEXT",
		"LONGTEXT",
		"DATE",
		"DATETIME",
		"TIMESTAMP",
		"TIME",
		"YEAR",
		"BOOLEAN",
		"JSON",
		"BLOB",
		"MEDIUMBLOB",
		"LONGBLOB",
		"ENUM",
		"SET",
	}
}

// GenerateCreateTableSQL generates CREATE TABLE SQL from definition
func GenerateCreateTableSQL(def TableDefinition) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("CREATE TABLE `%s` (\n", def.Name))

	var columnDefs []string
	var primaryKeys []string

	for _, col := range def.Columns {
		colDef := buildColumnDefinition(col)
		columnDefs = append(columnDefs, "  "+colDef)

		if col.PrimaryKey {
			primaryKeys = append(primaryKeys, fmt.Sprintf("`%s`", col.Name))
		}
	}

	// Add primary key constraint
	if len(primaryKeys) > 0 {
		columnDefs = append(columnDefs, fmt.Sprintf("  PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}

	// Add indexes
	for _, idx := range def.Indexes {
		var cols []string
		for _, c := range idx.Columns {
			cols = append(cols, fmt.Sprintf("`%s`", c))
		}
		if idx.Unique {
			columnDefs = append(columnDefs, fmt.Sprintf("  UNIQUE KEY `%s` (%s)", idx.Name, strings.Join(cols, ", ")))
		} else {
			columnDefs = append(columnDefs, fmt.Sprintf("  KEY `%s` (%s)", idx.Name, strings.Join(cols, ", ")))
		}
	}

	sb.WriteString(strings.Join(columnDefs, ",\n"))
	sb.WriteString("\n)")

	// Table options
	if def.Engine != "" {
		sb.WriteString(fmt.Sprintf(" ENGINE=%s", def.Engine))
	} else {
		sb.WriteString(" ENGINE=InnoDB")
	}

	if def.Charset != "" {
		sb.WriteString(fmt.Sprintf(" DEFAULT CHARSET=%s", def.Charset))
	} else {
		sb.WriteString(" DEFAULT CHARSET=utf8mb4")
	}

	if def.Collation != "" {
		sb.WriteString(fmt.Sprintf(" COLLATE=%s", def.Collation))
	} else {
		sb.WriteString(" COLLATE=utf8mb4_unicode_ci")
	}

	if def.Comment != "" {
		sb.WriteString(fmt.Sprintf(" COMMENT='%s'", escapeString(def.Comment)))
	}

	sb.WriteString(";")

	return sb.String()
}

func buildColumnDefinition(col ColumnDefinition) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("`%s`", col.Name))

	// Type with length
	if col.Length > 0 && needsLength(col.Type) {
		parts = append(parts, fmt.Sprintf("%s(%d)", col.Type, col.Length))
	} else {
		parts = append(parts, col.Type)
	}

	// Nullable
	if !col.Nullable {
		parts = append(parts, "NOT NULL")
	} else {
		parts = append(parts, "NULL")
	}

	// Default value
	if col.DefaultValue != nil {
		val := *col.DefaultValue
		if val == "CURRENT_TIMESTAMP" || val == "NULL" {
			parts = append(parts, fmt.Sprintf("DEFAULT %s", val))
		} else {
			parts = append(parts, fmt.Sprintf("DEFAULT '%s'", escapeString(val)))
		}
	}

	// Auto increment
	if col.AutoIncrement {
		parts = append(parts, "AUTO_INCREMENT")
	}

	// Unique (if not primary key)
	if col.Unique && !col.PrimaryKey {
		parts = append(parts, "UNIQUE")
	}

	// Comment
	if col.Comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", escapeString(col.Comment)))
	}

	return strings.Join(parts, " ")
}

func needsLength(dataType string) bool {
	upper := strings.ToUpper(dataType)
	return upper == "VARCHAR" || upper == "CHAR" || upper == "DECIMAL" ||
		upper == "FLOAT" || upper == "DOUBLE"
}

func escapeString(s string) string {
	s = strings.ReplaceAll(s, "'", "''")
	s = strings.ReplaceAll(s, "\\", "\\\\")
	return s
}

// CreateTable creates a table in the database
func CreateTable(config ConnectionConfig, def TableDefinition) error {
	db, err := Connect(config)
	if err != nil {
		return err
	}
	defer db.Close()

	sql := GenerateCreateTableSQL(def)
	_, err = db.Exec(sql)
	return err
}

// GetTableEngines returns available storage engines
func GetTableEngines() []string {
	return []string{"InnoDB", "MyISAM", "MEMORY", "CSV", "ARCHIVE"}
}

// GetCharsets returns common charsets
func GetCharsets() []string {
	return []string{"utf8mb4", "utf8", "latin1", "ascii", "gbk", "gb2312"}
}
