package sqlserver

import (
	"errors"
	"strings"

	"github.com/mstgnz/sqlporter"
)

type SQLServer struct {
	schema *sqlporter.SchemaType
}

// NewSQLServer yeni bir SQLServer parser instance'ı oluşturur
func NewSQLServer() *SQLServer {
	return &SQLServer{
		schema: &sqlporter.SchemaType{},
	}
}

// Parse SQLServer dump'ını parse eder
func (s *SQLServer) Parse(content string) (*sqlporter.SchemaType, error) {
	if content == "" {
		return nil, errors.New("boş içerik")
	}

	// TODO: SQLServer dump parsing işlemleri burada yapılacak
	return s.schema, nil
}

// Generate verilen şemadan SQLServer dump'ı oluşturur
func (s *SQLServer) Generate(schema *sqlporter.SchemaType) (string, error) {
	if schema == nil {
		return "", errors.New("boş şema")
	}

	var result strings.Builder

	// TODO: SQLServer dump generation işlemleri burada yapılacak

	return result.String(), nil
}
