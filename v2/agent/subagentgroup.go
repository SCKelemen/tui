package agent

import (
	"fmt"
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// SubagentGroup renders multiple SubagentPanels in one horizontal row.
type SubagentGroup struct {
	panels    []*SubagentPanel
	width     int
	gap       int
	maxPerRow int

	focused      bool
	focusedPanel int
	headerText   string
	footerText   string
	focusHint    string

	designTokens *design.DesignTokens
	startTime    time.Time
}

// SubagentGroupOption configures a SubagentGroup.
type SubagentGroupOption func(*SubagentGroup)

// WithSubagentGroupGap sets horizontal spacing between panels.
func WithSubagentGroupGap(gap int) SubagentGroupOption {
	return func(g *SubagentGroup) {
		if gap < 0 {
			gap = 0
		}
		g.gap = gap
	}
}

// WithSubagentGroupWidth sets the total width for the group.
func WithSubagentGroupWidth(w int) SubagentGroupOption {
	return func(g *SubagentGroup) {
		if w > 0 {
			g.width = w
		}
	}
}

// WithSubagentGroupMaxPerRow sets the max number of panels per row.
func WithSubagentGroupMaxPerRow(n int) SubagentGroupOption {
	return func(g *SubagentGroup) {
		if n <= 0 {
			n = 1
		}
		g.maxPerRow = n
	}
}

// WithSubagentGroupDesignTokens applies design-system tokens.
func WithSubagentGroupDesignTokens(tokens *design.DesignTokens) SubagentGroupOption {
	return func(g *SubagentGroup) {
		if tokens != nil {
			g.designTokens = tokens
		}
	}
}

// WithSubagentGroupTheme applies a named theme.
func WithSubagentGroupTheme(theme string) SubagentGroupOption {
	return func(g *SubagentGroup) {
		g.designTokens = style.DesignTokensForTheme(theme)
	}
}

// WithSubagentGroupFocusHint sets the optional header hint text.
func WithSubagentGroupFocusHint(hint string) SubagentGroupOption {
	return func(g *SubagentGroup) {
		g.focusHint = strings.TrimSpace(hint)
	}
}

// NewSubagentGroup creates a new SubagentGroup.
func NewSubagentGroup(opts ...SubagentGroupOption) *SubagentGroup {
	g := &SubagentGroup{
		panels:       []*SubagentPanel{},
		width:        120,
		gap:          1,
		maxPerRow:    3,
		focusedPanel: 0,
		focusHint:    "",
		designTokens: design.DefaultTheme(),
		startTime:    time.Now(),
	}
	for _, opt := range opts {
		opt(g)
	}

	return g
}

// AddPanel appends one panel to the group.
func (g *SubagentGroup) AddPanel(p *SubagentPanel) {
	if p == nil {
		return
	}
	g.panels = append(g.panels, p)
	g.syncPanelFocus()
}

// SetPanels replaces all panels in the group.
func (g *SubagentGroup) SetPanels(panels []*SubagentPanel) {
	if panels == nil {
		g.panels = []*SubagentPanel{}
	} else {
		g.panels = append([]*SubagentPanel(nil), panels...)
	}

	if len(g.panels) == 0 {
		g.focusedPanel = 0
	} else if g.focusedPanel >= len(g.panels) {
		g.focusedPanel = len(g.panels) - 1
	}

	g.syncPanelFocus()
}

// SetWidth updates the group's total rendering width.
func (g *SubagentGroup) SetWidth(w int) {
	if w > 0 {
		g.width = w
	}
}

// GetPanels returns the current panel slice.
func (g *SubagentGroup) GetPanels() []*SubagentPanel {
	return g.panels
}

// PanelCount returns the number of panels.
func (g *SubagentGroup) PanelCount() int {
	return len(g.panels)
}

// FocusPanel focuses a specific panel by index.
func (g *SubagentGroup) FocusPanel(index int) {
	if index < 0 || index >= len(g.panels) {
		return
	}
	g.focusedPanel = index
	g.syncPanelFocus()
}

// IsAllCompleted reports whether all panels are completed.
func (g *SubagentGroup) IsAllCompleted() bool {
	if len(g.panels) == 0 {
		return false
	}
	for _, p := range g.panels {
		if p == nil || p.GetStatus() != SubagentCompleted {
			return false
		}
	}
	return true
}

// CompletedCount returns how many panels are completed.
func (g *SubagentGroup) CompletedCount() int {
	count := 0
	for _, p := range g.panels {
		if p != nil && p.GetStatus() == SubagentCompleted {
			count++
		}
	}
	return count
}

// Init initializes all child panels.
func (g *SubagentGroup) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, panel := range g.panels {
		if panel == nil {
			continue
		}
		cmds = append(cmds, panel.Init())
	}
	return tea.Batch(cmds...)
}

