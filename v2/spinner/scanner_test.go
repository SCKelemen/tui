package spinner

import (
	"strings"
	"testing"
)

func TestNewScannerCreatesValidDefaultScanner(t *testing.T) {
	s := NewScanner(ScannerConfig{})
	if s == nil {
		t.Fatal("NewScanner returned nil")
	}

	if s.cfg.Width != 10 {
		t.Fatalf("expected default width 10, got %d", s.cfg.Width)
	}
	if s.cfg.TrailLength != 3 {
		t.Fatalf("expected default trail length 3, got %d", s.cfg.TrailLength)
	}
	if s.cfg.ActiveColor != "#FFFFFF" {
		t.Fatalf("expected default active color #FFFFFF, got %q", s.cfg.ActiveColor)
	}
	if !s.cfg.FadeTrail {
		t.Fatalf("expected default FadeTrail=true for zero config")
	}
}

func TestScannerViewRendersAndTickChangesFrame(t *testing.T) {
	s := NewScanner(ScannerConfig{
		Width:       6,
		Style:       ScannerDiamonds,
		Direction:   ScannerForward,
		ActiveColor: "#FF0000",
		TrailLength: 2,
		FadeTrail:   true,
	})

	before := s.View()
	if before == "" {
		t.Fatal("expected non-empty scanner view")
	}
	if !strings.Contains(before, "◆") {
		t.Fatalf("expected active diamond glyph in view, got %q", before)
	}

	initialFrame := s.frame
	s.Tick()
	after := s.View()

	if s.frame != initialFrame+1 {
		t.Fatalf("expected frame to increment by 1, got %d -> %d", initialFrame, s.frame)
	}
	if after == "" {
		t.Fatal("expected non-empty scanner view after tick")
	}
}
