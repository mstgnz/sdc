package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/mstgnz/sdc"
)

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

// validateIdentifierSafety checks if an identifier is safe to use
func validateIdentifierSafety(name string, options SecurityOptions) error {
	if name == "" {
		return fmt.Errorf("empty identifier name")
	}

	if len(name) > options.MaxIdentifierLength {
		return fmt.Errorf("identifier '%s' exceeds maximum length of %d", name, options.MaxIdentifierLength)
	}

	// Check for SQL injection patterns
	dangerousPatterns := []string{
		"--", "/*", "*/", "@@", "@",
		"EXEC", "EXECUTE", "SP_", "XP_",
		"WAITFOR", "DELAY", "SHUTDOWN",
	}

	upperName := strings.ToUpper(name)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(upperName, pattern) {
			return fmt.Errorf("identifier '%s' contains dangerous pattern: %s", name, pattern)
		}
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

	// Check for disallowed keywords based on security level
	if options.Level >= SecurityLevelHigh {
		for _, keyword := range options.DisallowedKeywords {
			pattern := fmt.Sprintf(`\b%s\b`, keyword)
			if matched, _ := regexp.MatchString(pattern, upperSQL); matched {
				return fmt.Errorf("SQL query contains disallowed keyword: %s", keyword)
			}
		}
	}

	return nil
}

// sanitizeString removes potentially dangerous characters from a string
func sanitizeString(s string) string {
	// Remove common SQL injection characters
	dangerous := []string{"'", "\"", ";", "--", "/*", "*/", "@@", "@"}
	result := s

	for _, ch := range dangerous {
		result = strings.ReplaceAll(result, ch, "")
	}

	return result
}

// logSensitiveData logs sensitive data based on security options
func logSensitiveData(data string, options SecurityOptions) {
	if !options.LogSensitiveData {
		// Mask sensitive data
		maskedData := maskSensitiveData(data)
		fmt.Printf("Masked sensitive data: %s\n", maskedData)
		return
	}

	fmt.Printf("Original data: %s\n", data)
}

// maskSensitiveData masks sensitive information in a string
func maskSensitiveData(data string) string {
	// Mask common sensitive patterns
	patterns := map[string]*regexp.Regexp{
		"EMAIL":    regexp.MustCompile(`\b[\w\.-]+@[\w\.-]+\.\w+\b`),
		"PASSWORD": regexp.MustCompile(`(?i)password\s*=\s*[^\s;]+`),
		"API_KEY":  regexp.MustCompile(`(?i)(api[_-]?key|access[_-]?key|secret[_-]?key)\s*=\s*[^\s;]+`),
		"TOKEN":    regexp.MustCompile(`(?i)(token|jwt|bearer)\s*=\s*[^\s;]+`),
	}

	result := data
	for name, pattern := range patterns {
		result = pattern.ReplaceAllString(result, fmt.Sprintf("[MASKED_%s]", name))
	}

	return result
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
	// Parse converts SQL dump to Table structure with security validation
	Parse(sql string, options SecurityOptions) (*sdc.Table, error)
	// Convert transforms Table structure to target database format
	Convert(table *sdc.Table, options SecurityOptions) (string, error)
	// ValidateSchema validates the schema structure with security checks
	ValidateSchema(table *sdc.Table, options SecurityOptions) error
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
