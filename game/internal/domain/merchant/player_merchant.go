package merchant

import (
	"errors"
	"sync"

	"github.com/yourusername/merchant-tails/game/internal/domain/item"
)

// PlayerMerchant represents the player's merchant character
type PlayerMerchant struct {
	ID         string
	Name       string
	Gold       int
	Reputation float64
	Level      int
	Experience int
	Inventory  *item.Inventory
	Stats      PlayerStats
	mu         sync.RWMutex
}

// PlayerStats tracks player performance
type PlayerStats struct {
	TotalProfit     int
	SuccessfulDeals int
	FailedDeals     int
	TotalVolume     int
	ItemsSold       int
	ItemsBought     int
}

// NewPlayerMerchant creates a new player merchant
func NewPlayerMerchant(id, name string, startingGold int) (*PlayerMerchant, error) {
	if id == "" {
		return nil, errors.New("merchant id cannot be empty")
	}
	if name == "" {
		return nil, errors.New("merchant name cannot be empty")
	}
	if startingGold < 0 {
		return nil, errors.New("starting gold must be positive")
	}

	return &PlayerMerchant{
		ID:         id,
		Name:       name,
		Gold:       startingGold,
		Reputation: 50.0, // Start with neutral reputation
		Level:      1,
		Experience: 0,
		Inventory:  item.NewInventory(),
		Stats:      PlayerStats{},
	}, nil
}

// BuyItem purchases an item
func (pm *PlayerMerchant) BuyItem(item *item.Item, quantity, pricePerUnit int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	totalCost := quantity * pricePerUnit
	if pm.Gold < totalCost {
		return errors.New("insufficient gold")
	}

	// For now, we'll skip inventory capacity check
	// TODO: Add proper inventory capacity management

	// Execute purchase
	pm.Gold -= totalCost
	err := pm.Inventory.AddItem(item, quantity)
	if err != nil {
		pm.Gold += totalCost // Rollback
		return err
	}

	// Update stats
	pm.Stats.ItemsBought += quantity
	pm.Stats.TotalVolume += totalCost
	pm.Stats.SuccessfulDeals++

	// Add experience
	pm.addExperience(10 * quantity)

	return nil
}

// SellItem sells an item from inventory
func (pm *PlayerMerchant) SellItem(itemID string, quantity, pricePerUnit int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check inventory
	if pm.Inventory.GetQuantity(itemID) < quantity {
		return errors.New("insufficient items in inventory")
	}

	// Execute sale
	err := pm.Inventory.RemoveItem(itemID, quantity)
	if err != nil {
		return err
	}

	totalRevenue := quantity * pricePerUnit
	pm.Gold += totalRevenue

	// Update stats
	pm.Stats.ItemsSold += quantity
	pm.Stats.TotalVolume += totalRevenue
	pm.Stats.TotalProfit += totalRevenue // Simplified profit calculation
	pm.Stats.SuccessfulDeals++

	// Add experience
	pm.addExperience(15 * quantity)

	// Update reputation based on successful sale
	pm.updateReputation(1.0)

	return nil
}

// addExperience adds experience and handles level up
func (pm *PlayerMerchant) addExperience(exp int) {
	pm.Experience += exp

	// Simple level up system
	requiredExp := pm.Level * 1000
	if pm.Experience >= requiredExp {
		pm.Level++
		pm.Experience -= requiredExp
		// Could trigger level up bonuses here
	}
}

// updateReputation updates the player's reputation
func (pm *PlayerMerchant) updateReputation(change float64) {
	pm.Reputation += change

	// Cap reputation between 0 and 100
	if pm.Reputation > 100 {
		pm.Reputation = 100
	} else if pm.Reputation < 0 {
		pm.Reputation = 0
	}
}

// GetGold returns current gold amount
func (pm *PlayerMerchant) GetGold() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.Gold
}

// GetStats returns player statistics
func (pm *PlayerMerchant) GetStats() PlayerStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.Stats
}

// GetLevel returns current level
func (pm *PlayerMerchant) GetLevel() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.Level
}

// GetReputation returns current reputation
func (pm *PlayerMerchant) GetReputation() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.Reputation
}

// CanAfford checks if player can afford a purchase
func (pm *PlayerMerchant) CanAfford(cost int) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.Gold >= cost
}

// GetInventoryValue calculates total inventory value at given prices
func (pm *PlayerMerchant) GetInventoryValue(prices map[string]int) int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	totalValue := 0
	items := pm.Inventory.GetItems()

	for _, inventoryItem := range items {
		if price, exists := prices[inventoryItem.Item.ID]; exists {
			totalValue += price * inventoryItem.Quantity
		}
	}

	return totalValue
}

// GetNetWorth calculates total net worth
func (pm *PlayerMerchant) GetNetWorth(prices map[string]int) int {
	return pm.GetGold() + pm.GetInventoryValue(prices)
}

// Reset resets the merchant to initial state (for new game)
func (pm *PlayerMerchant) Reset(startingGold int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.Gold = startingGold
	pm.Reputation = 50.0
	pm.Level = 1
	pm.Experience = 0
	pm.Inventory = item.NewInventory()
	pm.Stats = PlayerStats{}
}
