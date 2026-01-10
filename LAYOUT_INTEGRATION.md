# Layout System Integration Status

## Overview

The SCKelemen/layout system is **fully integrated and ready to use** in the TUI framework. This document explains the hybrid approach we're taking.

## Integration Status: ✅ Complete

### What's Working

1. **Layout System Dependencies** - All installed and configured
   - `github.com/SCKelemen/layout` - CSS Flexbox/Grid engine
   - `github.com/SCKelemen/color` - OKLCH color science
   - `github.com/SCKelemen/design-system` - Themes and tokens
   - `github.com/SCKelemen/cli/renderer` - Terminal rendering
   - `github.com/SCKelemen/text` - Unicode text measurement

2. **Working Demo** - `examples/layout_demo/`
   - Full stack integration example
   - CSS Flexbox layout with flex-grow
   - Viewport units (Px, Ch, Em, Vw, Vh)
   - Design tokens integration
   - Terminal rendering via cli/renderer

3. **Component Foundations**
   - Header component has layout foundation with `renderWithLayout()` skeleton
   - All components have design tokens available
   - Build system configured with local module replacements

## Hybrid Approach: Simple + Powerful

We're using a **pragmatic hybrid** strategy:

### ✅ Keep String-Based (Simple, Fast, Works Well)

These components use direct string concatenation and work perfectly:

- **StatusBar** - Single line status display
- **ActivityBar** - Single line with spinner animation
- **TextInput** - Single line text input
- **FileExplorer** - Tree view with simple indentation

**Why**: These are simple, linear components. String-based rendering is:
- ✅ Fast to render
- ✅ Easy to understand
- ✅ Minimal dependencies
- ✅ Already working well

### 🎯 Use Layout System (Complex, Powerful)

Use layout for **new** complex components or when rewriting:

**Perfect Use Cases**:
- **Multi-column layouts** - Headers with proportional columns
- **Centered overlays** - Modals, dialogs (auto-centering)
- **Grid layouts** - Card galleries, dashboards
- **Responsive sizing** - Components that adapt to viewport
- **Complex alignment** - Nested flexbox/grid structures

**Benefits**:
- Automatic centering and alignment
- Flex-grow for proportional distribution
- Viewport-relative sizing
- CSS-like declarative layout
- Consistent spacing with design tokens

## How to Use Layout System

### Example: Creating a Layout-Based Component

```go
package tui

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/SCKelemen/cli/renderer"
    "github.com/SCKelemen/color"
    design "github.com/SCKelemen/design-system"
    "github.com/SCKelemen/layout"
)

type MyComponent struct {
    width  int
    height int
    tokens *design.DesignTokens
}

func (c *MyComponent) View() string {
    // 1. Create layout context
    ctx := layout.NewLayoutContext(float64(c.width), float64(c.height), 16)

    // 2. Build layout tree with CSS Flexbox
    root := &layout.Node{
        Style: layout.Style{
            Display:        layout.DisplayFlex,
            FlexDirection:  layout.FlexDirectionRow,
            Width:          layout.Px(float64(c.width)),
            Height:         layout.Px(float64(c.height)),
            JustifyContent: layout.JustifyContentCenter,
            AlignItems:     layout.AlignItemsCenter,
            FlexGap:        layout.Ch(2),
        },
    }

    // 3. Add child nodes
    leftCol := &layout.Node{
        Style: layout.Style{FlexGrow: 1},
    }
    rightCol := &layout.Node{
        Style: layout.Style{FlexGrow: 2},
    }
    root.Children = []*layout.Node{leftCol, rightCol}

    // 4. Calculate layout
    constraints := layout.Tight(float64(c.width), float64(c.height))
    layout.Layout(root, constraints, ctx)

    // 5. Convert to styled nodes
    textColorRGBA, _ := color.HexToRGB(c.tokens.Color)
    var textColor color.Color = textColorRGBA

    rootStyled := renderer.NewStyledNode(root, &renderer.Style{
        Foreground: &textColor,
    })

    // 6. Set content
    leftStyled := renderer.NewStyledNode(leftCol, &renderer.Style{
        Foreground: &textColor,
    })
    leftStyled.Content = "Left content"

    rightStyled := renderer.NewStyledNode(rightCol, &renderer.Style{
        Foreground: &textColor,
    })
    rightStyled.Content = "Right content (2x wider)"

    // 7. Render
    screen := renderer.NewScreen(c.width, c.height)
    screen.Render(rootStyled)
    return screen.String()
}
```

