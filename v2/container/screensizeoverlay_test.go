package container

import (
	"strings"
	"testing"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	tea "github.com/charmbracelet/bubbletea"
)

var _ tui.Component = (*ScreenSizeOverlay)(nil)

type mockOverlayChild struct {
	focused     bool
	initCalled  bool
	updateCount int
	lastMsg     tea.Msg
	view        string
}

func (m *mockOverlayChild) Init() tea.Cmd {
	m.initCalled = true
	return func() tea.Msg { return "child-init" }
}

func (m *mockOverlayChild) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	m.updateCount++
	m.lastMsg = msg
	return m, func() tea.Msg { return "child-update" }
}

func (m *mockOverlayChild) View() string {
	if m.view == "" {
		return "child-view"
	}
	return m.view
}

func (m *mockOverlayChild) Focus() { m.focused = true }
func (m *mockOverlayChild) Blur()  { m.focused = false }
func (m *mockOverlayChild) Focused() bool {
	return m.focused
}

func TestScreenSizeOverlayCreationAndOptions(t *testing.T) {
	child := &mockOverlayChild{}
	o := NewScreenSizeOverlay(
		child,
		WithScreenSizeMinWidth(100),
		WithScreenSizeMinHeight(30),
		WithScreenSizeMessage("Need bigger terminal"),
		WithScreenSizeDesignTokens(design.DefaultTheme()),
	)

	if o == nil {
		t.Fatal("NewScreenSizeOverlay returned nil")
	}
	if o.child != child {
		t.Error("expected child to be set")
	}
	if o.minWidth != 100 || o.minHeight != 30 {
		t.Errorf("expected min size 100x30, got %dx%d", o.minWidth, o.minHeight)
	}
	if o.message != "Need bigger terminal" {
		t.Errorf("expected custom message, got %q", o.message)
	}
	if o.designTokens == nil {
		t.Error("expected design tokens to be set")
	}

	nilChild := NewScreenSizeOverlay(nil)
	if nilChild.child == nil {
		t.Fatal("expected nil child to be replaced with empty component")
	}
	if _, ok := nilChild.child.(*screenSizeOverlayEmptyComponent); !ok {
		t.Fatalf("expected screenSizeOverlayEmptyComponent, got %T", nilChild.child)
	}
}

func TestScreenSizeOverlayInitUpdateAndView(t *testing.T) {
	child := &mockOverlayChild{view: "actual app"}
	o := NewScreenSizeOverlay(child, WithScreenSizeMinWidth(60), WithScreenSizeMinHeight(20))

	if cmd := o.Init(); cmd == nil {
		t.Error("expected Init to forward child init command")
	}
	if !child.initCalled {
		t.Error("expected child Init to be called")
	}

	// Too small: should not forward updates to child and should render overlay.
	o.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	_, cmd := o.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if cmd != nil {
		t.Error("expected nil cmd when overlay is active")
	}
	if child.updateCount != 0 {
		t.Errorf("expected child not to receive updates while too small, got %d", child.updateCount)
	}
	overlayView := o.View()
	if strings.TrimSpace(overlayView) == "" {
		t.Fatal("expected overlay view to be non-empty when too small")
	}
	if !strings.Contains(overlayView, "Expand window to view") {
		t.Error("expected default overlay message")
	}
	if !strings.Contains(overlayView, "Current: 40x10") {
		t.Error("expected current size details in overlay")
	}

	// Large enough: should forward updates and render child.
	o.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	_, cmd = o.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Error("expected child command when minimum size is satisfied")
	}
	if child.updateCount == 0 {
		t.Error("expected child to receive updates when size is large enough")
	}
	if o.View() != "actual app" {
		t.Errorf("expected child view when not blocked, got %q", o.View())
	}
}

func TestScreenSizeOverlayFocusAndWindowSizeTracking(t *testing.T) {
	child := &mockOverlayChild{}
	o := NewScreenSizeOverlay(child)

	o.Focus()
	if !o.Focused() || !child.Focused() {
		t.Error("expected Focus to apply to overlay and child")
	}
	o.Blur()
	if o.Focused() || child.Focused() {
		t.Error("expected Blur to apply to overlay and child")
	}

	o.Update(tea.WindowSizeMsg{Width: 77, Height: 22})
	if o.width != 77 || o.height != 22 {
		t.Errorf("expected overlay dimensions to be updated, got %dx%d", o.width, o.height)
	}
}
