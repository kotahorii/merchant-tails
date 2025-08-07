package ai

import (
	"context"
)

// Context key for season
type seasonContextKey string

const SeasonKey seasonContextKey = "season"

// TradingStrategy defines the interface for trading strategies
type TradingStrategy interface {
	Evaluate(ctx context.Context, merchant *AIMerchant, marketData *MarketData) *AIDecision
	GetName() string
}

// ValueTradingStrategy buys undervalued items and sells overvalued ones
type ValueTradingStrategy struct {
	name string
}

// NewValueTradingStrategy creates a new value trading strategy
func NewValueTradingStrategy() TradingStrategy {
	return &ValueTradingStrategy{
		name: "value",
	}
}

// GetName returns the strategy name
func (s *ValueTradingStrategy) GetName() string {
	return s.name
}

// Evaluate evaluates market for value opportunities
func (s *ValueTradingStrategy) Evaluate(ctx context.Context, merchant *AIMerchant, marketData *MarketData) *AIDecision {
	if marketData == nil || len(marketData.Items) == 0 {
		return &AIDecision{
			Type:       DecisionHold,
			Confidence: 0.5,
			Reason:     "No market data available",
		}
	}

	var bestOpportunity *AIDecision
	bestValue := 0.0

	for _, itemData := range marketData.Items {
		// Calculate value score
		valueRatio := itemData.BasePrice / itemData.CurrentPrice

		// Check for undervalued items to buy
		if valueRatio > 1.2 { // 20% undervalued
			value := (valueRatio - 1.0) * float64(itemData.Demand) / float64(itemData.Supply+1)

			if value > bestValue && merchant.Gold() >= int(itemData.CurrentPrice) {
				bestValue = value
				bestOpportunity = &AIDecision{
					Type:       DecisionBuy,
					ItemID:     itemData.ItemID,
					Quantity:   calculateValueBuyQuantity(merchant, itemData),
					Price:      itemData.CurrentPrice,
					Confidence: calculateValueConfidence(valueRatio),
					Reason:     "Undervalued item with good demand",
				}
			}
		}

		// Check for overvalued items to sell
		if valueRatio < 0.8 { // 20% overvalued
			value := (1.0 - valueRatio) * float64(itemData.Supply) / float64(itemData.Demand+1)

			if value > bestValue {
				bestValue = value
				bestOpportunity = &AIDecision{
					Type:       DecisionSell,
					ItemID:     itemData.ItemID,
					Quantity:   5, // Default sell quantity
					Price:      itemData.CurrentPrice,
					Confidence: calculateValueConfidence(1.0 / valueRatio),
					Reason:     "Overvalued item with low demand",
				}
			}
		}
	}

	if bestOpportunity == nil {
		return &AIDecision{
			Type:       DecisionHold,
			Confidence: 0.3,
			Reason:     "No valuable opportunities found",
		}
	}

	return bestOpportunity
}

func calculateValueBuyQuantity(merchant *AIMerchant, itemData ItemData) int {
	maxAfford := int(float64(merchant.Gold()) / itemData.CurrentPrice)

	// Conservative approach for value trading
	quantity := maxAfford / 4
	if quantity == 0 && maxAfford > 0 {
		quantity = 1
	}

	return quantity
}

func calculateValueConfidence(valueRatio float64) float64 {
	// Higher value ratio = higher confidence
	confidence := (valueRatio - 1.0) * 0.5
	if confidence > 1.0 {
		confidence = 1.0
	} else if confidence < 0 {
		confidence = 0
	}

	return confidence
}

// MomentumTradingStrategy follows price trends
type MomentumTradingStrategy struct {
	name string
}

// NewMomentumTradingStrategy creates a new momentum trading strategy
func NewMomentumTradingStrategy() TradingStrategy {
	return &MomentumTradingStrategy{
		name: "momentum",
	}
}

// GetName returns the strategy name
func (s *MomentumTradingStrategy) GetName() string {
	return s.name
}

// Evaluate evaluates market for momentum opportunities
func (s *MomentumTradingStrategy) Evaluate(ctx context.Context, merchant *AIMerchant, marketData *MarketData) *AIDecision {
	if marketData == nil || len(marketData.Items) == 0 {
		return &AIDecision{
			Type:       DecisionHold,
			Confidence: 0.5,
			Reason:     "No market data available",
		}
	}

	var bestOpportunity *AIDecision
	bestMomentum := 0.0

	for _, itemData := range marketData.Items {
		if len(itemData.PriceHistory) < 2 {
			continue // Need history for momentum
		}

		// Calculate momentum (price change rate)
		momentum := calculateMomentum(itemData.PriceHistory)

		// Strong upward momentum - buy
		if momentum > 0.1 { // 10% positive momentum
			if momentum > bestMomentum && merchant.Gold() >= int(itemData.CurrentPrice) {
				bestMomentum = momentum
				bestOpportunity = &AIDecision{
					Type:       DecisionBuy,
					ItemID:     itemData.ItemID,
					Quantity:   calculateMomentumQuantity(merchant, itemData, momentum),
					Price:      itemData.CurrentPrice,
					Confidence: momentum * 2, // Scale momentum to confidence
					Reason:     "Strong upward price momentum",
				}
			}
		}

		// Strong downward momentum - sell (if we have it)
		if momentum < -0.1 { // 10% negative momentum
			absMomentum := -momentum
			if absMomentum > bestMomentum {
				bestMomentum = absMomentum
				bestOpportunity = &AIDecision{
					Type:       DecisionSell,
					ItemID:     itemData.ItemID,
					Quantity:   10, // Aggressive sell on downward momentum
					Price:      itemData.CurrentPrice,
					Confidence: absMomentum * 2,
					Reason:     "Strong downward price momentum",
				}
			}
		}
	}

	if bestOpportunity == nil {
		return &AIDecision{
			Type:       DecisionHold,
			Confidence: 0.3,
			Reason:     "No strong momentum detected",
		}
	}

	return bestOpportunity
}

