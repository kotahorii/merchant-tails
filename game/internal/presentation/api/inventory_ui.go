package api

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/inventory"
	"github.com/yourusername/merchant-tails/game/internal/domain/item"
)

// Fruit item constants
const (
	itemApple  = "apple"
	itemOrange = "orange"
	itemBanana = "banana"

	// Location constants
	locationShop      = "shop"
	locationWarehouse = "warehouse"

	// Category filter constant
	categoryAll = "all"
)

// InventoryUIItem represents an item in the inventory UI
type InventoryUIItem struct {
	ItemID        string        `json:"item_id"`
	Name          string        `json:"name"`
	Category      item.Category `json:"category"`
	Quantity      int           `json:"quantity"`
	Location      string        `json:"location"` // "shop" or "warehouse"
	PurchasePrice float64       `json:"purchase_price"`
	CurrentPrice  float64       `json:"current_price"`
	ProfitMargin  float64       `json:"profit_margin"`
	DaysInStock   int           `json:"days_in_stock"`
	Durability    int           `json:"durability"` // -1 for non-perishable
	SalesVelocity float64       `json:"sales_velocity"`
	SpaceUsed     int           `json:"space_used"`
	Icon          string        `json:"icon"`
}

// InventoryTransferRequest represents a request to transfer items
type InventoryTransferRequest struct {
	ItemID       string `json:"item_id"`
	Quantity     int    `json:"quantity"`
	FromLocation string `json:"from_location"` // "shop" or "warehouse"
	ToLocation   string `json:"to_location"`   // "shop" or "warehouse"
}

// InventoryTransferResult represents the result of a transfer
type InventoryTransferResult struct {
	Success      bool   `json:"success"`
	ItemID       string `json:"item_id"`
	Quantity     int    `json:"quantity"`
	FromLocation string `json:"from_location"`
	ToLocation   string `json:"to_location"`
	Message      string `json:"message"`
}

// BulkTransferRequest represents multiple transfer requests
type BulkTransferRequest struct {
	Transfers     []InventoryTransferRequest `json:"transfers"`
	OptimizeSpace bool                       `json:"optimize_space"`
}

// InventoryStats represents inventory statistics
type InventoryStats struct {
	ShopCapacity         int     `json:"shop_capacity"`
	ShopUsed             int     `json:"shop_used"`
	ShopUtilization      float64 `json:"shop_utilization"`
	WarehouseCapacity    int     `json:"warehouse_capacity"`
	WarehouseUsed        int     `json:"warehouse_used"`
	WarehouseUtilization float64 `json:"warehouse_utilization"`
	TotalItems           int     `json:"total_items"`
	TotalValue           float64 `json:"total_value"`
	PerishableItems      int     `json:"perishable_items"`
	ExpiringItems        int     `json:"expiring_items"` // Items expiring in next 3 days
}

// OptimizationSuggestion represents a suggested inventory optimization
type OptimizationSuggestion struct {
	Type          string  `json:"type"` // "move_to_shop", "move_to_warehouse", "sell_soon", "restock"
	ItemID        string  `json:"item_id"`
	ItemName      string  `json:"item_name"`
	Quantity      int     `json:"quantity"`
	Reason        string  `json:"reason"`
	Priority      int     `json:"priority"` // 1-5, higher is more urgent
	PotentialGain float64 `json:"potential_gain"`
}

// InventoryFilter represents filtering options for inventory display
type InventoryFilter struct {
	Category    string `json:"category"`
	Location    string `json:"location"`
	MinQuantity int    `json:"min_quantity"`
	MaxQuantity int    `json:"max_quantity"`
	Perishable  *bool  `json:"perishable,omitempty"`
	SortBy      string `json:"sort_by"`    // "name", "quantity", "value", "velocity", "age"
	SortOrder   string `json:"sort_order"` // "asc" or "desc"
}

// InventoryUIManager manages the inventory UI backend
type InventoryUIManager struct {
	gameManager   *GameManager
	inventory     *inventory.InventoryManager
	salesHistory  map[string]*SalesData
	optimizations []OptimizationSuggestion
	mu            sync.RWMutex
}

// SalesData tracks sales information for an item
type SalesData struct {
	ItemID       string
	TotalSold    int
	LastSaleTime time.Time
	AveragePrice float64
	DailySales   []int // Last 7 days
}

