# Architecture Overview

This document explains the relationship between the three CLI repositories and their supporting foundation libraries.

## Repository Organization

The SCKelemen CLI ecosystem consists of three main repositories with distinct, non-overlapping concerns:

```
┌─────────────────────────────────────────────────────────────┐
│ clix - CLI Application Framework                            │
│ github.com/SCKelemen/clix                                   │
│                                                             │
│ Purpose: Building traditional CLI applications             │
│ Features: Command parsing, flags, config, prompts          │
│ Similar to: cobra, urfave/cli                              │
│                                                             │
│ Use when: Building command-line tools with subcommands     │
│           (e.g., git, docker, kubectl)                     │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ tui - Interactive Terminal UI Framework                     │
│ github.com/SCKelemen/tui                                    │
│                                                             │
│ Purpose: Building interactive terminal user interfaces     │
│ Features: Dashboard, FileExplorer, Modal, CodeBlock, etc.  │
│ Built on: Bubble Tea + cli renderer                        │
│                                                             │
│ Use when: Building interactive TUI applications            │
│           (e.g., htop, lazygit, k9s)                       │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   │ depends on
                   ↓
┌─────────────────────────────────────────────────────────────┐
│ cli - Terminal Rendering Engine                            │
│ github.com/SCKelemen/cli                                   │
│                                                             │
│ Purpose: Low-level terminal rendering with CSS layouts     │
│ Features: Screen buffer, ANSI codes, Flexbox/Grid layouts  │
│ Components: Heatmap, LineGraph, BarChart, StatCard, etc.   │
│                                                             │
│ Use when: Building custom rendering or static output       │
│           (usually consumed via tui)                        │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   │ depends on
                   ↓
┌─────────────────────────────────────────────────────────────┐
│ Foundation Libraries                                        │
│                                                             │
│ • layout - CSS Grid/Flexbox layout engine                  │
│ • design-system - Design tokens and themes                 │
│ • color - OKLCH color manipulation                         │
│ • text - Unicode-aware text measurement                    │
│ • unicode - Grapheme cluster handling                      │
│ • units - CSS unit parsing (em, rem, vh, vw, etc.)        │
└─────────────────────────────────────────────────────────────┘
```

## Key Distinctions

### clix vs tui

**clix** and **tui** serve completely different use cases and do not overlap:

| Aspect | clix | tui |
|--------|------|-----|
| **Purpose** | CLI command parsing | Interactive TUI components |
| **User Interaction** | Command invocation with flags | Real-time keyboard/mouse navigation |
| **Output Model** | Line-based, append-only | Full-screen, redrawable canvas |
| **Examples** | git, docker, kubectl | htop, lazygit, k9s |
| **Key Dependencies** | Standard library, minimal | Bubble Tea, cli renderer |

### cli vs tui

**cli** is the rendering foundation that **tui** builds upon:

| Layer | Responsibility | Example Types |
|-------|----------------|---------------|
| **tui** | Interactive components with state | Dashboard, FileExplorer, Modal |
| **cli** | Static rendering, layout calculation | Screen buffer, ANSI codes, CSS layouts |

**Relationship**:
- `tui` components use `cli/renderer` for final rendering
- `tui` adds Bubble Tea integration (Init/Update/View pattern)
- `tui` adds focus management, keyboard navigation, state management

## Dependency Graph

```
tui/
├── Dashboard (interactive grid of StatCards)
├── FileExplorer (tree navigation with focus)
├── Modal (dialog boxes)
├── CodeBlock (collapsible code display)
└── DiffBlock (unified diff display)
    ↓ uses
cli/
├── renderer/ (screen buffer, ANSI codes, styling)
├── components/ (static viz: heatmap, linegraph, barchart)
└── layout integration (CSS Grid/Flexbox)
    ↓ uses
foundation/
├── layout (layout engine)
├── design-system (design tokens)
├── color (OKLCH gradients)
└── text (unicode width)

clix/
├── Command parsing
├── Flag/config management
└── Interactive prompts (separate from tui)
    ↓ uses
foundation/ (minimal)
└── Standard library
```

## When to Use Each Repository

### Use **clix** when you need:

✅ Command-line tools with subcommands
✅ Flag and configuration management
✅ Interactive prompts for missing arguments
✅ Config file persistence (`~/.config/app/config.yaml`)
✅ Shell completion scripts

**Example use cases**:
- `myapp deploy --env production --region us-west-2`
- `myapp config set api.token abc123`
- `myapp users create --name "John Doe" --email john@example.com`

