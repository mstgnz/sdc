package parser

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

var sampleSQL = `
CREATE TABLE users (
	id INT PRIMARY KEY AUTO_INCREMENT,
	username VARCHAR(50) NOT NULL UNIQUE,
	email VARCHAR(100) NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	status ENUM('active', 'inactive', 'suspended') DEFAULT 'active',
	CONSTRAINT fk_role FOREIGN KEY (role_id) REFERENCES roles(id)
);

CREATE INDEX idx_username ON users(username);
CREATE UNIQUE INDEX idx_email ON users(email);

ALTER TABLE users ADD COLUMN last_login TIMESTAMP NULL;
ALTER TABLE users ADD CONSTRAINT chk_email CHECK (email LIKE '%@%.%');
`

func BenchmarkMySQLParser_Parse(b *testing.B) {
	parser := &MySQLParser{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(sampleSQL)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPostgresParser_Parse(b *testing.B) {
	parser := &PostgresParser{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(sampleSQL)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSQLiteParser_Parse(b *testing.B) {
	parser := &SQLiteParser{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(sampleSQL)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOracleParser_Parse(b *testing.B) {
	parser := &OracleParser{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(sampleSQL)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSQLServerParser_Parse(b *testing.B) {
	parser := &SQLServerParser{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(sampleSQL)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Conversion Benchmarks
func BenchmarkMySQLToPostgres_Convert(b *testing.B) {
	mysqlParser := &MySQLParser{}
	pgParser := &PostgresParser{}

	entity, err := mysqlParser.Parse(sampleSQL)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pgParser.Convert(entity)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMySQLToSQLite_Convert(b *testing.B) {
	mysqlParser := &MySQLParser{}
	sqliteParser := &SQLiteParser{}

	entity, err := mysqlParser.Parse(sampleSQL)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sqliteParser.Convert(entity)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPostgresToMySQL_Convert(b *testing.B) {
	pgParser := &PostgresParser{}
	mysqlParser := &MySQLParser{}

	entity, err := pgParser.Parse(sampleSQL)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mysqlParser.Convert(entity)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Memory Allocation Benchmarks
func BenchmarkMySQLParser_Memory(b *testing.B) {
	parser := &MySQLParser{}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(sampleSQL)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPostgresParser_Memory(b *testing.B) {
	parser := &PostgresParser{}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(sampleSQL)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Parallel Benchmarks
func BenchmarkMySQLParser_Parallel(b *testing.B) {
	parser := &MySQLParser{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := parser.Parse(sampleSQL)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkPostgresParser_Parallel(b *testing.B) {
	parser := &PostgresParser{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := parser.Parse(sampleSQL)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkConverter_ConvertType(b *testing.B) {
	converter := NewConverter()
	converter.RegisterDefaultMappings()

	sourceVer := &Version{Major: 5, Minor: 7}
	targetVer := &Version{Major: 12}

	benchmarks := []struct {
		name       string
		sourceType string
		targetType string
		value      interface{}
	}{
		{
			name:       "Integer conversion",
			sourceType: "int",
			targetType: "integer",
			value:      42,
		},
		{
			name:       "String conversion",
			sourceType: "varchar",
			targetType: "character varying",
			value:      "test string",
		},
		{
			name:       "DateTime conversion",
			sourceType: "datetime",
			targetType: "timestamp",
			value:      "2023-01-01 12:00:00",
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := converter.ConvertType(bm.value, bm.sourceType, bm.targetType, sourceVer, targetVer)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkStreamParser_Parse(b *testing.B) {
	parser := NewStreamParser(StreamParserConfig{
		Workers:    4,
		BatchSize:  1024 * 1024,
		BufferSize: 32 * 1024,
	})

	// Generate test SQL data
	sql := `
		CREATE TABLE users (
			id INT PRIMARY KEY,
			name VARCHAR(100),
			email VARCHAR(255),
			created_at DATETIME
		);

		INSERT INTO users VALUES (1, 'John Doe', 'john@example.com', '2023-01-01 12:00:00');
		INSERT INTO users VALUES (2, 'Jane Doe', 'jane@example.com', '2023-01-01 12:00:00');
		INSERT INTO users VALUES (3, 'Bob Smith', 'bob@example.com', '2023-01-01 12:00:00');
	`

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := parser.ParseStream(context.Background(), strings.NewReader(sql))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBatchProcessor_Process(b *testing.B) {
	processor := NewBatchProcessor(BatchConfig{
		BatchSize: 1000,
		Workers:   4,
	})

	// Generate test statements
	stmts := make([]*Statement, 1000)
	for i := range stmts {
		stmts[i] = NewStatement(
			"INSERT INTO users VALUES (?, ?, ?, ?)",
			i,
			fmt.Sprintf("User %d", i),
			fmt.Sprintf("user%d@example.com", i),
			time.Now(),
		)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := processor.ProcessBatch(context.Background(), stmts, func(stmt *Statement) error {
			return nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCharSetConversion(b *testing.B) {
	converter := NewConverter()
	converter.RegisterDefaultMappings()

	text := "Hello, 世界! 🌍" // Mixed ASCII, Unicode, and Emoji

	b.Run("UTF8MB4 Validation", func(b *testing.B) {
		charset, err := converter.GetCharSet("utf8mb4")
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			if len(text) > charset.MaxLength*4 { // UTF8MB4 max 4 bytes per character
				b.Fatal("text too long for charset")
			}
		}
	})

	b.Run("Charset Lookup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := converter.GetCharSet("utf8mb4")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
