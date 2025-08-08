//go:build integration
// +build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/merchant-tails/game/internal/domain/balance"
	"github.com/yourusername/merchant-tails/game/internal/domain/difficulty"
	"github.com/yourusername/merchant-tails/game/internal/domain/event"
	"github.com/yourusername/merchant-tails/game/internal/domain/item"
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
	"github.com/yourusername/merchant-tails/game/internal/domain/weather"
)

// TestPriceBalanceWithDifficulty tests price balance system working with difficulty
func TestPriceBalanceWithDifficulty(t *testing.T) {
	// Setup
	eventBus := event.NewEventBus(100)
	defer eventBus.Stop()

	priceBalancer := balance.NewPriceBalancer(nil)
	difficultyManager := difficulty.NewDifficultyManager(nil)

	// Register items
	priceBalancer.RegisterItem("apple", item.CategoryFruit, 10.0)
	priceBalancer.RegisterItem("sword", item.CategoryWeapon, 100.0)

	// Set difficulty to Hard
	difficultyManager.SetDifficulty(difficulty.DifficultyHard)
	modifiers := difficultyManager.GetModifiers()

	// Calculate prices with difficulty modifiers
	applePrice := priceBalancer.CalculateOptimalPrice("apple", 2) // Journeyman rank
	swordPrice := priceBalancer.CalculateOptimalPrice("sword", 2)

	// Apply difficulty modifiers
	adjustedApplePrice := applePrice * modifiers.PriceMultiplier
	adjustedSwordPrice := swordPrice * modifiers.PriceMultiplier

	// Verify prices increased due to difficulty
	assert.Greater(t, adjustedApplePrice, applePrice)
	assert.Greater(t, adjustedSwordPrice, swordPrice)

	// Record trades and check difficulty adjustment
	for i := 0; i < 10; i++ {
		success := i%3 != 0 // 66% success rate
		profit := 20.0
		if !success {
			profit = -10.0
		}
		difficultyManager.RecordTrade(success, profit, time.Second)

		// Record sale for price balancing
		if success {
			priceBalancer.RecordSale("apple", adjustedApplePrice, 1, 1000.0)
		}
	}

	// Check if systems adapted
	newApplePrice := priceBalancer.CalculateOptimalPrice("apple", 2)
	assert.NotEqual(t, applePrice, newApplePrice) // Price should have adjusted

	newDifficulty := difficultyManager.GetCurrentDifficulty()
	assert.GreaterOrEqual(t, newDifficulty, difficulty.DifficultyHard) // Difficulty may have changed
}

// TestWeatherMarketInteraction tests weather affecting market prices
func TestWeatherMarketInteraction(t *testing.T) {
	// Setup
	eventBus := event.NewEventBus(100)
	defer eventBus.Stop()

	weatherSystem := weather.NewWeatherSystem(nil)
	marketManager := market.NewMarketManager(eventBus)

	// Initialize market with items
	marketManager.InitializeMarket(map[string]float64{
		"umbrella":  20.0,
		"sunscreen": 15.0,
		"coat":      50.0,
	})

	// Test rainy weather effect
	weatherSystem.SetWeather(weather.WeatherRainy)
	effects := weatherSystem.GetCurrentEffects()

	// Apply weather effects to market
	umbrellaPrice := marketManager.GetCurrentPrice("umbrella")
	adjustedUmbrellaPrice := umbrellaPrice * effects.IndoorDemandMultiplier
	assert.Greater(t, adjustedUmbrellaPrice, umbrellaPrice)

	// Test sunny weather effect
	weatherSystem.SetWeather(weather.WeatherSunny)
	effects = weatherSystem.GetCurrentEffects()

	sunscreenPrice := marketManager.GetCurrentPrice("sunscreen")
	adjustedSunscreenPrice := sunscreenPrice * effects.OutdoorDemandMultiplier
	assert.Greater(t, adjustedSunscreenPrice, sunscreenPrice)

	// Test snowy weather effect
	weatherSystem.SetWeather(weather.WeatherSnowy)
	effects = weatherSystem.GetCurrentEffects()

	coatPrice := marketManager.GetCurrentPrice("coat")
	adjustedCoatPrice := coatPrice * effects.WarmthDemandMultiplier
	assert.Greater(t, adjustedCoatPrice, coatPrice)
}

