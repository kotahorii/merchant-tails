package api

import (
	"fmt"
	"sync"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/item"
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
)

// PurchaseOption represents a purchasable item with current market data
type PurchaseOption struct {
	ItemID          string        `json:"item_id"`
	Name            string        `json:"name"`
	Category        item.Category `json:"category"`
	CurrentPrice    float64       `json:"current_price"`
	MarketTrend     string        `json:"market_trend"` // "up", "down", "stable"
	PriceChange     float64       `json:"price_change"` // Percentage change
	SupplyLevel     string        `json:"supply_level"` // "scarce", "low", "normal", "high", "abundant"
	QualityLevel    int           `json:"quality_level"`
	MaxQuantity     int           `json:"max_quantity"`
	RecommendedQty  int           `json:"recommended_qty"`
	ProfitPotential float64       `json:"profit_potential"`
	RiskLevel       string        `json:"risk_level"` // "low", "medium", "high"
	Description     string        `json:"description"`
	Icon            string        `json:"icon"`
}

// PurchaseRequest represents a purchase request from the UI
type PurchaseRequest struct {
	ItemID         string  `json:"item_id"`
	Quantity       int     `json:"quantity"`
	MaxPrice       float64 `json:"max_price"`
	NegotiatePrice bool    `json:"negotiate_price"`
}

// PurchaseResult represents the result of a purchase attempt
type PurchaseResult struct {
	Success        bool     `json:"success"`
	ItemID         string   `json:"item_id"`
	Quantity       int      `json:"quantity"`
	UnitPrice      float64  `json:"unit_price"`
	TotalCost      float64  `json:"total_cost"`
	GoldRemaining  float64  `json:"gold_remaining"`
	InventorySpace int      `json:"inventory_space"`
	Message        string   `json:"message"`
	Warnings       []string `json:"warnings"`
}

// BulkPurchaseRequest represents a bulk purchase of multiple items
type BulkPurchaseRequest struct {
	Purchases         []PurchaseRequest `json:"purchases"`
	TotalBudget       float64           `json:"total_budget"`
	OptimizeForProfit bool              `json:"optimize_for_profit"`
}

// QuickBuyPreset represents a preset purchase configuration
type QuickBuyPreset struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Items       []PurchaseRequest `json:"items"`
	TotalCost   float64           `json:"total_cost"`
	IconColor   string            `json:"icon_color"`
}

// PurchaseUIManager manages the purchase UI backend
type PurchaseUIManager struct {
	gameManager  *GameManager
	marketMgr    *market.MarketManager
	priceHistory map[string][]float64
	presets      map[string]*QuickBuyPreset
	mu           sync.RWMutex
}

// NewPurchaseUIManager creates a new purchase UI manager
func NewPurchaseUIManager(gameManager *GameManager, marketMgr *market.MarketManager) *PurchaseUIManager {
	return &PurchaseUIManager{
		gameManager:  gameManager,
		marketMgr:    marketMgr,
		priceHistory: make(map[string][]float64),
		presets:      createDefaultPresets(),
	}
}

// createDefaultPresets creates default quick buy presets
func createDefaultPresets() map[string]*QuickBuyPreset {
	presets := make(map[string]*QuickBuyPreset)

	// Basic supplies preset
	presets["basic_supplies"] = &QuickBuyPreset{
		ID:          "basic_supplies",
		Name:        "Basic Supplies",
		Description: "Common items for daily trading",
		Items: []PurchaseRequest{
			{ItemID: "apple", Quantity: 10},
			{ItemID: "bread", Quantity: 5},
			{ItemID: "potion_health", Quantity: 3},
		},
		IconColor: "#4CAF50",
	}

	// Adventure gear preset
	presets["adventure_gear"] = &QuickBuyPreset{
		ID:          "adventure_gear",
		Name:        "Adventure Gear",
		Description: "Equipment for adventurers",
		Items: []PurchaseRequest{
			{ItemID: "sword_iron", Quantity: 1},
			{ItemID: "shield_wooden", Quantity: 1},
			{ItemID: "potion_health", Quantity: 5},
		},
		IconColor: "#FF9800",
	}

	// Luxury items preset
	presets["luxury_items"] = &QuickBuyPreset{
		ID:          "luxury_items",
		Name:        "Luxury Items",
		Description: "High-value goods for wealthy customers",
		Items: []PurchaseRequest{
			{ItemID: "gem_diamond", Quantity: 1},
			{ItemID: "accessory_gold_ring", Quantity: 2},
			{ItemID: "spellbook_rare", Quantity: 1},
		},
		IconColor: "#9C27B0",
	}

	return presets
}

