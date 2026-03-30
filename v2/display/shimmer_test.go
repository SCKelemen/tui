package display

import (
	"strings"
	"testing"

	"github.com/SCKelemen/tui/v2/style"
)

func TestShimmerConstructor(t *testing.T) {
	s := NewShimmer()
	if s == nil {
		t.Fatal("expected non-nil Shimmer")
	}
	if s.width != defaultShimmerWidth {
		t.Fatalf("expected default width %d, got %d", defaultShimmerWidth, s.width)
	}
	if s.lines != defaultShimmerLines {
		t.Fatalf("expected default lines %d, got %d", defaultShimmerLines, s.lines)
	}
}

func TestShimmerLinesAndWidthOptions(t *testing.T) {
	s := NewShimmer(
		WithShimmerWidth(5),
		WithShimmerLines(2),
	)

	plain := stripANSI(s.View())
	lines := strings.Split(plain, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 shimmer lines, got %d in %q", len(lines), plain)
	}

	for i, line := range lines {
		if w := style.StringWidth(line); w != 5 {
			t.Fatalf("expected line %d width 5, got %d (%q)", i, w, line)
		}
	}
}

func TestShimmerViewNonEmpty(t *testing.T) {
	s := NewShimmer(WithShimmerWidth(3), WithShimmerLines(1))
	if got := s.View(); got == "" {
		t.Fatal("expected non-empty shimmer view")
	}
}
