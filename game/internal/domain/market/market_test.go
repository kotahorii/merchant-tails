package market

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/merchant-tails/game/internal/domain/item"
)

func TestNewMarket(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "create new market successfully",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			market := NewMarket()
			assert.NotNil(t, market)
			assert.NotNil(t, market.PricingEngine)
			assert.NotNil(t, market.State)
			assert.Equal(t, DemandNormal, market.State.CurrentDemand)
			assert.Equal(t, SupplyNormal, market.State.CurrentSupply)
		})
	}
}

func TestPricingEngine_CalculatePrice(t *testing.T) {
	engine := NewPricingEngine()

	testItem, err := item.NewItem("test_001", "Test Item", item.CategoryFruit, 100)
	require.NoError(t, err)

	tests := []struct {
		name        string
		item        *item.Item
		demand      DemandLevel
		supply      SupplyLevel
		season      item.Season
		expectedMin int
		expectedMax int
	}{
		{
			name:        "normal market conditions",
			item:        testItem,
			demand:      DemandNormal,
			supply:      SupplyNormal,
			season:      item.SeasonSpring,
			expectedMin: 95,
			expectedMax: 115,
		},
		{
			name:        "high demand low supply",
			item:        testItem,
			demand:      DemandHigh,
			supply:      SupplyLow,
			season:      item.SeasonSpring,
			expectedMin: 120,
			expectedMax: 160,
		},
		{
			name:        "low demand high supply",
			item:        testItem,
			demand:      DemandLow,
			supply:      SupplyHigh,
			season:      item.SeasonSpring,
			expectedMin: 60,
			expectedMax: 80,
		},
		{
			name:        "very high demand very low supply",
			item:        testItem,
			demand:      DemandVeryHigh,
			supply:      SupplyVeryLow,
			season:      item.SeasonSpring,
			expectedMin: 150,
			expectedMax: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &MarketState{
				CurrentDemand: tt.demand,
				CurrentSupply: tt.supply,
				CurrentSeason: tt.season,
			}

			price := engine.CalculatePrice(tt.item, state)
			assert.GreaterOrEqual(t, price, tt.expectedMin, "Price should be at least %d but was %d", tt.expectedMin, price)
			assert.LessOrEqual(t, price, tt.expectedMax, "Price should be at most %d but was %d", tt.expectedMax, price)
		})
	}
}

func TestMarket_UpdatePrices(t *testing.T) {
	market := NewMarket()

	// Add test items
	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	potion, _ := item.NewItem("potion_001", "Health Potion", item.CategoryPotion, 50)
	sword, _ := item.NewItem("sword_001", "Iron Sword", item.CategoryWeapon, 200)

	market.RegisterItem(apple)
	market.RegisterItem(potion)
	market.RegisterItem(sword)

	// Initial prices
	initialPrices := make(map[string]int)
	for id, history := range market.Prices {
		initialPrices[id] = history.CurrentPrice
	}

	// Update market conditions
	market.State.CurrentDemand = DemandHigh
	market.State.CurrentSupply = SupplyLow

	// Update prices
	market.UpdatePrices()

	// Check that prices have changed
	for id, history := range market.Prices {
		newPrice := history.CurrentPrice
		oldPrice := initialPrices[id]
		assert.NotEqual(t, oldPrice, newPrice, "Price for %s should have changed", id)
		assert.Greater(t, newPrice, oldPrice, "Price should increase with high demand and low supply")
	}
}

func TestMarketState_GetDemandModifier(t *testing.T) {
	tests := []struct {
		name     string
		demand   DemandLevel
		expected float64
	}{
		{"very low demand", DemandVeryLow, 0.7},
		{"low demand", DemandLow, 0.85},
		{"normal demand", DemandNormal, 1.0},
		{"high demand", DemandHigh, 1.2},
		{"very high demand", DemandVeryHigh, 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &MarketState{CurrentDemand: tt.demand}
			modifier := state.GetDemandModifier()
			assert.Equal(t, tt.expected, modifier)
		})
	}
}

func TestMarketState_GetSupplyModifier(t *testing.T) {
	tests := []struct {
		name     string
		supply   SupplyLevel
		expected float64
	}{
		{"very low supply", SupplyVeryLow, 1.3},
		{"low supply", SupplyLow, 1.15},
		{"normal supply", SupplyNormal, 1.0},
		{"high supply", SupplyHigh, 0.85},
		{"very high supply", SupplyVeryHigh, 0.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &MarketState{CurrentSupply: tt.supply}
			modifier := state.GetSupplyModifier()
			assert.Equal(t, tt.expected, modifier)
		})
	}
}

func TestMarket_SimulateEvent(t *testing.T) {
	market := NewMarket()

	// Register items
	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	market.RegisterItem(apple)

	tests := []struct {
		name       string
		event      *MarketEvent
		wantDemand DemandLevel
		wantSupply SupplyLevel
	}{
		{
			name: "dragon attack event",
			event: &MarketEvent{
				Type:        EventDragonAttack,
				Name:        "Dragon Attack",
				Description: "A dragon attacks the trade routes",
				Duration:    3,
				Effects: []EventEffect{
					{Type: EffectSupplyDecrease, Value: 2},
					{Type: EffectDemandIncrease, Value: 1},
				},
			},
			wantDemand: DemandHigh,
			wantSupply: SupplyVeryLow,
		},
		{
			name: "harvest festival",
			event: &MarketEvent{
				Type:        EventHarvestFestival,
				Name:        "Harvest Festival",
				Description: "The annual harvest festival begins",
				Duration:    7,
				Effects: []EventEffect{
					{Type: EffectSupplyIncrease, Value: 2},
					{Type: EffectPriceModifier, Value: 0.9},
				},
			},
			wantDemand: DemandNormal,
			wantSupply: SupplyVeryHigh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset market state
			market.State.CurrentDemand = DemandNormal
			market.State.CurrentSupply = SupplyNormal

			// Apply event
			market.ApplyEvent(tt.event)

			assert.Equal(t, tt.wantDemand, market.State.CurrentDemand)
			assert.Equal(t, tt.wantSupply, market.State.CurrentSupply)
			assert.Contains(t, market.ActiveEvents, tt.event)
		})
	}
}

