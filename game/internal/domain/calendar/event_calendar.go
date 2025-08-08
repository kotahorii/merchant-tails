package calendar

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// EventType represents the type of calendar event
type EventType int

const (
	EventTypeMarket EventType = iota
	EventTypeFestival
	EventTypeWeather
	EventTypeQuest
	EventTypeSpecial
	EventTypeMaintenance
)

// EventPriority represents the priority of an event
type EventPriority int

const (
	PriorityLow EventPriority = iota
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

// CalendarEvent represents an event in the game calendar
type CalendarEvent struct {
	ID          string
	Name        string
	Description string
	Type        EventType
	Priority    EventPriority
	StartDate   time.Time
	EndDate     time.Time
	Recurring   bool
	RecurPeriod time.Duration // For recurring events
	Effects     map[string]float64
	Rewards     map[string]int
	Active      bool
	Completed   bool
}

// EventCalendar manages all scheduled events
type EventCalendar struct {
	events          map[string]*CalendarEvent
	upcomingEvents  []*CalendarEvent
	activeEvents    []*CalendarEvent
	completedEvents map[string]bool
	currentDate     time.Time
	callbacks       []EventCallback
	mu              sync.RWMutex
}

// EventCallback is called when event status changes
type EventCallback func(event *CalendarEvent, status string)

// NewEventCalendar creates a new event calendar
func NewEventCalendar() *EventCalendar {
	ec := &EventCalendar{
		events:          make(map[string]*CalendarEvent),
		upcomingEvents:  make([]*CalendarEvent, 0),
		activeEvents:    make([]*CalendarEvent, 0),
		completedEvents: make(map[string]bool),
		currentDate:     time.Now(),
		callbacks:       make([]EventCallback, 0),
	}

	ec.initializeDefaultEvents()
	return ec
}

// initializeDefaultEvents sets up regular calendar events
func (ec *EventCalendar) initializeDefaultEvents() {
	// Weekly market day
	ec.AddEvent(&CalendarEvent{
		ID:          "weekly_market",
		Name:        "Weekly Market Day",
		Description: "Special market prices and increased customer traffic",
		Type:        EventTypeMarket,
		Priority:    PriorityMedium,
		Recurring:   true,
		RecurPeriod: 7 * 24 * time.Hour,
		Effects: map[string]float64{
			"customer_traffic": 1.5,
			"price_variance":   1.2,
		},
	})

	// Monthly festival
	ec.AddEvent(&CalendarEvent{
		ID:          "monthly_festival",
		Name:        "Town Festival",
		Description: "Celebration with special items and bonuses",
		Type:        EventTypeFestival,
		Priority:    PriorityHigh,
		Recurring:   true,
		RecurPeriod: 30 * 24 * time.Hour,
		Effects: map[string]float64{
			"luxury_demand": 2.0,
			"food_demand":   1.8,
			"customer_mood": 1.5,
		},
		Rewards: map[string]int{
			"reputation": 10,
			"gold":       500,
		},
	})

	// Seasonal events
	ec.addSeasonalEvents()
}

// addSeasonalEvents adds season-specific events
func (ec *EventCalendar) addSeasonalEvents() {
	// Spring Harvest
	ec.AddEvent(&CalendarEvent{
		ID:          "spring_harvest",
		Name:        "Spring Harvest Festival",
		Description: "Celebration of spring crops with reduced food prices",
		Type:        EventTypeFestival,
		Priority:    PriorityHigh,
		Effects: map[string]float64{
			"food_price":       0.7,
			"food_supply":      1.5,
			"customer_traffic": 1.3,
		},
	})

	// Summer Tournament
	ec.AddEvent(&CalendarEvent{
		ID:          "summer_tournament",
		Name:        "Knights' Tournament",
		Description: "Weapon and armor demand increases",
		Type:        EventTypeSpecial,
		Priority:    PriorityHigh,
		Effects: map[string]float64{
			"weapon_demand": 2.5,
			"armor_demand":  2.0,
			"tool_demand":   1.5,
		},
	})

	// Autumn Harvest
	ec.AddEvent(&CalendarEvent{
		ID:          "autumn_harvest",
		Name:        "Grand Harvest Festival",
		Description: "Biggest festival of the year",
		Type:        EventTypeFestival,
		Priority:    PriorityCritical,
		Effects: map[string]float64{
			"all_demand":       1.5,
			"customer_traffic": 2.0,
			"reputation_gain":  2.0,
		},
		Rewards: map[string]int{
			"reputation": 25,
			"gold":       1000,
			"exp":        500,
		},
	})

	// Winter Solstice
	ec.AddEvent(&CalendarEvent{
		ID:          "winter_solstice",
		Name:        "Winter Solstice Market",
		Description: "Special winter items and magical goods",
		Type:        EventTypeMarket,
		Priority:    PriorityMedium,
		Effects: map[string]float64{
			"magic_demand":    3.0,
			"clothing_demand": 2.0,
			"warmth_demand":   2.5,
		},
	})
}

// AddEvent adds a new event to the calendar
func (ec *EventCalendar) AddEvent(event *CalendarEvent) error {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if event.ID == "" {
		return fmt.Errorf("event ID cannot be empty")
	}

	if _, exists := ec.events[event.ID]; exists {
		return fmt.Errorf("event with ID %s already exists", event.ID)
	}

	ec.events[event.ID] = event

	// Add to upcoming if future event
	if event.StartDate.After(ec.currentDate) {
		ec.upcomingEvents = append(ec.upcomingEvents, event)
		ec.sortUpcomingEvents()
	}

	return nil
}

// UpdateDate updates the current date and activates/deactivates events
func (ec *EventCalendar) UpdateDate(newDate time.Time) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.currentDate = newDate

	// Check all events for status changes
	newActive := make([]*CalendarEvent, 0)
	newUpcoming := make([]*CalendarEvent, 0)

	// Check all events, not just upcoming ones
	for _, event := range ec.events {
		// Skip completed events
		if event.Completed {
			continue
		}

		// Check if event should be active
		if (event.StartDate.Before(newDate) || event.StartDate.Equal(newDate)) &&
			(event.EndDate.After(newDate) || event.EndDate.IsZero()) {
			if !event.Active {
				event.Active = true
				ec.notifyCallbacks(event, "activated")
			}
			newActive = append(newActive, event)
		} else if event.StartDate.After(newDate) {
			// Event is in the future
			newUpcoming = append(newUpcoming, event)
		} else if !event.EndDate.IsZero() && event.EndDate.Before(newDate) {
			// Event has ended
			if event.Active {
				event.Active = false
				event.Completed = true
				ec.completedEvents[event.ID] = true
				ec.notifyCallbacks(event, "completed")

				// Handle recurring events
				if event.Recurring && event.RecurPeriod > 0 {
					ec.scheduleRecurringEvent(event)
				}
			}
		}
	}

	// Update lists
	ec.activeEvents = newActive
	ec.upcomingEvents = newUpcoming
	ec.sortUpcomingEvents()
}

