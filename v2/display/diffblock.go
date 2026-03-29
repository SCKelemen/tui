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

// DiffLine represents a single line in a diff.
type DiffLine struct {
	Type    DiffType
	Content string
	LineNum int
}

// DiffType indicates the type of diff line.
type DiffType int

const (
	DiffUnchanged DiffType = iota
	DiffAdded
	DiffRemoved
)

// DiffBlock displays code changes with +/- indicators.
type DiffBlock struct {
	width   int
	height  int
	focused bool

	filename    string
	operation   string
	summary     string
	lines       []DiffLine
	oldStart    int
	newStart    int
	language    string
	highlighter *Highlighter
	expanded    bool
	showContext     int
	maxLines        int
	hideLineNumbers bool

	selMgr         *selection.SelectionManager
	mouseSelection bool
	// Design token colors
	addBgColor     string // background for added lines
	removeBgColor  string // background for removed lines
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

type DiffBlockOption func(*DiffBlock)

func WithDiffFilename(name string) DiffBlockOption { return func(db *DiffBlock) { db.filename = name } }

func WithDiffBlockFilename(name string) DiffBlockOption {
	return func(db *DiffBlock) {
		db.filename = name
		lang := DetectLanguage(name)
		db.language = string(lang)
		if lang == SyntaxLanguagePlain {
			db.highlighter = nil
			return
		}
		db.highlighter = NewHighlighter(lang)
	}
}

func WithDiffBlockLanguage(lang SyntaxLanguage) DiffBlockOption {
	return func(db *DiffBlock) {
		db.language = string(lang)
		if lang == SyntaxLanguagePlain {
			db.highlighter = nil
			return
		}
		db.highlighter = NewHighlighter(lang)
	}
}
func WithDiffOperation(op string) DiffBlockOption { return func(db *DiffBlock) { db.operation = op } }
func WithDiffSummary(summary string) DiffBlockOption {
	return func(db *DiffBlock) { db.summary = summary }
}
func WithDiffLines(lines []DiffLine) DiffBlockOption { return func(db *DiffBlock) { db.lines = lines } }
func WithDiffExpanded(expanded bool) DiffBlockOption {
	return func(db *DiffBlock) { db.expanded = expanded }
}
func WithDiffContext(n int) DiffBlockOption    { return func(db *DiffBlock) { db.showContext = n } }
func WithDiffMaxLines(max int) DiffBlockOption { return func(db *DiffBlock) { db.maxLines = max } }

// WithDiffHideLineNumbers hides line numbers from the rendered output.
// This makes terminal-native copy produce clean diff text.
func WithDiffHideLineNumbers(hide bool) DiffBlockOption {
	return func(d *DiffBlock) { d.hideLineNumbers = hide }
}

func WithDiffBlockMouseSelection(enabled bool) DiffBlockOption {
	return func(db *DiffBlock) { db.mouseSelection = enabled }
}
// WithDiffBlockDesignTokens applies design system tokens to the diff block.
func WithDiffBlockDesignTokens(dt *design.DesignTokens) DiffBlockOption {
	return func(d *DiffBlock) {
		if dt == nil {
			return
		}
		if v := dt.DiffAddBackground; v != "" {
			d.addBgColor = v
		}
		if v := dt.DiffRemoveBackground; v != "" {
			d.removeBgColor = v
		}
		if v := dt.SurfaceRaised; v != "" {
			d.surfaceColor = v
		}
		if v := dt.BorderSubtle; v != "" {
			d.borderColor = v
		}
		if v := dt.Accent; v != "" {
			d.accentColor = v
		}
		if v := dt.MutedColor; v != "" {
			d.mutedColor = v
		}
		if v := dt.SyntaxKeyword; v != "" {
			d.syntaxKeyword = v
		}
		if v := dt.SyntaxFunction; v != "" {
			d.syntaxFunction = v
		}
		if v := dt.SyntaxString; v != "" {
			d.syntaxString = v
		}
		if v := dt.SyntaxNumber; v != "" {
			d.syntaxNumber = v
		}
		if v := dt.SyntaxComment; v != "" {
			d.syntaxComment = v
		}
	}
}

func NewDiffBlock(opts ...DiffBlockOption) *DiffBlock {
	db := &DiffBlock{
		operation:   "Edit",
		showContext: 3,
		expanded:    false,
		oldStart:    1,
		newStart:    1,
		selMgr:      selection.NewSelectionManager(),
	}
	db.addBgColor = "#314D3B"
	db.removeBgColor = "#5F3439"
	db.surfaceColor = "#31353D"
	db.borderColor = "#3C414B"
	db.accentColor = "#61AFEF"
	db.mutedColor = "#7A818A"
	db.syntaxKeyword = "#C678DD"
	db.syntaxFunction = "#61AFEF"
	db.syntaxString = "#98C379"
	db.syntaxNumber = "#D19A66"
	db.syntaxComment = "#7F848E"
	for _, opt := range opts {
		opt(db)
	}
	return db
}

func NewDiffBlockFromStrings(old, new string, opts ...DiffBlockOption) *DiffBlock {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")
	diffLines := simpleDiff(oldLines, newLines)
	db := NewDiffBlock(opts...)
	db.lines = diffLines
	return db
}

func simpleDiff(oldLines, newLines []string) []DiffLine {
	var result []DiffLine
	commonPrefix := 0
	for commonPrefix < len(oldLines) && commonPrefix < len(newLines) && oldLines[commonPrefix] == newLines[commonPrefix] {
		result = append(result, DiffLine{Type: DiffUnchanged, Content: oldLines[commonPrefix], LineNum: commonPrefix + 1})
		commonPrefix++
	}

	commonSuffix := 0
	oldRemaining := len(oldLines) - commonPrefix
	newRemaining := len(newLines) - commonPrefix
	for commonSuffix < oldRemaining && commonSuffix < newRemaining && oldLines[len(oldLines)-1-commonSuffix] == newLines[len(newLines)-1-commonSuffix] {
		commonSuffix++
	}

	for i := commonPrefix; i < len(oldLines)-commonSuffix; i++ {
		result = append(result, DiffLine{Type: DiffRemoved, Content: oldLines[i], LineNum: i + 1})
	}
	for i := commonPrefix; i < len(newLines)-commonSuffix; i++ {
		result = append(result, DiffLine{Type: DiffAdded, Content: newLines[i], LineNum: i + 1})
	}
	for i := 0; i < commonSuffix; i++ {
		idx := len(oldLines) - commonSuffix + i
		result = append(result, DiffLine{Type: DiffUnchanged, Content: oldLines[idx], LineNum: idx + 1})
	}

	return result
}

func (db *DiffBlock) Init() tea.Cmd { return nil }

func (db *DiffBlock) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		db.width = msg.Width
		db.height = msg.Height
	case tea.MouseMsg:
		if db.mouseSelection && db.selMgr != nil {
			db.selMgr.HandleMouse(msg)
		}
	case tea.KeyMsg:
		if db.mouseSelection && db.selMgr != nil && db.selMgr.HasSelection() {
			switch msg.String() {
			case "y":
				if db.focused {
					return db, db.selMgr.CopySelection()
				}
			case "ctrl+c":
				return db, db.selMgr.CopySelection()
			}
		}
		if !db.focused {
			return db, nil
		}
		switch msg.String() {
		case "ctrl+o", "enter", " ":
			db.Toggle()
		}
	}
	return db, nil
}

