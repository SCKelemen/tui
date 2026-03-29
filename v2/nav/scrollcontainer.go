package nav

import (
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ScrollContainer is a viewport wrapper for heterogeneous child components.
type ScrollContainer struct {
	children           []tui.Component
	scrollOffset       int
	viewportHeight     int
	viewportWidth      int
	totalHeight        int
	focused            bool
	focusedChild       int
	showScrollbar      bool
	scrollbarChar      string
	designTokens       *design.DesignTokens
	autoScroll         bool
	autoScrollLocked   bool
	scrollAcceleration bool
	lastScrollEventAt  time.Time
	lastScrollDir      int
	scrollStep         int
}

// ScrollContainerOption configures a ScrollContainer.
type ScrollContainerOption func(*ScrollContainer)

// WithScrollContainerDesignTokens applies design-system tokens.
func WithScrollContainerDesignTokens(tokens *design.DesignTokens) ScrollContainerOption {
	return func(sc *ScrollContainer) {
		sc.designTokens = tokens
	}
}

// WithScrollContainerTheme applies a named design-system theme.
func WithScrollContainerTheme(theme string) ScrollContainerOption {
	return func(sc *ScrollContainer) {
		sc.designTokens = style.DesignTokensForTheme(theme)
	}
}

// WithShowScrollbar toggles scrollbar rendering.
func WithShowScrollbar(show bool) ScrollContainerOption {
	return func(sc *ScrollContainer) {
		sc.showScrollbar = show
	}
}

// WithScrollbarChar sets the scrollbar thumb character.
func WithScrollbarChar(char string) ScrollContainerOption {
	return func(sc *ScrollContainer) {
		if char == "" {
			return
		}
		sc.scrollbarChar = char
	}
}

// WithAutoScroll toggles auto-scroll-to-bottom when content changes.
func WithAutoScroll(auto bool) ScrollContainerOption {
	return func(sc *ScrollContainer) {
		sc.autoScroll = auto
		if !auto {
			sc.autoScrollLocked = false
		}
	}
}

// WithScrollAcceleration toggles repeated-scroll acceleration.
func WithScrollAcceleration(enabled bool) ScrollContainerOption {
	return func(sc *ScrollContainer) {
		sc.scrollAcceleration = enabled
		if !enabled {
			sc.scrollStep = 1
			sc.lastScrollDir = 0
			sc.lastScrollEventAt = time.Time{}
		}
	}
}

// WithViewportHeight sets an initial viewport height.
func WithViewportHeight(height int) ScrollContainerOption {
	return func(sc *ScrollContainer) {
		if height < 0 {
			height = 0
		}
		sc.viewportHeight = height
	}
}

// NewScrollContainer creates a new scroll container.
func NewScrollContainer(opts ...ScrollContainerOption) *ScrollContainer {
	sc := &ScrollContainer{
		children:           []tui.Component{},
		focusedChild:       -1,
		showScrollbar:      true,
		scrollbarChar:      "█",
		autoScroll:         false,
		scrollAcceleration: true,
		scrollStep:         1,
	}
	for _, opt := range opts {
		opt(sc)
	}

	return sc
}

// Init initializes the scroll container.
func (sc *ScrollContainer) Init() tea.Cmd {
	return nil
}

// Update handles Bubble Tea messages and scroll navigation.
func (sc *ScrollContainer) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		sc.viewportWidth = msg.Width
		sc.viewportHeight = msg.Height
		sc.updateTotalHeight()
		sc.clampScrollOffset()
		if sc.IsAtBottom() {
			sc.autoScrollLocked = false
		}

	case tea.MouseMsg:
		if sc.focused {
			switch msg.Button {
			case tea.MouseButtonWheelDown:
				step := sc.nextScrollStep(1)
				sc.applyUserScroll(step)
				return sc, nil
			case tea.MouseButtonWheelUp:
				step := sc.nextScrollStep(-1)
				sc.applyUserScroll(-step)
				return sc, nil
			}
		}

	case tea.KeyMsg:
		if !sc.focused {
			break
		}

		halfStep := sc.viewportHeight / 2
		if halfStep < 1 {
			halfStep = 1
		}

		handled := true
		switch msg.String() {
		case "j", "down":
			step := sc.nextScrollStep(1)
			sc.applyUserScroll(step)
		case "k", "up":
			step := sc.nextScrollStep(-1)
			sc.applyUserScroll(-step)
		case "ctrl+d", "pgdown":
			sc.resetScrollAcceleration()
			sc.applyUserScroll(halfStep)
		case "ctrl+u", "pgup":
			sc.resetScrollAcceleration()
			sc.applyUserScroll(-halfStep)
		case "G":
			sc.resetScrollAcceleration()
			sc.ScrollToBottom()
			return sc, nil
		case "g":
			sc.resetScrollAcceleration()
			sc.ScrollToTop()
			if sc.autoScroll {
				sc.autoScrollLocked = true
			}
			return sc, nil
		case "tab":
			sc.focusNextChild()
			return sc, nil
		case "shift+tab":
			sc.focusPrevChild()
			return sc, nil
		default:
			handled = false
		}

		if handled {
			return sc, nil
		}
	}

	if sc.focusedChild >= 0 && sc.focusedChild < len(sc.children) {
		updated, cmd := sc.children[sc.focusedChild].Update(msg)
		sc.children[sc.focusedChild] = updated
		sc.updateTotalHeight()
		if sc.autoScroll && !sc.autoScrollLocked {
			sc.ScrollToBottom()
		} else {
			sc.clampScrollOffset()
			if sc.IsAtBottom() {
				sc.autoScrollLocked = false
			}
		}
		return sc, cmd
	}

	sc.updateTotalHeight()
	sc.clampScrollOffset()
	if sc.IsAtBottom() {
		sc.autoScrollLocked = false
	}
	return sc, nil
}

