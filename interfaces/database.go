package interfaces

import (
	"context"
	"database/sql"
)

// QueryExecutor basic query operations interface
type QueryExecutor interface {
	Execute(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// TransactionManager transaction operations interface
type TransactionManager interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// ConnectionManager connection management interface
type ConnectionManager interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
}

// SchemaManager schema management interface
type SchemaManager interface {
	CreateTable(ctx context.Context, table string) error
	DropTable(ctx context.Context, table string) error
	AlterTable(ctx context.Context, table string, alterations []string) error
}

// MigrationManager migration operations interface
type MigrationManager interface {
	ApplyMigrations(ctx context.Context) error
	RollbackMigration(ctx context.Context) error
	GetMigrationStatus(ctx context.Context) ([]MigrationStatus, error)
}

// MigrationStatus migration status struct
type MigrationStatus struct {
	ID        string
	Name      string
	Version   string
	AppliedAt string
	Status    string
}
