package tui

import "testing"

func TestDetectThemeModeDark(t *testing.T) {
	t.Setenv("COLORFGBG", "15;0")
	if got := DetectThemeMode(); got != ThemeModeDark {
		t.Errorf("DetectThemeMode() = %v, want ThemeModeDark", got)
	}
}

func TestDetectThemeModeLight(t *testing.T) {
	t.Setenv("COLORFGBG", "0;15")
	if got := DetectThemeMode(); got != ThemeModeLight {
		t.Errorf("DetectThemeMode() = %v, want ThemeModeLight", got)
	}
}

func TestDetectThemeModeMissingReturnsUnknown(t *testing.T) {
	t.Setenv("COLORFGBG", "")
	if got := DetectThemeMode(); got != ThemeModeUnknown {
		t.Errorf("DetectThemeMode() = %v, want ThemeModeUnknown", got)
	}
}

func TestDetectThemeModeMalformedReturnsUnknown(t *testing.T) {
	cases := []string{
		"abc",
		"7",
		"a;b",
		"7;",
		";0",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			t.Setenv("COLORFGBG", c)
			if got := DetectThemeMode(); got != ThemeModeUnknown {
				t.Errorf("COLORFGBG=%q DetectThemeMode() = %v, want ThemeModeUnknown", c, got)
			}
		})
	}
}

func TestDetectThemeModeAmbiguousMidPaletteReturnsUnknown(t *testing.T) {
	// fg=5, bg=5: neither side passes the dark/light threshold.
	t.Setenv("COLORFGBG", "5;5")
	if got := DetectThemeMode(); got != ThemeModeUnknown {
		t.Errorf("DetectThemeMode() = %v, want ThemeModeUnknown for mid-palette values", got)
	}
}

func TestDetectThemeModeThreeFieldFormat(t *testing.T) {
	// rxvt-style "fg;default;bg"
	t.Setenv("COLORFGBG", "15;default;0")
	if got := DetectThemeMode(); got != ThemeModeDark {
		t.Errorf("DetectThemeMode() = %v, want ThemeModeDark", got)
	}
}

func TestDetectThemeModeCmdEmitsMsg(t *testing.T) {
	t.Setenv("COLORFGBG", "0;15")
	cmd := DetectThemeModeCmd()
	if cmd == nil {
		t.Fatal("DetectThemeModeCmd returned nil")
	}
	msg, ok := cmd().(ThemeModeMsg)
	if !ok {
		t.Fatalf("expected ThemeModeMsg, got %T", cmd())
	}
	if msg.Mode != ThemeModeLight {
		t.Errorf("expected ThemeModeLight, got %v", msg.Mode)
	}
}

func TestThemeModeStringStable(t *testing.T) {
	cases := map[ThemeMode]string{
		ThemeModeUnknown: "unknown",
		ThemeModeDark:    "dark",
		ThemeModeLight:   "light",
	}
	for m, want := range cases {
		if got := m.String(); got != want {
			t.Errorf("(%d).String() = %q, want %q", m, got, want)
		}
	}
}
