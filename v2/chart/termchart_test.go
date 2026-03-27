package chart

import (
	"strings"
	"testing"
)

func TestTermBarChartView(t *testing.T) {
	c := NewBarChart("Sales", []string{"A", "B", "C"}, []float64{3, 7, 5}, WithBarChartWidth(30), WithBarChartHeight(8))
	out := c.View()

	if !strings.Contains(out, "Sales") {
		t.Fatalf("expected title in output, got: %q", out)
	}
	if !strings.Contains(out, "A") {
		t.Fatalf("expected bar label in output, got: %q", out)
	}
	if !strings.Contains(out, "█") {
		t.Fatalf("expected rendered bar glyphs, got: %q", out)
	}
}

func TestTermLineChartView(t *testing.T) {
	c := NewLineChart("Trend", WithLineChartWidth(20), WithLineChartHeight(6))
	c.AddSeries("s1", []float64{1, 3, 2, 5, 4, 6})
	out := c.View()

	if !strings.Contains(out, "Trend") {
		t.Fatalf("expected title in output, got: %q", out)
	}
	if len(strings.Split(out, "\n")) < 2 {
		t.Fatalf("expected multi-line output, got: %q", out)
	}
}

func TestTermHeatMapView(t *testing.T) {
	c := NewHeatMap("Load", [][]float64{
		{0, 1, 2, 3},
		{3, 2, 1, 0},
	}, WithHeatMapWidth(4), WithHeatMapHeight(2))
	out := c.View()

	if !strings.Contains(out, "Load") {
		t.Fatalf("expected title in output, got: %q", out)
	}
	if !strings.ContainsAny(out, "░▒▓█") {
		t.Fatalf("expected heatmap block glyphs, got: %q", out)
	}
}

func TestTermScatterPlotView(t *testing.T) {
	c := NewScatterPlot(
		"Scatter",
		[]float64{0, 1, 2, 3, 4, 5},
		[]float64{5, 1, 4, 2, 3, 0},
		WithScatterPlotWidth(20),
		WithScatterPlotHeight(6),
	)
	out := c.View()

	if !strings.Contains(out, "Scatter") {
		t.Fatalf("expected title in output, got: %q", out)
	}
	if len(strings.Split(out, "\n")) < 2 {
		t.Fatalf("expected multi-line output, got: %q", out)
	}
}
