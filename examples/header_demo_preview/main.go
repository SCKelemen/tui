package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

func main() {
	// Create header with two columns like Codex CLI
	header := tui.NewHeader(
		tui.WithColumns(
			// Left column: centered content
			tui.HeaderColumn{
				Width: 40, // Percentage
				Align: tui.AlignCenter,
				Content: []string{
					"",
					"Welcome back!",
					"",
					"▐▛███▜▌",
					"▝▜█████▛▘",
					"▘▘ ▝▝",
					"",
					"TUI v1.0.0",
					"~/Code/github.com/SCKelemen/tui",
				},
			},
			// Right column: left-aligned sections
			tui.HeaderColumn{
				Width: 60, // Percentage
				Align: tui.AlignLeft,
			},
		),
		tui.WithColumnSections(1,
			tui.HeaderSection{
				Title:   "Tips for getting started",
				Content: []string{
					"Use Tab to navigate between components",
					"Press q to quit applications",
				},
			},
			tui.HeaderSection{
				Title:   "Recent activity",
				Content: []string{
					"No recent activity",
				},
				Divider: true,
			},
		),
		tui.WithVerticalDivider(true),
	)

	// Simulate terminal size
	header.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	fmt.Println("\n=== Header Demo Preview ===")
	fmt.Print(header.View())
	fmt.Println("\n(Run with: go run main.go for interactive version)")
}
