package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mstgnz/sdc/parser"
)

func main() {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Initialize memory optimizer
	memOptimizer := parser.NewMemoryOptimizer(1024, 0.8) // 1GB max memory, 80% GC threshold
	go memOptimizer.MonitorMemory(ctx)

	// Create worker pool
	wp := parser.NewWorkerPool(parser.WorkerConfig{
		Workers:      4,
		QueueSize:    1000,
		MemOptimizer: memOptimizer,
		ErrHandler: func(err error) {
			log.Printf("Worker error: %v", err)
		},
	})

	// Start worker pool
	wp.Start(ctx)
	defer wp.Stop()

	// Create batch processor
	bp := parser.NewBatchProcessor(parser.BatchConfig{
		BatchSize:    100,
		Workers:      4,
		Timeout:      30 * time.Second,
		MemOptimizer: memOptimizer,
		ErrorHandler: func(err error) {
			log.Printf("Batch error: %v", err)
		},
	})

	// Process SQL files in examples/files directory
	if err := processFiles(ctx, "examples/files", wp, bp); err != nil {
		log.Fatalf("Failed to process files: %v", err)
	}

	// Wait for all tasks to complete
	if err := wp.WaitForTasks(ctx); err != nil {
		log.Printf("Error waiting for tasks: %v", err)
	}

	// Print metrics
	metrics := wp.GetMetrics()
	fmt.Printf("\nWorker Pool Metrics:\n")
	fmt.Printf("Active Workers: %d\n", metrics.ActiveWorkers)
	fmt.Printf("Completed Tasks: %d\n", metrics.CompletedTasks)
	fmt.Printf("Failed Tasks: %d\n", metrics.FailedTasks)
	fmt.Printf("Average Latency: %v\n", metrics.AverageLatency)
}

func processFiles(ctx context.Context, dir string, wp *parser.WorkerPool, bp *parser.BatchProcessor) error {
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
		stmt := parser.NewStatement(string(content))

		// Submit task to worker pool
		task := parser.Task{
			ID:        path,
			Statement: stmt,
			Priority:  1,
			Timeout:   10 * time.Second,
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
