package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestStatusBarCreation tests that a status bar can be created
func TestStatusBarCreation(t *testing.T) {
	statusBar := NewStatusBar()

	if statusBar == nil {
		t.Fatal("Failed to create status bar")
	}

	if statusBar.message != "Ready" {
		t.Errorf("Expected default message='Ready', got '%s'", statusBar.message)
	}

	if statusBar.width != 0 {
		t.Errorf("Expected width=0, got %d", statusBar.width)
	}

	if statusBar.focused {
		t.Error("Status bar should not be focused initially")
	}
}

// TestStatusBarInit tests initialization
func TestStatusBarInit(t *testing.T) {
	statusBar := NewStatusBar()
	cmd := statusBar.Init()

	if cmd != nil {
		t.Error("Init should return nil command")
	}
}

// TestStatusBarWindowSizeUpdate tests window size handling
func TestStatusBarWindowSizeUpdate(t *testing.T) {
	statusBar := NewStatusBar()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	statusBar.Update(msg)

	if statusBar.width != 80 {
		t.Errorf("Expected width=80, got %d", statusBar.width)
	}
}

// TestStatusBarViewBasic tests basic rendering
func TestStatusBarViewBasic(t *testing.T) {
	statusBar := NewStatusBar()

	statusBar.width = 80
	view := statusBar.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	// Should contain default message
	if !strings.Contains(view, "Ready") {
		t.Error("View should contain 'Ready' message")
	}

	// Should contain keybindings
	if !strings.Contains(view, "Tab: Focus") {
		t.Error("View should contain 'Tab: Focus' keybinding")
	}

	if !strings.Contains(view, "q: Quit") {
		t.Error("View should contain 'q: Quit' keybinding")
	}
}

// TestStatusBarViewWithoutSize tests view before size is set
func TestStatusBarViewWithoutSize(t *testing.T) {
	statusBar := NewStatusBar()

	// Width is 0
	view := statusBar.View()

	if view != "" {
		t.Error("View should be empty without size")
	}
}

// TestStatusBarViewWithFocus tests focused styling
func TestStatusBarViewWithFocus(t *testing.T) {
	statusBar := NewStatusBar()
	statusBar.width = 80
	statusBar.Focus()

	view := statusBar.View()

	// Should contain inverted color ANSI code (7m)
	if !strings.Contains(view, "\033[7m") {
		t.Error("Focused status bar should have inverted colors")
	}
}

// TestStatusBarViewWithoutFocus tests unfocused styling
func TestStatusBarViewWithoutFocus(t *testing.T) {
	statusBar := NewStatusBar()
	statusBar.width = 80
	statusBar.Blur()

	view := statusBar.View()

	// Should contain dimmed ANSI code (2m)
	if !strings.Contains(view, "\033[2m") {
		t.Error("Unfocused status bar should be dimmed")
	}
}

// TestStatusBarFocusManagement tests focus state management
func TestStatusBarFocusManagement(t *testing.T) {
	statusBar := NewStatusBar()

	if statusBar.Focused() {
		t.Error("Status bar should not be focused initially")
	}

	statusBar.Focus()
	if !statusBar.Focused() {
		t.Error("Status bar should be focused after Focus()")
	}

	statusBar.Blur()
	if statusBar.Focused() {
		t.Error("Status bar should not be focused after Blur()")
	}
}

// TestStatusBarSetMessage tests setting custom message
func TestStatusBarSetMessage(t *testing.T) {
	statusBar := NewStatusBar()
	statusBar.width = 80

	statusBar.SetMessage("Processing...")

	if statusBar.message != "Processing..." {
		t.Errorf("Expected message='Processing...', got '%s'", statusBar.message)
	}

	view := statusBar.View()
	if !strings.Contains(view, "Processing...") {
		t.Error("View should contain custom message")
	}
}

// TestStatusBarSetMessageEmpty tests empty message
func TestStatusBarSetMessageEmpty(t *testing.T) {
	statusBar := NewStatusBar()
	statusBar.width = 80

	statusBar.SetMessage("")

	if statusBar.message != "" {
		t.Errorf("Expected empty message, got '%s'", statusBar.message)
	}

	view := statusBar.View()
	// Should still render with keybindings
	if !strings.Contains(view, "Tab: Focus") {
		t.Error("View should still contain keybindings with empty message")
	}
}

// TestStatusBarLongMessage tests message truncation
func TestStatusBarLongMessage(t *testing.T) {
	statusBar := NewStatusBar()
	statusBar.width = 40 // Narrow width

	// Very long message
	longMessage := "This is a very long status message that should be truncated"
	statusBar.SetMessage(longMessage)

	view := statusBar.View()

	// Should contain truncation indicator
	if !strings.Contains(view, "...") {
		t.Error("Long message should be truncated with ellipsis")
	}
}

