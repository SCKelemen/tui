package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const cellColorEpsilon = 3

// Cell represents a single terminal cell.
//
// Width is expected to be 1 or 2 for user-provided cells. Internally, the
// buffer uses Width == 0 as a continuation marker for the trailing half of a
// wide glyph.
type Cell struct {
	Rune          rune
	Grapheme      string
	Width         int
	Fg            lipgloss.Color
	Bg            lipgloss.Color
	Bold          bool
	Italic        bool
	Underline     bool
	Strikethrough bool
	Dim           bool
	Blink         bool
	Hyperlink     string
	Dirty         bool
}

// CellChange captures a changed cell position and value.
type CellChange struct {
	X, Y int
	Cell Cell
}

// CellBuffer manages a front/back terminal cell buffer for cell-level diffing.
type CellBuffer struct {
	width   int
	height  int
	front   []Cell
	back    []Cell
	changes []CellChange
}

// NewCellBuffer creates a new CellBuffer with the provided dimensions.
func NewCellBuffer(width, height int) *CellBuffer {
	cb := &CellBuffer{}
	cb.Resize(width, height)
	return cb
}

// Resize updates the buffer dimensions and clears both buffers.
func (cb *CellBuffer) Resize(width, height int) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	cb.width = width
	cb.height = height
	size := width * height
	cb.front = make([]Cell, size)
	cb.back = make([]Cell, size)
	for i := 0; i < size; i++ {
		cb.front[i] = blankCell()
		cb.back[i] = blankCell()
	}
	cb.changes = cb.changes[:0]
}

// SetCell writes a cell into the back buffer and marks it dirty if changed.
func (cb *CellBuffer) SetCell(x, y int, cell Cell) {
	if !cb.inBounds(x, y) {
		return
	}

	cell = normalizeCell(cell)
	if cell.Width == 2 && x >= cb.width-1 {
		return
	}

	cb.clearOverlapAt(x, y)
	if cell.Width == 2 {
		cb.clearOverlapAt(x+1, y)
	}

	idx := cb.index(x, y)
	cell.Dirty = !cellsEquivalent(cell, cb.front[idx])
	cb.back[idx] = cell

	if cell.Width == 2 {
		cont := continuationCell(cell)
		cont.Dirty = !cellsEquivalent(cont, cb.front[idx+1])
		cb.back[idx+1] = cont
	}
}

// WriteString writes text into the back buffer starting at x,y.
func (cb *CellBuffer) WriteString(x, y int, text string, fg, bg lipgloss.Color) {
	if y < 0 || y >= cb.height || x >= cb.width {
		return
	}
	if x < 0 {
		x = 0
	}

	col := x
	it := NewGraphemeIterator(text)
	for {
		cluster, ok := it.Next()
		if !ok {
			break
		}

		clusterText := cluster.String()
		if firstRuneInString(clusterText) == '\n' {
			break
		}
		if cluster.Width <= 0 {
			continue
		}
		if col+cluster.Width > cb.width {
			break
		}

		cb.SetCell(col, y, Cell{
			Rune:     firstRuneInString(clusterText),
			Grapheme: clusterText,
			Width:    cluster.Width,
			Fg:       fg,
			Bg:       bg,
		})
		col += cluster.Width
	}
}

// Clear resets the back buffer to blank cells.
func (cb *CellBuffer) Clear() {
	for i := range cb.back {
		blank := blankCell()
		blank.Dirty = !cellsEquivalent(blank, cb.front[i])
		cb.back[i] = blank
	}
	cb.changes = cb.changes[:0]
}

// Diff compares back and front buffers and returns only changed cells.
func (cb *CellBuffer) Diff() []CellChange {
	cb.changes = cb.changes[:0]
	for y := 0; y < cb.height; y++ {
		for x := 0; x < cb.width; x++ {
			idx := cb.index(x, y)
			backCell := cb.back[idx]
			if !backCell.Dirty && cellsEquivalent(backCell, cb.front[idx]) {
				continue
			}
			if cellsEquivalent(backCell, cb.front[idx]) {
				cb.back[idx].Dirty = false
				continue
			}
			cb.changes = append(cb.changes, CellChange{X: x, Y: y, Cell: backCell})
		}
	}
	return cb.changes
}

