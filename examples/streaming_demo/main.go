package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

// Simulated command output for demonstrations
var testOutput = []string{
	"=== RUN   TestApplicationCreation",
	"--- PASS: TestApplicationCreation (0.00s)",
	"=== RUN   TestComponentAddition",
	"--- PASS: TestComponentAddition (0.00s)",
	"=== RUN   TestFocusManagement",
	"--- PASS: TestFocusManagement (0.01s)",
	"=== RUN   TestWindowSizeMsg",
	"--- PASS: TestWindowSizeMsg (0.00s)",
	"=== RUN   TestQuitKeys",
	"--- PASS: TestQuitKeys (0.00s)",
	"=== RUN   TestStatusBarMessage",
	"--- PASS: TestStatusBarMessage (0.00s)",
	"=== RUN   TestStatusBarView",
	"--- PASS: TestStatusBarView (0.00s)",
	"PASS",
	"ok  \tgithub.com/SCKelemen/tui\t0.536s",
}

var buildOutput = []string{
	"Building project...",
	"Compiling main.go",
	"Compiling components.go",
	"Compiling utils.go",
	"Linking...",
	"Build complete!",
}

var errorOutput = []string{
	"Running tests...",
	"=== RUN   TestInvalidInput",
	"    main_test.go:42: Expected nil, got error",
	"--- FAIL: TestInvalidInput (0.02s)",
	"FAIL",
	"exit status 1",
}

type model struct {
	app           *tui.Application
	activityBar   *tui.ActivityBar
	testBlock     *tui.ToolBlock
	buildBlock    *tui.ToolBlock
	errorBlock    *tui.ToolBlock
	staticBlock   *tui.ToolBlock
	stage         int
	lineIndex     int
	currentBlock  *tui.ToolBlock
	currentOutput []string
}

type streamLineMsg struct{}
type nextStageMsg struct{}

func newModel() model {
	app := tui.NewApplication()

	// Activity bar
	activityBar := tui.NewActivityBar()
	app.AddComponent(activityBar)

	// Test streaming block (will stream test output)
	testBlock := tui.NewToolBlock(
		"Bash",
		"go test -v",
		[]string{},
		tui.WithMaxLines(5),
		tui.WithStreaming(),
	)
	app.AddComponent(testBlock)

	// Build streaming block
	buildBlock := tui.NewToolBlock(
		"Bash",
		"go build .",
		[]string{},
		tui.WithMaxLines(3),
		tui.WithStreaming(),
	)
	app.AddComponent(buildBlock)

	// Error streaming block
	errorBlock := tui.NewToolBlock(
		"Bash",
		"go test ./broken",
		[]string{},
		tui.WithMaxLines(4),
		tui.WithStreaming(),
	)
	app.AddComponent(errorBlock)

	// Static completed block for comparison
	staticBlock := tui.NewToolBlock(
		"Bash",
		"echo 'Static output'",
		[]string{
			"Static output",
			"This block was complete from the start",
			"No streaming here!",
		},
		tui.WithMaxLines(2),
		tui.WithStatus(tui.StatusComplete),
	)
	app.AddComponent(staticBlock)

	return model{
		app:           app,
		activityBar:   activityBar,
		testBlock:     testBlock,
		buildBlock:    buildBlock,
		errorBlock:    errorBlock,
		staticBlock:   staticBlock,
		stage:         0,
		currentBlock:  testBlock,
		currentOutput: testOutput,
	}
}

func (m model) Init() tea.Cmd {
	m.activityBar.Start("Running tests...")
	return tea.Batch(
		m.app.Init(),
		streamLine(),
	)
}

func streamLine() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return streamLineMsg{}
	})
}

func nextStage() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return nextStageMsg{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			// Reset demonstration
			newM := newModel()
			return newM, newM.Init()
		}

	case streamLineMsg:
		if m.lineIndex < len(m.currentOutput) {
			// Add next line to current block
			m.currentBlock.AppendLine(m.currentOutput[m.lineIndex])
			m.lineIndex++
			return m, streamLine()
		} else {
			// Finished streaming current block
			switch m.stage {
			case 0: // Test block done
				m.testBlock.SetStatus(tui.StatusComplete)
				m.activityBar.SetProgress("✓ Tests passed")
				return m, nextStage()
			case 1: // Build block done
				m.buildBlock.SetStatus(tui.StatusComplete)
				m.activityBar.SetProgress("✓ Build complete")
				return m, nextStage()
			case 2: // Error block done
				m.errorBlock.SetStatus(tui.StatusError)
				m.activityBar.Stop()
				return m, nil
			}
		}

	case nextStageMsg:
		m.stage++
		m.lineIndex = 0

		switch m.stage {
		case 1: // Start build
			m.activityBar.Start("Building project...")
			m.currentBlock = m.buildBlock
			m.currentOutput = buildOutput
			return m, streamLine()
		case 2: // Start error demo
			m.activityBar.Start("Running failing tests...")
			m.currentBlock = m.errorBlock
			m.currentOutput = errorOutput
			return m, streamLine()
		}
	}

	// Pass to app
	appModel, cmd := m.app.Update(msg)
	m.app = appModel.(*tui.Application)
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("\033[1m=== Streaming ToolBlock Demo ===\033[0m\n\n")

	b.WriteString(m.app.View())

	b.WriteString("\n\033[2m")
	b.WriteString("This demo shows streaming output in real-time:\n")
	b.WriteString("  • Test output streams line-by-line\n")
	b.WriteString("  • Build output follows after tests complete\n")
	b.WriteString("  • Error states are shown with ✗ indicators\n")
	b.WriteString("  • Complete states are shown with ✓ indicators\n")
	b.WriteString("  • Spinners animate while streaming\n")
	b.WriteString("\n")
	b.WriteString("Press 'r' to restart demo · 'q' to quit\n")
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
