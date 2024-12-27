package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDetectSourceType veritabanı tipini tespit etme fonksiyonunu test eder
func TestDetectSourceType(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "MySQL tespiti",
			content: "CREATE TABLE test (id INT) ENGINE=INNODB;",
			want:    "mysql",
		},
		{
			name:    "SQLite tespiti",
			content: "CREATE TABLE test (id INTEGER AUTOINCREMENT);",
			want:    "sqlite",
		},
		{
			name:    "SQL Server tespiti",
			content: "CREATE TABLE test (id INT IDENTITY(1,1));",
			want:    "sqlserver",
		},
		{
			name:    "PostgreSQL tespiti",
			content: "CREATE TABLE test (id SERIAL PRIMARY KEY);",
			want:    "postgres",
		},
		{
			name:    "Oracle tespiti",
			content: "CREATE TABLE test (id NUMBER(10));",
			want:    "oracle",
		},
		{
			name:    "Bilinmeyen veritabanı",
			content: "CREATE TABLE test (id INT);",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectSourceType(tt.content); got != tt.want {
				t.Errorf("detectSourceType() = %v, beklenilen %v", got, tt.want)
			}
		})
	}
}

// TestCreateOutputPath çıktı dosyası yolu oluşturma fonksiyonunu test eder
func TestCreateOutputPath(t *testing.T) {
	tests := []struct {
		name      string
		inputPath string
		targetDB  string
		want      string
	}{
		{
			name:      "Temel yol",
			inputPath: "test.sql",
			targetDB:  "mysql",
			want:      "test_mysql.sql",
		},
		{
			name:      "Dizinli yol",
			inputPath: "/path/to/test.sql",
			targetDB:  "postgres",
			want:      filepath.Join("/path/to", "test_postgres.sql"),
		},
		{
			name:      "Farklı uzantı",
			inputPath: "dump.txt",
			targetDB:  "sqlite",
			want:      "dump_sqlite.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createOutputPath(tt.inputPath, tt.targetDB); got != tt.want {
				t.Errorf("createOutputPath() = %v, beklenilen %v", got, tt.want)
			}
		})
	}
}

// TestCreateParser parser oluşturma fonksiyonunu test eder
func TestCreateParser(t *testing.T) {
	tests := []struct {
		name    string
		dbType  string
		wantNil bool
	}{
		{
			name:    "MySQL parser",
			dbType:  "mysql",
			wantNil: false,
		},
		{
			name:    "PostgreSQL parser",
			dbType:  "postgres",
			wantNil: false,
		},
		{
			name:    "SQLite parser",
			dbType:  "sqlite",
			wantNil: false,
		},
		{
			name:    "Oracle parser",
			dbType:  "oracle",
			wantNil: false,
		},
		{
			name:    "SQL Server parser",
			dbType:  "sqlserver",
			wantNil: false,
		},
		{
			name:    "Bilinmeyen veritabanı",
			dbType:  "unknown",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createParser(tt.dbType)
			if (got == nil) != tt.wantNil {
				t.Errorf("createParser() nil döndü: %v, beklenilen nil: %v", got == nil, tt.wantNil)
			}
		})
	}
}

// TestIntegration tüm sistemin entegrasyon testini gerçekleştirir
func TestIntegration(t *testing.T) {
	// Test dosyası oluştur
	testSQL := `
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE
);
`
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "test.sql")
	err := os.WriteFile(inputPath, []byte(testSQL), 0644)
	if err != nil {
		t.Fatalf("Test dosyası oluşturulamadı: %v", err)
	}

	// Test senaryoları
	tests := []struct {
		name     string
		targetDB string
		wantErr  bool
	}{
		{
			name:     "MySQL'e dönüştür",
			targetDB: "mysql",
			wantErr:  false,
		},
		{
			name:     "SQLite'a dönüştür",
			targetDB: "sqlite",
			wantErr:  false,
		},
		{
			name:     "Geçersiz hedef",
			targetDB: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test dosyasını oku
			content, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("Test dosyası okunamadı: %v", err)
			}

			// Kaynak tipi tespit et
			sourceType := detectSourceType(string(content))
			if sourceType != "postgres" {
				t.Errorf("Beklenen kaynak tipi postgres, alınan %s", sourceType)
			}

			// Parser'ları oluştur
			sourceParser := createParser(sourceType)
			targetParser := createParser(tt.targetDB)

			if tt.wantErr {
				if targetParser != nil {
					t.Errorf("Geçersiz hedef için nil parser bekleniyordu")
				}
				return
			}

			// Parse et
			schema, err := sourceParser.Parse(string(content))
			if err != nil {
				t.Fatalf("Parse işlemi başarısız: %v", err)
			}

			// Hedef SQL oluştur
			result, err := targetParser.Generate(schema)
			if err != nil {
				t.Fatalf("SQL oluşturma başarısız: %v", err)
			}

			// Sonuç dosyasını oluştur
			outputPath := createOutputPath(inputPath, tt.targetDB)
			err = os.WriteFile(outputPath, []byte(result), 0644)
			if err != nil {
				t.Fatalf("Çıktı dosyası yazılamadı: %v", err)
			}

			// Sonuç dosyasının varlığını kontrol et
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Errorf("Çıktı dosyası oluşturulmadı: %s", outputPath)
			}
		})
	}
}
