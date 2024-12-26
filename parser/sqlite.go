package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mstgnz/sdc"
)

// Precompiled regex patterns for better performance
var (
	createTableRegex      = regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:([^\s.]+)\.)?([^\s(]+)\s*\((.*)\)`)
	inlineCommentRegex    = regexp.MustCompile(`--.*$`)
	multiLineCommentRegex = regexp.MustCompile(`/\*.*?\*/`)
	defaultValueRegex     = regexp.MustCompile(`(?i)DEFAULT\s+(.+)`)
	foreignKeyRegex       = regexp.MustCompile(`(?i)REFERENCES\s+(?:([^\s.]+)\.)?([^\s(]+)\s*\(([^)]+)\)`)
	sqliteColumnRegex     = regexp.MustCompile(`(?i)^\s*"?([^"\s(]+)"?\s+([^(,\s]+)(?:\s*\(([^)]+)\))?\s*(.*)$`)
	sqliteDefaultRegex    = regexp.MustCompile(`(?i)DEFAULT\s+([^,\s]+|'[^']*'|"[^"]*")`)
)

// SQLiteParser implements the parser for SQLite database
type SQLiteParser struct {
	dbInfo DatabaseInfo
	// Add buffer pool for string builders to reduce allocations
	builderPool strings.Builder
}

// NewSQLiteParser creates a new SQLite parser
func NewSQLiteParser() *SQLiteParser {
	return &SQLiteParser{
		dbInfo: DatabaseInfoMap[SQLite],
	}
}

// Parse converts SQLite dump to Entity structure
func (p *SQLiteParser) Parse(sql string) (*sdc.Entity, error) {
	entity := &sdc.Entity{
		Tables: make([]*sdc.Table, 0), // Pre-allocate slice
	}

	// Remove comments and normalize whitespace once
	sql = removeComments(sql)
	sql = strings.TrimSpace(sql)

	// Split SQL statements efficiently
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

// parseCreateTable parses CREATE TABLE statement with optimized string handling
func (p *SQLiteParser) parseCreateTable(sql string) (*sdc.Table, error) {
	matches := createTableRegex.FindStringSubmatch(sql)
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid CREATE TABLE statement")
	}

	table := &sdc.Table{
		Schema:      strings.Trim(matches[1], "[]\""),
		Name:        strings.Trim(matches[2], "[]\""),
		Columns:     make([]*sdc.Column, 0),     // Pre-allocate slices
		Constraints: make([]*sdc.Constraint, 0), // Pre-allocate slices
	}

	// Parse column definitions efficiently
	defs := splitWithParentheses(matches[3])
	for _, def := range defs {
		def = strings.TrimSpace(def)
		if def == "" {
			continue
		}

		// Parse constraints
		upperDef := strings.ToUpper(def)
		if strings.HasPrefix(upperDef, "CONSTRAINT") ||
			strings.HasPrefix(upperDef, "PRIMARY KEY") ||
			strings.HasPrefix(upperDef, "FOREIGN KEY") ||
			strings.HasPrefix(upperDef, "UNIQUE") {
			constraint := p.parseConstraint(def)
			if constraint != nil {
				table.Constraints = append(table.Constraints, constraint)
			}
			continue
		}

		// Parse column definition
		column := p.parseColumn(def)
		if column != nil {
			table.Columns = append(table.Columns, column)
		}
	}

	return table, nil
}

