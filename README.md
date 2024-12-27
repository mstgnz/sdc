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