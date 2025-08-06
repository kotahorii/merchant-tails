package inventory

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/item"
)

// InventoryManager manages shop and warehouse inventories
type InventoryManager struct {
	ShopCapacity       int
	WarehouseCapacity  int
	ShopInventory      *item.Inventory
	WarehouseInventory *item.Inventory
	shopItems          map[string]*InventoryItem // Internal tracking
	warehouseItems     map[string]*InventoryItem // Internal tracking
	salesVelocity      map[string]float64
	minimumStock       map[string]int
	salesHistory       map[string]*SalesHistory
	spoiledItems       []*SpoiledItem
	mu                 sync.RWMutex
}

// InventoryItem represents an item with metadata
type InventoryItem struct {
	Item          *item.Item
	Quantity      int
	PurchaseDate  time.Time
	PurchasePrice int
	Location      InventoryLocation
}

// InventoryLocation represents where an item is stored
type InventoryLocation int

const (
	LocationShop InventoryLocation = iota
	LocationWarehouse
)

// SpoiledItem tracks spoiled items
type SpoiledItem struct {
	Item     *item.Item
	Quantity int
	Location InventoryLocation
	Date     time.Time
}

// SalesHistory tracks sales data for an item
type SalesHistory struct {
	TotalSold   int
	DaysTracked int
	LastSale    time.Time
}

// LowStockAlert represents an item below minimum stock
type LowStockAlert struct {
	Item         *item.Item
	CurrentStock int
	MinimumStock int
	Location     InventoryLocation
}

// InventorySnapshot represents a point-in-time inventory state
type InventorySnapshot struct {
	ShopItems      map[string]int
	WarehouseItems map[string]int
	TotalValue     int
	Timestamp      time.Time
}

// SellStrategy interface for different selling strategies
type SellStrategy interface {
	DetermineSellPriority(items []*InventoryItem, currentPrice int) []*InventoryItem
	GetName() string
}

// NewInventoryManager creates a new inventory manager
func NewInventoryManager(shopCapacity, warehouseCapacity int) (*InventoryManager, error) {
	if shopCapacity <= 0 {
		return nil, errors.New("shop capacity must be positive")
	}
	if warehouseCapacity <= 0 {
		return nil, errors.New("warehouse capacity must be positive")
	}

	return &InventoryManager{
		ShopCapacity:       shopCapacity,
		WarehouseCapacity:  warehouseCapacity,
		ShopInventory:      item.NewInventory(),
		WarehouseInventory: item.NewInventory(),
		shopItems:          make(map[string]*InventoryItem),
		warehouseItems:     make(map[string]*InventoryItem),
		salesVelocity:      make(map[string]float64),
		minimumStock:       make(map[string]int),
		salesHistory:       make(map[string]*SalesHistory),
		spoiledItems:       make([]*SpoiledItem, 0),
	}, nil
}

// AddToShop adds items to shop inventory
func (im *InventoryManager) AddToShop(item *item.Item, quantity int) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	currentTotal := im.getTotalShopItemsUnsafe()
	if currentTotal+quantity > im.ShopCapacity {
		return fmt.Errorf("exceeds shop capacity: current %d + new %d > capacity %d",
			currentTotal, quantity, im.ShopCapacity)
	}

	err := im.ShopInventory.AddItem(item, quantity)
	if err == nil {
		// Track internally
		if existing, exists := im.shopItems[item.ID]; exists {
			existing.Quantity += quantity
		} else {
			im.shopItems[item.ID] = &InventoryItem{
				Item:         item,
				Quantity:     quantity,
				PurchaseDate: time.Now(),
				Location:     LocationShop,
			}
		}
	}
	return err
}

// AddToWarehouse adds items to warehouse inventory
func (im *InventoryManager) AddToWarehouse(item *item.Item, quantity int) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	currentTotal := im.getTotalWarehouseItemsUnsafe()
	if currentTotal+quantity > im.WarehouseCapacity {
		return fmt.Errorf("exceeds warehouse capacity: current %d + new %d > capacity %d",
			currentTotal, quantity, im.WarehouseCapacity)
	}

	err := im.WarehouseInventory.AddItem(item, quantity)
	if err == nil {
		// Track internally
		if existing, exists := im.warehouseItems[item.ID]; exists {
			existing.Quantity += quantity
		} else {
			im.warehouseItems[item.ID] = &InventoryItem{
				Item:         item,
				Quantity:     quantity,
				PurchaseDate: time.Now(),
				Location:     LocationWarehouse,
			}
		}
	}
	return err
}

