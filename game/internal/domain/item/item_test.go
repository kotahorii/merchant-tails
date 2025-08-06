package item

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewItem(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		itemName    string
		category    Category
		basePrice   int
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid fruit item",
			id:        "apple_001",
			itemName:  "Apple",
			category:  CategoryFruit,
			basePrice: 10,
			wantErr:   false,
		},
		{
			name:      "valid potion item",
			id:        "potion_001",
			itemName:  "Health Potion",
			category:  CategoryPotion,
			basePrice: 50,
			wantErr:   false,
		},
		{
			name:        "invalid empty id",
			id:          "",
			itemName:    "Invalid Item",
			category:    CategoryWeapon,
			basePrice:   100,
			wantErr:     true,
			errContains: "item id cannot be empty",
		},
		{
			name:        "invalid empty name",
			id:          "item_001",
			itemName:    "",
			category:    CategoryWeapon,
			basePrice:   100,
			wantErr:     true,
			errContains: "item name cannot be empty",
		},
		{
			name:        "invalid negative price",
			id:          "item_001",
			itemName:    "Negative Item",
			category:    CategoryWeapon,
			basePrice:   -10,
			wantErr:     true,
			errContains: "base price must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := NewItem(tt.id, tt.itemName, tt.category, tt.basePrice)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, item)
			} else {
				require.NoError(t, err)
				require.NotNil(t, item)
				assert.Equal(t, tt.id, item.ID)
				assert.Equal(t, tt.itemName, item.Name)
				assert.Equal(t, tt.category, item.Category)
				assert.Equal(t, tt.basePrice, item.BasePrice)
			}
		})
	}
}

func TestItem_CalculatePrice(t *testing.T) {
	item, err := NewItem("apple_001", "Apple", CategoryFruit, 10)
	require.NoError(t, err)

	tests := []struct {
		name           string
		demandModifier float64
		seasonModifier float64
		expectedPrice  int
	}{
		{
			name:           "normal conditions",
			demandModifier: 1.0,
			seasonModifier: 1.0,
			expectedPrice:  10,
		},
		{
			name:           "high demand",
			demandModifier: 1.5,
			seasonModifier: 1.0,
			expectedPrice:  15,
		},
		{
			name:           "low demand in winter",
			demandModifier: 0.8,
			seasonModifier: 0.9,
			expectedPrice:  7,
		},
		{
			name:           "peak season high demand",
			demandModifier: 1.3,
			seasonModifier: 1.2,
			expectedPrice:  15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price := item.CalculatePrice(tt.demandModifier, tt.seasonModifier)
			assert.Equal(t, tt.expectedPrice, price)
		})
	}
}

func TestItem_UpdateDurability(t *testing.T) {
	tests := []struct {
		name             string
		category         Category
		initialDuration  int
		daysToPass       int
		expectedDuration int
		shouldSpoil      bool
	}{
		{
			name:             "fruit spoils after 3 days",
			category:         CategoryFruit,
			initialDuration:  3,
			daysToPass:       3,
			expectedDuration: 0,
			shouldSpoil:      true,
		},
		{
			name:             "fruit partially spoiled",
			category:         CategoryFruit,
			initialDuration:  3,
			daysToPass:       2,
			expectedDuration: 1,
			shouldSpoil:      false,
		},
		{
			name:             "weapon doesn't spoil",
			category:         CategoryWeapon,
			initialDuration:  -1, // infinite durability
			daysToPass:       100,
			expectedDuration: -1,
			shouldSpoil:      false,
		},
		{
			name:             "potion has long shelf life",
			category:         CategoryPotion,
			initialDuration:  30,
			daysToPass:       10,
			expectedDuration: 20,
			shouldSpoil:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := NewItem("test_001", "Test Item", tt.category, 100)
			require.NoError(t, err)

			item.Durability = tt.initialDuration

			for i := 0; i < tt.daysToPass; i++ {
				item.UpdateDurability()
			}

			assert.Equal(t, tt.expectedDuration, item.Durability)
			assert.Equal(t, tt.shouldSpoil, item.IsSpoiled())
		})
	}
}

func TestItem_GetVolatility(t *testing.T) {
	tests := []struct {
		name               string
		category           Category
		expectedVolatility float32
	}{
		{
			name:               "fruit has low volatility",
			category:           CategoryFruit,
			expectedVolatility: 0.1,
		},
		{
			name:               "potion has medium volatility",
			category:           CategoryPotion,
			expectedVolatility: 0.3,
		},
		{
			name:               "weapon has very low volatility",
			category:           CategoryWeapon,
			expectedVolatility: 0.05,
		},
		{
			name:               "accessory has high volatility",
			category:           CategoryAccessory,
			expectedVolatility: 0.5,
		},
		{
			name:               "magic book has low volatility",
			category:           CategoryMagicBook,
			expectedVolatility: 0.1,
		},
		{
			name:               "gem has very high volatility",
			category:           CategoryGem,
			expectedVolatility: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := NewItem("test_001", "Test Item", tt.category, 100)
			require.NoError(t, err)

			volatility := item.GetVolatility()
			assert.Equal(t, tt.expectedVolatility, volatility)
		})
	}
}

