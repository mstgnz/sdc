package parser

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewWorkerPool(t *testing.T) {
	tests := []struct {
		name   string
		config WorkerConfig
	}{
		{
			name:   "default configuration",
			config: WorkerConfig{},
		},
		{
			name: "custom configuration",
			config: WorkerConfig{
				Workers:    4,
				QueueSize:  100,
				ErrHandler: func(err error) {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewWorkerPool(tt.config)
			if pool == nil {
				t.Fatal("Expected non-nil WorkerPool")
				return
			}

			if pool.workers == 0 {
				t.Error("Expected non-zero workers")
			}
			if pool.tasks == nil {
				t.Error("Expected non-nil tasks channel")
			}
			if pool.results == nil {
				t.Error("Expected non-nil results channel")
			}
		})
	}
}

func TestWorkerPool_Submit(t *testing.T) {
	pool := NewWorkerPool(WorkerConfig{
		Workers:   2,
		QueueSize: 2,
	})

	// Test successful submission
	task := Task{
		ID: "test-task",
		Statement: &Statement{
			Query: "SELECT 1",
		},
	}

	err := pool.Submit(task)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test queue full
	for i := 0; i < 3; i++ {
		err = pool.Submit(task)
		if i == 2 && err != ErrQueueFull {
			t.Errorf("Expected ErrQueueFull, got %v", err)
		}
	}

	// Test submission after stop
	pool.Stop()
	err = pool.Submit(task)
	if err != ErrWorkerPoolStopped {
		t.Errorf("Expected ErrWorkerPoolStopped, got %v", err)
	}
}

func TestWorkerPool_ProcessTask(t *testing.T) {
	pool := NewWorkerPool(WorkerConfig{
		Workers:   1,
		QueueSize: 1,
	})

	ctx := context.Background()
	task := Task{
		ID: "test-task",
		Statement: &Statement{
			Query: "SELECT 1",
		},
		Timeout: time.Second,
	}

	result := pool.processTask(ctx, task)
	if result.TaskID != task.ID {
		t.Errorf("Expected task ID %s, got %s", task.ID, result.TaskID)
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got %v", result.Error)
	}
	if result.Duration <= 0 {
		t.Error("Expected positive duration")
	}
}

func TestWorkerPool_TaskTimeout(t *testing.T) {
	pool := NewWorkerPool(WorkerConfig{
		Workers:   1,
		QueueSize: 1,
	})

	ctx := context.Background()
	task := Task{
		ID: "timeout-task",
		Statement: &Statement{
			Query: "SELECT pg_sleep(2)", // A long running query
		},
		Timeout: time.Millisecond * 100, // Short timeout
	}

	result := pool.processTask(ctx, task)
	if result.Error == nil {
		t.Error("Expected timeout error")
	}
}

func TestWorkerPool_Metrics(t *testing.T) {
	pool := NewWorkerPool(WorkerConfig{
		Workers:   2,
		QueueSize: 10,
	})

	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop()

	// Submit some tasks
	for i := 0; i < 5; i++ {
		task := Task{
			ID: fmt.Sprintf("task-%d", i),
			Statement: &Statement{
				Query: "SELECT 1",
			},
		}
		err := pool.Submit(task)
		if err != nil {
			t.Errorf("Failed to submit task: %v", err)
		}
	}

	// Wait for tasks to complete
	time.Sleep(time.Second)

	metrics := pool.GetMetrics()
	if metrics == nil {
		t.Fatal("Expected non-nil metrics")
	}

	if metrics.ActiveWorkers != 2 {
		t.Errorf("Expected 2 active workers, got %d", metrics.ActiveWorkers)
	}

	if metrics.CompletedTasks < 5 {
		t.Errorf("Expected at least 5 completed tasks, got %d", metrics.CompletedTasks)
	}

	if metrics.AverageLatency <= 0 {
		t.Error("Expected positive average latency")
	}
}

func TestWorkerPool_ErrorHandling(t *testing.T) {
	var handledError error
	pool := NewWorkerPool(WorkerConfig{
		Workers:   1,
		QueueSize: 1,
		ErrHandler: func(err error) {
			handledError = err
		},
	})

	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop()

	task := Task{
		ID: "error-task",
		Statement: &Statement{
			Query: "INVALID SQL",
		},
	}

	err := pool.Submit(task)
	if err != nil {
		t.Errorf("Failed to submit task: %v", err)
	}

	// Wait for error handling
	time.Sleep(time.Second)

	if handledError == nil {
		t.Error("Expected error to be handled")
	}
}
