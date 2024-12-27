// Package mysql provides functionality for parsing and generating MySQL database schemas.
// It implements the Parser interface for handling MySQL specific SQL syntax and schema structures.
package mysql

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/mstgnz/sqlmapper"
)

// MySQL represents a MySQL parser implementation that handles parsing and generating
// MySQL database schemas. It maintains an internal schema representation and provides
// methods for converting between MySQL SQL and the common schema format.
type MySQL struct {
	schema *sqlmapper.Schema
}

// NewMySQL creates and initializes a new MySQL parser instance.
// It returns a parser that can handle MySQL specific SQL syntax and schema structures.
func NewMySQL() *MySQL {
	return &MySQL{
		schema: &sqlmapper.Schema{},
	}
}

// Parse takes a MySQL SQL dump content and parses it into a common schema structure.
// It processes various MySQL objects including:
// - Databases and schemas
// - Tables with columns and constraints
// - Indexes (including PRIMARY, UNIQUE, and FULLTEXT)
// - Views
// - Stored procedures and functions
// - Triggers
// - User privileges
//
// Parameters:
//   - content: The MySQL SQL dump content to parse
//
// Returns:
//   - *sqlmapper.Schema: The parsed schema structure
//   - error: An error if parsing fails
func (m *MySQL) Parse(content string) (*sqlmapper.Schema, error) {
	if content == "" {
		return nil, errors.New("empty content")
	}

	// Normalize content
	content = m.normalizeContent(content)

	// Parse schema objects
	if err := m.parseSchemas(content); err != nil {
		return nil, fmt.Errorf("error parsing schemas: %v", err)
	}

	if err := m.parseTables(content); err != nil {
		return nil, fmt.Errorf("error parsing tables: %v", err)
	}

	if err := m.parseIndexes(content); err != nil {
		return nil, fmt.Errorf("error parsing indexes: %v", err)
	}

	if err := m.parseViews(content); err != nil {
		return nil, fmt.Errorf("error parsing views: %v", err)
	}

	if err := m.parseFunctions(content); err != nil {
		return nil, fmt.Errorf("error parsing functions: %v", err)
	}

	if err := m.parseTriggers(content); err != nil {
		return nil, fmt.Errorf("error parsing triggers: %v", err)
	}

	if err := m.parsePermissions(content); err != nil {
		return nil, fmt.Errorf("error parsing permissions: %v", err)
	}

	return m.schema, nil
}

// Generate creates a MySQL SQL dump from a schema structure.
// It generates SQL statements for all database objects in the schema, including:
// - Tables with columns, indexes, and constraints
// - Views
// - Stored procedures and functions
// - Triggers
// - User privileges
//
// Parameters:
//   - schema: The schema structure to convert to MySQL SQL
//
// Returns:
//   - string: The generated MySQL SQL statements
//   - error: An error if generation fails
func (m *MySQL) Generate(schema *sqlmapper.Schema) (string, error) {
	if schema == nil {
		return "", errors.New("empty schema")
	}

	var result strings.Builder

	// Generate table creation
	for i, table := range schema.Tables {
		result.WriteString(m.generateTableSQL(table))
		if i < len(schema.Tables)-1 {
			result.WriteString("\n\n")
		}

		// Generate indexes for this table
		if len(table.Indexes) > 0 {
			result.WriteString("\n")
			for j, index := range table.Indexes {
				result.WriteString(m.generateIndexSQL(table.Name, index))
				if j < len(table.Indexes)-1 {
					result.WriteString("\n")
				}
			}
		}
	}

	return result.String(), nil
}

// normalizeContent preprocesses the SQL content by removing comments and normalizing whitespace.
// It handles MySQL specific comment styles (-- and #) and DELIMITER statements.
//
// Parameters:
//   - content: The SQL content to normalize
//
// Returns:
//   - string: The normalized SQL content
func (m *MySQL) normalizeContent(content string) string {
	// Remove comments
	re := regexp.MustCompile(`--.*$|#.*$`)
	content = re.ReplaceAllString(content, "")

	// Remove DELIMITER statements
	content = regexp.MustCompile(`DELIMITER\s+[^\s]+`).ReplaceAllString(content, "")

	// Normalize whitespace
	content = strings.TrimSpace(content)
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")

	return content
}

