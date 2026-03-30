package chart

import (
	"fmt"
	"math"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// UsageBar renders quota/usage as a segmented horizontal bar.
type UsageBar struct {
	used  float64
	total float64
	width int

	baseColor      string
	remainingColor string
	showLabels     bool
	label          string

	focused      bool
	designTokens *design.DesignTokens
}

// UsageBarOption configures a UsageBar.
type UsageBarOption func(*UsageBar)

// WithUsageBarWidth sets the bar width in characters.
func WithUsageBarWidth(width int) UsageBarOption {
	return func(ub *UsageBar) {
		if width > 0 {
			ub.width = width
		}
	}
}

// WithUsageBarBaseColor sets the used segment color (hex).
func WithUsageBarBaseColor(hex string) UsageBarOption {
	return func(ub *UsageBar) {
		if strings.TrimSpace(hex) != "" {
			ub.baseColor = strings.TrimSpace(hex)
		}
	}
}

// WithUsageBarRemainingColor sets the remaining segment color (hex).
func WithUsageBarRemainingColor(hex string) UsageBarOption {
	return func(ub *UsageBar) {
		if strings.TrimSpace(hex) != "" {
			ub.remainingColor = strings.TrimSpace(hex)
		}
	}
}

// WithUsageBarShowLabels toggles label rendering.
func WithUsageBarShowLabels(show bool) UsageBarOption {
	return func(ub *UsageBar) {
		ub.showLabels = show
	}
}

// WithUsageBarLabel sets an optional label (for example: "Subscription").
func WithUsageBarLabel(label string) UsageBarOption {
	return func(ub *UsageBar) {
		ub.label = strings.TrimSpace(label)
	}
}

// WithUsageBarDesignTokens applies design-system tokens.
func WithUsageBarDesignTokens(tokens *design.DesignTokens) UsageBarOption {
	return func(ub *UsageBar) {
		if tokens == nil {
			return
		}
		ub.designTokens = tokens
		if v := strings.TrimSpace(tokens.Accent); v != "" {
			ub.baseColor = v
		}
		if v := strings.TrimSpace(tokens.SuccessBright); v != "" {
			ub.remainingColor = v
		}
	}
}

// NewUsageBar creates a new usage bar from used and total values.
func NewUsageBar(used, total float64, opts ...UsageBarOption) *UsageBar {
	ub := &UsageBar{
		used:           used,
		total:          total,
		width:          24,
		baseColor:      "#D19A66",
		remainingColor: "#98C379",
		showLabels:     true,
		designTokens:   design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(ub)
	}

	return ub
}

// Init initializes the usage bar.
func (ub *UsageBar) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (ub *UsageBar) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	return ub, nil
}

// View renders the usage bar.
func (ub *UsageBar) View() string {
	used := ub.used
	total := ub.total
	if total < 0 {
		total = 0
	}
	if used < 0 {
		used = 0
	}
	if total > 0 && used > total {
		used = total
	}

	progress := 0.0
	if total > 0 {
		progress = used / total
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	filled := int(math.Round(progress * float64(ub.width)))
	if filled < 0 {
		filled = 0
	}
	if filled > ub.width {
		filled = ub.width
	}
	empty := ub.width - filled

	usedColor := style.ANSIColorFromHex(ub.baseColor)
	if usedColor == "" {
		usedColor = style.ANSIYellow
	}

	remainingColor := style.ANSIColorFromHex(ub.remainingColor)
	if remainingColor == "" {
		remainingColor = style.ANSIDim
	}

	bar := ""
	if filled > 0 {
		bar += usedColor + strings.Repeat("█", filled) + style.ANSIReset
	}
	if empty > 0 {
		bar += remainingColor + strings.Repeat("░", empty) + style.ANSIReset
	}

	remainingValue := total - used
	if remainingValue < 0 {
		remainingValue = 0
	}

	var lines []string
	var top strings.Builder
	if ub.label != "" {
		top.WriteString(ub.label)
		top.WriteString(" ")
	}

	if ub.showLabels {
		top.WriteString("Used ")
		top.WriteString(bar)
		top.WriteString(" Remaining")
		top.WriteString("  ")
		top.WriteString("(Available: ")
		top.WriteString(formatUsageValue(remainingValue))
		top.WriteString(")")
	} else {
		top.WriteString(bar)
	}

	lines = append(lines, top.String())

	if ub.showLabels {
		stats := fmt.Sprintf(
			"Used: %s  •  Remaining: %s  •  Total: %s",
			formatUsageValue(used),
			formatUsageValue(remainingValue),
			formatUsageValue(total),
		)
		lines = append(lines, style.ANSIDim+stats+style.ANSIReset)
	}

	return strings.Join(lines, "\n")
}

// Focus marks the component as focused.
func (ub *UsageBar) Focus() {
	ub.focused = true
}

// Blur marks the component as unfocused.
func (ub *UsageBar) Blur() {
	ub.focused = false
}

// Focused returns whether the component is focused.
func (ub *UsageBar) Focused() bool {
	return ub.focused
}

func formatUsageValue(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" || s == "-0" {
		return "0"
	}
	return s
}

var _ tui.Component = (*UsageBar)(nil)
