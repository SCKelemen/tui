package display

import (
	"strings"
	"testing"
)

func TestPeekViewConstructor(t *testing.T) {
	files := []PeekFile{{Path: "main.go", Language: "go", Content: "package main", StartLine: 1}}
	peek := NewPeekView("Definition", files, WithPeekViewWidth(70), WithPeekViewVisibleLines(6))

	if peek == nil {
		t.Fatal("NewPeekView returned nil")
	}
	if peek.width != 70 {
		t.Fatalf("expected width=70, got %d", peek.width)
	}
	if peek.visibleLines != 6 {
		t.Fatalf("expected visibleLines=6, got %d", peek.visibleLines)
	}
	if len(peek.files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(peek.files))
	}
}

func TestPeekViewRendersCodeAndReferences(t *testing.T) {
	files := []PeekFile{
		{Path: "a/main.go", Language: "go", Content: "package main\nfunc main() {}", StartLine: 1},
		{Path: "b/util.go", Language: "go", Content: "package b\nfunc Util() {}", StartLine: 10},
	}
	peek := NewPeekView("Peek References", files, WithPeekViewWidth(90), WithPeekViewVisibleLines(5))

	plain := stripANSI(peek.View())
	if !strings.Contains(plain, "Peek References") {
		t.Fatalf("expected title in view, got:\n%s", plain)
	}
	if !strings.Contains(plain, "main.go") || !strings.Contains(plain, "util.go") {
		t.Fatalf("expected tab labels (references) in view, got:\n%s", plain)
	}
	if !strings.Contains(plain, "func main() {}") {
		t.Fatalf("expected code in view, got:\n%s", plain)
	}
}
