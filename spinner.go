package tui

// Spinner defines an animation sequence
type Spinner struct {
	Frames []string
}

// Predefined spinner animations
var (
	// SpinnerDots - Braille dots spinner (smooth)
	SpinnerDots = Spinner{
		Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}

	// SpinnerLine - Classic line spinner
	SpinnerLine = Spinner{
		Frames: []string{"─", "\\", "|", "/"},
	}

	// SpinnerCircle - Rotating circle
	SpinnerCircle = Spinner{
		Frames: []string{"◴", "◷", "◶", "◵"},
	}

	// SpinnerThinking - Codex/Claude-style thinking animation (small to large)
	SpinnerThinking = Spinner{
		Frames: []string{".", "+", "*", "÷", "•"},
	}

	// SpinnerCodexThinking - Codex CLI style thinking animation.
	SpinnerCodexThinking = SpinnerThinking

	// SpinnerClaudeThinking - Backward-compatible alias for SpinnerThinking.
	SpinnerClaudeThinking = SpinnerThinking

	// SpinnerBlink - Simple blink (on/off)
	SpinnerBlink = Spinner{
		Frames: []string{"⏺", " "},
	}

	// SpinnerDotsJumping - Jumping dots
	SpinnerDotsJumping = Spinner{
		Frames: []string{"⢄", "⢂", "⢁", "⡁", "⡈", "⡐", "⡠"},
	}

	// SpinnerArc - Growing arc
	SpinnerArc = Spinner{
		Frames: []string{"◜", "◠", "◝", "◞", "◡", "◟"},
	}

	// SpinnerCircleQuarters - Circle with quarters
	SpinnerCircleQuarters = Spinner{
		Frames: []string{"◐", "◓", "◑", "◒"},
	}

	// SpinnerSquare - Rotating square corners
	SpinnerSquare = Spinner{
		Frames: []string{"◰", "◳", "◲", "◱"},
	}

	// SpinnerArrows - Rotating arrows
	SpinnerArrows = Spinner{
		Frames: []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"},
	}

	// SpinnerBouncingBar - Bouncing bar
	SpinnerBouncingBar = Spinner{
		Frames: []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█", "▇", "▆", "▅", "▄", "▃", "▁"},
	}

	// SpinnerBouncingBall - Bouncing ball
	SpinnerBouncingBall = Spinner{
		Frames: []string{"⠁", "⠂", "⠄", "⡀", "⢀", "⠠", "⠐", "⠈"},
	}

	// SpinnerPulse - Pulsing circle
	SpinnerPulse = Spinner{
		Frames: []string{"○", "◔", "◐", "◕", "●", "◕", "◐", "◔"},
	}
)

// IconSet defines icons for different statuses
type IconSet struct {
	Running string
	Success string
	Error   string
	Warning string
	Info    string
	None    string
}

// Predefined icon sets
var (
	// IconSetDefault - Default icons
	IconSetDefault = IconSet{
		Running: "⏺",
		Success: "⏺",
		Error:   "⏺",
		Warning: "⏺",
		Info:    "⏺",
		None:    "⏺",
	}

	// IconSetSymbols - Unicode symbols
	IconSetSymbols = IconSet{
		Running: "⏺",
		Success: "✓",
		Error:   "✗",
		Warning: "⚠",
		Info:    "ℹ",
		None:    "⏺",
	}

	// IconSetCircles - Circle-based icons
	IconSetCircles = IconSet{
		Running: "○",
		Success: "●",
		Error:   "◯",
		Warning: "◐",
		Info:    "○",
		None:    "○",
	}

	// IconSetEmoji - Emoji icons
	IconSetEmoji = IconSet{
		Running: "⏺",
		Success: "✅",
		Error:   "❌",
		Warning: "⚡",
		Info:    "💡",
		None:    "⏺",
	}

	// IconSetMinimal - Minimal ASCII
	IconSetMinimal = IconSet{
		Running: "·",
		Success: "+",
		Error:   "x",
		Warning: "!",
		Info:    "i",
		None:    "·",
	}

	// IconSetCodex - Codex CLI style
	IconSetCodex = IconSet{
		Running: "⏺",
		Success: "✓",
		Error:   "✗",
		Warning: "⚠",
		Info:    "⏺",
		None:    "⏺",
	}

	// IconSetClaude - Backward-compatible alias for IconSetCodex.
	IconSetClaude = IconSet{
		Running: IconSetCodex.Running,
		Success: IconSetCodex.Success,
		Error:   IconSetCodex.Error,
		Warning: IconSetCodex.Warning,
		Info:    IconSetCodex.Info,
		None:    IconSetCodex.None,
	}
)

// GetFrame returns the frame at the given index
func (s Spinner) GetFrame(index int) string {
	if len(s.Frames) == 0 {
		return ""
	}
	return s.Frames[index%len(s.Frames)]
}

// FrameCount returns the number of frames in the spinner
func (s Spinner) FrameCount() int {
	return len(s.Frames)
}
