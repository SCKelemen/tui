# tui/v2

A comprehensive terminal UI component library for Go, built on top of Bubble Tea and Lip Gloss.

**120+ components across 14 packages** — from basic inputs to full IDE-style experiences.

- Module: `github.com/SCKelemen/tui/v2`
- Focus: composable components, terminal-native UX, design-token theming, and AI/agent-oriented interfaces
- Scale: 120+ component files, 84 test files, broad component coverage

---

## Install

```bash
go get github.com/SCKelemen/tui/v2
```

---

## Why tui/v2?

`tui/v2` is designed for teams building serious terminal products, not just demos. It combines:

- **Reusable primitives** (inputs, panes, tabs, lists)
- **Advanced UX blocks** (command palettes, diff viewers, diagnostics, hover/peek flows)
- **AI-first interfaces** (tool execution views, context/token visualizations, agent panels)
- **Terminal compatibility support** (ANSI utilities, terminal capability detection, clipboard integration)
- **Design-token based theming** for consistent styling across components

---

## Quick Start

A minimal application using:
- the root `Application` model,
- a tokenized `CodeBlock`, and
- a `CommandPalette`.

```go
package main

import (
	"fmt"
	"log"

	"github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/display"
	"github.com/SCKelemen/tui/v2/input"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	tokens := style.DesignTokensForTheme("midnight")

	code := display.NewCodeBlock(
		display.WithCodeBlockFilename("main.go"),
		display.WithCode(`package main

func main() {
	println("hello from tui/v2")
}`),
		display.WithCodeBlockDesignTokens(tokens),
		display.WithExpanded(true),
	)

	palette := input.NewCommandPalette(
		[]input.Command{
			{Name: "Open File", Description: "Open a file", Category: "File", Keybinding: "Ctrl+O"},
			{Name: "Toggle Sidebar", Description: "Show/hide sidebar", Category: "View", Keybinding: "Ctrl+B"},
			{Name: "Quit", Description: "Exit application", Category: "System", Keybinding: "Ctrl+C", Action: func() tea.Cmd { return tea.Quit }},
		},
		input.WithCommandPaletteDesignTokens(tokens),
	)

	app := tui.NewApplication(tui.WithQuitKey("ctrl+c"))
	app.AddComponent(code)
	app.AddComponent(palette)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Goodbye")
}
```

> Tip: the palette is designed for VS Code-style interaction (`Ctrl+K` / `Ctrl+P`), depending on component focus and event flow.

---

## Packages

### `display` (55 components)

The largest package — rendering primitives, data visualization, IDE features, calendars, and specialized displays.

#### Code & Diff

| Component | Description |
|---|---|
| `CodeBlock` | Syntax-highlighted code with line numbers, selection, and clipboard support |
| `DiffBlock` | Unified diff rendering with add/remove highlights |
| `SideBySideDiff` | Two-pane diff comparison |
| `DiffDetailView` | Detailed diff viewer with context |
| `DiffFileList` | File list for multi-file diffs |
| `GitDiffView` | Git-style diff rendering |
| `GitFileList` | Git file status list |
| `Highlight` | Syntax highlighting engine |
| `Gutter` | Line number gutter with decorations |

#### IDE Features

| Component | Description |
|---|---|
| `Diagnostics` | LSP-style diagnostic list with severity icons |
| `HoverCard` | Hover documentation popup |
| `PeekView` | Peek definition inline viewer |
| `ToolUseBlock` | AI tool use display |
| `ToolUseLoader` | Tool execution progress |
| `ToolResultDisplay` | Tool result rendering |
| `ContextVisualization` | Token/context usage visualization |
| `TokenWarning` | Token limit warning display |
| `MemoryUsageIndicator` | Memory usage bar |

#### Calendar

| Component | Description |
|---|---|
| `MonthView` | Full month calendar grid |
| `WeekView` | 7-day week view with time slots |
| `ThreeDayView` | 3-day focused view |
| `DayView` | Single day with hour-by-hour timeline |
| `TaskView` | Task-focused calendar view |
| `CalendarTypes` | Shared calendar event/task types |

#### Text & Content

| Component | Description |
|---|---|
| `BigText` | Large ASCII art text renderer |
| `GradientText` | Text with color gradients |
| `ThemedText` | Token-aware styled text |
| `Markdown` | Markdown renderer |
| `RawAnsi` | Raw ANSI escape passthrough |
| `Link` | Clickable hyperlinks (OSC 8) |
| `Tips` | Rotating tip display |
| `GlimmerMessage` | Animated shimmer message |

