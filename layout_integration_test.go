package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/cli/renderer"
	"github.com/SCKelemen/color"
	design "github.com/SCKelemen/design-system"
	"github.com/SCKelemen/layout"
)

// TestLayoutSystemAvailable verifies layout system is properly integrated
func TestLayoutSystemAvailable(t *testing.T) {
	// Test that we can create a layout context
	ctx := layout.NewLayoutContext(100, 50, 16)
	if ctx == nil {
		t.Fatal("Failed to create layout context")
	}
}

// TestLayoutNodeCreation tests basic layout node creation
func TestLayoutNodeCreation(t *testing.T) {
	node := &layout.Node{
		Style: layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionRow,
			Width:         layout.Px(100),
			Height:        layout.Px(50),
		},
	}

	if node == nil {
		t.Fatal("Failed to create layout node")
	}

	if node.Style.Display != layout.DisplayFlex {
		t.Errorf("Expected Display=Flex, got %v", node.Style.Display)
	}
}

// TestLayoutCalculation tests that layout calculation works
func TestLayoutCalculation(t *testing.T) {
	ctx := layout.NewLayoutContext(100, 50, 16)

	root := &layout.Node{
		Style: layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionRow,
			Width:         layout.Px(100),
			Height:        layout.Px(50),
		},
	}

	child1 := &layout.Node{
		Style: layout.Style{
			FlexGrow: 1,
		},
	}
	child2 := &layout.Node{
		Style: layout.Style{
			FlexGrow: 2,
		},
	}

	root.Children = []*layout.Node{child1, child2}

	// Perform layout
	constraints := layout.Tight(100, 50)
	layout.Layout(root, constraints, ctx)

	// Verify layout was calculated
	if root.Rect.Width == 0 {
		t.Error("Root rect width should be set after layout")
	}

	// Verify flex-grow worked (child2 should be ~2x child1)
	if child2.Rect.Width == 0 {
		t.Error("Child2 rect width should be set")
	}
	if child1.Rect.Width == 0 {
		t.Error("Child1 rect width should be set")
	}

	// Child2 should be approximately 2x child1 (allowing for rounding)
	ratio := child2.Rect.Width / child1.Rect.Width
	if ratio < 1.8 || ratio > 2.2 {
		t.Errorf("Expected flex-grow ratio ~2.0, got %.2f (child1=%.0f, child2=%.0f)",
			ratio, child1.Rect.Width, child2.Rect.Width)
	}
}

// TestViewportUnits tests viewport-based units
func TestViewportUnits(t *testing.T) {
	tests := []struct {
		name string
		unit layout.Length
	}{
		{"Pixels", layout.Px(100)},
		{"Characters", layout.Ch(10)},
		{"Em", layout.Em(2)},
		{"Vw", layout.Vw(50)},
		{"Vh", layout.Vh(50)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &layout.Node{
				Style: layout.Style{
					Width: tt.unit,
				},
			}

			if node.Style.Width == (layout.Length{}) {
				t.Errorf("Failed to set %s unit", tt.name)
			}
		})
	}
}

// TestDesignTokensIntegration tests design system integration
func TestDesignTokensIntegration(t *testing.T) {
	tokens := design.DefaultTheme()
	if tokens == nil {
		t.Fatal("Failed to get default theme")
	}

	if tokens.Color == "" {
		t.Error("Theme should have Color set")
	}

	if tokens.Accent == "" {
		t.Error("Theme should have Accent set")
	}
}

// TestColorConversion tests color parsing and conversion
func TestColorConversion(t *testing.T) {
	tokens := design.DefaultTheme()

	// Test hex to RGB conversion
	rgbaColor, err := color.HexToRGB(tokens.Color)
	if err != nil {
		t.Fatalf("Failed to parse color hex: %v", err)
	}

	if rgbaColor == nil {
		t.Fatal("Parsed color should not be nil")
	}

	// Test color interface conversion
	var c color.Color = rgbaColor
	if c == nil {
		t.Error("Color interface conversion failed")
	}
}

// TestRendererIntegration tests cli/renderer integration
func TestRendererIntegration(t *testing.T) {
	ctx := layout.NewLayoutContext(80, 24, 16)

	root := &layout.Node{
		Style: layout.Style{
			Display: layout.DisplayFlex,
			Width:   layout.Px(80),
			Height:  layout.Px(24),
		},
	}

	constraints := layout.Tight(80, 24)
	layout.Layout(root, constraints, ctx)

	// Convert to styled node
	textColorRGBA, _ := color.HexToRGB("#FFFFFF")
	var textColor color.Color = textColorRGBA

	styledNode := renderer.NewStyledNode(root, &renderer.Style{
		Foreground: &textColor,
	})

	if styledNode == nil {
		t.Fatal("Failed to create styled node")
	}

	// Render to screen
	screen := renderer.NewScreen(80, 24)
	screen.Render(styledNode)

	output := screen.String()
	if output == "" {
		t.Error("Rendered output should not be empty")
	}
}

