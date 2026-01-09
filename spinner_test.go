package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestSpinnerAnimation verifies that streaming toolblocks animate their spinner
func TestSpinnerAnimation(t *testing.T) {
	// Create a streaming toolblock
	block := NewToolBlock(
		"Bash",
		"go test",
		[]string{},
		WithStreaming(),
	)

	// Initialize it to start the tick
	block.Init()

	// Simulate window size
	block.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Get initial view with spinner position
	initialView := block.View()
	initialSpinner := block.spinner

	// Simulate a few tick messages
	for i := 0; i < 5; i++ {
		msg := toolBlockTickMsg{id: block}
		block.Update(msg)
	}

	// Get final view
	finalView := block.View()
	finalSpinner := block.spinner

	// Verify spinner advanced
	if initialSpinner == finalSpinner {
		t.Errorf("Spinner did not advance: initial=%d, final=%d", initialSpinner, finalSpinner)
	}

	// Verify views are different (spinner character changed)
	if initialView == finalView {
		t.Error("View did not change after tick messages")
	}

	fmt.Printf("✓ Spinner animation test passed (initial spinner: %d, final spinner: %d)\n", initialSpinner, finalSpinner)
}
