package sqlite

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mstgnz/sqlmapper"
)

type SQLite struct {
	schema *sqlmapper.Schema
}

// NewSQLite creates a new SQLite parser instance
func NewSQLite() *SQLite {
	return &SQLite{
		schema: &sqlmapper.Schema{},
	}
}

// Parse parses SQLite dump content
func (s *SQLite) Parse(content string) (*sqlmapper.Schema, error) {
	if content == "" {
		return nil, errors.New("empty content")
	}

	// TODO: SQLite dump parsing operations will be implemented here
	return s.schema, nil
}

// Generate generates SQLite dump from schema
func (s *SQLite) Generate(schema *sqlmapper.Schema) (string, error) {
	if schema == nil {
		return "", errors.New("empty schema")
	}

	var result strings.Builder

	for _, table := range schema.Tables {
		result.WriteString("CREATE TABLE ")
		result.WriteString(table.Name)
		result.WriteString(" (\n")

		for i, col := range table.Columns {
			result.WriteString("    ")
			result.WriteString(col.Name)
			result.WriteString(" ")

			if col.IsPrimaryKey && col.DataType == "INTEGER" {
				result.WriteString("INTEGER PRIMARY KEY")
			} else {
				result.WriteString(col.DataType)
				if col.DataType != "TEXT" && col.Length > 0 {
					result.WriteString(fmt.Sprintf("(%d", col.Length))
					if col.Scale > 0 {
						result.WriteString(fmt.Sprintf(",%d", col.Scale))
					}
					result.WriteString(")")
				}

				if !col.IsNullable {
					result.WriteString(" NOT NULL")
				}

				if col.IsUnique {
					result.WriteString(" UNIQUE")
				}
			}

			if i < len(table.Columns)-1 {
				result.WriteString(",")
			}
			result.WriteString("\n")
		}

		result.WriteString(");\n")

		// Add indexes
		for _, idx := range table.Indexes {
			if idx.IsUnique {
				result.WriteString("CREATE UNIQUE INDEX ")
			} else {
				result.WriteString("CREATE INDEX ")
			}
			result.WriteString(idx.Name)
			result.WriteString(" ON ")
			result.WriteString(table.Name)
			result.WriteString("(")
			result.WriteString(strings.Join(idx.Columns, ", "))
			result.WriteString(");\n")
		}
	}

	return result.String(), nil
}

// getDataType maps common data types to SQLite data types
func (s *SQLite) getDataType(dataType string) string {
	dataType = strings.ToUpper(dataType)
	if strings.Contains(dataType, "(") {
		parts := strings.Split(dataType, "(")
		dataType = parts[0]
	}

	switch dataType {
	case "INT", "INTEGER", "BIGINT", "SMALLINT", "TINYINT":
		return "INTEGER"
	case "VARCHAR", "CHAR", "TEXT", "NVARCHAR", "NCHAR":
		return "TEXT"
	case "DECIMAL", "NUMERIC", "FLOAT", "DOUBLE", "REAL":
		return "REAL"
	case "BLOB", "BINARY", "VARBINARY":
		return "BLOB"
	default:
		return dataType
	}
}
