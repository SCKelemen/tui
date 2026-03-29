package input

import (
	"strings"
	"time"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// PromptSubmitMsg is emitted when a prompt is submitted.
type PromptSubmitMsg struct {
	Text string
}

// PromptHistoryEntry stores previously submitted prompt content.
type PromptHistoryEntry struct {
	Text      string
	Timestamp time.Time
}

// PromptOption configures a Prompt.
type PromptOption func(*Prompt)

// WithPromptPlaceholder sets the textarea placeholder text.
func WithPromptPlaceholder(placeholder string) PromptOption {
	return func(p *Prompt) {
		if strings.TrimSpace(placeholder) != "" {
			p.placeholder = placeholder
			p.textarea.Placeholder = placeholder
		}
	}
}

// WithPromptBorderColor sets the border color as a hex token.
func WithPromptBorderColor(color string) PromptOption {
	return func(p *Prompt) {
		if strings.TrimSpace(color) != "" {
			p.borderColor = color
		}
	}
}

// WithPromptWidth sets the preferred prompt width.
func WithPromptWidth(width int) PromptOption {
	return func(p *Prompt) {
		if width > 0 {
			p.width = width
			p.setTextareaWidth(width)
		}
	}
}

// WithPromptMaxHeight sets the maximum auto-expanded height in rows.
func WithPromptMaxHeight(maxHeight int) PromptOption {
	return func(p *Prompt) {
		if maxHeight > 0 {
			p.maxHeight = maxHeight
			p.textarea.MaxHeight = maxHeight
			p.adjustHeight()
		}
	}
}

// Prompt is a multi-line prompt input with history and stash support.
type Prompt struct {
	textarea    textarea.Model
	history     []PromptHistoryEntry
	historyIndex int
	stash       string
	placeholder string
	borderColor string
	focused     bool
	width       int
	height      int
	maxHeight   int
}

// NewPrompt creates a new Prompt component.
func NewPrompt(opts ...PromptOption) *Prompt {
	ta := textarea.New()
	ta.Placeholder = "Type your prompt..."
	ta.ShowLineNumbers = false
	ta.CharLimit = 10000
	ta.SetHeight(1)

	p := &Prompt{
		textarea:     ta,
		history:      make([]PromptHistoryEntry, 0),
		historyIndex: -1,
		placeholder:  "Type your prompt...",
		borderColor:  "#3C414B",
		maxHeight:    8,
		height:       3,
	}
	p.textarea.MaxHeight = p.maxHeight

	for _, opt := range opts {
		opt(p)
	}

	p.adjustHeight()
	return p
}

// Init initializes the prompt.
func (p *Prompt) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles Bubble Tea messages.
func (p *Prompt) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if p.width == 0 {
			p.width = msg.Width
		}
		p.setTextareaWidth(p.width)
		p.adjustHeight()
		return p, nil

	case tea.KeyMsg:
		if !p.focused {
			return p, nil
		}

		switch msg.Type {
		case tea.KeyEnter:
			content := strings.TrimSpace(p.textarea.Value())
			if content == "" {
				return p, nil
			}

			fullText := p.textarea.Value()
			p.history = append(p.history, PromptHistoryEntry{Text: fullText, Timestamp: time.Now()})
			p.historyIndex = -1
			p.stash = ""
			p.textarea.Reset()
			p.adjustHeight()

			return p, func() tea.Msg {
				return PromptSubmitMsg{Text: fullText}
			}

		case tea.KeyUp:
			if p.atTop() {
				p.browseHistoryUp()
				return p, nil
			}
		case tea.KeyDown:
			if p.atBottom() {
				p.browseHistoryDown()
				return p, nil
			}
		case tea.KeyCtrlS:
			p.stash = p.textarea.Value()
			return p, nil
		case tea.KeyCtrlR:
			p.historyIndex = -1
			p.textarea.SetValue(p.stash)
			p.adjustHeight()
			return p, nil
		case tea.KeyCtrlD:
			p.historyIndex = -1
			p.textarea.Reset()
			p.adjustHeight()
			return p, nil
		}
	}

	if p.focused {
		p.textarea, cmd = p.textarea.Update(msg)
		p.adjustHeight()
	}

	return p, cmd
}

