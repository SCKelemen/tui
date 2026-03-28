// Package spinner provides a library of terminal spinner animations.
package spinner

// Spinner holds the frames for an animation cycle.
type Spinner struct {
	Frames []string
}

// GetFrame returns the frame at the given index (wrapping).
func (s Spinner) GetFrame(index int) string {
	if len(s.Frames) == 0 {
		return ""
	}
	return s.Frames[index%len(s.Frames)]
}

// FrameCount returns the number of frames.
func (s Spinner) FrameCount() int {
	return len(s.Frames)
}

// New creates a custom spinner from the given frames.
func New(frames ...string) Spinner {
	return Spinner{Frames: frames}
}
