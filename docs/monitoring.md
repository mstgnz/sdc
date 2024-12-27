## Monitoring and Logging

### Logging System

SQLMapper provides a detailed logging system for stream processing operations:

```go
// Configure logging with custom options
logger := sqlmapper.NewLogger(sqlmapper.LogConfig{
    Level:        sqlmapper.InfoLevel,
    Format:       sqlmapper.JSONFormat,
    OutputPath:   "/var/log/sqlmapper/stream.log",
    ErrorPath:    "/var/log/sqlmapper/error.log",
    MaxSize:      100, // MB
    MaxBackups:   5,
    MaxAge:       30, // days
    Compress:     true,
})

parser := sqlmapper.NewStreamParser(sqlmapper.MySQL, sqlmapper.WithLogger(logger))
```

### Performance Metrics

Monitor stream processing performance with built-in metrics:

```go
metrics := sqlmapper.NewMetricsCollector()

err := parser.ParseStreamParallel(sqlContent, 4, func(obj interface{}) error {
    metrics.IncrementProcessedObjects()
    metrics.RecordProcessingTime(time.Since(start))
    return nil
}, sqlmapper.WithMetrics(metrics))

// Access metrics
fmt.Printf("Total objects processed: %d\n", metrics.TotalObjects())
fmt.Printf("Average processing time: %v\n", metrics.AverageProcessingTime())
fmt.Printf("Memory usage: %v MB\n", metrics.MemoryUsage())
```

### Available Metrics

1. **Processing Metrics**
   - Total objects processed
   - Objects processed per second
   - Processing time per object
   - Total processing time
   - Failed operations count

2. **Resource Metrics**
   - Memory usage
   - CPU utilization
   - Goroutine count
   - Channel buffer usage

3. **Error Metrics**
   - Error count by type
   - Error rate
   - Retry attempts
   - Recovery success rate

### Monitoring Integration

Integration with popular monitoring systems:

```go
// Prometheus integration
metrics := sqlmapper.NewPrometheusMetrics()
prometheus.MustRegister(metrics)

// Grafana dashboard configuration
grafanaConfig := sqlmapper.GrafanaConfig{
    URL:      "http://localhost:3000",
    APIKey:   os.Getenv("GRAFANA_API_KEY"),
    Dashboard: "sql-mapper-metrics",
}
metrics.EnableGrafana(grafanaConfig)
```

### Log Management

Configure log rotation and management:

```go
logConfig := sqlmapper.LogConfig{
    // Rotation settings
    RotateSize:    100, // MB
    RotateTime:    24,  // hours
    RotateBackups: 7,   // days
    
    // Compression
    CompressRotated: true,
    CompressFormat:  "gz",
    
    // Cleanup
    CleanupAge:     30, // days
    CleanupSize:    1,  // GB
}

logger := sqlmapper.NewLogger(logConfig)
```

### Alert Configuration

Set up alerts for important events:

```go
alerts := sqlmapper.NewAlertManager(sqlmapper.AlertConfig{
    Threshold: sqlmapper.AlertThreshold{
        ErrorRate:      0.01, // 1%
        ProcessingTime: 5 * time.Second,
        MemoryUsage:    80,  // percentage
    },
    Notifications: []sqlmapper.NotificationChannel{
        {Type: "email", Target: "admin@example.com"},
        {Type: "slack", Target: "monitoring-channel"},
    },
})

parser.SetAlertManager(alerts)
```

### Best Practices for Monitoring

1. **Log Levels**
   - Use appropriate log levels (DEBUG, INFO, WARN, ERROR)
   - Configure different outputs for different levels
   - Include contextual information in logs

2. **Metric Collection**
   - Collect metrics at appropriate intervals
   - Use aggregation for long-running processes
   - Monitor trends over time

3. **Resource Management**
   - Set up log rotation to prevent disk space issues
   - Archive old logs automatically
   - Monitor resource usage trends

4. **Alert Management**
   - Configure meaningful alert thresholds
   - Avoid alert fatigue with proper filtering
   - Set up escalation policies 