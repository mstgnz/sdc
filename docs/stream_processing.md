# Stream Processing

SQLMapper provides stream processing capabilities for handling large SQL files efficiently. This feature allows you to process SQL dumps in parallel, improving performance for large datasets.

## Overview

The stream processing feature:
- Processes SQL statements in parallel
- Uses worker pools for efficient resource utilization
- Supports all major database types
- Provides real-time callback for processed objects
- Handles large SQL files without loading entire content into memory

## Usage

### Basic Stream Processing

```go
parser := sqlmapper.NewStreamParser(sqlmapper.MySQL)

err := parser.ParseStream(sqlContent, func(obj interface{}) error {
    switch v := obj.(type) {
    case sqlmapper.Table:
        // Handle table
    case sqlmapper.View:
        // Handle view
    case sqlmapper.Function:
        // Handle function
    }
    return nil
})
```

### Parallel Stream Processing

```go
parser := sqlmapper.NewStreamParser(sqlmapper.MySQL)

// Process with 4 workers
err := parser.ParseStreamParallel(sqlContent, 4, func(obj interface{}) error {
    switch v := obj.(type) {
    case sqlmapper.Table:
        fmt.Printf("Processing table: %s\n", v.Name)
    case sqlmapper.View:
        fmt.Printf("Processing view: %s\n", v.Name)
    }
    return nil
})
```

## Configuration

### Worker Pool Size

The number of workers can be configured based on your system's capabilities:

```go
// For CPU-bound tasks
workers := runtime.NumCPU()

// For I/O-bound tasks
workers := runtime.NumCPU() * 2
```

### Supported Object Types

The stream processor can handle various SQL objects:

- Tables
- Views
- Functions
- Procedures
- Triggers
- Indexes
- Sequences

Each object is passed to the callback function as it's processed.

## Error Handling

Errors during stream processing are handled gracefully:

```go
err := parser.ParseStreamParallel(sqlContent, 4, func(obj interface{}) error {
    if err := processObject(obj); err != nil {
        return fmt.Errorf("failed to process object: %v", err)
    }
    return nil
})

if err != nil {
    log.Printf("Stream processing error: %v", err)
}
```

## Best Practices

1. **Worker Pool Size**
   - Start with number of CPU cores
   - Adjust based on performance monitoring
   - Consider memory constraints

2. **Error Handling**
   - Always check for errors in callback
   - Log errors appropriately
   - Consider implementing retry logic

3. **Memory Management**
   - Process objects as they come
   - Avoid storing large amounts of data
   - Use channels for communication

4. **Performance Optimization**
   - Monitor CPU and memory usage
   - Adjust worker count based on system load
   - Consider batch processing for small objects

## Examples

### Processing with Progress Tracking

```go
var processed int64

err := parser.ParseStreamParallel(sqlContent, 4, func(obj interface{}) error {
    atomic.AddInt64(&processed, 1)
    fmt.Printf("\rProcessed objects: %d", processed)
    return nil
})
```

### Custom Object Processing

```go
err := parser.ParseStreamParallel(sqlContent, 4, func(obj interface{}) error {
    switch v := obj.(type) {
    case sqlmapper.Table:
        return processTable(v)
    case sqlmapper.View:
        return processView(v)
    case sqlmapper.Function:
        return processFunction(v)
    default:
        return fmt.Errorf("unsupported object type")
    }
})
```

### Error Recovery

```go
err := parser.ParseStreamParallel(sqlContent, 4, func(obj interface{}) error {
    for attempts := 0; attempts < 3; attempts++ {
        if err := processObject(obj); err != nil {
            if attempts == 2 {
                return err
            }
            continue
        }
        break
    }
    return nil
})
```

## Limitations

- Maximum file size depends on available system memory
- Some complex SQL statements might require sequential processing
- Performance depends on system resources and SQL complexity

## Performance Considerations

- Worker pool size affects memory usage
- Large SQL files might require batching
- Consider network I/O for database operations
- Monitor system resources during processing 