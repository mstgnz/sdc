package parser

import (
	"testing"

	"github.com/mstgnz/sqlporter"
	"github.com/stretchr/testify/assert"
)

func TestOracleParser_ParseCreateTable(t *testing.T) {
	p := &OracleParser{}

	tests := []struct {
		name     string
		input    string
		expected *sqlporter.Table
		wantErr  bool
	}{
		{
			name: "Basic table with primary key",
			input: `CREATE TABLE "USERS" (
				"ID" NUMBER(10) GENERATED ALWAYS AS IDENTITY,
				"NAME" VARCHAR2(100) NOT NULL,
				"EMAIL" VARCHAR2(255),
				"CREATED_AT" TIMESTAMP DEFAULT SYSTIMESTAMP,
				CONSTRAINT "PK_USERS" PRIMARY KEY ("ID") USING INDEX TABLESPACE "USERS_IDX"
			) TABLESPACE "USERS_DATA"`,
			expected: &sqlporter.Table{
				Name: "USERS",
				Columns: []*sqlporter.Column{
					{
						Name: "ID",
						DataType: &sqlporter.DataType{
							Name:      "NUMBER",
							Precision: 10,
						},
						IsNullable: false,
						Extra:      "identity",
					},
					{
						Name: "NAME",
						DataType: &sqlporter.DataType{
							Name:   "VARCHAR2",
							Length: 100,
						},
						IsNullable: false,
					},
					{
						Name: "EMAIL",
						DataType: &sqlporter.DataType{
							Name:   "VARCHAR2",
							Length: 255,
						},
						IsNullable: true,
					},
					{
						Name: "CREATED_AT",
						DataType: &sqlporter.DataType{
							Name: "TIMESTAMP",
						},
						IsNullable: true,
						Default:    "SYSTIMESTAMP",
					},
				},
				Constraints: []*sqlporter.Constraint{
					{
						Name:    "PK_USERS",
						Type:    "PRIMARY KEY",
						Columns: []string{"ID"},
					},
				},
				TableSpace: "USERS_DATA",
				Indexes: []*sqlporter.Index{
					{
						Name:      "PK_USERS",
						Columns:   []string{"ID"},
						Unique:    true,
						FileGroup: "USERS_IDX",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Table with foreign key and unique constraint",
			input: `CREATE TABLE "ORDERS" (
				"ORDER_ID" NUMBER(10) GENERATED ALWAYS AS IDENTITY,
				"USER_ID" NUMBER(10) NOT NULL,
				"ORDER_NUMBER" VARCHAR2(50) NOT NULL,
				"TOTAL" NUMBER(10,2) DEFAULT 0.00,
				CONSTRAINT "PK_ORDERS" PRIMARY KEY ("ORDER_ID"),
				CONSTRAINT "UK_ORDER_NUMBER" UNIQUE ("ORDER_NUMBER") USING INDEX TABLESPACE "ORDERS_IDX",
				CONSTRAINT "FK_ORDERS_USERS" FOREIGN KEY ("USER_ID") REFERENCES "USERS" ("ID") ON DELETE CASCADE
			) TABLESPACE "ORDERS_DATA"
			STORAGE (INITIAL 1M NEXT 1M)`,
			expected: &sqlporter.Table{
				Name: "ORDERS",
				Columns: []*sqlporter.Column{
					{
						Name: "ORDER_ID",
						DataType: &sqlporter.DataType{
							Name:      "NUMBER",
							Precision: 10,
						},
						IsNullable: false,
						Extra:      "identity",
					},
					{
						Name: "USER_ID",
						DataType: &sqlporter.DataType{
							Name:      "NUMBER",
							Precision: 10,
						},
						IsNullable: false,
					},
					{
						Name: "ORDER_NUMBER",
						DataType: &sqlporter.DataType{
							Name:   "VARCHAR2",
							Length: 50,
						},
						IsNullable: false,
					},
					{
						Name: "TOTAL",
						DataType: &sqlporter.DataType{
							Name:      "NUMBER",
							Precision: 10,
							Scale:     2,
						},
						IsNullable: true,
						Default:    "0.00",
					},
				},
				Constraints: []*sqlporter.Constraint{
					{
						Name:    "PK_ORDERS",
						Type:    "PRIMARY KEY",
						Columns: []string{"ORDER_ID"},
					},
					{
						Name:    "UK_ORDER_NUMBER",
						Type:    "UNIQUE",
						Columns: []string{"ORDER_NUMBER"},
					},
					{
						Name:       "FK_ORDERS_USERS",
						Type:       "FOREIGN KEY",
						Columns:    []string{"USER_ID"},
						RefTable:   "USERS",
						RefColumns: []string{"ID"},
						OnDelete:   "CASCADE",
					},
				},
				TableSpace: "ORDERS_DATA",
				Indexes: []*sqlporter.Index{
					{
						Name:    "PK_ORDERS",
						Columns: []string{"ORDER_ID"},
						Unique:  true,
					},
					{
						Name:      "UK_ORDER_NUMBER",
						Columns:   []string{"ORDER_NUMBER"},
						Unique:    true,
						FileGroup: "ORDERS_IDX",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.parseCreateTable(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOracleParser_ConvertDataType(t *testing.T) {
	p := &OracleParser{}

	tests := []struct {
		name     string
		input    *sqlporter.DataType
		expected string
	}{
		{
			name: "VARCHAR2 with length",
			input: &sqlporter.DataType{
				Name:   "VARCHAR2",
				Length: 100,
			},
			expected: "VARCHAR2(100)",
		},
		{
			name: "VARCHAR2 without length",
			input: &sqlporter.DataType{
				Name: "VARCHAR2",
			},
			expected: "VARCHAR2(4000)",
		},
		{
			name: "NUMBER with precision and scale",
			input: &sqlporter.DataType{
				Name:      "NUMBER",
				Precision: 10,
				Scale:     2,
			},
			expected: "NUMBER(10,2)",
		},
		{
			name: "NUMBER with only precision",
			input: &sqlporter.DataType{
				Name:      "NUMBER",
				Precision: 10,
			},
			expected: "NUMBER(10)",
		},
		{
			name: "INTEGER to NUMBER",
			input: &sqlporter.DataType{
				Name: "INTEGER",
			},
			expected: "NUMBER(38)",
		},
		{
			name: "SMALLINT to NUMBER",
			input: &sqlporter.DataType{
				Name: "SMALLINT",
			},
			expected: "NUMBER(5)",
		},
		{
			name: "BOOLEAN to NUMBER",
			input: &sqlporter.DataType{
				Name: "BOOLEAN",
			},
			expected: "NUMBER(1)",
		},
		{
			name: "TEXT to CLOB",
			input: &sqlporter.DataType{
				Name: "TEXT",
			},
			expected: "CLOB",
		},
		{
			name: "NTEXT to NCLOB",
			input: &sqlporter.DataType{
				Name: "NTEXT",
			},
			expected: "NCLOB",
		},
		{
			name: "TIMESTAMP with scale",
			input: &sqlporter.DataType{
				Name:  "TIMESTAMP",
				Scale: 6,
			},
			expected: "TIMESTAMP(6)",
		},
		{
			name: "FLOAT with precision",
			input: &sqlporter.DataType{
				Name:      "FLOAT",
				Precision: 126,
			},
			expected: "FLOAT(126)",
		},
		{
			name: "REAL to FLOAT",
			input: &sqlporter.DataType{
				Name: "REAL",
			},
			expected: "FLOAT(63)",
		},
		{
			name: "Unknown type defaults to VARCHAR2",
			input: &sqlporter.DataType{
				Name: "UNKNOWN_TYPE",
			},
			expected: "VARCHAR2(4000)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.convertDataType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
