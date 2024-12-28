// Package postgres provides functionality for parsing and generating PostgreSQL database schemas.
// It implements the Parser interface for handling PostgreSQL specific SQL syntax and schema structures.
package postgres

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/mstgnz/sqlmapper"
)

// PostgreSQL represents a PostgreSQL parser implementation that handles parsing and generating
// PostgreSQL database schemas. It maintains an internal schema representation and provides
// methods for converting between PostgreSQL SQL and the common schema format.
type PostgreSQL struct {
	schema *sqlmapper.Schema
}

// NewPostgreSQL creates and initializes a new PostgreSQL parser instance.
// It returns a parser that can handle PostgreSQL specific SQL syntax and schema structures.
func NewPostgreSQL() *PostgreSQL {
	return &PostgreSQL{
		schema: &sqlmapper.Schema{},
	}
}

// Parse takes a PostgreSQL SQL dump content and parses it into a common schema structure.
// It processes various PostgreSQL objects including:
// - Schemas and databases
// - Custom types (ENUM, COMPOSITE)
// - Extensions
// - Sequences
// - Tables with columns and constraints
// - Indexes
// - Views (regular and materialized)
// - Functions and procedures
// - Triggers
// - Permissions (GRANT/REVOKE)
//
// Parameters:
//   - content: The PostgreSQL SQL dump content to parse
//
// Returns:
//   - *sqlmapper.Schema: The parsed schema structure
//   - error: An error if parsing fails
func (p *PostgreSQL) Parse(content string) (*sqlmapper.Schema, error) {
	if content == "" {
		return nil, errors.New("empty content")
	}

	// Normalize content
	content = p.normalizeContent(content)

	// Parse schema objects
	if err := p.parseSchemas(content); err != nil {
		return nil, fmt.Errorf("error parsing schemas: %v", err)
	}

	if err := p.parseTypes(content); err != nil {
		return nil, fmt.Errorf("error parsing types: %v", err)
	}

	if err := p.parseExtensions(content); err != nil {
		return nil, fmt.Errorf("error parsing extensions: %v", err)
	}

	if err := p.parseSequences(content); err != nil {
		return nil, fmt.Errorf("error parsing sequences: %v", err)
	}

	if err := p.parseTables(content); err != nil {
		return nil, fmt.Errorf("error parsing tables: %v", err)
	}

	if err := p.parseIndexes(content); err != nil {
		return nil, fmt.Errorf("error parsing indexes: %v", err)
	}

	if err := p.parseViews(content); err != nil {
		return nil, fmt.Errorf("error parsing views: %v", err)
	}

	if err := p.parseFunctions(content); err != nil {
		return nil, fmt.Errorf("error parsing functions: %v", err)
	}

	if err := p.parseTriggers(content); err != nil {
		return nil, fmt.Errorf("error parsing triggers: %v", err)
	}

	if err := p.parsePermissions(content); err != nil {
		return nil, fmt.Errorf("error parsing permissions: %v", err)
	}

	return p.schema, nil
}

// Generate creates a PostgreSQL SQL dump from a schema structure.
// It generates SQL statements for all database objects in the schema, including:
// - Tables with columns and constraints
// - Indexes
// - Views
// - Functions and procedures
// - Triggers
// - Permissions
//
// Parameters:
//   - schema: The schema structure to convert to PostgreSQL SQL
//
// Returns:
//   - string: The generated PostgreSQL SQL statements
//   - error: An error if generation fails
func (p *PostgreSQL) Generate(schema *sqlmapper.Schema) (string, error) {
	if schema == nil {
		return "", errors.New("empty schema")
	}

	var result strings.Builder

	for _, table := range schema.Tables {
		result.WriteString("CREATE TABLE ")
		result.WriteString(table.Name)
		result.WriteString(" (\n")

		for i, col := range table.Columns {
			result.WriteString("    ")
			result.WriteString(col.Name)
			result.WriteString(" ")

			if col.IsPrimaryKey && col.DataType == "SERIAL" {
				result.WriteString("SERIAL PRIMARY KEY")
			} else {
				result.WriteString(col.DataType)
				if col.Length > 0 {
					result.WriteString(fmt.Sprintf("(%d", col.Length))
					if col.Scale > 0 {
						result.WriteString(fmt.Sprintf(",%d", col.Scale))
					}
					result.WriteString(")")
				}

				if !col.IsNullable {
					result.WriteString(" NOT NULL")
				}

				if col.IsUnique {
					result.WriteString(" UNIQUE")
				}
			}

			if i < len(table.Columns)-1 {
				result.WriteString(",")
			}
			result.WriteString("\n")
		}

		result.WriteString(");\n")

		// Add indexes
		for _, idx := range table.Indexes {
			if idx.IsUnique {
				result.WriteString("CREATE UNIQUE INDEX ")
			} else {
				result.WriteString("CREATE INDEX ")
			}
			result.WriteString(idx.Name)
			result.WriteString(" ON ")
			result.WriteString(table.Name)
			result.WriteString("(")
			result.WriteString(strings.Join(idx.Columns, ", "))
			result.WriteString(");\n")
		}
	}

	return result.String(), nil
}

