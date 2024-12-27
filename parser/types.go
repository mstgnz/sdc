package parser

import (
	"fmt"
	"strings"
)

// DataType represents supported database types
type DataType int

const (
	TypeMySQL DataType = iota
	TypePostgreSQL
	TypeSQLite
	TypeOracle
	TypeSQLServer
)

// String returns string representation of DataType
func (dt DataType) String() string {
	switch dt {
	case TypeMySQL:
		return "mysql"
	case TypePostgreSQL:
		return "postgresql"
	case TypeSQLite:
		return "sqlite"
	case TypeOracle:
		return "oracle"
	case TypeSQLServer:
		return "sqlserver"
	default:
		return "unknown"
	}
}

// Version represents a database version
type Version struct {
	Major    int
	Minor    int
	Patch    int
	Metadata string
}

// TypeMapping represents data type mapping between databases
type TypeMapping struct {
	SourceType     string
	SourceVersion  *Version
	TargetType     string
	TargetVersion  *Version
	ConversionFunc func(interface{}) (interface{}, error)
	Constraints    []string
}

// CharSet represents character set configuration
type CharSet struct {
	Name        string
	Description string
	MaxLength   int
	Supported   map[DataType]bool
}

// CollationConfig represents collation configuration
type CollationConfig struct {
	Name        string
	CharSet     string
	Description string
	Supported   map[DataType]bool
}

// Converter handles data type conversions
type Converter struct {
	typeMappings map[string][]TypeMapping
	charSets     map[string]CharSet
	collations   map[string]CollationConfig
}

// NewConverter creates a new type converter
func NewConverter() *Converter {
	return &Converter{
		typeMappings: make(map[string][]TypeMapping),
		charSets:     make(map[string]CharSet),
		collations:   make(map[string]CollationConfig),
	}
}

// RegisterMapping registers a new type mapping
func (c *Converter) RegisterMapping(mapping TypeMapping) {
	if c.typeMappings == nil {
		c.typeMappings = make(map[string][]TypeMapping)
	}
	c.typeMappings[mapping.SourceType] = []TypeMapping{mapping}
}

// RegisterCharSet registers a new character set
func (c *Converter) RegisterCharSet(charset CharSet) {
	if c.charSets == nil {
		c.charSets = make(map[string]CharSet)
	}
	c.charSets[charset.Name] = charset
}

// RegisterCollation registers a new collation
func (c *Converter) RegisterCollation(collation CollationConfig) {
	if c.collations == nil {
		c.collations = make(map[string]CollationConfig)
	}
	c.collations[collation.Name] = collation
}

// ConvertType converts a value from one type to another
func (c *Converter) ConvertType(value interface{}, sourceType string, targetType string, sourceVersion, targetVersion *Version) (interface{}, error) {
	key := fmt.Sprintf("%s_%s", sourceType, targetType)
	mappings, exists := c.typeMappings[key]
	if !exists {
		return nil, fmt.Errorf("no mapping found for %s to %s", sourceType, targetType)
	}

	// Find suitable mapping based on versions
	var selectedMapping *TypeMapping
	for _, m := range mappings {
		if isVersionCompatible(m.SourceVersion, sourceVersion) &&
			isVersionCompatible(m.TargetVersion, targetVersion) {
			selectedMapping = &m
			break
		}
	}

	if selectedMapping == nil {
		return nil, fmt.Errorf("no compatible version mapping found")
	}

	return selectedMapping.ConversionFunc(value)
}

// GetCharSet returns character set information
func (c *Converter) GetCharSet(name string) (CharSet, error) {
	charset, exists := c.charSets[strings.ToUpper(name)]
	if !exists {
		return CharSet{}, fmt.Errorf("character set %s not found", name)
	}
	return charset, nil
}

// GetCollation returns collation information
func (c *Converter) GetCollation(name string) (CollationConfig, error) {
	collation, exists := c.collations[strings.ToUpper(name)]
	if !exists {
		return CollationConfig{}, fmt.Errorf("collation %s not found", name)
	}
	return collation, nil
}

// isVersionCompatible checks if versions are compatible
func isVersionCompatible(required, actual *Version) bool {
	if required == nil || actual == nil {
		return true
	}

	if required.Major != actual.Major {
		return false
	}

	if required.Minor > actual.Minor {
		return false
	}

	return true
}

// Common character sets
var (
	DefaultUTF8MB4 = CharSet{
		Name:        "utf8mb4",
		Description: "UTF-8 Unicode",
		MaxLength:   4,
		Supported: map[DataType]bool{
			TypeMySQL:      true,
			TypePostgreSQL: true,
			TypeSQLite:     true,
			TypeOracle:     true,
			TypeSQLServer:  true,
		},
	}

	DefaultLatin1 = CharSet{
		Name:        "latin1",
		Description: "cp1252 West European",
		MaxLength:   1,
		Supported: map[DataType]bool{
			TypeMySQL:      true,
			TypePostgreSQL: true,
			TypeSQLite:     true,
			TypeOracle:     true,
			TypeSQLServer:  true,
		},
	}
)

// Common collations
var (
	DefaultUTF8MB4Unicode = CollationConfig{
		Name:        "utf8mb4_unicode_ci",
		CharSet:     "utf8mb4",
		Description: "Unicode (case-insensitive)",
		Supported: map[DataType]bool{
			TypeMySQL:      true,
			TypePostgreSQL: true,
			TypeSQLite:     true,
		},
	}

	DefaultUTF8MB4Bin = CollationConfig{
		Name:        "utf8mb4_bin",
		CharSet:     "utf8mb4",
		Description: "Unicode (binary)",
		Supported: map[DataType]bool{
			TypeMySQL:      true,
			TypePostgreSQL: true,
			TypeSQLite:     true,
		},
	}
)
