# Changelog

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
