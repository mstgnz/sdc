package migration

import (
	"errors"
	"testing"
	"time"
)

// MockDriver implements Driver interface for testing
type MockDriver struct {
	migrations        []Migration
	createTableError  error
	getMigrationsErr  error
	applyMigrationErr error
	rollbackErr       error
}

func (m *MockDriver) CreateMigrationsTable() error {
	return m.createTableError
}

func (m *MockDriver) GetAppliedMigrations() ([]Migration, error) {
	if m.getMigrationsErr != nil {
		return nil, m.getMigrationsErr
	}
	return m.migrations, nil
}

func (m *MockDriver) ApplyMigration(migration Migration) error {
	if m.applyMigrationErr != nil {
		return m.applyMigrationErr
	}
	m.migrations = append(m.migrations, migration)
	return nil
}

func (m *MockDriver) RollbackMigration(migration Migration) error {
	if m.rollbackErr != nil {
		return m.rollbackErr
	}
	if len(m.migrations) > 0 {
		m.migrations = m.migrations[:len(m.migrations)-1]
	}
	return nil
}

func TestNewMigrationManager(t *testing.T) {
	driver := &MockDriver{}
	manager := NewMigrationManager(driver)
	if manager == nil {
		t.Error("Expected non-nil MigrationManager")
	}
	if manager.driver != driver {
		t.Error("Expected driver to be set correctly")
	}
}

func TestApplyMigrations(t *testing.T) {
	tests := []struct {
		name          string
		driver        *MockDriver
		migrations    []Migration
		expectedError bool
	}{
		{
			name:   "successful migration",
			driver: &MockDriver{},
			migrations: []Migration{
				{Version: "1", Name: "first migration"},
				{Version: "2", Name: "second migration"},
			},
			expectedError: false,
		},
		{
			name: "create table error",
			driver: &MockDriver{
				createTableError: errors.New("create table error"),
			},
			migrations: []Migration{
				{Version: "1", Name: "first migration"},
			},
			expectedError: true,
		},
		{
			name: "get migrations error",
			driver: &MockDriver{
				getMigrationsErr: errors.New("get migrations error"),
			},
			migrations: []Migration{
				{Version: "1", Name: "first migration"},
			},
			expectedError: true,
		},
		{
			name: "apply migration error",
			driver: &MockDriver{
				applyMigrationErr: errors.New("apply migration error"),
			},
			migrations: []Migration{
				{Version: "1", Name: "first migration"},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMigrationManager(tt.driver)
			err := manager.Apply(tt.migrations)
			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestRollback(t *testing.T) {
	tests := []struct {
		name          string
		driver        *MockDriver
		migrations    []Migration
		expectedError bool
	}{
		{
			name: "successful rollback",
			driver: &MockDriver{
				migrations: []Migration{
					{Version: "1", Name: "first migration", AppliedAt: time.Now()},
				},
			},
			expectedError: false,
		},
		{
			name:          "no migrations to rollback",
			driver:        &MockDriver{},
			expectedError: true,
		},
		{
			name: "rollback error",
			driver: &MockDriver{
				migrations: []Migration{
					{Version: "1", Name: "first migration", AppliedAt: time.Now()},
				},
				rollbackErr: errors.New("rollback error"),
			},
			expectedError: true,
		},
		{
			name: "get migrations error",
			driver: &MockDriver{
				getMigrationsErr: errors.New("get migrations error"),
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMigrationManager(tt.driver)
			err := manager.Rollback()
			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestIsApplied(t *testing.T) {
	applied := []Migration{
		{Version: "1", Name: "first"},
		{Version: "2", Name: "second"},
	}

	tests := []struct {
		name      string
		migration Migration
		expected  bool
	}{
		{
			name:      "migration is applied",
			migration: Migration{Version: "1", Name: "first"},
			expected:  true,
		},
		{
			name:      "migration is not applied",
			migration: Migration{Version: "3", Name: "third"},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isApplied(applied, tt.migration)
			if result != tt.expected {
				t.Errorf("Expected isApplied to return %v for migration %s", tt.expected, tt.migration.Version)
			}
		})
	}
}
