package selection

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSelectionManager(t *testing.T) {
	sm := NewSelectionManager()
	if sm == nil {
		t.Fatal("NewSelectionManager returned nil")
	}
	if sm.HasSelection() {
		t.Fatal("new manager should not have selection")
	}
}

func TestSelectionManagerSetRegionAndOffset(t *testing.T) {
	sm := NewSelectionManager()
	region := SelectableRegion{
		StartRow:    2,
		EndRow:      4,
		GutterWidth: 6,
		ContentLines: []string{
			"first",
			"second",
			"third",
		},
	}

	sm.SetRegion(region)
	sm.SetOffset(10, 20)

	if sm.region.StartRow != 2 || sm.region.EndRow != 4 {
		t.Fatalf("unexpected region rows: %+v", sm.region)
	}
	if sm.region.GutterWidth != 6 {
		t.Fatalf("unexpected gutter width: %d", sm.region.GutterWidth)
	}
	if sm.offsetX != 10 || sm.offsetY != 20 {
		t.Fatalf("unexpected offsets: x=%d y=%d", sm.offsetX, sm.offsetY)
	}
}

func TestHandleMousePressMotionRelease(t *testing.T) {
	sm := newSelectionManagerForTest()

	press := tea.MouseMsg{
		X:      5,
		Y:      1,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	if !sm.HandleMouse(press) {
		t.Fatal("expected press within region to return true")
	}
	if !sm.selection.Active {
		t.Fatal("selection should be active after press")
	}
	if sm.selection.StartX != 5 || sm.selection.StartY != 1 {
		t.Fatalf("unexpected start coords: %+v", sm.selection)
	}

	motion := tea.MouseMsg{
		X:      8,
		Y:      2,
		Action: tea.MouseActionMotion,
		Button: tea.MouseButtonLeft,
	}
	if !sm.HandleMouse(motion) {
		t.Fatal("expected motion within region to return true")
	}
	// EndX clamps to the target line length ("charlie" => 7 runes).
	if sm.selection.EndX != 7 || sm.selection.EndY != 2 {
		t.Fatalf("unexpected end coords after motion: %+v", sm.selection)
	}
	release := tea.MouseMsg{
		X:      10,
		Y:      2,
		Action: tea.MouseActionRelease,
		Button: tea.MouseButtonLeft,
	}
	if !sm.HandleMouse(release) {
		t.Fatal("expected release within region to return true")
	}
	if sm.selection.Active {
		t.Fatal("selection should be inactive after release")
	}
	if !sm.HasSelection() {
		t.Fatal("selection should be finalized after release")
	}
}

func TestHandleMouseOutsideRegion(t *testing.T) {
	sm := newSelectionManagerForTest()

	outsidePress := tea.MouseMsg{
		X:      1,
		Y:      99,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	if sm.HandleMouse(outsidePress) {
		t.Fatal("expected outside press to return false")
	}
	if sm.selection.Active {
		t.Fatal("selection should not become active for outside press")
	}

	if sm.HandleMouse(tea.MouseMsg{Action: tea.MouseActionMotion, X: 1, Y: 1, Button: tea.MouseButtonLeft}) {
		t.Fatal("expected motion with inactive selection to return false")
	}
	if sm.HandleMouse(tea.MouseMsg{Action: tea.MouseActionRelease, X: 1, Y: 1, Button: tea.MouseButtonLeft}) {
		t.Fatal("expected release with inactive selection to return false")
	}
}

func TestHandleMouseAdjustsForGutterWidth(t *testing.T) {
	sm := NewSelectionManager()
	sm.SetRegion(SelectableRegion{
		StartRow:    0,
		EndRow:      0,
		GutterWidth: 4,
		ContentLines: []string{
			"abcdef",
		},
	})
	sm.SetOffset(2, 3)

	press := tea.MouseMsg{X: 8, Y: 3, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}
	if !sm.HandleMouse(press) {
		t.Fatal("expected press in row to be handled")
	}
	if sm.selection.StartX != 2 {
		t.Fatalf("expected start x=2 after offset+gutter adjustment, got %d", sm.selection.StartX)
	}
	if sm.selection.StartY != 0 {
		t.Fatalf("expected start y=0 after offset adjustment, got %d", sm.selection.StartY)
	}
}

func TestGetSelectedTextSingleLine(t *testing.T) {
	sm := newSelectionManagerForTest()
	sm.selection = Selection{StartX: 1, StartY: 0, EndX: 4, EndY: 0}
	sm.hasSelection = true

	got := sm.GetSelectedText()
	if got != "lph" {
		t.Fatalf("expected selected text %q, got %q", "lph", got)
	}
}

func TestGetSelectedTextMultiLine(t *testing.T) {
	sm := newSelectionManagerForTest()
	sm.selection = Selection{StartX: 2, StartY: 0, EndX: 3, EndY: 2}
	sm.hasSelection = true

	got := sm.GetSelectedText()
	want := "pha\nbravo\ncha"
	if got != want {
		t.Fatalf("expected selected text %q, got %q", want, got)
	}
}

func TestGetSelectedTextBackwardSelection(t *testing.T) {
	sm := newSelectionManagerForTest()
	sm.selection = Selection{StartX: 3, StartY: 2, EndX: 1, EndY: 0}
	sm.hasSelection = true

	got := sm.GetSelectedText()
	want := "lpha\nbravo\ncha"
	if got != want {
		t.Fatalf("expected selected text %q, got %q", want, got)
	}
}

func TestGetSelectedTextEmpty(t *testing.T) {
	sm := newSelectionManagerForTest()
	if got := sm.GetSelectedText(); got != "" {
		t.Fatalf("expected empty selected text, got %q", got)
	}
}

func TestIsSelectedCoordinates(t *testing.T) {
	sm := newSelectionManagerForTest()
	sm.selection = Selection{StartX: 1, StartY: 0, EndX: 3, EndY: 2}
	sm.hasSelection = true

	testCases := []struct {
		name string
		row  int
		col  int
		want bool
	}{
		{name: "before start", row: 0, col: 0, want: false},
		{name: "at start", row: 0, col: 1, want: true},
		{name: "middle line", row: 1, col: 4, want: true},
		{name: "end exclusive", row: 2, col: 3, want: false},
		{name: "end selected", row: 2, col: 2, want: true},
		{name: "out of bounds row", row: 99, col: 0, want: false},
		{name: "negative col", row: 1, col: -1, want: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := sm.IsSelected(tc.row, tc.col)
			if got != tc.want {
				t.Fatalf("IsSelected(%d,%d) = %v, want %v", tc.row, tc.col, got, tc.want)
			}
		})
	}
}

func TestStyledLineHighlightsSelection(t *testing.T) {
	sm := newSelectionManagerForTest()
	sm.selection = Selection{StartX: 1, StartY: 1, EndX: 4, EndY: 1}
	sm.hasSelection = true

	got := sm.StyledLine("bravo", 1)
	reverseOn := string([]byte{0x1b}) + "[7m"
	reverseOff := string([]byte{0x1b}) + "[0m"

	if !strings.Contains(got, reverseOn+"rav"+reverseOff) {
		t.Fatalf("expected highlighted span in styled line, got %q", got)
	}

	plain := sm.StyledLine("alpha", 0)
	if strings.Contains(plain, reverseOn) {
		t.Fatalf("did not expect highlight on unselected row, got %q", plain)
	}
}

func TestClearSelectionResetsState(t *testing.T) {
	sm := newSelectionManagerForTest()
	sm.selection = Selection{Active: true, StartX: 1, StartY: 1, EndX: 2, EndY: 2}
	sm.hasSelection = true

	sm.ClearSelection()

	if sm.selection.Active {
		t.Fatal("expected selection to be inactive after clear")
	}
	if sm.HasSelection() {
		t.Fatal("expected hasSelection=false after clear")
	}
	if sm.GetSelectedText() != "" {
		t.Fatal("expected no selected text after clear")
	}
}

func TestCopySelectionReturnsCmdWhenSelectionExists(t *testing.T) {
	sm := newSelectionManagerForTest()
	sm.selection = Selection{StartX: 0, StartY: 0, EndX: 2, EndY: 0}
	sm.hasSelection = true

	if cmd := sm.CopySelection(); cmd == nil {
		t.Fatal("expected non-nil command when selection exists")
	}
}

func TestCopySelectionNilWhenEmpty(t *testing.T) {
	sm := newSelectionManagerForTest()
	if cmd := sm.CopySelection(); cmd != nil {
		t.Fatal("expected nil command for empty selection")
	}
}

func newSelectionManagerForTest() *SelectionManager {
	sm := NewSelectionManager()
	sm.SetRegion(SelectableRegion{
		StartRow:    0,
		EndRow:      2,
		GutterWidth: 0,
		ContentLines: []string{
			"alpha",
			"bravo",
			"charlie",
		},
	})
	return sm
}
