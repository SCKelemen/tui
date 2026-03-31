package display

import (
	"reflect"
	"sort"
	"strings"

	design "github.com/SCKelemen/design-system"
	"github.com/charmbracelet/lipgloss"
)

// DiagramBox represents a box/node in the diagram.
type DiagramBox struct {
	ID       string
	Title    string
	Lines    []string
	Row      int
	Col      int
	MinWidth int
}

// DiagramArrow connects two boxes.
type DiagramArrow struct {
	From      string
	To        string
	Label     string
	Direction ArrowDirection
	Style     ArrowStyle
}

// ArrowDirection controls arrowhead orientation.
type ArrowDirection int

const (
	ArrowBidirectional ArrowDirection = iota
	ArrowLeftToRight
	ArrowRightToLeft
	ArrowTopToBottom
	ArrowBottomToTop
)

// ArrowStyle controls the stroke style used for arrow lines.
type ArrowStyle int

const (
	ArrowSolid ArrowStyle = iota
	ArrowDashed
	ArrowDouble
)

// Diagram is the main architecture diagram component.
type Diagram struct {
	boxes   []DiagramBox
	arrows  []DiagramArrow
	title   string
	tokens  *design.DesignTokens
	padding int
}

// NewDiagram creates a diagram with sane defaults.
func NewDiagram() *Diagram {
	return &Diagram{padding: 4}
}

// SetTitle sets an optional title shown above the diagram.
func (d *Diagram) SetTitle(title string) *Diagram {
	d.title = title
	return d
}

// AddBox appends a box to the diagram definition.
func (d *Diagram) AddBox(box DiagramBox) *Diagram {
	d.boxes = append(d.boxes, box)
	return d
}

// AddArrow appends an arrow connection to the diagram definition.
func (d *Diagram) AddArrow(arrow DiagramArrow) *Diagram {
	d.arrows = append(d.arrows, arrow)
	return d
}

// WithDesignTokens applies optional design tokens for colorized output.
func (d *Diagram) WithDesignTokens(tokens *design.DesignTokens) *Diagram {
	d.tokens = tokens
	return d
}

// WithPadding sets horizontal padding between boxes.
func (d *Diagram) WithPadding(p int) *Diagram {
	if p > 0 {
		d.padding = p
	}
	return d
}

// View is a Bubble Tea friendly alias for Render.
func (d *Diagram) View() string {
	return d.Render()
}

// Render creates the full ASCII/Unicode diagram.
func (d *Diagram) Render() string {
	if len(d.boxes) == 0 {
		if strings.TrimSpace(d.title) == "" {
			return ""
		}
		return d.renderTitleLine(strings.TrimSpace(d.title), runeLen(strings.TrimSpace(d.title)))
	}

	layout := d.computeLayout()
	grid := newRuneGrid(maxInt(layout.width, 1), maxInt(layout.height, 1))

	for _, pb := range layout.ordered {
		d.drawBox(grid, pb)
	}
	for _, arrow := range d.arrows {
		d.drawArrow(grid, layout.byID, arrow)
	}

	lines := d.gridToLines(grid)
	if strings.TrimSpace(d.title) == "" {
		return strings.Join(lines, "\n")
	}

title := d.renderTitleLine(d.title, grid.width)
	return strings.TrimRight(title+"\n"+strings.Join(lines, "\n"), "\n")
}

type placedBox struct {
	box    DiagramBox
	x      int
	y      int
	width  int
	height int
}

type diagramLayout struct {
	ordered []placedBox
	byID    map[string]placedBox
	width   int
	height  int
}

