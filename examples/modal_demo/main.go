package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

type model struct {
	app            *tui.Application
	modal          *tui.Modal
	commandPalette *tui.CommandPalette
	statusBar      *tui.StatusBar
	messages       []string
	width          int
	height         int
}

type alertShownMsg struct{}
type confirmResultMsg struct{ confirmed bool }
type inputResultMsg struct{ value string }

func newModel() model {
	app := tui.NewApplication()

	// Create modal
	modal := tui.NewModal()
	app.AddComponent(modal)

	// Create command palette with demo commands
	commands := []tui.Command{
		{
			Name:        "Show Alert",
			Description: "Display an alert modal with OK button",
			Category:    "Modals",
			Keybinding:  "1",
			Action: func() tea.Cmd {
				return func() tea.Msg {
					return alertShownMsg{}
				}
			},
		},
		{
			Name:        "Show Confirmation",
			Description: "Display a confirmation modal with Yes/No buttons",
			Category:    "Modals",
			Keybinding:  "2",
			Action: func() tea.Cmd {
				return func() tea.Msg {
					return confirmResultMsg{}
				}
			},
		},
		{
			Name:        "Show Input",
			Description: "Display an input modal for text entry",
			Category:    "Modals",
			Keybinding:  "3",
			Action: func() tea.Cmd {
				return func() tea.Msg {
					return inputResultMsg{}
				}
			},
		},
		{
			Name:        "Quit",
			Description: "Exit the application",
			Category:    "Application",
			Keybinding:  "q",
			Action: func() tea.Cmd {
				return tea.Quit
			},
		},
	}

	commandPalette := tui.NewCommandPalette(commands)
	app.AddComponent(commandPalette)

	// Status bar
	statusBar := tui.NewStatusBar()
	statusBar.SetMessage("Press 1, 2, or 3 to open different modal types")
	app.AddComponent(statusBar)

	return model{
		app:            app,
		modal:          modal,
		commandPalette: commandPalette,
		statusBar:      statusBar,
		messages:       []string{"Welcome! Try different modal types:"},
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
		// Global shortcuts
		switch msg.String() {
		case "ctrl+k", "ctrl+p":
			if !m.commandPalette.IsVisible() && !m.modal.IsVisible() {
				m.commandPalette.Show()
				return m, nil
			}
		case "1":
			if !m.modal.IsVisible() && !m.commandPalette.IsVisible() {
				// Show alert modal
				m.modal.ShowAlert(
					"Information",
					"This is an alert dialog. It displays important information and has a single OK button.",
					func() tea.Cmd {
						m.messages = append(m.messages, "✓ Alert dismissed")
						m.statusBar.SetMessage("Alert was dismissed")
						return nil
					},
				)
				return m, nil
			}
		case "2":
			if !m.modal.IsVisible() && !m.commandPalette.IsVisible() {
				// Show confirmation modal
				m.modal.ShowConfirm(
					"Confirmation",
					"Are you sure you want to proceed with this action? This is a confirmation dialog with Yes and No buttons.",
					func() tea.Cmd {
						m.messages = append(m.messages, "✓ User selected: Yes")
						m.statusBar.SetMessage("Action confirmed")
						return nil
					},
					func() tea.Cmd {
						m.messages = append(m.messages, "✗ User selected: No")
						m.statusBar.SetMessage("Action cancelled")
						return nil
					},
				)
				return m, nil
			}
		case "3":
			if !m.modal.IsVisible() && !m.commandPalette.IsVisible() {
				// Show input modal
				m.modal.ShowInput(
					"User Input",
					"Please enter your name:",
					"John Doe",
					func(value string) tea.Cmd {
						if value == "" {
							m.messages = append(m.messages, "✗ No input provided")
							m.statusBar.SetMessage("Input was empty")
						} else {
							m.messages = append(m.messages, fmt.Sprintf("✓ User entered: %s", value))
							m.statusBar.SetMessage(fmt.Sprintf("Received input: %s", value))
						}
						return nil
					},
					func() tea.Cmd {
						m.messages = append(m.messages, "✗ Input cancelled")
						m.statusBar.SetMessage("Input was cancelled")
						return nil
					},
				)
				return m, nil
			}
		case "q":
			if !m.modal.IsVisible() && !m.commandPalette.IsVisible() {
				return m, tea.Quit
			}
		}

	case alertShownMsg:
		m.modal.ShowAlert(
			"Alert from Command Palette",
			"You triggered an alert from the command palette!",
			func() tea.Cmd {
				m.messages = append(m.messages, "✓ Command palette alert dismissed")
				return nil
			},
		)

	case confirmResultMsg:
		m.modal.ShowConfirm(
			"Confirm from Command Palette",
			"Do you want to continue?",
			func() tea.Cmd {
				m.messages = append(m.messages, "✓ Confirmed from command palette")
				return nil
			},
			func() tea.Cmd {
				m.messages = append(m.messages, "✗ Cancelled from command palette")
				return nil
			},
		)

	case inputResultMsg:
		m.modal.ShowInput(
			"Input from Command Palette",
			"Enter something:",
			"Type here...",
			func(value string) tea.Cmd {
				m.messages = append(m.messages, fmt.Sprintf("✓ Input from palette: %s", value))
				return nil
			},
			func() tea.Cmd {
				m.messages = append(m.messages, "✗ Input from palette cancelled")
				return nil
			},
		)
	}

	// Pass to app
	appModel, cmd := m.app.Update(msg)
	m.app = appModel.(*tui.Application)
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("\033[1m=== Modal Demo ===\033[0m\n\n")

	b.WriteString(m.app.View())

	// Message history
	b.WriteString("\n\033[2m┌─ Action History ─")
	b.WriteString(strings.Repeat("─", 50))
	b.WriteString("┐\033[0m\n")

	// Show last 5 messages
	startIdx := 0
	if len(m.messages) > 5 {
		startIdx = len(m.messages) - 5
	}

	for i := startIdx; i < len(m.messages); i++ {
		b.WriteString("\033[2m│\033[0m ")
		b.WriteString(m.messages[i])
		b.WriteString("\n")
	}

	b.WriteString("\033[2m└")
	b.WriteString(strings.Repeat("─", 69))
	b.WriteString("┘\033[0m\n\n")

	b.WriteString("\033[2m")
	b.WriteString("Modal Types:\n")
	b.WriteString("  1: Alert    - Information message with OK button\n")
	b.WriteString("  2: Confirm  - Yes/No question dialog\n")
	b.WriteString("  3: Input    - Text input with OK/Cancel\n")
	b.WriteString("\n")
	b.WriteString("Ctrl+K: Command Palette · q: Quit\n")
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
