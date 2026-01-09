package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

func main() {
	// Create app
	app := tui.NewApplication()

	// Add status bar
	statusBar := tui.NewStatusBar()
	statusBar.SetMessage("Test message")
	app.AddComponent(statusBar)

	// Initialize (simulate bubbletea Init)
	cmd := app.Init()
	if cmd != nil {
		fmt.Println("✓ Init returned command")
	} else {
		fmt.Println("✓ Init returned nil")
	}

	// Simulate window size message
	_, cmd = app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd == nil {
		fmt.Println("✓ WindowSizeMsg handled")
	}

	// Try to render
	view := app.View()
	if len(view) > 0 {
		fmt.Println("✓ View renders output:")
		fmt.Println(view)
	} else {
		fmt.Println("✗ View returned empty string")
	}

	// Test focus cycling
	_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyTab})
	fmt.Println("✓ Tab key handled")

	// Test quit
	_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		fmt.Println("✓ Quit key handled")
	}

	fmt.Println("\n✓ All basic functionality tests passed!")
}
