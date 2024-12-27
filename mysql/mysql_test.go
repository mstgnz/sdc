package mysql

import (
	"testing"

	"github.com/mstgnz/sqlporter"
)

func TestMySQL_Parse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "Empty content",
			content: "",
			wantErr: true,
		},
		// TODO: More test cases will be added
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMySQL()
			_, err := m.Parse(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMySQL_Generate(t *testing.T) {
	tests := []struct {
		name    string
		schema  *sqlporter.Schema
		wantErr bool
	}{
		{
			name:    "Empty schema",
			schema:  nil,
			wantErr: true,
		},
		// TODO: More test cases will be added
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMySQL()
			_, err := m.Generate(tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
