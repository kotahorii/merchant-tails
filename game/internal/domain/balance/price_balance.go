package balance

import (
	"math"
	"sync"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/item"
)

// PriceBalanceConfig defines the configuration for price balancing
type PriceBalanceConfig struct {
	// Target profit margins for different item categories
	TargetMargins map[item.Category]float64

	// Price volatility limits (min/max multipliers)
	MinPriceMultiplier float64
	MaxPriceMultiplier float64

	// Adjustment speeds
	PriceAdjustmentSpeed  float64 // How quickly prices adjust (0.0-1.0)
	MarginAdjustmentSpeed float64 // How quickly margins adjust (0.0-1.0)

	// Market saturation thresholds
	OversupplyThreshold float64 // When supply > demand * threshold
	ScarcityThreshold   float64 // When supply < demand * threshold

	// Player progression factors
	EarlyGameMultiplier float64 // Price multiplier for early game
	MidGameMultiplier   float64 // Price multiplier for mid game
	LateGameMultiplier  float64 // Price multiplier for late game

	// Economic health targets
	TargetInflationRate float64 // Annual inflation target
	TargetGoldVelocity  float64 // How fast gold should circulate
	MaxWealthGap        float64 // Maximum wealth disparity allowed
}

// DefaultPriceBalanceConfig returns the default configuration
func DefaultPriceBalanceConfig() *PriceBalanceConfig {
	return &PriceBalanceConfig{
		TargetMargins: map[item.Category]float64{
			item.CategoryFruit:     0.30, // 30% margin on fruits
			item.CategoryPotion:    0.45, // 45% margin on potions
			item.CategoryWeapon:    0.40, // 40% margin on weapons
			item.CategoryAccessory: 0.50, // 50% margin on accessories
			item.CategoryMagicBook: 0.55, // 55% margin on magic books
			item.CategoryGem:       0.60, // 60% margin on gems
		},
		MinPriceMultiplier:    0.5,  // Prices can drop to 50% minimum
		MaxPriceMultiplier:    3.0,  // Prices can rise to 300% maximum
		PriceAdjustmentSpeed:  0.1,  // 10% adjustment per update
		MarginAdjustmentSpeed: 0.05, // 5% margin adjustment per update
		OversupplyThreshold:   1.5,  // 150% supply vs demand
		ScarcityThreshold:     0.5,  // 50% supply vs demand
		EarlyGameMultiplier:   0.8,  // 80% prices in early game
		MidGameMultiplier:     1.0,  // 100% prices in mid game
		LateGameMultiplier:    1.3,  // 130% prices in late game
		TargetInflationRate:   0.02, // 2% annual inflation
		TargetGoldVelocity:    5.0,  // Gold should turn over 5x per year
		MaxWealthGap:          10.0, // Max 10x wealth difference
	}
}

// MarketMetrics tracks market health indicators
type MarketMetrics struct {
	TotalTransactions  int
	TotalVolume        float64
	AverageTransaction float64
	PriceIndex         float64 // Overall price level (1.0 = baseline)
	InflationRate      float64
	GoldVelocity       float64
	SupplyDemandRatio  float64
	PlayerProfitMargin float64
	MarketLiquidity    float64
	LastUpdated        time.Time
}

// ItemBalance tracks balance metrics for a specific item
type ItemBalance struct {
	ItemID          string
	Category        item.Category
	BasePrice       float64
	CurrentPrice    float64
	OptimalPrice    float64
	RecentSales     []SaleRecord
	SupplyLevel     int
	DemandLevel     int
	PriceMultiplier float64
	ProfitMargin    float64
	Volatility      float64
	LastAdjustment  time.Time
}

// SaleRecord represents a single sale transaction
type SaleRecord struct {
	Timestamp    time.Time
	Price        float64
	Quantity     int
	BuyerWealth  float64 // Wealth level of buyer
	SellerProfit float64
}

// PriceBalancer manages dynamic price balancing
type PriceBalancer struct {
	config              *PriceBalanceConfig
	metrics             *MarketMetrics
	itemBalance         map[string]*ItemBalance
	priceHistory        map[string][]PricePoint
	adjustmentCallbacks []AdjustmentCallback
	mu                  sync.RWMutex
}

// PricePoint represents a price at a point in time
type PricePoint struct {
	Timestamp time.Time
	Price     float64
	Volume    int
}

// AdjustmentCallback is called when prices are adjusted
type AdjustmentCallback func(itemID string, oldPrice, newPrice float64, reason string)

// NewPriceBalancer creates a new price balancer
func NewPriceBalancer(config *PriceBalanceConfig) *PriceBalancer {
	if config == nil {
		config = DefaultPriceBalanceConfig()
	}

	return &PriceBalancer{
		config:              config,
		metrics:             &MarketMetrics{LastUpdated: time.Now()},
		itemBalance:         make(map[string]*ItemBalance),
		priceHistory:        make(map[string][]PricePoint),
		adjustmentCallbacks: make([]AdjustmentCallback, 0),
	}
}

