package tui

// HitGrid stores component IDs per terminal cell for efficient mouse dispatch.
type HitGrid struct {
	width  int
	height int
	grid   [][]string
	next   [][]string
	dirty  bool
}

// NewHitGrid creates a new hit grid with double buffers.
func NewHitGrid(width, height int) *HitGrid {
	hg := &HitGrid{}
	hg.Resize(width, height)
	return hg
}

// Resize changes the grid dimensions and clears both buffers.
func (hg *HitGrid) Resize(width, height int) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	hg.width = width
	hg.height = height
	hg.grid = newHitGridBuffer(width, height)
	hg.next = newHitGridBuffer(width, height)
	hg.dirty = false
}

// Register marks a rectangular area in the next frame buffer with a component ID.
func (hg *HitGrid) Register(id string, x, y, w, h int) {
	if id == "" || w <= 0 || h <= 0 || hg.width == 0 || hg.height == 0 {
		return
	}

	startX := max(0, x)
	startY := max(0, y)
	endX := min(hg.width, x+w)
	endY := min(hg.height, y+h)
	if startX >= endX || startY >= endY {
		return
	}

	for row := startY; row < endY; row++ {
		for col := startX; col < endX; col++ {
			hg.next[row][col] = id
		}
	}

	hg.recomputeDirty()
}

// Clear resets the next frame buffer.
func (hg *HitGrid) Clear() {
	clearHitGridBuffer(hg.next)
	hg.recomputeDirty()
}

// HitTest returns the component ID at the current frame position.
func (hg *HitGrid) HitTest(x, y int) string {
	if x < 0 || y < 0 || x >= hg.width || y >= hg.height {
		return ""
	}
	return hg.grid[y][x]
}

// Swap promotes the next buffer to the current frame and clears the staging buffer.
func (hg *HitGrid) Swap() {
	hg.grid, hg.next = hg.next, hg.grid
	clearHitGridBuffer(hg.next)
	hg.dirty = false
}

// IsDirty reports whether the next frame differs from the current frame.
func (hg *HitGrid) IsDirty() bool {
	return hg.dirty
}

func (hg *HitGrid) recomputeDirty() {
	if len(hg.grid) != len(hg.next) {
		hg.dirty = true
		return
	}

	for row := 0; row < hg.height; row++ {
		for col := 0; col < hg.width; col++ {
			if hg.grid[row][col] != hg.next[row][col] {
				hg.dirty = true
				return
			}
		}
	}

	hg.dirty = false
}

func newHitGridBuffer(width, height int) [][]string {
	buffer := make([][]string, height)
	for row := 0; row < height; row++ {
		buffer[row] = make([]string, width)
	}
	return buffer
}

func clearHitGridBuffer(buffer [][]string) {
	for row := range buffer {
		for col := range buffer[row] {
			buffer[row][col] = ""
		}
	}
}
