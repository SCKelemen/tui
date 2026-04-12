package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type testFocusable struct {
	id        string
	focusable bool
	focused   bool
	tabIndex  int
}

func (f *testFocusable) ID() string              { return f.id }
func (f *testFocusable) IsFocusable() bool       { return f.focusable }
func (f *testFocusable) SetFocused(focused bool) { f.focused = focused }
func (f *testFocusable) IsFocused() bool         { return f.focused }
func (f *testFocusable) TabIndex() int           { return f.tabIndex }

func TestFocusManager(t *testing.T) {
	m := NewFocusManager()
	a := &testFocusable{id: "a", focusable: true, tabIndex: 1}
	b := &testFocusable{id: "b", focusable: true, tabIndex: 2}
	c := &testFocusable{id: "c", focusable: true, tabIndex: 3}
	d := &testFocusable{id: "d", focusable: false, tabIndex: 4}

	m.Register(c)
	m.Register(a)
	m.Register(d)
	m.Register(b)

	if got := m.Count(); got != 4 {
		t.Fatalf("Count() = %d, want 4", got)
	}
	if m.focusables[0].ID() != "a" || m.focusables[1].ID() != "b" || m.focusables[2].ID() != "c" || m.focusables[3].ID() != "d" {
		t.Fatalf("focus order = [%s %s %s %s], want [a b c d]",
			m.focusables[0].ID(), m.focusables[1].ID(), m.focusables[2].ID(), m.focusables[3].ID())
	}

	m.FocusNext()
	if got := m.CurrentID(); got != "a" {
		t.Fatalf("CurrentID() after FocusNext = %q, want %q", got, "a")
	}
	m.FocusNext()
	if got := m.CurrentID(); got != "b" {
		t.Fatalf("CurrentID() after second FocusNext = %q, want %q", got, "b")
	}
	m.FocusPrev()
	if got := m.CurrentID(); got != "a" {
		t.Fatalf("CurrentID() after FocusPrev = %q, want %q", got, "a")
	}

	if !m.FocusByID("c") || m.CurrentID() != "c" {
		t.Fatalf("FocusByID(c) failed, CurrentID() = %q", m.CurrentID())
	}
	if m.FocusByIndex(3) {
		t.Fatal("FocusByIndex(3) = true for non-focusable item, want false")
	}
	if !m.FocusByIndex(1) || m.CurrentID() != "b" {
		t.Fatalf("FocusByIndex(1) failed, CurrentID() = %q", m.CurrentID())
	}

	if !m.HandleKeyMsg(tea.KeyMsg{Type: tea.KeyTab}) || m.CurrentID() != "c" {
		t.Fatalf("HandleKeyMsg(Tab) failed, CurrentID() = %q", m.CurrentID())
	}
	if !m.HandleKeyMsg(tea.KeyMsg{Type: tea.KeyShiftTab}) || m.CurrentID() != "b" {
		t.Fatalf("HandleKeyMsg(Shift+Tab) failed, CurrentID() = %q", m.CurrentID())
	}
	if m.HandleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}) {
		t.Fatal("HandleKeyMsg(non-tab) = true, want false")
	}

	m.SetTrapFocus(true)
	if !m.trapFocus {
		t.Fatal("trapFocus = false after SetTrapFocus(true), want true")
	}

	m.Unregister("b")
	if got := m.Count(); got != 3 {
		t.Fatalf("Count() after Unregister = %d, want 3", got)
	}
	if got := m.CurrentID(); got != "c" {
		t.Fatalf("CurrentID() after unregistering focused item = %q, want %q", got, "c")
	}
}

func TestFocusGroup(t *testing.T) {
	parent := NewFocusManager()
	p1 := &testFocusable{id: "p1", focusable: true, tabIndex: 1}
	p2 := &testFocusable{id: "p2", focusable: true, tabIndex: 2}
	parent.Register(p1)
	parent.Register(p2)
	parent.FocusByID("p2")

	group := NewFocusGroup(parent)
	g1 := &testFocusable{id: "g1", focusable: true, tabIndex: 1}
	g2 := &testFocusable{id: "g2", focusable: true, tabIndex: 2}
	group.Manager().Register(g1)
	group.Manager().Register(g2)

	group.Enter()
	if !group.IsActive() {
		t.Fatal("IsActive() after Enter = false, want true")
	}
	if !group.manager.trapFocus || !parent.trapFocus {
		t.Fatal("trap focus not enabled on group/parent after Enter")
	}
	if got := group.manager.CurrentID(); got != "g1" {
		t.Fatalf("group current after Enter = %q, want %q", got, "g1")
	}

	group.Exit()
	if group.IsActive() {
		t.Fatal("IsActive() after Exit = true, want false")
	}
	if group.manager.trapFocus || parent.trapFocus {
		t.Fatal("trap focus still enabled after Exit")
	}
	if got := parent.CurrentID(); got != "p2" {
		t.Fatalf("parent focus after Exit = %q, want %q", got, "p2")
	}
}

func TestFocusRing(t *testing.T) {
	ring := NewFocusRing("a", "b", "c")

	if got := ring.Current(); got != "" {
		t.Fatalf("initial Current() = %q, want empty", got)
	}
	if got := ring.Next(); got != "a" {
		t.Fatalf("Next() = %q, want %q", got, "a")
	}
	if got := ring.Next(); got != "b" {
		t.Fatalf("second Next() = %q, want %q", got, "b")
	}
	if got := ring.Prev(); got != "a" {
		t.Fatalf("Prev() = %q, want %q", got, "a")
	}

	ring.SetCurrent("c")
	if got := ring.Current(); got != "c" {
		t.Fatalf("Current() after SetCurrent(c) = %q, want %q", got, "c")
	}
	if got := ring.Next(); got != "a" {
		t.Fatalf("Next() after SetCurrent(c) = %q, want %q", got, "a")
	}
}
