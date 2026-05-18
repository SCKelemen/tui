// Package markdown renders CommonMark to ANSI for the terminal.
//
// The rendering engine is github.com/charmbracelet/glamour, which transitively
// uses goldmark (CommonMark parser) and chroma (syntax highlighting). Wrapping
// glamour gives us a battle-tested implementation without adding two new
// top-level dependencies.
//
// Glamour ships built-in styles ("dark", "light", "notty", "auto", ...). The
// renderer's default is "auto", which lets glamour pick dark or light based
// on the detected terminal background. Call NewRenderer with WithStyle, or
// WithDesignTokens, to override.
package markdown

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"

	"github.com/SCKelemen/tui/v2/style/design"
)

// CodeHighlighter is a structural interface for syntax-highlighting fenced
// code blocks. It mirrors v2/syntax.Highlighter so callers can wire that
// implementation without introducing an import cycle.
type CodeHighlighter interface {
	Highlight(language, source string) string
}

// Renderer renders CommonMark markdown to ANSI-styled terminal output.
//
// A Renderer is safe to reuse across calls. Configuration is locked in at
// construction time; create a new Renderer to change width, style, etc.
type Renderer struct {
	width       int
	style       string
	highlighter CodeHighlighter
	tokens      *design.DesignTokens
}

// Option configures a Renderer.
type Option func(*Renderer)

// WithWidth sets the word-wrap width in terminal cells. Lines (including any
// margin/padding glamour adds for the active style) are bounded by this
// width. Defaults to 80; non-positive values are ignored.
func WithWidth(w int) Option {
	return func(r *Renderer) {
		if w > 0 {
			r.width = w
		}
	}
}

// WithStyle selects a built-in glamour style by name. Recognized values
// include "dark", "light", "notty", and "auto" (the default). An unknown
// name is forwarded to glamour and surfaces as an error from NewRenderer.
func WithStyle(name string) Option {
	return func(r *Renderer) {
		if name != "" {
			r.style = name
		}
	}
}

// WithCodeHighlighter installs a custom highlighter for fenced code blocks.
// When set, fenced blocks are extracted from the source, rendered by the
// highlighter, and re-inserted around the surrounding markdown which is
// still rendered by glamour. This lets callers wire v2/syntax.Highlighter
// (chroma-backed) without an import cycle from this package.
func WithCodeHighlighter(h CodeHighlighter) Option {
	return func(r *Renderer) {
		r.highlighter = h
	}
}

// WithDesignTokens bridges design-system tokens onto the renderer. Currently
// only the Mode field ("dark"/"light") is honored, selecting the matching
// glamour built-in style. Richer token bridging (custom colors, headings,
// code highlighting) is tracked as a follow-up. Nil tokens are a no-op.
func WithDesignTokens(t *design.DesignTokens) Option {
	return func(r *Renderer) {
		if t == nil {
			return
		}
		r.tokens = t
		switch strings.ToLower(strings.TrimSpace(t.Mode)) {
		case "dark":
			r.style = styles.DarkStyle
		case "light":
			r.style = styles.LightStyle
		}
	}
}

// NewRenderer constructs a Renderer from the supplied options. It validates
// the underlying glamour configuration eagerly: an unknown style name or
// other glamour configuration error is returned here rather than at Render
// time.
func NewRenderer(opts ...Option) (*Renderer, error) {
	r := &Renderer{
		width: 80,
		style: styles.AutoStyle,
	}
	for _, opt := range opts {
		opt(r)
	}
	if _, err := r.buildGlamour(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Renderer) buildGlamour() (*glamour.TermRenderer, error) {
	var styleOpt glamour.TermRendererOption
	switch r.style {
	case "", styles.AutoStyle:
		styleOpt = glamour.WithAutoStyle()
	default:
		styleOpt = glamour.WithStandardStyle(r.style)
	}
	return glamour.NewTermRenderer(styleOpt, glamour.WithWordWrap(r.width))
}

// fencedRE matches a single fenced code block. The language tag is captured
// in group 1 (may be empty) and the body (without the trailing newline
// preceding the closing fence) in group 2.
var fencedRE = regexp.MustCompile("(?ms)^```([^\n]*)\n(.*?)\n```[ \t]*\n?")

// Render renders markdown bytes to a styled terminal string.
func (r *Renderer) Render(source []byte) (string, error) {
	return r.RenderString(string(source))
}

// RenderString renders markdown text to a styled terminal string.
func (r *Renderer) RenderString(source string) (string, error) {
	if r.highlighter == nil {
		gr, err := r.buildGlamour()
		if err != nil {
			return "", err
		}
		return gr.Render(source)
	}

	// With a custom highlighter: split the source on fenced code blocks.
	// Render every text segment via glamour and every code segment via the
	// caller-supplied highlighter, then concatenate. This sidesteps
	// glamour's internal chroma without trying to splice ANSI back into
	// glamour's renderer mid-stream.
	gr, err := r.buildGlamour()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	matches := fencedRE.FindAllStringSubmatchIndex(source, -1)
	cursor := 0
	for _, m := range matches {
		blockStart, blockEnd := m[0], m[1]
		langStart, langEnd := m[2], m[3]
		bodyStart, bodyEnd := m[4], m[5]

		if blockStart > cursor {
			out, rerr := gr.Render(source[cursor:blockStart])
			if rerr != nil {
				return "", rerr
			}
			buf.WriteString(out)
		}

		lang := strings.TrimSpace(source[langStart:langEnd])
		body := source[bodyStart:bodyEnd]
		highlighted := r.highlighter.Highlight(lang, body)
		buf.WriteString(highlighted)
		if !strings.HasSuffix(highlighted, "\n") {
			buf.WriteByte('\n')
		}
		cursor = blockEnd
	}
	if cursor < len(source) {
		out, rerr := gr.Render(source[cursor:])
		if rerr != nil {
			return "", rerr
		}
		buf.WriteString(out)
	}
	return buf.String(), nil
}
