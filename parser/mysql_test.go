package parser

import (
	"testing"

	"github.com/mstgnz/sdc"
	"github.com/stretchr/testify/assert"
)

func TestMySQLParser_ParseCreateTable(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sdc.Table
		wantErr  bool
	}{
		{
			name: "Table with all features",
			sql: `CREATE TABLE users (
				id INT AUTO_INCREMENT PRIMARY KEY,
				username VARCHAR(50) NOT NULL UNIQUE,
				email VARCHAR(100) NOT NULL UNIQUE,
				password VARCHAR(100) NOT NULL,
				full_name VARCHAR(100),
				age INT CHECK (age >= 18),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP,
				CONSTRAINT users_email_check CHECK (email REGEXP '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
			expected: &sdc.Table{
				Name: "users",
				Columns: []*sdc.Column{
					{
						Name:          "id",
						DataType:      &sdc.DataType{Name: "INT"},
						PrimaryKey:    true,
						AutoIncrement: true,
					},
					{
						Name:     "username",
						DataType: &sdc.DataType{Name: "VARCHAR", Length: 50},
						Nullable: false,
						Unique:   true,
					},
					{
						Name:     "email",
						DataType: &sdc.DataType{Name: "VARCHAR", Length: 100},
						Nullable: false,
						Unique:   true,
					},
					{
						Name:     "password",
						DataType: &sdc.DataType{Name: "VARCHAR", Length: 100},
						Nullable: false,
					},
					{
						Name:     "full_name",
						DataType: &sdc.DataType{Name: "VARCHAR", Length: 100},
						Nullable: true,
					},
					{
						Name:     "age",
						DataType: &sdc.DataType{Name: "INT"},
						Check:    "age >= 18",
					},
					{
						Name:     "created_at",
						DataType: &sdc.DataType{Name: "TIMESTAMP"},
						Default:  "CURRENT_TIMESTAMP",
					},
					{
						Name:     "updated_at",
						DataType: &sdc.DataType{Name: "TIMESTAMP"},
						Nullable: true,
					},
				},
				Constraints: []*sdc.Constraint{
					{
						Name:  "users_email_check",
						Type:  "CHECK",
						Check: "email REGEXP '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$'",
					},
				},
				Options: map[string]string{
					"ENGINE":  "InnoDB",
					"CHARSET": "utf8mb4",
					"COLLATE": "utf8mb4_unicode_ci",
				},
			},
			wantErr: false,
		},
		{
			name: "Table with foreign keys",
			sql: `CREATE TABLE orders (
				id INT AUTO_INCREMENT PRIMARY KEY,
				user_id INT NOT NULL,
				product_id INT NOT NULL,
				quantity INT NOT NULL DEFAULT 1,
				status VARCHAR(20) NOT NULL DEFAULT 'pending',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
				FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT
			) ENGINE=InnoDB;`,
			expected: &sdc.Table{
				Name: "orders",
				Columns: []*sdc.Column{
					{
						Name:          "id",
						DataType:      &sdc.DataType{Name: "INT"},
						PrimaryKey:    true,
						AutoIncrement: true,
					},
					{
						Name:     "user_id",
						DataType: &sdc.DataType{Name: "INT"},
						Nullable: false,
					},
					{
						Name:     "product_id",
						DataType: &sdc.DataType{Name: "INT"},
						Nullable: false,
					},
					{
						Name:     "quantity",
						DataType: &sdc.DataType{Name: "INT"},
						Nullable: false,
						Default:  "1",
					},
					{
						Name:     "status",
						DataType: &sdc.DataType{Name: "VARCHAR", Length: 20},
						Nullable: false,
						Default:  "'pending'",
					},
					{
						Name:     "created_at",
						DataType: &sdc.DataType{Name: "TIMESTAMP"},
						Default:  "CURRENT_TIMESTAMP",
					},
				},
				Constraints: []*sdc.Constraint{
					{
						Type:       "FOREIGN KEY",
						Columns:    []string{"user_id"},
						RefTable:   "users",
						RefColumns: []string{"id"},
						OnDelete:   "CASCADE",
					},
					{
						Type:       "FOREIGN KEY",
						Columns:    []string{"product_id"},
						RefTable:   "products",
						RefColumns: []string{"id"},
						OnDelete:   "RESTRICT",
					},
				},
				Options: map[string]string{
					"ENGINE": "InnoDB",
				},
			},
			wantErr: false,
		},
	}

	parser := &MySQLParser{}

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

func TestMySQLParser_ParseAlterTable(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sdc.AlterTable
		wantErr  bool
	}{
		{
			name: "Add column",
			sql: `ALTER TABLE users 
				ADD COLUMN middle_name VARCHAR(50);`,
			expected: &sdc.AlterTable{
				Table:  "users",
				Action: "ADD COLUMN",
				Column: &sdc.Column{
					Name:     "middle_name",
					DataType: &sdc.DataType{Name: "VARCHAR", Length: 50},
				},
			},
			wantErr: false,
		},
		{
			name: "Drop column",
			sql: `ALTER TABLE users 
				DROP COLUMN middle_name;`,
			expected: &sdc.AlterTable{
				Table:  "users",
				Action: "DROP COLUMN",
				Column: &sdc.Column{
					Name: "middle_name",
				},
			},
			wantErr: false,
		},
		{
			name: "Add constraint",
			sql: `ALTER TABLE users 
				ADD CONSTRAINT users_age_check CHECK (age >= 21);`,
			expected: &sdc.AlterTable{
				Table:  "users",
				Action: "ADD CONSTRAINT",
				Constraint: &sdc.Constraint{
					Name:  "users_age_check",
					Type:  "CHECK",
					Check: "age >= 21",
				},
			},
			wantErr: false,
		},
	}

	parser := &MySQLParser{}

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

func TestMySQLParser_ParseDropTable(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sdc.DropTable
		wantErr  bool
	}{
		{
			name: "Drop table",
			sql:  "DROP TABLE users;",
			expected: &sdc.DropTable{
				Table: "users",
			},
			wantErr: false,
		},
		{
			name: "Drop table if exists",
			sql:  "DROP TABLE IF EXISTS users;",
			expected: &sdc.DropTable{
				Table:    "users",
				IfExists: true,
			},
			wantErr: false,
		},
		{
			name: "Drop table cascade",
			sql:  "DROP TABLE users CASCADE;",
			expected: &sdc.DropTable{
				Table:   "users",
				Cascade: true,
			},
			wantErr: false,
		},
	}

	parser := &MySQLParser{}

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

func TestMySQLParser_ParseCreateIndex(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sdc.Index
		wantErr  bool
	}{
		{
			name: "Create index",
			sql:  "CREATE INDEX idx_users_email ON users(email);",
			expected: &sdc.Index{
				Name:    "idx_users_email",
				Table:   "users",
				Columns: []string{"email"},
			},
			wantErr: false,
		},
		{
			name: "Create unique index",
			sql:  "CREATE UNIQUE INDEX idx_users_username ON users(username);",
			expected: &sdc.Index{
				Name:    "idx_users_username",
				Table:   "users",
				Columns: []string{"username"},
				Unique:  true,
			},
			wantErr: false,
		},
		{
			name: "Create index with include",
			sql:  "CREATE INDEX idx_users_name ON users(first_name, last_name) USING BTREE;",
			expected: &sdc.Index{
				Name:    "idx_users_name",
				Table:   "users",
				Columns: []string{"first_name", "last_name"},
				Options: map[string]string{
					"USING": "BTREE",
				},
			},
			wantErr: false,
		},
	}

	parser := &MySQLParser{}

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

func TestMySQLParser_ParseDropIndex(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sdc.DropIndex
		wantErr  bool
	}{
		{
			name: "Drop index",
			sql:  "DROP INDEX idx_users_email ON users;",
			expected: &sdc.DropIndex{
				Table: "users",
				Index: "idx_users_email",
			},
			wantErr: false,
		},
		{
			name: "Drop index if exists",
			sql:  "DROP INDEX IF EXISTS idx_users_email ON users;",
			expected: &sdc.DropIndex{
				Table:    "users",
				Index:    "idx_users_email",
				IfExists: true,
			},
			wantErr: false,
		},
	}

	parser := &MySQLParser{}

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
