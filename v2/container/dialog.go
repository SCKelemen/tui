package container

import (
	"strings"

	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DialogSize controls the dialog width preset.
type DialogSize int

const (
	// DialogMedium is the default width preset.
	DialogMedium DialogSize = iota
	// DialogLarge is a wider width preset.
	DialogLarge
)

// DialogResultMsg is emitted by dialog content models when the user makes a choice.
type DialogResultMsg struct {
	ID        string
	Value     interface{}
	Cancelled bool
}

// ShowDialogMsg requests that a dialog is pushed onto the manager stack.
type ShowDialogMsg struct {
	Dialog *Dialog
}

// CloseDialogMsg requests that the top dialog is closed.
type CloseDialogMsg struct{}

// Dialog is a generic container that wraps a tea.Model with dialog chrome.
type Dialog struct {
	ID      string
	Title   string
	Size    DialogSize
	Content tea.Model

	width  int
	height int
}

// DialogOption configures a Dialog.
type DialogOption func(*Dialog)

// WithDialogID sets the dialog ID.
func WithDialogID(id string) DialogOption {
	return func(d *Dialog) {
		d.ID = id
	}
}

// WithDialogTitle sets the dialog title.
func WithDialogTitle(title string) DialogOption {
	return func(d *Dialog) {
		d.Title = title
	}
}

// WithDialogSize sets the dialog size preset.
func WithDialogSize(size DialogSize) DialogOption {
	return func(d *Dialog) {
		d.Size = size
	}
}

// NewDialog creates a new dialog container.
func NewDialog(content tea.Model, opts ...DialogOption) *Dialog {
	d := &Dialog{
		Title:   "Dialog",
		Size:    DialogMedium,
		Content: content,
	}

	for _, opt := range opts {
		opt(d)
	}

	if d.Content == nil {
		d.Content = &dialogEmptyModel{}
	}

	return d
}

// Init initializes the wrapped content model.
func (d *Dialog) Init() tea.Cmd {
	if d.Content == nil {
		return nil
	}
	return d.Content.Init()
}

// Update forwards messages to the wrapped model and tracks current window size.
func (d *Dialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = m.Width
		d.height = m.Height
	}

	if d.Content == nil {
		return d, nil
	}

	updated, cmd := d.Content.Update(msg)
	d.Content = updated

	return d, cmd
}

// View renders the centered dialog with border, title bar, and content area.
func (d *Dialog) View() string {
	box := d.renderBox()
	if d.width <= 0 || d.height <= 0 {
		return box
	}

	return lipgloss.Place(d.width, d.height, lipgloss.Center, lipgloss.Center, box)
}

func (d *Dialog) renderBox() string {
	w := d.dialogWidth()
	innerWidth := max(10, w-4)

	title := strings.TrimSpace(d.Title)
	if title == "" {
		title = "Dialog"
	}

	if style.StringWidth(title) > innerWidth {
		title = style.Truncate(title, innerWidth, "…")
	}
	if style.StringWidth(title) < innerWidth {
		title = style.Pad(title, innerWidth)
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Width(innerWidth).
		Background(lipgloss.Color("63")).
		Foreground(lipgloss.Color("230")).
		Render(title)

	contentView := ""
	if d.Content != nil {
		contentView = d.Content.View()
	}

	bodyHeight := max(3, d.dialogHeight()-4)
	body := lipgloss.NewStyle().
		Width(innerWidth).
		Height(bodyHeight).
		Padding(1, 1).
		Render(contentView)

	return lipgloss.NewStyle().
		Width(w).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Render(header + "\n" + body)
}

func (d *Dialog) dialogWidth() int {
	if d.width <= 0 {
		if d.Size == DialogLarge {
			return 96
		}
		return 72
	}

	switch d.Size {
	case DialogLarge:
		return max(60, min(110, d.width-4))
	default:
		return max(50, min(80, d.width-8))
	}
}

func (d *Dialog) dialogHeight() int {
	if d.height <= 0 {
		return 16
	}
	return max(10, min(28, d.height-4))
}

// DialogManager manages a stack of open dialogs.
type DialogManager struct {
	stack  []*Dialog
	width  int
	height int
}

// NewDialogManager creates an empty dialog manager.
func NewDialogManager() *DialogManager {
	return &DialogManager{stack: make([]*Dialog, 0)}
}

// Init initializes the manager.
func (m *DialogManager) Init() tea.Cmd {
	return nil
}

// Update handles stack events and forwards input to the top dialog.
func (m *DialogManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch t := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = t.Width
		m.height = t.Height
		if top := m.Top(); top != nil {
			top.width = t.Width
			top.height = t.Height
		}
		return m, nil

	case ShowDialogMsg:
		if t.Dialog != nil {
			t.Dialog.width = m.width
			t.Dialog.height = m.height
			m.Push(t.Dialog)
			return m, t.Dialog.Init()
		}
		return m, nil

	case CloseDialogMsg:
		m.Pop()
		return m, nil

	case DialogResultMsg:
		m.Pop()
		return m, nil

	case tea.KeyMsg:
		switch t.String() {
		case "esc", "ctrl+c":
			if m.IsOpen() {
				m.Pop()
				return m, nil
			}
		}
	}

	top := m.Top()
	if top == nil {
		return m, nil
	}

	updated, cmd := top.Update(msg)
	if d, ok := updated.(*Dialog); ok {
		m.stack[len(m.stack)-1] = d
	}

	return m, cmd
}