// parseSchemas extracts database definitions from the SQL content.
// It handles CREATE DATABASE and USE statements.
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (m *MySQL) parseSchemas(content string) error {
	// Parse CREATE DATABASE
	dbRe := regexp.MustCompile(`CREATE\s+DATABASE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`)
	if matches := dbRe.FindStringSubmatch(content); len(matches) > 1 {
		m.schema.Name = matches[1]
	}

	return nil
}

// parseTables extracts table definitions from the SQL content.
// It processes table structure including columns, indexes, constraints,
// and table options like ENGINE, CHARSET, and COLLATE.
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (m *MySQL) parseTables(content string) error {
	re := regexp.MustCompile(`CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([.\w]+)\s*\((.*?)\)(?:\s+ENGINE\s*=\s*\w+)?(?:\s+DEFAULT\s+CHARSET\s*=\s*\w+)?(?:\s+COLLATE\s*=\s*\w+)?;`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 2 {
			tableName := match[1]
			columnDefs := match[2]

			table := sqlmapper.Table{}

			// Parse schema if exists
			parts := strings.Split(tableName, ".")
			if len(parts) > 1 {
				table.Schema = parts[0]
				table.Name = parts[1]
			} else {
				table.Name = tableName
			}

			// Parse columns and constraints
			if err := m.parseColumnsAndConstraints(columnDefs, &table); err != nil {
				return err
			}

			// Parse table comment
			tableCommentRe := regexp.MustCompile(`ALTER\s+TABLE\s+` + regexp.QuoteMeta(tableName) + `\s+COMMENT\s*=\s*'([^']+)';`)
			if tableCommentMatch := tableCommentRe.FindStringSubmatch(content); len(tableCommentMatch) > 1 {
				table.Comment = tableCommentMatch[1]
			}

			// Parse column comments
			commentRe := regexp.MustCompile(`ALTER\s+TABLE\s+` + regexp.QuoteMeta(tableName) + `\s+MODIFY\s+COLUMN\s+(\w+)[^']+COMMENT\s*'([^']+)';`)
			commentMatches := commentRe.FindAllStringSubmatch(content, -1)
			for _, commentMatch := range commentMatches {
				if len(commentMatch) > 2 {
					columnName := commentMatch[1]
					comment := commentMatch[2]
					for i := range table.Columns {
						if table.Columns[i].Name == columnName {
							table.Columns[i].Comment = comment
							break
						}
					}
				}
			}

			// Set column order
			for i := range table.Columns {
				table.Columns[i].Order = i + 1
			}

			m.schema.Tables = append(m.schema.Tables, table)
		}
	}

	return nil
}

