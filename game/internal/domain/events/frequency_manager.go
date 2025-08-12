package events

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// FrequencyManager manages event frequency and timing
type FrequencyManager struct {
	mu             sync.RWMutex
	config         *FrequencyConfig
	eventHistory   map[string]*EventFrequencyData
	currentPhase   GamePhase
	playerLevel    int
	dayCount       int
	lastAdjustment time.Time
	rng            *rand.Rand
}

// FrequencyConfig defines frequency settings for events
type FrequencyConfig struct {
	// Base frequencies (events per day)
	BaseCommonFrequency    float64
	BaseUncommonFrequency  float64
	BaseRareFrequency      float64
	BaseEpicFrequency      float64
	BaseLegendaryFrequency float64

	// Modifiers
	PlayerLevelMultiplier float64
	DayProgressMultiplier float64
	RecentEventPenalty    float64
	VarietyBonus          float64

	// Cooldowns (in game days)
	MinCooldownCommon    int
	MinCooldownUncommon  int
	MinCooldownRare      int
	MinCooldownEpic      int
	MinCooldownLegendary int

	// Phase-specific multipliers
	PhaseMultipliers map[GamePhase]float64

	// Event clustering
	ClusteringEnabled  bool
	ClusterProbability float64
	MaxClusterSize     int
}

// EventFrequencyData tracks frequency data for an event
type EventFrequencyData struct {
	EventID         string
	EventType       EventRarity
	LastOccurred    time.Time
	OccurrenceCount int
	AverageInterval time.Duration
	NextEligible    time.Time
	Weight          float64
}

// EventRarity represents how rare an event is
type EventRarity int

const (
	RarityCommon EventRarity = iota
	RarityUncommon
	RarityRare
	RarityEpic
	RarityLegendary
)

// GamePhase represents the current game phase
type GamePhase int

const (
	PhaseEarlyGame GamePhase = iota
	PhaseMidGame
	PhaseLateGame
	PhaseEndGame
)

// NewFrequencyManager creates a new frequency manager
func NewFrequencyManager() *FrequencyManager {
	return &FrequencyManager{
		config:       GetDefaultConfig(),
		eventHistory: make(map[string]*EventFrequencyData),
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetDefaultConfig returns default frequency configuration
func GetDefaultConfig() *FrequencyConfig {
	return &FrequencyConfig{
		// Base frequencies (events per game day)
		BaseCommonFrequency:    3.0,
		BaseUncommonFrequency:  1.5,
		BaseRareFrequency:      0.5,
		BaseEpicFrequency:      0.2,
		BaseLegendaryFrequency: 0.05,

		// Modifiers
		PlayerLevelMultiplier: 0.05, // 5% increase per level
		DayProgressMultiplier: 0.02, // 2% increase per day
		RecentEventPenalty:    0.3,  // 30% reduction for recent events
		VarietyBonus:          1.2,  // 20% bonus for variety

		// Cooldowns (game days)
		MinCooldownCommon:    0,
		MinCooldownUncommon:  1,
		MinCooldownRare:      3,
		MinCooldownEpic:      7,
		MinCooldownLegendary: 14,

		// Phase multipliers
		PhaseMultipliers: map[GamePhase]float64{
			PhaseEarlyGame: 0.8, // Fewer events early
			PhaseMidGame:   1.0, // Normal frequency
			PhaseLateGame:  1.2, // More events late
			PhaseEndGame:   1.5, // Most events at endgame
		},

		// Clustering
		ClusteringEnabled:  true,
		ClusterProbability: 0.15, // 15% chance of clustering
		MaxClusterSize:     3,
	}
}

// UpdateGameState updates the manager with current game state
func (fm *FrequencyManager) UpdateGameState(phase GamePhase, playerLevel, dayCount int) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.currentPhase = phase
	fm.playerLevel = playerLevel
	fm.dayCount = dayCount
}

// ShouldTriggerEvent determines if an event should trigger
func (fm *FrequencyManager) ShouldTriggerEvent(eventID string, rarity EventRarity) bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Check cooldown
	if !fm.checkCooldown(eventID, rarity) {
		return false
	}

	// Calculate probability
	probability := fm.calculateEventProbability(eventID, rarity)

	// Random check
	roll := fm.rng.Float64()
	shouldTrigger := roll < probability

	// Update history if triggered
	if shouldTrigger {
		fm.recordEventOccurrence(eventID, rarity)
	}

	return shouldTrigger
}

// checkCooldown checks if event is off cooldown
func (fm *FrequencyManager) checkCooldown(eventID string, rarity EventRarity) bool {
	data, exists := fm.eventHistory[eventID]
	if !exists {
		return true // Never occurred, no cooldown
	}

	// Check if enough time has passed
	return time.Now().After(data.NextEligible)
}

