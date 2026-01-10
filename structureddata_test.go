package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStructuredDataCreation(t *testing.T) {
	sd := NewStructuredData("Test")
	if sd == nil {
		t.Fatal("NewStructuredData returned nil")
	}

	if sd.title != "Test" {
		t.Errorf("Expected title 'Test', got %q", sd.title)
	}

	if !sd.expanded {
		t.Error("StructuredData should be expanded by default")
	}

	if len(sd.items) != 0 {
		t.Errorf("Expected 0 items initially, got %d", len(sd.items))
	}
}

func TestStructuredDataAddRow(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.AddRow("Key1", "Value1")

	if len(sd.items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(sd.items))
	}

	if sd.items[0].Type != ItemKeyValue {
		t.Error("Item should be KeyValue type")
	}

	if sd.items[0].Key != "Key1" {
		t.Errorf("Expected key 'Key1', got %q", sd.items[0].Key)
	}

	if sd.items[0].Value != "Value1" {
		t.Errorf("Expected value 'Value1', got %q", sd.items[0].Value)
	}
}

func TestStructuredDataBuilderPattern(t *testing.T) {
	sd := NewStructuredData("Test").
		AddRow("Key1", "Value1").
		AddRow("Key2", "Value2").
		AddHeader("Section").
		AddIndentedRow("Key3", "Value3", 1)

	if len(sd.items) != 4 {
		t.Errorf("Expected 4 items, got %d", len(sd.items))
	}

	if sd.items[2].Type != ItemHeader {
		t.Error("Third item should be Header type")
	}

	if sd.items[3].Indent != 1 {
		t.Errorf("Fourth item should have indent 1, got %d", sd.items[3].Indent)
	}
}

func TestStructuredDataAddHeader(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.AddHeader("Section 1")

	if len(sd.items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(sd.items))
	}

	if sd.items[0].Type != ItemHeader {
		t.Error("Item should be Header type")
	}

	if sd.items[0].Value != "Section 1" {
		t.Errorf("Expected value 'Section 1', got %q", sd.items[0].Value)
	}
}

func TestStructuredDataAddSeparator(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.AddSeparator()

	if len(sd.items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(sd.items))
	}

	if sd.items[0].Type != ItemSeparator {
		t.Error("Item should be Separator type")
	}
}

func TestStructuredDataAddValue(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.AddValue("Plain text value")

	if len(sd.items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(sd.items))
	}

	if sd.items[0].Type != ItemValue {
		t.Error("Item should be Value type")
	}

	if sd.items[0].Value != "Plain text value" {
		t.Errorf("Expected value 'Plain text value', got %q", sd.items[0].Value)
	}
}

func TestStructuredDataIndentation(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.AddIndentedRow("Key", "Value", 2)

	if sd.items[0].Indent != 2 {
		t.Errorf("Expected indent 2, got %d", sd.items[0].Indent)
	}
}

func TestStructuredDataColoredRow(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.AddColoredRow("Key", "Value", "\033[32m")

	if sd.items[0].Color != "\033[32m" {
		t.Errorf("Expected color '\\033[32m', got %q", sd.items[0].Color)
	}
}

func TestStructuredDataView(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.AddRow("Key", "Value")
	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := sd.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	if !strings.Contains(view, "Test") {
		t.Error("View should contain title")
	}

	if !strings.Contains(view, "Key") {
		t.Error("View should contain key")
	}

	if !strings.Contains(view, "Value") {
		t.Error("View should contain value")
	}

	if !strings.Contains(view, "⎿") {
		t.Error("View should contain tree connector")
	}
}

func TestStructuredDataViewEmpty(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := sd.View()
	if !strings.Contains(view, "(no data)") {
		t.Error("Empty view should show '(no data)'")
	}
}

func TestStructuredDataViewWithoutWidth(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.AddRow("Key", "Value")

	view := sd.View()
	if view != "" {
		t.Error("View should be empty when width is not set")
	}
}

func TestStructuredDataCollapsed(t *testing.T) {
	sd := NewStructuredData("Test", WithStructuredDataMaxLines(2))
	sd.AddRow("Item 1", "Value 1")
	sd.AddRow("Item 2", "Value 2")
	sd.AddRow("Item 3", "Value 3")
	sd.AddRow("Item 4", "Value 4")

	sd.SetExpanded(false)
	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := sd.View()

	if !strings.Contains(view, "Item 1") {
		t.Error("Collapsed view should contain Item 1")
	}

	if !strings.Contains(view, "Item 2") {
		t.Error("Collapsed view should contain Item 2")
	}

	if !strings.Contains(view, "+2 items") {
		t.Error("Collapsed view should show '+2 items' indicator")
	}

	if !strings.Contains(view, "ctrl+o to expand") {
		t.Error("Collapsed view should show expand hint")
	}
}

