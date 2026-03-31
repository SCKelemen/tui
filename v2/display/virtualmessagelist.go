package display

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// VirtualItem is an item rendered by VirtualMessageList.
type VirtualItem struct {
	ID     string
	Height int
	Render func(width int) string
}

// VirtualMessageListOption configures a VirtualMessageList.
type VirtualMessageListOption func(*VirtualMessageList)

// WithVirtualListWidth sets a fixed list width.
func WithVirtualListWidth(width int) VirtualMessageListOption {
	return func(v *VirtualMessageList) {
		if width >= 0 {
			v.width = width
		}
	}
}

// WithVirtualListHeight sets a fixed viewport height.
func WithVirtualListHeight(height int) VirtualMessageListOption {
	return func(v *VirtualMessageList) {
		if height >= 0 {
			v.height = height
		}
	}
}

// WithVirtualListStickyScroll enables/disables sticky scroll-to-bottom behavior.
func WithVirtualListStickyScroll(sticky bool) VirtualMessageListOption {
	return func(v *VirtualMessageList) {
		v.stickyScroll = sticky
	}
}

// WithVirtualListDesignTokens applies design-system tokens.
func WithVirtualListDesignTokens(tokens *design.DesignTokens) VirtualMessageListOption {
	return func(v *VirtualMessageList) {
		if tokens != nil {
			v.designTokens = tokens
		}
	}
}

// VirtualMessageList efficiently renders only items visible in the viewport.
type VirtualMessageList struct {
	items []VirtualItem

	// required state
	scrollOffset   int
	viewportHeight int
	totalHeight    int

	width        int
	height       int
	windowWidth  int
	windowHeight int
	focused      bool
	stickyScroll bool
	designTokens *design.DesignTokens

	itemStarts []int
}

// NewVirtualMessageList creates a new virtualized message list.
func NewVirtualMessageList(items []VirtualItem, opts ...VirtualMessageListOption) *VirtualMessageList {
	v := &VirtualMessageList{
		items:        append([]VirtualItem(nil), items...),
		stickyScroll: true,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(v)
	}

	v.rebuildOffsets()
	v.viewportHeight = v.effectiveHeight()
	v.clampScrollOffset()
	return v
}

// Init initializes the component.
func (v *VirtualMessageList) Init() tea.Cmd {
	return nil
}

// Update handles keyboard navigation and viewport updates.
func (v *VirtualMessageList) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		wasAtBottom := v.isAtBottom()
		v.windowWidth = msg.Width
		v.windowHeight = msg.Height
		v.viewportHeight = v.effectiveHeight()
		if wasAtBottom && v.stickyScroll {
			v.ScrollToBottom()
		} else {
			v.clampScrollOffset()
		}
		return v, nil
	case tea.KeyMsg:
		if !v.focused {
			return v, nil
		}

		pageStep := v.viewportHeight - 1
		if pageStep < 1 {
			pageStep = 1
		}

		switch msg.String() {
		case "up", "k":
			v.scrollBy(-1)
			return v, nil
		case "down", "j":
			v.scrollBy(1)
			return v, nil
		case "pgup", "ctrl+u":
			v.scrollBy(-pageStep)
			return v, nil
		case "pgdown", "ctrl+d":
			v.scrollBy(pageStep)
			return v, nil
		case "home", "g":
			v.scrollOffset = 0
			v.clampScrollOffset()
			return v, nil
		case "end", "G":
			v.ScrollToBottom()
			return v, nil
		}
	}

	return v, nil
}

// View renders only items intersecting the viewport.
func (v *VirtualMessageList) View() string {
	width := v.effectiveWidth()
	if width <= 0 {
		return ""
	}

	v.viewportHeight = v.effectiveHeight()
	v.clampScrollOffset()

	if v.viewportHeight <= 0 {
		return ""
	}

	viewportStart := v.scrollOffset
	viewportEnd := v.scrollOffset + v.viewportHeight

	rows := make([]string, 0, v.viewportHeight)
	contentWidth := width
	showScrollIndicator := width > 1
	if showScrollIndicator {
		contentWidth = width - 1
	}
	if contentWidth < 0 {
		contentWidth = 0
	}

	for i, item := range v.items {
		itemStart := v.itemStarts[i]
		itemHeight := v.normalizedItemHeight(item)
		itemEnd := itemStart + itemHeight

		if itemEnd <= viewportStart {
			continue
		}
		if itemStart >= viewportEnd {
			break
		}

		itemRows := v.renderItemRows(item, contentWidth)
		clipStart := vmlMax(itemStart, viewportStart)
		clipEnd := vmlMin(itemEnd, viewportEnd)
		localStart := clipStart - itemStart
		localEnd := localStart + (clipEnd - clipStart)

		if localStart < 0 {
			localStart = 0
		}
		if localEnd > len(itemRows) {
			localEnd = len(itemRows)
		}
		if localStart < localEnd {
			rows = append(rows, itemRows[localStart:localEnd]...)
		}

		if len(rows) >= v.viewportHeight {
			break
		}
	}

	for len(rows) < v.viewportHeight {
		rows = append(rows, strings.Repeat(" ", contentWidth))
	}

	for i := range rows {
		line := rows[i]
		line = style.Truncate(line, contentWidth, "…")
		pad := contentWidth - style.StringWidth(line)
		if pad > 0 {
			line += strings.Repeat(" ", pad)
		}
		if showScrollIndicator {
			line += v.scrollbarAtRow(i)
		}
		rows[i] = line
	}

	return strings.Join(rows, "\n")
}

