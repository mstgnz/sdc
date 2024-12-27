package oracle

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/mstgnz/sqlmapper"
)

// OracleStreamParser implements the StreamParser interface for Oracle
type OracleStreamParser struct {
	oracle *Oracle
}

// NewOracleStreamParser creates a new Oracle stream parser
func NewOracleStreamParser() *OracleStreamParser {
	return &OracleStreamParser{
		oracle: NewOracle(),
	}
}

// ParseStream implements the StreamParser interface
func (p *OracleStreamParser) ParseStream(reader io.Reader, callback func(sqlmapper.SchemaObject) error) error {
	streamReader := sqlmapper.NewStreamReader(reader, "/")

	for {
		statement, err := streamReader.ReadStatement()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading statement: %v", err)
		}

		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		// Parse CREATE TABLE statements
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE TABLE") {
			table, err := p.parseTableStatement(statement)
			if err != nil {
				return err
			}

			err = callback(sqlmapper.SchemaObject{
				Type: sqlmapper.TableObject,
				Data: table,
			})
			if err != nil {
				return err
			}
			continue
		}

		// Parse CREATE VIEW statements
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE VIEW") ||
			strings.HasPrefix(strings.ToUpper(statement), "CREATE MATERIALIZED VIEW") {
			view, err := p.parseViewStatement(statement)
			if err != nil {
				return err
			}

			err = callback(sqlmapper.SchemaObject{
				Type: sqlmapper.ViewObject,
				Data: view,
			})
			if err != nil {
				return err
			}
			continue
		}

		// Parse CREATE FUNCTION statements
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE FUNCTION") {
			function, err := p.parseFunctionStatement(statement)
			if err != nil {
				return err
			}

			err = callback(sqlmapper.SchemaObject{
				Type: sqlmapper.FunctionObject,
				Data: function,
			})
			if err != nil {
				return err
			}
			continue
		}

		// Parse CREATE PROCEDURE statements
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE PROCEDURE") {
			procedure, err := p.parseProcedureStatement(statement)
			if err != nil {
				return err
			}

			err = callback(sqlmapper.SchemaObject{
				Type: sqlmapper.ProcedureObject,
				Data: procedure,
			})
			if err != nil {
				return err
			}
			continue
		}

		// Parse CREATE TRIGGER statements
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE TRIGGER") {
			trigger, err := p.parseTriggerStatement(statement)
			if err != nil {
				return err
			}

			err = callback(sqlmapper.SchemaObject{
				Type: sqlmapper.TriggerObject,
				Data: trigger,
			})
			if err != nil {
				return err
			}
			continue
		}

		// Parse CREATE SEQUENCE statements
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE SEQUENCE") {
			sequence, err := p.parseSequenceStatement(statement)
			if err != nil {
				return err
			}

			err = callback(sqlmapper.SchemaObject{
				Type: sqlmapper.SequenceObject,
				Data: sequence,
			})
			if err != nil {
				return err
			}
			continue
		}

		// Parse CREATE TYPE statements
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE TYPE") {
			typ, err := p.parseTypeStatement(statement)
			if err != nil {
				return err
			}

			err = callback(sqlmapper.SchemaObject{
				Type: sqlmapper.TypeObject,
				Data: typ,
			})
			if err != nil {
				return err
			}
			continue
		}

		// Parse CREATE INDEX statements
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE INDEX") ||
			strings.HasPrefix(strings.ToUpper(statement), "CREATE UNIQUE INDEX") ||
			strings.HasPrefix(strings.ToUpper(statement), "CREATE BITMAP INDEX") {
			index, err := p.parseIndexStatement(statement)
			if err != nil {
				return err
			}

			err = callback(sqlmapper.SchemaObject{
				Type: sqlmapper.IndexObject,
				Data: index,
			})
			if err != nil {
				return err
			}
			continue
		}
	}

	return nil
}

// ParseStreamParallel implements parallel processing for Oracle stream parsing
func (p *OracleStreamParser) ParseStreamParallel(reader io.Reader, callback func(sqlmapper.SchemaObject) error, workers int) error {
	streamReader := sqlmapper.NewStreamReader(reader, "/")
	statements := make(chan string, workers)
	results := make(chan sqlmapper.SchemaObject, workers)
	errors := make(chan error, workers)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for statement := range statements {
				obj, err := p.parseStatement(statement)
				if err != nil {
					errors <- err
					return
				}
				if obj != nil {
					results <- *obj
				}
			}
		}()
	}

	// Start a goroutine to close results channel after all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Start a goroutine to read statements and send them to workers
	go func() {
		for {
			statement, err := streamReader.ReadStatement()
			if err == io.EOF {
				break
			}
			if err != nil {
				errors <- fmt.Errorf("error reading statement: %v", err)
				break
			}

			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}
			statements <- statement
		}
		close(statements)
	}()

	// Process results and handle errors
	for obj := range results {
		if err := callback(obj); err != nil {
			return err
		}
	}

	// Check for any errors from workers
	select {
	case err := <-errors:
		return err
	default:
		return nil
	}
}

