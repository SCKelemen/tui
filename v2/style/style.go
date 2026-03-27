package style

import (
	"fmt"
	"strconv"
	"strings"

	design "github.com/SCKelemen/design-system"
)

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
