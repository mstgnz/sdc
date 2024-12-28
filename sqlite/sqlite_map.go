package sqlite

// Data type conversion maps from SQLite to other database types
var (
	// SQLiteToMySQL Data type conversions from SQLite to MySQL
	SQLiteToMySQL = map[string]string{
		"INTEGER":  "int",
		"REAL":     "double",
		"TEXT":     "text",
		"BLOB":     "blob",
		"NUMERIC":  "decimal",
		"BOOLEAN":  "boolean",
		"DATETIME": "datetime",
		"DATE":     "date",
		"TIME":     "time",
	}

	// SQLiteToPostgreSQL Data type conversions from SQLite to PostgreSQL
	SQLiteToPostgreSQL = map[string]string{
		"INTEGER":  "integer",
		"REAL":     "double precision",
		"TEXT":     "text",
		"BLOB":     "bytea",
		"NUMERIC":  "numeric",
		"BOOLEAN":  "boolean",
		"DATETIME": "timestamp",
		"DATE":     "date",
		"TIME":     "time",
	}

	// SQLiteToSQLServer Data type conversions from SQLite to SQL Server
	SQLiteToSQLServer = map[string]string{
		"INTEGER":  "int",
		"REAL":     "float",
		"TEXT":     "nvarchar(max)",
		"BLOB":     "varbinary(max)",
		"NUMERIC":  "decimal",
		"BOOLEAN":  "bit",
		"DATETIME": "datetime2",
		"DATE":     "date",
		"TIME":     "time",
	}

	// SQLiteToOracle Data type conversions from SQLite to Oracle
	SQLiteToOracle = map[string]string{
		"INTEGER":  "NUMBER(10)",
		"REAL":     "BINARY_DOUBLE",
		"TEXT":     "CLOB",
		"BLOB":     "BLOB",
		"NUMERIC":  "NUMBER",
		"BOOLEAN":  "NUMBER(1)",
		"DATETIME": "TIMESTAMP",
		"DATE":     "DATE",
		"TIME":     "TIMESTAMP",
	}
)
