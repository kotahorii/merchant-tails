package merchant

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/merchant-tails/game/internal/domain/item"
)

func TestNewPlayerMerchant(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		merchantName string
		startingGold int
		wantErr      bool
		errContains  string
	}{
		{
			name:         "create valid player merchant",
			id:           "player_001",
			merchantName: "Hero Trader",
			startingGold: 1000,
			wantErr:      false,
		},
		{
			name:         "create merchant with minimum gold",
			id:           "player_002",
			merchantName: "Beginner",
			startingGold: 0,
			wantErr:      false,
		},
		{
			name:         "invalid empty id",
			id:           "",
			merchantName: "Invalid",
			startingGold: 1000,
			wantErr:      true,
			errContains:  "merchant id cannot be empty",
		},
		{
			name:         "invalid empty name",
			id:           "player_003",
			merchantName: "",
			startingGold: 1000,
			wantErr:      true,
			errContains:  "merchant name cannot be empty",
		},
		{
			name:         "invalid negative gold",
			id:           "player_004",
			merchantName: "Poor Trader",
			startingGold: -100,
			wantErr:      true,
			errContains:  "starting gold must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merchant, err := NewPlayerMerchant(tt.id, tt.merchantName, tt.startingGold)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, merchant)
			} else {
				require.NoError(t, err)
				require.NotNil(t, merchant)
				assert.Equal(t, tt.id, merchant.ID)
				assert.Equal(t, tt.merchantName, merchant.Name)
				assert.Equal(t, tt.startingGold, merchant.Gold)
				assert.Equal(t, 50.0, merchant.Reputation)
				assert.Equal(t, 1, merchant.Level)
				assert.Equal(t, 0, merchant.Experience)
				assert.NotNil(t, merchant.Inventory)
			}
		})
	}
}

func TestPlayerMerchant_BuyItem(t *testing.T) {
	merchant, err := NewPlayerMerchant("player_001", "Test Trader", 1000)
	require.NoError(t, err)

	testItem, err := item.NewItem("apple", "Red Apple", item.CategoryFruit, 10)
	require.NoError(t, err)

	tests := []struct {
		name         string
		item         *item.Item
		quantity     int
		pricePerUnit int
		wantErr      bool
		errContains  string
		expectedGold int
		expectedQty  int
	}{
		{
			name:         "successful purchase",
			item:         testItem,
			quantity:     5,
			pricePerUnit: 10,
			wantErr:      false,
			expectedGold: 950, // 1000 - (5 * 10)
			expectedQty:  5,
		},
		{
			name:         "insufficient gold",
			item:         testItem,
			quantity:     100,
			pricePerUnit: 20,
			wantErr:      true,
			errContains:  "insufficient gold",
			expectedGold: 950, // Should remain unchanged
			expectedQty:  5,   // Should remain unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := merchant.BuyItem(tt.item, tt.quantity, tt.pricePerUnit)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedGold, merchant.GetGold())
			assert.Equal(t, tt.expectedQty, merchant.Inventory.GetQuantity(tt.item.ID))
		})
	}
}

func TestPlayerMerchant_SellItem(t *testing.T) {
	merchant, err := NewPlayerMerchant("player_001", "Test Trader", 500)
	require.NoError(t, err)

	// Add items to inventory first
	testItem, err := item.NewItem("sword", "Iron Sword", item.CategoryWeapon, 100)
	require.NoError(t, err)
	err = merchant.Inventory.AddItem(testItem, 10)
	require.NoError(t, err)

	tests := []struct {
		name         string
		itemID       string
		quantity     int
		pricePerUnit int
		wantErr      bool
		errContains  string
		expectedGold int
		expectedQty  int
	}{
		{
			name:         "successful sale",
			itemID:       "sword",
			quantity:     3,
			pricePerUnit: 120,
			wantErr:      false,
			expectedGold: 860, // 500 + (3 * 120)
			expectedQty:  7,   // 10 - 3
		},
		{
			name:         "insufficient items",
			itemID:       "sword",
			quantity:     20,
			pricePerUnit: 120,
			wantErr:      true,
			errContains:  "insufficient items",
			expectedGold: 860, // Should remain unchanged
			expectedQty:  7,   // Should remain unchanged
		},
		{
			name:         "non-existent item",
			itemID:       "potion",
			quantity:     1,
			pricePerUnit: 50,
			wantErr:      true,
			errContains:  "insufficient items",
			expectedGold: 860, // Should remain unchanged
			expectedQty:  7,   // Should remain unchanged from sword
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := merchant.SellItem(tt.itemID, tt.quantity, tt.pricePerUnit)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedGold, merchant.GetGold())
			assert.Equal(t, tt.expectedQty, merchant.Inventory.GetQuantity("sword"))
		})
	}
}

