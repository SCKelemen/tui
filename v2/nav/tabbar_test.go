package nav

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTabBarCreation(t *testing.T) {
	tabBar := NewTabBar()
	if tabBar == nil {
		t.Fatal("NewTabBar returned nil")
	}

	if tabBar.TabCount() != 0 {
		t.Errorf("expected 0 tabs, got %d", tabBar.TabCount())
	}

	if tabBar.activeIdx != -1 {
		t.Errorf("expected activeIdx=-1, got %d", tabBar.activeIdx)
	}

	if tabBar.Focused() {
		t.Error("tab bar should not be focused initially")
	}
}

func TestTabBarWithTabs(t *testing.T) {
	tabBar := NewTabBar(
		WithTabs(
			Tab{ID: "files", Label: "Files"},
			Tab{ID: "search", Label: "Search"},
		),
	)

	if tabBar.TabCount() != 2 {
		t.Errorf("expected 2 tabs, got %d", tabBar.TabCount())
	}

	if tabBar.GetActiveID() != "files" {
		t.Errorf("expected active tab 'files', got %q", tabBar.GetActiveID())
	}
}

func TestTabBarAddRemoveTabs(t *testing.T) {
	tabBar := NewTabBar()

	tabBar.AddTab(Tab{ID: "files", Label: "Files"})
	tabBar.AddTab(Tab{ID: "search", Label: "Search"})
	tabBar.AddTab(Tab{ID: "terminal", Label: "Terminal"})

	if tabBar.TabCount() != 3 {
		t.Errorf("expected 3 tabs, got %d", tabBar.TabCount())
	}

	if tabBar.GetActiveID() != "files" {
		t.Errorf("expected first tab to be active, got %q", tabBar.GetActiveID())
	}

	tabBar.RemoveTab("search")
	if tabBar.TabCount() != 2 {
		t.Errorf("expected 2 tabs after remove, got %d", tabBar.TabCount())
	}

	tabBar.RemoveTab("files")
	if tabBar.GetActiveID() != "terminal" {
		t.Errorf("expected active tab to shift to terminal, got %q", tabBar.GetActiveID())
	}

	tabBar.RemoveTab("terminal")
	if tabBar.TabCount() != 0 {
		t.Errorf("expected 0 tabs after removing all, got %d", tabBar.TabCount())
	}

	if tabBar.GetActiveTab() != nil {
		t.Error("expected nil active tab when empty")
	}
}

func TestTabBarSetActiveAndGetters(t *testing.T) {
	tabBar := NewTabBar(
		WithTabs(
			Tab{ID: "files", Label: "Files"},
			Tab{ID: "search", Label: "Search"},
			Tab{ID: "terminal", Label: "Terminal"},
		),
	)

	tabBar.SetActive("terminal")
	if tabBar.GetActiveID() != "terminal" {
		t.Errorf("expected active id terminal, got %q", tabBar.GetActiveID())
	}

	active := tabBar.GetActiveTab()
	if active == nil {
		t.Fatal("expected active tab, got nil")
	}
	if active.Label != "Terminal" {
		t.Errorf("expected active label Terminal, got %q", active.Label)
	}

	// Unknown ID should leave active tab unchanged.
	tabBar.SetActive("missing")
	if tabBar.GetActiveID() != "terminal" {
		t.Errorf("expected active id to remain terminal, got %q", tabBar.GetActiveID())
	}
}

func TestTabBarSetBadge(t *testing.T) {
	tabBar := NewTabBar(
		WithTabs(
			Tab{ID: "terminal", Label: "Terminal"},
		),
	)

	tabBar.SetBadge("terminal", "3")
	view := tabBar.View()
	if !strings.Contains(view, "Terminal (3)") {
		t.Errorf("expected badge rendering in view, got %q", view)
	}
}

func TestTabBarFocusManagement(t *testing.T) {
	tabBar := NewTabBar()

	if tabBar.Focused() {
		t.Error("tab bar should not be focused initially")
	}

	tabBar.Focus()
	if !tabBar.Focused() {
		t.Error("tab bar should be focused after Focus")
	}

	tabBar.Blur()
	if tabBar.Focused() {
		t.Error("tab bar should not be focused after Blur")
	}
}

