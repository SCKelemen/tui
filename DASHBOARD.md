# Dashboard System

The Dashboard system provides a powerful, responsive grid layout for displaying metrics and data cards in terminal UIs. Built on top of the [layout system](LAYOUT_INTEGRATION.md), it offers automatic card positioning, responsive resizing, and real-time updates.

## Overview

The Dashboard system is **fully interactive** with keyboard navigation, focus management, and drill-down capabilities. It consists of four main components:

1. **Dashboard** (`dashboard.go`) - Interactive grid container with keyboard navigation
2. **StatCard** (`statcard.go`) - Individual metric cards with visual focus states
3. **DetailModal** (`detailmodal.go`) - Drill-down view for detailed metrics
4. **Layout Helpers** (`layout_helpers.go`) - Reusable layout patterns

### Key Features
- ⌨️  **Keyboard Navigation**: Arrow keys (←→↑↓) and vim bindings (hjkl)
- 🎯 **Focus Management**: Visual states (focused, selected, normal)
- 📊 **Drill-Down**: Press Enter on any card to view detailed modal
- 📈 **Large Trend Graphs**: 8-line graphs in modal view
- 📉 **Statistics**: Min, max, avg calculations
- 🔄 **Real-Time Updates**: Live metric updates with smooth transitions

## Components

### StatCard

A StatCard displays a single metric with:
- **Title** - Metric name (e.g., "CPU Usage")
- **Value** - Current value (bold, prominent)
- **Change Indicator** - ↑ (green), ↓ (red), or → (white) with percentage
- **Subtitle** - Additional context
- **Sparkline** - ASCII trend chart using block characters (▁▂▃▄▅▆▇█)

#### Creating StatCards

```go
cpuCard := tui.NewStatCard(
    tui.WithTitle("CPU Usage"),
    tui.WithValue("42%"),
    tui.WithSubtitle("8 cores active"),
    tui.WithChange(5, 13.5),              // +5, +13.5%
    tui.WithColor("#2196F3"),             // Accent color
    tui.WithTrendColor("#4CAF50"),        // Sparkline color
    tui.WithTrend([]float64{35, 38, 40, 42, 45}),
)
```

#### StatCard Options

| Option | Description | Example |
|--------|-------------|---------|
| `WithTitle(string)` | Set card title | `WithTitle("Memory")` |
| `WithValue(string)` | Set main value | `WithValue("8.2 GB")` |
| `WithSubtitle(string)` | Set subtitle | `WithSubtitle("of 16 GB total")` |
| `WithChange(int, float64)` | Set change value and percentage | `WithChange(100, 5.5)` |
| `WithTrend([]float64)` | Set sparkline data | `WithTrend(data)` |
| `WithColor(string)` | Set accent color | `WithColor("#FF5722")` |
| `WithTrendColor(string)` | Set sparkline color | `WithTrendColor("#4CAF50")` |

#### Change Indicators

Changes show direction and magnitude:
- `↑ +10 (+5.0%)` - **Green** for increases
- `↓ -5 (-2.5%)` - **Red** for decreases
- `→ 0 (0.0%)` - **White** for no change (only shown if non-zero)

#### Sparklines

Sparklines use Unicode block characters to visualize trends:

```
▁▂▃▄▅▆▇█
```

- Automatically normalizes data to 0-1 range
- Maps values to 8 block characters
- Samples data if more points than available width
- Renders with configurable trend color

#### Visual Focus States

StatCards have three visual states indicated by different border styles:

**Normal State** (thin borders `┌─┐ │ └─┘`):
```
┌──────────────────────────┐
│ CPU Usage                │
│ 42%                      │
│ ↑ 5 (13.5%)             │
│ ▁▂▃▄▅▆▇█                │
└──────────────────────────┘
```

**Focused State** (double-line cyan borders `╔═╗ ║ ╚═╝`):
```
╔══════════════════════════╗
║ CPU Usage                ║
║ 42%                      ║
║ ↑ 5 (13.5%)             ║
║ ▁▂▃▄▅▆▇█                ║
╚══════════════════════════╝
```