func TestPlayerMerchant_LevelUp(t *testing.T) {
	merchant, err := NewPlayerMerchant("player_001", "Test Trader", 1000)
	require.NoError(t, err)

	// Initial state
	assert.Equal(t, 1, merchant.GetLevel())
	assert.Equal(t, 0, merchant.Experience)

	// Add experience that should trigger level up
	testItem, err := item.NewItem("gem", "Ruby", item.CategoryGem, 500)
	require.NoError(t, err)

	// Buy 100 items to gain 1000 experience (10 * 100)
	err = merchant.BuyItem(testItem, 100, 5)
	require.NoError(t, err)

	// Should level up from 1 to 2 (requires 1000 exp)
	assert.Equal(t, 2, merchant.GetLevel())
	assert.Equal(t, 0, merchant.Experience) // Reset after level up

	// Sell items to gain more experience
	err = merchant.SellItem(testItem.ID, 50, 10)
	require.NoError(t, err)

	// Should have 750 experience (15 * 50)
	assert.Equal(t, 750, merchant.Experience)
	assert.Equal(t, 2, merchant.GetLevel()) // Still level 2
}

func TestPlayerMerchant_Reputation(t *testing.T) {
	merchant, err := NewPlayerMerchant("player_001", "Test Trader", 1000)
	require.NoError(t, err)

	// Initial reputation
	assert.Equal(t, 50.0, merchant.GetReputation())

	// Make a successful sale to increase reputation
	testItem, err := item.NewItem("potion", "Health Potion", item.CategoryPotion, 50)
	require.NoError(t, err)
	err = merchant.Inventory.AddItem(testItem, 10)
	require.NoError(t, err)

	err = merchant.SellItem(testItem.ID, 5, 60)
	require.NoError(t, err)

	// Reputation should increase
	assert.Equal(t, 51.0, merchant.GetReputation())

	// Test reputation cap
	merchant.Reputation = 99.5
	err = merchant.SellItem(testItem.ID, 1, 60)
	require.NoError(t, err)
	assert.Equal(t, 100.0, merchant.GetReputation()) // Should cap at 100
}

func TestPlayerMerchant_Stats(t *testing.T) {
	merchant, err := NewPlayerMerchant("player_001", "Test Trader", 1000)
	require.NoError(t, err)

	// Initial stats
	stats := merchant.GetStats()
	assert.Equal(t, 0, stats.TotalProfit)
	assert.Equal(t, 0, stats.SuccessfulDeals)
	assert.Equal(t, 0, stats.FailedDeals)
	assert.Equal(t, 0, stats.TotalVolume)
	assert.Equal(t, 0, stats.ItemsSold)
	assert.Equal(t, 0, stats.ItemsBought)

	// Make a purchase
	testItem, err := item.NewItem("apple", "Red Apple", item.CategoryFruit, 10)
	require.NoError(t, err)
	err = merchant.BuyItem(testItem, 5, 10)
	require.NoError(t, err)

	stats = merchant.GetStats()
	assert.Equal(t, 5, stats.ItemsBought)
	assert.Equal(t, 50, stats.TotalVolume) // 5 * 10
	assert.Equal(t, 1, stats.SuccessfulDeals)

	// Make a sale
	err = merchant.SellItem(testItem.ID, 3, 15)
	require.NoError(t, err)

	stats = merchant.GetStats()
	assert.Equal(t, 3, stats.ItemsSold)
	assert.Equal(t, 95, stats.TotalVolume)    // 50 + (3 * 15)
	assert.Equal(t, 45, stats.TotalProfit)    // 3 * 15
	assert.Equal(t, 2, stats.SuccessfulDeals) // Buy + Sell
}

func TestPlayerMerchant_CanAfford(t *testing.T) {
	merchant, err := NewPlayerMerchant("player_001", "Test Trader", 500)
	require.NoError(t, err)

	assert.True(t, merchant.CanAfford(500))
	assert.True(t, merchant.CanAfford(100))
	assert.False(t, merchant.CanAfford(501))
	assert.False(t, merchant.CanAfford(1000))
}

