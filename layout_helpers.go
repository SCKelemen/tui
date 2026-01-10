package tui

import (
	"github.com/SCKelemen/layout"
)

// LayoutHelper provides common layout patterns and utilities
type LayoutHelper struct{}

// NewLayoutHelper creates a new layout helper
func NewLayoutHelper() *LayoutHelper {
	return &LayoutHelper{}
}

// CenteredOverlay creates a centered overlay node (perfect for modals, dialogs)
// width and height are in viewport units (e.g., 60, 20 for 60ch x 20ch)
func (h *LayoutHelper) CenteredOverlay(width, height float64) *layout.Node {
	return &layout.Node{
		Style: layout.Style{
			Display:        layout.DisplayFlex,
			FlexDirection:  layout.FlexDirectionColumn,
			JustifyContent: layout.JustifyContentCenter,
			AlignItems:     layout.AlignItemsCenter,
			Width:          layout.Vw(100),
			Height:         layout.Vh(100),
		},
		Children: []*layout.Node{
			{
				Style: layout.Style{
					Width:  layout.Ch(width),
					Height: layout.Ch(height),
				},
			},
		},
	}
}

// TwoColumnLayout creates a two-column layout with adjustable ratio
// leftRatio:rightRatio determines the flex-grow distribution (e.g., 1:2)
func (h *LayoutHelper) TwoColumnLayout(leftRatio, rightRatio float64) *layout.Node {
	return &layout.Node{
		Style: layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionRow,
			Width:         layout.Vw(100),
			Height:        layout.Vh(100),
			FlexGap:       layout.Ch(2),
		},
		Children: []*layout.Node{
			{
				Style: layout.Style{
					FlexGrow: leftRatio,
				},
			},
			{
				Style: layout.Style{
					FlexGrow: rightRatio,
				},
			},
		},
	}
}

// ThreeColumnLayout creates a three-column layout with adjustable ratios
func (h *LayoutHelper) ThreeColumnLayout(leftRatio, centerRatio, rightRatio float64) *layout.Node {
	return &layout.Node{
		Style: layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionRow,
			Width:         layout.Vw(100),
			Height:        layout.Vh(100),
			FlexGap:       layout.Ch(2),
		},
		Children: []*layout.Node{
			{Style: layout.Style{FlexGrow: leftRatio}},
			{Style: layout.Style{FlexGrow: centerRatio}},
			{Style: layout.Style{FlexGrow: rightRatio}},
		},
	}
}

// SidebarLayout creates a sidebar + main content layout
// sidebarWidth is in characters (e.g., 20 for 20ch)
func (h *LayoutHelper) SidebarLayout(sidebarWidth float64) *layout.Node {
	return &layout.Node{
		Style: layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionRow,
			Width:         layout.Vw(100),
			Height:        layout.Vh(100),
		},
		Children: []*layout.Node{
			{
				Style: layout.Style{
					Width: layout.Ch(sidebarWidth),
				},
			},
			{
				Style: layout.Style{
					FlexGrow: 1,
				},
			},
		},
	}
}

// HeaderContentFooterLayout creates a classic header/content/footer layout
// headerHeight and footerHeight are in characters
func (h *LayoutHelper) HeaderContentFooterLayout(headerHeight, footerHeight float64) *layout.Node {
	return &layout.Node{
		Style: layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionColumn,
			Width:         layout.Vw(100),
			Height:        layout.Vh(100),
		},
		Children: []*layout.Node{
			{
				Style: layout.Style{
					Height: layout.Ch(headerHeight),
				},
			},
			{
				Style: layout.Style{
					FlexGrow: 1,
				},
			},
			{
				Style: layout.Style{
					Height: layout.Ch(footerHeight),
				},
			},
		},
	}
}

// GridLayout creates a CSS Grid layout with specified columns and rows
// columns is the number of columns, gap is the spacing between cells
func (h *LayoutHelper) GridLayout(columns int, gap float64) *layout.Node {
	// Create grid tracks (1fr for each column = equal distribution)
	gridColumns := make([]layout.GridTrack, columns)
	for i := range gridColumns {
		gridColumns[i] = layout.FractionTrack(1.0)
	}

	return &layout.Node{
		Style: layout.Style{
			Display:             layout.DisplayGrid,
			GridTemplateColumns: gridColumns,
			GridGap:             layout.Ch(gap),
			Width:               layout.Vw(100),
			Height:              layout.Vh(100),
		},
	}
}

