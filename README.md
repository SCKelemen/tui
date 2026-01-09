# tui

A comprehensive Terminal User Interface framework for building Claude Code-like CLI experiences.

## Overview

`tui` is a high-level framework built on top of the SCKelemen visualization stack, providing ready-to-use components for building sophisticated terminal applications with modern UX patterns.

## Features

- **Rich Components**: File explorers, command palettes, status bars, modals, tabs
- **Keyboard Navigation**: Vim-like bindings with customizable keymaps
- **Mouse Support**: Click, scroll, drag interactions
- **Focus Management**: Intuitive focus flow between components
- **Layout System**: Flexbox and grid layouts via `layout` package
- **Theme Support**: Full design token integration via `design-system`
- **Unicode Aware**: Proper handling of emoji, wide characters via `text`
- **Color Science**: Perceptually uniform gradients via `color` (OKLCH)

## Architecture

```
tui (high-level components)
 ├── cli (terminal rendering)
 ├── layout (flexbox/grid)
 ├── design-system (themes)
 ├── text (unicode width)
 └── color (OKLCH gradients)
```

## Installation

```bash
go get github.com/SCKelemen/tui@latest
```

## Quick Start

```go
package main

import (
    "github.com/SCKelemen/tui"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    app := tui.NewApplication()

    // Add components
    fileExplorer := tui.NewFileExplorer("/path/to/project")
    statusBar := tui.NewStatusBar()

    app.AddComponent(fileExplorer)
    app.AddComponent(statusBar)

    // Run
    p := tea.NewProgram(app)
    if _, err := p.Run(); err != nil {
        panic(err)
    }
}
```

## Components

### FileExplorer
Tree view with navigation, search, and file operations.

### CommandPalette
Fuzzy-searchable command launcher.

### StatusBar
Bottom status bar with context-aware keybindings.

### SidePanel
Collapsible side panels with sections.

### Editor
Text viewing/editing with syntax highlighting support.

### Modal
Dialog boxes for confirmations and inputs.

### Tabs
Multi-view tab management.

### SearchResults
Searchable result lists with context.

## Roadmap

- [ ] Core application framework
- [ ] FileExplorer component
- [ ] StatusBar component
- [ ] CommandPalette component
- [ ] Focus management system
- [ ] Keyboard binding system
- [ ] Theme customization
- [ ] Mouse event handling
- [ ] Window splitting
- [ ] Plugin system

## License

Bearware 1.0

## Related Projects

- [cli](https://github.com/SCKelemen/cli) - Low-level terminal rendering
- [layout](https://github.com/SCKelemen/layout) - CSS-like layout engine
- [design-system](https://github.com/SCKelemen/design-system) - Design tokens and themes
- [dataviz](https://github.com/SCKelemen/dataviz) - Data visualization components
