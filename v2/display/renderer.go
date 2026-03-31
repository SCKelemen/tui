package display

import (
	"io"
	"strings"

	tui "github.com/SCKelemen/tui/v2"
)

// FrameBuffer aliases the core TUI framebuffer for display-level rendering.
type FrameBuffer = tui.FrameBuffer

// Renderer wraps a FrameBuffer to provide a high-level rendering API.
// It accepts a full frame string (from View()), diffs it against the previous
// frame, and writes only the changed lines to the output.
type Renderer struct {
	fb     *FrameBuffer
	width  int
	height int
	out    io.Writer
}

// NewRenderer creates a FrameBuffer and wraps it.
func NewRenderer(width, height int, out io.Writer) *Renderer {
	if out == nil {
		out = io.Discard
	}

	return &Renderer{
		fb:     tui.NewFrameBuffer(io.Discard, width, height),
		width:  width,
		height: height,
		out:    out,
	}
}

// Render diffs the new frame with the previous frame and returns ANSI output.
func (r *Renderer) Render(frame string) string {
	lines := strings.Split(frame, "\n")
	if r.height > 0 {
		switch {
		case len(lines) < r.height:
			lines = append(lines, make([]string, r.height-len(lines))...)
		case len(lines) > r.height:
			lines = lines[:r.height]
		}
	}

	output := r.fb.Render(strings.Join(lines, "\n"))
	if output != "" {
		_, _ = io.WriteString(r.out, output)
	}
	return output
}

// Resize updates dimensions and resets the frame buffer.
func (r *Renderer) Resize(width, height int) {
	r.width = width
	r.height = height
	r.fb.Resize(width, height)
}

// Clear clears the renderer output and resets the frame state.
func (r *Renderer) Clear() string {
	r.fb.Clear()
	seq := "\033[2J\033[H"
	_, _ = io.WriteString(r.out, seq)
	return seq
}
