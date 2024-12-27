package mysql

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mstgnz/sqlporter/parser"
)

// MySQLParser implements the parser for MySQL database
type MySQLParser struct{}

// Parse converts MySQL dump to Entity structure
func (p *MySQLParser) Parse(sql string) (*parser.Entity, error) {
	entity := &parser.Entity{}

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

// parseCreateTable parses CREATE TABLE statement
func (p *MySQLParser) parseCreateTable(sql string) (*parser.Table, error) {
	table := &parser.Table{
		Options: &parser.TableOptions{},
	}

	// Extract table name
	tableNameRegex := regexp.MustCompile(`CREATE TABLE\s+(?:IF NOT EXISTS\s+)?(?:(\w+)\.)?(\w+)\s*\((.*)\)(.*)`)
	matches := tableNameRegex.FindStringSubmatch(sql)
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid CREATE TABLE statement")
	}

	table.Schema = matches[1]
	table.Name = matches[2]
	columnDefs := matches[3]
	tableOptions := matches[4]

	// Parse columns
	columnRegex := regexp.MustCompile(`(\w+)\s+([^,]+)(?:,|$)`)
	columnMatches := columnRegex.FindAllStringSubmatch(columnDefs, -1)

	for _, match := range columnMatches {
		if len(match) < 3 {
			continue
		}
		column := p.parseColumn(match[0])
		table.Columns = append(table.Columns, column)
	}

	// Parse table options
	if tableOptions != "" {
		engineRegex := regexp.MustCompile(`ENGINE\s*=\s*(\w+)`)
		if m := engineRegex.FindStringSubmatch(tableOptions); len(m) > 1 {
			table.Options.Engine = m[1]
		}

		charsetRegex := regexp.MustCompile(`(?:DEFAULT\s+)?CHARSET\s*=\s*(\w+)`)
		if m := charsetRegex.FindStringSubmatch(tableOptions); len(m) > 1 {
			table.Options.Charset = m[1]
		}

		collateRegex := regexp.MustCompile(`COLLATE\s*=\s*(\w+)`)
		if m := collateRegex.FindStringSubmatch(tableOptions); len(m) > 1 {
			table.Options.Collation = m[1]
		}

		commentRegex := regexp.MustCompile(`COMMENT\s*=\s*'([^']*)'`)
		if m := commentRegex.FindStringSubmatch(tableOptions); len(m) > 1 {
			table.Options.Comment = m[1]
		}
	}

	return table, nil
}

// parseColumn parses column definition
func (p *MySQLParser) parseColumn(columnDef string) *parser.Column {
	column := &parser.Column{}

	// Extract column name
	nameRegex := regexp.MustCompile(`^(\w+)\s+(.*)$`)
	nameMatches := nameRegex.FindStringSubmatch(columnDef)
	if len(nameMatches) < 3 {
		return column
	}

	column.Name = nameMatches[1]
	remainingDef := nameMatches[2]

	// Extract data type
	dataTypeRegex := regexp.MustCompile(`^(\w+)(?:\(([^)]+)\))?`)
	dataTypeMatches := dataTypeRegex.FindStringSubmatch(remainingDef)

	if len(dataTypeMatches) > 1 {
		dataType := &parser.ColumnType{
			Name: strings.ToUpper(dataTypeMatches[1]),
		}

		// Parse length/precision/scale
		if len(dataTypeMatches) > 2 && dataTypeMatches[2] != "" {
			params := strings.Split(dataTypeMatches[2], ",")
			if len(params) > 0 {
				if length, err := strconv.Atoi(strings.TrimSpace(params[0])); err == nil {
					dataType.Length = length
				}
				if len(params) > 1 {
					if scale, err := strconv.Atoi(strings.TrimSpace(params[1])); err == nil {
						dataType.Scale = scale
					}
				}
			}
		}

		column.DataType = dataType
	}

	// Parse column options
	column.IsNullable = !strings.Contains(strings.ToUpper(remainingDef), "NOT NULL")
	column.AutoIncrement = strings.Contains(strings.ToUpper(remainingDef), "AUTO_INCREMENT")
	column.PrimaryKey = strings.Contains(strings.ToUpper(remainingDef), "PRIMARY KEY")
	column.Unique = strings.Contains(strings.ToUpper(remainingDef), "UNIQUE")

	// Extract default value
	defaultRegex := regexp.MustCompile(`DEFAULT\s+(?:'([^']*)'|(\d+(?:\.\d+)?))`)
	if defaultMatch := defaultRegex.FindStringSubmatch(remainingDef); len(defaultMatch) > 1 {
		if defaultMatch[1] != "" {
			column.Default = defaultMatch[1]
		} else {
			column.Default = defaultMatch[2]
		}
	}

	// Extract collation
	collateRegex := regexp.MustCompile(`COLLATE\s+(\w+)`)
	if collateMatch := collateRegex.FindStringSubmatch(remainingDef); len(collateMatch) > 1 {
		column.Collation = collateMatch[1]
	}

	// Extract comment
	commentRegex := regexp.MustCompile(`COMMENT\s+'([^']*)'`)
	if commentMatch := commentRegex.FindStringSubmatch(remainingDef); len(commentMatch) > 1 {
		column.Comment = commentMatch[1]
	}

	return column
}

