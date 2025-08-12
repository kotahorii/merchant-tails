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

// RotationConfig configures log rotation
type RotationConfig struct {
	MaxSize    int64         // Maximum size in bytes before rotation (default: 100MB)
	MaxAge     time.Duration // Maximum age before rotation (default: 24h)
	MaxBackups int           // Maximum number of backup files to keep (default: 7)
	Compress   bool          // Whether to compress rotated files (default: true)
	LocalTime  bool          // Use local time for rotation filenames (default: true)
}

// DefaultRotationConfig returns default rotation configuration
func DefaultRotationConfig() *RotationConfig {
	return &RotationConfig{
		MaxSize:    100 * 1024 * 1024, // 100MB
		MaxAge:     24 * time.Hour,
		MaxBackups: 7,
		Compress:   true,
		LocalTime:  true,
	}
}

// RotatingFileWriter is a thread-safe writer that rotates log files
type RotatingFileWriter struct {
	config       *RotationConfig
	filename     string
	file         *os.File
	size         int64
	lastRotation time.Time
	mu           sync.Mutex
}

// NewRotatingFileWriter creates a new rotating file writer
func NewRotatingFileWriter(filename string, config *RotationConfig) (*RotatingFileWriter, error) {
	if config == nil {
		config = DefaultRotationConfig()
	}

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	rfw := &RotatingFileWriter{
		config:       config,
		filename:     filename,
		lastRotation: time.Now(),
	}

	// Open or create the log file
	if err := rfw.openFile(); err != nil {
		return nil, err
	}

	// Start rotation checker
	go rfw.rotationChecker()

	return rfw, nil
}

// Write implements io.Writer interface
func (rfw *RotatingFileWriter) Write(p []byte) (n int, err error) {
	rfw.mu.Lock()
	defer rfw.mu.Unlock()

	// Check if rotation is needed
	if rfw.shouldRotate() {
		if err := rfw.rotate(); err != nil {
			return 0, fmt.Errorf("failed to rotate log: %w", err)
		}
	}

	// Write to file
	n, err = rfw.file.Write(p)
	if err != nil {
		return n, err
	}

	rfw.size += int64(n)
	return n, nil
}

// Close closes the file writer
func (rfw *RotatingFileWriter) Close() error {
	rfw.mu.Lock()
	defer rfw.mu.Unlock()

	if rfw.file != nil {
		return rfw.file.Close()
	}
	return nil
}

// shouldRotate checks if rotation is needed
func (rfw *RotatingFileWriter) shouldRotate() bool {
	// Check size limit
	if rfw.config.MaxSize > 0 && rfw.size >= rfw.config.MaxSize {
		return true
	}

	// Check age limit
	if rfw.config.MaxAge > 0 && time.Since(rfw.lastRotation) >= rfw.config.MaxAge {
		return true
	}

	return false
}

// rotate performs log rotation
func (rfw *RotatingFileWriter) rotate() error {
	// Close current file
	if rfw.file != nil {
		if err := rfw.file.Close(); err != nil {
			return fmt.Errorf("failed to close current log file: %w", err)
		}
	}

	// Generate rotation filename
	rotationName := rfw.rotationFilename()

	// Rename current file
	if err := os.Rename(rfw.filename, rotationName); err != nil {
		return fmt.Errorf("failed to rename log file: %w", err)
	}

	// Compress if configured
	if rfw.config.Compress {
		go rfw.compressFile(rotationName)
	}

	// Clean up old files
	go rfw.cleanupOldFiles()

	// Open new file
	if err := rfw.openFile(); err != nil {
		return fmt.Errorf("failed to open new log file: %w", err)
	}

	rfw.lastRotation = time.Now()
	return nil
}

// openFile opens or creates the log file
func (rfw *RotatingFileWriter) openFile() error {
	file, err := os.OpenFile(rfw.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Get file size
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to stat log file: %w", err)
	}

	rfw.file = file
	rfw.size = info.Size()
	return nil
}

// rotationFilename generates a filename for rotation
func (rfw *RotatingFileWriter) rotationFilename() string {
	var t time.Time
	if rfw.config.LocalTime {
		t = time.Now()
	} else {
		t = time.Now().UTC()
	}

	// Format: filename.2006-01-02T15-04-05.log
	ext := filepath.Ext(rfw.filename)
	name := strings.TrimSuffix(rfw.filename, ext)
	timestamp := t.Format("2006-01-02T15-04-05")

	return fmt.Sprintf("%s.%s%s", name, timestamp, ext)
}

