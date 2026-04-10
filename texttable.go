package tui

import (
	"sort"
	"strings"
	"unicode/utf8"

	design "github.com/SCKelemen/design-system"
	tea "github.com/charmbracelet/bubbletea"
)

// TextTableAlignment controls how content is padded within a column.
type TextTableAlignment int

const (
	TextTableAlignLeft TextTableAlignment = iota
	TextTableAlignCenter
	TextTableAlignRight
)

// TextTableBorderStyle controls how the table frame renders.
type TextTableBorderStyle int

const (
	TextTableBorderNone TextTableBorderStyle = iota
	TextTableBorderSingle
	TextTableBorderDouble
	TextTableBorderRounded
)

// Column defines the metadata for a table column.
type Column struct {
	Name      string
	Width     int
	Alignment TextTableAlignment
	MinWidth  int
	MaxWidth  int
}

// TextTable renders tabular text with optional borders, sorting, and selection.
type TextTable struct {
	columns []Column
	rows    [][]string

	selectedRow   int
	zebraStriped  bool
	borderStyle   TextTableBorderStyle
	sortColumn    int
	sortAscending bool

	x int
	y int

	focused bool

	designTokens *design.DesignTokens
	headerStyle  string
	textStyle    string
	borderColor  string
	zebraStyle   string
	selectStyle  string
}

// TextTableOption configures a TextTable.
type TextTableOption func(*TextTable)

// WithTextTableRows sets the initial rows.
func WithTextTableRows(rows [][]string) TextTableOption {
	return func(tt *TextTable) {
		tt.SetRows(rows)
	}
}

// WithTextTableBorderStyle sets the table border style.
func WithTextTableBorderStyle(style TextTableBorderStyle) TextTableOption {
	return func(tt *TextTable) {
		tt.borderStyle = style
	}
}

// WithTextTableZebraStriping toggles zebra-striped row backgrounds.
func WithTextTableZebraStriping(enabled bool) TextTableOption {
	return func(tt *TextTable) {
		tt.zebraStriped = enabled
	}
}

// WithTextTableDesignTokens applies design-system colors.
func WithTextTableDesignTokens(tokens *design.DesignTokens) TextTableOption {
	return func(tt *TextTable) {
		tt.applyDesignTokens(tokens)
	}
}

// WithTextTableTheme applies a named design-system theme.
func WithTextTableTheme(theme string) TextTableOption {
	return func(tt *TextTable) {
		tt.applyDesignTokens(designTokensForTheme(theme))
	}
}

// NewTextTable creates a richer table renderable for TUI layouts.
func NewTextTable(columns []Column, opts ...TextTableOption) *TextTable {
	tt := &TextTable{
		columns:       cloneColumns(columns),
		rows:          [][]string{},
		selectedRow:   -1,
		borderStyle:   TextTableBorderRounded,
		sortColumn:    -1,
		sortAscending: true,
		headerStyle:   ansiBold,
		textStyle:     "",
		borderColor:   ansiDim,
		zebraStyle:    ansiDim,
		selectStyle:   ansiInverse,
		designTokens:  design.DefaultTheme(),
	}

	tt.applyDesignTokens(tt.designTokens)
	for _, opt := range opts {
		opt(tt)
	}
	return tt
}

// Init initializes the text table.
func (tt *TextTable) Init() tea.Cmd {
	return nil
}

// Update handles keyboard and mouse interactions for selection and header sorting.
func (tt *TextTable) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !tt.focused || len(tt.rows) == 0 {
			return tt, nil
		}
		switch msg.String() {
		case "up", "k":
			tt.moveSelection(-1)
		case "down", "j":
			tt.moveSelection(1)
		case "home", "g":
			tt.selectedRow = 0
		case "end", "G":
			tt.selectedRow = len(tt.rows) - 1
		}
	case tea.MouseMsg:
		tt.handleMouse(msg)
	}
	return tt, nil
}

