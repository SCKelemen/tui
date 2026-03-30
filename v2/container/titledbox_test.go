package container

import (
	"strings"
	"testing"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

var _ tui.Component = (*TitledBox)(nil)

func TestTitledBoxCreationAndOptions(t *testing.T) {
	b := NewTitledBox(
		"Header",
		"Line 1",
		WithTitledBoxWidth(42),
		WithTitledBoxBorderColor(style.ANSICyan),
		WithTitledBoxTitleColor(style.ANSIGreen),
	)

	if b == nil {
		t.Fatal("NewTitledBox returned nil")
	}
	if b.width != 42 {
		t.Errorf("expected width 42, got %d", b.width)
	}
	if b.borderColor != style.ANSICyan {
		t.Errorf("expected border color to be set, got %q", b.borderColor)
	}
	if b.titleColor != style.ANSIGreen {
		t.Errorf("expected title color to be set, got %q", b.titleColor)
	}

	defaults := NewTitledBox("", "")
	if defaults.width != 60 {
		t.Errorf("expected default width 60, got %d", defaults.width)
	}
}

func TestTitledBoxInitUpdateAndView(t *testing.T) {
	b := NewTitledBox("Service Status", "running")

	if b.Init() != nil {
		t.Error("TitledBox Init should return nil")
	}

	model, cmd := b.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if model != b {
		t.Error("expected Update to return same titled box")
	}
	if cmd != nil {
		t.Error("expected Update to return nil cmd")
	}

	view := b.View()
	if strings.TrimSpace(view) == "" {
		t.Fatal("expected non-empty titled box view")
	}
	if !strings.Contains(view, "Service Status") {
		t.Error("expected title in rendered view")
	}
	if !strings.Contains(view, "running") {
		t.Error("expected content in rendered view")
	}
	if !strings.Contains(view, "│") {
		t.Error("expected vertical borders in rendered view")
	}
}

func TestTitledBoxRenderHelperAndFocus(t *testing.T) {
	rendered := RenderTitledBox("Logs", "entry", 30)
	if strings.TrimSpace(rendered) == "" {
		t.Fatal("expected RenderTitledBox output to be non-empty")
	}
	if !strings.Contains(rendered, "Logs") {
		t.Error("expected helper output to contain title")
	}

	b := NewTitledBox("Focus", "test")
	if b.Focused() {
		t.Error("expected default focused=false")
	}
	b.Focus()
	if !b.Focused() {
		t.Error("expected Focus to set focused=true")
	}
	b.Blur()
	if b.Focused() {
		t.Error("expected Blur to set focused=false")
	}
}
