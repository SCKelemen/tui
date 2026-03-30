package input

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestContinuationPromptConstructor(t *testing.T) {
	p := NewContinuationPrompt()
	if p == nil {
		t.Fatal("NewContinuationPrompt returned nil")
	}

	if p.placeholder != "type your message..." {
		t.Fatalf("unexpected placeholder: %q", p.placeholder)
	}

	if p.Focused() {
		t.Fatal("prompt should not be focused by default")
	}
}

func TestContinuationPromptPlaceholder(t *testing.T) {
	p := NewContinuationPrompt(
		WithContinuationPromptPlaceholder("Say hello"),
		WithContinuationPromptPrefix("> "),
	)

	view := p.View()
	if view == "" {
		t.Fatal("view should not be empty")
	}

	if !strings.Contains(view, "Say hello") {
		t.Fatal("view should include placeholder")
	}
}

func TestContinuationPromptHistoryNavigation(t *testing.T) {
	p := NewContinuationPrompt(WithContinuationPromptHistory([]string{"first", "second"}))
	p.Focus()

	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("draft")})
	if p.input != "draft" {
		t.Fatalf("expected draft input, got %q", p.input)
	}

	p.Update(tea.KeyMsg{Type: tea.KeyUp})
	if p.input != "second" {
		t.Fatalf("expected most recent history entry, got %q", p.input)
	}

	p.Update(tea.KeyMsg{Type: tea.KeyUp})
	if p.input != "first" {
		t.Fatalf("expected older history entry, got %q", p.input)
	}

	p.Update(tea.KeyMsg{Type: tea.KeyDown})
	if p.input != "second" {
		t.Fatalf("expected newer history entry, got %q", p.input)
	}

	p.Update(tea.KeyMsg{Type: tea.KeyDown})
	if p.input != "draft" {
		t.Fatalf("expected restored draft input, got %q", p.input)
	}
}