// parseColumnsAndConstraints processes column and constraint definitions within a table.
// It handles various column attributes and both inline and table-level constraints.
//
// Parameters:
//   - columnDefs: The column definitions string to parse
//   - table: The table structure to populate
//
// Returns:
//   - error: An error if parsing fails
func (m *MySQL) parseColumnsAndConstraints(columnDefs string, table *sqlmapper.Table) error {
	// Split column definitions
	defs := strings.Split(columnDefs, ",")
	var currentDef strings.Builder
	var finalDefs []string

	// Handle nested parentheses in CHECK constraints
	parenCount := 0
	for _, def := range defs {
		parenCount += strings.Count(def, "(") - strings.Count(def, ")")
		if parenCount > 0 {
			currentDef.WriteString(def + ",")
		} else {
			if currentDef.Len() > 0 {
				currentDef.WriteString(def)
				finalDefs = append(finalDefs, currentDef.String())
				currentDef.Reset()
			} else {
				finalDefs = append(finalDefs, def)
			}
		}
	}

	for _, def := range finalDefs {
		def = strings.TrimSpace(def)

		// Skip empty definitions
		if def == "" {
			continue
		}

		// Parse constraints
		if strings.HasPrefix(strings.ToUpper(def), "CONSTRAINT") ||
			(strings.Contains(strings.ToUpper(def), "PRIMARY KEY") && !strings.Contains(strings.ToUpper(def), "AUTO_INCREMENT")) ||
			strings.Contains(strings.ToUpper(def), "FOREIGN KEY") ||
			(strings.Contains(strings.ToUpper(def), "UNIQUE") && !strings.Contains(strings.ToUpper(def), " ")) ||
			(strings.Contains(strings.ToUpper(def), "CHECK") && !strings.Contains(strings.ToUpper(def), " ")) {
			constraint, err := m.parseConstraint(def)
			if err != nil {
				return err
			}
			table.Constraints = append(table.Constraints, constraint)
			continue
		}

		// Parse column
		if strings.Contains(def, " ") {
			column, err := m.parseColumn(def)
			if err != nil {
				return err
			}
			table.Columns = append(table.Columns, column)

			// Check for inline constraints
			if strings.Contains(strings.ToUpper(def), "PRIMARY KEY") {
				table.Constraints = append(table.Constraints, sqlmapper.Constraint{
					Type:    "PRIMARY KEY",
					Columns: []string{column.Name},
				})
				column.IsPrimaryKey = true
				column.IsNullable = false
			}
			if strings.Contains(strings.ToUpper(def), "UNIQUE") {
				table.Constraints = append(table.Constraints, sqlmapper.Constraint{
					Type:    "UNIQUE",
					Columns: []string{column.Name},
				})
				column.IsUnique = true
			}
			if strings.Contains(strings.ToUpper(def), "CHECK") {
				re := regexp.MustCompile(`CHECK\s*\((.*?)\)`)
				if matches := re.FindStringSubmatch(def); len(matches) > 1 {
					table.Constraints = append(table.Constraints, sqlmapper.Constraint{
						Type:            "CHECK",
						Columns:         []string{column.Name},
						CheckExpression: matches[1],
					})
					column.CheckExpression = matches[1]
				}
			}
		}
	}

	return nil
}

// parseColumn processes a single column definition.
// It handles various column attributes including data type, length/precision,
// nullability, defaults, auto increment, and inline constraints.
//
// Parameters:
//   - def: The column definition string to parse
//
// Returns:
//   - sqlmapper.Column: The parsed column structure
//   - error: An error if parsing fails
func (m *MySQL) parseColumn(def string) (sqlmapper.Column, error) {
	parts := strings.Fields(def)
	if len(parts) < 2 {
		return sqlmapper.Column{}, fmt.Errorf("invalid column definition: %s", def)
	}

	column := sqlmapper.Column{
		Name:       parts[0],
		DataType:   parts[1],
		IsNullable: true,
	}

	// Handle AUTO_INCREMENT
	if strings.Contains(strings.ToUpper(def), "AUTO_INCREMENT") {
		column.AutoIncrement = true
	}

	// Parse length/precision
	if strings.Contains(column.DataType, "(") {
		re := regexp.MustCompile(`(\w+)\((\d+)(?:,(\d+))?\)`)
		if matches := re.FindStringSubmatch(column.DataType); len(matches) > 2 {
			column.DataType = matches[1]
			if len(matches[2]) > 0 {
				fmt.Sscanf(matches[2], "%d", &column.Length)
			}
			if len(matches) > 3 && len(matches[3]) > 0 {
				fmt.Sscanf(matches[3], "%d", &column.Scale)
			}
		}
	}

	// Parse default value
	if idx := strings.Index(strings.ToUpper(def), "DEFAULT"); idx >= 0 {
		defaultPart := def[idx+7:]
		defaultPart = strings.TrimSpace(defaultPart)

		// Handle function calls and keywords
		if strings.Contains(strings.ToUpper(defaultPart), "CURRENT_TIMESTAMP") {
			column.DefaultValue = "CURRENT_TIMESTAMP"
		} else if strings.HasPrefix(defaultPart, "'") {
			// Handle quoted string values
			re := regexp.MustCompile(`'([^']*)'`)
			if matches := re.FindStringSubmatch(defaultPart); len(matches) > 1 {
				column.DefaultValue = matches[1]
			}
		} else {
			// Handle other values
			endIdx := strings.Index(defaultPart, " ")
			if endIdx == -1 {
				endIdx = len(defaultPart)
			}
			defaultValue := strings.TrimSpace(defaultPart[:endIdx])
			// Remove trailing comma
			defaultValue = strings.TrimSuffix(defaultValue, ",")
			column.DefaultValue = defaultValue
		}
	}

	// Parse column constraints
	if strings.Contains(strings.ToUpper(def), "PRIMARY KEY") {
		column.IsPrimaryKey = true
		column.IsNullable = false
	}
	if strings.Contains(strings.ToUpper(def), "UNIQUE") {
		column.IsUnique = true
	}
	if strings.Contains(strings.ToUpper(def), "CHECK") {
		re := regexp.MustCompile(`CHECK\s*\((.*?)\)`)
		if matches := re.FindStringSubmatch(def); len(matches) > 1 {
			column.CheckExpression = matches[1]
		}
	}

	// Handle NOT NULL after other constraints
	if strings.Contains(strings.ToUpper(def), "NOT NULL") {
		column.IsNullable = false
	} else if strings.Contains(strings.ToUpper(def), "NULL") && !strings.Contains(strings.ToUpper(def), "NOT NULL") {
		column.IsNullable = true
	} else {
		// Default to NULL
		column.IsNullable = true
	}

	return column, nil
}

