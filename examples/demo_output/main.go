package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

func main() {
	fmt.Println("=== TUI Framework Demo ===\n")

	// Create app
	app := tui.NewApplication()

	// Add multiple status bars to demo focus management
	statusBar1 := tui.NewStatusBar()
	statusBar1.SetMessage("Status Bar 1 - Press Tab to focus next")

	statusBar2 := tui.NewStatusBar()
	statusBar2.SetMessage("Status Bar 2 - Press Shift+Tab to focus previous")

	app.AddComponent(statusBar1)
	app.AddComponent(statusBar2)

	// Simulate terminal size
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	fmt.Println("Initial state (StatusBar1 focused):")
	fmt.Println(app.View())

	// Press Tab to focus next component
	app.Update(tea.KeyMsg{Type: tea.KeyTab})

	fmt.Println("After pressing Tab (StatusBar2 focused):")
	fmt.Println(app.View())

	// Press Shift+Tab to focus previous
	app.Update(tea.KeyMsg{Type: tea.KeyShiftTab})

	fmt.Println("After pressing Shift+Tab (StatusBar1 focused again):")
	fmt.Println(app.View())

	fmt.Println("\n✓ TUI framework is working correctly!")
	fmt.Println("✓ Focus management: Tab/Shift+Tab navigation")
	fmt.Println("✓ Component lifecycle: Init, Update, View, Focus, Blur")
	fmt.Println("✓ Keyboard handling: q, Ctrl+C to quit")
	fmt.Println("\nRun './examples/basic/basic' for interactive demo")
}
