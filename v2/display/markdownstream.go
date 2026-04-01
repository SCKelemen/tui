package display

import (
	"strings"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style/design"
	tea "github.com/charmbracelet/bubbletea"
)

// MarkdownStream incrementally accepts markdown text and renders only committed lines.
//
// Commit behavior:
//   - Text is appended into an internal raw buffer.
//   - Lines are committed only after a newline is observed.
//   - Fenced code blocks are held until a closing ``` fence is committed.
//   - Finalize flushes any remaining partial line and unterminated code block.
type MarkdownStream struct {
	rawBuffer      strings.Builder
	committedLines []string
	codeLines      []string
	inCodeFence    bool
	width          int
	focused        bool
	designTokens   *design.DesignTokens
}

// MarkdownStreamOption configures a MarkdownStream instance.
type MarkdownStreamOption func(*MarkdownStream)

// WithMarkdownStreamWidth sets an optional markdown render width.
func WithMarkdownStreamWidth(width int) MarkdownStreamOption {
	return func(m *MarkdownStream) {
		if width >= 0 {
			m.width = width
		}
	}
}

// WithMarkdownStreamDesignTokens applies design tokens used for markdown styling.
func WithMarkdownStreamDesignTokens(tokens *design.DesignTokens) MarkdownStreamOption {
	return func(m *MarkdownStream) {
		if tokens != nil {
			m.designTokens = tokens
		}
	}
}

// NewMarkdownStream constructs a newline-gated incremental markdown renderer.
func NewMarkdownStream(opts ...MarkdownStreamOption) *MarkdownStream {
	m := &MarkdownStream{
		committedLines: make([]string, 0, 32),
		codeLines:      make([]string, 0, 16),
		designTokens:   design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// Init satisfies the Bubble Tea model contract.
func (m *MarkdownStream) Init() tea.Cmd { return nil }

// Update handles window-size updates for responsive wrapping.
func (m *MarkdownStream) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch t := msg.(type) {
	case tea.WindowSizeMsg:
		if m.width == 0 {
			m.width = t.Width
		}
	}
	return m, nil
}

// Append queues text into the stream and commits only complete logical lines.
func (m *MarkdownStream) Append(text string) {
	if text == "" {
		return
	}

	m.rawBuffer.WriteString(text)
	m.commitAvailableLines()
}

// Finalize flushes any pending line and unterminated fenced code block.
func (m *MarkdownStream) Finalize() {
	m.commitAvailableLines()
	remaining := strings.TrimRight(m.rawBuffer.String(), "\r")
	if remaining != "" {
		m.commitLine(remaining)
	}
	m.rawBuffer.Reset()

	if m.inCodeFence && len(m.codeLines) > 0 {
		m.committedLines = append(m.committedLines, m.codeLines...)
		m.codeLines = m.codeLines[:0]
		m.inCodeFence = false
	}
}

// View renders committed markdown content.
func (m *MarkdownStream) View() string {
	if len(m.committedLines) == 0 {
		return ""
	}

	theme := DefaultMarkdownTheme()
	if m.designTokens != nil {
		theme = MarkdownTheme{
			HeadingColor:    m.designTokens.Accent,
			BoldColor:       m.designTokens.Color,
			CodeColor:       m.designTokens.Color,
			LinkColor:       m.designTokens.Accent,
			BlockquoteColor: m.designTokens.MutedColor,
			CodeBgColor:     m.designTokens.SurfaceRaised,
			HRChar:          "─",
		}
	}

	renderer := NewMarkdown(strings.Join(m.committedLines, "\n"), WithMarkdownWidth(m.width), WithMarkdownTheme(theme))
	return renderer.View()
}

// Focus marks the component focused.
func (m *MarkdownStream) Focus() { m.focused = true }

// Blur marks the component unfocused.
func (m *MarkdownStream) Blur() { m.focused = false }

// Focused reports focus state.
func (m *MarkdownStream) Focused() bool { return m.focused }

func (m *MarkdownStream) commitAvailableLines() {
	buffer := m.rawBuffer.String()
	for {
		idx := strings.IndexByte(buffer, '\n')
		if idx < 0 {
			break
		}

		line := strings.TrimRight(buffer[:idx], "\r")
		m.commitLine(line)
		buffer = buffer[idx+1:]
	}

	m.rawBuffer.Reset()
	m.rawBuffer.WriteString(buffer)
}

func (m *MarkdownStream) commitLine(line string) {
	trimmed := strings.TrimSpace(line)
	isFence := strings.HasPrefix(trimmed, "```")

	if m.inCodeFence {
		m.codeLines = append(m.codeLines, line)
		if isFence {
			m.committedLines = append(m.committedLines, m.codeLines...)
			m.codeLines = m.codeLines[:0]
			m.inCodeFence = false
		}
		return
	}

	if isFence {
		m.inCodeFence = true
		m.codeLines = append(m.codeLines, line)
		return
	}

	m.committedLines = append(m.committedLines, line)
}

var _ tui.Component = (*MarkdownStream)(nil)
