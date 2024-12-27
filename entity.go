package sqlporter

// Database türlerini tanımlayan enum
type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgresql"
	SQLServer  DatabaseType = "sqlserver"
	Oracle     DatabaseType = "oracle"
	SQLite     DatabaseType = "sqlite"
)

// SchemaType veritabanı şemasını temsil eder
type SchemaType struct {
	Name        string
	Tables      []TableType
	Procedures  []ProcedureType
	Functions   []FunctionType
	Triggers    []TriggerType
	Views       []ViewType
	Sequences   []SequenceType
	Extensions  []ExtensionType
	Permissions []PermissionType
}

// TableType veritabanı tablosunu temsil eder
type TableType struct {
	Name        string
	Columns     []ColumnType
	Indexes     []IndexType
	Constraints []ConstraintType
	Data        []RowType
}

// ColumnType tablo kolonunu temsil eder
type ColumnType struct {
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

// IndexType tablo indeksini temsil eder
type IndexType struct {
	Name      string
	Columns   []string
	IsUnique  bool
	Type      string // BTREE, HASH vs.
	Condition string // WHERE koşulu
}

// ConstraintType tablo kısıtlamalarını temsil eder
type ConstraintType struct {
	Name            string
	Type            string // PRIMARY KEY, FOREIGN KEY, UNIQUE, CHECK
	Columns         []string
	RefTable        string
	RefColumns      []string
	UpdateRule      string
	DeleteRule      string
	CheckExpression string
}

// RowType tablo verisini temsil eder
type RowType struct {
	Values map[string]interface{}
}

// ProcedureType stored procedure'ü temsil eder
type ProcedureType struct {
	Name       string
	Parameters []ParameterType
	Body       string
	Language   string
}

// FunctionType fonksiyonu temsil eder
type FunctionType struct {
	Name       string
	Parameters []ParameterType
	Returns    string
	Body       string
	Language   string
}

// ParameterType prosedür ve fonksiyon parametrelerini temsil eder
type ParameterType struct {
	Name      string
	DataType  string
	Direction string // IN, OUT, INOUT
}

// TriggerType tetikleyiciyi temsil eder
type TriggerType struct {
	Name       string
	Table      string
	Timing     string // BEFORE, AFTER, INSTEAD OF
	Event      string // INSERT, UPDATE, DELETE
	Body       string
	ForEachRow bool
	Condition  string
}

// ViewType görünümü temsil eder
type ViewType struct {
	Name       string
	Definition string
	IsMaterial bool
}

// SequenceType sıralayıcıyı temsil eder
type SequenceType struct {
	Name       string
	Start      int64
	Increment  int64
	MinValue   int64
	MaxValue   int64
	Cache      int64
	Cycle      bool
	CurrentVal int64
}

// ExtensionType veritabanı eklentisini temsil eder
type ExtensionType struct {
	Name    string
	Version string
	Schema  string
}

// PermissionType yetkilendirmeyi temsil eder
type PermissionType struct {
	Object      string
	ObjectType  string
	Grantee     string
	Privileges  []string
	GrantOption bool
}

// ParserType veritabanı dump'larını işlemek için arayüz
type ParserType interface {
	Parse(content string) (*SchemaType, error)
	Generate(schema *SchemaType) (string, error)
}
