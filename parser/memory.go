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
	maxMemory      uint64
	gcThreshold    float64
	monitorTicker  time.Duration
	bufferPool     sync.Pool
	statsCollector *sync.Map
	isRunning      bool
	mu             sync.RWMutex
	stats          *MemoryStats
}

// MemoryStats holds memory statistics
type MemoryStats struct {
	Alloc      uint64
	TotalAlloc uint64
	Sys        uint64
	NumGC      uint32
}

// NewMemoryOptimizer creates a new memory optimizer
func NewMemoryOptimizer(maxMemoryMB int64, gcThreshold float64) *MemoryOptimizer {
	return &MemoryOptimizer{
		maxMemory:     uint64(maxMemoryMB * 1024 * 1024),
		gcThreshold:   gcThreshold,
		monitorTicker: time.Second,
		bufferPool: sync.Pool{
			New: func() interface{} {
				b := make([]byte, 32*1024) // 32KB default buffer size
				return &b
			},
		},
		statsCollector: &sync.Map{},
	}
}

// MonitorMemory monitors and optimizes memory usage
func (mo *MemoryOptimizer) MonitorMemory(ctx context.Context) {
	ticker := time.NewTicker(mo.monitorTicker)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := &runtime.MemStats{}
			runtime.ReadMemStats(stats)

			mo.mu.Lock()
			mo.stats = &MemoryStats{
				Alloc:      stats.Alloc,
				TotalAlloc: stats.TotalAlloc,
				Sys:        stats.Sys,
				NumGC:      stats.NumGC,
			}
			mo.mu.Unlock()

			if mo.stats.Alloc > mo.maxMemory {
				mo.ReleaseMemory()
			}

			mo.cleanupOldStats()
		}
	}
}

// GetBuffer gets a buffer from the pool
func (mo *MemoryOptimizer) GetBuffer() []byte {
	return *mo.bufferPool.Get().(*[]byte)
}

// PutBuffer returns a buffer to the pool
func (mo *MemoryOptimizer) PutBuffer(buf []byte) {
	// Clear sensitive data
	for i := range buf {
		buf[i] = 0
	}
	mo.bufferPool.Put(&buf)
}

// GetStats returns current memory statistics
func (mo *MemoryOptimizer) GetStats() *MemoryStats {
	mo.mu.Lock()
	defer mo.mu.Unlock()

	if mo.stats == nil {
		stats := &runtime.MemStats{}
		runtime.ReadMemStats(stats)
		mo.stats = &MemoryStats{
			Alloc:      stats.Alloc,
			TotalAlloc: stats.TotalAlloc,
			Sys:        stats.Sys,
			NumGC:      stats.NumGC,
		}
	}

	return mo.stats
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
	mo.maxMemory = uint64(maxMemoryMB * 1024 * 1024)
	mo.mu.Unlock()
}

// SetGCThreshold sets the garbage collection threshold
func (mo *MemoryOptimizer) SetGCThreshold(threshold float64) {
	mo.mu.Lock()
	mo.gcThreshold = threshold
	mo.mu.Unlock()
}

func (mo *MemoryOptimizer) ReleaseMemory() {
	runtime.GC()
	debug.FreeOSMemory()
}
