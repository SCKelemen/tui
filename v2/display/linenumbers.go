package display

import (
	"strconv"
	"strings"

	"github.com/SCKelemen/tui/v2/style"
	"github.com/SCKelemen/tui/v2/style/design"
)

// LineNumbers is a stateless gutter renderer that displays line numbers
// to the left of a viewport. It is intentionally simpler than GutterRenderer
// in gutter.go: it only renders numbers (absolute or relative), optional
// per-line signs, and current-line emphasis. It does not own decorations
// or interpret diagnostics.
//
// LineNumbers is not a tui.Component — it is a pure renderer that
// callers compose into their own layouts. The companion methods
// Render and RenderLine produce strings the caller is responsible
// for joining with the content viewport.
//
// Width contract: the gutter occupies a fixed number of columns equal
// to Width(). Numbers are right-aligned within the numeric column and
// followed by a single space separator. When any sign is configured,
// a single sign cell is inserted between the number and the separator.
type LineNumbers struct {
	start    int            // first absolute line number (1-based, inclusive)
	count    int            // number of lines to render
	minWidth int            // caller-supplied minimum width for the numeric column
	relative bool           // when true, non-current lines show distance from current
	current  int            // absolute line number considered "current"
	signs    map[int]string // optional per-line single-cell signs keyed by absolute line
	tokens   *design.DesignTokens

	// computed lazily; cleared whenever start/count/minWidth change.
	numberWidth int
	hasSigns    bool
}

// LineNumbersOption configures a LineNumbers renderer.
type LineNumbersOption func(*LineNumbers)

// WithLineNumbersStart sets the first absolute line number rendered.
// Defaults to 1. Values below 1 are coerced to 1.
func WithLineNumbersStart(n int) LineNumbersOption {
	return func(l *LineNumbers) {
		if n < 1 {
			n = 1
		}
		l.start = n
	}
}

// WithLineNumbersCount sets how many consecutive lines are rendered.
// Defaults to 1. Values below 1 are coerced to 1.
func WithLineNumbersCount(n int) LineNumbersOption {
	return func(l *LineNumbers) {
		if n < 1 {
			n = 1
		}
		l.count = n
	}
}

// WithLineNumbersWidth sets the minimum width of the numeric column.
// The renderer expands the column automatically when the largest line
// number requires more digits.
func WithLineNumbersWidth(w int) LineNumbersOption {
	return func(l *LineNumbers) {
		if w > 0 {
			l.minWidth = w
		}
	}
}

// WithLineNumbersRelative toggles relative numbering. In relative mode
// every non-current line shows its absolute distance from the current
// line while the current line keeps its absolute number.
func WithLineNumbersRelative(b bool) LineNumbersOption {
	return func(l *LineNumbers) { l.relative = b }
}

// WithLineNumbersCurrentLine sets the absolute line number considered
// the current line. The current line is rendered in bold; other lines
// are rendered dim.
func WithLineNumbersCurrentLine(n int) LineNumbersOption {
	return func(l *LineNumbers) { l.current = n }
}

// WithLineNumbersDesignTokens applies a design-system palette. When set
// and a suitable text color is present, the foreground of every gutter
// line is tinted with that color.
func WithLineNumbersDesignTokens(tokens *design.DesignTokens) LineNumbersOption {
	return func(l *LineNumbers) { l.tokens = tokens }
}

// WithLineNumbersSigns installs a map of per-line signs. The map is
// keyed by absolute line number; each value contributes a single cell
// drawn between the numeric column and the trailing separator.
// Multi-cell strings are truncated to their first rune.
func WithLineNumbersSigns(signs map[int]string) LineNumbersOption {
	return func(l *LineNumbers) {
		if signs == nil {
			l.signs = nil
			return
		}
		copyMap := make(map[int]string, len(signs))
		for k, v := range signs {
			copyMap[k] = v
		}
		l.signs = copyMap
	}
}

