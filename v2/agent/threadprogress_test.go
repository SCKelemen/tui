package agent

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestThreadProgressCreation(t *testing.T) {
	tp := NewThreadProgress()
	if tp == nil {
		t.Fatal("NewThreadProgress returned nil")
	}

	if tp.threadSpinner.FrameCount() == 0 {
		t.Error("default spinner should have frames")
	}

	if tp.iconSet.Success == "" {
		t.Error("default icon set should be configured")
	}

	if !tp.showElapsed {
		t.Error("showElapsed should default to true")
	}

	if len(tp.threads) != 0 {
		t.Errorf("expected 0 threads initially, got %d", len(tp.threads))
	}
}

func TestThreadProgressOptions(t *testing.T) {
	tp := NewThreadProgress(
		WithThreadSpinner(SpinnerCircleQuarters),
		WithThreadIconSet(IconSetCodex),
		WithMaxOutputLines(2),
		WithShowElapsed(false),
	)

	if tp.threadSpinner.FrameCount() != SpinnerCircleQuarters.FrameCount() {
		t.Error("WithThreadSpinner was not applied")
	}

	if tp.iconSet.Success != IconSetCodex.Success {
		t.Error("WithThreadIconSet was not applied")
	}

	if tp.maxOutputLines != 2 {
		t.Errorf("expected maxOutputLines=2, got %d", tp.maxOutputLines)
	}

	if tp.showElapsed {
		t.Error("WithShowElapsed(false) was not applied")
	}
}

func TestThreadProgressLifecycle(t *testing.T) {
	tp := NewThreadProgress()

	if cmd := tp.Init(); cmd != nil {
		t.Error("Init should return nil with no running threads")
	}

	tp.AddThread("t1", "worker 1")
	if cmd := tp.Init(); cmd == nil {
		t.Error("Init should return tick command when a thread is running")
	}

	tp.UpdateThread("t1", ThreadCompleted)
	if tp.hasRunningThreads() {
		t.Error("hasRunningThreads should be false after completion")
	}
}

func TestThreadProgressAddGetRemoveThread(t *testing.T) {
	tp := NewThreadProgress()
	tp.AddThread("t1", "worker 1")

	thread := tp.GetThread("t1")
	if thread == nil {
		t.Fatal("GetThread should return a thread")
	}

	if thread.Status != ThreadRunning {
		t.Errorf("expected default status running, got %v", thread.Status)
	}

	tp.RemoveThread("t1")
	if tp.GetThread("t1") != nil {
		t.Error("thread should be removed")
	}
}

func TestThreadProgressOutputAppendAndView(t *testing.T) {
	tp := NewThreadProgress(WithMaxOutputLines(2))
	tp.AddThread("t1", "worker 1")
	tp.SetThreadExpanded("t1", true)
	tp.AppendOutput("t1", "line 1")
	tp.AppendOutput("t1", "line 2")
	tp.AppendOutput("t1", "line 3")
	tp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := tp.View()
	if !strings.Contains(view, "line 1") || !strings.Contains(view, "line 2") {
		t.Error("view should contain first output lines")
	}
	if strings.Contains(view, "line 3") {
		t.Error("view should hide output beyond max lines")
	}
	if !strings.Contains(view, "+1 lines") {
		t.Error("view should show hidden line indicator")
	}
	if !strings.Contains(view, "├──") || !strings.Contains(view, "└──") {
		t.Error("view should contain tree connectors")
	}
}

func TestThreadProgressExpandCollapse(t *testing.T) {
	tp := NewThreadProgress()
	tp.AddThread("t1", "worker 1")
	tp.AddThread("t2", "worker 2")

	tp.SetThreadExpanded("t1", true)
	if !tp.GetThread("t1").Expanded {
		t.Error("SetThreadExpanded(true) should expand")
	}

	tp.Focus()
	tp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if tp.GetThread("t1").Expanded || tp.GetThread("t2").Expanded {
		t.Error("key 'c' should collapse all")
	}

	tp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if !tp.GetThread("t1").Expanded || !tp.GetThread("t2").Expanded {
		t.Error("key 'a' should expand all")
	}
}

