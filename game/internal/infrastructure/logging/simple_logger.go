package logging

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// SimpleLogger provides basic logging functionality
type SimpleLogger struct {
	level  LogLevel
	logger *log.Logger
	mu     sync.RWMutex
}

// NewSimpleLogger creates a basic logger
func NewSimpleLogger(level LogLevel) *SimpleLogger {
	return &SimpleLogger{
		level:  level,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// SetLevel sets the minimum log level
func (sl *SimpleLogger) SetLevel(level LogLevel) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.level = level
}

// Debug logs a debug message
func (sl *SimpleLogger) Debug(format string, args ...interface{}) {
	sl.log(LevelDebug, "DEBUG", format, args...)
}

// Info logs an info message
func (sl *SimpleLogger) Info(format string, args ...interface{}) {
	sl.log(LevelInfo, "INFO", format, args...)
}

// Warn logs a warning message
func (sl *SimpleLogger) Warn(format string, args ...interface{}) {
	sl.log(LevelWarn, "WARN", format, args...)
}

// Error logs an error message
func (sl *SimpleLogger) Error(format string, args ...interface{}) {
	sl.log(LevelError, "ERROR", format, args...)
}

// log handles the actual logging
func (sl *SimpleLogger) log(level LogLevel, prefix string, format string, args ...interface{}) {
	sl.mu.RLock()
	if level < sl.level {
		sl.mu.RUnlock()
		return
	}
	sl.mu.RUnlock()

	message := fmt.Sprintf(format, args...)
	sl.logger.Printf("[%s] %s", prefix, message)
}

// LogToFile sets up file logging
func (sl *SimpleLogger) LogToFile(filename string) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	sl.mu.Lock()
	sl.logger = log.New(file, "", log.LstdFlags)
	sl.mu.Unlock()

	return nil
}

// Close closes the logger (for file logging)
func (sl *SimpleLogger) Close() {
	// Nothing to close for stdout logger
	// File closing handled by garbage collector
}

// Global logger instance
var globalLogger = NewSimpleLogger(LevelInfo)

// SetGlobalLevel sets the global log level
func SetGlobalLevel(level LogLevel) {
	globalLogger.SetLevel(level)
}

// Debug logs a debug message globally
func Debug(format string, args ...interface{}) {
	globalLogger.Debug(format, args...)
}

// Info logs an info message globally
func Info(format string, args ...interface{}) {
	globalLogger.Info(format, args...)
}

// Warn logs a warning message globally
func Warn(format string, args ...interface{}) {
	globalLogger.Warn(format, args...)
}

// Error logs an error message globally
func Error(format string, args ...interface{}) {
	globalLogger.Error(format, args...)
}

// LogGameEvent logs a game event in a simple format
func LogGameEvent(eventType string, data map[string]interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	Info("GameEvent [%s] at %s: %v", eventType, timestamp, data)
}

// LogTransaction logs a transaction
func LogTransaction(playerID string, itemName string, quantity int, price float64, profit float64) {
	Info("Transaction: Player=%s Item=%s Qty=%d Price=%.2f Profit=%.2f",
		playerID, itemName, quantity, price, profit)
}

// LogError logs an error with context
func LogError(context string, err error) {
	if err != nil {
		Error("%s: %v", context, err)
	}
}

// GetLevelFromString converts a string to LogLevel
func GetLevelFromString(level string) LogLevel {
	switch level {
	case "debug", "DEBUG":
		return LevelDebug
	case "info", "INFO":
		return LevelInfo
	case "warn", "WARN", "warning", "WARNING":
		return LevelWarn
	case "error", "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}
