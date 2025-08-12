package item

import (
	"errors"
	"sync"
	"time"
)

// ErrItemNotFound is returned when an item is not found in the registry
var ErrItemNotFound = errors.New("item not found in registry")

// ItemRegistry manages all available items in the game
type ItemRegistry struct {
	items map[string]*ItemMaster
	mu    sync.RWMutex
}

var (
	globalRegistry *ItemRegistry
	registryOnce   sync.Once
)

// GetItemRegistry returns the global item registry
func GetItemRegistry() *ItemRegistry {
	registryOnce.Do(func() {
		globalRegistry = &ItemRegistry{
			items: make(map[string]*ItemMaster),
		}
		globalRegistry.initializeDefaultItems()
	})
	return globalRegistry
}

// initializeDefaultItems populates the registry with default game items
func (r *ItemRegistry) initializeDefaultItems() {
	// Fruits
	r.RegisterItem(&ItemMaster{
		ID:         "apple",
		Name:       "Fresh Apple",
		Category:   CategoryFruit,
		BasePrice:  10,
		Durability: 3,
		Volatility: 0.2,
		SeasonalModifiers: map[Season]float32{
			SeasonSpring: 1.1,
			SeasonSummer: 1.0,
			SeasonAutumn: 1.3,
			SeasonWinter: 0.8,
		},
	})

	r.RegisterItem(&ItemMaster{
		ID:         "orange",
		Name:       "Juicy Orange",
		Category:   CategoryFruit,
		BasePrice:  12,
		Durability: 4,
		Volatility: 0.15,
		SeasonalModifiers: map[Season]float32{
			SeasonSpring: 0.9,
			SeasonSummer: 1.2,
			SeasonAutumn: 1.0,
			SeasonWinter: 1.1,
		},
	})

	r.RegisterItem(&ItemMaster{
		ID:         "grapes",
		Name:       "Sweet Grapes",
		Category:   CategoryFruit,
		BasePrice:  15,
		Durability: 2,
		Volatility: 0.25,
		SeasonalModifiers: map[Season]float32{
			SeasonSpring: 0.8,
			SeasonSummer: 1.3,
			SeasonAutumn: 1.2,
			SeasonWinter: 0.7,
		},
	})

	// Potions
	r.RegisterItem(&ItemMaster{
		ID:         "health_potion",
		Name:       "Health Potion",
		Category:   CategoryPotion,
		BasePrice:  50,
		Durability: 30,
		Volatility: 0.3,
		SeasonalModifiers: map[Season]float32{
			SeasonSpring: 1.0,
			SeasonSummer: 0.9,
			SeasonAutumn: 1.1,
			SeasonWinter: 1.2,
		},
	})

	r.RegisterItem(&ItemMaster{
		ID:         "mana_potion",
		Name:       "Mana Potion",
		Category:   CategoryPotion,
		BasePrice:  60,
		Durability: 30,
		Volatility: 0.35,
		SeasonalModifiers: map[Season]float32{
			SeasonSpring: 1.1,
			SeasonSummer: 1.0,
			SeasonAutumn: 1.0,
			SeasonWinter: 1.1,
		},
	})

	r.RegisterItem(&ItemMaster{
		ID:         "stamina_potion",
		Name:       "Stamina Potion",
		Category:   CategoryPotion,
		BasePrice:  40,
		Durability: 30,
		Volatility: 0.25,
		SeasonalModifiers: map[Season]float32{
			SeasonSpring: 1.2,
			SeasonSummer: 1.3,
			SeasonAutumn: 0.9,
			SeasonWinter: 0.8,
		},
	})

	// Weapons
	r.RegisterItem(&ItemMaster{
		ID:         "iron_sword",
		Name:       "Iron Sword",
		Category:   CategoryWeapon,
		BasePrice:  150,
		Durability: -1, // Never spoils
		Volatility: 0.1,
		SeasonalModifiers: map[Season]float32{
			SeasonSpring: 1.0,
			SeasonSummer: 1.0,
			SeasonAutumn: 1.0,
			SeasonWinter: 1.0,
		},
	})

	r.RegisterItem(&ItemMaster{
		ID:         "steel_sword",
		Name:       "Steel Sword",
		Category:   CategoryWeapon,
		BasePrice:  300,
		Durability: -1,
		Volatility: 0.08,
		SeasonalModifiers: map[Season]float32{
			SeasonSpring: 1.0,
			SeasonSummer: 1.0,
			SeasonAutumn: 1.0,
			SeasonWinter: 1.0,
		},
	})

	r.RegisterItem(&ItemMaster{
		ID:         "magic_staff",
		Name:       "Magic Staff",
		Category:   CategoryWeapon,
		BasePrice:  500,
		Durability: -1,
		Volatility: 0.15,
		SeasonalModifiers: map[Season]float32{
			SeasonSpring: 1.1,
			SeasonSummer: 1.0,
			SeasonAutumn: 1.0,
			SeasonWinter: 0.9,
		},
	})
}

// RegisterItem adds an item to the registry
func (r *ItemRegistry) RegisterItem(master *ItemMaster) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[master.ID] = master
}

// GetItem retrieves an item from the registry
func (r *ItemRegistry) GetItem(id string) (*ItemMaster, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, exists := r.items[id]
	return item, exists
}

// GetAllItems returns all items in the registry
func (r *ItemRegistry) GetAllItems() []*ItemMaster {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*ItemMaster, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}
	return items
}

// GetItemsByCategory returns all items of a specific category
func (r *ItemRegistry) GetItemsByCategory(category Category) []*ItemMaster {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []*ItemMaster
	for _, item := range r.items {
		if item.Category == category {
			items = append(items, item)
		}
	}
	return items
}

// CreateItem creates a new item instance from master data
func (r *ItemRegistry) CreateItem(id string) (*Item, error) {
	master, exists := r.GetItem(id)
	if !exists {
		return nil, ErrItemNotFound
	}

	return &Item{
		ID:         master.ID,
		Name:       master.Name,
		Category:   master.Category,
		BasePrice:  master.BasePrice,
		Price:      master.BasePrice,
		Durability: master.Durability,
		CreatedAt:  time.Now(),
	}, nil
}
