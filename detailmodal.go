package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	design "github.com/SCKelemen/design-system"
)

// DetailModal displays detailed information about a StatCard
type DetailModal struct {
	width   int
	height  int
	visible bool
	focused bool
	tokens  *design.DesignTokens

	// Content from StatCard
	title      string
	value      string
	subtitle   string
	change     int
	changePct  float64
	trend      []float64
	color      string
	trendColor string

	// Additional details
	history []string // Historical data points
}

// DetailModalOption configures a DetailModal
type DetailModalOption func(*DetailModal)

// WithModalContent sets the content from a StatCard
func WithModalContent(card *StatCard) DetailModalOption {
	return func(m *DetailModal) {
		m.title = card.title
		m.value = card.value
		m.subtitle = card.subtitle
		m.change = card.change
		m.changePct = card.changePct
		m.trend = card.trend
		m.color = card.color
		m.trendColor = card.trendColor
	}
}

// WithHistory sets historical data points
func WithHistory(history []string) DetailModalOption {
	return func(m *DetailModal) {
		m.history = history
	}
}

// NewDetailModal creates a new detail modal
func NewDetailModal(opts ...DetailModalOption) *DetailModal {
	m := &DetailModal{
		tokens:  design.DefaultTheme(),
		visible: false,
		history: []string{},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// Init initializes the modal
func (m *DetailModal) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *DetailModal) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if !m.visible || !m.focused {
			return m, nil
		}

		switch msg.String() {
		case "esc", "q":
			m.Hide()
		}
	}

	return m, nil
}

