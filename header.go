package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/cli/renderer"
	"github.com/SCKelemen/color"
	design "github.com/SCKelemen/design-system"
	"github.com/SCKelemen/layout"
)

// ColumnAlign defines how content is aligned within a column
type ColumnAlign int

const (
	AlignLeft ColumnAlign = iota
	AlignCenter
	AlignRight
)

// HeaderColumn represents a column in the header
type HeaderColumn struct {
	Width   int         // Flex-grow value (0 = equal distribution with others)
	Align   ColumnAlign // Content alignment
	Content []string    // Lines of content
}

// HeaderSection represents a section within a column with optional divider
type HeaderSection struct {
	Title   string   // Section title (optional)
	Content []string // Section content lines
	Divider bool     // Show horizontal divider before section
}

// Header displays a multi-column header using the layout system
type Header struct {
	width       int
	height      int
	columns     []HeaderColumn
	sections    map[int][]HeaderSection
	showDivider bool
	focused     bool
	tokens      *design.DesignTokens
}

// HeaderOption configures a Header
type HeaderOption func(*Header)

// WithColumns sets the columns for the header
func WithColumns(columns ...HeaderColumn) HeaderOption {
	return func(h *Header) {
		h.columns = columns
	}
}

// WithColumnSections sets sections for a specific column
func WithColumnSections(columnIndex int, sections ...HeaderSection) HeaderOption {
	return func(h *Header) {
		if h.sections == nil {
			h.sections = make(map[int][]HeaderSection)
		}
		h.sections[columnIndex] = sections
	}
}

// WithVerticalDivider enables/disables the vertical divider between columns
func WithVerticalDivider(show bool) HeaderOption {
	return func(h *Header) {
		h.showDivider = show
	}
}

// NewHeader creates a new header component
func NewHeader(opts ...HeaderOption) *Header {
	h := &Header{
		columns:     []HeaderColumn{},
		sections:    make(map[int][]HeaderSection),
		showDivider: true,
		tokens:      design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Init initializes the header
func (h *Header) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (h *Header) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height
	}

	return h, nil
}

// View renders the header using layout system
func (h *Header) View() string {
	if h.width == 0 || len(h.columns) == 0 {
		return ""
	}

	// Fallback: use simple string rendering for now to avoid breaking things
	// TODO: Implement full layout-based rendering
	return h.renderSimple()
}

// renderSimple provides a simple string-based rendering (temporary)
func (h *Header) renderSimple() string {
	var b strings.Builder

	// Calculate column widths
	columnWidths := h.calculateColumnWidths()
	totalWidth := 0
	for _, w := range columnWidths {
		totalWidth += w
	}
	// Add space for dividers
	if h.showDivider && len(columnWidths) > 1 {
		totalWidth += len(columnWidths) - 1
	}
	// Add borders
	totalWidth += 2

	// Calculate content height
	contentHeight := h.calculateContentHeight()

	// Top border
	b.WriteString("╭")
	b.WriteString(strings.Repeat("─", totalWidth-2))
	b.WriteString("╮\n")

	// Render content rows
	for row := 0; row < contentHeight; row++ {
		b.WriteString("│")

		for colIdx, colWidth := range columnWidths {
			// Get content for this row/column
			content := h.getColumnContent(colIdx, row)

			// Apply alignment
			aligned := h.alignContent(content, colWidth, h.columns[colIdx].Align)
			b.WriteString(aligned)

			// Add vertical divider
			if h.showDivider && colIdx < len(columnWidths)-1 {
				b.WriteString("│")
			}
		}

		b.WriteString("│\n")
	}

	// Bottom border
	b.WriteString("╰")
	b.WriteString(strings.Repeat("─", totalWidth-2))
	b.WriteString("╯\n")

	return b.String()
}

// Focus is called when this component receives focus
func (h *Header) Focus() {
	h.focused = true
}

// Blur is called when this component loses focus
func (h *Header) Blur() {
	h.focused = false
}

// Focused returns whether this component is currently focused
func (h *Header) Focused() bool {
	return h.focused
}

// calculateColumnWidths calculates the actual pixel widths for each column using flex-grow
func (h *Header) calculateColumnWidths() []int {
	if len(h.columns) == 0 {
		return nil
	}

	availableWidth := h.width - 2 // Account for borders
	if h.showDivider && len(h.columns) > 1 {
		availableWidth -= len(h.columns) - 1 // Account for dividers
	}

	widths := make([]int, len(h.columns))

	// Calculate total flex-grow
	totalGrow := 0
	for _, col := range h.columns {
		if col.Width > 0 {
			totalGrow += col.Width
		} else {
			totalGrow += 1 // Default grow
		}
	}

	// Distribute width based on flex-grow ratios
	remaining := availableWidth
	for i, col := range h.columns {
		grow := col.Width
		if grow <= 0 {
			grow = 1
		}

		if i == len(h.columns)-1 {
			// Last column gets remaining width
			widths[i] = remaining
		} else {
			widths[i] = (availableWidth * grow) / totalGrow
			remaining -= widths[i]
		}
	}

	return widths
}

