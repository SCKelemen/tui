package chart

import (
	"fmt"
	"math"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// UsageSegment is one labeled value in a segmented usage bar.
type UsageSegment struct {
	Label string
	Value float64
	Color string
}

// MemoryUsageIndicatorOption configures a MemoryUsageIndicator.
type MemoryUsageIndicatorOption func(*MemoryUsageIndicator)

// WithMemoryUsageIndicatorWidth sets render width.
func WithMemoryUsageIndicatorWidth(width int) MemoryUsageIndicatorOption {
	return func(m *MemoryUsageIndicator) {
		if width > 0 {
			m.width = width
		}
	}
}

// WithMemoryUsageIndicatorTotal sets total capacity.
func WithMemoryUsageIndicatorTotal(total float64) MemoryUsageIndicatorOption {
	return func(m *MemoryUsageIndicator) {
		if total > 0 {
			m.total = total
		}
	}
}

// WithMemoryUsageIndicatorLabel sets optional heading label.
func WithMemoryUsageIndicatorLabel(label string) MemoryUsageIndicatorOption {
	return func(m *MemoryUsageIndicator) {
		m.label = strings.TrimSpace(label)
	}
}

// WithMemoryUsageIndicatorDesignTokens applies design-system tokens.
func WithMemoryUsageIndicatorDesignTokens(tokens *design.DesignTokens) MemoryUsageIndicatorOption {
	return func(m *MemoryUsageIndicator) {
		if tokens != nil {
			m.designTokens = tokens
		}
	}
}

// MemoryUsageIndicator renders a stacked usage bar and legend.
type MemoryUsageIndicator struct {
	segments     []UsageSegment
	total        float64
	label        string
	width        int
	windowWidth  int
	focused      bool
	designTokens *design.DesignTokens
}

// NewMemoryUsageIndicator creates a new MemoryUsageIndicator.
func NewMemoryUsageIndicator(segments []UsageSegment, opts ...MemoryUsageIndicatorOption) *MemoryUsageIndicator {
	m := &MemoryUsageIndicator{
		segments:     append([]UsageSegment(nil), segments...),
		total:        0,
		label:        "Memory usage",
		width:        32,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(m)
	}

	if m.total <= 0 {
		m.total = m.segmentSum()
	}
	if m.total <= 0 {
		m.total = 1
	}

	return m
}

// Init initializes the component.
func (m *MemoryUsageIndicator) Init() tea.Cmd { return nil }

// Update handles resize messages.
func (m *MemoryUsageIndicator) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		return m, nil
	}
	return m, nil
}

// View renders segmented bar and legend.
func (m *MemoryUsageIndicator) View() string {
	if len(m.segments) == 0 {
		return ""
	}

	barWidth := m.effectiveWidth()
	if barWidth <= 0 {
		barWidth = 32
	}

	var out []string
	if m.label != "" {
		out = append(out, m.label)
	}

	out = append(out, m.renderBar(barWidth))
	out = append(out, m.renderLegend(barWidth)...)
	return strings.Join(out, "\n")
}

// Focus marks the component focused.
func (m *MemoryUsageIndicator) Focus() { m.focused = true }

// Blur marks the component unfocused.
func (m *MemoryUsageIndicator) Blur() { m.focused = false }

// Focused reports whether this component is focused.
func (m *MemoryUsageIndicator) Focused() bool { return m.focused }

func (m *MemoryUsageIndicator) effectiveWidth() int {
	if m.width > 0 {
		return m.width
	}
	return m.windowWidth
}

func (m *MemoryUsageIndicator) renderBar(width int) string {
	remaining := width
	parts := make([]string, 0, len(m.segments))

	for i, seg := range m.segments {
		share := seg.Value / m.total
		if share < 0 {
			share = 0
		}
		segWidth := int(math.Round(share * float64(width)))
		if i == len(m.segments)-1 {
			segWidth = remaining
		}
		if segWidth > remaining {
			segWidth = remaining
		}
		if segWidth < 0 {
			segWidth = 0
		}
		remaining -= segWidth

		if segWidth == 0 {
			continue
		}
		color := style.Fg(seg.Color)
		if color == "" {
			color = m.defaultSegmentColor(i)
		}
		parts = append(parts, color+strings.Repeat("█", segWidth)+style.ANSIReset)
	}

	if remaining > 0 {
		parts = append(parts, style.ANSIDim+strings.Repeat("░", remaining)+style.ANSIReset)
	}

	return strings.Join(parts, "")
}

func (m *MemoryUsageIndicator) renderLegend(width int) []string {
	lines := make([]string, 0, len(m.segments))
	for i, seg := range m.segments {
		percent := 0.0
		if m.total > 0 {
			percent = (seg.Value / m.total) * 100
		}
		color := style.Fg(seg.Color)
		if color == "" {
			color = m.defaultSegmentColor(i)
		}
		line := fmt.Sprintf("%s■%s %s %.0f%%", color, style.ANSIReset, seg.Label, percent)
		if width > 0 {
			line = style.Pad(style.Truncate(line, width, "…"), width)
		}
		lines = append(lines, line)
	}
	return lines
}

func (m *MemoryUsageIndicator) segmentSum() float64 {
	total := 0.0
	for _, seg := range m.segments {
		total += seg.Value
	}
	return total
}

func (m *MemoryUsageIndicator) defaultSegmentColor(index int) string {
	defaults := []string{style.ANSICyan, style.ANSIYellow, style.ANSIGreen, style.ANSIRed}
	if index < len(defaults) {
		return defaults[index]
	}
	return style.ANSIWhite
}

var _ tui.Component = (*MemoryUsageIndicator)(nil)
