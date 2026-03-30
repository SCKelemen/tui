package input

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func testCompletionItems() []CompletionItem {
	return []CompletionItem{
		{Label: "fmt.Println", InsertText: "fmt.Println", Kind: KindFunction, Detail: "func(a ...any)"},
		{Label: "fmt.Printf", InsertText: "fmt.Printf", Kind: KindFunction, Detail: "func(format string, a ...any)"},
		{Label: "myVar", InsertText: "myVar", Kind: KindVariable},
	}
}

func TestIntelliSenseConstructor(t *testing.T) {
	is := NewIntelliSense(testCompletionItems(), WithIntelliSenseWidth(50), WithIntelliSenseMaxVisible(5))
	if is == nil {
		t.Fatal("NewIntelliSense returned nil")
	}

	if is.Visible() {
		t.Fatal("intellisense should be hidden by default")
	}

	if len(is.filtered) != len(testCompletionItems()) {
		t.Fatalf("expected %d filtered items, got %d", len(testCompletionItems()), len(is.filtered))
	}
}

func TestIntelliSenseCompletionItemsAndView(t *testing.T) {
	is := NewIntelliSense(testCompletionItems(), WithIntelliSenseWidth(60))
	is.Show()

	view := is.View()
	if view == "" {
		t.Fatal("view should not be empty when visible")
	}

	if !strings.Contains(view, "fmt.Println") {
		t.Fatal("view should contain completion item labels")
	}
}

func TestIntelliSenseSelection(t *testing.T) {
	is := NewIntelliSense(testCompletionItems())
	is.Show()

	is.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, cmd := is.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command on enter")
	}

	msg := cmd()
	selectedMsg, ok := msg.(CompletionSelectedMsg)
	if !ok {
		t.Fatalf("expected CompletionSelectedMsg, got %T", msg)
	}

	if selectedMsg.Item.Label != "fmt.Printf" {
		t.Fatalf("expected fmt.Printf selected, got %q", selectedMsg.Item.Label)
	}

	if is.Visible() {
		t.Fatal("intellisense should hide after selecting an item")
	}
}
