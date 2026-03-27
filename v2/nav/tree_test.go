package nav

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTreeCreationAndOptions(t *testing.T) {
	roots := []*TreeNode{{Label: "root", Children: []*TreeNode{{Label: "child"}}}}
	tree := NewTree(
		roots,
		WithTreeShowIcons(false),
		WithTreeShowGuides(false),
		WithTreeIndentSize(4),
		WithTreeDesignTokens("tokens"),
		WithTreeTheme("dark"),
	)

	if tree == nil {
		t.Fatal("expected tree to be created")
	}
	if tree.showIcons {
		t.Fatal("expected showIcons to be false")
	}
	if tree.showGuides {
		t.Fatal("expected showGuides to be false")
	}
	if tree.indentSize != 4 {
		t.Fatalf("expected indent size 4, got %d", tree.indentSize)
	}
	if tree.designTokens == nil {
		t.Fatalf("expected design tokens to be set, got %#v", tree.designTokens)
	}
	if tree.theme != "dark" {
		t.Fatalf("expected theme dark, got %q", tree.theme)
	}
	if !roots[0].Expanded {
		t.Fatal("expected root node to be expanded by default")
	}
}

func TestTreeNavigation(t *testing.T) {
	tree := NewTree(sampleTree())

	if got := tree.GetSelectedNode(); got == nil || got.Label != "root" {
		t.Fatalf("expected initial selection root, got %#v", got)
	}

	tree.Update(tea.KeyMsg{Type: tea.KeyDown})
	if got := tree.GetSelectedNode(); got == nil || got.Label != "child-a" {
		t.Fatalf("expected selection child-a after down, got %#v", got)
	}

	tree.Update(tea.KeyMsg{Type: tea.KeyUp})
	if got := tree.GetSelectedNode(); got == nil || got.Label != "root" {
		t.Fatalf("expected selection root after up, got %#v", got)
	}

	tree.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if got := tree.GetSelectedNode(); got == nil || got.Label != "root" {
		t.Fatalf("expected selection to stay at top, got %#v", got)
	}
}

func TestTreeToggleAndExpandCollapseKeys(t *testing.T) {
	tree := NewTree(sampleTree())
	root := tree.GetSelectedNode()
	if root == nil {
		t.Fatal("expected selected root node")
	}

	if !root.Expanded {
		t.Fatal("expected root expanded initially")
	}

	tree.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if root.Expanded {
		t.Fatal("expected enter to toggle root collapsed")
	}

	tree.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if !root.Expanded {
		t.Fatal("expected l to expand root")
	}

	tree.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if root.Expanded {
		t.Fatal("expected left to collapse root")
	}

	tree.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if root.Expanded {
		t.Fatal("expected h to keep root collapsed")
	}

	tree.Update(tea.KeyMsg{Type: tea.KeyRight})
	if !root.Expanded {
		t.Fatal("expected right to toggle root expanded")
	}

	tree.Update(tea.KeyMsg{Type: tea.KeySpace})
	if root.Expanded {
		t.Fatal("expected space to toggle root collapsed")
	}
}

func TestTreeGuidesAndNestedDepthRendering(t *testing.T) {
	roots := []*TreeNode{
		{
			Label:    "root",
			Expanded: true,
			Children: []*TreeNode{
				{
					Label:    "child-a",
					Expanded: true,
					Children: []*TreeNode{{Label: "grandchild"}},
				},
				{Label: "child-b"},
			},
		},
	}

	tree := NewTree(roots)
	view := tree.View()

	for _, token := range []string{"│", "├", "└", "▾", "grandchild"} {
		if !strings.Contains(view, token) {
			t.Fatalf("expected view to contain %q, got:\n%s", token, view)
		}
	}
}

func TestTreeIconsRendering(t *testing.T) {
	roots := []*TreeNode{
		{
			Label:    "root",
			Icon:     "📁",
			Expanded: true,
			Children: []*TreeNode{{Label: "file", Icon: "📄"}},
		},
	}

	withIcons := NewTree(roots, WithTreeShowIcons(true))
	viewWithIcons := withIcons.View()
	if !strings.Contains(viewWithIcons, "📁") || !strings.Contains(viewWithIcons, "📄") {
		t.Fatalf("expected icons in view, got:\n%s", viewWithIcons)
	}

	withoutIcons := NewTree(roots, WithTreeShowIcons(false))
	viewWithoutIcons := withoutIcons.View()
	if strings.Contains(viewWithoutIcons, "📁") || strings.Contains(viewWithoutIcons, "📄") {
		t.Fatalf("expected no icons in view, got:\n%s", viewWithoutIcons)
	}
}

func TestTreeExpandAllAndCollapseAll(t *testing.T) {
	tree := NewTree(sampleTree())

	tree.CollapseAll()
	collapsedLines := lineCount(tree.View())
	if collapsedLines != 2 {
		t.Fatalf("expected 2 visible lines after collapseAll, got %d", collapsedLines)
	}

	viewCollapsed := tree.View()
	if !strings.Contains(viewCollapsed, "▸") {
		t.Fatalf("expected collapsed indicator in view, got:\n%s", viewCollapsed)
	}

	tree.ExpandAll()
	expandedLines := lineCount(tree.View())
	if expandedLines != 4 {
		t.Fatalf("expected 4 visible lines after expandAll, got %d", expandedLines)
	}
}

func TestTreeSelectedNodeAndCursorBounds(t *testing.T) {
	tree := NewTree(sampleTree())

	tree.Update(tea.KeyMsg{Type: tea.KeyDown})
	tree.Update(tea.KeyMsg{Type: tea.KeyDown})
	tree.Update(tea.KeyMsg{Type: tea.KeyDown})
	tree.Update(tea.KeyMsg{Type: tea.KeyDown})
	tree.Update(tea.KeyMsg{Type: tea.KeyDown}) // stay at end

	if got := tree.GetSelectedNode(); got == nil || got.Label != "root-2" {
		t.Fatalf("expected selection to clamp at root-2, got %#v", got)
	}

	tree.CollapseAll()
	if got := tree.GetSelectedNode(); got == nil || got.Label != "root-2" {
		t.Fatalf("expected selected node still valid after collapseAll, got %#v", got)
	}

	tree.SetRoots(nil)
	if got := tree.GetSelectedNode(); got != nil {
		t.Fatalf("expected nil selected node for empty tree, got %#v", got)
	}
}

func sampleTree() []*TreeNode {
	return []*TreeNode{
		{
			Label:    "root",
			Icon:     "📁",
			Expanded: true,
			Children: []*TreeNode{
				{Label: "child-a", Icon: "📄"},
				{Label: "child-b", Icon: "📄"},
			},
		},
		{
			Label: "root-2",
			Icon:  "📁",
		},
	}
}

func lineCount(view string) int {
	if strings.TrimSpace(view) == "" {
		return 0
	}
	return len(strings.Split(view, "\n"))
}
