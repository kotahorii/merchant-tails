package event

import (
	"fmt"
	"reflect"
	"sync"
)

// Event is the base interface for all domain events
type Event interface {
	// EventName returns the name of the event
	EventName() string
	// OccurredAt returns when the event occurred
	OccurredAt() int64
}

// Handler is a function that handles an event
type Handler func(Event) error

// EventBus manages event publishing and subscription
type EventBus struct {
	handlers map[string][]Handler
	mu       sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]Handler),
	}
}

// Subscribe registers a handler for an event type
func (eb *EventBus) Subscribe(eventName string, handler Handler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventName] = append(eb.handlers[eventName], handler)
}

// SubscribeToType subscribes to events of a specific type using reflection
func (eb *EventBus) SubscribeToType(eventType Event, handler Handler) {
	eventName := reflect.TypeOf(eventType).String()
	eb.Subscribe(eventName, handler)
}

// Unsubscribe removes all handlers for an event type
func (eb *EventBus) Unsubscribe(eventName string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	delete(eb.handlers, eventName)
}

// Publish sends an event to all registered handlers
func (eb *EventBus) Publish(event Event) error {
	eb.mu.RLock()
	handlers, exists := eb.handlers[event.EventName()]
	eb.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		return nil
	}

	// Create a copy of handlers to avoid holding the lock during execution
	handlersCopy := make([]Handler, len(handlers))
	copy(handlersCopy, handlers)

	var errors []error
	for _, handler := range handlersCopy {
		if err := handler(event); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("event publish errors: %v", errors)
	}

	return nil
}

// PublishAsync publishes an event asynchronously
func (eb *EventBus) PublishAsync(event Event) {
	go func() {
		_ = eb.Publish(event)
	}()
}

// Clear removes all event handlers
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers = make(map[string][]Handler)
}

// HandlerCount returns the number of handlers for a specific event
func (eb *EventBus) HandlerCount(eventName string) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	return len(eb.handlers[eventName])
}

// HasHandlers checks if there are any handlers for an event
func (eb *EventBus) HasHandlers(eventName string) bool {
	return eb.HandlerCount(eventName) > 0
}
