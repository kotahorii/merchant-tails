package balance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/merchant-tails/game/internal/domain/item"
)

func TestNewPriceBalancer(t *testing.T) {
	// Test with default config
	pb := NewPriceBalancer(nil)
	assert.NotNil(t, pb)
	assert.NotNil(t, pb.config)
	assert.NotNil(t, pb.metrics)
	assert.NotNil(t, pb.itemBalance)
	assert.NotNil(t, pb.priceHistory)

	// Test with custom config
	config := &PriceBalanceConfig{
		MinPriceMultiplier: 0.3,
		MaxPriceMultiplier: 5.0,
	}
	pb2 := NewPriceBalancer(config)
	assert.Equal(t, 0.3, pb2.config.MinPriceMultiplier)
	assert.Equal(t, 5.0, pb2.config.MaxPriceMultiplier)
}

func TestRegisterItem(t *testing.T) {
	pb := NewPriceBalancer(nil)

	// Register an item
	pb.RegisterItem("apple", item.CategoryFruit, 10.0)

	// Check if item was registered
	balance, exists := pb.GetItemBalance("apple")
	assert.True(t, exists)
	assert.NotNil(t, balance)
	assert.Equal(t, "apple", balance.ItemID)
	assert.Equal(t, item.CategoryFruit, balance.Category)
	assert.Equal(t, 10.0, balance.BasePrice)
	assert.Equal(t, 10.0, balance.CurrentPrice)
	assert.Equal(t, 1.0, balance.PriceMultiplier)
}

func TestRecordSale(t *testing.T) {
	pb := NewPriceBalancer(nil)
	pb.RegisterItem("apple", item.CategoryFruit, 10.0)

	// Record a sale
	pb.RecordSale("apple", 12.0, 5, 1000.0)

	// Check metrics
	metrics := pb.GetMetrics()
	assert.Equal(t, 1, metrics.TotalTransactions)
	assert.Equal(t, 60.0, metrics.TotalVolume) // 12.0 * 5
	assert.Equal(t, 60.0, metrics.AverageTransaction)

	// Check item balance
	balance, _ := pb.GetItemBalance("apple")
	assert.Len(t, balance.RecentSales, 1)
	assert.Equal(t, 12.0, balance.RecentSales[0].Price)
	assert.Equal(t, 5, balance.RecentSales[0].Quantity)
}

func TestUpdateSupplyDemand(t *testing.T) {
	pb := NewPriceBalancer(nil)
	pb.RegisterItem("apple", item.CategoryFruit, 10.0)

	// Update supply and demand
	pb.UpdateSupplyDemand("apple", 100, 50)

	// Check if updated
	balance, _ := pb.GetItemBalance("apple")
	assert.Equal(t, 100, balance.SupplyLevel)
	assert.Equal(t, 50, balance.DemandLevel)
}

func TestCalculateOptimalPrice(t *testing.T) {
	pb := NewPriceBalancer(nil)
	pb.RegisterItem("apple", item.CategoryFruit, 10.0)

	// Test with normal supply/demand
	pb.UpdateSupplyDemand("apple", 100, 100)
	price := pb.CalculateOptimalPrice("apple", 1) // Apprentice rank

	// Should apply early game multiplier and margin
	// Base: 10 * 1.0 (normal ratio) * 0.8 (early game) * 1.3 (30% margin) = 10.4
	assert.InDelta(t, 10.4, price, 0.5)

	// Test with oversupply
	pb.UpdateSupplyDemand("apple", 200, 100)
	price = pb.CalculateOptimalPrice("apple", 1)
	assert.Less(t, price, 10.4) // Should be lower due to oversupply

	// Test with scarcity
	pb.UpdateSupplyDemand("apple", 40, 100)
	price = pb.CalculateOptimalPrice("apple", 1)
	assert.Greater(t, price, 10.4) // Should be higher due to scarcity
}

