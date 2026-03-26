package tui

import (
	"strings"
	"testing"
	"time"

	design "github.com/SCKelemen/design-system"
	tea "github.com/charmbracelet/bubbletea"
)

func TestToastCreation(t *testing.T) {
	toast := NewToast()
	if toast == nil {
		t.Fatal("NewToast returned nil")
	}

	if toast.Count() != 0 {
		t.Errorf("Expected empty toast list, got %d", toast.Count())
	}

	if toast.maxVisible != 5 {
		t.Errorf("Expected default maxVisible=5, got %d", toast.maxVisible)
	}

	if toast.defaultDuration != 4*time.Second {
		t.Errorf("Expected default duration=4s, got %v", toast.defaultDuration)
	}

	if toast.Focused() {
		t.Error("Toast should not be focused initially")
	}
}

func TestToastCreationWithOptions(t *testing.T) {
	toast := NewToast(
		WithMaxVisible(2),
		WithDefaultDuration(2*time.Second),
		WithToastIconSet(IconSetMinimal),
		WithToastTheme("midnight"),
	)

	if toast.maxVisible != 2 {
		t.Errorf("Expected maxVisible=2, got %d", toast.maxVisible)
	}

	if toast.defaultDuration != 2*time.Second {
		t.Errorf("Expected duration=2s, got %v", toast.defaultDuration)
	}

	if toast.iconSet.Success != IconSetMinimal.Success {
		t.Error("Expected minimal icon set to be applied")
	}
}

func TestToastInitAndWindowSizeUpdate(t *testing.T) {
	toast := NewToast()
	cmd := toast.Init()
	if cmd == nil {
		t.Error("Init should return tick command")
	}

	updated, _ := toast.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_ = updated
	if toast.width != 80 {
		t.Errorf("Expected width=80, got %d", toast.width)
	}
}

func TestToastPushAndDismiss(t *testing.T) {
	toast := NewToast()

	id1 := toast.Push("first", ToastInfo)
	if id1 == "" {
		t.Error("Push should return non-empty ID")
	}

	id2 := toast.PushWithDuration("second", ToastSuccess, 10*time.Second)
	if id2 == "" {
		t.Error("PushWithDuration should return non-empty ID")
	}

	if id1 == id2 {
		t.Error("Toast IDs must be unique")
	}

	if toast.Count() != 2 {
		t.Errorf("Expected 2 toasts, got %d", toast.Count())
	}

	// Most recent first
	if toast.items[0].ID != id2 {
		t.Error("Most recent toast should be first")
	}

	toast.Dismiss(id2)
	if toast.Count() != 1 {
		t.Errorf("Expected 1 toast after dismiss, got %d", toast.Count())
	}

	toast.Dismiss("does-not-exist")
	if toast.Count() != 1 {
		t.Errorf("Count should remain unchanged for unknown ID, got %d", toast.Count())
	}

	toast.DismissAll()
	if toast.Count() != 0 {
		t.Errorf("Expected 0 toasts after DismissAll, got %d", toast.Count())
	}
}

func TestToastAutoDismissTiming(t *testing.T) {
	toast := NewToast(WithDefaultDuration(5 * time.Second))
	id := toast.PushWithDuration("short", ToastWarning, 100*time.Millisecond)
	if id == "" {
		t.Fatal("Expected toast ID")
	}

	if toast.Count() != 1 {
		t.Fatalf("Expected 1 toast before tick, got %d", toast.Count())
	}

	// Force it to be old enough to expire.
	toast.items[0].CreatedAt = time.Now().Add(-200 * time.Millisecond)

	_, cmd := toast.Update(toastTickMsg(time.Now()))
	if cmd == nil {
		t.Error("Tick update should schedule next tick")
	}

	if toast.Count() != 0 {
		t.Errorf("Expected toast to auto-dismiss after tick, got %d", toast.Count())
	}
}

func TestToastMaxVisibleRendering(t *testing.T) {
	toast := NewToast(WithMaxVisible(2), WithToastIconSet(IconSetMinimal))
	toast.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	toast.Push("one", ToastInfo)
	toast.Push("two", ToastSuccess)
	toast.Push("three", ToastError)

	view := toast.View()
	if view == "" {
		t.Fatal("View should not be empty")
	}

	lines := strings.Split(strings.TrimSpace(view), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 visible lines, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "three") {
		t.Error("Most recent toast should be rendered first")
	}

	if strings.Contains(view, "one") {
		t.Error("Oldest toast should not be visible when maxVisible=2")
	}
}

func TestToastLevelRendering(t *testing.T) {
	icons := IconSet{
		Info:    "I",
		Success: "S",
		Warning: "W",
		Error:   "E",
	}

	toast := NewToast(WithToastIconSet(icons))
	toast.Update(tea.WindowSizeMsg{Width: 120, Height: 24})

	toast.Push("info", ToastInfo)
	toast.Push("success", ToastSuccess)
	toast.Push("warning", ToastWarning)
	toast.Push("error", ToastError)

	view := toast.View()
	if view == "" {
		t.Fatal("View should not be empty")
	}

	// Most recent first: error, warning, success, info
	if !strings.Contains(view, "\033[31mE\033[0m") {
		t.Error("Expected error icon with red color")
	}
	if !strings.Contains(view, "\033[33mW\033[0m") {
		t.Error("Expected warning icon with yellow color")
	}
	if !strings.Contains(view, "\033[32mS\033[0m") {
		t.Error("Expected success icon with green color")
	}
	if !strings.Contains(view, "\033[36mI\033[0m") {
		t.Error("Expected info icon with accent color")
	}
}

func TestToastViewAndFocus(t *testing.T) {
	toast := NewToast()

	if got := toast.View(); got != "" {
		t.Errorf("Expected empty view for empty toast list, got %q", got)
	}

	toast.Push("hello", ToastInfo)

	toast.Focus()
	if !toast.Focused() {
		t.Error("Toast should be focused after Focus()")
	}

	focusedView := toast.View()
	if !strings.Contains(focusedView, "\033[7m") {
		t.Error("Focused view should contain inverse styling")
	}

	toast.Blur()
	if toast.Focused() {
		t.Error("Toast should not be focused after Blur()")
	}
}

func TestToastDesignTokensOption(t *testing.T) {
	tokens := design.DefaultTheme()
	toast := NewToast(WithToastDesignTokens(tokens))

	if toast.textColor == "" {
		t.Error("Expected textColor to be set by design tokens")
	}
	if toast.infoColor == "" {
		t.Error("Expected infoColor to be set by design tokens")
	}
}
