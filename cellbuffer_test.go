package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNewCellBuffer(t *testing.T) {
	cb := NewCellBuffer(3, 2)

	if cb.width != 3 || cb.height != 2 {
		t.Fatalf("dimensions = %dx%d, want 3x2", cb.width, cb.height)
	}
	if len(cb.front) != 6 || len(cb.back) != 6 {
		t.Fatalf("buffer lengths = front=%d back=%d, want 6/6", len(cb.front), len(cb.back))
	}
	for i, cell := range cb.front {
		if cell != blankCell() {
			t.Fatalf("front[%d] = %#v, want blank cell %#v", i, cell, blankCell())
		}
	}
}

func TestSetCell(t *testing.T) {
	cb := NewCellBuffer(4, 1)
	cb.SetCell(1, 0, Cell{Grapheme: "界", Width: 2, Fg: lipgloss.Color("#112233")})

	base := cb.back[cb.index(1, 0)]
	if base.Grapheme != "界" || base.Width != 2 {
		t.Fatalf("base cell = %#v, want grapheme 界 width 2", base)
	}
	if !base.Dirty {
		t.Fatal("base cell Dirty = false, want true")
	}

	cont := cb.back[cb.index(2, 0)]
	if cont.Width != 0 {
		t.Fatalf("continuation width = %d, want 0", cont.Width)
	}
	if !cont.Dirty {
		t.Fatal("continuation Dirty = false, want true")
	}
}

func TestWriteString(t *testing.T) {
	cb := NewCellBuffer(4, 1)
	fg := lipgloss.Color("#010203")
	bg := lipgloss.Color("#AABBCC")

	cb.WriteString(0, 0, "A界", fg, bg)

	first := cb.back[cb.index(0, 0)]
	if first.Grapheme != "A" || first.Width != 1 {
		t.Fatalf("first cell = %#v, want A width 1", first)
	}
	if first.Fg != fg || first.Bg != bg {
		t.Fatalf("first colors = (%q,%q), want (%q,%q)", first.Fg, first.Bg, fg, bg)
	}

	wide := cb.back[cb.index(1, 0)]
	if wide.Grapheme != "界" || wide.Width != 2 {
		t.Fatalf("wide cell = %#v, want 界 width 2", wide)
	}
	if cb.back[cb.index(2, 0)].Width != 0 {
		t.Fatalf("continuation width = %d, want 0", cb.back[cb.index(2, 0)].Width)
	}
}

func TestClear(t *testing.T) {
	cb := NewCellBuffer(2, 1)
	cb.SetCell(0, 0, Cell{Rune: 'X'})
	_ = cb.Flush()

	cb.Clear()
	changes := cb.Diff()
	if len(changes) != 1 {
		t.Fatalf("len(Diff()) after Clear = %d, want 1", len(changes))
	}
	if changes[0].Cell.Grapheme != " " || changes[0].Cell.Width != 1 {
		t.Fatalf("changes[0].Cell = %#v, want blank-like cell", changes[0].Cell)
	}
}
func TestDiff(t *testing.T) {
	cb := NewCellBuffer(2, 1)
	cb.SetCell(0, 0, Cell{Rune: 'X'})

	changes := cb.Diff()
	if len(changes) != 1 {
		t.Fatalf("len(Diff()) = %d, want 1", len(changes))
	}
	if changes[0].X != 0 || changes[0].Y != 0 || changes[0].Cell.Grapheme != "X" {
		t.Fatalf("changes[0] = %#v, want change at 0,0 with X", changes[0])
	}

	_ = cb.Flush()
	cb.SetCell(0, 0, Cell{Rune: 'X'})
	if got := cb.Diff(); len(got) != 0 {
		t.Fatalf("len(Diff()) for unchanged cell = %d, want 0", len(got))
	}
}

func TestFlush(t *testing.T) {
	cb := NewCellBuffer(2, 1)
	cb.SetCell(0, 0, Cell{Rune: 'X', Fg: lipgloss.Color("#FF0000")})

	out := cb.Flush()
	if out == "" {
		t.Fatal("Flush() output is empty, want ANSI output")
	}
	if !strings.Contains(out, "\x1b[1;1H") {
		t.Fatalf("Flush() output = %q, want cursor move", out)
	}
	if !strings.Contains(out, "X") {
		t.Fatalf("Flush() output = %q, want rendered rune", out)
	}

	flushed := cb.CellAt(0, 0)
	if flushed.Grapheme != "X" {
		t.Fatalf("CellAt(0,0) after Flush = %#v, want X", flushed)
	}
	if out2 := cb.Flush(); out2 != "" {
		t.Fatalf("second Flush() = %q, want empty string", out2)
	}
}

func TestResize(t *testing.T) {
	cb := NewCellBuffer(2, 1)
	cb.SetCell(0, 0, Cell{Rune: 'X'})
	_ = cb.Flush()

	cb.Resize(3, 2)
	if cb.width != 3 || cb.height != 2 {
		t.Fatalf("dimensions after Resize = %dx%d, want 3x2", cb.width, cb.height)
	}
	if len(cb.front) != 6 || len(cb.back) != 6 {
		t.Fatalf("buffer lengths after Resize = front=%d back=%d, want 6/6", len(cb.front), len(cb.back))
	}
	if got := cb.CellAt(0, 0); got != blankCell() {
		t.Fatalf("CellAt(0,0) after Resize = %#v, want blank %#v", got, blankCell())
	}
}

func TestCellAt(t *testing.T) {
	cb := NewCellBuffer(2, 1)
	cb.SetCell(1, 0, Cell{Rune: 'Y'})
	_ = cb.Flush()

	if got := cb.CellAt(1, 0); got.Grapheme != "Y" {
		t.Fatalf("CellAt(1,0) = %#v, want Y", got)
	}
	if got := cb.CellAt(-1, 0); got != blankCell() {
		t.Fatalf("CellAt(-1,0) = %#v, want blank %#v", got, blankCell())
	}
}

func TestColorEpsilon(t *testing.T) {
	if !colorsEqual(lipgloss.Color("#010203"), lipgloss.Color("#030506"), cellColorEpsilon) {
		t.Fatal("colorsEqual() within epsilon = false, want true")
	}
	if colorsEqual(lipgloss.Color("#010203"), lipgloss.Color("#05090D"), cellColorEpsilon) {
		t.Fatal("colorsEqual() outside epsilon = true, want false")
	}
}
