# SQLPORTER (Parser, Mapper, Converter, Migrator, etc.)

A high-performance SQL parser and converter library written in Go, designed to handle large-scale database schema migrations and transformations.

## Features

- **Multi-Database Support**: Parse and convert SQL schemas between different database systems
  - PostgreSQL
  - MySQL
  - SQLite
  - SQL Server
  - Oracle

- **High Performance**:
  - Memory-optimized processing
  - Concurrent execution with worker pools
  - Streaming support for large files

- **Advanced Parsing**:
  - CREATE TABLE statements
  - ALTER TABLE operations
  - DROP TABLE commands
  - INDEX management
  - CREATE FUNCTION
  - CREATE PROCEDURE
  - CREATE VIEW
  - CREATE SEQUENCE
  - CREATE TYPE
  - CREATE EXTENSION
  - CREATE TRIGGER
  - CREATE INDEX
  - CREATE CONSTRAINT
  - INSERT statements
  - Constraints handling
  - Data type mappings

- **Memory Management**:
  - Efficient buffer pooling
  - Garbage collection optimization
  - Memory usage monitoring
  - Resource cleanup

## Requirements

- Go 1.23 or higher
- No external database dependencies required

## Installation

```bash
go get github.com/mstgnz/sqlporter
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Thanks to all contributors who have helped with the development
- Special thanks to the Go community for their excellent tools and libraries
