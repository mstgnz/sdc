package postgres

import (
	"errors"
	"strings"

	"github.com/mstgnz/sqlporter"
)

type PostgreSQL struct {
	schema *sqlporter.SchemaType
}

// NewPostgreSQL yeni bir PostgreSQL parser instance'ı oluşturur
func NewPostgreSQL() *PostgreSQL {
	return &PostgreSQL{
		schema: &sqlporter.SchemaType{},
	}
}

// Parse PostgreSQL dump'ını parse eder
func (p *PostgreSQL) Parse(content string) (*sqlporter.SchemaType, error) {
	if content == "" {
		return nil, errors.New("boş içerik")
	}

	// TODO: PostgreSQL dump parsing işlemleri burada yapılacak
	return p.schema, nil
}

// Generate verilen şemadan PostgreSQL dump'ı oluşturur
func (p *PostgreSQL) Generate(schema *sqlporter.SchemaType) (string, error) {
	if schema == nil {
		return "", errors.New("boş şema")
	}

	var result strings.Builder

	// TODO: PostgreSQL dump generation işlemleri burada yapılacak

	return result.String(), nil
}
