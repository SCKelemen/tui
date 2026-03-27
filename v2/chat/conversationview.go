package chat

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/selection"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// Spinner defines an animation sequence.
type Spinner struct {
	Frames []string
}

// SpinnerThinking - Codex/Claude-style thinking animation (small to large).
var SpinnerThinking = Spinner{
	Frames: []string{".", "+", "*", "÷", "•"},
}

// GetFrame returns the frame at the given index.
func (s Spinner) GetFrame(index int) string {
	if len(s.Frames) == 0 {
		return ""
	}
	return s.Frames[index%len(s.Frames)]
}

// MessageRole identifies who authored a message.
type MessageRole int

const (
	RoleUser MessageRole = iota
	RoleAssistant
	RoleSystem
)

// Message is a single conversation item.
type Message struct {
	ID        string
	Role      MessageRole
	Content   string
	Timestamp time.Time
	Streaming bool
}

// conversationTickMsg advances the streaming spinner.
type conversationTickMsg struct{}

// ConversationView renders a scrollable transcript.
type ConversationView struct {
	messages              []Message
	focused               bool
	scrollOffset          int
	autoScroll            bool
	designTokens          *design.DesignTokens
	spinner               Spinner
	spinIdx               int
	showTimestamps        bool
	width                 int
	height                int
	maxMessages           int
	selMgr                *selection.SelectionManager
	mouseSelectionEnabled bool
}

// ConversationViewOption configures a ConversationView.
type ConversationViewOption func(*ConversationView)

// WithConversationDesignTokens applies design-system tokens.
func WithConversationDesignTokens(tokens *design.DesignTokens) ConversationViewOption {
	return func(cv *ConversationView) {
		cv.designTokens = tokens
	}
}

// WithConversationTheme applies a named design-system theme.
func WithConversationTheme(theme string) ConversationViewOption {
	return func(cv *ConversationView) {
		cv.designTokens = style.DesignTokensForTheme(theme)
	}
}

// WithConversationSpinner sets the streaming spinner.
func WithConversationSpinner(spinner Spinner) ConversationViewOption {
	return func(cv *ConversationView) {
		cv.spinner = spinner
	}
}

// WithShowTimestamps controls timestamp visibility.
func WithShowTimestamps(show bool) ConversationViewOption {
	return func(cv *ConversationView) {
		cv.showTimestamps = show
	}
}

// WithConversationAutoScroll controls auto-scroll behavior.
func WithConversationAutoScroll(auto bool) ConversationViewOption {
	return func(cv *ConversationView) {
		cv.autoScroll = auto
	}
}

// WithConversationMouseSelection controls mouse text selection behavior.
func WithConversationMouseSelection(enabled bool) ConversationViewOption {
	return func(cv *ConversationView) {
		cv.mouseSelectionEnabled = enabled
	}
}

// WithMaxMessages limits retained messages (0 means unlimited).
func WithMaxMessages(max int) ConversationViewOption {
	return func(cv *ConversationView) {
		if max < 0 {
			max = 0
		}
		cv.maxMessages = max
	}
}

// NewConversationView creates a conversation component.
func NewConversationView(opts ...ConversationViewOption) *ConversationView {
	cv := &ConversationView{
		messages:       []Message{},
		autoScroll:     true,
		spinner:        SpinnerThinking,
		showTimestamps: true,
		selMgr:         selection.NewSelectionManager(),
	}
	for _, opt := range opts {
		opt(cv)
	}

	return cv
}

// Init initializes the component.
func (cv *ConversationView) Init() tea.Cmd {
	return cv.tick()
}

// Update handles Bubble Tea messages.
func (cv *ConversationView) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cv.width = msg.Width
		cv.height = msg.Height
		cv.clampScrollOffset()

	case conversationTickMsg:
		if cv.hasStreamingMessage() {
			cv.spinIdx++
		}
		return cv, cv.tick()

	case tea.MouseMsg:
		if cv.mouseSelectionEnabled && cv.selMgr != nil {
			cv.selMgr.HandleMouse(msg)
		}

	case tea.KeyMsg:
		if cv.mouseSelectionEnabled && cv.selMgr != nil && cv.selMgr.HasSelection() && msg.String() == "ctrl+c" {
			return cv, cv.selMgr.CopySelection()
		}

		if !cv.focused {
			return cv, nil
		}

		if cv.mouseSelectionEnabled && cv.selMgr != nil && cv.selMgr.HasSelection() && msg.String() == "y" {
			return cv, cv.selMgr.CopySelection()
		}

		switch msg.String() {
		case "j", "down":
			cv.scrollOffset++
			cv.clampScrollOffset()
		case "k", "up":
			cv.scrollOffset--
			if cv.scrollOffset < 0 {
				cv.scrollOffset = 0
			}
		case "g":
			cv.scrollOffset = 0
		case "G":
			cv.ScrollToBottom()
		}
	}

	return cv, nil
}

