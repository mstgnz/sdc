package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mstgnz/sdc"
)

// MySQLParser implements the parser for MySQL database
type MySQLParser struct{}

// Parse converts MySQL dump to Entity structure
func (p *MySQLParser) Parse(sql string) (*sdc.Entity, error) {
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

// parseCreateTable parses CREATE TABLE statement
func (p *MySQLParser) parseCreateTable(sql string) (*sdc.Table, error) {
	table := &sdc.Table{}

	// Extract basic table information
	tableRegex := regexp.MustCompile(`(?i)CREATE\s+(?:TEMPORARY\s+)?TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([^\s(]+)\s*\((.*?)\)(?:\s+ENGINE\s*=\s*(\w+))?(?:\s+DEFAULT\s+CHARSET\s*=\s*(\w+))?(?:\s+COLLATE\s*=\s*(\w+))?`)
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

	// Set table options
	if len(matches) > 3 && matches[3] != "" {
		if table.Options == nil {
			table.Options = make(map[string]string)
		}
		table.Options["ENGINE"] = matches[3]
	}

	if len(matches) > 4 && matches[4] != "" {
		if table.Options == nil {
			table.Options = make(map[string]string)
		}
		table.Options["DEFAULT CHARSET"] = matches[4]
	}

	if len(matches) > 5 && matches[5] != "" {
		table.Collation = matches[5]
	}

	if len(comments) > 0 {
		table.Comment = strings.Join(comments, ", ")
	}

	return table, nil
}

// parseColumn parses column definition
func (p *MySQLParser) parseColumn(columnDef string) *sdc.Column {
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
					precision, _ := strconv.Atoi(strings.TrimSpace(precisionScale[0]))
					dataType.Precision = precision
					if len(precisionScale) > 1 {
						scale, _ := strconv.Atoi(strings.TrimSpace(precisionScale[1]))
						dataType.Scale = scale
					}
				} else {
					length, _ := strconv.Atoi(strings.TrimSpace(precisionScale[0]))
					dataType.Length = length
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

	// Check for AUTO_INCREMENT
	if strings.Contains(strings.ToUpper(columnDef), "AUTO_INCREMENT") {
		column.AutoIncrement = true
	}

	// Check for COLLATE definition
	collateRegex := regexp.MustCompile(`(?i)COLLATE\s+([^,\s]+)`)
	collateMatches := collateRegex.FindStringSubmatch(columnDef)
	if len(collateMatches) > 1 {
		column.Collation = collateMatches[1]
	}

	// Check for COMMENT
	commentRegex := regexp.MustCompile(`(?i)COMMENT\s+'([^']+)'`)
	commentMatches := commentRegex.FindStringSubmatch(columnDef)
	if len(commentMatches) > 1 {
		column.Comment = commentMatches[1]
	}

	return column
}

// Convert transforms Entity structure to MySQL format
func (p *MySQLParser) Convert(entity *sdc.Entity) (string, error) {
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

			if column.AutoIncrement {
				result.WriteString(" AUTO_INCREMENT")
			}

			if column.Comment != "" {
				result.WriteString(fmt.Sprintf(" COMMENT '%s'", p.EscapeString(column.Comment)))
			}

			if i < len(table.Columns)-1 {
				result.WriteString(",\n")
			}
		}

		// Table options
		if table.Options != nil {
			if engine, ok := table.Options["ENGINE"]; ok {
				result.WriteString(fmt.Sprintf(" ENGINE=%s", engine))
			}

			if charset, ok := table.Options["DEFAULT CHARSET"]; ok {
				result.WriteString(fmt.Sprintf(" DEFAULT CHARSET=%s", charset))
			}
		}

		if table.Collation != "" {
			result.WriteString(fmt.Sprintf(" COLLATE=%s", table.Collation))
		}

		result.WriteString("\n);\n\n")
	}

	return result.String(), nil
}

// convertDataType converts data type to MySQL format
func (p *MySQLParser) convertDataType(dataType *sdc.DataType) string {
	if dataType == nil {
		return "VARCHAR(255)"
	}

	// Tip dönüşümü için map kullan
	if mappedType, ok := mysqlDataTypeMap[strings.ToLower(dataType.Name)]; ok {
		// Varsayılan uzunluk kontrolü
		if defaultLength, hasDefault := mysqlDefaultLengthMap[mappedType]; hasDefault {
			if dataType.Length == 0 {
				dataType.Length = defaultLength
			}
		}

		// Varsayılan precision kontrolü
		if defaultPrecision, hasDefault := mysqlDefaultPrecisionMap[mappedType]; hasDefault {
			if dataType.Precision == 0 {
				dataType.Precision = defaultPrecision
			}
		}

		// Varsayılan scale kontrolü
		if defaultScale, hasDefault := mysqlDefaultScaleMap[mappedType]; hasDefault {
			if dataType.Scale == 0 {
				dataType.Scale = defaultScale
			}
		}

		// Uzunluk, precision ve scale formatlaması
		switch mappedType {
		case "VARCHAR", "CHAR", "BINARY", "VARBINARY":
			return fmt.Sprintf("%s(%d)", mappedType, dataType.Length)
		case "DECIMAL", "NUMERIC":
			if dataType.Scale > 0 {
				return fmt.Sprintf("%s(%d,%d)", mappedType, dataType.Precision, dataType.Scale)
			}
			return fmt.Sprintf("%s(%d)", mappedType, dataType.Precision)
		default:
			return mappedType
		}
	}

	return "VARCHAR(255)"
}

