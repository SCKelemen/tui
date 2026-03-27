package chat

import (
	"strings"
	"testing"
	"time"

	design "github.com/SCKelemen/design-system"
	tea "github.com/charmbracelet/bubbletea"
)
func TestConversationViewCreation(t *testing.T) {
	cv := NewConversationView()
	if cv == nil {
		t.Fatal("NewConversationView returned nil")
	}

	if !cv.autoScroll {
		t.Error("autoScroll should default to true")
	}
	if !cv.showTimestamps {
		t.Error("showTimestamps should default to true")
	}
	if cv.maxMessages != 0 {
		t.Errorf("maxMessages should default to 0, got %d", cv.maxMessages)
	}
	if len(cv.spinner.Frames) != len(SpinnerThinking.Frames) {
		t.Error("default spinner should be SpinnerThinking")
	}
}

func TestConversationViewAddMessages(t *testing.T) {
	cv := NewConversationView()
	ts := time.Date(2026, 1, 1, 14, 32, 0, 0, time.UTC)

	cv.AddMessage(Message{ID: "1", Role: RoleUser, Content: "hi", Timestamp: ts})
	cv.AddMessage(Message{ID: "2", Role: RoleAssistant, Content: "hello", Timestamp: ts})

	if cv.MessageCount() != 2 {
		t.Fatalf("expected 2 messages, got %d", cv.MessageCount())
	}

	msgs := cv.GetMessages()
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages from GetMessages, got %d", len(msgs))
	}
	if msgs[0].ID != "1" || msgs[1].ID != "2" {
		t.Fatalf("unexpected message IDs: %+v", msgs)
	}
}

func TestConversationViewStreamingAppend(t *testing.T) {
	cv := NewConversationView()
	cv.AddMessage(Message{ID: "stream-1", Role: RoleAssistant, Content: "I will", Streaming: true, Timestamp: time.Now()})
	cv.AppendToLast(" investigate")

	msgs := cv.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Content != "I will investigate" {
		t.Fatalf("unexpected appended content: %q", msgs[0].Content)
	}

	cv.SetMessageStreaming("stream-1", false)
	msgs = cv.GetMessages()
	if msgs[0].Streaming {
		t.Fatal("streaming should be false after SetMessageStreaming")
	}
}

func TestConversationViewScrollBehavior(t *testing.T) {
	cv := NewConversationView()
	cv.Focus()
	cv.Update(tea.WindowSizeMsg{Width: 40, Height: 6})

	for i := 0; i < 4; i++ {
		cv.AddMessage(Message{ID: string(rune('a' + i)), Role: RoleUser, Content: "hello", Timestamp: time.Now()})
	}

	if cv.scrollOffset == 0 {
		t.Fatal("expected scrollOffset > 0 after auto-scroll with overflowing content")
	}

	bottom := cv.scrollOffset
	cv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if cv.scrollOffset >= bottom {
		t.Fatal("expected up key to move scroll up")
	}

	cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if cv.scrollOffset != cv.maxScrollOffset() {
		t.Fatal("expected G to jump to bottom")
	}

	cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if cv.scrollOffset != 0 {
		t.Fatal("expected g to jump to top")
	}
}

func TestConversationViewRoleRendering(t *testing.T) {
	testSpinner := Spinner{Frames: []string{".", "o", "O", "o"}}
	cv := NewConversationView(WithConversationSpinner(testSpinner), WithShowTimestamps(false))
	cv.Update(tea.WindowSizeMsg{Width: 50, Height: 30})
	cv.AddMessage(Message{ID: "u", Role: RoleUser, Content: "Can you fix this?", Timestamp: time.Now()})
	cv.AddMessage(Message{ID: "a", Role: RoleAssistant, Content: "Working on it", Streaming: true, Timestamp: time.Now()})
	cv.AddMessage(Message{ID: "s", Role: RoleSystem, Content: "Tool: grep auth", Timestamp: time.Now()})

	view := cv.View()
	if !strings.Contains(view, "user") {
		t.Fatal("view should contain user role")
	}
	if !strings.Contains(view, "assistant") {
		t.Fatal("view should contain assistant role")
	}
	if !strings.Contains(view, "system") {
		t.Fatal("view should contain system role")
	}
	if !strings.Contains(view, testSpinner.GetFrame(0)) {
		t.Fatal("streaming assistant message should include spinner frame")
	}
}
func TestConversationViewTimestampDisplay(t *testing.T) {
	ts := time.Date(2026, 1, 1, 14, 32, 0, 0, time.UTC)

	withTS := NewConversationView()
	withTS.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	withTS.AddMessage(Message{ID: "1", Role: RoleUser, Content: "x", Timestamp: ts})
	if !strings.Contains(withTS.View(), "14:32") {
		t.Fatal("expected timestamp in view when enabled")
	}

	withoutTS := NewConversationView(WithShowTimestamps(false))
	withoutTS.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	withoutTS.AddMessage(Message{ID: "1", Role: RoleUser, Content: "x", Timestamp: ts})
	if strings.Contains(withoutTS.View(), "14:32") {
		t.Fatal("did not expect timestamp in view when disabled")
	}
}