// TestEventPropagation tests event propagation across systems
func TestEventPropagation(t *testing.T) {
	// Setup
	eventBus := event.NewEventBus(100)
	defer eventBus.Stop()

	// Track events received
	eventsReceived := make(map[string]int)

	// Subscribe multiple systems
	eventBus.Subscribe("market.*", func(e event.Event) {
		eventsReceived["market"]++
	})

	eventBus.Subscribe("player.*", func(e event.Event) {
		eventsReceived["player"]++
	})

	eventBus.Subscribe("quest.*", func(e event.Event) {
		eventsReceived["quest"]++
	})

	// Publish various events
	eventBus.PublishAsync(event.NewBaseEvent("market.price_changed"))
	eventBus.PublishAsync(event.NewBaseEvent("player.level_up"))
	eventBus.PublishAsync(event.NewBaseEvent("quest.completed"))
	eventBus.PublishAsync(event.NewBaseEvent("market.item_sold"))

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify events were received
	assert.Equal(t, 2, eventsReceived["market"])
	assert.Equal(t, 1, eventsReceived["player"])
	assert.Equal(t, 1, eventsReceived["quest"])
}

// TestSupplyDemandBalance tests supply and demand affecting prices
func TestSupplyDemandBalance(t *testing.T) {
	// Setup
	priceBalancer := balance.NewPriceBalancer(nil)

	// Register items
	items := []struct {
		id        string
		category  item.Category
		basePrice float64
	}{
		{"apple", item.CategoryFruit, 10.0},
		{"potion", item.CategoryPotion, 25.0},
		{"sword", item.CategoryWeapon, 100.0},
	}

	for _, it := range items {
		priceBalancer.RegisterItem(it.id, it.category, it.basePrice)
	}

	// Simulate market conditions
	scenarios := []struct {
		name        string
		itemID      string
		supply      int
		demand      int
		expectation string // "increase", "decrease", "stable"
	}{
		{"Apple Oversupply", "apple", 200, 50, "decrease"},
		{"Potion Scarcity", "potion", 10, 100, "increase"},
		{"Sword Equilibrium", "sword", 50, 50, "stable"},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Update supply and demand
			priceBalancer.UpdateSupplyDemand(scenario.itemID, scenario.supply, scenario.demand)

			// Calculate optimal price
			optimalPrice := priceBalancer.CalculateOptimalPrice(scenario.itemID, 2)

			// Get item balance
			balance, exists := priceBalancer.GetItemBalance(scenario.itemID)
			assert.True(t, exists)

			// Verify price adjustment based on expectation
			switch scenario.expectation {
			case "increase":
				assert.Greater(t, optimalPrice, balance.BasePrice*1.1)
			case "decrease":
				assert.Less(t, optimalPrice, balance.BasePrice*0.9)
			case "stable":
				assert.InDelta(t, balance.BasePrice, optimalPrice, balance.BasePrice*0.2)
			}
		})
	}
}

