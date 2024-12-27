# SQLMapper

SQLMapper is a powerful Go library that enables SQL dump conversion between different database systems. It provides a comprehensive solution for database schema migration, comparison, and conversion tasks.

## Command Line Interface (CLI)

Convert SQL dumps between different database types with a simple command:

```bash
sqlmapper --file=<path_to_sql_file> --to=<target_database>
```

### Installation

```bash
go install github.com/mstgnz/sqlmapper/cmd/sqlmapper@latest
```

### Examples

Convert PostgreSQL dump to MySQL:
```bash
sqlmapper --file=database.sql --to=mysql
# Output: database_mysql.sql
```

Convert MySQL dump to SQLite:
```bash
sqlmapper --file=dump.sql --to=sqlite
# Output: dump_sqlite.sql
```

Convert Oracle dump to PostgreSQL:
```bash
sqlmapper --file=schema.sql --to=postgres
# Output: schema_postgres.sql
```

### Auto-detection

SQLMapper automatically detects the source database type based on SQL syntax patterns:
- MySQL: Detects `ENGINE=INNODB`
- SQLite: Detects `AUTOINCREMENT`
- SQL Server: Detects `IDENTITY`
- PostgreSQL: Detects `SERIAL`
- Oracle: Detects `NUMBER(` syntax

## Library Usage

### Installation

```bash
go get github.com/mstgnz/sqlmapper
```

### Basic Usage

```go
import "github.com/mstgnz/sqlmapper"

// Create a MySQL parser
parser := sqlmapper.NewMySQLParser()

// Parse MySQL dump
entity, err := parser.Parse(mysqlDump)
if err != nil {
    // handle error
}

// Convert to PostgreSQL
pgParser := sqlmapper.NewPostgresParser()
pgSQL, err := pgParser.Convert(entity)
```

### Schema Comparison

```go
import "github.com/mstgnz/sqlmapper/schema"

// Create schema comparer
comparer := schema.NewSchemaComparer(sourceTables, targetTables)

// Find differences
differences := comparer.Compare()
```

## Features

- **Multi-Database Support**: 
  - MySQL
  - PostgreSQL
  - SQLite
  - Oracle
  - SQL Server

- **Core Functionalities**:
  - SQL dump parsing and conversion
  - Schema migration support
  - Database schema comparison
  - Structured logging system
  - Thread-safe operations

## Supported Database Objects

- Tables with columns, indexes, and constraints
- Stored procedures and functions
- Triggers
- Views (including materialized views)
- Sequences
- Extensions
- Permissions
- User-defined types
- Partitions
- Database links
- Tablespaces
- Roles and users
- Clusters
- And more...

## Development

### Prerequisites

- Go 1.16 or higher
- Docker (for running test databases)

### Running Tests

```bash
# Start test databases
docker-compose up -d

# Run tests
go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache License - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions, please open an issue in the GitHub repository.

## Benchmark Results

The following benchmark results show the performance of SQLMapper for different database conversion scenarios. Tests were performed on Apple M1 processor.

```
goos: darwin
goarch: arm64
cpu: Apple M1

BenchmarkMySQLToPostgreSQL-8        2130           2500152 ns/op         6943466 B/op      16046 allocs/op
BenchmarkMySQLToSQLite-8            2101           2135750 ns/op         1656745 B/op       4274 allocs/op
BenchmarkPostgreSQLToMySQL-8        1783           6188440 ns/op         8975810 B/op     138055 allocs/op
BenchmarkSQLiteToMySQL-8          185398              6009 ns/op            9131 B/op        115 allocs/op
```

### Interpretation of Results

1. **MySQL to PostgreSQL**
   - Operations per second: ~2,130
   - Average time per operation: ~2.50ms
   - Memory allocation: ~6.9MB per operation
   - Number of allocations: 16,046 per operation

2. **MySQL to SQLite**
   - Operations per second: ~2,101
   - Average time per operation: ~2.14ms
   - Memory allocation: ~1.7MB per operation
   - Number of allocations: 4,274 per operation

3. **PostgreSQL to MySQL**
   - Operations per second: ~1,783
   - Average time per operation: ~6.19ms
   - Memory allocation: ~9.0MB per operation
   - Number of allocations: 138,055 per operation

4. **SQLite to MySQL** (with simplified schema)
   - Operations per second: ~185,398
   - Average time per operation: ~0.006ms
   - Memory allocation: ~9KB per operation
   - Number of allocations: 115 per operation

### Test Scenarios

1. **Complex Schema Tests** (MySQL to PostgreSQL/SQLite, PostgreSQL to MySQL):
   - Multiple tables with various column types
   - Foreign key constraints
   - Indexes (including fulltext)
   - Views
   - Triggers
   - JSON data type support
   - ENUM types
   - Timestamp auto-update features

2. **Simple Schema Test** (SQLite to MySQL):
   - Basic tables with common data types
   - Simple constraints
   - Basic view
   - No triggers or complex features

Note: The SQLite to MySQL test uses a simplified schema due to SQLite's limited support for complex database features.