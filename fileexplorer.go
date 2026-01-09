package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// FileNode represents a file or directory in the tree
type FileNode struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*FileNode
	Expanded bool
	Parent   *FileNode
}

// FileExplorer displays a navigable file tree
type FileExplorer struct {
	width         int
	height        int
	root          *FileNode
	selected      *FileNode
	visibleNodes  []*FileNode
	selectedIndex int
	scrollOffset  int
	focused       bool
	showHidden    bool
	basePath      string
}

// FileExplorerOption configures a FileExplorer
type FileExplorerOption func(*FileExplorer)

// WithShowHidden shows hidden files (starting with .)
func WithShowHidden(show bool) FileExplorerOption {
	return func(fe *FileExplorer) {
		fe.showHidden = show
	}
}

// NewFileExplorer creates a new file explorer starting at the given path
func NewFileExplorer(path string, opts ...FileExplorerOption) *FileExplorer {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	fe := &FileExplorer{
		basePath:   absPath,
		showHidden: false,
		height:     20, // Default height
	}

	for _, opt := range opts {
		opt(fe)
	}

	// Build initial tree
	fe.root = fe.buildTree(absPath, nil)
	fe.root.Expanded = true // Root is always expanded
	fe.updateVisibleNodes()
	if len(fe.visibleNodes) > 0 {
		fe.selected = fe.visibleNodes[0]
		fe.selectedIndex = 0
	}

	return fe
}

// Init initializes the file explorer
func (fe *FileExplorer) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (fe *FileExplorer) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		fe.width = msg.Width
		fe.height = msg.Height

	case tea.KeyMsg:
		if !fe.focused {
			return fe, nil
		}

		switch msg.String() {
		case "up", "k":
			fe.moveUp()
		case "down", "j":
			fe.moveDown()
		case "left", "h":
			fe.collapse()
		case "right", "l", "enter":
			fe.expand()
		case ".":
			fe.showHidden = !fe.showHidden
			fe.refresh()
		case "r":
			fe.refresh()
		}
	}

	return fe, nil
}

