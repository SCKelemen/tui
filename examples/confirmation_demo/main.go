package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

type model struct {
	confirmBlock *tui.ConfirmationBlock
}

func initialModel() model {
	yamlContent := `apiVersion: ownership.kelemen.com/v1
kind: Group
tags:
  - "team"
spec:
  id: Group:Test-TEAM@kelemen.com
  name: Test Team
  description: Team with validation issues
  type: Team
  email: test-team@kelemen.com
  slack: "#test"
  state: Active
  parent: group:security@kelemen.com
  owners:
    - user:Charlie@kelemen.com
    - user:Alice@kelemen.com
  members:
    - user:Bob@kelemen.com
    - user:Alice@kelemen.com`

	confirmBlock := tui.NewConfirmationBlock(
		tui.WithConfirmOperation("Write"),
		tui.WithConfirmFilepath("~/Code/github.com/SCKelemen/yaml-lsp/data/test-issues.yaml"),
		tui.WithConfirmDescription("Create file ../yaml-lsp/data/test-issues.yaml"),
		tui.WithConfirmCode(yamlContent),
		tui.WithConfirmOptions([]string{
			"Yes",
			"Yes, allow all edits in data/ during this session (shift+tab)",
			"No",
		}),
		tui.WithConfirmFooterHints([]string{
			"Esc to cancel",
			"Tab to add additional instructions",
		}),
	)
	confirmBlock.Focus()

	return model{
		confirmBlock: confirmBlock,
	}
}

func (m model) Init() tea.Cmd {
	return m.confirmBlock.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Check if confirmed
		if m.confirmBlock.IsConfirmed() {
			selection := m.confirmBlock.GetSelection()
			if selection == -1 {
				// Cancelled
				return m, tea.Quit
			}
			// Wait for another key press after confirmation
			if msg.String() == "enter" || msg.String() == "q" {
				return m, tea.Quit
			}
		}
	}

	// Pass to confirmation block
	var cmd tea.Cmd
	var component tui.Component
	component, cmd = m.confirmBlock.Update(msg)
	m.confirmBlock = component.(*tui.ConfirmationBlock)

	// If just confirmed, return quit command
	if m.confirmBlock.IsConfirmed() {
		return m, tea.Quit
	}

	return m, cmd
}

func (m model) View() string {
	return m.confirmBlock.View()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
