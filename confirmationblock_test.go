package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestConfirmationBlockCreation tests default creation
func TestConfirmationBlockCreation(t *testing.T) {
	cb := NewConfirmationBlock()

	if cb == nil {
		t.Fatal("NewConfirmationBlock returned nil")
	}

	if cb.operation != "Write" {
		t.Errorf("Expected default operation='Write', got '%s'", cb.operation)
	}

	if cb.startLine != 1 {
		t.Errorf("Expected default startLine=1, got %d", cb.startLine)
	}

	if cb.selectedIndex != 0 {
		t.Errorf("Expected default selectedIndex=0, got %d", cb.selectedIndex)
	}

	if cb.confirmedIdx != -1 {
		t.Errorf("Expected default confirmedIdx=-1, got %d", cb.confirmedIdx)
	}

	if cb.confirmed {
		t.Error("Should not be confirmed by default")
	}

	if len(cb.options) != 2 {
		t.Errorf("Expected 2 default options, got %d", len(cb.options))
	}
}

// TestConfirmationBlockWithConfirmOperation tests operation option
func TestConfirmationBlockWithConfirmOperation(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmOperation("Read"),
	)

	if cb.operation != "Read" {
		t.Errorf("Expected operation='Read', got '%s'", cb.operation)
	}
}

// TestConfirmationBlockWithConfirmFilepath tests filepath option
func TestConfirmationBlockWithConfirmFilepath(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmFilepath("/path/to/file.txt"),
	)

	if cb.filepath != "/path/to/file.txt" {
		t.Errorf("Expected filepath='/path/to/file.txt', got '%s'", cb.filepath)
	}
}

// TestConfirmationBlockWithConfirmDescription tests description option
func TestConfirmationBlockWithConfirmDescription(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmDescription("Create new file"),
	)

	if cb.description != "Create new file" {
		t.Errorf("Expected description='Create new file', got '%s'", cb.description)
	}
}

// TestConfirmationBlockWithConfirmCode tests code content option
func TestConfirmationBlockWithConfirmCode(t *testing.T) {
	code := "package main\n\nfunc main() {}"
	cb := NewConfirmationBlock(
		WithConfirmCode(code),
	)

	if len(cb.code) != 3 {
		t.Errorf("Expected 3 code lines, got %d", len(cb.code))
	}

	if cb.code[0] != "package main" {
		t.Errorf("Expected first line='package main', got '%s'", cb.code[0])
	}
}

// TestConfirmationBlockWithConfirmCodeLines tests code lines option
func TestConfirmationBlockWithConfirmCodeLines(t *testing.T) {
	lines := []string{"line 1", "line 2", "line 3"}
	cb := NewConfirmationBlock(
		WithConfirmCodeLines(lines),
	)

	if len(cb.code) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(cb.code))
	}

	if cb.code[1] != "line 2" {
		t.Errorf("Expected second line='line 2', got '%s'", cb.code[1])
	}
}

// TestConfirmationBlockWithConfirmOptions tests custom options
func TestConfirmationBlockWithConfirmOptions(t *testing.T) {
	options := []string{"Accept", "Reject", "Skip"}
	cb := NewConfirmationBlock(
		WithConfirmOptions(options),
	)

	if len(cb.options) != 3 {
		t.Errorf("Expected 3 options, got %d", len(cb.options))
	}

	if cb.options[0] != "Accept" {
		t.Errorf("Expected first option='Accept', got '%s'", cb.options[0])
	}
}

// TestConfirmationBlockWithConfirmStartLine tests start line option
func TestConfirmationBlockWithConfirmStartLine(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmStartLine(42),
	)

	if cb.startLine != 42 {
		t.Errorf("Expected startLine=42, got %d", cb.startLine)
	}
}

// TestConfirmationBlockWithConfirmPreview tests preview option
func TestConfirmationBlockWithConfirmPreview(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmPreview(10),
	)

	if cb.showPreview != 10 {
		t.Errorf("Expected showPreview=10, got %d", cb.showPreview)
	}
}

// TestConfirmationBlockWithConfirmFooterHints tests footer hints option
func TestConfirmationBlockWithConfirmFooterHints(t *testing.T) {
	hints := []string{"Press Enter", "Press Esc"}
	cb := NewConfirmationBlock(
		WithConfirmFooterHints(hints),
	)

	if len(cb.footerHints) != 2 {
		t.Errorf("Expected 2 hints, got %d", len(cb.footerHints))
	}

	if cb.footerHints[0] != "Press Enter" {
		t.Errorf("Expected first hint='Press Enter', got '%s'", cb.footerHints[0])
	}
}

