package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

type model struct {
	header *tui.Header
	width  int
	height int
}

func newModel() model {
	// Create header with two columns like Claude Code
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

	return model{
		header: header,
	}
}

func (m model) Init() tea.Cmd {
	return m.header.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	// Pass to header
	var cmd tea.Cmd
	var component tui.Component
	component, cmd = m.header.Update(msg)
	m.header = component.(*tui.Header)

	return m, cmd
}

func (m model) View() string {
	return "\033[1m=== Header Demo ===\033[0m\n\n" +
		m.header.View() +
		"\n\033[2mPress q to quit\033[0m\n"
}

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
