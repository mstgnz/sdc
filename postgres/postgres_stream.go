package postgres

import (
	"fmt"
	"io"
	"strings"

	"github.com/mstgnz/sqlmapper"
)

// PostgreSQLStreamParser implements the StreamParser interface for PostgreSQL
type PostgreSQLStreamParser struct {
	postgres *PostgreSQL
}

// NewPostgreSQLStreamParser creates a new PostgreSQL stream parser
func NewPostgreSQLStreamParser() *PostgreSQLStreamParser {
	return &PostgreSQLStreamParser{
		postgres: NewPostgreSQL(),
	}
}

// ParseStream implements the StreamParser interface
func (p *PostgreSQLStreamParser) ParseStream(reader io.Reader, callback func(sqlmapper.SchemaObject) error) error {
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

		// Parse GRANT/REVOKE statements
		if strings.HasPrefix(strings.ToUpper(statement), "GRANT") ||
			strings.HasPrefix(strings.ToUpper(statement), "REVOKE") {
			permission, err := p.parsePermissionStatement(statement)
			if err != nil {
				return err
			}

			err = callback(sqlmapper.SchemaObject{
				Type: sqlmapper.PermissionObject,
				Data: permission,
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
func (p *PostgreSQLStreamParser) GenerateStream(schema *sqlmapper.Schema, writer io.Writer) error {
	// Generate types
	for _, typ := range schema.Types {
		if typ.Kind == "ENUM" {
			sql := fmt.Sprintf("CREATE TYPE %s AS ENUM (%s);\n\n", typ.Name, typ.Definition)
			_, err := writer.Write([]byte(sql))
			if err != nil {
				return err
			}
		}
	}

	// Generate tables
	for _, table := range schema.Tables {
		sql := "CREATE TABLE " + table.Name + " (\n"

		// Generate columns
		for i, col := range table.Columns {
			sql += "    " + col.Name + " "

			if col.IsPrimaryKey && strings.ToUpper(col.DataType) == "SERIAL" {
				sql += "SERIAL PRIMARY KEY"
			} else {
				sql += col.DataType
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
			sql += index.Name + " ON " + table.Name
			if index.Type != "" {
				sql += " USING " + index.Type
			}
			sql += " (" + strings.Join(index.Columns, ", ") + ");\n"

			_, err := writer.Write([]byte(sql))
			if err != nil {
				return err
			}
		}
	}

	// Generate views
	for _, view := range schema.Views {
		if view.IsMaterialized {
			sql := fmt.Sprintf("CREATE MATERIALIZED VIEW %s AS\n%s;\n\n", view.Name, view.Definition)
			_, err := writer.Write([]byte(sql))
			if err != nil {
				return err
			}
		} else {
			sql := fmt.Sprintf("CREATE VIEW %s AS\n%s;\n\n", view.Name, view.Definition)
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
			sql += fmt.Sprintf(") RETURNS %s\nLANGUAGE %s\nAS $$\n%s\n$$;\n\n",
				function.Returns, function.Language, function.Body)
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
			sql += fmt.Sprintf(")\nLANGUAGE %s\nAS $$\n%s\n$$;\n\n",
				function.Language, function.Body)
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
			sql += "WHEN (" + trigger.Condition + ")\n"
		}
		sql += "EXECUTE FUNCTION " + trigger.Body + ";\n\n"
		_, err := writer.Write([]byte(sql))
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgreSQLStreamParser) parseTypeStatement(statement string) (*sqlmapper.Type, error) {
	// Create a temporary schema to parse the type
	tempSchema := &sqlmapper.Schema{}
	p.postgres.schema = tempSchema

	if err := p.postgres.parseTypes(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Types) == 0 {
		return nil, fmt.Errorf("no type found in statement")
	}

	return &tempSchema.Types[0], nil
}

func (p *PostgreSQLStreamParser) parseTableStatement(statement string) (*sqlmapper.Table, error) {
	// Create a temporary schema to parse the table
	tempSchema := &sqlmapper.Schema{}
	p.postgres.schema = tempSchema

	if err := p.postgres.parseTables(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Tables) == 0 {
		return nil, fmt.Errorf("no table found in statement")
	}

	return &tempSchema.Tables[0], nil
}

func (p *PostgreSQLStreamParser) parseViewStatement(statement string) (*sqlmapper.View, error) {
	// Create a temporary schema to parse the view
	tempSchema := &sqlmapper.Schema{}
	p.postgres.schema = tempSchema

	if err := p.postgres.parseViews(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Views) == 0 {
		return nil, fmt.Errorf("no view found in statement")
	}

	return &tempSchema.Views[0], nil
}

func (p *PostgreSQLStreamParser) parseFunctionStatement(statement string) (*sqlmapper.Function, error) {
	// Create a temporary schema to parse the function
	tempSchema := &sqlmapper.Schema{}
	p.postgres.schema = tempSchema

	if err := p.postgres.parseFunctions(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Functions) == 0 {
		return nil, fmt.Errorf("no function found in statement")
	}

	return &tempSchema.Functions[0], nil
}

func (p *PostgreSQLStreamParser) parseProcedureStatement(statement string) (*sqlmapper.Function, error) {
	// Create a temporary schema to parse the procedure
	tempSchema := &sqlmapper.Schema{}
	p.postgres.schema = tempSchema

	if err := p.postgres.parseFunctions(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Functions) == 0 {
		return nil, fmt.Errorf("no procedure found in statement")
	}

	return &tempSchema.Functions[0], nil
}

func (p *PostgreSQLStreamParser) parseTriggerStatement(statement string) (*sqlmapper.Trigger, error) {
	// Create a temporary schema to parse the trigger
	tempSchema := &sqlmapper.Schema{}
	p.postgres.schema = tempSchema

	if err := p.postgres.parseTriggers(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Triggers) == 0 {
		return nil, fmt.Errorf("no trigger found in statement")
	}

	return &tempSchema.Triggers[0], nil
}

func (p *PostgreSQLStreamParser) parseIndexStatement(statement string) (*sqlmapper.Index, error) {
	// Create a temporary schema to parse the index
	tempSchema := &sqlmapper.Schema{}
	p.postgres.schema = tempSchema

	if err := p.postgres.parseIndexes(statement); err != nil {
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

func (p *PostgreSQLStreamParser) parsePermissionStatement(statement string) (*sqlmapper.Permission, error) {
	// Create a temporary schema to parse the permission
	tempSchema := &sqlmapper.Schema{}
	p.postgres.schema = tempSchema

	if err := p.postgres.parsePermissions(statement); err != nil {
		return nil, err
	}

	if len(tempSchema.Permissions) == 0 {
		return nil, fmt.Errorf("no permission found in statement")
	}

	return &tempSchema.Permissions[0], nil
}
