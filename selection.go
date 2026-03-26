package tui

import (
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
)

// SelectableRegion defines the content area that can be selected.
type SelectableRegion struct {
	StartRow    int
	EndRow      int
	GutterWidth int
	ContentLines []string
}

// Selection stores the current selection coordinates.
type Selection struct {
	Active       bool
	StartX, StartY int
	EndX, EndY     int
}

// SelectionManager handles mouse-based text selection.
type SelectionManager struct {
	selection    Selection
	region       SelectableRegion
	offsetX      int
	offsetY      int
	hasSelection bool
}

// NewSelectionManager creates a new selection manager.
func NewSelectionManager() *SelectionManager {
	return &SelectionManager{}
}

// SetRegion sets the selectable region.
func (sm *SelectionManager) SetRegion(region SelectableRegion) {
	sm.region = region
}

// SetOffset sets the screen offset of the component.
func (sm *SelectionManager) SetOffset(x, y int) {
	sm.offsetX = x
	sm.offsetY = y
}

// HandleMouse handles mouse events for selection.
func (sm *SelectionManager) HandleMouse(msg tea.MouseMsg) bool {
	contentX, contentY, within := sm.toContentCoordinates(msg.X, msg.Y)

	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button != tea.MouseButtonLeft || !within {
			return false
		}

		sm.selection = Selection{
			Active: true,
			StartX: contentX,
			StartY: contentY,
			EndX:   contentX,
			EndY:   contentY,
		}
		sm.hasSelection = false
		return true

	case tea.MouseActionMotion:
		if !sm.selection.Active {
			return false
		}

		sm.selection.EndX = contentX
		sm.selection.EndY = contentY
		return within

	case tea.MouseActionRelease:
		if !sm.selection.Active {
			return false
		}

		sm.selection.EndX = contentX
		sm.selection.EndY = contentY
		sm.selection.Active = false
		sm.hasSelection = true
		return within
	}

	return false
}

// HasSelection reports whether there is a finalized selection.
func (sm *SelectionManager) HasSelection() bool {
	return sm.hasSelection
}

// GetSelectedText returns the selected text content.
func (sm *SelectionManager) GetSelectedText() string {
	if !sm.hasSelection || len(sm.region.ContentLines) == 0 {
		return ""
	}

	startRow, startCol, endRow, endCol := sm.normalizedSelection()
	if startRow < 0 || endRow < 0 || startRow >= len(sm.region.ContentLines) || endRow >= len(sm.region.ContentLines) {
		return ""
	}

	if startRow == endRow {
		line := sm.region.ContentLines[startRow]
		return substringByRune(line, startCol, endCol)
	}

	var selected []string
	firstLine := sm.region.ContentLines[startRow]
	selected = append(selected, substringByRune(firstLine, startCol, runeCount(firstLine)))

	for row := startRow + 1; row < endRow; row++ {
		selected = append(selected, sm.region.ContentLines[row])
	}

	lastLine := sm.region.ContentLines[endRow]
	selected = append(selected, substringByRune(lastLine, 0, endCol))

	return strings.Join(selected, "\n")
}

// ClearSelection resets the selection state.
func (sm *SelectionManager) ClearSelection() {
	sm.selection = Selection{}
	sm.hasSelection = false
}

// IsSelected reports whether the content-relative position is selected.
func (sm *SelectionManager) IsSelected(row, col int) bool {
	if !sm.hasSelection {
		return false
	}
	if row < 0 || row >= len(sm.region.ContentLines) || col < 0 {
		return false
	}

	startRow, startCol, endRow, endCol := sm.normalizedSelection()
	if row < startRow || row > endRow {
		return false
	}

	if startRow == endRow {
		return row == startRow && col >= startCol && col < endCol
	}

	if row == startRow {
		return col >= startCol
	}
	if row == endRow {
		return col < endCol
	}

	return true
}

// StyledLine applies reverse-video highlighting to selected spans on a line.
func (sm *SelectionManager) StyledLine(line string, row int) string {
	if !sm.hasSelection || line == "" {
		return line
	}

	reverseOn := string([]byte{0x1b}) + "[7m"
	reverseOff := string([]byte{0x1b}) + "[27m"

	var b strings.Builder
	selected := false

	for col, r := range []rune(line) {
		currentSelected := sm.IsSelected(row, col)
		if currentSelected && !selected {
			b.WriteString(reverseOn)
			selected = true
		}
		if !currentSelected && selected {
			b.WriteString(reverseOff)
			selected = false
		}
		b.WriteRune(r)
	}

	if selected {
		b.WriteString(reverseOff)
	}

	return b.String()
}

// CopySelection returns a command that copies selected text to clipboard.
func (sm *SelectionManager) CopySelection() tea.Cmd {
	text := sm.GetSelectedText()
	if text == "" {
		return nil
	}
	return WriteClipboard(text)
}

func (sm *SelectionManager) toContentCoordinates(x, y int) (contentX, contentY int, within bool) {
	rawX := x - sm.offsetX - sm.region.GutterWidth
	rawY := y - sm.offsetY

	if len(sm.region.ContentLines) == 0 {
		return 0, 0, false
	}

	minRow := sm.region.StartRow
	maxRow := sm.region.EndRow
	if maxRow < minRow {
		maxRow = minRow
	}

	within = rawY >= minRow && rawY <= maxRow

	clampedRow := clamp(rawY, minRow, maxRow)
	contentY = clampedRow - minRow
	if contentY < 0 {
		contentY = 0
	}
	if contentY >= len(sm.region.ContentLines) {
		contentY = len(sm.region.ContentLines) - 1
	}

	lineLen := runeCount(sm.region.ContentLines[contentY])
	contentX = clamp(rawX, 0, lineLen)

	return contentX, contentY, within
}

func (sm *SelectionManager) normalizedSelection() (startRow, startCol, endRow, endCol int) {
	startRow = sm.selection.StartY
	startCol = sm.selection.StartX
	endRow = sm.selection.EndY
	endCol = sm.selection.EndX

	if endRow < startRow || (endRow == startRow && endCol < startCol) {
		startRow, endRow = endRow, startRow
		startCol, endCol = endCol, startCol
	}

	startRow = clamp(startRow, 0, len(sm.region.ContentLines)-1)
	endRow = clamp(endRow, 0, len(sm.region.ContentLines)-1)

	startCol = clamp(startCol, 0, runeCount(sm.region.ContentLines[startRow]))
	endCol = clamp(endCol, 0, runeCount(sm.region.ContentLines[endRow]))

	return startRow, startCol, endRow, endCol
}

func substringByRune(s string, start, end int) string {
	runes := []rune(s)
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	if start > len(runes) {
		start = len(runes)
	}
	if end > len(runes) {
		end = len(runes)
	}
	if end < start {
		end = start
	}
	return string(runes[start:end])
}

func runeCount(s string) int {
	return utf8.RuneCountInString(s)
}

func clamp(v, min, max int) int {
	if max < min {
		max = min
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
