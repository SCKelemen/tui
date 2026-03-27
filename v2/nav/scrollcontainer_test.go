package nav

import (
	"strings"
	"testing"

	tui "github.com/SCKelemen/tui/v2"
	tea "github.com/charmbracelet/bubbletea"
)

type scrollTestChild struct {
	view       string
	focused    bool
	updateHits int
}

func (c *scrollTestChild) Init() tea.Cmd { return nil }

func (c *scrollTestChild) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	c.updateHits++
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		c.view = c.view + ""
		_ = ws
	}
	return c, nil
}

func (c *scrollTestChild) View() string  { return c.view }
func (c *scrollTestChild) Focus()        { c.focused = true }
func (c *scrollTestChild) Blur()         { c.focused = false }
func (c *scrollTestChild) Focused() bool { return c.focused }

func TestScrollContainerCreation(t *testing.T) {
	sc := NewScrollContainer()
	if sc == nil {
		t.Fatal("NewScrollContainer returned nil")
	}
	if sc.showScrollbar != true {
		t.Fatal("showScrollbar should default to true")
	}
	if sc.scrollbarChar != "█" {
		t.Fatalf("expected default scrollbar char █, got %q", sc.scrollbarChar)
	}
	if sc.autoScroll {
		t.Fatal("autoScroll should default to false")
	}
	if sc.focusedChild != -1 {
		t.Fatalf("focusedChild should default to -1, got %d", sc.focusedChild)
	}
}

func TestScrollContainerOptions(t *testing.T) {
	sc := NewScrollContainer(
		WithShowScrollbar(false),
		WithScrollbarChar("#"),
		WithAutoScroll(true),
		WithViewportHeight(7),
		WithScrollContainerTheme("nord"),
	)

	if sc.showScrollbar {
		t.Fatal("showScrollbar option not applied")
	}
	if sc.scrollbarChar != "#" {
		t.Fatalf("expected scrollbar char #, got %q", sc.scrollbarChar)
	}
	if !sc.autoScroll {
		t.Fatal("autoScroll option not applied")
	}
	if sc.viewportHeight != 7 {
		t.Fatalf("expected viewportHeight 7, got %d", sc.viewportHeight)
	}
	if sc.designTokens == nil {
		t.Fatal("theme option should set design tokens")
	}
}

func TestScrollContainerAddRemoveChildren(t *testing.T) {
	sc := NewScrollContainer()
	c1 := &scrollTestChild{view: "a"}
	c2 := &scrollTestChild{view: "b"}
	c3 := &scrollTestChild{view: "c"}

	sc.AddChild(c1)
	sc.AddChild(c2)
	sc.AddChild(c3)

	if sc.ChildCount() != 3 {
		t.Fatalf("expected 3 children, got %d", sc.ChildCount())
	}

	sc.RemoveChild(1)
	if sc.ChildCount() != 2 {
		t.Fatalf("expected 2 children after remove, got %d", sc.ChildCount())
	}
	children := sc.GetChildren()
	if children[0] != c1 || children[1] != c3 {
		t.Fatal("unexpected child order after remove")
	}

	sc.RemoveChild(-1)
	sc.RemoveChild(99)
	if sc.ChildCount() != 2 {
		t.Fatal("invalid remove should not change child count")
	}
}

func TestScrollContainerSetChildren(t *testing.T) {
	sc := NewScrollContainer()
	c1 := &scrollTestChild{view: "one"}
	c2 := &scrollTestChild{view: "two"}

	sc.SetChildren([]tui.Component{c1, c2})

	if sc.ChildCount() != 2 {
		t.Fatalf("expected 2 children, got %d", sc.ChildCount())
	}
	if sc.GetChildren()[0] != c1 || sc.GetChildren()[1] != c2 {
		t.Fatal("set children not applied")
	}
}

func TestScrollContainerScrollPosition(t *testing.T) {
	sc := NewScrollContainer(WithShowScrollbar(false), WithViewportHeight(2))
	sc.viewportWidth = 20
	sc.AddChild(&scrollTestChild{view: "1\n2\n3\n4\n5"})

	sc.ScrollTo(2)
	if sc.GetScrollOffset() != 2 {
		t.Fatalf("expected offset 2, got %d", sc.GetScrollOffset())
	}

	sc.ScrollTo(999)
	if sc.GetScrollOffset() != 3 {
		t.Fatalf("expected clamped offset 3, got %d", sc.GetScrollOffset())
	}

	sc.ScrollToTop()
	if sc.GetScrollOffset() != 0 {
		t.Fatal("ScrollToTop should set offset to 0")
	}

	sc.ScrollToBottom()
	if sc.GetScrollOffset() != 3 {
		t.Fatalf("ScrollToBottom expected 3, got %d", sc.GetScrollOffset())
	}
}

func TestScrollContainerViewportClipping(t *testing.T) {
	sc := NewScrollContainer(WithShowScrollbar(false), WithViewportHeight(3))
	sc.viewportWidth = 20
	sc.AddChild(&scrollTestChild{view: "line1\nline2\nline3\nline4\nline5"})

	sc.ScrollTo(1)
	view := sc.View()

	if !strings.Contains(view, "line2") || !strings.Contains(view, "line4") {
		t.Fatalf("expected clipped view to include line2..line4, got:\n%s", view)
	}
	if strings.Contains(view, "line1") || strings.Contains(view, "line5") {
		t.Fatalf("expected clipped view to exclude line1 and line5, got:\n%s", view)
	}
}