// NewInventoryUIManager creates a new inventory UI manager
func NewInventoryUIManager(gameManager *GameManager) *InventoryUIManager {
	return &InventoryUIManager{
		gameManager:   gameManager,
		inventory:     gameManager.inventory,
		salesHistory:  make(map[string]*SalesData),
		optimizations: make([]OptimizationSuggestion, 0),
	}
}

// GetInventoryItems returns all inventory items with UI data
func (iui *InventoryUIManager) GetInventoryItems(filter *InventoryFilter) ([]*InventoryUIItem, error) {
	iui.mu.RLock()
	defer iui.mu.RUnlock()

	items := make([]*InventoryUIItem, 0)

	// Get shop items
	shopItems := iui.getShopItems()
	for _, item := range shopItems {
		if iui.matchesFilter(item, filter) {
			items = append(items, item)
		}
	}

	// Get warehouse items
	warehouseItems := iui.getWarehouseItems()
	for _, item := range warehouseItems {
		if iui.matchesFilter(item, filter) {
			items = append(items, item)
		}
	}

	// Sort items
	if filter != nil && filter.SortBy != "" {
		iui.sortItems(items, filter.SortBy, filter.SortOrder)
	}

	return items, nil
}

// TransferItem transfers items between shop and warehouse
func (iui *InventoryUIManager) TransferItem(request *InventoryTransferRequest) (*InventoryTransferResult, error) {
	iui.mu.Lock()
	defer iui.mu.Unlock()

	// Validate request
	if request.Quantity <= 0 {
		return &InventoryTransferResult{
			Success: false,
			Message: "Invalid quantity",
		}, nil
	}

	if request.FromLocation == request.ToLocation {
		return &InventoryTransferResult{
			Success: false,
			Message: "Source and destination cannot be the same",
		}, nil
	}

	// Execute transfer based on direction
	var err error
	switch {
	case request.FromLocation == locationShop && request.ToLocation == locationWarehouse:
		err = iui.inventory.TransferToWarehouse(request.ItemID, request.Quantity)
	case request.FromLocation == locationWarehouse && request.ToLocation == locationShop:
		err = iui.inventory.TransferToShop(request.ItemID, request.Quantity)
	default:
		return &InventoryTransferResult{
			Success: false,
			Message: "Invalid transfer locations",
		}, nil
	}

	if err != nil {
		return &InventoryTransferResult{
			Success:      false,
			ItemID:       request.ItemID,
			Quantity:     request.Quantity,
			FromLocation: request.FromLocation,
			ToLocation:   request.ToLocation,
			Message:      err.Error(),
		}, nil
	}

	return &InventoryTransferResult{
		Success:      true,
		ItemID:       request.ItemID,
		Quantity:     request.Quantity,
		FromLocation: request.FromLocation,
		ToLocation:   request.ToLocation,
		Message:      "Transfer successful",
	}, nil
}

// BulkTransfer executes multiple transfers
func (iui *InventoryUIManager) BulkTransfer(request *BulkTransferRequest) ([]*InventoryTransferResult, error) {
	results := make([]*InventoryTransferResult, 0, len(request.Transfers))

	// Optimize transfer order if requested
	transfers := request.Transfers
	if request.OptimizeSpace {
		transfers = iui.optimizeTransferOrder(transfers)
	}

	for _, transfer := range transfers {
		result, err := iui.TransferItem(&transfer)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}

	return results, nil
}

