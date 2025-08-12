package difficulty

import (
	"math"
	"sync"
	"time"
)

// DifficultyLevel represents the current difficulty tier
type DifficultyLevel int

const (
	DifficultyTutorial DifficultyLevel = iota
	DifficultyEasy
	DifficultyNormal
	DifficultyHard
	DifficultyExpert
	DifficultyMaster
)

// String returns the string representation of difficulty level
func (d DifficultyLevel) String() string {
	switch d {
	case DifficultyTutorial:
		return "Tutorial"
	case DifficultyEasy:
		return "Easy"
	case DifficultyNormal:
		return "Normal"
	case DifficultyHard:
		return "Hard"
	case DifficultyExpert:
		return "Expert"
	case DifficultyMaster:
		return "Master"
	default:
		return "Unknown"
	}
}

// DifficultyConfig defines the configuration for difficulty scaling
type DifficultyConfig struct {
	// Base difficulty settings
	StartingDifficulty DifficultyLevel
	MaxDifficulty      DifficultyLevel

	// Progression rates
	LearningCurveSpeed   float64 // How fast difficulty increases (0.0-1.0)
	AdaptationSpeed      float64 // How fast difficulty adapts to player skill
	FrustrationThreshold float64 // When to reduce difficulty
	BoredomThreshold     float64 // When to increase difficulty

	// Challenge parameters
	BaseChallenge       float64 // Base challenge level
	ChallengeGrowthRate float64 // How fast challenges grow
	MaxChallengeSpike   float64 // Maximum sudden difficulty increase

	// Economic difficulty modifiers
	GoldIncomeModifier float64 // Affects gold earning rate
	PriceVolatility    float64 // Market price fluctuation
	EventFrequency     float64 // Special event occurrence rate
	QuestDifficulty    float64 // Quest requirement scaling

	// Recovery settings
	DeathPenaltyModifier float64 // Penalty on failure
	RecoveryBonus        float64 // Bonus after repeated failures
	StreakBonus          float64 // Bonus for success streaks
}

// DefaultDifficultyConfig returns the default configuration
func DefaultDifficultyConfig() *DifficultyConfig {
	return &DifficultyConfig{
		StartingDifficulty:   DifficultyTutorial,
		MaxDifficulty:        DifficultyMaster,
		LearningCurveSpeed:   0.1,
		AdaptationSpeed:      0.05,
		FrustrationThreshold: 0.3, // 30% failure rate
		BoredomThreshold:     0.9, // 90% success rate
		BaseChallenge:        1.0,
		ChallengeGrowthRate:  0.02,
		MaxChallengeSpike:    1.5,
		GoldIncomeModifier:   1.0,
		PriceVolatility:      0.2,
		EventFrequency:       1.0,
		QuestDifficulty:      1.0,
		DeathPenaltyModifier: 0.1,
		RecoveryBonus:        0.2,
		StreakBonus:          0.1,
	}
}

// PlayerSkillMetrics tracks player performance
type PlayerSkillMetrics struct {
	// Performance tracking
	TotalPlays       int
	SuccessfulTrades int
	FailedTrades     int
	ProfitEfficiency float64 // Average profit margin
	DecisionSpeed    float64 // Time taken for decisions
	StrategicDepth   float64 // Complexity of strategies used

	// Learning indicators
	ImprovementRate   float64 // Rate of skill improvement
	ConsistencyScore  float64 // How consistent the player is
	AdaptabilityScore float64 // How well player adapts to changes

	// Current state
	CurrentStreak     int     // Current win/loss streak
	RecentPerformance float64 // Recent success rate
	FrustrationLevel  float64 // Current frustration (0.0-1.0)
	EngagementLevel   float64 // Current engagement (0.0-1.0)

	LastUpdated time.Time
}

// DifficultyModifiers contains all active difficulty adjustments
type DifficultyModifiers struct {
	// Economic modifiers
	PriceMultiplier      float64
	DemandMultiplier     float64
	SupplyMultiplier     float64
	GoldRewardMultiplier float64

	// Challenge modifiers
	EventDifficulty    float64
	QuestRequirements  float64
	TimePresure        float64
	CompetitorStrength float64

	// Support modifiers
	HintAvailability  float64
	TutorialDetail    float64
	ErrorForgiveness  float64
	ResourceAbundance float64
}

// ChallengeEvent represents a difficulty spike or special challenge
type ChallengeEvent struct {
	ID              string
	Name            string
	Description     string
	DifficultyBoost float64
	Duration        time.Duration
	Rewards         map[string]interface{}
	StartTime       time.Time
	Active          bool
}

