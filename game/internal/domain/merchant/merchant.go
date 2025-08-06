package merchant

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/item"
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
)

// PersonalityType represents different merchant personalities
type PersonalityType int

const (
	PersonalityConservative PersonalityType = iota
	PersonalityBalanced
	PersonalityAggressive
	PersonalityOpportunistic
)

// TradingAction represents what action a merchant wants to take
type TradingAction int

const (
	ActionHold TradingAction = iota
	ActionBuy
	ActionSell
)

// Merchant represents an AI or player merchant
type Merchant struct {
	ID              string
	Name            string
	Personality     PersonalityType
	Gold            int
	Reputation      float64
	Inventory       *item.Inventory
	TradingStrategy TradingStrategy
	Stats           MerchantStats
	mu              sync.RWMutex
}

// MerchantStats tracks merchant performance
type MerchantStats struct {
	TotalProfit     int
	SuccessfulDeals int
	FailedDeals     int
	TotalVolume     int
}

// TradingStrategy interface for different trading behaviors
type TradingStrategy interface {
	ShouldBuy(priceRatio float64, trend market.PriceTrend) bool
	ShouldSell(priceRatio float64, trend market.PriceTrend) bool
	GetRiskTolerance() float64
}

// Trade represents a single trade transaction
type Trade struct {
	Action    TradingAction
	Item      *item.Item
	Quantity  int
	Price     int
	Timestamp time.Time
}

// TradingDecision represents a merchant's decision
type TradingDecision struct {
	Action     TradingAction
	Confidence float64
	Reasoning  string
}

// AIMerchant represents an AI-controlled merchant
type AIMerchant struct {
	*Merchant
	ActivityLevel  float64
	Specialization []item.Category
	LastAction     time.Time
}

// MarketInfluence represents how much a merchant affects the market
type MarketInfluence struct {
	TradeVolume int
	MarketSize  int
	Impact      float64
}

// MerchantNetwork represents relationships between merchants
type MerchantNetwork struct {
	Merchants     []*Merchant
	Relationships map[string]map[string]float64 // merchantID -> merchantID -> relationship strength
	mu            sync.RWMutex
}

// MarketInformation represents shared market knowledge
type MarketInformation struct {
	ItemID      string
	PriceChange float64
	Timestamp   time.Time
	Source      string
}

// NewMerchant creates a new merchant
func NewMerchant(id, name string, personality PersonalityType, startingGold int) (*Merchant, error) {
	if id == "" {
		return nil, errors.New("merchant id cannot be empty")
	}
	if name == "" {
		return nil, errors.New("merchant name cannot be empty")
	}
	if startingGold < 0 {
		return nil, errors.New("starting gold must be positive")
	}

	m := &Merchant{
		ID:          id,
		Name:        name,
		Personality: personality,
		Gold:        startingGold,
		Reputation:  0.5,
		Inventory:   item.NewInventory(),
		Stats:       MerchantStats{},
	}

	// Set trading strategy based on personality
	switch personality {
	case PersonalityAggressive:
		m.TradingStrategy = &AggressiveTradingStrategy{}
	case PersonalityConservative:
		m.TradingStrategy = &ConservativeTradingStrategy{}
	case PersonalityOpportunistic:
		m.TradingStrategy = &OpportunisticTradingStrategy{}
	default:
		m.TradingStrategy = &BalancedTradingStrategy{}
	}

	return m, nil
}

// EvaluateDeal determines if a merchant should accept a deal
func (m *Merchant) EvaluateDeal(item *item.Item, currentPrice, marketPrice int, isBuying bool) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	priceRatio := float64(currentPrice) / float64(marketPrice)

	if isBuying {
		// Buying: want price lower than market
		switch m.Personality {
		case PersonalityAggressive:
			return priceRatio <= 0.95 // Buy at 5% discount or better
		case PersonalityConservative:
			return priceRatio <= 0.75 // Need 25% discount
		case PersonalityOpportunistic:
			return priceRatio <= 0.85 // 15% discount
		default:
			return priceRatio <= 0.9 // 10% discount
		}
	} else {
		// Selling: want price higher than market
		switch m.Personality {
		case PersonalityAggressive:
			return priceRatio >= 1.05 // Sell at 5% profit or better
		case PersonalityConservative:
			return priceRatio >= 1.25 // Need 25% profit
		case PersonalityOpportunistic:
			return priceRatio >= 1.1 // 10% profit
		default:
			return priceRatio >= 1.15 // 15% profit
		}
	}
}

