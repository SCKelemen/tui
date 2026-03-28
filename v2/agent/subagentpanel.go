package agent

import (
	"fmt"
	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/spinner"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"time"
)

// SubagentStatus represents the subagent execution state.
type SubagentStatus int

const (
	SubagentRunning SubagentStatus = iota
	SubagentCompleted
	SubagentFailed
	SubagentAborted
)

// ToolStatus represents an individual tool execution state.
type ToolStatus int

const (
	ToolRunning ToolStatus = iota
	ToolCompleted
	ToolFailed
)

// SubagentTool represents one tool operation rendered in the panel.
type SubagentTool struct {
	Name   string
	Status ToolStatus
}

// SubagentPanel renders a single subagent panel.
type SubagentPanel struct {
	title    string
	thinking string

	status       SubagentStatus
	tools        []SubagentTool
	visibleTools int
	hiddenCount  int

	elapsed    time.Duration
	tokenCount string
	cost       string
	modelName  string

	width int

	spinner            spinner.Spinner
	spinnerIdx         int
	lastTick           time.Time
	focused            bool
	designTokens       *design.DesignTokens
	runningColor       string
	successBrightColor string // checkmarks, dots, 'Done'
	successMutedColor  string // durations, secondary success metadata
	errorBrightColor   string // crosses, failure dots
	errorMutedColor    string // failure durations, secondary error metadata
	abortedColor       string
	footerColor        string
	connectorColor     string
}

// SubagentPanelOption configures a SubagentPanel.
type SubagentPanelOption func(*SubagentPanel)

// WithSubagentTitle sets the panel title.
func WithSubagentTitle(title string) SubagentPanelOption {
	return func(p *SubagentPanel) {
		p.title = title
	}
}

// WithSubagentWidth sets the render width.
func WithSubagentWidth(width int) SubagentPanelOption {
	return func(p *SubagentPanel) {
		if width > 0 {
			p.width = width
		}
	}
}

// WithSubagentVisibleTools sets how many recent tools are shown.
func WithSubagentVisibleTools(n int) SubagentPanelOption {
	return func(p *SubagentPanel) {
		if n < 0 {
			n = 0
		}
		p.visibleTools = n
		p.recomputeHiddenCount()
	}
}

// WithSubagentDesignTokens applies design-system tokens.
func WithSubagentDesignTokens(tokens *design.DesignTokens) SubagentPanelOption {
	return func(p *SubagentPanel) {
		p.designTokens = tokens
		p.applyDesignTokens(tokens)
	}
}

// WithSubagentTheme applies a named design-system theme.
func WithSubagentTheme(theme string) SubagentPanelOption {
	return func(p *SubagentPanel) {
		tokens := style.DesignTokensForTheme(theme)
		p.designTokens = tokens
		p.applyDesignTokens(tokens)
	}
}

// WithSubagentModel sets the model name for footer rendering.
func WithSubagentModel(model string) SubagentPanelOption {
	return func(p *SubagentPanel) {
		p.modelName = model
	}
}

// WithSubagentSpinner sets the spinner animation.
func WithSubagentSpinner(s spinner.Spinner) SubagentPanelOption {
	return func(p *SubagentPanel) {
		p.spinner = s
	}
}

// WithSubagentThinking sets prose/thinking text rendered under the header.
func WithSubagentThinking(text string) SubagentPanelOption {
	return func(p *SubagentPanel) {
		p.thinking = text
	}
}

// subagentPanelTickMsg animates spinner + elapsed time updates.
type subagentPanelTickMsg struct {
	now time.Time
}

// NewSubagentPanel creates a new SubagentPanel.
func NewSubagentPanel(opts ...SubagentPanelOption) *SubagentPanel {
	p := &SubagentPanel{
		title:              "",
		status:             SubagentRunning,
		tools:              make([]SubagentTool, 0),
		visibleTools:       8,
		hiddenCount:        0,
		elapsed:            0,
		width:              36,
		spinner:            spinner.Braille,
		spinnerIdx:         0,
		runningColor:       style.Fg("#06B6D4"),
		successBrightColor: style.Fg("#88D67F"),
		successMutedColor:  style.Fg("#5B845C"),
		errorBrightColor:   style.Fg("#C96B72"),
		errorMutedColor:    style.Fg("#7A4044"),
		abortedColor:       style.ANSIDim,
		footerColor:        style.Fg("#7A818A"),
		connectorColor:     style.Fg("#7A818A"),
	}

	for _, opt := range opts {
		opt(p)
	}

	p.recomputeHiddenCount()
	return p
}

