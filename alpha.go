package tui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const opaqueAlpha = 1.0

// RGBA represents an RGB color with alpha.
type RGBA struct {
	R uint8
	G uint8
	B uint8
	A float64
}

// FromHex converts a hex color into RGBA.
func FromHex(hex string) RGBA {
	r, g, b, ok := parseHexColor(strings.TrimSpace(hex))
	if !ok {
		return RGBA{}
	}

	return RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: opaqueAlpha,
	}
}

// FromLipgloss converts a Lip Gloss color into RGBA.
func FromLipgloss(c lipgloss.Color) RGBA {
	value := strings.TrimSpace(string(c))
	if value == "" {
		return RGBA{}
	}

	r, g, b, ok := colorToRGB(value)
	if !ok {
		return RGBA{}
	}

	return RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: opaqueAlpha,
	}
}

// ToHex returns the color as a #RRGGBB string.
func (c RGBA) ToHex() string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

// ToLipgloss converts the color into a Lip Gloss color.
func (c RGBA) ToLipgloss() lipgloss.Color {
	if clampOpacity(c.A) == 0 {
		return lipgloss.Color("")
	}
	return lipgloss.Color(c.ToHex())
}

// WithAlpha returns a copy of the color with the provided alpha.
func (c RGBA) WithAlpha(a float64) RGBA {
	c.A = clampOpacity(a)
	return c
}

// BlendColors applies Porter-Duff source-over compositing.
func BlendColors(fg, bg RGBA) RGBA {
	srcA := clampOpacity(fg.A)
	dstA := clampOpacity(bg.A)
	outA := srcA + dstA*(1-srcA)
	if outA <= 0 {
		return RGBA{}
	}

	blendChannel := func(src, dst uint8) uint8 {
		value := (float64(src)*srcA + float64(dst)*dstA*(1-srcA)) / outA
		return uint8(math.Round(clampFloat(value, 0, 255)))
	}

	return RGBA{
		R: blendChannel(fg.R, bg.R),
		G: blendChannel(fg.G, bg.G),
		B: blendChannel(fg.B, bg.B),
		A: outA,
	}
}

// BlendHex blends a foreground hex color over a background hex color.
func BlendHex(fgHex, bgHex string, fgAlpha float64) string {
	fg := FromHex(fgHex).WithAlpha(fgAlpha)
	bg := FromHex(bgHex)
	return BlendColors(fg, bg).ToHex()
}

// OpacityStack tracks nested opacity multipliers.
type OpacityStack struct {
	stack []float64
}

// Push multiplies the current opacity by the provided opacity and pushes it.
func (s *OpacityStack) Push(opacity float64) {
	s.ensureBase()
	s.stack = append(s.stack, s.Current()*clampOpacity(opacity))
}

// Pop removes the most recent opacity value.
func (s *OpacityStack) Pop() {
	s.ensureBase()
	if len(s.stack) > 1 {
		s.stack = s.stack[:len(s.stack)-1]
	}
}

// Current returns the effective current opacity.
func (s *OpacityStack) Current() float64 {
	s.ensureBase()
	return s.stack[len(s.stack)-1]
}

// ApplyToColor multiplies the current opacity into the color alpha.
func (s *OpacityStack) ApplyToColor(c RGBA) RGBA {
	return c.WithAlpha(c.A * s.Current())
}

// Depth returns the number of pushed opacity scopes.
func (s *OpacityStack) Depth() int {
	s.ensureBase()
	return len(s.stack) - 1
}

func (s *OpacityStack) ensureBase() {
	if len(s.stack) == 0 {
		s.stack = []float64{opaqueAlpha}
	}
}

// LayerCompositor composites multiple rendered layers with opacity.
type LayerCompositor struct {
	layers []compositorLayer
}

type compositorLayer struct {
	content string
	x       int
	y       int
	opacity float64
}

// AddLayer adds a rendered layer at the provided coordinates.
func (lc *LayerCompositor) AddLayer(content string, x, y int, opacity float64) {
	lc.layers = append(lc.layers, compositorLayer{
		content: content,
		x:       x,
		y:       y,
		opacity: clampOpacity(opacity),
	})
}

// Composite merges layers from bottom to top into a CellBuffer.
func (lc *LayerCompositor) Composite(width, height int) *CellBuffer {
	result := NewCellBuffer(width, height)
	if width <= 0 || height <= 0 || len(lc.layers) == 0 {
		return result
	}

	composited := newAlphaBuffer(width, height)
	for _, layer := range lc.layers {
		if layer.opacity <= 0 {
			continue
		}

		layerBuffer := renderLayerToBuffer(width, height, layer.content, layer.x, layer.y)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				idx := layerBuffer.index(x, y)
				cell := layerBuffer.back[idx]
				if cell.Width == 0 {
					continue
				}
				composited.SetCell(x, y, alphaCellFromCell(cell, layer.opacity))
			}
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := composited.CellAt(x, y)
			if cell.Width == 0 || !cell.needsRendering() {
				continue
			}
			result.SetCell(x, y, cell.toCell())
		}
	}

	return result
}

