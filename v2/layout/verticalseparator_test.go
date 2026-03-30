package layout

import (
	"strings"
	"testing"

	"github.com/SCKelemen/tui/v2/style"
)

func TestVerticalSeparatorRendersExpectedLines(t *testing.T) {
	sep := NewVerticalSeparator(3, WithVerticalSeparatorColor(""))
	got := sep.View()

	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), got)
	}

	for i, line := range lines {
		if line != "│" {
			t.Fatalf("line %d = %q, want %q", i, line, "│")
		}
	}
}

func TestVerticalSeparatorOptionsCharAndColor(t *testing.T) {
	sep := NewVerticalSeparator(
		2,
		WithVerticalSeparatorChar("#"),
		WithVerticalSeparatorColor("#00FF00"),
	)

	got := sep.View()
	fg := style.ANSIColorFromHex("#00FF00")
	if !strings.Contains(got, fg) {
		t.Fatalf("expected separator to include ANSI fg color %q, got %q", fg, got)
	}

	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), got)
	}

	for i, line := range lines {
		if !strings.Contains(line, "#") {
			t.Fatalf("line %d missing custom char: %q", i, line)
		}
		if !strings.Contains(line, style.ANSIReset) {
			t.Fatalf("line %d missing ANSI reset: %q", i, line)
		}
	}
}
