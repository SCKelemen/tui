package container

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// FilePermissionResultMsg is emitted when a file operation is accepted or rejected.
type FilePermissionResultMsg struct {
	Accepted  bool
	Filepath  string
	Operation string
}

// FilePermissionViewOption configures a FilePermissionView.
type FilePermissionViewOption func(*FilePermissionView)

// WithFilePermissionViewWidth sets a fixed render width.
func WithFilePermissionViewWidth(width int) FilePermissionViewOption {
	return func(v *FilePermissionView) {
		if width > 0 {
			v.width = width
		}
	}
}

// WithFilePermissionViewDiffContent sets unified diff content for preview.
func WithFilePermissionViewDiffContent(content string) FilePermissionViewOption {
	return func(v *FilePermissionView) {
		v.diffContent = content
	}
}

// WithFilePermissionViewNewContent sets new file content when no diff is available.
func WithFilePermissionViewNewContent(content string) FilePermissionViewOption {
	return func(v *FilePermissionView) {
		v.newContent = content
	}
}

// WithFilePermissionViewTruncate sets maximum preview lines.
func WithFilePermissionViewTruncate(lines int) FilePermissionViewOption {
	return func(v *FilePermissionView) {
		if lines >= 0 {
			v.truncate = lines
		}
	}
}

// WithFilePermissionViewDesignTokens applies design-system colors.
func WithFilePermissionViewDesignTokens(tokens *design.DesignTokens) FilePermissionViewOption {
	return func(v *FilePermissionView) {
		if tokens == nil {
			return
		}
		v.colors = filePermissionViewColorsFromTokens(tokens)
	}
}

// FilePermissionView renders a file permission prompt and diff preview.
type FilePermissionView struct {
	filepath  string
	operation string

	diffContent string
	newContent  string
	truncate    int

	width       int
	windowWidth int
	focused     bool
	colors      filePermissionViewColors
}

type filePermissionViewColors struct {
	text      string
	muted     string
	accent    string
	allow     string
	reject    string
	readOp    string
	writeOp   string
	editOp    string
	codePlus  string
	codeMinus string
	codeDim   string
}

func defaultFilePermissionViewColors() filePermissionViewColors {
	return filePermissionViewColors{
		text:      style.ANSIWhite,
		muted:     style.ANSIDim,
		accent:    style.ANSICyan,
		allow:     style.ANSIGreen,
		reject:    style.ANSIRed,
		readOp:    style.ANSICyan,
		writeOp:   style.ANSIYellow,
		editOp:    style.ANSICyan,
		codePlus:  style.ANSIGreen,
		codeMinus: style.ANSIRed,
		codeDim:   style.ANSIDim,
	}
}

func filePermissionViewColorsFromTokens(tokens *design.DesignTokens) filePermissionViewColors {
	c := defaultFilePermissionViewColors()
	if tokens == nil {
		return c
	}
	if v := style.Fg(tokens.Color); v != "" {
		c.text = v
	}
	if v := style.Fg(tokens.MutedColor); v != "" {
		c.muted = v
		c.codeDim = v
	}
	if v := style.Fg(tokens.Accent); v != "" {
		c.accent = v
		c.readOp = v
	}
	if v := style.Fg(tokens.SuccessBright); v != "" {
		c.allow = v
		c.codePlus = v
	}
	if v := style.Fg(tokens.ErrorBright); v != "" {
		c.reject = v
		c.codeMinus = v
	}
	if v := style.Fg(tokens.PendingColor); v != "" {
		c.writeOp = v
	}
	if v := style.Fg(tokens.Accent); v != "" {
		c.editOp = v
	}
	return c
}

// NewFilePermissionView creates a file permission component.
func NewFilePermissionView(filepath string, operation string, opts ...FilePermissionViewOption) *FilePermissionView {
	v := &FilePermissionView{
		filepath:   strings.TrimSpace(filepath),
		operation:  strings.ToLower(strings.TrimSpace(operation)),
		truncate:   12,
		width:      92,
		focused:    false,
		colors:     defaultFilePermissionViewColors(),
		newContent: "",
	}

	for _, opt := range opts {
		opt(v)
	}

	if v.filepath == "" {
		v.filepath = "(unknown file)"
	}
	if v.operation == "" {
		v.operation = "read"
	}

	return v
}

