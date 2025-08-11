package monitoring

import (
	"runtime"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/gamestate"
	"github.com/yourusername/merchant-tails/game/internal/domain/inventory"
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
)

// MetricsMiddleware provides integration between game systems and metrics
type MetricsMiddleware struct {
	collector *MetricsCollector
	gameState *gamestate.GameState
	market    *market.Market
	inventory *inventory.InventoryManager
	ticker    *time.Ticker
	stopChan  chan bool
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(
	collector *MetricsCollector,
	gameState *gamestate.GameState,
	market *market.Market,
	inventory *inventory.InventoryManager,
) *MetricsMiddleware {
	return &MetricsMiddleware{
		collector: collector,
		gameState: gameState,
		market:    market,
		inventory: inventory,
		stopChan:  make(chan bool),
	}
}

// Start begins collecting metrics periodically
func (mm *MetricsMiddleware) Start(interval time.Duration) {
	mm.ticker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-mm.ticker.C:
				mm.collectMetrics()
			case <-mm.stopChan:
				return
			}
		}
	}()
}

// Stop stops the metrics collection
func (mm *MetricsMiddleware) Stop() {
	if mm.ticker != nil {
		mm.ticker.Stop()
	}
	mm.stopChan <- true
}

// collectMetrics gathers metrics from all game systems
func (mm *MetricsMiddleware) collectMetrics() {
	// Player metrics
	if mm.gameState != nil {
		mm.collector.UpdatePlayerMetrics(
			mm.gameState.GetGold(),
			int(mm.gameState.GetPlayerRank()),
			mm.gameState.GetReputation(),
		)
	}

	// Inventory metrics
	if mm.inventory != nil {
		shopItems := mm.inventory.GetTotalShopItems()
		shopInv := mm.inventory.GetShop()
		totalValue := 0.0
		if shopInv != nil {
			items := shopInv.GetAll()
			for itemID, quantity := range items {
				// Get item price from market
				if mm.market != nil {
					price := mm.market.GetPrice(itemID)
					totalValue += float64(price * quantity)
				}
			}
		}
		mm.collector.UpdateInventoryMetrics(shopItems, totalValue)
	}

	// Market metrics
	if mm.market != nil {
		items := mm.market.GetAllItems()
		for _, item := range items {
			price := mm.market.GetPrice(item.ID)
			mm.collector.UpdateMarketPrice(item.ID, float64(price))

			// Calculate volatility
			history := mm.market.GetPriceHistory(item.ID)
			if history != nil {
				volatility := mm.calculateVolatility(history)
				mm.collector.UpdateMarketVolatility(item.ID, volatility)
			}
		}
	}

	// System metrics
	mm.collectSystemMetrics()
}

// calculateVolatility calculates price volatility from history
func (mm *MetricsMiddleware) calculateVolatility(history *market.PriceHistory) float64 {
	if history == nil || len(history.Records) < 2 {
		return 0
	}

	// Calculate standard deviation of price changes
	var sum, sumSq float64
	var count int

	for i := 1; i < len(history.Records); i++ {
		change := float64(history.Records[i].Price - history.Records[i-1].Price)
		sum += change
		sumSq += change * change
		count++
	}

	if count == 0 {
		return 0
	}

	mean := sum / float64(count)
	variance := (sumSq / float64(count)) - (mean * mean)

	if variance < 0 {
		return 0
	}

	return variance // Return variance as volatility measure
}

// collectSystemMetrics collects system-level metrics
func (mm *MetricsMiddleware) collectSystemMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Memory metrics
	mm.collector.UpdateMemoryUsage(m.Alloc)

	// Goroutine count
	mm.collector.UpdateGoroutineCount(runtime.NumGoroutine())

	// GC metrics
	if m.NumGC > 0 {
		// Get the most recent GC pause
		lastGC := m.PauseNs[(m.NumGC+255)%256]
		mm.collector.RecordGCPause(time.Duration(lastGC))
	}
}

// RecordTransaction records a transaction with metrics
func (mm *MetricsMiddleware) RecordTransaction(success bool, revenue float64) {
	mm.collector.RecordTransaction(success, revenue)
}

// RecordFrameTime records frame rendering time
func (mm *MetricsMiddleware) RecordFrameTime(duration time.Duration) {
	mm.collector.RecordFrameTime(duration)
}

// RecordSaveOperation records save operation metrics
func (mm *MetricsMiddleware) RecordSaveOperation(success bool, duration time.Duration) {
	mm.collector.RecordSaveOperation(success, duration)
}

// RecordLoadOperation records load operation metrics
func (mm *MetricsMiddleware) RecordLoadOperation(success bool, duration time.Duration) {
	mm.collector.RecordLoadOperation(success, duration)
}

// RecordEvent records game event processing
func (mm *MetricsMiddleware) RecordEvent(eventType string, duration time.Duration) {
	mm.collector.RecordEvent(eventType, duration)
}

// UpdateEventQueueSize updates event queue size metric
func (mm *MetricsMiddleware) UpdateEventQueueSize(size int) {
	mm.collector.UpdateEventQueueSize(size)
}

// RecordQuestCompletion records quest completion
func (mm *MetricsMiddleware) RecordQuestCompletion() {
	mm.collector.RecordQuestCompletion()
}

// UpdateActiveQuests updates active quest count
func (mm *MetricsMiddleware) UpdateActiveQuests(count int) {
	mm.collector.UpdateActiveQuests(count)
}

// RecordAchievementUnlock records achievement unlock
func (mm *MetricsMiddleware) RecordAchievementUnlock() {
	mm.collector.RecordAchievementUnlock()
}

// UpdateAchievementProgress updates achievement progress
func (mm *MetricsMiddleware) UpdateAchievementProgress(achievementID string, progress float64) {
	mm.collector.UpdateAchievementProgress(achievementID, progress)
}
