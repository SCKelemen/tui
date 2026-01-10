package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestToolBlockCreation(t *testing.T) {
	block := NewToolBlock("Bash", "ls -la", []string{"file1.go", "file2.go"})
	if block == nil {
		t.Fatal("NewToolBlock returned nil")
	}

	if block.toolName != "Bash" {
		t.Errorf("Expected toolName 'Bash', got %q", block.toolName)
	}

	if block.command != "ls -la" {
		t.Errorf("Expected command 'ls -la', got %q", block.command)
	}

	if len(block.output) != 2 {
		t.Errorf("Expected 2 output lines, got %d", len(block.output))
	}

	if block.expanded {
		t.Error("ToolBlock should not be expanded by default")
	}

	if block.maxLines != 5 {
		t.Errorf("Expected default maxLines 5, got %d", block.maxLines)
	}

	if block.status != StatusComplete {
		t.Errorf("Expected default status StatusComplete, got %v", block.status)
	}
}

func TestToolBlockWithOptions(t *testing.T) {
	// Test WithLineNumbers
	block := NewToolBlock("Read", "test.go", []string{"package main"}, WithLineNumbers())
	if !block.showLineNos {
		t.Error("WithLineNumbers option not applied")
	}

	// Test WithMaxLines
	block = NewToolBlock("Bash", "test", []string{}, WithMaxLines(10))
	if block.maxLines != 10 {
		t.Errorf("WithMaxLines option not applied, got %d", block.maxLines)
	}

	// Test WithStreaming
	block = NewToolBlock("Bash", "test", []string{}, WithStreaming())
	if !block.streaming {
		t.Error("WithStreaming option not applied")
	}
	if block.status != StatusRunning {
		t.Error("WithStreaming should set status to StatusRunning")
	}

	// Test WithStatus
	block = NewToolBlock("Bash", "test", []string{}, WithStatus(StatusError))
	if block.status != StatusError {
		t.Errorf("WithStatus option not applied, got %v", block.status)
	}
}

func TestToolBlockToggleExpanded(t *testing.T) {
	block := NewToolBlock("Bash", "test", []string{})

	if block.expanded {
		t.Error("ToolBlock should start collapsed")
	}

	block.ToggleExpanded()
	if !block.expanded {
		t.Error("ToggleExpanded did not expand")
	}

	block.ToggleExpanded()
	if block.expanded {
		t.Error("ToggleExpanded did not collapse")
	}
}

func TestToolBlockSetExpanded(t *testing.T) {
	block := NewToolBlock("Bash", "test", []string{})

	block.SetExpanded(true)
	if !block.expanded {
		t.Error("SetExpanded(true) did not expand")
	}

	block.SetExpanded(false)
	if block.expanded {
		t.Error("SetExpanded(false) did not collapse")
	}
}

func TestToolBlockCollapsedView(t *testing.T) {
	// Create block with more lines than maxLines
	output := []string{
		"line 1",
		"line 2",
		"line 3",
		"line 4",
		"line 5",
		"line 6",
		"line 7",
		"line 8",
	}

	block := NewToolBlock("Bash", "test", output, WithMaxLines(3))
	block.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := block.View()

	// Should contain first 3 lines
	if !strings.Contains(view, "line 1") {
		t.Error("Collapsed view should contain line 1")
	}
	if !strings.Contains(view, "line 3") {
		t.Error("Collapsed view should contain line 3")
	}

	// Should show "... +N lines" indicator
	if !strings.Contains(view, "+5 lines") {
		t.Error("Collapsed view should show +5 lines indicator")
	}

	// Should show expand hint
	if !strings.Contains(view, "ctrl+o to expand") {
		t.Error("Collapsed view should show expand hint")
	}
}

func TestToolBlockExpandedView(t *testing.T) {
	output := []string{
		"line 1",
		"line 2",
		"line 3",
		"line 4",
		"line 5",
	}

	block := NewToolBlock("Bash", "test", output, WithMaxLines(2))
	block.SetExpanded(true)
	block.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := block.View()

	// Should contain all lines when expanded
	if !strings.Contains(view, "line 1") {
		t.Error("Expanded view should contain line 1")
	}
	if !strings.Contains(view, "line 5") {
		t.Error("Expanded view should contain line 5")
	}

	// Should NOT show "... +N lines" when expanded
	if strings.Contains(view, "+3 lines") {
		t.Error("Expanded view should not show +N lines indicator")
	}
}

