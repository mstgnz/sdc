package parser

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// Task represents a unit of work
type Task interface {
	Execute() error
}

// WorkerPool manages a pool of workers
type WorkerPool struct {
	workers       int
	queue         chan Task
	done          chan struct{}
	errorHandler  func(error)
	wg            sync.WaitGroup
	activeWorkers int32
	taskCount     int32
	mu            sync.RWMutex
	metrics       *WorkerMetrics
	taskTimeout   time.Duration
}

// WorkerMetrics holds worker pool metrics
type WorkerMetrics struct {
	TasksProcessed int64
	TasksSucceeded int64
	TasksFailed    int64
	mu             sync.RWMutex
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int, queueSize int) *WorkerPool {
	if workers <= 0 {
		workers = 1
	}
	if queueSize <= 0 {
		queueSize = 100
	}

	return &WorkerPool{
		workers:     workers,
		queue:       make(chan Task, queueSize),
		done:        make(chan struct{}),
		taskTimeout: 30 * time.Second,
		metrics:     &WorkerMetrics{},
		errorHandler: func(err error) {
			log.Printf("Worker error: %v", err)
		},
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.wg.Add(wp.workers)
	for i := 0; i < wp.workers; i++ {
		go wp.worker(ctx)
	}
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() {
	close(wp.done)
	wp.wg.Wait()
}

// Submit submits a task to the worker pool
func (wp *WorkerPool) Submit(task Task) error {
	select {
	case wp.queue <- task:
		atomic.AddInt32(&wp.taskCount, 1)
		return nil
	case <-wp.done:
		return errors.New("worker pool is stopped")
	default:
		return errors.New("worker queue is full")
	}
}

// ProcessTask processes a single task with timeout
func (wp *WorkerPool) ProcessTask(ctx context.Context, task Task) error {
	if task == nil {
		return errors.New("nil task")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, wp.taskTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- task.Execute()
	}()

	select {
	case <-timeoutCtx.Done():
		if wp.errorHandler != nil {
			wp.errorHandler(timeoutCtx.Err())
		}
		atomic.AddInt64(&wp.metrics.TasksFailed, 1)
		return fmt.Errorf("task timeout: %v", timeoutCtx.Err())
	case err := <-done:
		if err != nil {
			if wp.errorHandler != nil {
				wp.errorHandler(err)
			}
			atomic.AddInt64(&wp.metrics.TasksFailed, 1)
		} else {
			atomic.AddInt64(&wp.metrics.TasksSucceeded, 1)
		}
		atomic.AddInt64(&wp.metrics.TasksProcessed, 1)
		return err
	}
}

func (wp *WorkerPool) worker(ctx context.Context) {
	defer wp.wg.Done()
	atomic.AddInt32(&wp.activeWorkers, 1)
	defer atomic.AddInt32(&wp.activeWorkers, -1)

	for {
		select {
		case <-ctx.Done():
			return
		case <-wp.done:
			return
		case task := <-wp.queue:
			_ = wp.ProcessTask(ctx, task)
			atomic.AddInt32(&wp.taskCount, -1)
		}
	}
}

// GetMetrics returns current worker pool metrics
func (wp *WorkerPool) GetMetrics() *WorkerMetrics {
	wp.metrics.mu.RLock()
	defer wp.metrics.mu.RUnlock()

	return &WorkerMetrics{
		TasksProcessed: atomic.LoadInt64(&wp.metrics.TasksProcessed),
		TasksSucceeded: atomic.LoadInt64(&wp.metrics.TasksSucceeded),
		TasksFailed:    atomic.LoadInt64(&wp.metrics.TasksFailed),
	}
}

// GetActiveWorkers returns the number of active workers
func (wp *WorkerPool) GetActiveWorkers() int32 {
	return atomic.LoadInt32(&wp.activeWorkers)
}

// GetTaskCount returns the number of tasks in the queue
func (wp *WorkerPool) GetTaskCount() int32 {
	return atomic.LoadInt32(&wp.taskCount)
}
