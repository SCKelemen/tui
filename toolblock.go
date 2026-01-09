package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

// NewToolBlock creates a new tool block
func NewToolBlock(toolName, command string, output []string, opts ...ToolBlockOption) *ToolBlock {
	tb := &ToolBlock{
		toolName: toolName,
		command:  command,
		output:   output,
		expanded: false,
		maxLines: 5, // Default: show first 5 lines when collapsed
		icon:     getToolIcon(toolName),
	}

	for _, opt := range opts {
		opt(tb)
	}

	return tb
}

// Init initializes the tool block
func (tb *ToolBlock) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (tb *ToolBlock) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		tb.width = msg.Width

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

	// Header: ⏺ Bash(command)
	header := fmt.Sprintf("%s \033[1m%s\033[0m\033[2m(%s)\033[0m",
		tb.icon,
		tb.toolName,
		truncateString(tb.command, tb.width-len(tb.toolName)-10))

	if tb.focused {
		header = "\033[7m" + header + "\033[0m" // Inverted when focused
	}

	lines = append(lines, header)

	// Output with tree connector
	if len(tb.output) == 0 {
		lines = append(lines, "  \033[2m⎿  (no output)\033[0m")
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
