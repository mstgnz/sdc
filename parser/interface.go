package parser

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"unicode"
)

// DatabaseType represents the type of database
type DatabaseType string

// SecurityLevel represents the security level for parsing and validation
type SecurityLevel int

const (
	// SecurityLevelLow minimal security checks
	SecurityLevelLow SecurityLevel = iota
	// SecurityLevelMedium standard security checks
	SecurityLevelMedium
	// SecurityLevelHigh strict security checks
	SecurityLevelHigh
)

// SecurityOptions contains security-related configuration
type SecurityOptions struct {
	Level               SecurityLevel
	DisableComments     bool     // Disable SQL comments
	MaxIdentifierLength int      // Maximum length for identifiers
	MaxQueryLength      int      // Maximum length for SQL queries
	AllowedSchemas      []string // List of allowed schemas
	DisallowedKeywords  []string // List of disallowed keywords
	LogSensitiveData    bool     // Whether to log sensitive data
}

// DefaultSecurityOptions returns default security options
func DefaultSecurityOptions() SecurityOptions {
	return SecurityOptions{
		Level:               SecurityLevelMedium,
		DisableComments:     true,
		MaxIdentifierLength: 64,
		MaxQueryLength:      1000000,
		AllowedSchemas:      []string{"public", "dbo", "main"},
		DisallowedKeywords: []string{
			"DELETE", "DROP", "TRUNCATE", "ALTER", "GRANT", "REVOKE",
			"EXECUTE", "EXEC", "SHELL", "XP_", "SP_",
		},
		LogSensitiveData: false,
	}
}

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgresql"
	SQLServer  DatabaseType = "sqlserver"
	Oracle     DatabaseType = "oracle"
	SQLite     DatabaseType = "sqlite"
)

// DatabaseInfoMap contains database information for each supported database type
var DatabaseInfoMap = map[DatabaseType]DatabaseInfo{
	MySQL: {
		DefaultSchema:       "",
		IdentifierQuote:     "`",
		StringQuote:         "'",
		MaxIdentifierLength: 64,
	},
	PostgreSQL: {
		DefaultSchema:       "public",
		IdentifierQuote:     "\"",
		StringQuote:         "'",
		MaxIdentifierLength: 63,
	},
	SQLServer: {
		DefaultSchema:       "dbo",
		IdentifierQuote:     "\"",
		StringQuote:         "'",
		MaxIdentifierLength: 128,
	},
	Oracle: {
		DefaultSchema:       "",
		IdentifierQuote:     "\"",
		StringQuote:         "'",
		MaxIdentifierLength: 30,
	},
	SQLite: {
		DefaultSchema:       "main",
		IdentifierQuote:     "\"",
		StringQuote:         "'",
		MaxIdentifierLength: 0,
	},
}

// sanitizeString removes potentially dangerous characters from a string
func sanitizeString(s string) string {
	dangerous := []string{"'", "\"", ";", "--", "/*", "*/", "@@", "@"}
	result := s
	for _, ch := range dangerous {
		result = strings.ReplaceAll(result, ch, "")
	}
	return result
}

// logSensitiveData logs potentially sensitive data if logging is enabled
func logSensitiveData(data string, options SecurityOptions) {
	if options.LogSensitiveData {
		fmt.Printf("Warning: Potentially sensitive data found: %s\n", data)
	}
}

// validateIdentifierSafety checks if an identifier is safe to use
func validateIdentifierSafety(name string, options SecurityOptions) error {
	if name == "" {
		return fmt.Errorf("empty identifier name")
	}

	if len(name) > options.MaxIdentifierLength {
		return fmt.Errorf("identifier '%s' exceeds maximum length of %d", name, options.MaxIdentifierLength)
	}

	// Check for disallowed keywords
	for _, keyword := range options.DisallowedKeywords {
		if strings.EqualFold(name, keyword) {
			return fmt.Errorf("identifier '%s' matches disallowed keyword", name)
		}
	}

	// Validate first character
	firstChar := rune(name[0])
	if !unicode.IsLetter(firstChar) && firstChar != '_' {
		return fmt.Errorf("identifier '%s' must start with a letter or underscore", name)
	}

	// Validate remaining characters
	for _, ch := range name[1:] {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' && ch != '$' {
			return fmt.Errorf("identifier '%s' contains invalid character: %c", name, ch)
		}
	}

	return nil
}

