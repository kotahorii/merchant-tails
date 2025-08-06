package inventory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/merchant-tails/game/internal/domain/item"
)

func TestNewInventoryManager(t *testing.T) {
	tests := []struct {
		name         string
		shopCapacity int
		warehouseCap int
		wantErr      bool
		errContains  string
	}{
		{
			name:         "create valid inventory manager",
			shopCapacity: 20,
			warehouseCap: 100,
			wantErr:      false,
		},
		{
			name:         "invalid shop capacity",
			shopCapacity: -1,
			warehouseCap: 100,
			wantErr:      true,
			errContains:  "shop capacity must be positive",
		},
		{
			name:         "invalid warehouse capacity",
			shopCapacity: 20,
			warehouseCap: 0,
			wantErr:      true,
			errContains:  "warehouse capacity must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewInventoryManager(tt.shopCapacity, tt.warehouseCap)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, manager)
			} else {
				require.NoError(t, err)
				require.NotNil(t, manager)
				assert.Equal(t, tt.shopCapacity, manager.ShopCapacity)
				assert.Equal(t, tt.warehouseCap, manager.WarehouseCapacity)
				assert.NotNil(t, manager.ShopInventory)
				assert.NotNil(t, manager.WarehouseInventory)
			}
		})
	}
}

func TestInventoryManager_AddToShop(t *testing.T) {
	manager, _ := NewInventoryManager(20, 100)
	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)

	tests := []struct {
		name          string
		item          *item.Item
		quantity      int
		expectedQty   int
		shouldSucceed bool
	}{
		{
			name:          "add item within capacity",
			item:          apple,
			quantity:      5,
			expectedQty:   5,
			shouldSucceed: true,
		},
		{
			name:          "add more of same item",
			item:          apple,
			quantity:      3,
			expectedQty:   8,
			shouldSucceed: true,
		},
		{
			name:          "exceed shop capacity",
			item:          apple,
			quantity:      15,
			expectedQty:   8, // Should remain unchanged
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.AddToShop(tt.item, tt.quantity)

			if tt.shouldSucceed {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			actual := manager.GetShopQuantity(tt.item.ID)
			assert.Equal(t, tt.expectedQty, actual)
		})
	}
}

func TestInventoryManager_TransferToWarehouse(t *testing.T) {
	manager, _ := NewInventoryManager(20, 100)
	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)

	// Setup: Add items to shop
	_ = manager.AddToShop(apple, 10)

	tests := []struct {
		name              string
		itemID            string
		quantity          int
		expectedShop      int
		expectedWarehouse int
		shouldSucceed     bool
	}{
		{
			name:              "transfer partial quantity",
			itemID:            "apple_001",
			quantity:          3,
			expectedShop:      7,
			expectedWarehouse: 3,
			shouldSucceed:     true,
		},
		{
			name:              "transfer remaining quantity",
			itemID:            "apple_001",
			quantity:          7,
			expectedShop:      0,
			expectedWarehouse: 10,
			shouldSucceed:     true,
		},
		{
			name:              "transfer from empty shop",
			itemID:            "apple_001",
			quantity:          1,
			expectedShop:      0,
			expectedWarehouse: 10,
			shouldSucceed:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.TransferToWarehouse(tt.itemID, tt.quantity)

			if tt.shouldSucceed {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			shopQty := manager.GetShopQuantity(tt.itemID)
			warehouseQty := manager.GetWarehouseQuantity(tt.itemID)

			assert.Equal(t, tt.expectedShop, shopQty)
			assert.Equal(t, tt.expectedWarehouse, warehouseQty)
		})
	}
}

