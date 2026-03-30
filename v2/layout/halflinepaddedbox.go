package layout

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// HalfLinePaddedBox is a simple layout component that renders content with
// configurable half-line (single row) padding above and below.
type HalfLinePaddedBox struct {
	content            string
	paddingTop         int
	paddingBottom      int
	background         string
	width              int
	windowWidth        int
	focused            bool
	designTokens       *design.DesignTokens
	backgroundExplicit bool
}

// HalfLinePaddedBoxOption configures a HalfLinePaddedBox.
type HalfLinePaddedBoxOption func(*HalfLinePaddedBox)

// WithHalfLineBoxPaddingTop sets the number of half-lines rendered above content.
func WithHalfLineBoxPaddingTop(n int) HalfLinePaddedBoxOption {
	return func(h *HalfLinePaddedBox) {
		if n >= 0 {
			h.paddingTop = n
		}
	}
}

// WithHalfLineBoxPaddingBottom sets the number of half-lines rendered below content.
func WithHalfLineBoxPaddingBottom(n int) HalfLinePaddedBoxOption {
	return func(h *HalfLinePaddedBox) {
		if n >= 0 {
			h.paddingBottom = n
		}
	}
}

// WithHalfLineBoxBackground sets the box background color using a hex color value.
func WithHalfLineBoxBackground(hex string) HalfLinePaddedBoxOption {
	return func(h *HalfLinePaddedBox) {
		h.background = strings.TrimSpace(hex)
		h.backgroundExplicit = true
	}
}

// WithHalfLineBoxWidth forces the rendered width for the box.
// If width is 0, the component auto-sizes to its content.
func WithHalfLineBoxWidth(width int) HalfLinePaddedBoxOption {
	return func(h *HalfLinePaddedBox) {
		if width >= 0 {
			h.width = width
		}
	}
}

// WithHalfLineBoxDesignTokens applies design tokens to the box.
func WithHalfLineBoxDesignTokens(tokens *design.DesignTokens) HalfLinePaddedBoxOption {
	return func(h *HalfLinePaddedBox) {
		h.applyDesignTokens(tokens)
	}
}

// NewHalfLinePaddedBox creates a new HalfLinePaddedBox component.
func NewHalfLinePaddedBox(content string, opts ...HalfLinePaddedBoxOption) *HalfLinePaddedBox {
	h := &HalfLinePaddedBox{
		content:       content,
		paddingTop:    1,
		paddingBottom: 1,
		designTokens:  design.DefaultTheme(),
	}

	h.applyDesignTokens(h.designTokens)

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// HalfLinePad adds one empty line above and below content.
func HalfLinePad(content string) string {
	return "\n" + content + "\n"
}

// Init initializes the component.
func (h *HalfLinePaddedBox) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (h *HalfLinePaddedBox) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.windowWidth = msg.Width
	}

	return h, nil
}

// View renders the component.
func (h *HalfLinePaddedBox) View() string {
	lines := strings.Split(h.content, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}

	boxWidth := h.effectiveWidth(lines)
	if boxWidth < 1 {
		boxWidth = 1
	}

	bg := style.ANSIBackgroundColorFromHex(h.background)
	var b strings.Builder

	renderEmpty := func() {
		line := strings.Repeat(" ", boxWidth)
		if bg != "" {
			line = bg + line + style.ANSIReset
		}
		b.WriteString(line)
	}

	renderContent := func(line string) {
		content := line
		if style.StringWidth(content) > boxWidth {
			content = style.Truncate(content, boxWidth, "…")
		}

		padding := boxWidth - style.StringWidth(content)
		if padding < 0 {
			padding = 0
		}

		rendered := content + strings.Repeat(" ", padding)
		if bg != "" {
			rendered = bg + rendered + style.ANSIReset
		}
		b.WriteString(rendered)
	}

	totalLines := h.paddingTop + len(lines) + h.paddingBottom
	lineIndex := 0

	for i := 0; i < h.paddingTop; i++ {
		renderEmpty()
		lineIndex++
		if lineIndex < totalLines {
			b.WriteString("\n")
		}
	}

	for i, line := range lines {
		renderContent(line)
		lineIndex++
		if i < len(lines)-1 || lineIndex < totalLines {
			b.WriteString("\n")
		}
	}

	for i := 0; i < h.paddingBottom; i++ {
		renderEmpty()
		lineIndex++
		if lineIndex < totalLines {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// Focus marks the component as focused.
func (h *HalfLinePaddedBox) Focus() {
	h.focused = true
}

// Blur marks the component as unfocused.
func (h *HalfLinePaddedBox) Blur() {
	h.focused = false
}

// Focused returns whether this component is focused.
func (h *HalfLinePaddedBox) Focused() bool {
	return h.focused
}

func (h *HalfLinePaddedBox) effectiveWidth(lines []string) int {
	if h.width > 0 {
		if h.windowWidth > 0 && h.width > h.windowWidth {
			return h.windowWidth
		}
		return h.width
	}

	maxWidth := 1
	for _, line := range lines {
		if w := style.StringWidth(line); w > maxWidth {
			maxWidth = w
		}
	}

	if h.windowWidth > 0 && maxWidth > h.windowWidth {
		maxWidth = h.windowWidth
	}

	return maxWidth
}

func (h *HalfLinePaddedBox) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}

	h.designTokens = tokens

	if h.backgroundExplicit {
		return
	}

	if v := strings.TrimSpace(tokens.SurfaceRaised); v != "" {
		h.background = v
		return
	}

	if v := strings.TrimSpace(tokens.Background); v != "" {
		h.background = v
	}
}

var _ tui.Component = (*HalfLinePaddedBox)(nil)
