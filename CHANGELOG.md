# Changelog

## [tui/v2.20.1] - 2026-05-19

Patch release fixing two HIGH-severity concurrency bugs in `event.Bus` surfaced by a code-review sweep on `tui/v2.20.0`. Both bugs predated this session but were not caught earlier because they are logical races rather than data races — `go test -race` does not surface them.

### Fixed

- **`event.Publish` no longer panics with "send on closed channel" when a publisher races a `Subscription.Close()`.** Subscribers now own a `done` channel that is closed instead of closing the raw event channel. `Publish`'s select adds a `<-sub.done` case so it skips concurrently-torn-down subscribers without panicking. Raw channels are never closed; GC reclaims them when the adapter goroutine exits.
- **`Subscription` adapter goroutine no longer leaks when `Close` races a slow consumer.** The adapter's inner `typedCh <- value` send is now wrapped in `select { case typedCh <- v: case <-done: return }`, so closing `done` unblocks the adapter even when the typed channel buffer is full and the consumer has stopped draining.

### Tests

- `TestPublishConcurrentCloseDoesNotPanic` — 8 publishers vs 4 subscribe-then-close churners for 500ms; asserts no panic.
- `TestCloseUnblocksAdapterWithFullBuffer` — fills the typed channel, calls Close, asserts the adapter exits within 250ms.
- `TestSubscribeLegacyChannelStillDelivers` — regression guard for the legacy channel-returning `Subscribe` API after the refactor.

### Internal

- `Bus.subscribers` is now `map[string][]*subscriber` (was `map[string][]chan interface{}`) where `type subscriber struct { ch chan interface{}; done chan struct{} }`. All public APIs (`Subscribe`, `SubscribeWithHandle`, `Subscription.Chan`, `Subscription.Close`, `Publish`, `Unsubscribe`, `DroppedEvents`, `SetOnDrop`, `NewBus`, `Bus`, `BusMsg`, `Event`) retain their existing signatures.

## [tui/v2.20.0] - 2026-05-18

First release on the `tui/v2.x` line since `tui/v2.19.0`. Consolidates three waves of correctness fixes and OpenTUI-inspired feature work, plus dependency bumps. 48 commits, 83 changed files, +18,569 lines.

### Added

**Correctness foundation**
- `event.SubscribeWithHandle[T]` — returns a `*Subscription[T]` with a `Close()` method for explicit unsubscription. Prefer this in new code to avoid goroutine leaks from forgotten subscribers.
- `event.Publish` is non-blocking with bounded backpressure: events are dropped rather than blocking publishers when subscribers are slow. Track drops via `Bus.DroppedEvents()`.
- `KeyConsumer` interface — focused components can decline keys (e.g. `Tab`) so global shortcuts still work.
- `Bounded` interface — components expose their rendered geometry for hit-testing.
- `MouseMsg` routing now hit-tests against rendered components and re-focuses on click.
- Framebuffer line truncation and padding are display-width aware (east-asian width counted), not byte-length aware. Fixes width drift with CJK/wide characters.
- Cursor is hidden during framebuffer diff writes to prevent flicker.
- Baseline test coverage for the framebuffer (resize, clear, cursor ops) and selection manager.

**OpenTUI parity — system integration**
- `input.Slider` component.
- `FocusedMsg` / `BlurredMsg` — XTerm focus event tracking.
- `BracketedPasteMsg` — typed paste detection with content payload.
- `display.LineNumbers` gutter component.
- `CapabilityMsg` and `TerminalCapability` detection (truecolor, hyperlinks, etc.).
- `ThemeModeMsg` and light/dark theme-mode auto-detection.
- `selection.OSC52Sequence` and `selection.SupportsOSC52` — public OSC 52 clipboard helpers.

**OpenTUI parity — components**
- `markdown` package — glamour-backed terminal markdown renderer with theme-token bridging.
- `syntax` package — chroma-backed terminal syntax highlighter; optional truecolor output via `NewWithFormatter`.
- `animation` package — easings, tweens, timelines, and a `Tick` command.
- `headless` package — in-memory renderer for component testing without a real terminal.
- `Application.Use(InputHandler)` — composable input-handler chain for cross-cutting key/message processing.

**Other unreleased work consolidated in this tag**
- Codex-inspired components: markdown stream, commit-tick, history cell, exec cell, approval overlay, frame rate limiter, job control, network status.
- Claude-code-inspired components (two batches): VirtualMessageList, ToolUseBlock, ToolResultDisplay, GlimmerMessage, StatusLine, NoSelect, MessageBubble, TeammateSpinnerTree, BashPermissionView, FilePermissionView, BackgroundTaskPanel, MCPServerList, MCPToolDetail, ElicitationForm, AgentWizard, ModelPicker, ThemePicker, FuzzyPicker, RawAnsi.
- UI gap components: Divider, KeyboardShortcutHint, ListItem, LoadingState, StatusIcon, ThemedText, ThemedBox, Pane, OrderedList, ColorPicker, TreeSelect, DiffFileList, DiffDetailView, TokenWarning, MemoryUsageIndicator, ContextVisualization, ToolUseLoader.
- `FrameBuffer` renderer, virtualized `RecyclerView` / `VirtualList`.
- Diagram component with comprehensive README.
- Focus management, rope data structure, grapheme-aware text, scrollbar, text table.
- OpenTUI-inspired rendering optimizations: cell buffer, render pool, dirty tracking, alpha blending, hit grid.

