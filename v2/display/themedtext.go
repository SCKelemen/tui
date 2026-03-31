package display

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// TextVariant represents semantic text color variants.
type TextVariant int

const (
	VariantDefault TextVariant = iota
	VariantPrimary
	VariantSecondary
	VariantSuccess
	VariantError
	VariantWarning
	VariantMuted
)

// ThemedText renders text using semantic themed color variants.
type ThemedText struct {
	text         string
	variant      TextVariant
	bold         bool
	italic       bool
	underline    bool
	focused      bool
	designTokens *design.DesignTokens
}

// ThemedTextOption configures ThemedText.
type ThemedTextOption func(*ThemedText)

// WithThemedTextBold toggles bold style.
func WithThemedTextBold(bold bool) ThemedTextOption {
	return func(t *ThemedText) { t.bold = bold }
}

// WithThemedTextItalic toggles italic style.
func WithThemedTextItalic(italic bool) ThemedTextOption {
	return func(t *ThemedText) { t.italic = italic }
}

// WithThemedTextUnderline toggles underline style.
func WithThemedTextUnderline(underline bool) ThemedTextOption {
	return func(t *ThemedText) { t.underline = underline }
}

// WithThemedTextDesignTokens applies design tokens.
func WithThemedTextDesignTokens(tokens *design.DesignTokens) ThemedTextOption {
	return func(t *ThemedText) {
		if tokens != nil {
			t.designTokens = tokens
		}
	}
}

// NewThemedText creates a ThemedText component.
func NewThemedText(text string, variant TextVariant, opts ...ThemedTextOption) *ThemedText {
	t := &ThemedText{
		text:         text,
		variant:      variant,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Init initializes the component.
func (t *ThemedText) Init() tea.Cmd { return nil }

// Update handles Bubble Tea messages.
func (t *ThemedText) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	return t, nil
}

// View renders the themed text.
func (t *ThemedText) View() string {
	color := t.resolveVariantColor()
	var b strings.Builder
	if color != "" {
		b.WriteString(color)
	}
	if t.bold {
		b.WriteString(style.ANSIBold)
	}
	if t.italic {
		b.WriteString("\033[3m")
	}
	if t.underline {
		b.WriteString(style.ANSIUnderline)
	}
	b.WriteString(t.text)
	b.WriteString(style.ANSIReset)
	return b.String()
}

// Focus marks the component as focused.
func (t *ThemedText) Focus() { t.focused = true }

// Blur marks the component as unfocused.
func (t *ThemedText) Blur() { t.focused = false }

// Focused reports whether the component is focused.
func (t *ThemedText) Focused() bool { return t.focused }

func (t *ThemedText) resolveVariantColor() string {
	tokens := t.designTokens
	switch t.variant {
	case VariantPrimary:
		if tokens != nil {
			if c := style.Fg(tokens.Accent); c != "" {
				return c
			}
		}
		return style.ANSICyan
	case VariantSecondary:
		if tokens != nil {
			if c := style.Fg(tokens.Color); c != "" {
				return c
			}
		}
		return style.ANSIWhite
	case VariantSuccess:
		if tokens != nil {
			if c := style.Fg(tokens.SuccessBright); c != "" {
				return c
			}
		}
		return style.ANSIGreen
	case VariantError:
		if tokens != nil {
			if c := style.Fg(tokens.ErrorBright); c != "" {
				return c
			}
		}
		return style.ANSIRed
	case VariantWarning:
		if tokens != nil {
			if c := style.Fg(tokens.PendingColor); c != "" {
				return c
			}
		}
		return style.ANSIYellow
	case VariantMuted:
		if tokens != nil {
			if c := style.Fg(tokens.MutedColor); c != "" {
				return c
			}
		}
		return style.ANSIDim
	default:
		if tokens != nil {
			if c := style.Fg(tokens.Color); c != "" {
				return c
			}
		}
		return style.ANSIWhite
	}
}

var _ tui.Component = (*ThemedText)(nil)