// GetPurchaseOptions returns available items for purchase
func (pui *PurchaseUIManager) GetPurchaseOptions(category string, sortBy string) ([]*PurchaseOption, error) {
	pui.mu.RLock()
	defer pui.mu.RUnlock()

	options := make([]*PurchaseOption, 0)

	// Get all available items from market
	marketItems := pui.marketMgr.GetAvailableItems()

	for _, marketItem := range marketItems {
		// Filter by category if specified
		if category != "" && category != "all" && string(marketItem.Category) != category {
			continue
		}

		// Calculate market data
		currentPrice := pui.marketMgr.GetCurrentPrice(marketItem.ID)
		priceHistory := pui.getPriceHistory(marketItem.ID)
		trend := calculateTrend(priceHistory)
		priceChange := calculatePriceChange(priceHistory)
		supplyLevel := pui.marketMgr.GetSupplyLevel(marketItem.ID)

		// Calculate profit potential and risk
		profitPotential := calculateProfitPotential(currentPrice, priceHistory)
		riskLevel := calculateRiskLevel(marketItem.Category, priceChange, supplyLevel)

		// Calculate recommended quantity based on budget and risk
		playerGold := pui.gameManager.GetGameState().Gold
		recommendedQty := calculateRecommendedQuantity(currentPrice, playerGold, riskLevel)

		option := &PurchaseOption{
			ItemID:          marketItem.ID,
			Name:            marketItem.Name,
			Category:        marketItem.Category,
			CurrentPrice:    currentPrice,
			MarketTrend:     trend,
			PriceChange:     priceChange,
			SupplyLevel:     supplyLevel,
			QualityLevel:    marketItem.Quality,
			MaxQuantity:     marketItem.MaxSupply,
			RecommendedQty:  recommendedQty,
			ProfitPotential: profitPotential,
			RiskLevel:       riskLevel,
			Description:     marketItem.Description,
			Icon:            fmt.Sprintf("res://assets/items/%s.png", marketItem.ID),
		}

		options = append(options, option)
	}

	// Sort options based on criteria
	sortPurchaseOptions(options, sortBy)

	return options, nil
}

// ExecutePurchase executes a single purchase
func (pui *PurchaseUIManager) ExecutePurchase(request *PurchaseRequest) (*PurchaseResult, error) {
	pui.mu.Lock()
	defer pui.mu.Unlock()

	// Validate request
	if request.Quantity <= 0 {
		return &PurchaseResult{
			Success: false,
			Message: "Invalid quantity",
		}, nil
	}

	// Get current price
	currentPrice := pui.marketMgr.GetCurrentPrice(request.ItemID)

	// Apply negotiation if requested
	finalPrice := currentPrice
	if request.NegotiatePrice {
		finalPrice = pui.negotiatePrice(currentPrice, request.MaxPrice)
	}

	// Check if price is acceptable
	if request.MaxPrice > 0 && finalPrice > request.MaxPrice {
		return &PurchaseResult{
			Success: false,
			Message: fmt.Sprintf("Price too high: %.2f > %.2f", finalPrice, request.MaxPrice),
		}, nil
	}

	totalCost := finalPrice * float64(request.Quantity)

	// Check if player has enough gold
	playerGold := pui.gameManager.GetGameState().Gold
	if totalCost > playerGold {
		return &PurchaseResult{
			Success: false,
			Message: fmt.Sprintf("Insufficient gold: need %.2f, have %.2f", totalCost, playerGold),
		}, nil
	}

	// Check inventory space
	inventory := pui.gameManager.GetPlayerInventory()
	availableSpace := inventory.GetAvailableSpace()
	if request.Quantity > availableSpace {
		return &PurchaseResult{
			Success: false,
			Message: fmt.Sprintf("Insufficient inventory space: need %d, have %d", request.Quantity, availableSpace),
		}, nil
	}

	// Execute the purchase
	buyResult := pui.gameManager.BuyItem(request.ItemID, finalPrice, request.Quantity)

	if !buyResult.Success {
		return &PurchaseResult{
			Success: false,
			Message: buyResult.Message,
		}, nil
	}

	// Update price history
	pui.updatePriceHistory(request.ItemID, finalPrice)

	// Generate warnings if needed
	warnings := pui.generatePurchaseWarnings(request.ItemID, finalPrice, request.Quantity)

	return &PurchaseResult{
		Success:        true,
		ItemID:         request.ItemID,
		Quantity:       request.Quantity,
		UnitPrice:      finalPrice,
		TotalCost:      totalCost,
		GoldRemaining:  pui.gameManager.GetGameState().Gold,
		InventorySpace: inventory.GetAvailableSpace(),
		Message:        "Purchase successful",
		Warnings:       warnings,
	}, nil
}

