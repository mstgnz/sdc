package parser

import (
	"testing"
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
