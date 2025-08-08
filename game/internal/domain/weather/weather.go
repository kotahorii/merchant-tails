package weather

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// WeatherType represents different weather conditions
type WeatherType int

const (
	WeatherSunny WeatherType = iota
	WeatherCloudy
	WeatherRainy
	WeatherStormy
	WeatherSnowy
	WeatherFoggy
	WeatherWindy
	WeatherHot
	WeatherCold
)

// String returns the string representation of weather type
func (w WeatherType) String() string {
	switch w {
	case WeatherSunny:
		return "Sunny"
	case WeatherCloudy:
		return "Cloudy"
	case WeatherRainy:
		return "Rainy"
	case WeatherStormy:
		return "Stormy"
	case WeatherSnowy:
		return "Snowy"
	case WeatherFoggy:
		return "Foggy"
	case WeatherWindy:
		return "Windy"
	case WeatherHot:
		return "Hot"
	case WeatherCold:
		return "Cold"
	default:
		return "Unknown"
	}
}

// WeatherSeverity represents the intensity of weather
type WeatherSeverity int

const (
	SeverityMild WeatherSeverity = iota
	SeverityModerate
	SeveritySevere
	SeverityExtreme
)

// Weather represents current weather conditions
type Weather struct {
	Type        WeatherType
	Severity    WeatherSeverity
	Temperature int // In Celsius
	Humidity    int // Percentage
	WindSpeed   int // km/h
	Duration    int // Hours
	StartTime   time.Time
	Effects     *WeatherEffects
}

// WeatherEffects represents how weather affects the game
type WeatherEffects struct {
	// Market price modifiers (percentage)
	FoodDemand     float64
	ClothingDemand float64
	ToolDemand     float64
	LuxuryDemand   float64
	TransportCost  float64
	StorageCost    float64

	// Gameplay modifiers
	CustomerTraffic  float64
	ItemDurability   float64
	TravelSpeed      float64
	EventProbability float64
}

// WeatherForecast represents predicted weather
type WeatherForecast struct {
	Weather     *Weather
	Probability float64 // 0.0 to 1.0
	TimeUntil   time.Duration
}

// WeatherPattern represents seasonal weather patterns
type WeatherPattern struct {
	Season      string
	CommonTypes []WeatherType
	Temperature Range
	Humidity    Range
	WindSpeed   Range
}

// Range represents min and max values
type Range struct {
	Min int
	Max int
}

// WeatherManager manages weather system
type WeatherManager struct {
	currentWeather *Weather
	forecast       []*WeatherForecast
	patterns       map[string]*WeatherPattern
	history        []*Weather
	callbacks      []WeatherCallback
	lastUpdate     time.Time
	rng            *rand.Rand
	mu             sync.RWMutex
}

// WeatherCallback is called when weather changes
type WeatherCallback func(oldWeather, newWeather *Weather)