func TestAdjustPrices(t *testing.T) {
	pb := NewPriceBalancer(nil)

	// Register multiple items
	pb.RegisterItem("apple", item.CategoryFruit, 10.0)
	pb.RegisterItem("sword", item.CategoryWeapon, 100.0)

	// Set supply/demand
	pb.UpdateSupplyDemand("apple", 150, 100) // Oversupply
	pb.UpdateSupplyDemand("sword", 50, 100)  // Scarcity

	// Adjust prices
	adjustments := pb.AdjustPrices(1) // Apprentice rank

	// Check adjustments
	assert.Len(t, adjustments, 2)

	// Apple should decrease (oversupply)
	assert.Less(t, adjustments["apple"], 10.0)

	// Sword should increase (scarcity)
	assert.Greater(t, adjustments["sword"], 100.0)
}

func TestProgressionMultipliers(t *testing.T) {
	pb := NewPriceBalancer(nil)
	pb.RegisterItem("apple", item.CategoryFruit, 10.0)
	pb.UpdateSupplyDemand("apple", 100, 100)

	// Test different ranks
	priceApprentice := pb.CalculateOptimalPrice("apple", 1)
	priceJourneyman := pb.CalculateOptimalPrice("apple", 2)
	priceMaster := pb.CalculateOptimalPrice("apple", 4)

	// Prices should increase with rank
	assert.Less(t, priceApprentice, priceJourneyman)
	assert.Less(t, priceJourneyman, priceMaster)
}

func TestVolatilityCalculation(t *testing.T) {
	pb := NewPriceBalancer(nil)
	pb.RegisterItem("apple", item.CategoryFruit, 10.0)

	// Record sales with varying prices
	pb.RecordSale("apple", 10.0, 1, 1000.0)
	pb.RecordSale("apple", 12.0, 1, 1000.0)
	pb.RecordSale("apple", 8.0, 1, 1000.0)
	pb.RecordSale("apple", 15.0, 1, 1000.0)
	pb.RecordSale("apple", 5.0, 1, 1000.0)

	// Calculate optimal price - should factor in volatility
	price := pb.CalculateOptimalPrice("apple", 2)
	assert.Greater(t, price, 0.0)
}

func TestSaleRecordLimit(t *testing.T) {
	pb := NewPriceBalancer(nil)
	pb.RegisterItem("apple", item.CategoryFruit, 10.0)

	// Record more than 100 sales
	for i := 0; i < 150; i++ {
		pb.RecordSale("apple", 10.0, 1, 1000.0)
	}

	// Should keep only last 100
	balance, _ := pb.GetItemBalance("apple")
	assert.Len(t, balance.RecentSales, 100)
}

func TestAdjustmentCallbacks(t *testing.T) {
	pb := NewPriceBalancer(nil)
	pb.RegisterItem("apple", item.CategoryFruit, 10.0)

	callbackCalled := false
	var capturedReason string

	// Register callback
	pb.RegisterAdjustmentCallback(func(itemID string, oldPrice, newPrice float64, reason string) {
		callbackCalled = true
		capturedReason = reason
		assert.Equal(t, "apple", itemID)
		assert.Equal(t, 10.0, oldPrice)
		assert.NotEqual(t, oldPrice, newPrice)
	})

	// Set oversupply condition
	pb.UpdateSupplyDemand("apple", 200, 100)

	// Trigger adjustment
	pb.AdjustPrices(1)

	assert.True(t, callbackCalled)
	assert.Equal(t, "oversupply", capturedReason)
}

func TestMarketMetrics(t *testing.T) {
	pb := NewPriceBalancer(nil)

	// Register items
	pb.RegisterItem("apple", item.CategoryFruit, 10.0)
	pb.RegisterItem("sword", item.CategoryWeapon, 100.0)

	// Record some sales
	pb.RecordSale("apple", 12.0, 5, 1000.0)
	pb.RecordSale("sword", 110.0, 2, 5000.0)

	// Set supply/demand
	pb.UpdateSupplyDemand("apple", 100, 80)
	pb.UpdateSupplyDemand("sword", 50, 60)

	// Trigger price adjustment to update metrics
	pb.AdjustPrices(2)

	metrics := pb.GetMetrics()
	assert.Equal(t, 2, metrics.TotalTransactions)
	assert.Equal(t, 280.0, metrics.TotalVolume) // (12*5) + (110*2)
	assert.Equal(t, 140.0, metrics.AverageTransaction)
	assert.Greater(t, metrics.SupplyDemandRatio, 0.0)
}