type alphaCell struct {
	Rune          rune
	Width         int
	Fg            RGBA
	Bg            RGBA
	hasFg         bool
	hasBg         bool
	Bold          bool
	Italic        bool
	Underline     bool
	Strikethrough bool
	Dim           bool
	Blink         bool
	Hyperlink     string
}

func alphaCellFromCell(cell Cell, opacity float64) alphaCell {
	alpha := clampOpacity(opacity)
	fgValue := strings.TrimSpace(string(cell.Fg))
	bgValue := strings.TrimSpace(string(cell.Bg))

	alphaCell := alphaCell{
		Rune:          cell.Rune,
		Width:         cell.Width,
		Bold:          cell.Bold,
		Italic:        cell.Italic,
		Underline:     cell.Underline,
		Strikethrough: cell.Strikethrough,
		Dim:           cell.Dim,
		Blink:         cell.Blink,
		Hyperlink:     cell.Hyperlink,
	}

	if fgValue != "" {
		alphaCell.Fg = FromLipgloss(cell.Fg).WithAlpha(alpha)
		alphaCell.hasFg = alphaCell.Fg.A > 0
	}
	if bgValue != "" {
		alphaCell.Bg = FromLipgloss(cell.Bg).WithAlpha(alpha)
		alphaCell.hasBg = alphaCell.Bg.A > 0
	}

	return normalizeAlphaCell(alphaCell)
}

func blankAlphaCell() alphaCell {
	return alphaCell{Rune: ' ', Width: 1}
}

func continuationAlphaCell(base alphaCell) alphaCell {
	base.Rune = 0
	base.Width = 0
	return base
}

func normalizeAlphaCell(cell alphaCell) alphaCell {
	if cell.Width <= 0 {
		if cell.Rune == 0 {
			return blankAlphaCell()
		}
		cell.Width = runewidth.RuneWidth(cell.Rune)
	}
	if cell.Width <= 0 {
		cell = blankAlphaCell()
	}
	if cell.Width > 2 {
		cell.Width = 1
	}
	if cell.Rune == 0 && cell.Width > 0 {
		cell.Rune = ' '
	}
	return cell
}

func renderAlphaRune(cell alphaCell) rune {
	if cell.Rune == 0 {
		return ' '
	}
	return cell.Rune
}

func (c alphaCell) hasVisibleGlyph() bool {
	return c.Width > 0 && renderAlphaRune(c) != ' '
}

func (c alphaCell) hasTextAttributes() bool {
	return c.Bold || c.Italic || c.Underline || c.Strikethrough || c.Dim || c.Blink || c.Hyperlink != ""
}

func (c alphaCell) isTransparent() bool {
	return c.Width > 0 && !c.hasVisibleGlyph() && !c.hasFg && !c.hasBg && !c.hasTextAttributes()
}

func (c alphaCell) needsRendering() bool {
	return c.hasVisibleGlyph() || c.hasFg || c.hasBg || c.hasTextAttributes() || c.Width == 0
}

func (c alphaCell) visibleColor() RGBA {
	if c.hasVisibleGlyph() && c.hasFg {
		return c.Fg
	}
	if c.hasBg {
		return c.Bg
	}
	return RGBA{}
}

func (c alphaCell) toCell() Cell {
	cell := Cell{
		Rune:          c.Rune,
		Width:         c.Width,
		Bold:          c.Bold,
		Italic:        c.Italic,
		Underline:     c.Underline,
		Strikethrough: c.Strikethrough,
		Dim:           c.Dim,
		Blink:         c.Blink,
		Hyperlink:     c.Hyperlink,
	}
	if c.hasFg && c.Fg.A > 0 {
		cell.Fg = c.Fg.ToLipgloss()
	}
	if c.hasBg && c.Bg.A > 0 {
		cell.Bg = c.Bg.ToLipgloss()
	}
	return cell
}

type alphaBuffer struct {
	width  int
	height int
	cells  []alphaCell
}

func newAlphaBuffer(width, height int) *alphaBuffer {
	size := 0
	if width > 0 && height > 0 {
		size = width * height
	}

	cells := make([]alphaCell, size)
	for i := range cells {
		cells[i] = blankAlphaCell()
	}

	return &alphaBuffer{
		width:  width,
		height: height,
		cells:  cells,
	}
}

