package display

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// BigText renders text using 5-row block letter glyphs.
type BigText struct {
	text     string
	color    string
	gradient []string
	width    int
}

// BigTextOption configures a BigText component.
type BigTextOption func(*BigText)

// NewBigText creates a BigText component.
func NewBigText(text string, opts ...BigTextOption) *BigText {
	b := &BigText{
		text:  text,
		color: "#FFFFFF",
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// WithBigTextColor sets a single color for all characters.
func WithBigTextColor(color string) BigTextOption {
	return func(b *BigText) {
		b.color = color
	}
}

// WithBigTextGradient sets gradient colors to interpolate across characters.
func WithBigTextGradient(colors []string) BigTextOption {
	return func(b *BigText) {
		if len(colors) == 0 {
			b.gradient = nil
			return
		}

		b.gradient = append([]string(nil), colors...)
	}
}

// WithBigTextWidth sets the output width.
func WithBigTextWidth(w int) BigTextOption {
	return func(b *BigText) {
		b.width = w
	}
}

// View renders text using a block font made of full block characters.
func (b *BigText) View() string {
	if b.text == "" {
		return ""
	}

	runes := []rune(strings.ToUpper(b.text))
	rows := make([]strings.Builder, 5)

	for i, r := range runes {
		glyph, ok := blockFont[r]
		if !ok {
			glyph = blockFont['?']
		}

		charColor := b.color
		if len(b.gradient) > 0 {
			charColor = gradientColorAt(i, len(runes), b.gradient)
		}

		charStyle := lipgloss.NewStyle()
		if charColor != "" {
			charStyle = charStyle.Foreground(lipgloss.Color(charColor))
		}

		for row := range rows {
			if i > 0 {
				rows[row].WriteByte(' ')
			}
			rows[row].WriteString(charStyle.Render(glyph[row]))
		}
	}

	lines := make([]string, len(rows))
	for i := range rows {
		lines[i] = rows[i].String()
	}

	out := strings.Join(lines, "\n")
	if b.width > 0 {
		out = lipgloss.NewStyle().Width(b.width).Render(out)
	}

	return out
}

func gradientColorAt(index, total int, colors []string) string {
	if len(colors) == 0 {
		return ""
	}
	if len(colors) == 1 || total <= 1 {
		return colors[0]
	}

	t := float64(index) / float64(total-1)
	segments := len(colors) - 1
	position := t * float64(segments)
	segment := int(math.Floor(position))
	if segment >= segments {
		return colors[len(colors)-1]
	}

	localT := position - float64(segment)
	return interpolateHex(colors[segment], colors[segment+1], localT)
}

func interpolateHex(start, end string, t float64) string {
	sr, sg, sb, sok := parseHexColor(start)
	er, eg, eb, eok := parseHexColor(end)
	if !sok || !eok {
		if t < 0.5 {
			return start
		}
		return end
	}

	r := int(math.Round(float64(sr) + (float64(er)-float64(sr))*t))
	g := int(math.Round(float64(sg) + (float64(eg)-float64(sg))*t))
	bl := int(math.Round(float64(sb) + (float64(eb)-float64(sb))*t))

	return fmt.Sprintf("#%02X%02X%02X", clampColor(r), clampColor(g), clampColor(bl))
}

func parseHexColor(color string) (int, int, int, bool) {
	hex := strings.TrimPrefix(strings.TrimSpace(color), "#")

	switch len(hex) {
	case 3:
		r, errR := strconv.ParseUint(strings.Repeat(string(hex[0]), 2), 16, 8)
		g, errG := strconv.ParseUint(strings.Repeat(string(hex[1]), 2), 16, 8)
		b, errB := strconv.ParseUint(strings.Repeat(string(hex[2]), 2), 16, 8)
		if errR != nil || errG != nil || errB != nil {
			return 0, 0, 0, false
		}
		return int(r), int(g), int(b), true
	case 6:
		r, errR := strconv.ParseUint(hex[0:2], 16, 8)
		g, errG := strconv.ParseUint(hex[2:4], 16, 8)
		b, errB := strconv.ParseUint(hex[4:6], 16, 8)
		if errR != nil || errG != nil || errB != nil {
			return 0, 0, 0, false
		}
		return int(r), int(g), int(b), true
	default:
		return 0, 0, 0, false
	}
}

func clampColor(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

var blockFont = map[rune][5]string{
	' ': {"    ", "    ", "    ", "    ", "    "},
	'!': {" ██ ", " ██ ", " ██ ", "    ", " ██ "},
	'"': {"█ █ ", "█ █ ", "    ", "    ", "    "},
	'#': {" █ █ ", "█████", " █ █ ", "█████", " █ █ "},
	',': {"    ", "    ", "    ", " ██ ", "██  "},
	'-': {"     ", "     ", "█████", "     ", "     "},
	'.': {"    ", "    ", "    ", "    ", " ██ "},
	'/': {"    █", "   █ ", "  █  ", " █   ", "█    "},
	'0': {"████ ", "█  █ ", "█  █ ", "█  █ ", "████ "},
	'1': {" ██  ", "███  ", " ██  ", " ██  ", "████ "},
	'2': {"████ ", "   █ ", "████ ", "█    ", "████ "},
	'3': {"████ ", "   █ ", "████ ", "   █ ", "████ "},
	'4': {"█  █ ", "█  █ ", "█████", "   █ ", "   █ "},
	'5': {"█████", "█    ", "████ ", "   █ ", "████ "},
	'6': {" ███ ", "█    ", "████ ", "█  █ ", " ███ "},
	'7': {"█████", "   █ ", "  █  ", " █   ", " █   "},
	'8': {" ███ ", "█  █ ", " ███ ", "█  █ ", " ███ "},
	'9': {" ███ ", "█  █ ", " ████", "   █ ", " ███ "},
	':': {"    ", " ██ ", "    ", " ██ ", "    "},
	'?': {"████ ", "   █ ", " ██  ", "     ", " ██  "},
	'@': {" ████ ", "█   █ ", "█ ███ ", "█     ", " ████ "},
	'A': {" ███ ", "█   █", "█████", "█   █", "█   █"},
	'B': {"████ ", "█   █", "████ ", "█   █", "████ "},
	'C': {" ████", "█    ", "█    ", "█    ", " ████"},
	'D': {"████ ", "█   █", "█   █", "█   █", "████ "},
	'E': {"█████", "█    ", "████ ", "█    ", "█████"},
	'F': {"█████", "█    ", "████ ", "█    ", "█    "},
	'G': {" ████", "█    ", "█ ███", "█   █", " ███ "},
	'H': {"█   █", "█   █", "█████", "█   █", "█   █"},
	'I': {"████ ", " ██  ", " ██  ", " ██  ", "████ "},
	'J': {"█████", "   █ ", "   █ ", "█  █ ", " ██  "},
	'K': {"█   █", "█  █ ", "███  ", "█  █ ", "█   █"},
	'L': {"█    ", "█    ", "█    ", "█    ", "█████"},
	'M': {"█   █", "██ ██", "█ █ █", "█   █", "█   █"},
	'N': {"█   █", "██  █", "█ █ █", "█  ██", "█   █"},
	'O': {" ███ ", "█   █", "█   █", "█   █", " ███ "},
	'P': {"████ ", "█   █", "████ ", "█    ", "█    "},
	'Q': {" ███ ", "█   █", "█   █", "█  ██", " ████"},
	'R': {"████ ", "█   █", "████ ", "█  █ ", "█   █"},
	'S': {" ████", "█    ", " ███ ", "    █", "████ "},
	'T': {"█████", "  █  ", "  █  ", "  █  ", "  █  "},
	'U': {"█   █", "█   █", "█   █", "█   █", " ███ "},
	'V': {"█   █", "█   █", "█   █", " █ █ ", "  █  "},
	'W': {"█   █", "█   █", "█ █ █", "██ ██", "█   █"},
	'X': {"█   █", " █ █ ", "  █  ", " █ █ ", "█   █"},
	'Y': {"█   █", " █ █ ", "  █  ", "  █  ", "  █  "},
	'Z': {"█████", "   █ ", "  █  ", " █   ", "█████"},
	'_': {"     ", "     ", "     ", "     ", "█████"},
}
