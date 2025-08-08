package summary

import (
	"sync"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/market"
)

// DailySummary represents a summary of a single day's activities
type DailySummary struct {
	Date          time.Time
	DayNumber     int
	Season        string
	StartingGold  float64
	EndingGold    float64
	Profit        float64
	ProfitMargin  float64
	TotalRevenue  float64
	TotalExpenses float64

	// Trading statistics
	ItemsBought      int
	ItemsSold        int
	TotalTrades      int
	SuccessfulTrades int
	TradeSuccessRate float64

	// Top performers
	BestSellingItem  string
	BestProfitItem   string
	BestProfitMargin float64
	WorstSellingItem string

	// Market conditions
	MarketTrend        market.PriceTrend
	AveragePriceChange float64
	MostVolatileItem   string

	// Customer statistics
	CustomersServed    int
	AverageTransaction float64
	ReputationChange   int

	// Inventory
	ItemsSpoiled     int
	SpoilageValue    float64
	CurrentInventory int
	WarehouseUsage   float64

	// Events
	EventsOccurred []string
	EventImpact    float64

	// Achievements
	AchievementsEarned []string
	ExperienceGained   int
}

// SummaryAnalysis provides analysis of the daily summary
type SummaryAnalysis struct {
	PerformanceRating string // Excellent, Good, Average, Poor
	Strengths         []string
	Weaknesses        []string
	Recommendations   []string
	TomorrowFocus     []string
}

// WeeklySummary aggregates daily summaries for a week
type WeeklySummary struct {
	StartDate      time.Time
	EndDate        time.Time
	WeekNumber     int
	DailySummaries []*DailySummary

	// Aggregated stats
	TotalProfit        float64
	AverageDailyProfit float64
	BestDay            *DailySummary
	WorstDay           *DailySummary

	// Trends
	ProfitTrend     market.PriceTrend
	ReputationTrend market.PriceTrend
	CustomerTrend   market.PriceTrend

	// Goals progress
	WeeklyGoalsCompleted int
	WeeklyGoalsTotal     int
}

// SummaryManager manages daily and weekly summaries
type SummaryManager struct {
	currentDay      *DailySummary
	dailySummaries  []*DailySummary
	weeklySummaries []*WeeklySummary

	// Tracking for current day
	dayStartGold float64
	revenue      float64
	expenses     float64
	trades       []*TradeRecord
	events       []string
	achievements []string

	// Configuration
	autoSave       bool
	maxHistoryDays int

	mu sync.RWMutex
}

// TradeRecord represents a single trade
type TradeRecord struct {
	Timestamp    time.Time
	ItemID       string
	ItemName     string
	Quantity     int
	UnitPrice    float64
	TotalPrice   float64
	Type         TradeType
	Successful   bool
	CustomerID   string
	ProfitMargin float64
}

// TradeType represents the type of trade
type TradeType int

const (
	TradeBuy TradeType = iota
	TradeSell
	TradeWholesale
	TradeRetail
)

// NewSummaryManager creates a new summary manager
func NewSummaryManager() *SummaryManager {
	return &SummaryManager{
		dailySummaries:  make([]*DailySummary, 0),
		weeklySummaries: make([]*WeeklySummary, 0),
		trades:          make([]*TradeRecord, 0),
		events:          make([]string, 0),
		achievements:    make([]string, 0),
		autoSave:        true,
		maxHistoryDays:  30,
	}
}

// StartDay begins tracking for a new day
func (sm *SummaryManager) StartDay(dayNumber int, season string, startingGold float64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.currentDay = &DailySummary{
		Date:         time.Now(),
		DayNumber:    dayNumber,
		Season:       season,
		StartingGold: startingGold,
		EndingGold:   startingGold,
	}

	sm.dayStartGold = startingGold
	sm.revenue = 0
	sm.expenses = 0
	sm.trades = make([]*TradeRecord, 0)
	sm.events = make([]string, 0)
	sm.achievements = make([]string, 0)
}

