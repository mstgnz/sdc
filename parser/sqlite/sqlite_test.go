package sqlite

import (
	"testing"

	"github.com/mstgnz/sqlporter"
	"github.com/stretchr/testify/assert"
)

func TestSQLiteParser_ParseCreateTable(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sqlporter.Table
		wantErr  bool
	}{
		{
			name: "Table with all features",
			sql: `CREATE TABLE users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				username TEXT NOT NULL UNIQUE,
				email TEXT NOT NULL UNIQUE,
				password TEXT NOT NULL,
				full_name TEXT,
				age INTEGER CHECK (age >= 18),
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME,
				CONSTRAINT users_email_check CHECK (email REGEXP '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
			);`,
			expected: &sqlporter.Table{
				Name: "users",
				Columns: []*sqlporter.Column{
					{
						Name:       "id",
						DataType:   &sqlporter.DataType{Name: "INTEGER"},
						IsNullable: false,
						Extra:      "auto_increment",
					},
					{
						Name:       "username",
						DataType:   &sqlporter.DataType{Name: "TEXT"},
						IsNullable: false,
					},
					{
						Name:       "email",
						DataType:   &sqlporter.DataType{Name: "TEXT"},
						IsNullable: false,
					},
					{
						Name:       "password",
						DataType:   &sqlporter.DataType{Name: "TEXT"},
						IsNullable: false,
					},
					{
						Name:       "full_name",
						DataType:   &sqlporter.DataType{Name: "TEXT"},
						IsNullable: true,
					},
					{
						Name:       "age",
						DataType:   &sqlporter.DataType{Name: "INTEGER"},
						IsNullable: true,
					},
					{
						Name:       "created_at",
						DataType:   &sqlporter.DataType{Name: "DATETIME"},
						Default:    "CURRENT_TIMESTAMP",
						IsNullable: true,
					},
					{
						Name:       "updated_at",
						DataType:   &sqlporter.DataType{Name: "DATETIME"},
						IsNullable: true,
					},
				},
				Constraints: []*sqlporter.Constraint{
					{
						Name:    "pk_users",
						Type:    "PRIMARY KEY",
						Columns: []string{"id"},
					},
					{
						Name:  "users_email_check",
						Type:  "CHECK",
						Check: "email REGEXP '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$'",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Table with foreign keys",
			sql: `CREATE TABLE orders (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				product_id INTEGER NOT NULL,
				quantity INTEGER NOT NULL DEFAULT 1,
				status TEXT NOT NULL DEFAULT 'pending',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT
			);`,
			expected: &sqlporter.Table{
				Name: "orders",
				Columns: []*sqlporter.Column{
					{
						Name:       "id",
						DataType:   &sqlporter.DataType{Name: "INTEGER"},
						IsNullable: false,
						Extra:      "auto_increment",
					},
					{
						Name:       "user_id",
						DataType:   &sqlporter.DataType{Name: "INTEGER"},
						IsNullable: false,
					},
					{
						Name:       "product_id",
						DataType:   &sqlporter.DataType{Name: "INTEGER"},
						IsNullable: false,
					},
					{
						Name:       "quantity",
						DataType:   &sqlporter.DataType{Name: "INTEGER"},
						IsNullable: false,
						Default:    "1",
					},
					{
						Name:       "status",
						DataType:   &sqlporter.DataType{Name: "TEXT"},
						IsNullable: false,
						Default:    "'pending'",
					},
					{
						Name:       "created_at",
						DataType:   &sqlporter.DataType{Name: "DATETIME"},
						Default:    "CURRENT_TIMESTAMP",
						IsNullable: true,
					},
				},
				Constraints: []*sqlporter.Constraint{
					{
						Name:    "pk_orders",
						Type:    "PRIMARY KEY",
						Columns: []string{"id"},
					},
					{
						Name:       "fk_orders_users",
						Type:       "FOREIGN KEY",
						Columns:    []string{"user_id"},
						RefTable:   "users",
						RefColumns: []string{"id"},
						OnDelete:   "CASCADE",
					},
					{
						Name:       "fk_orders_products",
						Type:       "FOREIGN KEY",
						Columns:    []string{"product_id"},
						RefTable:   "products",
						RefColumns: []string{"id"},
						OnDelete:   "RESTRICT",
					},
				},
			},
			wantErr: false,
		},
	}

	parser := &SQLiteParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseCreateTable(tt.sql)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSQLiteParser_ParseAlterTable(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sqlporter.Table
		wantErr  bool
	}{
		{
			name: "Add column",
			sql: `ALTER TABLE users 
				ADD COLUMN middle_name TEXT;`,
			expected: &sqlporter.Table{
				Name: "users",
				Columns: []*sqlporter.Column{
					{
						Name:       "middle_name",
						DataType:   &sqlporter.DataType{Name: "TEXT"},
						IsNullable: true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Drop column",
			sql: `ALTER TABLE users 
				DROP COLUMN middle_name;`,
			expected: &sqlporter.Table{
				Name: "users",
			},
			wantErr: false,
		},
		{
			name: "Rename table",
			sql: `ALTER TABLE users 
				RENAME TO new_users;`,
			expected: &sqlporter.Table{
				Name: "new_users",
			},
			wantErr: false,
		},
	}

	parser := &SQLiteParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseAlterTable(tt.sql)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSQLiteParser_ParseDropTable(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sqlporter.Table
		wantErr  bool
	}{
		{
			name: "Drop table",
			sql:  "DROP TABLE users;",
			expected: &sqlporter.Table{
				Name: "users",
			},
			wantErr: false,
		},
		{
			name: "Drop table if exists",
			sql:  "DROP TABLE IF EXISTS users;",
			expected: &sqlporter.Table{
				Name: "users",
			},
			wantErr: false,
		},
	}

	parser := &SQLiteParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseDropTable(tt.sql)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSQLiteParser_ParseCreateIndex(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sqlporter.Index
		wantErr  bool
	}{
		{
			name: "Create index",
			sql:  "CREATE INDEX idx_users_email ON users(email);",
			expected: &sqlporter.Index{
				Name:    "idx_users_email",
				Columns: []string{"email"},
			},
			wantErr: false,
		},
		{
			name: "Create unique index",
			sql:  "CREATE UNIQUE INDEX idx_users_username ON users(username);",
			expected: &sqlporter.Index{
				Name:    "idx_users_username",
				Columns: []string{"username"},
				Unique:  true,
			},
			wantErr: false,
		},
		{
			name: "Create index with multiple columns",
			sql:  "CREATE INDEX idx_users_name ON users(first_name, last_name);",
			expected: &sqlporter.Index{
				Name:    "idx_users_name",
				Columns: []string{"first_name", "last_name"},
			},
			wantErr: false,
		},
	}

	parser := &SQLiteParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseCreateIndex(tt.sql)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSQLiteParser_ParseDropIndex(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sqlporter.Index
		wantErr  bool
	}{
		{
			name: "Drop index",
			sql:  "DROP INDEX idx_users_email;",
			expected: &sqlporter.Index{
				Name: "idx_users_email",
			},
			wantErr: false,
		},
		{
			name: "Drop index if exists",
			sql:  "DROP INDEX IF EXISTS idx_users_email;",
			expected: &sqlporter.Index{
				Name: "idx_users_email",
			},
			wantErr: false,
		},
	}

	parser := &SQLiteParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseDropIndex(tt.sql)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
