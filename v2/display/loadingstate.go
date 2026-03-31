package display

import (
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// LoadingState renders loading placeholders with shimmer animation.
type LoadingState struct {
	width       int
	height      int
	lines       int
	label       string
	focused     bool
	phase       int
	designTokens *design.DesignTokens
}

// LoadingStateOption configures a LoadingState.
type LoadingStateOption func(*LoadingState)

// WithLoadingStateWidth sets placeholder width.
func WithLoadingStateWidth(width int) LoadingStateOption {
	return func(l *LoadingState) {
		if width >= 0 {
			l.width = width
		}
	}
}

// WithLoadingStateHeight sets placeholder height.
func WithLoadingStateHeight(height int) LoadingStateOption {
	return func(l *LoadingState) {
		if height >= 0 {
			l.height = height
		}
	}
}

// WithLoadingStateLines sets placeholder line count.
func WithLoadingStateLines(lines int) LoadingStateOption {
	return func(l *LoadingState) {
		if lines >= 0 {
			l.lines = lines
		}
	}
}

// WithLoadingStateLabel sets an optional loading label.
func WithLoadingStateLabel(label string) LoadingStateOption {
	return func(l *LoadingState) { l.label = strings.TrimSpace(label) }
}

// WithLoadingStateDesignTokens applies design tokens.
func WithLoadingStateDesignTokens(tokens *design.DesignTokens) LoadingStateOption {
	return func(l *LoadingState) {
		if tokens != nil {
			l.designTokens = tokens
		}
	}
}

type loadingStateTickMsg time.Time

// NewLoadingState creates a LoadingState component.
func NewLoadingState(opts ...LoadingStateOption) *LoadingState {
	l := &LoadingState{
		width:        28,
		height:       0,
		lines:        3,
		label:        "Loading...",
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// Init initializes the component.
func (l *LoadingState) Init() tea.Cmd { return l.tick() }

// Update handles Bubble Tea messages.
func (l *LoadingState) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		if l.width == 0 {
			l.width = m.Width
		}
	case loadingStateTickMsg:
		l.phase++
		return l, l.tick()
	}
	return l, nil
}

// View renders the loading placeholders.
func (l *LoadingState) View() string {
	if l.width <= 0 {
		return ""
	}

	lineCount := l.lines
	if l.height > 0 {
		lineCount = l.height
	}
	if lineCount <= 0 {
		lineCount = 1
	}

	baseColor := style.ANSIDim
	if l.designTokens != nil {
		if c := style.Fg(l.designTokens.MutedColor); c != "" {
			baseColor = c
		}
	}

	chars := []rune{'░', '▒', '▓'}
	var b strings.Builder
	if l.label != "" {
		b.WriteString(baseColor)
		b.WriteString(l.label)
		b.WriteString(style.ANSIReset)
		b.WriteString("\n")
	}

	for i := 0; i < lineCount; i++ {
		ch := string(chars[(l.phase+i)%len(chars)])
		b.WriteString(baseColor)
		b.WriteString(strings.Repeat(ch, l.width))
		b.WriteString(style.ANSIReset)
		if i < lineCount-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}

// Focus marks the component as focused.
func (l *LoadingState) Focus() { l.focused = true }

// Blur marks the component as unfocused.
func (l *LoadingState) Blur() { l.focused = false }

// Focused reports whether the component is focused.
func (l *LoadingState) Focused() bool { return l.focused }

func (l *LoadingState) tick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(ts time.Time) tea.Msg {
		return loadingStateTickMsg(ts)
	})
}

var _ tui.Component = (*LoadingState)(nil)
