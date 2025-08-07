package event

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEvent is a test implementation of Event
type TestEvent struct {
	Name      string
	Timestamp int64
	Data      string
}

func (e *TestEvent) EventName() string {
	return e.Name
}

func (e *TestEvent) OccurredAt() int64 {
	return e.Timestamp
}

func TestNewEventBus(t *testing.T) {
	eb := NewEventBus()
	assert.NotNil(t, eb)
	assert.NotNil(t, eb.handlers)
	assert.Empty(t, eb.handlers)
}

func TestEventBus_Subscribe(t *testing.T) {
	eb := NewEventBus()

	var called bool
	handler := func(e Event) error {
		called = true
		return nil
	}

	eb.Subscribe("test.event", handler)
	assert.Equal(t, 1, eb.HandlerCount("test.event"))

	// Publish event to verify subscription
	event := &TestEvent{Name: "test.event", Timestamp: time.Now().Unix()}
	err := eb.Publish(event)
	require.NoError(t, err)
	assert.True(t, called)
}

func TestEventBus_MultipleHandlers(t *testing.T) {
	eb := NewEventBus()

	var count int32
	handler1 := func(e Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	}
	handler2 := func(e Event) error {
		atomic.AddInt32(&count, 2)
		return nil
	}

	eb.Subscribe("multi.event", handler1)
	eb.Subscribe("multi.event", handler2)

	assert.Equal(t, 2, eb.HandlerCount("multi.event"))

	event := &TestEvent{Name: "multi.event", Timestamp: time.Now().Unix()}
	err := eb.Publish(event)
	require.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&count))
}

func TestEventBus_Unsubscribe(t *testing.T) {
	eb := NewEventBus()

	handler := func(e Event) error {
		return nil
	}

	eb.Subscribe("remove.event", handler)
	assert.Equal(t, 1, eb.HandlerCount("remove.event"))

	eb.Unsubscribe("remove.event")
	assert.Equal(t, 0, eb.HandlerCount("remove.event"))
}

func TestEventBus_PublishWithError(t *testing.T) {
	eb := NewEventBus()

	expectedErr := errors.New("handler error")
	handler := func(e Event) error {
		return expectedErr
	}

	eb.Subscribe("error.event", handler)

	event := &TestEvent{Name: "error.event", Timestamp: time.Now().Unix()}
	err := eb.Publish(event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event publish errors")
}

func TestEventBus_PublishAsync(t *testing.T) {
	eb := NewEventBus()

	done := make(chan bool, 1)
	handler := func(e Event) error {
		done <- true
		return nil
	}

	eb.Subscribe("async.event", handler)

	event := &TestEvent{Name: "async.event", Timestamp: time.Now().Unix()}
	eb.PublishAsync(event)

	// Wait for async handler to complete
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Async handler not called within timeout")
	}
}

func TestEventBus_Clear(t *testing.T) {
	eb := NewEventBus()

	handler := func(e Event) error {
		return nil
	}

	eb.Subscribe("event1", handler)
	eb.Subscribe("event2", handler)
	eb.Subscribe("event3", handler)

	assert.True(t, eb.HasHandlers("event1"))
	assert.True(t, eb.HasHandlers("event2"))
	assert.True(t, eb.HasHandlers("event3"))

	eb.Clear()

	assert.False(t, eb.HasHandlers("event1"))
	assert.False(t, eb.HasHandlers("event2"))
	assert.False(t, eb.HasHandlers("event3"))
}

func TestEventBus_ConcurrentPublish(t *testing.T) {
	eb := NewEventBus()

	var count int32
	handler := func(e Event) error {
		atomic.AddInt32(&count, 1)
		time.Sleep(10 * time.Millisecond) // Simulate work
		return nil
	}

	eb.Subscribe("concurrent.event", handler)

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			event := &TestEvent{
				Name:      "concurrent.event",
				Timestamp: time.Now().Unix(),
				Data:      string(rune(id)),
			}
			_ = eb.Publish(event)
		}(i)
	}

	wg.Wait()
	assert.Equal(t, int32(numGoroutines), atomic.LoadInt32(&count))
}

func TestEventBus_SubscribeToType(t *testing.T) {
	eb := NewEventBus()

	var called bool
	handler := func(e Event) error {
		called = true
		return nil
	}

	// Subscribe using type
	eb.SubscribeToType(&TestEvent{}, handler)

	// The event name should be the type string
	typeName := "*event.TestEvent"
	assert.Equal(t, 1, eb.HandlerCount(typeName))

	// Verify handler is not called for wrong event name
	wrongEvent := &TestEvent{Name: "wrong.name", Timestamp: time.Now().Unix()}
	err := eb.Publish(wrongEvent)
	require.NoError(t, err)
	assert.False(t, called)
}

func TestEventBus_NoHandlers(t *testing.T) {
	eb := NewEventBus()

	event := &TestEvent{Name: "no.handlers", Timestamp: time.Now().Unix()}
	err := eb.Publish(event)
	assert.NoError(t, err) // Should not error when no handlers
}
