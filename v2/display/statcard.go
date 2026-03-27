package display

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// StatCard displays a single metric with title, value, change indicator, and optional
// sparkline trend visualization.
type StatCard struct {
	width    int
	height   int
	focused  bool
	selected bool
	tokens   *design.DesignTokens

	title      string
	value      string
	subtitle   string
	change     int
	changePct  float64
	trend      []float64
	color      string
	trendColor string
}

// StatCardOption configures a StatCard.
type StatCardOption func(*StatCard)

// WithTitle sets the card title.
func WithTitle(title string) StatCardOption {
	return func(s *StatCard) {
		s.title = title
	}
}

// WithValue sets the main value to display.
func WithValue(value string) StatCardOption {
	return func(s *StatCard) {
		s.value = value
	}
}

// WithSubtitle sets the subtitle/description.
func WithSubtitle(subtitle string) StatCardOption {
	return func(s *StatCard) {
		s.subtitle = subtitle
	}
}

// WithChange sets the change value and percentage.
func WithChange(change int, changePct float64) StatCardOption {
	return func(s *StatCard) {
		s.change = change
		s.changePct = changePct
	}
}

// WithTrend sets the sparkline trend data.
func WithTrend(trend []float64) StatCardOption {
	return func(s *StatCard) {
		s.trend = trend
	}
}

// WithColor sets the accent color.
func WithColor(color string) StatCardOption {
	return func(s *StatCard) {
		s.color = color
	}
}

// WithTrendColor sets the trend line color.
func WithTrendColor(color string) StatCardOption {
	return func(s *StatCard) {
		s.trendColor = color
	}
}

// NewStatCard creates a new stat card with the given configuration options.
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

// Init initializes the stat card.
func (s *StatCard) Init() tea.Cmd {
	return nil
}

// Update handles Bubble Tea messages.
func (s *StatCard) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	}

	return s, nil
}

// View renders the stat card.
func (s *StatCard) View() string {
	if s.width < 2 || s.height < 2 {
		return ""
	}
	return s.renderSimple()
}

// Focus is called when this component receives focus.
func (s *StatCard) Focus() {
	s.focused = true
}

// Blur is called when this component loses focus.
func (s *StatCard) Blur() {
	s.focused = false
}

// Focused returns whether this component is currently focused.
func (s *StatCard) Focused() bool {
	return s.focused
}

// Select marks the card as selected (for drill-down).
func (s *StatCard) Select() {
	s.selected = true
}

// Deselect marks the card as not selected.
func (s *StatCard) Deselect() {
	s.selected = false
}

// IsSelected returns whether this card is selected.
func (s *StatCard) IsSelected() bool {
	return s.selected
}

type borderStyle struct {
	topLeft, topRight, bottomLeft, bottomRight, horizontal, vertical string
	color                                                            string
}

func (s *StatCard) getBorderStyle() borderStyle {
	if s.focused {
		return borderStyle{
			topLeft: "╔", topRight: "╗",
			bottomLeft: "╚", bottomRight: "╝",
			horizontal: "═", vertical: "║",
			color: style.ANSICyan,
		}
	} else if s.selected {
		return borderStyle{
			topLeft: "┏", topRight: "┓",
			bottomLeft: "┗", bottomRight: "┛",
			horizontal: "━", vertical: "┃",
			color: style.ANSIYellow,
		}
	}

	return borderStyle{
		topLeft: "┌", topRight: "┐",
		bottomLeft: "└", bottomRight: "┘",
		horizontal: "─", vertical: "│",
		color: "",
	}
}

func (s *StatCard) writeBorder(b *strings.Builder, char string, st borderStyle) {
	if st.color != "" {
		b.WriteString(st.color)
	}
	b.WriteString(char)
	if st.color != "" {
		b.WriteString(style.ANSIReset)
	}
}

