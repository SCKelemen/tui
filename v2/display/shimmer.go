package display

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	defaultShimmerWidth = 28
	defaultShimmerLines = 3
	shimmerTickInterval = 80 * time.Millisecond
	shimmerWaveRadius   = 4.0
)

// ShimmerTickMsg drives shimmer animation frames.
type ShimmerTickMsg time.Time

// Shimmer renders an animated skeleton loading block.
type Shimmer struct {
	width  int
	lines  int
	color  string
	tokens *design.DesignTokens

	focused  bool
	position int
}

// ShimmerOption configures a Shimmer component.
type ShimmerOption func(*Shimmer)

// NewShimmer creates a shimmer/skeleton loading component.
func NewShimmer(opts ...ShimmerOption) *Shimmer {
	s := &Shimmer{
		width: defaultShimmerWidth,
		lines: defaultShimmerLines,
		color: "#6B7280",
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// WithShimmerWidth sets the number of placeholder columns.
func WithShimmerWidth(width int) ShimmerOption {
	return func(s *Shimmer) {
		if width > 0 {
			s.width = width
		}
	}
}

// WithShimmerLines sets the number of skeleton lines.
func WithShimmerLines(lines int) ShimmerOption {
	return func(s *Shimmer) {
		if lines > 0 {
			s.lines = lines
		}
	}
}

// WithShimmerColor sets the base shimmer color (hex).
func WithShimmerColor(color string) ShimmerOption {
	return func(s *Shimmer) {
		if strings.TrimSpace(color) != "" {
			s.color = color
		}
	}
}

// WithShimmerDesignTokens applies design-system tokens.
func WithShimmerDesignTokens(tokens *design.DesignTokens) ShimmerOption {
	return func(s *Shimmer) {
		if tokens == nil {
			return
		}

		s.tokens = tokens
		s.applyDesignTokens(tokens)
	}
}

// Init starts the shimmer animation ticker.
func (s *Shimmer) Init() tea.Cmd {
	return s.tick()
}

// Update advances the shimmer position on each animation tick.
func (s *Shimmer) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg.(type) {
	case ShimmerTickMsg:
		s.position++
		return s, s.tick()
	}

	return s, nil
}

// View renders the skeleton block with a moving brightness wave.
func (s *Shimmer) View() string {
	if s.width <= 0 || s.lines <= 0 {
		return ""
	}

	lineWidth := s.width
	cycleWidth := lineWidth + int(shimmerWaveRadius*2) + 1
	if cycleWidth <= 0 {
		cycleWidth = 1
	}

	var b strings.Builder
	for line := 0; line < s.lines; line++ {
		center := float64((s.position+line*2)%cycleWidth) - shimmerWaveRadius

		for col := 0; col < lineWidth; col++ {
			distance := math.Abs(float64(col) - center)
			glyph, brightness := shimmerGlyphAndBrightness(distance)
			color := adjustHexBrightness(s.color, brightness)

			if fg := style.Fg(color); fg != "" {
				b.WriteString(fg)
			}
			b.WriteString(glyph)
			b.WriteString(style.ANSIReset)
		}

		if line < s.lines-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}

// Focus marks the component as focused.
func (s *Shimmer) Focus() {
	s.focused = true
}

// Blur marks the component as unfocused.
func (s *Shimmer) Blur() {
	s.focused = false
}

// Focused reports whether this component is focused.
func (s *Shimmer) Focused() bool {
	return s.focused
}

func (s *Shimmer) tick() tea.Cmd {
	return tea.Tick(shimmerTickInterval, func(ts time.Time) tea.Msg {
		return ShimmerTickMsg(ts)
	})
}

func (s *Shimmer) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}

	switch {
	case strings.TrimSpace(tokens.MutedColor) != "":
		s.color = tokens.MutedColor
	case strings.TrimSpace(tokens.SurfaceRaised) != "":
		s.color = tokens.SurfaceRaised
	case strings.TrimSpace(tokens.Color) != "":
		s.color = tokens.Color
	}
}

func shimmerGlyphAndBrightness(distance float64) (string, float64) {
	switch {
	case distance <= 0.6:
		return "█", 1.35
	case distance <= 1.4:
		return "▓", 1.2
	case distance <= 2.4:
		return "▒", 1.05
	case distance <= shimmerWaveRadius:
		return "░", 0.9
	default:
		return "█", 0.62
	}
}

func adjustHexBrightness(hex string, factor float64) string {
	r, g, b, ok := parseHexRGB(hex)
	if !ok {
		return hex
	}

	r = clampColorInt(int(math.Round(float64(r) * factor)))
	g = clampColorInt(int(math.Round(float64(g) * factor)))
	b = clampColorInt(int(math.Round(float64(b) * factor)))

	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func parseHexRGB(value string) (int, int, int, bool) {
	hex := strings.TrimPrefix(strings.TrimSpace(value), "#")

	switch len(hex) {
	case 3:
		r, errR := strconv.ParseUint(strings.Repeat(string(hex[0]), 2), 16, 8)
		g, errG := strconv.ParseUint(strings.Repeat(string(hex[1]), 2), 16, 8)
		b, errB := strconv.ParseUint(strings.Repeat(string(hex[2]), 2), 16, 8)
		if errR != nil || errG != nil || errB != nil {
			return 0, 0, 0, false
		}
		return int(r), int(g), int(b), true
	case 6:
		r, errR := strconv.ParseUint(hex[0:2], 16, 8)
		g, errG := strconv.ParseUint(hex[2:4], 16, 8)
		b, errB := strconv.ParseUint(hex[4:6], 16, 8)
		if errR != nil || errG != nil || errB != nil {
			return 0, 0, 0, false
		}
		return int(r), int(g), int(b), true
	default:
		return 0, 0, 0, false
	}
}

func clampColorInt(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

var _ tui.Component = (*Shimmer)(nil)