// Init starts animation ticks when running.
func (p *SubagentPanel) Init() tea.Cmd {
	if p.shouldTick() {
		return p.tick()
	}
	return nil
}

// Update handles Bubble Tea messages.
func (p *SubagentPanel) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			p.width = msg.Width
		}

	case subagentPanelTickMsg:
		if p.shouldTick() {
			if p.spinner.FrameCount() > 0 {
				p.spinnerIdx = (p.spinnerIdx + 1) % p.spinner.FrameCount()
			}

			if p.status == SubagentRunning {
				if p.lastTick.IsZero() {
					p.lastTick = msg.now
				} else {
					delta := msg.now.Sub(p.lastTick)
					if delta > 0 {
						p.elapsed += delta
					}
					p.lastTick = msg.now
				}
			}

			return p, p.tick()
		}
	}

	return p, nil
}

// View renders the subagent panel.
func (p *SubagentPanel) View() string {
	if p.width <= 0 {
		return ""
	}

	var lines []string

	lines = append(lines, strings.Repeat("▄", p.width))
	lines = append(lines, p.fitToWidth(p.renderHeaderLine()))

	if p.thinking != "" {
		thinkLines := strings.Split(p.thinking, "\n")
		for _, tl := range thinkLines {
			trimmed := strings.TrimSpace(tl)
			if trimmed == "" {
				continue
			}
			line := fmt.Sprintf(" %s%s%s", style.ANSIDim, style.Truncate(trimmed, p.width-2, "…"), style.ANSIReset)
			lines = append(lines, p.fitToWidth(line))
		}
	}

	if p.hiddenCount > 0 {
		hidden := fmt.Sprintf(" %s... +%d earlier tools%s", style.ANSIDim, p.hiddenCount, style.ANSIReset)
		lines = append(lines, p.fitToWidth(hidden))
	}

	start := 0
	if len(p.tools) > p.visibleTools {
		start = len(p.tools) - p.visibleTools
	}
	visible := p.tools[start:]

	for i, tool := range visible {
		connector := "├"
		if i == len(visible)-1 {
			connector = "└"
		}
		plainPrefix := fmt.Sprintf(" %s %s ", connector, p.renderToolPlainIcon(tool.Status))
		nameWidth := p.width - style.StringWidth(plainPrefix)
		if nameWidth < 0 {
			nameWidth = 0
		}
		toolName := smartElideToolName(tool.Name, nameWidth)
		line := fmt.Sprintf(" %s%s %s %s", p.connectorColor, connector, p.renderToolIcon(tool.Status), toolName)

		// Dim older completed tools
		if tool.Status == ToolCompleted && i < len(visible)-1 {
			line = fmt.Sprintf(" %s%s %s %s%s%s", p.connectorColor, connector, p.renderToolIcon(tool.Status), p.successMutedColor, toolName, style.ANSIReset)
		}

		line += style.ANSIReset
		lines = append(lines, p.fitToWidth(line))
	}
	lines = append(lines, p.fitToWidth(p.renderFooterLine()))
	lines = append(lines, strings.Repeat("▀", p.width))

	return strings.Join(lines, "\n") + "\n"
}

// Focus marks the panel as focused.
func (p *SubagentPanel) Focus() {
	p.focused = true
}

// Blur marks the panel as not focused.
func (p *SubagentPanel) Blur() {
	p.focused = false
}

// Focused reports whether the panel is focused.
func (p *SubagentPanel) Focused() bool {
	return p.focused
}

// SetStatus updates panel status.
func (p *SubagentPanel) SetStatus(status SubagentStatus) {
	if p.status != SubagentRunning && status == SubagentRunning {
		p.lastTick = time.Time{}
	}
	p.status = status
}

