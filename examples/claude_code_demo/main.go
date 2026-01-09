package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

// demoModel wraps the TUI app with custom update logic
type demoModel struct {
	app         *tui.Application
	activityBar *tui.ActivityBar
	stage       int
	startTime   time.Time
}

func newDemoModel() demoModel {
	app := tui.NewApplication()

	// Activity bar at the top
	activityBar := tui.NewActivityBar()
	app.AddComponent(activityBar)

	// Tool blocks showing various operations
	bashBlock := tui.NewToolBlock(
		"Bash",
		"cd ~/Code/github.com/SCKelemen/tui/examples && go run demo_output.go",
		[]string{
			"=== TUI Framework Demo ===",
			"",
			"Initial state (StatusBar1 focused):",
			"Status Bar 1 - Press Tab to focus next                    Tab: Focus • q: Quit",
			"",
			"After pressing Tab (StatusBar2 focused):",
			"Status Bar 2 - Press Shift+Tab to focus previous          Tab: Focus • q: Quit",
			"",
			"After pressing Shift+Tab (StatusBar1 focused again):",
			"Status Bar 1 - Press Tab to focus next                    Tab: Focus • q: Quit",
			"",
			"",
			"✓ TUI framework is working correctly!",
			"✓ Focus management: Tab/Shift+Tab navigation",
			"✓ Component lifecycle: Init, Update, View, Focus, Blur",
			"✓ Keyboard handling: q, Ctrl+C to quit",
		},
		tui.WithMaxLines(3),
	)
	app.AddComponent(bashBlock)

	testBlock := tui.NewToolBlock(
		"Bash",
		"cd ~/Code/github.com/SCKelemen/tui && go test -v",
		[]string{
			"=== RUN   TestApplicationCreation",
			"--- PASS: TestApplicationCreation (0.00s)",
			"=== RUN   TestComponentAddition",
			"--- PASS: TestComponentAddition (0.00s)",
			"=== RUN   TestFocusManagement",
			"--- PASS: TestFocusManagement (0.00s)",
			"=== RUN   TestWindowSizeMsg",
			"--- PASS: TestWindowSizeMsg (0.00s)",
			"=== RUN   TestQuitKeys",
			"--- PASS: TestQuitKeys (0.00s)",
			"=== RUN   TestStatusBarMessage",
			"--- PASS: TestStatusBarMessage (0.00s)",
			"=== RUN   TestStatusBarView",
			"--- PASS: TestStatusBarView (0.00s)",
			"PASS",
			"ok  	github.com/SCKelemen/tui	0.768s",
		},
		tui.WithMaxLines(4),
	)
	app.AddComponent(testBlock)

	writeBlock := tui.NewToolBlock(
		"Write",
		"examples/demo_output.go",
		[]string{
			"package main",
			"",
			"import (",
			"    \"fmt\"",
			"    tea \"github.com/charmbracelet/bubbletea\"",
			"    \"github.com/SCKelemen/tui\"",
			")",
			"",
			"func main() {",
			"    fmt.Println(\"=== TUI Framework Demo ===\")",
			"    app := tui.NewApplication()",
			"    statusBar := tui.NewStatusBar()",
			"    app.AddComponent(statusBar)",
			"    app.Run()",
			"}",
		},
		tui.WithLineNumbers(),
		tui.WithMaxLines(7),
	)
	app.AddComponent(writeBlock)

	return demoModel{
		app:         app,
		activityBar: activityBar,
		stage:       0,
		startTime:   time.Now(),
	}
}

func (m demoModel) Init() tea.Cmd {
	return tea.Batch(
		m.app.Init(),
		m.activityBar.Start("Actualizing…"),
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type tickMsg time.Time

func (m demoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case tickMsg:
		elapsed := time.Since(m.startTime)

		// Update progress indicator
		tokens := int(elapsed.Seconds() * 0.5 * 1000) // ~500 tokens/sec
		m.activityBar.SetProgress(fmt.Sprintf("↓ %.1fk tokens", float64(tokens)/1000))

		return m, tickCmd()
	}

	// Pass to app
	appModel, cmd := m.app.Update(msg)
	m.app = appModel.(*tui.Application)
	return m, cmd
}

func (m demoModel) View() string {
	return m.app.View()
}

func main() {
	p := tea.NewProgram(newDemoModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
