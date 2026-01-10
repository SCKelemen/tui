package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestStatCardCreation tests that a stat card can be created
func TestStatCardCreation(t *testing.T) {
	card := NewStatCard()

	if card == nil {
		t.Fatal("Failed to create stat card")
	}

	if card.width != 30 {
		t.Errorf("Expected width=30, got %d", card.width)
	}

	if card.height != 8 {
		t.Errorf("Expected height=8, got %d", card.height)
	}

	if card.color == "" {
		t.Error("Color should have default value")
	}
}

// TestStatCardWithTitle tests title option
func TestStatCardWithTitle(t *testing.T) {
	card := NewStatCard(
		WithTitle("CPU Usage"),
	)

	if card.title != "CPU Usage" {
		t.Errorf("Expected title='CPU Usage', got '%s'", card.title)
	}
}

// TestStatCardWithValue tests value option
func TestStatCardWithValue(t *testing.T) {
	card := NewStatCard(
		WithValue("42%"),
	)

	if card.value != "42%" {
		t.Errorf("Expected value='42%%', got '%s'", card.value)
	}
}

// TestStatCardWithSubtitle tests subtitle option
func TestStatCardWithSubtitle(t *testing.T) {
	card := NewStatCard(
		WithSubtitle("Last hour"),
	)

	if card.subtitle != "Last hour" {
		t.Errorf("Expected subtitle='Last hour', got '%s'", card.subtitle)
	}
}

// TestStatCardWithChange tests change indicator
func TestStatCardWithChange(t *testing.T) {
	card := NewStatCard(
		WithChange(10, 5.5),
	)

	if card.change != 10 {
		t.Errorf("Expected change=10, got %d", card.change)
	}

	if card.changePct != 5.5 {
		t.Errorf("Expected changePct=5.5, got %.1f", card.changePct)
	}
}

// TestStatCardWithTrend tests trend data
func TestStatCardWithTrend(t *testing.T) {
	trend := []float64{10, 20, 15, 25, 30}
	card := NewStatCard(
		WithTrend(trend),
	)

	if len(card.trend) != 5 {
		t.Errorf("Expected 5 trend points, got %d", len(card.trend))
	}

	if card.trend[0] != 10 {
		t.Errorf("Expected first trend value=10, got %.0f", card.trend[0])
	}
}

// TestStatCardWithColor tests color option
func TestStatCardWithColor(t *testing.T) {
	card := NewStatCard(
		WithColor("#FF5722"),
	)

	if card.color != "#FF5722" {
		t.Errorf("Expected color='#FF5722', got '%s'", card.color)
	}
}

// TestStatCardWithTrendColor tests trend color option
func TestStatCardWithTrendColor(t *testing.T) {
	card := NewStatCard(
		WithTrendColor("#4CAF50"),
	)

	if card.trendColor != "#4CAF50" {
		t.Errorf("Expected trendColor='#4CAF50', got '%s'", card.trendColor)
	}
}

// TestStatCardInit tests initialization
func TestStatCardInit(t *testing.T) {
	card := NewStatCard()
	cmd := card.Init()

	if cmd != nil {
		t.Error("Init should return nil command")
	}
}

// TestStatCardFocusManagement tests focus management
func TestStatCardFocusManagement(t *testing.T) {
	card := NewStatCard()

	if card.Focused() {
		t.Error("Card should not be focused initially")
	}

	card.Focus()
	if !card.Focused() {
		t.Error("Card should be focused after Focus()")
	}

	card.Blur()
	if card.Focused() {
		t.Error("Card should not be focused after Blur()")
	}
}

// TestStatCardWindowSizeUpdate tests window size handling
func TestStatCardWindowSizeUpdate(t *testing.T) {
	card := NewStatCard()

	msg := tea.WindowSizeMsg{Width: 40, Height: 10}
	card.Update(msg)

	if card.width != 40 {
		t.Errorf("Expected width=40, got %d", card.width)
	}

	if card.height != 10 {
		t.Errorf("Expected height=10, got %d", card.height)
	}
}

