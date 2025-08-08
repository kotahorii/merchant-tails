package summary

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
)

func TestNewSummaryManager(t *testing.T) {
	sm := NewSummaryManager()

	assert.NotNil(t, sm)
	assert.NotNil(t, sm.dailySummaries)
	assert.NotNil(t, sm.weeklySummaries)
	assert.NotNil(t, sm.trades)
	assert.True(t, sm.autoSave)
	assert.Equal(t, 30, sm.maxHistoryDays)
}

func TestSummaryManager_StartDay(t *testing.T) {
	sm := NewSummaryManager()

	sm.StartDay(1, "Spring", 1000.0)

	currentDay := sm.GetCurrentDaySummary()
	assert.NotNil(t, currentDay)
	assert.Equal(t, 1, currentDay.DayNumber)
	assert.Equal(t, "Spring", currentDay.Season)
	assert.Equal(t, 1000.0, currentDay.StartingGold)
	assert.Equal(t, 1000.0, currentDay.EndingGold)
}

func TestSummaryManager_RecordTrade(t *testing.T) {
	sm := NewSummaryManager()
	sm.StartDay(1, "Spring", 1000.0)

	// Record a buy trade
	buyTrade := &TradeRecord{
		Timestamp:  time.Now(),
		ItemID:     "apple",
		ItemName:   "Apple",
		Quantity:   10,
		UnitPrice:  5.0,
		TotalPrice: 50.0,
		Type:       TradeBuy,
		Successful: true,
	}
	sm.RecordTrade(buyTrade)

	// Record a sell trade
	sellTrade := &TradeRecord{
		Timestamp:    time.Now(),
		ItemID:       "apple",
		ItemName:     "Apple",
		Quantity:     5,
		UnitPrice:    8.0,
		TotalPrice:   40.0,
		Type:         TradeSell,
		Successful:   true,
		ProfitMargin: 20.0,
	}
	sm.RecordTrade(sellTrade)

	currentDay := sm.GetCurrentDaySummary()
	assert.Equal(t, 2, currentDay.TotalTrades)
	assert.Equal(t, 2, currentDay.SuccessfulTrades)
	assert.Equal(t, 10, currentDay.ItemsBought)
	assert.Equal(t, 5, currentDay.ItemsSold)
	assert.Equal(t, 50.0, currentDay.TotalExpenses)
	assert.Equal(t, 40.0, currentDay.TotalRevenue)
}

func TestSummaryManager_RecordEvent(t *testing.T) {
	sm := NewSummaryManager()
	sm.StartDay(1, "Spring", 1000.0)

	sm.RecordEvent("Dragon Attack", -100.0)
	sm.RecordEvent("Festival", 50.0)

	currentDay := sm.GetCurrentDaySummary()
	assert.Len(t, currentDay.EventsOccurred, 2)
	assert.Equal(t, "Dragon Attack", currentDay.EventsOccurred[0])
	assert.Equal(t, "Festival", currentDay.EventsOccurred[1])
	assert.Equal(t, -50.0, currentDay.EventImpact)
}

func TestSummaryManager_RecordAchievement(t *testing.T) {
	sm := NewSummaryManager()
	sm.StartDay(1, "Spring", 1000.0)

	sm.RecordAchievement("First Sale", 10)
	sm.RecordAchievement("Profitable Day", 25)

	currentDay := sm.GetCurrentDaySummary()
	assert.Len(t, currentDay.AchievementsEarned, 2)
	assert.Equal(t, "First Sale", currentDay.AchievementsEarned[0])
	assert.Equal(t, 35, currentDay.ExperienceGained)
}

func TestSummaryManager_UpdateCustomerStats(t *testing.T) {
	sm := NewSummaryManager()
	sm.StartDay(1, "Spring", 1000.0)

	// Record some sales first
	for i := 0; i < 5; i++ {
		sm.RecordTrade(&TradeRecord{
			Type:       TradeSell,
			TotalPrice: 20.0,
			Successful: true,
		})
	}

	sm.UpdateCustomerStats(10, 5)

	currentDay := sm.GetCurrentDaySummary()
	assert.Equal(t, 10, currentDay.CustomersServed)
	assert.Equal(t, 5, currentDay.ReputationChange)
	assert.Equal(t, 10.0, currentDay.AverageTransaction) // 100 revenue / 10 customers
}

func TestSummaryManager_UpdateInventoryStats(t *testing.T) {
	sm := NewSummaryManager()
	sm.StartDay(1, "Spring", 1000.0)

	sm.UpdateInventoryStats(50, 100, 3, 15.0)

	currentDay := sm.GetCurrentDaySummary()
	assert.Equal(t, 50, currentDay.CurrentInventory)
	assert.Equal(t, 3, currentDay.ItemsSpoiled)
	assert.Equal(t, 15.0, currentDay.SpoilageValue)
	assert.Equal(t, 50.0, currentDay.WarehouseUsage)
}

