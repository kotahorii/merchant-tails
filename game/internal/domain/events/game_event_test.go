package events

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gametime "github.com/yourusername/merchant-tails/game/internal/domain/time"
)

func TestNewGameEvent(t *testing.T) {
	event := NewGameEvent(
		"payday",
		"Payday",
		"Citizens receive their monthly wages",
		EventTypeRegular,
		EventPriorityNormal,
	)

	assert.NotNil(t, event)
	assert.Equal(t, "payday", event.ID)
	assert.Equal(t, "Payday", event.Name)
	assert.Equal(t, EventTypeRegular, event.Type)
	assert.Equal(t, EventPriorityNormal, event.Priority)
	assert.True(t, event.IsActive)
}

func TestEventSchedule(t *testing.T) {
	t.Run("regular monthly event", func(t *testing.T) {
		schedule := &EventSchedule{
			Type:      ScheduleTypeMonthly,
			DayOfWeek: 15, // 15th of each month
		}

		currentTime := gametime.GameTime{
			Year:   1,
			Season: gametime.Spring,
			Day:    14,
		}

		// Should trigger on day 15
		currentTime.Day = 15
		assert.True(t, schedule.ShouldTrigger(currentTime))

		// Should not trigger on other days
		currentTime.Day = 16
		assert.False(t, schedule.ShouldTrigger(currentTime))
	})

	t.Run("seasonal event", func(t *testing.T) {
		schedule := &EventSchedule{
			Type:      ScheduleTypeSeasonal,
			Season:    gametime.Summer,
			DayOfWeek: 1, // First day of summer
		}

		currentTime := gametime.GameTime{
			Year:   1,
			Season: gametime.Summer,
			Day:    1,
		}

		// Should trigger on first day of summer
		assert.True(t, schedule.ShouldTrigger(currentTime))

		// Should not trigger on other days or seasons
		currentTime.Day = 2
		assert.False(t, schedule.ShouldTrigger(currentTime))

		currentTime.Season = gametime.Spring
		currentTime.Day = 1
		assert.False(t, schedule.ShouldTrigger(currentTime))
	})

	t.Run("one-time event", func(t *testing.T) {
		triggerTime := gametime.GameTime{
			Year:   2,
			Season: gametime.Winter,
			Day:    25,
		}

		schedule := &EventSchedule{
			Type:        ScheduleTypeOneTime,
			TriggerTime: &triggerTime,
		}

		// Should trigger at exact time
		assert.True(t, schedule.ShouldTrigger(triggerTime))

		// Should not trigger at other times
		otherTime := gametime.GameTime{
			Year:   2,
			Season: gametime.Winter,
			Day:    24,
		}
		assert.False(t, schedule.ShouldTrigger(otherTime))
	})
}

func TestEventManager(t *testing.T) {
	manager := NewEventManager()
	assert.NotNil(t, manager)

	// Register events
	paydayEvent := NewGameEvent(
		"payday",
		"Payday",
		"Monthly wages",
		EventTypeRegular,
		EventPriorityNormal,
	)
	paydayEvent.Schedule = &EventSchedule{
		Type:      ScheduleTypeMonthly,
		DayOfWeek: 15,
	}

	harvestEvent := NewGameEvent(
		"harvest",
		"Harvest Festival",
		"Annual harvest celebration",
		EventTypeSeasonal,
		EventPriorityHigh,
	)
	harvestEvent.Schedule = &EventSchedule{
		Type:      ScheduleTypeSeasonal,
		Season:    gametime.Autumn,
		DayOfWeek: 15,
	}

	manager.RegisterEvent(paydayEvent)
	manager.RegisterEvent(harvestEvent)

	// Check registered events
	assert.Equal(t, 2, len(manager.GetAllEvents()))
	assert.NotNil(t, manager.GetEvent("payday"))
	assert.NotNil(t, manager.GetEvent("harvest"))
}

