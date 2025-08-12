package api

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/event"
	"github.com/yourusername/merchant-tails/game/internal/domain/gameloop"
	"github.com/yourusername/merchant-tails/game/internal/domain/gamestate"
	"github.com/yourusername/merchant-tails/game/internal/domain/inventory"
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
	"github.com/yourusername/merchant-tails/game/internal/domain/progression"
	"github.com/yourusername/merchant-tails/game/internal/domain/settings"
	timemanager "github.com/yourusername/merchant-tails/game/internal/domain/time"
	"github.com/yourusername/merchant-tails/game/internal/infrastructure/logging"
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
	market      *market.Market
	inventory   *inventory.InventoryManager
	progression *progression.ProgressionManager

	// Infrastructure
	saveManager *persistence.SaveManager
	settings    *settings.SettingsManager

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
	// Create settings manager
	gm.settings = settings.NewSettingsManager("./config/settings.json")
	if err := gm.settings.LoadSettings(); err != nil {
		// Log error but continue with defaults
		logging.Warnf("Failed to load settings: %v", err)
	}

	// Create market
	gm.market = market.NewMarket()

	// Get capacity from settings
	gameSettings := gm.settings.GetSettings()
	shopCap := gameSettings.CustomSettings["shopCapacity"]
	warehouseCap := gameSettings.CustomSettings["warehouseCapacity"]

	shopCapacity := 100
	warehouseCapacity := 200

	if sc, ok := shopCap.(int); ok {
		shopCapacity = sc
	}
	if wc, ok := warehouseCap.(int); ok {
		warehouseCapacity = wc
	}

	// Create inventory manager with settings-based capacity
	invManager, err := inventory.NewInventoryManager(shopCapacity, warehouseCapacity)
	if err != nil {
		panic(err) // Should not happen with valid capacities
	}
	gm.inventory = invManager

	// Create progression manager
	gm.progression = progression.NewProgressionManager()

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
	// Set player name
	if err := gm.gameState.SetPlayerName(playerName); err != nil {
		return fmt.Errorf("failed to set player name: %w", err)
	}
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

	// Log game start
	logging.Infof("Game started - Gold: %d, Day: %d, Reputation: %.2f",
		gm.gameState.GetGold(), gm.gameState.GetCurrentDay(), gm.gameState.GetReputation())

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
	updateStart := time.Now()

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

	// Metrics removed - too complex
	_ = updateStart
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
		"isRunning":     gm.isRunning,
		"isPaused":      gm.isPaused,
		"playerName":    gm.gameState.GetPlayerName(),
		"gold":          gm.gameState.GetGold(),
		"rank":          gm.gameState.GetRank(),
		"rankName":      gamestate.GetRankName(gm.gameState.GetRank()),
		"rankProgress":  gm.gameState.GetRankProgress(),
		"reputation":    gm.gameState.GetReputation(),
		"currentDay":    gm.gameState.GetCurrentDay(),
		"currentSeason": gm.gameState.GetCurrentSeason(),
		"time":          gm.timeManager.GetCurrentTime(),
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
	// Restore player name
	if playerName, ok := saveData["playerName"].(string); ok && playerName != "" {
		_ = gm.gameState.SetPlayerName(playerName)
	}
	// Restore day and season
	if currentDay, ok := saveData["currentDay"].(float64); ok {
		gm.gameState.SetCurrentDay(int(currentDay))
	}
	if currentSeason, ok := saveData["currentSeason"].(string); ok {
		_ = gm.gameState.SetCurrentSeason(currentSeason)
	}
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
	// Advance game day
	gm.gameState.AdvanceDay()

	// Check for rank up
	if gm.gameState.CheckRankUp() {
		logging.Infof("Player ranked up to %s!", gamestate.GetRankName(gm.gameState.GetRank()))
		gm.eventBus.PublishAsync(event.NewBaseEvent("RankUp"))
	}

	// Update market prices based on time and season
	if gm.market != nil {
		gm.market.UpdatePrices()
	}

	// Event calendar removed - too complex
}

// handleTradeCompleted handles trade completion events
func (gm *GameManager) handleTradeCompleted(e event.Event) {
	// Extract trade data from event
	revenue := 0.0
	success := true
	transactionID := fmt.Sprintf("trans-%d", time.Now().Unix())

	// Log transaction
	logging.LogTransaction(transactionID, "unknown", 1, revenue, 0.0)

	// Metrics removed
	_ = success
	_ = revenue

	if gm.progression != nil {
		// Update progression with dummy values for now
		result := gm.progression.HandleTradeCompletion(0, 0)

		// Update game state
		if result.RankUp {
			gm.gameState.SetRank(gamestate.PlayerRank(result.NewRank))
			logging.Infof("Rank up - Gold: %d, Day: %d, Reputation: %.2f",
				gm.gameState.GetGold(), gm.gameState.GetCurrentDay(), gm.gameState.GetReputation())
		}
	}
}

