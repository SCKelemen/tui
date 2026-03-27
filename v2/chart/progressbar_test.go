package chart

import (
	"regexp"
	"strings"
	"testing"

	colorpkg "github.com/SCKelemen/color"
	"github.com/SCKelemen/tui/v2/style"
)

func TestProgressBarCreation(t *testing.T) {
	pb := NewProgressBar(0.5)
	if pb == nil {
		t.Fatal("NewProgressBar returned nil")
	}

	if pb.width != 30 {
		t.Errorf("expected default width=30, got %d", pb.width)
	}

	if pb.progress != 0.5 {
		t.Errorf("expected progress=0.5, got %f", pb.progress)
	}

	if !pb.showPercent {
		t.Error("showPercent should default to true")
	}

	if pb.style != ProgressBarBlock {
		t.Errorf("expected default style ProgressBarBlock, got %v", pb.style)
	}

	if pb.filledChar != '█' || pb.emptyChar != '░' {
		t.Errorf("unexpected default chars: filled=%q empty=%q", pb.filledChar, pb.emptyChar)
	}
}

func TestProgressBarOptions(t *testing.T) {
	pb := NewProgressBar(0.25,
		WithProgressBarWidth(20),
		WithProgressBarLabel("Compiling"),
		WithProgressBarShowPercent(false),
		WithProgressBarStyle(ProgressBarTick),
		WithProgressBarFilledChar('X'),
		WithProgressBarEmptyChar('_'),
	)

	if pb.width != 20 {
		t.Errorf("expected width=20, got %d", pb.width)
	}
	if pb.label != "Compiling" {
		t.Errorf("expected label Compiling, got %q", pb.label)
	}
	if pb.showPercent {
		t.Error("expected showPercent=false")
	}
	if pb.style != ProgressBarTick {
		t.Errorf("expected ProgressBarTick, got %v", pb.style)
	}
	if pb.filledChar != 'X' || pb.emptyChar != '_' {
		t.Errorf("expected custom chars X/_, got %q/%q", pb.filledChar, pb.emptyChar)
	}
}

func TestProgressBarSetProgressClamps(t *testing.T) {
	pb := NewProgressBar(0)

	pb.SetProgress(-0.5)
	if pb.GetProgress() != 0 {
		t.Errorf("expected clamp to 0, got %f", pb.GetProgress())
	}

	pb.SetProgress(2.0)
	if pb.GetProgress() != 1 {
		t.Errorf("expected clamp to 1, got %f", pb.GetProgress())
	}
}

func TestProgressBarIncrement(t *testing.T) {
	pb := NewProgressBar(0.2)

	pb.Increment(0.3)
	if pb.GetProgress() != 0.5 {
		t.Errorf("expected progress 0.5, got %f", pb.GetProgress())
	}

	pb.Increment(1.0)
	if pb.GetProgress() != 1.0 {
		t.Errorf("expected progress 1.0 after clamp, got %f", pb.GetProgress())
	}
}

func TestProgressBarStylesRender(t *testing.T) {
	tests := []struct {
		name        string
		style       ProgressBarStyle
		filledChar  string
		emptyChar   string
		hasBrackets bool
	}{
		{name: "block", style: ProgressBarBlock, filledChar: "█", emptyChar: "░"},
		{name: "tick", style: ProgressBarTick, filledChar: "▰", emptyChar: "▱"},
		{name: "classic", style: ProgressBarClassic, filledChar: "=", emptyChar: " ", hasBrackets: true},
		{name: "thin", style: ProgressBarThin, filledChar: "━", emptyChar: "╌"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewProgressBar(0.5,
				WithProgressBarWidth(10),
				WithProgressBarStyle(tt.style),
			)
			view := pb.View()
			plain := stripANSI(view)

			if tt.hasBrackets {
				if !strings.Contains(plain, "[") || !strings.Contains(plain, "]") {
					t.Fatalf("classic style should include brackets, got %q", plain)
				}
			}

			if !strings.Contains(plain, tt.filledChar) {
				t.Fatalf("expected filled char %q in %q", tt.filledChar, plain)
			}
			if !strings.Contains(plain, tt.emptyChar) {
				t.Fatalf("expected empty char %q in %q", tt.emptyChar, plain)
			}
		})
	}
}

