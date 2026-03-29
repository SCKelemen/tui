package display

import (
	"fmt"
	"strings"

	tui "github.com/SCKelemen/tui/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// changeKind describes the type of change represented by a row.
type changeKind int

const (
	changeEqual changeKind = iota
	changeAdded
	changeRemoved
	changeModified
)

// lineChange maps one rendered row to optional old/new line numbers.
type lineChange struct {
	kind            changeKind
	oldLine, newLine int
}

// SideBySideDiff renders a VS Code-style side-by-side diff view.
type SideBySideDiff struct {
	oldCode, newCode string
	oldLabel, newLabel string
	language string
	width int
	lineChanges []lineChange
	scrollOffset int
	visibleLines int
	focused bool
	showLineNumbers bool

	oldLines []string
	newLines []string
	highlighter *Highlighter
}

// SideBySideDiffOption configures SideBySideDiff.
type SideBySideDiffOption func(*SideBySideDiff)

// WithSideBySideDiffLanguage sets the syntax language used for highlighting.
func WithSideBySideDiffLanguage(lang string) SideBySideDiffOption {
	return func(d *SideBySideDiff) {
		d.language = strings.TrimSpace(lang)
	}
}

// WithSideBySideDiffLabels sets labels for the old and new panes.
func WithSideBySideDiffLabels(old, new string) SideBySideDiffOption {
	return func(d *SideBySideDiff) {
		d.oldLabel = old
		d.newLabel = new
	}
}

// WithSideBySideDiffWidth sets the render width.
func WithSideBySideDiffWidth(width int) SideBySideDiffOption {
	return func(d *SideBySideDiff) {
		if width > 0 {
			d.width = width
		}
	}
}

// WithSideBySideDiffVisibleLines sets the number of visible rows before scrolling.
func WithSideBySideDiffVisibleLines(lines int) SideBySideDiffOption {
	return func(d *SideBySideDiff) {
		if lines > 0 {
			d.visibleLines = lines
		}
	}
}

// NewSideBySideDiff creates a side-by-side diff component using an LCS-based line diff.
func NewSideBySideDiff(oldCode, newCode string, opts ...SideBySideDiffOption) *SideBySideDiff {
	d := &SideBySideDiff{
		oldCode: oldCode,
		newCode: newCode,
		oldLabel: "Old",
		newLabel: "New",
		width: 100,
		visibleLines: 20,
		showLineNumbers: true,
	}

	for _, opt := range opts {
		opt(d)
	}

	d.oldLines = splitCodeLines(oldCode)
	d.newLines = splitCodeLines(newCode)
	d.lineChanges = computeLineChanges(d.oldLines, d.newLines)

	if d.language == "" {
		lang := DetectLanguage(d.newLabel)
		if lang == SyntaxLanguagePlain {
			lang = DetectLanguage(d.oldLabel)
		}
		d.language = string(lang)
	}

	parsed := parseSyntaxLanguage(d.language)
	if parsed != SyntaxLanguagePlain {
		d.highlighter = NewHighlighter(parsed)
	}

	return d
}

// Init initializes the component.
func (d *SideBySideDiff) Init() tea.Cmd { return nil }

// Update handles window updates and keyboard scrolling when focused.
func (d *SideBySideDiff) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			d.width = msg.Width
		}
		d.clampScroll()
	case tea.KeyMsg:
		if !d.focused {
			return d, nil
		}

		jump := d.visibleLines - 1
		if jump < 1 {
			jump = 1
		}

		switch msg.String() {
		case "up":
			d.scrollOffset--
		case "down":
			d.scrollOffset++
		case "pgup", "pageup":
			d.scrollOffset -= jump
		case "pgdown", "pagedown":
			d.scrollOffset += jump
		case "home":
			d.scrollOffset = 0
		case "end":
			d.scrollOffset = d.maxScrollOffset()
		}
		d.clampScroll()
	}

	return d, nil
}

// View renders the full side-by-side diff.
func (d *SideBySideDiff) View() string {
	if len(d.lineChanges) == 0 {
		return d.renderHeader() + "\n"
	}

	viewWidth := d.width
	if viewWidth <= 0 {
		viewWidth = 100
	}

	separatorWidth := 7 // " │ x │ "
	colWidth := (viewWidth - separatorWidth) / 2
	if colWidth < 10 {
		colWidth = 10
	}

	oldNumWidth := d.lineNumberWidth(true)
	newNumWidth := d.lineNumberWidth(false)
	if !d.showLineNumbers {
		oldNumWidth = 0
		newNumWidth = 0
	}

	oldContentWidth := colWidth
	if d.showLineNumbers {
		oldContentWidth -= oldNumWidth + 1
	}
	if oldContentWidth < 1 {
		oldContentWidth = 1
	}

	newContentWidth := colWidth
	if d.showLineNumbers {
		newContentWidth -= newNumWidth + 1
	}
	if newContentWidth < 1 {
		newContentWidth = 1
	}

	start := d.scrollOffset
	end := start + d.visibleLines
	if end > len(d.lineChanges) {
		end = len(d.lineChanges)
	}

	var b strings.Builder
	b.WriteString(d.renderHeader())
	b.WriteString("\n")

	for i := start; i < end; i++ {
		change := d.lineChanges[i]
		oldText := d.contentAt(true, change.oldLine)
		newText := d.contentAt(false, change.newLine)
		if d.highlighter != nil {
			oldText = d.highlighter.HighlightLine(oldText)
			newText = d.highlighter.HighlightLine(newText)
		}

		left := d.renderPaneLine(true, change, oldText, colWidth, oldNumWidth, oldContentWidth)
		right := d.renderPaneLine(false, change, newText, colWidth, newNumWidth, newContentWidth)
		middle := d.renderCenter(change.kind)
		b.WriteString(left + middle + right + "\n")
	}

	if len(d.lineChanges) > d.visibleLines {
		b.WriteString(d.renderScrollIndicator())
	}

	return b.String()
}