func (db *DiffBlock) View() string {
	if len(db.lines) == 0 {
		return ""
	}

	var b strings.Builder
	accent := db.colorFg(db.accentColor, style.ANSIYellow)
	muted := db.colorFg(db.mutedColor, style.ANSIDim)
	border := db.colorFg(db.borderColor, "")

	icon := accent + "⏺" + style.ANSIReset
	b.WriteString(fmt.Sprintf("%s %s%s%s", icon, style.ANSIBold, db.operation, style.ANSIReset))
	if db.filename != "" {
		b.WriteString(fmt.Sprintf("(%s%s%s)", accent, db.filename, style.ANSIReset))
	}
	b.WriteString("\n")

	added, removed := db.countChanges()
	if db.summary != "" {
		b.WriteString(fmt.Sprintf("  %s⎿  %s%s\n", muted, db.summary, style.ANSIReset))
	} else {
		b.WriteString(fmt.Sprintf("  %s⎿  Added %d lines, removed %d lines%s\n", muted, added, removed, style.ANSIReset))
	}
	visibleLines, remaining, isTruncated := db.visibleLines()
	if db.mouseSelection && db.selMgr != nil {
		contentLines := make([]string, 0, len(visibleLines))
		for _, line := range visibleLines {
			contentLines = append(contentLines, line.Content)
		}
		const headerLines = 2
		startRow := headerLines
		endRow := startRow + len(contentLines) - 1
		if len(contentLines) == 0 {
			endRow = startRow
		}
		db.selMgr.SetRegion(selection.SelectableRegion{StartRow: startRow, EndRow: endRow, GutterWidth: 9, ContentLines: contentLines})
	}

	for i, line := range visibleLines {
		b.WriteString(db.renderDiffLine(line, i))
	}

	if remaining > 0 {
		if db.expanded && isTruncated {
			b.WriteString(fmt.Sprintf("     %s%s…%s %s+%d more lines (truncated)%s\n", border, muted, style.ANSIReset, muted, remaining, style.ANSIReset))
		} else {
			b.WriteString(fmt.Sprintf("     %s%s…%s %s+%d more lines (\033[3mctrl+o to expand%s%s)%s\n", border, muted, style.ANSIReset, muted, remaining, style.ANSIReset, muted, style.ANSIReset))
		}
	}
	return b.String()
}