// validateQuerySafety performs security checks on SQL queries
func validateQuerySafety(sql string, options SecurityOptions) error {
	if sql == "" {
		return fmt.Errorf("empty SQL query")
	}

	if len(sql) > options.MaxQueryLength {
		return fmt.Errorf("SQL query exceeds maximum length of %d", options.MaxQueryLength)
	}

	upperSQL := strings.ToUpper(sql)

	// Check for dangerous patterns
	dangerousPatterns := []string{
		"EXECUTE", "EXEC", "SP_", "XP_",
		"WAITFOR", "DELAY", "SHUTDOWN",
		"BULK INSERT", "OPENROWSET", "OPENQUERY",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(upperSQL, pattern) {
			return fmt.Errorf("SQL query contains dangerous pattern: %s", pattern)
		}
	}

	// Check for stacked queries
	if strings.Count(sql, ";") > 1 {
		return fmt.Errorf("multiple SQL statements are not allowed")
	}

	// Check for comments if disabled
	if options.DisableComments {
		if strings.Contains(sql, "--") || strings.Contains(sql, "/*") {
			return fmt.Errorf("SQL comments are not allowed")
		}
	}

	return nil
}

// DatabaseInfo holds database-specific information
type DatabaseInfo struct {
	DefaultSchema       string
	IdentifierQuote     string
	StringQuote         string
	MaxIdentifierLength int
	ReservedWords       []string
}

// Entity represents a database entity
type Entity struct {
	Tables []*Table
}

// Table represents a database table
type Table struct {
	Schema      string
	Name        string
	Columns     []*Column
	Constraints []*Constraint
	PrimaryKey  *Constraint
	ForeignKeys []*Constraint
}

// Column represents a table column
type Column struct {
	Name          string
	DataType      *DataType
	IsNullable    bool
	Nullable      bool
	Default       string
	AutoIncrement bool
	PrimaryKey    bool
	Unique        bool
	Collation     string
}

// Constraint represents a table constraint
type Constraint struct {
	Name       string
	Type       string
	Columns    []string
	RefTable   string
	RefColumns []string
	OnDelete   string
	OnUpdate   string
}

// Parser interface for SQL parsing and conversion
type Parser interface {
	Parse(sql string) (*Entity, error)
	Convert(entity *Entity) (string, error)
	ValidateIdentifier(name string) error
	EscapeIdentifier(name string) string
	EscapeString(value string) string
	GetDefaultSchema() string
	GetSchemaPrefix(schema string) string
	GetIdentifierQuote() string
	GetStringQuote() string
	GetMaxIdentifierLength() int
	GetReservedWords() []string
}

// QueryExecutor basic query operations interface
type QueryExecutor interface {
	Execute(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// TransactionManager transaction operations interface
type TransactionManager interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// ConnectionManager connection management interface
type ConnectionManager interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
}

// SchemaManager schema management interface
type SchemaManager interface {
	CreateTable(ctx context.Context, table string) error
	DropTable(ctx context.Context, table string) error
	AlterTable(ctx context.Context, table string, alterations []string) error
}

// MigrationManager migration operations interface
type MigrationManager interface {
	ApplyMigrations(ctx context.Context) error
	RollbackMigration(ctx context.Context) error
	GetMigrationStatus(ctx context.Context) ([]MigrationStatus, error)
}

// MigrationStatus migration status struct
type MigrationStatus struct {
	ID        string
	Name      string
	Version   string
	AppliedAt string
	Status    string
}
