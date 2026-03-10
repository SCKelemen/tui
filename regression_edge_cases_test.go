package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func assertNotPanics(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	fn()
}

func TestTinyWidthRenderingDoesNotPanic(t *testing.T) {
	t.Run("CommandPalette", func(t *testing.T) {
		assertNotPanics(t, func() {
			cp := NewCommandPalette([]Command{{Name: "Test"}})
			cp.Focus()
			cp.Show()
			cp.Update(tea.WindowSizeMsg{Width: 1, Height: 3})
			_ = cp.View()
		})
	})

	t.Run("Modal", func(t *testing.T) {
		assertNotPanics(t, func() {
			modal := NewModal()
			modal.Focus()
			modal.ShowAlert("Title", "Message", nil)
			modal.Update(tea.WindowSizeMsg{Width: 1, Height: 3})
			_ = modal.View()
		})
	})

	t.Run("DetailModal", func(t *testing.T) {
		assertNotPanics(t, func() {
			modal := NewDetailModal()
			modal.Show()
			modal.Update(tea.WindowSizeMsg{Width: 1, Height: 3})
			_ = modal.View()
		})
	})

	t.Run("StatCard", func(t *testing.T) {
		assertNotPanics(t, func() {
			card := NewStatCard(WithTitle("CPU"), WithValue("42%"))
			card.Update(tea.WindowSizeMsg{Width: 1, Height: 3})
			_ = card.View()
		})
	})

	t.Run("TextInput", func(t *testing.T) {
		assertNotPanics(t, func() {
			input := NewTextInput()
			input.Update(tea.WindowSizeMsg{Width: 1, Height: 3})
			_ = input.View()
		})
	})

	t.Run("ToolBlockTinyWidth", func(t *testing.T) {
		assertNotPanics(t, func() {
			block := NewToolBlock("Bash", "echo hi", []string{"one line"})
			block.Update(tea.WindowSizeMsg{Width: 1, Height: 3})
			_ = block.View()
		})
	})
}

func TestToolBlockLongToolNameDoesNotPanic(t *testing.T) {
	assertNotPanics(t, func() {
		block := NewToolBlock(strings.Repeat("T", 120), "echo hello", []string{"out"})
		block.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		_ = block.View()
	})
}

func TestCommandPaletteFuzzyMatchWithGaps(t *testing.T) {
	commands := []Command{
		{Name: "Save File"},
		{Name: "Open Window"},
	}

	cp := NewCommandPalette(commands)
	cp.Focus()
	cp.Show()

	for _, r := range []rune("sf") {
		cp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	if len(cp.filtered) == 0 {
		t.Fatal("expected fuzzy match for query 'sf'")
	}
	if cp.filtered[0].Name != "Save File" {
		t.Fatalf("expected 'Save File' first, got %q", cp.filtered[0].Name)
	}
}
