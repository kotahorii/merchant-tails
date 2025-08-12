package progression

import (
	"sync"
)

// MilestoneType represents the type of milestone
type MilestoneType int

const (
	MilestoneTypeGold MilestoneType = iota
	MilestoneTypeProfit
	MilestoneTypeTrades
	MilestoneTypeRank
)

// Milestone represents a game milestone
type Milestone struct {
	ID        string
	Name      string
	Threshold int
	Type      MilestoneType
	Reward    int // Experience points
	Achieved  bool
}

// MilestoneReward represents rewards for achieving a milestone
type MilestoneReward struct {
	Experience    int
	Gold          int
	AchievementID string
}

// TradeResult represents the result of a trade
type TradeResult struct {
	ExperienceGained     int
	BonusExperience      int
	AchievementsUnlocked []string
	FeaturesUnlocked     []string
	RankUp               bool
	NewRank              Rank
}

// ProgressionManager manages all progression systems
type ProgressionManager struct {
	rankSystem          *RankSystem
	achievementSystem   *AchievementSystem
	featureUnlockSystem *FeatureUnlockSystem
	playerStats         *PlayerStats
	milestones          map[string]*Milestone
	unlockedFeatures    map[string]bool
	tradeCounter        int
	mu                  sync.RWMutex
}

// NewProgressionManager creates a new progression manager
func NewProgressionManager() *ProgressionManager {
	return &ProgressionManager{
		rankSystem:          NewRankSystem(),
		achievementSystem:   NewAchievementSystem(),
		featureUnlockSystem: NewFeatureUnlockSystem(),
		playerStats:         NewPlayerStats(),
		milestones:          make(map[string]*Milestone),
		unlockedFeatures:    make(map[string]bool),
		tradeCounter:        0,
	}
}

// GetRankSystem returns the rank system
func (pm *ProgressionManager) GetRankSystem() *RankSystem {
	return pm.rankSystem
}

// GetAchievementSystem returns the achievement system
func (pm *ProgressionManager) GetAchievementSystem() *AchievementSystem {
	return pm.achievementSystem
}

// GetFeatureUnlockSystem returns the feature unlock system
func (pm *ProgressionManager) GetFeatureUnlockSystem() *FeatureUnlockSystem {
	return pm.featureUnlockSystem
}

// GetPlayerStats returns the player statistics
func (pm *ProgressionManager) GetPlayerStats() *PlayerStats {
	return pm.playerStats
}

// HandleTradeCompletion handles a completed trade
func (pm *ProgressionManager) HandleTradeCompletion(buyPrice, sellPrice int) *TradeResult {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	result := &TradeResult{
		ExperienceGained:     10, // Base experience for any trade
		BonusExperience:      0,
		AchievementsUnlocked: []string{},
		FeaturesUnlocked:     []string{},
		RankUp:               false,
	}

	// Calculate profit
	profit := sellPrice - buyPrice

	// Add bonus experience for profitable trades
	if profit > 0 {
		result.BonusExperience = profit / 10 // 10% of profit as bonus XP
		if result.BonusExperience < 5 {
			result.BonusExperience = 5 // Minimum 5 bonus XP
		}
	}

	// Record trade in statistics
	pm.playerStats.RecordTrade(buyPrice, sellPrice)

	// Update trade counter
	pm.tradeCounter++

	// Check for first trade achievement
	if pm.tradeCounter == 1 {
		if pm.achievementSystem.UnlockAchievement("first_trade") {
			result.AchievementsUnlocked = append(result.AchievementsUnlocked, "first_trade")
		}
	}

	// Update progressive achievements
	pm.achievementSystem.UpdateProgress("trades_10", 1)
	pm.achievementSystem.UpdateProgress("trade_10", 1) // Support both naming conventions
	pm.achievementSystem.UpdateProgress("trade_master", 1)

	// Add experience to rank system
	totalExp := result.ExperienceGained + result.BonusExperience
	promoted := pm.rankSystem.AddExperience(totalExp)

	if promoted {
		result.RankUp = true
		result.NewRank = pm.rankSystem.GetCurrentRank()
	}

	return result
}

