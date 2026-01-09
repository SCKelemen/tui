# File Explorer Demo

This demo showcases the FileExplorer component for browsing the file system with tree navigation.

## Features Demonstrated

- **Tree-based file navigation** with expand/collapse
- **Keyboard navigation** (vim-style or arrow keys)
- **Lazy loading** - directories load children on demand
- **Hidden files toggle** - press `.` to show/hide files starting with `.`
- **Visual indicators**:
  - 📁 Collapsed directory
  - 📂 Expanded directory
  - 📄 File
- **Scroll handling** for long directory listings
- **Focus states** - inverted highlight when focused, dimmed arrow when unfocused
- **Depth indentation** - visual tree structure with connectors
- **Status bar integration** - shows currently selected path

## Running the Demo

```bash
go run main.go
```

## Keyboard Controls

### Navigation
- **↑/k** - Move selection up
- **↓/j** - Move selection down
- **→/l or Enter** - Expand directory (if collapsed)
- **←/h** - Collapse directory (if expanded) or move to parent
- **q** - Quit

### Actions
- **.** - Toggle hidden files on/off
- **r** - Refresh current directory

## Implementation Details

### Tree Structure

Each file or directory is represented by a `FileNode`:

```go
type FileNode struct {
    Name     string
    Path     string
    IsDir    bool
    Children []*FileNode
    Expanded bool
    Parent   *FileNode
}
```

### Lazy Loading

Directories don't load their children until expanded for performance:

```go
func (fe *FileExplorer) expand() {
    if fe.selected.IsDir && !fe.selected.Expanded {
        if len(fe.selected.Children) == 0 {
            // Load children on first expand
            children := fe.loadChildren(fe.selected.Path, fe.selected)
            fe.selected.Children = children
        }
        fe.selected.Expanded = true
        fe.updateVisibleNodes()
    }
}
```

### Visible Nodes

Only expanded directories and their visible descendants are rendered:

```go
func (fe *FileExplorer) collectVisibleNodes(node *FileNode) {
    fe.visibleNodes = append(fe.visibleNodes, node)

    if node.IsDir && node.Expanded {
        for _, child := range node.Children {
            fe.collectVisibleNodes(child)
        }
    }
}
```

### Scrolling

The component automatically scrolls to keep the selected item visible:

```go
if fe.selectedIndex < fe.scrollOffset {
    fe.scrollOffset = fe.selectedIndex
}
if fe.selectedIndex >= fe.scrollOffset+maxVisible {
    fe.scrollOffset = fe.selectedIndex - maxVisible + 1
}
```

## Usage in Your Application

```go
// Create file explorer
explorer := tui.NewFileExplorer("/path/to/directory",
    tui.WithShowHidden(false))

// Add to application
app.AddComponent(explorer)

// Get selected path
path := explorer.GetSelectedPath()

// Get selected node
node := explorer.GetSelectedNode()
if node != nil {
    fmt.Printf("Selected: %s (IsDir: %v)\n", node.Name, node.IsDir)
}
```

## Customization

### Show Hidden Files by Default

```go
explorer := tui.NewFileExplorer("/path", tui.WithShowHidden(true))
```

### Start at Specific Directory

```go
explorer := tui.NewFileExplorer("/home/user/projects")
```

## Future Enhancements

Potential improvements for the FileExplorer component:

- File type icons based on extension
- File size display
- Last modified date
- Multi-selection support
- File operations (copy, move, delete)
- Search/filter functionality
- Bookmarks/favorites
- Git status indicators
- Custom sorting options