func TestStructuredDataExpanded(t *testing.T) {
	sd := NewStructuredData("Test", WithStructuredDataMaxLines(2))
	sd.AddRow("Item 1", "Value 1")
	sd.AddRow("Item 2", "Value 2")
	sd.AddRow("Item 3", "Value 3")

	sd.SetExpanded(true)
	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := sd.View()

	if !strings.Contains(view, "Item 3") {
		t.Error("Expanded view should contain all items")
	}

	if strings.Contains(view, "+1 items") {
		t.Error("Expanded view should not show items indicator")
	}
}

func TestStructuredDataToggleExpanded(t *testing.T) {
	sd := NewStructuredData("Test")

	if !sd.expanded {
		t.Error("Should start expanded")
	}

	sd.ToggleExpanded()
	if sd.expanded {
		t.Error("Should be collapsed after toggle")
	}

	sd.ToggleExpanded()
	if !sd.expanded {
		t.Error("Should be expanded after second toggle")
	}
}

func TestStructuredDataFocusManagement(t *testing.T) {
	sd := NewStructuredData("Test")

	if sd.Focused() {
		t.Error("Should not be focused initially")
	}

	sd.Focus()
	if !sd.Focused() {
		t.Error("Should be focused after Focus()")
	}

	sd.Blur()
	if sd.Focused() {
		t.Error("Should not be focused after Blur()")
	}
}

func TestStructuredDataKeyboardToggle(t *testing.T) {
	sd := NewStructuredData("Test", WithStructuredDataMaxLines(2))
	sd.Focus()
	sd.AddRow("Item 1", "Value 1")
	sd.AddRow("Item 2", "Value 2")
	sd.AddRow("Item 3", "Value 3")

	if !sd.expanded {
		t.Error("Should start expanded")
	}

	// Press Ctrl+O to collapse
	sd.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	if sd.expanded {
		t.Error("Ctrl+O should collapse")
	}

	// Press Enter to expand
	sd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !sd.expanded {
		t.Error("Enter should expand")
	}
}

func TestStructuredDataClear(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.AddRow("Key1", "Value1")
	sd.AddRow("Key2", "Value2")

	if len(sd.items) != 2 {
		t.Error("Should have 2 items before clear")
	}

	sd.Clear()

	if len(sd.items) != 0 {
		t.Errorf("Should have 0 items after clear, got %d", len(sd.items))
	}
}

func TestStructuredDataSetItems(t *testing.T) {
	sd := NewStructuredData("Test")

	items := []DataItem{
		{Type: ItemKeyValue, Key: "Key1", Value: "Value1"},
		{Type: ItemKeyValue, Key: "Key2", Value: "Value2"},
	}

	sd.SetItems(items)

	if len(sd.items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(sd.items))
	}
}

func TestStructuredDataFromMap(t *testing.T) {
	data := map[string]string{
		"Key1": "Value1",
		"Key2": "Value2",
	}

	sd := FromMap("Test", data)

	if len(sd.items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(sd.items))
	}

	if sd.title != "Test" {
		t.Errorf("Expected title 'Test', got %q", sd.title)
	}
}

func TestStructuredDataFromKeyValuePairs(t *testing.T) {
	sd := FromKeyValuePairs("Test", "Key1", "Value1", "Key2", "Value2")

	if len(sd.items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(sd.items))
	}

	if sd.items[0].Key != "Key1" {
		t.Errorf("Expected key 'Key1', got %q", sd.items[0].Key)
	}

	if sd.items[1].Value != "Value2" {
		t.Errorf("Expected value 'Value2', got %q", sd.items[1].Value)
	}
}

func TestStructuredDataKeyWidthCalculation(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.AddRow("Short", "Value")
	sd.AddRow("Very Long Key Name", "Value")

	width := sd.calculateKeyWidth()

	if width < len("Very Long Key Name") {
		t.Errorf("Key width should accommodate longest key, got %d", width)
	}

	if width > 40 {
		t.Errorf("Key width should be capped at 40, got %d", width)
	}
}

func TestStructuredDataCustomKeyWidth(t *testing.T) {
	sd := NewStructuredData("Test", WithKeyWidth(30))

	if sd.keyWidth != 30 {
		t.Errorf("Expected key width 30, got %d", sd.keyWidth)
	}
}

func TestStructuredDataCustomIcon(t *testing.T) {
	sd := NewStructuredData("Test", WithStructuredDataIcon("📊"))

	if sd.icon != "📊" {
		t.Errorf("Expected icon '📊', got %q", sd.icon)
	}
}

