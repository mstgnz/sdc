package parser

import (
	"fmt"
	"strings"
)

// TypeConverter handles data type conversions between different database systems
type TypeConverter struct {
	sourceDialect string
	targetDialect string
	typeMap       map[string]map[string]string
	defaultMap    map[string]string
}

// NewTypeConverter creates a new type converter
func NewTypeConverter(sourceDialect, targetDialect string) *TypeConverter {
	tc := &TypeConverter{
		sourceDialect: sourceDialect,
		targetDialect: targetDialect,
		typeMap:       make(map[string]map[string]string),
		defaultMap:    make(map[string]string),
	}

	tc.initializeTypeMaps()
	return tc
}

// Convert converts a data type from source dialect to target dialect
func (tc *TypeConverter) Convert(sourceType string) (string, error) {
	// Normalize source type
	sourceType = strings.ToUpper(strings.TrimSpace(sourceType))

	// Check if there's a direct mapping for this type
	if targetTypes, exists := tc.typeMap[tc.sourceDialect]; exists {
		if targetType, exists := targetTypes[sourceType]; exists {
			return targetType, nil
		}
	}

	// Try to match with default mappings
	if targetType, exists := tc.defaultMap[sourceType]; exists {
		return targetType, nil
	}

	return "", fmt.Errorf("no conversion found for type %s from %s to %s",
		sourceType, tc.sourceDialect, tc.targetDialect)
}

