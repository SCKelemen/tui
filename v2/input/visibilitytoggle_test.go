package input

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestVisibilityToggleConstructor(t *testing.T) {
	v := NewVisibilityToggle("Details", "hello world")
	if v == nil {
		t.Fatal("NewVisibilityToggle returned nil")
	}

	if v.expanded {
		t.Fatal("toggle should be collapsed by default")
	}
}

func TestVisibilityToggleExpandCollapse(t *testing.T) {
	v := NewVisibilityToggle("Details", "line1\nline2")
	v.Focus()

	_, cmd := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command when toggling")
	}
	msg := cmd()
	toggleMsg, ok := msg.(VisibilityToggleMsg)
	if !ok {
		t.Fatalf("expected VisibilityToggleMsg, got %T", msg)
	}
	if !toggleMsg.Visible {
		t.Fatal("expected expanded visibility true after first toggle")
	}

	_, cmd = v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command when toggling second time")
	}
	msg = cmd()
	toggleMsg, ok = msg.(VisibilityToggleMsg)
	if !ok {
		t.Fatalf("expected VisibilityToggleMsg, got %T", msg)
	}
	if toggleMsg.Visible {
		t.Fatal("expected collapsed visibility false after second toggle")
	}
}

func TestVisibilityToggleViewRendersIcons(t *testing.T) {
	v := NewVisibilityToggle("Details", "body")

	collapsed := v.View()
	if !strings.Contains(collapsed, "▶") {
		t.Fatal("collapsed view should render ▶")
	}

	v.Focus()
	v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	expanded := v.View()
	if !strings.Contains(expanded, "▼") {
		t.Fatal("expanded view should render ▼")
	}
	if !strings.Contains(expanded, "body") {
		t.Fatal("expanded view should render content")
	}
}
