package style

import (
	"testing"
)

func TestDetectColorSchemeReturnsKnownValue(t *testing.T) {
	scheme := DetectColorScheme()
	if scheme != ColorSchemeDark && scheme != ColorSchemeLight {
		t.Fatalf("unexpected color scheme: %q", scheme)
	}
}

func TestQueryTerminalColorsSensibleBehavior(t *testing.T) {
	colors, err := QueryTerminalColors()
	if err != nil {
		// Common in tests/non-interactive shells. Error is acceptable and sensible.
		return
	}

	if colors == nil {
		t.Fatal("expected non-nil TerminalColors when no error is returned")
	}
	if colors.Background == "" {
		t.Fatal("expected detected background color")
	}
	if colors.Foreground == "" {
		t.Fatal("expected detected foreground color")
	}
	for i, value := range colors.Palette {
		if value == "" {
			t.Fatalf("expected palette index %d to be populated", i)
		}
	}
}

func TestParseOSCColorAndHelpers(t *testing.T) {
	r, g, b, ok := parseOSCColor("#112233")
	if !ok || r != 0x11 || g != 0x22 || b != 0x33 {
		t.Fatalf("unexpected parse result for #112233: (%d,%d,%d,%v)", r, g, b, ok)
	}

	r, g, b, ok = parseOSCColor("rgb:ffff/0000/7fff")
	if !ok {
		t.Fatal("expected rgb: form to parse")
	}
	if r != 255 || g != 0 || b < 127 || b > 128 {
		t.Fatalf("unexpected rgb values: (%d,%d,%d)", r, g, b)
	}

	if lum := luminance(255, 255, 255); lum <= 128 {
		t.Fatalf("expected white luminance > 128, got %f", lum)
	}
	if lum := luminance(0, 0, 0); lum != 0 {
		t.Fatalf("expected black luminance 0, got %f", lum)
	}
}
