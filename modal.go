package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ModalType defines the type of modal dialog
type ModalType int

const (
	// ModalAlert shows a message with an OK button
	ModalAlert ModalType = iota
	// ModalConfirm shows a message with Yes/No or OK/Cancel buttons
	ModalConfirm
	// ModalInput shows a message with a text input field
	ModalInput
)

// ModalButton represents a button in the modal
type ModalButton struct {
	Label  string
	Action func(string) tea.Cmd // Input value passed for ModalInput, empty string otherwise
}

// Modal displays overlay dialogs for user interaction
type Modal struct {
	width      int
	height     int
	visible    bool
	focused    bool
	modalType  ModalType
	title      string
	message    string
	buttons    []ModalButton
	selected   int // Selected button index
	textInput  textinput.Model
	hasInput   bool
	onConfirm  func(string) tea.Cmd
	onCancel   func() tea.Cmd
}

// ModalOption configures a Modal
type ModalOption func(*Modal)

// WithModalType sets the modal type
func WithModalType(t ModalType) ModalOption {
	return func(m *Modal) {
		m.modalType = t
	}
}

// WithModalTitle sets the modal title
func WithModalTitle(title string) ModalOption {
	return func(m *Modal) {
		m.title = title
	}
}

// WithModalMessage sets the modal message
func WithModalMessage(message string) ModalOption {
	return func(m *Modal) {
		m.message = message
	}
}

// WithModalButtons sets custom buttons
func WithModalButtons(buttons []ModalButton) ModalOption {
	return func(m *Modal) {
		m.buttons = buttons
	}
}

// WithModalInput enables text input (for ModalInput type)
func WithModalInput(placeholder string) ModalOption {
	return func(m *Modal) {
		m.hasInput = true
		m.textInput.Placeholder = placeholder
	}
}

// WithModalOnConfirm sets the confirm callback
func WithModalOnConfirm(fn func(string) tea.Cmd) ModalOption {
	return func(m *Modal) {
		m.onConfirm = fn
	}
}

// WithModalOnCancel sets the cancel callback
func WithModalOnCancel(fn func() tea.Cmd) ModalOption {
	return func(m *Modal) {
		m.onCancel = fn
	}
}

// NewModal creates a new modal dialog
func NewModal(opts ...ModalOption) *Modal {
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 200
	ti.Width = 40

	m := &Modal{
		textInput: ti,
		visible:   false,
		modalType: ModalAlert,
	}

	for _, opt := range opts {
		opt(m)
	}

	// Set default buttons based on type if none provided
	if len(m.buttons) == 0 {
		switch m.modalType {
		case ModalAlert:
			m.buttons = []ModalButton{
				{Label: "OK", Action: func(s string) tea.Cmd { return nil }},
			}
		case ModalConfirm:
			m.buttons = []ModalButton{
				{Label: "Yes", Action: func(s string) tea.Cmd { return nil }},
				{Label: "No", Action: func(s string) tea.Cmd { return nil }},
			}
		case ModalInput:
			m.buttons = []ModalButton{
				{Label: "OK", Action: func(s string) tea.Cmd { return nil }},
				{Label: "Cancel", Action: func(s string) tea.Cmd { return nil }},
			}
			m.hasInput = true
		}
	}

	return m
}

// Init initializes the modal
func (m *Modal) Init() tea.Cmd {
	if m.hasInput {
		return textinput.Blink
	}
	return nil
}

