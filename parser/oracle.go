package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mstgnz/sdc"
)

var (
	// Pre-compiled regular expressions for better performance
	oracleColumnRegex  = regexp.MustCompile(`(?i)^\s*"?([^"\s(]+)"?\s+([^(,\s]+)(?:\s*\(([^)]+)\))?\s*(.*)$`)
	oracleDefaultRegex = regexp.MustCompile(`(?i)DEFAULT\s+([^,\s]+|'[^']*'|"[^"]*")`)
)

// OracleParser implements the parser for Oracle database
type OracleParser struct{}

// Parse converts Oracle dump to Entity structure
func (p *OracleParser) Parse(sql string) (*sdc.Entity, error) {
	entity := &sdc.Entity{}

	// Split SQL statements
	statements := strings.Split(sql, ";")

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Process CREATE TABLE statements
		if strings.HasPrefix(strings.ToUpper(stmt), "CREATE TABLE") {
			table, err := p.parseCreateTable(stmt)
			if err != nil {
				return nil, err
			}
			entity.Tables = append(entity.Tables, table)
		}
	}

	return entity, nil
}

// Convert transforms Entity structure to Oracle format
func (p *OracleParser) Convert(entity *sdc.Entity) (string, error) {
	var result strings.Builder

	for _, table := range entity.Tables {
		// CREATE TABLE statement
		result.WriteString(fmt.Sprintf("CREATE TABLE \"%s\" (\n", strings.ToUpper(table.Name)))

		// Columns
		for i, column := range table.Columns {
			result.WriteString(fmt.Sprintf("    \"%s\" %s", strings.ToUpper(column.Name), p.convertDataType(column.DataType)))

			if !column.IsNullable {
				result.WriteString(" NOT NULL")
			}

			if column.Default != "" {
				result.WriteString(fmt.Sprintf(" DEFAULT %s", column.Default))
			}

			if i < len(table.Columns)-1 {
				result.WriteString(",\n")
			}
		}

		result.WriteString("\n);\n\n")
	}

	return result.String(), nil
}

// convertDataType converts data type to Oracle format
func (p *OracleParser) convertDataType(dataType *sdc.DataType) string {
	if dataType == nil {
		return "VARCHAR2(4000)"
	}

	switch strings.ToUpper(dataType.Name) {
	case "VARCHAR", "NVARCHAR":
		if dataType.Length > 0 {
			return fmt.Sprintf("VARCHAR2(%d)", dataType.Length)
		}
		return "VARCHAR2(4000)"
	case "CHAR", "NCHAR":
		if dataType.Length > 0 {
			return fmt.Sprintf("CHAR(%d)", dataType.Length)
		}
		return "CHAR(1)"
	case "INT", "INTEGER":
		return "NUMBER(10)"
	case "BIGINT":
		return "NUMBER(19)"
	case "SMALLINT":
		return "NUMBER(5)"
	case "DECIMAL", "NUMERIC":
		if dataType.Precision > 0 {
			if dataType.Scale > 0 {
				return fmt.Sprintf("NUMBER(%d,%d)", dataType.Precision, dataType.Scale)
			}
			return fmt.Sprintf("NUMBER(%d)", dataType.Precision)
		}
		return "NUMBER"
	case "FLOAT", "REAL":
		return "FLOAT"
	case "DOUBLE":
		return "BINARY_DOUBLE"
	case "BOOLEAN", "BIT":
		return "NUMBER(1)"
	case "DATE":
		return "DATE"
	case "TIME":
		return "TIMESTAMP"
	case "TIMESTAMP":
		return "TIMESTAMP"
	case "TEXT", "NTEXT", "CLOB":
		return "CLOB"
	case "BLOB", "BINARY", "VARBINARY":
		return "BLOB"
	default:
		return "VARCHAR2(4000)"
	}
}

// GetDefaultSchema returns the default schema name
func (p *OracleParser) GetDefaultSchema() string {
	return ""
}

// GetSchemaPrefix returns the schema prefix
func (p *OracleParser) GetSchemaPrefix(schema string) string {
	if schema == "" {
		return ""
	}
	return schema + "."
}

// GetIdentifierQuote returns the identifier quote character
func (p *OracleParser) GetIdentifierQuote() string {
	return "\""
}

// GetStringQuote returns the string quote character
func (p *OracleParser) GetStringQuote() string {
	return "'"
}

// GetMaxIdentifierLength returns the maximum identifier length
func (p *OracleParser) GetMaxIdentifierLength() int {
	return 30
}

// GetReservedWords returns the reserved words
func (p *OracleParser) GetReservedWords() []string {
	return []string{
		"access", "add", "all", "alter", "and", "any", "as", "asc",
		"audit", "between", "by", "char", "check", "cluster", "column",
		"comment", "compress", "connect", "create", "current", "date",
		"decimal", "default", "delete", "desc", "distinct", "drop",
		"else", "exclusive", "exists", "file", "float", "for", "from",
		"grant", "group", "having", "identified", "immediate", "in",
		"increment", "index", "initial", "insert", "integer", "intersect",
		"into", "is", "level", "like", "lock", "long", "maxextents",
		"minus", "mlslabel", "mode", "modify", "noaudit", "nocompress",
		"not", "nowait", "null", "number", "of", "offline", "on",
		"online", "option", "or", "order", "pctfree", "prior",
		"privileges", "public", "raw", "rename", "resource", "revoke",
		"row", "rowid", "rownum", "rows", "select", "session", "set",
		"share", "size", "smallint", "start", "successful", "synonym",
		"sysdate", "table", "then", "to", "trigger", "uid", "union",
		"unique", "update", "user", "validate", "values", "varchar",
		"varchar2", "view", "whenever", "where", "with",
	}
}