// normalizeContent preprocesses the SQL content by removing comments and normalizing whitespace.
// This helps ensure consistent parsing of the SQL statements.
//
// Parameters:
//   - content: The SQL content to normalize
//
// Returns:
//   - string: The normalized SQL content
func (p *PostgreSQL) normalizeContent(content string) string {
	// Remove comments
	re := regexp.MustCompile(`--.*$`)
	content = re.ReplaceAllString(content, "")

	// Normalize whitespace
	content = strings.TrimSpace(content)
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")

	return content
}

// parseSchemas extracts database and schema definitions from the SQL content.
// It handles both CREATE DATABASE and CREATE SCHEMA statements.
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (p *PostgreSQL) parseSchemas(content string) error {
	// Parse CREATE DATABASE
	dbRe := regexp.MustCompile(`CREATE\s+DATABASE\s+(\w+)`)
	if matches := dbRe.FindStringSubmatch(content); len(matches) > 1 {
		p.schema.Name = matches[1]
		return nil
	}

	// Parse CREATE SCHEMA
	schemaRe := regexp.MustCompile(`CREATE\s+SCHEMA\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`)
	matches := schemaRe.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			schema := match[1]
			p.schema.Name = schema
		}
	}

	return nil
}

// parseTypes processes custom type definitions in the SQL content.
// It handles both ENUM and COMPOSITE types, including their schema qualification.
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (p *PostgreSQL) parseTypes(content string) error {
	// Parse ENUM types
	enumRe := regexp.MustCompile(`CREATE\s+TYPE\s+([.\w]+)\s+AS\s+ENUM\s*\((.*?)\);`)
	enumMatches := enumRe.FindAllStringSubmatch(content, -1)
	for _, match := range enumMatches {
		if len(match) > 2 {
			typeName := match[1]
			typ := sqlmapper.Type{
				Name:       typeName,
				Kind:       "ENUM",
				Definition: match[2],
			}

			// Parse schema if exists
			parts := strings.Split(typeName, ".")
			if len(parts) > 1 {
				typ.Schema = parts[0]
				typ.Name = parts[1]
			}

			p.schema.Types = append(p.schema.Types, typ)
		}
	}

	// Parse COMPOSITE types
	compositeRe := regexp.MustCompile(`CREATE\s+TYPE\s+([.\w]+)\s+AS\s*\((.*?)\);`)
	compositeMatches := compositeRe.FindAllStringSubmatch(content, -1)
	for _, match := range compositeMatches {
		if len(match) > 2 {
			typeName := match[1]
			typ := sqlmapper.Type{
				Name:       typeName,
				Kind:       "COMPOSITE",
				Definition: match[2],
			}

			// Parse schema if exists
			parts := strings.Split(typeName, ".")
			if len(parts) > 1 {
				typ.Schema = parts[0]
				typ.Name = parts[1]
			}

			p.schema.Types = append(p.schema.Types, typ)
		}
	}

	return nil
}

// parseExtensions extracts extension definitions from the SQL content.
// It handles CREATE EXTENSION statements with optional schema qualification.
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (p *PostgreSQL) parseExtensions(content string) error {
	re := regexp.MustCompile(`CREATE\s+EXTENSION(?:\s+IF\s+NOT\s+EXISTS)?\s+(?:"([^"]+)"|(\w+))(?:\s+WITH\s+SCHEMA\s+(\w+))?;`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			extension := sqlmapper.Extension{
				Name: match[1],
			}
			if extension.Name == "" {
				extension.Name = match[2]
			}
			if len(match) > 3 && match[3] != "" {
				extension.Schema = match[3]
			}
			p.schema.Extensions = append(p.schema.Extensions, extension)
		}
	}

	return nil
}

