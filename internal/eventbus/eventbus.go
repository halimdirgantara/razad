// Package eventbus provides a lightweight in-process event pub/sub system.
// For MVP this uses Go channels; a NATS-backed implementation can replace this
// when the cloud control plane is built.
package eventbus

import (
	"sync"
)

// Event is the interface that all bus events must implement.
type Event interface {
	// Type returns a stable event type identifier for routing.
	Type() string
}

// Handler processes events of a specific type.
type Handler func(Event)

// Bus is an in-process event pub/sub.
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// New creates a new event bus.
func New() *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
	}
}

// Subscribe registers a handler for events of the given type.
// Returns an unsubscribe function.
func (b *Bus) Subscribe(eventType string, handler Handler) func() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)

	// Capture the handler identity for unsubscribe.
	idx := len(b.handlers[eventType]) - 1
	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		handlers := b.handlers[eventType]
		if idx < len(handlers) {
			b.handlers[eventType] = append(handlers[:idx], handlers[idx+1:]...)
		}
	}
}

// Publish sends an event to all subscribed handlers. Each handler runs
// synchronously in a goroutine so a slow handler does not block others.
func (b *Bus) Publish(event Event) {
	b.mu.RLock()
	handlers := make([]Handler, len(b.handlers[event.Type()]))
	copy(handlers, b.handlers[event.Type()])
	b.mu.RUnlock()

	for _, h := range handlers {
		go h(event)
	}
}

// HasSubscribers returns true if the given event type has any subscribers.
func (b *Bus) HasSubscribers(eventType string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.handlers[eventType]) > 0
}
