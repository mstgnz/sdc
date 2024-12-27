// Package sqlite provides functionality for parsing and generating SQLite database schemas.
// It implements the Parser interface for handling SQLite specific SQL syntax and schema structures.
package sqlite

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/mstgnz/sqlmapper"
)

// SQLite represents a SQLite parser implementation that handles parsing and generating
// SQLite database schemas. It maintains an internal schema representation and provides
// methods for converting between SQLite SQL and the common schema format.
type SQLite struct {
	schema *sqlmapper.Schema
}

// NewSQLite creates and initializes a new SQLite parser instance.
// It returns a parser that can handle SQLite specific SQL syntax and schema structures.
func NewSQLite() *SQLite {
	return &SQLite{
		schema: &sqlmapper.Schema{},
	}
}

// Parse takes a SQLite SQL dump content and parses it into a common schema structure.
// It processes various SQLite objects including:
// - Tables with columns and constraints
// - Indexes (including UNIQUE indexes)
// - Views
// - Triggers
//
// Parameters:
//   - content: The SQLite SQL dump content to parse
//
// Returns:
//   - *sqlmapper.Schema: The parsed schema structure
//   - error: An error if parsing fails or if the content is empty
func (s *SQLite) Parse(content string) (*sqlmapper.Schema, error) {
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
func (s *SQLite) splitStatements(content string) []string {
	var statements []string
	var currentStmt strings.Builder

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
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
func (s *SQLite) parseCreateTable(stmt string) (sqlmapper.Table, error) {
	table := sqlmapper.Table{}

	// Extract table name
	tableNameRegex := regexp.MustCompile(`CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`)
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
func (s *SQLite) parseColumn(def string) sqlmapper.Column {
	parts := strings.Fields(def)
	if len(parts) < 2 {
		return sqlmapper.Column{}
	}

	column := sqlmapper.Column{
		Name:       parts[0],
		DataType:   s.getDataType(parts[1]),
		IsNullable: true, // SQLite columns are nullable by default
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

	// Handle DEFAULT
	if idx := strings.Index(def, "DEFAULT"); idx != -1 {
		restDef := def[idx+7:]
		endIdx := strings.IndexAny(restDef, " ,")
		if endIdx == -1 {
			endIdx = len(restDef)
		}
		column.DefaultValue = strings.Trim(restDef[:endIdx], "'")
	}

	return column
}

// parseConstraint parses a table constraint definition and returns a Constraint structure.
func (s *SQLite) parseConstraint(def string) sqlmapper.Constraint {
	constraint := sqlmapper.Constraint{}
	def = strings.TrimSpace(def)

	// Extract constraint name if exists
	if strings.HasPrefix(strings.ToUpper(def), "CONSTRAINT") {
		parts := strings.Fields(def)
		if len(parts) > 1 {
			constraint.Name = parts[1]
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
		refRegex := regexp.MustCompile(`REFERENCES\s+(\w+)\s*\((.*?)\)`)
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
	}

	return constraint
}

// parseCreateIndex parses a CREATE INDEX statement and adds the index to the appropriate table.
func (s *SQLite) parseCreateIndex(stmt string) error {
	isUnique := strings.HasPrefix(strings.ToUpper(stmt), "CREATE UNIQUE")

	// Extract index name and table name
	var indexRegex *regexp.Regexp
	if isUnique {
		indexRegex = regexp.MustCompile(`CREATE\s+UNIQUE\s+INDEX\s+(\w+)\s+ON\s+(\w+)`)
	} else {
		indexRegex = regexp.MustCompile(`CREATE\s+INDEX\s+(\w+)\s+ON\s+(\w+)`)
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

// parseCreateView parses a CREATE VIEW statement and returns a View structure.
func (s *SQLite) parseCreateView(stmt string) (sqlmapper.View, error) {
	view := sqlmapper.View{}

	// Extract view name
	viewRegex := regexp.MustCompile(`CREATE\s+(?:TEMP|TEMPORARY\s+)?VIEW\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`)
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
func (s *SQLite) parseCreateTrigger(stmt string) (sqlmapper.Trigger, error) {
	trigger := sqlmapper.Trigger{}

	// Extract trigger name
	triggerRegex := regexp.MustCompile(`CREATE\s+TRIGGER\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`)
	matches := triggerRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		trigger.Name = matches[1]
	}

	// Extract timing (BEFORE, AFTER, INSTEAD OF)
	if strings.Contains(strings.ToUpper(stmt), "BEFORE") {
		trigger.Timing = "BEFORE"
	} else if strings.Contains(strings.ToUpper(stmt), "AFTER") {
		trigger.Timing = "AFTER"
	} else if strings.Contains(strings.ToUpper(stmt), "INSTEAD OF") {
		trigger.Timing = "INSTEAD OF"
	}

	// Extract event (INSERT, UPDATE, DELETE)
	if strings.Contains(strings.ToUpper(stmt), "INSERT") {
		trigger.Event = "INSERT"
	} else if strings.Contains(strings.ToUpper(stmt), "UPDATE") {
		trigger.Event = "UPDATE"
	} else if strings.Contains(strings.ToUpper(stmt), "DELETE") {
		trigger.Event = "DELETE"
	}

	// Extract table name
	tableRegex := regexp.MustCompile(`ON\s+(\w+)`)
	matches = tableRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		trigger.Table = matches[1]
	}

	// Extract FOR EACH ROW
	trigger.ForEachRow = strings.Contains(strings.ToUpper(stmt), "FOR EACH ROW")

	// Extract trigger body
	beginIndex := strings.Index(strings.ToUpper(stmt), "BEGIN")
	endIndex := strings.LastIndex(strings.ToUpper(stmt), "END")
	if beginIndex != -1 && endIndex != -1 {
		trigger.Body = strings.TrimSpace(stmt[beginIndex : endIndex+3])
	}

	return trigger, nil
}

// splitAndTrim splits a string by commas and trims whitespace from each part.
func (s *SQLite) splitAndTrim(str string) []string {
	parts := strings.Split(str, ",")
	result := make([]string, len(parts))
	for i, part := range parts {
		result[i] = strings.TrimSpace(part)
	}
	return result
}

// extractColumns extracts column names from a constraint definition.
func (s *SQLite) extractColumns(def string, afterKeyword string) []string {
	regex := regexp.MustCompile(afterKeyword + `\s*\((.*?)\)`)
	if matches := regex.FindStringSubmatch(def); len(matches) > 1 {
		return s.splitAndTrim(matches[1])
	}
	return nil
}

// Generate creates a SQLite SQL dump from a schema structure.
// It generates SQL statements for all database objects in the schema, including:
// - Tables with columns and constraints
// - Indexes (including UNIQUE indexes)
// - Views
// - Triggers
//
// The generated SQL follows SQLite's specific syntax rules and type system.
// It handles SQLite-specific features such as INTEGER PRIMARY KEY for autoincrement.
//
// Parameters:
//   - schema: The schema structure to convert to SQLite SQL
//
// Returns:
//   - string: The generated SQLite SQL statements
//   - error: An error if generation fails or if the schema is nil
func (s *SQLite) Generate(schema *sqlmapper.Schema) (string, error) {
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

			if col.IsPrimaryKey && col.DataType == "INTEGER" {
				result.WriteString("INTEGER PRIMARY KEY")
			} else {
				result.WriteString(col.DataType)
				if col.DataType != "TEXT" && col.Length > 0 {
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

// getDataType maps common data types to SQLite data types.
// SQLite has a dynamic type system with storage classes:
// - NULL
// - INTEGER (1, 2, 3, 4, 6, or 8 bytes)
// - REAL (8-byte IEEE floating point)
// - TEXT (UTF-8, UTF-16BE or UTF-16LE)
// - BLOB (binary data)
//
// Parameters:
//   - dataType: The input data type to map to SQLite storage class
//
// Returns:
//   - string: The corresponding SQLite storage class
func (s *SQLite) getDataType(dataType string) string {
	dataType = strings.ToUpper(dataType)
	if strings.Contains(dataType, "(") {
		parts := strings.Split(dataType, "(")
		dataType = parts[0]
	}

	switch dataType {
	case "INT", "INTEGER", "BIGINT", "SMALLINT", "TINYINT":
		return "INTEGER"
	case "VARCHAR", "CHAR", "TEXT", "NVARCHAR", "NCHAR":
		return "TEXT"
	case "DECIMAL", "NUMERIC", "FLOAT", "DOUBLE", "REAL":
		return "REAL"
	case "BLOB", "BINARY", "VARBINARY":
		return "BLOB"
	default:
		return dataType
	}
}
