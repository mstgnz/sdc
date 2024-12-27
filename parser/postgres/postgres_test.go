package postgres

import (
	"testing"

	"github.com/mstgnz/sqlporter"
	"github.com/stretchr/testify/assert"
)

func TestPostgresParser_ParseCreateTable(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sqlporter.Table
		wantErr  bool
	}{
		{
			name: "Table with all features",
			sql: `CREATE TABLE IF NOT EXISTS public.users (
				id SERIAL PRIMARY KEY,
				username VARCHAR(50) NOT NULL UNIQUE,
				email VARCHAR(100) NOT NULL UNIQUE,
				password VARCHAR(100) NOT NULL,
				full_name VARCHAR(100),
				age INTEGER CHECK (age >= 18),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP,
				CONSTRAINT users_email_check CHECK (email ~* '^[A-Za-z0-9\\._%+-]+@[A-Za-z0-9\\.-]+\\.[A-Za-z]{2,}$')
			) TABLESPACE users_space;`,
			expected: &sqlporter.Table{
				Name:   "users",
				Schema: "public",
				Columns: []*sqlporter.Column{
					{
						Name:          "id",
						DataType:      &sqlporter.DataType{Name: "SERIAL"},
						PrimaryKey:    true,
						AutoIncrement: true,
					},
					{
						Name:     "username",
						DataType: &sqlporter.DataType{Name: "VARCHAR", Length: 50},
						Nullable: false,
						Unique:   true,
					},
					{
						Name:     "email",
						DataType: &sqlporter.DataType{Name: "VARCHAR", Length: 100},
						Nullable: false,
						Unique:   true,
					},
					{
						Name:     "password",
						DataType: &sqlporter.DataType{Name: "VARCHAR", Length: 100},
						Nullable: false,
					},
					{
						Name:     "full_name",
						DataType: &sqlporter.DataType{Name: "VARCHAR", Length: 100},
						Nullable: true,
					},
					{
						Name:     "age",
						DataType: &sqlporter.DataType{Name: "INTEGER"},
						Check:    "age >= 18",
					},
					{
						Name:     "created_at",
						DataType: &sqlporter.DataType{Name: "TIMESTAMP"},
						Default:  "CURRENT_TIMESTAMP",
					},
					{
						Name:     "updated_at",
						DataType: &sqlporter.DataType{Name: "TIMESTAMP"},
						Nullable: true,
					},
				},
				Constraints: []*sqlporter.Constraint{
					{
						Name:  "users_email_check",
						Type:  "CHECK",
						Check: "email ~* '^[A-Za-z0-9\\._%+-]+@[A-Za-z0-9\\.-]+\\.[A-Za-z]{2,}$'",
					},
				},
				FileGroup: "users_space",
			},
			wantErr: false,
		},
		{
			name: "Table with foreign keys",
			sql: `CREATE TABLE public.orders (
				id SERIAL PRIMARY KEY,
				user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				product_id INTEGER NOT NULL,
				quantity INTEGER NOT NULL DEFAULT 1,
				status VARCHAR(20) NOT NULL DEFAULT 'pending',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT
			);`,
			expected: &sqlporter.Table{
				Name:   "orders",
				Schema: "public",
				Columns: []*sqlporter.Column{
					{
						Name:          "id",
						DataType:      &sqlporter.DataType{Name: "SERIAL"},
						PrimaryKey:    true,
						AutoIncrement: true,
					},
					{
						Name:     "user_id",
						DataType: &sqlporter.DataType{Name: "INTEGER"},
						Nullable: false,
						ForeignKey: &sqlporter.ForeignKey{
							RefTable:  "users",
							RefColumn: "id",
							OnDelete:  "CASCADE",
						},
					},
					{
						Name:     "product_id",
						DataType: &sqlporter.DataType{Name: "INTEGER"},
						Nullable: false,
					},
					{
						Name:     "quantity",
						DataType: &sqlporter.DataType{Name: "INTEGER"},
						Nullable: false,
						Default:  "1",
					},
					{
						Name:     "status",
						DataType: &sqlporter.DataType{Name: "VARCHAR", Length: 20},
						Nullable: false,
						Default:  "'pending'",
					},
					{
						Name:     "created_at",
						DataType: &sqlporter.DataType{Name: "TIMESTAMP"},
						Default:  "CURRENT_TIMESTAMP",
					},
				},
				Constraints: []*sqlporter.Constraint{
					{
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

	parser := &PostgresParser{}

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

func TestPostgresParser_ParseAlterTable(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sqlporter.AlterTable
		wantErr  bool
	}{
		{
			name: "Add column",
			sql: `ALTER TABLE public.users 
				ADD COLUMN middle_name VARCHAR(50);`,
			expected: &sqlporter.AlterTable{
				Schema: "public",
				Table:  "users",
				Action: "ADD COLUMN",
				Column: &sqlporter.Column{
					Name:       "middle_name",
					DataType:   &sqlporter.DataType{Name: "VARCHAR", Length: 50},
					Nullable:   true,
					IsNullable: true,
				},
			},
			wantErr: false,
		},
		{
			name: "Drop column",
			sql: `ALTER TABLE public.users 
				DROP COLUMN middle_name;`,
			expected: &sqlporter.AlterTable{
				Schema: "public",
				Table:  "users",
				Action: "DROP COLUMN",
				Column: &sqlporter.Column{
					Name: "middle_name",
				},
			},
			wantErr: false,
		},
		{
			name: "Add constraint",
			sql: `ALTER TABLE public.users 
				ADD CONSTRAINT users_age_check CHECK (age >= 21);`,
			expected: &sqlporter.AlterTable{
				Schema: "public",
				Table:  "users",
				Action: "ADD CONSTRAINT",
				Constraint: &sqlporter.Constraint{
					Name:  "users_age_check",
					Type:  "CHECK",
					Check: "age >= 21",
				},
			},
			wantErr: false,
		},
	}

	parser := &PostgresParser{}

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

func TestPostgresParser_ParseDropTable(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sqlporter.DropTable
		wantErr  bool
	}{
		{
			name: "Drop table",
			sql:  "DROP TABLE public.users;",
			expected: &sqlporter.DropTable{
				Schema: "public",
				Table:  "users",
			},
			wantErr: false,
		},
		{
			name: "Drop table if exists",
			sql:  "DROP TABLE IF EXISTS public.users;",
			expected: &sqlporter.DropTable{
				Schema:   "public",
				Table:    "users",
				IfExists: true,
			},
			wantErr: false,
		},
		{
			name: "Drop table cascade",
			sql:  "DROP TABLE public.users CASCADE;",
			expected: &sqlporter.DropTable{
				Schema:  "public",
				Table:   "users",
				Cascade: true,
			},
			wantErr: false,
		},
	}

	parser := &PostgresParser{}

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

func TestPostgresParser_ParseCreateIndex(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sqlporter.Index
		wantErr  bool
	}{
		{
			name: "Create index",
			sql:  "CREATE INDEX idx_users_email ON public.users(email);",
			expected: &sqlporter.Index{
				Name:    "idx_users_email",
				Schema:  "public",
				Table:   "users",
				Columns: []string{"email"},
			},
			wantErr: false,
		},
		{
			name: "Create unique index",
			sql:  "CREATE UNIQUE INDEX idx_users_username ON public.users(username);",
			expected: &sqlporter.Index{
				Name:    "idx_users_username",
				Schema:  "public",
				Table:   "users",
				Columns: []string{"username"},
				Unique:  true,
			},
			wantErr: false,
		},
		{
			name: "Create index with include",
			sql:  "CREATE INDEX idx_users_name ON public.users(first_name, last_name) INCLUDE (email);",
			expected: &sqlporter.Index{
				Name:           "idx_users_name",
				Schema:         "public",
				Table:          "users",
				Columns:        []string{"first_name", "last_name"},
				IncludeColumns: []string{"email"},
			},
			wantErr: false,
		},
	}

	parser := &PostgresParser{}

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

func TestPostgresParser_ParseDropIndex(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sqlporter.DropIndex
		wantErr  bool
	}{
		{
			name: "Drop index",
			sql:  "DROP INDEX public.idx_users_email;",
			expected: &sqlporter.DropIndex{
				Schema: "public",
				Index:  "idx_users_email",
			},
			wantErr: false,
		},
		{
			name: "Drop index if exists",
			sql:  "DROP INDEX IF EXISTS public.idx_users_email;",
			expected: &sqlporter.DropIndex{
				Schema:   "public",
				Index:    "idx_users_email",
				IfExists: true,
			},
			wantErr: false,
		},
		{
			name: "Drop index cascade",
			sql:  "DROP INDEX public.idx_users_email CASCADE;",
			expected: &sqlporter.DropIndex{
				Schema:  "public",
				Index:   "idx_users_email",
				Cascade: true,
			},
			wantErr: false,
		},
	}

	parser := &PostgresParser{}

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
