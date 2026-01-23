package main

import (
	"fmt"

	"github.com/SCKelemen/tui/table"
)

func main() {
	fmt.Println("=== Table Example ===\n")

	// Example 1: Service status table (default rounded borders)
	fmt.Println("1. Service Status (Rounded borders):")
	t1 := table.New("Service", "Status", "Uptime", "Memory")
	t1.AddRow("api-gateway", "Running", "5d 3h", "256MB")
	t1.AddRow("auth-service", "Running", "5d 3h", "128MB")
	t1.AddRow("cache-redis", "Stopped", "0s", "0MB")
	t1.AddRow("db-postgres", "Running", "12d 5h", "2.1GB")
	t1.Print()

	fmt.Println()

	// Example 2: Double-line borders
	fmt.Println("2. User List (Double borders):")
	t2 := table.New("ID", "Username", "Role", "Last Login")
	t2.SetBorderStyle(table.BorderStyleDouble)
	t2.AddRow("001", "alice", "Admin", "2024-01-23 10:30")
	t2.AddRow("002", "bob", "User", "2024-01-23 09:15")
	t2.AddRow("003", "charlie", "Moderator", "2024-01-22 18:45")
	t2.Print()

	fmt.Println()

	// Example 3: ASCII borders (maximum compatibility)
	fmt.Println("3. Task List (ASCII borders):")
	t3 := table.New("Task", "Priority", "Status", "Assigned To")
	t3.SetBorderStyle(table.BorderStyleASCII)
	t3.AddRow("Fix login bug", "High", "In Progress", "Alice")
	t3.AddRow("Update docs", "Low", "Todo", "Bob")
	t3.AddRow("Deploy v2.0", "Critical", "Done", "Charlie")
	t3.Print()

	fmt.Println()

	// Example 4: Simple mode (no row separators)
	fmt.Println("4. Container List (Simple mode):")
	t4 := table.New("Container ID", "Image", "Status", "Ports")
	t4.AddRow("a1b2c3d4", "nginx:latest", "Up 3 hours", "80->8080")
	t4.AddRow("e5f6g7h8", "postgres:14", "Up 5 days", "5432->5432")
	t4.AddRow("i9j0k1l2", "redis:alpine", "Up 2 days", "6379->6379")
	t4.AddRow("m3n4o5p6", "mongo:5.0", "Exited", "-")
	t4.PrintSimple()

	fmt.Println()

	// Example 5: Headers without bold
	fmt.Println("5. Configuration (No bold headers):")
	t5 := table.New("Key", "Value", "Source")
	t5.SetHeaderBold(false)
	t5.AddRow("api.timeout", "30s", "config.yaml")
	t5.AddRow("api.retries", "3", "environment")
	t5.AddRow("log.level", "info", "default")
	t5.Print()

	fmt.Println()

	// Example 6: Dynamic data with AddRows
	fmt.Println("6. Metrics (Batch add):")
	t6 := table.New("Metric", "Value", "Unit", "Trend")
	metrics := [][]string{
		{"CPU Usage", "45%", "percent", "↑"},
		{"Memory", "2.1GB", "gigabytes", "→"},
		{"Disk I/O", "150MB/s", "megabytes/sec", "↓"},
		{"Network", "1.2Gbps", "gigabits/sec", "↑"},
	}
	t6.AddRows(metrics)
	t6.Print()

	fmt.Println()

	// Example 7: Empty and wide columns
	fmt.Println("7. Mixed Content:")
	t7 := table.New("Short", "Very Long Column Name Here", "Empty")
	t7.AddRow("A", "This is a much longer piece of text", "")
	t7.AddRow("B", "Short", "")
	t7.AddRow("C", "Another very long string that will expand the column", "X")
	t7.PrintSimple()

	fmt.Println("\n=== All Examples Complete ===")
}
