package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// StatusBar displays status information and keybindings at the bottom of the screen.
// It shows a status message on the left and keybinding hints on the right, with
// automatic truncation when the terminal is narrow.
//
// Visual styling changes based on focus state:
//   - Focused: Inverted colors (black on white)
//   - Unfocused: Dimmed text
//
// Example usage:
//
//	statusBar := tui.NewStatusBar()
//	statusBar.SetMessage("Processing...")
//	// Later: statusBar.SetMessage("Complete!")
type StatusBar struct {
	width   int
	message string
	focused bool
}

// NewStatusBar creates a new status bar with the default message "Ready".
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

// View renders the status bar as a single line with the status message on the left
// and keybinding hints on the right. The message is automatically truncated with "..."
// if the terminal is too narrow. Returns an empty string if width is zero.
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

// SetMessage updates the status message displayed on the left side of the status bar.
// The message will be automatically truncated with "..." if the terminal is too narrow
// to display both the message and keybinding hints.
func (s *StatusBar) SetMessage(msg string) {
	s.message = msg
}
