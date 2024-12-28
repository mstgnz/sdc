// Package sqlserver provides functionality for parsing and generating SQL Server database schemas.
// It implements the Parser interface for handling SQL Server specific SQL syntax and schema structures.
package sqlserver

import (
	"bytes"
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
	buf    *bytes.Buffer // Buffer for parsing operations
}

// NewSQLServer creates and initializes a new SQL Server parser instance.
// It returns a parser that can handle SQL Server specific SQL syntax and schema structures.
func NewSQLServer() sqlmapper.Database {
	return &SQLServer{
		schema: &sqlmapper.Schema{},
		buf:    bytes.NewBuffer(nil),
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

	// Convert content to bytes
	contentBytes := []byte(content)

	// Split content into statements
	statements := s.splitStatements(contentBytes)

	for _, stmt := range statements {
		stmt = bytes.TrimSpace(stmt)
		if len(stmt) == 0 {
			continue
		}

		upperStmt := bytes.ToUpper(stmt)

		switch {
		case bytes.HasPrefix(upperStmt, []byte("CREATE TABLE")):
			table, err := s.parseCreateTable(stmt)
			if err != nil {
				return nil, fmt.Errorf("error parsing CREATE TABLE: %v", err)
			}
			s.schema.Tables = append(s.schema.Tables, table)

		case bytes.HasPrefix(upperStmt, []byte("CREATE INDEX")) || bytes.HasPrefix(upperStmt, []byte("CREATE UNIQUE INDEX")):
			if err := s.parseCreateIndex(stmt); err != nil {
				return nil, fmt.Errorf("error parsing CREATE INDEX: %v", err)
			}

		case bytes.HasPrefix(upperStmt, []byte("ALTER TABLE")):
			if err := s.parseAlterTable(stmt); err != nil {
				return nil, fmt.Errorf("error parsing ALTER TABLE: %v", err)
			}

		case bytes.HasPrefix(upperStmt, []byte("CREATE VIEW")):
			view, err := s.parseCreateView(stmt)
			if err != nil {
				return nil, fmt.Errorf("error parsing CREATE VIEW: %v", err)
			}
			s.schema.Views = append(s.schema.Views, view)

		case bytes.HasPrefix(upperStmt, []byte("CREATE TRIGGER")):
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
func (s *SQLServer) splitStatements(content []byte) [][]byte {
	var statements [][]byte
	s.buf.Reset()

	// First split by GO statements
	goBlocks := bytes.Split(content, []byte("GO"))

	for _, block := range goBlocks {
		block = bytes.TrimSpace(block)
		if len(block) == 0 {
			continue
		}

		// Then split each GO block by semicolons
		stmts := bytes.Split(block, []byte(";"))
		for _, stmt := range stmts {
			stmt = bytes.TrimSpace(stmt)
			if len(stmt) == 0 || bytes.HasPrefix(stmt, []byte("--")) {
				continue
			}
			statements = append(statements, stmt)
		}
	}

	return statements
}

// parseCreateTable parses a CREATE TABLE statement and returns a Table structure.
func (s *SQLServer) parseCreateTable(stmt []byte) (sqlmapper.Table, error) {
	table := sqlmapper.Table{}

	// Extract table name using bytes.Index and bytes.LastIndex
	startIdx := bytes.Index(stmt, []byte("TABLE")) + 5
	endIdx := bytes.Index(stmt[startIdx:], []byte("("))
	if endIdx == -1 {
		return table, fmt.Errorf("invalid CREATE TABLE statement")
	}
	tableName := bytes.TrimSpace(stmt[startIdx : startIdx+endIdx])

	// Remove schema prefix if exists
	if idx := bytes.LastIndex(tableName, []byte(".")); idx != -1 {
		tableName = tableName[idx+1:]
	}
	// Remove brackets if exists
	tableName = bytes.Trim(tableName, "[]")
	table.Name = string(tableName)

	// Extract column definitions
	columnsBytes := stmt[startIdx+endIdx+1 : bytes.LastIndex(stmt, []byte(")"))]
	columnDefs := bytes.Split(columnsBytes, []byte(","))

	for _, colDef := range columnDefs {
		colDef = bytes.TrimSpace(colDef)
		if len(colDef) == 0 {
			continue
		}

		// Handle table constraints
		upperColDef := bytes.ToUpper(colDef)
		if bytes.HasPrefix(upperColDef, []byte("CONSTRAINT")) ||
			bytes.HasPrefix(upperColDef, []byte("PRIMARY KEY")) ||
			bytes.HasPrefix(upperColDef, []byte("FOREIGN KEY")) ||
			bytes.HasPrefix(upperColDef, []byte("UNIQUE")) {
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
func (s *SQLServer) parseColumn(def []byte) sqlmapper.Column {
	parts := bytes.Fields(def)
	if len(parts) < 2 {
		return sqlmapper.Column{}
	}

	column := sqlmapper.Column{
		Name:       string(bytes.Trim(parts[0], "[]")),
		DataType:   string(bytes.ToUpper(parts[1])),
		IsNullable: true, // SQL Server columns are nullable by default
	}

	// Parse length/precision
	if bytes.Contains(parts[1], []byte("(")) {
		startIdx := bytes.Index(parts[1], []byte("("))
		endIdx := bytes.Index(parts[1], []byte(")"))
		if startIdx != -1 && endIdx != -1 {
			column.DataType = string(parts[1][:startIdx])
			sizeStr := string(parts[1][startIdx+1 : endIdx])
			sizes := strings.Split(sizeStr, ",")
			if len(sizes) > 0 {
				fmt.Sscanf(sizes[0], "%d", &column.Length)
				if len(sizes) > 1 {
					fmt.Sscanf(sizes[1], "%d", &column.Scale)
				}
			}
		}
	}

	upperDef := bytes.ToUpper(def)

	// Handle PRIMARY KEY
	if bytes.Contains(upperDef, []byte("PRIMARY KEY")) {
		column.IsPrimaryKey = true
		column.IsNullable = false
	}

	// Handle NOT NULL
	if bytes.Contains(upperDef, []byte("NOT NULL")) {
		column.IsNullable = false
	}

	// Handle UNIQUE
	if bytes.Contains(upperDef, []byte("UNIQUE")) {
		column.IsUnique = true
	}

	// Handle IDENTITY
	if bytes.Contains(upperDef, []byte("IDENTITY")) {
		column.AutoIncrement = true
	}

	// Handle DEFAULT
	if idx := bytes.Index(upperDef, []byte("DEFAULT")); idx != -1 {
		restDef := upperDef[idx+7:]
		endIdx := bytes.IndexAny(restDef, " ,")
		if endIdx == -1 {
			endIdx = len(restDef)
		}
		column.DefaultValue = string(bytes.Trim(restDef[:endIdx], "'()"))
	}

	return column
}

// parseConstraint parses a table constraint definition and returns a Constraint structure.
func (s *SQLServer) parseConstraint(def []byte) sqlmapper.Constraint {
	constraint := sqlmapper.Constraint{}
	def = bytes.TrimSpace(def)

	// Extract constraint name if exists
	if bytes.HasPrefix(bytes.ToUpper(def), []byte("CONSTRAINT")) {
		parts := bytes.Fields(def)
		if len(parts) > 1 {
			constraint.Name = string(bytes.Trim(parts[1], "[]"))
			def = bytes.Join(parts[2:], []byte(" "))
		}
	}

	upperDef := bytes.ToUpper(def)
	switch {
	case bytes.Contains(upperDef, []byte("PRIMARY KEY")):
		constraint.Type = "PRIMARY KEY"
		constraint.Columns = s.extractColumns(def, "PRIMARY KEY")

	case bytes.Contains(upperDef, []byte("FOREIGN KEY")):
		constraint.Type = "FOREIGN KEY"
		constraint.Columns = s.extractColumns(def, "FOREIGN KEY")

		// Extract referenced table and columns
		if idx := bytes.Index(upperDef, []byte("REFERENCES")); idx != -1 {
			refPart := def[idx:]
			startIdx := bytes.Index(refPart, []byte("("))
			endIdx := bytes.Index(refPart, []byte(")"))
			if startIdx != -1 && endIdx != -1 {
				tableName := bytes.TrimSpace(refPart[9:startIdx])
				// Remove schema prefix and brackets
				if idx := bytes.LastIndex(tableName, []byte(".")); idx != -1 {
					tableName = tableName[idx+1:]
				}
				tableName = bytes.Trim(tableName, "[]")
				constraint.RefTable = string(tableName)

				colStr := string(refPart[startIdx+1 : endIdx])
				constraint.RefColumns = s.splitAndTrim(colStr)
			}
		}

		// Extract ON DELETE rule
		if bytes.Contains(upperDef, []byte("ON DELETE CASCADE")) {
			constraint.DeleteRule = "CASCADE"
		}

	case bytes.Contains(upperDef, []byte("UNIQUE")):
		constraint.Type = "UNIQUE"
		constraint.Columns = s.extractColumns(def, "UNIQUE")

	case bytes.Contains(upperDef, []byte("CHECK")):
		constraint.Type = "CHECK"
		if idx := bytes.Index(upperDef, []byte("CHECK")); idx != -1 {
			startIdx := bytes.Index(def[idx:], []byte("("))
			endIdx := bytes.LastIndex(def, []byte(")"))
			if startIdx != -1 && endIdx != -1 {
				constraint.CheckExpression = string(bytes.TrimSpace(def[idx+startIdx+1 : endIdx]))
			}
		}
	}

	return constraint
}

// extractColumns extracts column names from a constraint definition.
func (s *SQLServer) extractColumns(def []byte, afterKeyword string) []string {
	upperDef := bytes.ToUpper(def)
	keyword := []byte(afterKeyword)

	if idx := bytes.Index(upperDef, keyword); idx != -1 {
		rest := def[idx+len(keyword):]
		startIdx := bytes.Index(rest, []byte("("))
		endIdx := bytes.Index(rest, []byte(")"))
		if startIdx != -1 && endIdx != -1 {
			colStr := string(rest[startIdx+1 : endIdx])
			return s.splitAndTrim(colStr)
		}
	}
	return nil
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

	s.buf.Reset()

	for _, table := range schema.Tables {
		s.buf.WriteString("CREATE TABLE ")
		s.buf.WriteString(table.Name)
		s.buf.WriteString(" (\n")

		for i, col := range table.Columns {
			s.buf.WriteString("    ")
			s.buf.WriteString(col.Name)
			s.buf.WriteByte(' ')
			s.buf.WriteString(col.DataType)

			if col.Length > 0 {
				if col.Scale > 0 {
					fmt.Fprintf(s.buf, "(%d,%d)", col.Length, col.Scale)
				} else {
					fmt.Fprintf(s.buf, "(%d)", col.Length)
				}
			}

			if col.IsPrimaryKey {
				s.buf.WriteString(" PRIMARY KEY")
			} else if !col.IsNullable {
				s.buf.WriteString(" NOT NULL")
			}

			if col.IsUnique && !col.IsPrimaryKey {
				s.buf.WriteString(" UNIQUE")
			}

			if col.AutoIncrement {
				s.buf.WriteString(" IDENTITY(1,1)")
			}

			if i < len(table.Columns)-1 {
				s.buf.WriteByte(',')
			}
			s.buf.WriteByte('\n')
		}

		s.buf.WriteString(");\n")

		// Add indexes
		for _, idx := range table.Indexes {
			if idx.IsUnique {
				s.buf.WriteString("CREATE UNIQUE INDEX ")
			} else {
				s.buf.WriteString("CREATE INDEX ")
			}
			s.buf.WriteString(idx.Name)
			s.buf.WriteString(" ON ")
			s.buf.WriteString(table.Name)
			s.buf.WriteByte('(')
			s.buf.WriteString(strings.Join(idx.Columns, ", "))
			s.buf.WriteString(");\n")
		}
	}

	return s.buf.String(), nil
}

// parseCreateIndex parses a CREATE INDEX statement and adds the index to the appropriate table.
func (s *SQLServer) parseCreateIndex(stmt []byte) error {
	isUnique := bytes.HasPrefix(bytes.ToUpper(stmt), []byte("CREATE UNIQUE"))

	// Extract index name and table name
	parts := bytes.Fields(stmt)
	if len(parts) < 4 {
		return fmt.Errorf("invalid CREATE INDEX statement")
	}

	var indexNamePos, tableNamePos int
	if isUnique {
		indexNamePos = 3
		tableNamePos = 5
	} else {
		indexNamePos = 2
		tableNamePos = 4
	}

	if len(parts) <= tableNamePos {
		return fmt.Errorf("invalid CREATE INDEX statement: missing table name")
	}

	indexName := string(bytes.Trim(parts[indexNamePos], "[]"))
	tableName := string(bytes.Trim(parts[tableNamePos], "[]"))

	// Remove schema prefix if exists
	if idx := bytes.LastIndex(parts[tableNamePos], []byte(".")); idx != -1 {
		tableName = string(bytes.Trim(parts[tableNamePos][idx+1:], "[]"))
	}

	// Extract columns
	startIdx := bytes.LastIndex(stmt, []byte("("))
	endIdx := bytes.LastIndex(stmt, []byte(")"))
	if startIdx == -1 || endIdx == -1 {
		return fmt.Errorf("no columns found in CREATE INDEX statement")
	}

	columns := s.splitAndTrim(string(bytes.TrimSpace(stmt[startIdx+1 : endIdx])))

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
func (s *SQLServer) parseAlterTable(stmt []byte) error {
	// Extract table name
	parts := bytes.Fields(stmt)
	if len(parts) < 3 {
		return fmt.Errorf("invalid ALTER TABLE statement")
	}

	tableName := string(bytes.Trim(parts[2], "[]"))
	// Remove schema prefix if exists
	if idx := bytes.LastIndex(parts[2], []byte(".")); idx != -1 {
		tableName = string(bytes.Trim(parts[2][idx+1:], "[]"))
	}

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
	upperStmt := bytes.ToUpper(stmt)
	switch {
	case bytes.Contains(upperStmt, []byte("ADD CONSTRAINT")):
		if idx := bytes.Index(upperStmt, []byte("ADD CONSTRAINT")); idx != -1 {
			constraint := s.parseConstraint(stmt[idx:])
			s.schema.Tables[tableIndex].Constraints = append(s.schema.Tables[tableIndex].Constraints, constraint)
		}

	case bytes.Contains(upperStmt, []byte("ADD COLUMN")) || bytes.Contains(upperStmt, []byte("ADD ")):
		// Extract column definition
		addIndex := bytes.Index(upperStmt, []byte("ADD "))
		if addIndex == -1 {
			return fmt.Errorf("invalid ALTER TABLE ADD statement")
		}

		colDef := bytes.TrimSpace(stmt[addIndex+4:])
		if bytes.HasPrefix(bytes.ToUpper(colDef), []byte("COLUMN")) {
			colDef = bytes.TrimSpace(colDef[6:])
		}

		column := s.parseColumn(colDef)
		s.schema.Tables[tableIndex].Columns = append(s.schema.Tables[tableIndex].Columns, column)
	}

	return nil
}

// parseCreateView parses a CREATE VIEW statement and returns a View structure.
func (s *SQLServer) parseCreateView(stmt []byte) (sqlmapper.View, error) {
	view := sqlmapper.View{}

	// Extract view name
	parts := bytes.Fields(stmt)
	if len(parts) < 3 {
		return view, fmt.Errorf("invalid CREATE VIEW statement")
	}

	viewName := string(bytes.Trim(parts[2], "[]"))
	// Remove schema prefix if exists
	if idx := bytes.LastIndex(parts[2], []byte(".")); idx != -1 {
		viewName = string(bytes.Trim(parts[2][idx+1:], "[]"))
	}
	view.Name = viewName

	// Extract view definition
	if idx := bytes.Index(bytes.ToUpper(stmt), []byte(" AS ")); idx != -1 {
		view.Definition = string(bytes.TrimSpace(stmt[idx+4:]))
	}

	return view, nil
}

// parseCreateTrigger parses a CREATE TRIGGER statement and returns a Trigger structure.
func (s *SQLServer) parseCreateTrigger(stmt []byte) (sqlmapper.Trigger, error) {
	trigger := sqlmapper.Trigger{}

	// Extract trigger name
	parts := bytes.Fields(stmt)
	if len(parts) < 3 {
		return trigger, fmt.Errorf("invalid CREATE TRIGGER statement")
	}

	triggerName := string(bytes.Trim(parts[2], "[]"))
	trigger.Name = triggerName

	upperStmt := bytes.ToUpper(stmt)

	// Extract timing (AFTER/FOR/INSTEAD OF)
	switch {
	case bytes.Contains(upperStmt, []byte("AFTER")):
		trigger.Timing = "AFTER"
	case bytes.Contains(upperStmt, []byte("INSTEAD OF")):
		trigger.Timing = "INSTEAD OF"
	default:
		trigger.Timing = "FOR"
	}

	// Extract event (INSERT/UPDATE/DELETE)
	switch {
	case bytes.Contains(upperStmt, []byte("INSERT")):
		trigger.Event = "INSERT"
	case bytes.Contains(upperStmt, []byte("UPDATE")):
		trigger.Event = "UPDATE"
	case bytes.Contains(upperStmt, []byte("DELETE")):
		trigger.Event = "DELETE"
	}

	// Extract table name
	if idx := bytes.Index(upperStmt, []byte(" ON ")); idx != -1 {
		rest := stmt[idx+4:]
		if spaceIdx := bytes.Index(rest, []byte(" ")); spaceIdx != -1 {
			tableName := bytes.TrimSpace(rest[:spaceIdx])
			// Remove schema prefix if exists
			if dotIdx := bytes.LastIndex(tableName, []byte(".")); dotIdx != -1 {
				tableName = tableName[dotIdx+1:]
			}
			trigger.Table = string(bytes.Trim(tableName, "[]"))
		}
	}

	// Extract trigger body
	if idx := bytes.Index(upperStmt, []byte(" AS ")); idx != -1 {
		trigger.Body = string(bytes.TrimSpace(stmt[idx+4:]))
	}

	return trigger, nil
}

func (s *SQLServer) parseTables(statement string) error {
	re := regexp.MustCompile(`CREATE\s+TABLE\s+([.\w\[\]]+)\s*\((.*?)\)(?:\s+ON\s+(\w+))?`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) > 2 {
		tableName := matches[1]
		columnDefs := matches[2]

		table := sqlmapper.Table{}

		// Parse schema if exists
		parts := strings.Split(strings.Trim(tableName, "[]"), ".")
		if len(parts) > 1 {
			table.Schema = parts[0]
			table.Name = parts[1]
		} else {
			table.Name = tableName
		}

		// Parse filegroup if exists
		if len(matches) > 3 && matches[3] != "" {
			table.TableSpace = matches[3]
		}

		// Parse columns and constraints
		columns := strings.Split(columnDefs, ",")
		for _, col := range columns {
			col = strings.TrimSpace(col)
			if strings.HasPrefix(strings.ToUpper(col), "CONSTRAINT") {
				continue // Skip constraints for now
			}

			parts := strings.Fields(col)
			if len(parts) < 2 {
				continue
			}

			column := sqlmapper.Column{
				Name:       strings.Trim(parts[0], "[]"),
				DataType:   parts[1],
				IsNullable: true,
			}

			// Parse length/precision
			if strings.Contains(column.DataType, "(") {
				re := regexp.MustCompile(`(\w+)\((\d+|MAX)(?:,(\d+))?\)`)
				if matches := re.FindStringSubmatch(column.DataType); len(matches) > 2 {
					column.DataType = matches[1]
					if matches[2] == "MAX" {
						column.Length = -1
					} else {
						fmt.Sscanf(matches[2], "%d", &column.Length)
					}
					if len(matches) > 3 && matches[3] != "" {
						fmt.Sscanf(matches[3], "%d", &column.Scale)
					}
				}
			}

			// Parse constraints
			if strings.Contains(strings.ToUpper(col), "NOT NULL") {
				column.IsNullable = false
			}
			if strings.Contains(strings.ToUpper(col), "PRIMARY KEY") {
				column.IsPrimaryKey = true
			}
			if strings.Contains(strings.ToUpper(col), "IDENTITY") {
				column.AutoIncrement = true
			}
			if strings.Contains(strings.ToUpper(col), "UNIQUE") {
				column.IsUnique = true
			}
			if strings.Contains(strings.ToUpper(col), "DEFAULT") {
				re := regexp.MustCompile(`DEFAULT\s+([^,\s]+)`)
				if matches := re.FindStringSubmatch(col); len(matches) > 1 {
					column.DefaultValue = matches[1]
				}
			}

			table.Columns = append(table.Columns, column)
		}

		s.schema.Tables = append(s.schema.Tables, table)
	}

	return nil
}

func (s *SQLServer) parseViews(statement string) error {
	re := regexp.MustCompile(`CREATE\s+VIEW\s+([.\w\[\]]+)\s+AS\s+(.+)$`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) > 2 {
		viewName := matches[1]
		view := sqlmapper.View{
			Definition: matches[2],
		}

		// Parse schema if exists
		parts := strings.Split(strings.Trim(viewName, "[]"), ".")
		if len(parts) > 1 {
			view.Schema = parts[0]
			view.Name = parts[1]
		} else {
			view.Name = viewName
		}

		s.schema.Views = append(s.schema.Views, view)
	}

	return nil
}

func (s *SQLServer) parseFunctions(statement string) error {
	re := regexp.MustCompile(`CREATE\s+(FUNCTION|PROCEDURE)\s+([.\w\[\]]+)\s*\((.*?)\)(?:\s+RETURNS\s+(\w+(?:\s*\([^)]*\))?))?\s+AS\s+BEGIN\s+(.*?)\s+END`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) > 4 {
		isProc := matches[1] == "PROCEDURE"
		functionName := matches[2]
		function := sqlmapper.Function{
			IsProc: isProc,
			Body:   matches[5],
		}

		if !isProc && matches[4] != "" {
			function.Returns = matches[4]
		}

		// Parse schema if exists
		parts := strings.Split(strings.Trim(functionName, "[]"), ".")
		if len(parts) > 1 {
			function.Schema = parts[0]
			function.Name = parts[1]
		} else {
			function.Name = functionName
		}

		// Parse parameters
		if matches[3] != "" {
			params := strings.Split(matches[3], ",")
			for _, param := range params {
				parts := strings.Fields(strings.TrimSpace(param))
				if len(parts) >= 2 {
					parameter := sqlmapper.Parameter{
						Name:     strings.TrimPrefix(parts[0], "@"),
						DataType: parts[1],
					}
					function.Parameters = append(function.Parameters, parameter)
				}
			}
		}

		s.schema.Functions = append(s.schema.Functions, function)
	}

	return nil
}

func (s *SQLServer) parseTriggers(statement string) error {
	re := regexp.MustCompile(`CREATE\s+TRIGGER\s+([.\w\[\]]+)\s+ON\s+([.\w\[\]]+)\s+(AFTER|INSTEAD\s+OF|FOR)\s+(INSERT|UPDATE|DELETE)(?:\s*,\s*(INSERT|UPDATE|DELETE))*\s+AS\s+BEGIN\s+(.*?)\s+END`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) > 6 {
		triggerName := matches[1]
		trigger := sqlmapper.Trigger{
			Table:  matches[2],
			Timing: matches[3],
			Event:  matches[4],
			Body:   matches[6],
		}

		// Parse schema if exists
		parts := strings.Split(strings.Trim(triggerName, "[]"), ".")
		if len(parts) > 1 {
			trigger.Schema = parts[0]
			trigger.Name = parts[1]
		} else {
			trigger.Name = triggerName
		}

		s.schema.Triggers = append(s.schema.Triggers, trigger)
	}

	return nil
}

func (s *SQLServer) parseIndexes(statement string) error {
	re := regexp.MustCompile(`CREATE\s+(?:(UNIQUE|CLUSTERED|NONCLUSTERED)\s+)*INDEX\s+([.\w\[\]]+)\s+ON\s+([.\w\[\]]+)\s*\((.*?)\)(?:\s+INCLUDE\s*\((.*?)\))?(?:\s+WITH\s*\((.*?)\))?(?:\s+ON\s+(\w+))?`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) > 4 {
		indexName := matches[2]
		tableName := matches[3]
		columns := strings.Split(matches[4], ",")

		// Find the table
		for i, table := range s.schema.Tables {
			if table.Name == tableName || fmt.Sprintf("%s.%s", table.Schema, table.Name) == tableName {
				index := sqlmapper.Index{
					Name:        strings.Trim(indexName, "[]"),
					Columns:     make([]string, len(columns)),
					IsUnique:    strings.Contains(matches[1], "UNIQUE"),
					IsClustered: strings.Contains(matches[1], "CLUSTERED"),
				}

				// Clean column names
				for j, col := range columns {
					index.Columns[j] = strings.Trim(strings.TrimSpace(col), "[]")
				}

				// Skip INCLUDE columns since they're not supported in the common schema

				// Parse filegroup
				if len(matches) > 7 && matches[7] != "" {
					index.TableSpace = matches[7]
				}

				s.schema.Tables[i].Indexes = append(s.schema.Tables[i].Indexes, index)
				break
			}
		}
	}

	return nil
}

// generateTableSQL generates SQL for a table
func (s *SQLServer) generateTableSQL(table sqlmapper.Table) string {
	sql := "CREATE TABLE " + table.Name + " (\n"

	// Generate columns
	for i, col := range table.Columns {
		sql += "    " + col.Name + " " + col.DataType
		if col.Length > 0 {
			if strings.ToUpper(col.DataType) == "NVARCHAR" || strings.ToUpper(col.DataType) == "NCHAR" {
				if col.Length == -1 {
					sql += "(MAX)"
				} else {
					sql += fmt.Sprintf("(%d)", col.Length)
				}
			} else {
				sql += fmt.Sprintf("(%d", col.Length)
				if col.Scale > 0 {
					sql += fmt.Sprintf(",%d", col.Scale)
				}
				sql += ")"
			}
		}

		if col.IsPrimaryKey {
			sql += " PRIMARY KEY"
			if col.AutoIncrement {
				sql += " IDENTITY(1,1)"
			}
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

		if i < len(table.Columns)-1 {
			sql += ",\n"
		}
	}

	sql += "\n)"

	return sql
}

// generateIndexSQL generates SQL for an index
func (s *SQLServer) generateIndexSQL(tableName string, index sqlmapper.Index) string {
	var sql string
	if index.IsClustered {
		sql = "CREATE CLUSTERED "
	} else {
		sql = "CREATE NONCLUSTERED "
	}

	if index.IsUnique {
		sql += "UNIQUE INDEX "
	} else {
		sql += "INDEX "
	}

	sql += index.Name + " ON " + tableName + " (" + strings.Join(index.Columns, ", ") + ")"

	return sql
}
