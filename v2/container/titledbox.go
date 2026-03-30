package container

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// TitledBox renders a bordered container with a title embedded in the top border.
type TitledBox struct {
	title   string
	content string
	width   int

	focused bool

	borderColor string
	titleColor  string

	designTokens *design.DesignTokens
}

// TitledBoxOption configures a TitledBox.
type TitledBoxOption func(*TitledBox)

// WithTitledBoxWidth sets the render width.
func WithTitledBoxWidth(width int) TitledBoxOption {
	return func(t *TitledBox) {
		if width > 0 {
			t.width = width
		}
	}
}

// WithTitledBoxBorderColor sets the border color escape sequence.
func WithTitledBoxBorderColor(color string) TitledBoxOption {
	return func(t *TitledBox) {
		t.borderColor = color
	}
}

// WithTitledBoxTitleColor sets the title color escape sequence.
func WithTitledBoxTitleColor(color string) TitledBoxOption {
	return func(t *TitledBox) {
		t.titleColor = color
	}
}

// WithTitledBoxDesignTokens applies design-system tokens to titled box colors.
func WithTitledBoxDesignTokens(tokens *design.DesignTokens) TitledBoxOption {
	return func(t *TitledBox) {
		t.designTokens = tokens
	}
}

// NewTitledBox creates a new titled box container.
func NewTitledBox(title string, content string, opts ...TitledBoxOption) *TitledBox {
	t := &TitledBox{
		title:        title,
		content:      content,
		width:        60,
		borderColor:  style.ANSIDim,
		titleColor:   style.ANSIWhite,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// RenderTitledBox renders a titled box with an optional fixed width.
func RenderTitledBox(title, content string, width int) string {
	opts := make([]TitledBoxOption, 0, 1)
	if width > 0 {
		opts = append(opts, WithTitledBoxWidth(width))
	}
	return NewTitledBox(title, content, opts...).View()
}

// Init initializes the component.
func (t *TitledBox) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages.
func (t *TitledBox) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	_ = msg
	return t, nil
}

// View renders the titled box.
func (t *TitledBox) View() string {
	width := t.effectiveWidth()
	innerWidth := width - 4
	if innerWidth < 1 {
		innerWidth = 1
	}

	borderColor, titleColor := t.resolveColors()

	title := strings.TrimSpace(t.title)
	if title == "" {
		title = "Title"
	}
	title = style.Truncate(title, max(1, width-5), "…")

	topPrefix := "─── "
	topTitleSuffix := " "
	topPrefixWidth := style.StringWidth(topPrefix)
	topTitleWidth := style.StringWidth(title)
	topSuffixWidth := style.StringWidth(topTitleSuffix)
	remainingTop := width - topPrefixWidth - topTitleWidth - topSuffixWidth
	if remainingTop < 0 {
		remainingTop = 0
	}

	contentLines := strings.Split(t.content, "\n")
	if len(contentLines) == 0 {
		contentLines = []string{""}
	}

	var b strings.Builder

	b.WriteString(applyColor(borderColor, topPrefix))
	b.WriteString(applyColor(titleColor+style.ANSIBold, title))
	b.WriteString(applyColor(borderColor, topTitleSuffix+strings.Repeat("─", remainingTop)))
	b.WriteString("\n")

	for i, line := range contentLines {
		line = style.Truncate(line, innerWidth, "…")
		padding := innerWidth - style.StringWidth(line)
		if padding < 0 {
			padding = 0
		}

		b.WriteString(applyColor(borderColor, "│"))
		b.WriteString(" ")
		b.WriteString(line)
		b.WriteString(strings.Repeat(" ", padding))
		b.WriteString(" ")
		b.WriteString(applyColor(borderColor, "│"))
		if i < len(contentLines)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(applyColor(borderColor, strings.Repeat("─", width)))

	return b.String()
}

// Focus marks the component as focused.
func (t *TitledBox) Focus() {
	t.focused = true
}

// Blur marks the component as unfocused.
func (t *TitledBox) Blur() {
	t.focused = false
}

// Focused reports whether the component is focused.
func (t *TitledBox) Focused() bool {
	return t.focused
}

func (t *TitledBox) effectiveWidth() int {
	if t.width > 0 {
		return max(8, t.width)
	}

	computed := 8
	titleWidth := style.StringWidth(strings.TrimSpace(t.title)) + 5
	if titleWidth > computed {
		computed = titleWidth
	}
	for _, line := range strings.Split(t.content, "\n") {
		lineWidth := style.StringWidth(line) + 4
		if lineWidth > computed {
			computed = lineWidth
		}
	}
	if computed < 8 {
		computed = 8
	}
	return computed
}

func (t *TitledBox) resolveColors() (string, string) {
	borderColor := t.borderColor
	titleColor := t.titleColor

	if tokens := t.designTokens; tokens != nil {
		if borderColor == "" {
			if c := style.Fg(tokens.BorderSubtle); c != "" {
				borderColor = c
			} else if c := style.Fg(tokens.Accent); c != "" {
				borderColor = c
			}
		}
		if titleColor == "" {
			if c := style.Fg(tokens.Accent); c != "" {
				titleColor = c
			} else if c := style.Fg(tokens.Color); c != "" {
				titleColor = c
			}
		}
	}

	if borderColor == "" {
		borderColor = style.ANSIDim
	}
	if titleColor == "" {
		titleColor = style.ANSIWhite
	}

	return borderColor, titleColor
}

func applyColor(color, text string) string {
	if color == "" {
		return text
	}
	return color + text + style.ANSIReset
}

var _ tui.Component = (*TitledBox)(nil)
