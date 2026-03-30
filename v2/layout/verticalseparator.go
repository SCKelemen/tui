package layout

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// VerticalSeparator renders a vertical line for side-by-side layouts.
type VerticalSeparator struct {
	height        int
	color         string
	char          string
	focused       bool
	designTokens  *design.DesignTokens
	colorExplicit bool
}

// VerticalSeparatorOption configures a VerticalSeparator.
type VerticalSeparatorOption func(*VerticalSeparator)

// WithVerticalSeparatorColor sets the separator line color using a hex color value.
func WithVerticalSeparatorColor(hex string) VerticalSeparatorOption {
	return func(v *VerticalSeparator) {
		v.color = strings.TrimSpace(hex)
		v.colorExplicit = true
	}
}

// WithVerticalSeparatorChar sets the separator rune/string used for each line.
func WithVerticalSeparatorChar(ch string) VerticalSeparatorOption {
	return func(v *VerticalSeparator) {
		if strings.TrimSpace(ch) != "" {
			v.char = ch
		}
	}
}

// WithVerticalSeparatorDesignTokens applies design tokens to the separator.
func WithVerticalSeparatorDesignTokens(tokens *design.DesignTokens) VerticalSeparatorOption {
	return func(v *VerticalSeparator) {
		v.applyDesignTokens(tokens)
	}
}

// NewVerticalSeparator creates a new VerticalSeparator component.
func NewVerticalSeparator(height int, opts ...VerticalSeparatorOption) *VerticalSeparator {
	v := &VerticalSeparator{
		height:       height,
		char:         "│",
		designTokens: design.DefaultTheme(),
	}

	if v.height < 0 {
		v.height = 0
	}

	v.applyDesignTokens(v.designTokens)

	for _, opt := range opts {
		opt(v)
	}

	return v
}

// RenderVerticalSeparator renders a one-off vertical separator with an optional color.
func RenderVerticalSeparator(height int, color string) string {
	return NewVerticalSeparator(height, WithVerticalSeparatorColor(color)).View()
}

// Init initializes the component.
func (v *VerticalSeparator) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (v *VerticalSeparator) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	return v, nil
}

// View renders the separator.
func (v *VerticalSeparator) View() string {
	if v.height <= 0 {
		return ""
	}

	char := v.char
	if strings.TrimSpace(char) == "" {
		char = "│"
	}

	fg := style.ANSIColorFromHex(v.color)

	var b strings.Builder
	for i := 0; i < v.height; i++ {
		line := char
		if fg != "" {
			line = fg + line + style.ANSIReset
		}
		b.WriteString(line)
		if i < v.height-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// Focus marks the component as focused.
func (v *VerticalSeparator) Focus() {
	v.focused = true
}

// Blur marks the component as unfocused.
func (v *VerticalSeparator) Blur() {
	v.focused = false
}

// Focused returns whether this component is focused.
func (v *VerticalSeparator) Focused() bool {
	return v.focused
}

func (v *VerticalSeparator) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}

	v.designTokens = tokens

	if v.colorExplicit {
		return
	}

	if value := strings.TrimSpace(tokens.BorderSubtle); value != "" {
		v.color = value
		return
	}

	if value := strings.TrimSpace(tokens.Accent); value != "" {
		v.color = value
		return
	}

	if value := strings.TrimSpace(tokens.Color); value != "" {
		v.color = value
	}
}

var _ tui.Component = (*VerticalSeparator)(nil)