func TestConversationViewAutoScrollOption(t *testing.T) {
	cv := NewConversationView(WithConversationAutoScroll(false))
	cv.Update(tea.WindowSizeMsg{Width: 40, Height: 6})
	for i := 0; i < 4; i++ {
		cv.AddMessage(Message{ID: string(rune('a' + i)), Role: RoleAssistant, Content: "line", Timestamp: time.Now()})
	}

	if cv.scrollOffset != 0 {
		t.Fatalf("expected scrollOffset to remain 0 when auto-scroll disabled, got %d", cv.scrollOffset)
	}
}

func TestConversationViewMessageLimit(t *testing.T) {
	cv := NewConversationView(WithMaxMessages(2))

	cv.AddMessage(Message{ID: "1", Role: RoleUser, Content: "one", Timestamp: time.Now()})
	cv.AddMessage(Message{ID: "2", Role: RoleAssistant, Content: "two", Timestamp: time.Now()})
	cv.AddMessage(Message{ID: "3", Role: RoleSystem, Content: "three", Timestamp: time.Now()})

	if cv.MessageCount() != 2 {
		t.Fatalf("expected message count 2, got %d", cv.MessageCount())
	}

	msgs := cv.GetMessages()
	if msgs[0].ID != "2" || msgs[1].ID != "3" {
		t.Fatalf("expected oldest message to be dropped, got IDs %q and %q", msgs[0].ID, msgs[1].ID)
	}
}

func TestConversationViewMouseSelectionOption(t *testing.T) {
	disabled := NewConversationView()
	if disabled.mouseSelectionEnabled {
		t.Fatal("mouse selection should be disabled by default")
	}

	enabled := NewConversationView(WithConversationMouseSelection(true))
	if !enabled.mouseSelectionEnabled {
		t.Fatal("mouse selection should be enabled by option")
	}
}

func TestConversationViewGetSelectionManager(t *testing.T) {
	cv := NewConversationView()

	selMgr := cv.GetSelectionManager()
	if selMgr == nil {
		t.Fatal("expected non-nil selection manager")
	}
	if selMgr != cv.selMgr {
		t.Fatal("expected accessor to return internal selection manager")
	}
}

func TestConversationViewOptionsAndFocusLifecycle(t *testing.T) {
	cv := NewConversationView(
		WithConversationDesignTokens(design.NordTheme()),
		WithConversationTheme("paper"),
		WithMaxMessages(-5),
	)

	if cv.designTokens == nil {
		t.Fatal("expected design tokens from theme option")
	}
	if cv.maxMessages != 0 {
		t.Fatalf("expected negative max to clamp to 0, got %d", cv.maxMessages)
	}
	if cv.Focused() {
		t.Fatal("new view should not be focused")
	}

	cv.Focus()
	if !cv.Focused() {
		t.Fatal("view should be focused after Focus")
	}

	cv.Blur()
	if cv.Focused() {
		t.Fatal("view should not be focused after Blur")
	}
}

func TestConversationViewInitAndTickBehavior(t *testing.T) {
	cv := NewConversationView()
	cmd := cv.Init()
	if cmd == nil {
		t.Fatal("Init should return non-nil command")
	}

	msg := cmd()
	if _, ok := msg.(conversationTickMsg); !ok {
		t.Fatalf("expected conversationTickMsg, got %T", msg)
	}

	before := cv.spinIdx
	_, nextCmd := cv.Update(msg)
	if nextCmd == nil {
		t.Fatal("tick update should return next tick command")
	}
	if cv.spinIdx != before {
		t.Fatal("spin index should not advance without streaming message")
	}

	cv.AddMessage(Message{ID: "stream", Role: RoleAssistant, Content: "thinking", Streaming: true, Timestamp: time.Now()})
	before = cv.spinIdx
	_, nextCmd = cv.Update(conversationTickMsg{})
	if nextCmd == nil {
		t.Fatal("tick update should continue scheduling")
	}
	if cv.spinIdx != before+1 {
		t.Fatalf("expected spin index increment by 1, got %d -> %d", before, cv.spinIdx)
	}
}

func TestConversationViewViewEdgeCases(t *testing.T) {
	cv := NewConversationView()
	if got := cv.View(); got != "" {
		t.Fatalf("expected empty view at zero width, got %q", got)
	}

	cv.Update(tea.WindowSizeMsg{Width: 30, Height: 5})
	if got := cv.View(); got != "" {
		t.Fatalf("expected empty view with no messages, got %q", got)
	}

	cv.AddMessage(Message{ID: "1", Role: RoleUser, Content: "", Timestamp: time.Now()})
	if got := cv.View(); got == "" {
		t.Fatal("expected rendered frame for empty message content")
	}

	cv.Update(tea.WindowSizeMsg{Width: 0, Height: 5})
	if got := cv.View(); got != "" {
		t.Fatalf("expected empty view at zero width, got %q", got)
	}
}

