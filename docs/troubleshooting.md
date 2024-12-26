# Troubleshooting Guide

This guide helps you diagnose and fix common issues you might encounter while using SDC.

## Table of Contents

1. [Common Issues](#common-issues)
2. [Error Messages](#error-messages)
3. [Performance Issues](#performance-issues)
4. [Known Limitations](#known-limitations)

## Common Issues

### Connection Issues

#### Problem: Unable to connect to database
```
error: failed to connect to database: connection refused
```

**Possible causes:**
- Database server is not running
- Wrong connection credentials
- Firewall blocking the connection
- Wrong port number

**Solutions:**
1. Verify database server is running
2. Check credentials in configuration
3. Check firewall settings
4. Verify correct port number

### Parsing Issues

#### Problem: Invalid SQL syntax
```
error: failed to parse SQL: syntax error near 'TABLE'
```

**Possible causes:**
- Unsupported SQL syntax
- Malformed SQL statement
- Incorrect database dialect selected

**Solutions:**
1. Verify SQL syntax is supported
2. Check SQL statement format
3. Ensure correct parser is being used

### Migration Issues

#### Problem: Migration fails to apply
```
error: failed to apply migration: table already exists
```

**Possible causes:**
- Migration already applied
- Conflicting table names
- Insufficient permissions

**Solutions:**
1. Check migration status
2. Verify table names
3. Check database permissions

## Error Messages

### Connection Errors
- `ConnectionError: connection refused`: Database server is unreachable
- `ConnectionError: invalid credentials`: Wrong username or password
- `ConnectionError: database not found`: Database does not exist

### Parser Errors
- `ParserError: unsupported syntax`: SQL syntax not supported
- `ParserError: invalid type`: Data type not recognized
- `ParserError: missing required field`: Required field not provided

### Migration Errors
- `MigrationError: version conflict`: Migration version mismatch
- `MigrationError: failed to rollback`: Rollback operation failed
- `MigrationError: invalid state`: Migration in invalid state

## Performance Issues

### Slow Parsing

**Symptoms:**
- Long processing times for SQL files
- High memory usage

**Solutions:**
1. Break large SQL files into smaller chunks
2. Use streaming parser for large files
3. Optimize memory settings

### Memory Usage

**Symptoms:**
- Out of memory errors
- High memory consumption

**Solutions:**
1. Enable garbage collection
2. Use batch processing
3. Implement pagination

## Known Limitations

### SQL Support
- Complex stored procedures not supported
- Limited support for vendor-specific features
- Some advanced index types not supported

### Data Types
- Custom data types require manual mapping
- Some complex types may lose precision
- BLOB/CLOB size limitations

### Performance
- Large file processing may be slow
- Memory usage scales with file size
- Concurrent operations may be limited

## Best Practices

1. **Testing**
   - Always test conversions in development first
   - Use small datasets for initial testing
   - Maintain test coverage for custom conversions

2. **Backup**
   - Always backup database before migrations
   - Keep SQL dumps of original schema
   - Document all custom configurations

3. **Monitoring**
   - Enable logging for troubleshooting
   - Monitor memory usage
   - Track conversion times

## Getting Help

If you encounter issues not covered in this guide:

1. Check the [GitHub Issues](https://github.com/mstgnz/sdc/issues)
2. Search the [Documentation](https://github.com/mstgnz/sdc/docs)
3. Create a new issue with:
   - Error message
   - Steps to reproduce
   - Environment details
   - Sample code 