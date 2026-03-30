package input

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// RadioItem is one selectable radio option.
type RadioItem struct {
	Label       string
	Value       string
	Description string
}

// RadioButtonSelectMsg is emitted when the selected radio value changes.
type RadioButtonSelectMsg struct {
	Value string
}

// RadioButtonSelectOption configures a RadioButtonSelect.
type RadioButtonSelectOption func(*RadioButtonSelect)

// WithRadioButtonSelectDesignTokens applies design-system colors.
func WithRadioButtonSelectDesignTokens(tokens *design.DesignTokens) RadioButtonSelectOption {
	return func(r *RadioButtonSelect) {
		r.applyDesignTokens(tokens)
	}
}

// WithRadioButtonSelectSelected sets the initially selected index.
func WithRadioButtonSelectSelected(index int) RadioButtonSelectOption {
	return func(r *RadioButtonSelect) {
		r.selected = index
		r.cursor = index
	}
}

// WithRadioButtonSelectWidth sets a preferred render width.
func WithRadioButtonSelectWidth(width int) RadioButtonSelectOption {
	return func(r *RadioButtonSelect) {
		if width > 0 {
			r.width = width
		}
	}
}

// RadioButtonSelect renders a vertical list of radio button options.
type RadioButtonSelect struct {
	items        []RadioItem
	selected     int
	cursor       int
	focused      bool
	width        int
	designTokens *design.DesignTokens

	accentColor   string
	mutedColor    string
	descriptionFx string
	cursorBgColor string
}

// NewRadioButtonSelect creates a new radio button select component.
func NewRadioButtonSelect(items []RadioItem, opts ...RadioButtonSelectOption) *RadioButtonSelect {
	r := &RadioButtonSelect{
		items:         append([]RadioItem(nil), items...),
		selected:      0,
		cursor:        0,
		focused:       false,
		width:         0,
		designTokens:  design.DefaultTheme(),
		accentColor:   style.ANSICyan,
		mutedColor:    style.ANSIDim,
		descriptionFx: style.ANSIDim,
		cursorBgColor: style.ANSIInverse,
	}

	for _, opt := range opts {
		opt(r)
	}

	r.clampIndices()
	r.applyDesignTokens(r.designTokens)

	return r
}

// Init initializes the component.
func (r *RadioButtonSelect) Init() tea.Cmd {
	return nil
}

// Update handles keyboard input and window size updates.
func (r *RadioButtonSelect) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if r.width <= 0 {
			r.width = msg.Width
		}
		return r, nil
	case tea.KeyMsg:
		if !r.focused || len(r.items) == 0 {
			return r, nil
		}

		switch msg.String() {
		case "down", "j":
			if r.cursor < len(r.items)-1 {
				r.cursor++
			}
			return r, nil
		case "up", "k":
			if r.cursor > 0 {
				r.cursor--
			}
			return r, nil
		case "enter", " ":
			if r.selected == r.cursor {
				return r, nil
			}
			r.selected = r.cursor
			value := r.items[r.selected].Value
			return r, func() tea.Msg {
				return RadioButtonSelectMsg{Value: value}
			}
		}
	}

	return r, nil
}

// View renders radio rows and optional descriptions.
func (r *RadioButtonSelect) View() string {
	if len(r.items) == 0 {
		return ""
	}

	lines := make([]string, 0, len(r.items)*2)
	for i, item := range r.items {
		selected := i == r.selected
		cursor := i == r.cursor

		marker := "○"
		lineColor := r.mutedColor
		if selected {
			marker = "●"
			lineColor = r.accentColor
		}

		label := strings.TrimSpace(item.Label)
		if label == "" {
			label = item.Value
		}

		line := lineColor + marker + " " + label + style.ANSIReset
		line = r.fitWidth(line)
		if cursor {
			line = r.cursorBgColor + line + style.ANSIReset
		}
		lines = append(lines, line)

		if strings.TrimSpace(item.Description) != "" {
			desc := r.descriptionFx + "  " + item.Description + style.ANSIReset
			desc = r.fitWidth(desc)
			if cursor {
				desc = r.cursorBgColor + desc + style.ANSIReset
			}
			lines = append(lines, desc)
		}
	}

	return strings.Join(lines, "\n")
}

// Focus marks the component as focused.
func (r *RadioButtonSelect) Focus() {
	r.focused = true
}

// Blur marks the component as unfocused.
func (r *RadioButtonSelect) Blur() {
	r.focused = false
}

// Focused reports whether the component currently has focus.
func (r *RadioButtonSelect) Focused() bool {
	return r.focused
}

func (r *RadioButtonSelect) clampIndices() {
	if len(r.items) == 0 {
		r.selected = 0
		r.cursor = 0
		return
	}

	if r.selected < 0 {
		r.selected = 0
	}
	if r.selected >= len(r.items) {
		r.selected = len(r.items) - 1
	}
	if r.cursor < 0 {
		r.cursor = 0
	}
	if r.cursor >= len(r.items) {
		r.cursor = len(r.items) - 1
	}
}

func (r *RadioButtonSelect) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}
	r.designTokens = tokens

	if accent := style.ANSIColorFromHex(tokens.Accent); accent != "" {
		r.accentColor = accent
	}
	if muted := style.ANSIColorFromHex(tokens.Color); muted != "" {
		r.mutedColor = muted
	}
	if bg := style.ANSIBackgroundColorFromHex(tokens.Accent); bg != "" {
		r.cursorBgColor = bg
	}
}

func (r *RadioButtonSelect) fitWidth(line string) string {
	if r.width <= 0 {
		return line
	}
	if style.StringWidth(line) <= r.width {
		return style.Pad(line, r.width)
	}
	return style.Truncate(line, r.width, "…")
}

var _ tui.Component = (*RadioButtonSelect)(nil)
