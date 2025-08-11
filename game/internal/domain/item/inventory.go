package item

import (
	"errors"
	"sync"
)

// Inventory represents a collection of items
type Inventory struct {
	items map[string]int // itemID -> quantity
	mu    sync.RWMutex
}

// NewInventory creates a new inventory
func NewInventory() *Inventory {
	return &Inventory{
		items: make(map[string]int),
	}
}

// AddItem adds an item to the inventory
func (inv *Inventory) AddItem(item *Item, quantity int) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	inv.mu.Lock()
	defer inv.mu.Unlock()

	inv.items[item.ID] += quantity
	return nil
}

// RemoveItem removes an item from the inventory
func (inv *Inventory) RemoveItem(itemID string, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	inv.mu.Lock()
	defer inv.mu.Unlock()

	currentQty, exists := inv.items[itemID]
	if !exists {
		return errors.New("item not found in inventory")
	}

	if currentQty < quantity {
		return errors.New("insufficient quantity in inventory")
	}

	inv.items[itemID] -= quantity
	if inv.items[itemID] == 0 {
		delete(inv.items, itemID)
	}

	return nil
}

// GetQuantity returns the quantity of an item
func (inv *Inventory) GetQuantity(itemID string) int {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	return inv.items[itemID]
}

// GetAll returns all items in the inventory
func (inv *Inventory) GetAll() map[string]int {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]int)
	for id, qty := range inv.items {
		result[id] = qty
	}
	return result
}

// Clear removes all items from the inventory
func (inv *Inventory) Clear() {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	inv.items = make(map[string]int)
}

// IsEmpty returns true if the inventory is empty
func (inv *Inventory) IsEmpty() bool {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	return len(inv.items) == 0
}

// GetTotalItems returns the total number of items
func (inv *Inventory) GetTotalItems() int {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	total := 0
	for _, qty := range inv.items {
		total += qty
	}
	return total
}

// Contains checks if an item exists in the inventory
func (inv *Inventory) Contains(itemID string) bool {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	_, exists := inv.items[itemID]
	return exists
}

// GetItems returns all items as a map
func (inv *Inventory) GetItems() map[string]int {
	return inv.GetAll()
}

// HasItem checks if an item exists with sufficient quantity
func (inv *Inventory) HasItem(itemID string, quantity int) bool {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	currentQty, exists := inv.items[itemID]
	return exists && currentQty >= quantity
}
