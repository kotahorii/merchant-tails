package calendar

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewEventCalendar(t *testing.T) {
	ec := NewEventCalendar()

	assert.NotNil(t, ec)
	assert.NotNil(t, ec.events)
	assert.NotNil(t, ec.upcomingEvents)
	assert.NotNil(t, ec.activeEvents)
	assert.NotNil(t, ec.completedEvents)

	// Check default events are loaded
	assert.Greater(t, len(ec.events), 0)
}

func TestAddEvent(t *testing.T) {
	ec := NewEventCalendar()

	// Add a valid event
	event := &CalendarEvent{
		ID:          "test_event",
		Name:        "Test Event",
		Description: "A test event",
		Type:        EventTypeSpecial,
		Priority:    PriorityMedium,
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
	}

	err := ec.AddEvent(event)
	assert.NoError(t, err)

	// Try to add duplicate
	err = ec.AddEvent(event)
	assert.Error(t, err)

	// Try to add event without ID
	invalidEvent := &CalendarEvent{
		Name: "Invalid Event",
	}
	err = ec.AddEvent(invalidEvent)
	assert.Error(t, err)
}

func TestUpdateDate(t *testing.T) {
	ec := NewEventCalendar()

	// Add events with different dates
	now := time.Now()

	pastEvent := &CalendarEvent{
		ID:        "past_event",
		Name:      "Past Event",
		Type:      EventTypeSpecial,
		StartDate: now.Add(-48 * time.Hour),
		EndDate:   now.Add(-24 * time.Hour),
	}

	activeEvent := &CalendarEvent{
		ID:        "active_event",
		Name:      "Active Event",
		Type:      EventTypeSpecial,
		StartDate: now.Add(-12 * time.Hour),
		EndDate:   now.Add(12 * time.Hour),
	}

	futureEvent := &CalendarEvent{
		ID:        "future_event",
		Name:      "Future Event",
		Type:      EventTypeSpecial,
		StartDate: now.Add(24 * time.Hour),
		EndDate:   now.Add(48 * time.Hour),
	}

	ec.AddEvent(pastEvent)
	ec.AddEvent(activeEvent)
	ec.AddEvent(futureEvent)

	// Update to current date
	ec.UpdateDate(now)

	// Check active events
	active := ec.GetActiveEvents()
	found := false
	for _, e := range active {
		if e.ID == "active_event" {
			found = true
			break
		}
	}
	assert.True(t, found, "Active event should be in active list")

	// Check upcoming events
	upcoming := ec.GetUpcomingEvents(7)
	found = false
	for _, e := range upcoming {
		if e.ID == "future_event" {
			found = true
			break
		}
	}
	assert.True(t, found, "Future event should be in upcoming list")
}

func TestRecurringEvents(t *testing.T) {
	ec := NewEventCalendar()

	callbackCalled := 0
	ec.RegisterCallback(func(event *CalendarEvent, status string) {
		callbackCalled++
	})

	// Add a recurring event
	now := time.Now()
	recurringEvent := &CalendarEvent{
		ID:          "recurring_test",
		Name:        "Recurring Event",
		Type:        EventTypeMarket,
		Priority:    PriorityMedium,
		StartDate:   now.Add(-1 * time.Hour),
		EndDate:     now.Add(1 * time.Hour),
		Recurring:   true,
		RecurPeriod: 24 * time.Hour,
	}

	err := ec.AddEvent(recurringEvent)
	assert.NoError(t, err)

	// Activate the event
	ec.UpdateDate(now)

	// Move time forward to complete the event
	ec.UpdateDate(now.Add(2 * time.Hour))

	// Check that a new recurring event was scheduled
	upcoming := ec.GetUpcomingEvents(30)
	foundRecurring := false
	for _, e := range upcoming {
		if e.Recurring && e.Name == "Recurring Event" {
			foundRecurring = true
			break
		}
	}
	assert.True(t, foundRecurring, "New recurring event should be scheduled")
	assert.Greater(t, callbackCalled, 0, "Callback should have been called")
}

