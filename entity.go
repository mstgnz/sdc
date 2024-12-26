package sdc

// Entity represents the database structure.
// It contains all tables and their relationships in the database.
type Entity struct {
	Tables []*Table // List of tables in the database
}

// Table represents a database table structure.
// It contains all information about a table including columns, constraints,
// indexes, and various database-specific features.
type Table struct {
	Schema          string            // Schema name (database/namespace)
	Name            string            // Table name
	Columns         []*Column         // List of columns in the table
	PrimaryKey      []string          // List of column names that form the primary key
	ForeignKeys     []*ForeignKey     // List of foreign key constraints
	Indexes         []*Index          // List of indexes on the table
	Checks          []*Check          // List of check constraints
	Exclusions      []*Exclusion      // List of exclusion constraints (PostgreSQL)
	Engine          string            // Storage engine (MySQL specific)
	TableSpace      string            // Tablespace name (Oracle/PostgreSQL)
	CharacterSet    string            // Default character set for the table
	Collation       string            // Default collation for the table
	Comment         string            // Table comment or description
	IsTemporary     bool              // Whether this is a temporary table
	IsUnlogged      bool              // Whether this is an unlogged table (PostgreSQL)
	Partition       *Partition        // Partition information
	Storage         map[string]string // Storage parameters (WITH clause)
	Inherits        []string          // List of parent tables (PostgreSQL INHERITS)
	Like            *Like             // LIKE template specification (PostgreSQL)
	ReplicaIdentity string            // Replica identity setting (PostgreSQL)
	AccessMethod    string            // Table access method (PostgreSQL)
	RowSecurity     bool              // Whether row level security is enabled
	Options         map[string]string // Additional table options
}

// Column represents a database column structure.
// It contains all information about a column including its data type,
// constraints, and various database-specific features.
type Column struct {
	Name         string     // Column name
	DataType     *DataType  // Column data type information
	IsNullable   bool       // Whether NULL values are allowed
	IsUnsigned   bool       // Whether the numeric type is unsigned
	Default      string     // Default value expression
	Identity     *Identity  // Identity/Generated column information
	Generated    *Generated // Generated column information
	IsComputed   bool       // Whether this is a computed column
	ComputedExpr string     // Expression for computed column
	IsPersisted  bool       // Whether computed column is persisted
	IsSparse     bool       // Whether this is a sparse column (SQL Server)
	CharacterSet string     // Column-specific character set
	Collation    string     // Column-specific collation
	Comment      string     // Column comment or description
	Extra        string     // Additional column attributes
	Storage      string     // Storage directive (PostgreSQL)
	Compression  string     // Compression method
	Statistics   int        // Statistics target (PostgreSQL)
}

// Generated represents a generated/computed column specification.
// It defines how the column value is generated and stored.
type Generated struct {
	Expression string // The expression that generates the column value
	Type       string // Storage type (STORED/VIRTUAL)
	Scope      string // Generation scope (ALWAYS/BY DEFAULT)
}

// DataType represents a column's data type specification.
// It includes all type-specific information including precision,
// scale, and various type modifiers.
type DataType struct {
	Name          string   // Base type name
	Length        int      // Type length (e.g., VARCHAR(255))
	Precision     int      // Numeric precision
	Scale         int      // Numeric scale
	IsArray       bool     // Whether this is an array type
	ArrayDims     []int    // Array dimensions
	TimeZone      bool     // Whether type includes timezone
	IntervalField string   // Interval type field
	IsCustom      bool     // Whether this is a custom/user-defined type
	TypeModifiers []string // Additional type modifiers
}

// Identity represents identity/generated column properties.
// It defines how the column values are automatically generated.
type Identity struct {
	Generation string // Generation type (ALWAYS/BY DEFAULT)
	Start      int64  // Starting value
	Increment  int64  // Increment value
	MinValue   *int64 // Minimum allowed value
	MaxValue   *int64 // Maximum allowed value
	Cache      *int64 // Number of values to cache
	Cycle      bool   // Whether to cycle when limits are reached
}

