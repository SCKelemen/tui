package display

import (
	"strings"
	"testing"
)

func TestLinkConstructor(t *testing.T) {
	link := NewLink("https://docs.lovable.dev", "Docs")
	if link == nil {
		t.Fatal("NewLink returned nil")
	}
	if link.URL != "https://docs.lovable.dev" {
		t.Fatalf("unexpected URL: %q", link.URL)
	}
	if link.Text != "Docs" {
		t.Fatalf("unexpected text: %q", link.Text)
	}
}

func TestLinkViewRendersURLWithOSC8OrPlainText(t *testing.T) {
	link := NewLink("https://docs.lovable.dev", "Docs", WithLinkColor("39"))
	view := link.View()
	if !strings.Contains(view, "Docs") {
		t.Fatalf("expected link text in view, got %q", view)
	}
	if !(strings.Contains(view, "\033]8;;https://docs.lovable.dev\033\\") || view == "Docs") {
		t.Fatalf("expected OSC 8 hyperlink or plain text fallback, got %q", view)
	}

	rendered := RenderLink("https://example.com", "Example")
	if !strings.Contains(rendered, "Example") || !strings.Contains(rendered, "https://example.com") {
		t.Fatalf("RenderLink output missing expected content: %q", rendered)
	}
}
