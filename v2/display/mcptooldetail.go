package display

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// MCPToolCapability describes a tool safety profile.
type MCPToolCapability int

const (
	CapabilityReadOnly MCPToolCapability = iota
	CapabilityDestructive
	CapabilityOpenWorld
)

// MCPToolInfo contains metadata for one MCP tool.
type MCPToolInfo struct {
	Name         string
	Description  string
	Server       string
	Capabilities []MCPToolCapability
	InputSchema  string
}

// MCPToolDetailOption configures an MCPToolDetail component.
type MCPToolDetailOption func(*MCPToolDetail)

// WithMCPToolDetailWidth sets a fixed render width.
func WithMCPToolDetailWidth(width int) MCPToolDetailOption {
	return func(v *MCPToolDetail) {
		if width > 0 {
			v.width = width
		}
	}
}

// WithMCPToolDetailDesignTokens applies design-system colors.
func WithMCPToolDetailDesignTokens(tokens *design.DesignTokens) MCPToolDetailOption {
	return func(v *MCPToolDetail) {
		if tokens == nil {
			return
		}
		v.colors = mcpToolDetailColorsFromTokens(tokens)
	}
}

// MCPToolDetail renders metadata and schema for a selected MCP tool.
type MCPToolDetail struct {
	tool MCPToolInfo

	width       int
	windowWidth int
	focused     bool
	colors      mcpToolDetailColors
}

type mcpToolDetailColors struct {
	header      string
	text        string
	muted       string
	readOnly    string
	destructive string
	openWorld   string
	codeBg      string
	codeFg      string
}

func defaultMCPToolDetailColors() mcpToolDetailColors {
	return mcpToolDetailColors{
		header:      style.ANSICyan,
		text:        style.ANSIWhite,
		muted:       style.ANSIDim,
		readOnly:    style.ANSIGreen,
		destructive: style.ANSIRed,
		openWorld:   style.ANSIYellow,
		codeBg:      "\033[48;5;236m",
		codeFg:      style.ANSIWhite,
	}
}

func mcpToolDetailColorsFromTokens(tokens *design.DesignTokens) mcpToolDetailColors {
	c := defaultMCPToolDetailColors()
	if tokens == nil {
		return c
	}
	if v := style.Fg(tokens.Accent); v != "" {
		c.header = v
	}
	if v := style.Fg(tokens.Color); v != "" {
		c.text = v
		c.codeFg = v
	}
	if v := style.Fg(tokens.MutedColor); v != "" {
		c.muted = v
	}
	if v := style.Fg(tokens.SuccessBright); v != "" {
		c.readOnly = v
	}
	if v := style.Fg(tokens.ErrorBright); v != "" {
		c.destructive = v
	}
	if v := style.Fg(tokens.PendingColor); v != "" {
		c.openWorld = v
	}
	if v := style.Bg(tokens.SurfaceRaised); v != "" {
		c.codeBg = v
	}
	return c
}

// NewMCPToolDetail creates a new MCPToolDetail.
func NewMCPToolDetail(tool MCPToolInfo, opts ...MCPToolDetailOption) *MCPToolDetail {
	v := &MCPToolDetail{
		tool:    tool,
		width:   0,
		focused: false,
		colors:  defaultMCPToolDetailColors(),
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// Init initializes the component.
func (v *MCPToolDetail) Init() tea.Cmd {
	return nil
}

// Update handles window-size updates.
func (v *MCPToolDetail) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.windowWidth = msg.Width
	}
	return v, nil
}

// View renders name, capability badges, description, and input schema.
func (v *MCPToolDetail) View() string {
	width := v.renderWidth()
	if width < 50 {
		width = 50
	}
	lines := make([]string, 0, 32)

	name := strings.TrimSpace(v.tool.Name)
	if name == "" {
		name = "(unnamed tool)"
	}
	header := v.colors.header + style.ANSIBold + name + style.ANSIReset
	if server := strings.TrimSpace(v.tool.Server); server != "" {
		header += v.colors.muted + "  @" + server + style.ANSIReset
	}
	lines = append(lines, fitMCPToolDetailLine(header, width))

	lines = append(lines, fitMCPToolDetailLine(v.renderCapabilities(), width))

	desc := strings.TrimSpace(v.tool.Description)
	if desc == "" {
		desc = "No description available"
	}
	for _, row := range wrapMCPToolDetail(desc, width) {
		lines = append(lines, fitMCPToolDetailLine(v.colors.text+row+style.ANSIReset, width))
	}

	lines = append(lines, fitMCPToolDetailLine(v.colors.muted+"input schema"+style.ANSIReset, width))
	schema := strings.TrimSpace(v.tool.InputSchema)
	if schema == "" {
		schema = "{}"
	}
	for _, row := range strings.Split(schema, "\n") {
		content := " " + style.Pad(style.Truncate(row, width-2, "…"), width-2) + " "
		lines = append(lines, v.colors.codeBg+v.colors.codeFg+content+style.ANSIReset)
	}

	return strings.Join(lines, "\n") + "\n"
}

// Focus marks this component focused.
func (v *MCPToolDetail) Focus() {
	v.focused = true
}

// Blur marks this component unfocused.
func (v *MCPToolDetail) Blur() {
	v.focused = false
}

// Focused reports focus state.
func (v *MCPToolDetail) Focused() bool {
	return v.focused
}

func (v *MCPToolDetail) renderWidth() int {
	if v.width > 0 {
		return v.width
	}
	if v.windowWidth > 0 {
		return v.windowWidth
	}
	return 0
}

func (v *MCPToolDetail) renderCapabilities() string {
	if len(v.tool.Capabilities) == 0 {
		return v.colors.muted + "[NO CAPABILITIES]" + style.ANSIReset
	}
	parts := make([]string, 0, len(v.tool.Capabilities))
	for _, cap := range v.tool.Capabilities {
		label := "READ_ONLY"
		color := v.colors.readOnly
		switch cap {
		case CapabilityDestructive:
			label = "DESTRUCTIVE"
			color = v.colors.destructive
		case CapabilityOpenWorld:
			label = "OPEN_WORLD"
			color = v.colors.openWorld
		}
		parts = append(parts, color+style.ANSIBold+"["+label+"]"+style.ANSIReset)
	}
	return strings.Join(parts, " ")
}

func fitMCPToolDetailLine(s string, width int) string {
	if width <= 0 {
		return s
	}
	t := style.Truncate(s, width, "…")
	if style.StringWidth(t) < width {
		return style.Pad(t, width)
	}
	return t
}

func wrapMCPToolDetail(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	rows := make([]string, 0, len(words))
	line := words[0]
	for _, word := range words[1:] {
		candidate := line + " " + word
		if style.StringWidth(candidate) <= width {
			line = candidate
			continue
		}
		rows = append(rows, line)
		line = word
	}
	rows = append(rows, line)
	return rows
}

var _ tui.Component = (*MCPToolDetail)(nil)
