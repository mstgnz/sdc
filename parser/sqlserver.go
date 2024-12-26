package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mstgnz/sdc"
)

// Precompiled regex patterns for better performance
var (
	sqlServerCreateTableRegex = regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:\[?([^\]\.]+)\]?\.)?(?:\[?([^\]\s(]+)\]?)\s*\((.*)\)(?:\s+ON\s+\[?([^\]]+)\]?)?`)
	sqlServerColumnRegex      = regexp.MustCompile(`(?i)\[?([^\]]+)\]?\s+([^\s]+)(?:\s*\(([^)]+)\))?(.*)`)
	sqlServerIdentityRegex    = regexp.MustCompile(`(?i)IDENTITY(?:\s*\((\d+)\s*,\s*(\d+)\s*\))?`)
	sqlServerDefaultRegex     = regexp.MustCompile(`(?i)DEFAULT\s+([^,\s]+|'[^']*'|\([^)]+\))`)
	sqlserverColumnRegex      = regexp.MustCompile(`(?i)^\s*\[?([^\[\]\s(]+)\]?\s+([^(,\s]+)(?:\s*\(([^)]+)\))?\s*(.*)$`)
	sqlserverDefaultRegex     = regexp.MustCompile(`(?i)DEFAULT\s+([^,\s]+|'[^']*'|"[^"]*")`)
	sqlserverCollateRegex     = regexp.MustCompile(`(?i)COLLATE\s+([^\s,]+)`)
)

// SQLServerParser implements the parser for SQL Server database
type SQLServerParser struct {
	dbInfo DatabaseInfo
	// Add buffer pool for string builders to reduce allocations
	builderPool strings.Builder
}

