package input

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ThemeOption is one theme choice.
type ThemeOption struct {
	Name        string
	Description string
	Colors      ThemeColors
}

// ThemeColors defines the key semantic colors for a theme.
type ThemeColors struct {
	Primary    string
	Secondary  string
	Background string
	Surface    string
	Text       string
	Accent     string
}

// ThemeSelectedMsg is emitted when a theme is selected.
type ThemeSelectedMsg struct {
	Theme ThemeOption
}

// ThemePickerOption configures a ThemePicker.
type ThemePickerOption func(*ThemePicker)

// WithThemePickerWidth sets the preferred render width.
func WithThemePickerWidth(width int) ThemePickerOption {
	return func(t *ThemePicker) {
		if width > 0 {
			t.width = width
		}
	}
}

// WithThemePickerSelected sets the initial selected theme by name.
func WithThemePickerSelected(selected string) ThemePickerOption {
	return func(t *ThemePicker) {
		t.selected = strings.TrimSpace(selected)
	}
}

// WithThemePickerDesignTokens applies design-system tokens.
func WithThemePickerDesignTokens(tokens *design.DesignTokens) ThemePickerOption {
	return func(t *ThemePicker) {
		t.designTokens = tokens
		t.applyDesignTokens()
	}
}

// ThemePicker renders a theme list with a live preview panel.
type ThemePicker struct {
	themes   []ThemeOption
	cursor   int
	selected string
	focused  bool
	width    int

	designTokens *design.DesignTokens
	accentColor  string
	mutedColor   string
}

// NewThemePicker creates a new ThemePicker.
func NewThemePicker(themes []ThemeOption, opts ...ThemePickerOption) *ThemePicker {
	t := &ThemePicker{
		themes:       append([]ThemeOption(nil), themes...),
		cursor:       0,
		selected:     "",
		focused:      false,
		width:        96,
		designTokens: design.DefaultTheme(),
		accentColor:  style.ANSICyan,
		mutedColor:   style.ANSIDim,
	}

	for _, opt := range opts {
		opt(t)
	}

	t.applyDesignTokens()
	t.syncCursorToSelected()

	return t
}

// Init initializes the component.
func (t *ThemePicker) Init() tea.Cmd { return nil }

// Update handles keyboard and window messages.
func (t *ThemePicker) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if t.width <= 0 {
			t.width = msg.Width
		}
		return t, nil

	case tea.KeyMsg:
		if !t.focused || len(t.themes) == 0 {
			return t, nil
		}

		switch msg.String() {
		case "up", "k":
			if t.cursor > 0 {
				t.cursor--
			}
			return t, nil
		case "down", "j":
			if t.cursor < len(t.themes)-1 {
				t.cursor++
			}
			return t, nil
		case "enter":
			chosen := t.themes[t.cursor]
			t.selected = chosen.Name
			return t, func() tea.Msg {
				return ThemeSelectedMsg{Theme: chosen}
			}
		}
	}

	return t, nil
}

// View renders the list and preview pane.
func (t *ThemePicker) View() string {
	if len(t.themes) == 0 {
		return ""
	}

	width := t.width
	if width <= 0 {
		width = 96
	}
	if width < 48 {
		width = 48
	}

	leftWidth := width / 2
	rightWidth := width - leftWidth - 1

	leftLines := make([]string, 0, len(t.themes)+1)
	leftLines = append(leftLines, style.ANSIBold+"Themes"+style.ANSIReset)

	for i, theme := range t.themes {
		cursor := "  "
		if i == t.cursor {
			cursor = t.accentColor + "▸ " + style.ANSIReset
		}

		line := cursor + theme.Name + " " + t.renderSwatches(theme.Colors)
		if theme.Name == t.selected {
			line += " " + t.accentColor + "✓" + style.ANSIReset
		}
		leftLines = append(leftLines, style.Pad(style.Truncate(line, leftWidth, "…"), leftWidth))
	}

	preview := t.renderPreview(t.themes[t.cursor], rightWidth)
	previewLines := strings.Split(preview, "\n")

	maxLines := len(leftLines)
	if len(previewLines) > maxLines {
		maxLines = len(previewLines)
	}

	for len(leftLines) < maxLines {
		leftLines = append(leftLines, strings.Repeat(" ", leftWidth))
	}
	for len(previewLines) < maxLines {
		previewLines = append(previewLines, strings.Repeat(" ", rightWidth))
	}

	out := make([]string, 0, maxLines)
	for i := 0; i < maxLines; i++ {
		out = append(out, leftLines[i]+" "+previewLines[i])
	}

	return strings.Join(out, "\n")
}

// Focus marks the component focused.
func (t *ThemePicker) Focus() { t.focused = true }

// Blur marks the component unfocused.
func (t *ThemePicker) Blur() { t.focused = false }

// Focused reports focus state.
func (t *ThemePicker) Focused() bool { return t.focused }

func (t *ThemePicker) renderSwatches(colors ThemeColors) string {
	sw := []string{colors.Primary, colors.Secondary, colors.Background, colors.Surface, colors.Text, colors.Accent}
	parts := make([]string, 0, len(sw))
	for _, c := range sw {
		bg := style.ANSIBackgroundColorFromHex(c)
		if bg == "" {
			parts = append(parts, "··")
			continue
		}
		parts = append(parts, bg+"  "+style.ANSIReset)
	}
	return strings.Join(parts, "")
}

func (t *ThemePicker) renderPreview(theme ThemeOption, width int) string {
	if width < 18 {
		width = 18
	}

	title := style.Pad(style.Truncate("Preview: "+theme.Name, width, "…"), width)
	line1 := t.previewLine("Primary", theme.Colors.Primary, width)
	line2 := t.previewLine("Secondary", theme.Colors.Secondary, width)
	line3 := t.previewLine("Accent", theme.Colors.Accent, width)
	line4 := t.previewLine("Text", theme.Colors.Text, width)
	desc := style.Pad(style.Truncate(strings.TrimSpace(theme.Description), width, "…"), width)

	return strings.Join([]string{title, line1, line2, line3, line4, desc}, "\n")
}

func (t *ThemePicker) previewLine(label, hex string, width int) string {
	fg := style.ANSIColorFromHex(hex)
	content := fmt.Sprintf("%s: sample", label)
	if fg == "" {
		return style.Pad(style.Truncate(content, width, "…"), width)
	}
	line := fg + content + style.ANSIReset
	return style.Pad(style.Truncate(line, width, "…"), width)
}

func (t *ThemePicker) syncCursorToSelected() {
	if t.selected == "" {
		return
	}
	for i, theme := range t.themes {
		if strings.EqualFold(theme.Name, t.selected) {
			t.cursor = i
			return
		}
	}
}

func (t *ThemePicker) applyDesignTokens() {
	if t.designTokens == nil {
		return
	}
	if v := style.Fg(t.designTokens.Accent); v != "" {
		t.accentColor = v
	}
	if v := style.Fg(t.designTokens.MutedColor); v != "" {
		t.mutedColor = v
	}
}

var _ tui.Component = (*ThemePicker)(nil)
