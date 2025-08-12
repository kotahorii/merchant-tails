package api

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/item"
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
)

// Demand level constants
const (
	demandVeryHigh = "very_high"
	demandHigh     = "high"
	demandNormal   = "normal"
	demandLow      = "low"
	demandVeryLow  = "very_low"
)

// PriceSettingItem represents an item for price setting
type PriceSettingItem struct {
	ItemID           string        `json:"item_id"`
	Name             string        `json:"name"`
	Category         item.Category `json:"category"`
	Quantity         int           `json:"quantity"`
	PurchasePrice    float64       `json:"purchase_price"`
	CurrentPrice     float64       `json:"current_price"`
	MarketPrice      float64       `json:"market_price"`
	CompetitorPrice  float64       `json:"competitor_price"`
	RecommendedPrice float64       `json:"recommended_price"`
	MinPrice         float64       `json:"min_price"`
	MaxPrice         float64       `json:"max_price"`
	ProfitMargin     float64       `json:"profit_margin"`
	DemandLevel      string        `json:"demand_level"` // "very_low", "low", "normal", "high", "very_high"
	Elasticity       float64       `json:"elasticity"`   // Price elasticity of demand
	ExpectedSales    int           `json:"expected_sales"`
	Icon             string        `json:"icon"`
}

// PriceUpdateRequest represents a request to update item price
type PriceUpdateRequest struct {
	ItemID   string  `json:"item_id"`
	NewPrice float64 `json:"new_price"`
	Strategy string  `json:"strategy"` // "manual", "competitive", "profit_max", "volume_max"
}

// PriceUpdateResult represents the result of a price update
type PriceUpdateResult struct {
	Success          bool    `json:"success"`
	ItemID           string  `json:"item_id"`
	OldPrice         float64 `json:"old_price"`
	NewPrice         float64 `json:"new_price"`
	ExpectedRevenue  float64 `json:"expected_revenue"`
	ExpectedProfit   float64 `json:"expected_profit"`
	MarketComparison string  `json:"market_comparison"` // "below", "at", "above" market
	Message          string  `json:"message"`
}

// BulkPriceRequest represents a bulk price update request
type BulkPriceRequest struct {
	Updates  []PriceUpdateRequest `json:"updates"`
	Strategy string               `json:"strategy"`
}

// PricingStrategy represents different pricing strategies
type PricingStrategy struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	TargetProfit float64 `json:"target_profit"`
	RiskLevel    string  `json:"risk_level"`
}

// PriceAnalytics represents price analytics data
type PriceAnalytics struct {
	ItemID          string    `json:"item_id"`
	RevenueHistory  []float64 `json:"revenue_history"`
	ProfitHistory   []float64 `json:"profit_history"`
	SalesHistory    []int     `json:"sales_history"`
	OptimalPrice    float64   `json:"optimal_price"`
	PriceElasticity float64   `json:"price_elasticity"`
	LastUpdated     time.Time `json:"last_updated"`
}

// PriceRule represents an automated pricing rule
type PriceRule struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Condition  string  `json:"condition"` // "inventory_high", "demand_low", "competition", etc.
	Action     string  `json:"action"`    // "decrease_10", "match_market", "undercut", etc.
	Priority   int     `json:"priority"`
	Enabled    bool    `json:"enabled"`
	AppliesTo  string  `json:"applies_to"` // "all", "category:fruit", "item:apple", etc.
	Adjustment float64 `json:"adjustment"` // Percentage or fixed adjustment
}

// PriceSettingUIManager manages the price setting UI backend
type PriceSettingUIManager struct {
	gameManager      *GameManager
	market           *market.Market
	itemPrices       map[string]float64
	priceHistory     map[string][]PricePoint
	analytics        map[string]*PriceAnalytics
	strategies       map[string]*PricingStrategy
	rules            []*PriceRule
	competitorPrices map[string]float64
	mu               sync.RWMutex
}

// PricePoint represents a historical price point
type PricePoint struct {
	Price     float64
	Timestamp time.Time
	Sales     int
	Revenue   float64
}

