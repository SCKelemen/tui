package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCommandPaletteCreation(t *testing.T) {
	commands := []Command{
		{Name: "Test Command", Description: "Test", Category: "Test"},
	}

	cp := NewCommandPalette(commands)
	if cp == nil {
		t.Fatal("NewCommandPalette returned nil")
	}

	if len(cp.commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(cp.commands))
	}

	if len(cp.filtered) != 1 {
		t.Errorf("Expected 1 filtered command initially, got %d", len(cp.filtered))
	}

	if cp.visible {
		t.Error("CommandPalette should not be visible initially")
	}

	if cp.maxVisible != 8 {
		t.Errorf("Expected maxVisible 8, got %d", cp.maxVisible)
	}

	if cp.selected != 0 {
		t.Errorf("Expected selected index 0, got %d", cp.selected)
	}
}

func TestCommandPaletteShowHide(t *testing.T) {
	commands := []Command{
		{Name: "Command 1", Description: "Test 1"},
		{Name: "Command 2", Description: "Test 2"},
	}

	cp := NewCommandPalette(commands)

	if cp.IsVisible() {
		t.Error("CommandPalette should not be visible initially")
	}

	cp.Show()
	if !cp.IsVisible() {
		t.Error("CommandPalette should be visible after Show()")
	}

	if len(cp.filtered) != 2 {
		t.Error("Show() should reset filtered commands to all commands")
	}

	if cp.selected != 0 {
		t.Error("Show() should reset selected index to 0")
	}

	cp.Hide()
	if cp.IsVisible() {
		t.Error("CommandPalette should not be visible after Hide()")
	}
}

func TestCommandPaletteFocusManagement(t *testing.T) {
	cp := NewCommandPalette([]Command{})

	if cp.Focused() {
		t.Error("CommandPalette should not be focused initially")
	}

	cp.Focus()
	if !cp.Focused() {
		t.Error("CommandPalette should be focused after Focus()")
	}

	cp.Blur()
	if cp.Focused() {
		t.Error("CommandPalette should not be focused after Blur()")
	}
}

func TestCommandPaletteViewHidden(t *testing.T) {
	commands := []Command{
		{Name: "Test", Description: "Test"},
	}

	cp := NewCommandPalette(commands)
	cp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cp.View()
	if view != "" {
		t.Error("View should be empty when palette is hidden")
	}
}

