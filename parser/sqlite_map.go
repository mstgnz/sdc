package parser

// SQLite data type mapping
var sqliteDataTypeMap = map[string]string{
	"bigint":           "INTEGER",
	"binary":           "BLOB",
	"bit":              "INTEGER",
	"blob":             "BLOB",
	"bool":             "INTEGER",
	"boolean":          "INTEGER",
	"char":             "TEXT",
	"date":             "TEXT",
	"datetime":         "TEXT",
	"decimal":          "NUMERIC",
	"double":           "REAL",
	"double precision": "REAL",
	"float":            "REAL",
	"int":              "INTEGER",
	"integer":          "INTEGER",
	"json":             "TEXT",
	"longblob":         "BLOB",
	"longtext":         "TEXT",
	"mediumblob":       "BLOB",
	"mediumint":        "INTEGER",
	"mediumtext":       "TEXT",
	"numeric":          "NUMERIC",
	"real":             "REAL",
	"set":              "TEXT",
	"smallint":         "INTEGER",
	"text":             "TEXT",
	"time":             "TEXT",
	"timestamp":        "TEXT",
	"tinyblob":         "BLOB",
	"tinyint":          "INTEGER",
	"tinytext":         "TEXT",
	"varbinary":        "BLOB",
	"varchar":          "TEXT",
	"year":             "INTEGER",
}

// SQLite default length mapping
var sqliteDefaultLengthMap = map[string]int{
	"TEXT":    0,
	"BLOB":    0,
	"INTEGER": 0,
	"REAL":    0,
	"NUMERIC": 0,
}

// SQLite default precision mapping
var sqliteDefaultPrecisionMap = map[string]int{
	"NUMERIC": 10,
	"REAL":    0,
}

// SQLite default scale mapping
var sqliteDefaultScaleMap = map[string]int{
	"NUMERIC": 0,
	"REAL":    0,
}

// SQLite reserved words
var sqliteReservedWords = []string{
	"abort", "action", "add", "after", "all", "alter", "analyze", "and",
	"as", "asc", "attach", "autoincrement", "before", "begin", "between",
	"by", "cascade", "case", "cast", "check", "collate", "column",
	"commit", "conflict", "constraint", "create", "cross", "current_date",
	"current_time", "current_timestamp", "database", "default",
	"deferrable", "deferred", "delete", "desc", "detach", "distinct",
	"drop", "each", "else", "end", "escape", "except", "exclusive",
	"exists", "explain", "fail", "for", "foreign", "from", "full",
	"glob", "group", "having", "if", "ignore", "immediate", "in",
	"index", "indexed", "initially", "inner", "insert", "instead",
	"intersect", "into", "is", "isnull", "join", "key", "left",
	"like", "limit", "match", "natural", "no", "not", "notnull",
	"null", "of", "offset", "on", "or", "order", "outer", "plan",
	"pragma", "primary", "query", "raise", "recursive", "references",
	"regexp", "reindex", "release", "rename", "replace", "restrict",
	"right", "rollback", "row", "savepoint", "select", "set", "table",
	"temp", "temporary", "then", "to", "transaction", "trigger", "union",
	"unique", "update", "using", "vacuum", "values", "view", "virtual",
	"when", "where", "with", "without",
}