func TestStructuredDataWindowSizeUpdate(t *testing.T) {
	sd := NewStructuredData("Test")

	if sd.width != 0 {
		t.Error("Initial width should be 0")
	}

	sd.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if sd.width != 100 {
		t.Errorf("Expected width 100, got %d", sd.width)
	}
}

func TestStructuredDataMultipleSections(t *testing.T) {
	sd := NewStructuredData("Test").
		AddHeader("Section 1").
		AddRow("Key1", "Value1").
		AddSeparator().
		AddHeader("Section 2").
		AddRow("Key2", "Value2")

	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := sd.View()

	if !strings.Contains(view, "Section 1") {
		t.Error("View should contain Section 1")
	}

	if !strings.Contains(view, "Section 2") {
		t.Error("View should contain Section 2")
	}
}

func TestStructuredDataNestedIndentation(t *testing.T) {
	sd := NewStructuredData("Test").
		AddRow("Level 0", "Value").
		AddIndentedRow("Level 1", "Value", 1).
		AddIndentedRow("Level 2", "Value", 2).
		AddIndentedRow("Level 3", "Value", 3)

	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := sd.View()

	if view == "" {
		t.Error("View should not be empty with nested indentation")
	}
}

func TestStructuredDataUnicodeContent(t *testing.T) {
	sd := NewStructuredData("Test").
		AddRow("日本語", "テスト").
		AddRow("Emoji", "🎉 ✨ 🚀")

	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := sd.View()

	if !strings.Contains(view, "日本語") {
		t.Error("View should contain unicode content")
	}

	if !strings.Contains(view, "🎉") {
		t.Error("View should contain emoji")
	}
}

