package input

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ContinuationPromptOption configures a ContinuationPrompt.
type ContinuationPromptOption func(*ContinuationPrompt)

// WithContinuationPromptDesignTokens applies design-system colors.
func WithContinuationPromptDesignTokens(tokens *design.DesignTokens) ContinuationPromptOption {
	return func(c *ContinuationPrompt) {
		c.applyDesignTokens(tokens)
	}
}

// WithContinuationPromptWidth sets a preferred prompt width.
func WithContinuationPromptWidth(width int) ContinuationPromptOption {
	return func(c *ContinuationPrompt) {
		if width > 0 {
			c.width = width
		}
	}
}

// WithContinuationPromptPlaceholder sets placeholder text.
func WithContinuationPromptPlaceholder(placeholder string) ContinuationPromptOption {
	return func(c *ContinuationPrompt) {
		if strings.TrimSpace(placeholder) != "" {
			c.placeholder = placeholder
		}
	}
}

// WithContinuationPromptPrefix sets the line prefix.
func WithContinuationPromptPrefix(prefix string) ContinuationPromptOption {
	return func(c *ContinuationPrompt) {
		if strings.TrimSpace(prefix) != "" {
			c.prefix = prefix
		}
	}
}

// WithContinuationPromptHistory pre-populates prompt history.
func WithContinuationPromptHistory(history []string) ContinuationPromptOption {
	return func(c *ContinuationPrompt) {
		if len(history) == 0 {
			return
		}

		c.history = append([]string(nil), history...)
		c.historyIndex = -1
	}
}

// ContinuationPrompt is a single-line multi-turn prompt with history browsing.
type ContinuationPrompt struct {
	input       string
	cursorPos   int
	history     []string
	historyIndex int // -1 = current input, 0..N = history entries
	placeholder string
	prefix      string
	width       int
	focused     bool

	designTokens     *design.DesignTokens
	prefixStyle      string
	placeholderStyle string
	draftInput       string
}

// NewContinuationPrompt creates a new ContinuationPrompt.
func NewContinuationPrompt(opts ...ContinuationPromptOption) *ContinuationPrompt {
	c := &ContinuationPrompt{
		input:            "",
		cursorPos:        0,
		history:          make([]string, 0),
		historyIndex:     -1,
		placeholder:      "type your message...",
		prefix:           "❯ ",
		width:            0,
		focused:          false,
		designTokens:     design.DefaultTheme(),
		prefixStyle:      style.ANSIReset,
		placeholderStyle: style.ANSIDim,
		draftInput:       "",
	}

	for _, opt := range opts {
		opt(c)
	}

	c.applyDesignTokens(c.designTokens)
	c.clampCursor()
	return c
}

// Init initializes the continuation prompt.
func (c *ContinuationPrompt) Init() tea.Cmd {
	return nil
}

// Update handles Bubble Tea messages.
func (c *ContinuationPrompt) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if c.width <= 0 {
			c.width = msg.Width
		}
		return c, nil

	case tea.KeyMsg:
		if !c.focused {
			return c, nil
		}

		switch msg.Type {
		case tea.KeyEnter:
			submitted := strings.TrimSpace(c.input)
			if submitted == "" {
				return c, nil
			}

			full := c.input
			c.history = append(c.history, full)
			c.input = ""
			c.cursorPos = 0
			c.historyIndex = -1
			c.draftInput = ""

			return c, func() tea.Msg {
				return PromptSubmitMsg{Text: full}
			}

		case tea.KeyUp:
			c.browseHistoryUp()
			return c, nil

		case tea.KeyDown:
			c.browseHistoryDown()
			return c, nil

		case tea.KeyCtrlA:
			c.cursorPos = 0
			return c, nil

		case tea.KeyCtrlE:
			c.cursorPos = runeLen(c.input)
			return c, nil

		case tea.KeyBackspace:
			c.deleteBeforeCursor()
			return c, nil

		case tea.KeyCtrlU:
			c.input = ""
			c.cursorPos = 0
			if c.historyIndex == -1 {
				c.draftInput = ""
			}
			return c, nil

		case tea.KeyRunes:
			if len(msg.Runes) > 0 {
				c.insertRunes(string(msg.Runes))
				return c, nil
			}
		}
	}

	return c, nil
}

