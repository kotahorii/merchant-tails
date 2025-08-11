package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Default settings values
const (
	DefaultGameSpeed           = 1.0
	DefaultAutoSave            = true
	DefaultAutoSaveInterval    = 5
	DefaultMusicVolume         = 0.7
	DefaultSFXVolume           = 0.8
	DefaultLanguage            = "en"
	DefaultDifficulty          = "normal"
	DefaultFullscreen          = false
	DefaultVSync               = true
	DefaultAntiAliasing        = true
	DefaultShadowQuality       = "high"
	DefaultTextureQuality      = "high"
	DefaultEffectsQuality      = "high"
	DefaultNotifications       = true
	DefaultTutorialHints       = true
	DefaultConfirmationDialogs = true
	DefaultPauseOnFocusLoss    = true
	DefaultShowFPS             = false
	DefaultShowDebugInfo       = false
)

// Settings categories
type SettingsCategory string

const (
	CategoryGame     SettingsCategory = "game"
	CategoryGraphics SettingsCategory = "graphics"
	CategoryAudio    SettingsCategory = "audio"
	CategoryControls SettingsCategory = "controls"
	CategoryUI       SettingsCategory = "ui"
	CategoryAdvanced SettingsCategory = "advanced"
)

// Setting keys
const (
	SettingGameSpeed         = "game_speed"
	SettingMusicVolume       = "music_volume"
	SettingSFXVolume         = "sfx_volume"
	SettingDifficulty        = "difficulty"
	SettingAutoSave          = "auto_save"
	SettingAutoSaveInt       = "auto_save_interval"
	SettingLanguage          = "language"
	SettingFullscreen        = "fullscreen"
	SettingVSync             = "vsync"
	SettingTargetFPS         = "target_fps"
	SettingShadowQuality     = "shadow_quality"
	SettingTextureQuality    = "texture_quality"
	SettingEffectsQuality    = "effects_quality"
	SettingMasterVolume      = "master_volume"
	SettingUIVolume          = "ui_volume"
	SettingAmbientVolume     = "ambient_volume"
	SettingShowFPS           = "show_fps"
	SettingShowNotifications = "show_notifications"
	SettingShowTutorialHints = "show_tutorial_hints"
)

// Errors
var (
	ErrInvalidSetting     = errors.New("invalid setting value")
	ErrSettingNotFound    = errors.New("setting not found")
	ErrSettingsLocked     = errors.New("settings are locked")
	ErrInvalidRange       = errors.New("value out of valid range")
	ErrInvalidType        = errors.New("invalid value type")
	ErrSaveSettingsFailed = errors.New("failed to save settings")
	ErrLoadSettingsFailed = errors.New("failed to load settings")
)