// View renders the table.
func (tt *TextTable) View() string {
	if len(tt.columns) == 0 {
		return ""
	}

	widths := tt.columnWidths()
	border := tt.borderCharacters()
	lines := make([]string, 0, len(tt.rows)+4)

	if tt.borderStyle != TextTableBorderNone {
		lines = append(lines, tt.styleLine(tt.renderBorder(widths, border.topLeft, border.topT, border.topRight, border.horizontal), tt.borderColor))
	}

	lines = append(lines, tt.styleLine(tt.renderRow(tt.headerCells(), widths, true, false), tt.headerStyle))
	lines = append(lines, tt.styleLine(tt.renderBorder(widths, border.leftT, border.cross, border.rightT, border.horizontal), tt.borderColor))

	for i, row := range tt.rows {
		line := tt.renderRow(row, widths, false, i == tt.selectedRow)
		style := tt.textStyle
		if i == tt.selectedRow {
			style = tt.selectStyle
		} else if tt.zebraStriped && i%2 == 1 {
			style = tt.zebraStyle
		}
		lines = append(lines, tt.styleLine(line, style))
	}

	if tt.borderStyle != TextTableBorderNone {
		lines = append(lines, tt.styleLine(tt.renderBorder(widths, border.bottomLeft, border.bottomT, border.bottomRight, border.horizontal), tt.borderColor))
	}

	return strings.Join(lines, "\n")
}

// Focus marks the table as focused.
func (tt *TextTable) Focus() {
	tt.focused = true
}

// Blur marks the table as unfocused.
func (tt *TextTable) Blur() {
	tt.focused = false
}

// Focused reports whether the table is focused.
func (tt *TextTable) Focused() bool {
	return tt.focused
}

// SetColumns replaces the column definitions.
func (tt *TextTable) SetColumns(columns []Column) {
	tt.columns = cloneColumns(columns)
	if tt.sortColumn >= len(tt.columns) {
		tt.sortColumn = -1
	}
}

// SetRows replaces the table rows.
func (tt *TextTable) SetRows(rows [][]string) {
	tt.rows = cloneRows(rows)
	if len(tt.rows) == 0 {
		tt.selectedRow = -1
		return
	}
	if tt.selectedRow >= len(tt.rows) {
		tt.selectedRow = len(tt.rows) - 1
	}
}

// AddRow appends a row to the table.
func (tt *TextTable) AddRow(row ...string) {
	tt.rows = append(tt.rows, append([]string(nil), row...))
	if tt.selectedRow == -1 {
		tt.selectedRow = 0
	}
}