func (d *Diagram) computeLayout() diagramLayout {
	byRow := make(map[int][]DiagramBox)
	for _, box := range d.boxes {
		if strings.TrimSpace(box.ID) == "" {
			continue
		}
		byRow[box.Row] = append(byRow[box.Row], box)
	}

	rows := make([]int, 0, len(byRow))
	for row := range byRow {
		rows = append(rows, row)
	}
	sort.Ints(rows)

	ordered := make([]placedBox, 0, len(d.boxes))
	rowPlaced := make(map[int][]placedBox, len(rows))
	rowHeights := make(map[int]int, len(rows))
	maxWidth := 0

	for _, row := range rows {
		boxes := append([]DiagramBox(nil), byRow[row]...)
		sort.SliceStable(boxes, func(i, j int) bool {
			if boxes[i].Col == boxes[j].Col {
				return boxes[i].ID < boxes[j].ID
			}
			return boxes[i].Col < boxes[j].Col
		})

		x := 0
		placed := make([]placedBox, 0, len(boxes))
		rowMaxH := 0

		for i, box := range boxes {
			w, h := computeBoxSize(box)
			if i > 0 {
				prev := placed[i-1]
				x += prev.width + d.arrowGap(prev.box.ID, box.ID)
			}
			pb := placedBox{box: box, x: x, width: w, height: h}
			placed = append(placed, pb)
			if h > rowMaxH {
				rowMaxH = h
			}
		}

		rowPlaced[row] = placed
		rowHeights[row] = rowMaxH
		for _, pb := range placed {
			if pb.x+pb.width > maxWidth {
				maxWidth = pb.x + pb.width
			}
		}
	}

	currentY := 0
	baseRowGap := 4
	for i, row := range rows {
		placed := rowPlaced[row]
		rowH := rowHeights[row]
		for _, pb := range placed {
			pb.y = currentY + (rowH-pb.height)/2
			ordered = append(ordered, pb)
		}
		currentY += rowH
		if i < len(rows)-1 {
			currentY += baseRowGap
		}
	}

	byID := make(map[string]placedBox, len(ordered))
	maxHeight := 0
	for _, pb := range ordered {
		byID[pb.box.ID] = pb
		if pb.y+pb.height > maxHeight {
			maxHeight = pb.y + pb.height
		}
	}

	if maxHeight == 0 {
		maxHeight = 1
	}
	if maxWidth == 0 {
		maxWidth = 1
	}

	return diagramLayout{
		ordered: ordered,
		byID:    byID,
		width:   maxWidth,
		height:  maxHeight,
	}
}

func computeBoxSize(box DiagramBox) (width, height int) {
	maxContent := runeLen(box.Title)
	for _, line := range box.Lines {
		if l := runeLen(line); l > maxContent {
			maxContent = l
		}
	}
	if box.MinWidth > maxContent {
		maxContent = box.MinWidth
	}
	width = maxContent + 4
	if width < 6 {
		width = 6
	}
	height = 3 + len(box.Lines)
	if height < 3 {
		height = 3
	}
	return width, height
}

func (d *Diagram) arrowGap(leftID, rightID string) int {
	gap := maxInt(d.padding, 2)
	for _, arrow := range d.arrows {
		if (arrow.From == leftID && arrow.To == rightID) || (arrow.From == rightID && arrow.To == leftID) {
			need := runeLen(arrow.Label) + 4
			if need > gap {
				gap = need
			}
		}
	}
	return gap
}

func (d *Diagram) drawBox(grid *runeGrid, pb placedBox) {
	x := pb.x
	y := pb.y
	w := pb.width
	h := pb.height

	grid.set(x, y, '┌', roleBoxBorder)
	grid.set(x+w-1, y, '┐', roleBoxBorder)
	for cx := x + 1; cx < x+w-1; cx++ {
		grid.set(cx, y, '─', roleBoxBorder)
	}

	for cy := y + 1; cy < y+h-1; cy++ {
		grid.set(x, cy, '│', roleBoxBorder)
		grid.set(x+w-1, cy, '│', roleBoxBorder)
	}

	grid.set(x, y+h-1, '└', roleBoxBorder)
	grid.set(x+w-1, y+h-1, '┘', roleBoxBorder)
	for cx := x + 1; cx < x+w-1; cx++ {
		grid.set(cx, y+h-1, '─', roleBoxBorder)
	}

	innerW := w - 2
	title := padOrTrim(pb.box.Title, innerW)
	grid.writeText(x+1, y+1, title, roleBoxTitle)
	for i, line := range pb.box.Lines {
		text := padOrTrim(line, innerW)
		grid.writeText(x+1, y+2+i, text, roleBoxBody)
	}
}

func (d *Diagram) drawArrow(grid *runeGrid, byID map[string]placedBox, arrow DiagramArrow) {
	from, okFrom := byID[arrow.From]
	to, okTo := byID[arrow.To]
	if !okFrom || !okTo {
		return
	}

	if from.box.Row == to.box.Row {
		d.drawHorizontalArrow(grid, from, to, arrow)
		return
	}
	d.drawVerticalArrow(grid, from, to, arrow)
}

