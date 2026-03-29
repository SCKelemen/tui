package display

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func forceStatusBarTrueColor(t *testing.T) {
	t.Helper()
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(prev)
	})
}

func TestStatusBarView(t *testing.T) {
	bar := NewStatusBar(
		WithStatusBarWidth(20),
		WithStatusBarLeft(StatusBarSection{Text: "left"}),
	)

	view := bar.View()
	if view == "" {
		t.Fatal("expected non-empty status bar view")
	}

	plain := stripANSI(view)
	if got := lipgloss.Width(plain); got != 20 {
		t.Fatalf("expected rendered width 20, got %d (%q)", got, plain)
	}
	if strings.Contains(plain, "\n") {
		t.Fatalf("expected single-line bar, got %q", plain)
	}
}

func TestStatusBarLeftCenterRight(t *testing.T) {
	bar := NewStatusBar(
		WithStatusBarWidth(30),
		WithStatusBarLeft(StatusBarSection{Text: "L"}),
		WithStatusBarCenter(StatusBarSection{Text: "C"}),
		WithStatusBarRight(StatusBarSection{Text: "R"}),
	)

	plain := stripANSI(bar.View())
	if got := lipgloss.Width(plain); got != 30 {
		t.Fatalf("expected rendered width 30, got %d", got)
	}

	if !strings.HasPrefix(plain, "L") {
		t.Fatalf("expected left section at start, got %q", plain)
	}
	if !strings.HasSuffix(plain, "R") {
		t.Fatalf("expected right section at end, got %q", plain)
	}

	l := strings.Index(plain, "L")
	c := strings.Index(plain, "C")
	r := strings.LastIndex(plain, "R")
	if c == -1 || !(l < c && c < r) {
		t.Fatalf("expected center section between left and right, got %q", plain)
	}
}

func TestStatusBarWidth(t *testing.T) {
	bar := NewStatusBar(
		WithStatusBarWidth(10),
		WithStatusBarLeft(StatusBarSection{Text: "left"}),
	)

	narrow := stripANSI(bar.View())
	if got := lipgloss.Width(narrow); got != 10 {
		t.Fatalf("expected narrow width 10, got %d", got)
	}

	bar.SetWidth(20)
	wide := stripANSI(bar.View())
	if got := lipgloss.Width(wide); got != 20 {
		t.Fatalf("expected wide width 20, got %d", got)
	}

	if narrow == wide {
		t.Fatalf("expected different output for different widths, both were %q", narrow)
	}
}

func TestStatusBarSetters(t *testing.T) {
	bar := NewStatusBar(WithStatusBarWidth(30))

	bar.SetLeft(StatusBarSection{Text: "LEFT"})
	bar.SetCenter(StatusBarSection{Text: "CENTER"})
	bar.SetRight(StatusBarSection{Text: "RIGHT"})

	first := stripANSI(bar.View())
	if !strings.Contains(first, "LEFT") || !strings.Contains(first, "CENTER") || !strings.Contains(first, "RIGHT") {
		t.Fatalf("expected all sections after initial setters, got %q", first)
	}

	bar.SetLeft(StatusBarSection{Text: "L2"})
	bar.SetCenter(StatusBarSection{Text: "C2"})
	bar.SetRight(StatusBarSection{Text: "R2"})

	second := stripANSI(bar.View())
	if !strings.Contains(second, "L2") || !strings.Contains(second, "C2") || !strings.Contains(second, "R2") {
		t.Fatalf("expected updated sections after setters, got %q", second)
	}
	if strings.Contains(second, "LEFT") || strings.Contains(second, "CENTER") || strings.Contains(second, "RIGHT") {
		t.Fatalf("expected old section values to be replaced, got %q", second)
	}
}

func TestStatusBarBgColor(t *testing.T) {
	forceStatusBarTrueColor(t)

	bar := NewStatusBar(
		WithStatusBarWidth(12),
		WithStatusBarBg("#112233"),
		WithStatusBarLeft(StatusBarSection{Text: "bg"}),
	)

	view := bar.View()
	if !strings.Contains(view, "48;2;17;34;51") {
		t.Fatalf("expected custom background ANSI code in view, got %q", view)
	}
}

func TestStatusBarEmpty(t *testing.T) {
	forceStatusBarTrueColor(t)

	bar := NewStatusBar(
		WithStatusBarWidth(8),
		WithStatusBarBg("#123456"),
	)

	view := bar.View()
	plain := stripANSI(view)
	if plain != "        " {
		t.Fatalf("expected empty status bar to render spaces, got %q", plain)
	}
	if !strings.Contains(view, "48;2;18;52;86") {
		t.Fatalf("expected background color ANSI code in empty bar, got %q", view)
	}
}

func TestStatusBarBoldSection(t *testing.T) {
	forceStatusBarTrueColor(t)

	bar := NewStatusBar(
		WithStatusBarBg("#101010"),
	)

	rendered := bar.renderSections([]StatusBarSection{{Text: "BOLD", Bold: true}})
	if !strings.Contains(rendered, "BOLD") {
		t.Fatalf("expected bold section text in rendered output, got %q", rendered)
	}
	if !strings.Contains(rendered, "\x1b[1;") && !strings.Contains(rendered, "\x1b[1m") && !strings.Contains(rendered, ";1m") {
		t.Fatalf("expected ANSI bold code in rendered output, got %q", rendered)
	}
}
