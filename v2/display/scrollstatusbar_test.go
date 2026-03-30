package display

import (
	"strings"
	"testing"
)

func TestScrollStatusConstructor(t *testing.T) {
	bar := NewScrollStatusBar()
	if bar == nil {
		t.Fatal("expected non-nil ScrollStatusBar")
	}
	if bar.width != 80 {
		t.Fatalf("expected default width 80, got %d", bar.width)
	}
}

func TestScrollStatusPositionTracking(t *testing.T) {
	bar := NewScrollStatusBar(
		WithScrollStatusBarWidth(40),
		WithScrollStatusBarTotalLines(100),
		WithScrollStatusBarVisibleStart(10),
		WithScrollStatusBarVisibleEnd(20),
	)

	plain := stripANSI(bar.View())
	if !strings.Contains(plain, "Lines 10-20 of 100") {
		t.Fatalf("expected initial position text, got %q", plain)
	}

	bar.SetPosition(25, 30, 120)
	plain = stripANSI(bar.View())
	if !strings.Contains(plain, "Lines 25-30 of 120") {
		t.Fatalf("expected updated position text, got %q", plain)
	}
}

func TestScrollStatusViewNonEmpty(t *testing.T) {
	bar := NewScrollStatusBar(
		WithScrollStatusBarWidth(30),
		WithScrollStatusBarTotalLines(10),
		WithScrollStatusBarVisibleStart(1),
		WithScrollStatusBarVisibleEnd(5),
	)

	if got := bar.View(); got == "" {
		t.Fatal("expected non-empty scroll status view")
	}
}