**Selected State** (thick yellow borders `┏━┓ ┃ ┗━┛`):
```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ CPU Usage                ┃
┃ 42%                      ┃
┃ ↑ 5 (13.5%)             ┃
┃ ▁▂▃▄▅▆▇█                ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

Focus takes priority over selection in rendering.

### Dashboard

The Dashboard component arranges multiple StatCards in a responsive CSS Grid layout.

#### Creating Dashboards

```go
// Responsive layout (default)
dashboard := tui.NewDashboard(
    tui.WithDashboardTitle("System Metrics"),
    tui.WithResponsiveLayout(30),  // Min 30 chars per card
    tui.WithGap(2),                 // 2 character gap
    tui.WithCards(
        cpuCard,
        memoryCard,
        networkCard,
    ),
)

// Fixed column layout
dashboard := tui.NewDashboard(
    tui.WithGridColumns(3),  // Always 3 columns
    tui.WithGap(2),
    tui.WithCards(cards...),
)
```

#### Dashboard Options

| Option | Description | Example |
|--------|-------------|---------|
| `WithDashboardTitle(string)` | Set dashboard title | `WithDashboardTitle("Metrics")` |
| `WithGridColumns(int)` | Fixed number of columns | `WithGridColumns(4)` |
| `WithGap(float64)` | Gap between cards (in characters) | `WithGap(2)` |
| `WithResponsiveLayout(float64)` | Enable responsive mode with min card width | `WithResponsiveLayout(30)` |
| `WithCards(...*StatCard)` | Set initial cards | `WithCards(card1, card2)` |

#### Dynamic Card Management

```go
// Add cards
dashboard.AddCard(newCard)

// Remove cards by index
dashboard.RemoveCard(2)

// Replace all cards
dashboard.SetCards([]*StatCard{card1, card2, card3})

// Get all cards
cards := dashboard.GetCards()
```

#### Responsive vs Fixed Layout

**Responsive Mode** (default):
- Automatically calculates columns based on viewport width
- Cards reflow on terminal resize
- Formula: `columns = viewportWidth / (minCardWidth + gap)`
- Perfect for dashboards that adapt to different terminal sizes

**Fixed Column Mode**:
- Always displays specified number of columns
- Cards scale to fit available width
- Perfect for consistent layouts regardless of terminal size

### Interactive Features

#### Keyboard Navigation

The Dashboard supports full keyboard navigation with both arrow keys and vim-style bindings.

**Navigation Keys**:
```
←/h - Move focus left
→/l - Move focus right
↑/k - Move focus up (grid-aware)
↓/j - Move focus down (grid-aware)
Enter - Open DetailModal for focused card
ESC - Close modal / clear selection
```

**Enabling Navigation**:
```go
dashboard := tui.NewDashboard(
    tui.WithCards(cards...),
)
dashboard.Focus() // Enable keyboard navigation
```

**Grid-Aware Navigation**:
- Up/Down navigation respects column layout
- With 3 columns: card 0 → (down) → card 3
- Stops at boundaries (no wrapping)
- Works with both responsive and fixed layouts

**Example**:
```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "q" {
            return m, tea.Quit
        }
        // Forward all other keys to dashboard
        m.dashboard.Update(msg)
    }
    return m, nil
}
```

#### DetailModal

Press **Enter** on any focused card to open a detailed modal view with:

**Features**:
- **Large Trend Graph**: 8-line graph using Unicode blocks (▀▄█)
- **Statistics**: Min, max, average values
- **Change Indicator**: Detailed change information
- **Subtitle**: Additional context
- **Historical Data**: Optional history entries
- **Centered Display**: 70% width, 80% height
- **Close Controls**: ESC or 'q' to close

**Modal View Example**:
```
╔══════════════════════════════════════════════════════════╗
║ CPU Usage                           [ESC to close]        ║
╠══════════════════════════════════════════════════════════╣
║                                                            ║
║   42%                                                      ║
║                                                            ║
║   ↑ 5 (+13.5%)                                            ║
║                                                            ║
║   8 cores active                                          ║
║                                                            ║
║   Trend (Last 30 data points):                            ║
║                                                            ║
║     ██████▀▀▀                                             ║
║    ██    ██  ▄▄▄                                          ║
║   ██      ████  ██                                        ║
║  ██              ██▄▄                                     ║
║ ██                  ███▄                                  ║
║██                      ███                                ║
║                          ██▄                              ║
║                            ██                             ║
║                                                            ║
║   Min: 35.0  Max: 50.0  Avg: 42.5                        ║
║                                                            ║
╚══════════════════════════════════════════════════════════╝
```

**Automatic Content Population**:
```go
// When user presses Enter, the modal is automatically populated
// from the focused StatCard's data (title, value, trend, etc.)
```

**Modal Integration**:
The Dashboard handles modal lifecycle automatically:
1. User presses Enter → Dashboard calls `openDetailModal()`
2. Modal populated with card content via `SetContent(card)`
3. Modal shown with `Show()`, dashboard loses focus
4. User presses ESC → Modal hidden with `Hide()`, dashboard regains focus

**Focus Flow**:
```
Dashboard (focused)
  → User presses Enter
  → DetailModal (focused)
  → Dashboard (blurred)

