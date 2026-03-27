package display

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/SCKelemen/tui/v2/style"
)

// SyntaxLanguage identifies the source language used for highlighting.
type SyntaxLanguage string

const (
	SyntaxLanguageGo         SyntaxLanguage = "Go"
	SyntaxLanguageJavaScript SyntaxLanguage = "JavaScript"
	SyntaxLanguageTypeScript SyntaxLanguage = "TypeScript"
	SyntaxLanguagePython     SyntaxLanguage = "Python"
	SyntaxLanguageRust       SyntaxLanguage = "Rust"
	SyntaxLanguageSQL        SyntaxLanguage = "SQL"
	SyntaxLanguageShell      SyntaxLanguage = "Shell"
	SyntaxLanguageJSON       SyntaxLanguage = "JSON"
	SyntaxLanguageYAML       SyntaxLanguage = "YAML"
	SyntaxLanguageMarkdown   SyntaxLanguage = "Markdown"
	SyntaxLanguagePlain      SyntaxLanguage = "Plain"
)

// Highlighter is a simple regex-based syntax highlighter for terminal output.
type Highlighter struct {
	language SyntaxLanguage

	keywords  *regexp.Regexp
	types     *regexp.Regexp
	strings   *regexp.Regexp
	comments  *regexp.Regexp
	numbers   *regexp.Regexp
	operators *regexp.Regexp
}

// NewHighlighter creates a language-specific regex highlighter.
func NewHighlighter(lang SyntaxLanguage) *Highlighter {
	h := &Highlighter{
		language:  normalizeLanguage(lang),
		strings:   regexp.MustCompile("`[^`]*`|\"(?:\\\\.|[^\"\\\\])*\"|'(?:\\\\.|[^'\\\\])*'"),
		comments:  regexp.MustCompile(`//.*$|#.*$`),
		numbers:   regexp.MustCompile(`\b\d+\b`),
		operators: regexp.MustCompile(`==|!=|<=|>=|:=|&&|\|\||[=+\-*/%<>!&|]`),
	}

	keywordPatterns := map[SyntaxLanguage][]string{
		SyntaxLanguageGo:         {"func", "return", "if", "else", "for", "switch", "case", "import", "package", "type", "struct", "interface", "const", "var", "range", "go", "defer", "map", "select", "chan", "default"},
		SyntaxLanguageJavaScript: {"function", "return", "if", "else", "for", "switch", "case", "import", "export", "const", "let", "var", "class", "extends", "new", "try", "catch", "finally", "async", "await"},
		SyntaxLanguageTypeScript: {"function", "return", "if", "else", "for", "switch", "case", "import", "export", "const", "let", "var", "class", "extends", "new", "interface", "type", "implements", "enum", "namespace", "as", "readonly", "public", "private", "protected", "async", "await"},
		SyntaxLanguagePython:     {"def", "return", "if", "elif", "else", "for", "while", "import", "from", "class", "try", "except", "finally", "with", "as", "lambda", "pass", "yield", "async", "await"},
		SyntaxLanguageRust:       {"fn", "let", "mut", "impl", "trait", "struct", "enum", "mod", "use", "pub", "crate", "super", "self", "match", "if", "else", "for", "while", "loop", "return"},
		SyntaxLanguageSQL:        {"select", "from", "where", "insert", "into", "update", "delete", "join", "left", "right", "inner", "outer", "group", "by", "order", "limit", "having", "as", "on", "and", "or", "not", "null"},
		SyntaxLanguageShell:      {"if", "then", "else", "fi", "for", "do", "done", "case", "esac", "function", "in", "export", "local", "readonly", "return"},
		SyntaxLanguageJSON:       {"true", "false", "null"},
		SyntaxLanguageYAML:       {"true", "false", "null"},
	}

	typePatterns := map[SyntaxLanguage][]string{
		SyntaxLanguageGo:         {"string", "int", "int64", "int32", "bool", "float64", "float32", "error", "byte", "rune", "any"},
		SyntaxLanguageJavaScript: {"string", "number", "boolean", "object", "symbol", "bigint", "undefined"},
		SyntaxLanguageTypeScript: {"string", "number", "boolean", "any", "unknown", "never", "void", "object", "Record", "Array"},
		SyntaxLanguagePython:     {"str", "int", "float", "bool", "list", "dict", "tuple", "set", "None"},
		SyntaxLanguageRust:       {"String", "str", "i32", "i64", "u32", "u64", "f32", "f64", "bool", "char", "usize", "isize", "Result", "Option"},
		SyntaxLanguageSQL:        {"varchar", "text", "int", "integer", "bigint", "boolean", "timestamp", "date", "json", "jsonb"},
		SyntaxLanguageShell:      {"string", "int", "bool"},
		SyntaxLanguageJSON:       {"string", "number", "boolean", "object", "array"},
		SyntaxLanguageYAML:       {"string", "int", "bool", "float"},
	}

	if h.language == SyntaxLanguageSQL {
		h.keywords = compileWordsCaseInsensitive(keywordPatterns[h.language])
		h.types = compileWordsCaseInsensitive(typePatterns[h.language])
	} else {
		h.keywords = compileWords(keywordPatterns[h.language])
		h.types = compileWords(typePatterns[h.language])
	}

	return h
}

