package input

import (
	"fmt"
	"sort"
	"strings"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Command represents an executable command in the command palette with metadata
// for display and categorization.
type Command struct {
	Name        string         // Display name of the command
	Description string         // Brief description of what the command does
	Category    string         // Category for grouping (e.g., "File", "Edit", "View")
	Action      func() tea.Cmd // Function to execute when command is selected
	Keybinding  string         // Optional keyboard shortcut (e.g., "Ctrl+S")
}

// CommandPalette is a searchable command launcher inspired by VS Code's command
// palette. It provides a popup interface for quickly executing commands via keyboard.
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

// Init initializes the command palette.
func (cp *CommandPalette) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages.
func (cp *CommandPalette) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cp.width = msg.Width
		cp.height = msg.Height

	case tea.KeyMsg:
		if !cp.focused {
			return cp, nil
		}

		// Toggle visibility with Ctrl+K or Ctrl+P.
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
			cp.textInput, cmd = cp.textInput.Update(msg)
			cp.filterCommands()
			cp.selected = 0
			return cp, cmd
		}
	}

	if cp.visible && cp.focused {
		cp.textInput, cmd = cp.textInput.Update(msg)
		cp.filterCommands()
	}

	return cp, cmd
}

// View renders the command palette.
func (cp *CommandPalette) View() string {
	if !cp.visible || cp.width == 0 {
		return ""
	}

	var b strings.Builder

	paletteWidth := min(cp.width-4, 60)
	if paletteWidth < 20 {
		paletteWidth = 20
	}
	paletteHeight := min(cp.maxVisible+4, cp.height-4)
	startX := max(0, (cp.width-paletteWidth)/2)
	startY := max(2, (cp.height-paletteHeight)/4)

	for y := 0; y < cp.height; y++ {
		if y == startY {
			break
		}
	}

	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString("\033[1;44m")
	title := " Command Palette "
	padding := (paletteWidth - len(title)) / 2
	b.WriteString(strings.Repeat(" ", padding))
	b.WriteString(title)
	rightPadding := paletteWidth - padding - len(title)
	if rightPadding < 0 {
		rightPadding = 0
	}
	b.WriteString(strings.Repeat(" ", rightPadding))
	b.WriteString(style.ANSIReset + "\n")

	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString(style.ANSIDim + "┌")
	b.WriteString(strings.Repeat("─", paletteWidth-2))
	b.WriteString("┐" + style.ANSIReset + "\n")

	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString(style.ANSIDim + "│" + style.ANSIReset + " ")
	inputView := cp.textInput.View()
	b.WriteString(inputView)
	inputPadding := paletteWidth - len(stripANSICommandPalette(inputView)) - 4
	if inputPadding < 0 {
		inputPadding = 0
	}
	b.WriteString(strings.Repeat(" ", inputPadding))
	b.WriteString(" " + style.ANSIDim + "│" + style.ANSIReset + "\n")

	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString(style.ANSIDim + "├")
	b.WriteString(strings.Repeat("─", paletteWidth-2))
	b.WriteString("┤" + style.ANSIReset + "\n")

	visibleCommands := cp.filtered
	if len(visibleCommands) > cp.maxVisible {
		visibleCommands = visibleCommands[:cp.maxVisible]
	}

	if len(visibleCommands) == 0 {
		b.WriteString(strings.Repeat(" ", startX))
		b.WriteString(style.ANSIDim + "│" + style.ANSIReset + " ")
		noResults := "No commands found"
		b.WriteString(noResults)
		noResultsPadding := paletteWidth - len(noResults) - 4
		if noResultsPadding < 0 {
			noResultsPadding = 0
		}
		b.WriteString(strings.Repeat(" ", noResultsPadding))
		b.WriteString(" " + style.ANSIDim + "│" + style.ANSIReset + "\n")
	} else {
		for i, cmd := range visibleCommands {
			b.WriteString(strings.Repeat(" ", startX))

			if i == cp.selected {
				b.WriteString(style.ANSIDim + "│" + style.ANSIReset + style.ANSIInverse + " ▸ ")
				cmdLine := fmt.Sprintf("%-30s", cmd.Name)
				if len(cmdLine) > 30 {
					cmdLine = cmdLine[:27] + "..."
				}
				b.WriteString(cmdLine)

				if cmd.Keybinding != "" {
					b.WriteString(" " + style.ANSIDim)
					b.WriteString(cmd.Keybinding)
					b.WriteString(style.ANSIReset + style.ANSIInverse)
				}

				currentLen := 33 + len(cmd.Keybinding)
				linePadding := paletteWidth - currentLen - 3
				if linePadding < 0 {
					linePadding = 0
				}
				b.WriteString(strings.Repeat(" ", linePadding))
				b.WriteString(style.ANSIReset + style.ANSIDim + "│" + style.ANSIReset + "\n")
			} else {
				b.WriteString(style.ANSIDim + "│" + style.ANSIReset + "   ")
				cmdLine := fmt.Sprintf("%-30s", cmd.Name)
				if len(cmdLine) > 30 {
					cmdLine = cmdLine[:27] + "..."
				}
				b.WriteString(cmdLine)

				if cmd.Keybinding != "" {
					b.WriteString(" " + style.ANSIDim)
					b.WriteString(cmd.Keybinding)
					b.WriteString(style.ANSIReset)
				}

				currentLen := 33 + len(cmd.Keybinding)
				linePadding := paletteWidth - currentLen - 3
				if linePadding < 0 {
					linePadding = 0
				}
				b.WriteString(strings.Repeat(" ", linePadding))
				b.WriteString(style.ANSIDim + "│" + style.ANSIReset + "\n")
			}
		}
	}

	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString(style.ANSIDim + "└")
	footer := fmt.Sprintf(" %d commands ", len(cp.filtered))
	b.WriteString(footer)
	footerWidth := paletteWidth - len(footer) - 2
	if footerWidth < 0 {
		footerWidth = 0
	}
	b.WriteString(strings.Repeat("─", footerWidth))
	b.WriteString("┘" + style.ANSIReset + "\n")

	return b.String()
}

