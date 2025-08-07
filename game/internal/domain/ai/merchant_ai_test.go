package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMerchantPersonality(t *testing.T) {
	t.Run("aggressive personality", func(t *testing.T) {
		personality := NewAggressivePersonality()
		assert.Equal(t, PersonalityAggressive, personality.Type())
		assert.Equal(t, 0.8, personality.RiskTolerance())
		assert.Equal(t, 1.5, personality.TradingFrequency())
		assert.Equal(t, 0.3, personality.ProfitMarginTarget())
		assert.Equal(t, 1.2, personality.CompetitivenessFactor())
	})

	t.Run("conservative personality", func(t *testing.T) {
		personality := NewConservativePersonality()
		assert.Equal(t, PersonalityConservative, personality.Type())
		assert.Equal(t, 0.2, personality.RiskTolerance())
		assert.Equal(t, 0.7, personality.TradingFrequency())
		assert.Equal(t, 0.5, personality.ProfitMarginTarget())
		assert.Equal(t, 0.8, personality.CompetitivenessFactor())
	})

	t.Run("balanced personality", func(t *testing.T) {
		personality := NewBalancedPersonality()
		assert.Equal(t, PersonalityBalanced, personality.Type())
		assert.Equal(t, 0.5, personality.RiskTolerance())
		assert.Equal(t, 1.0, personality.TradingFrequency())
		assert.Equal(t, 0.4, personality.ProfitMarginTarget())
		assert.Equal(t, 1.0, personality.CompetitivenessFactor())
	})

	t.Run("opportunistic personality", func(t *testing.T) {
		personality := NewOpportunisticPersonality()
		assert.Equal(t, PersonalityOpportunistic, personality.Type())
		assert.Equal(t, 0.6, personality.RiskTolerance())
		assert.Equal(t, 1.3, personality.TradingFrequency())
		assert.Equal(t, 0.35, personality.ProfitMarginTarget())
		assert.Equal(t, 1.1, personality.CompetitivenessFactor())
	})
}

func TestAIMerchant(t *testing.T) {
	merchant := NewAIMerchant(
		"merchant_001",
		"John the Trader",
		1000,
		NewBalancedPersonality(),
	)

	assert.NotNil(t, merchant)
	assert.Equal(t, "merchant_001", merchant.ID())
	assert.Equal(t, "John the Trader", merchant.Name())
	assert.Equal(t, 1000, merchant.Gold())
	assert.Equal(t, PersonalityBalanced, merchant.Personality().Type())
	assert.Equal(t, 0.0, merchant.Reputation())
	assert.NotNil(t, merchant.Inventory())
}

func TestAIBehavior(t *testing.T) {
	behavior := NewStandardAIBehavior()
	assert.NotNil(t, behavior)

	merchant := NewAIMerchant(
		"merchant_001",
		"Test Merchant",
		1000,
		NewAggressivePersonality(),
	)

	// Create mock market data
	marketData := &MarketData{
		Items: []ItemData{
			{
				ItemID:       "sword_001",
				CurrentPrice: 100,
				BasePrice:    90,
				Supply:       50,
				Demand:       60,
				Volatility:   0.2,
			},
			{
				ItemID:       "potion_001",
				CurrentPrice: 20,
				BasePrice:    25,
				Supply:       100,
				Demand:       80,
				Volatility:   0.1,
			},
		},
	}

	ctx := context.Background()
	decision := behavior.MakeDecision(ctx, merchant, marketData)

	assert.NotNil(t, decision)
	assert.Contains(t, []DecisionType{DecisionBuy, DecisionSell, DecisionHold}, decision.Type)
}

