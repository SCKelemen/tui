package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// CodeBlock displays source code with line numbers, syntax highlighting (future), and collapse/expand
type CodeBlock struct {
	width     int
	height    int
	focused   bool

	// Content
	operation string   // e.g., "Write", "Read", "Edit"
	filename  string   // File being operated on
	summary   string   // e.g., "Wrote 253 lines to file.go"
	lines     []string // Code lines
	language  string   // Programming language (for future syntax highlighting)

	// Display state
	expanded     bool // Whether code is shown or collapsed
	maxLines     int  // Maximum lines to show when expanded (0 = show all)
	startLine    int  // Starting line number (1-indexed)
	showPreview  int  // Number of lines to show when collapsed (default 8)
}

// CodeBlockOption configures a CodeBlock
type CodeBlockOption func(*CodeBlock)

// WithCodeOperation sets the operation type (Write, Read, Edit, etc.)
func WithCodeOperation(op string) CodeBlockOption {
	return func(cb *CodeBlock) {
		cb.operation = op
	}
}

// WithCodeFilename sets the filename
func WithCodeFilename(name string) CodeBlockOption {
	return func(cb *CodeBlock) {
		cb.filename = name
	}
}

// WithCodeSummary sets the summary text
func WithCodeSummary(summary string) CodeBlockOption {
	return func(cb *CodeBlock) {
		cb.summary = summary
	}
}

// WithCodeLines sets the code content as lines
func WithCodeLines(lines []string) CodeBlockOption {
	return func(cb *CodeBlock) {
		cb.lines = lines
	}
}

// WithCode sets the code content from a single string
func WithCode(code string) CodeBlockOption {
	return func(cb *CodeBlock) {
		cb.lines = strings.Split(code, "\n")
	}
}

// WithLanguage sets the programming language
func WithLanguage(lang string) CodeBlockOption {
	return func(cb *CodeBlock) {
		cb.language = lang
	}
}

// WithStartLine sets the starting line number
func WithStartLine(line int) CodeBlockOption {
	return func(cb *CodeBlock) {
		cb.startLine = line
	}
}

// WithExpanded sets whether the block starts expanded
func WithExpanded(expanded bool) CodeBlockOption {
	return func(cb *CodeBlock) {
		cb.expanded = expanded
	}
}

// WithCodeMaxLines sets maximum lines to show when expanded
func WithCodeMaxLines(max int) CodeBlockOption {
	return func(cb *CodeBlock) {
		cb.maxLines = max
	}
}

// WithPreviewLines sets number of preview lines when collapsed
func WithPreviewLines(n int) CodeBlockOption {
	return func(cb *CodeBlock) {
		cb.showPreview = n
	}
}

// NewCodeBlock creates a new code block component
func NewCodeBlock(opts ...CodeBlockOption) *CodeBlock {
	cb := &CodeBlock{
		operation:   "Code",
		startLine:   1,
		showPreview: 8,
		expanded:    false,
	}

	for _, opt := range opts {
		opt(cb)
	}

	return cb
}

// Init initializes the code block
func (cb *CodeBlock) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (cb *CodeBlock) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cb.width = msg.Width
		cb.height = msg.Height

	case tea.KeyMsg:
		if !cb.focused {
			return cb, nil
		}

		switch msg.String() {
		case "ctrl+o", "enter", " ":
			cb.Toggle()
		}
	}

	return cb, nil
}

// View renders the code block
func (cb *CodeBlock) View() string {
	if len(cb.lines) == 0 {
		return ""
	}

	var b strings.Builder

	// Header: ⏺ Operation(filename)
	icon := cb.getOperationIcon()
	b.WriteString(fmt.Sprintf("%s \033[1m%s\033[0m", icon, cb.operation))
	if cb.filename != "" {
		b.WriteString(fmt.Sprintf("(\033[36m%s\033[0m)", cb.filename))
	}
	b.WriteString("\n")

	// Summary line
	if cb.summary != "" {
		b.WriteString(fmt.Sprintf("  \033[2m⎿  %s\033[0m\n", cb.summary))
	}

	// Code lines
	if cb.expanded {
		b.WriteString(cb.renderExpanded())
	} else {
		b.WriteString(cb.renderCollapsed())
	}

	return b.String()
}

// Focus is called when this component receives focus
func (cb *CodeBlock) Focus() {
	cb.focused = true
}

// Blur is called when this component loses focus
func (cb *CodeBlock) Blur() {
	cb.focused = false
}

// Focused returns whether this component is currently focused
func (cb *CodeBlock) Focused() bool {
	return cb.focused
}

// Toggle expands or collapses the code block
func (cb *CodeBlock) Toggle() {
	cb.expanded = !cb.expanded
}

// Expand shows the full code
func (cb *CodeBlock) Expand() {
	cb.expanded = true
}

// Collapse hides most of the code
func (cb *CodeBlock) Collapse() {
	cb.expanded = false
}

// IsExpanded returns whether the code is currently expanded
func (cb *CodeBlock) IsExpanded() bool {
	return cb.expanded
}

// getOperationIcon returns an icon for the operation type
func (cb *CodeBlock) getOperationIcon() string {
	switch strings.ToLower(cb.operation) {
	case "write", "create":
		return "\033[32m⏺\033[0m" // Green circle
	case "read", "view":
		return "\033[34m⏺\033[0m" // Blue circle
	case "edit", "update":
		return "\033[33m⏺\033[0m" // Yellow circle
	case "delete", "remove":
		return "\033[31m⏺\033[0m" // Red circle
	default:
		return "\033[37m⏺\033[0m" // White circle
	}
}

// renderCollapsed shows preview lines + "… +N lines" indicator
func (cb *CodeBlock) renderCollapsed() string {
	var b strings.Builder

	linesToShow := cb.showPreview
	if linesToShow > len(cb.lines) {
		linesToShow = len(cb.lines)
	}

	// Show preview lines
	for i := 0; i < linesToShow; i++ {
		lineNum := cb.startLine + i
		b.WriteString(cb.renderLine(lineNum, cb.lines[i]))
	}

	// Show "… +N lines" indicator
	remainingLines := len(cb.lines) - linesToShow
	if remainingLines > 0 {
		b.WriteString(fmt.Sprintf("     \033[2m… +%d lines (\033[3mctrl+o to expand\033[0m\033[2m)\033[0m\n", remainingLines))
	}

	return b.String()
}

// renderExpanded shows all lines (up to maxLines if set)
func (cb *CodeBlock) renderExpanded() string {
	var b strings.Builder

	linesToShow := len(cb.lines)
	if cb.maxLines > 0 && linesToShow > cb.maxLines {
		linesToShow = cb.maxLines
	}

	for i := 0; i < linesToShow; i++ {
		lineNum := cb.startLine + i
		b.WriteString(cb.renderLine(lineNum, cb.lines[i]))
	}

	// Show "… +N more lines" if truncated
	if cb.maxLines > 0 && len(cb.lines) > cb.maxLines {
		remainingLines := len(cb.lines) - cb.maxLines
		b.WriteString(fmt.Sprintf("     \033[2m… +%d more lines (truncated)\033[0m\n", remainingLines))
	}

	return b.String()
}

// renderLine renders a single line with line number
func (cb *CodeBlock) renderLine(lineNum int, content string) string {
	// Calculate width needed for line numbers
	maxLineNum := cb.startLine + len(cb.lines) - 1
	lineNumWidth := len(fmt.Sprintf("%d", maxLineNum))

	// Render: "      1 package main"
	return fmt.Sprintf("  \033[2m%*d\033[0m %s\n", lineNumWidth, lineNum, content)
}
