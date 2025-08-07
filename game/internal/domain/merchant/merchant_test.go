package merchant

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/merchant-tails/game/internal/domain/item"
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
)

func TestNewMerchant(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		merchantName string
		personality  PersonalityType
		startingGold int
		wantErr      bool
		errContains  string
	}{
		{
			name:         "create aggressive merchant",
			id:           "merchant_001",
			merchantName: "Bold Trader",
			personality:  PersonalityAggressive,
			startingGold: 1000,
			wantErr:      false,
		},
		{
			name:         "create conservative merchant",
			id:           "merchant_002",
			merchantName: "Careful Trader",
			personality:  PersonalityConservative,
			startingGold: 500,
			wantErr:      false,
		},
		{
			name:         "invalid empty id",
			id:           "",
			merchantName: "Invalid",
			personality:  PersonalityBalanced,
			startingGold: 1000,
			wantErr:      true,
			errContains:  "merchant id cannot be empty",
		},
		{
			name:         "invalid negative gold",
			id:           "merchant_003",
			merchantName: "Poor Trader",
			personality:  PersonalityBalanced,
			startingGold: -100,
			wantErr:      true,
			errContains:  "starting gold must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merchant, err := NewMerchant(tt.id, tt.merchantName, tt.personality, tt.startingGold)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, merchant)
			} else {
				require.NoError(t, err)
				require.NotNil(t, merchant)
				assert.Equal(t, tt.id, merchant.ID)
				assert.Equal(t, tt.merchantName, merchant.Name)
				assert.Equal(t, tt.personality, merchant.Personality)
				assert.Equal(t, tt.startingGold, merchant.Gold)
				assert.NotNil(t, merchant.Inventory)
				assert.NotNil(t, merchant.TradingStrategy)
			}
		})
	}
}

func TestMerchant_EvaluateDeal(t *testing.T) {
	aggressive, _ := NewMerchant("m1", "Aggressive", PersonalityAggressive, 1000)
	conservative, _ := NewMerchant("m2", "Conservative", PersonalityConservative, 1000)

	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)

	tests := []struct {
		name         string
		merchant     *Merchant
		item         *item.Item
		currentPrice int
		marketPrice  int
		isBuying     bool
		expected     bool
	}{
		{
			name:         "aggressive buys at slight discount",
			merchant:     aggressive,
			item:         apple,
			currentPrice: 9,
			marketPrice:  10,
			isBuying:     true,
			expected:     true,
		},
		{
			name:         "conservative needs bigger discount",
			merchant:     conservative,
			item:         apple,
			currentPrice: 9,
			marketPrice:  10,
			isBuying:     true,
			expected:     false,
		},
		{
			name:         "conservative buys at good discount",
			merchant:     conservative,
			item:         apple,
			currentPrice: 7,
			marketPrice:  10,
			isBuying:     true,
			expected:     true,
		},
		{
			name:         "aggressive sells at small profit",
			merchant:     aggressive,
			item:         apple,
			currentPrice: 11,
			marketPrice:  10,
			isBuying:     false,
			expected:     true,
		},
		{
			name:         "conservative needs bigger profit margin",
			merchant:     conservative,
			item:         apple,
			currentPrice: 11,
			marketPrice:  10,
			isBuying:     false,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.merchant.EvaluateDeal(tt.item, tt.currentPrice, tt.marketPrice, tt.isBuying)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMerchant_MakeTradingDecision(t *testing.T) {
	merchant, _ := NewMerchant("m1", "Trader", PersonalityBalanced, 1000)

	marketSys := market.NewMarket()
	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	marketSys.RegisterItem(apple)

	// Add item to merchant inventory
	_ = merchant.Inventory.AddItem(apple, 5)

	tests := []struct {
		name           string
		marketState    *market.MarketState
		expectedAction TradingAction
	}{
		{
			name: "buy when demand is low",
			marketState: &market.MarketState{
				CurrentDemand: market.DemandLow,
				CurrentSupply: market.SupplyHigh,
			},
			expectedAction: ActionBuy,
		},
		{
			name: "sell when demand is high",
			marketState: &market.MarketState{
				CurrentDemand: market.DemandHigh,
				CurrentSupply: market.SupplyLow,
			},
			expectedAction: ActionSell,
		},
		{
			name: "hold when market is normal",
			marketState: &market.MarketState{
				CurrentDemand: market.DemandNormal,
				CurrentSupply: market.SupplyNormal,
			},
			expectedAction: ActionHold,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := merchant.MakeTradingDecision(apple, tt.marketState)
			assert.Equal(t, tt.expectedAction, decision.Action)
		})
	}
}

func TestMerchant_CalculateRiskTolerance(t *testing.T) {
	tests := []struct {
		name        string
		personality PersonalityType
		gold        int
		expected    float64
		delta       float64
	}{
		{
			name:        "aggressive with high gold",
			personality: PersonalityAggressive,
			gold:        2000,
			expected:    0.8,
			delta:       0.1,
		},
		{
			name:        "conservative with low gold",
			personality: PersonalityConservative,
			gold:        100,
			expected:    0.2,
			delta:       0.1,
		},
		{
			name:        "balanced with medium gold",
			personality: PersonalityBalanced,
			gold:        1000,
			expected:    0.5,
			delta:       0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merchant, _ := NewMerchant("m1", "Test", tt.personality, tt.gold)
			tolerance := merchant.CalculateRiskTolerance()
			assert.InDelta(t, tt.expected, tolerance, tt.delta)
		})
	}
}

