package input

import (
	"fmt"
	"sort"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

// CommandPaletteOption configures a CommandPalette.
type CommandPaletteOption func(*CommandPalette)

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

	// Design token colors
	surfaceColor string // overlay background
	borderColor  string // border around palette
	accentColor  string // input cursor, selected item highlight
	mutedColor   string // descriptions, category labels
	textColor    string // primary text color
	selectionBg  string // selected item background
}

// WithCommandPaletteDesignTokens applies design-system colors.
func WithCommandPaletteDesignTokens(tokens *design.DesignTokens) CommandPaletteOption {
	return func(cp *CommandPalette) {
		if tokens == nil {
			return
		}
		if v := strings.TrimSpace(tokens.SurfaceOverlay); v != "" {
			cp.surfaceColor = v
		}
		if v := strings.TrimSpace(tokens.BorderSubtle); v != "" {
			cp.borderColor = v
		}
		if v := strings.TrimSpace(tokens.Accent); v != "" {
			cp.accentColor = v
		}
		if v := strings.TrimSpace(tokens.MutedColor); v != "" {
			cp.mutedColor = v
		}
		if v := strings.TrimSpace(tokens.Color); v != "" {
			cp.textColor = v
		}
		if v := strings.TrimSpace(tokens.SurfaceRaised); v != "" {
			cp.selectionBg = v
		}
		cp.applyTextInputStyles()
	}
}

// NewCommandPalette creates a new command palette with the given list of commands.
func NewCommandPalette(commands []Command, opts ...CommandPaletteOption) *CommandPalette {
	ti := textinput.New()
	ti.Placeholder = "Type to search commands..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	cp := &CommandPalette{
		textInput:     ti,
		commands:      commands,
		filtered:      commands,
		maxVisible:    8,
		visible:       false,
		surfaceColor:  "#353A43",
		borderColor:   "#3C414B",
		accentColor:   "#61AFEF",
		mutedColor:    "#7A818A",
		textColor:     "#E5E7EB",
		selectionBg:   "#31353D",
	}

	cp.applyTextInputStyles()
	for _, opt := range opts {
		opt(cp)
	}

	return cp
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

	border := ansiOrFallback(style.ANSIColorFromHex(cp.borderColor), style.ANSIDim)
	surfaceBg := style.ANSIBackgroundColorFromHex(cp.surfaceColor)
	accent := ansiOrFallback(style.ANSIColorFromHex(cp.accentColor), style.ANSICyan)
	text := ansiOrFallback(style.ANSIColorFromHex(cp.textColor), style.ANSIReset)
	muted := ansiOrFallback(style.ANSIColorFromHex(cp.mutedColor), style.ANSIDim)
	selectionBg := ansiOrFallback(style.ANSIBackgroundColorFromHex(cp.selectionBg), style.ANSIInverse)

	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString(surfaceBg)
	title := " Command Palette "
	padding := (paletteWidth - style.StringWidth(title)) / 2
	if padding < 0 {
		padding = 0
	}
	b.WriteString(strings.Repeat(" ", padding))
	b.WriteString(accent + title + style.ANSIReset + surfaceBg)
	rightPadding := paletteWidth - padding - style.StringWidth(title)
	if rightPadding < 0 {
		rightPadding = 0
	}
	b.WriteString(strings.Repeat(" ", rightPadding))
	b.WriteString(style.ANSIReset + "\n")

	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString(border + "┌")
	b.WriteString(strings.Repeat("─", paletteWidth-2))
	b.WriteString("┐" + style.ANSIReset + "\n")

	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString(border + "│" + style.ANSIReset + surfaceBg + " ")
	inputView := cp.textInput.View()
	b.WriteString(style.Pad(inputView, paletteWidth-4))
	b.WriteString(" " + style.ANSIReset + border + "│" + style.ANSIReset + "\n")

	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString(border + "├")
	b.WriteString(strings.Repeat("─", paletteWidth-2))
	b.WriteString("┤" + style.ANSIReset + "\n")

	visibleCommands := cp.filtered
	if len(visibleCommands) > cp.maxVisible {
		visibleCommands = visibleCommands[:cp.maxVisible]
	}

	if len(visibleCommands) == 0 {
		b.WriteString(strings.Repeat(" ", startX))
		b.WriteString(border + "│" + style.ANSIReset + surfaceBg + " ")
		b.WriteString(muted + style.Pad("No commands found", paletteWidth-4) + style.ANSIReset)
		b.WriteString(surfaceBg + " " + style.ANSIReset + border + "│" + style.ANSIReset + "\n")
	} else {
		for i, cmd := range visibleCommands {
			b.WriteString(strings.Repeat(" ", startX))
			b.WriteString(border + "│" + style.ANSIReset)

			name := style.Pad(style.TruncateMiddle(cmd.Name, 30, "…"), 30)
			currentLen := 33 + style.StringWidth(cmd.Keybinding)
			linePadding := paletteWidth - currentLen - 3
			if linePadding < 0 {
				linePadding = 0
			}

			if i == cp.selected {
				b.WriteString(selectionBg + " ")
				b.WriteString(accent + "▸" + style.ANSIReset + selectionBg + " ")
				b.WriteString(text + name + style.ANSIReset + selectionBg)
				if cmd.Keybinding != "" {
					b.WriteString(" " + muted + cmd.Keybinding + style.ANSIReset + selectionBg)
				}
				b.WriteString(strings.Repeat(" ", linePadding))
				b.WriteString(" " + style.ANSIReset)
			} else {
				b.WriteString(surfaceBg + "   ")
				b.WriteString(text + name + style.ANSIReset + surfaceBg)
				if cmd.Keybinding != "" {
					b.WriteString(" " + muted + cmd.Keybinding + style.ANSIReset + surfaceBg)
				}
				b.WriteString(strings.Repeat(" ", linePadding))
				b.WriteString(" " + style.ANSIReset)
			}

			b.WriteString(border + "│" + style.ANSIReset + "\n")
		}
	}

	b.WriteString(strings.Repeat(" ", startX))
	b.WriteString(border + "└")
	footer := fmt.Sprintf(" %d commands ", len(cp.filtered))
	b.WriteString(muted + footer + style.ANSIReset + border)
	footerWidth := paletteWidth - style.StringWidth(footer) - 2
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

func (cp *CommandPalette) applyTextInputStyles() {
	if strings.TrimSpace(cp.accentColor) != "" {
		cp.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(cp.accentColor))
		cp.textInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(cp.accentColor))
	}
	if strings.TrimSpace(cp.textColor) != "" {
		cp.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(cp.textColor))
	}
	if strings.TrimSpace(cp.mutedColor) != "" {
		cp.textInput.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(cp.mutedColor))
	}
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

func ansiOrFallback(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
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
