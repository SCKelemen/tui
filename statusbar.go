package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// StatusBar displays status information and keybindings at the bottom of the screen
type StatusBar struct {
	width   int
	message string
	focused bool
}

// NewStatusBar creates a new status bar
func NewStatusBar() *StatusBar {
	return &StatusBar{
		message: "Ready",
	}
}

// Init initializes the status bar
func (s *StatusBar) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *StatusBar) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
	}
	return s, nil
}

// View renders the status bar
func (s *StatusBar) View() string {
	if s.width == 0 {
		return ""
	}

	// Status message on the left
	left := s.message

	// Keybindings on the right
	right := "Tab: Focus • q: Quit"

	// Calculate spacing
	spacing := s.width - len(left) - len(right)
	if spacing < 0 {
		spacing = 0
		// Truncate left message if needed
		if len(left) > s.width-len(right)-3 {
			left = left[:s.width-len(right)-3] + "..."
		}
	}

	// Build status bar
	line := left + strings.Repeat(" ", spacing) + right

	// Add styling based on focus
	if s.focused {
		return fmt.Sprintf("\033[7m%s\033[0m\n", line) // Inverted colors when focused
	}
	return fmt.Sprintf("\033[2m%s\033[0m\n", line) // Dimmed when not focused
}

// Focus is called when this component receives focus
func (s *StatusBar) Focus() {
	s.focused = true
}

// Blur is called when this component loses focus
func (s *StatusBar) Blur() {
	s.focused = false
}

// Focused returns whether this component is currently focused
func (s *StatusBar) Focused() bool {
	return s.focused
}

// SetMessage updates the status message
func (s *StatusBar) SetMessage(msg string) {
	s.message = msg
}
