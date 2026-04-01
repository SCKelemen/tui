package display

import (
	"fmt"
	"strings"
	"time"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	"github.com/SCKelemen/tui/v2/style/design"
	tea "github.com/charmbracelet/bubbletea"
)
// ExecCellState represents execution lifecycle states.
type ExecCellState int

const (
	// ExecCellPending means the command has not started.
	ExecCellPending ExecCellState = iota
	// ExecCellRunning means the command is currently executing.
	ExecCellRunning
	// ExecCellCompleted means the command finished successfully.
	ExecCellCompleted
	// ExecCellFailed means the command finished with error.
	ExecCellFailed
	// ExecCellCancelled means the command was canceled.
	ExecCellCancelled
)

type execCellTickMsg time.Time

// ExecCell displays a command execution with spinner, stream output, and status.
type ExecCell struct {
	command      string
	state        ExecCellState
	stdout       []string
	stderr       []string
	scrollOffset int
	maxVisible   int
	width        int
	height       int
	startTime    time.Time
	endTime      time.Time
	exitCode     *int
	focused      bool
	spinnerFrame int
	designTokens *design.DesignTokens
}

// ExecCellOption configures an ExecCell.
type ExecCellOption func(*ExecCell)

// WithExecCellCommand sets command text.
func WithExecCellCommand(cmd string) ExecCellOption {
	return func(e *ExecCell) { e.command = strings.TrimSpace(cmd) }
}

// WithExecCellMaxVisibleLines configures visible output lines.
func WithExecCellMaxVisibleLines(lines int) ExecCellOption {
	return func(e *ExecCell) {
		if lines > 0 {
			e.maxVisible = lines
		}
	}
}

// WithExecCellDesignTokens applies design tokens.
func WithExecCellDesignTokens(tokens *design.DesignTokens) ExecCellOption {
	return func(e *ExecCell) {
		if tokens != nil {
			e.designTokens = tokens
		}
	}
}

