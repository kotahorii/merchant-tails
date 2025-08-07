package ai

import (
	"math"
)

// MarketInfluence represents a merchant's influence on the market
type MarketInfluence struct {
	PriceImpact  float64 // How much this merchant affects prices
	DemandImpact float64 // How much this merchant affects demand
	SupplyImpact float64 // How much this merchant affects supply
	Reputation   float64 // Overall market reputation
}

// MarketInfluenceCalculator calculates merchant influence on market
type MarketInfluenceCalculator struct {
	baseInfluence float64
}

// NewMarketInfluenceCalculator creates a new influence calculator
func NewMarketInfluenceCalculator() *MarketInfluenceCalculator {
	return &MarketInfluenceCalculator{
		baseInfluence: 0.01, // 1% base influence
	}
}

// CalculateInfluence calculates a merchant's market influence
func (c *MarketInfluenceCalculator) CalculateInfluence(merchant *AIMerchant) *MarketInfluence {
	// Base factors
	goldFactor := c.calculateGoldFactor(merchant.Gold())
	reputationFactor := c.calculateReputationFactor(merchant.Reputation())
	personalityFactor := c.calculatePersonalityFactor(merchant.Personality())

	// Calculate specific impacts
	priceImpact := c.baseInfluence * goldFactor * personalityFactor
	demandImpact := c.baseInfluence * reputationFactor * personalityFactor
	supplyImpact := c.baseInfluence * goldFactor * 0.5 // Less impact on supply

	return &MarketInfluence{
		PriceImpact:  math.Min(priceImpact, 0.2),   // Cap at 20% impact
		DemandImpact: math.Min(demandImpact, 0.15), // Cap at 15% impact
		SupplyImpact: math.Min(supplyImpact, 0.1),  // Cap at 10% impact
		Reputation:   merchant.Reputation(),
	}
}

// calculateGoldFactor returns influence factor based on gold
func (c *MarketInfluenceCalculator) calculateGoldFactor(gold int) float64 {
	// Logarithmic scale - more gold = more influence, but diminishing returns
	if gold <= 0 {
		return 0
	}

	// Every 1000 gold doubles influence, up to a point
	factor := math.Log10(float64(gold)/100.0 + 1)
	return math.Min(factor, 3.0) // Cap at 3x multiplier
}

// calculateReputationFactor returns influence factor based on reputation
func (c *MarketInfluenceCalculator) calculateReputationFactor(reputation float64) float64 {
	// Reputation ranges from -100 to 100
	// Convert to 0-2 scale (0 at -100, 1 at 0, 2 at 100)
	return (reputation + 100) / 100.0
}

// calculatePersonalityFactor returns influence factor based on personality
func (c *MarketInfluenceCalculator) calculatePersonalityFactor(personality MerchantPersonality) float64 {
	// Different personalities have different market impact
	competitiveness := personality.CompetitivenessFactor()
	frequency := personality.TradingFrequency()

	// Average of competitiveness and trading frequency
	return (competitiveness + frequency) / 2.0
}

// AggregateInfluence aggregates influence from multiple merchants
func (c *MarketInfluenceCalculator) AggregateInfluence(influences []*MarketInfluence) *MarketInfluence {
	if len(influences) == 0 {
		return &MarketInfluence{}
	}

	totalPrice := 0.0
	totalDemand := 0.0
	totalSupply := 0.0
	totalReputation := 0.0

	for _, influence := range influences {
		// Simple additive aggregation with diminishing returns
		totalPrice += influence.PriceImpact * (1.0 - totalPrice*0.5)
		totalDemand += influence.DemandImpact * (1.0 - totalDemand*0.5)
		totalSupply += influence.SupplyImpact * (1.0 - totalSupply*0.5)
		totalReputation += influence.Reputation
	}

	// Average reputation
	if len(influences) > 0 {
		totalReputation /= float64(len(influences))
	}

	return &MarketInfluence{
		PriceImpact:  math.Min(totalPrice, 0.5),  // Cap total at 50%
		DemandImpact: math.Min(totalDemand, 0.4), // Cap total at 40%
		SupplyImpact: math.Min(totalSupply, 0.3), // Cap total at 30%
		Reputation:   totalReputation,
	}
}

// CalculatePriceEffect calculates the price effect of an influence
func (i *MarketInfluence) CalculatePriceEffect(basePrice float64) float64 {
	// Apply influence to base price
	// Positive impact increases price, negative decreases
	return basePrice * (1.0 + i.PriceImpact)
}

// CalculateDemandEffect calculates the demand effect of an influence
func (i *MarketInfluence) CalculateDemandEffect(baseDemand int) int {
	// Apply influence to base demand
	effect := float64(baseDemand) * (1.0 + i.DemandImpact)
	return int(math.Round(effect))
}

// CalculateSupplyEffect calculates the supply effect of an influence
func (i *MarketInfluence) CalculateSupplyEffect(baseSupply int) int {
	// Apply influence to base supply
	effect := float64(baseSupply) * (1.0 + i.SupplyImpact)
	return int(math.Round(effect))
}
