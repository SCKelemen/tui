package display

import (
	"strings"
	"testing"
)

func TestHoverCardConstructor(t *testing.T) {
	sections := []HoverSection{{Content: "Hello"}}
	card := NewHoverCard(sections, WithHoverCardWidth(40), WithHoverCardBorderColor("#ff00aa"))

	if card == nil {
		t.Fatal("NewHoverCard returned nil")
	}
	if card.width != 40 {
		t.Fatalf("expected width=40, got %d", card.width)
	}
	if card.borderColor != "#ff00aa" {
		t.Fatalf("expected borderColor override, got %q", card.borderColor)
	}
}

func TestHoverCardContentRender(t *testing.T) {
	sections := []HoverSection{
		{Content: "**Type** `User`"},
		{Content: "func DoThing() error", Language: "go", IsCode: true},
	}
	card := NewHoverCard(sections, WithHoverCardWidth(50))

	plain := stripANSI(card.View())
	if strings.TrimSpace(plain) == "" {
		t.Fatal("expected non-empty rendered hover card")
	}
	if !strings.Contains(plain, "Type") {
		t.Fatalf("expected markdown content in view, got:\n%s", plain)
	}
	if !strings.Contains(plain, "func DoThing() error") {
		t.Fatalf("expected code content in view, got:\n%s", plain)
	}
}

func TestHoverCardPositioningByHeightClipping(t *testing.T) {
	sections := []HoverSection{{Content: strings.Repeat("line\n", 20)}}
	card := NewHoverCard(sections, WithHoverCardWidth(30), WithHoverCardMaxHeight(5))

	plain := stripANSI(card.View())
	lines := strings.Split(plain, "\n")
	if len(lines) != 5 {
		t.Fatalf("expected clipped height of 5 lines, got %d\n%s", len(lines), plain)
	}
	if !strings.Contains(plain, "...") {
		t.Fatalf("expected clipping indicator in view, got:\n%s", plain)
	}
}
