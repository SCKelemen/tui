package input

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPromptConstructor(t *testing.T) {
	p := NewPrompt()
	if p == nil {
		t.Fatal("NewPrompt returned nil")
	}

	if p.placeholder == "" {
		t.Fatal("expected default placeholder")
	}

	if p.Focused() {
		t.Fatal("prompt should not be focused initially")
	}
}

func TestPromptPlaceholder(t *testing.T) {
	p := NewPrompt(WithPromptPlaceholder("Ask something"), WithPromptWidth(60))
	p.Update(tea.WindowSizeMsg{Width: 60, Height: 20})

	view := p.View()
	if view == "" {
		t.Fatal("view should not be empty")
	}

	if !strings.Contains(view, "Ask something") {
		t.Fatal("view should include placeholder text")
	}
}

func TestPromptMultiline(t *testing.T) {
	p := NewPrompt(WithPromptWidth(80), WithPromptMaxHeight(8))
	p.SetValue("line1\nline2\nline3")

	if p.textarea.LineCount() != 3 {
		t.Fatalf("expected 3 lines, got %d", p.textarea.LineCount())
	}

	view := p.View()
	if view == "" {
		t.Fatal("view should not be empty")
	}

	if !strings.Contains(view, "line2") {
		t.Fatal("view should render multiline content")
	}
}

func TestPromptViewNonEmpty(t *testing.T) {
	p := NewPrompt(WithPromptWidth(50))
	p.Update(tea.WindowSizeMsg{Width: 50, Height: 20})

	view := p.View()
	if view == "" {
		t.Fatal("view should be non-empty when width is set")
	}

	if !strings.Contains(view, "┌") || !strings.Contains(view, "└") {
		t.Fatal("view should render bordered prompt")
	}
}
