package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mstgnz/sdc"
)

// Precompiled regex patterns for better performance
var (
	postgresCreateTableRegex = regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:([^\s.]+)\.)?([^\s(]+)\s*\((.*)\)`)
	postgresColumnRegex      = regexp.MustCompile(`(?i)^\s*"?([^"\s(]+)"?\s+([^(,\s]+)(?:\s*\(([^)]+)\))?\s*(.*)$`)
	postgresDefaultRegex     = regexp.MustCompile(`(?i)DEFAULT\s+([^,\s]+|'[^']*'|"[^"]*")`)
	postgresSerialRegex      = regexp.MustCompile(`(?i)^(SMALL|BIG)?SERIAL$`)
)

// PostgresParser implements the parser for PostgreSQL database
type PostgresParser struct {
	dbInfo DatabaseInfo
	// Add buffer pool for string builders to reduce allocations
	builderPool strings.Builder
}

// NewPostgresParser creates a new PostgreSQL parser
func NewPostgresParser() *PostgresParser {
	return &PostgresParser{
		dbInfo: DatabaseInfoMap[PostgreSQL],
	}
}

// Parse converts PostgreSQL dump to Entity structure
func (p *PostgresParser) Parse(sql string) (*sdc.Entity, error) {
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

// Convert transforms Entity structure to PostgreSQL format with optimized string handling
func (p *PostgresParser) Convert(entity *sdc.Entity) (string, error) {
	// Reset and reuse string builder
	p.builderPool.Reset()
	result := &p.builderPool

	for _, table := range entity.Tables {
		// CREATE TABLE statement
		result.WriteString("CREATE TABLE ")
		if table.Schema != "" {
			result.WriteString(table.Schema)
			result.WriteString(".")
		}
		result.WriteString(table.Name)
		result.WriteString(" (\n")

		// Columns
		for i, column := range table.Columns {
			result.WriteString("\t")
			result.WriteString(column.Name)
			result.WriteString(" ")
			result.WriteString(p.convertDataType(column.DataType))

			if column.Collation != "" {
				result.WriteString(" COLLATE ")
				result.WriteString(column.Collation)
			}

			if !column.IsNullable {
				result.WriteString(" NOT NULL")
			}

			if column.Default != "" {
				result.WriteString(" DEFAULT ")
				result.WriteString(column.Default)
			}

			if column.AutoIncrement {
				result.WriteString(" GENERATED ALWAYS AS IDENTITY")
			}

			if i < len(table.Columns)-1 {
				result.WriteString(",\n")
			}
		}

		// Add primary key constraint if exists
		if table.PrimaryKey != nil {
			pkColumns := table.PrimaryKey.Columns
			if len(pkColumns) > 0 {
				result.WriteString(",\n\tPRIMARY KEY (")
				result.WriteString(strings.Join(pkColumns, ", "))
				result.WriteString(")")
			}
		}

		// Add foreign key constraints if exist
		if len(table.ForeignKeys) > 0 {
			for _, fk := range table.ForeignKeys {
				result.WriteString(",\n\tFOREIGN KEY (")
				result.WriteString(strings.Join(fk.Columns, ", "))
				result.WriteString(") REFERENCES ")
				result.WriteString(fk.RefTable)
				result.WriteString(" (")
				result.WriteString(strings.Join(fk.RefColumns, ", "))
				result.WriteString(")")
				if fk.OnDelete != "" {
					result.WriteString(" ON DELETE ")
					result.WriteString(fk.OnDelete)
				}
				if fk.OnUpdate != "" {
					result.WriteString(" ON UPDATE ")
					result.WriteString(fk.OnUpdate)
				}
			}
		}

		result.WriteString("\n);\n\n")
	}

	return result.String(), nil
}

// parseCreateTable parses CREATE TABLE statement with optimized string handling
func (p *PostgresParser) parseCreateTable(sql string) (*sdc.Table, error) {
	matches := postgresCreateTableRegex.FindStringSubmatch(sql)
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid CREATE TABLE statement")
	}

	table := &sdc.Table{
		Schema:      strings.Trim(matches[1], "\""),
		Name:        strings.Trim(matches[2], "\""),
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

// parseColumn parses column definition with optimized string handling
func (p *PostgresParser) parseColumn(columnDef string) *sdc.Column {
	matches := postgresColumnRegex.FindStringSubmatch(columnDef)
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

	// Check for SERIAL types
	if postgresSerialRegex.MatchString(dataType.Name) {
		column.AutoIncrement = true
		// Convert SERIAL to INTEGER
		switch strings.ToUpper(dataType.Name) {
		case "SMALLSERIAL":
			dataType.Name = "SMALLINT"
		case "SERIAL":
			dataType.Name = "INTEGER"
		case "BIGSERIAL":
			dataType.Name = "BIGINT"
		}
	}

	// Parse length/precision/scale
	if matches[3] != "" {
		parts := strings.Split(matches[3], ",")
		if len(parts) > 0 {
			if dataType.Name == "numeric" || dataType.Name == "decimal" {
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
		if defaultMatches := postgresDefaultRegex.FindStringSubmatch(rest); defaultMatches != nil {
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

// parseConstraint parses constraint definition
func (p *PostgresParser) parseConstraint(constraintDef string) *sdc.Constraint {
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
		constraint.Name = strings.Trim(parts[1], "\"")
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
					constraint.Columns = append(constraint.Columns, strings.Trim(col, "\""))
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
					constraint.Columns = append(constraint.Columns, strings.Trim(col, "\""))
				}
				nextIdx++
			}
			// Parse REFERENCES
			for i := nextIdx; i < len(parts); i++ {
				if strings.ToUpper(parts[i]) == "REFERENCES" {
					if i+1 < len(parts) {
						constraint.RefTable = strings.Trim(parts[i+1], "\"")
						i++
						if i+1 < len(parts) && strings.HasPrefix(parts[i+1], "(") {
							refColStr := strings.Trim(parts[i+1], "()")
							refCols := strings.Split(refColStr, ",")
							for _, col := range refCols {
								constraint.RefColumns = append(constraint.RefColumns, strings.Trim(col, "\""))
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
				constraint.Columns = append(constraint.Columns, strings.Trim(col, "\""))
			}
		}
	}

	return constraint
}

// convertDataType converts data type to PostgreSQL format
func (p *PostgresParser) convertDataType(dataType *sdc.DataType) string {
	if dataType == nil {
		return "text"
	}

	switch strings.ToUpper(dataType.Name) {
	case "VARCHAR", "NVARCHAR":
		if dataType.Length > 0 {
			return fmt.Sprintf("varchar(%d)", dataType.Length)
		}
		return "text"
	case "CHAR", "NCHAR":
		if dataType.Length > 0 {
			return fmt.Sprintf("char(%d)", dataType.Length)
		}
		return "char"
	case "INT", "INTEGER":
		return "integer"
	case "BIGINT":
		return "bigint"
	case "SMALLINT":
		return "smallint"
	case "DECIMAL", "NUMERIC":
		if dataType.Precision > 0 {
			if dataType.Scale > 0 {
				return fmt.Sprintf("numeric(%d,%d)", dataType.Precision, dataType.Scale)
			}
			return fmt.Sprintf("numeric(%d)", dataType.Precision)
		}
		return "numeric"
	case "FLOAT":
		if dataType.Precision > 0 {
			return fmt.Sprintf("float(%d)", dataType.Precision)
		}
		return "float8"
	case "REAL":
		return "float4"
	case "BOOLEAN", "BIT":
		return "boolean"
	case "DATE":
		return "date"
	case "TIME":
		return "time"
	case "DATETIME", "TIMESTAMP":
		return "timestamp"
	case "TEXT", "NTEXT", "CLOB":
		return "text"
	case "BLOB", "BINARY", "VARBINARY":
		return "bytea"
	default:
		return strings.ToLower(dataType.Name)
	}
}

// ParseCreateTable parses CREATE TABLE statement
func (p *PostgresParser) ParseCreateTable(sql string) (*sdc.Table, error) {
	return p.parseCreateTable(sql)
}

// ParseAlterTable parses ALTER TABLE statement
func (p *PostgresParser) ParseAlterTable(sql string) (*sdc.AlterTable, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseDropTable parses DROP TABLE statement
func (p *PostgresParser) ParseDropTable(sql string) (*sdc.DropTable, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseCreateIndex parses CREATE INDEX statement
func (p *PostgresParser) ParseCreateIndex(sql string) (*sdc.Index, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseDropIndex parses DROP INDEX statement
func (p *PostgresParser) ParseDropIndex(sql string) (*sdc.DropIndex, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ValidateIdentifier validates the identifier name
func (p *PostgresParser) ValidateIdentifier(name string) error {
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
func (p *PostgresParser) EscapeIdentifier(name string) string {
	// Escape logic to be implemented
	return name
}

// EscapeString escapes the string value
func (p *PostgresParser) EscapeString(value string) string {
	// Escape logic to be implemented
	return value
}

// GetDefaultSchema returns the default schema name
func (p *PostgresParser) GetDefaultSchema() string {
	return p.dbInfo.DefaultSchema
}

// GetSchemaPrefix returns the schema prefix
func (p *PostgresParser) GetSchemaPrefix(schema string) string {
	if schema == "" || schema == p.dbInfo.DefaultSchema {
		return ""
	}
	return schema + "."
}

// GetIdentifierQuote returns the identifier quote character
func (p *PostgresParser) GetIdentifierQuote() string {
	return p.dbInfo.IdentifierQuote
}

// GetStringQuote returns the string quote character
func (p *PostgresParser) GetStringQuote() string {
	return p.dbInfo.StringQuote
}

// GetMaxIdentifierLength returns the maximum identifier length
func (p *PostgresParser) GetMaxIdentifierLength() int {
	return p.dbInfo.MaxIdentifierLength
}

// GetReservedWords returns the reserved words
func (p *PostgresParser) GetReservedWords() []string {
	return p.dbInfo.ReservedWords
}

// ConvertDataTypeFrom converts source database data type to PostgreSQL data type
func (p *PostgresParser) ConvertDataTypeFrom(sourceType string, length int, precision int, scale int) *sdc.DataType {
	return &sdc.DataType{
		Name:      sourceType,
		Length:    length,
		Precision: precision,
		Scale:     scale,
	}
}

// ParseSQL parses PostgreSQL SQL statements and returns an Entity
func (p *PostgresParser) ParseSQL(sql string) (*sdc.Entity, error) {
	entity := &sdc.Entity{}

	// Find CREATE TABLE statements
	createTableRegex := regexp.MustCompile(`(?i)CREATE\s+(?:UNLOGGED\s+)?TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([^\s(]+)\s*\((.*?)\)(?:\s+INHERITS\s*\((.*?)\))?(?:\s+PARTITION\s+BY\s+(.*?))?(?:\s+WITH\s*\((.*?)\))?(?:\s+TABLESPACE\s+(\w+))?;?`)
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
