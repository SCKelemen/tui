package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type mockComponent struct {
	focused     bool
	updateCount int
	lastWidth   int
	lastHeight  int
	view        string
}

func (m *mockComponent) Init() tea.Cmd {
	return nil
}

func (m *mockComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	m.updateCount++
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.lastWidth = size.Width
		m.lastHeight = size.Height
	}
	return m, nil
}

func (m *mockComponent) View() string {
	if m.view != "" {
		return m.view
	}
	return "mock"
}

func (m *mockComponent) Focus() {
	m.focused = true
}

func (m *mockComponent) Blur() {
	m.focused = false
}

func (m *mockComponent) Focused() bool {
	return m.focused
}

func TestApplicationCreation(t *testing.T) {
	app := NewApplication()
	if app == nil {
		t.Fatal("NewApplication returned nil")
	}

	if app.focused != -1 {
		t.Fatalf("expected initial focused index -1, got %d", app.focused)
	}
}

func TestComponentAddition(t *testing.T) {
	app := NewApplication()
	c1 := &mockComponent{}

	app.AddComponent(c1)

	if len(app.components) != 1 {
		t.Errorf("expected 1 component, got %d", len(app.components))
	}

	if app.focused != 0 {
		t.Errorf("expected focused index 0, got %d", app.focused)
	}

	if !c1.Focused() {
		t.Error("first component should be focused after adding")
	}
}

func TestFocusManagement(t *testing.T) {
	app := NewApplication()
	c1 := &mockComponent{}
	c2 := &mockComponent{}

	app.AddComponent(c1)
	app.AddComponent(c2)

	if !c1.Focused() {
		t.Error("first component should be focused")
	}

	if c2.Focused() {
		t.Error("second component should not be focused")
	}

	app.Update(tea.KeyMsg{Type: tea.KeyTab})

	if c1.Focused() {
		t.Error("first component should not be focused after tab")
	}

	if !c2.Focused() {
		t.Error("second component should be focused after tab")
	}

	app.Update(tea.KeyMsg{Type: tea.KeyShiftTab})

	if !c1.Focused() {
		t.Error("first component should be focused after shift+tab")
	}

	if c2.Focused() {
		t.Error("second component should not be focused after shift+tab")
	}
}

func TestFocusComponent(t *testing.T) {
	app := NewApplication()
	c1 := &mockComponent{}
	c2 := &mockComponent{}
	app.AddComponent(c1)
	app.AddComponent(c2)

	app.FocusComponent(1)
	if !c2.Focused() || c1.Focused() {
		t.Fatal("expected component 2 focused and component 1 blurred")
	}

	app.FocusComponent(10)
	if !c2.Focused() {
		t.Fatal("focus should not change on out-of-range index")
	}
}

func TestWindowSizeMsg(t *testing.T) {
	app := NewApplication()
	c1 := &mockComponent{}
	c2 := &mockComponent{}
	app.AddComponent(c1)
	app.AddComponent(c2)

	app.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if app.width != 100 {
		t.Errorf("expected width 100, got %d", app.width)
	}

	if app.height != 50 {
		t.Errorf("expected height 50, got %d", app.height)
	}

	if c1.lastWidth != 100 || c1.lastHeight != 50 {
		t.Errorf("expected c1 to receive size 100x50, got %dx%d", c1.lastWidth, c1.lastHeight)
	}

	if c2.lastWidth != 100 || c2.lastHeight != 50 {
		t.Errorf("expected c2 to receive size 100x50, got %dx%d", c2.lastWidth, c2.lastHeight)
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
			t.Errorf("expected quit command for key %v, got nil", key)
		}
	}
}

func TestView(t *testing.T) {
	app := NewApplication()
	if got := app.View(); got != "No components" {
		t.Fatalf("expected %q, got %q", "No components", got)
	}

	c1 := &mockComponent{view: "one"}
	c2 := &mockComponent{view: "two"}
	app.AddComponent(c1)
	app.AddComponent(c2)

	if got := app.View(); got != "onetwo" {
		t.Fatalf("expected concatenated view %q, got %q", "onetwo", got)
	}
}
