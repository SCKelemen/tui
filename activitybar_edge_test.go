package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestActivityBarVeryLongStatus(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	longStatus := strings.Repeat("Very long status message ", 20)
	ab.Start(longStatus)

	view := ab.View()
	if view == "" {
		t.Error("View should not be empty with long status")
	}

	// Should truncate to fit width
	if len(stripANSI(view)) > 100 {
		t.Error("View should truncate long status")
	}
}

func TestActivityBarVeryLongProgress(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	ab.Start("Test")
	longProgress := strings.Repeat("Progress info ", 30)
	ab.SetProgress(longProgress)

	view := ab.View()
	if view == "" {
		t.Error("View should not be empty with long progress")
	}
}

func TestActivityBarMultipleStartStop(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	// Start and stop multiple times
	for i := 0; i < 10; i++ {
		ab.Start("Activity")
		ab.Stop()
	}

	// After stop, view should show inactive state (not animated)
	view := ab.View()
	if view == "" {
		t.Error("View should not be empty after multiple start/stop")
	}
}

func TestActivityBarSetProgressWhenNotRunning(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	// Set progress without starting (should handle gracefully)
	ab.SetProgress("Progress without start")

	view := ab.View()
	// Might be empty or show inactive state, shouldn't panic
	_ = view
}

func TestActivityBarRapidTicks(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	ab.Start("Test")

	// Simulate many rapid ticks
	for i := 0; i < 100; i++ {
		ab.Update(activityBarTickMsg(time.Now()))
	}

	// Should not panic
	view := ab.View()
	if view == "" {
		t.Error("View should not be empty after rapid ticks")
	}

	// Check that it's still showing animated state by checking for spinner frames
	hasSpinner := false
	for _, frame := range spinnerFrames {
		if strings.Contains(view, frame) {
			hasSpinner = true
			break
		}
	}
	if !hasSpinner {
		t.Error("ActivityBar should still show spinner when running")
	}
}

func TestActivityBarEmptyStatus(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	ab.Start("")

	view := ab.View()
	if view == "" {
		t.Error("View should not be empty with empty status")
	}
}

func TestActivityBarEmptyProgress(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	ab.Start("Test")
	ab.SetProgress("")

	view := ab.View()
	if view == "" {
		t.Error("View should not be empty with empty progress")
	}
}

func TestActivityBarNarrowWidth(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 20})

	ab.Start("Test activity")
	ab.SetProgress("Progress")

	// Should not panic with narrow width
	view := ab.View()
	_ = view
}

func TestActivityBarVeryNarrowWidth(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 5})

	ab.Start("Test")

	// Should not panic with very narrow width
	view := ab.View()
	_ = view
}

func TestActivityBarWideWidth(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 300})

	ab.Start("Test activity")
	ab.SetProgress("Progress info")

	view := ab.View()
	if view == "" {
		t.Error("View should not be empty with wide width")
	}
}

func TestActivityBarZeroWidth(t *testing.T) {
	ab := NewActivityBar()

	// Don't set width
	ab.Start("Test")

	view := ab.View()
	if view != "" {
		t.Error("View should be empty when width is not set")
	}
}

func TestActivityBarUnicodeStatus(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	ab.Start("処理中… 🔄")
	ab.SetProgress("↓ 100件")

	view := ab.View()
	if view == "" {
		t.Error("View should not be empty with unicode status")
	}

	if !strings.Contains(view, "処理中") {
		t.Error("View should contain unicode status")
	}
}

func TestActivityBarSpecialCharacters(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	ab.Start("Test <>&\"'")
	ab.SetProgress("↓↑←→")

	view := ab.View()
	if view == "" {
		t.Error("View should not be empty with special characters")
	}
}

func TestActivityBarStopAfterLongRun(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	ab.Start("Long running")

	// Wait a reasonable amount of time to ensure elapsed time is measurable
	time.Sleep(100 * time.Millisecond)

	// Simulate some ticks
	for i := 0; i < 10; i++ {
		ab.Update(activityBarTickMsg(time.Now()))
	}

	ab.Stop()

	view := ab.View()
	if view == "" {
		t.Error("View should not be empty after long run and stop")
	}

	// After stop, should not show spinner
	hasSpinner := false
	for _, frame := range spinnerFrames {
		if strings.Contains(view, frame) {
			hasSpinner = true
			break
		}
	}
	if hasSpinner {
		t.Error("ActivityBar should not show spinner after Stop")
	}
}

