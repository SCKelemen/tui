## ⚠️ Deprecated — Use v2

This package (`github.com/SCKelemen/tui`) is deprecated.
Use [`github.com/SCKelemen/tui/v2`](./v2/) instead.

v2 provides the same components organized into focused sub-packages:
`agent`, `chart`, `chat`, `container`, `display`, `input`, `nav`, `selection`, `style`

See the [v2 README](./v2/README.md) for migration guide and API reference.

# tui
Terminal UI component library built on Bubble Tea with design token theming.

## Installation

```bash
go get github.com/SCKelemen/tui
```

## Features

- 25+ interactive components
- Design token theming (5 built-in themes)
- Mouse selection with OSC 52 clipboard
- Flexbox layout integration
- Keyboard-first with mouse support

## Components

### Layout

#### Application
Container for composing `tui.Component` instances with focus cycling.

```go
app := tui.NewApplication()
app.AddComponent(tui.NewStatusBar())
app.AddComponent(tui.NewActivityBar())
```

#### SplitPane
Two-pane layout with keyboard-driven focus and resize controls.

```go
split := tui.NewSplitPane(
    tui.WithLeftComponent(tui.NewFileExplorer(".")),
    tui.WithRightComponent(tui.NewConversationView()),
)
```

#### ScrollContainer
Planned standalone container; today, scrolling is built into components like `ConversationView`, `FileExplorer`, and `ThreadProgress`.

```go
cv := tui.NewConversationView()
cv.AddMessage(tui.Message{Role: tui.RoleAssistant, Content: "Hello"})
cv.ScrollToBottom()
```

#### Dashboard
Responsive metric grid for `StatCard` components with drill-down support.

```go
d := tui.NewDashboard(
    tui.WithDashboardTitle("System Metrics"),
    tui.WithCards(tui.NewStatCard(tui.WithTitle("CPU"))),
)
```

#### LayoutHelper
Helper for common `github.com/SCKelemen/layout` node patterns.

```go
helper := tui.NewLayoutHelper()
root := helper.TwoColumnLayout(1, 2)
_ = root
```

#### Header
Multi-column header with sections and alignment.

```go
h := tui.NewHeader(tui.WithColumns(
    tui.HeaderColumn{Align: tui.AlignLeft, Content: []string{"Project"}},
    tui.HeaderColumn{Align: tui.AlignRight, Content: []string{"main"}},
))
```

### Data Display

#### StatCard
Compact metric card with value, trend, and change indicators.

```go
card := tui.NewStatCard(
    tui.WithTitle("Requests"),
    tui.WithValue("12.4k"),
    tui.WithChange(220, 1.8),
)
```

#### StructuredData
Tree-like key/value renderer with status and spinner support.

```go
sd := tui.NewStructuredData("Run")
sd.AddRow("Status", "Running").AddRow("Duration", "14s")
sd.MarkInfo()
```

#### Table
Static CLI table renderer (subpackage: `tui/table`).

```go
tbl := table.New("Name", "Status")
tbl.AddRow("api", "ready")
fmt.Println(tbl.Render())
```

#### CodeBlock
Collapsible code view with optional mouse selection and copy.

```go
cb := tui.NewCodeBlock(
    tui.WithCodeFilename("main.go"),
    tui.WithCode("package main\nfunc main() {}"),
)
```

#### DiffBlock
Unified diff renderer with expand/collapse and selection support.

```go
db := tui.NewDiffBlockFromStrings("a\n", "a\nb\n",
    tui.WithDiffFilename("notes.txt"),
    tui.WithDiffSummary("Add one line"),
)
```

### Navigation

#### TabBar
Keyboard-driven tab strip with optional close behavior.

```go
tabs := tui.NewTabBar(tui.WithTabs(
    tui.Tab{ID: "one", Label: "Overview"},
    tui.Tab{ID: "two", Label: "Logs"},
))
```

#### Breadcrumb
Navigable path component with overflow truncation.

```go
bc := tui.NewBreadcrumb(tui.WithBreadcrumbItems(
    tui.BreadcrumbItem{ID: "root", Label: "workspace"},
    tui.BreadcrumbItem{ID: "repo", Label: "tui"},
))
```

#### CommandPalette
Fuzzy-search command launcher for keyboard-first workflows.

```go
cp := tui.NewCommandPalette([]tui.Command{{
    Name: "Refresh", Category: "View", Action: func() tea.Cmd { return nil },
}})
cp.Show()
```

#### FileExplorer
Tree file browser with expand/collapse and hidden file toggle.

```go
fe := tui.NewFileExplorer(".", tui.WithShowHidden(false))
_ = fe.GetSelectedPath()
fe.Focus()
```

### Input

#### TextInput
Multi-line textarea input with submit and clear shortcuts.

```go
ti := tui.NewTextInput()
ti.OnSubmit(func(s string) tea.Cmd { return nil })
ti.Focus()
```

#### Modal
Overlay dialog for alert, confirm, and input workflows.