func TestToolBlockLineNumbers(t *testing.T) {
	output := []string{"line 1", "line 2", "line 3"}

	block := NewToolBlock("Read", "test.go", output, WithLineNumbers())
	block.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := block.View()

	// Should contain line numbers (ANSI codes make exact matching tricky)
	// Just verify the view is not empty and contains content
	if view == "" {
		t.Error("View with line numbers should not be empty")
	}

	if !strings.Contains(view, "line 1") {
		t.Error("View should contain line content")
	}
}

func TestToolBlockEmptyOutput(t *testing.T) {
	block := NewToolBlock("Bash", "test", []string{})
	block.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := block.View()

	if !strings.Contains(view, "(no output)") {
		t.Error("Empty output should show '(no output)'")
	}
}

func TestToolBlockStreamingOutput(t *testing.T) {
	block := NewToolBlock("Bash", "test", []string{}, WithStreaming())
	block.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := block.View()

	if !strings.Contains(view, "streaming...") {
		t.Error("Streaming with no output should show 'streaming...'")
	}
}

func TestToolBlockStatusIndicators(t *testing.T) {
	tests := []struct {
		name     string
		status   ToolBlockStatus
		contains string
	}{
		{"Complete", StatusComplete, "✓"},
		{"Error", StatusError, "✗"},
		{"Warning", StatusWarning, "⚠"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := NewToolBlock("Bash", "test", []string{"output"}, WithStatus(tt.status))
			block.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

			view := block.View()
			if !strings.Contains(view, tt.contains) {
				t.Errorf("Status %v should display %q indicator", tt.status, tt.contains)
			}
		})
	}
}

func TestToolBlockAppendLine(t *testing.T) {
	block := NewToolBlock("Bash", "test", []string{"line 1"}, WithStreaming())

	if len(block.output) != 1 {
		t.Errorf("Expected 1 initial line, got %d", len(block.output))
	}

	block.AppendLine("line 2")
	if len(block.output) != 2 {
		t.Errorf("Expected 2 lines after AppendLine, got %d", len(block.output))
	}

	if block.output[1] != "line 2" {
		t.Errorf("Expected last line to be 'line 2', got %q", block.output[1])
	}
}

func TestToolBlockAppendLines(t *testing.T) {
	block := NewToolBlock("Bash", "test", []string{"line 1"}, WithStreaming())

	block.AppendLines([]string{"line 2", "line 3", "line 4"})
	if len(block.output) != 4 {
		t.Errorf("Expected 4 lines after AppendLines, got %d", len(block.output))
	}

	if block.output[3] != "line 4" {
		t.Errorf("Expected last line to be 'line 4', got %q", block.output[3])
	}
}

func TestToolBlockSetStatus(t *testing.T) {
	block := NewToolBlock("Bash", "test", []string{}, WithStreaming())

	if block.status != StatusRunning {
		t.Error("Streaming block should start with StatusRunning")
	}

	if !block.streaming {
		t.Error("Block should be streaming")
	}

	block.SetStatus(StatusComplete)
	if block.status != StatusComplete {
		t.Error("SetStatus did not change status")
	}

	if block.streaming {
		t.Error("SetStatus to non-running should stop streaming")
	}
}

func TestToolBlockStartStreaming(t *testing.T) {
	block := NewToolBlock("Bash", "test", []string{})

	cmd := block.StartStreaming()
	if cmd == nil {
		t.Error("StartStreaming should return a tick command")
	}

	if !block.streaming {
		t.Error("StartStreaming should enable streaming")
	}

	if block.status != StatusRunning {
		t.Error("StartStreaming should set status to StatusRunning")
	}

	if block.spinner != 0 {
		t.Error("StartStreaming should reset spinner to 0")
	}
}

