package market

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/item"
)

// DemandLevel represents the current demand level in the market
type DemandLevel int

const (
	DemandVeryLow DemandLevel = iota
	DemandLow
	DemandNormal
	DemandHigh
	DemandVeryHigh
)

// SupplyLevel represents the current supply level in the market
type SupplyLevel int

const (
	SupplyVeryLow SupplyLevel = iota
	SupplyLow
	SupplyNormal
	SupplyHigh
	SupplyVeryHigh
)

// EventType represents different types of market events
type EventType int

const (
	EventNormal EventType = iota
	EventDragonAttack
	EventHarvestFestival
	EventMarketCrash
	EventMarketBoom
)

// EffectType represents different types of event effects
type EffectType int

const (
	EffectNone EffectType = iota
	EffectSupplyDecrease
	EffectSupplyIncrease
	EffectDemandIncrease
	EffectDemandDecrease
	EffectPriceModifier
)

// PriceTrend represents the price movement trend
type PriceTrend int

const (
	TrendStable PriceTrend = iota
	TrendUp
	TrendDown
)

// TradeAction represents recommended trading actions
type TradeAction int

const (
	ActionHold TradeAction = iota
	ActionBuy
	ActionSell
)

// Market represents the game's market system
type Market struct {
	PricingEngine *PricingEngine
	State         *MarketState
	Prices        map[string]*PriceHistory
	ActiveEvents  []*MarketEvent
	items         map[string]*item.Item
	mu            sync.RWMutex
}

// MarketState represents the current state of the market
type MarketState struct {
	CurrentDemand DemandLevel
	CurrentSupply SupplyLevel
	CurrentSeason item.Season
	CurrentDay    int
}

// PricingEngine calculates prices based on various factors
type PricingEngine struct {
	baseFormula    PriceFormula
	modifiers      []PriceModifier
	volatilityCalc VolatilityCalculator
	random         *rand.Rand
}

// PriceFormula interface for price calculation
type PriceFormula interface {
	Calculate(item *item.Item, state *MarketState) float64
}

// PriceModifier interface for price modifiers
type PriceModifier interface {
	Apply(basePrice float64, item *item.Item, state *MarketState) float64
}

// VolatilityCalculator interface for volatility calculation
type VolatilityCalculator interface {
	Calculate(item *item.Item) float64
}

// MarketEvent represents a market event
type MarketEvent struct {
	Type        EventType
	Name        string
	Description string
	Duration    int
	StartDay    int
	Effects     []EventEffect
	IsActive    bool
}

// EventEffect represents an effect of a market event
type EventEffect struct {
	Type  EffectType
	Value float64
}

// PriceHistory tracks historical prices for an item
type PriceHistory struct {
	Records      []PriceRecord
	CurrentPrice int
	AveragePrice int
	Trend        PriceTrend
	MaxSize      int
	mu           sync.RWMutex
}

// PriceRecord represents a single price point
type PriceRecord struct {
	Price     int
	Timestamp time.Time
}

// NewMarket creates a new market instance
func NewMarket() *Market {
	return &Market{
		PricingEngine: NewPricingEngine(),
		State: &MarketState{
			CurrentDemand: DemandNormal,
			CurrentSupply: SupplyNormal,
			CurrentSeason: item.SeasonSpring,
			CurrentDay:    1,
		},
		Prices:       make(map[string]*PriceHistory),
		ActiveEvents: make([]*MarketEvent, 0),
		items:        make(map[string]*item.Item),
	}
}

// NewPricingEngine creates a new pricing engine
func NewPricingEngine() *PricingEngine {
	return &PricingEngine{
		baseFormula:    &DefaultPriceFormula{},
		modifiers:      []PriceModifier{},
		volatilityCalc: &DefaultVolatilityCalculator{},
		random:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// RegisterItem registers an item in the market
func (m *Market) RegisterItem(item *item.Item) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items[item.ID] = item
	m.Prices[item.ID] = &PriceHistory{
		Records:      make([]PriceRecord, 0),
		CurrentPrice: item.BasePrice,
		AveragePrice: item.BasePrice,
		Trend:        TrendStable,
		MaxSize:      10,
	}
}

// UpdatePrices updates all item prices based on current market conditions
func (m *Market) UpdatePrices() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, item := range m.items {
		newPrice := m.PricingEngine.CalculatePrice(item, m.State)
		history := m.Prices[id]

		// Add to history
		history.AddRecord(newPrice, time.Now())
		history.updateTrend()
	}
}