// MakeTradingDecision decides what action to take for an item
func (m *Merchant) MakeTradingDecision(item *item.Item, state *market.MarketState) *TradingDecision {
	decision := &TradingDecision{
		Action:     ActionHold,
		Confidence: 0.5,
		Reasoning:  "Market conditions normal",
	}

	// Simple decision based on market conditions
	if state.CurrentDemand == market.DemandHigh && state.CurrentSupply == market.SupplyLow {
		if m.Inventory.GetQuantity(item.ID) > 0 {
			decision.Action = ActionSell
			decision.Confidence = 0.8
			decision.Reasoning = "High demand, low supply - good time to sell"
		}
	} else if state.CurrentDemand == market.DemandLow && state.CurrentSupply == market.SupplyHigh {
		decision.Action = ActionBuy
		decision.Confidence = 0.7
		decision.Reasoning = "Low demand, high supply - good time to buy"
	}

	return decision
}

// CalculateRiskTolerance calculates merchant's risk tolerance
func (m *Merchant) CalculateRiskTolerance() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	baseTolerance := 0.5

	switch m.Personality {
	case PersonalityAggressive:
		baseTolerance = 0.8
	case PersonalityConservative:
		baseTolerance = 0.3
	case PersonalityOpportunistic:
		baseTolerance = 0.6
	}

	// Adjust based on current gold
	if m.Gold > 1500 {
		baseTolerance *= 1.1
	} else if m.Gold < 500 {
		baseTolerance *= 0.8
	}

	// Cap between 0 and 1
	if baseTolerance > 1.0 {
		baseTolerance = 1.0
	} else if baseTolerance < 0.1 {
		baseTolerance = 0.1
	}

	return baseTolerance
}

// ExecuteTrade executes a trading decision
func (m *Merchant) ExecuteTrade(trade *Trade) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	totalCost := trade.Quantity * trade.Price

	if trade.Action == ActionBuy {
		if m.Gold < totalCost {
			m.Stats.FailedDeals++
			return false
		}
		m.Gold -= totalCost
		m.Inventory.AddItem(trade.Item, trade.Quantity)
		m.Stats.SuccessfulDeals++
		m.Stats.TotalVolume += totalCost
	} else if trade.Action == ActionSell {
		if m.Inventory.GetQuantity(trade.Item.ID) < trade.Quantity {
			m.Stats.FailedDeals++
			return false
		}
		err := m.Inventory.RemoveItem(trade.Item.ID, trade.Quantity)
		if err != nil {
			m.Stats.FailedDeals++
			return false
		}
		m.Gold += totalCost
		m.Stats.SuccessfulDeals++
		m.Stats.TotalVolume += totalCost
		m.Stats.TotalProfit += totalCost - (trade.Item.BasePrice * trade.Quantity)
	}

	return true
}

// UpdateReputation updates merchant's reputation based on performance
func (m *Merchant) UpdateReputation() {
	m.mu.Lock()
	defer m.mu.Unlock()

	totalDeals := m.Stats.SuccessfulDeals + m.Stats.FailedDeals
	if totalDeals == 0 {
		return
	}

	successRate := float64(m.Stats.SuccessfulDeals) / float64(totalDeals)
	m.Reputation = successRate
}

// Trading Strategy Implementations

// AggressiveTradingStrategy takes more risks for higher rewards
type AggressiveTradingStrategy struct{}

func (s *AggressiveTradingStrategy) ShouldBuy(priceRatio float64, trend market.PriceTrend) bool {
	return priceRatio <= 0.95 // Buy at 5% discount
}

func (s *AggressiveTradingStrategy) ShouldSell(priceRatio float64, trend market.PriceTrend) bool {
	return priceRatio >= 1.05 // Sell at 5% profit
}

func (s *AggressiveTradingStrategy) GetRiskTolerance() float64 {
	return 0.8
}

// ConservativeTradingStrategy plays it safe
type ConservativeTradingStrategy struct{}

func (s *ConservativeTradingStrategy) ShouldBuy(priceRatio float64, trend market.PriceTrend) bool {
	return priceRatio <= 0.75 // Need 25% discount
}

func (s *ConservativeTradingStrategy) ShouldSell(priceRatio float64, trend market.PriceTrend) bool {
	return priceRatio >= 1.25 // Need 25% profit
}

func (s *ConservativeTradingStrategy) GetRiskTolerance() float64 {
	return 0.3
}

// BalancedTradingStrategy balances risk and reward
type BalancedTradingStrategy struct{}

func (s *BalancedTradingStrategy) ShouldBuy(priceRatio float64, trend market.PriceTrend) bool {
	return priceRatio <= 0.85
}

func (s *BalancedTradingStrategy) ShouldSell(priceRatio float64, trend market.PriceTrend) bool {
	return priceRatio >= 1.15
}

func (s *BalancedTradingStrategy) GetRiskTolerance() float64 {
	return 0.5
}

