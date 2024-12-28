// Package oracle provides functionality for parsing and generating Oracle database schemas.
// It implements the Parser interface for handling Oracle specific SQL syntax and schema structures.
package oracle

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/mstgnz/sqlmapper"
)

// Oracle represents an Oracle parser implementation that handles parsing and generating
// Oracle database schemas. It maintains an internal schema representation and provides
// methods for converting between Oracle SQL and the common schema format.
type Oracle struct {
	schema *sqlmapper.Schema
}

// NewOracle creates and initializes a new Oracle parser instance.
// It returns a parser that can handle Oracle specific SQL syntax and schema structures.
func NewOracle() *Oracle {
	return &Oracle{
		schema: &sqlmapper.Schema{},
	}
}

// Parse takes an Oracle SQL dump content and parses it into a common schema structure.
// It processes various Oracle objects including:
// - Tables with columns and constraints
// - Sequences
// - Views
// - Triggers
// - User privileges
//
// Parameters:
//   - content: The Oracle SQL dump content to parse
//
// Returns:
//   - *sqlmapper.Schema: The parsed schema structure
//   - error: An error if parsing fails or if the content is empty
func (o *Oracle) Parse(content string) (*sqlmapper.Schema, error) {
	if content == "" {
		return nil, errors.New("empty content")
	}

	// SQL ifadelerini ayır
	var statements []string
	var currentStmt strings.Builder
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if line == "/" {
			if currentStmt.Len() > 0 {
				statements = append(statements, currentStmt.String())
				currentStmt.Reset()
			}
			continue
		}

		if strings.HasSuffix(line, ";") {
			currentStmt.WriteString(line[:len(line)-1])
			if currentStmt.Len() > 0 {
				statements = append(statements, currentStmt.String())
				currentStmt.Reset()
			}
		} else {
			if currentStmt.Len() > 0 {
				currentStmt.WriteString(" ")
			}
			currentStmt.WriteString(line)
		}
	}

	if currentStmt.Len() > 0 {
		statements = append(statements, currentStmt.String())
	}

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// CREATE TABLE
		if strings.HasPrefix(strings.ToUpper(stmt), "CREATE TABLE") {
			table, err := o.parseCreateTable(stmt)
			if err != nil {
				return nil, err
			}
			o.schema.Tables = append(o.schema.Tables, table)
		}

		// CREATE SEQUENCE
		if strings.HasPrefix(strings.ToUpper(stmt), "CREATE SEQUENCE") {
			seq, err := o.parseCreateSequence(stmt)
			if err != nil {
				return nil, err
			}
			o.schema.Sequences = append(o.schema.Sequences, seq)
		}

		// CREATE VIEW
		if strings.HasPrefix(strings.ToUpper(stmt), "CREATE") && strings.Contains(strings.ToUpper(stmt), "VIEW") {
			view, err := o.parseCreateView(stmt)
			if err != nil {
				return nil, err
			}
			o.schema.Views = append(o.schema.Views, view)
		}

		// CREATE TRIGGER
		if strings.HasPrefix(strings.ToUpper(stmt), "CREATE") && strings.Contains(strings.ToUpper(stmt), "TRIGGER") {
			trigger, err := o.parseCreateTrigger(stmt)
			if err != nil {
				return nil, err
			}
			o.schema.Triggers = append(o.schema.Triggers, trigger)
		}
	}

	return o.schema, nil
}

