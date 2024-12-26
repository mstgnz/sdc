package parser

import (
	"testing"
)

func TestNewTypeConverter(t *testing.T) {
	tests := []struct {
		name          string
		sourceDialect string
		targetDialect string
	}{
		{
			name:          "mysql to postgres",
			sourceDialect: "mysql",
			targetDialect: "postgres",
		},
		{
			name:          "postgres to mysql",
			sourceDialect: "postgres",
			targetDialect: "mysql",
		},
		{
			name:          "oracle to postgres",
			sourceDialect: "oracle",
			targetDialect: "postgres",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewTypeConverter(tt.sourceDialect, tt.targetDialect)
			if converter == nil {
				t.Error("Expected non-nil TypeConverter")
			}
			if converter.sourceDialect != tt.sourceDialect {
				t.Errorf("Expected source dialect %s, got %s", tt.sourceDialect, converter.sourceDialect)
			}
			if converter.targetDialect != tt.targetDialect {
				t.Errorf("Expected target dialect %s, got %s", tt.targetDialect, converter.targetDialect)
			}
		})
	}
}

func TestTypeConverter_Convert(t *testing.T) {
	tests := []struct {
		name          string
		sourceDialect string
		targetDialect string
		sourceType    string
		expected      string
		expectError   bool
	}{
		{
			name:          "mysql int to postgres",
			sourceDialect: "mysql",
			targetDialect: "postgres",
			sourceType:    "INT",
			expected:      "INTEGER",
			expectError:   false,
		},
		{
			name:          "mysql varchar to postgres",
			sourceDialect: "mysql",
			targetDialect: "postgres",
			sourceType:    "VARCHAR",
			expected:      "VARCHAR",
			expectError:   false,
		},
		{
			name:          "postgres jsonb to mysql",
			sourceDialect: "postgres",
			targetDialect: "mysql",
			sourceType:    "JSONB",
			expected:      "JSON",
			expectError:   false,
		},
		{
			name:          "oracle varchar2 to postgres",
			sourceDialect: "oracle",
			targetDialect: "postgres",
			sourceType:    "VARCHAR2",
			expected:      "VARCHAR",
			expectError:   false,
		},
		{
			name:          "unknown type",
			sourceDialect: "mysql",
			targetDialect: "postgres",
			sourceType:    "UNKNOWN_TYPE",
			expected:      "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewTypeConverter(tt.sourceDialect, tt.targetDialect)
			result, err := converter.Convert(tt.sourceType)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTypeConverter_AddCustomMapping(t *testing.T) {
	converter := NewTypeConverter("mysql", "postgres")

	// Add custom mapping
	customSourceType := "CUSTOM_TYPE"
	customTargetType := "CUSTOM_TARGET"
	converter.AddCustomMapping(customSourceType, customTargetType)

	// Test the custom mapping
	result, err := converter.Convert(customSourceType)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
	if result != customTargetType {
		t.Errorf("Expected %s, got %s", customTargetType, result)
	}
}

func TestTypeConverter_GetSupportedTypes(t *testing.T) {
	tests := []struct {
		name     string
		dialect  string
		minTypes int // Minimum number of expected types
	}{
		{
			name:     "mysql types",
			dialect:  "mysql",
			minTypes: 20, // Adjust based on actual implementation
		},
		{
			name:     "postgres types",
			dialect:  "postgres",
			minTypes: 20,
		},
		{
			name:     "oracle types",
			dialect:  "oracle",
			minTypes: 15,
		},
		{
			name:     "sqlserver types",
			dialect:  "sqlserver",
			minTypes: 15,
		},
		{
			name:     "sqlite types",
			dialect:  "sqlite",
			minTypes: 5,
		},
	}

	converter := NewTypeConverter("mysql", "postgres") // Source and target don't matter for this test

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			types := converter.GetSupportedTypes(tt.dialect)
			if len(types) < tt.minTypes {
				t.Errorf("Expected at least %d types for %s, got %d", tt.minTypes, tt.dialect, len(types))
			}
		})
	}
}

func TestTypeConverter_CaseInsensitivity(t *testing.T) {
	converter := NewTypeConverter("mysql", "postgres")

	tests := []struct {
		name        string
		sourceType  string
		expected    string
		expectError bool
	}{
		{
			name:        "lowercase",
			sourceType:  "int",
			expected:    "INTEGER",
			expectError: false,
		},
		{
			name:        "uppercase",
			sourceType:  "INT",
			expected:    "INTEGER",
			expectError: false,
		},
		{
			name:        "mixed case",
			sourceType:  "InT",
			expected:    "INTEGER",
			expectError: false,
		},
		{
			name:        "with spaces",
			sourceType:  "  INT  ",
			expected:    "INTEGER",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.Convert(tt.sourceType)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
