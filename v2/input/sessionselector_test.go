package input

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func testSessions() []SessionItem {
	now := time.Now()
	return []SessionItem{
		{ID: "s1", Name: "Active session", Status: SessionActive, CreatedAt: now.Add(-2 * time.Minute), MessageCount: 3},
		{ID: "s2", Name: "Draft session", Status: SessionDraft, CreatedAt: now.Add(-1 * time.Hour), MessageCount: 1, IsChild: true},
		{ID: "s3", Name: "Completed session", Status: SessionCompleted, CreatedAt: now.Add(-24 * time.Hour), MessageCount: 5},
	}
}

func TestSessionSelectorConstructor(t *testing.T) {
	s := NewSessionSelector(testSessions(), WithSessionSelectorSelected(1), WithSessionSelectorWidth(80))
	if s == nil {
		t.Fatal("NewSessionSelector returned nil")
	}

	if s.cursor != 1 {
		t.Fatalf("expected cursor 1, got %d", s.cursor)
	}
}

func TestSessionSelectorSessionsListAndStatusRendering(t *testing.T) {
	s := NewSessionSelector(testSessions())
	view := s.View()
	if view == "" {
		t.Fatal("view should not be empty with sessions")
	}

	if !strings.Contains(view, "Active session") || !strings.Contains(view, "Draft session") {
		t.Fatal("view should include session names")
	}

	if !strings.Contains(view, "▶") || !strings.Contains(view, "✓") {
		t.Fatal("view should render status markers for active/completed")
	}
}

func TestSessionSelectorSelection(t *testing.T) {
	s := NewSessionSelector(testSessions())
	s.Focus()

	s.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command on enter")
	}

	msg := cmd()
	selectedMsg, ok := msg.(SessionSelectedMsg)
	if !ok {
		t.Fatalf("expected SessionSelectedMsg, got %T", msg)
	}

	if selectedMsg.ID != "s2" {
		t.Fatalf("expected selected session id s2, got %q", selectedMsg.ID)
	}
}