// parseCreateTable processes a CREATE TABLE statement and extracts table structure.
// It handles various table components including:
// - Table name and schema
// - Column definitions with data types and constraints
// - Table-level constraints (PRIMARY KEY, FOREIGN KEY, UNIQUE, CHECK)
// - Column-level constraints and properties
//
// Parameters:
//   - stmt: The CREATE TABLE statement to parse
//
// Returns:
//   - sqlmapper.Table: The parsed table structure
//   - error: An error if parsing fails
func (o *Oracle) parseCreateTable(stmt string) (sqlmapper.Table, error) {
	table := sqlmapper.Table{}

	// Tablo adını al
	tableNameRegex := regexp.MustCompile(`CREATE\s+TABLE\s+(\w+)`)
	matches := tableNameRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		table.Name = matches[1]
	}

	// Kolonları parse et
	columnsStr := stmt[strings.Index(stmt, "(")+1 : strings.LastIndex(stmt, ")")]
	columnDefs := strings.Split(columnsStr, ",")

	for _, colDef := range columnDefs {
		colDef = strings.TrimSpace(colDef)
		if strings.HasPrefix(colDef, "CONSTRAINT") {
			constraint := sqlmapper.Constraint{}

			// Constraint adını al
			nameRegex := regexp.MustCompile(`CONSTRAINT\s+(\w+)`)
			matches := nameRegex.FindStringSubmatch(colDef)
			if len(matches) > 1 {
				constraint.Name = matches[1]
			}

			if strings.Contains(colDef, "PRIMARY KEY") {
				constraint.Type = "PRIMARY KEY"
				// Kolonları al
				colsRegex := regexp.MustCompile(`PRIMARY\s+KEY\s*\(([^)]+)\)`)
				matches = colsRegex.FindStringSubmatch(colDef)
				if len(matches) > 1 {
					cols := strings.Split(matches[1], ",")
					for i, col := range cols {
						cols[i] = strings.TrimSpace(col)
					}
					constraint.Columns = cols
				}
			} else if strings.Contains(colDef, "FOREIGN KEY") {
				constraint.Type = "FOREIGN KEY"
				// FK kolonlarını al
				fkRegex := regexp.MustCompile(`FOREIGN\s+KEY\s*\(([^)]+)\)`)
				matches = fkRegex.FindStringSubmatch(colDef)
				if len(matches) > 1 {
					cols := strings.Split(matches[1], ",")
					for i, col := range cols {
						cols[i] = strings.TrimSpace(col)
					}
					constraint.Columns = cols
				}
				// Referans tabloyu ve kolonları al
				refRegex := regexp.MustCompile(`REFERENCES\s+(\w+)\s*\(([^)]+)\)`)
				matches = refRegex.FindStringSubmatch(colDef)
				if len(matches) > 2 {
					constraint.RefTable = matches[1]
					refCols := strings.Split(matches[2], ",")
					for i, col := range refCols {
						refCols[i] = strings.TrimSpace(col)
					}
					constraint.RefColumns = refCols
				}
				// ON DELETE kuralını al
				if strings.Contains(colDef, "ON DELETE") {
					if strings.Contains(colDef, "CASCADE") {
						constraint.DeleteRule = "CASCADE"
					}
				}
			} else if strings.Contains(colDef, "UNIQUE") {
				constraint.Type = "UNIQUE"
				// Kolonları al
				colsRegex := regexp.MustCompile(`UNIQUE\s*\(([^)]+)\)`)
				matches = colsRegex.FindStringSubmatch(colDef)
				if len(matches) > 1 {
					cols := strings.Split(matches[1], ",")
					for i, col := range cols {
						cols[i] = strings.TrimSpace(col)
					}
					constraint.Columns = cols
				}
			} else if strings.Contains(colDef, "CHECK") {
				constraint.Type = "CHECK"
				// Check ifadesini al
				checkRegex := regexp.MustCompile(`CHECK\s*\(([^)]+)\)`)
				matches = checkRegex.FindStringSubmatch(colDef)
				if len(matches) > 1 {
					constraint.CheckExpression = strings.TrimSpace(matches[1])
				}
			}
			table.Constraints = append(table.Constraints, constraint)
			continue
		}

		parts := strings.Fields(colDef)
		if len(parts) < 2 {
			continue
		}

		col := sqlmapper.Column{
			Name:     parts[0],
			DataType: parts[1],
		}

		if strings.Contains(colDef, "NOT NULL") {
			col.IsNullable = false
		}

		if strings.Contains(colDef, "DEFAULT") {
			defaultIdx := strings.Index(strings.ToUpper(colDef), "DEFAULT")
			restStr := colDef[defaultIdx+7:]
			defaultEnd := strings.Index(restStr, " ")
			if defaultEnd == -1 {
				defaultEnd = len(restStr)
			}
			col.DefaultValue = strings.TrimSpace(restStr[:defaultEnd])
		}

		if strings.Contains(colDef, "PRIMARY KEY") {
			col.IsPrimaryKey = true
			constraint := sqlmapper.Constraint{
				Type:    "PRIMARY KEY",
				Columns: []string{col.Name},
			}
			table.Constraints = append(table.Constraints, constraint)
		}

		if strings.Contains(colDef, "UNIQUE") {
			col.IsUnique = true
			constraint := sqlmapper.Constraint{
				Type:    "UNIQUE",
				Columns: []string{col.Name},
			}
			table.Constraints = append(table.Constraints, constraint)
		}

		if strings.Contains(colDef, "CHECK") {
			checkRegex := regexp.MustCompile(`CHECK\s*\(([^)]+)\)`)
			matches := checkRegex.FindStringSubmatch(colDef)
			if len(matches) > 1 {
				constraint := sqlmapper.Constraint{
					Type:            "CHECK",
					CheckExpression: strings.TrimSpace(matches[1]),
				}
				table.Constraints = append(table.Constraints, constraint)
			}
		}

		table.Columns = append(table.Columns, col)
	}

	return table, nil
}