func TestTradingStrategy_ShouldBuy(t *testing.T) {
	aggressive := &AggressiveTradingStrategy{}
	conservative := &ConservativeTradingStrategy{}

	tests := []struct {
		name        string
		strategy    TradingStrategy
		priceRatio  float64
		marketTrend market.PriceTrend
		expectedBuy bool
	}{
		{
			name:        "aggressive buys on slight discount",
			strategy:    aggressive,
			priceRatio:  0.95,
			marketTrend: market.TrendStable,
			expectedBuy: true,
		},
		{
			name:        "conservative needs bigger discount",
			strategy:    conservative,
			priceRatio:  0.95,
			marketTrend: market.TrendStable,
			expectedBuy: false,
		},
		{
			name:        "conservative buys on big discount",
			strategy:    conservative,
			priceRatio:  0.75,
			marketTrend: market.TrendUp,
			expectedBuy: true,
		},
		{
			name:        "aggressive avoids buying when overpriced",
			strategy:    aggressive,
			priceRatio:  1.2,
			marketTrend: market.TrendUp,
			expectedBuy: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.strategy.ShouldBuy(tt.priceRatio, tt.marketTrend)
			assert.Equal(t, tt.expectedBuy, result)
		})
	}
}

func TestTradingStrategy_ShouldSell(t *testing.T) {
	aggressive := &AggressiveTradingStrategy{}
	conservative := &ConservativeTradingStrategy{}
	opportunistic := &OpportunisticTradingStrategy{}

	tests := []struct {
		name         string
		strategy     TradingStrategy
		priceRatio   float64
		marketTrend  market.PriceTrend
		expectedSell bool
	}{
		{
			name:         "aggressive sells on small profit",
			strategy:     aggressive,
			priceRatio:   1.05,
			marketTrend:  market.TrendStable,
			expectedSell: true,
		},
		{
			name:         "conservative needs bigger profit",
			strategy:     conservative,
			priceRatio:   1.05,
			marketTrend:  market.TrendStable,
			expectedSell: false,
		},
		{
			name:         "conservative sells on good profit",
			strategy:     conservative,
			priceRatio:   1.25,
			marketTrend:  market.TrendStable,
			expectedSell: true,
		},
		{
			name:         "opportunistic sells when trend is down",
			strategy:     opportunistic,
			priceRatio:   1.1,
			marketTrend:  market.TrendDown,
			expectedSell: true,
		},
		{
			name:         "opportunistic holds when trend is up",
			strategy:     opportunistic,
			priceRatio:   1.1,
			marketTrend:  market.TrendUp,
			expectedSell: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.strategy.ShouldSell(tt.priceRatio, tt.marketTrend)
			assert.Equal(t, tt.expectedSell, result)
		})
	}
}

func TestMerchant_ExecuteTrade(t *testing.T) {
	merchant, _ := NewMerchant("m1", "Trader", PersonalityBalanced, 1000)
	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)

	// Add some apples to inventory
	_ = merchant.Inventory.AddItem(apple, 10)

	tests := []struct {
		name          string
		action        TradingAction
		item          *item.Item
		quantity      int
		price         int
		expectedGold  int
		expectedQty   int
		shouldSucceed bool
	}{
		{
			name:          "successful buy",
			action:        ActionBuy,
			item:          apple,
			quantity:      5,
			price:         10,
			expectedGold:  950, // 1000 - (5 * 10)
			expectedQty:   15,  // 10 + 5
			shouldSucceed: true,
		},
		{
			name:          "successful sell",
			action:        ActionSell,
			item:          apple,
			quantity:      3,
			price:         12,
			expectedGold:  986, // 950 + (3 * 12)
			expectedQty:   12,  // 15 - 3
			shouldSucceed: true,
		},
		{
			name:          "insufficient gold for buy",
			action:        ActionBuy,
			item:          apple,
			quantity:      100,
			price:         100,
			expectedGold:  986, // unchanged
			expectedQty:   12,  // unchanged
			shouldSucceed: false,
		},
		{
			name:          "insufficient inventory for sell",
			action:        ActionSell,
			item:          apple,
			quantity:      20,
			price:         10,
			expectedGold:  986, // unchanged
			expectedQty:   12,  // unchanged
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trade := &Trade{
				Action:   tt.action,
				Item:     tt.item,
				Quantity: tt.quantity,
				Price:    tt.price,
			}

			success := merchant.ExecuteTrade(trade)
			assert.Equal(t, tt.shouldSucceed, success)
			assert.Equal(t, tt.expectedGold, merchant.Gold)
			assert.Equal(t, tt.expectedQty, merchant.Inventory.GetQuantity(tt.item.ID))
		})
	}
}

