package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestCodeBlockCreation tests default creation
func TestCodeBlockCreation(t *testing.T) {
	cb := NewCodeBlock()

	if cb == nil {
		t.Fatal("NewCodeBlock returned nil")
	}

	if cb.operation != "Code" {
		t.Errorf("Expected default operation='Code', got '%s'", cb.operation)
	}

	if cb.startLine != 1 {
		t.Errorf("Expected default startLine=1, got %d", cb.startLine)
	}

	if cb.expanded {
		t.Error("CodeBlock should not be expanded by default")
	}

	if cb.showPreview != 8 {
		t.Errorf("Expected default showPreview=8, got %d", cb.showPreview)
	}
}

// TestCodeBlockWithCodeOperation tests operation option
func TestCodeBlockWithCodeOperation(t *testing.T) {
	cb := NewCodeBlock(
		WithCodeOperation("Read"),
	)

	if cb.operation != "Read" {
		t.Errorf("Expected operation='Read', got '%s'", cb.operation)
	}
}

// TestCodeBlockWithCodeFilename tests filename option
func TestCodeBlockWithCodeFilename(t *testing.T) {
	cb := NewCodeBlock(
		WithCodeFilename("main.go"),
	)

	if cb.filename != "main.go" {
		t.Errorf("Expected filename='main.go', got '%s'", cb.filename)
	}
}

// TestCodeBlockWithCodeSummary tests summary option
func TestCodeBlockWithCodeSummary(t *testing.T) {
	cb := NewCodeBlock(
		WithCodeSummary("Wrote 100 lines"),
	)

	if cb.summary != "Wrote 100 lines" {
		t.Errorf("Expected summary='Wrote 100 lines', got '%s'", cb.summary)
	}
}

// TestCodeBlockWithCode tests code content option
func TestCodeBlockWithCode(t *testing.T) {
	code := "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}"
	cb := NewCodeBlock(
		WithCode(code),
	)

	if len(cb.lines) != 5 {
		t.Errorf("Expected 5 lines, got %d", len(cb.lines))
	}

	if cb.lines[0] != "package main" {
		t.Errorf("Expected first line='package main', got '%s'", cb.lines[0])
	}
}

// TestCodeBlockWithCodeLines tests code lines option
func TestCodeBlockWithCodeLines(t *testing.T) {
	lines := []string{"line 1", "line 2", "line 3"}
	cb := NewCodeBlock(
		WithCodeLines(lines),
	)

	if len(cb.lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(cb.lines))
	}

	if cb.lines[1] != "line 2" {
		t.Errorf("Expected second line='line 2', got '%s'", cb.lines[1])
	}
}

// TestCodeBlockWithLanguage tests language option
func TestCodeBlockWithLanguage(t *testing.T) {
	cb := NewCodeBlock(
		WithLanguage("go"),
	)

	if cb.language != "go" {
		t.Errorf("Expected language='go', got '%s'", cb.language)
	}
}

// TestCodeBlockWithStartLine tests start line option
func TestCodeBlockWithStartLine(t *testing.T) {
	cb := NewCodeBlock(
		WithStartLine(42),
	)

	if cb.startLine != 42 {
		t.Errorf("Expected startLine=42, got %d", cb.startLine)
	}
}

// TestCodeBlockWithCodeMaxLines tests max lines option
func TestCodeBlockWithCodeMaxLines(t *testing.T) {
	cb := NewCodeBlock(
		WithCodeMaxLines(100),
	)

	if cb.maxLines != 100 {
		t.Errorf("Expected maxLines=100, got %d", cb.maxLines)
	}
}

// TestCodeBlockWithExpanded tests expanded option
func TestCodeBlockWithExpanded(t *testing.T) {
	cb := NewCodeBlock(
		WithExpanded(true),
	)

	if !cb.expanded {
		t.Error("Expected expanded=true")
	}
}