// HandleItemTrade handles a trade for a specific item
func (pm *ProgressionManager) HandleItemTrade(itemID string, buyPrice, sellPrice int) *TradeResult {
	// Record item-specific statistics
	pm.playerStats.RecordItemTrade(itemID, buyPrice, sellPrice)

	// Handle as a normal trade
	return pm.HandleTradeCompletion(buyPrice, sellPrice)
}

// IsFeatureAvailable checks if a feature is available to unlock
func (pm *ProgressionManager) IsFeatureAvailable(featureID string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Get current rank
	currentRank := pm.rankSystem.GetCurrentRank()

	// Get unlocked achievements
	unlockedAchievements := []string{}
	for _, achievement := range pm.achievementSystem.GetUnlockedAchievements() {
		unlockedAchievements = append(unlockedAchievements, achievement.ID)
	}

	// Get unlocked features
	unlockedFeatures := []string{}
	for id, unlocked := range pm.unlockedFeatures {
		if unlocked {
			unlockedFeatures = append(unlockedFeatures, id)
		}
	}

	// Check with all contexts
	return pm.featureUnlockSystem.IsAvailableWithAchievements(featureID, currentRank, unlockedAchievements) &&
		pm.featureUnlockSystem.IsAvailableWithFeatures(featureID, currentRank, unlockedFeatures)
}

// UnlockFeature unlocks a feature
func (pm *ProgressionManager) UnlockFeature(featureID string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.featureUnlockSystem.UnlockFeature(featureID) {
		pm.unlockedFeatures[featureID] = true
		return true
	}
	return false
}

// RegisterMilestone registers a new milestone
func (pm *ProgressionManager) RegisterMilestone(milestone *Milestone) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.milestones[milestone.ID] = milestone
}

// CheckMilestone checks if a milestone has been reached
func (pm *ProgressionManager) CheckMilestone(milestoneType MilestoneType, value int) *MilestoneReward {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, milestone := range pm.milestones {
		if milestone.Type == milestoneType &&
			!milestone.Achieved && value >= milestone.Threshold {

			// Mark as achieved
			milestone.Achieved = true

			// Grant rewards
			reward := &MilestoneReward{
				Experience: milestone.Reward,
			}

			// Add experience
			pm.rankSystem.AddExperience(milestone.Reward)

			return reward
		}
	}

	return nil
}

// GetProgressionScore calculates an overall progression score
func (pm *ProgressionManager) GetProgressionScore() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	score := 0

	// Add rank score
	rank := pm.rankSystem.GetCurrentRank()
	score += int(rank) * 1000

	// Add achievement points
	score += pm.achievementSystem.GetTotalPoints()

	// Add feature unlock count
	for _, unlocked := range pm.unlockedFeatures {
		if unlocked {
			score += 100
		}
	}

	// Add trading success
	successRate := pm.playerStats.GetSuccessRate()
	score += int(successRate * 500)

	return score
}

// GetProgressionSummary returns a summary of player progression
func (pm *ProgressionManager) GetProgressionSummary() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return map[string]interface{}{
		"rank":              GetRankName(pm.rankSystem.GetCurrentRank()),
		"experience":        pm.rankSystem.GetExperience(),
		"experienceToNext":  pm.rankSystem.GetExperienceToNextRank(),
		"achievementPoints": pm.achievementSystem.GetTotalPoints(),
		"achievementCount":  len(pm.achievementSystem.GetUnlockedAchievements()),
		"completionPercent": pm.achievementSystem.GetCompletionPercentage(),
		"unlockedFeatures":  len(pm.unlockedFeatures),
		"totalTrades":       pm.playerStats.GetTotalTrades(),
		"successRate":       pm.playerStats.GetSuccessRate(),
		"totalProfit":       pm.playerStats.GetTotalProfit(),
		"progressionScore":  pm.GetProgressionScore(),
	}
}

// ResetProgression resets all progression (for new game)
func (pm *ProgressionManager) ResetProgression() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Reset all systems
	pm.rankSystem = NewRankSystem()
	pm.achievementSystem.ResetAchievements()
	pm.featureUnlockSystem.ResetFeatures()
	pm.playerStats.ResetStats()

	// Reset milestones
	for _, milestone := range pm.milestones {
		milestone.Achieved = false
	}

	// Reset internal state
	pm.unlockedFeatures = make(map[string]bool)
	pm.tradeCounter = 0
}