// TestStatusBarVeryNarrowWidth tests rendering with very narrow width
func TestStatusBarVeryNarrowWidth(t *testing.T) {
	statusBar := NewStatusBar()
	statusBar.width = 50 // Narrow but reasonable

	statusBar.SetMessage("This is a longer test message")
	view := statusBar.View()

	if view == "" {
		t.Error("View should not be empty even with narrow width")
	}

	// Should handle narrow width gracefully
	lines := strings.Split(view, "\n")
	if len(lines) > 0 && len(strings.TrimSpace(lines[0])) > 0 {
		// Should have some content
		if !strings.Contains(view, "...") && !strings.Contains(view, "This is a longer") {
			t.Error("View should contain message or truncation")
		}
	}
}

// TestStatusBarWideWidth tests rendering with wide width
func TestStatusBarWideWidth(t *testing.T) {
	statusBar := NewStatusBar()
	statusBar.width = 200 // Very wide

	statusBar.SetMessage("Status")
	view := statusBar.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	// Should contain message on left
	if !strings.Contains(view, "Status") {
		t.Error("View should contain message")
	}

	// Should contain keybindings on right
	if !strings.Contains(view, "Tab: Focus") {
		t.Error("View should contain keybindings")
	}
}

// TestStatusBarMessageWithSpecialCharacters tests special characters
func TestStatusBarMessageWithSpecialCharacters(t *testing.T) {
	statusBar := NewStatusBar()
	statusBar.width = 80

	// Message with special characters
	statusBar.SetMessage("File: /path/to/file.txt • 42% complete")

	view := statusBar.View()
	if !strings.Contains(view, "File: /path/to/file.txt") {
		t.Error("View should contain message with special characters")
	}
}

// TestStatusBarMultipleUpdates tests multiple message updates
func TestStatusBarMultipleUpdates(t *testing.T) {
	statusBar := NewStatusBar()
	statusBar.width = 80

	messages := []string{"Loading...", "Processing...", "Complete!"}

	for _, msg := range messages {
		statusBar.SetMessage(msg)

		if statusBar.message != msg {
			t.Errorf("Expected message='%s', got '%s'", msg, statusBar.message)
		}

		view := statusBar.View()
		if !strings.Contains(view, msg) {
			t.Errorf("View should contain message '%s'", msg)
		}
	}
}

// TestStatusBarViewStructure tests the structure of the rendered view
func TestStatusBarViewStructure(t *testing.T) {
	statusBar := NewStatusBar()
	statusBar.width = 80
	statusBar.SetMessage("Test Message")

	view := statusBar.View()

	// Should end with newline
	if !strings.HasSuffix(view, "\n") {
		t.Error("View should end with newline")
	}

	// Should have ANSI codes
	if !strings.Contains(view, "\033[") {
		t.Error("View should contain ANSI escape codes")
	}

	// Should have reset code
	if !strings.Contains(view, "\033[0m") {
		t.Error("View should contain ANSI reset code")
	}
}

// TestStatusBarSpacing tests spacing calculation
func TestStatusBarSpacing(t *testing.T) {
	statusBar := NewStatusBar()
	statusBar.width = 100

	statusBar.SetMessage("Left")
	view := statusBar.View()

	// Remove ANSI codes for length check
	stripped := strings.ReplaceAll(view, "\033[2m", "")
	stripped = strings.ReplaceAll(stripped, "\033[7m", "")
	stripped = strings.ReplaceAll(stripped, "\033[0m", "")
	stripped = strings.TrimSuffix(stripped, "\n")

	// Should have spacing between left and right
	if !strings.Contains(stripped, "  ") {
		t.Error("View should have spacing between message and keybindings")
	}
}

// TestStatusBarUpdateWithOtherMessages tests that other messages don't affect state
func TestStatusBarUpdateWithOtherMessages(t *testing.T) {
	statusBar := NewStatusBar()
	statusBar.width = 80

	// Send a KeyMsg (should be ignored)
	statusBar.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Width should remain unchanged
	if statusBar.width != 80 {
		t.Error("Width should not change with non-WindowSizeMsg")
	}
}

// TestStatusBarToggleFocus tests toggling focus multiple times
func TestStatusBarToggleFocus(t *testing.T) {
	statusBar := NewStatusBar()
	statusBar.width = 80

	// Toggle focus multiple times
	for i := 0; i < 5; i++ {
		statusBar.Focus()
		if !statusBar.Focused() {
			t.Errorf("Iteration %d: should be focused", i)
		}

		statusBar.Blur()
		if statusBar.Focused() {
			t.Errorf("Iteration %d: should not be focused", i)
		}
	}
}

// TestStatusBarEmptyWidthAfterSetting tests setting width to 0
func TestStatusBarEmptyWidthAfterSetting(t *testing.T) {
	statusBar := NewStatusBar()
	statusBar.width = 80

	// Set to valid width first
	view := statusBar.View()
	if view == "" {
		t.Error("View should not be empty with width=80")
	}

	// Set width to 0
	msg := tea.WindowSizeMsg{Width: 0, Height: 24}
	statusBar.Update(msg)

	view = statusBar.View()
	if view != "" {
		t.Error("View should be empty with width=0")
	}
}