// View renders the bordered prompt and controls hint.
func (p *Prompt) View() string {
	if p.width < 2 {
		return ""
	}

	border := style.ANSIDim
	if c := style.ANSIColorFromHex(p.borderColor); strings.TrimSpace(c) != "" {
		border = c
	}

	var b strings.Builder

	b.WriteString(border + "┌")
	b.WriteString(strings.Repeat("─", p.width-2))
	b.WriteString("┐" + style.ANSIReset + "\n")

	lines := strings.Split(p.textarea.View(), "\n")
	for _, line := range lines {
		b.WriteString(border + "│" + style.ANSIReset + " ")
		b.WriteString(style.Pad(line, p.width-4))
		b.WriteString(" " + border + "│" + style.ANSIReset + "\n")
	}

	b.WriteString(border + "└")
	if p.focused {
		hint := "Enter: send · ↑/↓: history · Ctrl+S: stash · Ctrl+R: restore · Ctrl+D: clear"
		hintLen := style.StringWidth(hint)
		if hintLen < p.width-4 {
			b.WriteString(" " + style.ANSIDim + "\033[3m")
			b.WriteString(hint)
			b.WriteString(style.ANSIReset + border + " ")
			b.WriteString(strings.Repeat("─", p.width-hintLen-6))
		} else {
			b.WriteString(strings.Repeat("─", p.width-2))
		}
	} else {
		b.WriteString(strings.Repeat("─", p.width-2))
	}
	b.WriteString("┘" + style.ANSIReset + "\n")

	return b.String()
}

// Focus marks the prompt as focused.
func (p *Prompt) Focus() {
	p.focused = true
	p.textarea.Focus()
}

// Blur marks the prompt as unfocused.
func (p *Prompt) Blur() {
	p.focused = false
	p.textarea.Blur()
}

// Focused returns whether the prompt is focused.
func (p *Prompt) Focused() bool {
	return p.focused
}

// SetValue sets the prompt text.
func (p *Prompt) SetValue(s string) {
	p.textarea.SetValue(s)
	p.adjustHeight()
}

// Value returns the prompt text.
func (p *Prompt) Value() string {
	return p.textarea.Value()
}

func (p *Prompt) browseHistoryUp() {
	if len(p.history) == 0 {
		return
	}

	if p.historyIndex == -1 {
		p.stash = p.textarea.Value()
		p.historyIndex = len(p.history) - 1
	} else if p.historyIndex > 0 {
		p.historyIndex--
	}

	p.textarea.SetValue(p.history[p.historyIndex].Text)
	p.adjustHeight()
}

func (p *Prompt) browseHistoryDown() {
	if p.historyIndex == -1 {
		return
	}

	if p.historyIndex < len(p.history)-1 {
		p.historyIndex++
		p.textarea.SetValue(p.history[p.historyIndex].Text)
	} else {
		p.historyIndex = -1
		p.textarea.SetValue(p.stash)
	}

	p.adjustHeight()
}

func (p *Prompt) atTop() bool {
	if p.textarea.Line() != 0 {
		return false
	}
	lineInfo := p.textarea.LineInfo()
	return lineInfo.RowOffset == 0
}

func (p *Prompt) atBottom() bool {
	lineCount := p.textarea.LineCount()
	if lineCount <= 0 {
		return true
	}
	if p.textarea.Line() != lineCount-1 {
		return false
	}
	lineInfo := p.textarea.LineInfo()
	if lineInfo.Height <= 0 {
		return true
	}
	return lineInfo.RowOffset >= lineInfo.Height-1
}

func (p *Prompt) setTextareaWidth(componentWidth int) {
	textareaWidth := componentWidth - 4
	if textareaWidth < 1 {
		textareaWidth = 1
	}
	p.textarea.SetWidth(textareaWidth)
}

func (p *Prompt) adjustHeight() {
	rows := strings.Count(p.textarea.Value(), "\n") + 1
	if rows < 1 {
		rows = 1
	}
	if p.maxHeight > 0 && rows > p.maxHeight {
		rows = p.maxHeight
	}

	p.textarea.SetHeight(rows)
	p.height = rows + 2
}

var _ tui.Component = (*Prompt)(nil)
