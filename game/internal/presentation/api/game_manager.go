package api

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/ai"
	"github.com/yourusername/merchant-tails/game/internal/domain/event"
	"github.com/yourusername/merchant-tails/game/internal/domain/gameloop"
	"github.com/yourusername/merchant-tails/game/internal/domain/gamestate"
	"github.com/yourusername/merchant-tails/game/internal/domain/inventory"
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
	"github.com/yourusername/merchant-tails/game/internal/domain/progression"
	timemanager "github.com/yourusername/merchant-tails/game/internal/domain/time"
	"github.com/yourusername/merchant-tails/game/internal/domain/trading"
	"github.com/yourusername/merchant-tails/game/internal/infrastructure/persistence"
)

// GameManager manages the overall game state and coordinates between systems
type GameManager struct {
	// Core systems
	gameState   *gamestate.GameState
	gameLoop    gameloop.GameLoop
	timeManager *timemanager.TimeManager
	eventBus    *event.EventBus
	eventBridge *EventBridge

	// Game systems
	market      *market.Market
	inventory   *inventory.InventoryManager
	trading     *trading.TradingSystem
	progression *progression.ProgressionManager
	aiSystem    *ai.AIMerchantSystem

	// Infrastructure
	saveManager *persistence.SaveManager

	// State management
	isRunning bool
	isPaused  bool
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
}

// NewGameManager creates a new game manager instance
func NewGameManager() *GameManager {
	ctx, cancel := context.WithCancel(context.Background())

	gm := &GameManager{
		gameState:   gamestate.NewGameState(),
		timeManager: timemanager.NewTimeManager(),
		eventBus:    event.GetGlobalEventBus(),
		eventBridge: NewEventBridge(),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initialize systems
	gm.initializeSystems()

	return gm
}

// initializeSystems initializes all game systems
func (gm *GameManager) initializeSystems() {
	// Create market
	gm.market = market.NewMarket()

	// Create inventory manager
	invManager, err := inventory.NewInventoryManager(100, 200) // Shop: 100, Warehouse: 200
	if err != nil {
		panic(err) // Should not happen with valid capacities
	}
	gm.inventory = invManager

	// Create trading system
	gm.trading = trading.NewTradingSystem(gm.market)

	// Create progression manager
	gm.progression = progression.NewProgressionManager()

	// Create AI system
	gm.aiSystem = ai.NewAISystem()

	// Create save manager
	saveManager, err := persistence.NewSaveManager()
	if err != nil {
		// Log error but continue - save/load will be disabled
		fmt.Printf("Failed to initialize save manager: %v\n", err)
	}
	gm.saveManager = saveManager

	// Create game loop
	gm.gameLoop = gameloop.NewStandardGameLoop(60) // 60 FPS

	// Setup event listeners
	gm.setupEventListeners()
}

// setupEventListeners sets up event handlers
func (gm *GameManager) setupEventListeners() {
	// Listen for time events
	gm.eventBus.Subscribe(event.TimeAdvanced, func(e event.Event) {
		gm.handleTimeAdvanced(e)
	})

	// Listen for trade events
	gm.eventBus.Subscribe(event.TradeCompleted, func(e event.Event) {
		gm.handleTradeCompleted(e)
	})

	// Listen for market events
	gm.eventBus.Subscribe(event.MarketPriceChanged, func(e event.Event) {
		gm.handleMarketPriceChanged(e)
	})
}

// StartNewGame starts a new game
func (gm *GameManager) StartNewGame(playerName string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if gm.isRunning {
		return fmt.Errorf("game is already running")
	}

	// Reset game state
	gm.gameState = gamestate.NewGameState()
	gm.gameState.SetPlayerName(playerName)
	gm.gameState.SetGold(1000) // Starting gold
	gm.gameState.SetRank(gamestate.RankApprentice)

	// Reset systems
	gm.progression.ResetProgression()
	gm.market.Reset()
	gm.inventory.Clear()

	// Initialize AI merchants
	gm.initializeAIMerchants()

	// Start game loop
	gm.isRunning = true
	gm.isPaused = false

	go gm.runGameLoop()

	// Publish game started event
	gm.eventBus.Publish(event.Event{
		Type:      "GameStarted",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"playerName": playerName,
		},
	})

	return nil
}

