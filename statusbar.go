package tui

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
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
	width     int
	message   string
	focused   bool
	textColor string
	hintColor string
}

// StatusBarOption configures a StatusBar.
type StatusBarOption func(*StatusBar)

// WithStatusBarDesignTokens applies design-system colors to the status bar.
func WithStatusBarDesignTokens(tokens *design.DesignTokens) StatusBarOption {
	return func(s *StatusBar) {
		s.applyDesignTokens(tokens)
	}
}

// WithStatusBarTheme applies a named design-system theme.
func WithStatusBarTheme(theme string) StatusBarOption {
	return func(s *StatusBar) {
		s.applyDesignTokens(designTokensForTheme(theme))
	}
}

// NewStatusBar creates a new status bar with the default message "Ready".
func NewStatusBar(opts ...StatusBarOption) *StatusBar {
	s := &StatusBar{
		message:   "Ready",
		textColor: "\033[2m",
		hintColor: "\033[2m",
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
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
	right := s.hintColor + "Tab: Focus • q: Quit\033[0m"

	// Calculate spacing
	spacing := s.width - len(left) - len(stripANSI(right))
	if spacing < 0 {
		spacing = 0
		// Truncate left message if needed
		maxLeft := s.width - len(stripANSI(right)) - 3
		if maxLeft < 0 {
			maxLeft = 0
		}
		if len(left) > maxLeft {
			left = left[:maxLeft] + "..."
		}
	}

	// Build status bar
	line := left + strings.Repeat(" ", spacing) + right

	// Add styling based on focus
	if s.focused {
		return fmt.Sprintf("\033[7m%s\033[0m\n", line) // Inverted colors when focused
	}
	return fmt.Sprintf("%s%s\033[0m\n", s.textColor, line)
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

func (s *StatusBar) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}
	foreground := ansiColorFromHex(tokens.Color)
	accent := ansiColorFromHex(tokens.Accent)
	if foreground != "" {
		s.textColor = foreground
	}
	if accent != "" {
		s.hintColor = accent
	}
}