func TestPriceBounds(t *testing.T) {
	config := &PriceBalanceConfig{
		TargetMargins: map[item.Category]float64{
			item.CategoryFruit: 0.30,
		},
		MinPriceMultiplier:   0.5,
		MaxPriceMultiplier:   2.0,
		PriceAdjustmentSpeed: 0.1,
		OversupplyThreshold:  1.5,
		ScarcityThreshold:    0.5,
		EarlyGameMultiplier:  1.0,
		MidGameMultiplier:    1.0,
		LateGameMultiplier:   1.0,
	}

	pb := NewPriceBalancer(config)
	pb.RegisterItem("apple", item.CategoryFruit, 10.0)

	// Test minimum bound
	pb.UpdateSupplyDemand("apple", 1000, 10) // Extreme oversupply
	price := pb.CalculateOptimalPrice("apple", 2)
	assert.GreaterOrEqual(t, price, 5.0) // 10 * 0.5 minimum

	// Test maximum bound
	pb.UpdateSupplyDemand("apple", 1, 1000) // Extreme scarcity
	price = pb.CalculateOptimalPrice("apple", 2)
	assert.LessOrEqual(t, price, 20.0) // 10 * 2.0 maximum
}

func TestReset(t *testing.T) {
	pb := NewPriceBalancer(nil)

	// Add some data
	pb.RegisterItem("apple", item.CategoryFruit, 10.0)
	pb.RecordSale("apple", 12.0, 5, 1000.0)

	// Verify data exists
	balance, exists := pb.GetItemBalance("apple")
	assert.True(t, exists)
	assert.NotNil(t, balance)

	metrics := pb.GetMetrics()
	assert.Equal(t, 1, metrics.TotalTransactions)

	// Reset
	pb.Reset()

	// Verify data is cleared
	balance, exists = pb.GetItemBalance("apple")
	assert.False(t, exists)
	assert.Nil(t, balance)

	metrics = pb.GetMetrics()
	assert.Equal(t, 0, metrics.TotalTransactions)
}

func TestConcurrentAccess(t *testing.T) {
	pb := NewPriceBalancer(nil)
	pb.RegisterItem("apple", item.CategoryFruit, 10.0)

	done := make(chan bool, 3)

	// Concurrent sales recording
	go func() {
		for i := 0; i < 100; i++ {
			pb.RecordSale("apple", 10.0+float64(i%5), 1, 1000.0)
		}
		done <- true
	}()

	// Concurrent supply/demand updates
	go func() {
		for i := 0; i < 100; i++ {
			pb.UpdateSupplyDemand("apple", 100+i, 100-i)
		}
		done <- true
	}()

	// Concurrent price adjustments
	go func() {
		for i := 0; i < 50; i++ {
			pb.AdjustPrices(1)
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify state is consistent
	balance, exists := pb.GetItemBalance("apple")
	assert.True(t, exists)
	assert.NotNil(t, balance)

	metrics := pb.GetMetrics()
	assert.Greater(t, metrics.TotalTransactions, 0)
}

func TestAdjustmentReasons(t *testing.T) {
	pb := NewPriceBalancer(nil)
	pb.RegisterItem("apple", item.CategoryFruit, 10.0)

	reasons := make([]string, 0)
	pb.RegisterAdjustmentCallback(func(itemID string, oldPrice, newPrice float64, reason string) {
		reasons = append(reasons, reason)
	})

	// Test oversupply
	pb.UpdateSupplyDemand("apple", 200, 100)
	pb.AdjustPrices(1)
	assert.Contains(t, reasons, "oversupply")

	// Test scarcity
	pb.UpdateSupplyDemand("apple", 40, 100)
	pb.AdjustPrices(1)
	assert.Contains(t, reasons, "scarcity")

	// Test equilibrium
	pb.UpdateSupplyDemand("apple", 100, 100)
	pb.AdjustPrices(1)
	assert.Contains(t, reasons, "market_equilibrium")
}