// Convert transforms Entity structure to MySQL format
func (p *MySQLParser) Convert(entity *parser.Entity) (string, error) {
	var result strings.Builder

	for _, table := range entity.Tables {
		result.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", p.EscapeIdentifier(table.Name)))

		// Columns
		for i, column := range table.Columns {
			result.WriteString(fmt.Sprintf("  %s %s", p.EscapeIdentifier(column.Name), p.convertDataType(column.DataType)))

			if !column.IsNullable {
				result.WriteString(" NOT NULL")
			}
			if column.AutoIncrement {
				result.WriteString(" AUTO_INCREMENT")
			}
			if column.Default != "" {
				result.WriteString(fmt.Sprintf(" DEFAULT %s", column.Default))
			}
			if column.Collation != "" {
				result.WriteString(fmt.Sprintf(" COLLATE %s", column.Collation))
			}
			if column.Comment != "" {
				result.WriteString(fmt.Sprintf(" COMMENT '%s'", column.Comment))
			}

			if i < len(table.Columns)-1 {
				result.WriteString(",\n")
			}
		}

		// Table options
		result.WriteString("\n)")
		if table.Options != nil {
			if table.Options.Engine != "" {
				result.WriteString(fmt.Sprintf(" ENGINE=%s", table.Options.Engine))
			}
			if table.Options.Charset != "" {
				result.WriteString(fmt.Sprintf(" DEFAULT CHARSET=%s", table.Options.Charset))
			}
			if table.Options.Collation != "" {
				result.WriteString(fmt.Sprintf(" COLLATE=%s", table.Options.Collation))
			}
			if table.Options.Comment != "" {
				result.WriteString(fmt.Sprintf(" COMMENT='%s'", table.Options.Comment))
			}
		}
		result.WriteString(";\n\n")
	}

	return result.String(), nil
}

// convertDataType converts data type to MySQL format
func (p *MySQLParser) convertDataType(dataType *parser.ColumnType) string {
	if dataType == nil {
		return "VARCHAR(255)"
	}

	switch strings.ToUpper(dataType.Name) {
	case "VARCHAR", "CHAR", "VARBINARY", "BINARY":
		if dataType.Length > 0 {
			return fmt.Sprintf("%s(%d)", dataType.Name, dataType.Length)
		}
		return fmt.Sprintf("%s(255)", dataType.Name)

	case "DECIMAL", "NUMERIC":
		if dataType.Precision > 0 {
			if dataType.Scale > 0 {
				return fmt.Sprintf("%s(%d,%d)", dataType.Name, dataType.Precision, dataType.Scale)
			}
			return fmt.Sprintf("%s(%d)", dataType.Name, dataType.Precision)
		}
		return fmt.Sprintf("%s(10,0)", dataType.Name)

	case "FLOAT", "DOUBLE":
		if dataType.Length > 0 {
			if dataType.Scale > 0 {
				return fmt.Sprintf("%s(%d,%d)", dataType.Name, dataType.Length, dataType.Scale)
			}
			return fmt.Sprintf("%s(%d)", dataType.Name, dataType.Length)
		}
		return dataType.Name

	default:
		return dataType.Name
	}
}

