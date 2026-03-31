package input

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ModelProvider groups models under one provider label.
type ModelProvider struct {
	Name   string
	Models []ModelInfo
}

// ModelInfo describes one LLM model option.
type ModelInfo struct {
	ID            string
	Name          string
	Description   string
	ContextWindow int
}

// ModelSelectedMsg is emitted when a model is selected.
type ModelSelectedMsg struct {
	Provider string
	Model    ModelInfo
}

// ModelPickerOption configures a ModelPicker.
type ModelPickerOption func(*ModelPicker)

// WithModelPickerWidth sets the preferred render width.
func WithModelPickerWidth(width int) ModelPickerOption {
	return func(m *ModelPicker) {
		if width > 0 {
			m.width = width
		}
	}
}

// WithModelPickerSelected sets the initial selected model ID.
func WithModelPickerSelected(selected string) ModelPickerOption {
	return func(m *ModelPicker) {
		m.selectedID = strings.TrimSpace(selected)
	}
}

// WithModelPickerDesignTokens applies design-system tokens.
func WithModelPickerDesignTokens(tokens *design.DesignTokens) ModelPickerOption {
	return func(m *ModelPicker) {
		m.designTokens = tokens
		m.applyDesignTokens()
	}
}

type modelPickerRow struct {
	Provider string
	Model    ModelInfo
}

// ModelPicker renders grouped model options with keyboard navigation and filtering.
type ModelPicker struct {
	providers  []ModelProvider
	rows       []modelPickerRow
	cursor     int
	query      string
	selectedID string
	focused    bool
	width      int

	designTokens *design.DesignTokens
	accentColor  string
	mutedColor   string
	textColor    string
}

// NewModelPicker creates a new ModelPicker.
func NewModelPicker(providers []ModelProvider, opts ...ModelPickerOption) *ModelPicker {
	m := &ModelPicker{
		providers:    append([]ModelProvider(nil), providers...),
		rows:         make([]modelPickerRow, 0),
		cursor:       0,
		query:        "",
		selectedID:   "",
		focused:      false,
		width:        72,
		designTokens: design.DefaultTheme(),
		accentColor:  style.ANSICyan,
		mutedColor:   style.ANSIDim,
		textColor:    style.ANSIWhite,
	}

	for _, opt := range opts {
		opt(m)
	}

	m.applyDesignTokens()
	m.refreshRows()
	m.syncCursorToSelected()

	return m
}

// Init initializes the component.
func (m *ModelPicker) Init() tea.Cmd { return nil }

// Update handles keyboard and window messages.
func (m *ModelPicker) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.width <= 0 {
			m.width = msg.Width
		}
		return m, nil

	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.rows)-1 {
				m.cursor++
			}
			return m, nil
		case "enter":
			if len(m.rows) == 0 {
				return m, nil
			}
			selected := m.rows[m.cursor]
			m.selectedID = selected.Model.ID
			return m, func() tea.Msg {
				return ModelSelectedMsg{Provider: selected.Provider, Model: selected.Model}
			}
		case "backspace":
			if m.query != "" {
				r := []rune(m.query)
				m.query = string(r[:len(r)-1])
				m.refreshRows()
			}
			return m, nil
		case "esc":
			if m.query != "" {
				m.query = ""
				m.refreshRows()
			}
			return m, nil
		default:
			if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
				m.query += string(msg.Runes)
				m.refreshRows()
			}
			return m, nil
		}
	}

	return m, nil
}

// View renders the provider-grouped model list.
func (m *ModelPicker) View() string {
	if len(m.providers) == 0 {
		return ""
	}

	w := m.width
	if w <= 0 {
		w = 72
	}

	lines := []string{fmt.Sprintf("%sFilter:%s %s", m.mutedColor, style.ANSIReset, m.query)}
	if strings.TrimSpace(m.query) == "" {
		lines[0] = fmt.Sprintf("%sFilter:%s %s(type to filter)", m.mutedColor, style.ANSIReset, m.mutedColor)
	}

	rowIndex := 0
	for _, provider := range m.providers {
		providerRows := m.rowsForProvider(provider.Name)
		if len(providerRows) == 0 {
			continue
		}

		header := style.ANSIBold + provider.Name + style.ANSIReset
		lines = append(lines, header)

		for _, row := range providerRows {
			prefix := "  "
			if rowIndex == m.cursor {
				prefix = m.accentColor + "▸ " + style.ANSIReset
			}

			name := strings.TrimSpace(row.Model.Name)
			if name == "" {
				name = row.Model.ID
			}
			ctx := fmt.Sprintf("(%d)", row.Model.ContextWindow)
			line := fmt.Sprintf("%s%s %s", prefix, name, m.mutedColor+ctx+style.ANSIReset)
			if row.Model.ID == m.selectedID {
				line += " " + m.accentColor + "✓" + style.ANSIReset
			}
			lines = append(lines, line)
			rowIndex++
		}
	}

	for i := range lines {
		lines[i] = style.Pad(style.Truncate(lines[i], w, "…"), w)
	}

	return strings.Join(lines, "\n")
}

// Focus marks the component focused.
func (m *ModelPicker) Focus() { m.focused = true }

// Blur marks the component unfocused.
func (m *ModelPicker) Blur() { m.focused = false }

// Focused reports focus state.
func (m *ModelPicker) Focused() bool { return m.focused }

func (m *ModelPicker) refreshRows() {
	filtered := make([]modelPickerRow, 0)
	query := strings.ToLower(strings.TrimSpace(m.query))

	for _, provider := range m.providers {
		for _, model := range provider.Models {
			candidate := strings.ToLower(model.Name + " " + model.ID + " " + model.Description)
			if query != "" && !strings.Contains(candidate, query) {
				continue
			}
			filtered = append(filtered, modelPickerRow{Provider: provider.Name, Model: model})
		}
	}

	m.rows = filtered
	if len(m.rows) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(m.rows) {
		m.cursor = len(m.rows) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *ModelPicker) rowsForProvider(provider string) []modelPickerRow {
	out := make([]modelPickerRow, 0)
	for _, r := range m.rows {
		if r.Provider == provider {
			out = append(out, r)
		}
	}
	return out
}

func (m *ModelPicker) syncCursorToSelected() {
	if m.selectedID == "" {
		return
	}
	for i, row := range m.rows {
		if row.Model.ID == m.selectedID {
			m.cursor = i
			return
		}
	}
}

func (m *ModelPicker) applyDesignTokens() {
	if m.designTokens == nil {
		return
	}
	if v := style.Fg(m.designTokens.Accent); v != "" {
		m.accentColor = v
	}
	if v := style.Fg(m.designTokens.MutedColor); v != "" {
		m.mutedColor = v
	}
	if v := style.Fg(m.designTokens.Color); v != "" {
		m.textColor = v
	}
}

var _ tui.Component = (*ModelPicker)(nil)
