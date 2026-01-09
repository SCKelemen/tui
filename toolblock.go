package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ToolBlockStatus represents the execution state
type ToolBlockStatus int

const (
	StatusRunning ToolBlockStatus = iota
	StatusComplete
	StatusError
	StatusWarning
)

// ToolBlock represents a collapsible block showing tool execution results
type ToolBlock struct {
	width       int
	toolName    string // e.g., "Bash", "Write", "Read"
	command     string // e.g., "cd ~/Code && ls"
	output      []string
	expanded    bool
	focused     bool
	showLineNos bool // Show line numbers (for code files)
	icon        string
	maxLines    int // Max lines to show when collapsed (0 = show all)
	status      ToolBlockStatus
	spinner     int
	streaming   bool // Enable streaming mode
}

// ToolBlockOption configures a ToolBlock
type ToolBlockOption func(*ToolBlock)

// WithLineNumbers enables line numbers
func WithLineNumbers() ToolBlockOption {
	return func(tb *ToolBlock) {
		tb.showLineNos = true
	}
}

// WithMaxLines sets the maximum lines to show when collapsed
func WithMaxLines(n int) ToolBlockOption {
	return func(tb *ToolBlock) {
		tb.maxLines = n
	}
}

// WithStreaming enables streaming mode for real-time output
func WithStreaming() ToolBlockOption {
	return func(tb *ToolBlock) {
		tb.streaming = true
		tb.status = StatusRunning
	}
}

// WithStatus sets the initial status
func WithStatus(status ToolBlockStatus) ToolBlockOption {
	return func(tb *ToolBlock) {
		tb.status = status
	}
}

// NewToolBlock creates a new tool block
func NewToolBlock(toolName, command string, output []string, opts ...ToolBlockOption) *ToolBlock {
	tb := &ToolBlock{
		toolName: toolName,
		command:  command,
		output:   output,
		expanded: false,
		maxLines: 5, // Default: show first 5 lines when collapsed
		icon:     getToolIcon(toolName),
		status:   StatusComplete, // Default to complete
	}

	for _, opt := range opts {
		opt(tb)
	}

	return tb
}

// Init initializes the tool block
func (tb *ToolBlock) Init() tea.Cmd {
	if tb.streaming && tb.status == StatusRunning {
		return tb.tick()
	}
	return nil
}

// toolBlockTickMsg is sent to animate the spinner
type toolBlockTickMsg struct {
	id *ToolBlock
}

// Update handles messages
func (tb *ToolBlock) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		tb.width = msg.Width

	case toolBlockTickMsg:
		if msg.id == tb && tb.streaming && tb.status == StatusRunning {
			tb.spinner = (tb.spinner + 1) % len(spinnerFrames)
			return tb, tb.tick()
		}

	case tea.KeyMsg:
		if tb.focused {
			switch msg.String() {
			case "ctrl+o", "enter":
				tb.ToggleExpanded()
			}
		}
	}
	return tb, nil
}

