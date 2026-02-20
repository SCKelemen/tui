# TUI Components

Codex CLI-inspired components for building sophisticated terminal UIs.

## Available Components

### 1. ActivityBar

Animated status line with spinner, elapsed time, and progress indicators.

**Features:**
- Spinning animation (⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏)
- Elapsed time display (14s, 1m 14s format)
- Progress indicators (e.g., "↓ 2.5k tokens")
- Cancelable with Esc key
- Automatic color themes

**Options:**
- `WithActivityBarDesignTokens(tokens)` - Apply theme colors from `design-system`
- `WithActivityBarTheme(name)` - Apply named theme (`default`, `midnight`, `nord`, `paper`, `wrapped`)

**Example:**
```go
activityBar := tui.NewActivityBar()
activityBar.Start("Actualizing…")
activityBar.SetProgress("↓ 2.5k tokens")
// ... later
activityBar.Stop()
```

**Output:**
```
✳ Actualizing… (esc to interrupt · 1m 14s · ↓ 2.5k tokens)
```

---

### 2. ToolBlock

Collapsible content blocks for displaying tool execution results with real-time streaming support.

**Features:**
- Collapsible/expandable output
- Real-time streaming output (AppendLine/AppendLines)
- Status indicators: ✓ (complete), ✗ (error), ⚠ (warning), animated spinner (running)
- Color-coded by status (green, red, yellow, cyan)
- Line numbers for code files
- Tree-style indentation
- "... +N lines" summary when collapsed
- Tool-specific icons (⏺)
- Ctrl+O or Enter to expand/collapse

**Options:**
- `WithLineNumbers()` - Show line numbers (for code)
- `WithMaxLines(n)` - Limit visible lines when collapsed
- `WithStreaming()` - Enable streaming mode with running status
- `WithStatus(status)` - Set initial status (StatusComplete, StatusError, StatusWarning, StatusRunning)
- `WithToolBlockDesignTokens(tokens)` - Apply theme colors from `design-system`
- `WithToolBlockTheme(name)` - Apply named theme (`default`, `midnight`, `nord`, `paper`, `wrapped`)

**Example:**
```go
block := tui.NewToolBlock(
    "Bash",
    "go test -v",
    []string{"=== RUN   TestFoo", "--- PASS: TestFoo (0.00s)"},
    tui.WithMaxLines(3),
)
```

**Output (Collapsed):**
```
⏺ Bash(go test -v)
  ⎿  === RUN   TestFoo
     --- PASS: TestFoo (0.00s)
     === RUN   TestBar
     … +12 lines (ctrl+o to expand)
```

**Output (Expanded):**
```
⏺ Bash(go test -v)
  ⎿  === RUN   TestFoo
     --- PASS: TestFoo (0.00s)
     === RUN   TestBar
     --- PASS: TestBar (0.00s)
     PASS
```

**With Line Numbers:**
```go
block := tui.NewToolBlock(
    "Write",
    "main.go",
    []string{"package main", "", "func main() {", "    fmt.Println(\"hello\")", "}"},
    tui.WithLineNumbers(),
)
```

**Output:**
```
⏺ Write(main.go)
  ⎿    1 package main
       2
       3 func main() {
       4     fmt.Println("hello")
       5 }
```

**Streaming Mode:**
```go
// Create block in streaming mode
block := tui.NewToolBlock(
    "Bash",
    "go test -v",
    []string{},
    tui.WithStreaming(),
)

// Append output as it arrives
block.AppendLine("=== RUN   TestFoo")
block.AppendLine("--- PASS: TestFoo (0.00s)")
block.AppendLines([]string{
    "=== RUN   TestBar",
    "--- PASS: TestBar (0.00s)",
})

// Complete when done
block.SetStatus(tui.StatusComplete)
```

**Streaming Output:**
```
⏺ Bash(go test -v) ⠋   (while running with animated spinner)
  ⎿  streaming...

⏺ Bash(go test -v) ✓   (when complete)
  ⎿  === RUN   TestFoo
     --- PASS: TestFoo (0.00s)
     === RUN   TestBar
     --- PASS: TestBar (0.00s)
```

