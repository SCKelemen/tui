package markdown

import (
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/SCKelemen/tui/v2/style/design"
)

// ansiRE matches CSI and OSC escape sequences that glamour emits. It is
// deliberately permissive: we only use it to compute visible cell counts and
// to assert that visible substrings made it through rendering.
var ansiRE = regexp.MustCompile("\x1b\\[[0-9;]*[A-Za-z]|\x1b\\][^\x07]*\x07")

func stripANSI(s string) string {
	return ansiRE.ReplaceAllString(s, "")
}

func TestNewRendererDefault(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	out, err := r.RenderString("# Hello\n")
	if err != nil {
		t.Fatalf("RenderString: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
	if !strings.Contains(stripANSI(out), "Hello") {
		t.Fatalf("expected output to contain Hello, got %q", out)
	}
}

func TestRenderBytesEquivalentToString(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	const src = "# Hello\n"
	a, err := r.Render([]byte(src))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	b, err := r.RenderString(src)
	if err != nil {
		t.Fatalf("RenderString: %v", err)
	}
	if a != b {
		t.Fatalf("Render and RenderString diverged:\n  bytes:  %q\n  string: %q", a, b)
	}
}

func TestRenderConstructs(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	source := "# Heading One\n\n" +
		"A paragraph with **bolded** and *emphasized* and `inlinecode` text.\n\n" +
		"```go\nfmt.Println(\"hello\")\n```\n\n" +
		"1. firstitem\n2. seconditem\n\n" +
		"- alphaitem\n- betaitem\n\n" +
		"> quotedtext\n"

	out, err := r.RenderString(source)
	if err != nil {
		t.Fatalf("RenderString: %v", err)
	}
	visible := stripANSI(out)

	for _, want := range []string{
		"Heading One",
		"paragraph",
		"bolded",
		"emphasized",
		"inlinecode",
		"hello",
		"firstitem",
		"seconditem",
		"alphaitem",
		"betaitem",
		"quotedtext",
	} {
		if !strings.Contains(visible, want) {
			t.Errorf("expected output to contain %q\nvisible output:\n%s", want, visible)
		}
	}
}

type mockHighlighter struct {
	calls []mockCall
}

type mockCall struct {
	Language string
	Source   string
}

func (m *mockHighlighter) Highlight(language, source string) string {
	m.calls = append(m.calls, mockCall{Language: language, Source: source})
	return "<<HL:" + language + ":" + source + ">>"
}

func TestWithCodeHighlighter(t *testing.T) {
	h := &mockHighlighter{}
	r, err := NewRenderer(WithCodeHighlighter(h))
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	source := "before paragraph\n\n```go\nfmt.Println(\"hi\")\n```\n\nafter paragraph\n"
	out, err := r.RenderString(source)
	if err != nil {
		t.Fatalf("RenderString: %v", err)
	}

	if len(h.calls) != 1 {
		t.Fatalf("expected exactly one highlighter call, got %d", len(h.calls))
	}
	if got, want := h.calls[0].Language, "go"; got != want {
		t.Errorf("language: got %q want %q", got, want)
	}
	const wantSource = "fmt.Println(\"hi\")"
	if h.calls[0].Source != wantSource {
		t.Errorf("source: got %q want %q", h.calls[0].Source, wantSource)
	}
	if !strings.Contains(out, "<<HL:go:"+wantSource+">>") {
		t.Errorf("expected highlighter output spliced into render, got:\n%s", out)
	}
}

func TestWithWidthBoundsVisibleLines(t *testing.T) {
	const width = 40
	long := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 6)

	r, err := NewRenderer(WithWidth(width))
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	out, err := r.RenderString(long)
	if err != nil {
		t.Fatalf("RenderString: %v", err)
	}

	for i, line := range strings.Split(out, "\n") {
		visible := stripANSI(line)
		if utf8.RuneCountInString(visible) > width {
			t.Errorf("line %d exceeds width %d: %d runes: %q",
				i, width, utf8.RuneCountInString(visible), visible)
		}
	}
}

func TestWithStyleDarkAndLight(t *testing.T) {
	for _, name := range []string{"dark", "light"} {
		if _, err := NewRenderer(WithStyle(name)); err != nil {
			t.Errorf("WithStyle(%q): %v", name, err)
		}
	}
}

func TestWithStyleUnknownErrors(t *testing.T) {
	if _, err := NewRenderer(WithStyle("definitely-not-a-real-style")); err == nil {
		t.Fatal("expected error for unknown style, got nil")
	}
}

func TestWithDesignTokens(t *testing.T) {
	if _, err := NewRenderer(WithDesignTokens(&design.DesignTokens{Mode: "dark"})); err != nil {
		t.Errorf("dark tokens: %v", err)
	}
	if _, err := NewRenderer(WithDesignTokens(&design.DesignTokens{Mode: "light"})); err != nil {
		t.Errorf("light tokens: %v", err)
	}
	if _, err := NewRenderer(WithDesignTokens(nil)); err != nil {
		t.Errorf("nil tokens: %v", err)
	}
}
