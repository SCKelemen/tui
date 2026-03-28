package display

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/selection"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// CodeBlock displays source code with line numbers, syntax highlighting, and collapse/expand behavior.
type CodeBlock struct {
	width   int
	height  int
	focused bool

	selMgr                *selection.SelectionManager
	mouseSelectionEnabled bool

	operation   string
	filename    string
	summary     string
	lines       []string
	language    string
	highlighter *Highlighter

	expanded    bool
	maxLines    int
	startLine   int
	showPreview int

	// Design token colors
	surfaceColor   string // raised surface background
	borderColor    string // subtle border
	accentColor    string // filename, header accent
	mutedColor     string // line numbers, metadata
	syntaxKeyword  string
	syntaxFunction string
	syntaxString   string
	syntaxNumber   string
	syntaxComment  string
}

// CodeBlockOption configures a CodeBlock.
type CodeBlockOption func(*CodeBlock)

func WithCodeOperation(op string) CodeBlockOption {
	return func(cb *CodeBlock) { cb.operation = op }
}

func WithCodeFilename(name string) CodeBlockOption {
	return func(cb *CodeBlock) { cb.filename = name }
}

func WithCodeSummary(summary string) CodeBlockOption {
	return func(cb *CodeBlock) { cb.summary = summary }
}

func WithCodeLines(lines []string) CodeBlockOption {
	return func(cb *CodeBlock) { cb.lines = lines }
}

func WithCode(code string) CodeBlockOption {
	return func(cb *CodeBlock) { cb.lines = strings.Split(code, "\n") }
}

func WithLanguage(lang string) CodeBlockOption {
	return func(cb *CodeBlock) {
		cb.language = lang
		parsed := parseSyntaxLanguage(lang)
		if parsed == SyntaxLanguagePlain {
			cb.highlighter = nil
			return
		}
		cb.highlighter = NewHighlighter(parsed)
	}
}

func WithCodeBlockLanguage(lang SyntaxLanguage) CodeBlockOption {
	return func(cb *CodeBlock) {
		cb.language = string(lang)
		if lang == SyntaxLanguagePlain {
			cb.highlighter = nil
			return
		}
		cb.highlighter = NewHighlighter(lang)
	}
}

func WithCodeBlockFilename(filename string) CodeBlockOption {
	return func(cb *CodeBlock) {
		cb.filename = filename
		lang := DetectLanguage(filename)
		if lang == SyntaxLanguagePlain {
			cb.highlighter = nil
			return
		}
		cb.language = string(lang)
		cb.highlighter = NewHighlighter(lang)
	}
}

func WithStartLine(line int) CodeBlockOption {
	return func(cb *CodeBlock) { cb.startLine = line }
}

func WithExpanded(expanded bool) CodeBlockOption {
	return func(cb *CodeBlock) { cb.expanded = expanded }
}

func WithCodeMaxLines(max int) CodeBlockOption {
	return func(cb *CodeBlock) { cb.maxLines = max }
}

func WithPreviewLines(n int) CodeBlockOption {
	return func(cb *CodeBlock) { cb.showPreview = n }
}

func WithCodeBlockMouseSelection(enabled bool) CodeBlockOption {
	return func(cb *CodeBlock) { cb.mouseSelectionEnabled = enabled }
}

// WithCodeBlockDesignTokens applies design system tokens to the code block.
func WithCodeBlockDesignTokens(dt *design.DesignTokens) CodeBlockOption {
	return func(c *CodeBlock) {
		if dt == nil {
			return
		}
		if v := dt.SurfaceRaised; v != "" {
			c.surfaceColor = v
		}
		if v := dt.BorderSubtle; v != "" {
			c.borderColor = v
		}
		if v := dt.Accent; v != "" {
			c.accentColor = v
		}
		if v := dt.MutedColor; v != "" {
			c.mutedColor = v
		}
		if v := dt.SyntaxKeyword; v != "" {
			c.syntaxKeyword = v
		}
		if v := dt.SyntaxFunction; v != "" {
			c.syntaxFunction = v
		}
		if v := dt.SyntaxString; v != "" {
			c.syntaxString = v
		}
		if v := dt.SyntaxNumber; v != "" {
			c.syntaxNumber = v
		}
		if v := dt.SyntaxComment; v != "" {
			c.syntaxComment = v
		}
	}
}

