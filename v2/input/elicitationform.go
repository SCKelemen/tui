package input

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ElicitationFieldType identifies field UI behavior.
type ElicitationFieldType string

const (
	FieldTypeText    ElicitationFieldType = "text"
	FieldTypeNumber  ElicitationFieldType = "number"
	FieldTypeBoolean ElicitationFieldType = "boolean"
	FieldTypeSelect  ElicitationFieldType = "select"
	FieldTypeDate    ElicitationFieldType = "date"
)

// ElicitationField describes one dynamic form field.
type ElicitationField struct {
	Name        string
	Label       string
	Description string
	Type        ElicitationFieldType
	Required    bool
	Options     []string
	Default     string
}

// ElicitationSubmitMsg is emitted when the form is submitted.
type ElicitationSubmitMsg struct {
	Values map[string]string
}

// ElicitationFormOption configures an ElicitationForm.
type ElicitationFormOption func(*ElicitationForm)

// WithElicitationFormWidth sets a fixed render width.
func WithElicitationFormWidth(width int) ElicitationFormOption {
	return func(f *ElicitationForm) {
		if width > 0 {
			f.width = width
		}
	}
}

// WithElicitationFormTitle overrides the form title.
func WithElicitationFormTitle(title string) ElicitationFormOption {
	return func(f *ElicitationForm) {
		f.title = strings.TrimSpace(title)
	}
}

// WithElicitationFormDesignTokens applies design-system colors.
func WithElicitationFormDesignTokens(tokens *design.DesignTokens) ElicitationFormOption {
	return func(f *ElicitationForm) {
		if tokens == nil {
			return
		}
		f.colors = elicitationFormColorsFromTokens(tokens)
	}
}

// ElicitationForm renders a dynamic MCP elicitation form.
type ElicitationForm struct {
	fields []ElicitationField
	values map[string]string

	title      string
	width      int
	windowWidth int
	focused    bool
	cursor     int
	errorText  string
	colors     elicitationFormColors
}

type elicitationFormColors struct {
	title     string
	label     string
	value     string
	muted     string
	required  string
	focus     string
	errorText string
}

func defaultElicitationFormColors() elicitationFormColors {
	return elicitationFormColors{
		title:     style.ANSICyan,
		label:     style.ANSIWhite,
		value:     style.ANSIWhite,
		muted:     style.ANSIDim,
		required:  style.ANSIRed,
		focus:     style.ANSIInverse,
		errorText: style.ANSIRed,
	}
}

func elicitationFormColorsFromTokens(tokens *design.DesignTokens) elicitationFormColors {
	c := defaultElicitationFormColors()
	if tokens == nil {
		return c
	}
	if v := style.Fg(tokens.Accent); v != "" {
		c.title = v
	}
	if v := style.Fg(tokens.Color); v != "" {
		c.label = v
		c.value = v
	}
	if v := style.Fg(tokens.MutedColor); v != "" {
		c.muted = v
	}
	if v := style.Fg(tokens.ErrorBright); v != "" {
		c.required = v
		c.errorText = v
	}
	if v := style.Bg(tokens.SurfaceRaised); v != "" {
		c.focus = v
	}
	return c
}

// NewElicitationForm creates a dynamic form component.
func NewElicitationForm(fields []ElicitationField, opts ...ElicitationFormOption) *ElicitationForm {
	f := &ElicitationForm{
		fields:  append([]ElicitationField(nil), fields...),
		values:  map[string]string{},
		title:   "Elicitation Form",
		width:   0,
		focused: false,
		cursor:  0,
		colors:  defaultElicitationFormColors(),
	}
	for _, field := range f.fields {
		if strings.TrimSpace(field.Name) != "" && strings.TrimSpace(field.Default) != "" {
			f.values[field.Name] = field.Default
		}
	}
	for _, opt := range opts {
		opt(f)
	}
	f.clampCursor()
	return f
}

// Init initializes the component.
func (f *ElicitationForm) Init() tea.Cmd {
	return nil
}

