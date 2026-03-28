package spinner

import "testing"

func TestAllSpinnersHaveFrames(t *testing.T) {
	for name, s := range All() {
		if s.FrameCount() == 0 {
			t.Errorf("spinner %q has no frames", name)
		}
	}
}

func TestGetFrameWraps(t *testing.T) {
	s := Braille
	if s.GetFrame(0) != s.GetFrame(s.FrameCount()) {
		t.Error("GetFrame should wrap")
	}
}

func TestByNameFallback(t *testing.T) {
	s := ByName("nonexistent")
	if s.FrameCount() != Braille.FrameCount() {
		t.Error("ByName should fall back to Braille")
	}
}

func TestNew(t *testing.T) {
	s := New("a", "b", "c")
	if s.FrameCount() != 3 {
		t.Errorf("expected 3 frames, got %d", s.FrameCount())
	}
	if s.GetFrame(1) != "b" {
		t.Errorf("expected 'b', got %q", s.GetFrame(1))
	}
}

func TestNames(t *testing.T) {
	names := Names()
	if len(names) < 20 {
		t.Errorf("expected at least 20 spinners, got %d", len(names))
	}
	// Check sorted
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("names not sorted: %s before %s", names[i-1], names[i])
		}
	}
}