// Parse converts SQL Server dump to Entity structure
func (p *SQLServerParser) Parse(sql string) (*sdc.Entity, error) {
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

// Convert transforms Entity structure to SQL Server format with optimized string handling
func (p *SQLServerParser) Convert(entity *sdc.Entity) (string, error) {
	// Reset and reuse string builder
	p.builderPool.Reset()
	result := &p.builderPool

	for _, table := range entity.Tables {
		// CREATE TABLE statement
		result.WriteString("CREATE TABLE [")
		result.WriteString(table.Name)
		result.WriteString("] (\n")

		// Columns
		for i, column := range table.Columns {
			result.WriteString("    [")
			result.WriteString(column.Name)
			result.WriteString("] ")
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

			if column.Identity {
				if column.IdentitySeed != 0 || column.IdentityIncr != 0 {
					result.WriteString(" IDENTITY(")
					result.WriteString(fmt.Sprintf("%d,%d", column.IdentitySeed, column.IdentityIncr))
					result.WriteString(")")
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
			result.WriteString(") ON [")
			result.WriteString(table.FileGroup)
			result.WriteString("]")
		} else {
			result.WriteString(")")
		}

		result.WriteString("\n\n")
	}

	return result.String(), nil
}

// parseCreateTable parses CREATE TABLE statement with optimized string handling
func (p *SQLServerParser) parseCreateTable(sql string) (*sdc.Table, error) {
	matches := sqlServerCreateTableRegex.FindStringSubmatch(sql)
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid CREATE TABLE statement")
	}

	table := &sdc.Table{
		Schema:      strings.Trim(matches[1], "[]\""),
		Name:        strings.Trim(matches[2], "[]\""),
		FileGroup:   strings.Trim(matches[4], "[]\""),
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
func (p *SQLServerParser) parseColumn(columnDef string) *sdc.Column {
	matches := sqlserverColumnRegex.FindStringSubmatch(columnDef)
	if len(matches) < 3 {
		return nil
	}

	column := &sdc.Column{
		Name:       strings.Trim(matches[1], "[]\""),
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
		// Check for IDENTITY
		if strings.Contains(strings.ToUpper(rest), "IDENTITY") {
			column.Identity = true
			column.AutoIncrement = true
		}

		// Check for DEFAULT
		if defaultMatches := sqlserverDefaultRegex.FindStringSubmatch(rest); defaultMatches != nil {
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

		// Check for COLLATE
		if collateMatches := sqlserverCollateRegex.FindStringSubmatch(rest); collateMatches != nil {
			column.Collation = collateMatches[1]
		}
	}

	return column
}

// parseConstraint parses constraint definition
func (p *SQLServerParser) parseConstraint(constraintDef string) *sdc.Constraint {
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
			// Check for CLUSTERED/NONCLUSTERED
			nextIdx := startIdx + 2
			if nextIdx < len(parts) {
				switch strings.ToUpper(parts[nextIdx]) {
				case "CLUSTERED":
					constraint.Clustered = true
					nextIdx++
				case "NONCLUSTERED":
					constraint.NonClustered = true
					nextIdx++
				}
			}
			// Parse columns
			if nextIdx < len(parts) && strings.HasPrefix(parts[nextIdx], "(") {
				colStr := strings.Trim(parts[nextIdx], "()")
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
		// Check for CLUSTERED/NONCLUSTERED
		nextIdx := startIdx + 1
		if nextIdx < len(parts) {
			switch strings.ToUpper(parts[nextIdx]) {
			case "CLUSTERED":
				constraint.Clustered = true
				nextIdx++
			case "NONCLUSTERED":
				constraint.NonClustered = true
				nextIdx++
			}
		}
		// Parse columns
		if nextIdx < len(parts) && strings.HasPrefix(parts[nextIdx], "(") {
			colStr := strings.Trim(parts[nextIdx], "()")
			cols := strings.Split(colStr, ",")
			for _, col := range cols {
				constraint.Columns = append(constraint.Columns, strings.Trim(col, "[]\""))
			}
		}
	}

	return constraint
}

// convertDataType converts data type to SQL Server format
func (p *SQLServerParser) convertDataType(dataType *sdc.DataType) string {
	if dataType == nil {
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

// parseAlterTable parses ALTER TABLE statement
func (p *SQLServerParser) parseAlterTable(sql string) (*sdc.AlterTable, error) {
	// Remove comments and normalize whitespace
	sql = removeComments(sql)
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "ALTER TABLE") {
		return nil, fmt.Errorf("invalid ALTER TABLE statement")
	}

	// Extract schema, table name and action
	alterTableRegex := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+(?:\[?([^\]\.]+)\]?\.)?(?:\[?([^\]\s]+)\]?)\s+(?:ADD|DROP|ALTER\s+COLUMN)\s+(.+)`)
	matches := alterTableRegex.FindStringSubmatch(sql)
	if len(matches) < 4 {
		return nil, fmt.Errorf("could not parse ALTER TABLE statement")
	}

	schema := matches[1]
	if schema == "" {
		schema = "dbo" // Default schema for SQL Server
	}

	alterTable := &sdc.AlterTable{
		Schema: strings.Trim(schema, "[]\""),
		Table:  strings.Trim(matches[2], "[]\""),
	}

	// Parse the action and its details
	action := strings.ToUpper(matches[3])
	if strings.HasPrefix(action, "ADD CONSTRAINT") {
		alterTable.Action = "ADD CONSTRAINT"
		constraint := p.parseConstraint(matches[3][len("ADD CONSTRAINT"):])
		if constraint != nil {
			alterTable.Constraint = constraint
		}
	} else if strings.HasPrefix(action, "ADD") {
		alterTable.Action = "ADD COLUMN"
		column := p.parseColumn(matches[3][len("ADD"):])
		if column != nil {
			alterTable.Column = column
		}
	} else if strings.HasPrefix(action, "DROP CONSTRAINT") {
		alterTable.Action = "DROP CONSTRAINT"
		constraintName := strings.Trim(matches[3][len("DROP CONSTRAINT"):], " []\"")
		alterTable.Constraint = &sdc.Constraint{Name: constraintName}
	} else if strings.HasPrefix(action, "DROP COLUMN") || strings.HasPrefix(action, "DROP") {
		alterTable.Action = "DROP COLUMN"
		columnName := strings.Trim(matches[3][len("DROP COLUMN"):], " []\"")
		if strings.HasPrefix(columnName, "COLUMN ") {
			columnName = strings.Trim(columnName[len("COLUMN"):], " []\"")
		}
		alterTable.Column = &sdc.Column{Name: columnName}
	} else if strings.HasPrefix(action, "ALTER COLUMN") {
		alterTable.Action = "ALTER COLUMN"
		column := p.parseColumn(matches[3][len("ALTER COLUMN"):])
		if column != nil {
			alterTable.Column = column
		}
	}

	return alterTable, nil
}

// parseDropTable parses DROP TABLE statement
func (p *SQLServerParser) parseDropTable(sql string) (*sdc.DropTable, error) {
	// Remove comments and normalize whitespace
	sql = removeComments(sql)
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "DROP TABLE") {
		return nil, fmt.Errorf("invalid DROP TABLE statement")
	}

	// Check for IF EXISTS
	ifExists := strings.Contains(strings.ToUpper(sql), "IF EXISTS")

	// Extract schema and table name
	dropTableRegex := regexp.MustCompile(`(?i)DROP\s+TABLE\s+(?:IF\s+EXISTS\s+)?(?:\[?([^\]\.]+)\]?\.)?(?:\[?([^\]\s;]+)\]?)`)
	matches := dropTableRegex.FindStringSubmatch(sql)
	if len(matches) < 3 {
		return nil, fmt.Errorf("could not parse DROP TABLE statement")
	}

	schema := matches[1]
	if schema == "" {
		schema = "dbo" // Default schema for SQL Server
	}

	return &sdc.DropTable{
		Schema:   strings.Trim(schema, "[]\""),
		Table:    strings.Trim(matches[2], "[]\""),
		IfExists: ifExists,
		Cascade:  strings.Contains(strings.ToUpper(sql), "CASCADE"),
	}, nil
}

// parseCreateIndex parses CREATE INDEX statement
func (p *SQLServerParser) parseCreateIndex(sql string) (*sdc.Index, error) {
	// Remove comments and normalize whitespace
	sql = removeComments(sql)
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "CREATE") {
		return nil, fmt.Errorf("invalid CREATE INDEX statement")
	}

	// Check if it's a unique index
	isUnique := strings.HasPrefix(strings.ToUpper(sql), "CREATE UNIQUE")

	// Extract index details
	indexRegex := regexp.MustCompile(`(?i)CREATE\s+(?:UNIQUE\s+)?(?:CLUSTERED\s+)?(?:NONCLUSTERED\s+)?INDEX\s+(?:\[?([^\]\.]+)\]?)\s+ON\s+(?:\[?([^\]\.]+)\]?\.)?(?:\[?([^\]\s(]+)\]?)\s*\(([^)]+)\)(?:\s+INCLUDE\s*\(([^)]+)\))?(?:\s+WHERE\s+(.+))?(?:\s+WITH\s+\(([^)]+)\))?(?:\s+ON\s+(?:\[?([^\]\s]+)\]?))?`)
	matches := indexRegex.FindStringSubmatch(sql)
	if len(matches) < 5 {
		return nil, fmt.Errorf("could not parse CREATE INDEX statement")
	}

	schema := matches[2]
	if schema == "" {
		schema = "dbo" // Default schema for SQL Server
	}

	// Parse column names
	var columns []string
	columnList := strings.Split(matches[4], ",")
	for _, col := range columnList {
		col = strings.TrimSpace(col)
		// Remove ASC/DESC and other options, keep only the column name
		colParts := strings.Fields(col)
		if len(colParts) > 0 {
			columns = append(columns, strings.Trim(colParts[0], "[]\""))
		}
	}

	// Parse included columns if any
	var includeColumns []string
	if matches[5] != "" {
		includeList := strings.Split(matches[5], ",")
		for _, col := range includeList {
			col = strings.TrimSpace(col)
			includeColumns = append(includeColumns, strings.Trim(col, "[]\""))
		}
	}

	// Parse options if any
	var options map[string]string
	if matches[7] != "" {
		options = make(map[string]string)
		optionsList := strings.Split(matches[7], ",")
		for _, opt := range optionsList {
			opt = strings.TrimSpace(opt)
			parts := strings.SplitN(opt, "=", 2)
			if len(parts) == 2 {
				options[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	return &sdc.Index{
		Name:           strings.Trim(matches[1], "[]\""),
		Schema:         strings.Trim(schema, "[]\""),
		Table:          strings.Trim(matches[3], "[]\""),
		Columns:        columns,
		Unique:         isUnique,
		Clustered:      strings.Contains(strings.ToUpper(sql), "CLUSTERED") && !strings.Contains(strings.ToUpper(sql), "NONCLUSTERED"),
		NonClustered:   strings.Contains(strings.ToUpper(sql), "NONCLUSTERED"),
		FileGroup:      strings.Trim(matches[8], "[]\""),
		Filter:         strings.TrimSpace(matches[6]),
		IncludeColumns: includeColumns,
		Options:        options,
	}, nil
}

// parseDropIndex parses DROP INDEX statement
func (p *SQLServerParser) parseDropIndex(sql string) (*sdc.DropIndex, error) {
	// Remove comments and normalize whitespace
	sql = removeComments(sql)
	sql = strings.TrimSpace(sql)

	// Basic validation
	if !strings.HasPrefix(strings.ToUpper(sql), "DROP INDEX") {
		return nil, fmt.Errorf("invalid DROP INDEX statement")
	}

	// Check for IF EXISTS
	ifExists := strings.Contains(strings.ToUpper(sql), "IF EXISTS")

	// Extract index and table names
	dropIndexRegex := regexp.MustCompile(`(?i)DROP\s+INDEX\s+(?:IF\s+EXISTS\s+)?(?:\[?([^\]\.]+)\]?)\s+ON\s+(?:\[?([^\]\.]+)\]?\.)?(?:\[?([^\]\s;]+)\]?)`)
	matches := dropIndexRegex.FindStringSubmatch(sql)
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid DROP INDEX statement")
	}

	schema := matches[2]
	if schema == "" {
		schema = "dbo" // Default schema for SQL Server
	}

	return &sdc.DropIndex{
		Schema:   strings.Trim(schema, "[]\""),
		Table:    strings.Trim(matches[3], "[]\""),
		Index:    strings.Trim(matches[1], "[]\""),
		IfExists: ifExists,
		Cascade:  strings.Contains(strings.ToUpper(sql), "CASCADE"),
	}, nil
}
