package postgres

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mstgnz/sqlporter/parser"
)

// Precompiled regex patterns for better performance
var (
	postgresCreateTableRegex = regexp.MustCompile(`(?i)CREATE\s+(?:TEMPORARY\s+)?TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:(\w+)\.)?(["\w]+)\s*\((.*)\)(?:\s+TABLESPACE\s+(\w+))?(?:\s*;)?`)
	postgresColumnRegex      = regexp.MustCompile(`(?i)^\s*"?(\w+)"?\s+(\w+)(?:\s*\(([^)]+)\))?\s*(.*)$`)
	postgresDefaultRegex     = regexp.MustCompile(`(?i)DEFAULT\s+([^,\s]+|'[^']*'|"[^"]*")`)
	postgresSerialRegex      = regexp.MustCompile(`(?i)^(SMALL|BIG)?SERIAL$`)
	postgresRefRegex         = regexp.MustCompile(`(?i)REFERENCES\s+([^\s(]+)\s*\(([^)]+)\)(?:\s+ON\s+DELETE\s+(\w+))?(?:\s+ON\s+UPDATE\s+(\w+))?`)
	postgresCheckRegex       = regexp.MustCompile(`(?i)(?:CONSTRAINT\s+(\w+)\s+)?CHECK\s*\((.*?)\)`)
	postgresPrimaryKeyRegex  = regexp.MustCompile(`(?i)(?:CONSTRAINT\s+(\w+)\s+)?PRIMARY\s+KEY\s*\((.*?)\)`)
	postgresForeignKeyRegex  = regexp.MustCompile(`(?i)(?:CONSTRAINT\s+(\w+)\s+)?FOREIGN\s+KEY\s*\((.*?)\)\s*REFERENCES\s+(\w+)\s*\((.*?)\)(?:\s+ON\s+DELETE\s+(\w+))?(?:\s+ON\s+UPDATE\s+(\w+))?`)
	postgresUniqueRegex      = regexp.MustCompile(`(?i)(?:CONSTRAINT\s+(\w+)\s+)?UNIQUE\s*\((.*?)\)`)
)

// PostgresParser implements the parser for PostgreSQL database
type PostgresParser struct {
	dbInfo parser.DatabaseInfo
}

// NewPostgresParser creates a new PostgreSQL parser instance
func NewPostgresParser() *PostgresParser {
	return &PostgresParser{
		dbInfo: parser.DatabaseInfo{
			DefaultSchema:       "public",
			IdentifierQuote:     "\"",
			StringQuote:         "'",
			MaxIdentifierLength: 63,
		},
	}
}