// calculateContentHeight calculates the total height needed for content
func (h *Header) calculateContentHeight() int {
	maxHeight := 0

	for colIdx := range h.columns {
		// Check if column has sections
		if sections, ok := h.sections[colIdx]; ok {
			height := 0
			for _, section := range sections {
				if section.Divider && height > 0 {
					height++ // Divider line
				}
				if section.Title != "" {
					height++ // Title line
				}
				height += len(section.Content)
			}
			if height > maxHeight {
				maxHeight = height
			}
		} else {
			// Use column content
			if len(h.columns[colIdx].Content) > maxHeight {
				maxHeight = len(h.columns[colIdx].Content)
			}
		}
	}

	// Add padding
	return maxHeight + 2
}

// getColumnContent gets the content for a specific column and row
func (h *Header) getColumnContent(colIdx, row int) string {
	// Check if column has sections
	if sections, ok := h.sections[colIdx]; ok {
		currentRow := 0

		// Add top padding
		if row == 0 {
			return ""
		}
		currentRow++

		for _, section := range sections {
			// Divider
			if section.Divider && currentRow > 1 {
				if row == currentRow {
					// Return horizontal divider
					return "─────────────────────"
				}
				currentRow++
			}

			// Section title
			if section.Title != "" {
				if row == currentRow {
					return section.Title
				}
				currentRow++
			}

			// Section content
			for _, line := range section.Content {
				if row == currentRow {
					return line
				}
				currentRow++
			}
		}

		return ""
	}

	// Use column content
	if row == 0 || row >= len(h.columns[colIdx].Content)+1 {
		return "" // Padding
	}

	lineIdx := row - 1
	if lineIdx < len(h.columns[colIdx].Content) {
		return h.columns[colIdx].Content[lineIdx]
	}

	return ""
}

// alignContent aligns content within a given width
func (h *Header) alignContent(content string, width int, align ColumnAlign) string {
	contentWidth := len(content) // Simple length, should use text package for Unicode

	if contentWidth >= width {
		// Truncate if too long
		if contentWidth > width {
			if width > 3 {
				return content[:width-3] + "..."
			}
			return content[:width]
		}
		return content
	}

	padding := width - contentWidth

	switch align {
	case AlignLeft:
		return content + strings.Repeat(" ", padding)
	case AlignRight:
		return strings.Repeat(" ", padding) + content
	case AlignCenter:
		leftPad := padding / 2
		rightPad := padding - leftPad
		return strings.Repeat(" ", leftPad) + content + strings.Repeat(" ", rightPad)
	}

	return content
}

// renderWithLayout renders using the full layout system (work in progress)
func (h *Header) renderWithLayout() string {
	// Create layout context
	ctx := layout.NewLayoutContext(float64(h.width), float64(h.height), 16)

	// Create root flexbox container
	root := &layout.Node{
		Style: layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionRow,
			Width:         layout.Px(float64(h.width)),
			AlignItems:    layout.AlignItemsStretch,
			Padding: layout.Spacing{
				Top:    layout.Ch(0.5),
				Bottom: layout.Ch(0.5),
				Left:   layout.Ch(1),
				Right:  layout.Ch(1),
			},
		},
	}

	// Add columns with flex-grow
	for _, col := range h.columns {
		colNode := &layout.Node{
			Style: layout.Style{
				FlexGrow: float64(col.Width),
			},
		}

		if col.Width <= 0 {
			colNode.Style.FlexGrow = 1
		}

		root.Children = append(root.Children, colNode)
	}

	// Perform layout
	constraints := layout.Tight(float64(h.width), float64(h.height))
	layout.Layout(root, constraints, ctx)

	// Convert to styled nodes
	textColorRGBA, _ := color.HexToRGB(h.tokens.Color)
	var textColor color.Color = textColorRGBA

	rootStyled := renderer.NewStyledNode(root, &renderer.Style{
		Foreground: &textColor,
	})

	// TODO: Build content for each column child node

	// Render to screen
	screen := renderer.NewScreen(h.width, 10)
	screen.Render(rootStyled)

	return screen.String()
}