// splitWithParentheses splits a string by commas while respecting parentheses
func splitWithParentheses(s string) []string {
	result := make([]string, 0, 8) // Pre-allocate slice with reasonable capacity
	var current strings.Builder
	parentheses := 0
	inQuote := false
	quoteChar := rune(0)

	for _, char := range s {
		switch char {
		case '(':
			if !inQuote {
				parentheses++
			}
			current.WriteRune(char)
		case ')':
			if !inQuote {
				parentheses--
			}
			current.WriteRune(char)
		case '"', '\'', '`':
			if inQuote && char == quoteChar {
				inQuote = false
				quoteChar = 0
			} else if !inQuote {
				inQuote = true
				quoteChar = char
			}
			current.WriteRune(char)
		case ',':
			if parentheses == 0 && !inQuote {
				result = append(result, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		result = append(result, strings.TrimSpace(current.String()))
	}

	return result
}

// parseColumn parses column definition with optimized string handling
func (p *SQLiteParser) parseColumn(columnDef string) *sdc.Column {
	matches := sqliteColumnRegex.FindStringSubmatch(columnDef)
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
			if dataType.Name == "decimal" || dataType.Name == "numeric" {
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
		if defaultMatches := sqliteDefaultRegex.FindStringSubmatch(rest); defaultMatches != nil {
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
			if strings.Contains(strings.ToUpper(rest), "AUTOINCREMENT") {
				column.AutoIncrement = true
			}
		}

		// Check for UNIQUE
		if strings.Contains(strings.ToUpper(rest), "UNIQUE") {
			column.Unique = true
		}
	}

	return column
}

// splitQuotedString splits a string by spaces while respecting quotes
func splitQuotedString(s string) []string {
	var result []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)
	lastWasSpace := true

	for _, char := range s {
		switch {
		case char == '"' || char == '\'' || char == '`':
			if inQuote && char == quoteChar {
				inQuote = false
				quoteChar = 0
			} else if !inQuote {
				inQuote = true
				quoteChar = char
			}
			current.WriteRune(char)
			lastWasSpace = false
		case char == ' ' || char == '\t':
			if inQuote {
				current.WriteRune(char)
			} else if !lastWasSpace {
				if current.Len() > 0 {
					result = append(result, current.String())
					current.Reset()
				}
				lastWasSpace = true
			}
		default:
			current.WriteRune(char)
			lastWasSpace = false
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// removeComments removes SQL comments efficiently
func removeComments(sql string) string {
	// Remove inline comments
	sql = inlineCommentRegex.ReplaceAllString(sql, "")
	// Remove multi-line comments
	sql = multiLineCommentRegex.ReplaceAllString(sql, "")
	return sql
}

// Convert transforms Entity structure to SQLite format with optimized string handling
func (p *SQLiteParser) Convert(entity *sdc.Entity) (string, error) {
	// Reset and reuse string builder
	p.builderPool.Reset()
	result := &p.builderPool

	for _, table := range entity.Tables {
		// CREATE TABLE statement
		result.WriteString("CREATE TABLE ")
		result.WriteString(table.Name)
		result.WriteString(" (\n")

		// Columns
		for i, column := range table.Columns {
			result.WriteString("    ")
			result.WriteString(column.Name)
			result.WriteString(" ")
			result.WriteString(p.convertDataType(column.DataType))

			if !column.IsNullable {
				result.WriteString(" NOT NULL")
			}

			if column.Default != "" {
				result.WriteString(" DEFAULT ")
				result.WriteString(column.Default)
			}

			if column.PrimaryKey {
				result.WriteString(" PRIMARY KEY")
				if column.AutoIncrement {
					result.WriteString(" AUTOINCREMENT")
				}
			}

			if column.Unique {
				result.WriteString(" UNIQUE")
			}

			if i < len(table.Columns)-1 {
				result.WriteString(",\n")
			}
		}

		result.WriteString("\n);\n\n")
	}

	return result.String(), nil
}

// convertDataType converts data type to SQLite format
func (p *SQLiteParser) convertDataType(dataType *sdc.DataType) string {
	if dataType == nil {
		return "TEXT"
	}

	switch strings.ToUpper(dataType.Name) {
	case "VARCHAR", "NVARCHAR", "CHAR", "NCHAR", "TEXT", "NTEXT", "CLOB":
		return "TEXT"
	case "INT", "INTEGER", "BIGINT", "SMALLINT", "TINYINT":
		return "INTEGER"
	case "DECIMAL", "NUMERIC", "FLOAT", "REAL", "DOUBLE":
		return "REAL"
	case "BOOLEAN", "BIT":
		return "INTEGER"
	case "DATE", "TIME", "DATETIME", "TIMESTAMP":
		return "TEXT"
	case "BLOB", "BINARY", "VARBINARY":
		return "BLOB"
	default:
		return "TEXT"
	}
}

// GetDefaultSchema returns the default schema name
func (p *SQLiteParser) GetDefaultSchema() string {
	return p.dbInfo.DefaultSchema
}

// GetSchemaPrefix returns the schema prefix
func (p *SQLiteParser) GetSchemaPrefix(schema string) string {
	if schema == "" || schema == p.dbInfo.DefaultSchema {
		return ""
	}
	return schema + "."
}

// GetIdentifierQuote returns the identifier quote character
func (p *SQLiteParser) GetIdentifierQuote() string {
	return p.dbInfo.IdentifierQuote
}

// GetStringQuote returns the string quote character
func (p *SQLiteParser) GetStringQuote() string {
	return p.dbInfo.StringQuote
}

// GetMaxIdentifierLength returns the maximum identifier length
func (p *SQLiteParser) GetMaxIdentifierLength() int {
	return p.dbInfo.MaxIdentifierLength
}

// GetReservedWords returns the reserved words
func (p *SQLiteParser) GetReservedWords() []string {
	return p.dbInfo.ReservedWords
}

// ConvertDataTypeFrom converts source database data type to SQLite data type
func (p *SQLiteParser) ConvertDataTypeFrom(sourceType string, length int, precision int, scale int) *sdc.DataType {
	return &sdc.DataType{
		Name:      sourceType,
		Length:    length,
		Precision: precision,
		Scale:     scale,
	}
}

// ParseCreateTable parses CREATE TABLE statement
func (p *SQLiteParser) ParseCreateTable(sql string) (*sdc.Table, error) {
	// Remove comments and extra whitespace
	sql = removeComments(sql)
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "CREATE TABLE") {
		return nil, fmt.Errorf("could not parse CREATE TABLE statement")
	}

	// Extract table name and check for IF NOT EXISTS
	tableRegex := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:\[?([^\s\]]+)\]?)\s*\((.*)\)`)
	matches := tableRegex.FindStringSubmatch(sql)
	if len(matches) < 3 {
		return nil, fmt.Errorf("could not parse CREATE TABLE statement")
	}

	table := &sdc.Table{
		Name: strings.Trim(matches[1], "[]"),
	}

	// Split column definitions and constraints
	columnDefs := strings.Split(matches[2], ",")
	for _, def := range columnDefs {
		def = strings.TrimSpace(def)
		if def == "" {
			continue
		}

		// Parse constraints
		if strings.HasPrefix(strings.ToUpper(def), "CONSTRAINT") ||
			strings.HasPrefix(strings.ToUpper(def), "PRIMARY KEY") ||
			strings.HasPrefix(strings.ToUpper(def), "FOREIGN KEY") ||
			strings.HasPrefix(strings.ToUpper(def), "UNIQUE") ||
			strings.HasPrefix(strings.ToUpper(def), "CHECK") {
			constraint := p.parseConstraint(def)
			if constraint != nil {
				table.Constraints = append(table.Constraints, constraint)
			}
			continue
		}

		// Parse column
		column := p.parseColumn(def)
		if column != nil {
			table.Columns = append(table.Columns, column)
		}
	}

	return table, nil
}

// ParseAlterTable parses ALTER TABLE statement
func (p *SQLiteParser) ParseAlterTable(sql string) (*sdc.Table, error) {
	// Remove comments and extra whitespace
	sql = removeComments(sql)
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "ALTER TABLE") {
		return nil, fmt.Errorf("invalid ALTER TABLE statement")
	}

	// Extract schema and table name
	tableRegex := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+(?:([^.]+)\.)?([^\s]+)\s+(.*)`)
	matches := tableRegex.FindStringSubmatch(sql)
	if len(matches) < 4 {
		return nil, fmt.Errorf("could not parse ALTER TABLE statement")
	}

	table := &sdc.Table{
		Schema: matches[1],
		Name:   matches[2],
	}

	action := strings.TrimSpace(matches[3])

	// Parse ADD COLUMN
	if strings.HasPrefix(strings.ToUpper(action), "ADD") {
		columnDef := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(action, "ADD"), "COLUMN"))
		column := p.parseColumn(columnDef)
		if column != nil {
			table.Columns = append(table.Columns, column)
		}
		return table, nil
	}

	// Parse DROP COLUMN
	if strings.HasPrefix(strings.ToUpper(action), "DROP") {
		return table, nil
	}

	// Parse RENAME TO
	if strings.HasPrefix(strings.ToUpper(action), "RENAME TO") {
		newName := strings.TrimSpace(strings.TrimPrefix(action, "RENAME TO"))
		table.Name = strings.TrimSuffix(newName, ";")
		return table, nil
	}

	return nil, fmt.Errorf("unsupported ALTER TABLE action")
}

