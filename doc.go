/*
Package sqlmapper provides SQL dump conversion functionality between different database systems.

SQLPORTER (Parser, Mapper, Converter, Migrator, etc.) is a powerful Go library that allows you to convert SQL dump files
between different database systems. This library is particularly useful when you need to
migrate a database schema from one system to another.

Basic Usage:

	import "github.com/mstgnz/sqlmapper"

	// Create a MySQL parser
	parser := sqlmapper.NewMySQLParser()

	// Parse MySQL dump
	entity, err := Parse(mysqlDump)
	if err != nil {
		// handle error
	}

	// Convert to PostgreSQL
	pgParser := sqlmapper.NewPostgresParser()
	pgSQL, err := pgParser.Convert(entity)

Migration Support:

The package provides migration support through the migration package:

	import "github.com/mstgnz/sqlmapper/migration"

	// Create migration manager
	manager := migration.NewMigrationManager(driver)

	// Apply migrations
	err := manager.Apply(context.Background())

Schema Comparison:

Compare database schemas using the schema package:

	import "github.com/mstgnz/sqlmapper/schema"

	// Create schema comparer
	comparer := schema.NewSchemaComparer(sourceTables, targetTables)

	// Find differences
	differences := comparer.Compare()

Database Support:

The package supports the following databases:
  - MySQL
  - PostgreSQL
  - SQLite
  - Oracle
  - SQL Server

Each database has its own parser implementation that handles the specific syntax
and data types of that database system.

Error Handling:

All operations that can fail return an error as the last return value.
Errors should be checked and handled appropriately:

	if err != nil {
		switch {
		case errors.IsConnectionError(err):
			// handle connection error
		case errors.IsQueryError(err):
			// handle query error
		default:
			// handle other errors
		}
	}

Logging:

The package provides a structured logging system:

	import "github.com/mstgnz/sqlmapper/logger"

	log := logger.NewLogger(logger.Config{
		Level:  logger.INFO,
		Prefix: "[SQLPORTER] ",
	})

	log.Info("Starting conversion", map[string]interface{}{
		"source": "mysql",
		"target": "postgres",
	})

Configuration:

Most components can be configured through their respective Config structs:

	config := db.Config{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "mydb",
		Username: "user",
		Password: "pass",
	}

Thread Safety:

All public APIs in this package are thread-safe and can be used concurrently.

For more information and examples, visit: https://github.com/mstgnz/sqlmapper
*/
package sqlmapper
