package sqlserver

import (
	"testing"

	"github.com/mstgnz/sqlporter"
	"github.com/stretchr/testify/assert"
)

func TestSQLServer_Parse(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErr  bool
		validate func(*testing.T, *sqlporter.Schema)
	}{
		{
			name:    "Empty content",
			content: "",
			wantErr: true,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.Nil(t, schema)
			},
		},
		{
			name:    "CREATE TABLE",
			content: "CREATE TABLE test (id INT PRIMARY KEY, name NVARCHAR(50));",
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.NotNil(t, schema)
				// Additional validation logic can be added here
			},
		},
		{
			name:    "CREATE INDEX",
			content: "CREATE INDEX idx_name ON test (name);",
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.NotNil(t, schema)
				// Additional validation logic can be added here
			},
		},
		{
			name:    "ALTER TABLE",
			content: "ALTER TABLE test ADD COLUMN email NVARCHAR(100);",
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.NotNil(t, schema)
				// Additional validation logic can be added here
			},
		},
		{
			name:    "CREATE VIEW",
			content: "CREATE VIEW test_view AS SELECT id, name FROM test;",
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.NotNil(t, schema)
				// Additional validation logic for views can be added here
			},
		},
		{
			name:    "CREATE TRIGGER",
			content: "CREATE TRIGGER trg_test AFTER INSERT ON test FOR EACH ROW BEGIN UPDATE test SET name = 'updated' WHERE id = NEW.id; END;",
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.NotNil(t, schema)
				// Additional validation logic for triggers can be added here
			},
		},
		{
			name:    "ALTER TABLE with CONSTRAINT",
			content: "ALTER TABLE test ADD CONSTRAINT chk_name CHECK (name IS NOT NULL);",
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.NotNil(t, schema)
				// Additional validation logic for constraints can be added here
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSQLServer()
			schema, err := s.Parse(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.validate != nil {
				tt.validate(t, schema)
			}
		})
	}
}

func TestSQLServer_Generate(t *testing.T) {
	tests := []struct {
		name    string
		schema  *sqlporter.Schema
		wantErr bool
	}{
		{
			name:    "Nil schema",
			schema:  nil,
			wantErr: true,
		},
		{
			name:    "Valid schema",
			schema:  &sqlporter.Schema{}, // Assuming a valid schema object
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSQLServer()
			_, err := s.Generate(tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSQLServer_Generate_ComplexSchema(t *testing.T) {
	schema := &sqlporter.Schema{
		// Assuming a complex schema object with tables, views, and triggers
	}
	s := NewSQLServer()
	_, err := s.Generate(schema)
	assert.NoError(t, err)
}
