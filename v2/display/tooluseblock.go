package display

import (
	"fmt"
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ToolStatus represents tool execution state.
type ToolStatus int

const (
	ToolStatusPending ToolStatus = iota
	ToolStatusRunning
	ToolStatusSuccess
	ToolStatusError
	ToolStatusCanceled
	ToolStatusAwaitingPermission
)

// ToolUseBlock renders a single tool invocation with status and output.
type ToolUseBlock struct {
	toolName string
	status   ToolStatus

	args     string
	result   string
	errText  string
	duration time.Duration

	collapsed bool
	width     int

	windowWidth int
	focused     bool

	designTokens *design.DesignTokens

	successColor    string
	errorColor      string
	runningColor    string
	pendingColor    string
	canceledColor   string
	permissionColor string
	durationColor   string
}

// ToolUseBlockOption configures a ToolUseBlock.
type ToolUseBlockOption func(*ToolUseBlock)

// WithToolUseBlockStatus sets the current status.
func WithToolUseBlockStatus(status ToolStatus) ToolUseBlockOption {
	return func(b *ToolUseBlock) {
		b.status = status
	}
}

// WithToolUseBlockArgs sets args preview text.
func WithToolUseBlockArgs(args string) ToolUseBlockOption {
	return func(b *ToolUseBlock) {
		b.args = strings.TrimSpace(args)
	}
}

// WithToolUseBlockResult sets result text.
func WithToolUseBlockResult(result string) ToolUseBlockOption {
	return func(b *ToolUseBlock) {
		b.result = result
	}
}

// WithToolUseBlockError sets error text.
func WithToolUseBlockError(err string) ToolUseBlockOption {
	return func(b *ToolUseBlock) {
		b.errText = err
	}
}

// WithToolUseBlockDuration sets execution duration.
func WithToolUseBlockDuration(d time.Duration) ToolUseBlockOption {
	return func(b *ToolUseBlock) {
		if d < 0 {
			d = 0
		}
		b.duration = d
	}
}

// WithToolUseBlockCollapsed sets collapsed/expanded state.
func WithToolUseBlockCollapsed(collapsed bool) ToolUseBlockOption {
	return func(b *ToolUseBlock) {
		b.collapsed = collapsed
	}
}

// WithToolUseBlockWidth sets block width.
func WithToolUseBlockWidth(width int) ToolUseBlockOption {
	return func(b *ToolUseBlock) {
		if width >= 0 {
			b.width = width
		}
	}
}

// WithToolUseBlockDesignTokens applies design-system tokens.
func WithToolUseBlockDesignTokens(tokens *design.DesignTokens) ToolUseBlockOption {
	return func(b *ToolUseBlock) {
		b.designTokens = tokens
		b.applyDesignTokens(tokens)
	}
}

// NewToolUseBlock creates a ToolUseBlock.
func NewToolUseBlock(toolName string, opts ...ToolUseBlockOption) *ToolUseBlock {
	b := &ToolUseBlock{
		toolName:         strings.TrimSpace(toolName),
		status:           ToolStatusPending,
		collapsed:        false,
		width:            0,
		successColor:     style.ANSIGreen,
		errorColor:       style.ANSIRed,
		runningColor:     style.ANSICyan,
		pendingColor:     style.ANSIDim,
		canceledColor:    style.ANSIDim,
		permissionColor:  style.ANSIYellow,
		durationColor:    style.ANSIDim,
		designTokens:     nil,
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// Init initializes the component.
func (b *ToolUseBlock) Init() tea.Cmd {
	return nil
}

// Update handles resize and collapse toggle keys.
func (b *ToolUseBlock) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.windowWidth = msg.Width
	case tea.KeyMsg:
		if !b.focused {
			return b, nil
		}
		switch msg.String() {
		case "enter", " ":
			b.collapsed = !b.collapsed
		}
	}

	return b, nil
}

// View renders header and, when expanded, an indented result/error body.
func (b *ToolUseBlock) View() string {
	if b.toolName == "" {
		return ""
	}

	width := b.renderWidth()
	icon, plainIcon := b.renderStatusIcon()
	durationText := formatToolUseDuration(b.duration)

	argsSuffix := b.renderArgsSuffix(width, plainIcon, durationText)
	leftPlain := fmt.Sprintf("[%s] %s%s", plainIcon, b.toolName, argsSuffix)
	leftColored := fmt.Sprintf("[%s] %s%s", icon, b.toolName, argsSuffix)

	header := leftColored
	if width > 0 && durationText != "" {
		gap := width - style.StringWidth(leftPlain) - style.StringWidth(durationText)
		if gap < 1 {
			gap = 1
		}
		header = leftColored + strings.Repeat(" ", gap) + b.durationColor + durationText + style.ANSIReset
	} else if durationText != "" {
		header = leftColored + "  " + b.durationColor + durationText + style.ANSIReset
	}

	if width > 0 {
		header = b.fitToWidth(header, width)
	}

	if b.collapsed {
		return header + "\n"
	}

	body := b.bodyText()
	if strings.TrimSpace(body) == "" {
		return header + "\n"
	}

	bodyColor := b.bodyColor()
	bodyLines := strings.Split(body, "\n")
	renderedBody := make([]string, 0, len(bodyLines))
	for _, line := range bodyLines {
		rendered := "  " + bodyColor + line + style.ANSIReset
		if width > 0 {
			rendered = b.fitToWidth(rendered, width)
		}
		renderedBody = append(renderedBody, rendered)
	}

	return header + "\n" + strings.Join(renderedBody, "\n") + "\n"
}

// Focus marks component focused.
func (b *ToolUseBlock) Focus() {
	b.focused = true
}

// Blur marks component unfocused.
func (b *ToolUseBlock) Blur() {
	b.focused = false
}

// Focused reports focus state.
func (b *ToolUseBlock) Focused() bool {
	return b.focused
}

// SetStatus updates status.
func (b *ToolUseBlock) SetStatus(status ToolStatus) {
	b.status = status
}

// ToggleCollapsed toggles collapsed state.
func (b *ToolUseBlock) ToggleCollapsed() {
	b.collapsed = !b.collapsed
}

func (b *ToolUseBlock) renderStatusIcon() (string, string) {
	switch b.status {
	case ToolStatusPending:
		return b.pendingColor + "○" + style.ANSIReset, "○"
	case ToolStatusRunning:
		return b.runningColor + "◉" + style.ANSIReset, "◉"
	case ToolStatusSuccess:
		return b.successColor + "✓" + style.ANSIReset, "✓"
	case ToolStatusError:
		return b.errorColor + "✗" + style.ANSIReset, "✗"
	case ToolStatusCanceled:
		return b.canceledColor + "⊘" + style.ANSIReset, "⊘"
	case ToolStatusAwaitingPermission:
		return b.permissionColor + "⚑" + style.ANSIReset, "⚑"
	default:
		return b.pendingColor + "○" + style.ANSIReset, "○"
	}
}

func (b *ToolUseBlock) renderArgsSuffix(width int, iconPlain, durationText string) string {
	args := strings.TrimSpace(strings.ReplaceAll(b.args, "\n", " "))
	if args == "" {
		return ""
	}

	if width <= 0 {
		return " (" + args + ")"
	}

	available := width - style.StringWidth(iconPlain) - style.StringWidth(b.toolName) - 6
	if durationText != "" {
		available -= style.StringWidth(durationText) + 2
	}
	if available < 0 {
		available = 0
	}

	preview := style.Truncate(args, available, "…")
	if preview == "" {
		return ""
	}
	return " (" + preview + ")"
}

func (b *ToolUseBlock) bodyText() string {
	if b.status == ToolStatusError && strings.TrimSpace(b.errText) != "" {
		return b.errText
	}
	if strings.TrimSpace(b.result) != "" {
		return b.result
	}
	if strings.TrimSpace(b.errText) != "" {
		return b.errText
	}
	return ""
}

func (b *ToolUseBlock) bodyColor() string {
	switch b.status {
	case ToolStatusSuccess:
		return b.successColor
	case ToolStatusError:
		return b.errorColor
	case ToolStatusRunning:
		return b.runningColor
	case ToolStatusAwaitingPermission:
		return b.permissionColor
	default:
		return b.pendingColor
	}
}

func (b *ToolUseBlock) renderWidth() int {
	if b.width > 0 {
		return b.width
	}
	if b.windowWidth > 0 {
		return b.windowWidth
	}
	return 0
}

func (b *ToolUseBlock) fitToWidth(line string, width int) string {
	if width <= 0 {
		return line
	}
	plain := stripANSI(line)
	w := style.StringWidth(plain)
	if w > width {
		return truncateANSI(line, width)
	}
	if w < width {
		return line + strings.Repeat(" ", width-w)
	}
	return line
}

func (b *ToolUseBlock) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}
	if v := strings.TrimSpace(tokens.SuccessBright); v != "" {
		b.successColor = style.Fg(v)
	}
	if v := strings.TrimSpace(tokens.ErrorBright); v != "" {
		b.errorColor = style.Fg(v)
	}
	if v := strings.TrimSpace(tokens.Accent); v != "" {
		b.runningColor = style.Fg(v)
		b.permissionColor = style.Fg(v)
	}
	if v := strings.TrimSpace(tokens.MutedColor); v != "" {
		b.pendingColor = style.Fg(v)
		b.canceledColor = style.Fg(v)
		b.durationColor = style.Fg(v)
	}
}

func formatToolUseDuration(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	if d < time.Second {
		ms := d.Milliseconds()
		if ms < 1 {
			ms = 1
		}
		return fmt.Sprintf("%dms", ms)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d / time.Minute)
	seconds := int((d % time.Minute) / time.Second)
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

var _ tui.Component = (*ToolUseBlock)(nil)
