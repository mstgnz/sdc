package sqlporter

// DatabaseType represents the supported database types
type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgresql"
	SQLServer  DatabaseType = "sqlserver"
	Oracle     DatabaseType = "oracle"
	SQLite     DatabaseType = "sqlite"
)

// Schema represents a database schema
type Schema struct {
	Name        string
	Tables      []Table
	Procedures  []Procedure
	Functions   []Function
	Triggers    []Trigger
	Views       []View
	Sequences   []Sequence
	Extensions  []Extension
	Permissions []Permission
}

// Table represents a database table
type Table struct {
	Name        string
	Columns     []Column
	Indexes     []Index
	Constraints []Constraint
	Data        []Row
}

// Column represents a table column
type Column struct {
	Name          string
	DataType      string
	Length        int
	Precision     int
	Scale         int
	IsNullable    bool
	DefaultValue  string
	AutoIncrement bool
	Comment       string
}

// Index represents a table index
type Index struct {
	Name      string
	Columns   []string
	IsUnique  bool
	Type      string // BTREE, HASH etc.
	Condition string // WHERE clause
}

// Constraint represents a table constraint
type Constraint struct {
	Name            string
	Type            string // PRIMARY KEY, FOREIGN KEY, UNIQUE, CHECK
	Columns         []string
	RefTable        string
	RefColumns      []string
	UpdateRule      string
	DeleteRule      string
	CheckExpression string
}

// Row represents table data
type Row struct {
	Values map[string]interface{}
}

// Procedure represents a stored procedure
type Procedure struct {
	Name       string
	Parameters []Parameter
	Body       string
	Language   string
}

// Function represents a database function
type Function struct {
	Name       string
	Parameters []Parameter
	Returns    string
	Body       string
	Language   string
}

// Parameter represents a procedure or function parameter
type Parameter struct {
	Name      string
	DataType  string
	Direction string // IN, OUT, INOUT
}

// Trigger represents a database trigger
type Trigger struct {
	Name       string
	Table      string
	Timing     string // BEFORE, AFTER, INSTEAD OF
	Event      string // INSERT, UPDATE, DELETE
	Body       string
	ForEachRow bool
	Condition  string
}

// View represents a database view
type View struct {
	Name       string
	Definition string
	IsMaterial bool
}

// Sequence represents a database sequence
type Sequence struct {
	Name       string
	Start      int64
	Increment  int64
	MinValue   int64
	MaxValue   int64
	Cache      int64
	Cycle      bool
	CurrentVal int64
}

// Extension represents a database extension
type Extension struct {
	Name    string
	Version string
	Schema  string
}

// Permission represents a database permission
type Permission struct {
	Object      string
	ObjectType  string
	Grantee     string
	Privileges  []string
	GrantOption bool
}

// Parser represents an interface for database dump operations
type Parser interface {
	Parse(content string) (*Schema, error)
	Generate(schema *Schema) (string, error)
}
