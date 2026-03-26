package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestChecklistCreation(t *testing.T) {
	checklist := NewChecklist()
	if checklist == nil {
		t.Fatal("NewChecklist returned nil")
	}

	if checklist.title != "Checklist" {
		t.Errorf("expected default title Checklist, got %q", checklist.title)
	}

	if checklist.showProgress != true {
		t.Error("showProgress should default to true")
	}

	if checklist.ItemCount() != 0 {
		t.Errorf("expected 0 items initially, got %d", checklist.ItemCount())
	}
}

func TestChecklistOptions(t *testing.T) {
	items := []ChecklistItem{
		{ID: "1", Label: "one", Status: ItemPending},
		{ID: "2", Label: "two", Status: ItemDone},
	}

	checklist := NewChecklist(
		WithChecklistTitle("Build Pipeline"),
		WithChecklistItems(items...),
		WithChecklistIconSet(IconSetSymbols),
		WithShowProgress(false),
	)

	if checklist.title != "Build Pipeline" {
		t.Errorf("expected title Build Pipeline, got %q", checklist.title)
	}

	if checklist.ItemCount() != 2 {
		t.Errorf("expected 2 items, got %d", checklist.ItemCount())
	}

	if checklist.iconSet.Success != IconSetSymbols.Success {
		t.Error("WithChecklistIconSet was not applied")
	}

	if checklist.showProgress {
		t.Error("WithShowProgress(false) was not applied")
	}
}

func TestChecklistAddAndRemoveItem(t *testing.T) {
	checklist := NewChecklist()
	checklist.AddItem(ChecklistItem{ID: "compile", Label: "Compile", Status: ItemPending})
	checklist.AddItem(ChecklistItem{ID: "test", Label: "Test", Status: ItemPending})

	if checklist.ItemCount() != 2 {
		t.Errorf("expected 2 items after add, got %d", checklist.ItemCount())
	}

	checklist.RemoveItem("compile")
	if checklist.ItemCount() != 1 {
		t.Errorf("expected 1 item after remove, got %d", checklist.ItemCount())
	}

	if checklist.GetItem("compile") != nil {
		t.Error("removed item should no longer exist")
	}
}

func TestChecklistSetItemStatus(t *testing.T) {
	checklist := NewChecklist(
		WithChecklistItems(
			ChecklistItem{ID: "compile", Label: "Compile", Status: ItemPending},
		),
	)

	checklist.SetItemStatus("compile", ItemInProgress)
	item := checklist.GetItem("compile")
	if item == nil {
		t.Fatal("item should exist")
	}
	if item.Status != ItemInProgress {
		t.Errorf("expected status ItemInProgress, got %v", item.Status)
	}

	checklist.SetItemStatus("compile", ItemDone)
	if item.Status != ItemDone {
		t.Errorf("expected status ItemDone, got %v", item.Status)
	}
}

func TestChecklistToggleItem(t *testing.T) {
	checklist := NewChecklist(
		WithChecklistItems(
			ChecklistItem{ID: "compile", Label: "Compile", Status: ItemPending},
		),
	)

	checklist.ToggleItem("compile")
	if checklist.GetItem("compile").Status != ItemDone {
		t.Error("toggle should set pending item to done")
	}

	checklist.ToggleItem("compile")
	if checklist.GetItem("compile").Status != ItemPending {
		t.Error("toggle should set done item back to pending")
	}
}

func TestChecklistNestedItems(t *testing.T) {
	checklist := NewChecklist(
		WithChecklistItems(
			ChecklistItem{
				ID:       "parent",
				Label:    "Parent",
				Status:   ItemPending,
				Expanded: true,
				Children: []ChecklistItem{
					{ID: "child-1", Label: "Child 1", Status: ItemDone},
					{ID: "child-2", Label: "Child 2", Status: ItemPending},
				},
			},
		),
	)

	if checklist.ItemCount() != 3 {
		t.Errorf("expected nested item count 3, got %d", checklist.ItemCount())
	}

	child := checklist.GetItem("child-2")
	if child == nil {
		t.Fatal("expected nested child item to be found")
	}
	if child.Label != "Child 2" {
		t.Errorf("expected child label Child 2, got %q", child.Label)
	}

	checklist.RemoveItem("child-1")
	if checklist.ItemCount() != 2 {
		t.Errorf("expected 2 items after nested remove, got %d", checklist.ItemCount())
	}
}

func TestChecklistProgressCalculation(t *testing.T) {
	checklist := NewChecklist(
		WithChecklistItems(
			ChecklistItem{ID: "a", Label: "A", Status: ItemDone},
			ChecklistItem{ID: "b", Label: "B", Status: ItemPending},
			ChecklistItem{ID: "c", Label: "C", Status: ItemInProgress},
			ChecklistItem{ID: "d", Label: "D", Status: ItemSkipped},
			ChecklistItem{ID: "e", Label: "E", Status: ItemFailed},
		),
	)

	done, total := checklist.CompletionCount()
	if done != 4 {
		t.Errorf("expected done=4, got %d", done)
	}
	if total != 5 {
		t.Errorf("expected total=5, got %d", total)
	}

	checklist.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := checklist.View()
	if !strings.Contains(view, "[4/5") {
		t.Error("view should include progress count")
	}
	if !strings.Contains(view, "80%") {
		t.Error("view should include percentage")
	}
}

