package sqlserver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mstgnz/sqlporter/parser"
)

// Precompiled regex patterns for better performance
var (
	sqlServerCreateTableRegex = regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:\[?([^\]\.]+)\]?\.)?(?:\[?([^\]\s(]+)\]?)\s*\(([\s\S]*?)\)(?:\s+ON\s+\[?([^\]]+)\]?)?(?:\s*;)?`)
	sqlServerColumnRegex      = regexp.MustCompile(`(?i)^\s*\[?([^\[\]\s(]+)\]?\s+([^(,\s]+)(?:\s*\(([^)]+)\))?\s*(.*)$`)
	sqlServerDefaultRegex     = regexp.MustCompile(`(?i)DEFAULT\s+([^,\s]+|'[^']*'|"[^"]*")`)
	sqlServerIdentityRegex    = regexp.MustCompile(`(?i)IDENTITY(?:\s*\((\d+)\s*,\s*(\d+)\s*\))?`)
	sqlServerCollateRegex     = regexp.MustCompile(`(?i)COLLATE\s+([^\s,]+)`)
	sqlServerPrimaryKeyRegex  = regexp.MustCompile(`(?i)PRIMARY\s+KEY(?:\s*\((.*?)\))?`)
	sqlServerForeignKeyRegex  = regexp.MustCompile(`(?i)FOREIGN\s+KEY\s*\((.*?)\)\s*REFERENCES\s+([^\s(]+)\s*\((.*?)\)(?:\s+ON\s+DELETE\s+(\w+))?(?:\s+ON\s+UPDATE\s+(\w+))?`)
	sqlServerUniqueRegex      = regexp.MustCompile(`(?i)UNIQUE(?:\s*\((.*?)\))?`)
)

// SQLServerParser implements the parser for SQL Server database
type SQLServerParser struct {
	dbInfo parser.DatabaseInfo
}

// NewSQLServerParser creates a new SQL Server parser instance
func NewSQLServerParser() *SQLServerParser {
	return &SQLServerParser{
		dbInfo: parser.DatabaseInfo{
			DefaultSchema:       "dbo",
			IdentifierQuote:     "[",
			StringQuote:         "'",
			MaxIdentifierLength: 128,
		},
	}
}

