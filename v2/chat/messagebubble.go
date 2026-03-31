package chat

import (
	"regexp"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var messageBubbleANSIRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// Backward-compatible aliases for message roles used by MessageBubble.
const (
	MessageRoleUser      MessageRole = RoleUser
	MessageRoleAssistant MessageRole = RoleAssistant
	MessageRoleSystem    MessageRole = RoleSystem
	MessageRoleTool      MessageRole = 3
)

// MessageBubble renders a single chat message with role-aware styling.
type MessageBubble struct {
	role        MessageRole
	content     string
	maxWidth    int
	windowWidth int
	timestamp   string
	author      string
	hasError    bool
	streaming   bool
	focused     bool

	designTokens *design.DesignTokens
}

// MessageBubbleOption configures MessageBubble.
type MessageBubbleOption func(*MessageBubble)

// NewMessageBubble creates a new MessageBubble.
func NewMessageBubble(role MessageRole, content string, opts ...MessageBubbleOption) *MessageBubble {
	mb := &MessageBubble{
		role:     role,
		content:  content,
		maxWidth: 72,
	}

	for _, opt := range opts {
		opt(mb)
	}

	return mb
}

// WithMessageBubbleWidth sets the maximum rendered width for the bubble.
func WithMessageBubbleWidth(width int) MessageBubbleOption {
	return func(mb *MessageBubble) {
		if width > 0 {
			mb.maxWidth = width
		}
	}
}

// WithMessageBubbleTimestamp sets the timestamp text.
func WithMessageBubbleTimestamp(ts string) MessageBubbleOption {
	return func(mb *MessageBubble) {
		mb.timestamp = strings.TrimSpace(ts)
	}
}

// WithMessageBubbleAuthor sets the displayed author name.
func WithMessageBubbleAuthor(author string) MessageBubbleOption {
	return func(mb *MessageBubble) {
		mb.author = strings.TrimSpace(author)
	}
}

// WithMessageBubbleError toggles error styling.
func WithMessageBubbleError(hasError bool) MessageBubbleOption {
	return func(mb *MessageBubble) {
		mb.hasError = hasError
	}
}

// WithMessageBubbleStreaming toggles streaming cursor rendering.
func WithMessageBubbleStreaming(streaming bool) MessageBubbleOption {
	return func(mb *MessageBubble) {
		mb.streaming = streaming
	}
}

// WithMessageBubbleDesignTokens applies design-system tokens.
func WithMessageBubbleDesignTokens(tokens *design.DesignTokens) MessageBubbleOption {
	return func(mb *MessageBubble) {
		mb.designTokens = tokens
	}
}

// Init initializes the component.
func (mb *MessageBubble) Init() tea.Cmd { return nil }

// Update handles Bubble Tea messages.
func (mb *MessageBubble) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		mb.windowWidth = msg.Width
	}
	return mb, nil
}

// View renders the bubble.
func (mb *MessageBubble) View() string {
	width := mb.availableWidth()
	if width < 12 {
		width = 12
	}

	bodyWidth := width - 3 // border + one space + content
	if bodyWidth < 1 {
		bodyWidth = 1
	}

	header := mb.headerText()
	if mb.hasError {
		header = "⚠ " + header
	}

	content := mb.content
	if mb.streaming {
		content += "█"
	}

	lines := []string{mb.renderLine(style.Truncate(header, bodyWidth, "…"), true, bodyWidth)}
	for _, line := range wrapMessageBubbleText(content, bodyWidth) {
		lines = append(lines, mb.renderLine(line, false, bodyWidth))
	}

	return strings.Join(lines, "\n") + "\n"
}

// Focus marks this component focused.
func (mb *MessageBubble) Focus() { mb.focused = true }

// Blur marks this component blurred.
func (mb *MessageBubble) Blur() { mb.focused = false }

// Focused reports focus state.
func (mb *MessageBubble) Focused() bool { return mb.focused }

func (mb *MessageBubble) availableWidth() int {
	width := mb.maxWidth
	if mb.windowWidth > 0 && mb.windowWidth < width {
		width = mb.windowWidth
	}
	if width <= 0 {
		return 72
	}
	return width
}

func (mb *MessageBubble) headerText() string {
	author := mb.author
	if author == "" {
		author = mb.defaultAuthor()
	}
	if mb.timestamp == "" {
		return author
	}
	return author + " · " + mb.timestamp
}

func (mb *MessageBubble) defaultAuthor() string {
	switch mb.role {
	case MessageRoleUser:
		return "User"
	case MessageRoleAssistant:
		return "Assistant"
	case MessageRoleSystem:
		return "System"
	case MessageRoleTool:
		return "Tool"
	default:
		return "Message"
	}
}

func (mb *MessageBubble) borderColor() string {
	if mb.hasError {
		return style.ANSIRed
	}

	switch mb.role {
	case MessageRoleUser:
		if mb.designTokens != nil {
			if c := style.ANSIColorFromHex(mb.designTokens.Accent); c != "" {
				return c
			}
		}
		return style.ANSICyan
	case MessageRoleAssistant:
		return style.ANSIGreen
	case MessageRoleSystem:
		return style.ANSIDim + style.ANSIWhite
	case MessageRoleTool:
		return style.ANSIBlue
	default:
		return style.ANSIDim + style.ANSIWhite
	}
}

func (mb *MessageBubble) headerColor() string {
	if mb.hasError {
		return style.ANSIRed
	}
	return style.ANSIDim + style.ANSIWhite
}

func (mb *MessageBubble) contentColor() string {
	if mb.hasError {
		return style.ANSIRed
	}
	if mb.role == MessageRoleSystem {
		return style.ANSIDim
	}
	return ""
}

func (mb *MessageBubble) renderLine(content string, isHeader bool, width int) string {
	line := style.Pad(style.Truncate(content, width, ""), width)

	prefix := mb.borderColor() + "│" + style.ANSIReset + " "
	if isHeader {
		return prefix + mb.headerColor() + line + style.ANSIReset
	}

	if color := mb.contentColor(); color != "" {
		return prefix + color + line + style.ANSIReset
	}
	return prefix + line
}

func wrapMessageBubbleText(content string, width int) []string {
	if width <= 0 {
		return []string{""}
	}
	if content == "" {
		return []string{""}
	}

	paragraphs := strings.Split(content, "\n")
	out := make([]string, 0, len(paragraphs))

	for _, p := range paragraphs {
		if p == "" {
			out = append(out, "")
			continue
		}

		line := p
		for lipgloss.Width(line) > width {
			cut := len(line)
			for i := range line {
				segment := line[:i+1]
				if lipgloss.Width(segment) > width {
					cut = i
					break
				}
			}

			if cut <= 0 || cut >= len(line) {
				break
			}

			window := line[:cut]
			spaceIdx := strings.LastIndex(window, " ")
			if spaceIdx > 0 {
				cut = spaceIdx
			}

			out = append(out, strings.TrimSpace(line[:cut]))
			line = strings.TrimLeft(line[cut:], " ")
		}
		out = append(out, line)
	}

	if len(out) == 0 {
		return []string{""}
	}
	return out
}

func stripMessageBubbleANSI(s string) string {
	return messageBubbleANSIRegex.ReplaceAllString(s, "")
}

var _ tui.Component = (*MessageBubble)(nil)
