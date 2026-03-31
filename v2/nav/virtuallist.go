package nav

import (
	"strings"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// VirtualList is a fixed-height virtualized list optimized for large data sets.
type VirtualList[T any] struct {
	items        []T
	renderItem   func(index int, item T, width int, selected bool) string
	itemHeight   int
	viewport     int
	scrollOffset int
	selected     int
	width        int
	focused      bool
	windowWidth  int
	windowHeight int
}

// NewVirtualList creates a new fixed-height virtualized list.
func NewVirtualList[T any](items []T, render func(int, T, int, bool) string, itemHeight int) *VirtualList[T] {
	if itemHeight <= 0 {
		itemHeight = 1
	}
	v := &VirtualList[T]{
		items:      append([]T(nil), items...),
		renderItem: render,
		itemHeight: itemHeight,
		selected:   0,
	}
	v.clampSelection()
	v.clampScrollOffset()
	return v
}

// Init initializes the component.
func (v *VirtualList[T]) Init() tea.Cmd {
	return nil
}

// Update handles viewport changes and keyboard/mouse scrolling.
func (v *VirtualList[T]) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.windowWidth = msg.Width
		v.windowHeight = msg.Height
		v.clampScrollOffset()
		return v, nil
	case tea.MouseMsg:
		if !v.focused {
			return v, nil
		}
		switch msg.Button {
		case tea.MouseButtonWheelDown:
			v.scrollBy(v.itemHeight)
			return v, nil
		case tea.MouseButtonWheelUp:
			v.scrollBy(-v.itemHeight)
			return v, nil
		}
	case tea.KeyMsg:
		if !v.focused {
			return v, nil
		}
		switch msg.String() {
		case "down", "j":
			v.moveSelection(1)
			return v, nil
		case "up", "k":
			v.moveSelection(-1)
			return v, nil
		case "pgdown", "ctrl+d":
			v.pageMove(1)
			return v, nil
		case "pgup", "ctrl+u":
			v.pageMove(-1)
			return v, nil
		case "home", "g":
			v.selected = 0
			v.ensureSelectedVisible()
			return v, nil
		case "end", "G":
			v.selected = len(v.items) - 1
			v.ensureSelectedVisible()
			return v, nil
		}
	}
	return v, nil
}

// View renders only the visible rows.
func (v *VirtualList[T]) View() string {
	width := v.effectiveWidth()
	height := v.effectiveViewport()
	if width <= 0 || height <= 0 {
		return ""
	}

	if len(v.items) == 0 {
		rows := make([]string, height)
		for i := range rows {
			rows[i] = strings.Repeat(" ", width)
		}
		return strings.Join(rows, "\n")
	}

	v.clampSelection()
	v.clampScrollOffset()

	total := v.totalHeight()
	showScrollbar := width > 1 && total > height
	contentWidth := width
	if showScrollbar {
		contentWidth = width - 1
	}
	if contentWidth < 0 {
		contentWidth = 0
	}

	firstVisible := 0
	if v.itemHeight > 0 {
		firstVisible = v.scrollOffset / v.itemHeight
	}
	if firstVisible < 0 {
		firstVisible = 0
	}
	if firstVisible >= len(v.items) {
		firstVisible = len(v.items) - 1
	}
	lastVisible := firstVisible + (height / v.itemHeight) + 2
	if lastVisible > len(v.items) {
		lastVisible = len(v.items)
	}

	rows := make([]string, 0, height)
	viewportStart := v.scrollOffset
	viewportEnd := v.scrollOffset + height

	for i := firstVisible; i < lastVisible; i++ {
		itemStart := i * v.itemHeight
		itemEnd := itemStart + v.itemHeight
		if itemEnd <= viewportStart {
			continue
		}
		if itemStart >= viewportEnd {
			break
		}

		itemRows := v.renderRows(i, v.items[i], contentWidth)
		clipStart := vlMax(itemStart, viewportStart)
		clipEnd := vlMin(itemEnd, viewportEnd)
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
			line += v.scrollbarAtRow(i, height, total)
		}
		rows[i] = line
	}

	return strings.Join(rows, "\n")
}

// Focus marks the component focused.
func (v *VirtualList[T]) Focus() {
	v.focused = true
}

// Blur marks the component unfocused.
func (v *VirtualList[T]) Blur() {
	v.focused = false
}

// Focused reports focus state.
func (v *VirtualList[T]) Focused() bool {
	return v.focused
}

// SetItems replaces the data set.
func (v *VirtualList[T]) SetItems(items []T) {
	v.items = append([]T(nil), items...)
	v.clampSelection()
	v.clampScrollOffset()
}

// Selected returns the selected index.
func (v *VirtualList[T]) Selected() int {
	return v.selected
}

