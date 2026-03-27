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

	got := osc52Sequence(content, target)
	if got != expected {
		t.Fatalf("unexpected OSC 52 sequence\nexpected: %q\ngot:      %q", expected, got)
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
			got := osc52Sequence(tc.content, tc.target)
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