// Constraint represents common properties for all constraint types.
// It serves as a base for specific constraint implementations.
type Constraint struct {
	Name              string // Constraint name
	Type              string // Constraint type (PRIMARY KEY, FOREIGN KEY, etc.)
	Deferrable        bool   // Whether constraint check can be deferred
	InitiallyDeferred bool   // Whether constraint is initially deferred
	NoInherit         bool   // Whether constraint is not inherited
	Validated         bool   // Whether constraint is validated
}

// ForeignKey represents a foreign key constraint.
// It defines relationships between tables through their columns.
type ForeignKey struct {
	Constraint                 // Embedded constraint properties
	Columns           []string // Source columns
	ReferencedTable   string   // Referenced table name
	ReferencedColumns []string // Referenced columns
	OnDelete          string   // Action on delete (CASCADE, SET NULL, etc.)
	OnUpdate          string   // Action on update
	Match             string   // Match type (FULL/PARTIAL/SIMPLE)
}

// Index represents an index structure.
// It defines how table data is indexed for faster access.
type Index struct {
	Name           string            // Index name
	Columns        []string          // Indexed columns
	IsUnique       bool              // Whether this is a unique index
	IsClustered    bool              // Whether this is a clustered index
	Type           string            // Index type (BTREE/HASH/GiST etc.)
	TableSpace     string            // Index tablespace
	Comment        string            // Index comment
	IncludeColumns []string          // Included (covered) columns
	Filter         string            // Filter condition
	Predicate      string            // Partial index predicate
	OperatorClass  []string          // Operator classes for index
	Storage        map[string]string // Index storage parameters
	IsConcurrent   bool              // Whether index was created concurrently
	NullsOrder     string            // NULL ordering (NULLS FIRST/LAST)
}

// Check represents a check constraint.
// It defines conditions that must be true for all rows.
type Check struct {
	Constraint        // Embedded constraint properties
	Condition  string // Check condition expression
}

// Exclusion represents an exclusion constraint.
// It ensures that specific operations between rows yield no matches.
type Exclusion struct {
	Constraint                      // Embedded constraint properties
	Elements     []ExclusionElement // Exclusion elements
	AccessMethod string             // Index access method
}

// ExclusionElement represents an element in an exclusion constraint.
// It defines a column-operator pair for exclusion checking.
type ExclusionElement struct {
	Column   string // Column name
	Operator string // Exclusion operator
}

// Like represents a LIKE clause specification.
// It defines which properties to inherit from a template table.
type Like struct {
	Table              string // Template table name
	IncludeAll         bool   // Include all properties
	IncludeDefaults    bool   // Include default values
	IncludeConstraints bool   // Include constraints
	IncludeIndexes     bool   // Include indexes
	IncludeStorage     bool   // Include storage parameters
	IncludeComments    bool   // Include comments
}

// Partition represents table partitioning information.
// It defines how table data is partitioned across multiple storage units.
type Partition struct {
	Type         string     // Partition type (RANGE/LIST/HASH)
	Columns      []string   // Partitioning columns
	Strategy     string     // Partitioning strategy
	SubPartition *Partition // Sub-partition specification
	Bounds       []string   // Partition bounds/values
}

// Trigger represents a trigger definition.
// It defines automated actions to be taken on specific table events.
type Trigger struct {
	Name       string   // Trigger name
	Event      string   // Triggering event (INSERT/UPDATE/DELETE)
	Timing     string   // Trigger timing (BEFORE/AFTER/INSTEAD OF)
	Function   string   // Trigger function
	ForEach    string   // Execution scope (ROW/STATEMENT)
	When       string   // Conditional expression
	Columns    []string // Columns that trigger the action
	Referenced []string // Referenced tables in trigger
}
