package input

import (
	"fmt"
	"math"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ColorOption is one selectable color.
type ColorOption struct {
	Name string
	Hex  string
}

// ColorSelectedMsg is emitted when a color is confirmed with Enter.
type ColorSelectedMsg struct {
	Color ColorOption
}

// ColorPickerOption configures a ColorPicker.
type ColorPickerOption func(*ColorPicker)

// WithColorPickerWidth sets preferred render width.
func WithColorPickerWidth(width int) ColorPickerOption {
	return func(c *ColorPicker) {
		if width > 0 {
			c.width = width
		}
	}
}

// WithColorPickerSelected sets the initial selected color by name or hex value.
func WithColorPickerSelected(selected string) ColorPickerOption {
	return func(c *ColorPicker) {
		c.selected = strings.TrimSpace(selected)
	}
}

// WithColorPickerColumns sets number of swatch columns.
func WithColorPickerColumns(columns int) ColorPickerOption {
	return func(c *ColorPicker) {
		if columns > 0 {
			c.columns = columns
		}
	}
}

// WithColorPickerDesignTokens applies design-system tokens.
func WithColorPickerDesignTokens(tokens *design.DesignTokens) ColorPickerOption {
	return func(c *ColorPicker) {
		if tokens != nil {
			c.designTokens = tokens
		}
	}
}

// ColorPicker renders a color palette grid with keyboard navigation.
type ColorPicker struct {
	colors       []ColorOption
	cursor       int
	selected     string
	columns      int
	width        int
	windowWidth  int
	focused      bool
	designTokens *design.DesignTokens
}

// NewColorPicker creates a new ColorPicker.
func NewColorPicker(colors []ColorOption, opts ...ColorPickerOption) *ColorPicker {
	c := &ColorPicker{
		colors:       append([]ColorOption(nil), colors...),
		cursor:       0,
		selected:     "",
		columns:      4,
		width:        0,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(c)
	}

	c.syncCursorToSelected()
	c.clampCursor()
	return c
}

// Init initializes the component.
func (c *ColorPicker) Init() tea.Cmd { return nil }

// Update handles keyboard and resize messages.
func (c *ColorPicker) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.windowWidth = msg.Width
		return c, nil
	case tea.KeyMsg:
		if !c.focused || len(c.colors) == 0 {
			return c, nil
		}

		switch msg.String() {
		case "left", "h":
			if c.cursor > 0 {
				c.cursor--
			}
			return c, nil
		case "right", "l":
			if c.cursor < len(c.colors)-1 {
				c.cursor++
			}
			return c, nil
		case "up", "k":
			next := c.cursor - c.effectiveColumns()
			if next >= 0 {
				c.cursor = next
			}
			return c, nil
		case "down", "j":
			next := c.cursor + c.effectiveColumns()
			if next < len(c.colors) {
				c.cursor = next
			}
			return c, nil
		case "enter":
			selected := c.Selected()
			c.selected = selected.Name
			return c, func() tea.Msg {
				return ColorSelectedMsg{Color: selected}
			}
		}
	}

	return c, nil
}

// View renders the color swatch grid and selected color details.
func (c *ColorPicker) View() string {
	if len(c.colors) == 0 {
		return ""
	}

	width := c.effectiveWidth()
	cols := c.effectiveColumns()
	if cols < 1 {
		cols = 1
	}

	cellWidth := 12
	if width > 0 {
		candidate := width / cols
		if candidate > 6 {
			cellWidth = candidate
		}
	}

	rows := int(math.Ceil(float64(len(c.colors)) / float64(cols)))
	lines := make([]string, 0, rows+2)

	for r := 0; r < rows; r++ {
		cells := make([]string, 0, cols)
		for col := 0; col < cols; col++ {
			idx := r*cols + col
			if idx >= len(c.colors) {
				cells = append(cells, strings.Repeat(" ", cellWidth))
				continue
			}
			cells = append(cells, c.renderCell(idx, c.colors[idx], cellWidth))
		}
		lines = append(lines, strings.Join(cells, " "))
	}

	selected := c.Selected()
	detail := fmt.Sprintf("%s%s%s %s", style.ANSIBold, selected.Hex, style.ANSIReset, selected.Name)
	if width > 0 {
		detail = style.Pad(style.Truncate(detail, width, "…"), width)
	}
	lines = append(lines, detail)

	return strings.Join(lines, "\n")
}

// Focus marks the component focused.
func (c *ColorPicker) Focus() { c.focused = true }

// Blur marks the component unfocused.
func (c *ColorPicker) Blur() { c.focused = false }

// Focused reports whether this component is focused.
func (c *ColorPicker) Focused() bool { return c.focused }

// Selected returns the currently highlighted color.
func (c *ColorPicker) Selected() ColorOption {
	if len(c.colors) == 0 {
		return ColorOption{}
	}
	c.clampCursor()
	return c.colors[c.cursor]
}

func (c *ColorPicker) effectiveColumns() int {
	if c.columns > 0 {
		return c.columns
	}
	return 4
}

func (c *ColorPicker) effectiveWidth() int {
	if c.width > 0 {
		return c.width
	}
	return c.windowWidth
}

func (c *ColorPicker) clampCursor() {
	if len(c.colors) == 0 {
		c.cursor = 0
		return
	}
	if c.cursor < 0 {
		c.cursor = 0
	}
	if c.cursor >= len(c.colors) {
		c.cursor = len(c.colors) - 1
	}
}

func (c *ColorPicker) syncCursorToSelected() {
	if strings.TrimSpace(c.selected) == "" {
		return
	}
	for i, color := range c.colors {
		if strings.EqualFold(strings.TrimSpace(c.selected), strings.TrimSpace(color.Name)) || strings.EqualFold(strings.TrimSpace(c.selected), strings.TrimSpace(color.Hex)) {
			c.cursor = i
			return
		}
	}
}

func (c *ColorPicker) renderCell(index int, color ColorOption, cellWidth int) string {
	name := strings.TrimSpace(color.Name)
	if name == "" {
		name = strings.TrimPrefix(strings.TrimSpace(color.Hex), "#")
	}
	name = style.Truncate(name, 6, "…")

	swatch := "  "
	if bg := style.Bg(color.Hex); bg != "" {
		swatch = bg + "  " + style.ANSIReset
	}

	cursor := " "
	if index == c.cursor {
		cursor = "▸"
	}

	selectedMark := " "
	if strings.EqualFold(strings.TrimSpace(c.selected), strings.TrimSpace(color.Name)) || strings.EqualFold(strings.TrimSpace(c.selected), strings.TrimSpace(color.Hex)) {
		selectedMark = "✓"
	}

	line := fmt.Sprintf("%s%s %s%s", cursor, swatch, name, selectedMark)
	line = style.Truncate(line, cellWidth, "…")
	line = style.Pad(line, cellWidth)

	if index == c.cursor {
		line = style.ANSIInverse + line + style.ANSIReset
	}

	return line
}

var _ tui.Component = (*ColorPicker)(nil)
