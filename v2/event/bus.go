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

// subscriber is an internal handle for a single bus subscriber. Each
// subscriber owns a payload channel and a done channel used for
// cancellation signaling. The done channel is closed by Unsubscribe to
// signal the adapter goroutine and to let Publish skip torn-down
// subscribers without sending on a closed channel.
type subscriber struct {
	ch   chan interface{}
	done chan struct{}
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
	subscribers map[string][]*subscriber

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
		subscribers: make(map[string][]*subscriber),
	}
}

// Publish sends an event payload to all subscribers of the event name.
//
// Publish is non-blocking: if a subscriber's channel buffer is full, the
// payload is dropped for that subscriber rather than blocking the publisher
// or starving other subscribers. Dropped events are counted (see
// DroppedEvents) and optionally reported via the OnDrop callback.
//
// If a subscriber is concurrently torn down (its done channel is closed),
// Publish skips it without panicking or counting a drop.
func Publish[T any](bus *Bus, evt Event[T]) {
	bus.mu.RLock()
	subs := append([]*subscriber(nil), bus.subscribers[evt.Name]...)
	bus.mu.RUnlock()

	for _, sub := range subs {
		select {
		case sub.ch <- evt.Payload:
		case <-sub.done:
			// subscriber unsubscribed concurrently; skip without counting as a drop
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

// Subscription is a closeable handle for an event subscription. Use
// SubscribeWithHandle to obtain one. Call Close to unsubscribe and free the
// underlying goroutine; Chan exposes the typed receive channel.
//
// A Subscription owns both the raw fan-out channel registered on the bus and
// the goroutine that adapts payloads to the requested type.
type Subscription[T any] struct {
	bus     *Bus
	name    string
	sub     *subscriber
	typedCh chan T

	closeOnce sync.Once
}

// Chan returns the typed receive channel for this subscription. The channel
// is closed after Close has been called and the adapter goroutine has
// drained any pending payloads.
func (s *Subscription[T]) Chan() <-chan T {
	return s.typedCh
}

// Close removes this subscription from the bus and lets its adapter
// goroutine terminate. Close is idempotent and safe to call concurrently.
func (s *Subscription[T]) Close() {
	s.closeOnce.Do(func() {
		Unsubscribe(s.bus, s.name, (<-chan interface{})(s.sub.ch))
	})
}

// Subscribe registers a typed subscriber for the named event and returns
// the receive channel. The subscription lives until the bus is closed.
//
// Source-compatible with tui/v2.12.0..tui/v2.19.0. Prefer
// SubscribeWithHandle in new code so callers can Close() the subscription
// explicitly and avoid goroutine leaks.
func Subscribe[T any](bus *Bus, name string) <-chan T {
	return SubscribeWithHandle[T](bus, name).Chan()
}

// SubscribeWithHandle registers a typed subscriber and returns a
// Subscription handle. Call Close on the handle to unsubscribe and release
// the adapter goroutine. Use Chan to access the typed receive channel.
func SubscribeWithHandle[T any](bus *Bus, name string) *Subscription[T] {
	sub := &subscriber{
		ch:   make(chan interface{}, 1),
		done: make(chan struct{}),
	}
	typedCh := make(chan T, 1)

	bus.mu.Lock()
	bus.subscribers[name] = append(bus.subscribers[name], sub)
	bus.mu.Unlock()

	go func() {
		defer close(typedCh)
		for {
			select {
			case <-sub.done:
				return
			case payload := <-sub.ch:
				value, ok := payload.(T)
				if !ok {
					continue
				}
				select {
				case typedCh <- value:
				case <-sub.done:
					return
				}
			}
		}
	}()

	return &Subscription[T]{
		bus:     bus,
		name:    name,
		sub:     sub,
		typedCh: typedCh,
	}
}

// Unsubscribe removes a subscriber channel for an event name and signals
// the subscriber's done channel so the adapter goroutine exits cleanly.
// The raw payload channel is intentionally not closed; it becomes
// unreachable once the adapter exits and the subscriber is removed from the
// map, so GC reclaims it.
//
// Prefer Subscription.Close over this lower-level helper for typed
// subscribers created via SubscribeWithHandle; Close cleanly tears down
// both the raw channel and the adapter goroutine.
func Unsubscribe(bus *Bus, name string, ch <-chan interface{}) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	subs := bus.subscribers[name]
	for i, sub := range subs {
		if (<-chan interface{})(sub.ch) == ch {
			close(sub.done)
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
