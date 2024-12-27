package parser

import (
	"fmt"
	"strconv"
	"time"
)

// Common data type mappings
var (
	// MySQL to PostgreSQL mappings
	MySQLToPostgresMappings = []TypeMapping{
		{
			SourceType: "int",
			TargetType: "integer",
			ConversionFunc: func(v interface{}) (interface{}, error) {
				return v, nil
			},
		},
		{
			SourceType: "varchar",
			TargetType: "character varying",
			ConversionFunc: func(v interface{}) (interface{}, error) {
				return v, nil
			},
		},
		{
			SourceType: "datetime",
			TargetType: "timestamp",
			ConversionFunc: func(v interface{}) (interface{}, error) {
				if str, ok := v.(string); ok {
					return time.Parse("2006-01-02 15:04:05", str)
				}
				return v, nil
			},
		},
		{
			SourceType: "tinyint(1)",
			TargetType: "boolean",
			ConversionFunc: func(v interface{}) (interface{}, error) {
				switch val := v.(type) {
				case int64:
					return val != 0, nil
				case string:
					return strconv.ParseBool(val)
				default:
					return nil, fmt.Errorf("unsupported type for boolean conversion: %T", v)
				}
			},
		},
		{
			SourceType: "json",
			TargetType: "jsonb",
			ConversionFunc: func(v interface{}) (interface{}, error) {
				return v, nil
			},
		},
	}

	// PostgreSQL to MySQL mappings
	PostgresToMySQLMappings = []TypeMapping{
		{
			SourceType: "integer",
			TargetType: "int",
			ConversionFunc: func(v interface{}) (interface{}, error) {
				return v, nil
			},
		},
		{
			SourceType: "character varying",
			TargetType: "varchar",
			ConversionFunc: func(v interface{}) (interface{}, error) {
				return v, nil
			},
		},
		{
			SourceType: "timestamp",
			TargetType: "datetime",
			ConversionFunc: func(v interface{}) (interface{}, error) {
				if str, ok := v.(string); ok {
					t, err := time.Parse("2006-01-02 15:04:05.999999-07", str)
					if err != nil {
						return nil, err
					}
					return t.UTC(), nil
				}
				return v, nil
			},
		},
		{
			SourceType: "boolean",
			TargetType: "tinyint(1)",
			ConversionFunc: func(v interface{}) (interface{}, error) {
				if b, ok := v.(bool); ok {
					if b {
						return 1, nil
					}
					return 0, nil
				}
				return nil, fmt.Errorf("value is not a boolean: %v", v)
			},
		},
		{
			SourceType: "jsonb",
			TargetType: "json",
			ConversionFunc: func(v interface{}) (interface{}, error) {
				return v, nil
			},
		},
	}

	// SQLite specific mappings
	SQLiteMappings = []TypeMapping{
		{
			SourceType: "integer",
			TargetType: "integer",
			ConversionFunc: func(v interface{}) (interface{}, error) {
				return v, nil
			},
		},
		{
			SourceType: "text",
			TargetType: "text",
			ConversionFunc: func(v interface{}) (interface{}, error) {
				return v, nil
			},
		},
		{
			SourceType: "real",
			TargetType: "real",
			ConversionFunc: func(v interface{}) (interface{}, error) {
				return v, nil
			},
		},
		{
			SourceType: "blob",
			TargetType: "blob",
			ConversionFunc: func(v interface{}) (interface{}, error) {
				return v, nil
			},
		},
	}
)

// RegisterDefaultMappings registers default type mappings
func (c *Converter) RegisterDefaultMappings() {
	// Register MySQL to PostgreSQL mappings
	for _, m := range MySQLToPostgresMappings {
		c.RegisterMapping(m)
	}

	// Register PostgreSQL to MySQL mappings
	for _, m := range PostgresToMySQLMappings {
		c.RegisterMapping(m)
	}

	// Register SQLite mappings
	for _, m := range SQLiteMappings {
		c.RegisterMapping(m)
	}

	// Register default character sets
	c.RegisterCharSet(DefaultUTF8MB4)
	c.RegisterCharSet(DefaultLatin1)

	// Register default collations
	c.RegisterCollation(DefaultUTF8MB4Unicode)
	c.RegisterCollation(DefaultUTF8MB4Bin)
}