// parseConstraint processes a table constraint definition.
// It handles various constraint types including PRIMARY KEY, FOREIGN KEY,
// UNIQUE, and CHECK constraints.
//
// Parameters:
//   - def: The constraint definition string to parse
//
// Returns:
//   - sqlmapper.Constraint: The parsed constraint structure
//   - error: An error if parsing fails
func (m *MySQL) parseConstraint(def string) (sqlmapper.Constraint, error) {
	constraint := sqlmapper.Constraint{}

	// Extract constraint name if exists
	if strings.HasPrefix(strings.ToUpper(def), "CONSTRAINT") {
		re := regexp.MustCompile(`CONSTRAINT\s+(\w+)\s+(.*)`)
		if matches := re.FindStringSubmatch(def); len(matches) > 2 {
			constraint.Name = matches[1]
			def = matches[2]
		}
	}

	if strings.Contains(strings.ToUpper(def), "PRIMARY KEY") {
		constraint.Type = "PRIMARY KEY"
		re := regexp.MustCompile(`PRIMARY\s+KEY\s*\((.*?)\)`)
		if matches := re.FindStringSubmatch(def); len(matches) > 1 {
			constraint.Columns = strings.Split(matches[1], ",")
			for i := range constraint.Columns {
				constraint.Columns[i] = strings.TrimSpace(constraint.Columns[i])
			}
		}
	} else if strings.Contains(strings.ToUpper(def), "FOREIGN KEY") {
		constraint.Type = "FOREIGN KEY"
		re := regexp.MustCompile(`FOREIGN\s+KEY\s*\((.*?)\)\s*REFERENCES\s+([.\w]+)\s*\((.*?)\)`)
		if matches := re.FindStringSubmatch(def); len(matches) > 3 {
			constraint.Columns = strings.Split(matches[1], ",")
			constraint.RefTable = matches[2]
			constraint.RefColumns = strings.Split(matches[3], ",")
			for i := range constraint.Columns {
				constraint.Columns[i] = strings.TrimSpace(constraint.Columns[i])
			}
			for i := range constraint.RefColumns {
				constraint.RefColumns[i] = strings.TrimSpace(constraint.RefColumns[i])
			}
		}
		if strings.Contains(strings.ToUpper(def), "ON DELETE") {
			if strings.Contains(strings.ToUpper(def), "CASCADE") {
				constraint.DeleteRule = "CASCADE"
			} else if strings.Contains(strings.ToUpper(def), "SET NULL") {
				constraint.DeleteRule = "SET NULL"
			}
		}
	} else if strings.Contains(strings.ToUpper(def), "UNIQUE") {
		constraint.Type = "UNIQUE"
		re := regexp.MustCompile(`UNIQUE\s*\((.*?)\)`)
		if matches := re.FindStringSubmatch(def); len(matches) > 1 {
			constraint.Columns = strings.Split(matches[1], ",")
			for i := range constraint.Columns {
				constraint.Columns[i] = strings.TrimSpace(constraint.Columns[i])
			}
		}
	} else if strings.Contains(strings.ToUpper(def), "CHECK") {
		constraint.Type = "CHECK"
		re := regexp.MustCompile(`CHECK\s*\((.*?)\)`)
		if matches := re.FindStringSubmatch(def); len(matches) > 1 {
			constraint.CheckExpression = matches[1]
		}
	}

	return constraint, nil
}

