package input

import (
	"strings"
	"testing"

	design "github.com/SCKelemen/design-system"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSliderDefaults(t *testing.T) {
	s := NewSlider()
	if s == nil {
		t.Fatal("NewSlider returned nil")
	}
	if s.min != 0 || s.max != 100 {
		t.Errorf("default range = [%v, %v], want [0, 100]", s.min, s.max)
	}
	if s.value != 0 {
		t.Errorf("default value = %v, want 0", s.value)
	}
	if s.width != 30 {
		t.Errorf("default width = %d, want 30", s.width)
	}
	if !s.showValue {
		t.Error("showValue should default to true")
	}
	// Default step is (max-min)/100 = 1
	if s.step != 1 {
		t.Errorf("default step = %v, want 1", s.step)
	}
	if s.designTokens == nil {
		t.Error("designTokens should default to non-nil")
	}
}

func TestSliderOptionsSetFields(t *testing.T) {
	tokens := design.DefaultTheme()
	cb := func(v float64) tea.Cmd { return nil }
	s := NewSlider(
		WithSliderMin(-10),
		WithSliderMax(10),
		WithSliderValue(5),
		WithSliderStep(0.25),
		WithSliderWidth(20),
		WithSliderShowValue(false),
		WithSliderLabel("Vol"),
		WithSliderDesignTokens(tokens),
		WithSliderOnChange(cb),
	)
	if s.min != -10 || s.max != 10 {
		t.Errorf("range = [%v, %v], want [-10, 10]", s.min, s.max)
	}
	if s.value != 5 {
		t.Errorf("value = %v, want 5", s.value)
	}
	if s.step != 0.25 {
		t.Errorf("step = %v, want 0.25", s.step)
	}
	if s.width != 20 {
		t.Errorf("width = %d, want 20", s.width)
	}
	if s.showValue {
		t.Error("showValue should be false")
	}
	if s.label != "Vol" {
		t.Errorf("label = %q, want %q", s.label, "Vol")
	}
	if s.designTokens != tokens {
		t.Error("design tokens not propagated")
	}
	if s.onChange == nil {
		t.Error("onChange should be set")
	}
}

func TestSliderRightIncrements(t *testing.T) {
	s := NewSlider(WithSliderMin(0), WithSliderMax(10), WithSliderValue(3), WithSliderStep(1))
	s.Focus()
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyRight})
	if s.Value() != 4 {
		t.Errorf("after right: value = %v, want 4", s.Value())
	}
}

func TestSliderLKeyIncrements(t *testing.T) {
	s := NewSlider(WithSliderMin(0), WithSliderMax(10), WithSliderValue(3), WithSliderStep(2))
	s.Focus()
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if s.Value() != 5 {
		t.Errorf("after l: value = %v, want 5", s.Value())
	}
}

func TestSliderLeftDecrements(t *testing.T) {
	s := NewSlider(WithSliderMin(0), WithSliderMax(10), WithSliderValue(3), WithSliderStep(1))
	s.Focus()
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if s.Value() != 2 {
		t.Errorf("after left: value = %v, want 2", s.Value())
	}
}

func TestSliderHKeyDecrements(t *testing.T) {
	s := NewSlider(WithSliderMin(0), WithSliderMax(10), WithSliderValue(5), WithSliderStep(1))
	s.Focus()
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if s.Value() != 4 {
		t.Errorf("after h: value = %v, want 4", s.Value())
	}
}

func TestSliderRightAtMaxStays(t *testing.T) {
	s := NewSlider(WithSliderMin(0), WithSliderMax(10), WithSliderValue(10), WithSliderStep(1))
	s.Focus()
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyRight})
	if s.Value() != 10 {
		t.Errorf("value clamped to max, got %v", s.Value())
	}
}

func TestSliderLeftAtMinStays(t *testing.T) {
	s := NewSlider(WithSliderMin(0), WithSliderMax(10), WithSliderValue(0), WithSliderStep(1))
	s.Focus()
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if s.Value() != 0 {
		t.Errorf("value should clamp to min, got %v", s.Value())
	}
}

func TestSliderHomeAndEnd(t *testing.T) {
	s := NewSlider(WithSliderMin(0), WithSliderMax(10), WithSliderValue(5), WithSliderStep(1))
	s.Focus()
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyHome})
	if s.Value() != 0 {
		t.Errorf("after home: value = %v, want 0", s.Value())
	}
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if s.Value() != 10 {
		t.Errorf("after end: value = %v, want 10", s.Value())
	}
}

func TestSliderPageKeys(t *testing.T) {
	s := NewSlider(WithSliderMin(0), WithSliderMax(100), WithSliderValue(50), WithSliderStep(1))
	s.Focus()
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if s.Value() != 40 {
		t.Errorf("after pgup (10*step): value = %v, want 40", s.Value())
	}
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if s.Value() != 50 {
		t.Errorf("after pgdown: value = %v, want 50", s.Value())
	}
}

func TestSliderSetValueClamps(t *testing.T) {
	s := NewSlider(WithSliderMin(0), WithSliderMax(10))

	s.SetValue(-5)
	if s.Value() != 0 {
		t.Errorf("SetValue(-5) clamp to 0, got %v", s.Value())
	}
	s.SetValue(99)
	if s.Value() != 10 {
		t.Errorf("SetValue(99) clamp to 10, got %v", s.Value())
	}
}

