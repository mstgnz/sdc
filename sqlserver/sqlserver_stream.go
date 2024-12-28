package sqlserver

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/mstgnz/sqlmapper"
)

// SQLServerStreamParser implements the StreamParser interface for SQL Server
type SQLServerStreamParser struct {
	sqlserver *SQLServer
}

// NewSQLServerStreamParser creates a new SQL Server stream parser
func NewSQLServerStreamParser() *SQLServerStreamParser {
	return &SQLServerStreamParser{
		sqlserver: NewSQLServer(),
	}
}

// ParseStream implements the StreamParser interface
func (p *SQLServerStreamParser) ParseStream(reader io.Reader, callback func(sqlmapper.SchemaObject) error) error {
	streamReader := sqlmapper.NewStreamReader(reader, "GO")

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
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE VIEW") {
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
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE PROCEDURE") ||
			strings.HasPrefix(strings.ToUpper(statement), "CREATE PROC") {
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

		// Parse CREATE INDEX statements
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE INDEX") ||
			strings.HasPrefix(strings.ToUpper(statement), "CREATE UNIQUE INDEX") ||
			strings.HasPrefix(strings.ToUpper(statement), "CREATE CLUSTERED INDEX") ||
			strings.HasPrefix(strings.ToUpper(statement), "CREATE NONCLUSTERED INDEX") {
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

// ParseStreamParallel implements parallel processing for SQL Server stream parsing
func (p *SQLServerStreamParser) ParseStreamParallel(reader io.Reader, callback func(sqlmapper.SchemaObject) error, workers int) error {
	streamReader := sqlmapper.NewStreamReader(reader, "GO")
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
func (p *SQLServerStreamParser) parseStatement(statement string) (*sqlmapper.SchemaObject, error) {
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

	case strings.HasPrefix(upperStatement, "CREATE VIEW"):
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

	case strings.HasPrefix(upperStatement, "CREATE PROCEDURE") || strings.HasPrefix(upperStatement, "CREATE PROC"):
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

	case strings.HasPrefix(upperStatement, "CREATE INDEX") ||
		strings.HasPrefix(upperStatement, "CREATE UNIQUE INDEX") ||
		strings.HasPrefix(upperStatement, "CREATE CLUSTERED INDEX") ||
		strings.HasPrefix(upperStatement, "CREATE NONCLUSTERED INDEX"):
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
func (p *SQLServerStreamParser) GenerateStream(schema *sqlmapper.Schema, writer io.Writer) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	// Write tables
	for _, table := range schema.Tables {
		stmt := p.sqlserver.generateTableSQL(table)
		if _, err := writer.Write([]byte(stmt + "\nGO\n\n")); err != nil {
			return err
		}

		// Generate indexes for this table
		for _, index := range table.Indexes {
			stmt := p.sqlserver.generateIndexSQL(table.Name, index)
			if _, err := writer.Write([]byte(stmt + "\nGO\n")); err != nil {
				return err
			}
		}
	}

	// Write views
	for _, view := range schema.Views {
		stmt := fmt.Sprintf("CREATE VIEW %s AS\n%s", view.Name, view.Definition)
		if _, err := writer.Write([]byte(stmt + "\nGO\n\n")); err != nil {
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
				stmt += fmt.Sprintf("@%s %s", param.Name, param.DataType)
			}
			stmt += fmt.Sprintf(")\nRETURNS %s\nAS\nBEGIN\n%s\nEND",
				function.Returns, function.Body)
			if _, err := writer.Write([]byte(stmt + "\nGO\n\n")); err != nil {
				return err
			}
		}
	}

	// Write procedures
	for _, function := range schema.Functions {
		if function.IsProc {
			stmt := fmt.Sprintf("CREATE PROCEDURE %s", function.Name)
			if len(function.Parameters) > 0 {
				stmt += "("
				for i, param := range function.Parameters {
					if i > 0 {
						stmt += ", "
					}
					stmt += fmt.Sprintf("@%s %s", param.Name, param.DataType)
				}
				stmt += ")"
			}
			stmt += fmt.Sprintf("\nAS\nBEGIN\n%s\nEND", function.Body)
			if _, err := writer.Write([]byte(stmt + "\nGO\n\n")); err != nil {
				return err
			}
		}
	}

	// Write triggers
	for _, trigger := range schema.Triggers {
		stmt := fmt.Sprintf("CREATE TRIGGER %s ON %s\n%s %s\nAS\nBEGIN\n%s\nEND",
			trigger.Name, trigger.Table, trigger.Timing, trigger.Event, trigger.Body)
		if _, err := writer.Write([]byte(stmt + "\nGO\n\n")); err != nil {
			return err
		}
	}

	return nil
}

// parseTableStatement parses a CREATE TABLE statement
func (p *SQLServerStreamParser) parseTableStatement(statement string) (*sqlmapper.Table, error) {
	tempSchema := &sqlmapper.Schema{}
	p.sqlserver.schema = tempSchema

	if err := p.sqlserver.parseTables(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Tables) == 0 {
		return nil, fmt.Errorf("no table found in statement")
	}

	return &tempSchema.Tables[0], nil
}

// parseViewStatement parses a CREATE VIEW statement
func (p *SQLServerStreamParser) parseViewStatement(statement string) (*sqlmapper.View, error) {
	tempSchema := &sqlmapper.Schema{}
	p.sqlserver.schema = tempSchema

	if err := p.sqlserver.parseViews(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Views) == 0 {
		return nil, fmt.Errorf("no view found in statement")
	}

	return &tempSchema.Views[0], nil
}

// parseFunctionStatement parses a CREATE FUNCTION statement
func (p *SQLServerStreamParser) parseFunctionStatement(statement string) (*sqlmapper.Function, error) {
	tempSchema := &sqlmapper.Schema{}
	p.sqlserver.schema = tempSchema

	if err := p.sqlserver.parseFunctions(statement); err != nil {
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
func (p *SQLServerStreamParser) parseProcedureStatement(statement string) (*sqlmapper.Procedure, error) {
	tempSchema := &sqlmapper.Schema{}
	p.sqlserver.schema = tempSchema

	if err := p.sqlserver.parseFunctions(statement); err != nil {
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
func (p *SQLServerStreamParser) parseTriggerStatement(statement string) (*sqlmapper.Trigger, error) {
	tempSchema := &sqlmapper.Schema{}
	p.sqlserver.schema = tempSchema

	if err := p.sqlserver.parseTriggers(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Triggers) == 0 {
		return nil, fmt.Errorf("no trigger found in statement")
	}

	return &tempSchema.Triggers[0], nil
}

// parseIndexStatement parses a CREATE INDEX statement
func (p *SQLServerStreamParser) parseIndexStatement(statement string) (*sqlmapper.Index, error) {
	tempSchema := &sqlmapper.Schema{}
	p.sqlserver.schema = tempSchema

	if err := p.sqlserver.parseIndexes(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Tables) == 0 || len(tempSchema.Tables[0].Indexes) == 0 {
		return nil, fmt.Errorf("no index found in statement")
	}

	return &tempSchema.Tables[0].Indexes[0], nil
}
