package tui

import (
	"strings"
	"testing"
)

func TestNewScrollBar(t *testing.T) {
	sb := NewScrollBar()
	if sb == nil {
		t.Fatal("NewScrollBar returned nil")
	}
	if sb.orientation != ScrollBarVertical {
		t.Fatalf("orientation = %v, want vertical", sb.orientation)
	}
	if sb.total != 1 || sb.visible != 1 || sb.offset != 0 {
		t.Fatalf("unexpected default metrics: total=%d visible=%d offset=%d", sb.total, sb.visible, sb.offset)
	}
}

func TestScrollBarNew(t *testing.T) {
	TestNewScrollBar(t)
}

func TestSetPosition(t *testing.T) {
	sb := NewScrollBar(WithScrollBarMetrics(100, 20, 0))
	sb.SetPosition(500)
	if sb.offset != 80 {
		t.Fatalf("SetPosition() offset = %d, want 80", sb.offset)
	}
	sb.SetPosition(-5)
	if sb.offset != 0 {
		t.Fatalf("SetPosition(-5) offset = %d, want 0", sb.offset)
	}
}

func TestSetTotal(t *testing.T) {
	sb := NewScrollBar(WithScrollBarMetrics(10, 5, 4))
	sb.SetTotal(3)
	if sb.total != 3 || sb.visible != 3 || sb.offset != 0 {
		t.Fatalf("after SetTotal(3): total=%d visible=%d offset=%d", sb.total, sb.visible, sb.offset)
	}
}

func TestSetVisible(t *testing.T) {
	sb := NewScrollBar(WithScrollBarMetrics(10, 3, 8))
	sb.SetVisible(20)
	if sb.visible != 10 || sb.offset != 0 {
		t.Fatalf("after SetVisible(20): visible=%d offset=%d", sb.visible, sb.offset)
	}
}

func TestVerticalView(t *testing.T) {
	sb := NewScrollBar(
		WithScrollBarOrientation(ScrollBarVertical),
		WithScrollBarSize(1, 6),
		WithScrollBarMetrics(100, 20, 40),
	)

	view := stripANSI(sb.View())
	lines := strings.Split(view, "\n")
	if len(lines) != 6 {
		t.Fatalf("vertical line count = %d, want 6", len(lines))
	}
	if lines[0] != "▲" || lines[len(lines)-1] != "▼" {
		t.Fatalf("vertical arrows = [%q ... %q], want [▲ ... ▼]", lines[0], lines[len(lines)-1])
	}
	if !strings.Contains(view, "█") {
		t.Fatalf("vertical view missing thumb: %q", view)
	}
}

func TestHorizontalView(t *testing.T) {
	sb := NewScrollBar(
		WithScrollBarOrientation(ScrollBarHorizontal),
		WithScrollBarSize(8, 1),
		WithScrollBarMetrics(100, 25, 25),
	)

	view := stripANSI(sb.View())
	if !strings.HasPrefix(view, "◀") || !strings.HasSuffix(view, "▶") {
		t.Fatalf("horizontal view = %q, want ◀...▶", view)
	}
	if !strings.Contains(view, "█") {
		t.Fatalf("horizontal view missing thumb: %q", view)
	}
}