func TestTradingStrategy(t *testing.T) {
	t.Run("value trading strategy", func(t *testing.T) {
		strategy := NewValueTradingStrategy()
		merchant := NewAIMerchant("m1", "Value Trader", 1000, NewConservativePersonality())

		marketData := &MarketData{
			Items: []ItemData{
				{
					ItemID:       "undervalued_item",
					CurrentPrice: 80,
					BasePrice:    100,
					Supply:       30,
					Demand:       50,
				},
				{
					ItemID:       "overvalued_item",
					CurrentPrice: 120,
					BasePrice:    100,
					Supply:       50,
					Demand:       30,
				},
			},
		}

		decision := strategy.Evaluate(context.Background(), merchant, marketData)
		assert.NotNil(t, decision)

		// Should buy undervalued items
		if decision.Type == DecisionBuy {
			assert.Equal(t, "undervalued_item", decision.ItemID)
		}
	})

	t.Run("momentum trading strategy", func(t *testing.T) {
		strategy := NewMomentumTradingStrategy()
		merchant := NewAIMerchant("m2", "Momentum Trader", 1000, NewAggressivePersonality())

		marketData := &MarketData{
			Items: []ItemData{
				{
					ItemID:       "rising_item",
					CurrentPrice: 110,
					BasePrice:    100,
					PriceHistory: []float64{95, 100, 105, 110}, // Rising trend
				},
				{
					ItemID:       "falling_item",
					CurrentPrice: 90,
					BasePrice:    100,
					PriceHistory: []float64{105, 100, 95, 90}, // Falling trend
				},
			},
		}

		decision := strategy.Evaluate(context.Background(), merchant, marketData)
		assert.NotNil(t, decision)

		// Should buy items with upward momentum
		if decision.Type == DecisionBuy {
			assert.Equal(t, "rising_item", decision.ItemID)
		}
	})

	t.Run("seasonal trading strategy", func(t *testing.T) {
		strategy := NewSeasonalTradingStrategy()
		merchant := NewAIMerchant("m3", "Seasonal Trader", 1000, NewBalancedPersonality())

		// Mock seasonal context
		ctx := context.WithValue(context.Background(), SeasonKey, "summer")

		marketData := &MarketData{
			Items: []ItemData{
				{
					ItemID:       "ice_potion",
					CurrentPrice: 50,
					Category:     1, // CategoryPotion
					Tags:         []string{"summer", "cooling"},
				},
				{
					ItemID:       "fire_sword",
					CurrentPrice: 100,
					Category:     2, // CategoryWeapon
					Tags:         []string{"winter", "heating"},
				},
			},
		}

		decision := strategy.Evaluate(ctx, merchant, marketData)
		assert.NotNil(t, decision)

		// Should prefer summer items in summer
		if decision.Type == DecisionBuy {
			assert.Equal(t, "ice_potion", decision.ItemID)
		}
	})
}

func TestMarketInfluence(t *testing.T) {
	calculator := NewMarketInfluenceCalculator()

	merchant := NewAIMerchant("m1", "Influential Merchant", 5000, NewAggressivePersonality())
	merchant.SetReputation(80)

	// Calculate influence on market
	influence := calculator.CalculateInfluence(merchant)
	assert.Greater(t, influence.PriceImpact, 0.0)
	assert.Greater(t, influence.DemandImpact, 0.0)
	assert.Greater(t, influence.SupplyImpact, 0.0)

	// Higher gold and reputation should mean more influence
	poorMerchant := NewAIMerchant("m2", "Poor Merchant", 100, NewConservativePersonality())
	poorMerchant.SetReputation(10)

	poorInfluence := calculator.CalculateInfluence(poorMerchant)
	assert.Less(t, poorInfluence.PriceImpact, influence.PriceImpact)
}

func TestAIMerchantNetwork(t *testing.T) {
	network := NewMerchantNetwork()

	// Add merchants
	m1 := NewAIMerchant("m1", "Merchant 1", 1000, NewAggressivePersonality())
	m2 := NewAIMerchant("m2", "Merchant 2", 1500, NewBalancedPersonality())
	m3 := NewAIMerchant("m3", "Merchant 3", 800, NewConservativePersonality())

	network.AddMerchant(m1)
	network.AddMerchant(m2)
	network.AddMerchant(m3)

	// Create relationships
	network.AddRelationship(m1.ID(), m2.ID(), RelationshipFriendly)
	network.AddRelationship(m1.ID(), m3.ID(), RelationshipRival)
	network.AddRelationship(m2.ID(), m3.ID(), RelationshipNeutral)

	// Test information propagation
	info := &MarketInformation{
		ItemID:      "sword_001",
		PriceChange: 0.2,
		Source:      m1.ID(),
		Reliability: 1.0, // Initial reliability
	}

	propagated := network.PropagateInformation(info)
	assert.Greater(t, len(propagated), 0)

	// Friends should get information faster than rivals
	for _, target := range propagated {
		if target.TargetID == m2.ID() {
			assert.GreaterOrEqual(t, target.Reliability, 0.79) // Allow for floating point precision
		}
		if target.TargetID == m3.ID() {
			assert.Less(t, target.Reliability, 0.5)
		}
	}
}

