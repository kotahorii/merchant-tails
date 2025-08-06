package trading

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/merchant-tails/game/internal/domain/inventory"
	"github.com/yourusername/merchant-tails/game/internal/domain/item"
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeBuy   TransactionType = "BUY"
	TransactionTypeSell  TransactionType = "SELL"
	TransactionTypeTrade TransactionType = "TRADE"
)

// OrderType represents the type of market order
type OrderType string

const (
	OrderTypeBuy  OrderType = "BUY"
	OrderTypeSell OrderType = "SELL"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusCompleted OrderStatus = "COMPLETED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

// Transaction represents a trading transaction
type Transaction struct {
	ID        string
	Type      TransactionType
	ItemID    string
	ItemName  string
	Quantity  int
	UnitPrice int
	TotalCost int
	Timestamp time.Time
	Partner   string // Who we traded with
}

// MarketOrder represents a market order
type MarketOrder struct {
	ID          string
	ItemID      string
	Quantity    int
	Type        OrderType
	PriceLimit  int // 0 for market order, >0 for limit order
	Status      OrderStatus
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// SpecialDeal represents a time-limited special offer
type SpecialDeal struct {
	ID           string
	ItemID       string
	Quantity     int
	SpecialPrice int
	Supplier     string
	ExpiresAt    time.Time
}

// TradingSystem manages all trading operations
type TradingSystem struct {
	inventory        *inventory.InventoryManager
	market           *market.MarketSystem
	gold             int
	transactions     []*Transaction
	orders           map[string]*MarketOrder
	specialDeals     map[string]*SpecialDeal
	purchasePrices   map[string]int // Track purchase prices for profit calculation
	reputation       int            // 0-100, affects prices
	negotiationSkill int            // Percentage discount/markup ability
	fairPricing      bool
	totalProfit      int
	mu               sync.RWMutex
}

// NewTradingSystem creates a new trading system
func NewTradingSystem(inv *inventory.InventoryManager, mkt *market.MarketSystem) (*TradingSystem, error) {
	if inv == nil {
		return nil, errors.New("inventory manager is required")
	}
	if mkt == nil {
		return nil, errors.New("market system is required")
	}

	return &TradingSystem{
		inventory:        inv,
		market:           mkt,
		gold:             0,
		transactions:     make([]*Transaction, 0),
		orders:           make(map[string]*MarketOrder),
		specialDeals:     make(map[string]*SpecialDeal),
		purchasePrices:   make(map[string]int),
		reputation:       50, // Start with neutral reputation
		negotiationSkill: 0,
		fairPricing:      false,
		totalProfit:      0,
	}, nil
}

// SetGold sets the player's gold amount
func (ts *TradingSystem) SetGold(amount int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.gold = amount
}

// GetGold returns the current gold amount
func (ts *TradingSystem) GetGold() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.gold
}

// BuyFromSupplier purchases items from a supplier
func (ts *TradingSystem) BuyFromSupplier(item *item.Item, quantity int) (*Transaction, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Get current market price
	unitPrice := ts.market.GetCurrentPrice(item.ID)
	if unitPrice == 0 {
		unitPrice = item.BasePrice
	}

	// Apply negotiation if buying
	unitPrice = ts.negotiatePriceUnsafe(unitPrice, true)

	totalCost := unitPrice * quantity

	// Check if player has enough gold
	if ts.gold < totalCost {
		return nil, fmt.Errorf("insufficient gold: need %d, have %d", totalCost, ts.gold)
	}

	// Check inventory capacity
	totalCapacity := ts.inventory.ShopCapacity + ts.inventory.WarehouseCapacity
	currentTotal := ts.inventory.GetTotalShopItems()
	if currentTotal+quantity > totalCapacity {
		return nil, errors.New("insufficient inventory capacity")
	}

	// Perform the purchase
	ts.gold -= totalCost

	// Add to warehouse first (can transfer to shop later)
	err := ts.inventory.AddToWarehouse(item, quantity)
	if err != nil {
		ts.gold += totalCost // Rollback gold change
		return nil, err
	}

	// Record purchase price for profit tracking
	ts.purchasePrices[item.ID] = unitPrice

	// Create transaction record
	transaction := &Transaction{
		ID:        uuid.New().String(),
		Type:      TransactionTypeBuy,
		ItemID:    item.ID,
		ItemName:  item.Name,
		Quantity:  quantity,
		UnitPrice: unitPrice,
		TotalCost: totalCost,
		Timestamp: time.Now(),
		Partner:   "Supplier",
	}

	ts.transactions = append(ts.transactions, transaction)

	return transaction, nil
}

// SellToCustomer sells items to a customer
func (ts *TradingSystem) SellToCustomer(itemID string, quantity int, customerBudget int) (*Transaction, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Check if we have enough stock in shop
	shopQuantity := ts.inventory.GetShopQuantity(itemID)
	if shopQuantity < quantity {
		return nil, fmt.Errorf("insufficient stock: have %d, need %d", shopQuantity, quantity)
	}

	// Get current market price
	unitPrice := ts.market.GetCurrentPrice(itemID)
	if unitPrice == 0 {
		// Fallback to base price if market price not available
		unitPrice = 10 // Default fallback
	}

	// Apply negotiation if selling
	unitPrice = ts.negotiatePriceUnsafe(unitPrice, false)

	// Apply reputation adjustment
	unitPrice = ts.getReputationAdjustedPriceUnsafe(unitPrice, false)

	totalPrice := unitPrice * quantity

	// Check if customer can afford it
	if customerBudget < totalPrice {
		return nil, fmt.Errorf("customer cannot afford: price %d, budget %d", totalPrice, customerBudget)
	}

	// Remove from shop inventory
	err := ts.inventory.ShopInventory.RemoveItem(itemID, quantity)
	if err != nil {
		return nil, err
	}

	// Add gold
	ts.gold += totalPrice

	// Calculate and track profit
	if purchasePrice, exists := ts.purchasePrices[itemID]; exists {
		profit := (unitPrice - purchasePrice) * quantity
		ts.totalProfit += profit
	}

	// Improve reputation for successful sale
	ts.improveReputationUnsafe(1)

	// Create transaction record
	transaction := &Transaction{
		ID:        uuid.New().String(),
		Type:      TransactionTypeSell,
		ItemID:    itemID,
		ItemName:  "Item", // Would need item registry to get name
		Quantity:  quantity,
		UnitPrice: unitPrice,
		TotalCost: totalPrice,
		Timestamp: time.Now(),
		Partner:   "Customer",
	}

	ts.transactions = append(ts.transactions, transaction)

	return transaction, nil
}

// PlaceMarketOrder places a market or limit order
func (ts *TradingSystem) PlaceMarketOrder(itemID string, quantity int, orderType OrderType, priceLimit int) (*MarketOrder, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	order := &MarketOrder{
		ID:         uuid.New().String(),
		ItemID:     itemID,
		Quantity:   quantity,
		Type:       orderType,
		PriceLimit: priceLimit,
		Status:     OrderStatusPending,
		CreatedAt:  time.Now(),
	}

	ts.orders[order.ID] = order
	return order, nil
}

// ProcessOrder attempts to execute a pending order
func (ts *TradingSystem) ProcessOrder(order *MarketOrder) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if order.Status != OrderStatusPending {
		return false
	}

	currentPrice := ts.market.GetCurrentPrice(order.ItemID)

	// Check if price conditions are met
	if order.PriceLimit > 0 {
		if order.Type == OrderTypeSell && currentPrice < order.PriceLimit {
			// Selling but price too low
			return false
		}
		if order.Type == OrderTypeBuy && currentPrice > order.PriceLimit {
			// Buying but price too high
			return false
		}
	}

	// Execute the order
	order.Status = OrderStatusCompleted
	now := time.Now()
	order.CompletedAt = &now

	return true
}

