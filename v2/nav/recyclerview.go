package nav

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// RecyclerAdapter adapts a data source for RecyclerView rendering.
type RecyclerAdapter[T any] interface {
	ItemCount() int
	ItemHeight(index int) int
	RenderItem(index int, item T, width int, selected bool) string
	Items() []T
}

// viewHolder stores rendered output for one item index.
type viewHolder struct {
	index    int
	rendered string
	height   int
	dirty    bool
}

// RecyclerOption configures a RecyclerView.
type RecyclerOption[T any] func(*RecyclerView[T])

// RecyclerView efficiently renders only visible items and recycles view holders.
type RecyclerView[T any] struct {
	adapter      RecyclerAdapter[T]
	viewport     int
	scrollOffset int
	selected     int
	width        int
	pool         []viewHolder
	visible      []viewHolder
	tokens       *design.DesignTokens
	focused      bool
	windowWidth  int
	windowHeight int
}

// NewRecyclerView creates a new recycler view.
func NewRecyclerView[T any](adapter RecyclerAdapter[T], opts ...RecyclerOption[T]) *RecyclerView[T] {
	rv := &RecyclerView[T]{
		adapter:  adapter,
		selected: 0,
		pool:     make([]viewHolder, 0),
		visible:  make([]viewHolder, 0),
		tokens:   design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(rv)
	}

	rv.clampSelection()
	rv.clampScrollOffset()
	return rv
}

// WithViewport sets a fixed viewport height.
func WithViewport[T any](height int) RecyclerOption[T] {
	return func(rv *RecyclerView[T]) {
		if height >= 0 {
			rv.viewport = height
		}
	}
}

// WithWidth sets a fixed viewport width.
func WithWidth[T any](width int) RecyclerOption[T] {
	return func(rv *RecyclerView[T]) {
		if width >= 0 {
			rv.width = width
		}
	}
}

// Init initializes the component.
func (rv *RecyclerView[T]) Init() tea.Cmd {
	return nil
}

// Update handles navigation and scroll messages.
func (rv *RecyclerView[T]) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		rv.windowWidth = msg.Width
		rv.windowHeight = msg.Height
		rv.clampScrollOffset()
		return rv, nil
	case tea.MouseMsg:
		if !rv.focused {
			return rv, nil
		}
		switch msg.Button {
		case tea.MouseButtonWheelDown:
			rv.scrollBy(3)
			rv.recycleOffscreenViews()
			return rv, nil
		case tea.MouseButtonWheelUp:
			rv.scrollBy(-3)
			rv.recycleOffscreenViews()
			return rv, nil
		}
	case tea.KeyMsg:
		if !rv.focused {
			return rv, nil
		}

		prev := rv.selected
		handled := true
		switch msg.String() {
		case "up", "k":
			rv.moveSelection(-1)
		case "down", "j":
			rv.moveSelection(1)
		case "pgup", "ctrl+u":
			rv.pageMove(-1)
		case "pgdown", "ctrl+d":
			rv.pageMove(1)
		case "home", "g":
			rv.selected = 0
			rv.ensureSelectedVisible()
		case "end", "G":
			rv.selected = rv.itemCount() - 1
			rv.ensureSelectedVisible()
		default:
			handled = false
		}

		if handled {
			if prev != rv.selected {
				rv.NotifyItemChanged(prev)
				rv.NotifyItemChanged(rv.selected)
			}
			rv.recycleOffscreenViews()
			return rv, nil
		}
	}

	return rv, nil
}