// NewLineNumbers constructs a LineNumbers renderer with the supplied
// options applied in order. Defaults: start=1, count=1, no minimum
// width, absolute numbering, no current line, no signs.
func NewLineNumbers(opts ...LineNumbersOption) *LineNumbers {
	l := &LineNumbers{
		start:    1,
		count:    1,
		minWidth: 0,
		current:  0,
	}
	for _, opt := range opts {
		opt(l)
	}
	l.recomputeLayout()
	return l
}

// recomputeLayout refreshes derived layout fields (numberWidth, hasSigns).
func (l *LineNumbers) recomputeLayout() {
	last := l.start + l.count - 1
	if last < l.start {
		last = l.start
	}
	digits := len(strconv.Itoa(last))
	if digits < 1 {
		digits = 1
	}
	if l.minWidth > digits {
		l.numberWidth = l.minWidth
	} else {
		l.numberWidth = digits
	}
	l.hasSigns = len(l.signs) > 0
}

// SetCurrentLine updates the current-line marker after construction.
// This is the only mutating method on LineNumbers; all other state
// is configured at construction time.
func (l *LineNumbers) SetCurrentLine(n int) { l.current = n }

// Width returns the total column width consumed by the rendered gutter,
// including the numeric column, the optional sign cell, and the trailing
// separator space.
func (l *LineNumbers) Width() int {
	w := l.numberWidth + 1 // numeric column + separator space
	if l.hasSigns {
		w++ // sign cell
	}
	return w
}

// Render returns the full gutter as one string with newline-separated
// rows, suitable for placing alongside a vertically-aligned viewport.
func (l *LineNumbers) Render() string {
	if l.count <= 0 {
		return ""
	}
	rows := make([]string, 0, l.count)
	for i := 0; i < l.count; i++ {
		rows = append(rows, l.RenderLine(l.start+i))
	}
	return strings.Join(rows, "\n")
}

// RenderLine renders a single gutter row for the supplied absolute
// line number. Callers may use RenderLine to drive partial repaints
// without re-emitting the entire gutter.
func (l *LineNumbers) RenderLine(absoluteLine int) string {
	displayNum := absoluteLine
	if l.relative && l.current != 0 && absoluteLine != l.current {
		displayNum = absInt(absoluteLine - l.current)
	}

	numStr := strconv.Itoa(displayNum)
	padded := rightAlignInWidth(numStr, l.numberWidth)

	var b strings.Builder
	b.Grow(l.Width() + 16) // leave room for ANSI escape envelope

	// Style envelope: bold for the current line, dim for all others.
	isCurrent := l.current != 0 && absoluteLine == l.current
	if isCurrent {
		b.WriteString(style.ANSIBold)
	} else {
		b.WriteString(style.ANSIDim)
	}

	// Optional design-token foreground tint.
	if l.tokens != nil {
		if c := lineNumbersTokenColor(l.tokens); c != "" {
			b.WriteString(style.ANSIColorFromHex(c))
		}
	}

	b.WriteString(padded)

	if l.hasSigns {
		sign := " "
		if l.signs != nil {
			if s, ok := l.signs[absoluteLine]; ok && s != "" {
				sign = firstCell(s)
			}
		}
		b.WriteString(sign)
	}

	b.WriteString(" ") // trailing separator
	b.WriteString(style.ANSIReset)
	return b.String()
}

// rightAlignInWidth pads s with leading spaces so the result is exactly
// width columns wide. When s is already at least width wide the input
// is returned unchanged.
//
// This is intentionally separate from the leftPad helper in
// calendar_month.go: that helper pads on a different side and lives
// in the calendar render path.
func rightAlignInWidth(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

// absInt returns the absolute value of n.
func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// firstCell returns the first rune of s as a string, preserving sign
// glyphs that are exactly one cell wide. Multi-rune signs are truncated.
func firstCell(s string) string {
	for _, r := range s {
		return string(r)
	}
	return " "
}

// lineNumbersTokenColor resolves a foreground hex color from the
// supplied design-system tokens. It prefers muted/dim text colors so
// that the gutter does not compete with the primary content.
func lineNumbersTokenColor(t *design.DesignTokens) string {
	return tokenColor(t, "TextDim", "TextMuted", "MutedColor", "TextSecondary", "Color")
}