func TestTabBarKeyboardNavigation(t *testing.T) {
	tabBar := NewTabBar(
		WithTabs(
			Tab{ID: "files", Label: "Files"},
			Tab{ID: "search", Label: "Search"},
			Tab{ID: "terminal", Label: "Terminal"},
		),
	)
	tabBar.Focus()

	// tab -> next
	tabBar.Update(tea.KeyMsg{Type: tea.KeyTab})
	if tabBar.GetActiveID() != "search" {
		t.Errorf("expected search after tab, got %q", tabBar.GetActiveID())
	}

	// h -> previous
	tabBar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if tabBar.GetActiveID() != "files" {
		t.Errorf("expected files after h, got %q", tabBar.GetActiveID())
	}

	// l -> next
	tabBar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if tabBar.GetActiveID() != "search" {
		t.Errorf("expected search after l, got %q", tabBar.GetActiveID())
	}

	// direct jump
	tabBar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	if tabBar.GetActiveID() != "terminal" {
		t.Errorf("expected terminal after pressing 3, got %q", tabBar.GetActiveID())
	}

	// shift+tab wraps back
	tabBar.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if tabBar.GetActiveID() != "search" {
		t.Errorf("expected search after shift+tab, got %q", tabBar.GetActiveID())
	}
}

func TestTabBarOnSelectCallback(t *testing.T) {
	selected := ""
	tabBar := NewTabBar(
		WithTabs(
			Tab{ID: "files", Label: "Files"},
			Tab{ID: "search", Label: "Search"},
		),
		WithOnSelect(func(id string) tea.Cmd {
			selected = id
			return nil
		}),
	)
	tabBar.Focus()

	_, cmd := tabBar.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd != nil {
		cmd()
	}

	if selected != "search" {
		t.Errorf("expected onSelect search, got %q", selected)
	}
}

func TestTabBarCloseCallback(t *testing.T) {
	closed := ""
	tabBar := NewTabBar(
		WithTabs(
			Tab{ID: "files", Label: "Files", Closable: true},
			Tab{ID: "search", Label: "Search", Closable: false},
		),
		WithOnClose(func(id string) tea.Cmd {
			closed = id
			return nil
		}),
	)
	tabBar.Focus()

	_, cmd := tabBar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	if cmd != nil {
		cmd()
	}

	if closed != "files" {
		t.Errorf("expected closed id files, got %q", closed)
	}
	if tabBar.TabCount() != 1 {
		t.Errorf("expected 1 tab after close, got %d", tabBar.TabCount())
	}

	// Remaining tab is not closable.
	closed = ""
	_, cmd = tabBar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	if cmd != nil {
		cmd()
	}
	if closed != "" {
		t.Errorf("expected no close callback for non-closable tab, got %q", closed)
	}
}

func TestTabBarViewRendering(t *testing.T) {
	tabBar := NewTabBar(
		WithTabs(
			Tab{ID: "files", Label: "Files"},
			Tab{ID: "search", Label: "Search"},
			Tab{ID: "terminal", Label: "Terminal", Badge: "3"},
		),
	)
	tabBar.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	tabBar.SetActive("files")

	view := tabBar.View()
	if view == "" {
		t.Fatal("view should not be empty")
	}

	if !strings.Contains(view, "┌─ Files ─┐") {
		t.Errorf("view should render active tab frame, got %q", view)
	}
	if !strings.Contains(view, "Search") {
		t.Errorf("view should render inactive tab label, got %q", view)
	}
	if !strings.Contains(view, "Terminal (3)") {
		t.Errorf("view should render badge text, got %q", view)
	}
	if !strings.Contains(view, "┘") || !strings.Contains(view, "└") {
		t.Errorf("view should render connected bottom border corners, got %q", view)
	}
	if !strings.Contains(view, "─") {
		t.Errorf("view should render bottom border line, got %q", view)
	}
}

func TestTabBarIgnoresKeysWhenBlurred(t *testing.T) {
	tabBar := NewTabBar(
		WithTabs(
			Tab{ID: "files", Label: "Files"},
			Tab{ID: "search", Label: "Search"},
		),
	)
	tabBar.Blur()

	tabBar.Update(tea.KeyMsg{Type: tea.KeyTab})
	if tabBar.GetActiveID() != "files" {
		t.Errorf("expected active tab unchanged while blurred, got %q", tabBar.GetActiveID())
	}
}