func TestScrollContainerKeyboardNavigation(t *testing.T) {
	sc := NewScrollContainer(WithShowScrollbar(false), WithViewportHeight(4))
	sc.viewportWidth = 20
	sc.Focus()
	sc.AddChild(&scrollTestChild{view: "1\n2\n3\n4\n5\n6\n7\n8"})

	sc.Update(tea.KeyMsg{Type: tea.KeyDown})
	if sc.GetScrollOffset() != 1 {
		t.Fatalf("expected offset 1 after down key, got %d", sc.GetScrollOffset())
	}

	sc.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if sc.GetScrollOffset() != 3 {
		t.Fatalf("expected offset 3 after ctrl+d, got %d", sc.GetScrollOffset())
	}

	sc.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	if sc.GetScrollOffset() != 1 {
		t.Fatalf("expected offset 1 after ctrl+u, got %d", sc.GetScrollOffset())
	}

	sc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if sc.GetScrollOffset() != 4 {
		t.Fatalf("expected bottom offset 4 after G, got %d", sc.GetScrollOffset())
	}

	sc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if sc.GetScrollOffset() != 0 {
		t.Fatalf("expected top offset 0 after g, got %d", sc.GetScrollOffset())
	}
}

func TestScrollContainerChildFocusCyclingAndForwarding(t *testing.T) {
	sc := NewScrollContainer(WithShowScrollbar(false), WithViewportHeight(4))
	c1 := &scrollTestChild{view: "a"}
	c2 := &scrollTestChild{view: "b"}
	c3 := &scrollTestChild{view: "c"}

	sc.AddChild(c1)
	sc.AddChild(c2)
	sc.AddChild(c3)
	sc.Focus()

	if !c1.Focused() {
		t.Fatal("expected first child to be focused when container gains focus")
	}

	sc.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !c2.Focused() || c1.Focused() {
		t.Fatal("tab should move focus to second child")
	}

	sc.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if !c1.Focused() || c2.Focused() {
		t.Fatal("shift+tab should move focus back to first child")
	}

	hitsBefore := c1.updateHits
	sc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if c1.updateHits <= hitsBefore {
		t.Fatal("unhandled key should be forwarded to focused child")
	}
}

func TestScrollContainerScrollbarRendering(t *testing.T) {
	sc := NewScrollContainer(WithShowScrollbar(true), WithViewportHeight(4), WithScrollbarChar("▓"))
	sc.viewportWidth = 18
	sc.AddChild(&scrollTestChild{view: "1\n2\n3\n4\n5\n6\n7\n8\n9\n10"})

	view := sc.View()
	if !strings.Contains(view, "│") {
		t.Fatalf("expected scrollbar track in view, got:\n%s", view)
	}
	if !strings.Contains(view, "▓") {
		t.Fatalf("expected scrollbar thumb char in view, got:\n%s", view)
	}
}

func TestScrollContainerAutoScroll(t *testing.T) {
	sc := NewScrollContainer(WithShowScrollbar(false), WithViewportHeight(2), WithAutoScroll(true))
	sc.viewportWidth = 20

	sc.AddChild(&scrollTestChild{view: "row1"})
	if sc.GetScrollOffset() != 0 {
		t.Fatalf("offset should remain 0 for first row, got %d", sc.GetScrollOffset())
	}

	sc.AddChild(&scrollTestChild{view: "row2"})
	if sc.GetScrollOffset() != 0 {
		t.Fatalf("offset should remain 0 when content fits viewport, got %d", sc.GetScrollOffset())
	}

	sc.AddChild(&scrollTestChild{view: "row3"})
	if sc.GetScrollOffset() != 1 {
		t.Fatalf("expected auto-scroll offset 1, got %d", sc.GetScrollOffset())
	}
}

func TestScrollContainerWindowSizeHandling(t *testing.T) {
	sc := NewScrollContainer(WithShowScrollbar(false), WithViewportHeight(2))
	child := &scrollTestChild{view: "a\nb\nc"}
	sc.AddChild(child)
	sc.Focus()
	sc.FocusChild(0)

	sc.Update(tea.WindowSizeMsg{Width: 44, Height: 9})

	if sc.viewportWidth != 44 || sc.viewportHeight != 9 {
		t.Fatalf("window size not applied, got width=%d height=%d", sc.viewportWidth, sc.viewportHeight)
	}
	if child.updateHits == 0 {
		t.Fatal("window size should be forwarded to focused child")
	}
}

func TestScrollContainerMouseWheel(t *testing.T) {
	sc := NewScrollContainer(WithShowScrollbar(false), WithViewportHeight(3))
	sc.viewportWidth = 20
	sc.Focus()
	sc.AddChild(&scrollTestChild{view: "1\n2\n3\n4\n5\n6"})

	sc.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
	if sc.GetScrollOffset() != 1 {
		t.Fatalf("expected offset 1 after wheel down, got %d", sc.GetScrollOffset())
	}

	sc.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp})
	if sc.GetScrollOffset() != 0 {
		t.Fatalf("expected offset 0 after wheel up, got %d", sc.GetScrollOffset())
	}
}
