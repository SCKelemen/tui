package tui

import (
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
)

type testSyntaxHighlighter struct {
	highlight func(code string, language string) []HighlightedLine
	supported []string
}

func (tsh testSyntaxHighlighter) Highlight(code string, language string) []HighlightedLine {
	if tsh.highlight != nil {
		return tsh.highlight(code, language)
	}
	return nil
}

func (tsh testSyntaxHighlighter) SupportedLanguages() []string {
	return append([]string(nil), tsh.supported...)
}

func lineText(line HighlightedLine) string {
	var b strings.Builder
	for _, segment := range line.Segments {
		b.WriteString(segment.Text)
	}
	return b.String()
}

func lineHasTokenType(line HighlightedLine, tokenType string) bool {
	for _, segment := range line.Segments {
		if segment.Style.TokenType == tokenType {
			return true
		}
	}
	return false
}

func testThemeConfig(t *testing.T, theme ThemeConfig, name string) {
	t.Helper()
	if theme.Name != name {
		t.Fatalf("theme.Name = %q, want %q", theme.Name, name)
	}
	if theme.Tokens[TokenTypeKeyword].TokenType != TokenTypeKeyword {
		t.Fatalf("keyword TokenType = %q, want %q", theme.Tokens[TokenTypeKeyword].TokenType, TokenTypeKeyword)
	}
	if theme.Tokens[TokenTypePlain].Fg == "" {
		t.Fatal("plain token foreground should not be empty")
	}
}

func TestRegexHighlighter(t *testing.T) {
	h := NewRegexHighlighter(DarkTheme())

	t.Run("go", func(t *testing.T) {
		lines := h.Highlight("func add(x int) int { return 42 } // note", "go")
		if !lineHasTokenType(lines[0], TokenTypeKeyword) || !lineHasTokenType(lines[0], TokenTypeFunction) || !lineHasTokenType(lines[0], TokenTypeNumber) || !lineHasTokenType(lines[0], TokenTypeComment) {
			t.Fatalf("go highlight missing expected token types: %+v", lines[0].Segments)
		}
	})

	t.Run("python", func(t *testing.T) {
		lines := h.Highlight("def add(): return True # note", "python")
		if !lineHasTokenType(lines[0], TokenTypeKeyword) || !lineHasTokenType(lines[0], TokenTypeFunction) || !lineHasTokenType(lines[0], TokenTypeBoolean) || !lineHasTokenType(lines[0], TokenTypeComment) {
			t.Fatalf("python highlight missing expected token types: %+v", lines[0].Segments)
		}
	})

	t.Run("javascript", func(t *testing.T) {
		lines := h.Highlight("function add(){ const x = 1; return x }", "javascript")
		if !lineHasTokenType(lines[0], TokenTypeKeyword) || !lineHasTokenType(lines[0], TokenTypeFunction) || !lineHasTokenType(lines[0], TokenTypeVariable) || !lineHasTokenType(lines[0], TokenTypeNumber) {
			t.Fatalf("javascript highlight missing expected token types: %+v", lines[0].Segments)
		}
	})
}

func TestHighlightedLine(t *testing.T) {
	h := NewRegexHighlighter(DarkTheme())
	line := h.Highlight("func main() {}", "go")[0]
	if got := lineText(line); got != "func main() {}" {
		t.Fatalf("highlighted line text = %q, want original line", got)
	}
	if !lineHasTokenType(line, TokenTypeKeyword) || !lineHasTokenType(line, TokenTypeFunction) {
		t.Fatalf("highlighted line missing keyword/function segments: %+v", line.Segments)
	}
}

func TestDarkTheme(t *testing.T) {
	testThemeConfig(t, DarkTheme(), "dark")
}

func TestThemeDark(t *testing.T) {
	testThemeConfig(t, DarkTheme(), "dark")
}

func TestLightTheme(t *testing.T) {
	testThemeConfig(t, LightTheme(), "light")
}

func TestThemeLight(t *testing.T) {
	testThemeConfig(t, LightTheme(), "light")
}

func TestMonokaiTheme(t *testing.T) {
	testThemeConfig(t, MonokaiTheme(), "monokai")
}

func TestThemeMonokai(t *testing.T) {
	testThemeConfig(t, MonokaiTheme(), "monokai")
}

func TestSupportedLanguages(t *testing.T) {
	h := NewRegexHighlighter(DarkTheme())
	languages := h.SupportedLanguages()
	if !slices.Contains(languages, "go") || !slices.Contains(languages, "python") || !slices.Contains(languages, "javascript") {
		t.Fatalf("SupportedLanguages() missing expected languages: %v", languages)
	}
	if !sort.StringsAreSorted(languages) {
		t.Fatalf("SupportedLanguages() should be sorted, got %v", languages)
	}
}

func TestCompositeSyntaxHighlighter(t *testing.T) {
	primary := testSyntaxHighlighter{
		highlight: func(code string, language string) []HighlightedLine {
			return []HighlightedLine{{Segments: []TextSegment{{Text: code, Style: TokenStyle{TokenType: TokenTypePlain}}}}}
		},
		supported: []string{"go", "python"},
	}
	fallback := testSyntaxHighlighter{
		highlight: func(code string, language string) []HighlightedLine {
			return []HighlightedLine{{Segments: []TextSegment{{Text: "func", Style: TokenStyle{TokenType: TokenTypeKeyword}}, {Text: " main", Style: TokenStyle{TokenType: TokenTypePlain}}}}}
		},
		supported: []string{"javascript", "go"},
	}

	composite := NewCompositeSyntaxHighlighter(primary, fallback)
	lines := composite.Highlight("func main", "go")
	if !lineHasTokenType(lines[0], TokenTypeKeyword) {
		t.Fatalf("composite highlight should fall back to meaningful styling, got %+v", lines[0].Segments)
	}

	languages := composite.SupportedLanguages()
	want := []string{"go", "javascript", "python"}
	if !reflect.DeepEqual(languages, want) {
		t.Fatalf("SupportedLanguages() = %v, want %v", languages, want)
	}
}
