package display

import (
	"strings"
	"testing"
)

func TestActionRowConstructor(t *testing.T) {
	items := []ActionItem{{Key: "K", Label: "Keep"}}
	row := NewActionRow(items,
		WithActionRowWidth(40),
		WithActionRowSeparator(" | "),
		WithActionRowAlign("center"),
	)

	if row == nil {
		t.Fatal("expected non-nil ActionRow")
	}
	if row.width != 40 {
		t.Fatalf("expected width 40, got %d", row.width)
	}
	if row.separator != " | " {
		t.Fatalf("expected custom separator, got %q", row.separator)
	}
	if row.align != "center" {
		t.Fatalf("expected center alignment, got %q", row.align)
	}
	if len(row.items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(row.items))
	}
}

func TestActionRowItemsRender(t *testing.T) {
	row := NewActionRow([]ActionItem{
		{Key: "Ctrl+C", Label: "Cancel"},
		{Key: "Enter", Label: "Run"},
	})

	view := row.View()
	if view == "" {
		t.Fatal("expected non-empty action row view")
	}

	plain := stripANSI(view)
	if !strings.Contains(plain, "Ctrl+C Cancel") {
		t.Fatalf("expected first item in view, got %q", plain)
	}
	if !strings.Contains(plain, "Enter Run") {
		t.Fatalf("expected second item in view, got %q", plain)
	}
}

func TestActionRowSeparator(t *testing.T) {
	row := NewActionRow([]ActionItem{
		{Key: "A", Label: "Alpha"},
		{Key: "B", Label: "Beta"},
	}, WithActionRowSeparator(" | "))

	plain := stripANSI(row.View())
	if !strings.Contains(plain, "A Alpha | B Beta") {
		t.Fatalf("expected custom separator in view, got %q", plain)
	}
}

func TestActionRowAlignment(t *testing.T) {
	left := stripANSI(NewActionRow([]ActionItem{{Key: "K", Label: "Label"}}, WithActionRowWidth(20), WithActionRowAlign("left")).View())
	center := stripANSI(NewActionRow([]ActionItem{{Key: "K", Label: "Label"}}, WithActionRowWidth(20), WithActionRowAlign("center")).View())
	right := stripANSI(NewActionRow([]ActionItem{{Key: "K", Label: "Label"}}, WithActionRowWidth(20), WithActionRowAlign("right")).View())

	if strings.HasPrefix(left, " ") {
		t.Fatalf("expected left-aligned output to start at column 0, got %q", left)
	}
	if !strings.HasPrefix(right, " ") {
		t.Fatalf("expected right-aligned output to have leading spaces, got %q", right)
	}

	centerLeftPadding := len(center) - len(strings.TrimLeft(center, " "))
	centerRightPadding := len(center) - len(strings.TrimRight(center, " "))
	if centerLeftPadding == 0 || centerRightPadding == 0 {
		t.Fatalf("expected center-aligned output to have padding on both sides, got %q", center)
	}
}
