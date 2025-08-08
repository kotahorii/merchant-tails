package weather

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewWeatherManager(t *testing.T) {
	wm := NewWeatherManager()

	assert.NotNil(t, wm)
	assert.NotNil(t, wm.currentWeather)
	assert.NotNil(t, wm.patterns)
	assert.NotNil(t, wm.forecast)
	assert.NotNil(t, wm.history)

	// Check initial weather
	assert.Equal(t, WeatherSunny, wm.currentWeather.Type)
	assert.Equal(t, SeverityMild, wm.currentWeather.Severity)
	assert.NotNil(t, wm.currentWeather.Effects)

	// Check patterns are initialized
	assert.Equal(t, 4, len(wm.patterns))
	assert.NotNil(t, wm.patterns["spring"])
	assert.NotNil(t, wm.patterns["summer"])
	assert.NotNil(t, wm.patterns["autumn"])
	assert.NotNil(t, wm.patterns["winter"])
}

func TestWeatherTypeString(t *testing.T) {
	tests := []struct {
		weatherType WeatherType
		expected    string
	}{
		{WeatherSunny, "Sunny"},
		{WeatherCloudy, "Cloudy"},
		{WeatherRainy, "Rainy"},
		{WeatherStormy, "Stormy"},
		{WeatherSnowy, "Snowy"},
		{WeatherFoggy, "Foggy"},
		{WeatherWindy, "Windy"},
		{WeatherHot, "Hot"},
		{WeatherCold, "Cold"},
		{WeatherType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.weatherType.String())
		})
	}
}

func TestCalculateEffects(t *testing.T) {
	wm := NewWeatherManager()

	tests := []struct {
		name         string
		weatherType  WeatherType
		severity     WeatherSeverity
		checkEffects func(*WeatherEffects)
	}{
		{
			name:        "sunny mild",
			weatherType: WeatherSunny,
			severity:    SeverityMild,
			checkEffects: func(e *WeatherEffects) {
				assert.Less(t, e.FoodDemand, 1.0)
				assert.Greater(t, e.LuxuryDemand, 1.0)
				assert.Greater(t, e.CustomerTraffic, 1.0)
			},
		},
		{
			name:        "rainy moderate",
			weatherType: WeatherRainy,
			severity:    SeverityModerate,
			checkEffects: func(e *WeatherEffects) {
				assert.Greater(t, e.FoodDemand, 1.0)
				assert.Greater(t, e.ClothingDemand, 1.0)
				assert.Less(t, e.CustomerTraffic, 1.0)
			},
		},
		{
			name:        "stormy severe",
			weatherType: WeatherStormy,
			severity:    SeveritySevere,
			checkEffects: func(e *WeatherEffects) {
				assert.Greater(t, e.TransportCost, 1.3)
				assert.Less(t, e.CustomerTraffic, 0.5)
				assert.Less(t, e.ItemDurability, 0.8)
			},
		},
		{
			name:        "snowy extreme",
			weatherType: WeatherSnowy,
			severity:    SeverityExtreme,
			checkEffects: func(e *WeatherEffects) {
				assert.Greater(t, e.FoodDemand, 1.5)
				assert.Greater(t, e.ClothingDemand, 1.8)
				assert.Less(t, e.CustomerTraffic, 0.5)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effects := wm.calculateEffects(tt.weatherType, tt.severity)
			assert.NotNil(t, effects)
			tt.checkEffects(effects)
		})
	}
}

func TestUpdateWeather(t *testing.T) {
	wm := NewWeatherManager()

	// Register callback to track changes
	var oldW, newW *Weather
	wm.RegisterCallback(func(old, new *Weather) {
		oldW = old
		newW = new
	})

	// Force short duration weather
	wm.currentWeather.Duration = 0
	wm.currentWeather.StartTime = time.Now().Add(-1 * time.Hour)

	// Update weather
	wm.UpdateWeather("summer", 14)

	// Check callback was called
	assert.NotNil(t, oldW)
	assert.NotNil(t, newW)
	assert.Equal(t, WeatherSunny, oldW.Type)

	// Check new weather is from summer pattern
	summerTypes := wm.patterns["summer"].CommonTypes
	found := false
	for _, wType := range summerTypes {
		if newW.Type == wType {
			found = true
			break
		}
	}
	assert.True(t, found, "New weather should be from summer pattern")

	// Check history
	history := wm.GetWeatherHistory(10)
	assert.Greater(t, len(history), 0)
	assert.Equal(t, oldW, history[len(history)-1])

	// Check forecast was generated
	forecast := wm.GetForecast()
	assert.Greater(t, len(forecast), 0)
}

