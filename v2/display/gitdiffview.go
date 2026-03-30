package display

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// GitDiffFile is one changed file with its unified diff content.
type GitDiffFile struct {
	Path         string
	Status       FileStatus
	Diff         string
	LinesAdded   int
	LinesRemoved int
}

// GitDiffViewOption configures a GitDiffView.
type GitDiffViewOption func(*GitDiffView)

// WithGitDiffViewDesignTokens applies design-system tokens.
func WithGitDiffViewDesignTokens(tokens *design.DesignTokens) GitDiffViewOption {
	return func(v *GitDiffView) {
		if tokens != nil {
			v.designTokens = tokens
		}
	}
}

// WithGitDiffViewWidth sets a fixed rendering width.
func WithGitDiffViewWidth(width int) GitDiffViewOption {
	return func(v *GitDiffView) {
		if width >= 0 {
			v.width = width
		}
	}
}

// WithGitDiffViewHeight sets a fixed rendering height.
func WithGitDiffViewHeight(height int) GitDiffViewOption {
	return func(v *GitDiffView) {
		if height >= 0 {
			v.height = height
		}
	}
}

// GitDiffView renders a two-pane git diff explorer with file list + diff output.
type GitDiffView struct {
	files            []GitDiffFile
	fileList         *GitFileList
	selectedFile     int
	diffScrollOffset int
	focusedPane      int // 0=files, 1=diff
	width            int
	height           int
	designTokens     *design.DesignTokens
}

// NewGitDiffView creates a full-page git diff view.
func NewGitDiffView(files []GitDiffFile, opts ...GitDiffViewOption) *GitDiffView {
	v := &GitDiffView{
		files:            append([]GitDiffFile(nil), files...),
		selectedFile:     0,
		diffScrollOffset: 0,
		focusedPane:      0,
		designTokens:     design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(v)
	}

	v.fileList = NewGitFileList(v.buildFileEntries(),
		WithGitFileListDesignTokens(v.designTokens),
		WithGitFileListCursor(v.selectedFile),
	)
	v.fileList.Focus()
	v.clampSelectedFile()
	v.clampDiffScrollOffset()

	return v
}

// Init initializes the component.
func (v *GitDiffView) Init() tea.Cmd {
	if v.fileList == nil {
		return nil
	}
	return v.fileList.Init()
}

// Update handles pane focus, file selection, and diff scrolling.
func (v *GitDiffView) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if v.width <= 0 {
			v.width = msg.Width
		}
		if v.height <= 0 {
			v.height = msg.Height
		}
		if v.fileList != nil {
			leftWidth, _ := v.panelWidths()
			v.fileList.width = leftWidth
			_, cmd := v.fileList.Update(tea.WindowSizeMsg{Width: leftWidth, Height: v.height})
			v.clampDiffScrollOffset()
			return v, cmd
		}
		v.clampDiffScrollOffset()
		return v, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			v.focusedPane = (v.focusedPane + 1) % 2
			if v.focusedPane == 0 {
				if v.fileList != nil {
					v.fileList.Focus()
				}
			} else {
				if v.fileList != nil {
					v.fileList.Blur()
				}
			}
			return v, nil
		}

		if v.focusedPane == 0 {
			if v.fileList == nil {
				return v, nil
			}
			component, cmd := v.fileList.Update(msg)
			if updated, ok := component.(*GitFileList); ok {
				v.fileList = updated
			}
			return v, cmd
		}

		switch msg.String() {
		case "up", "k":
			v.diffScrollOffset--
			v.clampDiffScrollOffset()
		case "down", "j":
			v.diffScrollOffset++
			v.clampDiffScrollOffset()
		case "pgup", "pageup":
			jump := v.diffVisibleLines() - 1
			if jump < 1 {
				jump = 1
			}
			v.diffScrollOffset -= jump
			v.clampDiffScrollOffset()
		case "pgdown", "pagedown":
			jump := v.diffVisibleLines() - 1
			if jump < 1 {
				jump = 1
			}
			v.diffScrollOffset += jump
			v.clampDiffScrollOffset()
		case "home":
			v.diffScrollOffset = 0
		case "end":
			v.diffScrollOffset = v.maxDiffScrollOffset()
		}
		return v, nil
	case GitFileSelectedMsg:
		v.selectFileByPath(msg.Path)
		return v, nil
	}

	if v.fileList != nil && v.focusedPane == 0 {
		component, cmd := v.fileList.Update(msg)
		if updated, ok := component.(*GitFileList); ok {
			v.fileList = updated
		}
		return v, cmd
	}

	return v, nil
}