// NewPriceSettingUIManager creates a new price setting UI manager
func NewPriceSettingUIManager(gameManager *GameManager) *PriceSettingUIManager {
	return &PriceSettingUIManager{
		gameManager:      gameManager,
		market:           gameManager.market,
		itemPrices:       make(map[string]float64),
		priceHistory:     make(map[string][]PricePoint),
		analytics:        make(map[string]*PriceAnalytics),
		strategies:       createDefaultStrategies(),
		rules:            createDefaultRules(),
		competitorPrices: make(map[string]float64),
	}
}

// createDefaultStrategies creates default pricing strategies
func createDefaultStrategies() map[string]*PricingStrategy {
	strategies := make(map[string]*PricingStrategy)

	strategies["competitive"] = &PricingStrategy{
		ID:           "competitive",
		Name:         "Competitive Pricing",
		Description:  "Match or slightly undercut market prices",
		TargetProfit: 15.0,
		RiskLevel:    "low",
	}

	strategies["profit_max"] = &PricingStrategy{
		ID:           "profit_max",
		Name:         "Profit Maximization",
		Description:  "Set prices to maximize profit margins",
		TargetProfit: 30.0,
		RiskLevel:    "medium",
	}

	strategies["volume_max"] = &PricingStrategy{
		ID:           "volume_max",
		Name:         "Volume Maximization",
		Description:  "Lower prices to increase sales volume",
		TargetProfit: 10.0,
		RiskLevel:    "low",
	}

	strategies["premium"] = &PricingStrategy{
		ID:           "premium",
		Name:         "Premium Pricing",
		Description:  "Set high prices for perceived value",
		TargetProfit: 50.0,
		RiskLevel:    "high",
	}

	strategies["dynamic"] = &PricingStrategy{
		ID:           "dynamic",
		Name:         "Dynamic Pricing",
		Description:  "Adjust prices based on demand and competition",
		TargetProfit: 20.0,
		RiskLevel:    "medium",
	}

	return strategies
}

// createDefaultRules creates default pricing rules
func createDefaultRules() []*PriceRule {
	return []*PriceRule{
		{
			ID:         "rule_perishable",
			Name:       "Perishable Discount",
			Condition:  "item_expiring",
			Action:     "decrease_percentage",
			Priority:   1,
			Enabled:    true,
			AppliesTo:  "category:fruit",
			Adjustment: 20.0,
		},
		{
			ID:         "rule_high_inventory",
			Name:       "High Inventory Sale",
			Condition:  "inventory_high",
			Action:     "decrease_percentage",
			Priority:   2,
			Enabled:    true,
			AppliesTo:  "all",
			Adjustment: 15.0,
		},
		{
			ID:         "rule_low_demand",
			Name:       "Low Demand Adjustment",
			Condition:  "demand_low",
			Action:     "decrease_percentage",
			Priority:   3,
			Enabled:    true,
			AppliesTo:  "all",
			Adjustment: 10.0,
		},
		{
			ID:         "rule_competition",
			Name:       "Match Competition",
			Condition:  "competitor_lower",
			Action:     "match_competitor",
			Priority:   4,
			Enabled:    true,
			AppliesTo:  "all",
			Adjustment: 0.0,
		},
	}
}

// GetPriceSettingItems returns all items available for price setting
func (psu *PriceSettingUIManager) GetPriceSettingItems(filter string) ([]*PriceSettingItem, error) {
	psu.mu.RLock()
	defer psu.mu.RUnlock()

	items := make([]*PriceSettingItem, 0)

	// Get shop inventory items
	shopItems := psu.gameManager.inventory.ShopInventory.GetAll()

	for itemID, quantity := range shopItems {
		if quantity == 0 {
			continue
		}

		// Filter by category if specified
		if filter != "" && filter != "all" {
			category := psu.getItemCategory(itemID)
			if string(category) != filter {
				continue
			}
		}

		// Get various prices
		currentPrice := psu.getCurrentPrice(itemID)
		marketPrice := float64(psu.market.GetPrice(itemID))
		competitorPrice := psu.getCompetitorPrice(itemID)
		purchasePrice := psu.getPurchasePrice(itemID)

		// Calculate recommended price
		recommendedPrice := psu.calculateRecommendedPrice(itemID, marketPrice, competitorPrice, purchasePrice)

		// Calculate price bounds
		minPrice := purchasePrice * 1.05 // At least 5% markup
		maxPrice := marketPrice * 2.0    // Max 2x market price

		// Calculate profit margin
		profitMargin := 0.0
		if currentPrice > 0 {
			profitMargin = ((currentPrice - purchasePrice) / purchasePrice) * 100
		}

		// Get demand level
		demandLevel := psu.getDemandLevel(itemID)

		// Calculate elasticity
		elasticity := psu.calculateElasticity(itemID)

		// Estimate sales at current price
		expectedSales := psu.estimateSales(itemID, currentPrice, elasticity)

		item := &PriceSettingItem{
			ItemID:           itemID,
			Name:             psu.getItemName(itemID),
			Category:         psu.getItemCategory(itemID),
			Quantity:         quantity,
			PurchasePrice:    purchasePrice,
			CurrentPrice:     currentPrice,
			MarketPrice:      marketPrice,
			CompetitorPrice:  competitorPrice,
			RecommendedPrice: recommendedPrice,
			MinPrice:         minPrice,
			MaxPrice:         maxPrice,
			ProfitMargin:     profitMargin,
			DemandLevel:      demandLevel,
			Elasticity:       elasticity,
			ExpectedSales:    expectedSales,
			Icon:             fmt.Sprintf("res://assets/items/%s.png", itemID),
		}

		items = append(items, item)
	}

	// Sort by name by default
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return items, nil
}