// NewWeatherManager creates a new weather manager
func NewWeatherManager() *WeatherManager {
	wm := &WeatherManager{
		forecast:   make([]*WeatherForecast, 0),
		patterns:   make(map[string]*WeatherPattern),
		history:    make([]*Weather, 0, 100),
		callbacks:  make([]WeatherCallback, 0),
		lastUpdate: time.Now(),
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	wm.initializePatterns()
	wm.initializeWeather()
	return wm
}

// initializePatterns sets up seasonal weather patterns
func (wm *WeatherManager) initializePatterns() {
	// Spring pattern
	wm.patterns["spring"] = &WeatherPattern{
		Season:      "spring",
		CommonTypes: []WeatherType{WeatherSunny, WeatherCloudy, WeatherRainy, WeatherWindy},
		Temperature: Range{Min: 10, Max: 20},
		Humidity:    Range{Min: 40, Max: 70},
		WindSpeed:   Range{Min: 5, Max: 25},
	}

	// Summer pattern
	wm.patterns["summer"] = &WeatherPattern{
		Season:      "summer",
		CommonTypes: []WeatherType{WeatherSunny, WeatherHot, WeatherCloudy},
		Temperature: Range{Min: 20, Max: 35},
		Humidity:    Range{Min: 30, Max: 60},
		WindSpeed:   Range{Min: 0, Max: 15},
	}

	// Autumn pattern
	wm.patterns["autumn"] = &WeatherPattern{
		Season:      "autumn",
		CommonTypes: []WeatherType{WeatherCloudy, WeatherRainy, WeatherWindy, WeatherFoggy},
		Temperature: Range{Min: 5, Max: 15},
		Humidity:    Range{Min: 50, Max: 80},
		WindSpeed:   Range{Min: 10, Max: 30},
	}

	// Winter pattern
	wm.patterns["winter"] = &WeatherPattern{
		Season:      "winter",
		CommonTypes: []WeatherType{WeatherCold, WeatherSnowy, WeatherCloudy, WeatherStormy},
		Temperature: Range{Min: -10, Max: 5},
		Humidity:    Range{Min: 60, Max: 90},
		WindSpeed:   Range{Min: 15, Max: 40},
	}
}

// initializeWeather sets initial weather
func (wm *WeatherManager) initializeWeather() {
	// Start with sunny weather
	wm.currentWeather = &Weather{
		Type:        WeatherSunny,
		Severity:    SeverityMild,
		Temperature: 20,
		Humidity:    50,
		WindSpeed:   10,
		Duration:    8,
		StartTime:   time.Now(),
		Effects:     wm.calculateEffects(WeatherSunny, SeverityMild),
	}
}

// calculateEffects calculates weather effects on gameplay
func (wm *WeatherManager) calculateEffects(weatherType WeatherType, severity WeatherSeverity) *WeatherEffects {
	effects := &WeatherEffects{
		// Default values
		FoodDemand:       1.0,
		ClothingDemand:   1.0,
		ToolDemand:       1.0,
		LuxuryDemand:     1.0,
		TransportCost:    1.0,
		StorageCost:      1.0,
		CustomerTraffic:  1.0,
		ItemDurability:   1.0,
		TravelSpeed:      1.0,
		EventProbability: 1.0,
	}

	// Apply weather type effects
	switch weatherType {
	case WeatherSunny:
		effects.FoodDemand = 0.9      // Less food spoilage
		effects.LuxuryDemand = 1.2    // People feel good, buy luxuries
		effects.CustomerTraffic = 1.3 // More people out shopping
		effects.ItemDurability = 1.1  // Items last longer

	case WeatherRainy:
		effects.FoodDemand = 1.1      // Indoor activities, more eating
		effects.ClothingDemand = 1.3  // Rainwear needed
		effects.TransportCost = 1.2   // Slower transport
		effects.CustomerTraffic = 0.7 // Fewer people out
		effects.ItemDurability = 0.9  // Moisture damage

	case WeatherStormy:
		effects.TransportCost = 1.5    // Dangerous travel
		effects.StorageCost = 1.3      // Need better protection
		effects.CustomerTraffic = 0.3  // Very few customers
		effects.ItemDurability = 0.7   // Storm damage
		effects.EventProbability = 1.5 // More emergencies

	case WeatherSnowy:
		effects.FoodDemand = 1.3      // Stock up on food
		effects.ClothingDemand = 1.5  // Winter clothing
		effects.ToolDemand = 1.2      // Snow removal tools
		effects.TransportCost = 1.4   // Difficult travel
		effects.StorageCost = 1.2     // Heating costs
		effects.CustomerTraffic = 0.5 // Limited movement

	case WeatherHot:
		effects.FoodDemand = 0.8      // Less appetite, more drinks
		effects.ClothingDemand = 0.7  // Light clothing only
		effects.LuxuryDemand = 1.1    // Summer luxuries
		effects.StorageCost = 1.2     // Cooling costs
		effects.ItemDurability = 0.85 // Heat damage

	case WeatherCold:
		effects.FoodDemand = 1.4     // More calories needed
		effects.ClothingDemand = 1.6 // Heavy clothing
		effects.ToolDemand = 1.3     // Heating tools
		effects.TransportCost = 1.3  // Slower movement
		effects.StorageCost = 1.4    // Heating costs

	case WeatherFoggy:
		effects.TransportCost = 1.2   // Poor visibility
		effects.CustomerTraffic = 0.8 // Navigation difficulty
		effects.TravelSpeed = 0.7     // Slower travel

	case WeatherWindy:
		effects.TransportCost = 1.1    // Wind resistance
		effects.ItemDurability = 0.95  // Wind damage
		effects.EventProbability = 1.2 // More accidents
	}

	// Apply severity multiplier
	severityMultiplier := 1.0
	switch severity {
	case SeverityMild:
		severityMultiplier = 0.5
	case SeverityModerate:
		severityMultiplier = 1.0
	case SeveritySevere:
		severityMultiplier = 1.5
	case SeverityExtreme:
		severityMultiplier = 2.0
	}

	// Apply severity to effects (move them further from 1.0)
	effects.FoodDemand = 1.0 + (effects.FoodDemand-1.0)*severityMultiplier
	effects.ClothingDemand = 1.0 + (effects.ClothingDemand-1.0)*severityMultiplier
	effects.ToolDemand = 1.0 + (effects.ToolDemand-1.0)*severityMultiplier
	effects.LuxuryDemand = 1.0 + (effects.LuxuryDemand-1.0)*severityMultiplier
	effects.TransportCost = 1.0 + (effects.TransportCost-1.0)*severityMultiplier
	effects.StorageCost = 1.0 + (effects.StorageCost-1.0)*severityMultiplier
	effects.CustomerTraffic = 1.0 + (effects.CustomerTraffic-1.0)*severityMultiplier
	effects.ItemDurability = 1.0 + (effects.ItemDurability-1.0)*severityMultiplier
	effects.TravelSpeed = 1.0 + (effects.TravelSpeed-1.0)*severityMultiplier
	effects.EventProbability = 1.0 + (effects.EventProbability-1.0)*severityMultiplier

	return effects
}

// UpdateWeather updates current weather based on time and patterns
func (wm *WeatherManager) UpdateWeather(season string, currentHour int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Check if weather duration has expired
	elapsed := time.Since(wm.currentWeather.StartTime)
	if elapsed.Hours() < float64(wm.currentWeather.Duration) {
		return // Weather hasn't changed yet
	}

	// Store old weather
	oldWeather := wm.currentWeather
	wm.history = append(wm.history, oldWeather)
	if len(wm.history) > 100 {
		wm.history = wm.history[1:]
	}

	// Generate new weather based on season
	pattern, exists := wm.patterns[season]
	if !exists {
		pattern = wm.patterns["spring"] // Default
	}

	newWeather := wm.generateWeather(pattern, currentHour)
	wm.currentWeather = newWeather
	wm.lastUpdate = time.Now()

	// Generate forecast
	wm.generateForecast(pattern)

	// Notify callbacks
	for _, callback := range wm.callbacks {
		callback(oldWeather, newWeather)
	}
}

// generateWeather generates new weather based on pattern
func (wm *WeatherManager) generateWeather(pattern *WeatherPattern, currentHour int) *Weather {
	// Select weather type
	typeIndex := wm.rng.Intn(len(pattern.CommonTypes))
	weatherType := pattern.CommonTypes[typeIndex]

	// Determine severity
	severity := wm.generateSeverity()

	// Generate temperature
	tempRange := pattern.Temperature.Max - pattern.Temperature.Min
	temperature := pattern.Temperature.Min + wm.rng.Intn(tempRange+1)

	// Adjust temperature based on time of day
	if currentHour >= 0 && currentHour < 6 {
		temperature -= 5 // Early morning cold
	} else if currentHour >= 12 && currentHour < 16 {
		temperature += 5 // Afternoon heat
	}

	// Generate other parameters
	humidity := pattern.Humidity.Min + wm.rng.Intn(pattern.Humidity.Max-pattern.Humidity.Min+1)
	windSpeed := pattern.WindSpeed.Min + wm.rng.Intn(pattern.WindSpeed.Max-pattern.WindSpeed.Min+1)
	duration := 4 + wm.rng.Intn(12) // 4-16 hours

	return &Weather{
		Type:        weatherType,
		Severity:    severity,
		Temperature: temperature,
		Humidity:    humidity,
		WindSpeed:   windSpeed,
		Duration:    duration,
		StartTime:   time.Now(),
		Effects:     wm.calculateEffects(weatherType, severity),
	}
}

// generateSeverity generates weather severity with weighted probability
func (wm *WeatherManager) generateSeverity() WeatherSeverity {
	roll := wm.rng.Float64()
	if roll < 0.5 {
		return SeverityMild
	} else if roll < 0.8 {
		return SeverityModerate
	} else if roll < 0.95 {
		return SeveritySevere
	} else {
		return SeverityExtreme
	}
}

// generateForecast generates weather forecast
func (wm *WeatherManager) generateForecast(pattern *WeatherPattern) {
	wm.forecast = make([]*WeatherForecast, 0, 3)

	// Generate 3 forecasts
	for i := 1; i <= 3; i++ {
		// Generate possible weather
		typeIndex := wm.rng.Intn(len(pattern.CommonTypes))
		weatherType := pattern.CommonTypes[typeIndex]
		severity := wm.generateSeverity()

		forecastWeather := &Weather{
			Type:     weatherType,
			Severity: severity,
			Effects:  wm.calculateEffects(weatherType, severity),
		}

		// Calculate probability (decreases with time)
		probability := 0.9 - float64(i-1)*0.2
		if probability < 0.3 {
			probability = 0.3
		}

		forecast := &WeatherForecast{
			Weather:     forecastWeather,
			Probability: probability,
			TimeUntil:   time.Duration(wm.currentWeather.Duration+i*8) * time.Hour,
		}

		wm.forecast = append(wm.forecast, forecast)
	}
}

// GetCurrentWeather returns current weather
func (wm *WeatherManager) GetCurrentWeather() *Weather {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	return wm.currentWeather
}

// GetForecast returns weather forecast
func (wm *WeatherManager) GetForecast() []*WeatherForecast {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	return wm.forecast
}

// GetWeatherHistory returns recent weather history
func (wm *WeatherManager) GetWeatherHistory(count int) []*Weather {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if count > len(wm.history) {
		count = len(wm.history)
	}

	if count <= 0 {
		return []*Weather{}
	}

	start := len(wm.history) - count
	return wm.history[start:]
}

// GetMarketPriceModifier returns price modifier for item category
func (wm *WeatherManager) GetMarketPriceModifier(category string) float64 {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if wm.currentWeather == nil || wm.currentWeather.Effects == nil {
		return 1.0
	}

	switch category {
	case "FRUIT", "FOOD":
		return wm.currentWeather.Effects.FoodDemand
	case "CLOTHING", "ACCESSORY":
		return wm.currentWeather.Effects.ClothingDemand
	case "TOOL", "WEAPON":
		return wm.currentWeather.Effects.ToolDemand
	case "LUXURY", "GEM", "MAGIC":
		return wm.currentWeather.Effects.LuxuryDemand
	default:
		return 1.0
	}
}

// GetTransportCostModifier returns transport cost modifier
func (wm *WeatherManager) GetTransportCostModifier() float64 {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if wm.currentWeather == nil || wm.currentWeather.Effects == nil {
		return 1.0
	}

	return wm.currentWeather.Effects.TransportCost
}

// GetCustomerTrafficModifier returns customer traffic modifier
func (wm *WeatherManager) GetCustomerTrafficModifier() float64 {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if wm.currentWeather == nil || wm.currentWeather.Effects == nil {
		return 1.0
	}

	return wm.currentWeather.Effects.CustomerTraffic
}

// RegisterCallback registers a weather change callback
func (wm *WeatherManager) RegisterCallback(callback WeatherCallback) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.callbacks = append(wm.callbacks, callback)
}