// Parse converts PostgreSQL dump to Entity structure
func (p *PostgresParser) Parse(sql string) (*parser.Entity, error) {
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
func (p *PostgresParser) parseCreateTable(sql string) (*parser.Table, error) {
	table := &parser.Table{
		Options: &parser.TableOptions{},
	}

	matches := postgresCreateTableRegex.FindStringSubmatch(sql)
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid CREATE TABLE statement")
	}

	table.Schema = matches[1]
	if table.Schema == "" {
		table.Schema = p.GetDefaultSchema()
	}
	table.Name = matches[2]

	// Parse columns and constraints
	definitions := splitDefinitions(matches[3])
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
func (p *PostgresParser) parseColumn(columnDef string) *parser.Column {
	matches := postgresColumnRegex.FindStringSubmatch(columnDef)
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
	if defaultMatch := postgresDefaultRegex.FindStringSubmatch(matches[4]); len(defaultMatch) > 1 {
		column.Default = defaultMatch[1]
	}

	return column
}

// Convert transforms Entity structure to PostgreSQL format
func (p *PostgresParser) Convert(entity *parser.Entity) (string, error) {
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

// convertDataType converts data type to PostgreSQL format
func (p *PostgresParser) convertDataType(dataType *parser.ColumnType) string {
	if dataType == nil {
		return "VARCHAR(255)"
	}

	switch strings.ToUpper(dataType.Name) {
	case "VARCHAR", "CHAR", "CHARACTER VARYING", "CHARACTER":
		if dataType.Length > 0 {
			return fmt.Sprintf("%s(%d)", dataType.Name, dataType.Length)
		}
		return fmt.Sprintf("%s(255)", dataType.Name)

	case "NUMERIC", "DECIMAL":
		if dataType.Precision > 0 {
			if dataType.Scale > 0 {
				return fmt.Sprintf("%s(%d,%d)", dataType.Name, dataType.Precision, dataType.Scale)
			}
			return fmt.Sprintf("%s(%d)", dataType.Name, dataType.Precision)
		}
		return dataType.Name

	default:
		return dataType.Name
	}
}

// GetDefaultSchema returns the default schema name
func (p *PostgresParser) GetDefaultSchema() string {
	return p.dbInfo.DefaultSchema
}

// GetSchemaPrefix returns the schema prefix for identifiers
func (p *PostgresParser) GetSchemaPrefix(schema string) string {
	if schema != "" && schema != p.GetDefaultSchema() {
		return p.EscapeIdentifier(schema) + "."
	}
	return ""
}

// GetIdentifierQuote returns the quote character for identifiers
func (p *PostgresParser) GetIdentifierQuote() string {
	return p.dbInfo.IdentifierQuote
}

// GetStringQuote returns the quote character for strings
func (p *PostgresParser) GetStringQuote() string {
	return p.dbInfo.StringQuote
}

// GetMaxIdentifierLength returns the maximum length for identifiers
func (p *PostgresParser) GetMaxIdentifierLength() int {
	return p.dbInfo.MaxIdentifierLength
}

// GetReservedWords returns the list of reserved words
func (p *PostgresParser) GetReservedWords() []string {
	return []string{
		"ALL", "ANALYSE", "ANALYZE", "AND", "ANY", "ARRAY", "AS", "ASC",
		"ASYMMETRIC", "AUTHORIZATION", "BINARY", "BOTH", "CASE", "CAST",
		"CHECK", "COLLATE", "COLUMN", "CONSTRAINT", "CREATE", "CROSS",
		"CURRENT_DATE", "CURRENT_ROLE", "CURRENT_TIME", "CURRENT_TIMESTAMP",
		"CURRENT_USER", "DEFAULT", "DEFERRABLE", "DESC", "DISTINCT", "DO",
		"ELSE", "END", "EXCEPT", "FALSE", "FOR", "FOREIGN", "FREEZE", "FROM",
		"FULL", "GRANT", "GROUP", "HAVING", "ILIKE", "IN", "INITIALLY", "INNER",
		"INTERSECT", "INTO", "IS", "ISNULL", "JOIN", "LEADING", "LEFT", "LIKE",
		"LIMIT", "LOCALTIME", "LOCALTIMESTAMP", "NATURAL", "NOT", "NOTNULL",
		"NULL", "OFFSET", "ON", "ONLY", "OR", "ORDER", "OUTER", "OVERLAPS",
		"PLACING", "PRIMARY", "REFERENCES", "RIGHT", "SELECT", "SESSION_USER",
		"SIMILAR", "SOME", "SYMMETRIC", "TABLE", "THEN", "TO", "TRAILING",
		"TRUE", "UNION", "UNIQUE", "USER", "USING", "VERBOSE", "WHEN", "WHERE",
	}
}

// ValidateIdentifier validates an identifier
func (p *PostgresParser) ValidateIdentifier(name string) error {
	if len(name) > p.GetMaxIdentifierLength() {
		return fmt.Errorf("identifier '%s' is too long (max %d characters)", name, p.GetMaxIdentifierLength())
	}
	return nil
}

// EscapeIdentifier escapes an identifier
func (p *PostgresParser) EscapeIdentifier(name string) string {
	return fmt.Sprintf("%s%s%s", p.GetIdentifierQuote(), strings.Replace(name, p.GetIdentifierQuote(), p.GetIdentifierQuote()+p.GetIdentifierQuote(), -1), p.GetIdentifierQuote())
}

// EscapeString escapes a string value
func (p *PostgresParser) EscapeString(value string) string {
	return fmt.Sprintf("%s%s%s", p.GetStringQuote(), strings.Replace(value, p.GetStringQuote(), p.GetStringQuote()+p.GetStringQuote(), -1), p.GetStringQuote())
}

// ConvertDataTypeFrom converts source database data type to PostgreSQL data type
func (p *PostgresParser) ConvertDataTypeFrom(sourceType string, length int, precision int, scale int) *parser.ColumnType {
	return &parser.ColumnType{
		Name:      sourceType,
		Length:    length,
		Precision: precision,
		Scale:     scale,
	}
}

// Helper functions
func splitDefinitions(s string) []string {
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

// ParseCreateTable parses CREATE TABLE statement
func (p *PostgresParser) ParseCreateTable(sql string) (*parser.Table, error) {
	return p.parseCreateTable(sql)
}

// ParseAlterTable parses ALTER TABLE statement
func (p *PostgresParser) ParseAlterTable(sql string) (*parser.AlterTable, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseDropTable parses DROP TABLE statement
func (p *PostgresParser) ParseDropTable(sql string) (*parser.DropTable, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseCreateIndex parses CREATE INDEX statement
func (p *PostgresParser) ParseCreateIndex(sql string) (*parser.Index, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseDropIndex parses DROP INDEX statement
func (p *PostgresParser) ParseDropIndex(sql string) (*parser.DropIndex, error) {
	// Parse logic to be implemented
	return nil, nil
}

// parseConstraint parses table constraint definition
func (p *PostgresParser) parseConstraint(constraintDef string) *parser.Constraint {
	constraint := &parser.Constraint{}

	// Parse PRIMARY KEY constraint
	if pkMatch := postgresPrimaryKeyRegex.FindStringSubmatch(constraintDef); pkMatch != nil {
		constraint.Type = "PRIMARY KEY"
		if pkMatch[1] != "" {
			constraint.Name = pkMatch[1]
		}
		constraint.Columns = strings.Split(pkMatch[2], ",")
		for i := range constraint.Columns {
			constraint.Columns[i] = strings.TrimSpace(constraint.Columns[i])
		}
		return constraint
	}

	// Parse FOREIGN KEY constraint
	if fkMatch := postgresForeignKeyRegex.FindStringSubmatch(constraintDef); fkMatch != nil {
		constraint.Type = "FOREIGN KEY"
		if fkMatch[1] != "" {
			constraint.Name = fkMatch[1]
		}
		constraint.Columns = strings.Split(fkMatch[2], ",")
		for i := range constraint.Columns {
			constraint.Columns[i] = strings.TrimSpace(constraint.Columns[i])
		}
		constraint.RefTable = fkMatch[3]
		constraint.RefColumns = strings.Split(fkMatch[4], ",")
		for i := range constraint.RefColumns {
			constraint.RefColumns[i] = strings.TrimSpace(constraint.RefColumns[i])
		}
		if len(fkMatch) > 5 && fkMatch[5] != "" {
			constraint.OnDelete = fkMatch[5]
		}
		if len(fkMatch) > 6 && fkMatch[6] != "" {
			constraint.OnUpdate = fkMatch[6]
		}
		return constraint
	}

	// Parse UNIQUE constraint
	if uniqueMatch := postgresUniqueRegex.FindStringSubmatch(constraintDef); uniqueMatch != nil {
		constraint.Type = "UNIQUE"
		if uniqueMatch[1] != "" {
			constraint.Name = uniqueMatch[1]
		}
		constraint.Columns = strings.Split(uniqueMatch[2], ",")
		for i := range constraint.Columns {
			constraint.Columns[i] = strings.TrimSpace(constraint.Columns[i])
		}
		return constraint
	}

	return nil
}
