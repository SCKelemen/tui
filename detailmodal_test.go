package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestDetailModalCreation tests that a detail modal can be created
func TestDetailModalCreation(t *testing.T) {
	modal := NewDetailModal()

	if modal == nil {
		t.Fatal("Failed to create detail modal")
	}

	if modal.visible {
		t.Error("Modal should not be visible initially")
	}

	if modal.focused {
		t.Error("Modal should not be focused initially")
	}
}

// TestDetailModalShowHide tests show and hide functionality
func TestDetailModalShowHide(t *testing.T) {
	modal := NewDetailModal()

	// Initially hidden
	if modal.IsVisible() {
		t.Error("Modal should be hidden initially")
	}

	// Show modal
	modal.Show()
	if !modal.IsVisible() {
		t.Error("Modal should be visible after Show()")
	}
	if !modal.focused {
		t.Error("Modal should be focused after Show()")
	}

	// Hide modal
	modal.Hide()
	if modal.IsVisible() {
		t.Error("Modal should be hidden after Hide()")
	}
	if modal.focused {
		t.Error("Modal should not be focused after Hide()")
	}
}

// TestDetailModalSetContent tests setting content from StatCard
func TestDetailModalSetContent(t *testing.T) {
	card := NewStatCard(
		WithTitle("CPU Usage"),
		WithValue("42%"),
		WithSubtitle("8 cores active"),
		WithChange(5, 13.5),
		WithTrend([]float64{10, 20, 30, 40, 50}),
		WithColor("#2196F3"),
		WithTrendColor("#4CAF50"),
	)

	modal := NewDetailModal()
	modal.SetContent(card)

	if modal.title != "CPU Usage" {
		t.Errorf("Expected title='CPU Usage', got '%s'", modal.title)
	}

	if modal.value != "42%" {
		t.Errorf("Expected value='42%%', got '%s'", modal.value)
	}

	if modal.subtitle != "8 cores active" {
		t.Errorf("Expected subtitle='8 cores active', got '%s'", modal.subtitle)
	}

	if modal.change != 5 {
		t.Errorf("Expected change=5, got %d", modal.change)
	}

	if modal.changePct != 13.5 {
		t.Errorf("Expected changePct=13.5, got %.1f", modal.changePct)
	}

	if len(modal.trend) != 5 {
		t.Errorf("Expected 5 trend points, got %d", len(modal.trend))
	}

	if modal.color != "#2196F3" {
		t.Errorf("Expected color='#2196F3', got '%s'", modal.color)
	}

	if modal.trendColor != "#4CAF50" {
		t.Errorf("Expected trendColor='#4CAF50', got '%s'", modal.trendColor)
	}
}

// TestDetailModalWindowSizeUpdate tests window size handling
func TestDetailModalWindowSizeUpdate(t *testing.T) {
	modal := NewDetailModal()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	modal.Update(msg)

	if modal.width != 100 {
		t.Errorf("Expected width=100, got %d", modal.width)
	}

	if modal.height != 50 {
		t.Errorf("Expected height=50, got %d", modal.height)
	}
}

// TestDetailModalKeyHandling tests keyboard input handling
func TestDetailModalKeyHandling(t *testing.T) {
	modal := NewDetailModal()
	modal.Show()

	// Press ESC to close
	modal.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if modal.IsVisible() {
		t.Error("Modal should be hidden after ESC key")
	}

	// Show again and press 'q'
	modal.Show()
	modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if modal.IsVisible() {
		t.Error("Modal should be hidden after 'q' key")
	}
}

// TestDetailModalKeyHandlingWhenHidden tests keys are ignored when hidden
func TestDetailModalKeyHandlingWhenHidden(t *testing.T) {
	modal := NewDetailModal()

	// Modal is hidden, keys should be ignored
	modal.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if modal.IsVisible() {
		t.Error("Hidden modal should remain hidden")
	}
}

// TestDetailModalKeyHandlingWithoutFocus tests keys are ignored without focus
func TestDetailModalKeyHandlingWithoutFocus(t *testing.T) {
	modal := NewDetailModal()
	modal.visible = true
	modal.focused = false

	// Modal is visible but not focused, keys should be ignored
	modal.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !modal.visible {
		t.Error("Modal should remain visible when not focused")
	}
}