func TestChecklistKeyboardNavigation(t *testing.T) {
	checklist := NewChecklist(
		WithChecklistItems(
			ChecklistItem{ID: "a", Label: "A", Status: ItemPending},
			ChecklistItem{ID: "b", Label: "B", Status: ItemPending},
			ChecklistItem{ID: "c", Label: "C", Status: ItemPending},
		),
	)
	checklist.Focus()

	if checklist.selected != 0 {
		t.Errorf("expected selected=0, got %d", checklist.selected)
	}

	checklist.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if checklist.selected != 1 {
		t.Errorf("expected selected=1 after j, got %d", checklist.selected)
	}

	checklist.Update(tea.KeyMsg{Type: tea.KeyDown})
	if checklist.selected != 2 {
		t.Errorf("expected selected=2 after down, got %d", checklist.selected)
	}

	checklist.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if checklist.selected != 1 {
		t.Errorf("expected selected=1 after k, got %d", checklist.selected)
	}
}

func TestChecklistKeyboardToggleAndBulk(t *testing.T) {
	checklist := NewChecklist(
		WithChecklistItems(
			ChecklistItem{ID: "a", Label: "A", Status: ItemPending},
			ChecklistItem{ID: "b", Label: "B", Status: ItemPending},
		),
	)
	checklist.Focus()

	checklist.Update(tea.KeyMsg{Type: tea.KeySpace})
	if checklist.GetItem("a").Status != ItemDone {
		t.Error("space should toggle selected item to done")
	}

	checklist.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if checklist.GetItem("a").Status != ItemPending {
		t.Error("x should toggle selected item back to pending")
	}

	checklist.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if checklist.GetItem("a").Status != ItemDone || checklist.GetItem("b").Status != ItemDone {
		t.Error("a should mark all items done")
	}

	checklist.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	if checklist.GetItem("a").Status != ItemPending || checklist.GetItem("b").Status != ItemPending {
		t.Error("u should mark all items pending")
	}
}

func TestChecklistExpandCollapse(t *testing.T) {
	checklist := NewChecklist(
		WithChecklistItems(
			ChecklistItem{
				ID:       "parent",
				Label:    "Parent",
				Status:   ItemPending,
				Expanded: false,
				Children: []ChecklistItem{
					{ID: "child", Label: "Child", Status: ItemPending},
				},
			},
		),
	)
	checklist.Focus()
	checklist.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := checklist.View()
	if strings.Contains(view, "Child") {
		t.Error("child should not be visible while collapsed")
	}

	checklist.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !checklist.GetItem("parent").Expanded {
		t.Error("enter should expand selected parent item")
	}

	view = checklist.View()
	if !strings.Contains(view, "Child") {
		t.Error("child should be visible after expand")
	}

	checklist.SetExpanded("parent", false)
	if checklist.GetItem("parent").Expanded {
		t.Error("SetExpanded should collapse parent")
	}
}

func TestChecklistViewAndConnectors(t *testing.T) {
	checklist := NewChecklist(
		WithChecklistTitle("Build Pipeline"),
		WithChecklistItems(
			ChecklistItem{ID: "compile", Label: "Compile", Status: ItemDone},
			ChecklistItem{
				ID:       "tests",
				Label:    "Tests",
				Status:   ItemInProgress,
				Expanded: true,
				Children: []ChecklistItem{
					{ID: "unit", Label: "Unit", Status: ItemDone},
					{ID: "integration", Label: "Integration", Status: ItemPending},
				},
			},
		),
	)
	checklist.Update(tea.WindowSizeMsg{Width: 90, Height: 24})

	view := checklist.View()
	if view == "" {
		t.Fatal("view should not be empty when width is set")
	}

	if !strings.Contains(view, "Build Pipeline") {
		t.Error("view should include title")
	}

	if !strings.Contains(view, "├") || !strings.Contains(view, "└") {
		t.Error("view should include tree connectors")
	}

	if !strings.Contains(view, IconSetSymbols.Success) {
		t.Error("view should include done icon")
	}
}

func TestChecklistViewWithoutWidth(t *testing.T) {
	checklist := NewChecklist(
		WithChecklistItems(ChecklistItem{ID: "a", Label: "A", Status: ItemPending}),
	)

	if view := checklist.View(); view != "" {
		t.Error("view should be empty when width is not set")
	}
}

func TestChecklistFocusManagement(t *testing.T) {
	checklist := NewChecklist()
	if checklist.Focused() {
		t.Error("should not be focused initially")
	}

	checklist.Focus()
	if !checklist.Focused() {
		t.Error("should be focused after Focus")
	}

	checklist.Blur()
	if checklist.Focused() {
		t.Error("should not be focused after Blur")
	}
}

func TestChecklistDesignTokens(t *testing.T) {
	checklist := NewChecklist(
		WithChecklistTheme("nord"),
	)

	if checklist.designTokens == nil {
		t.Error("design tokens should be set")
	}

	if checklist.accentColor == "" {
		t.Error("accent color should be set from design tokens")
	}
}
