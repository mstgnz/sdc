package parser

import (
	"strconv"

	"github.com/mstgnz/sdc"
)

// Parser represents the interface for all database parsers
type Parser interface {
	// Parse converts SQL dump to Entity structure
	Parse(sql string) (*sdc.Entity, error)

	// Convert transforms Entity structure to target database format
	Convert(entity *sdc.Entity) (string, error)

	// ValidateSchema validates the schema structure
	ValidateSchema(entity *sdc.Entity) error

	// ParseCreateTable parses CREATE TABLE statement
	ParseCreateTable(sql string) (*sdc.Table, error)

	// ParseAlterTable parses ALTER TABLE statement
	ParseAlterTable(sql string) (*sdc.AlterTable, error)

	// ParseDropTable parses DROP TABLE statement
	ParseDropTable(sql string) (*sdc.DropTable, error)

	// ParseCreateIndex parses CREATE INDEX statement
	ParseCreateIndex(sql string) (*sdc.Index, error)

	// ParseDropIndex parses DROP INDEX statement
	ParseDropIndex(sql string) (*sdc.DropIndex, error)

	// ConvertDataType converts data type to target database format
	ConvertDataType(dataType *sdc.DataType) string

	// ConvertDataTypeFrom converts source database data type
	ConvertDataTypeFrom(sourceType string, length int, precision int, scale int) *sdc.DataType

	// GetDefaultSchema returns the default schema name
	GetDefaultSchema() string

	// GetSchemaPrefix returns the schema prefix
	GetSchemaPrefix(schema string) string

	// GetIdentifierQuote returns the identifier quote character
	GetIdentifierQuote() string

	// GetStringQuote returns the string quote character
	GetStringQuote() string

	// GetMaxIdentifierLength returns the maximum identifier length
	GetMaxIdentifierLength() int

	// GetReservedWords returns the reserved words
	GetReservedWords() []string

	// ValidateIdentifier validates the identifier name
	ValidateIdentifier(name string) error

	// EscapeIdentifier escapes the identifier name
	EscapeIdentifier(name string) string

	// EscapeString escapes the string value
	EscapeString(value string) string
}

// DatabaseType represents the type of database
type DatabaseType string

const (
	// MySQL database type
	MySQL DatabaseType = "mysql"
	// PostgreSQL database type
	PostgreSQL DatabaseType = "postgresql"
	// SQLServer database type
	SQLServer DatabaseType = "sqlserver"
	// Oracle database type
	Oracle DatabaseType = "oracle"
	// SQLite database type
	SQLite DatabaseType = "sqlite"
)

// DatabaseInfo contains database-specific information and limitations
type DatabaseInfo struct {
	Type                 DatabaseType
	DefaultSchema        string
	SchemaPrefix         string
	IdentifierQuote      string
	StringQuote          string
	MaxIdentifierLength  int
	ReservedWords        []string
	MaxTables            int // 0 means unlimited
	MaxColumns           int
	MaxIndexes           int // 0 means unlimited
	MaxForeignKeys       int // 0 means unlimited
	DefaultCharset       string
	DefaultCollation     string
	MaxPartitions        int // 0 means unlimited
	MaxTriggers          int // 0 means unlimited
	MaxViews             int // 0 means unlimited
	MaxStoredProcedures  int // 0 means unlimited
	MaxFunctions         int // 0 means unlimited
	MaxSequences         int // 0 means unlimited
	MaxMaterializedViews int // 0 means unlimited or not supported
}

