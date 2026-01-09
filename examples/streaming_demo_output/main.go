package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

func main() {
	fmt.Println("=== Streaming ToolBlock Components Demo ===\n")

	// 1. Running/Streaming State
	fmt.Println("1. ToolBlock in Streaming Mode (Running):\n")

	runningBlock := tui.NewToolBlock(
		"Bash",
		"go test -v ./...",
		[]string{},
		tui.WithStreaming(),
	)
	runningBlock.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	fmt.Println(runningBlock.View())

	// 2. Streaming with Partial Output
	fmt.Println("\n2. ToolBlock Streaming with Partial Output:\n")

	streamingBlock := tui.NewToolBlock(
		"Bash",
		"npm install",
		[]string{
			"npm WARN deprecated pkg@1.0.0",
			"added 42 packages in 2.1s",
			"downloading dependencies...",
		},
		tui.WithMaxLines(2),
		tui.WithStreaming(),
	)
	streamingBlock.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	fmt.Println(streamingBlock.View())

	// 3. Complete Status (Success)
	fmt.Println("\n3. ToolBlock Complete (Success ✓):\n")

	completeBlock := tui.NewToolBlock(
		"Bash",
		"go test -v",
		[]string{
			"=== RUN   TestApplicationCreation",
			"--- PASS: TestApplicationCreation (0.00s)",
			"=== RUN   TestComponentAddition",
			"--- PASS: TestComponentAddition (0.00s)",
			"=== RUN   TestFocusManagement",
			"--- PASS: TestFocusManagement (0.01s)",
			"PASS",
			"ok  \tgithub.com/SCKelemen/tui\t0.536s",
		},
		tui.WithMaxLines(4),
		tui.WithStatus(tui.StatusComplete),
	)
	completeBlock.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	fmt.Println(completeBlock.View())

	// 4. Error Status
	fmt.Println("\n4. ToolBlock with Error (✗):\n")

	errorBlock := tui.NewToolBlock(
		"Bash",
		"go test ./broken",
		[]string{
			"=== RUN   TestInvalidInput",
			"    main_test.go:42: Expected nil, got error: invalid input",
			"--- FAIL: TestInvalidInput (0.02s)",
			"FAIL",
			"exit status 1",
		},
		tui.WithMaxLines(3),
		tui.WithStatus(tui.StatusError),
	)
	errorBlock.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	fmt.Println(errorBlock.View())

	// 5. Warning Status
	fmt.Println("\n5. ToolBlock with Warning (⚠):\n")

	warningBlock := tui.NewToolBlock(
		"Bash",
		"npm audit",
		[]string{
			"found 3 vulnerabilities (2 moderate, 1 high)",
			"run `npm audit fix` to fix them",
			"",
			"Some issues need review",
		},
		tui.WithMaxLines(3),
		tui.WithStatus(tui.StatusWarning),
	)
	warningBlock.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	fmt.Println(warningBlock.View())

	// 6. API Methods Demo
	fmt.Println("\n6. API Methods for Streaming:\n")

	apiBlock := tui.NewToolBlock(
		"Bash",
		"curl https://api.example.com",
		[]string{},
		tui.WithStreaming(),
	)
	apiBlock.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Simulate streaming by appending lines
	apiBlock.AppendLine("Connecting to api.example.com...")
	apiBlock.AppendLine("Sending request...")
	apiBlock.AppendLines([]string{
		"Received response: 200 OK",
		"{\"status\": \"success\"}",
	})

	fmt.Println(apiBlock.View())

	// Feature Summary
	fmt.Println("\n✓ All streaming features working!\n")
	fmt.Println("Features:")
	fmt.Println("  • Real-time output streaming with AppendLine() and AppendLines()")
	fmt.Println("  • Status indicators: ✓ (success), ✗ (error), ⚠ (warning)")
	fmt.Println("  • Animated spinners while streaming")
	fmt.Println("  • Color-coded headers by status (green, red, yellow, cyan)")
	fmt.Println("  • StartStreaming() and StopStreaming() methods")
	fmt.Println("  • SetStatus() to change execution state")
	fmt.Println("  • WithStreaming() option to enable at creation")
	fmt.Println("\nRun 'go run examples/streaming_demo/main.go' for interactive demo")

	// 7. Multiple Blocks in Different States
	fmt.Println("\n\n=== Multiple ToolBlocks in Various States ===\n")

	blocks := []*tui.ToolBlock{
		tui.NewToolBlock("Bash", "git status", []string{"On branch main"}, tui.WithStatus(tui.StatusComplete)),
		tui.NewToolBlock("Bash", "npm test", []string{}, tui.WithStreaming()),
		tui.NewToolBlock("Read", "config.json", []string{"{\"port\": 8080}"}, tui.WithStatus(tui.StatusComplete)),
		tui.NewToolBlock("Write", "output.txt", []string{"Writing..."}, tui.WithStreaming()),
	}

	for _, block := range blocks {
		block.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		fmt.Println(block.View())
	}

	fmt.Println("\033[2mPress Tab to focus different blocks · Ctrl+O to expand/collapse\033[0m")
}