// GetPriceHistory returns the price history for an item
func (m *Market) GetPriceHistory(itemID string) *PriceHistory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.Prices[itemID]
}

// ApplyEvent applies a market event
func (m *Market) ApplyEvent(event *MarketEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, effect := range event.Effects {
		switch effect.Type {
		case EffectSupplyDecrease:
			m.State.CurrentSupply = m.adjustSupply(m.State.CurrentSupply, -int(effect.Value))
		case EffectSupplyIncrease:
			m.State.CurrentSupply = m.adjustSupply(m.State.CurrentSupply, int(effect.Value))
		case EffectDemandIncrease:
			m.State.CurrentDemand = m.adjustDemand(m.State.CurrentDemand, int(effect.Value))
		case EffectDemandDecrease:
			m.State.CurrentDemand = m.adjustDemand(m.State.CurrentDemand, -int(effect.Value))
		}
	}

	event.IsActive = true
	m.ActiveEvents = append(m.ActiveEvents, event)
}

// GetRecommendedAction returns a recommended trading action for an item
func (m *Market) GetRecommendedAction(itemID string) TradeAction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history := m.Prices[itemID]
	if history == nil {
		return ActionHold
	}

	priceRatio := float64(history.CurrentPrice) / float64(history.AveragePrice)

	switch history.Trend {
	case TrendUp:
		if priceRatio < 0.9 {
			return ActionBuy
		}
	case TrendDown:
		if priceRatio > 1.1 {
			return ActionSell
		}
	}

	return ActionHold
}

// GetDemandModifier returns the price modifier for the current demand level
func (s *MarketState) GetDemandModifier() float64 {
	modifiers := map[DemandLevel]float64{
		DemandVeryLow:  0.7,
		DemandLow:      0.85,
		DemandNormal:   1.0,
		DemandHigh:     1.2,
		DemandVeryHigh: 1.5,
	}
	return modifiers[s.CurrentDemand]
}

// GetSupplyModifier returns the price modifier for the current supply level
func (s *MarketState) GetSupplyModifier() float64 {
	modifiers := map[SupplyLevel]float64{
		SupplyVeryLow:  1.3,
		SupplyLow:      1.15,
		SupplyNormal:   1.0,
		SupplyHigh:     0.85,
		SupplyVeryHigh: 0.7,
	}
	return modifiers[s.CurrentSupply]
}

// CalculatePrice calculates the price for an item
func (pe *PricingEngine) CalculatePrice(item *item.Item, state *MarketState) int {
	basePrice := float64(item.BasePrice)

	// Apply demand and supply modifiers
	demandMod := state.GetDemandModifier()
	supplyMod := state.GetSupplyModifier()

	// Apply seasonal modifier
	seasonMod := pe.getSeasonalModifier(item, state.CurrentSeason)

	// Calculate base price with modifiers
	price := basePrice * demandMod * supplyMod * seasonMod

	// Apply volatility (reduced for more predictable pricing)
	volatility := item.GetVolatility()
	randomFactor := 1.0 + (pe.random.Float64()-0.5)*float64(volatility)*0.2
	price *= randomFactor

	// Ensure price doesn't go below 50% or above 200% of base
	minPrice := basePrice * 0.5
	maxPrice := basePrice * 2.0

	if price < minPrice {
		price = minPrice
	} else if price > maxPrice {
		price = maxPrice
	}

	return int(math.Round(price))
}

