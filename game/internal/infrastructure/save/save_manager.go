package save

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	savepb "github.com/yourusername/merchant-tails/game/internal/gen/save"
)

const (
	// Save file constants
	SaveFileExtension = ".mtsave"
	AutoSavePrefix    = "auto_"
	QuickSavePrefix   = "quick_"
	SaveVersion       = "1.0.0"

	// Encryption key (in production, this should be properly managed)
	encryptionKey = "merchant-tails-save-encryption-key-32bytes!!"
)

var (
	ErrSaveNotFound     = errors.New("save file not found")
	ErrCorruptedSave    = errors.New("save file is corrupted")
	ErrIncompatibleSave = errors.New("save file version is incompatible")
	ErrSaveInProgress   = errors.New("save operation already in progress")
	ErrLoadInProgress   = errors.New("load operation already in progress")
	ErrAutoSaveDisabled = errors.New("auto-save is disabled")
	ErrMaxSavesReached  = errors.New("maximum number of saves reached")
)

// SaveManager manages game saves
type SaveManager struct {
	saveDir           string
	currentSave       *savepb.GameState
	autoSaveEnabled   bool
	autoSaveInterval  time.Duration
	maxSaves          int
	maxAutoSaves      int
	compressionLevel  int
	encryptionEnabled bool

	// Auto-save management
	autoSaveTicker *time.Ticker
	autoSaveStop   chan bool
	lastAutoSave   time.Time

	// Thread safety
	saveMutex sync.Mutex
	loadMutex sync.Mutex

	// Callbacks
	onSaveStart    func()
	onSaveComplete func(string)
	onSaveError    func(error)
	onLoadStart    func()
	onLoadComplete func()
	onLoadError    func(error)
}

// SaveOptions contains options for saving
type SaveOptions struct {
	Compress  bool
	Encrypt   bool
	Overwrite bool
	Metadata  map[string]string
}

// SaveInfo contains information about a save file
type SaveInfo struct {
	FileName     string
	FilePath     string
	SaveTime     time.Time
	FileSize     int64
	Version      string
	PlayerName   string
	DayNumber    int32
	Gold         float64
	IsAutoSave   bool
	IsQuickSave  bool
	IsCompressed bool
	IsEncrypted  bool
}

// NewSaveManager creates a new save manager
func NewSaveManager(saveDir string) *SaveManager {
	sm := &SaveManager{
		saveDir:           saveDir,
		autoSaveEnabled:   true,
		autoSaveInterval:  5 * time.Minute,
		maxSaves:          100,
		maxAutoSaves:      5,
		compressionLevel:  gzip.DefaultCompression,
		encryptionEnabled: true,
		autoSaveStop:      make(chan bool),
	}

	// Ensure save directory exists
	_ = os.MkdirAll(saveDir, 0o755)

	return sm
}

// Save saves the current game state
func (sm *SaveManager) Save(state *savepb.GameState, fileName string, options *SaveOptions) error {
	sm.saveMutex.Lock()
	defer sm.saveMutex.Unlock()

	if sm.onSaveStart != nil {
		sm.onSaveStart()
	}

	// Set save metadata
	state.Version = SaveVersion
	state.SaveTime = timestamppb.Now()

	// Serialize the game state
	data, err := proto.Marshal(state)
	if err != nil {
		if sm.onSaveError != nil {
			sm.onSaveError(err)
		}
		return fmt.Errorf("failed to serialize game state: %w", err)
	}

	// Apply compression if requested
	if options != nil && options.Compress {
		data, err = sm.compressData(data)
		if err != nil {
			if sm.onSaveError != nil {
				sm.onSaveError(err)
			}
			return fmt.Errorf("failed to compress save data: %w", err)
		}
	}

	// Apply encryption if requested
	if options != nil && options.Encrypt && sm.encryptionEnabled {
		data, err = sm.encryptData(data)
		if err != nil {
			if sm.onSaveError != nil {
				sm.onSaveError(err)
			}
			return fmt.Errorf("failed to encrypt save data: %w", err)
		}
	}

	// Ensure file has correct extension
	if filepath.Ext(fileName) != SaveFileExtension {
		fileName += SaveFileExtension
	}

	// Create full file path
	filePath := filepath.Join(sm.saveDir, fileName)

	// Check if file exists and overwrite is not allowed
	if options != nil && !options.Overwrite {
		if _, err := os.Stat(filePath); err == nil {
			return errors.New("save file already exists")
		}
	}

	// Write to file
	err = os.WriteFile(filePath, data, 0o644)
	if err != nil {
		if sm.onSaveError != nil {
			sm.onSaveError(err)
		}
		return fmt.Errorf("failed to write save file: %w", err)
	}

	// Update current save
	sm.currentSave = state

	if sm.onSaveComplete != nil {
		sm.onSaveComplete(filePath)
	}

	// Clean up old auto-saves if this is an auto-save
	if strings.HasPrefix(fileName, AutoSavePrefix) {
		sm.cleanupAutoSaves()
	}

	return nil
}