// TestStatCardView tests basic rendering
func TestStatCardView(t *testing.T) {
	card := NewStatCard(
		WithTitle("Memory"),
		WithValue("8 GB"),
	)

	card.width = 30
	card.height = 8

	view := card.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	// Should contain title
	if !strings.Contains(view, "Memory") {
		t.Error("View should contain title")
	}

	// Should contain value
	if !strings.Contains(view, "8 GB") {
		t.Error("View should contain value")
	}

	// Should have borders
	if !strings.Contains(view, "┌") || !strings.Contains(view, "┐") {
		t.Error("View should have top border")
	}

	if !strings.Contains(view, "└") || !strings.Contains(view, "┘") {
		t.Error("View should have bottom border")
	}
}

// TestStatCardViewWithChange tests change indicator rendering
func TestStatCardViewWithChange(t *testing.T) {
	// Positive change
	cardUp := NewStatCard(
		WithTitle("Users"),
		WithValue("1,000"),
		WithChange(100, 11.1),
	)
	cardUp.width = 30
	cardUp.height = 8

	viewUp := cardUp.View()

	// Should show upward arrow for positive change
	if !strings.Contains(viewUp, "↑") {
		t.Error("View should contain upward arrow for positive change")
	}

	// Negative change
	cardDown := NewStatCard(
		WithTitle("Errors"),
		WithValue("50"),
		WithChange(-10, -16.7),
	)
	cardDown.width = 30
	cardDown.height = 8

	viewDown := cardDown.View()

	// Should show downward arrow for negative change
	if !strings.Contains(viewDown, "↓") {
		t.Error("View should contain downward arrow for negative change")
	}

	// Zero change - note: when both change and changePct are 0,
	// the change row is not rendered, so we don't test for arrow
}

// TestStatCardViewWithSubtitle tests subtitle rendering
func TestStatCardViewWithSubtitle(t *testing.T) {
	card := NewStatCard(
		WithTitle("Latency"),
		WithValue("42ms"),
		WithSubtitle("p95: 125ms"),
	)

	card.width = 30
	card.height = 10

	view := card.View()

	if !strings.Contains(view, "p95: 125ms") {
		t.Error("View should contain subtitle")
	}
}

// TestStatCardViewWithTrend tests sparkline rendering
func TestStatCardViewWithTrend(t *testing.T) {
	trend := []float64{10, 20, 15, 25, 30, 28, 35}
	card := NewStatCard(
		WithTitle("Requests"),
		WithValue("1,234"),
		WithTrend(trend),
	)

	card.width = 30
	card.height = 10

	view := card.View()

	// Should contain sparkline characters
	hasSparkline := false
	sparklineChars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	for _, char := range sparklineChars {
		if strings.Contains(view, char) {
			hasSparkline = true
			break
		}
	}

	if !hasSparkline {
		t.Error("View should contain sparkline characters")
	}
}

// TestStatCardViewWithoutSize tests view before size is set
func TestStatCardViewWithoutSize(t *testing.T) {
	card := NewStatCard(WithTitle("Test"))

	// Set width to 0
	card.width = 0

	view := card.View()

	if view != "" {
		t.Error("View should be empty without size")
	}
}

// TestStatCardRenderChange tests change rendering
func TestStatCardRenderChange(t *testing.T) {
	card := NewStatCard()

	// Positive change
	card.change = 10
	card.changePct = 5.5
	changeStr := card.renderChange()

	if !strings.Contains(changeStr, "↑") {
		t.Error("Should contain upward arrow")
	}

	if !strings.Contains(changeStr, "10") {
		t.Error("Should contain change value")
	}

	if !strings.Contains(changeStr, "5.5") {
		t.Error("Should contain percentage")
	}

	// Negative change
	card.change = -5
	card.changePct = -2.5
	changeStr = card.renderChange()

	if !strings.Contains(changeStr, "↓") {
		t.Error("Should contain downward arrow")
	}

	if !strings.Contains(changeStr, "5") {
		t.Error("Should contain absolute change value")
	}
}