// NewExecCell creates an execution display cell.
func NewExecCell(opts ...ExecCellOption) *ExecCell {
	e := &ExecCell{
		command:      "",
		state:        ExecCellPending,
		stdout:       make([]string, 0, 64),
		stderr:       make([]string, 0, 32),
		scrollOffset: 0,
		maxVisible:   12,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Init satisfies the Bubble Tea model contract.
func (e *ExecCell) Init() tea.Cmd {
	if e.state == ExecCellRunning {
		return e.tick()
	}
	return nil
}

// Update handles spinner progression, scrolling, and sizing.
func (e *ExecCell) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch t := msg.(type) {
	case tea.WindowSizeMsg:
		e.width = t.Width
		e.height = t.Height
	case execCellTickMsg:
		if e.state == ExecCellRunning {
			e.spinnerFrame++
			return e, e.tick()
		}
	case tea.KeyMsg:
		if !e.focused {
			return e, nil
		}
		switch t.String() {
		case "up", "k":
			if e.scrollOffset > 0 {
				e.scrollOffset--
			}
		case "down", "j":
			if e.scrollOffset < e.maxScroll() {
				e.scrollOffset++
			}
		}
	}

	return e, nil
}

// Start marks command execution as running.
func (e *ExecCell) Start() {
	now := time.Now()
	e.state = ExecCellRunning
	e.startTime = now
	e.endTime = time.Time{}
	e.exitCode = nil
}

// AppendStdout appends streamed stdout chunks.
func (e *ExecCell) AppendStdout(text string) {
	e.stdout = appendExecCellSplitLines(e.stdout, text)
}
// AppendStderr appends streamed stderr chunks.
func (e *ExecCell) AppendStderr(text string) {
	e.stderr = appendExecCellSplitLines(e.stderr, text)
}
// Complete marks execution complete with an exit code.
func (e *ExecCell) Complete(exitCode int) {
	e.state = ExecCellCompleted
	if exitCode != 0 {
		e.state = ExecCellFailed
	}
	e.endTime = time.Now()
	e.exitCode = &exitCode
}

// Cancel marks execution as canceled.
func (e *ExecCell) Cancel() {
	e.state = ExecCellCancelled
	e.endTime = time.Now()
}

// View renders execution state, command, and output region.
func (e *ExecCell) View() string {
	status := e.statusLine()
	cmd := "$ " + e.command
	if strings.TrimSpace(e.command) == "" {
		cmd = "$ (no command)"
	}

	lines := []string{status, cmd, ""}
	output := e.mergedOutput()
	start := e.scrollOffset
	end := start + e.maxVisible
	if end > len(output) {
		end = len(output)
	}
	if start < len(output) {
		lines = append(lines, output[start:end]...)
	}

	if len(output) == 0 {
		lines = append(lines, style.ANSIDim+"(no output yet)"+style.ANSIReset)
	}

	return strings.Join(lines, "\n")
}

// Focus marks focus state.
func (e *ExecCell) Focus() { e.focused = true }

// Blur marks blur state.
func (e *ExecCell) Blur() { e.focused = false }

// Focused reports focus state.
func (e *ExecCell) Focused() bool { return e.focused }

func (e *ExecCell) tick() tea.Cmd {
	return tea.Tick(90*time.Millisecond, func(t time.Time) tea.Msg {
		return execCellTickMsg(t)
	})
}

func (e *ExecCell) statusLine() string {
	stateText := "pending"
	stateColor := style.ANSIDim
	spinner := "•"

	switch e.state {
	case ExecCellRunning:
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinner = frames[e.spinnerFrame%len(frames)]
		stateText = "running"
		stateColor = style.ANSICyan
	case ExecCellCompleted:
		spinner = "✓"
		stateText = "completed"
		stateColor = style.ANSIGreen
	case ExecCellFailed:
		spinner = "✗"
		stateText = "failed"
		stateColor = style.ANSIRed
	case ExecCellCancelled:
		spinner = "⊘"
		stateText = "cancelled"
		stateColor = style.ANSIYellow
	}

	if e.designTokens != nil {
		switch e.state {
		case ExecCellRunning:
			if v := style.Fg(e.designTokens.Accent); v != "" {
				stateColor = v
			}
		case ExecCellCompleted:
			if v := style.Fg(e.designTokens.SuccessBright); v != "" {
				stateColor = v
			}
		case ExecCellFailed:
			if v := style.Fg(e.designTokens.ErrorBright); v != "" {
				stateColor = v
			}
		case ExecCellCancelled:
			if v := style.Fg(e.designTokens.PendingColor); v != "" {
				stateColor = v
			}
		}
	}

	elapsed := time.Since(e.startTime)
	if e.startTime.IsZero() {
		elapsed = 0
	}
	if !e.endTime.IsZero() {
		elapsed = e.endTime.Sub(e.startTime)
	}

	exitText := ""
	if e.exitCode != nil {
		exitText = fmt.Sprintf(" exit=%d", *e.exitCode)
	}

	return fmt.Sprintf("%s%s %s%s elapsed=%s%s", stateColor, spinner, stateText, style.ANSIReset, elapsed.Truncate(time.Millisecond), exitText)
}

func (e *ExecCell) mergedOutput() []string {
	out := make([]string, 0, len(e.stdout)+len(e.stderr))
	for _, line := range e.stdout {
		out = append(out, line)
	}
	for _, line := range e.stderr {
		out = append(out, style.ANSIRed+line+style.ANSIReset)
	}
	return out
}

func (e *ExecCell) maxScroll() int {
	max := len(e.mergedOutput()) - e.maxVisible
	if max < 0 {
		return 0
	}
	return max
}

func appendExecCellSplitLines(dst []string, text string) []string {
	if text == "" {
		return dst
	}

	parts := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for i, p := range parts {
		if i == len(parts)-1 && p == "" {
			continue
		}
		dst = append(dst, p)
	}

	return dst
}

var _ tui.Component = (*ExecCell)(nil)