// DatabaseInfoMap contains database information for each supported database type
var DatabaseInfoMap = map[DatabaseType]DatabaseInfo{
	MySQL: {
		Type:                 MySQL,
		DefaultSchema:        "",
		SchemaPrefix:         "",
		IdentifierQuote:      "`",
		StringQuote:          "'",
		MaxIdentifierLength:  64,
		ReservedWords:        []string{"accessible", "add", "all", "alter", "analyze", "and", "as", "asc", "asensitive", "before", "between", "bigint", "binary", "blob", "both", "by", "call", "cascade", "case", "change", "char", "character", "check", "collate", "column", "condition", "constraint", "continue", "convert", "create", "cross", "current_date", "current_time", "current_timestamp", "current_user", "cursor", "database", "databases", "day_hour", "day_microsecond", "day_minute", "day_second", "dec", "decimal", "declare", "default", "delayed", "delete", "desc", "describe", "deterministic", "distinct", "distinctrow", "div", "double", "drop", "dual", "each", "else", "elseif", "enclosed", "escaped", "exists", "exit", "explain", "false", "fetch", "float", "float4", "float8", "for", "force", "foreign", "from", "fulltext", "general", "grant", "group", "having", "high_priority", "hour_microsecond", "hour_minute", "hour_second", "if", "ignore", "in", "index", "infile", "inner", "inout", "insensitive", "insert", "int", "int1", "int2", "int3", "int4", "int8", "integer", "interval", "into", "is", "iterate", "join", "key", "keys", "kill", "leading", "leave", "left", "like", "limit", "linear", "lines", "load", "localtime", "localtimestamp", "lock", "long", "longblob", "longtext", "loop", "low_priority", "master_ssl_verify_server_cert", "match", "maxvalue", "mediumblob", "mediumint", "mediumtext", "middleint", "minute_microsecond", "minute_second", "mod", "modifies", "natural", "not", "no_write_to_binlog", "null", "numeric", "on", "optimize", "option", "optionally", "or", "order", "out", "outer", "outfile", "precision", "primary", "procedure", "purge", "range", "read", "reads", "read_write", "real", "references", "regexp", "release", "rename", "repeat", "replace", "require", "resignal", "restrict", "return", "revoke", "right", "rlike", "schema", "schemas", "second_microsecond", "select", "sensitive", "separator", "set", "show", "signal", "smallint", "spatial", "specific", "sql", "sqlexception", "sqlstate", "sqlwarning", "sql_big_result", "sql_calc_found_rows", "sql_small_result", "ssl", "starting", "straight_join", "table", "terminated", "then", "tinyblob", "tinyint", "tinytext", "to", "trailing", "trigger", "true", "undo", "union", "unique", "unlock", "unsigned", "update", "usage", "use", "using", "utc_date", "utc_time", "utc_timestamp", "values", "varbinary", "varchar", "varcharacter", "varying", "when", "where", "while", "with", "write", "xor", "year_month", "zerofill"},
		MaxTables:            4096,
		MaxColumns:           4096,
		MaxIndexes:           64,
		MaxForeignKeys:       64,
		DefaultCharset:       "utf8mb4",
		DefaultCollation:     "utf8mb4_unicode_ci",
		MaxPartitions:        8192,
		MaxTriggers:          0, // Unlimited
		MaxViews:             0, // Unlimited
		MaxStoredProcedures:  0, // Unlimited
		MaxFunctions:         0, // Unlimited
		MaxSequences:         0, // Not supported
		MaxMaterializedViews: 0, // Not supported
	},
	PostgreSQL: {
		Type:                 PostgreSQL,
		DefaultSchema:        "public",
		SchemaPrefix:         "",
		IdentifierQuote:      "\"",
		StringQuote:          "'",
		MaxIdentifierLength:  63,
		ReservedWords:        []string{"all", "analyse", "analyze", "and", "any", "array", "as", "asc", "asymmetric", "authorization", "binary", "both", "case", "cast", "check", "collate", "column", "constraint", "create", "cross", "current_date", "current_role", "current_time", "current_timestamp", "current_user", "default", "deferrable", "desc", "distinct", "do", "else", "end", "except", "false", "for", "foreign", "freeze", "from", "full", "grant", "group", "having", "ilike", "in", "initially", "inner", "intersect", "into", "is", "isnull", "join", "leading", "left", "like", "limit", "localtime", "localtimestamp", "natural", "new", "not", "notnull", "null", "off", "offset", "old", "on", "only", "or", "order", "outer", "overlaps", "placing", "primary", "references", "right", "select", "session_user", "similar", "some", "symmetric", "table", "then", "to", "trailing", "true", "union", "unique", "user", "using", "verbose", "when", "where", "with"},
		MaxTables:            0, // Unlimited
		MaxColumns:           1600,
		MaxIndexes:           0, // Unlimited
		MaxForeignKeys:       0, // Unlimited
		DefaultCharset:       "UTF8",
		DefaultCollation:     "en_US.UTF-8",
		MaxPartitions:        0, // Unlimited
		MaxTriggers:          0, // Unlimited
		MaxViews:             0, // Unlimited
		MaxStoredProcedures:  0, // Unlimited
		MaxFunctions:         0, // Unlimited
		MaxSequences:         0, // Unlimited
		MaxMaterializedViews: 0, // Unlimited
	},
	SQLServer: {
		Type:                 SQLServer,
		DefaultSchema:        "dbo",
		SchemaPrefix:         "",
		IdentifierQuote:      "\"",
		StringQuote:          "'",
		MaxIdentifierLength:  128,
		ReservedWords:        []string{"add", "all", "alter", "and", "any", "as", "asc", "authorization", "backup", "begin", "between", "break", "browse", "bulk", "by", "cascade", "case", "check", "checkpoint", "close", "clustered", "coalesce", "collate", "column", "commit", "compute", "constraint", "contains", "containstable", "continue", "convert", "create", "cross", "current", "current_date", "current_time", "current_timestamp", "current_user", "cursor", "database", "dbcc", "deallocate", "declare", "default", "delete", "deny", "desc", "disk", "distinct", "distributed", "double", "drop", "dump", "else", "end", "errlvl", "escape", "except", "exec", "execute", "exists", "exit", "external", "fetch", "file", "fillfactor", "for", "foreign", "freetext", "freetexttable", "from", "full", "function", "goto", "grant", "group", "having", "holdlock", "identity", "identity_insert", "identitycol", "if", "in", "index", "inner", "insert", "intersect", "into", "is", "join", "key", "kill", "left", "like", "lineno", "load", "merge", "national", "nocheck", "nonclustered", "not", "null", "nullif", "of", "off", "offsets", "on", "open", "opendatasource", "openquery", "openrowset", "openxml", "option", "or", "order", "outer", "over", "percent", "pivot", "plan", "precision", "primary", "print", "proc", "procedure", "public", "raiserror", "read", "readtext", "reconfigure", "references", "replication", "restore", "restrict", "return", "revert", "revoke", "right", "rollback", "rowcount", "rowguidcol", "rule", "save", "schema", "securityaudit", "select", "semantickeyphrasetable", "semanticsimilaritydetailstable", "semanticsimilaritytable", "session_user", "set", "setuser", "shutdown", "some", "statistics", "system_user", "table", "tablesample", "textsize", "then", "to", "top", "tran", "transaction", "trigger", "truncate", "try_convert", "tsequal", "union", "unique", "unpivot", "update", "updatetext", "use", "user", "values", "varying", "view", "waitfor", "when", "where", "while", "with", "within", "writetext"},
		MaxTables:            0, // Unlimited
		MaxColumns:           1024,
		MaxIndexes:           999,
		MaxForeignKeys:       253,
		DefaultCharset:       "UTF-8",
		DefaultCollation:     "SQL_Latin1_General_CP1_CI_AS",
		MaxPartitions:        15000,
		MaxTriggers:          0, // Unlimited
		MaxViews:             0, // Unlimited
		MaxStoredProcedures:  0, // Unlimited
		MaxFunctions:         0, // Unlimited
		MaxSequences:         0, // Unlimited
		MaxMaterializedViews: 0, // Not supported
	},
	Oracle: {
		Type:                 Oracle,
		DefaultSchema:        "",
		SchemaPrefix:         "",
		IdentifierQuote:      "\"",
		StringQuote:          "'",
		MaxIdentifierLength:  30,
		ReservedWords:        []string{"access", "add", "all", "alter", "and", "any", "as", "asc", "audit", "between", "by", "char", "check", "cluster", "column", "comment", "compress", "connect", "create", "current", "date", "decimal", "default", "delete", "desc", "distinct", "drop", "else", "exclusive", "exists", "file", "float", "for", "from", "grant", "group", "having", "identified", "immediate", "in", "increment", "index", "initial", "insert", "integer", "intersect", "into", "is", "level", "like", "lock", "long", "maxextents", "minus", "mlslabel", "mode", "modify", "noaudit", "nocompress", "not", "nowait", "null", "number", "of", "offline", "on", "online", "option", "or", "order", "pctfree", "prior", "privileges", "public", "raw", "rename", "resource", "revoke", "row", "rowid", "rownum", "rows", "select", "session", "set", "share", "size", "smallint", "start", "successful", "synonym", "sysdate", "table", "then", "to", "trigger", "uid", "union", "unique", "update", "user", "validate", "values", "varchar", "varchar2", "view", "whenever", "where", "with"},
		MaxTables:            0, // Unlimited
		MaxColumns:           1000,
		MaxIndexes:           0, // Unlimited
		MaxForeignKeys:       0, // Unlimited
		DefaultCharset:       "AL32UTF8",
		DefaultCollation:     "USING_NLS_COMP",
		MaxPartitions:        1024000,
		MaxTriggers:          0, // Unlimited
		MaxViews:             0, // Unlimited
		MaxStoredProcedures:  0, // Unlimited
		MaxFunctions:         0, // Unlimited
		MaxSequences:         0, // Unlimited
		MaxMaterializedViews: 0, // Unlimited
	},
	SQLite: {
		Type:                 SQLite,
		DefaultSchema:        "main",
		SchemaPrefix:         "",
		IdentifierQuote:      "\"",
		StringQuote:          "'",
		MaxIdentifierLength:  0, // No limit
		ReservedWords:        []string{"abort", "action", "add", "after", "all", "alter", "analyze", "and", "as", "asc", "attach", "autoincrement", "before", "begin", "between", "by", "cascade", "case", "cast", "check", "collate", "column", "commit", "conflict", "constraint", "create", "cross", "current_date", "current_time", "current_timestamp", "database", "default", "deferrable", "deferred", "delete", "desc", "detach", "distinct", "drop", "each", "else", "end", "escape", "except", "exclusive", "exists", "explain", "fail", "for", "foreign", "from", "full", "glob", "group", "having", "if", "ignore", "immediate", "in", "index", "indexed", "initially", "inner", "insert", "instead", "intersect", "into", "is", "isnull", "join", "key", "left", "like", "limit", "match", "natural", "no", "not", "notnull", "null", "of", "offset", "on", "or", "order", "outer", "plan", "pragma", "primary", "query", "raise", "recursive", "references", "regexp", "reindex", "release", "rename", "replace", "restrict", "right", "rollback", "row", "savepoint", "select", "set", "table", "temp", "temporary", "then", "to", "transaction", "trigger", "union", "unique", "update", "using", "vacuum", "values", "view", "virtual", "when", "where", "with", "without"},
		MaxTables:            0, // Unlimited
		MaxColumns:           2000,
		MaxIndexes:           0, // Unlimited
		MaxForeignKeys:       0, // Unlimited
		DefaultCharset:       "UTF-8",
		DefaultCollation:     "BINARY",
		MaxPartitions:        0, // Not supported
		MaxTriggers:          0, // Unlimited
		MaxViews:             0, // Unlimited
		MaxStoredProcedures:  0, // Not supported
		MaxFunctions:         0, // Not supported
		MaxSequences:         0, // Not supported
		MaxMaterializedViews: 0, // Not supported
	},
}

// parseInt converts a string to an integer, returning 0 if the conversion fails
func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
