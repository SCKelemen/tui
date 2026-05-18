package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// pasteAwareComponent records every BracketedPasteMsg and tea.KeyMsg it
// observes through Update so tests can assert exactly what the
// Application routed to it.
type pasteAwareComponent struct {
	mockComponent
	pasteEvents []BracketedPasteMsg
	keyEvents   []tea.KeyMsg
}

func (p *pasteAwareComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch m := msg.(type) {
	case BracketedPasteMsg:
		p.pasteEvents = append(p.pasteEvents, m)
	case tea.KeyMsg:
		p.keyEvents = append(p.keyEvents, m)
	}
	_, cmd := p.mockComponent.Update(msg)
	return p, cmd
}

func TestBracketedPasteFromKeyMsgIsRoutedAsBracketedPasteMsg(t *testing.T) {
	app := NewApplication()
	c := &pasteAwareComponent{}
	app.AddComponent(c)

	// Simulate bubbletea surfacing a bracketed paste: a KeyMsg with
	// Paste=true and the pasted content in Runes.
	payload := "hello\nworld"
	app.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(payload),
		Paste: true,
	})

	if len(c.pasteEvents) != 1 {
		t.Fatalf("expected component to receive 1 BracketedPasteMsg, got %d", len(c.pasteEvents))
	}
	if got := c.pasteEvents[0].Content; got != payload {
		t.Fatalf("expected paste content %q, got %q", payload, got)
	}
	if len(c.keyEvents) != 0 {
		t.Fatalf("expected no raw KeyMsg deliveries for paste, got %d", len(c.keyEvents))
	}
}

func TestBracketedPasteMsgDispatchedDirectlyReachesFocusedComponent(t *testing.T) {
	// Independent of upstream paste support, a consumer that synthesizes
	// BracketedPasteMsg into Update must have it routed to the focused
	// component just like any other untyped message.
	app := NewApplication()
	c := &pasteAwareComponent{}
	app.AddComponent(c)

	app.Update(BracketedPasteMsg{Content: "manual"})

	if len(c.pasteEvents) != 1 {
		t.Fatalf("expected 1 paste event, got %d", len(c.pasteEvents))
	}
	if got := c.pasteEvents[0].Content; got != "manual" {
		t.Fatalf("expected paste content %q, got %q", "manual", got)
	}
}

func TestBracketedPasteDoesNotTriggerQuitKey(t *testing.T) {
	// A pasted payload that contains the quit key must not quit the
	// application. The paste path bypasses Application shortcuts.
	app := NewApplication(WithQuitKey("q"))
	c := &pasteAwareComponent{}
	app.AddComponent(c)

	_, cmd := app.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("hi q there"),
		Paste: true,
	})
	if cmd != nil {
		t.Fatalf("expected paste with embedded quit key to return nil cmd, got %v", cmd)
	}
	if len(c.pasteEvents) != 1 {
		t.Fatalf("expected paste to be delivered exactly once, got %d", len(c.pasteEvents))
	}
}

func TestBracketedPasteDoesNotCycleFocus(t *testing.T) {
	// A pasted payload that contains a tab must not cycle focus.
	app := NewApplication()
	c1 := &pasteAwareComponent{}
	c2 := &pasteAwareComponent{}
	app.AddComponent(c1)
	app.AddComponent(c2)

	app.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("col1\tcol2"),
		Paste: true,
	})

	if !c1.Focused() || c2.Focused() {
		t.Fatal("paste containing tab must not cycle focus")
	}
	if len(c1.pasteEvents) != 1 {
		t.Fatalf("expected c1 to receive the paste, got %d events", len(c1.pasteEvents))
	}
}

func TestKeyMsgAsPasteRoundTripsContent(t *testing.T) {
	// Unit test the adapter directly to lock in its contract.
	in := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("multi\nline\tpayload"),
		Paste: true,
	}
	got, ok := keyMsgAsPaste(in)
	if !ok {
		t.Fatal("expected keyMsgAsPaste to recognize Paste=true KeyMsg")
	}
	if got.Content != "multi\nline\tpayload" {
		t.Fatalf("unexpected paste content: %q", got.Content)
	}

	if _, ok := keyMsgAsPaste(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}); ok {
		t.Fatal("expected non-paste KeyMsg to return ok=false")
	}
}
