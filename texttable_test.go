package tui

import (
	"strings"
	"testing"
)

func testTextTableColumns() []Column {
	return []Column{{Name: "Name", Width: 0}, {Name: "Role", Width: 0}}
}

func TestNewTextTable(t *testing.T) {
	tt := NewTextTable(testTextTableColumns())
	if tt == nil {
		t.Fatal("NewTextTable returned nil")
	}
	if tt.borderStyle != TextTableBorderRounded {
		t.Fatalf("borderStyle = %v, want rounded", tt.borderStyle)
	}
	if tt.selectedRow != -1 {
		t.Fatalf("selectedRow = %d, want -1", tt.selectedRow)
	}
}

func TestTextTableCreation(t *testing.T) {
	TestNewTextTable(t)
}

func TestAddRow(t *testing.T) {
	tt := NewTextTable(testTextTableColumns())
	tt.AddRow("Sam", "Engineer")
	if len(tt.rows) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(tt.rows))
	}
	if tt.selectedRow != 0 {
		t.Fatalf("selectedRow = %d, want 0", tt.selectedRow)
	}
}

func TestSetRows(t *testing.T) {
	tt := NewTextTable(testTextTableColumns())
	rows := [][]string{{"Sam", "Engineer"}, {"Pat", "Manager"}}
	tt.SetRows(rows)
	rows[0][0] = "mutated"
	if tt.rows[0][0] != "Sam" {
		t.Fatalf("SetRows should clone input rows, got %q", tt.rows[0][0])
	}
}

func TestAutoWidth(t *testing.T) {
	tt := NewTextTable([]Column{{Name: "Name"}, {Name: "Description"}})
	tt.SetRows([][]string{{"Sam", "Principal Engineer"}})
	widths := tt.columnWidths()
	if widths[0] != 4 {
		t.Fatalf("name column width = %d, want 4", widths[0])
	}
	if widths[1] != len("Principal Engineer") {
		t.Fatalf("description width = %d, want %d", widths[1], len("Principal Engineer"))
	}
}

func TestAlignment(t *testing.T) {
	tt := NewTextTable([]Column{{Name: "L", Width: 5, Alignment: TextTableAlignLeft}, {Name: "C", Width: 5, Alignment: TextTableAlignCenter}, {Name: "R", Width: 5, Alignment: TextTableAlignRight}}, WithTextTableBorderStyle(TextTableBorderNone))
	if got := tt.formatCell("a", 5, TextTableAlignLeft); got != "a    " {
		t.Fatalf("left aligned cell = %q, want %q", got, "a    ")
	}
	if got := tt.formatCell("b", 5, TextTableAlignCenter); got != "  b  " {
		t.Fatalf("center aligned cell = %q, want %q", got, "  b  ")
	}
	if got := tt.formatCell("c", 5, TextTableAlignRight); got != "    c" {
		t.Fatalf("right aligned cell = %q, want %q", got, "    c")
	}
}
func TestZebraStriping(t *testing.T) {
	tt := NewTextTable(testTextTableColumns(), WithTextTableZebraStriping(true))
	tt.SetRows([][]string{{"Sam", "Engineer"}, {"Pat", "Manager"}})
	view := tt.View()
	if !strings.Contains(view, tt.zebraStyle) {
		t.Fatalf("View() missing zebra style %q in %q", tt.zebraStyle, view)
	}
}

func TestBorderStyles(t *testing.T) {
	cases := []struct {
		style TextTableBorderStyle
		left  string
		right string
	}{
		{TextTableBorderSingle, "┌", "┐"},
		{TextTableBorderDouble, "╔", "╗"},
		{TextTableBorderRounded, "╭", "╮"},
	}

	for _, tc := range cases {
		t.Run(tc.left, func(t *testing.T) {
			tt := NewTextTable(testTextTableColumns(), WithTextTableBorderStyle(tc.style))
			tt.AddRow("Sam", "Engineer")
			plain := stripANSI(tt.View())
			if !strings.Contains(plain, tc.left) || !strings.Contains(plain, tc.right) {
				t.Fatalf("View() = %q, want border chars %q and %q", plain, tc.left, tc.right)
			}
		})
	}
}

func TestTruncation(t *testing.T) {
	tt := NewTextTable([]Column{{Name: "Name", MaxWidth: 4}}, WithTextTableBorderStyle(TextTableBorderNone))
	got := tt.formatCell("alphabet", 4, TextTableAlignLeft)
	if got != "alp…" {
		t.Fatalf("formatCell truncation = %q, want alp…", got)
	}
}

func TestView(t *testing.T) {
	tt := NewTextTable(testTextTableColumns(), WithTextTableBorderStyle(TextTableBorderSingle))
	tt.AddRow("Sam", "Engineer")
	tt.AddRow("Pat", "Manager")
	plain := stripANSI(tt.View())
	for _, want := range []string{"Name", "Role", "Sam", "Engineer", "Pat", "Manager", "┌", "└"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("View() = %q, missing %q", plain, want)
		}
	}
}
