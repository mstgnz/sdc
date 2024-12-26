package sdc

// Entity represents the database structure.
// It contains all tables and their relationships in the database.
type Entity struct {
	Tables     []*Table     // List of tables in the database
	Sequences  []*Sequence  // List of sequences in the database
	Views      []*View      // List of views in the database
	Functions  []*Function  // List of functions in the database
	Procedures []*Procedure // List of stored procedures in the database
	Triggers   []*Trigger   // List of triggers in the database
	Schemas    []*Schema    // List of schemas in the database
}

// Schema represents a database schema structure
type Schema struct {
	Name       string       // Schema name
	Owner      string       // Schema owner
	Comment    string       // Schema comment
	Privileges []*Privilege // List of privileges on the schema
}

// Sequence represents a database sequence structure
type Sequence struct {
	Name       string       // Sequence name
	Schema     string       // Schema name
	Start      int64        // Start value
	Increment  int64        // Increment value
	MinValue   int64        // Minimum value
	MaxValue   int64        // Maximum value
	Cache      int64        // Cache size
	Cycle      bool         // Whether sequence should cycle
	Owner      string       // Sequence owner
	Comment    string       // Sequence comment
	Privileges []*Privilege // List of privileges on the sequence
}

// View represents a database view structure
type View struct {
	Name         string       // View name
	Schema       string       // Schema name
	Query        string       // View query
	Materialized bool         // Whether this is a materialized view
	Owner        string       // View owner
	Comment      string       // View comment
	Privileges   []*Privilege // List of privileges on the view
}

// Function represents a database function structure
type Function struct {
	Name       string       // Function name
	Schema     string       // Schema name
	Arguments  []*Argument  // List of function arguments
	Returns    string       // Return type
	Body       string       // Function body
	Language   string       // Function language
	Owner      string       // Function owner
	Comment    string       // Function comment
	Privileges []*Privilege // List of privileges on the function
}

// Procedure represents a stored procedure structure
type Procedure struct {
	Name       string       // Procedure name
	Schema     string       // Schema name
	Arguments  []*Argument  // List of procedure arguments
	Body       string       // Procedure body
	Language   string       // Procedure language
	Owner      string       // Procedure owner
	Comment    string       // Procedure comment
	Privileges []*Privilege // List of privileges on the procedure
}

// Trigger represents a trigger definition.
// It defines automated actions to be taken on specific table events.
type Trigger struct {
	Name       string       // Trigger name
	Schema     string       // Schema name
	Table      string       // Table name
	Event      string       // Triggering event (INSERT/UPDATE/DELETE)
	Timing     string       // Trigger timing (BEFORE/AFTER/INSTEAD OF)
	Body       string       // Trigger body/function
	Condition  string       // Conditional expression (WHEN clause)
	Owner      string       // Owner of the trigger
	Comment    string       // Trigger comment or description
	Privileges []*Privilege // List of privileges on the trigger
}

// Argument represents a function/procedure argument structure
type Argument struct {
	Name     string    // Argument name
	DataType *DataType // Argument data type
	Mode     string    // Argument mode (IN/OUT/INOUT)
	Default  string    // Default value
}

// Privilege represents a privilege structure
type Privilege struct {
	Grantee     string // Privilege grantee
	Privilege   string // Privilege type
	Grantor     string // Privilege grantor
	GrantOption bool   // Whether grant option is included
}

// Table represents a database table structure.
// It contains all information about a table including columns, constraints,
// indexes, and various database-specific features.
type Table struct {
	Name        string            // Table name
	Schema      string            // Schema name (database/namespace)
	Columns     []*Column         // List of columns in the table
	Constraints []*Constraint     // List of constraints (PRIMARY KEY, FOREIGN KEY, etc.)
	Indexes     []*Index          // List of indexes on the table
	FileGroup   string            // FileGroup name (SQL Server specific)
	Options     map[string]string // Additional table options
	Comment     string            // Table comment or description
	Collation   string            // Default collation for the table
	Inherits    []string          // List of parent tables (PostgreSQL INHERITS)
	Partitions  []*Partition      // Partition information
	Like        *Like             // LIKE template specification
	Unlogged    bool              // Whether this is an unlogged table
	Temporary   bool              // Whether this is a temporary table
	IfNotExists bool              // Whether to use IF NOT EXISTS clause
}

// Partition represents table partitioning information.
// It defines how table data is partitioned across multiple storage units.
type Partition struct {
	Name      string   // Partition name
	Type      string   // Partition type (RANGE/LIST/HASH)
	Columns   []string // Partitioning columns
	Values    []string // Partition values/bounds
	FileGroup string   // FileGroup for the partition
}