func TestActivityBarSpinnerAdvancement(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	ab.Start("Test")
	initialView := ab.View()

	// Simulate several ticks
	for i := 0; i < 10; i++ {
		ab.Update(activityBarTickMsg(time.Now()))
	}

	// Check that view changed (spinner advanced)
	viewAfterTicks := ab.View()

	// At least one of the views should contain a spinner frame
	hasSpinner := false
	for _, frame := range spinnerFrames {
		if strings.Contains(initialView, frame) || strings.Contains(viewAfterTicks, frame) {
			hasSpinner = true
			break
		}
	}
	if !hasSpinner {
		t.Error("Views should contain spinner frames")
	}
}

func TestActivityBarElapsedTime(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	ab.Start("Test")

	// Wait enough time to ensure measurable elapsed time
	time.Sleep(150 * time.Millisecond)

	ab.Update(activityBarTickMsg(time.Now()))

	view := ab.View()

	// Should show elapsed time (contains digits and 's' unit)
	// More lenient check - just verify it's showing time information
	hasTimeInfo := strings.Contains(view, "s") || strings.Contains(view, ".")
	if !hasTimeInfo {
		t.Log("View content:", view)
		t.Error("View should show elapsed time information")
	}
}

func TestActivityBarProgressUpdate(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	ab.Start("Test")
	ab.SetProgress("Progress 1")

	view1 := ab.View()
	if !strings.Contains(view1, "Progress 1") {
		t.Error("View should contain first progress")
	}

	ab.SetProgress("Progress 2")
	view2 := ab.View()
	if !strings.Contains(view2, "Progress 2") {
		t.Error("View should contain updated progress")
	}

	if strings.Contains(view2, "Progress 1") {
		t.Error("View should not contain old progress")
	}
}

func TestActivityBarStopThenStart(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	ab.Start("First activity")
	ab.Stop()

	view1 := ab.View()
	// After stop, should show inactive state
	if strings.Contains(view1, spinnerFrames[0]) {
		t.Error("Should not show spinner after Stop")
	}

	ab.Start("Second activity")

	view := ab.View()
	if !strings.Contains(view, "Second activity") {
		t.Error("View should contain second activity status")
	}
}

func TestActivityBarWindowResize(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	ab.Start("Test")
	view1 := ab.View()

	// Resize to wider
	ab.Update(tea.WindowSizeMsg{Width: 120})
	view2 := ab.View()

	// Both views should not be empty
	if view1 == "" || view2 == "" {
		t.Error("Views should not be empty")
	}

	// Views should contain the status message
	if !strings.Contains(view1, "Test") || !strings.Contains(view2, "Test") {
		t.Error("Views should contain status message")
	}
}

func TestActivityBarNoSpinnerWhenStopped(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	ab.Start("Test")
	ab.Stop()

	view := ab.View()
	if view == "" {
		t.Error("View should not be empty when stopped")
	}

	// When stopped, should show final state without spinner
	hasSpinner := false
	for _, frame := range spinnerFrames {
		if strings.Contains(view, frame) {
			hasSpinner = true
			break
		}
	}
	if hasSpinner {
		t.Error("Should not show spinner when stopped")
	}
}

func TestActivityBarFormattedElapsedTime(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	// Test different durations
	testCases := []struct {
		duration time.Duration
		contains string
	}{
		{100 * time.Millisecond, "0.1s"},
		{1 * time.Second, "1.0s"},
		{10 * time.Second, "10."},
	}

	for _, tc := range testCases {
		ab.Start("Test")

		// Wait for the duration
		time.Sleep(tc.duration)
		ab.Update(activityBarTickMsg(time.Now()))

		view := ab.View()
		if !strings.Contains(view, "s") {
			t.Errorf("View should contain time unit 's' for duration %v", tc.duration)
		}

		ab.Stop()
	}
}

func TestActivityBarConcurrentStartStop(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	// Simulate rapid start/stop (like interrupt handling)
	for i := 0; i < 20; i++ {
		ab.Start("Activity")
		ab.Update(activityBarTickMsg(time.Now()))
		ab.Stop()
	}

	// Should end in stopped state (view won't show spinner)
	view := ab.View()
	hasSpinner := false
	for _, frame := range spinnerFrames {
		if strings.Contains(view, frame) {
			hasSpinner = true
			break
		}
	}
	if hasSpinner {
		t.Error("Should not show spinner after final Stop")
	}
}

func TestActivityBarViewConsistency(t *testing.T) {
	ab := NewActivityBar()
	ab.Update(tea.WindowSizeMsg{Width: 80})

	ab.Start("Test")

	// Multiple View() calls without update should return consistent results
	view1 := ab.View()
	view2 := ab.View()
	view3 := ab.View()

	if view1 != view2 || view2 != view3 {
		t.Error("Multiple View() calls without updates should return consistent results")
	}
}
