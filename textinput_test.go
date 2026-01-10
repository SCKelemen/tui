package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTextInputCreation(t *testing.T) {
	ti := NewTextInput()
	if ti == nil {
		t.Fatal("NewTextInput returned nil")
	}

	if ti.focused {
		t.Error("TextInput should not be focused initially")
	}

	if ti.height != 5 {
		t.Errorf("Expected default height 5, got %d", ti.height)
	}

	if ti.Value() != "" {
		t.Error("TextInput should be empty initially")
	}
}

func TestTextInputSetValue(t *testing.T) {
	ti := NewTextInput()

	testValue := "Hello, world!"
	ti.SetValue(testValue)

	if ti.Value() != testValue {
		t.Errorf("Expected value %q, got %q", testValue, ti.Value())
	}
}

func TestTextInputReset(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("Some text")

	if ti.Value() == "" {
		t.Error("Value should not be empty before reset")
	}

	ti.Reset()

	if ti.Value() != "" {
		t.Errorf("Value should be empty after reset, got %q", ti.Value())
	}
}

func TestTextInputFocusManagement(t *testing.T) {
	ti := NewTextInput()

	if ti.Focused() {
		t.Error("TextInput should not be focused initially")
	}

	ti.Focus()
	if !ti.Focused() {
		t.Error("TextInput should be focused after Focus()")
	}

	if !ti.textarea.Focused() {
		t.Error("Internal textarea should be focused when TextInput is focused")
	}

	ti.Blur()
	if ti.Focused() {
		t.Error("TextInput should not be focused after Blur()")
	}
}

func TestTextInputView(t *testing.T) {
	ti := NewTextInput()
	ti.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := ti.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// Check for top border
	if !strings.Contains(view, "┌") {
		t.Error("View should contain top-left corner (┌)")
	}

	// Check for bottom border
	if !strings.Contains(view, "└") {
		t.Error("View should contain bottom-left corner (└)")
	}

	// Check for vertical borders
	if !strings.Contains(view, "│") {
		t.Error("View should contain vertical borders (│)")
	}
}

func TestTextInputViewWithoutWidth(t *testing.T) {
	ti := NewTextInput()

	view := ti.View()
	if view != "" {
		t.Error("View should be empty when width is not set")
	}
}

func TestTextInputViewFocused(t *testing.T) {
	ti := NewTextInput()
	ti.Focus()
	ti.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := ti.View()

	// Should show keyboard hints when focused
	if !strings.Contains(view, "Ctrl+J: send") {
		t.Error("Focused view should show 'Ctrl+J: send' hint")
	}

	if !strings.Contains(view, "Ctrl+D: clear") {
		t.Error("Focused view should show 'Ctrl+D: clear' hint")
	}
}

func TestTextInputViewBlurred(t *testing.T) {
	ti := NewTextInput()
	ti.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := ti.View()

	// Should NOT show keyboard hints when blurred
	if strings.Contains(view, "Ctrl+J: send") {
		t.Error("Blurred view should not show keyboard hints")
	}
}

func TestTextInputWindowSizeUpdate(t *testing.T) {
	ti := NewTextInput()

	if ti.width != 0 {
		t.Error("Initial width should be 0")
	}

	ti.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if ti.width != 100 {
		t.Errorf("Expected width 100, got %d", ti.width)
	}
}

func TestTextInputCtrlJSubmit(t *testing.T) {
	ti := NewTextInput()
	ti.Focus()
	ti.SetValue("Test message")

	submitted := false
	var submittedText string

	ti.OnSubmit(func(text string) tea.Cmd {
		submitted = true
		submittedText = text
		return nil
	})

	// Send Ctrl+J
	ti.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})

	if !submitted {
		t.Error("Ctrl+J should trigger submit")
	}

	if submittedText != "Test message" {
		t.Errorf("Expected submitted text 'Test message', got %q", submittedText)
	}

	// Value should be cleared after submit
	if ti.Value() != "" {
		t.Error("Value should be cleared after submit")
	}
}

func TestTextInputCtrlJEmptyNoSubmit(t *testing.T) {
	ti := NewTextInput()
	ti.Focus()
	ti.SetValue("")

	submitted := false
	ti.OnSubmit(func(text string) tea.Cmd {
		submitted = true
		return nil
	})

	// Send Ctrl+J with empty value
	ti.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})

	if submitted {
		t.Error("Ctrl+J should not trigger submit when value is empty")
	}
}

func TestTextInputCtrlJWhitespaceNoSubmit(t *testing.T) {
	ti := NewTextInput()
	ti.Focus()
	ti.SetValue("   \n  \t  ")

	submitted := false
	ti.OnSubmit(func(text string) tea.Cmd {
		submitted = true
		return nil
	})

	// Send Ctrl+J with whitespace-only value
	ti.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})

	if submitted {
		t.Error("Ctrl+J should not trigger submit when value is only whitespace")
	}
}

