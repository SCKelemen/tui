package display

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	noSelectOpenANSI  = "\x1b[?2026h"
	noSelectCloseANSI = "\x1b[?2026l"
)

// NoSelect wraps content to mark it as non-selectable.
type NoSelect struct {
	content      string
	width        int
	fromLeftEdge bool
	focused      bool

	designTokens *design.DesignTokens
}

// NoSelectOption configures NoSelect.
type NoSelectOption func(*NoSelect)

// NewNoSelect creates a new NoSelect wrapper component.
func NewNoSelect(content string, opts ...NoSelectOption) *NoSelect {
	n := &NoSelect{
		content:      content,
		width:        0,
		fromLeftEdge: false,
	}

	for _, opt := range opts {
		opt(n)
	}

	return n
}

// WithNoSelectWidth sets the render width for wrapped content.
func WithNoSelectWidth(width int) NoSelectOption {
	return func(n *NoSelect) {
		if width >= 0 {
			n.width = width
		}
	}
}

// WithNoSelectFromLeftEdge toggles left-edge anchoring behavior.
func WithNoSelectFromLeftEdge(fromLeftEdge bool) NoSelectOption {
	return func(n *NoSelect) {
		n.fromLeftEdge = fromLeftEdge
	}
}

// WithNoSelectDesignTokens applies design-system tokens.
func WithNoSelectDesignTokens(tokens *design.DesignTokens) NoSelectOption {
	return func(n *NoSelect) {
		n.designTokens = tokens
	}
}

// Init initializes the component.
func (n *NoSelect) Init() tea.Cmd { return nil }

// Update handles Bubble Tea messages.
func (n *NoSelect) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if n.width == 0 {
			n.width = msg.Width
		}
	}
	return n, nil
}

// View renders the content wrapped in non-selectable ANSI markers.
func (n *NoSelect) View() string {
	content := n.content
	if n.width > 0 {
		content = n.fitContent(content)
	}
	return noSelectOpenANSI + content + noSelectCloseANSI
}

// Focus marks the component focused.
func (n *NoSelect) Focus() { n.focused = true }

// Blur marks the component unfocused.
func (n *NoSelect) Blur() { n.focused = false }

// Focused reports focus state.
func (n *NoSelect) Focused() bool { return n.focused }

func (n *NoSelect) fitContent(content string) string {
	if n.width <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if lipgloss.Width(line) > n.width {
			lines[i] = style.Truncate(line, n.width, "…")
			continue
		}

		if n.fromLeftEdge {
			lines[i] = style.Pad(line, n.width)
		} else {
			pad := n.width - lipgloss.Width(line)
			if pad < 0 {
				pad = 0
			}
			lines[i] = strings.Repeat(" ", pad) + line
		}
	}

	return strings.Join(lines, "\n")
}

var _ tui.Component = (*NoSelect)(nil)
