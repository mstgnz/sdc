# Troubleshooting Guide

This guide helps you diagnose and fix common issues you might encounter while using SDC.

## Table of Contents

1. [Common Issues](#common-issues)
2. [Error Messages](#error-messages)
3. [Performance Issues](#performance-issues)
4. [Known Limitations](#known-limitations)
5. [Database-Specific Issues](#database-specific-issues)
6. [Security Issues](#security-issues)
7. [Best Practices](#best-practices)

## Common Issues

### Connection Issues

#### Problem: Unable to connect to database
```
error: failed to connect to database: connection refused
```

**Possible causes:**
- Database server is not running
- Incorrect connection credentials
- Firewall blocking the connection
- Wrong port number
- SSL/TLS configuration issues

**Solutions:**
1. Verify database server is running
2. Check connection credentials
3. Check firewall settings
4. Verify port number
5. Check SSL/TLS certificates

### Parsing Issues

#### Problem: Invalid SQL syntax
```
error: failed to parse SQL: syntax error near 'TABLE'
```

**Possible causes:**
- Unsupported SQL syntax
- Malformed SQL statement
- Incorrect database dialect selected
- Character encoding issues

**Solutions:**
1. Verify SQL syntax is supported
2. Check SQL statement format
3. Ensure correct parser is being used
4. Check character encoding

### Migration Issues

#### Problem: Migration fails to apply
```
error: failed to apply migration: table already exists
```

**Possible causes:**
- Migration already applied
- Conflicting table names
- Insufficient permissions
- Data inconsistency

**Solutions:**
1. Check migration status
2. Verify table names
3. Check database permissions
4. Verify data integrity

## Database-Specific Issues

### MySQL Issues

#### Problem: Character set mismatch
```
error: Incorrect string value: '\xF0\x9F\x98\x83' for column
```

**Solution:**
1. Set database and table character set to UTF8MB4
2. Add `charset=utf8mb4` to connection URL

### PostgreSQL Issues

#### Problem: SSL connection error
```
error: SSL is not enabled on the server
```

**Solution:**
1. Set `ssl = on` in postgresql.conf
2. Place SSL certificates in correct location

## Security Issues

### SSL/TLS Configuration

#### Problem: Insecure connection
```
error: server does not support SSL, but SSL was required
```

**Solution:**
1. Configure SSL certificates
2. Set SSL mode in connection string
3. Verify certificate paths

### Authorization Issues

#### Problem: Insufficient permissions
```
error: permission denied for table users
```

**Solution:**
1. Adjust user permissions with GRANT commands
2. Implement Role-based access control (RBAC)

## Best Practices

### 1. Migration Management
- Break migrations into smaller chunks
- Create rollback plan for each migration
- Automate migration testing

### 2. Performance Optimization
- Use appropriate indexes
- Partition large tables
- Optimize queries

### 3. Security
- Encrypt sensitive data
- Perform regular security audits
- Maintain access logs

### 4. Backup
- Create regular backup schedule
- Test backups
- Store backups in different locations

## Getting Help

If you encounter issues not covered in this guide:

1. Check the [GitHub Issues](https://github.com/mstgnz/sdc/issues)
2. Search the [Documentation](https://github.com/mstgnz/sdc/docs)
3. Create a new issue with:
   - Error message
   - Steps to reproduce
   - Environment details
   - Sample code

## Version Migration Issues

### Version Upgrade
- Check changelog (CHANGELOG)
- Test upgrade in staging environment
- Backup before upgrade

### Version Downgrade
- Check backward compatibility
- Review data structure changes
- Review dependencies 