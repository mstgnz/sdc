package migration

import (
	"fmt"
	"time"
)

// Migration represents a database migration
type Migration struct {
	ID        int64
	Name      string
	Version   string
	Status    string
	AppliedAt time.Time
}

// MigrationManager handles database migrations
type MigrationManager struct {
	driver Driver
}

// Driver interface for database specific operations
type Driver interface {
	CreateMigrationsTable() error
	GetAppliedMigrations() ([]Migration, error)
	ApplyMigration(migration Migration) error
	RollbackMigration(migration Migration) error
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(driver Driver) *MigrationManager {
	return &MigrationManager{
		driver: driver,
	}
}

// Apply applies pending migrations
func (m *MigrationManager) Apply(migrations []Migration) error {
	if err := m.driver.CreateMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	applied, err := m.driver.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, migration := range migrations {
		if !isApplied(applied, migration) {
			if err := m.driver.ApplyMigration(migration); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", migration.Name, err)
			}
		}
	}

	return nil
}

// Rollback rolls back the last applied migration
func (m *MigrationManager) Rollback() error {
	applied, err := m.driver.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	lastMigration := applied[len(applied)-1]
	if err := m.driver.RollbackMigration(lastMigration); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", lastMigration.Name, err)
	}

	return nil
}

func isApplied(applied []Migration, migration Migration) bool {
	for _, m := range applied {
		if m.Version == migration.Version {
			return true
		}
	}
	return false
}
