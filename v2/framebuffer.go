package tui

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode/utf8"

	runewidth "github.com/mattn/go-runewidth"
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
			w := displayWidth(line)
			if w > maxWidth {
				maxWidth = w
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

// displayWidth returns the number of terminal cells a string occupies when
// rendered, ignoring ANSI escape sequences (CSI \x1b[...<final>, OSC
// \x1b]...(BEL|ST), and bare \x1b<single-byte>). Wide and zero-width runes
// are handled via runewidth.
func displayWidth(s string) int {
	w := 0
	i := 0
	for i < len(s) {
		if s[i] == 0x1b {
			i = skipEscape(s, i)
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			// Treat a stray invalid byte as one cell to avoid losing it.
			w++
			i++
			continue
		}
		w += runewidth.RuneWidth(r)
		i += size
	}
	return w
}

// skipEscape advances past an ANSI escape sequence starting at s[i] (which
// must be 0x1b) and returns the new index. Supports CSI (\x1b[...final),
// OSC (\x1b]...BEL or ST), and lone-ESC fallbacks.
func skipEscape(s string, i int) int {
	if i >= len(s) || s[i] != 0x1b {
		return i + 1
	}
	if i+1 >= len(s) {
		return len(s)
	}
	switch s[i+1] {
	case '[':
		// CSI: \x1b[ <params> <final 0x40-0x7e>
		j := i + 2
		for j < len(s) {
			c := s[j]
			if c >= 0x40 && c <= 0x7e {
				return j + 1
			}
			j++
		}
		return j
	case ']':
		// OSC: \x1b] ... (BEL=0x07 | ST=\x1b\\)
		j := i + 2
		for j < len(s) {
			if s[j] == 0x07 {
				return j + 1
			}
			if s[j] == 0x1b && j+1 < len(s) && s[j+1] == '\\' {
				return j + 2
			}
			j++
		}
		return j
	default:
		// Two-byte escape (e.g. \x1bM, \x1b=) or unknown — consume both.
		return i + 2
	}
}

// hasOpenStyle reports whether the line contains any ANSI SGR sequence that
// is not closed by a subsequent reset. It is a heuristic — it only checks
// for an SGR (CSI ... m) appearing without a trailing reset/SGR0.
func hasOpenStyle(s string) bool {
	open := false
	i := 0
	for i < len(s) {
		if s[i] != 0x1b {
			i++
			continue
		}
		if i+1 < len(s) && s[i+1] == '[' {
			// Parse CSI; check whether the final byte is 'm' (SGR).
			j := i + 2
			for j < len(s) && !(s[j] >= 0x40 && s[j] <= 0x7e) {
				j++
			}
			if j < len(s) && s[j] == 'm' {
				params := s[i+2 : j]
				if params == "" || params == "0" || params == "00" {
					open = false
				} else {
					open = true
				}
			}
			if j >= len(s) {
				return open
			}
			i = j + 1
			continue
		}
		i = skipEscape(s, i)
	}
	return open
}

// normalizeFrameLine truncates or pads a line so that it occupies exactly
// width display cells, while preserving ANSI escape sequences and respecting
// wide characters. Escape sequences are treated as zero-width and copied
// through. When truncating in the middle of styled content, a final SGR
// reset (\x1b[0m) is appended so styling does not bleed.
func normalizeFrameLine(line string, width int) string {
	if width <= 0 {
		return ""
	}

	var out strings.Builder
	out.Grow(len(line))

	used := 0
	i := 0
	for i < len(line) {
		if line[i] == 0x1b {
			j := skipEscape(line, i)
			out.WriteString(line[i:j])
			i = j
			continue
		}

		r, size := utf8.DecodeRuneInString(line[i:])
		runeW := runewidth.RuneWidth(r)
		if r == utf8.RuneError && size == 1 {
			runeW = 1
		}

		if used+runeW > width {
			break
		}

		out.WriteString(line[i : i+size])
		used += runeW
		i += size
	}

	// If we stopped early and the source had any open SGR styling, close it.
	if i < len(line) && hasOpenStyle(line[:i]) {
		out.WriteString("\x1b[0m")
	}

	if used < width {
		out.WriteString(strings.Repeat(" ", width-used))
	}

	return out.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
