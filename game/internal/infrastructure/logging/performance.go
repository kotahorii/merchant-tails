package logging

import (
	"fmt"
	"sync"
	"time"
)

// PerformanceLogger logs performance metrics
type PerformanceLogger struct {
	logger  *Logger
	metrics map[string]*PerformanceMetric
	mu      sync.RWMutex
}

// PerformanceMetric represents a performance measurement
type PerformanceMetric struct {
	Name         string
	Count        int64
	TotalTime    time.Duration
	MinTime      time.Duration
	MaxTime      time.Duration
	LastTime     time.Duration
	AverageTime  time.Duration
	LastRecorded time.Time
}

// NewPerformanceLogger creates a new performance logger
func NewPerformanceLogger(logger *Logger) *PerformanceLogger {
	return &PerformanceLogger{
		logger:  logger,
		metrics: make(map[string]*PerformanceMetric),
	}
}

// StartOperation starts timing an operation
func (pl *PerformanceLogger) StartOperation(name string) *OperationTimer {
	return &OperationTimer{
		name:      name,
		startTime: time.Now(),
		logger:    pl,
	}
}

// RecordOperation records the duration of an operation
func (pl *PerformanceLogger) RecordOperation(name string, duration time.Duration, metadata map[string]interface{}) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	// Get or create metric
	metric, exists := pl.metrics[name]
	if !exists {
		metric = &PerformanceMetric{
			Name:    name,
			MinTime: duration,
			MaxTime: duration,
		}
		pl.metrics[name] = metric
	}

	// Update metric
	metric.Count++
	metric.TotalTime += duration
	metric.LastTime = duration
	metric.LastRecorded = time.Now()

	if duration < metric.MinTime {
		metric.MinTime = duration
	}
	if duration > metric.MaxTime {
		metric.MaxTime = duration
	}

	metric.AverageTime = time.Duration(int64(metric.TotalTime) / metric.Count)

	// Log performance metric
	if pl.logger != nil {
		fields := map[string]interface{}{
			"operation":   name,
			"duration_ms": float64(duration.Milliseconds()),
			"count":       metric.Count,
			"avg_ms":      float64(metric.AverageTime.Milliseconds()),
			"min_ms":      float64(metric.MinTime.Milliseconds()),
			"max_ms":      float64(metric.MaxTime.Milliseconds()),
		}

		// Add metadata
		for k, v := range metadata {
			fields[k] = v
		}

		pl.logger.WithFields(fields).Debug("Performance metric recorded")
	}
}

// GetMetric returns a specific performance metric
func (pl *PerformanceLogger) GetMetric(name string) *PerformanceMetric {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return pl.metrics[name]
}

// GetAllMetrics returns all performance metrics
func (pl *PerformanceLogger) GetAllMetrics() map[string]*PerformanceMetric {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	result := make(map[string]*PerformanceMetric)
	for k, v := range pl.metrics {
		// Create a copy to avoid race conditions
		metricCopy := *v
		result[k] = &metricCopy
	}
	return result
}

// Reset resets all performance metrics
func (pl *PerformanceLogger) Reset() {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	pl.metrics = make(map[string]*PerformanceMetric)
}

// ResetMetric resets a specific metric
func (pl *PerformanceLogger) ResetMetric(name string) {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	delete(pl.metrics, name)
}

// OperationTimer times an operation
type OperationTimer struct {
	name      string
	startTime time.Time
	logger    *PerformanceLogger
	metadata  map[string]interface{}
}

// End ends the timer and records the operation
func (ot *OperationTimer) End() time.Duration {
	duration := time.Since(ot.startTime)
	ot.logger.RecordOperation(ot.name, duration, ot.metadata)
	return duration
}

// EndWithMetadata ends the timer with additional metadata
func (ot *OperationTimer) EndWithMetadata(metadata map[string]interface{}) time.Duration {
	ot.metadata = metadata
	return ot.End()
}

// AddMetadata adds metadata to the timer
func (ot *OperationTimer) AddMetadata(key string, value interface{}) *OperationTimer {
	if ot.metadata == nil {
		ot.metadata = make(map[string]interface{})
	}
	ot.metadata[key] = value
	return ot
}

// PerformanceReport generates a performance report
type PerformanceReport struct {
	GeneratedAt time.Time                     `json:"generated_at"`
	Metrics     map[string]*PerformanceMetric `json:"metrics"`
	Summary     map[string]interface{}        `json:"summary"`
	Alerts      []string                      `json:"alerts"`
}

