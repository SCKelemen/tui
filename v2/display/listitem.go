package display

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ListItem renders a generic list item row.
type ListItem struct {
	label       string
	width       int
	icon        string
	description string
	trailing    string
	selected    bool
	disabled    bool
	focused     bool

	designTokens *design.DesignTokens
}

// ListItemOption configures a ListItem.
type ListItemOption func(*ListItem)

// WithListItemWidth sets the list item width.
func WithListItemWidth(width int) ListItemOption {
	return func(l *ListItem) {
		if width >= 0 {
			l.width = width
		}
	}
}

// WithListItemIcon sets the optional icon.
func WithListItemIcon(icon string) ListItemOption {
	return func(l *ListItem) { l.icon = icon }
}

// WithListItemDescription sets the optional description.
func WithListItemDescription(description string) ListItemOption {
	return func(l *ListItem) { l.description = description }
}

// WithListItemTrailing sets the optional trailing text.
func WithListItemTrailing(trailing string) ListItemOption {
	return func(l *ListItem) { l.trailing = trailing }
}

// WithListItemSelected sets selected state.
func WithListItemSelected(selected bool) ListItemOption {
	return func(l *ListItem) { l.selected = selected }
}

// WithListItemDisabled sets disabled state.
func WithListItemDisabled(disabled bool) ListItemOption {
	return func(l *ListItem) { l.disabled = disabled }
}

// WithListItemDesignTokens applies design tokens.
func WithListItemDesignTokens(tokens *design.DesignTokens) ListItemOption {
	return func(l *ListItem) {
		if tokens != nil {
			l.designTokens = tokens
		}
	}
}

// NewListItem creates a new ListItem component.
func NewListItem(label string, opts ...ListItemOption) *ListItem {
	l := &ListItem{
		label:        label,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// Init initializes the component.
func (l *ListItem) Init() tea.Cmd { return nil }

// Update handles Bubble Tea messages.
func (l *ListItem) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		if l.width == 0 {
			l.width = m.Width
		}
	}
	return l, nil
}

// View renders the list item.
func (l *ListItem) View() string {
	main := l.mainLine()
	if strings.TrimSpace(l.description) == "" {
		return l.applyState(main)
	}

	descColor := style.ANSIDim
	if l.designTokens != nil {
		if c := style.Fg(l.designTokens.MutedColor); c != "" {
			descColor = c
		}
	}
	desc := descColor + strings.TrimSpace(l.description) + style.ANSIReset
	if l.width > 0 {
		desc = style.Truncate(desc, l.width, "…")
	}

	return l.applyState(main) + "\n" + l.applyState("  "+desc)
}

// Focus marks the component as focused.
func (l *ListItem) Focus() { l.focused = true }

// Blur marks the component as unfocused.
func (l *ListItem) Blur() { l.focused = false }

// Focused reports whether the component is focused.
func (l *ListItem) Focused() bool { return l.focused }

func (l *ListItem) mainLine() string {
	left := strings.TrimSpace(l.label)
	if icon := strings.TrimSpace(l.icon); icon != "" {
		left = icon + " " + left
	}

	if l.width <= 0 || strings.TrimSpace(l.trailing) == "" {
		return left
	}

	trail := strings.TrimSpace(l.trailing)
	leftWidth := style.StringWidth(left)
	trailWidth := style.StringWidth(trail)
	if leftWidth+1+trailWidth >= l.width {
		return style.Truncate(left+" "+trail, l.width, "…")
	}

	gap := l.width - leftWidth - trailWidth
	return left + strings.Repeat(" ", gap) + trail
}

func (l *ListItem) applyState(text string) string {
	out := text
	if l.disabled {
		out = style.ANSIDim + out + style.ANSIReset
	}
	if l.selected {
		out = style.ANSIInverse + out + style.ANSIReset
	}
	return out
}

var _ tui.Component = (*ListItem)(nil)
