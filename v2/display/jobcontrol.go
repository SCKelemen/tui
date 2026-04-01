package display

import (
	"fmt"
	"time"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	"github.com/SCKelemen/tui/v2/style/design"
	tea "github.com/charmbracelet/bubbletea"
)

// JobSuspendedMsg is emitted when job control transitions to suspended.
type JobSuspendedMsg struct{}

// JobResumedMsg is emitted when job control transitions back to running.
type JobResumedMsg struct{}

// JobControl handles suspend/resume flow and terminal-state bookkeeping.
type JobControl struct {
	suspended        bool
	suspendAt        time.Time
	savedWidth       int
	savedHeight      int
	focused          bool
	designTokens     *design.DesignTokens
	suspendedMessage string
}

// JobControlOption configures JobControl.
type JobControlOption func(*JobControl)

// WithJobControlDesignTokens applies design tokens.
func WithJobControlDesignTokens(tokens *design.DesignTokens) JobControlOption {
	return func(j *JobControl) {
		if tokens != nil {
			j.designTokens = tokens
		}
	}
}

// WithJobControlSuspendedMessage sets custom suspended text.
func WithJobControlSuspendedMessage(msg string) JobControlOption {
	return func(j *JobControl) {
		if msg != "" {
			j.suspendedMessage = msg
		}
	}
}

// NewJobControl creates a suspend/resume helper component.
func NewJobControl(opts ...JobControlOption) *JobControl {
	j := &JobControl{
		suspended:        false,
		designTokens:     design.DefaultTheme(),
		suspendedMessage: "suspended (press fg/resume to continue)",
	}
	for _, opt := range opts {
		opt(j)
	}
	return j
}

// Init satisfies the Bubble Tea model contract.
func (j *JobControl) Init() tea.Cmd { return nil }

// Update captures Ctrl+Z suspend and resume notifications.
func (j *JobControl) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch t := msg.(type) {
	case tea.WindowSizeMsg:
		if !j.suspended {
			j.savedWidth = t.Width
			j.savedHeight = t.Height
		}
	case tea.KeyMsg:
		if !j.focused {
			return j, nil
		}
		if t.String() == "ctrl+z" {
			j.Suspend()
			return j, func() tea.Msg { return JobSuspendedMsg{} }
		}
	case JobResumedMsg:
		j.Resume()
	}
	return j, nil
}

// Suspend enters suspended mode and captures terminal layout snapshot.
func (j *JobControl) Suspend() {
	if j.suspended {
		return
	}
	j.suspended = true
	j.suspendAt = time.Now()
}

// Resume exits suspended mode.
func (j *JobControl) Resume() {
	j.suspended = false
}

// Suspended reports whether job control is currently suspended.
func (j *JobControl) Suspended() bool { return j.suspended }

// View shows a suspended indicator when paused.
func (j *JobControl) View() string {
	if !j.suspended {
		return ""
	}
	color := style.ANSIYellow
	if j.designTokens != nil {
		if v := style.Fg(j.designTokens.PendingColor); v != "" {
			color = v
		}
	}
	elapsed := time.Since(j.suspendAt).Truncate(time.Second)
	return fmt.Sprintf("%s⏸ %s [%s]%s", color, j.suspendedMessage, elapsed, style.ANSIReset)
}

// Focus marks focus state.
func (j *JobControl) Focus() { j.focused = true }

// Blur marks blur state.
func (j *JobControl) Blur() { j.focused = false }

// Focused reports focus state.
func (j *JobControl) Focused() bool { return j.focused }

var _ tui.Component = (*JobControl)(nil)
