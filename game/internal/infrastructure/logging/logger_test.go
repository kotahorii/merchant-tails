package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoggerCreation(t *testing.T) {
	tests := []struct {
		name   string
		config *LoggerConfig
		want   LogLevel
	}{
		{
			name:   "Default config",
			config: DefaultConfig(),
			want:   InfoLevel,
		},
		{
			name: "Debug level config",
			config: &LoggerConfig{
				Level:   DebugLevel,
				Console: true,
				JSON:    true,
			},
			want: DebugLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			if logger.level != tt.want {
				t.Errorf("Logger level = %v, want %v", logger.level, tt.want)
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	config := &LoggerConfig{
		Level:   DebugLevel,
		Console: false,
		JSON:    true,
	}

	// Create logger with buffer output
	logger, _ := NewLogger(config)
	logger.logger = logger.logger.Output(&buf)

	// Test all log levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Error("Debug message not found")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Info message not found")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message not found")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message not found")
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	config := &LoggerConfig{
		Level:   InfoLevel,
		Console: false,
		JSON:    true,
	}

	logger, _ := NewLogger(config)
	logger.logger = logger.logger.Output(&buf)

	// Log with fields
	logger.WithFields(map[string]interface{}{
		"user_id": "123",
		"action":  "purchase",
		"amount":  99.99,
	}).Info("Transaction processed")

	output := buf.String()

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if logEntry["user_id"] != "123" {
		t.Error("user_id field not found or incorrect")
	}
	if logEntry["action"] != "purchase" {
		t.Error("action field not found or incorrect")
	}
}

func TestLoggerWithContext(t *testing.T) {
	var buf bytes.Buffer
	config := &LoggerConfig{
		Level:   InfoLevel,
		Console: false,
		JSON:    true,
	}

	logger, _ := NewLogger(config)
	logger.logger = logger.logger.Output(&buf)

	// Create context with values
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", "req-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")

	// Log with context
	logger.WithContext(ctx).Info("Request processed")

	output := buf.String()
	if !strings.Contains(output, "req-123") {
		t.Error("request_id not found in log")
	}
	if !strings.Contains(output, "user-456") {
		t.Error("user_id not found in log")
	}
}

func TestLoggerWithError(t *testing.T) {
	var buf bytes.Buffer
	config := &LoggerConfig{
		Level:   InfoLevel,
		Console: false,
		JSON:    true,
	}

	logger, _ := NewLogger(config)
	logger.logger = logger.logger.Output(&buf)

	err := errors.New("database connection failed")
	logger.WithError(err).Error("Failed to process request")

	output := buf.String()
	if !strings.Contains(output, "database connection failed") {
		t.Error("Error message not found in log")
	}
}

func TestJSONFormatter(t *testing.T) {
	formatter := NewJSONFormatter()
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   "Test message",
		Fields: map[string]interface{}{
			"key": "value",
		},
	}

	data, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Failed to format entry: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to parse formatted JSON: %v", err)
	}

	if result["message"] != "Test message" {
		t.Error("Message not found in formatted output")
	}
}

func TestTextFormatter(t *testing.T) {
	formatter := NewTextFormatter()
	formatter.DisableColors = true // Disable colors for testing

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   "Test message",
		Fields: map[string]interface{}{
			"user": "john",
			"age":  30,
		},
	}

	data, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Failed to format entry: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "[INFO]") {
		t.Error("Level not found in formatted output")
	}
	if !strings.Contains(output, "Test message") {
		t.Error("Message not found in formatted output")
	}
	if !strings.Contains(output, "user=john") {
		t.Error("Fields not found in formatted output")
	}
}

func TestRotatingFileWriter(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := &RotationConfig{
		MaxSize:    100, // 100 bytes for testing
		MaxAge:     1 * time.Hour,
		MaxBackups: 3,
		Compress:   false,
		LocalTime:  true,
	}

	writer, err := NewRotatingFileWriter(logFile, config)
	if err != nil {
		t.Fatalf("Failed to create rotating writer: %v", err)
	}
	defer writer.Close()

	// Write data that exceeds max size
	data := []byte("This is a test log message that will be repeated. ")
	for i := 0; i < 5; i++ {
		if _, err := writer.Write(data); err != nil {
			t.Fatalf("Failed to write: %v", err)
		}
	}

	// Check if rotation occurred
	files, err := filepath.Glob(filepath.Join(tempDir, "test*.log"))
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}

	if len(files) < 2 {
		t.Error("Expected rotation to create backup files")
	}
}

func TestErrorTracker(t *testing.T) {
	tracker := NewErrorTracker(10, 1*time.Hour)

	// Track multiple errors
	err1 := errors.New("connection refused")
	err2 := errors.New("timeout occurred")
	err3 := errors.New("connection refused") // Duplicate

	tracker.TrackError(err1, nil)
	tracker.TrackError(err2, nil)
	tracker.TrackError(err3, nil)

	stats := tracker.GetErrorStats()

	if stats["total_errors"].(int) != 3 {
		t.Errorf("Expected 3 total errors, got %v", stats["total_errors"])
	}

	if stats["unique_errors"].(int) != 2 {
		t.Errorf("Expected 2 unique errors, got %v", stats["unique_errors"])
	}
}

func TestErrorAnalyzer(t *testing.T) {
	analyzer := NewErrorAnalyzer()

	tests := []struct {
		err      error
		expected string
	}{
		{
			err:      errors.New("connection refused"),
			expected: "database_connection",
		},
		{
			err:      errors.New("no such file or directory"),
			expected: "file_not_found",
		},
		{
			err:      errors.New("operation timeout"),
			expected: "timeout",
		},
	}

	for _, tt := range tests {
		insights := analyzer.AnalyzeError(tt.err)
		if pattern, ok := insights["pattern"]; ok {
			if pattern != tt.expected {
				t.Errorf("Expected pattern %s, got %s", tt.expected, pattern)
			}
		} else {
			t.Errorf("No pattern found for error: %v", tt.err)
		}
	}
}

