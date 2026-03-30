package display

import (
	"strings"
	"testing"
)

func TestGutterRendererWithDecorations(t *testing.T) {
	decs := []GutterDecoration{{Line: 2, Mark: GutterError}, {Line: 3, Mark: GutterGitAdded}}
	g := NewGutterRenderer(4, decs)

	if g == nil {
		t.Fatal("NewGutterRenderer returned nil")
	}
	if g.maxLine != 4 {
		t.Fatalf("expected maxLine=4, got %d", g.maxLine)
	}
	if len(g.decorations) != 2 {
		t.Fatalf("expected 2 decorations, got %d", len(g.decorations))
	}
}

func TestGutterViewRendersMarkers(t *testing.T) {
	decs := []GutterDecoration{
		{Line: 1, Mark: GutterCurrentLine},
		{Line: 2, Mark: GutterWarning},
		{Line: 4, Mark: GutterBreakpoint},
	}
	g := NewGutterRenderer(4, decs)

	plain := stripANSI(g.View())
	if !strings.Contains(plain, "▶") {
		t.Fatalf("expected current-line marker in view, got:\n%s", plain)
	}
	if !strings.Contains(plain, "▲") {
		t.Fatalf("expected warning marker in view, got:\n%s", plain)
	}
	if !strings.Contains(plain, "⏺") {
		t.Fatalf("expected breakpoint marker in view, got:\n%s", plain)
	}
}