// ParseDropTable parses DROP TABLE statement
func (p *SQLiteParser) ParseDropTable(sql string) (*sdc.Table, error) {
	// Remove comments and extra whitespace
	sql = removeComments(sql)
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "DROP TABLE") {
		return nil, fmt.Errorf("invalid DROP TABLE statement")
	}

	// Extract schema and table name
	tableRegex := regexp.MustCompile(`(?i)DROP\s+TABLE\s+(?:IF\s+EXISTS\s+)?(?:([^.]+)\.)?([^\s;]+)`)
	matches := tableRegex.FindStringSubmatch(sql)
	if len(matches) < 3 {
		return nil, fmt.Errorf("could not parse DROP TABLE statement")
	}

	table := &sdc.Table{
		Schema: matches[1],
		Name:   matches[2],
	}

	return table, nil
}

// ParseCreateIndex parses CREATE INDEX statement
func (p *SQLiteParser) ParseCreateIndex(sql string) (*sdc.Index, error) {
	// Remove comments and extra whitespace
	sql = removeComments(sql)
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "CREATE") {
		return nil, fmt.Errorf("invalid CREATE INDEX statement")
	}

	index := &sdc.Index{}

	// Parse index name and table
	var indexRegex *regexp.Regexp
	if strings.Contains(strings.ToUpper(sql), "UNIQUE") {
		indexRegex = regexp.MustCompile(`CREATE\s+UNIQUE\s+INDEX\s+(\w+)\s+ON\s+\w+\s*\((.*?)\)`)
		index.Unique = true
	} else {
		indexRegex = regexp.MustCompile(`CREATE\s+INDEX\s+(\w+)\s+ON\s+\w+\s*\((.*?)\)`)
	}

	matches := indexRegex.FindStringSubmatch(sql)
	if len(matches) < 3 {
		return nil, fmt.Errorf("could not parse CREATE INDEX statement")
	}

	index.Name = matches[1]
	columnList := strings.Split(matches[2], ",")
	for _, col := range columnList {
		col = strings.TrimSpace(col)
		if col != "" {
			index.Columns = append(index.Columns, col)
		}
	}

	return index, nil
}