func TestMarket_PriceVolatility(t *testing.T) {
	market := NewMarket()

	// Create items with different volatilities
	fruit, _ := item.NewItem("fruit_001", "Apple", item.CategoryFruit, 10)
	weapon, _ := item.NewItem("weapon_001", "Sword", item.CategoryWeapon, 200)
	gem, _ := item.NewItem("gem_001", "Ruby", item.CategoryGem, 1000)

	market.RegisterItem(fruit)
	market.RegisterItem(weapon)
	market.RegisterItem(gem)

	// Simulate multiple price updates
	priceChanges := make(map[string][]int)

	for i := 0; i < 10; i++ {
		// Randomly change market conditions
		if i%2 == 0 {
			market.State.CurrentDemand = DemandHigh
		} else {
			market.State.CurrentDemand = DemandLow
		}

		market.UpdatePrices()

		for id, history := range market.Prices {
			priceChanges[id] = append(priceChanges[id], history.CurrentPrice)
		}
	}

	// Calculate variance for each item
	fruitVariance := calculateVariance(priceChanges["fruit_001"])
	weaponVariance := calculateVariance(priceChanges["weapon_001"])
	gemVariance := calculateVariance(priceChanges["gem_001"])

	// Gem should have highest variance, weapon lowest
	assert.Greater(t, gemVariance, fruitVariance, "Gem should be more volatile than fruit")
	assert.Greater(t, gemVariance, weaponVariance, "Gem should be more volatile than weapon")
}

func TestMarket_SeasonalPricing(t *testing.T) {
	market := NewMarket()

	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	market.RegisterItem(apple)

	seasons := []item.Season{
		item.SeasonSpring,
		item.SeasonSummer,
		item.SeasonAutumn,
		item.SeasonWinter,
	}

	prices := make(map[item.Season]int)

	for _, season := range seasons {
		market.State.CurrentSeason = season
		market.UpdatePrices()
		prices[season] = market.Prices["apple_001"].CurrentPrice
	}

	// Autumn should have highest price for fruit (harvest season)
	// Winter should have lowest
	assert.Greater(t, prices[item.SeasonAutumn], prices[item.SeasonSpring])
	assert.Greater(t, prices[item.SeasonAutumn], prices[item.SeasonSummer])
	assert.Less(t, prices[item.SeasonWinter], prices[item.SeasonAutumn])
}

func TestMarket_PriceHistory(t *testing.T) {
	market := NewMarket()

	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	market.RegisterItem(apple)

	// Update prices multiple times
	for i := 0; i < 10; i++ {
		market.UpdatePrices()
		time.Sleep(10 * time.Millisecond) // Small delay to ensure different timestamps
	}

	history := market.GetPriceHistory("apple_001")

	// Check history is maintained
	assert.NotNil(t, history)
	assert.GreaterOrEqual(t, len(history.Records), 5)
	assert.LessOrEqual(t, len(history.Records), 10)

	// Check trend calculation
	trend := history.GetTrend()
	assert.Contains(t, []PriceTrend{TrendUp, TrendDown, TrendStable}, trend)
}

func TestMarket_GetRecommendedAction(t *testing.T) {
	market := NewMarket()

	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	market.RegisterItem(apple)

	tests := []struct {
		name           string
		currentPrice   int
		averagePrice   int
		trend          PriceTrend
		expectedAction TradeAction
	}{
		{
			name:           "buy when price below average and trending up",
			currentPrice:   8,
			averagePrice:   10,
			trend:          TrendUp,
			expectedAction: ActionBuy,
		},
		{
			name:           "sell when price above average and trending down",
			currentPrice:   12,
			averagePrice:   10,
			trend:          TrendDown,
			expectedAction: ActionSell,
		},
		{
			name:           "hold when price stable near average",
			currentPrice:   10,
			averagePrice:   10,
			trend:          TrendStable,
			expectedAction: ActionHold,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock price history
			history := &PriceHistory{
				CurrentPrice: tt.currentPrice,
				AveragePrice: tt.averagePrice,
				Trend:        tt.trend,
			}
			market.Prices["apple_001"] = history

			action := market.GetRecommendedAction("apple_001")
			assert.Equal(t, tt.expectedAction, action)
		})
	}
}

// Helper function to calculate variance
func calculateVariance(prices []int) float64 {
	if len(prices) == 0 {
		return 0
	}

	// Calculate mean
	sum := 0
	for _, p := range prices {
		sum += p
	}
	mean := float64(sum) / float64(len(prices))

	// Calculate variance
	var variance float64
	for _, p := range prices {
		diff := float64(p) - mean
		variance += diff * diff
	}

	return variance / float64(len(prices))
}
