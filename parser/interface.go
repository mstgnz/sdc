package parser

import (
	"strconv"
	"strings"

	"github.com/mstgnz/sdc"
)

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

// Parser interface defines the methods that must be implemented by all database parsers
type Parser interface {
	Parse(sql string) (*sdc.Table, error)
	Convert(table *sdc.Table) (string, error)
}

// parseNumber safely parses a string to an integer
func parseNumber(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return n, nil
}