// runGameLoop runs the main game loop
func (gm *GameManager) runGameLoop() {
	gm.gameLoop.Start(gm.ctx, func(deltaTime float64) {
		if !gm.isPaused {
			gm.update(deltaTime)
		}
	})
}

// update updates all game systems
func (gm *GameManager) update(deltaTime float64) {
	// Update time
	gm.timeManager.Update(deltaTime)

	// Update market
	if gm.market != nil {
		gm.market.Update()
	}

	// Update AI system
	if gm.aiSystem != nil {
		gm.aiSystem.Update(gm.ctx, &ai.MarketData{
			Items: gm.getMarketItems(),
		})
	}

	// Check for game events
	gm.checkGameEvents()
}

// initializeAIMerchants sets up AI merchants
func (gm *GameManager) initializeAIMerchants() {
	// Create various AI merchants with different personalities
	personalities := []ai.MerchantPersonality{
		ai.NewAggressivePersonality(),
		ai.NewConservativePersonality(),
		ai.NewBalancedPersonality(),
		ai.NewOpportunisticPersonality(),
	}

	names := []string{"Marcus", "Elena", "Thorin", "Luna"}

	for i, personality := range personalities {
		merchant := ai.NewAIMerchant(
			fmt.Sprintf("merchant_%d", i),
			names[i],
			1500, // Starting gold
			personality,
		)
		gm.aiSystem.AddMerchant(merchant)
	}
}

// GetGameState returns the current game state as JSON
func (gm *GameManager) GetGameState() (string, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	state := map[string]interface{}{
		"isRunning":  gm.isRunning,
		"isPaused":   gm.isPaused,
		"gold":       gm.gameState.GetGold(),
		"rank":       gm.gameState.GetRank(),
		"reputation": gm.gameState.GetReputation(),
		"day":        gm.timeManager.GetCurrentDay(),
		"season":     gm.timeManager.GetCurrentSeason(),
		"timeOfDay":  gm.timeManager.GetTimeOfDay(),
	}

	jsonData, err := json.Marshal(state)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// PauseGame pauses the game
func (gm *GameManager) PauseGame() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.isPaused = true
	gm.eventBus.Publish(event.Event{
		Type:      "GamePaused",
		Timestamp: time.Now(),
	})
}

// ResumeGame resumes the game
func (gm *GameManager) ResumeGame() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.isPaused = false
	gm.eventBus.Publish(event.Event{
		Type:      "GameResumed",
		Timestamp: time.Now(),
	})
}

// SaveGame saves the current game state
func (gm *GameManager) SaveGame(slot int) error {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	if gm.saveManager == nil {
		return fmt.Errorf("save system not available")
	}

	// Save the game
	err := gm.saveManager.SaveGame(
		slot,
		gm.gameState,
		gm.market,
		gm.inventory,
		gm.progression,
	)

	if err != nil {
		return fmt.Errorf("failed to save game: %w", err)
	}

	// Publish save event
	gm.eventBus.Publish(event.Event{
		Type:      "GameSaved",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"slot": slot,
		},
	})

	return nil
}