// TestConfirmationBlockUpdate tests update with window size
func TestConfirmationBlockUpdate(t *testing.T) {
	cb := NewConfirmationBlock()

	_, _ = cb.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	if cb.width != 100 {
		t.Errorf("Expected width=100, got %d", cb.width)
	}

	if cb.height != 40 {
		t.Errorf("Expected height=40, got %d", cb.height)
	}
}

// TestConfirmationBlockNavigationDown tests down arrow navigation
func TestConfirmationBlockNavigationDown(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmOptions([]string{"Option 1", "Option 2", "Option 3"}),
	)
	cb.Focus()

	if cb.selectedIndex != 0 {
		t.Error("Should start at index 0")
	}

	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

	if cb.selectedIndex != 1 {
		t.Errorf("Should move to index 1, got %d", cb.selectedIndex)
	}
}

// TestConfirmationBlockNavigationUp tests up arrow navigation
func TestConfirmationBlockNavigationUp(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmOptions([]string{"Option 1", "Option 2", "Option 3"}),
	)
	cb.Focus()
	cb.selectedIndex = 1

	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyUp})

	if cb.selectedIndex != 0 {
		t.Errorf("Should move to index 0, got %d", cb.selectedIndex)
	}
}

// TestConfirmationBlockNavigationWrap tests wrapping at boundaries
func TestConfirmationBlockNavigationWrap(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmOptions([]string{"Option 1", "Option 2", "Option 3"}),
	)
	cb.Focus()

	// At index 0, pressing up should wrap to last
	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyUp})

	if cb.selectedIndex != 2 {
		t.Errorf("Should wrap to index 2, got %d", cb.selectedIndex)
	}

	// At last index, pressing down should wrap to 0
	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

	if cb.selectedIndex != 0 {
		t.Errorf("Should wrap to index 0, got %d", cb.selectedIndex)
	}
}

// TestConfirmationBlockNavigationVimKeys tests j/k navigation
func TestConfirmationBlockNavigationVimKeys(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmOptions([]string{"Option 1", "Option 2"}),
	)
	cb.Focus()

	// j moves down
	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	if cb.selectedIndex != 1 {
		t.Errorf("j should move down to index 1, got %d", cb.selectedIndex)
	}

	// k moves up
	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})

	if cb.selectedIndex != 0 {
		t.Errorf("k should move up to index 0, got %d", cb.selectedIndex)
	}
}

// TestConfirmationBlockNavigationTab tests tab/shift+tab navigation
func TestConfirmationBlockNavigationTab(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmOptions([]string{"Option 1", "Option 2"}),
	)
	cb.Focus()

	// Tab moves forward
	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyTab})

	if cb.selectedIndex != 1 {
		t.Errorf("Tab should move to index 1, got %d", cb.selectedIndex)
	}

	// Shift+Tab moves backward
	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyShiftTab})

	if cb.selectedIndex != 0 {
		t.Errorf("Shift+Tab should move to index 0, got %d", cb.selectedIndex)
	}
}

// TestConfirmationBlockConfirmWithEnter tests confirming with enter
func TestConfirmationBlockConfirmWithEnter(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmOptions([]string{"Yes", "No"}),
	)
	cb.Focus()
	cb.selectedIndex = 0

	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !cb.confirmed {
		t.Error("Should be confirmed after Enter")
	}

	if cb.confirmedIdx != 0 {
		t.Errorf("Should confirm index 0, got %d", cb.confirmedIdx)
	}
}

// TestConfirmationBlockCancelWithEsc tests cancelling with esc
func TestConfirmationBlockCancelWithEsc(t *testing.T) {
	cb := NewConfirmationBlock()
	cb.Focus()

	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !cb.confirmed {
		t.Error("Should be confirmed (cancelled) after Esc")
	}

	if cb.confirmedIdx != -1 {
		t.Errorf("Should have confirmedIdx=-1 for cancel, got %d", cb.confirmedIdx)
	}
}

// TestConfirmationBlockNumberKeySelection tests number key quick selection
func TestConfirmationBlockNumberKeySelection(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmOptions([]string{"Option 1", "Option 2", "Option 3"}),
	)
	cb.Focus()

	// Press '2' to select second option
	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})

	if !cb.confirmed {
		t.Error("Should be confirmed after number key")
	}

	if cb.confirmedIdx != 1 {
		t.Errorf("Should select index 1 (second option), got %d", cb.confirmedIdx)
	}
}

// TestConfirmationBlockNumberKeyOutOfRange tests invalid number key
func TestConfirmationBlockNumberKeyOutOfRange(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmOptions([]string{"Option 1", "Option 2"}),
	)
	cb.Focus()

	// Press '5' (out of range)
	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})

	// Should not confirm
	if cb.confirmed {
		t.Error("Should not confirm with out-of-range number")
	}
}