// Focus marks this component as focused.
func (d *SideBySideDiff) Focus() { d.focused = true }

// Blur marks this component as unfocused.
func (d *SideBySideDiff) Blur() { d.focused = false }

// Focused reports focus state.
func (d *SideBySideDiff) Focused() bool { return d.focused }

func (d *SideBySideDiff) renderHeader() string {
	viewWidth := d.width
	if viewWidth <= 0 {
		viewWidth = 100
	}

	separatorWidth := 7
	colWidth := (viewWidth - separatorWidth) / 2
	if colWidth < 10 {
		colWidth = 10
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#2A2D34")).Foreground(lipgloss.Color("#E6EAF0"))
	dividerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4D5566")).Background(lipgloss.Color("#2A2D34"))
	gutterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9AA4B2")).Background(lipgloss.Color("#2A2D34"))

	left := headerStyle.Width(colWidth).MaxWidth(colWidth).Render(" " + d.oldLabel)
	right := headerStyle.Width(colWidth).MaxWidth(colWidth).Render(" " + d.newLabel)
	middle := dividerStyle.Render(" ") + dividerStyle.Render("│") + gutterStyle.Render(" · ") + dividerStyle.Render("│") + dividerStyle.Render(" ")

	return left + middle + right
}

func (d *SideBySideDiff) renderPaneLine(isOld bool, change lineChange, content string, colWidth, numWidth, contentWidth int) string {
	numStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7F8795"))
	baseStyle := lipgloss.NewStyle().Width(colWidth).MaxWidth(colWidth)

	var lineNum int
	if isOld {
		lineNum = change.oldLine
	} else {
		lineNum = change.newLine
	}

	numPart := ""
	if d.showLineNumbers {
		if lineNum > 0 {
			numPart = numStyle.Render(fmt.Sprintf("%*d", numWidth, lineNum)) + " "
		} else {
			numPart = strings.Repeat(" ", numWidth+1)
		}
	}

	if content != "" {
		content = lipgloss.NewStyle().MaxWidth(contentWidth).Render(content)
	}

	line := numPart + content
	if lipgloss.Width(line) < colWidth {
		line += strings.Repeat(" ", colWidth-lipgloss.Width(line))
	}
	if lipgloss.Width(line) > colWidth {
		line = lipgloss.NewStyle().Width(colWidth).MaxWidth(colWidth).Render(line)
	}

	highlightStyle := d.highlightStyle(isOld, change.kind)
	if highlightStyle != nil {
		return highlightStyle.Width(colWidth).MaxWidth(colWidth).Render(line)
	}

	return baseStyle.Render(line)
}

func (d *SideBySideDiff) highlightStyle(isOld bool, kind changeKind) *lipgloss.Style {
	switch kind {
	case changeAdded:
		if !isOld {
			st := lipgloss.NewStyle().Background(lipgloss.Color("#1F3A2A"))
			return &st
		}
	case changeRemoved:
		if isOld {
			st := lipgloss.NewStyle().Background(lipgloss.Color("#4A2328"))
			return &st
		}
	case changeModified:
		st := lipgloss.NewStyle().Background(lipgloss.Color("#4A3F21"))
		return &st
	}
	return nil
}

func (d *SideBySideDiff) renderCenter(kind changeKind) string {
	dividerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4D5566"))
	gutterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9AA4B2"))

	indicator := " "
	switch kind {
	case changeAdded:
		indicator = lipgloss.NewStyle().Foreground(lipgloss.Color("#6BCB77")).Render("+")
	case changeRemoved:
		indicator = lipgloss.NewStyle().Foreground(lipgloss.Color("#E57373")).Render("-")
	case changeModified:
		indicator = lipgloss.NewStyle().Foreground(lipgloss.Color("#E7B567")).Render("~")
	}

	return dividerStyle.Render(" ") + dividerStyle.Render("│") + gutterStyle.Render(" ") + indicator + gutterStyle.Render(" ") + dividerStyle.Render("│") + dividerStyle.Render(" ")
}

func (d *SideBySideDiff) renderScrollIndicator() string {
	total := len(d.lineChanges)
	start := d.scrollOffset + 1
	end := d.scrollOffset + d.visibleLines
	if end > total {
		end = total
	}

	barWidth := 14
	thumbStart := 0
	thumbSize := barWidth
	if total > 0 && d.visibleLines > 0 {
		thumbSize = (d.visibleLines * barWidth) / total
		if thumbSize < 1 {
			thumbSize = 1
		}
		maxStart := barWidth - thumbSize
		if maxStart < 0 {
			maxStart = 0
		}
		if total > d.visibleLines {
			thumbStart = (d.scrollOffset * maxStart) / (total - d.visibleLines)
		}
	}

	track := strings.Repeat("─", barWidth)
	thumb := strings.Repeat("█", thumbSize)
	if thumbStart >= 0 && thumbStart+thumbSize <= len(track) {
		track = track[:thumbStart] + thumb + track[thumbStart+thumbSize:]
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#7F8795"))
	return style.Render(fmt.Sprintf("%s  lines %d-%d/%d", track, start, end, total))
}

func (d *SideBySideDiff) contentAt(isOld bool, line int) string {
	if line <= 0 {
		return ""
	}
	if isOld {
		if line > len(d.oldLines) {
			return ""
		}
		return d.oldLines[line-1]
	}
	if line > len(d.newLines) {
		return ""
	}
	return d.newLines[line-1]
}

func (d *SideBySideDiff) lineNumberWidth(isOld bool) int {
	if isOld {
		if len(d.oldLines) == 0 {
			return 1
		}
		return len(fmt.Sprintf("%d", len(d.oldLines)))
	}
	if len(d.newLines) == 0 {
		return 1
	}
	return len(fmt.Sprintf("%d", len(d.newLines)))
}

func (d *SideBySideDiff) maxScrollOffset() int {
	if len(d.lineChanges) <= d.visibleLines {
		return 0
	}
	return len(d.lineChanges) - d.visibleLines
}

func (d *SideBySideDiff) clampScroll() {
	if d.visibleLines <= 0 {
		d.visibleLines = 20
	}
	if d.scrollOffset < 0 {
		d.scrollOffset = 0
	}
	maxOffset := d.maxScrollOffset()
	if d.scrollOffset > maxOffset {
		d.scrollOffset = maxOffset
	}
}

type diffOpKind int

const (
	diffEqual diffOpKind = iota
	diffAdd
	diffRemove
)

type diffOp struct {
	kind diffOpKind
	oldLine int
	newLine int
}

func computeLineChanges(oldLines, newLines []string) []lineChange {
	ops := lcsDiffOps(oldLines, newLines)
	changes := make([]lineChange, 0, len(ops))

	for i := 0; i < len(ops); {
		if ops[i].kind == diffEqual {
			changes = append(changes, lineChange{kind: changeEqual, oldLine: ops[i].oldLine, newLine: ops[i].newLine})
			i++
			continue
		}

		if ops[i].kind == diffRemove || ops[i].kind == diffAdd {
			start := i
			for i < len(ops) && ops[i].kind != diffEqual {
				i++
			}
			changes = append(changes, collapseChangeRun(ops[start:i])...)
			continue
		}

		i++
	}

	return changes
}

func collapseChangeRun(run []diffOp) []lineChange {
	removed := make([]int, 0)
	added := make([]int, 0)

	for _, op := range run {
		switch op.kind {
		case diffRemove:
			removed = append(removed, op.oldLine)
		case diffAdd:
			added = append(added, op.newLine)
		}
	}

	result := make([]lineChange, 0, len(run))
	paired := len(removed)
	if len(added) < paired {
		paired = len(added)
	}

	for i := 0; i < paired; i++ {
		result = append(result, lineChange{kind: changeModified, oldLine: removed[i], newLine: added[i]})
	}
	for i := paired; i < len(removed); i++ {
		result = append(result, lineChange{kind: changeRemoved, oldLine: removed[i]})
	}
	for i := paired; i < len(added); i++ {
		result = append(result, lineChange{kind: changeAdded, newLine: added[i]})
	}

	return result
}

func lcsDiffOps(oldLines, newLines []string) []diffOp {
	n := len(oldLines)
	m := len(newLines)

	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}

	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if oldLines[i] == newLines[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	i, j := 0, 0
	ops := make([]diffOp, 0, n+m)
	for i < n && j < m {
		switch {
		case oldLines[i] == newLines[j]:
			ops = append(ops, diffOp{kind: diffEqual, oldLine: i + 1, newLine: j + 1})
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			ops = append(ops, diffOp{kind: diffRemove, oldLine: i + 1})
			i++
		default:
			ops = append(ops, diffOp{kind: diffAdd, newLine: j + 1})
			j++
		}
	}

	for i < n {
		ops = append(ops, diffOp{kind: diffRemove, oldLine: i + 1})
		i++
	}
	for j < m {
		ops = append(ops, diffOp{kind: diffAdd, newLine: j + 1})
		j++
	}

	return ops
}

func splitCodeLines(code string) []string {
	if code == "" {
		return []string{}
	}
	return strings.Split(code, "\n")
}

var _ tui.Component = (*SideBySideDiff)(nil)
