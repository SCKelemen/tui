package event

import (
	"runtime"
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
	sub := SubscribeWithHandle[string](bus, "updates")
	defer sub.Close()

	Publish(bus, Event[string]{Name: "updates", Payload: "hello"})

	select {
	case got := <-sub.Chan():
		if got != "hello" {
			t.Fatalf("expected payload hello, got %q", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for event")
	}
}

func TestPublishToMultipleSubscribers(t *testing.T) {
	bus := NewBus()
	sub1 := SubscribeWithHandle[int](bus, "counter")
	sub2 := SubscribeWithHandle[int](bus, "counter")
	defer sub1.Close()
	defer sub2.Close()

	Publish(bus, Event[int]{Name: "counter", Payload: 42})

	select {
	case got := <-sub1.Chan():
		if got != 42 {
			t.Fatalf("subscriber 1 expected 42, got %d", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for event on subscriber 1")
	}

	select {
	case got := <-sub2.Chan():
		if got != 42 {
			t.Fatalf("subscriber 2 expected 42, got %d", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for event on subscriber 2")
	}
}

func TestUnsubscribeStopsDelivery(t *testing.T) {
	bus := NewBus()
	sub1 := &subscriber{ch: make(chan interface{}, 1), done: make(chan struct{})}
	sub2 := &subscriber{ch: make(chan interface{}, 1), done: make(chan struct{})}
	bus.subscribers["topic"] = []*subscriber{sub1, sub2}

	Unsubscribe(bus, "topic", (<-chan interface{})(sub1.ch))

	if len(bus.subscribers["topic"]) != 1 {
		t.Fatalf("expected one remaining subscriber, got %d", len(bus.subscribers["topic"]))
	}

	if bus.subscribers["topic"][0] != sub2 {
		t.Fatal("expected remaining subscriber to be sub2")
	}

	// Verify the unsubscribed subscriber's done channel is closed.
	select {
	case <-sub1.done:
		// done was closed as expected
	default:
		t.Fatal("expected unsubscribed subscriber's done channel to be closed")
	}

	Publish(bus, Event[string]{Name: "topic", Payload: "still-delivered"})

	select {
	case got := <-sub2.ch:
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
	slowSub := &subscriber{ch: make(chan interface{}, 1), done: make(chan struct{})}
	bus.subscribers["topic"] = append(bus.subscribers["topic"], slowSub)

	// Fast typed subscriber via the public API.
	fast := SubscribeWithHandle[string](bus, "topic")
	defer fast.Close()

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
	case got := <-fast.Chan():
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
	slowSub := &subscriber{ch: make(chan interface{}, 1), done: make(chan struct{})}
	bus.subscribers["topic"] = append(bus.subscribers["topic"], slowSub)

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

func TestSubscriptionCloseRemovesSubscriber(t *testing.T) {
	bus := NewBus()
	sub := SubscribeWithHandle[string](bus, "topic")

	bus.mu.RLock()
	if got := len(bus.subscribers["topic"]); got != 1 {
		bus.mu.RUnlock()
		t.Fatalf("expected one subscriber registered, got %d", got)
	}
	bus.mu.RUnlock()

	sub.Close()

	bus.mu.RLock()
	if got := len(bus.subscribers["topic"]); got != 0 {
		bus.mu.RUnlock()
		t.Fatalf("expected subscriber removed after Close, got %d", got)
	}
	bus.mu.RUnlock()

	// Adapter goroutine should have drained and closed the typed channel.
	select {
	case _, ok := <-sub.Chan():
		if ok {
			t.Fatal("expected typed channel to be closed after Close")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("typed channel was not closed after Close")
	}
}

func TestSubscriptionCloseIdempotent(t *testing.T) {
	bus := NewBus()
	sub := SubscribeWithHandle[int](bus, "topic")

	sub.Close()

	// Second Close must not panic, must not double-close.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("second Close panicked: %v", r)
		}
	}()
	sub.Close()

	// Concurrent Close should also be safe.
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sub.Close()
		}()
	}
	wg.Wait()
}

func TestSubscribeCloseDoesNotLeakGoroutines(t *testing.T) {
	bus := NewBus()

	// Warm up to stabilize the goroutine count.
	warm := SubscribeWithHandle[int](bus, "warm")
	warm.Close()
	// Allow the warm-up adapter goroutine to exit.
	waitForGoroutines(50 * time.Millisecond)

	before := runtime.NumGoroutine()

	const N = 200
	for i := 0; i < N; i++ {
		sub := SubscribeWithHandle[int](bus, "tight")
		sub.Close()
	}

	// Adapter goroutines exit asynchronously after done close.
	waitForGoroutines(250 * time.Millisecond)
	after := runtime.NumGoroutine()

	// Allow a small grace for runtime bookkeeping goroutines.
	if after > before+5 {
		t.Fatalf("goroutine leak suspected: before=%d after=%d (delta=%d)", before, after, after-before)
	}
}

func waitForGoroutines(d time.Duration) {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		runtime.Gosched()
		time.Sleep(2 * time.Millisecond)
	}
}

// TestSubscribe_LegacyChannelAPI verifies that Subscribe retains its
// channel-returning signature from tui/v2.12.0..tui/v2.19.0: it returns a
// <-chan T directly and delivers published events on that channel.
func TestSubscribe_LegacyChannelAPI(t *testing.T) {
	bus := NewBus()

	// Compile-time assertion: Subscribe must return <-chan T.
	var ch <-chan string = Subscribe[string](bus, "legacy")

	Publish(bus, Event[string]{Name: "legacy", Payload: "world"})

	select {
	case got, ok := <-ch:
		if !ok {
			t.Fatal("legacy channel closed before receiving event")
		}
		if got != "world" {
			t.Fatalf("expected payload world, got %q", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for event on legacy channel")
	}
}

// TestSubscribeWithHandle_Close verifies that calling Close on a
// SubscribeWithHandle subscription stops delivery: subsequent publishes do
// not deliver to the closed sub, and its typed channel is observed closed.
func TestSubscribeWithHandle_Close(t *testing.T) {
	bus := NewBus()
	sub := SubscribeWithHandle[int](bus, "handle-close")

	// Establish that delivery works before Close.
	Publish(bus, Event[int]{Name: "handle-close", Payload: 1})
	select {
	case got := <-sub.Chan():
		if got != 1 {
			t.Fatalf("expected 1, got %d", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for pre-Close event")
	}

	sub.Close()

	// Publishing after Close must not deliver to the closed sub.
	Publish(bus, Event[int]{Name: "handle-close", Payload: 2})

	// Drain Chan and assert it closes without yielding payload 2.
	deadline := time.After(250 * time.Millisecond)
	for {
		select {
		case got, ok := <-sub.Chan():
			if !ok {
				// Channel closed cleanly: contract satisfied.
				return
			}
			if got == 2 {
				t.Fatalf("received event after Close: %d", got)
			}
			// Drain any pre-Close residual payloads (none expected here)
			// and keep waiting for the close signal.
		case <-deadline:
			t.Fatal("typed channel was not closed after Close")
		}
	}
}

// ---------------------------------------------------------------------------
// New tests for tui/v2.20.1 concurrency fixes
// ---------------------------------------------------------------------------

// TestPublishConcurrentCloseDoesNotPanic reproduces Bug A: Publish racing
// Subscription.Close used to send on a closed channel and panic. With the
// done-channel fix, Publish's select skips torn-down subscribers.
func TestPublishConcurrentCloseDoesNotPanic(t *testing.T) {
	bus := NewBus()
	var wg sync.WaitGroup
	panicCh := make(chan any, 64)
	stop := make(chan struct{})

	safeGo := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					select {
					case panicCh <- r:
					default:
					}
				}
			}()
			fn()
		}()
	}

	// Publishers
	for i := 0; i < 8; i++ {
		safeGo(func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				Publish(bus, Event[int]{Name: "race", Payload: 42})
			}
		})
	}

	// Subscribe-then-Close churners
	for i := 0; i < 4; i++ {
		safeGo(func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				sub := SubscribeWithHandle[int](bus, "race")
				sub.Close()
			}
		})
	}

	time.Sleep(500 * time.Millisecond)
	close(stop)
	wg.Wait()

	select {
	case r := <-panicCh:
		t.Fatalf("panic during concurrent Publish/Close: %v", r)
	default:
	}
}

