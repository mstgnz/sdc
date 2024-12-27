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
- Parallel stream processing for large SQL files
- Support for various SQL objects:
  - Tables
  - Views
  - Functions
  - Procedures
  - Triggers
  - Indexes
  - Sequences

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

### Stream Processing

For large SQL files, you can use the stream processing feature with parallel execution:

```go
package main

import (
    "fmt"
    "github.com/mstgnz/sqlmapper"
)

func main() {
    // Create a stream parser
    parser := sqlmapper.NewStreamParser(sqlmapper.MySQL)
    
    // Process SQL stream with parallel execution
    err := parser.ParseStreamParallel(sqlContent, 4, func(obj interface{}) error {
        switch v := obj.(type) {
        case sqlmapper.Table:
            fmt.Printf("Processed table: %s\n", v.Name)
        case sqlmapper.View:
            fmt.Printf("Processed view: %s\n", v.Name)
        }
        return nil
    })
    
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
  - Indexes
- Views
- Functions
- Stored Procedures
- Triggers
- Sequences
- User-defined Types

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.