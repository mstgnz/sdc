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
	Name             string
	Tables           []Table
	Procedures       []Procedure
	Functions        []Function
	Triggers         []Trigger
	Views            []View
	Sequences        []Sequence
	Extensions       []Extension
	Permissions      []Permission
	UserDefinedTypes []UserDefinedType
	Partitions       map[string][]Partition // table_name -> partitions
	DatabaseLinks    []DatabaseLink
	Tablespaces      []Tablespace
	Roles            []Role
	Users            []User
	Clusters         []Cluster
	MaterializedLogs []MaterializedViewLog
	Types            []Type
}

// Table represents a database table
type Table struct {
	Name        string
	Schema      string
	Columns     []Column
	Indexes     []Index
	Constraints []Constraint
	Data        []Row
	TableSpace  string
	Storage     *StorageClause
	Temporary   bool
	Comment     string
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
	Collation     string
	CharacterSet  string
}

// Index represents a table index
type Index struct {
	Name        string
	Columns     []string
	IsUnique    bool
	Type        string // BTREE, HASH etc.
	Condition   string // WHERE clause
	TableSpace  string
	Storage     *StorageClause
	Compression bool
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
	Deferrable      bool
	Initially       string // IMMEDIATE, DEFERRED
}

// Row represents table data
type Row struct {
	Values map[string]interface{}
}

// Procedure represents a stored procedure
type Procedure struct {
	Name          string
	Schema        string
	Parameters    []Parameter
	Body          string
	Language      string
	Security      string // DEFINER, INVOKER
	SQLSecurity   string
	Deterministic bool
	Comment       string
}

// Function represents a database function
type Function struct {
	Name          string
	Schema        string
	Parameters    []Parameter
	Returns       string
	Body          string
	Language      string
	Security      string // DEFINER, INVOKER
	SQLSecurity   string
	Deterministic bool
	Comment       string
}

// Parameter represents a procedure or function parameter
type Parameter struct {
	Name      string
	DataType  string
	Direction string // IN, OUT, INOUT
	Default   string
}

// Trigger represents a database trigger
type Trigger struct {
	Name       string
	Schema     string
	Table      string
	Timing     string // BEFORE, AFTER, INSTEAD OF
	Event      string // INSERT, UPDATE, DELETE
	Body       string
	ForEachRow bool
	Condition  string
	Enabled    bool
}

// View represents a database view
type View struct {
	Name        string
	Schema      string
	Definition  string
	IsMaterial  bool
	Columns     []string
	CheckOption string // NONE, LOCAL, CASCADED
	Security    string // DEFINER, INVOKER
	Comment     string
}

// Sequence represents a database sequence
type Sequence struct {
	Name       string
	Schema     string
	Start      int64
	Increment  int64
	MinValue   int64
	MaxValue   int64
	Cache      int64
	Cycle      bool
	CurrentVal int64
	Order      bool
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

// UserDefinedType represents custom data types
type UserDefinedType struct {
	Name       string
	Schema     string
	BaseType   string
	Properties map[string]interface{}
}

// Partition represents table partition information
type Partition struct {
	Name          string
	Type          string // RANGE, LIST, HASH
	SubPartitions []SubPartition
	Expression    string
	Values        []string
	TableSpace    string
	Storage       *StorageClause
}

// SubPartition represents table sub-partition information
type SubPartition struct {
	Name       string
	Type       string
	Expression string
	Values     []string
	TableSpace string
	Storage    *StorageClause
}

// MaterializedViewLog represents materialized view log information
type MaterializedViewLog struct {
	Name           string
	Schema         string
	TableName      string
	Columns        []string
	RowID          bool
	PrimaryKey     bool
	SequenceNumber bool
	CommitSCN      bool
	Storage        *StorageClause
}

// DatabaseLink represents database link information
type DatabaseLink struct {
	Name        string
	Owner       string
	ConnectInfo string
	Public      bool
}

// Tablespace represents tablespace information
type Tablespace struct {
	Name        string
	Type        string // PERMANENT, TEMPORARY
	Status      string
	Autoextend  bool
	MaxSize     int64
	InitialSize int64
	DataFile    string
	BlockSize   int
	Logging     bool
}

// Role represents database role information
type Role struct {
	Name        string
	Password    string
	Permissions []Permission
	Members     []string
	System      bool
}

// User represents database user information
type User struct {
	Name        string
	Password    string
	DefaultRole string
	Roles       []string
	Permissions []Permission
	Profile     string
	Status      string
	TableSpace  string
	TempSpace   string
}

// Cluster represents Oracle cluster information
type Cluster struct {
	Name       string
	Schema     string
	TableSpace string
	Key        []string
	Tables     []string
	Size       int
	HashKeys   int
	Storage    *StorageClause
}

// StorageClause represents storage properties
type StorageClause struct {
	Initial     int64
	Next        int64
	MinExtents  int
	MaxExtents  int
	Pctincrease int
	Buffer      int
	TableSpace  string
	Logging     bool
}

// Parser represents an interface for database dump operations
type Parser interface {
	Parse(content string) (*Schema, error)
	Generate(schema *Schema) (string, error)
}

// Type represents a database type
type Type struct {
	Name       string
	Schema     string
	Kind       string // ENUM, COMPOSITE, DOMAIN, etc.
	Definition string
}