// GetInventoryStats returns current inventory statistics
func (iui *InventoryUIManager) GetInventoryStats() (*InventoryStats, error) {
	iui.mu.RLock()
	defer iui.mu.RUnlock()

	// Calculate shop usage
	shopUsed := 0
	shopItems := iui.inventory.ShopInventory.GetAll()
	for _, quantity := range shopItems {
		shopUsed += quantity
	}

	// Calculate warehouse usage
	warehouseUsed := 0
	warehouseItems := iui.inventory.WarehouseInventory.GetAll()
	for _, quantity := range warehouseItems {
		warehouseUsed += quantity
	}

	// Calculate total value
	totalValue := 0.0
	totalItems := 0
	perishableItems := 0
	expiringItems := 0

	// Count items and calculate values
	for itemID, quantity := range shopItems {
		totalItems += quantity
		price := float64(iui.gameManager.market.GetPrice(itemID))
		totalValue += price * float64(quantity)

		// Check if perishable
		if iui.isPerishable(itemID) {
			perishableItems++
			if iui.isExpiringSoon(itemID, 3) {
				expiringItems++
			}
		}
	}

	for itemID, quantity := range warehouseItems {
		totalItems += quantity
		price := float64(iui.gameManager.market.GetPrice(itemID))
		totalValue += price * float64(quantity)

		if iui.isPerishable(itemID) {
			perishableItems++
			if iui.isExpiringSoon(itemID, 3) {
				expiringItems++
			}
		}
	}

	stats := &InventoryStats{
		ShopCapacity:         iui.inventory.ShopCapacity,
		ShopUsed:             shopUsed,
		ShopUtilization:      float64(shopUsed) / float64(iui.inventory.ShopCapacity),
		WarehouseCapacity:    iui.inventory.WarehouseCapacity,
		WarehouseUsed:        warehouseUsed,
		WarehouseUtilization: float64(warehouseUsed) / float64(iui.inventory.WarehouseCapacity),
		TotalItems:           totalItems,
		TotalValue:           totalValue,
		PerishableItems:      perishableItems,
		ExpiringItems:        expiringItems,
	}

	return stats, nil
}

// GetOptimizationSuggestions returns inventory optimization suggestions
func (iui *InventoryUIManager) GetOptimizationSuggestions() ([]*OptimizationSuggestion, error) {
	iui.mu.Lock()
	defer iui.mu.Unlock()

	suggestions := make([]*OptimizationSuggestion, 0)

	// Analyze shop items
	shopItems := iui.inventory.ShopInventory.GetAll()
	for itemID, quantity := range shopItems {
		// Check for slow-moving items
		velocity := iui.getSalesVelocity(itemID)
		if velocity < 0.1 && quantity > 5 {
			suggestions = append(suggestions, &OptimizationSuggestion{
				Type:     "move_to_warehouse",
				ItemID:   itemID,
				ItemName: itemID, // Would be replaced with actual name
				Quantity: quantity / 2,
				Reason:   fmt.Sprintf("Low sales velocity (%.2f items/day)", velocity),
				Priority: 2,
			})
		}

		// Check for expiring items
		if iui.isExpiringSoon(itemID, 2) {
			suggestions = append(suggestions, &OptimizationSuggestion{
				Type:     "sell_soon",
				ItemID:   itemID,
				ItemName: itemID,
				Quantity: quantity,
				Reason:   "Item expiring in 2 days",
				Priority: 5,
			})
		}
	}

	// Analyze warehouse items
	warehouseItems := iui.inventory.WarehouseInventory.GetAll()
	for itemID, quantity := range warehouseItems {
		// Check for high-demand items
		velocity := iui.getSalesVelocity(itemID)
		if velocity > 1.0 && quantity > 0 {
			// Check if shop has room
			shopSpace := iui.inventory.ShopCapacity - iui.getShopUsage()
			if shopSpace > 0 {
				moveQty := min(quantity, shopSpace)
				suggestions = append(suggestions, &OptimizationSuggestion{
					Type:     "move_to_shop",
					ItemID:   itemID,
					ItemName: itemID,
					Quantity: moveQty,
					Reason:   fmt.Sprintf("High sales velocity (%.2f items/day)", velocity),
					Priority: 3,
				})
			}
		}
	}

	// Check for restock opportunities
	for itemID, salesData := range iui.salesHistory {
		currentStock := iui.getTotalStock(itemID)
		avgDailySales := iui.getAverageDailySales(salesData)

		if float64(currentStock) < avgDailySales*3 { // Less than 3 days of stock
			suggestions = append(suggestions, &OptimizationSuggestion{
				Type:     "restock",
				ItemID:   itemID,
				ItemName: itemID,
				Quantity: int(avgDailySales * 7), // Suggest 1 week of stock
				Reason:   fmt.Sprintf("Low stock (%.1f days remaining)", float64(currentStock)/avgDailySales),
				Priority: 4,
			})
		}
	}

	// Sort by priority
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Priority > suggestions[j].Priority
	})

	// Convert to non-pointer slice for storage
	nonPointerSlice := make([]OptimizationSuggestion, 0, len(suggestions))
	for _, s := range suggestions {
		if s != nil {
			nonPointerSlice = append(nonPointerSlice, *s)
		}
	}
	iui.optimizations = nonPointerSlice
	return suggestions, nil
}

