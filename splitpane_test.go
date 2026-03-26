package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type mockSplitComponent struct {
	focused      bool
	width        int
	height       int
	view         string
	updateCount  int
	lastKey      string
	windowUpdate int
}

func (m *mockSplitComponent) Init() tea.Cmd { return nil }

func (m *mockSplitComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height
		m.windowUpdate++
	case tea.KeyMsg:
		m.lastKey = v.String()
		m.updateCount++
	default:
		m.updateCount++
	}
	return m, nil
}

func (m *mockSplitComponent) View() string { return m.view }
func (m *mockSplitComponent) Focus()       { m.focused = true }
func (m *mockSplitComponent) Blur()        { m.focused = false }
func (m *mockSplitComponent) Focused() bool {
	return m.focused
}

func TestSplitPaneCreation(t *testing.T) {
	s := NewSplitPane()
	if s == nil {
		t.Fatal("NewSplitPane returned nil")
	}

	if s.direction != SplitHorizontal {
		t.Errorf("expected default direction SplitHorizontal, got %v", s.direction)
	}
	if s.ratio != 0.5 {
		t.Errorf("expected default ratio 0.5, got %.2f", s.ratio)
	}
	if s.minRatio != 0.1 || s.maxRatio != 0.9 {
		t.Errorf("expected min/max 0.1/0.9, got %.2f/%.2f", s.minRatio, s.maxRatio)
	}
	if !s.showDivider {
		t.Error("expected showDivider=true by default")
	}
	if s.dividerChar != "│" {
		t.Errorf("expected default dividerChar │, got %q", s.dividerChar)
	}
}

func TestSplitPaneOptions(t *testing.T) {
	left := &mockSplitComponent{}
	right := &mockSplitComponent{}

	s := NewSplitPane(
		WithDirection(SplitVertical),
		WithRatio(0.7),
		WithMinMaxRatio(0.2, 0.8),
		WithLeftComponent(left),
		WithRightComponent(right),
		WithShowDivider(false),
		WithDividerChar("#"),
	)

	if s.direction != SplitVertical {
		t.Error("direction option not applied")
	}
	if s.ratio != 0.7 {
		t.Errorf("expected ratio 0.7, got %.2f", s.ratio)
	}
	if s.minRatio != 0.2 || s.maxRatio != 0.8 {
		t.Errorf("expected min/max 0.2/0.8, got %.2f/%.2f", s.minRatio, s.maxRatio)
	}
	if s.left != left || s.right != right {
		t.Error("left/right component options not applied")
	}
	if s.showDivider {
		t.Error("showDivider option not applied")
	}
	if s.dividerChar != "#" {
		t.Errorf("expected dividerChar #, got %q", s.dividerChar)
	}
}

func TestSplitPaneRatioClamping(t *testing.T) {
	s := NewSplitPane(WithMinMaxRatio(0.2, 0.8))

	s.SetRatio(0.95)
	if s.GetRatio() != 0.8 {
		t.Errorf("expected ratio clamped to 0.8, got %.2f", s.GetRatio())
	}

	s.SetRatio(0.05)
	if s.GetRatio() != 0.2 {
		t.Errorf("expected ratio clamped to 0.2, got %.2f", s.GetRatio())
	}

	s.SetRatio(0.45)
	if s.GetRatio() != 0.45 {
		t.Errorf("expected ratio 0.45, got %.2f", s.GetRatio())
	}
}

func TestSplitPaneFocusSwitching(t *testing.T) {
	left := &mockSplitComponent{}
	right := &mockSplitComponent{}
	s := NewSplitPane(WithLeftComponent(left), WithRightComponent(right))

	s.Focus()
	if s.GetFocusedPane() != 0 {
		t.Errorf("expected focused pane 0, got %d", s.GetFocusedPane())
	}
	if !left.Focused() || right.Focused() {
		t.Error("expected left focused and right blurred")
	}

	s.FocusRight()
	if s.GetFocusedPane() != 1 {
		t.Errorf("expected focused pane 1, got %d", s.GetFocusedPane())
	}
	if !right.Focused() || left.Focused() {
		t.Error("expected right focused and left blurred")
	}

	s.ToggleFocus()
	if s.GetFocusedPane() != 0 {
		t.Errorf("expected focused pane 0 after toggle, got %d", s.GetFocusedPane())
	}

	s.FocusTop()
	if s.GetFocusedPane() != 0 {
		t.Error("FocusTop should alias FocusLeft")
	}
	s.FocusBottom()
	if s.GetFocusedPane() != 1 {
		t.Error("FocusBottom should alias FocusRight")
	}
}