// UpdatePrice updates the price of a single item
func (psu *PriceSettingUIManager) UpdatePrice(request *PriceUpdateRequest) (*PriceUpdateResult, error) {
	psu.mu.Lock()
	defer psu.mu.Unlock()

	// Validate request
	if request.NewPrice <= 0 {
		return &PriceUpdateResult{
			Success: false,
			Message: "Price must be positive",
		}, nil
	}

	// Get current price
	oldPrice := psu.getCurrentPrice(request.ItemID)

	// Apply strategy if specified
	finalPrice := request.NewPrice
	if request.Strategy != "manual" && request.Strategy != "" {
		finalPrice = psu.applyStrategy(request.ItemID, request.Strategy)
	}

	// Validate price bounds
	purchasePrice := psu.getPurchasePrice(request.ItemID)
	if finalPrice < purchasePrice {
		return &PriceUpdateResult{
			Success: false,
			Message: fmt.Sprintf("Price cannot be below purchase price (%.2f)", purchasePrice),
		}, nil
	}

	// Update the price
	psu.itemPrices[request.ItemID] = finalPrice

	// Record price history
	psu.recordPriceChange(request.ItemID, finalPrice)

	// Calculate expected outcomes
	elasticity := psu.calculateElasticity(request.ItemID)
	expectedSales := psu.estimateSales(request.ItemID, finalPrice, elasticity)
	expectedRevenue := finalPrice * float64(expectedSales)
	expectedProfit := (finalPrice - purchasePrice) * float64(expectedSales)

	// Determine market comparison
	marketPrice := float64(psu.market.GetPrice(request.ItemID))
	marketComparison := "at"
	if finalPrice < marketPrice*0.95 {
		marketComparison = "below"
	} else if finalPrice > marketPrice*1.05 {
		marketComparison = "above"
	}

	return &PriceUpdateResult{
		Success:          true,
		ItemID:           request.ItemID,
		OldPrice:         oldPrice,
		NewPrice:         finalPrice,
		ExpectedRevenue:  expectedRevenue,
		ExpectedProfit:   expectedProfit,
		MarketComparison: marketComparison,
		Message:          "Price updated successfully",
	}, nil
}

// BulkUpdatePrices updates prices for multiple items
func (psu *PriceSettingUIManager) BulkUpdatePrices(request *BulkPriceRequest) ([]*PriceUpdateResult, error) {
	results := make([]*PriceUpdateResult, 0, len(request.Updates))

	for _, update := range request.Updates {
		// Override individual strategies with bulk strategy if specified
		if request.Strategy != "" {
			update.Strategy = request.Strategy
		}

		result, err := psu.UpdatePrice(&update)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}

	return results, nil
}

// GetPricingStrategies returns available pricing strategies
func (psu *PriceSettingUIManager) GetPricingStrategies() []*PricingStrategy {
	psu.mu.RLock()
	defer psu.mu.RUnlock()

	strategies := make([]*PricingStrategy, 0, len(psu.strategies))
	for _, strategy := range psu.strategies {
		strategies = append(strategies, strategy)
	}

	return strategies
}

