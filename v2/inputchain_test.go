package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// recordingHandler returns an InputHandler that appends every observed
// message to *seen and reports the supplied handled flag. The optional
// cmd is returned as-is.
func recordingHandler(seen *[]tea.Msg, handled bool, cmd tea.Cmd) InputHandler {
	return func(msg tea.Msg) (bool, tea.Cmd) {
		*seen = append(*seen, msg)
		return handled, cmd
	}
}

func TestInputChainHandlerObservesMessage(t *testing.T) {
	app := NewApplication()
	c := &mockComponent{}
	app.AddComponent(c)

	var seen []tea.Msg
	app.Use(recordingHandler(&seen, false, nil))

	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	app.Update(key)

	if len(seen) != 1 {
		t.Fatalf("expected handler to observe 1 message, got %d", len(seen))
	}
	got, ok := seen[0].(tea.KeyMsg)
	if !ok {
		t.Fatalf("expected handler to see a tea.KeyMsg, got %T", seen[0])
	}
	if got.String() != key.String() {
		t.Fatalf("expected handler to see key %q, got %q", key.String(), got.String())
	}
}

func TestInputChainHandledShortCircuitsComponentRouting(t *testing.T) {
	app := NewApplication()
	c := &mockComponent{}
	app.AddComponent(c)

	// Resetting because AddComponent may have triggered Focus side
	// effects but updateCount tracks Update calls only.
	if c.updateCount != 0 {
		t.Fatalf("expected fresh component to have updateCount=0, got %d", c.updateCount)
	}

	var seen []tea.Msg
	app.Use(recordingHandler(&seen, true, nil))

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	if len(seen) != 1 {
		t.Fatalf("expected handler to observe message, got %d observations", len(seen))
	}
	if c.updateCount != 0 {
		t.Fatalf("expected focused component to be skipped when handler consumes; updateCount=%d", c.updateCount)
	}
}

func TestInputChainUnhandledFallsThroughToComponent(t *testing.T) {
	app := NewApplication()
	c := &mockComponent{}
	app.AddComponent(c)

	var seen []tea.Msg
	app.Use(recordingHandler(&seen, false, nil))

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	if len(seen) != 1 {
		t.Fatalf("expected handler to observe message, got %d observations", len(seen))
	}
	if c.updateCount != 1 {
		t.Fatalf("expected focused component to receive message when handler does not consume; updateCount=%d", c.updateCount)
	}
}

func TestInputChainOrderingFirstUnhandledSecondConsumes(t *testing.T) {
	app := NewApplication()
	c := &mockComponent{}
	app.AddComponent(c)

	var order []int
	app.Use(func(msg tea.Msg) (bool, tea.Cmd) {
		order = append(order, 1)
		return false, nil
	})
	app.Use(func(msg tea.Msg) (bool, tea.Cmd) {
		order = append(order, 2)
		return true, nil
	})
	// A third handler that should never run because handler 2 consumes.
	app.Use(func(msg tea.Msg) (bool, tea.Cmd) {
		order = append(order, 3)
		return false, nil
	})

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Fatalf("expected handlers to run in order [1,2] then short-circuit, got %v", order)
	}
	if c.updateCount != 0 {
		t.Fatalf("expected component to be skipped when chain consumes; updateCount=%d", c.updateCount)
	}
}

// sentinelMsg is a unique message type returned by a handler-produced
// tea.Cmd so the test can confirm the cmd was actually placed in the
// batched result.
type sentinelMsg struct{}

func TestInputChainBatchesHandlerCommand(t *testing.T) {
	app := NewApplication()
	c := &mockComponent{}
	app.AddComponent(c)

	sentinel := func() tea.Msg { return sentinelMsg{} }
	app.Use(func(msg tea.Msg) (bool, tea.Cmd) {
		return false, sentinel
	})

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if cmd == nil {
		t.Fatal("expected non-nil batched command from chain")
	}

	if !batchContainsSentinel(cmd) {
		t.Fatal("expected batched command sequence to include sentinelMsg")
	}
}

func TestInputChainBatchesHandlerCommandWhenConsumed(t *testing.T) {
	app := NewApplication()
	c := &mockComponent{}
	app.AddComponent(c)

	sentinel := func() tea.Msg { return sentinelMsg{} }
	app.Use(func(msg tea.Msg) (bool, tea.Cmd) {
		return true, sentinel
	})

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if cmd == nil {
		t.Fatal("expected non-nil batched command from consuming handler")
	}

	if !batchContainsSentinel(cmd) {
		t.Fatal("expected consumed-chain batched command sequence to include sentinelMsg")
	}
	if c.updateCount != 0 {
		t.Fatalf("expected component to be skipped when handler consumes; updateCount=%d", c.updateCount)
	}
}

func TestInputChainDeregisterStopsHandler(t *testing.T) {
	app := NewApplication()
	c := &mockComponent{}
	app.AddComponent(c)

	var seen []tea.Msg
	deregister := app.Use(recordingHandler(&seen, false, nil))

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if len(seen) != 1 {
		t.Fatalf("expected 1 observation before deregister, got %d", len(seen))
	}

	deregister()

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if len(seen) != 1 {
		t.Fatalf("expected handler to stop firing after deregister, got %d observations", len(seen))
	}

	// Calling deregister twice must be a no-op (no panic, no extra
	// removals that could corrupt the slice).
	deregister()

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if len(seen) != 1 {
		t.Fatalf("expected handler to remain deregistered after second deregister call, got %d observations", len(seen))
	}
}

func TestInputChainDeregisterMiddleHandler(t *testing.T) {
	app := NewApplication()
	c := &mockComponent{}
	app.AddComponent(c)

	var seen []int
	app.Use(func(msg tea.Msg) (bool, tea.Cmd) {
		seen = append(seen, 1)
		return false, nil
	})
	deregisterMiddle := app.Use(func(msg tea.Msg) (bool, tea.Cmd) {
		seen = append(seen, 2)
		return false, nil
	})
	app.Use(func(msg tea.Msg) (bool, tea.Cmd) {
		seen = append(seen, 3)
		return false, nil
	})

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if len(seen) != 3 || seen[0] != 1 || seen[1] != 2 || seen[2] != 3 {
		t.Fatalf("expected initial order [1,2,3], got %v", seen)
	}

	seen = nil
	deregisterMiddle()

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if len(seen) != 2 || seen[0] != 1 || seen[1] != 3 {
		t.Fatalf("expected order after deregistering middle to be [1,3], got %v", seen)
	}
}

func TestInputChainNilHandlerNoop(t *testing.T) {
	app := NewApplication()
	c := &mockComponent{}
	app.AddComponent(c)

	deregister := app.Use(nil)
	if deregister == nil {
		t.Fatal("expected Use(nil) to return a non-nil deregister function")
	}
	deregister() // must not panic

	// Component still receives messages because no real handlers exist.
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if c.updateCount != 1 {
		t.Fatalf("expected component to receive message with only nil-handler registration; updateCount=%d", c.updateCount)
	}
}

// batchContainsSentinel inspects a tea.Cmd produced by tea.Batch and
// reports whether evaluating any of the underlying commands yields a
// sentinelMsg. tea.Batch returns a tea.Cmd that produces a tea.BatchMsg
// — a []tea.Cmd — when called. We walk that recursively so handlers
// that return their own tea.Batch values are also covered.
func batchContainsSentinel(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	msg := cmd()
	switch m := msg.(type) {
	case sentinelMsg:
		return true
	case tea.BatchMsg:
		for _, c := range m {
			if batchContainsSentinel(c) {
				return true
			}
		}
	}
	return false
}
