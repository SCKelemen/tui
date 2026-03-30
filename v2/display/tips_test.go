package display

import (
	"strings"
	"testing"
)

func TestTipsConstructorWithTips(t *testing.T) {
	tips := []Tip{{Text: "Use {highlight}go test ./...{/highlight} often."}}
	v := NewTips(tips, WithTipsHighlightColor("196"), WithTipsDimColor("245"), WithTipsWidth(80))
	if v == nil {
		t.Fatal("NewTips returned nil")
	}
	if len(v.tips) != 1 {
		t.Fatalf("expected 1 tip, got %d", len(v.tips))
	}
	if v.currentIndex != 0 {
		t.Fatalf("expected currentIndex=0, got %d", v.currentIndex)
	}
	if v.highlightColor != "196" {
		t.Fatalf("expected highlightColor=196, got %q", v.highlightColor)
	}
	if v.width != 80 {
		t.Fatalf("expected width=80, got %d", v.width)
	}
}

func TestTipsViewRendersTipContent(t *testing.T) {
	v := NewTips([]Tip{{Text: "Use {highlight}go test -run TestName -v{/highlight} to iterate."}})
	view := stripANSI(v.View())
	if strings.TrimSpace(view) == "" {
		t.Fatal("expected non-empty tips view")
	}
	if !strings.Contains(view, "Tip:") {
		t.Fatalf("expected Tip prefix, got %q", view)
	}
	if !strings.Contains(view, "go test -run TestName -v") {
		t.Fatalf("expected highlighted tip content, got %q", view)
	}
	if strings.Contains(view, "{highlight}") || strings.Contains(view, "{/highlight}") {
		t.Fatalf("expected markup tags to be removed, got %q", view)
	}
}