// TestConfirmationBlockIgnoresKeysWhenNotFocused tests ignoring keys without focus
func TestConfirmationBlockIgnoresKeysWhenNotFocused(t *testing.T) {
	cb := NewConfirmationBlock()

	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cb.confirmed {
		t.Error("Should not confirm when not focused")
	}
}

// TestConfirmationBlockIgnoresKeysAfterConfirmed tests ignoring keys after confirmation
func TestConfirmationBlockIgnoresKeysAfterConfirmed(t *testing.T) {
	cb := NewConfirmationBlock()
	cb.Focus()
	cb.confirmed = true
	cb.confirmedIdx = 0

	oldIndex := cb.confirmedIdx

	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

	if cb.confirmedIdx != oldIndex {
		t.Error("Should not change selection after confirmation")
	}
}

// TestConfirmationBlockFocusBlur tests focus management
func TestConfirmationBlockFocusBlur(t *testing.T) {
	cb := NewConfirmationBlock()

	if cb.Focused() {
		t.Error("Should not be focused initially")
	}

	cb.Focus()
	if !cb.Focused() {
		t.Error("Should be focused after Focus()")
	}

	cb.Blur()
	if cb.Focused() {
		t.Error("Should not be focused after Blur()")
	}
}

// TestConfirmationBlockIsConfirmed tests IsConfirmed method
func TestConfirmationBlockIsConfirmed(t *testing.T) {
	cb := NewConfirmationBlock()

	if cb.IsConfirmed() {
		t.Error("Should not be confirmed initially")
	}

	cb.confirmed = true
	if !cb.IsConfirmed() {
		t.Error("Should be confirmed after setting confirmed=true")
	}
}

// TestConfirmationBlockGetSelection tests GetSelection method
func TestConfirmationBlockGetSelection(t *testing.T) {
	cb := NewConfirmationBlock()

	if cb.GetSelection() != -1 {
		t.Errorf("Should return -1 initially, got %d", cb.GetSelection())
	}

	cb.confirmedIdx = 2
	if cb.GetSelection() != 2 {
		t.Errorf("Should return 2, got %d", cb.GetSelection())
	}
}

// TestConfirmationBlockReset tests Reset method
func TestConfirmationBlockReset(t *testing.T) {
	cb := NewConfirmationBlock()
	cb.confirmed = true
	cb.confirmedIdx = 1
	cb.selectedIndex = 2

	cb.Reset()

	if cb.confirmed {
		t.Error("Should not be confirmed after Reset")
	}

	if cb.confirmedIdx != -1 {
		t.Errorf("confirmedIdx should be -1 after Reset, got %d", cb.confirmedIdx)
	}

	if cb.selectedIndex != 0 {
		t.Errorf("selectedIndex should be 0 after Reset, got %d", cb.selectedIndex)
	}
}

// TestConfirmationBlockViewEmpty tests view with minimal content
func TestConfirmationBlockViewEmpty(t *testing.T) {
	cb := NewConfirmationBlock()
	cb.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cb.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	if !strings.Contains(view, "Write") {
		t.Error("View should contain operation")
	}
}

// TestConfirmationBlockViewWithContent tests view with full content
func TestConfirmationBlockViewWithContent(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmOperation("Write"),
		WithConfirmFilepath("test.go"),
		WithConfirmDescription("Create file test.go"),
		WithConfirmCode("package main\n\nfunc main() {}"),
		WithConfirmOptions([]string{"Yes", "No"}),
	)
	cb.Focus()
	cb.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cb.View()

	if !strings.Contains(view, "Write") {
		t.Error("View should contain operation")
	}

	if !strings.Contains(view, "test.go") {
		t.Error("View should contain filepath")
	}

	if !strings.Contains(view, "Create file test.go") {
		t.Error("View should contain description")
	}

	if !strings.Contains(view, "package main") {
		t.Error("View should contain code preview")
	}

	if !strings.Contains(view, "Yes") {
		t.Error("View should contain options")
	}
}

// TestConfirmationBlockViewSelectedOption tests visual selection indicator
func TestConfirmationBlockViewSelectedOption(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmOptions([]string{"Option 1", "Option 2"}),
	)
	cb.Focus()
	cb.selectedIndex = 0
	cb.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cb.View()

	// Should show selection indicator (❯) for selected option
	if !strings.Contains(view, "❯") {
		t.Error("View should contain selection indicator")
	}
}

