package tui

import (
	"testing"

	"github.com/SCKelemen/layout"
)

// TestLayoutHelperCreation tests that LayoutHelper can be created
func TestLayoutHelperCreation(t *testing.T) {
	helper := NewLayoutHelper()
	if helper == nil {
		t.Fatal("Failed to create LayoutHelper")
	}

	// Test global instance
	if LayoutHelpers == nil {
		t.Fatal("Global LayoutHelpers instance should not be nil")
	}
}

// TestCenteredOverlay tests centered overlay layout
func TestCenteredOverlay(t *testing.T) {
	node := LayoutHelpers.CenteredOverlay(60, 20)

	if node == nil {
		t.Fatal("CenteredOverlay returned nil")
	}

	if node.Style.Display != layout.DisplayFlex {
		t.Error("Expected DisplayFlex")
	}

	if node.Style.JustifyContent != layout.JustifyContentCenter {
		t.Error("Expected JustifyContentCenter")
	}

	if node.Style.AlignItems != layout.AlignItemsCenter {
		t.Error("Expected AlignItemsCenter")
	}

	if len(node.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(node.Children))
	}
}

// TestTwoColumnLayout tests two-column layout
func TestTwoColumnLayout(t *testing.T) {
	node := LayoutHelpers.TwoColumnLayout(1, 2)

	if node == nil {
		t.Fatal("TwoColumnLayout returned nil")
	}

	if node.Style.Display != layout.DisplayFlex {
		t.Error("Expected DisplayFlex")
	}

	if node.Style.FlexDirection != layout.FlexDirectionRow {
		t.Error("Expected FlexDirectionRow")
	}

	if len(node.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(node.Children))
	}

	// Check flex-grow ratios
	if node.Children[0].Style.FlexGrow != 1 {
		t.Errorf("Expected left column flex-grow=1, got %.0f", node.Children[0].Style.FlexGrow)
	}

	if node.Children[1].Style.FlexGrow != 2 {
		t.Errorf("Expected right column flex-grow=2, got %.0f", node.Children[1].Style.FlexGrow)
	}
}

// TestThreeColumnLayout tests three-column layout
func TestThreeColumnLayout(t *testing.T) {
	node := LayoutHelpers.ThreeColumnLayout(1, 2, 1)

	if node == nil {
		t.Fatal("ThreeColumnLayout returned nil")
	}

	if len(node.Children) != 3 {
		t.Fatalf("Expected 3 children, got %d", len(node.Children))
	}

	// Check flex-grow ratios
	if node.Children[0].Style.FlexGrow != 1 {
		t.Errorf("Expected left column flex-grow=1, got %.0f", node.Children[0].Style.FlexGrow)
	}

	if node.Children[1].Style.FlexGrow != 2 {
		t.Errorf("Expected center column flex-grow=2, got %.0f", node.Children[1].Style.FlexGrow)
	}

	if node.Children[2].Style.FlexGrow != 1 {
		t.Errorf("Expected right column flex-grow=1, got %.0f", node.Children[2].Style.FlexGrow)
	}
}

// TestSidebarLayout tests sidebar layout
func TestSidebarLayout(t *testing.T) {
	node := LayoutHelpers.SidebarLayout(20)

	if node == nil {
		t.Fatal("SidebarLayout returned nil")
	}

	if len(node.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(node.Children))
	}

	// Sidebar should have fixed width
	// Main content should have flex-grow=1
	if node.Children[1].Style.FlexGrow != 1 {
		t.Error("Expected main content to have flex-grow=1")
	}
}

// TestHeaderContentFooterLayout tests header/content/footer layout
func TestHeaderContentFooterLayout(t *testing.T) {
	node := LayoutHelpers.HeaderContentFooterLayout(3, 1)

	if node == nil {
		t.Fatal("HeaderContentFooterLayout returned nil")
	}

	if node.Style.FlexDirection != layout.FlexDirectionColumn {
		t.Error("Expected FlexDirectionColumn")
	}

	if len(node.Children) != 3 {
		t.Fatalf("Expected 3 children, got %d", len(node.Children))
	}

	// Content should have flex-grow=1
	if node.Children[1].Style.FlexGrow != 1 {
		t.Error("Expected content area to have flex-grow=1")
	}
}

// TestGridLayout tests CSS Grid layout
func TestGridLayout(t *testing.T) {
	node := LayoutHelpers.GridLayout(3, 2)

	if node == nil {
		t.Fatal("GridLayout returned nil")
	}

	if node.Style.Display != layout.DisplayGrid {
		t.Error("Expected DisplayGrid")
	}

	if len(node.Style.GridTemplateColumns) != 3 {
		t.Errorf("Expected 3 grid columns, got %d", len(node.Style.GridTemplateColumns))
	}

	// Check all columns are fractional
	for i, track := range node.Style.GridTemplateColumns {
		if track.Fraction != 1.0 {
			t.Errorf("Column %d: expected fraction=1.0, got %.1f", i, track.Fraction)
		}
	}
}

// TestResponsiveGridLayout tests responsive grid layout
func TestResponsiveGridLayout(t *testing.T) {
	node := LayoutHelpers.ResponsiveGridLayout(30, 2)

	if node == nil {
		t.Fatal("ResponsiveGridLayout returned nil")
	}

	if node.Style.Display != layout.DisplayGrid {
		t.Error("Expected DisplayGrid")
	}

	if len(node.Style.GridTemplateColumns) != 1 {
		t.Errorf("Expected 1 grid column template (minmax), got %d", len(node.Style.GridTemplateColumns))
	}
}

