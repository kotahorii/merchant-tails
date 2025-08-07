package progression

import (
	"sync"
	"time"
)

// PlayerStats tracks player statistics
type PlayerStats struct {
	totalTrades      int
	profitableTrades int
	totalProfit      int
	totalLoss        int
	bestTrade        int
	worstTrade       int
	totalGoldEarned  int
	totalGoldSpent   int
	itemStats        map[string]*ItemTradeStats
	dailyStats       map[string]*DailyStats
	playTime         time.Duration
	gameStartTime    time.Time
	lastPlayedTime   time.Time
	mu               sync.RWMutex
}

// ItemTradeStats tracks statistics for a specific item
type ItemTradeStats struct {
	ItemID      string
	TradeCount  int
	TotalProfit int
	TotalLoss   int
	BestDeal    int
	WorstDeal   int
}

// DailyStats tracks daily trading statistics
type DailyStats struct {
	Date       string
	TradeCount int
	Profit     int
	GoldEarned int
	GoldSpent  int
}

// NewPlayerStats creates a new player statistics tracker
func NewPlayerStats() *PlayerStats {
	return &PlayerStats{
		totalTrades:      0,
		profitableTrades: 0,
		totalProfit:      0,
		totalLoss:        0,
		bestTrade:        0,
		worstTrade:       0,
		totalGoldEarned:  0,
		totalGoldSpent:   0,
		itemStats:        make(map[string]*ItemTradeStats),
		dailyStats:       make(map[string]*DailyStats),
		gameStartTime:    time.Now(),
		lastPlayedTime:   time.Now(),
	}
}

// RecordTrade records a trade transaction
func (ps *PlayerStats) RecordTrade(buyPrice, sellPrice int) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.totalTrades++
	profit := sellPrice - buyPrice

	if profit > 0 {
		ps.profitableTrades++
		ps.totalProfit += profit
		if profit > ps.bestTrade {
			ps.bestTrade = profit
		}
	} else {
		ps.totalLoss += -profit
		if profit < ps.worstTrade {
			ps.worstTrade = profit
		}
	}

	ps.totalGoldSpent += buyPrice
	ps.totalGoldEarned += sellPrice

	// Update daily stats
	today := time.Now().Format("2006-01-02")
	if daily, exists := ps.dailyStats[today]; exists {
		daily.TradeCount++
		daily.Profit += profit
		daily.GoldEarned += sellPrice
		daily.GoldSpent += buyPrice
	} else {
		ps.dailyStats[today] = &DailyStats{
			Date:       today,
			TradeCount: 1,
			Profit:     profit,
			GoldEarned: sellPrice,
			GoldSpent:  buyPrice,
		}
	}
}

// RecordItemTrade records a trade for a specific item
func (ps *PlayerStats) RecordItemTrade(itemID string, buyPrice, sellPrice int) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	profit := sellPrice - buyPrice

	// Get or create item stats
	stats, exists := ps.itemStats[itemID]
	if !exists {
		stats = &ItemTradeStats{
			ItemID:      itemID,
			TradeCount:  0,
			TotalProfit: 0,
			TotalLoss:   0,
			BestDeal:    0,
			WorstDeal:   0,
		}
		ps.itemStats[itemID] = stats
	}

	// Update item stats
	stats.TradeCount++
	if profit > 0 {
		stats.TotalProfit += profit
		if profit > stats.BestDeal {
			stats.BestDeal = profit
		}
	} else {
		stats.TotalLoss += -profit
		if profit < stats.WorstDeal {
			stats.WorstDeal = profit
		}
	}

	// Also record general trade
	ps.totalTrades++
	if profit > 0 {
		ps.profitableTrades++
		ps.totalProfit += profit
		if profit > ps.bestTrade {
			ps.bestTrade = profit
		}
	} else {
		ps.totalLoss += -profit
		if profit < ps.worstTrade {
			ps.worstTrade = profit
		}
	}
}

// GetTotalTrades returns the total number of trades
func (ps *PlayerStats) GetTotalTrades() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.totalTrades
}

// GetProfitableTrades returns the number of profitable trades
func (ps *PlayerStats) GetProfitableTrades() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.profitableTrades
}

// GetTotalProfit returns the total profit (profit - losses)
func (ps *PlayerStats) GetTotalProfit() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.totalProfit - ps.totalLoss
}