// View renders the modal
func (m *DetailModal) View() string {
	if !m.visible || m.width == 0 {
		return ""
	}

	// Modal dimensions - 70% of viewport width, 80% of height
	modalWidth := int(float64(m.width) * 0.7)
	if modalWidth < 60 {
		modalWidth = 60
	}
	if modalWidth > m.width-4 {
		modalWidth = m.width - 4
	}

	modalHeight := int(float64(m.height) * 0.8)
	if modalHeight < 20 {
		modalHeight = 20
	}
	if modalHeight > m.height-4 {
		modalHeight = m.height - 4
	}

	// Center the modal
	offsetX := (m.width - modalWidth) / 2
	offsetY := (m.height - modalHeight) / 2

	// Render modal content
	modalContent := m.renderModalContent(modalWidth, modalHeight)

	// Position modal with centering
	var b strings.Builder
	lines := strings.Split(modalContent, "\n")

	// Add empty lines before modal (vertical centering)
	for i := 0; i < offsetY; i++ {
		b.WriteString("\n")
	}

	// Add each modal line with horizontal centering
	for _, line := range lines {
		if line == "" {
			b.WriteString("\n")
			continue
		}
		// Add horizontal offset
		b.WriteString(strings.Repeat(" ", offsetX))
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

// Focus is called when this component receives focus
func (m *DetailModal) Focus() {
	m.focused = true
}

// Blur is called when this component loses focus
func (m *DetailModal) Blur() {
	m.focused = false
}

// Focused returns whether this component is currently focused
func (m *DetailModal) Focused() bool {
	return m.focused
}

// Show displays the modal
func (m *DetailModal) Show() {
	m.visible = true
	m.focused = true
}

// Hide hides the modal
func (m *DetailModal) Hide() {
	m.visible = false
	m.focused = false
}

// IsVisible returns whether the modal is visible
func (m *DetailModal) IsVisible() bool {
	return m.visible
}

// SetContent updates the modal content from a StatCard
func (m *DetailModal) SetContent(card *StatCard) {
	m.title = card.title
	m.value = card.value
	m.subtitle = card.subtitle
	m.change = card.change
	m.changePct = card.changePct
	m.trend = card.trend
	m.color = card.color
	m.trendColor = card.trendColor
}

// renderModalContent renders the modal content box
func (m *DetailModal) renderModalContent(width, height int) string {
	var b strings.Builder

	contentWidth := width - 4 // Account for borders and padding

	// Top border
	b.WriteString("╔")
	b.WriteString(strings.Repeat("═", width-2))
	b.WriteString("╗\n")

	// Title bar with close hint
	b.WriteString("║ ")
	titleLine := fmt.Sprintf("\033[1m%s\033[0m", m.title)
	closeHint := "[ESC to close]"
	titleLen := len(m.title) // Visible length without ANSI
	spacing := contentWidth - titleLen - len(closeHint)
	if spacing < 1 {
		spacing = 1
	}
	b.WriteString(titleLine)
	b.WriteString(strings.Repeat(" ", spacing))
	b.WriteString("\033[90m" + closeHint + "\033[0m") // Gray
	b.WriteString(" ║\n")

	// Separator
	b.WriteString("╠")
	b.WriteString(strings.Repeat("═", width-2))
	b.WriteString("╣\n")

	// Empty line
	m.writeModalLine(&b, "", contentWidth)

	// Value (large display)
	valueLine := fmt.Sprintf("  \033[1;36m%s\033[0m", m.value) // Bold cyan
	m.writeModalLine(&b, valueLine, contentWidth)

	// Empty line
	m.writeModalLine(&b, "", contentWidth)

	// Change indicator
	if m.change != 0 || m.changePct != 0 {
		var changeColor, arrow string
		if m.change > 0 {
			changeColor = "\033[32m" // Green
			arrow = "↑"
		} else if m.change < 0 {
			changeColor = "\033[31m" // Red
			arrow = "↓"
		} else {
			changeColor = "\033[37m" // White
			arrow = "→"
		}
		changeStr := fmt.Sprintf("  %s%s %d (%+.1f%%)%s",
			changeColor, arrow, abs(m.change), m.changePct, "\033[0m")
		m.writeModalLine(&b, changeStr, contentWidth)
		m.writeModalLine(&b, "", contentWidth)
	}

	// Subtitle
	if m.subtitle != "" {
		subtitleLine := fmt.Sprintf("  \033[90m%s\033[0m", m.subtitle) // Gray
		m.writeModalLine(&b, subtitleLine, contentWidth)
		m.writeModalLine(&b, "", contentWidth)
	}

	// Trend section
	var trendLines []string
	if len(m.trend) > 0 {
		m.writeModalLine(&b, "  Trend (Last 30 data points):", contentWidth)
		m.writeModalLine(&b, "", contentWidth)

		// Render large trend graph
		trendLines = m.renderLargeTrendGraph(contentWidth - 4)
		for _, line := range trendLines {
			m.writeModalLine(&b, "  "+line, contentWidth)
		}
		m.writeModalLine(&b, "", contentWidth)

		// Statistics
		minVal, maxVal, avg := m.calculateStats()
		statsLine := fmt.Sprintf("  Min: %.1f  Max: %.1f  Avg: %.1f", minVal, maxVal, avg)
		m.writeModalLine(&b, statsLine, contentWidth)
		m.writeModalLine(&b, "", contentWidth)
	}

	// Historical data if available
	if len(m.history) > 0 {
		m.writeModalLine(&b, "  Recent History:", contentWidth)
		m.writeModalLine(&b, "", contentWidth)
		for i, entry := range m.history {
			if i >= 5 { // Show only 5 most recent
				break
			}
			m.writeModalLine(&b, "  "+entry, contentWidth)
		}
	}

	// Fill remaining height
	currentLines := 8 + // Fixed lines (borders, title, value, etc)
		len(trendLines) +
		min(len(m.history), 5)

	if m.change != 0 || m.changePct != 0 {
		currentLines += 2
	}
	if m.subtitle != "" {
		currentLines += 2
	}
	if len(m.trend) > 0 {
		currentLines += 6 // Trend section
	}

	for currentLines < height-1 {
		m.writeModalLine(&b, "", contentWidth)
		currentLines++
	}

	// Bottom border
	b.WriteString("╚")
	b.WriteString(strings.Repeat("═", width-2))
	b.WriteString("╝")

	return b.String()
}

// writeModalLine writes a line with proper border and padding
func (m *DetailModal) writeModalLine(b *strings.Builder, content string, width int) {
	b.WriteString("║ ")

	// Calculate visible length (excluding ANSI codes)
	visibleLen := m.visibleLength(content)

	b.WriteString(content)
	if visibleLen < width {
		b.WriteString(strings.Repeat(" ", width-visibleLen))
	}
	b.WriteString(" ║\n")
}

// renderLargeTrendGraph renders a multi-line trend graph
func (m *DetailModal) renderLargeTrendGraph(width int) []string {
	if len(m.trend) == 0 {
		return []string{}
	}

	height := 8 // Graph height in lines
	lines := make([]string, height)

	// Find min and max
	min, max := m.trend[0], m.trend[0]
	for _, v := range m.trend {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Normalize to 0-1 range
	normalize := func(v float64) float64 {
		if max == min {
			return 0.5
		}
		return (v - min) / (max - min)
	}

	// Determine how many data points to show
	pointsToShow := len(m.trend)
	if pointsToShow > width {
		pointsToShow = width
	}

	// Sample data if needed
	step := 1
	if len(m.trend) > width {
		step = len(m.trend) / width
	}

	// Build graph from top to bottom
	for row := 0; row < height; row++ {
		var line strings.Builder
		threshold := 1.0 - float64(row)/float64(height-1)

		line.WriteString("\033[38;2;76;175;80m") // Green color

		for i := 0; i < pointsToShow; i++ {
			dataIndex := i * step
			if dataIndex >= len(m.trend) {
				dataIndex = len(m.trend) - 1
			}

			normalizedValue := normalize(m.trend[dataIndex])

			if normalizedValue >= threshold {
				if row == 0 {
					line.WriteString("▀")
				} else if normalizedValue >= threshold+(1.0/float64(height)) {
					line.WriteString("█")
				} else {
					line.WriteString("▄")
				}
			} else {
				line.WriteString(" ")
			}
		}

		line.WriteString("\033[0m")
		lines[row] = line.String()
	}

	return lines
}

// calculateStats calculates min, max, and average from trend data
func (m *DetailModal) calculateStats() (min, max, avg float64) {
	if len(m.trend) == 0 {
		return 0, 0, 0
	}

	min, max = m.trend[0], m.trend[0]
	sum := 0.0

	for _, v := range m.trend {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}

	avg = sum / float64(len(m.trend))
	return min, max, avg
}

// visibleLength calculates the visible length of a string (excluding ANSI codes)
func (m *DetailModal) visibleLength(str string) int {
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
