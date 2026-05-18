package animation

import (
	"math"
	"testing"
	"time"
)

const eps = 1e-9

func approxEqual(got, want float64) bool {
	return math.Abs(got-want) < eps
}

func TestEasingEndpoints(t *testing.T) {
	cases := []struct {
		name   string
		easing Easing
	}{
		{"EaseLinear", EaseLinear},
		{"EaseInQuad", EaseInQuad},
		{"EaseOutQuad", EaseOutQuad},
		{"EaseInOutQuad", EaseInOutQuad},
		{"EaseInCubic", EaseInCubic},
		{"EaseOutCubic", EaseOutCubic},
		{"EaseInOutCubic", EaseInOutCubic},
		{"EaseInExpo", EaseInExpo},
		{"EaseOutExpo", EaseOutExpo},
		{"EaseInOutExpo", EaseInOutExpo},
		{"EaseInBack", EaseInBack},
		{"EaseOutBack", EaseOutBack},
	}
	for _, c := range cases {
		if got := c.easing(0); !approxEqual(got, 0) {
			t.Errorf("%s(0) = %v, want 0", c.name, got)
		}
		if got := c.easing(1); !approxEqual(got, 1) {
			t.Errorf("%s(1) = %v, want 1", c.name, got)
		}
	}
}

func TestEaseLinearMidpoint(t *testing.T) {
	if got := EaseLinear(0.5); !approxEqual(got, 0.5) {
		t.Fatalf("EaseLinear(0.5) = %v, want 0.5", got)
	}
}

func TestEaseInQuadMidpoint(t *testing.T) {
	if got := EaseInQuad(0.5); !approxEqual(got, 0.25) {
		t.Fatalf("EaseInQuad(0.5) = %v, want 0.25", got)
	}
}

func TestTweenLifecycle(t *testing.T) {
	const duration = 100 * time.Millisecond
	tw := NewTween(10, 30, duration, EaseLinear)
	start := time.Unix(1_700_000_000, 0)

	// Before Start: At should return from.
	if got := tw.At(start); !approxEqual(got, 10) {
		t.Fatalf("At before Start = %v, want 10", got)
	}
	if tw.Done(start) {
		t.Fatalf("Done before Start = true, want false")
	}

	tw.Start(start)

	// At t == start, value is from.
	if got := tw.At(start); !approxEqual(got, 10) {
		t.Fatalf("At at start = %v, want 10", got)
	}

	// Linear at 50% should be midpoint.
	mid := start.Add(duration / 2)
	if got := tw.At(mid); !approxEqual(got, 20) {
		t.Fatalf("At midpoint = %v, want 20", got)
	}
	if got := tw.Progress(mid); !approxEqual(got, 0.5) {
		t.Fatalf("Progress midpoint = %v, want 0.5", got)
	}

	// After duration: value is to, Done flips true.
	end := start.Add(duration)
	if got := tw.At(end); !approxEqual(got, 30) {
		t.Fatalf("At end = %v, want 30", got)
	}
	if !tw.Done(end) {
		t.Fatalf("Done at end = false, want true")
	}

	// Well past duration still returns to.
	past := start.Add(duration + 50*time.Millisecond)
	if got := tw.At(past); !approxEqual(got, 30) {
		t.Fatalf("At past end = %v, want 30", got)
	}
}

func TestTweenNilEasingDefaultsToLinear(t *testing.T) {
	tw := NewTween(0, 10, 100*time.Millisecond, nil)
	start := time.Unix(1_700_000_000, 0)
	tw.Start(start)
	mid := start.Add(50 * time.Millisecond)
	if got := tw.At(mid); !approxEqual(got, 5) {
		t.Fatalf("nil easing midpoint = %v, want 5", got)
	}
}

func TestTimelineWithDelay(t *testing.T) {
	const (
		delay    = 100 * time.Millisecond
		duration = 200 * time.Millisecond
	)
	start := time.Unix(1_700_000_000, 0)
	tw := NewTween(5, 25, duration, EaseLinear)
	tl := NewTimeline().Add("a", tw, delay)
	tl.Start(start)

	// At start: value should equal from, ok=true.
	got, ok := tl.Value("a", start)
	if !ok {
		t.Fatalf("expected ok at start")
	}
	if !approxEqual(got, 5) {
		t.Fatalf("value at start = %v, want 5", got)
	}

	// At start+delay: still from (tween just begins).
	atDelay := start.Add(delay)
	got, ok = tl.Value("a", atDelay)
	if !ok {
		t.Fatalf("expected ok at delay boundary")
	}
	if !approxEqual(got, 5) {
		t.Fatalf("value at delay = %v, want 5", got)
	}

	// At start+delay+duration: value is to.
	atEnd := start.Add(delay + duration)
	got, ok = tl.Value("a", atEnd)
	if !ok {
		t.Fatalf("expected ok at end")
	}
	if !approxEqual(got, 25) {
		t.Fatalf("value at end = %v, want 25", got)
	}

	if !tl.Done(atEnd) {
		t.Fatalf("Done at end = false, want true")
	}
	if tl.Done(atDelay) {
		t.Fatalf("Done at delay = true, want false")
	}
}

func TestTimelineUnknownName(t *testing.T) {
	tl := NewTimeline()
	if _, ok := tl.Value("missing", time.Now()); ok {
		t.Fatalf("Value(\"missing\") ok = true, want false")
	}
}

func TestTickCmd(t *testing.T) {
	cmd := TickCmd(60)
	if cmd == nil {
		t.Fatalf("TickCmd(60) returned nil")
	}

	before := time.Now()
	msg := cmd()
	tick, ok := msg.(TickMsg)
	if !ok {
		t.Fatalf("TickCmd produced %T, want TickMsg", msg)
	}
	if tick.Time.Before(before.Add(-time.Second)) || tick.Time.After(time.Now().Add(time.Second)) {
		t.Fatalf("TickMsg.Time = %v, expected within +/- 1s of now", tick.Time)
	}
}

func TestTickCmdDefaultFPS(t *testing.T) {
	if cmd := TickCmd(0); cmd == nil {
		t.Fatalf("TickCmd(0) returned nil; expected default 60 fps cmd")
	}
	if cmd := TickCmd(-5); cmd == nil {
		t.Fatalf("TickCmd(-5) returned nil; expected default 60 fps cmd")
	}
}
