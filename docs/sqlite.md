# SQLite User Guide

## Contents
- [Connection Configuration](#connection-configuration)
- [Basic Usage](#basic-usage)
- [Data Types](#data-types)
- [SQLite Features](#sqlite-features)
- [Example Scenarios](#example-scenarios)

## Connection Configuration

You can use the following configuration for SQLite connection:

```go
config := db.Config{
    Driver:   "sqlite3",
    Database: "mydb.sqlite",
}
```

## Basic Usage

To convert SQLite dump file:

```go
// Create SQLite parser
parser := sqlporter.NewSQLiteParser()

// Parse SQLite dump
entity, err := Parse(sqliteDump)
if err != nil {
    log.Error("Parse error", err)
    return
}

// Convert to PostgreSQL
pgParser := sqlporter.NewPostgresParser()
pgSQL, err := pgParser.Convert(entity)
if err != nil {
    log.Error("Conversion error", err)
    return
}
```

## Data Types

Data type mappings when converting from SQLite to other databases:

| SQLite Data Type | MySQL | PostgreSQL | Oracle | SQL Server |
|------------------|-------|------------|---------|------------|
| INTEGER         | INT   | INTEGER    | NUMBER  | INT        |
| REAL            | FLOAT | REAL       | NUMBER  | FLOAT      |
| TEXT            | TEXT  | TEXT       | CLOB    | TEXT       |
| BLOB            | BLOB  | BYTEA      | BLOB    | VARBINARY  |
| NUMERIC         | DECIMAL| NUMERIC    | NUMBER  | DECIMAL    |

## SQLite Features

### Dynamic Type System

SQLite's flexible type system:
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,  -- No need to specify length
    data ANY    -- Can be any data type
);
```

### AUTOINCREMENT

Auto-incrementing fields in SQLite:
```sql
CREATE TABLE products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);
```

### Limitations and Features

SQLite-specific limitations and features:
- Maximum database size: 140 TB
- Maximum number of tables: 2000000
- Maximum row size: 1 GB
- Single file-based database
- No server required
- Zero configuration

## Example Scenarios

### 1. Simple Table Creation

```go
sqliteDump := `CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

// Convert from SQLite to MySQL
mysqlSQL, err := Convert(sqliteDump, "sqlite", "mysql")
```

### 2. Indexes

```go
sqliteDump := `
CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    total REAL NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE INDEX idx_user_id ON orders(user_id);`

// Convert from SQLite to PostgreSQL
pgSQL, err := Convert(sqliteDump, "sqlite", "postgres")
```

### 3. Triggers

```go
sqliteDump := `
CREATE TRIGGER update_timestamp 
AFTER UPDATE ON users
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.id;
END;`

// Convert from SQLite to Oracle
oracleSQL, err := Convert(sqliteDump, "sqlite", "oracle")
```

## Best Practices

### 1. Database Maintenance
- Perform regular VACUUM operations
- Optimize indexes
- Regular backups

### 2. Performance Tips
- Use appropriate indexes
- Execute batch operations within transactions
- Use WAL (Write-Ahead Logging) mode

### 3. Security
- Set proper file permissions
- Use encryption (with SQLCipher)
- Implement input validation 