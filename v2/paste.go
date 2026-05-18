package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// BracketedPasteMsg is dispatched when the user pastes text into the
// terminal while bracketed paste mode is enabled. Components that want
// to handle paste explicitly — for example a text input that needs to
// insert the entire payload as a single edit instead of one keypress at
// a time — can match on this message in their own Update method.
//
// Upstream support: bubbletea v1.3.10 surfaces bracketed paste as a
// tea.KeyMsg whose underlying Key.Paste field is true and whose Runes
// slice contains the pasted content. Application.Update converts those
// messages into BracketedPasteMsg before routing them so consumers can
// match on the typed event instead of inspecting Key.Paste directly.
// If a newer bubbletea release introduces a dedicated paste message,
// the adapter in keyMsgAsPaste can be extended without changing this
// public API.
type BracketedPasteMsg struct {
	// Content is the full pasted payload as a UTF-8 string. It may
	// contain newlines, tabs, and other control characters exactly as
	// the user copied them.
	Content string
}

// keyMsgAsPaste reports whether a bubbletea key message represents a
// bracketed-paste payload and, if so, returns the equivalent
// BracketedPasteMsg. The boolean return is false for ordinary
// keypresses so callers can continue handling them through the normal
// key path.
//
// This adapter exists to insulate the rest of the package — and any
// downstream consumers — from the precise shape bubbletea uses to
// surface paste. v1.3.10 sets Key.Paste=true and packs the payload
// into Runes; future releases may introduce a dedicated message type.
func keyMsgAsPaste(msg tea.KeyMsg) (BracketedPasteMsg, bool) {
	if !msg.Paste {
		return BracketedPasteMsg{}, false
	}
	return BracketedPasteMsg{Content: string(msg.Runes)}, true
}
