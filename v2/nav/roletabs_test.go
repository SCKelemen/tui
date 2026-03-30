package nav

import (
	"strings"
	"testing"

	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

func TestRoleTabsConstructorAndInitialSelection(t *testing.T) {
	tabs := []RoleTab{
		{Label: "Planner", Role: "planner"},
		{Label: "Coder", Role: "coder", Active: true},
	}

	r := NewRoleTabs(tabs)
	if r == nil {
		t.Fatal("NewRoleTabs returned nil")
	}

	if r.selected != 1 {
		t.Fatalf("expected selected index 1 from active tab, got %d", r.selected)
	}

	if !r.tabs[1].Active {
		t.Fatalf("expected index 1 to remain active")
	}
}

func TestRoleTabsSelectionAndEnterEmitsMessage(t *testing.T) {
	r := NewRoleTabs([]RoleTab{
		{Label: "Planner", Role: "planner"},
		{Label: "Coder", Role: "coder"},
	})
	r.Focus()

	_, _ = r.Update(tea.KeyMsg{Type: tea.KeyRight})
	if r.selected != 1 {
		t.Fatalf("expected selected index 1 after right, got %d", r.selected)
	}
	if !r.tabs[1].Active {
		t.Fatalf("expected tab 1 active after selection")
	}

	_, cmd := r.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected enter to return a command")
	}

	msg := cmd()
	selectedMsg, ok := msg.(RoleTabSelectedMsg)
	if !ok {
		t.Fatalf("expected RoleTabSelectedMsg, got %T", msg)
	}

	if selectedMsg.Index != 1 || selectedMsg.Role != "coder" {
		t.Fatalf("unexpected selected msg: %+v", selectedMsg)
	}
}

func TestRoleTabsViewUsesRoleColors(t *testing.T) {
	r := NewRoleTabs(
		[]RoleTab{
			{Label: "Planner", Role: "planner", Active: true},
			{Label: "Coder", Role: "coder"},
		},
		WithRoleTabsPalettes(map[string]RolePalette{
			"planner": {"#101010", "#202020", "#303030", "#404040", "#505050"},
			"coder":   {"#111111", "#222222", "#ABCDEF", "#444444", "#555555"},
		}),
	)

	got := r.View()
	if !strings.Contains(got, "[ Planner ]") || !strings.Contains(got, "[ Coder ]") {
		t.Fatalf("expected both tab labels in view, got %q", got)
	}

	if !strings.Contains(got, style.ANSIBold) || !strings.Contains(got, style.ANSIUnderline) {
		t.Fatalf("expected selected tab styling (bold+underline), got %q", got)
	}

	inactiveColor := style.Fg("#ABCDEF")
	if inactiveColor != "" && !strings.Contains(got, inactiveColor) {
		t.Fatalf("expected inactive tab color %q in view, got %q", inactiveColor, got)
	}
}
