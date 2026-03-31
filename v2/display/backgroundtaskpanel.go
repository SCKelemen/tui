package display

import (
	"fmt"
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// TaskType identifies background task category.
type TaskType string

const (
	TaskTypeShell  TaskType = "shell"
	TaskTypeAgent  TaskType = "agent"
	TaskTypeRemote TaskType = "remote"
)

// BackgroundTaskPanelOption configures a BackgroundTaskPanel.
type BackgroundTaskPanelOption func(*BackgroundTaskPanel)

// WithBackgroundTaskPanelWidth sets fixed width.
func WithBackgroundTaskPanelWidth(width int) BackgroundTaskPanelOption {
	return func(p *BackgroundTaskPanel) {
		if width > 0 {
			p.width = width
		}
	}
}

// WithBackgroundTaskPanelStatus sets display status.
func WithBackgroundTaskPanelStatus(status string) BackgroundTaskPanelOption {
	return func(p *BackgroundTaskPanel) {
		p.status = strings.TrimSpace(status)
	}
}

// WithBackgroundTaskPanelProgress sets progress [0,1].
func WithBackgroundTaskPanelProgress(progress float64) BackgroundTaskPanelOption {
	return func(p *BackgroundTaskPanel) {
		if progress < 0 {
			progress = 0
		}
		if progress > 1 {
			progress = 1
		}
		p.progress = progress
	}
}

// WithBackgroundTaskPanelDuration sets elapsed duration.
func WithBackgroundTaskPanelDuration(d time.Duration) BackgroundTaskPanelOption {
	return func(p *BackgroundTaskPanel) {
		if d < 0 {
			d = 0
		}
		p.duration = d
	}
}

// WithBackgroundTaskPanelCollapsed sets collapsed state.
func WithBackgroundTaskPanelCollapsed(collapsed bool) BackgroundTaskPanelOption {
	return func(p *BackgroundTaskPanel) {
		p.collapsed = collapsed
	}
}

// WithBackgroundTaskPanelDesignTokens applies design-system colors.
func WithBackgroundTaskPanelDesignTokens(tokens *design.DesignTokens) BackgroundTaskPanelOption {
	return func(p *BackgroundTaskPanel) {
		if tokens == nil {
			return
		}
		p.colors = backgroundTaskPanelColorsFromTokens(tokens)
	}
}

// BackgroundTaskPanel renders running background task status with progress.
type BackgroundTaskPanel struct {
	name     string
	taskType TaskType
	status   string
	progress float64
	duration time.Duration
	collapsed bool

	width       int
	windowWidth int
	focused     bool
	colors      backgroundTaskPanelColors
}

type backgroundTaskPanelColors struct {
	text     string
	muted    string
	accent   string
	success  string
	pending  string
	focusBg  string
}

func defaultBackgroundTaskPanelColors() backgroundTaskPanelColors {
	return backgroundTaskPanelColors{
		text:    style.ANSIWhite,
		muted:   style.ANSIDim,
		accent:  style.ANSICyan,
		success: style.ANSIGreen,
		pending: style.ANSIYellow,
		focusBg: style.ANSIInverse,
	}
}

func backgroundTaskPanelColorsFromTokens(tokens *design.DesignTokens) backgroundTaskPanelColors {
	c := defaultBackgroundTaskPanelColors()
	if tokens == nil {
		return c
	}
	if v := style.Fg(tokens.Color); v != "" {
		c.text = v
	}
	if v := style.Fg(tokens.MutedColor); v != "" {
		c.muted = v
	}
	if v := style.Fg(tokens.Accent); v != "" {
		c.accent = v
	}
	if v := style.Fg(tokens.SuccessBright); v != "" {
		c.success = v
	}
	if v := style.Fg(tokens.PendingColor); v != "" {
		c.pending = v
	}
	if v := style.Bg(tokens.SurfaceRaised); v != "" {
		c.focusBg = v
	}
	return c
}

// NewBackgroundTaskPanel creates a BackgroundTaskPanel component.
func NewBackgroundTaskPanel(name string, taskType TaskType, opts ...BackgroundTaskPanelOption) *BackgroundTaskPanel {
	p := &BackgroundTaskPanel{
		name:      strings.TrimSpace(name),
		taskType:  taskType,
		status:    "running",
		progress:  0,
		duration:  0,
		collapsed: false,
		width:     0,
		focused:   false,
		colors:    defaultBackgroundTaskPanelColors(),
	}
	for _, opt := range opts {
		opt(p)
	}
	if p.name == "" {
		p.name = "background task"
	}
	if strings.TrimSpace(string(p.taskType)) == "" {
		p.taskType = TaskTypeAgent
	}
	return p
}

// Init initializes the component.
func (p *BackgroundTaskPanel) Init() tea.Cmd {
	return nil
}

// Update handles collapse toggle on Enter/Space.
func (p *BackgroundTaskPanel) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.windowWidth = msg.Width
		return p, nil
	case tea.KeyMsg:
		if !p.focused {
			return p, nil
		}
		switch msg.String() {
		case "enter", " ":
			p.collapsed = !p.collapsed
		}
	}
	return p, nil
}

