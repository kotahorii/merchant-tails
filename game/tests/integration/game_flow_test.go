//go:build integration
// +build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/yourusername/merchant-tails/game/internal/domain/achievement"
	"github.com/yourusername/merchant-tails/game/internal/domain/balance"
	"github.com/yourusername/merchant-tails/game/internal/domain/bank"
	"github.com/yourusername/merchant-tails/game/internal/domain/calendar"
	"github.com/yourusername/merchant-tails/game/internal/domain/difficulty"
	"github.com/yourusername/merchant-tails/game/internal/domain/event"
	"github.com/yourusername/merchant-tails/game/internal/domain/gamestate"
	"github.com/yourusername/merchant-tails/game/internal/domain/item"
	"github.com/yourusername/merchant-tails/game/internal/domain/progression"
	"github.com/yourusername/merchant-tails/game/internal/domain/quest"
	"github.com/yourusername/merchant-tails/game/internal/domain/tutorial"
	"github.com/yourusername/merchant-tails/game/internal/domain/weather"
	"github.com/yourusername/merchant-tails/game/internal/presentation/api"
)

// GameFlowTestSuite tests the complete game flow integration
type GameFlowTestSuite struct {
	suite.Suite
	gameManager       *api.GameManager
	eventBus          *event.EventBus
	achievementSystem *achievement.AchievementSystem
	questSystem       *quest.QuestSystem
	priceBalancer     *balance.PriceBalancer
	difficultyManager *difficulty.DifficultyManager
	weatherSystem     *weather.WeatherSystem
	eventCalendar     *calendar.EventCalendar
}

func (suite *GameFlowTestSuite) SetupTest() {
	// Initialize event bus
	suite.eventBus = event.NewEventBus(100)

	// Initialize game manager with all systems
	suite.gameManager = api.NewGameManager(suite.eventBus)

	// Initialize individual systems for direct testing
	suite.achievementSystem = achievement.NewAchievementSystem()
	suite.questSystem = quest.NewQuestSystem()
	suite.priceBalancer = balance.NewPriceBalancer(nil)
	suite.difficultyManager = difficulty.NewDifficultyManager(nil)
	suite.weatherSystem = weather.NewWeatherSystem(nil)
	suite.eventCalendar = calendar.NewEventCalendar()
}

func (suite *GameFlowTestSuite) TearDownTest() {
	// Clean up
	suite.eventBus.Stop()
}

// TestNewGameFlow tests starting a new game
func (suite *GameFlowTestSuite) TestNewGameFlow() {
	// Start new game
	gameID, err := suite.gameManager.NewGame("TestPlayer")
	suite.NoError(err)
	suite.NotEmpty(gameID)

	// Verify initial state
	state := suite.gameManager.GetGameState()
	suite.NotNil(state)
	suite.Equal("TestPlayer", state.PlayerName)
	suite.Equal(1000.0, state.Gold)
	suite.Equal(gamestate.RankApprentice, state.Rank)
	suite.Equal(gamestate.StatePlaying, state.State)

	// Verify tutorial starts
	tutorialManager := suite.gameManager.GetTutorialManager()
	suite.NotNil(tutorialManager)
	suite.True(tutorialManager.IsActive())
	suite.Equal(tutorial.StepWelcome, tutorialManager.GetCurrentStep())
}

// TestBuyAndSellFlow tests basic trading flow
func (suite *GameFlowTestSuite) TestBuyAndSellFlow() {
	// Setup game
	suite.gameManager.NewGame("Trader")

	// Get initial gold
	initialGold := suite.gameManager.GetGameState().Gold

	// Buy an item
	buyResult := suite.gameManager.BuyItem("apple", 10.0, 5)
	suite.True(buyResult.Success)
	suite.Equal(50.0, buyResult.TotalCost)

	// Verify gold decreased
	suite.Equal(initialGold-50.0, suite.gameManager.GetGameState().Gold)

	// Verify inventory
	inventory := suite.gameManager.GetPlayerInventory()
	suite.NotNil(inventory)
	apples, exists := inventory.GetItem("apple")
	suite.True(exists)
	suite.Equal(5, apples.Quantity)

	// Sell the item at higher price
	sellResult := suite.gameManager.SellItem("apple", 15.0, 3)
	suite.True(sellResult.Success)
	suite.Equal(45.0, sellResult.TotalRevenue)
	suite.Equal(15.0, sellResult.Profit) // 45 - (10*3) = 15

	// Verify gold increased
	suite.Equal(initialGold-50.0+45.0, suite.gameManager.GetGameState().Gold)

	// Verify inventory updated
	apples, exists = inventory.GetItem("apple")
	suite.True(exists)
	suite.Equal(2, apples.Quantity)
}

