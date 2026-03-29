package display

import (
	"fmt"
	"sort"
	"strings"

	tui "github.com/SCKelemen/tui/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DiagnosticSeverity describes the importance level of a diagnostic.
type DiagnosticSeverity int

const (
	// SeverityError indicates a blocking error.
	SeverityError DiagnosticSeverity = iota
	// SeverityWarning indicates a non-blocking warning.
	SeverityWarning
	// SeverityInfo indicates informational output.
	SeverityInfo
	// SeverityHint indicates a low-priority suggestion.
	SeverityHint
)

// SeverityIcon returns the icon used for a severity level.
func SeverityIcon(s DiagnosticSeverity) string {
	switch s {
	case SeverityError:
		return "✗"
	case SeverityWarning:
		return "⚠"
	case SeverityInfo:
		return "ℹ"
	case SeverityHint:
		return "💡"
	default:
		return "•"
	}
}

// SeverityColor returns the color name used for a severity level.
func SeverityColor(s DiagnosticSeverity) string {
	switch s {
	case SeverityError:
		return "red"
	case SeverityWarning:
		return "yellow"
	case SeverityInfo:
		return "blue"
	case SeverityHint:
		return "green"
	default:
		return "white"
	}
}

// Diagnostic is a single compiler/linter problem entry.
type Diagnostic struct {
	Message     string
	Severity    DiagnosticSeverity
	File        string
	Line        int
	Column      int
	Source      string
	Code        string
	RelatedInfo []string
}

// DiagnosticSelectedMsg is emitted when Enter is pressed on a diagnostic.
type DiagnosticSelectedMsg struct {
	Diagnostic Diagnostic
}

// DiagnosticListOption configures a DiagnosticList.
type DiagnosticListOption func(*DiagnosticList)

// WithDiagnosticListWidth sets a fixed width for the panel.
func WithDiagnosticListWidth(width int) DiagnosticListOption {
	return func(d *DiagnosticList) {
		if width >= 0 {
			d.width = width
		}
	}
}

// WithDiagnosticListGroupByFile controls file grouping behavior.
func WithDiagnosticListGroupByFile(group bool) DiagnosticListOption {
	return func(d *DiagnosticList) {
		d.groupByFile = group
	}
}

// WithDiagnosticShowCounts controls whether the summary header is rendered.
func WithDiagnosticShowCounts(show bool) DiagnosticListOption {
	return func(d *DiagnosticList) {
		d.showCounts = show
	}
}

// DiagnosticList renders a VS Code-style problems panel.
type DiagnosticList struct {
	diagnostics []Diagnostic
	cursor      int
	width       int
	groupByFile bool
	showCounts  bool
	focused     bool
	windowWidth int
}

// NewDiagnosticList creates a diagnostic list with optional configuration.
func NewDiagnosticList(diagnostics []Diagnostic, opts ...DiagnosticListOption) *DiagnosticList {
	d := &DiagnosticList{
		diagnostics: append([]Diagnostic(nil), diagnostics...),
		cursor:      0,
		groupByFile: true,
		showCounts:  true,
	}

	for _, opt := range opts {
		opt(d)
	}

	d.clampCursor()
	return d
}

// Init initializes the component.
func (d *DiagnosticList) Init() tea.Cmd {
	return nil
}

// Update handles keyboard navigation and selection.
func (d *DiagnosticList) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.windowWidth = msg.Width
		return d, nil
	case tea.KeyMsg:
		if !d.focused || len(d.diagnostics) == 0 {
			return d, nil
		}

		switch msg.String() {
		case "up":
			if d.cursor > 0 {
				d.cursor--
			}
			return d, nil
		case "down":
			if d.cursor < len(d.diagnostics)-1 {
				d.cursor++
			}
			return d, nil
		case "enter":
			selected := d.selectedDiagnostic()
			if selected == nil {
				return d, nil
			}
			return d, func() tea.Msg {
				return DiagnosticSelectedMsg{Diagnostic: *selected}
			}
		}
	}

	return d, nil
}

// View renders the problems panel.
func (d *DiagnosticList) View() string {
	if len(d.diagnostics) == 0 {
		empty := lipgloss.NewStyle().Faint(true).Render("No problems")
		return d.wrapPanel(empty)
	}

	lines := make([]string, 0, len(d.diagnostics)+8)
	if d.showCounts {
		lines = append(lines, d.renderHeader())
	}

	if d.groupByFile {
		lines = append(lines, d.renderGrouped()...)
	} else {
		for i := range d.diagnostics {
			lines = append(lines, d.renderDiagnosticLine(i, d.diagnostics[i], ""))
		}
	}

	return d.wrapPanel(strings.Join(lines, "\n"))
}

// Focus marks the component as focused.
func (d *DiagnosticList) Focus() {
	d.focused = true
}

// Blur marks the component as unfocused.
func (d *DiagnosticList) Blur() {
	d.focused = false
}

// Focused reports whether the component is focused.
func (d *DiagnosticList) Focused() bool {
	return d.focused
}

// SetDiagnostics replaces current diagnostics and clamps cursor bounds.
func (d *DiagnosticList) SetDiagnostics(diags []Diagnostic) {
	d.diagnostics = append([]Diagnostic(nil), diags...)
	d.clampCursor()
}

// Counts returns the number of diagnostics by severity.
func (d *DiagnosticList) Counts() (errors, warnings, infos, hints int) {
	for _, diag := range d.diagnostics {
		switch diag.Severity {
		case SeverityError:
			errors++
		case SeverityWarning:
			warnings++
		case SeverityInfo:
			infos++
		case SeverityHint:
			hints++
		}
	}
	return errors, warnings, infos, hints
}

