package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestDashboardCreation tests that a dashboard can be created
func TestDashboardCreation(t *testing.T) {
	dashboard := NewDashboard()

	if dashboard == nil {
		t.Fatal("Failed to create dashboard")
	}

	if dashboard.columns != 3 {
		t.Errorf("Expected 3 columns, got %d", dashboard.columns)
	}

	if !dashboard.responsive {
		t.Error("Expected responsive=true by default")
	}

	if dashboard.gap != 2 {
		t.Errorf("Expected gap=2, got %.0f", dashboard.gap)
	}
}

// TestDashboardWithTitle tests dashboard with title
func TestDashboardWithTitle(t *testing.T) {
	dashboard := NewDashboard(
		WithDashboardTitle("Test Dashboard"),
	)

	if dashboard.title != "Test Dashboard" {
		t.Errorf("Expected title='Test Dashboard', got '%s'", dashboard.title)
	}
}

// TestDashboardWithGridColumns tests fixed column layout
func TestDashboardWithGridColumns(t *testing.T) {
	dashboard := NewDashboard(
		WithGridColumns(4),
	)

	if dashboard.columns != 4 {
		t.Errorf("Expected 4 columns, got %d", dashboard.columns)
	}

	if dashboard.responsive {
		t.Error("Expected responsive=false with fixed columns")
	}
}

// TestDashboardWithGap tests gap configuration
func TestDashboardWithGap(t *testing.T) {
	dashboard := NewDashboard(
		WithGap(3),
	)

	if dashboard.gap != 3 {
		t.Errorf("Expected gap=3, got %.0f", dashboard.gap)
	}
}

// TestDashboardWithResponsiveLayout tests responsive layout
func TestDashboardWithResponsiveLayout(t *testing.T) {
	dashboard := NewDashboard(
		WithResponsiveLayout(40),
	)

	if !dashboard.responsive {
		t.Error("Expected responsive=true")
	}

	if dashboard.minCardWidth != 40 {
		t.Errorf("Expected minCardWidth=40, got %.0f", dashboard.minCardWidth)
	}
}

// TestDashboardWithCards tests adding cards
func TestDashboardWithCards(t *testing.T) {
	card1 := NewStatCard(WithTitle("Card 1"))
	card2 := NewStatCard(WithTitle("Card 2"))

	dashboard := NewDashboard(
		WithCards(card1, card2),
	)

	if len(dashboard.cards) != 2 {
		t.Errorf("Expected 2 cards, got %d", len(dashboard.cards))
	}

	if dashboard.cards[0] != card1 {
		t.Error("First card not set correctly")
	}

	if dashboard.cards[1] != card2 {
		t.Error("Second card not set correctly")
	}
}

// TestDashboardAddCard tests dynamically adding cards
func TestDashboardAddCard(t *testing.T) {
	dashboard := NewDashboard()

	card := NewStatCard(WithTitle("New Card"))
	dashboard.AddCard(card)

	if len(dashboard.cards) != 1 {
		t.Errorf("Expected 1 card, got %d", len(dashboard.cards))
	}

	if dashboard.cards[0] != card {
		t.Error("Card not added correctly")
	}
}

// TestDashboardRemoveCard tests removing cards
func TestDashboardRemoveCard(t *testing.T) {
	card1 := NewStatCard(WithTitle("Card 1"))
	card2 := NewStatCard(WithTitle("Card 2"))
	card3 := NewStatCard(WithTitle("Card 3"))

	dashboard := NewDashboard(
		WithCards(card1, card2, card3),
	)

	dashboard.RemoveCard(1) // Remove card2

	if len(dashboard.cards) != 2 {
		t.Errorf("Expected 2 cards, got %d", len(dashboard.cards))
	}

	if dashboard.cards[0] != card1 {
		t.Error("First card should still be card1")
	}

	if dashboard.cards[1] != card3 {
		t.Error("Second card should be card3")
	}
}

// TestDashboardRemoveCardInvalidIndex tests removing with invalid index
func TestDashboardRemoveCardInvalidIndex(t *testing.T) {
	card := NewStatCard(WithTitle("Card"))
	dashboard := NewDashboard(WithCards(card))

	// Try to remove with invalid indices
	dashboard.RemoveCard(-1)
	dashboard.RemoveCard(5)

	if len(dashboard.cards) != 1 {
		t.Error("Card should not be removed with invalid index")
	}
}

