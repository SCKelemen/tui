package display

import (
	"strconv"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// OrderedList renders a numbered list.
type OrderedList struct {
	items        []string
	width        int
	startNumber  int
	focused      bool
	designTokens *design.DesignTokens
}

// OrderedListOption configures an OrderedList.
type OrderedListOption func(*OrderedList)

// WithOrderedListWidth sets the list width.
func WithOrderedListWidth(width int) OrderedListOption {
	return func(o *OrderedList) {
		if width >= 0 {
			o.width = width
		}
	}
}

// WithOrderedListStartNumber sets the first item number.
func WithOrderedListStartNumber(start int) OrderedListOption {
	return func(o *OrderedList) {
		if start > 0 {
			o.startNumber = start
		}
	}
}

// WithOrderedListDesignTokens applies design tokens.
func WithOrderedListDesignTokens(tokens *design.DesignTokens) OrderedListOption {
	return func(o *OrderedList) {
		if tokens != nil {
			o.designTokens = tokens
		}
	}
}

// NewOrderedList creates an OrderedList component.
func NewOrderedList(items []string, opts ...OrderedListOption) *OrderedList {
	o := &OrderedList{
		items:        append([]string(nil), items...),
		startNumber:  1,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// Init initializes the component.
func (o *OrderedList) Init() tea.Cmd { return nil }

// Update handles Bubble Tea messages.
func (o *OrderedList) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		if o.width == 0 {
			o.width = m.Width
		}
	}
	return o, nil
}

// View renders the ordered list.
func (o *OrderedList) View() string {
	if len(o.items) == 0 {
		return ""
	}

	numberColor := style.ANSIBold
	textColor := ""
	if o.designTokens != nil {
		if c := style.Fg(o.designTokens.Accent); c != "" {
			numberColor = c + style.ANSIBold
		}
		if c := style.Fg(o.designTokens.Color); c != "" {
			textColor = c
		}
	}

	var b strings.Builder
	for i, item := range o.items {
		number := o.startNumber + i
		prefix := strconv.Itoa(number) + ". "
		line := strings.TrimSpace(item)
		if o.width > 0 {
			line = style.Truncate(line, max(1, o.width-style.StringWidth(prefix)), "…")
		}

		b.WriteString(numberColor)
		b.WriteString(prefix)
		b.WriteString(style.ANSIReset)
		if textColor != "" {
			b.WriteString(textColor)
		}
		b.WriteString(line)
		b.WriteString(style.ANSIReset)
		if i < len(o.items)-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}

// Focus marks the component as focused.
func (o *OrderedList) Focus() { o.focused = true }

// Blur marks the component as unfocused.
func (o *OrderedList) Blur() { o.focused = false }

// Focused reports whether the component is focused.
func (o *OrderedList) Focused() bool { return o.focused }

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var _ tui.Component = (*OrderedList)(nil)
