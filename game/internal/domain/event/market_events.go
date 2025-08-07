package event

import (
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
)

// MarketEventAdapter adapts market.MarketEvent to the Event interface
type MarketEventAdapter struct {
	*BaseEvent
	MarketEvent *market.MarketEvent
}

// NewMarketEventAdapter creates a new adapter for market events
func NewMarketEventAdapter(me *market.MarketEvent) *MarketEventAdapter {
	eventName := EventNameMarketEventOccurred
	if me != nil {
		switch me.Type {
		case market.EventDragonAttack:
			eventName = "market.dragon_attack"
		case market.EventHarvestFestival:
			eventName = "market.harvest_festival"
		case market.EventMarketCrash:
			eventName = "market.crash"
		case market.EventMarketBoom:
			eventName = "market.boom"
		}
	}

	return &MarketEventAdapter{
		BaseEvent:   NewBaseEvent(eventName),
		MarketEvent: me,
	}
}

// PublishMarketEvent publishes a market event through the event bus
func PublishMarketEvent(me *market.MarketEvent) error {
	adapter := NewMarketEventAdapter(me)
	return Publish(adapter)
}

// PublishMarketEventAsync publishes a market event asynchronously
func PublishMarketEventAsync(me *market.MarketEvent) {
	adapter := NewMarketEventAdapter(me)
	PublishAsync(adapter)
}