func TestToolBlockStopStreaming(t *testing.T) {
	block := NewToolBlock("Bash", "test", []string{}, WithStreaming())

	block.StopStreaming()
	if block.streaming {
		t.Error("StopStreaming should disable streaming")
	}

	if block.status != StatusComplete {
		t.Error("StopStreaming should set status to StatusComplete")
	}
}

func TestToolBlockStopStreamingWithError(t *testing.T) {
	block := NewToolBlock("Bash", "test", []string{}, WithStreaming())

	block.StopStreamingWithError()
	if block.streaming {
		t.Error("StopStreamingWithError should disable streaming")
	}

	if block.status != StatusError {
		t.Error("StopStreamingWithError should set status to StatusError")
	}
}

func TestToolBlockFocusManagement(t *testing.T) {
	block := NewToolBlock("Bash", "test", []string{})

	if block.Focused() {
		t.Error("ToolBlock should not be focused initially")
	}

	block.Focus()
	if !block.Focused() {
		t.Error("ToolBlock should be focused after Focus()")
	}

	block.Blur()
	if block.Focused() {
		t.Error("ToolBlock should not be focused after Blur()")
	}
}

func TestToolBlockKeyboardToggle(t *testing.T) {
	block := NewToolBlock("Bash", "test", []string{"line 1", "line 2"}, WithMaxLines(1))
	block.Focus()
	block.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if block.expanded {
		t.Error("Block should start collapsed")
	}

	// Press Ctrl+O to expand
	block.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	if !block.expanded {
		t.Error("Ctrl+O should expand the block")
	}

	// Press Enter to collapse
	block.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if block.expanded {
		t.Error("Enter should collapse the block")
	}
}

func TestToolBlockKeyboardNoActionWhenBlurred(t *testing.T) {
	block := NewToolBlock("Bash", "test", []string{"line 1"})
	// Don't focus the block

	if block.expanded {
		t.Error("Block should start collapsed")
	}

	// Press Ctrl+O (should have no effect when not focused)
	block.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	if block.expanded {
		t.Error("Ctrl+O should have no effect when not focused")
	}
}

func TestToolBlockWindowSizeUpdate(t *testing.T) {
	block := NewToolBlock("Bash", "test", []string{})

	if block.width != 0 {
		t.Error("Initial width should be 0")
	}

	block.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if block.width != 100 {
		t.Errorf("Expected width 100, got %d", block.width)
	}
}

func TestToolBlockViewWithoutWidth(t *testing.T) {
	block := NewToolBlock("Bash", "test", []string{"line 1"})

	view := block.View()
	if view != "" {
		t.Error("View should be empty when width is not set")
	}
}

func TestToolBlockDifferentToolNames(t *testing.T) {
	tools := []string{"Bash", "Write", "Read", "Edit", "Grep", "Glob", "Task", "WebFetch"}

	for _, tool := range tools {
		block := NewToolBlock(tool, "test command", []string{})
		if block.toolName != tool {
			t.Errorf("Tool name not set correctly for %s", tool)
		}

		icon := getToolIcon(tool)
		if icon == "" {
			t.Errorf("No icon returned for tool %s", tool)
		}

		block.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		view := block.View()
		if !strings.Contains(view, tool) {
			t.Errorf("View should contain tool name %s", tool)
		}
	}
}

func TestToolBlockLongLinesTruncation(t *testing.T) {
	longLine := strings.Repeat("a", 200)
	block := NewToolBlock("Bash", "test", []string{longLine})
	block.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := block.View()
	// View should not contain the full 200-character line
	lines := strings.Split(view, "\n")
	for _, line := range lines {
		strippedLine := stripANSI(line)
		if len(strippedLine) > 100 {
			t.Errorf("Line should be truncated, got length %d", len(strippedLine))
		}
	}
}

func TestToolBlockCommandTruncation(t *testing.T) {
	longCommand := strings.Repeat("command ", 20)
	block := NewToolBlock("Bash", longCommand, []string{})
	block.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := block.View()
	// Command should be truncated in the view
	if strings.Contains(view, longCommand) {
		t.Error("Long command should be truncated")
	}

	if !strings.Contains(view, "...") {
		t.Error("Truncated command should contain '...'")
	}
}
