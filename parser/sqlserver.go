package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mstgnz/sdc"
)

// SQLServerParser implements the parser for SQL Server database
type SQLServerParser struct {
	dbInfo       DatabaseInfo
	currentTable *sdc.Table // Geçerli işlenen tablo
}

// NewSQLServerParser creates a new SQL Server parser
func NewSQLServerParser() *SQLServerParser {
	return &SQLServerParser{
		dbInfo: DatabaseInfoMap[SQLServer],
	}
}

// Parse converts SQL Server dump to Entity structure
func (p *SQLServerParser) Parse(sql string) (*sdc.Entity, error) {
	entity := &sdc.Entity{}

	// Find CREATE TABLE statements
	createTableRegex := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:\[([^\]]+)\]|\[?(\w+)\]?)\s*\((.*?)\)(?:\s+ON\s+\[?(\w+)\]?)?(?:\s+TEXTIMAGE_ON\s+\[?(\w+)\]?)?`)
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
func (p *SQLServerParser) parseCreateTable(sql string) (*sdc.Table, error) {
	table := &sdc.Table{}

	// Extract basic table information
	tableRegex := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:\[([^\]]+)\]|\[?(\w+)\]?)\s*\((.*?)\)(?:\s+ON\s+\[?(\w+)\]?)?(?:\s+TEXTIMAGE_ON\s+\[?(\w+)\]?)?`)
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

		// Parse column definition
		column := p.parseColumn(columnDef)
		if column != nil {
			table.Columns = append(table.Columns, column)
		}
	}

	// Set FileGroup if specified
	if len(matches) > 4 && matches[4] != "" {
		table.FileGroup = matches[4]
	}

	if len(comments) > 0 {
		table.Comment = strings.Join(comments, ", ")
	}

	return table, nil
}

