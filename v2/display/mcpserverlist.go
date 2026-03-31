package display

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// MCPServerScope groups MCP servers by source scope.
type MCPServerScope string

const (
	ScopeProject    MCPServerScope = "project"
	ScopeUser       MCPServerScope = "user"
	ScopeLocal      MCPServerScope = "local"
	ScopeEnterprise MCPServerScope = "enterprise"
	ScopeDynamic    MCPServerScope = "dynamic"
)

// MCPServerInfo is one MCP server entry.
type MCPServerInfo struct {
	Name      string
	URL       string
	Scope     MCPServerScope
	ToolCount int
	Connected bool
}

// MCPServerSelectedMsg is emitted when Enter confirms a selected server.
type MCPServerSelectedMsg struct {
	Server MCPServerInfo
	Index  int
}

// MCPServerListOption configures an MCPServerList.
type MCPServerListOption func(*MCPServerList)

// WithMCPServerListWidth sets fixed width.
func WithMCPServerListWidth(width int) MCPServerListOption {
	return func(l *MCPServerList) {
		if width > 0 {
			l.width = width
		}
	}
}

// WithMCPServerListDesignTokens applies design-system colors.
func WithMCPServerListDesignTokens(tokens *design.DesignTokens) MCPServerListOption {
	return func(l *MCPServerList) {
		if tokens == nil {
			return
		}
		l.colors = mcpServerListColorsFromTokens(tokens)
	}
}

// MCPServerList renders MCP servers grouped by scope.
type MCPServerList struct {
	servers []MCPServerInfo
	cursor  int
	focused bool

	width       int
	windowWidth int
	colors      mcpServerListColors
}

type mcpServerListColors struct {
	header    string
	scope     string
	text      string
	muted     string
	connected string
	offline   string
	cursor    string
}

func defaultMCPServerListColors() mcpServerListColors {
	return mcpServerListColors{
		header:    style.ANSICyan,
		scope:     style.ANSIDim,
		text:      style.ANSIWhite,
		muted:     style.ANSIDim,
		connected: style.ANSIGreen,
		offline:   style.ANSIRed,
		cursor:    style.ANSIInverse,
	}
}

func mcpServerListColorsFromTokens(tokens *design.DesignTokens) mcpServerListColors {
	c := defaultMCPServerListColors()
	if tokens == nil {
		return c
	}
	if v := style.Fg(tokens.Accent); v != "" {
		c.header = v
	}
	if v := style.Fg(tokens.MutedColor); v != "" {
		c.scope = v
		c.muted = v
	}
	if v := style.Fg(tokens.Color); v != "" {
		c.text = v
	}
	if v := style.Fg(tokens.SuccessBright); v != "" {
		c.connected = v
	}
	if v := style.Fg(tokens.ErrorBright); v != "" {
		c.offline = v
	}
	if v := style.Bg(tokens.SurfaceRaised); v != "" {
		c.cursor = v
	}
	return c
}

// NewMCPServerList creates an MCP server list component.
func NewMCPServerList(servers []MCPServerInfo, opts ...MCPServerListOption) *MCPServerList {
	l := &MCPServerList{
		servers: append([]MCPServerInfo(nil), servers...),
		cursor:  0,
		focused: false,
		width:   0,
		colors:  defaultMCPServerListColors(),
	}
	for _, opt := range opts {
		opt(l)
	}
	l.clampCursor()
	return l
}

// Init initializes the component.
func (l *MCPServerList) Init() tea.Cmd {
	return nil
}

// Update handles keyboard navigation and selection.
func (l *MCPServerList) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.windowWidth = msg.Width
		return l, nil
	case tea.KeyMsg:
		if !l.focused {
			return l, nil
		}
		ordered := l.orderedServerIndices()
		if len(ordered) == 0 {
			return l, nil
		}
		switch msg.String() {
		case "up", "k":
			if l.cursor > 0 {
				l.cursor--
			}
		case "down", "j":
			if l.cursor < len(ordered)-1 {
				l.cursor++
			}
		case "enter":
			idx := ordered[l.cursor]
			selection := MCPServerSelectedMsg{Server: l.servers[idx], Index: idx}
			return l, func() tea.Msg {
				return selection
			}
		}
	}
	return l, nil
}

