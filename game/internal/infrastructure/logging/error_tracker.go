package logging

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ErrorTracker tracks and analyzes errors
type ErrorTracker struct {
	errors      map[string]*ErrorInfo
	errorCounts map[string]int
	mu          sync.RWMutex
	maxErrors   int
	ttl         time.Duration
}

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Error       error
	Message     string
	StackTrace  string
	FirstSeen   time.Time
	LastSeen    time.Time
	Count       int
	Context     map[string]interface{}
	Fingerprint string
}

// NewErrorTracker creates a new error tracker
func NewErrorTracker(maxErrors int, ttl time.Duration) *ErrorTracker {
	et := &ErrorTracker{
		errors:      make(map[string]*ErrorInfo),
		errorCounts: make(map[string]int),
		maxErrors:   maxErrors,
		ttl:         ttl,
	}

	// Start cleanup routine
	go et.cleanupRoutine()

	return et
}

// TrackError tracks an error occurrence
func (et *ErrorTracker) TrackError(err error, context map[string]interface{}) *ErrorInfo {
	if err == nil {
		return nil
	}

	// Generate fingerprint for error grouping
	fingerprint := et.generateFingerprint(err)

	et.mu.Lock()
	defer et.mu.Unlock()

	// Check if error already exists
	if info, exists := et.errors[fingerprint]; exists {
		info.Count++
		info.LastSeen = time.Now()
		et.errorCounts[fingerprint]++
		return info
	}

	// Create new error info
	info := &ErrorInfo{
		Error:       err,
		Message:     err.Error(),
		StackTrace:  et.captureStackTrace(),
		FirstSeen:   time.Now(),
		LastSeen:    time.Now(),
		Count:       1,
		Context:     context,
		Fingerprint: fingerprint,
	}

	// Store error
	et.errors[fingerprint] = info
	et.errorCounts[fingerprint] = 1

	// Limit stored errors
	if len(et.errors) > et.maxErrors {
		et.evictOldestError()
	}

	return info
}

// GetErrorStats returns error statistics
func (et *ErrorTracker) GetErrorStats() map[string]interface{} {
	et.mu.RLock()
	defer et.mu.RUnlock()

	totalErrors := 0
	uniqueErrors := len(et.errors)
	topErrors := make([]map[string]interface{}, 0)

	// Calculate total errors
	for _, count := range et.errorCounts {
		totalErrors += count
	}

	// Get top errors
	for fingerprint, info := range et.errors {
		topErrors = append(topErrors, map[string]interface{}{
			"fingerprint": fingerprint,
			"message":     info.Message,
			"count":       info.Count,
			"first_seen":  info.FirstSeen,
			"last_seen":   info.LastSeen,
		})
	}

	// Sort by count (simple bubble sort for small datasets)
	for i := 0; i < len(topErrors)-1; i++ {
		for j := 0; j < len(topErrors)-i-1; j++ {
			if topErrors[j]["count"].(int) < topErrors[j+1]["count"].(int) {
				topErrors[j], topErrors[j+1] = topErrors[j+1], topErrors[j]
			}
		}
	}

	// Limit to top 10
	if len(topErrors) > 10 {
		topErrors = topErrors[:10]
	}

	return map[string]interface{}{
		"total_errors":  totalErrors,
		"unique_errors": uniqueErrors,
		"top_errors":    topErrors,
	}
}

// GetError returns error info by fingerprint
func (et *ErrorTracker) GetError(fingerprint string) *ErrorInfo {
	et.mu.RLock()
	defer et.mu.RUnlock()
	return et.errors[fingerprint]
}

// Clear clears all tracked errors
func (et *ErrorTracker) Clear() {
	et.mu.Lock()
	defer et.mu.Unlock()
	et.errors = make(map[string]*ErrorInfo)
	et.errorCounts = make(map[string]int)
}

// generateFingerprint generates a unique fingerprint for error grouping
func (et *ErrorTracker) generateFingerprint(err error) string {
	// Use error type and message for fingerprinting
	errType := fmt.Sprintf("%T", err)
	errMsg := err.Error()

	// Remove variable parts from error message (numbers, IDs, etc.)
	errMsg = et.normalizeErrorMessage(errMsg)

	return fmt.Sprintf("%s:%s", errType, errMsg)
}

// normalizeErrorMessage removes variable parts from error messages
func (et *ErrorTracker) normalizeErrorMessage(msg string) string {
	// Remove numbers
	msg = strings.ReplaceAll(msg, "0", "N")
	msg = strings.ReplaceAll(msg, "1", "N")
	msg = strings.ReplaceAll(msg, "2", "N")
	msg = strings.ReplaceAll(msg, "3", "N")
	msg = strings.ReplaceAll(msg, "4", "N")
	msg = strings.ReplaceAll(msg, "5", "N")
	msg = strings.ReplaceAll(msg, "6", "N")
	msg = strings.ReplaceAll(msg, "7", "N")
	msg = strings.ReplaceAll(msg, "8", "N")
	msg = strings.ReplaceAll(msg, "9", "N")

	// Remove common variable patterns
	// This is a simple implementation; in production, use regex
	return msg
}

