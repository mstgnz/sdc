package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mstgnz/sdc"
)

// SQLiteParser implements the parser for SQLite database
type SQLiteParser struct {
	dbInfo       DatabaseInfo
	currentTable *sdc.Table // Geçerli işlenen tablo
}

// NewSQLiteParser creates a new SQLite parser
func NewSQLiteParser() *SQLiteParser {
	return &SQLiteParser{
		dbInfo: DatabaseInfoMap[SQLite],
	}
}

// Parse converts SQLite dump to Entity structure
func (p *SQLiteParser) Parse(sql string) (*sdc.Entity, error) {
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

// parseCreateTable parses CREATE TABLE statement
func (p *SQLiteParser) parseCreateTable(sql string) (*sdc.Table, error) {
	// Remove comments and extra whitespace
	sql = removeComments(sql)
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "CREATE TABLE") {
		return nil, fmt.Errorf("invalid CREATE TABLE statement")
	}

	// Extract table name
	tableNameRegex := regexp.MustCompile(`CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([^\s(]+)\s*\(`)
	matches := tableNameRegex.FindStringSubmatch(sql)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not extract table name")
	}

	table := &sdc.Table{
		Name: matches[1],
	}

	// Extract column definitions
	columnDefsRegex := regexp.MustCompile(`\((.*)\)`)
	matches = columnDefsRegex.FindStringSubmatch(sql)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not extract column definitions")
	}

	// Parse column definitions
	columnDefs := strings.Split(matches[1], ",")
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

// parseColumn parses column definition
func (p *SQLiteParser) parseColumn(columnDef string) *sdc.Column {
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

	// Check for PRIMARY KEY
	if strings.Contains(strings.ToUpper(columnDef), "PRIMARY KEY") {
		column.PrimaryKey = true
		if strings.Contains(strings.ToUpper(columnDef), "AUTOINCREMENT") {
			column.AutoIncrement = true
		}
	}

	// Check for UNIQUE
	if strings.Contains(strings.ToUpper(columnDef), "UNIQUE") {
		column.Unique = true
	}

	return column
}

// removeComments removes SQL comments from the input string
func removeComments(sql string) string {
	// Remove inline comments (--...)
	inlineCommentRegex := regexp.MustCompile(`--.*$`)
	sql = inlineCommentRegex.ReplaceAllString(sql, "")

	// Remove multi-line comments (/* ... */)
	multiLineCommentRegex := regexp.MustCompile(`/\*.*?\*/`)
	sql = multiLineCommentRegex.ReplaceAllString(sql, "")

	return sql
}

// Convert transforms Entity structure to SQLite format
func (p *SQLiteParser) Convert(entity *sdc.Entity) (string, error) {
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

	// Sütun sayısı kontrolü
	if p.dbInfo.MaxColumns > 0 && p.currentTable != nil && len(p.currentTable.Columns) >= p.dbInfo.MaxColumns {
		// Hata durumunda varsayılan tip dönüyoruz
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
		return nil, fmt.Errorf("invalid CREATE TABLE statement")
	}

	// Extract table name
	tableNameRegex := regexp.MustCompile(`CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([^\s(]+)\s*\(`)
	matches := tableNameRegex.FindStringSubmatch(sql)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not extract table name")
	}

	table := &sdc.Table{
		Name: matches[1],
	}

	// Extract column definitions
	columnDefsRegex := regexp.MustCompile(`\((.*)\)`)
	matches = columnDefsRegex.FindStringSubmatch(sql)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not extract column definitions")
	}

	// Parse column definitions
	columnDefs := strings.Split(matches[1], ",")
	for _, columnDef := range columnDefs {
		columnDef = strings.TrimSpace(columnDef)
		if columnDef == "" {
			continue
		}

		// Check for constraints
		if strings.HasPrefix(strings.ToUpper(columnDef), "CONSTRAINT") {
			constraint := p.parseConstraint(columnDef)
			if constraint != nil {
				table.Constraints = append(table.Constraints, constraint)
			}
			continue
		}

		// Check for table-level constraints
		if strings.HasPrefix(strings.ToUpper(columnDef), "PRIMARY KEY") ||
			strings.HasPrefix(strings.ToUpper(columnDef), "FOREIGN KEY") ||
			strings.HasPrefix(strings.ToUpper(columnDef), "UNIQUE") ||
			strings.HasPrefix(strings.ToUpper(columnDef), "CHECK") {
			constraint := p.parseConstraint(columnDef)
			if constraint != nil {
				table.Constraints = append(table.Constraints, constraint)
			}
			continue
		}

		// Parse column definition
		column := p.parseColumn(columnDef)
		if column != nil {
			table.Columns = append(table.Columns, column)
		}
	}

	return table, nil
}

// ParseAlterTable parses ALTER TABLE statement
func (p *SQLiteParser) ParseAlterTable(sql string) (*sdc.AlterTable, error) {
	// Remove comments and extra whitespace
	sql = removeComments(sql)
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "ALTER TABLE") {
		return nil, fmt.Errorf("invalid ALTER TABLE statement")
	}

	alterTable := &sdc.AlterTable{}

	// Extract table name and action
	alterRegex := regexp.MustCompile(`ALTER\s+TABLE\s+([^\s]+)\s+(.*)`)
	matches := alterRegex.FindStringSubmatch(sql)
	if len(matches) < 3 {
		return nil, fmt.Errorf("could not parse ALTER TABLE statement")
	}

	alterTable.Table = matches[1]
	action := strings.TrimSpace(matches[2])

	// Parse action
	if strings.HasPrefix(strings.ToUpper(action), "ADD COLUMN") {
		alterTable.Action = "ADD COLUMN"
		alterTable.Column = p.parseColumn(action[10:])
	} else if strings.HasPrefix(strings.ToUpper(action), "DROP COLUMN") {
		alterTable.Action = "DROP COLUMN"
		alterTable.Column = &sdc.Column{Name: strings.TrimSpace(action[11:])}
	} else if strings.HasPrefix(strings.ToUpper(action), "RENAME TO") {
		alterTable.Action = "RENAME TO"
		alterTable.NewName = strings.TrimSpace(action[9:])
	}

	return alterTable, nil
}

// ParseDropTable parses DROP TABLE statement
func (p *SQLiteParser) ParseDropTable(sql string) (*sdc.DropTable, error) {
	// Remove comments and extra whitespace
	sql = removeComments(sql)
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "DROP TABLE") {
		return nil, fmt.Errorf("invalid DROP TABLE statement")
	}

	dropTable := &sdc.DropTable{}

	// Extract table name
	dropRegex := regexp.MustCompile(`DROP\s+TABLE\s+(?:IF\s+EXISTS\s+)?([^\s;]+)`)
	matches := dropRegex.FindStringSubmatch(sql)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not parse DROP TABLE statement")
	}

	dropTable.Table = matches[1]
	dropTable.IfExists = strings.Contains(strings.ToUpper(sql), "IF EXISTS")

	return dropTable, nil
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

	// Extract constraint name if exists
	if strings.HasPrefix(strings.ToUpper(constraintDef), "CONSTRAINT") {
		nameRegex := regexp.MustCompile(`CONSTRAINT\s+(\w+)\s+(.*)`)
		matches := nameRegex.FindStringSubmatch(constraintDef)
		if len(matches) < 3 {
			return nil
		}
		constraint.Name = matches[1]
		constraintDef = matches[2]
	}

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
