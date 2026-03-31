package display

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// DividerOrientation controls whether a Divider is horizontal or vertical.
type DividerOrientation int

const (
	// Horizontal renders a horizontal divider line.
	Horizontal DividerOrientation = iota
	// Vertical renders a vertical divider line.
	Vertical
)

// Divider renders a horizontal or vertical divider with optional label.
type Divider struct {
	width       int
	orientation DividerOrientation
	label       string
	character   rune
	focused     bool

	designTokens *design.DesignTokens
	color        string

	widthExplicit bool
}

// DividerOption configures a Divider.
type DividerOption func(*Divider)

// WithDividerWidth sets the divider width.
func WithDividerWidth(width int) DividerOption {
	return func(d *Divider) {
		if width >= 0 {
			d.width = width
			d.widthExplicit = true
		}
	}
}

// WithDividerOrientation sets the divider orientation.
func WithDividerOrientation(orientation DividerOrientation) DividerOption {
	return func(d *Divider) {
		d.orientation = orientation
	}
}

// WithDividerLabel sets an optional divider label.
func WithDividerLabel(label string) DividerOption {
	return func(d *Divider) {
		d.label = strings.TrimSpace(label)
	}
}

// WithDividerCharacter sets the divider line rune.
func WithDividerCharacter(ch rune) DividerOption {
	return func(d *Divider) {
		if ch != 0 {
			d.character = ch
		}
	}
}

// WithDividerDesignTokens applies design tokens.
func WithDividerDesignTokens(tokens *design.DesignTokens) DividerOption {
	return func(d *Divider) {
		d.applyDesignTokens(tokens)
	}
}

// NewDivider creates a Divider component.
func NewDivider(opts ...DividerOption) *Divider {
	d := &Divider{
		width:        40,
		orientation:  Horizontal,
		character:    '─',
		designTokens: design.DefaultTheme(),
	}

	d.applyDesignTokens(d.designTokens)

	for _, opt := range opts {
		opt(d)
	}

	if d.orientation == Vertical && d.character == '─' {
		d.character = '│'
	}
	if d.orientation == Horizontal && d.character == '│' {
		d.character = '─'
	}

	return d
}

// Init initializes the component.
func (d *Divider) Init() tea.Cmd { return nil }

// Update handles Bubble Tea messages.
func (d *Divider) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		if !d.widthExplicit {
			d.width = m.Width
		}
	}
	return d, nil
}

// View renders the divider.
func (d *Divider) View() string {
	if d.width <= 0 {
		return ""
	}

	if d.orientation == Vertical {
		return d.renderVertical()
	}
	return d.renderHorizontal()
}

// Focus marks the component as focused.
func (d *Divider) Focus() { d.focused = true }

// Blur marks the component as unfocused.
func (d *Divider) Blur() { d.focused = false }

// Focused reports whether the component is focused.
func (d *Divider) Focused() bool { return d.focused }

func (d *Divider) renderHorizontal() string {
	ch := d.character
	if ch == 0 {
		ch = '─'
	}

	label := strings.TrimSpace(d.label)
	if label == "" {
		line := strings.Repeat(string(ch), d.width)
		return d.applyColor(line)
	}

	middle := " " + label + " "
	middleWidth := style.StringWidth(middle)
	if middleWidth >= d.width {
		truncated := style.Truncate(middle, d.width, "…")
		return d.applyColor(truncated)
	}

	remaining := d.width - middleWidth
	left := remaining / 2
	right := remaining - left
	line := strings.Repeat(string(ch), left) + middle + strings.Repeat(string(ch), right)
	return d.applyColor(line)
}

func (d *Divider) renderVertical() string {
	ch := d.character
	if ch == 0 {
		ch = '│'
	}

	line := d.applyColor(string(ch))
	return strings.Repeat(line+"\n", d.width-1) + line
}

func (d *Divider) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}

	d.designTokens = tokens

	if value := strings.TrimSpace(tokens.BorderSubtle); value != "" {
		d.color = value
		return
	}
	if value := strings.TrimSpace(tokens.Accent); value != "" {
		d.color = value
	}
}

func (d *Divider) applyColor(text string) string {
	fg := style.Fg(d.color)
	if fg == "" {
		return text
	}
	return fg + text + style.ANSIReset
}

var _ tui.Component = (*Divider)(nil)
