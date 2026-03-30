package agent

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSubagentDiffListConstructor(t *testing.T) {
	items := []SubagentDiffItem{{FilePath: "a.go", LinesAdded: 1, LinesRemoved: 0}}
	list := NewSubagentDiffList(items, WithSubagentDiffListWidth(64))
	if list == nil {
		t.Fatal("NewSubagentDiffList returned nil")
	}

	if list.width != 64 {
		t.Fatalf("expected width=64, got %d", list.width)
	}
	if list.focus {
		t.Fatal("expected list to be unfocused by default")
	}
	if list.cursor != 0 {
		t.Fatalf("expected cursor=0, got %d", list.cursor)
	}
	if len(list.items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list.items))
	}

	items[0].FilePath = "mutated.go"
	if list.items[0].FilePath != "a.go" {
		t.Fatal("expected constructor to copy items slice")
	}
}

func TestSubagentDiffListDiffItemsAndSummary(t *testing.T) {
	list := NewSubagentDiffList([]SubagentDiffItem{
		{FilePath: "agent/a.go", Status: "M", LinesAdded: 12, LinesRemoved: 3},
		{FilePath: "display/b.go", Status: "A", LinesAdded: 5, LinesRemoved: 1},
	})

	view := stripANSI(list.View())
	if !strings.Contains(view, "agent/a.go") || !strings.Contains(view, "display/b.go") {
		t.Fatalf("expected file paths in view, got:\n%s", view)
	}
	if !strings.Contains(view, "[M]") || !strings.Contains(view, "[A]") {
		t.Fatalf("expected statuses in view, got:\n%s", view)
	}
	if !strings.Contains(view, "2 files changed, +17 -4") {
		t.Fatalf("expected summary totals in view, got:\n%s", view)
	}
}

func TestSubagentDiffListExpandCollapse(t *testing.T) {
	list := NewSubagentDiffList([]SubagentDiffItem{
		{FilePath: "x.go", Status: "M", Diff: "@@ -1 +1 @@\n-old\n+new"},
		{FilePath: "y.go", Status: "M", Diff: "@@ -2 +2 @@\n-a\n+b"},
	})

	if list.items[0].Expanded {
		t.Fatal("expected first item collapsed by default")
	}

	list.Update(SubagentDiffExpandMsg{Index: 0})
	if !list.items[0].Expanded {
		t.Fatal("expected first item to expand on SubagentDiffExpandMsg")
	}
	list.Update(SubagentDiffExpandMsg{Index: 0})
	if list.items[0].Expanded {
		t.Fatal("expected first item to collapse on second SubagentDiffExpandMsg")
	}

	list.Focus()
	list.Update(tea.KeyMsg{Type: tea.KeyDown})
	if list.cursor != 1 {
		t.Fatalf("expected cursor=1 after down key, got %d", list.cursor)
	}

	list.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !list.items[1].Expanded {
		t.Fatal("expected enter to toggle current item expansion")
	}

	list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if !list.items[0].Expanded || !list.items[1].Expanded {
		t.Fatal("expected 'a' to expand all items")
	}

	list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if list.items[0].Expanded || list.items[1].Expanded {
		t.Fatal("expected 'c' to collapse all items")
	}
}

func TestSubagentDiffListViewNonEmpty(t *testing.T) {
	list := NewSubagentDiffList([]SubagentDiffItem{{
		FilePath:     "calendar/month.go",
		Status:       "M",
		LinesAdded:   3,
		LinesRemoved: 1,
		Diff:         "@@ -1 +1 @@\n-old\n+new",
		Expanded:     true,
	}})

	view := stripANSI(list.View())
	if strings.TrimSpace(view) == "" {
		t.Fatal("expected non-empty view")
	}
	if !strings.Contains(view, "@@ -1 +1 @@") || !strings.Contains(view, "+new") {
		t.Fatalf("expected expanded diff lines in view, got:\n%s", view)
	}
}
