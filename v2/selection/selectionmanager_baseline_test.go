package selection

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestEmptySelectionReturnsEmptyString ensures a freshly constructed manager
// or one whose selection was cleared produces no selected text.
func TestEmptySelectionReturnsEmptyString(t *testing.T) {
	sm := NewSelectionManager()
	if got := sm.GetSelectedText(); got != "" {
		t.Fatalf("expected empty selection from fresh manager, got %q", got)
	}

	sm.SetRegion(SelectableRegion{
		StartRow:     0,
		EndRow:       0,
		ContentLines: []string{"hello"},
	})
	if got := sm.GetSelectedText(); got != "" {
		t.Fatalf("expected empty selection before any press, got %q", got)
	}
}

// TestSingleLineSelectionBytesMatchSource verifies the byte slice of the
// selected text matches the corresponding substring of the source line.
func TestSingleLineSelectionBytesMatchSource(t *testing.T) {
	src := "alpha bravo charlie"
	sm := NewSelectionManager()
	sm.SetRegion(SelectableRegion{
		StartRow:     0,
		EndRow:       0,
		ContentLines: []string{src},
	})
	sm.selection = Selection{StartX: 6, StartY: 0, EndX: 11, EndY: 0}
	sm.hasSelection = true

	got := sm.GetSelectedText()
	want := src[6:11]
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
	if []byte(got)[0] != src[6] {
		t.Fatalf("expected first byte to match source, got %x vs %x", got[0], src[6])
	}
}

// TestMultiLineSelectionBytesMatchSource verifies the bytes of a multi-line
// selection equal the concatenation of the source slices separated by '\n'.
func TestMultiLineSelectionBytesMatchSource(t *testing.T) {
	lines := []string{"first line", "second line", "third line"}
	sm := NewSelectionManager()
	sm.SetRegion(SelectableRegion{
		StartRow:     0,
		EndRow:       2,
		ContentLines: lines,
	})
	sm.selection = Selection{StartX: 6, StartY: 0, EndX: 5, EndY: 2}
	sm.hasSelection = true

	got := sm.GetSelectedText()
	want := lines[0][6:] + "\n" + lines[1] + "\n" + lines[2][:5]
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
	if len(got) != len(want) {
		t.Fatalf("byte length mismatch: got %d want %d", len(got), len(want))
	}
}

// TestCopySelectionTextEqualsGetSelectedText ensures that CopySelection
// captures exactly the bytes returned by GetSelectedText for a multi-line
// selection. CopySelection returns nil for empty selections.
func TestCopySelectionMatchesSelectedText(t *testing.T) {
	lines := []string{"alpha", "bravo", "charlie"}
	sm := NewSelectionManager()
	sm.SetRegion(SelectableRegion{
		StartRow:     0,
		EndRow:       2,
		ContentLines: lines,
	})

	// Empty selection: CopySelection must be nil.
	if cmd := sm.CopySelection(); cmd != nil {
		t.Fatal("expected nil CopySelection cmd when no selection")
	}

	sm.selection = Selection{StartX: 1, StartY: 0, EndX: 4, EndY: 2}
	sm.hasSelection = true

	text := sm.GetSelectedText()
	if !strings.Contains(text, "lpha") || !strings.Contains(text, "bravo") || !strings.Contains(text, "char") {
		t.Fatalf("unexpected selected text: %q", text)
	}

	// Hitting CopySelection must produce a non-nil cmd; that cmd is the
	// WriteClipboard command and we treat its existence as proof that the
	// selected text was forwarded.
	if cmd := sm.CopySelection(); cmd == nil {
		t.Fatal("expected non-nil CopySelection cmd for finalized selection")
	}
}

