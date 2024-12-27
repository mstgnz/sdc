package parser

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

func TestNewStreamParser(t *testing.T) {
	tests := []struct {
		name   string
		config StreamParserConfig
	}{
		{
			name:   "default configuration",
			config: StreamParserConfig{},
		},
		{
			name: "custom configuration",
			config: StreamParserConfig{
				Workers:      4,
				BatchSize:    2048,
				BufferSize:   64 * 1024,
				Timeout:      time.Second * 10,
				MaxRetries:   5,
				ErrorHandler: func(err error) {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewStreamParser(tt.config)
			if parser == nil {
				t.Fatal("Expected non-nil StreamParser")
				return
			}

			if parser.workers == 0 {
				t.Error("Expected non-zero workers")
			}
			if parser.batchSize == 0 {
				t.Error("Expected non-zero batch size")
			}
			if len(parser.buffer) == 0 {
				t.Error("Expected non-zero buffer size")
			}
		})
	}
}

func TestStreamParser_ParseStream(t *testing.T) {
	parser := NewStreamParser(StreamParserConfig{
		Workers:   2,
		BatchSize: 1024,
	})

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "empty input",
			input:       "",
			expectError: false,
		},
		{
			name:        "single statement",
			input:       "SELECT * FROM users;",
			expectError: false,
		},
		{
			name: "multiple statements",
			input: `
				CREATE TABLE users (id INT);
				INSERT INTO users VALUES (1);
				SELECT * FROM users;
			`,
			expectError: false,
		},
		{
			name:        "nil reader",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reader io.Reader
			if tt.name != "nil reader" {
				reader = strings.NewReader(tt.input)
			}

			err := parser.ParseStream(context.Background(), reader)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestStreamParser_ProcessBatch(t *testing.T) {
	parser := NewStreamParser(StreamParserConfig{
		Workers:   1,
		BatchSize: 1024,
		Timeout:   time.Second,
	})

	ctx := context.Background()
	batch := []byte("SELECT * FROM users; INSERT INTO users VALUES (1);")

	err := parser.processBatch(ctx, batch)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
}

func TestStreamParser_Timeout(t *testing.T) {
	parser := NewStreamParser(StreamParserConfig{
		Workers:   1,
		BatchSize: 1024,
		Timeout:   time.Millisecond * 100,
	})

	// Create a context that will timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	// Create a large input that will take time to process
	largeInput := strings.Repeat("SELECT * FROM users;", 1000)
	reader := strings.NewReader(largeInput)

	err := parser.ParseStream(ctx, reader)
	if err == nil {
		t.Error("Expected timeout error but got none")
	}
}

type mockTemporaryError struct{}

func (e mockTemporaryError) Error() string   { return "temporary error" }
func (e mockTemporaryError) Temporary() bool { return true }

func TestStreamParser_TemporaryError(t *testing.T) {
	err := mockTemporaryError{}
	if !isTemporaryError(err) {
		t.Error("Expected temporary error")
	}

	permanentErr := errors.New("permanent error")
	if isTemporaryError(permanentErr) {
		t.Error("Expected permanent error")
	}
}

func TestStreamParser_Configuration(t *testing.T) {
	parser := NewStreamParser(StreamParserConfig{})

	// Test SetTimeout
	newTimeout := time.Second * 5
	parser.SetTimeout(newTimeout)
	if parser.timeout != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, parser.timeout)
	}
}

func TestStreamParser_BufferHandling(t *testing.T) {
	parser := NewStreamParser(StreamParserConfig{
		BatchSize:  1024,
		BufferSize: 1024,
	})

	// Test buffer acquisition and return
	buf := parser.getBuffer()
	if len(buf) != parser.batchSize {
		t.Errorf("Expected buffer size %d, got %d", parser.batchSize, len(buf))
	}

	// Write some data
	testData := []byte("test data")
	copy(buf, testData)

	// Return buffer
	parser.putBuffer(buf)

	// Get buffer again and verify it's from pool
	buf = parser.getBuffer()
	if len(buf) != parser.batchSize {
		t.Errorf("Expected buffer size %d, got %d", parser.batchSize, len(buf))
	}
}

func TestStreamParser_ConcurrentParsing(t *testing.T) {
	parser := NewStreamParser(StreamParserConfig{
		Workers:   4,
		BatchSize: 1024,
	})

	// Create multiple inputs
	inputs := []string{
		"SELECT * FROM users;",
		"INSERT INTO users VALUES (1);",
		"UPDATE users SET name = 'test';",
		"DELETE FROM users WHERE id = 1;",
	}

	// Parse concurrently
	ctx := context.Background()
	errChan := make(chan error, len(inputs))
	for _, input := range inputs {
		go func(sql string) {
			reader := strings.NewReader(sql)
			errChan <- parser.ParseStream(ctx, reader)
		}(input)
	}

	// Check results
	for i := 0; i < len(inputs); i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	}
}