// scheduleRecurringEvent creates the next occurrence of a recurring event
func (ec *EventCalendar) scheduleRecurringEvent(event *CalendarEvent) {
	nextEvent := &CalendarEvent{
		ID:          fmt.Sprintf("%s_%d", event.ID, time.Now().Unix()),
		Name:        event.Name,
		Description: event.Description,
		Type:        event.Type,
		Priority:    event.Priority,
		StartDate:   event.StartDate.Add(event.RecurPeriod),
		EndDate:     event.EndDate,
		Recurring:   true,
		RecurPeriod: event.RecurPeriod,
		Effects:     event.Effects,
		Rewards:     event.Rewards,
		Active:      false,
		Completed:   false,
	}

	if !nextEvent.EndDate.IsZero() {
		nextEvent.EndDate = nextEvent.EndDate.Add(event.RecurPeriod)
	}

	ec.events[nextEvent.ID] = nextEvent
	ec.upcomingEvents = append(ec.upcomingEvents, nextEvent)
	ec.sortUpcomingEvents()
}

// GetUpcomingEvents returns events in the next N days
func (ec *EventCalendar) GetUpcomingEvents(days int) []*CalendarEvent {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	cutoffDate := ec.currentDate.AddDate(0, 0, days)
	upcoming := make([]*CalendarEvent, 0)

	for _, event := range ec.upcomingEvents {
		if event.StartDate.Before(cutoffDate) {
			upcoming = append(upcoming, event)
		}
	}

	return upcoming
}

