package style

import (
	"fmt"
	"math"
	"strings"

	colorpkg "github.com/SCKelemen/color"
	design "github.com/SCKelemen/design-system"
)

// TokenToColor parses a design-system hex color token into a color object.
func TokenToColor(hexToken string) (colorpkg.Color, error) {
	token := strings.TrimSpace(hexToken)
	if token == "" {
		return nil, fmt.Errorf("empty color token")
	}
	if !strings.HasPrefix(token, "#") {
		token = "#" + token
	}
	return colorpkg.ParseColor(token)
}

// ColorToANSIFg converts a color object to a 24-bit ANSI foreground escape sequence.
func ColorToANSIFg(c colorpkg.Color) string {
	if c == nil {
		return ""
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", toByte(r), toByte(g), toByte(b))
}

// ColorToANSIBg converts a color object to a 24-bit ANSI background escape sequence.
func ColorToANSIBg(c colorpkg.Color) string {
	if c == nil {
		return ""
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", toByte(r), toByte(g), toByte(b))
}

// ThemeAccentColor returns the accent token as a color object.
func ThemeAccentColor(dt *design.DesignTokens) colorpkg.Color {
	if dt == nil {
		return colorpkg.RGB(0, 0, 0)
	}
	c, err := TokenToColor(dt.Accent)
	if err == nil {
		return c
	}
	return colorpkg.RGB(0, 0, 0)
}

// ThemeGradient returns a gradient from accent to background with the requested number of steps.
func ThemeGradient(dt *design.DesignTokens, steps int) []colorpkg.Color {
	if steps <= 0 {
		return nil
	}

	accent := ThemeAccentColor(dt)
	var background colorpkg.Color = colorpkg.RGB(0, 0, 0)
	if dt != nil {
		if bg, err := TokenToColor(dt.Background); err == nil {
			background = bg
		}
	}
	if steps == 1 {
		return []colorpkg.Color{accent}
	}

	ar, ag, ab, aa := accent.RGBA()
	br, bg, bb, ba := background.RGBA()

	gradient := make([]colorpkg.Color, 0, steps)
	for i := range steps {
		t := float64(i) / float64(steps-1)
		r := ar + (br-ar)*t
		g := ag + (bg-ag)*t
		b := ab + (bb-ab)*t
		a := aa + (ba-aa)*t
		gradient = append(gradient, colorpkg.NewRGBA(r, g, b, a))
	}
	return gradient
}

func toByte(v float64) int {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return int(math.Round(v * 255))
}

// ShimmerText applies a sweeping highlight effect across text.
// phase is 0.0-1.0, representing the position of the highlight window.
// The highlight slides from left to right across the visible characters.
func ShimmerText(text string, baseHex string, highlightHex string, phase float64) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}

	baseColor, err := TokenToColor(baseHex)
	if err != nil {
		return text
	}
	highlightColor, err := TokenToColor(highlightHex)
	if err != nil {
		return text
	}

	br, bg, bb, _ := baseColor.RGBA()
	hr, hg, hb, _ := highlightColor.RGBA()

	n := len(runes)
	// The highlight center position sweeps across the text
	center := phase * float64(n)
	// Highlight window half-width (in characters)
	halfWidth := 4.0
	if float64(n) < halfWidth*2 {
		halfWidth = float64(n) / 2.0
	}

	var sb strings.Builder
	for i, r := range runes {
		// Distance from the highlight center, normalized to 0-1
		dist := math.Abs(float64(i)-center) / halfWidth
		if dist > 1.0 {
			dist = 1.0
		}
		// Smooth falloff: 1.0 at center, 0.0 at edges
		t := 1.0 - dist
		t = t * t // quadratic falloff for smoother shimmer

		// Interpolate between base and highlight
		cr := br + (hr-br)*t
		cg := bg + (hg-bg)*t
		cb := bb + (hb-bb)*t

		sb.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm", toByte(cr), toByte(cg), toByte(cb)))
		sb.WriteRune(r)
	}
	sb.WriteString(ANSIReset)
	return sb.String()
}