func TestGetEventsByType(t *testing.T) {
	ec := NewEventCalendar()

	// Add events of different types
	marketEvent := &CalendarEvent{
		ID:   "market_1",
		Name: "Market Event",
		Type: EventTypeMarket,
	}

	festivalEvent := &CalendarEvent{
		ID:   "festival_1",
		Name: "Festival Event",
		Type: EventTypeFestival,
	}

	ec.AddEvent(marketEvent)
	ec.AddEvent(festivalEvent)

	// Get market events
	marketEvents := ec.GetEventsByType(EventTypeMarket)
	found := false
	for _, e := range marketEvents {
		if e.ID == "market_1" {
			found = true
			break
		}
	}
	assert.True(t, found, "Market event should be found")

	// Get festival events
	festivalEvents := ec.GetEventsByType(EventTypeFestival)
	found = false
	for _, e := range festivalEvents {
		if e.ID == "festival_1" {
			found = true
			break
		}
	}
	assert.True(t, found, "Festival event should be found")
}

func TestGetEventEffects(t *testing.T) {
	ec := NewEventCalendar()

	// Clear default events for clean test
	ec.Reset()

	now := time.Now()

	// Add multiple active events with effects
	event1 := &CalendarEvent{
		ID:        "effect_1",
		Name:      "Effect Event 1",
		Type:      EventTypeMarket,
		StartDate: now.Add(-1 * time.Hour),
		EndDate:   now.Add(1 * time.Hour),
		Effects: map[string]float64{
			"price_modifier": 1.5,
			"demand":         1.2,
		},
	}

	event2 := &CalendarEvent{
		ID:        "effect_2",
		Name:      "Effect Event 2",
		Type:      EventTypeFestival,
		StartDate: now.Add(-1 * time.Hour),
		EndDate:   now.Add(1 * time.Hour),
		Effects: map[string]float64{
			"price_modifier": 1.2, // Should multiply with event1
			"supply":         0.8,
		},
	}

	ec.AddEvent(event1)
	ec.AddEvent(event2)

	// Activate events
	ec.UpdateDate(now)

	// Get combined effects
	effects := ec.GetEventEffects()

	// Check multiplied effect
	assert.InDelta(t, 1.8, effects["price_modifier"], 0.01) // 1.5 * 1.2
	assert.Equal(t, 1.2, effects["demand"])
	assert.Equal(t, 0.8, effects["supply"])
}

func TestGetHighPriorityEvents(t *testing.T) {
	ec := NewEventCalendar()

	// Add events with different priorities
	lowPriority := &CalendarEvent{
		ID:       "low",
		Name:     "Low Priority",
		Priority: PriorityLow,
	}

	highPriority := &CalendarEvent{
		ID:       "high",
		Name:     "High Priority",
		Priority: PriorityHigh,
	}

	criticalPriority := &CalendarEvent{
		ID:       "critical",
		Name:     "Critical Priority",
		Priority: PriorityCritical,
	}

	ec.AddEvent(lowPriority)
	ec.AddEvent(highPriority)
	ec.AddEvent(criticalPriority)

	// Get high priority events
	highEvents := ec.GetHighPriorityEvents()

	// Should include high and critical, but not low
	foundHigh := false
	foundCritical := false
	foundLow := false

	for _, e := range highEvents {
		switch e.ID {
		case "high":
			foundHigh = true
		case "critical":
			foundCritical = true
		case "low":
			foundLow = true
		}
	}

	assert.True(t, foundHigh, "High priority event should be included")
	assert.True(t, foundCritical, "Critical priority event should be included")
	assert.False(t, foundLow, "Low priority event should not be included")
}