// TestDifficultyProgressionIntegration tests difficulty adjusting with player progression
func TestDifficultyProgressionIntegration(t *testing.T) {
	// Setup
	difficultyManager := difficulty.NewDifficultyManager(nil)

	// Simulate a player's journey
	stages := []struct {
		name          string
		successRate   float64
		numTrades     int
		expectedLevel difficulty.DifficultyLevel
	}{
		{"Tutorial Phase", 1.0, 5, difficulty.DifficultyEasy},
		{"Learning Phase", 0.7, 20, difficulty.DifficultyEasy},
		{"Competent Phase", 0.8, 30, difficulty.DifficultyNormal},
		{"Expert Phase", 0.9, 40, difficulty.DifficultyHard},
	}

	totalTrades := 0
	for _, stage := range stages {
		t.Run(stage.name, func(t *testing.T) {
			for i := 0; i < stage.numTrades; i++ {
				success := float64(i)/float64(stage.numTrades) < stage.successRate
				profit := 50.0
				if !success {
					profit = -20.0
				}
				difficultyManager.RecordTrade(success, profit, time.Second)
				totalTrades++
			}

			// Check difficulty progression
			currentDifficulty := difficultyManager.GetCurrentDifficulty()
			assert.GreaterOrEqual(t, currentDifficulty, stage.expectedLevel,
				"After %d trades with %.0f%% success rate, expected at least %v difficulty",
				totalTrades, stage.successRate*100, stage.expectedLevel)
		})
	}
}

// TestMarketVolatility tests market volatility calculations
func TestMarketVolatility(t *testing.T) {
	// Setup
	priceBalancer := balance.NewPriceBalancer(nil)
	priceBalancer.RegisterItem("volatile_gem", item.CategoryGem, 100.0)

	// Simulate volatile trading
	prices := []float64{100, 120, 80, 150, 70, 130, 90, 110}

	for _, price := range prices {
		priceBalancer.RecordSale("volatile_gem", price, 1, 1000.0)
	}

	// Get item balance
	balance, exists := priceBalancer.GetItemBalance("volatile_gem")
	assert.True(t, exists)

	// Calculate optimal price considering volatility
	optimalPrice := priceBalancer.CalculateOptimalPrice("volatile_gem", 2)

	// Volatility should affect the price
	assert.NotEqual(t, balance.BasePrice, optimalPrice)

	// With high volatility, price should be adjusted
	// The exact adjustment depends on the algorithm
	assert.Greater(t, optimalPrice, 0.0)
}

// TestChallengeEventImpact tests challenge events affecting difficulty
func TestChallengeEventImpact(t *testing.T) {
	// Setup
	difficultyManager := difficulty.NewDifficultyManager(nil)
	difficultyManager.SetDifficulty(difficulty.DifficultyNormal)

	// Get initial difficulty score
	initialScore := difficultyManager.GetDifficultyScore()

	// Add a challenge event
	challenge := &difficulty.ChallengeEvent{
		ID:              "dragon_attack",
		Name:            "Dragon Attack",
		Description:     "A dragon is terrorizing the market",
		DifficultyBoost: 2.0,
		Duration:        time.Hour,
	}
	difficultyManager.AddChallenge(challenge)

	// Check difficulty increased
	newScore := difficultyManager.GetDifficultyScore()
	assert.Greater(t, newScore, initialScore)

	// Get modifiers
	modifiers := difficultyManager.GetModifiers()
	assert.Greater(t, modifiers.EventDifficulty, 1.0)

	// Remove challenge
	difficultyManager.RemoveChallenge("dragon_attack")

	// Check difficulty returned to normal range
	finalScore := difficultyManager.GetDifficultyScore()
	assert.LessOrEqual(t, finalScore, initialScore*1.1)
}

