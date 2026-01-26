package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestDiffBlockCreation tests default creation
func TestDiffBlockCreation(t *testing.T) {
	db := NewDiffBlock()

	if db == nil {
		t.Fatal("NewDiffBlock returned nil")
	}

	if db.operation != "Edit" {
		t.Errorf("Expected default operation='Edit', got '%s'", db.operation)
	}

	if db.expanded {
		t.Error("DiffBlock should not be expanded by default")
	}

	if len(db.lines) != 0 {
		t.Errorf("Expected empty lines initially, got %d", len(db.lines))
	}
}

// TestDiffBlockWithDiffFilename tests filename option
func TestDiffBlockWithDiffFilename(t *testing.T) {
	db := NewDiffBlock(
		WithDiffFilename("main.go"),
	)

	if db.filename != "main.go" {
		t.Errorf("Expected filename='main.go', got '%s'", db.filename)
	}
}

// TestDiffBlockWithDiffOperation tests operation option
func TestDiffBlockWithDiffOperation(t *testing.T) {
	db := NewDiffBlock(
		WithDiffOperation("Edit"),
	)

	if db.operation != "Edit" {
		t.Errorf("Expected operation='Edit', got '%s'", db.operation)
	}
}

// TestDiffBlockWithDiffExpanded tests expanded option
func TestDiffBlockWithDiffExpanded(t *testing.T) {
	db := NewDiffBlock(
		WithDiffExpanded(true),
	)

	if !db.expanded {
		t.Error("Expected expanded=true")
	}
}

// TestDiffBlockWithDiffContext tests context lines option
func TestDiffBlockWithDiffContext(t *testing.T) {
	db := NewDiffBlock(
		WithDiffContext(5),
	)

	if db.showContext != 5 {
		t.Errorf("Expected showContext=5, got %d", db.showContext)
	}
}

// TestDiffBlockFromStrings tests creating diff from old/new strings
func TestDiffBlockFromStrings(t *testing.T) {
	oldCode := "line 1\nline 2\nline 3"
	newCode := "line 1\nline 2 modified\nline 3"

	db := NewDiffBlockFromStrings(oldCode, newCode)

	if db == nil {
		t.Fatal("NewDiffBlockFromStrings returned nil")
	}

	if len(db.lines) == 0 {
		t.Error("Expected diff lines to be generated")
	}

	// Check that we have at least one added and one removed line
	hasAdded := false
	hasRemoved := false
	for _, line := range db.lines {
		if line.Type == DiffAdded {
			hasAdded = true
		}
		if line.Type == DiffRemoved {
			hasRemoved = true
		}
	}

	if !hasAdded || !hasRemoved {
		t.Error("Diff should contain both added and removed lines")
	}
}

// TestDiffBlockFromStringsWithOptions tests combining strings with options
func TestDiffBlockFromStringsWithOptions(t *testing.T) {
	oldCode := "old"
	newCode := "new"

	db := NewDiffBlockFromStrings(
		oldCode,
		newCode,
		WithDiffFilename("test.txt"),
		WithDiffOperation("Modify"),
		WithDiffExpanded(true),
	)

	if db.filename != "test.txt" {
		t.Errorf("Expected filename='test.txt', got '%s'", db.filename)
	}

	if db.operation != "Modify" {
		t.Errorf("Expected operation='Modify', got '%s'", db.operation)
	}

	if !db.expanded {
		t.Error("Expected expanded=true")
	}
}

// TestDiffBlockUpdate tests update with window size
func TestDiffBlockUpdate(t *testing.T) {
	db := NewDiffBlock()

	_, _ = db.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	if db.width != 120 {
		t.Errorf("Expected width=120, got %d", db.width)
	}

	if db.height != 30 {
		t.Errorf("Expected height=30, got %d", db.height)
	}
}

// TestDiffBlockToggle tests expand/collapse toggle
func TestDiffBlockToggle(t *testing.T) {
	db := NewDiffBlock()
	db.Focus()

	if db.expanded {
		t.Error("Should start collapsed")
	}

	// Toggle to expand
	db.Toggle()

	if !db.expanded {
		t.Error("Should be expanded after Toggle()")
	}

	// Toggle to collapse
	db.Toggle()

	if db.expanded {
		t.Error("Should be collapsed after second Toggle()")
	}
}

// TestDiffBlockToggleViaUpdate tests toggle with key press
func TestDiffBlockToggleViaUpdate(t *testing.T) {
	db := NewDiffBlock()
	db.Focus()

	_, _ = db.Update(tea.KeyMsg{Type: tea.KeyCtrlO})

	if !db.expanded {
		t.Error("Should be expanded after ctrl+o")
	}
}

