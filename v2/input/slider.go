// Package input contains interactive input components.
//
// Slider is a horizontal numeric slider rendered with Unicode block
// characters and a single-cell knob. Keyboard navigation uses the
// arrow keys (or h/l), Home/End, and Page Up/Page Down. Mouse clicks
// inside the slider's bounds snap the knob to the click position via
// the Application's Bounded hit-testing path (see v2/tui.go).
package input

import (
	"fmt"
	"math"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// Slider visual constants.
const (
	sliderFilledChar = '█'
	sliderEmptyChar  = '░'
	sliderKnobChar   = '◉'
)

// Slider is an interactive numeric slider component.
type Slider struct {
	min   float64
	max   float64
	value float64
	step  float64

	width      int
	label      string
	showValue  bool
	focused    bool
	stepIsAuto bool

	// Render bounds, used by the Application hit-test path to route mouse
	// clicks to the slider. Width and Height reflect the rendered cell
	// extent of the track itself; callers set these via SetBounds after
	// layout.
	bx, by, bw, bh int

	designTokens *design.DesignTokens
	onChange     func(value float64) tea.Cmd
}

// SliderOption configures a Slider.
type SliderOption func(*Slider)

// WithSliderMin sets the slider's minimum value.
func WithSliderMin(min float64) SliderOption {
	return func(s *Slider) { s.min = min }
}

// WithSliderMax sets the slider's maximum value.
func WithSliderMax(max float64) SliderOption {
	return func(s *Slider) { s.max = max }
}

// WithSliderValue sets the slider's initial value. Clamped to [min, max]
// during NewSlider after all options have been applied.
func WithSliderValue(v float64) SliderOption {
	return func(s *Slider) { s.value = v }
}

// WithSliderStep sets the increment applied per key press. Defaults to
// (max-min)/100 when not set or when zero.
func WithSliderStep(step float64) SliderOption {
	return func(s *Slider) {
		if step > 0 {
			s.step = step
			s.stepIsAuto = false
		}
	}
}

// WithSliderWidth sets the total track width in cells. Default is 30.
func WithSliderWidth(w int) SliderOption {
	return func(s *Slider) {
		if w > 0 {
			s.width = w
		}
	}
}

// WithSliderShowValue toggles whether the numeric value is rendered
// next to the track.
func WithSliderShowValue(b bool) SliderOption {
	return func(s *Slider) { s.showValue = b }
}

// WithSliderLabel sets an optional label rendered to the left of the
// track.
func WithSliderLabel(s string) SliderOption {
	return func(sl *Slider) { sl.label = s }
}

// WithSliderDesignTokens applies design-system tokens for color theming.
func WithSliderDesignTokens(tokens *design.DesignTokens) SliderOption {
	return func(s *Slider) {
		if tokens == nil {
			return
		}
		s.designTokens = tokens
	}
}

// WithSliderOnChange registers a callback fired after every value change.
// The returned tea.Cmd, if any, is bubbled out of Update.
func WithSliderOnChange(fn func(value float64) tea.Cmd) SliderOption {
	return func(s *Slider) { s.onChange = fn }
}

// NewSlider creates a slider with the provided options.
func NewSlider(opts ...SliderOption) *Slider {
	s := &Slider{
		min:          0,
		max:          100,
		value:        0,
		width:        30,
		showValue:    true,
		designTokens: design.DefaultTheme(),
		stepIsAuto:   true,
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.max < s.min {
		s.max = s.min
	}
	if s.stepIsAuto {
		s.step = autoStep(s.min, s.max)
	}
	s.value = clampFloat(s.value, s.min, s.max)

	return s
}

// Value returns the current value.
func (s *Slider) Value() float64 { return s.value }

// SetValue clamps v to [min, max] and stores it. Does not fire OnChange —
// callers using SetValue programmatically can dispatch their own commands.
func (s *Slider) SetValue(v float64) {
	s.value = clampFloat(v, s.min, s.max)
}

// SetBounds records the slider's rendered screen rectangle, enabling the
// Application's mouse hit-test to refocus the slider and route clicks
// to its Update. The y dimension is treated as a single row (height = 1)
// unless h is positive.
func (s *Slider) SetBounds(x, y, w, h int) {
	s.bx = x
	s.by = y
	if w > 0 {
		s.bw = w
	}
	if h > 0 {
		s.bh = h
	} else if s.bh == 0 {
		s.bh = 1
	}
}

// Bounds returns the slider's rendered screen rectangle. Implements
// tui.Bounded so the Application can route mouse clicks to the slider.
func (s *Slider) Bounds() (int, int, int, int) {
	return s.bx, s.by, s.bw, s.bh
}

// Init satisfies tui.Component.
func (s *Slider) Init() tea.Cmd { return nil }

// Update handles keyboard and mouse messages. Mouse presses are
// processed regardless of focus state — the Application's hit-test path
// refocuses this component before forwarding the event, so by the time
// Update sees the press the slider is already focused.
func (s *Slider) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		if !s.focused {
			return s, nil
		}
		before := s.value
		switch m.String() {
		case "left", "h":
			s.value = clampFloat(s.value-s.step, s.min, s.max)
		case "right", "l":
			s.value = clampFloat(s.value+s.step, s.min, s.max)
		case "home":
			s.value = s.min
		case "end":
			s.value = s.max
		case "pgup", "pageup":
			s.value = clampFloat(s.value-10*s.step, s.min, s.max)
		case "pgdown", "pagedown":
			s.value = clampFloat(s.value+10*s.step, s.min, s.max)
		default:
			return s, nil
		}
		if s.value != before && s.onChange != nil {
			return s, s.onChange(s.value)
		}
		return s, nil

	case tea.MouseMsg:
		if m.Action != tea.MouseActionPress || m.Button != tea.MouseButtonLeft {
			return s, nil
		}
		if s.bw <= 0 {
			return s, nil
		}
		// Snap the value to the click's horizontal cell within the track.
		// The track width covers s.bw cells starting at s.bx.
		offset := m.X - s.bx
		if offset < 0 {
			offset = 0
		}
		if offset > s.bw-1 {
			offset = s.bw - 1
		}
		before := s.value
		denom := float64(s.bw - 1)
		if denom <= 0 {
			s.value = s.min
		} else {
			ratio := float64(offset) / denom
			s.value = clampFloat(s.min+ratio*(s.max-s.min), s.min, s.max)
		}
		if s.value != before && s.onChange != nil {
			return s, s.onChange(s.value)
		}
		return s, nil
	}

	return s, nil
}