// parseCreateSequence processes a CREATE SEQUENCE statement.
// It extracts sequence properties including:
// - Sequence name and schema
// - Start value and increment
// - Min and max values
// - Cycle option
//
// Parameters:
//   - stmt: The CREATE SEQUENCE statement to parse
//
// Returns:
//   - sqlmapper.Sequence: The parsed sequence structure
//   - error: An error if parsing fails
func (o *Oracle) parseCreateSequence(stmt string) (sqlmapper.Sequence, error) {
	seq := sqlmapper.Sequence{}

	// Sequence adını al
	seqNameRegex := regexp.MustCompile(`CREATE\s+SEQUENCE\s+(\w+)`)
	matches := seqNameRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		seq.Name = matches[1]
	}

	// START WITH değerini al
	startWithRegex := regexp.MustCompile(`START\s+WITH\s+(\d+)`)
	matches = startWithRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		seq.StartValue = 1 // Default değer
	}

	// INCREMENT BY değerini al
	incrementByRegex := regexp.MustCompile(`INCREMENT\s+BY\s+(\d+)`)
	matches = incrementByRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		seq.IncrementBy = 1 // Default değer
	}

	return seq, nil
}

// parseCreateView processes a CREATE VIEW statement.
// It extracts view properties including:
// - View name and schema
// - View definition (SELECT statement)
// - View options (FORCE, WITH CHECK OPTION)
//
// Parameters:
//   - stmt: The CREATE VIEW statement to parse
//
// Returns:
//   - sqlmapper.View: The parsed view structure
//   - error: An error if parsing fails
func (o *Oracle) parseCreateView(stmt string) (sqlmapper.View, error) {
	view := sqlmapper.View{}

	// View adını al
	viewNameRegex := regexp.MustCompile(`CREATE\s+(?:OR\s+REPLACE\s+)?VIEW\s+(\w+)`)
	matches := viewNameRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		view.Name = matches[1]
	}

	// View tanımını al
	asIndex := strings.Index(strings.ToUpper(stmt), " AS ")
	if asIndex != -1 {
		view.Definition = strings.TrimSpace(stmt[asIndex+4:])
	}

	return view, nil
}

// parseCreateTrigger processes a CREATE TRIGGER statement.
// It extracts trigger properties including:
// - Trigger name and schema
// - Triggering event (INSERT, UPDATE, DELETE)
// - Trigger timing (BEFORE, AFTER, INSTEAD OF)
// - Table name
// - Trigger body
//
// Parameters:
//   - stmt: The CREATE TRIGGER statement to parse
//
// Returns:
//   - sqlmapper.Trigger: The parsed trigger structure
//   - error: An error if parsing fails
func (o *Oracle) parseCreateTrigger(stmt string) (sqlmapper.Trigger, error) {
	trigger := sqlmapper.Trigger{}

	// Trigger adını al
	triggerNameRegex := regexp.MustCompile(`CREATE\s+(?:OR\s+REPLACE\s+)?TRIGGER\s+(\w+)`)
	matches := triggerNameRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		trigger.Name = matches[1]
	}

	// Trigger zamanlamasını al
	if strings.Contains(strings.ToUpper(stmt), "BEFORE") {
		trigger.Timing = "BEFORE"
	} else if strings.Contains(strings.ToUpper(stmt), "AFTER") {
		trigger.Timing = "AFTER"
	}

	// Trigger olayını al
	if strings.Contains(strings.ToUpper(stmt), "INSERT") {
		trigger.Event = "INSERT"
	} else if strings.Contains(strings.ToUpper(stmt), "UPDATE") {
		trigger.Event = "UPDATE"
	} else if strings.Contains(strings.ToUpper(stmt), "DELETE") {
		trigger.Event = "DELETE"
	}

	// Tablo adını al
	tableRegex := regexp.MustCompile(`ON\s+(\w+)`)
	matches = tableRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		trigger.Table = matches[1]
	}

	// FOR EACH ROW kontrolü
	trigger.ForEachRow = strings.Contains(strings.ToUpper(stmt), "FOR EACH ROW")

	// Trigger gövdesini al
	beginIndex := strings.Index(strings.ToUpper(stmt), "BEGIN")
	endIndex := strings.LastIndex(strings.ToUpper(stmt), "END")
	if beginIndex != -1 && endIndex != -1 {
		trigger.Body = strings.TrimSpace(stmt[beginIndex : endIndex+3])
	}

	return trigger, nil
}