func TestStructuredDataInit(t *testing.T) {
	sd := NewStructuredData("Test")
	cmd := sd.Init()

	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestStructuredDataDataStatusRunning(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	cmd := sd.StartRunning()
	if cmd == nil {
		t.Error("StartRunning should return tick command")
	}

	if sd.GetStatus() != DataStatusRunning {
		t.Error("Status should be Running")
	}

	view1 := sd.View()
	if view1 == "" {
		t.Error("View should not be empty")
	}

	// Simulate tick to advance animation
	sd.Update(structuredDataTickMsg{})
	view2 := sd.View()

	// Views should differ due to blinking animation
	if view1 == view2 {
		t.Error("Views should differ due to animation")
	}
}

func TestStructuredDataDataStatusSuccess(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	sd.AddRow("Key", "Value")

	sd.MarkSuccess()

	if sd.GetStatus() != DataStatusSuccess {
		t.Error("Status should be Success")
	}

	view := sd.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// Green ANSI code should be present
	if !strings.Contains(view, "\033[32m") {
		t.Error("View should contain green color code for success")
	}
}

func TestStructuredDataDataStatusError(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	sd.AddRow("Key", "Value")

	sd.MarkError()

	if sd.GetStatus() != DataStatusError {
		t.Error("Status should be Error")
	}

	view := sd.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// Red ANSI code should be present
	if !strings.Contains(view, "\033[31m") {
		t.Error("View should contain red color code for error")
	}
}

func TestStructuredDataDataStatusWarning(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	sd.AddRow("Key", "Value")

	sd.MarkWarning()

	if sd.GetStatus() != DataStatusWarning {
		t.Error("Status should be Warning")
	}

	view := sd.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// Yellow ANSI code should be present
	if !strings.Contains(view, "\033[33m") {
		t.Error("View should contain yellow color code for warning")
	}
}

func TestStructuredDataDataStatusInfo(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	sd.AddRow("Key", "Value")

	sd.MarkInfo()

	if sd.GetStatus() != DataStatusInfo {
		t.Error("Status should be Info")
	}

	view := sd.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// White ANSI code should be present
	if !strings.Contains(view, "\033[37m") {
		t.Error("View should contain white color code for info")
	}
}

func TestStructuredDataSetStatus(t *testing.T) {
	sd := NewStructuredData("Test")

	// Test setting to running
	cmd := sd.SetStatus(DataStatusRunning)
	if cmd == nil {
		t.Error("SetStatus(Running) should return tick command")
	}
	if sd.GetStatus() != DataStatusRunning {
		t.Error("Status should be Running")
	}

	// Test setting to success (no command)
	cmd = sd.SetStatus(DataStatusSuccess)
	if cmd != nil {
		t.Error("SetStatus(Success) should return nil")
	}
	if sd.GetStatus() != DataStatusSuccess {
		t.Error("Status should be Success")
	}
}

func TestStructuredDataAnimationFrameAdvances(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	sd.StartRunning()
	initialFrame := sd.animationFrame

	// Simulate several ticks
	for i := 0; i < 5; i++ {
		sd.Update(structuredDataTickMsg{})
	}

	if sd.animationFrame <= initialFrame {
		t.Error("Animation frame should advance on tick")
	}
}

func TestStructuredDataInitWithRunningStatus(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.StartRunning()

	cmd := sd.Init()
	if cmd == nil {
		t.Error("Init should return tick command when status is Running")
	}
}

func TestStructuredDataStatusTransitions(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Start running
	sd.StartRunning()
	if sd.GetStatus() != DataStatusRunning {
		t.Error("Should be running")
	}

	// Mark success
	sd.MarkSuccess()
	if sd.GetStatus() != DataStatusSuccess {
		t.Error("Should be success")
	}

	// Tick should not continue animation after success
	sd.Update(structuredDataTickMsg{})
	// No error expected, just verify it doesn't panic
}

func TestStructuredDataBlinkingAnimation(t *testing.T) {
	sd := NewStructuredData("Test")
	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	sd.StartRunning()

	// Get views at different animation frames
	views := make([]string, 4)
	for i := 0; i < 4; i++ {
		views[i] = sd.View()
		sd.Update(structuredDataTickMsg{})
	}

	// Should see alternating patterns (blink on/off)
	if views[0] == views[1] && views[1] == views[2] {
		t.Error("Views should differ due to blinking animation")
	}
}

func TestStructuredDataCustomRunningColor(t *testing.T) {
	// Test with custom cyan color
	sd := NewStructuredData("Test", WithRunningColor("\033[36m"))
	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	sd.StartRunning()

	view := sd.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// Cyan ANSI code should be present (when icon is visible)
	if !strings.Contains(view, "\033[36m") {
		t.Error("View should contain custom cyan color code for running")
	}
}

func TestStructuredDataDefaultRunningColor(t *testing.T) {
	sd := NewStructuredData("Test")

	// Default running color should be white
	if sd.runningColor != "\033[37m" {
		t.Errorf("Expected default running color to be white \\033[37m, got %q", sd.runningColor)
	}
}

func TestStructuredDataCustomSpinner(t *testing.T) {
	sd := NewStructuredData("Test", WithSpinner(SpinnerThinking))
	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if sd.spinner.FrameCount() != 5 {
		t.Errorf("Expected SpinnerThinking to have 5 frames, got %d", sd.spinner.FrameCount())
	}

	sd.StartRunning()

	// Verify different frames appear
	frames := make(map[string]bool)
	for i := 0; i < 5; i++ {
		view := sd.View()
		frames[view] = true
		sd.Update(structuredDataTickMsg{})
	}

	// Should have seen multiple different frames
	if len(frames) < 2 {
		t.Error("Should see multiple different spinner frames")
	}
}

func TestStructuredDataCustomIconSet(t *testing.T) {
	sd := NewStructuredData("Test", WithIconSet(IconSetSymbols))
	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	sd.MarkSuccess()
	view := sd.View()

	// Should contain the checkmark from IconSetSymbols
	if !strings.Contains(view, "✓") {
		t.Error("View should contain checkmark icon from IconSetSymbols")
	}
}

func TestStructuredDataIconSetClaude(t *testing.T) {
	sd := NewStructuredData("Test", WithIconSet(IconSetClaude))
	sd.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Test each status
	sd.MarkSuccess()
	viewSuccess := sd.View()
	if !strings.Contains(viewSuccess, "✓") {
		t.Error("Success should show checkmark")
	}

	sd.MarkError()
	viewError := sd.View()
	if !strings.Contains(viewError, "✗") {
		t.Error("Error should show X mark")
	}

	sd.MarkWarning()
	viewWarning := sd.View()
	if !strings.Contains(viewWarning, "⚠") {
		t.Error("Warning should show warning symbol")
	}
}

func TestSpinnerGetFrame(t *testing.T) {
	spinner := SpinnerThinking

	// Test frame cycling
	if spinner.GetFrame(0) != "." {
		t.Error("Frame 0 should be '.'")
	}
	if spinner.GetFrame(1) != "*" {
		t.Error("Frame 1 should be '*'")
	}
	if spinner.GetFrame(5) != "." {
		t.Error("Frame 5 should wrap to '.'")
	}
}

func TestSpinnerFrameCount(t *testing.T) {
	if SpinnerThinking.FrameCount() != 5 {
		t.Errorf("SpinnerThinking should have 5 frames, got %d", SpinnerThinking.FrameCount())
	}

	if SpinnerDots.FrameCount() != 10 {
		t.Errorf("SpinnerDots should have 10 frames, got %d", SpinnerDots.FrameCount())
	}

	if SpinnerBlink.FrameCount() != 2 {
		t.Errorf("SpinnerBlink should have 2 frames, got %d", SpinnerBlink.FrameCount())
	}
}
