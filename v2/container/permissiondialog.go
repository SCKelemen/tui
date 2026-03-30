package container

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// PermissionRequest describes a tool permission request.
type PermissionRequest struct {
	ToolName    string
	Description string
	Command     string
	Risk        string
}

// PermissionResult is the user's decision from the permission dialog.
type PermissionResult struct {
	Allowed     bool
	AlwaysAllow bool
	Exceptions  []string
}

// PermissionResultMsg is emitted when a permission decision is made.
type PermissionResultMsg struct {
	Result PermissionResult
}

// PermissionDialogOption configures a PermissionDialog.
type PermissionDialogOption func(*PermissionDialog)

// WithPermissionDialogDesignTokens applies design-system tokens to dialog colors.
func WithPermissionDialogDesignTokens(tokens *design.DesignTokens) PermissionDialogOption {
	return func(p *PermissionDialog) {
		if tokens == nil {
			return
		}
		p.colors = permissionDialogColorsFromTokens(tokens)
	}
}

// WithPermissionDialogWidth sets the preferred render width.
func WithPermissionDialogWidth(width int) PermissionDialogOption {
	return func(p *PermissionDialog) {
		if width > 0 {
			p.width = width
		}
	}
}

// WithPermissionDialogExceptions pre-populates exception entries.
func WithPermissionDialogExceptions(exceptions []string) PermissionDialogOption {
	return func(p *PermissionDialog) {
		p.exceptions = append([]string(nil), exceptions...)
	}
}

// PermissionDialog shows tool permission details with allow/deny choices.
type PermissionDialog struct {
	request PermissionRequest
	cursor  int // 0=Allow, 1=Always Allow, 2=Deny
	visible bool
	width   int
	focused bool

	exceptions []string
	colors     permissionDialogColors

	windowWidth int
}

type permissionDialogColors struct {
	warning   string
	accent    string
	text      string
	muted     string
	riskLow   string
	riskMed   string
	riskHigh  string
	separator string
	cursorBg  string
	codeBg    string
	codeText  string
}

func defaultPermissionDialogColors() permissionDialogColors {
	return permissionDialogColors{
		warning:   style.ANSIYellow,
		accent:    style.ANSICyan,
		text:      style.ANSIWhite,
		muted:     style.ANSIDim,
		riskLow:   style.ANSIGreen,
		riskMed:   style.ANSIYellow,
		riskHigh:  style.ANSIRed,
		separator: style.ANSIDim,
		cursorBg:  style.ANSIInverse,
		codeBg:    "\033[48;5;236m",
		codeText:  style.ANSIWhite,
	}
}

func permissionDialogColorsFromTokens(tokens *design.DesignTokens) permissionDialogColors {
	c := defaultPermissionDialogColors()
	if tokens == nil {
		return c
	}
	if v := style.Fg(tokens.PendingColor); v != "" {
		c.warning = v
	}
	if v := style.Fg(tokens.Accent); v != "" {
		c.accent = v
	}
	if v := style.Fg(tokens.Color); v != "" {
		c.text = v
	}
	if v := style.Fg(tokens.MutedColor); v != "" {
		c.muted = v
	}
	if v := style.Fg(tokens.SuccessBright); v != "" {
		c.riskLow = v
	}
	if v := style.Fg(tokens.PendingColor); v != "" {
		c.riskMed = v
	}
	if v := style.Fg(tokens.ErrorBright); v != "" {
		c.riskHigh = v
	}
	if v := style.Fg(tokens.BorderSubtle); v != "" {
		c.separator = v
	}
	if v := style.Bg(tokens.Accent); v != "" {
		c.cursorBg = v
	}
	if v := style.Bg(tokens.SurfaceRaised); v != "" {
		c.codeBg = v
	}
	if v := style.Fg(tokens.Color); v != "" {
		c.codeText = v
	}
	return c
}

// NewPermissionDialog creates a permission approval dialog component.
func NewPermissionDialog(request PermissionRequest, opts ...PermissionDialogOption) *PermissionDialog {
	p := &PermissionDialog{
		request:     request,
		cursor:      0,
		visible:     true,
		width:       76,
		focused:     false,
		exceptions:  nil,
		colors:      defaultPermissionDialogColors(),
		windowWidth: 0,
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.cursor < 0 || p.cursor > 2 {
		p.cursor = 0
	}

	return p
}

// Init initializes the component.
func (p *PermissionDialog) Init() tea.Cmd {
	return nil
}

// Update handles keyboard navigation and selection.
func (p *PermissionDialog) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.windowWidth = msg.Width
		return p, nil
	case tea.KeyMsg:
		if !p.focused || !p.visible {
			return p, nil
		}

		switch msg.String() {
		case "down", "j":
			if p.cursor < 2 {
				p.cursor++
			}
			return p, nil
		case "up", "k":
			if p.cursor > 0 {
				p.cursor--
			}
			return p, nil
		case "esc":
			p.cursor = 2
			return p, p.emitResult(false, false)
		case "enter":
			switch p.cursor {
			case 0:
				return p, p.emitResult(true, false)
			case 1:
				return p, p.emitResult(true, true)
			default:
				return p, p.emitResult(false, false)
			}
		}
	}

	return p, nil
}

