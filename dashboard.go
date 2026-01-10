package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/cli/renderer"
	"github.com/SCKelemen/color"
	design "github.com/SCKelemen/design-system"
	"github.com/SCKelemen/layout"
)

// Dashboard displays multiple stat cards in a responsive grid layout
type Dashboard struct {
	width   int
	height  int
	focused bool
	tokens  *design.DesignTokens

	// Layout configuration
	columns      int     // Number of columns in grid
	gap          float64 // Gap between cards in characters
	minCardWidth float64 // Minimum card width for responsive layout
	responsive   bool    // Use responsive grid layout

	// Cards
	cards []*StatCard

	// Navigation
	focusedCardIndex int // Index of currently focused card (-1 = none)
	selectedCardIndex int // Index of selected card for drill-down (-1 = none)

	// Title
	title string
}

// DashboardOption configures a Dashboard
type DashboardOption func(*Dashboard)

// WithDashboardTitle sets the dashboard title
func WithDashboardTitle(title string) DashboardOption {
	return func(d *Dashboard) {
		d.title = title
	}
}

// WithGridColumns sets the number of columns in the grid
func WithGridColumns(columns int) DashboardOption {
	return func(d *Dashboard) {
		d.columns = columns
		d.responsive = false
	}
}

// WithGap sets the gap between cards
func WithGap(gap float64) DashboardOption {
	return func(d *Dashboard) {
		d.gap = gap
	}
}

// WithResponsiveLayout enables responsive grid layout
func WithResponsiveLayout(minCardWidth float64) DashboardOption {
	return func(d *Dashboard) {
		d.responsive = true
		d.minCardWidth = minCardWidth
	}
}

// WithCards sets the stat cards to display
func WithCards(cards ...*StatCard) DashboardOption {
	return func(d *Dashboard) {
		d.cards = cards
	}
}

// NewDashboard creates a new dashboard
func NewDashboard(opts ...DashboardOption) *Dashboard {
	d := &Dashboard{
		columns:           3,
		gap:               2,
		minCardWidth:      30,
		responsive:        true,
		tokens:            design.DefaultTheme(),
		cards:             []*StatCard{},
		focusedCardIndex:  -1, // No card focused initially
		selectedCardIndex: -1, // No card selected initially
	}

	for _, opt := range opts {
		opt(d)
	}

	// Focus first card if cards exist
	if len(d.cards) > 0 {
		d.focusedCardIndex = 0
		d.cards[0].Focus()
	}

	return d
}

// Init initializes the dashboard
func (d *Dashboard) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (d *Dashboard) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height

		// Update card dimensions based on grid layout
		d.updateCardDimensions()

		// Don't forward window size to cards - we already calculated their dimensions

	case tea.KeyMsg:
		// Only handle keys if dashboard is focused
		if !d.focused {
			return d, nil
		}

		switch msg.String() {
		case "up", "k":
			d.moveFocusUp()
		case "down", "j":
			d.moveFocusDown()
		case "left", "h":
			d.moveFocusLeft()
		case "right", "l":
			d.moveFocusRight()
		case "enter":
			d.toggleSelection()
		case "esc":
			d.clearSelection()
		}
	}

	return d, nil
}

// View renders the dashboard
func (d *Dashboard) View() string {
	if d.width == 0 || len(d.cards) == 0 {
		return ""
	}

	// Use layout-based rendering for grid
	return d.renderWithLayout()
}

// Focus is called when this component receives focus
func (d *Dashboard) Focus() {
	d.focused = true
}

// Blur is called when this component loses focus
func (d *Dashboard) Blur() {
	d.focused = false
}

// Focused returns whether this component is currently focused
func (d *Dashboard) Focused() bool {
	return d.focused
}

// moveFocusUp moves focus to the card above
func (d *Dashboard) moveFocusUp() {
	if len(d.cards) == 0 {
		return
	}

	// Calculate current row and column
	cols := d.getColumnCount()
	newIndex := d.focusedCardIndex - cols
	if newIndex >= 0 {
		d.setFocusedCard(newIndex)
	}
}

// moveFocusDown moves focus to the card below
func (d *Dashboard) moveFocusDown() {
	if len(d.cards) == 0 {
		return
	}

	// Calculate current row and column
	cols := d.getColumnCount()
	newIndex := d.focusedCardIndex + cols
	if newIndex < len(d.cards) {
		d.setFocusedCard(newIndex)
	}
}

