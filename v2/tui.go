package tui

import tea "github.com/charmbracelet/bubbletea"

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

// Application represents the main TUI application.
type Application struct {
	width      int
	height     int
	components []Component
	focused    int // Index of currently focused component.
}

// NewApplication creates a new TUI application.
func NewApplication() *Application {
	return &Application{
		components: make([]Component, 0),
		focused:    -1,
	}
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
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return a, tea.Quit
		case "tab":
			return a, a.focusNext()
		case "shift+tab":
			return a, a.focusPrev()
		case "ctrl+c":
			// Forward ctrl+c to focused component first (e.g. copy selection).
			// If the component returns a command, honor it; otherwise quit.
			if a.focused >= 0 && a.focused < len(a.components) {
				var cmd tea.Cmd
				a.components[a.focused], cmd = a.components[a.focused].Update(msg)
				if cmd != nil {
					return a, cmd
				}
			}
			return a, tea.Quit
		}
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

		var cmds []tea.Cmd
		for i, c := range a.components {
			var cmd tea.Cmd
			a.components[i], cmd = c.Update(msg)
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)
	}

	// Forward all other messages (mouse events, remaining keys) to focused component.
	if a.focused >= 0 && a.focused < len(a.components) {
		var cmd tea.Cmd
		a.components[a.focused], cmd = a.components[a.focused].Update(msg)
		return a, cmd
	}

	return a, nil
}

// View renders the application.
func (a *Application) View() string {
	if len(a.components) == 0 {
		return "No components"
	}

	var view string
	for _, c := range a.components {
		view += c.View()
	}
	return view
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
