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
