package container

import (
	"strings"
	"testing"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	tea "github.com/charmbracelet/bubbletea"
)

var _ tui.Component = (*PermissionDialog)(nil)

func TestPermissionDialogCreationAndOptions(t *testing.T) {
	request := PermissionRequest{
		ToolName:    "terminal.run",
		Description: "Execute shell command",
		Command:     "rm -rf /tmp/example",
		Risk:        "high",
	}
	tokens := design.DefaultTheme()
	exceptions := []string{"/tmp", "./sandbox"}

	d := NewPermissionDialog(
		request,
		WithPermissionDialogWidth(90),
		WithPermissionDialogDesignTokens(tokens),
		WithPermissionDialogExceptions(exceptions),
	)

	if d == nil {
		t.Fatal("NewPermissionDialog returned nil")
	}
	if d.width != 90 {
		t.Errorf("expected width 90, got %d", d.width)
	}
	if d.cursor != 0 {
		t.Errorf("expected default cursor 0, got %d", d.cursor)
	}
	if !d.visible {
		t.Error("permission dialog should be visible by default")
	}
	if len(d.exceptions) != 2 {
		t.Errorf("expected 2 exceptions, got %d", len(d.exceptions))
	}
	if d.colors != permissionDialogColorsFromTokens(tokens) {
		t.Error("expected colors to be derived from provided design tokens")
	}
}

func TestPermissionDialogInitViewAndWindowSize(t *testing.T) {
	d := NewPermissionDialog(PermissionRequest{ToolName: "git.push", Description: "Push branch", Command: "git push", Risk: "medium"})
	if d.Init() != nil {
		t.Error("PermissionDialog Init should return nil")
	}

	d.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if d.windowWidth != 80 {
		t.Errorf("expected windowWidth 80, got %d", d.windowWidth)
	}

	view := d.View()
	if strings.TrimSpace(view) == "" {
		t.Fatal("expected non-empty view")
	}
	if !strings.Contains(view, "Permission Required") {
		t.Error("expected warning title in view")
	}
	if !strings.Contains(view, "git.push") {
		t.Error("expected tool name in view")
	}
}

func TestPermissionDialogAllowAlwaysAndDeny(t *testing.T) {
	d := NewPermissionDialog(
		PermissionRequest{ToolName: "terminal.run", Description: "Run", Command: "echo hi", Risk: "low"},
		WithPermissionDialogExceptions([]string{"./allowed"}),
	)

	// Without focus, key presses should be ignored.
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected enter to be ignored when dialog is not focused")
	}

	d.Focus()
	if !d.Focused() {
		t.Fatal("expected dialog to be focused")
	}

	// Allow once.
	d.cursor = 0
	_, cmd = d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected enter to emit allow result")
	}
	msg := cmd()
	resMsg, ok := msg.(PermissionResultMsg)
	if !ok {
		t.Fatalf("expected PermissionResultMsg, got %T", msg)
	}
	if !resMsg.Result.Allowed || resMsg.Result.AlwaysAllow {
		t.Errorf("unexpected allow-once result: %+v", resMsg.Result)
	}
	if len(resMsg.Result.Exceptions) != 1 || resMsg.Result.Exceptions[0] != "./allowed" {
		t.Errorf("unexpected exceptions: %+v", resMsg.Result.Exceptions)
	}

	// Always allow.
	d.cursor = 0
	d.Update(tea.KeyMsg{Type: tea.KeyDown})
	if d.cursor != 1 {
		t.Fatalf("expected cursor at always-allow option, got %d", d.cursor)
	}
	_, cmd = d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	resMsg = cmd().(PermissionResultMsg)
	if !resMsg.Result.Allowed || !resMsg.Result.AlwaysAllow {
		t.Errorf("unexpected always-allow result: %+v", resMsg.Result)
	}

	// Deny via Esc.
	_, cmd = d.Update(tea.KeyMsg{Type: tea.KeyEsc})
	resMsg = cmd().(PermissionResultMsg)
	if resMsg.Result.Allowed || resMsg.Result.AlwaysAllow {
		t.Errorf("expected deny result on esc, got %+v", resMsg.Result)
	}
	if d.cursor != 2 {
		t.Errorf("expected cursor to move to deny on esc, got %d", d.cursor)
	}
}

func TestPermissionDialogVisibilityAndFocus(t *testing.T) {
	d := NewPermissionDialog(PermissionRequest{})
	d.Hide()
	if d.IsVisible() {
		t.Error("expected dialog to be hidden")
	}
	if d.View() != "" {
		t.Error("expected hidden dialog view to be empty")
	}

	d.Show()
	if !d.IsVisible() {
		t.Error("expected dialog to be visible after Show")
	}

	d.Focus()
	if !d.Focused() {
		t.Error("expected Focus to set focused=true")
	}
	d.Blur()
	if d.Focused() {
		t.Error("expected Blur to set focused=false")
	}
}