func TestEventManagerUpdate(t *testing.T) {
	manager := NewEventManager()

	triggeredEvents := []string{}
	manager.RegisterEventHandler(func(event *GameEvent) {
		triggeredEvents = append(triggeredEvents, event.ID)
	})

	// Create a monthly event
	paydayEvent := NewGameEvent(
		"payday",
		"Payday",
		"Monthly wages",
		EventTypeRegular,
		EventPriorityNormal,
	)
	paydayEvent.Schedule = &EventSchedule{
		Type:      ScheduleTypeMonthly,
		DayOfWeek: 15,
	}
	manager.RegisterEvent(paydayEvent)

	// Update with time that should trigger
	gameTime := gametime.GameTime{
		Year:   1,
		Season: gametime.Spring,
		Day:    15,
	}

	ctx := context.Background()
	manager.Update(ctx, gameTime)

	assert.Contains(t, triggeredEvents, "payday")
}

func TestEventEffects(t *testing.T) {
	// Test payday effect
	paydayEffect := &PaydayEffect{
		WageMultiplier: 1.0,
		BaseWage:       100,
	}

	ctx := context.Background()
	result := paydayEffect.Apply(ctx, nil)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Contains(t, result.Changes, "gold_distributed")

	// Test market crash effect
	crashEffect := &MarketCrashEffect{
		PriceReduction: 0.3, // 30% price drop
	}

	result = crashEffect.Apply(ctx, nil)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Contains(t, result.Changes, "prices_reduced")
}

func TestEventPriority(t *testing.T) {
	manager := NewEventManager()

	executionOrder := []string{}
	manager.RegisterEventHandler(func(event *GameEvent) {
		executionOrder = append(executionOrder, event.ID)
	})

	// Register events with different priorities
	highPriorityEvent := NewGameEvent("high", "High Priority", "", EventTypeMajor, EventPriorityHigh)
	highPriorityEvent.Schedule = &EventSchedule{Type: ScheduleTypeMonthly, DayOfWeek: 1}

	normalPriorityEvent := NewGameEvent("normal", "Normal Priority", "", EventTypeRegular, EventPriorityNormal)
	normalPriorityEvent.Schedule = &EventSchedule{Type: ScheduleTypeMonthly, DayOfWeek: 1}

	lowPriorityEvent := NewGameEvent("low", "Low Priority", "", EventTypeRegular, EventPriorityLow)
	lowPriorityEvent.Schedule = &EventSchedule{Type: ScheduleTypeMonthly, DayOfWeek: 1}

	// Register in random order
	manager.RegisterEvent(normalPriorityEvent)
	manager.RegisterEvent(lowPriorityEvent)
	manager.RegisterEvent(highPriorityEvent)

	// Update with time that triggers all events
	gameTime := gametime.GameTime{Year: 1, Season: gametime.Spring, Day: 1}
	ctx := context.Background()
	manager.Update(ctx, gameTime)

	// Check execution order (high priority first)
	require.Equal(t, 3, len(executionOrder))
	assert.Equal(t, "high", executionOrder[0])
	assert.Equal(t, "normal", executionOrder[1])
	assert.Equal(t, "low", executionOrder[2])
}

func TestEventConditions(t *testing.T) {
	event := NewGameEvent(
		"dragon_attack",
		"Dragon Attack",
		"A dragon attacks the town",
		EventTypeMajor,
		EventPriorityUrgent,
	)

	// Add condition: only trigger if player rank is Expert or higher
	event.Conditions = []EventCondition{
		&RankCondition{MinRank: "Expert"},
	}

	// Test with insufficient rank
	ctx := context.WithValue(context.Background(), PlayerRankKey, "Apprentice")
	assert.False(t, event.CheckConditions(ctx))

	// Test with sufficient rank
	ctx = context.WithValue(context.Background(), PlayerRankKey, "Expert")
	assert.True(t, event.CheckConditions(ctx))
}

func TestEventChain(t *testing.T) {
	manager := NewEventManager()

	// Create main event that triggers follow-up events
	dragonEvent := NewGameEvent(
		"dragon_attack",
		"Dragon Attack",
		"Dragon attacks the town",
		EventTypeMajor,
		EventPriorityUrgent,
	)

	// Add follow-up events
	dragonEvent.FollowUpEvents = []string{"rebuild_town", "hero_celebration"}

	rebuildEvent := NewGameEvent(
		"rebuild_town",
		"Rebuild Town",
		"Town rebuilding after dragon attack",
		EventTypeRegular,
		EventPriorityNormal,
	)

	celebrationEvent := NewGameEvent(
		"hero_celebration",
		"Hero Celebration",
		"Celebrating the dragon's defeat",
		EventTypeRegular,
		EventPriorityNormal,
	)

	manager.RegisterEvent(dragonEvent)
	manager.RegisterEvent(rebuildEvent)
	manager.RegisterEvent(celebrationEvent)

	// Track triggered events
	var triggeredEvents []string
	manager.RegisterEventHandler(func(event *GameEvent) {
		triggeredEvents = append(triggeredEvents, event.ID)
	})

	// Trigger dragon event (which should chain the others)
	ctx := context.Background()
	manager.TriggerEvent(ctx, "dragon_attack")

	// Check that all events in chain were triggered
	assert.Contains(t, triggeredEvents, "dragon_attack")
	assert.Contains(t, triggeredEvents, "rebuild_town")
	assert.Contains(t, triggeredEvents, "hero_celebration")
}

