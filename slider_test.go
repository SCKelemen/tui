package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSlider(t *testing.T) {
	s := NewSlider()
	if s == nil {
		t.Fatal("NewSlider returned nil")
	}
	if s.min != 0 || s.max != 1 || s.step != 0.1 {
		t.Fatalf("unexpected defaults: min=%v max=%v step=%v", s.min, s.max, s.step)
	}
	if s.width != 10 || s.orientation != SliderHorizontal {
		t.Fatalf("unexpected width/orientation: width=%d orientation=%v", s.width, s.orientation)
	}
}

func TestSliderView(t *testing.T) {
	s := NewSlider(
		WithSliderRange(0, 10),
		WithSliderValue(5),
		WithSliderStep(1),
		WithSliderWidth(5),
	)

	view := stripANSI(s.View())
	if !strings.HasPrefix(view, "[") || !strings.HasSuffix(view, "]") {
		t.Fatalf("View() = %q, want bracketed slider", view)
	}
	if !strings.Contains(view, "█") || !strings.Contains(view, "░") {
		t.Fatalf("View() = %q, want filled and empty cells", view)
	}
}

func TestSliderKeyboard(t *testing.T) {
	s := NewSlider(
		WithSliderRange(0, 10),
		WithSliderValue(5),
		WithSliderStep(1),
	)
	s.Focus()

	s.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if got := s.Value(); got != 4 {
		t.Fatalf("after left key Value() = %v, want 4", got)
	}

	s.Update(tea.KeyMsg{Type: tea.KeyRight})
	if got := s.Value(); got != 5 {
		t.Fatalf("after right key Value() = %v, want 5", got)
	}
}

func TestSliderValue(t *testing.T) {
	s := NewSlider(
		WithSliderRange(10, 0),
		WithSliderValue(11),
		WithSliderStep(0.5),
		WithSliderWidth(11),
	)

	if got := s.Value(); got != 10 {
		t.Fatalf("normalized/clamped Value() = %v, want 10", got)
	}

	s.SetValue(4.6)
	if got := s.Value(); got != 4.5 {
		t.Fatalf("snapped Value() = %v, want 4.5", got)
	}
	if got := s.NormalizedValue(); got != 0.45 {
		t.Fatalf("NormalizedValue() = %v, want 0.45", got)
	}
}
