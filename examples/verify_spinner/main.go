package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

type model struct {
	app     *tui.Application
	block1  *tui.ToolBlock
	block2  *tui.ToolBlock
	elapsed time.Duration
	start   time.Time
}

func newModel() model {
	app := tui.NewApplication()

	// Two streaming blocks
	block1 := tui.NewToolBlock(
		"Bash",
		"npm install",
		[]string{},
		tui.WithStreaming(),
	)
	app.AddComponent(block1)

	block2 := tui.NewToolBlock(
		"Bash",
		"go test -v",
		[]string{},
		tui.WithStreaming(),
	)
	app.AddComponent(block2)

	return model{
		app:    app,
		block1: block1,
		block2: block2,
		start:  time.Now(),
	}
}

type addLineMsg struct{}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.app.Init(),
		addLineTick(),
	)
}

func addLineTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return addLineMsg{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}

	case addLineMsg:
		m.elapsed = time.Since(m.start)

		// Add lines to show progress
		if m.elapsed < 3*time.Second {
			m.block1.AppendLine(fmt.Sprintf("Installing package %d...", int(m.elapsed.Seconds())+1))
			m.block2.AppendLine(fmt.Sprintf("Running test %d...", int(m.elapsed.Seconds())+1))
			return m, addLineTick()
		} else if m.elapsed < 4*time.Second {
			m.block1.SetStatus(tui.StatusComplete)
			m.block2.SetStatus(tui.StatusComplete)
			return m, addLineTick()
		}
	}

	// Pass to app
	appModel, cmd := m.app.Update(msg)
	m.app = appModel.(*tui.Application)
	return m, cmd
}

func (m model) View() string {
	header := fmt.Sprintf("\033[1m=== Spinner Animation Verification ===\033[0m\n\n")
	header += "Watch the spinners (⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏) animate while streaming.\n"
	header += "Both blocks should animate independently.\n\n"
	return header + m.app.View() + "\n\nPress 'q' to quit"
}

func main() {
	p := tea.NewProgram(newModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
