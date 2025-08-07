package event

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
)

func TestEventBus_Integration(t *testing.T) {
	// Reset global event bus for clean test
	ResetGlobalEventBus()

	t.Run("ItemRegisteredEvent", func(t *testing.T) {
		var called bool
		Subscribe(EventNameItemRegistered, func(e Event) error {
			ire, ok := e.(*ItemRegisteredEvent)
			require.True(t, ok)
			assert.Equal(t, "ITEM001", ire.ItemID)
			assert.Equal(t, "Apple", ire.ItemName)
			assert.Equal(t, "Fruit", ire.Category)
			assert.Equal(t, 100, ire.Price)
			called = true
			return nil
		})

		event := NewItemRegisteredEvent("ITEM001", "Apple", "Fruit", 100)
		err := Publish(event)
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("PriceUpdatedEvent", func(t *testing.T) {
		var called bool
		Subscribe(EventNamePriceUpdated, func(e Event) error {
			pue, ok := e.(*PriceUpdatedEvent)
			require.True(t, ok)
			assert.Equal(t, "ITEM001", pue.ItemID)
			assert.Equal(t, 100, pue.OldPrice)
			assert.Equal(t, 120, pue.NewPrice)
			assert.Equal(t, "Market demand increased", pue.Reason)
			called = true
			return nil
		})

		event := NewPriceUpdatedEvent("ITEM001", 100, 120, "Market demand increased")
		err := Publish(event)
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("TransactionCompleteEvent", func(t *testing.T) {
		var called bool
		Subscribe(EventNameTransactionComplete, func(e Event) error {
			tce, ok := e.(*TransactionCompleteEvent)
			require.True(t, ok)
			assert.Equal(t, "TX001", tce.TransactionID)
			assert.Equal(t, "buy", tce.Type)
			assert.Equal(t, "ITEM001", tce.ItemID)
			assert.Equal(t, 5, tce.Quantity)
			assert.Equal(t, 500, tce.TotalPrice)
			assert.Equal(t, "MERCHANT001", tce.PartyID)
			called = true
			return nil
		})

		event := NewTransactionCompleteEvent("TX001", "buy", "ITEM001", 5, 500, "MERCHANT001")
		err := Publish(event)
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("MarketEventAdapter", func(t *testing.T) {
		var called bool
		Subscribe("market.dragon_attack", func(e Event) error {
			mea, ok := e.(*MarketEventAdapter)
			require.True(t, ok)
			assert.Equal(t, market.EventDragonAttack, mea.MarketEvent.Type)
			assert.Equal(t, "Dragon Attack", mea.MarketEvent.Name)
			called = true
			return nil
		})

		marketEvent := &market.MarketEvent{
			Type:        market.EventDragonAttack,
			Name:        "Dragon Attack",
			Description: "A dragon attacks the trade routes",
			Duration:    3,
			Effects: []market.EventEffect{
				{Type: market.EffectSupplyDecrease, Value: 2},
				{Type: market.EffectDemandIncrease, Value: 1},
			},
		}

		err := PublishMarketEvent(marketEvent)
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("MultipleSubscribers", func(t *testing.T) {
		var count int32

		// Subscribe multiple handlers to the same event
		Subscribe(EventNameDayEnded, func(e Event) error {
			atomic.AddInt32(&count, 1)
			return nil
		})

		Subscribe(EventNameDayEnded, func(e Event) error {
			dee, ok := e.(*DayEndedEvent)
			require.True(t, ok)
			assert.Equal(t, 10, dee.DayNumber)
			atomic.AddInt32(&count, 10)
			return nil
		})

		Subscribe(EventNameDayEnded, func(e Event) error {
			atomic.AddInt32(&count, 100)
			return nil
		})

		event := NewDayEndedEvent(10, 5000, 3000, 2000)
		err := Publish(event)
		require.NoError(t, err)
		assert.Equal(t, int32(111), atomic.LoadInt32(&count))
	})

	t.Run("AsyncPublishing", func(t *testing.T) {
		done := make(chan bool, 1)

		Subscribe(EventNameSeasonChanged, func(e Event) error {
			sce, ok := e.(*SeasonChangedEvent)
			require.True(t, ok)
			assert.Equal(t, "Summer", sce.OldSeason)
			assert.Equal(t, "Autumn", sce.NewSeason)
			done <- true
			return nil
		})

		effects := map[string]float64{
			"fruit_price_modifier": 1.2,
			"harvest_bonus":        1.5,
		}
		event := NewSeasonChangedEvent("Summer", "Autumn", effects)
		PublishAsync(event)

		select {
		case <-done:
			// Success
		case <-time.After(1 * time.Second):
			t.Fatal("Async event not received within timeout")
		}
	})
}

func TestEventBus_CleanupBetweenTests(t *testing.T) {
	// Test that ResetGlobalEventBus properly cleans up
	ResetGlobalEventBus()

	var count int
	Subscribe("test.event", func(e Event) error {
		count++
		return nil
	})

	event := &TestEvent{Name: "test.event", Timestamp: time.Now().Unix()}
	_ = Publish(event)
	assert.Equal(t, 1, count)

	// Reset and verify handlers are cleared
	ResetGlobalEventBus()
	_ = Publish(event)
	assert.Equal(t, 1, count) // Count should not increase
}