func TestSummaryManager_UpdateMarketConditions(t *testing.T) {
	sm := NewSummaryManager()
	sm.StartDay(1, "Spring", 1000.0)

	sm.UpdateMarketConditions(market.TrendUp, 5.5, "Gold Ring")

	currentDay := sm.GetCurrentDaySummary()
	assert.Equal(t, market.TrendUp, currentDay.MarketTrend)
	assert.Equal(t, 5.5, currentDay.AveragePriceChange)
	assert.Equal(t, "Gold Ring", currentDay.MostVolatileItem)
}

func TestSummaryManager_EndDay(t *testing.T) {
	sm := NewSummaryManager()
	sm.StartDay(1, "Spring", 1000.0)

	// Simulate day's activities
	sm.RecordTrade(&TradeRecord{
		ItemName:   "Apple",
		Quantity:   10,
		UnitPrice:  5.0,
		TotalPrice: 50.0,
		Type:       TradeBuy,
		Successful: true,
	})

	sm.RecordTrade(&TradeRecord{
		ItemName:     "Apple",
		Quantity:     8,
		UnitPrice:    8.0,
		TotalPrice:   64.0,
		Type:         TradeSell,
		Successful:   true,
		ProfitMargin: 25.0,
	})

	summary := sm.EndDay(1014.0)

	assert.NotNil(t, summary)
	assert.Equal(t, 1014.0, summary.EndingGold)
	assert.Equal(t, 14.0, summary.Profit)
	assert.Greater(t, summary.ProfitMargin, 0.0)
	assert.Equal(t, 100.0, summary.TradeSuccessRate)
	assert.Equal(t, "Apple", summary.BestSellingItem)

	// Check that summary was saved
	dailySummaries := sm.GetDailySummaries()
	assert.Len(t, dailySummaries, 1)
}

func TestSummaryManager_WeeklySummary(t *testing.T) {
	sm := NewSummaryManager()

	// Simulate 7 days
	for i := 1; i <= 7; i++ {
		sm.StartDay(i, "Spring", 1000.0+float64(i-1)*10)

		// Simulate some trades
		sm.RecordTrade(&TradeRecord{
			ItemName:   "Apple",
			Quantity:   5,
			UnitPrice:  10.0,
			TotalPrice: 50.0,
			Type:       TradeSell,
			Successful: true,
		})

		sm.EndDay(1000.0 + float64(i)*10)
	}

	// Debug: check daily summaries count
	dailySummaries := sm.GetDailySummaries()
	t.Logf("Daily summaries count: %d", len(dailySummaries))
	t.Logf("Should generate weekly: %v", len(dailySummaries)%7 == 0)

	weeklySummaries := sm.GetWeeklySummaries()
	t.Logf("Weekly summaries count: %d", len(weeklySummaries))
	assert.Len(t, weeklySummaries, 1)

	weekly := weeklySummaries[0]
	assert.NotNil(t, weekly)
	assert.Equal(t, 1, weekly.WeekNumber)
	assert.Len(t, weekly.DailySummaries, 7)
	assert.Greater(t, weekly.TotalProfit, 0.0)
	assert.NotNil(t, weekly.BestDay)
	assert.NotNil(t, weekly.WorstDay)
	// Since all days have same profit (10), trend should be stable
	assert.Equal(t, market.TrendStable, weekly.ProfitTrend)
}

func TestSummaryManager_AnalyzePerformance(t *testing.T) {
	sm := NewSummaryManager()

	tests := []struct {
		name     string
		summary  *DailySummary
		expected string
	}{
		{
			name: "Excellent performance",
			summary: &DailySummary{
				ProfitMargin:     35.0,
				TradeSuccessRate: 85.0,
				CustomersServed:  25,
				ItemsSpoiled:     0,
			},
			expected: "Excellent",
		},
		{
			name: "Good performance",
			summary: &DailySummary{
				ProfitMargin:     22.0,
				TradeSuccessRate: 75.0,
				CustomersServed:  15,
				ItemsSpoiled:     1,
			},
			expected: "Good",
		},
		{
			name: "Poor performance",
			summary: &DailySummary{
				ProfitMargin:     8.0,
				TradeSuccessRate: 40.0,
				CustomersServed:  8,
				ItemsSpoiled:     8,
			},
			expected: "Poor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := sm.AnalyzePerformance(tt.summary)
			assert.NotNil(t, analysis)
			assert.Equal(t, tt.expected, analysis.PerformanceRating)

			// Check that appropriate strengths/weaknesses are identified
			if tt.summary.ProfitMargin > 25 {
				assert.Contains(t, analysis.Strengths, "Strong profit margins")
			}
			if tt.summary.TradeSuccessRate < 50 {
				assert.Contains(t, analysis.Weaknesses, "Low trade success rate")
			}
			if tt.summary.ItemsSpoiled > 5 {
				assert.Contains(t, analysis.Weaknesses, "High inventory spoilage")
			}
		})
	}
}

