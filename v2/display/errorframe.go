package display

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ErrorFrame displays a bordered error box with a title and message body.
type ErrorFrame struct {
	title        string
	message      string
	width        int
	windowWidth  int
	focused      bool
	designTokens *design.DesignTokens
}

// ErrorFrameOption configures an ErrorFrame.
type ErrorFrameOption func(*ErrorFrame)

// WithErrorFrameWidth sets a fixed content width for the error frame.
// If width is 0, the frame auto-sizes to content.
func WithErrorFrameWidth(width int) ErrorFrameOption {
	return func(e *ErrorFrame) {
		if width >= 0 {
			e.width = width
		}
	}
}

// WithErrorFrameDesignTokens applies design tokens to the error frame.
func WithErrorFrameDesignTokens(tokens *design.DesignTokens) ErrorFrameOption {
	return func(e *ErrorFrame) {
		e.designTokens = tokens
	}
}

// WithErrorFrameTheme applies a named design-system theme.
func WithErrorFrameTheme(theme string) ErrorFrameOption {
	return func(e *ErrorFrame) {
		e.designTokens = style.DesignTokensForTheme(theme)
	}
}

// NewErrorFrame creates a new ErrorFrame component.
func NewErrorFrame(title, message string, opts ...ErrorFrameOption) *ErrorFrame {
	e := &ErrorFrame{
		title:        title,
		message:      message,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Init initializes the error frame.
func (e *ErrorFrame) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (e *ErrorFrame) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.windowWidth = msg.Width
	}

	return e, nil
}

// View renders the error frame.
func (e *ErrorFrame) View() string {
	contentWidth := e.contentWidth()
	if contentWidth < 1 {
		contentWidth = 1
	}

	titleText := elideErrorFrameTitle(e.title, contentWidth)
	messageLines := e.wrappedMessageLines(contentWidth)
	if len(messageLines) == 0 {
		messageLines = []string{""}
	}

	innerWidth := contentWidth + 2
	borderColor := e.errorColor()
	titleColor := style.ANSIBold + borderColor
	messageColor := borderColor

	var b strings.Builder

	b.WriteString(borderColor)
	b.WriteString("┌")
	b.WriteString(strings.Repeat("─", innerWidth))
	b.WriteString("┐")
	b.WriteString(style.ANSIReset)
	b.WriteString("\n")

	b.WriteString(e.renderContentLine(titleText, contentWidth, borderColor, titleColor))
	b.WriteString("\n")

	for i, line := range messageLines {
		b.WriteString(e.renderContentLine(line, contentWidth, borderColor, messageColor))
		if i < len(messageLines)-1 {
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")

	b.WriteString(borderColor)
	b.WriteString("└")
	b.WriteString(strings.Repeat("─", innerWidth))
	b.WriteString("┘")
	b.WriteString(style.ANSIReset)
	b.WriteString("\n")

	return b.String()
}

// Focus marks the component as focused.
func (e *ErrorFrame) Focus() {
	e.focused = true
}

// Blur marks the component as unfocused.
func (e *ErrorFrame) Blur() {
	e.focused = false
}

// Focused returns whether the component is focused.
func (e *ErrorFrame) Focused() bool {
	return e.focused
}

// SetTitle updates the frame title.
func (e *ErrorFrame) SetTitle(s string) {
	e.title = s
}

// SetMessage updates the frame message.
func (e *ErrorFrame) SetMessage(s string) {
	e.message = s
}

func (e *ErrorFrame) renderContentLine(text string, contentWidth int, borderColor, textColor string) string {
	if style.StringWidth(text) > contentWidth {
		text = style.Truncate(text, contentWidth, "…")
	}

	padding := contentWidth - style.StringWidth(text)
	if padding < 0 {
		padding = 0
	}

	var b strings.Builder
	b.WriteString(borderColor)
	b.WriteString("│")
	b.WriteString(style.ANSIReset)
	b.WriteString(" ")
	b.WriteString(textColor)
	b.WriteString(text)
	b.WriteString(style.ANSIReset)
	b.WriteString(strings.Repeat(" ", padding))
	b.WriteString(" ")
	b.WriteString(borderColor)
	b.WriteString("│")
	b.WriteString(style.ANSIReset)
	return b.String()
}

func (e *ErrorFrame) contentWidth() int {
	if e.width > 0 {
		return e.width
	}

	longest := style.StringWidth(e.title)
	for _, line := range strings.Split(e.message, "\n") {
		if n := style.StringWidth(line); n > longest {
			longest = n
		}
	}

	if longest < 1 {
		longest = 1
	}

	if e.windowWidth > 0 {
		maxContent := e.windowWidth - 4
		if maxContent < 1 {
			maxContent = 1
		}
		if longest > maxContent {
			longest = maxContent
		}
	}

	return longest
}

func (e *ErrorFrame) wrappedMessageLines(width int) []string {
	rawLines := strings.Split(e.message, "\n")
	if len(rawLines) == 0 {
		return []string{""}
	}

	wrapped := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		parts := wrapLine(line, width)
		wrapped = append(wrapped, parts...)
	}
	if len(wrapped) == 0 {
		return []string{""}
	}
	return wrapped
}

func (e *ErrorFrame) errorColor() string {
	if e.designTokens != nil {
		if c := style.ANSIColorFromHex(e.designTokens.Accent); c != "" {
			return c
		}
	}
	return style.ANSIRed
}

func elideErrorFrameTitle(title string, maxWidth int) string {
	if strings.HasPrefix(title, "http") {
		return style.ElideURL(title, maxWidth)
	}
	if strings.Contains(title, "/") {
		return style.ElidePath(title, maxWidth)
	}
	return style.Truncate(title, maxWidth, "…")
}

func wrapLine(line string, width int) []string {
	if width <= 0 {
		return []string{line}
	}
	if line == "" {
		return []string{""}
	}

	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}

	lines := make([]string, 0, 1)
	current := ""

	for _, word := range words {
		for style.StringWidth(word) > width {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
			chunk, rest := splitAtRunes(word, width)
			lines = append(lines, chunk)
			word = rest
		}

		if word == "" {
			continue
		}

		if current == "" {
			current = word
			continue
		}

		if style.StringWidth(current)+1+style.StringWidth(word) <= width {
			current += " " + word
		} else {
			lines = append(lines, current)
			current = word
		}
	}

	if current != "" {
		lines = append(lines, current)
	}

	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func splitAtRunes(s string, n int) (string, string) {
	if n <= 0 {
		return "", s
	}

	runes := []rune(s)
	if len(runes) <= n {
		return s, ""
	}
	return string(runes[:n]), string(runes[n:])
}

var _ tui.Component = (*ErrorFrame)(nil)
