package selection

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func TestOSC52Sequence(t *testing.T) {
	content := "hello"
	target := ClipboardSystem

	expectedBase64 := base64.StdEncoding.EncodeToString([]byte(content))
	expected := fmt.Sprintf(string([]byte{0x1b})+"]52;%s;%s"+string([]byte{0x07}), target, expectedBase64)

	got := OSC52Sequence(target, content)
	if got != expected {
		t.Fatalf("unexpected OSC 52 sequence\nexpected: %q\ngot:      %q", expected, got)
	}
}

func TestOSC52SequenceMatchesHelloFixture(t *testing.T) {
	// Locked-down assertion to guard against accidental wire-format
	// changes. The reference bytes come from the xterm OSC 52 spec.
	got := OSC52Sequence(ClipboardSystem, "hello")
	want := "\x1b]52;c;aGVsbG8=\x07"
	if got != want {
		t.Fatalf("OSC52Sequence(ClipboardSystem, %q) = %q, want %q", "hello", got, want)
	}
}

func TestOSC52SequenceBase64EncodingVariants(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		target  ClipboardTarget
	}{
		{name: "empty", content: "", target: ClipboardSystem},
		{name: "ascii", content: "clipboard-content-123", target: ClipboardSystem},
		{name: "unicode", content: "Hello, 世界 👋", target: ClipboardPrimary},
		{name: "multiline", content: "line1\nline2\nline3", target: ClipboardSystem},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := OSC52Sequence(tc.target, tc.content)
			expectedBase64 := base64.StdEncoding.EncodeToString([]byte(tc.content))

			prefix := fmt.Sprintf(string([]byte{0x1b})+"]52;%s;", tc.target)
			suffix := string([]byte{0x07})

			if !strings.HasPrefix(got, prefix) {
				t.Fatalf("sequence missing expected prefix %q: %q", prefix, got)
			}
			if !strings.HasSuffix(got, suffix) {
				t.Fatalf("sequence missing expected BEL suffix: %q", got)
			}

			encodedContent := strings.TrimSuffix(strings.TrimPrefix(got, prefix), suffix)
			if encodedContent != expectedBase64 {
				t.Fatalf("unexpected base64 payload\nexpected: %q\ngot:      %q", expectedBase64, encodedContent)
			}
		})
	}
}

func TestWriteClipboardReturnsCmd(t *testing.T) {
	cmd := WriteClipboard("test")
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}
}

func TestSupportsOSC52KnownGoodTerm(t *testing.T) {
	cases := []string{"iTerm.app", "ghostty", "WezTerm", "kitty", "vscode", "Apple_Terminal", "alacritty", "tmux"}
	for _, term := range cases {
		t.Run(term, func(t *testing.T) {
			t.Setenv("TERM_PROGRAM", term)
			t.Setenv("TERM", "xterm-256color")
			t.Setenv("KITTY_WINDOW_ID", "")
			if !SupportsOSC52() {
				t.Errorf("SupportsOSC52() = false for TERM_PROGRAM=%s, want true", term)
			}
		})
	}
}

func TestSupportsOSC52UnknownReturnsFalse(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("TERM", "")
	t.Setenv("KITTY_WINDOW_ID", "")
	if SupportsOSC52() {
		t.Error("SupportsOSC52() = true for empty env, want false")
	}
}

func TestSupportsOSC52KittyWindowID(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("TERM", "")
	t.Setenv("KITTY_WINDOW_ID", "42")
	if !SupportsOSC52() {
		t.Error("SupportsOSC52() = false with KITTY_WINDOW_ID set, want true")
	}
}

func TestSupportsOSC52DumbTermReturnsFalse(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "ghostty")
	t.Setenv("TERM", "dumb")
	t.Setenv("KITTY_WINDOW_ID", "")
	if SupportsOSC52() {
		t.Error("SupportsOSC52() = true for TERM=dumb, want false")
	}
}

func TestClipboardWriteMsgFields(t *testing.T) {
	msg := ClipboardWriteMsg{
		Content: "content",
		Target:  ClipboardPrimary,
		Err:     nil,
	}

	if msg.Content != "content" {
		t.Fatalf("unexpected content: %q", msg.Content)
	}
	if msg.Target != ClipboardPrimary {
		t.Fatalf("unexpected target: %q", msg.Target)
	}
	if msg.Err != nil {
		t.Fatalf("unexpected error: %v", msg.Err)
	}
}