// Load loads a game state from a save file
func (sm *SaveManager) Load(fileName string) (*savepb.GameState, error) {
	sm.loadMutex.Lock()
	defer sm.loadMutex.Unlock()

	if sm.onLoadStart != nil {
		sm.onLoadStart()
	}

	// Ensure file has correct extension
	if filepath.Ext(fileName) != SaveFileExtension {
		fileName += SaveFileExtension
	}

	// Create full file path
	filePath := filepath.Join(sm.saveDir, fileName)

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			if sm.onLoadError != nil {
				sm.onLoadError(ErrSaveNotFound)
			}
			return nil, ErrSaveNotFound
		}
		if sm.onLoadError != nil {
			sm.onLoadError(err)
		}
		return nil, fmt.Errorf("failed to read save file: %w", err)
	}

	// Try to decrypt if enabled
	if sm.encryptionEnabled {
		decrypted, err := sm.decryptData(data)
		if err == nil {
			data = decrypted
		}
		// If decryption fails, assume data is not encrypted
	}

	// Try to decompress
	decompressed, err := sm.decompressData(data)
	if err == nil {
		data = decompressed
	}
	// If decompression fails, assume data is not compressed

	// Deserialize the game state
	state := &savepb.GameState{}
	err = proto.Unmarshal(data, state)
	if err != nil {
		if sm.onLoadError != nil {
			sm.onLoadError(ErrCorruptedSave)
		}
		return nil, ErrCorruptedSave
	}

	// Check version compatibility
	if !sm.isVersionCompatible(state.Version) {
		if sm.onLoadError != nil {
			sm.onLoadError(ErrIncompatibleSave)
		}
		return nil, ErrIncompatibleSave
	}

	// Update current save
	sm.currentSave = state

	if sm.onLoadComplete != nil {
		sm.onLoadComplete()
	}

	return state, nil
}

// QuickSave performs a quick save
func (sm *SaveManager) QuickSave(state *savepb.GameState) error {
	fileName := fmt.Sprintf("%s%s", QuickSavePrefix, time.Now().Format("20060102_150405"))
	options := &SaveOptions{
		Compress:  true,
		Encrypt:   sm.encryptionEnabled,
		Overwrite: true,
	}
	return sm.Save(state, fileName, options)
}

// AutoSave performs an auto-save
func (sm *SaveManager) AutoSave(state *savepb.GameState) error {
	if !sm.autoSaveEnabled {
		return ErrAutoSaveDisabled
	}

	fileName := fmt.Sprintf("%s%s", AutoSavePrefix, time.Now().Format("20060102_150405"))
	options := &SaveOptions{
		Compress:  true,
		Encrypt:   sm.encryptionEnabled,
		Overwrite: true,
	}

	err := sm.Save(state, fileName, options)
	if err == nil {
		sm.lastAutoSave = time.Now()
	}
	return err
}

// StartAutoSave starts the auto-save timer
func (sm *SaveManager) StartAutoSave(stateProvider func() *savepb.GameState) {
	if !sm.autoSaveEnabled || sm.autoSaveTicker != nil {
		return
	}

	sm.autoSaveTicker = time.NewTicker(sm.autoSaveInterval)
	go func() {
		for {
			select {
			case <-sm.autoSaveTicker.C:
				if state := stateProvider(); state != nil {
					_ = sm.AutoSave(state)
				}
			case <-sm.autoSaveStop:
				return
			}
		}
	}()
}

// StopAutoSave stops the auto-save timer
func (sm *SaveManager) StopAutoSave() {
	if sm.autoSaveTicker != nil {
		sm.autoSaveTicker.Stop()
		sm.autoSaveStop <- true
		sm.autoSaveTicker = nil
	}
}

// ListSaves returns a list of all save files
func (sm *SaveManager) ListSaves() ([]*SaveInfo, error) {
	files, err := os.ReadDir(sm.saveDir)
	if err != nil {
		return nil, err
	}

	saves := make([]*SaveInfo, 0)
	for _, file := range files {
		if filepath.Ext(file.Name()) != SaveFileExtension {
			continue
		}

		info, err := sm.GetSaveInfo(file.Name())
		if err != nil {
			continue
		}
		saves = append(saves, info)
	}

	return saves, nil
}