// parseColumn parses column definition
func (p *SQLServerParser) parseColumn(columnDef string) *sdc.Column {
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
				if strings.Contains(dataType.Name, "DECIMAL") || strings.Contains(dataType.Name, "NUMERIC") {
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

	// Check for IDENTITY
	if strings.Contains(strings.ToUpper(columnDef), "IDENTITY") {
		column.Identity = true
		identityRegex := regexp.MustCompile(`IDENTITY\s*\((\d+),\s*(\d+)\)`)
		identityMatches := identityRegex.FindStringSubmatch(columnDef)
		if len(identityMatches) > 2 {
			column.IdentitySeed = parseInt64(identityMatches[1])
			column.IdentityIncr = parseInt64(identityMatches[2])
		}
	}

	// Check for COLLATE definition
	collateRegex := regexp.MustCompile(`(?i)COLLATE\s+([^\s,]+)`)
	collateMatches := collateRegex.FindStringSubmatch(columnDef)
	if len(collateMatches) > 1 {
		column.Collation = collateMatches[1]
	}

	// Check for SPARSE
	if strings.Contains(strings.ToUpper(columnDef), "SPARSE") {
		column.Sparse = true
	}

	// Check for FILESTREAM
	if strings.Contains(strings.ToUpper(columnDef), "FILESTREAM") {
		column.FileStream = true
	}

	// Check for ROWGUIDCOL
	if strings.Contains(strings.ToUpper(columnDef), "ROWGUIDCOL") {
		column.RowGUIDCol = true
	}

	return column
}

// parseInt64 safely converts string to int64
func parseInt64(s string) int64 {
	s = strings.TrimSpace(s)
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}

// Convert transforms Entity structure to SQL Server format
func (p *SQLServerParser) Convert(entity *sdc.Entity) (string, error) {
	var result strings.Builder

	for _, table := range entity.Tables {
		// CREATE TABLE statement
		result.WriteString(fmt.Sprintf("CREATE TABLE [%s] (\n", table.Name))

		// Columns
		for i, column := range table.Columns {
			result.WriteString(fmt.Sprintf("    [%s] %s", column.Name, p.convertDataType(column.DataType)))

			if column.Collation != "" {
				result.WriteString(fmt.Sprintf(" COLLATE %s", column.Collation))
			}

			if !column.IsNullable {
				result.WriteString(" NOT NULL")
			}

			if column.Default != "" {
				result.WriteString(fmt.Sprintf(" DEFAULT %s", column.Default))
			}

			if column.Identity {
				if column.IdentitySeed != 0 || column.IdentityIncr != 0 {
					result.WriteString(fmt.Sprintf(" IDENTITY(%d,%d)", column.IdentitySeed, column.IdentityIncr))
				} else {
					result.WriteString(" IDENTITY")
				}
			}

			if column.Sparse {
				result.WriteString(" SPARSE")
			}

			if column.FileStream {
				result.WriteString(" FILESTREAM")
			}

			if column.RowGUIDCol {
				result.WriteString(" ROWGUIDCOL")
			}

			if i < len(table.Columns)-1 {
				result.WriteString(",\n")
			}
		}

		// FileGroup
		if table.FileGroup != "" {
			result.WriteString(fmt.Sprintf(") ON [%s]", table.FileGroup))
		} else {
			result.WriteString(")")
		}

		result.WriteString("\n\n")
	}

	return result.String(), nil
}

// convertDataType converts data type to SQL Server format
func (p *SQLServerParser) convertDataType(dataType *sdc.DataType) string {
	if dataType == nil {
		return "varchar(max)"
	}

	// Sütun sayısı kontrolü
	if p.dbInfo.MaxColumns > 0 && p.currentTable != nil && len(p.currentTable.Columns) >= p.dbInfo.MaxColumns {
		// Hata durumunda varsayılan tip dönüyoruz
		return "varchar(max)"
	}

	switch strings.ToUpper(dataType.Name) {
	case "VARCHAR", "NVARCHAR":
		if dataType.Length > 0 {
			return fmt.Sprintf("%s(%d)", strings.ToUpper(dataType.Name), dataType.Length)
		}
		return dataType.Name + "(max)"
	case "CHAR", "NCHAR":
		if dataType.Length > 0 {
			return fmt.Sprintf("%s(%d)", strings.ToUpper(dataType.Name), dataType.Length)
		}
		return dataType.Name
	case "INT", "INTEGER":
		return "int"
	case "BIGINT":
		return "bigint"
	case "SMALLINT":
		return "smallint"
	case "DECIMAL", "NUMERIC":
		if dataType.Precision > 0 {
			if dataType.Scale > 0 {
				return fmt.Sprintf("%s(%d,%d)", strings.ToUpper(dataType.Name), dataType.Precision, dataType.Scale)
			}
			return fmt.Sprintf("%s(%d)", strings.ToUpper(dataType.Name), dataType.Precision)
		}
		return strings.ToUpper(dataType.Name)
	case "FLOAT":
		if dataType.Precision > 0 {
			return fmt.Sprintf("float(%d)", dataType.Precision)
		}
		return "float"
	case "REAL":
		return "real"
	case "BOOLEAN", "BIT":
		return "bit"
	case "DATE":
		return "date"
	case "TIME":
		return "time"
	case "DATETIME", "TIMESTAMP":
		return "datetime2"
	case "TEXT", "NTEXT", "CLOB":
		return "varchar(max)"
	case "BLOB", "BINARY", "VARBINARY":
		if dataType.Length > 0 {
			return fmt.Sprintf("varbinary(%d)", dataType.Length)
		}
		return "varbinary(max)"
	default:
		return strings.ToUpper(dataType.Name)
	}
}

// GetDefaultSchema returns the default schema name
func (p *SQLServerParser) GetDefaultSchema() string {
	return p.dbInfo.DefaultSchema
}

// GetSchemaPrefix returns the schema prefix
func (p *SQLServerParser) GetSchemaPrefix(schema string) string {
	if schema == "" || schema == p.dbInfo.DefaultSchema {
		return ""
	}
	return schema + "."
}

// GetIdentifierQuote returns the identifier quote character
func (p *SQLServerParser) GetIdentifierQuote() string {
	return p.dbInfo.IdentifierQuote
}

// GetStringQuote returns the string quote character
func (p *SQLServerParser) GetStringQuote() string {
	return p.dbInfo.StringQuote
}

// GetMaxIdentifierLength returns the maximum identifier length
func (p *SQLServerParser) GetMaxIdentifierLength() int {
	return p.dbInfo.MaxIdentifierLength
}

// GetReservedWords returns the reserved words
func (p *SQLServerParser) GetReservedWords() []string {
	return p.dbInfo.ReservedWords
}

// ConvertDataTypeFrom converts source database data type to SQL Server data type
func (p *SQLServerParser) ConvertDataTypeFrom(sourceType string, length int, precision int, scale int) *sdc.DataType {
	return &sdc.DataType{
		Name:      sourceType,
		Length:    length,
		Precision: precision,
		Scale:     scale,
	}
}

// ParseCreateTable parses CREATE TABLE statement
func (p *SQLServerParser) ParseCreateTable(sql string) (*sdc.Table, error) {
	// Remove comments and extra whitespace
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "CREATE TABLE") {
		return nil, fmt.Errorf("invalid CREATE TABLE statement")
	}

	// Extract table name and schema
	tableRegex := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:\[?(\w+)\]?\.)?(?:\[([^\]]+)\]|\[?(\w+)\]?)\s*\((.*?)\)(?:\s+ON\s+\[?(\w+)\]?)?`)
	matches := tableRegex.FindStringSubmatch(sql)
	if len(matches) < 5 {
		return nil, fmt.Errorf("could not parse CREATE TABLE statement")
	}

	table := &sdc.Table{
		Schema:    matches[1],
		Name:      matches[2],
		FileGroup: matches[4],
	}

	if table.Schema == "" {
		table.Schema = "dbo"
	}

	// Parse column definitions
	columnDefs := strings.Split(matches[3], ",")
	for _, columnDef := range columnDefs {
		columnDef = strings.TrimSpace(columnDef)
		if columnDef == "" {
			continue
		}

		// Parse column
		if !strings.HasPrefix(strings.ToUpper(columnDef), "CONSTRAINT") {
			column := p.parseColumn(columnDef)
			if column != nil {
				table.Columns = append(table.Columns, column)
			}
		} else {
			// Parse constraint
			constraint := p.parseConstraint(columnDef)
			if constraint != nil {
				table.Constraints = append(table.Constraints, constraint)
			}
		}
	}

	return table, nil
}

// ParseAlterTable parses ALTER TABLE statement
func (p *SQLServerParser) ParseAlterTable(sql string) (*sdc.AlterTable, error) {
	// Remove comments and extra whitespace
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "ALTER TABLE") {
		return nil, fmt.Errorf("invalid ALTER TABLE statement")
	}

	alterTable := &sdc.AlterTable{}

	// Extract table name and schema
	tableRegex := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+(?:\[?(\w+)\]?\.)?(?:\[([^\]]+)\]|\[?(\w+)\]?)\s+(.*)`)
	matches := tableRegex.FindStringSubmatch(sql)
	if len(matches) < 5 {
		return nil, fmt.Errorf("could not parse ALTER TABLE statement")
	}

	alterTable.Schema = matches[1]
	alterTable.Table = matches[2]
	if alterTable.Schema == "" {
		alterTable.Schema = "dbo"
	}

	// Parse action
	action := strings.TrimSpace(matches[4])
	if strings.HasPrefix(strings.ToUpper(action), "ADD") {
		alterTable.Action = "ADD"
		if strings.HasPrefix(strings.ToUpper(action), "ADD CONSTRAINT") {
			alterTable.Action = "ADD CONSTRAINT"
			alterTable.Constraint = p.parseConstraint(action[14:])
		} else {
			alterTable.Column = p.parseColumn(action[4:])
		}
	} else if strings.HasPrefix(strings.ToUpper(action), "DROP") {
		alterTable.Action = "DROP"
		if strings.HasPrefix(strings.ToUpper(action), "DROP CONSTRAINT") {
			alterTable.Action = "DROP CONSTRAINT"
			alterTable.Constraint = &sdc.Constraint{Name: strings.TrimSpace(action[15:])}
		} else {
			alterTable.Column = &sdc.Column{Name: strings.TrimSpace(action[5:])}
		}
	}

	return alterTable, nil
}

