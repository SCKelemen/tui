package display

import (
	"fmt"
	"path/filepath"
	"strings"

	tui "github.com/SCKelemen/tui/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PeekFile is one file entry shown in a peek view.
type PeekFile struct {
	Path      string
	Language  string
	Content   string
	StartLine int
}

// PeekSelectedMsg is emitted when the active peek location is selected.
type PeekSelectedMsg struct {
	File PeekFile
	Line int
}

// PeekClosedMsg is emitted when the peek view is dismissed.
type PeekClosedMsg struct{}

// PeekView renders a VS Code-style peek definition panel.
type PeekView struct {
	title         string
	files         []PeekFile
	activeFile    int
	scrollOffset  int
	visibleLines  int
	width         int
	highlightLine int
	borderColor   string
	focused       bool
}

// PeekViewOption configures a PeekView.
type PeekViewOption func(*PeekView)

// WithPeekViewWidth sets the panel width in terminal cells.
func WithPeekViewWidth(width int) PeekViewOption {
	return func(p *PeekView) {
		if width > 0 {
			p.width = width
		}
	}
}

// WithPeekViewVisibleLines sets the number of visible code lines.
func WithPeekViewVisibleLines(lines int) PeekViewOption {
	return func(p *PeekView) {
		if lines > 0 {
			p.visibleLines = lines
		}
	}
}

// WithPeekViewHighlightLine sets the 1-indexed absolute line to accent.
func WithPeekViewHighlightLine(line int) PeekViewOption {
	return func(p *PeekView) {
		if line > 0 {
			p.highlightLine = line
		}
	}
}

// WithPeekViewBorderColor sets the left accent border color.
func WithPeekViewBorderColor(color string) PeekViewOption {
	return func(p *PeekView) {
		if strings.TrimSpace(color) != "" {
			p.borderColor = color
		}
	}
}

// NewPeekView creates a new peek definition view.
func NewPeekView(title string, files []PeekFile, opts ...PeekViewOption) *PeekView {
	p := &PeekView{
		title:         strings.TrimSpace(title),
		files:         append([]PeekFile(nil), files...),
		activeFile:    0,
		scrollOffset:  0,
		visibleLines:  10,
		width:         92,
		highlightLine: 0,
		borderColor:   "#22C55E",
	}

	for _, opt := range opts {
		opt(p)
	}

	p.ensureValidState()
	return p
}

// Init initializes the component.
func (p *PeekView) Init() tea.Cmd {
	return nil
}

// Update handles navigation, selection, and close behavior.
func (p *PeekView) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if p.width <= 0 || p.width > msg.Width {
			if msg.Width > 0 {
				p.width = msg.Width
			}
		}
		p.ensureValidState()
		return p, nil
	case tea.KeyMsg:
		if !p.focused {
			return p, nil
		}

		switch msg.String() {
		case "up", "k":
			if p.scrollOffset > 0 {
				p.scrollOffset--
			}
			return p, nil
		case "down", "j":
			maxOffset := p.maxScrollOffset()
			if p.scrollOffset < maxOffset {
				p.scrollOffset++
			}
			return p, nil
		case "left", "h":
			if len(p.files) > 1 {
				p.activeFile--
				if p.activeFile < 0 {
					p.activeFile = len(p.files) - 1
				}
				p.scrollOffset = 0
				p.ensureValidState()
			}
			return p, nil
		case "right", "l":
			if len(p.files) > 1 {
				p.activeFile = (p.activeFile + 1) % len(p.files)
				p.scrollOffset = 0
				p.ensureValidState()
			}
			return p, nil
		case "enter":
			selectedFile, ok := p.currentFile()
			if !ok {
				return p, nil
			}
			line := p.currentSelectedLine(selectedFile)
			return p, func() tea.Msg {
				return PeekSelectedMsg{File: selectedFile, Line: line}
			}
		case "esc":
			return p, func() tea.Msg {
				return PeekClosedMsg{}
			}
		}
	}

	return p, nil
}

