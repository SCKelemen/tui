package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/cli/renderer"
	"github.com/SCKelemen/color"
	design "github.com/SCKelemen/design-system"
	"github.com/SCKelemen/layout"
)

// StatCard displays a metric with title, value, change indicator, and optional sparkline
type StatCard struct {
	width   int
	height  int
	focused bool
	tokens  *design.DesignTokens

	// Content
	title      string
	value      string
	subtitle   string
	change     int       // Absolute change
	changePct  float64   // Percentage change
	trend      []float64 // Sparkline data
	color      string    // Accent color for highlights
	trendColor string    // Color for trend/sparkline
}

// StatCardOption configures a StatCard
type StatCardOption func(*StatCard)

// WithTitle sets the card title
func WithTitle(title string) StatCardOption {
	return func(s *StatCard) {
		s.title = title
	}
}

// WithValue sets the main value to display
func WithValue(value string) StatCardOption {
	return func(s *StatCard) {
		s.value = value
	}
}

// WithSubtitle sets the subtitle/description
func WithSubtitle(subtitle string) StatCardOption {
	return func(s *StatCard) {
		s.subtitle = subtitle
	}
}

// WithChange sets the change value and percentage
func WithChange(change int, changePct float64) StatCardOption {
	return func(s *StatCard) {
		s.change = change
		s.changePct = changePct
	}
}

// WithTrend sets the sparkline trend data
func WithTrend(trend []float64) StatCardOption {
	return func(s *StatCard) {
		s.trend = trend
	}
}

// WithColor sets the accent color
func WithColor(color string) StatCardOption {
	return func(s *StatCard) {
		s.color = color
	}
}

// WithTrendColor sets the trend line color
func WithTrendColor(color string) StatCardOption {
	return func(s *StatCard) {
		s.trendColor = color
	}
}

// NewStatCard creates a new stat card
func NewStatCard(opts ...StatCardOption) *StatCard {
	s := &StatCard{
		width:      30,
		height:     8,
		tokens:     design.DefaultTheme(),
		color:      "#2196F3",
		trendColor: "#4CAF50",
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Init initializes the stat card
func (s *StatCard) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *StatCard) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	}

	return s, nil
}

// View renders the stat card
func (s *StatCard) View() string {
	if s.width == 0 {
		return ""
	}

	// Use simple string-based rendering for now
	return s.renderSimple()
}

// Focus is called when this component receives focus
func (s *StatCard) Focus() {
	s.focused = true
}

// Blur is called when this component loses focus
func (s *StatCard) Blur() {
	s.focused = false
}

// Focused returns whether this component is currently focused
func (s *StatCard) Focused() bool {
	return s.focused
}

// renderSimple provides string-based rendering
func (s *StatCard) renderSimple() string {
	var b strings.Builder

	// Calculate dimensions
	contentWidth := s.width - 4 // Account for borders and padding
	if contentWidth < 10 {
		contentWidth = 10
	}

	// Top border
	b.WriteString("┌")
	b.WriteString(strings.Repeat("─", s.width-2))
	b.WriteString("┐\n")

	// Title row
	b.WriteString("│ ")
	b.WriteString(s.truncate(s.title, contentWidth))
	b.WriteString(" │\n")

	// Value row
	b.WriteString("│ ")
	valueStr := "\033[1m" + s.value + "\033[0m" // Bold
	b.WriteString(valueStr)
	// Use visible length to account for ANSI codes
	visibleValueLen := s.visibleLength(valueStr)
	if visibleValueLen < contentWidth {
		b.WriteString(strings.Repeat(" ", contentWidth-visibleValueLen))
	}
	b.WriteString(" │\n")

	// Change indicator row
	if s.change != 0 || s.changePct != 0 {
		b.WriteString("│ ")
		changeStr := s.renderChange()
		b.WriteString(changeStr)
		// Calculate visible length (without ANSI codes)
		visibleLen := s.visibleLength(changeStr)
		if visibleLen < contentWidth {
			b.WriteString(strings.Repeat(" ", contentWidth-visibleLen))
		}
		b.WriteString(" │\n")
	}

	// Subtitle row
	if s.subtitle != "" {
		b.WriteString("│ ")
		b.WriteString(s.truncate(s.subtitle, contentWidth))
		b.WriteString(" │\n")
	}

	// Sparkline row
	if len(s.trend) > 0 {
		b.WriteString("│ ")
		sparkline := s.renderSparkline(contentWidth)
		b.WriteString(sparkline)
		b.WriteString(" │\n")
	}

	// Fill remaining height
	currentHeight := 3 // Top border + title + value
	if s.change != 0 || s.changePct != 0 {
		currentHeight++
	}
	if s.subtitle != "" {
		currentHeight++
	}
	if len(s.trend) > 0 {
		currentHeight++
	}

	for currentHeight < s.height-1 {
		b.WriteString("│")
		b.WriteString(strings.Repeat(" ", s.width-2))
		b.WriteString("│\n")
		currentHeight++
	}

	// Bottom border
	b.WriteString("└")
	b.WriteString(strings.Repeat("─", s.width-2))
	b.WriteString("┘\n")

	return b.String()
}