// ParseDropIndex parses DROP INDEX statement
func (p *SQLiteParser) ParseDropIndex(sql string) (*sdc.Index, error) {
	// Remove comments and extra whitespace
	sql = removeComments(sql)
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "DROP INDEX") {
		return nil, fmt.Errorf("invalid DROP INDEX statement")
	}

	index := &sdc.Index{}

	// Extract index name
	dropRegex := regexp.MustCompile(`DROP\s+INDEX\s+(?:IF\s+EXISTS\s+)?([^\s;]+)`)
	matches := dropRegex.FindStringSubmatch(sql)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not parse DROP INDEX statement")
	}

	index.Name = matches[1]

	return index, nil
}

// ValidateIdentifier validates the identifier name
func (p *SQLiteParser) ValidateIdentifier(name string) error {
	// Maksimum uzunluk kontrolü
	if p.dbInfo.MaxIdentifierLength > 0 && len(name) > p.dbInfo.MaxIdentifierLength {
		return fmt.Errorf("identifier '%s' exceeds maximum length of %d", name, p.dbInfo.MaxIdentifierLength)
	}

	// Ayrılmış kelime kontrolü
	for _, word := range p.dbInfo.ReservedWords {
		if strings.ToUpper(name) == strings.ToUpper(word) {
			return fmt.Errorf("identifier '%s' is a reserved word", name)
		}
	}

	return nil
}

// EscapeIdentifier escapes the identifier name
func (p *SQLiteParser) EscapeIdentifier(name string) string {
	// Escape logic to be implemented
	return name
}

// EscapeString escapes the string value
func (p *SQLiteParser) EscapeString(value string) string {
	// Escape logic to be implemented
	return value
}