// GetActiveEvents returns currently active events
func (ec *EventCalendar) GetActiveEvents() []*CalendarEvent {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	active := make([]*CalendarEvent, len(ec.activeEvents))
	copy(active, ec.activeEvents)
	return active
}

// GetEventsByType returns events of a specific type
func (ec *EventCalendar) GetEventsByType(eventType EventType) []*CalendarEvent {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	events := make([]*CalendarEvent, 0)
	for _, event := range ec.events {
		if event.Type == eventType {
			events = append(events, event)
		}
	}
	return events
}

// GetEventEffects returns combined effects of all active events
func (ec *EventCalendar) GetEventEffects() map[string]float64 {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	effects := make(map[string]float64)

	for _, event := range ec.activeEvents {
		for effect, value := range event.Effects {
			if current, exists := effects[effect]; exists {
				// Multiply effects
				effects[effect] = current * value
			} else {
				effects[effect] = value
			}
		}
	}

	return effects
}

// RegisterCallback registers a callback for event status changes
func (ec *EventCalendar) RegisterCallback(callback EventCallback) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.callbacks = append(ec.callbacks, callback)
}

// notifyCallbacks notifies all registered callbacks
func (ec *EventCalendar) notifyCallbacks(event *CalendarEvent, status string) {
	for _, callback := range ec.callbacks {
		callback(event, status)
	}
}

// sortUpcomingEvents sorts upcoming events by start date
func (ec *EventCalendar) sortUpcomingEvents() {
	sort.Slice(ec.upcomingEvents, func(i, j int) bool {
		return ec.upcomingEvents[i].StartDate.Before(ec.upcomingEvents[j].StartDate)
	})
}

// RemoveEvent removes an event from the calendar
func (ec *EventCalendar) RemoveEvent(eventID string) error {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if _, exists := ec.events[eventID]; !exists {
		return fmt.Errorf("event with ID %s not found", eventID)
	}

	delete(ec.events, eventID)

	// Remove from upcoming events
	newUpcoming := make([]*CalendarEvent, 0)
	for _, event := range ec.upcomingEvents {
		if event.ID != eventID {
			newUpcoming = append(newUpcoming, event)
		}
	}
	ec.upcomingEvents = newUpcoming

	// Remove from active events
	newActive := make([]*CalendarEvent, 0)
	for _, event := range ec.activeEvents {
		if event.ID != eventID {
			newActive = append(newActive, event)
		}
	}
	ec.activeEvents = newActive

	return nil
}

// GetEventByID returns a specific event
func (ec *EventCalendar) GetEventByID(eventID string) (*CalendarEvent, bool) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	event, exists := ec.events[eventID]
	return event, exists
}

// GetHighPriorityEvents returns events with high or critical priority
func (ec *EventCalendar) GetHighPriorityEvents() []*CalendarEvent {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	events := make([]*CalendarEvent, 0)
	for _, event := range ec.events {
		if event.Priority >= PriorityHigh {
			events = append(events, event)
		}
	}
	return events
}

// Reset resets the calendar
func (ec *EventCalendar) Reset() {
	ec.mu.Lock()
	ec.events = make(map[string]*CalendarEvent)
	ec.upcomingEvents = make([]*CalendarEvent, 0)
	ec.activeEvents = make([]*CalendarEvent, 0)
	ec.completedEvents = make(map[string]bool)
	ec.currentDate = time.Now()
	ec.mu.Unlock()

	// Initialize default events after unlocking
	ec.initializeDefaultEvents()
}