// View renders the visible content window, with optional scrollbar.
func (sc *ScrollContainer) View() string {
	if sc.viewportWidth <= 0 {
		return ""
	}

	allLines := sc.renderedLines()
	sc.totalHeight = len(allLines)
	sc.clampScrollOffset()

	start := sc.scrollOffset
	if start < 0 {
		start = 0
	}
	if start > len(allLines) {
		start = len(allLines)
	}

	end := len(allLines)
	if sc.viewportHeight > 0 {
		end = start + sc.viewportHeight
		if end > len(allLines) {
			end = len(allLines)
		}
	}

	visible := allLines[start:end]
	renderHeight := len(visible)
	if sc.viewportHeight > 0 {
		renderHeight = sc.viewportHeight
	}

	contentWidth := sc.viewportWidth
	withScrollbar := sc.showScrollbar && sc.viewportWidth > 1
	if withScrollbar {
		contentWidth--
	}
	if contentWidth < 0 {
		contentWidth = 0
	}

	rows := make([]string, 0, renderHeight)
	for i := 0; i < renderHeight; i++ {
		line := ""
		if i < len(visible) {
			line = visible[i]
		}

		if contentWidth > 0 {
			line = style.Truncate(stripANSI(line), contentWidth, "…")
			line += strings.Repeat(" ", max(0, contentWidth-style.StringWidth(stripANSI(line))))
		} else {
			line = ""
		}
		if withScrollbar {
			line += sc.scrollbarAtRow(i, renderHeight)
		}

		rows = append(rows, line)
	}

	return strings.Join(rows, "\n")
}

// Focus marks the container as focused.
func (sc *ScrollContainer) Focus() {
	sc.focused = true
	if len(sc.children) > 0 && sc.focusedChild == -1 {
		sc.focusedChild = 0
	}
	if sc.focusedChild >= 0 && sc.focusedChild < len(sc.children) {
		sc.children[sc.focusedChild].Focus()
	}
}

// Blur marks the container as unfocused.
func (sc *ScrollContainer) Blur() {
	sc.focused = false
	if sc.focusedChild >= 0 && sc.focusedChild < len(sc.children) {
		sc.children[sc.focusedChild].Blur()
	}
}

// Focused reports focus state.
func (sc *ScrollContainer) Focused() bool {
	return sc.focused
}

// AddChild appends a child component.
func (sc *ScrollContainer) AddChild(c tui.Component) {
	if c == nil {
		return
	}
	sc.children = append(sc.children, c)
	if sc.focused && sc.focusedChild == -1 {
		sc.focusedChild = len(sc.children) - 1
		sc.children[sc.focusedChild].Focus()
	}
	sc.updateTotalHeight()
	if sc.autoScroll && !sc.autoScrollLocked {
		sc.ScrollToBottom()
	} else {
		sc.clampScrollOffset()
		if sc.IsAtBottom() {
			sc.autoScrollLocked = false
		}
	}
}

