package display

import (
	"strings"

	colorpkg "github.com/SCKelemen/color"
	"github.com/charmbracelet/lipgloss"
)

// GradientText renders text with a smooth multi-stop gradient across characters.
type GradientText struct {
	text  string
	stops []string
	width int
}

// GradientTextOption configures a GradientText component.
type GradientTextOption func(*GradientText)

// NewGradientText creates a new GradientText component.
func NewGradientText(text string, stops []string) *GradientText {
	return &GradientText{
		text:  text,
		stops: append([]string(nil), stops...),
		width: 0,
	}
}

// WithGradientTextWidth sets the maximum number of characters to render.
func WithGradientTextWidth(w int) GradientTextOption {
	return func(g *GradientText) {
		if w >= 0 {
			g.width = w
		}
	}
}

// View renders the gradient text.
func (g *GradientText) View() string {
	if g == nil {
		return ""
	}

	text := g.text
	if g.width > 0 {
		runes := []rune(text)
		if len(runes) > g.width {
			text = string(runes[:g.width])
		}
	}

	return RenderGradient(text, g.stops)
}

// RenderGradient renders text using a smooth OKLCH multi-stop gradient.
func RenderGradient(text string, stops []string) string {
	if text == "" {
		return ""
	}

	runes := []rune(text)
	if len(runes) == 0 || len(stops) == 0 {
		return text
	}

	gradientStops := make([]colorpkg.GradientStop, 0, len(stops))
	for i, stop := range stops {
		parsed, err := colorpkg.ParseColor(stop)
		if err != nil {
			return text
		}

		position := 0.0
		if len(stops) > 1 {
			position = float64(i) / float64(len(stops)-1)
		}

		gradientStops = append(gradientStops, colorpkg.GradientStop{
			Color:    parsed,
			Position: position,
		})
	}

	colors := colorpkg.GradientMultiStop(gradientStops, len(runes), colorpkg.GradientOKLCH)
	if len(colors) != len(runes) {
		return text
	}

	var b strings.Builder
	for i, r := range runes {
		hex := colorpkg.ToLipglossColor(colors[i])
		char := string(r)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Render(char))
	}

	return b.String()
}

// LovableGradient renders text with Lovable's signature gradient.
func LovableGradient(text string) string {
	return RenderGradient(text, []string{"#FF8E63", "#FF7EB0", "#4B73FF"})
}