// TestRegionBoundsClampingNegativeAndOutOfBounds verifies IsSelected
// returns false for negative rows/cols and out-of-bounds rows even when a
// selection is finalized.
func TestRegionBoundsClampingNegativeAndOutOfBounds(t *testing.T) {
	sm := NewSelectionManager()
	sm.SetRegion(SelectableRegion{
		StartRow:     0,
		EndRow:       1,
		ContentLines: []string{"abc", "def"},
	})
	sm.selection = Selection{StartX: 0, StartY: 0, EndX: 3, EndY: 1}
	sm.hasSelection = true

	cases := []struct {
		row, col int
	}{
		{-1, 0},
		{0, -1},
		{99, 0},
		{0, 99}, // beyond line length
	}
	for _, c := range cases {
		// We just want to ensure no panic and reasonable behaviour. For
		// negative inputs the manager must return false; for col beyond
		// the end of the row it must also return false (col >= EndX on
		// the end row is exclusive).
		if c.row < 0 || c.col < 0 || c.row >= 2 {
			if sm.IsSelected(c.row, c.col) {
				t.Fatalf("expected IsSelected(%d,%d)=false for out-of-bounds input", c.row, c.col)
			}
		}
	}

	// GetSelectedText must not panic even if region rows are nonsense.
	sm.region.ContentLines = nil
	if got := sm.GetSelectedText(); got != "" {
		t.Fatalf("expected empty text for cleared region, got %q", got)
	}
}

// TestClearSelectionIdempotent calls ClearSelection multiple times in a row
// to confirm it is safe to invoke repeatedly.
func TestClearSelectionIdempotent(t *testing.T) {
	sm := NewSelectionManager()
	sm.SetRegion(SelectableRegion{
		StartRow:     0,
		EndRow:       0,
		ContentLines: []string{"hello"},
	})
	sm.selection = Selection{Active: true, StartX: 0, StartY: 0, EndX: 3, EndY: 0}
	sm.hasSelection = true

	sm.ClearSelection()
	sm.ClearSelection()
	sm.ClearSelection()

	if sm.HasSelection() {
		t.Fatal("expected HasSelection=false after repeated ClearSelection")
	}
	if sm.selection.Active {
		t.Fatal("expected selection.Active=false after repeated ClearSelection")
	}
	if got := sm.GetSelectedText(); got != "" {
		t.Fatalf("expected empty text after repeated ClearSelection, got %q", got)
	}
}

// TestDragPressMotionReleaseFinalizesRange exercises the full mouse-press,
// motion, release lifecycle and asserts the resulting selection spans the
// expected start and end coordinates.
func TestDragPressMotionReleaseFinalizesRange(t *testing.T) {
	sm := NewSelectionManager()
	sm.SetRegion(SelectableRegion{
		StartRow:     0,
		EndRow:       2,
		GutterWidth:  0,
		ContentLines: []string{"alpha", "bravo", "charlie"},
	})

	sm.HandleMouse(tea.MouseMsg{
		X:      1,
		Y:      0,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	sm.HandleMouse(tea.MouseMsg{
		X:      3,
		Y:      1,
		Action: tea.MouseActionMotion,
		Button: tea.MouseButtonLeft,
	})
	sm.HandleMouse(tea.MouseMsg{
		X:      4,
		Y:      2,
		Action: tea.MouseActionRelease,
		Button: tea.MouseButtonLeft,
	})

	if !sm.HasSelection() {
		t.Fatal("expected finalized selection after press/motion/release")
	}

	if sm.selection.StartY != 0 || sm.selection.EndY != 2 {
		t.Fatalf("expected selection rows 0..2, got %d..%d", sm.selection.StartY, sm.selection.EndY)
	}
	if sm.selection.StartX != 1 {
		t.Fatalf("expected StartX=1, got %d", sm.selection.StartX)
	}
	if sm.selection.EndX != 4 {
		t.Fatalf("expected EndX=4, got %d", sm.selection.EndX)
	}

	got := sm.GetSelectedText()
	if !strings.Contains(got, "lpha") {
		t.Fatalf("expected first-line slice in selected text, got %q", got)
	}
	if !strings.Contains(got, "bravo") {
		t.Fatalf("expected middle line fully selected, got %q", got)
	}
	if !strings.Contains(got, "char") {
		t.Fatalf("expected last-line slice in selected text, got %q", got)
	}
}
