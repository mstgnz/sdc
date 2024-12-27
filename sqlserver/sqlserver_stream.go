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
	// Generate tables
	for _, table := range schema.Tables {
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

		sql += "\n);\nGO\n\n"

		_, err := writer.Write([]byte(sql))
		if err != nil {
			return err
		}

		// Generate indexes
		for _, index := range table.Indexes {
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

			sql += index.Name + " ON " + table.Name + " (" + strings.Join(index.Columns, ", ") + ");\nGO\n"

			_, err := writer.Write([]byte(sql))
			if err != nil {
				return err
			}
		}
	}

	// Generate views
	for _, view := range schema.Views {
		sql := fmt.Sprintf("CREATE VIEW %s AS\n%s;\nGO\n\n", view.Name, view.Definition)
		_, err := writer.Write([]byte(sql))
		if err != nil {
			return err
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
				sql += fmt.Sprintf("@%s %s", param.Name, param.DataType)
			}
			sql += fmt.Sprintf(")\nRETURNS %s\nAS\nBEGIN\n%s\nEND;\nGO\n\n",
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
			sql := fmt.Sprintf("CREATE PROCEDURE %s", function.Name)
			if len(function.Parameters) > 0 {
				sql += "(\n"
				for i, param := range function.Parameters {
					if i > 0 {
						sql += ",\n"
					}
					sql += fmt.Sprintf("    @%s %s", param.Name, param.DataType)
				}
				sql += "\n)"
			}
			sql += fmt.Sprintf("\nAS\nBEGIN\n%s\nEND;\nGO\n\n", function.Body)
			_, err := writer.Write([]byte(sql))
			if err != nil {
				return err
			}
		}
	}

	// Generate triggers
	for _, trigger := range schema.Triggers {
		sql := fmt.Sprintf("CREATE TRIGGER %s\nON %s\n%s %s\nAS\nBEGIN\n%s\nEND;\nGO\n\n",
			trigger.Name, trigger.Table, trigger.Timing, trigger.Event, trigger.Body)
		_, err := writer.Write([]byte(sql))
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *SQLServerStreamParser) parseTableStatement(statement string) (*sqlmapper.Table, error) {
	// Create a temporary schema to parse the table
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

func (p *SQLServerStreamParser) parseViewStatement(statement string) (*sqlmapper.View, error) {
	// Create a temporary schema to parse the view
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

func (p *SQLServerStreamParser) parseFunctionStatement(statement string) (*sqlmapper.Function, error) {
	// Create a temporary schema to parse the function
	tempSchema := &sqlmapper.Schema{}
	p.sqlserver.schema = tempSchema

	if err := p.sqlserver.parseFunctions(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Functions) == 0 {
		return nil, fmt.Errorf("no function found in statement")
	}

	return &tempSchema.Functions[0], nil
}

func (p *SQLServerStreamParser) parseProcedureStatement(statement string) (*sqlmapper.Function, error) {
	// Create a temporary schema to parse the procedure
	tempSchema := &sqlmapper.Schema{}
	p.sqlserver.schema = tempSchema

	if err := p.sqlserver.parseFunctions(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Functions) == 0 {
		return nil, fmt.Errorf("no procedure found in statement")
	}

	return &tempSchema.Functions[0], nil
}

func (p *SQLServerStreamParser) parseTriggerStatement(statement string) (*sqlmapper.Trigger, error) {
	// Create a temporary schema to parse the trigger
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

func (p *SQLServerStreamParser) parseIndexStatement(statement string) (*sqlmapper.Index, error) {
	// Create a temporary schema to parse the index
	tempSchema := &sqlmapper.Schema{}
	p.sqlserver.schema = tempSchema

	if err := p.sqlserver.parseIndexes(statement); err != nil {
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