func TestSliderViewContainsLabelAndValueAndChars(t *testing.T) {
	s := NewSlider(
		WithSliderMin(0),
		WithSliderMax(10),
		WithSliderValue(4),
		WithSliderWidth(10),
		WithSliderLabel("Vol"),
		WithSliderShowValue(true),
	)
	out := s.View()
	if !strings.Contains(out, "Vol") {
		t.Errorf("View should contain label, got %q", out)
	}
	if !strings.HasSuffix(stripANSI(out), " 4") {
		t.Errorf("View should end with value 4, got %q", stripANSI(out))
	}
	if !strings.ContainsRune(out, sliderFilledChar) {
		t.Errorf("View should contain filled char, got %q", out)
	}
	if !strings.ContainsRune(out, sliderEmptyChar) {
		t.Errorf("View should contain empty char, got %q", out)
	}
	if !strings.ContainsRune(out, sliderKnobChar) {
		t.Errorf("View should contain knob, got %q", out)
	}
}

func TestSliderViewTrackLengthMatchesWidth(t *testing.T) {
	s := NewSlider(WithSliderMin(0), WithSliderMax(10), WithSliderValue(5), WithSliderWidth(11))
	stripped := stripANSI(s.View())
	// Count slider runes (filled + empty + knob) — should equal width.
	count := 0
	for _, r := range stripped {
		if r == sliderFilledChar || r == sliderEmptyChar || r == sliderKnobChar {
			count++
		}
	}
	if count != 11 {
		t.Errorf("track length = %d, want 11; view=%q", count, stripped)
	}
}

func TestSliderFocusStyling(t *testing.T) {
	s := NewSlider(WithSliderMin(0), WithSliderMax(10), WithSliderValue(5))
	s.Focus()
	focusedOut := s.View()
	if !strings.Contains(focusedOut, style.ANSIBold) {
		t.Errorf("focused View should contain ANSIBold, got %q", focusedOut)
	}
	s.Blur()
	blurredOut := s.View()
	if strings.Contains(blurredOut, style.ANSIBold) {
		// Allow bold elsewhere? Knob renders dim when blurred, so no Bold.
		t.Errorf("blurred View should not contain ANSIBold, got %q", blurredOut)
	}
}

func TestSliderOnChangeFires(t *testing.T) {
	var got float64
	called := false
	s := NewSlider(
		WithSliderMin(0),
		WithSliderMax(10),
		WithSliderValue(3),
		WithSliderStep(1),
		WithSliderOnChange(func(v float64) tea.Cmd {
			got = v
			called = true
			return nil
		}),
	)
	s.Focus()
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyRight})
	if !called {
		t.Fatal("OnChange not called")
	}
	if got != 4 {
		t.Errorf("OnChange got %v, want 4", got)
	}
}

func TestSliderOnChangeNotFiredWhenAtBoundary(t *testing.T) {
	called := false
	s := NewSlider(
		WithSliderMin(0),
		WithSliderMax(10),
		WithSliderValue(10),
		WithSliderStep(1),
		WithSliderOnChange(func(v float64) tea.Cmd {
			called = true
			return nil
		}),
	)
	s.Focus()
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyRight})
	if called {
		t.Errorf("OnChange should not fire when value doesn't change")
	}
}

func TestSliderKeyIgnoredWhenBlurred(t *testing.T) {
	s := NewSlider(WithSliderMin(0), WithSliderMax(10), WithSliderValue(3), WithSliderStep(1))
	// Not focused — key should be ignored.
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyRight})
	if s.Value() != 3 {
		t.Errorf("blurred slider should ignore key, got value = %v", s.Value())
	}
}

func TestSliderBoundsAndMouseSnap(t *testing.T) {
	s := NewSlider(WithSliderMin(0), WithSliderMax(100), WithSliderValue(0), WithSliderWidth(11))
	s.SetBounds(5, 2, 11, 1)

	bx, by, bw, bh := s.Bounds()
	if bx != 5 || by != 2 || bw != 11 || bh != 1 {
		t.Fatalf("Bounds returned (%d,%d,%d,%d), want (5,2,11,1)", bx, by, bw, bh)
	}

	s.Focus()
	// Click at the far right of the track: x = bx + bw - 1 = 15.
	_, _ = s.Update(tea.MouseMsg{
		X:      15,
		Y:      2,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	if s.Value() != 100 {
		t.Errorf("click at far right: value = %v, want 100", s.Value())
	}

	// Click at the middle: offset 5, ratio 5/10 -> 50.
	_, _ = s.Update(tea.MouseMsg{
		X:      10,
		Y:      2,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	if s.Value() != 50 {
		t.Errorf("click at middle: value = %v, want 50", s.Value())
	}
}

func TestSliderImplementsComponentAndBounded(t *testing.T) {
	var s any = NewSlider()
	if _, ok := s.(interface {
		Init() tea.Cmd
		View() string
		Focus()
		Blur()
		Focused() bool
	}); !ok {
		t.Fatalf("Slider must satisfy tui.Component shape")
	}
	if _, ok := s.(interface{ Bounds() (int, int, int, int) }); !ok {
		t.Fatalf("Slider must implement Bounded")
	}
}

// stripANSI returns s with ANSI CSI escape sequences removed. Sufficient
// for the SGR sequences emitted by Slider.View.
func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		if inEsc {
			if r == 'm' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		if r == 0x1b {
			inEsc = true
			continue
		}
		if r == '[' {
			// stripANSI is conservative — only the actual ESC start a CSI.
		}
		b.WriteRune(r)
	}
	return b.String()
}
