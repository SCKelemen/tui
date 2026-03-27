package nav

import (
	"strings"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// TreeNode represents a node in a collapsible tree.
type TreeNode struct {
	Label    string
	Icon     string
	Children []*TreeNode
	Expanded bool
	Data     interface{}
}

// TreeOption configures a tree component.
type TreeOption func(*tree)

type tree struct {
	roots        []*TreeNode
	cursor       int
	focused      bool
	showIcons    bool
	showGuides   bool
	indentSize   int
	designTokens interface{}
	theme        string
	width        int
}

type visibleTreeNode struct {
	node          *TreeNode
	depth         int
	ancestorIsEnd []bool
	isEnd         bool
}

// WithTreeShowIcons configures whether node icons are rendered.
func WithTreeShowIcons(show bool) TreeOption {
	return func(t *tree) {
		t.showIcons = show
	}
}

// WithTreeShowGuides configures whether tree guides are rendered.
func WithTreeShowGuides(show bool) TreeOption {
	return func(t *tree) {
		t.showGuides = show
	}
}

// WithTreeIndentSize configures the indentation size.
func WithTreeIndentSize(size int) TreeOption {
	return func(t *tree) {
		if size > 0 {
			t.indentSize = size
		}
	}
}

// WithTreeDesignTokens configures design tokens for future style integrations.
func WithTreeDesignTokens(tokens interface{}) TreeOption {
	return func(t *tree) {
		t.designTokens = tokens
	}
}

// WithTreeTheme configures a named theme for future style integrations.
func WithTreeTheme(theme string) TreeOption {
	return func(t *tree) {
		t.theme = theme
		t.designTokens = style.DesignTokensForTheme(theme)
	}
}

// NewTree creates a standalone collapsible, navigable tree component.
func NewTree(roots []*TreeNode, opts ...TreeOption) *tree {
	t := &tree{
		roots:      make([]*TreeNode, 0),
		cursor:     0,
		showIcons:  true,
		showGuides: true,
		indentSize: 2,
	}

	for _, opt := range opts {
		opt(t)
	}

	t.SetRoots(roots)
	return t
}

// Init initializes the tree component.
func (t *tree) Init() tea.Cmd {
	return nil
}

// Update handles key navigation and expand/collapse interactions.
func (t *tree) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			t.moveCursor(1)
		case "k", "up":
			t.moveCursor(-1)
		case "enter", " ", "right":
			t.ToggleSelected()
		case "left", "h":
			t.collapseSelected()
		case "l":
			t.expandSelected()
		}
	}

	return t, nil
}

