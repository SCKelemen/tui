package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

type model struct {
	app          *tui.Application
	fileExplorer *tui.FileExplorer
	statusBar    *tui.StatusBar
	width        int
	height       int
}

func newModel() model {
	app := tui.NewApplication()

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	// Create file explorer starting at current directory
	fileExplorer := tui.NewFileExplorer(cwd, tui.WithShowHidden(false))
	app.AddComponent(fileExplorer)

	// Add status bar to show selected path
	statusBar := tui.NewStatusBar()
	statusBar.SetMessage("Browse files and directories")
	app.AddComponent(statusBar)

	return model{
		app:          app,
		fileExplorer: fileExplorer,
		statusBar:    statusBar,
	}
}

func (m model) Init() tea.Cmd {
	return m.app.Init()
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

		// Update status bar with selected path
		if m.fileExplorer.GetSelectedNode() != nil {
			node := m.fileExplorer.GetSelectedNode()
			fileType := "File"
			if node.IsDir {
				fileType = "Directory"
			}
			m.statusBar.SetMessage(fmt.Sprintf("%s: %s", fileType, node.Path))
		}
	}

	// Pass to app
	appModel, cmd := m.app.Update(msg)
	m.app = appModel.(*tui.Application)
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("\033[1m=== File Explorer Demo ===\033[0m\n\n")

	b.WriteString(m.app.View())

	b.WriteString("\n\033[2m")
	b.WriteString("Navigate through your file system:\n")
	b.WriteString("  • ↑/↓ or j/k: Move selection\n")
	b.WriteString("  • →/l or Enter: Expand directory\n")
	b.WriteString("  • ←/h: Collapse directory or move to parent\n")
	b.WriteString("  • .: Toggle hidden files\n")
	b.WriteString("  • r: Refresh current directory\n")
	b.WriteString("  • q: Quit\n")
	b.WriteString("\033[0m")

	return b.String()
}

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
