package logging

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Formatter interface for log formatting
type Formatter interface {
	Format(entry *LogEntry) ([]byte, error)
}

// JSONFormatter formats logs as JSON
type JSONFormatter struct {
	PrettyPrint bool
	TimeFormat  string
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		PrettyPrint: false,
		TimeFormat:  time.RFC3339Nano,
	}
}

// Format formats the log entry as JSON
func (f *JSONFormatter) Format(entry *LogEntry) ([]byte, error) {
	// Format timestamp
	formattedEntry := map[string]interface{}{
		"timestamp": entry.Timestamp.Format(f.TimeFormat),
		"level":     entry.Level,
		"message":   entry.Message,
		"service":   "merchant-tails",
	}

	// Merge custom fields
	if len(entry.Fields) > 0 {
		formattedEntry["fields"] = entry.Fields
	}

	var data []byte
	var err error
	if f.PrettyPrint {
		data, err = json.MarshalIndent(formattedEntry, "", "  ")
	} else {
		data, err = json.Marshal(formattedEntry)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Add newline
	data = append(data, '\n')
	return data, nil
}

// TextFormatter formats logs as human-readable text
type TextFormatter struct {
	TimeFormat       string
	DisableColors    bool
	FullTimestamp    bool
	DisableSorting   bool
	QuoteEmptyFields bool
}

// NewTextFormatter creates a new text formatter
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
		TimeFormat:       "2006-01-02 15:04:05",
		DisableColors:    false,
		FullTimestamp:    true,
		DisableSorting:   false,
		QuoteEmptyFields: false,
	}
}

// Format formats the log entry as text
func (f *TextFormatter) Format(entry *LogEntry) ([]byte, error) {
	var sb strings.Builder

	// Add timestamp
	if f.FullTimestamp {
		sb.WriteString(entry.Timestamp.Format(f.TimeFormat))
		sb.WriteString(" ")
	}

	// Add level with color if enabled
	levelStr := entry.Level
	if !f.DisableColors {
		levelStr = f.colorizeLevel(entry.Level)
	}
	sb.WriteString(fmt.Sprintf("[%s]", levelStr))
	sb.WriteString(" ")

	// Add message
	sb.WriteString(entry.Message)

	// Add fields
	for key, value := range entry.Fields {
		sb.WriteString(" ")
		sb.WriteString(key)
		sb.WriteString("=")

		valueStr := fmt.Sprintf("%v", value)
		if f.QuoteEmptyFields && valueStr == "" {
			valueStr = `""`
		} else if strings.Contains(valueStr, " ") {
			valueStr = fmt.Sprintf("%q", valueStr)
		}
		sb.WriteString(valueStr)
	}

	sb.WriteString("\n")
	return []byte(sb.String()), nil
}

// colorizeLevel adds ANSI color codes to log levels
func (f *TextFormatter) colorizeLevel(level string) string {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return "\033[36m" + level + "\033[0m" // Cyan
	case "INFO":
		return "\033[32m" + level + "\033[0m" // Green
	case "WARN":
		return "\033[33m" + level + "\033[0m" // Yellow
	case "ERROR":
		return "\033[31m" + level + "\033[0m" // Red
	case "FATAL":
		return "\033[35m" + level + "\033[0m" // Magenta
	default:
		return level
	}
}

// CompactFormatter formats logs in a compact format
type CompactFormatter struct {
	TimeFormat string
}

// NewCompactFormatter creates a new compact formatter
func NewCompactFormatter() *CompactFormatter {
	return &CompactFormatter{
		TimeFormat: "15:04:05",
	}
}

// Format formats the log entry in compact format
func (f *CompactFormatter) Format(entry *LogEntry) ([]byte, error) {
	// Format: [TIME] LEVEL: MESSAGE key=value ...
	var parts []string

	parts = append(parts, fmt.Sprintf("[%s]", entry.Timestamp.Format(f.TimeFormat)))
	parts = append(parts, fmt.Sprintf("%s:", entry.Level))
	parts = append(parts, entry.Message)

	// Add select fields
	for key, value := range entry.Fields {
		// Only include important fields in compact format
		if key == "duration" || key == "status" || key == "count" || key == "size" {
			parts = append(parts, fmt.Sprintf("%s=%v", key, value))
		}
	}

	result := strings.Join(parts, " ") + "\n"
	return []byte(result), nil
}

// LogstashFormatter formats logs for Logstash/ELK stack
type LogstashFormatter struct {
	AppName string
	Version string
}

// NewLogstashFormatter creates a new Logstash formatter
func NewLogstashFormatter(appName, version string) *LogstashFormatter {
	return &LogstashFormatter{
		AppName: appName,
		Version: version,
	}
}

// Format formats the log entry for Logstash
func (f *LogstashFormatter) Format(entry *LogEntry) ([]byte, error) {
	logstashEntry := map[string]interface{}{
		"@timestamp": entry.Timestamp.Format(time.RFC3339),
		"@version":   f.Version,
		"app":        f.AppName,
		"level":      entry.Level,
		"message":    entry.Message,
		"logger":     "merchant-tails",
	}

	// Add custom fields under "fields" namespace
	if len(entry.Fields) > 0 {
		logstashEntry["fields"] = entry.Fields
	}

	data, err := json.Marshal(logstashEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal logstash entry: %w", err)
	}

	data = append(data, '\n')
	return data, nil
}