// TransferToWarehouse moves items from shop to warehouse
func (im *InventoryManager) TransferToWarehouse(itemID string, quantity int) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	shopQty := im.ShopInventory.GetQuantity(itemID)
	if shopQty < quantity {
		return fmt.Errorf("insufficient quantity in shop: have %d, need %d", shopQty, quantity)
	}

	warehouseTotal := im.getTotalWarehouseItemsUnsafe()
	if warehouseTotal+quantity > im.WarehouseCapacity {
		return errors.New("exceeds warehouse capacity")
	}

	// Get the item reference
	var itemRef *item.Item
	if entry, exists := im.shopItems[itemID]; exists {
		itemRef = entry.Item
	} else {
		return errors.New("item not found in shop")
	}

	// Transfer
	if err := im.ShopInventory.RemoveItem(itemID, quantity); err != nil {
		return err
	}

	// Update internal tracking
	if entry, exists := im.shopItems[itemID]; exists {
		entry.Quantity -= quantity
		if entry.Quantity == 0 {
			delete(im.shopItems, itemID)
		}
	}

	err := im.WarehouseInventory.AddItem(itemRef, quantity)
	if err == nil {
		if existing, exists := im.warehouseItems[itemID]; exists {
			existing.Quantity += quantity
		} else {
			im.warehouseItems[itemID] = &InventoryItem{
				Item:         itemRef,
				Quantity:     quantity,
				PurchaseDate: time.Now(),
				Location:     LocationWarehouse,
			}
		}
	}
	return err
}

// TransferToShop moves items from warehouse to shop
func (im *InventoryManager) TransferToShop(itemID string, quantity int) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	warehouseQty := im.WarehouseInventory.GetQuantity(itemID)
	if warehouseQty < quantity {
		return fmt.Errorf("insufficient quantity in warehouse: have %d, need %d",
			warehouseQty, quantity)
	}

	shopTotal := im.getTotalShopItemsUnsafe()
	if shopTotal+quantity > im.ShopCapacity {
		return errors.New("exceeds shop capacity")
	}

	// Get the item reference
	var itemRef *item.Item
	if entry, exists := im.warehouseItems[itemID]; exists {
		itemRef = entry.Item
	} else {
		return errors.New("item not found in warehouse")
	}

	// Transfer
	if err := im.WarehouseInventory.RemoveItem(itemID, quantity); err != nil {
		return err
	}

	// Update internal tracking
	if entry, exists := im.warehouseItems[itemID]; exists {
		entry.Quantity -= quantity
		if entry.Quantity == 0 {
			delete(im.warehouseItems, itemID)
		}
	}

	err := im.ShopInventory.AddItem(itemRef, quantity)
	if err == nil {
		if existing, exists := im.shopItems[itemID]; exists {
			existing.Quantity += quantity
		} else {
			im.shopItems[itemID] = &InventoryItem{
				Item:         itemRef,
				Quantity:     quantity,
				PurchaseDate: time.Now(),
				Location:     LocationShop,
			}
		}
	}
	return err
}

// GetShopQuantity returns quantity of an item in shop
func (im *InventoryManager) GetShopQuantity(itemID string) int {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.ShopInventory.GetQuantity(itemID)
}

// GetWarehouseQuantity returns quantity of an item in warehouse
func (im *InventoryManager) GetWarehouseQuantity(itemID string) int {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.WarehouseInventory.GetQuantity(itemID)
}

// GetTotalShopItems returns total number of items in shop
func (im *InventoryManager) GetTotalShopItems() int {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.getTotalShopItemsUnsafe()
}

// getTotalShopItemsUnsafe returns total items without locking
func (im *InventoryManager) getTotalShopItemsUnsafe() int {
	total := 0
	for _, entry := range im.shopItems {
		total += entry.Quantity
	}
	return total
}

// getTotalWarehouseItemsUnsafe returns total items without locking
func (im *InventoryManager) getTotalWarehouseItemsUnsafe() int {
	total := 0
	for _, entry := range im.warehouseItems {
		total += entry.Quantity
	}
	return total
}

// SetSalesVelocity sets the sales velocity for an item
func (im *InventoryManager) SetSalesVelocity(itemID string, velocity float64) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.salesVelocity[itemID] = velocity
}

// GetSalesVelocity returns the sales velocity for an item
func (im *InventoryManager) GetSalesVelocity(itemID string) float64 {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.salesVelocity[itemID]
}

