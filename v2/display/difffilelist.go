package display

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// FileChangeStatus identifies a diff file change type.
type FileChangeStatus int

const (
	StatusAdded FileChangeStatus = iota
	StatusModified
	StatusDeleted
	StatusRenamed
)

// DiffFileEntry is one changed file row in a diff list.
type DiffFileEntry struct {
	Path         string
	Status       FileChangeStatus
	LinesAdded   int
	LinesRemoved int
}

// DiffFileSelectedMsg is emitted when a file is selected with Enter.
type DiffFileSelectedMsg struct {
	File DiffFileEntry
}

// DiffFileListOption configures a DiffFileList.
type DiffFileListOption func(*DiffFileList)

// WithDiffFileListWidth sets preferred render width.
func WithDiffFileListWidth(width int) DiffFileListOption {
	return func(d *DiffFileList) {
		if width > 0 {
			d.width = width
		}
	}
}

// WithDiffFileListDesignTokens applies design-system tokens.
func WithDiffFileListDesignTokens(tokens *design.DesignTokens) DiffFileListOption {
	return func(d *DiffFileList) {
		if tokens != nil {
			d.designTokens = tokens
		}
	}
}

// DiffFileList renders changed files with colored status badges and line counts.
type DiffFileList struct {
	files        []DiffFileEntry
	cursor       int
	width        int
	windowWidth  int
	focused      bool
	designTokens *design.DesignTokens
}

// NewDiffFileList creates a new DiffFileList.
func NewDiffFileList(files []DiffFileEntry, opts ...DiffFileListOption) *DiffFileList {
	d := &DiffFileList{
		files:        append([]DiffFileEntry(nil), files...),
		cursor:       0,
		width:        0,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(d)
	}

	d.clampCursor()
	return d
}

// Init initializes the component.
func (d *DiffFileList) Init() tea.Cmd { return nil }

// Update handles keyboard navigation and selection.
func (d *DiffFileList) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.windowWidth = msg.Width
		return d, nil
	case tea.KeyMsg:
		if !d.focused || len(d.files) == 0 {
			return d, nil
		}
		switch msg.String() {
		case "up", "k":
			if d.cursor > 0 {
				d.cursor--
			}
			return d, nil
		case "down", "j":
			if d.cursor < len(d.files)-1 {
				d.cursor++
			}
			return d, nil
		case "enter":
			selected := d.files[d.cursor]
			return d, func() tea.Msg {
				return DiffFileSelectedMsg{File: selected}
			}
		}
	}
	return d, nil
}

// View renders the file list.
func (d *DiffFileList) View() string {
	if len(d.files) == 0 {
		return "No changed files"
	}

	width := d.effectiveWidth()
	lines := make([]string, 0, len(d.files))

	for i, file := range d.files {
		line := d.renderLine(file, width)
		if i == d.cursor {
			line = style.ANSIInverse + line + style.ANSIReset
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// Focus marks this component focused.
func (d *DiffFileList) Focus() { d.focused = true }

// Blur marks this component unfocused.
func (d *DiffFileList) Blur() { d.focused = false }

// Focused reports focus state.
func (d *DiffFileList) Focused() bool { return d.focused }

func (d *DiffFileList) effectiveWidth() int {
	if d.width > 0 {
		return d.width
	}
	if d.windowWidth > 0 {
		return d.windowWidth
	}
	return 80
}

func (d *DiffFileList) clampCursor() {
	if len(d.files) == 0 {
		d.cursor = 0
		return
	}
	if d.cursor < 0 {
		d.cursor = 0
	}
	if d.cursor >= len(d.files) {
		d.cursor = len(d.files) - 1
	}
}

func (d *DiffFileList) renderLine(file DiffFileEntry, width int) string {
	badgePlain, badgeStyled := d.badge(file.Status)
	statsPlain := fmt.Sprintf("+%d -%d", nonNegativeInt(file.LinesAdded), nonNegativeInt(file.LinesRemoved))
	statsStyled := d.successColor() + fmt.Sprintf("+%d", nonNegativeInt(file.LinesAdded)) + style.ANSIReset + " " + d.errorColor() + fmt.Sprintf("-%d", nonNegativeInt(file.LinesRemoved)) + style.ANSIReset
	base := style.StringWidth(badgePlain) + 1 + 1 + style.StringWidth(statsPlain)
	pathWidth := width - base
	if pathWidth < 1 {
		pathWidth = 1
	}
	path := style.Pad(style.ElidePath(file.Path, pathWidth), pathWidth)

	line := badgeStyled + " " + path + " " + statsStyled
	return style.Pad(style.Truncate(line, width, "…"), width)
}

func (d *DiffFileList) badge(status FileChangeStatus) (plain string, styled string) {
	letter := "?"
	color := style.ANSIWhite
	switch status {
	case StatusAdded:
		letter = "A"
		color = d.successColor()
	case StatusModified:
		letter = "M"
		color = d.warningColor()
	case StatusDeleted:
		letter = "D"
		color = d.errorColor()
	case StatusRenamed:
		letter = "R"
		color = d.accentColor()
	}
	plain = "[" + letter + "]"
	styled = color + plain + style.ANSIReset
	return plain, styled
}

func (d *DiffFileList) successColor() string {
	if d.designTokens != nil {
		if c := style.Fg(d.designTokens.SuccessBright); c != "" {
			return c
		}
	}
	return style.ANSIGreen
}

func (d *DiffFileList) errorColor() string {
	if d.designTokens != nil {
		if c := style.Fg(d.designTokens.ErrorBright); c != "" {
			return c
		}
	}
	return style.ANSIRed
}

func (d *DiffFileList) warningColor() string {
	if d.designTokens != nil {
		if c := style.Fg(d.designTokens.PendingColor); c != "" {
			return c
		}
	}
	return style.ANSIYellow
}

func (d *DiffFileList) accentColor() string {
	if d.designTokens != nil {
		if c := style.Fg(d.designTokens.Accent); c != "" {
			return c
		}
	}
	return style.ANSICyan
}

func nonNegativeInt(v int) int {
	if v < 0 {
		return 0
	}
	return v
}
var _ tui.Component = (*DiffFileList)(nil)
