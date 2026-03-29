package event

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// Event represents a named event with a typed payload.
type Event[T any] struct {
	Name    string
	Payload T
}

// Bus is a thread-safe event bus for named subscriptions.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan interface{}
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
func Publish[T any](bus *Bus, evt Event[T]) {
	bus.mu.RLock()
	subs := append([]chan interface{}(nil), bus.subscribers[evt.Name]...)
	bus.mu.RUnlock()

	for _, ch := range subs {
		ch <- evt.Payload
	}
}

// Subscribe registers a typed subscriber for an event name.
func Subscribe[T any](bus *Bus, name string) <-chan T {
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

	return typedCh
}

// Unsubscribe removes and closes a subscriber channel for an event name.
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