// GetDefaultSchema returns the default schema name
func (p *MySQLParser) GetDefaultSchema() string {
	return ""
}

// GetSchemaPrefix returns the schema prefix
func (p *MySQLParser) GetSchemaPrefix(schema string) string {
	if schema == "" {
		return ""
	}
	return schema + "."
}

// GetIdentifierQuote returns the identifier quote character
func (p *MySQLParser) GetIdentifierQuote() string {
	return "`"
}

// GetStringQuote returns the string quote character
func (p *MySQLParser) GetStringQuote() string {
	return "'"
}

// GetMaxIdentifierLength returns the maximum identifier length
func (p *MySQLParser) GetMaxIdentifierLength() int {
	return 64
}

// GetReservedWords returns the reserved words
func (p *MySQLParser) GetReservedWords() []string {
	return []string{
		"accessible", "add", "all", "alter", "analyze", "and", "as", "asc",
		"asensitive", "before", "between", "bigint", "binary", "blob",
		"both", "by", "call", "cascade", "case", "change", "char",
		"character", "check", "collate", "column", "condition",
		"constraint", "continue", "convert", "create", "cross",
		"current_date", "current_time", "current_timestamp", "current_user",
		"cursor", "database", "databases", "day_hour", "day_microsecond",
		"day_minute", "day_second", "dec", "decimal", "declare", "default",
		"delayed", "delete", "desc", "describe", "deterministic",
		"distinct", "distinctrow", "div", "double", "drop", "dual", "each",
		"else", "elseif", "enclosed", "escaped", "exists", "exit",
		"explain", "false", "fetch", "float", "float4", "float8", "for",
		"force", "foreign", "from", "fulltext", "grant", "group",
		"having", "high_priority", "hour_microsecond", "hour_minute",
		"hour_second", "if", "ignore", "in", "index", "infile", "inner",
		"inout", "insensitive", "insert", "int", "int1", "int2", "int3",
		"int4", "int8", "integer", "interval", "into", "is", "iterate",
		"join", "key", "keys", "kill", "leading", "leave", "left", "like",
		"limit", "linear", "lines", "load", "localtime", "localtimestamp",
		"lock", "long", "longblob", "longtext", "loop", "low_priority",
		"match", "mediumblob", "mediumint", "mediumtext", "middleint",
		"minute_microsecond", "minute_second", "mod", "modifies",
		"natural", "not", "no_write_to_binlog", "null", "numeric", "on",
		"optimize", "option", "optionally", "or", "order", "out", "outer",
		"outfile", "precision", "primary", "procedure", "purge", "range",
		"read", "reads", "read_write", "real", "references", "regexp",
		"release", "rename", "repeat", "replace", "require", "restrict",
		"return", "revoke", "right", "rlike", "schema", "schemas",
		"second_microsecond", "select", "sensitive", "separator", "set",
		"show", "smallint", "spatial", "specific", "sql", "sqlexception",
		"sqlstate", "sqlwarning", "sql_big_result", "sql_calc_found_rows",
		"sql_small_result", "ssl", "starting", "straight_join", "table",
		"terminated", "then", "tinyblob", "tinyint", "tinytext", "to",
		"trailing", "trigger", "true", "undo", "union", "unique", "unlock",
		"unsigned", "update", "usage", "use", "using", "utc_date",
		"utc_time", "utc_timestamp", "values", "varbinary", "varchar",
		"varcharacter", "varying", "when", "where", "while", "with",
		"write", "xor", "year_month", "zerofill",
	}
}

// ConvertDataTypeFrom converts source database data type to MySQL data type
func (p *MySQLParser) ConvertDataTypeFrom(sourceType string, length int, precision int, scale int) *sdc.DataType {
	return &sdc.DataType{
		Name:      sourceType,
		Length:    length,
		Precision: precision,
		Scale:     scale,
	}
}

// ParseCreateTable parses CREATE TABLE statement
func (p *MySQLParser) ParseCreateTable(sql string) (*sdc.Table, error) {
	return p.parseCreateTable(sql)
}

// ParseAlterTable parses ALTER TABLE statement
func (p *MySQLParser) ParseAlterTable(sql string) (*sdc.AlterTable, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseDropTable parses DROP TABLE statement
func (p *MySQLParser) ParseDropTable(sql string) (*sdc.DropTable, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseCreateIndex parses CREATE INDEX statement
func (p *MySQLParser) ParseCreateIndex(sql string) (*sdc.Index, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ParseDropIndex parses DROP INDEX statement
func (p *MySQLParser) ParseDropIndex(sql string) (*sdc.DropIndex, error) {
	// Parse logic to be implemented
	return nil, nil
}

// ValidateIdentifier validates the identifier name
func (p *MySQLParser) ValidateIdentifier(name string) error {
	// Validation logic to be implemented
	return nil
}

// EscapeIdentifier escapes the identifier name
func (p *MySQLParser) EscapeIdentifier(name string) string {
	// Escape logic to be implemented
	return name
}

// EscapeString escapes the string value
func (p *MySQLParser) EscapeString(value string) string {
	// Escape logic to be implemented
	return value
}