func TestRemoveEvent(t *testing.T) {
	ec := NewEventCalendar()

	// Add an event
	event := &CalendarEvent{
		ID:        "to_remove",
		Name:      "Event to Remove",
		Type:      EventTypeSpecial,
		StartDate: time.Now().Add(24 * time.Hour),
	}

	ec.AddEvent(event)

	// Verify it exists
	retrieved, exists := ec.GetEventByID("to_remove")
	assert.True(t, exists)
	assert.NotNil(t, retrieved)

	// Remove it
	err := ec.RemoveEvent("to_remove")
	assert.NoError(t, err)

	// Verify it's gone
	retrieved, exists = ec.GetEventByID("to_remove")
	assert.False(t, exists)
	assert.Nil(t, retrieved)

	// Try to remove non-existent event
	err = ec.RemoveEvent("non_existent")
	assert.Error(t, err)
}

func TestSeasonalEvents(t *testing.T) {
	ec := NewEventCalendar()

	// Check that seasonal events are loaded
	events := ec.GetEventsByType(EventTypeFestival)

	seasonalEventNames := []string{
		"Spring Harvest Festival",
		"Grand Harvest Festival",
	}

	for _, name := range seasonalEventNames {
		found := false
		for _, e := range events {
			if e.Name == name {
				found = true
				break
			}
		}
		assert.True(t, found, "Seasonal event %s should be loaded", name)
	}
}

func TestEventCallbacks(t *testing.T) {
	ec := NewEventCalendar()

	activatedEvents := make([]string, 0)
	completedEvents := make([]string, 0)

	ec.RegisterCallback(func(event *CalendarEvent, status string) {
		switch status {
		case "activated":
			activatedEvents = append(activatedEvents, event.ID)
		case "completed":
			completedEvents = append(completedEvents, event.ID)
		}
	})

	now := time.Now()

	// Add an event that will activate and complete
	event := &CalendarEvent{
		ID:        "callback_test",
		Name:      "Callback Test",
		Type:      EventTypeSpecial,
		StartDate: now.Add(-1 * time.Hour),
		EndDate:   now.Add(1 * time.Hour),
	}

	ec.AddEvent(event)

	// Activate the event
	ec.UpdateDate(now)
	assert.Contains(t, activatedEvents, "callback_test")

	// Complete the event
	ec.UpdateDate(now.Add(2 * time.Hour))
	assert.Contains(t, completedEvents, "callback_test")
}

func TestConcurrentAccess(t *testing.T) {
	ec := NewEventCalendar()

	done := make(chan bool, 4)

	// Concurrent adds
	go func() {
		for i := 0; i < 100; i++ {
			event := &CalendarEvent{
				ID:   fmt.Sprintf("concurrent_%d", i),
				Name: fmt.Sprintf("Event %d", i),
				Type: EventTypeSpecial,
			}
			ec.AddEvent(event)
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			ec.GetActiveEvents()
			ec.GetUpcomingEvents(7)
			ec.GetEventEffects()
		}
		done <- true
	}()

	// Concurrent updates
	go func() {
		for i := 0; i < 100; i++ {
			ec.UpdateDate(time.Now().Add(time.Duration(i) * time.Hour))
		}
		done <- true
	}()

	// Concurrent removes
	go func() {
		for i := 0; i < 50; i++ {
			ec.RemoveEvent(fmt.Sprintf("concurrent_%d", i))
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}

	// Verify state is consistent
	events := ec.GetEventsByType(EventTypeSpecial)
	assert.NotNil(t, events)
}

func TestReset(t *testing.T) {
	ec := NewEventCalendar()

	// Add custom event
	customEvent := &CalendarEvent{
		ID:   "custom",
		Name: "Custom Event",
		Type: EventTypeSpecial,
	}
	ec.AddEvent(customEvent)

	// Verify it exists
	_, exists := ec.GetEventByID("custom")
	assert.True(t, exists)

	// Reset
	ec.Reset()

	// Custom event should be gone
	_, exists = ec.GetEventByID("custom")
	assert.False(t, exists)

	// Default events should be restored
	assert.Greater(t, len(ec.events), 0)
}
