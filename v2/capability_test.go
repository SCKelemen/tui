package tui

import (
	"testing"
)

// clearTerminalEnv resets every environment variable that DetectCapabilities
// consults, so each test starts from a deterministic baseline.
func clearTerminalEnv(t *testing.T) {
	t.Helper()
	t.Setenv("COLORTERM", "")
	t.Setenv("TERM", "")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("KITTY_WINDOW_ID", "")
}

func TestDetectCapabilitiesEmptyEnvReturnsAllFalse(t *testing.T) {
	clearTerminalEnv(t)
	got := DetectCapabilities()
	want := TerminalCapabilities{}
	if got != want {
		t.Fatalf("DetectCapabilities() = %+v, want %+v", got, want)
	}
}

func TestDetectCapabilitiesTruecolorFromColorterm(t *testing.T) {
	clearTerminalEnv(t)
	t.Setenv("COLORTERM", "truecolor")
	if !DetectCapabilities().Truecolor {
		t.Errorf("COLORTERM=truecolor should set Truecolor")
	}
	t.Setenv("COLORTERM", "24bit")
	if !DetectCapabilities().Truecolor {
		t.Errorf("COLORTERM=24bit should set Truecolor")
	}
}

func TestDetectCapabilitiesTruecolorFromTermDirect(t *testing.T) {
	clearTerminalEnv(t)
	t.Setenv("TERM", "xterm-direct")
	if !DetectCapabilities().Truecolor {
		t.Errorf("TERM containing 'direct' should set Truecolor")
	}
}

func TestDetectCapabilitiesColor256(t *testing.T) {
	clearTerminalEnv(t)
	t.Setenv("TERM", "xterm-256color")
	caps := DetectCapabilities()
	if !caps.Color256 {
		t.Errorf("TERM=xterm-256color should set Color256")
	}
	if !caps.Mouse {
		t.Errorf("TERM=xterm-256color should set Mouse")
	}
}

func TestDetectCapabilitiesMouseFromTermProgram(t *testing.T) {
	candidates := []string{"iTerm.app", "vscode", "Apple_Terminal", "WezTerm", "ghostty", "kitty", "alacritty"}
	for _, name := range candidates {
		t.Run(name, func(t *testing.T) {
			clearTerminalEnv(t)
			t.Setenv("TERM_PROGRAM", name)
			if !DetectCapabilities().Mouse {
				t.Errorf("TERM_PROGRAM=%s should set Mouse", name)
			}
		})
	}
}

func TestDetectCapabilitiesKittyFromTerm(t *testing.T) {
	clearTerminalEnv(t)
	t.Setenv("TERM", "xterm-kitty")
	if !DetectCapabilities().Kitty {
		t.Errorf("TERM=xterm-kitty should set Kitty")
	}
}

func TestDetectCapabilitiesKittyFromWindowID(t *testing.T) {
	clearTerminalEnv(t)
	t.Setenv("KITTY_WINDOW_ID", "1")
	if !DetectCapabilities().Kitty {
		t.Errorf("KITTY_WINDOW_ID set should imply Kitty")
	}
}

func TestDetectCapabilitiesSixelFromTerm(t *testing.T) {
	clearTerminalEnv(t)
	t.Setenv("TERM", "xterm-sixel")
	if !DetectCapabilities().Sixel {
		t.Errorf("TERM containing 'sixel' should set Sixel")
	}
}

func TestDetectCapabilitiesSixelFromHostTerminal(t *testing.T) {
	for _, name := range []string{"iTerm.app", "ghostty"} {
		t.Run(name, func(t *testing.T) {
			clearTerminalEnv(t)
			t.Setenv("TERM_PROGRAM", name)
			if !DetectCapabilities().Sixel {
				t.Errorf("TERM_PROGRAM=%s should set Sixel", name)
			}
		})
	}
}

func TestDetectCapabilitiesDerivedFlagsFollowModernSignals(t *testing.T) {
	clearTerminalEnv(t)
	t.Setenv("TERM_PROGRAM", "ghostty")
	caps := DetectCapabilities()
	if !caps.BracketedPaste || !caps.FocusEvents || !caps.OSC52 {
		t.Errorf("modern terminal should imply BracketedPaste / FocusEvents / OSC52, got %+v", caps)
	}
}

func TestDetectCapabilitiesUnknownTerminalLeavesDerivedFlagsFalse(t *testing.T) {
	clearTerminalEnv(t)
	t.Setenv("TERM", "dumb")
	caps := DetectCapabilities()
	if caps.BracketedPaste || caps.FocusEvents || caps.OSC52 {
		t.Errorf("unknown terminal should not advertise modern features, got %+v", caps)
	}
}

func TestDetectCapabilitiesCmdEmitsCapabilityMsg(t *testing.T) {
	clearTerminalEnv(t)
	t.Setenv("COLORTERM", "truecolor")
	cmd := DetectCapabilitiesCmd()
	if cmd == nil {
		t.Fatal("DetectCapabilitiesCmd returned nil")
	}
	msg := cmd()
	got, ok := msg.(CapabilityMsg)
	if !ok {
		t.Fatalf("expected CapabilityMsg, got %T", msg)
	}
	if !got.Caps.Truecolor {
		t.Errorf("expected Truecolor in CapabilityMsg.Caps, got %+v", got.Caps)
	}
}