DetailModal (focused)
  → User presses ESC
  → DetailModal (hidden)
  → Dashboard (focused)
```

### Layout Helpers

The `LayoutHelper` provides 15+ reusable layout patterns:

#### Grid Layouts

```go
// Equal-width columns (1fr each)
grid := tui.LayoutHelpers.GridLayout(3, 2)  // 3 columns, 2 char gap

// Responsive grid (auto-adjusting columns)
grid := tui.LayoutHelpers.ResponsiveGridLayout(30, 2)  // Min 30 chars per card
```

#### Column Layouts

```go
// Two columns with ratio
layout := tui.LayoutHelpers.TwoColumnLayout(1, 2)  // 1:2 ratio

// Three columns with ratios
layout := tui.LayoutHelpers.ThreeColumnLayout(1, 2, 1)  // 1:2:1 ratio

// Sidebar layout
layout := tui.LayoutHelpers.SidebarLayout(20)  // 20 char sidebar
```

#### Structural Layouts

```go
// Header/Content/Footer
layout := tui.LayoutHelpers.HeaderContentFooterLayout(3, 1)  // 3 header, 1 footer

// Centered overlay (modals, dialogs)
modal := tui.LayoutHelpers.CenteredOverlay(60, 20)  // 60x20 overlay

// Centered content
content := tui.LayoutHelpers.CenteredContent()
```

#### Stack Layouts

```go
// Vertical stack
stack := tui.LayoutHelpers.StackLayout(1)  // 1 char gap

// Horizontal stack
row := tui.LayoutHelpers.HorizontalStackLayout(2)  // 2 char gap

// Space-between row (toolbar style)
toolbar := tui.LayoutHelpers.SpaceBetweenRow()
```

#### Utility Helpers

```go
// Card container with padding
card := tui.LayoutHelpers.CardLayout(1)  // 1 char padding

// Absolute positioning
overlay := tui.LayoutHelpers.AbsolutePosition(10, 20, 100, 50)

// Flex-grow node
flexNode := tui.LayoutHelpers.FlexGrowNode(2)  // flex-grow: 2

// Fixed size node
fixedNode := tui.LayoutHelpers.FixedSizeNode(100, 50)
```

## Usage Patterns

### Basic Dashboard

```go
package main

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/SCKelemen/tui"
)