// TestCloseUnblocksAdapterWithFullBuffer reproduces Bug B: the adapter
// goroutine used to leak when Close was called while the adapter was
// parked on typedCh <- value because the consumer stopped draining. The
// done-aware inner select now unblocks the adapter.
func TestCloseUnblocksAdapterWithFullBuffer(t *testing.T) {
	bus := NewBus()
	sub := SubscribeWithHandle[int](bus, "topic")

	// Fill the typed channel buffer (size 1) by publishing and not draining.
	Publish(bus, Event[int]{Name: "topic", Payload: 1})
	// Now the adapter should be parked on typedCh <- value for the next event,
	// OR have already delivered the first event into typedCh and be blocked on
	// rawCh receive. Publish enough to guarantee the adapter is blocked on send.
	for i := 0; i < 4; i++ {
		Publish(bus, Event[int]{Name: "topic", Payload: i + 2})
	}

	// Close while the adapter is potentially parked on typedCh <- value.
	sub.Close()

	// Within 250ms, Chan() must close (adapter exited, defer close(typedCh) ran).
	done := make(chan struct{})
	go func() {
		for range sub.Chan() {
			// drain remaining buffered items
		}
		close(done)
	}()

	select {
	case <-done:
		// adapter exited cleanly
	case <-time.After(250 * time.Millisecond):
		t.Fatal("adapter goroutine did not exit within 250ms after Close — Bug B regression")
	}
}

// TestSubscribeLegacyChannelStillDelivers is a belt-and-suspenders
// regression guard for the legacy channel-returning Subscribe API after
// the subscriber struct refactor.
func TestSubscribeLegacyChannelStillDelivers(t *testing.T) {
	bus := NewBus()
	ch := Subscribe[string](bus, "topic")

	Publish(bus, Event[string]{Name: "topic", Payload: "hello"})

	select {
	case got := <-ch:
		if got != "hello" {
			t.Fatalf("expected hello, got %q", got)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("legacy Subscribe did not deliver event within 100ms")
	}
}