// View renders the file explorer
func (fe *FileExplorer) View() string {
	if fe.width == 0 {
		return ""
	}

	var b strings.Builder

	// Header with current path
	header := fmt.Sprintf("\033[1m📁 %s\033[0m", fe.basePath)
	if len(header) > fe.width {
		header = fmt.Sprintf("\033[1m📁 ...%s\033[0m", fe.basePath[len(fe.basePath)-(fe.width-10):])
	}
	b.WriteString(header)
	b.WriteString("\n")

	// Calculate visible range
	maxVisible := fe.height - 3 // Account for header and hints
	if maxVisible < 1 {
		maxVisible = 1
	}

	// Adjust scroll offset if needed
	if fe.selectedIndex < fe.scrollOffset {
		fe.scrollOffset = fe.selectedIndex
	}
	if fe.selectedIndex >= fe.scrollOffset+maxVisible {
		fe.scrollOffset = fe.selectedIndex - maxVisible + 1
	}

	// Render visible nodes
	end := fe.scrollOffset + maxVisible
	if end > len(fe.visibleNodes) {
		end = len(fe.visibleNodes)
	}

	for i := fe.scrollOffset; i < end; i++ {
		node := fe.visibleNodes[i]
		isSelected := node == fe.selected

		// Indent based on depth
		depth := fe.getDepth(node)
		indent := strings.Repeat("  ", depth)

		// Tree lines
		var connector string
		if depth > 0 {
			connector = "├─ "
			// TODO: Use └─ for last child
		}

		// Icon
		icon := "📄"
		if node.IsDir {
			if node.Expanded {
				icon = "📂"
			} else {
				icon = "📁"
			}
		}

		// Build line
		line := fmt.Sprintf("%s%s%s %s", indent, connector, icon, node.Name)

		// Highlight if selected
		if isSelected {
			if fe.focused {
				line = fmt.Sprintf("\033[7m%s\033[0m", line) // Inverted
			} else {
				line = fmt.Sprintf("\033[2m▸ %s\033[0m", line) // Dimmed with arrow
			}
		} else {
			line = "  " + line
		}

		// Truncate if too long
		if len(stripANSI(line)) > fe.width {
			line = truncateANSI(line, fe.width-3) + "..."
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(fe.visibleNodes) > maxVisible {
		b.WriteString(fmt.Sprintf("\033[2m[%d/%d]\033[0m\n", fe.selectedIndex+1, len(fe.visibleNodes)))
	}

	// Hints
	if fe.focused {
		b.WriteString("\033[2m↑↓: navigate · Enter: open · .: toggle hidden · r: refresh\033[0m")
	}

	return b.String()
}

// Focus is called when this component receives focus
func (fe *FileExplorer) Focus() {
	fe.focused = true
}

// Blur is called when this component loses focus
func (fe *FileExplorer) Blur() {
	fe.focused = false
}

// Focused returns whether this component is currently focused
func (fe *FileExplorer) Focused() bool {
	return fe.focused
}

// GetSelectedPath returns the path of the currently selected node
func (fe *FileExplorer) GetSelectedPath() string {
	if fe.selected != nil {
		return fe.selected.Path
	}
	return ""
}

// GetSelectedNode returns the currently selected node
func (fe *FileExplorer) GetSelectedNode() *FileNode {
	return fe.selected
}

// moveUp moves selection up
func (fe *FileExplorer) moveUp() {
	if fe.selectedIndex > 0 {
		fe.selectedIndex--
		fe.selected = fe.visibleNodes[fe.selectedIndex]
	}
}

// moveDown moves selection down
func (fe *FileExplorer) moveDown() {
	if fe.selectedIndex < len(fe.visibleNodes)-1 {
		fe.selectedIndex++
		fe.selected = fe.visibleNodes[fe.selectedIndex]
	}
}

// expand expands a directory or opens a file
func (fe *FileExplorer) expand() {
	if fe.selected == nil {
		return
	}

	if fe.selected.IsDir {
		if !fe.selected.Expanded {
			// Load children if not already loaded
			if len(fe.selected.Children) == 0 {
				children := fe.loadChildren(fe.selected.Path, fe.selected)
				fe.selected.Children = children
			}
			fe.selected.Expanded = true
			fe.updateVisibleNodes()
		}
	}
}

// collapse collapses a directory or moves to parent
func (fe *FileExplorer) collapse() {
	if fe.selected == nil {
		return
	}

	if fe.selected.IsDir && fe.selected.Expanded {
		fe.selected.Expanded = false
		fe.updateVisibleNodes()
	} else if fe.selected.Parent != nil {
		// Move to parent
		for i, node := range fe.visibleNodes {
			if node == fe.selected.Parent {
				fe.selectedIndex = i
				fe.selected = node
				break
			}
		}
	}
}

// refresh reloads the current directory
func (fe *FileExplorer) refresh() {
	selectedPath := ""
	if fe.selected != nil {
		selectedPath = fe.selected.Path
	}

	fe.root = fe.buildTree(fe.basePath, nil)
	fe.root.Expanded = true
	fe.updateVisibleNodes()

	// Try to restore selection
	if selectedPath != "" {
		for i, node := range fe.visibleNodes {
			if node.Path == selectedPath {
				fe.selectedIndex = i
				fe.selected = node
				return
			}
		}
	}

	// Default to first node
	if len(fe.visibleNodes) > 0 {
		fe.selectedIndex = 0
		fe.selected = fe.visibleNodes[0]
	}
}

// buildTree builds a file tree starting at path
func (fe *FileExplorer) buildTree(path string, parent *FileNode) *FileNode {
	info, err := os.Stat(path)
	if err != nil {
		return &FileNode{
			Name:   filepath.Base(path),
			Path:   path,
			IsDir:  false,
			Parent: parent,
		}
	}

	node := &FileNode{
		Name:   filepath.Base(path),
		Path:   path,
		IsDir:  info.IsDir(),
		Parent: parent,
	}

	// Don't load children initially (lazy load on expand)
	return node
}

// loadChildren loads child nodes for a directory
func (fe *FileExplorer) loadChildren(path string, parent *FileNode) []*FileNode {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	var children []*FileNode
	for _, entry := range entries {
		// Skip hidden files if not showing
		if !fe.showHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		childPath := filepath.Join(path, entry.Name())
		child := &FileNode{
			Name:   entry.Name(),
			Path:   childPath,
			IsDir:  entry.IsDir(),
			Parent: parent,
		}
		children = append(children, child)
	}

	// Sort: directories first, then alphabetically
	sort.Slice(children, func(i, j int) bool {
		if children[i].IsDir != children[j].IsDir {
			return children[i].IsDir
		}
		return children[i].Name < children[j].Name
	})

	return children
}

// updateVisibleNodes updates the list of visible nodes based on expansion state
func (fe *FileExplorer) updateVisibleNodes() {
	fe.visibleNodes = nil
	fe.collectVisibleNodes(fe.root)
}

// collectVisibleNodes recursively collects visible nodes
func (fe *FileExplorer) collectVisibleNodes(node *FileNode) {
	if node == nil {
		return
	}

	fe.visibleNodes = append(fe.visibleNodes, node)

	if node.IsDir && node.Expanded {
		// Ensure children are loaded
		if len(node.Children) == 0 {
			node.Children = fe.loadChildren(node.Path, node)
		}

		for _, child := range node.Children {
			fe.collectVisibleNodes(child)
		}
	}
}

// getDepth returns the depth of a node in the tree
func (fe *FileExplorer) getDepth(node *FileNode) int {
	depth := 0
	current := node
	for current.Parent != nil {
		depth++
		current = current.Parent
	}
	// Don't count the root, but ensure we never return negative
	if depth > 0 {
		return depth - 1
	}
	return 0
}
