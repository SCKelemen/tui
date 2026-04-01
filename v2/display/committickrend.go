package display

import (
	"strings"
	"time"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	"github.com/SCKelemen/tui/v2/style/design"
	tea "github.com/charmbracelet/bubbletea"
)

// commitTickMsg advances queued text into the visible buffer.
type commitTickMsg time.Time

// CommitTickRenderer provides smooth, queue-backed streaming text output.
//
// The renderer drains from an enqueue queue on each tick. It uses adaptive chunking:
//   - smooth mode: single-rune commits for typing-like animation
//   - catch-up mode: larger rune chunks when backlog grows
type CommitTickRenderer struct {
	queue           []rune
	visible         strings.Builder
	tickRate        time.Duration
	smoothChunk     int
	catchUpChunk    int
	catchUpAt       int
	maxRunesPerTick int
	focused         bool
	designTokens    *design.DesignTokens
}

// CommitTickRendererOption configures a CommitTickRenderer.
type CommitTickRendererOption func(*CommitTickRenderer)

// WithCommitTickRate sets the drain cadence.
func WithCommitTickRate(rate time.Duration) CommitTickRendererOption {
	return func(r *CommitTickRenderer) {
		if rate > 0 {
			r.tickRate = rate
		}
	}
}

// WithCommitTickSmoothChunk sets smooth-mode chunk size.
func WithCommitTickSmoothChunk(chunk int) CommitTickRendererOption {
	return func(r *CommitTickRenderer) {
		if chunk > 0 {
			r.smoothChunk = chunk
		}
	}
}

// WithCommitTickCatchUpChunk sets catch-up chunk size.
func WithCommitTickCatchUpChunk(chunk int) CommitTickRendererOption {
	return func(r *CommitTickRenderer) {
		if chunk > 0 {
			r.catchUpChunk = chunk
		}
	}
}

// WithCommitTickBacklogThreshold sets queue size that enables catch-up mode.
func WithCommitTickBacklogThreshold(threshold int) CommitTickRendererOption {
	return func(r *CommitTickRenderer) {
		if threshold > 0 {
			r.catchUpAt = threshold
		}
	}
}

// WithCommitTickMaxRunesPerTick rate-limits total runes committed in one tick.
func WithCommitTickMaxRunesPerTick(max int) CommitTickRendererOption {
	return func(r *CommitTickRenderer) {
		if max > 0 {
			r.maxRunesPerTick = max
		}
	}
}

// WithCommitTickDesignTokens applies design tokens.
func WithCommitTickDesignTokens(tokens *design.DesignTokens) CommitTickRendererOption {
	return func(r *CommitTickRenderer) {
		if tokens != nil {
			r.designTokens = tokens
		}
	}
}

// NewCommitTickRenderer creates a queue-backed incremental renderer.
func NewCommitTickRenderer(opts ...CommitTickRendererOption) *CommitTickRenderer {
	r := &CommitTickRenderer{
		queue:           make([]rune, 0, 256),
		tickRate:        16 * time.Millisecond,
		smoothChunk:     1,
		catchUpChunk:    8,
		catchUpAt:       120,
		maxRunesPerTick: 32,
		designTokens:    design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Init satisfies the Bubble Tea model contract.
func (r *CommitTickRenderer) Init() tea.Cmd { return nil }

// Update handles tick messages and drains queued content.
func (r *CommitTickRenderer) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg.(type) {
	case commitTickMsg:
		r.drainOneTick()
		if len(r.queue) > 0 {
			return r, r.Tick()
		}
	}
	return r, nil
}

// Enqueue adds text to the render queue.
func (r *CommitTickRenderer) Enqueue(text string) {
	if text == "" {
		return
	}
	r.queue = append(r.queue, []rune(text)...)
}

// Tick returns a command that commits queue data at the configured rate.
func (r *CommitTickRenderer) Tick() tea.Cmd {
	return tea.Tick(r.tickRate, func(t time.Time) tea.Msg {
		return commitTickMsg(t)
	})
}

// View renders drained output.
func (r *CommitTickRenderer) View() string {
	out := r.visible.String()
	if r.designTokens == nil {
		return out
	}
	color := style.Fg(r.designTokens.Color)
	if color == "" {
		return out
	}
	return color + out + style.ANSIReset
}

// Focus marks focus state.
func (r *CommitTickRenderer) Focus() { r.focused = true }

// Blur marks blur state.
func (r *CommitTickRenderer) Blur() { r.focused = false }

// Focused reports focus state.
func (r *CommitTickRenderer) Focused() bool { return r.focused }

func (r *CommitTickRenderer) drainOneTick() {
	if len(r.queue) == 0 {
		return
	}

	chunk := r.smoothChunk
	if len(r.queue) >= r.catchUpAt {
		chunk = r.catchUpChunk
	}
	if chunk > r.maxRunesPerTick {
		chunk = r.maxRunesPerTick
	}
	if chunk > len(r.queue) {
		chunk = len(r.queue)
	}

	r.visible.WriteString(string(r.queue[:chunk]))
	r.queue = r.queue[chunk:]
}

var _ tui.Component = (*CommitTickRenderer)(nil)
