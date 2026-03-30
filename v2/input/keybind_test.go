package input

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type keybindTestMsg struct{}

func TestKeybindManagerRegisterLookup(t *testing.T) {
	km := NewKeybindManager()
	km.Register(Keybind{Key: "ctrl+s", Description: "save"})

	cmd, ok := km.Handle("ctrl+s")
	if !ok {
		t.Fatal("expected registered keybind to be handled")
	}

	if cmd != nil {
		t.Fatal("expected nil command for keybind without action")
	}

	_, ok = km.Handle("ctrl+x")
	if ok {
		t.Fatal("expected unknown keybind to not be handled")
	}
}

func TestKeybindManagerHandlerInvocation(t *testing.T) {
	km := NewKeybindManager()
	called := false

	km.Register(Keybind{
		Key: "enter",
		Action: func() tea.Cmd {
			called = true
			return func() tea.Msg { return keybindTestMsg{} }
		},
	})

	cmd, ok := km.HandleMsg(tea.KeyMsg{Type: tea.KeyEnter})
	if !ok {
		t.Fatal("expected keybind to be handled")
	}
	if cmd == nil {
		t.Fatal("expected command from action")
	}

	if !called {
		t.Fatal("expected action to be invoked")
	}

	if _, ok := cmd().(keybindTestMsg); !ok {
		t.Fatalf("expected keybindTestMsg from command, got %T", cmd())
	}
}
