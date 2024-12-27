# SQLMapper Troubleshooting Guide

## Table of Contents
1. [Common Error Messages](#common-error-messages)
2. [Database-Specific Issues](#database-specific-issues)
3. [Performance Issues](#performance-issues)
4. [Conversion Problems](#conversion-problems)
5. [CLI Issues](#cli-issues)

## Common Error Messages

### "Invalid SQL Syntax"
**Problem**: SQLMapper fails to parse the input SQL file.
**Solution**:
1. Verify the source database type is correctly detected
2. Check for unsupported SQL features
3. Remove any database-specific comments or directives
4. Split complex statements into simpler ones

### "Unsupported Data Type"
**Problem**: Encountered a data type that doesn't have a direct mapping.
**Solution**:
1. Check the data type mapping documentation
2. Use the `--force-type-mapping` flag with a custom mapping
3. Modify the source SQL to use a supported type
4. Implement a custom type converter

### "File Access Error"
**Problem**: Cannot read input file or write output file.
**Solution**:
1. Verify file permissions
2. Check if the file path is correct
3. Ensure sufficient disk space
4. Run the command with appropriate privileges

## Database-Specific Issues

### MySQL Issues

#### Character Set Conversion
**Problem**: UTF8MB4 character set conversion fails.
**Solution**:
```sql
-- Original MySQL
ALTER TABLE users CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Modified for PostgreSQL
-- Add this before table creation:
SET client_encoding = 'UTF8';
```

#### JSON Column Conversion
**Problem**: JSON columns not properly converted.
**Solution**:
```sql
-- MySQL
ALTER TABLE data MODIFY COLUMN json_col JSON;

-- PostgreSQL equivalent
ALTER TABLE data ALTER COLUMN json_col TYPE JSONB USING json_col::JSONB;

-- SQLite equivalent (stored as TEXT)
ALTER TABLE data ALTER COLUMN json_col TEXT;
```

### PostgreSQL Issues

#### Array Type Conversion
**Problem**: Array types not supported in target database.
**Solution**:
```sql
-- PostgreSQL source
CREATE TABLE items (
    id SERIAL,
    tags TEXT[]
);

-- MySQL conversion
CREATE TABLE items (
    id INT AUTO_INCREMENT,
    tags JSON,
    PRIMARY KEY (id)
);

-- Add migration script:
INSERT INTO items_new (id, tags)
SELECT id, JSON_ARRAY(UNNEST(tags)) FROM items;
```

### Oracle Issues

#### ROWID Handling
**Problem**: ROWID pseudo-column not supported.
**Solution**:
```sql
-- Oracle source
SELECT * FROM employees WHERE ROWID = 'AAASuZAABAAALvVAAA';

-- MySQL conversion
ALTER TABLE employees ADD COLUMN row_identifier BIGINT AUTO_INCREMENT;

-- PostgreSQL conversion
ALTER TABLE employees ADD COLUMN row_identifier BIGSERIAL;
```

## Performance Issues

### Large File Processing

**Problem**: Memory usage spikes with large SQL files.
**Solution**:
1. Use the `--chunk-size` flag to process in smaller batches
2. Enable streaming mode with `--stream`
3. Split large files into smaller ones
4. Use the `--optimize-memory` flag

Example:
```bash
sqlmapper --file=large_dump.sql --to=mysql --chunk-size=1000 --optimize-memory
```

### Slow Conversion Speed

**Problem**: Conversion takes longer than expected.
**Solution**:
1. Enable parallel processing:
```bash
sqlmapper --file=dump.sql --to=postgres --parallel-workers=4
```

2. Optimize input SQL:
```sql
-- Before
SELECT * FROM large_table WHERE complex_condition;

-- After
SELECT needed_columns FROM large_table WHERE simple_condition;
```

## Conversion Problems

### Data Loss Prevention

**Problem**: Potential data truncation during conversion.
**Solution**:
1. Use the `--strict` flag to fail on potential data loss
2. Add data validation queries:

```sql
-- Check for oversized data
SELECT column_name, MAX(LENGTH(column_name)) as max_length
FROM table_name
GROUP BY column_name
HAVING max_length > new_size_limit;
```

### Constraint Violations

**Problem**: Foreign key constraints fail after conversion.
**Solution**:
1. Use `--defer-constraints` flag
2. Add this to your conversion:
```sql
-- At the start
SET FOREIGN_KEY_CHECKS = 0;  -- MySQL
SET CONSTRAINTS ALL DEFERRED;  -- PostgreSQL

-- At the end
SET FOREIGN_KEY_CHECKS = 1;  -- MySQL
SET CONSTRAINTS ALL IMMEDIATE;  -- PostgreSQL
```

## CLI Issues

### Command Line Arguments

**Problem**: Incorrect argument format
**Solution**:
```bash
# Incorrect
sqlmapper -file dump.sql -to mysql

# Correct
sqlmapper --file=dump.sql --to=mysql
```

### Environment Setup

**Problem**: Path or environment issues
**Solution**:
1. Add to PATH:
```bash
export PATH=$PATH:/path/to/sqlmapper/bin
```

2. Set required environment variables:
```bash
export SQLMAPPER_HOME=/path/to/sqlmapper
export SQLMAPPER_CONFIG=/path/to/config.json
```

### Debug Mode

When encountering unexpected issues:
```bash
sqlmapper --file=dump.sql --to=mysql --debug --verbose
```

This will provide:
- Detailed error messages
- Stack traces
- SQL parsing steps
- Conversion process logs

### Common Flags Reference

```bash
--file=<path>           # Input SQL file
--to=<database>         # Target database type
--debug                 # Enable debug mode
--verbose              # Detailed output
--strict               # Strict conversion mode
--force                # Ignore non-critical errors
--dry-run              # Preview conversion
--config=<path>        # Custom config file
--output=<path>        # Output file path
``` 