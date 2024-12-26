package parser

import (
	"context"
	"runtime"
	"time"
)

// MemoryOptimizer handles memory optimization
type MemoryOptimizer struct {
	maxMemory   int64
	gcThreshold float64
}

// NewMemoryOptimizer creates a new memory optimizer
func NewMemoryOptimizer(maxMemoryMB int64, gcThreshold float64) *MemoryOptimizer {
	return &MemoryOptimizer{
		maxMemory:   maxMemoryMB * 1024 * 1024,
		gcThreshold: gcThreshold,
	}
}

// MonitorMemory monitors and optimizes memory usage
func (mo *MemoryOptimizer) MonitorMemory(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var ms runtime.MemStats
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runtime.ReadMemStats(&ms)
			if float64(ms.Alloc) > float64(mo.maxMemory)*mo.gcThreshold {
				runtime.GC()
			}
		}
	}
}