func (d *DiagnosticList) clampCursor() {
	if len(d.diagnostics) == 0 {
		d.cursor = 0
		return
	}
	if d.cursor < 0 {
		d.cursor = 0
	}
	if d.cursor >= len(d.diagnostics) {
		d.cursor = len(d.diagnostics) - 1
	}
}

func (d *DiagnosticList) selectedDiagnostic() *Diagnostic {
	if len(d.diagnostics) == 0 {
		return nil
	}
	d.clampCursor()
	return &d.diagnostics[d.cursor]
}

func (d *DiagnosticList) renderHeader() string {
	errors, warnings, infos, hints := d.Counts()
	parts := make([]string, 0, 4)

	if errors > 0 {
		parts = append(parts, d.renderCountPart(SeverityError, errors, "error"))
	}
	if warnings > 0 {
		parts = append(parts, d.renderCountPart(SeverityWarning, warnings, "warning"))
	}
	if infos > 0 {
		parts = append(parts, d.renderCountPart(SeverityInfo, infos, "info"))
	}
	if hints > 0 {
		parts = append(parts, d.renderCountPart(SeverityHint, hints, "hint"))
	}

	if len(parts) == 0 {
		parts = append(parts, lipgloss.NewStyle().Faint(true).Render("No problems"))
	}

	return lipgloss.NewStyle().Bold(true).Render(strings.Join(parts, "  "))
}

func (d *DiagnosticList) renderCountPart(severity DiagnosticSeverity, n int, noun string) string {
	if n != 1 {
		noun += "s"
	}
	icon := d.severityStyle(severity).Render(SeverityIcon(severity))
	text := lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("%d %s", n, noun))
	return icon + " " + text
}

func (d *DiagnosticList) renderGrouped() []string {
	grouped := make(map[string][]int)
	order := make([]string, 0)
	for i, diag := range d.diagnostics {
		file := strings.TrimSpace(diag.File)
		if file == "" {
			file = "(unknown file)"
		}
		if _, ok := grouped[file]; !ok {
			order = append(order, file)
		}
		grouped[file] = append(grouped[file], i)
	}

	sort.Strings(order)

	out := make([]string, 0, len(d.diagnostics)+len(order))
	for _, file := range order {
		indices := grouped[file]
		header := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%s (%d)", file, len(indices)))
		out = append(out, header)
		for _, idx := range indices {
			out = append(out, d.renderDiagnosticLine(idx, d.diagnostics[idx], "  "))
		}
	}

	return out
}

func (d *DiagnosticList) renderDiagnosticLine(index int, diag Diagnostic, indent string) string {
	icon := d.severityStyle(diag.Severity).Render(SeverityIcon(diag.Severity))
	message := strings.TrimSpace(diag.Message)
	if message == "" {
		message = "(no message)"
	}

	metaParts := make([]string, 0, 2)
	metaParts = append(metaParts, d.locationString(diag))
	if srcCode := d.sourceCodeString(diag); srcCode != "" {
		metaParts = append(metaParts, srcCode)
	}
	meta := lipgloss.NewStyle().Faint(true).Render(strings.Join(metaParts, " "))

	prefix := "  "
	if index == d.cursor {
		if d.focused {
			prefix = lipgloss.NewStyle().Foreground(lipgloss.Color("#61AFEF")).Bold(true).Render("❯ ")
		} else {
			prefix = lipgloss.NewStyle().Faint(true).Render("• ")
		}
	}

	line := fmt.Sprintf("%s%s%s %s", indent, prefix, icon, message)
	if meta != "" {
		line += " " + meta
	}

	if d.focused && index == d.cursor {
		line = lipgloss.NewStyle().Bold(true).Render(line)
	}
	return line
}

func (d *DiagnosticList) locationString(diag Diagnostic) string {
	file := strings.TrimSpace(diag.File)
	if file == "" {
		file = "(unknown file)"
	}
	line := diag.Line
	column := diag.Column
	if line < 0 {
		line = 0
	}
	if column < 0 {
		column = 0
	}
	return fmt.Sprintf("%s:%d:%d", file, line, column)
}

func (d *DiagnosticList) sourceCodeString(diag Diagnostic) string {
	bits := make([]string, 0, 2)
	if strings.TrimSpace(diag.Source) != "" {
		bits = append(bits, strings.TrimSpace(diag.Source))
	}
	if strings.TrimSpace(diag.Code) != "" {
		bits = append(bits, strings.TrimSpace(diag.Code))
	}
	if len(bits) == 0 {
		return ""
	}
	return "(" + strings.Join(bits, " ") + ")"
}

func (d *DiagnosticList) severityStyle(severity DiagnosticSeverity) lipgloss.Style {
	color := SeverityColor(severity)
	switch color {
	case "red":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#E06C75"))
	case "yellow":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#E5C07B"))
	case "blue":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#61AFEF"))
	case "green":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#98C379"))
	default:
		return lipgloss.NewStyle()
	}
}

func (d *DiagnosticList) wrapPanel(content string) string {
	width := d.width
	if width <= 0 {
		width = d.windowWidth
	}
	if width <= 0 {
		return content
	}
	return lipgloss.NewStyle().Width(width).MaxWidth(width).Render(content)
}

var _ tui.Component = (*DiagnosticList)(nil)