func TestItemInventory_AddItem(t *testing.T) {
	inventory := NewInventory()

	apple, err := NewItem("apple_001", "Apple", CategoryFruit, 10)
	require.NoError(t, err)

	t.Run("add new item", func(t *testing.T) {
		err := inventory.AddItem(apple, 5)
		require.NoError(t, err)
		assert.Equal(t, 5, inventory.GetQuantity("apple_001"))
	})

	t.Run("add more of existing item", func(t *testing.T) {
		err := inventory.AddItem(apple, 3)
		require.NoError(t, err)
		assert.Equal(t, 8, inventory.GetQuantity("apple_001"))
	})

	t.Run("cannot add negative quantity", func(t *testing.T) {
		err := inventory.AddItem(apple, -1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "quantity must be positive")
	})
}

func TestItemInventory_RemoveItem(t *testing.T) {
	inventory := NewInventory()

	apple, err := NewItem("apple_001", "Apple", CategoryFruit, 10)
	require.NoError(t, err)

	err = inventory.AddItem(apple, 10)
	require.NoError(t, err)

	t.Run("remove partial quantity", func(t *testing.T) {
		err := inventory.RemoveItem("apple_001", 3)
		require.NoError(t, err)
		assert.Equal(t, 7, inventory.GetQuantity("apple_001"))
	})

	t.Run("cannot remove more than available", func(t *testing.T) {
		err := inventory.RemoveItem("apple_001", 10) // Try to remove 10 when only 7 available
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient quantity")
		assert.Equal(t, 7, inventory.GetQuantity("apple_001")) // Quantity unchanged
	})

	t.Run("remove all quantity", func(t *testing.T) {
		err := inventory.RemoveItem("apple_001", 7)
		require.NoError(t, err)
		assert.Equal(t, 0, inventory.GetQuantity("apple_001"))
	})

	t.Run("cannot remove from zero quantity item", func(t *testing.T) {
		// After removing all quantity, the item is deleted from inventory
		err := inventory.RemoveItem("apple_001", 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "item not found")
	})

	t.Run("cannot remove from non-existent item", func(t *testing.T) {
		err := inventory.RemoveItem("nonexistent", 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "item not found")
	})
}

func TestItemMaster_GetSeasonalModifier(t *testing.T) {
	master := &ItemMaster{
		ID:        "apple_001",
		Name:      "Apple",
		Category:  CategoryFruit,
		BasePrice: 10,
		SeasonalModifiers: map[Season]float32{
			SeasonSpring: 1.2,
			SeasonSummer: 1.0,
			SeasonAutumn: 1.5,
			SeasonWinter: 0.8,
		},
	}

	tests := []struct {
		name     string
		season   Season
		expected float32
	}{
		{"spring bonus", SeasonSpring, 1.2},
		{"summer normal", SeasonSummer, 1.0},
		{"autumn peak", SeasonAutumn, 1.5},
		{"winter penalty", SeasonWinter, 0.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := master.GetSeasonalModifier(tt.season)
			assert.Equal(t, tt.expected, modifier)
		})
	}
}

func TestItemPriceHistory(t *testing.T) {
	history := NewPriceHistory(5)

	t.Run("add price records", func(t *testing.T) {
		history.AddRecord(100, time.Now())
		history.AddRecord(110, time.Now().Add(time.Hour))
		history.AddRecord(105, time.Now().Add(2*time.Hour))

		assert.Equal(t, 3, len(history.Records))
		assert.Equal(t, 105, history.GetLatestPrice())
	})

	t.Run("calculate average price", func(t *testing.T) {
		avg := history.GetAveragePrice()
		expected := (100 + 110 + 105) / 3
		assert.Equal(t, expected, avg)
	})

	t.Run("maintain max size", func(t *testing.T) {
		history.AddRecord(120, time.Now().Add(3*time.Hour))
		history.AddRecord(115, time.Now().Add(4*time.Hour))
		history.AddRecord(125, time.Now().Add(5*time.Hour))

		assert.Equal(t, 5, len(history.Records))
		assert.Equal(t, 110, history.Records[0].Price) // First record (100) should be removed
	})

	t.Run("calculate price trend", func(t *testing.T) {
		trend := history.GetPriceTrend()
		assert.Equal(t, PriceTrendUp, trend) // Price went from 110 to 125
	})
}
