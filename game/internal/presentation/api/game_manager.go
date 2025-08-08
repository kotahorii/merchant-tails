package api

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/calendar"
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
	timeManager timemanager.TimeManager
	eventBus    *event.EventBus
	eventBridge *EventBridge

	// Game systems
	market        *market.Market
	inventory     *inventory.InventoryManager
	trading       *trading.TradingSystem
	progression   *progression.ProgressionManager
	eventCalendar *calendar.EventCalendar

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
		gameState:   gamestate.NewGameState(nil), // Use default config
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
	tradingSystem, err := trading.NewTradingSystem(gm.inventory, nil)
	if err != nil {
		panic(err) // Should not happen with valid parameters
	}
	gm.trading = tradingSystem

	// Create progression manager
	gm.progression = progression.NewProgressionManager()

	// Create event calendar
	gm.eventCalendar = calendar.NewEventCalendar()

	// Register calendar callback to publish events
	gm.eventCalendar.RegisterCallback(func(calEvent *calendar.CalendarEvent, status string) {
		eventName := fmt.Sprintf("calendar.event.%s", status)
		gm.eventBus.PublishAsync(event.NewBaseEvent(eventName))
	})

	// AI system removed - single player only

	// Create save manager
	saveManager, err := persistence.NewSaveManager()
	if err != nil {
		// Log error but continue - save/load will be disabled
		fmt.Printf("Failed to initialize save manager: %v\n", err)
	}
	gm.saveManager = saveManager

	// Create game loop
	gm.gameLoop = gameloop.NewStandardGameLoop(&gameloop.Config{
		TargetFPS: 60,
	})

	// Create time manager
	gm.timeManager = timemanager.NewStandardTimeManager(gm.gameLoop, 5*time.Minute)

	// Register update callback
	gm.gameLoop.RegisterUpdateCallback(func(deltaTime time.Duration) error {
		if !gm.isPaused {
			gm.update(deltaTime.Seconds())
		}
		return nil
	})

	// Setup event listeners
	gm.setupEventListeners()
}

