package postgres

// PostgreSQL data type mapping
var postgresDataTypeMap = map[string]string{
	"bigint":           "BIGINT",
	"binary":           "BYTEA",
	"bit":              "BIT",
	"blob":             "BYTEA",
	"bool":             "BOOLEAN",
	"boolean":          "BOOLEAN",
	"char":             "CHAR",
	"date":             "DATE",
	"datetime":         "TIMESTAMP",
	"decimal":          "DECIMAL",
	"double":           "DOUBLE PRECISION",
	"double precision": "DOUBLE PRECISION",
	"float":            "REAL",
	"int":              "INTEGER",
	"integer":          "INTEGER",
	"json":             "JSON",
	"longblob":         "BYTEA",
	"longtext":         "TEXT",
	"mediumblob":       "BYTEA",
	"mediumint":        "INTEGER",
	"mediumtext":       "TEXT",
	"numeric":          "NUMERIC",
	"real":             "REAL",
	"set":              "TEXT",
	"smallint":         "SMALLINT",
	"text":             "TEXT",
	"time":             "TIME",
	"timestamp":        "TIMESTAMP",
	"tinyblob":         "BYTEA",
	"tinyint":          "SMALLINT",
	"tinytext":         "TEXT",
	"varbinary":        "BYTEA",
	"varchar":          "VARCHAR",
	"year":             "SMALLINT",
}

// PostgreSQL default length mapping
var postgresDefaultLengthMap = map[string]int{
	"BIT":       1,
	"CHAR":      1,
	"VARCHAR":   255,
	"BYTEA":     0,
	"TEXT":      0,
	"SMALLINT":  0,
	"INTEGER":   0,
	"BIGINT":    0,
	"DECIMAL":   0,
	"NUMERIC":   0,
	"REAL":      0,
	"DOUBLE":    0,
	"DATE":      0,
	"TIME":      0,
	"TIMESTAMP": 0,
}

// PostgreSQL default precision mapping
var postgresDefaultPrecisionMap = map[string]int{
	"DECIMAL": 10,
	"NUMERIC": 10,
	"REAL":    0,
	"DOUBLE":  0,
}

// PostgreSQL default scale mapping
var postgresDefaultScaleMap = map[string]int{
	"DECIMAL": 0,
	"NUMERIC": 0,
	"REAL":    0,
	"DOUBLE":  0,
}

// PostgreSQL reserved words
var postgresReservedWords = []string{
	"all", "analyse", "analyze", "and", "any", "array", "as", "asc",
	"asymmetric", "authorization", "between", "binary", "both", "case",
	"cast", "check", "collate", "column", "constraint", "create",
	"cross", "current_catalog", "current_date", "current_role",
	"current_schema", "current_time", "current_timestamp", "current_user",
	"default", "deferrable", "desc", "distinct", "do", "else", "end",
	"except", "false", "fetch", "for", "foreign", "freeze", "from",
	"full", "grant", "group", "having", "ilike", "in", "initially",
	"inner", "intersect", "into", "is", "isnull", "join", "lateral",
	"leading", "left", "like", "limit", "localtime", "localtimestamp",
	"natural", "not", "notnull", "null", "offset", "on", "only", "or",
	"order", "outer", "overlaps", "placing", "primary", "references",
	"returning", "right", "select", "session_user", "similar", "some",
	"symmetric", "table", "then", "to", "trailing", "true", "union",
	"unique", "user", "using", "variadic", "verbose", "when", "where",
	"window", "with",
}