// GameSettings contains all game settings
type GameSettings struct {
	// Game settings
	GameSpeed        float64 `json:"game_speed"`
	Difficulty       string  `json:"difficulty"`
	AutoSave         bool    `json:"auto_save"`
	AutoSaveInterval int     `json:"auto_save_interval_minutes"`
	PauseOnFocusLoss bool    `json:"pause_on_focus_loss"`
	SkipAnimations   bool    `json:"skip_animations"`
	FastForwardSpeed float64 `json:"fast_forward_speed"`
	Language         string  `json:"language"`
	Currency         string  `json:"currency"`
	DateFormat       string  `json:"date_format"`

	// Graphics settings
	Resolution     Resolution `json:"resolution"`
	Fullscreen     bool       `json:"fullscreen"`
	VSync          bool       `json:"vsync"`
	TargetFPS      int        `json:"target_fps"`
	AntiAliasing   bool       `json:"anti_aliasing"`
	ShadowQuality  string     `json:"shadow_quality"`
	TextureQuality string     `json:"texture_quality"`
	EffectsQuality string     `json:"effects_quality"`
	ParticleCount  int        `json:"particle_count"`
	UIScale        float64    `json:"ui_scale"`

	// Audio settings
	MasterVolume       float64 `json:"master_volume"`
	MusicVolume        float64 `json:"music_volume"`
	SFXVolume          float64 `json:"sfx_volume"`
	UIVolume           float64 `json:"ui_volume"`
	AmbientVolume      float64 `json:"ambient_volume"`
	MuteWhenUnfocused  bool    `json:"mute_when_unfocused"`
	EnableDynamicMusic bool    `json:"enable_dynamic_music"`

	// Control settings
	MouseSensitivity float64           `json:"mouse_sensitivity"`
	InvertMouseY     bool              `json:"invert_mouse_y"`
	EdgeScrolling    bool              `json:"edge_scrolling"`
	EdgeScrollSpeed  float64           `json:"edge_scroll_speed"`
	KeyBindings      map[string]string `json:"key_bindings"`
	GamepadEnabled   bool              `json:"gamepad_enabled"`
	GamepadVibration bool              `json:"gamepad_vibration"`

	// UI settings
	ShowNotifications    bool `json:"show_notifications"`
	NotificationDuration int  `json:"notification_duration_seconds"`
	ShowTutorialHints    bool `json:"show_tutorial_hints"`
	ShowTooltips         bool `json:"show_tooltips"`
	TooltipDelay         int  `json:"tooltip_delay_ms"`
	ConfirmationDialogs  bool `json:"confirmation_dialogs"`
	ShowFPS              bool `json:"show_fps"`
	ShowClock            bool `json:"show_clock"`
	ShowMinimap          bool `json:"show_minimap"`
	MinimapSize          int  `json:"minimap_size"`

	// Advanced settings
	EnableDebugMode bool   `json:"enable_debug_mode"`
	ShowDebugInfo   bool   `json:"show_debug_info"`
	EnableConsole   bool   `json:"enable_console"`
	LogLevel        string `json:"log_level"`
	MaxAutoSaves    int    `json:"max_auto_saves"`
	CompressSaves   bool   `json:"compress_saves"`
	EncryptSaves    bool   `json:"encrypt_saves"`
	NetworkTimeout  int    `json:"network_timeout_seconds"`
	CacheSize       int    `json:"cache_size_mb"`

	// Accessibility settings
	ColorblindMode      string `json:"colorblind_mode"`
	HighContrast        bool   `json:"high_contrast"`
	ScreenReaderEnabled bool   `json:"screen_reader_enabled"`
	SubtitlesEnabled    bool   `json:"subtitles_enabled"`
	SubtitleSize        int    `json:"subtitle_size"`

	// Custom settings
	CustomSettings map[string]interface{} `json:"custom_settings"`

	// Metadata
	Version      string    `json:"version"`
	LastModified time.Time `json:"last_modified"`
}

// Resolution represents a screen resolution
type Resolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// SettingsManager manages game settings
type SettingsManager struct {
	settings        *GameSettings
	defaultSettings *GameSettings
	settingsPath    string
	autoSave        bool
	locked          bool

	// Validation rules
	validators map[string]SettingValidator

	// Change callbacks
	changeCallbacks map[string][]SettingChangeCallback

	// Thread safety
	mu sync.RWMutex
}

// SettingValidator validates a setting value
type SettingValidator func(value interface{}) error

// SettingChangeCallback is called when a setting changes
type SettingChangeCallback func(oldValue, newValue interface{})

// NewSettingsManager creates a new settings manager
func NewSettingsManager(settingsPath string) *SettingsManager {
	sm := &SettingsManager{
		settingsPath:    settingsPath,
		autoSave:        true,
		validators:      make(map[string]SettingValidator),
		changeCallbacks: make(map[string][]SettingChangeCallback),
	}

	// Initialize default settings
	sm.defaultSettings = sm.createDefaultSettings()
	sm.settings = sm.createDefaultSettings()

	// Register validators
	sm.registerValidators()

	return sm
}