// LoadGame loads a saved game
func (gm *GameManager) LoadGame(slot int) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if gm.saveManager == nil {
		return fmt.Errorf("save system not available")
	}

	// Load the save data
	saveData, err := gm.saveManager.LoadGame(slot)
	if err != nil {
		return fmt.Errorf("failed to load game: %w", err)
	}

	// Restore game state
	gm.gameState = gamestate.NewGameState()
	gm.gameState.SetPlayerName(saveData.Player.Name)
	gm.gameState.SetGold(int(saveData.Player.Gold))
	gm.gameState.SetReputation(int(saveData.Player.Reputation))

	// Convert and set rank
	switch saveData.Player.Rank {
	case 1: // RANK_APPRENTICE
		gm.gameState.SetRank(gamestate.RankApprentice)
	case 2: // RANK_JOURNEYMAN
		gm.gameState.SetRank(gamestate.RankJourneyman)
	case 3: // RANK_VETERAN
		gm.gameState.SetRank(gamestate.RankVeteran)
	case 4: // RANK_MASTER
		gm.gameState.SetRank(gamestate.RankMaster)
	}

	// Restore inventory
	gm.inventory.Clear()
	for _, ownedItem := range saveData.Player.Inventory {
		if ownedItem.Location == 1 { // LOCATION_SHOP
			// TODO: Add to shop
		} else if ownedItem.Location == 2 { // LOCATION_WAREHOUSE
			gm.inventory.AddToWarehouse(ownedItem.ItemId, int(ownedItem.Quantity), 0)
		}
	}

	// Restore progression
	// TODO: Restore achievements and stats

	// Publish load event
	gm.eventBus.Publish(event.Event{
		Type:      "GameLoaded",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"slot": slot,
		},
	})

	return nil
}

// GetSaveSlots returns information about save slots
func (gm *GameManager) GetSaveSlots() (string, error) {
	if gm.saveManager == nil {
		return "[]", nil
	}

	slots, err := gm.saveManager.GetSaveSlots()
	if err != nil {
		return "[]", err
	}

	jsonData, err := json.Marshal(slots)
	if err != nil {
		return "[]", err
	}

	return string(jsonData), nil
}

// handleTimeAdvanced handles time advancement events
func (gm *GameManager) handleTimeAdvanced(e event.Event) {
	// Update market prices based on time
	if gm.market != nil {
		gm.market.UpdatePrices()
	}
}

// handleTradeCompleted handles trade completion events
func (gm *GameManager) handleTradeCompleted(e event.Event) {
	if data, ok := e.Data.(map[string]interface{}); ok {
		buyPrice := data["buyPrice"].(int)
		sellPrice := data["sellPrice"].(int)

		// Update progression
		result := gm.progression.HandleTradeCompletion(buyPrice, sellPrice)

		// Update game state
		if result.RankUp {
			gm.gameState.SetRank(gamestate.PlayerRank(result.NewRank))
		}
	}
}

// handleMarketPriceChanged handles market price change events
func (gm *GameManager) handleMarketPriceChanged(e event.Event) {
	// Notify UI of price changes
	// This would be sent to Godot
}

// getMarketItems returns current market items
func (gm *GameManager) getMarketItems() []ai.ItemData {
	items := []ai.ItemData{}

	if gm.market != nil {
		// Convert market items to AI item data
		// TODO: Implement proper conversion
	}

	return items
}

// GetMarketData returns current market data
func (gm *GameManager) GetMarketData() map[string]interface{} {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	marketItems := []map[string]interface{}{}

	if gm.market != nil {
		// TODO: Get actual market items
		// For now, return mock data
		marketItems = append(marketItems, map[string]interface{}{
			"id":           "apple",
			"name":         "Fresh Apple",
			"category":     "fruits",
			"basePrice":    10,
			"currentPrice": gm.market.GetPrice("apple"),
			"demand":       80,
			"supply":       100,
		})
	}

	return map[string]interface{}{
		"items": marketItems,
	}
}

