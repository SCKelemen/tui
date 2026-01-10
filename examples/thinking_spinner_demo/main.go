package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

type tickMsg time.Time

type model struct {
	sd     *tui.StructuredData
	width  int
	height int
}

func initialModel() model {
	// Create with Claude's thinking spinner
	sd := tui.NewStructuredData("Claude is thinking",
		tui.WithSpinner(tui.SpinnerThinking),
		tui.WithIconSet(tui.IconSetClaude))

	sd.AddRow("Analyzing", "Your request...")
	sd.AddRow("Status", "Processing")

	return model{
		sd: sd,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.sd.StartRunning(),
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*10, func(t time.Time) tea.Msg {
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
		case "s":
			m.sd.MarkSuccess()
		case "e":
			m.sd.MarkError()
		case "w":
			m.sd.MarkWarning()
		case "r":
			return m, m.sd.StartRunning()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		// Mark as success after 10 seconds
		m.sd.MarkSuccess()
		m.sd.Clear()
		m.sd.AddRow("Status", "Complete!")
		m.sd.AddRow("Result", "✓ Task finished successfully")
	}

	// Update component
	comp, cmd := m.sd.Update(msg)
	m.sd = comp.(*tui.StructuredData)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	s := "\n=== Claude's Thinking Spinner Demo ===\n\n"
	s += "Watch the thinking animation: . * ÷ + •\n\n"
	s += m.sd.View() + "\n\n"
	s += "Press:\n"
	s += "  'r' to restart animation\n"
	s += "  's' for success (✓)\n"
	s += "  'e' for error (✗)\n"
	s += "  'w' for warning (⚠)\n"
	s += "  'q' to quit\n"

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
