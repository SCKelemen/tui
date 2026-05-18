package tui

import (
	"regexp"
	"strings"
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
	t.Run("default quit key is ctrl+c only", func(t *testing.T) {
		app := NewApplication()

		_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if cmd != nil {
			t.Fatal("expected no quit command for key q by default")
		}

		_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		if cmd == nil {
			t.Fatal("expected quit command for key ctrl+c by default")
		}
	})

	t.Run("custom quit key via option", func(t *testing.T) {
		app := NewApplication(WithQuitKey("q"))

		_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if cmd == nil {
			t.Fatal("expected quit command for key q when configured with WithQuitKey")
		}

		_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		if cmd != nil {
			t.Fatal("expected no quit command for key ctrl+c when quit key is set to q")
		}
	})
}
var ansiControlRE = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)

func normalizeFrameBufferOutput(s string) string {
	cleaned := ansiControlRE.ReplaceAllString(s, "")
	return strings.TrimRight(cleaned, " \n\r\t")
}

func TestView(t *testing.T) {
	app := NewApplication()
	if got := normalizeFrameBufferOutput(app.View()); got != "No components" {
		t.Fatalf("expected %q, got %q", "No components", got)
	}

	c1 := &mockComponent{view: "one"}
	c2 := &mockComponent{view: "two"}
	app.AddComponent(c1)
	app.AddComponent(c2)

	if got := normalizeFrameBufferOutput(app.View()); got != "onetwo" {
		t.Fatalf("expected concatenated view %q, got %q", "onetwo", got)
	}
}

// keyClaimingComponent is a mockComponent that implements KeyConsumer.
type keyClaimingComponent struct {
	mockComponent
	claim    map[string]bool
	lastKeys []string
}

func (k *keyClaimingComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		k.lastKeys = append(k.lastKeys, km.String())
	}
	_, cmd := k.mockComponent.Update(msg)
	return k, cmd
}

func (k *keyClaimingComponent) HandlesKey(key string) bool {
	return k.claim[key]
}

func TestKeyConsumerClaimsTab(t *testing.T) {
	app := NewApplication()
	other := &mockComponent{}
	claimer := &keyClaimingComponent{claim: map[string]bool{"tab": true}}

	app.AddComponent(claimer)
	app.AddComponent(other)

	// claimer is focused (index 0).
	if !claimer.Focused() {
		t.Fatal("expected claimer to be focused initially")
	}

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd != nil {
		t.Fatalf("expected no command when key is claimed, got %v", cmd)
	}

	// Focus must NOT have cycled.
	if !claimer.Focused() {
		t.Fatal("claimer should still be focused — Tab was claimed")
	}
	if other.Focused() {
		t.Fatal("other should not be focused after claimed Tab")
	}

	// Claimer should have received the Tab key.
	if len(claimer.lastKeys) == 0 || claimer.lastKeys[len(claimer.lastKeys)-1] != "tab" {
		t.Fatalf("expected claimer to receive tab key, got %v", claimer.lastKeys)
	}
}

func TestKeyConsumerDoesNotInterceptUnclaimedKeys(t *testing.T) {
	app := NewApplication()
	claimer := &keyClaimingComponent{claim: map[string]bool{"enter": true}}
	other := &mockComponent{}
	app.AddComponent(claimer)
	app.AddComponent(other)

	// claimer claims "enter" but not "tab" — Tab should still cycle focus.
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !other.Focused() {
		t.Fatal("expected Tab to cycle to second component when not claimed")
	}
}

func TestQuitKeyFiresWhenNoComponentFocused(t *testing.T) {
	app := NewApplication()
	// No components added — focused remains -1.
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected quit command when no component is focused")
	}
}

// mouseAwareComponent records every MouseMsg it sees so tests can assert
// that the Application routed the click correctly.
type mouseAwareComponent struct {
	mockComponent
	mouseEvents []tea.MouseMsg
}

func (m *mouseAwareComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	if mm, ok := msg.(tea.MouseMsg); ok {
		m.mouseEvents = append(m.mouseEvents, mm)
	}
	_, cmd := m.mockComponent.Update(msg)
	return m, cmd
}

// boundedComponent embeds mouseAwareComponent and implements Bounded.
type boundedComponent struct {
	mouseAwareComponent
	bx, by, bw, bh int
}

func (b *boundedComponent) Bounds() (int, int, int, int) {
	return b.bx, b.by, b.bw, b.bh
}

func (b *boundedComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	if mm, ok := msg.(tea.MouseMsg); ok {
		b.mouseEvents = append(b.mouseEvents, mm)
	}
	_, cmd := b.mockComponent.Update(msg)
	return b, cmd
}

