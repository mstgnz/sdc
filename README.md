# SDC - SQL Dump Converter

SDC (SQL Dump Converter) is a powerful Go library that allows you to convert SQL dump files between different database systems. This library is particularly useful when you need to migrate a database schema from one system to another.

## Features

### Core Features
- Convert SQL dump files to Go struct format
- Schema conversion between different database systems
- Supported SQL statements:
  - CREATE TABLE
  - ALTER TABLE
  - DROP TABLE
  - CREATE INDEX
  - DROP INDEX
- Advanced parser features:
  - Schema analysis
  - Data type conversions
  - Constraint support
  - Index management

### Advanced Features
- Migration support
  - Automatic migration table creation
  - Migration apply and rollback
  - Migration status tracking
- Schema comparison
  - Table comparison
  - Column comparison
  - Index comparison
  - Constraint comparison
- Extended data type conversions
  - Special conversions for all popular databases
  - Customizable conversion rules
- Database connection management
  - Connection pooling
  - Automatic reconnection
  - Connection status monitoring

## Supported Databases

- [x] MySQL
- [x] PostgreSQL
- [x] SQLite
- [x] Oracle
- [x] SQL Server

## Installation

### As Go Module
```bash
go get -u github.com/mstgnz/sdc
```

### Using Docker
```bash
# Pull Docker image
docker pull mstgnz/sdc:latest

# or run with docker-compose
docker-compose up -d
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/mstgnz/sdc"
    "github.com/mstgnz/sdc/logger"
)

func main() {
    // Create logger
    log := logger.NewLogger(logger.Config{
        Level:  logger.INFO,
        Prefix: "[SDC] ",
    })

    // MySQL dump
    mysqlDump := `CREATE TABLE users (
        id INT PRIMARY KEY AUTO_INCREMENT,
        username VARCHAR(50) NOT NULL UNIQUE,
        email VARCHAR(100) NOT NULL
    );`

    // Create MySQL parser
    parser := sdc.NewMySQLParser()

    // Parse the dump
    entity, err := parser.Parse(mysqlDump)
    if err != nil {
        log.Error("Parse error", map[string]interface{}{
            "error": err.Error(),
        })
        return
    }

    // Convert to PostgreSQL
    pgParser := sdc.NewPostgresParser()
    pgSQL, err := pgParser.Convert(entity)
    if err != nil {
        log.Error("Conversion error", map[string]interface{}{
            "error": err.Error(),
        })
        return
    }

    fmt.Println(pgSQL)
}
```

### Migration Example

```go
package main

import (
    "context"
    "github.com/mstgnz/sdc/migration"
)

func main() {
    // Create migration manager
    manager := migration.NewMigrationManager(driver)

    // Apply migrations
    err := manager.Apply(context.Background())
    if err != nil {
        panic(err)
    }
}
```

### Schema Comparison Example

```go
package main

import (
    "github.com/mstgnz/sdc/schema"
)

func main() {
    // Create schema comparer
    comparer := schema.NewSchemaComparer(sourceTables, targetTables)

    // Find differences
    differences := comparer.Compare()
    for _, diff := range differences {
        fmt.Printf("Type: %s, Name: %s, Change: %s\n", 
            diff.Type, diff.Name, diff.ChangeType)
    }
}
```

## Development

### Requirements

- Go 1.21 or higher
- Docker (optional)

### Testing

```bash
# Run all tests
go test -v ./...

# Run benchmark tests
go test -bench=. ./...
```

### Docker Build

```bash
# Build image
docker build -t sdc .

# Run container
docker run -d sdc
```

## CI/CD

The project includes automated CI/CD pipelines with GitHub Actions:

- Automatic testing and building on every push
- Lint checking on pull requests
- Automatic release creation on tags
- Automatic push to Docker Hub
- Semantic versioning and automatic CHANGELOG

## Contributing

1. Fork this repository
2. Create a new feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push your branch (`git push origin feature/amazing-feature`)
5. Create a Pull Request

## Commit Messages

The project uses semantic versioning. Format your commit messages as follows:

- `feat: ` - New feature
- `fix: ` - Bug fix
- `docs: ` - Documentation changes only
- `style: ` - Code formatting changes
- `refactor: ` - Code refactoring
- `perf: ` - Performance improvements
- `test: ` - Adding or modifying tests
- `chore: ` - General maintenance

## License

This project is licensed under the APACHE License - see the [LICENSE](LICENSE) file for details.