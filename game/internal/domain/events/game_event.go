// Package events provides the game event system
package events

import (
	"context"
	"sort"
	"sync"

	gametime "github.com/yourusername/merchant-tails/game/internal/domain/time"
)

// EventType defines the type of event
type EventType int

const (
	EventTypeRegular EventType = iota
	EventTypeSeasonal
	EventTypeMajor
	EventTypeRandom
)

// EventPriority defines the priority of event execution
type EventPriority int

const (
	EventPriorityLow EventPriority = iota
	EventPriorityNormal
	EventPriorityHigh
	EventPriorityUrgent
)

// ScheduleType defines how an event is scheduled
type ScheduleType int

const (
	ScheduleTypeOneTime ScheduleType = iota
	ScheduleTypeMonthly
	ScheduleTypeSeasonal
	ScheduleTypeRandom
)

// GameEvent represents a game event
type GameEvent struct {
	ID               string
	Name             string
	Description      string
	Type             EventType
	Priority         EventPriority
	IsActive         bool
	Schedule         *EventSchedule
	Conditions       []EventCondition
	Effects          []EventEffect
	Rewards          *EventRewards
	FollowUpEvents   []string
	NotificationDays int // Days in advance to notify
}

// EventSchedule defines when an event should trigger
type EventSchedule struct {
	Type        ScheduleType
	DayOfWeek   int // Day of month for monthly, day of season for seasonal
	Season      gametime.Season
	TriggerTime *gametime.GameTime // For one-time events
	Probability float64            // For random events
}

// EventCondition interface for event trigger conditions
type EventCondition interface {
	Check(ctx context.Context) bool
}

// EventEffect interface for event effects
type EventEffect interface {
	Apply(ctx context.Context, data interface{}) *EffectResult
}

// EffectResult contains the result of applying an effect
type EffectResult struct {
	Success bool
	Changes map[string]interface{}
	Error   error
}

// EventRewards contains rewards for completing an event
type EventRewards struct {
	Gold       int
	Reputation int
	Items      []string
	Experience int
}

// EventNotification represents a notification about an upcoming event
type EventNotification struct {
	EventID   string
	EventName string
	DaysUntil int
	Message   string
}

// EventTriggerResult contains the result of triggering an event
type EventTriggerResult struct {
	Success bool
	Event   *GameEvent
	Effects []*EffectResult
	Rewards *EventRewards
	Error   error
}

// NewGameEvent creates a new game event
func NewGameEvent(id, name, description string, eventType EventType, priority EventPriority) *GameEvent {
	return &GameEvent{
		ID:          id,
		Name:        name,
		Description: description,
		Type:        eventType,
		Priority:    priority,
		IsActive:    true,
		Conditions:  make([]EventCondition, 0),
		Effects:     make([]EventEffect, 0),
	}
}

// ShouldTrigger checks if the event should trigger at the given time
func (s *EventSchedule) ShouldTrigger(currentTime gametime.GameTime) bool {
	switch s.Type {
	case ScheduleTypeMonthly:
		return currentTime.Day == s.DayOfWeek
	case ScheduleTypeSeasonal:
		return currentTime.Season == s.Season && currentTime.Day == s.DayOfWeek
	case ScheduleTypeOneTime:
		if s.TriggerTime == nil {
			return false
		}
		return currentTime.Year == s.TriggerTime.Year &&
			currentTime.Season == s.TriggerTime.Season &&
			currentTime.Day == s.TriggerTime.Day
	case ScheduleTypeRandom:
		// Random events would use probability
		// Implementation would depend on random number generation
		return false
	default:
		return false
	}
}

// CheckConditions checks if all conditions are met
func (e *GameEvent) CheckConditions(ctx context.Context) bool {
	for _, condition := range e.Conditions {
		if !condition.Check(ctx) {
			return false
		}
	}
	return true
}

// EventManager manages all game events
type EventManager struct {
	events        map[string]*GameEvent
	eventHandlers []func(*GameEvent)
	mu            sync.RWMutex
}

// NewEventManager creates a new event manager
func NewEventManager() *EventManager {
	return &EventManager{
		events:        make(map[string]*GameEvent),
		eventHandlers: make([]func(*GameEvent), 0),
	}
}

// RegisterEvent registers a new event
func (em *EventManager) RegisterEvent(event *GameEvent) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.events[event.ID] = event
}

// GetEvent retrieves an event by ID
func (em *EventManager) GetEvent(id string) *GameEvent {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.events[id]
}

// GetAllEvents returns all registered events
func (em *EventManager) GetAllEvents() []*GameEvent {
	em.mu.RLock()
	defer em.mu.RUnlock()

	events := make([]*GameEvent, 0, len(em.events))
	for _, event := range em.events {
		events = append(events, event)
	}
	return events
}

// RegisterEventHandler registers a handler for events
func (em *EventManager) RegisterEventHandler(handler func(*GameEvent)) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.eventHandlers = append(em.eventHandlers, handler)
}