func TestConversationViewViewHeightAndLongWrap(t *testing.T) {
	cv := NewConversationView(WithShowTimestamps(false))
	cv.Update(tea.WindowSizeMsg{Width: 20, Height: 0})
	cv.AddMessage(Message{ID: "1", Role: RoleAssistant, Content: "this is a very long message that should wrap across multiple lines", Timestamp: time.Now()})

	full := cv.View()
	if !strings.Contains(full, "assistant") {
		t.Fatal("expected role in rendered view")
	}
	if !strings.Contains(full, "this is a") {
		t.Fatal("expected wrapped content in rendered view")
	}

	cv.Update(tea.WindowSizeMsg{Width: 20, Height: 2})
	truncated := cv.View()
	if truncated == "" {
		t.Fatal("expected non-empty truncated view")
	}
}

func TestConversationViewUpdateWindowAndKeyControls(t *testing.T) {
	cv := NewConversationView()
	cv.Update(tea.WindowSizeMsg{Width: 30, Height: 4})
	cv.Focus()

	for i := 0; i < 5; i++ {
		cv.AddMessage(Message{ID: string(rune('a' + i)), Role: RoleUser, Content: "line", Timestamp: time.Now()})
	}

	bottom := cv.maxScrollOffset()
	if cv.scrollOffset != bottom {
		t.Fatalf("expected auto-scroll to bottom %d, got %d", bottom, cv.scrollOffset)
	}

	cv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cv.scrollOffset != bottom {
		t.Fatal("down at bottom should remain clamped")
	}

	cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if cv.scrollOffset >= bottom {
		t.Fatal("k should scroll up")
	}

	cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if cv.scrollOffset != bottom {
		t.Fatal("j should scroll down back to bottom")
	}

	cv.Blur()
	before := cv.scrollOffset
	cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if cv.scrollOffset != before {
		t.Fatal("blurred view should ignore scroll keys")
	}
}

func TestConversationViewMouseSelectionAndCopyPaths(t *testing.T) {
	cv := NewConversationView(WithConversationMouseSelection(true), WithShowTimestamps(false))
	cv.Update(tea.WindowSizeMsg{Width: 40, Height: 8})
	cv.AddMessage(Message{ID: "1", Role: RoleAssistant, Content: "alpha bravo", Timestamp: time.Now()})

	_ = cv.View()

	cv.Update(tea.MouseMsg{X: 2, Y: 1, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	cv.Update(tea.MouseMsg{X: 7, Y: 1, Action: tea.MouseActionMotion, Button: tea.MouseButtonLeft})
	cv.Update(tea.MouseMsg{X: 7, Y: 1, Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft})

	if !cv.selMgr.HasSelection() {
		t.Fatal("expected finalized selection after mouse drag")
	}

	_, cmd := cv.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("ctrl+c should return copy command when selection exists")
	}

	cv.Focus()
	_, cmd = cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("y should return copy command when focused and selection exists")
	}

	selectedView := cv.View()
	if !strings.Contains(selectedView, "\x1b[7m") {
		t.Fatal("expected selection highlighting to be applied in view")
	}
}

func TestConversationViewClearAndColorFallbacks(t *testing.T) {
	cv := NewConversationView()
	cv.Update(tea.WindowSizeMsg{Width: 32, Height: 6})
	cv.AddMessage(Message{ID: "1", Role: RoleSystem, Content: "notice", Timestamp: time.Now()})
	cv.AppendToLast(" now")
	cv.AppendToLast("")

	if cv.MessageCount() != 1 {
		t.Fatalf("expected 1 message before clear, got %d", cv.MessageCount())
	}

	cv.Clear()
	if cv.MessageCount() != 0 {
		t.Fatalf("expected 0 messages after clear, got %d", cv.MessageCount())
	}
	if cv.scrollOffset != 0 {
		t.Fatalf("expected scroll offset reset after clear, got %d", cv.scrollOffset)
	}

	cv.AppendToLast("ignored")
	if cv.MessageCount() != 0 {
		t.Fatal("append on empty conversation should do nothing")
	}

	tokens := &design.DesignTokens{Accent: "not-a-hex", Color: "not-a-hex"}
	colored := NewConversationView(WithConversationDesignTokens(tokens))
	if got := colored.userColor(); got == "" {
		t.Fatal("user color should always have fallback")
	}
	if got := colored.systemColor(); got == "" {
		t.Fatal("system color should always have fallback")
	}
}