// OptimizeLayout automatically optimizes inventory layout
func (iui *InventoryUIManager) OptimizeLayout() error {
	suggestions, err := iui.GetOptimizationSuggestions()
	if err != nil {
		return err
	}

	// Execute high-priority suggestions automatically
	for _, suggestion := range suggestions {
		if suggestion.Priority >= 4 {
			switch suggestion.Type {
			case "move_to_shop":
				_, _ = iui.TransferItem(&InventoryTransferRequest{
					ItemID:       suggestion.ItemID,
					Quantity:     suggestion.Quantity,
					FromLocation: "warehouse",
					ToLocation:   "shop",
				})
			case "move_to_warehouse":
				_, _ = iui.TransferItem(&InventoryTransferRequest{
					ItemID:       suggestion.ItemID,
					Quantity:     suggestion.Quantity,
					FromLocation: "shop",
					ToLocation:   "warehouse",
				})
			}
		}
	}

	return nil
}

// Helper functions

func (iui *InventoryUIManager) getShopItems() []*InventoryUIItem {
	items := make([]*InventoryUIItem, 0)
	shopItems := iui.inventory.ShopInventory.GetAll()

	for itemID, quantity := range shopItems {
		currentPrice := float64(iui.gameManager.market.GetPrice(itemID))
		purchasePrice := iui.getPurchasePrice(itemID)

		item := &InventoryUIItem{
			ItemID:        itemID,
			Name:          iui.getItemName(itemID),
			Category:      iui.getItemCategory(itemID),
			Quantity:      quantity,
			Location:      "shop",
			PurchasePrice: purchasePrice,
			CurrentPrice:  currentPrice,
			ProfitMargin:  ((currentPrice - purchasePrice) / purchasePrice) * 100,
			DaysInStock:   iui.getDaysInStock(itemID),
			Durability:    iui.getItemDurability(itemID),
			SalesVelocity: iui.getSalesVelocity(itemID),
			SpaceUsed:     quantity,
			Icon:          fmt.Sprintf("res://assets/items/%s.png", itemID),
		}
		items = append(items, item)
	}

	return items
}

func (iui *InventoryUIManager) getWarehouseItems() []*InventoryUIItem {
	items := make([]*InventoryUIItem, 0)
	warehouseItems := iui.inventory.WarehouseInventory.GetAll()

	for itemID, quantity := range warehouseItems {
		currentPrice := float64(iui.gameManager.market.GetPrice(itemID))
		purchasePrice := iui.getPurchasePrice(itemID)

		item := &InventoryUIItem{
			ItemID:        itemID,
			Name:          iui.getItemName(itemID),
			Category:      iui.getItemCategory(itemID),
			Quantity:      quantity,
			Location:      "warehouse",
			PurchasePrice: purchasePrice,
			CurrentPrice:  currentPrice,
			ProfitMargin:  ((currentPrice - purchasePrice) / purchasePrice) * 100,
			DaysInStock:   iui.getDaysInStock(itemID),
			Durability:    iui.getItemDurability(itemID),
			SalesVelocity: iui.getSalesVelocity(itemID),
			SpaceUsed:     quantity,
			Icon:          fmt.Sprintf("res://assets/items/%s.png", itemID),
		}
		items = append(items, item)
	}

	return items
}

func (iui *InventoryUIManager) matchesFilter(item *InventoryUIItem, filter *InventoryFilter) bool {
	if filter == nil {
		return true
	}

	if filter.Category != "" && filter.Category != categoryAll && string(item.Category) != filter.Category {
		return false
	}

	if filter.Location != "" && filter.Location != categoryAll && item.Location != filter.Location {
		return false
	}

	if filter.MinQuantity > 0 && item.Quantity < filter.MinQuantity {
		return false
	}

	if filter.MaxQuantity > 0 && item.Quantity > filter.MaxQuantity {
		return false
	}

	if filter.Perishable != nil {
		isPerishable := item.Durability > 0
		if *filter.Perishable != isPerishable {
			return false
		}
	}

	return true
}