// ExecuteBulkPurchase executes multiple purchases
func (pui *PurchaseUIManager) ExecuteBulkPurchase(request *BulkPurchaseRequest) ([]*PurchaseResult, error) {
	results := make([]*PurchaseResult, 0, len(request.Purchases))

	// Optimize purchase order if requested
	purchases := request.Purchases
	if request.OptimizeForProfit {
		purchases = pui.optimizePurchaseOrder(purchases)
	}

	totalSpent := 0.0
	for _, purchase := range purchases {
		// Check budget constraint
		if request.TotalBudget > 0 && totalSpent >= request.TotalBudget {
			break
		}

		// Adjust max price based on remaining budget
		if request.TotalBudget > 0 {
			remainingBudget := request.TotalBudget - totalSpent
			maxPossiblePrice := remainingBudget / float64(purchase.Quantity)
			if purchase.MaxPrice == 0 || purchase.MaxPrice > maxPossiblePrice {
				purchase.MaxPrice = maxPossiblePrice
			}
		}

		result, err := pui.ExecutePurchase(&purchase)
		if err != nil {
			return results, err
		}

		results = append(results, result)
		if result.Success {
			totalSpent += result.TotalCost
		}
	}

	return results, nil
}

// GetQuickBuyPresets returns available quick buy presets
func (pui *PurchaseUIManager) GetQuickBuyPresets() []*QuickBuyPreset {
	pui.mu.RLock()
	defer pui.mu.RUnlock()

	presets := make([]*QuickBuyPreset, 0, len(pui.presets))
	for _, preset := range pui.presets {
		// Update total cost
		totalCost := 0.0
		for _, item := range preset.Items {
			price := pui.marketMgr.GetCurrentPrice(item.ItemID)
			totalCost += price * float64(item.Quantity)
		}
		preset.TotalCost = totalCost
		presets = append(presets, preset)
	}

	return presets
}

// ExecuteQuickBuy executes a quick buy preset
func (pui *PurchaseUIManager) ExecuteQuickBuy(presetID string) ([]*PurchaseResult, error) {
	pui.mu.RLock()
	preset, exists := pui.presets[presetID]
	pui.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("preset not found: %s", presetID)
	}

	// Convert preset to bulk purchase request
	bulkRequest := &BulkPurchaseRequest{
		Purchases:         preset.Items,
		OptimizeForProfit: false,
	}

	return pui.ExecuteBulkPurchase(bulkRequest)
}

// SaveCustomPreset saves a custom quick buy preset
func (pui *PurchaseUIManager) SaveCustomPreset(preset *QuickBuyPreset) error {
	pui.mu.Lock()
	defer pui.mu.Unlock()

	if preset.ID == "" {
		preset.ID = fmt.Sprintf("custom_%d", time.Now().Unix())
	}

	pui.presets[preset.ID] = preset
	return nil
}

// Helper functions

func (pui *PurchaseUIManager) getPriceHistory(itemID string) []float64 {
	if history, exists := pui.priceHistory[itemID]; exists {
		return history
	}
	return []float64{pui.marketMgr.GetCurrentPrice(itemID)}
}

