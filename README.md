> ⚠️ **This module (`github.com/SCKelemen/tui`, v1) is deprecated.**
>
> All development has moved to **[`github.com/SCKelemen/tui/v2`](./v2)**, a clean rewrite around Bubble Tea, [`SCKelemen/text`](https://github.com/SCKelemen/text) (UAX #29 grapheme handling), and [`SCKelemen/layout`](https://github.com/SCKelemen/layout).
>
> The last fully-supported v1 release is **v1.6.0** — the tag remains valid for module proxy fetches. Migrate by changing your import path:
>
> ```go
> import tui "github.com/SCKelemen/tui/v2"
> ```

# tui

Terminal UI component library built on Bubble Tea with design token theming.

## Migration to v2

v2 provides the same components reorganized into focused sub-packages:
`agent`, `chart`, `chat`, `container`, `display`, `input`, `nav`, `selection`, `style`

See the [v2 README](./v2/README.md) for the full API reference and migration guide.

### Quick migration

```diff
- import "github.com/SCKelemen/tui"
+ import tui "github.com/SCKelemen/tui/v2"
```

v2 has a substantially different API around lifecycle, the event bus, and layout integration. No drop-in compatibility shim is provided.

### Pinning to v1

If you need to stay on v1 temporarily:

```bash
go get github.com/SCKelemen/tui@v1.6.0
```

The `v1.6.0` tag remains valid for module proxy fetches. The git history preserves all v1 source code.

## What was in v1

v1 shipped 25+ interactive components including:

- **Layout**: Application, SplitPane, ScrollContainer, Dashboard, Header
- **Data Display**: StatCard, StructuredData, Table (`tui/table`), CodeBlock, DiffBlock
- **Navigation**: TabBar, Breadcrumb, CommandPalette, FileExplorer
- **Input**: TextInput, Modal, Checklist, FloatingPalette
- **Feedback**: ActivityBar, StatusBar, Toast, ToolBlock, ThreadProgress, Spinner
- **Communication**: ConversationView
- **Utilities**: SelectionManager, Clipboard (OSC 52)

All of these have v2 equivalents — see the [v2 README](./v2/README.md).

## License

Bearware 1.0