// TestProgressionFlow tests player progression through ranks
func (suite *GameFlowTestSuite) TestProgressionFlow() {
	// Setup
	suite.gameManager.NewGame("ProPlayer")
	progressionManager := progression.NewProgressionManager(suite.eventBus)

	// Simulate successful trades to gain experience
	for i := 0; i < 20; i++ {
		progressionManager.RecordTrade(100.0, 20.0) // 20% profit margin
	}

	// Check if player should rank up
	suite.True(progressionManager.ShouldRankUp())

	// Rank up
	newRank := progressionManager.RankUp()
	suite.Equal(gamestate.RankJourneyman, newRank)

	// Verify rank benefits
	benefits := progressionManager.GetRankBenefits(newRank)
	suite.Greater(benefits.ShopCapacity, 20)
	suite.Greater(benefits.WarehouseCapacity, 50)
}

// TestMarketEventFlow tests market events affecting prices
func (suite *GameFlowTestSuite) TestMarketEventFlow() {
	// Setup
	suite.gameManager.NewGame("EventTrader")
	marketManager := suite.gameManager.GetMarketManager()

	// Get initial price
	initialPrice := marketManager.GetCurrentPrice("apple")

	// Trigger a harvest festival event (should reduce fruit prices)
	festivalEvent := &event.GameEvent{
		Base: event.NewBaseEvent("market.harvest_festival"),
		Data: map[string]interface{}{
			"category":         item.CategoryFruit,
			"price_multiplier": 0.7,
		},
	}
	suite.eventBus.Publish(festivalEvent)

	// Give time for event processing
	time.Sleep(100 * time.Millisecond)

	// Check price decreased
	newPrice := marketManager.GetCurrentPrice("apple")
	suite.Less(newPrice, initialPrice)
	suite.InDelta(initialPrice*0.7, newPrice, 1.0)
}

// TestQuestFlow tests quest system integration
func (suite *GameFlowTestSuite) TestQuestFlow() {
	// Setup
	suite.gameManager.NewGame("QuestPlayer")

	// Start a quest
	firstProfitQuest := suite.questSystem.GetQuest("first_profit")
	suite.NotNil(firstProfitQuest)
	suite.questSystem.StartQuest("first_profit")

	// Complete quest objective (make profitable trade)
	suite.gameManager.BuyItem("sword", 100.0, 1)
	suite.gameManager.SellItem("sword", 130.0, 1)

	// Update quest progress
	suite.questSystem.UpdateProgress("first_profit", "profit", 30.0)

	// Check quest completion
	suite.True(suite.questSystem.IsQuestCompleted("first_profit"))

	// Claim rewards
	rewards := suite.questSystem.ClaimRewards("first_profit")
	suite.NotNil(rewards)
	suite.Equal(100.0, rewards.Gold)
	suite.Equal(50, rewards.Experience)
}

// TestDifficultyAdaptation tests difficulty system adapting to player
func (suite *GameFlowTestSuite) TestDifficultyAdaptation() {
	// Setup
	suite.gameManager.NewGame("AdaptivePlayer")

	// Record successful trades
	for i := 0; i < 15; i++ {
		suite.difficultyManager.RecordTrade(true, 50.0, time.Second)
	}

	// Check difficulty increased
	currentDifficulty := suite.difficultyManager.GetCurrentDifficulty()
	suite.Greater(currentDifficulty, difficulty.DifficultyTutorial)

	// Get difficulty modifiers
	modifiers := suite.difficultyManager.GetModifiers()
	suite.NotNil(modifiers)
	suite.Greater(modifiers.PriceMultiplier, 1.0)
	suite.Less(modifiers.GoldRewardMultiplier, 1.0)
}