**Code example**:
```go
import "github.com/SCKelemen/clix"

app := clix.NewApp("myapp")
app.Root = clix.NewCommand("myapp")
app.Root.Children = []*clix.Command{
    deployCmd,
    configCmd,
    usersCmd,
}
app.Run(context.Background(), os.Args[1:])
```

### Use **tui** when you need:

✅ Interactive dashboards with real-time updates
✅ File explorers with keyboard navigation
✅ Modal dialogs and confirmations
✅ Full-screen applications with focus management
✅ Code/diff viewers with syntax highlighting

**Example use cases**:
- System monitoring dashboard (like htop)
- Git UI (like lazygit)
- File manager (like ranger)
- Log viewer with search/filter
- Development tool with multiple views

**Code example**:
```go
import (
    "github.com/SCKelemen/tui"
    tea "github.com/charmbracelet/bubbletea"
)

dashboard := tui.NewDashboard(
    tui.WithDashboardTitle("System Metrics"),
    tui.WithCards(cpuCard, memoryCard, diskCard),
)

p := tea.NewProgram(model{dashboard: dashboard}, tea.WithAltScreen())
p.Run()
```

### Use **cli** directly when you need:

✅ Custom rendering without Bubble Tea
✅ Static output generation (SVG-like layouts in terminal)
✅ Building your own TUI framework
✅ Non-interactive visualizations

**Example use cases**:
- Generating terminal "images" for documentation
- Custom layout engine experimentation
- Low-level screen manipulation
- Building a different TUI framework on top

**Code example**:
```go
import (
    "github.com/SCKelemen/cli/renderer"
    "github.com/SCKelemen/cli/components"
)

screen := renderer.NewScreen(80, 24)
heatmap := components.NewHeatmap(data)
styledNode := heatmap.ToStyledNode()
screen.Render(styledNode)
print(screen.String())
```

## Common Patterns

### Pattern 1: Interactive TUI with Command Parsing

You might want **both** clix and tui in the same application:

```go
// Use clix to parse the command
app := clix.NewApp("myapp")

watchCmd := clix.NewCommand("watch")
watchCmd.Run = func(ctx *clix.Context) error {
    // Launch tui dashboard
    dashboard := tui.NewDashboard(...)
    p := tea.NewProgram(model{dashboard: dashboard})
    return p.Start()
}

app.Root.Children = []*clix.Command{watchCmd}
app.Run(ctx, os.Args[1:])
```

**Example**: `myapp watch --interval 1s` parses flags with clix, then launches a tui dashboard.

### Pattern 2: TUI Component in Non-Interactive Context

You might want tui components for rendering but not Bubble Tea:

```go
// Create tui component but render statically using cli
codeBlock := tui.NewCodeBlock(
    tui.WithCodeOperation("Write"),
    tui.WithCodeFilename("main.go"),
    tui.WithCode(sourceCode),
)

// Extract the cli renderer output
// (tui components internally use cli/renderer)
output := codeBlock.View()
fmt.Print(output)
```

### Pattern 3: Custom Visualizations with cli

Build custom charts using the low-level renderer:

```go
import (
    "github.com/SCKelemen/cli/renderer"
    "github.com/SCKelemen/layout"
)

// Create layout tree
root := &layout.Node{
    Style: layout.Style{
        Display: layout.DisplayGrid,
        GridTemplateColumns: []layout.TrackSize{
            layout.Fr(1), layout.Fr(1),
        },
        Width: 80, Height: 24,
    },
}

// Add styled nodes
styledRoot := renderer.NewStyledNode(root, nil)
styledRoot.AddChild(leftPanel)
styledRoot.AddChild(rightPanel)

// Compute layout
layout.Layout(root, layout.Tight(80, 24))

// Render to screen
screen := renderer.NewScreen(80, 24)
screen.Render(styledRoot)
print(screen.String())
```

## Component Availability Matrix

### cli/components (Low-Level Rendering)

| Component | Purpose | Interactive | State Management |
|-----------|---------|-------------|------------------|
| Heatmap | Contribution-style heatmap | ❌ | ❌ |
| LineGraph | Time-series line charts | ❌ | ❌ |
| BarChart | Horizontal/vertical bars | ❌ | ❌ |
| StatCard | Metric card with trend | ❌ | ❌ |
| AreaChart | Filled area chart | ❌ | ❌ |
| ScatterPlot | Scatter plot | ❌ | ❌ |
| Collapsible | Expandable sections | ❌ | ❌ |
| Loading | Spinner, progress bar | ❌ | ❌ |
| Message | Styled message blocks | ❌ | ❌ |