func calculateMomentum(priceHistory []float64) float64 {
	if len(priceHistory) < 2 {
		return 0
	}

	// Simple momentum: recent price change rate
	recentPrice := priceHistory[len(priceHistory)-1]
	previousPrice := priceHistory[len(priceHistory)-2]

	if previousPrice == 0 {
		return 0
	}

	return (recentPrice - previousPrice) / previousPrice
}

func calculateMomentumQuantity(merchant *AIMerchant, itemData ItemData, momentum float64) int {
	maxAfford := int(float64(merchant.Gold()) / itemData.CurrentPrice)

	// Aggressive approach for momentum trading
	quantity := int(float64(maxAfford) * momentum * 2)
	if quantity > maxAfford/2 {
		quantity = maxAfford / 2 // Cap at half of affordable amount
	}
	if quantity == 0 && maxAfford > 0 {
		quantity = 1
	}

	return quantity
}

// SeasonalTradingStrategy trades based on seasonal patterns
type SeasonalTradingStrategy struct {
	name string
}

// NewSeasonalTradingStrategy creates a new seasonal trading strategy
func NewSeasonalTradingStrategy() TradingStrategy {
	return &SeasonalTradingStrategy{
		name: "seasonal",
	}
}

// GetName returns the strategy name
func (s *SeasonalTradingStrategy) GetName() string {
	return s.name
}

// Evaluate evaluates market for seasonal opportunities
func (s *SeasonalTradingStrategy) Evaluate(ctx context.Context, merchant *AIMerchant, marketData *MarketData) *AIDecision {
	if marketData == nil || len(marketData.Items) == 0 {
		return &AIDecision{
			Type:       DecisionHold,
			Confidence: 0.5,
			Reason:     "No market data available",
		}
	}

	// Get current season from context
	season, ok := ctx.Value(SeasonKey).(string)
	if !ok {
		season = "spring" // Default season
	}

	var bestOpportunity *AIDecision
	bestScore := 0.0

	for _, itemData := range marketData.Items {
		// Check if item is seasonal
		seasonalScore := calculateSeasonalScore(itemData, season)

		if seasonalScore > 0.5 && seasonalScore > bestScore {
			// Seasonal item in season - buy
			if merchant.Gold() >= int(itemData.CurrentPrice) {
				bestScore = seasonalScore
				bestOpportunity = &AIDecision{
					Type:       DecisionBuy,
					ItemID:     itemData.ItemID,
					Quantity:   calculateSeasonalQuantity(merchant, itemData, seasonalScore),
					Price:      itemData.CurrentPrice,
					Confidence: seasonalScore,
					Reason:     "Seasonal item in high demand",
				}
			}
		} else if seasonalScore < -0.5 && -seasonalScore > bestScore {
			// Out of season item - sell if we have it
			bestScore = -seasonalScore
			bestOpportunity = &AIDecision{
				Type:       DecisionSell,
				ItemID:     itemData.ItemID,
				Quantity:   5,
				Price:      itemData.CurrentPrice,
				Confidence: -seasonalScore,
				Reason:     "Out of season item",
			}
		}
	}

	if bestOpportunity == nil {
		return &AIDecision{
			Type:       DecisionHold,
			Confidence: 0.3,
			Reason:     "No seasonal opportunities",
		}
	}

	return bestOpportunity
}

func calculateSeasonalScore(itemData ItemData, season string) float64 {
	// Check item tags for seasonal indicators
	for _, tag := range itemData.Tags {
		if tag == season {
			return 0.8 // High score for in-season items
		}

		// Check for opposite season
		oppositeSeasons := map[string]string{
			"summer": "winter",
			"winter": "summer",
			"spring": "autumn",
			"autumn": "spring",
		}

		if opposite, exists := oppositeSeasons[season]; exists && tag == opposite {
			return -0.8 // Negative score for out-of-season items
		}
	}

	// Check category-based seasonal patterns
	switch itemData.Category {
	case 1: // CategoryPotion
		// Potions are more valuable in winter (healing) and summer (cooling)
		if season == "winter" || season == "summer" {
			return 0.4
		}
	case 2: // CategoryFood
		// Food items vary by season
		if season == "autumn" {
			return 0.6 // Harvest season
		}
	}

	return 0 // Neutral
}

func calculateSeasonalQuantity(merchant *AIMerchant, itemData ItemData, seasonalScore float64) int {
	maxAfford := int(float64(merchant.Gold()) / itemData.CurrentPrice)

	// Moderate approach for seasonal trading
	quantity := int(float64(maxAfford) * seasonalScore * 0.3)
	if quantity == 0 && maxAfford > 0 {
		quantity = 1
	}

	return quantity
}