func (pui *PurchaseUIManager) updatePriceHistory(itemID string, price float64) {
	if _, exists := pui.priceHistory[itemID]; !exists {
		pui.priceHistory[itemID] = make([]float64, 0, 100)
	}

	pui.priceHistory[itemID] = append(pui.priceHistory[itemID], price)

	// Keep only last 100 prices
	if len(pui.priceHistory[itemID]) > 100 {
		pui.priceHistory[itemID] = pui.priceHistory[itemID][1:]
	}
}

func (pui *PurchaseUIManager) negotiatePrice(currentPrice, maxPrice float64) float64 {
	// Simple negotiation logic - can be enhanced
	negotiationPower := 0.1 // 10% negotiation power
	discount := currentPrice * negotiationPower
	finalPrice := currentPrice - discount

	if maxPrice > 0 && finalPrice > maxPrice {
		finalPrice = maxPrice
	}

	return finalPrice
}

func (pui *PurchaseUIManager) generatePurchaseWarnings(itemID string, price float64, quantity int) []string {
	warnings := make([]string, 0)

	// Check if price is above average
	history := pui.getPriceHistory(itemID)
	if len(history) > 0 {
		avg := calculateAverage(history)
		if price > avg*1.2 {
			warnings = append(warnings, fmt.Sprintf("Price is %.0f%% above average", ((price/avg)-1)*100))
		}
	}

	// Check if buying large quantity of volatile item
	marketItem := pui.marketMgr.GetItemInfo(itemID)
	if marketItem != nil && marketItem.Volatility > 0.5 && quantity > 10 {
		warnings = append(warnings, "High quantity of volatile item - increased risk")
	}

	// Check if item is perishable
	if marketItem != nil && marketItem.Category == item.CategoryFruit {
		warnings = append(warnings, "Perishable item - sell quickly to avoid losses")
	}

	return warnings
}

func (pui *PurchaseUIManager) optimizePurchaseOrder(purchases []PurchaseRequest) []PurchaseRequest {
	// Sort by profit potential (simplified)
	// In real implementation, would use more sophisticated optimization
	return purchases
}

func calculateTrend(history []float64) string {
	if len(history) < 2 {
		return "stable"
	}

	recent := history[len(history)-1]
	previous := history[len(history)-2]

	change := (recent - previous) / previous
	if change > 0.05 {
		return "up"
	} else if change < -0.05 {
		return "down"
	}
	return "stable"
}

func calculatePriceChange(history []float64) float64 {
	if len(history) < 2 {
		return 0.0
	}

	recent := history[len(history)-1]
	previous := history[len(history)-2]

	return ((recent - previous) / previous) * 100
}

func calculateProfitPotential(currentPrice float64, history []float64) float64 {
	if len(history) == 0 {
		return 0.0
	}

	// Calculate based on historical max price
	maxPrice := 0.0
	for _, price := range history {
		if price > maxPrice {
			maxPrice = price
		}
	}

	if maxPrice > currentPrice {
		return ((maxPrice - currentPrice) / currentPrice) * 100
	}
	return 0.0
}

func calculateRiskLevel(category item.Category, priceChange float64, supplyLevel string) string {
	risk := 0.0

	// Category risk
	switch category {
	case item.CategoryFruit:
		risk += 0.3 // Perishable
	case item.CategoryGem:
		risk += 0.5 // Volatile
	case item.CategoryWeapon:
		risk += 0.2 // Stable
	}

	// Price volatility risk
	if priceChange > 20 || priceChange < -20 {
		risk += 0.3
	}

	// Supply risk
	switch supplyLevel {
	case "scarce":
		risk += 0.4
	case "abundant":
		risk += 0.2
	}

	if risk > 0.6 {
		return "high"
	} else if risk > 0.3 {
		return "medium"
	}
	return "low"
}

func calculateRecommendedQuantity(price, budget float64, risk string) int {
	// Base on available budget
	maxAffordable := int(budget * 0.2 / price) // Use 20% of budget max

	// Adjust based on risk
	switch risk {
	case "high":
		maxAffordable = maxAffordable / 3
	case "medium":
		maxAffordable = maxAffordable / 2
	}

	if maxAffordable < 1 {
		maxAffordable = 1
	}

	return maxAffordable
}

func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func sortPurchaseOptions(options []*PurchaseOption, sortBy string) {
	// Implement sorting logic based on sortBy parameter
	// Options: "name", "price", "profit", "risk", "trend"
}