// TestWeatherImpact tests weather affecting market
func (suite *GameFlowTestSuite) TestWeatherImpact() {
	// Setup
	suite.gameManager.NewGame("WeatherTrader")

	// Set rainy weather
	suite.weatherSystem.SetWeather(weather.WeatherRainy)

	// Get weather effects
	effects := suite.weatherSystem.GetCurrentEffects()
	suite.NotNil(effects)

	// Check market impact
	suite.Greater(effects.IndoorDemandMultiplier, 1.0)
	suite.Less(effects.CustomerTrafficMultiplier, 1.0)

	// Verify umbrella demand increased
	marketManager := suite.gameManager.GetMarketManager()
	// Assuming umbrella is categorized as accessory
	umbrellaPrice := marketManager.GetCurrentPrice("umbrella")
	suite.Greater(umbrellaPrice, 0)
}

// TestBankingFlow tests banking system integration
func (suite *GameFlowTestSuite) TestBankingFlow() {
	// Setup
	suite.gameManager.NewGame("BankCustomer")
	bankManager := bank.NewBankManager(suite.eventBus)

	// Make a deposit
	depositAmount := 500.0
	err := bankManager.Deposit(depositAmount)
	suite.NoError(err)

	// Simulate time passing for interest
	bankManager.ProcessDailyInterest()

	// Check balance with interest
	balance := bankManager.GetBalance()
	suite.Greater(balance, depositAmount)

	// Take a loan
	loanAmount := 1000.0
	err = bankManager.TakeLoan(loanAmount, 30) // 30 day loan
	suite.NoError(err)

	// Verify loan details
	loanDetails := bankManager.GetCurrentLoan()
	suite.NotNil(loanDetails)
	suite.Equal(loanAmount, loanDetails.Principal)
	suite.Greater(loanDetails.TotalDue, loanAmount)
}

// TestAchievementUnlock tests achievement system
func (suite *GameFlowTestSuite) TestAchievementUnlock() {
	// Setup
	suite.gameManager.NewGame("AchieverPlayer")

	// Make first sale to unlock achievement
	suite.gameManager.BuyItem("apple", 10.0, 1)
	suite.gameManager.SellItem("apple", 12.0, 1)

	// Update achievement progress
	suite.achievementSystem.UpdateProgress("first_sale", 1)

	// Check achievement unlocked
	unlockedAchievements := suite.achievementSystem.GetUnlockedAchievements()
	suite.Contains(unlockedAchievements, "first_sale")

	// Get achievement details
	achievement := suite.achievementSystem.GetAchievement("first_sale")
	suite.NotNil(achievement)
	suite.True(achievement.IsUnlocked)
}

// TestEventCalendarFlow tests event calendar system
func (suite *GameFlowTestSuite) TestEventCalendarFlow() {
	// Setup
	suite.gameManager.NewGame("EventPlayer")

	// Add a market day event
	marketDay := &calendar.CalendarEvent{
		ID:          "market_day",
		Name:        "Market Day",
		Description: "Increased customer traffic",
		StartDate:   time.Now().Add(24 * time.Hour),
		Duration:    8 * time.Hour,
		EventType:   calendar.EventTypeMarket,
		Priority:    calendar.PriorityHigh,
		Effects: map[string]float64{
			"customer_traffic": 1.5,
			"price_tolerance":  1.2,
		},
	}
	suite.eventCalendar.AddEvent(marketDay)

	// Get upcoming events
	upcomingEvents := suite.eventCalendar.GetUpcomingEvents(7)
	suite.NotEmpty(upcomingEvents)
	suite.Contains(upcomingEvents, marketDay)

	// Simulate time passing to activate event
	suite.eventCalendar.Update(time.Now().Add(25 * time.Hour))

	// Check event is active
	activeEvents := suite.eventCalendar.GetActiveEvents()
	suite.Contains(activeEvents, marketDay)
}

