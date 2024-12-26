package parser

import (
	"testing"
	"time"
)

func TestConverter_ConvertType(t *testing.T) {
	tests := []struct {
		name        string
		sourceType  string
		targetType  string
		value       interface{}
		expected    interface{}
		expectError bool
		sourceVer   *Version
		targetVer   *Version
	}{
		{
			name:       "MySQL int to PostgreSQL integer",
			sourceType: "int",
			targetType: "integer",
			value:      42,
			expected:   42,
			sourceVer:  &Version{Major: 5, Minor: 7},
			targetVer:  &Version{Major: 12},
		},
		{
			name:       "MySQL varchar to PostgreSQL character varying",
			sourceType: "varchar",
			targetType: "character varying",
			value:      "test",
			expected:   "test",
			sourceVer:  &Version{Major: 5, Minor: 7},
			targetVer:  &Version{Major: 12},
		},
		{
			name:       "MySQL datetime to PostgreSQL timestamp",
			sourceType: "datetime",
			targetType: "timestamp",
			value:      "2023-01-01 12:00:00",
			expected:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			sourceVer:  &Version{Major: 5, Minor: 7},
			targetVer:  &Version{Major: 12},
		},
		{
			name:        "Invalid type conversion",
			sourceType:  "invalid",
			targetType:  "invalid",
			value:       42,
			expectError: true,
		},
	}

	converter := NewConverter()
	converter.RegisterDefaultMappings()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ConvertType(tt.value, tt.sourceType, tt.targetType, tt.sourceVer, tt.targetVer)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestConverter_CharSetOperations(t *testing.T) {
	converter := NewConverter()
	converter.RegisterDefaultMappings()

	tests := []struct {
		name        string
		charset     string
		dbType      DataType
		expectError bool
	}{
		{
			name:    "UTF8MB4 supported in MySQL",
			charset: "utf8mb4",
			dbType:  TypeMySQL,
		},
		{
			name:    "UTF8MB4 supported in PostgreSQL",
			charset: "utf8mb4",
			dbType:  TypePostgreSQL,
		},
		{
			name:        "Invalid charset",
			charset:     "invalid",
			dbType:      TypeMySQL,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charset, err := converter.GetCharSet(tt.charset)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !charset.Supported[tt.dbType] {
				t.Errorf("expected charset %s to be supported in %s", tt.charset, tt.dbType)
			}
		})
	}
}
