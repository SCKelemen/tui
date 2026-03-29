package display

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

// Markdown renders common Markdown elements for terminal output.
type Markdown struct {
	content string
	width   int
	theme   MarkdownTheme
}

// MarkdownTheme controls colors and characters used by Markdown rendering.
type MarkdownTheme struct {
	HeadingColor    string
	BoldColor       string
	CodeColor       string
	LinkColor       string
	BlockquoteColor string
	CodeBgColor     string
	HRChar          string
}

// MarkdownOption configures a Markdown renderer.
type MarkdownOption func(*Markdown)

// WithMarkdownWidth sets the target render width.
func WithMarkdownWidth(w int) MarkdownOption {
	return func(m *Markdown) {
		if w >= 0 {
			m.width = w
		}
	}
}

// WithMarkdownTheme applies a custom theme.
func WithMarkdownTheme(theme MarkdownTheme) MarkdownOption {
	return func(m *Markdown) {
		m.theme = theme
	}
}

// DefaultMarkdownTheme returns the default Markdown renderer theme.
func DefaultMarkdownTheme() MarkdownTheme {
	return MarkdownTheme{
		HeadingColor:    "#61AFEF",
		BoldColor:       "#E5C07B",
		CodeColor:       "#E06C75",
		LinkColor:       "#56B6C2",
		BlockquoteColor: "#7F848E",
		CodeBgColor:     "#2C313A",
		HRChar:          "─",
	}
}

// NewMarkdown creates a Markdown renderer with optional configuration.
func NewMarkdown(content string, opts ...MarkdownOption) *Markdown {
	m := &Markdown{
		content: content,
		theme:   DefaultMarkdownTheme(),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// View renders Markdown content with terminal styling.
func (m *Markdown) View() string {
	if m.content == "" {
		return ""
	}

	lines := strings.Split(m.content, "\n")
	output := make([]string, 0, len(lines))

	inCodeBlock := false
	codeLang := ""
	codeLines := make([]string, 0)

	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				output = append(output, m.renderCodeBlock(codeLines, codeLang)...)
				inCodeBlock = false
				codeLang = ""
				codeLines = codeLines[:0]
			} else {
				inCodeBlock = true
				codeLang = strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
			}
			continue
		}

		if inCodeBlock {
			codeLines = append(codeLines, line)
			continue
		}

		switch {
		case isHorizontalRule(trimmed):
			output = append(output, m.renderHorizontalRule())
		case strings.HasPrefix(trimmed, "### "):
			output = append(output, m.renderHeading(strings.TrimSpace(trimmed[4:]), 3))
		case strings.HasPrefix(trimmed, "## "):
			output = append(output, m.renderHeading(strings.TrimSpace(trimmed[3:]), 2))
		case strings.HasPrefix(trimmed, "# "):
			output = append(output, m.renderHeading(strings.TrimSpace(trimmed[2:]), 1))
		case strings.HasPrefix(trimmed, ">"):
			output = append(output, m.renderBlockquote(strings.TrimSpace(strings.TrimPrefix(trimmed, ">")))...)
		case isUnorderedListItem(trimmed):
			item := strings.TrimSpace(trimmed[1:])
			output = append(output, m.renderListItem(item, "•")...)
		case isOrderedListItem(trimmed):
			number, item := splitOrderedListItem(trimmed)
			output = append(output, m.renderListItem(item, number+".")...)
		case trimmed == "":
			output = append(output, "")
		default:
			output = append(output, m.renderParagraph(line)...)
		}
	}

	if inCodeBlock {
		output = append(output, m.renderCodeBlock(codeLines, codeLang)...)
	}

	return strings.Join(output, "\n")
}

// RenderMarkdown renders Markdown content with a one-off width override.
func RenderMarkdown(content string, width int) string {
	return NewMarkdown(content, WithMarkdownWidth(width)).View()
}

func (m *Markdown) renderHeading(text string, level int) string {
	base := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.HeadingColor)).
		Bold(true)

	var rendered string
	switch level {
	case 1:
		rendered = base.Underline(true).Render(strings.ToUpper(m.renderInline(text)))
	case 2:
		rendered = base.Render(m.renderInline(text))
	default:
		rendered = base.Faint(true).Render(m.renderInline(text))
	}

	return rendered
}

func (m *Markdown) renderParagraph(text string) []string {
	wrapped := m.wrapText(strings.TrimSpace(text), 0)
	if len(wrapped) == 0 {
		return []string{""}
	}

	result := make([]string, 0, len(wrapped))
	for _, line := range wrapped {
		result = append(result, m.renderInline(line))
	}
	return result
}

func (m *Markdown) renderBlockquote(text string) []string {
	prefixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.BlockquoteColor))
	quoteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.BlockquoteColor)).Faint(true)

	wrapped := m.wrapText(text, 2)
	if len(wrapped) == 0 {
		wrapped = []string{""}
	}

	result := make([]string, 0, len(wrapped))
	for _, line := range wrapped {
		rendered := quoteStyle.Render(m.renderInline(line))
		result = append(result, fmt.Sprintf("%s %s", prefixStyle.Render("│"), rendered))
	}
	return result
}

