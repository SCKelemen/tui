package tui

import (
	"math"
	"strings"

	design "github.com/SCKelemen/design-system"
	tea "github.com/charmbracelet/bubbletea"
)

// ScrollBarOrientation controls the direction a scrollbar renders.
type ScrollBarOrientation int

const (
	ScrollBarHorizontal ScrollBarOrientation = iota
	ScrollBarVertical
)

// ScrollBar renders a standalone horizontal or vertical scrollbar.
type ScrollBar struct {
	orientation ScrollBarOrientation
	total       int
	visible     int
	offset      int
	width       int
	height      int

	trackChar rune
	thumbChar rune
	prevArrow rune
	nextArrow rune

	trackColor string
	thumbColor string
	arrowColor string

	designTokens *design.DesignTokens
	focused      bool
}

// ScrollBarOption configures a ScrollBar.
type ScrollBarOption func(*ScrollBar)

// WithScrollBarOrientation sets the scrollbar orientation.
func WithScrollBarOrientation(orientation ScrollBarOrientation) ScrollBarOption {
	return func(sb *ScrollBar) {
		sb.orientation = orientation
		sb.applyOrientationDefaults()
	}
}

// WithScrollBarSize sets the rendered width and height.
func WithScrollBarSize(width, height int) ScrollBarOption {
	return func(sb *ScrollBar) {
		if width > 0 {
			sb.width = width
		}
		if height > 0 {
			sb.height = height
		}
	}
}

// WithScrollBarMetrics sets the total content size, visible size, and offset.
func WithScrollBarMetrics(total, visible, offset int) ScrollBarOption {
	return func(sb *ScrollBar) {
		sb.total = total
		sb.visible = visible
		sb.offset = offset
		sb.normalizeMetrics()
	}
}

// WithScrollBarDesignTokens applies design-system colors.
func WithScrollBarDesignTokens(tokens *design.DesignTokens) ScrollBarOption {
	return func(sb *ScrollBar) {
		sb.applyDesignTokens(tokens)
	}
}

// WithScrollBarTheme applies a named design-system theme.
func WithScrollBarTheme(theme string) ScrollBarOption {
	return func(sb *ScrollBar) {
		sb.applyDesignTokens(designTokensForTheme(theme))
	}
}

// NewScrollBar creates a new standalone scrollbar.
func NewScrollBar(opts ...ScrollBarOption) *ScrollBar {
	sb := &ScrollBar{
		orientation:  ScrollBarVertical,
		total:        1,
		visible:      1,
		offset:       0,
		width:        10,
		height:       10,
		trackChar:    '░',
		thumbChar:    '█',
		trackColor:   ansiDim,
		thumbColor:   ansiGreen,
		arrowColor:   ansiWhite,
		designTokens: design.DefaultTheme(),
	}

	sb.applyOrientationDefaults()
	sb.applyDesignTokens(sb.designTokens)
	for _, opt := range opts {
		opt(sb)
	}
	sb.normalizeMetrics()
	return sb
}

// Init initializes the scrollbar.
func (sb *ScrollBar) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (sb *ScrollBar) Update(msg tea.Msg) (Component, tea.Cmd) {
	return sb, nil
}

// View renders the scrollbar in its current orientation.
func (sb *ScrollBar) View() string {
	sb.normalizeMetrics()
	if sb.orientation == ScrollBarHorizontal {
		return sb.horizontalView()
	}
	return sb.verticalView()
}

// Focus marks the scrollbar as focused.
func (sb *ScrollBar) Focus() {
	sb.focused = true
}

// Blur marks the scrollbar as unfocused.
func (sb *ScrollBar) Blur() {
	sb.focused = false
}

// Focused reports whether the scrollbar is focused.
func (sb *ScrollBar) Focused() bool {
	return sb.focused
}

// SetPosition sets the scrollbar offset.
func (sb *ScrollBar) SetPosition(offset int) {
	sb.offset = offset
	sb.normalizeMetrics()
}

// SetTotal sets the total scrollable size.
func (sb *ScrollBar) SetTotal(total int) {
	sb.total = total
	sb.normalizeMetrics()
}

// SetVisible sets the visible viewport size.
func (sb *ScrollBar) SetVisible(visible int) {
	sb.visible = visible
	sb.normalizeMetrics()
}

// SetWidth sets the horizontal rendered width.
func (sb *ScrollBar) SetWidth(width int) {
	if width > 0 {
		sb.width = width
	}
}

