package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

type model struct {
	codeBlock *tui.CodeBlock
	diffBlock *tui.DiffBlock
	width     int
	height    int
}

func initialModel() model {
	// Create a code block showing a file write
	codeBlock := tui.NewCodeBlock(
		tui.WithCodeOperation("Write"),
		tui.WithCodeFilename("examples/selection_range_example.go"),
		tui.WithCodeSummary("Wrote 253 lines to examples/selection_range_example.go"),
		tui.WithCode(`package examples

import (
    "go/ast"
    "go/parser"
    "go/token"
    "strings"
    "github.com/SCKelemen/lsp/core"
)

func SelectionRange(filename string, line, char int) (*core.SelectionRange, error) {
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
    if err != nil {
        return nil, err
    }

    // Find the position in the file
    pos := fset.Position(token.Pos(line*1000 + char))

    // ... rest of implementation
    return &core.SelectionRange{}, nil
}`),
		tui.WithStartLine(1),
		tui.WithExpanded(false), // Start collapsed
	)

	// Create a diff block showing code changes
	oldCode := `func processData(data string) error {
    if data == "" {
        return errors.New("empty data")
    }

    result := strings.ToUpper(data)
    fmt.Println(result)
    return nil
}`

	newCode := `func processData(data string) error {
    if data == "" {
        return fmt.Errorf("empty data")
    }

    // Normalize the data
    result := strings.ToLower(data)
    log.Printf("Processed: %s", result)
    return nil
}`

	diffBlock := tui.NewDiffBlockFromStrings(oldCode, newCode,
		tui.WithDiffFilename("processor.go"),
		tui.WithDiffOperation("Edit"),
		tui.WithDiffSummary("Updated error handling and changed normalization"),
		tui.WithDiffExpanded(false),
	)

	return model{
		codeBlock: codeBlock,
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
		case "1":
			m.codeBlock.Toggle()
		case "2":
			m.diffBlock.Toggle()
		case "ctrl+o":
			// Toggle both
			m.codeBlock.Toggle()
			m.diffBlock.Toggle()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.codeBlock.Update(msg)
		m.diffBlock.Update(msg)
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	var s strings.Builder

	// Header
	s.WriteString("\033[1m╭─ Code Block & Diff Demo ─────────────────────────────────────────╮\033[0m\n")
	s.WriteString("\033[1m│\033[0m Press \033[33m1\033[0m to toggle code block, \033[33m2\033[0m to toggle diff, \033[33mCtrl+O\033[0m both    \033[1m│\033[0m\n")
	s.WriteString("\033[1m│\033[0m Press \033[33mq\033[0m to quit                                              \033[1m│\033[0m\n")
	s.WriteString("\033[1m╰───────────────────────────────────────────────────────────────────╯\033[0m\n\n")

	// Code block
	s.WriteString("\033[1m📝 Source Code Block:\033[0m\n")
	s.WriteString(m.codeBlock.View())
	s.WriteString("\n")

	// Diff block
	s.WriteString("\033[1m🔄 Diff Block:\033[0m\n")
	s.WriteString(m.diffBlock.View())
	s.WriteString("\n")

	// Footer hints
	s.WriteString("\033[2m" + strings.Repeat("─", m.width-1) + "\033[0m\n")
	if !m.codeBlock.IsExpanded() {
		s.WriteString("\033[2mCode block is collapsed. Press \033[33m1\033[0m or \033[33mCtrl+O\033[0m to expand\033[0m\n")
	} else {
		s.WriteString("\033[2mCode block is expanded. Press \033[33m1\033[0m to collapse\033[0m\n")
	}
	if !m.diffBlock.IsExpanded() {
		s.WriteString("\033[2mDiff block is collapsed. Press \033[33m2\033[0m or \033[33mCtrl+O\033[0m to expand\033[0m\n")
	} else {
		s.WriteString("\033[2mDiff block is expanded. Press \033[33m2\033[0m to collapse\033[0m\n")
	}

	return s.String()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