func TestAIDecisionMaking(t *testing.T) {
	merchant := NewAIMerchant("m1", "Smart Merchant", 2000, NewOpportunisticPersonality())
	decisionMaker := NewDecisionMaker()

	// Add some items to inventory
	merchant.AddItem("sword_001", 5, 100)
	merchant.AddItem("potion_001", 20, 15)

	marketData := &MarketData{
		Items: []ItemData{
			{
				ItemID:       "sword_001",
				CurrentPrice: 150, // 50% profit opportunity
				Supply:       30,
				Demand:       50,
			},
			{
				ItemID:       "potion_001",
				CurrentPrice: 12, // Loss if sold
				Supply:       100,
				Demand:       60,
			},
			{
				ItemID:       "gem_001",
				CurrentPrice: 500,
				BasePrice:    450,
				Supply:       5,
				Demand:       10,
			},
		},
	}

	ctx := context.Background()
	decisions := decisionMaker.MakeDecisions(ctx, merchant, marketData, 3)

	require.NotNil(t, decisions)
	assert.LessOrEqual(t, len(decisions), 3)

	// Should prioritize selling swords (high profit)
	foundSellSword := false
	for _, decision := range decisions {
		if decision.Type == DecisionSell && decision.ItemID == "sword_001" {
			foundSellSword = true
			break
		}
	}
	assert.True(t, foundSellSword, "Should decide to sell swords for profit")
}

func TestAILearning(t *testing.T) {
	learner := NewAILearner()
	merchant := NewAIMerchant("m1", "Learning Merchant", 1000, NewBalancedPersonality())

	// Record trading outcomes
	outcome1 := &TradingOutcome{
		Decision: AIDecision{
			Type:     DecisionBuy,
			ItemID:   "sword_001",
			Quantity: 5,
			Price:    100,
		},
		Profit:      50,
		Success:     true,
		MarketState: "normal",
	}

	outcome2 := &TradingOutcome{
		Decision: AIDecision{
			Type:     DecisionBuy,
			ItemID:   "potion_001",
			Quantity: 10,
			Price:    20,
		},
		Profit:      -30,
		Success:     false,
		MarketState: "volatile",
	}

	learner.RecordOutcome(merchant.ID(), outcome1)
	learner.RecordOutcome(merchant.ID(), outcome2)

	// Get learned preferences
	preferences := learner.GetPreferences(merchant.ID())
	assert.NotNil(t, preferences)

	// Should prefer swords over potions based on outcomes
	swordPref := preferences.GetItemPreference("sword_001")
	potionPref := preferences.GetItemPreference("potion_001")
	assert.Greater(t, swordPref, potionPref)

	// Should adapt strategy based on market conditions
	normalStrategy := preferences.GetStrategyForMarket("normal")
	volatileStrategy := preferences.GetStrategyForMarket("volatile")
	assert.NotEqual(t, normalStrategy, volatileStrategy)
}

func TestAIMerchantIntegration(t *testing.T) {
	// Create a full AI merchant system
	system := NewAIMerchantSystem()

	// Add multiple merchants with different personalities
	aggressive := NewAIMerchant("aggressive", "Aggressive Al", 2000, NewAggressivePersonality())
	conservative := NewAIMerchant("conservative", "Conservative Carl", 2000, NewConservativePersonality())
	balanced := NewAIMerchant("balanced", "Balanced Bob", 2000, NewBalancedPersonality())

	system.AddMerchant(aggressive)
	system.AddMerchant(conservative)
	system.AddMerchant(balanced)

	// Create market conditions
	marketData := &MarketData{
		Items: []ItemData{
			{
				ItemID:       "volatile_item",
				CurrentPrice: 100,
				BasePrice:    100,
				Volatility:   0.5,
				Supply:       50,
				Demand:       50,
			},
			{
				ItemID:       "stable_item",
				CurrentPrice: 50,
				BasePrice:    50,
				Volatility:   0.1,
				Supply:       100,
				Demand:       100,
			},
		},
	}

	// Simulate a trading round
	ctx := context.Background()
	results := system.SimulateTradingRound(ctx, marketData)

	assert.NotNil(t, results)
	assert.Equal(t, 3, len(results))

	// Check that different personalities made different decisions
	aggressiveResult := results["aggressive"]
	conservativeResult := results["conservative"]

	// Aggressive should take more risks
	if aggressiveResult.Decision.Type == DecisionBuy {
		assert.Equal(t, "volatile_item", aggressiveResult.Decision.ItemID)
	}

	// Conservative should prefer stable items
	if conservativeResult.Decision.Type == DecisionBuy {
		assert.Equal(t, "stable_item", conservativeResult.Decision.ItemID)
	}
}