// View renders the permission dialog.
func (p *PermissionDialog) View() string {
	if !p.visible {
		return ""
	}

	width := p.width
	if width <= 0 {
		width = 76
	}
	if p.windowWidth > 0 {
		maxWidth := p.windowWidth - 4
		if maxWidth < 40 {
			maxWidth = 40
		}
		if width > maxWidth {
			width = maxWidth
		}
	}

	inner := width
	if inner < 40 {
		inner = 40
	}

	lines := make([]string, 0, 24)

	title := p.colors.warning + style.ANSIBold + "⚠ Permission Required" + style.ANSIReset
	lines = append(lines, p.fitWidth(title, inner))

	toolName := strings.TrimSpace(p.request.ToolName)
	if toolName == "" {
		toolName = "(unknown tool)"
	}
	toolLine := style.ANSIBold + "Tool: " + p.colors.accent + style.ANSIBold + toolName + style.ANSIReset
	lines = append(lines, p.fitWidth(toolLine, inner))

	desc := strings.TrimSpace(p.request.Description)
	if desc == "" {
		desc = "No description provided"
	}
	lines = append(lines, p.wrapText(desc, inner)...)

	cmd := strings.TrimSpace(p.request.Command)
	if cmd == "" {
		cmd = "(no command preview)"
	}
	cmdPrefix := p.colors.muted + "Command:" + style.ANSIReset
	lines = append(lines, p.fitWidth(cmdPrefix, inner))
	for _, row := range p.wrapPlainText(cmd, inner-2) {
		line := p.colors.codeBg + p.colors.codeText + " " + style.Truncate(row, inner-2, "…") + style.Pad("", max(0, inner-2-style.StringWidth(row))) + " " + style.ANSIReset
		lines = append(lines, p.fitWidth(line, inner))
	}

	if risk := strings.TrimSpace(p.request.Risk); risk != "" {
		badge := p.renderRiskBadge(risk)
		lines = append(lines, p.fitWidth("Risk: "+badge, inner))
	}

	lines = append(lines, p.fitWidth(p.colors.separator+strings.Repeat("─", inner)+style.ANSIReset, inner))

	options := []string{"Allow once", "Always allow", "Deny"}
	for i, option := range options {
		marker := "○"
		if i == p.cursor {
			marker = "●"
		}
		line := marker + " " + option
		if i == p.cursor {
			line = p.colors.cursorBg + p.colors.accent + style.ANSIBold + line + style.ANSIReset
		} else {
			line = p.colors.text + line + style.ANSIReset
		}
		lines = append(lines, p.fitWidth(line, inner))
	}

	if len(p.exceptions) > 0 {
		lines = append(lines, "")
		lines = append(lines, p.fitWidth("- with exceptions to:", inner))
		for _, ex := range p.exceptions {
			exText := strings.TrimSpace(ex)
			if exText == "" {
				continue
			}
			for _, row := range p.wrapText("  • "+exText, inner) {
				lines = append(lines, row)
			}
		}
	}

	lines = append(lines, "")
	footer := p.colors.muted + "Enter to confirm  •  Esc to deny" + style.ANSIReset
	lines = append(lines, p.fitWidth(footer, inner))

	return strings.Join(lines, "\n")
}

// Focus marks the component as focused.
func (p *PermissionDialog) Focus() {
	p.focused = true
}

// Blur marks the component as unfocused.
func (p *PermissionDialog) Blur() {
	p.focused = false
}

// Focused reports whether the component currently has focus.
func (p *PermissionDialog) Focused() bool {
	return p.focused
}

// Show makes the dialog visible.
func (p *PermissionDialog) Show() {
	p.visible = true
}

// Hide conceals the dialog.
func (p *PermissionDialog) Hide() {
	p.visible = false
}

// IsVisible reports whether the dialog is currently visible.
func (p *PermissionDialog) IsVisible() bool {
	return p.visible
}

func (p *PermissionDialog) emitResult(allowed, always bool) tea.Cmd {
	result := PermissionResult{
		Allowed:     allowed,
		AlwaysAllow: always,
		Exceptions:  append([]string(nil), p.exceptions...),
	}
	return func() tea.Msg {
		return PermissionResultMsg{Result: result}
	}
}

func (p *PermissionDialog) renderRiskBadge(risk string) string {
	norm := strings.ToLower(strings.TrimSpace(risk))
	label := strings.ToUpper(norm)
	if label == "" {
		label = "UNKNOWN"
	}

	color := p.colors.muted
	switch norm {
	case "low":
		color = p.colors.riskLow
	case "medium":
		color = p.colors.riskMed
	case "high":
		color = p.colors.riskHigh
	}

	return color + style.ANSIBold + "[" + label + "]" + style.ANSIReset
}

func (p *PermissionDialog) fitWidth(line string, width int) string {
	if width <= 0 {
		return line
	}
	if style.StringWidth(line) <= width {
		return style.Pad(line, width)
	}
	return style.Truncate(line, width, "…")
}

func (p *PermissionDialog) wrapText(text string, width int) []string {
	rows := p.wrapPlainText(text, width)
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, p.fitWidth(p.colors.text+row+style.ANSIReset, width))
	}
	return out
}

func (p *PermissionDialog) wrapPlainText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return []string{""}
	}

	rows := make([]string, 0, len(words))
	line := words[0]
	for _, word := range words[1:] {
		candidate := line + " " + word
		if style.StringWidth(candidate) <= width {
			line = candidate
			continue
		}
		rows = append(rows, line)
		line = word
	}
	rows = append(rows, line)
	return rows
}

var _ tui.Component = (*PermissionDialog)(nil)
