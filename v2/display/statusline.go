package display

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusLineItem is a single labeled value segment in the status line.
type StatusLineItem struct {
	Label string
	Value string
	Color string
	Bold  bool
}

// StatusLine renders a full-width bottom status line with left, center, and right sections.
type StatusLine struct {
	left                  []StatusLineItem
	center                []StatusLineItem
	right                 []StatusLineItem
	width                 int
	bgColor               string
	focused               bool
	designTokens          *design.DesignTokens
	backgroundColorSetOpt bool
}

// StatusLineOption configures a StatusLine.
type StatusLineOption func(*StatusLine)

// NewStatusLine creates a StatusLine with defaults.
func NewStatusLine(opts ...StatusLineOption) *StatusLine {
	s := &StatusLine{
		width:        80,
		bgColor:      "#1F1F1F",
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(s)
	}

	s.resolveBackgroundColor()
	return s
}

// WithStatusLineLeft sets the left section items.
func WithStatusLineLeft(items []StatusLineItem) StatusLineOption {
	return func(s *StatusLine) {
		s.left = append([]StatusLineItem(nil), items...)
	}
}

// WithStatusLineCenter sets the center section items.
func WithStatusLineCenter(items []StatusLineItem) StatusLineOption {
	return func(s *StatusLine) {
		s.center = append([]StatusLineItem(nil), items...)
	}
}

// WithStatusLineRight sets the right section items.
func WithStatusLineRight(items []StatusLineItem) StatusLineOption {
	return func(s *StatusLine) {
		s.right = append([]StatusLineItem(nil), items...)
	}
}

// WithStatusLineWidth sets the rendered width.
func WithStatusLineWidth(width int) StatusLineOption {
	return func(s *StatusLine) {
		if width >= 0 {
			s.width = width
		}
	}
}

// WithStatusLineBackgroundColor sets the status line background color.
func WithStatusLineBackgroundColor(color string) StatusLineOption {
	return func(s *StatusLine) {
		s.bgColor = strings.TrimSpace(color)
		s.backgroundColorSetOpt = true
	}
}

// WithStatusLineDesignTokens applies design-system tokens.
func WithStatusLineDesignTokens(tokens *design.DesignTokens) StatusLineOption {
	return func(s *StatusLine) {
		if tokens != nil {
			s.designTokens = tokens
		}
	}
}

// Init initializes the component.
func (s *StatusLine) Init() tea.Cmd { return nil }

// Update handles Bubble Tea messages.
func (s *StatusLine) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
	}
	return s, nil
}

// View renders a full-width single-line status line.
func (s *StatusLine) View() string {
	if s.width <= 0 {
		return ""
	}

	s.resolveBackgroundColor()

	left := s.renderItems(s.left)
	center := s.renderItems(s.center)
	right := s.renderItems(s.right)

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)

	if leftWidth > s.width {
		left = lipgloss.NewStyle().MaxWidth(s.width).Render(left)
		leftWidth = lipgloss.Width(left)
		right = ""
		rightWidth = 0
		center = ""
	}

	if rightWidth > s.width {
		right = lipgloss.NewStyle().MaxWidth(s.width).Render(right)
		rightWidth = lipgloss.Width(right)
		left = ""
		leftWidth = 0
		center = ""
	}

	if leftWidth+rightWidth > s.width {
		leftAvail := s.width / 2
		rightAvail := s.width - leftAvail

		left = lipgloss.NewStyle().MaxWidth(leftAvail).Render(left)
		right = lipgloss.NewStyle().MaxWidth(rightAvail).Render(right)
		leftWidth = lipgloss.Width(left)
		rightWidth = lipgloss.Width(right)
		center = ""
	}

	middleWidth := s.width - leftWidth - rightWidth
	if middleWidth < 0 {
		middleWidth = 0
	}

	center = lipgloss.NewStyle().MaxWidth(middleWidth).Render(center)
	centerWidth := lipgloss.Width(center)

	leftPad := 0
	rightPad := 0
	if middleWidth > centerWidth {
		leftPad = (middleWidth - centerWidth) / 2
		rightPad = middleWidth - centerWidth - leftPad
	}

	line := left + strings.Repeat(" ", leftPad) + center + strings.Repeat(" ", rightPad) + right

	barStyle := lipgloss.NewStyle().Background(lipgloss.Color(s.bgColor)).Width(s.width).MaxWidth(s.width)
	return barStyle.Render(line)
}

// Focus marks the component as focused.
func (s *StatusLine) Focus() { s.focused = true }

// Blur marks the component as unfocused.
func (s *StatusLine) Blur() { s.focused = false }

// Focused reports whether the component is focused.
func (s *StatusLine) Focused() bool { return s.focused }

func (s *StatusLine) resolveBackgroundColor() {
	if s.backgroundColorSetOpt && strings.TrimSpace(s.bgColor) != "" {
		return
	}

	if s.designTokens != nil {
		if bg := strings.TrimSpace(s.designTokens.Background); bg != "" {
			s.bgColor = bg
			return
		}
	}

	if strings.TrimSpace(s.bgColor) == "" {
		s.bgColor = "#1F1F1F"
	}
}

func (s *StatusLine) renderItems(items []StatusLineItem) string {
	if len(items) == 0 {
		return ""
	}

	parts := make([]string, 0, len(items))
	for _, item := range items {
		text := s.itemText(item)
		if text == "" {
			continue
		}

		itemStyle := lipgloss.NewStyle().Background(lipgloss.Color(s.bgColor))
		if color := strings.TrimSpace(item.Color); color != "" {
			itemStyle = itemStyle.Foreground(lipgloss.Color(color))
		}
		if item.Bold {
			itemStyle = itemStyle.Bold(true)
		}

		parts = append(parts, itemStyle.Render(text))
	}

	return strings.Join(parts, " │ ")
}

func (s *StatusLine) itemText(item StatusLineItem) string {
	label := strings.TrimSpace(item.Label)
	value := strings.TrimSpace(item.Value)

	switch {
	case label != "" && value != "":
		return label + ": " + value
	case label != "":
		return label
	default:
		return value
	}
}

var _ tui.Component = (*StatusLine)(nil)