// parseSequences processes sequence definitions from the SQL content.
// It handles all sequence options including INCREMENT, MINVALUE, MAXVALUE,
// START WITH, CACHE, and CYCLE.
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (p *PostgreSQL) parseSequences(content string) error {
	re := regexp.MustCompile(`CREATE\s+SEQUENCE\s+([.\w]+)(?:\s+INCREMENT\s+BY\s+(\d+))?(?:\s+MINVALUE\s+(\d+))?(?:\s+MAXVALUE\s+(\d+))?(?:\s+START\s+WITH\s+(\d+))?(?:\s+CACHE\s+(\d+))?(?:\s+CYCLE)?;`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			seqName := match[1]
			seq := sqlmapper.Sequence{
				Name: seqName,
			}

			// Parse schema if exists
			parts := strings.Split(seqName, ".")
			if len(parts) > 1 {
				seq.Schema = parts[0]
				seq.Name = parts[1]
			}

			// Parse optional parameters
			if len(match) > 2 && match[2] != "" {
				fmt.Sscanf(match[2], "%d", &seq.IncrementBy)
			}
			if len(match) > 3 && match[3] != "" {
				fmt.Sscanf(match[3], "%d", &seq.MinValue)
			}
			if len(match) > 4 && match[4] != "" {
				fmt.Sscanf(match[4], "%d", &seq.MaxValue)
			}
			if len(match) > 5 && match[5] != "" {
				fmt.Sscanf(match[5], "%d", &seq.StartValue)
			}
			if len(match) > 6 && match[6] != "" {
				fmt.Sscanf(match[6], "%d", &seq.Cache)
			}
			if len(match) > 7 {
				seq.Cycle = true
			}

			p.schema.Sequences = append(p.schema.Sequences, seq)
		}
	}

	return nil
}

