package parser

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewWorkerPool(t *testing.T) {
	tests := []struct {
		name      string
		workers   int
		queueSize int
	}{
		{
			name:      "default configuration",
			workers:   1,
			queueSize: 100,
		},
		{
			name:      "custom configuration",
			workers:   4,
			queueSize: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewWorkerPool(tt.workers, tt.queueSize)
			if pool == nil {
				t.Fatal("Expected non-nil WorkerPool")
				return
			}

			if pool.workers == 0 {
				t.Error("Expected non-zero workers")
			}
			if pool.queue == nil {
				t.Error("Expected non-nil queue channel")
			}
		})
	}
}

func TestWorkerPool_Submit(t *testing.T) {
	pool := NewWorkerPool(2, 2)

	// Test successful submission
	task := &testTask{id: "test-task"}

	err := pool.Submit(task)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test queue full
	for i := 0; i < 3; i++ {
		err = pool.Submit(task)
		if i == 2 && err == nil {
			t.Error("Expected error when queue is full")
		}
	}

	// Test submission after stop
	pool.Stop()
	err = pool.Submit(task)
	if err == nil {
		t.Error("Expected error after worker pool is stopped")
	}
}

type testTask struct {
	id string
}

func (tt *testTask) Execute() error {
	return nil
}

func TestWorkerPool_ProcessTask(t *testing.T) {
	pool := NewWorkerPool(1, 1)
	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop()

	task := &testTask{id: "test-task"}
	err := pool.Submit(task)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Wait for task to be processed
	time.Sleep(time.Millisecond * 100)
}

func TestWorkerPool_Metrics(t *testing.T) {
	pool := NewWorkerPool(2, 10)
	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop()

	// Submit some tasks
	for i := 0; i < 5; i++ {
		task := &testTask{id: fmt.Sprintf("task-%d", i)}
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

	if metrics.TasksProcessed < 5 {
		t.Errorf("Expected at least 5 processed tasks, got %d", metrics.TasksProcessed)
	}
}
