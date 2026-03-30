package display

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ShellOutputView displays a terminal command, its output, and exit status.
type ShellOutputView struct {
	command string
	output  string

	exitCode  int
	cwd       string
	truncated bool
	maxLines  int
	width     int

	focused     bool
	windowWidth int

	// Inline colors
	successColor string
	errorColor   string
	mutedColor   string
	borderColor  string
	outputBg     string
}

// ShellOutputViewOption configures a ShellOutputView.
type ShellOutputViewOption func(*ShellOutputView)

// WithShellOutputViewExitCode sets the command exit code.
func WithShellOutputViewExitCode(code int) ShellOutputViewOption {
	return func(v *ShellOutputView) { v.exitCode = code }
}

// WithShellOutputViewCwd sets the working directory shown in the header.
func WithShellOutputViewCwd(cwd string) ShellOutputViewOption {
	return func(v *ShellOutputView) { v.cwd = cwd }
}

// WithShellOutputViewTruncated controls whether output truncation is indicated.
func WithShellOutputViewTruncated(truncated bool) ShellOutputViewOption {
	return func(v *ShellOutputView) { v.truncated = truncated }
}

// WithShellOutputViewMaxLines sets the maximum number of output lines shown.
func WithShellOutputViewMaxLines(max int) ShellOutputViewOption {
	return func(v *ShellOutputView) {
		if max > 0 {
			v.maxLines = max
		}
	}
}

// WithShellOutputViewDesignTokens applies design-system tokens.
func WithShellOutputViewDesignTokens(dt *design.DesignTokens) ShellOutputViewOption {
	return func(v *ShellOutputView) {
		if dt == nil {
			return
		}
		if c := dt.SuccessBright; c != "" {
			v.successColor = c
		}
		if c := dt.ErrorBright; c != "" {
			v.errorColor = c
		}
		if c := dt.MutedColor; c != "" {
			v.mutedColor = c
		}
		if c := dt.BorderSubtle; c != "" {
			v.borderColor = c
		}
		if c := dt.SurfaceRaised; c != "" {
			v.outputBg = c
		}
	}
}

// WithShellOutputViewWidth sets a fixed rendering width.
func WithShellOutputViewWidth(width int) ShellOutputViewOption {
	return func(v *ShellOutputView) {
		if width >= 0 {
			v.width = width
		}
	}
}

// NewShellOutputView creates a new ShellOutputView.
func NewShellOutputView(command string, output string, opts ...ShellOutputViewOption) *ShellOutputView {
	v := &ShellOutputView{
		command:      command,
		output:       output,
		exitCode:     0,
		maxLines:     20,
		successColor: "#98C379",
		errorColor:   "#E06C75",
		mutedColor:   "#7A818A",
		borderColor:  "#3C414B",
		outputBg:     "#31353D",
	}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

// Init initializes the component.
func (v *ShellOutputView) Init() tea.Cmd { return nil }

// Update handles component updates.
func (v *ShellOutputView) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.windowWidth = msg.Width
	}
	return v, nil
}

