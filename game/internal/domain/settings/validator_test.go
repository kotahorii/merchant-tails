package settings

import (
	"fmt"
	"regexp"
	"testing"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	if v == nil {
		t.Fatal("NewValidator returned nil")
	}
	if v.rules == nil {
		t.Fatal("Validator rules map is nil")
	}
	if len(v.rules) == 0 {
		t.Error("Validator should have default rules")
	}
}

func TestValidatePlayerName(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name     string
		settings map[string]interface{}
		wantErr  bool
		errField string
	}{
		{
			name:     "Valid player name",
			settings: map[string]interface{}{"playerName": "TestPlayer"},
			wantErr:  false,
		},
		{
			name:     "Player name with spaces",
			settings: map[string]interface{}{"playerName": "Test Player"},
			wantErr:  false,
		},
		{
			name:     "Player name with numbers",
			settings: map[string]interface{}{"playerName": "Player123"},
			wantErr:  false,
		},
		{
			name:     "Empty player name",
			settings: map[string]interface{}{"playerName": ""},
			wantErr:  true,
			errField: "playerName",
		},
		{
			name:     "Player name too long",
			settings: map[string]interface{}{"playerName": "ThisPlayerNameIsWayTooLongForTheSystem"},
			wantErr:  true,
			errField: "playerName",
		},
		{
			name:     "Player name with special chars",
			settings: map[string]interface{}{"playerName": "Player@#$"},
			wantErr:  true,
			errField: "playerName",
		},
		{
			name:     "Missing player name",
			settings: map[string]interface{}{},
			wantErr:  true,
			errField: "playerName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use ValidatePartial to only validate playerName field
			result := v.ValidatePartial(tt.settings, []string{"playerName"})
			if tt.wantErr {
				if result.Valid {
					t.Error("Expected validation to fail but it passed")
				}
				if tt.errField != "" {
					found := false
					for _, err := range result.Errors {
						if err.Field == tt.errField {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error for field %s but not found", tt.errField)
					}
				}
			} else if !result.Valid {
				t.Errorf("Expected validation to pass but got errors: %v", result.Errors)
			}
		})
	}
}

func TestValidateAudioSettings(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name     string
		settings map[string]interface{}
		wantErr  bool
	}{
		{
			name: "Valid audio settings",
			settings: map[string]interface{}{
				"masterVolume": 0.5,
				"sfxVolume":    0.7,
				"musicVolume":  0.3,
			},
			wantErr: false,
		},
		{
			name: "Volume at minimum",
			settings: map[string]interface{}{
				"masterVolume": 0.0,
				"sfxVolume":    0.0,
				"musicVolume":  0.0,
			},
			wantErr: false,
		},
		{
			name: "Volume at maximum",
			settings: map[string]interface{}{
				"masterVolume": 1.0,
				"sfxVolume":    1.0,
				"musicVolume":  1.0,
			},
			wantErr: false,
		},
		{
			name: "Volume too high",
			settings: map[string]interface{}{
				"masterVolume": 1.5,
			},
			wantErr: true,
		},
		{
			name: "Negative volume",
			settings: map[string]interface{}{
				"sfxVolume": -0.1,
			},
			wantErr: true,
		},
		{
			name: "Invalid volume type",
			settings: map[string]interface{}{
				"musicVolume": "loud",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.ValidatePartial(tt.settings, []string{"masterVolume", "sfxVolume", "musicVolume"})
			if tt.wantErr {
				if result.Valid {
					t.Error("Expected validation to fail but it passed")
				}
			} else if !result.Valid {
				t.Errorf("Expected validation to pass but got errors: %v", result.Errors)
			}
		})
	}
}

