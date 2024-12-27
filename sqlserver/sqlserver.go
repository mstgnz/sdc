package sqlserver

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mstgnz/sqlporter"
)

type SQLServer struct {
	schema *sqlporter.Schema
}

// NewSQLServer creates a new SQLServer parser instance
func NewSQLServer() *SQLServer {
	return &SQLServer{
		schema: &sqlporter.Schema{},
	}
}

// Parse parses SQLServer dump content
func (s *SQLServer) Parse(content string) (*sqlporter.Schema, error) {
	if content == "" {
		return nil, errors.New("empty content")
	}

	// TODO: SQLServer dump parsing operations will be implemented here
	return s.schema, nil
}

// Generate generates SQLServer dump from schema
func (s *SQLServer) Generate(schema *sqlporter.Schema) (string, error) {
	if schema == nil {
		return "", errors.New("empty schema")
	}

	var result strings.Builder

	// Generate table creation
	for i, table := range schema.Tables {
		result.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", table.Name))

		// Generate columns
		for j, column := range table.Columns {
			result.WriteString(fmt.Sprintf("    %s %s", column.Name, s.getDataType(column.DataType, column.Length)))

			if column.IsPrimaryKey {
				result.WriteString(" PRIMARY KEY")
			} else if !column.IsNullable {
				result.WriteString(" NOT NULL")
			}

			if column.IsUnique && !column.IsPrimaryKey {
				result.WriteString(" UNIQUE")
			}

			if j < len(table.Columns)-1 {
				result.WriteString(",")
			}
			result.WriteString("\n")
		}

		result.WriteString(");\n")

		// Generate indexes
		for _, index := range table.Indexes {
			if index.IsUnique {
				result.WriteString(fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s(%s);\n",
					index.Name, table.Name, strings.Join(index.Columns, ", ")))
			} else {
				result.WriteString(fmt.Sprintf("CREATE INDEX %s ON %s(%s);\n",
					index.Name, table.Name, strings.Join(index.Columns, ", ")))
			}
		}

		if i < len(schema.Tables)-1 {
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

// getDataType maps common data types to SQL Server data types
func (s *SQLServer) getDataType(dataType string, length int) string {
	dataType = strings.ToUpper(dataType)
	if strings.Contains(dataType, "(") {
		parts := strings.Split(dataType, "(")
		dataType = parts[0]
	}

	switch dataType {
	case "INT", "INTEGER":
		return "INT"
	case "VARCHAR", "CHAR":
		if length > 0 {
			return fmt.Sprintf("NVARCHAR(%d)", length)
		}
		return "NVARCHAR(MAX)"
	case "DECIMAL", "NUMERIC":
		if length > 0 {
			return fmt.Sprintf("DECIMAL(%d,2)", length)
		}
		return "DECIMAL(18,2)"
	case "FLOAT", "DOUBLE":
		return "FLOAT"
	case "BLOB", "BINARY":
		return "VARBINARY(MAX)"
	default:
		if length > 0 {
			return fmt.Sprintf("%s(%d)", dataType, length)
		}
		return dataType
	}
}