// SelectedItem returns the selected item, or zero-value when list is empty.
func (v *VirtualList[T]) SelectedItem() T {
	var zero T
	if len(v.items) == 0 || v.selected < 0 || v.selected >= len(v.items) {
		return zero
	}
	return v.items[v.selected]
}

// ScrollTo scrolls so index is visible.
func (v *VirtualList[T]) ScrollTo(index int) {
	if len(v.items) == 0 {
		v.scrollOffset = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(v.items) {
		index = len(v.items) - 1
	}

	itemStart := index * v.itemHeight
	itemEnd := itemStart + v.itemHeight
	viewportEnd := v.scrollOffset + v.effectiveViewport()
	if itemStart < v.scrollOffset {
		v.scrollOffset = itemStart
	} else if itemEnd > viewportEnd {
		v.scrollOffset = itemEnd - v.effectiveViewport()
	}
	v.clampScrollOffset()
}

func (v *VirtualList[T]) moveSelection(delta int) {
	if len(v.items) == 0 {
		v.selected = 0
		v.scrollOffset = 0
		return
	}
	v.selected += delta
	v.clampSelection()
	v.ensureSelectedVisible()
}

func (v *VirtualList[T]) pageMove(direction int) {
	if len(v.items) == 0 {
		v.selected = 0
		v.scrollOffset = 0
		return
	}
	targetLine := v.scrollOffset + direction*v.effectiveViewport()
	if targetLine < 0 {
		targetLine = 0
	}
	maxLine := v.totalHeight() - 1
	if targetLine > maxLine {
		targetLine = maxLine
	}
	if targetLine < 0 {
		targetLine = 0
	}
	v.selected = targetLine / v.itemHeight
	v.clampSelection()
	v.ensureSelectedVisible()
}

func (v *VirtualList[T]) ensureSelectedVisible() {
	if len(v.items) == 0 {
		v.scrollOffset = 0
		return
	}
	itemStart := v.selected * v.itemHeight
	itemEnd := itemStart + v.itemHeight
	viewportEnd := v.scrollOffset + v.effectiveViewport()
	if itemStart < v.scrollOffset {
		v.scrollOffset = itemStart
	} else if itemEnd > viewportEnd {
		v.scrollOffset = itemEnd - v.effectiveViewport()
	}
	v.clampScrollOffset()
}

func (v *VirtualList[T]) renderRows(index int, item T, width int) []string {
	if v.itemHeight <= 0 {
		v.itemHeight = 1
	}
	if v.renderItem == nil {
		rows := make([]string, v.itemHeight)
		for i := range rows {
			rows[i] = ""
		}
		return rows
	}
	rendered := v.renderItem(index, item, width, index == v.selected)
	lines := strings.Split(strings.TrimSuffix(rendered, "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}
	rows := make([]string, 0, v.itemHeight)
	for i := 0; i < v.itemHeight; i++ {
		if i < len(lines) {
			rows = append(rows, lines[i])
		} else {
			rows = append(rows, "")
		}
	}
	return rows
}

func (v *VirtualList[T]) totalHeight() int {
	if len(v.items) == 0 {
		return 0
	}
	return len(v.items) * v.itemHeight
}

func (v *VirtualList[T]) scrollBy(delta int) {
	v.scrollOffset += delta
	v.clampScrollOffset()
}

func (v *VirtualList[T]) clampSelection() {
	if len(v.items) == 0 {
		v.selected = 0
		return
	}
	if v.selected < 0 {
		v.selected = 0
	}
	if v.selected >= len(v.items) {
		v.selected = len(v.items) - 1
	}
}

func (v *VirtualList[T]) clampScrollOffset() {
	if v.scrollOffset < 0 {
		v.scrollOffset = 0
	}
	maxOffset := v.maxScrollOffset()
	if v.scrollOffset > maxOffset {
		v.scrollOffset = maxOffset
	}
}

func (v *VirtualList[T]) maxScrollOffset() int {
	total := v.totalHeight()
	viewport := v.effectiveViewport()
	if viewport <= 0 || total <= viewport {
		return 0
	}
	return total - viewport
}

func (v *VirtualList[T]) effectiveViewport() int {
	if v.viewport > 0 {
		return v.viewport
	}
	if v.windowHeight > 0 {
		return v.windowHeight
	}
	return 0
}

func (v *VirtualList[T]) effectiveWidth() int {
	if v.width > 0 {
		return v.width
	}
	if v.windowWidth > 0 {
		return v.windowWidth
	}
	return 0
}

func (v *VirtualList[T]) scrollbarAtRow(row, viewport, total int) string {
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

	maxOffset := v.maxScrollOffset()
	thumbPos := 0
	if maxOffset > 0 {
		thumbPos = (v.scrollOffset * (viewport - thumbSize)) / maxOffset
	}

	if row >= thumbPos && row < thumbPos+thumbSize {
		return thumb
	}
	return track
}

func vlMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func vlMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var _ tui.Component = (*VirtualList[int])(nil)
