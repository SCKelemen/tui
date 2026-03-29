package display

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// GutterMark identifies the type of decoration shown in the gutter.
type GutterMark int

const (
	GutterError GutterMark = iota
	GutterWarning
	GutterInfo
	GutterHint
	GutterBreakpoint
	GutterBookmark
	GutterGitAdded
	GutterGitModified
	GutterGitDeleted
	GutterCurrentLine
	GutterFoldOpen
	GutterFoldClosed
)

// GutterDecoration describes a decoration for a single line in the gutter.
type GutterDecoration struct {
	Line    int
	Mark    GutterMark
	Tooltip string
}

// MarkIcon returns the icon used for a gutter mark.
func MarkIcon(m GutterMark) string {
	switch m {
	case GutterError:
		return "●"
	case GutterWarning:
		return "▲"
	case GutterInfo:
		return "◆"
	case GutterHint:
		return "◇"
	case GutterBreakpoint:
		return "⏺"
	case GutterBookmark:
		return "★"
	case GutterGitAdded, GutterGitModified, GutterGitDeleted:
		return "┃"
	case GutterCurrentLine:
		return "▶"
	case GutterFoldOpen:
		return "▾"
	case GutterFoldClosed:
		return "▸"
	default:
		return " "
	}
}

// MarkColor returns the color used for a gutter mark.
func MarkColor(m GutterMark) string {
	switch m {
	case GutterError:
		return "red"
	case GutterWarning:
		return "yellow"
	case GutterInfo:
		return "blue"
	case GutterHint:
		return "green"
	case GutterBreakpoint:
		return "red"
	case GutterBookmark:
		return "gold"
	case GutterGitAdded:
		return "green"
	case GutterGitModified:
		return "blue"
	case GutterGitDeleted:
		return "red"
	case GutterCurrentLine:
		return "yellow"
	case GutterFoldOpen, GutterFoldClosed:
		return "240"
	default:
		return "250"
	}
}

// GutterOption configures a GutterRenderer.
type GutterOption func(*GutterRenderer)

// WithGutterWidth sets the width of the mark column.
func WithGutterWidth(width int) GutterOption {
	return func(g *GutterRenderer) {
		if width > 0 {
			g.gutterWidth = width
		}
	}
}

// WithGutterShowLineNumbers controls whether line numbers are rendered.
func WithGutterShowLineNumbers(show bool) GutterOption {
	return func(g *GutterRenderer) {
		g.showLineNumbers = show
	}
}

// GutterRenderer renders line-number and decoration gutters.
type GutterRenderer struct {
	decorations     []GutterDecoration
	maxLine         int
	gutterWidth     int
	showLineNumbers bool
	width           int
}

// NewGutterRenderer creates a new gutter renderer.
func NewGutterRenderer(maxLine int, decorations []GutterDecoration, opts ...GutterOption) *GutterRenderer {
	if maxLine < 1 {
		maxLine = 1
	}

	g := &GutterRenderer{
		decorations:     append([]GutterDecoration(nil), decorations...),
		maxLine:         maxLine,
		gutterWidth:     3,
		showLineNumbers: true,
		width:           len(strconv.Itoa(maxLine)),
	}

	for _, opt := range opts {
		opt(g)
	}

	if g.width < 1 {
		g.width = 1
	}
	if g.gutterWidth < 1 {
		g.gutterWidth = 1
	}

	return g
}

// SetDecorations replaces current gutter decorations.
func (g *GutterRenderer) SetDecorations(d []GutterDecoration) {
	g.decorations = append([]GutterDecoration(nil), d...)
}

// RenderLine renders one gutter row for the provided line number.
func (g *GutterRenderer) RenderLine(lineNum int) string {
	if lineNum > g.maxLine {
		g.maxLine = lineNum
		g.width = len(strconv.Itoa(g.maxLine))
	}

	lineStyle := lipgloss.NewStyle().Width(g.width).Align(lipgloss.Right).Faint(true)
	markStyle := lipgloss.NewStyle().Width(g.gutterWidth).Align(lipgloss.Center).Faint(true)

	icon := ""
	if dec, ok := g.decorationForLine(lineNum); ok {
		icon = MarkIcon(dec.Mark)
		markStyle = lipgloss.NewStyle().
			Width(g.gutterWidth).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color(MarkColor(dec.Mark)))
	}

	markCell := markStyle.Render(icon)
	if !g.showLineNumbers {
		return markCell
	}

	return lineStyle.Render(fmt.Sprintf("%d", lineNum)) + " " + markCell
}

// RenderGutter renders a gutter column from startLine through endLine, inclusive.
func (g *GutterRenderer) RenderGutter(startLine, endLine int) string {
	if startLine < 1 {
		startLine = 1
	}
	if endLine < startLine {
		return ""
	}

	if endLine > g.maxLine {
		g.maxLine = endLine
		g.width = len(strconv.Itoa(g.maxLine))
	}

	lines := make([]string, 0, endLine-startLine+1)
	for lineNum := startLine; lineNum <= endLine; lineNum++ {
		lines = append(lines, g.RenderLine(lineNum))
	}
	return strings.Join(lines, "\n")
}

// View renders the full gutter for all lines.
func (g *GutterRenderer) View() string {
	return g.RenderGutter(1, g.maxLine)
}

// ApplyToCode renders highlighted code with the gutter placed before each line.
func (g *GutterRenderer) ApplyToCode(code string, language string) string {
	if code == "" {
		return ""
	}

	highlighted := HighlightCode(code, language)
	codeLines := strings.Split(highlighted, "\n")

	if len(codeLines) > g.maxLine {
		g.maxLine = len(codeLines)
		g.width = len(strconv.Itoa(g.maxLine))
	}

	rows := make([]string, 0, len(codeLines))
	for i, line := range codeLines {
		rows = append(rows, g.RenderLine(i+1)+" "+line)
	}

	return strings.Join(rows, "\n")
}

func (g *GutterRenderer) decorationForLine(lineNum int) (GutterDecoration, bool) {
	for i := len(g.decorations) - 1; i >= 0; i-- {
		if g.decorations[i].Line == lineNum {
			return g.decorations[i], true
		}
	}
	return GutterDecoration{}, false
}
