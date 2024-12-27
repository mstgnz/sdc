// Package sqlserver provides functionality for parsing and generating SQL Server database schemas.
// It implements the Parser interface for handling SQL Server specific SQL syntax and schema structures.
package sqlserver

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/mstgnz/sqlmapper"
)

// SQLServer represents a SQL Server parser implementation that handles parsing and generating
// SQL Server database schemas. It maintains an internal schema representation and provides
// methods for converting between SQL Server SQL and the common schema format.
type SQLServer struct {
	schema *sqlmapper.Schema
}

// NewSQLServer creates and initializes a new SQL Server parser instance.
// It returns a parser that can handle SQL Server specific SQL syntax and schema structures.
func NewSQLServer() *SQLServer {
	return &SQLServer{
		schema: &sqlmapper.Schema{},
	}
}

// Parse takes a SQL Server SQL dump content and parses it into a common schema structure.
// It processes various SQL Server objects including:
// - Tables with columns and constraints
// - Indexes (including UNIQUE indexes)
// - Views
// - Triggers
// - ALTER TABLE statements
//
// Parameters:
//   - content: The SQL Server SQL dump content to parse
//
// Returns:
//   - *sqlmapper.Schema: The parsed schema structure
//   - error: An error if parsing fails or if the content is empty
func (s *SQLServer) Parse(content string) (*sqlmapper.Schema, error) {
	if content == "" {
		return nil, errors.New("empty content")
	}

	// Split content into statements
	statements := s.splitStatements(content)

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		upperStmt := strings.ToUpper(stmt)

		switch {
		case strings.HasPrefix(upperStmt, "CREATE TABLE"):
			table, err := s.parseCreateTable(stmt)
			if err != nil {
				return nil, fmt.Errorf("error parsing CREATE TABLE: %v", err)
			}
			s.schema.Tables = append(s.schema.Tables, table)

		case strings.HasPrefix(upperStmt, "CREATE INDEX") || strings.HasPrefix(upperStmt, "CREATE UNIQUE INDEX"):
			if err := s.parseCreateIndex(stmt); err != nil {
				return nil, fmt.Errorf("error parsing CREATE INDEX: %v", err)
			}

		case strings.HasPrefix(upperStmt, "ALTER TABLE"):
			if err := s.parseAlterTable(stmt); err != nil {
				return nil, fmt.Errorf("error parsing ALTER TABLE: %v", err)
			}

		case strings.HasPrefix(upperStmt, "CREATE VIEW"):
			view, err := s.parseCreateView(stmt)
			if err != nil {
				return nil, fmt.Errorf("error parsing CREATE VIEW: %v", err)
			}
			s.schema.Views = append(s.schema.Views, view)

		case strings.HasPrefix(upperStmt, "CREATE TRIGGER"):
			trigger, err := s.parseCreateTrigger(stmt)
			if err != nil {
				return nil, fmt.Errorf("error parsing CREATE TRIGGER: %v", err)
			}
			s.schema.Triggers = append(s.schema.Triggers, trigger)
		}
	}

	return s.schema, nil
}

// splitStatements splits the SQL content into individual statements.
// It handles both semicolon and GO statement terminators.
func (s *SQLServer) splitStatements(content string) []string {
	var statements []string
	var currentStmt strings.Builder

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		if strings.EqualFold(line, "GO") {
			if currentStmt.Len() > 0 {
				statements = append(statements, strings.TrimSpace(currentStmt.String()))
				currentStmt.Reset()
			}
			continue
		}

		if strings.Contains(line, ";") {
			parts := strings.Split(line, ";")
			for i, part := range parts {
				if i < len(parts)-1 {
					currentStmt.WriteString(part)
					if currentStmt.Len() > 0 {
						statements = append(statements, strings.TrimSpace(currentStmt.String()))
						currentStmt.Reset()
					}
				} else if part != "" {
					if currentStmt.Len() > 0 {
						currentStmt.WriteString(" ")
					}
					currentStmt.WriteString(part)
				}
			}
		} else {
			if currentStmt.Len() > 0 {
				currentStmt.WriteString(" ")
			}
			currentStmt.WriteString(line)
		}
	}

	if currentStmt.Len() > 0 {
		statements = append(statements, strings.TrimSpace(currentStmt.String()))
	}

	return statements
}