// View renders the tree.
func (t *tree) View() string {
	visible := t.visibleNodes()
	if len(visible) == 0 {
		return ""
	}

	lines := make([]string, 0, len(visible))
	for i, vn := range visible {
		line := t.renderNode(vn)
		if i == t.cursor {
			if t.focused {
				line = style.ANSIInverse + line + style.ANSIReset
			} else {
				line = style.ANSIBold + line + style.ANSIReset
			}
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// Focus marks the component as focused.
func (t *tree) Focus() {
	t.focused = true
}

// Blur marks the component as unfocused.
func (t *tree) Blur() {
	t.focused = false
}

// Focused reports whether the component is focused.
func (t *tree) Focused() bool {
	return t.focused
}

// SetRoots replaces tree roots and resets cursor to a valid position.
func (t *tree) SetRoots(roots []*TreeNode) {
	t.roots = roots
	for _, root := range t.roots {
		root.Expanded = true
	}
	t.ensureCursorBounds()
}

// GetRoots returns the current root nodes.
func (t *tree) GetRoots() []*TreeNode {
	return t.roots
}

// GetSelectedNode returns the currently selected visible node.
func (t *tree) GetSelectedNode() *TreeNode {
	visible := t.visibleNodes()
	if len(visible) == 0 || t.cursor < 0 || t.cursor >= len(visible) {
		return nil
	}
	return visible[t.cursor].node
}

// ExpandAll expands all nodes recursively.
func (t *tree) ExpandAll() {
	for _, root := range t.roots {
		t.expandRecursive(root)
	}
	t.ensureCursorBounds()
}

// CollapseAll collapses all nodes recursively.
func (t *tree) CollapseAll() {
	for _, root := range t.roots {
		t.collapseRecursive(root)
	}
	t.ensureCursorBounds()
}

// ToggleSelected toggles the expanded state of the selected node.
func (t *tree) ToggleSelected() {
	n := t.GetSelectedNode()
	if n == nil || len(n.Children) == 0 {
		return
	}
	n.Expanded = !n.Expanded
	t.ensureCursorBounds()
}

func (t *tree) moveCursor(delta int) {
	visible := t.visibleNodes()
	if len(visible) == 0 {
		t.cursor = 0
		return
	}

	t.cursor += delta
	if t.cursor < 0 {
		t.cursor = 0
	}
	if t.cursor >= len(visible) {
		t.cursor = len(visible) - 1
	}
}

func (t *tree) collapseSelected() {
	n := t.GetSelectedNode()
	if n == nil || len(n.Children) == 0 {
		return
	}
	n.Expanded = false
	t.ensureCursorBounds()
}

func (t *tree) expandSelected() {
	n := t.GetSelectedNode()
	if n == nil || len(n.Children) == 0 {
		return
	}
	n.Expanded = true
	t.ensureCursorBounds()
}

func (t *tree) ensureCursorBounds() {
	visible := t.visibleNodes()
	if len(visible) == 0 {
		t.cursor = 0
		return
	}
	if t.cursor < 0 {
		t.cursor = 0
	}
	if t.cursor >= len(visible) {
		t.cursor = len(visible) - 1
	}
}

func (t *tree) visibleNodes() []visibleTreeNode {
	out := make([]visibleTreeNode, 0)
	for i, root := range t.roots {
		isLast := i == len(t.roots)-1
		t.walkVisible(root, 0, nil, isLast, &out)
	}
	return out
}

func (t *tree) walkVisible(node *TreeNode, depth int, ancestorIsEnd []bool, isEnd bool, out *[]visibleTreeNode) {
	if node == nil {
		return
	}

	pathCopy := append([]bool(nil), ancestorIsEnd...)
	*out = append(*out, visibleTreeNode{
		node:          node,
		depth:         depth,
		ancestorIsEnd: pathCopy,
		isEnd:         isEnd,
	})

	if len(node.Children) == 0 || !node.Expanded {
		return
	}

	childAncestors := append(pathCopy, isEnd)
	for i, child := range node.Children {
		childIsEnd := i == len(node.Children)-1
		t.walkVisible(child, depth+1, childAncestors, childIsEnd, out)
	}
}

func (t *tree) renderNode(vn visibleTreeNode) string {
	var b strings.Builder

	if t.showGuides {
		for _, ancestorEnd := range vn.ancestorIsEnd {
			if ancestorEnd {
				b.WriteString(strings.Repeat(" ", t.indentSize))
			} else {
				if t.indentSize > 1 {
					b.WriteString("│" + strings.Repeat(" ", t.indentSize-1))
				} else {
					b.WriteString("│")
				}
			}
		}

		if vn.depth > 0 {
			if vn.isEnd {
				b.WriteString("└")
			} else {
				b.WriteString("├")
			}
			if t.indentSize > 1 {
				b.WriteString(strings.Repeat("─", t.indentSize-1))
			}
		}
	} else if vn.depth > 0 {
		b.WriteString(strings.Repeat(" ", vn.depth*t.indentSize))
	}

	if len(vn.node.Children) > 0 {
		if vn.node.Expanded {
			b.WriteString("▾ ")
		} else {
			b.WriteString("▸ ")
		}
	} else {
		b.WriteString("  ")
	}

	label := vn.node.Label
	if t.showIcons && vn.node.Icon != "" {
		label = vn.node.Icon + " " + label
	}

	b.WriteString(label)
	return b.String()
}

func (t *tree) expandRecursive(node *TreeNode) {
	if node == nil {
		return
	}
	if len(node.Children) > 0 {
		node.Expanded = true
	}
	for _, child := range node.Children {
		t.expandRecursive(child)
	}
}

func (t *tree) collapseRecursive(node *TreeNode) {
	if node == nil {
		return
	}
	if len(node.Children) > 0 {
		node.Expanded = false
	}
	for _, child := range node.Children {
		t.collapseRecursive(child)
	}
}
