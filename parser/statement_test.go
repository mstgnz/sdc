package parser

import (
	"testing"
)

func TestNewStatement(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		args     []interface{}
		expected *Statement
	}{
		{
			name:  "empty statement",
			query: "",
			args:  nil,
			expected: &Statement{
				Query: "",
				Args:  nil,
			},
		},
		{
			name:  "query without args",
			query: "SELECT * FROM users",
			args:  nil,
			expected: &Statement{
				Query: "SELECT * FROM users",
				Args:  nil,
			},
		},
		{
			name:  "query with args",
			query: "SELECT * FROM users WHERE id = ?",
			args:  []interface{}{1},
			expected: &Statement{
				Query: "SELECT * FROM users WHERE id = ?",
				Args:  []interface{}{1},
			},
		},
		{
			name:  "query with multiple args",
			query: "INSERT INTO users (name, age) VALUES (?, ?)",
			args:  []interface{}{"John", 30},
			expected: &Statement{
				Query: "INSERT INTO users (name, age) VALUES (?, ?)",
				Args:  []interface{}{"John", 30},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt := NewStatement(tt.query, tt.args...)

			if stmt == nil {
				t.Fatal("Expected non-nil Statement")
			}

			if stmt.Query != tt.expected.Query {
				t.Errorf("Expected query %q, got %q", tt.expected.Query, stmt.Query)
			}

			if len(stmt.Args) != len(tt.expected.Args) {
				t.Errorf("Expected %d args, got %d", len(tt.expected.Args), len(stmt.Args))
			}

			for i, arg := range stmt.Args {
				if arg != tt.expected.Args[i] {
					t.Errorf("Expected arg %v at position %d, got %v", tt.expected.Args[i], i, arg)
				}
			}
		})
	}
}
