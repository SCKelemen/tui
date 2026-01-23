// Package table provides static table rendering for CLI output.
//
// This package is designed for non-interactive command output (like kubectl get, ls -l, etc.)
// and is separate from the interactive Bubble Tea components in the parent tui package.
//
// Example usage:
//
//	table := table.New("Name", "Status", "Age")
//	table.AddRow("service-a", "Running", "2d")
//	table.AddRow("service-b", "Stopped", "5h")
//	fmt.Println(table.Render())
//
// Output:
//
//	┌───────────┬─────────┬─────┐
//	│ Name      │ Status  │ Age │
//	├───────────┼─────────┼─────┤
//	│ service-a │ Running │ 2d  │
//	├───────────┼─────────┼─────┤
//	│ service-b │ Stopped │ 5h  │
//	└───────────┴─────────┴─────┘
package table

import (
	"fmt"
	"strings"
)

// BorderStyle defines the characters used for table borders
type BorderStyle struct {
	TopLeft      string
	TopRight     string
	BottomLeft   string
	BottomRight  string
	Horizontal   string
	Vertical     string
	Cross        string
	LeftT        string
	RightT       string
	TopT         string
	BottomT      string
}

// Common border styles
var (
	// BorderStyleRounded uses rounded Unicode box-drawing characters
	BorderStyleRounded = BorderStyle{
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
		Horizontal:  "─",
		Vertical:    "│",
		Cross:       "┼",
		LeftT:       "├",
		RightT:      "┤",
		TopT:        "┬",
		BottomT:     "┴",
	}

	// BorderStyleDouble uses double-line box-drawing characters
	BorderStyleDouble = BorderStyle{
		TopLeft:     "╔",
		TopRight:    "╗",
		BottomLeft:  "╚",
		BottomRight: "╝",
		Horizontal:  "═",
		Vertical:    "║",
		Cross:       "╬",
		LeftT:       "╠",
		RightT:      "╣",
		TopT:        "╦",
		BottomT:     "╩",
	}

	// BorderStyleASCII uses plain ASCII characters for maximum compatibility
	BorderStyleASCII = BorderStyle{
		TopLeft:     "+",
		TopRight:    "+",
		BottomLeft:  "+",
		BottomRight: "+",
		Horizontal:  "-",
		Vertical:    "|",
		Cross:       "+",
		LeftT:       "+",
		RightT:      "+",
		TopT:        "+",
		BottomT:     "+",
	}
)

// Table represents a static table for CLI output
type Table struct {
	headers     []string
	rows        [][]string
	widths      []int
	borderStyle BorderStyle
	headerBold  bool
}

// New creates a new table with the given headers
func New(headers ...string) *Table {
	t := &Table{
		headers:     headers,
		widths:      make([]int, len(headers)),
		borderStyle: BorderStyleRounded,
		headerBold:  true,
	}

	// Initialize widths with header lengths
	for i, h := range headers {
		t.widths[i] = len(h)
	}

	return t
}

// SetBorderStyle sets the border style for the table
func (t *Table) SetBorderStyle(style BorderStyle) {
	t.borderStyle = style
}

// SetHeaderBold controls whether headers are rendered in bold
func (t *Table) SetHeaderBold(bold bool) {
	t.headerBold = bold
}

// AddRow adds a row to the table
func (t *Table) AddRow(cells ...string) {
	// Pad cells to match header count
	row := make([]string, len(t.headers))
	copy(row, cells)

	// Update column widths
	for i, cell := range row {
		if i < len(t.widths) && len(cell) > t.widths[i] {
			t.widths[i] = len(cell)
		}
	}

	t.rows = append(t.rows, row)
}

// AddRows adds multiple rows to the table
func (t *Table) AddRows(rows [][]string) {
	for _, row := range rows {
		t.AddRow(row...)
	}
}

// Clear removes all rows but keeps headers
func (t *Table) Clear() {
	t.rows = nil
	// Reset widths to header lengths
	for i, h := range t.headers {
		t.widths[i] = len(h)
	}
}

// Render renders the table to a string
func (t *Table) Render() string {
	if len(t.headers) == 0 {
		return ""
	}

	var b strings.Builder

	// Top border
	b.WriteString(t.renderBorder(t.borderStyle.TopLeft, t.borderStyle.TopT, t.borderStyle.TopRight))
	b.WriteString("\n")

	// Headers
	b.WriteString(t.renderRow(t.headers, t.headerBold))
	b.WriteString("\n")

	// Header separator
	b.WriteString(t.renderBorder(t.borderStyle.LeftT, t.borderStyle.Cross, t.borderStyle.RightT))
	b.WriteString("\n")

	// Rows
	for i, row := range t.rows {
		b.WriteString(t.renderRow(row, false))
		b.WriteString("\n")

		// Row separator (except for last row)
		if i < len(t.rows)-1 {
			b.WriteString(t.renderBorder(t.borderStyle.LeftT, t.borderStyle.Cross, t.borderStyle.RightT))
			b.WriteString("\n")
		}
	}

	// Bottom border
	b.WriteString(t.renderBorder(t.borderStyle.BottomLeft, t.borderStyle.BottomT, t.borderStyle.BottomRight))

	return b.String()
}

// String implements the Stringer interface
func (t *Table) String() string {
	return t.Render()
}

// renderBorder renders a horizontal border line
func (t *Table) renderBorder(left, middle, right string) string {
	var parts []string
	for _, width := range t.widths {
		parts = append(parts, strings.Repeat(t.borderStyle.Horizontal, width+2))
	}
	return left + strings.Join(parts, middle) + right
}

// renderRow renders a single row
func (t *Table) renderRow(cells []string, bold bool) string {
	var parts []string
	for i, cell := range cells {
		width := t.widths[i]
		padded := t.pad(cell, width)
		if bold {
			padded = "\033[1m" + padded + "\033[0m"
		}
		parts = append(parts, " "+padded+" ")
	}
	return t.borderStyle.Vertical + strings.Join(parts, t.borderStyle.Vertical) + t.borderStyle.Vertical
}

// pad pads a string to the specified width
func (t *Table) pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// Print renders and prints the table to stdout
func (t *Table) Print() {
	fmt.Println(t.Render())
}

// RenderSimple renders the table without row separators (more compact)
func (t *Table) RenderSimple() string {
	if len(t.headers) == 0 {
		return ""
	}

	var b strings.Builder

	// Top border
	b.WriteString(t.renderBorder(t.borderStyle.TopLeft, t.borderStyle.TopT, t.borderStyle.TopRight))
	b.WriteString("\n")

	// Headers
	b.WriteString(t.renderRow(t.headers, t.headerBold))
	b.WriteString("\n")

	// Header separator
	b.WriteString(t.renderBorder(t.borderStyle.LeftT, t.borderStyle.Cross, t.borderStyle.RightT))
	b.WriteString("\n")

	// Rows (without separators between)
	for _, row := range t.rows {
		b.WriteString(t.renderRow(row, false))
		b.WriteString("\n")
	}

	// Bottom border
	b.WriteString(t.renderBorder(t.borderStyle.BottomLeft, t.borderStyle.BottomT, t.borderStyle.BottomRight))

	return b.String()
}

// PrintSimple renders and prints the table in simple mode (no row separators)
func (t *Table) PrintSimple() {
	fmt.Println(t.RenderSimple())
}