func TestValidateGraphicsSettings(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name     string
		settings map[string]interface{}
		wantErr  bool
	}{
		{
			name: "Valid resolution",
			settings: map[string]interface{}{
				"resolution": "1920x1080",
			},
			wantErr: false,
		},
		{
			name: "Another valid resolution",
			settings: map[string]interface{}{
				"resolution": "1366x768",
			},
			wantErr: false,
		},
		{
			name: "Invalid resolution format",
			settings: map[string]interface{}{
				"resolution": "1920*1080",
			},
			wantErr: true,
		},
		{
			name: "Invalid resolution value",
			settings: map[string]interface{}{
				"resolution": "9999x9999",
			},
			wantErr: true,
		},
		{
			name: "Valid graphics quality",
			settings: map[string]interface{}{
				"graphicsQuality": "high",
			},
			wantErr: false,
		},
		{
			name: "Invalid graphics quality",
			settings: map[string]interface{}{
				"graphicsQuality": "super",
			},
			wantErr: true,
		},
		{
			name: "Valid fullscreen setting",
			settings: map[string]interface{}{
				"fullscreen": true,
			},
			wantErr: false,
		},
		{
			name: "Valid vsync setting",
			settings: map[string]interface{}{
				"vsync": false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := []string{}
			for k := range tt.settings {
				fields = append(fields, k)
			}
			result := v.ValidatePartial(tt.settings, fields)
			if tt.wantErr {
				if result.Valid {
					t.Error("Expected validation to fail but it passed")
				}
			} else if !result.Valid {
				t.Errorf("Expected validation to pass but got errors: %v", result.Errors)
			}
		})
	}
}

func TestValidateGameplaySettings(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name     string
		settings map[string]interface{}
		wantErr  bool
	}{
		{
			name: "Valid difficulty",
			settings: map[string]interface{}{
				"difficulty": "normal",
			},
			wantErr: false,
		},
		{
			name: "All valid difficulties",
			settings: map[string]interface{}{
				"difficulty": "easy",
			},
			wantErr: false,
		},
		{
			name: "Invalid difficulty",
			settings: map[string]interface{}{
				"difficulty": "impossible",
			},
			wantErr: true,
		},
		{
			name: "Valid auto-save interval",
			settings: map[string]interface{}{
				"autoSaveInterval": 300.0,
			},
			wantErr: false,
		},
		{
			name: "Auto-save disabled",
			settings: map[string]interface{}{
				"autoSaveInterval": 0.0,
			},
			wantErr: false,
		},
		{
			name: "Auto-save interval too high",
			settings: map[string]interface{}{
				"autoSaveInterval": 7200.0,
			},
			wantErr: true,
		},
		{
			name: "Valid language",
			settings: map[string]interface{}{
				"language": "en",
			},
			wantErr: false,
		},
		{
			name: "Invalid language",
			settings: map[string]interface{}{
				"language": "klingon",
			},
			wantErr: true,
		},
		{
			name: "Show tutorial setting",
			settings: map[string]interface{}{
				"showTutorial": true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := []string{}
			for k := range tt.settings {
				fields = append(fields, k)
			}
			result := v.ValidatePartial(tt.settings, fields)
			if tt.wantErr {
				if result.Valid {
					t.Error("Expected validation to fail but it passed")
				}
			} else if !result.Valid {
				t.Errorf("Expected validation to pass but got errors: %v", result.Errors)
			}
		})
	}
}

func TestValidateEconomySettings(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name     string
		settings map[string]interface{}
		wantErr  bool
	}{
		{
			name: "Valid starting gold",
			settings: map[string]interface{}{
				"startingGold": 500.0,
			},
			wantErr: false,
		},
		{
			name: "Starting gold at minimum",
			settings: map[string]interface{}{
				"startingGold": 100.0,
			},
			wantErr: false,
		},
		{
			name: "Starting gold at maximum",
			settings: map[string]interface{}{
				"startingGold": 10000.0,
			},
			wantErr: false,
		},
		{
			name: "Starting gold too low",
			settings: map[string]interface{}{
				"startingGold": 50.0,
			},
			wantErr: true,
		},
		{
			name: "Starting gold too high",
			settings: map[string]interface{}{
				"startingGold": 20000.0,
			},
			wantErr: true,
		},
		{
			name: "Valid shop capacity",
			settings: map[string]interface{}{
				"shopCapacity": 100.0,
			},
			wantErr: false,
		},
		{
			name: "Shop capacity too low",
			settings: map[string]interface{}{
				"shopCapacity": 5.0,
			},
			wantErr: true,
		},
		{
			name: "Valid warehouse capacity",
			settings: map[string]interface{}{
				"warehouseCapacity": 500.0,
			},
			wantErr: false,
		},
		{
			name: "Warehouse capacity too high",
			settings: map[string]interface{}{
				"warehouseCapacity": 10000.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := []string{}
			for k := range tt.settings {
				fields = append(fields, k)
			}
			result := v.ValidatePartial(tt.settings, fields)
			if tt.wantErr {
				if result.Valid {
					t.Error("Expected validation to fail but it passed")
				}
			} else if !result.Valid {
				t.Errorf("Expected validation to pass but got errors: %v", result.Errors)
			}
		})
	}
}

