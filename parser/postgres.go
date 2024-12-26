package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mstgnz/sdc"
)

// PostgresParser implements the parser for PostgreSQL database
type PostgresParser struct {
	dbInfo       DatabaseInfo
	currentTable *sdc.Table // Geçerli işlenen tablo
}

// NewPostgresParser creates a new PostgreSQL parser
func NewPostgresParser() *PostgresParser {
	return &PostgresParser{
		dbInfo: DatabaseInfoMap[PostgreSQL],
	}
}

// Parse converts PostgreSQL dump to Entity structure
func (p *PostgresParser) Parse(sql string) (*sdc.Entity, error) {
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

// parseCreateTable parses CREATE TABLE statement
func (p *PostgresParser) parseCreateTable(sql string) (*sdc.Table, error) {
	table := &sdc.Table{}

	// Extract basic table information
	tableRegex := regexp.MustCompile(`(?i)CREATE\s+(?:UNLOGGED\s+)?TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([^\s(]+)\s*\((.*?)\)(?:\s+INHERITS\s*\((.*?)\))?(?:\s+PARTITION\s+BY\s+(.*?))?(?:\s+WITH\s*\((.*?)\))?(?:\s+TABLESPACE\s+(\w+))?;?`)
	matches := tableRegex.FindStringSubmatch(sql)

	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid CREATE TABLE statement")
	}

	// Set table name
	table.Name = matches[1]

	// Parse column definitions
	columnDefs := strings.Split(matches[2], ",")
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

		// Check for LIKE definition
		if strings.HasPrefix(strings.ToUpper(columnDef), "LIKE") {
			comments = append(comments, columnDef)
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

// parseColumn parses column definition
func (p *PostgresParser) parseColumn(columnDef string) *sdc.Column {
	column := &sdc.Column{}

	// Split column name and data type
	parts := strings.Fields(columnDef)
	if len(parts) < 2 {
		return nil
	}

	column.Name = parts[0]

	// Parse data type and length/scale information
	dataTypeRegex := regexp.MustCompile(`(?i)(\w+(?:\s+\w+)?)\s*(?:\(([^)]+)\))?`)
	dataTypeMatches := dataTypeRegex.FindStringSubmatch(parts[1])

	if len(dataTypeMatches) > 1 {
		dataType := &sdc.DataType{
			Name: strings.ToUpper(dataTypeMatches[1]),
		}

		// Parse length and scale information
		if len(dataTypeMatches) > 2 && dataTypeMatches[2] != "" {
			precisionScale := strings.Split(dataTypeMatches[2], ",")
			if len(precisionScale) > 0 {
				if strings.Contains(dataType.Name, "NUMERIC") || strings.Contains(dataType.Name, "DECIMAL") {
					dataType.Precision = parseInt(precisionScale[0])
					if len(precisionScale) > 1 {
						dataType.Scale = parseInt(precisionScale[1])
					}
				} else {
					dataType.Length = parseInt(precisionScale[0])
				}
			}
		}

		column.DataType = dataType
	}

	// Check for NOT NULL constraint
	if strings.Contains(strings.ToUpper(columnDef), "NOT NULL") {
		column.IsNullable = false
	} else {
		column.IsNullable = true
	}

	// Check for DEFAULT value
	defaultRegex := regexp.MustCompile(`(?i)DEFAULT\s+([^,\s]+)`)
	defaultMatches := defaultRegex.FindStringSubmatch(columnDef)
	if len(defaultMatches) > 1 {
		column.Default = defaultMatches[1]
	}

	// Check for COLLATE definition
	collateRegex := regexp.MustCompile(`(?i)COLLATE\s+"([^"]+)"`)
	collateMatches := collateRegex.FindStringSubmatch(columnDef)
	if len(collateMatches) > 1 {
		column.Collation = collateMatches[1]
	}

	return column
}

// Convert transforms Entity structure to PostgreSQL format
func (p *PostgresParser) Convert(entity *sdc.Entity) (string, error) {
	var result strings.Builder

	for _, table := range entity.Tables {
		// CREATE TABLE statement
		result.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", table.Name))

		// Columns
		for i, column := range table.Columns {
			result.WriteString(fmt.Sprintf("    %s %s", column.Name, p.convertDataType(column.DataType)))

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

// convertDataType converts data type to PostgreSQL format
func (p *PostgresParser) convertDataType(dataType *sdc.DataType) string {
	if dataType == nil {
		return "text"
	}

	// Sütun sayısı kontrolü
	if p.dbInfo.MaxColumns > 0 && p.currentTable != nil && len(p.currentTable.Columns) >= p.dbInfo.MaxColumns {
		// Hata durumunda varsayılan tip dönüyoruz
		return "text"
	}

	switch strings.ToUpper(dataType.Name) {
	case "VARCHAR", "NVARCHAR", "CHAR", "NCHAR":
		if dataType.Length > 0 {
			return fmt.Sprintf("%s(%d)", strings.ToLower(dataType.Name), dataType.Length)
		}
		return "text"
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
	case "FLOAT", "REAL":
		return "real"
	case "DOUBLE", "DOUBLE PRECISION":
		return "double precision"
	case "BOOLEAN", "BIT":
		return "boolean"
	case "DATE":
		return "date"
	case "TIME":
		return "time"
	case "TIMESTAMP":
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