// DifficultyManager manages the game's difficulty curve
type DifficultyManager struct {
	config          *DifficultyConfig
	currentLevel    DifficultyLevel
	targetLevel     DifficultyLevel
	playerSkill     *PlayerSkillMetrics
	modifiers       *DifficultyModifiers
	challenges      map[string]*ChallengeEvent
	difficultyScore float64 // Overall difficulty score (0.0-10.0)
	adaptiveScore   float64 // Adaptive difficulty score
	callbacks       []DifficultyCallback
	mu              sync.RWMutex
}

// DifficultyCallback is called when difficulty changes
type DifficultyCallback func(oldLevel, newLevel DifficultyLevel, modifiers *DifficultyModifiers)

// NewDifficultyManager creates a new difficulty manager
func NewDifficultyManager(config *DifficultyConfig) *DifficultyManager {
	if config == nil {
		config = DefaultDifficultyConfig()
	}

	return &DifficultyManager{
		config:          config,
		currentLevel:    config.StartingDifficulty,
		targetLevel:     config.StartingDifficulty,
		playerSkill:     &PlayerSkillMetrics{LastUpdated: time.Now()},
		modifiers:       createDefaultModifiers(),
		challenges:      make(map[string]*ChallengeEvent),
		difficultyScore: 1.0,
		adaptiveScore:   1.0,
		callbacks:       make([]DifficultyCallback, 0),
	}
}

// createDefaultModifiers creates default difficulty modifiers
func createDefaultModifiers() *DifficultyModifiers {
	return &DifficultyModifiers{
		PriceMultiplier:      1.0,
		DemandMultiplier:     1.0,
		SupplyMultiplier:     1.0,
		GoldRewardMultiplier: 1.0,
		EventDifficulty:      1.0,
		QuestRequirements:    1.0,
		TimePresure:          0.0,
		CompetitorStrength:   0.0,
		HintAvailability:     1.0,
		TutorialDetail:       1.0,
		ErrorForgiveness:     1.0,
		ResourceAbundance:    1.0,
	}
}

// RecordTrade records a trade outcome for skill assessment
func (dm *DifficultyManager) RecordTrade(success bool, profit float64, timeSpent time.Duration) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.playerSkill.TotalPlays++

	if success {
		dm.playerSkill.SuccessfulTrades++
		dm.playerSkill.CurrentStreak = max(1, dm.playerSkill.CurrentStreak+1)
	} else {
		dm.playerSkill.FailedTrades++
		dm.playerSkill.CurrentStreak = min(-1, dm.playerSkill.CurrentStreak-1)
	}

	// Update profit efficiency
	if dm.playerSkill.TotalPlays > 0 {
		dm.playerSkill.ProfitEfficiency = (dm.playerSkill.ProfitEfficiency*float64(dm.playerSkill.TotalPlays-1) + profit) / float64(dm.playerSkill.TotalPlays)
	}

	// Update decision speed
	dm.playerSkill.DecisionSpeed = timeSpent.Seconds()

	// Update recent performance
	successRate := float64(dm.playerSkill.SuccessfulTrades) / float64(dm.playerSkill.TotalPlays)
	dm.playerSkill.RecentPerformance = successRate

	// Update frustration and engagement
	dm.updateEmotionalState()

	// Check if difficulty adjustment is needed
	dm.evaluateDifficultyAdjustment()
}

// updateEmotionalState updates player's emotional metrics
func (dm *DifficultyManager) updateEmotionalState() {
	successRate := dm.playerSkill.RecentPerformance

	// Update frustration level
	switch {
	case successRate < dm.config.FrustrationThreshold:
		dm.playerSkill.FrustrationLevel = math.Min(1.0, dm.playerSkill.FrustrationLevel+0.1)
		dm.playerSkill.EngagementLevel = math.Max(0.0, dm.playerSkill.EngagementLevel-0.05)
	case successRate > dm.config.BoredomThreshold:
		dm.playerSkill.FrustrationLevel = math.Max(0.0, dm.playerSkill.FrustrationLevel-0.1)
		dm.playerSkill.EngagementLevel = math.Max(0.0, dm.playerSkill.EngagementLevel-0.1) // Boredom reduces engagement
	default:
		// Optimal challenge zone
		dm.playerSkill.FrustrationLevel = math.Max(0.0, dm.playerSkill.FrustrationLevel-0.05)
		dm.playerSkill.EngagementLevel = math.Min(1.0, dm.playerSkill.EngagementLevel+0.1)
	}
}