// View renders the slider. The output has the form:
//
//	Volume: ████◉░░░░░ 40
//
// The label and value sections are omitted when empty.
func (s *Slider) View() string {
	track := s.renderTrack()

	var b strings.Builder
	if strings.TrimSpace(s.label) != "" {
		b.WriteString(s.label)
		b.WriteString(": ")
	}
	b.WriteString(track)
	if s.showValue {
		b.WriteString(" ")
		b.WriteString(formatValue(s.value))
	}
	return b.String()
}

// Focus marks the slider as focused.
func (s *Slider) Focus() { s.focused = true }

// Blur marks the slider as unfocused.
func (s *Slider) Blur() { s.focused = false }

// Focused reports whether the slider is currently focused.
func (s *Slider) Focused() bool { return s.focused }

func (s *Slider) renderTrack() string {
	if s.width <= 0 {
		return ""
	}

	knobPos := s.knobPosition()
	filled := knobPos
	empty := s.width - knobPos - 1
	if empty < 0 {
		empty = 0
	}

	accent := s.accentANSI()

	var b strings.Builder
	if filled > 0 {
		if accent != "" {
			b.WriteString(accent)
		}
		b.WriteString(strings.Repeat(string(sliderFilledChar), filled))
		b.WriteString(style.ANSIReset)
	}

	// Knob — bold and accent-colored when focused, dim otherwise.
	if s.focused {
		b.WriteString(style.ANSIBold)
		if accent != "" {
			b.WriteString(accent)
		}
	} else {
		b.WriteString(style.ANSIDim)
	}
	b.WriteRune(sliderKnobChar)
	b.WriteString(style.ANSIReset)

	if empty > 0 {
		b.WriteString(style.ANSIDim)
		b.WriteString(strings.Repeat(string(sliderEmptyChar), empty))
		b.WriteString(style.ANSIReset)
	}

	return b.String()
}

func (s *Slider) knobPosition() int {
	if s.width <= 1 {
		return 0
	}
	span := s.max - s.min
	if span <= 0 {
		return 0
	}
	ratio := (s.value - s.min) / span
	pos := int(math.Round(ratio * float64(s.width-1)))
	if pos < 0 {
		pos = 0
	}
	if pos > s.width-1 {
		pos = s.width - 1
	}
	return pos
}

func (s *Slider) accentANSI() string {
	if s.designTokens == nil {
		return style.ANSICyan
	}
	if accent := style.ANSIColorFromHex(s.designTokens.Accent); accent != "" {
		return accent
	}
	return style.ANSICyan
}

func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func autoStep(min, max float64) float64 {
	span := max - min
	if span <= 0 {
		return 1
	}
	return span / 100
}

func formatValue(v float64) string {
	// Render whole numbers without a decimal point; otherwise trim
	// trailing zeros for compact output.
	if v == math.Trunc(v) {
		return fmt.Sprintf("%d", int64(v))
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.4f", v), "0"), ".")
}