// parseIndexes extracts index definitions from the SQL content.
// It handles various index types including PRIMARY KEY, UNIQUE,
// FULLTEXT, and regular indexes.
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (m *MySQL) parseIndexes(content string) error {
	re := regexp.MustCompile(`CREATE\s+(?:UNIQUE\s+)?(?:FULLTEXT\s+)?INDEX\s+(\w+)\s+ON\s+([.\w]+)\s*\((.*?)\)`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 3 {
			indexName := match[1]
			tableName := match[2]
			columns := strings.Split(match[3], ",")

			// Find the table
			for i, table := range m.schema.Tables {
				if table.Name == tableName || fmt.Sprintf("%s.%s", table.Schema, table.Name) == tableName {
					index := sqlmapper.Index{
						Name:     indexName,
						Columns:  make([]string, len(columns)),
						IsUnique: strings.Contains(match[0], "UNIQUE"),
					}

					// Clean column names
					for j, col := range columns {
						index.Columns[j] = strings.TrimSpace(col)
					}

					m.schema.Tables[i].Indexes = append(m.schema.Tables[i].Indexes, index)
					break
				}
			}
		}
	}

	return nil
}

// parseViews processes view definitions from the SQL content.
// It handles both regular and updatable views with their definitions.
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (m *MySQL) parseViews(content string) error {
	viewRe := regexp.MustCompile(`CREATE(?:\s+OR\s+REPLACE)?\s+VIEW\s+([.\w]+)\s+AS\s+(.*?);`)
	viewMatches := viewRe.FindAllStringSubmatch(content, -1)

	for _, match := range viewMatches {
		if len(match) > 2 {
			viewName := match[1]
			view := sqlmapper.View{
				Definition: match[2],
			}

			// Parse schema if exists
			parts := strings.Split(viewName, ".")
			if len(parts) > 1 {
				view.Schema = parts[0]
				view.Name = parts[1]
			} else {
				view.Name = viewName
			}

			m.schema.Views = append(m.schema.Views, view)
		}
	}

	return nil
}

// parseFunctions extracts function and procedure definitions from the SQL content.
// It handles various routine attributes including parameters, return types,
// and procedure parameter directions (IN/OUT/INOUT).
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (m *MySQL) parseFunctions(content string) error {
	// Parse functions
	funcRe := regexp.MustCompile(`CREATE\s+FUNCTION\s+([.\w]+)\s*\((.*?)\)\s+RETURNS\s+(\w+)\s+BEGIN\s+(.*?)\s+END`)
	funcMatches := funcRe.FindAllStringSubmatch(content, -1)

	for _, match := range funcMatches {
		if len(match) > 4 {
			functionName := match[1]
			function := sqlmapper.Function{
				Returns: match[3],
				Body:    match[4],
			}

			// Parse schema if exists
			parts := strings.Split(functionName, ".")
			if len(parts) > 1 {
				function.Schema = parts[0]
				function.Name = parts[1]
			} else {
				function.Name = functionName
			}

			// Parse parameters
			if match[2] != "" {
				params := strings.Split(match[2], ",")
				for _, param := range params {
					parts := strings.Fields(strings.TrimSpace(param))
					if len(parts) >= 2 {
						parameter := sqlmapper.Parameter{
							Name:     parts[0],
							DataType: parts[1],
						}
						function.Parameters = append(function.Parameters, parameter)
					}
				}
			}

			m.schema.Functions = append(m.schema.Functions, function)
		}
	}

	// Parse procedures
	procRe := regexp.MustCompile(`CREATE\s+PROCEDURE\s+([.\w]+)\s*\((.*?)\)\s+BEGIN\s+(.*?)\s+END`)
	procMatches := procRe.FindAllStringSubmatch(content, -1)

	for _, match := range procMatches {
		if len(match) > 3 {
			procName := match[1]
			function := sqlmapper.Function{
				Name:   procName,
				Body:   match[3],
				IsProc: true,
			}

			// Parse schema if exists
			parts := strings.Split(procName, ".")
			if len(parts) > 1 {
				function.Schema = parts[0]
				function.Name = parts[1]
			}

			// Parse parameters
			if match[2] != "" {
				params := strings.Split(match[2], ",")
				for _, param := range params {
					parts := strings.Fields(strings.TrimSpace(param))
					if len(parts) >= 3 { // IN/OUT/INOUT parameter_name type
						parameter := sqlmapper.Parameter{
							Name:      parts[1],
							DataType:  parts[2],
							Direction: parts[0],
						}
						function.Parameters = append(function.Parameters, parameter)
					}
				}
			}

			m.schema.Functions = append(m.schema.Functions, function)
		}
	}

	return nil
}

