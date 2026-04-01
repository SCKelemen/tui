package display

import (
	"fmt"
	"time"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style/design"
	tea "github.com/charmbracelet/bubbletea"
)

type frameReadyMsg time.Time

// FrameRateLimiter coalesces rapid frame requests into a target FPS cadence.
type FrameRateLimiter struct {
	targetFPS    int
	interval     time.Duration
	pending      bool
	lastEmit     time.Time
	statsWindow  time.Time
	framesWindow int
	actualFPS    float64
	focused      bool
	designTokens *design.DesignTokens
}

// FrameRateLimiterOption configures a FrameRateLimiter.
type FrameRateLimiterOption func(*FrameRateLimiter)

// WithFrameRateLimiterFPS sets the target FPS (default 120).
func WithFrameRateLimiterFPS(fps int) FrameRateLimiterOption {
	return func(f *FrameRateLimiter) {
		if fps > 0 {
			f.targetFPS = fps
			f.interval = time.Second / time.Duration(fps)
		}
	}
}

// WithFrameRateLimiterDesignTokens applies design tokens.
func WithFrameRateLimiterDesignTokens(tokens *design.DesignTokens) FrameRateLimiterOption {
	return func(f *FrameRateLimiter) {
		if tokens != nil {
			f.designTokens = tokens
		}
	}
}

// NewFrameRateLimiter creates a frame pacing utility model.
func NewFrameRateLimiter(opts ...FrameRateLimiterOption) *FrameRateLimiter {
	f := &FrameRateLimiter{
		targetFPS:    120,
		interval:     time.Second / 120,
		statsWindow:  time.Now(),
		designTokens: design.DefaultTheme(),
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// Init satisfies Bubble Tea model contract.
func (f *FrameRateLimiter) Init() tea.Cmd { return nil }

// Update marks emitted frames and updates actual FPS statistics.
func (f *FrameRateLimiter) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch t := msg.(type) {
	case frameReadyMsg:
		f.pending = false
		now := time.Time(t)
		f.lastEmit = now
		f.framesWindow++
		elapsed := now.Sub(f.statsWindow)
		if elapsed >= time.Second {
			f.actualFPS = float64(f.framesWindow) / elapsed.Seconds()
			f.framesWindow = 0
			f.statsWindow = now
		}
	}
	return f, nil
}

// Tick returns a coalesced frame command that fires no faster than the target interval.
func (f *FrameRateLimiter) Tick() tea.Cmd {
	if f.pending {
		return nil
	}
	f.pending = true

	wait := f.interval
	if !f.lastEmit.IsZero() {
		delta := time.Since(f.lastEmit)
		if delta < f.interval {
			wait = f.interval - delta
		} else {
			wait = 0
		}
	}

	return tea.Tick(wait, func(t time.Time) tea.Msg {
		return frameReadyMsg(t)
	})
}

// TargetFPS returns desired frame rate.
func (f *FrameRateLimiter) TargetFPS() int { return f.targetFPS }

// ActualFPS returns measured frame rate over the rolling stats window.
func (f *FrameRateLimiter) ActualFPS() float64 { return f.actualFPS }

// View returns frame pacing debug text.
func (f *FrameRateLimiter) View() string {
	return fmt.Sprintf("fps %.1f/%d", f.actualFPS, f.targetFPS)
}

// Focus marks focus state.
func (f *FrameRateLimiter) Focus() { f.focused = true }

// Blur marks blur state.
func (f *FrameRateLimiter) Blur() { f.focused = false }

// Focused reports focus state.
func (f *FrameRateLimiter) Focused() bool { return f.focused }

var _ tui.Component = (*FrameRateLimiter)(nil)