// SetHeight sets the vertical rendered height.
func (sb *ScrollBar) SetHeight(height int) {
	if height > 0 {
		sb.height = height
	}
}

// SetOrientation sets the scrollbar orientation.
func (sb *ScrollBar) SetOrientation(orientation ScrollBarOrientation) {
	sb.orientation = orientation
	sb.applyOrientationDefaults()
}

func (sb *ScrollBar) horizontalView() string {
	trackLength := max(0, sb.width-2)
	thumbStart, thumbLength := sb.thumbRange(trackLength)

	if sb.width <= 0 {
		return ""
	}
	if sb.width == 1 {
		return sb.renderStyledRune(sb.prevArrow, sb.arrowColor)
	}

	var b strings.Builder
	b.WriteString(sb.renderStyledRune(sb.prevArrow, sb.arrowColor))
	for i := 0; i < trackLength; i++ {
		if i >= thumbStart && i < thumbStart+thumbLength {
			b.WriteString(sb.renderStyledRune(sb.thumbChar, sb.thumbColor))
			continue
		}
		b.WriteString(sb.renderStyledRune(sb.trackChar, sb.trackColor))
	}
	b.WriteString(sb.renderStyledRune(sb.nextArrow, sb.arrowColor))
	return b.String()
}

func (sb *ScrollBar) verticalView() string {
	trackLength := max(0, sb.height-2)
	thumbStart, thumbLength := sb.thumbRange(trackLength)

	if sb.height <= 0 {
		return ""
	}
	if sb.height == 1 {
		return sb.renderStyledRune(sb.prevArrow, sb.arrowColor)
	}

	lines := make([]string, 0, sb.height)
	lines = append(lines, sb.renderStyledRune(sb.prevArrow, sb.arrowColor))
	for i := 0; i < trackLength; i++ {
		if i >= thumbStart && i < thumbStart+thumbLength {
			lines = append(lines, sb.renderStyledRune(sb.thumbChar, sb.thumbColor))
			continue
		}
		lines = append(lines, sb.renderStyledRune(sb.trackChar, sb.trackColor))
	}
	lines = append(lines, sb.renderStyledRune(sb.nextArrow, sb.arrowColor))
	return strings.Join(lines, "\n")
}

func (sb *ScrollBar) thumbRange(trackLength int) (int, int) {
	if trackLength <= 0 {
		return 0, 0
	}

	total := max(1, sb.total)
	visible := min(max(0, sb.visible), total)
	if visible >= total {
		return 0, trackLength
	}

	thumbLength := int(math.Round(float64(visible) / float64(total) * float64(trackLength)))
	if thumbLength < 1 {
		thumbLength = 1
	}
	if thumbLength > trackLength {
		thumbLength = trackLength
	}

	maxOffset := max(0, total-visible)
	if maxOffset == 0 {
		return 0, thumbLength
	}

	thumbStart := int(math.Round(float64(sb.offset) / float64(maxOffset) * float64(trackLength-thumbLength)))
	if thumbStart < 0 {
		thumbStart = 0
	}
	if thumbStart > trackLength-thumbLength {
		thumbStart = trackLength - thumbLength
	}
	return thumbStart, thumbLength
}

func (sb *ScrollBar) normalizeMetrics() {
	if sb.total < 0 {
		sb.total = 0
	}
	if sb.visible < 0 {
		sb.visible = 0
	}
	if sb.total == 0 {
		sb.visible = 0
		sb.offset = 0
		return
	}
	if sb.visible > sb.total {
		sb.visible = sb.total
	}
	maxOffset := max(0, sb.total-sb.visible)
	if sb.offset < 0 {
		sb.offset = 0
	}
	if sb.offset > maxOffset {
		sb.offset = maxOffset
	}
}

func (sb *ScrollBar) applyOrientationDefaults() {
	if sb.orientation == ScrollBarHorizontal {
		sb.prevArrow = '◀'
		sb.nextArrow = '▶'
		return
	}
	sb.prevArrow = '▲'
	sb.nextArrow = '▼'
}

func (sb *ScrollBar) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}
	sb.designTokens = tokens
	if color := ansiColorFromHex(tokens.Color); color != "" {
		sb.trackColor = color
		sb.arrowColor = color
	}
	if accent := ansiColorFromHex(tokens.Accent); accent != "" {
		sb.thumbColor = accent
	}
}
func (sb *ScrollBar) renderStyledRune(r rune, color string) string {
	if color == "" {
		return string(r)
	}
	return color + string(r) + ansiReset
}
