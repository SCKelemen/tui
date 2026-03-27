package nav

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBreadcrumbCreation(t *testing.T) {
	breadcrumb := NewBreadcrumb(
		WithBreadcrumbItems(
			BreadcrumbItem{ID: "home", Label: "Home"},
			BreadcrumbItem{ID: "projects", Label: "Projects"},
		),
	)

	if breadcrumb == nil {
		t.Fatal("NewBreadcrumb returned nil")
	}

	if breadcrumb.separator != " › " {
		t.Errorf("expected default separator ' › ', got %q", breadcrumb.separator)
	}

	if !breadcrumb.truncateFromLeft {
		t.Error("expected truncateFromLeft to default to true")
	}

	if breadcrumb.ItemCount() != 2 {
		t.Errorf("expected 2 items, got %d", breadcrumb.ItemCount())
	}

	selected := breadcrumb.GetSelected()
	if selected == nil || selected.ID != "projects" {
		t.Fatalf("expected selected item to be 'projects', got %+v", selected)
	}
}

func TestBreadcrumbPushPop(t *testing.T) {
	breadcrumb := NewBreadcrumb(
		WithBreadcrumbItems(
			BreadcrumbItem{ID: "home", Label: "Home"},
		),
	)

	breadcrumb.Push(BreadcrumbItem{ID: "stack", Label: "stack"})
	if breadcrumb.ItemCount() != 2 {
		t.Errorf("expected 2 items after push, got %d", breadcrumb.ItemCount())
	}

	selected := breadcrumb.GetSelected()
	if selected == nil || selected.ID != "stack" {
		t.Fatalf("expected selected item to be 'stack' after push, got %+v", selected)
	}

	popped := breadcrumb.Pop()
	if popped == nil || popped.ID != "stack" {
		t.Fatalf("expected popped item 'stack', got %+v", popped)
	}

	if breadcrumb.ItemCount() != 1 {
		t.Errorf("expected 1 item after pop, got %d", breadcrumb.ItemCount())
	}

	popped = breadcrumb.Pop()
	if popped == nil || popped.ID != "home" {
		t.Fatalf("expected popped item 'home', got %+v", popped)
	}

	if breadcrumb.Pop() != nil {
		t.Error("expected pop on empty breadcrumb to return nil")
	}
}

func TestBreadcrumbOverflowTruncation(t *testing.T) {
	breadcrumb := NewBreadcrumb(
		WithBreadcrumbItems(
			BreadcrumbItem{ID: "home", Label: "Home"},
			BreadcrumbItem{ID: "projects", Label: "Projects"},
			BreadcrumbItem{ID: "stack", Label: "stack"},
			BreadcrumbItem{ID: "tui", Label: "tui"},
		),
	)

	breadcrumb.Update(tea.WindowSizeMsg{Width: 20, Height: 1})
	view := breadcrumb.View()

	if view == "" {
		t.Fatal("view should not be empty")
	}

	if !strings.Contains(view, "...") {
		t.Errorf("expected overflow to include ellipsis, got %q", view)
	}

	if !strings.Contains(view, "stack") || !strings.Contains(view, "tui") {
		t.Errorf("expected truncated view to keep right-most items, got %q", view)
	}
}

func TestBreadcrumbKeyboardNavigation(t *testing.T) {
	breadcrumb := NewBreadcrumb(
		WithBreadcrumbItems(
			BreadcrumbItem{ID: "home", Label: "Home"},
			BreadcrumbItem{ID: "projects", Label: "Projects"},
			BreadcrumbItem{ID: "tui", Label: "TUI"},
		),
	)
	breadcrumb.Focus()

	// Start on last item by default.
	if got := breadcrumb.GetSelected(); got == nil || got.ID != "tui" {
		t.Fatalf("expected default selected item 'tui', got %+v", got)
	}

	// h moves left.
	breadcrumb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if got := breadcrumb.GetSelected(); got == nil || got.ID != "projects" {
		t.Fatalf("expected selected item 'projects' after h, got %+v", got)
	}

	// right arrow moves right.
	breadcrumb.Update(tea.KeyMsg{Type: tea.KeyRight})
	if got := breadcrumb.GetSelected(); got == nil || got.ID != "tui" {
		t.Fatalf("expected selected item 'tui' after right, got %+v", got)
	}

	// left arrow moves left.
	breadcrumb.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if got := breadcrumb.GetSelected(); got == nil || got.ID != "projects" {
		t.Fatalf("expected selected item 'projects' after left, got %+v", got)
	}
}

func TestBreadcrumbSelectionCallback(t *testing.T) {
	selectedID := ""
	breadcrumb := NewBreadcrumb(
		WithBreadcrumbItems(
			BreadcrumbItem{ID: "home", Label: "Home"},
			BreadcrumbItem{ID: "projects", Label: "Projects"},
		),
		WithOnBreadcrumbSelect(func(id string) tea.Cmd {
			selectedID = id
			return nil
		}),
	)
	breadcrumb.Focus()
	breadcrumb.SetSelected("home")

	_, cmd := breadcrumb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		cmd()
	}

	if selectedID != "home" {
		t.Errorf("expected callback with id 'home', got %q", selectedID)
	}
}

func TestBreadcrumbSeparatorCustomization(t *testing.T) {
	breadcrumb := NewBreadcrumb(
		WithSeparator(" / "),
		WithBreadcrumbItems(
			BreadcrumbItem{ID: "home", Label: "Home"},
			BreadcrumbItem{ID: "stack", Label: "stack"},
		),
	)

	view := breadcrumb.View()
	if !strings.Contains(view, " / ") {
		t.Errorf("expected custom separator in view, got %q", view)
	}
	if strings.Contains(view, " › ") {
		t.Errorf("did not expect default separator in view, got %q", view)
	}
}
