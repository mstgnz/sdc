package parser

import (
	"context"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// MemoryOptimizer handles memory optimization
type MemoryOptimizer struct {
	maxMemory      int64
	gcThreshold    float64
	monitorTicker  time.Duration
	bufferPool     *sync.Pool
	statsCollector *sync.Map
	isRunning      bool
	mu             sync.RWMutex
}

// MemoryStats holds memory statistics
type MemoryStats struct {
	AllocatedBytes    uint64
	TotalAllocBytes   uint64
	SystemBytes       uint64
	GCCycles          uint32
	LastGCPauseNanos  uint64
	TotalGCPauseNanos uint64
}

// NewMemoryOptimizer creates a new memory optimizer
func NewMemoryOptimizer(maxMemoryMB int64, gcThreshold float64) *MemoryOptimizer {
	return &MemoryOptimizer{
		maxMemory:     maxMemoryMB * 1024 * 1024,
		gcThreshold:   gcThreshold,
		monitorTicker: time.Second,
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 32*1024) // 32KB default buffer size
			},
		},
		statsCollector: &sync.Map{},
	}
}

// MonitorMemory monitors and optimizes memory usage
func (mo *MemoryOptimizer) MonitorMemory(ctx context.Context) {
	mo.mu.Lock()
	if mo.isRunning {
		mo.mu.Unlock()
		return
	}
	mo.isRunning = true
	mo.mu.Unlock()

	ticker := time.NewTicker(mo.monitorTicker)
	defer ticker.Stop()

	var ms runtime.MemStats
	for {
		select {
		case <-ctx.Done():
			mo.mu.Lock()
			mo.isRunning = false
			mo.mu.Unlock()
			return
		case <-ticker.C:
			runtime.ReadMemStats(&ms)

			// Collect stats
			stats := &MemoryStats{
				AllocatedBytes:   ms.Alloc,
				TotalAllocBytes:  ms.TotalAlloc,
				SystemBytes:      ms.Sys,
				GCCycles:         ms.NumGC,
				LastGCPauseNanos: ms.PauseNs[(ms.NumGC+255)%256],
			}
			for i := range ms.PauseNs {
				stats.TotalGCPauseNanos += ms.PauseNs[i]
			}
			mo.statsCollector.Store(time.Now().UnixNano(), stats)

			// Check if we need to trigger GC
			if float64(ms.Alloc) > float64(mo.maxMemory)*mo.gcThreshold {
				runtime.GC()
				debug.FreeOSMemory() // Return memory to OS
			}

			// Clean up old stats
			mo.cleanupOldStats()
		}
	}
}

// GetBuffer gets a buffer from the pool
func (mo *MemoryOptimizer) GetBuffer() []byte {
	return mo.bufferPool.Get().([]byte)
}

// PutBuffer returns a buffer to the pool
func (mo *MemoryOptimizer) PutBuffer(buf []byte) {
	// Clear sensitive data
	for i := range buf {
		buf[i] = 0
	}
	mo.bufferPool.Put(buf)
}

// GetStats returns current memory statistics
func (mo *MemoryOptimizer) GetStats() *MemoryStats {
	var latest *MemoryStats
	var latestTime int64

	mo.statsCollector.Range(func(key, value interface{}) bool {
		timestamp := key.(int64)
		if timestamp > latestTime {
			latestTime = timestamp
			latest = value.(*MemoryStats)
		}
		return true
	})

	return latest
}

// cleanupOldStats removes stats older than 1 hour
func (mo *MemoryOptimizer) cleanupOldStats() {
	threshold := time.Now().Add(-1 * time.Hour).UnixNano()
	mo.statsCollector.Range(func(key, _ interface{}) bool {
		timestamp := key.(int64)
		if timestamp < threshold {
			mo.statsCollector.Delete(key)
		}
		return true
	})
}

// SetMonitoringInterval sets the monitoring interval
func (mo *MemoryOptimizer) SetMonitoringInterval(d time.Duration) {
	mo.mu.Lock()
	mo.monitorTicker = d
	mo.mu.Unlock()
}

// SetMaxMemory sets the maximum memory threshold
func (mo *MemoryOptimizer) SetMaxMemory(maxMemoryMB int64) {
	mo.mu.Lock()
	mo.maxMemory = maxMemoryMB * 1024 * 1024
	mo.mu.Unlock()
}

// SetGCThreshold sets the garbage collection threshold
func (mo *MemoryOptimizer) SetGCThreshold(threshold float64) {
	mo.mu.Lock()
	mo.gcThreshold = threshold
	mo.mu.Unlock()
}
