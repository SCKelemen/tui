// Package animation provides small, dependency-free primitives for building
// time-based animations on top of Bubble Tea. It ships with a catalogue of
// Robert Penner easing functions, a value tween, a named timeline composed of
// delayed tweens, and a tea.Cmd-based ticker that drives redraws at a target
// frame rate.
//
// Usage:
//
//	t := animation.NewTween(0, 100, 250*time.Millisecond, animation.EaseOutCubic)
//	t.Start(time.Now())
//
//	// Inside Update:
//	return m, animation.TickCmd(60)
//
//	// Inside View:
//	value := t.At(time.Now())
//
// Timelines fan out the same idea across multiple named tweens with optional
// delays so groups of values can be choreographed together:
//
//	tl := animation.NewTimeline().
//	    Add("fade", animation.NewTween(0, 1, 200*time.Millisecond, animation.EaseLinear), 0).
//	    Add("slide", animation.NewTween(-10, 0, 250*time.Millisecond, animation.EaseOutBack), 100*time.Millisecond)
//	tl.Start(time.Now())
package animation

import (
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Easing maps a normalised time in [0, 1] to a normalised progress value.
// Implementations are expected to satisfy f(0) == 0 and f(1) == 1; intermediate
// values may overshoot the unit interval (for example EaseInBack and
// EaseOutBack briefly leave [0, 1]).
type Easing func(t float64) float64

// backOvershoot is the canonical Penner overshoot constant used by the
// EaseInBack and EaseOutBack curves.
const backOvershoot = 1.70158

// EaseLinear is the identity easing; values change at a constant rate.
func EaseLinear(t float64) float64 { return t }

// EaseInQuad accelerates from zero with a quadratic curve.
func EaseInQuad(t float64) float64 { return t * t }

// EaseOutQuad decelerates to one with a quadratic curve.
func EaseOutQuad(t float64) float64 { return 1 - (1-t)*(1-t) }

// EaseInOutQuad combines EaseInQuad and EaseOutQuad around t = 0.5.
func EaseInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	u := 1 - t
	return 1 - 2*u*u
}

// EaseInCubic accelerates from zero with a cubic curve.
func EaseInCubic(t float64) float64 { return t * t * t }

// EaseOutCubic decelerates to one with a cubic curve.
func EaseOutCubic(t float64) float64 {
	u := 1 - t
	return 1 - u*u*u
}

// EaseInOutCubic combines EaseInCubic and EaseOutCubic around t = 0.5.
func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	u := 1 - t
	return 1 - 4*u*u*u
}

// EaseInExpo accelerates from zero with an exponential curve.
func EaseInExpo(t float64) float64 {
	if t <= 0 {
		return 0
	}
	return math.Pow(2, 10*(t-1))
}

// EaseOutExpo decelerates to one with an exponential curve.
func EaseOutExpo(t float64) float64 {
	if t >= 1 {
		return 1
	}
	return 1 - math.Pow(2, -10*t)
}

// EaseInOutExpo combines EaseInExpo and EaseOutExpo around t = 0.5.
func EaseInOutExpo(t float64) float64 {
	switch {
	case t <= 0:
		return 0
	case t >= 1:
		return 1
	case t < 0.5:
		return math.Pow(2, 20*t-10) / 2
	default:
		return (2 - math.Pow(2, -20*t+10)) / 2
	}
}

// EaseInBack accelerates from zero with an anticipation/overshoot curve.
func EaseInBack(t float64) float64 {
	c1 := backOvershoot
	c3 := c1 + 1
	return c3*t*t*t - c1*t*t
}

// EaseOutBack decelerates to one with an overshoot curve.
func EaseOutBack(t float64) float64 {
	c1 := backOvershoot
	c3 := c1 + 1
	u := t - 1
	return 1 + c3*u*u*u + c1*u*u
}

// Tween animates a single float64 value from From to To over Duration, applying
// Easing to the normalised progress. Tweens are inert until Start is called.
type Tween struct {
	from     float64
	to       float64
	duration time.Duration
	easing   Easing
	start    time.Time
	started  bool
}