// OptimizePlacement optimizes item placement based on sales velocity
func (im *InventoryManager) OptimizePlacement() {
	im.mu.Lock()
	defer im.mu.Unlock()

	type itemPriority struct {
		itemID   string
		velocity float64
		item     *item.Item
		quantity int
	}

	// Collect all items from warehouse with their velocities
	priorities := make([]itemPriority, 0)
	for id, entry := range im.warehouseItems {
		if velocity, exists := im.salesVelocity[id]; exists && velocity > 0 {
			priorities = append(priorities, itemPriority{
				itemID:   id,
				velocity: velocity,
				item:     entry.Item,
				quantity: entry.Quantity,
			})
		}
	}

	// Sort by velocity (highest first)
	sort.Slice(priorities, func(i, j int) bool {
		return priorities[i].velocity > priorities[j].velocity
	})

	// Move high-velocity items to shop
	shopSpace := im.ShopCapacity - im.getTotalShopItemsUnsafe()
	for _, p := range priorities {
		if shopSpace <= 0 {
			break
		}

		// Calculate how many to move
		toMove := p.quantity
		if toMove > shopSpace {
			toMove = shopSpace
		}

		// Move items
		_ = im.WarehouseInventory.RemoveItem(p.itemID, toMove)
		_ = im.ShopInventory.AddItem(p.item, toMove)
		shopSpace -= toMove
	}
}

// ProcessDailyUpdate processes daily inventory updates
func (im *InventoryManager) ProcessDailyUpdate() {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Process spoilage for shop items
	for _, entry := range im.shopItems {
		entry.Item.UpdateDurability()
		if entry.Item.IsSpoiled() {
			im.spoiledItems = append(im.spoiledItems, &SpoiledItem{
				Item:     entry.Item,
				Quantity: entry.Quantity,
				Location: LocationShop,
				Date:     time.Now(),
			})
		}
	}

	// Process spoilage for warehouse items
	for _, entry := range im.warehouseItems {
		entry.Item.UpdateDurability()
		if entry.Item.IsSpoiled() {
			im.spoiledItems = append(im.spoiledItems, &SpoiledItem{
				Item:     entry.Item,
				Quantity: entry.Quantity,
				Location: LocationWarehouse,
				Date:     time.Now(),
			})
		}
	}
}

// GetSpoiledItems returns list of spoiled items
func (im *InventoryManager) GetSpoiledItems() []*SpoiledItem {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.spoiledItems
}

// SetMinimumStock sets minimum stock level for an item
func (im *InventoryManager) SetMinimumStock(itemID string, minimum int) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.minimumStock[itemID] = minimum
}

// GetLowStockItems returns items below minimum stock
func (im *InventoryManager) GetLowStockItems() []*LowStockAlert {
	im.mu.RLock()
	defer im.mu.RUnlock()

	alerts := make([]*LowStockAlert, 0)

	for itemID, minStock := range im.minimumStock {
		currentStock := im.ShopInventory.GetQuantity(itemID) +
			im.WarehouseInventory.GetQuantity(itemID)

		if currentStock < minStock {
			// Find the item
			var itemRef *item.Item
			if entry, exists := im.shopItems[itemID]; exists {
				itemRef = entry.Item
			} else if entry, exists := im.warehouseItems[itemID]; exists {
				itemRef = entry.Item
			}

			if itemRef != nil {
				alerts = append(alerts, &LowStockAlert{
					Item:         itemRef,
					CurrentStock: currentStock,
					MinimumStock: minStock,
					Location:     LocationShop,
				})
			}
		}
	}

	return alerts
}

// CalculateRestockQuantity calculates optimal restock quantity
func (im *InventoryManager) CalculateRestockQuantity(itemID string, currentStock, availableGold, unitPrice int) int {
	im.mu.RLock()
	defer im.mu.RUnlock()

	velocity := im.salesVelocity[itemID]
	if velocity == 0 {
		return 0
	}

	// Calculate based on velocity (3 days worth)
	targetStock := int(velocity * 3) // 3 days worth
	needed := targetStock

	if needed <= 0 {
		return 0
	}

	// Limit by available gold
	maxAffordable := availableGold / unitPrice
	if needed > maxAffordable {
		needed = maxAffordable
	}

	// Limit by capacity
	totalSpace := im.ShopCapacity + im.WarehouseCapacity
	currentTotal := im.getTotalShopItemsUnsafe() + im.getTotalWarehouseItemsUnsafe()
	availableSpace := totalSpace - currentTotal

	if needed > availableSpace {
		needed = availableSpace
	}

	return needed
}

// CreateSnapshot creates a snapshot of current inventory
func (im *InventoryManager) CreateSnapshot() *InventorySnapshot {
	im.mu.RLock()
	defer im.mu.RUnlock()

	snapshot := &InventorySnapshot{
		ShopItems:      make(map[string]int),
		WarehouseItems: make(map[string]int),
		TotalValue:     0,
		Timestamp:      time.Now(),
	}

	// Copy shop items
	for id, entry := range im.shopItems {
		snapshot.ShopItems[id] = entry.Quantity
		snapshot.TotalValue += entry.Item.BasePrice * entry.Quantity
	}

	// Copy warehouse items
	for id, entry := range im.warehouseItems {
		snapshot.WarehouseItems[id] = entry.Quantity
	}

	return snapshot
}