func (ab *alphaBuffer) SetCell(x, y int, source alphaCell) {
	if !ab.inBounds(x, y) {
		return
	}

	source = normalizeAlphaCell(source)
	if source.Width == 2 && x >= ab.width-1 {
		return
	}

	ab.clearOverlapAt(x, y)
	if source.Width == 2 {
		ab.clearOverlapAt(x+1, y)
	}

	idx := ab.index(x, y)
	ab.cells[idx] = compositeAlphaCell(ab.cells[idx], source)

	if source.Width == 2 {
		ab.cells[idx+1] = compositeAlphaCell(ab.cells[idx+1], continuationAlphaCell(source))
	}
}

func (ab *alphaBuffer) CellAt(x, y int) alphaCell {
	if !ab.inBounds(x, y) {
		return blankAlphaCell()
	}
	return ab.cells[ab.index(x, y)]
}

func (ab *alphaBuffer) clearOverlapAt(x, y int) {
	if !ab.inBounds(x, y) {
		return
	}

	idx := ab.index(x, y)
	current := ab.cells[idx]
	if current.Width == 2 && x+1 < ab.width {
		ab.cells[ab.index(x+1, y)] = blankAlphaCell()
	}

	if current.Width == 0 && x > 0 {
		prevIdx := ab.index(x-1, y)
		if ab.cells[prevIdx].Width == 2 {
			ab.cells[prevIdx] = blankAlphaCell()
		}
	}

	if x > 0 {
		prevIdx := ab.index(x-1, y)
		if ab.cells[prevIdx].Width == 2 {
			ab.cells[prevIdx] = blankAlphaCell()
		}
	}

	ab.cells[idx] = blankAlphaCell()
}

func (ab *alphaBuffer) inBounds(x, y int) bool {
	return x >= 0 && x < ab.width && y >= 0 && y < ab.height
}

func (ab *alphaBuffer) index(x, y int) int {
	return y*ab.width + x
}

func compositeAlphaCell(dst, src alphaCell) alphaCell {
	if src.isTransparent() {
		return dst
	}

	out := dst
	if src.hasBg {
		out.Bg = BlendColors(src.Bg, dst.Bg)
		out.hasBg = out.Bg.A > 0
	}

	switch {
	case src.Width == 0:
		out = src
		if src.hasBg {
			out.Bg = BlendColors(src.Bg, dst.Bg)
			out.hasBg = out.Bg.A > 0
		} else {
			out.Bg = dst.Bg
			out.hasBg = dst.hasBg
		}
		return out
	case src.hasVisibleGlyph():
		out = src
		if src.hasBg {
			out.Bg = BlendColors(src.Bg, dst.Bg)
			out.hasBg = out.Bg.A > 0
		} else {
			out.Bg = dst.Bg
			out.hasBg = dst.hasBg
		}
		if src.hasFg {
			out.Fg = BlendColors(src.Fg, dst.visibleColor())
			out.hasFg = out.Fg.A > 0
		}
		return out
	case src.hasTextAttributes() || src.hasFg || src.Bg.A >= 0.999999:
		out = src
		if src.hasBg {
			out.Bg = BlendColors(src.Bg, dst.Bg)
			out.hasBg = out.Bg.A > 0
		} else {
			out.Bg = dst.Bg
			out.hasBg = dst.hasBg
		}
		if src.hasFg {
			out.Fg = BlendColors(src.Fg, dst.visibleColor())
			out.hasFg = out.Fg.A > 0
		}
		return out
	default:
		return out
	}
}

type ansiRenderState struct {
	fg            lipgloss.Color
	bg            lipgloss.Color
	bold          bool
	italic        bool
	underline     bool
	strikethrough bool
	dim           bool
	blink         bool
	hyperlink     string
}

func renderLayerToBuffer(width, height int, content string, startX, startY int) *CellBuffer {
	buffer := NewCellBuffer(width, height)
	state := ansiRenderState{}
	x := startX
	y := startY

	for i := 0; i < len(content); {
		if content[i] == '\x1b' {
			if consumed, ok := consumeEscapeSequence(content[i:], &state); ok {
				i += consumed
				continue
			}
		}

		r, size := utf8.DecodeRuneInString(content[i:])
		if r == utf8.RuneError && size == 1 {
			i++
			continue
		}

		switch r {
		case '\n':
			x = startX
			y++
			i += size
			continue
		case '\r':
			x = startX
			i += size
			continue
		case '\t':
			spaces := 4 - positiveMod(x-startX, 4)
			for j := 0; j < spaces; j++ {
				writeLayerRune(buffer, x, y, state, ' ')
				x++
			}
			i += size
			continue
		}

		width := runewidth.RuneWidth(r)
		if width <= 0 {
			i += size
			continue
		}
		if width > 2 {
			width = 1
		}

		writeLayerCell(buffer, x, y, state, r, width)
		x += width
		i += size
	}

	return buffer
}