// GenerateReport generates a performance report
func (pl *PerformanceLogger) GenerateReport() *PerformanceReport {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	report := &PerformanceReport{
		GeneratedAt: time.Now(),
		Metrics:     make(map[string]*PerformanceMetric),
		Summary:     make(map[string]interface{}),
		Alerts:      make([]string, 0),
	}

	// Copy metrics
	var totalOperations int64
	var totalTime time.Duration
	var slowestOperation string
	var slowestTime time.Duration

	for name, metric := range pl.metrics {
		metricCopy := *metric
		report.Metrics[name] = &metricCopy

		totalOperations += metric.Count
		totalTime += metric.TotalTime

		if metric.MaxTime > slowestTime {
			slowestTime = metric.MaxTime
			slowestOperation = name
		}

		// Check for performance alerts
		if metric.AverageTime > 1*time.Second {
			report.Alerts = append(report.Alerts,
				fmt.Sprintf("Operation '%s' has high average time: %v", name, metric.AverageTime))
		}

		if metric.MaxTime > 5*time.Second {
			report.Alerts = append(report.Alerts,
				fmt.Sprintf("Operation '%s' has very high max time: %v", name, metric.MaxTime))
		}
	}

	// Generate summary
	report.Summary["total_operations"] = totalOperations
	report.Summary["total_time"] = totalTime.String()
	report.Summary["unique_operations"] = len(pl.metrics)
	report.Summary["slowest_operation"] = slowestOperation
	report.Summary["slowest_time"] = slowestTime.String()

	if totalOperations > 0 {
		avgTime := time.Duration(int64(totalTime) / totalOperations)
		report.Summary["average_time"] = avgTime.String()
	}

	return report
}

// ThresholdMonitor monitors performance thresholds
type ThresholdMonitor struct {
	thresholds map[string]time.Duration
	callbacks  map[string]func(string, time.Duration)
	mu         sync.RWMutex
}

// NewThresholdMonitor creates a new threshold monitor
func NewThresholdMonitor() *ThresholdMonitor {
	return &ThresholdMonitor{
		thresholds: make(map[string]time.Duration),
		callbacks:  make(map[string]func(string, time.Duration)),
	}
}

// SetThreshold sets a performance threshold for an operation
func (tm *ThresholdMonitor) SetThreshold(operation string, threshold time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.thresholds[operation] = threshold
}

// SetCallback sets a callback for threshold violations
func (tm *ThresholdMonitor) SetCallback(operation string, callback func(string, time.Duration)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.callbacks[operation] = callback
}

// CheckThreshold checks if a duration exceeds the threshold
func (tm *ThresholdMonitor) CheckThreshold(operation string, duration time.Duration) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	threshold, exists := tm.thresholds[operation]
	if !exists {
		return false
	}

	if duration > threshold {
		// Call callback if exists
		if callback, ok := tm.callbacks[operation]; ok {
			go callback(operation, duration)
		}
		return true
	}

	return false
}

// PerformanceHook is a log hook that tracks performance
type PerformanceHook struct {
	perfLogger *PerformanceLogger
	monitor    *ThresholdMonitor
}

// NewPerformanceHook creates a new performance hook
func NewPerformanceHook(perfLogger *PerformanceLogger) *PerformanceHook {
	return &PerformanceHook{
		perfLogger: perfLogger,
		monitor:    NewThresholdMonitor(),
	}
}

// Fire processes log entries for performance metrics
func (h *PerformanceHook) Fire(entry *LogEntry) error {
	// Look for performance-related fields
	if operation, ok := entry.Fields["operation"].(string); ok {
		if duration, ok := entry.Fields["duration"].(time.Duration); ok {
			// Record the operation
			metadata := make(map[string]interface{})
			for k, v := range entry.Fields {
				if k != "operation" && k != "duration" {
					metadata[k] = v
				}
			}
			h.perfLogger.RecordOperation(operation, duration, metadata)

			// Check threshold
			if h.monitor.CheckThreshold(operation, duration) {
				// Log threshold violation
				fmt.Printf("Performance threshold exceeded for %s: %v\n", operation, duration)
			}
		}
	}

	return nil
}

// Levels returns all log levels for performance tracking
func (h *PerformanceHook) Levels() []LogLevel {
	return []LogLevel{
		DebugLevel,
		InfoLevel,
		WarnLevel,
		ErrorLevel,
	}
}

// Benchmark runs a benchmark on a function
func Benchmark(name string, iterations int, fn func()) *PerformanceMetric {
	var totalDuration time.Duration
	minDuration := time.Duration(1<<63 - 1)
	var maxDuration time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		fn()
		duration := time.Since(start)

		totalDuration += duration
		if duration < minDuration {
			minDuration = duration
		}
		if duration > maxDuration {
			maxDuration = duration
		}
	}

	return &PerformanceMetric{
		Name:         name,
		Count:        int64(iterations),
		TotalTime:    totalDuration,
		MinTime:      minDuration,
		MaxTime:      maxDuration,
		AverageTime:  time.Duration(int64(totalDuration) / int64(iterations)),
		LastRecorded: time.Now(),
	}
}

// MeasureFunc measures the execution time of a function
func MeasureFunc(name string, logger *PerformanceLogger, fn func() error) error {
	timer := logger.StartOperation(name)
	err := fn()
	timer.AddMetadata("success", err == nil)
	if err != nil {
		timer.AddMetadata("error", err.Error())
	}
	timer.End()
	return err
}
