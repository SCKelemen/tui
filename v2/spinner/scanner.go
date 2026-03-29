package spinner

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ScannerStyle defines the glyph family used by the scanner.
type ScannerStyle int

const (
	// ScannerBlocks uses square block glyphs.
	ScannerBlocks ScannerStyle = iota
	// ScannerDiamonds uses diamond glyphs.
	ScannerDiamonds
)

// ScannerDirection controls how the scanner moves across the track.
type ScannerDirection int

const (
	// ScannerBidirectional bounces left-to-right and right-to-left.
	ScannerBidirectional ScannerDirection = iota
	// ScannerForward continuously moves from left to right.
	ScannerForward
	// ScannerBackward continuously moves from right to left.
	ScannerBackward
)

// ScannerConfig configures the Knight Rider-style scanner.
type ScannerConfig struct {
	// Width is the number of cells in the scanner track.
	Width int
	// Style picks the active/inactive glyph set.
	Style ScannerStyle
	// Direction controls the scanner movement pattern.
	Direction ScannerDirection
	// ActiveColor is the primary scanner color in hex format.
	ActiveColor string
	// TrailLength is the number of cells behind the active scanner cell.
	TrailLength int
	// HoldFrames is the number of frames to pause at each edge in bidirectional mode.
	HoldFrames int
	// FadeTrail controls whether trail intensity falls off with distance.
	FadeTrail bool
}

// Scanner renders and advances a gradient scanner animation.
type Scanner struct {
	cfg           ScannerConfig
	frame         int
	position      int
	movingForward bool
	holdCount     int
}

// NewScanner creates a scanner with sane defaults applied.
func NewScanner(cfg ScannerConfig) *Scanner {
	cfg = normalizeScannerConfig(cfg)

	s := &Scanner{
		cfg:           cfg,
		movingForward: true,
	}

	switch cfg.Direction {
	case ScannerBackward:
		s.position = cfg.Width - 1
		s.movingForward = false
	case ScannerForward:
		s.position = 0
		s.movingForward = true
	default:
		s.position = 0
		s.movingForward = true
	}

	return s
}

// DefaultScanner returns a scanner with practical defaults.
func DefaultScanner() *Scanner {
	return NewScanner(ScannerConfig{})
}

// Tick advances the scanner by one frame.
func (s *Scanner) Tick() {
	s.frame++
	if s.cfg.Width <= 1 {
		s.position = 0
		return
	}

	switch s.cfg.Direction {
	case ScannerForward:
		s.position++
		if s.position >= s.cfg.Width {
			s.position = 0
		}
	case ScannerBackward:
		s.position--
		if s.position < 0 {
			s.position = s.cfg.Width - 1
		}
	default:
		if s.holdCount > 0 {
			s.holdCount--
			return
		}

		if s.movingForward {
			s.position++
			if s.position >= s.cfg.Width-1 {
				s.position = s.cfg.Width - 1
				s.movingForward = false
				s.holdCount = s.cfg.HoldFrames
			}
		} else {
			s.position--
			if s.position <= 0 {
				s.position = 0
				s.movingForward = true
				s.holdCount = s.cfg.HoldFrames
			}
		}
	}
}

// View renders the current scanner frame with gradient trail coloring.
func (s *Scanner) View() string {
	if s.cfg.Width <= 0 {
		return ""
	}

	activeRune, inactiveRune := scannerRunes(s.cfg.Style)
	activeRGB := parseHexColor(s.cfg.ActiveColor)
	inactiveRGB := scaleColor(activeRGB, 0.15)

	var b strings.Builder
	b.Grow(s.cfg.Width * 12)

	trailForward := s.trailForward()

	for i := 0; i < s.cfg.Width; i++ {
		if i == s.position {
			b.WriteString(renderColor(activeRune, activeRGB))
			continue
		}

		distance := s.trailDistance(i, trailForward)
		if distance > 0 && distance <= s.cfg.TrailLength {
			trailRGB := s.trailColor(activeRGB, inactiveRGB, distance)
			b.WriteString(renderColor(activeRune, trailRGB))
			continue
		}

		b.WriteString(renderColor(inactiveRune, inactiveRGB))
	}

	return b.String()
}

