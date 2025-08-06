package trading

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/merchant-tails/game/internal/domain/inventory"
	"github.com/yourusername/merchant-tails/game/internal/domain/item"
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
)

func TestNewTradingSystem(t *testing.T) {
	invManager, _ := inventory.NewInventoryManager(20, 100)
	marketSystem := market.NewMarketSystem()

	tests := []struct {
		name        string
		inventory   *inventory.InventoryManager
		market      *market.MarketSystem
		wantErr     bool
		errContains string
	}{
		{
			name:      "create valid trading system",
			inventory: invManager,
			market:    marketSystem,
			wantErr:   false,
		},
		{
			name:        "nil inventory manager",
			inventory:   nil,
			market:      marketSystem,
			wantErr:     true,
			errContains: "inventory manager is required",
		},
		{
			name:        "nil market system",
			inventory:   invManager,
			market:      nil,
			wantErr:     true,
			errContains: "market system is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system, err := NewTradingSystem(tt.inventory, tt.market)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, system)
			} else {
				require.NoError(t, err)
				require.NotNil(t, system)
			}
		})
	}
}

func TestTradingSystem_BuyFromSupplier(t *testing.T) {
	invManager, _ := inventory.NewInventoryManager(20, 100)
	marketSystem := market.NewMarketSystem()
	tradingSystem, _ := NewTradingSystem(invManager, marketSystem)

	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	marketSystem.SetBasePrice("apple_001", 10)

	tests := []struct {
		name          string
		item          *item.Item
		quantity      int
		availableGold int
		expectedGold  int
		shouldSucceed bool
	}{
		{
			name:          "buy within budget and capacity",
			item:          apple,
			quantity:      5,
			availableGold: 100,
			expectedGold:  50, // 5 * 10
			shouldSucceed: true,
		},
		{
			name:          "exceed budget",
			item:          apple,
			quantity:      20,
			availableGold: 100,
			expectedGold:  0,
			shouldSucceed: false,
		},
		{
			name:          "exceed capacity",
			item:          apple,
			quantity:      200,
			availableGold: 5000,
			expectedGold:  0,
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tradingSystem.SetGold(tt.availableGold)

			transaction, err := tradingSystem.BuyFromSupplier(tt.item, tt.quantity)

			if tt.shouldSucceed {
				require.NoError(t, err)
				require.NotNil(t, transaction)
				assert.Equal(t, TransactionTypeBuy, transaction.Type)
				assert.Equal(t, tt.item.ID, transaction.ItemID)
				assert.Equal(t, tt.quantity, transaction.Quantity)
				assert.Equal(t, tt.expectedGold, transaction.TotalCost)
				assert.Equal(t, tt.availableGold-tt.expectedGold, tradingSystem.GetGold())
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestTradingSystem_SellToCustomer(t *testing.T) {
	invManager, _ := inventory.NewInventoryManager(20, 100)
	marketSystem := market.NewMarketSystem()
	tradingSystem, _ := NewTradingSystem(invManager, marketSystem)

	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	marketSystem.SetBasePrice("apple_001", 10)

	// Setup: Add items to shop
	invManager.AddToShop(apple, 10)
	tradingSystem.SetGold(100)

	tests := []struct {
		name            string
		itemID          string
		quantity        int
		customerBudget  int
		expectedRevenue int
		shouldSucceed   bool
	}{
		{
			name:            "sell within stock",
			itemID:          "apple_001",
			quantity:        3,
			customerBudget:  100,
			expectedRevenue: 30, // 3 * 10
			shouldSucceed:   true,
		},
		{
			name:            "exceed stock",
			itemID:          "apple_001",
			quantity:        15,
			customerBudget:  200,
			expectedRevenue: 0,
			shouldSucceed:   false,
		},
		{
			name:            "customer budget too low",
			itemID:          "apple_001",
			quantity:        5,
			customerBudget:  20,
			expectedRevenue: 0,
			shouldSucceed:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialGold := tradingSystem.GetGold()

			transaction, err := tradingSystem.SellToCustomer(
				tt.itemID,
				tt.quantity,
				tt.customerBudget,
			)

			if tt.shouldSucceed {
				require.NoError(t, err)
				require.NotNil(t, transaction)
				assert.Equal(t, TransactionTypeSell, transaction.Type)
				assert.Equal(t, tt.itemID, transaction.ItemID)
				assert.Equal(t, tt.quantity, transaction.Quantity)
				assert.Equal(t, tt.expectedRevenue, transaction.TotalCost)
				assert.Equal(t, initialGold+tt.expectedRevenue, tradingSystem.GetGold())
			} else {
				require.Error(t, err)
				assert.Equal(t, initialGold, tradingSystem.GetGold())
			}
		})
	}
}

func TestTradingSystem_MarketOrder(t *testing.T) {
	invManager, _ := inventory.NewInventoryManager(20, 100)
	marketSystem := market.NewMarketSystem()
	tradingSystem, _ := NewTradingSystem(invManager, marketSystem)

	sword, _ := item.NewItem("sword_001", "Iron Sword", item.CategoryWeapon, 200)
	marketSystem.SetBasePrice("sword_001", 200)

	// Setup: Add items and set gold
	invManager.AddToShop(sword, 5)
	tradingSystem.SetGold(500)

	tests := []struct {
		name            string
		itemID          string
		quantity        int
		orderType       OrderType
		priceLimit      int
		expectedSuccess bool
	}{
		{
			name:            "market sell order",
			itemID:          "sword_001",
			quantity:        2,
			orderType:       OrderTypeSell,
			priceLimit:      0, // Market order
			expectedSuccess: true,
		},
		{
			name:            "limit sell order - price met",
			itemID:          "sword_001",
			quantity:        1,
			orderType:       OrderTypeSell,
			priceLimit:      180,
			expectedSuccess: true,
		},
		{
			name:            "limit sell order - price not met",
			itemID:          "sword_001",
			quantity:        1,
			orderType:       OrderTypeSell,
			priceLimit:      250,
			expectedSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := tradingSystem.PlaceMarketOrder(
				tt.itemID,
				tt.quantity,
				tt.orderType,
				tt.priceLimit,
			)

			require.NoError(t, err)
			require.NotNil(t, order)

			// Process order
			success := tradingSystem.ProcessOrder(order)
			assert.Equal(t, tt.expectedSuccess, success)

			if tt.expectedSuccess {
				assert.Equal(t, OrderStatusCompleted, order.Status)
			} else {
				assert.Equal(t, OrderStatusPending, order.Status)
			}
		})
	}
}

func TestTradingSystem_Negotiation(t *testing.T) {
	invManager, _ := inventory.NewInventoryManager(20, 100)
	marketSystem := market.NewMarketSystem()
	tradingSystem, _ := NewTradingSystem(invManager, marketSystem)

	potion, _ := item.NewItem("potion_001", "Health Potion", item.CategoryPotion, 50)
	marketSystem.SetBasePrice("potion_001", 50)

	// Add negotiation skill
	tradingSystem.SetNegotiationSkill(10) // 10% discount/markup ability

	tests := []struct {
		name          string
		item          *item.Item
		basePrice     int
		isBuying      bool
		expectedPrice int
	}{
		{
			name:          "negotiate buy price down",
			item:          potion,
			basePrice:     50,
			isBuying:      true,
			expectedPrice: 45, // 10% discount
		},
		{
			name:          "negotiate sell price up",
			item:          potion,
			basePrice:     50,
			isBuying:      false,
			expectedPrice: 55, // 10% markup
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			negotiatedPrice := tradingSystem.NegotiatePrice(
				tt.basePrice,
				tt.isBuying,
			)

			assert.Equal(t, tt.expectedPrice, negotiatedPrice)
		})
	}
}

func TestTradingSystem_BulkDiscount(t *testing.T) {
	invManager, _ := inventory.NewInventoryManager(20, 100)
	marketSystem := market.NewMarketSystem()
	tradingSystem, _ := NewTradingSystem(invManager, marketSystem)

	// gem, _ := item.NewItem("gem_001", "Ruby", item.CategoryGem, 500) // Not used in test

	tests := []struct {
		name             string
		quantity         int
		unitPrice        int
		expectedDiscount float64
		expectedTotal    int
	}{
		{
			name:             "no discount for small quantity",
			quantity:         5,
			unitPrice:        500,
			expectedDiscount: 0,
			expectedTotal:    2500,
		},
		{
			name:             "5% discount for 10+ items",
			quantity:         10,
			unitPrice:        500,
			expectedDiscount: 0.05,
			expectedTotal:    4750,
		},
		{
			name:             "10% discount for 20+ items",
			quantity:         20,
			unitPrice:        500,
			expectedDiscount: 0.10,
			expectedTotal:    9000,
		},
		{
			name:             "15% discount for 50+ items",
			quantity:         50,
			unitPrice:        500,
			expectedDiscount: 0.15,
			expectedTotal:    21250,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discount := tradingSystem.CalculateBulkDiscount(tt.quantity)
			assert.Equal(t, tt.expectedDiscount, discount)

			total := tradingSystem.CalculateTotalWithDiscount(
				tt.unitPrice,
				tt.quantity,
			)
			assert.Equal(t, tt.expectedTotal, total)
		})
	}
}

