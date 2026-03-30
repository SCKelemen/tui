package agent

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

var _ tui.Component = (*SubagentDiffList)(nil)

// SubagentDiffItem represents one file diff produced by a subagent.
type SubagentDiffItem struct {
	FilePath     string
	Status       string
	LinesAdded   int
	LinesRemoved int
	Diff         string
	Expanded     bool
}

// SubagentDiffExpandMsg expands/collapses one diff item by index.
type SubagentDiffExpandMsg struct {
	Index int
}

// SubagentDiffList renders grouped file diffs from a subagent's work.
type SubagentDiffList struct {
	items  []SubagentDiffItem
	width  int
	focus  bool
	cursor int

	designTokens *design.DesignTokens

	headerColor  string
	hunkColor    string
	addedColor   string
	removedColor string
	contextColor string
	mutedColor   string
}

// SubagentDiffListOption configures a SubagentDiffList.
type SubagentDiffListOption func(*SubagentDiffList)

// WithSubagentDiffListDesignTokens applies design-system tokens.
func WithSubagentDiffListDesignTokens(tokens *design.DesignTokens) SubagentDiffListOption {
	return func(d *SubagentDiffList) {
		d.designTokens = tokens
		d.applyDesignTokens(tokens)
	}
}

// WithSubagentDiffListWidth sets the render width.
func WithSubagentDiffListWidth(width int) SubagentDiffListOption {
	return func(d *SubagentDiffList) {
		if width > 0 {
			d.width = width
		}
	}
}

// NewSubagentDiffList creates a new SubagentDiffList.
func NewSubagentDiffList(items []SubagentDiffItem, opts ...SubagentDiffListOption) *SubagentDiffList {
	d := &SubagentDiffList{
		items:        []SubagentDiffItem{},
		width:        80,
		focus:        false,
		cursor:       0,
		headerColor:  style.ANSIReset,
		hunkColor:    style.Fg("#3B82F6"),
		addedColor:   style.Fg("#22C55E"),
		removedColor: style.Fg("#EF4444"),
		contextColor: style.ANSIDim,
		mutedColor:   style.ANSIDim,
	}

	if items != nil {
		d.items = append([]SubagentDiffItem(nil), items...)
	}

	for _, opt := range opts {
		opt(d)
	}

	if d.cursor >= len(d.items) {
		d.cursor = len(d.items) - 1
	}
	if d.cursor < 0 {
		d.cursor = 0
	}

	return d
}

// Init initializes the component.
func (d *SubagentDiffList) Init() tea.Cmd {
	return nil
}

// Update handles Bubble Tea messages.
func (d *SubagentDiffList) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			d.width = msg.Width
		}

	case SubagentDiffExpandMsg:
		if msg.Index >= 0 && msg.Index < len(d.items) {
			d.items[msg.Index].Expanded = !d.items[msg.Index].Expanded
		}

	case tea.KeyMsg:
		if !d.focus || len(d.items) == 0 {
			break
		}
		switch msg.String() {
		case "j", "down":
			d.cursor++
			if d.cursor >= len(d.items) {
				d.cursor = len(d.items) - 1
			}
		case "k", "up":
			d.cursor--
			if d.cursor < 0 {
				d.cursor = 0
			}
		case "enter", " ":
			d.items[d.cursor].Expanded = !d.items[d.cursor].Expanded
		case "a":
			for i := range d.items {
				d.items[i].Expanded = true
			}
		case "c":
			for i := range d.items {
				d.items[i].Expanded = false
			}
		}
	}

	return d, nil
}

// View renders the grouped file diffs.
func (d *SubagentDiffList) View() string {
	if d.width <= 0 {
		return ""
	}

	if len(d.items) == 0 {
		return fmt.Sprintf("%s(no file diffs)%s\n", d.mutedColor, style.ANSIReset)
	}

	var lines []string
	totalAdded := 0
	totalRemoved := 0

	for i, item := range d.items {
		totalAdded += item.LinesAdded
		totalRemoved += item.LinesRemoved

		caret := "▶"
		if item.Expanded {
			caret = "▼"
		}

		status := strings.TrimSpace(item.Status)
		if status == "" {
			status = "?"
		}

		header := fmt.Sprintf(
			"%s %s  [%s]  %s+%d%s %s-%d%s",
			caret,
			item.FilePath,
			status,
			d.addedColor,
			item.LinesAdded,
			style.ANSIReset,
			d.removedColor,
			item.LinesRemoved,
			style.ANSIReset,
		)
		header = d.headerColor + header + style.ANSIReset
		if d.focus && i == d.cursor {
			header = style.ANSIInverse + header + style.ANSIReset
		}
		lines = append(lines, d.fitToWidth(header))

		if !item.Expanded {
			continue
		}

		diffLines := strings.Split(item.Diff, "\n")
		for _, diffLine := range diffLines {
			if diffLine == "" {
				continue
			}

			lineColor := d.mutedColor
			switch {
			case strings.HasPrefix(diffLine, "@@"):
				lineColor = d.hunkColor
			case strings.HasPrefix(diffLine, "+"):
				lineColor = d.addedColor
			case strings.HasPrefix(diffLine, "-"):
				lineColor = d.removedColor
			case strings.HasPrefix(diffLine, " "):
				lineColor = d.contextColor
			}

			rendered := "  " + lineColor + diffLine + style.ANSIReset
			lines = append(lines, d.fitToWidth(rendered))
		}
	}

	summary := fmt.Sprintf(
		"%d files changed, %s+%d%s %s-%d%s",
		len(d.items),
		d.addedColor,
		totalAdded,
		style.ANSIReset,
		d.removedColor,
		totalRemoved,
		style.ANSIReset,
	)
	lines = append(lines, d.fitToWidth(d.mutedColor+summary+style.ANSIReset))

	return strings.Join(lines, "\n") + "\n"
}

// Focus marks the component as focused.
func (d *SubagentDiffList) Focus() {
	d.focus = true
}

// Blur marks the component as unfocused.
func (d *SubagentDiffList) Blur() {
	d.focus = false
}

// Focused returns whether the component is focused.
func (d *SubagentDiffList) Focused() bool {
	return d.focus
}

func (d *SubagentDiffList) fitToWidth(line string) string {
	if d.width <= 0 {
		return line
	}
	stripped := stripANSI(line)
	w := style.StringWidth(stripped)
	if w <= d.width {
		return line
	}
	return truncatePreservingANSI(line, d.width) + style.ANSIReset
}

func (d *SubagentDiffList) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}

	if v := strings.TrimSpace(tokens.Color); v != "" {
		d.headerColor = style.Fg(v)
	}
	if v := strings.TrimSpace(tokens.MutedColor); v != "" {
		d.mutedColor = style.Fg(v)
		d.contextColor = style.Fg(v)
	}
	if v := strings.TrimSpace(tokens.RunningColor); v != "" {
		d.hunkColor = style.Fg(v)
	}
	if v := strings.TrimSpace(tokens.SuccessBright); v != "" {
		d.addedColor = style.Fg(v)
	}
	if v := strings.TrimSpace(tokens.ErrorBright); v != "" {
		d.removedColor = style.Fg(v)
	}
}
