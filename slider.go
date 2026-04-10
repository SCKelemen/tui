package tui

import (
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SliderOrientation controls the direction a slider renders and tracks mouse movement.
type SliderOrientation int

const (
	SliderHorizontal SliderOrientation = iota
	SliderVertical
)

// Slider is an interactive range control with keyboard and mouse support.
type Slider struct {
	value       float64
	min         float64
	max         float64
	step        float64
	width       int
	orientation SliderOrientation

	trackStyle  string
	thumbStyle  string
	activeStyle string

	focused  bool
	dragging bool
	x        int
	y        int
	onChange func(float64) tea.Cmd
}

// SliderOption configures a Slider.
type SliderOption func(*Slider)

// WithSliderValue sets the initial value.
func WithSliderValue(value float64) SliderOption {
	return func(s *Slider) {
		s.value = value
	}
}

// WithSliderRange sets the minimum and maximum values.
func WithSliderRange(minValue, maxValue float64) SliderOption {
	return func(s *Slider) {
		s.min = minValue
		s.max = maxValue
	}
}

// WithSliderStep sets the keyboard and snap step size.
func WithSliderStep(step float64) SliderOption {
	return func(s *Slider) {
		s.step = step
	}
}

// WithSliderWidth sets the rendered width or height of the slider track.
func WithSliderWidth(width int) SliderOption {
	return func(s *Slider) {
		if width > 0 {
			s.width = width
		}
	}
}

// WithSliderOrientation sets the slider orientation.
func WithSliderOrientation(orientation SliderOrientation) SliderOption {
	return func(s *Slider) {
		s.orientation = orientation
	}
}

// WithSliderTrackStyle sets the ANSI style applied to inactive track cells.
func WithSliderTrackStyle(style string) SliderOption {
	return func(s *Slider) {
		s.trackStyle = style
	}
}

// WithSliderThumbStyle sets the ANSI style applied to the thumb cell.
func WithSliderThumbStyle(style string) SliderOption {
	return func(s *Slider) {
		s.thumbStyle = style
	}
}

// WithSliderActiveStyle sets the ANSI style applied to active track cells.
func WithSliderActiveStyle(style string) SliderOption {
	return func(s *Slider) {
		s.activeStyle = style
	}
}

// WithSliderOnChange sets the callback invoked whenever the slider value changes.
func WithSliderOnChange(fn func(float64) tea.Cmd) SliderOption {
	return func(s *Slider) {
		s.onChange = fn
	}
}

// NewSlider creates a new slider.
func NewSlider(opts ...SliderOption) *Slider {
	s := &Slider{
		min:         0,
		max:         1,
		step:        0.1,
		width:       10,
		orientation: SliderHorizontal,
		trackStyle:  ansiDim,
		activeStyle: ansiGreen,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.normalizeRange()
	s.value = s.clampAndSnap(s.value)

	return s
}

// Init initializes the slider.
func (s *Slider) Init() tea.Cmd {
	return nil
}

// Update handles keyboard and mouse input.
func (s *Slider) Update(msg tea.Msg) (Component, tea.Cmd) {
	s.normalizeRange()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !s.focused {
			return s, nil
		}

		switch msg.Type {
		case tea.KeyLeft:
			return s, s.adjust(-s.effectiveStep())
		case tea.KeyRight:
			return s, s.adjust(s.effectiveStep())
		case tea.KeyDown:
			if s.orientation == SliderVertical {
				return s, s.adjust(-s.effectiveStep())
			}
		case tea.KeyUp:
			if s.orientation == SliderVertical {
				return s, s.adjust(s.effectiveStep())
			}
		}
	case tea.MouseMsg:
		return s, s.handleMouse(msg)
	}

	return s, nil
}

// View renders the slider.
func (s *Slider) View() string {
	if s.width <= 0 {
		return "[]"
	}

	if s.orientation == SliderVertical {
		return s.verticalView()
	}
	return s.horizontalView()
}

// Focus marks the slider as focused.
func (s *Slider) Focus() {
	s.focused = true
}

// Blur marks the slider as unfocused and ends any active drag.
func (s *Slider) Blur() {
	s.focused = false
	s.dragging = false
}

// Focused reports whether the slider has focus.
func (s *Slider) Focused() bool {
	return s.focused
}

// SetPosition sets the slider's top-left screen coordinate for mouse hit testing.
func (s *Slider) SetPosition(x, y int) {
	s.x = x
	s.y = y
}

// SetRange sets the slider range and clamps the current value.
func (s *Slider) SetRange(minValue, maxValue float64) {
	s.min = minValue
	s.max = maxValue
	s.normalizeRange()
	s.value = s.clampAndSnap(s.value)
}

// SetStep sets the slider step size.
func (s *Slider) SetStep(step float64) {
	s.step = step
}

// SetWidth sets the slider width.
func (s *Slider) SetWidth(width int) {
	if width > 0 {
		s.width = width
	}
}

// SetOrientation sets the slider orientation.
func (s *Slider) SetOrientation(orientation SliderOrientation) {
	s.orientation = orientation
}

// SetStyles updates the inactive track, thumb, and active track ANSI styles.
func (s *Slider) SetStyles(trackStyle, thumbStyle, activeStyle string) {
	s.trackStyle = trackStyle
	s.thumbStyle = thumbStyle
	s.activeStyle = activeStyle
}

// OnChange sets the change callback.
func (s *Slider) OnChange(fn func(float64) tea.Cmd) {
	s.onChange = fn
}

// Value returns the current slider value.
func (s *Slider) Value() float64 {
	return s.value
}

// SetValue sets the current slider value.
func (s *Slider) SetValue(value float64) {
	_ = s.updateValue(value)
}

// NormalizedValue returns the current value normalized to [0, 1].
func (s *Slider) NormalizedValue() float64 {
	s.normalizeRange()
	if s.max <= s.min {
		return 0
	}
	return (s.value - s.min) / (s.max - s.min)
}

func (s *Slider) horizontalView() string {
	filledWidth := int(math.Floor(s.NormalizedValue() * float64(s.width)))
	thumbIndex := s.thumbIndex()

	var b strings.Builder
	b.WriteByte('[')
	for i := 0; i < s.width; i++ {
		switch {
		case i == thumbIndex:
			b.WriteString(s.thumbANSI())
			b.WriteRune('█')
			b.WriteString(ansiReset)
		case i < filledWidth:
			b.WriteString(s.activeANSI())
			b.WriteRune('█')
			b.WriteString(ansiReset)
		default:
			b.WriteString(s.trackANSI())
			b.WriteRune('░')
			b.WriteString(ansiReset)
		}
	}
	b.WriteByte(']')
	return b.String()
}

func (s *Slider) verticalView() string {
	filledHeight := int(math.Floor(s.NormalizedValue() * float64(s.width)))
	thumbIndex := s.thumbIndex()

	var b strings.Builder
	for row := s.width - 1; row >= 0; row-- {
		b.WriteByte('[')
		switch {
		case row == thumbIndex:
			b.WriteString(s.thumbANSI())
			b.WriteRune('█')
			b.WriteString(ansiReset)
		case row < filledHeight:
			b.WriteString(s.activeANSI())
			b.WriteRune('█')
			b.WriteString(ansiReset)
		default:
			b.WriteString(s.trackANSI())
			b.WriteRune('░')
			b.WriteString(ansiReset)
		}
		b.WriteByte(']')
		if row > 0 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (s *Slider) handleMouse(msg tea.MouseMsg) tea.Cmd {
	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button != tea.MouseButtonLeft || !s.contains(msg.X, msg.Y) {
			return nil
		}
		s.dragging = true
		return s.setValueFromPosition(msg.X, msg.Y)
	case tea.MouseActionMotion:
		if !s.dragging {
			return nil
		}
		return s.setValueFromPosition(msg.X, msg.Y)
	case tea.MouseActionRelease:
		if !s.dragging {
			return nil
		}
		s.dragging = false
		return s.setValueFromPosition(msg.X, msg.Y)
	}

	return nil
}

func (s *Slider) contains(x, y int) bool {
	if s.orientation == SliderVertical {
		return x >= s.x && x < s.x+3 && y >= s.y && y < s.y+s.width
	}
	return y == s.y && x >= s.x && x < s.x+s.width+2
}

func (s *Slider) setValueFromPosition(x, y int) tea.Cmd {
	if s.width <= 1 {
		return s.updateValue(s.max)
	}

	var ratio float64
	if s.orientation == SliderVertical {
		cell := (s.y + s.width - 1) - y
		if cell < 0 {
			cell = 0
		}
		if cell > s.width-1 {
			cell = s.width - 1
		}
		ratio = float64(cell) / float64(s.width-1)
	} else {
		cell := x - s.x - 1
		if cell < 0 {
			cell = 0
		}
		if cell > s.width-1 {
			cell = s.width - 1
		}
		ratio = float64(cell) / float64(s.width-1)
	}

	value := s.min + ratio*(s.max-s.min)
	return s.updateValue(value)
}

func (s *Slider) adjust(delta float64) tea.Cmd {
	if delta == 0 {
		return nil
	}
	return s.updateValue(s.value + delta)
}

func (s *Slider) updateValue(value float64) tea.Cmd {
	next := s.clampAndSnap(value)
	if math.Abs(next-s.value) < 1e-9 {
		return nil
	}

	s.value = next
	if s.onChange != nil {
		return s.onChange(s.value)
	}
	return nil
}

func (s *Slider) effectiveStep() float64 {
	if s.step > 0 {
		return s.step
	}
	if s.max <= s.min {
		return 0
	}
	if s.width <= 1 {
		return s.max - s.min
	}
	return (s.max - s.min) / float64(s.width-1)
}

func (s *Slider) clampAndSnap(value float64) float64 {
	s.normalizeRange()
	if value < s.min {
		value = s.min
	}
	if value > s.max {
		value = s.max
	}

	step := math.Abs(s.step)
	if step <= 0 {
		return value
	}

	snapped := s.min + math.Round((value-s.min)/step)*step
	if snapped < s.min {
		return s.min
	}
	if snapped > s.max {
		return s.max
	}
	return snapped
}

func (s *Slider) normalizeRange() {
	if s.max < s.min {
		s.min, s.max = s.max, s.min
	}
}

func (s *Slider) thumbIndex() int {
	if s.width <= 1 {
		return 0
	}
	index := int(math.Round(s.NormalizedValue() * float64(s.width-1)))
	if index < 0 {
		return 0
	}
	if index >= s.width {
		return s.width - 1
	}
	return index
}

func (s *Slider) trackANSI() string {
	if s.trackStyle != "" {
		return s.trackStyle
	}
	return ansiDim
}

func (s *Slider) activeANSI() string {
	if s.activeStyle != "" {
		return s.activeStyle
	}
	return ansiGreen
}

func (s *Slider) thumbANSI() string {
	if s.thumbStyle != "" {
		return s.thumbStyle
	}
	if s.focused {
		return s.activeANSI() + ansiBold + ansiInverse
	}
	return s.activeANSI() + ansiBold
}

// Ensure compile-time interface compliance.
var _ Component = (*Slider)(nil)