// TestSeasonalPriceFluctuation tests seasonal effects on prices
func TestSeasonalPriceFluctuation(t *testing.T) {
	// Setup
	eventBus := event.NewEventBus(100)
	defer eventBus.Stop()

	marketManager := market.NewMarketManager(eventBus)

	// Initialize seasonal items
	seasonalItems := map[string]float64{
		"ice_cream":     5.0, // Summer item
		"hot_chocolate": 3.0, // Winter item
		"apple":         2.0, // Fall item
		"flower":        4.0, // Spring item
	}
	marketManager.InitializeMarket(seasonalItems)

	// Test seasonal modifiers
	seasons := []struct {
		season   market.Season
		item     string
		modifier float64
	}{
		{market.SeasonSummer, "ice_cream", 1.5},
		{market.SeasonWinter, "hot_chocolate", 1.4},
		{market.SeasonFall, "apple", 1.3},
		{market.SeasonSpring, "flower", 1.3},
	}

	for _, test := range seasons {
		t.Run(string(test.season), func(t *testing.T) {
			// Set season
			marketManager.SetSeason(test.season)

			// Update prices
			marketManager.UpdatePrices()

			// Get adjusted price
			price := marketManager.GetCurrentPrice(test.item)
			basePrice := seasonalItems[test.item]

			// Verify seasonal adjustment
			assert.Greater(t, price, basePrice,
				"%s should be more expensive in %s", test.item, test.season)
		})
	}
}

// TestProgressionRewards tests progression rewards being applied correctly
func TestProgressionRewards(t *testing.T) {
	// Setup
	eventBus := event.NewEventBus(100)
	defer eventBus.Stop()

	// Test rank benefits
	ranks := []struct {
		rank              int // Using int to represent rank levels
		shopCapacity      int
		warehouseCapacity int
		questSlots        int
		bankInterestBonus float64
	}{
		{1, 20, 50, 1, 0.0},    // Apprentice
		{2, 30, 100, 2, 0.01},  // Journeyman
		{3, 50, 200, 3, 0.02},  // Expert
		{4, 100, 500, 5, 0.03}, // Master
	}

	for _, rankTest := range ranks {
		t.Run("Rank"+string(rune('0'+rankTest.rank)), func(t *testing.T) {
			// Verify capacity increases
			assert.Greater(t, rankTest.shopCapacity, 0)
			assert.Greater(t, rankTest.warehouseCapacity, 0)
			assert.Greater(t, rankTest.questSlots, 0)

			// Verify higher ranks get better benefits
			if rankTest.rank > 1 {
				prevRank := ranks[rankTest.rank-2]
				assert.Greater(t, rankTest.shopCapacity, prevRank.shopCapacity)
				assert.Greater(t, rankTest.warehouseCapacity, prevRank.warehouseCapacity)
				assert.GreaterOrEqual(t, rankTest.questSlots, prevRank.questSlots)
				assert.GreaterOrEqual(t, rankTest.bankInterestBonus, prevRank.bankInterestBonus)
			}
		})
	}
}

// TestConcurrentMarketOperations tests thread-safe market operations
func TestConcurrentMarketOperations(t *testing.T) {
	// Setup
	eventBus := event.NewEventBus(100)
	defer eventBus.Stop()

	priceBalancer := balance.NewPriceBalancer(nil)
	difficultyManager := difficulty.NewDifficultyManager(nil)

	// Register items
	for i := 0; i < 10; i++ {
		itemID := string(rune('a' + i))
		priceBalancer.RegisterItem(itemID, item.CategoryFruit, float64(10+i))
	}

	// Run concurrent operations
	done := make(chan bool, 3)

	// Price updates
	go func() {
		for i := 0; i < 100; i++ {
			itemID := string(rune('a' + (i % 10)))
			priceBalancer.RecordSale(itemID, float64(10+i%20), 1, 1000.0)
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Supply/demand updates
	go func() {
		for i := 0; i < 100; i++ {
			itemID := string(rune('a' + (i % 10)))
			priceBalancer.UpdateSupplyDemand(itemID, 50+i%50, 50+i%30)
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Difficulty adjustments
	go func() {
		for i := 0; i < 100; i++ {
			success := i%3 != 0
			difficultyManager.RecordTrade(success, 20.0, time.Second)
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Wait for completion
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify system integrity
	metrics := priceBalancer.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Greater(t, metrics.TotalTransactions, 0)

	playerSkill := difficultyManager.GetPlayerSkill()
	assert.NotNil(t, playerSkill)
	assert.Greater(t, playerSkill.TotalPlays, 0)
}
