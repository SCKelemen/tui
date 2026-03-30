package spinner

import "testing"

func TestCatalogAllSpinnerTypesDefinedWithFrames(t *testing.T) {
	all := All()
	if len(all) == 0 {
		t.Fatal("expected spinner catalog to be non-empty")
	}

	required := []string{
		"braille", "braille2", "dots", "dots2", "dots3", "circle", "circleQuarters",
		"line", "simpleDots", "arrow", "arrow2", "bounce", "bouncingBar",
		"growVertical", "growHorizontal", "pulse", "moon", "earth", "clock", "star",
		"hamburger", "pipe", "toggle", "toggle2", "ellipsis", "arc", "verticalBars", "miniBar",
	}

	for _, name := range required {
		s, ok := all[name]
		if !ok {
			t.Fatalf("missing spinner type %q", name)
		}
		if s.FrameCount() == 0 {
			t.Fatalf("spinner %q has no frames", name)
		}
	}
}

func TestCatalogNamesAndByName(t *testing.T) {
	names := Names()
	if len(names) == 0 {
		t.Fatal("expected names to be non-empty")
	}

	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Fatalf("expected sorted names, but %q < %q", names[i], names[i-1])
		}
	}

	for _, name := range names {
		s := ByName(name)
		if s.FrameCount() == 0 {
			t.Fatalf("ByName(%q) returned spinner with no frames", name)
		}
	}
}
