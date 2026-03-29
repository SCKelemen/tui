package nav

import tea "github.com/charmbracelet/bubbletea"

// Route describes a named route and the Bubble Tea model rendered for it.
type Route struct {
	Name string
	View tea.Model
}

// NavigateMsg requests navigation to another route.
type NavigateMsg struct {
	RouteName string
	Params    map[string]string
}

// Router provides simple view-based routing for Bubble Tea models.
type Router struct {
	routes  map[string]Route
	current string
	history []string
	params  map[string]string
	width   int
	height  int
}

// NewRouter creates a new Router.
func NewRouter() *Router {
	return &Router{
		routes:  make(map[string]Route),
		history: make([]string, 0),
		params:  make(map[string]string),
	}
}

// AddRoute registers a route and returns the router for chaining.
func (r *Router) AddRoute(name string, view tea.Model) *Router {
	if name == "" || view == nil {
		return r
	}

	r.routes[name] = Route{Name: name, View: view}
	if r.current == "" {
		r.current = name
	}

	return r
}

// Navigate returns a command that sends a NavigateMsg for the requested route.
func (r *Router) Navigate(name string, params ...map[string]string) tea.Cmd {
	routeParams := map[string]string{}
	if len(params) > 0 && params[0] != nil {
		routeParams = copyParams(params[0])
	}

	return func() tea.Msg {
		return NavigateMsg{
			RouteName: name,
			Params:    routeParams,
		}
	}
}

// Back returns a command that navigates to the previous route in history.
func (r *Router) Back() tea.Cmd {
	if len(r.history) == 0 {
		return nil
	}

	prev := r.history[len(r.history)-1]
	return r.Navigate(prev)
}

// Current returns the current route name.
func (r *Router) Current() string {
	return r.current
}

// CurrentParams returns a copy of the current route parameters.
func (r *Router) CurrentParams() map[string]string {
	return copyParams(r.params)
}

// Init initializes the current route model.
func (r *Router) Init() tea.Cmd {
	route, ok := r.routes[r.current]
	if !ok || route.View == nil {
		return nil
	}
	return route.View.Init()
}

// Update handles navigation and forwards all other messages to the active route.
func (r *Router) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width = msg.Width
		r.height = msg.Height
		return r.forwardToCurrent(msg)
	case NavigateMsg:
		return r.handleNavigate(msg)
	default:
		return r.forwardToCurrent(msg)
	}
}

// View renders the current route's view.
func (r *Router) View() string {
	route, ok := r.routes[r.current]
	if !ok || route.View == nil {
		return ""
	}

	return route.View.View()
}

func (r *Router) handleNavigate(msg NavigateMsg) (tea.Model, tea.Cmd) {
	route, ok := r.routes[msg.RouteName]
	if !ok || route.View == nil {
		return r, nil
	}

	if r.current != "" && r.current != msg.RouteName {
		r.history = append(r.history, r.current)
	}

	if len(r.history) > 0 && msg.RouteName == r.history[len(r.history)-1] {
		r.history = r.history[:len(r.history)-1]
	}

	r.current = msg.RouteName
	r.params = copyParams(msg.Params)

	var cmds []tea.Cmd
	if r.width > 0 || r.height > 0 {
		updated, cmd := route.View.Update(tea.WindowSizeMsg{Width: r.width, Height: r.height})
		route.View = updated
		r.routes[msg.RouteName] = route
		cmds = append(cmds, cmd)
	}

	cmds = append(cmds, route.View.Init())
	return r, tea.Batch(cmds...)
}

func (r *Router) forwardToCurrent(msg tea.Msg) (tea.Model, tea.Cmd) {
	route, ok := r.routes[r.current]
	if !ok || route.View == nil {
		return r, nil
	}

	updated, cmd := route.View.Update(msg)
	route.View = updated
	r.routes[r.current] = route

	return r, cmd
}

func copyParams(params map[string]string) map[string]string {
	if params == nil {
		return map[string]string{}
	}

	copied := make(map[string]string, len(params))
	for k, v := range params {
		copied[k] = v
	}
	return copied
}