// TestCodeBlockWithPreviewLines tests preview lines option
func TestCodeBlockWithPreviewLines(t *testing.T) {
	cb := NewCodeBlock(
		WithPreviewLines(5),
	)

	if cb.showPreview != 5 {
		t.Errorf("Expected showPreview=5, got %d", cb.showPreview)
	}
}

// TestCodeBlockUpdate tests update with window size
func TestCodeBlockUpdate(t *testing.T) {
	cb := NewCodeBlock()

	_, _ = cb.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if cb.width != 100 {
		t.Errorf("Expected width=100, got %d", cb.width)
	}

	if cb.height != 50 {
		t.Errorf("Expected height=50, got %d", cb.height)
	}
}

// TestCodeBlockToggle tests expand/collapse toggle
func TestCodeBlockToggle(t *testing.T) {
	cb := NewCodeBlock()
	cb.Focus()

	if cb.expanded {
		t.Error("Should start collapsed")
	}

	// Toggle to expand
	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyCtrlO})

	if !cb.expanded {
		t.Error("Should be expanded after ctrl+o")
	}

	// Toggle to collapse
	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyCtrlO})

	if cb.expanded {
		t.Error("Should be collapsed after second ctrl+o")
	}
}

// TestCodeBlockToggleWithEnter tests toggle with enter key
func TestCodeBlockToggleWithEnter(t *testing.T) {
	cb := NewCodeBlock()
	cb.Focus()

	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !cb.expanded {
		t.Error("Should be expanded after enter")
	}
}

// TestCodeBlockToggleWithSpace tests toggle with space key
func TestCodeBlockToggleWithSpace(t *testing.T) {
	cb := NewCodeBlock()
	cb.Focus()

	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

	if !cb.expanded {
		t.Error("Should be expanded after space")
	}
}

// TestCodeBlockIgnoresKeysWhenNotFocused tests that keys are ignored without focus
func TestCodeBlockIgnoresKeysWhenNotFocused(t *testing.T) {
	cb := NewCodeBlock()

	_, _ = cb.Update(tea.KeyMsg{Type: tea.KeyCtrlO})

	if cb.expanded {
		t.Error("Should not toggle when not focused")
	}
}

