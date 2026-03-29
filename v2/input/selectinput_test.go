package input

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func selectInputTestItems() []SelectItem {
	return []SelectItem{
		{Label: "One", Value: "1"},
		{Label: "Two", Value: "2"},
		{Label: "Three", Value: "3"},
		{Label: "Four", Value: "4"},
		{Label: "Five", Value: "5"},
	}
}

func TestSelectInputView(t *testing.T) {
	si := NewSelectInput(selectInputTestItems(), WithSelectInputIndicator(">"))

	view := si.View()
	if view == "" {
		t.Fatal("View should not be empty when items exist")
	}

	if !strings.Contains(view, "> ") {
		t.Error("View should render indicator for selected item")
	}

	if !strings.Contains(view, "One") {
		t.Error("View should render item labels")
	}
}

func TestSelectInputNavigation(t *testing.T) {
	si := NewSelectInput(selectInputTestItems())
	si.Focus()

	if si.SelectedIndex() != 0 {
		t.Fatalf("expected initial index 0, got %d", si.SelectedIndex())
	}

	si.Update(tea.KeyMsg{Type: tea.KeyDown})
	if si.SelectedIndex() != 1 {
		t.Fatalf("expected index 1 after down, got %d", si.SelectedIndex())
	}

	si.Update(tea.KeyMsg{Type: tea.KeyUp})
	if si.SelectedIndex() != 0 {
		t.Fatalf("expected index 0 after up, got %d", si.SelectedIndex())
	}
}

func TestSelectInputSelect(t *testing.T) {
	si := NewSelectInput(selectInputTestItems())
	si.Focus()

	si.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, cmd := si.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected enter to return a command")
	}

	msg := cmd()
	selectMsg, ok := msg.(SelectInputMsg)
	if !ok {
		t.Fatalf("expected SelectInputMsg, got %T", msg)
	}

	if selectMsg.Item.Label != "Two" {
		t.Fatalf("expected selected item Two, got %q", selectMsg.Item.Label)
	}
}

func TestSelectInputScroll(t *testing.T) {
	si := NewSelectInput(selectInputTestItems(), WithSelectInputHeight(3))
	si.Focus()

	for i := 0; i < 3; i++ {
		si.Update(tea.KeyMsg{Type: tea.KeyDown})
	}

	if si.SelectedIndex() != 3 {
		t.Fatalf("expected cursor at index 3, got %d", si.SelectedIndex())
	}

	view := si.View()
	if strings.Contains(view, "One") {
		t.Error("expected scrolled view to not include first item")
	}
	if !strings.Contains(view, "Four") {
		t.Error("expected scrolled view to include current item")
	}
}

func TestSelectInputFocusBlur(t *testing.T) {
	si := NewSelectInput(selectInputTestItems(), WithSelectInputIndicator(">"))

	blurredView := si.View()
	si.Update(tea.KeyMsg{Type: tea.KeyDown})
	if si.SelectedIndex() != 0 {
		t.Fatal("cursor should not move while blurred")
	}

	si.Focus()
	si.Update(tea.KeyMsg{Type: tea.KeyDown})
	focusedView := si.View()
	if si.SelectedIndex() != 1 {
		t.Fatal("cursor should move while focused")
	}
	if focusedView == blurredView {
		t.Error("view should change after focus and navigation")
	}

	si.Blur()
	afterBlurView := si.View()
	si.Update(tea.KeyMsg{Type: tea.KeyDown})
	if si.SelectedIndex() != 1 {
		t.Fatal("cursor should stop moving after blur")
	}
	if afterBlurView != si.View() {
		t.Error("view should remain unchanged while blurred and handling keys")
	}
}

func TestSelectInputSetItems(t *testing.T) {
	si := NewSelectInput(selectInputTestItems())
	si.Focus()
	si.Update(tea.KeyMsg{Type: tea.KeyDown})
	si.Update(tea.KeyMsg{Type: tea.KeyDown})

	replacement := []SelectItem{
		{Label: "Alpha", Value: "a"},
		{Label: "Beta", Value: "b"},
	}
	si.SetItems(replacement)

	if si.SelectedIndex() != 1 {
		t.Fatalf("expected clamped cursor at 1, got %d", si.SelectedIndex())
	}
	if si.Selected().Label != "Beta" {
		t.Fatalf("expected selected label Beta, got %q", si.Selected().Label)
	}
	view := si.View()
	if !strings.Contains(view, "Alpha") || !strings.Contains(view, "Beta") {
		t.Error("view should contain replacement items")
	}
	if strings.Contains(view, "One") {
		t.Error("view should not contain old items")
	}
}

func TestSelectInputSelected(t *testing.T) {
	si := NewSelectInput(selectInputTestItems())
	si.Focus()
	si.Update(tea.KeyMsg{Type: tea.KeyDown})

	selected := si.Selected()
	if selected.Label != "Two" || selected.Value != "2" {
		t.Fatalf("expected selected item Two/2, got %q/%q", selected.Label, selected.Value)
	}
}

func TestSelectInputEmpty(t *testing.T) {
	si := NewSelectInput(nil)

	if si.View() != "" {
		t.Error("expected empty view for empty items")
	}

	if si.SelectedIndex() != -1 {
		t.Fatalf("expected SelectedIndex -1 for empty items, got %d", si.SelectedIndex())
	}

	selected := si.Selected()
	if selected != (SelectItem{}) {
		t.Fatalf("expected zero SelectItem for empty items, got %+v", selected)
	}

	si.Focus()
	_, cmd := si.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil cmd when pressing enter with empty items")
	}
}
