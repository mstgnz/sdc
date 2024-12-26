package parser

import (
	"testing"

	"github.com/mstgnz/sdc"
	"github.com/stretchr/testify/assert"
)

func TestSQLServerParser_ParseCreateTable(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sdc.Table
		wantErr  bool
	}{
		{
			name: "Table with all features",
			sql: `CREATE TABLE [dbo].[users] (
				[id] INT IDENTITY(1,1) PRIMARY KEY,
				[username] NVARCHAR(50) NOT NULL UNIQUE,
				[email] NVARCHAR(100) NOT NULL UNIQUE,
				[password] NVARCHAR(100) NOT NULL,
				[full_name] NVARCHAR(100),
				[age] INT CHECK (age >= 18),
				[created_at] DATETIME2 DEFAULT GETDATE(),
				[updated_at] DATETIME2,
				CONSTRAINT [users_email_check] CHECK (email LIKE '%@%.%')
			) ON [PRIMARY];`,
			expected: &sdc.Table{
				Name:   "users",
				Schema: "dbo",
				Columns: []*sdc.Column{
					{
						Name:          "id",
						DataType:      &sdc.DataType{Name: "INT"},
						PrimaryKey:    true,
						AutoIncrement: true,
						Identity:      true,
						IdentitySeed:  1,
						IdentityIncr:  1,
					},
					{
						Name:     "username",
						DataType: &sdc.DataType{Name: "NVARCHAR", Length: 50},
						Nullable: false,
						Unique:   true,
					},
					{
						Name:     "email",
						DataType: &sdc.DataType{Name: "NVARCHAR", Length: 100},
						Nullable: false,
						Unique:   true,
					},
					{
						Name:     "password",
						DataType: &sdc.DataType{Name: "NVARCHAR", Length: 100},
						Nullable: false,
					},
					{
						Name:     "full_name",
						DataType: &sdc.DataType{Name: "NVARCHAR", Length: 100},
						Nullable: true,
					},
					{
						Name:     "age",
						DataType: &sdc.DataType{Name: "INT"},
						Check:    "age >= 18",
					},
					{
						Name:     "created_at",
						DataType: &sdc.DataType{Name: "DATETIME2"},
						Default:  "GETDATE()",
					},
					{
						Name:     "updated_at",
						DataType: &sdc.DataType{Name: "DATETIME2"},
						Nullable: true,
					},
				},
				Constraints: []*sdc.Constraint{
					{
						Name:  "users_email_check",
						Type:  "CHECK",
						Check: "email LIKE '%@%.%'",
					},
				},
				FileGroup: "PRIMARY",
			},
			wantErr: false,
		},
		{
			name: "Table with foreign keys",
			sql: `CREATE TABLE [dbo].[orders] (
				[id] INT IDENTITY(1,1) PRIMARY KEY,
				[user_id] INT NOT NULL REFERENCES [users]([id]) ON DELETE CASCADE,
				[product_id] INT NOT NULL,
				[quantity] INT NOT NULL DEFAULT 1,
				[status] NVARCHAR(20) NOT NULL DEFAULT 'pending',
				[created_at] DATETIME2 DEFAULT GETDATE(),
				FOREIGN KEY ([product_id]) REFERENCES [products]([id]) ON DELETE NO ACTION
			) ON [PRIMARY];`,
			expected: &sdc.Table{
				Name:   "orders",
				Schema: "dbo",
				Columns: []*sdc.Column{
					{
						Name:          "id",
						DataType:      &sdc.DataType{Name: "INT"},
						PrimaryKey:    true,
						AutoIncrement: true,
						Identity:      true,
						IdentitySeed:  1,
						IdentityIncr:  1,
					},
					{
						Name:     "user_id",
						DataType: &sdc.DataType{Name: "INT"},
						Nullable: false,
						ForeignKey: &sdc.ForeignKey{
							RefTable:  "users",
							RefColumn: "id",
							OnDelete:  "CASCADE",
						},
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
						DataType: &sdc.DataType{Name: "NVARCHAR", Length: 20},
						Nullable: false,
						Default:  "'pending'",
					},
					{
						Name:     "created_at",
						DataType: &sdc.DataType{Name: "DATETIME2"},
						Default:  "GETDATE()",
					},
				},
				Constraints: []*sdc.Constraint{
					{
						Type:       "FOREIGN KEY",
						Columns:    []string{"product_id"},
						RefTable:   "products",
						RefColumns: []string{"id"},
						OnDelete:   "NO ACTION",
					},
				},
				FileGroup: "PRIMARY",
			},
			wantErr: false,
		},
	}

	parser := &SQLServerParser{}

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