// createDefaultSettings creates default game settings
func (sm *SettingsManager) createDefaultSettings() *GameSettings {
	return &GameSettings{
		// Game
		GameSpeed:        DefaultGameSpeed,
		Difficulty:       DefaultDifficulty,
		AutoSave:         DefaultAutoSave,
		AutoSaveInterval: DefaultAutoSaveInterval,
		PauseOnFocusLoss: DefaultPauseOnFocusLoss,
		SkipAnimations:   false,
		FastForwardSpeed: 5.0,
		Language:         DefaultLanguage,
		Currency:         "gold",
		DateFormat:       "MM/DD/YYYY",

		// Graphics
		Resolution:     Resolution{Width: 1920, Height: 1080},
		Fullscreen:     DefaultFullscreen,
		VSync:          DefaultVSync,
		TargetFPS:      60,
		AntiAliasing:   DefaultAntiAliasing,
		ShadowQuality:  DefaultShadowQuality,
		TextureQuality: DefaultTextureQuality,
		EffectsQuality: DefaultEffectsQuality,
		ParticleCount:  100,
		UIScale:        1.0,

		// Audio
		MasterVolume:       1.0,
		MusicVolume:        DefaultMusicVolume,
		SFXVolume:          DefaultSFXVolume,
		UIVolume:           0.6,
		AmbientVolume:      0.5,
		MuteWhenUnfocused:  false,
		EnableDynamicMusic: true,

		// Controls
		MouseSensitivity: 1.0,
		InvertMouseY:     false,
		EdgeScrolling:    true,
		EdgeScrollSpeed:  1.0,
		KeyBindings:      sm.getDefaultKeyBindings(),
		GamepadEnabled:   true,
		GamepadVibration: true,

		// UI
		ShowNotifications:    DefaultNotifications,
		NotificationDuration: 5,
		ShowTutorialHints:    DefaultTutorialHints,
		ShowTooltips:         true,
		TooltipDelay:         500,
		ConfirmationDialogs:  DefaultConfirmationDialogs,
		ShowFPS:              DefaultShowFPS,
		ShowClock:            true,
		ShowMinimap:          true,
		MinimapSize:          200,

		// Advanced
		EnableDebugMode: false,
		ShowDebugInfo:   DefaultShowDebugInfo,
		EnableConsole:   false,
		LogLevel:        "info",
		MaxAutoSaves:    5,
		CompressSaves:   true,
		EncryptSaves:    true,
		NetworkTimeout:  30,
		CacheSize:       100,

		// Accessibility
		ColorblindMode:      "none",
		HighContrast:        false,
		ScreenReaderEnabled: false,
		SubtitlesEnabled:    false,
		SubtitleSize:        16,

		// Custom
		CustomSettings: make(map[string]interface{}),

		// Metadata
		Version:      "1.0.0",
		LastModified: time.Now(),
	}
}

// getDefaultKeyBindings returns default key bindings
func (sm *SettingsManager) getDefaultKeyBindings() map[string]string {
	return map[string]string{
		"pause":            "Escape",
		"quicksave":        "F5",
		"quickload":        "F9",
		"inventory":        "I",
		"market":           "M",
		"journal":          "J",
		"settings":         "O",
		"help":             "F1",
		"screenshot":       "F12",
		"speedup":          "Space",
		"zoomin":           "Plus",
		"zoomout":          "Minus",
		"togglefullscreen": "F11",
		"confirm":          "Enter",
		"cancel":           "Escape",
	}
}

// registerValidators registers setting validators
func (sm *SettingsManager) registerValidators() {
	// Volume validators (0.0 to 1.0)
	volumeValidator := func(value interface{}) error {
		v, ok := value.(float64)
		if !ok {
			return ErrInvalidType
		}
		if v < 0.0 || v > 1.0 {
			return ErrInvalidRange
		}
		return nil
	}

	sm.validators["master_volume"] = volumeValidator
	sm.validators["music_volume"] = volumeValidator
	sm.validators["sfx_volume"] = volumeValidator
	sm.validators["ui_volume"] = volumeValidator
	sm.validators["ambient_volume"] = volumeValidator

	// Game speed validator (0.1 to 10.0)
	sm.validators["game_speed"] = func(value interface{}) error {
		v, ok := value.(float64)
		if !ok {
			return ErrInvalidType
		}
		if v < 0.1 || v > 10.0 {
			return ErrInvalidRange
		}
		return nil
	}

	// FPS validator (30 to 240)
	sm.validators["target_fps"] = func(value interface{}) error {
		v, ok := value.(int)
		if !ok {
			return ErrInvalidType
		}
		if v < 30 || v > 240 {
			return ErrInvalidRange
		}
		return nil
	}

	// Difficulty validator
	sm.validators["difficulty"] = func(value interface{}) error {
		v, ok := value.(string)
		if !ok {
			return ErrInvalidType
		}
		validDifficulties := []string{"easy", "normal", "hard", "expert"}
		for _, d := range validDifficulties {
			if v == d {
				return nil
			}
		}
		return ErrInvalidSetting
	}

	// Quality validators
	qualityValidator := func(value interface{}) error {
		v, ok := value.(string)
		if !ok {
			return ErrInvalidType
		}
		validQualities := []string{"low", "medium", "high", "ultra"}
		for _, q := range validQualities {
			if v == q {
				return nil
			}
		}
		return ErrInvalidSetting
	}

	sm.validators["shadow_quality"] = qualityValidator
	sm.validators["texture_quality"] = qualityValidator
	sm.validators["effects_quality"] = qualityValidator
}