// initializeTypeMaps sets up the type conversion mappings
func (tc *TypeConverter) initializeTypeMaps() {
	// Initialize maps for each dialect
	tc.typeMap["mysql"] = make(map[string]string)
	tc.typeMap["postgres"] = make(map[string]string)
	tc.typeMap["oracle"] = make(map[string]string)
	tc.typeMap["sqlserver"] = make(map[string]string)
	tc.typeMap["sqlite"] = make(map[string]string)

	// MySQL to PostgreSQL
	tc.addMapping("mysql", "postgres", map[string]string{
		"TINYINT":    "SMALLINT",
		"SMALLINT":   "SMALLINT",
		"MEDIUMINT":  "INTEGER",
		"INT":        "INTEGER",
		"BIGINT":     "BIGINT",
		"FLOAT":      "REAL",
		"DOUBLE":     "DOUBLE PRECISION",
		"DECIMAL":    "DECIMAL",
		"DATE":       "DATE",
		"DATETIME":   "TIMESTAMP",
		"TIMESTAMP":  "TIMESTAMP",
		"TIME":       "TIME",
		"YEAR":       "INTEGER",
		"CHAR":       "CHAR",
		"VARCHAR":    "VARCHAR",
		"TINYTEXT":   "TEXT",
		"TEXT":       "TEXT",
		"MEDIUMTEXT": "TEXT",
		"LONGTEXT":   "TEXT",
		"BINARY":     "BYTEA",
		"VARBINARY":  "BYTEA",
		"TINYBLOB":   "BYTEA",
		"BLOB":       "BYTEA",
		"MEDIUMBLOB": "BYTEA",
		"LONGBLOB":   "BYTEA",
		"ENUM":       "VARCHAR",
		"SET":        "VARCHAR[]",
		"JSON":       "JSONB",
		"BOOL":       "BOOLEAN",
	})

	// PostgreSQL to MySQL
	tc.addMapping("postgres", "mysql", map[string]string{
		"SMALLINT":         "SMALLINT",
		"INTEGER":          "INT",
		"BIGINT":           "BIGINT",
		"REAL":             "FLOAT",
		"DOUBLE PRECISION": "DOUBLE",
		"DECIMAL":          "DECIMAL",
		"NUMERIC":          "DECIMAL",
		"MONEY":            "DECIMAL(19,4)",
		"CHAR":             "CHAR",
		"VARCHAR":          "VARCHAR",
		"TEXT":             "TEXT",
		"BYTEA":            "BLOB",
		"TIMESTAMP":        "DATETIME",
		"DATE":             "DATE",
		"TIME":             "TIME",
		"BOOLEAN":          "TINYINT(1)",
		"SERIAL":           "INT AUTO_INCREMENT",
		"BIGSERIAL":        "BIGINT AUTO_INCREMENT",
		"JSONB":            "JSON",
		"UUID":             "CHAR(36)",
		"CIDR":             "VARCHAR(43)",
		"INET":             "VARCHAR(43)",
		"MACADDR":          "VARCHAR(17)",
		"BIT":              "BIT",
		"BIT VARYING":      "VARCHAR",
		"POINT":            "POINT",
		"LINE":             "LINESTRING",
		"POLYGON":          "POLYGON",
	})

	// Oracle specific mappings
	tc.addMapping("oracle", "postgres", map[string]string{
		"NUMBER":        "NUMERIC",
		"FLOAT":         "DOUBLE PRECISION",
		"VARCHAR2":      "VARCHAR",
		"NVARCHAR2":     "VARCHAR",
		"CHAR":          "CHAR",
		"NCHAR":         "CHAR",
		"DATE":          "TIMESTAMP",
		"TIMESTAMP":     "TIMESTAMP",
		"CLOB":          "TEXT",
		"NCLOB":         "TEXT",
		"BLOB":          "BYTEA",
		"BFILE":         "BYTEA",
		"RAW":           "BYTEA",
		"LONG RAW":      "BYTEA",
		"ROWID":         "VARCHAR(18)",
		"UROWID":        "VARCHAR(18)",
		"XMLTYPE":       "XML",
		"INTERVAL YEAR": "INTERVAL YEAR TO MONTH",
		"INTERVAL DAY":  "INTERVAL DAY TO SECOND",
		"BINARY_FLOAT":  "REAL",
		"BINARY_DOUBLE": "DOUBLE PRECISION",
		"LONG":          "TEXT",
	})

	// SQL Server specific mappings
	tc.addMapping("sqlserver", "postgres", map[string]string{
		"BIGINT":           "BIGINT",
		"BINARY":           "BYTEA",
		"BIT":              "BOOLEAN",
		"CHAR":             "CHAR",
		"DATE":             "DATE",
		"DATETIME":         "TIMESTAMP",
		"DATETIME2":        "TIMESTAMP",
		"DATETIMEOFFSET":   "TIMESTAMP WITH TIME ZONE",
		"DECIMAL":          "DECIMAL",
		"FLOAT":            "DOUBLE PRECISION",
		"IMAGE":            "BYTEA",
		"INT":              "INTEGER",
		"MONEY":            "DECIMAL(19,4)",
		"NCHAR":            "CHAR",
		"NTEXT":            "TEXT",
		"NUMERIC":          "NUMERIC",
		"NVARCHAR":         "VARCHAR",
		"REAL":             "REAL",
		"SMALLDATETIME":    "TIMESTAMP",
		"SMALLINT":         "SMALLINT",
		"SMALLMONEY":       "DECIMAL(10,4)",
		"TEXT":             "TEXT",
		"TIME":             "TIME",
		"TINYINT":          "SMALLINT",
		"UNIQUEIDENTIFIER": "UUID",
		"VARBINARY":        "BYTEA",
		"VARCHAR":          "VARCHAR",
		"XML":              "XML",
	})

	// SQLite specific mappings
	tc.addMapping("sqlite", "postgres", map[string]string{
		"INTEGER":  "INTEGER",
		"REAL":     "REAL",
		"TEXT":     "TEXT",
		"BLOB":     "BYTEA",
		"NUMERIC":  "NUMERIC",
		"BOOLEAN":  "BOOLEAN",
		"DATETIME": "TIMESTAMP",
		"DATE":     "DATE",
		"TIME":     "TIME",
	})

	// Default mappings for common types
	tc.defaultMap = map[string]string{
		"INT":       "INTEGER",
		"VARCHAR":   "VARCHAR",
		"TEXT":      "TEXT",
		"TIMESTAMP": "TIMESTAMP",
		"BOOLEAN":   "BOOLEAN",
		"DECIMAL":   "DECIMAL",
		"FLOAT":     "FLOAT",
		"DOUBLE":    "DOUBLE PRECISION",
		"DATE":      "DATE",
		"TIME":      "TIME",
		"BLOB":      "BYTEA",
		"CHAR":      "CHAR",
	}
}

// addMapping adds type mappings for a specific source and target dialect
func (tc *TypeConverter) addMapping(source, target string, mappings map[string]string) {
	if _, exists := tc.typeMap[source]; !exists {
		tc.typeMap[source] = make(map[string]string)
	}
	for sourceType, targetType := range mappings {
		tc.typeMap[source][sourceType] = targetType
	}
}

// AddCustomMapping adds a custom type mapping
func (tc *TypeConverter) AddCustomMapping(sourceType, targetType string) {
	if _, exists := tc.typeMap[tc.sourceDialect]; !exists {
		tc.typeMap[tc.sourceDialect] = make(map[string]string)
	}
	tc.typeMap[tc.sourceDialect][sourceType] = targetType
}

// GetSupportedTypes returns a list of supported types for a given dialect
func (tc *TypeConverter) GetSupportedTypes(dialect string) []string {
	var types []string
	if mappings, exists := tc.typeMap[dialect]; exists {
		for t := range mappings {
			types = append(types, t)
		}
	}
	return types
}