// calculateEventProbability calculates the probability of an event occurring
func (fm *FrequencyManager) calculateEventProbability(eventID string, rarity EventRarity) float64 {
	// Get base frequency
	baseFreq := fm.getBaseFrequency(rarity)

	// Apply phase multiplier
	phaseMultiplier := fm.config.PhaseMultipliers[fm.currentPhase]

	// Apply player level multiplier
	levelMultiplier := 1.0 + (float64(fm.playerLevel) * fm.config.PlayerLevelMultiplier)

	// Apply day progress multiplier
	dayMultiplier := 1.0 + (float64(fm.dayCount) * fm.config.DayProgressMultiplier)

	// Apply recent event penalty
	recentPenalty := fm.calculateRecentPenalty(eventID)

	// Apply variety bonus
	varietyBonus := fm.calculateVarietyBonus()

	// Calculate final probability
	probability := baseFreq * phaseMultiplier * levelMultiplier * dayMultiplier * recentPenalty * varietyBonus

	// Cap at reasonable maximum (50% chance)
	return math.Min(probability, 0.5)
}

// getBaseFrequency returns base frequency for rarity
func (fm *FrequencyManager) getBaseFrequency(rarity EventRarity) float64 {
	switch rarity {
	case RarityCommon:
		return fm.config.BaseCommonFrequency / 100.0 // Convert to probability
	case RarityUncommon:
		return fm.config.BaseUncommonFrequency / 100.0
	case RarityRare:
		return fm.config.BaseRareFrequency / 100.0
	case RarityEpic:
		return fm.config.BaseEpicFrequency / 100.0
	case RarityLegendary:
		return fm.config.BaseLegendaryFrequency / 100.0
	default:
		return fm.config.BaseCommonFrequency / 100.0
	}
}

// calculateRecentPenalty calculates penalty for recently occurred events
func (fm *FrequencyManager) calculateRecentPenalty(eventID string) float64 {
	data, exists := fm.eventHistory[eventID]
	if !exists {
		return 1.0 // No penalty for first occurrence
	}

	// Calculate time since last occurrence
	timeSince := time.Since(data.LastOccurred)
	daysSince := timeSince.Hours() / 24

	// Apply penalty if recent (within 7 days)
	if daysSince < 7 {
		penaltyFactor := daysSince / 7.0 // Linear scaling
		return fm.config.RecentEventPenalty + (1.0-fm.config.RecentEventPenalty)*penaltyFactor
	}

	return 1.0 // No penalty
}

// calculateVarietyBonus calculates bonus for event variety
func (fm *FrequencyManager) calculateVarietyBonus() float64 {
	// Count unique events in recent history
	recentThreshold := time.Now().Add(-24 * time.Hour * 3) // Last 3 days
	uniqueRecent := 0

	for _, data := range fm.eventHistory {
		if data.LastOccurred.After(recentThreshold) {
			uniqueRecent++
		}
	}

	// More variety = higher bonus
	if uniqueRecent >= 5 {
		return fm.config.VarietyBonus
	} else if uniqueRecent >= 3 {
		return 1.0 + (fm.config.VarietyBonus-1.0)*0.5
	}

	return 1.0 // No bonus
}

// recordEventOccurrence records that an event occurred
func (fm *FrequencyManager) recordEventOccurrence(eventID string, rarity EventRarity) {
	now := time.Now()

	data, exists := fm.eventHistory[eventID]
	if !exists {
		data = &EventFrequencyData{
			EventID:   eventID,
			EventType: rarity,
		}
		fm.eventHistory[eventID] = data
	}

	// Update occurrence data
	if data.OccurrenceCount > 0 {
		interval := now.Sub(data.LastOccurred)
		// Update average interval (exponential moving average)
		alpha := 0.3 // Smoothing factor
		if data.AverageInterval == 0 {
			data.AverageInterval = interval
		} else {
			data.AverageInterval = time.Duration(
				alpha*float64(interval) + (1-alpha)*float64(data.AverageInterval),
			)
		}
	}

	data.LastOccurred = now
	data.OccurrenceCount++
	data.NextEligible = now.Add(time.Duration(fm.getCooldownDays(rarity)) * 24 * time.Hour)
}

// getCooldownDays returns cooldown period for rarity
func (fm *FrequencyManager) getCooldownDays(rarity EventRarity) int {
	switch rarity {
	case RarityCommon:
		return fm.config.MinCooldownCommon
	case RarityUncommon:
		return fm.config.MinCooldownUncommon
	case RarityRare:
		return fm.config.MinCooldownRare
	case RarityEpic:
		return fm.config.MinCooldownEpic
	case RarityLegendary:
		return fm.config.MinCooldownLegendary
	default:
		return fm.config.MinCooldownCommon
	}
}