func TestSQLServerParser_ParseAlterTable(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sdc.AlterTable
		wantErr  bool
	}{
		{
			name: "Add column",
			sql: `ALTER TABLE [dbo].[users] 
				ADD [middle_name] NVARCHAR(50);`,
			expected: &sdc.AlterTable{
				Schema: "dbo",
				Table:  "users",
				Action: "ADD COLUMN",
				Column: &sdc.Column{
					Name:     "middle_name",
					DataType: &sdc.DataType{Name: "NVARCHAR", Length: 50},
				},
			},
			wantErr: false,
		},
		{
			name: "Drop column",
			sql: `ALTER TABLE [dbo].[users] 
				DROP COLUMN [middle_name];`,
			expected: &sdc.AlterTable{
				Schema: "dbo",
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
			sql: `ALTER TABLE [dbo].[users] 
				ADD CONSTRAINT [users_age_check] CHECK (age >= 21);`,
			expected: &sdc.AlterTable{
				Schema: "dbo",
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

	parser := &SQLServerParser{}

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

func TestSQLServerParser_ParseDropTable(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sdc.DropTable
		wantErr  bool
	}{
		{
			name: "Drop table",
			sql:  "DROP TABLE [dbo].[users];",
			expected: &sdc.DropTable{
				Schema: "dbo",
				Table:  "users",
			},
			wantErr: false,
		},
		{
			name: "Drop table if exists",
			sql:  "IF OBJECT_ID('[dbo].[users]', 'U') IS NOT NULL DROP TABLE [dbo].[users];",
			expected: &sdc.DropTable{
				Schema:   "dbo",
				Table:    "users",
				IfExists: true,
			},
			wantErr: false,
		},
	}

	parser := &SQLServerParser{}

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

func TestSQLServerParser_ParseCreateIndex(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sdc.Index
		wantErr  bool
	}{
		{
			name: "Create index",
			sql:  "CREATE INDEX [idx_users_email] ON [dbo].[users]([email]);",
			expected: &sdc.Index{
				Name:    "idx_users_email",
				Schema:  "dbo",
				Table:   "users",
				Columns: []string{"email"},
			},
			wantErr: false,
		},
		{
			name: "Create unique index",
			sql:  "CREATE UNIQUE INDEX [idx_users_username] ON [dbo].[users]([username]);",
			expected: &sdc.Index{
				Name:    "idx_users_username",
				Schema:  "dbo",
				Table:   "users",
				Columns: []string{"username"},
				Unique:  true,
			},
			wantErr: false,
		},
		{
			name: "Create clustered index",
			sql:  "CREATE CLUSTERED INDEX [idx_users_name] ON [dbo].[users]([first_name], [last_name]);",
			expected: &sdc.Index{
				Name:      "idx_users_name",
				Schema:    "dbo",
				Table:     "users",
				Columns:   []string{"first_name", "last_name"},
				Clustered: true,
			},
			wantErr: false,
		},
	}

	parser := &SQLServerParser{}

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

func TestSQLServerParser_ParseDropIndex(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected *sdc.DropIndex
		wantErr  bool
	}{
		{
			name: "Drop index",
			sql:  "DROP INDEX [idx_users_email] ON [dbo].[users];",
			expected: &sdc.DropIndex{
				Schema: "dbo",
				Table:  "users",
				Index:  "idx_users_email",
			},
			wantErr: false,
		},
		{
			name: "Drop index if exists",
			sql:  "IF EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_users_email') DROP INDEX [idx_users_email] ON [dbo].[users];",
			expected: &sdc.DropIndex{
				Schema:   "dbo",
				Table:    "users",
				Index:    "idx_users_email",
				IfExists: true,
			},
			wantErr: false,
		},
	}

	parser := &SQLServerParser{}

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
