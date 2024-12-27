package parser

import (
	"strings"
	"testing"
)

func TestDefaultSecurityOptions(t *testing.T) {
	options := DefaultSecurityOptions()

	if options.Level != SecurityLevelMedium {
		t.Errorf("Expected security level %v, got %v", SecurityLevelMedium, options.Level)
	}

	if !options.DisableComments {
		t.Error("Expected comments to be disabled by default")
	}

	if options.MaxIdentifierLength != 64 {
		t.Errorf("Expected max identifier length 64, got %d", options.MaxIdentifierLength)
	}

	if options.MaxQueryLength != 1000000 {
		t.Errorf("Expected max query length 1000000, got %d", options.MaxQueryLength)
	}

	expectedSchemas := []string{"public", "dbo", "main"}
	if len(options.AllowedSchemas) != len(expectedSchemas) {
		t.Errorf("Expected %d allowed schemas, got %d", len(expectedSchemas), len(options.AllowedSchemas))
	}

	if options.LogSensitiveData {
		t.Error("Expected sensitive data logging to be disabled by default")
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "sql injection attempt",
			input:    "DROP TABLE users; --",
			expected: "DROP TABLE users",
		},
		{
			name:     "multiple dangerous characters",
			input:    "SELECT * FROM users WHERE name = 'admin' AND password = '' OR '1'='1'",
			expected: "SELECT * FROM users WHERE name = admin AND password =  OR 1=1",
		},
		{
			name:     "safe string",
			input:    "SELECT * FROM users",
			expected: "SELECT * FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValidateIdentifierSafety(t *testing.T) {
	options := DefaultSecurityOptions()

	tests := []struct {
		name          string
		identifier    string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid identifier",
			identifier:  "users_table",
			expectError: false,
		},
		{
			name:          "empty identifier",
			identifier:    "",
			expectError:   true,
			errorContains: "empty identifier",
		},
		{
			name:          "too long identifier",
			identifier:    "very_very_very_very_very_very_very_very_very_long_identifier_name",
			expectError:   true,
			errorContains: "exceeds maximum length",
		},
		{
			name:          "disallowed keyword",
			identifier:    "DROP",
			expectError:   true,
			errorContains: "matches disallowed keyword",
		},
		{
			name:          "invalid start character",
			identifier:    "1users",
			expectError:   true,
			errorContains: "must start with a letter or underscore",
		},
		{
			name:          "invalid character",
			identifier:    "users@table",
			expectError:   true,
			errorContains: "contains invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIdentifierSafety(tt.identifier, options)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateQuerySafety(t *testing.T) {
	options := DefaultSecurityOptions()

	tests := []struct {
		name          string
		query         string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid query",
			query:       "SELECT * FROM users WHERE id = 1",
			expectError: false,
		},
		{
			name:          "empty query",
			query:         "",
			expectError:   true,
			errorContains: "empty SQL query",
		},
		{
			name:          "too long query",
			query:         strings.Repeat("SELECT ", 200000),
			expectError:   true,
			errorContains: "exceeds maximum length",
		},
		{
			name:          "dangerous pattern",
			query:         "EXEC sp_executesql N'SELECT * FROM users'",
			expectError:   true,
			errorContains: "dangerous pattern",
		},
		{
			name:          "stacked queries",
			query:         "SELECT * FROM users; DROP TABLE users",
			expectError:   true,
			errorContains: "multiple SQL statements",
		},
		{
			name:          "comment in query",
			query:         "SELECT * FROM users -- get all users",
			expectError:   true,
			errorContains: "comments are not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQuerySafety(tt.query, options)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestDatabaseInfoMap(t *testing.T) {
	tests := []struct {
		name         DatabaseType
		expectExists bool
	}{
		{
			name:         MySQL,
			expectExists: true,
		},
		{
			name:         PostgreSQL,
			expectExists: true,
		},
		{
			name:         SQLServer,
			expectExists: true,
		},
		{
			name:         Oracle,
			expectExists: true,
		},
		{
			name:         SQLite,
			expectExists: true,
		},
		{
			name:         "unknown",
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			info, exists := DatabaseInfoMap[tt.name]
			if tt.expectExists {
				if !exists {
					t.Errorf("Expected database info for %s to exist", tt.name)
				}
				if info.IdentifierQuote == "" {
					t.Error("Expected non-empty identifier quote")
				}
				if info.StringQuote == "" {
					t.Error("Expected non-empty string quote")
				}
			} else if exists {
				t.Errorf("Expected database info for %s to not exist", tt.name)
			}
		})
	}
}
