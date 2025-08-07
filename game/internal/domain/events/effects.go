package events

import (
	"context"
	"errors"
)

// Common errors
var (
	ErrEventNotFound = errors.New("event not found")
	ErrEffectFailed  = errors.New("effect application failed")
)

// PaydayEffect represents the effect of a payday event
type PaydayEffect struct {
	WageMultiplier float64
	BaseWage       int
}

// Apply implements EventEffect
func (e *PaydayEffect) Apply(ctx context.Context, data interface{}) *EffectResult {
	goldDistributed := int(float64(e.BaseWage) * e.WageMultiplier * 100) // 100 citizens

	return &EffectResult{
		Success: true,
		Changes: map[string]interface{}{
			"gold_distributed": goldDistributed,
			"citizens_paid":    100,
			"wage_per_citizen": int(float64(e.BaseWage) * e.WageMultiplier),
		},
	}
}

// MarketCrashEffect represents a market crash
type MarketCrashEffect struct {
	PriceReduction float64
}

// Apply implements EventEffect
func (e *MarketCrashEffect) Apply(ctx context.Context, data interface{}) *EffectResult {
	return &EffectResult{
		Success: true,
		Changes: map[string]interface{}{
			"prices_reduced":    true,
			"reduction_percent": e.PriceReduction * 100,
		},
	}
}

// MarketBoostEffect represents a market boost
type MarketBoostEffect struct {
	PriceMultiplier float64
	Duration        int // days
}

// Apply implements EventEffect
func (e *MarketBoostEffect) Apply(ctx context.Context, data interface{}) *EffectResult {
	return &EffectResult{
		Success: true,
		Changes: map[string]interface{}{
			"prices_boosted": true,
			"boost_percent":  (e.PriceMultiplier - 1) * 100,
			"duration_days":  e.Duration,
		},
	}
}

// ReputationEffect modifies player reputation
type ReputationEffect struct {
	Amount int
}

// Apply implements EventEffect
func (e *ReputationEffect) Apply(ctx context.Context, data interface{}) *EffectResult {
	return &EffectResult{
		Success: true,
		Changes: map[string]interface{}{
			"reputation_change": e.Amount,
		},
	}
}

// UnlockFeatureEffect unlocks a game feature
type UnlockFeatureEffect struct {
	Feature string
}

// Apply implements EventEffect
func (e *UnlockFeatureEffect) Apply(ctx context.Context, data interface{}) *EffectResult {
	return &EffectResult{
		Success: true,
		Changes: map[string]interface{}{
			"feature_unlocked": e.Feature,
		},
	}
}

// ItemSpawnEffect spawns items in the market
type ItemSpawnEffect struct {
	ItemIDs    []string
	Quantities []int
}

// Apply implements EventEffect
func (e *ItemSpawnEffect) Apply(ctx context.Context, data interface{}) *EffectResult {
	items := make(map[string]int)
	for i, itemID := range e.ItemIDs {
		if i < len(e.Quantities) {
			items[itemID] = e.Quantities[i]
		}
	}

	return &EffectResult{
		Success: true,
		Changes: map[string]interface{}{
			"items_spawned": items,
		},
	}
}

// QuestStartEffect starts a quest
type QuestStartEffect struct {
	QuestID    string
	QuestName  string
	Objectives []string
	TimeLimit  int // days
	RewardGold int
}

// Apply implements EventEffect
func (e *QuestStartEffect) Apply(ctx context.Context, data interface{}) *EffectResult {
	return &EffectResult{
		Success: true,
		Changes: map[string]interface{}{
			"quest_started": e.QuestID,
			"quest_name":    e.QuestName,
			"objectives":    e.Objectives,
			"time_limit":    e.TimeLimit,
			"reward_gold":   e.RewardGold,
		},
	}
}

// WeatherEffect changes weather conditions
type WeatherEffect struct {
	Weather  string // "sunny", "rainy", "stormy", "snowy"
	Duration int    // days
}

// Apply implements EventEffect
func (e *WeatherEffect) Apply(ctx context.Context, data interface{}) *EffectResult {
	return &EffectResult{
		Success: true,
		Changes: map[string]interface{}{
			"weather_changed": e.Weather,
			"duration_days":   e.Duration,
		},
	}
}

// TaxEffect applies taxes
type TaxEffect struct {
	TaxRate float64
}

// Apply implements EventEffect
func (e *TaxEffect) Apply(ctx context.Context, data interface{}) *EffectResult {
	return &EffectResult{
		Success: true,
		Changes: map[string]interface{}{
			"tax_applied": true,
			"tax_rate":    e.TaxRate * 100,
		},
	}
}

// CompetitorEffect introduces a new competitor
type CompetitorEffect struct {
	CompetitorID   string
	CompetitorName string
	Strength       float64 // 0.0 to 1.0
}

// Apply implements EventEffect
func (e *CompetitorEffect) Apply(ctx context.Context, data interface{}) *EffectResult {
	return &EffectResult{
		Success: true,
		Changes: map[string]interface{}{
			"competitor_added": e.CompetitorID,
			"competitor_name":  e.CompetitorName,
			"strength":         e.Strength,
		},
	}
}
