# SQLMapper

SQLMapper is a powerful SQL schema parser and generator that supports multiple database systems. It can parse SQL dump files and generate schema definitions in a standardized format.

## Features

- Multi-database support:
  - MySQL
  - PostgreSQL
  - SQLite
  - SQL Server
  - Oracle
- Schema parsing and generation
- Support for various SQL objects:
  - Tables
  - Views
  - Functions
  - Procedures
  - Triggers
  - Indexes
  - Sequences

## Development Status

- Basic schema parsing and generation is implemented
- Stream processing feature is under development
  - Basic stream parsing functionality is implemented
  - Tests for stream processing are pending
  - Parallel stream processing is planned
- Documentation will be updated as features are completed

## Installation

```bash
go get github.com/mstgnz/sqlmapper
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/mstgnz/sqlmapper"
)

func main() {
    // Create a new parser for your database type
    parser := sqlmapper.NewParser(sqlmapper.MySQL)
    
    // Parse SQL content
    schema, err := parser.Parse(sqlContent)
    if err != nil {
        panic(err)
    }
    
    // Generate SQL from schema
    sql, err := parser.Generate(schema)
    if err != nil {
        panic(err)
    }
}
```

## Supported SQL Objects

- Tables
  - Columns with data types
  - Primary keys
  - Foreign keys
  - Unique constraints
  - Check constraints
- Views
- Functions
- Procedures
- Triggers
- Indexes
- Sequences

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This work is licensed under the Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License (CC BY-NC-SA 4.0).