package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// InputHandler is a function registered on an Application via
// Application.Use. Handlers run in registration order before component
// routing in Application.Update. A handler can short-circuit further
// routing by returning handled=true; any non-nil tea.Cmd is collected
// and batched with the eventual return value of Update.
//
// This mirrors OpenTUI's chained input-handler model: a small,
// composable pipeline that sits in front of component dispatch and
// lets application-wide concerns (global hotkeys, focus management,
// command palettes, recorders, etc.) observe or consume messages
// before any focused component sees them.
type InputHandler func(msg tea.Msg) (handled bool, cmd tea.Cmd)

// registeredHandler is the internal storage entry for a handler. It
// holds both the handler function and a stable id so that the
// deregister closure returned from Use can locate the entry even after
// other handlers have been added or removed.
type registeredHandler struct {
	id int
	fn InputHandler
}

// nextHandlerID returns the next id to assign to a registered handler.
// Application is single-goroutine like the rest of bubbletea's Update
// path, so a plain integer counter is sufficient — no locking required.
func (a *Application) nextHandlerID() int {
	// Use len(handlers) + a monotonically increasing field would also
	// work; using a counter on the slice avoids adding another field
	// to Application. We derive a fresh id by scanning for the max
	// existing id and adding one. Handler counts are expected to be
	// small (tens at most) so the O(n) scan is fine.
	maxID := 0
	for _, h := range a.handlers {
		if h.id > maxID {
			maxID = h.id
		}
	}
	return maxID + 1
}

// Use registers an InputHandler with the Application. Handlers are
// invoked in registration order before component routing in
// Application.Update. The returned function deregisters the handler in
// O(n); calling it more than once is a no-op.
//
// A nil handler is a no-op: Use returns a no-op deregister function and
// does not add anything to the chain.
func (a *Application) Use(h InputHandler) (deregister func()) {
	if h == nil {
		return func() {}
	}

	id := a.nextHandlerID()
	a.handlers = append(a.handlers, registeredHandler{id: id, fn: h})

	deregistered := false
	return func() {
		if deregistered {
			return
		}
		deregistered = true
		for i := range a.handlers {
			if a.handlers[i].id == id {
				a.handlers = append(a.handlers[:i], a.handlers[i+1:]...)
				return
			}
		}
	}
}

// runInputChain executes the registered InputHandler chain in
// registration order for msg. It returns the collected non-nil commands
// and a flag indicating whether any handler consumed the message. Once a
// handler returns handled=true, no later handlers are invoked — their
// turn is skipped and consumed=true short-circuits component routing in
// Application.Update.
func (a *Application) runInputChain(msg tea.Msg) (cmds []tea.Cmd, consumed bool) {
	if len(a.handlers) == 0 {
		return nil, false
	}
	// Snapshot the slice header so handlers that call Use or the
	// returned deregister function during iteration do not skip or
	// double-visit entries. Re-resolving by id keeps the iteration
	// stable even if mutations reslice a.handlers underneath us.
	snapshot := make([]registeredHandler, len(a.handlers))
	copy(snapshot, a.handlers)
	for _, h := range snapshot {
		if h.fn == nil {
			continue
		}
		handled, cmd := h.fn(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		if handled {
			return cmds, true
		}
	}
	return cmds, false
}