// TestDashboardSetCards tests replacing all cards
func TestDashboardSetCards(t *testing.T) {
	card1 := NewStatCard(WithTitle("Card 1"))
	dashboard := NewDashboard(WithCards(card1))

	card2 := NewStatCard(WithTitle("Card 2"))
	card3 := NewStatCard(WithTitle("Card 3"))
	newCards := []*StatCard{card2, card3}

	dashboard.SetCards(newCards)

	if len(dashboard.cards) != 2 {
		t.Errorf("Expected 2 cards, got %d", len(dashboard.cards))
	}

	if dashboard.cards[0] != card2 || dashboard.cards[1] != card3 {
		t.Error("Cards not replaced correctly")
	}
}

// TestDashboardGetCards tests getting all cards
func TestDashboardGetCards(t *testing.T) {
	card1 := NewStatCard(WithTitle("Card 1"))
	card2 := NewStatCard(WithTitle("Card 2"))

	dashboard := NewDashboard(WithCards(card1, card2))

	cards := dashboard.GetCards()

	if len(cards) != 2 {
		t.Errorf("Expected 2 cards, got %d", len(cards))
	}
}

// TestDashboardInit tests initialization
func TestDashboardInit(t *testing.T) {
	dashboard := NewDashboard()
	cmd := dashboard.Init()

	if cmd != nil {
		t.Error("Init should return nil command")
	}
}

// TestDashboardFocusManagement tests focus management
func TestDashboardFocusManagement(t *testing.T) {
	dashboard := NewDashboard()

	if dashboard.Focused() {
		t.Error("Dashboard should not be focused initially")
	}

	dashboard.Focus()
	if !dashboard.Focused() {
		t.Error("Dashboard should be focused after Focus()")
	}

	dashboard.Blur()
	if dashboard.Focused() {
		t.Error("Dashboard should not be focused after Blur()")
	}
}

// TestDashboardWindowSizeUpdate tests window size handling
func TestDashboardWindowSizeUpdate(t *testing.T) {
	card1 := NewStatCard(WithTitle("Card 1"))
	card2 := NewStatCard(WithTitle("Card 2"))
	dashboard := NewDashboard(WithCards(card1, card2))

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	dashboard.Update(msg)

	if dashboard.width != 100 {
		t.Errorf("Expected width=100, got %d", dashboard.width)
	}

	if dashboard.height != 50 {
		t.Errorf("Expected height=50, got %d", dashboard.height)
	}

	// Cards should have updated dimensions
	if card1.width == 0 || card1.height == 0 {
		t.Error("Card dimensions should be updated")
	}
}

// TestDashboardUpdateCardDimensions tests card dimension calculation
func TestDashboardUpdateCardDimensions(t *testing.T) {
	card1 := NewStatCard(WithTitle("Card 1"))
	card2 := NewStatCard(WithTitle("Card 2"))
	card3 := NewStatCard(WithTitle("Card 3"))

	dashboard := NewDashboard(
		WithGridColumns(3),
		WithGap(2),
		WithCards(card1, card2, card3),
	)

	// Set viewport size
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	dashboard.Update(msg)

	// All cards should have same width (3 equal columns)
	if card1.width != card2.width || card2.width != card3.width {
		t.Errorf("Cards should have equal width in 3-column grid (got %d, %d, %d)",
			card1.width, card2.width, card3.width)
	}

	// Width should account for gap
	// (100 - 2*2) / 3 = ~32 per card
	expectedWidth := (100 - 2*2) / 3
	if card1.width < expectedWidth-1 || card1.width > expectedWidth+1 {
		t.Errorf("Expected card width ~%d, got %d", expectedWidth, card1.width)
	}
}