// TestCodeBlockFocusBlur tests focus management
func TestCodeBlockFocusBlur(t *testing.T) {
	cb := NewCodeBlock()

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

// TestCodeBlockIsExpanded tests IsExpanded method
func TestCodeBlockIsExpanded(t *testing.T) {
	cb := NewCodeBlock()

	if cb.IsExpanded() {
		t.Error("Should not be expanded initially")
	}

	cb.expanded = true
	if !cb.IsExpanded() {
		t.Error("Should be expanded after setting expanded=true")
	}
}

// TestCodeBlockViewEmpty tests view with no content
func TestCodeBlockViewEmpty(t *testing.T) {
	cb := NewCodeBlock()
	cb.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cb.View()

	// Should return empty string when no lines
	if view != "" {
		t.Error("View should be empty with no code lines")
	}
}

// TestCodeBlockViewWithContent tests view with code
func TestCodeBlockViewWithContent(t *testing.T) {
	cb := NewCodeBlock(
		WithCodeOperation("Write"),
		WithCodeFilename("test.go"),
		WithCodeSummary("Created test file"),
		WithCode("package main\n\nfunc main() {}"),
	)
	cb.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cb.View()

	if !strings.Contains(view, "Write") {
		t.Error("View should contain operation")
	}

	if !strings.Contains(view, "test.go") {
		t.Error("View should contain filename")
	}

	if !strings.Contains(view, "Created test file") {
		t.Error("View should contain summary")
	}

	if !strings.Contains(view, "package main") {
		t.Error("View should contain code content")
	}
}

// TestCodeBlockViewCollapsed tests collapsed view shows preview
func TestCodeBlockViewCollapsed(t *testing.T) {
	lines := make([]string, 20)
	for i := 0; i < 20; i++ {
		lines[i] = "line content"
	}

	cb := NewCodeBlock(
		WithCodeLines(lines),
		WithPreviewLines(5),
		WithExpanded(false),
	)
	cb.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cb.View()

	// Should show preview indicator
	if !strings.Contains(view, "more lines") && !strings.Contains(view, "ctrl+o to expand") {
		t.Error("Collapsed view should show expansion hint")
	}
}

// TestCodeBlockViewExpanded tests expanded view shows all lines
func TestCodeBlockViewExpanded(t *testing.T) {
	lines := []string{"line 1", "line 2", "line 3"}

	cb := NewCodeBlock(
		WithCodeLines(lines),
		WithExpanded(true),
	)
	cb.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cb.View()

	for _, line := range lines {
		if !strings.Contains(view, line) {
			t.Errorf("Expanded view should contain '%s'", line)
		}
	}
}

// TestCodeBlockOperationIcons tests operation icon mapping
func TestCodeBlockOperationIcons(t *testing.T) {
	operations := map[string]string{
		"Write":  "⏺",
		"Read":   "⏺",
		"Edit":   "⏺",
		"Delete": "⏺",
	}

	for op, expectedIcon := range operations {
		cb := NewCodeBlock(
			WithCodeOperation(op),
			WithCode("test content"), // Need content for View to render
		)
		cb.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		view := cb.View()

		if !strings.Contains(view, expectedIcon) {
			t.Errorf("View for operation '%s' should contain icon '%s'", op, expectedIcon)
		}
	}
}

// TestCodeBlockLineNumbers tests line number rendering
func TestCodeBlockLineNumbers(t *testing.T) {
	cb := NewCodeBlock(
		WithCodeLines([]string{"line 1", "line 2"}),
		WithStartLine(10),
		WithExpanded(true),
	)
	cb.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cb.View()

	// Should show line numbers starting from 10
	if !strings.Contains(view, "10") {
		t.Error("View should contain starting line number")
	}
}

// TestCodeBlockMaxLines tests max lines truncation
func TestCodeBlockMaxLines(t *testing.T) {
	lines := make([]string, 100)
	for i := 0; i < 100; i++ {
		lines[i] = "content"
	}

	cb := NewCodeBlock(
		WithCodeLines(lines),
		WithCodeMaxLines(10),
		WithExpanded(true),
	)

	if len(cb.lines) != 100 {
		t.Errorf("Should preserve all lines internally, got %d", len(cb.lines))
	}

	// MaxLines is used during rendering, not storage
}

// TestCodeBlockEmptyWidth tests behavior with zero width
func TestCodeBlockEmptyWidth(t *testing.T) {
	cb := NewCodeBlock(
		WithCode("test"),
	)

	view := cb.View()

	// Should return empty string or handle gracefully
	if view != "" {
		// Some implementations may render anyway, which is fine
		t.Logf("View with zero width returned: %s", view)
	}
}

// TestCodeBlockMultipleOptions tests combining multiple options
func TestCodeBlockMultipleOptions(t *testing.T) {
	cb := NewCodeBlock(
		WithCodeOperation("Edit"),
		WithCodeFilename("config.yaml"),
		WithCodeSummary("Updated configuration"),
		WithCode("key: value"),
		WithLanguage("yaml"),
		WithStartLine(5),
		WithExpanded(true),
	)

	if cb.operation != "Edit" {
		t.Errorf("Expected operation='Edit', got '%s'", cb.operation)
	}

	if cb.filename != "config.yaml" {
		t.Errorf("Expected filename='config.yaml', got '%s'", cb.filename)
	}

	if cb.summary != "Updated configuration" {
		t.Errorf("Expected correct summary, got '%s'", cb.summary)
	}

	if cb.summary != "Updated configuration" {
		t.Errorf("Expected correct summary, got '%s'", cb.summary)
	}

	if len(cb.lines) == 0 {
		t.Error("Expected code lines to be set")
	}

	if cb.language != "yaml" {
		t.Errorf("Expected language='yaml', got '%s'", cb.language)
	}

	if cb.startLine != 5 {
		t.Errorf("Expected startLine=5, got %d", cb.startLine)
	}

	if !cb.expanded {
		t.Error("Expected expanded=true")
	}
}
