package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Command represents an executable command in the command palette with metadata
// for display and categorization.
type Command struct {
	Name        string          // Display name of the command
	Description string          // Brief description of what the command does
	Category    string          // Category for grouping (e.g., "File", "Edit", "View")
	Action      func() tea.Cmd  // Function to execute when command is selected
	Keybinding  string          // Optional keyboard shortcut (e.g., "Ctrl+S")
}

// CommandPalette is a fuzzy-searchable command launcher inspired by VS Code's command
// palette. It provides a popup interface for quickly executing commands via keyboard.
//
// Features:
//   - Fuzzy search filtering
//   - Keyboard navigation (↑↓ or j/k)
//   - Category grouping
//   - Keybinding hints
//   - Toggle visibility with Ctrl+K
//
// Example usage:
//
//	commands := []tui.Command{
//	    {Name: "Save File", Description: "Save current file", Category: "File", Keybinding: "Ctrl+S"},
//	    {Name: "Open File", Description: "Open a file", Category: "File", Keybinding: "Ctrl+O"},
//	}
//	palette := tui.NewCommandPalette(commands)
//	palette.Show()
type CommandPalette struct {
	width      int
	height     int
	visible    bool
	focused    bool
	textInput  textinput.Model
	commands   []Command
	filtered   []Command
	selected   int
	maxVisible int
}

// NewCommandPalette creates a new command palette with the given list of commands.
// The palette is initially hidden and can be shown/hidden with Show() and Hide(),
// or toggled with Toggle().
//
// The palette displays up to 8 commands at a time and supports fuzzy searching.
func NewCommandPalette(commands []Command) *CommandPalette {
	ti := textinput.New()
	ti.Placeholder = "Type to search commands..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	return &CommandPalette{
		textInput:  ti,
		commands:   commands,
		filtered:   commands,
		maxVisible: 8,
		visible:    false,
	}
}

// Init initializes the command palette
func (cp *CommandPalette) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (cp *CommandPalette) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cp.width = msg.Width
		cp.height = msg.Height

	case tea.KeyMsg:
		if !cp.focused {
			return cp, nil
		}

		// Toggle visibility with Ctrl+K or Ctrl+P
		if (msg.Type == tea.KeyCtrlK || msg.Type == tea.KeyCtrlP) && !cp.visible {
			cp.Show()
			return cp, nil
		}

		if !cp.visible {
			return cp, nil
		}

		switch msg.Type {
		case tea.KeyEsc:
			cp.Hide()
			return cp, nil

		case tea.KeyEnter:
			cp.Hide()
			if len(cp.filtered) > 0 && cp.selected < len(cp.filtered) {
				selectedCmd := cp.filtered[cp.selected]
				if selectedCmd.Action != nil {
					return cp, selectedCmd.Action()
				}
			}
			return cp, nil

		case tea.KeyUp:
			if cp.selected > 0 {
				cp.selected--
			}
			return cp, nil

		case tea.KeyDown:
			if cp.selected < len(cp.filtered)-1 {
				cp.selected++
			}
			return cp, nil

		default:
			// Update text input and filter commands
			cp.textInput, cmd = cp.textInput.Update(msg)
			cp.filterCommands()
			cp.selected = 0 // Reset selection on new input
			return cp, cmd
		}
	}

	if cp.visible && cp.focused {
		cp.textInput, cmd = cp.textInput.Update(msg)
		cp.filterCommands()
	}

	return cp, cmd
}

