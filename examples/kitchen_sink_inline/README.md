# Kitchen Sink Inline Mode

This demo showcases TUI components in **inline/history mode**, where content persists in your terminal after the program exits (like Codex CLI).

## Key Differences from Standard Kitchen Sink

### Inline Mode (this demo)
- **No `tea.WithAltScreen()`** - content flows naturally in terminal
- **Content persists** - scroll up in your terminal to see previous output
- **No viewport needed** - terminal handles scrolling natively
- **No modal/palette overlays** - these require alt screen buffer
- **Simpler interaction** - focus on core component rendering

### Full-Screen Mode (`kitchen_sink`)
- **Uses `tea.WithAltScreen()`** - creates separate screen buffer
- **Content disappears on exit** - clean terminal state
- **Viewport scrolling** - built-in scroll with indicators
- **Full overlays** - modals and command palettes work
- **Complex interactions** - section switching, multiple input modes

## Usage

```bash
go run main.go
go run main.go --theme=claude
```

The demo will:
1. Display all components with live animations
2. Update status indicators every 3 seconds
3. Allow you to start/stop activity bar with 'r' and 's'
4. Quit with 'q' or Ctrl+C
5. **Leave all output in your terminal** - scroll up to review

## When to Use Inline Mode

Use inline mode when:
- You want output to persist in terminal history
- Building CLI tools that feel like regular commands
- Users need to scroll back through results
- You're creating log viewers or streaming output displays
- Content should be copyable from terminal scrollback

## When to Use Full-Screen Mode

Use full-screen mode when:
- Building interactive applications (editors, dashboards)
- You need modal dialogs or overlays
- Clean exit is important (don't clutter terminal)
- Content is meant to be ephemeral
- Building TUI apps rather than CLI tools
