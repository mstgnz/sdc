package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mstgnz/sqlporter/statement"
	"github.com/mstgnz/sqlporter/worker"
)

type SQLTask struct {
	id        string
	statement *statement.Statement
	timeout   time.Duration
}

func (t *SQLTask) Execute() error {
	// Task execution logic here
	return nil
}

func main() {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Initialize worker pool
	wp := worker.NewWorkerPool(4, 1000)

	// Start worker pool
	wp.Start(ctx)
	defer wp.Stop()

	// Process SQL files in examples/files directory
	if err := processFiles(ctx, "examples/files", wp); err != nil {
		log.Fatalf("Failed to process files: %v", err)
	}
}

func processFiles(_ context.Context, dir string, wp *worker.WorkerPool) error {
	// Walk through directory
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Process only .sql files
		if filepath.Ext(path) != ".sql" {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Create statement
		stmt := statement.NewStatement(string(content))

		// Submit task to worker pool
		task := &SQLTask{
			id:        path,
			statement: stmt,
			timeout:   10 * time.Second,
		}

		if err := wp.Submit(task); err != nil {
			return fmt.Errorf("failed to submit task for file %s: %w", path, err)
		}

		return nil
	})
}

// Example SQL files to be placed in examples/files:
/*
-- examples/files/create_tables.sql
CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- examples/files/create_indexes.sql
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_posts_user ON posts(user_id);

-- examples/files/alter_tables.sql
ALTER TABLE users ADD COLUMN last_login TIMESTAMP NULL;
ALTER TABLE posts ADD COLUMN updated_at TIMESTAMP NULL;
*/
