package style

import (
	"strings"
	"testing"

	design "github.com/SCKelemen/design-system"
)

func TestDesignTokensForTheme(t *testing.T) {
	tests := []struct {
		name      string
		theme     string
		expected  string
		assertMsg string
	}{
		{name: "midnight", theme: "midnight", expected: "midnight", assertMsg: "midnight should resolve to midnight theme"},
		{name: "nord", theme: "nord", expected: "nord", assertMsg: "nord should resolve to nord theme"},
		{name: "paper", theme: "paper", expected: "paper", assertMsg: "paper should resolve to paper theme"},
		{name: "wrapped", theme: "wrapped", expected: "wrapped", assertMsg: "wrapped should resolve to wrapped theme"},
		{name: "case insensitive", theme: "NoRd", expected: "nord", assertMsg: "mixed case should resolve to nord theme"},
		{name: "trim spaces", theme: "  midnight  ", expected: "midnight", assertMsg: "surrounding whitespace should be ignored"},
		{name: "default", theme: "default", expected: "default", assertMsg: "default should resolve to default theme"},
		{name: "empty", theme: "", expected: "default", assertMsg: "empty should resolve to default theme"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tokens := DesignTokensForTheme(tc.theme)
			if tokens == nil {
				t.Fatalf("%s: got nil tokens", tc.assertMsg)
			}
			if tokens.Theme != tc.expected {
				t.Fatalf("%s: got %q, want %q", tc.assertMsg, tokens.Theme, tc.expected)
			}
		})
	}
}

func TestDesignTokensForThemeReturnsDefaultForUnknown(t *testing.T) {
	unknown := DesignTokensForTheme("does-not-exist")
	if unknown == nil {
		t.Fatal("expected non-nil tokens for unknown theme")
	}

	def := design.DefaultTheme()
	if unknown.Theme != def.Theme {
		t.Fatalf("theme fallback mismatch: got %q, want %q", unknown.Theme, def.Theme)
	}
	if unknown.Background != def.Background {
		t.Fatalf("background fallback mismatch: got %q, want %q", unknown.Background, def.Background)
	}
	if unknown.Accent != def.Accent {
		t.Fatalf("accent fallback mismatch: got %q, want %q", unknown.Accent, def.Accent)
	}
}

func TestAnsiColorFromHex(t *testing.T) {
	tests := []struct {
		name     string
		hex      string
		expected string
	}{
		{name: "orange", hex: "#FF5733", expected: "\033[38;2;255;87;51m"},
		{name: "black", hex: "#000000", expected: "\033[38;2;0;0;0m"},
		{name: "white", hex: "#FFFFFF", expected: "\033[38;2;255;255;255m"},
		{name: "trim spaces", hex: "  #00ff00  ", expected: "\033[38;2;0;255;0m"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ANSIColorFromHex(tc.hex)
			if got != tc.expected {
				t.Fatalf("ANSIColorFromHex(%q) = %q, want %q", tc.hex, got, tc.expected)
			}
		})
	}
}

func TestAnsiColorFromHexInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		hex  string
	}{
		{name: "empty", hex: ""},
		{name: "no hash", hex: "FF5733"},
		{name: "too short", hex: "#FFF"},
		{name: "non-hex", hex: "#GGGGGG"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ANSIColorFromHex(tc.hex)
			if got != "" {
				t.Fatalf("ANSIColorFromHex(%q) = %q, want empty string", tc.hex, got)
			}
		})
	}
}

func TestANSIBackgroundColorFromHex(t *testing.T) {
	tests := []struct {
		name     string
		hex      string
		expected string
	}{
		{name: "orange", hex: "#FF5733", expected: "\033[48;2;255;87;51m"},
		{name: "black", hex: "#000000", expected: "\033[48;2;0;0;0m"},
		{name: "white", hex: "#FFFFFF", expected: "\033[48;2;255;255;255m"},
		{name: "trim spaces", hex: "  #00ff00  ", expected: "\033[48;2;0;255;0m"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ANSIBackgroundColorFromHex(tc.hex)
			if got != tc.expected {
				t.Fatalf("ANSIBackgroundColorFromHex(%q) = %q, want %q", tc.hex, got, tc.expected)
			}
		})
	}
}

func TestANSIBackgroundColorFromHexInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		hex  string
	}{
		{name: "empty", hex: ""},
		{name: "no hash", hex: "FF5733"},
		{name: "too short", hex: "#FFF"},
		{name: "non-hex", hex: "#GGGGGG"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ANSIBackgroundColorFromHex(tc.hex)
			if got != "" {
				t.Fatalf("ANSIBackgroundColorFromHex(%q) = %q, want empty string", tc.hex, got)
			}
		})
	}
}
func TestANSIConstants(t *testing.T) {
	constants := map[string]string{
		"ANSIReset":     ANSIReset,
		"ANSIBold":      ANSIBold,
		"ANSIDim":       ANSIDim,
		"ANSIUnderline": ANSIUnderline,
		"ANSIInverse":   ANSIInverse,
		"ANSIRed":       ANSIRed,
		"ANSIGreen":     ANSIGreen,
		"ANSIYellow":    ANSIYellow,
		"ANSIBlue":      ANSIBlue,
		"ANSICyan":      ANSICyan,
		"ANSIWhite":     ANSIWhite,
	}

	for name, value := range constants {
		if !strings.HasPrefix(value, "\033[") {
			t.Fatalf("%s = %q, expected prefix %q", name, value, "\\033[")
		}
	}
}
