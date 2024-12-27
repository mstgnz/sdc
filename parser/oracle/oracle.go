package oracle

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mstgnz/sqlporter/parser"
)

// parseNumber safely parses a string to an integer
func parseNumber(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	return strconv.ParseInt(s, 10, 64)
}

var (
	// Pre-compiled regular expressions for better performance
	oracleColumnRegex      = regexp.MustCompile(`(?i)^\s*"?([^"\s(]+)"?\s+([^(,\s]+)(?:\s*\(([^)]+)\))?\s*(.*)$`)
	oracleDefaultRegex     = regexp.MustCompile(`(?i)DEFAULT\s+([^,\s]+|'[^']*'|"[^"]*")`)
	oraclePrimaryKeyRegex  = regexp.MustCompile(`(?i)PRIMARY\s+KEY(?:\s*\((.*?)\))?`)
	oracleForeignKeyRegex  = regexp.MustCompile(`(?i)FOREIGN\s+KEY\s*\((.*?)\)\s*REFERENCES\s+([^\s(]+)\s*\((.*?)\)(?:\s+ON\s+DELETE\s+(\w+))?(?:\s+ON\s+UPDATE\s+(\w+))?`)
	oracleUniqueRegex      = regexp.MustCompile(`(?i)UNIQUE(?:\s*\((.*?)\))?`)
	oracleCreateTableRegex = regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:([^\s.]+)\.)?([^\s(]+)\s*\(([\s\S]*?)\)(?:\s+TABLESPACE\s+(\w+))?(?:\s*;)?`)
)

// OracleParser implements the parser for Oracle database
type OracleParser struct {
	dbInfo parser.DatabaseInfo
}

// NewOracleParser creates a new Oracle parser instance
func NewOracleParser() *OracleParser {
	return &OracleParser{
		dbInfo: parser.DatabaseInfo{
			DefaultSchema:       "SYSTEM",
			IdentifierQuote:     "\"",
			StringQuote:         "'",
			MaxIdentifierLength: 30,
		},
	}
}

// Parse converts Oracle dump to Entity structure
func (p *OracleParser) Parse(sql string) (*parser.Entity, error) {
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
func (p *OracleParser) parseCreateTable(sql string) (*parser.Table, error) {
	table := &parser.Table{
		Options: &parser.TableOptions{},
	}

	matches := oracleCreateTableRegex.FindStringSubmatch(sql)
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid CREATE TABLE statement")
	}

	table.Schema = matches[1]
	if table.Schema == "" {
		table.Schema = p.GetDefaultSchema()
	}
	table.Name = matches[2]

	// Parse columns and constraints
	definitions := splitWithParentheses(matches[3])
	for _, def := range definitions {
		def = strings.TrimSpace(def)
		if def == "" {
			continue
		}

		if strings.HasPrefix(strings.ToUpper(def), "CONSTRAINT") ||
			strings.HasPrefix(strings.ToUpper(def), "PRIMARY KEY") ||
			strings.HasPrefix(strings.ToUpper(def), "FOREIGN KEY") ||
			strings.HasPrefix(strings.ToUpper(def), "UNIQUE") {
			constraint := p.parseConstraint(def)
			if constraint != nil {
				table.Constraints = append(table.Constraints, constraint)
			}
		} else {
			column := p.parseColumn(def)
			if column != nil {
				table.Columns = append(table.Columns, column)
			}
		}
	}

	return table, nil
}

// parseColumn parses column definition
func (p *OracleParser) parseColumn(columnDef string) *parser.Column {
	matches := oracleColumnRegex.FindStringSubmatch(columnDef)
	if len(matches) < 5 {
		return nil
	}

	column := &parser.Column{
		Name: matches[1],
		DataType: &parser.ColumnType{
			Name: strings.ToUpper(matches[2]),
		},
	}

	// Parse length/precision/scale
	if matches[3] != "" {
		params := strings.Split(matches[3], ",")
		if len(params) > 0 {
			if length, err := strconv.Atoi(strings.TrimSpace(params[0])); err == nil {
				column.DataType.Length = length
			}
			if len(params) > 1 {
				if scale, err := strconv.Atoi(strings.TrimSpace(params[1])); err == nil {
					column.DataType.Scale = scale
				}
			}
		}
	}

	// Parse column options
	options := strings.ToUpper(matches[4])
	column.IsNullable = !strings.Contains(options, "NOT NULL")
	column.PrimaryKey = strings.Contains(options, "PRIMARY KEY")
	column.Unique = strings.Contains(options, "UNIQUE")

	// Parse default value
	if defaultMatch := oracleDefaultRegex.FindStringSubmatch(matches[4]); len(defaultMatch) > 1 {
		column.Default = defaultMatch[1]
	}

	return column
}

// Convert transforms Entity structure to Oracle format
func (p *OracleParser) Convert(entity *parser.Entity) (string, error) {
	var result strings.Builder

	for _, table := range entity.Tables {
		result.WriteString(fmt.Sprintf("CREATE TABLE %s%s (\n",
			p.GetSchemaPrefix(table.Schema),
			p.EscapeIdentifier(table.Name)))

		// Columns
		for i, column := range table.Columns {
			result.WriteString(fmt.Sprintf("  %s %s",
				p.EscapeIdentifier(column.Name),
				p.convertDataType(column.DataType)))

			if !column.IsNullable {
				result.WriteString(" NOT NULL")
			}
			if column.Default != "" {
				result.WriteString(fmt.Sprintf(" DEFAULT %s", column.Default))
			}
			if column.PrimaryKey {
				result.WriteString(" PRIMARY KEY")
			}
			if column.Unique {
				result.WriteString(" UNIQUE")
			}

			if i < len(table.Columns)-1 || len(table.Constraints) > 0 {
				result.WriteString(",")
			}
			result.WriteString("\n")
		}

		// Constraints
		for i, constraint := range table.Constraints {
			result.WriteString(fmt.Sprintf("  CONSTRAINT %s %s",
				p.EscapeIdentifier(constraint.Name),
				constraint.Type))

			if len(constraint.Columns) > 0 {
				result.WriteString(fmt.Sprintf(" (%s)",
					strings.Join(constraint.Columns, ", ")))
			}

			if constraint.RefTable != "" {
				result.WriteString(fmt.Sprintf(" REFERENCES %s (%s)",
					p.EscapeIdentifier(constraint.RefTable),
					strings.Join(constraint.RefColumns, ", ")))

				if constraint.OnDelete != "" {
					result.WriteString(fmt.Sprintf(" ON DELETE %s", constraint.OnDelete))
				}
				if constraint.OnUpdate != "" {
					result.WriteString(fmt.Sprintf(" ON UPDATE %s", constraint.OnUpdate))
				}
			}

			if i < len(table.Constraints)-1 {
				result.WriteString(",")
			}
			result.WriteString("\n")
		}

		result.WriteString(");\n\n")
	}

	return result.String(), nil
}

// convertDataType converts data type to Oracle format
func (p *OracleParser) convertDataType(dataType *parser.ColumnType) string {
	if dataType == nil {
		return "VARCHAR2(255)"
	}

	switch strings.ToUpper(dataType.Name) {
	case "VARCHAR", "VARCHAR2", "CHAR", "NCHAR", "NVARCHAR2":
		if dataType.Length > 0 {
			return fmt.Sprintf("%s(%d)", dataType.Name, dataType.Length)
		}
		return fmt.Sprintf("%s(255)", dataType.Name)

	case "NUMBER", "DECIMAL":
		if dataType.Precision > 0 {
			if dataType.Scale > 0 {
				return fmt.Sprintf("%s(%d,%d)", dataType.Name, dataType.Precision, dataType.Scale)
			}
			return fmt.Sprintf("%s(%d)", dataType.Name, dataType.Precision)
		}
		return dataType.Name

	case "INTEGER", "INT":
		return "NUMBER(10)"

	case "FLOAT", "DOUBLE":
		return "NUMBER"

	default:
		return dataType.Name
	}
}

// GetDefaultSchema returns the default schema name
func (p *OracleParser) GetDefaultSchema() string {
	return p.dbInfo.DefaultSchema
}

// GetSchemaPrefix returns the schema prefix for identifiers
func (p *OracleParser) GetSchemaPrefix(schema string) string {
	if schema != "" && schema != p.GetDefaultSchema() {
		return p.EscapeIdentifier(schema) + "."
	}
	return ""
}

// GetIdentifierQuote returns the quote character for identifiers
func (p *OracleParser) GetIdentifierQuote() string {
	return p.dbInfo.IdentifierQuote
}

// GetStringQuote returns the quote character for strings
func (p *OracleParser) GetStringQuote() string {
	return p.dbInfo.StringQuote
}

// GetMaxIdentifierLength returns the maximum length for identifiers
func (p *OracleParser) GetMaxIdentifierLength() int {
	return p.dbInfo.MaxIdentifierLength
}

// GetReservedWords returns the list of reserved words
func (p *OracleParser) GetReservedWords() []string {
	return []string{
		"ACCESS", "ADD", "ALL", "ALTER", "AND", "ANY", "AS", "ASC",
		"AUDIT", "BETWEEN", "BY", "CHAR", "CHECK", "CLUSTER", "COLUMN",
		"COMMENT", "COMPRESS", "CONNECT", "CREATE", "CURRENT", "DATE",
		"DECIMAL", "DEFAULT", "DELETE", "DESC", "DISTINCT", "DROP",
		"ELSE", "EXCLUSIVE", "EXISTS", "FILE", "FLOAT", "FOR", "FROM",
		"GRANT", "GROUP", "HAVING", "IDENTIFIED", "IMMEDIATE", "IN",
		"INCREMENT", "INDEX", "INITIAL", "INSERT", "INTEGER", "INTERSECT",
		"INTO", "IS", "LEVEL", "LIKE", "LOCK", "LONG", "MAXEXTENTS",
		"MINUS", "MLSLABEL", "MODE", "MODIFY", "NOAUDIT", "NOCOMPRESS",
		"NOT", "NOWAIT", "NULL", "NUMBER", "OF", "OFFLINE", "ON",
		"ONLINE", "OPTION", "OR", "ORDER", "PCTFREE", "PRIOR",
		"PRIVILEGES", "PUBLIC", "RAW", "RENAME", "RESOURCE", "REVOKE",
		"ROW", "ROWID", "ROWNUM", "ROWS", "SELECT", "SESSION", "SET",
		"SHARE", "SIZE", "SMALLINT", "START", "SUCCESSFUL", "SYNONYM",
		"SYSDATE", "TABLE", "THEN", "TO", "TRIGGER", "UID", "UNION",
		"UNIQUE", "UPDATE", "USER", "VALIDATE", "VALUES", "VARCHAR",
		"VARCHAR2", "VIEW", "WHENEVER", "WHERE", "WITH",
	}
}

// ValidateIdentifier validates an identifier
func (p *OracleParser) ValidateIdentifier(name string) error {
	if len(name) > p.GetMaxIdentifierLength() {
		return fmt.Errorf("identifier '%s' is too long (max %d characters)", name, p.GetMaxIdentifierLength())
	}
	return nil
}

// EscapeIdentifier escapes an identifier
func (p *OracleParser) EscapeIdentifier(name string) string {
	return fmt.Sprintf("%s%s%s", p.GetIdentifierQuote(), strings.Replace(name, p.GetIdentifierQuote(), p.GetIdentifierQuote()+p.GetIdentifierQuote(), -1), p.GetIdentifierQuote())
}

// EscapeString escapes a string value
func (p *OracleParser) EscapeString(value string) string {
	return fmt.Sprintf("%s%s%s", p.GetStringQuote(), strings.Replace(value, p.GetStringQuote(), p.GetStringQuote()+p.GetStringQuote(), -1), p.GetStringQuote())
}

// ConvertDataTypeFrom converts source database data type to Oracle data type
func (p *OracleParser) ConvertDataTypeFrom(sourceType string, length int, precision int, scale int) *parser.ColumnType {
	return &parser.ColumnType{
		Name:      sourceType,
		Length:    length,
		Precision: precision,
		Scale:     scale,
	}
}

// Helper functions
func splitWithParentheses(s string) []string {
	var result []string
	var current strings.Builder
	var depth int
	var inQuote bool
	var quoteChar rune

	for _, r := range s {
		switch {
		case r == '(' && !inQuote:
			depth++
			current.WriteRune(r)
		case r == ')' && !inQuote:
			depth--
			current.WriteRune(r)
		case (r == '\'' || r == '"') && (quoteChar == 0 || quoteChar == r):
			inQuote = !inQuote
			if inQuote {
				quoteChar = r
			} else {
				quoteChar = 0
			}
			current.WriteRune(r)
		case r == ',' && depth == 0 && !inQuote:
			result = append(result, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	// Trim spaces
	for i := range result {
		result[i] = strings.TrimSpace(result[i])
	}

	return result
}

// parseConstraint parses table constraint definition
func (p *OracleParser) parseConstraint(constraintDef string) *parser.Constraint {
	constraint := &parser.Constraint{}

	// Parse PRIMARY KEY constraint
	if pkMatch := oraclePrimaryKeyRegex.FindStringSubmatch(constraintDef); pkMatch != nil {
		constraint.Type = "PRIMARY KEY"
		if len(pkMatch) > 1 && pkMatch[1] != "" {
			constraint.Columns = strings.Split(pkMatch[1], ",")
			for i := range constraint.Columns {
				constraint.Columns[i] = strings.TrimSpace(constraint.Columns[i])
			}
		}
		return constraint
	}

	// Parse FOREIGN KEY constraint
	if fkMatch := oracleForeignKeyRegex.FindStringSubmatch(constraintDef); fkMatch != nil {
		constraint.Type = "FOREIGN KEY"
		constraint.Columns = strings.Split(fkMatch[1], ",")
		for i := range constraint.Columns {
			constraint.Columns[i] = strings.TrimSpace(constraint.Columns[i])
		}
		constraint.RefTable = fkMatch[2]
		constraint.RefColumns = strings.Split(fkMatch[3], ",")
		for i := range constraint.RefColumns {
			constraint.RefColumns[i] = strings.TrimSpace(constraint.RefColumns[i])
		}
		if len(fkMatch) > 4 && fkMatch[4] != "" {
			constraint.OnDelete = fkMatch[4]
		}
		if len(fkMatch) > 5 && fkMatch[5] != "" {
			constraint.OnUpdate = fkMatch[5]
		}
		return constraint
	}

	// Parse UNIQUE constraint
	if uniqueMatch := oracleUniqueRegex.FindStringSubmatch(constraintDef); uniqueMatch != nil {
		constraint.Type = "UNIQUE"
		if len(uniqueMatch) > 1 && uniqueMatch[1] != "" {
			constraint.Columns = strings.Split(uniqueMatch[1], ",")
			for i := range constraint.Columns {
				constraint.Columns[i] = strings.TrimSpace(constraint.Columns[i])
			}
		}
		return constraint
	}

	return nil
}

// ParseCreateTable parses CREATE TABLE statement
func (p *OracleParser) ParseCreateTable(sql string) (*parser.Table, error) {
	return p.parseCreateTable(sql)
}

// ParseAlterTable parses ALTER TABLE statement
func (p *OracleParser) ParseAlterTable(sql string) (*parser.AlterTable, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseDropTable parses DROP TABLE statement
func (p *OracleParser) ParseDropTable(sql string) (*parser.DropTable, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseCreateIndex parses CREATE INDEX statement
func (p *OracleParser) ParseCreateIndex(sql string) (*parser.Index, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseDropIndex parses DROP INDEX statement
func (p *OracleParser) ParseDropIndex(sql string) (*parser.DropIndex, error) {
	// Parse logic to be implemented
	return nil, nil
}