// OpportunisticTradingStrategy adapts to market trends
type OpportunisticTradingStrategy struct{}

func (s *OpportunisticTradingStrategy) ShouldBuy(priceRatio float64, trend market.PriceTrend) bool {
	if trend == market.TrendUp {
		return priceRatio <= 0.9 // Less strict when trend is up
	}
	return priceRatio <= 0.8
}

func (s *OpportunisticTradingStrategy) ShouldSell(priceRatio float64, trend market.PriceTrend) bool {
	if trend == market.TrendDown {
		return priceRatio >= 1.05 // Sell quicker when trend is down
	}
	return priceRatio >= 1.15
}

func (s *OpportunisticTradingStrategy) GetRiskTolerance() float64 {
	return 0.6
}

// SimulateBehavior simulates AI merchant behavior
func (ai *AIMerchant) SimulateBehavior(state *market.MarketState) []*Trade {
	trades := make([]*Trade, 0)

	// Random chance based on activity level
	if rand.Float64() > ai.ActivityLevel {
		return trades
	}

	// Focus on specialized categories
	for range ai.Specialization {
		// Simulate some trading logic
		// This is simplified - real implementation would be more complex
		action := ActionHold
		if state.CurrentDemand == market.DemandHigh {
			action = ActionSell
		} else if state.CurrentDemand == market.DemandLow {
			action = ActionBuy
		}

		if action != ActionHold {
			// Create a simulated trade
			trade := &Trade{
				Action:    action,
				Quantity:  rand.Intn(10) + 1,
				Timestamp: time.Now(),
			}
			trades = append(trades, trade)
		}
	}

	ai.LastAction = time.Now()
	return trades
}

// CalculateImpact calculates market impact of trades
func (mi *MarketInfluence) CalculateImpact() float64 {
	if mi.MarketSize == 0 {
		return 0
	}
	mi.Impact = float64(mi.TradeVolume) / float64(mi.MarketSize)
	return mi.Impact
}

// ShareInformation propagates market information through network
func (mn *MerchantNetwork) ShareInformation(info *MarketInformation) map[string]bool {
	mn.mu.RLock()
	defer mn.mu.RUnlock()

	propagation := make(map[string]bool)

	// Find source merchant's relationships
	if relationships, exists := mn.Relationships[info.Source]; exists {
		for targetID, strength := range relationships {
			// Information spreads based on relationship strength
			if rand.Float64() < strength {
				propagation[targetID] = true
			}
		}
	}

	return propagation
}

// GetStrongestRelationship finds the merchant with strongest relationship
func (mn *MerchantNetwork) GetStrongestRelationship(merchantID string) (string, float64) {
	mn.mu.RLock()
	defer mn.mu.RUnlock()

	var strongestID string
	var strongestValue float64

	if relationships, exists := mn.Relationships[merchantID]; exists {
		for id, strength := range relationships {
			if strength > strongestValue {
				strongestValue = strength
				strongestID = id
			}
		}
	}

	return strongestID, strongestValue
}

// CalculateNetworkInfluence calculates total network influence
func (mn *MerchantNetwork) CalculateNetworkInfluence() float64 {
	mn.mu.RLock()
	defer mn.mu.RUnlock()

	totalInfluence := 0.0
	connectionCount := 0

	for _, relationships := range mn.Relationships {
		for _, strength := range relationships {
			totalInfluence += strength
			connectionCount++
		}
	}

	if connectionCount == 0 {
		return 0
	}

	return totalInfluence / float64(connectionCount)
}

// GetMerchantByID finds a merchant by ID
func (mn *MerchantNetwork) GetMerchantByID(id string) *Merchant {
	mn.mu.RLock()
	defer mn.mu.RUnlock()

	for _, merchant := range mn.Merchants {
		if merchant.ID == id {
			return merchant
		}
	}
	return nil
}

// CalculateProfitMargin calculates profit margin for a trade
func CalculateProfitMargin(buyPrice, sellPrice int) float64 {
	if buyPrice == 0 {
		return 0
	}
	return float64(sellPrice-buyPrice) / float64(buyPrice)
}

// DetermineOptimalQuantity determines optimal trade quantity
func DetermineOptimalQuantity(availableGold, pricePerUnit, maxInventorySpace, currentInventory int) int {
	// Calculate based on gold constraint
	maxFromGold := availableGold / pricePerUnit

	// Calculate based on inventory constraint
	maxFromInventory := maxInventorySpace - currentInventory

	// Return the minimum to ensure both constraints are met
	optimal := int(math.Min(float64(maxFromGold), float64(maxFromInventory)))

	if optimal < 0 {
		return 0
	}
	return optimal
}
