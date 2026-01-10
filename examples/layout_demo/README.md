# Layout System Demo

This demo shows the **SCKelemen/layout** system integrated with TUI components, demonstrating how to build terminal UIs using CSS Flexbox layout with viewport-based units.

## What is the Layout System?

The [SCKelemen/layout](https://github.com/SCKelemen/layout) package provides a pure Go implementation of CSS Grid and CSS Flexbox layout engines. It calculates positions and sizes for UI elements using the same algorithms as web browsers.

## Stack Architecture

```
tui (TUI components)
 ├── layout (CSS Flexbox/Grid engine)
 ├── cli/renderer (terminal rendering)
 ├── design-system (themes & tokens)
 ├── color (OKLCH color science)
 └── text (Unicode text measurement)
```

## Key Concepts

### 1. Layout Context

```go
ctx := layout.NewLayoutContext(float64(width), float64(height), 16)
```

The layout context tracks viewport dimensions and font size (16px base) for unit calculations.

### 2. Layout Nodes

```go
root := &layout.Node{
    Style: layout.Style{
        Display:       layout.DisplayFlex,
        FlexDirection: layout.FlexDirectionColumn,
        Width:         layout.Px(float64(width)),
        Height:        layout.Px(float64(height)),
        FlexGap:       layout.Ch(1),  // 1 character gap
    },
}
```

Nodes define the layout tree with CSS-like style properties.

### 3. Viewport Units

- `layout.Px(n)` - Absolute pixels
- `layout.Ch(n)` - Character units (based on font size)
- `layout.Em(n)` - Relative to font size
- `layout.Vw(n)` - Percentage of viewport width
- `layout.Vh(n)` - Percentage of viewport height

### 4. Flexbox Properties

```go
Style: layout.Style{
    Display:        layout.DisplayFlex,
    FlexDirection:  layout.FlexDirectionRow,    // or Column
    FlexGrow:       1,                           // Expand to fill space
    FlexGap:        layout.Ch(2),                // Gap between items
    JustifyContent: layout.JustifyContentCenter,
    AlignItems:     layout.AlignItemsCenter,
}
```

### 5. Layout Calculation

```go
constraints := layout.Tight(float64(width), float64(height))
layout.Layout(root, constraints, ctx)

// Access computed positions
fmt.Printf("x=%0.f y=%0.f w=%0.f h=%0.f\n",
    node.Rect.X, node.Rect.Y,
    node.Rect.Width, node.Rect.Height)
```

### 6. Rendering

```go
// Convert to styled nodes
style := &renderer.Style{
    Foreground: &textColor,
}
styledNode := renderer.NewStyledNode(layoutNode, style)
styledNode.Content = "Hello, World!"

// Render to terminal
screen := renderer.NewScreen(width, height)
screen.Render(styledNode)
return screen.String()
```

## Demo Features

This demo creates a classic 3-section layout:

```
┌─────────────────────────────────┐
│ Header (flex row, 3 columns)   │
│  - Left: 1x grow                │
│  - Center: 2x grow              │
│  - Right: 1x grow               │
├─────────────────────────────────┤
│                                 │
│ Content (flex-grow: 1)          │
│  - Fills remaining space        │
│  - Shows computed positions     │
│                                 │
├─────────────────────────────────┤
│ Footer (fixed height)           │
└─────────────────────────────────┘
```

### Layout Hierarchy

```
root (flex column)
 ├── header (flex row, Ch(3) height)
 │    ├── leftCol (flex-grow: 1)
 │    ├── centerCol (flex-grow: 2)
 │    └── rightCol (flex-grow: 1)
 ├── content (flex-grow: 1)
 └── footer (Ch(1) height)
```

## Usage

```bash
cd examples/layout_demo
go run main.go
```

### Controls

- **q** or **Ctrl+C** - Quit

## Why Use the Layout System?

### Benefits

1. **Responsive** - Flex-grow automatically distributes space
2. **Precise** - Viewport units adapt to terminal size
3. **Maintainable** - CSS-like properties familiar to web developers
4. **Powerful** - Full Flexbox and Grid support
5. **Unicode-aware** - Accurate text width calculations
6. **Type-safe** - Compile-time checking of layout properties

### Compared to Manual Layout

**Manual (string concatenation)**:
```go
// Hard to maintain, brittle
leftWidth := width / 3
centerWidth := width / 3
rightWidth := width - leftWidth - centerWidth
// Handle padding, alignment, wrapping manually...
```

**Layout System**:
```go
// Declarative, automatic
Style: layout.Style{
    Display: layout.DisplayFlex,
    FlexDirection: layout.FlexDirectionRow,
    FlexGap: layout.Ch(2),
}
// Children get flex-grow: 1, 2, 1
```

## Integration with Kitchen Sink

Both kitchen sink demos could be enhanced with the layout system:

### Current (String-based)
- Manual width calculations
- String concatenation
- Hard-coded spacing
- Brittle on resize

### Layout-based (Future)
- CSS Flexbox for headers, columns
- Grid for card galleries
- Viewport units for responsive sizing
- Automatic reflow on resize

## Next Steps

1. **Convert Components** - Migrate Header, StatusBar, StructuredData to use layout
2. **Grid Layouts** - Use CSS Grid for card galleries and dashboards
3. **Viewport Units** - Make all sizing viewport-responsive
4. **Theme Integration** - Deep integration with design tokens
5. **Text Measurement** - Use `text` package for accurate wrapping

## References

- [Layout Package](https://github.com/SCKelemen/layout) - CSS Flexbox/Grid engine
- [CLI Renderer](https://github.com/SCKelemen/cli) - Terminal rendering
- [Design System](https://github.com/SCKelemen/design-system) - Themes and tokens
- [Color](https://github.com/SCKelemen/color) - OKLCH color science
- [Text](https://github.com/SCKelemen/text) - Unicode text measurement

## CSS Flexbox Resources

- [MDN: Flexbox](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_Flexible_Box_Layout)
- [CSS Flexbox Spec](https://www.w3.org/TR/css-flexbox-1/)
- [Flexbox Froggy](https://flexboxfroggy.com/) - Interactive tutorial
