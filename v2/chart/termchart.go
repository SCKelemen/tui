package chart

import (
	"math"
	"strings"

	dataviz "github.com/SCKelemen/dataviz/charts"
	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// BarChart wraps dataviz terminal bar chart rendering.
type BarChart struct {
	title  string
	labels []string
	values []float64

	width  int
	height int

	focused      bool
	designTokens *design.DesignTokens
}

// BarChartOption configures a BarChart.
type BarChartOption func(*BarChart)

// WithBarChartWidth sets chart width.
func WithBarChartWidth(width int) BarChartOption {
	return func(c *BarChart) {
		if width > 0 {
			c.width = width
		}
	}
}

// WithBarChartHeight sets chart height.
func WithBarChartHeight(height int) BarChartOption {
	return func(c *BarChart) {
		if height > 0 {
			c.height = height
		}
	}
}

// WithBarChartDesignTokens applies design-system tokens.
func WithBarChartDesignTokens(tokens *design.DesignTokens) BarChartOption {
	return func(c *BarChart) {
		if tokens != nil {
			c.designTokens = tokens
		}
	}
}

// WithBarChartTheme applies a named theme.
func WithBarChartTheme(theme string) BarChartOption {
	return func(c *BarChart) {
		c.designTokens = style.DesignTokensForTheme(theme)
	}
}

// NewBarChart creates a new BarChart.
func NewBarChart(title string, labels []string, values []float64, opts ...BarChartOption) *BarChart {
	c := &BarChart{
		title:        title,
		labels:       append([]string(nil), labels...),
		values:       append([]float64(nil), values...),
		width:        40,
		height:       10,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// SetData replaces labels and values.
func (c *BarChart) SetData(labels []string, values []float64) {
	c.labels = append([]string(nil), labels...)
	c.values = append([]float64(nil), values...)
}

// Init initializes the component.
func (c *BarChart) Init() tea.Cmd { return nil }

// Update handles messages.
func (c *BarChart) Update(msg tea.Msg) (tui.Component, tea.Cmd) { return c, nil }

// View renders the bar chart.
func (c *BarChart) View() string {
	bars := make([]dataviz.BarData, 0, minInt(len(c.labels), len(c.values)))
	for i := 0; i < len(c.labels) && i < len(c.values); i++ {
		v := int(math.Round(c.values[i]))
		if v < 0 {
			v = 0
		}
		bars = append(bars, dataviz.BarData{Label: c.labels[i], Value: v})
	}

	r := dataviz.NewTerminalRenderer()
	output := r.RenderBarChart(
		dataviz.BarChartData{Bars: bars},
		dataviz.Bounds{Width: c.width, Height: c.height},
		dataviz.RenderConfig{DesignTokens: c.designTokens},
	)

	return withTitle(c.title, output.String())
}

// Focus marks the component focused.
func (c *BarChart) Focus() { c.focused = true }

// Blur marks the component unfocused.
func (c *BarChart) Blur() { c.focused = false }

// Focused reports focus state.
func (c *BarChart) Focused() bool { return c.focused }

// LineChart wraps terminal line chart rendering using a braille canvas.
type LineChart struct {
	title        string
	series       [][]float64
	seriesLabels []string

	width  int
	height int

	focused      bool
	designTokens *design.DesignTokens
}

// LineChartOption configures a LineChart.
type LineChartOption func(*LineChart)

// WithLineChartWidth sets chart width.
func WithLineChartWidth(width int) LineChartOption {
	return func(c *LineChart) {
		if width > 0 {
			c.width = width
		}
	}
}

// WithLineChartHeight sets chart height.
func WithLineChartHeight(height int) LineChartOption {
	return func(c *LineChart) {
		if height > 0 {
			c.height = height
		}
	}
}

// WithLineChartDesignTokens applies design-system tokens.
func WithLineChartDesignTokens(tokens *design.DesignTokens) LineChartOption {
	return func(c *LineChart) {
		if tokens != nil {
			c.designTokens = tokens
		}
	}
}

// WithLineChartTheme applies a named theme.
func WithLineChartTheme(theme string) LineChartOption {
	return func(c *LineChart) {
		c.designTokens = style.DesignTokensForTheme(theme)
	}
}

// NewLineChart creates a new LineChart.
func NewLineChart(title string, opts ...LineChartOption) *LineChart {
	c := &LineChart{
		title:        title,
		width:        40,
		height:       10,
		designTokens: design.DefaultTheme(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// AddSeries appends a named data series.
func (c *LineChart) AddSeries(label string, data []float64) {
	c.seriesLabels = append(c.seriesLabels, label)
	c.series = append(c.series, append([]float64(nil), data...))
}

// Init initializes the component.
func (c *LineChart) Init() tea.Cmd { return nil }

// Update handles messages.
func (c *LineChart) Update(msg tea.Msg) (tui.Component, tea.Cmd) { return c, nil }

// View renders the line chart using braille cells.
func (c *LineChart) View() string {
	if len(c.series) == 0 {
		return withTitle(c.title, "")
	}

	minVal, maxVal, ok := seriesMinMax(c.series)
	if !ok {
		return withTitle(c.title, "")
	}

	canvas := dataviz.NewBrailleCanvas(c.width, c.height)
	pixelHeight := c.height*4 - 1
	pixelWidth := c.width*2 - 1
	if pixelHeight < 1 || pixelWidth < 1 {
		return withTitle(c.title, "")
	}

	for _, s := range c.series {
		if len(s) == 0 {
			continue
		}
		points := make([]dataviz.Point, 0, len(s))
		for i, v := range s {
			x := 0.0
			if len(s) > 1 {
				x = float64(i) / float64(len(s)-1) * float64(pixelWidth)
			}
			norm := normalize(v, minVal, maxVal)
			y := float64(pixelHeight) - norm*float64(pixelHeight)
			points = append(points, dataviz.Point{X: x, Y: y})
		}
		canvas.DrawCurve(points)
	}

	return withTitle(c.title, canvas.Render())
}

// Focus marks the component focused.
func (c *LineChart) Focus() { c.focused = true }

// Blur marks the component unfocused.
func (c *LineChart) Blur() { c.focused = false }

// Focused reports focus state.
func (c *LineChart) Focused() bool { return c.focused }

// HeatMap renders a matrix as colored terminal blocks.
type HeatMap struct {
	title string
	data  [][]float64

	xLabels []string
	yLabels []string

	width  int
	height int

	focused      bool
	designTokens *design.DesignTokens
}

// HeatMapOption configures a HeatMap.
type HeatMapOption func(*HeatMap)

// WithHeatMapWidth sets chart width.
func WithHeatMapWidth(width int) HeatMapOption {
	return func(c *HeatMap) {
		if width > 0 {
			c.width = width
		}
	}
}

// WithHeatMapHeight sets chart height.
func WithHeatMapHeight(height int) HeatMapOption {
	return func(c *HeatMap) {
		if height > 0 {
			c.height = height
		}
	}
}

// WithHeatMapDesignTokens applies design-system tokens.
func WithHeatMapDesignTokens(tokens *design.DesignTokens) HeatMapOption {
	return func(c *HeatMap) {
		if tokens != nil {
			c.designTokens = tokens
		}
	}
}

// WithHeatMapTheme applies a named theme.
func WithHeatMapTheme(theme string) HeatMapOption {
	return func(c *HeatMap) {
		c.designTokens = style.DesignTokensForTheme(theme)
	}
}

// NewHeatMap creates a new HeatMap.
func NewHeatMap(title string, data [][]float64, opts ...HeatMapOption) *HeatMap {
	copied := make([][]float64, len(data))
	for i := range data {
		copied[i] = append([]float64(nil), data[i]...)
	}

	c := &HeatMap{
		title:        title,
		data:         copied,
		width:        32,
		height:       8,
		designTokens: design.DefaultTheme(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Init initializes the component.
func (c *HeatMap) Init() tea.Cmd { return nil }

// Update handles messages.
func (c *HeatMap) Update(msg tea.Msg) (tui.Component, tea.Cmd) { return c, nil }

// View renders the heatmap using ANSI-colored blocks.
func (c *HeatMap) View() string {
	if len(c.data) == 0 {
		return withTitle(c.title, "")
	}

	rows := minInt(c.height, len(c.data))
	if rows <= 0 {
		rows = len(c.data)
	}

	minVal, maxVal := matrixMinMax(c.data)
	blocks := []string{" ", "░", "▒", "▓", "█"}

	colorMode := dataviz.TerminalColorTrue
	if c.designTokens != nil && c.designTokens.Mode == "basic" {
		colorMode = dataviz.TerminalColor256
	}

	accent := "#40C463"
	if c.designTokens != nil && c.designTokens.Accent != "" {
		accent = c.designTokens.Accent
	}
	gradient := dataviz.InterpolateColorGradient("#161B22", accent, len(blocks), colorMode)
	ansiReset := "\x1b[0m"

	var b strings.Builder
	for r := 0; r < rows; r++ {
		row := c.data[r]
		cols := len(row)
		if c.width > 0 {
			cols = minInt(cols, c.width)
		}
		for col := 0; col < cols; col++ {
			n := normalize(row[col], minVal, maxVal)
			idx := int(math.Round(n * float64(len(blocks)-1)))
			if idx < 0 {
				idx = 0
			}
			if idx >= len(blocks) {
				idx = len(blocks) - 1
			}
			if gradient[idx] != "" {
				b.WriteString(gradient[idx])
			}
			b.WriteString(blocks[idx])
			if gradient[idx] != "" {
				b.WriteString(ansiReset)
			}
		}
		if r < rows-1 {
			b.WriteString("\n")
		}
	}

	return withTitle(c.title, b.String())
}

// Focus marks the component focused.
func (c *HeatMap) Focus() { c.focused = true }

// Blur marks the component unfocused.
func (c *HeatMap) Blur() { c.focused = false }

// Focused reports focus state.
func (c *HeatMap) Focused() bool { return c.focused }

// ScatterPlot renders a terminal scatter plot using braille dots.
type ScatterPlot struct {
	title string
	xData []float64
	yData []float64

	width  int
	height int

	focused      bool
	designTokens *design.DesignTokens
}

// ScatterPlotOption configures a ScatterPlot.
type ScatterPlotOption func(*ScatterPlot)

// WithScatterPlotWidth sets chart width.
func WithScatterPlotWidth(width int) ScatterPlotOption {
	return func(c *ScatterPlot) {
		if width > 0 {
			c.width = width
		}
	}
}

// WithScatterPlotHeight sets chart height.
func WithScatterPlotHeight(height int) ScatterPlotOption {
	return func(c *ScatterPlot) {
		if height > 0 {
			c.height = height
		}
	}
}

// WithScatterPlotDesignTokens applies design-system tokens.
func WithScatterPlotDesignTokens(tokens *design.DesignTokens) ScatterPlotOption {
	return func(c *ScatterPlot) {
		if tokens != nil {
			c.designTokens = tokens
		}
	}
}

// WithScatterPlotTheme applies a named theme.
func WithScatterPlotTheme(theme string) ScatterPlotOption {
	return func(c *ScatterPlot) {
		c.designTokens = style.DesignTokensForTheme(theme)
	}
}

// NewScatterPlot creates a new ScatterPlot.
func NewScatterPlot(title string, xData, yData []float64, opts ...ScatterPlotOption) *ScatterPlot {
	c := &ScatterPlot{
		title:        title,
		xData:        append([]float64(nil), xData...),
		yData:        append([]float64(nil), yData...),
		width:        40,
		height:       10,
		designTokens: design.DefaultTheme(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Init initializes the component.
func (c *ScatterPlot) Init() tea.Cmd { return nil }

// Update handles messages.
func (c *ScatterPlot) Update(msg tea.Msg) (tui.Component, tea.Cmd) { return c, nil }

// View renders the scatter plot using braille dots.
func (c *ScatterPlot) View() string {
	n := minInt(len(c.xData), len(c.yData))
	if n == 0 {
		return withTitle(c.title, "")
	}

	canvas := dataviz.NewBrailleCanvas(c.width, c.height)
	pixelHeight := c.height*4 - 1
	pixelWidth := c.width*2 - 1
	if pixelHeight < 1 || pixelWidth < 1 {
		return withTitle(c.title, "")
	}

	xMin, xMax := sliceMinMax(c.xData[:n])
	yMin, yMax := sliceMinMax(c.yData[:n])

	for i := 0; i < n; i++ {
		xNorm := normalize(c.xData[i], xMin, xMax)
		yNorm := normalize(c.yData[i], yMin, yMax)

		x := int(math.Round(xNorm * float64(pixelWidth)))
		y := int(math.Round(float64(pixelHeight) - yNorm*float64(pixelHeight)))
		canvas.DrawPoint(x, y)
	}

	return withTitle(c.title, canvas.Render())
}

// Focus marks the component focused.
func (c *ScatterPlot) Focus() { c.focused = true }

// Blur marks the component unfocused.
func (c *ScatterPlot) Blur() { c.focused = false }

// Focused reports focus state.
func (c *ScatterPlot) Focused() bool { return c.focused }

func withTitle(title, body string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return body
	}
	if body == "" {
		return title
	}
	return title + "\n" + body
}

func normalize(v, minV, maxV float64) float64 {
	rng := maxV - minV
	if rng == 0 {
		return 0.5
	}
	n := (v - minV) / rng
	if n < 0 {
		return 0
	}
	if n > 1 {
		return 1
	}
	return n
}

func seriesMinMax(series [][]float64) (float64, float64, bool) {
	found := false
	var minV, maxV float64
	for _, s := range series {
		for _, v := range s {
			if !found {
				minV, maxV = v, v
				found = true
				continue
			}
			if v < minV {
				minV = v
			}
			if v > maxV {
				maxV = v
			}
		}
	}
	return minV, maxV, found
}

func sliceMinMax(values []float64) (float64, float64) {
	minV := values[0]
	maxV := values[0]
	for _, v := range values[1:] {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	return minV, maxV
}

func matrixMinMax(matrix [][]float64) (float64, float64) {
	found := false
	var minV, maxV float64
	for _, row := range matrix {
		for _, v := range row {
			if !found {
				minV, maxV = v, v
				found = true
				continue
			}
			if v < minV {
				minV = v
			}
			if v > maxV {
				maxV = v
			}
		}
	}
	if !found {
		return 0, 1
	}
	return minV, maxV
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var (
	_ tui.Component = (*BarChart)(nil)
	_ tui.Component = (*LineChart)(nil)
	_ tui.Component = (*HeatMap)(nil)
	_ tui.Component = (*ScatterPlot)(nil)
)