func NewCodeBlock(opts ...CodeBlockOption) *CodeBlock {
	cb := &CodeBlock{
		operation:   "Code",
		startLine:   1,
		showPreview: 8,
		expanded:    false,
		selMgr:      selection.NewSelectionManager(),
	}

	cb.surfaceColor = "#31353D"
	cb.borderColor = "#3C414B"
	cb.accentColor = "#61AFEF"
	cb.mutedColor = "#7A818A"
	cb.syntaxKeyword = "#C678DD"
	cb.syntaxFunction = "#61AFEF"
	cb.syntaxString = "#98C379"
	cb.syntaxNumber = "#D19A66"
	cb.syntaxComment = "#7F848E"

	for _, opt := range opts {
		opt(cb)
	}

	return cb
}

func (cb *CodeBlock) Init() tea.Cmd { return nil }

func (cb *CodeBlock) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cb.width = msg.Width
		cb.height = msg.Height
	case tea.MouseMsg:
		if cb.mouseSelectionEnabled && cb.selMgr != nil {
			cb.selMgr.HandleMouse(msg)
		}
	case tea.KeyMsg:
		if cb.selMgr != nil && cb.selMgr.HasSelection() {
			switch msg.String() {
			case "ctrl+c":
				return cb, cb.selMgr.CopySelection()
			}
		}

		if !cb.focused {
			return cb, nil
		}

		if cb.selMgr != nil && cb.selMgr.HasSelection() {
			switch msg.String() {
			case "y":
				return cb, cb.selMgr.CopySelection()
			}
		}

		switch msg.String() {
		case "ctrl+o", "enter", " ":
			cb.Toggle()
		}
	}

	return cb, nil
}

func (cb *CodeBlock) View() string {
	if len(cb.lines) == 0 {
		if cb.selMgr != nil {
			cb.selMgr.SetRegion(selection.SelectableRegion{})
		}
		return ""
	}

	var b strings.Builder
	icon := cb.getOperationIcon()
	b.WriteString(fmt.Sprintf("%s %s%s%s", icon, style.ANSIBold, cb.operation, style.ANSIReset))
	headerLines := 1
	if cb.filename != "" {
		b.WriteString(fmt.Sprintf("(%s%s%s)", style.ANSICyan, cb.filename, style.ANSIReset))
	}
	b.WriteString("\n")

	if cb.summary != "" {
		b.WriteString(fmt.Sprintf("  %s⎿  %s%s\n", style.ANSIDim, cb.summary, style.ANSIReset))
		headerLines++
	}

	if cb.expanded {
		b.WriteString(cb.renderExpanded())
	} else {
		b.WriteString(cb.renderCollapsed())
	}

	cb.updateSelectionRegion(headerLines)
	return b.String()
}

func (cb *CodeBlock) Focus()           { cb.focused = true }
func (cb *CodeBlock) Blur()            { cb.focused = false }
func (cb *CodeBlock) Focused() bool    { return cb.focused }
func (cb *CodeBlock) Toggle()          { cb.expanded = !cb.expanded }
func (cb *CodeBlock) Expand()          { cb.expanded = true }
func (cb *CodeBlock) Collapse()        { cb.expanded = false }
func (cb *CodeBlock) IsExpanded() bool { return cb.expanded }

// GetSelectionManager returns the selection manager for this code block.
func (cb *CodeBlock) GetSelectionManager() *selection.SelectionManager { return cb.selMgr }

