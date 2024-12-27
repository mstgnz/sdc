package sqlserver

import (
	"testing"

	"github.com/mstgnz/sqlporter"
)

func TestSQLServer_Parse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "Boş içerik",
			content: "",
			wantErr: true,
		},
		// TODO: Daha fazla test senaryosu eklenecek
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSQLServer()
			_, err := s.Parse(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSQLServer_Generate(t *testing.T) {
	tests := []struct {
		name    string
		schema  *sqlporter.SchemaType
		wantErr bool
	}{
		{
			name:    "Boş şema",
			schema:  nil,
			wantErr: true,
		},
		// TODO: Daha fazla test senaryosu eklenecek
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSQLServer()
			_, err := s.Generate(tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
