# tui

A comprehensive Terminal User Interface framework for building Claude Code-like CLI experiences with modern, interactive dashboard capabilities.

## Overview

`tui` is a production-ready framework built on top of the SCKelemen visualization stack, providing fully-tested, ready-to-use components for building sophisticated terminal applications with modern UX patterns.

## ✨ Highlights

- **🎨 Interactive Dashboards**: Responsive grid layouts with real-time metrics, sparklines, and drill-down modals
- **⌨️  Keyboard Navigation**: Full arrow key + vim-style (hjkl) navigation with visual focus indicators
- **📊 Data Visualization**: StatCards with change indicators (↑↓→), trend graphs, and detailed modal views
- **🎯 Focus Management**: Intuitive focus flow with visual states (focused, selected, normal)
- **📏 Responsive Layouts**: Auto-adjusting grids that adapt to terminal size
- **🧪 Battle-Tested**: 354 tests with 82.9% coverage

## Features

- **Rich Components**: Dashboards, file explorers, command palettes, status bars, modals
- **Keyboard Navigation**: Arrow keys + vim bindings (hjkl) with customizable keymaps
- **Mouse Support**: Click, scroll, drag interactions
- **Focus Management**: Visual focus indicators with three states (focused/selected/normal)
- **Layout System**: CSS Grid and Flexbox layouts via `layout` package
- **Theme Support**: Full design token integration via `design-system`
- **Unicode Aware**: Proper handling of emoji, wide characters via `text`
- **Color Science**: Perceptually uniform gradients via `color` (OKLCH)

## Architecture

```
tui (high-level components)
 ├── cli (terminal rendering)
 ├── layout (flexbox/grid)
 ├── design-system (themes)
 ├── text (unicode width)
 └── color (OKLCH gradients)
```

## Installation

```bash
go get github.com/SCKelemen/tui@latest
```

## Quick Start

### Interactive Dashboard Example

```go
package main

import (
    "github.com/SCKelemen/tui"
    tea "github.com/charmbracelet/bubbletea"
)

type model struct {
    dashboard *tui.Dashboard
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.dashboard.Update(msg)
    case tea.KeyMsg:
        if msg.String() == "q" {
            return m, tea.Quit
        }
        m.dashboard.Update(msg)
    }
    return m, nil
}

func (m model) View() string {
    return m.dashboard.View()
}

func main() {
    // Create stat cards
    cpuCard := tui.NewStatCard(
        tui.WithTitle("CPU Usage"),
        tui.WithValue("42%"),
        tui.WithChange(5, 13.5),
        tui.WithTrend([]float64{35, 38, 40, 42, 45}),
    )

    memoryCard := tui.NewStatCard(
        tui.WithTitle("Memory"),
        tui.WithValue("8.2 GB"),
        tui.WithSubtitle("of 16 GB total"),
        tui.WithChange(-200, -2.4),
    )

    // Create responsive dashboard
    dashboard := tui.NewDashboard(
        tui.WithDashboardTitle("System Metrics"),
        tui.WithResponsiveLayout(30),
        tui.WithCards(cpuCard, memoryCard),
    )
    dashboard.Focus() // Enable keyboard navigation

    p := tea.NewProgram(
        model{dashboard: dashboard},
        tea.WithAltScreen(),
    )
    if _, err := p.Run(); err != nil {
        panic(err)
    }
}
```

**Features shown:**
- ✅ Responsive grid layout
- ✅ StatCards with values and trends
- ✅ Change indicators (↑↓→)
- ✅ Keyboard navigation (arrow keys/hjkl)
- ✅ Focus management

See [examples/dashboard_demo](examples/dashboard_demo/) for a complete example with real-time updates!

## Components

### 📊 Dashboard System (Interactive)

**Dashboard** - Responsive grid container for metrics visualization
- Keyboard navigation (←→↑↓ or hjkl)
- Auto-adjusting column layout
- Real-time updates
- Focus management

**StatCard** - Individual metric cards with:
- Title, value, subtitle
- Change indicators (↑↓→) with color coding
- Sparkline trends (▁▂▃▄▅▆▇█)
- Visual focus states (focused/selected/normal)

**DetailModal** - Drill-down view for detailed metrics
- Large 8-line trend graphs
- Statistics (min, max, avg)
- Press Enter to open, ESC to close

See [DASHBOARD.md](DASHBOARD.md) for complete documentation.

### FileExplorer
Tree view with navigation, search, and file operations.

### CommandPalette
Fuzzy-searchable command launcher with keyboard shortcuts.

### StatusBar
Bottom status bar with context-aware keybindings and focus indicators.

### Modal
Dialog boxes for confirmations, inputs, and detail views.

### ActivityBar
Animated status line with spinner, elapsed time, and progress.

### ToolBlock
Collapsible content blocks for tool execution results with streaming support.

### TextInput
Text input fields with validation and keyboard controls.

### Header
Top bar with title, breadcrumbs, and navigation.

### StructuredData
Formatted JSON/data display with syntax highlighting.

See [COMPONENTS.md](COMPONENTS.md) for detailed documentation on all components.

## Status & Roadmap

### ✅ Completed
- [x] Interactive Dashboard system with keyboard navigation
- [x] StatCard component with sparklines and change indicators
- [x] DetailModal for drill-down views
- [x] Focus management with visual states
- [x] FileExplorer component
- [x] StatusBar component
- [x] CommandPalette component
- [x] Keyboard navigation (arrow keys + vim bindings)
- [x] ActivityBar with spinner animations
- [x] ToolBlock with streaming support
- [x] Modal dialogs
- [x] TextInput fields
- [x] Header component
- [x] StructuredData display
- [x] Comprehensive test coverage (354 tests, 82.9%)

### 🚧 In Progress
- [ ] Theme customization and dark/light modes
- [ ] Mouse event handling improvements
- [ ] Additional chart types (bar, pie, gauge)

### 📋 Planned
- [ ] Window splitting and panes
- [ ] Plugin system
- [ ] Animation system
- [ ] More visualization components

## Testing

The library has comprehensive test coverage:

- **354 total tests** across all components
- **82.9% code coverage**
- **93.75% component coverage** (15 of 16 files)
- Tests for: creation, rendering, updates, focus management, keyboard navigation, edge cases

Run tests:
```bash
go test ./...                    # Run all tests
go test -v -run TestDashboard   # Dashboard tests
go test -v -run TestStatCard    # StatCard tests
go test -v -run TestDetailModal # Modal tests
go test -cover ./...            # With coverage report
```

## License

Bearware 1.0

## Related Projects

- [cli](https://github.com/SCKelemen/cli) - Low-level terminal rendering
- [layout](https://github.com/SCKelemen/layout) - CSS-like layout engine
- [design-system](https://github.com/SCKelemen/design-system) - Design tokens and themes
- [dataviz](https://github.com/SCKelemen/dataviz) - Data visualization components