// SetElapsed sets elapsed duration.
func (p *SubagentPanel) SetElapsed(d time.Duration) {
	if d < 0 {
		d = 0
	}
	p.elapsed = d
}

// AdvanceSpinner manually advances the spinner frame.
// Use this when rendering outside of Bubble Tea's Update loop.
func (p *SubagentPanel) AdvanceSpinner() {
	if p.spinner.FrameCount() > 0 {
		p.spinnerIdx = (p.spinnerIdx + 1) % p.spinner.FrameCount()
	}
}

// SetTokenCount sets token count text.
func (p *SubagentPanel) SetTokenCount(s string) {
	p.tokenCount = strings.TrimSpace(s)
}

// SetCost sets cost text.
func (p *SubagentPanel) SetCost(s string) {
	p.cost = strings.TrimSpace(s)
}

// SetThinking sets prose/thinking text rendered under the header.
func (p *SubagentPanel) SetThinking(text string) {
	p.thinking = text
}

// AddTool appends a tool and recomputes hidden count.
func (p *SubagentPanel) AddTool(tool SubagentTool) {
	p.tools = append(p.tools, tool)
	p.recomputeHiddenCount()
}

// SetTools replaces tools and recomputes hidden count.
func (p *SubagentPanel) SetTools(tools []SubagentTool) {
	if tools == nil {
		p.tools = []SubagentTool{}
	} else {
		p.tools = append([]SubagentTool(nil), tools...)
	}
	p.recomputeHiddenCount()
}

// GetStatus returns the current panel status.
func (p *SubagentPanel) GetStatus() SubagentStatus {
	return p.status
}

func (p *SubagentPanel) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(now time.Time) tea.Msg {
		return subagentPanelTickMsg{now: now}
	})
}

func (p *SubagentPanel) shouldTick() bool {
	if p.status == SubagentRunning {
		return true
	}
	for _, t := range p.tools {
		if t.Status == ToolRunning {
			return true
		}
	}
	return false
}

func (p *SubagentPanel) recomputeHiddenCount() {
	hidden := len(p.tools) - p.visibleTools
	if hidden < 0 {
		hidden = 0
	}
	p.hiddenCount = hidden
}

func (p *SubagentPanel) renderHeaderLine() string {
	icon, plainIcon := p.renderStatusIcon()
	prefixPlain := " " + plainIcon + " Subagent: "
	available := p.width - style.StringWidth(prefixPlain)
	if available < 0 {
		available = 0
	}
	title := style.Truncate(p.title, available, "…")
	return fmt.Sprintf(" %s Subagent: %s", icon, title)
}
func (p *SubagentPanel) renderStatusIcon() (string, string) {
	switch p.status {
	case SubagentRunning:
		frame := "◐"
		if p.spinner.FrameCount() > 0 {
			frame = p.spinner.GetFrame(p.spinnerIdx)
		}
		return p.runningColor + frame + style.ANSIReset, frame
	case SubagentCompleted:
		return p.successBrightColor + style.ANSIBold + "●" + style.ANSIReset, "●"
	case SubagentFailed:
		return p.errorBrightColor + "✗" + style.ANSIReset, "✗"
	case SubagentAborted:
		return p.abortedColor + "○" + style.ANSIReset, "○"
	default:
		return "◐", "◐"
	}
}

func (p *SubagentPanel) renderToolIcon(status ToolStatus) string {
	switch status {
	case ToolRunning:
		frame := "◐"
		if p.spinner.FrameCount() > 0 {
			frame = p.spinner.GetFrame(p.spinnerIdx)
		}
		return p.runningColor + frame + style.ANSIReset
	case ToolCompleted:
		return p.successBrightColor + "✓" + style.ANSIReset
	case ToolFailed:
		return p.errorBrightColor + "✗" + style.ANSIReset
	default:
		return "·"
	}
}

