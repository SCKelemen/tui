package tui

import (
	"hash/fnv"
	"time"
)

// ComponentState stores cached render state for a component.
type ComponentState struct {
	lastView       string
	lastHash       uint64
	dirty          bool
	lastRenderTime time.Time
}

// DirtyTrackerStats captures cache hit/miss and invalidation counts.
type DirtyTrackerStats struct {
	Hits          uint64
	Misses        uint64
	Invalidations uint64
}

// DirtyTracker tracks which component views must be re-rendered.
type DirtyTracker struct {
	componentStates map[string]*ComponentState
	hits            uint64
	misses          uint64
	invalidations   uint64
}

// NewDirtyTracker creates a new DirtyTracker.
func NewDirtyTracker() *DirtyTracker {
	return &DirtyTracker{
		componentStates: make(map[string]*ComponentState),
	}
}

// MarkDirty marks a component as needing re-rendering.
func (dt *DirtyTracker) MarkDirty(id string) {
	if id == "" {
		return
	}

	state := dt.ensureState(id)
	if !state.dirty {
		dt.invalidations++
	}
	state.dirty = true
}

// IsDirty reports whether a component needs to be re-rendered.
func (dt *DirtyTracker) IsDirty(id string) bool {
	if id == "" {
		return true
	}

	state, ok := dt.componentStates[id]
	if !ok {
		return true
	}
	return state.dirty
}

// CacheView stores the rendered view for a component and clears its dirty flag.
func (dt *DirtyTracker) CacheView(id string, view string) {
	if id == "" {
		return
	}

	state := dt.ensureState(id)
	state.lastView = view
	state.lastHash = hashView(view)
	state.dirty = false
	state.lastRenderTime = time.Now()
}

// GetCachedView returns the cached view when it is still valid.
func (dt *DirtyTracker) GetCachedView(id string) (string, bool) {
	if id == "" {
		dt.misses++
		return "", false
	}

	state, ok := dt.componentStates[id]
	if !ok || state.dirty {
		dt.misses++
		return "", false
	}

	dt.hits++
	return state.lastView, true
}

// InvalidateAll marks every tracked component dirty.
func (dt *DirtyTracker) InvalidateAll() {
	for _, state := range dt.componentStates {
		if state == nil {
			continue
		}
		if !state.dirty {
			dt.invalidations++
		}
		state.dirty = true
	}
}

// Stats returns dirty-tracker cache statistics.
func (dt *DirtyTracker) Stats() DirtyTrackerStats {
	return DirtyTrackerStats{
		Hits:          dt.hits,
		Misses:        dt.misses,
		Invalidations: dt.invalidations,
	}
}

func (dt *DirtyTracker) ensureState(id string) *ComponentState {
	state, ok := dt.componentStates[id]
	if !ok || state == nil {
		state = &ComponentState{dirty: true}
		dt.componentStates[id] = state
	}
	return state
}

func hashView(view string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(view))
	return h.Sum64()
}

// ViewportCuller determines whether component rectangles intersect the viewport.
type ViewportCuller struct {
	screenWidth  int
	screenHeight int
}

// NewViewportCuller creates a new viewport culler.
func NewViewportCuller() *ViewportCuller {
	return &ViewportCuller{}
}

// SetViewport updates the active viewport size.
func (vc *ViewportCuller) SetViewport(width, height int) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	vc.screenWidth = width
	vc.screenHeight = height
}

// IsVisible reports whether a rectangle intersects the viewport.
func (vc *ViewportCuller) IsVisible(x, y, w, h int) bool {
	if w <= 0 || h <= 0 || vc.screenWidth <= 0 || vc.screenHeight <= 0 {
		return false
	}

	_, _, cw, ch := vc.ClipRect(x, y, w, h)
	return cw > 0 && ch > 0
}