// TestCardLayout tests card layout
func TestCardLayout(t *testing.T) {
	node := LayoutHelpers.CardLayout(1)

	if node == nil {
		t.Fatal("CardLayout returned nil")
	}

	if node.Style.Display != layout.DisplayFlex {
		t.Error("Expected DisplayFlex")
	}

	if node.Style.FlexDirection != layout.FlexDirectionColumn {
		t.Error("Expected FlexDirectionColumn")
	}

	// Check padding exists
	// Padding values should be non-zero
}

// TestStackLayout tests vertical stack layout
func TestStackLayout(t *testing.T) {
	node := LayoutHelpers.StackLayout(1)

	if node == nil {
		t.Fatal("StackLayout returned nil")
	}

	if node.Style.FlexDirection != layout.FlexDirectionColumn {
		t.Error("Expected FlexDirectionColumn")
	}
}

// TestHorizontalStackLayout tests horizontal stack layout
func TestHorizontalStackLayout(t *testing.T) {
	node := LayoutHelpers.HorizontalStackLayout(1)

	if node == nil {
		t.Fatal("HorizontalStackLayout returned nil")
	}

	if node.Style.FlexDirection != layout.FlexDirectionRow {
		t.Error("Expected FlexDirectionRow")
	}

	if node.Style.AlignItems != layout.AlignItemsCenter {
		t.Error("Expected AlignItemsCenter")
	}
}

// TestSpaceBetweenRow tests space-between row layout
func TestSpaceBetweenRow(t *testing.T) {
	node := LayoutHelpers.SpaceBetweenRow()

	if node == nil {
		t.Fatal("SpaceBetweenRow returned nil")
	}

	if node.Style.JustifyContent != layout.JustifyContentSpaceBetween {
		t.Error("Expected JustifyContentSpaceBetween")
	}

	if node.Style.AlignItems != layout.AlignItemsCenter {
		t.Error("Expected AlignItemsCenter")
	}
}

// TestCenteredContent tests centered content layout
func TestCenteredContent(t *testing.T) {
	node := LayoutHelpers.CenteredContent()

	if node == nil {
		t.Fatal("CenteredContent returned nil")
	}

	if node.Style.JustifyContent != layout.JustifyContentCenter {
		t.Error("Expected JustifyContentCenter")
	}

	if node.Style.AlignItems != layout.AlignItemsCenter {
		t.Error("Expected AlignItemsCenter")
	}
}

// TestAbsolutePosition tests absolute positioning
func TestAbsolutePosition(t *testing.T) {
	node := LayoutHelpers.AbsolutePosition(10, 20, 100, 50)

	if node == nil {
		t.Fatal("AbsolutePosition returned nil")
	}

	if node.Style.Position != layout.PositionAbsolute {
		t.Error("Expected PositionAbsolute")
	}
}

// TestFlexGrowNode tests flex-grow node
func TestFlexGrowNode(t *testing.T) {
	node := LayoutHelpers.FlexGrowNode(2)

	if node == nil {
		t.Fatal("FlexGrowNode returned nil")
	}

	if node.Style.FlexGrow != 2 {
		t.Errorf("Expected flex-grow=2, got %.0f", node.Style.FlexGrow)
	}
}

// TestFixedSizeNode tests fixed size node
func TestFixedSizeNode(t *testing.T) {
	node := LayoutHelpers.FixedSizeNode(100, 50)

	if node == nil {
		t.Fatal("FixedSizeNode returned nil")
	}

	// Width and height should be set
}

// TestLayoutHelperWithContext tests that helpers work with layout context
func TestLayoutHelperWithContext(t *testing.T) {
	ctx := layout.NewLayoutContext(100, 50, 16)

	node := LayoutHelpers.GridLayout(3, 2)

	// Add some child nodes
	for i := 0; i < 6; i++ {
		child := &layout.Node{
			Style: layout.Style{
				Width:  layout.Ch(10),
				Height: layout.Ch(5),
			},
		}
		node.Children = append(node.Children, child)
	}

	// Perform layout
	constraints := layout.Tight(100, 50)
	layout.Layout(node, constraints, ctx)

	// Verify layout was calculated
	if node.Rect.Width == 0 {
		t.Error("Node width should be set after layout")
	}

	if node.Rect.Height == 0 {
		t.Error("Node height should be set after layout")
	}
}

// TestTwoColumnLayoutCalculation tests flex-grow calculation in two columns
func TestTwoColumnLayoutCalculation(t *testing.T) {
	ctx := layout.NewLayoutContext(100, 50, 16)

	node := LayoutHelpers.TwoColumnLayout(1, 2)
	node.Style.Width = layout.Px(100)
	node.Style.Height = layout.Px(50)

	// Perform layout
	constraints := layout.Tight(100, 50)
	layout.Layout(node, constraints, ctx)

	// Get child widths
	left := node.Children[0]
	right := node.Children[1]

	if left.Rect.Width == 0 || right.Rect.Width == 0 {
		t.Error("Child widths should be set after layout")
	}

	// Right column should be approximately 2x left column
	ratio := right.Rect.Width / left.Rect.Width
	if ratio < 1.8 || ratio > 2.2 {
		t.Errorf("Expected flex-grow ratio ~2.0, got %.2f (left=%.0f, right=%.0f)",
			ratio, left.Rect.Width, right.Rect.Width)
	}
}
