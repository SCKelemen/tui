package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

// model wraps the TUI app with message history
type model struct {
	app            *tui.Application
	textInput      *tui.TextInput
	commandPalette *tui.CommandPalette
	activityBar    *tui.ActivityBar
	messages       []string
	toolBlocks     []*tui.ToolBlock
}

type messageSubmittedMsg string

func newModel() model {
	app := tui.NewApplication()

	// Activity bar
	activityBar := tui.NewActivityBar()
	app.AddComponent(activityBar)

	// Create command palette with sample commands
	commands := []tui.Command{
		{
			Name:        "Clear Messages",
			Description: "Clear all message history",
			Category:    "Edit",
			Keybinding:  "Ctrl+L",
			Action: func() tea.Cmd {
				return func() tea.Msg {
					return clearMessagesMsg{}
				}
			},
		},
		{
			Name:        "Toggle Activity Bar",
			Description: "Start/stop activity animation",
			Category:    "View",
			Keybinding:  "Ctrl+A",
			Action: func() tea.Cmd {
				return func() tea.Msg {
					return toggleActivityMsg{}
				}
			},
		},
		{
			Name:        "Add Sample Tool Block",
			Description: "Add a sample tool execution result",
			Category:    "Debug",
			Keybinding:  "",
			Action: func() tea.Cmd {
				return func() tea.Msg {
					return addToolBlockMsg{}
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

	// Text input for user messages
	textInput := tui.NewTextInput()
	textInput.OnSubmit(func(text string) tea.Cmd {
		return func() tea.Msg {
			return messageSubmittedMsg(text)
		}
	})
	app.AddComponent(textInput)

	// IMPORTANT: Focus the text input so user can type
	// (By default, first component added gets focus, which is activityBar)
	app.FocusComponent(2) // Index 2 is textInput (0=activityBar, 1=commandPalette, 2=textInput)

	return model{
		app:            app,
		textInput:      textInput,
		commandPalette: commandPalette,
		activityBar:    activityBar,
		messages:       []string{"Welcome! Type your message below and press Ctrl+J to send."},
		toolBlocks:     []*tui.ToolBlock{},
	}
}

type clearMessagesMsg struct{}
type toggleActivityMsg struct{}
type addToolBlockMsg struct{}

func (m model) Init() tea.Cmd {
	return m.app.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global shortcuts
		switch msg.String() {
		case "ctrl+k", "ctrl+p":
			if !m.commandPalette.IsVisible() {
				m.commandPalette.Show()
				return m, nil
			}
		case "q":
			if !m.commandPalette.IsVisible() && !m.textInput.Focused() {
				return m, tea.Quit
			}
		}

	case messageSubmittedMsg:
		m.messages = append(m.messages, fmt.Sprintf("You: %s", string(msg)))
		m.messages = append(m.messages, fmt.Sprintf("Bot: Echo - %s", string(msg)))

	case clearMessagesMsg:
		m.messages = []string{"Messages cleared."}

	case toggleActivityMsg:
		if m.activityBar.Focused() {
			m.activityBar.Stop()
			return m, nil
		} else {
			activityCmd := m.activityBar.Start("Processing...")
			return m, activityCmd
		}

	case addToolBlockMsg:
		block := tui.NewToolBlock(
			"Bash",
			"echo 'Hello from command palette'",
			[]string{
				"Hello from command palette",
				"Command executed successfully",
			},
			tui.WithMaxLines(2),
		)
		m.toolBlocks = append(m.toolBlocks, block)
		m.app.AddComponent(block)
	}

	// Pass to app
	appModel, cmd := m.app.Update(msg)
	m.app = appModel.(*tui.Application)
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	// App components (activity bar, command palette, tool blocks)
	b.WriteString(m.app.View())

	// Message history (scrollable area in the middle)
	b.WriteString("\n\033[2m┌─ Message History ─")
	b.WriteString(strings.Repeat("─", 60))
	b.WriteString("┐\033[0m\n")

	// Show last 10 messages
	startIdx := 0
	if len(m.messages) > 10 {
		startIdx = len(m.messages) - 10
	}

	for i := startIdx; i < len(m.messages); i++ {
		b.WriteString("\033[2m│\033[0m ")
		b.WriteString(m.messages[i])
		b.WriteString("\n")
	}

	b.WriteString("\033[2m└")
	b.WriteString(strings.Repeat("─", 78))
	b.WriteString("┘\033[0m\n\n")

	// Keybinding hints
	b.WriteString("\033[2mCtrl+K: Command Palette · Ctrl+J: Send Message · q: Quit\033[0m\n")

	return b.String()
}

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