// handleMarketPriceChanged handles market price change events
func (gm *GameManager) handleMarketPriceChanged(e event.Event) {
	// Update market prices for all items
	if gm.market != nil {
		for _, itemID := range []string{"apple", "potion", "sword"} {
			oldPrice := float64(gm.market.GetPrice(itemID))
			gm.market.UpdatePrice(itemID)
			newPrice := float64(gm.market.GetPrice(itemID))

			// Log market event
			impact := (newPrice - oldPrice) / oldPrice * 100
			logging.Infof("Market price change - Item: %s, Old: %.2f, New: %.2f, Impact: %.2f%%",
				itemID, oldPrice, newPrice, impact)
		}
	}

	// Notify UI of price changes
	// This would be sent to Godot
}

// getMarketItems removed - AI system no longer needed

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
	shopCapacity := 100
	warehouseCapacity := 200
	shopUsed := 0
	warehouseUsed := 0

	if gm.inventory != nil {
		// Update capacity utilization tracking
		gm.inventory.UpdateCapacity()

		// Get capacity stats
		capacityStats := gm.inventory.GetCapacityStats()
		if capacityStats != nil {
			shopCapacity = capacityStats.CurrentShopCapacity
			warehouseCapacity = capacityStats.CurrentWarehouseCapacity
		}

		// Get shop inventory
		shop := gm.inventory.GetShop()
		for itemID, quantity := range shop.GetAll() {
			price := 100 // Default price, should get from market
			if gm.market != nil {
				price = gm.market.GetPrice(itemID)
			}
			shopItems = append(shopItems, map[string]interface{}{
				"id":       itemID,
				"quantity": quantity,
				"price":    price,
			})
			shopUsed += quantity
		}

		// Get warehouse inventory
		warehouse := gm.inventory.GetWarehouse()
		for itemID, quantity := range warehouse.GetAll() {
			price := 100 // Default price, should get from market
			if gm.market != nil {
				price = gm.market.GetPrice(itemID)
			}
			warehouseItems = append(warehouseItems, map[string]interface{}{
				"id":       itemID,
				"quantity": quantity,
				"price":    price,
			})
			warehouseUsed += quantity
		}
	}

	result := map[string]interface{}{
		locationShop:           shopItems,
		locationWarehouse:      warehouseItems,
		"shopCapacity":         shopCapacity,
		"warehouseCapacity":    warehouseCapacity,
		"shopUsed":             shopUsed,
		"warehouseUsed":        warehouseUsed,
		"shopUtilization":      float64(shopUsed) / float64(shopCapacity),
		"warehouseUtilization": float64(warehouseUsed) / float64(warehouseCapacity),
	}

	// Add capacity alerts if any
	if gm.inventory != nil {
		alerts := gm.inventory.GetCapacityAlerts()
		if len(alerts) > 0 {
			alertMaps := []map[string]interface{}{}
			for _, alert := range alerts {
				alertMaps = append(alertMaps, map[string]interface{}{
					"type":        alert.Type,
					"location":    alert.Location,
					"severity":    alert.Severity,
					"message":     alert.Message,
					"utilization": alert.Utilization,
				})
			}
			result["capacityAlerts"] = alertMaps
		}
	}

	return result
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
		_ = gm.inventory.AddToWarehouseByID(itemID, quantity, int(price))
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
		_ = shop.RemoveItem(itemID, quantity)
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

// GetUpcomingEvents returns upcoming events (simplified)
func (gm *GameManager) GetUpcomingEvents(days int) (string, error) {
	// Event calendar removed - too complex
	return "[]", nil
}

// GetActiveEvents returns currently active events (simplified)
func (gm *GameManager) GetActiveEvents() (string, error) {
	// Event calendar removed - too complex
	return "[]", nil
}

// GetEventEffects returns the combined effects of all active events (simplified)
func (gm *GameManager) GetEventEffects() (string, error) {
	// Event calendar removed - too complex
	return "{}", nil
}

// SetEventCallback sets the callback for sending events to Godot
func (gm *GameManager) SetEventCallback(callback func(string, string)) {
	gm.eventBridge.SetGodotCallback(callback)
}