// Like represents a LIKE clause specification.
// It defines which properties to inherit from a template table.
type Like struct {
	Table         string   // Template table name
	Including     []string // Properties to include
	Excluding     []string // Properties to exclude
	WithDefaults  bool     // Include default values
	WithIndexes   bool     // Include indexes
	WithStorage   bool     // Include storage parameters
	WithComments  bool     // Include comments
	WithCollation bool     // Include collation
}

// Index represents an index structure.
// It defines how table data is indexed for faster access.
type Index struct {
	Name           string            // Index name
	Schema         string            // Schema name
	Table          string            // Table name
	Columns        []string          // Indexed columns
	Unique         bool              // Whether this is a unique index
	Clustered      bool              // Whether this is a clustered index
	NonClustered   bool              // Whether this is a non-clustered index
	FileGroup      string            // FileGroup for the index
	Filter         string            // Filter condition
	IncludeColumns []string          // Included (covered) columns
	Options        map[string]string // Index options
}

// DataType represents a column's data type specification.
// It includes all type-specific information including precision,
// scale, and various type modifiers.
type DataType struct {
	Name      string // Base type name
	Length    int    // Type length (e.g., VARCHAR(255))
	Precision int    // Numeric precision
	Scale     int    // Numeric scale
}

// ForeignKey represents a foreign key constraint.
// It defines relationships between tables through their columns.
type ForeignKey struct {
	Constraint        // Embedded constraint properties
	Name       string // Constraint name
	Column     string // Source column
	RefTable   string // Referenced table name
	RefColumn  string // Referenced column
	OnDelete   string // Action on delete (CASCADE, SET NULL, etc.)
	OnUpdate   string // Action on update
	Clustered  bool   // Whether this is a clustered foreign key
	FileGroup  string // FileGroup for the foreign key
}

// Column represents a database column structure.
// It contains all information about a column including its data type,
// constraints, and various database-specific features.
type Column struct {
	Name          string      // Column name
	DataType      *DataType   // Column data type information
	Length        int         // Type length
	Precision     int         // Numeric precision
	Scale         int         // Numeric scale
	IsNullable    bool        // Whether NULL values are allowed (parser compatibility)
	Nullable      bool        // Whether NULL values are allowed
	Default       string      // Default value expression
	AutoIncrement bool        // Whether column auto-increments
	PrimaryKey    bool        // Whether column is part of primary key
	Unique        bool        // Whether column has unique constraint
	Check         string      // Check constraint expression
	ForeignKey    *ForeignKey // Foreign key reference
	Comment       string      // Column comment or description
	Collation     string      // Column-specific collation
	Sparse        bool        // Whether this is a sparse column
	Computed      bool        // Whether this is a computed column
	ComputedExpr  string      // Expression for computed column
	Identity      bool        // Whether this is an identity column
	IdentitySeed  int64       // Identity seed value
	IdentityIncr  int64       // Identity increment value
	FileStream    bool        // Whether this is a FileStream column
	FileGroup     string      // FileGroup for the column
	RowGuidCol    bool        // Whether this is a rowguid column
	Persisted     bool        // Whether computed column is persisted
	Extra         string      // Additional column attributes (parser compatibility)
}

// Constraint represents a table constraint.
// It defines various types of constraints that can be applied to a table.
type Constraint struct {
	Name         string   // Constraint name
	Type         string   // Constraint type (PRIMARY KEY, FOREIGN KEY, UNIQUE, CHECK)
	Columns      []string // Constrained columns
	RefTable     string   // Referenced table (for foreign keys)
	RefColumns   []string // Referenced columns (for foreign keys)
	OnDelete     string   // Action on delete (for foreign keys)
	OnUpdate     string   // Action on update (for foreign keys)
	Check        string   // Check constraint expression
	Clustered    bool     // Whether this is a clustered constraint
	NonClustered bool     // Whether this is a non-clustered constraint
	FileGroup    string   // FileGroup for the constraint
}

// AlterTable represents an ALTER TABLE statement structure
type AlterTable struct {
	Schema     string      // Schema name
	Table      string      // Table name
	Action     string      // Action to perform
	Column     *Column     // Column to alter
	NewName    string      // New name for rename operations
	Constraint *Constraint // Constraint to add/modify
}

// DropTable represents a DROP TABLE statement structure
type DropTable struct {
	Schema   string // Schema name
	Table    string // Table name
	IfExists bool   // Whether to use IF EXISTS clause
	Cascade  bool   // Whether to cascade the drop operation
}

// DropIndex represents a DROP INDEX statement structure
type DropIndex struct {
	Schema   string // Schema name
	Table    string // Table name
	Index    string // Index name
	IfExists bool   // Whether to use IF EXISTS clause
	Cascade  bool   // Whether to cascade the drop operation
}
