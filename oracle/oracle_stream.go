package oracle

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/mstgnz/sqlmapper"
	"github.com/mstgnz/sqlmapper/stream"
)

// OracleStreamParser implements the StreamParser interface for Oracle
type OracleStreamParser struct {
	oracle *Oracle
}

// NewOracleStreamParser creates a new Oracle stream parser
func NewOracleStreamParser() *OracleStreamParser {
	return &OracleStreamParser{
		oracle: NewOracle().(*Oracle),
	}
}

// ParseStream implements the StreamParser interface
func (p *OracleStreamParser) ParseStream(reader io.Reader, callback func(stream.SchemaObject) error) error {
	streamReader := stream.NewStreamReader(reader, "/")

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

			err = callback(stream.SchemaObject{
				Type: stream.TableObject,
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

			err = callback(stream.SchemaObject{
				Type: stream.ViewObject,
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

			err = callback(stream.SchemaObject{
				Type: stream.FunctionObject,
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

			err = callback(stream.SchemaObject{
				Type: stream.ProcedureObject,
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

			err = callback(stream.SchemaObject{
				Type: stream.TriggerObject,
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

			err = callback(stream.SchemaObject{
				Type: stream.SequenceObject,
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

			err = callback(stream.SchemaObject{
				Type: stream.TypeObject,
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

			err = callback(stream.SchemaObject{
				Type: stream.IndexObject,
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
func (p *OracleStreamParser) ParseStreamParallel(reader io.Reader, callback func(stream.SchemaObject) error, workers int) error {
	streamReader := stream.NewStreamReader(reader, "/")
	statements := make(chan string, workers)
	results := make(chan stream.SchemaObject, workers)
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
func (p *OracleStreamParser) parseStatement(statement string) (*stream.SchemaObject, error) {
	upperStatement := strings.ToUpper(statement)

	switch {
	case strings.HasPrefix(upperStatement, "CREATE TABLE"):
		table, err := p.parseTableStatement(statement)
		if err != nil {
			return nil, err
		}
		return &stream.SchemaObject{
			Type: stream.TableObject,
			Data: table,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE VIEW"),
		strings.HasPrefix(upperStatement, "CREATE MATERIALIZED VIEW"):
		view, err := p.parseViewStatement(statement)
		if err != nil {
			return nil, err
		}
		return &stream.SchemaObject{
			Type: stream.ViewObject,
			Data: view,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE FUNCTION"):
		function, err := p.parseFunctionStatement(statement)
		if err != nil {
			return nil, err
		}
		return &stream.SchemaObject{
			Type: stream.FunctionObject,
			Data: function,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE PROCEDURE"):
		procedure, err := p.parseProcedureStatement(statement)
		if err != nil {
			return nil, err
		}
		return &stream.SchemaObject{
			Type: stream.ProcedureObject,
			Data: procedure,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE TRIGGER"):
		trigger, err := p.parseTriggerStatement(statement)
		if err != nil {
			return nil, err
		}
		return &stream.SchemaObject{
			Type: stream.TriggerObject,
			Data: trigger,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE SEQUENCE"):
		sequence, err := p.parseSequenceStatement(statement)
		if err != nil {
			return nil, err
		}
		return &stream.SchemaObject{
			Type: stream.SequenceObject,
			Data: sequence,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE TYPE"):
		typ, err := p.parseTypeStatement(statement)
		if err != nil {
			return nil, err
		}
		return &stream.SchemaObject{
			Type: stream.TypeObject,
			Data: typ,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE INDEX"),
		strings.HasPrefix(upperStatement, "CREATE UNIQUE INDEX"),
		strings.HasPrefix(upperStatement, "CREATE BITMAP INDEX"):
		index, err := p.parseIndexStatement(statement)
		if err != nil {
			return nil, err
		}
		return &stream.SchemaObject{
			Type: stream.IndexObject,
			Data: index,
		}, nil
	}

	return nil, nil
}

// parseTableStatement parses a CREATE TABLE statement
func (p *OracleStreamParser) parseTableStatement(statement string) (*sqlmapper.Table, error) {
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

// parseViewStatement parses a CREATE VIEW statement
func (p *OracleStreamParser) parseViewStatement(statement string) (*sqlmapper.View, error) {
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

// parseFunctionStatement parses a CREATE FUNCTION statement
func (p *OracleStreamParser) parseFunctionStatement(statement string) (*sqlmapper.Function, error) {
	tempSchema := &sqlmapper.Schema{}
	p.oracle.schema = tempSchema

	if err := p.oracle.parseFunctions(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Functions) == 0 {
		return nil, fmt.Errorf("no function found in statement")
	}

	for _, fn := range tempSchema.Functions {
		if !fn.IsProc {
			return &fn, nil
		}
	}

	return nil, fmt.Errorf("no function found in statement")
}

// parseProcedureStatement parses a CREATE PROCEDURE statement
func (p *OracleStreamParser) parseProcedureStatement(statement string) (*sqlmapper.Procedure, error) {
	tempSchema := &sqlmapper.Schema{}
	p.oracle.schema = tempSchema

	if err := p.oracle.parseFunctions(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Functions) == 0 {
		return nil, fmt.Errorf("no procedure found in statement")
	}

	for _, fn := range tempSchema.Functions {
		if fn.IsProc {
			proc := &sqlmapper.Procedure{
				Name:       fn.Name,
				Parameters: fn.Parameters,
				Body:       fn.Body,
				Schema:     fn.Schema,
			}
			return proc, nil
		}
	}

	return nil, fmt.Errorf("no procedure found in statement")
}

// parseTriggerStatement parses a CREATE TRIGGER statement
func (p *OracleStreamParser) parseTriggerStatement(statement string) (*sqlmapper.Trigger, error) {
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

// parseSequenceStatement parses a CREATE SEQUENCE statement
func (p *OracleStreamParser) parseSequenceStatement(statement string) (*sqlmapper.Sequence, error) {
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

// parseTypeStatement parses a CREATE TYPE statement
func (p *OracleStreamParser) parseTypeStatement(statement string) (*sqlmapper.Type, error) {
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

// parseIndexStatement parses a CREATE INDEX statement
func (p *OracleStreamParser) parseIndexStatement(statement string) (*sqlmapper.Index, error) {
	tempSchema := &sqlmapper.Schema{}
	p.oracle.schema = tempSchema

	if err := p.oracle.parseIndexes(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Tables) == 0 || len(tempSchema.Tables[0].Indexes) == 0 {
		return nil, fmt.Errorf("no index found in statement")
	}

	return &tempSchema.Tables[0].Indexes[0], nil
}

// GenerateStream implements the StreamParser interface
func (p *OracleStreamParser) GenerateStream(schema *sqlmapper.Schema, writer io.Writer) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	// Write sequences
	for _, sequence := range schema.Sequences {
		stmt := p.oracle.generateSequenceSQL(sequence)
		if _, err := writer.Write([]byte(stmt + ";\n\n")); err != nil {
			return err
		}
	}

	// Write types
	for _, typ := range schema.Types {
		stmt := p.oracle.generateTypeSQL(typ)
		if _, err := writer.Write([]byte(stmt + ";\n\n")); err != nil {
			return err
		}
	}

	// Write tables
	for _, table := range schema.Tables {
		stmt := p.oracle.generateTableSQL(table)
		if _, err := writer.Write([]byte(stmt + ";\n\n")); err != nil {
			return err
		}

		// Generate indexes for this table
		for _, index := range table.Indexes {
			stmt := p.oracle.generateIndexSQL(table.Name, index)
			if _, err := writer.Write([]byte(stmt + ";\n")); err != nil {
				return err
			}
		}
	}

	// Write views
	for _, view := range schema.Views {
		stmt := fmt.Sprintf("CREATE VIEW %s AS %s", view.Name, view.Definition)
		if _, err := writer.Write([]byte(stmt + ";\n\n")); err != nil {
			return err
		}
	}

	// Write functions
	for _, function := range schema.Functions {
		if !function.IsProc {
			stmt := fmt.Sprintf("CREATE FUNCTION %s(", function.Name)
			for i, param := range function.Parameters {
				if i > 0 {
					stmt += ", "
				}
				stmt += fmt.Sprintf("%s %s", param.Name, param.DataType)
			}
			stmt += fmt.Sprintf(") RETURN %s\n%s", function.Returns, function.Body)
			if _, err := writer.Write([]byte(stmt + ";\n\n")); err != nil {
				return err
			}
		}
	}

	// Write procedures
	for _, function := range schema.Functions {
		if function.IsProc {
			stmt := fmt.Sprintf("CREATE PROCEDURE %s(", function.Name)
			for i, param := range function.Parameters {
				if i > 0 {
					stmt += ", "
				}
				stmt += fmt.Sprintf("%s %s", param.Name, param.DataType)
			}
			stmt += fmt.Sprintf(")\n%s", function.Body)
			if _, err := writer.Write([]byte(stmt + ";\n\n")); err != nil {
				return err
			}
		}
	}

	// Write triggers
	for _, trigger := range schema.Triggers {
		stmt := fmt.Sprintf("CREATE TRIGGER %s %s %s ON %s\n%s",
			trigger.Name, trigger.Timing, trigger.Event, trigger.Table, trigger.Body)
		if _, err := writer.Write([]byte(stmt + ";\n\n")); err != nil {
			return err
		}
	}

	return nil
}