// renderChange renders the change indicator with color
func (s *StatCard) renderChange() string {
	var changeColor string
	var arrow string

	if s.change > 0 {
		changeColor = "\033[32m" // Green
		arrow = "↑"
	} else if s.change < 0 {
		changeColor = "\033[31m" // Red
		arrow = "↓"
	} else {
		changeColor = "\033[37m" // White
		arrow = "→"
	}

	changeStr := fmt.Sprintf("%s%s %d (%.1f%%)%s",
		changeColor, arrow, abs(s.change), s.changePct, "\033[0m")

	return changeStr
}

// renderSparkline renders a simple ASCII sparkline
func (s *StatCard) renderSparkline(width int) string {
	if len(s.trend) == 0 {
		return ""
	}

	// Use block characters for sparkline
	blocks := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	// Find min and max
	min, max := s.trend[0], s.trend[0]
	for _, v := range s.trend {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Normalize to 0-7 range for block selection
	normalize := func(v float64) int {
		if max == min {
			return 4 // Middle block
		}
		normalized := (v - min) / (max - min)
		index := int(normalized * 7)
		if index < 0 {
			index = 0
		}
		if index > 7 {
			index = 7
		}
		return index
	}

	var b strings.Builder

	// Determine how many data points to show
	pointsToShow := len(s.trend)
	if pointsToShow > width {
		pointsToShow = width
	}

	// Calculate step size if we need to sample
	step := 1
	if len(s.trend) > width {
		step = len(s.trend) / width
	}

	// Render sparkline with trend color
	b.WriteString("\033[38;2;76;175;80m") // Green color for trend
	for i := 0; i < pointsToShow; i++ {
		dataIndex := i * step
		if dataIndex >= len(s.trend) {
			dataIndex = len(s.trend) - 1
		}
		blockIndex := normalize(s.trend[dataIndex])
		b.WriteString(blocks[blockIndex])
	}
	b.WriteString("\033[0m")

	// Pad to width
	sparklineLen := pointsToShow
	if sparklineLen < width {
		b.WriteString(strings.Repeat(" ", width-sparklineLen))
	}

	return b.String()
}

// renderWithLayout renders using the full layout system (future)
func (s *StatCard) renderWithLayout() string {
	// Create layout context
	ctx := layout.NewLayoutContext(float64(s.width), float64(s.height), 16)

	// Use CardLayout helper
	card := LayoutHelpers.CardLayout(1)
	card.Style.Width = layout.Px(float64(s.width))
	card.Style.Height = layout.Px(float64(s.height))

	// Perform layout
	constraints := layout.Tight(float64(s.width), float64(s.height))
	layout.Layout(card, constraints, ctx)

	// Convert to styled nodes
	textColorRGBA, _ := color.HexToRGB(s.tokens.Color)
	var textColor color.Color = textColorRGBA

	cardStyled := renderer.NewStyledNode(card, &renderer.Style{
		Foreground: &textColor,
	})

	// Build content
	var content strings.Builder
	content.WriteString(s.title + "\n")
	content.WriteString(s.value + "\n")
	if s.change != 0 || s.changePct != 0 {
		content.WriteString(s.renderChange() + "\n")
	}
	if s.subtitle != "" {
		content.WriteString(s.subtitle + "\n")
	}

	cardStyled.Content = content.String()

	// Render to screen
	screen := renderer.NewScreen(s.width, s.height)
	screen.Render(cardStyled)

	return screen.String()
}

// truncate truncates a string to fit within width (using rune count for better unicode support)
func (s *StatCard) truncate(str string, width int) string {
	runes := []rune(str)
	runeLen := len(runes)

	if runeLen <= width {
		return str + strings.Repeat(" ", width-runeLen)
	}
	if width > 3 {
		return string(runes[:width-3]) + "..."
	}
	if width > 0 {
		return string(runes[:width])
	}
	return ""
}

// visibleLength calculates the visible length of a string (excluding ANSI codes, counting runes)
func (s *StatCard) visibleLength(str string) int {
	// Count runes while skipping ANSI escape sequences
	inEscape := false
	count := 0
	for _, ch := range str {
		if ch == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if ch == 'm' {
				inEscape = false
			}
			continue
		}
		count++
	}
	return count
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