// Flush generates minimal ANSI output for changed cells and syncs buffers.
func (cb *CellBuffer) Flush() string {
	changes := cb.Diff()
	if len(changes) == 0 {
		cb.syncBuffers()
		return ""
	}

	var out strings.Builder
	state := ansiState{row: -1, col: -1}
	for i := 0; i < len(changes); {
		change := changes[i]
		start := i
		end := i + 1
		for end < len(changes) {
			next := changes[end]
			if next.Y != change.Y || next.X != changes[end-1].X+1 {
				break
			}
			end++
		}

		cb.writeRun(&out, &state, changes[start:end])
		i = end
	}

	if state.hyperlink != "" {
		out.WriteString(closeHyperlink())
	}
	if state.hasStyle {
		out.WriteString("\x1b[0m")
	}
	cb.syncBuffers()
	return out.String()
}

// CellAt returns the currently displayed cell from the front buffer.
func (cb *CellBuffer) CellAt(x, y int) Cell {
	if !cb.inBounds(x, y) {
		return blankCell()
	}
	return cb.front[cb.index(x, y)]
}

func (cb *CellBuffer) writeRun(out *strings.Builder, state *ansiState, run []CellChange) {
	if len(run) == 0 {
		return
	}

	moveCursor(out, state, run[0].Y+1, run[0].X+1)
	for _, change := range run {
		cell := change.Cell
		if cell.Width == 0 {
			continue
		}
		applyCellStyle(out, state, cell)
		out.WriteString(renderCellText(cell))
		advanceCursor(state, cell)
	}
}
func (cb *CellBuffer) clearOverlapAt(x, y int) {
	if !cb.inBounds(x, y) {
		return
	}

	idx := cb.index(x, y)
	current := cb.back[idx]
	if current.Width == 2 && x+1 < cb.width {
		nextIdx := cb.index(x+1, y)
		blankNext := blankCell()
		blankNext.Dirty = !cellsEquivalent(blankNext, cb.front[nextIdx])
		cb.back[nextIdx] = blankNext
	}

	if current.Width == 0 && x > 0 {
		prevIdx := cb.index(x-1, y)
		if cb.back[prevIdx].Width == 2 {
			blankPrev := blankCell()
			blankPrev.Dirty = !cellsEquivalent(blankPrev, cb.front[prevIdx])
			cb.back[prevIdx] = blankPrev
		}
	}

	if x > 0 {
		prevIdx := cb.index(x-1, y)
		if cb.back[prevIdx].Width == 2 {
			blankPrev := blankCell()
			blankPrev.Dirty = !cellsEquivalent(blankPrev, cb.front[prevIdx])
			cb.back[prevIdx] = blankPrev
		}
	}

	blank := blankCell()
	blank.Dirty = !cellsEquivalent(blank, cb.front[idx])
	cb.back[idx] = blank
}

func (cb *CellBuffer) syncBuffers() {
	if len(cb.front) != len(cb.back) {
		cb.front = make([]Cell, len(cb.back))
	}
	if len(cb.back) == 0 {
		cb.changes = cb.changes[:0]
		return
	}

	cb.front, cb.back = cb.back, cb.front
	if len(cb.back) != len(cb.front) {
		cb.back = make([]Cell, len(cb.front))
	}

	for i := range cb.front {
		cb.front[i].Dirty = false
	}
	copy(cb.back, cb.front)
	for i := range cb.back {
		cb.back[i].Dirty = false
	}
	cb.changes = cb.changes[:0]
}
func (cb *CellBuffer) inBounds(x, y int) bool {
	return x >= 0 && x < cb.width && y >= 0 && y < cb.height
}

func (cb *CellBuffer) index(x, y int) int {
	return y*cb.width + x
}

func blankCell() Cell {
	return Cell{Rune: ' ', Grapheme: " ", Width: 1}
}

func continuationCell(base Cell) Cell {
	base.Rune = 0
	base.Grapheme = ""
	base.Width = 0
	return base
}

