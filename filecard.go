package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	design "github.com/SCKelemen/design-system"
	tea "github.com/charmbracelet/bubbletea"
)

// FileCard displays a compact filename badge with diff stats.
type FileCard struct {
	filename     string
	additions    int
	deletions    int
	focused      bool
	designTokens *design.DesignTokens
}

// FileCardOption configures a FileCard.
type FileCardOption func(*FileCard)

// WithFileCardDesignTokens applies design-system tokens to a file card.
func WithFileCardDesignTokens(tokens *design.DesignTokens) FileCardOption {
	return func(c *FileCard) {
		if tokens != nil {
			c.designTokens = tokens
		}
	}
}

// WithFileCardTheme applies a named design-system theme.
func WithFileCardTheme(theme string) FileCardOption {
	return func(c *FileCard) {
		c.designTokens = designTokensForTheme(theme)
	}
}

// NewFileCard creates a new file card.
func NewFileCard(filename string, additions, deletions int, opts ...FileCardOption) *FileCard {
	c := &FileCard{
		filename:     filename,
		additions:    additions,
		deletions:    deletions,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Init initializes the component.
func (c *FileCard) Init() tea.Cmd {
	return nil
}

// Update handles Bubble Tea messages.
func (c *FileCard) Update(msg tea.Msg) (Component, tea.Cmd) {
	return c, nil
}

// View renders the file card.
func (c *FileCard) View() string {
	plainContent := c.plainContent()
	styledContent := c.styledContent()

	innerWidth := utf8.RuneCountInString(plainContent) + 2 // one space of padding on each side
	if innerWidth < 2 {
		innerWidth = 2
	}

	top := "╭" + strings.Repeat("─", innerWidth) + "╮"
	mid := "│ " + styledContent + " │"
	bot := "╰" + strings.Repeat("─", innerWidth) + "╯"

	return top + "\n" + mid + "\n" + bot
}

// Focus marks the component as focused.
func (c *FileCard) Focus() {
	c.focused = true
}

// Blur marks the component as unfocused.
func (c *FileCard) Blur() {
	c.focused = false
}

// Focused reports whether the component is focused.
func (c *FileCard) Focused() bool {
	return c.focused
}

// SetAdditions updates the added-line count.
func (c *FileCard) SetAdditions(n int) {
	c.additions = n
}

// SetDeletions updates the removed-line count.
func (c *FileCard) SetDeletions(n int) {
	c.deletions = n
}

// SetFilename updates the filename.
func (c *FileCard) SetFilename(s string) {
	c.filename = s
}

// NewFileCardRow renders multiple cards inline with a one-space gap.
func NewFileCardRow(cards ...*FileCard) string {
	nonNil := make([]*FileCard, 0, len(cards))
	for _, card := range cards {
		if card != nil {
			nonNil = append(nonNil, card)
		}
	}
	if len(nonNil) == 0 {
		return ""
	}

	rows := make([][]string, 0, len(nonNil))
	for _, card := range nonNil {
		rows = append(rows, strings.Split(card.View(), "\n"))
	}

	lineCount := len(rows[0])
	if lineCount == 0 {
		return ""
	}

	outLines := make([]string, lineCount)
	for i := 0; i < lineCount; i++ {
		parts := make([]string, 0, len(rows))
		for _, row := range rows {
			if i < len(row) {
				parts = append(parts, row[i])
			}
		}
		outLines[i] = strings.Join(parts, " ")
	}

	return strings.Join(outLines, "\n")
}

func (c *FileCard) plainContent() string {
	var b strings.Builder
	b.WriteString(c.filename)
	if c.additions > 0 {
		b.WriteString(fmt.Sprintf(" +%d", c.additions))
	}
	if c.deletions > 0 {
		b.WriteString(fmt.Sprintf(" -%d", c.deletions))
	}
	return b.String()
}

func (c *FileCard) styledContent() string {
	var b strings.Builder
	b.WriteString(c.filename)
	if c.additions > 0 {
		b.WriteString(" ")
		b.WriteString(ansiGreen)
		b.WriteString(fmt.Sprintf("+%d", c.additions))
		b.WriteString(ansiReset)
	}
	if c.deletions > 0 {
		b.WriteString(" ")
		b.WriteString(ansiRed)
		b.WriteString(fmt.Sprintf("-%d", c.deletions))
		b.WriteString(ansiReset)
	}
	return b.String()
}
