package ai

// MockInventory is a simple inventory implementation for AI merchants
type MockInventory struct {
	capacity int
	items    map[string]int
}

// NewMockInventory creates a new mock inventory
func NewMockInventory(capacity int) *MockInventory {
	return &MockInventory{
		capacity: capacity,
		items:    make(map[string]int),
	}
}

// MarketData represents market data for AI decision making
type MarketData struct {
	Items []ItemData
}

// ItemData represents data about an item in the market
type ItemData struct {
	ItemID       string
	CurrentPrice float64
	BasePrice    float64
	Supply       int
	Demand       int
	Volatility   float64
	PriceHistory []float64
	Category     int
	Tags         []string
}