// GetPlayerInfo returns detailed player information
func (gm *GameManager) GetPlayerInfo() map[string]interface{} {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	// Get rank bonuses
	shopBonus, warehouseBonus, priceDiscount := gm.gameState.GetRankBonus()

	return map[string]interface{}{
		"name":              gm.gameState.GetPlayerName(),
		"rank":              gm.gameState.GetRank(),
		"rankName":          gamestate.GetRankName(gm.gameState.GetRank()),
		"rankProgress":      gm.gameState.GetRankProgress(),
		"gold":              gm.gameState.GetGold(),
		"reputation":        gm.gameState.GetReputation(),
		"currentDay":        gm.gameState.GetCurrentDay(),
		"currentSeason":     gm.gameState.GetCurrentSeason(),
		"totalTransactions": gm.gameState.GetTotalTransactions(),
		"totalProfit":       gm.gameState.GetTotalProfit(),
		"profitMargin":      gm.gameState.GetProfitMargin(),
		"shopCapacity":      gm.gameState.GetShopCapacity(),
		"warehouseCapacity": gm.gameState.GetWarehouseCapacity(),
		"rankBonuses": map[string]interface{}{
			"shopCapacityBonus":      shopBonus,
			"warehouseCapacityBonus": warehouseBonus,
			"priceDiscount":          priceDiscount,
		},
	}
}

// UpdatePlayerName updates the player's name
func (gm *GameManager) UpdatePlayerName(name string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	return gm.gameState.SetPlayerName(name)
}

// AdvanceTime advances the game time by specified number of days
func (gm *GameManager) AdvanceTime(days int) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	for i := 0; i < days; i++ {
		gm.gameState.AdvanceDay()

		// Check for rank up after each day
		if gm.gameState.CheckRankUp() {
			logging.Infof("Player ranked up to %s!", gamestate.GetRankName(gm.gameState.GetRank()))
			gm.eventBus.PublishAsync(event.NewBaseEvent("RankUp"))
		}
	}

	// Update market and events
	if gm.market != nil {
		gm.market.UpdatePrices()
	}
}

// UpgradeInventoryCapacity upgrades shop or warehouse capacity
func (gm *GameManager) UpgradeInventoryCapacity(location string, amount int, cost int) map[string]interface{} {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Check if player has enough gold
	currentGold := gm.gameState.GetGold()
	if currentGold < cost {
		return map[string]interface{}{
			"success":   false,
			"message":   "Insufficient gold for upgrade",
			"required":  cost,
			"available": currentGold,
		}
	}

	// Determine location
	var invLocation inventory.InventoryLocation
	if location == locationShop {
		invLocation = inventory.LocationShop
	} else if location == locationWarehouse {
		invLocation = inventory.LocationWarehouse
	} else {
		return map[string]interface{}{
			"success": false,
			"message": "Invalid location. Must be 'shop' or 'warehouse'",
		}
	}

	// Perform upgrade
	if gm.inventory != nil {
		err := gm.inventory.UpgradeCapacity(invLocation, amount)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"message": err.Error(),
			}
		}

		// Deduct gold
		gm.gameState.SetGold(currentGold - cost)

		// Get new capacity stats
		stats := gm.inventory.GetCapacityStats()
		newCapacity := 0
		if invLocation == inventory.LocationShop && stats != nil {
			newCapacity = stats.CurrentShopCapacity
		} else if stats != nil {
			newCapacity = stats.CurrentWarehouseCapacity
		}

		// Metrics removed

		return map[string]interface{}{
			"success":       true,
			"message":       fmt.Sprintf("Successfully upgraded %s capacity", location),
			"newCapacity":   newCapacity,
			"goldRemaining": gm.gameState.GetGold(),
		}
	}

	return map[string]interface{}{
		"success": false,
		"message": "Inventory system not initialized",
	}
}

// OptimizeInventory optimizes inventory placement
func (gm *GameManager) OptimizeInventory() map[string]interface{} {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if gm.inventory == nil {
		return map[string]interface{}{
			"success": false,
			"message": "Inventory system not initialized",
		}
	}

	// Perform optimization
	err := gm.inventory.OptimizeCapacityUsage()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	// Also optimize placement based on sales velocity
	gm.inventory.OptimizePlacement()

	return map[string]interface{}{
		"success": true,
		"message": "Inventory optimized successfully",
	}
}

