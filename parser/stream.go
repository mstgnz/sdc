package parser

import (
	"bufio"
	"context"
	"io"
	"runtime"
	"strings"
	"sync"
)

// StreamParser handles streaming SQL parsing
type StreamParser struct {
	buffer     []byte
	bufferPool *sync.Pool
	workers    int
	batchSize  int
}

// StreamParserConfig represents parser configuration
type StreamParserConfig struct {
	Workers    int // Number of concurrent workers
	BatchSize  int // Size of each batch in bytes
	BufferSize int // Size of read buffer
}

// NewStreamParser creates a new stream parser
func NewStreamParser(config StreamParserConfig) *StreamParser {
	if config.Workers == 0 {
		config.Workers = runtime.NumCPU()
	}
	if config.BatchSize == 0 {
		config.BatchSize = 1024 * 1024 // 1MB default batch size
	}
	if config.BufferSize == 0 {
		config.BufferSize = 32 * 1024 // 32KB default buffer size
	}

	return &StreamParser{
		buffer: make([]byte, config.BufferSize),
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, config.BatchSize)
			},
		},
		workers:   config.Workers,
		batchSize: config.BatchSize,
	}
}

// ParseStream parses SQL dump file in streaming mode
func (p *StreamParser) ParseStream(ctx context.Context, reader io.Reader, handler func(stmt *Statement) error) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(p.buffer, p.batchSize)

	// Create work channels
	jobs := make(chan string, p.workers)
	results := make(chan error, p.workers)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for query := range jobs {
				if err := handler(NewStatement(query)); err != nil {
					select {
					case results <- err:
					default:
					}
				}
			}
		}()
	}

	// Process SQL statements
	var currentStmt strings.Builder
	inStatement := false

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			close(jobs)
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		if !inStatement {
			currentStmt.Reset()
			inStatement = true
		}

		currentStmt.WriteString(line)
		currentStmt.WriteString(" ")

		if strings.HasSuffix(line, ";") {
			query := currentStmt.String()
			select {
			case jobs <- query:
			case err := <-results:
				close(jobs)
				return err
			}
			inStatement = false
		}
	}

	close(jobs)
	wg.Wait()

	select {
	case err := <-results:
		return err
	default:
		return scanner.Err()
	}
}