// RestoreFromSnapshot restores inventory from a snapshot
func (im *InventoryManager) RestoreFromSnapshot(snapshot *InventorySnapshot) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Clear current inventories
	im.ShopInventory = item.NewInventory()
	im.WarehouseInventory = item.NewInventory()

	// Restore shop items
	for itemID, quantity := range snapshot.ShopItems {
		// Note: In real implementation, would need item registry to get item details
		// For now, creating dummy items
		restoredItem, _ := item.NewItem(itemID, "Restored Item", item.CategoryFruit, 10)
		_ = im.ShopInventory.AddItem(restoredItem, quantity)
	}

	// Restore warehouse items
	for itemID, quantity := range snapshot.WarehouseItems {
		restoredItem, _ := item.NewItem(itemID, "Restored Item", item.CategoryWeapon, 200)
		_ = im.WarehouseInventory.AddItem(restoredItem, quantity)
	}

	return nil
}

// RecordSale records a sale for tracking purposes
func (im *InventoryManager) RecordSale(itemID string, quantity, days int) {
	im.mu.Lock()
	defer im.mu.Unlock()

	if _, exists := im.salesHistory[itemID]; !exists {
		im.salesHistory[itemID] = &SalesHistory{}
	}

	im.salesHistory[itemID].TotalSold = quantity
	im.salesHistory[itemID].DaysTracked = days
	im.salesHistory[itemID].LastSale = time.Now()

	// Update velocity
	if days > 0 {
		im.salesVelocity[itemID] = float64(quantity) / float64(days)
	}
}

// GetTurnoverRate calculates inventory turnover rate
func (im *InventoryManager) GetTurnoverRate(itemID string) float64 {
	im.mu.RLock()
	defer im.mu.RUnlock()

	history, exists := im.salesHistory[itemID]
	if !exists || history.DaysTracked == 0 {
		return 0
	}

	currentStock := im.ShopInventory.GetQuantity(itemID) +
		im.WarehouseInventory.GetQuantity(itemID)

	if currentStock == 0 {
		return 0
	}

	avgDailySales := float64(history.TotalSold) / float64(history.DaysTracked)
	turnover := avgDailySales / float64(currentStock) * 30 // Monthly turnover

	return turnover
}

// Sell Strategy Implementations

// FIFOStrategy sells oldest items first
type FIFOStrategy struct{}

func (s *FIFOStrategy) DetermineSellPriority(items []*InventoryItem, currentPrice int) []*InventoryItem {
	// Sort by purchase date (oldest first)
	sort.Slice(items, func(i, j int) bool {
		return items[i].PurchaseDate.Before(items[j].PurchaseDate)
	})
	return items
}

func (s *FIFOStrategy) GetName() string {
	return "FIFO"
}

// LIFOStrategy sells newest items first
type LIFOStrategy struct{}

func (s *LIFOStrategy) DetermineSellPriority(items []*InventoryItem, currentPrice int) []*InventoryItem {
	// Sort by purchase date (newest first)
	sort.Slice(items, func(i, j int) bool {
		return items[i].PurchaseDate.After(items[j].PurchaseDate)
	})
	return items
}

func (s *LIFOStrategy) GetName() string {
	return "LIFO"
}

// ProfitMaximizationStrategy sells highest profit items first
type ProfitMaximizationStrategy struct{}

func (s *ProfitMaximizationStrategy) DetermineSellPriority(items []*InventoryItem, currentPrice int) []*InventoryItem {
	// Sort by profit margin (highest first)
	sort.Slice(items, func(i, j int) bool {
		profitI := currentPrice - items[i].PurchasePrice
		profitJ := currentPrice - items[j].PurchasePrice
		return profitI > profitJ
	})
	return items
}

func (s *ProfitMaximizationStrategy) GetName() string {
	return "ProfitMaximization"
}

// VelocityBasedStrategy prioritizes fast-moving items
type VelocityBasedStrategy struct {
	velocities map[string]float64
}

func (s *VelocityBasedStrategy) DetermineSellPriority(items []*InventoryItem, currentPrice int) []*InventoryItem {
	// Sort by velocity (highest first)
	sort.Slice(items, func(i, j int) bool {
		velI := 0.0
		velJ := 0.0
		if s.velocities != nil {
			velI = s.velocities[items[i].Item.ID]
			velJ = s.velocities[items[j].Item.ID]
		}
		return velI > velJ
	})
	return items
}

func (s *VelocityBasedStrategy) GetName() string {
	return "VelocityBased"
}
