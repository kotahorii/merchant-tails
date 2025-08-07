package ai

import (
	"sync"
)

// TradingOutcome represents the result of a trading decision
type TradingOutcome struct {
	Decision    AIDecision
	Profit      float64
	Success     bool
	MarketState string
	Timestamp   int64
}

// AILearner manages learning from trading outcomes
type AILearner struct {
	outcomes    map[string][]*TradingOutcome   // merchantID -> outcomes
	preferences map[string]*TradingPreferences // merchantID -> preferences
	mu          sync.RWMutex
}

// NewAILearner creates a new AI learner
func NewAILearner() *AILearner {
	return &AILearner{
		outcomes:    make(map[string][]*TradingOutcome),
		preferences: make(map[string]*TradingPreferences),
	}
}

// RecordOutcome records a trading outcome for learning
func (l *AILearner) RecordOutcome(merchantID string, outcome *TradingOutcome) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Initialize if needed
	if l.outcomes[merchantID] == nil {
		l.outcomes[merchantID] = make([]*TradingOutcome, 0)
	}

	// Add outcome
	l.outcomes[merchantID] = append(l.outcomes[merchantID], outcome)

	// Keep only last 100 outcomes
	if len(l.outcomes[merchantID]) > 100 {
		l.outcomes[merchantID] = l.outcomes[merchantID][1:]
	}

	// Update preferences based on outcome
	l.updatePreferences(merchantID, outcome)
}

// updatePreferences updates trading preferences based on outcomes
func (l *AILearner) updatePreferences(merchantID string, outcome *TradingOutcome) {
	// Initialize preferences if needed
	if l.preferences[merchantID] == nil {
		l.preferences[merchantID] = &TradingPreferences{
			PreferredItems:      make(map[string]float64),
			AvoidedItems:        make(map[string]float64),
			PreferredStrategies: make(map[string]float64),
			MarketConditions:    make(map[string]float64),
		}
	}

	prefs := l.preferences[merchantID]

	// Update item preferences
	if outcome.Success {
		prefs.PreferredItems[outcome.Decision.ItemID] += 0.1
		if prefs.PreferredItems[outcome.Decision.ItemID] > 1.0 {
			prefs.PreferredItems[outcome.Decision.ItemID] = 1.0
		}
		// Remove from avoided if successful
		delete(prefs.AvoidedItems, outcome.Decision.ItemID)
	} else {
		prefs.PreferredItems[outcome.Decision.ItemID] -= 0.2
		if prefs.PreferredItems[outcome.Decision.ItemID] < -1.0 {
			// Move to avoided items
			prefs.AvoidedItems[outcome.Decision.ItemID] = -prefs.PreferredItems[outcome.Decision.ItemID]
			delete(prefs.PreferredItems, outcome.Decision.ItemID)
		}
	}

	// Update market condition preferences
	if outcome.MarketState != "" {
		if outcome.Success {
			prefs.MarketConditions[outcome.MarketState] += 0.05
		} else {
			prefs.MarketConditions[outcome.MarketState] -= 0.05
		}
	}
}

// GetPreferences gets learned preferences for a merchant
func (l *AILearner) GetPreferences(merchantID string) *TradingPreferences {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if prefs, exists := l.preferences[merchantID]; exists {
		// Return a copy to prevent external modification
		return &TradingPreferences{
			PreferredItems:      copyFloatMap(prefs.PreferredItems),
			AvoidedItems:        copyFloatMap(prefs.AvoidedItems),
			PreferredStrategies: copyFloatMap(prefs.PreferredStrategies),
			MarketConditions:    copyFloatMap(prefs.MarketConditions),
		}
	}

	// Return default preferences
	return &TradingPreferences{
		PreferredItems:      make(map[string]float64),
		AvoidedItems:        make(map[string]float64),
		PreferredStrategies: make(map[string]float64),
		MarketConditions:    make(map[string]float64),
	}
}

// GetSuccessRate calculates the success rate for a merchant
func (l *AILearner) GetSuccessRate(merchantID string) float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	outcomes := l.outcomes[merchantID]
	if len(outcomes) == 0 {
		return 0.5 // Default to 50%
	}

	successCount := 0
	for _, outcome := range outcomes {
		if outcome.Success {
			successCount++
		}
	}

	return float64(successCount) / float64(len(outcomes))
}

// GetAverageProfit calculates average profit for a merchant
func (l *AILearner) GetAverageProfit(merchantID string) float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	outcomes := l.outcomes[merchantID]
	if len(outcomes) == 0 {
		return 0
	}

	totalProfit := 0.0
	for _, outcome := range outcomes {
		totalProfit += outcome.Profit
	}

	return totalProfit / float64(len(outcomes))
}

// AnalyzePatterns identifies trading patterns from outcomes
func (l *AILearner) AnalyzePatterns(merchantID string) *TradingPattern {
	l.mu.RLock()
	defer l.mu.RUnlock()

	outcomes := l.outcomes[merchantID]
	if len(outcomes) < 10 { // Need minimum data
		return nil
	}

	pattern := &TradingPattern{
		MostProfitableItem:   "",
		MostProfitableTime:   "",
		MostSuccessfulAction: DecisionHold,
		OptimalMarketState:   "",
	}

	// Analyze item profitability
	itemProfits := make(map[string]float64)
	actionSuccess := make(map[DecisionType]int)
	actionCount := make(map[DecisionType]int)
	marketSuccess := make(map[string]int)
	marketCount := make(map[string]int)

	for _, outcome := range outcomes {
		// Track item profits
		itemProfits[outcome.Decision.ItemID] += outcome.Profit

		// Track action success
		actionCount[outcome.Decision.Type]++
		if outcome.Success {
			actionSuccess[outcome.Decision.Type]++
		}

		// Track market state success
		if outcome.MarketState != "" {
			marketCount[outcome.MarketState]++
			if outcome.Success {
				marketSuccess[outcome.MarketState]++
			}
		}
	}

	// Find most profitable item
	maxProfit := 0.0
	for item, profit := range itemProfits {
		if profit > maxProfit {
			maxProfit = profit
			pattern.MostProfitableItem = item
		}
	}

	// Find most successful action
	maxSuccessRate := 0.0
	for action, count := range actionCount {
		if count > 0 {
			successRate := float64(actionSuccess[action]) / float64(count)
			if successRate > maxSuccessRate {
				maxSuccessRate = successRate
				pattern.MostSuccessfulAction = action
			}
		}
	}

	// Find optimal market state
	maxMarketSuccess := 0.0
	for state, count := range marketCount {
		if count > 0 {
			successRate := float64(marketSuccess[state]) / float64(count)
			if successRate > maxMarketSuccess {
				maxMarketSuccess = successRate
				pattern.OptimalMarketState = state
			}
		}
	}

	return pattern
}

// TradingPattern represents identified trading patterns
type TradingPattern struct {
	MostProfitableItem   string
	MostProfitableTime   string
	MostSuccessfulAction DecisionType
	OptimalMarketState   string
}

// Helper function to copy map
func copyFloatMap(m map[string]float64) map[string]float64 {
	copy := make(map[string]float64)
	for k, v := range m {
		copy[k] = v
	}
	return copy
}
