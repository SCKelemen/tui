package event

import (
	"sync"
	"testing"
	"time"
)

func TestNewBus(t *testing.T) {
	bus := NewBus()
	if bus == nil {
		t.Fatal("NewBus returned nil")
	}
	if bus.subscribers == nil {
		t.Fatal("expected subscribers map to be initialized")
	}
	if len(bus.subscribers) != 0 {
		t.Fatalf("expected no subscribers, got %d", len(bus.subscribers))
	}
}

func TestSubscribeAndPublishDelivers(t *testing.T) {
	bus := NewBus()
	ch := Subscribe[string](bus, "updates")

	Publish(bus, Event[string]{Name: "updates", Payload: "hello"})

	select {
	case got := <-ch:
		if got != "hello" {
			t.Fatalf("expected payload hello, got %q", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for event")
	}
}

func TestPublishToMultipleSubscribers(t *testing.T) {
	bus := NewBus()
	ch1 := Subscribe[int](bus, "counter")
	ch2 := Subscribe[int](bus, "counter")

	Publish(bus, Event[int]{Name: "counter", Payload: 42})

	select {
	case got := <-ch1:
		if got != 42 {
			t.Fatalf("subscriber 1 expected 42, got %d", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for event on subscriber 1")
	}

	select {
	case got := <-ch2:
		if got != 42 {
			t.Fatalf("subscriber 2 expected 42, got %d", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for event on subscriber 2")
	}
}

func TestUnsubscribeStopsDelivery(t *testing.T) {
	bus := NewBus()
	raw1 := make(chan interface{}, 1)
	raw2 := make(chan interface{}, 1)
	bus.subscribers["topic"] = []chan interface{}{raw1, raw2}

	Unsubscribe(bus, "topic", (<-chan interface{})(raw1))

	if len(bus.subscribers["topic"]) != 1 {
		t.Fatalf("expected one remaining subscriber, got %d", len(bus.subscribers["topic"]))
	}

	if bus.subscribers["topic"][0] != raw2 {
		t.Fatal("expected remaining subscriber to be raw2")
	}

	select {
	case _, ok := <-raw1:
		if ok {
			t.Fatal("expected unsubscribed channel to be closed")
		}
	default:
		t.Fatal("expected unsubscribed channel to be immediately readable as closed")
	}

	Publish(bus, Event[string]{Name: "topic", Payload: "still-delivered"})

	select {
	case got := <-raw2:
		if got != "still-delivered" {
			t.Fatalf("expected payload still-delivered, got %v", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for event on remaining subscriber")
	}
}

func TestPublishUnknownTopicNoPanic(t *testing.T) {
	bus := NewBus()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Publish panicked for unknown topic: %v", r)
		}
	}()

	Publish(bus, Event[string]{Name: "does-not-exist", Payload: "ignored"})
}

func TestPublishDoesNotBlockOnSlowSubscriber(t *testing.T) {
	bus := NewBus()

	// Slow subscriber: never read from this raw channel.
	slowRaw := make(chan interface{}, 1)
	bus.subscribers["topic"] = append(bus.subscribers["topic"], slowRaw)

	// Fast typed subscriber via the public API.
	fast := Subscribe[string](bus, "topic")

	done := make(chan struct{})
	go func() {
		// Publish enough events to overflow the slow subscriber's buffer.
		for i := 0; i < 5; i++ {
			Publish(bus, Event[string]{Name: "topic", Payload: "hello"})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Publish blocked on slow subscriber")
	}

	// Fast subscriber should still receive at least one event.
	select {
	case got := <-fast:
		if got != "hello" {
			t.Fatalf("expected fast subscriber to receive hello, got %q", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("fast subscriber starved by slow subscriber")
	}

	if bus.DroppedEvents() == 0 {
		t.Fatal("expected dropped event counter to be incremented")
	}
}

func TestPublishOnDropCallback(t *testing.T) {
	bus := NewBus()

	// Slow subscriber whose buffer fills immediately.
	slowRaw := make(chan interface{}, 1)
	bus.subscribers["topic"] = append(bus.subscribers["topic"], slowRaw)

	var mu sync.Mutex
	var dropped []string
	bus.SetOnDrop(func(name string) {
		mu.Lock()
		dropped = append(dropped, name)
		mu.Unlock()
	})

	for i := 0; i < 3; i++ {
		Publish(bus, Event[int]{Name: "topic", Payload: i})
	}

	mu.Lock()
	defer mu.Unlock()
	if len(dropped) == 0 {
		t.Fatal("expected at least one drop callback")
	}
	for _, n := range dropped {
		if n != "topic" {
			t.Fatalf("unexpected drop name: %q", n)
		}
	}
}
