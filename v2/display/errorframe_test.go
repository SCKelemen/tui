package display

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestErrorFrameBasicRendering(t *testing.T) {
	e := NewErrorFrame("UnknownError", "502 error code: 502")
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	view := stripANSI(e.View())
	lines := splitFrameLines(view)
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}

	if !strings.HasPrefix(lines[0], "┌") || !strings.HasSuffix(lines[0], "┐") {
		t.Fatalf("expected top border, got %q", lines[0])
	}
	if !strings.Contains(lines[1], "UnknownError") {
		t.Fatalf("expected title line to include title, got %q", lines[1])
	}
	if !strings.Contains(lines[2], "502 error code: 502") {
		t.Fatalf("expected message line to include message, got %q", lines[2])
	}
	if !strings.HasPrefix(lines[3], "└") || !strings.HasSuffix(lines[3], "┘") {
		t.Fatalf("expected bottom border, got %q", lines[3])
	}

	if len([]rune(lines[0])) != len([]rune(lines[3])) {
		t.Fatal("top and bottom borders should have equal width")
	}
}

func TestErrorFrameMultilineMessage(t *testing.T) {
	e := NewErrorFrame("UnknownError", "first line\nsecond line")
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	lines := splitFrameLines(stripANSI(e.View()))
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines for multiline message, got %d", len(lines))
	}

	if !strings.Contains(lines[2], "first line") {
		t.Fatalf("expected first message line, got %q", lines[2])
	}
	if !strings.Contains(lines[3], "second line") {
		t.Fatalf("expected second message line, got %q", lines[3])
	}
}

func TestErrorFrameAutoWidth(t *testing.T) {
	e := NewErrorFrame("Err", "long message")
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	lines := splitFrameLines(stripANSI(e.View()))
	topWidth := len([]rune(lines[0]))
	want := len([]rune("long message")) + 4
	if topWidth != want {
		t.Fatalf("expected auto width %d, got %d", want, topWidth)
	}
}

func TestErrorFrameFixedWidth(t *testing.T) {
	e := NewErrorFrame("Err", "this message wraps", WithErrorFrameWidth(10))
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	lines := splitFrameLines(stripANSI(e.View()))
	for i, line := range lines {
		if got := len([]rune(line)); got != 14 {
			t.Fatalf("line %d expected width 14, got %d (%q)", i, got, line)
		}
	}
}

func TestErrorFrameEmptyMessage(t *testing.T) {
	e := NewErrorFrame("Err", "")
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	lines := splitFrameLines(stripANSI(e.View()))
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines for empty message, got %d", len(lines))
	}

	msg := extractFrameContent(lines[2])
	if strings.TrimSpace(msg) != "" {
		t.Fatalf("expected empty message line, got %q", msg)
	}
}

func TestErrorFrameLongTitleAutoWidth(t *testing.T) {
	title := "VeryLongErrorTitleWithoutAnySpaces"
	e := NewErrorFrame(title, "msg")
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	lines := splitFrameLines(stripANSI(e.View()))
	want := len([]rune(title)) + 4
	if got := len([]rune(lines[0])); got != want {
		t.Fatalf("expected width from title %d, got %d", want, got)
	}
}

func TestErrorFrameWordWrapping(t *testing.T) {
	e := NewErrorFrame("Err", "this is a wrapped message body", WithErrorFrameWidth(12))
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	lines := splitFrameLines(stripANSI(e.View()))
	if len(lines) < 5 {
		t.Fatalf("expected wrapped message to use multiple lines, got %d", len(lines))
	}

	for i := 1; i < len(lines)-1; i++ {
		content := extractFrameContent(lines[i])
		if len([]rune(content)) > 12 {
			t.Fatalf("content line %d exceeds width: %q", i, content)
		}
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "wrapped") {
		t.Fatalf("expected wrapped content to be rendered: %q", joined)
	}
}

func TestErrorFrameAutoWidthUsesWindowSize(t *testing.T) {
	e := NewErrorFrame("VeryLongErrorTitle", "This is a long message line that should be constrained")
	e.Update(tea.WindowSizeMsg{Width: 20, Height: 30})

	lines := splitFrameLines(stripANSI(e.View()))
	if got := len([]rune(lines[0])); got != 20 {
		t.Fatalf("expected frame width to match window cap 20, got %d", got)
	}
}

func splitFrameLines(view string) []string {
	trimmed := strings.TrimSuffix(view, "\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

func extractFrameContent(frameLine string) string {
	runes := []rune(frameLine)
	if len(runes) < 4 {
		return ""
	}
	// line format: │ + space + content + space + │
	return string(runes[2 : len(runes)-2])
}