// RemoveChild removes a child by index.
func (sc *ScrollContainer) RemoveChild(index int) {
	if index < 0 || index >= len(sc.children) {
		return
	}

	if index == sc.focusedChild {
		sc.children[index].Blur()
		sc.focusedChild = -1
	}

	sc.children = append(sc.children[:index], sc.children[index+1:]...)

	if sc.focusedChild > index {
		sc.focusedChild--
	}
	if sc.focused && len(sc.children) > 0 && sc.focusedChild == -1 {
		sc.focusedChild = min(index, len(sc.children)-1)
		sc.children[sc.focusedChild].Focus()
	}

	sc.updateTotalHeight()
	sc.clampScrollOffset()
}

// SetChildren replaces all child components.
func (sc *ScrollContainer) SetChildren(children []tui.Component) {
	if sc.focusedChild >= 0 && sc.focusedChild < len(sc.children) {
		sc.children[sc.focusedChild].Blur()
	}

	sc.children = append([]tui.Component(nil), children...)
	sc.focusedChild = -1
	if sc.focused && len(sc.children) > 0 {
		sc.focusedChild = 0
		sc.children[0].Focus()
	}

	sc.updateTotalHeight()
	if sc.autoScroll && !sc.autoScrollLocked {
		sc.ScrollToBottom()
	} else {
		sc.clampScrollOffset()
		if sc.IsAtBottom() {
			sc.autoScrollLocked = false
		}
	}
}

// GetChildren returns a copy of child components.
func (sc *ScrollContainer) GetChildren() []tui.Component {
	out := make([]tui.Component, len(sc.children))
	copy(out, sc.children)
	return out
}

// ChildCount returns the number of children.
func (sc *ScrollContainer) ChildCount() int {
	return len(sc.children)
}

// ScrollTo sets the scroll offset (clamped).
func (sc *ScrollContainer) ScrollTo(offset int) {
	sc.scrollOffset = offset
	sc.clampScrollOffset()
	if sc.IsAtBottom() {
		sc.autoScrollLocked = false
	}
}

// ScrollToBottom jumps to the end of content.
func (sc *ScrollContainer) ScrollToBottom() {
	sc.updateTotalHeight()
	sc.scrollOffset = sc.maxScrollOffset()
	sc.autoScrollLocked = false
}

// ScrollToTop jumps to the start of content.
func (sc *ScrollContainer) ScrollToTop() {
	sc.scrollOffset = 0
}

// GetScrollOffset returns the current scroll offset.
func (sc *ScrollContainer) GetScrollOffset() int {
	return sc.scrollOffset
}

// IsAtBottom reports whether the viewport is at the bottom.
func (sc *ScrollContainer) IsAtBottom() bool {
	sc.updateTotalHeight()
	return sc.scrollOffset >= sc.maxScrollOffset()
}

// IsAtTop reports whether the viewport is at the top.
func (sc *ScrollContainer) IsAtTop() bool {
	return sc.scrollOffset <= 0
}

// ScrollLocked reports whether auto-scroll is disabled by user scroll.
func (sc *ScrollContainer) ScrollLocked() bool {
	return sc.autoScroll && sc.autoScrollLocked
}

// FocusChild focuses a specific child index.
func (sc *ScrollContainer) FocusChild(index int) {
	if index < 0 || index >= len(sc.children) {
		return
	}
	if sc.focusedChild >= 0 && sc.focusedChild < len(sc.children) {
		sc.children[sc.focusedChild].Blur()
	}
	sc.focusedChild = index
	if sc.focused {
		sc.children[sc.focusedChild].Focus()
	}
}

func (sc *ScrollContainer) renderedLines() []string {
	lines := make([]string, 0, len(sc.children)*3)
	for _, child := range sc.children {
		if child == nil {
			continue
		}
		view := child.View()
		if view == "" {
			continue
		}
		parts := strings.Split(strings.TrimSuffix(view, "\n"), "\n")
		if len(parts) == 1 && parts[0] == "" {
			continue
		}
		lines = append(lines, parts...)
	}
	return lines
}

func (sc *ScrollContainer) updateTotalHeight() {
	sc.totalHeight = len(sc.renderedLines())
}

