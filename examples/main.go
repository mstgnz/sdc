package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mstgnz/sqlmapper/mysql"
	"github.com/mstgnz/sqlmapper/oracle"
	"github.com/mstgnz/sqlmapper/postgres"
	"github.com/mstgnz/sqlmapper/sqlite"
	"github.com/mstgnz/sqlmapper/sqlserver"
)

func main() {
	// Run example conversions
	postgresqlToMysql()
	mysqlToPostgresql()
	oracleToMysql()
	sqlserverToPostgresql()
	sqliteToOracle()
}

// PostgreSQL -> MySQL conversion
func postgresqlToMysql() {
	fmt.Println("\n=== PostgreSQL -> MySQL Conversion ===")

	// Read PostgreSQL dump file
	pgDump, err := os.ReadFile("examples/files/postgres.sql")
	if err != nil {
		log.Fatalf("Failed to read PostgreSQL dump: %v", err)
	}

	// Create PostgreSQL parser
	pgParser := postgres.NewPostgreSQL()

	// Parse PostgreSQL dump
	entity, err := pgParser.Parse(string(pgDump))
	if err != nil {
		log.Fatalf("PostgreSQL parse error: %v", err)
	}

	// Create MySQL parser
	mysqlParser := mysql.NewMySQL()

	// Generate MySQL format from entity
	mysqlDump, err := mysqlParser.Generate(entity)
	if err != nil {
		log.Fatalf("MySQL generate error: %v", err)
	}

	// Write result to file
	err = os.WriteFile("examples/files/output/postgres_to_mysql.sql", []byte(mysqlDump), 0644)
	if err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	fmt.Println("Conversion completed: examples/files/output/postgres_to_mysql.sql")
}

// MySQL -> PostgreSQL conversion
func mysqlToPostgresql() {
	fmt.Println("\n=== MySQL -> PostgreSQL Conversion ===")

	// Read MySQL dump file
	mysqlDump, err := os.ReadFile("examples/files/mysql.sql")
	if err != nil {
		log.Fatalf("Failed to read MySQL dump: %v", err)
	}

	// Create MySQL parser
	mysqlParser := mysql.NewMySQL()

	// Parse MySQL dump
	entity, err := mysqlParser.Parse(string(mysqlDump))
	if err != nil {
		log.Fatalf("MySQL parse error: %v", err)
	}

	// Create PostgreSQL parser
	pgParser := postgres.NewPostgreSQL()

	// Generate PostgreSQL format from entity
	pgDump, err := pgParser.Generate(entity)
	if err != nil {
		log.Fatalf("PostgreSQL generate error: %v", err)
	}

	// Write result to file
	err = os.WriteFile("examples/files/output/mysql_to_postgres.sql", []byte(pgDump), 0644)
	if err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	fmt.Println("Conversion completed: examples/files/output/mysql_to_postgres.sql")
}

// Oracle -> MySQL conversion
func oracleToMysql() {
	fmt.Println("\n=== Oracle -> MySQL Conversion ===")

	// Read Oracle dump file
	oracleDump, err := os.ReadFile("examples/files/oracle.sql")
	if err != nil {
		log.Fatalf("Failed to read Oracle dump: %v", err)
	}

	// Create Oracle parser
	oracleParser := oracle.NewOracle()

	// Parse Oracle dump
	entity, err := oracleParser.Parse(string(oracleDump))
	if err != nil {
		log.Fatalf("Oracle parse error: %v", err)
	}

	// Create MySQL parser
	mysqlParser := mysql.NewMySQL()

	// Generate MySQL format from entity
	mysqlDump, err := mysqlParser.Generate(entity)
	if err != nil {
		log.Fatalf("MySQL generate error: %v", err)
	}

	// Write result to file
	err = os.WriteFile("examples/files/output/oracle_to_mysql.sql", []byte(mysqlDump), 0644)
	if err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	fmt.Println("Conversion completed: examples/files/output/oracle_to_mysql.sql")
}

// SQL Server -> PostgreSQL conversion
func sqlserverToPostgresql() {
	fmt.Println("\n=== SQL Server -> PostgreSQL Conversion ===")

	// Read SQL Server dump file
	sqlserverDump, err := os.ReadFile("examples/files/sqlserver.sql")
	if err != nil {
		log.Fatalf("Failed to read SQL Server dump: %v", err)
	}

	// Create SQL Server parser
	sqlserverParser := sqlserver.NewSQLServer()

	// Parse SQL Server dump
	entity, err := sqlserverParser.Parse(string(sqlserverDump))
	if err != nil {
		log.Fatalf("SQL Server parse error: %v", err)
	}

	// Create PostgreSQL parser
	pgParser := postgres.NewPostgreSQL()

	// Generate PostgreSQL format from entity
	pgDump, err := pgParser.Generate(entity)
	if err != nil {
		log.Fatalf("PostgreSQL generate error: %v", err)
	}

	// Write result to file
	err = os.WriteFile("examples/files/output/sqlserver_to_postgres.sql", []byte(pgDump), 0644)
	if err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	fmt.Println("Conversion completed: examples/files/output/sqlserver_to_postgres.sql")
}

// SQLite -> Oracle conversion
func sqliteToOracle() {
	fmt.Println("\n=== SQLite -> Oracle Conversion ===")

	// Read SQLite dump file
	sqliteDump, err := os.ReadFile("examples/files/sqlite.sql")
	if err != nil {
		log.Fatalf("Failed to read SQLite dump: %v", err)
	}

	// Create SQLite parser
	sqliteParser := sqlite.NewSQLite()

	// Parse SQLite dump
	entity, err := sqliteParser.Parse(string(sqliteDump))
	if err != nil {
		log.Fatalf("SQLite parse error: %v", err)
	}

	// Create Oracle parser
	oracleParser := oracle.NewOracle()

	// Generate Oracle format from entity
	oracleDump, err := oracleParser.Generate(entity)
	if err != nil {
		log.Fatalf("Oracle generate error: %v", err)
	}

	// Write result to file
	err = os.WriteFile("examples/files/output/sqlite_to_oracle.sql", []byte(oracleDump), 0644)
	if err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	fmt.Println("Conversion completed: examples/files/output/sqlite_to_oracle.sql")
}

func init() {
	// Create output directory
	outputDir := "examples/files/output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
}