func writeLayerRune(buffer *CellBuffer, x, y int, state ansiRenderState, r rune) {
	writeLayerCell(buffer, x, y, state, r, 1)
}

func writeLayerCell(buffer *CellBuffer, x, y int, state ansiRenderState, r rune, width int) {
	if y < 0 || y >= buffer.height || x >= buffer.width {
		return
	}
	if x < 0 || x+width > buffer.width {
		return
	}

	buffer.SetCell(x, y, Cell{
		Rune:          r,
		Width:         width,
		Fg:            state.fg,
		Bg:            state.bg,
		Bold:          state.bold,
		Italic:        state.italic,
		Underline:     state.underline,
		Strikethrough: state.strikethrough,
		Dim:           state.dim,
		Blink:         state.blink,
		Hyperlink:     state.hyperlink,
	})
}

func consumeEscapeSequence(s string, state *ansiRenderState) (int, bool) {
	if strings.HasPrefix(s, "\x1b[") {
		end := strings.IndexByte(s, 'm')
		if end < 0 {
			return 0, false
		}
		applySGR(state, s[2:end])
		return end + 1, true
	}

	if strings.HasPrefix(s, "\x1b]8;;") {
		payload := s[len("\x1b]8;;"):]
		if idx := strings.IndexByte(payload, '\a'); idx >= 0 {
			state.hyperlink = payload[:idx]
			return len("\x1b]8;;") + idx + 1, true
		}
		if idx := strings.Index(payload, "\x1b\\"); idx >= 0 {
			state.hyperlink = payload[:idx]
			return len("\x1b]8;;") + idx + 2, true
		}
	}

	return 0, false
}

func applySGR(state *ansiRenderState, raw string) {
	params := []int{0}
	if raw != "" {
		parts := strings.Split(raw, ";")
		params = make([]int, 0, len(parts))
		for _, part := range parts {
			if part == "" {
				params = append(params, 0)
				continue
			}
			value, err := strconv.Atoi(part)
			if err != nil {
				params = append(params, 0)
				continue
			}
			params = append(params, value)
		}
	}

	for i := 0; i < len(params); i++ {
		switch code := params[i]; {
		case code == 0:
			*state = ansiRenderState{}
		case code == 1:
			state.bold = true
		case code == 2:
			state.dim = true
		case code == 3:
			state.italic = true
		case code == 4:
			state.underline = true
		case code == 5:
			state.blink = true
		case code == 9:
			state.strikethrough = true
		case code == 22:
			state.bold = false
			state.dim = false
		case code == 23:
			state.italic = false
		case code == 24:
			state.underline = false
		case code == 25:
			state.blink = false
		case code == 29:
			state.strikethrough = false
		case code == 39:
			state.fg = lipgloss.Color("")
		case code == 49:
			state.bg = lipgloss.Color("")
		case code >= 30 && code <= 37:
			state.fg = lipgloss.Color(strconv.Itoa(code - 30))
		case code >= 40 && code <= 47:
			state.bg = lipgloss.Color(strconv.Itoa(code - 40))
		case code >= 90 && code <= 97:
			state.fg = lipgloss.Color(strconv.Itoa(code - 90 + 8))
		case code >= 100 && code <= 107:
			state.bg = lipgloss.Color(strconv.Itoa(code - 100 + 8))
		case code == 38 || code == 48:
			if i+1 >= len(params) {
				continue
			}
			mode := params[i+1]
			switch mode {
			case 2:
				if i+4 >= len(params) {
					continue
				}
				color := lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", clampColorComponent(params[i+2]), clampColorComponent(params[i+3]), clampColorComponent(params[i+4])))
				if code == 38 {
					state.fg = color
				} else {
					state.bg = color
				}
				i += 4
			case 5:
				if i+2 >= len(params) {
					continue
				}
				color := lipgloss.Color(strconv.Itoa(clampANSIIndex(params[i+2])))
				if code == 38 {
					state.fg = color
				} else {
					state.bg = color
				}
				i += 2
			}
		}
	}
}

func clampOpacity(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func clampColorComponent(value int) int {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return value
}
func clampANSIIndex(value int) int {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return value
}

func positiveMod(value, mod int) int {
	if mod == 0 {
		return 0
	}
	result := value % mod
	if result < 0 {
		result += mod
	}
	return result
}
