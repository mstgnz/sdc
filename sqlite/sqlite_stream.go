package sqlite

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/mstgnz/sqlmapper"
)

// SQLiteStreamParser implements the StreamParser interface for SQLite
type SQLiteStreamParser struct {
	sqlite *SQLite
}

// NewSQLiteStreamParser creates a new SQLite stream parser
func NewSQLiteStreamParser() *SQLiteStreamParser {
	return &SQLiteStreamParser{
		sqlite: NewSQLite(),
	}
}

// ParseStream implements the StreamParser interface
func (p *SQLiteStreamParser) ParseStream(reader io.Reader, callback func(sqlmapper.SchemaObject) error) error {
	streamReader := sqlmapper.NewStreamReader(reader, ";")

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

		// Parse CREATE INDEX statements
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE INDEX") ||
			strings.HasPrefix(strings.ToUpper(statement), "CREATE UNIQUE INDEX") {
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
	}

	return nil
}

// ParseStreamParallel implements parallel processing for SQLite stream parsing
func (p *SQLiteStreamParser) ParseStreamParallel(reader io.Reader, callback func(sqlmapper.SchemaObject) error, workers int) error {
	streamReader := sqlmapper.NewStreamReader(reader, ";")
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
func (p *SQLiteStreamParser) parseStatement(statement string) (*sqlmapper.SchemaObject, error) {
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

	case strings.HasPrefix(upperStatement, "CREATE INDEX"):
		index, err := p.parseIndexStatement(statement)
		if err != nil {
			return nil, err
		}
		return &sqlmapper.SchemaObject{
			Type: sqlmapper.IndexObject,
			Data: index,
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
	}

	return nil, nil
}

// GenerateStream implements the StreamParser interface
func (p *SQLiteStreamParser) GenerateStream(schema *sqlmapper.Schema, writer io.Writer) error {
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

			if col.IsPrimaryKey {
				sql += " PRIMARY KEY"
				if col.AutoIncrement {
					sql += " AUTOINCREMENT"
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

		sql += "\n);\n\n"

		_, err := writer.Write([]byte(sql))
		if err != nil {
			return err
		}

		// Generate indexes
		for _, index := range table.Indexes {
			if index.IsUnique {
				sql = "CREATE UNIQUE INDEX "
			} else {
				sql = "CREATE INDEX "
			}
			sql += index.Name + " ON " + table.Name + " (" + strings.Join(index.Columns, ", ") + ");\n"

			_, err := writer.Write([]byte(sql))
			if err != nil {
				return err
			}
		}
	}

	// Generate views
	for _, view := range schema.Views {
		sql := fmt.Sprintf("CREATE VIEW %s AS\n%s;\n\n", view.Name, view.Definition)
		_, err := writer.Write([]byte(sql))
		if err != nil {
			return err
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
		sql += "BEGIN\n" + trigger.Body + "\nEND;\n\n"
		_, err := writer.Write([]byte(sql))
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *SQLiteStreamParser) parseTableStatement(statement string) (*sqlmapper.Table, error) {
	// Create a temporary schema to parse the table
	tempSchema := &sqlmapper.Schema{}
	p.sqlite.schema = tempSchema

	if err := p.sqlite.parseTables(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Tables) == 0 {
		return nil, fmt.Errorf("no table found in statement")
	}

	return &tempSchema.Tables[0], nil
}

func (p *SQLiteStreamParser) parseViewStatement(statement string) (*sqlmapper.View, error) {
	// Create a temporary schema to parse the view
	tempSchema := &sqlmapper.Schema{}
	p.sqlite.schema = tempSchema

	if err := p.sqlite.parseViews(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Views) == 0 {
		return nil, fmt.Errorf("no view found in statement")
	}

	return &tempSchema.Views[0], nil
}

func (p *SQLiteStreamParser) parseIndexStatement(statement string) (*sqlmapper.Index, error) {
	// Create a temporary schema to parse the index
	tempSchema := &sqlmapper.Schema{}
	p.sqlite.schema = tempSchema

	if err := p.sqlite.parseIndexes(statement); err != nil {
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

func (p *SQLiteStreamParser) parseTriggerStatement(statement string) (*sqlmapper.Trigger, error) {
	// Create a temporary schema to parse the trigger
	tempSchema := &sqlmapper.Schema{}
	p.sqlite.schema = tempSchema

	if err := p.sqlite.parseTriggers(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Triggers) == 0 {
		return nil, fmt.Errorf("no trigger found in statement")
	}

	return &tempSchema.Triggers[0], nil
}
