package tui

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	TokenTypePlain    = "plain"
	TokenTypeKeyword  = "keyword"
	TokenTypeString   = "string"
	TokenTypeComment  = "comment"
	TokenTypeNumber   = "number"
	TokenTypeFunction = "function"
	TokenTypeType     = "type"
	TokenTypeOperator = "operator"
	TokenTypeProperty = "property"
	TokenTypeTag      = "tag"
	TokenTypeAttr     = "attr"
	TokenTypeBoolean  = "boolean"
	TokenTypeVariable = "variable"
	TokenTypeSelector = "selector"
)

// SyntaxHighlighter highlights code into structured text segments.
type SyntaxHighlighter interface {
	Highlight(code string, language string) []HighlightedLine
	SupportedLanguages() []string
}

// HighlightedLine represents one highlighted source line.
type HighlightedLine struct {
	Segments []TextSegment
}

// TextSegment represents a contiguous styled span of text.
type TextSegment struct {
	Text  string
	Style TokenStyle
}

// TokenStyle describes how a token should be rendered.
type TokenStyle struct {
	Fg        string
	Bg        string
	Bold      bool
	Italic    bool
	Underline bool
	TokenType string
}

// ThemeConfig maps token types to visual styles.
type ThemeConfig struct {
	Name   string
	Tokens map[string]TokenStyle
}

// RegexHighlighter provides a built-in regex-based syntax highlighter.
type RegexHighlighter struct {
	theme     ThemeConfig
	grammars  map[string]languageGrammar
	aliases   map[string]string
	supported []string
}

// CompositeSyntaxHighlighter tries the primary highlighter first and falls back when needed.
// This allows a future tree-sitter-backed highlighter to share the same interface and gracefully
// fall back to the built-in regex implementation.
type CompositeSyntaxHighlighter struct {
	Primary  SyntaxHighlighter
	Fallback SyntaxHighlighter
}

type languageGrammar struct {
	Patterns []tokenPattern
}

type tokenPattern struct {
	Regex     *regexp.Regexp
	TokenType string
	Group     int
}

type tokenMatch struct {
	Start     int
	End       int
	TokenType string
}

// NewRegexHighlighter creates a regex highlighter using the supplied theme.
func NewRegexHighlighter(theme ThemeConfig) *RegexHighlighter {
	if len(theme.Tokens) == 0 {
		theme = DarkTheme()
	}

	grammars, aliases := defaultRegexGrammars()
	supported := make([]string, 0, len(grammars))
	for language := range grammars {
		supported = append(supported, language)
	}
	sort.Strings(supported)

	return &RegexHighlighter{
		theme:     theme,
		grammars:  grammars,
		aliases:   aliases,
		supported: supported,
	}
}

// NewCompositeSyntaxHighlighter creates a highlighter that falls back when the primary cannot highlight.
func NewCompositeSyntaxHighlighter(primary, fallback SyntaxHighlighter) *CompositeSyntaxHighlighter {
	return &CompositeSyntaxHighlighter{Primary: primary, Fallback: fallback}
}

// SupportedLanguages returns the languages supported by the regex highlighter.
func (h *RegexHighlighter) SupportedLanguages() []string {
	if h == nil {
		return nil
	}
	languages := make([]string, len(h.supported))
	copy(languages, h.supported)
	return languages
}

// SupportedLanguages returns the union of primary and fallback support.
func (h *CompositeSyntaxHighlighter) SupportedLanguages() []string {
	if h == nil {
		return nil
	}

	seen := map[string]struct{}{}
	languages := make([]string, 0)
	for _, highlighter := range []SyntaxHighlighter{h.Primary, h.Fallback} {
		if highlighter == nil {
			continue
		}
		for _, language := range highlighter.SupportedLanguages() {
			normalized := strings.ToLower(strings.TrimSpace(language))
			if normalized == "" {
				continue
			}
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			languages = append(languages, normalized)
		}
	}

	sort.Strings(languages)
	return languages
}

