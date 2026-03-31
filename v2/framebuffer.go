package tui

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

// FrameBuffer manages double-buffered terminal rendering.
// It renders the full next frame to a back buffer, diffs against
// the front buffer, and only writes changed lines to the terminal.
type FrameBuffer struct {
	front     []string // currently displayed lines
	back      []string // next frame being built
	width     int
	height    int
	writer    io.Writer
	mu        sync.Mutex
	dirty     bool
	cursorRow int // tracked cursor position
	cursorCol int
}

// NewFrameBuffer creates a new double-buffered frame renderer.
// writer is typically os.Stdout. width/height are terminal dimensions.
func NewFrameBuffer(writer io.Writer, width, height int) *FrameBuffer {
	if writer == nil {
		writer = io.Discard
	}

	return &FrameBuffer{
		width:     width,
		height:    height,
		writer:    writer,
		back:      make([]string, max(height, 0)),
		dirty:     true,
		cursorRow: 1,
		cursorCol: 1,
	}
}

// Resize updates the buffer dimensions and clears both buffers.
func (fb *FrameBuffer) Resize(width, height int) {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	fb.width = width
	fb.height = height
	fb.front = nil
	fb.back = make([]string, max(height, 0))
	fb.dirty = true
}

// Render takes a complete rendered frame string (the output of View()),
// splits it into lines, diffs against the front buffer, and returns ANSI
// sequences for only the changed lines.
func (fb *FrameBuffer) Render(frame string) string {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	fb.buildBackBuffer(frame)

	if len(fb.front) == 0 || fb.dirty {
		return fb.flushLocked()
	}

	var out strings.Builder
	changed := false

	for i := 0; i < fb.height; i++ {
		if fb.back[i] == fb.front[i] {
			continue
		}

		fmt.Fprintf(&out, "\033[%d;1H%s\033[K", i+1, fb.back[i])
		changed = true
	}

	output := ""
	if changed {
		output = out.String()
		_, _ = io.WriteString(fb.writer, output)
	}

	fb.syncFrontToBack()
	return output
}

// WriteFrame writes the rendered ANSI diff output to the configured writer.
func (fb *FrameBuffer) WriteFrame(frame string) {
	_ = fb.Render(frame)
}

// Flush writes the entire back buffer to the terminal. Used on first render
// or after resize when a full redraw is needed.
func (fb *FrameBuffer) Flush() {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	_ = fb.flushLocked()
}

// Clear clears both buffers and the terminal screen.
func (fb *FrameBuffer) Clear() {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	fb.front = nil
	fb.back = make([]string, max(fb.height, 0))
	fb.dirty = true

	var out strings.Builder
	out.WriteString("\033[2J")
	out.WriteString("\033[H")
	_, _ = io.WriteString(fb.writer, out.String())
	fb.cursorRow = 1
	fb.cursorCol = 1
}

// CursorTo moves the cursor to a specific position (1-indexed).
func (fb *FrameBuffer) CursorTo(row, col int) {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if row < 1 {
		row = 1
	}
	if col < 1 {
		col = 1
	}

	var out strings.Builder
	fmt.Fprintf(&out, "\033[%d;%dH", row, col)
	_, _ = io.WriteString(fb.writer, out.String())
	fb.cursorRow = row
	fb.cursorCol = col
}

// HideCursor hides the terminal cursor.
func (fb *FrameBuffer) HideCursor() {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	_, _ = io.WriteString(fb.writer, "\033[?25l")
}

// ShowCursor shows the terminal cursor.
func (fb *FrameBuffer) ShowCursor() {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	_, _ = io.WriteString(fb.writer, "\033[?25h")
}

func (fb *FrameBuffer) flushLocked() string {
	if fb.height <= 0 {
		fb.front = nil
		fb.dirty = false
		return ""
	}

	var out strings.Builder
	for i := 0; i < fb.height; i++ {
		fmt.Fprintf(&out, "\033[%d;1H%s\033[K", i+1, fb.back[i])
	}

	output := out.String()
	_, _ = io.WriteString(fb.writer, output)
	fb.syncFrontToBack()
	fb.dirty = false
	return output
}
func (fb *FrameBuffer) buildBackBuffer(frame string) {
	if fb.height < 0 {
		fb.height = 0
	}
	if fb.width < 0 {
		fb.width = 0
	}

	lines := strings.Split(frame, "\n")
	if fb.height == 0 {
		fb.height = len(lines)
	}
	if fb.width == 0 {
		maxWidth := 0
		for _, line := range lines {
			if len(line) > maxWidth {
				maxWidth = len(line)
			}
		}
		fb.width = maxWidth
	}

	if len(fb.back) != fb.height {
		fb.back = make([]string, fb.height)
	}

	for i := 0; i < fb.height; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		fb.back[i] = normalizeFrameLine(line, fb.width)
	}
}
func (fb *FrameBuffer) syncFrontToBack() {
	if len(fb.front) != fb.height {
		fb.front = make([]string, fb.height)
	}
	copy(fb.front, fb.back)
}

func normalizeFrameLine(line string, width int) string {
	if width <= 0 {
		return ""
	}

	if len(line) > width {
		return line[:width]
	}

	if len(line) < width {
		return line + strings.Repeat(" ", width-len(line))
	}

	return line
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