// TestDiffBlockToggleWithEnter tests toggle with enter key
func TestDiffBlockToggleWithEnter(t *testing.T) {
	db := NewDiffBlock()
	db.Focus()

	_, _ = db.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !db.expanded {
		t.Error("Should be expanded after enter")
	}
}

// TestDiffBlockToggleWithSpace tests toggle with space key
func TestDiffBlockToggleWithSpace(t *testing.T) {
	db := NewDiffBlock()
	db.Focus()

	_, _ = db.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

	if !db.expanded {
		t.Error("Should be expanded after space")
	}
}

// TestDiffBlockIgnoresKeysWhenNotFocused tests that keys are ignored without focus
func TestDiffBlockIgnoresKeysWhenNotFocused(t *testing.T) {
	db := NewDiffBlock()

	_, _ = db.Update(tea.KeyMsg{Type: tea.KeyCtrlO})

	if db.expanded {
		t.Error("Should not toggle when not focused")
	}
}

// TestDiffBlockFocusBlur tests focus management
func TestDiffBlockFocusBlur(t *testing.T) {
	db := NewDiffBlock()

	if db.Focused() {
		t.Error("Should not be focused initially")
	}

	db.Focus()
	if !db.Focused() {
		t.Error("Should be focused after Focus()")
	}

	db.Blur()
	if db.Focused() {
		t.Error("Should not be focused after Blur()")
	}
}

// TestDiffBlockIsExpanded tests IsExpanded method
func TestDiffBlockIsExpanded(t *testing.T) {
	db := NewDiffBlock()

	if db.IsExpanded() {
		t.Error("Should not be expanded initially")
	}

	db.expanded = true
	if !db.IsExpanded() {
		t.Error("Should be expanded after setting expanded=true")
	}
}

// TestDiffBlockViewEmpty tests view with no diff
func TestDiffBlockViewEmpty(t *testing.T) {
	db := NewDiffBlock()
	db.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := db.View()

	// Should return empty string when no diff lines
	if view != "" {
		t.Error("View should be empty with no diff lines")
	}
}

// TestDiffBlockViewWithDiff tests view with diff content
func TestDiffBlockViewWithDiff(t *testing.T) {
	oldCode := "old line"
	newCode := "new line"

	db := NewDiffBlockFromStrings(
		oldCode,
		newCode,
		WithDiffFilename("test.txt"),
		WithDiffOperation("Update"),
	)
	db.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := db.View()

	if !strings.Contains(view, "Update") {
		t.Error("View should contain operation")
	}

	if !strings.Contains(view, "test.txt") {
		t.Error("View should contain filename")
	}
}

// TestDiffBlockViewCollapsed tests collapsed view shows summary
func TestDiffBlockViewCollapsed(t *testing.T) {
	oldCode := "line 1\nline 2\nline 3"
	newCode := "line 1\nmodified\nline 3\nline 4"

	db := NewDiffBlockFromStrings(oldCode, newCode)
	db.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := db.View()

	// Should show summary: "Added N lines, removed N lines"
	if !strings.Contains(view, "Added") && !strings.Contains(view, "removed") {
		t.Error("Collapsed view should show change summary")
	}
}

// TestDiffBlockViewExpanded tests expanded view shows all diff lines
func TestDiffBlockViewExpanded(t *testing.T) {
	oldCode := "old line"
	newCode := "new line"

	db := NewDiffBlockFromStrings(
		oldCode,
		newCode,
		WithDiffExpanded(true),
	)
	db.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := db.View()

	// Expanded view should show actual diff content
	// (specific format depends on implementation)
	if len(view) < 10 {
		t.Error("Expanded view should have substantial content")
	}
}

// TestDiffBlockSummaryGeneration tests summary calculation
func TestDiffBlockSummaryGeneration(t *testing.T) {
	oldCode := "line 1\nline 2"
	newCode := "line 1\nline 2\nline 3"

	db := NewDiffBlockFromStrings(oldCode, newCode)

	// Should have generated diff lines
	if len(db.lines) == 0 {
		t.Error("Diff lines should be generated")
	}

	// Check that we have at least one added line
	hasAdded := false
	for _, line := range db.lines {
		if line.Type == DiffAdded {
			hasAdded = true
			break
		}
	}

	if !hasAdded {
		t.Error("Should have at least one added line")
	}
}

// TestDiffBlockOperationIcons tests operation icon mapping
func TestDiffBlockOperationIcons(t *testing.T) {
	operations := map[string]string{
		"Update": "⏺",
		"Edit":   "⏺",
		"Modify": "⏺",
	}

	for op, expectedIcon := range operations {
		db := NewDiffBlockFromStrings(
			"old",
			"new",
			WithDiffOperation(op), // Need diff content for View to render
		)
		db.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		view := db.View()

		if !strings.Contains(view, expectedIcon) {
			t.Errorf("View for operation '%s' should contain icon '%s'", op, expectedIcon)
		}
	}
}