**Status States:**
```go
// Success
tui.WithStatus(tui.StatusComplete)  // Green ✓

// Error
tui.WithStatus(tui.StatusError)     // Red ✗

// Warning
tui.WithStatus(tui.StatusWarning)   // Yellow ⚠

// Running (auto-set with WithStreaming)
tui.WithStatus(tui.StatusRunning)   // Cyan with spinner
```

---

### 3. StatusBar

Simple status bar with message and keybindings.

**Features:**
- Left-aligned status message
- Right-aligned keybinding hints
- Visual feedback when focused (inverted colors)
- Auto-truncation for narrow terminals

**Options:**
- `WithStatusBarDesignTokens(tokens)` - Apply theme colors from `design-system`
- `WithStatusBarTheme(name)` - Apply named theme (`default`, `midnight`, `nord`, `paper`, `wrapped`)

**Example:**
```go
statusBar := tui.NewStatusBar()
statusBar.SetMessage("Processing files...")
```

**Output:**
```
Processing files...                                    Tab: Focus • q: Quit
```

---

### 4. TextInput

Multi-line text input component for user messages.

**Features:**
- Multi-line text editing with textarea support
- Submit with Ctrl+J (Ctrl+Enter)
- Clear with Ctrl+D
- Bordered container with visual hints
- Callback support for message submission
- Placeholder text when empty
- Character limit (10,000 by default)

**Example:**
```go
textInput := tui.NewTextInput()
textInput.OnSubmit(func(text string) tea.Cmd {
    // Handle submitted message
    fmt.Println("User said:", text)
    return nil
})
app.AddComponent(textInput)
```

**Output:**
```
┌──────────────────────────────────────────┐
│ ┃ Type your message here...              │
│ ┃                                         │
│ ┃                                         │
└ Ctrl+J: send · Ctrl+D: clear ────────────┘
```

---

### 5. CommandPalette