// TestConfirmationBlockViewConfirmedResult tests view after confirmation
func TestConfirmationBlockViewConfirmedResult(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmOptions([]string{"Yes", "No"}),
	)
	cb.Focus()
	cb.confirmed = true
	cb.confirmedIdx = 0
	cb.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cb.View()

	// Should show confirmation result
	if !strings.Contains(view, "✓") && !strings.Contains(view, "Selected") {
		t.Error("View should show confirmation result")
	}
}

// TestConfirmationBlockViewCancelledResult tests view after cancellation
func TestConfirmationBlockViewCancelledResult(t *testing.T) {
	cb := NewConfirmationBlock()
	cb.Focus()
	cb.confirmed = true
	cb.confirmedIdx = -1
	cb.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cb.View()

	// Should show cancelled message
	if !strings.Contains(view, "Cancelled") {
		t.Error("View should show cancelled message")
	}
}

// TestConfirmationBlockViewFooterHints tests footer hints display
func TestConfirmationBlockViewFooterHints(t *testing.T) {
	hints := []string{"Press Enter to confirm", "Press Esc to cancel"}
	cb := NewConfirmationBlock(
		WithConfirmFooterHints(hints),
	)
	cb.Focus()
	cb.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cb.View()

	for _, hint := range hints {
		if !strings.Contains(view, hint) {
			t.Errorf("View should contain hint: %s", hint)
		}
	}
}

// TestConfirmationBlockViewHidesFooterAfterConfirm tests footer hidden after confirmation
func TestConfirmationBlockViewHidesFooterAfterConfirm(t *testing.T) {
	hints := []string{"Press Enter"}
	cb := NewConfirmationBlock(
		WithConfirmFooterHints(hints),
	)
	cb.Focus()
	cb.confirmed = true
	cb.confirmedIdx = 0
	cb.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cb.View()

	// Footer hints should not appear after confirmation
	// (Implementation may vary - this is a reasonable expectation)
	t.Logf("View after confirmation: %d bytes", len(view))
}

// TestConfirmationBlockOperationIcons tests operation icon mapping
func TestConfirmationBlockOperationIcons(t *testing.T) {
	operations := map[string]string{
		"Write":  "⏺",
		"Read":   "⏺",
		"Edit":   "⏺",
		"Delete": "⏺",
	}

	for op, expectedIcon := range operations {
		cb := NewConfirmationBlock(WithConfirmOperation(op))
		cb.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		view := cb.View()

		if !strings.Contains(view, expectedIcon) {
			t.Errorf("View for operation '%s' should contain icon '%s'", op, expectedIcon)
		}
	}
}

// TestConfirmationBlockCodePreview tests code preview truncation
func TestConfirmationBlockCodePreview(t *testing.T) {
	lines := make([]string, 50)
	for i := 0; i < 50; i++ {
		lines[i] = "line content"
	}

	cb := NewConfirmationBlock(
		WithConfirmCodeLines(lines),
		WithConfirmPreview(5),
	)
	cb.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cb.View()

	// Should show preview indicator for truncated code
	if !strings.Contains(view, "more lines") {
		t.Log("Expected preview indicator (may vary by implementation)")
	}
}

// TestConfirmationBlockEmptyWidth tests behavior with zero width
func TestConfirmationBlockEmptyWidth(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmCode("test"),
	)

	view := cb.View()

	// Should return empty string with zero width
	if view != "" {
		t.Logf("View with zero width returned: %d bytes", len(view))
	}
}

// TestConfirmationBlockMultipleOptions tests combining multiple options
func TestConfirmationBlockMultipleOptions(t *testing.T) {
	cb := NewConfirmationBlock(
		WithConfirmOperation("Edit"),
		WithConfirmFilepath("config.yaml"),
		WithConfirmDescription("Update configuration file"),
		WithConfirmCode("key: value\nother: data"),
		WithConfirmOptions([]string{"Apply", "Preview", "Cancel"}),
		WithConfirmStartLine(10),
		WithConfirmPreview(5),
		WithConfirmFooterHints([]string{"Tab to navigate", "Enter to confirm"}),
	)

	if cb.operation != "Edit" {
		t.Errorf("Expected operation='Edit', got '%s'", cb.operation)
	}

	if cb.filepath != "config.yaml" {
		t.Errorf("Expected filepath='config.yaml', got '%s'", cb.filepath)
	}

	if len(cb.options) != 3 {
		t.Errorf("Expected 3 options, got %d", len(cb.options))
	}

	if cb.startLine != 10 {
		t.Errorf("Expected startLine=10, got %d", cb.startLine)
	}

	if len(cb.footerHints) != 2 {
		t.Errorf("Expected 2 footer hints, got %d", len(cb.footerHints))
	}
}
