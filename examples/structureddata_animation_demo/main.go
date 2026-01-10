package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

type tickMsg time.Time

type model struct {
	sd1    *tui.StructuredData // Running animation
	sd2    *tui.StructuredData // Success
	sd3    *tui.StructuredData // Error
	sd4    *tui.StructuredData // Info
	width  int
	height int
	step   int
}

func initialModel() model {
	// Create structured data components with different statuses
	sd1 := tui.NewStructuredData("Processing Task").
		AddRow("Status", "In progress...").
		AddRow("Items processed", "127/500").
		AddRow("Elapsed time", "2.3s")

	sd2 := tui.NewStructuredData("Completed Task").
		AddRow("Status", "Done").
		AddRow("Items processed", "500/500").
		AddRow("Elapsed time", "8.7s").
		AddRow("Success rate", "100%")

	sd3 := tui.NewStructuredData("Failed Task").
		AddRow("Status", "Failed").
		AddRow("Items processed", "245/500").
		AddRow("Error", "Connection timeout").
		AddRow("Elapsed time", "15.2s")

	sd4 := tui.NewStructuredData("Info Task").
		AddRow("Status", "Informational").
		AddRow("Message", "Configuration loaded").
		AddRow("Version", "1.0.0")

	return model{
		sd1:  sd1,
		sd2:  sd2,
		sd3:  sd3,
		sd4:  sd4,
		step: 0,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.sd1.StartRunning(), // Start blinking animation
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		m.step++

		// Simulate status transitions
		switch m.step {
		case 1:
			m.sd2.MarkSuccess()
		case 2:
			m.sd3.MarkError()
		case 3:
			m.sd4.MarkInfo()
		case 4:
			// Stop running animation after 8 seconds
			m.sd1.MarkSuccess()
		}

		return m, tickCmd()
	}

	// Update all components
	comp1, cmd1 := m.sd1.Update(msg)
	m.sd1 = comp1.(*tui.StructuredData)
	cmds = append(cmds, cmd1)

	comp2, cmd2 := m.sd2.Update(msg)
	m.sd2 = comp2.(*tui.StructuredData)
	cmds = append(cmds, cmd2)

	comp3, cmd3 := m.sd3.Update(msg)
	m.sd3 = comp3.(*tui.StructuredData)
	cmds = append(cmds, cmd3)

	comp4, cmd4 := m.sd4.Update(msg)
	m.sd4 = comp4.(*tui.StructuredData)
	cmds = append(cmds, cmd4)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	s := "\n=== Animated Status Icons Demo ===\n\n"
	s += "Watch the icons animate and change color based on status!\n\n"
	s += "  • Blinking cyan (⏺) = Running\n"
	s += "  • Green (⏺) = Success\n"
	s += "  • Red (⏺) = Error\n"
	s += "  • White (⏺) = Info\n\n"

	s += m.sd1.View() + "\n"
	s += m.sd2.View() + "\n"
	s += m.sd3.View() + "\n"
	s += m.sd4.View() + "\n"

	s += "\nPress 'q' to quit\n"

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