// RecordTrade records a trade transaction
func (sm *SummaryManager) RecordTrade(trade *TradeRecord) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.currentDay == nil {
		return
	}

	sm.trades = append(sm.trades, trade)

	// Update statistics
	sm.currentDay.TotalTrades++

	if trade.Type == TradeBuy || trade.Type == TradeWholesale {
		sm.currentDay.ItemsBought += trade.Quantity
		sm.expenses += trade.TotalPrice
		sm.currentDay.TotalExpenses = sm.expenses
	} else {
		sm.currentDay.ItemsSold += trade.Quantity
		sm.revenue += trade.TotalPrice
		sm.currentDay.TotalRevenue = sm.revenue
	}

	if trade.Successful {
		sm.currentDay.SuccessfulTrades++
	}
}

// RecordEvent records a game event
func (sm *SummaryManager) RecordEvent(eventName string, impact float64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.currentDay == nil {
		return
	}

	sm.events = append(sm.events, eventName)
	sm.currentDay.EventsOccurred = sm.events
	sm.currentDay.EventImpact += impact
}

// RecordAchievement records an earned achievement
func (sm *SummaryManager) RecordAchievement(achievement string, experience int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.currentDay == nil {
		return
	}

	sm.achievements = append(sm.achievements, achievement)
	sm.currentDay.AchievementsEarned = sm.achievements
	sm.currentDay.ExperienceGained += experience
}

// UpdateCustomerStats updates customer-related statistics
func (sm *SummaryManager) UpdateCustomerStats(customersServed int, reputationChange int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.currentDay == nil {
		return
	}

	sm.currentDay.CustomersServed = customersServed
	sm.currentDay.ReputationChange = reputationChange

	if customersServed > 0 && sm.revenue > 0 {
		sm.currentDay.AverageTransaction = sm.revenue / float64(customersServed)
	}
}

// UpdateInventoryStats updates inventory-related statistics
func (sm *SummaryManager) UpdateInventoryStats(currentInventory int, warehouseCapacity int, itemsSpoiled int, spoilageValue float64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.currentDay == nil {
		return
	}

	sm.currentDay.CurrentInventory = currentInventory
	sm.currentDay.ItemsSpoiled = itemsSpoiled
	sm.currentDay.SpoilageValue = spoilageValue

	if warehouseCapacity > 0 {
		sm.currentDay.WarehouseUsage = float64(currentInventory) / float64(warehouseCapacity) * 100
	}
}

// UpdateMarketConditions updates market-related statistics
func (sm *SummaryManager) UpdateMarketConditions(trend market.PriceTrend, avgPriceChange float64, volatileItem string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.currentDay == nil {
		return
	}

	sm.currentDay.MarketTrend = trend
	sm.currentDay.AveragePriceChange = avgPriceChange
	sm.currentDay.MostVolatileItem = volatileItem
}

// EndDay finalizes the current day and generates summary
func (sm *SummaryManager) EndDay(endingGold float64) *DailySummary {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.currentDay == nil {
		return nil
	}

	sm.currentDay.EndingGold = endingGold
	sm.finalizeDaySummary()

	summary := sm.currentDay
	sm.dailySummaries = append(sm.dailySummaries, summary)

	// Check if week is complete (7 days have been completed)
	if len(sm.dailySummaries) >= 7 && len(sm.dailySummaries)%7 == 0 {
		sm.generateWeeklySummary()
	}

	// Cleanup old summaries
	sm.cleanupOldSummaries()

	// Clear current day
	sm.currentDay = nil

	return summary
}

// finalizeDaySummary calculates final statistics for the day
func (sm *SummaryManager) finalizeDaySummary() {
	if sm.currentDay == nil {
		return
	}

	// Calculate profit
	sm.currentDay.Profit = sm.currentDay.EndingGold - sm.currentDay.StartingGold

	// Calculate profit margin
	if sm.currentDay.TotalRevenue > 0 {
		sm.currentDay.ProfitMargin = (sm.currentDay.Profit / sm.currentDay.TotalRevenue) * 100
	}

	// Calculate trade success rate
	if sm.currentDay.TotalTrades > 0 {
		sm.currentDay.TradeSuccessRate = float64(sm.currentDay.SuccessfulTrades) / float64(sm.currentDay.TotalTrades) * 100
	}

	// Analyze trades for best/worst performers
	sm.analyzeTradePerformance()
}

