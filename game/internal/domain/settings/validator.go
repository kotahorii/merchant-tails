package settings

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Validator provides validation for game settings
type Validator struct {
	rules map[string]ValidationRule
}

// ValidationRule defines a validation rule for a setting field
type ValidationRule struct {
	FieldName     string
	Required      bool
	MinValue      *float64
	MaxValue      *float64
	MinLength     *int
	MaxLength     *int
	Pattern       *regexp.Regexp
	CustomFunc    func(interface{}) error
	AllowedValues []interface{}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

// ValidationResult contains validation results
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// NewValidator creates a new settings validator
func NewValidator() *Validator {
	v := &Validator{
		rules: make(map[string]ValidationRule),
	}
	v.setupDefaultRules()
	return v
}

// setupDefaultRules sets up default validation rules
func (v *Validator) setupDefaultRules() {
	// Player settings
	v.AddRule("playerName", ValidationRule{
		FieldName: "playerName",
		Required:  true,
		MinLength: intPtr(1),
		MaxLength: intPtr(30),
		Pattern:   regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`),
	})

	// Game difficulty
	v.AddRule("difficulty", ValidationRule{
		FieldName:     "difficulty",
		Required:      true,
		AllowedValues: []interface{}{"easy", "normal", "hard", "expert"},
	})

	// Audio settings
	v.AddRule("masterVolume", ValidationRule{
		FieldName: "masterVolume",
		Required:  true,
		MinValue:  float64Ptr(0.0),
		MaxValue:  float64Ptr(1.0),
	})

	v.AddRule("sfxVolume", ValidationRule{
		FieldName: "sfxVolume",
		Required:  true,
		MinValue:  float64Ptr(0.0),
		MaxValue:  float64Ptr(1.0),
	})

	v.AddRule("musicVolume", ValidationRule{
		FieldName: "musicVolume",
		Required:  true,
		MinValue:  float64Ptr(0.0),
		MaxValue:  float64Ptr(1.0),
	})

	// Graphics settings
	v.AddRule("resolution", ValidationRule{
		FieldName: "resolution",
		Required:  true,
		Pattern:   regexp.MustCompile(`^\d{3,4}x\d{3,4}$`),
		AllowedValues: []interface{}{
			"1920x1080", "1680x1050", "1600x900", "1440x900",
			"1366x768", "1280x720", "1024x768", "800x600",
		},
	})

	v.AddRule("fullscreen", ValidationRule{
		FieldName: "fullscreen",
		Required:  true,
	})

	v.AddRule("vsync", ValidationRule{
		FieldName: "vsync",
		Required:  false,
	})

	v.AddRule("graphicsQuality", ValidationRule{
		FieldName:     "graphicsQuality",
		Required:      true,
		AllowedValues: []interface{}{"low", "medium", "high", "ultra"},
	})

	// Gameplay settings
	v.AddRule("autoSaveInterval", ValidationRule{
		FieldName: "autoSaveInterval",
		Required:  true,
		MinValue:  float64Ptr(0),    // 0 means disabled
		MaxValue:  float64Ptr(3600), // Max 1 hour
	})

	v.AddRule("showTutorial", ValidationRule{
		FieldName: "showTutorial",
		Required:  false,
	})

	v.AddRule("language", ValidationRule{
		FieldName:     "language",
		Required:      true,
		AllowedValues: []interface{}{"en", "ja", "es", "fr", "de", "zh", "ko"},
	})

	// Economy settings
	v.AddRule("startingGold", ValidationRule{
		FieldName: "startingGold",
		Required:  true,
		MinValue:  float64Ptr(100),
		MaxValue:  float64Ptr(10000),
	})

	v.AddRule("shopCapacity", ValidationRule{
		FieldName: "shopCapacity",
		Required:  true,
		MinValue:  float64Ptr(10),
		MaxValue:  float64Ptr(1000),
	})

	v.AddRule("warehouseCapacity", ValidationRule{
		FieldName: "warehouseCapacity",
		Required:  true,
		MinValue:  float64Ptr(50),
		MaxValue:  float64Ptr(5000),
	})

	// Market settings
	v.AddRule("priceFluctuation", ValidationRule{
		FieldName: "priceFluctuation",
		Required:  true,
		MinValue:  float64Ptr(0.1),
		MaxValue:  float64Ptr(2.0),
	})

	v.AddRule("demandSensitivity", ValidationRule{
		FieldName: "demandSensitivity",
		Required:  true,
		MinValue:  float64Ptr(0.1),
		MaxValue:  float64Ptr(3.0),
	})

	// Network settings
	v.AddRule("serverAddress", ValidationRule{
		FieldName: "serverAddress",
		Required:  false,
		Pattern:   regexp.MustCompile(`^(https?://)?([a-zA-Z0-9.-]+)(:\d+)?(/.*)?$`),
		MaxLength: intPtr(255),
	})

	v.AddRule("connectionTimeout", ValidationRule{
		FieldName: "connectionTimeout",
		Required:  false,
		MinValue:  float64Ptr(5),
		MaxValue:  float64Ptr(300),
	})

	// Notification settings
	v.AddRule("enableNotifications", ValidationRule{
		FieldName: "enableNotifications",
		Required:  false,
	})

	v.AddRule("notificationTypes", ValidationRule{
		FieldName: "notificationTypes",
		Required:  false,
		CustomFunc: func(value interface{}) error {
			types, ok := value.([]string)
			if !ok {
				return errors.New("must be array of strings")
			}
			allowedTypes := map[string]bool{
				"trade": true, "quest": true, "achievement": true,
				"market": true, "inventory": true, "system": true,
			}
			for _, t := range types {
				if !allowedTypes[t] {
					return fmt.Errorf("invalid notification type: %s", t)
				}
			}
			return nil
		},
	})
}

// AddRule adds a validation rule
func (v *Validator) AddRule(field string, rule ValidationRule) {
	v.rules[field] = rule
}

// RemoveRule removes a validation rule
func (v *Validator) RemoveRule(field string) {
	delete(v.rules, field)
}

// Validate validates a settings object
func (v *Validator) Validate(settings map[string]interface{}) *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	// Check each rule
	for field, rule := range v.rules {
		value, exists := settings[field]

		// Check required fields
		if rule.Required && !exists {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   field,
				Message: "field is required",
				Value:   nil,
			})
			continue
		}

		// Skip validation if field doesn't exist and is not required
		if !exists {
			continue
		}

		// Validate the field value
		if err := v.validateField(field, value, rule); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   field,
				Message: err.Error(),
				Value:   value,
			})
		}
	}

	return result
}

