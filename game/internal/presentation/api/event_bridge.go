package api

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/event"
)

// EventBridge manages event communication between Go and Godot
type EventBridge struct {
	eventBus      *event.EventBus
	eventQueue    []QueuedEvent
	subscribers   map[string][]EventHandler
	godotCallback GodotEventCallback
	mu            sync.RWMutex
}

// QueuedEvent represents an event waiting to be sent to Godot
type QueuedEvent struct {
	Name      string
	Data      string
	Timestamp time.Time
}

// EventHandler is a function that handles events
type EventHandler func(event event.Event)

// GodotEventCallback is the callback function to send events to Godot
type GodotEventCallback func(eventName string, eventData string)

// NewEventBridge creates a new event bridge
func NewEventBridge() *EventBridge {
	eb := &EventBridge{
		eventBus:    event.GetGlobalEventBus(),
		eventQueue:  make([]QueuedEvent, 0),
		subscribers: make(map[string][]EventHandler),
	}

	// Subscribe to all game events
	eb.setupEventSubscriptions()

	return eb
}

// SetGodotCallback sets the callback function for sending events to Godot
func (eb *EventBridge) SetGodotCallback(callback GodotEventCallback) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.godotCallback = callback
}

// setupEventSubscriptions subscribes to all relevant game events
func (eb *EventBridge) setupEventSubscriptions() {
	// Game state events
	eb.subscribeToEvent("game.started")
	eb.subscribeToEvent("game.paused")
	eb.subscribeToEvent("game.resumed")
	eb.subscribeToEvent("GameVictory")
	eb.subscribeToEvent("GameDefeat")

	// Trading events
	eb.subscribeToEvent(event.EventNameTransactionComplete)
	eb.subscribeToEvent("trade.failed")

	// Market events
	eb.subscribeToEvent(event.EventNamePriceUpdated)
	eb.subscribeToEvent(event.EventNameMarketEventOccurred)

	// Progression events
	eb.subscribeToEvent("RankUp")
	eb.subscribeToEvent("AchievementUnlocked")
	eb.subscribeToEvent("FeatureUnlocked")

	// Time events
	eb.subscribeToEvent("time.advanced")
	eb.subscribeToEvent(event.EventNameDayEnded)
	eb.subscribeToEvent(event.EventNameSeasonChanged)

	// Inventory events
	eb.subscribeToEvent(event.EventNameInventoryChanged)
	eb.subscribeToEvent("ItemSpoiled")
}

// subscribeToEvent subscribes to a specific event type
func (eb *EventBridge) subscribeToEvent(eventType string) {
	eb.eventBus.Subscribe(eventType, func(e event.Event) {
		eb.handleEvent(e)
	})
}

// handleEvent processes an event and queues it for Godot
func (eb *EventBridge) handleEvent(e event.Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Convert event data to JSON
	jsonData, err := json.Marshal(e.Data)
	if err != nil {
		jsonData = []byte("{}")
	}

	// Queue the event
	queuedEvent := QueuedEvent{
		Name:      e.Type,
		Data:      string(jsonData),
		Timestamp: e.Timestamp,
	}
	eb.eventQueue = append(eb.eventQueue, queuedEvent)

	// If we have a Godot callback, send immediately
	if eb.godotCallback != nil {
		eb.godotCallback(queuedEvent.Name, queuedEvent.Data)
		// Clear the queue after sending
		eb.eventQueue = eb.eventQueue[:0]
	}
}

// FlushEvents sends all queued events to Godot
func (eb *EventBridge) FlushEvents() []QueuedEvent {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if len(eb.eventQueue) == 0 {
		return nil
	}

	// Return copy of events and clear queue
	events := make([]QueuedEvent, len(eb.eventQueue))
	copy(events, eb.eventQueue)
	eb.eventQueue = eb.eventQueue[:0]

	return events
}

// PublishToGodot publishes an event directly to Godot
func (eb *EventBridge) PublishToGodot(eventName string, data map[string]interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		jsonData = []byte("{}")
	}

	eb.mu.RLock()
	callback := eb.godotCallback
	eb.mu.RUnlock()

	if callback != nil {
		callback(eventName, string(jsonData))
	}
}

// GetQueuedEvents returns all queued events without clearing them
func (eb *EventBridge) GetQueuedEvents() []QueuedEvent {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	events := make([]QueuedEvent, len(eb.eventQueue))
	copy(events, eb.eventQueue)
	return events
}

// ClearEventQueue clears all queued events
func (eb *EventBridge) ClearEventQueue() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.eventQueue = eb.eventQueue[:0]
}
