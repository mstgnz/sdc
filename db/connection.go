package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ConnectionStatus represents the current status of connections
type ConnectionStatus struct {
	OpenConnections  int64
	InUseConnections int64
	IdleConnections  int64
	WaitCount        int64
	WaitDuration     time.Duration
}

// Config represents database connection configuration
type Config struct {
	Driver              string
	Host                string
	Port                int
	Database            string
	Username            string
	Password            string
	MaxOpenConns        int           // Maximum number of open connections
	MaxIdleConns        int           // Maximum number of idle connections
	ConnMaxLifetime     time.Duration // Maximum lifetime of a connection
	ConnMaxIdleTime     time.Duration // Maximum idle time of a connection
	ConnectionString    string        // Custom connection string (optional)
	RetryAttempts       int           // Number of connection retry attempts
	RetryDelay          time.Duration // Delay between retry attempts
	HealthCheckInterval time.Duration // Interval for health checks
	Timeout             time.Duration // Connection timeout
	ReadTimeout         time.Duration // Read operation timeout
	WriteTimeout        time.Duration // Write operation timeout
	SSLMode             string        // SSL mode (disable, require, verify-ca, verify-full)
	SSLCert             string        // Path to SSL certificate
	SSLKey              string        // Path to SSL key
	SSLRootCert         string        // Path to SSL root certificate
}

// ConnectionManager manages database connections with enhanced features
type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]*managedConnection
	configs     map[string]Config
	metrics     *connectionMetrics
}

// managedConnection wraps sql.DB with additional management features
type managedConnection struct {
	db          *sql.DB
	status      atomic.Value // stores ConnectionStatus
	lastChecked time.Time
	healthCheck chan struct{}
}

// connectionMetrics tracks connection pool metrics
type connectionMetrics struct {
	totalConnections atomic.Int64
	activeQueries    atomic.Int64
	errorCount       atomic.Int64
	queryDurations   sync.Map // stores time.Duration
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*managedConnection),
		configs:     make(map[string]Config),
		metrics:     &connectionMetrics{},
	}
}

// RegisterConnection registers a new database connection configuration
func (cm *ConnectionManager) RegisterConnection(name string, config Config) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.configs[name]; exists {
		return fmt.Errorf("connection %s already registered", name)
	}

	// Set default values if not specified
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 10
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 5
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = time.Hour
	}
	if config.ConnMaxIdleTime == 0 {
		config.ConnMaxIdleTime = 30 * time.Minute
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = time.Second
	}
	if config.HealthCheckInterval == 0 {
		config.HealthCheckInterval = time.Minute
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	cm.configs[name] = config
	return nil
}

// GetConnection returns a database connection with retry mechanism
func (cm *ConnectionManager) GetConnection(name string) (*sql.DB, error) {
	cm.mu.RLock()
	conn, exists := cm.connections[name]
	cm.mu.RUnlock()

	if exists {
		if err := conn.db.Ping(); err == nil {
			return conn.db, nil
		}
	}

	return cm.connectWithRetry(name)
}

// connectWithRetry attempts to establish a connection with retry mechanism
func (cm *ConnectionManager) connectWithRetry(name string) (*sql.DB, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	config, exists := cm.configs[name]
	if !exists {
		return nil, fmt.Errorf("connection %s not registered", name)
	}

	var db *sql.DB
	var err error

	for attempt := 1; attempt <= config.RetryAttempts; attempt++ {
		db, err = cm.connect(config)
		if err == nil {
			break
		}

		if attempt < config.RetryAttempts {
			time.Sleep(config.RetryDelay)
		}
	}

	if err != nil {
		cm.metrics.errorCount.Add(1)
		return nil, fmt.Errorf("failed to establish connection after %d attempts: %w",
			config.RetryAttempts, err)
	}

	conn := &managedConnection{
		db:          db,
		healthCheck: make(chan struct{}),
	}
	conn.status.Store(ConnectionStatus{})

	cm.connections[name] = conn
	cm.metrics.totalConnections.Add(1)

	// Start health check routine
	go cm.startHealthCheck(name, conn)

	return db, nil
}

// connect establishes a new database connection
func (cm *ConnectionManager) connect(config Config) (*sql.DB, error) {
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

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// startHealthCheck starts a health check routine for the connection
func (cm *ConnectionManager) startHealthCheck(name string, conn *managedConnection) {
	config := cm.configs[name]
	ticker := time.NewTicker(config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := conn.db.Ping(); err != nil {
				cm.metrics.errorCount.Add(1)
				// Attempt to reconnect
				if _, reconnectErr := cm.connectWithRetry(name); reconnectErr != nil {
					// Log reconnection failure
				}
			}
			cm.updateConnectionStatus(conn)
		case <-conn.healthCheck:
			return
		}
	}
}

// updateConnectionStatus updates the connection status
func (cm *ConnectionManager) updateConnectionStatus(conn *managedConnection) {
	stats := conn.db.Stats()
	status := ConnectionStatus{
		OpenConnections:  int64(stats.OpenConnections),
		InUseConnections: int64(stats.InUse),
		IdleConnections:  int64(stats.Idle),
		WaitCount:        int64(stats.WaitCount),
		WaitDuration:     stats.WaitDuration,
	}
	conn.status.Store(status)
}

// GetConnectionStatus returns the current status of a connection
func (cm *ConnectionManager) GetConnectionStatus(name string) (ConnectionStatus, error) {
	cm.mu.RLock()
	conn, exists := cm.connections[name]
	cm.mu.RUnlock()

	if !exists {
		return ConnectionStatus{}, fmt.Errorf("connection %s not found", name)
	}

	return conn.status.Load().(ConnectionStatus), nil
}

// GetMetrics returns the current connection metrics
func (cm *ConnectionManager) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"total_connections": cm.metrics.totalConnections.Load(),
		"active_queries":    cm.metrics.activeQueries.Load(),
		"error_count":       cm.metrics.errorCount.Load(),
	}
}

// Close closes all database connections
func (cm *ConnectionManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var errs []error
	for name, conn := range cm.connections {
		close(conn.healthCheck) // Stop health check routine
		if err := conn.db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection %s: %w", name, err))
		}
		delete(cm.connections, name)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}
	return nil
}

// buildConnectionString creates a connection string from config
func buildConnectionString(config Config) string {
	// Implementation depends on the database driver
	switch config.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			config.Username, config.Password, config.Host, config.Port, config.Database)
	// Add other database drivers as needed
	default:
		return ""
	}
}

// Example usage in comments:
/*
	// Create connection manager
	manager := NewConnectionManager()

	// Register connection with configuration
	err := manager.RegisterConnection("main", Config{
		Driver:             "postgres",
		Host:              "localhost",
		Port:              5432,
		Database:          "myapp",
		Username:          "user",
		Password:          "pass",
		MaxOpenConns:      20,
		MaxIdleConns:      10,
		ConnMaxLifetime:   time.Hour,
		ConnMaxIdleTime:   30 * time.Minute,
		RetryAttempts:     3,
		RetryDelay:        time.Second,
		HealthCheckInterval: time.Minute,
		Timeout:           30 * time.Second,
		SSLMode:          "require",
	})

	// Get connection
	db, err := manager.GetConnection("main")
	if err != nil {
		log.Fatal(err)
	}

	// Get connection status
	status, err := manager.GetConnectionStatus("main")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Open connections: %d\n", status.OpenConnections)

	// Get metrics
	metrics := manager.GetMetrics()
	fmt.Printf("Total connections: %d\n", metrics["total_connections"])

	// Close all connections
	manager.Close()
*/