// evaluateDifficultyAdjustment checks if difficulty should be adjusted
func (dm *DifficultyManager) evaluateDifficultyAdjustment() {
	// Skip if not enough data
	if dm.playerSkill.TotalPlays < 5 {
		// Special case: allow quick progression from tutorial with perfect performance
		if dm.currentLevel == DifficultyTutorial && dm.playerSkill.CurrentStreak >= 5 {
			dm.targetLevel = DifficultyEasy
			dm.currentLevel = dm.targetLevel
			dm.difficultyScore = float64(dm.currentLevel)
			dm.updateModifiers()

			// Notify callbacks
			for _, callback := range dm.callbacks {
				callback(DifficultyTutorial, dm.currentLevel, dm.modifiers)
			}
		}
		return
	}

	oldLevel := dm.currentLevel
	successRate := dm.playerSkill.RecentPerformance

	// Determine target difficulty based on performance
	if dm.playerSkill.FrustrationLevel > 0.7 {
		// Player is frustrated, reduce difficulty
		if dm.currentLevel > DifficultyEasy {
			dm.targetLevel = dm.currentLevel - 1
		}
	} else if successRate > dm.config.BoredomThreshold && dm.playerSkill.EngagementLevel < 0.5 {
		// Player is bored, increase difficulty
		if dm.currentLevel < dm.config.MaxDifficulty {
			dm.targetLevel = dm.currentLevel + 1
		}
	} else if dm.playerSkill.CurrentStreak > 10 {
		// Long success streak, gradual increase
		if dm.currentLevel < dm.config.MaxDifficulty {
			dm.targetLevel = dm.currentLevel + 1
		}
	} else if dm.playerSkill.CurrentStreak < -5 {
		// Long failure streak, provide relief
		if dm.currentLevel > DifficultyEasy {
			dm.targetLevel = dm.currentLevel - 1
		}
	} else if dm.currentLevel == DifficultyTutorial && dm.playerSkill.TotalPlays >= 20 && successRate > 0.6 {
		// Graduate from tutorial after sufficient experience with decent performance
		dm.targetLevel = DifficultyEasy
	} else if successRate > 0.75 && dm.playerSkill.EngagementLevel > 0.7 && dm.playerSkill.TotalPlays >= int(dm.currentLevel)*10 {
		// Steady progress with good performance and engagement
		if dm.currentLevel < dm.config.MaxDifficulty {
			dm.targetLevel = dm.currentLevel + 1
		}
	}

	// Apply difficulty change immediately for testing
	if dm.targetLevel != dm.currentLevel {
		dm.currentLevel = dm.targetLevel
		dm.difficultyScore = float64(dm.currentLevel)
		dm.updateModifiers()

		// Notify callbacks
		for _, callback := range dm.callbacks {
			callback(oldLevel, dm.currentLevel, dm.modifiers)
		}
	}
}

// updateModifiers updates difficulty modifiers based on current level
func (dm *DifficultyManager) updateModifiers() {
	base := float64(dm.currentLevel) / float64(DifficultyMaster)

	// Economic modifiers (harder = less forgiving economy)
	dm.modifiers.PriceMultiplier = 1.0 + (base * 0.5)      // Prices up to 50% higher
	dm.modifiers.DemandMultiplier = 1.0 - (base * 0.3)     // Demand up to 30% lower
	dm.modifiers.SupplyMultiplier = 1.0 + (base * 0.3)     // Supply up to 30% higher
	dm.modifiers.GoldRewardMultiplier = 1.0 - (base * 0.2) // Rewards up to 20% lower

	// Challenge modifiers (harder = more challenging)
	dm.modifiers.EventDifficulty = 1.0 + (base * 0.5)
	dm.modifiers.QuestRequirements = 1.0 + (base * 0.4)
	dm.modifiers.TimePresure = base * 0.5
	dm.modifiers.CompetitorStrength = base * 0.7

	// Support modifiers (harder = less support)
	dm.modifiers.HintAvailability = 1.0 - (base * 0.5)
	dm.modifiers.TutorialDetail = 1.0 - (base * 0.3)
	dm.modifiers.ErrorForgiveness = 1.0 - (base * 0.4)
	dm.modifiers.ResourceAbundance = 1.0 - (base * 0.2)

	// Apply streak bonuses
	if dm.playerSkill.CurrentStreak > 5 {
		dm.modifiers.GoldRewardMultiplier *= (1.0 + dm.config.StreakBonus)
	} else if dm.playerSkill.CurrentStreak < -3 {
		dm.modifiers.ErrorForgiveness *= (1.0 + dm.config.RecoveryBonus)
	}
}