// TestDashboardResponsiveColumns tests responsive column calculation
func TestDashboardResponsiveColumns(t *testing.T) {
	cards := []*StatCard{
		NewStatCard(WithTitle("Card 1")),
		NewStatCard(WithTitle("Card 2")),
		NewStatCard(WithTitle("Card 3")),
		NewStatCard(WithTitle("Card 4")),
		NewStatCard(WithTitle("Card 5")),
		NewStatCard(WithTitle("Card 6")),
	}

	dashboard := NewDashboard(
		WithResponsiveLayout(30), // Min 30 chars per card
		WithCards(cards...),
	)

	// Wide viewport should have multiple columns
	wideMsg := tea.WindowSizeMsg{Width: 150, Height: 50}
	dashboard.Update(wideMsg)

	// Should fit ~5 columns (150 / 30 = 5)
	// All cards should have positive dimensions
	for i, card := range cards {
		if card.width <= 0 || card.height <= 0 {
			t.Errorf("Card %d has invalid dimensions: %dx%d", i, card.width, card.height)
		}
	}

	// Narrow viewport should have fewer columns
	narrowMsg := tea.WindowSizeMsg{Width: 60, Height: 50}
	dashboard.Update(narrowMsg)

	// Should fit ~2 columns (60 / 30 = 2)
	// Cards should still have positive dimensions
	for i, card := range cards {
		if card.width <= 0 || card.height <= 0 {
			t.Errorf("Card %d has invalid dimensions after resize: %dx%d", i, card.width, card.height)
		}
	}
}

// TestDashboardView tests rendering
func TestDashboardView(t *testing.T) {
	card1 := NewStatCard(WithTitle("Card 1"), WithValue("100"))
	card2 := NewStatCard(WithTitle("Card 2"), WithValue("200"))

	dashboard := NewDashboard(
		WithDashboardTitle("Test Dashboard"),
		WithCards(card1, card2),
	)

	// Set size
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	dashboard.Update(msg)

	view := dashboard.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	// Should contain title
	if !strings.Contains(view, "Test Dashboard") {
		t.Error("View should contain dashboard title")
	}

	// Should contain card content
	if !strings.Contains(view, "Card 1") || !strings.Contains(view, "Card 2") {
		t.Error("View should contain card titles")
	}
}

// TestDashboardViewWithoutSize tests view before size is set
func TestDashboardViewWithoutSize(t *testing.T) {
	card := NewStatCard(WithTitle("Card"))
	dashboard := NewDashboard(WithCards(card))

	view := dashboard.View()

	if view != "" {
		t.Error("View should be empty without size")
	}
}

// TestDashboardViewEmpty tests view with no cards
func TestDashboardViewEmpty(t *testing.T) {
	dashboard := NewDashboard()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	dashboard.Update(msg)

	view := dashboard.View()

	if view != "" {
		t.Error("View should be empty without cards")
	}
}

// TestDashboardManyCards tests rendering many cards
func TestDashboardManyCards(t *testing.T) {
	var cards []*StatCard
	for i := 0; i < 12; i++ {
		card := NewStatCard(
			WithTitle("Card"),
			WithValue("100"),
		)
		cards = append(cards, card)
	}

	dashboard := NewDashboard(
		WithGridColumns(3),
		WithCards(cards...),
	)

	msg := tea.WindowSizeMsg{Width: 120, Height: 100}
	dashboard.Update(msg)

	view := dashboard.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	// All cards should have valid dimensions
	for i, card := range cards {
		if card.width <= 0 || card.height <= 0 {
			t.Errorf("Card %d has invalid dimensions: %dx%d", i, card.width, card.height)
		}
	}
}

// TestDashboardInitialCardFocus tests first card is focused on creation
func TestDashboardInitialCardFocus(t *testing.T) {
	card1 := NewStatCard(WithTitle("Card 1"))
	card2 := NewStatCard(WithTitle("Card 2"))
	card3 := NewStatCard(WithTitle("Card 3"))

	dashboard := NewDashboard(
		WithCards(card1, card2, card3),
	)

	// First card should be focused
	if dashboard.focusedCardIndex != 0 {
		t.Errorf("Expected focusedCardIndex=0, got %d", dashboard.focusedCardIndex)
	}

	if !card1.Focused() {
		t.Error("First card should be focused")
	}
}

// TestDashboardNavigationRight tests navigating right
func TestDashboardNavigationRight(t *testing.T) {
	card1 := NewStatCard(WithTitle("Card 1"))
	card2 := NewStatCard(WithTitle("Card 2"))
	card3 := NewStatCard(WithTitle("Card 3"))

	dashboard := NewDashboard(
		WithGridColumns(3),
		WithCards(card1, card2, card3),
	)
	dashboard.Focus()

	// Navigate right from card 0 to card 1
	msg := tea.KeyMsg{Type: tea.KeyRight}
	dashboard.Update(msg)

	if dashboard.focusedCardIndex != 1 {
		t.Errorf("Expected focusedCardIndex=1, got %d", dashboard.focusedCardIndex)
	}

	if !card2.Focused() {
		t.Error("Card 2 should be focused after right navigation")
	}

	if card1.Focused() {
		t.Error("Card 1 should be blurred")
	}
}