func (sc *ScrollContainer) maxScrollOffset() int {
	if sc.viewportHeight <= 0 {
		return 0
	}
	if sc.totalHeight <= sc.viewportHeight {
		return 0
	}
	return sc.totalHeight - sc.viewportHeight
}

func (sc *ScrollContainer) clampScrollOffset() {
	if sc.scrollOffset < 0 {
		sc.scrollOffset = 0
	}
	maxOffset := sc.maxScrollOffset()
	if sc.scrollOffset > maxOffset {
		sc.scrollOffset = maxOffset
	}
}

func (sc *ScrollContainer) applyUserScroll(delta int) {
	if delta == 0 {
		return
	}

	sc.scrollOffset += delta
	sc.clampScrollOffset()

	if sc.autoScroll {
		if delta < 0 {
			sc.autoScrollLocked = true
		}
		if sc.IsAtBottom() {
			sc.autoScrollLocked = false
		}
	}
}

func (sc *ScrollContainer) nextScrollStep(direction int) int {
	if direction == 0 {
		sc.resetScrollAcceleration()
		return 1
	}
	if !sc.scrollAcceleration {
		return 1
	}

	now := time.Now()
	if sc.lastScrollDir == direction && !sc.lastScrollEventAt.IsZero() && now.Sub(sc.lastScrollEventAt) <= 100*time.Millisecond {
		sc.scrollStep = min(sc.scrollStep+1, 5)
	} else {
		sc.scrollStep = 1
	}

	sc.lastScrollDir = direction
	sc.lastScrollEventAt = now
	return sc.scrollStep
}

func (sc *ScrollContainer) resetScrollAcceleration() {
	sc.scrollStep = 1
	sc.lastScrollDir = 0
	sc.lastScrollEventAt = time.Time{}
}

func (sc *ScrollContainer) focusNextChild() {
	if len(sc.children) == 0 {
		return
	}
	if sc.focusedChild >= 0 && sc.focusedChild < len(sc.children) {
		sc.children[sc.focusedChild].Blur()
	}
	sc.focusedChild = (sc.focusedChild + 1) % len(sc.children)
	sc.children[sc.focusedChild].Focus()
}

func (sc *ScrollContainer) focusPrevChild() {
	if len(sc.children) == 0 {
		return
	}
	if sc.focusedChild >= 0 && sc.focusedChild < len(sc.children) {
		sc.children[sc.focusedChild].Blur()
	}
	sc.focusedChild--
	if sc.focusedChild < 0 {
		sc.focusedChild = len(sc.children) - 1
	}
	sc.children[sc.focusedChild].Focus()
}

func (sc *ScrollContainer) scrollbarAtRow(row, renderHeight int) string {
	track := style.ANSIDim + "│" + style.ANSIReset
	thumb := style.ANSIBold + sc.scrollbarChar + style.ANSIReset
	if sc.totalHeight <= 0 || renderHeight <= 0 {
		return track
	}

	thumbSize := renderHeight
	thumbPos := 0
	if sc.totalHeight > renderHeight {
		thumbSize = (renderHeight * renderHeight) / sc.totalHeight
		if thumbSize < 1 {
			thumbSize = 1
		}
		if thumbSize > renderHeight {
			thumbSize = renderHeight
		}

		maxOffset := sc.maxScrollOffset()
		if maxOffset > 0 {
			thumbPos = (sc.scrollOffset * (renderHeight - thumbSize)) / maxOffset
		}
	}

	if row >= thumbPos && row < thumbPos+thumbSize {
		return thumb
	}
	return track
}

func stripANSI(s string) string {
	var result strings.Builder
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
		} else if inEscape {
			if r == 'm' {
				inEscape = false
			}
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func truncateANSI(s string, maxWidth int) string {
	var result strings.Builder
	visualWidth := 0
	inEscape := false

	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inEscape = true
			result.WriteByte(s[i])
		} else if inEscape {
			result.WriteByte(s[i])
			if s[i] == 'm' {
				inEscape = false
			}
		} else {
			if visualWidth >= maxWidth {
				break
			}
			result.WriteByte(s[i])
			visualWidth++
		}
	}

	return result.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Ensure compile-time interface compliance.
var _ tui.Component = (*ScrollContainer)(nil)
