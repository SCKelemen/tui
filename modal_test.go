package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModalCreation(t *testing.T) {
	modal := NewModal()
	if modal == nil {
		t.Fatal("NewModal returned nil")
	}

	if modal.visible {
		t.Error("Modal should not be visible on creation")
	}
}

func TestModalShow(t *testing.T) {
	modal := NewModal()
	modal.Show()

	if !modal.visible {
		t.Error("Modal should be visible after Show()")
	}
}

func TestModalHide(t *testing.T) {
	modal := NewModal()
	modal.Show()
	modal.Hide()

	if modal.visible {
		t.Error("Modal should not be visible after Hide()")
	}
}

func TestModalAlert(t *testing.T) {
	modal := NewModal()
	modal.Focus()
	modal.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	called := false
	modal.ShowAlert("Test", "Test message", func() tea.Cmd {
		called = true
		return nil
	})

	if !modal.IsVisible() {
		t.Error("Modal should be visible after ShowAlert")
	}

	// Simulate Enter key
	modal.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !called {
		t.Error("Alert callback was not called")
	}
}

func TestModalConfirm(t *testing.T) {
	modal := NewModal()
	modal.Focus()
	modal.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	yesCalled := false
	noCalled := false

	modal.ShowConfirm(
		"Test",
		"Confirm?",
		func() tea.Cmd {
			yesCalled = true
			return nil
		},
		func() tea.Cmd {
			noCalled = true
			return nil
		},
	)

	// Press Enter (should call Yes)
	modal.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !yesCalled {
		t.Error("Yes callback was not called")
	}

	// Test No button
	yesCalled = false
	modal.ShowConfirm(
		"Test",
		"Confirm?",
		func() tea.Cmd {
			yesCalled = true
			return nil
		},
		func() tea.Cmd {
			noCalled = true
			return nil
		},
	)

	// Move to No button (Tab once)
	modal.Update(tea.KeyMsg{Type: tea.KeyTab})
	// Press Enter (should call No)
	modal.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !noCalled {
		t.Error("No callback was not called")
	}
}

func TestModalInput(t *testing.T) {
	modal := NewModal()
	modal.Focus()
	modal.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	var receivedValue string
	modal.ShowInput(
		"Test",
		"Enter value:",
		"Placeholder",
		func(value string) tea.Cmd {
			receivedValue = value
			return nil
		},
		func() tea.Cmd {
			return nil
		},
	)

	if !modal.hasInput {
		t.Error("Modal should have input field")
	}

	// Type some text
	modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test input")})

	// Press Enter
	modal.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if receivedValue == "" {
		t.Error("Input value was not received")
	}
}

func TestModalView(t *testing.T) {
	modal := NewModal()
	modal.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	modal.ShowAlert("Test Title", "Test message", nil)

	view := modal.View()

	// Check for rounded corners
	if !strings.Contains(view, "╭") {
		t.Error("Modal should contain top-left rounded corner (╭)")
	}
	if !strings.Contains(view, "╮") {
		t.Error("Modal should contain top-right rounded corner (╮)")
	}
	if !strings.Contains(view, "╰") {
		t.Error("Modal should contain bottom-left rounded corner (╰)")
	}
	if !strings.Contains(view, "╯") {
		t.Error("Modal should contain bottom-right rounded corner (╯)")
	}

	// Check for title in the border
	if !strings.Contains(view, "Test Title") {
		t.Error("Modal should contain title")
	}

	// Check for message
	if !strings.Contains(view, "Test message") {
		t.Error("Modal should contain message")
	}
}

func TestModalButtonNavigation(t *testing.T) {
	modal := NewModal()
	modal.Focus()
	modal.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	modal.ShowConfirm("Test", "Message", nil, nil)

	// Initial selection should be 0
	if modal.selected != 0 {
		t.Errorf("Initial selection should be 0, got %d", modal.selected)
	}

	// Press Tab to move to next button
	modal.Update(tea.KeyMsg{Type: tea.KeyTab})
	if modal.selected != 1 {
		t.Errorf("Selection should be 1 after Tab, got %d", modal.selected)
	}

	// Press Tab again (should wrap to 0)
	modal.Update(tea.KeyMsg{Type: tea.KeyTab})
	if modal.selected != 0 {
		t.Errorf("Selection should wrap to 0, got %d", modal.selected)
	}

	// Press Shift+Tab to move backwards
	modal.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if modal.selected != 1 {
		t.Errorf("Selection should be 1 after Shift+Tab, got %d", modal.selected)
	}
}

func TestModalEscapeKey(t *testing.T) {
	modal := NewModal()
	modal.Focus()
	modal.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	modal.ShowAlert("Test", "Message", nil)

	// Press Escape
	modal.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if modal.IsVisible() {
		t.Error("Modal should be hidden after Escape")
	}
}
