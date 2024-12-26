package parser

import (
	"reflect"
	"testing"

	"github.com/mstgnz/sqlporter"
)

func TestSQLServerParser_ParseCreateTable(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		want    *sqlporter.Table
		wantErr bool
	}{
		{
			name: "Table with all features",
			sql: `CREATE TABLE [dbo].[Users] (
				[Id] INT IDENTITY(1,1) PRIMARY KEY,
				[Username] NVARCHAR(50) NOT NULL UNIQUE,
				[Email] NVARCHAR(100) NOT NULL,
				[CreatedAt] DATETIME2 DEFAULT GETDATE(),
				[Status] TINYINT DEFAULT 1,
				CONSTRAINT FK_Users_Roles FOREIGN KEY ([RoleId]) REFERENCES [Roles]([Id])
			) ON [PRIMARY]`,
			want: &sqlporter.Table{
				Schema:    "dbo",
				Name:      "Users",
				FileGroup: "PRIMARY",
				Columns: []*sqlporter.Column{
					{
						Name:         "Id",
						DataType:     &sqlporter.DataType{Name: "INT"},
						Identity:     true,
						IdentitySeed: 1,
						IdentityIncr: 1,
						PrimaryKey:   true,
						IsNullable:   false,
						Nullable:     false,
					},
					{
						Name:       "Username",
						DataType:   &sqlporter.DataType{Name: "NVARCHAR", Length: 50},
						IsNullable: false,
						Nullable:   false,
						Unique:     true,
					},
					{
						Name:       "Email",
						DataType:   &sqlporter.DataType{Name: "NVARCHAR", Length: 100},
						IsNullable: false,
						Nullable:   false,
					},
					{
						Name:     "CreatedAt",
						DataType: &sqlporter.DataType{Name: "DATETIME2"},
						Default:  "GETDATE()",
					},
					{
						Name:     "Status",
						DataType: &sqlporter.DataType{Name: "TINYINT"},
						Default:  "1",
					},
				},
				Constraints: []*sqlporter.Constraint{
					{
						Name:       "FK_Users_Roles",
						Type:       "FOREIGN KEY",
						Columns:    []string{"RoleId"},
						RefTable:   "Roles",
						RefColumns: []string{"Id"},
					},
				},
			},
			wantErr: false,
		},
	}

	parser := &SQLServerParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.parseCreateTable(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLServerParser.parseCreateTable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLServerParser.parseCreateTable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLServerParser_ParseAlterTable(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		want    *sqlporter.AlterTable
		wantErr bool
	}{
		{
			name: "Add column",
			sql:  "ALTER TABLE [dbo].[Users] ADD [LastLoginDate] DATETIME2 NULL",
			want: &sqlporter.AlterTable{
				Schema: "dbo",
				Table:  "Users",
				Action: "ADD COLUMN",
				Column: &sqlporter.Column{
					Name:     "LastLoginDate",
					DataType: &sqlporter.DataType{Name: "DATETIME2"},
				},
			},
			wantErr: false,
		},
	}

	parser := &SQLServerParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.parseAlterTable(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLServerParser.parseAlterTable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLServerParser.parseAlterTable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLServerParser_ParseDropTable(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		want    *sqlporter.DropTable
		wantErr bool
	}{
		{
			name: "Drop table",
			sql:  "DROP TABLE [dbo].[Users]",
			want: &sqlporter.DropTable{
				Schema: "dbo",
				Table:  "Users",
			},
			wantErr: false,
		},
	}

	parser := &SQLServerParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.parseDropTable(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLServerParser.parseDropTable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLServerParser.parseDropTable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLServerParser_ParseCreateIndex(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		want    *sqlporter.Index
		wantErr bool
	}{
		{
			name: "Create index",
			sql:  "CREATE INDEX [IX_Users_Email] ON [dbo].[Users] ([Email])",
			want: &sqlporter.Index{
				Name:    "IX_Users_Email",
				Schema:  "dbo",
				Table:   "Users",
				Columns: []string{"Email"},
			},
			wantErr: false,
		},
	}

	parser := &SQLServerParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.parseCreateIndex(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLServerParser.parseCreateIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLServerParser.parseCreateIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLServerParser_ParseDropIndex(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		want    *sqlporter.DropIndex
		wantErr bool
	}{
		{
			name: "Drop index",
			sql:  "DROP INDEX [IX_Users_Email] ON [dbo].[Users]",
			want: &sqlporter.DropIndex{
				Schema: "dbo",
				Table:  "Users",
				Index:  "IX_Users_Email",
			},
			wantErr: false,
		},
	}

	parser := &SQLServerParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.parseDropIndex(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLServerParser.parseDropIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLServerParser.parseDropIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}