### Key Concepts

**Layout Nodes** - Define structure with CSS properties:
```go
&layout.Node{
    Style: layout.Style{
        Display:       layout.DisplayFlex,
        FlexDirection: layout.FlexDirectionRow,
        FlexGrow:      1,
        Width:         layout.Px(100),
        Height:        layout.Ch(3),
        Padding:       layout.Spacing{Top: layout.Ch(1)},
    },
}
```

**Viewport Units**:
- `layout.Px(n)` - Absolute pixels
- `layout.Ch(n)` - Character units (based on font size)
- `layout.Em(n)` - Relative to font size
- `layout.Vw(n)` - Percentage of viewport width
- `layout.Vh(n)` - Percentage of viewport height

**Flexbox Properties**:
- `FlexDirection` - Row, Column, RowReverse, ColumnReverse
- `FlexGrow` - Proportional sizing (1, 2, 3 = 1:2:3 ratio)
- `FlexGap` - Spacing between flex items
- `JustifyContent` - Main axis alignment
- `AlignItems` - Cross axis alignment

## Decision Guide

**Use String-Based When**:
- ✅ Component is simple (single line, straightforward layout)
- ✅ Already works well
- ✅ Performance is critical (hot path)
- ✅ No responsive/complex layout needs

**Use Layout System When**:
- ✅ Multi-column proportional layouts
- ✅ Centering/overlay positioning
- ✅ Responsive viewport-relative sizing
- ✅ Complex nested structures
- ✅ Grid-based card layouts
- ✅ Building new complex components

## Examples

### Simple String-Based: StatusBar
```go
func (s *StatusBar) View() string {
    return s.message + strings.Repeat(" ", s.width-len(s.message))
}
```
**Why**: Single line, simple padding. String concat is perfect.

### Layout-Based: Centered Modal
```go
root := &layout.Node{
    Style: layout.Style{
        Display:        layout.DisplayFlex,
        JustifyContent: layout.JustifyContentCenter,  // Auto-center
        AlignItems:     layout.AlignItemsCenter,      // Auto-center
        Width:          layout.Vw(100),                // Full viewport
        Height:         layout.Vh(100),                // Full viewport
    },
}

modal := &layout.Node{
    Style: layout.Style{
        Width:  layout.Ch(60),  // Fixed width
        Height: layout.Ch(20),  // Fixed height
    },
}
root.Children = []*layout.Node{modal}
```
**Why**: Automatic centering in viewport. Layout system excels here.

## Current Component Status

| Component | Type | Reason |
|-----------|------|--------|
| StatusBar | String | ✅ Simple single line |
| ActivityBar | String | ✅ Simple single line + spinner |
| TextInput | String | ✅ Simple single line input |
| FileExplorer | String | ✅ Works well with indentation |
| Header | Hybrid | 🔄 Has layout foundation, uses flex-grow |
| Modal | String | ⏳ Could benefit from layout for centering |
| CommandPalette | String | ⏳ Could benefit from layout for overlay |
| StructuredData | String | ⏳ Could benefit from layout for alignment |
| ToolBlock | String | ⏳ Could benefit from layout for sections |

## Next Steps

1. **Use Layout for New Components** - Start with layout system
2. **Gradually Migrate Complex Ones** - As they're refactored
3. **Document Patterns** - Build pattern library for common layouts
4. **Performance Test** - Measure layout vs string rendering

## Resources

- [Layout Package](https://github.com/SCKelemen/layout) - Full documentation
- [layout_demo/](examples/layout_demo/) - Working example
- [CSS Flexbox Guide](https://css-tricks.com/snippets/css/a-guide-to-flexbox/)
- [CSS Grid Guide](https://css-tricks.com/snippets/css/complete-guide-grid/)

## Conclusion

The layout system is **ready to use** whenever you need powerful, flexible layouts. For simple components, string-based rendering is perfectly fine. **Use the right tool for the job.**