// SetSelectedRow updates the selected row index.
func (tt *TextTable) SetSelectedRow(index int) {
	if len(tt.rows) == 0 {
		tt.selectedRow = -1
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(tt.rows) {
		index = len(tt.rows) - 1
	}
	tt.selectedRow = index
}

// SetPosition sets the top-left screen coordinate used for mouse hit testing.
func (tt *TextTable) SetPosition(x, y int) {
	tt.x = x
	tt.y = y
}

// SetBorderStyle updates the border style.
func (tt *TextTable) SetBorderStyle(style TextTableBorderStyle) {
	tt.borderStyle = style
}

// SetZebraStriping toggles zebra striping.
func (tt *TextTable) SetZebraStriping(enabled bool) {
	tt.zebraStriped = enabled
}

// SortByColumn sorts rows by the selected column, toggling direction on repeated calls.
func (tt *TextTable) SortByColumn(index int) {
	if index < 0 || index >= len(tt.columns) {
		return
	}
	if tt.sortColumn == index {
		tt.sortAscending = !tt.sortAscending
	} else {
		tt.sortColumn = index
		tt.sortAscending = true
	}

	sort.SliceStable(tt.rows, func(i, j int) bool {
		left := strings.ToLower(tt.cellValue(tt.rows[i], index))
		right := strings.ToLower(tt.cellValue(tt.rows[j], index))
		if tt.sortAscending {
			return left < right
		}
		return left > right
	})

	if len(tt.rows) == 0 {
		tt.selectedRow = -1
		return
	}
	if tt.selectedRow < 0 {
		tt.selectedRow = 0
	}
	if tt.selectedRow >= len(tt.rows) {
		tt.selectedRow = len(tt.rows) - 1
	}
}

func (tt *TextTable) handleMouse(msg tea.MouseMsg) {
	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionRelease {
		return
	}
	widths := tt.columnWidths()
	if !tt.contains(msg.X, msg.Y, widths) {
		return
	}

	headerRowY := tt.headerRowY()
	if msg.Y == headerRowY {
		if col, ok := tt.columnAtX(msg.X, widths); ok {
			tt.SortByColumn(col)
		}
		return
	}

	rowIndex := tt.rowAtY(msg.Y)
	if rowIndex >= 0 && rowIndex < len(tt.rows) {
		tt.selectedRow = rowIndex
	}
}

func (tt *TextTable) moveSelection(delta int) {
	if len(tt.rows) == 0 {
		tt.selectedRow = -1
		return
	}
	if tt.selectedRow == -1 {
		tt.selectedRow = 0
		return
	}
	tt.selectedRow += delta
	if tt.selectedRow < 0 {
		tt.selectedRow = 0
	}
	if tt.selectedRow >= len(tt.rows) {
		tt.selectedRow = len(tt.rows) - 1
	}
}

func (tt *TextTable) columnWidths() []int {
	widths := make([]int, len(tt.columns))
	for i, column := range tt.columns {
		width := column.Width
		if width <= 0 {
			width = max(textWidth(tt.headerCell(i)), textWidth(column.Name))
			for _, row := range tt.rows {
				width = max(width, textWidth(tt.cellValue(row, i)))
			}
		}
		if column.MinWidth > 0 && width < column.MinWidth {
			width = column.MinWidth
		}
		if column.MaxWidth > 0 && width > column.MaxWidth {
			width = column.MaxWidth
		}
		if width < 1 {
			width = 1
		}
		widths[i] = width
	}
	return widths
}

func (tt *TextTable) headerCells() []string {
	cells := make([]string, len(tt.columns))
	for i := range tt.columns {
		cells[i] = tt.headerCell(i)
	}
	return cells
}

func (tt *TextTable) headerCell(index int) string {
	name := tt.columns[index].Name
	if tt.sortColumn != index {
		return name
	}
	if tt.sortAscending {
		return name + " ↑"
	}
	return name + " ↓"
}

func (tt *TextTable) renderRow(cells []string, widths []int, header, selected bool) string {
	if tt.borderStyle == TextTableBorderNone {
		parts := make([]string, len(widths))
		for i, width := range widths {
			parts[i] = tt.formatCell(tt.cellValue(cells, i), width, tt.columns[i].Alignment)
		}
		return strings.Join(parts, "  ")
	}

	border := tt.borderCharacters()
	parts := make([]string, len(widths))
	for i, width := range widths {
		parts[i] = " " + tt.formatCell(tt.cellValue(cells, i), width, tt.columns[i].Alignment) + " "
	}
	return border.vertical + strings.Join(parts, border.vertical) + border.vertical
}

func (tt *TextTable) renderBorder(widths []int, left, middle, right, horizontal string) string {
	if tt.borderStyle == TextTableBorderNone {
		parts := make([]string, len(widths))
		for i, width := range widths {
			parts[i] = strings.Repeat("-", width)
		}
		return strings.Join(parts, "  ")
	}

	parts := make([]string, len(widths))
	for i, width := range widths {
		parts[i] = strings.Repeat(horizontal, width+2)
	}
	return left + strings.Join(parts, middle) + right
}

func (tt *TextTable) formatCell(value string, width int, alignment TextTableAlignment) string {
	text := tt.truncateWithEllipsis(value, width)
	padding := width - textWidth(text)
	if padding <= 0 {
		return text
	}

	switch alignment {
	case TextTableAlignRight:
		return strings.Repeat(" ", padding) + text
	case TextTableAlignCenter:
		left := padding / 2
		right := padding - left
		return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
	default:
		return text + strings.Repeat(" ", padding)
	}
}

func (tt *TextTable) truncateWithEllipsis(value string, width int) string {
	plain := stripANSI(value)
	if width <= 0 {
		return ""
	}
	if textWidth(plain) <= width {
		return plain
	}
	if width == 1 {
		return "…"
	}

	runes := []rune(plain)
	if len(runes) > width-1 {
		runes = runes[:width-1]
	}
	return string(runes) + "…"
}

func (tt *TextTable) contains(x, y int, widths []int) bool {
	if y < tt.y || y >= tt.y+tt.totalHeight() {
		return false
	}
	return x >= tt.x && x < tt.x+tt.renderWidth(widths)
}

func (tt *TextTable) renderWidth(widths []int) int {
	line := tt.renderRow(tt.headerCells(), widths, true, false)
	return textWidth(line)
}

func (tt *TextTable) totalHeight() int {
	height := len(tt.rows) + 2
	if tt.borderStyle != TextTableBorderNone {
		height += 2
	}
	return height
}

func (tt *TextTable) headerRowY() int {
	if tt.borderStyle == TextTableBorderNone {
		return tt.y
	}
	return tt.y + 1
}

func (tt *TextTable) rowAtY(y int) int {
	start := tt.y + 2
	if tt.borderStyle != TextTableBorderNone {
		start = tt.y + 3
	}
	return y - start
}

func (tt *TextTable) columnAtX(x int, widths []int) (int, bool) {
	cursor := tt.x
	if tt.borderStyle != TextTableBorderNone {
		cursor++
		for i, width := range widths {
			span := width + 2
			if x >= cursor && x < cursor+span {
				return i, true
			}
			cursor += span + 1
		}
		return 0, false
	}

	for i, width := range widths {
		if x >= cursor && x < cursor+width {
			return i, true
		}
		cursor += width
		if i < len(widths)-1 {
			cursor += 2
		}
	}
	return 0, false
}

func (tt *TextTable) cellValue(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return row[index]
}

func (tt *TextTable) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}
	tt.designTokens = tokens
	if accent := ansiColorFromHex(tokens.Accent); accent != "" {
		tt.headerStyle = ansiBold + accent
	}
	if foreground := ansiColorFromHex(tokens.Color); foreground != "" {
		tt.textStyle = foreground
		tt.borderColor = foreground
		tt.zebraStyle = ansiDim + foreground
	}
	selectionStyle := ""
	if bg := ansiBackgroundColorFromHex(tokens.Accent); bg != "" {
		selectionStyle += bg
	}
	if fg := ansiColorFromHex(tokens.Background); fg != "" {
		selectionStyle += fg
	} else if fg := ansiColorFromHex(tokens.Color); fg != "" {
		selectionStyle += fg
	}
	if selectionStyle != "" {
		tt.selectStyle = selectionStyle
	}
}
func (tt *TextTable) styleLine(line, style string) string {
	if style == "" {
		return line
	}
	return style + line + ansiReset
}

