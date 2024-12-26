# API Documentation

## Contents
- [Parser API](#parser-api)
- [Converter API](#converter-api)
- [Migration API](#migration-api)
- [Database API](#database-api)
- [Helper Functions](#helper-functions)

## Parser API

### NewParser

Creates a new parser.

```go
func NewParser(dialect string) (Parser, error)
```

**Parameters:**
- `dialect`: Database dialect ("mysql", "postgres", "sqlite", "oracle", "sqlserver")

**Return Values:**
- `Parser`: Object implementing the Parser interface
- `error`: In case of error

### Parse

Parses SQL dump file.

```go
func (p *Parser) Parse(sql string) (*Entity, error)
```

**Parameters:**
- `sql`: SQL dump text to parse

**Return Values:**
- `*Entity`: Parsed data structure
- `error`: In case of error

## Converter API

### Convert

Converts a database schema to another database format.

```go
func Convert(sql, sourceDialect, targetDialect string) (string, error)
```

**Parameters:**
- `sql`: SQL text to convert
- `sourceDialect`: Source database dialect
- `targetDialect`: Target database dialect

**Return Values:**
- `string`: Converted SQL text
- `error`: In case of error

### ConvertEntity

Converts Entity structure to SQL.

```go
func (p *Parser) ConvertEntity(entity *Entity) (string, error)
```

**Parameters:**
- `entity`: Entity structure to convert

**Return Values:**
- `string`: Generated SQL text
- `error`: In case of error

## Migration API

### NewMigrationManager

Creates a new migration manager.

```go
func NewMigrationManager(db *sql.DB) *MigrationManager
```

**Parameters:**
- `db`: Database connection

### Apply

Applies migrations.

```go
func (m *MigrationManager) Apply(ctx context.Context) error
```

**Parameters:**
- `ctx`: Context object

**Return Values:**
- `error`: In case of error

### Rollback

Rolls back the last migration.

```go
func (m *MigrationManager) Rollback(ctx context.Context) error
```

**Parameters:**
- `ctx`: Context object

**Return Values:**
- `error`: In case of error

## Database API

### NewConnection

Creates a new database connection.

```go
func NewConnection(config *Config) (*sql.DB, error)
```

**Parameters:**
- `config`: Database configuration

**Return Values:**
- `*sql.DB`: Database connection
- `error`: In case of error

### Config Structure

```go
type Config struct {
    Driver   string
    Host     string
    Port     int
    Database string
    Username string
    Password string
    SSLMode  string
    Options  map[string]string
}
```

## Helper Functions

### ValidateSQL

Validates SQL text.

```go
func ValidateSQL(sql string) error
```

**Parameters:**
- `sql`: SQL text to validate

**Return Values:**
- `error`: In case of error

### FormatSQL

Formats SQL text.

```go
func FormatSQL(sql string) string
```

**Parameters:**
- `sql`: SQL text to format

**Return Values:**
- `string`: Formatted SQL text

## Error Types

```go
var (
    ErrInvalidDialect    = errors.New("invalid database dialect")
    ErrParseFailure      = errors.New("SQL parse error")
    ErrConversionFailure = errors.New("conversion error")
    ErrConnectionFailure = errors.New("connection error")
    ErrMigrationFailure  = errors.New("migration error")
)
```

## Examples

### Simple Conversion

```go
package main

import (
    "github.com/mstgnz/sdc"
    "log"
)

func main() {
    mysqlSQL := `CREATE TABLE users (
        id INT PRIMARY KEY AUTO_INCREMENT,
        name VARCHAR(100) NOT NULL
    );`

    // Convert from MySQL to PostgreSQL
    pgSQL, err := sdc.Convert(mysqlSQL, "mysql", "postgres")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println(pgSQL)
}
```

### Migration Usage

```go
package main

import (
    "context"
    "github.com/mstgnz/sdc/migration"
    "log"
)

func main() {
    // Create migration manager
    manager := migration.NewMigrationManager(db)

    // Apply migrations
    err := manager.Apply(context.Background())
    if err != nil {
        log.Fatal(err)
    }
} 