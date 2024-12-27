package sqlite

// SQLite'dan diğer veritabanı tiplerine veri tipi dönüşüm haritaları
var (
	// SQLiteToMySQL SQLite'dan MySQL'e veri tipi dönüşümleri
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

	// SQLiteToPostgreSQL SQLite'dan PostgreSQL'e veri tipi dönüşümleri
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

	// SQLiteToSQLServer SQLite'dan SQL Server'a veri tipi dönüşümleri
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

	// SQLiteToOracle SQLite'dan Oracle'a veri tipi dönüşümleri
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
