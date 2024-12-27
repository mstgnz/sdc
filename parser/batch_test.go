package parser

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNewBatchProcessor(t *testing.T) {
	tests := []struct {
		name   string
		config BatchConfig
	}{
		{
			name:   "default configuration",
			config: BatchConfig{},
		},
		{
			name: "custom configuration",
			config: BatchConfig{
				BatchSize:    100,
				Workers:      4,
				Timeout:      time.Second * 10,
				ErrorHandler: func(err error) {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewBatchProcessor(tt.config)
			if processor == nil {
				t.Fatal("Expected non-nil BatchProcessor")
				return
			}

			if processor.batchSize == 0 {
				t.Error("Expected non-zero batch size")
			}
			if processor.workers == 0 {
				t.Error("Expected non-zero workers")
			}
			if processor.timeout == 0 {
				t.Error("Expected non-zero timeout")
			}
		})
	}
}

func TestBatchProcessor_ProcessBatch(t *testing.T) {
	processor := NewBatchProcessor(BatchConfig{
		BatchSize: 10,
		Workers:   2,
		Timeout:   time.Second,
	})

	tests := []struct {
		name          string
		statements    []*Statement
		handler       func(*Statement) error
		expectedError bool
	}{
		{
			name: "successful processing",
			statements: []*Statement{
				{Query: "SELECT 1"},
				{Query: "SELECT 2"},
			},
			handler: func(stmt *Statement) error {
				return nil
			},
			expectedError: false,
		},
		{
			name: "handler error",
			statements: []*Statement{
				{Query: "SELECT 1"},
			},
			handler: func(stmt *Statement) error {
				return errors.New("handler error")
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.ProcessBatch(context.Background(), tt.statements, tt.handler)
			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestBatchProcessor_Timeout(t *testing.T) {
	processor := NewBatchProcessor(BatchConfig{
		BatchSize: 10,
		Workers:   1,
		Timeout:   time.Millisecond * 100,
	})

	statements := []*Statement{
		{Query: "SELECT pg_sleep(1)"}, // Long running query
	}

	handler := func(stmt *Statement) error {
		time.Sleep(time.Second) // Simulate long processing
		return nil
	}

	err := processor.ProcessBatch(context.Background(), statements, handler)
	if err == nil {
		t.Error("Expected timeout error but got none")
	}
}

func TestBatchProcessor_Stop(t *testing.T) {
	processor := NewBatchProcessor(BatchConfig{
		BatchSize: 10,
		Workers:   2,
	})

	// Stop the processor
	processor.Stop()

	// Try to process after stopping
	err := processor.ProcessBatch(context.Background(), []*Statement{{Query: "SELECT 1"}}, func(stmt *Statement) error {
		return nil
	})

	if err != ErrBatchProcessorStopped {
		t.Errorf("Expected ErrBatchProcessorStopped but got: %v", err)
	}
}

func TestBatchProcessor_Configuration(t *testing.T) {
	processor := NewBatchProcessor(BatchConfig{})

	// Test SetTimeout
	newTimeout := time.Second * 5
	processor.SetTimeout(newTimeout)
	if processor.timeout != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, processor.timeout)
	}

	// Test SetBatchSize
	newBatchSize := 500
	processor.SetBatchSize(newBatchSize)
	if processor.batchSize != newBatchSize {
		t.Errorf("Expected batch size %d, got %d", newBatchSize, processor.batchSize)
	}

	// Test SetErrorHandler
	var handledError error
	newHandler := func(err error) {
		handledError = err
	}
	processor.SetErrorHandler(newHandler)

	testError := errors.New("test error")
	processor.errorHandler(testError)
	if handledError != testError {
		t.Error("Error handler not properly set")
	}
}

func TestBatchProcessor_ConcurrentProcessing(t *testing.T) {
	processor := NewBatchProcessor(BatchConfig{
		BatchSize: 100,
		Workers:   4,
	})

	// Create a large number of statements
	statements := make([]*Statement, 50)
	for i := range statements {
		statements[i] = &Statement{Query: "SELECT 1"}
	}

	// Create a handler that simulates some work
	processed := 0
	var mu sync.Mutex
	handler := func(stmt *Statement) error {
		time.Sleep(time.Millisecond) // Simulate work
		mu.Lock()
		processed++
		mu.Unlock()
		return nil
	}

	// Process statements
	err := processor.ProcessBatch(context.Background(), statements, handler)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	// Check if all statements were processed
	if processed != len(statements) {
		t.Errorf("Expected %d processed statements, got %d", len(statements), processed)
	}
}