// ApplyStrategy applies a pricing strategy to items
func (psu *PriceSettingUIManager) ApplyStrategy(strategyID string, itemIDs []string) ([]*PriceUpdateResult, error) {
	psu.mu.RLock()
	strategy, exists := psu.strategies[strategyID]
	psu.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("strategy not found: %s", strategyID)
	}

	results := make([]*PriceUpdateResult, 0, len(itemIDs))

	for _, itemID := range itemIDs {
		price := psu.calculateStrategyPrice(itemID, strategy)

		request := &PriceUpdateRequest{
			ItemID:   itemID,
			NewPrice: price,
			Strategy: strategyID,
		}

		result, err := psu.UpdatePrice(request)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}

	return results, nil
}

// GetPriceAnalytics returns analytics for an item
func (psu *PriceSettingUIManager) GetPriceAnalytics(itemID string) (*PriceAnalytics, error) {
	psu.mu.RLock()
	defer psu.mu.RUnlock()

	if analytics, exists := psu.analytics[itemID]; exists {
		return analytics, nil
	}

	// Generate analytics if not cached
	analytics := psu.generateAnalytics(itemID)
	psu.analytics[itemID] = analytics

	return analytics, nil
}

// GetPricingRules returns all pricing rules
func (psu *PriceSettingUIManager) GetPricingRules() []*PriceRule {
	psu.mu.RLock()
	defer psu.mu.RUnlock()

	return psu.rules
}

// ToggleRule enables or disables a pricing rule
func (psu *PriceSettingUIManager) ToggleRule(ruleID string, enabled bool) error {
	psu.mu.Lock()
	defer psu.mu.Unlock()

	for _, rule := range psu.rules {
		if rule.ID == ruleID {
			rule.Enabled = enabled
			return nil
		}
	}

	return fmt.Errorf("rule not found: %s", ruleID)
}

// ApplyRules automatically applies enabled pricing rules
func (psu *PriceSettingUIManager) ApplyRules() ([]*PriceUpdateResult, error) {
	psu.mu.Lock()
	defer psu.mu.Unlock()

	results := make([]*PriceUpdateResult, 0)

	// Sort rules by priority
	sort.Slice(psu.rules, func(i, j int) bool {
		return psu.rules[i].Priority < psu.rules[j].Priority
	})

	// Get all shop items
	shopItems := psu.gameManager.inventory.ShopInventory.GetAll()

	for itemID := range shopItems {
		for _, rule := range psu.rules {
			if !rule.Enabled {
				continue
			}

			// Check if rule applies to this item
			if !psu.ruleApplies(rule, itemID) {
				continue
			}

			// Check if condition is met
			if !psu.checkCondition(rule, itemID) {
				continue
			}

			// Apply the rule
			newPrice := psu.applyRule(rule, itemID)

			request := &PriceUpdateRequest{
				ItemID:   itemID,
				NewPrice: newPrice,
				Strategy: "rule_" + rule.ID,
			}

			result, _ := psu.UpdatePrice(request)
			if result.Success {
				results = append(results, result)
			}
		}
	}

	return results, nil
}

// Helper functions

func (psu *PriceSettingUIManager) getCurrentPrice(itemID string) float64 {
	if price, exists := psu.itemPrices[itemID]; exists {
		return price
	}
	// Default to market price
	return float64(psu.market.GetPrice(itemID))
}

func (psu *PriceSettingUIManager) getCompetitorPrice(itemID string) float64 {
	if price, exists := psu.competitorPrices[itemID]; exists {
		return price
	}
	// Simulate competitor price as 95-105% of market
	marketPrice := float64(psu.market.GetPrice(itemID))
	return marketPrice * (0.95 + 0.1*math.Sin(float64(time.Now().Unix())))
}

func (psu *PriceSettingUIManager) getPurchasePrice(itemID string) float64 {
	// Simplified - in production would track actual purchase prices
	return float64(psu.market.GetPrice(itemID)) * 0.7
}