// RegisterItem registers an item for balance tracking
func (pb *PriceBalancer) RegisterItem(itemID string, category item.Category, basePrice float64) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.itemBalance[itemID] = &ItemBalance{
		ItemID:          itemID,
		Category:        category,
		BasePrice:       basePrice,
		CurrentPrice:    basePrice,
		OptimalPrice:    basePrice,
		RecentSales:     make([]SaleRecord, 0),
		PriceMultiplier: 1.0,
		ProfitMargin:    pb.config.TargetMargins[category],
		LastAdjustment:  time.Now(),
	}
}

// RecordSale records a sale transaction for balance analysis
func (pb *PriceBalancer) RecordSale(itemID string, price float64, quantity int, buyerWealth float64) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if balance, exists := pb.itemBalance[itemID]; exists {
		saleRecord := SaleRecord{
			Timestamp:    time.Now(),
			Price:        price,
			Quantity:     quantity,
			BuyerWealth:  buyerWealth,
			SellerProfit: price - balance.BasePrice,
		}

		balance.RecentSales = append(balance.RecentSales, saleRecord)

		// Keep only recent sales (last 100)
		if len(balance.RecentSales) > 100 {
			balance.RecentSales = balance.RecentSales[len(balance.RecentSales)-100:]
		}

		// Update metrics
		pb.metrics.TotalTransactions++
		pb.metrics.TotalVolume += price * float64(quantity)
		pb.metrics.AverageTransaction = pb.metrics.TotalVolume / float64(pb.metrics.TotalTransactions)

		// Add to price history
		pb.priceHistory[itemID] = append(pb.priceHistory[itemID], PricePoint{
			Timestamp: time.Now(),
			Price:     price,
			Volume:    quantity,
		})
	}
}

// UpdateSupplyDemand updates supply and demand levels for an item
func (pb *PriceBalancer) UpdateSupplyDemand(itemID string, supply, demand int) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if balance, exists := pb.itemBalance[itemID]; exists {
		balance.SupplyLevel = supply
		balance.DemandLevel = demand
	}
}

// CalculateOptimalPrice calculates the optimal price for an item
func (pb *PriceBalancer) CalculateOptimalPrice(itemID string, playerRank int) float64 {
	pb.mu.RLock()
	defer pb.mu.RUnlock()

	balance, exists := pb.itemBalance[itemID]
	if !exists {
		return 0
	}

	// Base optimal price
	optimal := balance.BasePrice

	// Apply supply/demand adjustment
	if balance.DemandLevel > 0 {
		ratio := float64(balance.SupplyLevel) / float64(balance.DemandLevel)

		if ratio > pb.config.OversupplyThreshold {
			// Oversupply - reduce price
			optimal *= 0.8
		} else if ratio < pb.config.ScarcityThreshold {
			// Scarcity - increase price
			optimal *= 1.3
		} else {
			// Normal supply/demand
			optimal *= (2.0 - ratio) // Inverse relationship
		}
	}

	// Apply player progression multiplier
	progressMultiplier := pb.getProgressionMultiplier(playerRank)
	optimal *= progressMultiplier

	// Apply profit margin target
	targetMargin := pb.config.TargetMargins[balance.Category]
	optimal *= (1.0 + targetMargin)

	// Apply volatility based on recent sales
	volatility := pb.calculateVolatility(balance.RecentSales)
	optimal *= (1.0 + volatility*0.1) // Max 10% volatility adjustment

	// Clamp to min/max bounds
	minPrice := balance.BasePrice * pb.config.MinPriceMultiplier
	maxPrice := balance.BasePrice * pb.config.MaxPriceMultiplier
	optimal = math.Max(minPrice, math.Min(maxPrice, optimal))

	return optimal
}

// AdjustPrices performs automatic price adjustments for market balance
func (pb *PriceBalancer) AdjustPrices(playerRank int) map[string]float64 {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	adjustments := make(map[string]float64)

	for itemID, balance := range pb.itemBalance {
		oldPrice := balance.CurrentPrice

		// Calculate optimal price
		optimal := pb.CalculateOptimalPriceInternal(itemID, playerRank)

		// Smooth adjustment towards optimal
		adjustment := (optimal - oldPrice) * pb.config.PriceAdjustmentSpeed
		newPrice := oldPrice + adjustment

		// Update balance
		balance.CurrentPrice = newPrice
		balance.OptimalPrice = optimal
		balance.PriceMultiplier = newPrice / balance.BasePrice
		balance.LastAdjustment = time.Now()

		adjustments[itemID] = newPrice

		// Notify callbacks
		reason := pb.getAdjustmentReason(balance)
		for _, callback := range pb.adjustmentCallbacks {
			callback(itemID, oldPrice, newPrice, reason)
		}
	}

	// Update overall metrics
	pb.updateMetrics()

	return adjustments
}