func (m *Markdown) renderListItem(text, marker string) []string {
	markerStyle := lipgloss.NewStyle().Bold(true)
	wrapped := m.wrapText(text, 4)
	if len(wrapped) == 0 {
		wrapped = []string{""}
	}

	result := make([]string, 0, len(wrapped))
	for i, line := range wrapped {
		if i == 0 {
			result = append(result, fmt.Sprintf("%s %s", markerStyle.Render(marker), m.renderInline(line)))
			continue
		}
		result = append(result, "  "+m.renderInline(line))
	}
	return result
}

func (m *Markdown) renderCodeBlock(lines []string, lang string) []string {
	if len(lines) == 0 {
		return nil
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.CodeColor)).
		Background(lipgloss.Color(m.theme.CodeBgColor)).
		PaddingLeft(1).
		PaddingRight(1)

	result := make([]string, 0, len(lines)+1)
	if lang != "" {
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.BlockquoteColor)).Faint(true)
		result = append(result, labelStyle.Render("  "+lang))
	}

	for _, line := range lines {
		codeLine := line
		if m.width > 0 {
			max := m.width - 4
			if max > 0 && lipgloss.Width(codeLine) > max {
				codeLine = truncatePlain(codeLine, max)
			}
		}
		result = append(result, "  "+style.Render(codeLine))
	}

	return result
}

func (m *Markdown) renderHorizontalRule() string {
	width := m.width
	if width <= 0 {
		width = 40
	}
	if width < 3 {
		width = 3
	}

	hrChar := m.theme.HRChar
	if hrChar == "" {
		hrChar = "─"
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.BlockquoteColor)).
		Render(strings.Repeat(hrChar, width))
}

func (m *Markdown) renderInline(text string) string {
	if text == "" {
		return ""
	}

	var b strings.Builder
	for i := 0; i < len(text); {
		switch {
		case strings.HasPrefix(text[i:], "`"):
			end := strings.Index(text[i+1:], "`")
			if end >= 0 {
				segment := text[i+1 : i+1+end]
				styled := lipgloss.NewStyle().
					Foreground(lipgloss.Color(m.theme.CodeColor)).
					Background(lipgloss.Color(m.theme.CodeBgColor)).
					PaddingLeft(1).
					PaddingRight(1).
					Render(segment)
				b.WriteString(styled)
				i += end + 2
				continue
			}
		case strings.HasPrefix(text[i:], "**"):
			end := strings.Index(text[i+2:], "**")
			if end >= 0 {
				segment := text[i+2 : i+2+end]
				styled := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color(m.theme.BoldColor)).
					Render(segment)
				b.WriteString(styled)
				i += end + 4
				continue
			}
		case strings.HasPrefix(text[i:], "*"):
			end := strings.Index(text[i+1:], "*")
			if end >= 0 {
				segment := text[i+1 : i+1+end]
				styled := lipgloss.NewStyle().Faint(true).Italic(true).Render(segment)
				b.WriteString(styled)
				i += end + 2
				continue
			}
		case strings.HasPrefix(text[i:], "["):
			if rendered, consumed, ok := m.tryRenderLink(text[i:]); ok {
				b.WriteString(rendered)
				i += consumed
				continue
			}
		}

		b.WriteByte(text[i])
		i++
	}

	return b.String()
}

func (m *Markdown) tryRenderLink(s string) (string, int, bool) {
	closeBracket := strings.IndexByte(s, ']')
	if closeBracket <= 0 || closeBracket+1 >= len(s) || s[closeBracket+1] != '(' {
		return "", 0, false
	}

	closeParen := strings.IndexByte(s[closeBracket+2:], ')')
	if closeParen < 0 {
		return "", 0, false
	}

	text := s[1:closeBracket]
	url := s[closeBracket+2 : closeBracket+2+closeParen]
	consumed := closeBracket + 3 + closeParen
	if text == "" || url == "" {
		return "", 0, false
	}

	return NewLink(url, text, WithLinkColor(m.theme.LinkColor)).View(), consumed, true
}

func (m *Markdown) wrapText(text string, indent int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if m.width <= 0 {
		return []string{text}
	}

	limit := m.width - indent
	if limit < 10 {
		limit = 10
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	lines := make([]string, 0, 1)
	current := words[0]

	for _, word := range words[1:] {
		candidate := current + " " + word
		if lipgloss.Width(candidate) <= limit {
			current = candidate
			continue
		}
		lines = append(lines, current)
		current = word
	}
	lines = append(lines, current)
	return lines
}

func isHorizontalRule(s string) bool {
	if len(s) < 3 {
		return false
	}
	for _, r := range s {
		if r != '-' {
			return false
		}
	}
	return true
}

func isUnorderedListItem(s string) bool {
	if len(s) < 2 {
		return false
	}
	if (s[0] != '-' && s[0] != '*') || !unicode.IsSpace(rune(s[1])) {
		return false
	}
	return true
}

var orderedListRe = regexp.MustCompile(`^(\d+)\.\s+(.+)$`)

func isOrderedListItem(s string) bool {
	return orderedListRe.MatchString(s)
}

func splitOrderedListItem(s string) (string, string) {
	parts := orderedListRe.FindStringSubmatch(s)
	if len(parts) != 3 {
		return "1", s
	}

	if _, err := strconv.Atoi(parts[1]); err != nil {
		return "1", strings.TrimSpace(parts[2])
	}

	return parts[1], strings.TrimSpace(parts[2])
}

func truncatePlain(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}

	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)+"…") > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}
