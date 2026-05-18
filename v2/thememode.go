package tui

import (
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ThemeMode reports whether the host terminal is presenting a dark or
// light color scheme. Applications can use ThemeMode to flip between
// matching color palettes without asking the user.
type ThemeMode int

const (
	// ThemeModeUnknown indicates detection failed; the application
	// should fall back to a default theme of its choice.
	ThemeModeUnknown ThemeMode = iota
	// ThemeModeDark indicates a dark background.
	ThemeModeDark
	// ThemeModeLight indicates a light background.
	ThemeModeLight
)

// String returns a stable, lowercase tag suitable for logging.
func (m ThemeMode) String() string {
	switch m {
	case ThemeModeDark:
		return "dark"
	case ThemeModeLight:
		return "light"
	default:
		return "unknown"
	}
}

// ThemeModeMsg delivers an asynchronously-detected ThemeMode to the
// program's Update method. Apps that want to react to theme changes
// should match on this message type.
type ThemeModeMsg struct {
	Mode ThemeMode
}

// DetectThemeMode inspects the COLORFGBG environment variable and
// classifies the host terminal as dark, light, or unknown.
//
// COLORFGBG has the shape "<fg>;<bg>" (and occasionally "<fg>;<sep>;<bg>"
// in some emulators) where each value is a decimal ANSI palette index
// in the 0-15 range. Modern terminals export it on launch when they
// believe their palette has a stable foreground/background pair. The
// classification mirrors xterm's convention: high foreground + low
// background means dark mode; low foreground + high background means
// light mode; any other combination, or a missing/malformed value,
// returns ThemeModeUnknown.
//
// The DEC 2031 query path (writing `\x1b]11;?\x07` and reading the
// response) is intentionally not implemented: bubbletea owns stdin
// and stdout while the program is running, so issuing the query from
// inside this package would deadlock with the runtime. Callers that
// need a synchronous probe can issue the OSC sequence themselves
// before starting their bubbletea program.
func DetectThemeMode() ThemeMode {
	raw := strings.TrimSpace(os.Getenv("COLORFGBG"))
	if raw == "" {
		return ThemeModeUnknown
	}

	parts := strings.Split(raw, ";")
	if len(parts) < 2 {
		return ThemeModeUnknown
	}

	// Some terminals (rxvt) emit "fg;default;bg"; take the first and
	// last fields so both shapes parse correctly.
	fg, errFg := strconv.Atoi(strings.TrimSpace(parts[0]))
	bg, errBg := strconv.Atoi(strings.TrimSpace(parts[len(parts)-1]))
	if errFg != nil || errBg != nil {
		return ThemeModeUnknown
	}

	switch {
	case fg >= 7 && bg <= 6:
		return ThemeModeDark
	case fg <= 6 && bg >= 7:
		return ThemeModeLight
	default:
		return ThemeModeUnknown
	}
}

// DetectThemeModeCmd returns a tea.Cmd that emits a ThemeModeMsg with
// the result of DetectThemeMode. The command is opt-in: Application.Init
// does not invoke it automatically, so existing apps see no behaviour
// change. Apps that care about theme mode can return DetectThemeModeCmd()
// from their own Init.
func DetectThemeModeCmd() tea.Cmd {
	return func() tea.Msg {
		return ThemeModeMsg{Mode: DetectThemeMode()}
	}
}
