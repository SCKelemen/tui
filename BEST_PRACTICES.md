# TUI Best Practices Guide

This guide covers design patterns, common pitfalls, performance optimization, and testing strategies for building robust terminal applications with the TUI framework.

## Table of Contents

- [Component Composition Patterns](#component-composition-patterns)
- [Focus Management](#focus-management)
- [Layout and Responsive Design](#layout-and-responsive-design)
- [Common Pitfalls](#common-pitfalls)
- [Performance Optimization](#performance-optimization)
- [Testing Strategies](#testing-strategies)
- [Error Handling](#error-handling)
- [Keyboard Navigation](#keyboard-navigation)
- [Visual Design](#visual-design)

---

## Component Composition Patterns

### Building Complex UIs

**DO**: Compose UIs from small, focused components

```go
type AppModel struct {
    dashboard   *tui.Dashboard
    statusBar   *tui.StatusBar
    commandPal  *tui.CommandPalette
    fileExplorer *tui.FileExplorer
    activeView   string // "dashboard", "explorer", etc.
}

func (m AppModel) View() string {
    var content string

    switch m.activeView {
    case "dashboard":
        content = m.dashboard.View()
    case "explorer":
        content = m.fileExplorer.View()
    }

    return content + "\n" + m.statusBar.View()
}
```

**DON'T**: Create monolithic components that do everything

```go
// Bad: Single component handling too many responsibilities
type MegaComponent struct {
    dashboard       bool
    fileTree        bool
    statusBar       bool
    commandPalette  bool
    // ... too many concerns
}
```

### Parent-Child Communication

**Pattern: Message Delegation**

```go
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // Propagate resize to all children
        m.dashboard.Update(msg)
        m.statusBar.Update(msg)
        m.fileExplorer.Update(msg)

    case tea.KeyMsg:
        // Route keyboard input to focused component
        if m.commandPal.IsVisible() {
            m.commandPal.Update(msg)
        } else if m.dashboard.Focused() {
            m.dashboard.Update(msg)
        }
    }

    return m, nil
}
```

### When to Create Custom Components

**Create a new component when**:
- Logic is reused across multiple views
- Component has clear, single responsibility
- State management becomes complex
- You need isolated testing

**Use existing components when**:
- Simple, one-off UI elements
- Minimal state management needed
- Already covered by TUI library

### Component Interface Pattern

All TUI components implement this interface:

```go
type Component interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (Component, tea.Cmd)
    View() string
    Focus()
    Blur()
    Focused() bool
}
```

**Always implement all methods** even if some are no-ops:

```go
func (c *MyComponent) Init() tea.Cmd {
    return nil // No initialization needed
}
```

---

## Focus Management

### Focus Lifecycle

**Pattern: Single Focus Source**

Only ONE component should be focused at any time.

```go
type AppModel struct {
    dashboard *tui.Dashboard
    modal     *tui.DetailModal
}

func (m *AppModel) showModal() {
    // Transfer focus: dashboard → modal
    m.dashboard.Blur()
    m.modal.Show() // Show() automatically focuses
}

func (m *AppModel) closeModal() {
    // Restore focus: modal → dashboard
    m.modal.Hide() // Hide() automatically blurs
    m.dashboard.Focus()
}
```

### Focus State Tracking

**DO**: Use boolean flags for focus state

```go
type Component struct {
    focused bool
}

func (c *Component) Focus() {
    c.focused = true
}

func (c *Component) Blur() {
    c.focused = false
}

func (c *Component) Focused() bool {
    return c.focused
}
```

**DON'T**: Rely on derived state

```go
// Bad: No explicit focus tracking
func (c *Component) Focused() bool {
    return c.borderColor == "cyan" // Fragile!
}
```

### Modal Focus Pattern

When opening modals, follow this pattern:

```go
func (d *Dashboard) openModal() {
    // 1. Populate modal with data
    card := d.cards[d.focusedCardIndex]
    d.modal.SetContent(card)

    // 2. Show modal (focuses automatically)
    d.modal.Show()

    // 3. Blur parent
    d.focused = false
}

func (d *Dashboard) Update(msg tea.Msg) (Component, tea.Cmd) {
    // Check if modal is visible FIRST
    if d.modal.IsVisible() {
        d.modal.Update(msg)

        // Restore focus when modal closes
        if !d.modal.IsVisible() {
            d.focused = true
        }

        return d, nil
    }

    // Handle dashboard input only if focused
    if d.focused {
        // ... handle keyboard navigation
    }

    return d, nil
}
```

### Visual Focus Indicators

**Use distinct visual states**:

- **Normal**: Dim colors, thin borders `┌─┐`
- **Focused**: Bright colors, double borders `╔═╗`
- **Selected**: Accent colors, thick borders `┏━┓`

```go
func (s *StatCard) getBorderStyle() borderStyle {
    if s.focused {
        return borderStyle{
            topLeft: "╔", horizontal: "═",
            color: "\033[36m", // Cyan
        }
    } else if s.selected {
        return borderStyle{
            topLeft: "┏", horizontal: "━",
            color: "\033[33m", // Yellow
        }
    }
    return borderStyle{
        topLeft: "┌", horizontal: "─",
        color: "", // Default
    }
}
```

---

## Layout and Responsive Design

### Responsive Grid Patterns

**DO**: Calculate columns based on terminal width

```go
func (d *Dashboard) getColumnCount() int {
    if !d.responsive {
        return d.columns
    }

    // Auto-calculate based on width
    minCardWidth := 30
    maxColumns := d.width / (minCardWidth + d.gap)

    if maxColumns < 1 {
        return 1
    }
    if maxColumns > d.columns {
        return d.columns
    }

    return maxColumns
}
```

**DON'T**: Use fixed layouts

```go
// Bad: Breaks on narrow terminals
func (d *Dashboard) render() string {
    // Always 3 columns, even if terminal is 60 chars wide
    columns := 3
    cardWidth := d.width / 3 // May be too narrow!
}
```

### Content Area Calculation

**Pattern: Subtract borders and padding**

```go
func (c *Card) contentArea() (x, y, width, height int) {
    borderWidth := 2  // Left + right borders
    borderHeight := 2 // Top + bottom borders

    x = c.x + 1
    y = c.y + 1
    width = c.width - borderWidth
    height = c.height - borderHeight

    return
}
```

### Handling Window Resize

**DO**: Update all dimensions on resize

```go
func (d *Dashboard) Update(msg tea.Msg) (Component, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // 1. Update viewport dimensions
        d.width = msg.Width
        d.height = msg.Height

        // 2. Recalculate card dimensions
        d.updateCardDimensions()

        // 3. Propagate to children
        d.modal.Update(msg)
    }

    return d, nil
}
```

**DON'T**: Cache layout calculations across resizes

```go
// Bad: Stale layout after resize
type Dashboard struct {
    cardWidth  int // Cached, never updated
    cardHeight int // Will be wrong after resize!
}
```

### Grid-Aware Navigation

**Pattern: Column-based movement**

```go
func (d *Dashboard) moveFocusDown() {
    cols := d.getColumnCount()
    newIndex := d.focusedCardIndex + cols

    if newIndex < len(d.cards) {
        d.setFocusedCard(newIndex)
    }
}

func (d *Dashboard) moveFocusUp() {
    cols := d.getColumnCount()
    newIndex := d.focusedCardIndex - cols

    if newIndex >= 0 {
        d.setFocusedCard(newIndex)
    }
}
```

---

## Common Pitfalls

### ANSI Code Width Calculation

**PITFALL**: Counting ANSI escape codes as visible characters

```go
// Bad: Includes ANSI codes in length
text := "\033[32mGreen\033[0m"
length := len(text) // 18 (wrong!)

// Good: Strip ANSI codes first
func visibleLength(s string) int {
    // Remove ANSI escape sequences
    ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
    stripped := ansiRegex.ReplaceAllString(s, "")
    return len(stripped) // 5 (correct!)
}
```

**Impact**: Misaligned text, broken borders, incorrect padding

**Solution**: Use ANSI-aware width calculation

```go
func (m *DetailModal) centerText(text string, width int) string {
    visibleLen := m.visibleLength(text)
    padding := (width - visibleLen) / 2

    if padding < 0 {
        padding = 0
    }

    return strings.Repeat(" ", padding) + text
}
```

### Window Resize Race Conditions

**PITFALL**: Rendering before dimensions are set

```go
// Bad: Renders with zero width
func (s *StatusBar) View() string {
    // s.width is still 0!
    return s.renderContent(s.width)
}
```

**Solution**: Guard against uninitialized dimensions

```go
func (s *StatusBar) View() string {
    if s.width == 0 {
        return ""
    }

    return s.renderContent(s.width)
}
```

### Focus State Bugs

**PITFALL**: Multiple components focused simultaneously

```go
// Bad: Both focused at once
dashboard.Focus()
modal.Show() // Also focuses, but dashboard still focused!
```

**Solution**: Always blur before transferring focus

```go
// Good: Explicit focus transfer
dashboard.Blur()
modal.Show()
```

### Message Handling Order

**PITFALL**: Processing messages in wrong order

```go
// Bad: Modal input leaked to parent
func (d *Dashboard) Update(msg tea.Msg) (Component, tea.Cmd) {
    if d.focused {
        d.handleKeyboard(msg) // Runs even when modal open!
    }

    if d.modal.IsVisible() {
        d.modal.Update(msg)
    }
}
```

**Solution**: Check modal FIRST

```go
// Good: Modal intercepts input
func (d *Dashboard) Update(msg tea.Msg) (Component, tea.Cmd) {
    if d.modal.IsVisible() {
        d.modal.Update(msg)
        return d, nil // Stop processing
    }

    if d.focused {
        d.handleKeyboard(msg)
    }
}
```

### String Concatenation Performance

**PITFALL**: Using `+` in loops

```go
// Bad: Quadratic time complexity
var output string
for i := 0; i < 1000; i++ {
    output += line + "\n" // Allocates new string each time
}
```

**Solution**: Use `strings.Builder`

```go
// Good: Linear time complexity
var builder strings.Builder
for i := 0; i < 1000; i++ {
    builder.WriteString(line)
    builder.WriteString("\n")
}
output := builder.String()
```

### Truncation Edge Cases

**PITFALL**: Negative slice bounds on narrow widths

```go
// Bad: Panics when width < 10
func truncate(s string, width int) string {
    return s[:width-3] + "..." // Panic if width = 5!
}
```

**Solution**: Guard against edge cases

```go
// Good: Safe truncation
func truncate(s string, width int) string {
    if len(s) <= width {
        return s
    }

    if width < 4 {
        return s[:width] // No room for "..."
    }

    return s[:width-3] + "..."
}
```

---

## Performance Optimization

### Efficient Rendering

**DO**: Minimize string allocations

```go
// Good: Pre-allocate builder
var builder strings.Builder
builder.Grow(estimatedSize) // Avoid resizing
builder.WriteString(content)
```

**DON'T**: Rebuild entire view on every update

```go
// Bad: Expensive full rebuild
func (d *Dashboard) View() string {
    // Rebuilds all cards even if only one changed
    return d.renderAllCards()
}
```

### Caching Layout Calculations

**Pattern: Lazy evaluation with cache invalidation**

```go
type Dashboard struct {
    width, height int

    // Cached layout
    cardWidth  int
    cardHeight int
    layoutDirty bool
}

func (d *Dashboard) Update(msg tea.Msg) (Component, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        d.width = msg.Width
        d.height = msg.Height
        d.layoutDirty = true // Invalidate cache
    }

    return d, nil
}

func (d *Dashboard) View() string {
    if d.layoutDirty {
        d.calculateLayout()
        d.layoutDirty = false
    }

    return d.render()
}
```

### Avoid Redundant Calculations

**DO**: Calculate once, reuse

```go
func (d *Dashboard) render() string {
    cols := d.getColumnCount() // Calculate once

    for i, card := range d.cards {
        row := i / cols    // Use cached value
        col := i % cols
        // ...
    }
}
```

**DON'T**: Recalculate in loops

```go
// Bad: Calls getColumnCount() 100 times
for i, card := range d.cards {
    row := i / d.getColumnCount()
    col := i % d.getColumnCount()
}
```

### Minimize ANSI Code Usage

**Pattern: Group styling**

```go
// Good: Apply style once to entire block
func render(lines []string, color string) string {
    return color + strings.Join(lines, "\n") + "\033[0m"
}
```

```go
// Bad: Styling every line individually
func render(lines []string, color string) string {
    var output string
    for _, line := range lines {
        output += color + line + "\033[0m\n" // Redundant resets
    }
    return output
}
```

### Debounce Expensive Operations

**Pattern: Rate limiting updates**

```go
type Component struct {
    lastUpdate time.Time
    updateDebounce time.Duration
}

func (c *Component) Update(msg tea.Msg) (Component, tea.Cmd) {
    if time.Since(c.lastUpdate) < c.updateDebounce {
        return c, nil // Skip update
    }

    c.lastUpdate = time.Now()
    // ... expensive update logic
}
```

---

## Testing Strategies

### Component Isolation

**Pattern: Test components independently**

```go
func TestStatCardCreation(t *testing.T) {
    card := NewStatCard(
        WithTitle("CPU"),
        WithValue("42%"),
    )

    if card.title != "CPU" {
        t.Errorf("Expected title='CPU', got '%s'", card.title)
    }
}
```

**Avoid**: Testing through parent components

```go
// Bad: Testing StatCard via Dashboard
func TestStatCardTitle(t *testing.T) {
    dashboard := NewDashboard(
        WithCards(NewStatCard(WithTitle("CPU"))),
    )

    view := dashboard.View()
    if !strings.Contains(view, "CPU") {
        t.Error("Card title not found")
    }
}
```

### Focus State Testing

**Pattern: Verify state transitions**

```go
func TestFocusLifecycle(t *testing.T) {
    card := NewStatCard()

    // Initial state
    if card.Focused() {
        t.Error("Should not be focused initially")
    }

    // Focus transition
    card.Focus()
    if !card.Focused() {
        t.Error("Should be focused after Focus()")
    }

    // Blur transition
    card.Blur()
    if card.Focused() {
        t.Error("Should not be focused after Blur()")
    }
}
```

### View Rendering Assertions

**Pattern: Check visual output**

```go
func TestStatCardBorderFocused(t *testing.T) {
    card := NewStatCard()
    card.width = 30
    card.height = 8
    card.Focus()

    view := card.View()

    // Check for double-line borders
    if !strings.Contains(view, "╔") || !strings.Contains(view, "╗") {
        t.Error("Focused card should have double-line borders")
    }

    // Check for cyan color
    if !strings.Contains(view, "\033[36m") {
        t.Error("Focused card should be cyan")
    }
}
```

### Integration Testing

**Pattern: Test component interactions**

```go
func TestDashboardModalIntegration(t *testing.T) {
    card := NewStatCard(WithTitle("CPU"))
    dashboard := NewDashboard(WithCards(card))
    dashboard.Focus()

    // Set viewport
    dashboard.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

    // Open modal
    dashboard.Update(tea.KeyMsg{Type: tea.KeyEnter})

    // Verify state changes
    if !dashboard.detailModal.IsVisible() {
        t.Error("Modal should be visible")
    }

    if dashboard.focused {
        t.Error("Dashboard should lose focus")
    }

    // Close modal
    dashboard.detailModal.Update(tea.KeyMsg{Type: tea.KeyEsc})

    if dashboard.detailModal.IsVisible() {
        t.Error("Modal should be hidden")
    }
}
```

### Edge Case Testing

**Pattern: Test boundary conditions**

```go
func TestStatusBarNarrowWidth(t *testing.T) {
    statusBar := NewStatusBar()
    statusBar.width = 20 // Very narrow

    longMessage := "This is a very long message"
    statusBar.SetMessage(longMessage)

    view := statusBar.View()

    // Should truncate gracefully
    if !strings.Contains(view, "...") {
        t.Error("Should truncate long message")
    }
}

func TestDashboardEmptyCards(t *testing.T) {
    dashboard := NewDashboard() // No cards
    dashboard.Focus()

    // Should not panic on navigation
    dashboard.Update(tea.KeyMsg{Type: tea.KeyRight})
    dashboard.Update(tea.KeyMsg{Type: tea.KeyEnter})
}
```

### Test Coverage Goals

**Aim for**:
- **80%+ code coverage** overall
- **90%+ component coverage** (files with tests)
- **100% public API coverage**

**Test categories**:
1. **Creation tests**: Verify initialization and defaults
2. **State management tests**: Focus, selection, visibility
3. **View rendering tests**: Visual output correctness
4. **Keyboard handling tests**: Input processing
5. **Integration tests**: Component interactions
6. **Edge case tests**: Boundary conditions

---

## Error Handling

### Graceful Degradation

**Pattern: Render empty on invalid state**

```go
func (c *Component) View() string {
    if c.width == 0 || c.height == 0 {
        return "" // Not initialized yet
    }

    if len(c.items) == 0 {
        return c.renderEmptyState()
    }

    return c.render()
}
```

### Input Validation

**DO**: Validate option parameters

```go
func WithColumns(cols int) DashboardOption {
    return func(d *Dashboard) {
        if cols < 1 {
            cols = 1 // Clamp to minimum
        }
        if cols > 12 {
            cols = 12 // Clamp to maximum
        }
        d.columns = cols
    }
}
```

**DON'T**: Panic on invalid input

```go
// Bad: Crashes on zero
func WithColumns(cols int) DashboardOption {
    return func(d *Dashboard) {
        d.columns = cols // May be 0 or negative!
    }
}
```

### Safe Array Access

**Pattern: Bounds checking**

```go
func (d *Dashboard) setFocusedCard(index int) {
    if index < 0 || index >= len(d.cards) {
        return // Ignore invalid index
    }

    // Blur previous
    if d.focusedCardIndex >= 0 && d.focusedCardIndex < len(d.cards) {
        d.cards[d.focusedCardIndex].Blur()
    }

    // Focus new
    d.focusedCardIndex = index
    d.cards[index].Focus()
}
```

---

## Keyboard Navigation

### Consistent Key Bindings

**Standard conventions**:
- **Arrow keys**: `←↑→↓` for navigation
- **Vim keys**: `hjkl` as alternatives
- **Enter**: Confirm/select/drill-down
- **ESC**: Cancel/close/go back
- **Tab**: Switch focus between components
- **q**: Quit (top level only)

**DO**: Support both arrow and vim keys

```go
switch msg.String() {
case "up", "k":
    d.moveFocusUp()
case "down", "j":
    d.moveFocusDown()
case "left", "h":
    d.moveFocusLeft()
case "right", "l":
    d.moveFocusRight()
}
```

### Keyboard Focus Scope

**Pattern: Only handle input when focused**

```go
func (c *Component) Update(msg tea.Msg) (Component, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if !c.focused {
            return c, nil // Ignore input
        }

        // Handle keys
    }

    return c, nil
}
```

### Navigation Boundaries

**DO**: Stop at edges, don't wrap

```go
func (d *Dashboard) moveFocusRight() {
    newIndex := d.focusedCardIndex + 1

    // Stop at last card
    if newIndex < len(d.cards) {
        d.setFocusedCard(newIndex)
    }
}
```

**Alternative**: Wrap around for infinite scrolling

```go
func (f *FileExplorer) moveFocusDown() {
    newIndex := (f.selectedIndex + 1) % len(f.items)
    f.setSelected(newIndex)
}
```

---

## Visual Design

### Color Usage

**Semantic colors**:
- **Cyan** (`\033[36m`): Focus indicators
- **Yellow** (`\033[33m`): Selection, warnings
- **Green** (`\033[32m`): Positive changes, success
- **Red** (`\033[31m`): Negative changes, errors
- **Dim** (`\033[2m`): Secondary text, unfocused

**DO**: Use ANSI color codes consistently

```go
const (
    ColorReset  = "\033[0m"
    ColorCyan   = "\033[36m"
    ColorYellow = "\033[33m"
    ColorGreen  = "\033[32m"
    ColorRed    = "\033[31m"
    ColorDim    = "\033[2m"
)
```

### Unicode Box Drawing

**Standard border sets**:

**Normal** (thin):
```
┌─┬─┐
│ │ │
├─┼─┤
│ │ │
└─┴─┘
```

**Focused** (double):
```
╔═╦═╗
║ ║ ║
╠═╬═╣
║ ║ ║
╚═╩═╝
```

**Selected** (thick):
```
┏━┳━┓
┃ ┃ ┃
┣━╋━┫
┃ ┃ ┃
┗━┻━┛
```

### Spacing and Padding

**Consistent spacing**:
- **Cards**: 2-4 char gap between cards
- **Borders**: 1 char padding inside borders
- **Sections**: 1 blank line between sections
- **Headers**: 2 blank lines before major sections

```go
func (c *Card) render() string {
    // Top border
    output := "┌" + strings.Repeat("─", c.width-2) + "┐\n"

    // Content with padding
    output += "│ " + c.title + strings.Repeat(" ", c.width-len(c.title)-3) + "│\n"

    // Bottom border
    output += "└" + strings.Repeat("─", c.width-2) + "┘\n"

    return output
}
```

### Text Alignment

**Pattern: Center, left-align, right-align helpers**

```go
func centerText(text string, width int) string {
    padding := (width - len(text)) / 2
    return strings.Repeat(" ", padding) + text
}

func rightAlign(text string, width int) string {
    padding := width - len(text)
    if padding < 0 {
        padding = 0
    }
    return strings.Repeat(" ", padding) + text
}
```

---

## Summary Checklist

### Component Design
- [ ] Single responsibility per component
- [ ] Implement full Component interface
- [ ] Use functional options pattern
- [ ] Validate all inputs

### Focus Management
- [ ] Only one component focused at a time
- [ ] Explicit focus transfer (blur → focus)
- [ ] Visual focus indicators (3 states)
- [ ] Modal focus lifecycle

### Layout
- [ ] Responsive grid calculations
- [ ] Handle window resize
- [ ] ANSI-aware width calculations
- [ ] Safe content area calculation

### Performance
- [ ] Use strings.Builder for concatenation
- [ ] Cache layout calculations
- [ ] Minimize ANSI code usage
- [ ] Avoid redundant calculations

### Testing
- [ ] 80%+ code coverage
- [ ] Test component isolation
- [ ] Test focus lifecycle
- [ ] Test view rendering
- [ ] Test edge cases
- [ ] Integration tests

### Error Handling
- [ ] Graceful degradation
- [ ] Input validation
- [ ] Bounds checking
- [ ] Safe array access

### Keyboard Navigation
- [ ] Support arrow + vim keys
- [ ] Only handle input when focused
- [ ] Consistent key bindings
- [ ] Navigation boundaries

### Visual Design
- [ ] Semantic color usage
- [ ] Consistent border styles
- [ ] Proper spacing/padding
- [ ] ANSI code cleanup

---

## Additional Resources

- [API Reference](API_REFERENCE.md) - Complete API documentation
- [Dashboard Guide](DASHBOARD.md) - Interactive dashboard system
- [Card System](CARD_SYSTEM.md) - Card-based layouts
- [Bubble Tea Docs](https://github.com/charmbracelet/bubbletea) - Framework documentation

---

**Last Updated**: 2024-01-11
**TUI Version**: v1.0.0