// Update handles messages
func (m *Modal) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if !m.focused || !m.visible {
			return m, nil
		}

		switch msg.Type {
		case tea.KeyEsc:
			// Cancel/close modal
			m.Hide()
			if m.onCancel != nil {
				return m, m.onCancel()
			}
			return m, nil

		case tea.KeyEnter:
			// Activate selected button
			if m.selected < len(m.buttons) {
				btn := m.buttons[m.selected]
				value := ""
				if m.hasInput {
					value = m.textInput.Value()
				}
				m.Hide()
				if btn.Action != nil {
					return m, btn.Action(value)
				}
			}
			return m, nil

		case tea.KeyTab, tea.KeyRight:
			// Move to next button
			if m.selected < len(m.buttons)-1 {
				m.selected++
			} else {
				m.selected = 0
			}
			return m, nil

		case tea.KeyShiftTab, tea.KeyLeft:
			// Move to previous button
			if m.selected > 0 {
				m.selected--
			} else {
				m.selected = len(m.buttons) - 1
			}
			return m, nil

		default:
			// Pass to text input if present
			if m.hasInput {
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		}
	}

	// Update text input if focused and has input
	if m.visible && m.focused && m.hasInput {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

// View renders the modal
func (m *Modal) View() string {
	if !m.visible || m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Calculate dimensions - ensure we don't exceed terminal width
	modalWidth := min(60, m.width-4)
	messageLines := wrapText(m.message, modalWidth-4)
	startX := (m.width - modalWidth) / 2
	if startX < 0 {
		startX = 0
	}

	// Render minimal backdrop (just 2 lines of spacing)
	b.WriteString("\n\n")

	// Top border with integrated title
	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString("╭─")
	title := m.title
	if title == "" {
		title = "Dialog"
	}
	titleText := "── " + title + " "
	b.WriteString(titleText)
	// Calculate remaining width more carefully
	remainingWidth := modalWidth - len(titleText) - 4 // Account for ╭─ and ─╮
	if remainingWidth > 0 {
		b.WriteString(strings.Repeat("─", remainingWidth))
	}
	b.WriteString("╮\n")

	// Empty line
	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString("│")
	b.WriteString(strings.Repeat(" ", modalWidth-2))
	b.WriteString("│\n")

	// Message content
	for _, line := range messageLines {
		b.WriteString(strings.Repeat(" ", startX))
		b.WriteString("│ ")
		b.WriteString(line)
		// Pad to width
		if len(line) < modalWidth-4 {
			b.WriteString(strings.Repeat(" ", modalWidth-4-len(line)))
		}
		b.WriteString(" │\n")
	}

	// Empty line
	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString("│")
	b.WriteString(strings.Repeat(" ", modalWidth-2))
	b.WriteString("│\n")

	// Text input (if present)
	if m.hasInput {
		b.WriteString(strings.Repeat(" ", startX))
		b.WriteString("│ ")
		inputView := m.textInput.View()
		b.WriteString(inputView)
		inputLen := len(stripANSI(inputView))
		if inputLen < modalWidth-4 {
			b.WriteString(strings.Repeat(" ", modalWidth-4-inputLen))
		}
		b.WriteString(" │\n")

		// Empty line after input
		b.WriteString(strings.Repeat(" ", startX))
		b.WriteString("│")
		b.WriteString(strings.Repeat(" ", modalWidth-2))
		b.WriteString("│\n")
	}

	// Buttons
	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString("│")

	// Calculate button layout
	totalButtonWidth := 0
	for _, btn := range m.buttons {
		totalButtonWidth += len(btn.Label) + 4 // [Label] + spacing
	}

	buttonStartX := (modalWidth - totalButtonWidth) / 2
	b.WriteString(strings.Repeat(" ", buttonStartX))

	for i, btn := range m.buttons {
		if i == m.selected {
			// Highlighted button
			b.WriteString("\033[7m[ ")
			b.WriteString(btn.Label)
			b.WriteString(" ]\033[0m")
		} else {
			// Normal button
			b.WriteString("\033[2m[ \033[0m")
			b.WriteString(btn.Label)
			b.WriteString("\033[2m ]\033[0m")
		}

		if i < len(m.buttons)-1 {
			b.WriteString("  ")
		}
	}

	// Pad to width (modalWidth-2 for interior, minus what we've used)
	padding := (modalWidth - 2) - buttonStartX - totalButtonWidth
	if padding > 0 {
		b.WriteString(strings.Repeat(" ", padding))
	}
	b.WriteString("│\n")

	// Empty line
	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString("│")
	b.WriteString(strings.Repeat(" ", modalWidth-2))
	b.WriteString("│\n")

	// Bottom border with hints
	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString("╰")
	hints := "─ Tab: navigate · Enter: confirm · Esc: cancel "
	// Calculate remaining dash width: modalWidth - corners(2) - hints length
	remainingDashes := modalWidth - 2 - len(hints)
	if remainingDashes > 0 {
		b.WriteString("\033[2m")
		b.WriteString(hints)
		b.WriteString(strings.Repeat("─", remainingDashes))
		b.WriteString("\033[0m")
	} else {
		// If hints too long, just use dashes
		b.WriteString(strings.Repeat("─", modalWidth-2))
	}
	b.WriteString("╯\n")

	return b.String()
}

// Focus is called when this component receives focus
func (m *Modal) Focus() {
	m.focused = true
	if m.hasInput {
		m.textInput.Focus()
	}
}

// Blur is called when this component loses focus
func (m *Modal) Blur() {
	m.focused = false
	if m.hasInput {
		m.textInput.Blur()
	}
}

// Focused returns whether this component is currently focused
func (m *Modal) Focused() bool {
	return m.focused
}

// Show displays the modal
func (m *Modal) Show() {
	m.visible = true
	m.selected = 0
	if m.hasInput {
		m.textInput.SetValue("")
		m.textInput.Focus()
	}
}

// Hide conceals the modal
func (m *Modal) Hide() {
	m.visible = false
	if m.hasInput {
		m.textInput.Blur()
	}
}

// IsVisible returns whether the modal is currently visible
func (m *Modal) IsVisible() bool {
	return m.visible
}

// SetTitle updates the modal title
func (m *Modal) SetTitle(title string) {
	m.title = title
}

// SetMessage updates the modal message
func (m *Modal) SetMessage(message string) {
	m.message = message
}

// ShowAlert displays an alert modal
func (m *Modal) ShowAlert(title, message string, onOK func() tea.Cmd) {
	m.modalType = ModalAlert
	m.title = title
	m.message = message
	m.buttons = []ModalButton{
		{Label: "OK", Action: func(s string) tea.Cmd {
			if onOK != nil {
				return onOK()
			}
			return nil
		}},
	}
	m.hasInput = false
	m.Show()
}

// ShowConfirm displays a confirmation modal
func (m *Modal) ShowConfirm(title, message string, onYes, onNo func() tea.Cmd) {
	m.modalType = ModalConfirm
	m.title = title
	m.message = message
	m.buttons = []ModalButton{
		{Label: "Yes", Action: func(s string) tea.Cmd {
			if onYes != nil {
				return onYes()
			}
			return nil
		}},
		{Label: "No", Action: func(s string) tea.Cmd {
			if onNo != nil {
				return onNo()
			}
			return nil
		}},
	}
	m.hasInput = false
	m.Show()
}

// ShowInput displays an input modal
func (m *Modal) ShowInput(title, message, placeholder string, onOK func(string) tea.Cmd, onCancel func() tea.Cmd) {
	m.modalType = ModalInput
	m.title = title
	m.message = message
	m.textInput.Placeholder = placeholder
	m.buttons = []ModalButton{
		{Label: "OK", Action: func(s string) tea.Cmd {
			if onOK != nil {
				return onOK(s)
			}
			return nil
		}},
		{Label: "Cancel", Action: func(s string) tea.Cmd {
			if onCancel != nil {
				return onCancel()
			}
			return nil
		}},
	}
	m.hasInput = true
	m.Show()
}

// wrapText wraps text to fit within a given width
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		if currentLine.Len() == 0 {
			currentLine.WriteString(word)
		} else if currentLine.Len()+1+len(word) <= width {
			currentLine.WriteString(" ")
			currentLine.WriteString(word)
		} else {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		}
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}
