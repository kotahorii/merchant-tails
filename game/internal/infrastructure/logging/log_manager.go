package logging

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// LogManager manages log files, rotation, and buffering
type LogManager struct {
	config         *LogManagerConfig
	currentFile    *os.File
	currentSize    int64
	buffer         chan LogEntry
	wg             sync.WaitGroup
	stopChan       chan struct{}
	logger         *Logger
	mu             sync.RWMutex
	rotationTicker *time.Ticker
}

// LogManagerConfig configures the log manager
type LogManagerConfig struct {
	LogDir          string
	MaxFileSize     int64         // Max size in bytes before rotation
	MaxBackups      int           // Max number of backup files
	MaxAge          int           // Max age in days
	Compress        bool          // Compress rotated files
	BufferSize      int           // Size of log buffer
	FlushInterval   time.Duration // How often to flush buffer
	RotationTime    time.Duration // Time-based rotation (e.g., daily)
	FileNamePattern string        // Log file name pattern
}

// LogEntry represents a buffered log entry
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Fields    map[string]interface{}
}

// DefaultLogManagerConfig returns default configuration
func DefaultLogManagerConfig() *LogManagerConfig {
	return &LogManagerConfig{
		LogDir:          "./logs",
		MaxFileSize:     100 * 1024 * 1024, // 100MB
		MaxBackups:      10,
		MaxAge:          30,
		Compress:        true,
		BufferSize:      1000,
		FlushInterval:   time.Second,
		RotationTime:    24 * time.Hour,
		FileNamePattern: "merchant-tails-%s.log",
	}
}

// NewLogManager creates a new log manager
func NewLogManager(config *LogManagerConfig) (*LogManager, error) {
	if config == nil {
		config = DefaultLogManagerConfig()
	}

	// Create log directory
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open initial log file
	logFile, err := openLogFile(config.LogDir, config.FileNamePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Get current file size
	info, err := logFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat log file: %w", err)
	}

	manager := &LogManager{
		config:      config,
		currentFile: logFile,
		currentSize: info.Size(),
		buffer:      make(chan LogEntry, config.BufferSize),
		stopChan:    make(chan struct{}),
	}

	// Create logger
	loggerConfig := &LoggerConfig{
		Level:      InfoLevel,
		OutputPath: logFile.Name(),
		Console:    false,
		JSON:       true,
	}
	logger, err := NewLogger(loggerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	manager.logger = logger

	// Start background workers
	manager.startWorkers()

	return manager, nil
}

// startWorkers starts background workers for log processing
func (lm *LogManager) startWorkers() {
	// Start buffer flusher
	lm.wg.Add(1)
	go lm.flushWorker()

	// Start rotation ticker if configured
	if lm.config.RotationTime > 0 {
		lm.rotationTicker = time.NewTicker(lm.config.RotationTime)
		lm.wg.Add(1)
		go lm.rotationWorker()
	}

	// Start cleanup worker
	lm.wg.Add(1)
	go lm.cleanupWorker()
}

// Write writes a log entry
func (lm *LogManager) Write(entry LogEntry) {
	select {
	case lm.buffer <- entry:
		// Successfully buffered
	default:
		// Buffer full, write directly
		lm.writeEntry(entry)
	}
}

// writeEntry writes an entry to the current log file
func (lm *LogManager) writeEntry(entry LogEntry) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Format log entry
	logLine := lm.formatEntry(entry)

	// Check if rotation is needed
	if lm.shouldRotate(len(logLine)) {
		if err := lm.rotate(); err != nil {
			fmt.Printf("Failed to rotate log: %v\n", err)
		}
	}

	// Write to file
	n, err := lm.currentFile.WriteString(logLine)
	if err != nil {
		fmt.Printf("Failed to write log: %v\n", err)
		return
	}

	lm.currentSize += int64(n)
}

// formatEntry formats a log entry
func (lm *LogManager) formatEntry(entry LogEntry) string {
	// Simple JSON format
	fields := make([]string, 0, len(entry.Fields)+3)
	fields = append(fields, fmt.Sprintf(`"timestamp":"%s"`, entry.Timestamp.Format(time.RFC3339)))
	fields = append(fields, fmt.Sprintf(`"level":"%s"`, entry.Level))
	fields = append(fields, fmt.Sprintf(`"message":"%s"`, entry.Message))

	for key, value := range entry.Fields {
		fields = append(fields, fmt.Sprintf(`"%s":%v`, key, formatValue(value)))
	}

	return fmt.Sprintf("{%s}\n", strings.Join(fields, ","))
}

// formatValue formats a value for JSON
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, v)
	case int, int64, float64, bool:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf(`"%v"`, v)
	}
}