#### Data Display

| Component | Description |
|---|---|
| `StatusBar` | Multi-section status bar |
| `StatusLine` | Single-line status with indicators |
| `StatusIcon` | Semantic status icons |
| `StatCard` | Statistic card with label and value |
| `FileCard` | File information card |
| `FileStatusBadge` | Git-style file status badge (M/A/D/R) |
| `ErrorFrame` | Error display frame |
| `Toast` | Notification toast |
| `LoadingState` | Loading placeholder |
| `Shimmer` | Animated loading shimmer |
| `ListItem` | Styled list item |
| `OrderedList` | Numbered list |
| `KeyboardShortcutHint` | Keybind display |
| `Divider` | Horizontal divider |
| `ActionRow` | Action button row |
| `RatingDots` | Rating indicator dots |
| `ScrollStatusBar` | Scroll position indicator |

#### Architecture Diagrams

| Component | Description |
|---|---|
| `Diagram` | Programmatic box-and-arrow architecture diagrams with Unicode rendering |

#### Agents & AI

| Component | Description |
|---|---|
| `BackgroundTaskPanel` | Background task monitoring |
| `MCPServerList` | MCP server connection list |
| `MCPToolDetail` | MCP tool detail view |
| `ShellOutputView` | Shell command output display |
| `VirtualMessageList` | Virtualized message list with viewport scrolling |
| `NoSelect` | Non-selectable text wrapper |

---

### `input` (23 components)

Everything interactive — text entry, selection, search, code intelligence, permissions.

#### Text Input

| Component | Description |
|---|---|
| `TextInput` | Single-line text input |
| `Prompt` | Multi-line prompt with history |
| `ContinuationPrompt` | Multi-turn continuation input |

#### Selection & Search

| Component | Description |
|---|---|
| `SelectInput` | Single-select dropdown |
| `RadioButtonSelect` | Radio button group |
| `Checklist` | Multi-select checklist |
| `Autocomplete` | Input with autocomplete suggestions |
| `FuzzyPicker` | Fuzzy-search picker |
| `FuzzyMatch` | fzf-style fuzzy matching engine (scoring, smart case, match highlighting) |
| `CommandPalette` | VS Code-style command palette (Ctrl+K) |
| `FloatingPalette` | Floating search/filter palette |
| `TreeSelect` | Tree-structured selection |
| `ColorPicker` | Color selector |

#### IDE Features

| Component | Description |
|---|---|
| `IntelliSense` | Code completion popup with filtering |
| `CodeAction` | Quick fix / refactoring menu |

#### Permissions & Approval

| Component | Description |
|---|---|
| `Approval` | Accept/reject prompt |
| `VisibilityToggle` | Show/hide toggle |
| `ElicitationForm` | Dynamic form generation |

#### Configuration

| Component | Description |
|---|---|
| `ModelPicker` | LLM model selector |
| `ThemePicker` | Theme selector |
| `SessionSelector` | Session picker |
| `KeyBind` | Keybind manager and display |
| `KeyOverlay` | Keyboard shortcut overlay |

---

### `container` (11 components)

Layout containers, dialogs, and permission views.

| Component | Description |
|---|---|
| `Dialog` | Modal dialog with actions |
| `Modal` | Generic modal overlay |
| `SplitPane` | Resizable split view |
| `Pane` | Basic content pane |
| `ThemedBox` | Token-themed container |
| `TitledBox` | Container with title bar |
| `AgentWizard` | Step-by-step agent setup wizard |
| `PermissionDialog` | Permission request dialog |
| `BashPermissionView` | Shell command permission view |
| `FilePermissionView` | File access permission view |
| `ScreenSizeOverlay` | Terminal size display overlay |

---

### `agent` (5 components)

AI agent-specific UI — subagent panels, thread progress, diff lists.

| Component | Description |
|---|---|
| `SubagentPanel` | Subagent work display with status footer |
| `SubagentGroup` | Multi-panel agent group with height matching |
| `SubagentDiffList` | Agent-generated diff list |
| `TeammateSpinnerTree` | Hierarchical agent activity tree |
| `ThreadProgress` | Thread execution progress |

---

### `chart` (5 components)

Data visualization and metrics.

