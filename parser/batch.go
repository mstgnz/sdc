package parser

import (
	"context"
	"runtime"
	"sync"
)

// BatchProcessor handles batch processing of SQL statements
type BatchProcessor struct {
	batchSize int
	workers   int
	queue     chan *Statement
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(batchSize, workers int) *BatchProcessor {
	if workers == 0 {
		workers = runtime.NumCPU()
	}
	if batchSize == 0 {
		batchSize = 1000
	}

	return &BatchProcessor{
		batchSize: batchSize,
		workers:   workers,
		queue:     make(chan *Statement, batchSize),
	}
}

// ProcessBatch processes statements in batches
func (bp *BatchProcessor) ProcessBatch(ctx context.Context, stmts []*Statement, handler func(*Statement) error) error {
	var wg sync.WaitGroup
	errChan := make(chan error, bp.workers)

	// Start worker pool
	for i := 0; i < bp.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for stmt := range bp.queue {
				if err := handler(stmt); err != nil {
					select {
					case errChan <- err:
					default:
					}
					return
				}
			}
		}()
	}

	// Send statements to workers
	for _, stmt := range stmts {
		select {
		case <-ctx.Done():
			close(bp.queue)
			return ctx.Err()
		case bp.queue <- stmt:
		case err := <-errChan:
			close(bp.queue)
			return err
		}
	}

	close(bp.queue)
	wg.Wait()

	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}
