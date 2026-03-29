package input

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAutocompleteView(t *testing.T) {
	a := NewAutocomplete(nil, WithAutocompletePlaceholder("Search..."))
	a.SetValue("hello")

	view := a.View()
	if view == "" {
		t.Fatal("expected non-empty view")
	}
	if !strings.Contains(view, "hello") {
		t.Fatalf("expected view to contain input value, got: %q", view)
	}
}

func TestAutocompleteSuggestions(t *testing.T) {
	calls := 0
	lastInput := ""
	a := NewAutocomplete(func(input string) []Suggestion {
		calls++
		lastInput = input
		if input == "a" {
			return []Suggestion{{Text: "alpha"}}
		}
		return nil
	})

	a.Focus()
	_, _ = a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	if calls == 0 {
		t.Fatal("expected suggestFn to be called")
	}
	if lastInput != "a" {
		t.Fatalf("expected suggestFn input %q, got %q", "a", lastInput)
	}
	if len(a.suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(a.suggestions))
	}
}

func TestAutocompleteAccept(t *testing.T) {
	a := NewAutocomplete(func(input string) []Suggestion {
		if input == "a" {
			return []Suggestion{
				{Text: "alpha", Description: "first", Value: "A"},
				{Text: "beta", Description: "second", Value: "B"},
			}
		}
		return nil
	})

	a.SetValue("a")
	a.Focus()

	_, cmd := a.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected enter to return a command")
	}

	msg := cmd()
	autocompleteMsg, ok := msg.(AutocompleteMsg)
	if !ok {
		t.Fatalf("expected AutocompleteMsg, got %T", msg)
	}
	if autocompleteMsg.Suggestion.Text != "alpha" {
		t.Fatalf("expected first suggestion to be selected, got %q", autocompleteMsg.Suggestion.Text)
	}
	if autocompleteMsg.Suggestion.Value != "A" {
		t.Fatalf("expected suggestion value %q, got %q", "A", autocompleteMsg.Suggestion.Value)
	}
}

func TestAutocompleteDismiss(t *testing.T) {
	a := NewAutocomplete(func(input string) []Suggestion {
		if input == "a" {
			return []Suggestion{{Text: "alpha"}}
		}
		return nil
	})

	a.SetValue("a")
	a.Focus()
	if !a.showSuggestions {
		t.Fatal("expected suggestions to be visible when focused with matches")
	}

	_, _ = a.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if a.showSuggestions {
		t.Fatal("expected Esc to hide suggestions")
	}
}

func TestAutocompleteFocusBlur(t *testing.T) {
	a := NewAutocomplete(nil)

	if a.Focused() {
		t.Fatal("expected autocomplete to be unfocused initially")
	}

	a.Focus()
	if !a.Focused() {
		t.Fatal("expected autocomplete to be focused after Focus()")
	}

	a.Blur()
	if a.Focused() {
		t.Fatal("expected autocomplete to be unfocused after Blur()")
	}
}

func TestAutocompleteValueSetValue(t *testing.T) {
	a := NewAutocomplete(nil)

	a.SetValue("hello")
	if a.Value() != "hello" {
		t.Fatalf("expected value %q, got %q", "hello", a.Value())
	}

	a.SetValue("world")
	if a.Value() != "world" {
		t.Fatalf("expected value %q, got %q", "world", a.Value())
	}
}

func TestAutocompleteMaxSuggestions(t *testing.T) {
	a := NewAutocomplete(func(input string) []Suggestion {
		if input == "a" {
			return []Suggestion{
				{Text: "alpha"},
				{Text: "beta"},
				{Text: "gamma"},
				{Text: "delta"},
			}
		}
		return nil
	}, WithAutocompleteMaxSuggestions(2))

	a.SetValue("a")
	a.Focus()

	if len(a.suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(a.suggestions))
	}

	view := a.View()
	if !strings.Contains(view, "alpha") || !strings.Contains(view, "beta") {
		t.Fatalf("expected view to contain first two suggestions, got: %q", view)
	}
	if strings.Contains(view, "gamma") || strings.Contains(view, "delta") {
		t.Fatalf("expected view to hide suggestions beyond max, got: %q", view)
	}
}

func TestAutocompleteEmptySuggestions(t *testing.T) {
	a := NewAutocomplete(func(input string) []Suggestion {
		if strings.TrimSpace(input) == "" {
			return nil
		}
		return []Suggestion{{Text: "alpha"}}
	})

	a.Focus()
	a.SetValue("")

	if len(a.suggestions) != 0 {
		t.Fatalf("expected no suggestions for empty input, got %d", len(a.suggestions))
	}
	if a.showSuggestions {
		t.Fatal("expected dropdown to be hidden for empty input")
	}

	view := a.View()
	if strings.Contains(view, "▸") {
		t.Fatalf("expected no suggestion marker in view, got: %q", view)
	}
}