// GetInventoryData returns current inventory data
func (gm *GameManager) GetInventoryData() map[string]interface{} {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	shopItems := []map[string]interface{}{}
	warehouseItems := []map[string]interface{}{}

	if gm.inventory != nil {
		// Get shop inventory
		shop := gm.inventory.GetShop()
		for _, item := range shop.GetItems() {
			shopItems = append(shopItems, map[string]interface{}{
				"id":       item.ID,
				"quantity": item.Quantity,
				"price":    item.Price,
			})
		}

		// Get warehouse inventory
		warehouse := gm.inventory.GetWarehouse()
		for _, item := range warehouse.GetItems() {
			warehouseItems = append(warehouseItems, map[string]interface{}{
				"id":       item.ID,
				"quantity": item.Quantity,
				"price":    item.Price,
			})
		}
	}

	return map[string]interface{}{
		"shop":              shopItems,
		"warehouse":         warehouseItems,
		"shopCapacity":      100,
		"warehouseCapacity": 200,
		"shopUsed":          len(shopItems),
		"warehouseUsed":     len(warehouseItems),
	}
}

// BuyItem handles item purchase
func (gm *GameManager) BuyItem(itemID string, quantity int, price float64) map[string]interface{} {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	totalCost := int(price * float64(quantity))
	currentGold := gm.gameState.GetGold()

	if currentGold < totalCost {
		return map[string]interface{}{
			"success": false,
			"message": "Insufficient gold",
		}
	}

	// Deduct gold
	gm.gameState.SetGold(currentGold - totalCost)

	// Add to inventory
	if gm.inventory != nil {
		gm.inventory.AddToWarehouse(itemID, quantity, int(price))
	}

	// Track with progression
	if gm.progression != nil {
		gm.progression.HandleTradeCompletion(int(price), 0)
	}

	return map[string]interface{}{
		"success":        true,
		"message":        "Item purchased",
		"gold_remaining": gm.gameState.GetGold(),
	}
}

// SellItem handles item sale
func (gm *GameManager) SellItem(itemID string, quantity int, price float64) map[string]interface{} {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Check if item exists in shop
	if gm.inventory != nil {
		shop := gm.inventory.GetShop()
		if !shop.HasItem(itemID, quantity) {
			return map[string]interface{}{
				"success": false,
				"message": "Insufficient quantity in shop",
			}
		}

		// Remove from shop
		shop.RemoveItem(itemID, quantity)
	}

	// Add gold
	totalGain := int(price * float64(quantity))
	gm.gameState.SetGold(gm.gameState.GetGold() + totalGain)

	// Track with progression
	if gm.progression != nil {
		gm.progression.HandleTradeCompletion(0, int(price))
	}

	return map[string]interface{}{
		"success":     true,
		"message":     "Item sold",
		"gold_gained": totalGain,
	}
}

// checkGameEvents checks for special game events
func (gm *GameManager) checkGameEvents() {
	// Check victory conditions
	if gm.gameState.GetGold() >= 100000 {
		gm.triggerVictory("wealth")
	}

	// Check defeat conditions
	if gm.gameState.GetGold() <= 0 && gm.inventory.IsEmpty() {
		gm.triggerDefeat("bankruptcy")
	}
}

// triggerVictory triggers a victory condition
func (gm *GameManager) triggerVictory(victoryType string) {
	gm.eventBus.Publish(event.Event{
		Type:      "GameVictory",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"type": victoryType,
		},
	})
}

// triggerDefeat triggers a defeat condition
func (gm *GameManager) triggerDefeat(defeatType string) {
	gm.eventBus.Publish(event.Event{
		Type:      "GameDefeat",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"type": defeatType,
		},
	})
}

// GetQueuedEvents returns all queued events for Godot
func (gm *GameManager) GetQueuedEvents() (string, error) {
	events := gm.eventBridge.FlushEvents()
	if len(events) == 0 {
		return "[]", nil
	}

	jsonData, err := json.Marshal(events)
	if err != nil {
		return "[]", err
	}

	return string(jsonData), nil
}

// SetEventCallback sets the callback for sending events to Godot
func (gm *GameManager) SetEventCallback(callback func(string, string)) {
	gm.eventBridge.SetGodotCallback(callback)
}

// Cleanup cleans up resources
func (gm *GameManager) Cleanup() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.isRunning = false
	gm.cancel()

	// Clean up systems
	if gm.aiSystem != nil {
		// AI system cleanup
	}
}