// moveFocusLeft moves focus to the card on the left
func (d *Dashboard) moveFocusLeft() {
	if len(d.cards) == 0 {
		return
	}

	newIndex := d.focusedCardIndex - 1
	if newIndex >= 0 {
		d.setFocusedCard(newIndex)
	}
}

// moveFocusRight moves focus to the card on the right
func (d *Dashboard) moveFocusRight() {
	if len(d.cards) == 0 {
		return
	}

	newIndex := d.focusedCardIndex + 1
	if newIndex < len(d.cards) {
		d.setFocusedCard(newIndex)
	}
}

// toggleSelection toggles selection of the currently focused card
func (d *Dashboard) toggleSelection() {
	if d.focusedCardIndex < 0 || d.focusedCardIndex >= len(d.cards) {
		return
	}

	if d.selectedCardIndex == d.focusedCardIndex {
		// Deselect
		d.cards[d.selectedCardIndex].Deselect()
		d.selectedCardIndex = -1
	} else {
		// Deselect previous if any
		if d.selectedCardIndex >= 0 && d.selectedCardIndex < len(d.cards) {
			d.cards[d.selectedCardIndex].Deselect()
		}
		// Select current
		d.selectedCardIndex = d.focusedCardIndex
		d.cards[d.selectedCardIndex].Select()
	}
}

// clearSelection clears the selection
func (d *Dashboard) clearSelection() {
	if d.selectedCardIndex >= 0 && d.selectedCardIndex < len(d.cards) {
		d.cards[d.selectedCardIndex].Deselect()
		d.selectedCardIndex = -1
	}
}

// setFocusedCard sets the focused card by index
func (d *Dashboard) setFocusedCard(index int) {
	if index < 0 || index >= len(d.cards) {
		return
	}

	// Blur previous card
	if d.focusedCardIndex >= 0 && d.focusedCardIndex < len(d.cards) {
		d.cards[d.focusedCardIndex].Blur()
	}

	// Focus new card
	d.focusedCardIndex = index
	d.cards[d.focusedCardIndex].Focus()
}

// getColumnCount returns the current number of columns in the grid
func (d *Dashboard) getColumnCount() int {
	if !d.responsive {
		return d.columns
	}

	// Calculate responsive columns
	availableWidth := float64(d.width)
	cols := int(availableWidth / (d.minCardWidth + d.gap))
	if cols < 1 {
		cols = 1
	}
	if cols > len(d.cards) {
		cols = len(d.cards)
	}
	return cols
}

// updateCardDimensions calculates and updates card dimensions based on grid layout
func (d *Dashboard) updateCardDimensions() {
	if len(d.cards) == 0 {
		return
	}

	// Calculate columns
	cols := d.columns
	if d.responsive {
		// Calculate columns based on viewport width and min card width
		availableWidth := float64(d.width)
		cols = int(availableWidth / (d.minCardWidth + d.gap))
		if cols < 1 {
			cols = 1
		}
		if cols > len(d.cards) {
			cols = len(d.cards)
		}
	}

	// Calculate card dimensions
	gapTotal := d.gap * float64(cols-1)
	cardWidth := int((float64(d.width) - gapTotal) / float64(cols))
	if cardWidth < 20 {
		cardWidth = 20
	}

	// Calculate rows
	rows := (len(d.cards) + cols - 1) / cols

	// Title takes 3 lines
	titleHeight := 0
	if d.title != "" {
		titleHeight = 3
	}

	// Calculate card height
	availableHeight := d.height - titleHeight
	if rows > 0 {
		gapTotalVertical := d.gap * float64(rows-1)
		cardHeight := int((float64(availableHeight) - gapTotalVertical) / float64(rows))
		if cardHeight < 8 {
			cardHeight = 8
		}

		// Update all cards
		for _, card := range d.cards {
			card.width = cardWidth
			card.height = cardHeight
		}
	}
}

// renderWithLayout renders using the full layout system with CSS Grid
func (d *Dashboard) renderWithLayout() string {
	// For now, use simple string-based rendering since we need to render cards
	// Full layout integration will render cards as layout nodes
	return d.renderSimple()
}