func TestGenerateWeather(t *testing.T) {
	wm := NewWeatherManager()

	pattern := wm.patterns["winter"]
	weather := wm.generateWeather(pattern, 3)

	assert.NotNil(t, weather)
	assert.NotNil(t, weather.Effects)
	assert.Greater(t, weather.Duration, 0)
	assert.NotZero(t, weather.StartTime)

	// Check temperature is in winter range
	assert.GreaterOrEqual(t, weather.Temperature, pattern.Temperature.Min-5)
	assert.LessOrEqual(t, weather.Temperature, pattern.Temperature.Max+5)

	// Check humidity is in range
	assert.GreaterOrEqual(t, weather.Humidity, pattern.Humidity.Min)
	assert.LessOrEqual(t, weather.Humidity, pattern.Humidity.Max)
}

func TestGenerateForecast(t *testing.T) {
	wm := NewWeatherManager()

	pattern := wm.patterns["spring"]
	wm.generateForecast(pattern)

	forecast := wm.GetForecast()
	assert.Equal(t, 3, len(forecast))

	// Check forecast properties
	for i, f := range forecast {
		assert.NotNil(t, f.Weather)
		assert.NotNil(t, f.Weather.Effects)
		assert.Greater(t, f.Probability, 0.0)
		assert.LessOrEqual(t, f.Probability, 1.0)
		assert.Greater(t, f.TimeUntil, time.Duration(0))

		// Probability should generally decrease
		if i > 0 {
			assert.LessOrEqual(t, f.Probability, forecast[i-1].Probability+0.2)
		}
	}
}

func TestGetMarketPriceModifier(t *testing.T) {
	wm := NewWeatherManager()

	// Set specific weather
	wm.ForceWeather(WeatherSnowy, SeverityModerate, 8)

	tests := []struct {
		category string
		modifier float64
	}{
		{"FRUIT", wm.currentWeather.Effects.FoodDemand},
		{"FOOD", wm.currentWeather.Effects.FoodDemand},
		{"CLOTHING", wm.currentWeather.Effects.ClothingDemand},
		{"ACCESSORY", wm.currentWeather.Effects.ClothingDemand},
		{"TOOL", wm.currentWeather.Effects.ToolDemand},
		{"WEAPON", wm.currentWeather.Effects.ToolDemand},
		{"LUXURY", wm.currentWeather.Effects.LuxuryDemand},
		{"GEM", wm.currentWeather.Effects.LuxuryDemand},
		{"UNKNOWN", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			modifier := wm.GetMarketPriceModifier(tt.category)
			assert.Equal(t, tt.modifier, modifier)
		})
	}
}

func TestGetTransportCostModifier(t *testing.T) {
	wm := NewWeatherManager()

	// Normal weather
	assert.Equal(t, 1.0, wm.GetTransportCostModifier())

	// Stormy weather
	wm.ForceWeather(WeatherStormy, SeverityModerate, 4)
	assert.Greater(t, wm.GetTransportCostModifier(), 1.0)
}

func TestGetCustomerTrafficModifier(t *testing.T) {
	wm := NewWeatherManager()

	// Sunny weather
	wm.ForceWeather(WeatherSunny, SeverityModerate, 8)
	assert.Greater(t, wm.GetCustomerTrafficModifier(), 1.0)

	// Stormy weather
	wm.ForceWeather(WeatherStormy, SeverityModerate, 4)
	assert.Less(t, wm.GetCustomerTrafficModifier(), 1.0)
}

