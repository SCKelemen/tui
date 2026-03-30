package display

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// GitFileEntry represents one changed file row in a git file list.
type GitFileEntry struct {
	Path         string
	Status       FileStatus
	LinesAdded   int
	LinesRemoved int
	Selected     bool
}

// GitFileSelectedMsg is emitted when Enter is pressed on a file row.
type GitFileSelectedMsg struct {
	Path string
}

// GitFileListOption configures a GitFileList.
type GitFileListOption func(*GitFileList)

// WithGitFileListDesignTokens applies design-system tokens.
func WithGitFileListDesignTokens(tokens *design.DesignTokens) GitFileListOption {
	return func(g *GitFileList) {
		if tokens != nil {
			g.designTokens = tokens
		}
	}
}

// WithGitFileListWidth sets the list width in terminal cells.
func WithGitFileListWidth(width int) GitFileListOption {
	return func(g *GitFileList) {
		if width >= 0 {
			g.width = width
		}
	}
}

// WithGitFileListCursor sets the initial cursor position.
func WithGitFileListCursor(cursor int) GitFileListOption {
	return func(g *GitFileList) {
		g.cursor = cursor
	}
}

// GitFileList renders a list of changed files with git status and diff stats.
type GitFileList struct {
	files        []GitFileEntry
	cursor       int
	width        int
	windowWidth  int
	focused      bool
	designTokens *design.DesignTokens
}

// NewGitFileList creates a new git file list.
func NewGitFileList(files []GitFileEntry, opts ...GitFileListOption) *GitFileList {
	g := &GitFileList{
		files:        append([]GitFileEntry(nil), files...),
		cursor:       0,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(g)
	}

	g.clampCursor()
	return g
}

// Init initializes the component.
func (g *GitFileList) Init() tea.Cmd {
	return nil
}

// Update handles keyboard navigation and selection.
func (g *GitFileList) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		g.windowWidth = msg.Width
		return g, nil
	case tea.KeyMsg:
		if !g.focused || len(g.files) == 0 {
			return g, nil
		}

		switch msg.String() {
		case "down", "j":
			if g.cursor < len(g.files)-1 {
				g.cursor++
			}
			return g, nil
		case "up", "k":
			if g.cursor > 0 {
				g.cursor--
			}
			return g, nil
		case "enter":
			selected := g.selectedFile()
			if selected == nil {
				return g, nil
			}
			return g, func() tea.Msg {
				return GitFileSelectedMsg{Path: selected.Path}
			}
		}
	}

	return g, nil
}

// View renders the changed file list.
func (g *GitFileList) View() string {
	if len(g.files) == 0 {
		return "No changed files"
	}

	width := g.effectiveWidth()
	g.clampCursor()

	lines := make([]string, 0, len(g.files)+1)
	for i := range g.files {
		lines = append(lines, g.renderFileLine(i, g.files[i], width))
	}
	lines = append(lines, g.renderSummaryLine())

	return strings.Join(lines, "\n")
}

// Focus marks the component as focused.
func (g *GitFileList) Focus() {
	g.focused = true
}

// Blur marks the component as unfocused.
func (g *GitFileList) Blur() {
	g.focused = false
}

// Focused reports whether this component is focused.
func (g *GitFileList) Focused() bool {
	return g.focused
}

func (g *GitFileList) selectedFile() *GitFileEntry {
	if len(g.files) == 0 {
		return nil
	}
	g.clampCursor()
	return &g.files[g.cursor]
}

func (g *GitFileList) clampCursor() {
	if len(g.files) == 0 {
		g.cursor = 0
		return
	}
	if g.cursor < 0 {
		g.cursor = 0
	}
	if g.cursor >= len(g.files) {
		g.cursor = len(g.files) - 1
	}
}

func (g *GitFileList) effectiveWidth() int {
	if g.width > 0 {
		return g.width
	}
	if g.windowWidth > 0 {
		return g.windowWidth
	}
	return 80
}

func (g *GitFileList) renderFileLine(index int, file GitFileEntry, width int) string {
	badge := NewFileStatusBadge(file.Status, WithFileStatusBadgeDesignTokens(g.designTokens))
	badgeText := badge.View()

	statsPlain, statsStyled := g.renderStats(file.LinesAdded, file.LinesRemoved)
	statsWidth := style.StringWidth(statsPlain)

	const badgeWidth = 3 // [X]
	baseWidth := badgeWidth + 1
	if statsWidth > 0 {
		baseWidth += 1 + statsWidth
	}

	pathWidth := width - baseWidth
	if pathWidth < 1 {
		pathWidth = 1
	}

	path := style.ElidePath(strings.TrimSpace(file.Path), pathWidth)
	if path == "" {
		path = " "
	}
	path = style.Pad(path, pathWidth)

	line := badgeText + " " + path
	if statsWidth > 0 {
		line += " " + statsStyled
	}

	if index == g.cursor {
		if bg := g.cursorBackground(); bg != "" {
			line = style.Bg(bg) + line + style.ANSIReset
		} else {
			line = style.ANSIInverse + line + style.ANSIReset
		}
	}

	return line
}

func (g *GitFileList) renderSummaryLine() string {
	fileCount := len(g.files)
	filesWord := "files"
	if fileCount == 1 {
		filesWord = "file"
	}

	totalAdded := 0
	totalRemoved := 0
	for _, file := range g.files {
		if file.LinesAdded > 0 {
			totalAdded += file.LinesAdded
		}
		if file.LinesRemoved > 0 {
			totalRemoved += file.LinesRemoved
		}
	}

	parts := []string{fmt.Sprintf("%d %s changed", fileCount, filesWord)}
	if totalAdded > 0 {
		parts = append(parts, g.successColor()+fmt.Sprintf("+%d", totalAdded)+style.ANSIReset)
	}
	if totalRemoved > 0 {
		parts = append(parts, g.errorColor()+fmt.Sprintf("-%d", totalRemoved)+style.ANSIReset)
	}

	return style.ANSIDim + strings.Join(parts, ", ") + style.ANSIReset
}

func (g *GitFileList) renderStats(added, removed int) (plain string, styled string) {
	partsPlain := make([]string, 0, 2)
	partsStyled := make([]string, 0, 2)

	if added > 0 {
		p := fmt.Sprintf("+%d", added)
		partsPlain = append(partsPlain, p)
		partsStyled = append(partsStyled, g.successColor()+p+style.ANSIReset)
	}
	if removed > 0 {
		p := fmt.Sprintf("-%d", removed)
		partsPlain = append(partsPlain, p)
		partsStyled = append(partsStyled, g.errorColor()+p+style.ANSIReset)
	}

	plain = strings.Join(partsPlain, " ")
	styled = strings.Join(partsStyled, " ")
	return plain, styled
}

func (g *GitFileList) successColor() string {
	if g.designTokens != nil {
		if v := strings.TrimSpace(g.designTokens.SuccessBright); v != "" {
			if c := style.Fg(v); c != "" {
				return c
			}
		}
	}
	return style.ANSIGreen
}

func (g *GitFileList) errorColor() string {
	if g.designTokens != nil {
		if v := strings.TrimSpace(g.designTokens.ErrorBright); v != "" {
			if c := style.Fg(v); c != "" {
				return c
			}
		}
	}
	return style.ANSIRed
}

func (g *GitFileList) cursorBackground() string {
	if g.designTokens != nil {
		if v := strings.TrimSpace(g.designTokens.SurfaceRaised); v != "" {
			if c := style.Bg(v); c != "" {
				return c
			}
		}
	}
	return style.Bg("#31353D")
}

var _ tui.Component = (*GitFileList)(nil)