// SetNegotiationSkill sets the negotiation skill level
func (ts *TradingSystem) SetNegotiationSkill(skill int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.negotiationSkill = skill
}

// NegotiatePrice applies negotiation skill to get better prices
func (ts *TradingSystem) NegotiatePrice(basePrice int, isBuying bool) int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.negotiatePriceUnsafe(basePrice, isBuying)
}

func (ts *TradingSystem) negotiatePriceUnsafe(basePrice int, isBuying bool) int {
	if ts.negotiationSkill == 0 {
		return basePrice
	}

	adjustment := float64(basePrice) * float64(ts.negotiationSkill) / 100

	if isBuying {
		// When buying, we want lower prices
		return basePrice - int(adjustment)
	} else {
		// When selling, we want higher prices
		return basePrice + int(adjustment)
	}
}

// CalculateBulkDiscount calculates discount based on quantity
func (ts *TradingSystem) CalculateBulkDiscount(quantity int) float64 {
	switch {
	case quantity >= 50:
		return 0.15 // 15% discount
	case quantity >= 20:
		return 0.10 // 10% discount
	case quantity >= 10:
		return 0.05 // 5% discount
	default:
		return 0
	}
}

// CalculateTotalWithDiscount calculates total price with bulk discount
func (ts *TradingSystem) CalculateTotalWithDiscount(unitPrice, quantity int) int {
	discount := ts.CalculateBulkDiscount(quantity)
	total := float64(unitPrice * quantity)
	return int(total * (1 - discount))
}

// GetTransactionHistory returns all transactions
func (ts *TradingSystem) GetTransactionHistory() []*Transaction {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.transactions
}