func TestTradingSystem_TransactionHistory(t *testing.T) {
	invManager, _ := inventory.NewInventoryManager(20, 100)
	marketSystem := market.NewMarketSystem()
	tradingSystem, _ := NewTradingSystem(invManager, marketSystem)

	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	sword, _ := item.NewItem("sword_001", "Iron Sword", item.CategoryWeapon, 200)

	marketSystem.SetBasePrice("apple_001", 10)
	marketSystem.SetBasePrice("sword_001", 200)

	// Perform some transactions
	invManager.AddToShop(apple, 10)
	invManager.AddToShop(sword, 5)
	tradingSystem.SetGold(1000)

	// Buy apples
	tradingSystem.BuyFromSupplier(apple, 5)

	// Sell sword
	tradingSystem.SellToCustomer("sword_001", 2, 500)

	// Get history
	history := tradingSystem.GetTransactionHistory()

	assert.Len(t, history, 2)

	// Verify first transaction (buy)
	assert.Equal(t, TransactionTypeBuy, history[0].Type)
	assert.Equal(t, "apple_001", history[0].ItemID)
	assert.Equal(t, 5, history[0].Quantity)

	// Verify second transaction (sell)
	assert.Equal(t, TransactionTypeSell, history[1].Type)
	assert.Equal(t, "sword_001", history[1].ItemID)
	assert.Equal(t, 2, history[1].Quantity)

	// Test filtering by type
	buyHistory := tradingSystem.GetTransactionsByType(TransactionTypeBuy)
	assert.Len(t, buyHistory, 1)

	sellHistory := tradingSystem.GetTransactionsByType(TransactionTypeSell)
	assert.Len(t, sellHistory, 1)

	// Test date range filtering
	today := time.Now()
	yesterday := today.Add(-24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	todayHistory := tradingSystem.GetTransactionsByDateRange(yesterday, tomorrow)
	assert.Len(t, todayHistory, 2)
}

func TestTradingSystem_ProfitCalculation(t *testing.T) {
	invManager, _ := inventory.NewInventoryManager(20, 100)
	marketSystem := market.NewMarketSystem()
	tradingSystem, _ := NewTradingSystem(invManager, marketSystem)

	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	marketSystem.SetBasePrice("apple_001", 10)

	// Setup
	invManager.AddToShop(apple, 10)
	tradingSystem.SetGold(1000)

	// Record buy price
	tradingSystem.RecordPurchasePrice("apple_001", 8) // Bought at 8 gold each

	// Sell at market price
	tradingSystem.SellToCustomer("apple_001", 5, 100)

	// Calculate profit
	profit := tradingSystem.CalculateProfit("apple_001", 5)
	expectedProfit := (10 - 8) * 5 // (sell price - buy price) * quantity

	assert.Equal(t, expectedProfit, profit)

	// Test overall profit
	totalProfit := tradingSystem.GetTotalProfit()
	assert.Equal(t, expectedProfit, totalProfit)

	// Test profit margin
	margin := tradingSystem.GetProfitMargin("apple_001")
	expectedMargin := float64(10-8) / float64(8) * 100 // 25%
	assert.Equal(t, expectedMargin, margin)
}

func TestTradingSystem_CustomerReputation(t *testing.T) {
	invManager, _ := inventory.NewInventoryManager(20, 100)
	marketSystem := market.NewMarketSystem()
	tradingSystem, _ := NewTradingSystem(invManager, marketSystem)

	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	marketSystem.SetBasePrice("apple_001", 10)
	invManager.AddToShop(apple, 20)

	// Initial reputation
	assert.Equal(t, 50, tradingSystem.GetReputation())

	// Successful transaction improves reputation
	tradingSystem.SellToCustomer("apple_001", 5, 100)
	assert.Greater(t, tradingSystem.GetReputation(), 50)

	// Fair pricing improves reputation
	tradingSystem.SetFairPricing(true)
	tradingSystem.SellToCustomer("apple_001", 3, 50)
	assert.Greater(t, tradingSystem.GetReputation(), 51)

	// Reputation affects prices
	highRepPrice := tradingSystem.GetReputationAdjustedPrice(100, false) // selling
	assert.GreaterOrEqual(t, highRepPrice, 100)                          // Can charge more or same with good reputation

	// Test reputation decay
	tradingSystem.ProcessDailyReputationDecay()
	currentRep := tradingSystem.GetReputation()
	assert.Less(t, currentRep, 100) // Should decay towards neutral (50)
}

func TestTradingSystem_SpecialDeals(t *testing.T) {
	invManager, _ := inventory.NewInventoryManager(20, 100)
	marketSystem := market.NewMarketSystem()
	tradingSystem, _ := NewTradingSystem(invManager, marketSystem)

	// Create special deal
	deal := &SpecialDeal{
		ItemID:       "potion_001",
		Quantity:     10,
		SpecialPrice: 35, // Normal price is 50
		Supplier:     "Alchemist Guild",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	tradingSystem.AddSpecialDeal(deal)

	// Get available deals
	deals := tradingSystem.GetAvailableDeals()
	assert.Len(t, deals, 1)
	assert.Equal(t, "potion_001", deals[0].ItemID)

	// Accept deal
	// potion, _ := item.NewItem("potion_001", "Health Potion", item.CategoryPotion, 50) // Not used
	tradingSystem.SetGold(500)

	transaction, err := tradingSystem.AcceptSpecialDeal(deal.ID)
	require.NoError(t, err)
	assert.Equal(t, 350, transaction.TotalCost)   // 10 * 35
	assert.Equal(t, 150, tradingSystem.GetGold()) // 500 - 350

	// Deal should be removed after acceptance
	deals = tradingSystem.GetAvailableDeals()
	assert.Len(t, deals, 0)

	// Test expired deals
	expiredDeal := &SpecialDeal{
		ItemID:       "sword_001",
		Quantity:     2,
		SpecialPrice: 150,
		Supplier:     "Blacksmith",
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // Already expired
	}

	tradingSystem.AddSpecialDeal(expiredDeal)
	tradingSystem.CleanExpiredDeals()

	deals = tradingSystem.GetAvailableDeals()
	assert.Len(t, deals, 0) // Expired deal should be removed
}
