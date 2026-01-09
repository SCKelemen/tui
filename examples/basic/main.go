package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

func main() {
	app := tui.NewApplication()

	// Add a status bar
	statusBar := tui.NewStatusBar()
	statusBar.SetMessage("Welcome to TUI framework example")
	app.AddComponent(statusBar)

	// Run the application
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
