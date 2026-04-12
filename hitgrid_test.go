package tui

import "testing"

func TestNewHitGrid(t *testing.T) {
	hg := NewHitGrid(3, 2)

	if hg.width != 3 || hg.height != 2 {
		t.Fatalf("dimensions = %dx%d, want 3x2", hg.width, hg.height)
	}
	if len(hg.grid) != 2 || len(hg.next) != 2 {
		t.Fatalf("row counts = grid=%d next=%d, want 2/2", len(hg.grid), len(hg.next))
	}
	if hg.IsDirty() {
		t.Fatal("IsDirty() = true for new grid, want false")
	}
}

func TestRegister(t *testing.T) {
	hg := NewHitGrid(4, 3)
	hg.Register("button", 1, 1, 2, 2)

	if !hg.IsDirty() {
		t.Fatal("IsDirty() after Register = false, want true")
	}
	if hg.next[1][1] != "button" || hg.next[2][2] != "button" {
		t.Fatalf("next buffer registration incorrect: %#v", hg.next)
	}
}

func TestHitTest(t *testing.T) {
	hg := NewHitGrid(4, 3)
	hg.Register("button", 1, 1, 2, 1)
	hg.Swap()

	if got := hg.HitTest(1, 1); got != "button" {
		t.Fatalf("HitTest(1,1) = %q, want %q", got, "button")
	}
	if got := hg.HitTest(0, 0); got != "" {
		t.Fatalf("HitTest(0,0) = %q, want empty", got)
	}
}

func TestHitGridClear(t *testing.T) {
	hg := NewHitGrid(3, 2)
	hg.Register("item", 0, 0, 2, 1)
	hg.Swap()

	hg.Clear()
	if !hg.IsDirty() {
		t.Fatal("IsDirty() after Clear = false, want true")
	}
	hg.Swap()
	if got := hg.HitTest(0, 0); got != "" {
		t.Fatalf("HitTest(0,0) after Clear+Swap = %q, want empty", got)
	}
}

func TestSwap(t *testing.T) {
	hg := NewHitGrid(2, 1)
	hg.Register("x", 0, 0, 1, 1)
	hg.Swap()

	if got := hg.HitTest(0, 0); got != "x" {
		t.Fatalf("HitTest(0,0) after Swap = %q, want %q", got, "x")
	}
	if hg.IsDirty() {
		t.Fatal("IsDirty() after Swap = true, want false")
	}
	if hg.next[0][0] != "" {
		t.Fatalf("next buffer after Swap = %q at 0,0, want empty", hg.next[0][0])
	}
}

func TestHitGridResize(t *testing.T) {
	hg := NewHitGrid(2, 1)
	hg.Register("x", 0, 0, 1, 1)
	hg.Resize(3, 2)

	if hg.width != 3 || hg.height != 2 {
		t.Fatalf("dimensions after Resize = %dx%d, want 3x2", hg.width, hg.height)
	}
	if len(hg.grid) != 2 || len(hg.grid[0]) != 3 {
		t.Fatalf("grid shape after Resize invalid: %#v", hg.grid)
	}
	if hg.IsDirty() {
		t.Fatal("IsDirty() after Resize = true, want false")
	}
}

func TestIsDirty(t *testing.T) {
	hg := NewHitGrid(2, 1)
	if hg.IsDirty() {
		t.Fatal("initial IsDirty() = true, want false")
	}

	hg.Register("x", 0, 0, 1, 1)
	if !hg.IsDirty() {
		t.Fatal("IsDirty() after Register = false, want true")
	}

	hg.Swap()
	if hg.IsDirty() {
		t.Fatal("IsDirty() after Swap = true, want false")
	}

	hg.Clear()
	if !hg.IsDirty() {
		t.Fatal("IsDirty() after Clear of non-empty grid = false, want true")
	}
}
