package ai

import (
	"context"
	"sync"
)

// AIMerchantSystem manages all AI merchants in the game
type AIMerchantSystem struct {
	merchants map[string]*AIMerchant
	network   *MerchantNetwork
	learner   *AILearner
	behavior  AIBehavior
	mu        sync.RWMutex
}

// NewAIMerchantSystem creates a new AI merchant system
func NewAIMerchantSystem() *AIMerchantSystem {
	return &AIMerchantSystem{
		merchants: make(map[string]*AIMerchant),
		network:   NewMerchantNetwork(),
		learner:   NewAILearner(),
		behavior:  NewStandardAIBehavior(),
	}
}

// AddMerchant adds a merchant to the system
func (s *AIMerchantSystem) AddMerchant(merchant *AIMerchant) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.merchants[merchant.ID()] = merchant
	s.network.AddMerchant(merchant)
}

// RemoveMerchant removes a merchant from the system
func (s *AIMerchantSystem) RemoveMerchant(merchantID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.merchants, merchantID)
	s.network.RemoveMerchant(merchantID)
}

// GetMerchant retrieves a merchant by ID
func (s *AIMerchantSystem) GetMerchant(merchantID string) *AIMerchant {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.merchants[merchantID]
}

// GetAllMerchants returns all merchants in the system
func (s *AIMerchantSystem) GetAllMerchants() []*AIMerchant {
	s.mu.RLock()
	defer s.mu.RUnlock()

	merchants := make([]*AIMerchant, 0, len(s.merchants))
	for _, merchant := range s.merchants {
		merchants = append(merchants, merchant)
	}

	return merchants
}

// TradingRoundResult represents the result of a trading round for a merchant
type TradingRoundResult struct {
	MerchantID string
	Decision   *AIDecision
	Outcome    *TradingOutcome
}

// SimulateTradingRound simulates a trading round for all merchants
func (s *AIMerchantSystem) SimulateTradingRound(ctx context.Context, marketData interface{}) map[string]*TradingRoundResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make(map[string]*TradingRoundResult)

	// For now, create mock decisions since we don't have real market data
	for merchantID, merchant := range s.merchants {
		// Create a mock decision based on personality
		decision := &AIDecision{
			Type:       DecisionHold,
			Confidence: 0.5,
			Reason:     "Market analysis",
		}

		// Personality-based decision type
		personality := merchant.Personality()
		if personality.RiskTolerance() > 0.6 {
			decision.Type = DecisionBuy
			decision.ItemID = "volatile_item"
			decision.Quantity = 5
			decision.Price = 100
			decision.Confidence = personality.RiskTolerance()
		} else if personality.RiskTolerance() < 0.4 {
			decision.Type = DecisionBuy
			decision.ItemID = "stable_item"
			decision.Quantity = 3
			decision.Price = 50
			decision.Confidence = 1.0 - personality.RiskTolerance()
		}

		// Create outcome
		outcome := &TradingOutcome{
			Decision:    *decision,
			Profit:      0,
			Success:     decision.Confidence > 0.5,
			MarketState: "normal",
		}

		// Record outcome for learning
		s.learner.RecordOutcome(merchantID, outcome)

		results[merchantID] = &TradingRoundResult{
			MerchantID: merchantID,
			Decision:   decision,
			Outcome:    outcome,
		}
	}

	return results
}

// UpdateNetwork updates merchant relationships based on interactions
func (s *AIMerchantSystem) UpdateNetwork(merchantA, merchantB string, interaction float64) {
	s.network.UpdateRelationshipStrength(merchantA, merchantB, interaction)
}

// GetMarketInfluence calculates total market influence of all merchants
func (s *AIMerchantSystem) GetMarketInfluence() *MarketInfluence {
	s.mu.RLock()
	defer s.mu.RUnlock()

	calculator := NewMarketInfluenceCalculator()
	influences := make([]*MarketInfluence, 0, len(s.merchants))

	for _, merchant := range s.merchants {
		influence := calculator.CalculateInfluence(merchant)
		influences = append(influences, influence)
	}

	return calculator.AggregateInfluence(influences)
}