// TestDetailModalView tests basic rendering
func TestDetailModalView(t *testing.T) {
	card := NewStatCard(
		WithTitle("Memory"),
		WithValue("8 GB"),
		WithSubtitle("of 16 GB total"),
	)

	modal := NewDetailModal()
	modal.SetContent(card)
	modal.Show()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	modal.Update(msg)

	view := modal.View()

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

	// Should contain subtitle
	if !strings.Contains(view, "of 16 GB total") {
		t.Error("View should contain subtitle")
	}

	// Should have modal borders
	if !strings.Contains(view, "╔") || !strings.Contains(view, "╗") {
		t.Error("View should have double-line top border")
	}

	if !strings.Contains(view, "╚") || !strings.Contains(view, "╝") {
		t.Error("View should have double-line bottom border")
	}

	// Should show close hint
	if !strings.Contains(view, "ESC to close") {
		t.Error("View should show close hint")
	}
}

// TestDetailModalViewWithChange tests change indicator rendering
func TestDetailModalViewWithChange(t *testing.T) {
	card := NewStatCard(
		WithTitle("Users"),
		WithValue("1,000"),
		WithChange(100, 11.1),
	)

	modal := NewDetailModal()
	modal.SetContent(card)
	modal.Show()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	modal.Update(msg)

	view := modal.View()

	// Should show upward arrow for positive change
	if !strings.Contains(view, "↑") {
		t.Error("View should contain upward arrow for positive change")
	}
}

// TestDetailModalViewWithTrend tests trend graph rendering
func TestDetailModalViewWithTrend(t *testing.T) {
	trend := []float64{10, 20, 15, 25, 30, 28, 35, 40}
	card := NewStatCard(
		WithTitle("Requests"),
		WithValue("1,234"),
		WithTrend(trend),
	)

	modal := NewDetailModal()
	modal.SetContent(card)
	modal.Show()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	modal.Update(msg)

	view := modal.View()

	// Should contain trend header
	if !strings.Contains(view, "Trend") {
		t.Error("View should contain 'Trend' header")
	}

	// Should contain statistics
	if !strings.Contains(view, "Min:") || !strings.Contains(view, "Max:") || !strings.Contains(view, "Avg:") {
		t.Error("View should contain statistics (Min, Max, Avg)")
	}

	// Should contain block characters for trend graph
	hasBlock := false
	blockChars := []string{"▀", "▄", "█"}
	for _, char := range blockChars {
		if strings.Contains(view, char) {
			hasBlock = true
			break
		}
	}

	if !hasBlock {
		t.Error("View should contain block characters for trend graph")
	}
}

// TestDetailModalViewHidden tests view when hidden
func TestDetailModalViewHidden(t *testing.T) {
	modal := NewDetailModal()

	view := modal.View()

	if view != "" {
		t.Error("View should be empty when modal is hidden")
	}
}

// TestDetailModalViewWithoutSize tests view before size is set
func TestDetailModalViewWithoutSize(t *testing.T) {
	modal := NewDetailModal()
	modal.Show()

	// Width is 0
	view := modal.View()

	if view != "" {
		t.Error("View should be empty without size")
	}
}

// TestDetailModalCalculateStats tests statistics calculation
func TestDetailModalCalculateStats(t *testing.T) {
	modal := NewDetailModal()
	modal.trend = []float64{10, 20, 15, 25, 30}

	min, max, avg := modal.calculateStats()

	if min != 10 {
		t.Errorf("Expected min=10, got %.0f", min)
	}

	if max != 30 {
		t.Errorf("Expected max=30, got %.0f", max)
	}

	expectedAvg := (10 + 20 + 15 + 25 + 30) / 5.0
	if avg != expectedAvg {
		t.Errorf("Expected avg=%.1f, got %.1f", expectedAvg, avg)
	}
}

// TestDetailModalCalculateStatsEmpty tests stats with no data
func TestDetailModalCalculateStatsEmpty(t *testing.T) {
	modal := NewDetailModal()
	modal.trend = []float64{}

	min, max, avg := modal.calculateStats()

	if min != 0 || max != 0 || avg != 0 {
		t.Error("Empty trend should return all zeros")
	}
}