func (db *DiffBlock) Focus()                                           { db.focused = true }
func (db *DiffBlock) Blur()                                            { db.focused = false }
func (db *DiffBlock) Focused() bool                                    { return db.focused }
func (db *DiffBlock) GetSelectionManager() *selection.SelectionManager { return db.selMgr }
func (db *DiffBlock) Toggle()                                          { db.expanded = !db.expanded }
func (db *DiffBlock) Expand()                                          { db.expanded = true }
func (db *DiffBlock) Collapse()                                        { db.expanded = false }
func (db *DiffBlock) IsExpanded() bool                                 { return db.expanded }

func (db *DiffBlock) countChanges() (added, removed int) {
	for _, line := range db.lines {
		switch line.Type {
		case DiffAdded:
			added++
		case DiffRemoved:
			removed++
		}
	}
	return
}

func (db *DiffBlock) visibleLines() ([]DiffLine, int, bool) {
	if db.expanded {
		linesToShow := len(db.lines)
		if db.maxLines > 0 && linesToShow > db.maxLines {
			linesToShow = db.maxLines
		}
		if db.maxLines > 0 && len(db.lines) > db.maxLines {
			remaining := len(db.lines) - db.maxLines
			return db.lines[:linesToShow], remaining, true
		}
		return db.lines[:linesToShow], 0, false
	}

	visible := make([]DiffLine, 0, len(db.lines))
	maxPreview := 15
	hasChanges := false
	for i, line := range db.lines {
		if line.Type == DiffUnchanged {
			showContext := false
			contextWindow := 2
			for j := i - contextWindow; j <= i+contextWindow; j++ {
				if j >= 0 && j < len(db.lines) && db.lines[j].Type != DiffUnchanged {
					showContext = true
					break
				}
			}
			if !showContext && hasChanges {
				break
			}
			if !showContext {
				continue
			}
		} else {
			hasChanges = true
		}

		if len(visible) >= maxPreview {
			break
		}
		visible = append(visible, line)
	}

	if len(db.lines) > len(visible) {
		return visible, len(db.lines) - len(visible), false
	}
	return visible, 0, false
}

