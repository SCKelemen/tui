package input

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// VisibilityToggleMsg is emitted when visibility is toggled.
type VisibilityToggleMsg struct {
	Visible bool
}

// VisibilityToggleOption configures a VisibilityToggle.
type VisibilityToggleOption func(*VisibilityToggle)

// WithVisibilityToggleDesignTokens applies design-system colors.
func WithVisibilityToggleDesignTokens(tokens *design.DesignTokens) VisibilityToggleOption {
	return func(v *VisibilityToggle) {
		v.applyDesignTokens(tokens)
	}
}

// WithVisibilityToggleExpanded sets the initial expanded state.
func WithVisibilityToggleExpanded(expanded bool) VisibilityToggleOption {
	return func(v *VisibilityToggle) {
		v.expanded = expanded
	}
}

// WithVisibilityToggleWidth sets a preferred render width.
func WithVisibilityToggleWidth(width int) VisibilityToggleOption {
	return func(v *VisibilityToggle) {
		if width > 0 {
			v.width = width
		}
	}
}

// VisibilityToggle is a simple expandable show/hide toggle.
type VisibilityToggle struct {
	label   string
	content string

	expanded bool
	focused  bool
	width    int

	designTokens *design.DesignTokens
	accentColor  string
	mutedColor   string
}

// NewVisibilityToggle creates a new VisibilityToggle component.
func NewVisibilityToggle(label string, content string, opts ...VisibilityToggleOption) *VisibilityToggle {
	v := &VisibilityToggle{
		label:        strings.TrimSpace(label),
		content:      content,
		expanded:     false,
		focused:      false,
		width:        0,
		designTokens: design.DefaultTheme(),
		accentColor:  style.ANSICyan,
		mutedColor:   style.ANSIDim,
	}

	for _, opt := range opts {
		opt(v)
	}

	if v.label == "" {
		v.label = "Details"
	}

	v.applyDesignTokens(v.designTokens)
	return v
}

// Init initializes the component.
func (v *VisibilityToggle) Init() tea.Cmd {
	return nil
}

// Update handles keyboard and window size messages.
func (v *VisibilityToggle) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if v.width <= 0 {
			v.width = msg.Width
		}
		return v, nil
	case tea.KeyMsg:
		if !v.focused {
			return v, nil
		}

		switch msg.String() {
		case "enter", " ":
			v.expanded = !v.expanded
			visible := v.expanded
			return v, func() tea.Msg {
				return VisibilityToggleMsg{Visible: visible}
			}
		}
	}

	return v, nil
}

// View renders the collapsed or expanded toggle.
func (v *VisibilityToggle) View() string {
	icon := "▶"
	lineColor := v.mutedColor
	if v.expanded {
		icon = "▼"
		lineColor = v.accentColor
	}

	header := lineColor + icon + " " + v.label + style.ANSIReset
	header = v.fitWidth(header)
	if !v.expanded {
		return header
	}

	contentLines := strings.Split(v.content, "\n")
	for i, line := range contentLines {
		contentLines[i] = v.fitWidth("  " + line)
	}

	if len(contentLines) == 0 {
		return header
	}
	return header + "\n" + strings.Join(contentLines, "\n")
}

// Focus marks the component as focused.
func (v *VisibilityToggle) Focus() {
	v.focused = true
}

// Blur marks the component as unfocused.
func (v *VisibilityToggle) Blur() {
	v.focused = false
}

// Focused reports whether the component currently has focus.
func (v *VisibilityToggle) Focused() bool {
	return v.focused
}

func (v *VisibilityToggle) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}
	v.designTokens = tokens

	if accent := style.ANSIColorFromHex(tokens.Accent); accent != "" {
		v.accentColor = accent
	}
	if muted := style.ANSIColorFromHex(tokens.MutedColor); muted != "" {
		v.mutedColor = muted
	} else if color := style.ANSIColorFromHex(tokens.Color); color != "" {
		v.mutedColor = color
	}
}

func (v *VisibilityToggle) fitWidth(line string) string {
	if v.width <= 0 {
		return line
	}
	if style.StringWidth(line) <= v.width {
		return style.Pad(line, v.width)
	}
	return style.Truncate(line, v.width, "…")
}

var _ tui.Component = (*VisibilityToggle)(nil)
