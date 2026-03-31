package display

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// TokenWarningOption configures a TokenWarning.
type TokenWarningOption func(*TokenWarning)

// WithTokenWarningWidth sets preferred render width.
func WithTokenWarningWidth(width int) TokenWarningOption {
	return func(t *TokenWarning) {
		if width > 0 {
			t.width = width
		}
	}
}

// WithTokenWarningLabel sets the warning label prefix.
func WithTokenWarningLabel(label string) TokenWarningOption {
	return func(t *TokenWarning) {
		if strings.TrimSpace(label) != "" {
			t.label = strings.TrimSpace(label)
		}
	}
}

// WithTokenWarningDesignTokens applies design-system tokens.
func WithTokenWarningDesignTokens(tokens *design.DesignTokens) TokenWarningOption {
	return func(t *TokenWarning) {
		if tokens != nil {
			t.designTokens = tokens
		}
	}
}

// TokenWarning renders token/context utilization warnings.
type TokenWarning struct {
	current      int
	max          int
	width        int
	windowWidth  int
	label        string
	focused      bool
	designTokens *design.DesignTokens
}

// NewTokenWarning creates a TokenWarning component.
func NewTokenWarning(current, max int, opts ...TokenWarningOption) *TokenWarning {
	t := &TokenWarning{
		current:      maxInt(current, 0),
		max:          maxInt(max, 0),
		width:        0,
		label:        "Token usage",
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Init initializes the component.
func (t *TokenWarning) Init() tea.Cmd { return nil }

// Update handles resize messages.
func (t *TokenWarning) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.windowWidth = msg.Width
		return t, nil
	}
	return t, nil
}

// View renders one warning line with threshold-based coloring.
func (t *TokenWarning) View() string {
	percent := 0
	if t.max > 0 {
		percent = int(float64(t.current) / float64(t.max) * 100)
		if percent > 999 {
			percent = 999
		}
	}

	text := fmt.Sprintf("⚠ %s: %s / %s (%d%%)", t.label, formatIntCommas(t.current), formatIntCommas(t.max), percent)
	colored := t.colorForPercent(percent) + text + style.ANSIReset

	width := t.effectiveWidth()
	if width > 0 {
		return style.Pad(style.Truncate(colored, width, "…"), width)
	}
	return colored
}

// Focus marks the component focused.
func (t *TokenWarning) Focus() { t.focused = true }

// Blur marks the component unfocused.
func (t *TokenWarning) Blur() { t.focused = false }

// Focused reports whether this component is focused.
func (t *TokenWarning) Focused() bool { return t.focused }

func (t *TokenWarning) effectiveWidth() int {
	if t.width > 0 {
		return t.width
	}
	return t.windowWidth
}

func (t *TokenWarning) colorForPercent(percent int) string {
	if percent > 90 {
		if t.designTokens != nil {
			if c := style.Fg(t.designTokens.ErrorBright); c != "" {
				return c
			}
		}
		return style.ANSIRed
	}
	if percent > 70 {
		if t.designTokens != nil {
			if c := style.Fg(t.designTokens.PendingColor); c != "" {
				return c
			}
		}
		return style.ANSIYellow
	}
	if t.designTokens != nil {
		if c := style.Fg(t.designTokens.MutedColor); c != "" {
			return c
		}
	}
	return style.ANSIDim
}

func formatIntCommas(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	negative := strings.HasPrefix(s, "-")
	if negative {
		s = s[1:]
	}
	parts := make([]string, 0)
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	if s != "" {
		parts = append([]string{s}, parts...)
	}
	out := strings.Join(parts, ",")
	if negative {
		return "-" + out
	}
	return out
}

var _ tui.Component = (*TokenWarning)(nil)