func TestEventNotification(t *testing.T) {
	manager := NewEventManager()

	// Create event with advance notification
	festivalEvent := NewGameEvent(
		"harvest_festival",
		"Harvest Festival",
		"Annual harvest celebration",
		EventTypeSeasonal,
		EventPriorityNormal,
	)

	festivalEvent.Schedule = &EventSchedule{
		Type:      ScheduleTypeSeasonal,
		Season:    gametime.Autumn,
		DayOfWeek: 15,
	}
	festivalEvent.NotificationDays = 3 // Notify 3 days in advance

	manager.RegisterEvent(festivalEvent)

	// Check for upcoming events
	currentTime := gametime.GameTime{
		Year:   1,
		Season: gametime.Autumn,
		Day:    12, // 3 days before event
	}

	upcomingEvents := manager.GetUpcomingEvents(currentTime, 5)
	assert.Contains(t, upcomingEvents, festivalEvent)

	// Check notification
	notifications := manager.GetEventNotifications(currentTime)
	assert.Equal(t, 1, len(notifications))
	assert.Equal(t, "harvest_festival", notifications[0].EventID)
	assert.Equal(t, 3, notifications[0].DaysUntil)
}

func TestSeasonalEvents(t *testing.T) {
	manager := NewEventManager()

	// Register seasonal events
	springEvent := NewSeasonalEvent("spring_sale", "Spring Sale", gametime.Spring, 1)
	summerEvent := NewSeasonalEvent("summer_festival", "Summer Festival", gametime.Summer, 15)
	autumnEvent := NewSeasonalEvent("harvest_time", "Harvest Time", gametime.Autumn, 20)
	winterEvent := NewSeasonalEvent("winter_market", "Winter Market", gametime.Winter, 10)

	manager.RegisterEvent(springEvent)
	manager.RegisterEvent(summerEvent)
	manager.RegisterEvent(autumnEvent)
	manager.RegisterEvent(winterEvent)

	// Get events for each season
	springEvents := manager.GetEventsForSeason(gametime.Spring)
	assert.Equal(t, 1, len(springEvents))
	assert.Equal(t, "spring_sale", springEvents[0].ID)

	summerEvents := manager.GetEventsForSeason(gametime.Summer)
	assert.Equal(t, 1, len(summerEvents))
	assert.Equal(t, "summer_festival", summerEvents[0].ID)
}

func TestMajorEvents(t *testing.T) {
	manager := NewEventManager()

	// Create major event with complex effects
	dragonEvent := &GameEvent{
		ID:          "dragon_defeat",
		Name:        "Dragon Defeated",
		Description: "The town celebrates defeating the dragon",
		Type:        EventTypeMajor,
		Priority:    EventPriorityUrgent,
		IsActive:    true,
		Effects: []EventEffect{
			&ReputationEffect{Amount: 50},
			&MarketBoostEffect{PriceMultiplier: 1.5, Duration: 7},
			&UnlockFeatureEffect{Feature: "dragon_scales_trading"},
		},
		Rewards: &EventRewards{
			Gold:       10000,
			Reputation: 50,
			Items:      []string{"dragon_scale", "hero_medal"},
		},
	}

	manager.RegisterEvent(dragonEvent)

	// Trigger the event
	ctx := context.Background()
	results := manager.TriggerEvent(ctx, "dragon_defeat")

	assert.NotNil(t, results)
	assert.True(t, results.Success)
	assert.Equal(t, 10000, results.Rewards.Gold)
	assert.Equal(t, 50, results.Rewards.Reputation)
	assert.Contains(t, results.Rewards.Items, "dragon_scale")
}
