package parser

// Oracle data type mapping
var oracleDataTypeMap = map[string]string{
	"bigint":           "NUMBER(19)",
	"binary":           "RAW",
	"bit":              "NUMBER(1)",
	"blob":             "BLOB",
	"bool":             "NUMBER(1)",
	"boolean":          "NUMBER(1)",
	"char":             "CHAR",
	"date":             "DATE",
	"datetime":         "TIMESTAMP",
	"decimal":          "NUMBER",
	"double":           "BINARY_DOUBLE",
	"double precision": "BINARY_DOUBLE",
	"float":            "BINARY_FLOAT",
	"int":              "NUMBER(10)",
	"integer":          "NUMBER(10)",
	"json":             "CLOB",
	"longblob":         "BLOB",
	"longtext":         "CLOB",
	"mediumblob":       "BLOB",
	"mediumint":        "NUMBER(7)",
	"mediumtext":       "CLOB",
	"numeric":          "NUMBER",
	"real":             "BINARY_FLOAT",
	"set":              "VARCHAR2(4000)",
	"smallint":         "NUMBER(5)",
	"text":             "CLOB",
	"time":             "TIMESTAMP",
	"timestamp":        "TIMESTAMP",
	"tinyblob":         "BLOB",
	"tinyint":          "NUMBER(3)",
	"tinytext":         "CLOB",
	"varbinary":        "RAW",
	"varchar":          "VARCHAR2",
	"year":             "NUMBER(4)",
}

// Oracle default length mapping
var oracleDefaultLengthMap = map[string]int{
	"CHAR":     1,
	"RAW":      2000,
	"VARCHAR2": 4000,
}

// Oracle default precision mapping
var oracleDefaultPrecisionMap = map[string]int{
	"NUMBER": 38,
}

// Oracle default scale mapping
var oracleDefaultScaleMap = map[string]int{
	"NUMBER": 0,
}

// Oracle reserved words
var oracleReservedWords = []string{
	"access", "add", "all", "alter", "and", "any", "as", "asc",
	"audit", "between", "by", "char", "check", "cluster", "column",
	"comment", "compress", "connect", "create", "current", "date",
	"decimal", "default", "delete", "desc", "distinct", "drop",
	"else", "exclusive", "exists", "file", "float", "for", "from",
	"grant", "group", "having", "identified", "immediate", "in",
	"increment", "index", "initial", "insert", "integer", "intersect",
	"into", "is", "level", "like", "lock", "long", "maxextents",
	"minus", "mlslabel", "mode", "modify", "noaudit", "nocompress",
	"not", "nowait", "null", "number", "of", "offline", "on",
	"online", "option", "or", "order", "pctfree", "prior",
	"privileges", "public", "raw", "rename", "resource", "revoke",
	"row", "rowid", "rownum", "rows", "select", "session", "set",
	"share", "size", "smallint", "start", "successful", "synonym",
	"sysdate", "table", "then", "to", "trigger", "uid", "union",
	"unique", "update", "user", "validate", "values", "varchar",
	"varchar2", "view", "whenever", "where", "with",
}