func TestForceWeather(t *testing.T) {
	wm := NewWeatherManager()

	// Register callback
	callbackCalled := false
	wm.RegisterCallback(func(old, new *Weather) {
		callbackCalled = true
		assert.Equal(t, WeatherSunny, old.Type)
		assert.Equal(t, WeatherStormy, new.Type)
	})

	// Force weather
	wm.ForceWeather(WeatherStormy, SeverityExtreme, 12)

	assert.True(t, callbackCalled)
	assert.Equal(t, WeatherStormy, wm.currentWeather.Type)
	assert.Equal(t, SeverityExtreme, wm.currentWeather.Severity)
	assert.Equal(t, 12, wm.currentWeather.Duration)
}

func TestIsExtremeWeather(t *testing.T) {
	wm := NewWeatherManager()

	// Normal weather
	assert.False(t, wm.IsExtremeWeather())

	// Extreme severity
	wm.ForceWeather(WeatherSunny, SeverityExtreme, 4)
	assert.True(t, wm.IsExtremeWeather())

	// Stormy weather
	wm.ForceWeather(WeatherStormy, SeverityMild, 4)
	assert.True(t, wm.IsExtremeWeather())
}

func TestGetWeatherWarnings(t *testing.T) {
	wm := NewWeatherManager()

	// Normal weather - no warnings
	warnings := wm.GetWeatherWarnings()
	assert.Equal(t, 0, len(warnings))

	// Extreme weather
	wm.ForceWeather(WeatherStormy, SeverityExtreme, 4)
	warnings = wm.GetWeatherWarnings()
	assert.Greater(t, len(warnings), 0)
	assert.Contains(t, warnings[0], "EXTREME")
	assert.Contains(t, warnings[1], "STORM")

	// Heat warning
	wm.currentWeather.Temperature = 40
	warnings = wm.GetWeatherWarnings()
	assert.Greater(t, len(warnings), 0)
	found := false
	for _, w := range warnings {
		if contains(w, "HEAT") {
			found = true
			break
		}
	}
	assert.True(t, found)

	// Cold warning
	wm.currentWeather.Temperature = -10
	warnings = wm.GetWeatherWarnings()
	found = false
	for _, w := range warnings {
		if contains(w, "COLD") {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestWeatherHistory(t *testing.T) {
	wm := NewWeatherManager()

	// Initially empty
	history := wm.GetWeatherHistory(10)
	assert.Equal(t, 0, len(history))

	// Force multiple weather changes
	for i := 0; i < 5; i++ {
		wm.currentWeather.Duration = 0
		wm.currentWeather.StartTime = time.Now().Add(-1 * time.Hour)
		wm.UpdateWeather("spring", 12)
	}

	// Check history
	history = wm.GetWeatherHistory(10)
	assert.Equal(t, 5, len(history))

	// Check limited history
	history = wm.GetWeatherHistory(3)
	assert.Equal(t, 3, len(history))
}

func TestSeasonalPatterns(t *testing.T) {
	wm := NewWeatherManager()

	seasons := []string{"spring", "summer", "autumn", "winter"}

	for _, season := range seasons {
		t.Run(season, func(t *testing.T) {
			pattern := wm.patterns[season]
			assert.NotNil(t, pattern)
			assert.Equal(t, season, pattern.Season)
			assert.Greater(t, len(pattern.CommonTypes), 0)
			assert.Less(t, pattern.Temperature.Min, pattern.Temperature.Max)
			assert.Less(t, pattern.Humidity.Min, pattern.Humidity.Max)
			assert.Less(t, pattern.WindSpeed.Min, pattern.WindSpeed.Max)
		})
	}
}

func TestSeverityMultiplier(t *testing.T) {
	wm := NewWeatherManager()

	// Test that severity affects effects magnitude
	mildEffects := wm.calculateEffects(WeatherRainy, SeverityMild)
	moderateEffects := wm.calculateEffects(WeatherRainy, SeverityModerate)
	severeEffects := wm.calculateEffects(WeatherRainy, SeveritySevere)
	extremeEffects := wm.calculateEffects(WeatherRainy, SeverityExtreme)

	// Food demand should increase with severity
	assert.Less(t, mildEffects.FoodDemand, moderateEffects.FoodDemand)
	assert.Less(t, moderateEffects.FoodDemand, severeEffects.FoodDemand)
	assert.Less(t, severeEffects.FoodDemand, extremeEffects.FoodDemand)

	// Customer traffic should decrease more with severity
	assert.Greater(t, mildEffects.CustomerTraffic, moderateEffects.CustomerTraffic)
	assert.Greater(t, moderateEffects.CustomerTraffic, severeEffects.CustomerTraffic)
	assert.Greater(t, severeEffects.CustomerTraffic, extremeEffects.CustomerTraffic)
}

func TestReset(t *testing.T) {
	wm := NewWeatherManager()

	// Change weather and add history
	wm.ForceWeather(WeatherStormy, SeverityExtreme, 4)
	wm.currentWeather.Duration = 0
	wm.currentWeather.StartTime = time.Now().Add(-1 * time.Hour)
	wm.UpdateWeather("winter", 12)

	// Generate forecast
	wm.generateForecast(wm.patterns["winter"])

	// Verify state before reset
	assert.NotEqual(t, WeatherSunny, wm.currentWeather.Type)
	assert.Greater(t, len(wm.history), 0)
	assert.Greater(t, len(wm.forecast), 0)

	// Reset
	wm.Reset()

	// Verify reset state
	assert.Equal(t, WeatherSunny, wm.currentWeather.Type)
	assert.Equal(t, SeverityMild, wm.currentWeather.Severity)
	assert.Equal(t, 0, len(wm.history))
	assert.Equal(t, 0, len(wm.forecast))
}

func TestConcurrentAccess(t *testing.T) {
	wm := NewWeatherManager()

	done := make(chan bool, 4)

	// Concurrent weather updates
	go func() {
		for i := 0; i < 100; i++ {
			wm.UpdateWeather("spring", i%24)
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			_ = wm.GetCurrentWeather()
			_ = wm.GetForecast()
			_ = wm.GetWeatherHistory(10)
		}
		done <- true
	}()

	// Concurrent modifier checks
	go func() {
		for i := 0; i < 100; i++ {
			_ = wm.GetMarketPriceModifier("FOOD")
			_ = wm.GetTransportCostModifier()
			_ = wm.GetCustomerTrafficModifier()
		}
		done <- true
	}()

	// Concurrent force weather
	go func() {
		for i := 0; i < 100; i++ {
			wType := WeatherType(i % 9)
			severity := WeatherSeverity(i % 4)
			wm.ForceWeather(wType, severity, 4)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}

	// Verify state is consistent
	weather := wm.GetCurrentWeather()
	assert.NotNil(t, weather)
	assert.NotNil(t, weather.Effects)
}

func TestTimeOfDayTemperature(t *testing.T) {
	wm := NewWeatherManager()
	pattern := wm.patterns["summer"]

	// Due to randomness, we can't guarantee exact values,
	// but we can test the general trend over multiple generations
	morningTemps := 0
	afternoonTemps := 0

	for i := 0; i < 10; i++ {
		morning := wm.generateWeather(pattern, 3)
		afternoon := wm.generateWeather(pattern, 14)
		morningTemps += morning.Temperature
		afternoonTemps += afternoon.Temperature
	}

	// On average, afternoon should be warmer
	assert.Less(t, morningTemps, afternoonTemps)
}

func TestWeatherEffectsConsistency(t *testing.T) {
	wm := NewWeatherManager()

	// Test that all weather types have proper effects
	weatherTypes := []WeatherType{
		WeatherSunny, WeatherCloudy, WeatherRainy, WeatherStormy,
		WeatherSnowy, WeatherFoggy, WeatherWindy, WeatherHot, WeatherCold,
	}

	for _, wType := range weatherTypes {
		t.Run(wType.String(), func(t *testing.T) {
			effects := wm.calculateEffects(wType, SeverityModerate)

			// All modifiers should be positive
			assert.Greater(t, effects.FoodDemand, 0.0)
			assert.Greater(t, effects.ClothingDemand, 0.0)
			assert.Greater(t, effects.ToolDemand, 0.0)
			assert.Greater(t, effects.LuxuryDemand, 0.0)
			assert.Greater(t, effects.TransportCost, 0.0)
			assert.Greater(t, effects.StorageCost, 0.0)
			assert.Greater(t, effects.CustomerTraffic, 0.0)
			assert.Greater(t, effects.ItemDurability, 0.0)
			assert.Greater(t, effects.TravelSpeed, 0.0)
			assert.Greater(t, effects.EventProbability, 0.0)
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && contains(s[1:], substr)
}
