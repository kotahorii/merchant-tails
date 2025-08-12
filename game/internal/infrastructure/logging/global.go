package logging

import (
	"context"
	"sync"
	"time"
)

var (
	globalLogger     *Logger
	globalLogManager *LogManager
	once             sync.Once
)

// Initialize initializes the global logger
func Initialize(config *LoggerConfig, managerConfig *LogManagerConfig) error {
	var err error
	once.Do(func() {
		// Create log manager if config provided
		if managerConfig != nil {
			globalLogManager, err = NewLogManager(managerConfig)
			if err != nil {
				return
			}
		}

		// Create logger
		if config == nil {
			config = DefaultConfig()
		}
		globalLogger, err = NewLogger(config)
	})
	return err
}

// InitializeDefault initializes with default configuration
func InitializeDefault() error {
	return Initialize(DefaultConfig(), DefaultLogManagerConfig())
}

// Get returns the global logger
func Get() *Logger {
	if globalLogger == nil {
		// Initialize with defaults if not already initialized
		_ = InitializeDefault()
	}
	return globalLogger
}

// SetLogger sets the global logger
func SetLogger(logger *Logger) {
	globalLogger = logger
}

// Close closes the global logger and manager
func Close() error {
	if globalLogManager != nil {
		return globalLogManager.Close()
	}
	return nil
}

// Helper functions for direct logging

// Debug logs a debug message
func Debug(msg string) {
	Get().Debug(msg)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	Get().Debugf(format, args...)
}

// Info logs an info message
func Info(msg string) {
	Get().Info(msg)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	Get().Infof(format, args...)
}

// Warn logs a warning message
func Warn(msg string) {
	Get().Warn(msg)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	Get().Warnf(format, args...)
}

// Error logs an error message
func Error(msg string) {
	Get().Error(msg)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	Get().Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string) {
	Get().Fatal(msg)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	Get().Fatalf(format, args...)
}

// WithField adds a field to the logger
func WithField(key string, value interface{}) *Logger {
	return Get().WithField(key, value)
}

// WithFields adds multiple fields to the logger
func WithFields(fields map[string]interface{}) *Logger {
	return Get().WithFields(fields)
}

// WithContext adds context to the logger
func WithContext(ctx context.Context) *Logger {
	return Get().WithContext(ctx)
}

// WithError adds an error field to the logger
func WithError(err error) *Logger {
	return Get().WithError(err)
}

// LogEvent logs a structured event
func LogEvent(eventType string, fields map[string]interface{}) {
	Get().LogEvent(eventType, fields)
}

// LogTransaction logs a transaction
func LogTransaction(transactionID string, itemID string, quantity int, price float64, success bool) {
	Get().LogTransaction(transactionID, itemID, quantity, price, success)
}

// LogPerformance logs performance metrics
func LogPerformance(operation string, duration time.Duration, metadata map[string]interface{}) {
	Get().LogPerformance(operation, duration, metadata)
}

// LogGameState logs game state changes
func LogGameState(state string, gold int, level int, reputation float64) {
	Get().LogGameState(state, gold, level, reputation)
}

// LogMarketEvent logs market events
func LogMarketEvent(eventType string, itemID string, oldPrice float64, newPrice float64, impact float64) {
	Get().LogMarketEvent(eventType, itemID, oldPrice, newPrice, impact)
}

// LogError logs an error with context
func LogError(err error, operation string, metadata map[string]interface{}) {
	Get().LogError(err, operation, metadata)
}

// LogSaveOperation logs save/load operations
func LogSaveOperation(operation string, filename string, success bool, duration time.Duration, size int64) {
	Get().LogSaveOperation(operation, filename, success, duration, size)
}

// LogQuestEvent logs quest-related events
func LogQuestEvent(questID string, eventType string, progress float64, rewards map[string]interface{}) {
	Get().LogQuestEvent(questID, eventType, progress, rewards)
}

// LogAchievement logs achievement unlocks
func LogAchievement(achievementID string, playerLevel int, timestamp time.Time) {
	Get().LogAchievement(achievementID, playerLevel, timestamp)
}

// LogWeatherChange logs weather changes
func LogWeatherChange(oldWeather string, newWeather string, marketImpact map[string]float64) {
	Get().LogWeatherChange(oldWeather, newWeather, marketImpact)
}