// captureStackTrace captures the current stack trace
func (et *ErrorTracker) captureStackTrace() string {
	const maxStackDepth = 32
	pc := make([]uintptr, maxStackDepth)
	n := runtime.Callers(3, pc) // Skip runtime.Callers, captureStackTrace, and TrackError

	if n == 0 {
		return "No stack trace available"
	}

	frames := runtime.CallersFrames(pc[:n])
	var sb strings.Builder

	for {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}

	return sb.String()
}

// evictOldestError removes the oldest error from tracking
func (et *ErrorTracker) evictOldestError() {
	var oldestKey string
	var oldestTime time.Time

	for key, info := range et.errors {
		if oldestKey == "" || info.LastSeen.Before(oldestTime) {
			oldestKey = key
			oldestTime = info.LastSeen
		}
	}

	if oldestKey != "" {
		delete(et.errors, oldestKey)
		delete(et.errorCounts, oldestKey)
	}
}

// cleanupRoutine periodically removes expired errors
func (et *ErrorTracker) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		et.mu.Lock()
		now := time.Now()

		for key, info := range et.errors {
			if now.Sub(info.LastSeen) > et.ttl {
				delete(et.errors, key)
				delete(et.errorCounts, key)
			}
		}

		et.mu.Unlock()
	}
}

// ErrorPattern represents a pattern for error matching
type ErrorPattern struct {
	Name        string
	Pattern     string
	Severity    string
	Action      string
	Description string
}

// ErrorAnalyzer analyzes errors for patterns and insights
type ErrorAnalyzer struct {
	patterns []ErrorPattern
	mu       sync.RWMutex
}

// NewErrorAnalyzer creates a new error analyzer
func NewErrorAnalyzer() *ErrorAnalyzer {
	return &ErrorAnalyzer{
		patterns: []ErrorPattern{
			{
				Name:        "database_connection",
				Pattern:     "connection refused",
				Severity:    "critical",
				Action:      "Check database connection and credentials",
				Description: "Database connection failure",
			},
			{
				Name:        "file_not_found",
				Pattern:     "no such file or directory",
				Severity:    "warning",
				Action:      "Verify file path and permissions",
				Description: "File system access error",
			},
			{
				Name:        "out_of_memory",
				Pattern:     "out of memory",
				Severity:    "critical",
				Action:      "Increase memory allocation or optimize memory usage",
				Description: "Memory exhaustion",
			},
			{
				Name:        "timeout",
				Pattern:     "timeout",
				Severity:    "warning",
				Action:      "Increase timeout or optimize operation",
				Description: "Operation timeout",
			},
			{
				Name:        "permission_denied",
				Pattern:     "permission denied",
				Severity:    "error",
				Action:      "Check file/resource permissions",
				Description: "Permission error",
			},
		},
	}
}

// AnalyzeError analyzes an error and returns insights
func (ea *ErrorAnalyzer) AnalyzeError(err error) map[string]interface{} {
	if err == nil {
		return nil
	}

	ea.mu.RLock()
	defer ea.mu.RUnlock()

	errMsg := strings.ToLower(err.Error())
	insights := make(map[string]interface{})

	// Check against patterns
	for _, pattern := range ea.patterns {
		if strings.Contains(errMsg, strings.ToLower(pattern.Pattern)) {
			insights["pattern"] = pattern.Name
			insights["severity"] = pattern.Severity
			insights["action"] = pattern.Action
			insights["description"] = pattern.Description
			break
		}
	}

	// Add error type
	insights["error_type"] = fmt.Sprintf("%T", err)

	// Check if it's a wrapped error
	if unwrapped := fmt.Errorf("%w", err); unwrapped != err {
		insights["wrapped"] = true
	}

	return insights
}

// AddPattern adds a custom error pattern
func (ea *ErrorAnalyzer) AddPattern(pattern ErrorPattern) {
	ea.mu.Lock()
	defer ea.mu.Unlock()
	ea.patterns = append(ea.patterns, pattern)
}

// ErrorHook is a log hook that tracks errors
type ErrorHook struct {
	tracker  *ErrorTracker
	analyzer *ErrorAnalyzer
}

// NewErrorHook creates a new error hook
func NewErrorHook(tracker *ErrorTracker) *ErrorHook {
	return &ErrorHook{
		tracker:  tracker,
		analyzer: NewErrorAnalyzer(),
	}
}

// Fire processes log entries for errors
func (h *ErrorHook) Fire(entry *LogEntry) error {
	// Only track error and fatal levels
	if entry.Level != "ERROR" && entry.Level != "FATAL" {
		return nil
	}

	// Extract error from entry
	var err error
	if errField, ok := entry.Fields["error"].(string); ok && errField != "" {
		err = fmt.Errorf("%s", errField)
	} else {
		err = fmt.Errorf("%s", entry.Message)
	}

	// Track error
	errorInfo := h.tracker.TrackError(err, entry.Fields)

	// Analyze error
	if errorInfo != nil && errorInfo.Count == 1 {
		insights := h.analyzer.AnalyzeError(err)
		if insights != nil {
			// Log insights (avoid recursion by not using the logger)
			fmt.Printf("Error Analysis: %v\n", insights)
		}
	}

	return nil
}

// Levels returns the levels this hook is interested in
func (h *ErrorHook) Levels() []LogLevel {
	return []LogLevel{ErrorLevel, FatalLevel}
}
