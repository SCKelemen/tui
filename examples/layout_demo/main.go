package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/SCKelemen/cli/renderer"
	"github.com/SCKelemen/color"
	design "github.com/SCKelemen/design-system"
	"github.com/SCKelemen/layout"
)

type model struct {
	width  int
	height int
	tokens *design.DesignTokens
}

func initialModel() model {
	// Use default design tokens
	tokens := design.DefaultTheme()

	return model{
		tokens: tokens,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Create layout context with viewport dimensions
	ctx := layout.NewLayoutContext(float64(m.width), float64(m.height), 16)

	// Create root flexbox container
	root := &layout.Node{
		Style: layout.Style{
			Display:        layout.DisplayFlex,
			FlexDirection:  layout.FlexDirectionColumn,
			Width:          layout.Px(float64(m.width)),
			Height:         layout.Px(float64(m.height)),
			FlexGap:        layout.Ch(1),
			Padding:        layout.Spacing{Top: layout.Ch(1), Bottom: layout.Ch(1)},
		},
	}

	// Add header (fixed height)
	header := &layout.Node{
		Style: layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionRow,
			Height:        layout.Ch(3),
			FlexGap:       layout.Ch(2),
			Padding:       layout.Spacing{Left: layout.Ch(2), Right: layout.Ch(2)},
		},
	}
	root.Children = append(root.Children, header)

	// Header columns (using flex-grow)
	leftCol := &layout.Node{
		Style: layout.Style{
			FlexGrow: 1,
		},
	}
	centerCol := &layout.Node{
		Style: layout.Style{
			FlexGrow: 2,
		},
	}
	rightCol := &layout.Node{
		Style: layout.Style{
			FlexGrow: 1,
		},
	}
	header.Children = append(header.Children, leftCol, centerCol, rightCol)

	// Add main content area (flex-grow to fill remaining space)
	content := &layout.Node{
		Style: layout.Style{
			FlexGrow:  1,
			Padding:   layout.Spacing{Left: layout.Ch(2), Right: layout.Ch(2)},
		},
	}
	root.Children = append(root.Children, content)

	// Add footer (fixed height)
	footer := &layout.Node{
		Style: layout.Style{
			Height:  layout.Ch(1),
			Padding: layout.Spacing{Left: layout.Ch(2), Right: layout.Ch(2)},
		},
	}
	root.Children = append(root.Children, footer)

	// Perform layout calculation
	constraints := layout.Tight(float64(m.width), float64(m.height))
	layout.Layout(root, constraints, ctx)

	// Convert to styled nodes for rendering
	accentColorRGBA, _ := color.HexToRGB(m.tokens.Accent)
	textColorRGBA, _ := color.HexToRGB(m.tokens.Color)

	// Convert to Color interface
	var accentColor color.Color = accentColorRGBA
	var textColor color.Color = textColorRGBA

	rootStyled := m.toStyledNode(root, &textColor)

	// Build content strings based on computed layout
	var leftContent strings.Builder
	leftContent.WriteString("┌─────────────┐\n")
	leftContent.WriteString("│ Layout Demo │\n")
	leftContent.WriteString("└─────────────┘")

	var centerContent strings.Builder
	centerContent.WriteString("══════════════════════\n")
	centerContent.WriteString("   TUI Layout System  \n")
	centerContent.WriteString("══════════════════════")

	var rightContent strings.Builder
	rightContent.WriteString(fmt.Sprintf("Size: %dx%d", m.width, m.height))

	// Set content for each section
	leftStyled := m.toStyledNode(leftCol, &textColor)
	leftStyled.Content = leftContent.String()

	centerStyled := m.toStyledNode(centerCol, &accentColor)
	centerStyled.Content = centerContent.String()

	rightStyled := m.toStyledNode(rightCol, &textColor)
	rightStyled.Content = rightContent.String()

	// Main content
	var mainContent strings.Builder
	mainContent.WriteString("This demo shows the SCKelemen/layout system in action:\n\n")
	mainContent.WriteString("✓ CSS Flexbox layout with flex-grow\n")
	mainContent.WriteString("✓ Viewport-based units (Px, Ch, etc.)\n")
	mainContent.WriteString("✓ Design tokens integration\n")
	mainContent.WriteString("✓ Terminal rendering via cli/renderer\n")
	mainContent.WriteString("✓ Unicode text support via text package\n\n")

	mainContent.WriteString(fmt.Sprintf("Layout calculated positions:\n"))
	mainContent.WriteString(fmt.Sprintf("  Header: x=%0.f y=%0.f w=%0.f h=%0.f\n",
		header.Rect.X, header.Rect.Y,
		header.Rect.Width, header.Rect.Height))
	mainContent.WriteString(fmt.Sprintf("  Content: x=%0.f y=%0.f w=%0.f h=%0.f\n",
		content.Rect.X, content.Rect.Y,
		content.Rect.Width, content.Rect.Height))
	mainContent.WriteString(fmt.Sprintf("  Footer: x=%0.f y=%0.f w=%0.f h=%0.f\n",
		footer.Rect.X, footer.Rect.Y,
		footer.Rect.Width, footer.Rect.Height))

	contentStyled := m.toStyledNode(content, &textColor)
	contentStyled.Content = mainContent.String()

	// Footer
	footerStyled := m.toStyledNode(footer, &textColor)
	footerStyled.Content = "Press 'q' to quit"

	// Render to screen
	screen := renderer.NewScreen(m.width, m.height)
	screen.Render(rootStyled)

	return screen.String()
}

// toStyledNode converts a layout node to a styled rendering node
func (m model) toStyledNode(node *layout.Node, fgColor *color.Color) *renderer.StyledNode {
	style := &renderer.Style{
		Foreground: fgColor,
	}
	return renderer.NewStyledNode(node, style)
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
