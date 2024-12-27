package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mstgnz/sqlmapper"
	"github.com/mstgnz/sqlmapper/mysql"
	"github.com/mstgnz/sqlmapper/postgres"
	"github.com/stretchr/testify/assert"
)

func TestMySQLStreamParser(t *testing.T) {
	parser := mysql.NewMySQLStreamParser()

	// Test input
	input := `
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    status ENUM('active', 'inactive') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE VIEW active_users AS
SELECT * FROM users WHERE status = 'active';

CREATE FUNCTION get_user_status(user_id INT)
RETURNS VARCHAR(50)
BEGIN
    DECLARE status VARCHAR(50);
    SELECT status INTO status FROM users WHERE id = user_id;
    RETURN status;
END;

CREATE TRIGGER before_user_update
BEFORE UPDATE ON users
FOR EACH ROW
BEGIN
    SET NEW.updated_at = CURRENT_TIMESTAMP;
END;
`

	// Test parsing
	var objects []sqlmapper.SchemaObject
	err := parser.ParseStream(strings.NewReader(input), func(obj sqlmapper.SchemaObject) error {
		objects = append(objects, obj)
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 4, len(objects))

	// Test table object
	assert.Equal(t, sqlmapper.TableObject, objects[0].Type)
	table := objects[0].Data.(*sqlmapper.Table)
	assert.Equal(t, "users", table.Name)
	assert.Equal(t, 5, len(table.Columns))

	// Test view object
	assert.Equal(t, sqlmapper.ViewObject, objects[1].Type)
	view := objects[1].Data.(*sqlmapper.View)
	assert.Equal(t, "active_users", view.Name)

	// Test function object
	assert.Equal(t, sqlmapper.FunctionObject, objects[2].Type)
	function := objects[2].Data.(*sqlmapper.Function)
	assert.Equal(t, "get_user_status", function.Name)

	// Test trigger object
	assert.Equal(t, sqlmapper.TriggerObject, objects[3].Type)
	trigger := objects[3].Data.(*sqlmapper.Trigger)
	assert.Equal(t, "before_user_update", trigger.Name)

	// Test generation
	var output bytes.Buffer
	schema := &sqlmapper.Schema{
		Tables:    []sqlmapper.Table{*table},
		Views:     []sqlmapper.View{*view},
		Functions: []sqlmapper.Function{*function},
		Triggers:  []sqlmapper.Trigger{*trigger},
	}

	err = parser.GenerateStream(schema, &output)
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "CREATE TABLE users")
	assert.Contains(t, output.String(), "CREATE VIEW active_users")
	assert.Contains(t, output.String(), "CREATE FUNCTION get_user_status")
	assert.Contains(t, output.String(), "CREATE TRIGGER before_user_update")
}

func TestPostgreSQLStreamParser(t *testing.T) {
	parser := postgres.NewPostgreSQLStreamParser()

	// Test input
	input := `
CREATE TYPE user_status AS ENUM ('active', 'inactive');

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    status user_status DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users USING btree (email);

CREATE MATERIALIZED VIEW active_users AS
SELECT * FROM users WHERE status = 'active'
WITH DATA;

CREATE FUNCTION get_user_status(user_id INTEGER)
RETURNS user_status
LANGUAGE plpgsql
AS $$
DECLARE
    status user_status;
BEGIN
    SELECT status INTO status FROM users WHERE id = user_id;
    RETURN status;
END;
$$;

CREATE TRIGGER before_user_update
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();

GRANT SELECT, INSERT ON users TO app_user;
`

	// Test parsing
	var objects []sqlmapper.SchemaObject
	err := parser.ParseStream(strings.NewReader(input), func(obj sqlmapper.SchemaObject) error {
		objects = append(objects, obj)
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 7, len(objects))

	// Test type object
	assert.Equal(t, sqlmapper.TypeObject, objects[0].Type)
	typ := objects[0].Data.(*sqlmapper.Type)
	assert.Equal(t, "user_status", typ.Name)

	// Test table object
	assert.Equal(t, sqlmapper.TableObject, objects[1].Type)
	table := objects[1].Data.(*sqlmapper.Table)
	assert.Equal(t, "users", table.Name)
	assert.Equal(t, 5, len(table.Columns))

	// Test index object
	assert.Equal(t, sqlmapper.IndexObject, objects[2].Type)
	index := objects[2].Data.(*sqlmapper.Index)
	assert.Equal(t, "idx_users_email", index.Name)

	// Test view object
	assert.Equal(t, sqlmapper.ViewObject, objects[3].Type)
	view := objects[3].Data.(*sqlmapper.View)
	assert.Equal(t, "active_users", view.Name)
	assert.True(t, view.IsMaterialized)

	// Test function object
	assert.Equal(t, sqlmapper.FunctionObject, objects[4].Type)
	function := objects[4].Data.(*sqlmapper.Function)
	assert.Equal(t, "get_user_status", function.Name)

	// Test trigger object
	assert.Equal(t, sqlmapper.TriggerObject, objects[5].Type)
	trigger := objects[5].Data.(*sqlmapper.Trigger)
	assert.Equal(t, "before_user_update", trigger.Name)

	// Test permission object
	assert.Equal(t, sqlmapper.PermissionObject, objects[6].Type)
	permission := objects[6].Data.(*sqlmapper.Permission)
	assert.Equal(t, "GRANT", permission.Type)

	// Test generation
	var output bytes.Buffer
	schema := &sqlmapper.Schema{
		Types:       []sqlmapper.Type{*typ},
		Tables:      []sqlmapper.Table{*table},
		Views:       []sqlmapper.View{*view},
		Functions:   []sqlmapper.Function{*function},
		Triggers:    []sqlmapper.Trigger{*trigger},
		Permissions: []sqlmapper.Permission{*permission},
	}

	err = parser.GenerateStream(schema, &output)
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "CREATE TYPE user_status")
	assert.Contains(t, output.String(), "CREATE TABLE users")
	assert.Contains(t, output.String(), "CREATE MATERIALIZED VIEW active_users")
	assert.Contains(t, output.String(), "CREATE FUNCTION get_user_status")
	assert.Contains(t, output.String(), "CREATE TRIGGER before_user_update")
}