// View renders the shell command output card.
func (v *ShellOutputView) View() string {
	visibleLines, remaining := v.visibleOutputLines()
	contentWidth := v.contentWidth(visibleLines, remaining)
	if contentWidth < 1 {
		contentWidth = 1
	}

	headerColor := v.successANSI()
	if v.exitCode != 0 {
		headerColor = v.errorANSI()
	}
	muted := v.mutedANSI()
	border := v.borderANSI()
	outputBg := v.outputBgANSI()

	commandLine := "$ " + strings.TrimSpace(v.command)
	if commandLine == "$" {
		commandLine = "$"
	}
	commandLine = style.Truncate(commandLine, contentWidth, "…")

	footerText := "✓ Exit 0"
	footerColor := v.successANSI()
	if v.exitCode != 0 {
		footerText = fmt.Sprintf("✗ Exit %d", v.exitCode)
		footerColor = v.errorANSI()
	}
	footerText = style.Truncate(footerText, contentWidth, "…")

	var b strings.Builder
	b.WriteString(border)
	b.WriteString(strings.Repeat("─", contentWidth+2))
	b.WriteString(style.ANSIReset)
	b.WriteString("\n")

	b.WriteString(headerColor)
	b.WriteString(style.ANSIBold)
	b.WriteString(commandLine)
	b.WriteString(style.ANSIReset)
	b.WriteString("\n")

	if strings.TrimSpace(v.cwd) != "" {
		cwdLine := "in " + strings.TrimSpace(v.cwd)
		cwdLine = style.Truncate(cwdLine, contentWidth, "…")
		b.WriteString(muted)
		b.WriteString(cwdLine)
		b.WriteString(style.ANSIReset)
		b.WriteString("\n")
	}

	for _, line := range visibleLines {
		text := style.Truncate(line, contentWidth, "…")
		padding := contentWidth - style.StringWidth(text)
		if padding < 0 {
			padding = 0
		}

		b.WriteString(outputBg)
		b.WriteString(" ")
		b.WriteString(text)
		b.WriteString(strings.Repeat(" ", padding))
		b.WriteString(" ")
		b.WriteString(style.ANSIReset)
		b.WriteString("\n")
	}

	if remaining > 0 {
		line := fmt.Sprintf("... (%d more lines)", remaining)
		line = style.Truncate(line, contentWidth, "…")
		b.WriteString(muted)
		b.WriteString(line)
		b.WriteString(style.ANSIReset)
		b.WriteString("\n")
	} else if v.truncated {
		line := style.Truncate("... (truncated)", contentWidth, "…")
		b.WriteString(muted)
		b.WriteString(line)
		b.WriteString(style.ANSIReset)
		b.WriteString("\n")
	}

	b.WriteString(footerColor)
	b.WriteString(style.ANSIBold)
	b.WriteString(footerText)
	b.WriteString(style.ANSIReset)
	b.WriteString("\n")

	b.WriteString(border)
	b.WriteString(strings.Repeat("─", contentWidth+2))
	b.WriteString(style.ANSIReset)
	b.WriteString("\n")

	return b.String()
}

// Focus marks the component as focused.
func (v *ShellOutputView) Focus() { v.focused = true }

// Blur marks the component as unfocused.
func (v *ShellOutputView) Blur() { v.focused = false }

// Focused reports whether the component is focused.
func (v *ShellOutputView) Focused() bool { return v.focused }

func (v *ShellOutputView) visibleOutputLines() ([]string, int) {
	lines := strings.Split(v.output, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return nil, 0
	}

	if v.maxLines > 0 && len(lines) > v.maxLines {
		return lines[:v.maxLines], len(lines) - v.maxLines
	}

	return lines, 0
}

func (v *ShellOutputView) contentWidth(lines []string, remaining int) int {
	if v.width > 0 {
		return v.width
	}

	maxWidth := style.StringWidth("$ " + strings.TrimSpace(v.command))

	if strings.TrimSpace(v.cwd) != "" {
		if n := style.StringWidth("in " + strings.TrimSpace(v.cwd)); n > maxWidth {
			maxWidth = n
		}
	}

	for _, line := range lines {
		if n := style.StringWidth(line); n > maxWidth {
			maxWidth = n
		}
	}

	if remaining > 0 {
		hiddenLine := fmt.Sprintf("... (%d more lines)", remaining)
		if n := style.StringWidth(hiddenLine); n > maxWidth {
			maxWidth = n
		}
	} else if v.truncated {
		if n := style.StringWidth("... (truncated)"); n > maxWidth {
			maxWidth = n
		}
	}

	footer := "✓ Exit 0"
	if v.exitCode != 0 {
		footer = fmt.Sprintf("✗ Exit %d", v.exitCode)
	}
	if n := style.StringWidth(footer); n > maxWidth {
		maxWidth = n
	}

	if maxWidth < 1 {
		maxWidth = 1
	}

	if v.windowWidth > 0 {
		maxAllowed := v.windowWidth - 2
		if maxAllowed < 1 {
			maxAllowed = 1
		}
		if maxWidth > maxAllowed {
			maxWidth = maxAllowed
		}
	}

	return maxWidth
}

func (v *ShellOutputView) successANSI() string {
	if c := style.Fg(v.successColor); c != "" {
		return c
	}
	return style.ANSIGreen
}

func (v *ShellOutputView) errorANSI() string {
	if c := style.Fg(v.errorColor); c != "" {
		return c
	}
	return style.ANSIRed
}

func (v *ShellOutputView) mutedANSI() string {
	if c := style.Fg(v.mutedColor); c != "" {
		return c
	}
	return style.ANSIDim
}

func (v *ShellOutputView) borderANSI() string {
	if c := style.Fg(v.borderColor); c != "" {
		return c
	}
	return style.ANSIDim
}

func (v *ShellOutputView) outputBgANSI() string {
	if c := style.Bg(v.outputBg); c != "" {
		return c
	}
	return ""
}

var _ tui.Component = (*ShellOutputView)(nil)
