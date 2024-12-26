package parser

// SQL Server data type mapping
var sqlserverDataTypeMap = map[string]string{
	"bigint":           "BIGINT",
	"binary":           "BINARY",
	"bit":              "BIT",
	"blob":             "VARBINARY(MAX)",
	"bool":             "BIT",
	"boolean":          "BIT",
	"char":             "CHAR",
	"date":             "DATE",
	"datetime":         "DATETIME2",
	"decimal":          "DECIMAL",
	"double":           "FLOAT",
	"double precision": "FLOAT",
	"float":            "REAL",
	"int":              "INT",
	"integer":          "INT",
	"json":             "NVARCHAR(MAX)",
	"longblob":         "VARBINARY(MAX)",
	"longtext":         "NVARCHAR(MAX)",
	"mediumblob":       "VARBINARY(MAX)",
	"mediumint":        "INT",
	"mediumtext":       "NVARCHAR(MAX)",
	"numeric":          "NUMERIC",
	"real":             "REAL",
	"set":              "NVARCHAR(MAX)",
	"smallint":         "SMALLINT",
	"text":             "NVARCHAR(MAX)",
	"time":             "TIME",
	"timestamp":        "DATETIME2",
	"tinyblob":         "VARBINARY(MAX)",
	"tinyint":          "TINYINT",
	"tinytext":         "NVARCHAR(MAX)",
	"varbinary":        "VARBINARY",
	"varchar":          "VARCHAR",
	"year":             "SMALLINT",
}

// SQL Server default length mapping
var sqlserverDefaultLengthMap = map[string]int{
	"BINARY":    1,
	"CHAR":      1,
	"NCHAR":     1,
	"NVARCHAR":  255,
	"VARBINARY": 1,
	"VARCHAR":   255,
}

// SQL Server default precision mapping
var sqlserverDefaultPrecisionMap = map[string]int{
	"DECIMAL": 18,
	"NUMERIC": 18,
	"FLOAT":   53,
	"REAL":    24,
}

// SQL Server default scale mapping
var sqlserverDefaultScaleMap = map[string]int{
	"DECIMAL": 0,
	"NUMERIC": 0,
	"FLOAT":   0,
	"REAL":    0,
}

// SQL Server reserved words
var sqlserverReservedWords = []string{
	"add", "all", "alter", "and", "any", "as", "asc", "authorization",
	"backup", "begin", "between", "break", "browse", "bulk", "by",
	"cascade", "case", "check", "checkpoint", "close", "clustered",
	"coalesce", "collate", "column", "commit", "compute", "constraint",
	"contains", "containstable", "continue", "convert", "create", "cross",
	"current", "current_date", "current_time", "current_timestamp",
	"current_user", "cursor", "database", "dbcc", "deallocate",
	"declare", "default", "delete", "deny", "desc", "disk", "distinct",
	"distributed", "double", "drop", "dump", "else", "end", "errlvl",
	"escape", "except", "exec", "execute", "exists", "exit", "external",
	"fetch", "file", "fillfactor", "for", "foreign", "freetext",
	"freetexttable", "from", "full", "function", "goto", "grant",
	"group", "having", "holdlock", "identity", "identity_insert",
	"identitycol", "if", "in", "index", "inner", "insert", "intersect",
	"into", "is", "join", "key", "kill", "left", "like", "lineno",
	"load", "merge", "national", "nocheck", "nonclustered", "not",
	"null", "nullif", "of", "off", "offsets", "on", "open",
	"opendatasource", "openquery", "openrowset", "openxml", "option",
	"or", "order", "outer", "over", "percent", "pivot", "plan",
	"precision", "primary", "print", "proc", "procedure", "public",
	"raiserror", "read", "readtext", "reconfigure", "references",
	"replication", "restore", "restrict", "return", "revert", "revoke",
	"right", "rollback", "rowcount", "rowguidcol", "rule", "save",
	"schema", "securityaudit", "select", "semantickeyphrasetable",
	"semanticsimilaritydetailstable", "semanticsimilaritytable",
	"session_user", "set", "setuser", "shutdown", "some", "statistics",
	"system_user", "table", "tablesample", "textsize", "then", "to",
	"top", "tran", "transaction", "trigger", "truncate", "try_convert",
	"tsequal", "union", "unique", "unpivot", "update", "updatetext",
	"use", "user", "values", "varying", "view", "waitfor", "when",
	"where", "while", "with", "within", "writetext",
}