// View renders the conversation.
func (cv *ConversationView) View() string {
	if cv.width == 0 {
		return ""
	}

	lines, region := cv.renderLines(false)
	if len(lines) == 0 {
		return ""
	}

	if cv.mouseSelectionEnabled && cv.selMgr != nil {
		cv.selMgr.SetRegion(region)
		if cv.selMgr.HasSelection() {
			lines, _ = cv.renderLines(true)
		}
	}

	if cv.height <= 0 || len(lines) <= cv.height {
		return strings.Join(lines, "\n") + "\n"
	}

	cv.clampScrollOffset()
	end := cv.scrollOffset + cv.height
	if end > len(lines) {
		end = len(lines)
	}
	if cv.scrollOffset >= end {
		return ""
	}

	return strings.Join(lines[cv.scrollOffset:end], "\n") + "\n"
}

// Focus marks this component as focused.
func (cv *ConversationView) Focus() {
	cv.focused = true
}

// Blur marks this component as blurred.
func (cv *ConversationView) Blur() {
	cv.focused = false
}

// Focused reports focus state.
func (cv *ConversationView) Focused() bool {
	return cv.focused
}

// AddMessage appends a new message.
func (cv *ConversationView) AddMessage(msg Message) {
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	cv.messages = append(cv.messages, msg)
	cv.applyMessageLimit()

	if cv.autoScroll {
		cv.ScrollToBottom()
	} else {
		cv.clampScrollOffset()
	}
}

// AppendToLast appends text to the latest message.
func (cv *ConversationView) AppendToLast(content string) {
	if len(cv.messages) == 0 {
		return
	}

	cv.messages[len(cv.messages)-1].Content += content

	if cv.autoScroll {
		cv.ScrollToBottom()
	} else {
		cv.clampScrollOffset()
	}
}

// SetMessageStreaming updates streaming state for a message by ID.
func (cv *ConversationView) SetMessageStreaming(id string, streaming bool) {
	for i := range cv.messages {
		if cv.messages[i].ID == id {
			cv.messages[i].Streaming = streaming
			break
		}
	}
}

// GetMessages returns a copy of all messages.
func (cv *ConversationView) GetMessages() []Message {
	out := make([]Message, len(cv.messages))
	copy(out, cv.messages)
	return out
}

// Clear removes all messages and resets scroll.
func (cv *ConversationView) Clear() {
	cv.messages = []Message{}
	cv.scrollOffset = 0
}

// ScrollToBottom jumps to latest content.
func (cv *ConversationView) ScrollToBottom() {
	maxOffset := cv.maxScrollOffset()
	cv.scrollOffset = maxOffset
}

// MessageCount returns number of retained messages.
func (cv *ConversationView) MessageCount() int {
	return len(cv.messages)
}

// GetSelectionManager returns the conversation selection manager.
func (cv *ConversationView) GetSelectionManager() *selection.SelectionManager {
	return cv.selMgr
}

func (cv *ConversationView) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return conversationTickMsg{}
	})
}

func (cv *ConversationView) applyMessageLimit() {
	if cv.maxMessages <= 0 || len(cv.messages) <= cv.maxMessages {
		return
	}
	trim := len(cv.messages) - cv.maxMessages
	cv.messages = cv.messages[trim:]
}

func (cv *ConversationView) maxScrollOffset() int {
	if cv.height <= 0 {
		return 0
	}
	lines, _ := cv.renderLines(false)
	if len(lines) <= cv.height {
		return 0
	}
	return len(lines) - cv.height
}

func (cv *ConversationView) clampScrollOffset() {
	maxOffset := cv.maxScrollOffset()
	if cv.scrollOffset < 0 {
		cv.scrollOffset = 0
	}
	if cv.scrollOffset > maxOffset {
		cv.scrollOffset = maxOffset
	}
}

func (cv *ConversationView) hasStreamingMessage() bool {
	for _, msg := range cv.messages {
		if msg.Streaming {
			return true
		}
	}
	return false
}

