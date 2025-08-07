package ai

import (
	"context"
	"math"
	"math/rand"
)

// DecisionType represents the type of trading decision
type DecisionType int

const (
	DecisionBuy DecisionType = iota
	DecisionSell
	DecisionHold
)

// AIDecision represents a trading decision made by AI
type AIDecision struct {
	Type       DecisionType
	ItemID     string
	Quantity   int
	Price      float64
	Confidence float64 // 0.0 to 1.0
	Reason     string
}

// AIBehavior defines the interface for AI merchant behavior
type AIBehavior interface {
	MakeDecision(ctx context.Context, merchant *AIMerchant, marketData *MarketData) *AIDecision
	EvaluateMarket(ctx context.Context, marketData *MarketData) float64
	DetermineStrategy(merchant *AIMerchant, marketCondition string) TradingStrategy
}

// StandardAIBehavior implements standard AI merchant behavior
type StandardAIBehavior struct {
	strategies map[string]TradingStrategy
}

// NewStandardAIBehavior creates a new standard AI behavior
func NewStandardAIBehavior() AIBehavior {
	return &StandardAIBehavior{
		strategies: map[string]TradingStrategy{
			"value":    NewValueTradingStrategy(),
			"momentum": NewMomentumTradingStrategy(),
			"seasonal": NewSeasonalTradingStrategy(),
		},
	}
}

// MakeDecision makes a trading decision based on market data and merchant state
func (b *StandardAIBehavior) MakeDecision(ctx context.Context, merchant *AIMerchant, marketData *MarketData) *AIDecision {
	// Evaluate market condition
	marketScore := b.EvaluateMarket(ctx, marketData)

	// Determine market condition
	var marketCondition string
	switch {
	case marketScore > 0.7:
		marketCondition = "bullish"
	case marketScore < 0.3:
		marketCondition = "bearish"
	default:
		marketCondition = "neutral"
	}

	// Select strategy based on personality and market
	strategy := b.DetermineStrategy(merchant, marketCondition)

	// Make decision using selected strategy
	decision := strategy.Evaluate(ctx, merchant, marketData)

	// Apply personality modifiers
	decision = b.applyPersonalityModifiers(decision, merchant)

	return decision
}

// EvaluateMarket evaluates overall market conditions
func (b *StandardAIBehavior) EvaluateMarket(ctx context.Context, marketData *MarketData) float64 {
	if marketData == nil || len(marketData.Items) == 0 {
		return 0.5 // Neutral if no data
	}

	totalScore := 0.0
	count := 0

	for _, item := range marketData.Items {
		// Calculate individual item market score
		priceRatio := item.CurrentPrice / item.BasePrice
		supplyDemandRatio := float64(item.Demand) / float64(item.Supply+1) // +1 to avoid division by zero

		// Combine factors
		itemScore := (priceRatio + supplyDemandRatio) / 2.0

		// Normalize to 0-1 range
		itemScore = math.Max(0, math.Min(1, itemScore))

		totalScore += itemScore
		count++
	}

	if count == 0 {
		return 0.5
	}

	return totalScore / float64(count)
}

// DetermineStrategy selects appropriate trading strategy
func (b *StandardAIBehavior) DetermineStrategy(merchant *AIMerchant, marketCondition string) TradingStrategy {
	personality := merchant.Personality()

	// Select strategy based on personality and market
	switch personality.Type() {
	case PersonalityAggressive:
		if marketCondition == "bullish" {
			return b.strategies["momentum"]
		}
		return b.strategies["value"]

	case PersonalityConservative:
		return b.strategies["value"]

	case PersonalityOpportunistic:
		if marketCondition == "bearish" {
			return b.strategies["value"]
		}
		return b.strategies["momentum"]

	default: // Balanced
		// Use seasonal strategy if available, otherwise value
		if _, exists := b.strategies["seasonal"]; exists {
			return b.strategies["seasonal"]
		}
		return b.strategies["value"]
	}
}

// applyPersonalityModifiers adjusts decision based on personality
func (b *StandardAIBehavior) applyPersonalityModifiers(decision *AIDecision, merchant *AIMerchant) *AIDecision {
	if decision == nil {
		return nil
	}

	personality := merchant.Personality()

	// Adjust quantity based on risk tolerance
	riskFactor := personality.RiskTolerance()
	decision.Quantity = int(float64(decision.Quantity) * (0.5 + riskFactor))

	// Adjust confidence based on personality
	decision.Confidence *= personality.TradingFrequency()

	// Ensure confidence stays in valid range
	decision.Confidence = math.Max(0, math.Min(1, decision.Confidence))

	return decision
}