// setupEventListeners sets up event handlers
func (gm *GameManager) setupEventListeners() {
	// Listen for time events
	gm.eventBus.Subscribe("time.advanced", func(e event.Event) error {
		gm.handleTimeAdvanced(e)
		return nil
	})

	// Listen for trade events
	gm.eventBus.Subscribe(event.EventNameTransactionComplete, func(e event.Event) error {
		gm.handleTradeCompleted(e)
		return nil
	})

	// Listen for market events
	gm.eventBus.Subscribe(event.EventNamePriceUpdated, func(e event.Event) error {
		gm.handleMarketPriceChanged(e)
		return nil
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
	gm.gameState = gamestate.NewGameState(nil)
	// TODO: Add player name support to GameState
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
	gm.eventBus.PublishAsync(event.NewBaseEvent("GameStarted"))

	return nil
}

// runGameLoop runs the main game loop
func (gm *GameManager) runGameLoop() {
	err := gm.gameLoop.Start(gm.ctx)
	if err != nil {
		// Log error but continue
		fmt.Printf("Game loop error: %v\n", err)
	}
}

// update updates all game systems
func (gm *GameManager) update(deltaTime float64) {
	// Update time
	duration := time.Duration(deltaTime * float64(time.Second))
	gm.timeManager.Update(duration)

	// Update market
	if gm.market != nil {
		gm.market.Update()
	}

	// AI system removed - single player only

	// Check for game events
	gm.checkGameEvents()
}

// initializeAIMerchants removed - single player only
func (gm *GameManager) initializeAIMerchants() {
	// AI merchants removed for single player focus
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
		"time":       gm.timeManager.GetCurrentTime(),
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
	gm.eventBus.PublishAsync(event.NewBaseEvent("GamePaused"))
}

// ResumeGame resumes the game
func (gm *GameManager) ResumeGame() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.isPaused = false
	gm.eventBus.PublishAsync(event.NewBaseEvent("GameResumed"))
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
	gm.eventBus.PublishAsync(event.NewBaseEvent("GameSaved"))

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
	gm.gameState = gamestate.NewGameState(nil)
	// TODO: Add player name support to GameState
	gm.gameState.SetGold(int(saveData["gold"].(float64)))
	gm.gameState.SetReputation(saveData["reputation"].(float64))

	// Convert and set rank
	if rank, ok := saveData["rank"].(float64); ok {
		switch int(rank) {
		case 1: // RANK_APPRENTICE
			gm.gameState.SetRank(gamestate.RankApprentice)
		case 2: // RANK_JOURNEYMAN
			gm.gameState.SetRank(gamestate.RankJourneyman)
		case 3: // RANK_EXPERT
			gm.gameState.SetRank(gamestate.RankExpert)
		case 4: // RANK_MASTER
			gm.gameState.SetRank(gamestate.RankMaster)
		}
	}

	// Restore inventory
	gm.inventory.Clear()
	// TODO: Implement inventory restoration from save data

	// Restore progression
	// TODO: Restore achievements and stats

	// Publish load event
	gm.eventBus.PublishAsync(event.NewBaseEvent("GameLoaded"))

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

	// Update event calendar
	if gm.eventCalendar != nil {
		gm.eventCalendar.UpdateDate(time.Now())
	}
}

// handleTradeCompleted handles trade completion events
func (gm *GameManager) handleTradeCompleted(e event.Event) {
	// TODO: Extract trade data from event
	// For now, just log that trade was completed
	if gm.progression != nil {
		// Update progression with dummy values for now
		result := gm.progression.HandleTradeCompletion(0, 0)

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

// getMarketItems removed - AI system no longer needed
func (gm *GameManager) getMarketItems() []interface{} {
	return []interface{}{}
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
				"id":       item.Item.ID,
				"quantity": item.Quantity,
				"price":    item.Item.BasePrice,
			})
		}

		// Get warehouse inventory
		warehouse := gm.inventory.GetWarehouse()
		for _, item := range warehouse.GetItems() {
			warehouseItems = append(warehouseItems, map[string]interface{}{
				"id":       item.Item.ID,
				"quantity": item.Quantity,
				"price":    item.Item.BasePrice,
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
		gm.inventory.AddToWarehouseByID(itemID, quantity, int(price))
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
	gm.eventBus.PublishAsync(event.NewBaseEvent("GameVictory"))
}

// triggerDefeat triggers a defeat condition
func (gm *GameManager) triggerDefeat(defeatType string) {
	gm.eventBus.PublishAsync(event.NewBaseEvent("GameDefeat"))
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

// GetUpcomingEvents returns upcoming calendar events
func (gm *GameManager) GetUpcomingEvents(days int) (string, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	if gm.eventCalendar == nil {
		return "[]", nil
	}

	events := gm.eventCalendar.GetUpcomingEvents(days)

	// Convert to simpler format for Godot
	simpleEvents := make([]map[string]interface{}, 0, len(events))
	for _, event := range events {
		simpleEvents = append(simpleEvents, map[string]interface{}{
			"id":          event.ID,
			"name":        event.Name,
			"description": event.Description,
			"type":        int(event.Type),
			"priority":    int(event.Priority),
			"startDate":   event.StartDate.Format(time.RFC3339),
			"endDate":     event.EndDate.Format(time.RFC3339),
			"effects":     event.Effects,
			"rewards":     event.Rewards,
		})
	}

	jsonData, err := json.Marshal(simpleEvents)
	if err != nil {
		return "[]", err
	}

	return string(jsonData), nil
}

// GetActiveEvents returns currently active calendar events
func (gm *GameManager) GetActiveEvents() (string, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	if gm.eventCalendar == nil {
		return "[]", nil
	}

	events := gm.eventCalendar.GetActiveEvents()

	// Convert to simpler format for Godot
	simpleEvents := make([]map[string]interface{}, 0, len(events))
	for _, event := range events {
		simpleEvents = append(simpleEvents, map[string]interface{}{
			"id":          event.ID,
			"name":        event.Name,
			"description": event.Description,
			"type":        int(event.Type),
			"priority":    int(event.Priority),
			"effects":     event.Effects,
			"rewards":     event.Rewards,
		})
	}

	jsonData, err := json.Marshal(simpleEvents)
	if err != nil {
		return "[]", err
	}

	return string(jsonData), nil
}

// GetEventEffects returns the combined effects of all active events
func (gm *GameManager) GetEventEffects() (string, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	if gm.eventCalendar == nil {
		return "{}", nil
	}

	effects := gm.eventCalendar.GetEventEffects()

	jsonData, err := json.Marshal(effects)
	if err != nil {
		return "{}", err
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
	// AI system removed - single player only
}
