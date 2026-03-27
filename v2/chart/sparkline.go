package chart

import (
	"math"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

var sparklineBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Sparkline renders a compact trend line using Unicode block characters.
type Sparkline struct {
	data      []float64
	width     int
	minVal    float64
	maxVal    float64
	autoRange bool
	label     string

	focused      bool
	designTokens *design.DesignTokens
}

// SparklineOption configures a Sparkline.
type SparklineOption func(*Sparkline)

// WithSparklineData sets the sparkline series values.
func WithSparklineData(data []float64) SparklineOption {
	return func(s *Sparkline) {
		s.data = append([]float64(nil), data...)
	}
}

// WithSparklineWidth sets the sparkline width in characters.
func WithSparklineWidth(width int) SparklineOption {
	return func(s *Sparkline) {
		if width > 0 {
			s.width = width
		}
	}
}

// WithSparklineRange sets a fixed value range and disables auto-range.
func WithSparklineRange(minVal, maxVal float64) SparklineOption {
	return func(s *Sparkline) {
		s.minVal = minVal
		s.maxVal = maxVal
		s.autoRange = false
	}
}

// WithSparklineAutoRange toggles dynamic min/max range calculation.
func WithSparklineAutoRange(auto bool) SparklineOption {
	return func(s *Sparkline) {
		s.autoRange = auto
	}
}

// WithSparklineLabel sets a right-side label.
func WithSparklineLabel(label string) SparklineOption {
	return func(s *Sparkline) {
		s.label = label
	}
}

// WithSparklineDesignTokens applies design-system tokens.
func WithSparklineDesignTokens(tokens *design.DesignTokens) SparklineOption {
	return func(s *Sparkline) {
		if tokens == nil {
			return
		}
		s.designTokens = tokens
	}
}

// WithSparklineTheme applies a named design-system theme.
func WithSparklineTheme(theme string) SparklineOption {
	return func(s *Sparkline) {
		s.designTokens = style.DesignTokensForTheme(theme)
	}
}

// NewSparkline creates a new sparkline component.
func NewSparkline(data []float64, opts ...SparklineOption) *Sparkline {
	s := &Sparkline{
		data:         append([]float64(nil), data...),
		width:        40,
		autoRange:    true,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Init initializes the sparkline.
func (s *Sparkline) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (s *Sparkline) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	return s, nil
}

// View renders the sparkline and optional label.
func (s *Sparkline) View() string {
	series := s.lastPoints()
	if len(series) == 0 {
		if strings.TrimSpace(s.label) != "" {
			return s.label
		}
		return ""
	}

	minVal, maxVal := s.rangeFor(series)
	delta := maxVal - minVal

	var b strings.Builder
	b.Grow(len(series))

	for _, v := range series {
		idx := 0
		if delta > 0 {
			norm := (v - minVal) / delta
			if norm < 0 {
				norm = 0
			}
			if norm > 1 {
				norm = 1
			}
			idx = int(math.Round(norm * float64(len(sparklineBlocks)-1)))
		}
		b.WriteRune(sparklineBlocks[idx])
	}

	spark := b.String()
	if strings.TrimSpace(s.label) != "" {
		return spark + " " + s.label
	}
	return spark
}

// Focus marks the component as focused.
func (s *Sparkline) Focus() {
	s.focused = true
}

// Blur marks the component as unfocused.
func (s *Sparkline) Blur() {
	s.focused = false
}

// Focused returns whether the component is focused.
func (s *Sparkline) Focused() bool {
	return s.focused
}

// SetData replaces the sparkline data series.
func (s *Sparkline) SetData(data []float64) {
	s.data = append([]float64(nil), data...)
}

// SetLabel sets the right-side label.
func (s *Sparkline) SetLabel(label string) {
	s.label = label
}

func (s *Sparkline) lastPoints() []float64 {
	if len(s.data) <= s.width {
		return s.data
	}
	start := len(s.data) - s.width
	return s.data[start:]
}

func (s *Sparkline) rangeFor(series []float64) (float64, float64) {
	if !s.autoRange {
		if s.maxVal < s.minVal {
			return s.maxVal, s.minVal
		}
		return s.minVal, s.maxVal
	}

	minVal := series[0]
	maxVal := series[0]
	for _, v := range series[1:] {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	return minVal, maxVal
}