type model struct {
    dashboard *tui.Dashboard
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.dashboard.Update(msg)
    case tea.KeyMsg:
        if msg.String() == "q" {
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m model) View() string {
    return m.dashboard.View()
}

func main() {
    // Create cards
    cards := []*tui.StatCard{
        tui.NewStatCard(
            tui.WithTitle("CPU"),
            tui.WithValue("42%"),
            tui.WithChange(5, 13.5),
        ),
        tui.NewStatCard(
            tui.WithTitle("Memory"),
            tui.WithValue("8 GB"),
            tui.WithChange(-100, -1.2),
        ),
    }

    // Create dashboard
    dashboard := tui.NewDashboard(
        tui.WithDashboardTitle("System Metrics"),
        tui.WithCards(cards...),
    )

    p := tea.NewProgram(model{dashboard: dashboard}, tea.WithAltScreen())
    p.Run()
}
```

### Real-Time Updates

```go
type tickMsg time.Time

func tickCmd() tea.Cmd {
    return tea.Tick(time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tickMsg:
        // Update metrics
        m.updateMetrics()
        return m, tickCmd()
    }
    return m, nil
}

func (m *model) updateMetrics() {
    cards := m.dashboard.GetCards()

    // Update card values
    cards[0] = tui.NewStatCard(
        tui.WithTitle("CPU Usage"),
        tui.WithValue(fmt.Sprintf("%.0f%%", getCPUUsage())),
        tui.WithTrend(getCPUTrend()),
    )

    m.dashboard.SetCards(cards)
}
```

### Custom Grid Layout

```go
// Create a 2x3 grid of specific cards
dashboard := tui.NewDashboard(
    tui.WithGridColumns(3),  // 3 columns
    tui.WithGap(2),
    tui.WithCards(
        // Row 1
        cpuCard, memoryCard, diskCard,
        // Row 2
        networkCard, usersCard, requestsCard,
    ),
)
```

### Responsive Breakpoints

```go
func (m *model) updateLayout(width int) {
    var cards []*tui.StatCard

    if width >= 120 {
        // Wide layout: 4 columns
        m.dashboard = tui.NewDashboard(
            tui.WithGridColumns(4),
            tui.WithCards(allCards...),
        )
    } else if width >= 80 {
        // Medium layout: 3 columns
        m.dashboard = tui.NewDashboard(
            tui.WithGridColumns(3),
            tui.WithCards(importantCards...),
        )
    } else {
        // Narrow layout: 2 columns
        m.dashboard = tui.NewDashboard(
            tui.WithGridColumns(2),
            tui.WithCards(essentialCards...),
        )
    }
}
```

## Examples

### Dashboard Demo

See `examples/dashboard_demo/` for a complete working example with:
- 9 different metric cards (CPU, Memory, Network, etc.)
- Real-time updates every second
- Responsive grid layout
- Change indicators
- Sparklines

Run it:
```bash
cd examples/dashboard_demo
go run main.go
```

### Sparkline Generation

```go
// Generate realistic trend data
func generateTrendData(points int, baseValue float64, volatility float64) []float64 {
    trend := make([]float64, points)
    value := baseValue

    for i := 0; i < points; i++ {
        change := (rand.Float64() - 0.5) * volatility
        value += change

        if value < 0 {
            value = 0
        }

        trend[i] = value
    }

    return trend
}

// Use in card
card := tui.NewStatCard(
    tui.WithTitle("Requests"),
    tui.WithValue("45.2k"),
    tui.WithTrend(generateTrendData(30, 42000, 5000)),
)
```

## Testing

The Dashboard system has **comprehensive test coverage** with tests for all features:

**Dashboard Tests** (32 tests):
- Creation and configuration
- Card management (add, remove, set)
- Window size updates and responsive layout
- **Keyboard navigation** (up, down, left, right, vim keys)
- **Focus management** (initial focus, boundaries, grid-aware)
- **Selection and clearing**
- Column count calculation
- Rendering

**StatCard Tests** (33 tests):
- Creation with all options
- **Visual states** (normal, focused, selected, priority)
- **Border rendering** for each state
- Change indicators and sparklines
- Truncation and ANSI-aware lengths
- Window size handling
- Edge cases (narrow, empty, constant values)

**DetailModal Tests** (21 tests):
- Creation, show/hide, visibility
- Content population from StatCards
- **Keyboard handling** (ESC, 'q' to close)
- **Focus and state management**
- **Large trend graph rendering** (8-line graphs)
- Statistics calculation (min, max, avg)
- Historical data display
- Modal dimensions and centering
- **Integration with Dashboard**

**Layout Helpers Tests** (18 tests):
- All 15+ layout helper functions
- Grid, column, and stack layouts
- Centering and positioning

**Total**: **104 dashboard-related tests** out of 354 total tests

Run tests:
```bash
go test -v                       # All tests
go test -run TestDashboard       # Dashboard navigation tests
go test -run TestStatCard        # StatCard rendering & states
go test -run TestDetailModal     # Modal tests
go test -run TestLayoutHelper    # Layout helpers
go test -cover ./...            # With coverage report (82.9%)
```

## Performance

The Dashboard system is highly performant:

- **Rendering**: Handles 12+ cards at 60 FPS
- **Updates**: Efficient string concatenation
- **Responsive**: Instant reflow on terminal resize
- **Memory**: Minimal allocations

## Design Decisions

### Why String-Based Rendering?

The Dashboard currently uses string concatenation for rendering rather than full layout system integration because:

1. **Performance** - String operations are fast for terminal rendering
2. **Simplicity** - Easier to debug and maintain
3. **Compatibility** - Works with existing terminal libraries
4. **Future-Ready** - Skeleton code exists for layout integration

Future versions will migrate to full layout-based rendering for:
- More complex layouts
- Nested dashboards
- Interactive elements

### Why Sparklines?

Sparklines provide instant visual context for trends:
- Show patterns at a glance
- Minimal space usage
- No scrolling required
- Familiar from monitoring tools

### Why Responsive Mode?

Responsive mode ensures dashboards work across terminal sizes:
- Works on narrow terminals (80 cols)
- Takes advantage of wide terminals (200+ cols)
- Adapts to window resize
- Mobile-friendly when using mosh/ssh on mobile

## Real-World Use Cases

Perfect for:

### System Monitoring
- CPU, memory, disk, network metrics
- Process monitoring
- Resource usage dashboards

### Application Metrics
- Request rate, latency, errors
- Queue lengths, throughput
- Active connections, sessions

### Business KPIs
- Sales, revenue, growth
- User counts, engagement
- Conversion rates, churn

### DevOps Dashboards
- Service health, uptime
- Deployment status
- Build/test results

### CI/CD Pipelines
- Build duration, success rate
- Test coverage, failures
- Deployment frequency

## Comparison to Web Dashboards

Similar to:
- **Grafana** - Metrics visualization
- **Datadog** - APM dashboards
- **New Relic** - Application monitoring
- **Kibana** - Log visualization

But with advantages:
- ✅ No browser required
- ✅ SSH-friendly
- ✅ Low bandwidth
- ✅ Fast startup (<100ms)
- ✅ Keyboard-driven
- ✅ Works over slow connections

## Architecture

```
Dashboard (Grid Container)
  ├── StatCard (Metric 1)
  │   ├── Title
  │   ├── Value (bold)
  │   ├── Change Indicator (↑↓→)
  │   ├── Subtitle
  │   └── Sparkline (▁▂▃▄▅▆▇█)
  ├── StatCard (Metric 2)
  └── StatCard (Metric 3)

Layout System (Future)
  ├── CSS Grid (rows x columns)
  ├── Flexbox (for card internals)
  ├── Viewport Units (Px, Ch, Em, Vw, Vh)
  └── Design Tokens (colors, spacing)
```

## Future Enhancements

Planned features:

1. **Chart Types** - Bar charts, line graphs, gauges
2. **Interactive Cards** - Click to drill down
3. **Themes** - Dark mode, light mode, custom colors
4. **Export** - Save dashboard as JSON or image
5. **Filtering** - Show/hide specific metrics
6. **Sorting** - Reorder cards dynamically
7. **Alerts** - Highlight cards exceeding thresholds
8. **Animations** - Smooth transitions for value changes
9. **Full Layout Integration** - Use layout system for rendering
10. **Card Types** - Progress bars, tables, lists

## See Also

- [Layout System Integration](LAYOUT_INTEGRATION.md) - CSS Flexbox/Grid for Go
- [Layout Helpers](layout_helpers.go) - Reusable layout patterns
- [Dashboard Demo](examples/dashboard_demo/) - Complete working example
- [CSS Grid Guide](https://css-tricks.com/snippets/css/complete-guide-grid/)

## API Reference

### StatCard

```go
type StatCard struct {
    width      int
    height     int
    title      string
    value      string
    subtitle   string
    change     int
    changePct  float64
    trend      []float64
    color      string
    trendColor string
}

func NewStatCard(opts ...StatCardOption) *StatCard
func (s *StatCard) Init() tea.Cmd
func (s *StatCard) Update(msg tea.Msg) (Component, tea.Cmd)
func (s *StatCard) View() string
func (s *StatCard) Focus()
func (s *StatCard) Blur()
func (s *StatCard) Focused() bool
```

### Dashboard

```go
type Dashboard struct {
    width        int
    height       int
    columns      int
    gap          float64
    minCardWidth float64
    responsive   bool
    cards        []*StatCard
    title        string
}

func NewDashboard(opts ...DashboardOption) *Dashboard
func (d *Dashboard) Init() tea.Cmd
func (d *Dashboard) Update(msg tea.Msg) (Component, tea.Cmd)
func (d *Dashboard) View() string
func (d *Dashboard) Focus()
func (d *Dashboard) Blur()
func (d *Dashboard) Focused() bool
func (d *Dashboard) AddCard(card *StatCard)
func (d *Dashboard) RemoveCard(index int)
func (d *Dashboard) GetCards() []*StatCard
func (d *Dashboard) SetCards(cards []*StatCard)
```

### LayoutHelper

```go
type LayoutHelper struct{}

func NewLayoutHelper() *LayoutHelper
func (h *LayoutHelper) CenteredOverlay(width, height float64) *layout.Node
func (h *LayoutHelper) TwoColumnLayout(leftRatio, rightRatio float64) *layout.Node
func (h *LayoutHelper) ThreeColumnLayout(left, center, right float64) *layout.Node
func (h *LayoutHelper) SidebarLayout(sidebarWidth float64) *layout.Node
func (h *LayoutHelper) HeaderContentFooterLayout(headerHeight, footerHeight float64) *layout.Node
func (h *LayoutHelper) GridLayout(columns int, gap float64) *layout.Node
func (h *LayoutHelper) ResponsiveGridLayout(minCardWidth, gap float64) *layout.Node
func (h *LayoutHelper) CardLayout(paddingCh float64) *layout.Node
func (h *LayoutHelper) StackLayout(gap float64) *layout.Node
func (h *LayoutHelper) HorizontalStackLayout(gap float64) *layout.Node
func (h *LayoutHelper) SpaceBetweenRow() *layout.Node
func (h *LayoutHelper) CenteredContent() *layout.Node
func (h *LayoutHelper) AbsolutePosition(top, left, width, height float64) *layout.Node
func (h *LayoutHelper) FlexGrowNode(grow float64) *layout.Node
func (h *LayoutHelper) FixedSizeNode(width, height float64) *layout.Node

// Global instance
var LayoutHelpers = NewLayoutHelper()
```

## Conclusion

The Dashboard system provides a powerful, flexible way to build terminal dashboards with minimal code. The combination of StatCards, responsive grid layout, and layout helpers makes it easy to create professional-looking metrics displays that adapt to any terminal size.

Start with the [dashboard demo](examples/dashboard_demo/) to see it in action, then build your own dashboards using the patterns documented here.