### tui (High-Level Interactive Components)

| Component | Purpose | Interactive | State Management |
|-----------|---------|-------------|------------------|
| Dashboard | Grid of stat cards | ✅ (kbd nav) | ✅ (focus, selection) |
| StatCard | Metric card wrapper | ✅ (focus states) | ✅ (focus, selection) |
| FileExplorer | File tree navigation | ✅ (kbd nav) | ✅ (selection, expansion) |
| CommandPalette | Fuzzy command search | ✅ (search, select) | ✅ (input, filtering) |
| Modal | Dialog boxes | ✅ (focus trap) | ✅ (open/close) |
| DetailModal | Drill-down view | ✅ (ESC to close) | ✅ (open/close) |
| StatusBar | Context-aware status | ❌ | ✅ (content updates) |
| Header | Title and breadcrumbs | ❌ | ✅ (content updates) |
| ActivityBar | Animated status line | ❌ | ✅ (spinner, progress) |
| ToolBlock | Collapsible content | ✅ (expand/collapse) | ✅ (expanded state) |
| CodeBlock | Syntax-highlighted code | ✅ (expand/collapse) | ✅ (expanded state) |
| DiffBlock | Unified diff display | ✅ (expand/collapse) | ✅ (expanded state) |
| ConfirmationBlock | File operation prompts | ✅ (option selection) | ✅ (selection, confirmed) |
| TextInput | Text entry field | ✅ (typing, cursor) | ✅ (value, cursor pos) |
| StructuredData | JSON/data display | ✅ (expand/collapse) | ✅ (expanded nodes) |

## Testing Strategy

Each layer has distinct testing concerns:

### clix Tests
- Command parsing correctness
- Flag precedence (cmd flags > app flags > env > config > defaults)
- Argument validation and prompting
- Config file persistence

### tui Tests
- Component initialization
- Keyboard event handling
- Focus state transitions
- Bubble Tea message flow
- Visual state rendering

### cli Tests
- Layout calculations (CSS Grid/Flexbox)
- Screen buffer correctness
- ANSI escape code generation
- Unicode width calculations

## Migration Guide

### From Separate Components to tui

If you've been using cli components directly, migrate to tui wrappers:

**Before** (cli):
```go
import "github.com/SCKelemen/cli/components"

statcard := components.NewStatCard(data)
screen := renderer.NewScreen(80, 24)
screen.Render(statcard.ToStyledNode())
```

**After** (tui):
```go
import "github.com/SCKelemen/tui"

card := tui.NewStatCard(
    tui.WithTitle("CPU"),
    tui.WithValue("42%"),
    tui.WithChange(5, 13.5),
)
// Use in Bubble Tea program
p := tea.NewProgram(model{card: card})
```

### From Manual Rendering to Bubble Tea

If you've been manually rendering frames, use tui's Update/View pattern:

**Before** (manual):
```go
for {
    screen.Clear()
    screen.Render(root)
    fmt.Print(screen.String())
    time.Sleep(100 * time.Millisecond)
}
```

**After** (Bubble Tea):
```go
type model struct {
    dashboard *tui.Dashboard
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.dashboard.Update(msg)
    return m, nil
}

func (m model) View() string {
    return m.dashboard.View()
}
```

## Future Direction

### Planned Enhancements

**tui**:
- [ ] Window splitting and panes
- [ ] Animation system
- [ ] More chart wrappers (BarChart, LineGraph, Heatmap)
- [ ] Plugin system

**cli**:
- [ ] Additional layout modes
- [ ] More animation easing functions
- [ ] Performance optimizations

**clix**:
- [ ] Plugin system
- [ ] More extension types
- [ ] Advanced prompt widgets

### Stability Commitment

- **clix**: Stable API, following semver
- **tui**: Stable core, expanding components
- **cli**: Stable renderer, evolving layouts

## Contributing

Each repository has its own contribution guidelines:

- **clix**: See [clix/README.md](https://github.com/SCKelemen/clix)
- **tui**: See [tui/README.md](https://github.com/SCKelemen/tui)
- **cli**: See [cli/README.md](https://github.com/SCKelemen/cli)

## Questions?

For questions about which repository to use:

1. **Need command parsing?** → Use **clix**
2. **Need interactive UI?** → Use **tui**
3. **Need custom rendering?** → Use **cli**
4. **Need both CLI and TUI?** → Use **clix + tui**

See the README.md in each repository for detailed API documentation and examples.
