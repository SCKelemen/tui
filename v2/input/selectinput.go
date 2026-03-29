package input

import (
	"strings"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SelectItem is one selectable option in a SelectInput.
type SelectItem struct {
	Label       string
	Value       string
	Description string
}

// SelectInputMsg is emitted when selection is confirmed with Enter.
type SelectInputMsg struct {
	Item SelectItem
}

// SelectInputOption configures a SelectInput.
type SelectInputOption func(*SelectInput)

// WithSelectInputWidth sets a preferred render width in terminal cells.
func WithSelectInputWidth(width int) SelectInputOption {
	return func(s *SelectInput) {
		if width > 0 {
			s.width = width
		}
	}
}

// WithSelectInputHeight sets the maximum number of visible items.
func WithSelectInputHeight(height int) SelectInputOption {
	return func(s *SelectInput) {
		if height > 0 {
			s.height = height
		}
	}
}

// WithSelectInputIndicator sets the cursor indicator for the selected row.
func WithSelectInputIndicator(indicator string) SelectInputOption {
	return func(s *SelectInput) {
		if strings.TrimSpace(indicator) != "" {
			s.indicator = indicator
		}
	}
}

// SelectInput is a keyboard-driven select list component.
type SelectInput struct {
	items        []SelectItem
	cursor       int
	focused      bool
	width        int
	height       int
	scrollOffset int

	indicator      string
	indicatorStyle lipgloss.Style
	itemStyle      lipgloss.Style
	selectedStyle  lipgloss.Style
}

// NewSelectInput creates a new SelectInput.
func NewSelectInput(items []SelectItem, opts ...SelectInputOption) *SelectInput {
	si := &SelectInput{
		items:      append([]SelectItem(nil), items...),
		cursor:     0,
		focused:    false,
		width:      0,
		height:     8,
		indicator:  "❯",
		itemStyle:  lipgloss.NewStyle(),
		selectedStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#61AFEF")),
		indicatorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#61AFEF")),
	}

	for _, opt := range opts {
		opt(si)
	}

	si.clampCursor()
	si.ensureCursorVisible()

	return si
}

// Init initializes the component.
func (s *SelectInput) Init() tea.Cmd {
	return nil
}

// Update handles keyboard and window messages.
func (s *SelectInput) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if s.width <= 0 {
			s.width = msg.Width
		}
		s.ensureCursorVisible()
		return s, nil
	case tea.KeyMsg:
		if !s.focused || len(s.items) == 0 {
			return s, nil
		}

		switch msg.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
				s.ensureCursorVisible()
			}
			return s, nil
		case "down", "j":
			if s.cursor < len(s.items)-1 {
				s.cursor++
				s.ensureCursorVisible()
			}
			return s, nil
		case "enter":
			selected := s.Selected()
			return s, func() tea.Msg {
				return SelectInputMsg{Item: selected}
			}
		}
	}

	return s, nil
}

// View renders the visible item list with the active cursor indicator.
func (s *SelectInput) View() string {
	if len(s.items) == 0 {
		return ""
	}

	start, end := s.visibleRange()
	lines := make([]string, 0, end-start)

	for i := start; i < end; i++ {
		item := s.items[i]
		isSelected := i == s.cursor

		indicator := "  "
		if isSelected {
			indicator = s.indicatorStyle.Render(s.indicator + " ")
		}

		line := indicator + s.renderItemLabel(item, isSelected)
		if s.width > 0 {
			line = style.Pad(style.Truncate(line, s.width, ""), s.width)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// Focus marks the component as focused.
func (s *SelectInput) Focus() {
	s.focused = true
}

// Blur marks the component as unfocused.
func (s *SelectInput) Blur() {
	s.focused = false
}

// Focused reports whether the component is focused.
func (s *SelectInput) Focused() bool {
	return s.focused
}

// Selected returns the currently selected item.
func (s *SelectInput) Selected() SelectItem {
	if len(s.items) == 0 {
		return SelectItem{}
	}
	s.clampCursor()
	return s.items[s.cursor]
}

// SelectedIndex returns the current selected index.
func (s *SelectInput) SelectedIndex() int {
	if len(s.items) == 0 {
		return -1
	}
	s.clampCursor()
	return s.cursor
}

// SetItems replaces the list of selectable items.
func (s *SelectInput) SetItems(items []SelectItem) {
	s.items = append([]SelectItem(nil), items...)
	s.clampCursor()
	s.ensureCursorVisible()
}

func (s *SelectInput) renderItemLabel(item SelectItem, isSelected bool) string {
	label := item.Label
	if strings.TrimSpace(item.Description) != "" {
		label += " " + lipgloss.NewStyle().Faint(true).Render(item.Description)
	}

	if isSelected {
		return s.selectedStyle.Render(label)
	}
	return s.itemStyle.Render(label)
}

func (s *SelectInput) clampCursor() {
	if len(s.items) == 0 {
		s.cursor = 0
		s.scrollOffset = 0
		return
	}
	if s.cursor < 0 {
		s.cursor = 0
	}
	if s.cursor >= len(s.items) {
		s.cursor = len(s.items) - 1
	}
}

func (s *SelectInput) visibleRange() (start int, end int) {
	if len(s.items) == 0 {
		return 0, 0
	}
	maxVisible := s.height
	if maxVisible <= 0 || maxVisible > len(s.items) {
		maxVisible = len(s.items)
	}

	start = s.scrollOffset
	if start < 0 {
		start = 0
	}
	if start > len(s.items)-1 {
		start = len(s.items) - 1
	}

	end = start + maxVisible
	if end > len(s.items) {
		end = len(s.items)
	}

	if s.cursor < start {
		start = s.cursor
		end = start + maxVisible
		if end > len(s.items) {
			end = len(s.items)
		}
	}
	if s.cursor >= end {
		start = s.cursor - maxVisible + 1
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > len(s.items) {
			end = len(s.items)
		}
	}

	return start, end
}

func (s *SelectInput) ensureCursorVisible() {
	if len(s.items) == 0 {
		s.scrollOffset = 0
		return
	}

	s.clampCursor()
	maxVisible := s.height
	if maxVisible <= 0 || maxVisible > len(s.items) {
		maxVisible = len(s.items)
	}

	if s.cursor < s.scrollOffset {
		s.scrollOffset = s.cursor
	}
	if s.cursor >= s.scrollOffset+maxVisible {
		s.scrollOffset = s.cursor - maxVisible + 1
	}
	if s.scrollOffset < 0 {
		s.scrollOffset = 0
	}
	maxOffset := len(s.items) - maxVisible
	if maxOffset < 0 {
		maxOffset = 0
	}
	if s.scrollOffset > maxOffset {
		s.scrollOffset = maxOffset
	}
}

var _ tui.Component = (*SelectInput)(nil)