func normalizeLanguage(lang SyntaxLanguage) SyntaxLanguage {
	switch lang {
	case SyntaxLanguageGo,
		SyntaxLanguageJavaScript,
		SyntaxLanguageTypeScript,
		SyntaxLanguagePython,
		SyntaxLanguageRust,
		SyntaxLanguageSQL,
		SyntaxLanguageShell,
		SyntaxLanguageJSON,
		SyntaxLanguageYAML,
		SyntaxLanguageMarkdown,
		SyntaxLanguagePlain:
		return lang
	default:
		return SyntaxLanguagePlain
	}
}

func parseSyntaxLanguage(lang string) SyntaxLanguage {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "go":
		return SyntaxLanguageGo
	case "javascript", "js":
		return SyntaxLanguageJavaScript
	case "typescript", "ts":
		return SyntaxLanguageTypeScript
	case "python", "py":
		return SyntaxLanguagePython
	case "rust", "rs":
		return SyntaxLanguageRust
	case "sql":
		return SyntaxLanguageSQL
	case "shell", "sh", "bash", "zsh":
		return SyntaxLanguageShell
	case "json":
		return SyntaxLanguageJSON
	case "yaml", "yml":
		return SyntaxLanguageYAML
	case "markdown", "md":
		return SyntaxLanguageMarkdown
	case "plain", "text", "txt":
		return SyntaxLanguagePlain
	default:
		return SyntaxLanguagePlain
	}
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

// DetectLanguage detects language based on a filename extension.
func DetectLanguage(filename string) SyntaxLanguage {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(filename)))
	switch ext {
	case ".go":
		return SyntaxLanguageGo
	case ".js", ".mjs", ".cjs":
		return SyntaxLanguageJavaScript
	case ".ts", ".tsx":
		return SyntaxLanguageTypeScript
	case ".py":
		return SyntaxLanguagePython
	case ".rs":
		return SyntaxLanguageRust
	case ".sql":
		return SyntaxLanguageSQL
	case ".sh", ".bash", ".zsh":
		return SyntaxLanguageShell
	case ".json":
		return SyntaxLanguageJSON
	case ".yaml", ".yml":
		return SyntaxLanguageYAML
	case ".md", ".markdown":
		return SyntaxLanguageMarkdown
	default:
		return SyntaxLanguagePlain
	}
}

// HighlightLine applies syntax highlighting to a single line.
func (h *Highlighter) HighlightLine(line string) string {
	if h == nil || h.language == SyntaxLanguagePlain || line == "" {
		return line
	}

	comment := ""
	if idx := h.comments.FindStringIndex(line); idx != nil {
		comment = line[idx[0]:]
		line = line[:idx[0]]
	}

	stringMatches := h.strings.FindAllStringIndex(line, -1)
	stringLiterals := make([]string, 0, len(stringMatches))
	var builder strings.Builder
	last := 0
	for i, match := range stringMatches {
		if match[0] > last {
			builder.WriteString(line[last:match[0]])
		}
		placeholder := placeholderToken(i)
		builder.WriteString(placeholder)
		stringLiterals = append(stringLiterals, line[match[0]:match[1]])
		last = match[1]
	}
	if last < len(line) {
		builder.WriteString(line[last:])
	}
	working := builder.String()

	if h.numbers != nil {
		working = h.numbers.ReplaceAllStringFunc(working, colorNumber)
	}
	if h.operators != nil {
		working = h.operators.ReplaceAllStringFunc(working, colorOperator)
	}
	if h.types != nil {
		working = h.types.ReplaceAllStringFunc(working, colorType)
	}
	if h.keywords != nil {
		working = h.keywords.ReplaceAllStringFunc(working, colorKeyword)
	}

	for i, literal := range stringLiterals {
		working = strings.ReplaceAll(working, placeholderToken(i), colorString(literal))
	}

	if comment != "" {
		working += colorComment(comment)
	}

	return working
}

func placeholderToken(index int) string {
	return "__SCKSTR" + strconv.Itoa(index) + "__"
}

func colorKeyword(s string) string {
	return style.ANSIBold + style.ANSIBlue + s + style.ANSIReset
}

func colorType(s string) string {
	return style.ANSICyan + s + style.ANSIReset
}

func colorString(s string) string {
	return style.ANSIGreen + s + style.ANSIReset
}

func colorComment(s string) string {
	return style.ANSIDim + s + style.ANSIReset
}

func colorNumber(s string) string {
	return style.ANSIYellow + s + style.ANSIReset
}

func colorOperator(s string) string {
	return style.ANSIDim + style.ANSIRed + s + style.ANSIReset
}

// HighlightLines highlights multiple lines.
func (h *Highlighter) HighlightLines(lines []string) []string {
	if len(lines) == 0 {
		return lines
	}
	out := make([]string, len(lines))
	for i := range lines {
		out[i] = h.HighlightLine(lines[i])
	}
	return out
}
