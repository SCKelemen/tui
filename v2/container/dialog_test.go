package container

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var _ tea.Model = (*Dialog)(nil)
var _ tea.Model = (*DialogManager)(nil)
var _ tea.Model = (*DialogAlert)(nil)
var _ tea.Model = (*DialogConfirm)(nil)

type mockDialogContent struct {
	initCalled bool
	updateMsgs []tea.Msg
	view       string
}

func (m *mockDialogContent) Init() tea.Cmd {
	m.initCalled = true
	return func() tea.Msg { return "init" }
}

func (m *mockDialogContent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.updateMsgs = append(m.updateMsgs, msg)
	return m, func() tea.Msg { return "updated" }
}

func (m *mockDialogContent) View() string {
	if m.view == "" {
		return "mock dialog content"
	}
	return m.view
}

func TestDialogCreationAndOptions(t *testing.T) {
	content := &mockDialogContent{view: "body"}
	d := NewDialog(
		content,
		WithDialogID("dlg-1"),
		WithDialogTitle("Test Dialog"),
		WithDialogSize(DialogLarge),
	)

	if d == nil {
		t.Fatal("NewDialog returned nil")
	}
	if d.ID != "dlg-1" {
		t.Errorf("expected ID dlg-1, got %q", d.ID)
	}
	if d.Title != "Test Dialog" {
		t.Errorf("expected title Test Dialog, got %q", d.Title)
	}
	if d.Size != DialogLarge {
		t.Errorf("expected size DialogLarge, got %v", d.Size)
	}
	if d.Content != content {
		t.Error("expected content model to be assigned")
	}

	empty := NewDialog(nil)
	if empty.Content == nil {
		t.Fatal("expected nil content to be replaced with empty model")
	}
	if _, ok := empty.Content.(*dialogEmptyModel); !ok {
		t.Fatalf("expected dialogEmptyModel, got %T", empty.Content)
	}
}

func TestDialogInitUpdateAndView(t *testing.T) {
	content := &mockDialogContent{view: "hello from content"}
	d := NewDialog(content)

	if cmd := d.Init(); cmd == nil {
		t.Error("expected Init to return content init command")
	}
	if !content.initCalled {
		t.Error("expected wrapped content Init to be called")
	}

	model, cmd := d.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if model != d {
		t.Error("expected Update to return same dialog instance")
	}
	if cmd == nil {
		t.Error("expected Update to return child update command")
	}
	if d.width != 100 || d.height != 40 {
		t.Errorf("expected size 100x40, got %dx%d", d.width, d.height)
	}

	view := d.View()
	if strings.TrimSpace(view) == "" {
		t.Fatal("expected non-empty view")
	}
	if !strings.Contains(view, "Dialog") {
		t.Error("expected default dialog title in view")
	}
	if !strings.Contains(view, "hello from content") {
		t.Error("expected child content in dialog view")
	}
}

func TestDialogManagerShowCloseAndEscape(t *testing.T) {
	mgr := NewDialogManager()
	if mgr == nil {
		t.Fatal("NewDialogManager returned nil")
	}
	if mgr.Init() != nil {
		t.Error("DialogManager Init should return nil")
	}

	mgr.Update(tea.WindowSizeMsg{Width: 120, Height: 32})

	d := NewDialog(&mockDialogContent{view: "managed body"}, WithDialogTitle("Managed Dialog"))
	_, cmd := mgr.Update(ShowDialogMsg{Dialog: d})
	if cmd == nil {
		t.Error("expected show dialog to return dialog init command")
	}
	if !mgr.IsOpen() {
		t.Fatal("expected manager to have open dialog")
	}
	if mgr.Top() == nil {
		t.Fatal("expected top dialog to be set")
	}
	if mgr.Top().width != 120 || mgr.Top().height != 32 {
		t.Errorf("expected top dialog to receive manager size, got %dx%d", mgr.Top().width, mgr.Top().height)
	}

	view := mgr.View()
	if strings.TrimSpace(view) == "" {
		t.Error("expected manager view to be non-empty while dialog is open")
	}

	mgr.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if mgr.IsOpen() {
		t.Error("expected esc to close top dialog")
	}

	mgr.Update(CloseDialogMsg{})
	if mgr.IsOpen() {
		t.Error("expected manager to remain closed")
	}
}

func TestDialogAlertKeyResults(t *testing.T) {
	alert := NewDialogAlert("Delete", "Are you sure?")
	if alert.Init() != nil {
		t.Error("DialogAlert Init should return nil")
	}
	if strings.TrimSpace(alert.View()) == "" {
		t.Error("DialogAlert view should be non-empty")
	}

	_, enterCmd := alert.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if enterCmd == nil {
		t.Fatal("expected enter to emit result command")
	}
	msg := enterCmd()
	result, ok := msg.(DialogResultMsg)
	if !ok {
		t.Fatalf("expected DialogResultMsg, got %T", msg)
	}
	if result.ID != "Delete" || result.Cancelled {
		t.Errorf("unexpected enter result: %+v", result)
	}

	_, escCmd := alert.Update(tea.KeyMsg{Type: tea.KeyEsc})
	msg = escCmd()
	result, ok = msg.(DialogResultMsg)
	if !ok {
		t.Fatalf("expected DialogResultMsg on esc, got %T", msg)
	}
	if !result.Cancelled {
		t.Errorf("expected esc to cancel, got %+v", result)
	}
}

func TestDialogConfirmSelectionAndResult(t *testing.T) {
	confirm := NewDialogConfirm("Exit", "Exit now?")
	if confirm.Init() != nil {
		t.Error("DialogConfirm Init should return nil")
	}
	if strings.TrimSpace(confirm.View()) == "" {
		t.Error("DialogConfirm view should be non-empty")
	}

	confirm.Update(tea.KeyMsg{Type: tea.KeyRight})
	_, cmd := confirm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected enter to emit result")
	}
	msg := cmd()
	result, ok := msg.(DialogResultMsg)
	if !ok {
		t.Fatalf("expected DialogResultMsg, got %T", msg)
	}
	if result.Value != false || !result.Cancelled {
		t.Errorf("expected cancel selection result, got %+v", result)
	}

	confirm.Update(tea.KeyMsg{Type: tea.KeyLeft})
	_, cmd = confirm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	msg = cmd()
	result = msg.(DialogResultMsg)
	if result.Value != true || result.Cancelled {
		t.Errorf("expected confirm selection result, got %+v", result)
	}

	confirm.Update(tea.KeyMsg{Type: tea.KeyTab})
	if confirm.selected != 1 {
		t.Errorf("expected tab to move selection to 1, got %d", confirm.selected)
	}
}
