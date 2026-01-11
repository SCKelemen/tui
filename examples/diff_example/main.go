package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

type model struct {
	diffBlock *tui.DiffBlock
	width     int
	height    int
}

func initialModel() model {
	// Show the actual diff from renaming WithOperation -> WithCodeOperation
	oldCode := `func initialModel() model {
	// Create a code block showing a file write
	codeBlock := tui.NewCodeBlock(
		tui.WithOperation("Write"),
		tui.WithFilename("examples/selection_range_example.go"),
		tui.WithSummary("Wrote 253 lines to examples/selection_range_example.go"),
		tui.WithCode(` + "`package examples`" + `)`

	newCode := `func initialModel() model {
	// Create a code block showing a file write
	codeBlock := tui.NewCodeBlock(
		tui.WithCodeOperation("Write"),
		tui.WithCodeFilename("examples/selection_range_example.go"),
		tui.WithCodeSummary("Wrote 253 lines to examples/selection_range_example.go"),
		tui.WithCode(` + "`package examples`" + `)`

	diffBlock := tui.NewDiffBlockFromStrings(oldCode, newCode,
		tui.WithDiffFilename("examples/code_demo/main.go"),
		tui.WithDiffOperation("Update"),
		tui.WithDiffExpanded(true), // Start expanded to show the full diff
	)

	return model{
		diffBlock: diffBlock,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "ctrl+o", "enter", " ":
			m.diffBlock.Toggle()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.diffBlock.Update(msg)
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	var s string
	s += "\033[1m╭─ Claude Code-Style Diff Display ────────────────────────────────╮\033[0m\n"
	s += "\033[1m│\033[0m Press \033[33mctrl+o\033[0m or \033[33mspace\033[0m to toggle, \033[33mq\033[0m to quit                   \033[1m│\033[0m\n"
	s += "\033[1m╰──────────────────────────────────────────────────────────────────╯\033[0m\n\n"

	s += m.diffBlock.View()

	s += "\n\033[2m────────────────────────────────────────────────────────────────────\033[0m\n"
	if m.diffBlock.IsExpanded() {
		s += "\033[2mDiff is expanded. Press \033[33mctrl+o\033[0m to collapse\033[0m\n"
	} else {
		s += "\033[2mDiff is collapsed. Press \033[33mctrl+o\033[0m to expand\033[0m\n"
	}

	return s
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
