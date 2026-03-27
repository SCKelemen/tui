package display

import (
	"strings"
	"testing"

	"github.com/SCKelemen/tui/v2/style"
)

func TestHighlightGoKeyword(t *testing.T) {
	h := NewHighlighter(SyntaxLanguageGo)
	out := h.HighlightLine("func main() { return }")
	keyword := style.ANSIBold + style.ANSIBlue + "func" + style.ANSIReset
	if !strings.Contains(out, keyword) {
		t.Fatalf("expected highlighted keyword in %q", out)
	}
}

func TestHighlightString(t *testing.T) {
	h := NewHighlighter(SyntaxLanguageGo)
	out := h.HighlightLine(`name := "sam"`)
	want := style.ANSIGreen + `"sam"` + style.ANSIReset
	if !strings.Contains(out, want) {
		t.Fatalf("expected highlighted string in %q", out)
	}
}

func TestHighlightComment(t *testing.T) {
	h := NewHighlighter(SyntaxLanguageGo)
	out := h.HighlightLine("x := 1 // comment")
	want := style.ANSIDim + "// comment" + style.ANSIReset
	if !strings.Contains(out, want) {
		t.Fatalf("expected highlighted comment in %q", out)
	}
}

func TestHighlightNumber(t *testing.T) {
	h := NewHighlighter(SyntaxLanguageGo)
	out := h.HighlightLine("x := 42")
	want := style.ANSIYellow + "42" + style.ANSIReset
	if !strings.Contains(out, want) {
		t.Fatalf("expected highlighted number in %q", out)
	}
}

func TestHighlightDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     SyntaxLanguage
	}{
		{name: "go", filename: "main.go", want: SyntaxLanguageGo},
		{name: "typescript", filename: "component.ts", want: SyntaxLanguageTypeScript},
		{name: "python", filename: "app.py", want: SyntaxLanguagePython},
		{name: "shell", filename: "run.sh", want: SyntaxLanguageShell},
		{name: "yaml", filename: "config.yml", want: SyntaxLanguageYAML},
		{name: "plain", filename: "README.unknown", want: SyntaxLanguagePlain},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectLanguage(tc.filename)
			if got != tc.want {
				t.Fatalf("DetectLanguage(%q)=%q want %q", tc.filename, got, tc.want)
			}
		})
	}
}