// TestStatCardRenderSparkline tests sparkline rendering
func TestStatCardRenderSparkline(t *testing.T) {
	trend := []float64{0, 25, 50, 75, 100}
	card := NewStatCard(WithTrend(trend))

	sparkline := card.renderSparkline(20)

	if len(sparkline) == 0 {
		t.Error("Sparkline should not be empty")
	}

	// Should contain block characters
	hasBlock := false
	for _, ch := range sparkline {
		if ch >= '▁' && ch <= '█' {
			hasBlock = true
			break
		}
	}

	if !hasBlock {
		t.Error("Sparkline should contain block characters")
	}
}

// TestStatCardRenderSparklineEmpty tests sparkline with no data
func TestStatCardRenderSparklineEmpty(t *testing.T) {
	card := NewStatCard()

	sparkline := card.renderSparkline(20)

	if sparkline != "" {
		t.Error("Sparkline should be empty with no trend data")
	}
}

// TestStatCardRenderSparklineConstant tests sparkline with constant values
func TestStatCardRenderSparklineConstant(t *testing.T) {
	// All same values
	trend := []float64{50, 50, 50, 50, 50}
	card := NewStatCard(WithTrend(trend))

	sparkline := card.renderSparkline(20)

	// Should use middle block character
	if !strings.Contains(sparkline, "▄") && !strings.Contains(sparkline, "▅") {
		t.Error("Constant trend should use middle block characters")
	}
}

// TestStatCardTruncate tests string truncation
func TestStatCardTruncate(t *testing.T) {
	card := NewStatCard()

	// Short string
	result := card.truncate("Short", 20)
	if len(result) != 20 {
		t.Errorf("Expected length 20, got %d", len(result))
	}

	// Long string
	long := "This is a very long string that needs truncation"
	result = card.truncate(long, 20)
	if !strings.Contains(result, "...") {
		t.Error("Long string should be truncated with ellipsis")
	}
}

// TestStatCardVisibleLength tests ANSI-aware length calculation
func TestStatCardVisibleLength(t *testing.T) {
	card := NewStatCard()

	// Plain string
	plain := "Hello"
	length := card.visibleLength(plain)
	if length != 5 {
		t.Errorf("Expected length 5, got %d", length)
	}

	// String with ANSI codes
	ansi := "\033[32mGreen\033[0m"
	length = card.visibleLength(ansi)
	if length != 5 {
		t.Errorf("Expected visible length 5, got %d", length)
	}
}

// TestStatCardAllOptions tests combining all options
func TestStatCardAllOptions(t *testing.T) {
	trend := []float64{10, 15, 20, 25, 30}
	card := NewStatCard(
		WithTitle("Full Card"),
		WithValue("9,999"),
		WithSubtitle("Last 24 hours"),
		WithChange(500, 5.3),
		WithTrend(trend),
		WithColor("#2196F3"),
		WithTrendColor("#4CAF50"),
	)

	card.width = 40
	card.height = 12

	view := card.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	// Should contain all elements
	if !strings.Contains(view, "Full Card") {
		t.Error("Should contain title")
	}

	if !strings.Contains(view, "9,999") {
		t.Error("Should contain value")
	}

	if !strings.Contains(view, "Last 24 hours") {
		t.Error("Should contain subtitle")
	}

	if !strings.Contains(view, "↑") {
		t.Error("Should contain change indicator")
	}
}

// TestStatCardVeryNarrowWidth tests rendering with narrow width
func TestStatCardVeryNarrowWidth(t *testing.T) {
	card := NewStatCard(
		WithTitle("Test"),
		WithValue("100"),
	)

	card.width = 15
	card.height = 8

	view := card.View()

	if view == "" {
		t.Error("View should not be empty even with narrow width")
	}
}

// TestStatCardVeryShortHeight tests rendering with short height
func TestStatCardVeryShortHeight(t *testing.T) {
	card := NewStatCard(
		WithTitle("Test"),
		WithValue("100"),
	)

	card.width = 30
	card.height = 5

	view := card.View()

	if view == "" {
		t.Error("View should not be empty even with short height")
	}
}

// TestStatCardEmptyValues tests rendering with empty values
func TestStatCardEmptyValues(t *testing.T) {
	card := NewStatCard()

	card.width = 30
	card.height = 8

	view := card.View()

	if view == "" {
		t.Error("View should not be empty even without values")
	}

	// Should still have borders
	if !strings.Contains(view, "┌") {
		t.Error("Should have borders")
	}
}
