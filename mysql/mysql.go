package mysql

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/mstgnz/sqlporter"
)

type MySQL struct {
	schema *sqlporter.Schema
}

// NewMySQL creates a new MySQL parser instance
func NewMySQL() *MySQL {
	return &MySQL{
		schema: &sqlporter.Schema{},
	}
}

// Parse parses MySQL dump content
func (m *MySQL) Parse(content string) (*sqlporter.Schema, error) {
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

// Generate generates MySQL dump from schema
func (m *MySQL) Generate(schema *sqlporter.Schema) (string, error) {
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

// Helper functions for parsing

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

func (m *MySQL) parseSchemas(content string) error {
	// Parse CREATE DATABASE
	dbRe := regexp.MustCompile(`CREATE\s+DATABASE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`)
	if matches := dbRe.FindStringSubmatch(content); len(matches) > 1 {
		m.schema.Name = matches[1]
	}

	return nil
}

func (m *MySQL) parseTables(content string) error {
	re := regexp.MustCompile(`CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([.\w]+)\s*\((.*?)\)(?:\s+ENGINE\s*=\s*\w+)?(?:\s+DEFAULT\s+CHARSET\s*=\s*\w+)?(?:\s+COLLATE\s*=\s*\w+)?;`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 2 {
			tableName := match[1]
			columnDefs := match[2]

			table := sqlporter.Table{}

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

func (m *MySQL) parseColumnsAndConstraints(columnDefs string, table *sqlporter.Table) error {
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
				table.Constraints = append(table.Constraints, sqlporter.Constraint{
					Type:    "PRIMARY KEY",
					Columns: []string{column.Name},
				})
				column.IsPrimaryKey = true
				column.IsNullable = false
			}
			if strings.Contains(strings.ToUpper(def), "UNIQUE") {
				table.Constraints = append(table.Constraints, sqlporter.Constraint{
					Type:    "UNIQUE",
					Columns: []string{column.Name},
				})
				column.IsUnique = true
			}
			if strings.Contains(strings.ToUpper(def), "CHECK") {
				re := regexp.MustCompile(`CHECK\s*\((.*?)\)`)
				if matches := re.FindStringSubmatch(def); len(matches) > 1 {
					table.Constraints = append(table.Constraints, sqlporter.Constraint{
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

func (m *MySQL) parseColumn(def string) (sqlporter.Column, error) {
	parts := strings.Fields(def)
	if len(parts) < 2 {
		return sqlporter.Column{}, fmt.Errorf("invalid column definition: %s", def)
	}

	column := sqlporter.Column{
		Name:     parts[0],
		DataType: parts[1],
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

func (m *MySQL) parseConstraint(def string) (sqlporter.Constraint, error) {
	constraint := sqlporter.Constraint{}

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
					index := sqlporter.Index{
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

func (m *MySQL) parseViews(content string) error {
	viewRe := regexp.MustCompile(`CREATE(?:\s+OR\s+REPLACE)?\s+VIEW\s+([.\w]+)\s+AS\s+(.*?);`)
	viewMatches := viewRe.FindAllStringSubmatch(content, -1)

	for _, match := range viewMatches {
		if len(match) > 2 {
			viewName := match[1]
			view := sqlporter.View{
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

func (m *MySQL) parseFunctions(content string) error {
	// Parse functions
	funcRe := regexp.MustCompile(`CREATE\s+FUNCTION\s+([.\w]+)\s*\((.*?)\)\s+RETURNS\s+(\w+)\s+BEGIN\s+(.*?)\s+END`)
	funcMatches := funcRe.FindAllStringSubmatch(content, -1)

	for _, match := range funcMatches {
		if len(match) > 4 {
			functionName := match[1]
			function := sqlporter.Function{
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
						parameter := sqlporter.Parameter{
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
			function := sqlporter.Function{
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
						parameter := sqlporter.Parameter{
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

func (m *MySQL) parseTriggers(content string) error {
	triggerRe := regexp.MustCompile(`CREATE\s+TRIGGER\s+(\w+)\s+(BEFORE|AFTER)\s+(INSERT|UPDATE|DELETE)\s+ON\s+([.\w]+)\s+FOR\s+EACH\s+ROW\s+BEGIN\s+(.*?)\s+END`)
	triggerMatches := triggerRe.FindAllStringSubmatch(content, -1)

	for _, match := range triggerMatches {
		if len(match) > 5 {
			trigger := sqlporter.Trigger{
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

			perm := sqlporter.Permission{
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
			perm := sqlporter.Permission{
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

			perm := sqlporter.Permission{
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

// Helper functions for generating SQL

func (m *MySQL) generateTableSQL(table sqlporter.Table) string {
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

func (m *MySQL) generateColumnSQL(column sqlporter.Column) string {
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
	} else if !column.IsNullable {
		// Handle NOT NULL only if not PRIMARY KEY and explicitly set to false
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

func (m *MySQL) generateIndexSQL(tableName string, index sqlporter.Index) string {
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