// Update checks for events that should trigger
func (em *EventManager) Update(ctx context.Context, currentTime gametime.GameTime) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	// Collect events that should trigger
	var triggeredEvents []*GameEvent
	for _, event := range em.events {
		if !event.IsActive {
			continue
		}

		if event.Schedule != nil && event.Schedule.ShouldTrigger(currentTime) {
			if event.CheckConditions(ctx) {
				triggeredEvents = append(triggeredEvents, event)
			}
		}
	}

	// Sort by priority (higher priority first)
	sort.Slice(triggeredEvents, func(i, j int) bool {
		return triggeredEvents[i].Priority > triggeredEvents[j].Priority
	})

	// Trigger events in priority order
	for _, event := range triggeredEvents {
		em.triggerEvent(ctx, event)
	}
}

// TriggerEvent manually triggers an event
func (em *EventManager) TriggerEvent(ctx context.Context, eventID string) *EventTriggerResult {
	em.mu.RLock()
	event, exists := em.events[eventID]
	em.mu.RUnlock()

	if !exists {
		return &EventTriggerResult{
			Success: false,
			Error:   ErrEventNotFound,
		}
	}

	result := em.triggerEvent(ctx, event)

	// Trigger follow-up events
	for _, followUpID := range event.FollowUpEvents {
		em.TriggerEvent(ctx, followUpID)
	}

	return result
}

// triggerEvent executes an event
func (em *EventManager) triggerEvent(ctx context.Context, event *GameEvent) *EventTriggerResult {
	result := &EventTriggerResult{
		Success: true,
		Event:   event,
		Effects: make([]*EffectResult, 0),
		Rewards: event.Rewards,
	}

	// Apply effects
	for _, effect := range event.Effects {
		effectResult := effect.Apply(ctx, nil)
		result.Effects = append(result.Effects, effectResult)
		if !effectResult.Success {
			result.Success = false
		}
	}

	// Notify handlers
	for _, handler := range em.eventHandlers {
		handler(event)
	}

	return result
}

// GetUpcomingEvents returns events scheduled within the specified days
func (em *EventManager) GetUpcomingEvents(currentTime gametime.GameTime, daysAhead int) []*GameEvent {
	em.mu.RLock()
	defer em.mu.RUnlock()

	upcomingEvents := make([]*GameEvent, 0)

	for _, event := range em.events {
		if !event.IsActive || event.Schedule == nil {
			continue
		}

		// Check if event is within the specified days ahead
		for i := 0; i <= daysAhead; i++ {
			futureTime := currentTime
			futureTime.Day += i

			// Handle month/season transitions
			if futureTime.Day > gametime.DaysPerSeason {
				futureTime.Day -= gametime.DaysPerSeason
				futureTime.Season++
				if futureTime.Season > gametime.Winter {
					futureTime.Season = gametime.Spring
					futureTime.Year++
				}
			}

			if event.Schedule.ShouldTrigger(futureTime) {
				upcomingEvents = append(upcomingEvents, event)
				break
			}
		}
	}

	return upcomingEvents
}

// GetEventNotifications returns notifications for upcoming events
func (em *EventManager) GetEventNotifications(currentTime gametime.GameTime) []*EventNotification {
	em.mu.RLock()
	defer em.mu.RUnlock()

	notifications := make([]*EventNotification, 0)

	for _, event := range em.events {
		if !event.IsActive || event.Schedule == nil || event.NotificationDays == 0 {
			continue
		}

		// Check if we should notify about this event
		upcomingEvents := em.GetUpcomingEvents(currentTime, event.NotificationDays)
		for _, upcomingEvent := range upcomingEvents {
			if upcomingEvent.ID == event.ID {
				// Calculate days until event
				daysUntil := 0
				for i := 1; i <= event.NotificationDays; i++ {
					futureTime := currentTime
					futureTime.Day += i
					if futureTime.Day > gametime.DaysPerSeason {
						futureTime.Day -= gametime.DaysPerSeason
						futureTime.Season++
					}
					if event.Schedule.ShouldTrigger(futureTime) {
						daysUntil = i
						break
					}
				}

				notifications = append(notifications, &EventNotification{
					EventID:   event.ID,
					EventName: event.Name,
					DaysUntil: daysUntil,
					Message:   event.Description,
				})
			}
		}
	}

	return notifications
}

// GetEventsForSeason returns all events scheduled for a specific season
func (em *EventManager) GetEventsForSeason(season gametime.Season) []*GameEvent {
	em.mu.RLock()
	defer em.mu.RUnlock()

	seasonalEvents := make([]*GameEvent, 0)
	for _, event := range em.events {
		if event.Schedule != nil &&
			event.Schedule.Type == ScheduleTypeSeasonal &&
			event.Schedule.Season == season {
			seasonalEvents = append(seasonalEvents, event)
		}
	}
	return seasonalEvents
}

// NewSeasonalEvent creates a new seasonal event
func NewSeasonalEvent(id, name string, season gametime.Season, day int) *GameEvent {
	event := NewGameEvent(id, name, "", EventTypeSeasonal, EventPriorityNormal)
	event.Schedule = &EventSchedule{
		Type:      ScheduleTypeSeasonal,
		Season:    season,
		DayOfWeek: day,
	}
	return event
}
