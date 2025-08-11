package item

import (
	"errors"
	"sync"
	"time"
)

// Category represents the type of item
type Category string

const (
	CategoryFruit     Category = "FRUIT"
	CategoryPotion    Category = "POTION"
	CategoryWeapon    Category = "WEAPON"
	CategoryAccessory Category = "ACCESSORY"
	CategoryMagicBook Category = "MAGIC_BOOK"
	CategoryGem       Category = "GEM"
)

// Season represents the game season
type Season string

const (
	SeasonSpring Season = "SPRING"
	SeasonSummer Season = "SUMMER"
	SeasonAutumn Season = "AUTUMN"
	SeasonWinter Season = "WINTER"
)

// PriceTrend represents price movement direction
type PriceTrend int

const (
	PriceTrendStable PriceTrend = iota
	PriceTrendUp
	PriceTrendDown
)

// Item represents a tradeable item in the game
type Item struct {
	ID         string
	Name       string
	Category   Category
	BasePrice  int
	Price      int // Current price (can differ from BasePrice)
	Durability int // Days until spoilage, -1 for infinite
	CreatedAt  time.Time
}

// NewItem creates a new item with validation
func NewItem(id, name string, category Category, basePrice int) (*Item, error) {
	if id == "" {
		return nil, errors.New("item id cannot be empty")
	}
	if name == "" {
		return nil, errors.New("item name cannot be empty")
	}
	if basePrice <= 0 {
		return nil, errors.New("base price must be positive")
	}

	durability := getDurabilityByCategory(category)

	return &Item{
		ID:         id,
		Name:       name,
		Category:   category,
		BasePrice:  basePrice,
		Price:      basePrice, // Initialize current price to base price
		Durability: durability,
		CreatedAt:  time.Now(),
	}, nil
}

// CalculatePrice calculates the current price based on modifiers
func (i *Item) CalculatePrice(demandModifier, seasonModifier float64) int {
	price := float64(i.BasePrice) * demandModifier * seasonModifier
	return int(price)
}

// UpdateDurability decreases durability by one day
func (i *Item) UpdateDurability() {
	if i.Durability > 0 {
		i.Durability--
	}
}

// IsSpoiled checks if the item has spoiled
func (i *Item) IsSpoiled() bool {
	return i.Durability == 0
}

// GetVolatility returns the price volatility for the item category
func (i *Item) GetVolatility() float32 {
	volatilityMap := map[Category]float32{
		CategoryFruit:     0.1,
		CategoryPotion:    0.3,
		CategoryWeapon:    0.05,
		CategoryAccessory: 0.5,
		CategoryMagicBook: 0.1,
		CategoryGem:       0.7,
	}

	if v, ok := volatilityMap[i.Category]; ok {
		return v
	}
	return 0.1 // default volatility
}

// getDurabilityByCategory returns the default durability for each category
func getDurabilityByCategory(category Category) int {
	switch category {
	case CategoryFruit:
		return 3 // Spoils in 3 days
	case CategoryPotion:
		return 30 // 30 days shelf life
	case CategoryWeapon, CategoryMagicBook, CategoryGem:
		return -1 // Never spoils
	case CategoryAccessory:
		return -1 // Never spoils but fashion changes
	default:
		return -1
	}
}

// ItemMaster represents master data for an item type
type ItemMaster struct {
	ID                string
	Name              string
	Category          Category
	BasePrice         int
	Durability        int
	Volatility        float32
	SeasonalModifiers map[Season]float32
}

// GetSeasonalModifier returns the price modifier for a given season
func (im *ItemMaster) GetSeasonalModifier(season Season) float32 {
	if modifier, ok := im.SeasonalModifiers[season]; ok {
		return modifier
	}
	return 1.0 // No modifier
}

// PriceRecord represents a historical price point
type PriceRecord struct {
	Price     int
	Timestamp time.Time
}

// PriceHistory tracks historical prices for an item
type PriceHistory struct {
	Records []PriceRecord
	MaxSize int
	mu      sync.RWMutex
}

// NewPriceHistory creates a new price history tracker
func NewPriceHistory(maxSize int) *PriceHistory {
	return &PriceHistory{
		Records: make([]PriceRecord, 0, maxSize),
		MaxSize: maxSize,
	}
}

// AddRecord adds a new price record
func (ph *PriceHistory) AddRecord(price int, timestamp time.Time) {
	ph.mu.Lock()
	defer ph.mu.Unlock()

	ph.Records = append(ph.Records, PriceRecord{
		Price:     price,
		Timestamp: timestamp,
	})

	// Maintain max size by removing oldest records
	if len(ph.Records) > ph.MaxSize {
		ph.Records = ph.Records[len(ph.Records)-ph.MaxSize:]
	}
}

// GetLatestPrice returns the most recent price
func (ph *PriceHistory) GetLatestPrice() int {
	ph.mu.RLock()
	defer ph.mu.RUnlock()

	if len(ph.Records) == 0 {
		return 0
	}
	return ph.Records[len(ph.Records)-1].Price
}

// GetAveragePrice calculates the average price from history
func (ph *PriceHistory) GetAveragePrice() int {
	ph.mu.RLock()
	defer ph.mu.RUnlock()

	if len(ph.Records) == 0 {
		return 0
	}

	sum := 0
	for _, record := range ph.Records {
		sum += record.Price
	}
	return sum / len(ph.Records)
}

// GetPriceTrend analyzes the price trend
func (ph *PriceHistory) GetPriceTrend() PriceTrend {
	ph.mu.RLock()
	defer ph.mu.RUnlock()

	if len(ph.Records) < 2 {
		return PriceTrendStable
	}

	firstPrice := ph.Records[0].Price
	lastPrice := ph.Records[len(ph.Records)-1].Price

	if lastPrice > firstPrice {
		return PriceTrendUp
	} else if lastPrice < firstPrice {
		return PriceTrendDown
	}
	return PriceTrendStable
}