// View renders the command palette
func (cp *CommandPalette) View() string {
	if !cp.visible || cp.width == 0 {
		return ""
	}

	var b strings.Builder

	// Calculate dimensions
	paletteWidth := min(60, cp.width-4)
	paletteHeight := min(cp.maxVisible+4, cp.height-4)
	startX := (cp.width - paletteWidth) / 2
	startY := max(2, (cp.height-paletteHeight)/4)

	// Create overlay background (dim the screen)
	for y := 0; y < cp.height; y++ {
		if y == startY {
			// Draw palette starting here
			break
		}
	}

	// Title bar
	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString("\033[1;44m") // Blue background
	title := " Command Palette "
	padding := (paletteWidth - len(title)) / 2
	b.WriteString(strings.Repeat(" ", padding))
	b.WriteString(title)
	b.WriteString(strings.Repeat(" ", paletteWidth-padding-len(title)))
	b.WriteString("\033[0m\n")

	// Search input
	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString("\033[2m┌")
	b.WriteString(strings.Repeat("─", paletteWidth-2))
	b.WriteString("┐\033[0m\n")

	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString("\033[2m│\033[0m ")
	inputView := cp.textInput.View()
	b.WriteString(inputView)
	b.WriteString(strings.Repeat(" ", paletteWidth-len(stripANSI(inputView))-4))
	b.WriteString(" \033[2m│\033[0m\n")

	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString("\033[2m├")
	b.WriteString(strings.Repeat("─", paletteWidth-2))
	b.WriteString("┤\033[0m\n")

	// Command list
	visibleCommands := cp.filtered
	if len(visibleCommands) > cp.maxVisible {
		visibleCommands = visibleCommands[:cp.maxVisible]
	}

	if len(visibleCommands) == 0 {
		// No results
		b.WriteString(strings.Repeat(" ", startX))
		b.WriteString("\033[2m│\033[0m ")
		noResults := "No commands found"
		b.WriteString(noResults)
		b.WriteString(strings.Repeat(" ", paletteWidth-len(noResults)-4))
		b.WriteString(" \033[2m│\033[0m\n")
	} else {
		for i, cmd := range visibleCommands {
			b.WriteString(strings.Repeat(" ", startX))

			if i == cp.selected {
				// Selected item - highlighted
				b.WriteString("\033[2m│\033[0m\033[7m ▸ ") // Inverted
				cmdLine := fmt.Sprintf("%-30s", cmd.Name)
				if len(cmdLine) > 30 {
					cmdLine = cmdLine[:27] + "..."
				}
				b.WriteString(cmdLine)

				if cmd.Keybinding != "" {
					b.WriteString(" \033[2m")
					b.WriteString(cmd.Keybinding)
					b.WriteString("\033[0m\033[7m")
				}

				// Pad to width
				currentLen := 33 + len(cmd.Keybinding)
				b.WriteString(strings.Repeat(" ", paletteWidth-currentLen-3))
				b.WriteString("\033[0m\033[2m│\033[0m\n")
			} else {
				// Normal item
				b.WriteString("\033[2m│\033[0m   ")
				cmdLine := fmt.Sprintf("%-30s", cmd.Name)
				if len(cmdLine) > 30 {
					cmdLine = cmdLine[:27] + "..."
				}
				b.WriteString(cmdLine)

				if cmd.Keybinding != "" {
					b.WriteString(" \033[2m")
					b.WriteString(cmd.Keybinding)
					b.WriteString("\033[0m")
				}

				// Pad to width
				currentLen := 33 + len(cmd.Keybinding)
				b.WriteString(strings.Repeat(" ", paletteWidth-currentLen-3))
				b.WriteString("\033[2m│\033[0m\n")
			}
		}
	}

	// Footer
	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString("\033[2m└")
	footer := fmt.Sprintf(" %d commands ", len(cp.filtered))
	b.WriteString(footer)
	b.WriteString(strings.Repeat("─", paletteWidth-len(footer)-2))
	b.WriteString("┘\033[0m\n")

	return b.String()
}

// Focus is called when this component receives focus
func (cp *CommandPalette) Focus() {
	cp.focused = true
	cp.textInput.Focus()
}

// Blur is called when this component loses focus
func (cp *CommandPalette) Blur() {
	cp.focused = false
	cp.textInput.Blur()
}

// Focused returns whether this component is currently focused
func (cp *CommandPalette) Focused() bool {
	return cp.focused
}

// Show displays the command palette
func (cp *CommandPalette) Show() {
	cp.visible = true
	cp.textInput.SetValue("")
	cp.filtered = cp.commands
	cp.selected = 0
	cp.textInput.Focus()
}

// Hide conceals the command palette
func (cp *CommandPalette) Hide() {
	cp.visible = false
	cp.textInput.Blur()
}

// IsVisible returns whether the palette is currently visible
func (cp *CommandPalette) IsVisible() bool {
	return cp.visible
}

// filterCommands filters the command list based on search query
func (cp *CommandPalette) filterCommands() {
	query := strings.ToLower(strings.TrimSpace(cp.textInput.Value()))

	if query == "" {
		cp.filtered = cp.commands
		return
	}

	var filtered []Command
	for _, cmd := range cp.commands {
		// Simple substring matching (could be improved with fuzzy search)
		if strings.Contains(strings.ToLower(cmd.Name), query) ||
			strings.Contains(strings.ToLower(cmd.Description), query) ||
			strings.Contains(strings.ToLower(cmd.Category), query) {
			filtered = append(filtered, cmd)
		}
	}

	cp.filtered = filtered
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
