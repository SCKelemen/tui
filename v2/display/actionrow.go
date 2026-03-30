package display

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ActionItem is a single action/keybinding hint in the action row.
type ActionItem struct {
	Key   string
	Label string
	Color string
}

// ActionRow renders a horizontal row of action keybinding hints.
type ActionRow struct {
	items        []ActionItem
	separator    string
	align        string
	width        int
	focused      bool
	designTokens *design.DesignTokens
}

// ActionRowOption configures an ActionRow.
type ActionRowOption func(*ActionRow)

// WithActionRowDesignTokens applies design-system tokens.
func WithActionRowDesignTokens(tokens *design.DesignTokens) ActionRowOption {
	return func(a *ActionRow) {
		if tokens != nil {
			a.designTokens = tokens
		}
	}
}

// WithActionRowWidth sets the rendered width.
func WithActionRowWidth(width int) ActionRowOption {
	return func(a *ActionRow) {
		if width >= 0 {
			a.width = width
		}
	}
}

// WithActionRowSeparator sets the separator used between action items.
func WithActionRowSeparator(separator string) ActionRowOption {
	return func(a *ActionRow) {
		a.separator = separator
	}
}

// WithActionRowAlign sets alignment: left, center, or right.
func WithActionRowAlign(align string) ActionRowOption {
	return func(a *ActionRow) {
		a.align = normalizeActionRowAlign(align)
	}
}

// NewActionRow creates a new ActionRow.
func NewActionRow(items []ActionItem, opts ...ActionRowOption) *ActionRow {
	a := &ActionRow{
		items:        append([]ActionItem(nil), items...),
		separator:    "  •  ",
		align:        "left",
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(a)
	}

	a.align = normalizeActionRowAlign(a.align)
	return a
}

// Init initializes the component.
func (a *ActionRow) Init() tea.Cmd { return nil }

// Update handles Bubble Tea messages.
func (a *ActionRow) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
	}
	return a, nil
}

// View renders the action row.
func (a *ActionRow) View() string {
	if len(a.items) == 0 {
		return ""
	}

	styledParts := make([]string, 0, len(a.items))
	plainParts := make([]string, 0, len(a.items))
	for _, item := range a.items {
		styled, plain := a.renderActionItem(item)
		styledParts = append(styledParts, styled)
		plainParts = append(plainParts, plain)
	}

	styledLine := strings.Join(styledParts, a.separator)
	plainLine := strings.Join(plainParts, a.separator)

	if a.width <= 0 {
		return styledLine
	}

	lineWidth := style.StringWidth(plainLine)
	if lineWidth >= a.width {
		return styledLine
	}

	padding := a.width - lineWidth
	leftPad := 0
	rightPad := 0

	switch normalizeActionRowAlign(a.align) {
	case "right":
		leftPad = padding
	case "center":
		leftPad = padding / 2
		rightPad = padding - leftPad
	default:
		rightPad = padding
	}

	return strings.Repeat(" ", leftPad) + styledLine + strings.Repeat(" ", rightPad)
}

// Focus marks the component as focused.
func (a *ActionRow) Focus() { a.focused = true }

// Blur marks the component as unfocused.
func (a *ActionRow) Blur() { a.focused = false }

// Focused reports whether the component is focused.
func (a *ActionRow) Focused() bool { return a.focused }

func (a *ActionRow) renderActionItem(item ActionItem) (styled string, plain string) {
	key := strings.TrimSpace(item.Key)
	label := strings.TrimSpace(item.Label)

	plain = strings.TrimSpace(strings.TrimSpace(key + " " + label))
	if key == "" {
		plain = label
	} else if label == "" {
		plain = key
	}

	keyColor := a.keyColor(item.Color)
	labelColor := a.labelColor()

	if key != "" && label != "" {
		styled = keyColor + style.ANSIBold + key + style.ANSIReset + " " + labelColor + label + style.ANSIReset
		return styled, plain
	}

	if key != "" {
		styled = keyColor + style.ANSIBold + key + style.ANSIReset
		return styled, plain
	}

	styled = labelColor + label + style.ANSIReset
	return styled, plain
}

func (a *ActionRow) keyColor(override string) string {
	if c := ansiColor(override); c != "" {
		return c
	}

	if a.designTokens != nil {
		if c := ansiColor(a.designTokens.Accent); c != "" {
			return c
		}
	}

	return style.ANSICyan
}

func (a *ActionRow) labelColor() string {
	if a.designTokens != nil {
		if c := ansiColor(a.designTokens.MutedColor); c != "" {
			return c
		}
		if c := ansiColor(a.designTokens.Color); c != "" {
			return c
		}
	}

	return style.ANSIDim
}

func ansiColor(token string) string {
	v := strings.TrimSpace(token)
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, "\033[") {
		return v
	}
	if c := style.ANSIColorFromHex(v); c != "" {
		return c
	}
	return ""
}

func normalizeActionRowAlign(align string) string {
	switch strings.ToLower(strings.TrimSpace(align)) {
	case "center", "right":
		return strings.ToLower(strings.TrimSpace(align))
	default:
		return "left"
	}
}

var _ tui.Component = (*ActionRow)(nil)
