package event

import (
	"sync"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
)

// Event represents a named event with a typed payload.
type Event[T any] struct {
	Name    string
	Payload T
}

// Bus is a thread-safe event bus for named subscriptions.
//
// Subscribers receive payloads on buffered channels. If a subscriber is slow
// (its channel buffer is full when Publish runs), Publish does not block; the
// event is dropped for that subscriber and the bus drop counter is incremented.
// Use DroppedEvents to observe how many events have been dropped, or
// SetOnDrop to register a callback for each drop.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan interface{}

	dropped atomic.Uint64

	dropMu sync.RWMutex
	onDrop func(name string)
}

// BusMsg is a Bubble Tea message that carries a bus event payload.
type BusMsg struct {
	Name    string
	Payload interface{}
}

// NewBus creates a new event bus.
func NewBus() *Bus {
	return &Bus{
		subscribers: make(map[string][]chan interface{}),
	}
}

// Publish sends an event payload to all subscribers of the event name.
//
// Publish is non-blocking: if a subscriber's channel buffer is full, the
// payload is dropped for that subscriber rather than blocking the publisher
// or starving other subscribers. Dropped events are counted (see
// DroppedEvents) and optionally reported via the OnDrop callback.
func Publish[T any](bus *Bus, evt Event[T]) {
	bus.mu.RLock()
	subs := append([]chan interface{}(nil), bus.subscribers[evt.Name]...)
	bus.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- evt.Payload:
		default:
			bus.dropped.Add(1)
			bus.dropMu.RLock()
			cb := bus.onDrop
			bus.dropMu.RUnlock()
			if cb != nil {
				cb(evt.Name)
			}
		}
	}
}

// DroppedEvents returns the total number of events that have been dropped
// across all subscribers because a subscriber's channel buffer was full
// when Publish was called.
func (bus *Bus) DroppedEvents() uint64 {
	return bus.dropped.Load()
}

// SetOnDrop registers a callback invoked whenever Publish drops an event
// because a subscriber's buffer was full. Pass nil to clear the callback.
// The callback must be cheap and non-blocking; it runs on the publisher
// goroutine.
func (bus *Bus) SetOnDrop(cb func(name string)) {
	bus.dropMu.Lock()
	bus.onDrop = cb
	bus.dropMu.Unlock()
}

// Subscription is a handle to an active typed subscription on a Bus. It owns
// both the raw fan-out channel registered on the bus and the goroutine that
// adapts payloads to the requested type. Callers must invoke Close when they
// are done to release these resources.
type Subscription[T any] struct {
	bus     *Bus
	name    string
	rawCh   chan interface{}
	typedCh chan T

	closeOnce sync.Once
}

// C returns the typed receive channel for this subscription. The channel is
// closed after Close has been called and the adapter goroutine has drained
// any pending payloads.
func (s *Subscription[T]) C() <-chan T {
	return s.typedCh
}

// Close removes this subscription from the bus and lets its adapter
// goroutine terminate. Close is idempotent and safe to call concurrently.
// It always returns nil; the signature returns error to leave room for
// future failure modes (e.g. a closed bus).
func (s *Subscription[T]) Close() error {
	s.closeOnce.Do(func() {
		Unsubscribe(s.bus, s.name, (<-chan interface{})(s.rawCh))
	})
	return nil
}

// Subscribe registers a typed subscriber for an event name and returns a
// Subscription handle. Callers must invoke Subscription.Close to release the
// underlying channel and stop the adapter goroutine. Use Subscription.C to
// access the typed receive channel.
func Subscribe[T any](bus *Bus, name string) *Subscription[T] {
	rawCh := make(chan interface{}, 1)
	typedCh := make(chan T, 1)

	bus.mu.Lock()
	bus.subscribers[name] = append(bus.subscribers[name], rawCh)
	bus.mu.Unlock()

	go func() {
		defer close(typedCh)
		for payload := range rawCh {
			value, ok := payload.(T)
			if !ok {
				continue
			}
			typedCh <- value
		}
	}()

	return &Subscription[T]{
		bus:     bus,
		name:    name,
		rawCh:   rawCh,
		typedCh: typedCh,
	}
}

// Unsubscribe removes and closes a subscriber channel for an event name.
//
// Prefer Subscription.Close over this lower-level helper for typed
// subscribers created via Subscribe; Close cleanly tears down both the raw
// channel and the adapter goroutine.
func Unsubscribe(bus *Bus, name string, ch <-chan interface{}) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	subs := bus.subscribers[name]
	for i, sub := range subs {
		if (<-chan interface{})(sub) == ch {
			close(sub)
			subs = append(subs[:i], subs[i+1:]...)
			if len(subs) == 0 {
				delete(bus.subscribers, name)
			} else {
				bus.subscribers[name] = subs
			}
			return
		}
	}
}

// PublishCmd returns a Bubble Tea command that emits a BusMsg.
func PublishCmd[T any](evt Event[T]) tea.Cmd {
	return func() tea.Msg {
		return BusMsg{
			Name:    evt.Name,
			Payload: evt.Payload,
		}
	}
}
