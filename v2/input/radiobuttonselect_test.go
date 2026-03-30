package input

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func testRadioItems() []RadioItem {
	return []RadioItem{
		{Label: "Alpha", Value: "a", Description: "first"},
		{Label: "Beta", Value: "b", Description: "second"},
		{Label: "Gamma", Value: "c", Description: "third"},
	}
}

func TestRadioButtonSelectConstructor(t *testing.T) {
	r := NewRadioButtonSelect(testRadioItems(), WithRadioButtonSelectSelected(1))
	if r == nil {
		t.Fatal("NewRadioButtonSelect returned nil")
	}

	if r.selected != 1 {
		t.Fatalf("expected selected index 1, got %d", r.selected)
	}
}

func TestRadioButtonSelectItemsAndView(t *testing.T) {
	r := NewRadioButtonSelect(testRadioItems())
	view := r.View()
	if view == "" {
		t.Fatal("view should not be empty with items")
	}

	if !strings.Contains(view, "Alpha") || !strings.Contains(view, "Beta") {
		t.Fatal("view should render item labels")
	}

	if !strings.Contains(view, "○") || !strings.Contains(view, "●") {
		t.Fatal("view should render radio button markers")
	}
}

func TestRadioButtonSelectSelection(t *testing.T) {
	r := NewRadioButtonSelect(testRadioItems())
	r.Focus()

	r.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, cmd := r.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command on selection")
	}

	msg := cmd()
	selectedMsg, ok := msg.(RadioButtonSelectMsg)
	if !ok {
		t.Fatalf("expected RadioButtonSelectMsg, got %T", msg)
	}

	if selectedMsg.Value != "b" {
		t.Fatalf("expected selected value b, got %q", selectedMsg.Value)
	}
}
