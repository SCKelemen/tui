package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/tui"
)

func main() {
	fmt.Println("=== StructuredData Component Demo ===")

	// Example 1: Cost Summary (like Codex CLI's /cost command)
	fmt.Println("1. Cost Summary:")
	cost := tui.NewStructuredData("Session Summary").
		AddRow("Total cost", "$122.25").
		AddRow("Total duration (API)", "6h 10m 48s").
		AddRow("Total duration (wall)", "1d 20h 37m").
		AddRow("Total code changes", "26773 lines added, 2436 lines removed").
		AddSeparator().
		AddHeader("Usage by model").
		AddIndentedRow("codex-mini", "797.2k input, 65.9k output, 1.9m cache read, 233.5k cache write ($1.61)", 1).
		AddIndentedRow("codex-pro", "44.6k input, 970.4k output, 189.5m cache read, 13.1m cache write ($120.63)", 1)

	cost.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	fmt.Println(cost.View())

	// Example 2: System Information
	fmt.Println("\n2. System Information:")
	sysInfo := tui.NewStructuredData("System Info").
		AddRow("Operating System", "macOS 14.2.1").
		AddRow("Architecture", "arm64").
		AddRow("CPU", "Apple M2 Pro (12 cores)").
		AddRow("Memory", "32 GB").
		AddSeparator().
		AddHeader("Go Environment").
		AddIndentedRow("Version", "go1.21.5", 1).
		AddIndentedRow("GOPATH", "/Users/user/go", 1).
		AddIndentedRow("GOROOT", "/usr/local/go", 1)

	sysInfo.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	fmt.Println(sysInfo.View())

	// Example 3: Test Results Summary
	fmt.Println("\n3. Test Results:")
	tests := tui.NewStructuredData("Test Summary").
		AddColoredRow("Total Tests", "170", "\033[32m").     // Green
		AddColoredRow("Passed", "170", "\033[32m").          // Green
		AddColoredRow("Failed", "0", "\033[2m").             // Dim
		AddColoredRow("Skipped", "0", "\033[2m").            // Dim
		AddRow("Duration", "11.942s").
		AddSeparator().
		AddHeader("Coverage by Component").
		AddIndentedRow("ToolBlock", "27 tests", 1).
		AddIndentedRow("TextInput", "26 tests", 1).
		AddIndentedRow("CommandPalette", "28 tests", 1).
		AddIndentedRow("FileExplorer", "21 tests (3 core + 18 edge)", 1).
		AddIndentedRow("Header", "34 tests (9 core + 25 edge)", 1).
		AddIndentedRow("ActivityBar", "25 tests (2 core + 23 edge)", 1)

	tests.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	fmt.Println(tests.View())

	// Example 4: API Response
	fmt.Println("\n4. API Response:")
	api := tui.NewStructuredData("API Response").
		AddRow("Status", "200 OK").
		AddRow("Content-Type", "application/json").
		AddRow("Response Time", "142ms").
		AddSeparator().
		AddHeader("Response Body").
		AddIndentedValue("{", 1).
		AddIndentedRow("\"id\"", "12345", 2).
		AddIndentedRow("\"name\"", "\"John Doe\"", 2).
		AddIndentedRow("\"email\"", "\"john@example.com\"", 2).
		AddIndentedRow("\"created_at\"", "\"2024-01-10T15:30:00Z\"", 2).
		AddIndentedValue("}", 1)

	api.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	fmt.Println(api.View())

	// Example 5: Configuration Display
	fmt.Println("\n5. Configuration:")
	config := tui.NewStructuredData("App Configuration").
		AddHeader("Server").
		AddIndentedRow("Host", "localhost", 1).
		AddIndentedRow("Port", "8080", 1).
		AddIndentedRow("Protocol", "http", 1).
		AddSeparator().
		AddHeader("Database").
		AddIndentedRow("Driver", "postgresql", 1).
		AddIndentedRow("Host", "db.example.com", 1).
		AddIndentedRow("Database", "myapp_production", 1).
		AddIndentedRow("Pool Size", "20", 1).
		AddSeparator().
		AddHeader("Features").
		AddIndentedRow("Authentication", "enabled", 1).
		AddIndentedRow("Rate Limiting", "enabled (100 req/min)", 1).
		AddIndentedRow("Caching", "redis", 1)

	config.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	fmt.Println(config.View())

	// Example 6: Collapsed View
	fmt.Println("\n6. Collapsed View (MaxLines = 5):")
	large := tui.NewStructuredData("Large Dataset", tui.WithStructuredDataMaxLines(5)).
		AddRow("Item 1", "Value 1").
		AddRow("Item 2", "Value 2").
		AddRow("Item 3", "Value 3").
		AddRow("Item 4", "Value 4").
		AddRow("Item 5", "Value 5").
		AddRow("Item 6", "Value 6").
		AddRow("Item 7", "Value 7").
		AddRow("Item 8", "Value 8").
		AddRow("Item 9", "Value 9").
		AddRow("Item 10", "Value 10")

	large.SetExpanded(false)
	large.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	fmt.Println(large.View())

	// Example 7: Using FromKeyValuePairs helper
	fmt.Println("\n7. Using Helper Functions:")
	simple := tui.FromKeyValuePairs(
		"Quick Data",
		"Name", "StructuredData Component",
		"Type", "TUI Component",
		"Status", "Active",
		"Version", "1.0.0",
	)

	simple.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	fmt.Println(simple.View())

	fmt.Println("\nInteractive features:")
	fmt.Println("  • Tab/Shift+Tab: Navigate between components")
	fmt.Println("  • Ctrl+O or Enter: Expand/collapse (when maxLines is set)")
	fmt.Println("  • Builder pattern for ergonomic API")
}
