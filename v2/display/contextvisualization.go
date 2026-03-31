package display

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// SegmentType identifies the source/type of context tokens.
type SegmentType int

const (
	SegmentSystem SegmentType = iota
	SegmentUser
	SegmentAssistant
	SegmentTool
	SegmentFree
)

// ContextSegment describes one portion of context window usage.
type ContextSegment struct {
	Label  string
	Tokens int
	Type   SegmentType
}

// ContextVisualizationOption configures a ContextVisualization.
type ContextVisualizationOption func(*ContextVisualization)

// WithContextVisualizationWidth sets preferred render width.
func WithContextVisualizationWidth(width int) ContextVisualizationOption {
	return func(c *ContextVisualization) {
		if width > 0 {
			c.width = width
		}
	}
}

// WithContextVisualizationDesignTokens applies design-system tokens.
func WithContextVisualizationDesignTokens(tokens *design.DesignTokens) ContextVisualizationOption {
	return func(c *ContextVisualization) {
		if tokens != nil {
			c.designTokens = tokens
		}
	}
}

// ContextVisualization renders context window utilization as a segmented bar.
type ContextVisualization struct {
	segments     []ContextSegment
	maxTokens    int
	width        int
	windowWidth  int
	focused      bool
	designTokens *design.DesignTokens
}

// NewContextVisualization creates a new ContextVisualization.
func NewContextVisualization(segments []ContextSegment, maxTokens int, opts ...ContextVisualizationOption) *ContextVisualization {
	c := &ContextVisualization{
		segments:     append([]ContextSegment(nil), segments...),
		maxTokens:    maxInt(maxTokens, 1),
		width:        40,
		designTokens: design.DefaultTheme(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Init initializes the component.
func (c *ContextVisualization) Init() tea.Cmd { return nil }

// Update handles resize messages.
func (c *ContextVisualization) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.windowWidth = msg.Width
		return c, nil
	}
	return c, nil
}

// View renders a segmented horizontal bar and labeled percentages.
func (c *ContextVisualization) View() string {
	if len(c.segments) == 0 {
		return ""
	}

	width := c.effectiveWidth()
	if width <= 0 {
		width = 40
	}

	bar := c.renderBar(width)
	legend := c.renderLegend(width)

	parts := []string{bar}
	parts = append(parts, legend...)
	return strings.Join(parts, "\n")
}

// Focus marks this component focused.
func (c *ContextVisualization) Focus() { c.focused = true }

// Blur marks this component unfocused.
func (c *ContextVisualization) Blur() { c.focused = false }

// Focused reports whether this component is focused.
func (c *ContextVisualization) Focused() bool { return c.focused }

func (c *ContextVisualization) effectiveWidth() int {
	if c.width > 0 {
		return c.width
	}
	return c.windowWidth
}

func (c *ContextVisualization) renderBar(width int) string {
	remaining := width
	parts := make([]string, 0, len(c.segments))

	for i, seg := range c.segments {
		w := seg.Tokens * width / c.maxTokens
		if i == len(c.segments)-1 {
			w = remaining
		}
		if w > remaining {
			w = remaining
		}
		if w < 0 {
			w = 0
		}
		remaining -= w
		if w == 0 {
			continue
		}
		parts = append(parts, c.segmentColor(seg.Type)+strings.Repeat("█", w)+style.ANSIReset)
	}
	if remaining > 0 {
		parts = append(parts, style.ANSIDim+strings.Repeat("░", remaining)+style.ANSIReset)
	}
	return strings.Join(parts, "")
}

func (c *ContextVisualization) renderLegend(width int) []string {
	lines := make([]string, 0, len(c.segments))
	for _, seg := range c.segments {
		pct := float64(seg.Tokens) / float64(c.maxTokens) * 100
		line := fmt.Sprintf("%s■%s %s %d (%.1f%%)", c.segmentColor(seg.Type), style.ANSIReset, seg.Label, seg.Tokens, pct)
		if width > 0 {
			line = style.Pad(style.Truncate(line, width, "…"), width)
		}
		lines = append(lines, line)
	}
	return lines
}

func (c *ContextVisualization) segmentColor(t SegmentType) string {
	if c.designTokens != nil {
		switch t {
		case SegmentSystem:
			if fg := style.Fg(c.designTokens.Accent); fg != "" {
				return fg
			}
		case SegmentUser:
			if fg := style.Fg(c.designTokens.SuccessBright); fg != "" {
				return fg
			}
		case SegmentAssistant:
			if fg := style.Fg(c.designTokens.PendingColor); fg != "" {
				return fg
			}
		case SegmentTool:
			if fg := style.Fg(c.designTokens.ErrorBright); fg != "" {
				return fg
			}
		case SegmentFree:
			if fg := style.Fg(c.designTokens.MutedColor); fg != "" {
				return fg
			}
		}
	}

	switch t {
	case SegmentSystem:
		return style.ANSICyan
	case SegmentUser:
		return style.ANSIGreen
	case SegmentAssistant:
		return style.ANSIYellow
	case SegmentTool:
		return style.ANSIRed
	case SegmentFree:
		return style.ANSIDim
	default:
		return style.ANSIWhite
	}
}

var _ tui.Component = (*ContextVisualization)(nil)