// TestDetailModalRenderLargeTrendGraph tests large trend graph rendering
func TestDetailModalRenderLargeTrendGraph(t *testing.T) {
	modal := NewDetailModal()
	modal.trend = []float64{0, 25, 50, 75, 100}

	lines := modal.renderLargeTrendGraph(50)

	if len(lines) != 8 {
		t.Errorf("Expected 8 lines, got %d", len(lines))
	}

	// Each line should have content
	for i, line := range lines {
		if len(line) == 0 {
			t.Errorf("Line %d should not be empty", i)
		}
	}
}

// TestDetailModalRenderLargeTrendGraphEmpty tests empty trend
func TestDetailModalRenderLargeTrendGraphEmpty(t *testing.T) {
	modal := NewDetailModal()
	modal.trend = []float64{}

	lines := modal.renderLargeTrendGraph(50)

	if len(lines) != 0 {
		t.Error("Empty trend should return no lines")
	}
}

// TestDetailModalVisibleLength tests ANSI-aware length calculation
func TestDetailModalVisibleLength(t *testing.T) {
	modal := NewDetailModal()

	// Plain string
	plain := "Hello"
	length := modal.visibleLength(plain)
	if length != 5 {
		t.Errorf("Expected length 5, got %d", length)
	}

	// String with ANSI codes
	ansi := "\033[32mGreen\033[0m"
	length = modal.visibleLength(ansi)
	if length != 5 {
		t.Errorf("Expected visible length 5, got %d", length)
	}
}

// TestDetailModalWithHistory tests historical data rendering
func TestDetailModalWithHistory(t *testing.T) {
	modal := NewDetailModal(
		WithHistory([]string{
			"2024-01-10: 1,234 users",
			"2024-01-09: 1,189 users",
			"2024-01-08: 1,156 users",
		}),
	)

	card := NewStatCard(WithTitle("Users"), WithValue("1,234"))
	modal.SetContent(card)
	modal.Show()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	modal.Update(msg)

	view := modal.View()

	// Should contain history header
	if !strings.Contains(view, "Recent History") {
		t.Error("View should contain 'Recent History' header")
	}

	// Should contain history entries
	if !strings.Contains(view, "2024-01-10") {
		t.Error("View should contain history entries")
	}
}

// TestDetailModalDimensions tests modal sizing
func TestDetailModalDimensions(t *testing.T) {
	modal := NewDetailModal()
	modal.Show()

	// Set viewport size
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	modal.Update(msg)

	view := modal.View()

	// Modal should be centered and sized appropriately
	// 70% of 100 = 70 width
	// 80% of 50 = 40 height

	if view == "" {
		t.Error("View should not be empty")
	}

	// Check that modal has reasonable size (not too small)
	lines := strings.Split(view, "\n")
	if len(lines) < 20 {
		t.Error("Modal should have reasonable height")
	}
}

// TestDetailModalNarrowViewport tests modal with narrow viewport
func TestDetailModalNarrowViewport(t *testing.T) {
	modal := NewDetailModal()
	modal.Show()

	card := NewStatCard(WithTitle("Test"))
	modal.SetContent(card)

	// Very narrow viewport
	msg := tea.WindowSizeMsg{Width: 50, Height: 20}
	modal.Update(msg)

	view := modal.View()

	// Modal should still render but use minimum size
	if view == "" {
		t.Error("View should not be empty even with narrow viewport")
	}
}

// TestDetailModalIntegrationWithDashboard tests modal integration
func TestDetailModalIntegrationWithDashboard(t *testing.T) {
	card := NewStatCard(
		WithTitle("CPU Usage"),
		WithValue("42%"),
	)

	dashboard := NewDashboard(
		WithCards(card),
	)
	dashboard.Focus()

	// Set viewport size
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	dashboard.Update(msg)

	// Modal should not be visible initially
	view := dashboard.View()
	if strings.Contains(view, "ESC to close") {
		t.Error("Modal should not be visible initially")
	}

	// Open modal by pressing Enter
	dashboard.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Modal should now be visible
	if !dashboard.detailModal.IsVisible() {
		t.Error("Modal should be visible after pressing Enter")
	}

	// Dashboard should lose focus
	if dashboard.focused {
		t.Error("Dashboard should lose focus when modal opens")
	}

	// View should show modal
	view = dashboard.View()
	if !strings.Contains(view, "ESC to close") {
		t.Error("View should show modal overlay")
	}

	// Close modal by pressing ESC
	dashboard.detailModal.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Modal should be hidden
	if dashboard.detailModal.IsVisible() {
		t.Error("Modal should be hidden after pressing ESC")
	}
}