// View renders servers grouped by scope with selection cursor.
func (l *MCPServerList) View() string {
	width := l.renderWidth()
	lines := []string{fitMCPServerListLine(l.colors.header+style.ANSIBold+"MCP Servers"+style.ANSIReset, width)}
	if len(l.servers) == 0 {
		lines = append(lines, fitMCPServerListLine(l.colors.muted+"(no servers configured)"+style.ANSIReset, width))
		return strings.Join(lines, "\n") + "\n"
	}

	ordered := l.orderedServerIndices()
	selectedIndex := -1
	if len(ordered) > 0 && l.cursor >= 0 && l.cursor < len(ordered) {
		selectedIndex = ordered[l.cursor]
	}

	for _, scope := range orderedScopes() {
		indexes := l.scopeIndexes(scope)
		if len(indexes) == 0 {
			continue
		}
		scopeHeader := l.colors.scope + strings.ToUpper(string(scope)) + style.ANSIReset
		lines = append(lines, fitMCPServerListLine(scopeHeader, width))
		for _, idx := range indexes {
			s := l.servers[idx]
			status := l.colors.offline + "offline" + style.ANSIReset
			if s.Connected {
				status = l.colors.connected + "connected" + style.ANSIReset
			}
			name := strings.TrimSpace(s.Name)
			if name == "" {
				name = "(unnamed)"
			}
			url := style.ElideURL(strings.TrimSpace(s.URL), maxMCPServerURLWidth(width))
			line := fmt.Sprintf("  %s (%d tools) · %s · %s", name, s.ToolCount, status, url)
			if idx == selectedIndex {
				line = "❯ " + line[2:]
				if l.focused {
					line = l.colors.cursor + line + style.ANSIReset
				}
			}
			lines = append(lines, fitMCPServerListLine(line, width))
		}
	}

	return strings.Join(lines, "\n") + "\n"
}

// Focus marks this component focused.
func (l *MCPServerList) Focus() {
	l.focused = true
}

// Blur marks this component unfocused.
func (l *MCPServerList) Blur() {
	l.focused = false
}

// Focused reports focus state.
func (l *MCPServerList) Focused() bool {
	return l.focused
}

func (l *MCPServerList) orderedServerIndices() []int {
	result := make([]int, 0, len(l.servers))
	for _, scope := range orderedScopes() {
		for i, server := range l.servers {
			if server.Scope == scope {
				result = append(result, i)
			}
		}
	}
	return result
}

func (l *MCPServerList) scopeIndexes(scope MCPServerScope) []int {
	idx := []int{}
	for i, server := range l.servers {
		if server.Scope == scope {
			idx = append(idx, i)
		}
	}
	return idx
}

func orderedScopes() []MCPServerScope {
	return []MCPServerScope{ScopeProject, ScopeUser, ScopeLocal, ScopeEnterprise, ScopeDynamic}
}

func (l *MCPServerList) clampCursor() {
	ordered := l.orderedServerIndices()
	if len(ordered) == 0 {
		l.cursor = 0
		return
	}
	if l.cursor < 0 {
		l.cursor = 0
	}
	if l.cursor >= len(ordered) {
		l.cursor = len(ordered) - 1
	}
}

func (l *MCPServerList) renderWidth() int {
	if l.width > 0 {
		return l.width
	}
	if l.windowWidth > 0 {
		return l.windowWidth
	}
	return 0
}

func fitMCPServerListLine(s string, width int) string {
	if width <= 0 {
		return s
	}
	t := style.Truncate(s, width, "…")
	if style.StringWidth(t) < width {
		return style.Pad(t, width)
	}
	return t
}

func maxMCPServerURLWidth(width int) int {
	if width <= 0 {
		return 64
	}
	w := width / 3
	if w < 24 {
		w = 24
	}
	return w
}

var _ tui.Component = (*MCPServerList)(nil)
