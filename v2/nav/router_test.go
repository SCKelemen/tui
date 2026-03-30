package nav

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type testRouteModel struct {
	viewText string
	updates  int
	inits    int
	width    int
	height   int
}

func (m *testRouteModel) Init() tea.Cmd {
	m.inits++
	return nil
}

func (m *testRouteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.updates++
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = ws.Width
		m.height = ws.Height
	}
	return m, nil
}

func (m *testRouteModel) View() string {
	return m.viewText
}

func TestRouterAddRouteNavigateCurrentAndView(t *testing.T) {
	home := &testRouteModel{viewText: "home view"}
	settings := &testRouteModel{viewText: "settings view"}

	r := NewRouter()
	r.AddRoute("home", home).AddRoute("settings", settings)

	if got := r.Current(); got != "home" {
		t.Fatalf("expected first added route to be current, got %q", got)
	}

	if got := r.View(); got != "home view" {
		t.Fatalf("expected home view, got %q", got)
	}

	navCmd := r.Navigate("settings", map[string]string{"section": "advanced"})
	if navCmd == nil {
		t.Fatal("Navigate should return command")
	}

	navMsg := navCmd()
	_, _ = r.Update(navMsg)

	if got := r.Current(); got != "settings" {
		t.Fatalf("expected current route settings, got %q", got)
	}

	if got := r.View(); got != "settings view" {
		t.Fatalf("expected settings view, got %q", got)
	}

	if got := r.CurrentParams()["section"]; got != "advanced" {
		t.Fatalf("expected params to contain section=advanced, got %q", got)
	}
}

func TestRouterForwardsWindowSizeAndBack(t *testing.T) {
	home := &testRouteModel{viewText: "home"}
	settings := &testRouteModel{viewText: "settings"}

	r := NewRouter()
	r.AddRoute("home", home).AddRoute("settings", settings)

	_, _ = r.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	_, _ = r.Update(r.Navigate("settings")())
	if settings.width != 80 || settings.height != 24 {
		t.Fatalf("expected routed model to receive window size 80x24, got %dx%d", settings.width, settings.height)
	}

	back := r.Back()
	if back == nil {
		t.Fatal("expected Back command after navigation")
	}
	_, _ = r.Update(back())

	if got := r.Current(); got != "home" {
		t.Fatalf("expected current route home after back, got %q", got)
	}

	if !strings.Contains(r.View(), "home") {
		t.Fatalf("expected home view after back, got %q", r.View())
	}
}
