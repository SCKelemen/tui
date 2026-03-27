package display

import (
	"fmt"
	"strings"

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

	filename  string
	operation string
	summary   string
	lines     []DiffLine
	oldStart  int
	newStart  int

	expanded    bool
	showContext int
	maxLines    int

	selMgr         *selection.SelectionManager
	mouseSelection bool
}

type DiffBlockOption func(*DiffBlock)

func WithDiffFilename(name string) DiffBlockOption { return func(db *DiffBlock) { db.filename = name } }
func WithDiffOperation(op string) DiffBlockOption  { return func(db *DiffBlock) { db.operation = op } }
func WithDiffSummary(summary string) DiffBlockOption {
	return func(db *DiffBlock) { db.summary = summary }
}
func WithDiffLines(lines []DiffLine) DiffBlockOption { return func(db *DiffBlock) { db.lines = lines } }
func WithDiffExpanded(expanded bool) DiffBlockOption {
	return func(db *DiffBlock) { db.expanded = expanded }
}
func WithDiffContext(n int) DiffBlockOption    { return func(db *DiffBlock) { db.showContext = n } }
func WithDiffMaxLines(max int) DiffBlockOption { return func(db *DiffBlock) { db.maxLines = max } }
func WithDiffBlockMouseSelection(enabled bool) DiffBlockOption {
	return func(db *DiffBlock) { db.mouseSelection = enabled }
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
	icon := style.ANSIYellow + "⏺" + style.ANSIReset
	b.WriteString(fmt.Sprintf("%s %s%s%s", icon, style.ANSIBold, db.operation, style.ANSIReset))
	if db.filename != "" {
		b.WriteString(fmt.Sprintf("(\033[36m%s%s)", db.filename, style.ANSIReset))
	}
	b.WriteString("\n")

	added, removed := db.countChanges()
	if db.summary != "" {
		b.WriteString(fmt.Sprintf("  %s⎿  %s%s\n", style.ANSIDim, db.summary, style.ANSIReset))
	} else {
		b.WriteString(fmt.Sprintf("  %s⎿  Added %d lines, removed %d lines%s\n", style.ANSIDim, added, removed, style.ANSIReset))
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
			b.WriteString(fmt.Sprintf("     %s… +%d more lines (truncated)%s\n", style.ANSIDim, remaining, style.ANSIReset))
		} else {
			b.WriteString(fmt.Sprintf("     %s… +%d more lines (\033[3mctrl+o to expand%s%s)%s\n", style.ANSIDim, remaining, style.ANSIReset, style.ANSIDim, style.ANSIReset))
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
	content := line.Content
	if db.mouseSelection && db.selMgr != nil && db.selMgr.HasSelection() {
		content = db.selMgr.StyledLine(content, row)
	}

	switch line.Type {
	case DiffAdded:
		return fmt.Sprintf("  %s %s+%s%s\n", lineNumStr, style.ANSIGreen, content, style.ANSIReset)
	case DiffRemoved:
		return fmt.Sprintf("  %s %s-%s%s\n", lineNumStr, style.ANSIRed, content, style.ANSIReset)
	case DiffUnchanged:
		return fmt.Sprintf("  %s  %s\n", lineNumStr, content)
	default:
		return fmt.Sprintf("        %s\n", content)
	}
}

var _ tui.Component = (*DiffBlock)(nil)