// Highlight highlights code using the regex grammar for the requested language.
func (h *RegexHighlighter) Highlight(code string, language string) []HighlightedLine {
	lines := strings.Split(code, "\n")
	if code == "" {
		lines = []string{""}
	}

	result := make([]HighlightedLine, len(lines))
	grammar, ok := h.lookupGrammar(language)
	for i, line := range lines {
		result[i] = h.highlightLine(line, grammar, ok)
	}
	return result
}

// Highlight highlights with the primary highlighter and falls back when it produces no styling.
func (h *CompositeSyntaxHighlighter) Highlight(code string, language string) []HighlightedLine {
	if h == nil {
		return nil
	}

	if h.Primary != nil {
		highlighted := h.Primary.Highlight(code, language)
		if hasMeaningfulHighlighting(highlighted) {
			return highlighted
		}
	}

	if h.Fallback != nil {
		return h.Fallback.Highlight(code, language)
	}

	return nil
}

// DetectSyntaxLanguage attempts to infer a supported language from a filename.
func DetectSyntaxLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(filename)))
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js", ".mjs", ".cjs", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".rs":
		return "rust"
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".sh", ".bash", ".zsh":
		return "bash"
	case ".sql":
		return "sql"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	default:
		return ""
	}
}

func (h *RegexHighlighter) lookupGrammar(language string) (languageGrammar, bool) {
	if h == nil {
		return languageGrammar{}, false
	}

	normalized := strings.ToLower(strings.TrimSpace(language))
	if normalized == "" {
		return languageGrammar{}, false
	}

	if canonical, ok := h.aliases[normalized]; ok {
		grammar, ok := h.grammars[canonical]
		return grammar, ok
	}

	if detected := DetectSyntaxLanguage(normalized); detected != "" {
		grammar, ok := h.grammars[detected]
		return grammar, ok
	}

	grammar, ok := h.grammars[normalized]
	return grammar, ok
}

func (h *RegexHighlighter) highlightLine(line string, grammar languageGrammar, ok bool) HighlightedLine {
	if !ok || line == "" {
		return HighlightedLine{Segments: []TextSegment{{Text: line, Style: h.styleFor(TokenTypePlain)}}}
	}

	matches := collectMatches(line, grammar.Patterns)
	if len(matches) == 0 {
		return HighlightedLine{Segments: []TextSegment{{Text: line, Style: h.styleFor(TokenTypePlain)}}}
	}

	segments := make([]TextSegment, 0, len(matches)*2+1)
	cursor := 0
	for _, match := range matches {
		if match.Start > cursor {
			segments = append(segments, TextSegment{Text: line[cursor:match.Start], Style: h.styleFor(TokenTypePlain)})
		}
		segments = append(segments, TextSegment{Text: line[match.Start:match.End], Style: h.styleFor(match.TokenType)})
		cursor = match.End
	}
	if cursor < len(line) {
		segments = append(segments, TextSegment{Text: line[cursor:], Style: h.styleFor(TokenTypePlain)})
	}

	return HighlightedLine{Segments: mergeAdjacentSegments(segments)}
}

func (h *RegexHighlighter) styleFor(tokenType string) TokenStyle {
	if h == nil {
		return TokenStyle{TokenType: tokenType}
	}

	if style, ok := h.theme.Tokens[tokenType]; ok {
		if style.TokenType == "" {
			style.TokenType = tokenType
		}
		return style
	}

	if style, ok := h.theme.Tokens[TokenTypePlain]; ok {
		style.TokenType = tokenType
		return style
	}

	return TokenStyle{TokenType: tokenType}
}

func hasMeaningfulHighlighting(lines []HighlightedLine) bool {
	for _, line := range lines {
		for _, segment := range line.Segments {
			style := segment.Style
			if style.TokenType != "" && style.TokenType != TokenTypePlain {
				return true
			}
			if style.Fg != "" || style.Bg != "" || style.Bold || style.Italic || style.Underline {
				return true
			}
		}
	}
	return false
}