// analyzeTradePerformance analyzes trade records to find best/worst performers
func (sm *SummaryManager) analyzeTradePerformance() {
	if len(sm.trades) == 0 {
		return
	}

	itemProfits := make(map[string]float64)
	itemQuantities := make(map[string]int)
	itemMargins := make(map[string]float64)

	for _, trade := range sm.trades {
		if trade.Type == TradeSell || trade.Type == TradeRetail {
			// Calculate profit: positive margin means profit, negative means loss
			profit := trade.TotalPrice * (trade.ProfitMargin / 100.0)
			itemProfits[trade.ItemName] += profit
			itemQuantities[trade.ItemName] += trade.Quantity

			if trade.ProfitMargin > itemMargins[trade.ItemName] {
				itemMargins[trade.ItemName] = trade.ProfitMargin
			}
		}
	}

	// Find best selling item by quantity
	maxQuantity := 0
	for item, qty := range itemQuantities {
		if qty > maxQuantity {
			maxQuantity = qty
			sm.currentDay.BestSellingItem = item
		}
	}

	// Find best and worst profit items
	maxProfit := -99999.0
	minProfit := 99999.0
	for item, profit := range itemProfits {
		if profit > maxProfit {
			maxProfit = profit
			sm.currentDay.BestProfitItem = item
		}
		if profit < minProfit {
			minProfit = profit
			sm.currentDay.WorstSellingItem = item
		}
	}

	// Find best profit margin
	for _, margin := range itemMargins {
		if margin > sm.currentDay.BestProfitMargin {
			sm.currentDay.BestProfitMargin = margin
		}
	}
}

// generateWeeklySummary creates a weekly summary from the last 7 days
func (sm *SummaryManager) generateWeeklySummary() {
	if len(sm.dailySummaries) < 7 {
		return
	}

	startIdx := len(sm.dailySummaries) - 7
	weekDays := sm.dailySummaries[startIdx:]

	weekly := &WeeklySummary{
		StartDate:      weekDays[0].Date,
		EndDate:        weekDays[6].Date,
		WeekNumber:     len(sm.weeklySummaries) + 1,
		DailySummaries: weekDays,
	}

	// Calculate aggregated stats
	var totalProfit float64
	var bestProfit float64 = -99999
	var worstProfit float64 = 99999

	for _, day := range weekDays {
		totalProfit += day.Profit

		if day.Profit > bestProfit {
			bestProfit = day.Profit
			weekly.BestDay = day
		}

		if day.Profit < worstProfit {
			worstProfit = day.Profit
			weekly.WorstDay = day
		}
	}

	weekly.TotalProfit = totalProfit
	weekly.AverageDailyProfit = totalProfit / 7

	// Determine trends
	if weekDays[6].Profit > weekDays[0].Profit {
		weekly.ProfitTrend = market.TrendUp
	} else if weekDays[6].Profit < weekDays[0].Profit {
		weekly.ProfitTrend = market.TrendDown
	} else {
		weekly.ProfitTrend = market.TrendStable
	}

	sm.weeklySummaries = append(sm.weeklySummaries, weekly)
}

// cleanupOldSummaries removes summaries older than maxHistoryDays
func (sm *SummaryManager) cleanupOldSummaries() {
	if len(sm.dailySummaries) > sm.maxHistoryDays {
		sm.dailySummaries = sm.dailySummaries[len(sm.dailySummaries)-sm.maxHistoryDays:]
	}
}

// GetCurrentDaySummary returns the current day's summary
func (sm *SummaryManager) GetCurrentDaySummary() *DailySummary {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentDay
}

// GetDailySummaries returns all daily summaries
func (sm *SummaryManager) GetDailySummaries() []*DailySummary {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]*DailySummary, len(sm.dailySummaries))
	copy(result, sm.dailySummaries)
	return result
}

// GetWeeklySummaries returns all weekly summaries
func (sm *SummaryManager) GetWeeklySummaries() []*WeeklySummary {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]*WeeklySummary, len(sm.weeklySummaries))
	copy(result, sm.weeklySummaries)
	return result
}