// Update handles group and panel events.
func (g *SubagentGroup) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			g.width = msg.Width
		}

		var cmds []tea.Cmd
		for _, panel := range g.panels {
			if panel == nil {
				continue
			}
			var cmd tea.Cmd
			_, cmd = panel.Update(msg)
			cmds = append(cmds, cmd)
		}
		return g, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+o":
			g.focused = !g.focused
			g.syncPanelFocus()
			return g, nil
		}

		if g.focused && len(g.panels) > 0 {
			switch msg.String() {
			case "tab", "right", "l":
				g.focusedPanel = (g.focusedPanel + 1) % len(g.panels)
				g.syncPanelFocus()
				return g, nil
			case "shift+tab", "left", "h":
				g.focusedPanel--
				if g.focusedPanel < 0 {
					g.focusedPanel = len(g.panels) - 1
				}
				g.syncPanelFocus()
				return g, nil
			}
		}
	}

	if len(g.panels) == 0 {
		return g, nil
	}

	idx := g.focusedPanel
	if idx < 0 || idx >= len(g.panels) {
		idx = 0
	}
	if g.panels[idx] == nil {
		return g, nil
	}

	updated, cmd := g.panels[idx].Update(msg)
	panel, ok := updated.(*SubagentPanel)
	if ok {
		g.panels[idx] = panel
	}
	return g, cmd
}