func (cb *CodeBlock) getOperationIcon() string {
	switch strings.ToLower(cb.operation) {
	case "write", "create":
		return style.ANSIGreen + "⏺" + style.ANSIReset
	case "read", "view":
		return style.ANSIBlue + "⏺" + style.ANSIReset
	case "edit", "update":
		return style.ANSIYellow + "⏺" + style.ANSIReset
	case "delete", "remove":
		return style.ANSIRed + "⏺" + style.ANSIReset
	default:
		return style.ANSIWhite + "⏺" + style.ANSIReset
	}
}

func (cb *CodeBlock) renderCollapsed() string {
	var b strings.Builder
	linesToShow := cb.showPreview
	if linesToShow > len(cb.lines) {
		linesToShow = len(cb.lines)
	}
	for i := 0; i < linesToShow; i++ {
		lineNum := cb.startLine + i
		content := cb.lines[i]
		if cb.highlighter != nil {
			content = cb.highlighter.HighlightLine(content)
		}
		if cb.selMgr != nil && cb.selMgr.HasSelection() {
			content = cb.selMgr.StyledLine(content, i)
		}
		b.WriteString(cb.renderLine(lineNum, content))
	}

	remainingLines := len(cb.lines) - linesToShow
	if remainingLines > 0 {
		b.WriteString(fmt.Sprintf("     %s… +%d lines (\033[3mctrl+o to expand%s%s)%s\n", style.ANSIDim, remainingLines, style.ANSIReset, style.ANSIDim, style.ANSIReset))
	}
	return b.String()
}

func (cb *CodeBlock) renderExpanded() string {
	var b strings.Builder
	linesToShow := len(cb.lines)
	if cb.maxLines > 0 && linesToShow > cb.maxLines {
		linesToShow = cb.maxLines
	}
	for i := 0; i < linesToShow; i++ {
		lineNum := cb.startLine + i
		content := cb.lines[i]
		if cb.highlighter != nil {
			content = cb.highlighter.HighlightLine(content)
		}
		if cb.selMgr != nil && cb.selMgr.HasSelection() {
			content = cb.selMgr.StyledLine(content, i)
		}
		b.WriteString(cb.renderLine(lineNum, content))
	}
	if cb.maxLines > 0 && len(cb.lines) > cb.maxLines {
		remainingLines := len(cb.lines) - cb.maxLines
		b.WriteString(fmt.Sprintf("     %s… +%d more lines (truncated)%s\n", style.ANSIDim, remainingLines, style.ANSIReset))
	}
	return b.String()
}

func (cb *CodeBlock) renderLine(lineNum int, content string) string {
	lineNumWidth := cb.lineNumWidth()
	return fmt.Sprintf("  %s%*d%s %s\n", style.ANSIDim, lineNumWidth, lineNum, style.ANSIReset, content)
}

func (cb *CodeBlock) lineNumWidth() int {
	maxLineNum := cb.startLine + len(cb.lines) - 1
	return len(fmt.Sprintf("%d", maxLineNum))
}

func (cb *CodeBlock) visibleLines() []string {
	if cb.expanded {
		linesToShow := len(cb.lines)
		if cb.maxLines > 0 && linesToShow > cb.maxLines {
			linesToShow = cb.maxLines
		}
		return cb.lines[:linesToShow]
	}
	linesToShow := cb.showPreview
	if linesToShow > len(cb.lines) {
		linesToShow = len(cb.lines)
	}
	return cb.lines[:linesToShow]
}

func (cb *CodeBlock) updateSelectionRegion(headerLines int) {
	if cb.selMgr == nil {
		return
	}
	visible := cb.visibleLines()
	contentLines := make([]string, len(visible))
	copy(contentLines, visible)

	startRow := headerLines
	endRow := headerLines
	if len(contentLines) > 0 {
		endRow = startRow + len(contentLines) - 1
	}

	region := selection.SelectableRegion{
		StartRow:     startRow,
		EndRow:       endRow,
		GutterWidth:  2 + cb.lineNumWidth() + 1,
		ContentLines: contentLines,
	}
	cb.selMgr.SetRegion(region)
}

var _ tui.Component = (*CodeBlock)(nil)