func (psu *PriceSettingUIManager) calculateRecommendedPrice(itemID string, marketPrice, competitorPrice, purchasePrice float64) float64 {
	// Basic recommendation algorithm
	targetMargin := 0.25 // 25% profit margin
	minPrice := purchasePrice * (1 + targetMargin)

	// Consider market and competitor prices
	avgMarketPrice := (marketPrice + competitorPrice) / 2

	// Adjust based on demand
	demandMultiplier := 1.0
	demandLevel := psu.getDemandLevel(itemID)
	switch demandLevel {
	case demandVeryHigh:
		demandMultiplier = 1.2
	case demandHigh:
		demandMultiplier = 1.1
	case demandLow:
		demandMultiplier = 0.9
	case demandVeryLow:
		demandMultiplier = 0.8
	}

	recommendedPrice := avgMarketPrice * demandMultiplier

	// Ensure minimum profit margin
	if recommendedPrice < minPrice {
		recommendedPrice = minPrice
	}

	return recommendedPrice
}

func (psu *PriceSettingUIManager) getDemandLevel(itemID string) string {
	// Use market state to determine demand
	state := psu.market.State
	if state == nil {
		return demandNormal
	}

	demand := state.CurrentDemand

	if demand == market.DemandVeryHigh {
		return demandVeryHigh
	} else if demand == market.DemandHigh {
		return demandHigh
	} else if demand == market.DemandLow {
		return demandLow
	} else if demand == market.DemandVeryLow {
		return demandVeryLow
	}

	return demandNormal
}

func (psu *PriceSettingUIManager) calculateElasticity(itemID string) float64 {
	// Simplified elasticity calculation
	// Luxury items are more elastic, necessities less elastic
	category := psu.getItemCategory(itemID)

	switch category {
	case item.CategoryFruit:
		return 0.5 // Inelastic (necessity)
	case item.CategoryPotion:
		return 0.7 // Somewhat inelastic
	case item.CategoryWeapon:
		return 1.0 // Unit elastic
	case item.CategoryAccessory:
		return 1.5 // Elastic
	case item.CategoryGem:
		return 2.0 // Very elastic (luxury)
	case item.CategoryMagicBook:
		return 1.8 // Elastic
	default:
		return 1.0
	}
}

func (psu *PriceSettingUIManager) estimateSales(itemID string, price, elasticity float64) int {
	// Estimate sales based on price and elasticity
	baselineSales := 10
	marketPrice := float64(psu.market.GetPrice(itemID))

	if marketPrice == 0 {
		return baselineSales
	}

	// Calculate price change percentage
	priceChange := (price - marketPrice) / marketPrice

	// Apply elasticity to estimate quantity change
	quantityChange := -priceChange * elasticity

	// Calculate expected sales
	expectedSales := float64(baselineSales) * (1 + quantityChange)

	if expectedSales < 0 {
		expectedSales = 0
	}

	return int(expectedSales)
}

func (psu *PriceSettingUIManager) recordPriceChange(itemID string, newPrice float64) {
	if _, exists := psu.priceHistory[itemID]; !exists {
		psu.priceHistory[itemID] = make([]PricePoint, 0, 100)
	}

	point := PricePoint{
		Price:     newPrice,
		Timestamp: time.Now(),
		Sales:     0, // Will be updated when sales occur
		Revenue:   0,
	}

	psu.priceHistory[itemID] = append(psu.priceHistory[itemID], point)

	// Keep only last 100 points
	if len(psu.priceHistory[itemID]) > 100 {
		psu.priceHistory[itemID] = psu.priceHistory[itemID][1:]
	}
}

func (psu *PriceSettingUIManager) applyStrategy(itemID, strategyID string) float64 {
	strategy, exists := psu.strategies[strategyID]
	if !exists {
		return psu.getCurrentPrice(itemID)
	}

	return psu.calculateStrategyPrice(itemID, strategy)
}

func (psu *PriceSettingUIManager) calculateStrategyPrice(itemID string, strategy *PricingStrategy) float64 {
	purchasePrice := psu.getPurchasePrice(itemID)
	marketPrice := float64(psu.market.GetPrice(itemID))
	competitorPrice := psu.getCompetitorPrice(itemID)

	switch strategy.ID {
	case "competitive":
		// Match or slightly undercut competitors
		return competitorPrice * 0.98

	case "profit_max":
		// Set price for target profit margin
		return purchasePrice * (1 + strategy.TargetProfit/100)

	case "volume_max":
		// Price below market to increase volume
		return marketPrice * 0.85

	case "premium":
		// Price above market for premium positioning
		return marketPrice * 1.3

	case "dynamic":
		// Adjust based on demand
		demandLevel := psu.getDemandLevel(itemID)
		multiplier := 1.0
		switch demandLevel {
		case demandVeryHigh:
			multiplier = 1.25
		case demandHigh:
			multiplier = 1.15
		case demandLow:
			multiplier = 0.9
		case demandVeryLow:
			multiplier = 0.8
		}
		return marketPrice * multiplier

	default:
		return marketPrice
	}
}

