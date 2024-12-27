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
  - Batch processing capabilities
  - Streaming support for large files

- **Advanced Parsing**:
  - CREATE TABLE statements
  - ALTER TABLE operations
  - DROP TABLE commands
  - INDEX management
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

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    
    "time"
)

func main() {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    // Initialize parser with memory optimization
    memOptimizer := NewMemoryOptimizer(1024, 0.8) // 1GB max memory
    go memOptimizer.MonitorMemory(ctx)

    // Create PostgreSQL parser
    postgresParser := NewPostgresParser()

    // Parse SQL
    sql := `CREATE TABLE users (
        id INT PRIMARY KEY,
        username VARCHAR(50) NOT NULL,
        email VARCHAR(100) UNIQUE,
        created_at TIMESTAMP
    )`

    entity, err := postgresParser.Parse(sql)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Parsed table: %s\n", entity.Tables[0].Name)
}
```

## Advanced Usage

### Worker Pool

```go
wp := NewWorkerPool(WorkerConfig{
    Workers:      4,
    QueueSize:    1000,
    MemOptimizer: memOptimizer,
    ErrHandler: func(err error) {
        log.Printf("Worker error: %v", err)
    },
})

wp.Start(ctx)
defer wp.Stop()
```

### Batch Processing

```go
bp := NewBatchProcessor(BatchConfig{
    BatchSize:    100,
    Workers:      4,
    Timeout:      30 * time.Second,
    MemOptimizer: memOptimizer,
    ErrorHandler: func(err error) {
        log.Printf("Batch error: %v", err)
    },
})
```

### Stream Processing

```go
sp := NewStreamParser(StreamParserConfig{
    Workers:      4,
    BatchSize:    1024 * 1024, // 1MB
    BufferSize:   32 * 1024,   // 32KB
    Timeout:      30 * time.Second,
    MemOptimizer: memOptimizer,
})
```

## Performance Optimization

The library includes several features for optimizing performance:

1. **Memory Management**:
   - Buffer pooling
   - GC threshold control
   - Memory usage monitoring

2. **Concurrent Processing**:
   - Worker pools
   - Batch processing
   - Stream parsing

3. **Resource Management**:
   - Automatic cleanup
   - Resource pooling
   - Timeout handling

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Thanks to all contributors who have helped with the development
- Special thanks to the Go community for their excellent tools and libraries

## Contact

- GitHub: [@mstgnz](https://github.com/mstgnz)
- Project Link: [https://github.com/mstgnz/sqlporter](https://github.com/mstgnz/sqlporter)