// View renders the two-pane git diff layout.
func (v *GitDiffView) View() string {
	leftWidth, rightWidth := v.panelWidths()
	viewHeight := v.effectiveHeight()

	if v.fileList != nil {
		v.fileList.width = leftWidth
	}

	leftPane := v.renderLeftPane(leftWidth, viewHeight)
	rightPane := v.renderRightPane(rightWidth, viewHeight)
	separator := v.separatorStyle() + "│" + style.ANSIReset

	out := make([]string, 0, viewHeight)
	for i := 0; i < viewHeight; i++ {
		out = append(out, leftPane[i]+separator+rightPane[i])
	}
	return strings.Join(out, "\n")
}

// Focus marks this component as focused.
func (v *GitDiffView) Focus() {
	if v.focusedPane == 0 {
		if v.fileList != nil {
			v.fileList.Focus()
		}
		return
	}
	if v.fileList != nil {
		v.fileList.Blur()
	}
}

// Blur marks this component as unfocused.
func (v *GitDiffView) Blur() {
	if v.fileList != nil {
		v.fileList.Blur()
	}
}

// Focused reports whether this component currently has focus.
func (v *GitDiffView) Focused() bool {
	if v.focusedPane == 0 {
		return v.fileList != nil && v.fileList.Focused()
	}
	return true
}

func (v *GitDiffView) buildFileEntries() []GitFileEntry {
	entries := make([]GitFileEntry, 0, len(v.files))
	for i, f := range v.files {
		entries = append(entries, GitFileEntry{
			Path:         f.Path,
			Status:       f.Status,
			LinesAdded:   f.LinesAdded,
			LinesRemoved: f.LinesRemoved,
			Selected:     i == v.selectedFile,
		})
	}
	return entries
}

func (v *GitDiffView) selectFileByPath(path string) {
	for i := range v.files {
		if v.files[i].Path == path {
			v.selectedFile = i
			v.diffScrollOffset = 0
			v.clampSelectedFile()
			if v.fileList != nil {
				v.fileList.cursor = i
				v.fileList.clampCursor()
			}
			return
		}
	}
}

func (v *GitDiffView) renderLeftPane(width, height int) []string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	content := "No changed files"
	if v.fileList != nil {
		content = v.fileList.View()
	}
	return v.fitPane(content, width, height)
}

func (v *GitDiffView) renderRightPane(width, height int) []string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	if len(v.files) == 0 {
		return v.fitPane(style.ANSIDim+"No diff selected"+style.ANSIReset, width, height)
	}

	v.clampSelectedFile()
	selected := v.files[v.selectedFile]

	headline := style.Truncate(selected.Path, width, "…")
	headline = style.Pad(headline, width)
	headline = style.ANSIBold + headline + style.ANSIReset

	stats := fmt.Sprintf("+%d -%d", maxInt(selected.LinesAdded, 0), maxInt(selected.LinesRemoved, 0))
	if v.focusedPane == 1 {
		stats = stats + " · diff"
	} else {
		stats = stats + " · files"
	}
	stats = style.Truncate(stats, width, "…")
	stats = style.Pad(stats, width)
	stats = style.ANSIDim + stats + style.ANSIReset

	diffLines := v.diffLinesForSelected()
	visible := v.diffVisibleLines()
	if visible < 1 {
		visible = 1
	}

	start := v.diffScrollOffset
	end := start + visible
	if end > len(diffLines) {
		end = len(diffLines)
	}

	rendered := make([]string, 0, height)
	rendered = append(rendered, headline)
	rendered = append(rendered, stats)

	for _, raw := range diffLines[start:end] {
		rendered = append(rendered, v.renderDiffLine(raw, width))
	}

	for len(rendered) < height {
		rendered = append(rendered, style.Pad("", width))
	}

	if len(rendered) > height {
		rendered = rendered[:height]
	}

	return rendered
}