// Parse converts SQL Server dump to Entity structure
func (p *SQLServerParser) Parse(sql string) (*parser.Entity, error) {
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
func (p *SQLServerParser) parseCreateTable(sql string) (*parser.Table, error) {
	table := &parser.Table{
		Options: &parser.TableOptions{},
	}

	matches := sqlServerCreateTableRegex.FindStringSubmatch(sql)
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
func (p *SQLServerParser) parseColumn(columnDef string) *parser.Column {
	matches := sqlServerColumnRegex.FindStringSubmatch(columnDef)
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
	column.AutoIncrement = strings.Contains(options, "IDENTITY")
	column.Unique = strings.Contains(options, "UNIQUE")

	// Parse default value
	if defaultMatch := sqlServerDefaultRegex.FindStringSubmatch(matches[4]); len(defaultMatch) > 1 {
		column.Default = defaultMatch[1]
	}

	return column
}

// Convert transforms Entity structure to SQL Server format
func (p *SQLServerParser) Convert(entity *parser.Entity) (string, error) {
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
			if column.AutoIncrement {
				result.WriteString(" IDENTITY(1,1)")
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

// convertDataType converts data type to SQL Server format
func (p *SQLServerParser) convertDataType(dataType *parser.ColumnType) string {
	if dataType == nil {
		return "VARCHAR(255)"
	}

	switch strings.ToUpper(dataType.Name) {
	case "VARCHAR", "NVARCHAR", "CHAR", "NCHAR":
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
		return dataType.Name

	case "FLOAT", "REAL":
		if dataType.Precision > 0 {
			return fmt.Sprintf("%s(%d)", dataType.Name, dataType.Precision)
		}
		return dataType.Name

	default:
		return dataType.Name
	}
}

// GetDefaultSchema returns the default schema name
func (p *SQLServerParser) GetDefaultSchema() string {
	return p.dbInfo.DefaultSchema
}

// GetSchemaPrefix returns the schema prefix for identifiers
func (p *SQLServerParser) GetSchemaPrefix(schema string) string {
	if schema != "" && schema != p.GetDefaultSchema() {
		return p.EscapeIdentifier(schema) + "."
	}
	return ""
}

// GetIdentifierQuote returns the quote character for identifiers
func (p *SQLServerParser) GetIdentifierQuote() string {
	return p.dbInfo.IdentifierQuote
}

// GetStringQuote returns the quote character for strings
func (p *SQLServerParser) GetStringQuote() string {
	return p.dbInfo.StringQuote
}

// GetMaxIdentifierLength returns the maximum length for identifiers
func (p *SQLServerParser) GetMaxIdentifierLength() int {
	return p.dbInfo.MaxIdentifierLength
}

// GetReservedWords returns the list of reserved words
func (p *SQLServerParser) GetReservedWords() []string {
	return []string{
		"ADD", "ALL", "ALTER", "AND", "ANY", "AS", "ASC", "AUTHORIZATION",
		"BACKUP", "BEGIN", "BETWEEN", "BREAK", "BROWSE", "BULK", "BY",
		"CASCADE", "CASE", "CHECK", "CHECKPOINT", "CLOSE", "CLUSTERED",
		"COALESCE", "COLLATE", "COLUMN", "COMMIT", "COMPUTE", "CONSTRAINT",
		"CONTAINS", "CONTAINSTABLE", "CONTINUE", "CONVERT", "CREATE",
		"CROSS", "CURRENT", "CURRENT_DATE", "CURRENT_TIME",
		"CURRENT_TIMESTAMP", "CURRENT_USER", "CURSOR", "DATABASE", "DBCC",
		"DEALLOCATE", "DECLARE", "DEFAULT", "DELETE", "DENY", "DESC",
		"DISK", "DISTINCT", "DISTRIBUTED", "DOUBLE", "DROP", "DUMP",
		"ELSE", "END", "ERRLVL", "ESCAPE", "EXCEPT", "EXEC", "EXECUTE",
		"EXISTS", "EXIT", "EXTERNAL", "FETCH", "FILE", "FILLFACTOR",
		"FOR", "FOREIGN", "FREETEXT", "FREETEXTTABLE", "FROM", "FULL",
		"FUNCTION", "GOTO", "GRANT", "GROUP", "HAVING", "HOLDLOCK",
		"IDENTITY", "IDENTITY_INSERT", "IDENTITYCOL", "IF", "IN", "INDEX",
		"INNER", "INSERT", "INTERSECT", "INTO", "IS", "JOIN", "KEY",
		"KILL", "LEFT", "LIKE", "LINENO", "LOAD", "MERGE", "NATIONAL",
		"NOCHECK", "NONCLUSTERED", "NOT", "NULL", "NULLIF", "OF", "OFF",
		"OFFSETS", "ON", "OPEN", "OPENDATASOURCE", "OPENQUERY",
		"OPENROWSET", "OPENXML", "OPTION", "OR", "ORDER", "OUTER", "OVER",
		"PERCENT", "PIVOT", "PLAN", "PRECISION", "PRIMARY", "PRINT",
		"PROC", "PROCEDURE", "PUBLIC", "RAISERROR", "READ", "READTEXT",
		"RECONFIGURE", "REFERENCES", "REPLICATION", "RESTORE", "RESTRICT",
		"RETURN", "REVERT", "REVOKE", "RIGHT", "ROLLBACK", "ROWCOUNT",
		"ROWGUIDCOL", "RULE", "SAVE", "SCHEMA", "SECURITYAUDIT",
		"SELECT", "SEMANTICKEYPHRASETABLE", "SEMANTICSIMILARITYDETAILSTABLE",
		"SEMANTICSIMILARITYTABLE", "SESSION_USER", "SET", "SETUSER",
		"SHUTDOWN", "SOME", "STATISTICS", "SYSTEM_USER", "TABLE",
		"TABLESAMPLE", "TEXTSIZE", "THEN", "TO", "TOP", "TRAN",
		"TRANSACTION", "TRIGGER", "TRUNCATE", "TRY_CONVERT", "TSEQUAL",
		"UNION", "UNIQUE", "UNPIVOT", "UPDATE", "UPDATETEXT", "USE",
		"USER", "VALUES", "VARYING", "VIEW", "WAITFOR", "WHEN", "WHERE",
		"WHILE", "WITH", "WITHIN GROUP", "WRITETEXT",
	}
}

// ValidateIdentifier validates an identifier
func (p *SQLServerParser) ValidateIdentifier(name string) error {
	if len(name) > p.GetMaxIdentifierLength() {
		return fmt.Errorf("identifier '%s' is too long (max %d characters)", name, p.GetMaxIdentifierLength())
	}
	return nil
}

// EscapeIdentifier escapes an identifier
func (p *SQLServerParser) EscapeIdentifier(name string) string {
	return fmt.Sprintf("%s%s%s", p.GetIdentifierQuote(), strings.Replace(name, p.GetIdentifierQuote(), p.GetIdentifierQuote()+p.GetIdentifierQuote(), -1), p.GetIdentifierQuote())
}

// EscapeString escapes a string value
func (p *SQLServerParser) EscapeString(value string) string {
	return fmt.Sprintf("%s%s%s", p.GetStringQuote(), strings.Replace(value, p.GetStringQuote(), p.GetStringQuote()+p.GetStringQuote(), -1), p.GetStringQuote())
}

// ConvertDataTypeFrom converts source database data type to SQL Server data type
func (p *SQLServerParser) ConvertDataTypeFrom(sourceType string, length int, precision int, scale int) *parser.ColumnType {
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
func (p *SQLServerParser) parseConstraint(constraintDef string) *parser.Constraint {
	constraint := &parser.Constraint{}

	// Parse PRIMARY KEY constraint
	if pkMatch := sqlServerPrimaryKeyRegex.FindStringSubmatch(constraintDef); pkMatch != nil {
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
	if fkMatch := sqlServerForeignKeyRegex.FindStringSubmatch(constraintDef); fkMatch != nil {
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
	if uniqueMatch := sqlServerUniqueRegex.FindStringSubmatch(constraintDef); uniqueMatch != nil {
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
func (p *SQLServerParser) ParseCreateTable(sql string) (*parser.Table, error) {
	return p.parseCreateTable(sql)
}

// ParseAlterTable parses ALTER TABLE statement
func (p *SQLServerParser) ParseAlterTable(sql string) (*parser.AlterTable, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseDropTable parses DROP TABLE statement
func (p *SQLServerParser) ParseDropTable(sql string) (*parser.DropTable, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseCreateIndex parses CREATE INDEX statement
func (p *SQLServerParser) ParseCreateIndex(sql string) (*parser.Index, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseDropIndex parses DROP INDEX statement
func (p *SQLServerParser) ParseDropIndex(sql string) (*parser.DropIndex, error) {
	// Parse logic to be implemented
	return nil, nil
}
