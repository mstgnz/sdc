package gosql

type Entity struct {
	tables  []*Table
	content []*Data
}

type Table struct {
	columnName    string
	dataType      DataType
	isNullable    bool
	isUnsigned    bool
	columnDefault string
	check         string
	comment       string
	characterSet  string
	collation     string
	extra         string
	foreignKey    ForeignKey
	index         Index
}

type Data struct {
	tableName string
	keys      []string
	values    []string
}

type DataType struct {
	name  string
	param string
}

type Index struct {
	name        string
	algorithm   string
	isUnique    bool
	columnNames []string
	condition   string
	comment     string
}

type ForeignKey struct {
	table             string
	columns           []string
	referencedTable   string
	referencedColumns []string
	onUpdate          string // NO ACTION, RESTRICT, CASCADE,  SET NULL, SET DEFAULT
	onDelete          string
}

type Trigger struct {
	name     string
	event    string // INSERT, UPDATE, DELETE
	timing   string // BEFORE, AFTER
	function string
}
