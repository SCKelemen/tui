package display

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HoverSection is a single section rendered in a hover card.
type HoverSection struct {
	Content  string
	Language string
	IsCode   bool
}

// HoverCard renders a VS Code-style hover tooltip.
type HoverCard struct {
	sections    []HoverSection
	width       int
	maxWidth    int
	maxHeight   int
	borderColor string
}

// HoverCardOption configures a HoverCard.
type HoverCardOption func(*HoverCard)

// WithHoverCardWidth sets a fixed inner width for card content.
func WithHoverCardWidth(width int) HoverCardOption {
	return func(h *HoverCard) {
		if width > 0 {
			h.width = width
		}
	}
}

// WithHoverCardMaxWidth sets the maximum inner content width.
func WithHoverCardMaxWidth(maxWidth int) HoverCardOption {
	return func(h *HoverCard) {
		if maxWidth > 0 {
			h.maxWidth = maxWidth
		}
	}
}

// WithHoverCardMaxHeight sets the maximum rendered card height in lines.
func WithHoverCardMaxHeight(maxHeight int) HoverCardOption {
	return func(h *HoverCard) {
		if maxHeight > 0 {
			h.maxHeight = maxHeight
		}
	}
}

// WithHoverCardBorderColor sets the card border color.
func WithHoverCardBorderColor(borderColor string) HoverCardOption {
	return func(h *HoverCard) {
		if strings.TrimSpace(borderColor) != "" {
			h.borderColor = borderColor
		}
	}
}

// NewHoverCard creates a hover tooltip component.
func NewHoverCard(sections []HoverSection, opts ...HoverCardOption) *HoverCard {
	h := &HoverCard{
		sections:    sections,
		maxWidth:    60,
		maxHeight:   20,
		borderColor: "#3C414B",
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// View renders the hover card.
func (h *HoverCard) View() string {
	innerWidth := h.innerWidth()
	if innerWidth <= 0 {
		innerWidth = 60
	}

	baseContentStyle := lipgloss.NewStyle().
		Width(innerWidth).
		Background(lipgloss.Color("#2C313A"))

	contentLines := h.renderContentLines(innerWidth)
	if len(contentLines) == 0 {
		contentLines = []string{""}
	}

	styledLines := make([]string, len(contentLines))
	for i := range contentLines {
		styledLines[i] = baseContentStyle.Render(contentLines[i])
	}

	inner := strings.Join(styledLines, "\n")

	border := lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
	}

	card := lipgloss.NewStyle().
		Border(border, true).
		BorderForeground(lipgloss.Color(h.borderColor)).
		Padding(0, 1).
		Render(inner)

	return h.clipCardHeight(card, innerWidth)
}

// NewTypeHoverCard creates a hover card for type information.
func NewTypeHoverCard(typeName, typeKind, signature, doc string) *HoverCard {
	sections := make([]HoverSection, 0, 3)

	header := strings.TrimSpace(strings.Join([]string{typeName, typeKind}, " "))
	if header != "" {
		sections = append(sections, HoverSection{Content: fmt.Sprintf("**%s**", header)})
	}
	if strings.TrimSpace(signature) != "" {
		sections = append(sections, HoverSection{Content: signature, Language: "go", IsCode: true})
	}
	if strings.TrimSpace(doc) != "" {
		sections = append(sections, HoverSection{Content: doc})
	}

	return NewHoverCard(sections)
}

// NewFunctionHoverCard creates a hover card for function information.
func NewFunctionHoverCard(funcName, signature, doc string, params []string) *HoverCard {
	sections := make([]HoverSection, 0, 4)

	if strings.TrimSpace(funcName) != "" {
		sections = append(sections, HoverSection{Content: fmt.Sprintf("**%s**", strings.TrimSpace(funcName))})
	}
	if strings.TrimSpace(signature) != "" {
		sections = append(sections, HoverSection{Content: signature, Language: "go", IsCode: true})
	}
	if len(params) > 0 {
		var b strings.Builder
		b.WriteString("### Parameters\n")
		for _, param := range params {
			if strings.TrimSpace(param) == "" {
				continue
			}
			b.WriteString("- ")
			b.WriteString(param)
			b.WriteString("\n")
		}
		sections = append(sections, HoverSection{Content: strings.TrimSpace(b.String())})
	}
	if strings.TrimSpace(doc) != "" {
		sections = append(sections, HoverSection{Content: doc})
	}

	return NewHoverCard(sections)
}

func (h *HoverCard) innerWidth() int {
	if h.width > 0 {
		return h.width
	}
	if h.maxWidth > 0 {
		return h.maxWidth
	}
	return 60
}

func (h *HoverCard) renderContentLines(innerWidth int) []string {
	result := make([]string, 0)
	for i, section := range h.sections {
		sectionLines := h.renderSection(section, innerWidth)
		result = append(result, sectionLines...)
		if i < len(h.sections)-1 {
			result = append(result, h.renderDivider(innerWidth))
		}
	}
	return result
}

func (h *HoverCard) renderSection(section HoverSection, innerWidth int) []string {
	content := strings.TrimSpace(section.Content)
	if content == "" {
		return []string{""}
	}

	if section.IsCode {
		return h.renderCodeSection(content, section.Language, innerWidth)
	}

	rendered := NewMarkdown(content, WithMarkdownWidth(innerWidth)).View()
	if strings.TrimSpace(rendered) == "" {
		return []string{""}
	}

	lines := strings.Split(rendered, "\n")
	for i := range lines {
		lines[i] = truncateANSI(lines[i], innerWidth)
	}

	return lines
}

func (h *HoverCard) renderCodeSection(content, language string, innerWidth int) []string {
	highlighted := HighlightCode(content, language)
	lines := strings.Split(highlighted, "\n")

	codeLineStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#31353D"))

	result := make([]string, 0, len(lines))
	for _, line := range lines {
		truncated := truncateANSI(line, innerWidth)
		result = append(result, codeLineStyle.Render(truncated))
	}

	if len(result) == 0 {
		return []string{""}
	}

	return result
}

func (h *HoverCard) renderDivider(innerWidth int) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.borderColor)).
		Render(strings.Repeat("─", innerWidth))
}

func (h *HoverCard) clipCardHeight(card string, innerWidth int) string {
	if h.maxHeight <= 0 {
		return card
	}

	lines := strings.Split(card, "\n")
	if len(lines) <= h.maxHeight {
		return card
	}
	if h.maxHeight < 3 {
		return strings.Join(lines[:h.maxHeight], "\n")
	}

	clipped := make([]string, 0, h.maxHeight)
	clipped = append(clipped, lines[:h.maxHeight]...)

	indicatorWidth := lipgloss.Width(clipped[h.maxHeight-2])
	if indicatorWidth <= 2 {
		indicatorWidth = innerWidth + 2
	}
	indicatorContent := truncateANSI("...", indicatorWidth-2)
	clipped[h.maxHeight-2] = "│" + lipgloss.NewStyle().
		Width(indicatorWidth-2).
		Align(lipgloss.Left).
		Background(lipgloss.Color("#2C313A")).
		Render(indicatorContent) + "│"

	return strings.Join(clipped, "\n")
}
