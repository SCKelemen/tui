package tui

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestActivityBarSpinnerAnimation verifies that ActivityBar animates its spinner
func TestActivityBarSpinnerAnimation(t *testing.T) {
	// Create an activity bar
	bar := NewActivityBar()

	// Start the activity
	bar.Start("Testing...")

	// Simulate window size
	bar.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Get initial spinner position
	initialSpinner := bar.spinner

	// Simulate a few tick messages
	for i := 0; i < 5; i++ {
		msg := activityBarTickMsg{}
		bar.Update(msg)
	}

	// Get final spinner position
	finalSpinner := bar.spinner

	// Verify spinner advanced
	if initialSpinner == finalSpinner {
		t.Errorf("ActivityBar spinner did not advance: initial=%d, final=%d", initialSpinner, finalSpinner)
	}

	// Verify it's still active
	if !bar.active {
		t.Error("ActivityBar should still be active")
	}

	fmt.Printf("✓ ActivityBar spinner animation test passed (initial: %d, final: %d)\n", initialSpinner, finalSpinner)
}

// TestActivityBarWithApplication verifies ActivityBar spinner in an Application context
func TestActivityBarWithApplication(t *testing.T) {
	app := NewApplication()
	bar := NewActivityBar()
	app.AddComponent(bar)

	// Initialize
	app.Init()

	// Start the activity
	bar.Start("Processing...")

	// Simulate window size
	appModel, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app = appModel.(*Application)

	// Get view before any ticks
	viewBefore := app.View()

	// Simulate a tick message - should be broadcast to ActivityBar
	msg := activityBarTickMsg(time.Now())
	appModel, _ = app.Update(msg)
	app = appModel.(*Application)

	// Get view after tick
	viewAfter := app.View()

	// Verify view changed (spinner character updated)
	if viewBefore == viewAfter {
		t.Error("ActivityBar view should have changed after tick message")
	}

	// Access the ActivityBar from components to check spinner directly
	if len(app.components) > 0 {
		updatedBar := app.components[0].(*ActivityBar)
		if updatedBar.spinner == 0 {
			t.Error("Spinner should have advanced from 0")
		}
		fmt.Printf("✓ Spinner advanced to position %d\n", updatedBar.spinner)
	}

	fmt.Printf("✓ ActivityBar in Application animation test passed\n")
}
