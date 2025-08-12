package difficulty

import "sync"

// SimpleDifficulty represents simplified difficulty levels
type SimpleDifficulty int

const (
	DifficultyEasy SimpleDifficulty = iota
	DifficultyNormal
	DifficultyHard
)

// SimpleDifficultyManager manages game difficulty
type SimpleDifficultyManager struct {
	currentDifficulty SimpleDifficulty
	priceMultiplier   float64
	demandMultiplier  float64
	eventFrequency    float64
	mu                sync.RWMutex
}

// NewSimpleDifficultyManager creates a new difficulty manager
func NewSimpleDifficultyManager() *SimpleDifficultyManager {
	return &SimpleDifficultyManager{
		currentDifficulty: DifficultyNormal,
		priceMultiplier:   1.0,
		demandMultiplier:  1.0,
		eventFrequency:    1.0,
	}
}

// SetDifficulty sets the game difficulty
func (dm *SimpleDifficultyManager) SetDifficulty(difficulty SimpleDifficulty) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.currentDifficulty = difficulty

	switch difficulty {
	case DifficultyEasy:
		dm.priceMultiplier = 0.8  // Lower prices
		dm.demandMultiplier = 1.2 // Higher demand
		dm.eventFrequency = 0.5   // Fewer negative events
	case DifficultyNormal:
		dm.priceMultiplier = 1.0
		dm.demandMultiplier = 1.0
		dm.eventFrequency = 1.0
	case DifficultyHard:
		dm.priceMultiplier = 1.2  // Higher prices
		dm.demandMultiplier = 0.8 // Lower demand
		dm.eventFrequency = 1.5   // More frequent events
	}
}

// GetDifficulty returns the current difficulty
func (dm *SimpleDifficultyManager) GetDifficulty() SimpleDifficulty {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.currentDifficulty
}

// GetPriceMultiplier returns the price adjustment for difficulty
func (dm *SimpleDifficultyManager) GetPriceMultiplier() float64 {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.priceMultiplier
}

// GetDemandMultiplier returns the demand adjustment for difficulty
func (dm *SimpleDifficultyManager) GetDemandMultiplier() float64 {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.demandMultiplier
}

// GetEventFrequency returns the event frequency multiplier
func (dm *SimpleDifficultyManager) GetEventFrequency() float64 {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.eventFrequency
}

// GetDifficultyName returns the name of the current difficulty
func (dm *SimpleDifficultyManager) GetDifficultyName() string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	switch dm.currentDifficulty {
	case DifficultyEasy:
		return "Easy"
	case DifficultyNormal:
		return "Normal"
	case DifficultyHard:
		return "Hard"
	default:
		return "Unknown"
	}
}

// GetDifficultySettings returns all difficulty settings
func (dm *SimpleDifficultyManager) GetDifficultySettings() map[string]float64 {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return map[string]float64{
		"price_multiplier":  dm.priceMultiplier,
		"demand_multiplier": dm.demandMultiplier,
		"event_frequency":   dm.eventFrequency,
	}
}

// ApplyPriceDifficulty applies difficulty to a base price
func (dm *SimpleDifficultyManager) ApplyPriceDifficulty(basePrice float64) float64 {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return basePrice * dm.priceMultiplier
}

// ApplyDemandDifficulty applies difficulty to base demand
func (dm *SimpleDifficultyManager) ApplyDemandDifficulty(baseDemand float64) float64 {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return baseDemand * dm.demandMultiplier
}

// ShouldTriggerEvent determines if an event should trigger based on difficulty
func (dm *SimpleDifficultyManager) ShouldTriggerEvent(baseChance float64) bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	adjustedChance := baseChance * dm.eventFrequency
	// Simple random check (in real implementation, use proper random)
	// For simplicity, just return based on threshold
	return adjustedChance > 0.5
}
