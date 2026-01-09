// Package tui provides a comprehensive Terminal User Interface framework
// for building sophisticated CLI applications with modern UX patterns.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Application represents the main TUI application
type Application struct {
	width      int
	height     int
	components []Component
	focused    int // Index of currently focused component
}

// Component is the interface all TUI components must implement
type Component interface {
	// Init initializes the component
	Init() tea.Cmd

	// Update handles messages and updates component state
	Update(msg tea.Msg) (Component, tea.Cmd)

	// View renders the component
	View() string

	// Focus is called when this component receives focus
	Focus()

	// Blur is called when this component loses focus
	Blur()

	// Focused returns whether this component is currently focused
	Focused() bool
}

// NewApplication creates a new TUI application
func NewApplication() *Application {
	return &Application{
		components: make([]Component, 0),
		focused:    -1,
	}
}

// AddComponent adds a component to the application
func (a *Application) AddComponent(c Component) {
	a.components = append(a.components, c)
	if a.focused == -1 && len(a.components) > 0 {
		a.focused = 0
		a.components[0].Focus()
	}
}

// Init initializes the application
func (a *Application) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, c := range a.components {
		cmds = append(cmds, c.Init())
	}
	return tea.Batch(cmds...)
}

// Update handles messages
func (a *Application) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "tab":
			// Cycle focus forward
			return a, a.focusNext()
		case "shift+tab":
			// Cycle focus backward
			return a, a.focusPrev()
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Window size messages should go to all components
		for i, c := range a.components {
			var cmd tea.Cmd
			a.components[i], cmd = c.Update(msg)
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)
	}

	// Check if this is a tick message (these need to go to all components for animations)
	// We do this by checking the message type name
	if isTickMessage(msg) {
		// Broadcast tick messages to all components for animations
		for i, c := range a.components {
			var cmd tea.Cmd
			a.components[i], cmd = c.Update(msg)
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)
	}

	// Pass other messages only to focused component
	if a.focused >= 0 && a.focused < len(a.components) {
		var cmd tea.Cmd
		a.components[a.focused], cmd = a.components[a.focused].Update(msg)
		return a, cmd
	}

	return a, nil
}

// isTickMessage checks if a message is a tick-related message that should be broadcast
func isTickMessage(msg tea.Msg) bool {
	// Check for specific tick message types
	switch msg.(type) {
	case activityBarTickMsg:
		return true
	case toolBlockTickMsg:
		return true
	default:
		// Check the type name for any message containing "tick" or "Tick"
		typeName := fmt.Sprintf("%T", msg)
		return strings.Contains(typeName, "tick") || strings.Contains(typeName, "Tick")
	}
}

// View renders the application
func (a *Application) View() string {
	if len(a.components) == 0 {
		return "No components"
	}

	var view string
	for _, c := range a.components {
		view += c.View()
	}
	return view
}

// focusNext moves focus to the next component
func (a *Application) focusNext() tea.Cmd {
	if len(a.components) == 0 {
		return nil
	}

	// Blur current
	if a.focused >= 0 {
		a.components[a.focused].Blur()
	}

	// Focus next
	a.focused = (a.focused + 1) % len(a.components)
	a.components[a.focused].Focus()

	return nil
}

// focusPrev moves focus to the previous component
func (a *Application) focusPrev() tea.Cmd {
	if len(a.components) == 0 {
		return nil
	}

	// Blur current
	if a.focused >= 0 {
		a.components[a.focused].Blur()
	}

	// Focus previous
	a.focused--
	if a.focused < 0 {
		a.focused = len(a.components) - 1
	}
	a.components[a.focused].Focus()

	return nil
}
