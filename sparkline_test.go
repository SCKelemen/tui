package tui

import (
	"strings"
	"testing"
)

func TestSparklineCreation(t *testing.T) {
	s := NewSparkline([]float64{1, 2, 3})
	if s == nil {
		t.Fatal("NewSparkline returned nil")
	}

	if s.width != 40 {
		t.Errorf("expected default width=40, got %d", s.width)
	}

	if !s.autoRange {
		t.Error("expected autoRange=true by default")
	}
}

func TestSparklineOptions(t *testing.T) {
	s := NewSparkline([]float64{1, 2, 3},
		WithSparklineWidth(8),
		WithSparklineLabel("CPU"),
		WithSparklineRange(0, 10),
	)

	if s.width != 8 {
		t.Errorf("expected width=8, got %d", s.width)
	}
	if s.label != "CPU" {
		t.Errorf("expected label CPU, got %q", s.label)
	}
	if s.autoRange {
		t.Error("expected autoRange=false after fixed range option")
	}
	if s.minVal != 0 || s.maxVal != 10 {
		t.Errorf("expected range 0..10, got %f..%f", s.minVal, s.maxVal)
	}
}

func TestSparklineView(t *testing.T) {
	s := NewSparkline([]float64{0, 1, 2, 3, 4, 5, 6, 7},
		WithSparklineWidth(8),
		WithSparklineLabel("trend"),
	)

	view := s.View()
	if !strings.HasSuffix(view, " trend") {
		t.Fatalf("expected label suffix, got %q", view)
	}

	spark := strings.TrimSuffix(view, " trend")
	if len([]rune(spark)) != 8 {
		t.Fatalf("expected 8 sparkline chars, got %d in %q", len([]rune(spark)), spark)
	}

	if !strings.Contains(spark, "▁") || !strings.Contains(spark, "█") {
		t.Fatalf("expected low and high block chars, got %q", spark)
	}
}

func TestSparklineViewUsesLastWidthDataPoints(t *testing.T) {
	s := NewSparkline([]float64{1, 2, 3, 4, 5, 6},
		WithSparklineWidth(3),
	)

	view := s.View()
	if len([]rune(view)) != 3 {
		t.Fatalf("expected width-limited output of 3 chars, got %q", view)
	}

	if view != "▁▅█" {
		t.Fatalf("expected tail series mapping \"▁▅█\", got %q", view)
	}
}

func TestSparklineFixedRange(t *testing.T) {
	s := NewSparkline([]float64{10, 20, 30},
		WithSparklineWidth(3),
		WithSparklineRange(0, 40),
	)

	view := s.View()
	if view != "▃▅▆" {
		t.Fatalf("expected fixed-range mapping \"▃▅▆\", got %q", view)
	}
}

func TestSparklineFocusManagement(t *testing.T) {
	s := NewSparkline(nil)
	if s.Focused() {
		t.Fatal("expected unfocused by default")
	}

	s.Focus()
	if !s.Focused() {
		t.Fatal("expected focused after Focus")
	}

	s.Blur()
	if s.Focused() {
		t.Fatal("expected unfocused after Blur")
	}
}
