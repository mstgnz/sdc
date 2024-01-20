package gosql

type Entity struct {
	tables  []*Table
	content []*Content
}

type Table struct {
	columnName    string
	dataType      DataType
	isNullable    bool
	columnDefault string
	check         string
	foreignKey    string
	comment       string
	characterSet  string
	collation     string
	extra         string
}

type Content struct {
	tableName string
	keys      []string
	values    []string
}

type DataType struct {
	name           string
	internalParams []int    // varchar(20), int(11), ...
	externalParams []string // unsigned, current_timestamp, ...
}