| Component | Description |
|---|---|
| `ProgressBar` | Gradient progress bar with OKLCH interpolation |
| `Sparkline` | Inline sparkline chart |
| `TermChart` | Terminal bar/line charts |
| `UsageBar` | Segmented usage bar |
| `MemoryUsageIndicator` | Memory usage display |

---

### `chat` (2 components)

Conversation UI.

| Component | Description |
|---|---|
| `ConversationView` | Chat message thread |
| `MessageBubble` | Styled message bubble with role indicators |

---

### `nav` (6 components)

Navigation and routing.

| Component | Description |
|---|---|
| `Router` | View-based routing system |
| `TabBar` | Tab navigation bar |
| `Breadcrumb` | Breadcrumb navigation |
| `RoleTabs` | Role-based tab selector (user/assistant/system) |
| `ScrollContainer` | Scrollable viewport |
| `Tree` | Expandable tree view |

---

### `spinner` (3 components)

Loading animations.

| Component | Description |
|---|---|
| `Spinner` | Configurable spinner with 10+ styles (dots, moon, earth, etc.) |
| `Scanner` | Knight Rider-style scanning animation |
| `Catalog` | Spinner style catalog/preview |

---

### `event` (1 component)

Event system.

| Component | Description |
|---|---|
| `Bus` | Typed pub/sub event bus |

---

### `selection` (2 components)

Text selection and clipboard.

| Component | Description |
|---|---|
| `SelectionManager` | Mouse-driven text selection with content-aware boundaries |
| `Clipboard` | Cross-platform clipboard with OSC 52 support |

---

### `style` (4 files)

Styling utilities and terminal detection.

| File / Utility | Description |
|---|---|
| `Style` | Design token resolution and theme management |
| `ColorBridge` | Color space conversion (OKLCH, hex, ANSI) |
| `Terminal` | Terminal capability detection (truecolor, mouse, Unicode) |
| `TextUtil` | Text measurement and wrapping |

---

### `layout` (2 components)

Layout primitives.

| Component | Description |
|---|---|
| `HalfLinePaddedBox` | Half-line padding container |
| `VerticalSeparator` | Vertical divider line |

---

### Root (1 file)

| File | Description |
|---|---|
| `tui.go` | Application model, design token types, Bubble Tea integration |

---

## Architecture Diagrams

Use `display.Diagram` for terminal-rendered architecture drawings:

```go
d := display.NewDiagram().
	SetTitle("System Architecture").
	AddBox(display.DiagramBox{ID: "client", Title: "CLI Client", Lines: []string{"(TUI + Auth)"}, Row: 0, Col: 0}).
	AddBox(display.DiagramBox{ID: "server", Title: "Backend Svc", Lines: []string{"(Context Mgr)", "(Tool Runner)", "(Session DB)"}, Row: 0, Col: 1}).
	AddBox(display.DiagramBox{ID: "llm", Title: "LLM APIs", Lines: []string{"(OpenAI,", " Anthropic)"}, Row: 0, Col: 2}).
	AddArrow(display.DiagramArrow{From: "client", To: "server", Label: "gRPC bidi stream", Direction: display.ArrowBidirectional}).
	AddArrow(display.DiagramArrow{From: "server", To: "llm", Label: "LLM API calls", Direction: display.ArrowBidirectional})

fmt.Println(d.Render())
```

---

## Fuzzy Search

`input.FuzzyMatcher` provides fzf-style scoring and match position metadata:

```go
matcher := input.NewFuzzyMatcher()
result := matcher.Match("cmdpal", "CommandPalette")

// result.Score
// result.Matched
// result.Positions
```

This powers palette/picker experiences with:
- smart-case behavior,
- scoring bonuses for boundaries/consecutive matches,
- ranked candidate sorting,
- highlighted match positions.

---

## Design Tokens

The library uses a design token approach for visual consistency across components.

Many components expose package-specific token options such as:

- `display.WithCodeBlockDesignTokens(tokens)`
- `input.WithCommandPaletteDesignTokens(tokens)`

A common flow is:

```go
tokens := style.DesignTokensForTheme("midnight")
```

Then pass `tokens` into components so colors, accents, surfaces, and muted text remain consistent across the application.

---

## Testing

`tui/v2` includes **84 test files** with broad component-level validation.

Run all tests:

```bash
go test ./...
```

---

## Compatibility Notes

- Built for modern terminals with ANSI support.
- Includes capability-aware styling and fallbacks.
- Clipboard support includes OSC 52 paths where available.

---

## License

MIT — Samuel Kelemen
