package display

import (
	"fmt"
	"math"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ScrollStatusBar renders a compact line range indicator and mini scrollbar.
type ScrollStatusBar struct {
	totalLines   int
	visibleStart int
	visibleEnd   int
	width        int

	focused      bool
	designTokens *design.DesignTokens
}

// ScrollStatusBarOption configures a ScrollStatusBar.
type ScrollStatusBarOption func(*ScrollStatusBar)

// WithScrollStatusBarTotalLines sets the total document line count.
func WithScrollStatusBarTotalLines(total int) ScrollStatusBarOption {
	return func(s *ScrollStatusBar) {
		if total >= 0 {
			s.totalLines = total
		}
	}
}

// WithScrollStatusBarVisibleStart sets the first visible line (1-based).
func WithScrollStatusBarVisibleStart(start int) ScrollStatusBarOption {
	return func(s *ScrollStatusBar) {
		s.visibleStart = start
	}
}

// WithScrollStatusBarVisibleEnd sets the last visible line (1-based).
func WithScrollStatusBarVisibleEnd(end int) ScrollStatusBarOption {
	return func(s *ScrollStatusBar) {
		s.visibleEnd = end
	}
}

// WithScrollStatusBarWidth sets the rendered width.
func WithScrollStatusBarWidth(width int) ScrollStatusBarOption {
	return func(s *ScrollStatusBar) {
		if width >= 0 {
			s.width = width
		}
	}
}

// WithScrollStatusBarDesignTokens applies design-system tokens.
func WithScrollStatusBarDesignTokens(tokens *design.DesignTokens) ScrollStatusBarOption {
	return func(s *ScrollStatusBar) {
		if tokens != nil {
			s.designTokens = tokens
		}
	}
}

// NewScrollStatusBar creates a ScrollStatusBar with defaults.
func NewScrollStatusBar(opts ...ScrollStatusBarOption) *ScrollStatusBar {
	s := &ScrollStatusBar{
		totalLines:   0,
		visibleStart: 0,
		visibleEnd:   0,
		width:        80,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(s)
	}

	s.clampPosition()
	return s
}

// SetPosition updates the visible range and total line count.
func (s *ScrollStatusBar) SetPosition(start, end, total int) {
	s.visibleStart = start
	s.visibleEnd = end
	s.totalLines = total
	s.clampPosition()
}

// Init initializes the component.
func (s *ScrollStatusBar) Init() tea.Cmd { return nil }

// Update handles Bubble Tea messages.
func (s *ScrollStatusBar) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
	}
	return s, nil
}

// View renders the status text and compact scrollbar.
func (s *ScrollStatusBar) View() string {
	if s.width <= 0 {
		return ""
	}

	start, end, total := s.normalizedPosition()
	left := fmt.Sprintf("Lines %d-%d of %d", start, end, total)

	available := s.width - style.StringWidth(left)
	if available <= 0 {
		return style.Truncate(left, s.width, "…")
	}

	scrollbarWidth := available
	if scrollbarWidth > 10 {
		scrollbarWidth = 10
	}
	if scrollbarWidth < 3 {
		return style.Truncate(left, s.width, "…")
	}

	right := s.renderMiniScrollbar(scrollbarWidth, start, end, total)
	gap := s.width - style.StringWidth(left) - style.StringWidth(right)
	if gap < 1 {
		leftMax := s.width - style.StringWidth(right) - 1
		if leftMax < 0 {
			leftMax = 0
		}
		left = style.Truncate(left, leftMax, "…")
		gap = s.width - style.StringWidth(left) - style.StringWidth(right)
		if gap < 1 {
			gap = 1
		}
	}

	line := left + strings.Repeat(" ", gap) + right
	if style.StringWidth(line) > s.width {
		line = style.Truncate(line, s.width, "…")
	}
	return line
}

// Focus marks the component as focused.
func (s *ScrollStatusBar) Focus() { s.focused = true }

// Blur marks the component as unfocused.
func (s *ScrollStatusBar) Blur() { s.focused = false }

// Focused reports whether the component is focused.
func (s *ScrollStatusBar) Focused() bool { return s.focused }

func (s *ScrollStatusBar) clampPosition() {
	if s.totalLines < 0 {
		s.totalLines = 0
	}

	if s.totalLines == 0 {
		s.visibleStart = 0
		s.visibleEnd = 0
		return
	}

	if s.visibleStart < 1 {
		s.visibleStart = 1
	}
	if s.visibleStart > s.totalLines {
		s.visibleStart = s.totalLines
	}

	if s.visibleEnd < s.visibleStart {
		s.visibleEnd = s.visibleStart
	}
	if s.visibleEnd > s.totalLines {
		s.visibleEnd = s.totalLines
	}
}

func (s *ScrollStatusBar) normalizedPosition() (start, end, total int) {
	s.clampPosition()
	return s.visibleStart, s.visibleEnd, s.totalLines
}

func (s *ScrollStatusBar) renderMiniScrollbar(width, start, end, total int) string {
	if width <= 0 {
		return ""
	}
	if width == 1 {
		return "▲"
	}
	if width == 2 {
		return "▲▼"
	}

	trackWidth := width - 2
	visible := end - start + 1
	if visible < 0 {
		visible = 0
	}

	thumbSize := trackWidth
	if total > 0 && visible > 0 {
		ratio := float64(visible) / float64(total)
		thumbSize = int(math.Round(ratio * float64(trackWidth)))
	}
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > trackWidth {
		thumbSize = trackWidth
	}

	thumbStart := 0
	maxThumbStart := trackWidth - thumbSize
	if total > visible && total > 0 && visible > 0 && maxThumbStart > 0 {
		ratio := float64(start-1) / float64(total-visible)
		thumbStart = int(math.Round(ratio * float64(maxThumbStart)))
	}
	if thumbStart < 0 {
		thumbStart = 0
	}
	if thumbStart > maxThumbStart {
		thumbStart = maxThumbStart
	}

	track := make([]rune, trackWidth)
	for i := range track {
		track[i] = '░'
	}
	for i := thumbStart; i < thumbStart+thumbSize && i < len(track); i++ {
		track[i] = '▓'
	}

	return "▲" + string(track) + "▼"
}

var _ tui.Component = (*ScrollStatusBar)(nil)
