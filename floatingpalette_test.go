package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func testPaletteCommands() []PaletteCommand {
	return []PaletteCommand{
		{Name: "diff", Description: "Show git diff for current directory"},
		{Name: "sessions", Description: "Switch to a different session"},
		{Name: "subagents", Description: "View subagent session details and diffs"},
	}
}

func TestFloatingPaletteCreation(t *testing.T) {
	fp := NewFloatingPalette(testPaletteCommands())
	if fp == nil {
		t.Fatal("NewFloatingPalette returned nil")
	}

	if fp.maxResults != 10 {
		t.Errorf("expected default maxResults=10, got %d", fp.maxResults)
	}
	if fp.horizontalMargin != 8 {
		t.Errorf("expected default horizontalMargin=8, got %d", fp.horizontalMargin)
	}
	if fp.prompt != "> " {
		t.Errorf("expected default prompt '> ', got %q", fp.prompt)
	}
	if fp.IsVisible() {
		t.Error("palette should not be visible initially")
	}
	if len(fp.filtered) != 3 {
		t.Errorf("expected all commands initially filtered, got %d", len(fp.filtered))
	}
}

func TestFloatingPaletteOptions(t *testing.T) {
	fp := NewFloatingPalette(
		testPaletteCommands(),
		WithFloatingPaletteMaxResults(2),
		WithFloatingPaletteMargin(4),
		WithFloatingPalettePrompt("/ "),
		WithFloatingPaletteTheme("nord"),
	)

	if fp.maxResults != 2 {
		t.Errorf("expected maxResults=2, got %d", fp.maxResults)
	}
	if fp.horizontalMargin != 4 {
		t.Errorf("expected horizontalMargin=4, got %d", fp.horizontalMargin)
	}
	if fp.prompt != "/ " {
		t.Errorf("expected prompt '/ ', got %q", fp.prompt)
	}
	if fp.designTokens == nil {
		t.Error("expected design tokens to be applied")
	}
}

func TestFloatingPaletteShowHideToggle(t *testing.T) {
	fp := NewFloatingPalette(testPaletteCommands())

	fp.Show()
	if !fp.IsVisible() {
		t.Error("palette should be visible after Show")
	}
	if fp.input != "" {
		t.Errorf("expected input reset on Show, got %q", fp.input)
	}

	fp.Hide()
	if fp.IsVisible() {
		t.Error("palette should not be visible after Hide")
	}

	fp.Toggle()
	if !fp.IsVisible() {
		t.Error("palette should be visible after Toggle from hidden")
	}

	fp.Toggle()
	if fp.IsVisible() {
		t.Error("palette should be hidden after Toggle from visible")
	}
}

func TestFloatingPaletteFuzzyFiltering(t *testing.T) {
	fp := NewFloatingPalette(testPaletteCommands())
	fp.Focus()
	fp.Show()

	fp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d', 'i'}})
	if fp.input != "di" {
		t.Fatalf("expected input 'di', got %q", fp.input)
	}

	if len(fp.filtered) != 1 {
		t.Fatalf("expected 1 filtered result for 'di', got %d", len(fp.filtered))
	}
	if fp.filtered[0].Name != "diff" {
		t.Errorf("expected 'diff' to match, got %q", fp.filtered[0].Name)
	}

	fp.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if fp.input != "d" {
		t.Fatalf("expected input 'd' after backspace, got %q", fp.input)
	}
	if len(fp.filtered) != 1 {
		t.Fatalf("expected 1 result for 'd', got %d", len(fp.filtered))
	}
}
func TestFloatingPaletteSelectionNavigation(t *testing.T) {
	fp := NewFloatingPalette(testPaletteCommands())
	fp.Focus()
	fp.Show()

	if fp.selectedIndex != 0 {
		t.Fatalf("expected initial selectedIndex=0, got %d", fp.selectedIndex)
	}

	fp.Update(tea.KeyMsg{Type: tea.KeyDown})
	if fp.selectedIndex != 1 {
		t.Errorf("expected selectedIndex=1 after KeyDown, got %d", fp.selectedIndex)
	}

	fp.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	if fp.selectedIndex != 2 {
		t.Errorf("expected selectedIndex=2 after KeyCtrlN, got %d", fp.selectedIndex)
	}

	fp.Update(tea.KeyMsg{Type: tea.KeyDown})
	if fp.selectedIndex != 2 {
		t.Errorf("expected selectedIndex to stay at end (2), got %d", fp.selectedIndex)
	}

	fp.Update(tea.KeyMsg{Type: tea.KeyUp})
	if fp.selectedIndex != 1 {
		t.Errorf("expected selectedIndex=1 after KeyUp, got %d", fp.selectedIndex)
	}

	fp.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	if fp.selectedIndex != 0 {
		t.Errorf("expected selectedIndex=0 after KeyCtrlP, got %d", fp.selectedIndex)
	}
}