// Update handles navigation, value edits, and submit.
func (f *ElicitationForm) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.windowWidth = msg.Width
		return f, nil
	case tea.KeyMsg:
		if !f.focused {
			return f, nil
		}
		if len(f.fields) == 0 {
			if msg.String() == "enter" {
				return f, f.emitSubmit()
			}
			return f, nil
		}

		switch msg.String() {
		case "tab":
			f.cursor = (f.cursor + 1) % len(f.fields)
			f.errorText = ""
			return f, nil
		case "shift+tab":
			f.cursor--
			if f.cursor < 0 {
				f.cursor = len(f.fields) - 1
			}
			f.errorText = ""
			return f, nil
		case "enter":
			if err := f.validateRequired(); err != "" {
				f.errorText = err
				return f, nil
			}
			f.errorText = ""
			return f, f.emitSubmit()
		case "left", "right", "up", "down", " ":
			f.mutateChoice(msg.String())
			f.errorText = ""
			return f, nil
		case "backspace", "ctrl+h":
			f.backspaceCurrent()
			f.errorText = ""
			return f, nil
		}

		if msg.Type == tea.KeyRunes {
			f.appendRunes(msg.Runes)
			f.errorText = ""
		}
	}
	return f, nil
}

// View renders the form title and fields in order.
func (f *ElicitationForm) View() string {
	width := f.renderWidth()
	lines := []string{fitElicitationLine(f.colors.title+style.ANSIBold+f.title+style.ANSIReset, width)}
	if len(f.fields) == 0 {
		lines = append(lines, fitElicitationLine(f.colors.muted+"(no fields)"+style.ANSIReset, width))
		return strings.Join(lines, "\n") + "\n"
	}

	for i, field := range f.fields {
		label := strings.TrimSpace(field.Label)
		if label == "" {
			label = field.Name
		}
		required := ""
		if field.Required {
			required = " " + f.colors.required + "*" + style.ANSIReset
		}
		prefix := "  "
		if i == f.cursor {
			prefix = "› "
		}
		value := f.renderFieldValue(field)
		row := prefix + f.colors.label + label + style.ANSIReset + required + ": " + value
		if i == f.cursor && f.focused {
			row = f.colors.focus + row + style.ANSIReset
		}
		lines = append(lines, fitElicitationLine(row, width))

		desc := strings.TrimSpace(field.Description)
		if desc != "" {
			lines = append(lines, fitElicitationLine("   "+f.colors.muted+desc+style.ANSIReset, width))
		}
	}

	if strings.TrimSpace(f.errorText) != "" {
		lines = append(lines, fitElicitationLine(f.colors.errorText+f.errorText+style.ANSIReset, width))
	}

	hint := f.colors.muted + "tab/shift-tab navigate · enter submit" + style.ANSIReset
	lines = append(lines, fitElicitationLine(hint, width))
	return strings.Join(lines, "\n") + "\n"
}

// Focus marks the component focused.
func (f *ElicitationForm) Focus() {
	f.focused = true
}

// Blur marks the component unfocused.
func (f *ElicitationForm) Blur() {
	f.focused = false
}

// Focused reports focus state.
func (f *ElicitationForm) Focused() bool {
	return f.focused
}

func (f *ElicitationForm) clampCursor() {
	if len(f.fields) == 0 {
		f.cursor = 0
		return
	}
	if f.cursor < 0 {
		f.cursor = 0
	}
	if f.cursor >= len(f.fields) {
		f.cursor = len(f.fields) - 1
	}
}

func (f *ElicitationForm) currentField() *ElicitationField {
	if len(f.fields) == 0 {
		return nil
	}
	f.clampCursor()
	return &f.fields[f.cursor]
}

func (f *ElicitationForm) fieldValue(field ElicitationField) string {
	if v, ok := f.values[field.Name]; ok {
		return v
	}
	return field.Default
}

func (f *ElicitationForm) setFieldValue(field ElicitationField, value string) {
	if strings.TrimSpace(field.Name) == "" {
		return
	}
	f.values[field.Name] = value
}

