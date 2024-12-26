# SDC - SQL Dump Converter

SDC (SQL Dump Converter) is a powerful Go library that allows you to convert SQL dump files between different database systems. This library is useful when you need to migrate a database schema from one database system to another.

## Features

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

## Supported Databases

- [x] MySQL
- [x] PostgreSQL
- [x] SQLite
- [x] Oracle
- [x] SQL Server

## Installation

```bash
go get -u github.com/mstgnz/sdc
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/mstgnz/sdc"
)

func main() {
    // Read MySQL dump
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
        panic(err)
    }

    // Convert to PostgreSQL
    pgParser := sdc.NewPostgresParser()
    pgSQL, err := pgParser.Convert(entity)
    if err != nil {
        panic(err)
    }

    fmt.Println(pgSQL)
}
```

### Supported SQL Features

- Table operations
  - Create table (CREATE TABLE)
  - Alter table (ALTER TABLE)
  - Drop table (DROP TABLE)
- Column properties
  - Data types
  - NULL/NOT NULL
  - DEFAULT values
  - AUTO_INCREMENT/IDENTITY
  - Column constraints
- Constraints
  - PRIMARY KEY
  - FOREIGN KEY
  - UNIQUE
  - CHECK
- Indexes
  - Normal indexes
  - Unique indexes
  - Clustered/Non-clustered indexes (SQL Server)

## Contributing

1. Fork this repository
2. Create a new feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push your branch (`git push origin feature/amazing-feature`)
5. Create a Pull Request

## Testing

To run the tests:

```bash
go test ./... -v
```

## License

This project is licensed under the APACHE License - see the [LICENSE](LICENSE) file for details.