// Init initializes the component.
func (v *FilePermissionView) Init() tea.Cmd {
	return nil
}

// Update handles accept/reject key actions.
func (v *FilePermissionView) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.windowWidth = msg.Width
		return v, nil
	case tea.KeyMsg:
		if !v.focused {
			return v, nil
		}
		switch msg.String() {
		case "a":
			return v, v.emitDecision(true)
		case "r", "esc":
			return v, v.emitDecision(false)
		}
	}
	return v, nil
}

// View renders file details, operation badge, and preview.
func (v *FilePermissionView) View() string {
	width := v.renderWidth()
	if width < 50 {
		width = 50
	}

	lines := make([]string, 0, 24)
	header := style.ANSIBold + "File Permission" + style.ANSIReset
	lines = append(lines, v.fit(header, width))

	pathLine := v.colors.text + style.ElidePath(v.filepath, filePermissionMax(10, width)) + style.ANSIReset
	lines = append(lines, v.fit(pathLine, width))

	opBadge := v.operationBadge()
	lines = append(lines, v.fit("operation: "+opBadge, width))
	lines = append(lines, v.fit(v.colors.muted+strings.Repeat("─", width)+style.ANSIReset, width))

	preview := v.previewLines()
	for _, line := range preview {
		lines = append(lines, v.fit(line, width))
	}

	hint := v.colors.allow + "[a]" + style.ANSIReset + "ccept  " + v.colors.reject + "[r]" + style.ANSIReset + "eject"
	lines = append(lines, "")
	lines = append(lines, v.fit(hint, width))

	return strings.Join(lines, "\n") + "\n"
}

// Focus marks this component as focused.
func (v *FilePermissionView) Focus() {
	v.focused = true
}

// Blur marks this component as unfocused.
func (v *FilePermissionView) Blur() {
	v.focused = false
}

// Focused reports focus state.
func (v *FilePermissionView) Focused() bool {
	return v.focused
}

func (v *FilePermissionView) emitDecision(accepted bool) tea.Cmd {
	msg := FilePermissionResultMsg{
		Accepted:  accepted,
		Filepath:  v.filepath,
		Operation: v.operation,
	}
	return func() tea.Msg {
		return msg
	}
}

func (v *FilePermissionView) renderWidth() int {
	width := v.width
	if width <= 0 {
		width = 92
	}
	if v.windowWidth > 0 {
		maxWidth := v.windowWidth - 4
		if maxWidth < 50 {
			maxWidth = 50
		}
		if width > maxWidth {
			width = maxWidth
		}
	}
	return width
}

func (v *FilePermissionView) fit(s string, width int) string {
	if width <= 0 {
		return s
	}
	t := style.Truncate(s, width, "…")
	if style.StringWidth(t) < width {
		return style.Pad(t, width)
	}
	return t
}

func (v *FilePermissionView) operationBadge() string {
	label := strings.ToUpper(v.operation)
	color := v.colors.readOp
	switch v.operation {
	case "write":
		color = v.colors.writeOp
	case "edit":
		color = v.colors.editOp
	case "read":
		color = v.colors.readOp
	default:
		color = v.colors.accent
	}
	return color + style.ANSIBold + "[" + label + "]" + style.ANSIReset
}

func (v *FilePermissionView) previewLines() []string {
	raw := strings.TrimSpace(v.diffContent)
	if raw == "" {
		raw = strings.TrimSpace(v.newContent)
	}
	if raw == "" {
		return []string{v.colors.muted + "(no preview available)" + style.ANSIReset}
	}

	lines := strings.Split(raw, "\n")
	if v.truncate > 0 && len(lines) > v.truncate {
		lines = append(lines[:v.truncate], "...")
	}

	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimRight(line, "\t ")
		color := v.colors.text
		switch {
		case strings.HasPrefix(trimmed, "+"):
			color = v.colors.codePlus
		case strings.HasPrefix(trimmed, "-"):
			color = v.colors.codeMinus
		case strings.HasPrefix(trimmed, "@@"), strings.HasPrefix(trimmed, "diff "), strings.HasPrefix(trimmed, "---"), strings.HasPrefix(trimmed, "+++"):
			color = v.colors.codeDim
		}
		out = append(out, color+trimmed+style.ANSIReset)
	}
	return out
}

func filePermissionMax(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

var _ tui.Component = (*FilePermissionView)(nil)