func (db *DiffBlock) renderDiffLine(line DiffLine, row int) string {
	lineNumStr := fmt.Sprintf("%6d", line.LineNum)
	content := db.highlightDiffContent(line)
	if db.mouseSelection && db.selMgr != nil && db.selMgr.HasSelection() {
		content = db.selMgr.StyledLine(content, row)
	}

	lineNumColor := db.colorFg(db.mutedColor, style.ANSIDim)
	addColor := style.ANSIGreen
	removeColor := style.ANSIRed
	addBg := db.colorBg(db.addBgColor, "")
	removeBg := db.colorBg(db.removeBgColor, "")
	surfaceBg := db.colorBg(db.surfaceColor, "")

	if db.hideLineNumbers {
		switch line.Type {
		case DiffAdded:
			return fmt.Sprintf("  %s+%s %s%s%s%s\n", addColor, style.ANSIReset, addBg, db.tintHighlightedContent(content, addColor), style.ANSIReset, surfaceBg)
		case DiffRemoved:
			return fmt.Sprintf("  %s-%s %s%s%s%s\n", removeColor, style.ANSIReset, removeBg, db.tintHighlightedContent(content, removeColor), style.ANSIReset, surfaceBg)
		case DiffUnchanged:
			return fmt.Sprintf("    %s%s%s\n", surfaceBg, content, style.ANSIReset)
		default:
			return fmt.Sprintf("    %s\n", content)
		}
	}

	switch line.Type {
	case DiffAdded:
		return fmt.Sprintf("  %s%s%s %s+%s %s%s%s%s\n", lineNumColor, lineNumStr, style.ANSIReset, addColor, style.ANSIReset, addBg, db.tintHighlightedContent(content, addColor), style.ANSIReset, surfaceBg)
	case DiffRemoved:
		return fmt.Sprintf("  %s%s%s %s-%s %s%s%s%s\n", lineNumColor, lineNumStr, style.ANSIReset, removeColor, style.ANSIReset, removeBg, db.tintHighlightedContent(content, removeColor), style.ANSIReset, surfaceBg)
	case DiffUnchanged:
		return fmt.Sprintf("  %s%s%s  %s%s%s\n", lineNumColor, lineNumStr, style.ANSIReset, surfaceBg, content, style.ANSIReset)
	default:
		return fmt.Sprintf("        %s\n", content)
	}
}
func (db *DiffBlock) highlightDiffContent(line DiffLine) string {
	content := line.Content
	if db.highlighter == nil {
		return content
	}

	trimmed := content
	if len(trimmed) > 0 {
		switch line.Type {
		case DiffAdded:
			if trimmed[0] == '+' {
				trimmed = trimmed[1:]
			}
		case DiffRemoved:
			if trimmed[0] == '-' {
				trimmed = trimmed[1:]
			}
		}
	}
	highlighted := db.highlighter.HighlightLine(trimmed)
	return db.applySyntaxTokenColors(highlighted)
}

func (db *DiffBlock) tintHighlightedContent(content, diffColor string) string {
	if content == "" || diffColor == "" {
		return content
	}
	tinted := strings.ReplaceAll(content, style.ANSIReset, style.ANSIReset+diffColor)
	return diffColor + tinted + style.ANSIReset
}

func (db *DiffBlock) applySyntaxTokenColors(content string) string {
	if content == "" {
		return content
	}

	replacements := map[string]string{
		style.ANSIBold + style.ANSIBlue: db.colorFg(db.syntaxKeyword, style.ANSIBold+style.ANSIBlue),
		style.ANSICyan:                  db.colorFg(db.syntaxFunction, style.ANSICyan),
		style.ANSIGreen:                 db.colorFg(db.syntaxString, style.ANSIGreen),
		style.ANSIYellow:                db.colorFg(db.syntaxNumber, style.ANSIYellow),
		style.ANSIDim:                   db.colorFg(db.syntaxComment, style.ANSIDim),
	}
	for from, to := range replacements {
		content = strings.ReplaceAll(content, from, to)
	}
	return content
}

func (db *DiffBlock) colorFg(hex, fallback string) string {
	if c := style.Fg(hex); c != "" {
		return c
	}
	return fallback
}

func (db *DiffBlock) colorBg(hex, fallback string) string {
	if c := style.Bg(hex); c != "" {
		return c
	}
	return fallback
}

var _ tui.Component = (*DiffBlock)(nil)