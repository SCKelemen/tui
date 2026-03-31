package display

import (
	"fmt"
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/spinner"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ToolUseLoaderOption configures a ToolUseLoader.
type ToolUseLoaderOption func(*ToolUseLoader)

// WithToolUseLoaderWidth sets preferred render width.
func WithToolUseLoaderWidth(width int) ToolUseLoaderOption {
	return func(t *ToolUseLoader) {
		if width > 0 {
			t.width = width
		}
	}
}

// WithToolUseLoaderStatus sets status text.
func WithToolUseLoaderStatus(status string) ToolUseLoaderOption {
	return func(t *ToolUseLoader) {
		t.status = strings.TrimSpace(status)
	}
}

// WithToolUseLoaderElapsedTime sets elapsed execution duration.
func WithToolUseLoaderElapsedTime(elapsed time.Duration) ToolUseLoaderOption {
	return func(t *ToolUseLoader) {
		if elapsed >= 0 {
			t.elapsed = elapsed
		}
	}
}

// WithToolUseLoaderDesignTokens applies design-system tokens.
func WithToolUseLoaderDesignTokens(tokens *design.DesignTokens) ToolUseLoaderOption {
	return func(t *ToolUseLoader) {
		if tokens != nil {
			t.designTokens = tokens
		}
	}
}

// ToolUseLoaderTickMsg drives loader animation.
type ToolUseLoaderTickMsg struct{}

// ToolUseLoader renders a loading line for tool execution.
type ToolUseLoader struct {
	toolName     string
	status       string
	elapsed      time.Duration
	width        int
	windowWidth  int
	focused      bool
	designTokens *design.DesignTokens

	spinner spinner.Spinner
	frame   int
	phase   int
}

// NewToolUseLoader creates a new ToolUseLoader.
func NewToolUseLoader(toolName string, opts ...ToolUseLoaderOption) *ToolUseLoader {
	t := &ToolUseLoader{
		toolName:     strings.TrimSpace(toolName),
		status:       "running",
		elapsed:      0,
		width:        0,
		designTokens: design.DefaultTheme(),
		spinner:      spinner.Braille,
		frame:        0,
		phase:        0,
	}
	for _, opt := range opts {
		opt(t)
	}
	if t.toolName == "" {
		t.toolName = "tool"
	}
	return t
}

// Init initializes animation.
func (t *ToolUseLoader) Init() tea.Cmd {
	return t.tick()
}

// Update handles animation and resize events.
func (t *ToolUseLoader) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.windowWidth = msg.Width
		return t, nil
	case ToolUseLoaderTickMsg:
		t.frame = (t.frame + 1) % maxInt(t.spinner.FrameCount(), 1)
		t.phase = (t.phase + 1) % 8
		return t, t.tick()
	}
	return t, nil
}

// View renders spinner, shimmering tool name, status, and elapsed time.
func (t *ToolUseLoader) View() string {
	frame := t.spinner.GetFrame(t.frame)
	tool := t.shimmerToolName()
	status := strings.TrimSpace(t.status)
	if status == "" {
		status = "running"
	}
	elapsed := formatToolUseDuration(t.elapsed)
	if elapsed == "" {
		elapsed = "0ms"
	}

	line := fmt.Sprintf("[%s] %s %s  %s", frame, tool, status, elapsed)
	width := t.effectiveWidth()
	if width > 0 {
		line = style.Pad(style.Truncate(line, width, "…"), width)
	}
	return line
}

// Focus marks this component focused.
func (t *ToolUseLoader) Focus() { t.focused = true }

// Blur marks this component unfocused.
func (t *ToolUseLoader) Blur() { t.focused = false }

// Focused reports whether this component is focused.
func (t *ToolUseLoader) Focused() bool { return t.focused }

func (t *ToolUseLoader) effectiveWidth() int {
	if t.width > 0 {
		return t.width
	}
	return t.windowWidth
}

func (t *ToolUseLoader) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg {
		return ToolUseLoaderTickMsg{}
	})
}

func (t *ToolUseLoader) shimmerToolName() string {
	runes := []rune(t.toolName)
	if len(runes) == 0 {
		return t.toolName
	}
	index := t.phase % len(runes)

	var b strings.Builder
	for i, r := range runes {
		if i == index {
			b.WriteString(style.ANSIBold)
			b.WriteString(string(r))
			b.WriteString(style.ANSIReset)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

var _ tui.Component = (*ToolUseLoader)(nil)
