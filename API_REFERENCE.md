# TUI API Reference

Complete API documentation for the TUI library.

## Table of Contents

- [Core Interfaces](#core-interfaces)
- [Dashboard System](#dashboard-system)
  - [Dashboard](#dashboard)
  - [StatCard](#statcard)
  - [DetailModal](#detailmodal)
- [Interactive Components](#interactive-components)
  - [CommandPalette](#commandpalette)
  - [FileExplorer](#fileexplorer)
  - [StatusBar](#statusbar)
  - [Modal](#modal)
- [Display Components](#display-components)
  - [ActivityBar](#activitybar)
  - [ToolBlock](#toolblock)
  - [Header](#header)
  - [TextInput](#textinput)
  - [StructuredData](#structureddata)
- [Layout Helpers](#layout-helpers)
- [Common Types](#common-types)

---

## Core Interfaces

### Component

All TUI components implement the `Component` interface for consistent interaction with the Bubble Tea framework.

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

**Methods**:

- **`Init() tea.Cmd`**
  - Initialize the component
  - Returns initial command to execute (or nil)
  - Called once when component is created

- **`Update(msg tea.Msg) (Component, tea.Cmd)`**
  - Handle incoming messages
  - **Parameters**: `msg` - Message to process (KeyMsg, WindowSizeMsg, etc.)
  - **Returns**: Updated component and optional command
  - Core event handler for all interactions

- **`View() string`**
  - Render the component as a string
  - **Returns**: ANSI-formatted string for terminal display
  - Called on every frame

- **`Focus()`**
  - Set component as focused
  - Enables keyboard input handling
  - May update visual appearance

- **`Blur()`**
  - Remove focus from component
  - Disables keyboard input handling
  - May update visual appearance

- **`Focused() bool`**
  - **Returns**: true if component is currently focused

---

## Dashboard System

### Dashboard

Interactive grid container for displaying metric cards with keyboard navigation.

#### Type Definition

```go
type Dashboard struct {
    // Viewport dimensions
    width   int
    height  int

    // Focus state
    focused bool

    // Layout configuration
    columns      int     // Number of columns in grid
    gap          float64 // Gap between cards (characters)
    minCardWidth float64 // Minimum card width for responsive layout
    responsive   bool    // Use responsive grid layout

    // Cards
    cards []*StatCard

    // Navigation
    focusedCardIndex  int // Index of currently focused card (-1 = none)
    selectedCardIndex int // Index of selected card for drill-down (-1 = none)

    // Detail modal for drill-down
    detailModal *DetailModal

    // Title
    title string
}
```

#### Constructor

```go
func NewDashboard(opts ...DashboardOption) *Dashboard
```

Creates a new dashboard with optional configuration.

**Parameters**:
- `opts` - Variable number of configuration options

**Returns**: Configured `*Dashboard` instance

**Defaults**:
- `columns`: 3
- `gap`: 2
- `minCardWidth`: 30
- `responsive`: true
- `focusedCardIndex`: 0 (if cards exist)
- `selectedCardIndex`: -1

**Example**:
```go
dashboard := tui.NewDashboard(
    tui.WithDashboardTitle("Metrics"),
    tui.WithResponsiveLayout(30),
    tui.WithGap(2),
    tui.WithCards(card1, card2, card3),
)
```

#### Options

**`WithDashboardTitle(title string) DashboardOption`**

Set the dashboard title displayed at the top.

```go
dashboard := tui.NewDashboard(
    tui.WithDashboardTitle("System Metrics Dashboard"),
)
```

**`WithGridColumns(columns int) DashboardOption`**

Set fixed number of columns (disables responsive mode).

```go
dashboard := tui.NewDashboard(
    tui.WithGridColumns(4), // Always 4 columns
)
```

**`WithGap(gap float64) DashboardOption`**

Set gap between cards in characters.

```go
dashboard := tui.NewDashboard(
    tui.WithGap(3), // 3 character gap
)
```

**`WithResponsiveLayout(minCardWidth float64) DashboardOption`**

Enable responsive layout with minimum card width.

```go
dashboard := tui.NewDashboard(
    tui.WithResponsiveLayout(30), // Min 30 chars per card
)
```

**`WithCards(cards ...*StatCard) DashboardOption`**

Set initial stat cards.

```go
dashboard := tui.NewDashboard(
    tui.WithCards(cpuCard, memoryCard, networkCard),
)
```

#### Methods

**`Init() tea.Cmd`**

Initialize the dashboard (implements Component interface).

**Returns**: nil (no initial command)

---

**`Update(msg tea.Msg) (Component, tea.Cmd)`**

Handle messages including window resize and keyboard navigation.

**Handles**:
- `tea.WindowSizeMsg` - Updates dimensions and card layout
- `tea.KeyMsg` - Navigation and modal control (if focused)

**Keyboard Controls** (when focused):
- `←`, `h` - Move focus left
- `→`, `l` - Move focus right
- `↑`, `k` - Move focus up (grid-aware)
- `↓`, `j` - Move focus down (grid-aware)
- `Enter` - Open DetailModal for focused card
- `ESC` - Clear selection

**Returns**: Updated dashboard and optional command

**Example**:
```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.dashboard.Update(msg)
    case tea.KeyMsg:
        m.dashboard.Update(msg)
    }
    return m, nil
}
```

---

**`View() string`**

Render the dashboard with all cards and optional modal overlay.

**Returns**: ANSI-formatted string with:
- Optional title (if set)
- Grid of StatCards
- DetailModal overlay (if visible)

---

**`Focus()`**

Enable keyboard navigation for the dashboard.

**Example**:
```go
dashboard := tui.NewDashboard(tui.WithCards(cards...))
dashboard.Focus() // Enable keyboard navigation
```

---

**`Blur()`**

Disable keyboard navigation.

---

**`Focused() bool`**

**Returns**: true if dashboard has focus

---

**`AddCard(card *StatCard)`**

Add a card to the dashboard.

**Parameters**:
- `card` - StatCard to add

**Example**:
```go
newCard := tui.NewStatCard(tui.WithTitle("Disk"))
dashboard.AddCard(newCard)
```

---

**`RemoveCard(index int)`**

Remove a card by index.

**Parameters**:
- `index` - Card index to remove (0-based)

**Example**:
```go
dashboard.RemoveCard(2) // Remove third card
```

---

**`GetCards() []*StatCard`**

Get all cards.

**Returns**: Slice of all StatCard pointers

**Example**:
```go
cards := dashboard.GetCards()
for i, card := range cards {
    fmt.Printf("Card %d: %v\n", i, card)
}
```

---

**`SetCards(cards []*StatCard)`**

Replace all cards.

**Parameters**:
- `cards` - Slice of StatCard pointers

**Example**:
```go
newCards := []*tui.StatCard{card1, card2, card3}
dashboard.SetCards(newCards)
```

---

### StatCard

Individual metric card with title, value, change indicator, and sparkline.

#### Type Definition

```go
type StatCard struct {
    // Dimensions
    width    int
    height   int

    // State
    focused  bool
    selected bool

    // Content
    title      string
    value      string
    subtitle   string
    change     int       // Absolute change
    changePct  float64   // Percentage change
    trend      []float64 // Sparkline data
    color      string    // Accent color
    trendColor string    // Color for trend/sparkline
}
```

#### Constructor

```go
func NewStatCard(opts ...StatCardOption) *StatCard
```

Creates a new stat card with optional configuration.

**Parameters**:
- `opts` - Variable number of configuration options

**Returns**: Configured `*StatCard` instance

**Defaults**:
- `width`: 30
- `height`: 8
- `color`: "#2196F3" (blue)
- `trendColor`: "#4CAF50" (green)

#### Options

**`WithTitle(title string) StatCardOption`**

Set card title.

```go
card := tui.NewStatCard(
    tui.WithTitle("CPU Usage"),
)
```

---

**`WithValue(value string) StatCardOption`**

Set main value (displayed prominently).

```go
card := tui.NewStatCard(
    tui.WithValue("42%"),
)
```

---

**`WithSubtitle(subtitle string) StatCardOption`**

Set subtitle text.

```go
card := tui.NewStatCard(
    tui.WithSubtitle("8 cores active"),
)
```

---

**`WithChange(change int, changePct float64) StatCardOption`**

Set change value and percentage.

**Parameters**:
- `change` - Absolute change value
- `changePct` - Percentage change

**Display**:
- Positive: `↑ +5 (+13.5%)` in green
- Negative: `↓ -10 (-5.2%)` in red
- Zero: `→ 0 (0.0%)` in white (only if non-zero percentage)

```go
card := tui.NewStatCard(
    tui.WithChange(100, 5.5), // +100, +5.5%
)
```

---

**`WithTrend(trend []float64) StatCardOption`**

Set sparkline trend data.

**Parameters**:
- `trend` - Slice of numeric values

**Display**: Uses Unicode blocks `▁▂▃▄▅▆▇█`

```go
card := tui.NewStatCard(
    tui.WithTrend([]float64{10, 20, 15, 25, 30}),
)
```

---

**`WithColor(color string) StatCardOption`**

Set accent color (hex format).

```go
card := tui.NewStatCard(
    tui.WithColor("#FF5722"),
)
```

---

**`WithTrendColor(color string) StatCardOption`**

Set sparkline color (hex format).

```go
card := tui.NewStatCard(
    tui.WithTrendColor("#4CAF50"),
)
```

#### Methods

**`Init() tea.Cmd`**

Initialize the card.

**Returns**: nil

---

**`Update(msg tea.Msg) (Component, tea.Cmd)`**

Handle window size messages.

**Handles**:
- `tea.WindowSizeMsg` - Updates card dimensions

**Returns**: Updated card and nil

---

**`View() string`**

Render the card with appropriate border style based on state.

**Border Styles**:
- **Normal**: Thin single-line `┌─┐ │ └─┘`
- **Focused**: Double-line cyan `╔═╗ ║ ╚═╝`
- **Selected**: Thick yellow `┏━┓ ┃ ┗━┛`

**Returns**: ANSI-formatted string

---

**`Focus()`**

Set card as focused (changes border to double-line cyan).

---

**`Blur()`**

Remove focus from card (changes border to normal).

---

**`Focused() bool`**

**Returns**: true if card is focused

---

**`Select()`**

Mark card as selected (changes border to thick yellow).

---

**`Deselect()`**

Remove selection from card.

---

**`IsSelected() bool`**

**Returns**: true if card is selected

---

### DetailModal

Drill-down modal view for detailed metrics with large trend graphs.

#### Type Definition

```go
type DetailModal struct {
    // Dimensions
    width   int
    height  int

    // State
    visible bool
    focused bool

    // Content from StatCard
    title      string
    value      string
    subtitle   string
    change     int
    changePct  float64
    trend      []float64
    color      string
    trendColor string

    // Additional details
    history []string // Historical data points
}
```

#### Constructor

```go
func NewDetailModal(opts ...DetailModalOption) *DetailModal
```

Creates a new detail modal.

**Parameters**:
- `opts` - Variable number of configuration options

**Returns**: Configured `*DetailModal` instance

**Defaults**:
- `visible`: false
- `focused`: false
- `history`: empty slice

#### Options

**`WithModalContent(card *StatCard) DetailModalOption`**

Set content from a StatCard.

```go
modal := tui.NewDetailModal(
    tui.WithModalContent(cpuCard),
)
```

---

**`WithHistory(history []string) DetailModalOption`**

Set historical data points.

```go
modal := tui.NewDetailModal(
    tui.WithHistory([]string{
        "2024-01-10: 1,234 users",
        "2024-01-09: 1,189 users",
    }),
)
```

#### Methods

**`Init() tea.Cmd`**

Initialize the modal.

**Returns**: nil

---

**`Update(msg tea.Msg) (Component, tea.Cmd)`**

Handle window size and keyboard messages.

**Handles**:
- `tea.WindowSizeMsg` - Updates modal dimensions
- `tea.KeyMsg` - Close controls (if visible and focused)

**Keyboard Controls** (when visible and focused):
- `ESC` - Close modal
- `q` - Close modal

**Returns**: Updated modal and nil

---

**`View() string`**

Render the modal centered on screen.

**Layout**:
- Size: 70% width, 80% height of viewport
- Min: 60x20, Max: viewport-4
- Centered horizontally and vertically

**Content**:
- Title bar with close hint `[ESC to close]`
- Large value display (bold cyan)
- Change indicator with arrow (↑↓→)
- Subtitle (gray)
- 8-line trend graph using `▀▄█` characters
- Statistics (Min, Max, Avg)
- Historical data (up to 5 entries)

**Returns**: ANSI-formatted string (empty if not visible)

---

**`Focus()`**

Set modal as focused (enables keyboard handling).

---

**`Blur()`**

Remove focus from modal.

---

**`Focused() bool`**

**Returns**: true if modal is focused

---

**`Show()`**

Display the modal and set as focused.

```go
modal.Show()
```

---

**`Hide()`**

Hide the modal and remove focus.

```go
modal.Hide()
```

---

**`IsVisible() bool`**

**Returns**: true if modal is visible

---

**`SetContent(card *StatCard)`**

Update modal content from a StatCard.

**Parameters**:
- `card` - StatCard to copy content from

**Example**:
```go
modal.SetContent(focusedCard)
modal.Show()
```

---

## Interactive Components

### CommandPalette

Fuzzy-searchable command launcher with keyboard shortcuts.

#### Type Definition

```go
type CommandPalette struct {
    width      int
    height     int
    visible    bool
    focused    bool
    input      string
    commands   []Command
    filtered   []Command
    selected   int
    maxVisible int
}
```

#### Constructor

```go
func NewCommandPalette(commands []Command) *CommandPalette
```

Creates a new command palette.

**Parameters**:
- `commands` - Slice of available commands

**Returns**: Configured `*CommandPalette` instance

#### Methods

**`Toggle()`**

Toggle visibility (show if hidden, hide if visible).

---

**`Show()`**

Display the command palette.

---

**`Hide()`**

Hide the command palette.

---

**`IsVisible() bool`**

**Returns**: true if visible

---

### FileExplorer

Tree view file browser with navigation and search.

#### Type Definition

```go
type FileExplorer struct {
    width       int
    height      int
    rootPath    string
    currentPath string
    files       []FileEntry
    selected    int
    expanded    map[string]bool
    focused     bool
}
```

#### Constructor

```go
func NewFileExplorer(rootPath string) *FileExplorer
```

Creates a new file explorer.

**Parameters**:
- `rootPath` - Root directory path to explore

**Returns**: Configured `*FileExplorer` instance

---

### StatusBar

Bottom status bar with keybindings and messages.

#### Type Definition

```go
type StatusBar struct {
    width   int
    message string
    focused bool
}
```

#### Constructor

```go
func NewStatusBar() *StatusBar
```

Creates a new status bar.

**Returns**: Configured `*StatusBar` instance

**Defaults**:
- `message`: "Ready"

#### Methods

**`SetMessage(msg string)`**

Update the status message.

**Parameters**:
- `msg` - Message to display

**Example**:
```go
statusBar.SetMessage("Processing...")
```

---

### Modal

Dialog box for confirmations and inputs.

#### Type Definition

```go
type Modal struct {
    width   int
    height  int
    visible bool
    focused bool
    title   string
    content string
    buttons []string
    selected int
}
```

#### Constructor

```go
func NewModal(title, content string) *Modal
```

Creates a new modal dialog.

**Parameters**:
- `title` - Modal title
- `content` - Modal content text

**Returns**: Configured `*Modal` instance

---

## Display Components

### ActivityBar

Animated status line with spinner and progress.

#### Type Definition

```go
type ActivityBar struct {
    width      int
    height     int
    message    string
    active     bool
    startTime  time.Time
    elapsed    time.Duration
    spinner    int
    progress   string
    cancelable bool
}
```

#### Constructor

```go
func NewActivityBar() *ActivityBar
```

Creates a new activity bar.

**Returns**: Configured `*ActivityBar` instance

#### Methods

**`Start(message string) tea.Cmd`**

Start the activity animation.

**Parameters**:
- `message` - Message to display

**Returns**: Command to tick animation

**Example**:
```go
cmd := activityBar.Start("Loading...")
```

---

**`Stop()`**

Stop the activity animation.

---

**`SetProgress(progress string)`**

Update progress indicator.

**Parameters**:
- `progress` - Progress text (e.g., "↓ 2.5k tokens")

**Example**:
```go
activityBar.SetProgress("↓ 2.5k tokens")
```

---

### ToolBlock

Collapsible content block for tool execution results.

#### Type Definition

```go
type ToolBlock struct {
    toolName  string
    args      string
    lines     []string
    expanded  bool
    status    ToolBlockStatus
    maxLines  int
    lineNumbers bool
}
```

#### Constructor

```go
func NewToolBlock(toolName, args string, lines []string, opts ...ToolBlockOption) *ToolBlock
```

Creates a new tool block.

**Parameters**:
- `toolName` - Name of tool (e.g., "Bash", "Read")
- `args` - Tool arguments
- `lines` - Output lines
- `opts` - Configuration options

**Returns**: Configured `*ToolBlock` instance

---

## Layout Helpers

### LayoutHelper

Provides reusable layout patterns.

```go
var LayoutHelpers = NewLayoutHelper()
```

#### Grid Layouts

**`GridLayout(columns int, gap float64) *layout.Node`**

Equal-width column grid.

**Parameters**:
- `columns` - Number of columns
- `gap` - Gap between columns (characters)

**Returns**: Grid layout node

---

**`ResponsiveGridLayout(minCardWidth, gap float64) *layout.Node`**

Responsive grid with minimum card width.

**Parameters**:
- `minCardWidth` - Minimum width per card
- `gap` - Gap between cards

**Returns**: Responsive grid node

---

#### Column Layouts

**`TwoColumnLayout(leftRatio, rightRatio float64) *layout.Node`**

Two-column layout with ratios.

**Parameters**:
- `leftRatio` - Left column ratio
- `rightRatio` - Right column ratio

**Example**: `TwoColumnLayout(1, 2)` creates 1:2 ratio

---

**`ThreeColumnLayout(left, center, right float64) *layout.Node`**

Three-column layout with ratios.

---

**`SidebarLayout(sidebarWidth float64) *layout.Node`**

Sidebar with content area.

**Parameters**:
- `sidebarWidth` - Sidebar width in characters

---

#### Structural Layouts

**`HeaderContentFooterLayout(headerHeight, footerHeight float64) *layout.Node`**

Header/content/footer structure.

---

**`CenteredOverlay(width, height float64) *layout.Node`**

Centered overlay (for modals).

**Parameters**:
- `width` - Overlay width
- `height` - Overlay height

---

**`CenteredContent() *layout.Node`**

Center content horizontally and vertically.

---

#### Stack Layouts

**`StackLayout(gap float64) *layout.Node`**

Vertical stack with gap.

---

**`HorizontalStackLayout(gap float64) *layout.Node`**

Horizontal row with gap.

---

**`SpaceBetweenRow() *layout.Node`**

Horizontal row with space-between.

---

#### Utility Helpers

**`CardLayout(paddingCh float64) *layout.Node`**

Card container with padding.

---

**`AbsolutePosition(top, left, width, height float64) *layout.Node`**

Absolute positioning.

---

**`FlexGrowNode(grow float64) *layout.Node`**

Node with flex-grow property.

---

**`FixedSizeNode(width, height float64) *layout.Node`**

Fixed-size node.

---

## Common Types

### Command

```go
type Command struct {
    Name        string
    Description string
    Keybinding  string
    Action      func()
}
```

### FileEntry

```go
type FileEntry struct {
    Name  string
    Path  string
    IsDir bool
    Size  int64
}
```

### ToolBlockStatus

```go
const (
    StatusComplete ToolBlockStatus = iota
    StatusError
    StatusWarning
    StatusRunning
)
```

---

## Best Practices

### Focus Management

Always enable focus before expecting keyboard input:

```go
dashboard.Focus()  // Enable navigation
```

### Window Size Handling

Forward `tea.WindowSizeMsg` to all components:

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.dashboard.Update(msg)
        m.statusBar.Update(msg)
    }
    return m, nil
}
```

### Modal Integration

Let components manage their own modals:

```go
// Dashboard handles modal automatically
dashboard.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Opens modal
```

### Resource Cleanup

Components are garbage-collected; no explicit cleanup needed.

---

## See Also

- [Dashboard Guide](DASHBOARD.md) - Complete dashboard documentation
- [Components Guide](COMPONENTS.md) - All component documentation
- [Best Practices](BEST_PRACTICES.md) - Design patterns and tips
- [Layout Integration](LAYOUT_INTEGRATION.md) - Layout system details
