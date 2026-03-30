package display

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// RatingDots displays a row of filled and empty dots (for example: ●●●○○)
// with an optional label.
type RatingDots struct {
	value      int
	max        int
	label      string
	color      string
	emptyColor string
	bold       bool

	focused bool
	tokens  *design.DesignTokens
}

// RatingDotsOption configures a RatingDots component.
type RatingDotsOption func(*RatingDots)

// WithRatingDotsColor sets the hex color for filled dots.
func WithRatingDotsColor(color string) RatingDotsOption {
	return func(r *RatingDots) {
		r.color = color
	}
}

// WithRatingDotsEmptyColor sets the hex color for empty dots.
func WithRatingDotsEmptyColor(color string) RatingDotsOption {
	return func(r *RatingDots) {
		r.emptyColor = color
	}
}

// WithRatingDotsLabel sets the optional label shown before the dots.
func WithRatingDotsLabel(label string) RatingDotsOption {
	return func(r *RatingDots) {
		r.label = label
	}
}

// WithRatingDotsDesignTokens applies design tokens.
func WithRatingDotsDesignTokens(tokens *design.DesignTokens) RatingDotsOption {
	return func(r *RatingDots) {
		if tokens != nil {
			r.tokens = tokens
		}
	}
}

// WithRatingDotsBold enables or disables bold styling for filled dots.
func WithRatingDotsBold(bold bool) RatingDotsOption {
	return func(r *RatingDots) {
		r.bold = bold
	}
}

// NewRatingDots creates a new RatingDots component.
func NewRatingDots(value int, max int, opts ...RatingDotsOption) *RatingDots {
	r := &RatingDots{
		value:      value,
		max:        max,
		color:      "#4CAF50",
		emptyColor: "#808080",
		tokens:     design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Init initializes the component.
func (r *RatingDots) Init() tea.Cmd {
	return nil
}

// Update handles Bubble Tea messages.
func (r *RatingDots) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	return r, nil
}

// View renders the rating dots with optional label.
func (r *RatingDots) View() string {
	max := r.max
	if max < 0 {
		max = 0
	}

	value := r.value
	if value < 0 {
		value = 0
	}
	if value > max {
		value = max
	}

	filledColor := style.ANSIColorFromHex(r.color)
	emptyColor := style.ANSIColorFromHex(r.emptyColor)

	var b strings.Builder
	if r.label != "" {
		b.WriteString(r.label)
		b.WriteString("  ")
	}

	for i := 0; i < max; i++ {
		if i < value {
			if filledColor != "" {
				b.WriteString(filledColor)
			}
			if r.bold {
				b.WriteString(style.ANSIBold)
			}
			b.WriteString("●")
			if filledColor != "" || r.bold {
				b.WriteString(style.ANSIReset)
			}
		} else {
			if emptyColor != "" {
				b.WriteString(emptyColor)
			}
			b.WriteString("○")
			if emptyColor != "" {
				b.WriteString(style.ANSIReset)
			}
		}
	}

	return b.String()
}

// Focus marks the component as focused.
func (r *RatingDots) Focus() {
	r.focused = true
}

// Blur marks the component as unfocused.
func (r *RatingDots) Blur() {
	r.focused = false
}

// Focused reports whether the component is focused.
func (r *RatingDots) Focused() bool {
	return r.focused
}

var _ tui.Component = (*RatingDots)(nil)
