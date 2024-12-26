package parser

// Statement represents a SQL statement
type Statement struct {
	Query string
	Args  []interface{}
}

// NewStatement creates a new SQL statement
func NewStatement(query string, args ...interface{}) *Statement {
	return &Statement{
		Query: query,
		Args:  args,
	}
}