func TestPlayerMerchant_GetInventoryValue(t *testing.T) {
	merchant, err := NewPlayerMerchant("player_001", "Test Trader", 1000)
	require.NoError(t, err)

	// Add various items
	apple, err := item.NewItem("apple", "Red Apple", item.CategoryFruit, 10)
	require.NoError(t, err)
	sword, err := item.NewItem("sword", "Iron Sword", item.CategoryWeapon, 100)
	require.NoError(t, err)

	err = merchant.Inventory.AddItem(apple, 10)
	require.NoError(t, err)
	err = merchant.Inventory.AddItem(sword, 2)
	require.NoError(t, err)

	// Create price map
	prices := map[string]int{
		"apple": 12,  // Current price higher than base
		"sword": 110, // Current price higher than base
	}

	value := merchant.GetInventoryValue(prices)
	assert.Equal(t, 340, value) // (10 * 12) + (2 * 110)
}

func TestPlayerMerchant_GetNetWorth(t *testing.T) {
	merchant, err := NewPlayerMerchant("player_001", "Test Trader", 500)
	require.NoError(t, err)

	// Add items
	gem, err := item.NewItem("ruby", "Ruby", item.CategoryGem, 200)
	require.NoError(t, err)
	err = merchant.Inventory.AddItem(gem, 3)
	require.NoError(t, err)

	prices := map[string]int{
		"ruby": 250,
	}

	netWorth := merchant.GetNetWorth(prices)
	assert.Equal(t, 1250, netWorth) // 500 gold + (3 * 250)
}

func TestPlayerMerchant_Reset(t *testing.T) {
	merchant, err := NewPlayerMerchant("player_001", "Test Trader", 1000)
	require.NoError(t, err)

	// Modify merchant state
	testItem, err := item.NewItem("potion", "Health Potion", item.CategoryPotion, 50)
	require.NoError(t, err)
	err = merchant.BuyItem(testItem, 5, 40)
	require.NoError(t, err)

	merchant.Level = 5
	merchant.Experience = 2500
	merchant.Reputation = 75.0

	// Reset
	merchant.Reset(1500)

	// Check reset state
	assert.Equal(t, 1500, merchant.GetGold())
	assert.Equal(t, 50.0, merchant.GetReputation())
	assert.Equal(t, 1, merchant.GetLevel())
	assert.Equal(t, 0, merchant.Experience)
	assert.Equal(t, 0, merchant.Inventory.GetQuantity(testItem.ID))

	stats := merchant.GetStats()
	assert.Equal(t, 0, stats.TotalProfit)
	assert.Equal(t, 0, stats.SuccessfulDeals)
	assert.Equal(t, 0, stats.ItemsBought)
	assert.Equal(t, 0, stats.ItemsSold)
}

func TestPlayerMerchant_ConcurrentAccess(t *testing.T) {
	merchant, err := NewPlayerMerchant("player_001", "Test Trader", 10000)
	require.NoError(t, err)

	testItem, err := item.NewItem("gem", "Diamond", item.CategoryGem, 1000)
	require.NoError(t, err)
	err = merchant.Inventory.AddItem(testItem, 100)
	require.NoError(t, err)

	// Run concurrent operations
	done := make(chan bool, 4)

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			_ = merchant.GetGold()
			_ = merchant.GetLevel()
			_ = merchant.GetReputation()
		}
		done <- true
	}()

	// Concurrent buys
	go func() {
		for i := 0; i < 10; i++ {
			_ = merchant.BuyItem(testItem, 1, 100)
		}
		done <- true
	}()

	// Concurrent sells
	go func() {
		for i := 0; i < 10; i++ {
			_ = merchant.SellItem(testItem.ID, 1, 150)
		}
		done <- true
	}()

	// Concurrent stats reads
	go func() {
		for i := 0; i < 100; i++ {
			_ = merchant.GetStats()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}

	// Verify merchant is still in valid state
	assert.GreaterOrEqual(t, merchant.GetGold(), 0)
	assert.GreaterOrEqual(t, merchant.GetLevel(), 1)
	assert.GreaterOrEqual(t, merchant.GetReputation(), 0.0)
	assert.LessOrEqual(t, merchant.GetReputation(), 100.0)
}