func TestValidateMarketSettings(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name     string
		settings map[string]interface{}
		wantErr  bool
	}{
		{
			name: "Valid price fluctuation",
			settings: map[string]interface{}{
				"priceFluctuation": 1.0,
			},
			wantErr: false,
		},
		{
			name: "Price fluctuation at minimum",
			settings: map[string]interface{}{
				"priceFluctuation": 0.1,
			},
			wantErr: false,
		},
		{
			name: "Price fluctuation too low",
			settings: map[string]interface{}{
				"priceFluctuation": 0.05,
			},
			wantErr: true,
		},
		{
			name: "Valid demand sensitivity",
			settings: map[string]interface{}{
				"demandSensitivity": 1.5,
			},
			wantErr: false,
		},
		{
			name: "Demand sensitivity too high",
			settings: map[string]interface{}{
				"demandSensitivity": 5.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := []string{}
			for k := range tt.settings {
				fields = append(fields, k)
			}
			result := v.ValidatePartial(tt.settings, fields)
			if tt.wantErr {
				if result.Valid {
					t.Error("Expected validation to fail but it passed")
				}
			} else if !result.Valid {
				t.Errorf("Expected validation to pass but got errors: %v", result.Errors)
			}
		})
	}
}

func TestValidateNetworkSettings(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name     string
		settings map[string]interface{}
		wantErr  bool
	}{
		{
			name: "Valid server address with http",
			settings: map[string]interface{}{
				"serverAddress": "http://localhost:8080",
			},
			wantErr: false,
		},
		{
			name: "Valid server address with https",
			settings: map[string]interface{}{
				"serverAddress": "https://api.example.com",
			},
			wantErr: false,
		},
		{
			name: "Valid server address with path",
			settings: map[string]interface{}{
				"serverAddress": "https://api.example.com/game",
			},
			wantErr: false,
		},
		{
			name: "Invalid server address",
			settings: map[string]interface{}{
				"serverAddress": "not a url",
			},
			wantErr: true,
		},
		{
			name: "Server address too long",
			settings: map[string]interface{}{
				"serverAddress": "https://" + string(make([]byte, 300)),
			},
			wantErr: true,
		},
		{
			name: "Valid connection timeout",
			settings: map[string]interface{}{
				"connectionTimeout": 30.0,
			},
			wantErr: false,
		},
		{
			name: "Connection timeout too low",
			settings: map[string]interface{}{
				"connectionTimeout": 2.0,
			},
			wantErr: true,
		},
		{
			name: "Connection timeout too high",
			settings: map[string]interface{}{
				"connectionTimeout": 600.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := []string{}
			for k := range tt.settings {
				fields = append(fields, k)
			}
			result := v.ValidatePartial(tt.settings, fields)
			if tt.wantErr {
				if result.Valid {
					t.Error("Expected validation to fail but it passed")
				}
			} else if !result.Valid {
				t.Errorf("Expected validation to pass but got errors: %v", result.Errors)
			}
		})
	}
}

func TestValidateNotificationSettings(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name     string
		settings map[string]interface{}
		wantErr  bool
	}{
		{
			name: "Valid notification types",
			settings: map[string]interface{}{
				"notificationTypes": []string{"trade", "quest", "achievement"},
			},
			wantErr: false,
		},
		{
			name: "All valid notification types",
			settings: map[string]interface{}{
				"notificationTypes": []string{"trade", "quest", "achievement", "market", "inventory", "system"},
			},
			wantErr: false,
		},
		{
			name: "Invalid notification type",
			settings: map[string]interface{}{
				"notificationTypes": []string{"trade", "invalid"},
			},
			wantErr: true,
		},
		{
			name: "Wrong type for notification types",
			settings: map[string]interface{}{
				"notificationTypes": "trade",
			},
			wantErr: true,
		},
		{
			name: "Enable notifications setting",
			settings: map[string]interface{}{
				"enableNotifications": true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := []string{}
			for k := range tt.settings {
				fields = append(fields, k)
			}
			result := v.ValidatePartial(tt.settings, fields)
			if tt.wantErr {
				if result.Valid {
					t.Error("Expected validation to fail but it passed")
				}
			} else if !result.Valid {
				t.Errorf("Expected validation to pass but got errors: %v", result.Errors)
			}
		})
	}
}

