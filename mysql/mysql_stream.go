package mysql

import (
	"fmt"
	"io"
	"strings"

	"github.com/mstgnz/sqlmapper"
)

// MySQLStreamParser implements the StreamParser interface for MySQL
type MySQLStreamParser struct {
	mysql *MySQL
}

// NewMySQLStreamParser creates a new MySQL stream parser
func NewMySQLStreamParser() *MySQLStreamParser {
	return &MySQLStreamParser{
		mysql: NewMySQL(),
	}
}

// ParseStream implements the StreamParser interface
func (p *MySQLStreamParser) ParseStream(reader io.Reader, callback func(sqlmapper.SchemaObject) error) error {
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
	}

	return nil
}

// GenerateStream implements the StreamParser interface
func (p *MySQLStreamParser) GenerateStream(schema *sqlmapper.Schema, writer io.Writer) error {
	// Generate tables
	for _, table := range schema.Tables {
		sql := p.mysql.generateTableSQL(table)
		_, err := writer.Write([]byte(sql + ";\n\n"))
		if err != nil {
			return err
		}

		// Generate indexes for this table
		for _, index := range table.Indexes {
			sql := p.mysql.generateIndexSQL(table.Name, index)
			_, err := writer.Write([]byte(sql + ";\n"))
			if err != nil {
				return err
			}
		}
	}

	// Generate views
	for _, view := range schema.Views {
		sql := fmt.Sprintf("CREATE VIEW %s AS %s", view.Name, view.Definition)
		_, err := writer.Write([]byte(sql + ";\n\n"))
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
				sql += fmt.Sprintf("%s %s", param.Name, param.DataType)
			}
			sql += fmt.Sprintf(") RETURNS %s\n%s", function.Returns, function.Body)
			_, err := writer.Write([]byte(sql + ";\n\n"))
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
			sql += fmt.Sprintf(")\n%s", function.Body)
			_, err := writer.Write([]byte(sql + ";\n\n"))
			if err != nil {
				return err
			}
		}
	}

	// Generate triggers
	for _, trigger := range schema.Triggers {
		sql := fmt.Sprintf("CREATE TRIGGER %s %s %s ON %s\n%s",
			trigger.Name, trigger.Timing, trigger.Event, trigger.Table, trigger.Body)
		_, err := writer.Write([]byte(sql + ";\n\n"))
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *MySQLStreamParser) parseTableStatement(statement string) (*sqlmapper.Table, error) {
	// Create a temporary schema to parse the table
	tempSchema := &sqlmapper.Schema{}
	p.mysql.schema = tempSchema

	if err := p.mysql.parseTables(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Tables) == 0 {
		return nil, fmt.Errorf("no table found in statement")
	}

	return &tempSchema.Tables[0], nil
}

func (p *MySQLStreamParser) parseViewStatement(statement string) (*sqlmapper.View, error) {
	// Create a temporary schema to parse the view
	tempSchema := &sqlmapper.Schema{}
	p.mysql.schema = tempSchema

	if err := p.mysql.parseViews(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Views) == 0 {
		return nil, fmt.Errorf("no view found in statement")
	}

	return &tempSchema.Views[0], nil
}

func (p *MySQLStreamParser) parseFunctionStatement(statement string) (*sqlmapper.Function, error) {
	// Create a temporary schema to parse the function
	tempSchema := &sqlmapper.Schema{}
	p.mysql.schema = tempSchema

	if err := p.mysql.parseFunctions(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Functions) == 0 {
		return nil, fmt.Errorf("no function found in statement")
	}

	return &tempSchema.Functions[0], nil
}

func (p *MySQLStreamParser) parseProcedureStatement(statement string) (*sqlmapper.Function, error) {
	// Create a temporary schema to parse the procedure
	tempSchema := &sqlmapper.Schema{}
	p.mysql.schema = tempSchema

	if err := p.mysql.parseFunctions(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Functions) == 0 {
		return nil, fmt.Errorf("no procedure found in statement")
	}

	return &tempSchema.Functions[0], nil
}

func (p *MySQLStreamParser) parseTriggerStatement(statement string) (*sqlmapper.Trigger, error) {
	// Create a temporary schema to parse the trigger
	tempSchema := &sqlmapper.Schema{}
	p.mysql.schema = tempSchema

	if err := p.mysql.parseTriggers(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Triggers) == 0 {
		return nil, fmt.Errorf("no trigger found in statement")
	}

	return &tempSchema.Triggers[0], nil
}
