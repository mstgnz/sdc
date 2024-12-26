package parser

import (
	"testing"
	"time"
)

func TestMySQLToPostgresMappings(t *testing.T) {
	tests := []struct {
		name        string
		sourceType  string
		value       interface{}
		expected    interface{}
		expectError bool
	}{
		{
			name:        "int to integer",
			sourceType:  "int",
			value:       42,
			expected:    42,
			expectError: false,
		},
		{
			name:        "varchar to character varying",
			sourceType:  "varchar",
			value:       "test",
			expected:    "test",
			expectError: false,
		},
		{
			name:        "datetime to timestamp",
			sourceType:  "datetime",
			value:       "2023-01-01 12:00:00",
			expected:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "tinyint(1) to boolean - true",
			sourceType:  "tinyint(1)",
			value:       int64(1),
			expected:    true,
			expectError: false,
		},
		{
			name:        "tinyint(1) to boolean - false",
			sourceType:  "tinyint(1)",
			value:       int64(0),
			expected:    false,
			expectError: false,
		},
		{
			name:        "json to jsonb",
			sourceType:  "json",
			value:       `{"key": "value"}`,
			expected:    `{"key": "value"}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mapping TypeMapping
			for _, m := range MySQLToPostgresMappings {
				if m.SourceType == tt.sourceType {
					mapping = m
					break
				}
			}

			result, err := mapping.ConversionFunc(tt.value)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestPostgresToMySQLMappings(t *testing.T) {
	tests := []struct {
		name        string
		sourceType  string
		value       interface{}
		expected    interface{}
		expectError bool
	}{
		{
			name:        "integer to int",
			sourceType:  "integer",
			value:       42,
			expected:    42,
			expectError: false,
		},
		{
			name:        "character varying to varchar",
			sourceType:  "character varying",
			value:       "test",
			expected:    "test",
			expectError: false,
		},
		{
			name:        "timestamp to datetime",
			sourceType:  "timestamp",
			value:       "2023-01-01 12:00:00.000000+00",
			expected:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "boolean to tinyint(1) - true",
			sourceType:  "boolean",
			value:       true,
			expected:    1,
			expectError: false,
		},
		{
			name:        "boolean to tinyint(1) - false",
			sourceType:  "boolean",
			value:       false,
			expected:    0,
			expectError: false,
		},
		{
			name:        "jsonb to json",
			sourceType:  "jsonb",
			value:       `{"key": "value"}`,
			expected:    `{"key": "value"}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mapping TypeMapping
			for _, m := range PostgresToMySQLMappings {
				if m.SourceType == tt.sourceType {
					mapping = m
					break
				}
			}

			result, err := mapping.ConversionFunc(tt.value)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestSQLiteMappings(t *testing.T) {
	tests := []struct {
		name        string
		sourceType  string
		value       interface{}
		expected    interface{}
		expectError bool
	}{
		{
			name:        "integer",
			sourceType:  "integer",
			value:       42,
			expected:    42,
			expectError: false,
		},
		{
			name:        "text",
			sourceType:  "text",
			value:       "test",
			expected:    "test",
			expectError: false,
		},
		{
			name:        "real",
			sourceType:  "real",
			value:       3.14,
			expected:    3.14,
			expectError: false,
		},
		{
			name:        "blob",
			sourceType:  "blob",
			value:       []byte("binary data"),
			expected:    []byte("binary data"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mapping TypeMapping
			for _, m := range SQLiteMappings {
				if m.SourceType == tt.sourceType {
					mapping = m
					break
				}
			}

			result, err := mapping.ConversionFunc(tt.value)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestRegisterDefaultMappings(t *testing.T) {
	converter := &Converter{
		typeMappings: make(map[string][]TypeMapping),
		charSets:     make(map[string]CharSet),
		collations:   make(map[string]CollationConfig),
	}

	converter.RegisterDefaultMappings()

	// Check MySQL to PostgreSQL mappings
	for _, m := range MySQLToPostgresMappings {
		if _, exists := converter.typeMappings[m.SourceType]; !exists {
			t.Errorf("MySQL to PostgreSQL mapping not registered: %s", m.SourceType)
		}
	}

	// Check PostgreSQL to MySQL mappings
	for _, m := range PostgresToMySQLMappings {
		if _, exists := converter.typeMappings[m.SourceType]; !exists {
			t.Errorf("PostgreSQL to MySQL mapping not registered: %s", m.SourceType)
		}
	}

	// Check SQLite mappings
	for _, m := range SQLiteMappings {
		if _, exists := converter.typeMappings[m.SourceType]; !exists {
			t.Errorf("SQLite mapping not registered: %s", m.SourceType)
		}
	}

	// Check character sets
	if _, exists := converter.charSets[DefaultUTF8MB4.Name]; !exists {
		t.Error("UTF8MB4 character set not registered")
	}
	if _, exists := converter.charSets[DefaultLatin1.Name]; !exists {
		t.Error("Latin1 character set not registered")
	}

	// Check collations
	if _, exists := converter.collations[DefaultUTF8MB4Unicode.Name]; !exists {
		t.Error("UTF8MB4 Unicode collation not registered")
	}
	if _, exists := converter.collations[DefaultUTF8MB4Bin.Name]; !exists {
		t.Error("UTF8MB4 Binary collation not registered")
	}
}