// parseStatement parses a single SQL statement and returns a SchemaObject
func (p *OracleStreamParser) parseStatement(statement string) (*sqlmapper.SchemaObject, error) {
	upperStatement := strings.ToUpper(statement)

	switch {
	case strings.HasPrefix(upperStatement, "CREATE TABLE"):
		table, err := p.parseTableStatement(statement)
		if err != nil {
			return nil, err
		}
		return &sqlmapper.SchemaObject{
			Type: sqlmapper.TableObject,
			Data: table,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE VIEW") ||
		strings.HasPrefix(upperStatement, "CREATE MATERIALIZED VIEW"):
		view, err := p.parseViewStatement(statement)
		if err != nil {
			return nil, err
		}
		return &sqlmapper.SchemaObject{
			Type: sqlmapper.ViewObject,
			Data: view,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE FUNCTION"):
		function, err := p.parseFunctionStatement(statement)
		if err != nil {
			return nil, err
		}
		return &sqlmapper.SchemaObject{
			Type: sqlmapper.FunctionObject,
			Data: function,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE PROCEDURE"):
		procedure, err := p.parseProcedureStatement(statement)
		if err != nil {
			return nil, err
		}
		return &sqlmapper.SchemaObject{
			Type: sqlmapper.ProcedureObject,
			Data: procedure,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE TRIGGER"):
		trigger, err := p.parseTriggerStatement(statement)
		if err != nil {
			return nil, err
		}
		return &sqlmapper.SchemaObject{
			Type: sqlmapper.TriggerObject,
			Data: trigger,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE SEQUENCE"):
		sequence, err := p.parseSequenceStatement(statement)
		if err != nil {
			return nil, err
		}
		return &sqlmapper.SchemaObject{
			Type: sqlmapper.SequenceObject,
			Data: sequence,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE TYPE"):
		typ, err := p.parseTypeStatement(statement)
		if err != nil {
			return nil, err
		}
		return &sqlmapper.SchemaObject{
			Type: sqlmapper.TypeObject,
			Data: typ,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE INDEX") ||
		strings.HasPrefix(upperStatement, "CREATE UNIQUE INDEX") ||
		strings.HasPrefix(upperStatement, "CREATE BITMAP INDEX"):
		index, err := p.parseIndexStatement(statement)
		if err != nil {
			return nil, err
		}
		return &sqlmapper.SchemaObject{
			Type: sqlmapper.IndexObject,
			Data: index,
		}, nil
	}

	return nil, nil
}

// GenerateStream implements the StreamParser interface
func (p *OracleStreamParser) GenerateStream(schema *sqlmapper.Schema, writer io.Writer) error {
	// Generate types
	for _, typ := range schema.Types {
		sql := fmt.Sprintf("CREATE TYPE %s AS %s;\n/\n\n", typ.Name, typ.Definition)
		_, err := writer.Write([]byte(sql))
		if err != nil {
			return err
		}
	}

	// Generate sequences
	for _, sequence := range schema.Sequences {
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
		sql += ";\n/\n\n"

		_, err := writer.Write([]byte(sql))
		if err != nil {
			return err
		}
	}

	// Generate tables
	for _, table := range schema.Tables {
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

		sql += "\n);\n/\n\n"

		_, err := writer.Write([]byte(sql))
		if err != nil {
			return err
		}

		// Generate indexes
		for _, index := range table.Indexes {
			if index.IsBitmap {
				sql = "CREATE BITMAP INDEX "
			} else if index.IsUnique {
				sql = "CREATE UNIQUE INDEX "
			} else {
				sql = "CREATE INDEX "
			}

			sql += index.Name + " ON " + table.Name + " (" + strings.Join(index.Columns, ", ") + ");\n/\n"

			_, err := writer.Write([]byte(sql))
			if err != nil {
				return err
			}
		}
	}

	// Generate views
	for _, view := range schema.Views {
		if view.IsMaterialized {
			sql := fmt.Sprintf("CREATE MATERIALIZED VIEW %s AS\n%s;\n/\n\n", view.Name, view.Definition)
			_, err := writer.Write([]byte(sql))
			if err != nil {
				return err
			}
		} else {
			sql := fmt.Sprintf("CREATE VIEW %s AS\n%s;\n/\n\n", view.Name, view.Definition)
			_, err := writer.Write([]byte(sql))
			if err != nil {
				return err
			}
		}
	}

	// Generate functions
	for _, function := range schema.Functions {
		if !function.IsProc {
			sql := fmt.Sprintf("CREATE FUNCTION %s(", function.Name)
			for i, param := range function.Parameters {
				if i > 0 {
					sql += ", "
				}
				sql += fmt.Sprintf("%s %s", param.Name, param.DataType)
			}
			sql += fmt.Sprintf(")\nRETURN %s\nIS\nBEGIN\n%s\nEND;\n/\n\n",
				function.Returns, function.Body)
			_, err := writer.Write([]byte(sql))
			if err != nil {
				return err
			}
		}
	}

	// Generate procedures
	for _, function := range schema.Functions {
		if function.IsProc {
			sql := fmt.Sprintf("CREATE PROCEDURE %s(", function.Name)
			for i, param := range function.Parameters {
				if i > 0 {
					sql += ", "
				}
				sql += fmt.Sprintf("%s %s", param.Name, param.DataType)
			}
			sql += fmt.Sprintf(")\nIS\nBEGIN\n%s\nEND;\n/\n\n", function.Body)
			_, err := writer.Write([]byte(sql))
			if err != nil {
				return err
			}
		}
	}

	// Generate triggers
	for _, trigger := range schema.Triggers {
		sql := fmt.Sprintf("CREATE TRIGGER %s\n%s %s ON %s\n",
			trigger.Name, trigger.Timing, trigger.Event, trigger.Table)
		if trigger.ForEachRow {
			sql += "FOR EACH ROW\n"
		}
		if trigger.Condition != "" {
			sql += "WHEN " + trigger.Condition + "\n"
		}
		sql += fmt.Sprintf("BEGIN\n%s\nEND;\n/\n\n", trigger.Body)
		_, err := writer.Write([]byte(sql))
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *OracleStreamParser) parseTableStatement(statement string) (*sqlmapper.Table, error) {
	// Create a temporary schema to parse the table
	tempSchema := &sqlmapper.Schema{}
	p.oracle.schema = tempSchema

	if err := p.oracle.parseTables(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Tables) == 0 {
		return nil, fmt.Errorf("no table found in statement")
	}

	return &tempSchema.Tables[0], nil
}

func (p *OracleStreamParser) parseViewStatement(statement string) (*sqlmapper.View, error) {
	// Create a temporary schema to parse the view
	tempSchema := &sqlmapper.Schema{}
	p.oracle.schema = tempSchema

	if err := p.oracle.parseViews(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Views) == 0 {
		return nil, fmt.Errorf("no view found in statement")
	}

	return &tempSchema.Views[0], nil
}

func (p *OracleStreamParser) parseFunctionStatement(statement string) (*sqlmapper.Function, error) {
	// Create a temporary schema to parse the function
	tempSchema := &sqlmapper.Schema{}
	p.oracle.schema = tempSchema

	if err := p.oracle.parseFunctions(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Functions) == 0 {
		return nil, fmt.Errorf("no function found in statement")
	}

	return &tempSchema.Functions[0], nil
}

func (p *OracleStreamParser) parseProcedureStatement(statement string) (*sqlmapper.Function, error) {
	// Create a temporary schema to parse the procedure
	tempSchema := &sqlmapper.Schema{}
	p.oracle.schema = tempSchema

	if err := p.oracle.parseFunctions(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Functions) == 0 {
		return nil, fmt.Errorf("no procedure found in statement")
	}

	return &tempSchema.Functions[0], nil
}

func (p *OracleStreamParser) parseTriggerStatement(statement string) (*sqlmapper.Trigger, error) {
	// Create a temporary schema to parse the trigger
	tempSchema := &sqlmapper.Schema{}
	p.oracle.schema = tempSchema

	if err := p.oracle.parseTriggers(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Triggers) == 0 {
		return nil, fmt.Errorf("no trigger found in statement")
	}

	return &tempSchema.Triggers[0], nil
}

func (p *OracleStreamParser) parseSequenceStatement(statement string) (*sqlmapper.Sequence, error) {
	// Create a temporary schema to parse the sequence
	tempSchema := &sqlmapper.Schema{}
	p.oracle.schema = tempSchema

	if err := p.oracle.parseSequences(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Sequences) == 0 {
		return nil, fmt.Errorf("no sequence found in statement")
	}

	return &tempSchema.Sequences[0], nil
}

func (p *OracleStreamParser) parseTypeStatement(statement string) (*sqlmapper.Type, error) {
	// Create a temporary schema to parse the type
	tempSchema := &sqlmapper.Schema{}
	p.oracle.schema = tempSchema

	if err := p.oracle.parseTypes(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Types) == 0 {
		return nil, fmt.Errorf("no type found in statement")
	}

	return &tempSchema.Types[0], nil
}

func (p *OracleStreamParser) parseIndexStatement(statement string) (*sqlmapper.Index, error) {
	// Create a temporary schema to parse the index
	tempSchema := &sqlmapper.Schema{}
	p.oracle.schema = tempSchema

	if err := p.oracle.parseIndexes(statement); err != nil {
		return nil, err
	}

	// Find the first table with indexes
	for _, table := range tempSchema.Tables {
		if len(table.Indexes) > 0 {
			return &table.Indexes[0], nil
		}
	}

	return nil, fmt.Errorf("no index found in statement")
}
