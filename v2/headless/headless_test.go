package headless

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	tui "github.com/SCKelemen/tui/v2"
)

// mockComp is a minimal tui.Component used to observe what the
// renderer does to it.
type mockComp struct {
	inits   int
	updates int
	lastMsg tea.Msg
	view    string
	focused bool

	// initCmd is invoked, if non-nil, by Init.
	initCmd tea.Cmd
	// updateCmd is invoked, if non-nil, by Update.
	updateCmd tea.Cmd
}

func (m *mockComp) Init() tea.Cmd {
	m.inits++
	return m.initCmd
}

func (m *mockComp) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	m.updates++
	m.lastMsg = msg
	return m, m.updateCmd
}

func (m *mockComp) View() string  { return m.view }
func (m *mockComp) Focus()        { m.focused = true }
func (m *mockComp) Blur()         { m.focused = false }
func (m *mockComp) Focused() bool { return m.focused }

func TestMountSendsWindowSizeAndRunsInit(t *testing.T) {
	m := &mockComp{}
	r := NewRenderer(80, 24)
	if err := r.Mount(m); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	if m.inits != 1 {
		t.Fatalf("inits = %d, want 1", m.inits)
	}
	ws, ok := m.lastMsg.(tea.WindowSizeMsg)
	if !ok {
		t.Fatalf("lastMsg = %#v, want tea.WindowSizeMsg", m.lastMsg)
	}
	if ws.Width != 80 || ws.Height != 24 {
		t.Fatalf("size = %dx%d, want 80x24", ws.Width, ws.Height)
	}
	if r.Component() != m {
		t.Fatalf("Component() returned %#v, want mock", r.Component())
	}
}

func TestMountSkipsWindowSizeWhenWidthZero(t *testing.T) {
	m := &mockComp{}
	r := NewRenderer(0, 24)
	if err := r.Mount(m); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	if m.updates != 0 {
		t.Fatalf("updates = %d, want 0", m.updates)
	}
}

func TestSendIncrementsUpdate(t *testing.T) {
	m := &mockComp{}
	r := NewRenderer(10, 3)
	if err := r.Mount(m); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	before := m.updates
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	if err := r.Send(key); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if m.updates != before+1 {
		t.Fatalf("updates = %d, want %d", m.updates, before+1)
	}
	got, ok := m.lastMsg.(tea.KeyMsg)
	if !ok {
		t.Fatalf("lastMsg = %#v, want tea.KeyMsg", m.lastMsg)
	}
	if !reflect.DeepEqual(got, key) {
		t.Fatalf("lastMsg = %#v, want %#v", got, key)
	}
}

func TestViewPreservesANSIAndPlainStrips(t *testing.T) {
	m := &mockComp{view: "\x1b[31mred\x1b[0m"}
	r := NewRenderer(0, 0)
	if err := r.Mount(m); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	if !strings.Contains(r.View(), "\x1b[") {
		t.Fatalf("View() lost ANSI: %q", r.View())
	}
	if r.ViewPlain() != "red" {
		t.Fatalf("ViewPlain() = %q, want %q", r.ViewPlain(), "red")
	}
}