// View renders the visible slice of content and optional scrollbar.
func (rv *RecyclerView[T]) View() string {
	if rv.adapter == nil {
		return ""
	}
	width := rv.effectiveWidth()
	height := rv.effectiveViewport()
	if width <= 0 || height <= 0 {
		return ""
	}

	rv.clampSelection()
	rv.clampScrollOffset()

	total := rv.totalHeight()
	showScrollbar := width > 1 && total > height
	contentWidth := width
	if showScrollbar {
		contentWidth = width - 1
	}
	if contentWidth < 0 {
		contentWidth = 0
	}

	visibleIndices := rv.visibleIndices()
	rv.updateVisibleHolders(visibleIndices, contentWidth)

	rows := make([]string, 0, height)
	viewportStart := rv.scrollOffset
	viewportEnd := rv.scrollOffset + height
	items := rv.adapter.Items()

	for _, h := range rv.visible {
		if h.index < 0 || h.index >= len(items) {
			continue
		}

		itemStart := rv.itemTop(h.index)
		itemEnd := itemStart + h.height
		if itemEnd <= viewportStart || itemStart >= viewportEnd {
			continue
		}

		itemRows := rv.holderRows(h, contentWidth)
		clipStart := rvMax(itemStart, viewportStart)
		clipEnd := rvMin(itemEnd, viewportEnd)
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

		if len(rows) >= height {
			break
		}
	}

	for len(rows) < height {
		rows = append(rows, strings.Repeat(" ", contentWidth))
	}

	for i := range rows {
		line := style.Truncate(rows[i], contentWidth, "…")
		pad := contentWidth - style.StringWidth(line)
		if pad > 0 {
			line += strings.Repeat(" ", pad)
		}
		if showScrollbar {
			line += rv.scrollbarAtRow(i, height, total)
		}
		rows[i] = line
	}

	return strings.Join(rows, "\n")
}

// Focus marks this component focused.
func (rv *RecyclerView[T]) Focus() {
	rv.focused = true
}

// Blur marks this component unfocused.
func (rv *RecyclerView[T]) Blur() {
	rv.focused = false
}

// Focused reports focus state.
func (rv *RecyclerView[T]) Focused() bool {
	return rv.focused
}

// SetAdapter replaces the adapter and resets cached holders.
func (rv *RecyclerView[T]) SetAdapter(adapter RecyclerAdapter[T]) {
	rv.adapter = adapter
	rv.selected = 0
	rv.scrollOffset = 0
	rv.NotifyDataSetChanged()
	rv.clampSelection()
	rv.clampScrollOffset()
}

// SetViewport updates visible height in terminal lines.
func (rv *RecyclerView[T]) SetViewport(height int) {
	if height < 0 {
		height = 0
	}
	rv.viewport = height
	rv.clampScrollOffset()
	rv.recycleOffscreenViews()
}

// SetWidth updates the render width.
func (rv *RecyclerView[T]) SetWidth(width int) {
	if width < 0 {
		width = 0
	}
	if rv.width == width {
		return
	}
	rv.width = width
	for i := range rv.visible {
		rv.visible[i].dirty = true
	}
}