// ResponsiveGridLayout creates a grid that adapts to viewport width
// minCardWidth is the minimum width for each card in characters
// gap is the spacing between cards
func (h *LayoutHelper) ResponsiveGridLayout(minCardWidth, gap float64) *layout.Node {
	// Use auto-fill with minmax for responsive columns
	// This makes the grid automatically adjust column count based on available space
	return &layout.Node{
		Style: layout.Style{
			Display: layout.DisplayGrid,
			GridTemplateColumns: []layout.GridTrack{
				layout.MinMaxTrack(layout.Ch(minCardWidth), layout.FractionTrack(1).MaxSize),
			},
			GridAutoRows: layout.AutoTrack(),
			GridGap:      layout.Ch(gap),
			Width:        layout.Vw(100),
			Height:       layout.Vh(100),
		},
	}
}

// CardLayout creates a card-style container with padding and borders
// paddingCh is padding in characters
func (h *LayoutHelper) CardLayout(paddingCh float64) *layout.Node {
	return &layout.Node{
		Style: layout.Style{
			Display: layout.DisplayFlex,
			Padding: layout.Spacing{
				Top:    layout.Ch(paddingCh),
				Bottom: layout.Ch(paddingCh),
				Left:   layout.Ch(paddingCh * 2), // More horizontal padding
				Right:  layout.Ch(paddingCh * 2),
			},
			FlexDirection: layout.FlexDirectionColumn,
		},
	}
}

// StackLayout creates a vertical stack with gap between items
func (h *LayoutHelper) StackLayout(gap float64) *layout.Node {
	return &layout.Node{
		Style: layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionColumn,
			FlexGap:       layout.Ch(gap),
			Width:         layout.Vw(100),
		},
	}
}

// HorizontalStackLayout creates a horizontal stack with gap between items
func (h *LayoutHelper) HorizontalStackLayout(gap float64) *layout.Node {
	return &layout.Node{
		Style: layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionRow,
			FlexGap:       layout.Ch(gap),
			AlignItems:    layout.AlignItemsCenter,
		},
	}
}

// SpaceBetweenRow creates a row with space-between justification
// Perfect for toolbars, headers with left/right content
func (h *LayoutHelper) SpaceBetweenRow() *layout.Node {
	return &layout.Node{
		Style: layout.Style{
			Display:        layout.DisplayFlex,
			FlexDirection:  layout.FlexDirectionRow,
			JustifyContent: layout.JustifyContentSpaceBetween,
			AlignItems:     layout.AlignItemsCenter,
			Width:          layout.Vw(100),
		},
	}
}

// CenteredContent creates a container with centered content
func (h *LayoutHelper) CenteredContent() *layout.Node {
	return &layout.Node{
		Style: layout.Style{
			Display:        layout.DisplayFlex,
			FlexDirection:  layout.FlexDirectionColumn,
			JustifyContent: layout.JustifyContentCenter,
			AlignItems:     layout.AlignItemsCenter,
			Width:          layout.Vw(100),
			Height:         layout.Vh(100),
		},
	}
}

// AbsolutePosition creates a node with absolute positioning
// Useful for overlays, tooltips, floating elements
func (h *LayoutHelper) AbsolutePosition(top, left, width, height float64) *layout.Node {
	return &layout.Node{
		Style: layout.Style{
			Position: layout.PositionAbsolute,
			Top:      layout.Ch(top),
			Left:     layout.Ch(left),
			Width:    layout.Ch(width),
			Height:   layout.Ch(height),
		},
	}
}

// FlexGrowNode creates a simple node that grows to fill available space
func (h *LayoutHelper) FlexGrowNode(grow float64) *layout.Node {
	return &layout.Node{
		Style: layout.Style{
			FlexGrow: grow,
		},
	}
}

// FixedSizeNode creates a node with fixed width and height
func (h *LayoutHelper) FixedSizeNode(width, height float64) *layout.Node {
	return &layout.Node{
		Style: layout.Style{
			Width:  layout.Ch(width),
			Height: layout.Ch(height),
		},
	}
}

// Global helper instance for convenience
var LayoutHelpers = NewLayoutHelper()
