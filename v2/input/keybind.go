package input

import (
	"sort"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// Keybind describes a keyboard shortcut, what it does, and how to execute it.
type Keybind struct {
	Key         string
	Description string
	Action      func() tea.Cmd
	Category    string
}

// KeybindManager stores keybindings and dispatches actions.
type KeybindManager struct {
	bindings  map[string]*Keybind
	suspended bool
	mu        sync.RWMutex
}

// NewKeybindManager creates a new keybind manager.
func NewKeybindManager() *KeybindManager {
	return &KeybindManager{
		bindings: make(map[string]*Keybind),
	}
}

// Register adds or replaces a keybind for the key.
func (km *KeybindManager) Register(kb Keybind) {
	km.mu.Lock()
	defer km.mu.Unlock()

	copyKB := kb
	km.bindings[kb.Key] = &copyKB
}

// Unregister removes a keybind by key.
func (km *KeybindManager) Unregister(key string) {
	km.mu.Lock()
	defer km.mu.Unlock()

	delete(km.bindings, key)
}

// Suspend temporarily disables all keybind handling.
func (km *KeybindManager) Suspend() {
	km.mu.Lock()
	defer km.mu.Unlock()

	km.suspended = true
}

// Resume re-enables keybind handling.
func (km *KeybindManager) Resume() {
	km.mu.Lock()
	defer km.mu.Unlock()

	km.suspended = false
}

// IsSuspended reports whether keybind handling is currently suspended.
func (km *KeybindManager) IsSuspended() bool {
	km.mu.RLock()
	defer km.mu.RUnlock()

	return km.suspended
}

// Handle executes the action for a key if registered and enabled.
func (km *KeybindManager) Handle(key string) (tea.Cmd, bool) {
	km.mu.RLock()
	if km.suspended {
		km.mu.RUnlock()
		return nil, false
	}

	kb, ok := km.bindings[key]
	if !ok {
		km.mu.RUnlock()
		return nil, false
	}
	action := kb.Action
	km.mu.RUnlock()

	if action == nil {
		return nil, true
	}

	return action(), true
}

// List returns all registered keybinds sorted by category and then key.
func (km *KeybindManager) List() []Keybind {
	km.mu.RLock()
	result := make([]Keybind, 0, len(km.bindings))
	for _, kb := range km.bindings {
		result = append(result, *kb)
	}
	km.mu.RUnlock()

	sort.Slice(result, func(i, j int) bool {
		if result[i].Category == result[j].Category {
			return result[i].Key < result[j].Key
		}
		return result[i].Category < result[j].Category
	})

	return result
}

// HandleMsg converts a tea.KeyMsg into a key string and handles it.
func (km *KeybindManager) HandleMsg(msg tea.KeyMsg) (tea.Cmd, bool) {
	return km.Handle(msg.String())
}