// NewTween constructs a Tween. A nil easing defaults to EaseLinear; a
// non-positive duration is treated as zero and causes the tween to finish
// immediately on the first call to At after Start.
func NewTween(from, to float64, duration time.Duration, easing Easing) *Tween {
	if easing == nil {
		easing = EaseLinear
	}
	if duration < 0 {
		duration = 0
	}
	return &Tween{
		from:     from,
		to:       to,
		duration: duration,
		easing:   easing,
	}
}

// Start anchors the tween to the provided wall-clock time. Subsequent calls to
// At, Progress, and Done are interpreted relative to that anchor.
func (t *Tween) Start(now time.Time) {
	t.start = now
	t.started = true
}

// Progress returns the elapsed fraction of the tween clamped to [0, 1]. If the
// tween has not been started Progress returns 0.
func (t *Tween) Progress(now time.Time) float64 {
	if !t.started || t.duration <= 0 {
		if t.started && t.duration <= 0 {
			return 1
		}
		return 0
	}
	elapsed := now.Sub(t.start)
	if elapsed <= 0 {
		return 0
	}
	if elapsed >= t.duration {
		return 1
	}
	return float64(elapsed) / float64(t.duration)
}

// At returns the eased value at the given time. Before Start, or while elapsed
// is non-positive, At returns From. Once elapsed reaches Duration it returns To
// exactly.
func (t *Tween) At(now time.Time) float64 {
	if !t.started {
		return t.from
	}
	if t.duration <= 0 {
		return t.to
	}
	elapsed := now.Sub(t.start)
	if elapsed <= 0 {
		return t.from
	}
	if elapsed >= t.duration {
		return t.to
	}
	p := float64(elapsed) / float64(t.duration)
	return t.from + (t.to-t.from)*t.easing(p)
}

// Done reports whether the tween has reached its end time.
func (t *Tween) Done(now time.Time) bool {
	if !t.started {
		return false
	}
	if t.duration <= 0 {
		return true
	}
	return now.Sub(t.start) >= t.duration
}

// timelineEntry binds a named tween to a delay relative to the timeline start.
type timelineEntry struct {
	name  string
	tween *Tween
	delay time.Duration
}

// Timeline schedules a collection of named Tweens, each with its own delay
// measured from the timeline's Start time.
type Timeline struct {
	entries []timelineEntry
	start   time.Time
	started bool
}

// NewTimeline returns an empty Timeline ready for Add and Start.
func NewTimeline() *Timeline {
	return &Timeline{}
}

// Add registers a tween under name with the given delay. Add returns the
// receiver for fluent composition. If the timeline has already been started,
// the new entry is anchored immediately using the existing start time.
func (tl *Timeline) Add(name string, tween *Tween, delay time.Duration) *Timeline {
	if delay < 0 {
		delay = 0
	}
	tl.entries = append(tl.entries, timelineEntry{name: name, tween: tween, delay: delay})
	if tl.started && tween != nil {
		tween.Start(tl.start.Add(delay))
	}
	return tl
}

// Start anchors the timeline to now and starts every registered tween at
// now + delay.
func (tl *Timeline) Start(now time.Time) {
	tl.start = now
	tl.started = true
	for _, e := range tl.entries {
		if e.tween != nil {
			e.tween.Start(tl.start.Add(e.delay))
		}
	}
}

// Value returns the current value of the named tween together with an ok flag.
// ok is false when no entry with the given name exists.
func (tl *Timeline) Value(name string, now time.Time) (float64, bool) {
	for _, e := range tl.entries {
		if e.name != name {
			continue
		}
		if e.tween == nil {
			return 0, true
		}
		return e.tween.At(now), true
	}
	return 0, false
}

// Done reports whether every registered tween has completed at now.
func (tl *Timeline) Done(now time.Time) bool {
	if !tl.started {
		return false
	}
	for _, e := range tl.entries {
		if e.tween == nil {
			continue
		}
		if !e.tween.Done(now) {
			return false
		}
	}
	return true
}

// TickMsg is delivered by TickCmd on each tick and carries the wall-clock time
// at which the tick fired.
type TickMsg struct {
	Time time.Time
}

// TickCmd returns a tea.Cmd that fires a TickMsg once per frame at the given
// frames-per-second. Values ≤ 0 fall back to 60 fps.
func TickCmd(fps int) tea.Cmd {
	if fps <= 0 {
		fps = 60
	}
	interval := time.Second / time.Duration(fps)
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return TickMsg{Time: t}
	})
}