// GetSaveInfo returns information about a save file
func (sm *SaveManager) GetSaveInfo(fileName string) (*SaveInfo, error) {
	filePath := filepath.Join(sm.saveDir, fileName)
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	// Try to load the save to get metadata
	state, err := sm.Load(fileName)
	if err != nil {
		return nil, err
	}

	info := &SaveInfo{
		FileName:     fileName,
		FilePath:     filePath,
		SaveTime:     state.SaveTime.AsTime(),
		FileSize:     stat.Size(),
		Version:      state.Version,
		IsAutoSave:   strings.HasPrefix(fileName, AutoSavePrefix),
		IsQuickSave:  strings.HasPrefix(fileName, QuickSavePrefix),
		IsCompressed: true, // We always compress
		IsEncrypted:  sm.encryptionEnabled,
	}

	// Extract player info if available
	if state.Player != nil {
		info.PlayerName = state.Player.Name
		info.DayNumber = state.Player.DayNumber
		info.Gold = state.Player.Gold
	}

	return info, nil
}

// DeleteSave deletes a save file
func (sm *SaveManager) DeleteSave(fileName string) error {
	if filepath.Ext(fileName) != SaveFileExtension {
		fileName += SaveFileExtension
	}
	filePath := filepath.Join(sm.saveDir, fileName)
	return os.Remove(filePath)
}

// compressData compresses data using gzip
func (sm *SaveManager) compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, sm.compressionLevel)
	if err != nil {
		return nil, err
	}
	defer func() { _ = gz.Close() }()

	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decompressData decompresses gzip data
func (sm *SaveManager) decompressData(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()

	return io.ReadAll(r)
}

// encryptData encrypts data using AES
func (sm *SaveManager) encryptData(data []byte) ([]byte, error) {
	key := sha256.Sum256([]byte(encryptionKey))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	encrypted := gcm.Seal(nonce, nonce, data, nil)
	return []byte(base64.StdEncoding.EncodeToString(encrypted)), nil
}

// decryptData decrypts AES encrypted data
func (sm *SaveManager) decryptData(data []byte) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, err
	}

	key := sha256.Sum256([]byte(encryptionKey))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(decoded) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := decoded[:nonceSize], decoded[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// isVersionCompatible checks if a save version is compatible
func (sm *SaveManager) isVersionCompatible(version string) bool {
	// For now, we only support exact version match
	// In the future, we could support migration between versions
	return version == SaveVersion
}

// cleanupAutoSaves removes old auto-saves keeping only the most recent ones
func (sm *SaveManager) cleanupAutoSaves() {
	saves, err := sm.ListSaves()
	if err != nil {
		return
	}

	autoSaves := make([]*SaveInfo, 0)
	for _, save := range saves {
		if save.IsAutoSave {
			autoSaves = append(autoSaves, save)
		}
	}

	// Sort by save time (newest first)
	sort.Slice(autoSaves, func(i, j int) bool {
		return autoSaves[i].SaveTime.After(autoSaves[j].SaveTime)
	})

	// Delete old auto-saves
	for i := sm.maxAutoSaves; i < len(autoSaves); i++ {
		_ = sm.DeleteSave(autoSaves[i].FileName)
	}
}

// SetAutoSaveInterval sets the auto-save interval
func (sm *SaveManager) SetAutoSaveInterval(interval time.Duration) {
	sm.autoSaveInterval = interval
	if sm.autoSaveTicker != nil {
		sm.autoSaveTicker.Reset(interval)
	}
}

// SetAutoSaveEnabled enables or disables auto-save
func (sm *SaveManager) SetAutoSaveEnabled(enabled bool) {
	sm.autoSaveEnabled = enabled
}

// SetEncryptionEnabled enables or disables encryption
func (sm *SaveManager) SetEncryptionEnabled(enabled bool) {
	sm.encryptionEnabled = enabled
}

// SetCallbacks sets callback functions for save/load events
func (sm *SaveManager) SetCallbacks(
	onSaveStart func(),
	onSaveComplete func(string),
	onSaveError func(error),
	onLoadStart func(),
	onLoadComplete func(),
	onLoadError func(error),
) {
	sm.onSaveStart = onSaveStart
	sm.onSaveComplete = onSaveComplete
	sm.onSaveError = onSaveError
	sm.onLoadStart = onLoadStart
	sm.onLoadComplete = onLoadComplete
	sm.onLoadError = onLoadError
}
