package display

import (
	"strings"
	"testing"

	"github.com/SCKelemen/tui/v2/style"
)

func TestFileCardBasicRendering(t *testing.T) {
	card := NewFileCard("main.go", 323, 0)
	view := card.View()
	plain := stripANSI(view)

	expected := strings.Join([]string{
		"╭──────────────╮",
		"│ main.go +323 │",
		"╰──────────────╯",
	}, "\n")

	if plain != expected {
		t.Fatalf("unexpected rendering:\n%s", plain)
	}
}

func TestFileCardAdditionsOnly(t *testing.T) {
	card := NewFileCard("main.go", 57, 0)
	view := card.View()
	plain := stripANSI(view)

	if !strings.Contains(plain, "main.go +57") {
		t.Fatalf("expected additions text in view, got:\n%s", plain)
	}
	if strings.Contains(plain, "-") {
		t.Fatalf("did not expect deletions text in view, got:\n%s", plain)
	}
	if !strings.Contains(view, style.ANSIGreen+"+57"+style.ANSIReset) {
		t.Fatalf("expected green additions styling, got:\n%s", view)
	}
}

func TestFileCardDeletionsOnly(t *testing.T) {
	card := NewFileCard("main.go", 0, 120)
	view := card.View()
	plain := stripANSI(view)

	if !strings.Contains(plain, "main.go -120") {
		t.Fatalf("expected deletions text in view, got:\n%s", plain)
	}
	if strings.Contains(plain, "+") {
		t.Fatalf("did not expect additions text in view, got:\n%s", plain)
	}
	if !strings.Contains(view, style.ANSIRed+"-120"+style.ANSIReset) {
		t.Fatalf("expected red deletions styling, got:\n%s", view)
	}
}

func TestFileCardBothAdditionsAndDeletions(t *testing.T) {
	card := NewFileCard("main.go", 57, 120)
	view := card.View()
	plain := stripANSI(view)

	expected := strings.Join([]string{
		"╭──────────────────╮",
		"│ main.go +57 -120 │",
		"╰──────────────────╯",
	}, "\n")

	if plain != expected {
		t.Fatalf("unexpected rendering:\n%s", plain)
	}
	if !strings.Contains(view, style.ANSIGreen+"+57"+style.ANSIReset) {
		t.Fatalf("expected green additions styling, got:\n%s", view)
	}
	if !strings.Contains(view, style.ANSIRed+"-120"+style.ANSIReset) {
		t.Fatalf("expected red deletions styling, got:\n%s", view)
	}
}

func TestFileCardNeitherAdditionsNorDeletions(t *testing.T) {
	card := NewFileCard("main.go", 0, 0)
	plain := stripANSI(card.View())

	expected := strings.Join([]string{
		"╭─────────╮",
		"│ main.go │",
		"╰─────────╯",
	}, "\n")

	if plain != expected {
		t.Fatalf("unexpected rendering:\n%s", plain)
	}
}

func TestFileCardLongFilename(t *testing.T) {
	card := NewFileCard("very-long-filename-with-many-segments.go", 10, 2)
	view := card.View()
	plain := stripANSI(view)
	lines := strings.Split(plain, "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[1], "very-long-filename-with-many-segments.go +10 -2") {
		t.Fatalf("expected full long filename and stats, got:\n%s", plain)
	}
	if len([]rune(lines[0])) != len([]rune(lines[1])) || len([]rune(lines[1])) != len([]rune(lines[2])) {
		t.Fatalf("expected all lines to be equal width, got:\n%s", plain)
	}
}

func TestFileCardRowWithMultipleCards(t *testing.T) {
	cardA := NewFileCard("main.go", 10, 0)
	cardB := NewFileCard("utils.go", 0, 4)
	cardC := NewFileCard("README.md", 0, 0)

	row := NewFileCardRow(cardA, cardB, cardC)
	plain := stripANSI(row)
	lines := strings.Split(plain, "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[1], "main.go +10") || !strings.Contains(lines[1], "utils.go -4") || !strings.Contains(lines[1], "README.md") {
		t.Fatalf("expected all card contents on middle row, got:\n%s", plain)
	}

	if strings.Count(lines[0], " ") < 2 || strings.Count(lines[1], " ") < 2 || strings.Count(lines[2], " ") < 2 {
		t.Fatalf("expected visible gaps between cards, got:\n%s", plain)
	}
}

func TestFileCardSetters(t *testing.T) {
	card := NewFileCard("a.go", 1, 1)
	card.SetFilename("b.go")
	card.SetAdditions(42)
	card.SetDeletions(7)

	plain := stripANSI(card.View())
	if !strings.Contains(plain, "b.go +42 -7") {
		t.Fatalf("expected updated values, got:\n%s", plain)
	}
}