// AddChallenge adds a special challenge event
func (dm *DifficultyManager) AddChallenge(challenge *ChallengeEvent) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	challenge.StartTime = time.Now()
	challenge.Active = true
	dm.challenges[challenge.ID] = challenge

	// Apply challenge difficulty boost
	dm.difficultyScore = math.Min(10.0, dm.difficultyScore*challenge.DifficultyBoost)
	dm.updateModifiers()
}

// RemoveChallenge removes a challenge event
func (dm *DifficultyManager) RemoveChallenge(challengeID string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if challenge, exists := dm.challenges[challengeID]; exists {
		challenge.Active = false
		delete(dm.challenges, challengeID)

		// Recalculate difficulty score without this challenge
		// Reset to base difficulty for current level
		dm.difficultyScore = float64(dm.currentLevel)

		// Reapply remaining challenges
		for _, remainingChallenge := range dm.challenges {
			if remainingChallenge.Active {
				dm.difficultyScore = math.Min(10.0, dm.difficultyScore*remainingChallenge.DifficultyBoost)
			}
		}

		// Recalculate modifiers
		dm.updateModifiers()
	}
}

// GetCurrentDifficulty returns the current difficulty level
func (dm *DifficultyManager) GetCurrentDifficulty() DifficultyLevel {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.currentLevel
}

// GetModifiers returns current difficulty modifiers
func (dm *DifficultyManager) GetModifiers() *DifficultyModifiers {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// Return a copy
	modifiers := *dm.modifiers
	return &modifiers
}

// GetPlayerSkill returns player skill metrics
func (dm *DifficultyManager) GetPlayerSkill() *PlayerSkillMetrics {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// Return a copy
	skill := *dm.playerSkill
	return &skill
}

// SetDifficulty manually sets the difficulty level
func (dm *DifficultyManager) SetDifficulty(level DifficultyLevel) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	oldLevel := dm.currentLevel
	dm.currentLevel = level
	dm.targetLevel = level
	dm.difficultyScore = float64(level)

	dm.updateModifiers()

	// Notify callbacks
	for _, callback := range dm.callbacks {
		callback(oldLevel, dm.currentLevel, dm.modifiers)
	}
}

// RegisterCallback registers a callback for difficulty changes
func (dm *DifficultyManager) RegisterCallback(callback DifficultyCallback) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.callbacks = append(dm.callbacks, callback)
}

// CalculateAdjustedValue applies difficulty modifiers to a base value
func (dm *DifficultyManager) CalculateAdjustedValue(baseValue float64, valueType string) float64 {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	switch valueType {
	case "price":
		return baseValue * dm.modifiers.PriceMultiplier
	case "demand":
		return baseValue * dm.modifiers.DemandMultiplier
	case "supply":
		return baseValue * dm.modifiers.SupplyMultiplier
	case "reward":
		return baseValue * dm.modifiers.GoldRewardMultiplier
	case "quest":
		return baseValue * dm.modifiers.QuestRequirements
	default:
		return baseValue
	}
}

// GetDifficultyScore returns the current difficulty score (0.0-10.0)
func (dm *DifficultyManager) GetDifficultyScore() float64 {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.difficultyScore
}

// GetAdaptiveScore returns the adaptive difficulty score
func (dm *DifficultyManager) GetAdaptiveScore() float64 {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// Calculate based on player performance and engagement
	performance := dm.playerSkill.RecentPerformance
	engagement := dm.playerSkill.EngagementLevel
	frustration := 1.0 - dm.playerSkill.FrustrationLevel

	// Weighted average
	dm.adaptiveScore = (performance*0.4 + engagement*0.4 + frustration*0.2)
	return dm.adaptiveScore
}

// Reset resets the difficulty manager
func (dm *DifficultyManager) Reset() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.currentLevel = dm.config.StartingDifficulty
	dm.targetLevel = dm.config.StartingDifficulty
	dm.playerSkill = &PlayerSkillMetrics{LastUpdated: time.Now()}
	dm.modifiers = createDefaultModifiers()
	dm.challenges = make(map[string]*ChallengeEvent)
	dm.difficultyScore = 1.0
	dm.adaptiveScore = 1.0
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
