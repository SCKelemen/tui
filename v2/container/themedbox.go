package container

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ThemedBox renders themed content with optional title and border.
type ThemedBox struct {
	content      string
	width        int
	title        string
	border       bool
	padding      int
	focused      bool
	designTokens *design.DesignTokens
}

// ThemedBoxOption configures a ThemedBox.
type ThemedBoxOption func(*ThemedBox)

// WithThemedBoxWidth sets the render width.
func WithThemedBoxWidth(width int) ThemedBoxOption {
	return func(t *ThemedBox) {
		if width >= 0 {
			t.width = width
		}
	}
}

// WithThemedBoxTitle sets the optional title.
func WithThemedBoxTitle(title string) ThemedBoxOption {
	return func(t *ThemedBox) {
		t.title = strings.TrimSpace(title)
	}
}

// WithThemedBoxBorder toggles border rendering.
func WithThemedBoxBorder(border bool) ThemedBoxOption {
	return func(t *ThemedBox) {
		t.border = border
	}
}

// WithThemedBoxPadding sets internal padding.
func WithThemedBoxPadding(padding int) ThemedBoxOption {
	return func(t *ThemedBox) {
		if padding >= 0 {
			t.padding = padding
		}
	}
}

// WithThemedBoxDesignTokens applies design tokens.
func WithThemedBoxDesignTokens(tokens *design.DesignTokens) ThemedBoxOption {
	return func(t *ThemedBox) {
		if tokens != nil {
			t.designTokens = tokens
		}
	}
}

// NewThemedBox creates a new ThemedBox component.
func NewThemedBox(content string, opts ...ThemedBoxOption) *ThemedBox {
	t := &ThemedBox{
		content:      content,
		width:        60,
		border:       true,
		padding:      1,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Init initializes the component.
func (t *ThemedBox) Init() tea.Cmd { return nil }

// Update handles Bubble Tea messages.
func (t *ThemedBox) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		if t.width == 0 {
			t.width = m.Width
		}
	}
	return t, nil
}

// View renders the themed box.
func (t *ThemedBox) View() string {
	width := t.width
	if width <= 0 {
		width = 1
	}

	content := t.content
	if t.padding > 0 {
		pad := strings.Repeat(" ", t.padding)
		lines := strings.Split(content, "\n")
		for i := range lines {
			lines[i] = pad + lines[i] + pad
		}
		content = strings.Join(lines, "\n")
	}

	if !t.border {
		if t.title == "" {
			return content
		}
		titleColor := style.ANSIBold
		if t.designTokens != nil {
			if c := style.Fg(t.designTokens.Accent); c != "" {
				titleColor = c + style.ANSIBold
			}
		}
		return titleColor + t.title + style.ANSIReset + "\n" + content
	}

	borderColor := style.ANSIDim
	if t.designTokens != nil {
		if c := style.Fg(t.designTokens.BorderSubtle); c != "" {
			borderColor = c
		}
	}

	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}

	lines := strings.Split(content, "\n")
	var b strings.Builder

	top := strings.Repeat("─", innerWidth)
	if t.title != "" {
		title := " " + style.Truncate(t.title, innerWidth-2, "…") + " "
		titleWidth := style.StringWidth(title)
		if titleWidth < innerWidth {
			top = title + strings.Repeat("─", innerWidth-titleWidth)
		} else {
			top = style.Truncate(title, innerWidth, "…")
		}
	}

	b.WriteString(borderColor + "┌" + top + "┐" + style.ANSIReset)
	if len(lines) > 0 {
		b.WriteByte('\n')
	}

	for i, line := range lines {
		trimmed := style.Truncate(line, innerWidth, "…")
		pad := innerWidth - style.StringWidth(trimmed)
		if pad < 0 {
			pad = 0
		}
		b.WriteString(borderColor + "│" + style.ANSIReset)
		b.WriteString(trimmed)
		b.WriteString(strings.Repeat(" ", pad))
		b.WriteString(borderColor + "│" + style.ANSIReset)
		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}

	if len(lines) > 0 {
		b.WriteByte('\n')
	}
	b.WriteString(borderColor + "└" + strings.Repeat("─", innerWidth) + "┘" + style.ANSIReset)

	return b.String()
}

// Focus marks the component as focused.
func (t *ThemedBox) Focus() { t.focused = true }

// Blur marks the component as unfocused.
func (t *ThemedBox) Blur() { t.focused = false }

// Focused reports whether the component is focused.
func (t *ThemedBox) Focused() bool { return t.focused }

var _ tui.Component = (*ThemedBox)(nil)
