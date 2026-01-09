package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

type model struct {
	app     *tui.Application
	counter int
}

func newModel() model {
	app := tui.NewApplication()

	block := tui.NewToolBlock(
		"Test",
		fmt.Sprintf("Counter: %d", 0),
		[]string{"Press 'r' to restart", "Counter should reset to 0"},
		tui.WithStatus(tui.StatusComplete),
	)
	app.AddComponent(block)

	return model{
		app:     app,
		counter: 0,
	}
}

func (m model) Init() tea.Cmd {
	fmt.Println("Init called, counter:", m.counter)
	return m.app.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		fmt.Println("Key pressed:", msg.String())
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "r":
			fmt.Println("Restarting! Old counter:", m.counter)
			newM := newModel()
			fmt.Println("New model created, counter:", newM.counter)
			return newM, newM.Init()
		case "space":
			m.counter++
			fmt.Println("Counter incremented:", m.counter)
		}
	}

	// Pass to app
	appModel, cmd := m.app.Update(msg)
	m.app = appModel.(*tui.Application)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf("=== Restart Test ===\n\nCounter: %d\n\n%s\n\nPress 'space' to increment · 'r' to restart · 'q' to quit",
		m.counter, m.app.View())
}

func main() {
	fmt.Println("Starting restart test...")
	p := tea.NewProgram(newModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