func (iui *InventoryUIManager) sortItems(items []*InventoryUIItem, sortBy, sortOrder string) {
	sort.Slice(items, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "name":
			less = items[i].Name < items[j].Name
		case "quantity":
			less = items[i].Quantity < items[j].Quantity
		case "value":
			less = items[i].CurrentPrice*float64(items[i].Quantity) < items[j].CurrentPrice*float64(items[j].Quantity)
		case "velocity":
			less = items[i].SalesVelocity < items[j].SalesVelocity
		case "age":
			less = items[i].DaysInStock < items[j].DaysInStock
		default:
			less = items[i].Name < items[j].Name
		}

		if sortOrder == "desc" {
			return !less
		}
		return less
	})
}

func (iui *InventoryUIManager) optimizeTransferOrder(transfers []InventoryTransferRequest) []InventoryTransferRequest {
	// Simple optimization: execute warehouse->shop transfers first to free up space
	optimized := make([]InventoryTransferRequest, 0, len(transfers))

	// First, add all warehouse to shop transfers
	for _, transfer := range transfers {
		if transfer.FromLocation == "warehouse" && transfer.ToLocation == "shop" {
			optimized = append(optimized, transfer)
		}
	}

	// Then add all shop to warehouse transfers
	for _, transfer := range transfers {
		if transfer.FromLocation == "shop" && transfer.ToLocation == "warehouse" {
			optimized = append(optimized, transfer)
		}
	}

	return optimized
}

func (iui *InventoryUIManager) getSalesVelocity(itemID string) float64 {
	if data, exists := iui.salesHistory[itemID]; exists {
		return iui.getAverageDailySales(data)
	}
	return 0.0
}

func (iui *InventoryUIManager) getAverageDailySales(data *SalesData) float64 {
	if len(data.DailySales) == 0 {
		return 0.0
	}

	total := 0
	for _, sales := range data.DailySales {
		total += sales
	}
	return float64(total) / float64(len(data.DailySales))
}

func (iui *InventoryUIManager) isPerishable(itemID string) bool {
	// Check if item is in fruit category or has durability
	return itemID == itemApple || itemID == itemOrange || itemID == itemBanana
}

func (iui *InventoryUIManager) isExpiringSoon(itemID string, days int) bool {
	if !iui.isPerishable(itemID) {
		return false
	}

	daysInStock := iui.getDaysInStock(itemID)
	durability := iui.getItemDurability(itemID)

	if durability > 0 {
		return durability-daysInStock <= days
	}
	return false
}

func (iui *InventoryUIManager) getShopUsage() int {
	total := 0
	for _, quantity := range iui.inventory.ShopInventory.GetAll() {
		total += quantity
	}
	return total
}

func (iui *InventoryUIManager) getTotalStock(itemID string) int {
	shopQty := iui.inventory.ShopInventory.GetQuantity(itemID)
	warehouseQty := iui.inventory.WarehouseInventory.GetQuantity(itemID)
	return shopQty + warehouseQty
}

func (iui *InventoryUIManager) getPurchasePrice(itemID string) float64 {
	// Simplified - in production would track actual purchase prices
	return float64(iui.gameManager.market.GetPrice(itemID)) * 0.8
}

func (iui *InventoryUIManager) getItemName(itemID string) string {
	// Map item IDs to display names
	names := map[string]string{
		"apple":               "Apple",
		"sword_iron":          "Iron Sword",
		"potion_health":       "Health Potion",
		"gem_diamond":         "Diamond",
		"spellbook_rare":      "Rare Spellbook",
		"accessory_gold_ring": "Gold Ring",
	}

	if name, exists := names[itemID]; exists {
		return name
	}
	return itemID
}

func (iui *InventoryUIManager) getItemCategory(itemID string) item.Category {
	// Map item IDs to categories
	categories := map[string]item.Category{
		"apple":               item.CategoryFruit,
		"sword_iron":          item.CategoryWeapon,
		"potion_health":       item.CategoryPotion,
		"gem_diamond":         item.CategoryGem,
		"spellbook_rare":      item.CategoryMagicBook,
		"accessory_gold_ring": item.CategoryAccessory,
	}

	if cat, exists := categories[itemID]; exists {
		return cat
	}
	return item.CategoryFruit
}

func (iui *InventoryUIManager) getDaysInStock(_ string) int {
	// Simplified - return random value for demo
	return 3
}

func (iui *InventoryUIManager) getItemDurability(itemID string) int {
	if iui.isPerishable(itemID) {
		return 7 // 7 days for fruits
	}
	return -1 // Non-perishable
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