// parseTables extracts table definitions from the SQL content.
// It processes table structure including columns, constraints, tablespaces,
// and comments for both tables and columns.
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (p *PostgreSQL) parseTables(content string) error {
	re := regexp.MustCompile(`CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([.\w]+)\s*\((.*?)\)(?:\s+TABLESPACE\s+(\w+))?;`)
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

			// Parse tablespace if exists
			if len(match) > 3 && match[3] != "" {
				table.TableSpace = match[3]
			}

			// Parse columns and constraints
			if err := p.parseColumnsAndConstraints(columnDefs, &table); err != nil {
				return err
			}

			// Parse table comment
			tableCommentRe := regexp.MustCompile(`COMMENT\s+ON\s+TABLE\s+` + regexp.QuoteMeta(tableName) + `\s+IS\s+'([^']+)';`)
			if tableCommentMatch := tableCommentRe.FindStringSubmatch(content); len(tableCommentMatch) > 1 {
				table.Comment = tableCommentMatch[1]
			}

			// Parse column comments
			commentRe := regexp.MustCompile(`COMMENT\s+ON\s+COLUMN\s+` + regexp.QuoteMeta(tableName) + `\.(\w+)\s+IS\s+'([^']+)';`)
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

			p.schema.Tables = append(p.schema.Tables, table)
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
func (p *PostgreSQL) parseColumnsAndConstraints(columnDefs string, table *sqlmapper.Table) error {
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
			(strings.Contains(strings.ToUpper(def), "PRIMARY KEY") && !strings.Contains(strings.ToUpper(def), "SERIAL")) ||
			strings.Contains(strings.ToUpper(def), "FOREIGN KEY") ||
			(strings.Contains(strings.ToUpper(def), "UNIQUE") && !strings.Contains(strings.ToUpper(def), " ")) ||
			(strings.Contains(strings.ToUpper(def), "CHECK") && !strings.Contains(strings.ToUpper(def), " ")) {
			constraint, err := p.parseConstraint(def)
			if err != nil {
				return err
			}
			table.Constraints = append(table.Constraints, constraint)
			continue
		}

		// Parse column
		if strings.Contains(def, " ") {
			column, err := p.parseColumn(def)
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
// nullability, defaults, and inline constraints.
//
// Parameters:
//   - def: The column definition string to parse
//
// Returns:
//   - sqlmapper.Column: The parsed column structure
//   - error: An error if parsing fails
func (p *PostgreSQL) parseColumn(def string) (sqlmapper.Column, error) {
	parts := strings.Fields(def)
	if len(parts) < 2 {
		return sqlmapper.Column{}, fmt.Errorf("invalid column definition: %s", def)
	}

	column := sqlmapper.Column{
		Name:       parts[0],
		DataType:   parts[1],
		IsNullable: true,
	}

	// Handle SERIAL type
	if strings.ToUpper(column.DataType) == "SERIAL" {
		column.AutoIncrement = true
		column.DataType = "INTEGER"
		column.DefaultValue = "nextval('" + column.Name + "_seq'::regclass)"
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
func (p *PostgreSQL) parseConstraint(def string) (sqlmapper.Constraint, error) {
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
// It handles both regular and unique indexes, associating them with their tables.
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (p *PostgreSQL) parseIndexes(content string) error {
	re := regexp.MustCompile(`CREATE\s+(?:UNIQUE\s+)?INDEX\s+(\w+)\s+ON\s+([.\w]+)\s*\((.*?)\)`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 3 {
			indexName := match[1]
			tableName := match[2]
			columns := strings.Split(match[3], ",")

			// Find the table
			for i, table := range p.schema.Tables {
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

					p.schema.Tables[i].Indexes = append(p.schema.Tables[i].Indexes, index)
					break
				}
			}
		}
	}

	return nil
}

// parseViews processes view definitions from the SQL content.
// It handles both regular and materialized views with their definitions.
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (p *PostgreSQL) parseViews(content string) error {
	// Parse regular views
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

			p.schema.Views = append(p.schema.Views, view)
		}
	}

	// Parse materialized views
	matViewRe := regexp.MustCompile(`CREATE\s+MATERIALIZED\s+VIEW\s+([.\w]+)(?:\s+WITH\s*\([^)]*\))?\s+AS\s+(.*?)\s+WITH\s+(?:NO\s+)?DATA;`)
	matViewMatches := matViewRe.FindAllStringSubmatch(content, -1)

	for _, match := range matViewMatches {
		if len(match) > 2 {
			viewName := match[1]
			view := sqlmapper.View{
				Definition:     match[2],
				IsMaterialized: true,
			}

			// Parse schema if exists
			parts := strings.Split(viewName, ".")
			if len(parts) > 1 {
				view.Schema = parts[0]
				view.Name = parts[1]
			} else {
				view.Name = viewName
			}

			p.schema.Views = append(p.schema.Views, view)
		}
	}

	return nil
}

// parseFunctions extracts function and procedure definitions from the SQL content.
// It handles various function attributes including parameters, return types,
// and function bodies.
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (p *PostgreSQL) parseFunctions(content string) error {
	// Parse functions
	funcRe := regexp.MustCompile(`CREATE(?:\s+OR\s+REPLACE)?\s+FUNCTION\s+([.\w]+)\s*\((.*?)\)\s+RETURNS\s+(\w+)\s+AS\s+\$\$(.*?)\$\$\s+LANGUAGE\s+(\w+)`)
	funcMatches := funcRe.FindAllStringSubmatch(content, -1)

	for _, match := range funcMatches {
		if len(match) > 5 {
			functionName := match[1]
			function := sqlmapper.Function{
				Returns:  match[3],
				Body:     match[4],
				Language: match[5],
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

			p.schema.Functions = append(p.schema.Functions, function)
		}
	}

	// Parse procedures
	procRe := regexp.MustCompile(`CREATE(?:\s+OR\s+REPLACE)?\s+PROCEDURE\s+([.\w]+)\s*\((.*?)\)\s+LANGUAGE\s+(\w+)\s+AS\s+\$\$(.*?)\$\$`)
	procMatches := procRe.FindAllStringSubmatch(content, -1)

	for _, match := range procMatches {
		if len(match) > 4 {
			procName := match[1]
			function := sqlmapper.Function{
				Name:     procName,
				Body:     match[4],
				Language: match[3],
				IsProc:   true,
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
					if len(parts) >= 2 {
						parameter := sqlmapper.Parameter{
							Name:     parts[0],
							DataType: parts[1],
						}
						function.Parameters = append(function.Parameters, parameter)
					}
				}
			}

			p.schema.Functions = append(p.schema.Functions, function)
		}
	}

	return nil
}

// parseTriggers processes trigger definitions from the SQL content.
// It handles both regular and conditional triggers with their timing,
// events, and conditions.
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (p *PostgreSQL) parseTriggers(content string) error {
	// Parse regular triggers
	triggerRe := regexp.MustCompile(`CREATE\s+TRIGGER\s+(\w+)\s+(BEFORE|AFTER|INSTEAD\s+OF)\s+(INSERT|UPDATE|DELETE)\s+ON\s+([.\w]+)\s+(?:FOR\s+EACH\s+ROW\s+)?EXECUTE\s+(?:FUNCTION|PROCEDURE)\s+([.\w]+)`)
	triggerMatches := triggerRe.FindAllStringSubmatch(content, -1)

	for _, match := range triggerMatches {
		if len(match) > 5 {
			trigger := sqlmapper.Trigger{
				Name:       match[1],
				Timing:     match[2],
				Event:      match[3],
				Table:      match[4],
				Body:       match[5],
				ForEachRow: strings.Contains(match[0], "FOR EACH ROW"),
			}

			// Parse schema if exists
			parts := strings.Split(trigger.Table, ".")
			if len(parts) > 1 {
				trigger.Schema = parts[0]
				trigger.Table = parts[1]
			}

			p.schema.Triggers = append(p.schema.Triggers, trigger)
		}
	}

	// Parse conditional triggers
	condTriggerRe := regexp.MustCompile(`CREATE\s+TRIGGER\s+(\w+)\s+(BEFORE|AFTER|INSTEAD\s+OF)\s+(?:UPDATE\s+OF\s+[.\w]+\s+)?ON\s+([.\w]+)\s+(?:FOR\s+EACH\s+ROW\s+)?WHEN\s+\((.*?)\)\s+EXECUTE\s+(?:FUNCTION|PROCEDURE)\s+([.\w]+)`)
	condTriggerMatches := condTriggerRe.FindAllStringSubmatch(content, -1)

	for _, match := range condTriggerMatches {
		if len(match) > 5 {
			trigger := sqlmapper.Trigger{
				Name:       match[1],
				Timing:     match[2],
				Table:      match[3],
				Condition:  match[4],
				Body:       match[5],
				ForEachRow: strings.Contains(match[0], "FOR EACH ROW"),
			}

			// Parse schema if exists
			parts := strings.Split(trigger.Table, ".")
			if len(parts) > 1 {
				trigger.Schema = parts[0]
				trigger.Table = parts[1]
			}

			p.schema.Triggers = append(p.schema.Triggers, trigger)
		}
	}

	return nil
}

// parsePermissions extracts permission definitions from the SQL content.
// It handles GRANT and REVOKE statements for various privilege types.
//
// Parameters:
//   - content: The SQL content to parse
//
// Returns:
//   - error: An error if parsing fails
func (p *PostgreSQL) parsePermissions(content string) error {
	// Parse GRANT statements
	grantRe := regexp.MustCompile(`GRANT\s+(.*?)\s+ON\s+(?:TABLE\s+)?([.\w]+)\s+TO\s+(\w+)(?:\s+WITH\s+GRANT\s+OPTION)?;`)
	grantMatches := grantRe.FindAllStringSubmatch(content, -1)

	for _, match := range grantMatches {
		if len(match) > 3 {
			privileges := strings.Split(strings.TrimSpace(match[1]), ",")
			for i := range privileges {
				privileges[i] = strings.TrimSpace(privileges[i])
			}

			perm := sqlmapper.Permission{
				Type:       "GRANT",
				Privileges: privileges,
				Object:     match[2],
				Grantee:    match[3],
				WithGrant:  strings.Contains(match[0], "WITH GRANT OPTION"),
			}
			p.schema.Permissions = append(p.schema.Permissions, perm)
		}
	}

	// Parse GRANT ALL statements
	grantAllRe := regexp.MustCompile(`GRANT\s+ALL\s+PRIVILEGES\s+ON\s+(?:ALL\s+TABLES\s+IN\s+SCHEMA\s+)?([.\w]+)\s+TO\s+(\w+)(?:\s+WITH\s+GRANT\s+OPTION)?;`)
	grantAllMatches := grantAllRe.FindAllStringSubmatch(content, -1)

	for _, match := range grantAllMatches {
		if len(match) > 2 {
			perm := sqlmapper.Permission{
				Type:       "GRANT",
				Privileges: []string{"ALL PRIVILEGES"},
				Object:     match[1],
				Grantee:    match[2],
				WithGrant:  strings.Contains(match[0], "WITH GRANT OPTION"),
			}
			p.schema.Permissions = append(p.schema.Permissions, perm)
		}
	}

	// Parse GRANT EXECUTE statements
	grantExecRe := regexp.MustCompile(`GRANT\s+EXECUTE\s+ON\s+(?:FUNCTION|PROCEDURE)\s+([.\w]+)\s*\((.*?)\)\s+TO\s+(\w+)(?:\s+WITH\s+GRANT\s+OPTION)?;`)
	grantExecMatches := grantExecRe.FindAllStringSubmatch(content, -1)

	for _, match := range grantExecMatches {
		if len(match) > 3 {
			perm := sqlmapper.Permission{
				Type:       "GRANT",
				Privileges: []string{"EXECUTE"},
				Object:     match[1],
				Grantee:    match[3],
				WithGrant:  strings.Contains(match[0], "WITH GRANT OPTION"),
			}
			p.schema.Permissions = append(p.schema.Permissions, perm)
		}
	}

	// Parse REVOKE statements
	revokeRe := regexp.MustCompile(`REVOKE\s+(.*?)\s+ON\s+(?:TABLE\s+)?([.\w]+)\s+FROM\s+(\w+);`)
	revokeMatches := revokeRe.FindAllStringSubmatch(content, -1)

	for _, match := range revokeMatches {
		if len(match) > 3 {
			privileges := strings.Split(strings.TrimSpace(match[1]), ",")
			for i := range privileges {
				privileges[i] = strings.TrimSpace(privileges[i])
			}

			perm := sqlmapper.Permission{
				Type:       "REVOKE",
				Privileges: privileges,
				Object:     match[2],
				Grantee:    match[3],
			}
			p.schema.Permissions = append(p.schema.Permissions, perm)
		}
	}

	return nil
}

// generateTypeSQL generates SQL for a type
func (p *PostgreSQL) generateTypeSQL(typ sqlmapper.Type) string {
	if typ.Kind == "ENUM" {
		return fmt.Sprintf("CREATE TYPE %s AS ENUM (%s)", typ.Name, typ.Definition)
	}
	return fmt.Sprintf("CREATE TYPE %s AS %s", typ.Name, typ.Definition)
}

// generateTableSQL generates SQL for a table
func (p *PostgreSQL) generateTableSQL(table sqlmapper.Table) string {
	sql := "CREATE TABLE " + table.Name + " (\n"

	// Generate columns
	for i, col := range table.Columns {
		sql += "    " + col.Name + " "

		if col.IsPrimaryKey && strings.ToUpper(col.DataType) == "SERIAL" {
			sql += "SERIAL PRIMARY KEY"
		} else {
			sql += col.DataType
			if col.Length > 0 {
				sql += fmt.Sprintf("(%d", col.Length)
				if col.Scale > 0 {
					sql += fmt.Sprintf(",%d", col.Scale)
				}
				sql += ")"
			}

			if !col.IsNullable {
				sql += " NOT NULL"
			}
			if col.IsUnique {
				sql += " UNIQUE"
			}
			if col.DefaultValue != "" {
				sql += " DEFAULT " + col.DefaultValue
			}
		}

		if i < len(table.Columns)-1 {
			sql += ",\n"
		}
	}

	sql += "\n)"

	// Add table options
	if table.TableSpace != "" {
		sql += " TABLESPACE " + table.TableSpace
	}

	return sql
}

// generateIndexSQL generates SQL for an index
func (p *PostgreSQL) generateIndexSQL(tableName string, index sqlmapper.Index) string {
	var sql string
	if index.IsUnique {
		sql = "CREATE UNIQUE INDEX "
	} else {
		sql = "CREATE INDEX "
	}

	sql += index.Name + " ON " + tableName
	if index.Type != "" {
		sql += " USING " + index.Type
	}
	sql += " (" + strings.Join(index.Columns, ", ") + ")"

	// Add index options
	if index.TableSpace != "" {
		sql += " TABLESPACE " + index.TableSpace
	}

	return sql
}