// View renders header, panel row, and optional completion footer.
func (g *SubagentGroup) View() string {
	if g.width <= 0 {
		return ""
	}

	g.headerText = g.renderHeaderText()

	var lines []string
	lines = append(lines, g.fitToWidth(g.headerText))

	panelLines := g.renderPanelRowLines()
	lines = append(lines, panelLines...)

	if g.IsAllCompleted() {
		g.footerText = g.renderFooterText()
		for _, footerLine := range strings.Split(g.footerText, "\n") {
			lines = append(lines, g.fitToWidth(footerLine))
		}
	} else {
		g.footerText = ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// Focus marks the group as focused.
func (g *SubagentGroup) Focus() {
	g.focused = true
	g.syncPanelFocus()
}

// Blur marks the group as not focused.
func (g *SubagentGroup) Blur() {
	g.focused = false
	g.syncPanelFocus()
}

// Focused returns whether the group is focused.
func (g *SubagentGroup) Focused() bool {
	return g.focused
}

func (g *SubagentGroup) syncPanelFocus() {
	for i, p := range g.panels {
		if p == nil {
			continue
		}
		if g.focused && i == g.focusedPanel {
			p.Focus()
		} else {
			p.Blur()
		}
	}
}

func (g *SubagentGroup) renderHeaderText() string {
	total := len(g.panels)
	completed := g.CompletedCount()

	text := "Running 0 subagents... (0/0 completed)"
	if total > 0 {
		if g.IsAllCompleted() {
			text = fmt.Sprintf("Ran %d subagents ✓", total)
		} else {
			text = fmt.Sprintf("Running %d subagents... (%d/%d completed)", total, completed, total)
		}
	}

	if g.focusHint != "" {
		text += " " + g.focusHint
	}
	return text
}

func (g *SubagentGroup) renderFooterText() string {
	total := len(g.panels)
	elapsed := time.Since(g.startTime)
	seconds := int(elapsed.Seconds())
	if seconds < 0 {
		seconds = 0
	}
	line1 := fmt.Sprintf(" Finished running %d subagents %s (%ds) ✓", total, g.renderProgressBar(), seconds)
	line2 := fmt.Sprintf(" %s└ Gathering results…%s", style.ANSIDim, style.ANSIReset)
	return line1 + "\n" + line2
}
func (g *SubagentGroup) renderProgressBar() string {
	if len(g.panels) == 0 {
		return ""
	}

	var b strings.Builder
	for _, p := range g.panels {
		if p != nil && p.GetStatus() == SubagentCompleted {
			b.WriteString(style.ANSIBold)
			b.WriteRune('▰')
			b.WriteString(style.ANSIReset)
		} else {
			b.WriteString(style.ANSIDim)
			b.WriteRune('▰')
			b.WriteString(style.ANSIReset)
		}
	}
	return b.String()
}

func (g *SubagentGroup) renderPanelRowLines() []string {
	if len(g.panels) == 0 {
		return nil
	}

	var rows [][]*SubagentPanel
	for i := 0; i < len(g.panels); i += g.maxPerRow {
		end := i + g.maxPerRow
		if end > len(g.panels) {
			end = len(g.panels)
		}
		rows = append(rows, g.panels[i:end])
	}

	var allLines []string
	for rowIdx, row := range rows {
		if rowIdx > 0 {
			allLines = append(allLines, "")
		}
		rowLines := g.renderSingleRow(row)
		allLines = append(allLines, rowLines...)
	}
	return allLines
}

func (g *SubagentGroup) renderSingleRow(panels []*SubagentPanel) []string {
	n := len(panels)
	if n == 0 {
		return nil
	}

	widths := g.computeRowWidths(n)
	perPanel := make([][]string, n)
	maxHeight := 0

	for i, panel := range panels {
		if panel == nil {
			perPanel[i] = []string{}
			continue
		}
		panel.width = widths[i]
		raw := strings.TrimSuffix(panel.View(), "\n")
		if raw == "" {
			perPanel[i] = []string{}
		} else {
			perPanel[i] = strings.Split(raw, "\n")
		}
		if len(perPanel[i]) > maxHeight {
			maxHeight = len(perPanel[i])
		}
	}

	for i := range perPanel {
		lines := perPanel[i]
		deficit := maxHeight - len(lines)
		if deficit <= 0 {
			continue
		}
		// Get this panel's surface color for padding background
		surfHex := "#31353D"
		if i < len(panels) && panels[i] != nil && panels[i].surfaceColor != "" {
			surfHex = panels[i].surfaceColor
		}
		padLine := style.Bg(surfHex) + strings.Repeat(" ", widths[i]) + style.ANSIReset

		// Ask the panel how many bottom lines to pin (footer + chips + border).
		pinned := 2 // default: footer + border
		if i < len(panels) && panels[i] != nil {
			pinned = panels[i].BottomPinnedLines()
		}

		// Insert padding before the pinned bottom section.
		if len(lines) > pinned {
			bottom := make([]string, pinned)
			copy(bottom, lines[len(lines)-pinned:])
			content := lines[:len(lines)-pinned]
			pad := make([]string, deficit)
			for j := range pad {
				pad[j] = padLine
			}
			perPanel[i] = append(content, pad...)
			perPanel[i] = append(perPanel[i], bottom...)
		} else {
			// Panel is too short; just append padding at the end
			for len(perPanel[i]) < maxHeight {
				perPanel[i] = append(perPanel[i], padLine)
			}
		}
	}
	gap := strings.Repeat(" ", g.gap)
	result := make([]string, 0, maxHeight)
	for lineIdx := 0; lineIdx < maxHeight; lineIdx++ {
		parts := make([]string, 0, n)
		for panelIdx := range perPanel {
			parts = append(parts, perPanel[panelIdx][lineIdx])
		}
		result = append(result, strings.Join(parts, gap))
	}
	return result
}
func (g *SubagentGroup) computeRowWidths(n int) []int {
	if n == 0 {
		return nil
	}

	available := g.width - (g.gap * (n - 1))
	if available < n {
		available = n
	}

	base := available / n
	rem := available % n

	widths := make([]int, n)
	for i := 0; i < n; i++ {
		widths[i] = base
		if i < rem {
			widths[i]++
		}
		if widths[i] <= 0 {
			widths[i] = 1
		}
	}

	return widths
}
func (g *SubagentGroup) fitToWidth(line string) string {
	if g.width <= 0 {
		return line
	}
	stripped := stripANSI(line)
	if style.StringWidth(stripped) <= g.width {
		return line
	}
	return style.Truncate(stripped, g.width, "…")
}
