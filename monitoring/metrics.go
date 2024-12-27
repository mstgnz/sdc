package monitoring

import (
	"sync"
	"sync/atomic"
	"time"
)

// MetricsCollector collects and manages performance metrics
type MetricsCollector struct {
	totalObjects        int64
	totalProcessingTime int64
	failedOperations    int64
	memoryUsage         int64
	cpuUtilization      float64
	goroutineCount      int64
	channelBufferUsage  int64
	errorCount          map[string]int64
	errorCountMutex     sync.RWMutex
	retryAttempts       int64
	recoverySuccess     int64
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		errorCount: make(map[string]int64),
	}
}

// IncrementProcessedObjects increments the total objects counter
func (m *MetricsCollector) IncrementProcessedObjects() {
	atomic.AddInt64(&m.totalObjects, 1)
}

// RecordProcessingTime adds processing time to total
func (m *MetricsCollector) RecordProcessingTime(duration time.Duration) {
	atomic.AddInt64(&m.totalProcessingTime, int64(duration))
}

// IncrementFailedOperations increments the failed operations counter
func (m *MetricsCollector) IncrementFailedOperations() {
	atomic.AddInt64(&m.failedOperations, 1)
}

// SetMemoryUsage sets the current memory usage
func (m *MetricsCollector) SetMemoryUsage(bytes int64) {
	atomic.StoreInt64(&m.memoryUsage, bytes)
}

// SetCPUUtilization sets the current CPU utilization
func (m *MetricsCollector) SetCPUUtilization(percentage float64) {
	m.cpuUtilization = percentage
}

// SetGoroutineCount sets the current number of goroutines
func (m *MetricsCollector) SetGoroutineCount(count int64) {
	atomic.StoreInt64(&m.goroutineCount, count)
}

// SetChannelBufferUsage sets the current channel buffer usage
func (m *MetricsCollector) SetChannelBufferUsage(usage int64) {
	atomic.StoreInt64(&m.channelBufferUsage, usage)
}

// IncrementErrorCount increments the error count for a specific error type
func (m *MetricsCollector) IncrementErrorCount(errorType string) {
	m.errorCountMutex.Lock()
	m.errorCount[errorType]++
	m.errorCountMutex.Unlock()
}

// IncrementRetryAttempts increments the retry attempts counter
func (m *MetricsCollector) IncrementRetryAttempts() {
	atomic.AddInt64(&m.retryAttempts, 1)
}

// IncrementRecoverySuccess increments the recovery success counter
func (m *MetricsCollector) IncrementRecoverySuccess() {
	atomic.AddInt64(&m.recoverySuccess, 1)
}

// GetMetrics returns all current metrics
func (m *MetricsCollector) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"total_objects":         atomic.LoadInt64(&m.totalObjects),
		"total_processing_time": atomic.LoadInt64(&m.totalProcessingTime),
		"failed_operations":     atomic.LoadInt64(&m.failedOperations),
		"memory_usage":          atomic.LoadInt64(&m.memoryUsage),
		"cpu_utilization":       m.cpuUtilization,
		"goroutine_count":       atomic.LoadInt64(&m.goroutineCount),
		"channel_buffer_usage":  atomic.LoadInt64(&m.channelBufferUsage),
		"error_count":           m.errorCount,
		"retry_attempts":        atomic.LoadInt64(&m.retryAttempts),
		"recovery_success":      atomic.LoadInt64(&m.recoverySuccess),
	}
}

// TotalObjects returns the total number of processed objects
func (m *MetricsCollector) TotalObjects() int64 {
	return atomic.LoadInt64(&m.totalObjects)
}

// AverageProcessingTime returns the average processing time per object
func (m *MetricsCollector) AverageProcessingTime() time.Duration {
	total := atomic.LoadInt64(&m.totalObjects)
	if total == 0 {
		return 0
	}
	return time.Duration(atomic.LoadInt64(&m.totalProcessingTime) / total)
}

// MemoryUsage returns the current memory usage in bytes
func (m *MetricsCollector) MemoryUsage() int64 {
	return atomic.LoadInt64(&m.memoryUsage)
}

// ErrorRate returns the error rate as a percentage
func (m *MetricsCollector) ErrorRate() float64 {
	total := atomic.LoadInt64(&m.totalObjects)
	if total == 0 {
		return 0
	}
	failed := atomic.LoadInt64(&m.failedOperations)
	return float64(failed) / float64(total) * 100
}

// RecoveryRate returns the recovery success rate as a percentage
func (m *MetricsCollector) RecoveryRate() float64 {
	attempts := atomic.LoadInt64(&m.retryAttempts)
	if attempts == 0 {
		return 0
	}
	success := atomic.LoadInt64(&m.recoverySuccess)
	return float64(success) / float64(attempts) * 100
}