Fuzzy-searchable command launcher (like VS Code's Ctrl+P).

**Features:**
- Modal overlay that appears on Ctrl+K or Ctrl+P
- Fuzzy search filtering as you type
- Up/Down arrow navigation
- Enter to execute selected command
- Esc to dismiss
- Shows command name, description, and keybinding
- Category grouping support
- Custom action callbacks

**Example:**
```go
commands := []tui.Command{
    {
        Name:        "Clear Messages",
        Description: "Clear all message history",
        Category:    "Edit",
        Keybinding:  "Ctrl+L",
        Action: func() tea.Cmd {
            return clearMessagesCmd()
        },
    },
    {
        Name:        "Toggle Activity",
        Description: "Start/stop activity animation",
        Category:    "View",
        Keybinding:  "Ctrl+A",
        Action: func() tea.Cmd {
            return toggleActivityCmd()
        },
    },
}

palette := tui.NewCommandPalette(commands)
app.AddComponent(palette)
```

**Output:**
```
          ┌────────── Command Palette ──────────┐
          │ > clear                             │
          ├─────────────────────────────────────┤
          │ ▸ Clear Messages           Ctrl+L  │
          │   Clear Cache                       │
          └ 2 commands ─────────────────────────┘
```

---

### 6. Application

Container for managing multiple components with focus.

**Features:**
- Component lifecycle management (Init, Update, View, Focus, Blur)
- Tab/Shift+Tab focus navigation
- Window size handling
- Quit keys (q, Ctrl+C)

**Example:**
```go
app := tui.NewApplication()

activityBar := tui.NewActivityBar()
toolBlock := tui.NewToolBlock("Bash", "ls", []string{"file1", "file2"})

app.AddComponent(activityBar)
app.AddComponent(toolBlock)

p := tea.NewProgram(app, tea.WithAltScreen())
p.Run()
```

---

### 6. FileExplorer

Tree-based file system navigator with keyboard controls.

**Features:**
- Tree view with expand/collapse
- Lazy loading (directories load on expand)
- Show/hide hidden files (toggle with `.`)
- Keyboard navigation (vim-style or arrows)
- Visual indicators: 📁 (collapsed), 📂 (expanded), 📄 (file)
- Depth indentation with tree connectors
- Scroll handling for long lists
- Parent/child relationships
- Refresh on demand

**Example:**
```go
fileExplorer := tui.NewFileExplorer("/path/to/directory",
    tui.WithShowHidden(false))
app.AddComponent(fileExplorer)

// Get selected path
path := fileExplorer.GetSelectedPath()

// Get selected node
node := fileExplorer.GetSelectedNode()
if node != nil {
    fmt.Printf("Selected: %s (IsDir: %v)\n", node.Name, node.IsDir)
}
```

**Keyboard Controls:**
- `↑/k` - Move selection up
- `↓/j` - Move selection down
- `→/l or Enter` - Expand directory
- `←/h` - Collapse directory or move to parent
- `.` - Toggle hidden files
- `r` - Refresh current directory

**Output:**
```
📁 /home/user/projects

  📂 myproject
  ├─ 📁 src
  ├─ 📄 go.mod
  ├─ 📄 go.sum
  └─ 📄 README.md

[1/15]
↑↓: navigate · Enter: open · .: toggle hidden · r: refresh
```

---

### 7. Modal

Overlay dialogs for user interaction (alerts, confirmations, input).

**Features:**
- Three modal types: Alert, Confirm, Input
- Centered overlay with backdrop
- Keyboard navigation between buttons (Tab/Shift+Tab)
- Text wrapping for long messages
- Optional text input field
- Callback support for user actions
- ESC to cancel, Enter to confirm
- Customizable buttons and actions

**Modal Types:**

**Alert** - Information with OK button:
```go
modal.ShowAlert(
    "Success",
    "Operation completed successfully!",
    func() tea.Cmd {
        // Handle OK
        return nil
    },
)
```

**Confirm** - Yes/No question:
```go
modal.ShowConfirm(
    "Delete File",
    "Are you sure you want to delete this file?",
    func() tea.Cmd {
        // Handle Yes
        return deleteFileCmd()
    },
    func() tea.Cmd {
        // Handle No
        return nil
    },
)
```

**Input** - Text entry with OK/Cancel:
```go
modal.ShowInput(
    "Enter Name",
    "Please enter your name:",
    "John Doe", // placeholder
    func(value string) tea.Cmd {
        // Handle OK with value
        return processNameCmd(value)
    },
    func() tea.Cmd {
        // Handle Cancel
        return nil
    },
)
```

**Keyboard Controls:**
- `Tab / →` - Next button
- `Shift+Tab / ←` - Previous button
- `Enter` - Confirm selected button
- `Esc` - Cancel/close modal

**Output:**
```
╭─── Confirmation ──────────────────────────╮
│                                           │
│  Are you sure you want to proceed with   │
│  this action? This cannot be undone.     │
│                                           │
│            [ Yes ]  [ No ]               │
│                                           │
└─ Tab: navigate · Enter: confirm · Esc ───┘
```

**Custom Buttons:**
```go
modal := tui.NewModal(
    tui.WithModalTitle("Choose Option"),
    tui.WithModalMessage("Select one:"),
    tui.WithModalButtons([]tui.ModalButton{
        {Label: "Option 1", Action: func(s string) tea.Cmd { return nil }},
        {Label: "Option 2", Action: func(s string) tea.Cmd { return nil }},
        {Label: "Cancel", Action: func(s string) tea.Cmd { return nil }},
    }),
)
```

---

### 8. Application

Container for managing multiple components with focus.

**Features:**
- Component lifecycle management (Init, Update, View, Focus, Blur)
- Tab/Shift+Tab focus navigation
- Window size handling
- Quit keys (q, Ctrl+C)

**Example:**
```go
app := tui.NewApplication()

activityBar := tui.NewActivityBar()
toolBlock := tui.NewToolBlock("Bash", "ls", []string{"file1", "file2"})

app.AddComponent(activityBar)
app.AddComponent(toolBlock)

p := tea.NewProgram(app, tea.WithAltScreen())
p.Run()
```

---

### 9. Header

Multi-column header with rounded borders and vertical dividers.

**Features:**
- Multi-column layout with configurable widths
- Content alignment per column (left, center, right)
- Sections within columns with optional titles
- Horizontal dividers between sections
- Vertical dividers between columns
- Rounded corner borders (╭╮╰╯)
- UTF-8 aware width calculations

**Example:**
```go
header := tui.NewHeader(
    tui.WithColumns(
        // Left column: centered
        tui.HeaderColumn{
            Width:   40,
            Align:   tui.AlignCenter,
            Content: []string{
                "",
                "Welcome back!",
                "",
                "TUI Framework v1.0",
                "",
            },
        },
        // Right column: left-aligned with sections
        tui.HeaderColumn{
            Width: 60,
            Align: tui.AlignLeft,
        },
    ),
    tui.WithColumnSections(1,
        tui.HeaderSection{
            Title:   "Tips for getting started",
            Content: []string{
                "Use Tab to navigate between components",
                "Press q to quit applications",
            },
        },
        tui.HeaderSection{
            Title:   "Recent activity",
            Content: []string{
                "No recent activity",
            },
            Divider: true, // Add horizontal divider before this section
        },
    ),
    tui.WithVerticalDivider(true),
)
```

**Output:**
```
╭────────────────────────────────────────────────────────────────────╮
│                                    │ Tips for getting started      │
│          Welcome back!             │ Use Tab to navigate...        │
│                                    │ Press q to quit...            │
│       TUI Framework v1.0           │ ─────────────────────────────│
│                                    │ Recent activity               │
│                                    │ No recent activity            │
╰────────────────────────────────────────────────────────────────────╯
```

**Column Alignment:**
- `AlignLeft` - Content aligned to the left
- `AlignCenter` - Content centered in column
- `AlignRight` - Content aligned to the right

**Sections:**
- Add multiple sections to a column with `WithColumnSections(columnIndex, ...sections)`
- Each section can have a title and content
- Use `Divider: true` to add horizontal separator before section

---

### 10. StructuredData

Displays formatted key-value data with tree connectors, similar to ToolBlock but optimized for structured information.

**Features:**
- Builder pattern API for ergonomic data construction
- Key-value pairs with automatic alignment
- Section headers
- Nested/indented items
- Colored rows
- Collapsible when using `WithStructuredDataMaxLines()`
- Ctrl+O or Enter to expand/collapse
- Helper functions: `FromMap()`, `FromKeyValuePairs()`

**API:**
```go
// Create with builder pattern
sd := tui.NewStructuredData("Session Summary").
    AddRow("Total cost", "$122.25").
    AddRow("Total duration", "6h 10m 48s").
    AddSeparator().
    AddHeader("Usage by model").
    AddIndentedRow("codex-mini", "797.2k input", 1).
    AddIndentedRow("codex-pro", "970.4k output", 1)

// Or use helper functions
sd := tui.FromKeyValuePairs("Config",
    "Host", "localhost",
    "Port", "8080",
    "SSL", "enabled",
)
```

**Item Types:**
- `AddRow(key, value)` - Key-value pair
- `AddColoredRow(key, value, color)` - Colored key-value pair
- `AddIndentedRow(key, value, indent)` - Indented key-value pair
- `AddHeader(text)` - Section header (bold)
- `AddSeparator()` - Blank line
- `AddValue(text)` - Value-only line (no key)
- `AddIndentedValue(text, indent)` - Indented value-only line

**Options:**
- `WithStructuredDataMaxLines(n)` - Collapse to N lines
- `WithKeyWidth(width)` - Fixed width for key column (auto if 0)
- `WithStructuredDataIcon(icon)` - Custom icon (deprecated, use WithIconSet)
- `WithRunningColor(color)` - ANSI color code for running status (default: white "\033[37m")
- `WithSpinner(spinner)` - Spinner animation (default: SpinnerBlink)
- `WithIconSet(iconSet)` - Icon set for status indicators (default: IconSetDefault)
- `WithStructuredDataDesignTokens(tokens)` - Apply theme colors from `design-system`
- `WithStructuredDataTheme(name)` - Apply named theme (`default`, `midnight`, `nord`, `paper`, `wrapped`)

**Animated Spinners & Status Icons:**

The component supports configurable spinner animations and icon sets:

```go
// Use different spinner animations
sd := tui.NewStructuredData("Task", tui.WithSpinner(tui.SpinnerThinking))
sd := tui.NewStructuredData("Task", tui.WithSpinner(tui.SpinnerDots))

// Use different icon sets
sd := tui.NewStructuredData("Task", tui.WithIconSet(tui.IconSetCodex))
sd := tui.NewStructuredData("Task", tui.WithIconSet(tui.IconSetSymbols))

// Start animation
cmd := sd.StartRunning()

// Mark as complete with status-based color
sd.MarkSuccess()  // Green icon
sd.MarkError()    // Red icon
sd.MarkWarning()  // Yellow icon
sd.MarkInfo()     // White icon

// Customize running color
sd := tui.NewStructuredData("Task",
    tui.WithSpinner(tui.SpinnerThinking),
    tui.WithIconSet(tui.IconSetCodex),
    tui.WithRunningColor("\033[36m"))
```

**Available Spinners:**
- `SpinnerBlink` - Simple blink on/off (default)
- `SpinnerThinking` - Codex CLI style (small to large): . + * ÷ •
- `SpinnerCodexThinking` - Alias for `SpinnerThinking`
- `SpinnerClaudeThinking` - Backward-compatible alias for `SpinnerThinking`
- `SpinnerDots` - Braille dots (smooth): ⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏
- `SpinnerLine` - Classic line: ─ \ | /
- `SpinnerCircle` - Rotating circle: ◴◷◶◵
- `SpinnerPulse` - Pulsing circle: ○◔◐◕●◕◐◔
- `SpinnerArrows` - Rotating arrows: ←↖↑↗→↘↓↙
- `SpinnerArc` - Growing arc: ◜◠◝◞◡◟
- `SpinnerCircleQuarters` - Circle quarters: ◐◓◑◒
- `SpinnerSquare` - Rotating square: ◰◳◲◱
- `SpinnerDotsJumping` - Jumping dots: ⢄⢂⢁⡁⡈⡐⡠
- `SpinnerBouncingBar` - Bouncing bar animation
- `SpinnerBouncingBall` - Bouncing ball animation

**Available Icon Sets:**
- `IconSetDefault` - All ⏺ (default)
- `IconSetCodex` - Codex CLI style: ⏺ ✓ ✗ ⚠ ⏺
- `IconSetClaude` - Backward-compatible alias for `IconSetCodex`
- `IconSetSymbols` - Unicode symbols: ⏺ ✓ ✗ ⚠ ℹ
- `IconSetEmoji` - Emoji: ⏺ ✅ ❌ ⚡ 💡
- `IconSetCircles` - Circles: ○ ● ◯ ◐ ○
- `IconSetMinimal` - ASCII: · + x ! i

**Status Colors:**
- `DataStatusRunning` - Animated spinner (color configurable)
- `DataStatusSuccess` - Static green icon
- `DataStatusError` - Static red icon
- `DataStatusWarning` - Static yellow icon
- `DataStatusInfo` - Static white icon
- `DataStatusNone` - Static cyan icon (default)

The animation runs at 500ms intervals and automatically stops when status changes from Running to a final state.

**Output:**
```
⏺ Session Summary
  ⎿  Total cost:           $122.25
     Total duration (API): 6h 10m 48s
     Total duration (wall): 1d 20h 37m
     Total code changes:   26773 lines added, 2436 lines removed

     Usage by model
       codex-mini:       797.2k input, 65.9k output
       codex-pro:      970.4k output, 189.5m cache read
```

**Use Cases:**
- Cost/billing summaries
- Configuration display
- API response formatting
- Test results summary
- System stats
- Any tabular key-value data

**Advanced:**
```go
// Multiple sections with indentation
sd := tui.NewStructuredData("Config").
    AddHeader("Server").
    AddIndentedRow("Host", "localhost", 1).
    AddIndentedRow("Port", "8080", 1).
    AddSeparator().
    AddHeader("Database").
    AddIndentedRow("Driver", "postgresql", 1).
    AddIndentedRow("Pool Size", "20", 1).
    AddSeparator().
    AddHeader("Features").
    AddIndentedRow("Auth", "enabled", 1).
    AddIndentedRow("Caching", "redis", 1)

// Color coding
sd.AddColoredRow("Passed", "170", "\033[32m")    // Green
sd.AddColoredRow("Failed", "0", "\033[2m")       // Dim
```

---

## Component Interface

All components implement:

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

---

## Keyboard Shortcuts

### Global
| Key | Action |
|-----|--------|
| Tab | Focus next component |
| Shift+Tab | Focus previous component |
| q or Ctrl+C | Quit application |

### ToolBlock
| Key | Action |
|-----|--------|
| Ctrl+O or Enter | Expand/collapse ToolBlock |

### TextInput
| Key | Action |
|-----|--------|
| Ctrl+J | Submit text |
| Ctrl+D | Clear text |

### CommandPalette
| Key | Action |
|-----|--------|
| Ctrl+K or Ctrl+P | Open CommandPalette |
| Up/Down | Navigate items |
| Enter | Execute selected command |
| Esc | Close palette |

### FileExplorer
| Key | Action |
|-----|--------|
| ↑/k | Move selection up |
| ↓/j | Move selection down |
| →/l or Enter | Expand directory |
| ←/h | Collapse directory or move to parent |
| . | Toggle hidden files |
| r | Refresh directory |

### Modal
| Key | Action |
|-----|--------|
| Tab / → | Next button |
| Shift+Tab / ← | Previous button |
| Enter | Confirm selected button |
| Esc | Cancel/close modal |

### ActivityBar
| Key | Action |
|-----|--------|
| Esc | Interrupt running activity |

---

## Examples

### Basic Demo
```bash
go run examples/basic/main.go
```

### Codex CLI Style Demo
```bash
go run examples/codex_code_demo/main.go
```

### Claude Style Demo (Compatibility)
```bash
go run examples/claude_code_demo/main.go
```

### Input Components Demo (Non-interactive)
```bash
go run examples/input_demo_output/main.go
```

### Input Components Demo (Interactive)
```bash
go run examples/input_demo/main.go
```

### Streaming ToolBlocks Demo (Non-interactive)
```bash
go run examples/streaming_demo_output/main.go
```

### Streaming ToolBlocks Demo (Interactive)
```bash
go run examples/streaming_demo/main.go
```

### FileExplorer Demo (Interactive)
```bash
go run examples/fileexplorer_demo/main.go
```

### Modal Demo (Interactive)
```bash
go run examples/modal_demo/main.go
```

### Header Demo (Interactive)
```bash
go run examples/header_demo/main.go
```

---

## Future Components (Planned)

- **Editor**: Text viewing/editing with syntax highlighting
- **Tabs**: Multi-view tab management
- **SidePanel**: Collapsible side panels with sections
- **SearchResults**: Searchable result lists with context
- **DiffViewer**: Side-by-side or unified diff display
- **ProgressBar**: Progress indicator for long-running operations
- **Table**: Sortable, scrollable data tables

---

## Integration with SCKelemen Stack

Future v2 components will leverage:

- **cli/renderer**: Double-buffered screen rendering, ANSI output
- **layout**: Flexbox/grid layouts for complex UIs
- **text**: Unicode-aware text width measurement
- **design-system**: Design tokens and theme management
- **color**: OKLCH color space, gradients, accessibility
- **units**: CSS-like units (px, ch, vw, vh)

**Status**: v1 components use simple ANSI rendering for immediate usability. v2 refactor will add full stack integration when all packages are public.

**ActivityBarV2**: An experimental v2 implementation (`activitybar_v2.go`) exists that demonstrates full stack integration. It requires private packages and is gated behind a build tag:

```bash
# Standard build (v1 components only)
go build

# Build with stack integration (requires private repos)
go build -tags stack
```

This pattern will be extended to other components as the stack packages become public.

---

## License

Bearware 1.0