func (psu *PriceSettingUIManager) generateAnalytics(itemID string) *PriceAnalytics {
	// Generate analytics from price history
	history := psu.priceHistory[itemID]

	revenueHistory := make([]float64, 0)
	profitHistory := make([]float64, 0)
	salesHistory := make([]int, 0)

	for _, point := range history {
		revenueHistory = append(revenueHistory, point.Revenue)
		profit := point.Revenue - (psu.getPurchasePrice(itemID) * float64(point.Sales))
		profitHistory = append(profitHistory, profit)
		salesHistory = append(salesHistory, point.Sales)
	}

	// Calculate optimal price (simplified)
	optimalPrice := psu.calculateOptimalPrice(itemID)

	return &PriceAnalytics{
		ItemID:          itemID,
		RevenueHistory:  revenueHistory,
		ProfitHistory:   profitHistory,
		SalesHistory:    salesHistory,
		OptimalPrice:    optimalPrice,
		PriceElasticity: psu.calculateElasticity(itemID),
		LastUpdated:     time.Now(),
	}
}

func (psu *PriceSettingUIManager) calculateOptimalPrice(itemID string) float64 {
	// Simplified optimal price calculation
	purchasePrice := psu.getPurchasePrice(itemID)
	elasticity := psu.calculateElasticity(itemID)

	// Basic formula: optimal markup = 1 / elasticity
	optimalMarkup := 1.0 / elasticity

	return purchasePrice * (1 + optimalMarkup)
}

func (psu *PriceSettingUIManager) ruleApplies(rule *PriceRule, itemID string) bool {
	if rule.AppliesTo == "all" {
		return true
	}

	// Check category rules
	if len(rule.AppliesTo) > 9 && rule.AppliesTo[:9] == "category:" {
		categoryStr := rule.AppliesTo[9:]
		itemCategory := psu.getItemCategory(itemID)
		return string(itemCategory) == categoryStr
	}

	// Check specific item rules
	if len(rule.AppliesTo) > 5 && rule.AppliesTo[:5] == "item:" {
		targetItem := rule.AppliesTo[5:]
		return itemID == targetItem
	}

	return false
}

func (psu *PriceSettingUIManager) checkCondition(rule *PriceRule, itemID string) bool {
	switch rule.Condition {
	case "item_expiring":
		// Check if item is perishable and expiring soon
		return psu.isExpiringSoon(itemID)

	case "inventory_high":
		// Check if inventory is above threshold
		quantity := psu.gameManager.inventory.ShopInventory.GetQuantity(itemID)
		return quantity > 20

	case "demand_low":
		demandLevel := psu.getDemandLevel(itemID)
		return demandLevel == demandLow || demandLevel == demandVeryLow

	case "competitor_lower":
		currentPrice := psu.getCurrentPrice(itemID)
		competitorPrice := psu.getCompetitorPrice(itemID)
		return competitorPrice < currentPrice

	default:
		return false
	}
}

func (psu *PriceSettingUIManager) applyRule(rule *PriceRule, itemID string) float64 {
	currentPrice := psu.getCurrentPrice(itemID)

	switch rule.Action {
	case "decrease_percentage":
		return currentPrice * (1 - rule.Adjustment/100)

	case "increase_percentage":
		return currentPrice * (1 + rule.Adjustment/100)

	case "match_competitor":
		return psu.getCompetitorPrice(itemID)

	case "match_market":
		return float64(psu.market.GetPrice(itemID))

	case "undercut":
		competitorPrice := psu.getCompetitorPrice(itemID)
		return competitorPrice * 0.95

	default:
		return currentPrice
	}
}

func (psu *PriceSettingUIManager) isExpiringSoon(itemID string) bool {
	// Check if item is perishable and expiring within 2 days
	return itemID == "apple" || itemID == "orange" || itemID == "banana"
}

func (psu *PriceSettingUIManager) getItemName(itemID string) string {
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

func (psu *PriceSettingUIManager) getItemCategory(itemID string) item.Category {
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