func normalizeCell(cell Cell) Cell {
	if cell.Grapheme == "" && cell.Rune != 0 {
		cell.Grapheme = string(cell.Rune)
	}
	if cell.Width <= 0 {
		switch {
		case cell.Grapheme != "":
			cell.Width = StringWidth(cell.Grapheme)
		case cell.Rune != 0:
			cell.Width = runewidth.RuneWidth(cell.Rune)
		}
	}
	if cell.Width <= 0 {
		cell = blankCell()
	}
	if cell.Width > 2 {
		cell.Width = 2
	}
	if cell.Grapheme == "" && cell.Width > 0 {
		if cell.Rune == 0 {
			cell.Rune = ' '
		}
		cell.Grapheme = string(cell.Rune)
	}
	if cell.Rune == 0 && cell.Grapheme != "" {
		cell.Rune = firstRuneInString(cell.Grapheme)
	}
	return cell
}

func renderCellText(cell Cell) string {
	if cell.Grapheme != "" {
		return cell.Grapheme
	}
	if cell.Rune == 0 {
		return " "
	}
	return string(cell.Rune)
}

func cellsEquivalent(a, b Cell) bool {
	return renderCellText(a) == renderCellText(b) &&
		a.Width == b.Width &&
		colorsEqual(a.Fg, b.Fg, cellColorEpsilon) &&
		colorsEqual(a.Bg, b.Bg, cellColorEpsilon) &&
		a.Bold == b.Bold &&
		a.Italic == b.Italic &&
		a.Underline == b.Underline &&
		a.Strikethrough == b.Strikethrough &&
		a.Dim == b.Dim &&
		a.Blink == b.Blink &&
		a.Hyperlink == b.Hyperlink
}
func colorsEqual(a, b lipgloss.Color, epsilon int) bool {
	sa := string(a)
	sb := string(b)
	if sa == sb {
		return true
	}
	if sa == "" || sb == "" {
		return false
	}

	ra, ga, ba, okA := colorToRGB(sa)
	rb, gb, bb, okB := colorToRGB(sb)
	if !okA || !okB {
		return false
	}

	return channelDiff(ra, rb) <= epsilon && channelDiff(ga, gb) <= epsilon && channelDiff(ba, bb) <= epsilon
}

func channelDiff(a, b int) int {
	if a > b {
		return a - b
	}
	return b - a
}

func colorToRGB(value string) (int, int, int, bool) {
	if value == "" {
		return 0, 0, 0, false
	}

	if strings.HasPrefix(value, "#") {
		return parseHexColor(value)
	}

	n, err := strconv.Atoi(value)
	if err == nil && n >= 0 && n <= 255 {
		return ansi256ToRGB(n)
	}

	return 0, 0, 0, false
}

func parseHexColor(value string) (int, int, int, bool) {
	hex := strings.TrimPrefix(value, "#")
	switch len(hex) {
	case 3:
		r, err := strconv.ParseUint(strings.Repeat(string(hex[0]), 2), 16, 8)
		if err != nil {
			return 0, 0, 0, false
		}
		g, err := strconv.ParseUint(strings.Repeat(string(hex[1]), 2), 16, 8)
		if err != nil {
			return 0, 0, 0, false
		}
		b, err := strconv.ParseUint(strings.Repeat(string(hex[2]), 2), 16, 8)
		if err != nil {
			return 0, 0, 0, false
		}
		return int(r), int(g), int(b), true
	case 6:
		r, err := strconv.ParseUint(hex[0:2], 16, 8)
		if err != nil {
			return 0, 0, 0, false
		}
		g, err := strconv.ParseUint(hex[2:4], 16, 8)
		if err != nil {
			return 0, 0, 0, false
		}
		b, err := strconv.ParseUint(hex[4:6], 16, 8)
		if err != nil {
			return 0, 0, 0, false
		}
		return int(r), int(g), int(b), true
	default:
		return 0, 0, 0, false
	}
}

func ansi256ToRGB(n int) (int, int, int, bool) {
	base := [16][3]int{
		{0, 0, 0},
		{128, 0, 0},
		{0, 128, 0},
		{128, 128, 0},
		{0, 0, 128},
		{128, 0, 128},
		{0, 128, 128},
		{192, 192, 192},
		{128, 128, 128},
		{255, 0, 0},
		{0, 255, 0},
		{255, 255, 0},
		{0, 0, 255},
		{255, 0, 255},
		{0, 255, 255},
		{255, 255, 255},
	}

	switch {
	case n < 16:
		v := base[n]
		return v[0], v[1], v[2], true
	case n >= 16 && n <= 231:
		n -= 16
		r := n / 36
		g := (n / 6) % 6
		b := n % 6
		levels := [6]int{0, 95, 135, 175, 215, 255}
		return levels[r], levels[g], levels[b], true
	case n >= 232 && n <= 255:
		gray := 8 + (n-232)*10
		return gray, gray, gray, true
	default:
		return 0, 0, 0, false
	}
}

