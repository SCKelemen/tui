package tui

import (
	"fmt"
	"strconv"
	"strings"

	design "github.com/SCKelemen/design-system"
)

func designTokensForTheme(theme string) *design.DesignTokens {
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

func ansiColorFromHex(hex string) string {
	s := strings.TrimSpace(strings.TrimPrefix(hex, "#"))
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
