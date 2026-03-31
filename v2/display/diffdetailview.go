package display

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// DiffFile is one file and its rendered diff hunks.
type DiffFile struct {
	Path  string
	Hunks []string
}

// DiffDetailCloseMsg is emitted when the detail view is closed.
type DiffDetailCloseMsg struct{}

// DiffDetailViewOption configures a DiffDetailView.
type DiffDetailViewOption func(*DiffDetailView)

// WithDiffDetailViewWidth sets a preferred render width.
func WithDiffDetailViewWidth(width int) DiffDetailViewOption {
	return func(d *DiffDetailView) {
		if width > 0 {
			d.width = width
		}
	}
}

// WithDiffDetailViewDesignTokens applies design-system tokens.
func WithDiffDetailViewDesignTokens(tokens *design.DesignTokens) DiffDetailViewOption {
	return func(d *DiffDetailView) {
		if tokens != nil {
			d.designTokens = tokens
		}
	}
}

// DiffDetailView renders file-by-file unified diff hunks.
type DiffDetailView struct {
	files        []DiffFile
	fileIndex    int
	scrollOffset int
	width        int
	windowWidth  int
	focused      bool
	designTokens *design.DesignTokens
}

// NewDiffDetailView creates a new DiffDetailView.
func NewDiffDetailView(files []DiffFile, opts ...DiffDetailViewOption) *DiffDetailView {
	d := &DiffDetailView{
		files:        append([]DiffFile(nil), files...),
		fileIndex:    0,
		scrollOffset: 0,
		width:        0,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(d)
	}

	d.clampFileIndex()
	d.clampScrollOffset()
	return d
}

// Init initializes the component.
func (d *DiffDetailView) Init() tea.Cmd { return nil }

// Update handles scrolling, file navigation, and close interactions.
func (d *DiffDetailView) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.windowWidth = msg.Width
		return d, nil
	case tea.KeyMsg:
		if !d.focused {
			return d, nil
		}
		switch msg.String() {
		case "j", "down":
			d.scrollOffset++
			d.clampScrollOffset()
			return d, nil
		case "k", "up":
			d.scrollOffset--
			d.clampScrollOffset()
			return d, nil
		case "n":
			if d.fileIndex < len(d.files)-1 {
				d.fileIndex++
				d.scrollOffset = 0
			}
			return d, nil
		case "p":
			if d.fileIndex > 0 {
				d.fileIndex--
				d.scrollOffset = 0
			}
			return d, nil
		case "q":
			return d, func() tea.Msg { return DiffDetailCloseMsg{} }
		}
	}
	return d, nil
}

// View renders the active file header, hunks, and navigation footer.
func (d *DiffDetailView) View() string {
	if len(d.files) == 0 {
		return "No diff files"
	}
	d.clampFileIndex()
	d.clampScrollOffset()

	width := d.effectiveWidth()
	active := d.files[d.fileIndex]

	header := fmt.Sprintf("%s[%d/%d]%s %s", style.ANSIBold, d.fileIndex+1, len(d.files), style.ANSIReset, active.Path)
	if width > 0 {
		header = style.Pad(style.Truncate(header, width, "…"), width)
	}

	allLines := d.renderHunkLines(active)
	visible := allLines
	if d.scrollOffset > 0 && d.scrollOffset < len(allLines) {
		visible = allLines[d.scrollOffset:]
	}

	footer := fmt.Sprintf("%s[j/k]%s scroll  %s[n/p]%s file  %s[q]%s close", d.mutedColor(), style.ANSIReset, d.mutedColor(), style.ANSIReset, d.mutedColor(), style.ANSIReset)
	if width > 0 {
		footer = style.Pad(style.Truncate(footer, width, "…"), width)
	}

	parts := []string{header}
	parts = append(parts, visible...)
	parts = append(parts, footer)
	return strings.Join(parts, "\n")
}

// Focus marks the component focused.
func (d *DiffDetailView) Focus() { d.focused = true }

// Blur marks the component unfocused.
func (d *DiffDetailView) Blur() { d.focused = false }

// Focused reports whether this component is focused.
func (d *DiffDetailView) Focused() bool { return d.focused }

func (d *DiffDetailView) effectiveWidth() int {
	if d.width > 0 {
		return d.width
	}
	return d.windowWidth
}

func (d *DiffDetailView) clampFileIndex() {
	if len(d.files) == 0 {
		d.fileIndex = 0
		return
	}
	if d.fileIndex < 0 {
		d.fileIndex = 0
	}
	if d.fileIndex >= len(d.files) {
		d.fileIndex = len(d.files) - 1
	}
}

func (d *DiffDetailView) clampScrollOffset() {
	lines := d.renderHunkLines(d.currentFile())
	if d.scrollOffset < 0 {
		d.scrollOffset = 0
	}
	if d.scrollOffset >= len(lines) {
		d.scrollOffset = len(lines) - 1
	}
	if d.scrollOffset < 0 {
		d.scrollOffset = 0
	}
}

func (d *DiffDetailView) currentFile() DiffFile {
	if len(d.files) == 0 {
		return DiffFile{}
	}
	d.clampFileIndex()
	return d.files[d.fileIndex]
}

func (d *DiffDetailView) renderHunkLines(file DiffFile) []string {
	if len(file.Hunks) == 0 {
		return []string{"(no hunks)"}
	}
	width := d.effectiveWidth()
	lines := make([]string, 0, len(file.Hunks))
	for _, hunk := range file.Hunks {
		line := hunk
		switch {
		case strings.HasPrefix(hunk, "+"):
			line = d.addColor() + hunk + style.ANSIReset
		case strings.HasPrefix(hunk, "-"):
			line = d.removeColor() + hunk + style.ANSIReset
		}
		if width > 0 {
			line = style.Pad(style.Truncate(line, width, "…"), width)
		}
		lines = append(lines, line)
	}
	return lines
}

func (d *DiffDetailView) addColor() string {
	if d.designTokens != nil {
		if c := style.Fg(d.designTokens.SuccessBright); c != "" {
			return c
		}
	}
	return style.ANSIGreen
}

func (d *DiffDetailView) removeColor() string {
	if d.designTokens != nil {
		if c := style.Fg(d.designTokens.ErrorBright); c != "" {
			return c
		}
	}
	return style.ANSIRed
}

func (d *DiffDetailView) mutedColor() string {
	if d.designTokens != nil {
		if c := style.Fg(d.designTokens.MutedColor); c != "" {
			return c
		}
	}
	return style.ANSIDim
}

var _ tui.Component = (*DiffDetailView)(nil)