// GetTransactionsByType filters transactions by type
func (ts *TradingSystem) GetTransactionsByType(transType TransactionType) []*Transaction {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	filtered := make([]*Transaction, 0)
	for _, t := range ts.transactions {
		if t.Type == transType {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// GetTransactionsByDateRange filters transactions by date range
func (ts *TradingSystem) GetTransactionsByDateRange(start, end time.Time) []*Transaction {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	filtered := make([]*Transaction, 0)
	for _, t := range ts.transactions {
		if t.Timestamp.After(start) && t.Timestamp.Before(end) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// RecordPurchasePrice records the purchase price of an item
func (ts *TradingSystem) RecordPurchasePrice(itemID string, price int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.purchasePrices[itemID] = price
}

// CalculateProfit calculates profit for a specific item sale
func (ts *TradingSystem) CalculateProfit(itemID string, quantity int) int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	purchasePrice, exists := ts.purchasePrices[itemID]
	if !exists {
		return 0
	}

	currentPrice := ts.market.GetCurrentPrice(itemID)
	if currentPrice == 0 {
		currentPrice = 10 // Default
	}

	return (currentPrice - purchasePrice) * quantity
}

// GetTotalProfit returns the total profit made
func (ts *TradingSystem) GetTotalProfit() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.totalProfit
}

// GetProfitMargin calculates profit margin percentage for an item
func (ts *TradingSystem) GetProfitMargin(itemID string) float64 {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	purchasePrice, exists := ts.purchasePrices[itemID]
	if !exists || purchasePrice == 0 {
		return 0
	}

	currentPrice := ts.market.GetCurrentPrice(itemID)
	if currentPrice == 0 {
		currentPrice = 10 // Default
	}

	return float64(currentPrice-purchasePrice) / float64(purchasePrice) * 100
}

// GetReputation returns the current reputation
func (ts *TradingSystem) GetReputation() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.reputation
}

// SetFairPricing sets whether fair pricing is enabled
func (ts *TradingSystem) SetFairPricing(fair bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.fairPricing = fair
}

// improveReputationUnsafe improves reputation (must be called with lock held)
func (ts *TradingSystem) improveReputationUnsafe(amount int) {
	ts.reputation += amount
	if ts.reputation > 100 {
		ts.reputation = 100
	}
}

// GetReputationAdjustedPrice adjusts price based on reputation
func (ts *TradingSystem) GetReputationAdjustedPrice(basePrice int, isBuying bool) int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.getReputationAdjustedPriceUnsafe(basePrice, isBuying)
}

func (ts *TradingSystem) getReputationAdjustedPriceUnsafe(basePrice int, isBuying bool) int {
	// Reputation affects prices: high reputation = better prices
	repModifier := float64(ts.reputation-50) / 100.0 // -0.5 to +0.5

	if isBuying {
		// When buying, high reputation gets discounts
		return basePrice - int(float64(basePrice)*repModifier*0.1)
	} else {
		// When selling, high reputation allows higher prices
		return basePrice + int(float64(basePrice)*repModifier*0.1)
	}
}

// ProcessDailyReputationDecay slowly moves reputation towards neutral
func (ts *TradingSystem) ProcessDailyReputationDecay() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.reputation > 50 {
		ts.reputation--
	} else if ts.reputation < 50 {
		ts.reputation++
	}
}

// AddSpecialDeal adds a special deal offer
func (ts *TradingSystem) AddSpecialDeal(deal *SpecialDeal) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if deal.ID == "" {
		deal.ID = uuid.New().String()
	}
	ts.specialDeals[deal.ID] = deal
}

// GetAvailableDeals returns non-expired special deals
func (ts *TradingSystem) GetAvailableDeals() []*SpecialDeal {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	deals := make([]*SpecialDeal, 0)
	now := time.Now()

	for _, deal := range ts.specialDeals {
		if deal.ExpiresAt.After(now) {
			deals = append(deals, deal)
		}
	}

	return deals
}

// AcceptSpecialDeal accepts and processes a special deal
func (ts *TradingSystem) AcceptSpecialDeal(dealID string) (*Transaction, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	deal, exists := ts.specialDeals[dealID]
	if !exists {
		return nil, errors.New("deal not found")
	}

	if deal.ExpiresAt.Before(time.Now()) {
		delete(ts.specialDeals, dealID)
		return nil, errors.New("deal has expired")
	}

	totalCost := deal.SpecialPrice * deal.Quantity

	if ts.gold < totalCost {
		return nil, fmt.Errorf("insufficient gold: need %d, have %d", totalCost, ts.gold)
	}

	ts.gold -= totalCost

	// Create transaction
	transaction := &Transaction{
		ID:        uuid.New().String(),
		Type:      TransactionTypeBuy,
		ItemID:    deal.ItemID,
		ItemName:  "Special Deal Item",
		Quantity:  deal.Quantity,
		UnitPrice: deal.SpecialPrice,
		TotalCost: totalCost,
		Timestamp: time.Now(),
		Partner:   deal.Supplier,
	}

	ts.transactions = append(ts.transactions, transaction)

	// Remove the accepted deal
	delete(ts.specialDeals, dealID)

	return transaction, nil
}

// CleanExpiredDeals removes expired special deals
func (ts *TradingSystem) CleanExpiredDeals() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	now := time.Now()
	for id, deal := range ts.specialDeals {
		if deal.ExpiresAt.Before(now) {
			delete(ts.specialDeals, id)
		}
	}
}
