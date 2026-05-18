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

// Bounded is an optional interface a Component can implement to advertise
// the screen rectangle (in cells, top-left origin) where its current view
// is rendered. Application uses Bounds to hit-test mouse clicks so that
// pressing in a component's rect refocuses that component and forwards the
// event. Components rendered through a container that does its own layout
// can either implement Bounded themselves or have their bounds registered
// via Application.SetComponentBounds.
type Bounded interface {
	Bounds() (x, y, w, h int)
}

// KeyConsumer is an optional interface a Component can implement to claim
// keys that would otherwise be handled by Application-level shortcuts such
// as Tab, Shift+Tab and the quit key.
//
// When the focused component reports HandlesKey(k) == true for a given
// key string (as produced by tea.KeyMsg.String), Application will deliver
// the key to the component and skip its own shortcut routing for that
// key. This lets, for example, a focused text input bind Tab without
// having the Application steal it for focus cycling.
//
// Components that do not implement KeyConsumer keep the legacy behavior:
// keys are forwarded to the focused component, and Application shortcuts
// still fire as usual.
type KeyConsumer interface {
	HandlesKey(key string) bool
}

// ApplicationOption configures a new Application.
type ApplicationOption func(*Application)

// WithQuitKey sets the quit key for the application.
func WithQuitKey(key string) ApplicationOption {
	return func(a *Application) {
		a.SetQuitKey(key)
	}
}

// componentRect holds an explicit, externally registered bounding box for a
// component that does not implement the Bounded interface.
type componentRect struct {
	x, y, w, h int
	set        bool
}

// Application represents the main TUI application.
type Application struct {
	width      int
	height     int
	components []Component
	bounds     []componentRect // parallel to components; bounds[i].set toggles validity
	focused    int             // Index of currently focused component.
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
	a.bounds = append(a.bounds, componentRect{})
	if a.focused == -1 && len(a.components) > 0 {
		a.focused = 0
		a.components[0].Focus()
	}
}

// SetComponentBounds registers the screen rectangle for the component at
// index i. Bounds are used for mouse hit-testing — a press inside the rect
// refocuses that component and routes the event to it. Pass a non-positive
// width or height to clear previously registered bounds.
//
// Components that implement the Bounded interface do not need to call
// SetComponentBounds; bounds reported by Bounded.Bounds take precedence
// over the registered rectangle when both are present.
func (a *Application) SetComponentBounds(i, x, y, w, h int) {
	if i < 0 || i >= len(a.bounds) {
		return
	}
	if w <= 0 || h <= 0 {
		a.bounds[i] = componentRect{}
		return
	}
	a.bounds[i] = componentRect{x: x, y: y, w: w, h: h, set: true}
}

// componentBoundsAt returns the registered or component-reported bounds for
// the component at index i, plus whether bounds were available.
func (a *Application) componentBoundsAt(i int) (x, y, w, h int, ok bool) {
	if i < 0 || i >= len(a.components) {
		return 0, 0, 0, 0, false
	}
	if b, has := a.components[i].(Bounded); has {
		bx, by, bw, bh := b.Bounds()
		if bw > 0 && bh > 0 {
			return bx, by, bw, bh, true
		}
	}
	if r := a.bounds[i]; r.set {
		return r.x, r.y, r.w, r.h, true
	}
	return 0, 0, 0, 0, false
}

// pointInBounds reports whether (x, y) is within the rectangle described by
// (bx, by, bw, bh) using inclusive top-left, exclusive bottom-right.
func pointInBounds(x, y, bx, by, bw, bh int) bool {
	return x >= bx && y >= by && x < bx+bw && y < by+bh
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
		// Hit-test presses so a click in another component's rect refocuses
		// that component and routes the message to it. Motion and release
		// stay with the currently focused component — drag tracking is a
		// larger feature.
		if msg.Action == tea.MouseActionPress {
			for i := range a.components {
				bx, by, bw, bh, ok := a.componentBoundsAt(i)
				if !ok {
					continue
				}
				if !pointInBounds(msg.X, msg.Y, bx, by, bw, bh) {
					continue
				}
				if i != a.focused {
					a.FocusComponent(i)
				}
				var cmd tea.Cmd
				a.components[i], cmd = a.components[i].Update(msg)
				return a, cmd
			}
		}

		if a.focused >= 0 && a.focused < len(a.components) {
			var cmd tea.Cmd
			a.components[a.focused], cmd = a.components[a.focused].Update(msg)
			return a, cmd
		}
		return a, nil
	case tea.KeyMsg:
		key := msg.String()

		// Always forward keys to the focused component first.
		var focusedCmd tea.Cmd
		claimed := false
		if a.focused >= 0 && a.focused < len(a.components) {
			focused := a.components[a.focused]
			a.components[a.focused], focusedCmd = focused.Update(msg)
			if kc, ok := focused.(KeyConsumer); ok && kc.HandlesKey(key) {
				claimed = true
			}
		}

		// If the focused component explicitly claimed the key, do not let
		// Application-level shortcuts intercept it.
		if claimed {
			return a, focusedCmd
		}

		// Preserve legacy behavior: a non-nil command from the focused
		// component short-circuits Application shortcuts as well.
		if focusedCmd != nil {
			return a, focusedCmd
		}

		switch key {
		case "tab":
			return a, a.focusNext()
		case "shift+tab":
			return a, a.focusPrev()
		}

		if a.quitKey != "" && key == a.quitKey {
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