// ConvertDataTypeFrom converts source database data type to Oracle data type
func (p *OracleParser) ConvertDataTypeFrom(sourceType string, length int, precision int, scale int) *sdc.DataType {
	return &sdc.DataType{
		Name:      sourceType,
		Length:    length,
		Precision: precision,
		Scale:     scale,
	}
}

// ParseCreateTable parses CREATE TABLE statement
func (p *OracleParser) ParseCreateTable(sql string) (*sdc.Table, error) {
	return p.parseCreateTable(sql)
}

// ParseAlterTable parses ALTER TABLE statement
func (p *OracleParser) ParseAlterTable(sql string) (*sdc.AlterTable, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseDropTable parses DROP TABLE statement
func (p *OracleParser) ParseDropTable(sql string) (*sdc.DropTable, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseCreateIndex parses CREATE INDEX statement
func (p *OracleParser) ParseCreateIndex(sql string) (*sdc.Index, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseDropIndex parses DROP INDEX statement
func (p *OracleParser) ParseDropIndex(sql string) (*sdc.DropIndex, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ValidateIdentifier validates the identifier name
func (p *OracleParser) ValidateIdentifier(name string) error {
	// Validation logic to be implemented
	return nil
}

// EscapeIdentifier escapes the identifier name
func (p *OracleParser) EscapeIdentifier(name string) string {
	// Escape logic to be implemented
	return name
}

// EscapeString escapes the string value
func (p *OracleParser) EscapeString(value string) string {
	// Escape logic to be implemented
	return value
}

// parseCreateTable parses CREATE TABLE statement
func (p *OracleParser) parseCreateTable(sql string) (*sdc.Table, error) {
	table := &sdc.Table{}

	// Extract basic table information
	tableRegex := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:"([^"]+)"|([^\s(]+))\s*\((.*?)\)`)
	matches := tableRegex.FindStringSubmatch(sql)

	if len(matches) < 3 {
		return nil, fmt.Errorf("invalid CREATE TABLE statement")
	}

	// Set table name
	if matches[1] != "" {
		table.Name = matches[1]
	} else {
		table.Name = matches[2]
	}

	// Parse column definitions
	columnDefs := strings.Split(matches[3], ",")
	var comments []string

	for _, columnDef := range columnDefs {
		columnDef = strings.TrimSpace(columnDef)
		if columnDef == "" {
			continue
		}

		// Check for constraints
		if strings.HasPrefix(strings.ToUpper(columnDef), "CONSTRAINT") ||
			strings.HasPrefix(strings.ToUpper(columnDef), "PRIMARY KEY") ||
			strings.HasPrefix(strings.ToUpper(columnDef), "FOREIGN KEY") ||
			strings.HasPrefix(strings.ToUpper(columnDef), "UNIQUE") {
			comments = append(comments, columnDef)
			continue
		}

		// Check for CHECK constraints
		if strings.HasPrefix(strings.ToUpper(columnDef), "CHECK") {
			checkRegex := regexp.MustCompile(`(?i)CHECK\s*\((.*?)\)`)
			checkMatches := checkRegex.FindStringSubmatch(columnDef)
			if len(checkMatches) > 1 {
				comments = append(comments, fmt.Sprintf("CHECK (%s)", checkMatches[1]))
			}
			continue
		}

		// Parse column definition
		column := p.parseColumn(columnDef)
		if column != nil {
			table.Columns = append(table.Columns, column)
		}
	}

	if len(comments) > 0 {
		table.Comment = strings.Join(comments, ", ")
	}

	return table, nil
}

// parseColumn parses column definition with optimized string handling
func (p *OracleParser) parseColumn(columnDef string) *sdc.Column {
	matches := oracleColumnRegex.FindStringSubmatch(columnDef)
	if len(matches) < 3 {
		return nil
	}

	column := &sdc.Column{
		Name:       strings.Trim(matches[1], "\""),
		IsNullable: true, // Default to nullable
		Nullable:   true,
	}

	// Parse data type
	dataType := &sdc.DataType{
		Name: matches[2],
	}

	// Parse length/precision/scale
	if matches[3] != "" {
		parts := strings.Split(matches[3], ",")
		if len(parts) > 0 {
			if dataType.Name == "NUMBER" {
				dataType.Precision, _ = parseNumber(parts[0])
				if len(parts) > 1 {
					dataType.Scale, _ = parseNumber(parts[1])
				}
			} else {
				dataType.Length, _ = parseNumber(parts[0])
			}
		}
	}

	column.DataType = dataType

	// Parse additional properties
	rest := matches[4]
	if rest != "" {
		// Check for DEFAULT
		if defaultMatches := oracleDefaultRegex.FindStringSubmatch(rest); defaultMatches != nil {
			column.Default = defaultMatches[1]
		}

		// Check for NOT NULL
		if strings.Contains(strings.ToUpper(rest), "NOT NULL") {
			column.IsNullable = false
			column.Nullable = false
		}

		// Check for PRIMARY KEY
		if strings.Contains(strings.ToUpper(rest), "PRIMARY KEY") {
			column.PrimaryKey = true
			column.IsNullable = false
			column.Nullable = false
		}

		// Check for UNIQUE
		if strings.Contains(strings.ToUpper(rest), "UNIQUE") {
			column.Unique = true
		}
	}

	return column
}