// GetCapacityRecommendations returns capacity upgrade recommendations
func (gm *GameManager) GetCapacityRecommendations() map[string]interface{} {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	if gm.inventory == nil {
		return map[string]interface{}{
			"hasRecommendations": false,
		}
	}

	stats := gm.inventory.GetCapacityStats()
	if stats == nil {
		return map[string]interface{}{
			"hasRecommendations": false,
		}
	}

	recommendations := []map[string]interface{}{}

	// Check if shop needs upgrade
	if stats.PeakShopUtilization > 0.8 {
		upgradeAmount := stats.RecommendedShopCapacity - stats.CurrentShopCapacity
		if upgradeAmount > 0 {
			recommendations = append(recommendations, map[string]interface{}{
				"location":        locationShop,
				"currentCapacity": stats.CurrentShopCapacity,
				"recommended":     stats.RecommendedShopCapacity,
				"upgradeAmount":   upgradeAmount,
				"reason":          fmt.Sprintf("Peak utilization reached %.1f%%", stats.PeakShopUtilization*100),
				"estimatedCost":   upgradeAmount * 50, // 50 gold per capacity unit
			})
		}
	}

	// Check if warehouse needs upgrade
	if stats.PeakWarehouseUtilization > 0.8 {
		upgradeAmount := stats.RecommendedWarehouseCapacity - stats.CurrentWarehouseCapacity
		if upgradeAmount > 0 {
			recommendations = append(recommendations, map[string]interface{}{
				"location":        locationWarehouse,
				"currentCapacity": stats.CurrentWarehouseCapacity,
				"recommended":     stats.RecommendedWarehouseCapacity,
				"upgradeAmount":   upgradeAmount,
				"reason":          fmt.Sprintf("Peak utilization reached %.1f%%", stats.PeakWarehouseUtilization*100),
				"estimatedCost":   upgradeAmount * 20, // 20 gold per capacity unit
			})
		}
	}

	return map[string]interface{}{
		"hasRecommendations": len(recommendations) > 0,
		"recommendations":    recommendations,
		"statistics": map[string]interface{}{
			"avgShopUtilization":       stats.AvgShopUtilization,
			"avgWarehouseUtilization":  stats.AvgWarehouseUtilization,
			"peakShopUtilization":      stats.PeakShopUtilization,
			"peakWarehouseUtilization": stats.PeakWarehouseUtilization,
		},
	}
}

// GetSeasonalEffects returns the current seasonal effects on the market
func (gm *GameManager) GetSeasonalEffects() map[string]interface{} {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	season := gm.gameState.GetCurrentSeason()
	effects := make(map[string]interface{})

	// Define seasonal effects
	switch season {
	case "Spring":
		effects["description"] = "Spring: Fresh produce is abundant"
		effects["priceModifiers"] = map[string]float64{
			"fruits":     0.9,  // 10% cheaper
			"vegetables": 0.85, // 15% cheaper
			"weapons":    1.0,
			"potions":    1.05, // 5% more expensive
		}
	case "Summer":
		effects["description"] = "Summer: Travel season increases demand for supplies"
		effects["priceModifiers"] = map[string]float64{
			"fruits":     0.95,
			"vegetables": 0.9,
			"weapons":    1.1,  // 10% more expensive
			"potions":    1.15, // 15% more expensive
		}
	case "Autumn":
		effects["description"] = "Autumn: Harvest season brings plenty"
		effects["priceModifiers"] = map[string]float64{
			"fruits":     0.8,  // 20% cheaper
			"vegetables": 0.75, // 25% cheaper
			"weapons":    1.05,
			"potions":    1.0,
		}
	case "Winter":
		effects["description"] = "Winter: Scarcity drives prices up"
		effects["priceModifiers"] = map[string]float64{
			"fruits":     1.3, // 30% more expensive
			"vegetables": 1.4, // 40% more expensive
			"weapons":    0.95,
			"potions":    1.2, // 20% more expensive
		}
	}

	effects["currentSeason"] = season
	effects["currentDay"] = gm.gameState.GetCurrentDay()

	return effects
}

// GetSettings returns current game settings
func (gm *GameManager) GetSettings() map[string]interface{} {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	if gm.settings == nil {
		return map[string]interface{}{}
	}

	gameSettings := gm.settings.GetSettings()
	return map[string]interface{}{
		"game": map[string]interface{}{
			"difficulty":       gameSettings.Difficulty,
			"autoSave":         gameSettings.AutoSave,
			"autoSaveInterval": gameSettings.AutoSaveInterval,
			"language":         gameSettings.Language,
			"pauseOnFocusLoss": gameSettings.PauseOnFocusLoss,
		},
		"graphics": map[string]interface{}{
			"resolution":     gameSettings.Resolution,
			"fullscreen":     gameSettings.Fullscreen,
			"vsync":          gameSettings.VSync,
			"targetFPS":      gameSettings.TargetFPS,
			"shadowQuality":  gameSettings.ShadowQuality,
			"textureQuality": gameSettings.TextureQuality,
			"effectsQuality": gameSettings.EffectsQuality,
		},
		"audio": map[string]interface{}{
			"masterVolume": gameSettings.MasterVolume,
			"musicVolume":  gameSettings.MusicVolume,
			"sfxVolume":    gameSettings.SFXVolume,
			"uiVolume":     gameSettings.UIVolume,
		},
		"ui": map[string]interface{}{
			"showFPS":           gameSettings.ShowFPS,
			"showNotifications": gameSettings.ShowNotifications,
			"showTutorialHints": gameSettings.ShowTutorialHints,
			"showTooltips":      gameSettings.ShowTooltips,
		},
	}
}