// ParseSQL parses SQLite SQL statements and returns an Entity
func (p *SQLiteParser) ParseSQL(sql string) (*sdc.Entity, error) {
	entity := &sdc.Entity{}

	// Find CREATE TABLE statements
	createTableRegex := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([^\s(]+)\s*\((.*?)\)`)
	matches := createTableRegex.FindStringSubmatch(sql)

	if len(matches) > 0 {
		table, err := p.parseCreateTable(sql)
		if err != nil {
			return nil, err
		}
		entity.Tables = append(entity.Tables, table)
	}

	return entity, nil
}

// parseConstraint parses constraint definition
func (p *SQLiteParser) parseConstraint(constraintDef string) *sdc.Constraint {
	constraint := &sdc.Constraint{}

	// Split constraint definition into parts
	parts := strings.Fields(constraintDef)
	if len(parts) < 2 {
		return nil
	}

	// Check if it starts with CONSTRAINT keyword
	startIdx := 0
	if strings.ToUpper(parts[0]) == "CONSTRAINT" {
		if len(parts) < 3 {
			return nil
		}
		constraint.Name = strings.Trim(parts[1], "[]\"")
		startIdx = 2
	}

	// Parse constraint type and details
	constraintType := strings.ToUpper(parts[startIdx])
	switch constraintType {
	case "PRIMARY":
		if startIdx+1 < len(parts) && strings.ToUpper(parts[startIdx+1]) == "KEY" {
			constraint.Type = "PRIMARY KEY"
			// Parse columns
			if startIdx+2 < len(parts) && strings.HasPrefix(parts[startIdx+2], "(") {
				colStr := strings.Trim(parts[startIdx+2], "()")
				cols := strings.Split(colStr, ",")
				for _, col := range cols {
					constraint.Columns = append(constraint.Columns, strings.Trim(col, "[]\""))
				}
			}
		}
	case "FOREIGN":
		if startIdx+1 < len(parts) && strings.ToUpper(parts[startIdx+1]) == "KEY" {
			constraint.Type = "FOREIGN KEY"
			// Parse columns
			nextIdx := startIdx + 2
			if nextIdx < len(parts) && strings.HasPrefix(parts[nextIdx], "(") {
				colStr := strings.Trim(parts[nextIdx], "()")
				cols := strings.Split(colStr, ",")
				for _, col := range cols {
					constraint.Columns = append(constraint.Columns, strings.Trim(col, "[]\""))
				}
				nextIdx++
			}
			// Parse REFERENCES
			for i := nextIdx; i < len(parts); i++ {
				if strings.ToUpper(parts[i]) == "REFERENCES" {
					if i+1 < len(parts) {
						constraint.RefTable = strings.Trim(parts[i+1], "[]\"")
						i++
						if i+1 < len(parts) && strings.HasPrefix(parts[i+1], "(") {
							refColStr := strings.Trim(parts[i+1], "()")
							refCols := strings.Split(refColStr, ",")
							for _, col := range refCols {
								constraint.RefColumns = append(constraint.RefColumns, strings.Trim(col, "[]\""))
							}
							i++
						}
					}
					// Parse ON DELETE/UPDATE actions
					for j := i + 1; j < len(parts); j++ {
						if strings.ToUpper(parts[j]) == "ON" && j+2 < len(parts) {
							action := strings.ToUpper(parts[j+1])
							if action == "DELETE" {
								constraint.OnDelete = strings.ToUpper(parts[j+2])
								j += 2
								i = j
							} else if action == "UPDATE" {
								constraint.OnUpdate = strings.ToUpper(parts[j+2])
								j += 2
								i = j
							}
						}
					}
					break
				}
			}
		}
	case "UNIQUE":
		constraint.Type = "UNIQUE"
		// Parse columns
		if startIdx+1 < len(parts) && strings.HasPrefix(parts[startIdx+1], "(") {
			colStr := strings.Trim(parts[startIdx+1], "()")
			cols := strings.Split(colStr, ",")
			for _, col := range cols {
				constraint.Columns = append(constraint.Columns, strings.Trim(col, "[]\""))
			}
		}
	}

	return constraint
}
