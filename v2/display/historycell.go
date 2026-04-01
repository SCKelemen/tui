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

// HistoryCellKind classifies the conversation turn style.
type HistoryCellKind int

const (
	// HistoryCellUser renders a user message bubble.
	HistoryCellUser HistoryCellKind = iota
	// HistoryCellAssistant renders assistant output text.
	HistoryCellAssistant
	// HistoryCellToolUse renders a tool invocation block.
	HistoryCellToolUse
	// HistoryCellToolResult renders a tool result block.
	HistoryCellToolResult
)

// HistoryCell represents one conversation entry in history.
type HistoryCell struct {
	kind          HistoryCellKind
	content       string
	toolName      string
	toolStatus    string
	timestamp     time.Time
	tokenCount    int
	collapsed     bool
	width         int
	focused       bool
	designTokens  *design.DesignTokens
	showTimestamp bool
}

// HistoryCellOption configures a HistoryCell.
type HistoryCellOption func(*HistoryCell)

// WithHistoryCellToolName sets an optional tool name.
func WithHistoryCellToolName(name string) HistoryCellOption {
	return func(c *HistoryCell) { c.toolName = strings.TrimSpace(name) }
}

// WithHistoryCellToolStatus sets tool status text.
func WithHistoryCellToolStatus(status string) HistoryCellOption {
	return func(c *HistoryCell) { c.toolStatus = strings.TrimSpace(status) }
}

// WithHistoryCellTimestamp sets an explicit timestamp.
func WithHistoryCellTimestamp(ts time.Time) HistoryCellOption {
	return func(c *HistoryCell) { c.timestamp = ts }
}

// WithHistoryCellTokenCount sets token count metadata.
func WithHistoryCellTokenCount(count int) HistoryCellOption {
	return func(c *HistoryCell) {
		if count >= 0 {
			c.tokenCount = count
		}
	}
}

// WithHistoryCellCollapsed configures initial collapsed state.
func WithHistoryCellCollapsed(collapsed bool) HistoryCellOption {
	return func(c *HistoryCell) { c.collapsed = collapsed }
}

// WithHistoryCellWidth sets preferred width.
func WithHistoryCellWidth(width int) HistoryCellOption {
	return func(c *HistoryCell) {
		if width >= 0 {
			c.width = width
		}
	}
}

// WithHistoryCellDesignTokens applies design tokens.
func WithHistoryCellDesignTokens(tokens *design.DesignTokens) HistoryCellOption {
	return func(c *HistoryCell) {
		if tokens != nil {
			c.designTokens = tokens
		}
	}
}

// WithHistoryCellShowTimestamp enables/disables timestamp rendering.
func WithHistoryCellShowTimestamp(show bool) HistoryCellOption {
	return func(c *HistoryCell) { c.showTimestamp = show }
}

// NewHistoryCell creates a new conversation history cell.
func NewHistoryCell(kind HistoryCellKind, content string, opts ...HistoryCellOption) *HistoryCell {
	c := &HistoryCell{
		kind:          kind,
		content:       content,
		toolStatus:    "pending",
		timestamp:     time.Now(),
		tokenCount:    0,
		collapsed:     false,
		width:         0,
		designTokens:  design.DefaultTheme(),
		showTimestamp: true,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Init satisfies the Bubble Tea model contract.
func (c *HistoryCell) Init() tea.Cmd { return nil }

// Update handles key interactions, including collapse toggling for tool results.
func (c *HistoryCell) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch t := msg.(type) {
	case tea.WindowSizeMsg:
		if c.width == 0 {
			c.width = t.Width
		}
	case tea.KeyMsg:
		if !c.focused {
			return c, nil
		}
		if c.kind == HistoryCellToolResult && (t.String() == "enter" || t.String() == " ") {
			c.collapsed = !c.collapsed
		}
	}
	return c, nil
}

// View renders the cell with role-specific visual treatment.
func (c *HistoryCell) View() string {
	meta := c.renderMeta()
	body := c.content
	if c.kind == HistoryCellToolResult && c.collapsed {
		body = "(collapsed)"
	}

	switch c.kind {
	case HistoryCellUser:
		return c.userBubble(meta, body)
	case HistoryCellAssistant:
		return c.assistantText(meta, body)
	case HistoryCellToolUse:
		return c.toolCall(meta, body)
	case HistoryCellToolResult:
		return c.toolResult(meta, body)
	default:
		return meta + "\n" + body
	}
}

// Focus marks focus state.
func (c *HistoryCell) Focus() { c.focused = true }

// Blur marks blur state.
func (c *HistoryCell) Blur() { c.focused = false }

// Focused reports focus state.
func (c *HistoryCell) Focused() bool { return c.focused }

func (c *HistoryCell) renderMeta() string {
	parts := make([]string, 0, 3)
	if c.showTimestamp {
		parts = append(parts, c.timestamp.Format("15:04:05"))
	}
	if c.tokenCount > 0 {
		parts = append(parts, fmt.Sprintf("%dt", c.tokenCount))
	}
	if c.kind == HistoryCellToolUse || c.kind == HistoryCellToolResult {
		if c.toolName != "" {
			parts = append(parts, c.toolName)
		}
		if c.toolStatus != "" {
			parts = append(parts, c.toolStatus)
		}
	}

	line := strings.Join(parts, " • ")
	if c.designTokens == nil {
		return line
	}
	muted := style.Fg(c.designTokens.MutedColor)
	if muted == "" {
		return line
	}
	return muted + line + style.ANSIReset
}

func (c *HistoryCell) userBubble(meta, body string) string {
	accent := style.ANSICyan
	if c.designTokens != nil {
		if v := style.Fg(c.designTokens.Accent); v != "" {
			accent = v
		}
	}
	return meta + "\n" + accent + "╭─ user" + style.ANSIReset + "\n" + accent + "│ " + style.ANSIReset + body + "\n" + accent + "╰" + style.ANSIReset
}

func (c *HistoryCell) assistantText(meta, body string) string {
	color := ""
	if c.designTokens != nil {
		color = style.Fg(c.designTokens.Color)
	}
	return meta + "\n" + color + body + style.ANSIReset
}

func (c *HistoryCell) toolCall(meta, body string) string {
	pending := style.ANSIYellow
	if c.designTokens != nil {
		if v := style.Fg(c.designTokens.PendingColor); v != "" {
			pending = v
		}
	}
	return meta + "\n" + pending + "⧉ tool call" + style.ANSIReset + "\n" + body
}

func (c *HistoryCell) toolResult(meta, body string) string {
	statusColor := style.ANSIGreen
	if strings.EqualFold(c.toolStatus, "failed") || strings.EqualFold(c.toolStatus, "error") {
		statusColor = style.ANSIRed
	}
	if c.designTokens != nil {
		if strings.EqualFold(c.toolStatus, "failed") || strings.EqualFold(c.toolStatus, "error") {
			if v := style.Fg(c.designTokens.ErrorBright); v != "" {
				statusColor = v
			}
		} else if v := style.Fg(c.designTokens.SuccessBright); v != "" {
			statusColor = v
		}
	}
	return meta + "\n" + statusColor + "✓ tool result" + style.ANSIReset + "\n" + body
}

var _ tui.Component = (*HistoryCell)(nil)