// LoadSettings loads settings from file
func (sm *SettingsManager) LoadSettings() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.locked {
		return ErrSettingsLocked
	}

	// Check if settings file exists
	if _, err := os.Stat(sm.settingsPath); os.IsNotExist(err) {
		// Use default settings
		return nil
	}

	// Read settings file
	data, err := os.ReadFile(sm.settingsPath)
	if err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	// Parse JSON
	var settings GameSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	sm.settings = &settings
	return nil
}

// SaveSettings saves settings to file
func (sm *SettingsManager) SaveSettings() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.locked {
		return ErrSettingsLocked
	}

	// Update last modified time
	sm.settings.LastModified = time.Now()

	// Convert to JSON
	data, err := json.MarshalIndent(sm.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize settings: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(sm.settingsPath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(sm.settingsPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// GetSetting gets a setting value
func (sm *SettingsManager) GetSetting(key string) (interface{}, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	switch key {
	// Game settings
	case SettingGameSpeed:
		return sm.settings.GameSpeed, nil
	case SettingDifficulty:
		return sm.settings.Difficulty, nil
	case SettingAutoSave:
		return sm.settings.AutoSave, nil
	case SettingAutoSaveInt:
		return sm.settings.AutoSaveInterval, nil
	case SettingLanguage:
		return sm.settings.Language, nil

	// Graphics settings
	case SettingFullscreen:
		return sm.settings.Fullscreen, nil
	case SettingVSync:
		return sm.settings.VSync, nil
	case SettingTargetFPS:
		return sm.settings.TargetFPS, nil
	case SettingShadowQuality:
		return sm.settings.ShadowQuality, nil
	case SettingTextureQuality:
		return sm.settings.TextureQuality, nil
	case SettingEffectsQuality:
		return sm.settings.EffectsQuality, nil

	// Audio settings
	case SettingMasterVolume:
		return sm.settings.MasterVolume, nil
	case SettingMusicVolume:
		return sm.settings.MusicVolume, nil
	case SettingSFXVolume:
		return sm.settings.SFXVolume, nil
	case SettingUIVolume:
		return sm.settings.UIVolume, nil
	case SettingAmbientVolume:
		return sm.settings.AmbientVolume, nil

	// UI settings
	case SettingShowFPS:
		return sm.settings.ShowFPS, nil
	case SettingShowNotifications:
		return sm.settings.ShowNotifications, nil
	case SettingShowTutorialHints:
		return sm.settings.ShowTutorialHints, nil

	default:
		// Check custom settings
		if val, ok := sm.settings.CustomSettings[key]; ok {
			return val, nil
		}
		return nil, ErrSettingNotFound
	}
}

// SetSetting sets a setting value
func (sm *SettingsManager) SetSetting(key string, value interface{}) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.locked {
		return ErrSettingsLocked
	}

	// Validate if validator exists
	if validator, ok := sm.validators[key]; ok {
		if err := validator(value); err != nil {
			return err
		}
	}

	// Store old value for callbacks
	oldValue, _ := sm.GetSettingUnlocked(key)

	// Set the value
	switch key {
	// Game settings
	case SettingGameSpeed:
		if v, ok := value.(float64); ok {
			sm.settings.GameSpeed = v
		} else {
			return ErrInvalidType
		}
	case SettingDifficulty:
		if v, ok := value.(string); ok {
			sm.settings.Difficulty = v
		} else {
			return ErrInvalidType
		}
	case SettingAutoSave:
		if v, ok := value.(bool); ok {
			sm.settings.AutoSave = v
		} else {
			return ErrInvalidType
		}
	case SettingAutoSaveInt:
		if v, ok := value.(int); ok {
			sm.settings.AutoSaveInterval = v
		} else {
			return ErrInvalidType
		}

	// Audio settings
	case SettingMusicVolume:
		if v, ok := value.(float64); ok {
			sm.settings.MusicVolume = v
		} else {
			return ErrInvalidType
		}
	case SettingSFXVolume:
		if v, ok := value.(float64); ok {
			sm.settings.SFXVolume = v
		} else {
			return ErrInvalidType
		}

	default:
		// Set custom setting
		sm.settings.CustomSettings[key] = value
	}

	// Trigger change callbacks
	if callbacks, ok := sm.changeCallbacks[key]; ok {
		for _, callback := range callbacks {
			callback(oldValue, value)
		}
	}

	// Auto-save if enabled
	if sm.autoSave {
		return sm.SaveSettings()
	}

	return nil
}

// GetSettingUnlocked gets a setting without locking (internal use)
func (sm *SettingsManager) GetSettingUnlocked(key string) (interface{}, error) {
	switch key {
	case SettingGameSpeed:
		return sm.settings.GameSpeed, nil
	case SettingMusicVolume:
		return sm.settings.MusicVolume, nil
	case SettingSFXVolume:
		return sm.settings.SFXVolume, nil
	default:
		if val, ok := sm.settings.CustomSettings[key]; ok {
			return val, nil
		}
		return nil, ErrSettingNotFound
	}
}

// ResetToDefaults resets all settings to defaults
func (sm *SettingsManager) ResetToDefaults() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.locked {
		return ErrSettingsLocked
	}

	sm.settings = sm.createDefaultSettings()

	if sm.autoSave {
		return sm.SaveSettings()
	}

	return nil
}

// ResetCategory resets a category of settings to defaults
func (sm *SettingsManager) ResetCategory(category SettingsCategory) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.locked {
		return ErrSettingsLocked
	}

	defaults := sm.createDefaultSettings()

	switch category {
	case CategoryGame:
		sm.settings.GameSpeed = defaults.GameSpeed
		sm.settings.Difficulty = defaults.Difficulty
		sm.settings.AutoSave = defaults.AutoSave
		sm.settings.AutoSaveInterval = defaults.AutoSaveInterval

	case CategoryGraphics:
		sm.settings.Resolution = defaults.Resolution
		sm.settings.Fullscreen = defaults.Fullscreen
		sm.settings.VSync = defaults.VSync
		sm.settings.ShadowQuality = defaults.ShadowQuality
		sm.settings.TextureQuality = defaults.TextureQuality
		sm.settings.EffectsQuality = defaults.EffectsQuality

	case CategoryAudio:
		sm.settings.MasterVolume = defaults.MasterVolume
		sm.settings.MusicVolume = defaults.MusicVolume
		sm.settings.SFXVolume = defaults.SFXVolume
		sm.settings.UIVolume = defaults.UIVolume
		sm.settings.AmbientVolume = defaults.AmbientVolume

	case CategoryControls:
		sm.settings.MouseSensitivity = defaults.MouseSensitivity
		sm.settings.InvertMouseY = defaults.InvertMouseY
		sm.settings.KeyBindings = defaults.KeyBindings

	case CategoryUI:
		sm.settings.ShowNotifications = defaults.ShowNotifications
		sm.settings.ShowTutorialHints = defaults.ShowTutorialHints
		sm.settings.ShowFPS = defaults.ShowFPS

	case CategoryAdvanced:
		sm.settings.EnableDebugMode = defaults.EnableDebugMode
		sm.settings.ShowDebugInfo = defaults.ShowDebugInfo
		sm.settings.MaxAutoSaves = defaults.MaxAutoSaves
	}

	if sm.autoSave {
		return sm.SaveSettings()
	}

	return nil
}

