package parser

import (
	"context"
	"testing"
	"time"
)

func TestNewMemoryOptimizer(t *testing.T) {
	tests := []struct {
		name        string
		maxMemoryMB int64
		gcThreshold float64
	}{
		{
			name:        "default values",
			maxMemoryMB: 1024,
			gcThreshold: 0.8,
		},
		{
			name:        "custom values",
			maxMemoryMB: 2048,
			gcThreshold: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			optimizer := NewMemoryOptimizer(tt.maxMemoryMB, tt.gcThreshold)
			if optimizer == nil {
				t.Fatal("Expected non-nil MemoryOptimizer")
				return
			}

			if optimizer.maxMemory != uint64(tt.maxMemoryMB*1024*1024) {
				t.Errorf("Expected max memory %d, got %d", tt.maxMemoryMB*1024*1024, optimizer.maxMemory)
			}
			if optimizer.gcThreshold != tt.gcThreshold {
				t.Errorf("Expected GC threshold %f, got %f", tt.gcThreshold, optimizer.gcThreshold)
			}
		})
	}
}

func TestMemoryOptimizer_BufferPool(t *testing.T) {
	optimizer := NewMemoryOptimizer(1024, 0.8)

	// Get buffer
	buf := optimizer.GetBuffer()
	if len(buf) != 32*1024 {
		t.Errorf("Expected buffer size 32KB, got %d", len(buf))
	}

	// Write some data
	testData := []byte("test data")
	copy(buf, testData)

	// Put buffer back
	optimizer.PutBuffer(buf)

	// Get buffer again and verify it's cleared
	buf = optimizer.GetBuffer()
	for i, b := range buf {
		if b != 0 {
			t.Errorf("Buffer not cleared at position %d: got %d", i, b)
		}
	}
}

func TestMemoryOptimizer_MonitorMemory(t *testing.T) {
	optimizer := NewMemoryOptimizer(1024, 0.8)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Start monitoring
	go optimizer.MonitorMemory(ctx)

	// Wait for some stats to be collected
	time.Sleep(time.Millisecond * 100)

	// Get stats
	stats := optimizer.GetStats()
	if stats == nil {
		t.Error("Expected non-nil memory stats")
	}
}

func TestMemoryOptimizer_Configuration(t *testing.T) {
	optimizer := NewMemoryOptimizer(1024, 0.8)

	// Test SetMonitoringInterval
	newInterval := time.Second * 2
	optimizer.SetMonitoringInterval(newInterval)
	if optimizer.monitorTicker != newInterval {
		t.Errorf("Expected monitoring interval %v, got %v", newInterval, optimizer.monitorTicker)
	}

	// Test SetMaxMemory
	newMaxMemory := int64(2048)
	optimizer.SetMaxMemory(newMaxMemory)
	if optimizer.maxMemory != uint64(newMaxMemory*1024*1024) {
		t.Errorf("Expected max memory %d, got %d", newMaxMemory*1024*1024, optimizer.maxMemory)
	}

	// Test SetGCThreshold
	newThreshold := 0.7
	optimizer.SetGCThreshold(newThreshold)
	if optimizer.gcThreshold != newThreshold {
		t.Errorf("Expected GC threshold %f, got %f", newThreshold, optimizer.gcThreshold)
	}
}

func TestMemoryOptimizer_StatsCollection(t *testing.T) {
	optimizer := NewMemoryOptimizer(1024, 0.8)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Start monitoring
	go optimizer.MonitorMemory(ctx)

	// Wait for multiple stats collections
	time.Sleep(time.Millisecond * 200)

	// Get stats
	stats := optimizer.GetStats()
	if stats == nil {
		t.Fatal("Expected non-nil memory stats")
	}

	// Verify stats fields
	if stats.Alloc == 0 {
		t.Error("Expected non-zero allocated bytes")
	}
	if stats.TotalAlloc == 0 {
		t.Error("Expected non-zero total allocated bytes")
	}
	if stats.Sys == 0 {
		t.Error("Expected non-zero system bytes")
	}
}

func TestMemoryOptimizer_ConcurrentAccess(t *testing.T) {
	optimizer := NewMemoryOptimizer(1024, 0.8)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Start monitoring
	go optimizer.MonitorMemory(ctx)

	// Concurrent buffer operations
	const goroutines = 10
	done := make(chan bool)

	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				buf := optimizer.GetBuffer()
				optimizer.PutBuffer(buf)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func TestMemoryOptimizer_MonitorMemoryRestart(t *testing.T) {
	optimizer := NewMemoryOptimizer(1024, 0.8)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	// Start monitoring multiple times
	go optimizer.MonitorMemory(ctx)
	go optimizer.MonitorMemory(ctx) // Should not start a second monitor

	// Wait for monitoring to stop
	time.Sleep(time.Millisecond * 300)

	// Verify isRunning is false after context cancellation
	optimizer.mu.RLock()
	if optimizer.isRunning {
		t.Error("Expected isRunning to be false after context cancellation")
	}
	optimizer.mu.RUnlock()
}