// ParseDropTable parses DROP TABLE statement
func (p *SQLServerParser) ParseDropTable(sql string) (*sdc.DropTable, error) {
	// Remove comments and extra whitespace
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "DROP TABLE") {
		return nil, fmt.Errorf("invalid DROP TABLE statement")
	}

	dropTable := &sdc.DropTable{}

	// Extract table name and schema
	tableRegex := regexp.MustCompile(`(?i)DROP\s+TABLE\s+(?:IF\s+EXISTS\s+)?(?:\[?(\w+)\]?\.)?(?:\[([^\]]+)\]|\[?(\w+)\]?)`)
	matches := tableRegex.FindStringSubmatch(sql)
	if len(matches) < 3 {
		return nil, fmt.Errorf("could not parse DROP TABLE statement")
	}

	dropTable.Schema = matches[1]
	dropTable.Table = matches[2]
	if dropTable.Schema == "" {
		dropTable.Schema = "dbo"
	}
	dropTable.IfExists = strings.Contains(strings.ToUpper(sql), "IF EXISTS")

	return dropTable, nil
}

// ParseCreateIndex parses CREATE INDEX statement
func (p *SQLServerParser) ParseCreateIndex(sql string) (*sdc.Index, error) {
	// Remove comments and extra whitespace
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "CREATE") {
		return nil, fmt.Errorf("invalid CREATE INDEX statement")
	}

	index := &sdc.Index{}

	// Extract index information
	var indexRegex *regexp.Regexp
	if strings.Contains(strings.ToUpper(sql), "UNIQUE") {
		indexRegex = regexp.MustCompile(`(?i)CREATE\s+UNIQUE\s+(?:CLUSTERED\s+)?INDEX\s+\[?(\w+)\]?\s+ON\s+(?:\[?(\w+)\]?\.)?(?:\[?(\w+)\]?)\s*\((.*?)\)(?:\s+INCLUDE\s*\((.*?)\))?(?:\s+WHERE\s+(.*))?(?:\s+WITH\s*\((.*?)\))?(?:\s+ON\s+\[?(\w+)\]?)?`)
		index.Unique = true
	} else {
		indexRegex = regexp.MustCompile(`(?i)CREATE\s+(?:CLUSTERED\s+)?INDEX\s+\[?(\w+)\]?\s+ON\s+(?:\[?(\w+)\]?\.)?(?:\[?(\w+)\]?)\s*\((.*?)\)(?:\s+INCLUDE\s*\((.*?)\))?(?:\s+WHERE\s+(.*))?(?:\s+WITH\s*\((.*?)\))?(?:\s+ON\s+\[?(\w+)\]?)?`)
	}

	matches := indexRegex.FindStringSubmatch(sql)
	if len(matches) < 5 {
		return nil, fmt.Errorf("could not parse CREATE INDEX statement")
	}

	index.Name = matches[1]
	index.Schema = matches[2]
	index.Table = matches[3]
	if index.Schema == "" {
		index.Schema = "dbo"
	}

	// Parse columns
	columnList := strings.Split(matches[4], ",")
	for _, col := range columnList {
		col = strings.TrimSpace(col)
		if col != "" {
			index.Columns = append(index.Columns, col)
		}
	}

	// Parse included columns
	if len(matches) > 5 && matches[5] != "" {
		includeList := strings.Split(matches[5], ",")
		for _, col := range includeList {
			col = strings.TrimSpace(col)
			if col != "" {
				index.IncludeColumns = append(index.IncludeColumns, col)
			}
		}
	}

	// Parse filter
	if len(matches) > 6 && matches[6] != "" {
		index.Filter = matches[6]
	}

	// Parse filegroup
	if len(matches) > 8 && matches[8] != "" {
		index.FileGroup = matches[8]
	}

	// Check for CLUSTERED
	if strings.Contains(strings.ToUpper(sql), "CLUSTERED") {
		index.Clustered = true
	}

	return index, nil
}