// View renders a dimmed backdrop and the top dialog centered.
func (m *DialogManager) View() string {
	top := m.Top()
	if top == nil {
		return ""
	}

	if m.width <= 0 || m.height <= 0 {
		return top.View()
	}

	top.width = m.width
	top.height = m.height

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		top.renderBox(),
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("8")),
		lipgloss.WithWhitespaceBackground(lipgloss.Color("236")),
	)
}

// Push opens a new dialog on top of the stack.
func (m *DialogManager) Push(d *Dialog) {
	if d == nil {
		return
	}
	m.stack = append(m.stack, d)
}

// Pop closes the top dialog.
func (m *DialogManager) Pop() {
	if len(m.stack) == 0 {
		return
	}
	m.stack = m.stack[:len(m.stack)-1]
}

// Top returns the top-most open dialog.
func (m *DialogManager) Top() *Dialog {
	if len(m.stack) == 0 {
		return nil
	}
	return m.stack[len(m.stack)-1]
}

// IsOpen reports whether at least one dialog is open.
func (m *DialogManager) IsOpen() bool {
	return len(m.stack) > 0
}

// Clear closes all dialogs.
func (m *DialogManager) Clear() {
	m.stack = m.stack[:0]
}

// DialogAlert is a basic content model with a message and a single OK action.
type DialogAlert struct {
	title   string
	message string
	focused bool
}

// NewDialogAlert creates an alert content model.
func NewDialogAlert(title, message string) *DialogAlert {
	return &DialogAlert{title: title, message: message, focused: true}
}

// Init initializes the alert model.
func (a *DialogAlert) Init() tea.Cmd {
	return nil
}

// Update handles alert interactions.
func (a *DialogAlert) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m, ok := msg.(tea.KeyMsg); ok {
		switch m.String() {
		case "enter":
			return a, emitDialogResult(DialogResultMsg{ID: a.title, Cancelled: false})
		case "esc":
			return a, emitDialogResult(DialogResultMsg{ID: a.title, Cancelled: true})
		}
	}
	return a, nil
}

// View renders alert message content.
func (a *DialogAlert) View() string {
	buttonStyle := lipgloss.NewStyle().Padding(0, 1)
	if a.focused {
		buttonStyle = buttonStyle.Foreground(lipgloss.Color("230")).Background(lipgloss.Color("63")).Bold(true)
	} else {
		buttonStyle = buttonStyle.Foreground(lipgloss.Color("250")).Background(lipgloss.Color("238"))
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		a.message,
		"",
		buttonStyle.Render("[ OK ]"),
	)
}

// DialogConfirm is a content model with confirm/cancel actions.
type DialogConfirm struct {
	title    string
	message  string
	selected int
}

// NewDialogConfirm creates a confirmation content model.
func NewDialogConfirm(title, message string) *DialogConfirm {
	return &DialogConfirm{title: title, message: message}
}

// Init initializes the confirm model.
func (c *DialogConfirm) Init() tea.Cmd {
	return nil
}

// Update handles left/right selection and enter confirmation.
func (c *DialogConfirm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m, ok := msg.(tea.KeyMsg); ok {
		switch m.String() {
		case "left", "h":
			c.selected = 0
		case "right", "l":
			c.selected = 1
		case "tab":
			c.selected = (c.selected + 1) % 2
		case "shift+tab":
			c.selected = (c.selected + 1) % 2
		case "enter":
			confirmed := c.selected == 0
			return c, emitDialogResult(DialogResultMsg{ID: c.title, Value: confirmed, Cancelled: !confirmed})
		case "esc":
			return c, emitDialogResult(DialogResultMsg{ID: c.title, Value: false, Cancelled: true})
		}
	}
	return c, nil
}

// View renders confirmation message content.
func (c *DialogConfirm) View() string {
	confirmStyle := lipgloss.NewStyle().Padding(0, 1)
	cancelStyle := lipgloss.NewStyle().Padding(0, 1)

	if c.selected == 0 {
		confirmStyle = confirmStyle.Foreground(lipgloss.Color("230")).Background(lipgloss.Color("34")).Bold(true)
		cancelStyle = cancelStyle.Foreground(lipgloss.Color("250")).Background(lipgloss.Color("238"))
	} else {
		confirmStyle = confirmStyle.Foreground(lipgloss.Color("250")).Background(lipgloss.Color("238"))
		cancelStyle = cancelStyle.Foreground(lipgloss.Color("230")).Background(lipgloss.Color("160")).Bold(true)
	}

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Left,
		confirmStyle.Render("[ Confirm ]"),
		"  ",
		cancelStyle.Render("[ Cancel ]"),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		c.message,
		"",
		buttons,
	)
}

type dialogEmptyModel struct{}

func (m *dialogEmptyModel) Init() tea.Cmd { return nil }

func (m *dialogEmptyModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }

func (m *dialogEmptyModel) View() string { return "" }

func emitDialogResult(result DialogResultMsg) tea.Cmd {
	return func() tea.Msg { return result }
}