func TestThreadProgressKeyboardNavigation(t *testing.T) {
	tp := NewThreadProgress()
	tp.AddThread("t1", "worker 1")
	tp.AddThread("t2", "worker 2")
	tp.AddThread("t3", "worker 3")
	tp.Focus()

	if tp.selected != 0 {
		t.Errorf("expected initial selected=0, got %d", tp.selected)
	}

	tp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if tp.selected != 1 {
		t.Errorf("expected selected=1 after j, got %d", tp.selected)
	}

	tp.Update(tea.KeyMsg{Type: tea.KeyDown})
	if tp.selected != 2 {
		t.Errorf("expected selected=2 after down, got %d", tp.selected)
	}

	tp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if tp.selected != 1 {
		t.Errorf("expected selected=1 after k, got %d", tp.selected)
	}

	// enter toggles current thread expanded state
	current := tp.threads[tp.selected]
	if current.Expanded {
		t.Error("thread should start collapsed")
	}
	tp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !current.Expanded {
		t.Error("enter should toggle expanded")
	}
}

func TestThreadProgressStatusTransitions(t *testing.T) {
	tp := NewThreadProgress(WithThreadIconSet(IconSetCodex))
	tp.AddThread("t1", "worker 1")
	thread := tp.GetThread("t1")
	if thread == nil {
		t.Fatal("thread should exist")
	}

	tp.UpdateThread("t1", ThreadFailed)
	if thread.Status != ThreadFailed {
		t.Errorf("expected failed status, got %v", thread.Status)
	}
	if thread.EndTime.IsZero() {
		t.Error("failed thread should have end time")
	}

	tp.UpdateThread("t1", ThreadEscalated)
	if thread.Status != ThreadEscalated {
		t.Errorf("expected escalated status, got %v", thread.Status)
	}

	tp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := tp.View()
	if !strings.Contains(view, IconSetCodex.Warning) {
		t.Error("escalated status should render warning icon")
	}

	cmd := tp.StartThread("t1")
	if cmd == nil {
		t.Error("StartThread should return a tick command")
	}
	if thread.Status != ThreadRunning {
		t.Error("StartThread should set running status")
	}
}

func TestThreadProgressTickSpinnerUpdate(t *testing.T) {
	tp := NewThreadProgress(WithThreadSpinner(SpinnerCircleQuarters))
	tp.AddThread("t1", "worker 1")
	thread := tp.GetThread("t1")
	if thread == nil {
		t.Fatal("thread should exist")
	}

	initial := thread.spinIdx
	component, cmd := tp.Update(threadProgressTickMsg{})
	if cmd == nil {
		t.Error("tick update should return follow-up tick command")
	}
	updated := component.(*ThreadProgress)
	if updated.GetThread("t1").spinIdx == initial {
		t.Error("spin index should advance on tick")
	}
}

func TestThreadProgressFocusManagement(t *testing.T) {
	tp := NewThreadProgress()
	if tp.Focused() {
		t.Error("should not be focused initially")
	}

	tp.Focus()
	if !tp.Focused() {
		t.Error("should be focused after Focus")
	}

	tp.Blur()
	if tp.Focused() {
		t.Error("should not be focused after Blur")
	}
}

func TestThreadProgressElapsedFormatting(t *testing.T) {
	tp := NewThreadProgress()

	if got := tp.formatDuration(14 * time.Second); got != "14s" {
		t.Errorf("expected 14s, got %q", got)
	}

	if got := tp.formatDuration(74 * time.Second); got != "1m 14s" {
		t.Errorf("expected 1m 14s, got %q", got)
	}
}

func TestThreadProgressViewWithoutWidth(t *testing.T) {
	tp := NewThreadProgress()
	tp.AddThread("t1", "worker 1")

	if view := tp.View(); view != "" {
		t.Error("view should be empty when width is not set")
	}
}

func TestThreadProgressViewWithoutElapsed(t *testing.T) {
	tp := NewThreadProgress(WithShowElapsed(false))
	tp.AddThread("t1", "worker 1")
	tp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := tp.View()
	if strings.Contains(view, "(") && strings.Contains(view, "s)") {
		t.Error("elapsed time should not render when disabled")
	}
}