func (p *SubagentPanel) renderToolPlainIcon(status ToolStatus) string {
	switch status {
	case ToolRunning:
		if p.spinner.FrameCount() > 0 {
			return p.spinner.GetFrame(p.spinnerIdx)
		}
		return "◐"
	case ToolCompleted:
		return "✓"
	case ToolFailed:
		return "✗"
	default:
		return "·"
	}
}
func (p *SubagentPanel) renderFooterLine() string {
	dur := formatSubagentDuration(p.elapsed)
	meta := p.buildMetaSuffix()

	switch p.status {
	case SubagentCompleted:
		left := fmt.Sprintf("%sDone%s %sin %s%s", p.successBrightColor, style.ANSIReset, p.successMutedColor, dur, style.ANSIReset)
		if meta != "" {
			return fmt.Sprintf(" %s %s%s%s", left, p.successMutedColor, meta, style.ANSIReset)
		}
		return " " + left
	case SubagentFailed:
		left := fmt.Sprintf("%sFailed%s %safter %s%s", p.errorBrightColor, style.ANSIReset, p.errorMutedColor, dur, style.ANSIReset)
		if meta != "" {
			return fmt.Sprintf(" %s %s%s%s", left, p.errorMutedColor, meta, style.ANSIReset)
		}
		return " " + left
	case SubagentAborted:
		left := fmt.Sprintf("%sAborted%s %safter %s%s", p.abortedColor, style.ANSIReset, p.footerColor, dur, style.ANSIReset)
		if meta != "" {
			return fmt.Sprintf(" %s %s%s%s", left, p.footerColor, meta, style.ANSIReset)
		}
		return " " + left
	default:
		left := fmt.Sprintf("%sWorking…%s %s%s%s", p.runningColor, style.ANSIReset, p.footerColor, dur, style.ANSIReset)
		if meta != "" {
			return fmt.Sprintf(" %s %s%s%s", left, p.footerColor, meta, style.ANSIReset)
		}
		return " " + left
	}
}

func (p *SubagentPanel) buildMetaSuffix() string {
	var parts []string
	if p.modelName != "" {
		parts = append(parts, p.modelName)
	}
	if p.tokenCount != "" {
		parts = append(parts, p.tokenCount)
	}
	if p.cost != "" {
		parts = append(parts, p.cost)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " · ")
}

func formatSubagentDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

func smartElideToolName(name string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if !strings.Contains(name, "/") {
		return style.Truncate(name, maxWidth, "…")
	}

	slashIdx := strings.Index(name, "/")
	start := strings.LastIndex(name[:slashIdx+1], " ") + 1
	end := len(name)
	if tailSpace := strings.Index(name[slashIdx:], " "); tailSpace >= 0 {
		end = slashIdx + tailSpace
	}

	prefix := name[:start]
	pathPart := name[start:end]
	suffix := name[end:]

	pathWidth := maxWidth - style.StringWidth(prefix) - style.StringWidth(suffix)
	if pathWidth <= 0 {
		return style.Truncate(name, maxWidth, "…")
	}

	candidate := prefix + style.ElidePath(pathPart, pathWidth) + suffix
	if style.StringWidth(candidate) <= maxWidth {
		return candidate
	}
	return style.Truncate(candidate, maxWidth, "…")
}

func (p *SubagentPanel) fitToWidth(line string) string {
	if p.width <= 0 {
		return line
	}
	stripped := stripANSI(line)
	w := style.StringWidth(stripped)
	if w > p.width {
		return style.Truncate(stripped, p.width, "…")
	}
	if w < p.width {
		return line + strings.Repeat(" ", p.width-w)
	}
	return line
}
func (p *SubagentPanel) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}
	if v := strings.TrimSpace(tokens.SuccessBright); v != "" {
		p.successBrightColor = style.Fg(v)
	}
	if v := strings.TrimSpace(tokens.SuccessMuted); v != "" {
		p.successMutedColor = style.Fg(v)
	}
	if v := strings.TrimSpace(tokens.ErrorBright); v != "" {
		p.errorBrightColor = style.Fg(v)
	}
	if v := strings.TrimSpace(tokens.ErrorMuted); v != "" {
		p.errorMutedColor = style.Fg(v)
	}
	if v := strings.TrimSpace(tokens.RunningColor); v != "" {
		p.runningColor = style.Fg(v)
	}
	if v := strings.TrimSpace(tokens.MutedColor); v != "" {
		p.footerColor = style.Fg(v)
		p.connectorColor = style.Fg(v)
		p.abortedColor = style.Fg(v)
	}
}