func TestFloatingPaletteEnterConfirmsSelection(t *testing.T) {
	fp := NewFloatingPalette(testPaletteCommands())
	fp.Focus()
	fp.Show()

	// Move to sessions
	fp.Update(tea.KeyMsg{Type: tea.KeyDown})

	_, cmd := fp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected enter to return a command")
	}
	if fp.IsVisible() {
		t.Error("palette should hide after enter")
	}

	msg := cmd()
	selectMsg, ok := msg.(PaletteSelectMsg)
	if !ok {
		t.Fatalf("expected PaletteSelectMsg, got %T", msg)
	}
	if selectMsg.Command.Name != "sessions" {
		t.Errorf("expected selected command 'sessions', got %q", selectMsg.Command.Name)
	}
}

func TestFloatingPaletteEscapeDismisses(t *testing.T) {
	fp := NewFloatingPalette(testPaletteCommands())
	fp.Focus()
	fp.Show()

	_, cmd := fp.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected escape to return a command")
	}
	if fp.IsVisible() {
		t.Error("palette should hide after escape")
	}

	msg := cmd()
	if _, ok := msg.(PaletteDismissMsg); !ok {
		t.Fatalf("expected PaletteDismissMsg, got %T", msg)
	}
}

func TestFloatingPaletteRenderingCentering(t *testing.T) {
	fp := NewFloatingPalette(testPaletteCommands())
	fp.Focus()
	fp.Show()
	fp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := fp.View()
	if view == "" {
		t.Fatal("expected non-empty view when visible")
	}

	lines := strings.Split(view, "\n")
	if len(lines) != 24 {
		t.Fatalf("expected 24 lines to match terminal height, got %d", len(lines))
	}

	contentWidth := 80 - 2*8
	border := strings.Repeat("▄", contentWidth)
	foundBorder := false
	for _, line := range lines {
		if strings.Contains(line, border) {
			foundBorder = true
			if !strings.HasPrefix(line, strings.Repeat(" ", 8)) {
				t.Error("expected border line to include horizontal margin")
			}
			break
		}
	}
	if !foundBorder {
		t.Error("expected to find top border in rendered output")
	}
}

func TestFloatingPaletteRenderingResults(t *testing.T) {
	fp := NewFloatingPalette(testPaletteCommands(), WithFloatingPaletteMaxResults(2))
	fp.Focus()
	fp.Show()
	fp.Update(tea.WindowSizeMsg{Width: 100, Height: 20})

	view := fp.View()
	if !strings.Contains(view, "diff") {
		t.Error("expected rendered output to contain 'diff'")
	}
	if !strings.Contains(view, "sessions") {
		t.Error("expected rendered output to contain 'sessions'")
	}
	if strings.Contains(view, "subagents") {
		t.Error("expected max results limit to hide 'subagents'")
	}
}

func TestFloatingPaletteSetCommandsAndGetSelected(t *testing.T) {
	fp := NewFloatingPalette(nil)
	fp.SetCommands([]PaletteCommand{{Name: "one"}, {Name: "two"}})
	fp.Show()

	selected := fp.GetSelectedCommand()
	if selected == nil || selected.Name != "one" {
		t.Fatalf("expected selected command 'one', got %+v", selected)
	}
}
