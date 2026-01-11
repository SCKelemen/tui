// Package tui provides a comprehensive Terminal User Interface framework
// for building sophisticated CLI applications with modern UX patterns.
//
// The framework is built on top of Bubble Tea and provides high-level components
// for common UI patterns including interactive dashboards, file explorers, command
// palettes, status bars, and modal dialogs.
//
// Key features:
//   - Interactive dashboards with keyboard navigation and drill-down modals
//   - StatCards with sparklines, change indicators, and focus states
//   - Command palettes with fuzzy search
//   - File explorers with tree navigation
//   - Status bars with keybinding hints
//   - Modal dialogs for confirmations and details
//   - Full keyboard navigation with vim-style bindings
//   - Responsive layouts that adapt to terminal size
//   - Theme support via design tokens
//
// Architecture:
//
// All components implement the Component interface, which follows the Bubble Tea
// pattern with Init/Update/View methods plus Focus/Blur/Focused for focus management.
//
// Components can be composed together to build complex UIs. The Application type
// provides a container for managing multiple components with automatic focus cycling.
//
// Example usage:
//
//	// Create components
//	dashboard := tui.NewDashboard(
//	    tui.WithDashboardTitle("Metrics"),
//	    tui.WithCards(cpuCard, memCard),
//	)
//	dashboard.Focus()
//
//	statusBar := tui.NewStatusBar()
//	statusBar.SetMessage("Ready")
//
//	// Run with Bubble Tea
//	p := tea.NewProgram(model{
//	    dashboard: dashboard,
//	    statusBar: statusBar,
//	}, tea.WithAltScreen())
//	p.Run()
//
// For detailed component documentation, see the individual component types.
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

// FocusComponent focuses a specific component by index, blurring the currently focused one
func (a *Application) FocusComponent(index int) {
	if index < 0 || index >= len(a.components) {
		return
	}

	// Blur current
	if a.focused >= 0 && a.focused < len(a.components) {
		a.components[a.focused].Blur()
	}

	// Focus new
	a.focused = index
	a.components[index].Focus()
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
