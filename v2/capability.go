package tui

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// TerminalCapabilities describes the optional features the host
// terminal is believed to support. The values are derived heuristically
// from environment variables alone; no terminal queries are issued.
// Detection is best-effort and intentionally conservative: a true value
// means a reasonably modern terminal advertises the feature, while a
// false value only means the heuristic could not confirm support, not
// that the feature is definitely missing.
type TerminalCapabilities struct {
	// Truecolor reports whether the terminal advertises 24-bit color
	// via COLORTERM or a "direct" TERM entry.
	Truecolor bool
	// Color256 reports whether the terminal advertises the 256-color
	// xterm palette via TERM.
	Color256 bool
	// Mouse reports whether the terminal is one of the well-known
	// emulators that supports xterm-style mouse reporting.
	Mouse bool
	// BracketedPaste reports whether the terminal is expected to
	// support DEC bracketed paste mode.
	BracketedPaste bool
	// FocusEvents reports whether the terminal is expected to emit
	// CSI focus/blur events when configured to do so.
	FocusEvents bool
	// OSC52 reports whether the terminal is expected to honor OSC 52
	// clipboard sequences.
	OSC52 bool
	// Kitty reports whether the terminal speaks the kitty keyboard
	// or graphics protocols.
	Kitty bool
	// Sixel reports whether the terminal supports the Sixel graphics
	// protocol.
	Sixel bool
}

// CapabilityMsg carries the result of an asynchronous capability probe.
// Applications that want to gate behaviour on terminal capabilities can
// match on this message in their own Update method.
type CapabilityMsg struct {
	Caps TerminalCapabilities
}

// modernMouseTerminals are the TERM_PROGRAM values known to support
// xterm-style mouse reporting. The check is case-insensitive.
var modernMouseTerminals = map[string]struct{}{
	"iterm.app":      {},
	"vscode":         {},
	"apple_terminal": {},
	"wezterm":        {},
	"ghostty":        {},
	"kitty":          {},
	"alacritty":      {},
}

// sixelHostTerminals are the TERM_PROGRAM values known to bundle a
// Sixel implementation.
var sixelHostTerminals = map[string]struct{}{
	"iterm.app": {},
	"ghostty":   {},
}

// DetectCapabilities inspects the process environment and returns the
// inferred terminal capability set. It performs no I/O. The result is
// stable for the lifetime of the process unless the environment is
// mutated (for example by tests using t.Setenv).
func DetectCapabilities() TerminalCapabilities {
	colorterm := strings.ToLower(strings.TrimSpace(os.Getenv("COLORTERM")))
	term := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))
	termProgram := strings.ToLower(strings.TrimSpace(os.Getenv("TERM_PROGRAM")))

	caps := TerminalCapabilities{}

	// Truecolor: COLORTERM=truecolor|24bit, or TERM contains "direct".
	if colorterm == "truecolor" || colorterm == "24bit" || strings.Contains(term, "direct") {
		caps.Truecolor = true
	}

	// 256-color: TERM contains "256".
	if strings.Contains(term, "256") {
		caps.Color256 = true
	}

	// Mouse: well-known program, or any xterm-derived TERM entry.
	if _, ok := modernMouseTerminals[termProgram]; ok {
		caps.Mouse = true
	} else if strings.Contains(term, "xterm") {
		caps.Mouse = true
	}

	// Kitty: TERM is one of the kitty values, or KITTY_WINDOW_ID is set.
	if term == "kitty" || term == "xterm-kitty" || os.Getenv("KITTY_WINDOW_ID") != "" {
		caps.Kitty = true
	}

	// Sixel: TERM advertises sixel, or the host terminal is known to
	// implement it.
	if strings.Contains(term, "sixel") {
		caps.Sixel = true
	} else if _, ok := sixelHostTerminals[termProgram]; ok {
		caps.Sixel = true
	}

	// BracketedPaste / FocusEvents / OSC52 are true if any of the
	// modern detection signals fired. This keeps the heuristic
	// conservative: when we can't recognise the terminal at all, we
	// avoid promising features that may not work.
	modern := caps.Mouse || caps.Kitty || caps.Truecolor
	caps.BracketedPaste = modern
	caps.FocusEvents = modern
	caps.OSC52 = modern

	return caps
}

// DetectCapabilitiesCmd returns a tea.Cmd that emits a CapabilityMsg
// with the result of DetectCapabilities. The command is opt-in:
// Application.Init does not invoke it automatically so that existing
// apps see no behaviour change. Apps that care about capabilities can
// return DetectCapabilitiesCmd() from their own Init.
func DetectCapabilitiesCmd() tea.Cmd {
	return func() tea.Msg {
		return CapabilityMsg{Caps: DetectCapabilities()}
	}
}