func (s *StatCard) renderSimple() string {
	var b strings.Builder

	contentWidth := s.width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	st := s.getBorderStyle()

	s.writeBorder(&b, st.topLeft, st)
	s.writeBorder(&b, strings.Repeat(st.horizontal, s.width-2), st)
	s.writeBorder(&b, st.topRight, st)
	b.WriteString("\n")

	s.writeBorder(&b, st.vertical, st)
	b.WriteString(" ")
	b.WriteString(s.truncate(s.title, contentWidth))
	b.WriteString(" ")
	s.writeBorder(&b, st.vertical, st)
	b.WriteString("\n")

	s.writeBorder(&b, st.vertical, st)
	b.WriteString(" ")
	valueStr := style.ANSIBold + s.value + style.ANSIReset
	b.WriteString(valueStr)
	visibleValueLen := s.visibleLength(valueStr)
	if visibleValueLen < contentWidth {
		b.WriteString(strings.Repeat(" ", contentWidth-visibleValueLen))
	}
	b.WriteString(" ")
	s.writeBorder(&b, st.vertical, st)
	b.WriteString("\n")

	if s.change != 0 || s.changePct != 0 {
		s.writeBorder(&b, st.vertical, st)
		b.WriteString(" ")
		changeStr := s.renderChange()
		b.WriteString(changeStr)
		visibleLen := s.visibleLength(changeStr)
		if visibleLen < contentWidth {
			b.WriteString(strings.Repeat(" ", contentWidth-visibleLen))
		}
		b.WriteString(" ")
		s.writeBorder(&b, st.vertical, st)
		b.WriteString("\n")
	}

	if s.subtitle != "" {
		s.writeBorder(&b, st.vertical, st)
		b.WriteString(" ")
		b.WriteString(s.truncate(s.subtitle, contentWidth))
		b.WriteString(" ")
		s.writeBorder(&b, st.vertical, st)
		b.WriteString("\n")
	}

	if len(s.trend) > 0 {
		s.writeBorder(&b, st.vertical, st)
		b.WriteString(" ")
		sparkline := s.renderSparkline(contentWidth)
		b.WriteString(sparkline)
		b.WriteString(" ")
		s.writeBorder(&b, st.vertical, st)
		b.WriteString("\n")
	}

	currentHeight := 3
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
		s.writeBorder(&b, st.vertical, st)
		b.WriteString(strings.Repeat(" ", s.width-2))
		s.writeBorder(&b, st.vertical, st)
		b.WriteString("\n")
		currentHeight++
	}

	s.writeBorder(&b, st.bottomLeft, st)
	s.writeBorder(&b, strings.Repeat(st.horizontal, s.width-2), st)
	s.writeBorder(&b, st.bottomRight, st)
	b.WriteString("\n")

	return b.String()
}

func (s *StatCard) renderChange() string {
	var changeColor string
	var arrow string

	if s.change > 0 {
		changeColor = style.ANSIGreen
		arrow = "↑"
	} else if s.change < 0 {
		changeColor = style.ANSIRed
		arrow = "↓"
	} else {
		changeColor = style.ANSIWhite
		arrow = "→"
	}

	return fmt.Sprintf("%s%s %d (%.1f%%)%s", changeColor, arrow, abs(s.change), s.changePct, style.ANSIReset)
}

func (s *StatCard) renderSparkline(width int) string {
	if len(s.trend) == 0 {
		return ""
	}

	blocks := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	min, max := s.trend[0], s.trend[0]
	for _, v := range s.trend {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	normalize := func(v float64) int {
		if max == min {
			return 4
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
	pointsToShow := len(s.trend)
	if pointsToShow > width {
		pointsToShow = width
	}
	step := 1
	if len(s.trend) > width {
		step = len(s.trend) / width
	}

	trendAnsi := style.ANSIColorFromHex(s.trendColor)
	if trendAnsi == "" {
		trendAnsi = style.ANSIGreen
	}
	b.WriteString(trendAnsi)
	for i := 0; i < pointsToShow; i++ {
		dataIndex := i * step
		if dataIndex >= len(s.trend) {
			dataIndex = len(s.trend) - 1
		}
		blockIndex := normalize(s.trend[dataIndex])
		b.WriteString(blocks[blockIndex])
	}
	b.WriteString(style.ANSIReset)

	if pointsToShow < width {
		b.WriteString(strings.Repeat(" ", width-pointsToShow))
	}

	return b.String()
}

func (s *StatCard) truncate(str string, width int) string {
	if width <= 0 {
		return ""
	}
	truncated := style.Truncate(str, width, "...")
	pad := width - style.StringWidth(truncated)
	if pad > 0 {
		return truncated + strings.Repeat(" ", pad)
	}
	return truncated
}

func (s *StatCard) visibleLength(str string) int {
	return style.StringWidth(stripANSI(str))
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

var _ tui.Component = (*StatCard)(nil)