### Changed

- **Source compatibility restored**: `event.Subscribe[T]` retains its `<-chan T` return type from `tui/v2.12.0..tui/v2.19.0`. The handle-returning variant introduced earlier in this branch is preserved as the new `event.SubscribeWithHandle[T]`. Callers of the legacy signature need no changes; callers wanting `Close()` should migrate to `SubscribeWithHandle`.

### Dependencies

- `github.com/SCKelemen/design-system` v1.0.2 → v1.3.0.
- `github.com/junegunn/fzf` v0.70.0 → v0.72.0.
- `github.com/SCKelemen/layout` v1.1.3 → v1.1.4 (a follow-up will bump to v1.2.0).
- `github.com/charmbracelet/bubbles` v0.21.0 → v1.0.0.
- **New**: `github.com/charmbracelet/glamour` (markdown).
- **New**: `github.com/alecthomas/chroma/v2` (syntax highlighting).

### CI
- `actions/checkout` v4 → v6.
- `actions/setup-go` v5 → v6.
- `softprops/action-gh-release` v2 → v3.

## [v2.2.0] - 2025-03-27
### Added
- Comprehensive test suite: 30 test files across all 10 packages
- 8 cross-package E2E integration tests
- Style package unit tests (theme resolution, ANSI parsing, constants)
- Coverage: 68-86% across all packages

## [v2.1.0] - 2025-03-27
### Added
- Regex-based syntax highlighter for 11 languages (Go, JS, TS, Python, Rust, SQL, Shell, JSON, YAML, Markdown, Plain)
- CodeBlock integration: WithCodeBlockLanguage, WithCodeBlockFilename
- design-system ↔ SCKelemen/color bridge: TokenToColor, ColorToANSIFg/Bg, ThemeAccentColor, ThemeGradient
- 4 dataviz terminal chart components: BarChart, LineChart, HeatMap, ScatterPlot
- ANSIBlue style constant

## [v2.0.0] - 2025-03-27
### Changed
- **BREAKING**: Module path changed to `github.com/SCKelemen/tui/v2`
- All components reorganized into sub-packages:
  - `agent/` — SubagentPanel, SubagentGroup, ThreadProgress
  - `chart/` — ProgressBar, Sparkline
  - `chat/` — ConversationView
  - `container/` — SplitPane, Modal
  - `display/` — CodeBlock, DiffBlock, ErrorFrame, FileCard, Toast, StatCard
  - `input/` — TextInput, CommandPalette, FloatingPalette, Checklist
  - `nav/` — TabBar, Breadcrumb, Tree, ScrollContainer
  - `selection/` — SelectionManager, Clipboard (OSC 52)
  - `style/` — ANSI constants, design token resolution
- Root package exports Component interface and Application type
- Shared ANSI/theme helpers centralized in style/ package

## [Unreleased] (v1.6.0)
### Added
- ProgressBar with 4 styles and SCKelemen/color gradient support
- Tree component with collapsible nodes, vim-style navigation, icons
- Sparkline inline chart using Unicode block characters
### Changed
- Design token wiring for Toast, ErrorFrame, SelectionManager
- Updated showcase demo

## [v1.5.0] - 2026-03-26
### Added
- SubagentPanel: bordered card with status indicators and tool tree
- SubagentGroup: horizontal equal-height panel layout with progress bar
- FileCard: inline diff stat badge with rounded box drawing
- ErrorFrame: large bordered error display
- FloatingPalette: centered overlay command palette with fuzzy filter

## [v1.4.0] - 2026-03-26
### Added
- ScrollContainer with scrollbar and child focus management
- Showcase demo application
- Comprehensive package README
### Changed
- Design system token expansion (typography, selection, toast)
- ConversationView WithAutoScroll renamed to WithConversationAutoScroll

## [v1.3.0] - 2026-03-26
### Added
- SelectionManager for mouse text selection
- OSC 52 clipboard support (Ghostty, iTerm2, kitty, etc.)
- Mouse selection wired into CodeBlock, DiffBlock, ConversationView

## [v1.2.0] - 2026-03-26
### Added
- ThreadProgress: animated thread status with icon sets
- Toast: timed notification with severity levels
- TabBar: switchable tab navigation
- SplitPane: horizontal/vertical split layout
- Breadcrumb: overflow-aware path breadcrumbs
- Checklist: checkable item list
- ConversationView: chat message display with auto-scroll