// CalculateOptimalPriceInternal is the internal version without lock
func (pb *PriceBalancer) CalculateOptimalPriceInternal(itemID string, playerRank int) float64 {
	balance, exists := pb.itemBalance[itemID]
	if !exists {
		return 0
	}

	// Same logic as CalculateOptimalPrice but without locking
	optimal := balance.BasePrice

	if balance.DemandLevel > 0 {
		ratio := float64(balance.SupplyLevel) / float64(balance.DemandLevel)

		if ratio > pb.config.OversupplyThreshold {
			optimal *= 0.8
		} else if ratio < pb.config.ScarcityThreshold {
			optimal *= 1.3
		} else {
			optimal *= (2.0 - ratio)
		}
	}

	progressMultiplier := pb.getProgressionMultiplier(playerRank)
	optimal *= progressMultiplier

	targetMargin := pb.config.TargetMargins[balance.Category]
	optimal *= (1.0 + targetMargin)

	volatility := pb.calculateVolatility(balance.RecentSales)
	optimal *= (1.0 + volatility*0.1)

	minPrice := balance.BasePrice * pb.config.MinPriceMultiplier
	maxPrice := balance.BasePrice * pb.config.MaxPriceMultiplier
	optimal = math.Max(minPrice, math.Min(maxPrice, optimal))

	return optimal
}

// getProgressionMultiplier returns the price multiplier based on player rank
func (pb *PriceBalancer) getProgressionMultiplier(playerRank int) float64 {
	switch playerRank {
	case 1: // Apprentice
		return pb.config.EarlyGameMultiplier
	case 2: // Journeyman
		return pb.config.MidGameMultiplier
	case 3, 4: // Expert, Master
		return pb.config.LateGameMultiplier
	default:
		return 1.0
	}
}

// calculateVolatility calculates price volatility from recent sales
func (pb *PriceBalancer) calculateVolatility(sales []SaleRecord) float64 {
	if len(sales) < 2 {
		return 0
	}

	// Calculate standard deviation of prices
	var sum, sumSq float64
	for _, sale := range sales {
		sum += sale.Price
		sumSq += sale.Price * sale.Price
	}

	n := float64(len(sales))
	mean := sum / n
	variance := (sumSq / n) - (mean * mean)

	if variance < 0 {
		return 0
	}

	stdDev := math.Sqrt(variance)
	return stdDev / mean // Coefficient of variation
}

// getAdjustmentReason determines the reason for price adjustment
func (pb *PriceBalancer) getAdjustmentReason(balance *ItemBalance) string {
	ratio := float64(balance.SupplyLevel) / float64(balance.DemandLevel+1)

	if ratio > pb.config.OversupplyThreshold {
		return "oversupply"
	} else if ratio < pb.config.ScarcityThreshold {
		return "scarcity"
	} else if balance.Volatility > 0.2 {
		return "high_volatility"
	} else {
		return "market_equilibrium"
	}
}

// updateMetrics updates overall market metrics
func (pb *PriceBalancer) updateMetrics() {
	totalSupply := 0
	totalDemand := 0
	totalPrice := 0.0
	count := 0

	for _, balance := range pb.itemBalance {
		totalSupply += balance.SupplyLevel
		totalDemand += balance.DemandLevel
		totalPrice += balance.CurrentPrice
		count++
	}

	if count > 0 {
		pb.metrics.PriceIndex = totalPrice / (float64(count) * 100) // Normalized to 100 base
	}

	if totalDemand > 0 {
		pb.metrics.SupplyDemandRatio = float64(totalSupply) / float64(totalDemand)
	}

	pb.metrics.LastUpdated = time.Now()

	// Calculate inflation rate (simplified)
	if pb.metrics.PriceIndex > 0 {
		pb.metrics.InflationRate = (pb.metrics.PriceIndex - 1.0) * 0.02 // Annual rate
	}
}

// RegisterAdjustmentCallback registers a callback for price adjustments
func (pb *PriceBalancer) RegisterAdjustmentCallback(callback AdjustmentCallback) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.adjustmentCallbacks = append(pb.adjustmentCallbacks, callback)
}

// GetMetrics returns current market metrics
func (pb *PriceBalancer) GetMetrics() *MarketMetrics {
	pb.mu.RLock()
	defer pb.mu.RUnlock()

	// Return a copy
	metrics := *pb.metrics
	return &metrics
}

// GetItemBalance returns balance information for a specific item
func (pb *PriceBalancer) GetItemBalance(itemID string) (*ItemBalance, bool) {
	pb.mu.RLock()
	defer pb.mu.RUnlock()

	balance, exists := pb.itemBalance[itemID]
	if !exists {
		return nil, false
	}

	// Return a copy
	balanceCopy := *balance
	return &balanceCopy, true
}

// Reset resets the price balancer
func (pb *PriceBalancer) Reset() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.metrics = &MarketMetrics{LastUpdated: time.Now()}
	pb.itemBalance = make(map[string]*ItemBalance)
	pb.priceHistory = make(map[string][]PricePoint)
}
