package style

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	design "github.com/SCKelemen/design-system"
)

// ColorMode represents the terminal's color capability.
type ColorMode int

const (
	// ColorBasic is 16-color ANSI (fallback for limited terminals).
	ColorBasic ColorMode = iota
	// Color256 is 256-color xterm.
	Color256
	// ColorTrueColor is 24-bit RGB (Ghostty, Kitty, iTerm2, WezTerm, etc).
	ColorTrueColor
)

// DetectColorMode checks the terminal's color capability.
// It inspects COLORTERM, TERM, and known terminal emulators.
func DetectColorMode() ColorMode {
	// COLORTERM=truecolor or 24bit is the standard signal
	ct := os.Getenv("COLORTERM")
	if ct == "truecolor" || ct == "24bit" {
		return ColorTrueColor
	}
	// Known true-color terminals by TERM_PROGRAM
	tp := os.Getenv("TERM_PROGRAM")
	switch strings.ToLower(tp) {
	case "ghostty", "kitty", "iterm.app", "wezterm", "alacritty", "warp":
		return ColorTrueColor
	}
	// xterm-256color and similar
	term := os.Getenv("TERM")
	if strings.Contains(term, "256color") {
		return Color256
	}
	// Default to truecolor on modern systems (most terminals support it now)
	if term != "" && term != "dumb" {
		return ColorTrueColor
	}
	return ColorBasic
}

// colorMode is detected once at init time.
var colorMode = DetectColorMode()

// GetColorMode returns the current detected color mode.
func GetColorMode() ColorMode { return colorMode }

// SetColorMode overrides the detected color mode (useful for testing).
func SetColorMode(m ColorMode) { colorMode = m }

// Fg returns a foreground color escape sequence.
// On truecolor terminals, uses 24-bit RGB from the hex value.
// Falls back to the closest basic ANSI color on limited terminals.
func Fg(hex string) string {
	if colorMode == ColorTrueColor || colorMode == Color256 {
		return ANSIColorFromHex(hex)
	}
	return hexToBasicFg(hex)
}

// Bg returns a background color escape sequence.
// On truecolor terminals, uses 24-bit RGB from the hex value.
// Falls back to the closest basic ANSI background on limited terminals.
func Bg(hex string) string {
	if colorMode == ColorTrueColor || colorMode == Color256 {
		return ANSIBackgroundColorFromHex(hex)
	}
	return hexToBasicBg(hex)
}

// hexToBasicFg maps a hex color to the nearest basic 16-color ANSI foreground.
func hexToBasicFg(hex string) string {
	s := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(s) != 6 {
		return ""
	}
	value, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return ""
	}
	r, g, b := int((value>>16)&0xFF), int((value>>8)&0xFF), int(value&0xFF)
	return nearestANSI(r, g, b, false)
}

// hexToBasicBg maps a hex color to the nearest basic 16-color ANSI background.
func hexToBasicBg(hex string) string {
	s := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(s) != 6 {
		return ""
	}
	value, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return ""
	}
	r, g, b := int((value>>16)&0xFF), int((value>>8)&0xFF), int(value&0xFF)
	return nearestANSI(r, g, b, true)
}

// nearestANSI maps RGB to the closest basic ANSI color.
func nearestANSI(r, g, b int, bg bool) string {
	type ansiColor struct {
		r, g, b int
		fg, bg  string
	}
	palette := []ansiColor{
		{0, 0, 0, "\033[30m", "\033[40m"},         // black
		{170, 0, 0, "\033[31m", "\033[41m"},       // red
		{0, 170, 0, "\033[32m", "\033[42m"},       // green
		{170, 170, 0, "\033[33m", "\033[43m"},     // yellow
		{0, 0, 170, "\033[34m", "\033[44m"},       // blue
		{170, 0, 170, "\033[35m", "\033[45m"},     // magenta
		{0, 170, 170, "\033[36m", "\033[46m"},     // cyan
		{170, 170, 170, "\033[37m", "\033[47m"},   // white
		{85, 85, 85, "\033[90m", "\033[100m"},     // bright black
		{255, 85, 85, "\033[91m", "\033[101m"},    // bright red
		{85, 255, 85, "\033[92m", "\033[102m"},    // bright green
		{255, 255, 85, "\033[93m", "\033[103m"},   // bright yellow
		{85, 85, 255, "\033[94m", "\033[104m"},    // bright blue
		{255, 85, 255, "\033[95m", "\033[105m"},   // bright magenta
		{85, 255, 255, "\033[96m", "\033[106m"},   // bright cyan
		{255, 255, 255, "\033[97m", "\033[107m"},  // bright white
	}
	bestDist := 1 << 30
	bestIdx := 0
	for i, c := range palette {
		dr, dg, db := r-c.r, g-c.g, b-c.b
		dist := dr*dr + dg*dg + db*db
		if dist < bestDist {
			bestDist = dist
			bestIdx = i
		}
	}
	if bg {
		return palette[bestIdx].bg
	}
	return palette[bestIdx].fg
}

const (
	ANSIReset     = "\033[0m"
	ANSIBold      = "\033[1m"
	ANSIDim       = "\033[2m"
	ANSIUnderline = "\033[4m"
	ANSIInverse   = "\033[7m"

	ANSIRed    = "\033[31m"
	ANSIGreen  = "\033[32m"
	ANSIYellow = "\033[33m"
	ANSIBlue   = "\033[34m"
	ANSICyan   = "\033[36m"
	ANSIWhite  = "\033[37m"
)

// DesignTokensForTheme resolves design tokens by theme name.
func DesignTokensForTheme(theme string) *design.DesignTokens {
	switch strings.ToLower(strings.TrimSpace(theme)) {
	case "midnight":
		return design.MidnightTheme()
	case "nord":
		return design.NordTheme()
	case "paper":
		return design.PaperTheme()
	case "wrapped":
		return design.WrappedTheme()
	default:
		return design.DefaultTheme()
	}
}

// ANSIColorFromHex converts a hex color to a 24-bit ANSI foreground color sequence.
func ANSIColorFromHex(hex string) string {
	token := strings.TrimSpace(hex)
	if !strings.HasPrefix(token, "#") {
		return ""
	}

	s := strings.TrimPrefix(token, "#")
	if len(s) != 6 {
		return ""
	}

	value, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return ""
	}

	r := (value >> 16) & 0xFF
	g := (value >> 8) & 0xFF
	b := value & 0xFF
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

// ANSIBackgroundColorFromHex converts a hex color to a 24-bit ANSI background color sequence.
func ANSIBackgroundColorFromHex(hex string) string {
	token := strings.TrimSpace(hex)
	if !strings.HasPrefix(token, "#") {
		return ""
	}

	s := strings.TrimPrefix(token, "#")
	if len(s) != 6 {
		return ""
	}

	value, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return ""
	}

	r := (value >> 16) & 0xFF
	g := (value >> 8) & 0xFF
	b := value & 0xFF
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
}
