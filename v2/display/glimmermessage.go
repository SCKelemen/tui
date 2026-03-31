package display

import (
	"fmt"
	"math"
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	defaultGlimmerMessageWidth = 28
	defaultGlimmerMessageLines = 3
	glimmerWaveRadius          = 5.0
)

// GlimmerMessageOption configures a GlimmerMessage component.
type GlimmerMessageOption func(*GlimmerMessage)

// GlimmerMessage renders an animated thinking/streaming shimmer placeholder.
type GlimmerMessage struct {
	width  int
	lines  int
	label  string
	color  string
	tokens *design.DesignTokens

	focused bool
	offset  int
}

// GlimmerMessageTickMsg drives GlimmerMessage animation frames.
type GlimmerMessageTickMsg time.Time

// NewGlimmerMessage creates a new animated GlimmerMessage component.
func NewGlimmerMessage(opts ...GlimmerMessageOption) *GlimmerMessage {
	g := &GlimmerMessage{
		width: defaultGlimmerMessageWidth,
		lines: defaultGlimmerMessageLines,
		color: "#6B7280",
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

// WithGlimmerMessageWidth sets the number of columns for each glimmer line.
func WithGlimmerMessageWidth(width int) GlimmerMessageOption {
	return func(g *GlimmerMessage) {
		if width > 0 {
			g.width = width
		}
	}
}

// WithGlimmerMessageLines sets the number of shimmer lines.
func WithGlimmerMessageLines(lines int) GlimmerMessageOption {
	return func(g *GlimmerMessage) {
		if lines > 0 {
			g.lines = lines
		}
	}
}

// WithGlimmerMessageLabel sets an optional label rendered above shimmer lines.
func WithGlimmerMessageLabel(label string) GlimmerMessageOption {
	return func(g *GlimmerMessage) {
		g.label = strings.TrimSpace(label)
	}
}

// WithGlimmerMessageColor sets the base shimmer color (hex preferred).
func WithGlimmerMessageColor(color string) GlimmerMessageOption {
	return func(g *GlimmerMessage) {
		if strings.TrimSpace(color) != "" {
			g.color = color
		}
	}
}

// WithGlimmerMessageDesignTokens applies design-system tokens.
func WithGlimmerMessageDesignTokens(tokens *design.DesignTokens) GlimmerMessageOption {
	return func(g *GlimmerMessage) {
		if tokens == nil {
			return
		}

		g.tokens = tokens
		g.applyDesignTokens(tokens)
	}
}

// Init starts the animation tick loop.
func (g *GlimmerMessage) Init() tea.Cmd {
	return g.tick()
}

// Update advances the shimmer wave on tick messages.
func (g *GlimmerMessage) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg.(type) {
	case GlimmerMessageTickMsg:
		g.offset++
		return g, g.tick()
	}

	return g, nil
}

// View renders label and animated shimmer lines.
func (g *GlimmerMessage) View() string {
	if g.width <= 0 || g.lines <= 0 {
		if g.label == "" {
			return ""
		}
		return g.renderLabel()
	}

	highlight := adjustHexBrightness(g.color, 1.45)
	cycleWidth := g.width + int(glimmerWaveRadius*2) + 1
	if cycleWidth <= 0 {
		cycleWidth = 1
	}

	parts := make([]string, 0, g.lines+1)
	if g.label != "" {
		parts = append(parts, g.renderLabel())
	}

	for line := 0; line < g.lines; line++ {
		center := float64((g.offset+line*3)%cycleWidth) - glimmerWaveRadius

		var b strings.Builder
		for col := 0; col < g.width; col++ {
			distance := math.Abs(float64(col) - center)
			intensity := glimmerIntensity(distance)
			glyph := glimmerGlyph(intensity)

			color := interpolateHex(g.color, highlight, intensity)
			if fg := style.Fg(color); fg != "" {
				b.WriteString(fg)
			}
			b.WriteString(glyph)
			b.WriteString(style.ANSIReset)
		}

		parts = append(parts, b.String())
	}

	return strings.Join(parts, "\n")
}

// Focus marks the component as focused.
func (g *GlimmerMessage) Focus() {
	g.focused = true
}

// Blur marks the component as unfocused.
func (g *GlimmerMessage) Blur() {
	g.focused = false
}

// Focused reports whether this component is focused.
func (g *GlimmerMessage) Focused() bool {
	return g.focused
}

func (g *GlimmerMessage) tick() tea.Cmd {
	return tea.Tick(shimmerTickInterval, func(ts time.Time) tea.Msg {
		return GlimmerMessageTickMsg(ts)
	})
}

func (g *GlimmerMessage) renderLabel() string {
	labelColor := adjustHexBrightness(g.color, 1.15)
	if fg := style.Fg(labelColor); fg != "" {
		return fmt.Sprintf("%s%s%s", fg, g.label, style.ANSIReset)
	}
	return g.label
}

func (g *GlimmerMessage) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}

	switch {
	case strings.TrimSpace(tokens.MutedColor) != "":
		g.color = tokens.MutedColor
	case strings.TrimSpace(tokens.SurfaceRaised) != "":
		g.color = tokens.SurfaceRaised
	case strings.TrimSpace(tokens.Color) != "":
		g.color = tokens.Color
	}
}

func glimmerIntensity(distance float64) float64 {
	if distance > glimmerWaveRadius {
		return 0.08
	}

	v := 1.0 - (distance / glimmerWaveRadius)
	if v < 0 {
		v = 0
	}
	return 0.08 + (v * v)
}

func glimmerGlyph(intensity float64) string {
	switch {
	case intensity >= 0.82:
		return "█"
	case intensity >= 0.58:
		return "▓"
	case intensity >= 0.34:
		return "▒"
	default:
		return "░"
	}
}

var _ tui.Component = (*GlimmerMessage)(nil)