func TestSummaryManager_GetRecentPerformance(t *testing.T) {
	sm := NewSummaryManager()

	// Simulate 5 days
	for i := 1; i <= 5; i++ {
		sm.StartDay(i, "Spring", 1000.0)

		// Record some trades
		sm.RecordTrade(&TradeRecord{
			ItemName:   "Apple",
			Quantity:   10,
			UnitPrice:  10.0,
			TotalPrice: 100.0,
			Type:       TradeSell,
			Successful: true,
		})

		sm.EndDay(1000.0 + float64(i)*20)
	}

	metrics := sm.GetRecentPerformance(3)

	assert.Greater(t, metrics["average_daily_profit"], 0.0)
	assert.Greater(t, metrics["total_profit"], 0.0)
	assert.Greater(t, metrics["average_daily_revenue"], 0.0)
	assert.Equal(t, 100.0, metrics["trade_success_rate"])
}

func TestSummaryManager_TradePerformanceAnalysis(t *testing.T) {
	sm := NewSummaryManager()
	sm.StartDay(1, "Spring", 1000.0)

	// Record various trades
	trades := []struct {
		item         string
		quantity     int
		price        float64
		profitMargin float64
	}{
		{"Apple", 20, 100.0, 15.0},
		{"Sword", 5, 500.0, 30.0},
		{"Potion", 15, 150.0, 20.0},
		{"Apple", 10, 50.0, 15.0},
		{"Ring", 2, 300.0, -10.0}, // Loss
	}

	for _, trade := range trades {
		sm.RecordTrade(&TradeRecord{
			ItemName:     trade.item,
			Quantity:     trade.quantity,
			UnitPrice:    trade.price / float64(trade.quantity),
			TotalPrice:   trade.price,
			Type:         TradeSell,
			Successful:   true,
			ProfitMargin: trade.profitMargin,
		})
	}

	summary := sm.EndDay(2000.0)

	assert.Equal(t, "Apple", summary.BestSellingItem) // 30 total quantity
	assert.Equal(t, "Sword", summary.BestProfitItem)  // Highest profit
	assert.Equal(t, 30.0, summary.BestProfitMargin)   // Sword's margin
	assert.Equal(t, "Ring", summary.WorstSellingItem) // Negative profit
}

func TestSummaryManager_CleanupOldSummaries(t *testing.T) {
	sm := NewSummaryManager()
	sm.maxHistoryDays = 5 // Set low for testing

	// Create 10 days of summaries
	for i := 1; i <= 10; i++ {
		sm.StartDay(i, "Spring", 1000.0)
		sm.EndDay(1010.0)
	}

	dailySummaries := sm.GetDailySummaries()
	assert.Len(t, dailySummaries, 5)                // Should only keep last 5 days
	assert.Equal(t, 6, dailySummaries[0].DayNumber) // First should be day 6
}

func TestSummaryManager_NilChecks(t *testing.T) {
	sm := NewSummaryManager()

	// Try to record without starting day
	sm.RecordTrade(&TradeRecord{})
	sm.RecordEvent("Test", 0)
	sm.RecordAchievement("Test", 0)
	sm.UpdateCustomerStats(0, 0)
	sm.UpdateInventoryStats(0, 0, 0, 0)
	sm.UpdateMarketConditions(market.TrendStable, 0, "")

	// Should not panic
	assert.Nil(t, sm.GetCurrentDaySummary())

	summary := sm.EndDay(0)
	assert.Nil(t, summary)
}

func TestSummaryManager_TomorrowFocus(t *testing.T) {
	sm := NewSummaryManager()

	// Test market trend recommendations
	summaryUp := &DailySummary{
		MarketTrend:      market.TrendUp,
		ReputationChange: 0,
	}
	analysisUp := sm.AnalyzePerformance(summaryUp)
	assert.Contains(t, analysisUp.TomorrowFocus, "Sell high-value items while prices are up")

	summaryDown := &DailySummary{
		MarketTrend:      market.TrendDown,
		ReputationChange: 0,
	}
	analysisDown := sm.AnalyzePerformance(summaryDown)
	assert.Contains(t, analysisDown.TomorrowFocus, "Stock up on discounted items")

	// Test reputation focus
	summaryRep := &DailySummary{
		MarketTrend:      market.TrendStable,
		ReputationChange: -5,
	}
	analysisRep := sm.AnalyzePerformance(summaryRep)
	assert.Contains(t, analysisRep.TomorrowFocus, "Focus on customer satisfaction")
}
