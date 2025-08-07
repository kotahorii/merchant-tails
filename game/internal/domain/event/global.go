package event

import "sync"

var (
	globalEventBus *EventBus
	once           sync.Once
)

// GetGlobalEventBus returns the global event bus instance
func GetGlobalEventBus() *EventBus {
	once.Do(func() {
		globalEventBus = NewEventBus()
	})
	return globalEventBus
}

// Publish publishes an event to the global event bus
func Publish(event Event) error {
	return GetGlobalEventBus().Publish(event)
}

// PublishAsync publishes an event asynchronously to the global event bus
func PublishAsync(event Event) {
	GetGlobalEventBus().PublishAsync(event)
}

// Subscribe subscribes to events on the global event bus
func Subscribe(eventName string, handler Handler) {
	GetGlobalEventBus().Subscribe(eventName, handler)
}

// Unsubscribe removes handlers from the global event bus
func Unsubscribe(eventName string) {
	GetGlobalEventBus().Unsubscribe(eventName)
}

// ResetGlobalEventBus resets the global event bus (useful for testing)
func ResetGlobalEventBus() {
	globalEventBus = NewEventBus()
}