func TestInventoryManager_TransferToShop(t *testing.T) {
	manager, _ := NewInventoryManager(20, 100)
	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)

	// Setup: Add items to warehouse
	_ = manager.AddToWarehouse(apple, 15)

	tests := []struct {
		name              string
		itemID            string
		quantity          int
		expectedShop      int
		expectedWarehouse int
		shouldSucceed     bool
	}{
		{
			name:              "transfer within shop capacity",
			itemID:            "apple_001",
			quantity:          5,
			expectedShop:      5,
			expectedWarehouse: 10,
			shouldSucceed:     true,
		},
		{
			name:              "transfer more within capacity",
			itemID:            "apple_001",
			quantity:          10,
			expectedShop:      15,
			expectedWarehouse: 0,
			shouldSucceed:     true,
		},
		{
			name:              "exceed shop capacity",
			itemID:            "apple_001",
			quantity:          10,
			expectedShop:      15,
			expectedWarehouse: 0,
			shouldSucceed:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.TransferToShop(tt.itemID, tt.quantity)

			if tt.shouldSucceed {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			shopQty := manager.GetShopQuantity(tt.itemID)
			warehouseQty := manager.GetWarehouseQuantity(tt.itemID)

			assert.Equal(t, tt.expectedShop, shopQty)
			assert.Equal(t, tt.expectedWarehouse, warehouseQty)
		})
	}
}

func TestInventoryManager_OptimizePlacement(t *testing.T) {
	manager, _ := NewInventoryManager(20, 100)

	// Create items with different priorities
	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	sword, _ := item.NewItem("sword_001", "Iron Sword", item.CategoryWeapon, 200)
	potion, _ := item.NewItem("potion_001", "Health Potion", item.CategoryPotion, 50)

	// Add items to warehouse
	_ = manager.AddToWarehouse(apple, 10)
	_ = manager.AddToWarehouse(sword, 5)
	_ = manager.AddToWarehouse(potion, 8)

	// Set sales velocity (simulating which items sell faster)
	manager.SetSalesVelocity("apple_001", 5.0)
	manager.SetSalesVelocity("sword_001", 1.0)
	manager.SetSalesVelocity("potion_001", 3.0)

	// Optimize placement
	manager.OptimizePlacement()

	// High velocity items should be in shop
	assert.Greater(t, manager.GetShopQuantity("apple_001"), 0, "High velocity apple should be in shop")
	assert.Greater(t, manager.GetShopQuantity("potion_001"), 0, "Medium velocity potion should be in shop")

	// Check total capacity is not exceeded
	totalInShop := manager.GetTotalShopItems()
	assert.LessOrEqual(t, totalInShop, 20, "Shop capacity should not be exceeded")
}

func TestInventoryManager_HandleSpoilage(t *testing.T) {
	manager, _ := NewInventoryManager(20, 100)

	// Create perishable and non-perishable items
	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	apple.Durability = 3 // Will spoil in 3 days

	sword, _ := item.NewItem("sword_001", "Iron Sword", item.CategoryWeapon, 200)
	sword.Durability = -1 // Never spoils

	// Add items
	_ = manager.AddToShop(apple, 5)
	_ = manager.AddToShop(sword, 2)
	_ = manager.AddToWarehouse(apple, 10)

	// Simulate passage of time
	for i := 0; i < 3; i++ {
		manager.ProcessDailyUpdate()
	}

	// Check spoilage
	spoiledItems := manager.GetSpoiledItems()
	assert.Greater(t, len(spoiledItems), 0, "Should have spoiled items")

	// Swords should not spoil
	assert.Equal(t, 2, manager.GetShopQuantity("sword_001"), "Swords should not spoil")
}

func TestSellStrategy_DetermineSellPriority(t *testing.T) {
	strategies := []SellStrategy{
		&FIFOStrategy{},
		&LIFOStrategy{},
		&ProfitMaximizationStrategy{},
		&VelocityBasedStrategy{},
	}

	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	sword, _ := item.NewItem("sword_001", "Iron Sword", item.CategoryWeapon, 200)

	items := []*InventoryItem{
		{
			Item:          apple,
			Quantity:      10,
			PurchaseDate:  time.Now().Add(-5 * 24 * time.Hour),
			PurchasePrice: 8,
		},
		{
			Item:          sword,
			Quantity:      2,
			PurchaseDate:  time.Now().Add(-2 * 24 * time.Hour),
			PurchasePrice: 180,
		},
	}

	for _, strategy := range strategies {
		t.Run(strategy.GetName(), func(t *testing.T) {
			priority := strategy.DetermineSellPriority(items, 15)
			assert.NotNil(t, priority)
			assert.GreaterOrEqual(t, len(priority), 0)
		})
	}
}