// parseCreateTable parses a CREATE TABLE statement and returns a Table structure.
func (s *SQLServer) parseCreateTable(stmt string) (sqlmapper.Table, error) {
	table := sqlmapper.Table{}

	// Extract table name
	tableNameRegex := regexp.MustCompile(`CREATE\s+TABLE\s+(?:\[?dbo\]?\.)?\[?(\w+)\]?`)
	matches := tableNameRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		table.Name = matches[1]
	}

	// Extract column definitions
	columnsStr := stmt[strings.Index(stmt, "(")+1 : strings.LastIndex(stmt, ")")]
	columnDefs := strings.Split(columnsStr, ",")

	for _, colDef := range columnDefs {
		colDef = strings.TrimSpace(colDef)
		if colDef == "" {
			continue
		}

		// Handle table constraints
		if strings.HasPrefix(strings.ToUpper(colDef), "CONSTRAINT") ||
			strings.HasPrefix(strings.ToUpper(colDef), "PRIMARY KEY") ||
			strings.HasPrefix(strings.ToUpper(colDef), "FOREIGN KEY") ||
			strings.HasPrefix(strings.ToUpper(colDef), "UNIQUE") {
			constraint := s.parseConstraint(colDef)
			table.Constraints = append(table.Constraints, constraint)
			continue
		}

		// Parse column
		column := s.parseColumn(colDef)
		table.Columns = append(table.Columns, column)
	}

	return table, nil
}

// parseColumn parses a column definition and returns a Column structure.
func (s *SQLServer) parseColumn(def string) sqlmapper.Column {
	parts := strings.Fields(def)
	if len(parts) < 2 {
		return sqlmapper.Column{}
	}

	column := sqlmapper.Column{
		Name:       strings.Trim(parts[0], "[]"),
		DataType:   strings.ToUpper(parts[1]),
		IsNullable: true, // SQL Server columns are nullable by default
	}

	// Parse length/precision
	if strings.Contains(column.DataType, "(") {
		re := regexp.MustCompile(`(\w+)\((\d+)(?:,(\d+))?\)`)
		if matches := re.FindStringSubmatch(column.DataType); len(matches) > 2 {
			column.DataType = matches[1]
			fmt.Sscanf(matches[2], "%d", &column.Length)
			if len(matches) > 3 && matches[3] != "" {
				fmt.Sscanf(matches[3], "%d", &column.Scale)
			}
		}
	}

	def = strings.ToUpper(def)

	// Handle PRIMARY KEY
	if strings.Contains(def, "PRIMARY KEY") {
		column.IsPrimaryKey = true
		column.IsNullable = false
	}

	// Handle NOT NULL
	if strings.Contains(def, "NOT NULL") {
		column.IsNullable = false
	}

	// Handle UNIQUE
	if strings.Contains(def, "UNIQUE") {
		column.IsUnique = true
	}

	// Handle IDENTITY
	if strings.Contains(def, "IDENTITY") {
		column.AutoIncrement = true
	}

	// Handle DEFAULT
	if idx := strings.Index(def, "DEFAULT"); idx != -1 {
		restDef := def[idx+7:]
		endIdx := strings.IndexAny(restDef, " ,")
		if endIdx == -1 {
			endIdx = len(restDef)
		}
		column.DefaultValue = strings.Trim(restDef[:endIdx], "'()")
	}

	return column
}

// parseConstraint parses a table constraint definition and returns a Constraint structure.
func (s *SQLServer) parseConstraint(def string) sqlmapper.Constraint {
	constraint := sqlmapper.Constraint{}
	def = strings.TrimSpace(def)

	// Extract constraint name if exists
	if strings.HasPrefix(strings.ToUpper(def), "CONSTRAINT") {
		parts := strings.Fields(def)
		if len(parts) > 1 {
			constraint.Name = strings.Trim(parts[1], "[]")
			def = strings.Join(parts[2:], " ")
		}
	}

	switch {
	case strings.Contains(strings.ToUpper(def), "PRIMARY KEY"):
		constraint.Type = "PRIMARY KEY"
		columns := s.extractColumns(def, "PRIMARY KEY")
		constraint.Columns = columns

	case strings.Contains(strings.ToUpper(def), "FOREIGN KEY"):
		constraint.Type = "FOREIGN KEY"
		columns := s.extractColumns(def, "FOREIGN KEY")
		constraint.Columns = columns

		// Extract referenced table and columns
		refRegex := regexp.MustCompile(`REFERENCES\s+(?:\[?dbo\]?\.)?\[?(\w+)\]?\s*\((.*?)\)`)
		if matches := refRegex.FindStringSubmatch(def); len(matches) > 2 {
			constraint.RefTable = matches[1]
			constraint.RefColumns = s.splitAndTrim(matches[2])
		}

		// Extract ON DELETE rule
		if strings.Contains(strings.ToUpper(def), "ON DELETE CASCADE") {
			constraint.DeleteRule = "CASCADE"
		}

	case strings.Contains(strings.ToUpper(def), "UNIQUE"):
		constraint.Type = "UNIQUE"
		columns := s.extractColumns(def, "UNIQUE")
		constraint.Columns = columns

	case strings.Contains(strings.ToUpper(def), "CHECK"):
		constraint.Type = "CHECK"
		checkRegex := regexp.MustCompile(`CHECK\s*\((.*?)\)`)
		if matches := checkRegex.FindStringSubmatch(def); len(matches) > 1 {
			constraint.CheckExpression = matches[1]
		}
	}

	return constraint
}

