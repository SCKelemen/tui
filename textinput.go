package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// TextInput is a multi-line text input component for user messages
type TextInput struct {
	width      int
	height     int
	textarea   textarea.Model
	focused    bool
	placeholder string
	onSubmit   func(string) tea.Cmd
}

// NewTextInput creates a new text input component
func NewTextInput() *TextInput {
	ta := textarea.New()
	ta.Placeholder = "Type your message... (Ctrl+J to send)"
	ta.ShowLineNumbers = false
	ta.CharLimit = 10000
	ta.SetHeight(3)

	return &TextInput{
		textarea:    ta,
		placeholder: "Type your message... (Ctrl+J to send)",
		height:      5, // 3 lines + border
	}
}

// Init initializes the text input
func (t *TextInput) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages
func (t *TextInput) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.textarea.SetWidth(msg.Width - 4) // Account for border

	case tea.KeyMsg:
		if !t.focused {
			return t, nil
		}

		// Handle Ctrl+Enter to submit (Ctrl+J in terminal)
		if msg.Type == tea.KeyCtrlJ || (msg.Type == tea.KeyEnter && msg.Alt) {
			content := strings.TrimSpace(t.textarea.Value())
			if content != "" {
				t.textarea.Reset()
				if t.onSubmit != nil {
					return t, t.onSubmit(content)
				}
			}
			return t, nil
		}

		// Handle Ctrl+D to clear
		if msg.Type == tea.KeyCtrlD {
			t.textarea.Reset()
			return t, nil
		}
	}

	// Pass to textarea
	if t.focused {
		t.textarea, cmd = t.textarea.Update(msg)
	}

	return t, cmd
}

// View renders the text input
func (t *TextInput) View() string {
	if t.width == 0 {
		return ""
	}

	var b strings.Builder

	// Top border
	b.WriteString("\033[2m┌")
	b.WriteString(strings.Repeat("─", t.width-2))
	b.WriteString("┐\033[0m\n")

	// Textarea content
	lines := strings.Split(t.textarea.View(), "\n")
	for _, line := range lines {
		b.WriteString("\033[2m│\033[0m ")
		b.WriteString(line)

		// Pad to width
		visualLen := len(stripANSI(line))
		if visualLen < t.width-4 {
			b.WriteString(strings.Repeat(" ", t.width-4-visualLen))
		}

		b.WriteString(" \033[2m│\033[0m\n")
	}

	// Bottom border with hint
	b.WriteString("\033[2m└")
	if t.focused {
		hint := "Ctrl+J: send · Ctrl+D: clear"
		hintLen := len(hint)
		if hintLen < t.width-4 {
			b.WriteString(" \033[3m")
			b.WriteString(hint)
			b.WriteString("\033[0m\033[2m ")
			b.WriteString(strings.Repeat("─", t.width-hintLen-6))
		} else {
			b.WriteString(strings.Repeat("─", t.width-2))
		}
	} else {
		b.WriteString(strings.Repeat("─", t.width-2))
	}
	b.WriteString("┘\033[0m\n")

	return b.String()
}

// Focus is called when this component receives focus
func (t *TextInput) Focus() {
	t.focused = true
	t.textarea.Focus()
}

// Blur is called when this component loses focus
func (t *TextInput) Blur() {
	t.focused = false
	t.textarea.Blur()
}

// Focused returns whether this component is currently focused
func (t *TextInput) Focused() bool {
	return t.focused
}

// OnSubmit sets the callback for when text is submitted
func (t *TextInput) OnSubmit(fn func(string) tea.Cmd) {
	t.onSubmit = fn
}

// Value returns the current text value
func (t *TextInput) Value() string {
	return t.textarea.Value()
}

// SetValue sets the text value
func (t *TextInput) SetValue(value string) {
	t.textarea.SetValue(value)
}

// Reset clears the text input
func (t *TextInput) Reset() {
	t.textarea.Reset()
}