// parseTriggers processes trigger definitions from the SQL content.
// It handles trigger timing (BEFORE/AFTER), events (INSERT/UPDATE/DELETE),
// and trigger bodies.
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (m *MySQL) parseTriggers(content string) error {
	triggerRe := regexp.MustCompile(`CREATE\s+TRIGGER\s+(\w+)\s+(BEFORE|AFTER)\s+(INSERT|UPDATE|DELETE)\s+ON\s+([.\w]+)\s+FOR\s+EACH\s+ROW\s+BEGIN\s+(.*?)\s+END`)
	triggerMatches := triggerRe.FindAllStringSubmatch(content, -1)

	for _, match := range triggerMatches {
		if len(match) > 5 {
			trigger := sqlmapper.Trigger{
				Name:       match[1],
				Timing:     match[2],
				Event:      match[3],
				Table:      match[4],
				Body:       match[5],
				ForEachRow: true,
			}

			// Parse schema if exists
			parts := strings.Split(trigger.Table, ".")
			if len(parts) > 1 {
				trigger.Schema = parts[0]
				trigger.Table = parts[1]
			}

			m.schema.Triggers = append(m.schema.Triggers, trigger)
		}
	}

	return nil
}

// parsePermissions extracts user privilege definitions from the SQL content.
// It handles GRANT and REVOKE statements for various privilege types,
// including table privileges and routine (PROCEDURE/FUNCTION) privileges.
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (m *MySQL) parsePermissions(content string) error {
	// Parse GRANT statements for tables
	grantRe := regexp.MustCompile(`GRANT\s+(.*?)\s+ON\s+([.\w*]+)\s+TO\s+'([^']+)'@'([^']+)'(?:\s+WITH\s+GRANT\s+OPTION)?;`)
	grantMatches := grantRe.FindAllStringSubmatch(content, -1)

	for _, match := range grantMatches {
		if len(match) > 4 {
			// Split privileges and handle multiple privileges in one statement
			privilegeStr := strings.TrimSpace(match[1])
			var privileges []string
			if strings.ToUpper(privilegeStr) == "ALL PRIVILEGES" {
				privileges = []string{"ALL PRIVILEGES"}
			} else {
				// Handle comma-separated privileges and potential whitespace
				for _, priv := range strings.Split(privilegeStr, ",") {
					privs := strings.Fields(strings.TrimSpace(priv))
					privileges = append(privileges, privs...)
				}
			}

			perm := sqlmapper.Permission{
				Type:       "GRANT",
				Privileges: privileges,
				Object:     match[2],
				Grantee:    fmt.Sprintf("%s@%s", match[3], match[4]),
				WithGrant:  strings.Contains(match[0], "WITH GRANT OPTION"),
			}
			m.schema.Permissions = append(m.schema.Permissions, perm)
		}
	}

	// Parse GRANT statements for procedures and functions
	grantProcRe := regexp.MustCompile(`GRANT\s+EXECUTE\s+ON\s+(?:PROCEDURE|FUNCTION)\s+(\w+)\s+TO\s+'([^']+)'@'([^']+)'(?:\s+WITH\s+GRANT\s+OPTION)?;`)
	grantProcMatches := grantProcRe.FindAllStringSubmatch(content, -1)

	for _, match := range grantProcMatches {
		if len(match) > 3 {
			perm := sqlmapper.Permission{
				Type:       "GRANT",
				Privileges: []string{"EXECUTE"},
				Object:     match[1],
				Grantee:    fmt.Sprintf("%s@%s", match[2], match[3]),
				WithGrant:  strings.Contains(match[0], "WITH GRANT OPTION"),
			}
			m.schema.Permissions = append(m.schema.Permissions, perm)
		}
	}

	// Parse REVOKE statements
	revokeRe := regexp.MustCompile(`REVOKE\s+(.*?)\s+ON\s+([.\w*]+)\s+FROM\s+'([^']+)'@'([^']+)';`)
	revokeMatches := revokeRe.FindAllStringSubmatch(content, -1)

	for _, match := range revokeMatches {
		if len(match) > 4 {
			// Split privileges and handle multiple privileges in one statement
			privilegeStr := strings.TrimSpace(match[1])
			var privileges []string
			if strings.ToUpper(privilegeStr) == "ALL PRIVILEGES" {
				privileges = []string{"ALL PRIVILEGES"}
			} else {
				// Handle comma-separated privileges and potential whitespace
				for _, priv := range strings.Split(privilegeStr, ",") {
					privs := strings.Fields(strings.TrimSpace(priv))
					privileges = append(privileges, privs...)
				}
			}

			perm := sqlmapper.Permission{
				Type:       "REVOKE",
				Privileges: privileges,
				Object:     match[2],
				Grantee:    fmt.Sprintf("%s@%s", match[3], match[4]),
			}
			m.schema.Permissions = append(m.schema.Permissions, perm)
		}
	}

	return nil
}

