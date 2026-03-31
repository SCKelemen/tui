package input

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// TreeNode is one selectable node in a TreeSelect.
type TreeNode struct {
	ID       string
	Label    string
	Children []TreeNode
	Disabled bool
}

// TreeSelectMsg is emitted when a node is selected with Enter.
type TreeSelectMsg struct {
	Node TreeNode
}

// TreeSelectOption configures a TreeSelect.
type TreeSelectOption func(*TreeSelect)

// WithTreeSelectWidth sets preferred render width.
func WithTreeSelectWidth(width int) TreeSelectOption {
	return func(t *TreeSelect) {
		if width > 0 {
			t.width = width
		}
	}
}

// WithTreeSelectDesignTokens applies design-system tokens.
func WithTreeSelectDesignTokens(tokens *design.DesignTokens) TreeSelectOption {
	return func(t *TreeSelect) {
		if tokens != nil {
			t.designTokens = tokens
		}
	}
}

type treeSelectVisibleNode struct {
	node  TreeNode
	depth int
}

// TreeSelect renders a collapsible tree selector.
type TreeSelect struct {
	roots        []TreeNode
	expanded     map[string]bool
	cursor       int
	focused      bool
	width        int
	windowWidth  int
	designTokens *design.DesignTokens
}

// NewTreeSelect creates a new TreeSelect.
func NewTreeSelect(roots []TreeNode, opts ...TreeSelectOption) *TreeSelect {
	t := &TreeSelect{
		roots:        append([]TreeNode(nil), roots...),
		expanded:     make(map[string]bool),
		cursor:       0,
		focused:      false,
		width:        0,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(t)
	}

	for _, root := range t.roots {
		t.expanded[t.nodeID(root)] = true
	}
	return t
}

// Init initializes the component.
func (t *TreeSelect) Init() tea.Cmd { return nil }

// Update handles keyboard navigation and selection.
func (t *TreeSelect) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.windowWidth = msg.Width
		return t, nil
	case tea.KeyMsg:
		if !t.focused {
			return t, nil
		}

		visible := t.visibleNodes()
		if len(visible) == 0 {
			return t, nil
		}
		t.clampCursor(len(visible))

		switch msg.String() {
		case "up", "k":
			if t.cursor > 0 {
				t.cursor--
			}
			return t, nil
		case "down", "j":
			if t.cursor < len(visible)-1 {
				t.cursor++
			}
			return t, nil
		case "right", "l":
			curr := visible[t.cursor].node
			if len(curr.Children) > 0 {
				t.expanded[t.nodeID(curr)] = true
			}
			return t, nil
		case "left", "h":
			curr := visible[t.cursor].node
			id := t.nodeID(curr)
			if len(curr.Children) > 0 && t.expanded[id] {
				t.expanded[id] = false
			}
			return t, nil
		case "enter":
			curr := visible[t.cursor].node
			if curr.Disabled {
				return t, nil
			}
			return t, func() tea.Msg {
				return TreeSelectMsg{Node: curr}
			}
		}
	}

	return t, nil
}

// View renders the tree with indentation and expand/collapse markers.
func (t *TreeSelect) View() string {
	visible := t.visibleNodes()
	if len(visible) == 0 {
		return ""
	}
	t.clampCursor(len(visible))

	width := t.effectiveWidth()
	lines := make([]string, 0, len(visible))

	for i, vn := range visible {
		line := t.renderNodeLine(vn)
		if width > 0 {
			line = style.Pad(style.Truncate(line, width, "…"), width)
		}
		if i == t.cursor {
			line = style.ANSIInverse + line + style.ANSIReset
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// Focus marks component focused.
func (t *TreeSelect) Focus() { t.focused = true }

// Blur marks component unfocused.
func (t *TreeSelect) Blur() { t.focused = false }

// Focused reports whether component is focused.
func (t *TreeSelect) Focused() bool { return t.focused }

func (t *TreeSelect) visibleNodes() []treeSelectVisibleNode {
	out := make([]treeSelectVisibleNode, 0)
	for _, root := range t.roots {
		t.walkVisible(root, 0, &out)
	}
	return out
}

func (t *TreeSelect) walkVisible(node TreeNode, depth int, out *[]treeSelectVisibleNode) {
	*out = append(*out, treeSelectVisibleNode{node: node, depth: depth})
	if len(node.Children) == 0 {
		return
	}
	if !t.expanded[t.nodeID(node)] {
		return
	}
	for _, child := range node.Children {
		t.walkVisible(child, depth+1, out)
	}
}

func (t *TreeSelect) renderNodeLine(v treeSelectVisibleNode) string {
	indent := strings.Repeat("  ", v.depth)
	marker := "  "
	if len(v.node.Children) > 0 {
		if t.expanded[t.nodeID(v.node)] {
			marker = "▾ "
		} else {
			marker = "▸ "
		}
	}
	label := v.node.Label
	if strings.TrimSpace(label) == "" {
		label = v.node.ID
	}
	if v.node.Disabled {
		label = style.ANSIDim + label + style.ANSIReset
	}
	return indent + marker + label
}

func (t *TreeSelect) effectiveWidth() int {
	if t.width > 0 {
		return t.width
	}
	return t.windowWidth
}

func (t *TreeSelect) clampCursor(total int) {
	if total <= 0 {
		t.cursor = 0
		return
	}
	if t.cursor < 0 {
		t.cursor = 0
	}
	if t.cursor >= total {
		t.cursor = total - 1
	}
}

func (t *TreeSelect) nodeID(node TreeNode) string {
	if strings.TrimSpace(node.ID) != "" {
		return node.ID
	}
	if strings.TrimSpace(node.Label) != "" {
		return node.Label
	}
	return "node"
}

var _ tui.Component = (*TreeSelect)(nil)
