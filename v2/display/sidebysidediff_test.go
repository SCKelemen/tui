package display

import (
	"strings"
	"testing"
)

func TestSideBySideDiffConstructorWithHunks(t *testing.T) {
	oldCode := "line1\nline2\nline3"
	newCode := "line1\nline2 changed\nline3\nline4"

	diff := NewSideBySideDiff(oldCode, newCode, WithSideBySideDiffLabels("old.go", "new.go"), WithSideBySideDiffVisibleLines(8))
	if diff == nil {
		t.Fatal("NewSideBySideDiff returned nil")
	}
	if len(diff.lineChanges) == 0 {
		t.Fatal("expected computed line changes")
	}
	if diff.visibleLines != 8 {
		t.Fatalf("expected visibleLines=8, got %d", diff.visibleLines)
	}
}

func TestSideBySideDiffRender(t *testing.T) {
	oldCode := "func main() {\n\tprintln(\"old\")\n}"
	newCode := "func main() {\n\tprintln(\"new\")\n}"

	diff := NewSideBySideDiff(oldCode, newCode, WithSideBySideDiffLabels("old.go", "new.go"), WithSideBySideDiffWidth(100), WithSideBySideDiffVisibleLines(6))
	plain := stripANSI(diff.View())

	if strings.TrimSpace(plain) == "" {
		t.Fatal("expected non-empty side-by-side view")
	}
	if !strings.Contains(plain, "old.go") || !strings.Contains(plain, "new.go") {
		t.Fatalf("expected pane labels in view, got:\n%s", plain)
	}
	if !strings.Contains(plain, "println(\"old\")") || !strings.Contains(plain, "println(\"new\")") {
		t.Fatalf("expected old/new code content in view, got:\n%s", plain)
	}
	if !strings.Contains(plain, "~") {
		t.Fatalf("expected modified hunk indicator in view, got:\n%s", plain)
	}
}
