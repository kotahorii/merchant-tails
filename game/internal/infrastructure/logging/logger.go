package logging

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogLevel represents the logging level
type LogLevel int

const (
	// DebugLevel logs everything
	DebugLevel LogLevel = iota
	// InfoLevel logs info, warnings, errors
	InfoLevel
	// WarnLevel logs warnings and errors
	WarnLevel
	// ErrorLevel logs only errors
	ErrorLevel
	// FatalLevel logs only fatal errors
	FatalLevel
)

// Logger wraps zerolog for structured logging
type Logger struct {
	logger zerolog.Logger
	level  LogLevel
	fields map[string]interface{}
	mu     sync.RWMutex
}

// LoggerConfig configures the logger
type LoggerConfig struct {
	Level      LogLevel
	OutputPath string
	Console    bool
	JSON       bool
	TimeFormat string
	Context    map[string]interface{}
}

// DefaultConfig returns default logger configuration
func DefaultConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:      InfoLevel,
		Console:    true,
		JSON:       false,
		TimeFormat: time.RFC3339,
		Context:    make(map[string]interface{}),
	}
}

// NewLogger creates a new structured logger
func NewLogger(config *LoggerConfig) (*Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Set global log level
	level := parseZerologLevel(config.Level)
	zerolog.SetGlobalLevel(level)

	// Configure output
	var output io.Writer
	if config.Console {
		if config.JSON {
			output = os.Stdout
		} else {
			output = zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: config.TimeFormat,
			}
		}
	}

	// Add file output if specified
	if config.OutputPath != "" {
		// Create log directory if it doesn't exist
		dir := filepath.Dir(config.OutputPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		file, err := os.OpenFile(config.OutputPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		if output != nil {
			output = io.MultiWriter(output, file)
		} else {
			output = file
		}
	}

	// Create logger with context
	logger := zerolog.New(output).With().
		Timestamp().
		Str("service", "merchant-tails").
		Logger()

	// Add custom context fields
	for key, value := range config.Context {
		logger = logger.With().Interface(key, value).Logger()
	}

	return &Logger{
		logger: logger,
		level:  config.Level,
		fields: make(map[string]interface{}),
	}, nil
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newLogger := &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
		level:  l.level,
		fields: make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	newLogger.fields[key] = value

	return newLogger
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newLogger := &Logger{
		logger: l.logger,
		level:  l.level,
		fields: make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for key, value := range fields {
		newLogger.logger = newLogger.logger.With().Interface(key, value).Logger()
		newLogger.fields[key] = value
	}

	return newLogger
}

// WithContext adds context to the logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	if ctx == nil {
		return l
	}

	// Extract common context values
	fields := make(map[string]interface{})

	// Add request ID if present
	if reqID := ctx.Value("request_id"); reqID != nil {
		fields["request_id"] = reqID
	}

	// Add user ID if present
	if userID := ctx.Value("user_id"); userID != nil {
		fields["user_id"] = userID
	}

	// Add session ID if present
	if sessionID := ctx.Value("session_id"); sessionID != nil {
		fields["session_id"] = sessionID
	}

	if len(fields) > 0 {
		return l.WithFields(fields)
	}

	return l
}

// WithError adds an error field to the logger
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}
	return l.WithField("error", err.Error())
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// Panic logs a panic message and panics
func (l *Logger) Panic(msg string) {
	l.logger.Panic().Msg(msg)
}

// Panicf logs a formatted panic message and panics
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.logger.Panic().Msgf(format, args...)
}

// LogEvent logs a structured event
func (l *Logger) LogEvent(eventType string, fields map[string]interface{}) {
	event := l.logger.Info().
		Str("event_type", eventType).
		Time("timestamp", time.Now())

	for key, value := range fields {
		event = event.Interface(key, value)
	}

	event.Msg("Event occurred")
}

// LogTransaction logs a transaction
func (l *Logger) LogTransaction(transactionID string, itemID string, quantity int, price float64, success bool) {
	l.logger.Info().
		Str("transaction_id", transactionID).
		Str("item_id", itemID).
		Int("quantity", quantity).
		Float64("price", price).
		Bool("success", success).
		Msg("Transaction processed")
}

// LogPerformance logs performance metrics
func (l *Logger) LogPerformance(operation string, duration time.Duration, metadata map[string]interface{}) {
	event := l.logger.Info().
		Str("operation", operation).
		Dur("duration", duration).
		Float64("duration_ms", float64(duration.Milliseconds()))

	for key, value := range metadata {
		event = event.Interface(key, value)
	}

	event.Msg("Performance metric")
}

// LogGameState logs game state changes
func (l *Logger) LogGameState(state string, gold int, level int, reputation float64) {
	l.logger.Info().
		Str("state", state).
		Int("gold", gold).
		Int("level", level).
		Float64("reputation", reputation).
		Msg("Game state changed")
}

// LogMarketEvent logs market events
func (l *Logger) LogMarketEvent(eventType string, itemID string, oldPrice float64, newPrice float64, impact float64) {
	l.logger.Info().
		Str("event_type", eventType).
		Str("item_id", itemID).
		Float64("old_price", oldPrice).
		Float64("new_price", newPrice).
		Float64("impact", impact).
		Msg("Market event")
}

// LogError logs an error with context
func (l *Logger) LogError(err error, operation string, metadata map[string]interface{}) {
	if err == nil {
		return
	}

	// Get caller information
	_, file, line, _ := runtime.Caller(1)

	event := l.logger.Error().
		Err(err).
		Str("operation", operation).
		Str("file", filepath.Base(file)).
		Int("line", line)

	for key, value := range metadata {
		event = event.Interface(key, value)
	}

	event.Msg("Error occurred")
}

// LogSaveOperation logs save/load operations
func (l *Logger) LogSaveOperation(operation string, filename string, success bool, duration time.Duration, size int64) {
	l.logger.Info().
		Str("operation", operation).
		Str("filename", filename).
		Bool("success", success).
		Dur("duration", duration).
		Int64("size_bytes", size).
		Msg("Save operation")
}

// LogQuestEvent logs quest-related events
func (l *Logger) LogQuestEvent(questID string, eventType string, progress float64, rewards map[string]interface{}) {
	event := l.logger.Info().
		Str("quest_id", questID).
		Str("event_type", eventType).
		Float64("progress", progress)

	if rewards != nil {
		event = event.Interface("rewards", rewards)
	}

	event.Msg("Quest event")
}

// LogAchievement logs achievement unlocks
func (l *Logger) LogAchievement(achievementID string, playerLevel int, timestamp time.Time) {
	l.logger.Info().
		Str("achievement_id", achievementID).
		Int("player_level", playerLevel).
		Time("unlocked_at", timestamp).
		Msg("Achievement unlocked")
}

// LogWeatherChange logs weather changes
func (l *Logger) LogWeatherChange(oldWeather string, newWeather string, marketImpact map[string]float64) {
	l.logger.Info().
		Str("old_weather", oldWeather).
		Str("new_weather", newWeather).
		Interface("market_impact", marketImpact).
		Msg("Weather changed")
}

// parseZerologLevel converts LogLevel to zerolog.Level
func parseZerologLevel(level LogLevel) zerolog.Level {
	switch level {
	case DebugLevel:
		return zerolog.DebugLevel
	case InfoLevel:
		return zerolog.InfoLevel
	case WarnLevel:
		return zerolog.WarnLevel
	case ErrorLevel:
		return zerolog.ErrorLevel
	case FatalLevel:
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	return &Logger{
		logger: log.Logger,
		level:  InfoLevel,
		fields: make(map[string]interface{}),
	}
}