// parseCreateIndex parses a CREATE INDEX statement and adds the index to the appropriate table.
func (s *SQLServer) parseCreateIndex(stmt string) error {
	isUnique := strings.HasPrefix(strings.ToUpper(stmt), "CREATE UNIQUE")

	// Extract index name and table name
	var indexRegex *regexp.Regexp
	if isUnique {
		indexRegex = regexp.MustCompile(`CREATE\s+UNIQUE\s+INDEX\s+\[?(\w+)\]?\s+ON\s+(?:\[?dbo\]?\.)?\[?(\w+)\]?`)
	} else {
		indexRegex = regexp.MustCompile(`CREATE\s+INDEX\s+\[?(\w+)\]?\s+ON\s+(?:\[?dbo\]?\.)?\[?(\w+)\]?`)
	}

	matches := indexRegex.FindStringSubmatch(stmt)
	if len(matches) < 3 {
		return fmt.Errorf("invalid CREATE INDEX statement: %s", stmt)
	}

	indexName := matches[1]
	tableName := matches[2]

	// Extract columns
	columnsRegex := regexp.MustCompile(`\((.*?)\)`)
	columnMatches := columnsRegex.FindStringSubmatch(stmt)
	if len(columnMatches) < 2 {
		return fmt.Errorf("no columns found in CREATE INDEX statement: %s", stmt)
	}

	columns := s.splitAndTrim(columnMatches[1])

	// Find the table and add the index
	for i, table := range s.schema.Tables {
		if table.Name == tableName {
			s.schema.Tables[i].Indexes = append(s.schema.Tables[i].Indexes, sqlmapper.Index{
				Name:     indexName,
				Columns:  columns,
				IsUnique: isUnique,
			})
			return nil
		}
	}

	return fmt.Errorf("table not found for index: %s", tableName)
}

// parseAlterTable parses an ALTER TABLE statement and modifies the appropriate table.
func (s *SQLServer) parseAlterTable(stmt string) error {
	// Extract table name
	tableNameRegex := regexp.MustCompile(`ALTER\s+TABLE\s+(?:\[?dbo\]?\.)?\[?(\w+)\]?`)
	matches := tableNameRegex.FindStringSubmatch(stmt)
	if len(matches) < 2 {
		return fmt.Errorf("invalid ALTER TABLE statement: %s", stmt)
	}

	tableName := matches[1]

	// Find the table
	var tableIndex = -1
	for i, table := range s.schema.Tables {
		if table.Name == tableName {
			tableIndex = i
			break
		}
	}

	if tableIndex == -1 {
		// Create new table if it doesn't exist
		s.schema.Tables = append(s.schema.Tables, sqlmapper.Table{Name: tableName})
		tableIndex = len(s.schema.Tables) - 1
	}

	// Handle different ALTER TABLE operations
	upperStmt := strings.ToUpper(stmt)
	switch {
	case strings.Contains(upperStmt, "ADD CONSTRAINT"):
		constraint := s.parseConstraint(stmt[strings.Index(strings.ToUpper(stmt), "ADD CONSTRAINT"):])
		s.schema.Tables[tableIndex].Constraints = append(s.schema.Tables[tableIndex].Constraints, constraint)

	case strings.Contains(upperStmt, "ADD COLUMN") || strings.Contains(upperStmt, "ADD "):
		// Extract column definition
		addIndex := strings.Index(strings.ToUpper(stmt), "ADD ")
		if addIndex == -1 {
			return fmt.Errorf("invalid ALTER TABLE ADD statement: %s", stmt)
		}

		colDef := strings.TrimSpace(stmt[addIndex+4:])
		if strings.HasPrefix(strings.ToUpper(colDef), "COLUMN") {
			colDef = strings.TrimSpace(colDef[6:])
		}

		column := s.parseColumn(colDef)
		s.schema.Tables[tableIndex].Columns = append(s.schema.Tables[tableIndex].Columns, column)
	}

	return nil
}