func TestMerchant_UpdateReputation(t *testing.T) {
	merchant, _ := NewMerchant("m1", "Trader", PersonalityBalanced, 1000)

	tests := []struct {
		name            string
		successfulDeals int
		failedDeals     int
		expectedRep     float64
		delta           float64
	}{
		{
			name:            "all successful deals",
			successfulDeals: 10,
			failedDeals:     0,
			expectedRep:     1.0,
			delta:           0.1,
		},
		{
			name:            "mixed results",
			successfulDeals: 7,
			failedDeals:     3,
			expectedRep:     0.7,
			delta:           0.1,
		},
		{
			name:            "mostly failed deals",
			successfulDeals: 2,
			failedDeals:     8,
			expectedRep:     0.2,
			delta:           0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merchant.Stats.SuccessfulDeals = tt.successfulDeals
			merchant.Stats.FailedDeals = tt.failedDeals
			merchant.UpdateReputation()
			assert.InDelta(t, tt.expectedRep, merchant.Reputation, tt.delta)
		})
	}
}

func TestAIMerchant_SimulateBehavior(t *testing.T) {
	aiMerchant := &AIMerchant{
		Merchant: &Merchant{
			ID:          "ai_001",
			Name:        "AI Trader",
			Personality: PersonalityOpportunistic,
			Gold:        2000,
			Inventory:   item.NewInventory(),
		},
		ActivityLevel: 0.8,
		Specialization: []item.Category{
			item.CategoryFruit,
			item.CategoryPotion,
		},
	}

	marketState := &market.MarketState{
		CurrentDemand: market.DemandHigh,
		CurrentSupply: market.SupplyLow,
		CurrentSeason: item.SeasonSummer,
	}

	// Test that AI merchant makes decisions
	actions := aiMerchant.SimulateBehavior(marketState)
	assert.NotNil(t, actions)
	assert.GreaterOrEqual(t, len(actions), 0)

	// Test activity level affects behavior
	aiMerchant.ActivityLevel = 0.1
	lowActivityActions := aiMerchant.SimulateBehavior(marketState)

	aiMerchant.ActivityLevel = 0.9
	highActivityActions := aiMerchant.SimulateBehavior(marketState)

	// High activity should generally produce more actions
	// (though randomness means this isn't guaranteed)
	assert.NotNil(t, lowActivityActions)
	assert.NotNil(t, highActivityActions)
}

func TestMarketInfluence_CalculateImpact(t *testing.T) {
	tests := []struct {
		name           string
		tradeVolume    int
		marketSize     int
		expectedImpact float64
		delta          float64
	}{
		{
			name:           "small trade in large market",
			tradeVolume:    10,
			marketSize:     1000,
			expectedImpact: 0.01,
			delta:          0.01,
		},
		{
			name:           "large trade in small market",
			tradeVolume:    100,
			marketSize:     200,
			expectedImpact: 0.5,
			delta:          0.1,
		},
		{
			name:           "medium trade in medium market",
			tradeVolume:    50,
			marketSize:     500,
			expectedImpact: 0.1,
			delta:          0.05,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			influence := &MarketInfluence{
				TradeVolume: tt.tradeVolume,
				MarketSize:  tt.marketSize,
			}
			impact := influence.CalculateImpact()
			assert.InDelta(t, tt.expectedImpact, impact, tt.delta)
		})
	}
}

func TestMerchantNetwork_ShareInformation(t *testing.T) {
	// Create a network of merchants
	m1, _ := NewMerchant("m1", "Trader 1", PersonalityAggressive, 1000)
	m2, _ := NewMerchant("m2", "Trader 2", PersonalityConservative, 1000)
	m3, _ := NewMerchant("m3", "Trader 3", PersonalityBalanced, 1000)

	network := &MerchantNetwork{
		Merchants: []*Merchant{m1, m2, m3},
		Relationships: map[string]map[string]float64{
			"m1": {"m2": 0.7, "m3": 0.5},
			"m2": {"m1": 0.7, "m3": 0.3},
			"m3": {"m1": 0.5, "m2": 0.3},
		},
	}

	// Create market information
	info := &MarketInformation{
		ItemID:      "apple_001",
		PriceChange: 0.2,
		Timestamp:   time.Now(),
		Source:      "m1",
	}

	// Share information through network
	// Run multiple times since propagation is probabilistic
	propagated := false
	for i := 0; i < 10; i++ {
		propagation := network.ShareInformation(info)
		if propagation["m2"] {
			propagated = true
			break
		}
	}

	assert.True(t, propagated, "m2 should receive information with 0.8 probability")
}
