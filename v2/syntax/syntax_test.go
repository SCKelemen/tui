package syntax

import (
	"regexp"
	"strings"
	"testing"
)

// ansiEscape matches CSI / OSC escape sequences emitted by chroma's
// terminal256 formatter so tests can compare against the raw payload.
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

func stripANSI(s string) string { return ansiEscape.ReplaceAllString(s, "") }

func TestHighlightGo(t *testing.T) {
	h := New("monokai")
	src := `fmt.Println("hi")`
	out := h.Highlight("go", src)

	if out == "" {
		t.Fatalf("expected non-empty output")
	}
	if !strings.Contains(out, "\x1b") {
		t.Fatalf("expected at least one ESC byte in highlighted output, got %q", out)
	}
	stripped := stripANSI(out)
	for _, tok := range []string{"fmt", "Println", `"hi"`} {
		if !strings.Contains(stripped, tok) {
			t.Errorf("expected token %q in ANSI-stripped output %q", tok, stripped)
		}
	}
}

func TestHighlightUnknownLanguageReturnsSourceVerbatim(t *testing.T) {
	h := New("monokai")
	src := "this is not really code"
	out := h.Highlight("definitely-not-a-language", src)

	if out != src {
		t.Errorf("expected source returned verbatim, got %q", out)
	}
	if strings.Contains(out, "\x1b") {
		t.Errorf("expected no ANSI escapes, got %q", out)
	}
}

func TestSupportsLanguage(t *testing.T) {
	h := New("monokai")
	if !h.SupportsLanguage("go") {
		t.Errorf("expected SupportsLanguage(\"go\") = true")
	}
	if h.SupportsLanguage("notalanguage") {
		t.Errorf("expected SupportsLanguage(\"notalanguage\") = false")
	}
}

func TestDetectLanguage(t *testing.T) {
	got := DetectLanguage("foo.go")
	if !strings.EqualFold(got, "go") {
		t.Errorf("DetectLanguage(\"foo.go\") = %q, want case-insensitive match of \"go\"", got)
	}

	if got := DetectLanguage("foo.unknown"); got != "" {
		t.Errorf("DetectLanguage(\"foo.unknown\") = %q, want \"\"", got)
	}
}

func TestPlainHighlighterPassThrough(t *testing.T) {
	h := PlainHighlighter()
	src := `fmt.Println("hi")`
	if got := h.Highlight("go", src); got != src {
		t.Errorf("PlainHighlighter.Highlight: got %q, want %q", got, src)
	}
	if h.SupportsLanguage("go") {
		t.Errorf("PlainHighlighter.SupportsLanguage should always be false")
	}
}

func TestListStylesNonEmpty(t *testing.T) {
	if len(ListStyles()) == 0 {
		t.Fatalf("expected at least one chroma style registered")
	}
}

func TestListLanguagesNonEmpty(t *testing.T) {
	if len(ListLanguages()) == 0 {
		t.Fatalf("expected at least one chroma lexer registered")
	}
}

func TestNewWithFormatterFallbacks(t *testing.T) {
	// Unknown style and formatter names must fall back rather than panic
	// or return nil. The resulting highlighter should still produce output.
	h := NewWithFormatter("definitely-not-a-style", "definitely-not-a-formatter")
	out := h.Highlight("go", `package main`)
	if out == "" {
		t.Fatalf("expected non-empty output from fallback highlighter")
	}
}