// View renders task summary and progress.
func (p *BackgroundTaskPanel) View() string {
	width := p.renderWidth()
	line := p.renderHeader(width)
	if p.focused {
		line = p.colors.focusBg + line + style.ANSIReset
	}
	if p.collapsed {
		return line + "\n"
	}
	hint := p.colors.muted + "enter/space to collapse" + style.ANSIReset
	if width > 0 {
		hint = fitBackgroundTaskPanelLine(hint, width)
	}
	return line + "\n" + hint + "\n"
}

// Focus marks this component focused.
func (p *BackgroundTaskPanel) Focus() {
	p.focused = true
}

// Blur marks this component unfocused.
func (p *BackgroundTaskPanel) Blur() {
	p.focused = false
}

// Focused reports focus state.
func (p *BackgroundTaskPanel) Focused() bool {
	return p.focused
}

func (p *BackgroundTaskPanel) renderHeader(width int) string {
	icon := p.icon()
	typeLabel := strings.ToUpper(string(p.taskType))
	status := strings.TrimSpace(p.status)
	if status == "" {
		status = "running"
	}
	bar := backgroundTaskPanelProgressBar(p.progress, 14)
	duration := formatBackgroundTaskPanelDuration(p.duration)
	text := fmt.Sprintf("[%s] %s %s  %s  %s  %s", icon, p.name, typeLabel, status, bar, duration)
	if width > 0 {
		return fitBackgroundTaskPanelLine(text, width)
	}
	return text
}

func (p *BackgroundTaskPanel) icon() string {
	switch p.taskType {
	case TaskTypeShell:
		return p.colors.accent + "⌘" + style.ANSIReset
	case TaskTypeRemote:
		return p.colors.pending + "☁" + style.ANSIReset
	default:
		return p.colors.success + "⚙" + style.ANSIReset
	}
}

func (p *BackgroundTaskPanel) renderWidth() int {
	if p.width > 0 {
		return p.width
	}
	if p.windowWidth > 0 {
		return p.windowWidth
	}
	return 0
}

func backgroundTaskPanelProgressBar(progress float64, width int) string {
	if width < 4 {
		width = 4
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}

func formatBackgroundTaskPanelDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	if d < time.Second {
		ms := d.Milliseconds()
		if ms < 1 {
			ms = 1
		}
		return fmt.Sprintf("%dms", ms)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	m := int(d / time.Minute)
	s := int((d % time.Minute) / time.Second)
	return fmt.Sprintf("%dm%02ds", m, s)
}

func fitBackgroundTaskPanelLine(s string, width int) string {
	if width <= 0 {
		return s
	}
	t := style.Truncate(s, width, "…")
	if style.StringWidth(t) < width {
		return style.Pad(t, width)
	}
	return t
}

var _ tui.Component = (*BackgroundTaskPanel)(nil)
