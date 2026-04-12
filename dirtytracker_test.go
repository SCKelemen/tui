package tui

import "testing"

func TestDirtyTracker(t *testing.T) {
	dt := NewDirtyTracker()

	if !dt.IsDirty("component-a") {
		t.Fatal("IsDirty(missing) = false, want true")
	}
	if view, ok := dt.GetCachedView("component-a"); ok || view != "" {
		t.Fatalf("GetCachedView(missing) = (%q, %v), want (\"\", false)", view, ok)
	}

	dt.MarkDirty("component-a")
	if !dt.IsDirty("component-a") {
		t.Fatal("IsDirty(after MarkDirty) = false, want true")
	}

	dt.CacheView("component-a", "view-a")
	stateA := dt.componentStates["component-a"]
	if stateA == nil || stateA.dirty {
		t.Fatalf("stateA = %#v, want cached non-dirty state", stateA)
	}
	if stateA.lastHash == 0 {
		t.Fatal("stateA.lastHash = 0, want non-zero hash")
	}
	if stateA.lastRenderTime.IsZero() {
		t.Fatal("stateA.lastRenderTime is zero, want render timestamp")
	}

	if view, ok := dt.GetCachedView("component-a"); !ok || view != "view-a" {
		t.Fatalf("GetCachedView(component-a) = (%q, %v), want (\"view-a\", true)", view, ok)
	}

	dt.MarkDirty("component-a")
	if !dt.IsDirty("component-a") {
		t.Fatal("IsDirty(after second MarkDirty) = false, want true")
	}
	if view, ok := dt.GetCachedView("component-a"); ok || view != "" {
		t.Fatalf("GetCachedView(dirty) = (%q, %v), want (\"\", false)", view, ok)
	}

	dt.CacheView("component-b", "view-b")
	dt.InvalidateAll()
	if !dt.IsDirty("component-a") || !dt.IsDirty("component-b") {
		t.Fatal("InvalidateAll() did not mark all components dirty")
	}

	stats := dt.Stats()
	if stats.Hits != 1 || stats.Misses != 2 || stats.Invalidations != 2 {
		t.Fatalf("Stats() = %#v, want Hits=1 Misses=2 Invalidations=2", stats)
	}
}

func TestViewportCuller(t *testing.T) {
	vc := NewViewportCuller()
	vc.SetViewport(10, 5)

	if !vc.IsVisible(0, 0, 1, 1) {
		t.Fatal("IsVisible(0,0,1,1) = false, want true")
	}
	if !vc.IsVisible(-1, -1, 2, 2) {
		t.Fatal("IsVisible(partial overlap) = false, want true")
	}
	if vc.IsVisible(10, 0, 1, 1) {
		t.Fatal("IsVisible(outside) = true, want false")
	}

	cx, cy, cw, ch := vc.ClipRect(-1, 1, 4, 3)
	if cx != 0 || cy != 1 || cw != 3 || ch != 3 {
		t.Fatalf("ClipRect(-1,1,4,3) = (%d,%d,%d,%d), want (0,1,3,3)", cx, cy, cw, ch)
	}

	start, end := vc.VisibleRows(100, 10, 20)
	if start != 10 || end != 30 {
		t.Fatalf("VisibleRows(100,10,20) = (%d,%d), want (10,30)", start, end)
	}
	start, end = vc.VisibleRows(10, 8, 5)
	if start != 8 || end != 10 {
		t.Fatalf("VisibleRows(10,8,5) = (%d,%d), want (8,10)", start, end)
	}
}

func TestScissorStack(t *testing.T) {
	ss := NewScissorStack()

	if got := ss.Depth(); got != 0 {
		t.Fatalf("initial Depth() = %d, want 0", got)
	}
	if !ss.IsInside(100, 100) {
		t.Fatal("IsInside() with empty stack = false, want true")
	}

	ss.Push(0, 0, 10, 10)
	x, y, w, h := ss.Current()
	if x != 0 || y != 0 || w != 10 || h != 10 {
		t.Fatalf("Current() after first Push = (%d,%d,%d,%d), want (0,0,10,10)", x, y, w, h)
	}
	if got := ss.Depth(); got != 1 {
		t.Fatalf("Depth() after first Push = %d, want 1", got)
	}
	if !ss.IsInside(5, 5) || ss.IsInside(10, 10) {
		t.Fatal("IsInside() returned unexpected result for first clip rect")
	}

	ss.Push(5, 5, 10, 10)
	x, y, w, h = ss.Current()
	if x != 5 || y != 5 || w != 5 || h != 5 {
		t.Fatalf("Current() after intersecting Push = (%d,%d,%d,%d), want (5,5,5,5)", x, y, w, h)
	}
	if !ss.IsInside(7, 7) || ss.IsInside(4, 4) {
		t.Fatal("IsInside() returned unexpected result for intersected clip rect")
	}

	ss.Pop()
	x, y, w, h = ss.Current()
	if x != 0 || y != 0 || w != 10 || h != 10 {
		t.Fatalf("Current() after Pop = (%d,%d,%d,%d), want (0,0,10,10)", x, y, w, h)
	}

	ss.Pop()
	if got := ss.Depth(); got != 0 {
		t.Fatalf("Depth() after popping all = %d, want 0", got)
	}
	x, y, w, h = ss.Current()
	if x != 0 || y != 0 || w != 0 || h != 0 {
		t.Fatalf("Current() after emptying stack = (%d,%d,%d,%d), want zeros", x, y, w, h)
	}
}