// Focus is called when this component receives focus.
func (cp *CommandPalette) Focus() {
	cp.focused = true
	cp.textInput.Focus()
}

// Blur is called when this component loses focus.
func (cp *CommandPalette) Blur() {
	cp.focused = false
	cp.textInput.Blur()
}

// Focused returns whether this component is currently focused.
func (cp *CommandPalette) Focused() bool {
	return cp.focused
}

// Show displays the command palette.
func (cp *CommandPalette) Show() {
	cp.visible = true
	cp.textInput.SetValue("")
	cp.filtered = cp.commands
	cp.selected = 0
	cp.textInput.Focus()
}

// Hide conceals the command palette.
func (cp *CommandPalette) Hide() {
	cp.visible = false
	cp.textInput.Blur()
}

// IsVisible returns whether the palette is currently visible.
func (cp *CommandPalette) IsVisible() bool {
	return cp.visible
}

// filterCommands filters the command list based on search query.
func (cp *CommandPalette) filterCommands() {
	query := strings.ToLower(strings.TrimSpace(cp.textInput.Value()))

	if query == "" {
		cp.filtered = cp.commands
		return
	}

	type scoredCommand struct {
		command Command
		score   int
	}

	scored := make([]scoredCommand, 0, len(cp.commands))
	for _, cmd := range cp.commands {
		score, ok := commandMatchScore(query, cmd)
		if !ok {
			continue
		}
		scored = append(scored, scoredCommand{command: cmd, score: score})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return strings.ToLower(scored[i].command.Name) < strings.ToLower(scored[j].command.Name)
		}
		return scored[i].score > scored[j].score
	})

	cp.filtered = make([]Command, 0, len(scored))
	for _, item := range scored {
		cp.filtered = append(cp.filtered, item.command)
	}
}

func commandMatchScore(query string, cmd Command) (int, bool) {
	name := strings.ToLower(cmd.Name)
	desc := strings.ToLower(cmd.Description)
	cat := strings.ToLower(cmd.Category)

	best := -1

	if name == query {
		best = max(best, 300)
	} else if strings.HasPrefix(name, query) {
		best = max(best, 250)
	} else if strings.Contains(name, query) {
		best = max(best, 200)
	}

	if desc == query {
		best = max(best, 160)
	} else if strings.HasPrefix(desc, query) {
		best = max(best, 130)
	} else if strings.Contains(desc, query) {
		best = max(best, 100)
	}

	if cat == query {
		best = max(best, 90)
	} else if strings.HasPrefix(cat, query) {
		best = max(best, 70)
	} else if strings.Contains(cat, query) {
		best = max(best, 50)
	}

	if best >= 0 {
		return best, true
	}

	if subsequenceMatch(name, query) {
		return 40, true
	}
	if subsequenceMatch(desc, query) {
		return 20, true
	}
	if subsequenceMatch(cat, query) {
		return 10, true
	}

	return 0, false
}

func subsequenceMatch(candidate, query string) bool {
	if query == "" {
		return true
	}
	qi := 0
	for _, r := range candidate {
		if qi < len(query) && byte(r) == query[qi] {
			qi++
		}
		if qi == len(query) {
			return true
		}
	}
	return false
}

func stripANSICommandPalette(s string) string {
	var result strings.Builder
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
		} else if inEscape {
			if r == 'm' {
				inEscape = false
			}
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// Helper functions.
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

var _ tui.Component = (*CommandPalette)(nil)