```go
m := tui.NewModal(tui.WithModalTitle("Confirm"))
m.ShowConfirm("Delete", "Delete file?", func() tea.Cmd { return nil }, nil)
m.Focus()
```

#### ConfirmationBlock
Inline confirmation prompt with code preview and choice list.

```go
confirm := tui.NewConfirmationBlock(
    tui.WithConfirmDescription("Write config file"),
    tui.WithConfirmOptions([]string{"Yes", "No"}),
)
```

#### Checklist
Interactive checklist with nested items and status icons.

```go
cl := tui.NewChecklist(tui.WithChecklistItems(
    tui.ChecklistItem{ID: "1", Label: "Run tests"},
    tui.ChecklistItem{ID: "2", Label: "Update docs"},
))
```

### Feedback

#### ActivityBar
Animated status line with elapsed time and progress hint.

```go
ab := tui.NewActivityBar()
ab.Start("Deploying…")
ab.SetProgress("step 2/4")
```

#### StatusBar
Bottom status row for state and keybinding hints.

```go
sb := tui.NewStatusBar()
sb.SetMessage("Ready")
sb.Focus()
```

#### Toast
Stacked auto-dismiss notifications.

```go
toast := tui.NewToast()
toast.PushSuccess("Saved")
toast.PushWarning("Low disk space")
```

#### ToolBlock
Collapsible tool output block with streaming status support.

```go
tb := tui.NewToolBlock("Bash", "go test ./...", nil, tui.WithStreaming())
tb.AppendLine("ok  ./pkg")
tb.SetStatus(tui.StatusComplete)
```

#### ThreadProgress
Concurrent task/thread progress view with per-thread output.

```go
tp := tui.NewThreadProgress()
tp.UpsertThread("lint", "Lint", tui.ThreadRunning)
tp.AppendOutput("lint", "running staticcheck")
```

### Communication

#### ConversationView
Scrollable transcript view with roles, timestamps, and streaming.

```go
cv := tui.NewConversationView()
cv.AddMessage(tui.Message{Role: tui.RoleUser, Content: "Summarize tests"})
cv.AddMessage(tui.Message{Role: tui.RoleAssistant, Content: "All green"})
```

### Utilities

#### Spinner
Prebuilt animation frame sets for running states.

```go
frame := tui.SpinnerDots.GetFrame(0)
count := tui.SpinnerThinking.FrameCount()
_, _ = frame, count
```

#### IconSet
Status icon families for components that render state.

```go
icons := tui.IconSetCodex
fmt.Println(icons.Success, icons.Error)
```

#### SelectionManager
Mouse-based text selection helper used by render components.

```go
sm := tui.NewSelectionManager()
sm.SetOffset(0, 0)
_ = sm.HasSelection()
```

#### Clipboard (OSC 52)
Terminal clipboard support via OSC 52 escape sequences.

```go
cmd := tui.WriteClipboard("copied from tui")
_ = cmd
_ = tui.WriteClipboardTarget("primary", tui.ClipboardPrimary)
```

## Theming

`tui` uses `github.com/SCKelemen/design-system` tokens. You can provide tokens directly or use named theme options where supported.

```go
tokens := design.DefaultTheme()
midnight := design.MidnightTheme()
nord := design.NordTheme()
paper := design.PaperTheme()
wrapped := design.WrappedTheme()
```

Example applying tokens to components:

```go
status := tui.NewStatusBar(tui.WithStatusBarDesignTokens(tokens))
activity := tui.NewActivityBar(tui.WithActivityBarDesignTokens(midnight))
threads := tui.NewThreadProgress(tui.WithThreadProgressDesignTokens(nord))
```

Example applying a named theme directly:

```go
sd := tui.NewStructuredData("Plan", tui.WithStructuredDataTheme("paper"))
tool := tui.NewToolBlock("Read", "README.md", nil, tui.WithToolBlockTheme("wrapped"))
_ = sd
_ = tool
```

## Mouse Selection

Mouse selection is available on text-heavy components and copies via OSC 52.

- `WithCodeBlockMouseSelection(true)`
- `WithDiffBlockMouseSelection(true)`
- `WithConversationMouseSelection(true)`

When a selection exists:
- `Ctrl+C` copies selected text
- `y` copies selected text when the component is focused
- clipboard write uses `WriteClipboard` / `WriteClipboardTarget` (OSC 52)

```go
cb := tui.NewCodeBlock(
    tui.WithCode("line1\nline2"),
    tui.WithCodeBlockMouseSelection(true),
)
_ = cb
```

## Dependencies

SCKelemen stack dependencies in this module:

- `github.com/SCKelemen/cli`
- `github.com/SCKelemen/color`
- `github.com/SCKelemen/design-system`
- `github.com/SCKelemen/layout`
- `github.com/SCKelemen/text`

Additional SCKelemen modules used indirectly:

- `github.com/SCKelemen/unicode`
- `github.com/SCKelemen/units`

## License

Bearware 1.0