func (v *GitDiffView) renderDiffLine(line string, width int) string {
	trimmed := style.Truncate(line, width, "…")
	padded := style.Pad(trimmed, width)

	switch {
	case strings.HasPrefix(line, "@@"):
		return v.accentStyle() + padded + style.ANSIReset
	case strings.HasPrefix(line, "+"):
		return v.successStyle() + padded + style.ANSIReset
	case strings.HasPrefix(line, "-"):
		return v.errorStyle() + padded + style.ANSIReset
	default:
		return style.ANSIDim + padded + style.ANSIReset
	}
}

func (v *GitDiffView) fitPane(content string, width, height int) []string {
	lines := strings.Split(content, "\n")
	fitted := make([]string, 0, height)

	for i := 0; i < len(lines) && len(fitted) < height; i++ {
		line := style.Truncate(lines[i], width, "…")
		fitted = append(fitted, style.Pad(line, width))
	}

	for len(fitted) < height {
		fitted = append(fitted, style.Pad("", width))
	}

	return fitted
}

func (v *GitDiffView) diffLinesForSelected() []string {
	if len(v.files) == 0 {
		return []string{"No diff"}
	}
	v.clampSelectedFile()
	lines := strings.Split(v.files[v.selectedFile].Diff, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return []string{"(no diff content)"}
	}
	return lines
}

func (v *GitDiffView) diffVisibleLines() int {
	visible := v.effectiveHeight() - 2
	if visible < 1 {
		return 1
	}
	return visible
}

func (v *GitDiffView) maxDiffScrollOffset() int {
	lines := len(v.diffLinesForSelected())
	visible := v.diffVisibleLines()
	if lines <= visible {
		return 0
	}
	return lines - visible
}

func (v *GitDiffView) clampSelectedFile() {
	if len(v.files) == 0 {
		v.selectedFile = 0
		return
	}
	if v.selectedFile < 0 {
		v.selectedFile = 0
	}
	if v.selectedFile >= len(v.files) {
		v.selectedFile = len(v.files) - 1
	}
}

func (v *GitDiffView) clampDiffScrollOffset() {
	if v.diffScrollOffset < 0 {
		v.diffScrollOffset = 0
	}
	maxOffset := v.maxDiffScrollOffset()
	if v.diffScrollOffset > maxOffset {
		v.diffScrollOffset = maxOffset
	}
}

func (v *GitDiffView) panelWidths() (left int, right int) {
	w := v.effectiveWidth()
	if w < 3 {
		w = 3
	}

	left = (w * 30) / 100
	if left < 1 {
		left = 1
	}
	if left > w-2 {
		left = w / 2
	}

	right = w - left - 1
	if right < 1 {
		right = 1
		if left > 1 {
			left = w - right - 1
		}
	}

	return left, right
}

func (v *GitDiffView) effectiveWidth() int {
	if v.width > 0 {
		return v.width
	}
	return 100
}

func (v *GitDiffView) effectiveHeight() int {
	if v.height > 0 {
		return v.height
	}
	return 24
}

func (v *GitDiffView) separatorStyle() string {
	if v.designTokens != nil {
		if c := style.Fg(v.designTokens.BorderSubtle); c != "" {
			return c
		}
	}
	return style.ANSIDim
}

func (v *GitDiffView) accentStyle() string {
	if v.designTokens != nil {
		if c := style.Fg(v.designTokens.Accent); c != "" {
			return c
		}
	}
	return style.ANSIBlue
}

func (v *GitDiffView) successStyle() string {
	if v.designTokens != nil {
		if c := style.Fg(v.designTokens.SuccessBright); c != "" {
			return c
		}
	}
	return style.ANSIGreen
}

func (v *GitDiffView) errorStyle() string {
	if v.designTokens != nil {
		if c := style.Fg(v.designTokens.ErrorBright); c != "" {
			return c
		}
	}
	return style.ANSIRed
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var _ tui.Component = (*GitDiffView)(nil)
