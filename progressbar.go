package tui

import (
	"fmt"
	"strings"

	colorpkg "github.com/SCKelemen/color"
	design "github.com/SCKelemen/design-system"
	tea "github.com/charmbracelet/bubbletea"
)

// ProgressBarStyle defines visual styles for the progress bar.
type ProgressBarStyle int

const (
	ProgressBarBlock ProgressBarStyle = iota
	ProgressBarTick
	ProgressBarClassic
	ProgressBarThin
)

// ProgressBar renders a horizontal progress bar with optional gradient coloring.
type ProgressBar struct {
	progress    float64
	width       int
	label       string
	showPercent bool
	style       ProgressBarStyle

	useGradient   bool
	gradientStart colorpkg.Color
	gradientEnd   colorpkg.Color
	gradientStops []colorpkg.GradientStop

	filledChar rune
	emptyChar  rune

	filledCharSet bool
	emptyCharSet  bool

	focused      bool
	designTokens *design.DesignTokens
}

// ProgressBarOption configures a ProgressBar.
type ProgressBarOption func(*ProgressBar)

// WithProgressBarWidth sets the bar width in characters.
func WithProgressBarWidth(width int) ProgressBarOption {
	return func(pb *ProgressBar) {
		if width > 0 {
			pb.width = width
		}
	}
}

// WithProgressBarLabel sets the right-side label.
func WithProgressBarLabel(label string) ProgressBarOption {
	return func(pb *ProgressBar) {
		pb.label = label
	}
}

// WithProgressBarShowPercent toggles percent rendering.
func WithProgressBarShowPercent(show bool) ProgressBarOption {
	return func(pb *ProgressBar) {
		pb.showPercent = show
	}
}

// WithProgressBarStyle sets the bar style.
func WithProgressBarStyle(style ProgressBarStyle) ProgressBarOption {
	return func(pb *ProgressBar) {
		pb.style = style
		pb.applyStyleDefaults()
	}
}

// WithProgressBarGradient configures a two-stop gradient.
func WithProgressBarGradient(start, end colorpkg.Color) ProgressBarOption {
	return func(pb *ProgressBar) {
		pb.useGradient = true
		pb.gradientStart = start
		pb.gradientEnd = end
		pb.gradientStops = nil
	}
}

// WithProgressBarMultiStopGradient configures a multi-stop gradient.
func WithProgressBarMultiStopGradient(stops []colorpkg.GradientStop) ProgressBarOption {
	return func(pb *ProgressBar) {
		pb.useGradient = len(stops) > 0
		if len(stops) == 0 {
			pb.gradientStops = nil
			return
		}
		pb.gradientStops = append([]colorpkg.GradientStop(nil), stops...)
	}
}

// WithProgressBarFilledChar sets a custom filled segment character.
func WithProgressBarFilledChar(r rune) ProgressBarOption {
	return func(pb *ProgressBar) {
		if r == 0 {
			return
		}
		pb.filledChar = r
		pb.filledCharSet = true
	}
}

// WithProgressBarEmptyChar sets a custom empty segment character.
func WithProgressBarEmptyChar(r rune) ProgressBarOption {
	return func(pb *ProgressBar) {
		if r == 0 {
			return
		}
		pb.emptyChar = r
		pb.emptyCharSet = true
	}
}

// WithProgressBarDesignTokens applies design-system tokens.
func WithProgressBarDesignTokens(tokens *design.DesignTokens) ProgressBarOption {
	return func(pb *ProgressBar) {
		if tokens == nil {
			return
		}
		pb.designTokens = tokens
	}
}

// WithProgressBarTheme applies a named design-system theme.
func WithProgressBarTheme(theme string) ProgressBarOption {
	return func(pb *ProgressBar) {
		pb.designTokens = designTokensForTheme(theme)
	}
}

// NewProgressBar creates a new progress bar.
func NewProgressBar(progress float64, opts ...ProgressBarOption) *ProgressBar {
	pb := &ProgressBar{
		progress:     clampProgress(progress),
		width:        30,
		showPercent:  true,
		style:        ProgressBarBlock,
		designTokens: design.DefaultTheme(),
	}

	pb.applyStyleDefaults()

	for _, opt := range opts {
		opt(pb)
	}

	return pb
}

// Init initializes the progress bar.
func (pb *ProgressBar) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (pb *ProgressBar) Update(msg tea.Msg) (Component, tea.Cmd) {
	return pb, nil
}

// View renders the progress bar.
func (pb *ProgressBar) View() string {
	filledWidth := int(pb.progress * float64(pb.width))
	if filledWidth < 0 {
		filledWidth = 0
	}
	if filledWidth > pb.width {
		filledWidth = pb.width
	}
	emptyWidth := pb.width - filledWidth

	bar := pb.renderBody(filledWidth, emptyWidth)
	if pb.style == ProgressBarClassic {
		bar = "[" + bar + "]"
	}

	parts := []string{bar}
	if pb.showPercent {
		parts = append(parts, fmt.Sprintf("%d%%", int(pb.progress*100)))
	}
	if strings.TrimSpace(pb.label) != "" {
		parts = append(parts, pb.label)
	}

	return strings.Join(parts, " ")
}