// DecisionMaker makes complex trading decisions
type DecisionMaker struct {
	behavior AIBehavior
}

// NewDecisionMaker creates a new decision maker
func NewDecisionMaker() *DecisionMaker {
	return &DecisionMaker{
		behavior: NewStandardAIBehavior(),
	}
}

// MakeDecisions makes multiple trading decisions
func (dm *DecisionMaker) MakeDecisions(ctx context.Context, merchant *AIMerchant, marketData *MarketData, maxDecisions int) []*AIDecision {
	decisions := make([]*AIDecision, 0, maxDecisions)

	// Analyze each item in market
	for _, item := range marketData.Items {
		if len(decisions) >= maxDecisions {
			break
		}

		// Calculate profit potential
		inventoryItems := merchant.Inventory() // This would need proper inventory lookup

		// Decide whether to buy or sell
		if shouldSell(merchant, item, inventoryItems) {
			decision := &AIDecision{
				Type:       DecisionSell,
				ItemID:     item.ItemID,
				Quantity:   calculateSellQuantity(merchant, item),
				Price:      item.CurrentPrice,
				Confidence: calculateConfidence(merchant, item),
				Reason:     "High profit margin",
			}
			decisions = append(decisions, decision)
		} else if shouldBuy(merchant, item) {
			decision := &AIDecision{
				Type:       DecisionBuy,
				ItemID:     item.ItemID,
				Quantity:   calculateBuyQuantity(merchant, item),
				Price:      item.CurrentPrice,
				Confidence: calculateConfidence(merchant, item),
				Reason:     "Good value opportunity",
			}
			decisions = append(decisions, decision)
		}
	}

	// Sort decisions by confidence (highest first)
	// In a real implementation, we'd sort the slice

	return decisions
}

// Helper functions for decision making

func shouldSell(merchant *AIMerchant, item ItemData, _ interface{}) bool {
	// Check if we have the item and if price is good
	profitMargin := (item.CurrentPrice - item.BasePrice) / item.BasePrice
	targetMargin := merchant.Personality().ProfitMarginTarget()

	return profitMargin >= targetMargin
}

func shouldBuy(merchant *AIMerchant, item ItemData) bool {
	// Check if item is undervalued and we have gold
	if merchant.Gold() < int(item.CurrentPrice) {
		return false
	}

	// Check if item is below base price (potential profit)
	discount := (item.BasePrice - item.CurrentPrice) / item.BasePrice
	riskTolerance := merchant.Personality().RiskTolerance()

	// More risk-tolerant merchants buy at smaller discounts
	requiredDiscount := 0.1 * (1.0 - riskTolerance)

	return discount >= requiredDiscount
}

func calculateBuyQuantity(merchant *AIMerchant, item ItemData) int {
	// Calculate how many we can afford
	maxAfford := int(float64(merchant.Gold()) / item.CurrentPrice)

	// Apply risk factor
	riskFactor := merchant.Personality().RiskTolerance()
	quantity := int(float64(maxAfford) * riskFactor * 0.3) // Don't spend all gold

	// Ensure at least 1 if we can afford it
	if quantity == 0 && maxAfford > 0 {
		quantity = 1
	}

	return quantity
}

func calculateSellQuantity(merchant *AIMerchant, _ ItemData) int {
	// In a real implementation, this would check actual inventory
	// For now, return a reasonable amount based on personality
	baseQuantity := 5

	// Aggressive merchants sell more at once
	if merchant.Personality().Type() == PersonalityAggressive {
		baseQuantity = 10
	} else if merchant.Personality().Type() == PersonalityConservative {
		baseQuantity = 2
	}

	return baseQuantity
}

func calculateConfidence(merchant *AIMerchant, item ItemData) float64 {
	// Base confidence on volatility and price ratio
	priceRatio := item.CurrentPrice / item.BasePrice

	// Lower volatility = higher confidence
	volatilityFactor := 1.0 - item.Volatility

	// Price far from base = opportunity but also risk
	priceFactor := 1.0 - math.Abs(1.0-priceRatio)

	confidence := (volatilityFactor + priceFactor) / 2.0

	// Apply personality factor
	confidence *= merchant.Personality().TradingFrequency()

	// Add some randomness
	// #nosec G404 - This is for game mechanics, not security
	confidence += (rand.Float64() - 0.5) * 0.1

	return math.Max(0, math.Min(1, confidence))
}
