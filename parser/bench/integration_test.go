package bench

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/mstgnz/sqlporter/parser"
)

type testDB struct {
	db      *sql.DB
	driver  string
	cleanup func()
	version *parser.Version
}

func setupTestDatabases(t *testing.T) map[parser.DataType]testDB {
	dbs := make(map[parser.DataType]testDB)

	// MySQL test database
	if mysqlURL := os.Getenv("TEST_MYSQL_URL"); mysqlURL != "" {
		db, err := sql.Open("mysql", mysqlURL)
		if err != nil {
			t.Fatalf("failed to connect to MySQL: %v", err)
		}
		dbs[parser.TypeMySQL] = testDB{
			db:      db,
			driver:  "mysql",
			version: &parser.Version{Major: 5, Minor: 7},
			cleanup: func() {
				db.Close()
			},
		}
	}

	// PostgreSQL test database
	if pgURL := os.Getenv("TEST_POSTGRES_URL"); pgURL != "" {
		db, err := sql.Open("postgres", pgURL)
		if err != nil {
			t.Fatalf("failed to connect to PostgreSQL: %v", err)
		}
		dbs[parser.TypePostgreSQL] = testDB{
			db:      db,
			driver:  "postgres",
			version: &parser.Version{Major: 12},
			cleanup: func() {
				db.Close()
			},
		}
	}

	return dbs
}

func TestIntegration_DataTypeConversion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbs := setupTestDatabases(t)
	defer func() {
		for _, db := range dbs {
			db.cleanup()
		}
	}()

	tests := []struct {
		name       string
		sourceType parser.DataType
		targetType parser.DataType
		setup      func(*sql.DB) error
		verify     func(*sql.DB) error
	}{
		{
			name:       "MySQL to PostgreSQL integer conversion",
			sourceType: parser.TypeMySQL,
			targetType: parser.TypePostgreSQL,
			setup: func(db *sql.DB) error {
				_, err := db.Exec(`
					CREATE TABLE test_int (
						id INT PRIMARY KEY,
						value INT
					)
				`)
				if err != nil {
					return err
				}
				_, err = db.Exec("INSERT INTO test_int VALUES (1, 42)")
				return err
			},
			verify: func(db *sql.DB) error {
				var value int
				err := db.QueryRow("SELECT value FROM test_int WHERE id = 1").Scan(&value)
				if err != nil {
					return err
				}
				if value != 42 {
					t.Errorf("expected 42, got %d", value)
				}
				return nil
			},
		},
		{
			name:       "MySQL to PostgreSQL timestamp conversion",
			sourceType: parser.TypeMySQL,
			targetType: parser.TypePostgreSQL,
			setup: func(db *sql.DB) error {
				_, err := db.Exec(`
					CREATE TABLE test_time (
						id INT PRIMARY KEY,
						created_at DATETIME
					)
				`)
				if err != nil {
					return err
				}
				_, err = db.Exec("INSERT INTO test_time VALUES (1, '2023-01-01 12:00:00')")
				return err
			},
			verify: func(db *sql.DB) error {
				var createdAt time.Time
				err := db.QueryRow("SELECT created_at FROM test_time WHERE id = 1").Scan(&createdAt)
				if err != nil {
					return err
				}
				expected := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
				if !createdAt.Equal(expected) {
					t.Errorf("expected %v, got %v", expected, createdAt)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceDB, ok := dbs[tt.sourceType]
			if !ok {
				t.Skipf("source database %s not configured", tt.sourceType)
			}

			targetDB, ok := dbs[tt.targetType]
			if !ok {
				t.Skipf("target database %s not configured", tt.targetType)
			}

			if err := tt.setup(sourceDB.db); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			converter := parser.NewConverter()
			converter.RegisterDefaultMappings()

			if err := tt.verify(targetDB.db); err != nil {
				t.Errorf("verification failed: %v", err)
			}
		})
	}
}

func TestIntegration_CharSetConversion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbs := setupTestDatabases(t)
	defer func() {
		for _, db := range dbs {
			db.cleanup()
		}
	}()

	tests := []struct {
		name       string
		sourceType parser.DataType
		targetType parser.DataType
		charset    string
		text       string
	}{
		{
			name:       "UTF8MB4 text conversion MySQL to PostgreSQL",
			sourceType: parser.TypeMySQL,
			targetType: parser.TypePostgreSQL,
			charset:    "utf8mb4",
			text:       "Hello, 世界! 🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceDB, ok := dbs[tt.sourceType]
			if !ok {
				t.Skipf("source database %s not configured", tt.sourceType)
			}

			targetDB, ok := dbs[tt.targetType]
			if !ok {
				t.Skipf("target database %s not configured", tt.targetType)
			}

			// Create test table in source database
			_, err := sourceDB.db.Exec(`
				CREATE TABLE test_text (
					id INT PRIMARY KEY,
					content TEXT CHARACTER SET utf8mb4
				)
			`)
			if err != nil {
				t.Fatalf("failed to create source table: %v", err)
			}

			// Insert test data
			_, err = sourceDB.db.Exec("INSERT INTO test_text VALUES (1, ?)", tt.text)
			if err != nil {
				t.Fatalf("failed to insert test data: %v", err)
			}

			// Create test table in target database
			_, err = targetDB.db.Exec(`
				CREATE TABLE test_text (
					id INTEGER PRIMARY KEY,
					content TEXT
				)
			`)
			if err != nil {
				t.Fatalf("failed to create target table: %v", err)
			}

			// Verify data
			var content string
			err = targetDB.db.QueryRow("SELECT content FROM test_text WHERE id = 1").Scan(&content)
			if err != nil {
				t.Fatalf("failed to read test data: %v", err)
			}

			if content != tt.text {
				t.Errorf("expected %q, got %q", tt.text, content)
			}
		})
	}
}