// RegisterChangeCallback registers a callback for setting changes
func (sm *SettingsManager) RegisterChangeCallback(key string, callback SettingChangeCallback) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.changeCallbacks[key] == nil {
		sm.changeCallbacks[key] = make([]SettingChangeCallback, 0)
	}
	sm.changeCallbacks[key] = append(sm.changeCallbacks[key], callback)
}

// GetSettings returns a copy of all settings
func (sm *SettingsManager) GetSettings() *GameSettings {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Create a copy to prevent external modification
	settingsCopy := *sm.settings
	return &settingsCopy
}

// ApplySettings applies a set of settings
func (sm *SettingsManager) ApplySettings(settings *GameSettings) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.locked {
		return ErrSettingsLocked
	}

	// Validate all settings first
	// TODO: Add validation for all fields

	sm.settings = settings

	if sm.autoSave {
		return sm.SaveSettings()
	}

	return nil
}

// Lock prevents settings from being changed
func (sm *SettingsManager) Lock() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.locked = true
}

// Unlock allows settings to be changed
func (sm *SettingsManager) Unlock() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.locked = false
}

// SetAutoSave enables or disables auto-save
func (sm *SettingsManager) SetAutoSave(enabled bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.autoSave = enabled
}

// ExportSettings exports settings to a JSON string
func (sm *SettingsManager) ExportSettings() (string, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	data, err := json.MarshalIndent(sm.settings, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ImportSettings imports settings from a JSON string
func (sm *SettingsManager) ImportSettings(jsonStr string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.locked {
		return ErrSettingsLocked
	}

	var settings GameSettings
	if err := json.Unmarshal([]byte(jsonStr), &settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	sm.settings = &settings

	if sm.autoSave {
		return sm.SaveSettings()
	}

	return nil
}