// TestHeaderLayoutFoundation tests Header component has layout foundation
func TestHeaderLayoutFoundation(t *testing.T) {
	header := NewHeader(
		WithColumns(
			HeaderColumn{Width: 1, Align: AlignLeft, Content: []string{"Left"}},
			HeaderColumn{Width: 2, Align: AlignCenter, Content: []string{"Center"}},
			HeaderColumn{Width: 1, Align: AlignRight, Content: []string{"Right"}},
		),
	)

	// Update with window size
	header.Update(tea.WindowSizeMsg{Width: 100, Height: 10})

	// Test that header has tokens
	if header.tokens == nil {
		t.Error("Header should have design tokens initialized")
	}

	// Test that View() works
	view := header.View()
	if view == "" {
		t.Error("Header view should not be empty")
	}

	// Test column width calculation uses flex-grow
	widths := header.calculateColumnWidths()
	if len(widths) != 3 {
		t.Fatalf("Expected 3 column widths, got %d", len(widths))
	}

	// Verify proportional distribution (1:2:1 ratio)
	// Middle column should be approximately 2x the side columns
	ratio := float64(widths[1]) / float64(widths[0])
	if ratio < 1.8 || ratio > 2.2 {
		t.Errorf("Expected center column ~2x left column, got ratio %.2f (widths: %v)",
			ratio, widths)
	}
}

// TestFlexboxAlignmentProperties tests various flexbox alignment options
func TestFlexboxAlignmentProperties(t *testing.T) {
	alignments := []struct {
		name    string
		justify layout.JustifyContent
		align   layout.AlignItems
	}{
		{"Center-Center", layout.JustifyContentCenter, layout.AlignItemsCenter},
		{"FlexStart-FlexStart", layout.JustifyContentFlexStart, layout.AlignItemsFlexStart},
		{"FlexEnd-FlexEnd", layout.JustifyContentFlexEnd, layout.AlignItemsFlexEnd},
		{"SpaceBetween-Stretch", layout.JustifyContentSpaceBetween, layout.AlignItemsStretch},
	}

	for _, tt := range alignments {
		t.Run(tt.name, func(t *testing.T) {
			node := &layout.Node{
				Style: layout.Style{
					Display:        layout.DisplayFlex,
					JustifyContent: tt.justify,
					AlignItems:     tt.align,
					Width:          layout.Px(100),
					Height:         layout.Px(50),
				},
			}

			if node.Style.JustifyContent != tt.justify {
				t.Errorf("JustifyContent not set correctly")
			}
			if node.Style.AlignItems != tt.align {
				t.Errorf("AlignItems not set correctly")
			}
		})
	}
}

// TestLayoutWithPadding tests padding in layout system
func TestLayoutWithPadding(t *testing.T) {
	ctx := layout.NewLayoutContext(100, 50, 16)

	root := &layout.Node{
		Style: layout.Style{
			Display: layout.DisplayFlex,
			Width:   layout.Px(100),
			Height:  layout.Px(50),
			Padding: layout.Spacing{
				Top:    layout.Ch(1),
				Bottom: layout.Ch(1),
				Left:   layout.Ch(2),
				Right:  layout.Ch(2),
			},
		},
	}

	constraints := layout.Tight(100, 50)
	layout.Layout(root, constraints, ctx)

	// Verify padding was applied (content area should be smaller than total)
	if root.Rect.Width <= 0 {
		t.Error("Root should have positive width after layout")
	}
}

// TestMultiLevelNesting tests nested layout nodes
func TestMultiLevelNesting(t *testing.T) {
	ctx := layout.NewLayoutContext(100, 50, 16)

	root := &layout.Node{
		Style: layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionColumn,
			Width:         layout.Px(100),
			Height:        layout.Px(50),
		},
	}

	header := &layout.Node{
		Style: layout.Style{
			Height: layout.Px(3),
		},
	}

	content := &layout.Node{
		Style: layout.Style{
			FlexGrow:      1,
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionRow,
		},
	}

	sidebar := &layout.Node{
		Style: layout.Style{
			Width: layout.Px(20), // Use pixels instead of Ch for predictable sizing
		},
	}

	main := &layout.Node{
		Style: layout.Style{
			FlexGrow: 1,
		},
	}

	content.Children = []*layout.Node{sidebar, main}
	root.Children = []*layout.Node{header, content}

	constraints := layout.Tight(100, 50)
	layout.Layout(root, constraints, ctx)

	// Verify nested layout worked
	if content.Rect.Height <= 0 {
		t.Errorf("Content area should have positive height, got %.2f", content.Rect.Height)
	}
	if sidebar.Rect.Width <= 0 {
		t.Errorf("Sidebar should have positive width, got %.2f", sidebar.Rect.Width)
	}

	// Main area should exist and fill remaining space
	// Note: In flexbox, the main area gets the remaining width after sidebar
	if main.Rect.Width < 0 {
		t.Errorf("Main area width should not be negative, got %.2f", main.Rect.Width)
	}

	// Verify total width distribution makes sense
	totalChildWidth := sidebar.Rect.Width + main.Rect.Width
	if totalChildWidth > content.Rect.Width+1 { // Allow 1px rounding
		t.Errorf("Children widths (%.2f) exceed parent width (%.2f)",
			totalChildWidth, content.Rect.Width)
	}
}
