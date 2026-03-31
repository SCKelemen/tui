package container

import (
	"fmt"
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// BashPermissionResult is emitted when the user decides on shell command approval.
type BashPermissionResult struct {
	Accepted bool
	Rejected bool
	Edited   bool
	Command  string
	Cwd      string
	Timeout  time.Duration
	Risky    bool
}

// BashPermissionViewOption configures a BashPermissionView.
type BashPermissionViewOption func(*BashPermissionView)

// WithBashPermissionViewWidth sets a fixed render width.
func WithBashPermissionViewWidth(width int) BashPermissionViewOption {
	return func(v *BashPermissionView) {
		if width > 0 {
			v.width = width
		}
	}
}

// WithBashPermissionViewCwd sets the shell working directory shown in the prompt.
func WithBashPermissionViewCwd(cwd string) BashPermissionViewOption {
	return func(v *BashPermissionView) {
		v.cwd = strings.TrimSpace(cwd)
	}
}

// WithBashPermissionViewTimeout sets the command timeout shown in the prompt.
func WithBashPermissionViewTimeout(timeout time.Duration) BashPermissionViewOption {
	return func(v *BashPermissionView) {
		if timeout < 0 {
			timeout = 0
		}
		v.timeout = timeout
	}
}

// WithBashPermissionViewRisky toggles the risky-command indicator.
func WithBashPermissionViewRisky(risky bool) BashPermissionViewOption {
	return func(v *BashPermissionView) {
		v.risky = risky
	}
}

// WithBashPermissionViewDesignTokens applies design-system colors.
func WithBashPermissionViewDesignTokens(tokens *design.DesignTokens) BashPermissionViewOption {
	return func(v *BashPermissionView) {
		if tokens == nil {
			return
		}
		v.colors = bashPermissionViewColorsFromTokens(tokens)
	}
}

// BashPermissionView renders shell command approval details and keybind hints.
type BashPermissionView struct {
	command string
	cwd     string
	timeout time.Duration
	risky   bool

	width       int
	windowWidth int
	focused     bool
	colors      bashPermissionViewColors
}

type bashPermissionViewColors struct {
	title   string
	accent  string
	text    string
	muted   string
	warning string
	codeBg  string
	codeFg  string
}

func defaultBashPermissionViewColors() bashPermissionViewColors {
	return bashPermissionViewColors{
		title:   style.ANSICyan,
		accent:  style.ANSICyan,
		text:    style.ANSIWhite,
		muted:   style.ANSIDim,
		warning: style.ANSIYellow,
		codeBg:  "\033[48;5;236m",
		codeFg:  style.ANSIWhite,
	}
}

func bashPermissionViewColorsFromTokens(tokens *design.DesignTokens) bashPermissionViewColors {
	c := defaultBashPermissionViewColors()
	if tokens == nil {
		return c
	}
	if v := style.Fg(tokens.Accent); v != "" {
		c.title = v
		c.accent = v
	}
	if v := style.Fg(tokens.Color); v != "" {
		c.text = v
		c.codeFg = v
	}
	if v := style.Fg(tokens.MutedColor); v != "" {
		c.muted = v
	}
	if v := style.Fg(tokens.PendingColor); v != "" {
		c.warning = v
	}
	if v := style.Bg(tokens.SurfaceRaised); v != "" {
		c.codeBg = v
	}
	return c
}

// NewBashPermissionView creates a shell command approval component.
func NewBashPermissionView(command string, opts ...BashPermissionViewOption) *BashPermissionView {
	v := &BashPermissionView{
		command: strings.TrimSpace(command),
		width:   90,
		focused: false,
		colors:  defaultBashPermissionViewColors(),
	}

	for _, opt := range opts {
		opt(v)
	}

	if v.command == "" {
		v.command = "(empty command)"
	}

	return v
}

// Init initializes the component.
func (v *BashPermissionView) Init() tea.Cmd {
	return nil
}

// Update handles key actions: [a]ccept, [r]eject, [e]dit.
func (v *BashPermissionView) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
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
			return v, v.emitResult("accept")
		case "r", "esc":
			return v, v.emitResult("reject")
		case "e":
			return v, v.emitResult("edit")
		}
	}

	return v, nil
}

// View renders the approval UI.
func (v *BashPermissionView) View() string {
	w := v.renderWidth()
	inner := w
	if inner < 48 {
		inner = 48
	}

	lines := make([]string, 0, 18)
	lines = append(lines, v.fit(v.colors.title+style.ANSIBold+"Shell Command Approval"+style.ANSIReset, inner))

	if v.risky {
		warn := v.colors.warning + style.ANSIBold + "⚠ Risky command" + style.ANSIReset
		lines = append(lines, v.fit(warn, inner))
	}

	if strings.TrimSpace(v.cwd) != "" {
		cwdText := "cwd: " + style.ElidePath(v.cwd, bashPermissionMax(10, inner-5))
		lines = append(lines, v.fit(v.colors.muted+cwdText+style.ANSIReset, inner))
	}
	if v.timeout > 0 {
		lines = append(lines, v.fit(v.colors.muted+fmt.Sprintf("timeout: %s", v.timeout)+style.ANSIReset, inner))
	}

	lines = append(lines, v.fit(v.colors.muted+"command"+style.ANSIReset, inner))
	for _, row := range v.wrapPlain(v.command, bashPermissionMax(8, inner-4)) {
		body := " " + style.Pad(style.Truncate(row, inner-2, "…"), inner-2) + " "
		lines = append(lines, v.colors.codeBg+v.colors.codeFg+body+style.ANSIReset)
	}

	hints := v.colors.accent + "[a]" + style.ANSIReset + "ccept  " +
		v.colors.accent + "[r]" + style.ANSIReset + "eject  " +
		v.colors.accent + "[e]" + style.ANSIReset + "dit"
	lines = append(lines, "")
	lines = append(lines, v.fit(hints, inner))

	return strings.Join(lines, "\n") + "\n"
}

// Focus marks this component focused.
func (v *BashPermissionView) Focus() {
	v.focused = true
}

// Blur marks this component unfocused.
func (v *BashPermissionView) Blur() {
	v.focused = false
}

// Focused reports focus state.
func (v *BashPermissionView) Focused() bool {
	return v.focused
}

func (v *BashPermissionView) emitResult(action string) tea.Cmd {
	result := BashPermissionResult{
		Accepted: action == "accept",
		Rejected: action == "reject",
		Edited:   action == "edit",
		Command:  v.command,
		Cwd:      v.cwd,
		Timeout:  v.timeout,
		Risky:    v.risky,
	}
	return func() tea.Msg {
		return result
	}
}

func (v *BashPermissionView) renderWidth() int {
	w := v.width
	if w <= 0 {
		w = 90
	}
	if v.windowWidth > 0 {
		maxWidth := v.windowWidth - 4
		if maxWidth < 48 {
			maxWidth = 48
		}
		if w > maxWidth {
			w = maxWidth
		}
	}
	return w
}

func (v *BashPermissionView) fit(s string, width int) string {
	if width <= 0 {
		return s
	}
	plain := style.Truncate(s, width, "…")
	if style.StringWidth(plain) < width {
		return style.Pad(plain, width)
	}
	return plain
}

func (v *BashPermissionView) wrapPlain(s string, width int) []string {
	if width <= 0 {
		return []string{s}
	}
	words := strings.Fields(strings.TrimSpace(s))
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

func bashPermissionMax(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

var _ tui.Component = (*BashPermissionView)(nil)