func TestTextInputCtrlDClear(t *testing.T) {
	ti := NewTextInput()
	ti.Focus()
	ti.SetValue("Some text to clear")

	if ti.Value() == "" {
		t.Error("Value should not be empty before Ctrl+D")
	}

	// Send Ctrl+D
	ti.Update(tea.KeyMsg{Type: tea.KeyCtrlD})

	if ti.Value() != "" {
		t.Errorf("Value should be empty after Ctrl+D, got %q", ti.Value())
	}
}

func TestTextInputNoSubmitWhenBlurred(t *testing.T) {
	ti := NewTextInput()
	// Don't focus
	ti.SetValue("Test message")

	submitted := false
	ti.OnSubmit(func(text string) tea.Cmd {
		submitted = true
		return nil
	})

	// Send Ctrl+J (should have no effect when not focused)
	ti.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})

	if submitted {
		t.Error("Ctrl+J should not trigger submit when not focused")
	}
}

func TestTextInputNoClearWhenBlurred(t *testing.T) {
	ti := NewTextInput()
	// Don't focus
	ti.SetValue("Test message")

	// Send Ctrl+D (should have no effect when not focused)
	ti.Update(tea.KeyMsg{Type: tea.KeyCtrlD})

	if ti.Value() == "" {
		t.Error("Ctrl+D should not clear when not focused")
	}
}

func TestTextInputMultilineSupport(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("Line 1\nLine 2\nLine 3")

	value := ti.Value()
	if !strings.Contains(value, "\n") {
		t.Error("TextInput should support multi-line text")
	}

	lines := strings.Split(value, "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}
}

func TestTextInputOnSubmitCallback(t *testing.T) {
	ti := NewTextInput()
	ti.Focus()
	ti.SetValue("Test")

	callbackCalled := false
	ti.OnSubmit(func(text string) tea.Cmd {
		callbackCalled = true
		return func() tea.Msg {
			return "custom message"
		}
	})

	_, cmd := ti.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})

	if !callbackCalled {
		t.Error("OnSubmit callback should be called")
	}

	if cmd == nil {
		t.Error("OnSubmit should return the command from callback")
	}
}

func TestTextInputNoCallbackNoError(t *testing.T) {
	ti := NewTextInput()
	ti.Focus()
	ti.SetValue("Test")

	// Don't set OnSubmit callback
	_, cmd := ti.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})

	// Should not panic, just return nil
	if cmd != nil {
		t.Error("Without callback, command should be nil")
	}

	// Value should still be cleared
	if ti.Value() != "" {
		t.Error("Value should be cleared even without callback")
	}
}

func TestTextInputViewWithContent(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("Hello\nWorld")
	ti.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := ti.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	// View should contain borders and content
	lineCount := strings.Count(view, "\n")
	if lineCount < 3 {
		t.Errorf("View should have at least 3 lines (top border + content + bottom border), got %d", lineCount)
	}
}

func TestTextInputInit(t *testing.T) {
	ti := NewTextInput()
	cmd := ti.Init()

	if cmd == nil {
		t.Error("Init should return a blink command")
	}
}

func TestTextInputPlaceholder(t *testing.T) {
	ti := NewTextInput()

	if ti.placeholder == "" {
		t.Error("TextInput should have a default placeholder")
	}

	if !strings.Contains(ti.placeholder, "Ctrl+J") {
		t.Error("Placeholder should mention Ctrl+J for sending")
	}
}

func TestTextInputVeryLongText(t *testing.T) {
	ti := NewTextInput()
	ti.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Set very long text
	longText := strings.Repeat("This is a very long line of text. ", 100)
	ti.SetValue(longText)

	// Should not panic
	view := ti.View()
	if view == "" {
		t.Error("View should not be empty with long text")
	}
}

func TestTextInputNarrowWidth(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("Test")
	ti.Update(tea.WindowSizeMsg{Width: 20, Height: 24})

	// Should not panic with narrow width
	view := ti.View()
	if view == "" {
		t.Error("View should not be empty with narrow width")
	}

	// Check that borders exist
	if !strings.Contains(view, "┌") || !strings.Contains(view, "└") {
		t.Error("View should still have borders with narrow width")
	}
}

func TestTextInputAltEnterSubmit(t *testing.T) {
	ti := NewTextInput()
	ti.Focus()
	ti.SetValue("Test message")

	submitted := false
	ti.OnSubmit(func(text string) tea.Cmd {
		submitted = true
		return nil
	})

	// Send Alt+Enter (alternative submit key)
	ti.Update(tea.KeyMsg{Type: tea.KeyEnter, Alt: true})

	if !submitted {
		t.Error("Alt+Enter should trigger submit")
	}
}

func TestTextInputRegularEnterNoSubmit(t *testing.T) {
	ti := NewTextInput()
	ti.Focus()
	ti.SetValue("Test")

	submitted := false
	ti.OnSubmit(func(text string) tea.Cmd {
		submitted = true
		return nil
	})

	// Send regular Enter (should be handled by textarea for newline)
	ti.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if submitted {
		t.Error("Regular Enter should not trigger submit")
	}
}