func (s *Scanner) trailForward() bool {
	switch s.cfg.Direction {
	case ScannerForward:
		return true
	case ScannerBackward:
		return false
	default:
		return s.movingForward
	}
}

func (s *Scanner) trailDistance(index int, trailForward bool) int {
	if trailForward {
		if index >= s.position {
			return -1
		}
		return s.position - index
	}

	if index <= s.position {
		return -1
	}
	return index - s.position
}

func (s *Scanner) trailColor(active, inactive rgb, distance int) rgb {
	if s.cfg.TrailLength <= 0 {
		return inactive
	}
	if !s.cfg.FadeTrail {
		return blendColor(inactive, active, 0.55)
	}

	ratio := 1.0 - (float64(distance) / float64(s.cfg.TrailLength+1))
	if ratio < 0 {
		ratio = 0
	}
	return blendColor(inactive, active, ratio)
}

func normalizeScannerConfig(cfg ScannerConfig) ScannerConfig {
	isZeroConfig := cfg == (ScannerConfig{})

	if cfg.Width <= 0 {
		cfg.Width = 10
	}
	if cfg.ActiveColor == "" {
		cfg.ActiveColor = "#FFFFFF"
	}
	if cfg.TrailLength <= 0 {
		cfg.TrailLength = 3
	}
	if cfg.HoldFrames < 0 {
		cfg.HoldFrames = 0
	}

	// The zero-value config should use fading and a small edge hold by default.
	if isZeroConfig {
		cfg.FadeTrail = true
		cfg.HoldFrames = 2
	}

	if cfg.Direction != ScannerBidirectional && cfg.Direction != ScannerForward && cfg.Direction != ScannerBackward {
		cfg.Direction = ScannerBidirectional
	}
	if cfg.Style != ScannerBlocks && cfg.Style != ScannerDiamonds {
		cfg.Style = ScannerBlocks
	}

	return cfg
}

func scannerRunes(style ScannerStyle) (active string, inactive string) {
	switch style {
	case ScannerDiamonds:
		return "◆", "◇"
	default:
		return "■", "□"
	}
}

type rgb struct {
	r int
	g int
	b int
}

func parseHexColor(hex string) rgb {
	trimmed := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(trimmed) != 6 {
		return rgb{r: 255, g: 255, b: 255}
	}

	r, errR := strconv.ParseInt(trimmed[0:2], 16, 32)
	g, errG := strconv.ParseInt(trimmed[2:4], 16, 32)
	b, errB := strconv.ParseInt(trimmed[4:6], 16, 32)
	if errR != nil || errG != nil || errB != nil {
		return rgb{r: 255, g: 255, b: 255}
	}

	return rgb{r: int(r), g: int(g), b: int(b)}
}

func scaleColor(c rgb, factor float64) rgb {
	if factor < 0 {
		factor = 0
	}
	return rgb{
		r: clampColor(float64(c.r) * factor),
		g: clampColor(float64(c.g) * factor),
		b: clampColor(float64(c.b) * factor),
	}
}

func blendColor(from, to rgb, ratio float64) rgb {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	return rgb{
		r: clampColor(float64(from.r) + (float64(to.r)-float64(from.r))*ratio),
		g: clampColor(float64(from.g) + (float64(to.g)-float64(from.g))*ratio),
		b: clampColor(float64(from.b) + (float64(to.b)-float64(from.b))*ratio),
	}
}

func clampColor(v float64) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return int(v + 0.5)
}

func renderColor(glyph string, c rgb) string {
	hex := fmt.Sprintf("#%02X%02X%02X", c.r, c.g, c.b)
	return lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Render(glyph)
}
