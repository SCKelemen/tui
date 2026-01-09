package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

type model struct {
	app   *tui.Application
	block *tui.ToolBlock
	ticks int
}

func newModel() model {
	app := tui.NewApplication()

	// Create a streaming block
	block := tui.NewToolBlock(
		"Bash",
		"go test -v",
		[]string{},
		tui.WithStreaming(),
	)
	app.AddComponent(block)

	return model{
		app:   app,
		block: block,
	}
}

func (m model) Init() tea.Cmd {
	return m.app.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}
	}

	// Count ticks to verify spinner is animating
	typeName := fmt.Sprintf("%T", msg)
	if typeName == "tui.toolBlockTickMsg" {
		m.ticks++
		if m.ticks%10 == 0 {
			// Add a line every 10 ticks to show it's working
			m.block.AppendLine(fmt.Sprintf("Line %d", m.ticks/10))
		}
		if m.ticks > 50 {
			// Stop after 50 ticks (~5 seconds)
			m.block.SetStatus(tui.StatusComplete)
		}
	}

	// Pass to app
	appModel, cmd := m.app.Update(msg)
	m.app = appModel.(*tui.Application)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf("Spinner Test (ticks: %d)\n\n%s\nPress 'q' to quit", m.ticks, m.app.View())
}

func main() {
	p := tea.NewProgram(newModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