// Focus marks the component as focused.
func (v *VirtualMessageList) Focus() { v.focused = true }

// Blur marks the component as unfocused.
func (v *VirtualMessageList) Blur() { v.focused = false }

// Focused reports focus state.
func (v *VirtualMessageList) Focused() bool { return v.focused }

// ScrollToBottom jumps to the end of the list.
func (v *VirtualMessageList) ScrollToBottom() {
	v.scrollOffset = v.maxScrollOffset()
}

// ScrollToItem scrolls so the target item starts at the top of the viewport.
func (v *VirtualMessageList) ScrollToItem(id string) {
	for i := range v.items {
		if v.items[i].ID == id {
			v.scrollOffset = v.itemStarts[i]
			v.clampScrollOffset()
			return
		}
	}
}

// AppendItem appends one item and keeps bottom sticky when appropriate.
func (v *VirtualMessageList) AppendItem(item VirtualItem) {
	wasAtBottom := v.isAtBottom()
	v.items = append(v.items, item)
	v.itemStarts = append(v.itemStarts, v.totalHeight)
	v.totalHeight += v.normalizedItemHeight(item)

	if v.stickyScroll && wasAtBottom {
		v.ScrollToBottom()
	} else {
		v.clampScrollOffset()
	}
}

// ItemCount returns number of items.
func (v *VirtualMessageList) ItemCount() int {
	return len(v.items)
}

func (v *VirtualMessageList) rebuildOffsets() {
	v.itemStarts = make([]int, len(v.items))
	running := 0
	for i := range v.items {
		v.itemStarts[i] = running
		running += v.normalizedItemHeight(v.items[i])
	}
	v.totalHeight = running
}

func (v *VirtualMessageList) effectiveWidth() int {
	if v.width > 0 {
		return v.width
	}
	if v.windowWidth > 0 {
		return v.windowWidth
	}
	return 0
}

func (v *VirtualMessageList) effectiveHeight() int {
	if v.height > 0 {
		return v.height
	}
	if v.windowHeight > 0 {
		return v.windowHeight
	}
	if v.viewportHeight > 0 {
		return v.viewportHeight
	}
	return 0
}
func (v *VirtualMessageList) maxScrollOffset() int {
	if v.viewportHeight <= 0 || v.totalHeight <= v.viewportHeight {
		return 0
	}
	return v.totalHeight - v.viewportHeight
}

func (v *VirtualMessageList) clampScrollOffset() {
	if v.scrollOffset < 0 {
		v.scrollOffset = 0
	}
	maxOffset := v.maxScrollOffset()
	if v.scrollOffset > maxOffset {
		v.scrollOffset = maxOffset
	}
}

func (v *VirtualMessageList) scrollBy(delta int) {
	v.scrollOffset += delta
	v.clampScrollOffset()
}

func (v *VirtualMessageList) isAtBottom() bool {
	return v.scrollOffset >= v.maxScrollOffset()
}

func (v *VirtualMessageList) normalizedItemHeight(item VirtualItem) int {
	if item.Height > 0 {
		return item.Height
	}
	return 1
}

func (v *VirtualMessageList) renderItemRows(item VirtualItem, width int) []string {
	height := v.normalizedItemHeight(item)
	if item.Render == nil {
		rows := make([]string, height)
		for i := range rows {
			rows[i] = ""
		}
		return rows
	}

	rendered := item.Render(width)
	lines := strings.Split(strings.TrimSuffix(rendered, "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}

	rows := make([]string, 0, height)
	for i := 0; i < height; i++ {
		if i < len(lines) {
			rows = append(rows, lines[i])
		} else {
			rows = append(rows, "")
		}
	}
	return rows
}

func (v *VirtualMessageList) scrollbarAtRow(row int) string {
	track := style.ANSIDim + "│" + style.ANSIReset
	thumb := style.ANSIBold + "█" + style.ANSIReset

	if v.totalHeight <= 0 || v.viewportHeight <= 0 {
		return track
	}

	thumbSize := v.viewportHeight
	thumbPos := 0
	if v.totalHeight > v.viewportHeight {
		thumbSize = (v.viewportHeight * v.viewportHeight) / v.totalHeight
		if thumbSize < 1 {
			thumbSize = 1
		}
		if thumbSize > v.viewportHeight {
			thumbSize = v.viewportHeight
		}

		maxOffset := v.maxScrollOffset()
		if maxOffset > 0 {
			thumbPos = (v.scrollOffset * (v.viewportHeight - thumbSize)) / maxOffset
		}
	}

	if row >= thumbPos && row < thumbPos+thumbSize {
		return thumb
	}
	return track
}

func vmlMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func vmlMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var _ tui.Component = (*VirtualMessageList)(nil)