// validateField validates a single field
func (v *Validator) validateField(_ string, value interface{}, rule ValidationRule) error {
	// Check nil values
	if value == nil {
		if rule.Required {
			return errors.New("cannot be nil")
		}
		return nil
	}

	// String validation
	if strValue, ok := value.(string); ok {
		if rule.MinLength != nil && len(strValue) < *rule.MinLength {
			return fmt.Errorf("must be at least %d characters", *rule.MinLength)
		}
		if rule.MaxLength != nil && len(strValue) > *rule.MaxLength {
			return fmt.Errorf("must be at most %d characters", *rule.MaxLength)
		}
		if rule.Pattern != nil && !rule.Pattern.MatchString(strValue) {
			return fmt.Errorf("invalid format")
		}
	}

	// Numeric validation
	var numValue float64
	switch v := value.(type) {
	case int:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	case float32:
		numValue = float64(v)
	case float64:
		numValue = v
	default:
		// Not a numeric value, skip numeric validation
		goto checkAllowed
	}

	if rule.MinValue != nil && numValue < *rule.MinValue {
		return fmt.Errorf("must be at least %v", *rule.MinValue)
	}
	if rule.MaxValue != nil && numValue > *rule.MaxValue {
		return fmt.Errorf("must be at most %v", *rule.MaxValue)
	}

checkAllowed:
	// Check allowed values
	if len(rule.AllowedValues) > 0 {
		found := false
		for _, allowed := range rule.AllowedValues {
			if value == allowed {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("must be one of: %v", rule.AllowedValues)
		}
	}

	// Custom validation function
	if rule.CustomFunc != nil {
		if err := rule.CustomFunc(value); err != nil {
			return err
		}
	}

	return nil
}

// ValidatePartial validates only specified fields
func (v *Validator) ValidatePartial(settings map[string]interface{}, fields []string) *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	for _, field := range fields {
		rule, exists := v.rules[field]
		if !exists {
			continue
		}

		value, exists := settings[field]
		if !exists && rule.Required {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   field,
				Message: "field is required",
				Value:   nil,
			})
			continue
		}

		if exists {
			if err := v.validateField(field, value, rule); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   field,
					Message: err.Error(),
					Value:   value,
				})
			}
		}
	}

	return result
}

// SanitizeString sanitizes a string value
func SanitizeString(value string, maxLength int) string {
	// Trim whitespace
	value = strings.TrimSpace(value)

	// Limit length
	if len(value) > maxLength {
		value = value[:maxLength]
	}

	// Remove control characters
	value = regexp.MustCompile(`[\x00-\x1F\x7F]`).ReplaceAllString(value, "")

	return value
}

// SanitizeNumber ensures a number is within bounds
func SanitizeNumber(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

// Error implements the error interface for ValidationResult
func (vr *ValidationResult) Error() string {
	if vr.Valid {
		return ""
	}

	var messages []string
	for _, err := range vr.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}
