package table

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	table := New("Name", "Age", "City")
	if len(table.headers) != 3 {
		t.Errorf("expected 3 headers, got %d", len(table.headers))
	}
	if len(table.widths) != 3 {
		t.Errorf("expected 3 width entries, got %d", len(table.widths))
	}
	// Widths should be initialized to header lengths
	if table.widths[0] != 4 { // "Name"
		t.Errorf("expected width[0]=4, got %d", table.widths[0])
	}
}

func TestAddRow(t *testing.T) {
	table := New("Name", "Age")
	table.AddRow("Alice", "30")
	table.AddRow("Bob", "25")

	if len(table.rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(table.rows))
	}

	// Check width calculation
	if table.widths[0] < 5 { // "Alice" is 5 chars
		t.Errorf("expected widths[0] >= 5, got %d", table.widths[0])
	}
}

func TestAddRowWithPadding(t *testing.T) {
	table := New("Name", "Age")
	table.AddRow("Alice") // Missing second column

	if len(table.rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(table.rows))
	}
	if len(table.rows[0]) != 2 {
		t.Errorf("expected row to be padded to 2 columns, got %d", len(table.rows[0]))
	}
}

func TestRender(t *testing.T) {
	table := New("Name", "Age")
	table.AddRow("Alice", "30")
	table.AddRow("Bob", "25")

	output := table.Render()

	// Check that output contains expected elements
	if !strings.Contains(output, "Name") {
		t.Error("output should contain header 'Name'")
	}
	if !strings.Contains(output, "Alice") {
		t.Error("output should contain 'Alice'")
	}
	if !strings.Contains(output, "Bob") {
		t.Error("output should contain 'Bob'")
	}

	// Should have borders
	if !strings.Contains(output, "┌") || !strings.Contains(output, "└") {
		t.Error("output should contain border characters")
	}
}

func TestRenderSimple(t *testing.T) {
	table := New("Name", "Age")
	table.AddRow("Alice", "30")
	table.AddRow("Bob", "25")

	simpleOutput := table.RenderSimple()
	fullOutput := table.Render()

	// Simple output should be shorter (no row separators)
	if len(simpleOutput) >= len(fullOutput) {
		t.Error("simple output should be shorter than full output")
	}

	// Should still contain data
	if !strings.Contains(simpleOutput, "Alice") {
		t.Error("simple output should contain 'Alice'")
	}
}

func TestBorderStyles(t *testing.T) {
	table := New("A", "B")
	table.AddRow("1", "2")

	// Test rounded (default)
	roundedOutput := table.Render()
	if !strings.Contains(roundedOutput, "┌") {
		t.Error("rounded border should contain ┌")
	}

	// Test double
	table.SetBorderStyle(BorderStyleDouble)
	doubleOutput := table.Render()
	if !strings.Contains(doubleOutput, "╔") {
		t.Error("double border should contain ╔")
	}

	// Test ASCII
	table.SetBorderStyle(BorderStyleASCII)
	asciiOutput := table.Render()
	if !strings.Contains(asciiOutput, "+") {
		t.Error("ASCII border should contain +")
	}
	if !strings.Contains(asciiOutput, "-") {
		t.Error("ASCII border should contain -")
	}
}

func TestHeaderBold(t *testing.T) {
	table := New("Name")
	table.AddRow("Alice")

	// With bold (default)
	withBold := table.Render()
	if !strings.Contains(withBold, "\033[1m") {
		t.Error("output with bold should contain ANSI bold code")
	}

	// Without bold
	table.SetHeaderBold(false)
	withoutBold := table.Render()
	if strings.Contains(withoutBold, "\033[1m") {
		t.Error("output without bold should not contain ANSI bold code")
	}
}

func TestClear(t *testing.T) {
	table := New("Name", "Age")
	table.AddRow("Alice", "30")
	table.AddRow("Bob", "25")

	if len(table.rows) != 2 {
		t.Errorf("expected 2 rows before clear, got %d", len(table.rows))
	}

	table.Clear()

	if len(table.rows) != 0 {
		t.Errorf("expected 0 rows after clear, got %d", len(table.rows))
	}

	// Headers should still be there
	if len(table.headers) != 2 {
		t.Errorf("expected headers to remain after clear, got %d", len(table.headers))
	}
}

func TestAddRows(t *testing.T) {
	table := New("Name", "Age")
	rows := [][]string{
		{"Alice", "30"},
		{"Bob", "25"},
		{"Charlie", "35"},
	}
	table.AddRows(rows)

	if len(table.rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(table.rows))
	}
}

func TestEmptyTable(t *testing.T) {
	table := New()
	output := table.Render()

	if output != "" {
		t.Error("empty table should render as empty string")
	}
}

func TestWidthCalculation(t *testing.T) {
	table := New("A", "B")
	table.AddRow("Short", "Very Long String")

	// Width should expand to fit longest content
	if table.widths[1] < len("Very Long String") {
		t.Errorf("expected width[1] >= %d, got %d", len("Very Long String"), table.widths[1])
	}
}

func TestString(t *testing.T) {
	table := New("Name")
	table.AddRow("Alice")

	// String() should be the same as Render()
	if table.String() != table.Render() {
		t.Error("String() should return same output as Render()")
	}
}

// Example test that demonstrates usage
func ExampleTable() {
	table := New("Name", "Status", "Age")
	table.AddRow("service-a", "Running", "2d")
	table.AddRow("service-b", "Stopped", "5h")
	table.Print()
	// Output will show a formatted table
}

func BenchmarkRender(b *testing.B) {
	table := New("Name", "Age", "City", "Country")
	for i := 0; i < 100; i++ {
		table.AddRow("Alice", "30", "New York", "USA")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = table.Render()
	}
}

func BenchmarkAddRow(b *testing.B) {
	table := New("Name", "Age", "City", "Country")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		table.AddRow("Alice", "30", "New York", "USA")
	}
}