// GetDefaultSchema returns the default schema name
func (p *MySQLParser) GetDefaultSchema() string {
	return ""
}

// GetSchemaPrefix returns the schema prefix for identifiers
func (p *MySQLParser) GetSchemaPrefix(schema string) string {
	if schema != "" {
		return p.EscapeIdentifier(schema) + "."
	}
	return ""
}

// GetIdentifierQuote returns the quote character for identifiers
func (p *MySQLParser) GetIdentifierQuote() string {
	return "`"
}

// GetStringQuote returns the quote character for strings
func (p *MySQLParser) GetStringQuote() string {
	return "'"
}

// GetMaxIdentifierLength returns the maximum length for identifiers
func (p *MySQLParser) GetMaxIdentifierLength() int {
	return 64
}

// GetReservedWords returns the list of reserved words
func (p *MySQLParser) GetReservedWords() []string {
	return []string{
		"ADD", "ALL", "ALTER", "ANALYZE", "AND", "AS", "ASC",
		"BEFORE", "BETWEEN", "BINARY", "BOTH", "BY",
		"CASE", "CHANGE", "CHARACTER", "CHECK", "COLLATE", "COLUMN", "CONDITION",
		"CONSTRAINT", "CONTINUE", "CONVERT", "CREATE", "CROSS", "CURRENT_DATE",
		"CURRENT_TIME", "CURRENT_TIMESTAMP", "CURRENT_USER", "CURSOR",
		"DATABASE", "DATABASES", "DAY_HOUR", "DAY_MICROSECOND", "DAY_MINUTE",
		"DAY_SECOND", "DEC", "DECIMAL", "DECLARE", "DEFAULT", "DELAYED", "DELETE",
		"DESC", "DESCRIBE", "DETERMINISTIC", "DISTINCT", "DISTINCTROW", "DIV",
		"DOUBLE", "DROP", "DUAL",
		// ... diğer anahtar kelimeler ...
	}
}

// ValidateIdentifier validates an identifier
func (p *MySQLParser) ValidateIdentifier(name string) error {
	if len(name) > p.GetMaxIdentifierLength() {
		return fmt.Errorf("identifier '%s' is too long (max %d characters)", name, p.GetMaxIdentifierLength())
	}
	return nil
}

// EscapeIdentifier escapes an identifier
func (p *MySQLParser) EscapeIdentifier(name string) string {
	return fmt.Sprintf("`%s`", strings.Replace(name, "`", "``", -1))
}

// EscapeString escapes a string value
func (p *MySQLParser) EscapeString(value string) string {
	return fmt.Sprintf("'%s'", strings.Replace(value, "'", "''", -1))
}

// ConvertDataTypeFrom converts source database data type to MySQL data type
func (p *MySQLParser) ConvertDataTypeFrom(sourceType string, length int, precision int, scale int) *parser.ColumnType {
	return &parser.ColumnType{
		Name:      sourceType,
		Length:    length,
		Precision: precision,
		Scale:     scale,
	}
}

// ParseCreateTable parses CREATE TABLE statement
func (p *MySQLParser) ParseCreateTable(sql string) (*parser.Table, error) {
	return p.parseCreateTable(sql)
}

// ParseAlterTable parses ALTER TABLE statement
func (p *MySQLParser) ParseAlterTable(sql string) (*parser.AlterTable, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseDropTable parses DROP TABLE statement
func (p *MySQLParser) ParseDropTable(sql string) (*parser.DropTable, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseCreateIndex parses CREATE INDEX statement
func (p *MySQLParser) ParseCreateIndex(sql string) (*parser.Index, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseDropIndex parses DROP INDEX statement
func (p *MySQLParser) ParseDropIndex(sql string) (*parser.DropIndex, error) {
	// Parse logic to be implemented
	return nil, nil
}