func cellANSIColorCode(prefix string, color lipgloss.Color) string {
	value := string(color)
	if value == "" {
		if prefix == "38" {
			return "39"
		}
		return "49"
	}

	if strings.HasPrefix(value, "#") {
		r, g, b, ok := parseHexColor(value)
		if ok {
			return fmt.Sprintf("%s;2;%d;%d;%d", prefix, r, g, b)
		}
	}

	n, err := strconv.Atoi(value)
	if err == nil && n >= 0 && n <= 255 {
		return fmt.Sprintf("%s;5;%d", prefix, n)
	}

	return ""
}

type ansiState struct {
	fg            lipgloss.Color
	bg            lipgloss.Color
	bold          bool
	italic        bool
	underline     bool
	strikethrough bool
	dim           bool
	blink         bool
	hyperlink     string
	hasStyle      bool
	row           int
	col           int
}

func moveCursor(out *strings.Builder, state *ansiState, row, col int) {
	if state.row == row && state.col == col {
		return
	}
	fmt.Fprintf(out, "\x1b[%d;%dH", row, col)
	state.row = row
	state.col = col
}

func advanceCursor(state *ansiState, cell Cell) {
	width := cell.Width
	if width <= 0 {
		width = 1
	}
	state.col += width
}

func applyCellStyle(out *strings.Builder, state *ansiState, cell Cell) {
	if state.hyperlink != cell.Hyperlink {
		if state.hyperlink != "" {
			out.WriteString(closeHyperlink())
		}
		if cell.Hyperlink != "" {
			out.WriteString(openHyperlink(cell.Hyperlink))
		}
		state.hyperlink = cell.Hyperlink
	}

	codes := make([]string, 0, 8)

	if state.bold != cell.Bold || state.dim != cell.Dim {
		codes = append(codes, "22")
		if cell.Bold {
			codes = append(codes, "1")
		}
		if cell.Dim {
			codes = append(codes, "2")
		}
	}
	if state.italic != cell.Italic {
		if cell.Italic {
			codes = append(codes, "3")
		} else {
			codes = append(codes, "23")
		}
	}
	if state.underline != cell.Underline {
		if cell.Underline {
			codes = append(codes, "4")
		} else {
			codes = append(codes, "24")
		}
	}
	if state.blink != cell.Blink {
		if cell.Blink {
			codes = append(codes, "5")
		} else {
			codes = append(codes, "25")
		}
	}
	if state.strikethrough != cell.Strikethrough {
		if cell.Strikethrough {
			codes = append(codes, "9")
		} else {
			codes = append(codes, "29")
		}
	}
	if !colorsEqual(state.fg, cell.Fg, 0) {
		if code := cellANSIColorCode("38", cell.Fg); code != "" {
			codes = append(codes, code)
		} else if string(cell.Fg) == "" {
			codes = append(codes, "39")
		}
	}
	if !colorsEqual(state.bg, cell.Bg, 0) {
		if code := cellANSIColorCode("48", cell.Bg); code != "" {
			codes = append(codes, code)
		} else if string(cell.Bg) == "" {
			codes = append(codes, "49")
		}
	}

	if len(codes) > 0 {
		out.WriteString("\x1b[")
		out.WriteString(strings.Join(codes, ";"))
		out.WriteByte('m')
		state.hasStyle = true
	}

	state.fg = cell.Fg
	state.bg = cell.Bg
	state.bold = cell.Bold
	state.italic = cell.Italic
	state.underline = cell.Underline
	state.strikethrough = cell.Strikethrough
	state.dim = cell.Dim
	state.blink = cell.Blink
}

// OSC 8 hyperlinks: https://iterm2.com/feature-reporting/Hyperlinks_in_Terminal_Emulators.html
func openHyperlink(target string) string {
	return "\x1b]8;;" + target + "\x1b\\"
}

func closeHyperlink() string {
	return "\x1b]8;;\x1b\\"
}