// View renders the peek panel.
func (p *PeekView) View() string {
	p.ensureValidState()

	if len(p.files) == 0 {
		return p.renderPanel([]string{"No results"})
	}

	active, _ := p.currentFile()
	innerWidth := p.innerWidth()

	lines := make([]string, 0, p.visibleLines+8)
	lines = append(lines, p.renderHeader(active, innerWidth))

	if len(p.files) > 1 {
		lines = append(lines, p.renderTabs(innerWidth))
	}

	codeLines := p.renderCodeLines(active, innerWidth)
	lines = append(lines, codeLines...)

	if p.maxScrollOffset() > 0 {
		lines = append(lines, p.renderScrollIndicator(innerWidth))
	}

	return p.renderPanel(lines)
}

// Focus marks this component as focused.
func (p *PeekView) Focus() {
	p.focused = true
}

// Blur marks this component as unfocused.
func (p *PeekView) Blur() {
	p.focused = false
}

// Focused reports whether this component currently has focus.
func (p *PeekView) Focused() bool {
	return p.focused
}

func (p *PeekView) ensureValidState() {
	if p.visibleLines <= 0 {
		p.visibleLines = 10
	}
	if p.width <= 0 {
		p.width = 92
	}
	if len(p.files) == 0 {
		p.activeFile = 0
		p.scrollOffset = 0
		return
	}
	if p.activeFile < 0 {
		p.activeFile = 0
	}
	if p.activeFile >= len(p.files) {
		p.activeFile = len(p.files) - 1
	}
	if p.scrollOffset < 0 {
		p.scrollOffset = 0
	}
	maxOffset := p.maxScrollOffset()
	if p.scrollOffset > maxOffset {
		p.scrollOffset = maxOffset
	}
}

func (p *PeekView) currentFile() (PeekFile, bool) {
	if len(p.files) == 0 {
		return PeekFile{}, false
	}
	idx := p.activeFile
	if idx < 0 || idx >= len(p.files) {
		idx = 0
	}
	file := p.files[idx]
	if file.StartLine <= 0 {
		file.StartLine = 1
	}
	return file, true
}

func (p *PeekView) currentSelectedLine(file PeekFile) int {
	lines := splitLines(file.Content)
	if len(lines) == 0 {
		return file.StartLine
	}

	if p.highlightLine > 0 {
		min := file.StartLine
		max := file.StartLine + len(lines) - 1
		if p.highlightLine >= min && p.highlightLine <= max {
			return p.highlightLine
		}
	}

	return file.StartLine + p.scrollOffset
}

func (p *PeekView) maxScrollOffset() int {
	file, ok := p.currentFile()
	if !ok {
		return 0
	}
	lines := splitLines(file.Content)
	if len(lines) <= p.visibleLines {
		return 0
	}
	return len(lines) - p.visibleLines
}

func (p *PeekView) innerWidth() int {
	inner := p.width - 3
	if inner < 30 {
		inner = 30
	}
	return inner
}

func (p *PeekView) renderHeader(file PeekFile, width int) string {
	title := p.title
	if title == "" {
		title = "Peek"
	}
	fileCount := fmt.Sprintf("%d file", len(p.files))
	if len(p.files) != 1 {
		fileCount = fmt.Sprintf("%d files", len(p.files))
	}
	pathLabel := strings.TrimSpace(file.Path)
	if pathLabel == "" {
		pathLabel = "(untitled)"
	}

	left := fmt.Sprintf("%s • %s", title, fileCount)
	header := left + "  " + pathLabel
	if lipgloss.Width(header) > width {
		header = truncateANSI(header, width)
	}

	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#E5E7EB")).
		Background(lipgloss.Color("#1F2937")).
		Width(width).
		Render(header)
}

func (p *PeekView) renderTabs(width int) string {
	parts := make([]string, 0, len(p.files))
	for i := range p.files {
		name := filepath.Base(strings.TrimSpace(p.files[i].Path))
		if name == "" {
			name = fmt.Sprintf("file-%d", i+1)
		}
		label := " " + name + " "
		if i == p.activeFile {
			parts = append(parts, lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#E5E7EB")).
				Background(lipgloss.Color("#374151")).
				Render(label))
		} else {
			parts = append(parts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF")).
				Background(lipgloss.Color("#111827")).
				Render(label))
		}
	}

	tabLine := strings.Join(parts, " ")
	if lipgloss.Width(tabLine) > width {
		tabLine = truncateANSI(tabLine, width)
	}

	return lipgloss.NewStyle().Width(width).Render(tabLine)
}