func (d *Diagram) drawHorizontalArrow(grid *runeGrid, from, to placedBox, arrow DiagramArrow) {
	left := from
	right := to
	if from.x > to.x {
		left = to
		right = from
	}

	xStart := left.x + left.width
	xEnd := right.x - 1
	if xEnd <= xStart {
		return
	}

	y := (from.y+from.height/2 + to.y+to.height/2) / 2
	lineRune := horizontalLineRune(arrow.Style)
	for x := xStart; x <= xEnd; x++ {
		grid.set(x, y, lineRune, roleArrowLine)
	}

	switch arrow.Direction {
	case ArrowBidirectional:
		grid.set(xStart, y, '◄', roleArrowLine)
		grid.set(xEnd, y, '►', roleArrowLine)
	case ArrowRightToLeft:
		grid.set(xStart, y, '◄', roleArrowLine)
	case ArrowTopToBottom, ArrowBottomToTop:
		// Keep line-only if direction doesn't match axis.
	default:
		grid.set(xEnd, y, '►', roleArrowLine)
	}

	if strings.TrimSpace(arrow.Label) != "" {
		label := arrow.Label
		labelX := xStart + (xEnd-xStart+1-runeLen(label))/2
		labelY := y - 1
		if labelY < 0 {
			labelY = y + 1
		}
		grid.writeText(labelX, labelY, label, roleArrowLabel)
	}
}

func (d *Diagram) drawVerticalArrow(grid *runeGrid, from, to placedBox, arrow DiagramArrow) {
	top := from
	bottom := to
	if from.y > to.y {
		top = to
		bottom = from
	}

	x := from.x + from.width/2
	yStart := top.y + top.height
	yEnd := bottom.y - 1
	if yEnd <= yStart {
		return
	}

	lineRune := verticalLineRune(arrow.Style)
	for y := yStart; y <= yEnd; y++ {
		grid.set(x, y, lineRune, roleArrowLine)
	}

	switch arrow.Direction {
	case ArrowBidirectional:
		grid.set(x, yStart, '▲', roleArrowLine)
		grid.set(x, yEnd, '▼', roleArrowLine)
	case ArrowBottomToTop:
		grid.set(x, yStart, '▲', roleArrowLine)
	default:
		grid.set(x, yEnd, '▼', roleArrowLine)
	}

	if strings.TrimSpace(arrow.Label) != "" {
		yMid := yStart + (yEnd-yStart)/2
		grid.writeText(x+2, yMid, arrow.Label, roleArrowLabel)
	}
}

func horizontalLineRune(style ArrowStyle) rune {
	switch style {
	case ArrowDashed:
		return '╌'
	case ArrowDouble:
		return '═'
	default:
		return '─'
	}
}

func verticalLineRune(style ArrowStyle) rune {
	switch style {
	case ArrowDashed:
		return '┆'
	case ArrowDouble:
		return '║'
	default:
		return '│'
	}
}

type cellRole int

const (
	roleNone cellRole = iota
	roleBoxBorder
	roleBoxTitle
	roleBoxBody
	roleArrowLine
	roleArrowLabel
)

type gridCell struct {
	ch   rune
	role cellRole
}

type runeGrid struct {
	cells  [][]gridCell
	width  int
	height int
}

func newRuneGrid(width, height int) *runeGrid {
	g := &runeGrid{
		width:  maxInt(width, 1),
		height: maxInt(height, 1),
	}
	g.cells = make([][]gridCell, g.height)
	for y := 0; y < g.height; y++ {
		g.cells[y] = make([]gridCell, g.width)
		for x := 0; x < g.width; x++ {
			g.cells[y][x] = gridCell{ch: ' ', role: roleNone}
		}
	}
	return g
}

func (g *runeGrid) ensure(x, y int) {
	if x < 0 || y < 0 {
		return
	}
	if y >= g.height {
		oldHeight := g.height
		g.height = y + 1
		rows := make([][]gridCell, g.height)
		copy(rows, g.cells)
		for i := oldHeight; i < g.height; i++ {
			rows[i] = make([]gridCell, g.width)
			for x := 0; x < g.width; x++ {
				rows[i][x] = gridCell{ch: ' ', role: roleNone}
			}
		}
		g.cells = rows
	}
	if x >= g.width {
		oldWidth := g.width
		g.width = x + 1
		for i := 0; i < g.height; i++ {
			row := make([]gridCell, g.width)
			copy(row, g.cells[i])
			for j := oldWidth; j < g.width; j++ {
				row[j] = gridCell{ch: ' ', role: roleNone}
			}
			g.cells[i] = row
		}
	}
}

