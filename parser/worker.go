package parser

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrWorkerPoolStopped indicates the worker pool has been stopped
	ErrWorkerPoolStopped = errors.New("worker pool stopped")
	// ErrTaskTimeout indicates a task execution timeout
	ErrTaskTimeout = errors.New("task execution timeout")
	// ErrQueueFull indicates the task queue is full
	ErrQueueFull = errors.New("task queue is full")
)

// TaskStatus represents the status of a task
type TaskStatus int32

const (
	// Task statuses
	TaskPending TaskStatus = iota
	TaskRunning
	TaskCompleted
	TaskFailed
)

// Task represents a unit of work
type Task struct {
	ID        string
	Statement *Statement
	Priority  int
	Timeout   time.Duration
	Status    TaskStatus
	StartTime time.Time
	EndTime   time.Time
	Error     error
}

// Result represents the result of a task
type Result struct {
	TaskID    string
	Data      interface{}
	Error     error
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// WorkerPool manages a pool of workers
type WorkerPool struct {
	workers       int
	tasks         chan Task
	results       chan Result
	done          chan struct{}
	errHandler    func(error)
	memOptimizer  *MemoryOptimizer
	wg            sync.WaitGroup
	activeWorkers int32
	taskCount     int32
	mu            sync.RWMutex
	metrics       *WorkerMetrics
}

// WorkerMetrics holds worker pool metrics
type WorkerMetrics struct {
	ActiveWorkers   int32
	CompletedTasks  int32
	FailedTasks     int32
	AverageLatency  time.Duration
	ProcessingTasks int32
	QueueLength     int32
	mu              sync.RWMutex
}

// WorkerConfig represents worker pool configuration
type WorkerConfig struct {
	Workers      int
	QueueSize    int
	ErrHandler   func(error)
	MemOptimizer *MemoryOptimizer
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(config WorkerConfig) *WorkerPool {
	if config.Workers == 0 {
		config.Workers = runtime.NumCPU()
	}
	if config.QueueSize == 0 {
		config.QueueSize = 1000
	}
	if config.ErrHandler == nil {
		config.ErrHandler = func(err error) {
			// Default error handler just ignores the error
		}
	}

	return &WorkerPool{
		workers:      config.Workers,
		tasks:        make(chan Task, config.QueueSize),
		results:      make(chan Result, config.QueueSize),
		done:         make(chan struct{}),
		errHandler:   config.ErrHandler,
		memOptimizer: config.MemOptimizer,
		metrics: &WorkerMetrics{
			mu: sync.RWMutex{},
		},
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx)
	}
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() {
	close(wp.done)
	wp.wg.Wait()
	close(wp.results)
}

// Submit submits a task to the pool
func (wp *WorkerPool) Submit(task Task) error {
	select {
	case wp.tasks <- task:
		atomic.AddInt32(&wp.taskCount, 1)
		atomic.AddInt32(&wp.metrics.QueueLength, 1)
		return nil
	case <-wp.done:
		return ErrWorkerPoolStopped
	default:
		return ErrQueueFull
	}
}

// Results returns the results channel
func (wp *WorkerPool) Results() <-chan Result {
	return wp.results
}

// worker processes tasks
func (wp *WorkerPool) worker(ctx context.Context) {
	defer wp.wg.Done()
	atomic.AddInt32(&wp.activeWorkers, 1)
	atomic.AddInt32(&wp.metrics.ActiveWorkers, 1)
	defer func() {
		atomic.AddInt32(&wp.activeWorkers, -1)
		atomic.AddInt32(&wp.metrics.ActiveWorkers, -1)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-wp.done:
			return
		case task := <-wp.tasks:
			atomic.AddInt32(&wp.metrics.QueueLength, -1)
			atomic.AddInt32(&wp.metrics.ProcessingTasks, 1)
			result := wp.processTask(ctx, task)
			atomic.AddInt32(&wp.metrics.ProcessingTasks, -1)

			if result.Error != nil {
				atomic.AddInt32(&wp.metrics.FailedTasks, 1)
			} else {
				atomic.AddInt32(&wp.metrics.CompletedTasks, 1)
			}

			wp.updateMetrics(result)

			select {
			case wp.results <- result:
			case <-wp.done:
				return
			}
		}
	}
}

// processTask processes a single task
func (wp *WorkerPool) processTask(ctx context.Context, task Task) Result {
	result := Result{
		TaskID:    task.ID,
		StartTime: time.Now(),
	}

	// Create task context with timeout
	taskCtx := ctx
	if task.Timeout > 0 {
		var cancel context.CancelFunc
		taskCtx, cancel = context.WithTimeout(ctx, task.Timeout)
		defer cancel()
	}

	// Get buffer from memory optimizer if available
	if wp.memOptimizer != nil {
		buf := wp.memOptimizer.GetBuffer()
		defer wp.memOptimizer.PutBuffer(buf)
	}

	// Execute task with timeout
	done := make(chan struct{})
	go func() {
		defer close(done)
		atomic.StoreInt32((*int32)(&task.Status), int32(TaskRunning))
		if res, err := executeStatement(task.Statement); err != nil {
			result.Error = err
			atomic.StoreInt32((*int32)(&task.Status), int32(TaskFailed))
		} else {
			result.Data = res
			atomic.StoreInt32((*int32)(&task.Status), int32(TaskCompleted))
		}
	}()

	select {
	case <-taskCtx.Done():
		result.Error = fmt.Errorf("task timeout: %w", taskCtx.Err())
		atomic.StoreInt32((*int32)(&task.Status), int32(TaskFailed))
	case <-done:
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result
}

// updateMetrics updates worker pool metrics
func (wp *WorkerPool) updateMetrics(result Result) {
	wp.metrics.mu.Lock()
	defer wp.metrics.mu.Unlock()

	// Update average latency using exponential moving average
	alpha := 0.1 // Smoothing factor
	currentAvg := wp.metrics.AverageLatency
	newLatency := result.Duration
	wp.metrics.AverageLatency = time.Duration(float64(currentAvg)*(1-alpha) + float64(newLatency)*alpha)
}

// GetMetrics returns current worker pool metrics
func (wp *WorkerPool) GetMetrics() WorkerMetrics {
	wp.metrics.mu.RLock()
	defer wp.metrics.mu.RUnlock()
	return *wp.metrics
}

// WaitForTasks waits for all submitted tasks to complete
func (wp *WorkerPool) WaitForTasks(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if atomic.LoadInt32(&wp.taskCount) == 0 &&
				atomic.LoadInt32(&wp.metrics.ProcessingTasks) == 0 {
				return nil
			}
		}
	}
}

// executeStatement executes a SQL statement
func executeStatement(stmt *Statement) (interface{}, error) {
	// Implementation depends on the statement type
	// This is just a placeholder
	return nil, nil
}
