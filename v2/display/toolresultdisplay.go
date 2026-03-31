package display

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ToolResultStatus represents the state of a tool execution.
type ToolResultStatus int

const (
	ToolResultSuccess ToolResultStatus = iota
	ToolResultError
	ToolResultCanceled
	ToolResultRejected
)

// ToolResultDisplay renders a tool execution result with status-aware styling.
type ToolResultDisplay struct {
	status        ToolResultStatus
	toolName      string
	content       string
	width         int
	truncateLines int
	windowWidth   int
	focused       bool
	designTokens  *design.DesignTokens
}

// ToolResultDisplayOption configures a ToolResultDisplay.
type ToolResultDisplayOption func(*ToolResultDisplay)

// WithToolResultToolName sets the tool name shown in the header.
func WithToolResultToolName(name string) ToolResultDisplayOption {
	return func(d *ToolResultDisplay) {
		d.toolName = strings.TrimSpace(name)
	}
}

// WithToolResultWidth sets a fixed content width.
func WithToolResultWidth(width int) ToolResultDisplayOption {
	return func(d *ToolResultDisplay) {
		if width >= 0 {
			d.width = width
		}
	}
}

// WithToolResultTruncateLines sets the max number of content lines before truncation.
func WithToolResultTruncateLines(max int) ToolResultDisplayOption {
	return func(d *ToolResultDisplay) {
		if max >= 0 {
			d.truncateLines = max
		}
	}
}

// WithToolResultDesignTokens applies design-system tokens.
func WithToolResultDesignTokens(tokens *design.DesignTokens) ToolResultDisplayOption {
	return func(d *ToolResultDisplay) {
		d.designTokens = tokens
	}
}

// NewToolResultDisplay creates a new ToolResultDisplay component.
func NewToolResultDisplay(status ToolResultStatus, content string, opts ...ToolResultDisplayOption) *ToolResultDisplay {
	d := &ToolResultDisplay{
		status:        status,
		toolName:      "tool",
		content:       content,
		truncateLines: 0,
		designTokens:  design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(d)
	}

	if strings.TrimSpace(d.toolName) == "" {
		d.toolName = "tool"
	}

	return d
}

// Init initializes the component.
func (d *ToolResultDisplay) Init() tea.Cmd {
	return nil
}

// Update handles component updates.
func (d *ToolResultDisplay) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.windowWidth = msg.Width
	}

	return d, nil
}

// View renders the tool result.
func (d *ToolResultDisplay) View() string {
	header := d.headerText()
	allLines := d.contentLines()
	visibleLines, hidden := d.truncatedLines(allLines)

	contentWidth := d.contentWidth(header, visibleLines, hidden)
	if contentWidth < 1 {
		contentWidth = 1
	}

	statusColor := d.statusColorANSI()
	headerColor := statusColor
	contentColor := style.ANSIReset
	if d.status == ToolResultError {
		contentColor = statusColor
	}
	truncationColor := d.linkColorANSI()

	var b strings.Builder

	renderedHeader := style.Truncate(header, contentWidth, "…")
	b.WriteString(d.renderLine(renderedHeader, contentWidth, statusColor, headerColor))

	for _, line := range visibleLines {
		b.WriteString("\n")
		rendered := style.Truncate(line, contentWidth, "…")
		b.WriteString(d.renderLine(rendered, contentWidth, statusColor, contentColor))
	}

	if hidden > 0 {
		b.WriteString("\n")
		linkText := fmt.Sprintf("... (%d more lines)", hidden)
		linkText = style.Truncate(linkText, contentWidth, "…")
		b.WriteString(d.renderLine(linkText, contentWidth, statusColor, style.ANSIUnderline+truncationColor))
	}

	b.WriteString("\n")
	return b.String()
}

// Focus marks the component as focused.
func (d *ToolResultDisplay) Focus() {
	d.focused = true
}

// Blur marks the component as unfocused.
func (d *ToolResultDisplay) Blur() {
	d.focused = false
}

// Focused reports whether the component is focused.
func (d *ToolResultDisplay) Focused() bool {
	return d.focused
}

func (d *ToolResultDisplay) renderLine(text string, contentWidth int, borderColor, textColor string) string {
	padding := contentWidth - style.StringWidth(text)
	if padding < 0 {
		padding = 0
	}

	var b strings.Builder
	b.WriteString(borderColor)
	b.WriteString("│")
	b.WriteString(style.ANSIReset)
	b.WriteString(" ")
	b.WriteString(textColor)
	b.WriteString(text)
	b.WriteString(style.ANSIReset)
	b.WriteString(strings.Repeat(" ", padding))
	return b.String()
}

func (d *ToolResultDisplay) headerText() string {
	name := strings.TrimSpace(d.toolName)
	if name == "" {
		name = "tool"
	}

	switch d.status {
	case ToolResultError:
		return "✗ " + name
	case ToolResultCanceled:
		return "⊘ " + name + " — canceled"
	case ToolResultRejected:
		return "⊘ " + name + " — rejected by user"
	default:
		return "✓ " + name
	}
}

func (d *ToolResultDisplay) contentLines() []string {
	if strings.TrimSpace(d.content) == "" {
		return nil
	}

	raw := strings.Split(d.content, "\n")
	if len(raw) > 0 && raw[len(raw)-1] == "" {
		raw = raw[:len(raw)-1]
	}

	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		lines = append(lines, line)
	}

	return lines
}

func (d *ToolResultDisplay) truncatedLines(lines []string) ([]string, int) {
	if d.truncateLines <= 0 || len(lines) <= d.truncateLines {
		return lines, 0
	}
	return lines[:d.truncateLines], len(lines) - d.truncateLines
}

func (d *ToolResultDisplay) contentWidth(header string, lines []string, hidden int) int {
	if d.width > 0 {
		return d.width
	}

	maxWidth := style.StringWidth(header)

	for _, line := range lines {
		if w := style.StringWidth(line); w > maxWidth {
			maxWidth = w
		}
	}

	if hidden > 0 {
		if w := style.StringWidth(fmt.Sprintf("... (%d more lines)", hidden)); w > maxWidth {
			maxWidth = w
		}
	}

	if maxWidth < 1 {
		maxWidth = 1
	}

	if d.windowWidth > 0 {
		maxContent := d.windowWidth - 2
		if maxContent < 1 {
			maxContent = 1
		}
		if maxWidth > maxContent {
			maxWidth = maxContent
		}
	}

	return maxWidth
}

func (d *ToolResultDisplay) statusColorANSI() string {
	switch d.status {
	case ToolResultError:
		return style.ANSIRed
	case ToolResultCanceled:
		return style.ANSIYellow
	case ToolResultRejected:
		return style.ANSIColorFromHex("#F2994A")
	default:
		return style.ANSIGreen
	}
}

func (d *ToolResultDisplay) linkColorANSI() string {
	if d.designTokens != nil {
		if c := style.ANSIColorFromHex(d.designTokens.Accent); c != "" {
			return c
		}
	}
	return style.ANSICyan
}

var _ tui.Component = (*ToolResultDisplay)(nil)