// renderSimple provides string-based rendering with grid-like layout
func (d *Dashboard) renderSimple() string {
	var b strings.Builder

	// Render title if present
	if d.title != "" {
		b.WriteString("╭")
		b.WriteString(strings.Repeat("─", d.width-2))
		b.WriteString("╮\n")

		// Center title
		titlePadding := (d.width - len(d.title) - 2) / 2
		if titlePadding < 0 {
			titlePadding = 0
		}
		b.WriteString("│")
		b.WriteString(strings.Repeat(" ", titlePadding))
		b.WriteString("\033[1m" + d.title + "\033[0m") // Bold
		b.WriteString(strings.Repeat(" ", d.width-len(d.title)-titlePadding-2))
		b.WriteString("│\n")

		b.WriteString("╰")
		b.WriteString(strings.Repeat("─", d.width-2))
		b.WriteString("╯\n")
	}

	// Calculate columns
	cols := d.columns
	if d.responsive {
		availableWidth := float64(d.width)
		cols = int(availableWidth / (d.minCardWidth + d.gap))
		if cols < 1 {
			cols = 1
		}
		if cols > len(d.cards) {
			cols = len(d.cards)
		}
	}

	// Render cards in grid
	for i := 0; i < len(d.cards); i += cols {
		// Get row of cards
		rowCards := d.cards[i:]
		if len(rowCards) > cols {
			rowCards = rowCards[:cols]
		}

		// Render this row
		cardViews := make([]string, len(rowCards))
		maxCardHeight := 0
		for j, card := range rowCards {
			view := card.View()
			cardViews[j] = view
			lines := strings.Split(view, "\n")
			if len(lines) > maxCardHeight {
				maxCardHeight = len(lines)
			}
		}

		// Render row line by line
		for line := 0; line < maxCardHeight; line++ {
			for j, view := range cardViews {
				lines := strings.Split(view, "\n")
				if line < len(lines) {
					b.WriteString(lines[line])
				} else {
					// Pad with spaces if card is shorter
					if len(rowCards) > 0 {
						b.WriteString(strings.Repeat(" ", rowCards[j].width))
					}
				}

				// Add gap between cards
				if j < len(rowCards)-1 {
					b.WriteString(strings.Repeat(" ", int(d.gap)))
				}
			}
			b.WriteString("\n")
		}

		// Add vertical gap between rows
		if i+cols < len(d.cards) {
			for g := 0; g < int(d.gap); g++ {
				b.WriteString("\n")
			}
		}
	}

	return b.String()
}

// AddCard adds a stat card to the dashboard
func (d *Dashboard) AddCard(card *StatCard) {
	d.cards = append(d.cards, card)
	d.updateCardDimensions()
}

// RemoveCard removes a stat card from the dashboard by index
func (d *Dashboard) RemoveCard(index int) {
	if index >= 0 && index < len(d.cards) {
		d.cards = append(d.cards[:index], d.cards[index+1:]...)
		d.updateCardDimensions()
	}
}

// GetCards returns all stat cards
func (d *Dashboard) GetCards() []*StatCard {
	return d.cards
}

// SetCards replaces all stat cards
func (d *Dashboard) SetCards(cards []*StatCard) {
	d.cards = cards
	d.updateCardDimensions()
}

// renderWithGridLayout demonstrates using CSS Grid layout (future enhancement)
func (d *Dashboard) renderWithGridLayout() string {
	// Create layout context
	ctx := layout.NewLayoutContext(float64(d.width), float64(d.height), 16)

	// Choose grid helper based on responsive setting
	var grid *layout.Node
	if d.responsive {
		grid = LayoutHelpers.ResponsiveGridLayout(d.minCardWidth, d.gap)
	} else {
		grid = LayoutHelpers.GridLayout(d.columns, d.gap)
	}

	grid.Style.Width = layout.Px(float64(d.width))
	grid.Style.Height = layout.Px(float64(d.height))

	// Add card nodes as children
	for _, card := range d.cards {
		cardNode := &layout.Node{
			Style: layout.Style{
				Width:  layout.Ch(float64(card.width)),
				Height: layout.Ch(float64(card.height)),
			},
		}
		grid.Children = append(grid.Children, cardNode)
	}

	// Perform layout
	constraints := layout.Tight(float64(d.width), float64(d.height))
	layout.Layout(grid, constraints, ctx)

	// Convert to styled nodes
	textColorRGBA, _ := color.HexToRGB(d.tokens.Color)
	var textColor color.Color = textColorRGBA

	gridStyled := renderer.NewStyledNode(grid, &renderer.Style{
		Foreground: &textColor,
	})

	// Render each card into its grid cell
	for i, card := range d.cards {
		if i < len(grid.Children) {
			cellNode := grid.Children[i]
			cellStyled := renderer.NewStyledNode(cellNode, &renderer.Style{
				Foreground: &textColor,
			})
			cellStyled.Content = card.View()
		}
	}

	// Render to screen
	screen := renderer.NewScreen(d.width, d.height)
	screen.Render(gridStyled)

	return screen.String()
}