// ClipRect clips a rectangle to the active viewport.
func (vc *ViewportCuller) ClipRect(x, y, w, h int) (cx, cy, cw, ch int) {
	if w <= 0 || h <= 0 || vc.screenWidth <= 0 || vc.screenHeight <= 0 {
		return 0, 0, 0, 0
	}

	right := x + w
	bottom := y + h
	if right <= 0 || bottom <= 0 || x >= vc.screenWidth || y >= vc.screenHeight {
		return 0, 0, 0, 0
	}

	cx = maxInt(x, 0)
	cy = maxInt(y, 0)
	cright := minInt(right, vc.screenWidth)
	cbottom := minInt(bottom, vc.screenHeight)
	cw = cright - cx
	ch = cbottom - cy
	if cw <= 0 || ch <= 0 {
		return 0, 0, 0, 0
	}
	return cx, cy, cw, ch
}

// VisibleRows returns the visible row window for a scrolling viewport.
func (vc *ViewportCuller) VisibleRows(totalRows, scrollOffset, viewportHeight int) (start, end int) {
	if totalRows <= 0 || viewportHeight <= 0 {
		return 0, 0
	}

	start = clampInt(scrollOffset, 0, totalRows)
	end = start + viewportHeight
	if end > totalRows {
		end = totalRows
	}
	return start, end
}

type clipRect struct {
	x int
	y int
	w int
	h int
}

// ScissorStack manages nested clipping rectangles.
type ScissorStack struct {
	stack []clipRect
}

// NewScissorStack creates a new scissor stack.
func NewScissorStack() *ScissorStack {
	return &ScissorStack{}
}

// Push adds a new clipping rectangle, intersecting it with the current clip.
func (ss *ScissorStack) Push(x, y, w, h int) {
	rect := normalizeClipRect(x, y, w, h)
	if len(ss.stack) > 0 {
		rect = intersectClipRect(ss.stack[len(ss.stack)-1], rect)
	}
	ss.stack = append(ss.stack, rect)
}

// Pop restores the previous clipping rectangle.
func (ss *ScissorStack) Pop() {
	if len(ss.stack) == 0 {
		return
	}
	ss.stack = ss.stack[:len(ss.stack)-1]
}

// Current returns the currently active clipping rectangle.
func (ss *ScissorStack) Current() (x, y, w, h int) {
	if len(ss.stack) == 0 {
		return 0, 0, 0, 0
	}
	current := ss.stack[len(ss.stack)-1]
	return current.x, current.y, current.w, current.h
}

// IsInside reports whether a point is inside the current clipping rectangle.
func (ss *ScissorStack) IsInside(x, y int) bool {
	if len(ss.stack) == 0 {
		return true
	}
	current := ss.stack[len(ss.stack)-1]
	if current.w <= 0 || current.h <= 0 {
		return false
	}
	return x >= current.x && x < current.x+current.w && y >= current.y && y < current.y+current.h
}

// Depth returns the number of active clipping rectangles.
func (ss *ScissorStack) Depth() int {
	return len(ss.stack)
}

func normalizeClipRect(x, y, w, h int) clipRect {
	if w <= 0 || h <= 0 {
		return clipRect{x: x, y: y}
	}
	return clipRect{x: x, y: y, w: w, h: h}
}

func intersectClipRect(a, b clipRect) clipRect {
	if a.w <= 0 || a.h <= 0 || b.w <= 0 || b.h <= 0 {
		return clipRect{x: maxInt(a.x, b.x), y: maxInt(a.y, b.y)}
	}

	left := maxInt(a.x, b.x)
	top := maxInt(a.y, b.y)
	right := minInt(a.x+a.w, b.x+b.w)
	bottom := minInt(a.y+a.h, b.y+b.h)
	if right <= left || bottom <= top {
		return clipRect{x: left, y: top}
	}
	return clipRect{x: left, y: top, w: right - left, h: bottom - top}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampInt(v, minValue, maxValue int) int {
	if v < minValue {
		return minValue
	}
	if v > maxValue {
		return maxValue
	}
	return v
}