// GetNextEvents suggests next events based on frequency
func (fm *FrequencyManager) GetNextEvents(count int) []EventSuggestion {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	suggestions := make([]EventSuggestion, 0)

	// Calculate weights for all eligible events
	for eventID, data := range fm.eventHistory {
		if fm.checkCooldown(eventID, data.EventType) {
			weight := fm.calculateEventWeight(eventID, data)
			suggestions = append(suggestions, EventSuggestion{
				EventID:     eventID,
				Weight:      weight,
				Probability: fm.calculateEventProbability(eventID, data.EventType),
				Rarity:      data.EventType,
			})
		}
	}

	// Sort by weight
	sortEventSuggestions(suggestions)

	// Return top N
	if len(suggestions) > count {
		return suggestions[:count]
	}
	return suggestions
}

// EventSuggestion represents a suggested event
type EventSuggestion struct {
	EventID     string
	Weight      float64
	Probability float64
	Rarity      EventRarity
}

// calculateEventWeight calculates weight for event selection
func (fm *FrequencyManager) calculateEventWeight(eventID string, data *EventFrequencyData) float64 {
	// Base weight by rarity
	weight := 1.0 / (float64(data.EventType) + 1.0)

	// Adjust by time since last occurrence
	daysSince := time.Since(data.LastOccurred).Hours() / 24
	timeWeight := math.Min(daysSince/7.0, 2.0) // Cap at 2x after a week

	// Adjust by occurrence count (prefer less frequent)
	frequencyWeight := 1.0 / math.Max(float64(data.OccurrenceCount), 1.0)

	return weight * timeWeight * frequencyWeight
}

// sortEventSuggestions sorts suggestions by weight (descending)
func sortEventSuggestions(suggestions []EventSuggestion) {
	for i := 0; i < len(suggestions)-1; i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[i].Weight < suggestions[j].Weight {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}
}

// AdjustFrequency dynamically adjusts frequency based on player engagement
func (fm *FrequencyManager) AdjustFrequency(engagementScore float64) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Don't adjust too frequently
	if time.Since(fm.lastAdjustment) < time.Hour {
		return
	}

	// Engagement score: 0.0 (low) to 1.0 (high)
	// Adjust base frequencies based on engagement
	if engagementScore < 0.3 {
		// Low engagement - reduce frequency
		fm.config.BaseCommonFrequency *= 0.9
		fm.config.BaseUncommonFrequency *= 0.9
	} else if engagementScore > 0.7 {
		// High engagement - increase frequency
		fm.config.BaseCommonFrequency *= 1.1
		fm.config.BaseUncommonFrequency *= 1.1
	}

	// Keep frequencies within reasonable bounds
	fm.config.BaseCommonFrequency = math.Max(1.0, math.Min(5.0, fm.config.BaseCommonFrequency))
	fm.config.BaseUncommonFrequency = math.Max(0.5, math.Min(2.5, fm.config.BaseUncommonFrequency))

	fm.lastAdjustment = time.Now()
}

// ShouldClusterEvents determines if events should cluster
func (fm *FrequencyManager) ShouldClusterEvents() (bool, int) {
	if !fm.config.ClusteringEnabled {
		return false, 0
	}

	// Random check for clustering
	if fm.rng.Float64() < fm.config.ClusterProbability {
		// Determine cluster size
		size := fm.rng.Intn(fm.config.MaxClusterSize) + 1
		return true, size
	}

	return false, 0
}

// GetEventStats returns statistics about event frequency
func (fm *FrequencyManager) GetEventStats() EventFrequencyStats {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	stats := EventFrequencyStats{
		TotalEvents:      len(fm.eventHistory),
		EventsByRarity:   make(map[EventRarity]int),
		AverageIntervals: make(map[EventRarity]time.Duration),
	}

	// Calculate statistics
	for _, data := range fm.eventHistory {
		stats.EventsByRarity[data.EventType]++
		stats.TotalOccurrences += data.OccurrenceCount

		if data.AverageInterval > 0 {
			current := stats.AverageIntervals[data.EventType]
			count := stats.EventsByRarity[data.EventType]
			// Update running average
			stats.AverageIntervals[data.EventType] =
				(current*time.Duration(count-1) + data.AverageInterval) / time.Duration(count)
		}
	}

	return stats
}

// EventFrequencyStats contains frequency statistics
type EventFrequencyStats struct {
	TotalEvents      int
	TotalOccurrences int
	EventsByRarity   map[EventRarity]int
	AverageIntervals map[EventRarity]time.Duration
}

// ResetEventHistory clears event history (for new game)
func (fm *FrequencyManager) ResetEventHistory() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.eventHistory = make(map[string]*EventFrequencyData)
	fm.dayCount = 0
	fm.playerLevel = 1
	fm.currentPhase = PhaseEarlyGame
}

// SetConfig updates the frequency configuration
func (fm *FrequencyManager) SetConfig(config *FrequencyConfig) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.config = config
}