// getSeasonalModifier returns the seasonal price modifier for an item
func (pe *PricingEngine) getSeasonalModifier(i *item.Item, season item.Season) float64 {
	// Simple seasonal effects for different item categories
	switch i.Category {
	case item.CategoryFruit:
		switch season {
		case item.SeasonSpring:
			return 1.1
		case item.SeasonSummer:
			return 1.0
		case item.SeasonAutumn:
			return 1.3
		case item.SeasonWinter:
			return 0.8
		}
	case item.CategoryPotion:
		// Potions are more needed in winter (cold season)
		if season == item.SeasonWinter {
			return 1.2
		}
	}
	return 1.0
}

// AddRecord adds a new price record to the history
func (ph *PriceHistory) AddRecord(price int, timestamp time.Time) {
	ph.mu.Lock()
	defer ph.mu.Unlock()

	ph.Records = append(ph.Records, PriceRecord{
		Price:     price,
		Timestamp: timestamp,
	})

	// Maintain max size
	if len(ph.Records) > ph.MaxSize {
		ph.Records = ph.Records[len(ph.Records)-ph.MaxSize:]
	}

	ph.CurrentPrice = price
	ph.updateAverage()
}

// GetTrend returns the current price trend
func (ph *PriceHistory) GetTrend() PriceTrend {
	ph.mu.RLock()
	defer ph.mu.RUnlock()
	return ph.Trend
}

// updateAverage updates the average price
func (ph *PriceHistory) updateAverage() {
	if len(ph.Records) == 0 {
		return
	}

	sum := 0
	for _, record := range ph.Records {
		sum += record.Price
	}
	ph.AveragePrice = sum / len(ph.Records)
}

// updateTrend updates the price trend based on recent history
func (ph *PriceHistory) updateTrend() {
	if len(ph.Records) < 2 {
		ph.Trend = TrendStable
		return
	}

	// Compare recent prices with older prices
	recentAvg := 0
	olderAvg := 0
	halfPoint := len(ph.Records) / 2

	for i := 0; i < halfPoint; i++ {
		olderAvg += ph.Records[i].Price
	}
	olderAvg /= halfPoint

	for i := halfPoint; i < len(ph.Records); i++ {
		recentAvg += ph.Records[i].Price
	}
	recentAvg /= (len(ph.Records) - halfPoint)

	threshold := float64(olderAvg) * 0.05 // 5% threshold for trend detection

	if float64(recentAvg) > float64(olderAvg)+threshold {
		ph.Trend = TrendUp
	} else if float64(recentAvg) < float64(olderAvg)-threshold {
		ph.Trend = TrendDown
	} else {
		ph.Trend = TrendStable
	}
}

// adjustDemand adjusts the demand level by the given steps
func (m *Market) adjustDemand(current DemandLevel, steps int) DemandLevel {
	newLevel := int(current) + steps
	if newLevel < int(DemandVeryLow) {
		return DemandVeryLow
	}
	if newLevel > int(DemandVeryHigh) {
		return DemandVeryHigh
	}
	return DemandLevel(newLevel)
}

// adjustSupply adjusts the supply level by the given steps
func (m *Market) adjustSupply(current SupplyLevel, steps int) SupplyLevel {
	newLevel := int(current) + steps
	if newLevel < int(SupplyVeryLow) {
		return SupplyVeryLow
	}
	if newLevel > int(SupplyVeryHigh) {
		return SupplyVeryHigh
	}
	return SupplyLevel(newLevel)
}

// DefaultPriceFormula is the default implementation of PriceFormula
type DefaultPriceFormula struct{}

// Calculate implements the PriceFormula interface
func (f *DefaultPriceFormula) Calculate(item *item.Item, state *MarketState) float64 {
	return float64(item.BasePrice) * state.GetDemandModifier() * state.GetSupplyModifier()
}

// DefaultVolatilityCalculator is the default implementation of VolatilityCalculator
type DefaultVolatilityCalculator struct{}

// Calculate implements the VolatilityCalculator interface
func (v *DefaultVolatilityCalculator) Calculate(item *item.Item) float64 {
	return float64(item.GetVolatility())
}
