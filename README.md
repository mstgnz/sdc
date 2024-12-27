# SQLPorter

SQLPorter is a powerful Go library that enables SQL dump conversion between different database systems. It provides a comprehensive solution for database schema migration, comparison, and conversion tasks.

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

## Installation

```bash
go get github.com/mstgnz/sqlmapper
```

## Quick Start

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

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions, please open an issue in the GitHub repository.

## Documentation

For detailed documentation and examples, visit our [GitHub repository](https://github.com/mstgnz/sqlmapper).