// shouldRotate checks if log rotation is needed
func (lm *LogManager) shouldRotate(nextWriteSize int) bool {
	return lm.currentSize+int64(nextWriteSize) > lm.config.MaxFileSize
}

// rotate performs log rotation
func (lm *LogManager) rotate() error {
	// Close current file
	if err := lm.currentFile.Close(); err != nil {
		return fmt.Errorf("failed to close current log file: %w", err)
	}

	// Rename current file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	oldPath := lm.currentFile.Name()
	newPath := strings.Replace(oldPath, ".log", fmt.Sprintf("-%s.log", timestamp), 1)

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename log file: %w", err)
	}

	// Compress if configured
	if lm.config.Compress {
		go lm.compressFile(newPath)
	}

	// Open new log file
	logFile, err := openLogFile(lm.config.LogDir, lm.config.FileNamePattern)
	if err != nil {
		return fmt.Errorf("failed to open new log file: %w", err)
	}

	lm.currentFile = logFile
	lm.currentSize = 0

	return nil
}

// compressFile compresses a log file
func (lm *LogManager) compressFile(filePath string) {
	source, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Failed to open file for compression: %v\n", err)
		return
	}
	defer func() { _ = source.Close() }()

	dest, err := os.Create(filePath + ".gz")
	if err != nil {
		fmt.Printf("Failed to create compressed file: %v\n", err)
		return
	}
	defer func() { _ = dest.Close() }()

	gz := gzip.NewWriter(dest)
	defer func() { _ = gz.Close() }()

	if _, err := io.Copy(gz, source); err != nil {
		fmt.Printf("Failed to compress file: %v\n", err)
		return
	}

	// Remove original file after successful compression
	_ = os.Remove(filePath)
}

// flushWorker periodically flushes the buffer
func (lm *LogManager) flushWorker() {
	defer lm.wg.Done()
	ticker := time.NewTicker(lm.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case entry := <-lm.buffer:
			lm.writeEntry(entry)
		case <-ticker.C:
			// Flush any remaining entries
			lm.flush()
		case <-lm.stopChan:
			// Final flush before stopping
			lm.flush()
			return
		}
	}
}

// flush flushes all buffered entries
func (lm *LogManager) flush() {
	for {
		select {
		case entry := <-lm.buffer:
			lm.writeEntry(entry)
		default:
			return
		}
	}
}

// rotationWorker performs time-based rotation
func (lm *LogManager) rotationWorker() {
	defer lm.wg.Done()

	for {
		select {
		case <-lm.rotationTicker.C:
			lm.mu.Lock()
			if err := lm.rotate(); err != nil {
				fmt.Printf("Failed to rotate log: %v\n", err)
			}
			lm.mu.Unlock()
		case <-lm.stopChan:
			return
		}
	}
}

// cleanupWorker removes old log files
func (lm *LogManager) cleanupWorker() {
	defer lm.wg.Done()
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lm.cleanup()
		case <-lm.stopChan:
			return
		}
	}
}

// cleanup removes old log files
func (lm *LogManager) cleanup() {
	files, err := filepath.Glob(filepath.Join(lm.config.LogDir, "*.log*"))
	if err != nil {
		fmt.Printf("Failed to list log files: %v\n", err)
		return
	}

	// Sort files by modification time
	type fileInfo struct {
		path    string
		modTime time.Time
	}

	var fileInfos []fileInfo
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, fileInfo{
			path:    file,
			modTime: info.ModTime(),
		})
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].modTime.After(fileInfos[j].modTime)
	})

	// Remove old files
	cutoffTime := time.Now().AddDate(0, 0, -lm.config.MaxAge)
	keepCount := 0

	for _, fi := range fileInfos {
		keepCount++

		// Skip current log file
		if fi.path == lm.currentFile.Name() {
			continue
		}

		// Remove if too old or exceeds max backups
		if fi.modTime.Before(cutoffTime) || keepCount > lm.config.MaxBackups {
			_ = os.Remove(fi.path)
		}
	}
}

// Close closes the log manager
func (lm *LogManager) Close() error {
	close(lm.stopChan)

	// Stop rotation ticker
	if lm.rotationTicker != nil {
		lm.rotationTicker.Stop()
	}

	// Wait for workers to finish
	lm.wg.Wait()

	// Close current file
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.currentFile != nil {
		return lm.currentFile.Close()
	}

	return nil
}

// openLogFile opens or creates a log file
func openLogFile(logDir, pattern string) (*os.File, error) {
	timestamp := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf(pattern, timestamp)
	filepath := filepath.Join(logDir, filename)

	return os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
}
