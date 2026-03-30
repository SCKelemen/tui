package input

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func testCodeActions() []CodeAction {
	return []CodeAction{
		{Title: "Fix import", Kind: KindQuickFix, Description: "Add missing fmt import"},
		{Title: "Extract variable", Kind: KindRefactorExtract, Description: "Extract expression"},
	}
}

func TestCodeActionMenuConstructor(t *testing.T) {
	menu := NewCodeActionMenu(testCodeActions(), WithCodeActionWidth(64), WithCodeActionAnchor(1))
	if menu == nil {
		t.Fatal("NewCodeActionMenu returned nil")
	}

	if menu.width != 64 {
		t.Fatalf("expected width 64, got %d", menu.width)
	}

	if menu.anchorLine != 1 {
		t.Fatalf("expected anchor line 1, got %d", menu.anchorLine)
	}

	if menu.Visible() {
		t.Fatal("menu should be hidden by default")
	}
}

func TestCodeActionMenuItemsRender(t *testing.T) {
	menu := NewCodeActionMenu(testCodeActions())
	menu.Show()
	menu.Focus()

	view := menu.View()
	if view == "" {
		t.Fatal("view should not be empty when visible")
	}

	if !strings.Contains(view, "Fix import") || !strings.Contains(view, "Extract variable") {
		t.Fatal("view should include action titles")
	}
}

func TestCodeActionMenuSelection(t *testing.T) {
	menu := NewCodeActionMenu(testCodeActions())
	menu.Show()
	menu.Focus()

	menu.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, cmd := menu.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected enter to return command")
	}

	msg := cmd()
	selectedMsg, ok := msg.(CodeActionSelectedMsg)
	if !ok {
		t.Fatalf("expected CodeActionSelectedMsg, got %T", msg)
	}

	if selectedMsg.Action.Title != "Extract variable" {
		t.Fatalf("expected selected action 'Extract variable', got %q", selectedMsg.Action.Title)
	}

	if menu.Visible() {
		t.Fatal("menu should hide after selection")
	}
}