// generateTableSQL creates a CREATE TABLE statement for the given table.
// It includes column definitions, constraints, indexes, and table options.
//
// Parameters:
//   - table: The table structure to generate SQL for
//
// Returns:
//   - string: The generated CREATE TABLE statement
func (m *MySQL) generateTableSQL(table sqlmapper.Table) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", table.Name))

	// Columns
	for i, column := range table.Columns {
		result.WriteString("    " + m.generateColumnSQL(column))
		if i < len(table.Columns)-1 {
			result.WriteString(",")
		}
		result.WriteString("\n")
	}

	result.WriteString(");")
	return result.String()
}

// generateColumnSQL creates the SQL definition for a single column.
// It handles various column attributes including data type, length/precision,
// nullability, defaults, auto increment, and constraints.
//
// Parameters:
//   - column: The column structure to generate SQL for
//
// Returns:
//   - string: The generated column definition
func (m *MySQL) generateColumnSQL(column sqlmapper.Column) string {
	var parts []string
	parts = append(parts, column.Name)

	// Data type with length/precision
	if column.Length > 0 {
		if column.Scale > 0 {
			parts = append(parts, fmt.Sprintf("%s(%d,%d)", column.DataType, column.Length, column.Scale))
		} else {
			parts = append(parts, fmt.Sprintf("%s(%d)", column.DataType, column.Length))
		}
	} else {
		parts = append(parts, column.DataType)
	}

	// Handle AUTO_INCREMENT and PRIMARY KEY
	if column.AutoIncrement {
		parts = append(parts, "AUTO_INCREMENT")
	}
	if column.IsPrimaryKey {
		parts = append(parts, "PRIMARY KEY")
	} else if !column.IsNullable && column.DefaultValue == "" {
		parts = append(parts, "NOT NULL")
	}

	if column.DefaultValue != "" {
		if strings.Contains(column.DefaultValue, " ") ||
			strings.ToUpper(column.DefaultValue) == "CURRENT_TIMESTAMP" {
			parts = append(parts, "DEFAULT", column.DefaultValue)
		} else {
			parts = append(parts, "DEFAULT", fmt.Sprintf("'%s'", column.DefaultValue))
		}
	}

	if column.IsUnique && !column.IsPrimaryKey {
		parts = append(parts, "UNIQUE")
	}

	return strings.Join(parts, " ")
}

// generateIndexSQL creates a CREATE INDEX statement for the given index.
// It handles various index types including UNIQUE and regular indexes.
//
// Parameters:
//   - tableName: The name of the table the index belongs to
//   - index: The index structure to generate SQL for
//
// Returns:
//   - string: The generated CREATE INDEX statement
func (m *MySQL) generateIndexSQL(tableName string, index sqlmapper.Index) string {
	var result strings.Builder

	if index.IsUnique {
		result.WriteString("CREATE UNIQUE INDEX ")
	} else {
		result.WriteString("CREATE INDEX ")
	}

	result.WriteString(fmt.Sprintf("%s ON %s(%s);",
		index.Name,
		tableName,
		strings.Join(index.Columns, ", ")))

	return result.String()
}
