package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestApplicationCreation(t *testing.T) {
	app := NewApplication()
	if app == nil {
		t.Fatal("NewApplication returned nil")
	}
}

func TestComponentAddition(t *testing.T) {
	app := NewApplication()
	statusBar := NewStatusBar()

	app.AddComponent(statusBar)

	if len(app.components) != 1 {
		t.Errorf("Expected 1 component, got %d", len(app.components))
	}

	if app.focused != 0 {
		t.Errorf("Expected focused index 0, got %d", app.focused)
	}
}

func TestFocusManagement(t *testing.T) {
	app := NewApplication()
	statusBar1 := NewStatusBar()
	statusBar2 := NewStatusBar()

	app.AddComponent(statusBar1)
	app.AddComponent(statusBar2)

	// First component should be focused
	if !statusBar1.Focused() {
		t.Error("First component should be focused")
	}

	if statusBar2.Focused() {
		t.Error("Second component should not be focused")
	}

	// Tab to next component
	app.Update(tea.KeyMsg{Type: tea.KeyTab})

	if statusBar1.Focused() {
		t.Error("First component should not be focused after tab")
	}

	if !statusBar2.Focused() {
		t.Error("Second component should be focused after tab")
	}
}

func TestWindowSizeMsg(t *testing.T) {
	app := NewApplication()
	statusBar := NewStatusBar()
	app.AddComponent(statusBar)

	app.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if app.width != 100 {
		t.Errorf("Expected width 100, got %d", app.width)
	}

	if app.height != 50 {
		t.Errorf("Expected height 50, got %d", app.height)
	}
}

func TestQuitKeys(t *testing.T) {
	app := NewApplication()

	tests := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'q'}},
		{Type: tea.KeyCtrlC},
	}

	for _, key := range tests {
		_, cmd := app.Update(key)
		if cmd == nil {
			t.Errorf("Expected quit command for key %v, got nil", key)
		}
	}
}

func TestStatusBarMessage(t *testing.T) {
	statusBar := NewStatusBar()
	testMsg := "Test status message"

	statusBar.SetMessage(testMsg)

	if statusBar.message != testMsg {
		t.Errorf("Expected message %q, got %q", testMsg, statusBar.message)
	}
}

func TestStatusBarView(t *testing.T) {
	statusBar := NewStatusBar()
	statusBar.SetMessage("Test")

	// Without width set, should return empty
	view := statusBar.View()
	if view != "" {
		t.Error("Expected empty view without width")
	}

	// Set width via Update
	statusBar.Update(tea.WindowSizeMsg{Width: 80})

	view = statusBar.View()
	if len(view) == 0 {
		t.Error("Expected non-empty view after setting width")
	}
}
