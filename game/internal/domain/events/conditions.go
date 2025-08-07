package events

import (
	"context"
)

// Context key type for type-safe context values
type contextKey string

// Context keys
const (
	PlayerRankKey       contextKey = "playerRank"
	PlayerGoldKey       contextKey = "playerGold"
	PlayerReputationKey contextKey = "playerReputation"
	CurrentSeasonKey    contextKey = "currentSeason"
	CompletedQuestsKey  contextKey = "completedQuests"
	PlayerInventoryKey  contextKey = "playerInventory"
	RandomValueKey      contextKey = "randomValue"
	DaysPassedKey       contextKey = "daysPassed"
	ShopLevelKey        contextKey = "shopLevel"
)

// RankCondition checks if player has minimum rank
type RankCondition struct {
	MinRank string
}

// Check implements EventCondition
func (c *RankCondition) Check(ctx context.Context) bool {
	playerRank, ok := ctx.Value(PlayerRankKey).(string)
	if !ok {
		return false
	}

	rankOrder := map[string]int{
		"Apprentice": 1,
		"Journeyman": 2,
		"Expert":     3,
		"Master":     4,
	}

	playerRankValue, exists := rankOrder[playerRank]
	if !exists {
		return false
	}

	minRankValue, exists := rankOrder[c.MinRank]
	if !exists {
		return false
	}

	return playerRankValue >= minRankValue
}

// GoldCondition checks if player has minimum gold
type GoldCondition struct {
	MinGold int
}

// Check implements EventCondition
func (c *GoldCondition) Check(ctx context.Context) bool {
	playerGold, ok := ctx.Value(PlayerGoldKey).(int)
	if !ok {
		return false
	}
	return playerGold >= c.MinGold
}

// ReputationCondition checks if player has minimum reputation
type ReputationCondition struct {
	MinReputation float64
}

// Check implements EventCondition
func (c *ReputationCondition) Check(ctx context.Context) bool {
	playerReputation, ok := ctx.Value(PlayerReputationKey).(float64)
	if !ok {
		return false
	}
	return playerReputation >= c.MinReputation
}

// SeasonCondition checks if it's a specific season
type SeasonCondition struct {
	RequiredSeason int
}

// Check implements EventCondition
func (c *SeasonCondition) Check(ctx context.Context) bool {
	currentSeason, ok := ctx.Value(CurrentSeasonKey).(int)
	if !ok {
		return false
	}
	return currentSeason == c.RequiredSeason
}

// QuestCondition checks if a quest is completed
type QuestCondition struct {
	QuestID string
}

// Check implements EventCondition
func (c *QuestCondition) Check(ctx context.Context) bool {
	completedQuests, ok := ctx.Value(CompletedQuestsKey).([]string)
	if !ok {
		return false
	}

	for _, questID := range completedQuests {
		if questID == c.QuestID {
			return true
		}
	}
	return false
}

// ItemCondition checks if player has specific items
type ItemCondition struct {
	RequiredItems map[string]int // item ID -> quantity
}

// Check implements EventCondition
func (c *ItemCondition) Check(ctx context.Context) bool {
	playerInventory, ok := ctx.Value(PlayerInventoryKey).(map[string]int)
	if !ok {
		return false
	}

	for itemID, requiredQty := range c.RequiredItems {
		if playerQty, exists := playerInventory[itemID]; !exists || playerQty < requiredQty {
			return false
		}
	}
	return true
}

// RandomCondition triggers based on probability
type RandomCondition struct {
	Probability float64 // 0.0 to 1.0
}

// Check implements EventCondition
func (c *RandomCondition) Check(ctx context.Context) bool {
	// In a real implementation, this would use a random number generator
	// For testing, we can use a value from context
	randomValue, ok := ctx.Value(RandomValueKey).(float64)
	if !ok {
		// Default to 50% chance if no random value provided
		randomValue = 0.5
	}
	return randomValue < c.Probability
}

// CompoundCondition combines multiple conditions
type CompoundCondition struct {
	Conditions []EventCondition
	RequireAll bool // true = AND, false = OR
}

// Check implements EventCondition
func (c *CompoundCondition) Check(ctx context.Context) bool {
	if c.RequireAll {
		// AND logic - all conditions must be true
		for _, condition := range c.Conditions {
			if !condition.Check(ctx) {
				return false
			}
		}
		return true
	} else {
		// OR logic - at least one condition must be true
		for _, condition := range c.Conditions {
			if condition.Check(ctx) {
				return true
			}
		}
		return false
	}
}

// TimeCondition checks if enough time has passed
type TimeCondition struct {
	MinDaysPassed int
}

// Check implements EventCondition
func (c *TimeCondition) Check(ctx context.Context) bool {
	daysPassed, ok := ctx.Value(DaysPassedKey).(int)
	if !ok {
		return false
	}
	return daysPassed >= c.MinDaysPassed
}

// ShopLevelCondition checks shop level
type ShopLevelCondition struct {
	MinLevel int
}

// Check implements EventCondition
func (c *ShopLevelCondition) Check(ctx context.Context) bool {
	shopLevel, ok := ctx.Value(ShopLevelKey).(int)
	if !ok {
		return false
	}
	return shopLevel >= c.MinLevel
}
