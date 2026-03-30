package display

import (
	"strings"
	"testing"
)

func TestRatingConstructor(t *testing.T) {
	r := NewRatingDots(3, 5, WithRatingDotsLabel("Score"))
	if r == nil {
		t.Fatal("expected non-nil RatingDots")
	}
	if r.value != 3 || r.max != 5 {
		t.Fatalf("expected value/max 3/5, got %d/%d", r.value, r.max)
	}
	if r.label != "Score" {
		t.Fatalf("expected label Score, got %q", r.label)
	}
}

func TestRatingFilledAndEmptyDots(t *testing.T) {
	r := NewRatingDots(3, 5)
	plain := stripANSI(r.View())

	if got := strings.Count(plain, "●"); got != 3 {
		t.Fatalf("expected 3 filled dots, got %d in %q", got, plain)
	}
	if got := strings.Count(plain, "○"); got != 2 {
		t.Fatalf("expected 2 empty dots, got %d in %q", got, plain)
	}
}

func TestRatingViewNonEmpty(t *testing.T) {
	r := NewRatingDots(1, 1)
	if got := r.View(); got == "" {
		t.Fatal("expected non-empty rating view")
	}
}
