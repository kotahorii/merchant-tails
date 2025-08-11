package monitoring

import (
	"runtime"
	"time"
)

// RuntimeCollector collects Go runtime metrics
type RuntimeCollector struct {
	metrics  *MetricsCollector
	stopChan chan struct{}
}

// NewRuntimeCollector creates a new runtime metrics collector
func NewRuntimeCollector(metrics *MetricsCollector) *RuntimeCollector {
	return &RuntimeCollector{
		metrics:  metrics,
		stopChan: make(chan struct{}),
	}
}

// Start begins collecting runtime metrics
func (rc *RuntimeCollector) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				rc.collect()
			case <-rc.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops collecting runtime metrics
func (rc *RuntimeCollector) Stop() {
	close(rc.stopChan)
}

// collect gathers current runtime metrics
func (rc *RuntimeCollector) collect() {
	// Memory statistics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Update memory usage
	rc.metrics.UpdateMemoryUsage(m.Alloc)

	// Update goroutine count
	rc.metrics.UpdateGoroutineCount(runtime.NumGoroutine())

	// Record GC pause time (most recent)
	if m.NumGC > 0 {
		pauseNs := m.PauseNs[(m.NumGC+255)%256]
		rc.metrics.RecordGCPause(time.Duration(pauseNs))
	}
}

// CollectNow performs an immediate collection
func (rc *RuntimeCollector) CollectNow() {
	rc.collect()
}
