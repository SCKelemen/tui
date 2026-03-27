package input

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestKeyOverlayCreation(t *testing.T) {
	bindings := []KeyBinding{{Key: "?", Description: "Show shortcuts", Category: "General"}}
	ko := NewKeyOverlay(
		WithKeyOverlayTitle("Help"),
		WithKeyOverlayWidth(72),
		WithKeyOverlayBindings(bindings),
	)

	if ko == nil {
		t.Fatal("NewKeyOverlay returned nil")
	}
	if ko.title != "Help" {
		t.Errorf("expected title Help, got %q", ko.title)
	}
	if ko.overlayWidth != 72 {
		t.Errorf("expected overlayWidth 72, got %d", ko.overlayWidth)
	}
	if len(ko.bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(ko.bindings))
	}
	if ko.Visible() {
		t.Error("overlay should not be visible initially")
	}
}

func TestKeyOverlayAddBindings(t *testing.T) {
	ko := NewKeyOverlay()
	ko.AddBinding("?", "Show shortcuts", "General")
	ko.AddBindings([]KeyBinding{
		{Key: "j", Description: "Move down", Category: "Navigation"},
		{Key: "k", Description: "Move up", Category: "Navigation"},
	})

	if len(ko.bindings) != 3 {
		t.Fatalf("expected 3 bindings, got %d", len(ko.bindings))
	}

	if ko.bindings[0].Key != "?" {
		t.Errorf("expected first key '?', got %q", ko.bindings[0].Key)
	}
	if ko.bindings[1].Category != "Navigation" || ko.bindings[2].Category != "Navigation" {
		t.Error("expected navigation category for appended bindings")
	}
}

func TestKeyOverlayToggleVisibility(t *testing.T) {
	ko := NewKeyOverlay()

	ko.Toggle()
	if !ko.Visible() {
		t.Error("expected visible after first toggle")
	}

	ko.Toggle()
	if ko.Visible() {
		t.Error("expected hidden after second toggle")
	}

	ko.SetVisible(true)
	if !ko.Visible() {
		t.Error("expected visible after SetVisible(true)")
	}

	ko.SetVisible(false)
	if ko.Visible() {
		t.Error("expected hidden after SetVisible(false)")
	}
}

func TestKeyOverlayKeyboardToggle(t *testing.T) {
	ko := NewKeyOverlay()
	ko.Focus()

	ko.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !ko.Visible() {
		t.Error("expected '?' to toggle visible on")
	}

	ko.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if ko.Visible() {
		t.Error("expected Escape to toggle visible off")
	}
}

func TestKeyOverlayViewWhenHidden(t *testing.T) {
	ko := NewKeyOverlay()
	ko.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if view := ko.View(); view != "" {
		t.Errorf("expected empty view when hidden, got %q", view)
	}
}

func TestKeyOverlayViewWhenVisible(t *testing.T) {
	ko := NewKeyOverlay(
		WithKeyOverlayTitle("Keyboard Shortcuts"),
		WithKeyOverlayBindings([]KeyBinding{{Key: "?", Description: "Show help", Category: "General"}}),
	)
	ko.Focus()
	ko.SetVisible(true)
	ko.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := ko.View()
	if view == "" {
		t.Fatal("expected non-empty view when visible")
	}

	if !strings.Contains(view, "╭") || !strings.Contains(view, "╰") {
		t.Error("expected bordered box in view")
	}
	if !strings.Contains(view, "Keyboard Shortcuts") {
		t.Error("expected title in view")
	}
	if !strings.Contains(view, "Show help") {
		t.Error("expected binding description in view")
	}
}

func TestKeyOverlayCategoryGrouping(t *testing.T) {
	ko := NewKeyOverlay(
		WithKeyOverlayBindings([]KeyBinding{
			{Key: "?", Description: "Show shortcuts", Category: "General"},
			{Key: "j", Description: "Move down", Category: "Navigation"},
			{Key: "k", Description: "Move up", Category: "Navigation"},
		}),
	)
	ko.Focus()
	ko.SetVisible(true)
	ko.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	view := stripANSIKeyOverlay(ko.View())
	if !strings.Contains(view, "General") {
		t.Error("expected General category header")
	}
	if !strings.Contains(view, "Navigation") {
		t.Error("expected Navigation category header")
	}

	if strings.Count(view, "Navigation") != 1 {
		t.Error("expected one Navigation header")
	}
	if !strings.Contains(view, "Move down") || !strings.Contains(view, "Move up") {
		t.Error("expected both navigation bindings in grouped output")
	}
}

func TestKeyOverlayFocusManagement(t *testing.T) {
	ko := NewKeyOverlay()

	if ko.Focused() {
		t.Error("overlay should not be focused initially")
	}

	ko.Focus()
	if !ko.Focused() {
		t.Error("overlay should be focused after Focus()")
	}

	ko.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !ko.Visible() {
		t.Error("expected key handling when focused")
	}

	ko.Blur()
	if ko.Focused() {
		t.Error("overlay should not be focused after Blur()")
	}

	ko.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !ko.Visible() {
		t.Error("expected no key handling when blurred")
	}
}
