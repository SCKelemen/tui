package tui

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

// GraphemeCluster represents a single user-perceived character.
type GraphemeCluster struct {
	Runes []rune
	Width int
}

// String returns the grapheme cluster as a string.
func (g GraphemeCluster) String() string {
	return string(g.Runes)
}

// GraphemeIterator iterates over grapheme clusters in a string.
type GraphemeIterator struct {
	source string
	iter   *uniseg.Graphemes
	offset int
}

// NewGraphemeIterator creates a grapheme cluster iterator for s.
func NewGraphemeIterator(s string) *GraphemeIterator {
	return &GraphemeIterator{
		source: s,
		iter:   uniseg.NewGraphemes(s),
	}
}

// Next returns the next grapheme cluster.
func (it *GraphemeIterator) Next() (GraphemeCluster, bool) {
	if it == nil || it.iter == nil {
		return GraphemeCluster{}, false
	}
	if !it.iter.Next() {
		it.offset = len(it.source)
		return GraphemeCluster{}, false
	}

	_, end := it.iter.Positions()
	it.offset = end

	return GraphemeCluster{
		Runes: append([]rune(nil), it.iter.Runes()...),
		Width: clampClusterWidth(it.iter.Width()),
	}, true
}

// Remaining returns the unconsumed suffix of the iterator's source string.
func (it *GraphemeIterator) Remaining() string {
	if it == nil {
		return ""
	}
	if it.offset >= len(it.source) {
		return ""
	}
	return it.source[it.offset:]
}

// StringWidth returns the terminal display width of s.
func StringWidth(s string) int {
	width := 0
	it := NewGraphemeIterator(s)
	for {
		cluster, ok := it.Next()
		if !ok {
			break
		}
		width += cluster.Width
	}
	return width
}

// Truncate truncates s at a grapheme boundary so it fits within maxWidth.
func Truncate(s string, maxWidth int) string {
	if maxWidth <= 0 || s == "" {
		return ""
	}
	if StringWidth(s) <= maxWidth {
		return s
	}

	var out strings.Builder
	width := 0
	it := NewGraphemeIterator(s)
	for {
		cluster, ok := it.Next()
		if !ok {
			break
		}
		if cluster.Width > 0 && width+cluster.Width > maxWidth {
			break
		}
		out.WriteString(cluster.String())
		width += cluster.Width
	}

	return out.String()
}

// TruncateWithEllipsis truncates s at a grapheme boundary and appends an ellipsis.
func TruncateWithEllipsis(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if StringWidth(s) <= maxWidth {
		return s
	}

	const ellipsis = "…"
	ellipsisWidth := StringWidth(ellipsis)
	if maxWidth <= ellipsisWidth {
		return Truncate(ellipsis, maxWidth)
	}

	return Truncate(s, maxWidth-ellipsisWidth) + ellipsis
}

// PadRight pads s on the right with spaces until it reaches width.
func PadRight(s string, width int) string {
	current := StringWidth(s)
	if current >= width {
		return s
	}
	return s + strings.Repeat(" ", width-current)
}

// PadLeft pads s on the left with spaces until it reaches width.
func PadLeft(s string, width int) string {
	current := StringWidth(s)
	if current >= width {
		return s
	}
	return strings.Repeat(" ", width-current) + s
}

// PadCenter centers s within width using spaces.
func PadCenter(s string, width int) string {
	current := StringWidth(s)
	if current >= width {
		return s
	}
	pad := width - current
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// WrapText wraps s to width at grapheme boundaries.
func WrapText(s string, width int) []string {
	if width <= 0 {
		return []string{""}
	}
	if s == "" {
		return []string{""}
	}

	paragraphs := strings.Split(s, "\n")
	lines := make([]string, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		lines = append(lines, wrapParagraph(paragraph, width)...)
	}
	return lines
}

// IsWide reports whether r is double-width in terminal cells.
func IsWide(r rune) bool {
	return runewidth.RuneWidth(r) == 2
}

// IsCombining reports whether r is a combining character.
func IsCombining(r rune) bool {
	return unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) || unicode.Is(unicode.Me, r)
}

// IsZeroWidth reports whether r occupies no terminal cells.
func IsZeroWidth(r rune) bool {
	if IsCombining(r) {
		return true
	}
	if unicode.Is(unicode.Cf, r) {
		return true
	}
	if unicode.IsControl(r) {
		return true
	}
	return runewidth.RuneWidth(r) == 0
}