// TestDashboardNavigationLeft tests navigating left
func TestDashboardNavigationLeft(t *testing.T) {
	card1 := NewStatCard(WithTitle("Card 1"))
	card2 := NewStatCard(WithTitle("Card 2"))
	card3 := NewStatCard(WithTitle("Card 3"))

	dashboard := NewDashboard(
		WithGridColumns(3),
		WithCards(card1, card2, card3),
	)
	dashboard.Focus()

	// Navigate to card 1 first
	dashboard.setFocusedCard(1)

	// Navigate left from card 1 to card 0
	msg := tea.KeyMsg{Type: tea.KeyLeft}
	dashboard.Update(msg)

	if dashboard.focusedCardIndex != 0 {
		t.Errorf("Expected focusedCardIndex=0, got %d", dashboard.focusedCardIndex)
	}

	if !card1.Focused() {
		t.Error("Card 1 should be focused after left navigation")
	}
}

// TestDashboardNavigationDown tests navigating down
func TestDashboardNavigationDown(t *testing.T) {
	// Create 6 cards in 3 columns (2 rows)
	cards := []*StatCard{
		NewStatCard(WithTitle("Card 1")),
		NewStatCard(WithTitle("Card 2")),
		NewStatCard(WithTitle("Card 3")),
		NewStatCard(WithTitle("Card 4")),
		NewStatCard(WithTitle("Card 5")),
		NewStatCard(WithTitle("Card 6")),
	}

	dashboard := NewDashboard(
		WithGridColumns(3),
		WithCards(cards...),
	)
	dashboard.Focus()

	// Navigate down from card 0 (row 0, col 0) to card 3 (row 1, col 0)
	msg := tea.KeyMsg{Type: tea.KeyDown}
	dashboard.Update(msg)

	if dashboard.focusedCardIndex != 3 {
		t.Errorf("Expected focusedCardIndex=3, got %d", dashboard.focusedCardIndex)
	}

	if !cards[3].Focused() {
		t.Error("Card 4 should be focused after down navigation")
	}
}

// TestDashboardNavigationUp tests navigating up
func TestDashboardNavigationUp(t *testing.T) {
	// Create 6 cards in 3 columns (2 rows)
	cards := []*StatCard{
		NewStatCard(WithTitle("Card 1")),
		NewStatCard(WithTitle("Card 2")),
		NewStatCard(WithTitle("Card 3")),
		NewStatCard(WithTitle("Card 4")),
		NewStatCard(WithTitle("Card 5")),
		NewStatCard(WithTitle("Card 6")),
	}

	dashboard := NewDashboard(
		WithGridColumns(3),
		WithCards(cards...),
	)
	dashboard.Focus()

	// Start at card 3 (row 1, col 0)
	dashboard.setFocusedCard(3)

	// Navigate up to card 0 (row 0, col 0)
	msg := tea.KeyMsg{Type: tea.KeyUp}
	dashboard.Update(msg)

	if dashboard.focusedCardIndex != 0 {
		t.Errorf("Expected focusedCardIndex=0, got %d", dashboard.focusedCardIndex)
	}

	if !cards[0].Focused() {
		t.Error("Card 1 should be focused after up navigation")
	}
}

// TestDashboardNavigationVimKeys tests vim-style navigation (hjkl)
func TestDashboardNavigationVimKeys(t *testing.T) {
	cards := []*StatCard{
		NewStatCard(WithTitle("Card 1")),
		NewStatCard(WithTitle("Card 2")),
		NewStatCard(WithTitle("Card 3")),
	}

	dashboard := NewDashboard(
		WithGridColumns(3),
		WithCards(cards...),
	)
	dashboard.Focus()

	// Test 'l' (right)
	dashboard.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if dashboard.focusedCardIndex != 1 {
		t.Error("'l' should navigate right")
	}

	// Test 'h' (left)
	dashboard.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if dashboard.focusedCardIndex != 0 {
		t.Error("'h' should navigate left")
	}
}