func (p *PeekView) renderCodeLines(file PeekFile, width int) []string {
	lines := splitLines(file.Content)
	if len(lines) == 0 {
		return []string{lipgloss.NewStyle().Faint(true).Width(width).Render("(empty)")}
	}

	start := p.scrollOffset
	if start < 0 {
		start = 0
	}
	if start > len(lines) {
		start = len(lines)
	}

	end := start + p.visibleLines
	if end > len(lines) {
		end = len(lines)
	}

	lineNoWidth := len(fmt.Sprintf("%d", file.StartLine+len(lines)-1))
	highlighter := p.highlighterForFile(file)

	result := make([]string, 0, p.visibleLines)
	for i := start; i < end; i++ {
		lineNum := file.StartLine + i
		content := lines[i]
		if highlighter != nil {
			content = highlighter.HighlightLine(content)
		}

		gutter := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Render(fmt.Sprintf("%*d", lineNoWidth, lineNum))

		body := fmt.Sprintf("%s %s", gutter, content)
		body = truncateANSI(body, width)

		lineStyle := lipgloss.NewStyle().Width(width).Background(lipgloss.Color("#111827"))
		if lineNum == p.highlightLine {
			lineStyle = lineStyle.Background(lipgloss.Color("#1F2937"))
		}

		result = append(result, lineStyle.Render(body))
	}

	for len(result) < p.visibleLines {
		result = append(result, lipgloss.NewStyle().Width(width).Background(lipgloss.Color("#111827")).Render(""))
	}

	return result
}

func (p *PeekView) renderScrollIndicator(width int) string {
	file, ok := p.currentFile()
	if !ok {
		return ""
	}
	lines := splitLines(file.Content)
	if len(lines) == 0 {
		return ""
	}

	top := p.scrollOffset + 1
	bottom := p.scrollOffset + p.visibleLines
	if bottom > len(lines) {
		bottom = len(lines)
	}

	indicator := fmt.Sprintf("Lines %d-%d of %d", top, bottom, len(lines))
	return lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("#9CA3AF")).
		Width(width).
		Render(indicator)
}

func (p *PeekView) renderPanel(lines []string) string {
	innerWidth := p.innerWidth()
	if len(lines) == 0 {
		lines = []string{""}
	}

	leftBorder := lipgloss.NewStyle().Foreground(lipgloss.Color(p.borderColor)).Render("│")
	rightBorder := lipgloss.NewStyle().Foreground(lipgloss.Color("#4B5563")).Render("│")
	topRight := lipgloss.NewStyle().Foreground(lipgloss.Color("#4B5563")).Render("╮")
	bottomRight := lipgloss.NewStyle().Foreground(lipgloss.Color("#4B5563")).Render("╯")
	horizontal := lipgloss.NewStyle().Foreground(lipgloss.Color("#4B5563")).Render(strings.Repeat("─", innerWidth+1))

	out := make([]string, 0, len(lines)+2)
	out = append(out, leftBorder+horizontal+topRight)

	for i := range lines {
		line := truncateANSI(lines[i], innerWidth)
		content := lipgloss.NewStyle().Width(innerWidth).Render(line)
		out = append(out, leftBorder+" "+content+rightBorder)
	}

	out = append(out, leftBorder+horizontal+bottomRight)
	return strings.Join(out, "\n")
}

func (p *PeekView) highlighterForFile(file PeekFile) *Highlighter {
	lang := parseSyntaxLanguage(file.Language)
	if lang == SyntaxLanguagePlain {
		if detected := DetectLanguage(file.Path); detected != SyntaxLanguagePlain {
			lang = detected
		}
	}
	if lang == SyntaxLanguagePlain {
		return nil
	}
	return NewHighlighter(lang)
}

func splitLines(content string) []string {
	if content == "" {
		return nil
	}
	trimmed := strings.ReplaceAll(content, "\r\n", "\n")
	trimmed = strings.ReplaceAll(trimmed, "\r", "\n")
	return strings.Split(trimmed, "\n")
}

var _ tui.Component = (*PeekView)(nil)