func TestProgressBarGradientRendering(t *testing.T) {
	start := colorpkg.NewRGBA(1, 0, 0, 1)
	end := colorpkg.NewRGBA(0, 0, 1, 1)

	pb := NewProgressBar(0.6,
		WithProgressBarWidth(10),
		WithProgressBarGradient(start, end),
	)

	view := pb.View()
	if !strings.Contains(view, "\x1b[38;2;") {
		t.Fatalf("expected ANSI truecolor sequence in gradient render, got %q", view)
	}
}

func TestProgressBarMultiStopGradient(t *testing.T) {
	stops := []colorpkg.GradientStop{
		{Color: colorpkg.NewRGBA(1, 0, 0, 1), Position: 0.0},
		{Color: colorpkg.NewRGBA(0, 1, 0, 1), Position: 0.5},
		{Color: colorpkg.NewRGBA(0, 0, 1, 1), Position: 1.0},
	}

	pb := NewProgressBar(0.8,
		WithProgressBarWidth(10),
		WithProgressBarMultiStopGradient(stops),
	)

	view := pb.View()
	if !strings.Contains(view, "\x1b[38;2;") {
		t.Fatalf("expected ANSI truecolor sequence in multi-stop gradient render, got %q", view)
	}
}

func TestProgressBarNonGradientRendering(t *testing.T) {
	pb := NewProgressBar(0.5, WithProgressBarWidth(10))
	view := pb.View()

	if !strings.Contains(view, "\x1b[") {
		t.Fatalf("expected ANSI styling in output, got %q", view)
	}
	if !strings.Contains(view, style.ANSIDim) {
		t.Fatalf("expected dim empty segment style, got %q", view)
	}
}

func TestProgressBarPercentDisplay(t *testing.T) {
	pb := NewProgressBar(0.67, WithProgressBarWidth(10), WithProgressBarShowPercent(true))
	plain := stripANSI(pb.View())
	if !strings.Contains(plain, "67%") {
		t.Fatalf("expected percent in output, got %q", plain)
	}

	pb2 := NewProgressBar(0.67, WithProgressBarWidth(10), WithProgressBarShowPercent(false))
	plain2 := stripANSI(pb2.View())
	if strings.Contains(plain2, "%") {
		t.Fatalf("expected no percent in output, got %q", plain2)
	}
}

func TestProgressBarLabelDisplay(t *testing.T) {
	pb := NewProgressBar(0.4,
		WithProgressBarWidth(10),
		WithProgressBarLabel("Building..."),
	)

	plain := stripANSI(pb.View())
	if !strings.Contains(plain, "Building...") {
		t.Fatalf("expected label in output, got %q", plain)
	}
}

func TestProgressBarEdgeCases(t *testing.T) {
	zero := NewProgressBar(0,
		WithProgressBarWidth(10),
		WithProgressBarShowPercent(true),
	)
	plainZero := stripANSI(zero.View())
	if !strings.Contains(plainZero, "0%") {
		t.Fatalf("expected 0%%, got %q", plainZero)
	}

	full := NewProgressBar(1,
		WithProgressBarWidth(10),
		WithProgressBarShowPercent(true),
	)
	plainFull := stripANSI(full.View())
	if !strings.Contains(plainFull, "100%") {
		t.Fatalf("expected 100%%, got %q", plainFull)
	}
	if strings.Contains(plainFull, "░") || strings.Contains(plainFull, "▱") || strings.Contains(plainFull, "╌") {
		t.Fatalf("full progress should not contain default empty chars, got %q", plainFull)
	}
}

func TestProgressBarSetLabel(t *testing.T) {
	pb := NewProgressBar(0.5)
	pb.SetLabel("Compiling")

	if pb.label != "Compiling" {
		t.Fatalf("expected label Compiling, got %q", pb.label)
	}
}

func TestProgressBarFocusManagement(t *testing.T) {
	pb := NewProgressBar(0.5)
	if pb.Focused() {
		t.Fatal("expected unfocused by default")
	}

	pb.Focus()
	if !pb.Focused() {
		t.Fatal("expected focused after Focus")
	}

	pb.Blur()
	if pb.Focused() {
		t.Fatal("expected unfocused after Blur")
	}
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}
