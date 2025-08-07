package ai

import (
	"sync"
)

// AIMerchant represents an AI-controlled merchant
type AIMerchant struct {
	id           string
	name         string
	gold         int
	reputation   float64
	personality  MerchantPersonality
	inventory    *MockInventory
	tradingStats *TradingStatistics
	preferences  *TradingPreferences
	mu           sync.RWMutex
}

// NewAIMerchant creates a new AI merchant
func NewAIMerchant(id, name string, startingGold int, personality MerchantPersonality) *AIMerchant {
	return &AIMerchant{
		id:          id,
		name:        name,
		gold:        startingGold,
		reputation:  0.0,
		personality: personality,
		inventory:   NewMockInventory(100), // Starting with 100 capacity
		tradingStats: &TradingStatistics{
			TotalTrades:  0,
			TotalProfit:  0,
			BestTrade:    0,
			WorstTrade:   0,
			SuccessRate:  0.0,
			TradeHistory: make([]TradeRecord, 0),
		},
		preferences: &TradingPreferences{
			PreferredItems:      make(map[string]float64),
			AvoidedItems:        make(map[string]float64),
			PreferredStrategies: make(map[string]float64),
			MarketConditions:    make(map[string]float64),
		},
	}
}

// ID returns the merchant's unique identifier
func (m *AIMerchant) ID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.id
}

// Name returns the merchant's name
func (m *AIMerchant) Name() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.name
}

// Gold returns the merchant's current gold
func (m *AIMerchant) Gold() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.gold
}

// SetGold sets the merchant's gold amount
func (m *AIMerchant) SetGold(amount int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gold = amount
}

// AddGold adds gold to the merchant's purse
func (m *AIMerchant) AddGold(amount int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gold += amount
}

// RemoveGold removes gold from the merchant's purse
func (m *AIMerchant) RemoveGold(amount int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.gold >= amount {
		m.gold -= amount
		return true
	}
	return false
}

// Reputation returns the merchant's reputation
func (m *AIMerchant) Reputation() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.reputation
}

// SetReputation sets the merchant's reputation
func (m *AIMerchant) SetReputation(reputation float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Clamp reputation between -100 and 100
	if reputation > 100 {
		reputation = 100
	} else if reputation < -100 {
		reputation = -100
	}
	m.reputation = reputation
}

// AdjustReputation adjusts the merchant's reputation by delta
func (m *AIMerchant) AdjustReputation(delta float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reputation += delta
	// Clamp reputation between -100 and 100
	if m.reputation > 100 {
		m.reputation = 100
	} else if m.reputation < -100 {
		m.reputation = -100
	}
}

// Personality returns the merchant's personality
func (m *AIMerchant) Personality() MerchantPersonality {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.personality
}

// Inventory returns the merchant's inventory
func (m *AIMerchant) Inventory() *MockInventory {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.inventory
}

// AddItem adds an item to the merchant's inventory
func (m *AIMerchant) AddItem(itemID string, quantity int, purchasePrice float64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// For simplicity, we'll just track quantity for now
	// In a real implementation, this would interact with the inventory system
	return true
}

// RemoveItem removes an item from the merchant's inventory
func (m *AIMerchant) RemoveItem(itemID string, quantity int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// For simplicity, we'll just track quantity for now
	// In a real implementation, this would interact with the inventory system
	return true
}

// GetTradingStats returns the merchant's trading statistics
func (m *AIMerchant) GetTradingStats() *TradingStatistics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tradingStats
}

// RecordTrade records a completed trade
func (m *AIMerchant) RecordTrade(record TradeRecord) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tradingStats.TotalTrades++
	m.tradingStats.TotalProfit += record.Profit

	if record.Profit > m.tradingStats.BestTrade {
		m.tradingStats.BestTrade = record.Profit
	}
	if record.Profit < m.tradingStats.WorstTrade {
		m.tradingStats.WorstTrade = record.Profit
	}

	// Calculate success rate
	if record.Profit > 0 {
		successCount := int(m.tradingStats.SuccessRate * float64(m.tradingStats.TotalTrades-1))
		successCount++
		m.tradingStats.SuccessRate = float64(successCount) / float64(m.tradingStats.TotalTrades)
	} else {
		successCount := int(m.tradingStats.SuccessRate * float64(m.tradingStats.TotalTrades-1))
		m.tradingStats.SuccessRate = float64(successCount) / float64(m.tradingStats.TotalTrades)
	}

	// Keep last 100 trades in history
	m.tradingStats.TradeHistory = append(m.tradingStats.TradeHistory, record)
	if len(m.tradingStats.TradeHistory) > 100 {
		m.tradingStats.TradeHistory = m.tradingStats.TradeHistory[1:]
	}
}

// GetPreferences returns the merchant's trading preferences
func (m *AIMerchant) GetPreferences() *TradingPreferences {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.preferences
}

// UpdatePreferences updates the merchant's trading preferences based on outcomes
func (m *AIMerchant) UpdatePreferences(itemID string, profit float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Adjust item preference based on profit
	currentPref := m.preferences.PreferredItems[itemID]
	if profit > 0 {
		// Increase preference for profitable items
		m.preferences.PreferredItems[itemID] = currentPref + 0.1
	} else {
		// Decrease preference for unprofitable items
		m.preferences.PreferredItems[itemID] = currentPref - 0.1
		if m.preferences.PreferredItems[itemID] < -1.0 {
			// Move to avoided items if very unprofitable
			m.preferences.AvoidedItems[itemID] = -m.preferences.PreferredItems[itemID]
			delete(m.preferences.PreferredItems, itemID)
		}
	}
}

// TradingStatistics tracks merchant's trading performance
type TradingStatistics struct {
	TotalTrades  int
	TotalProfit  float64
	BestTrade    float64
	WorstTrade   float64
	SuccessRate  float64
	TradeHistory []TradeRecord
}

// TradeRecord represents a single trade
type TradeRecord struct {
	ItemID    string
	Quantity  int
	BuyPrice  float64
	SellPrice float64
	Profit    float64
	Timestamp int64
}

// TradingPreferences represents merchant's learned preferences
type TradingPreferences struct {
	PreferredItems      map[string]float64 // item ID -> preference score
	AvoidedItems        map[string]float64 // item ID -> avoidance score
	PreferredStrategies map[string]float64 // strategy name -> preference score
	MarketConditions    map[string]float64 // condition -> preference score
}

// GetItemPreference returns the preference score for an item
func (p *TradingPreferences) GetItemPreference(itemID string) float64 {
	if pref, exists := p.PreferredItems[itemID]; exists {
		return pref
	}
	if avoid, exists := p.AvoidedItems[itemID]; exists {
		return -avoid
	}
	return 0.0 // Neutral preference
}

// GetStrategyForMarket returns the preferred strategy for market conditions
func (p *TradingPreferences) GetStrategyForMarket(marketState string) string {
	// If no strategies are set, look at market conditions
	if len(p.PreferredStrategies) == 0 {
		// Use market condition preference
		switch marketState {
		case "volatile":
			return "momentum"
		case "stable":
			return "value"
		default:
			return "balanced"
		}
	}

	// Return the strategy with highest preference
	bestStrategy := "balanced"
	bestScore := 0.0

	for strategy, score := range p.PreferredStrategies {
		if score > bestScore {
			bestStrategy = strategy
			bestScore = score
		}
	}

	return bestStrategy
}
