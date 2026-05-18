// Package syntax provides terminal syntax highlighting backed by
// github.com/alecthomas/chroma/v2.
//
// Chroma is the de facto standard Go syntax-highlighting library; it is the
// engine behind glamour, bat (via a Go shim), gotty, and many other Go-based
// terminal tools. This package wraps Chroma with a small, opinionated
// interface tailored for the v2 TUI: a single Highlighter that takes a
// language identifier plus source text and returns a string of ANSI-escaped,
// terminal-ready output.
//
// All entry points are intentionally infallible. Lexer errors, unknown
// languages, or formatter failures cause the original source text to be
// returned unchanged. Nothing in this package panics or returns an error.
package syntax

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// Highlighter renders source code into terminal-ready text.
//
// Implementations must be safe for concurrent use by multiple goroutines and
// must never panic; tokenization, formatting, or lookup failures result in
// the source string being returned verbatim.
type Highlighter interface {
	// Highlight returns source rendered for the terminal using the lexer
	// associated with language (a chroma lexer name or alias, e.g. "go",
	// "python", "json"). If the language is unknown or rendering fails,
	// the original source is returned unchanged.
	Highlight(language, source string) string

	// SupportsLanguage reports whether the implementation recognizes the
	// given language identifier as a real (non-fallback) lexer.
	SupportsLanguage(language string) bool
}

// chromaHighlighter is the default chroma-backed Highlighter.
type chromaHighlighter struct {
	style     *chroma.Style
	formatter chroma.Formatter
}

// New returns a Highlighter that emits ANSI escape sequences using chroma's
// terminal256 formatter and the named style. Unknown style names fall back
// to styles.Fallback rather than producing an error.
func New(styleName string) Highlighter {
	return NewWithFormatter(styleName, "terminal256")
}

// NewWithFormatter returns a Highlighter that uses the named chroma style
// and formatter. Both names are looked up against chroma's global registry;
// unknown style names fall back to styles.Fallback and unknown formatter
// names fall back to formatters.Fallback.
func NewWithFormatter(styleName, formatterName string) Highlighter {
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}
	f := formatters.Get(formatterName)
	if f == nil {
		f = formatters.Fallback
	}
	return &chromaHighlighter{style: style, formatter: f}
}

// Highlight implements Highlighter.
func (h *chromaHighlighter) Highlight(language, source string) string {
	lexer := lexers.Get(language)
	if lexer == nil || lexer == lexers.Fallback {
		return source
	}
	lexer = chroma.Coalesce(lexer)
	iter, err := lexer.Tokenise(nil, source)
	if err != nil {
		return source
	}
	var sb strings.Builder
	if err := h.formatter.Format(&sb, h.style, iter); err != nil {
		return source
	}
	return sb.String()
}

// SupportsLanguage implements Highlighter.
func (h *chromaHighlighter) SupportsLanguage(language string) bool {
	l := lexers.Get(language)
	return l != nil && l != lexers.Fallback
}

// plainHighlighter is a no-op pass-through implementation.
type plainHighlighter struct{}

// PlainHighlighter returns a Highlighter that performs no highlighting and
// reports no language support. Useful for headless tests, dumb terminals,
// or when ANSI output must be suppressed.
func PlainHighlighter() Highlighter { return plainHighlighter{} }

// Highlight implements Highlighter.
func (plainHighlighter) Highlight(_, source string) string { return source }

// SupportsLanguage implements Highlighter.
func (plainHighlighter) SupportsLanguage(_ string) bool { return false }

// DetectLanguage returns the chroma lexer name that best matches filename
// (typically by extension or shebang). It returns the empty string when no
// lexer matches.
func DetectLanguage(filename string) string {
	l := lexers.Match(filename)
	if l == nil {
		return ""
	}
	return l.Config().Name
}

// ListStyles returns the names of all chroma styles registered in the
// process, sorted alphabetically.
func ListStyles() []string { return styles.Names() }

// ListLanguages returns the names of all chroma lexers registered in the
// process, sorted alphabetically.
func ListLanguages() []string { return lexers.Names(false) }
