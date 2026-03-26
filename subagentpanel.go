package tui

import (
	"fmt"
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tea "github.com/charmbracelet/bubbletea"
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
	title string

	status SubagentStatus
	tools  []SubagentTool

	visibleTools int
	hiddenCount  int

	elapsed    time.Duration
	tokenCount string
	cost       string
	modelName  string

	width int

	spinner       Spinner
	spinnerIdx    int
	lastTick      time.Time
	focused       bool
	designTokens  *design.DesignTokens
	runningColor  string
	successColor  string
	errorColor    string
	abortedColor  string
	footerColor   string
	connectorColor string
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
		tokens := designTokensForTheme(theme)
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

// subagentPanelTickMsg animates spinner + elapsed time updates.
type subagentPanelTickMsg struct {
	now time.Time
}

// NewSubagentPanel creates a new SubagentPanel.
func NewSubagentPanel(opts ...SubagentPanelOption) *SubagentPanel {
	p := &SubagentPanel{
		title:         "",
		status:        SubagentRunning,
		tools:         make([]SubagentTool, 0),
		visibleTools:  8,
		hiddenCount:   0,
		elapsed:       0,
		width:         36,
		spinner:       SpinnerCircleQuarters,
		spinnerIdx:    0,
		runningColor:  ansiCyan,
		successColor:  ansiGreen,
		errorColor:    ansiRed,
		abortedColor:  ansiDim,
		footerColor:   ansiDim,
		connectorColor: ansiDim,
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
func (p *SubagentPanel) Update(msg tea.Msg) (Component, tea.Cmd) {
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

	if p.hiddenCount > 0 {
		hidden := fmt.Sprintf(" %s... +%d earlier tools%s", ansiDim, p.hiddenCount, ansiReset)
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
		line := fmt.Sprintf(" %s%s %s %s", p.connectorColor, connector, p.renderToolIcon(tool.Status), tool.Name)
		line += ansiReset
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

// SetTokenCount sets token count text.
func (p *SubagentPanel) SetTokenCount(s string) {
	p.tokenCount = strings.TrimSpace(s)
}

// SetCost sets cost text.
func (p *SubagentPanel) SetCost(s string) {
	p.cost = strings.TrimSpace(s)
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
	available := p.width - len(prefixPlain)
	if available < 0 {
		available = 0
	}
	title := truncateString(p.title, available)
	return fmt.Sprintf(" %s Subagent: %s", icon, title)
}

func (p *SubagentPanel) renderStatusIcon() (string, string) {
	switch p.status {
	case SubagentRunning:
		frame := "◐"
		if p.spinner.FrameCount() > 0 {
			frame = p.spinner.GetFrame(p.spinnerIdx)
		}
		return p.runningColor + frame + ansiReset, frame
	case SubagentCompleted:
		return p.successColor + ansiBold + "●" + ansiReset, "●"
	case SubagentFailed:
		return p.errorColor + "✗" + ansiReset, "✗"
	case SubagentAborted:
		return p.abortedColor + "○" + ansiReset, "○"
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
		return p.runningColor + frame + ansiReset
	case ToolCompleted:
		return p.successColor + "✓" + ansiReset
	case ToolFailed:
		return p.errorColor + "✗" + ansiReset
	default:
		return "·"
	}
}

func (p *SubagentPanel) renderFooterLine() string {
	statusText := "Done in"
	if p.status == SubagentRunning {
		statusText = "Working…"
	}

	parts := []string{statusText, formatSubagentDuration(p.elapsed)}
	if p.modelName != "" {
		parts = append(parts, p.modelName)
	}
	if p.tokenCount != "" {
		parts = append(parts, p.tokenCount)
	}
	if p.cost != "" {
		parts = append(parts, p.cost)
	}

	if len(parts) <= 2 {
		return " " + p.footerColor + strings.Join(parts, " ") + ansiReset
	}

	left := strings.Join(parts[:2], " ")
	right := strings.Join(parts[2:], " · ")
	return fmt.Sprintf(" %s%s %s%s%s", p.footerColor, left, right, ansiReset, ansiReset)
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

func (p *SubagentPanel) fitToWidth(line string) string {
	if p.width <= 0 {
		return line
	}
	if len(stripANSI(line)) <= p.width {
		return line
	}
	if p.width <= 3 {
		return truncateANSI(line, p.width)
	}
	return truncateANSI(line, p.width-3) + "..."
}

func (p *SubagentPanel) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}

	accent := ansiColorFromHex(tokens.Accent)
	foreground := ansiColorFromHex(tokens.Color)
	if accent != "" {
		p.runningColor = accent
		p.successColor = accent
		p.connectorColor = accent
	}
	if foreground != "" {
		p.errorColor = foreground
		p.abortedColor = foreground
		p.footerColor = foreground
	}
}
