package container

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// Pane renders a panel with title header, content area, and optional footer.
type Pane struct {
	title        string
	width        int
	height       int
	content      string
	footer       string
	focused      bool
	designTokens *design.DesignTokens
}

// PaneOption configures a Pane.
type PaneOption func(*Pane)

// WithPaneWidth sets the pane width.
func WithPaneWidth(width int) PaneOption {
	return func(p *Pane) {
		if width >= 0 {
			p.width = width
		}
	}
}

// WithPaneHeight sets the pane height.
func WithPaneHeight(height int) PaneOption {
	return func(p *Pane) {
		if height >= 0 {
			p.height = height
		}
	}
}

// WithPaneContent sets pane body content.
func WithPaneContent(content string) PaneOption {
	return func(p *Pane) {
		p.content = content
	}
}

// WithPaneFooter sets optional footer text.
func WithPaneFooter(footer string) PaneOption {
	return func(p *Pane) {
		p.footer = strings.TrimSpace(footer)
	}
}

// WithPaneDesignTokens applies design tokens.
func WithPaneDesignTokens(tokens *design.DesignTokens) PaneOption {
	return func(p *Pane) {
		if tokens != nil {
			p.designTokens = tokens
		}
	}
}

// NewPane creates a new Pane component.
func NewPane(title string, opts ...PaneOption) *Pane {
	p := &Pane{
		title:        strings.TrimSpace(title),
		width:        60,
		height:       12,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Init initializes the component.
func (p *Pane) Init() tea.Cmd { return nil }

// Update handles Bubble Tea messages.
func (p *Pane) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		if p.width == 0 {
			p.width = m.Width
		}
		if p.height == 0 {
			p.height = m.Height
		}
	}
	return p, nil
}

// View renders the pane.
func (p *Pane) View() string {
	width := p.width
	if width <= 0 {
		width = 1
	}

	height := p.height
	if height <= 0 {
		height = 3
	}

	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}

	borderColor := style.ANSIDim
	titleColor := style.ANSIBold
	footerColor := style.ANSIDim
	if p.designTokens != nil {
		if c := style.Fg(p.designTokens.BorderSubtle); c != "" {
			borderColor = c
		}
		if c := style.Fg(p.designTokens.Accent); c != "" {
			titleColor = c + style.ANSIBold
		}
		if c := style.Fg(p.designTokens.MutedColor); c != "" {
			footerColor = c
		}
	}

	title := style.Truncate(p.title, innerWidth, "…")
	topFill := innerWidth - style.StringWidth(title)
	if topFill < 0 {
		topFill = 0
	}

	contentLines := strings.Split(p.content, "\n")
	bodyLines := height - 2
	if p.footer != "" {
		bodyLines--
	}
	if bodyLines < 0 {
		bodyLines = 0
	}

	var b strings.Builder
	b.WriteString(borderColor + "┌" + style.ANSIReset)
	b.WriteString(titleColor + title + style.ANSIReset)
	b.WriteString(borderColor + strings.Repeat("─", topFill) + "┐" + style.ANSIReset)
	b.WriteByte('\n')

	for i := 0; i < bodyLines; i++ {
		line := ""
		if i < len(contentLines) {
			line = contentLines[i]
		}
		line = style.Truncate(line, innerWidth, "…")
		pad := innerWidth - style.StringWidth(line)
		if pad < 0 {
			pad = 0
		}
		b.WriteString(borderColor + "│" + style.ANSIReset)
		b.WriteString(line)
		b.WriteString(strings.Repeat(" ", pad))
		b.WriteString(borderColor + "│" + style.ANSIReset)
		b.WriteByte('\n')
	}

	if p.footer != "" {
		footer := style.Truncate(p.footer, innerWidth, "…")
		pad := innerWidth - style.StringWidth(footer)
		if pad < 0 {
			pad = 0
		}
		b.WriteString(borderColor + "│" + style.ANSIReset)
		b.WriteString(footerColor + footer + style.ANSIReset)
		b.WriteString(strings.Repeat(" ", pad))
		b.WriteString(borderColor + "│" + style.ANSIReset)
		b.WriteByte('\n')
	}

	b.WriteString(borderColor + "└" + strings.Repeat("─", innerWidth) + "┘" + style.ANSIReset)
	return b.String()
}

// Focus marks the component as focused.
func (p *Pane) Focus() { p.focused = true }

// Blur marks the component as unfocused.
func (p *Pane) Blur() { p.focused = false }

// Focused reports whether the component is focused.
func (p *Pane) Focused() bool { return p.focused }

var _ tui.Component = (*Pane)(nil)
