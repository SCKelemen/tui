package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ActivityBar displays an animated status line with spinner, elapsed time, and progress
type ActivityBar struct {
	width      int
	message    string
	active     bool
	startTime  time.Time
	elapsed    time.Duration
	spinner    int
	focused    bool
	progress   string // e.g., "↓ 2.5k tokens"
	cancelable bool
}

// tickMsg is sent periodically to update the spinner and timer
type tickMsg time.Time

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewActivityBar creates a new activity bar
func NewActivityBar() *ActivityBar {
	return &ActivityBar{
		message:    "Ready",
		cancelable: true,
	}
}

// Init initializes the activity bar
func (a *ActivityBar) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (a *ActivityBar) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width

	case tickMsg:
		if a.active {
			a.spinner = (a.spinner + 1) % len(spinnerFrames)
			a.elapsed = time.Since(a.startTime)
			return a, a.tick()
		}

	case tea.KeyMsg:
		if a.focused && a.active && a.cancelable && msg.String() == "esc" {
			a.Stop()
		}
	}
	return a, nil
}

// View renders the activity bar
func (a *ActivityBar) View() string {
	if a.width == 0 {
		return ""
	}

	if !a.active {
		// Inactive state - simple message
		return fmt.Sprintf("\033[2m%s\033[0m\n", a.message)
	}

	// Active state - animated spinner
	var parts []string

	// Spinner + message
	spinner := spinnerFrames[a.spinner]
	parts = append(parts, fmt.Sprintf("\033[1;36m%s\033[0m %s", spinner, a.message))

	// Build status info
	var status []string

	// Cancelable hint
	if a.cancelable {
		status = append(status, "\033[2mesc to interrupt\033[0m")
	}

	// Elapsed time
	if a.elapsed > 0 {
		status = append(status, a.formatDuration(a.elapsed))
	}

	// Progress indicator
	if a.progress != "" {
		status = append(status, fmt.Sprintf("\033[36m%s\033[0m", a.progress))
	}

	if len(status) > 0 {
		parts = append(parts, fmt.Sprintf("(\033[2m%s\033[0m)", strings.Join(status, " · ")))
	}

	line := strings.Join(parts, " ")

	// Truncate if too long
	if len(stripANSI(line)) > a.width {
		line = truncateANSI(line, a.width-3) + "..."
	}

	return line + "\n"
}

// Focus is called when this component receives focus
func (a *ActivityBar) Focus() {
	a.focused = true
}

// Blur is called when this component loses focus
func (a *ActivityBar) Blur() {
	a.focused = false
}

// Focused returns whether this component is currently focused
func (a *ActivityBar) Focused() bool {
	return a.focused
}

// Start begins the activity animation
func (a *ActivityBar) Start(message string) tea.Cmd {
	a.message = message
	a.active = true
	a.startTime = time.Now()
	a.elapsed = 0
	a.spinner = 0
	return a.tick()
}

// Stop stops the activity animation
func (a *ActivityBar) Stop() {
	a.active = false
	a.message = "Ready"
	a.progress = ""
}

// SetProgress updates the progress indicator
func (a *ActivityBar) SetProgress(progress string) {
	a.progress = progress
}

// tick returns a command that sends a tickMsg after a delay
func (a *ActivityBar) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// formatDuration formats a duration as "1m 14s" or "14s"
func (a *ActivityBar) formatDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

// stripANSI removes ANSI escape codes (reused from border_components.go logic)
func stripANSI(s string) string {
	var result strings.Builder
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
		} else if inEscape {
			if r == 'm' {
				inEscape = false
			}
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// truncateANSI truncates preserving ANSI codes (reused from border_components.go logic)
func truncateANSI(s string, maxWidth int) string {
	var result strings.Builder
	visualWidth := 0
	inEscape := false

	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inEscape = true
			result.WriteByte(s[i])
		} else if inEscape {
			result.WriteByte(s[i])
			if s[i] == 'm' {
				inEscape = false
			}
		} else {
			if visualWidth >= maxWidth {
				break
			}
			result.WriteByte(s[i])
			visualWidth++
		}
	}

	return result.String()
}