// ParseDropIndex parses DROP INDEX statement
func (p *SQLServerParser) ParseDropIndex(sql string) (*sdc.DropIndex, error) {
	// Remove comments and extra whitespace
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "DROP INDEX") {
		return nil, fmt.Errorf("invalid DROP INDEX statement")
	}

	dropIndex := &sdc.DropIndex{}

	// Extract index information
	indexRegex := regexp.MustCompile(`(?i)DROP\s+INDEX\s+(?:IF\s+EXISTS\s+)?\[?(\w+)\]?\s+ON\s+(?:\[?(\w+)\]?\.)?(?:\[?(\w+)\]?)`)
	matches := indexRegex.FindStringSubmatch(sql)
	if len(matches) < 4 {
		return nil, fmt.Errorf("could not parse DROP INDEX statement")
	}

	dropIndex.Index = matches[1]
	dropIndex.Schema = matches[2]
	dropIndex.Table = matches[3]
	if dropIndex.Schema == "" {
		dropIndex.Schema = "dbo"
	}
	dropIndex.IfExists = strings.Contains(strings.ToUpper(sql), "IF EXISTS")

	return dropIndex, nil
}

// parseConstraint parses constraint definition
func (p *SQLServerParser) parseConstraint(constraintDef string) *sdc.Constraint {
	constraint := &sdc.Constraint{}

	// Extract constraint name
	nameRegex := regexp.MustCompile(`(?i)CONSTRAINT\s+\[?(\w+)\]?\s+(.*)`)
	matches := nameRegex.FindStringSubmatch(constraintDef)
	if len(matches) < 3 {
		return nil
	}

	constraint.Name = matches[1]
	constraintDef = matches[2]

	// Determine constraint type and parse accordingly
	if strings.HasPrefix(strings.ToUpper(constraintDef), "PRIMARY KEY") {
		constraint.Type = "PRIMARY KEY"
		columnsRegex := regexp.MustCompile(`\((.*?)\)`)
		columnsMatch := columnsRegex.FindStringSubmatch(constraintDef)
		if len(columnsMatch) > 1 {
			columns := strings.Split(columnsMatch[1], ",")
			for _, col := range columns {
				constraint.Columns = append(constraint.Columns, strings.TrimSpace(col))
			}
		}
	} else if strings.HasPrefix(strings.ToUpper(constraintDef), "FOREIGN KEY") {
		constraint.Type = "FOREIGN KEY"
		fkRegex := regexp.MustCompile(`FOREIGN\s+KEY\s*\((.*?)\)\s*REFERENCES\s+(\w+)\s*\((.*?)\)(?:\s+ON\s+DELETE\s+(\w+(?:\s+\w+)?))?(?:\s+ON\s+UPDATE\s+(\w+(?:\s+\w+)?))?`)
		fkMatch := fkRegex.FindStringSubmatch(constraintDef)
		if len(fkMatch) > 3 {
			columns := strings.Split(fkMatch[1], ",")
			for _, col := range columns {
				constraint.Columns = append(constraint.Columns, strings.TrimSpace(col))
			}
			constraint.RefTable = fkMatch[2]
			refColumns := strings.Split(fkMatch[3], ",")
			for _, col := range refColumns {
				constraint.RefColumns = append(constraint.RefColumns, strings.TrimSpace(col))
			}
			if len(fkMatch) > 4 && fkMatch[4] != "" {
				constraint.OnDelete = fkMatch[4]
			}
			if len(fkMatch) > 5 && fkMatch[5] != "" {
				constraint.OnUpdate = fkMatch[5]
			}
		}
	} else if strings.HasPrefix(strings.ToUpper(constraintDef), "UNIQUE") {
		constraint.Type = "UNIQUE"
		columnsRegex := regexp.MustCompile(`\((.*?)\)`)
		columnsMatch := columnsRegex.FindStringSubmatch(constraintDef)
		if len(columnsMatch) > 1 {
			columns := strings.Split(columnsMatch[1], ",")
			for _, col := range columns {
				constraint.Columns = append(constraint.Columns, strings.TrimSpace(col))
			}
		}
	} else if strings.HasPrefix(strings.ToUpper(constraintDef), "CHECK") {
		constraint.Type = "CHECK"
		checkRegex := regexp.MustCompile(`CHECK\s*\((.*?)\)`)
		checkMatch := checkRegex.FindStringSubmatch(constraintDef)
		if len(checkMatch) > 1 {
			constraint.Check = checkMatch[1]
		}
	}

	return constraint
}

// ValidateIdentifier validates the identifier name
func (p *SQLServerParser) ValidateIdentifier(name string) error {
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
func (p *SQLServerParser) EscapeIdentifier(name string) string {
	// Escape logic to be implemented
	return name
}

// EscapeString escapes the string value
func (p *SQLServerParser) EscapeString(value string) string {
	// Escape logic to be implemented
	return value
}

// ParseSQL parses SQL Server SQL statements and returns an Entity
func (p *SQLServerParser) ParseSQL(sql string) (*sdc.Entity, error) {
	entity := &sdc.Entity{}

	// Find CREATE TABLE statements
	createTableRegex := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:\[([^\]]+)\]|\[?(\w+)\]?)\s*\((.*?)\)(?:\s+ON\s+\[?(\w+)\]?)?(?:\s+TEXTIMAGE_ON\s+\[?(\w+)\]?)?`)
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
