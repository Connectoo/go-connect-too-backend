package events

import (
	"context"
	"sync"
	"time"
)

// Handler processes a published event.
type Handler func(ctx context.Context, event Event)

// Dispatcher routes in-app events to subscribed handlers.
type Dispatcher struct {
	mu       sync.RWMutex
	handlers map[Type][]Handler
}

// NewDispatcher creates an event dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[Type][]Handler),
	}
}

// Subscribe registers a handler for an event type.
func (d *Dispatcher) Subscribe(eventType Type, handler Handler) {
	if handler == nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

// Publish notifies subscribers without blocking the caller.
func (d *Dispatcher) Publish(ctx context.Context, event Event) {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}

	d.mu.RLock()
	handlers := append([]Handler(nil), d.handlers[event.Type]...)
	d.mu.RUnlock()

	for _, handler := range handlers {
		go handler(ctx, event)
	}
}