func clampClusterWidth(width int) int {
	if width <= 0 {
		return 0
	}
	if width >= 2 {
		return 2
	}
	return 1
}

func firstRuneInString(s string) rune {
	r, _ := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && s == "" {
		return 0
	}
	return r
}

type wrapToken struct {
	text  string
	width int
	space bool
}

func wrapParagraph(s string, width int) []string {
	if s == "" {
		return []string{""}
	}

	tokens := tokenizeWrapText(s)
	lines := make([]string, 0, 1)

	var line strings.Builder
	lineWidth := 0
	pendingSpaceText := ""
	pendingSpaceWidth := 0

	flush := func() {
		lines = append(lines, line.String())
		line.Reset()
		lineWidth = 0
		pendingSpaceText = ""
		pendingSpaceWidth = 0
	}
	appendText := func(text string, textWidth int) {
		line.WriteString(text)
		lineWidth += textWidth
	}

	for _, token := range tokens {
		if token.space {
			if lineWidth > 0 {
				pendingSpaceText = token.text
				pendingSpaceWidth = token.width
			}
			continue
		}

		if token.width > width {
			if lineWidth > 0 {
				flush()
			}
			parts := splitTokenByWidth(token.text, width)
			for i, part := range parts {
				appendText(part, StringWidth(part))
				if i < len(parts)-1 {
					flush()
				}
			}
			pendingSpaceText = ""
			pendingSpaceWidth = 0
			continue
		}

		if lineWidth == 0 {
			appendText(token.text, token.width)
			pendingSpaceText = ""
			pendingSpaceWidth = 0
			continue
		}

		if pendingSpaceWidth > 0 && lineWidth+pendingSpaceWidth+token.width <= width {
			appendText(pendingSpaceText, pendingSpaceWidth)
			appendText(token.text, token.width)
			pendingSpaceText = ""
			pendingSpaceWidth = 0
			continue
		}

		if lineWidth+token.width <= width {
			appendText(token.text, token.width)
			pendingSpaceText = ""
			pendingSpaceWidth = 0
			continue
		}

		flush()
		appendText(token.text, token.width)
	}

	if lineWidth > 0 || len(lines) == 0 {
		lines = append(lines, line.String())
	}

	return lines
}

func tokenizeWrapText(s string) []wrapToken {
	tokens := make([]wrapToken, 0, 8)
	it := NewGraphemeIterator(s)

	var current strings.Builder
	currentWidth := 0
	currentSpace := false
	hasCurrent := false

	flush := func() {
		if !hasCurrent {
			return
		}
		tokens = append(tokens, wrapToken{
			text:  current.String(),
			width: currentWidth,
			space: currentSpace,
		})
		current.Reset()
		currentWidth = 0
		hasCurrent = false
	}

	for {
		cluster, ok := it.Next()
		if !ok {
			break
		}
		space := isWhitespaceCluster(cluster)
		if !hasCurrent {
			currentSpace = space
			hasCurrent = true
		} else if currentSpace != space {
			flush()
			currentSpace = space
			hasCurrent = true
		}

		current.WriteString(cluster.String())
		currentWidth += cluster.Width
	}
	flush()

	return tokens
}

func splitTokenByWidth(s string, width int) []string {
	if s == "" {
		return nil
	}
	if width <= 0 {
		return []string{s}
	}

	parts := make([]string, 0, 1)
	var current strings.Builder
	currentWidth := 0

	flush := func() {
		if current.Len() == 0 {
			return
		}
		parts = append(parts, current.String())
		current.Reset()
		currentWidth = 0
	}

	it := NewGraphemeIterator(s)
	for {
		cluster, ok := it.Next()
		if !ok {
			break
		}
		clusterText := cluster.String()
		clusterWidth := cluster.Width

		if current.Len() == 0 {
			current.WriteString(clusterText)
			currentWidth = clusterWidth
			if clusterWidth >= width {
				flush()
			}
			continue
		}

		if clusterWidth > 0 && currentWidth+clusterWidth > width {
			flush()
			current.WriteString(clusterText)
			currentWidth = clusterWidth
			if clusterWidth >= width {
				flush()
			}
			continue
		}

		current.WriteString(clusterText)
		currentWidth += clusterWidth
	}
	flush()

	if len(parts) == 0 {
		return []string{""}
	}
	return parts
}

func isWhitespaceCluster(cluster GraphemeCluster) bool {
	if len(cluster.Runes) == 0 {
		return false
	}
	for _, r := range cluster.Runes {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}