// TestDashboardNavigationBoundaries tests navigation at boundaries
func TestDashboardNavigationBoundaries(t *testing.T) {
	cards := []*StatCard{
		NewStatCard(WithTitle("Card 1")),
		NewStatCard(WithTitle("Card 2")),
		NewStatCard(WithTitle("Card 3")),
	}

	dashboard := NewDashboard(
		WithGridColumns(3),
		WithCards(cards...),
	)
	dashboard.Focus()

	// Try to navigate left from first card - should stay at 0
	dashboard.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if dashboard.focusedCardIndex != 0 {
		t.Error("Should stay at first card when navigating left from boundary")
	}

	// Navigate to last card
	dashboard.setFocusedCard(2)

	// Try to navigate right from last card - should stay at 2
	dashboard.Update(tea.KeyMsg{Type: tea.KeyRight})
	if dashboard.focusedCardIndex != 2 {
		t.Error("Should stay at last card when navigating right from boundary")
	}

	// Try to navigate up from first row - should stay at 2
	dashboard.Update(tea.KeyMsg{Type: tea.KeyUp})
	if dashboard.focusedCardIndex != 2 {
		t.Error("Should stay in same position when navigating up from first row")
	}

	// Try to navigate down from last row - should stay at 2
	dashboard.Update(tea.KeyMsg{Type: tea.KeyDown})
	if dashboard.focusedCardIndex != 2 {
		t.Error("Should stay in same position when navigating down from last row")
	}
}

// TestDashboardNavigationWithoutFocus tests navigation requires focus
func TestDashboardNavigationWithoutFocus(t *testing.T) {
	card1 := NewStatCard(WithTitle("Card 1"))
	card2 := NewStatCard(WithTitle("Card 2"))

	dashboard := NewDashboard(
		WithCards(card1, card2),
	)

	// Don't call dashboard.Focus()
	dashboard.Blur()

	initialIndex := dashboard.focusedCardIndex

	// Try to navigate - should not work
	dashboard.Update(tea.KeyMsg{Type: tea.KeyRight})

	if dashboard.focusedCardIndex != initialIndex {
		t.Error("Navigation should not work without focus")
	}
}

// TestDashboardGetColumnCount tests column count calculation
func TestDashboardGetColumnCount(t *testing.T) {
	cards := []*StatCard{
		NewStatCard(WithTitle("Card 1")),
		NewStatCard(WithTitle("Card 2")),
		NewStatCard(WithTitle("Card 3")),
		NewStatCard(WithTitle("Card 4")),
	}

	// Fixed columns
	dashboard := NewDashboard(
		WithGridColumns(2),
		WithCards(cards...),
	)

	cols := dashboard.getColumnCount()
	if cols != 2 {
		t.Errorf("Expected 2 columns, got %d", cols)
	}

	// Responsive layout
	dashboardResp := NewDashboard(
		WithResponsiveLayout(30),
		WithGap(2),
		WithCards(cards...),
	)
	dashboardResp.width = 100

	cols = dashboardResp.getColumnCount()
	// (100 / (30 + 2)) = 3 columns
	if cols != 3 {
		t.Errorf("Expected 3 columns for width=100, got %d", cols)
	}
}

// TestDashboardClearSelection tests clearing card selection
func TestDashboardClearSelection(t *testing.T) {
	card1 := NewStatCard(WithTitle("Card 1"))
	card2 := NewStatCard(WithTitle("Card 2"))

	dashboard := NewDashboard(
		WithCards(card1, card2),
	)

	// Select a card manually
	dashboard.selectedCardIndex = 0
	card1.Select()

	dashboard.clearSelection()

	if dashboard.selectedCardIndex != -1 {
		t.Errorf("Expected selectedCardIndex=-1, got %d", dashboard.selectedCardIndex)
	}

	if card1.IsSelected() {
		t.Error("Card should not be selected after clearSelection()")
	}
}

// TestDashboardEscapeClearsSelection tests ESC key clears selection
func TestDashboardEscapeClearsSelection(t *testing.T) {
	card1 := NewStatCard(WithTitle("Card 1"))
	card2 := NewStatCard(WithTitle("Card 2"))

	dashboard := NewDashboard(
		WithCards(card1, card2),
	)
	dashboard.Focus()

	// Select a card
	dashboard.selectedCardIndex = 0
	card1.Select()

	// Press ESC
	dashboard.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if dashboard.selectedCardIndex != -1 {
		t.Error("ESC should clear selection")
	}

	if card1.IsSelected() {
		t.Error("Card should not be selected after ESC")
	}
}