// parseCreateView parses a CREATE VIEW statement and returns a View structure.
func (s *SQLServer) parseCreateView(stmt string) (sqlmapper.View, error) {
	view := sqlmapper.View{}

	// Extract view name
	viewRegex := regexp.MustCompile(`CREATE\s+VIEW\s+(?:\[?dbo\]?\.)?\[?(\w+)\]?`)
	matches := viewRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		view.Name = matches[1]
	}

	// Extract view definition
	asIndex := strings.Index(strings.ToUpper(stmt), " AS ")
	if asIndex != -1 {
		view.Definition = strings.TrimSpace(stmt[asIndex+4:])
	}

	return view, nil
}

// parseCreateTrigger parses a CREATE TRIGGER statement and returns a Trigger structure.
func (s *SQLServer) parseCreateTrigger(stmt string) (sqlmapper.Trigger, error) {
	trigger := sqlmapper.Trigger{}

	// Extract trigger name
	triggerRegex := regexp.MustCompile(`CREATE\s+TRIGGER\s+\[?(\w+)\]?`)
	matches := triggerRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		trigger.Name = matches[1]
	}

	// Extract timing (AFTER/FOR/INSTEAD OF)
	if strings.Contains(strings.ToUpper(stmt), "AFTER") {
		trigger.Timing = "AFTER"
	} else if strings.Contains(strings.ToUpper(stmt), "INSTEAD OF") {
		trigger.Timing = "INSTEAD OF"
	} else {
		trigger.Timing = "FOR"
	}

	// Extract event (INSERT/UPDATE/DELETE)
	if strings.Contains(strings.ToUpper(stmt), "INSERT") {
		trigger.Event = "INSERT"
	} else if strings.Contains(strings.ToUpper(stmt), "UPDATE") {
		trigger.Event = "UPDATE"
	} else if strings.Contains(strings.ToUpper(stmt), "DELETE") {
		trigger.Event = "DELETE"
	}

	// Extract table name
	tableRegex := regexp.MustCompile(`ON\s+(?:\[?dbo\]?\.)?\[?(\w+)\]?`)
	matches = tableRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		trigger.Table = matches[1]
	}

	// Extract trigger body
	asIndex := strings.Index(strings.ToUpper(stmt), " AS ")
	if asIndex != -1 {
		trigger.Body = strings.TrimSpace(stmt[asIndex+4:])
	}

	return trigger, nil
}

// splitAndTrim splits a string by commas and trims whitespace and brackets from each part.
func (s *SQLServer) splitAndTrim(str string) []string {
	parts := strings.Split(str, ",")
	result := make([]string, len(parts))
	for i, part := range parts {
		result[i] = strings.Trim(strings.TrimSpace(part), "[]")
	}
	return result
}

// extractColumns extracts column names from a constraint definition.
func (s *SQLServer) extractColumns(def string, afterKeyword string) []string {
	regex := regexp.MustCompile(afterKeyword + `\s*\((.*?)\)`)
	if matches := regex.FindStringSubmatch(def); len(matches) > 1 {
		return s.splitAndTrim(matches[1])
	}
	return nil
}

// Generate creates a SQL Server SQL dump from a schema structure.
// It generates SQL statements for all database objects in the schema, including:
// - Tables with columns and constraints
// - Indexes (including UNIQUE indexes)
// - Views
// - Triggers
//
// Parameters:
//   - schema: The schema structure to convert to SQL Server SQL
//
// Returns:
//   - string: The generated SQL Server SQL statements
//   - error: An error if generation fails or if the schema is nil
func (s *SQLServer) Generate(schema *sqlmapper.Schema) (string, error) {
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
			result.WriteString(col.DataType)

			if col.Length > 0 {
				if col.Scale > 0 {
					result.WriteString(fmt.Sprintf("(%d,%d)", col.Length, col.Scale))
				} else {
					result.WriteString(fmt.Sprintf("(%d)", col.Length))
				}
			}

			if col.IsPrimaryKey {
				result.WriteString(" PRIMARY KEY")
			} else if !col.IsNullable {
				result.WriteString(" NOT NULL")
			}

			if col.IsUnique && !col.IsPrimaryKey {
				result.WriteString(" UNIQUE")
			}

			if col.AutoIncrement {
				result.WriteString(" IDENTITY(1,1)")
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
