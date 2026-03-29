package display

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var gradientANSIPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripGradientANSI(s string) string {
	return gradientANSIPattern.ReplaceAllString(s, "")
}

func forceGradientTrueColor(t *testing.T) {
	t.Helper()
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(prev)
	})
}

func TestGradientTextView(t *testing.T) {
	forceGradientTrueColor(t)

	g := NewGradientText("hello", []string{"#FF0000", "#00FF00"})
	out := g.View()

	if out == "" {
		t.Fatal("expected non-empty output")
	}

	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected ANSI-colored output, got %q", out)
	}
}

func TestGradientTextEmpty(t *testing.T) {
	g := NewGradientText("", []string{"#FF0000", "#00FF00"})
	if got := g.View(); got != "" {
		t.Fatalf("expected empty output, got %q", got)
	}
}

func TestRenderGradient(t *testing.T) {
	forceGradientTrueColor(t)

	out := RenderGradient("gradient", []string{"#FF0000", "#0000FF"})
	if out == "" {
		t.Fatal("expected non-empty output")
	}

	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected ANSI-colored output, got %q", out)
	}
}

func TestLovableGradient(t *testing.T) {
	forceGradientTrueColor(t)

	text := "lovable"
	out := LovableGradient(text)
	if out == "" {
		t.Fatal("expected non-empty output")
	}

	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected ANSI-colored output, got %q", out)
	}
}

func TestGradientTextWidth(t *testing.T) {
	g := NewGradientText("abcdef", []string{"#FF0000", "#00FF00"})
	WithGradientTextWidth(3)(g)

	out := g.View()
	plain := stripGradientANSI(out)

	if plain != "abc" {
		t.Fatalf("expected truncated output %q, got %q", "abc", plain)
	}
}

func TestGradientSingleStop(t *testing.T) {
	forceGradientTrueColor(t)

	out := RenderGradient("single", []string{"#112233"})
	if out == "" {
		t.Fatal("expected non-empty output")
	}

	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected ANSI-colored output, got %q", out)
	}
}

func TestGradientInvalidStop(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RenderGradient panicked for invalid stop: %v", r)
		}
	}()

	text := "fallback"
	out := RenderGradient(text, []string{"#ZZZZZZ", "#00FF00"})
	if out != text {
		t.Fatalf("expected graceful fallback to original text %q, got %q", text, out)
	}
}
