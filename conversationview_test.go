package tui

import (
	"strings"
	"testing"
	"time"

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
	if cv.spinner.FrameCount() != SpinnerThinking.FrameCount() {
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
	cv := NewConversationView(WithConversationSpinner(SpinnerDots), WithShowTimestamps(false))
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
	if !strings.Contains(view, SpinnerDots.GetFrame(0)) {
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