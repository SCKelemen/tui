package display

import (
	"regexp"
	"strings"
	"testing"
)

var markdownANSICSIRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripMarkdownANSI(s string) string {
	return markdownANSICSIRe.ReplaceAllString(s, "")
}

func TestMarkdownHeadings(t *testing.T) {
	in := "# one\n## two\n### three"
	out := NewMarkdown(in).View()
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	plain0 := stripMarkdownANSI(lines[0])
	plain1 := stripMarkdownANSI(lines[1])
	plain2 := stripMarkdownANSI(lines[2])

	if !strings.Contains(plain0, "ONE") {
		t.Fatalf("expected H1 to be uppercased, got %q", plain0)
	}
	if !strings.Contains(plain1, "two") {
		t.Fatalf("expected H2 text, got %q", plain1)
	}
	if !strings.Contains(plain2, "three") {
		t.Fatalf("expected H3 text, got %q", plain2)
	}

	if strings.Contains(plain0, "#") || strings.Contains(plain1, "##") || strings.Contains(plain2, "###") {
		t.Fatalf("expected heading markers to be removed, got %q", out)
	}
}

func TestMarkdownBoldItalic(t *testing.T) {
	out := NewMarkdown("**bold** and *italic*").View()
	plain := stripMarkdownANSI(out)

	if !strings.Contains(plain, "bold") || !strings.Contains(plain, "italic") {
		t.Fatalf("expected bold and italic text in output: %q", out)
	}
	if strings.Contains(plain, "**") || strings.Contains(plain, "*") {
		t.Fatalf("expected markdown emphasis markers to be removed, got %q", plain)
	}
}

func TestMarkdownInlineCode(t *testing.T) {
	out := NewMarkdown("Use `fmt.Println` now").View()
	plain := stripMarkdownANSI(out)

	if !strings.Contains(plain, " fmt.Println ") {
		t.Fatalf("expected inline code with padded segment, got %q", plain)
	}
	if strings.Contains(plain, "`") {
		t.Fatalf("expected backticks to be removed from inline code, got %q", plain)
	}
}

func TestMarkdownCodeBlock(t *testing.T) {
	in := "```go\nfmt.Println(\"x\")\n```"
	out := NewMarkdown(in).View()
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines for code block with language label, got %d: %q", len(lines), out)
	}

	label := stripMarkdownANSI(lines[0])
	code := stripMarkdownANSI(lines[1])

	if strings.TrimSpace(label) != "go" {
		t.Fatalf("expected language label 'go', got %q", label)
	}
	if !strings.HasPrefix(code, "  ") {
		t.Fatalf("expected code block to be indented, got %q", code)
	}
	if !strings.Contains(code, `fmt.Println("x")`) {
		t.Fatalf("expected code line content, got %q", code)
	}
}

func TestMarkdownBlockquote(t *testing.T) {
	out := NewMarkdown("> quoted line").View()
	plain := stripMarkdownANSI(out)

	if !strings.HasPrefix(plain, "│ ") {
		t.Fatalf("expected blockquote prefix, got %q", plain)
	}
	if !strings.Contains(plain, "quoted line") {
		t.Fatalf("expected quoted content, got %q", plain)
	}
}

func TestMarkdownUnorderedList(t *testing.T) {
	out := NewMarkdown("- item one").View()
	plain := stripMarkdownANSI(out)

	if !strings.HasPrefix(plain, "• ") {
		t.Fatalf("expected bullet prefix, got %q", plain)
	}
	if !strings.Contains(plain, "item one") {
		t.Fatalf("expected list item text, got %q", plain)
	}
}

func TestMarkdownOrderedList(t *testing.T) {
	out := NewMarkdown("12. item one").View()
	plain := stripMarkdownANSI(out)

	if !strings.HasPrefix(plain, "12. ") {
		t.Fatalf("expected ordered list prefix, got %q", plain)
	}
	if !strings.Contains(plain, "item one") {
		t.Fatalf("expected list item text, got %q", plain)
	}
}

func TestMarkdownHR(t *testing.T) {
	out := NewMarkdown("---", WithMarkdownWidth(5)).View()
	plain := stripMarkdownANSI(out)
	want := strings.Repeat("─", 5)

	if plain != want {
		t.Fatalf("expected horizontal rule %q, got %q", want, plain)
	}
}

func TestMarkdownLink(t *testing.T) {
	out := NewMarkdown("[site](https://example.com)").View()

	if !strings.Contains(out, "\x1b]8;;https://example.com\x1b\\") {
		t.Fatalf("expected OSC 8 link opener, got %q", out)
	}
	if !strings.Contains(out, "site") {
		t.Fatalf("expected link text in output, got %q", out)
	}
	if !strings.Contains(out, "\x1b]8;;\x1b\\") {
		t.Fatalf("expected OSC 8 link closer, got %q", out)
	}
}

func TestMarkdownEmpty(t *testing.T) {
	if got := NewMarkdown("").View(); got != "" {
		t.Fatalf("expected empty output, got %q", got)
	}
}

func TestRenderMarkdown(t *testing.T) {
	in := "- item"
	got := RenderMarkdown(in, 0)
	want := NewMarkdown(in, WithMarkdownWidth(0)).View()

	if got != want {
		t.Fatalf("RenderMarkdown mismatch:\n got: %q\nwant: %q", got, want)
	}
	if got == "" {
		t.Fatal("expected non-empty output from RenderMarkdown")
	}
}

func TestDefaultMarkdownTheme(t *testing.T) {
	theme := DefaultMarkdownTheme()

	if theme.HeadingColor == "" || theme.BoldColor == "" || theme.CodeColor == "" || theme.LinkColor == "" || theme.BlockquoteColor == "" || theme.CodeBgColor == "" || theme.HRChar == "" {
		t.Fatalf("expected non-zero default theme, got %+v", theme)
	}
}