// AnalyzePerformance provides analysis of a daily summary
func (sm *SummaryManager) AnalyzePerformance(summary *DailySummary) *SummaryAnalysis {
	if summary == nil {
		return nil
	}

	analysis := &SummaryAnalysis{
		Strengths:       make([]string, 0),
		Weaknesses:      make([]string, 0),
		Recommendations: make([]string, 0),
		TomorrowFocus:   make([]string, 0),
	}

	// Determine performance rating
	if summary.ProfitMargin > 30 {
		analysis.PerformanceRating = "Excellent"
	} else if summary.ProfitMargin > 20 {
		analysis.PerformanceRating = "Good"
	} else if summary.ProfitMargin > 10 {
		analysis.PerformanceRating = "Average"
	} else {
		analysis.PerformanceRating = "Poor"
	}

	// Identify strengths
	if summary.TradeSuccessRate > 80 {
		analysis.Strengths = append(analysis.Strengths, "High trade success rate")
	}
	if summary.ProfitMargin > 25 {
		analysis.Strengths = append(analysis.Strengths, "Strong profit margins")
	}
	if summary.CustomersServed > 20 {
		analysis.Strengths = append(analysis.Strengths, "High customer volume")
	}
	if summary.ItemsSpoiled == 0 {
		analysis.Strengths = append(analysis.Strengths, "No inventory spoilage")
	}

	// Identify weaknesses
	if summary.TradeSuccessRate < 50 {
		analysis.Weaknesses = append(analysis.Weaknesses, "Low trade success rate")
	}
	if summary.ProfitMargin < 10 {
		analysis.Weaknesses = append(analysis.Weaknesses, "Poor profit margins")
	}
	if summary.ItemsSpoiled > 5 {
		analysis.Weaknesses = append(analysis.Weaknesses, "High inventory spoilage")
	}
	if summary.CustomersServed < 10 {
		analysis.Weaknesses = append(analysis.Weaknesses, "Low customer traffic")
	}

	// Generate recommendations
	if summary.ProfitMargin < 20 {
		analysis.Recommendations = append(analysis.Recommendations, "Focus on higher margin items")
	}
	if summary.ItemsSpoiled > 0 {
		analysis.Recommendations = append(analysis.Recommendations, "Improve inventory rotation")
	}
	if summary.WarehouseUsage > 80 {
		analysis.Recommendations = append(analysis.Recommendations, "Consider expanding warehouse capacity")
	}
	if summary.TradeSuccessRate < 70 {
		analysis.Recommendations = append(analysis.Recommendations, "Be more selective with trades")
	}

	// Tomorrow's focus areas
	if summary.MarketTrend == market.TrendUp {
		analysis.TomorrowFocus = append(analysis.TomorrowFocus, "Sell high-value items while prices are up")
	} else if summary.MarketTrend == market.TrendDown {
		analysis.TomorrowFocus = append(analysis.TomorrowFocus, "Stock up on discounted items")
	}

	if summary.ReputationChange < 0 {
		analysis.TomorrowFocus = append(analysis.TomorrowFocus, "Focus on customer satisfaction")
	}

	return analysis
}

// GetRecentPerformance returns performance metrics for the last N days
func (sm *SummaryManager) GetRecentPerformance(days int) map[string]float64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	metrics := make(map[string]float64)

	if len(sm.dailySummaries) == 0 {
		return metrics
	}

	startIdx := 0
	if len(sm.dailySummaries) > days {
		startIdx = len(sm.dailySummaries) - days
	}

	recentDays := sm.dailySummaries[startIdx:]

	totalProfit := 0.0
	totalRevenue := 0.0
	totalTrades := 0
	successfulTrades := 0

	for _, day := range recentDays {
		totalProfit += day.Profit
		totalRevenue += day.TotalRevenue
		totalTrades += day.TotalTrades
		successfulTrades += day.SuccessfulTrades
	}

	metrics["average_daily_profit"] = totalProfit / float64(len(recentDays))
	metrics["total_profit"] = totalProfit
	metrics["average_daily_revenue"] = totalRevenue / float64(len(recentDays))

	if totalTrades > 0 {
		metrics["trade_success_rate"] = float64(successfulTrades) / float64(totalTrades) * 100
	}

	return metrics
}