func (f *ElicitationForm) renderFieldValue(field ElicitationField) string {
	value := f.fieldValue(field)
	switch field.Type {
	case FieldTypeBoolean:
		checked := strings.EqualFold(value, "true")
		if checked {
			return f.colors.value + "[x] true" + style.ANSIReset
		}
		return f.colors.value + "[ ] false" + style.ANSIReset
	case FieldTypeSelect:
		if strings.TrimSpace(value) == "" && len(field.Options) > 0 {
			value = field.Options[0]
		}
		if strings.TrimSpace(value) == "" {
			value = "(none)"
		}
		return f.colors.value + "<" + value + ">" + style.ANSIReset
	default:
		if strings.TrimSpace(value) == "" {
			return f.colors.muted + "(empty)" + style.ANSIReset
		}
		return f.colors.value + value + style.ANSIReset
	}
}

func (f *ElicitationForm) appendRunes(runes []rune) {
	field := f.currentField()
	if field == nil {
		return
	}
	if field.Type == FieldTypeBoolean || field.Type == FieldTypeSelect {
		return
	}
	current := f.fieldValue(*field)
	for _, r := range runes {
		if !isAllowedRune(field.Type, r) {
			continue
		}
		current += string(r)
	}
	f.setFieldValue(*field, current)
}

func (f *ElicitationForm) backspaceCurrent() {
	field := f.currentField()
	if field == nil {
		return
	}
	if field.Type == FieldTypeBoolean || field.Type == FieldTypeSelect {
		return
	}
	current := []rune(f.fieldValue(*field))
	if len(current) == 0 {
		return
	}
	f.setFieldValue(*field, string(current[:len(current)-1]))
}

func (f *ElicitationForm) mutateChoice(key string) {
	field := f.currentField()
	if field == nil {
		return
	}
	switch field.Type {
	case FieldTypeBoolean:
		current := strings.EqualFold(f.fieldValue(*field), "true")
		if key == " " || key == "left" || key == "right" || key == "up" || key == "down" {
			current = !current
		}
		if current {
			f.setFieldValue(*field, "true")
		} else {
			f.setFieldValue(*field, "false")
		}
	case FieldTypeSelect:
		if len(field.Options) == 0 {
			return
		}
		current := f.fieldValue(*field)
		index := 0
		for i, option := range field.Options {
			if option == current {
				index = i
				break
			}
		}
		if key == "left" || key == "up" {
			index--
		} else if key == "right" || key == "down" || key == " " {
			index++
		} else {
			return
		}
		if index < 0 {
			index = len(field.Options) - 1
		}
		if index >= len(field.Options) {
			index = 0
		}
		f.setFieldValue(*field, field.Options[index])
	}
}

func (f *ElicitationForm) validateRequired() string {
	for _, field := range f.fields {
		if !field.Required {
			continue
		}
		value := strings.TrimSpace(f.fieldValue(field))
		if value == "" {
			label := strings.TrimSpace(field.Label)
			if label == "" {
				label = field.Name
			}
			if label == "" {
				label = "required field"
			}
			return label + " is required"
		}
	}
	return ""
}

func (f *ElicitationForm) emitSubmit() tea.Cmd {
	values := map[string]string{}
	for _, field := range f.fields {
		if strings.TrimSpace(field.Name) == "" {
			continue
		}
		values[field.Name] = f.fieldValue(field)
	}
	msg := ElicitationSubmitMsg{Values: values}
	return func() tea.Msg {
		return msg
	}
}

func (f *ElicitationForm) renderWidth() int {
	if f.width > 0 {
		return f.width
	}
	if f.windowWidth > 0 {
		return f.windowWidth
	}
	return 0
}

func fitElicitationLine(s string, width int) string {
	if width <= 0 {
		return s
	}
	t := style.Truncate(s, width, "…")
	if style.StringWidth(t) < width {
		return style.Pad(t, width)
	}
	return t
}

func isAllowedRune(fieldType ElicitationFieldType, r rune) bool {
	if r < 32 {
		return false
	}
	switch fieldType {
	case FieldTypeNumber:
		if (r >= '0' && r <= '9') || r == '.' || r == '-' {
			return true
		}
		return false
	case FieldTypeDate:
		if (r >= '0' && r <= '9') || r == '-' || r == '/' {
			return true
		}
		return false
	default:
		return true
	}
}

var _ tui.Component = (*ElicitationForm)(nil)