// Generate creates an Oracle SQL dump from a schema structure.
// It generates SQL statements for all database objects in the schema, including:
// - Tables with columns, constraints, and indexes
// - Sequences
// - Views
// - Triggers
// - User privileges
//
// Parameters:
//   - schema: The schema structure to convert to Oracle SQL
//
// Returns:
//   - string: The generated Oracle SQL statements
//   - error: An error if generation fails or if the schema is nil
func (o *Oracle) Generate(schema *sqlmapper.Schema) (string, error) {
	if schema == nil {
		return "", errors.New("empty schema")
	}

	var result strings.Builder

	// Create sequences
	for _, seq := range schema.Sequences {
		result.WriteString(fmt.Sprintf("CREATE SEQUENCE %s START WITH %d INCREMENT BY %d;\n\n",
			seq.Name, seq.StartValue, seq.IncrementBy))
	}

	// Create tables
	for _, table := range schema.Tables {
		result.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", table.Name))

		// Add columns
		for i, col := range table.Columns {
			result.WriteString(fmt.Sprintf("    %s %s", col.Name, col.DataType))
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
			if col.DefaultValue != "" {
				// Add quotes for default values of type String
				if strings.HasPrefix(col.DataType, "VARCHAR") || strings.HasPrefix(col.DataType, "CHAR") {
					result.WriteString(fmt.Sprintf(" DEFAULT '%s'", col.DefaultValue))
				} else {
					result.WriteString(fmt.Sprintf(" DEFAULT %s", col.DefaultValue))
				}
			}
			if col.IsUnique && !col.IsPrimaryKey {
				result.WriteString(" UNIQUE")
			}
			if i < len(table.Columns)-1 || len(table.Constraints) > 0 {
				result.WriteString(",")
			}
			result.WriteString("\n")
		}

		// Add Constraint
		for i, constraint := range table.Constraints {
			if constraint.Name == "" {
				continue // Skip unnamed constraints as they are handled with column definitions
			}
			result.WriteString(fmt.Sprintf("    CONSTRAINT %s %s", constraint.Name, constraint.Type))
			if len(constraint.Columns) > 0 {
				result.WriteString(fmt.Sprintf(" (%s)", strings.Join(constraint.Columns, ", ")))
			}
			if constraint.Type == "FOREIGN KEY" && constraint.RefTable != "" {
				result.WriteString(fmt.Sprintf(" REFERENCES %s", constraint.RefTable))
				if len(constraint.RefColumns) > 0 {
					result.WriteString(fmt.Sprintf("(%s)", strings.Join(constraint.RefColumns, ", ")))
				}
				if constraint.DeleteRule != "" {
					result.WriteString(fmt.Sprintf(" ON DELETE %s", constraint.DeleteRule))
				}
			}
			if i < len(table.Constraints)-1 {
				result.WriteString(",")
			}
			result.WriteString("\n")
		}

		result.WriteString(");\n")

		// Index'leri oluştur
		for _, index := range table.Indexes {
			if index.IsUnique {
				result.WriteString(fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s(%s);\n",
					index.Name, table.Name, strings.Join(index.Columns, ", ")))
			} else {
				result.WriteString(fmt.Sprintf("CREATE INDEX %s ON %s(%s);\n",
					index.Name, table.Name, strings.Join(index.Columns, ", ")))
			}
		}

		result.WriteString("\n")
	}

	// Create views
	for _, view := range schema.Views {
		result.WriteString(fmt.Sprintf("CREATE OR REPLACE VIEW %s AS\n%s;\n\n",
			view.Name, view.Definition))
	}

	// Create triggers
	for _, trigger := range schema.Triggers {
		result.WriteString(fmt.Sprintf("CREATE OR REPLACE TRIGGER %s\n", trigger.Name))
		if trigger.Timing != "" {
			result.WriteString(trigger.Timing + " ")
		}
		if trigger.Event != "" {
			result.WriteString(trigger.Event + " ")
		}
		if trigger.Table != "" {
			result.WriteString("ON " + trigger.Table + " ")
		}
		if trigger.ForEachRow {
			result.WriteString("FOR EACH ROW\n")
		}
		if trigger.Body != "" {
			result.WriteString(trigger.Body)
		}
		result.WriteString("\n/\n\n")
	}

	return result.String(), nil
}