func collectMatches(line string, patterns []tokenPattern) []tokenMatch {
	if line == "" || len(patterns) == 0 {
		return nil
	}

	taken := make([]bool, len(line))
	matches := make([]tokenMatch, 0)

	for _, pattern := range patterns {
		if pattern.Regex == nil {
			continue
		}

		for _, submatch := range pattern.Regex.FindAllStringSubmatchIndex(line, -1) {
			start, end, ok := submatchRange(submatch, pattern.Group)
			if !ok || start >= end {
				continue
			}
			if overlaps(taken, start, end) {
				continue
			}
			markTaken(taken, start, end)
			matches = append(matches, tokenMatch{Start: start, End: end, TokenType: pattern.TokenType})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Start == matches[j].Start {
			return matches[i].End < matches[j].End
		}
		return matches[i].Start < matches[j].Start
	})

	return matches
}

func submatchRange(indices []int, group int) (int, int, bool) {
	if group < 0 {
		group = 0
	}
	index := group * 2
	if len(indices) <= index+1 {
		return 0, 0, false
	}
	start, end := indices[index], indices[index+1]
	if start < 0 || end < 0 {
		return 0, 0, false
	}
	return start, end, true
}

func overlaps(taken []bool, start int, end int) bool {
	for i := start; i < end; i++ {
		if taken[i] {
			return true
		}
	}
	return false
}

func markTaken(taken []bool, start int, end int) {
	for i := start; i < end; i++ {
		taken[i] = true
	}
}

func mergeAdjacentSegments(segments []TextSegment) []TextSegment {
	if len(segments) == 0 {
		return nil
	}

	merged := make([]TextSegment, 0, len(segments))
	for _, segment := range segments {
		if segment.Text == "" {
			continue
		}
		if len(merged) == 0 {
			merged = append(merged, segment)
			continue
		}
		last := &merged[len(merged)-1]
		if last.Style == segment.Style {
			last.Text += segment.Text
			continue
		}
		merged = append(merged, segment)
	}
	return merged
}