func (cv *ConversationView) renderLines(applySelection bool) ([]string, selection.SelectableRegion) {
	all := make([]string, 0, len(cv.messages)*4)
	region := selection.SelectableRegion{GutterWidth: 2}
	contentRow := 0
	lineRow := 0
	regionInitialized := false

	for _, msg := range cv.messages {
		rendered, contentLines, startRow, endRow := cv.renderMessage(msg, applySelection, contentRow, lineRow)
		all = append(all, rendered...)
		region.ContentLines = append(region.ContentLines, contentLines...)

		if len(contentLines) > 0 {
			if !regionInitialized {
				region.StartRow = startRow
				regionInitialized = true
			}
			region.EndRow = endRow
		}

		contentRow += len(contentLines)
		lineRow += len(rendered)
	}

	if !regionInitialized {
		region.StartRow = 0
		region.EndRow = 0
	}

	return all, region
}

func (cv *ConversationView) renderMessage(msg Message, applySelection bool, contentRowStart, lineRowStart int) ([]string, []string, int, int) {
	innerWidth := cv.width - 2
	if innerWidth < 12 {
		innerWidth = 12
	}

	bodyWidth := innerWidth - 2
	if bodyWidth < 1 {
		bodyWidth = 1
	}

	left := cv.renderRole(msg)
	if msg.Streaming {
		left = left + " " + cv.spinner.GetFrame(cv.spinIdx)
	}

	right := ""
	if cv.showTimestamps {
		right = msg.Timestamp.Format("15:04")
	}

	header := cv.renderHeader(innerWidth, left, right)
	top := "┌" + header + "┐"

	wrapped := wrapConversationText(msg.Content, bodyWidth)
	contentLines := make([]string, 0, len(wrapped))
	body := make([]string, 0, len(wrapped))
	for i, line := range wrapped {
		contentLine := padRight(line, bodyWidth)
		contentLines = append(contentLines, contentLine)
		if applySelection && cv.mouseSelectionEnabled && cv.selMgr != nil && cv.selMgr.HasSelection() {
			contentLine = cv.selMgr.StyledLine(contentLine, contentRowStart+i)
		}
		body = append(body, "│ "+contentLine+" │")
	}

	bottom := "└" + strings.Repeat("─", innerWidth) + "┘"

	lines := []string{top}
	lines = append(lines, body...)
	lines = append(lines, bottom)

	startRow := lineRowStart + 1
	endRow := startRow + len(contentLines) - 1
	return lines, contentLines, startRow, endRow
}

func (cv *ConversationView) renderRole(msg Message) string {
	text := "assistant"
	color := cv.assistantColor()

	switch msg.Role {
	case RoleUser:
		text = "user"
		color = cv.userColor()
	case RoleAssistant:
		text = "assistant"
		color = cv.assistantColor()
	case RoleSystem:
		text = "system"
		color = cv.systemColor()
	}

	return color + text + style.ANSIReset
}

func (cv *ConversationView) renderHeader(innerWidth int, left, right string) string {
	if right == "" {
		base := left + " "
		remaining := innerWidth - len(stripANSI(base))
		if remaining < 0 {
			remaining = 0
		}
		return base + strings.Repeat("─", remaining)
	}

	base := left + " "
	remaining := innerWidth - len(stripANSI(base)) - len(right) - 1
	if remaining < 1 {
		remaining = 1
	}
	return base + strings.Repeat("─", remaining) + " " + right
}

func (cv *ConversationView) userColor() string {
	if cv.designTokens == nil {
		return style.ANSICyan
	}
	accent := style.ANSIColorFromHex(cv.designTokens.Accent)
	if accent == "" {
		return style.ANSICyan
	}
	return accent
}

func (cv *ConversationView) assistantColor() string {
	return style.ANSIGreen
}

func (cv *ConversationView) systemColor() string {
	if cv.designTokens != nil {
		if fg := style.ANSIColorFromHex(cv.designTokens.Color); fg != "" {
			return fg
		}
	}
	return style.ANSIDim + style.ANSIWhite
}

func wrapConversationText(content string, width int) []string {
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
		for len(line) > width {
			cut := width
			spaceIdx := strings.LastIndex(line[:cut], " ")
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

func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// Ensure compile-time interface compliance.
var _ tui.Component = (*ConversationView)(nil)

func (r MessageRole) String() string {
	switch r {
	case RoleUser:
		return "user"
	case RoleAssistant:
		return "assistant"
	case RoleSystem:
		return "system"
	default:
		return fmt.Sprintf("role(%d)", int(r))
	}
}
