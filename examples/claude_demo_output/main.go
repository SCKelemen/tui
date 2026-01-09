package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

func main() {
	fmt.Println("=== Claude Code-Style TUI Components Demo ===\n")

	// Demo 1: Activity Bar
	fmt.Println("1. Animated Activity Bar:")
	fmt.Println("   (Shows spinner, elapsed time, and progress)")
	fmt.Println()

	activityBar := tui.NewActivityBar()
	activityBar.Update(tea.WindowSizeMsg{Width: 80})

	// Start activity
	activityBar.Start("Actualizing…")

	// Simulate some time passing
	for i := 0; i < 5; i++ {
		activityBar.Update(tickMsg(time.Now()))
		activityBar.SetProgress(fmt.Sprintf("↓ %.1fk tokens", float64(i)*0.5))
		fmt.Print(activityBar.View())
		time.Sleep(200 * time.Millisecond)
	}

	activityBar.Stop()
	fmt.Println(activityBar.View())
	fmt.Println()

	// Demo 2: Tool Blocks - Collapsed
	fmt.Println("2. Collapsible Tool Blocks (Collapsed):")
	fmt.Println()

	bashBlock := tui.NewToolBlock(
		"Bash",
		"cd ~/Code/github.com/SCKelemen/tui/examples && go run demo_output.go",
		[]string{
			"=== TUI Framework Demo ===",
			"",
			"Initial state (StatusBar1 focused):",
			"Status Bar 1 - Press Tab to focus next",
			"",
			"After pressing Tab (StatusBar2 focused):",
			"Status Bar 2 - Press Shift+Tab to focus previous",
			"",
			"✓ TUI framework is working correctly!",
			"✓ Focus management: Tab/Shift+Tab navigation",
			"✓ Component lifecycle: Init, Update, View, Focus, Blur",
			"✓ Keyboard handling: q, Ctrl+C to quit",
			"",
			"Run './examples/basic/basic' for interactive demo",
		},
		tui.WithMaxLines(3),
	)
	bashBlock.Update(tea.WindowSizeMsg{Width: 80})
	fmt.Print(bashBlock.View())
	fmt.Println()

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
	testBlock.Update(tea.WindowSizeMsg{Width: 80})
	fmt.Print(testBlock.View())
	fmt.Println()

	// Demo 3: Tool Block with Line Numbers
	fmt.Println("3. Tool Block with Line Numbers (Code File):")
	fmt.Println()

	writeBlock := tui.NewToolBlock(
		"Write",
		"examples/demo_output.go",
		[]string{
			"package main",
			"",
			"import (",
			"    \"fmt\"",
			"",
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
	writeBlock.Update(tea.WindowSizeMsg{Width: 80})
	fmt.Print(writeBlock.View())
	fmt.Println()

	// Demo 4: Expanded Tool Block
	fmt.Println("4. Expanded Tool Block:")
	fmt.Println()

	expandedBlock := tui.NewToolBlock(
		"Read",
		"tui.go",
		[]string{
			"package tui",
			"",
			"type Application struct {",
			"    width      int",
			"    height     int",
			"    components []Component",
			"}",
		},
		tui.WithLineNumbers(),
	)
	expandedBlock.SetExpanded(true)
	expandedBlock.Update(tea.WindowSizeMsg{Width: 80})
	fmt.Print(expandedBlock.View())
	fmt.Println()

	fmt.Println("✓ All Claude Code-style components rendering correctly!")
	fmt.Println()
	fmt.Println("Interactive features in the full TUI:")
	fmt.Println("  • Tab/Shift+Tab: Navigate between components")
	fmt.Println("  • Ctrl+O or Enter: Expand/collapse tool blocks")
	fmt.Println("  • Esc: Interrupt active operations (when activity bar is running)")
	fmt.Println()
	fmt.Println("Run 'go run examples/claude_code_demo/main.go' for interactive demo")
}

type tickMsg time.Time