func cloneColumns(columns []Column) []Column {
	if len(columns) == 0 {
		return nil
	}
	cloned := make([]Column, len(columns))
	copy(cloned, columns)
	return cloned
}

func cloneRows(rows [][]string) [][]string {
	if len(rows) == 0 {
		return [][]string{}
	}
	cloned := make([][]string, len(rows))
	for i, row := range rows {
		cloned[i] = append([]string(nil), row...)
	}
	return cloned
}

type textTableBorders struct {
	topLeft     string
	topRight    string
	bottomLeft  string
	bottomRight string
	horizontal  string
	vertical    string
	cross       string
	leftT       string
	rightT      string
	topT        string
	bottomT     string
}

func (tt *TextTable) borderCharacters() textTableBorders {
	switch tt.borderStyle {
	case TextTableBorderSingle:
		return textTableBorders{
			topLeft:     "┌",
			topRight:    "┐",
			bottomLeft:  "└",
			bottomRight: "┘",
			horizontal:  "─",
			vertical:    "│",
			cross:       "┼",
			leftT:       "├",
			rightT:      "┤",
			topT:        "┬",
			bottomT:     "┴",
		}
	case TextTableBorderDouble:
		return textTableBorders{
			topLeft:     "╔",
			topRight:    "╗",
			bottomLeft:  "╚",
			bottomRight: "╝",
			horizontal:  "═",
			vertical:    "║",
			cross:       "╬",
			leftT:       "╠",
			rightT:      "╣",
			topT:        "╦",
			bottomT:     "╩",
		}
	case TextTableBorderRounded:
		return textTableBorders{
			topLeft:     "╭",
			topRight:    "╮",
			bottomLeft:  "╰",
			bottomRight: "╯",
			horizontal:  "─",
			vertical:    "│",
			cross:       "┼",
			leftT:       "├",
			rightT:      "┤",
			topT:        "┬",
			bottomT:     "┴",
		}
	default:
		return textTableBorders{}
	}
}

func textWidth(value string) int {
	return utf8.RuneCountInString(stripANSI(value))
}
