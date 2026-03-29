package display

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusBarSection is a single styled text segment within the status bar.
type StatusBarSection struct {
	Text  string
	Color string
	Bold  bool
}

// StatusBar renders a full-width bottom status bar with left, center, and right groups.
type StatusBar struct {
	left    []StatusBarSection
	center  []StatusBarSection
	right   []StatusBarSection
	bgColor string
	width   int
}

// StatusBarOption configures a StatusBar.
type StatusBarOption func(*StatusBar)

// NewStatusBar creates a StatusBar with default configuration.
func NewStatusBar(opts ...StatusBarOption) *StatusBar {
	s := &StatusBar{
		bgColor: "#1F1F1F",
		width:   80,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// WithStatusBarLeft sets the left-aligned section group.
func WithStatusBarLeft(sections ...StatusBarSection) StatusBarOption {
	return func(s *StatusBar) {
		s.left = append([]StatusBarSection(nil), sections...)
	}
}

// WithStatusBarCenter sets the centered section group.
func WithStatusBarCenter(sections ...StatusBarSection) StatusBarOption {
	return func(s *StatusBar) {
		s.center = append([]StatusBarSection(nil), sections...)
	}
}

// WithStatusBarRight sets the right-aligned section group.
func WithStatusBarRight(sections ...StatusBarSection) StatusBarOption {
	return func(s *StatusBar) {
		s.right = append([]StatusBarSection(nil), sections...)
	}
}

// WithStatusBarBg sets the background color used across the status bar.
func WithStatusBarBg(color string) StatusBarOption {
	return func(s *StatusBar) {
		s.bgColor = color
	}
}

// WithStatusBarWidth sets the status bar width.
func WithStatusBarWidth(w int) StatusBarOption {
	return func(s *StatusBar) {
		s.width = w
	}
}

// SetLeft updates the left-aligned section group.
func (s *StatusBar) SetLeft(sections ...StatusBarSection) {
	s.left = append([]StatusBarSection(nil), sections...)
}

// SetCenter updates the centered section group.
func (s *StatusBar) SetCenter(sections ...StatusBarSection) {
	s.center = append([]StatusBarSection(nil), sections...)
}

// SetRight updates the right-aligned section group.
func (s *StatusBar) SetRight(sections ...StatusBarSection) {
	s.right = append([]StatusBarSection(nil), sections...)
}

// SetWidth updates the status bar width.
func (s *StatusBar) SetWidth(w int) {
	s.width = w
}

// View renders a full-width single-line status bar.
func (s *StatusBar) View() string {
	if s.width <= 0 {
		return ""
	}

	left := s.renderSections(s.left)
	center := s.renderSections(s.center)
	right := s.renderSections(s.right)

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

func (s *StatusBar) renderSections(sections []StatusBarSection) string {
	if len(sections) == 0 {
		return ""
	}

	parts := make([]string, 0, len(sections))
	for _, section := range sections {
		sectionStyle := lipgloss.NewStyle().Background(lipgloss.Color(s.bgColor))
		if section.Color != "" {
			sectionStyle = sectionStyle.Foreground(lipgloss.Color(section.Color))
		}
		if section.Bold {
			sectionStyle = sectionStyle.Bold(true)
		}
		parts = append(parts, sectionStyle.Render(section.Text))
	}

	return strings.Join(parts, " ")
}