func (g *runeGrid) set(x, y int, ch rune, role cellRole) {
	if x < 0 || y < 0 {
		return
	}
	g.ensure(x, y)
	g.cells[y][x] = gridCell{ch: ch, role: role}
}

func (g *runeGrid) writeText(x, y int, text string, role cellRole) {
	if y < 0 {
		return
	}
	rx := x
	for _, r := range text {
		if rx >= 0 {
			g.set(rx, y, r, role)
		}
		rx++
	}
}

func (d *Diagram) gridToLines(grid *runeGrid) []string {
	lines := make([]string, 0, grid.height)
	styles := d.styles()

	for y := 0; y < grid.height; y++ {
		row := grid.cells[y]
		last := -1
		for x := len(row) - 1; x >= 0; x-- {
			if row[x].ch != ' ' {
				last = x
				break
			}
		}
		if last < 0 {
			lines = append(lines, "")
			continue
		}

		var b strings.Builder
		for x := 0; x <= last; x++ {
			cell := row[x]
			if cell.ch == 0 {
				cell.ch = ' '
			}
			if styles == nil || cell.role == roleNone {
				b.WriteRune(cell.ch)
				continue
			}
			st := styles[cell.role]
			b.WriteString(st.Render(string(cell.ch)))
		}
		lines = append(lines, b.String())
	}

	return lines
}

func (d *Diagram) renderTitleLine(title string, width int) string {
	clean := strings.TrimSpace(title)
	if clean == "" {
		return ""
	}
	line := centerText(clean, maxInt(width, runeLen(clean)))
	if d.tokens == nil {
		return strings.TrimRight(line, " ")
	}
	st := lipgloss.NewStyle().Bold(true)
	if c := tokenColor(d.tokens, "Primary", "Accent", "Color"); c != "" {
		st = st.Foreground(lipgloss.Color(c))
	}
	return strings.TrimRight(st.Render(line), " ")
}

func (d *Diagram) styles() map[cellRole]lipgloss.Style {
	if d.tokens == nil {
		return nil
	}

	border := lipgloss.NewStyle()
	if c := tokenColor(d.tokens, "Surface", "Border", "MutedColor", "Color"); c != "" {
		border = border.Foreground(lipgloss.Color(c))
	}

	title := lipgloss.NewStyle().Bold(true)
	if c := tokenColor(d.tokens, "Primary", "Accent", "Color"); c != "" {
		title = title.Foreground(lipgloss.Color(c))
	}

	body := lipgloss.NewStyle()
	if c := tokenColor(d.tokens, "TextMuted", "MutedColor", "Color"); c != "" {
		body = body.Foreground(lipgloss.Color(c))
	}

	arrow := lipgloss.NewStyle()
	if c := tokenColor(d.tokens, "TextDim", "MutedColor", "Color"); c != "" {
		arrow = arrow.Foreground(lipgloss.Color(c))
	}

	arrowLabel := lipgloss.NewStyle()
	if c := tokenColor(d.tokens, "TextSecondary", "Color", "MutedColor"); c != "" {
		arrowLabel = arrowLabel.Foreground(lipgloss.Color(c))
	}

	return map[cellRole]lipgloss.Style{
		roleBoxBorder: border,
		roleBoxTitle:  title,
		roleBoxBody:   body,
		roleArrowLine: arrow,
		roleArrowLabel: arrowLabel,
	}
}

func tokenColor(tokens *design.DesignTokens, fieldNames ...string) string {
	if tokens == nil {
		return ""
	}
	v := reflect.ValueOf(tokens)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return ""
	}
	for _, name := range fieldNames {
		f := v.FieldByName(name)
		if !f.IsValid() || f.Kind() != reflect.String {
			continue
		}
		color := strings.TrimSpace(f.String())
		if color != "" {
			return color
		}
	}
	return ""
}

func padOrTrim(s string, width int) string {
	r := []rune(s)
	if width <= 0 {
		return ""
	}
	if len(r) > width {
		return string(r[:width])
	}
	if len(r) < width {
		return string(r) + strings.Repeat(" ", width-len(r))
	}
	return s
}

func centerText(text string, width int) string {
	contentW := runeLen(text)
	if width <= contentW {
		return text
	}
	left := (width - contentW) / 2
	right := width - contentW - left
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}