// View renders the prompt line.
func (c *ContinuationPrompt) View() string {
	prefix := c.prefixStyle + c.prefix + style.ANSIReset

	if c.input == "" {
		line := prefix + c.placeholderStyle + c.placeholder + style.ANSIReset
		return c.fitWidth(line)
	}

	left, right := splitAtCursor(c.input, c.cursorPos)
	cursor := ""
	if c.focused {
		cursor = "█"
	}

	line := prefix + left + cursor + right
	return c.fitWidth(line)
}

// Focus marks the prompt as focused.
func (c *ContinuationPrompt) Focus() {
	c.focused = true
}

// Blur marks the prompt as unfocused.
func (c *ContinuationPrompt) Blur() {
	c.focused = false
}

// Focused reports whether the prompt has focus.
func (c *ContinuationPrompt) Focused() bool {
	return c.focused
}

func (c *ContinuationPrompt) applyDesignTokens(tokens *design.DesignTokens) {
	c.designTokens = tokens
	if tokens == nil {
		c.prefixStyle = style.ANSIReset
		c.placeholderStyle = style.ANSIDim
		return
	}

	c.prefixStyle = style.ANSIReset
	if accent := style.ANSIColorFromHex(tokens.Accent); accent != "" {
		c.prefixStyle = accent
	}

	c.placeholderStyle = style.ANSIDim
	if muted := style.ANSIColorFromHex(tokens.MutedColor); muted != "" {
		c.placeholderStyle = muted
	} else if color := style.ANSIColorFromHex(tokens.Color); color != "" {
		c.placeholderStyle = color
	}
}

func (c *ContinuationPrompt) browseHistoryUp() {
	if len(c.history) == 0 {
		return
	}

	if c.historyIndex == -1 {
		c.draftInput = c.input
		c.historyIndex = len(c.history) - 1
	} else if c.historyIndex > 0 {
		c.historyIndex--
	}

	c.input = c.history[c.historyIndex]
	c.cursorPos = runeLen(c.input)
}

func (c *ContinuationPrompt) browseHistoryDown() {
	if c.historyIndex == -1 {
		return
	}

	if c.historyIndex < len(c.history)-1 {
		c.historyIndex++
		c.input = c.history[c.historyIndex]
	} else {
		c.historyIndex = -1
		c.input = c.draftInput
	}

	c.cursorPos = runeLen(c.input)
}

func (c *ContinuationPrompt) insertRunes(s string) {
	if c.historyIndex != -1 {
		c.historyIndex = -1
	}

	left, right := splitAtCursor(c.input, c.cursorPos)
	c.input = left + s + right
	c.cursorPos += runeLen(s)
	c.draftInput = c.input
}

func (c *ContinuationPrompt) deleteBeforeCursor() {
	if c.cursorPos <= 0 {
		return
	}

	if c.historyIndex != -1 {
		c.historyIndex = -1
	}

	runes := []rune(c.input)
	idx := c.cursorPos - 1
	if idx < 0 || idx >= len(runes) {
		return
	}

	c.input = string(append(runes[:idx], runes[idx+1:]...))
	c.cursorPos--
	c.clampCursor()
	c.draftInput = c.input
}

func (c *ContinuationPrompt) clampCursor() {
	max := runeLen(c.input)
	if c.cursorPos < 0 {
		c.cursorPos = 0
	}
	if c.cursorPos > max {
		c.cursorPos = max
	}
}

func (c *ContinuationPrompt) fitWidth(line string) string {
	if c.width <= 0 {
		return line
	}
	return style.Truncate(line, c.width, "…")
}

func splitAtCursor(s string, cursor int) (string, string) {
	runes := []rune(s)
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}
	return string(runes[:cursor]), string(runes[cursor:])
}

func runeLen(s string) int {
	return len([]rune(s))
}

var _ tui.Component = (*ContinuationPrompt)(nil)