func TestFrameLayout(t *testing.T) {
	m := &mockComp{view: "hi\nworld"}
	r := NewRenderer(10, 3)
	if err := r.Mount(m); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	frame := r.Frame()
	if len(frame) != 3 {
		t.Fatalf("rows = %d, want 3", len(frame))
	}
	for i, row := range frame {
		if len(row) != 10 {
			t.Fatalf("row %d width = %d, want 10", i, len(row))
		}
	}

	wantRow0 := []rune{'h', 'i', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '}
	if !reflect.DeepEqual(frame[0], wantRow0) {
		t.Fatalf("row 0 = %q, want %q", string(frame[0]), string(wantRow0))
	}
	if string(frame[1][:5]) != "world" {
		t.Fatalf("row 1 prefix = %q, want %q", string(frame[1][:5]), "world")
	}
	for j := 5; j < 10; j++ {
		if frame[1][j] != ' ' {
			t.Fatalf("row 1 col %d = %q, want space", j, frame[1][j])
		}
	}
	for j := 0; j < 10; j++ {
		if frame[2][j] != ' ' {
			t.Fatalf("row 2 col %d = %q, want space", j, frame[2][j])
		}
	}
}

func TestFrameWideRune(t *testing.T) {
	// '世' has width 2 in East Asian Width tables.
	m := &mockComp{view: "世a"}
	r := NewRenderer(4, 1)
	if err := r.Mount(m); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	frame := r.Frame()
	if frame[0][0] != '世' {
		t.Fatalf("col 0 = %q, want 世", frame[0][0])
	}
	if frame[0][1] != 0 {
		t.Fatalf("col 1 = %d, want 0 (wide-rune right cell)", frame[0][1])
	}
	if frame[0][2] != 'a' {
		t.Fatalf("col 2 = %q, want a", frame[0][2])
	}
	if frame[0][3] != ' ' {
		t.Fatalf("col 3 = %q, want space", frame[0][3])
	}
}

func TestKeystrokeSendsEachRune(t *testing.T) {
	m := &mockComp{}
	r := NewRenderer(10, 3)
	if err := r.Mount(m); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	before := m.updates
	if err := r.Keystroke("ab"); err != nil {
		t.Fatalf("Keystroke: %v", err)
	}
	if m.updates-before != 2 {
		t.Fatalf("updates delta = %d, want 2", m.updates-before)
	}
	last, ok := m.lastMsg.(tea.KeyMsg)
	if !ok {
		t.Fatalf("lastMsg = %#v, want tea.KeyMsg", m.lastMsg)
	}
	if len(last.Runes) != 1 || last.Runes[0] != 'b' {
		t.Fatalf("last runes = %v, want [b]", last.Runes)
	}
}

func TestKeyEnter(t *testing.T) {
	m := &mockComp{}
	r := NewRenderer(10, 3)
	if err := r.Mount(m); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	if err := r.Key("enter"); err != nil {
		t.Fatalf("Key: %v", err)
	}
	got, ok := m.lastMsg.(tea.KeyMsg)
	if !ok {
		t.Fatalf("lastMsg = %#v, want tea.KeyMsg", m.lastMsg)
	}
	if got.Type != tea.KeyEnter {
		t.Fatalf("type = %v, want tea.KeyEnter", got.Type)
	}
}

func TestKeyUnknown(t *testing.T) {
	r := NewRenderer(10, 3)
	if err := r.Mount(&mockComp{}); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	if err := r.Key("nope"); err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestRecursionCap(t *testing.T) {
	// loopComp's Update returns a cmd that re-emits a custom message
	// forever; the renderer must bail out with an error.
	type loopMsg struct{}
	m := &loopComp{}
	m.updateCmd = func() tea.Msg { return loopMsg{} }
	r := NewRenderer(0, 0)
	if err := r.Mount(m); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	err := r.Send(loopMsg{})
	if err == nil {
		t.Fatal("expected recursion cap error, got nil")
	}
	if !strings.Contains(err.Error(), "exceeded") {
		t.Fatalf("error = %v, want 'exceeded'", err)
	}
	if m.updates > maxMessages+1 {
		t.Fatalf("updates = %d, want <= %d", m.updates, maxMessages+1)
	}
}

// loopComp is a stripped-down component whose Update always returns
// the same self-looping cmd. It cannot be a mockComp because mockComp
// uses a fixed cmd field but does not re-set it; here we just want
// the same behavior with no other state.
type loopComp struct {
	updates   int
	updateCmd tea.Cmd
}

func (l *loopComp) Init() tea.Cmd { return nil }
func (l *loopComp) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	l.updates++
	return l, l.updateCmd
}
func (l *loopComp) View() string  { return "" }
func (l *loopComp) Focus()        {}
func (l *loopComp) Blur()         {}
func (l *loopComp) Focused() bool { return false }

func TestBatchMsgDrains(t *testing.T) {
	// A batch with two no-op cmds should leave the component with
	// two extra updates: one per dispatched message. tea.BatchMsg
	// itself is processed but does not count toward Update calls.
	m := &mockComp{}
	r := NewRenderer(0, 0)
	if err := r.Mount(m); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	before := m.updates

	tick := func() tea.Cmd { return func() tea.Msg { return struct{ Name string }{"tick"} } }
	batch := tea.BatchMsg{tick(), tick()}
	if err := r.Send(batch); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if m.updates-before != 2 {
		t.Fatalf("updates delta = %d, want 2", m.updates-before)
	}
}

func TestQuitMsgStopsChain(t *testing.T) {
	// A cmd returning tea.QuitMsg should not recurse into Update.
	m := &mockComp{}
	m.updateCmd = tea.Quit
	r := NewRenderer(0, 0)
	if err := r.Mount(m); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	before := m.updates
	if err := r.Send(struct{}{}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	// Exactly one Update (the original message); Quit must not be
	// fed back through Update.
	if m.updates-before != 1 {
		t.Fatalf("updates delta = %d, want 1", m.updates-before)
	}
}