func TestPerformanceLogger(t *testing.T) {
	logger, _ := NewLogger(DefaultConfig())
	perfLogger := NewPerformanceLogger(logger)

	// Record multiple operations
	for i := 0; i < 5; i++ {
		timer := perfLogger.StartOperation("test_operation")
		time.Sleep(10 * time.Millisecond)
		timer.End()
	}

	// Get metrics
	metric := perfLogger.GetMetric("test_operation")
	if metric == nil {
		t.Fatal("Metric not found")
	}

	if metric.Count != 5 {
		t.Errorf("Expected count 5, got %d", metric.Count)
	}

	if metric.AverageTime < 10*time.Millisecond {
		t.Error("Average time is less than expected")
	}
}

func TestPerformanceReport(t *testing.T) {
	logger, _ := NewLogger(DefaultConfig())
	perfLogger := NewPerformanceLogger(logger)

	// Record operations
	perfLogger.RecordOperation("fast_op", 100*time.Millisecond, nil)
	perfLogger.RecordOperation("slow_op", 2*time.Second, nil)
	perfLogger.RecordOperation("medium_op", 500*time.Millisecond, nil)

	// Generate report
	report := perfLogger.GenerateReport()

	if report.Summary["unique_operations"].(int) != 3 {
		t.Error("Expected 3 unique operations")
	}

	if len(report.Alerts) == 0 {
		t.Error("Expected alerts for slow operations")
	}
}

func TestBenchmark(t *testing.T) {
	result := Benchmark("test_function", 100, func() {
		// Simulate some work
		time.Sleep(1 * time.Millisecond)
	})

	if result.Count != 100 {
		t.Errorf("Expected 100 iterations, got %d", result.Count)
	}

	if result.AverageTime < 1*time.Millisecond {
		t.Error("Average time is less than expected")
	}
}

// TODO: Implement LoggerMiddleware functionality
// func TestLoggerMiddleware(t *testing.T) {
// 	logger, _ := NewLogger(DefaultConfig())
// 	middleware := NewLoggerMiddleware(logger)

// 	// Test transaction logging
// 	ctx := context.Background()
// 	err := middleware.LogTransaction(ctx, "txn-123", func() error {
// 		return nil
// 	})

// 	if err != nil {
// 		t.Error("Transaction should have succeeded")
// 	}

// 	// Test failed transaction
// 	err = middleware.LogTransaction(ctx, "txn-456", func() error {
// 		return errors.New("transaction failed")
// 	})

// 	if err == nil {
// 		t.Error("Transaction should have failed")
// 	}

// 	// Get error stats
// 	stats := middleware.GetErrorStats()
// 	if stats["total_errors"].(int) != 1 {
// 		t.Error("Expected 1 error in stats")
// 	}
// }

// TODO: Implement GameSystemLogger functionality
// func TestGameSystemLogger(t *testing.T) {
// 	sysLogger := NewGameSystemLogger("TestSystem")
// 	sysLogger.SetContext("session_id", "sess-123")

// 	// Test logging with context
// 	sysLogger.Info("System initialized", map[string]interface{}{
// 		"version": "1.0.0",
// 	})

// 	// Test operation logging
// 	err := sysLogger.LogOperation("startup", func() error {
// 		time.Sleep(10 * time.Millisecond)
// 		return nil
// 	})

// 	if err != nil {
// 		t.Error("Operation should have succeeded")
// 	}
// }

// TODO: Implement ConfigureLogging functionality
// func TestConfigureLogging(t *testing.T) {
// 	tempDir := t.TempDir()
// 	logFile := filepath.Join(tempDir, "app.log")

// 	err := ConfigureLogging("debug", logFile, true)
// 	if err != nil {
// 		t.Fatalf("Failed to configure logging: %v", err)
// 	}

// 	// Test global logger
// 	logger := GetLogger()
// 	if logger == nil {
// 		t.Fatal("Global logger not initialized")
// 	}

// 	logger.Info("Test message")

// 	// Check if log file was created
// 	if _, err := os.Stat(logFile); os.IsNotExist(err) {
// 		t.Error("Log file was not created")
// 	}
// }

func TestMeasureFunc(t *testing.T) {
	logger, _ := NewLogger(DefaultConfig())
	perfLogger := NewPerformanceLogger(logger)

	// Measure successful function
	err := MeasureFunc("test_func", perfLogger, func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Error("Function should have succeeded")
	}

	// Check metric was recorded
	metric := perfLogger.GetMetric("test_func")
	if metric == nil {
		t.Fatal("Metric not found")
	}

	if metric.Count != 1 {
		t.Error("Expected 1 execution")
	}
}

func BenchmarkLogger(b *testing.B) {
	config := &LoggerConfig{
		Level:   InfoLevel,
		Console: false,
		JSON:    true,
	}

	logger, _ := NewLogger(config)
	logger.logger = logger.logger.Output(io.Discard)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message")
	}
}

func BenchmarkLoggerWithFields(b *testing.B) {
	config := &LoggerConfig{
		Level:   InfoLevel,
		Console: false,
		JSON:    true,
	}

	logger, _ := NewLogger(config)
	logger.logger = logger.logger.Output(io.Discard)

	fields := map[string]interface{}{
		"user_id": "123",
		"action":  "test",
		"value":   42,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(fields).Info("Benchmark message")
	}
}