func TestSplitPaneResizeMode(t *testing.T) {
	s := NewSplitPane()
	s.Focus()

	if s.IsResizing() {
		t.Error("resizing should be false by default")
	}

	s.EnterResizeMode()
	if !s.IsResizing() {
		t.Error("EnterResizeMode should enable resizing")
	}

	s.ExitResizeMode()
	if s.IsResizing() {
		t.Error("ExitResizeMode should disable resizing")
	}

	s.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	if !s.IsResizing() {
		t.Error("ctrl+r should toggle resizing on")
	}
	s.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	if s.IsResizing() {
		t.Error("ctrl+r should toggle resizing off")
	}

	s.EnterResizeMode()
	s.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if s.IsResizing() {
		t.Error("esc should exit resize mode")
	}
}

func TestSplitPaneResizeKeyboardAdjustments(t *testing.T) {
	s := NewSplitPane(WithMinMaxRatio(0.1, 0.9), WithRatio(0.5))
	s.Focus()
	s.EnterResizeMode()

	s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if s.GetRatio() != 0.55 {
		t.Errorf("expected ratio 0.55 after l, got %.2f", s.GetRatio())
	}

	s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if s.GetRatio() != 0.50 {
		t.Errorf("expected ratio 0.50 after h, got %.2f", s.GetRatio())
	}

	s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	if s.GetRatio() != 0.55 {
		t.Errorf("expected ratio 0.55 after +, got %.2f", s.GetRatio())
	}

	for i := 0; i < 20; i++ {
		s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	}
	if s.GetRatio() != 0.1 {
		t.Errorf("expected ratio clamped at 0.1, got %.2f", s.GetRatio())
	}
}

func TestSplitPaneComponentDelegation(t *testing.T) {
	left := &mockSplitComponent{}
	right := &mockSplitComponent{}
	s := NewSplitPane(WithLeftComponent(left), WithRightComponent(right))
	s.Focus()

	s.Update(tea.WindowSizeMsg{Width: 100, Height: 20})
	if left.windowUpdate == 0 || right.windowUpdate == 0 {
		t.Error("window size should be delegated to both child panes")
	}

	s.FocusLeft()
	s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if left.lastKey != "x" {
		t.Error("focused left pane should receive delegated key")
	}
	if right.lastKey == "x" {
		t.Error("unfocused right pane should not receive delegated key")
	}

	s.FocusRight()
	s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if right.lastKey != "y" {
		t.Error("focused right pane should receive delegated key")
	}
}

func TestSplitPaneNilComponentsAndView(t *testing.T) {
	s := NewSplitPane()
	s.Focus()
	s.Update(tea.WindowSizeMsg{Width: 20, Height: 4})

	view := s.View()
	if view == "" {
		t.Error("view should render even when child components are nil")
	}

	lines := strings.Split(view, "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d", len(lines))
	}
	for _, line := range lines {
		if len(line) < 20 {
			t.Errorf("expected each line to fill width, got len=%d line=%q", len(line), line)
		}
	}
}

func TestSplitPaneDirectionAndAliases(t *testing.T) {
	top := &mockSplitComponent{}
	bottom := &mockSplitComponent{}
	s := NewSplitPane(
		WithDirection(SplitVertical),
		WithTopComponent(top),
		WithBottomComponent(bottom),
	)

	if s.direction != SplitVertical {
		t.Error("expected vertical direction")
	}
	if s.left != top || s.right != bottom {
		t.Error("top/bottom aliases should set left/right components")
	}

	s.Update(tea.WindowSizeMsg{Width: 30, Height: 10})
	view := s.View()
	if view == "" {
		t.Error("vertical view should not be empty")
	}
}

func TestSplitPaneSetComponents(t *testing.T) {
	s := NewSplitPane()
	left := &mockSplitComponent{view: "left"}
	right := &mockSplitComponent{view: "right"}

	s.SetLeftComponent(left)
	s.SetRightComponent(right)

	if s.left != left || s.right != right {
		t.Error("SetLeftComponent/SetRightComponent should update children")
	}
}