// ScrollTo scrolls so the target index is visible.
func (rv *RecyclerView[T]) ScrollTo(index int) {
	if rv.itemCount() == 0 {
		rv.scrollOffset = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= rv.itemCount() {
		index = rv.itemCount() - 1
	}

	itemStart := rv.itemTop(index)
	itemEnd := itemStart + rv.itemHeight(index)
	viewportEnd := rv.scrollOffset + rv.effectiveViewport()

	if itemStart < rv.scrollOffset {
		rv.scrollOffset = itemStart
	} else if itemEnd > viewportEnd {
		rv.scrollOffset = itemEnd - rv.effectiveViewport()
	}

	rv.clampScrollOffset()
	rv.recycleOffscreenViews()
}

// Selected returns the currently selected item index.
func (rv *RecyclerView[T]) Selected() int {
	return rv.selected
}

// NotifyDataSetChanged invalidates all cached rendered holders.
func (rv *RecyclerView[T]) NotifyDataSetChanged() {
	for i := range rv.visible {
		rv.visible[i].dirty = true
	}
	for i := range rv.pool {
		rv.pool[i].dirty = true
		rv.pool[i].index = -1
		rv.pool[i].rendered = ""
		rv.pool[i].height = 0
	}
	rv.clampSelection()
	rv.clampScrollOffset()
}

// NotifyItemChanged invalidates a specific item holder.
func (rv *RecyclerView[T]) NotifyItemChanged(index int) {
	if index < 0 {
		return
	}
	for i := range rv.visible {
		if rv.visible[i].index == index {
			rv.visible[i].dirty = true
		}
	}
}

func (rv *RecyclerView[T]) moveSelection(delta int) {
	if rv.itemCount() == 0 {
		rv.selected = 0
		rv.scrollOffset = 0
		return
	}
	rv.selected += delta
	rv.clampSelection()
	rv.ensureSelectedVisible()
}

func (rv *RecyclerView[T]) pageMove(direction int) {
	if rv.itemCount() == 0 {
		rv.selected = 0
		rv.scrollOffset = 0
		return
	}
	targetLine := rv.scrollOffset + direction*rv.effectiveViewport()
	if targetLine < 0 {
		targetLine = 0
	}
	if targetLine > rv.totalHeight()-1 {
		targetLine = rv.totalHeight() - 1
	}
	if targetLine < 0 {
		targetLine = 0
	}
	rv.selected = rv.indexAtLine(targetLine)
	rv.ensureSelectedVisible()
}

func (rv *RecyclerView[T]) ensureSelectedVisible() {
	if rv.itemCount() == 0 {
		rv.scrollOffset = 0
		return
	}
	if rv.selected < 0 {
		rv.selected = 0
	}
	if rv.selected >= rv.itemCount() {
		rv.selected = rv.itemCount() - 1
	}

	itemStart := rv.itemTop(rv.selected)
	itemEnd := itemStart + rv.itemHeight(rv.selected)
	viewportEnd := rv.scrollOffset + rv.effectiveViewport()

	if itemStart < rv.scrollOffset {
		rv.scrollOffset = itemStart
	} else if itemEnd > viewportEnd {
		rv.scrollOffset = itemEnd - rv.effectiveViewport()
	}
	if rv.scrollOffset < 0 {
		rv.scrollOffset = 0
	}
	rv.clampScrollOffset()
}

func (rv *RecyclerView[T]) recycleOffscreenViews() {
	visibleIndices := rv.visibleIndices()
	visibleSet := make(map[int]struct{}, len(visibleIndices))
	for _, idx := range visibleIndices {
		visibleSet[idx] = struct{}{}
	}

	stillVisible := make([]viewHolder, 0, len(rv.visible))
	for _, h := range rv.visible {
		if _, ok := visibleSet[h.index]; ok {
			stillVisible = append(stillVisible, h)
			continue
		}
		h.index = -1
		h.rendered = ""
		h.height = 0
		h.dirty = true
		rv.pool = append(rv.pool, h)
	}
	rv.visible = stillVisible
}

func (rv *RecyclerView[T]) updateVisibleHolders(indices []int, width int) {
	oldVisible := rv.visible
	oldPosByIndex := make(map[int]int, len(oldVisible))
	for i, h := range oldVisible {
		oldPosByIndex[h.index] = i
	}
	usedOld := make(map[int]struct{}, len(oldVisible))
	newVisible := make([]viewHolder, 0, len(indices))
	items := rv.adapter.Items()

	for _, idx := range indices {
		if idx < 0 || idx >= len(items) {
			continue
		}

		var h viewHolder
		if oldPos, ok := oldPosByIndex[idx]; ok {
			h = oldVisible[oldPos]
			usedOld[oldPos] = struct{}{}
		} else {
			h = rv.obtainHolder()
		}

		height := rv.itemHeight(idx)
		if h.index != idx {
			h.index = idx
			h.dirty = true
		}
		if h.height != height {
			h.height = height
			h.dirty = true
		}

		if h.dirty || h.rendered == "" {
			h.rendered = rv.adapter.RenderItem(idx, items[idx], width, idx == rv.selected)
			h.dirty = false
		}

		newVisible = append(newVisible, h)
	}

	for i, h := range oldVisible {
		if _, ok := usedOld[i]; ok {
			continue
		}
		h.index = -1
		h.rendered = ""
		h.height = 0
		h.dirty = true
		rv.pool = append(rv.pool, h)
	}

	rv.visible = newVisible
}

func (rv *RecyclerView[T]) obtainHolder() viewHolder {
	if len(rv.pool) == 0 {
		return viewHolder{index: -1, dirty: true}
	}
	last := len(rv.pool) - 1
	h := rv.pool[last]
	rv.pool = rv.pool[:last]
	h.index = -1
	h.dirty = true
	h.height = 0
	h.rendered = ""
	return h
}

func (rv *RecyclerView[T]) holderRows(h viewHolder, width int) []string {
	if h.height <= 0 {
		return nil
	}
	lines := strings.Split(strings.TrimSuffix(h.rendered, "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}
	rows := make([]string, 0, h.height)
	for i := 0; i < h.height; i++ {
		if i < len(lines) {
			rows = append(rows, lines[i])
		} else {
			rows = append(rows, "")
		}
	}
	return rows
}

func (rv *RecyclerView[T]) visibleIndices() []int {
	count := rv.itemCount()
	if count == 0 {
		return nil
	}
	viewport := rv.effectiveViewport()
	if viewport <= 0 {
		return nil
	}
	start := rv.scrollOffset
	end := rv.scrollOffset + viewport

	indices := make([]int, 0)
	running := 0
	for i := 0; i < count; i++ {
		h := rv.itemHeight(i)
		itemStart := running
		itemEnd := running + h
		running = itemEnd

		if itemEnd <= start {
			continue
		}
		if itemStart >= end {
			break
		}
		indices = append(indices, i)
	}
	return indices
}

func (rv *RecyclerView[T]) totalHeight() int {
	total := 0
	for i := 0; i < rv.itemCount(); i++ {
		total += rv.itemHeight(i)
	}
	return total
}

func (rv *RecyclerView[T]) itemTop(index int) int {
	if index <= 0 {
		return 0
	}
	if index > rv.itemCount() {
		index = rv.itemCount()
	}
	top := 0
	for i := 0; i < index; i++ {
		top += rv.itemHeight(i)
	}
	return top
}

func (rv *RecyclerView[T]) indexAtLine(line int) int {
	if rv.itemCount() == 0 {
		return 0
	}
	if line <= 0 {
		return 0
	}
	running := 0
	for i := 0; i < rv.itemCount(); i++ {
		running += rv.itemHeight(i)
		if line < running {
			return i
		}
	}
	return rv.itemCount() - 1
}

func (rv *RecyclerView[T]) scrollBy(delta int) {
	rv.scrollOffset += delta
	rv.clampScrollOffset()
}

func (rv *RecyclerView[T]) clampScrollOffset() {
	if rv.scrollOffset < 0 {
		rv.scrollOffset = 0
	}
	maxOffset := rv.maxScrollOffset()
	if rv.scrollOffset > maxOffset {
		rv.scrollOffset = maxOffset
	}
}

func (rv *RecyclerView[T]) clampSelection() {
	count := rv.itemCount()
	if count == 0 {
		rv.selected = 0
		return
	}
	if rv.selected < 0 {
		rv.selected = 0
	}
	if rv.selected >= count {
		rv.selected = count - 1
	}
}

func (rv *RecyclerView[T]) maxScrollOffset() int {
	total := rv.totalHeight()
	viewport := rv.effectiveViewport()
	if viewport <= 0 || total <= viewport {
		return 0
	}
	return total - viewport
}

func (rv *RecyclerView[T]) itemCount() int {
	if rv.adapter == nil {
		return 0
	}
	count := rv.adapter.ItemCount()
	if count < 0 {
		return 0
	}
	items := rv.adapter.Items()
	if len(items) < count {
		return len(items)
	}
	return count
}

func (rv *RecyclerView[T]) itemHeight(index int) int {
	if rv.adapter == nil {
		return 1
	}
	h := rv.adapter.ItemHeight(index)
	if h <= 0 {
		return 1
	}
	return h
}

func (rv *RecyclerView[T]) effectiveViewport() int {
	if rv.viewport > 0 {
		return rv.viewport
	}
	if rv.windowHeight > 0 {
		return rv.windowHeight
	}
	return 0
}

func (rv *RecyclerView[T]) effectiveWidth() int {
	if rv.width > 0 {
		return rv.width
	}
	if rv.windowWidth > 0 {
		return rv.windowWidth
	}
	return 0
}

func (rv *RecyclerView[T]) scrollbarAtRow(row, viewport, total int) string {
	track := style.ANSIDim + "│" + style.ANSIReset
	thumb := style.ANSIBold + "█" + style.ANSIReset
	if viewport <= 0 || total <= viewport {
		return track
	}

	thumbSize := (viewport * viewport) / total
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > viewport {
		thumbSize = viewport
	}

	maxOffset := rv.maxScrollOffset()
	thumbPos := 0
	if maxOffset > 0 {
		thumbPos = (rv.scrollOffset * (viewport - thumbSize)) / maxOffset
	}

	if row >= thumbPos && row < thumbPos+thumbSize {
		return thumb
	}
	return track
}

func rvMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func rvMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var _ tui.Component = (*RecyclerView[int])(nil)