// ForceWeather forces specific weather (for testing or events)
func (wm *WeatherManager) ForceWeather(weatherType WeatherType, severity WeatherSeverity, duration int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	oldWeather := wm.currentWeather

	wm.currentWeather = &Weather{
		Type:        weatherType,
		Severity:    severity,
		Temperature: 20,
		Humidity:    50,
		WindSpeed:   10,
		Duration:    duration,
		StartTime:   time.Now(),
		Effects:     wm.calculateEffects(weatherType, severity),
	}

	// Notify callbacks
	for _, callback := range wm.callbacks {
		callback(oldWeather, wm.currentWeather)
	}
}

// IsExtremeWeather checks if current weather is extreme
func (wm *WeatherManager) IsExtremeWeather() bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if wm.currentWeather == nil {
		return false
	}

	return wm.currentWeather.Severity == SeverityExtreme ||
		wm.currentWeather.Type == WeatherStormy
}

// GetWeatherWarnings returns any weather warnings
func (wm *WeatherManager) GetWeatherWarnings() []string {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	warnings := []string{}

	if wm.currentWeather == nil {
		return warnings
	}

	// Check for extreme weather
	if wm.currentWeather.Severity == SeverityExtreme {
		warnings = append(warnings, fmt.Sprintf("EXTREME %s WARNING: Take shelter immediately!", wm.currentWeather.Type.String()))
	}

	// Check for storms
	if wm.currentWeather.Type == WeatherStormy {
		warnings = append(warnings, "STORM WARNING: Travel not recommended!")
	}

	// Check for extreme temperatures
	if wm.currentWeather.Temperature > 35 {
		warnings = append(warnings, "HEAT WARNING: Stay hydrated and avoid outdoor activities!")
	} else if wm.currentWeather.Temperature < -5 {
		warnings = append(warnings, "COLD WARNING: Bundle up and limit exposure!")
	}

	// Check forecast for incoming severe weather
	for _, forecast := range wm.forecast {
		if forecast.Weather.Severity >= SeveritySevere && forecast.Probability > 0.7 {
			warnings = append(warnings, fmt.Sprintf("FORECAST: %s weather likely in %d hours",
				forecast.Weather.Type.String(), int(forecast.TimeUntil.Hours())))
		}
	}

	return warnings
}

// Reset resets weather to default
func (wm *WeatherManager) Reset() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.initializeWeather()
	wm.forecast = make([]*WeatherForecast, 0)
	wm.history = make([]*Weather, 0, 100)
	wm.lastUpdate = time.Now()
}