// GetSuccessRate returns the success rate (profitable trades / total trades)
func (ps *PlayerStats) GetSuccessRate() float64 {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.totalTrades == 0 {
		return 0.0
	}

	return float64(ps.profitableTrades) / float64(ps.totalTrades)
}

// GetBestTrade returns the most profitable single trade
func (ps *PlayerStats) GetBestTrade() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.bestTrade
}

// GetWorstTrade returns the worst single trade
func (ps *PlayerStats) GetWorstTrade() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.worstTrade
}

// GetItemStats returns statistics for a specific item
func (ps *PlayerStats) GetItemStats(itemID string) ItemTradeStats {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if stats, exists := ps.itemStats[itemID]; exists {
		return *stats
	}

	return ItemTradeStats{
		ItemID:      itemID,
		TradeCount:  0,
		TotalProfit: 0,
		TotalLoss:   0,
		BestDeal:    0,
		WorstDeal:   0,
	}
}

// GetTopTradedItems returns the most traded items
func (ps *PlayerStats) GetTopTradedItems(limit int) []ItemTradeStats {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Convert map to slice for sorting
	items := make([]ItemTradeStats, 0, len(ps.itemStats))
	for _, stats := range ps.itemStats {
		items = append(items, *stats)
	}

	// Simple bubble sort by trade count (for small datasets)
	for i := 0; i < len(items)-1; i++ {
		for j := 0; j < len(items)-i-1; j++ {
			if items[j].TradeCount < items[j+1].TradeCount {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
	}

	// Return top N items
	if limit > len(items) {
		limit = len(items)
	}

	return items[:limit]
}

// GetMostProfitableItems returns the most profitable items
func (ps *PlayerStats) GetMostProfitableItems(limit int) []ItemTradeStats {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Convert map to slice for sorting
	items := make([]ItemTradeStats, 0, len(ps.itemStats))
	for _, stats := range ps.itemStats {
		items = append(items, *stats)
	}

	// Sort by total profit
	for i := 0; i < len(items)-1; i++ {
		for j := 0; j < len(items)-i-1; j++ {
			profit1 := items[j].TotalProfit - items[j].TotalLoss
			profit2 := items[j+1].TotalProfit - items[j+1].TotalLoss
			if profit1 < profit2 {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
	}

	// Return top N items
	if limit > len(items) {
		limit = len(items)
	}

	return items[:limit]
}

// GetDailyStats returns statistics for a specific date
func (ps *PlayerStats) GetDailyStats(date string) *DailyStats {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if stats, exists := ps.dailyStats[date]; exists {
		return stats
	}

	return &DailyStats{
		Date:       date,
		TradeCount: 0,
		Profit:     0,
		GoldEarned: 0,
		GoldSpent:  0,
	}
}

// GetRecentDailyStats returns recent daily statistics
func (ps *PlayerStats) GetRecentDailyStats(days int) []*DailyStats {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	stats := make([]*DailyStats, 0, days)
	now := time.Now()

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		if daily, exists := ps.dailyStats[date]; exists {
			stats = append(stats, daily)
		} else {
			stats = append(stats, &DailyStats{
				Date:       date,
				TradeCount: 0,
				Profit:     0,
				GoldEarned: 0,
				GoldSpent:  0,
			})
		}
	}

	return stats
}

// UpdatePlayTime updates the total play time
func (ps *PlayerStats) UpdatePlayTime() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	now := time.Now()
	ps.playTime += now.Sub(ps.lastPlayedTime)
	ps.lastPlayedTime = now
}

// GetPlayTime returns the total play time
func (ps *PlayerStats) GetPlayTime() time.Duration {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.playTime
}

// ResetStats resets all statistics (for new game)
func (ps *PlayerStats) ResetStats() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.totalTrades = 0
	ps.profitableTrades = 0
	ps.totalProfit = 0
	ps.totalLoss = 0
	ps.bestTrade = 0
	ps.worstTrade = 0
	ps.totalGoldEarned = 0
	ps.totalGoldSpent = 0
	ps.itemStats = make(map[string]*ItemTradeStats)
	ps.dailyStats = make(map[string]*DailyStats)
	ps.playTime = 0
	ps.gameStartTime = time.Now()
	ps.lastPlayedTime = time.Now()
}
