package main

import (
	"fmt"
	"strings"

	"github.com/SCKelemen/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	fmt.Println("=== Claude Code-Style Input Components Demo ===")

	// 1. TextInput Component
	fmt.Println("1. TextInput Component (Multi-line text entry):")

	textInput := tui.NewTextInput()
	textInput.SetValue("This is a sample message\nwith multiple lines\nfor demonstration.")

	// Simulate window size
	textInput.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	textInput.Focus()

	fmt.Println(textInput.View())

	// 2. CommandPalette Component
	fmt.Println("\n2. CommandPalette Component (Quick command launcher):")

	commands := []tui.Command{
		{
			Name:        "Clear Messages",
			Description: "Clear all message history",
			Category:    "Edit",
			Keybinding:  "Ctrl+L",
		},
		{
			Name:        "Toggle Activity Bar",
			Description: "Start/stop activity animation",
			Category:    "View",
			Keybinding:  "Ctrl+A",
		},
		{
			Name:        "Add Tool Block",
			Description: "Add a sample tool execution result",
			Category:    "Debug",
		},
		{
			Name:        "Open File",
			Description: "Open a file for editing",
			Category:    "File",
			Keybinding:  "Ctrl+O",
		},
		{
			Name:        "Save File",
			Description: "Save the current file",
			Category:    "File",
			Keybinding:  "Ctrl+S",
		},
		{
			Name:        "Find in Files",
			Description: "Search across all project files",
			Category:    "Search",
			Keybinding:  "Ctrl+Shift+F",
		},
		{
			Name:        "Quit",
			Description: "Exit the application",
			Category:    "Application",
			Keybinding:  "q",
		},
	}

	palette := tui.NewCommandPalette(commands)
	palette.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	palette.Focus()
	palette.Show()

	fmt.Println(palette.View())

	// 3. Command Palette with Search
	fmt.Println("\n3. CommandPalette with Search Filter ('file'):")

	palette2 := tui.NewCommandPalette(commands)
	palette2.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	palette2.Focus()
	palette2.Show()

	// Simulate typing "file"
	palette2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	palette2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	palette2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	palette2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	fmt.Println(palette2.View())

	// 4. Feature Summary
	fmt.Println("\n✓ All input components rendering correctly!")
	fmt.Println("Features:")
	fmt.Println("  • TextInput: Multi-line text entry with Ctrl+J submit, Ctrl+D clear")
	fmt.Println("  • CommandPalette: Quick command launcher with fuzzy search")
	fmt.Println("  • Keyboard shortcuts: Ctrl+K/Ctrl+P to open palette")
	fmt.Println("  • Navigation: Up/Down arrows, Enter to select, Esc to close")
	fmt.Println("  • Search filtering: Type to narrow command list")
	fmt.Println("  • Visual feedback: Selected items highlighted")
	fmt.Println("\nRun 'go run examples/input_demo/main.go' for interactive demo")

	// 5. Integration Example
	fmt.Println("\n\n=== Integration with Other Components ===")

	// Show how it looks with activity bar and tool blocks
	activityBar := tui.NewActivityBar()
	activityBar.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	activityBar.Start("Processing your request...")
	activityBar.SetProgress("↓ 3.2k tokens")

	fmt.Println(activityBar.View())

	toolBlock := tui.NewToolBlock(
		"Bash",
		"git status",
		[]string{
			"On branch main",
			"Your branch is up to date with 'origin/main'.",
			"",
			"nothing to commit, working tree clean",
		},
		tui.WithMaxLines(2),
	)
	toolBlock.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	fmt.Println(toolBlock.View())

	var b strings.Builder
	b.WriteString("\033[2m┌─ Message History ")
	b.WriteString(strings.Repeat("─", 60))
	b.WriteString("┐\033[0m\n")
	b.WriteString("\033[2m│\033[0m You: What's the git status?\n")
	b.WriteString("\033[2m│\033[0m Bot: Let me check that for you...\n")
	b.WriteString("\033[2m└")
	b.WriteString(strings.Repeat("─", 78))
	b.WriteString("┘\033[0m\n")

	fmt.Println(b.String())

	fmt.Println(textInput.View())

	fmt.Println("\033[2mCtrl+K: Command Palette · Ctrl+J: Send Message · q: Quit\033[0m")
}
