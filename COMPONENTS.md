# TUI Components

Claude Code-inspired components for building sophisticated terminal UIs.

## Available Components

### 1. ActivityBar

Animated status line with spinner, elapsed time, and progress indicators.

**Features:**
- Spinning animation (⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏)
- Elapsed time display (14s, 1m 14s format)
- Progress indicators (e.g., "↓ 2.5k tokens")
- Cancelable with Esc key
- Automatic color themes

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

### Claude Code Style Demo
```bash
go run examples/claude_demo_output/main.go
```

### Interactive Demo
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

---

## Future Components (Planned)

- **Header**: Fancy multi-column headers with rounded corners (like Claude Code's welcome screen)
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
