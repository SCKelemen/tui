package layout

import (
	"strings"
	"testing"

	"github.com/SCKelemen/tui/v2/style"
)

func TestHalfLinePad(t *testing.T) {
	got := HalfLinePad("content")
	want := "\ncontent\n"
	if got != want {
		t.Fatalf("HalfLinePad() = %q, want %q", got, want)
	}
}

func TestHalfLinePaddedBoxRendersWithPadding(t *testing.T) {
	box := NewHalfLinePaddedBox(
		"abc",
		WithHalfLineBoxPaddingTop(1),
		WithHalfLineBoxPaddingBottom(1),
		WithHalfLineBoxBackground(""),
	)

	got := box.View()
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), got)
	}

	if lines[0] != "   " {
		t.Fatalf("top padding line = %q, want %q", lines[0], "   ")
	}
	if lines[1] != "abc" {
		t.Fatalf("content line = %q, want %q", lines[1], "abc")
	}
	if lines[2] != "   " {
		t.Fatalf("bottom padding line = %q, want %q", lines[2], "   ")
	}
}

func TestHalfLinePaddedBoxWidthAndBackgroundOptions(t *testing.T) {
	box := NewHalfLinePaddedBox(
		"abc",
		WithHalfLineBoxPaddingTop(1),
		WithHalfLineBoxPaddingBottom(1),
		WithHalfLineBoxWidth(5),
		WithHalfLineBoxBackground("#112233"),
	)

	got := box.View()
	bg := style.ANSIBackgroundColorFromHex("#112233")

	if !strings.Contains(got, bg) {
		t.Fatalf("expected view to contain background ANSI sequence %q, got %q", bg, got)
	}

	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), got)
	}

	for i, line := range lines {
		if !strings.Contains(line, style.ANSIReset) {
			t.Fatalf("line %d should contain ANSI reset, got %q", i, line)
		}
	}

	if !strings.Contains(lines[1], "abc  ") {
		t.Fatalf("content line should be padded to width 5, got %q", lines[1])
	}
}