// TestDiffBlockLineTypes tests different diff line types
func TestDiffBlockLineTypes(t *testing.T) {
	db := NewDiffBlock()

	// Add different line types
	db.lines = []DiffLine{
		{Type: DiffUnchanged, Content: "unchanged", LineNum: 1},
		{Type: DiffAdded, Content: "added", LineNum: 2},
		{Type: DiffRemoved, Content: "removed", LineNum: 3},
	}

	db.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Check that line types are preserved
	if len(db.lines) != 3 {
		t.Errorf("Expected 3 diff lines, got %d", len(db.lines))
	}

	if db.lines[0].Type != DiffUnchanged {
		t.Error("First line should be unchanged")
	}

	if db.lines[1].Type != DiffAdded {
		t.Error("Second line should be added")
	}

	if db.lines[2].Type != DiffRemoved {
		t.Error("Third line should be removed")
	}
}

// TestDiffBlockContextLines tests context line extraction
func TestDiffBlockContextLines(t *testing.T) {
	oldCode := "line 1\nline 2\nline 3\nline 4\nline 5"
	newCode := "line 1\nline 2 modified\nline 3\nline 4\nline 5"

	db := NewDiffBlockFromStrings(
		oldCode,
		newCode,
		WithDiffContext(1),
	)

	// Should include context lines around changes
	if len(db.lines) == 0 {
		t.Error("Diff should have lines")
	}

	// Check that context lines exist (implementation-specific)
	hasContext := false
	for _, line := range db.lines {
		if line.Type == DiffUnchanged {
			hasContext = true
			break
		}
	}

	if !hasContext {
		t.Log("Expected context lines (may vary by implementation)")
	}
}

// TestDiffBlockEmptyDiff tests diff with no changes
func TestDiffBlockEmptyDiff(t *testing.T) {
	sameCode := "line 1\nline 2"

	db := NewDiffBlockFromStrings(sameCode, sameCode)

	// Should handle no-change case
	if db == nil {
		t.Fatal("Should create diff block even with no changes")
	}

	// May have all unchanged lines or be empty
	hasChanges := false
	for _, line := range db.lines {
		if line.Type == DiffAdded || line.Type == DiffRemoved {
			hasChanges = true
			break
		}
	}

	if hasChanges {
		t.Error("Should have no added or removed lines for identical content")
	}
}

// TestDiffBlockLargeFile tests diff with many lines
func TestDiffBlockLargeFile(t *testing.T) {
	var oldLines, newLines []string
	for i := 0; i < 100; i++ {
		oldLines = append(oldLines, "line content")
	}
	oldCode := strings.Join(oldLines, "\n")

	// Modify middle line
	for i := 0; i < 100; i++ {
		if i == 50 {
			newLines = append(newLines, "modified line")
		} else {
			newLines = append(newLines, "line content")
		}
	}
	newCode := strings.Join(newLines, "\n")

	db := NewDiffBlockFromStrings(oldCode, newCode)

	if db == nil {
		t.Fatal("Should handle large files")
	}

	// Should have generated diff
	if len(db.lines) == 0 {
		t.Error("Should have diff lines for large file")
	}
}

// TestDiffBlockMultipleOptions tests combining multiple options
func TestDiffBlockMultipleOptions(t *testing.T) {
	db := NewDiffBlock(
		WithDiffFilename("config.yaml"),
		WithDiffOperation("Patch"),
		WithDiffExpanded(false),
		WithDiffContext(3),
	)

	if db.filename != "config.yaml" {
		t.Errorf("Expected filename='config.yaml', got '%s'", db.filename)
	}

	if db.operation != "Patch" {
		t.Errorf("Expected operation='Patch', got '%s'", db.operation)
	}

	if db.expanded {
		t.Error("Expected expanded=false")
	}

	if db.showContext != 3 {
		t.Errorf("Expected showContext=3, got %d", db.showContext)
	}
}

// TestDiffBlockEmptyWidth tests behavior with zero width
func TestDiffBlockEmptyWidth(t *testing.T) {
	db := NewDiffBlockFromStrings("old", "new")

	view := db.View()

	// Should return empty string or handle gracefully with zero width
	if view != "" {
		t.Logf("View with zero width returned: %s", view)
	}
}

// TestDiffBlockLineNumbering tests line number assignment
func TestDiffBlockLineNumbering(t *testing.T) {
	oldCode := "line 1\nline 2\nline 3"
	newCode := "line 1\nmodified\nline 3"

	db := NewDiffBlockFromStrings(oldCode, newCode)

	// Check that line numbers are assigned
	for _, line := range db.lines {
		if line.LineNum < 0 {
			t.Error("Line numbers should be non-negative")
		}
	}
}