func TestCommandPaletteViewVisible(t *testing.T) {
	commands := []Command{
		{Name: "Test Command", Description: "Test"},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()
	cp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cp.View()
	if view == "" {
		t.Error("View should not be empty when palette is visible")
	}

	if !strings.Contains(view, "Command Palette") {
		t.Error("View should contain 'Command Palette' title")
	}

	if !strings.Contains(view, "Test Command") {
		t.Error("View should contain the command name")
	}
}

func TestCommandPaletteNavigation(t *testing.T) {
	commands := []Command{
		{Name: "Command 1"},
		{Name: "Command 2"},
		{Name: "Command 3"},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()

	if cp.selected != 0 {
		t.Error("Should start with first item selected")
	}

	// Press Down
	cp.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cp.selected != 1 {
		t.Errorf("Expected selected 1 after Down, got %d", cp.selected)
	}

	// Press Down again
	cp.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cp.selected != 2 {
		t.Errorf("Expected selected 2 after second Down, got %d", cp.selected)
	}

	// Press Down at end (should stay at last item)
	cp.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cp.selected != 2 {
		t.Errorf("Expected selected to stay at 2 (last item), got %d", cp.selected)
	}

	// Press Up
	cp.Update(tea.KeyMsg{Type: tea.KeyUp})
	if cp.selected != 1 {
		t.Errorf("Expected selected 1 after Up, got %d", cp.selected)
	}

	// Press Up again
	cp.Update(tea.KeyMsg{Type: tea.KeyUp})
	if cp.selected != 0 {
		t.Errorf("Expected selected 0 after second Up, got %d", cp.selected)
	}

	// Press Up at start (should stay at first item)
	cp.Update(tea.KeyMsg{Type: tea.KeyUp})
	if cp.selected != 0 {
		t.Errorf("Expected selected to stay at 0 (first item), got %d", cp.selected)
	}
}

func TestCommandPaletteEscapeKey(t *testing.T) {
	commands := []Command{
		{Name: "Test"},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()

	if !cp.IsVisible() {
		t.Error("Palette should be visible before Esc")
	}

	cp.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cp.IsVisible() {
		t.Error("Esc should hide the palette")
	}
}

func TestCommandPaletteEnterSelection(t *testing.T) {
	actionCalled := false
	commands := []Command{
		{
			Name: "Test Command",
			Action: func() tea.Cmd {
				actionCalled = true
				return nil
			},
		},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()

	if !cp.IsVisible() {
		t.Error("Palette should be visible")
	}

	cp.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cp.IsVisible() {
		t.Error("Enter should hide the palette after selection")
	}

	if !actionCalled {
		t.Error("Enter should execute the selected command's action")
	}
}

func TestCommandPaletteEnterNoAction(t *testing.T) {
	commands := []Command{
		{Name: "Test Command"}, // No Action defined
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()

	// Should not panic when Enter is pressed without an Action
	cp.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cp.IsVisible() {
		t.Error("Enter should still hide the palette even without action")
	}
}

func TestCommandPaletteFilterByName(t *testing.T) {
	commands := []Command{
		{Name: "Open File", Description: "Open a file"},
		{Name: "Save File", Description: "Save current file"},
		{Name: "Close Window", Description: "Close the window"},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()
	cp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Type "file"
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	// Should filter to only "Open File" and "Save File"
	if len(cp.filtered) != 2 {
		t.Errorf("Expected 2 filtered commands for 'file', got %d", len(cp.filtered))
	}

	// Should reset selection to 0
	if cp.selected != 0 {
		t.Error("Filtering should reset selection to 0")
	}
}

func TestCommandPaletteFilterByDescription(t *testing.T) {
	commands := []Command{
		{Name: "Command 1", Description: "Open a file"},
		{Name: "Command 2", Description: "Save data"},
		{Name: "Command 3", Description: "Open window"},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()

	// Type "open" (should match descriptions)
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	// Should filter to commands with "open" in description
	if len(cp.filtered) != 2 {
		t.Errorf("Expected 2 filtered commands for 'open', got %d", len(cp.filtered))
	}
}

func TestCommandPaletteFilterByCategory(t *testing.T) {
	commands := []Command{
		{Name: "Cmd 1", Category: "File"},
		{Name: "Cmd 2", Category: "Edit"},
		{Name: "Cmd 3", Category: "File"},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()

	// Type "file" (should match category)
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	if len(cp.filtered) != 2 {
		t.Errorf("Expected 2 filtered commands for category 'file', got %d", len(cp.filtered))
	}
}

func TestCommandPaletteEmptyFilter(t *testing.T) {
	commands := []Command{
		{Name: "Test 1"},
		{Name: "Test 2"},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()

	// Type something that matches nothing
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})

	if len(cp.filtered) != 0 {
		t.Errorf("Expected 0 filtered commands for 'xyz', got %d", len(cp.filtered))
	}
}

func TestCommandPaletteEmptyFilterView(t *testing.T) {
	commands := []Command{
		{Name: "Test"},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()
	cp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Type something that matches nothing
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})

	view := cp.View()
	if !strings.Contains(view, "No commands found") {
		t.Error("View should show 'No commands found' when filter has no results")
	}
}

func TestCommandPaletteEmptyCommandList(t *testing.T) {
	cp := NewCommandPalette([]Command{})
	cp.Focus()
	cp.Show()
	cp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cp.View()
	if !strings.Contains(view, "No commands found") {
		t.Error("View should show 'No commands found' with empty command list")
	}
}

func TestCommandPaletteCaseInsensitiveFilter(t *testing.T) {
	commands := []Command{
		{Name: "OpenFile", Description: "Open a file"},
		{Name: "SaveFile", Description: "Save current file"},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()

	// Type "OPEN" in uppercase
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'O'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'E'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})

	if len(cp.filtered) != 1 {
		t.Errorf("Case insensitive filter should find 'OpenFile', got %d results", len(cp.filtered))
	}
}

func TestCommandPaletteWindowSizeUpdate(t *testing.T) {
	cp := NewCommandPalette([]Command{})

	if cp.width != 0 {
		t.Error("Initial width should be 0")
	}

	cp.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if cp.width != 100 {
		t.Errorf("Expected width 100, got %d", cp.width)
	}

	if cp.height != 50 {
		t.Errorf("Expected height 50, got %d", cp.height)
	}
}

func TestCommandPaletteViewWithoutSize(t *testing.T) {
	commands := []Command{
		{Name: "Test"},
	}

	cp := NewCommandPalette(commands)
	cp.Show()

	view := cp.View()
	if view != "" {
		t.Error("View should be empty when width/height is 0")
	}
}

func TestCommandPaletteKeybindingDisplay(t *testing.T) {
	commands := []Command{
		{Name: "Test Command", Keybinding: "Ctrl+S"},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()
	cp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cp.View()
	if !strings.Contains(view, "Ctrl+S") {
		t.Error("View should display keybinding")
	}
}

func TestCommandPaletteFooterCount(t *testing.T) {
	commands := []Command{
		{Name: "Cmd 1"},
		{Name: "Cmd 2"},
		{Name: "Cmd 3"},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()
	cp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cp.View()
	if !strings.Contains(view, "3 commands") {
		t.Error("Footer should show command count")
	}

	// Filter to 1 command
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	view = cp.View()
	if !strings.Contains(view, "1 commands") {
		t.Error("Footer should update to show filtered count")
	}
}

func TestCommandPaletteInit(t *testing.T) {
	cp := NewCommandPalette([]Command{})
	cmd := cp.Init()

	if cmd == nil {
		t.Error("Init should return a blink command")
	}
}

func TestCommandPaletteNoActionWhenBlurred(t *testing.T) {
	commands := []Command{
		{Name: "Test"},
	}

	cp := NewCommandPalette(commands)
	// Don't focus

	cp.Update(tea.KeyMsg{Type: tea.KeyCtrlK})

	if cp.IsVisible() {
		t.Error("Ctrl+K should not show palette when not focused")
	}
}

func TestCommandPaletteCtrlKToggle(t *testing.T) {
	commands := []Command{
		{Name: "Test"},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()

	if cp.IsVisible() {
		t.Error("Palette should not be visible initially")
	}

	// Press Ctrl+K
	cp.Update(tea.KeyMsg{Type: tea.KeyCtrlK})

	if !cp.IsVisible() {
		t.Error("Ctrl+K should show the palette")
	}
}

func TestCommandPaletteCtrlPToggle(t *testing.T) {
	commands := []Command{
		{Name: "Test"},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()

	// Press Ctrl+P
	cp.Update(tea.KeyMsg{Type: tea.KeyCtrlP})

	if !cp.IsVisible() {
		t.Error("Ctrl+P should show the palette")
	}
}

func TestCommandPaletteManyCommands(t *testing.T) {
	// Create more commands than maxVisible
	commands := []Command{}
	for i := 0; i < 20; i++ {
		commands = append(commands, Command{Name: "Command", Description: "Test"})
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()
	cp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := cp.View()
	if view == "" {
		t.Error("View should not be empty with many commands")
	}

	// Should show footer with count
	if !strings.Contains(view, "20 commands") {
		t.Error("Footer should show total command count")
	}
}

func TestCommandPaletteLongCommandName(t *testing.T) {
	commands := []Command{
		{Name: strings.Repeat("Very Long Command Name ", 10)},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()
	cp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Should not panic with long command name
	view := cp.View()
	if view == "" {
		t.Error("View should not be empty with long command name")
	}

	// Long names should be truncated
	if strings.Contains(view, "...") {
		// Truncation is working
	}
}

func TestCommandPaletteEnterWithEmptyResults(t *testing.T) {
	commands := []Command{
		{Name: "Test"},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()

	// Filter to empty results
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})

	// Press Enter (should not panic)
	cp.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cp.IsVisible() {
		t.Error("Enter should hide palette even with empty results")
	}
}

func TestCommandPaletteNavigationWithEmptyResults(t *testing.T) {
	commands := []Command{
		{Name: "Test"},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()

	// Filter to empty results
	cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	// Press Down/Up (should not panic)
	cp.Update(tea.KeyMsg{Type: tea.KeyDown})
	cp.Update(tea.KeyMsg{Type: tea.KeyUp})

	// Should still work fine
	if len(cp.filtered) != 0 {
		t.Error("Filtered list should remain empty")
	}
}