func defaultRegexGrammars() (map[string]languageGrammar, map[string]string) {
	grammars := map[string]languageGrammar{
		"go": {
			Patterns: []tokenPattern{
				{Regex: regexp.MustCompile(`//.*$`), TokenType: TokenTypeComment},
				{Regex: regexp.MustCompile("`[^`]*`|\"(?:\\\\.|[^\"\\\\])*\"|'(?:\\\\.|[^'\\\\])*'"), TokenType: TokenTypeString},
				{Regex: regexp.MustCompile(`\b(?:0[xX][0-9a-fA-F]+|\d+(?:\.\d+)?)\b`), TokenType: TokenTypeNumber},
				{Regex: regexp.MustCompile(`\b(true|false|nil|iota)\b`), TokenType: TokenTypeBoolean},
				{Regex: compileWords([]string{"break", "case", "chan", "const", "continue", "default", "defer", "else", "fallthrough", "for", "func", "go", "goto", "if", "import", "interface", "map", "package", "range", "return", "select", "struct", "switch", "type", "var"}), TokenType: TokenTypeKeyword},
				{Regex: compileWords([]string{"any", "bool", "byte", "comparable", "complex64", "complex128", "error", "float32", "float64", "int", "int8", "int16", "int32", "int64", "rune", "string", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr"}), TokenType: TokenTypeType},
				{Regex: regexp.MustCompile(`\bfunc\s*(?:\([^)]*\)\s*)?([A-Za-z_][A-Za-z0-9_]*)`), TokenType: TokenTypeFunction, Group: 1},
				{Regex: regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\s*:=`), TokenType: TokenTypeVariable, Group: 1},
				{Regex: regexp.MustCompile(`==|!=|<=|>=|:=|&&|\|\||<-|[+\-*/%=&|^!<>]`), TokenType: TokenTypeOperator},
			},
		},
		"python": {
			Patterns: []tokenPattern{
				{Regex: regexp.MustCompile(`#.*$`), TokenType: TokenTypeComment},
				{Regex: regexp.MustCompile(`""".*?"""|'''.*?'''|"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'`), TokenType: TokenTypeString},
				{Regex: regexp.MustCompile(`\b(?:0[xX][0-9a-fA-F]+|\d+(?:\.\d+)?)\b`), TokenType: TokenTypeNumber},
				{Regex: regexp.MustCompile(`\b(True|False|None)\b`), TokenType: TokenTypeBoolean},
				{Regex: compileWords([]string{"and", "as", "assert", "async", "await", "break", "class", "continue", "def", "del", "elif", "else", "except", "finally", "for", "from", "global", "if", "import", "in", "is", "lambda", "nonlocal", "not", "or", "pass", "raise", "return", "try", "while", "with", "yield"}), TokenType: TokenTypeKeyword},
				{Regex: compileWords([]string{"bool", "bytes", "dict", "float", "frozenset", "int", "list", "set", "str", "tuple"}), TokenType: TokenTypeType},
				{Regex: regexp.MustCompile(`\bdef\s+([A-Za-z_][A-Za-z0-9_]*)`), TokenType: TokenTypeFunction, Group: 1},
				{Regex: regexp.MustCompile(`==|!=|<=|>=|//=|\*\*|:=|[+\-*/%=&|^~<>]`), TokenType: TokenTypeOperator},
			},
		},
		"javascript": {
			Patterns: []tokenPattern{
				{Regex: regexp.MustCompile(`//.*$|/\*.*\*/`), TokenType: TokenTypeComment},
				{Regex: regexp.MustCompile("`(?:\\\\.|[^`\\\\])*`|\"(?:\\\\.|[^\"\\\\])*\"|'(?:\\\\.|[^'\\\\])*'"), TokenType: TokenTypeString},
				{Regex: regexp.MustCompile(`\b(?:0[xX][0-9a-fA-F]+|\d+(?:\.\d+)?)\b`), TokenType: TokenTypeNumber},
				{Regex: regexp.MustCompile(`\b(true|false|null|undefined)\b`), TokenType: TokenTypeBoolean},
				{Regex: compileWords([]string{"async", "await", "break", "case", "catch", "class", "const", "continue", "default", "delete", "do", "else", "export", "extends", "finally", "for", "function", "if", "import", "in", "instanceof", "let", "new", "return", "super", "switch", "this", "throw", "try", "typeof", "var", "while", "yield"}), TokenType: TokenTypeKeyword},
				{Regex: compileWords([]string{"Array", "BigInt", "Boolean", "Date", "Map", "Number", "Object", "Promise", "RegExp", "Set", "String", "Symbol"}), TokenType: TokenTypeType},
				{Regex: regexp.MustCompile(`\bfunction\s+([A-Za-z_$][A-Za-z0-9_$]*)`), TokenType: TokenTypeFunction, Group: 1},
				{Regex: regexp.MustCompile(`\bclass\s+([A-Za-z_$][A-Za-z0-9_$]*)`), TokenType: TokenTypeType, Group: 1},
				{Regex: regexp.MustCompile(`\b(?:const|let|var)\s+([A-Za-z_$][A-Za-z0-9_$]*)`), TokenType: TokenTypeVariable, Group: 1},
				{Regex: regexp.MustCompile(`===|!==|==|!=|<=|>=|=>|&&|\|\||\?\?|\+\+|--|[+\-*/%=&|^!<>?:~]`), TokenType: TokenTypeOperator},
			},
		},
		"typescript": {
			Patterns: []tokenPattern{
				{Regex: regexp.MustCompile(`//.*$|/\*.*\*/`), TokenType: TokenTypeComment},
				{Regex: regexp.MustCompile("`(?:\\\\.|[^`\\\\])*`|\"(?:\\\\.|[^\"\\\\])*\"|'(?:\\\\.|[^'\\\\])*'"), TokenType: TokenTypeString},
				{Regex: regexp.MustCompile(`\b(?:0[xX][0-9a-fA-F]+|\d+(?:\.\d+)?)\b`), TokenType: TokenTypeNumber},
				{Regex: regexp.MustCompile(`\b(true|false|null|undefined)\b`), TokenType: TokenTypeBoolean},
				{Regex: compileWords([]string{"abstract", "any", "as", "asserts", "async", "await", "break", "case", "catch", "class", "const", "continue", "declare", "default", "do", "else", "enum", "export", "extends", "finally", "for", "function", "if", "implements", "import", "in", "infer", "interface", "is", "keyof", "let", "module", "namespace", "new", "private", "protected", "public", "readonly", "return", "satisfies", "super", "switch", "this", "throw", "try", "type", "typeof", "var", "while"}), TokenType: TokenTypeKeyword},
				{Regex: compileWords([]string{"Array", "Promise", "Record", "ReadonlyArray", "boolean", "never", "number", "object", "string", "symbol", "unknown", "void"}), TokenType: TokenTypeType},
				{Regex: regexp.MustCompile(`\bfunction\s+([A-Za-z_$][A-Za-z0-9_$]*)`), TokenType: TokenTypeFunction, Group: 1},
				{Regex: regexp.MustCompile(`\b(?:interface|type|class|enum)\s+([A-Za-z_$][A-Za-z0-9_$]*)`), TokenType: TokenTypeType, Group: 1},
				{Regex: regexp.MustCompile(`\b(?:const|let|var)\s+([A-Za-z_$][A-Za-z0-9_$]*)`), TokenType: TokenTypeVariable, Group: 1},
				{Regex: regexp.MustCompile(`===|!==|==|!=|<=|>=|=>|&&|\|\||\?\?|\+\+|--|[+\-*/%=&|^!<>?:~]`), TokenType: TokenTypeOperator},
			},
		},
		"rust": {
			Patterns: []tokenPattern{
				{Regex: regexp.MustCompile(`//.*$|/\*.*\*/`), TokenType: TokenTypeComment},
				{Regex: regexp.MustCompile("b?\"(?:\\\\.|[^\"\\\\])*\"|r#*\".*?\"#*|'(?:\\\\.|[^'\\\\])*'"), TokenType: TokenTypeString},
				{Regex: regexp.MustCompile(`\b(?:0[xX][0-9a-fA-F]+|\d+(?:\.\d+)?(?:_[0-9]+)*(?:[iu](?:8|16|32|64|128|size)|f(?:32|64))?)\b`), TokenType: TokenTypeNumber},
				{Regex: regexp.MustCompile(`\b(true|false|None|Some|Ok|Err)\b`), TokenType: TokenTypeBoolean},
				{Regex: compileWords([]string{"as", "async", "await", "break", "const", "continue", "crate", "dyn", "else", "enum", "extern", "fn", "for", "if", "impl", "in", "let", "loop", "match", "mod", "move", "mut", "pub", "ref", "return", "self", "Self", "static", "struct", "super", "trait", "type", "unsafe", "use", "where", "while"}), TokenType: TokenTypeKeyword},
				{Regex: compileWords([]string{"String", "Vec", "bool", "char", "f32", "f64", "i8", "i16", "i32", "i64", "i128", "isize", "str", "u8", "u16", "u32", "u64", "u128", "usize", "Option", "Result"}), TokenType: TokenTypeType},
				{Regex: regexp.MustCompile(`\bfn\s+([A-Za-z_][A-Za-z0-9_]*)`), TokenType: TokenTypeFunction, Group: 1},
				{Regex: regexp.MustCompile(`\b(?:struct|enum|trait|type|impl)\s+([A-Za-z_][A-Za-z0-9_]*)`), TokenType: TokenTypeType, Group: 1},
				{Regex: regexp.MustCompile(`==|!=|<=|>=|=>|->|&&|\|\||::|[+\-*/%=&|^!<>?:]`), TokenType: TokenTypeOperator},
			},
		},
		"yaml": {
			Patterns: []tokenPattern{
				{Regex: regexp.MustCompile(`#.*$`), TokenType: TokenTypeComment},
				{Regex: regexp.MustCompile(`"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'`), TokenType: TokenTypeString},
				{Regex: regexp.MustCompile(`\b(?:true|false|null|yes|no|on|off)\b`), TokenType: TokenTypeBoolean},
				{Regex: regexp.MustCompile(`\b(?:0[xX][0-9a-fA-F]+|\d+(?:\.\d+)?)\b`), TokenType: TokenTypeNumber},
				{Regex: regexp.MustCompile(`^\s*[-?]\s+`), TokenType: TokenTypeOperator},
				{Regex: regexp.MustCompile(`^\s*([A-Za-z0-9_.-]+)\s*:`), TokenType: TokenTypeProperty, Group: 1},
				{Regex: regexp.MustCompile(`&[A-Za-z0-9_-]+|\*[A-Za-z0-9_-]+`), TokenType: TokenTypeVariable},
			},
		},
		"json": {
			Patterns: []tokenPattern{
				{Regex: regexp.MustCompile(`"(?:\\.|[^"\\])*"\s*:`), TokenType: TokenTypeProperty},
				{Regex: regexp.MustCompile(`"(?:\\.|[^"\\])*"`), TokenType: TokenTypeString},
				{Regex: regexp.MustCompile(`\b(?:true|false|null)\b`), TokenType: TokenTypeBoolean},
				{Regex: regexp.MustCompile(`\b(?:-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)\b`), TokenType: TokenTypeNumber},
				{Regex: regexp.MustCompile(`[:{},\[\]]`), TokenType: TokenTypeOperator},
			},
		},
		"bash": {
			Patterns: []tokenPattern{
				{Regex: regexp.MustCompile(`#.*$`), TokenType: TokenTypeComment},
				{Regex: regexp.MustCompile("\"(?:\\\\.|[^\"\\\\])*\"|'(?:\\\\.|[^'\\\\])*'"), TokenType: TokenTypeString},
				{Regex: regexp.MustCompile(`\$\{[^}]+\}|\$[A-Za-z_][A-Za-z0-9_]*|\$[0-9@#?*!$-]`), TokenType: TokenTypeVariable},
				{Regex: regexp.MustCompile(`\b(?:0[xX][0-9a-fA-F]+|\d+)\b`), TokenType: TokenTypeNumber},
				{Regex: compileWords([]string{"if", "then", "else", "elif", "fi", "for", "while", "in", "do", "done", "case", "esac", "function", "select", "time", "coproc", "until"}), TokenType: TokenTypeKeyword},
				{Regex: compileWords([]string{"alias", "bg", "bind", "builtin", "cd", "command", "echo", "eval", "exec", "exit", "export", "jobs", "kill", "printf", "pwd", "read", "readonly", "set", "shift", "source", "test", "trap", "type", "ulimit", "umask", "unset", "wait"}), TokenType: TokenTypeFunction},
				{Regex: regexp.MustCompile(`\|\||&&|>>|<<|[|&;<>=()]`), TokenType: TokenTypeOperator},
			},
		},
		"sql": {
			Patterns: []tokenPattern{
				{Regex: regexp.MustCompile(`--.*$|/\*.*\*/`), TokenType: TokenTypeComment},
				{Regex: regexp.MustCompile(`"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'`), TokenType: TokenTypeString},
				{Regex: regexp.MustCompile(`\b(?:-?\d+(?:\.\d+)?)\b`), TokenType: TokenTypeNumber},
				{Regex: regexp.MustCompile(`(?i)\b(true|false|null)\b`), TokenType: TokenTypeBoolean},
				{Regex: compileWordsCaseInsensitive([]string{"alter", "and", "as", "by", "case", "create", "delete", "desc", "distinct", "drop", "else", "end", "exists", "from", "group", "having", "in", "inner", "insert", "into", "join", "left", "limit", "not", "null", "offset", "on", "or", "order", "outer", "primary", "right", "select", "set", "table", "then", "union", "update", "values", "when", "where"}), TokenType: TokenTypeKeyword},
				{Regex: compileWordsCaseInsensitive([]string{"bigint", "blob", "boolean", "date", "datetime", "decimal", "double", "float", "integer", "json", "jsonb", "numeric", "real", "serial", "text", "time", "timestamp", "uuid", "varchar"}), TokenType: TokenTypeType},
				{Regex: regexp.MustCompile(`(?i)\b(?:from|join|into|table|update)\s+([A-Za-z_][A-Za-z0-9_\.]*)`), TokenType: TokenTypeFunction, Group: 1},
				{Regex: regexp.MustCompile(`<>|!=|<=|>=|:=|[+\-*/%=&|<>]`), TokenType: TokenTypeOperator},
			},
		},
		"html": {
			Patterns: []tokenPattern{
				{Regex: regexp.MustCompile(`<!--.*?-->`), TokenType: TokenTypeComment},
				{Regex: regexp.MustCompile(`</?([A-Za-z][A-Za-z0-9:-]*)`), TokenType: TokenTypeTag, Group: 1},
				{Regex: regexp.MustCompile(`\s([A-Za-z_:][-A-Za-z0-9_:.]*)=`), TokenType: TokenTypeAttr, Group: 1},
				{Regex: regexp.MustCompile(`"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'`), TokenType: TokenTypeString},
				{Regex: regexp.MustCompile(`[<>/=]+`), TokenType: TokenTypeOperator},
			},
		},
		"css": {
			Patterns: []tokenPattern{
				{Regex: regexp.MustCompile(`/\*.*\*/`), TokenType: TokenTypeComment},
				{Regex: regexp.MustCompile(`"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'`), TokenType: TokenTypeString},
				{Regex: regexp.MustCompile(`(^|\s)([#.]?[A-Za-z_-][A-Za-z0-9_-]*)\s*\{`), TokenType: TokenTypeSelector, Group: 2},
				{Regex: regexp.MustCompile(`\b([a-z-]+)\s*:`), TokenType: TokenTypeProperty, Group: 1},
				{Regex: regexp.MustCompile(`\b(?:#[0-9a-fA-F]{3,8}|\d+(?:\.\d+)?(?:px|em|rem|vh|vw|%)?)\b`), TokenType: TokenTypeNumber},
				{Regex: compileWords([]string{"important", "inherit", "initial", "unset", "auto", "block", "flex", "grid", "inline", "none", "relative", "absolute", "fixed"}), TokenType: TokenTypeKeyword},
				{Regex: regexp.MustCompile(`[:;{},()]`), TokenType: TokenTypeOperator},
			},
		},
	}

	aliases := map[string]string{
		"bash":       "bash",
		"css":        "css",
		"go":         "go",
		"golang":     "go",
		"html":       "html",
		"htm":        "html",
		"javascript": "javascript",
		"js":         "javascript",
		"json":       "json",
		"python":     "python",
		"py":         "python",
		"rust":       "rust",
		"rs":         "rust",
		"shell":      "bash",
		"sh":         "bash",
		"sql":        "sql",
		"ts":         "typescript",
		"tsx":        "typescript",
		"typescript": "typescript",
		"yaml":       "yaml",
		"yml":        "yaml",
		"zsh":        "bash",
	}

	return grammars, aliases
}

func compileWords(words []string) *regexp.Regexp {
	if len(words) == 0 {
		return nil
	}
	escaped := make([]string, len(words))
	for i := range words {
		escaped[i] = regexp.QuoteMeta(words[i])
	}
	return regexp.MustCompile(`\b(?:` + strings.Join(escaped, "|") + `)\b`)
}

func compileWordsCaseInsensitive(words []string) *regexp.Regexp {
	if len(words) == 0 {
		return nil
	}
	escaped := make([]string, len(words))
	for i := range words {
		escaped[i] = regexp.QuoteMeta(words[i])
	}
	return regexp.MustCompile(`(?i)\b(?:` + strings.Join(escaped, "|") + `)\b`)
}

// DarkTheme returns a VS Code dark-inspired theme.
func DarkTheme() ThemeConfig {
	return ThemeConfig{
		Name: "dark",
		Tokens: map[string]TokenStyle{
			TokenTypePlain:    {Fg: "#D4D4D4", TokenType: TokenTypePlain},
			TokenTypeKeyword:  {Fg: "#C586C0", Bold: true, TokenType: TokenTypeKeyword},
			TokenTypeString:   {Fg: "#CE9178", TokenType: TokenTypeString},
			TokenTypeComment:  {Fg: "#6A9955", Italic: true, TokenType: TokenTypeComment},
			TokenTypeNumber:   {Fg: "#B5CEA8", TokenType: TokenTypeNumber},
			TokenTypeFunction: {Fg: "#DCDCAA", TokenType: TokenTypeFunction},
			TokenTypeType:     {Fg: "#4EC9B0", TokenType: TokenTypeType},
			TokenTypeOperator: {Fg: "#D4D4D4", TokenType: TokenTypeOperator},
			TokenTypeProperty: {Fg: "#9CDCFE", TokenType: TokenTypeProperty},
			TokenTypeTag:      {Fg: "#569CD6", TokenType: TokenTypeTag},
			TokenTypeAttr:     {Fg: "#9CDCFE", TokenType: TokenTypeAttr},
			TokenTypeBoolean:  {Fg: "#569CD6", Bold: true, TokenType: TokenTypeBoolean},
			TokenTypeVariable: {Fg: "#9CDCFE", TokenType: TokenTypeVariable},
			TokenTypeSelector: {Fg: "#D7BA7D", TokenType: TokenTypeSelector},
		},
	}
}

// LightTheme returns a VS Code light-inspired theme.
func LightTheme() ThemeConfig {
	return ThemeConfig{
		Name: "light",
		Tokens: map[string]TokenStyle{
			TokenTypePlain:    {Fg: "#24292E", TokenType: TokenTypePlain},
			TokenTypeKeyword:  {Fg: "#AF00DB", Bold: true, TokenType: TokenTypeKeyword},
			TokenTypeString:   {Fg: "#A31515", TokenType: TokenTypeString},
			TokenTypeComment:  {Fg: "#008000", Italic: true, TokenType: TokenTypeComment},
			TokenTypeNumber:   {Fg: "#098658", TokenType: TokenTypeNumber},
			TokenTypeFunction: {Fg: "#795E26", TokenType: TokenTypeFunction},
			TokenTypeType:     {Fg: "#267F99", TokenType: TokenTypeType},
			TokenTypeOperator: {Fg: "#000000", TokenType: TokenTypeOperator},
			TokenTypeProperty: {Fg: "#001080", TokenType: TokenTypeProperty},
			TokenTypeTag:      {Fg: "#800000", TokenType: TokenTypeTag},
			TokenTypeAttr:     {Fg: "#FF0000", TokenType: TokenTypeAttr},
			TokenTypeBoolean:  {Fg: "#0000FF", Bold: true, TokenType: TokenTypeBoolean},
			TokenTypeVariable: {Fg: "#001080", TokenType: TokenTypeVariable},
			TokenTypeSelector: {Fg: "#800000", TokenType: TokenTypeSelector},
		},
	}
}

// MonokaiTheme returns a Monokai-inspired theme.
func MonokaiTheme() ThemeConfig {
	return ThemeConfig{
		Name: "monokai",
		Tokens: map[string]TokenStyle{
			TokenTypePlain:    {Fg: "#F8F8F2", TokenType: TokenTypePlain},
			TokenTypeKeyword:  {Fg: "#F92672", Bold: true, TokenType: TokenTypeKeyword},
			TokenTypeString:   {Fg: "#E6DB74", TokenType: TokenTypeString},
			TokenTypeComment:  {Fg: "#75715E", Italic: true, TokenType: TokenTypeComment},
			TokenTypeNumber:   {Fg: "#AE81FF", TokenType: TokenTypeNumber},
			TokenTypeFunction: {Fg: "#A6E22E", TokenType: TokenTypeFunction},
			TokenTypeType:     {Fg: "#66D9EF", TokenType: TokenTypeType},
			TokenTypeOperator: {Fg: "#F92672", TokenType: TokenTypeOperator},
			TokenTypeProperty: {Fg: "#A6E22E", TokenType: TokenTypeProperty},
			TokenTypeTag:      {Fg: "#F92672", TokenType: TokenTypeTag},
			TokenTypeAttr:     {Fg: "#A6E22E", TokenType: TokenTypeAttr},
			TokenTypeBoolean:  {Fg: "#AE81FF", Bold: true, TokenType: TokenTypeBoolean},
			TokenTypeVariable: {Fg: "#FD971F", TokenType: TokenTypeVariable},
			TokenTypeSelector: {Fg: "#A6E22E", TokenType: TokenTypeSelector},
		},
	}
}