// Focus marks the component as focused.
func (pb *ProgressBar) Focus() {
	pb.focused = true
}

// Blur marks the component as unfocused.
func (pb *ProgressBar) Blur() {
	pb.focused = false
}

// Focused returns whether the component is focused.
func (pb *ProgressBar) Focused() bool {
	return pb.focused
}

// SetProgress sets and clamps progress to [0, 1].
func (pb *ProgressBar) SetProgress(p float64) {
	pb.progress = clampProgress(p)
}

// SetLabel sets the right-side label.
func (pb *ProgressBar) SetLabel(s string) {
	pb.label = s
}

// GetProgress returns current progress.
func (pb *ProgressBar) GetProgress() float64 {
	return pb.progress
}

// Increment changes progress by delta and clamps to [0, 1].
func (pb *ProgressBar) Increment(delta float64) {
	pb.progress = clampProgress(pb.progress + delta)
}

func (pb *ProgressBar) renderBody(filledWidth, emptyWidth int) string {
	var b strings.Builder

	if filledWidth > 0 {
		if pb.useGradient {
			gradient := pb.gradientColors(filledWidth)
			for i := 0; i < filledWidth; i++ {
				colorCode := colorToANSIFg(gradient[i])
				b.WriteString(colorCode)
				b.WriteRune(pb.filledChar)
				b.WriteString(ansiReset)
			}
		} else {
			filledColor := pb.filledANSIColor()
			if filledColor != "" {
				b.WriteString(filledColor)
			}
			b.WriteString(strings.Repeat(string(pb.filledChar), filledWidth))
			b.WriteString(ansiReset)
		}
	}

	if emptyWidth > 0 {
		b.WriteString(ansiDim)
		b.WriteString(strings.Repeat(string(pb.emptyChar), emptyWidth))
		b.WriteString(ansiReset)
	}

	return b.String()
}

func (pb *ProgressBar) gradientColors(steps int) []colorpkg.Color {
	if steps <= 0 {
		return nil
	}

	if len(pb.gradientStops) > 0 {
		colors := colorpkg.GradientMultiStop(pb.gradientStops, steps, colorpkg.GradientOKLCH)
		if len(colors) >= steps {
			return colors[:steps]
		}
		return padGradient(colors, steps)
	}

	start := pb.gradientStart
	end := pb.gradientEnd
	if start == nil {
		start = colorpkg.NewRGBA(0, 1, 0, 1)
	}
	if end == nil {
		end = start
	}

	colors := colorpkg.Gradient(start, end, steps)
	if len(colors) >= steps {
		return colors[:steps]
	}
	return padGradient(colors, steps)
}

func padGradient(colors []colorpkg.Color, steps int) []colorpkg.Color {
	if len(colors) == 0 {
		fallback := colorpkg.NewRGBA(0, 1, 0, 1)
		padded := make([]colorpkg.Color, steps)
		for i := range padded {
			padded[i] = fallback
		}
		return padded
	}

	padded := make([]colorpkg.Color, 0, steps)
	padded = append(padded, colors...)
	last := colors[len(colors)-1]
	for len(padded) < steps {
		padded = append(padded, last)
	}
	return padded
}

func (pb *ProgressBar) filledANSIColor() string {
	if pb.designTokens != nil {
		if accent := ansiColorFromHex(pb.designTokens.Accent); accent != "" {
			return accent
		}
	}
	return ansiGreen
}

func (pb *ProgressBar) applyStyleDefaults() {
	var defaultFilled rune
	var defaultEmpty rune

	switch pb.style {
	case ProgressBarTick:
		defaultFilled = '▰'
		defaultEmpty = '▱'
	case ProgressBarClassic:
		defaultFilled = '='
		defaultEmpty = ' '
	case ProgressBarThin:
		defaultFilled = '━'
		defaultEmpty = '╌'
	case ProgressBarBlock:
		fallthrough
	default:
		defaultFilled = '█'
		defaultEmpty = '░'
	}

	if !pb.filledCharSet {
		pb.filledChar = defaultFilled
	}
	if !pb.emptyCharSet {
		pb.emptyChar = defaultEmpty
	}
}

func clampProgress(p float64) float64 {
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return p
}

func colorToANSIFg(c colorpkg.Color) string {
	r, g, b, _ := c.RGBA()
	esc := string([]byte{0x1b})
	r8 := int(r * 255)
	g8 := int(g * 255)
	b8 := int(b * 255)
	if r8 < 0 {
		r8 = 0
	} else if r8 > 255 {
		r8 = 255
	}
	if g8 < 0 {
		g8 = 0
	} else if g8 > 255 {
		g8 = 255
	}
	if b8 < 0 {
		b8 = 0
	} else if b8 > 255 {
		b8 = 255
	}
	return fmt.Sprintf("%s[38;2;%d;%d;%dm", esc, r8, g8, b8)
}