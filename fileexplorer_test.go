package tui

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFileExplorerDepthCalculation(t *testing.T) {
	// Get current directory for testing
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("Cannot get current directory")
	}

	// Create file explorer
	fe := NewFileExplorer(cwd)

	// Send window size message to initialize
	fe.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Get the root node's depth - should be 0, not negative
	if fe.root != nil {
		depth := fe.getDepth(fe.root)
		if depth < 0 {
			t.Errorf("Root node depth is negative: %d", depth)
		}

		// View should not panic
		view := fe.View()
		if view == "" {
			t.Error("View returned empty string")
		}
	}
}

func TestFileExplorerWithShowHidden(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("Cannot get current directory")
	}

	// Create with show hidden option
	fe := NewFileExplorer(cwd, WithShowHidden(true))
	if !fe.showHidden {
		t.Error("ShowHidden option not applied")
	}

	fe.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := fe.View()
	if view == "" {
		t.Error("View returned empty string")
	}
}

func TestFileExplorerNavigation(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("Cannot get current directory")
	}

	fe := NewFileExplorer(cwd)
	fe.Focus()
	fe.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	initialIndex := fe.selectedIndex

	// Try moving down
	fe.Update(tea.KeyMsg{Type: tea.KeyDown})
	if len(fe.visibleNodes) > 1 && fe.selectedIndex == initialIndex {
		t.Error("Down key did not move selection")
	}

	// Try moving up
	fe.Update(tea.KeyMsg{Type: tea.KeyUp})
	if fe.selectedIndex != initialIndex {
		t.Error("Up key did not move selection back")
	}
}
