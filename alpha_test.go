package tui

import (
	"math"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRGBAFromHex(t *testing.T) {
	tests := []struct {
		name string
		hex  string
		want RGBA
	}{
		{
			name: "short hex",
			hex:  "#0f8",
			want: RGBA{R: 0x00, G: 0xff, B: 0x88, A: 1},
		},
		{
			name: "long hex with spaces",
			hex:  "  #112233  ",
			want: RGBA{R: 0x11, G: 0x22, B: 0x33, A: 1},
		},
		{
			name: "invalid",
			hex:  "not-a-color",
			want: RGBA{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromHex(tt.hex)
			if got != tt.want {
				t.Fatalf("FromHex(%q) = %#v, want %#v", tt.hex, got, tt.want)
			}
		})
	}
}

func TestRGBAToHex(t *testing.T) {
	got := (RGBA{R: 0x12, G: 0xab, B: 0x00, A: 0.25}).ToHex()
	if got != "#12AB00" {
		t.Fatalf("ToHex() = %q, want %q", got, "#12AB00")
	}
}

func TestBlendColors(t *testing.T) {
	tests := []struct {
		name string
		fg   RGBA
		bg   RGBA
		want RGBA
	}{
		{
			name: "opaque foreground replaces background",
			fg:   RGBA{R: 255, G: 0, B: 0, A: 1},
			bg:   RGBA{R: 0, G: 0, B: 255, A: 1},
			want: RGBA{R: 255, G: 0, B: 0, A: 1},
		},
		{
			name: "semi transparent foreground blends over opaque background",
			fg:   RGBA{R: 255, G: 0, B: 0, A: 0.5},
			bg:   RGBA{R: 0, G: 0, B: 255, A: 1},
			want: RGBA{R: 128, G: 0, B: 128, A: 1},
		},
		{
			name: "transparent foreground keeps background",
			fg:   RGBA{R: 255, G: 255, B: 255, A: 0},
			bg:   RGBA{R: 10, G: 20, B: 30, A: 1},
			want: RGBA{R: 10, G: 20, B: 30, A: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BlendColors(tt.fg, tt.bg)
			if got.R != tt.want.R || got.G != tt.want.G || got.B != tt.want.B || math.Abs(got.A-tt.want.A) > 1e-9 {
				t.Fatalf("BlendColors(%#v, %#v) = %#v, want %#v", tt.fg, tt.bg, got, tt.want)
			}
		})
	}
}

func TestBlendHex(t *testing.T) {
	got := BlendHex("#FF0000", "#0000FF", 0.5)
	if got != "#800080" {
		t.Fatalf("BlendHex() = %q, want %q", got, "#800080")
	}
}

func TestOpacityStack(t *testing.T) {
	var stack OpacityStack

	if got := stack.Current(); got != 1 {
		t.Fatalf("initial Current() = %v, want 1", got)
	}
	if got := stack.Depth(); got != 0 {
		t.Fatalf("initial Depth() = %d, want 0", got)
	}

	stack.Push(0.5)
	if got := stack.Current(); got != 0.5 {
		t.Fatalf("Current() after first Push = %v, want 0.5", got)
	}
	if got := stack.Depth(); got != 1 {
		t.Fatalf("Depth() after first Push = %d, want 1", got)
	}

	stack.Push(0.4)
	if got := stack.Current(); math.Abs(got-0.2) > 1e-9 {
		t.Fatalf("Current() after second Push = %v, want 0.2", got)
	}
	if got := stack.Depth(); got != 2 {
		t.Fatalf("Depth() after second Push = %d, want 2", got)
	}

	applied := stack.ApplyToColor(RGBA{R: 1, G: 2, B: 3, A: 0.5})
	if math.Abs(applied.A-0.1) > 1e-9 {
		t.Fatalf("ApplyToColor().A = %v, want 0.1", applied.A)
	}

	stack.Pop()
	if got := stack.Current(); got != 0.5 {
		t.Fatalf("Current() after Pop = %v, want 0.5", got)
	}
	stack.Pop()
	stack.Pop()
	if got := stack.Current(); got != 1 {
		t.Fatalf("Current() after popping to base = %v, want 1", got)
	}
	if got := stack.Depth(); got != 0 {
		t.Fatalf("Depth() after popping to base = %d, want 0", got)
	}
}

func TestLayerCompositor(t *testing.T) {
	var compositor LayerCompositor
	compositor.AddLayer("\x1b[38;2;255;0;0mA", 0, 0, 1)
	compositor.AddLayer("\x1b[38;2;0;0;255mB", 0, 0, 0.5)

	buffer := compositor.Composite(2, 1)

	left := buffer.back[buffer.index(0, 0)]
	if left.Grapheme != "B" {
		t.Fatalf("left Grapheme = %q, want %q", left.Grapheme, "B")
	}
	if left.Fg != lipgloss.Color("#0000FF") {
		t.Fatalf("left Fg = %q, want %q", left.Fg, lipgloss.Color("#0000FF"))
	}

	right := buffer.back[buffer.index(1, 0)]
	if right != blankCell() {
		t.Fatalf("right cell = %#v, want blank %#v", right, blankCell())
	}
}
