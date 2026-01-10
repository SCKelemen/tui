# Dashboard Demo

This demo showcases the **Dashboard** component with **StatCard** widgets, demonstrating responsive CSS Grid layout, viewport-based units, and real-time metric updates.

## Features

### Dashboard Component

The Dashboard component provides:
- **CSS Grid Layout** - Automatic card positioning with configurable columns
- **Responsive Sizing** - Cards automatically reflow based on viewport width
- **Configurable Gap** - Adjustable spacing between cards
- **Dynamic Updates** - Cards update in real-time

### StatCard Component

Each StatCard displays:
- **Title** - Metric name
- **Value** - Current metric value (bold, prominent)
- **Change Indicator** - Shows increase (↑ green), decrease (↓ red), or no change (→ white)
- **Percentage Change** - Change as a percentage
- **Subtitle** - Additional context or description
- **Sparkline** - ASCII trend chart using block characters (▁▂▃▄▅▆▇█)

### Metrics Displayed

The demo shows 9 different system metrics:

1. **CPU Usage** - Processor utilization with trend
2. **Memory** - RAM usage with total capacity
3. **Network** - Download/upload speed
4. **Disk I/O** - Read/write throughput
5. **Active Users** - Current online users
6. **Requests** - Request rate per minute
7. **Error Rate** - Percentage of failed requests
8. **Avg Latency** - Response time with p95 percentile
9. **Uptime** - System availability percentage

## Layout System Integration

The Dashboard uses layout helpers for CSS Grid:

```go
// Responsive grid - automatically adjusts columns
dashboard := tui.NewDashboard(
    tui.WithResponsiveLayout(30), // Min 30 characters per card
    tui.WithGap(2),                // 2 character gap
)

// Fixed column grid
dashboard := tui.NewDashboard(
    tui.WithGridColumns(3),  // Always 3 columns
    tui.WithGap(2),
)
```

### Grid Behavior

**Responsive Mode** (default):
- Calculates columns based on viewport width and minimum card width
- Cards reflow automatically on terminal resize
- Uses `LayoutHelpers.ResponsiveGridLayout(minCardWidth, gap)`

**Fixed Columns Mode**:
- Always displays specified number of columns
- Cards scale to fit available width
- Uses `LayoutHelpers.GridLayout(columns, gap)`

### Card Dimensions

Cards automatically calculate dimensions based on:
- Available viewport width
- Number of columns
- Gap spacing
- Number of rows needed

```go
cardWidth = (viewportWidth - gapTotal) / columns
cardHeight = (viewportHeight - titleHeight - gapTotal) / rows
```

## Usage

```bash
cd examples/dashboard_demo
go run main.go
```

### Controls

- **q** or **Ctrl+C** - Quit
- **r** - Refresh (regenerate all metrics)

### Auto-Update

Metrics update automatically every second, simulating real-time monitoring. In a production app, you would:
1. Fetch metrics from a data source (Prometheus, StatsD, etc.)
2. Update StatCard values on each tick
3. Optionally add animations for value changes

## Code Structure

### Creating StatCards

```go
cpuCard := tui.NewStatCard(
    tui.WithTitle("CPU Usage"),
    tui.WithValue("42%"),
    tui.WithSubtitle("8 cores active"),
    tui.WithChange(5, 13.5),          // +5, +13.5%
    tui.WithColor("#2196F3"),         // Accent color
    tui.WithTrendColor("#4CAF50"),    // Trend line color
    tui.WithTrend([]float64{...}),    // Sparkline data
)
```

### Creating Dashboard

```go
dashboard := tui.NewDashboard(
    tui.WithDashboardTitle("System Metrics Dashboard"),
    tui.WithResponsiveLayout(30),  // Min card width
    tui.WithGap(2),                 // Gap between cards
    tui.WithCards(
        cpuCard,
        memoryCard,
        networkCard,
        // ... more cards
    ),
)
```

### Updating Metrics

```go
// Get current cards
cards := dashboard.GetCards()

// Update a card
cards[0] = tui.NewStatCard(
    tui.WithTitle("CPU Usage"),
    tui.WithValue(fmt.Sprintf("%.0f%%", newValue)),
    // ... other options
)

// Set updated cards
dashboard.SetCards(cards)
```

## Sparkline Rendering

Sparklines use Unicode block characters to show trends:
- `▁` - Minimum value (12.5% height)
- `▂` - 25% height
- `▃` - 37.5% height
- `▄` - 50% height
- `▅` - 62.5% height
- `▆` - 75% height
- `▇` - 87.5% height
- `█` - Maximum value (100% height)

The sparkline:
1. Normalizes data to 0-1 range
2. Maps to 8 block characters (0-7 index)
3. Renders with trend color
4. Samples data if there are more points than available width

## Change Indicators

Changes show direction and magnitude:
- `↑ +10 (+5.0%)` - **Green** for increases
- `↓ -5 (-2.5%)` - **Red** for decreases
- `→ 0 (0.0%)` - **White** for no change

## Layout Helpers Used

This demo showcases:
- `GridLayout(columns, gap)` - Fixed column CSS Grid
- `ResponsiveGridLayout(minCardWidth, gap)` - Auto-adjusting grid
- `CardLayout(padding)` - Card containers with padding (future)

## Future Enhancements

Potential additions:
1. **Card Types** - Different card layouts (gauge, progress bar, list)
2. **Interactive Cards** - Click to drill down into metrics
3. **Chart Types** - Bar charts, pie charts, line graphs
4. **Themes** - Dark mode, light mode, custom color schemes
5. **Export** - Save dashboard as JSON or image
6. **Filtering** - Show/hide specific metrics
7. **Sorting** - Reorder cards by value, change, or name
8. **Alerts** - Highlight cards that exceed thresholds

## Real-World Use Cases

This dashboard pattern is perfect for:
- **System Monitoring** - CPU, memory, disk, network
- **Application Metrics** - Request rate, latency, errors
- **Business KPIs** - Sales, users, revenue, growth
- **CI/CD Pipelines** - Build status, test results, deployments
- **DevOps Dashboards** - Service health, uptime, logs

## Performance

The dashboard efficiently renders:
- **9 cards** with sparklines at **60 FPS**
- Responsive to terminal resize
- Minimal CPU usage when idle
- Efficient string concatenation

## Comparison to Web Dashboards

This TUI dashboard provides similar functionality to web dashboards like:
- **Grafana** - Metrics visualization
- **Datadog** - APM dashboards
- **New Relic** - Application monitoring
- **Kibana** - Log visualization

But runs entirely in the terminal with:
- ✅ No browser required
- ✅ SSH-friendly
- ✅ Low bandwidth
- ✅ Fast startup
- ✅ Keyboard-driven

## References

- [Layout System Integration](../../LAYOUT_INTEGRATION.md)
- [Layout Helpers](../../layout_helpers.go)
- [StatCard Component](../../statcard.go)
- [Dashboard Component](../../dashboard.go)
- [CSS Grid Guide](https://css-tricks.com/snippets/css/complete-guide-grid/)
