package display

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	tea "github.com/charmbracelet/bubbletea"
)

// RawAnsiOption configures a RawAnsi component.
type RawAnsiOption func(*RawAnsi)

// WithRawAnsiWidth sets the preferred render width.
func WithRawAnsiWidth(width int) RawAnsiOption {
	return func(r *RawAnsi) {
		if width > 0 {
			r.width = width
		}
	}
}

// WithRawAnsiDesignTokens accepts design tokens for API symmetry.
func WithRawAnsiDesignTokens(tokens *design.DesignTokens) RawAnsiOption {
	return func(r *RawAnsi) {
		r.designTokens = tokens
	}
}

// RawAnsi is a fast-path component for pre-rendered ANSI text.
type RawAnsi struct {
	lines        []string
	width        int
	focused      bool
	designTokens *design.DesignTokens
}

// NewRawAnsi creates a new RawAnsi renderer.
func NewRawAnsi(lines []string, opts ...RawAnsiOption) *RawAnsi {
	r := &RawAnsi{
		lines:   append([]string(nil), lines...),
		width:   0,
		focused: false,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Init initializes the component.
func (r *RawAnsi) Init() tea.Cmd { return nil }

// Update handles Bubble Tea messages.
func (r *RawAnsi) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if r.width <= 0 {
			r.width = msg.Width
		}
	}
	return r, nil
}

// View joins the raw ANSI lines and truncates to width if configured.
func (r *RawAnsi) View() string {
	if len(r.lines) == 0 {
		return ""
	}

	if r.width <= 0 {
		return strings.Join(r.lines, "\n")
	}

	out := make([]string, len(r.lines))
	for i, line := range r.lines {
		out[i] = rawAnsiTruncateLine(line, r.width)
	}
	return strings.Join(out, "\n")
}

// Focus marks the component focused.
func (r *RawAnsi) Focus() { r.focused = true }

// Blur marks the component unfocused.
func (r *RawAnsi) Blur() { r.focused = false }

// Focused reports focus state.
func (r *RawAnsi) Focused() bool { return r.focused }

func rawAnsiTruncateLine(line string, width int) string {
	if width <= 0 {
		return line
	}
	runes := []rune(line)
	if len(runes) <= width {
		return line
	}
	return string(runes[:width])
}

var _ tui.Component = (*RawAnsi)(nil)