// UpdateSettings updates game settings with validation
func (gm *GameManager) UpdateSettings(category string, updates map[string]interface{}) map[string]interface{} {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if gm.settings == nil {
		return map[string]interface{}{
			"success": false,
			"message": "Settings manager not initialized",
		}
	}

	// Apply updates based on category
	var errors []string
	for key, value := range updates {
		fullKey := fmt.Sprintf("%s_%s", category, key)
		if err := gm.settings.SetSetting(fullKey, value); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", key, err))
		}
	}

	if len(errors) > 0 {
		return map[string]interface{}{
			"success": false,
			"errors":  errors,
		}
	}

	// Save settings
	if err := gm.settings.SaveSettings(); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to save settings: %v", err),
		}
	}

	// Apply certain settings immediately
	gm.applySettingsChanges(category, updates)

	return map[string]interface{}{
		"success": true,
		"message": "Settings updated successfully",
	}
}

// applySettingsChanges applies settings changes that need immediate effect
func (gm *GameManager) applySettingsChanges(category string, updates map[string]interface{}) {
	switch category {
	case "audio":
		// Audio changes would be applied to the audio system
		if volume, ok := updates["musicVolume"].(float64); ok {
			logging.Infof("Music volume changed to: %.2f", volume)
		}
	case "game":
		// Game settings changes
		if autoSave, ok := updates["autoSave"].(bool); ok && gm.saveManager != nil {
			if autoSave {
				logging.Infof("Auto-save enabled")
			} else {
				logging.Infof("Auto-save disabled")
			}
		}
	case "graphics":
		// Graphics settings would trigger rendering updates
		if fps, ok := updates["targetFPS"].(int); ok && gm.gameLoop != nil {
			logging.Infof("Target FPS changed to: %d", fps)
		}
	}
}

// ResetSettings resets settings to defaults
func (gm *GameManager) ResetSettings(category string) map[string]interface{} {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if gm.settings == nil {
		return map[string]interface{}{
			"success": false,
			"message": "Settings manager not initialized",
		}
	}

	var err error
	if category == categoryAll {
		err = gm.settings.ResetToDefaults()
	} else {
		categoryMap := map[string]settings.SettingsCategory{
			"game":     settings.CategoryGame,
			"graphics": settings.CategoryGraphics,
			"audio":    settings.CategoryAudio,
			"controls": settings.CategoryControls,
			"ui":       settings.CategoryUI,
			"advanced": settings.CategoryAdvanced,
		}

		if cat, ok := categoryMap[category]; ok {
			err = gm.settings.ResetCategory(cat)
		} else {
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("Unknown category: %s", category),
			}
		}
	}

	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to reset settings: %v", err),
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Settings reset to defaults for category: %s", category),
	}
}

// ValidateSettings validates settings before applying
func (gm *GameManager) ValidateSettings(updates map[string]interface{}) map[string]interface{} {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	// Create a validator
	validator := settings.NewValidator()

	// Validate the updates
	result := validator.ValidatePartial(updates, getSettingKeys(updates))

	if result.Valid {
		return map[string]interface{}{
			"valid": true,
		}
	}

	// Return validation errors
	errors := make([]map[string]interface{}, 0)
	for _, err := range result.Errors {
		errors = append(errors, map[string]interface{}{
			"field":   err.Field,
			"message": err.Message,
			"value":   err.Value,
		})
	}

	return map[string]interface{}{
		"valid":  false,
		"errors": errors,
	}
}

// getSettingKeys extracts keys from settings map
func getSettingKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Cleanup cleans up resources
func (gm *GameManager) Cleanup() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Save settings before cleanup
	if gm.settings != nil {
		_ = gm.settings.SaveSettings()
	}

	gm.isRunning = false
	gm.cancel()

	// Clean up systems
	// AI system removed - single player only
}