// View renders the tool block
func (tb *ToolBlock) View() string {
	if tb.width == 0 {
		return ""
	}

	var lines []string

	// Get status indicator and color
	statusIcon, statusColor := tb.getStatusIndicator()

	// Header with status: [icon] Bash(command) [status]
	var header string
	if tb.streaming && tb.status == StatusRunning {
		// Show spinner when streaming
		spinner := spinnerFrames[tb.spinner]
		header = fmt.Sprintf("%s%s\033[0m \033[1m%s\033[0m\033[2m(%s)\033[0m %s%s\033[0m",
			statusColor,
			tb.icon,
			tb.toolName,
			truncateString(tb.command, tb.width-len(tb.toolName)-20),
			statusColor,
			spinner)
	} else {
		header = fmt.Sprintf("%s%s\033[0m \033[1m%s\033[0m\033[2m(%s)\033[0m %s",
			statusColor,
			tb.icon,
			tb.toolName,
			truncateString(tb.command, tb.width-len(tb.toolName)-15),
			statusIcon)
	}

	if tb.focused {
		header = "\033[7m" + header + "\033[0m" // Inverted when focused
	}

	lines = append(lines, header)

	// Output with tree connector
	if len(tb.output) == 0 {
		if tb.streaming && tb.status == StatusRunning {
			lines = append(lines, "  \033[2m⎿  \033[3mstreaming...\033[0m")
		} else {
			lines = append(lines, "  \033[2m⎿  (no output)\033[0m")
		}
		return strings.Join(lines, "\n") + "\n"
	}

	outputLines := tb.output
	hiddenCount := 0

	if !tb.expanded && tb.maxLines > 0 && len(tb.output) > tb.maxLines {
		outputLines = tb.output[:tb.maxLines]
		hiddenCount = len(tb.output) - tb.maxLines
	}

	for i, line := range outputLines {
		prefix := "  \033[2m⎿\033[0m  "
		if i > 0 {
			prefix = "     " // Indent continuation lines
		}

		// Add line numbers if enabled
		if tb.showLineNos {
			lineNo := fmt.Sprintf("\033[2m%3d\033[0m ", i+1)
			prefix += lineNo
		}

		// Truncate long lines
		displayLine := line
		maxWidth := tb.width - len(stripANSI(prefix)) - 2
		if len(displayLine) > maxWidth {
			displayLine = displayLine[:maxWidth-3] + "..."
		}

		lines = append(lines, prefix+displayLine)
	}

	// Show "... +N lines" if collapsed
	if hiddenCount > 0 {
		expandHint := fmt.Sprintf("     \033[2m… +%d lines \033[0m\033[3m(ctrl+o to expand)\033[0m",
			hiddenCount)
		lines = append(lines, expandHint)
	}

	return strings.Join(lines, "\n") + "\n"
}

// Focus is called when this component receives focus
func (tb *ToolBlock) Focus() {
	tb.focused = true
}

// Blur is called when this component loses focus
func (tb *ToolBlock) Blur() {
	tb.focused = false
}

// Focused returns whether this component is currently focused
func (tb *ToolBlock) Focused() bool {
	return tb.focused
}

// ToggleExpanded toggles the expanded state
func (tb *ToolBlock) ToggleExpanded() {
	tb.expanded = !tb.expanded
}

// SetExpanded sets the expanded state
func (tb *ToolBlock) SetExpanded(expanded bool) {
	tb.expanded = expanded
}

// getToolIcon returns an icon for the tool type
func getToolIcon(toolName string) string {
	icons := map[string]string{
		"Bash":   "⏺",
		"Write":  "⏺",
		"Read":   "⏺",
		"Edit":   "⏺",
		"Grep":   "⏺",
		"Glob":   "⏺",
		"Task":   "⏺",
		"WebFetch": "⏺",
	}

	if icon, ok := icons[toolName]; ok {
		return icon
	}
	return "⏺" // Default icon
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// getStatusIndicator returns the icon and color for the current status
func (tb *ToolBlock) getStatusIndicator() (string, string) {
	switch tb.status {
	case StatusRunning:
		return "", "\033[36m" // Cyan
	case StatusComplete:
		return "\033[32m✓\033[0m", "\033[32m" // Green
	case StatusError:
		return "\033[31m✗\033[0m", "\033[31m" // Red
	case StatusWarning:
		return "\033[33m⚠\033[0m", "\033[33m" // Yellow
	default:
		return "", "\033[0m"
	}
}

// tick returns a command that sends a tick message for spinner animation
func (tb *ToolBlock) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return toolBlockTickMsg{id: tb}
	})
}

// AppendLine adds a single line to the output (for streaming)
func (tb *ToolBlock) AppendLine(line string) {
	tb.output = append(tb.output, line)
}

// AppendLines adds multiple lines to the output (for streaming)
func (tb *ToolBlock) AppendLines(lines []string) {
	tb.output = append(tb.output, lines...)
}

// SetStatus updates the status and stops streaming if completed
func (tb *ToolBlock) SetStatus(status ToolBlockStatus) {
	tb.status = status
	if status != StatusRunning {
		tb.streaming = false
	}
}

// StartStreaming begins streaming mode with running status
func (tb *ToolBlock) StartStreaming() tea.Cmd {
	tb.streaming = true
	tb.status = StatusRunning
	tb.spinner = 0
	return tb.tick()
}

// StopStreaming stops streaming and sets status to complete
func (tb *ToolBlock) StopStreaming() {
	tb.streaming = false
	tb.status = StatusComplete
}

// StopStreamingWithError stops streaming and sets status to error
func (tb *ToolBlock) StopStreamingWithError() {
	tb.streaming = false
	tb.status = StatusError
}
