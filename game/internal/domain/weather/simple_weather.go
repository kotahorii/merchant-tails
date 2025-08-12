package weather

import (
	"math/rand"
	"sync"
	"time"
)

// SimpleWeatherType represents basic weather conditions
type SimpleWeatherType int

const (
	WeatherSunny SimpleWeatherType = iota
	WeatherRainy
	WeatherCloudy
)

// SimpleWeather represents the simplified weather system
type SimpleWeather struct {
	currentWeather SimpleWeatherType
	priceEffect    float64
	demandEffect   float64
	mu             sync.RWMutex
}

// NewSimpleWeather creates a basic weather system
func NewSimpleWeather() *SimpleWeather {
	return &SimpleWeather{
		currentWeather: WeatherSunny,
		priceEffect:    1.0,
		demandEffect:   1.0,
	}
}

// UpdateWeather randomly changes the weather
func (sw *SimpleWeather) UpdateWeather() {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	// Simple random weather change
	rand.Seed(time.Now().UnixNano())
	weatherType := SimpleWeatherType(rand.Intn(3))

	sw.currentWeather = weatherType

	// Set simple effects based on weather
	switch weatherType {
	case WeatherSunny:
		sw.priceEffect = 0.95 // Slightly lower prices (good harvest)
		sw.demandEffect = 1.1 // Higher customer traffic
	case WeatherRainy:
		sw.priceEffect = 1.05 // Slightly higher prices (transport issues)
		sw.demandEffect = 0.9 // Lower customer traffic
	case WeatherCloudy:
		sw.priceEffect = 1.0  // Normal prices
		sw.demandEffect = 1.0 // Normal traffic
	}
}

// GetCurrentWeather returns the current weather
func (sw *SimpleWeather) GetCurrentWeather() string {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	switch sw.currentWeather {
	case WeatherSunny:
		return "Sunny"
	case WeatherRainy:
		return "Rainy"
	case WeatherCloudy:
		return "Cloudy"
	default:
		return "Unknown"
	}
}

// GetPriceEffect returns the price multiplier for current weather
func (sw *SimpleWeather) GetPriceEffect() float64 {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.priceEffect
}

// GetDemandEffect returns the demand multiplier for current weather
func (sw *SimpleWeather) GetDemandEffect() float64 {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.demandEffect
}

// GetWeatherEffects returns all weather effects
func (sw *SimpleWeather) GetWeatherEffects() map[string]float64 {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	return map[string]float64{
		"price_effect":  sw.priceEffect,
		"demand_effect": sw.demandEffect,
	}
}

// GetWeatherDescription returns a simple description of the weather
func (sw *SimpleWeather) GetWeatherDescription() string {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	switch sw.currentWeather {
	case WeatherSunny:
		return "Clear skies boost customer traffic!"
	case WeatherRainy:
		return "Rain keeps some customers away."
	case WeatherCloudy:
		return "Overcast but business as usual."
	default:
		return "Normal weather conditions."
	}
}

// SetWeather manually sets the weather (for testing or events)
func (sw *SimpleWeather) SetWeather(weather SimpleWeatherType) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.currentWeather = weather

	// Update effects
	switch weather {
	case WeatherSunny:
		sw.priceEffect = 0.95
		sw.demandEffect = 1.1
	case WeatherRainy:
		sw.priceEffect = 1.05
		sw.demandEffect = 0.9
	case WeatherCloudy:
		sw.priceEffect = 1.0
		sw.demandEffect = 1.0
	}
}