func TestMouseHitTestRefocusesAndRoutesClick(t *testing.T) {
	app := NewApplication()
	left := &boundedComponent{bx: 0, by: 0, bw: 20, bh: 10}
	right := &boundedComponent{bx: 20, by: 0, bw: 20, bh: 10}

	app.AddComponent(left)
	app.AddComponent(right)

	if !left.Focused() {
		t.Fatal("expected left to be focused initially")
	}

	// Click in the right component's rect.
	_, _ = app.Update(tea.MouseMsg{
		X:      25,
		Y:      4,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})

	if !right.Focused() {
		t.Fatal("expected right to be focused after click in its rect")
	}
	if left.Focused() {
		t.Fatal("expected left to be blurred after focus moved")
	}
	if len(right.mouseEvents) != 1 {
		t.Fatalf("expected right to receive 1 mouse event, got %d", len(right.mouseEvents))
	}
	if len(left.mouseEvents) != 0 {
		t.Fatalf("expected left to receive 0 mouse events, got %d", len(left.mouseEvents))
	}

	// Click back into the left component's rect.
	_, _ = app.Update(tea.MouseMsg{
		X:      5,
		Y:      2,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})

	if !left.Focused() {
		t.Fatal("expected left to be focused after click in its rect")
	}
	if right.Focused() {
		t.Fatal("expected right to be blurred after focus moved back")
	}
	if len(left.mouseEvents) != 1 {
		t.Fatalf("expected left to receive 1 mouse event, got %d", len(left.mouseEvents))
	}
}

func TestSetComponentBoundsHitTest(t *testing.T) {
	app := NewApplication()
	a := &mouseAwareComponent{}
	b := &mouseAwareComponent{}
	app.AddComponent(a)
	app.AddComponent(b)

	// No Bounded interface — use SetComponentBounds explicitly.
	app.SetComponentBounds(0, 0, 0, 10, 5)
	app.SetComponentBounds(1, 10, 0, 10, 5)

	_, _ = app.Update(tea.MouseMsg{
		X:      12,
		Y:      2,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})

	if !b.Focused() {
		t.Fatal("expected b to be focused after click in its registered rect")
	}
	if len(b.mouseEvents) != 1 {
		t.Fatalf("expected b to receive 1 mouse event, got %d", len(b.mouseEvents))
	}
}

func TestMouseClickOutsideAllBoundsStaysWithFocused(t *testing.T) {
	app := NewApplication()
	left := &boundedComponent{bx: 0, by: 0, bw: 10, bh: 5}
	right := &boundedComponent{bx: 20, by: 0, bw: 10, bh: 5}
	app.AddComponent(left)
	app.AddComponent(right)

	// Click in a gap between components — should not change focus.
	_, _ = app.Update(tea.MouseMsg{
		X:      15,
		Y:      2,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})

	if !left.Focused() {
		t.Fatal("expected left to remain focused after click in gap")
	}
	if len(left.mouseEvents) != 1 {
		t.Fatalf("expected left (focused) to receive the gap click, got %d", len(left.mouseEvents))
	}
}

func TestMouseMotionStaysWithFocused(t *testing.T) {
	app := NewApplication()
	left := &boundedComponent{bx: 0, by: 0, bw: 10, bh: 5}
	right := &boundedComponent{bx: 20, by: 0, bw: 10, bh: 5}
	app.AddComponent(left)
	app.AddComponent(right)

	// Motion event over right's rect must NOT refocus.
	_, _ = app.Update(tea.MouseMsg{
		X:      25,
		Y:      2,
		Action: tea.MouseActionMotion,
		Button: tea.MouseButtonLeft,
	})

	if !left.Focused() {
		t.Fatal("expected left to remain focused; motion must not refocus")
	}
	if len(left.mouseEvents) != 1 {
		t.Fatalf("expected left (focused) to receive motion, got %d", len(left.mouseEvents))
	}
	if len(right.mouseEvents) != 0 {
		t.Fatalf("expected right to receive no motion events, got %d", len(right.mouseEvents))
	}
}

func TestExistingFocusCyclingPreservedForNonClaimingComponents(t *testing.T) {
	app := NewApplication()
	c1 := &mockComponent{}
	c2 := &mockComponent{}
	c3 := &mockComponent{}
	app.AddComponent(c1)
	app.AddComponent(c2)
	app.AddComponent(c3)

	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !c2.Focused() {
		t.Fatal("expected c2 focused after first Tab")
	}
	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !c3.Focused() {
		t.Fatal("expected c3 focused after second Tab")
	}
	app.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if !c2.Focused() {
		t.Fatal("expected c2 focused after Shift+Tab")
	}
}