func TestAddAndRemoveRule(t *testing.T) {
	v := NewValidator()

	// Add a custom rule
	customRule := ValidationRule{
		FieldName: "customField",
		Required:  true,
		MinValue:  float64Ptr(10),
		MaxValue:  float64Ptr(100),
	}
	v.AddRule("customField", customRule)

	// Test with valid value
	settings := map[string]interface{}{
		"customField": 50.0,
	}
	result := v.ValidatePartial(settings, []string{"customField"})
	if !result.Valid {
		t.Errorf("Custom rule validation failed: %v", result.Errors)
	}

	// Test with invalid value
	settings["customField"] = 150.0
	result = v.ValidatePartial(settings, []string{"customField"})
	if result.Valid {
		t.Error("Expected custom rule validation to fail for out of range value")
	}

	// Remove the rule
	v.RemoveRule("customField")

	// Should not validate anymore
	result = v.ValidatePartial(settings, []string{"customField"})
	if !result.Valid {
		t.Error("Rule should have been removed")
	}
}

func TestCustomValidationFunction(t *testing.T) {
	v := NewValidator()

	// Add rule with custom validation function
	customRule := ValidationRule{
		FieldName: "email",
		Required:  true,
		CustomFunc: func(value interface{}) error {
			str, ok := value.(string)
			if !ok {
				return nil
			}
			emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
			if !emailRegex.MatchString(str) {
				return errorf("invalid email format")
			}
			return nil
		},
	}
	v.AddRule("email", customRule)

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"Valid email", "test@example.com", false},
		{"Another valid email", "user.name+tag@domain.co.uk", false},
		{"Invalid email - no @", "invalid.email", true},
		{"Invalid email - no domain", "test@", true},
		{"Invalid email - no local part", "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := map[string]interface{}{"email": tt.email}
			result := v.ValidatePartial(settings, []string{"email"})
			if tt.wantErr {
				if result.Valid {
					t.Error("Expected validation to fail but it passed")
				}
			} else if !result.Valid {
				t.Errorf("Expected validation to pass but got errors: %v", result.Errors)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "Normal string",
			input:     "Hello World",
			maxLength: 20,
			expected:  "Hello World",
		},
		{
			name:      "String with whitespace",
			input:     "  Hello World  ",
			maxLength: 20,
			expected:  "Hello World",
		},
		{
			name:      "String too long",
			input:     "This is a very long string that exceeds the limit",
			maxLength: 10,
			expected:  "This is a ",
		},
		{
			name:      "String with control characters",
			input:     "Hello\x00World\x1F",
			maxLength: 20,
			expected:  "HelloWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input, tt.maxLength)
			if result != tt.expected {
				t.Errorf("SanitizeString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizeNumber(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		min      float64
		max      float64
		expected float64
	}{
		{
			name:     "Value within range",
			value:    50,
			min:      0,
			max:      100,
			expected: 50,
		},
		{
			name:     "Value below minimum",
			value:    -10,
			min:      0,
			max:      100,
			expected: 0,
		},
		{
			name:     "Value above maximum",
			value:    150,
			min:      0,
			max:      100,
			expected: 100,
		},
		{
			name:     "Value at minimum",
			value:    0,
			min:      0,
			max:      100,
			expected: 0,
		},
		{
			name:     "Value at maximum",
			value:    100,
			min:      0,
			max:      100,
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeNumber(tt.value, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("SanitizeNumber() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidationResultError(t *testing.T) {
	// Test valid result
	validResult := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}
	if validResult.Error() != "" {
		t.Error("Valid result should return empty error string")
	}

	// Test invalid result with errors
	invalidResult := &ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{Field: "field1", Message: "is required"},
			{Field: "field2", Message: "must be positive"},
		},
	}
	errStr := invalidResult.Error()
	if errStr == "" {
		t.Error("Invalid result should return non-empty error string")
	}
	if errStr != "field1: is required; field2: must be positive" {
		t.Errorf("Unexpected error string: %s", errStr)
	}
}

// Helper function for tests
func errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