func (o *Oracle) parseTables(statement string) error {
	re := regexp.MustCompile(`CREATE\s+TABLE\s+([.\w]+)\s*\((.*?)\)(?:\s+TABLESPACE\s+(\w+))?`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) > 2 {
		tableName := matches[1]
		columnDefs := matches[2]

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
				Name:       parts[0],
				DataType:   parts[1],
				IsNullable: true,
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

			// Parse constraints
			if strings.Contains(strings.ToUpper(col), "NOT NULL") {
				column.IsNullable = false
			}
			if strings.Contains(strings.ToUpper(col), "PRIMARY KEY") {
				column.IsPrimaryKey = true
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

		o.schema.Tables = append(o.schema.Tables, table)
	}

	return nil
}

func (o *Oracle) parseViews(statement string) error {
	re := regexp.MustCompile(`CREATE(?:\s+OR\s+REPLACE)?\s+VIEW\s+([.\w]+)\s+AS\s+(.*?)(?:WITH\s+READ\s+ONLY)?$`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) > 2 {
		viewName := matches[1]
		view := sqlmapper.View{
			Definition: matches[2],
		}

		// Parse schema if exists
		parts := strings.Split(viewName, ".")
		if len(parts) > 1 {
			view.Schema = parts[0]
			view.Name = parts[1]
		} else {
			view.Name = viewName
		}

		o.schema.Views = append(o.schema.Views, view)
	}

	return nil
}

func (o *Oracle) parseFunctions(statement string) error {
	re := regexp.MustCompile(`CREATE(?:\s+OR\s+REPLACE)?\s+(FUNCTION|PROCEDURE)\s+([.\w]+)\s*\((.*?)\)(?:\s+RETURN\s+(\w+))?\s+IS|AS\s+(.*?)(?:END\s+\w+)?$`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) > 4 {
		isProc := matches[1] == "PROCEDURE"
		functionName := matches[2]
		function := sqlmapper.Function{
			IsProc: isProc,
			Body:   matches[5],
		}

		if !isProc {
			function.Returns = matches[4]
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
		if matches[3] != "" {
			params := strings.Split(matches[3], ",")
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

		o.schema.Functions = append(o.schema.Functions, function)
	}

	return nil
}

func (o *Oracle) parseTriggers(statement string) error {
	re := regexp.MustCompile(`CREATE(?:\s+OR\s+REPLACE)?\s+TRIGGER\s+([.\w]+)\s+(BEFORE|AFTER|INSTEAD\s+OF)\s+(INSERT|UPDATE|DELETE)\s+ON\s+([.\w]+)(?:\s+FOR\s+EACH\s+ROW)?\s+(.*?)(?:END\s+\w+)?$`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) > 5 {
		triggerName := matches[1]
		trigger := sqlmapper.Trigger{
			Timing:     matches[2],
			Event:      matches[3],
			Table:      matches[4],
			Body:       matches[5],
			ForEachRow: strings.Contains(statement, "FOR EACH ROW"),
		}

		// Parse schema if exists
		parts := strings.Split(triggerName, ".")
		if len(parts) > 1 {
			trigger.Schema = parts[0]
			trigger.Name = parts[1]
		} else {
			trigger.Name = triggerName
		}

		o.schema.Triggers = append(o.schema.Triggers, trigger)
	}

	return nil
}

func (o *Oracle) parseSequences(statement string) error {
	re := regexp.MustCompile(`CREATE\s+SEQUENCE\s+([.\w]+)(?:\s+START\s+WITH\s+(\d+))?(?:\s+INCREMENT\s+BY\s+(\d+))?(?:\s+MINVALUE\s+(\d+))?(?:\s+MAXVALUE\s+(\d+))?(?:\s+CACHE\s+(\d+))?(?:\s+CYCLE)?`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) > 1 {
		seqName := matches[1]
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
		if len(matches) > 2 && matches[2] != "" {
			fmt.Sscanf(matches[2], "%d", &seq.StartValue)
		}
		if len(matches) > 3 && matches[3] != "" {
			fmt.Sscanf(matches[3], "%d", &seq.IncrementBy)
		}
		if len(matches) > 4 && matches[4] != "" {
			fmt.Sscanf(matches[4], "%d", &seq.MinValue)
		}
		if len(matches) > 5 && matches[5] != "" {
			fmt.Sscanf(matches[5], "%d", &seq.MaxValue)
		}
		if len(matches) > 6 && matches[6] != "" {
			fmt.Sscanf(matches[6], "%d", &seq.Cache)
		}
		seq.Cycle = strings.Contains(statement, "CYCLE")

		o.schema.Sequences = append(o.schema.Sequences, seq)
	}

	return nil
}

func (o *Oracle) parseTypes(statement string) error {
	re := regexp.MustCompile(`CREATE(?:\s+OR\s+REPLACE)?\s+TYPE\s+([.\w]+)\s+(?:AS\s+|IS\s+|UNDER\s+)?(.+?)(?:NOT\s+FINAL)?$`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) > 2 {
		typeName := matches[1]
		typ := sqlmapper.Type{
			Definition: matches[2],
		}

		// Parse schema if exists
		parts := strings.Split(typeName, ".")
		if len(parts) > 1 {
			typ.Schema = parts[0]
			typ.Name = parts[1]
		} else {
			typ.Name = typeName
		}

		o.schema.Types = append(o.schema.Types, typ)
	}

	return nil
}

func (o *Oracle) parseIndexes(statement string) error {
	re := regexp.MustCompile(`CREATE(?:\s+UNIQUE|\s+BITMAP)?\s+INDEX\s+([.\w]+)\s+ON\s+([.\w]+)\s*\((.*?)\)(?:\s+TABLESPACE\s+(\w+))?`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) > 3 {
		indexName := matches[1]
		tableName := matches[2]
		columns := strings.Split(matches[3], ",")

		// Find the table
		for i, table := range o.schema.Tables {
			if table.Name == tableName || fmt.Sprintf("%s.%s", table.Schema, table.Name) == tableName {
				index := sqlmapper.Index{
					Name:     indexName,
					Columns:  make([]string, len(columns)),
					IsUnique: strings.Contains(statement, "UNIQUE"),
					IsBitmap: strings.Contains(statement, "BITMAP"),
				}

				// Clean column names
				for j, col := range columns {
					index.Columns[j] = strings.TrimSpace(col)
				}

				// Parse tablespace if exists
				if len(matches) > 4 && matches[4] != "" {
					index.TableSpace = matches[4]
				}

				o.schema.Tables[i].Indexes = append(o.schema.Tables[i].Indexes, index)
				break
			}
		}
	}

	return nil
}

// generateSequenceSQL generates SQL for a sequence
func (o *Oracle) generateSequenceSQL(sequence sqlmapper.Sequence) string {
	sql := fmt.Sprintf("CREATE SEQUENCE %s\n", sequence.Name)
	if sequence.StartValue > 0 {
		sql += fmt.Sprintf("START WITH %d\n", sequence.StartValue)
	}
	if sequence.IncrementBy > 0 {
		sql += fmt.Sprintf("INCREMENT BY %d\n", sequence.IncrementBy)
	}
	if sequence.MinValue > 0 {
		sql += fmt.Sprintf("MINVALUE %d\n", sequence.MinValue)
	}
	if sequence.MaxValue > 0 {
		sql += fmt.Sprintf("MAXVALUE %d\n", sequence.MaxValue)
	}
	if sequence.Cache > 0 {
		sql += fmt.Sprintf("CACHE %d\n", sequence.Cache)
	}
	if sequence.Cycle {
		sql += "CYCLE\n"
	}
	return sql
}

// generateTypeSQL generates SQL for a type
func (o *Oracle) generateTypeSQL(typ sqlmapper.Type) string {
	return fmt.Sprintf("CREATE TYPE %s AS %s", typ.Name, typ.Definition)
}

// generateTableSQL generates SQL for a table
func (o *Oracle) generateTableSQL(table sqlmapper.Table) string {
	sql := "CREATE TABLE " + table.Name + " (\n"

	// Generate columns
	for i, col := range table.Columns {
		sql += "    " + col.Name + " " + col.DataType
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
func (o *Oracle) generateIndexSQL(tableName string, index sqlmapper.Index) string {
	var sql string
	if index.IsBitmap {
		sql = "CREATE BITMAP INDEX "
	} else if index.IsUnique {
		sql = "CREATE UNIQUE INDEX "
	} else {
		sql = "CREATE INDEX "
	}

	sql += index.Name + " ON " + tableName + " (" + strings.Join(index.Columns, ", ") + ")"

	// Add index options
	if index.TableSpace != "" {
		sql += " TABLESPACE " + index.TableSpace
	}

	return sql
}