func TestInventoryManager_GetLowStockItems(t *testing.T) {
	manager, _ := NewInventoryManager(20, 100)

	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	potion, _ := item.NewItem("potion_001", "Health Potion", item.CategoryPotion, 50)

	// Set minimum stock levels
	manager.SetMinimumStock("apple_001", 10)
	manager.SetMinimumStock("potion_001", 5)

	// Add items below minimum
	_ = manager.AddToShop(apple, 3)
	_ = manager.AddToShop(potion, 6)

	lowStock := manager.GetLowStockItems()

	assert.Len(t, lowStock, 1, "Should have one low stock item")
	assert.Equal(t, "apple_001", lowStock[0].Item.ID)
	assert.Equal(t, 10, lowStock[0].MinimumStock)
	assert.Equal(t, 3, lowStock[0].CurrentStock)
}

func TestInventoryManager_CalculateRestockQuantity(t *testing.T) {
	manager, _ := NewInventoryManager(20, 100)

	tests := []struct {
		name          string
		itemID        string
		currentStock  int
		salesVelocity float64
		availableGold int
		unitPrice     int
		expectedQty   int
	}{
		{
			name:          "normal restock",
			itemID:        "apple_001",
			currentStock:  5,
			salesVelocity: 3.0,
			availableGold: 100,
			unitPrice:     10,
			expectedQty:   9, // Based on velocity and gold
		},
		{
			name:          "limited by gold",
			itemID:        "sword_001",
			currentStock:  1,
			salesVelocity: 2.0,
			availableGold: 50,
			unitPrice:     200,
			expectedQty:   0, // Can't afford any
		},
		{
			name:          "high velocity item",
			itemID:        "potion_001",
			currentStock:  2,
			salesVelocity: 10.0,
			availableGold: 500,
			unitPrice:     50,
			expectedQty:   10, // Max based on gold
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager.SetSalesVelocity(tt.itemID, tt.salesVelocity)
			qty := manager.CalculateRestockQuantity(
				tt.itemID,
				tt.currentStock,
				tt.availableGold,
				tt.unitPrice,
			)
			assert.Equal(t, tt.expectedQty, qty)
		})
	}
}

func TestInventorySnapshot(t *testing.T) {
	manager, _ := NewInventoryManager(20, 100)

	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	sword, _ := item.NewItem("sword_001", "Iron Sword", item.CategoryWeapon, 200)

	_ = manager.AddToShop(apple, 5)
	_ = manager.AddToWarehouse(sword, 2)

	// Create snapshot
	snapshot := manager.CreateSnapshot()

	assert.NotNil(t, snapshot)
	assert.Equal(t, 5, snapshot.ShopItems["apple_001"])
	assert.Equal(t, 2, snapshot.WarehouseItems["sword_001"])
	assert.Equal(t, 50, snapshot.TotalValue) // 5*10 for apples
	assert.NotZero(t, snapshot.Timestamp)

	// Restore from snapshot
	newManager, _ := NewInventoryManager(20, 100)
	err := newManager.RestoreFromSnapshot(snapshot)

	require.NoError(t, err)
	assert.Equal(t, 5, newManager.GetShopQuantity("apple_001"))
	assert.Equal(t, 2, newManager.GetWarehouseQuantity("sword_001"))
}

func TestInventoryManager_GetTurnoverRate(t *testing.T) {
	manager, _ := NewInventoryManager(20, 100)

	// Set up sales history
	manager.RecordSale("apple_001", 50, 30) // 50 sold over 30 days
	manager.RecordSale("sword_001", 5, 30)  // 5 sold over 30 days

	apple, _ := item.NewItem("apple_001", "Apple", item.CategoryFruit, 10)
	sword, _ := item.NewItem("sword_001", "Iron Sword", item.CategoryWeapon, 200)

	_ = manager.AddToShop(apple, 10)
	_ = manager.AddToShop(sword, 2)

	appleTurnover := manager.GetTurnoverRate("apple_001")
	swordTurnover := manager.GetTurnoverRate("sword_001")

	assert.Greater(t, appleTurnover, swordTurnover, "Apples should have higher turnover")
	assert.Greater(t, appleTurnover, 1.0, "Apple turnover should be > 1")
}
