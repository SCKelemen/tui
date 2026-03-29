package display

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var bigTextANSIPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripBigTextANSI(s string) string {
	return bigTextANSIPattern.ReplaceAllString(s, "")
}

func forceTrueColor(t *testing.T) {
	t.Helper()
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(prev)
	})
}

func TestBigTextView(t *testing.T) {
	out := NewBigText("AB", WithBigTextColor("")).View()
	if out == "" {
		t.Fatal("expected non-empty output")
	}

	lines := strings.Split(out, "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
}

func TestBigTextEmpty(t *testing.T) {
	out := NewBigText("").View()
	if out != "" {
		t.Fatalf("expected empty output, got %q", out)
	}
}

func TestBigTextColor(t *testing.T) {
	forceTrueColor(t)

	out := NewBigText("A", WithBigTextColor("#FF0000")).View()
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected ANSI escape sequences in output, got %q", out)
	}
}

func TestBigTextGradient(t *testing.T) {
	forceTrueColor(t)

	out := NewBigText("AB", WithBigTextGradient([]string{"#FF0000", "#0000FF"})).View()
	matches := bigTextANSIPattern.FindAllString(out, -1)

	seen := map[string]struct{}{}
	for _, m := range matches {
		if strings.Contains(m, "38;") {
			seen[m] = struct{}{}
		}
	}

	if len(seen) < 2 {
		t.Fatalf("expected at least 2 distinct color sequences for gradient, got %d in %q", len(seen), out)
	}
}

func TestBigTextWidth(t *testing.T) {
	const width = 40
	out := NewBigText("HI", WithBigTextColor(""), WithBigTextWidth(width)).View()
	lines := strings.Split(out, "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}

	for i, line := range lines {
		if got := lipgloss.Width(stripBigTextANSI(line)); got != width {
			t.Fatalf("line %d width=%d, want %d; line=%q", i, got, width, line)
		}
	}
}

func TestBigTextSpecialChars(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("expected no panic, got %v", r)
		}
	}()

	out := NewBigText("!@#-.,/:\"", WithBigTextColor("")).View()
	if out == "" {
		t.Fatal("expected non-empty output for punctuation")
	}

	lines := strings.Split(out, "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
}

func TestBigTextUnknownChar(t *testing.T) {
	unknown := NewBigText("€", WithBigTextColor("")).View()
	fallback := NewBigText("?", WithBigTextColor("")).View()

	if unknown != fallback {
		t.Fatalf("expected unknown character to fallback to '?' glyph\nunknown:\n%q\nfallback:\n%q", unknown, fallback)
	}
}
