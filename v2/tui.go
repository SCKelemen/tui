package tui

import (
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Component is the interface all TUI components must implement.
type Component interface {
	// Init initializes the component.
	Init() tea.Cmd

	// Update handles messages and updates component state.
	Update(msg tea.Msg) (Component, tea.Cmd)

	// View renders the component.
	View() string

	// Focus is called when this component receives focus.
	Focus()

	// Blur is called when this component loses focus.
	Blur()

	// Focused returns whether this component is currently focused.
	Focused() bool
}

// ApplicationOption configures a new Application.
type ApplicationOption func(*Application)

// WithQuitKey sets the quit key for the application.
func WithQuitKey(key string) ApplicationOption {
	return func(a *Application) {
		a.SetQuitKey(key)
	}
}

// Application represents the main TUI application.
type Application struct {
	width      int
	height     int
	components []Component
	focused    int // Index of currently focused component.
	quitKey    string
	frameBuf   *FrameBuffer
}

// NewApplication creates a new TUI application.
func NewApplication(opts ...ApplicationOption) *Application {
	a := &Application{
		components: make([]Component, 0),
		focused:    -1,
		quitKey:    "ctrl+c",
		frameBuf:   NewFrameBuffer(io.Discard, 0, 0),
	}
	for _, opt := range opts {
		opt(a)
	}

	return a
}

// SetQuitKey updates the key used to quit the application.
func (a *Application) SetQuitKey(key string) {
	a.quitKey = strings.ToLower(strings.TrimSpace(key))
}

// AddComponent adds a component to the application.
func (a *Application) AddComponent(c Component) {
	a.components = append(a.components, c)
	if a.focused == -1 && len(a.components) > 0 {
		a.focused = 0
		a.components[0].Focus()
	}
}

// FocusComponent focuses a specific component by index, blurring the currently focused one.
func (a *Application) FocusComponent(index int) {
	if index < 0 || index >= len(a.components) {
		return
	}

	if a.focused >= 0 && a.focused < len(a.components) {
		a.components[a.focused].Blur()
	}

	a.focused = index
	a.components[index].Focus()
}

// Init initializes the application.
func (a *Application) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, c := range a.components {
		cmds = append(cmds, c.Init())
	}
	return tea.Batch(cmds...)
}

// Update handles messages.
func (a *Application) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		if a.frameBuf != nil {
			a.frameBuf.Resize(msg.Width, msg.Height)
		}

		var cmds []tea.Cmd
		for i, c := range a.components {
			var cmd tea.Cmd
			a.components[i], cmd = c.Update(msg)
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)
	case tea.MouseMsg:
		if a.focused >= 0 && a.focused < len(a.components) {
			var cmd tea.Cmd
			a.components[a.focused], cmd = a.components[a.focused].Update(msg)
			return a, cmd
		}
		return a, nil
	case tea.KeyMsg:
		// Always forward keys to the focused component first.
		if a.focused >= 0 && a.focused < len(a.components) {
			var cmd tea.Cmd
			a.components[a.focused], cmd = a.components[a.focused].Update(msg)
			if cmd != nil {
				return a, cmd
			}
		}

		switch msg.String() {
		case "tab":
			return a, a.focusNext()
		case "shift+tab":
			return a, a.focusPrev()
		}

		if a.quitKey != "" && msg.String() == a.quitKey {
			return a, tea.Quit
		}

		return a, nil
	}

	if a.focused >= 0 && a.focused < len(a.components) {
		var cmd tea.Cmd
		a.components[a.focused], cmd = a.components[a.focused].Update(msg)
		return a, cmd
	}

	return a, nil
}

// View renders the application.
func (a *Application) View() string {
	frame := "No components"
	if len(a.components) > 0 {
		var view string
		for _, c := range a.components {
			view += c.View()
		}
		frame = view
	}

	if a.frameBuf == nil {
		a.frameBuf = NewFrameBuffer(io.Discard, a.width, a.height)
	}

	return a.frameBuf.Render(frame)
}
func (a *Application) focusNext() tea.Cmd {
	if len(a.components) == 0 {
		return nil
	}

	if a.focused >= 0 {
		a.components[a.focused].Blur()
	}

	a.focused = (a.focused + 1) % len(a.components)
	a.components[a.focused].Focus()

	return nil
}

func (a *Application) focusPrev() tea.Cmd {
	if len(a.components) == 0 {
		return nil
	}

	if a.focused >= 0 {
		a.components[a.focused].Blur()
	}

	a.focused--
	if a.focused < 0 {
		a.focused = len(a.components) - 1
	}
	a.components[a.focused].Focus()

	return nil
}
