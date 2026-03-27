package style

import (
	"testing"

	colorpkg "github.com/SCKelemen/color"
	design "github.com/SCKelemen/design-system"
)

func TestColorTokenToColor(t *testing.T) {
	c, err := TokenToColor("#FF5733")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil color")
	}

	r, g, b, _ := c.RGBA()
	if toByte(r) != 255 || toByte(g) != 87 || toByte(b) != 51 {
		t.Fatalf("unexpected RGB values: got %d,%d,%d", toByte(r), toByte(g), toByte(b))
	}
}

func TestColorToANSI(t *testing.T) {
	c := colorpkg.RGB(1.0, 0.0, 0.0)

	fg := ColorToANSIFg(c)
	if fg != "\033[38;2;255;0;0m" {
		t.Fatalf("unexpected fg sequence: %q", fg)
	}

	bg := ColorToANSIBg(c)
	if bg != "\033[48;2;255;0;0m" {
		t.Fatalf("unexpected bg sequence: %q", bg)
	}
}

func TestColorThemeAccentColor(t *testing.T) {
	dt := &design.DesignTokens{Accent: "#336699"}
	c := ThemeAccentColor(dt)
	if c == nil {
		t.Fatal("expected non-nil accent color")
	}

	r, g, b, _ := c.RGBA()
	if toByte(r) != 51 || toByte(g) != 102 || toByte(b) != 153 {
		t.Fatalf("unexpected accent RGB values: got %d,%d,%d", toByte(r), toByte(g), toByte(b))
	}
}

func TestColorThemeGradient(t *testing.T) {
	dt := &design.DesignTokens{Accent: "#FF0000", Background: "#0000FF"}
	gradient := ThemeGradient(dt, 3)
	if len(gradient) != 3 {
		t.Fatalf("expected 3 gradient colors, got %d", len(gradient))
	}

	r0, g0, b0, _ := gradient[0].RGBA()
	if toByte(r0) != 255 || toByte(g0) != 0 || toByte(b0) != 0 {
		t.Fatalf("unexpected first gradient color: %d,%d,%d", toByte(r0), toByte(g0), toByte(b0))
	}

	r2, g2, b2, _ := gradient[2].RGBA()
	if toByte(r2) != 0 || toByte(g2) != 0 || toByte(b2) != 255 {
		t.Fatalf("unexpected last gradient color: %d,%d,%d", toByte(r2), toByte(g2), toByte(b2))
	}
}

func TestTokenToColorErrorPaths(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{name: "empty token", token: ""},
		{name: "whitespace token", token: "   "},
		{name: "invalid hex with hash", token: "#GGGGGG"},
		{name: "invalid hex without hash", token: "GGGGGG"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := TokenToColor(tc.token)
			if err == nil {
				t.Fatalf("TokenToColor(%q) expected error, got nil", tc.token)
			}
		})
	}
}

func TestTokenToColorTrimsAndAddsHash(t *testing.T) {
	c, err := TokenToColor("  ff00aa  ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	r, g, b, _ := c.RGBA()
	if toByte(r) != 255 || toByte(g) != 0 || toByte(b) != 170 {
		t.Fatalf("unexpected RGB values: got %d,%d,%d", toByte(r), toByte(g), toByte(b))
	}
}

func TestColorToANSIFormatAndNil(t *testing.T) {
	if got := ColorToANSIFg(nil); got != "" {
		t.Fatalf("ColorToANSIFg(nil) = %q, want empty string", got)
	}
	if got := ColorToANSIBg(nil); got != "" {
		t.Fatalf("ColorToANSIBg(nil) = %q, want empty string", got)
	}

	c := colorpkg.RGB(0.5, 0.25, 0.75)
	if got := ColorToANSIFg(c); got != "\033[38;2;128;64;191m" {
		t.Fatalf("ColorToANSIFg format mismatch: got %q", got)
	}
	if got := ColorToANSIBg(c); got != "\033[48;2;128;64;191m" {
		t.Fatalf("ColorToANSIBg format mismatch: got %q", got)
	}
}

func TestThemeGradientStepCounts(t *testing.T) {
	dt := &design.DesignTokens{Accent: "#FF0000", Background: "#0000FF"}

	if g := ThemeGradient(dt, 0); g != nil {
		t.Fatalf("expected nil gradient for 0 steps, got len=%d", len(g))
	}

	g1 := ThemeGradient(dt, 1)
	if len(g1) != 1 {
		t.Fatalf("expected 1 gradient color, got %d", len(g1))
	}
	r, g, b, _ := g1[0].RGBA()
	if toByte(r) != 255 || toByte(g) != 0 || toByte(b) != 0 {
		t.Fatalf("unexpected single-step gradient color: %d,%d,%d", toByte(r), toByte(g), toByte(b))
	}

	large := ThemeGradient(dt, 64)
	if len(large) != 64 {
		t.Fatalf("expected 64 gradient colors, got %d", len(large))
	}
	r0, g0, b0, _ := large[0].RGBA()
	rN, gN, bN, _ := large[len(large)-1].RGBA()
	if toByte(r0) != 255 || toByte(g0) != 0 || toByte(b0) != 0 {
		t.Fatalf("unexpected first large gradient color: %d,%d,%d", toByte(r0), toByte(g0), toByte(b0))
	}
	if toByte(rN) != 0 || toByte(gN) != 0 || toByte(bN) != 255 {
		t.Fatalf("unexpected last large gradient color: %d,%d,%d", toByte(rN), toByte(gN), toByte(bN))
	}
}

func TestThemeGradientFallbackPaths(t *testing.T) {
	g := ThemeGradient(nil, 2)
	if len(g) != 2 {
		t.Fatalf("expected 2 gradient colors, got %d", len(g))
	}
	r0, g0, b0, _ := g[0].RGBA()
	r1, g1, b1, _ := g[1].RGBA()
	if toByte(r0) != 0 || toByte(g0) != 0 || toByte(b0) != 0 {
		t.Fatalf("unexpected nil-theme start color: %d,%d,%d", toByte(r0), toByte(g0), toByte(b0))
	}
	if toByte(r1) != 0 || toByte(g1) != 0 || toByte(b1) != 0 {
		t.Fatalf("unexpected nil-theme end color: %d,%d,%d", toByte(r1), toByte(g1), toByte(b1))
	}

	dt := &design.DesignTokens{Accent: "#GGGGGG", Background: "#0000FF"}
	g = ThemeGradient(dt, 2)
	rA, gA, bA, _ := g[0].RGBA()
	rB, gB, bB, _ := g[1].RGBA()
	if toByte(rA) != 0 || toByte(gA) != 0 || toByte(bA) != 0 {
		t.Fatalf("unexpected fallback accent color: %d,%d,%d", toByte(rA), toByte(gA), toByte(bA))
	}
	if toByte(rB) != 0 || toByte(gB) != 0 || toByte(bB) != 255 {
		t.Fatalf("unexpected background color: %d,%d,%d", toByte(rB), toByte(gB), toByte(bB))
	}
}