// TestSaveLoadFlow tests save and load functionality
func (suite *GameFlowTestSuite) TestSaveLoadFlow() {
	// Setup and modify game state
	gameID, _ := suite.gameManager.NewGame("SavePlayer")
	suite.gameManager.BuyItem("apple", 10.0, 5)
	suite.gameManager.UpdatePlayerGold(2000.0)

	// Save game
	saveData, err := suite.gameManager.SaveGame()
	suite.NoError(err)
	suite.NotNil(saveData)

	// Create new game manager
	newGameManager := api.NewGameManager(suite.eventBus)

	// Load saved game
	err = newGameManager.LoadGame(saveData)
	suite.NoError(err)

	// Verify state restored
	loadedState := newGameManager.GetGameState()
	suite.Equal("SavePlayer", loadedState.PlayerName)
	suite.Equal(2000.0, loadedState.Gold)
	suite.Equal(gameID, loadedState.ID)

	// Verify inventory restored
	inventory := newGameManager.GetPlayerInventory()
	apples, exists := inventory.GetItem("apple")
	suite.True(exists)
	suite.Equal(5, apples.Quantity)
}

// TestCompleteGameSession tests a complete game session
func (suite *GameFlowTestSuite) TestCompleteGameSession() {
	// Start new game
	suite.gameManager.NewGame("CompletePlayer")

	// Complete tutorial
	tutorialManager := suite.gameManager.GetTutorialManager()
	for tutorialManager.IsActive() {
		tutorialManager.NextStep()
	}

	// Make some trades
	trades := []struct {
		item      string
		buyPrice  float64
		sellPrice float64
		quantity  int
	}{
		{"apple", 10.0, 12.0, 10},
		{"sword", 100.0, 130.0, 2},
		{"potion", 25.0, 35.0, 5},
	}

	totalProfit := 0.0
	for _, trade := range trades {
		suite.gameManager.BuyItem(trade.item, trade.buyPrice, trade.quantity)
		result := suite.gameManager.SellItem(trade.item, trade.sellPrice, trade.quantity)
		totalProfit += result.Profit
	}

	// Check profit
	suite.Greater(totalProfit, 0.0)

	// Process daily summary
	summary := suite.gameManager.GetDailySummary()
	suite.NotNil(summary)
	suite.Greater(summary.TotalRevenue, 0.0)
	suite.Greater(summary.TotalProfit, 0.0)

	// Check if any achievements unlocked
	achievements := suite.achievementSystem.GetUnlockedAchievements()
	suite.NotEmpty(achievements)

	// Save game
	saveData, err := suite.gameManager.SaveGame()
	suite.NoError(err)
	suite.NotNil(saveData)
}

// TestConcurrentSystems tests multiple systems working concurrently
func (suite *GameFlowTestSuite) TestConcurrentSystems() {
	// Setup
	suite.gameManager.NewGame("ConcurrentPlayer")

	// Start multiple goroutines simulating different systems
	done := make(chan bool, 4)

	// Trading goroutine
	go func() {
		for i := 0; i < 10; i++ {
			suite.gameManager.BuyItem("apple", 10.0, 1)
			suite.gameManager.SellItem("apple", 12.0, 1)
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// Weather change goroutine
	go func() {
		weathers := []weather.WeatherType{
			weather.WeatherSunny,
			weather.WeatherRainy,
			weather.WeatherSnowy,
		}
		for i := 0; i < 10; i++ {
			suite.weatherSystem.SetWeather(weathers[i%3])
			time.Sleep(15 * time.Millisecond)
		}
		done <- true
	}()

	// Event processing goroutine
	go func() {
		for i := 0; i < 10; i++ {
			event := event.NewBaseEvent("test.event")
			suite.eventBus.PublishAsync(event)
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// Achievement checking goroutine
	go func() {
		for i := 0; i < 10; i++ {
			suite.achievementSystem.UpdateProgress("trades_made", 1)
			time.Sleep(12 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}

	// Verify system state is consistent
	state := suite.gameManager.GetGameState()
	suite.NotNil(state)
	suite.GreaterOrEqual(state.Gold, 0.0)
}

func TestGameFlowSuite(t *testing.T) {
	suite.Run(t, new(GameFlowTestSuite))
}
