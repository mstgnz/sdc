package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// Config represents database connection configuration
type Config struct {
	Driver           string
	Host             string
	Port             int
	Database         string
	Username         string
	Password         string
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetime  time.Duration
	ConnMaxIdleTime  time.Duration
	ConnectionString string
}

// ConnectionManager manages database connections
type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]*sql.DB
	configs     map[string]Config
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*sql.DB),
		configs:     make(map[string]Config),
	}
}

// RegisterConnection registers a new database connection configuration
func (cm *ConnectionManager) RegisterConnection(name string, config Config) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.configs[name]; exists {
		return fmt.Errorf("connection %s already registered", name)
	}

	cm.configs[name] = config
	return nil
}

// GetConnection returns a database connection
func (cm *ConnectionManager) GetConnection(name string) (*sql.DB, error) {
	cm.mu.RLock()
	if db, exists := cm.connections[name]; exists {
		cm.mu.RUnlock()
		return db, nil
	}
	cm.mu.RUnlock()

	return cm.connect(name)
}

// connect establishes a new database connection
func (cm *ConnectionManager) connect(name string) (*sql.DB, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	config, exists := cm.configs[name]
	if !exists {
		return nil, fmt.Errorf("connection %s not registered", name)
	}

	var connStr string
	if config.ConnectionString != "" {
		connStr = config.ConnectionString
	} else {
		connStr = buildConnectionString(config)
	}

	db, err := sql.Open(config.Driver, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	cm.connections[name] = db
	return db, nil
}

// Close closes all database connections
func (cm *ConnectionManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var errs []error
	for name, db := range cm.connections {
		if err := db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection %s: %w", name, err))
		}
		delete(cm.connections, name)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}
	return nil
}

// buildConnectionString builds a connection string based on the configuration
func buildConnectionString(config Config) string {
	switch config.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			config.Username, config.Password, config.Host, config.Port, config.Database)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Host, config.Port, config.Username, config.Password, config.Database)
	case "sqlserver":
		return fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s",
			config.Host, config.Username, config.Password, config.Port, config.Database)
	case "oracle":
		return fmt.Sprintf("user=%s password=%s connectString=%s:%d/%s",
			config.Username, config.Password, config.Host, config.Port, config.Database)
	case "sqlite3":
		return config.Database
	default:
		return ""
	}
}
