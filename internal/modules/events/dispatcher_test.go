package events

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestDispatcherPublishNotifiesSubscribers(t *testing.T) {
	d := NewDispatcher()

	var count atomic.Int32
	d.Subscribe(TypeBookingCreated, func(_ context.Context, event Event) {
		if event.Type != TypeBookingCreated {
			t.Errorf("event type = %q, want %q", event.Type, TypeBookingCreated)
		}
		count.Add(1)
	})

	d.Publish(context.Background(), Event{
		Type:    TypeBookingCreated,
		Payload: map[string]any{"booking_id": "abc"},
	})

	deadline := time.Now().Add(500 * time.Millisecond)
	for count.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if count.Load() != 1 {
		t.Fatalf("handler call count = %d, want 1", count.Load())
	}
}