// compressFile compresses a rotated log file
func (rfw *RotatingFileWriter) compressFile(filename string) {
	// Open source file
	src, err := os.Open(filename)
	if err != nil {
		return
	}
	defer func() { _ = src.Close() }()

	// Create compressed file
	dst, err := os.Create(filename + ".gz")
	if err != nil {
		return
	}
	defer func() { _ = dst.Close() }()

	// Create gzip writer
	gz := gzip.NewWriter(dst)
	defer func() { _ = gz.Close() }()

	// Copy data
	if _, err := io.Copy(gz, src); err != nil {
		return
	}

	// Remove original file after successful compression
	_ = os.Remove(filename)
}

// cleanupOldFiles removes old backup files
func (rfw *RotatingFileWriter) cleanupOldFiles() {
	if rfw.config.MaxBackups <= 0 {
		return
	}

	// Get directory and base name
	dir := filepath.Dir(rfw.filename)
	base := filepath.Base(rfw.filename)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Find all backup files
	var backups []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Check if it's a backup file
		filename := filepath.Base(path)
		if strings.HasPrefix(filename, name+".") && filename != base {
			// Check if it's a log file or compressed log file
			if strings.HasSuffix(filename, ext) || strings.HasSuffix(filename, ext+".gz") {
				backups = append(backups, path)
			}
		}

		return nil
	})

	// Sort by modification time (oldest first)
	sort.Slice(backups, func(i, j int) bool {
		infoI, _ := os.Stat(backups[i])
		infoJ, _ := os.Stat(backups[j])
		if infoI == nil || infoJ == nil {
			return false
		}
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// Remove old backups
	if len(backups) > rfw.config.MaxBackups {
		for _, backup := range backups[:len(backups)-rfw.config.MaxBackups] {
			_ = os.Remove(backup)
		}
	}
}

// rotationChecker periodically checks if rotation is needed
func (rfw *RotatingFileWriter) rotationChecker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rfw.mu.Lock()
		if rfw.shouldRotate() {
			_ = rfw.rotate()
		}
		rfw.mu.Unlock()
	}
}

// RotationHook is a log hook that triggers rotation
type RotationHook struct {
	writer *RotatingFileWriter
	levels []LogLevel
}

// NewRotationHook creates a new rotation hook
func NewRotationHook(writer *RotatingFileWriter, levels ...LogLevel) *RotationHook {
	if len(levels) == 0 {
		levels = []LogLevel{
			DebugLevel,
			InfoLevel,
			WarnLevel,
			ErrorLevel,
			FatalLevel,
		}
	}
	return &RotationHook{
		writer: writer,
		levels: levels,
	}
}

// Fire writes the log entry to the rotating file
func (h *RotationHook) Fire(entry *LogEntry) error {
	// Format entry as JSON
	formatter := NewJSONFormatter()
	data, err := formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = h.writer.Write(data)
	return err
}

// Levels returns the levels this hook is interested in
func (h *RotationHook) Levels() []LogLevel {
	return h.levels
}

// SizeBasedRotationStrategy rotates based on file size
type SizeBasedRotationStrategy struct {
	MaxSize int64
}

// ShouldRotate checks if rotation is needed based on size
func (s *SizeBasedRotationStrategy) ShouldRotate(size int64, age time.Duration) bool {
	return size >= s.MaxSize
}

// TimeBasedRotationStrategy rotates based on time
type TimeBasedRotationStrategy struct {
	MaxAge time.Duration
}

// ShouldRotate checks if rotation is needed based on age
func (t *TimeBasedRotationStrategy) ShouldRotate(size int64, age time.Duration) bool {
	return age >= t.MaxAge
}

// DailyRotationStrategy rotates daily at a specific time
type DailyRotationStrategy struct {
	Hour         int
	Minute       int
	lastRotation time.Time
}

// ShouldRotate checks if daily rotation is needed
func (d *DailyRotationStrategy) ShouldRotate(size int64, age time.Duration) bool {
	now := time.Now()
	rotationTime := time.Date(now.Year(), now.Month(), now.Day(), d.Hour, d.Minute, 0, 0, now.Location())

	// If rotation time has passed today and we haven't rotated yet
	if now.After(rotationTime) && d.lastRotation.Before(rotationTime) {
		d.lastRotation = now
		return true
	}

	return false
}
