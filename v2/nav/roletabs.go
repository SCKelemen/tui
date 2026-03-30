package nav

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// RoleTab represents one role-aware tab item.
type RoleTab struct {
	Label  string
	Role   string
	Active bool
}

// RolePalette is a 5-step color intensity palette for a role.
type RolePalette []string

// RoleTabSelectedMsg is emitted when Enter is pressed on the selected role tab.
type RoleTabSelectedMsg struct {
	Index int
	Role  string
}

// RoleTabsOption configures a RoleTabs component.
type RoleTabsOption func(*RoleTabs)

// RoleTabs renders horizontal role-based tabs.
type RoleTabs struct {
	tabs         []RoleTab
	selected     int
	focused      bool
	designTokens *design.DesignTokens
	palettes     map[string]RolePalette
}

var _ tui.Component = (*RoleTabs)(nil)

// WithRoleTabsDesignTokens applies design-system tokens.
func WithRoleTabsDesignTokens(tokens *design.DesignTokens) RoleTabsOption {
	return func(r *RoleTabs) {
		if tokens != nil {
			r.designTokens = tokens
		}
	}
}

// WithRoleTabsPalettes sets role-specific palettes.
func WithRoleTabsPalettes(palettes map[string]RolePalette) RoleTabsOption {
	return func(r *RoleTabs) {
		if len(palettes) == 0 {
			return
		}
		r.palettes = copyRolePalettes(palettes)
		r.ensureDefaultPalette()
	}
}

// WithRoleTabsSelected sets the initial selected tab index.
func WithRoleTabsSelected(index int) RoleTabsOption {
	return func(r *RoleTabs) {
		r.selected = index
	}
}

// NewRoleTabs creates a new RoleTabs component.
func NewRoleTabs(tabs []RoleTab, opts ...RoleTabsOption) *RoleTabs {
	r := &RoleTabs{
		tabs:         append([]RoleTab(nil), tabs...),
		selected:     0,
		designTokens: design.DefaultTheme(),
		palettes:     defaultRolePalettes(),
	}

	activeIdx := -1
	for i := range r.tabs {
		if r.tabs[i].Active {
			activeIdx = i
			break
		}
	}
	if activeIdx >= 0 {
		r.selected = activeIdx
	}

	for _, opt := range opts {
		opt(r)
	}

	r.ensureDefaultPalette()
	r.clampSelected()
	r.syncActive()

	return r
}

// Init initializes the component.
func (r *RoleTabs) Init() tea.Cmd {
	return nil
}

// Update handles keyboard interaction and selection.
func (r *RoleTabs) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		_ = msg
		return r, nil
	case tea.KeyMsg:
		if !r.focused || len(r.tabs) == 0 {
			return r, nil
		}

		switch msg.String() {
		case "h", "left":
			r.selectPrev()
			return r, nil
		case "l", "right":
			r.selectNext()
			return r, nil
		case "enter":
			selected := r.current()
			if selected == nil {
				return r, nil
			}
			return r, func() tea.Msg {
				return RoleTabSelectedMsg{Index: r.selected, Role: selected.Role}
			}
		}
	}

	return r, nil
}

// View renders the role tabs.
func (r *RoleTabs) View() string {
	if len(r.tabs) == 0 {
		return ""
	}

	parts := make([]string, 0, len(r.tabs))
	for i := range r.tabs {
		tab := r.tabs[i]
		palette := r.paletteForRole(tab.Role)

		label := strings.TrimSpace(tab.Label)
		if label == "" {
			label = tab.Role
		}
		if label == "" {
			label = "Tab"
		}

		text := "[ " + label + " ]"

		if i == r.selected {
			st := style.Fg(paletteColor(palette, 0)) + style.ANSIBold + style.ANSIUnderline
			parts = append(parts, st+text+style.ANSIReset)
			continue
		}

		st := style.Fg(paletteColor(palette, 2))
		parts = append(parts, st+text+style.ANSIReset)
	}

	return strings.Join(parts, "  ")
}

// Focus marks the component as focused.
func (r *RoleTabs) Focus() {
	r.focused = true
}

// Blur marks the component as blurred.
func (r *RoleTabs) Blur() {
	r.focused = false
}

// Focused reports whether the component currently has focus.
func (r *RoleTabs) Focused() bool {
	return r.focused
}

func (r *RoleTabs) selectNext() {
	if len(r.tabs) == 0 {
		return
	}
	r.selected = (r.selected + 1) % len(r.tabs)
	r.syncActive()
}

func (r *RoleTabs) selectPrev() {
	if len(r.tabs) == 0 {
		return
	}
	r.selected--
	if r.selected < 0 {
		r.selected = len(r.tabs) - 1
	}
	r.syncActive()
}

func (r *RoleTabs) current() *RoleTab {
	if r.selected < 0 || r.selected >= len(r.tabs) {
		return nil
	}
	return &r.tabs[r.selected]
}

func (r *RoleTabs) clampSelected() {
	if len(r.tabs) == 0 {
		r.selected = -1
		return
	}
	if r.selected < 0 {
		r.selected = 0
	}
	if r.selected >= len(r.tabs) {
		r.selected = len(r.tabs) - 1
	}
}

func (r *RoleTabs) syncActive() {
	for i := range r.tabs {
		r.tabs[i].Active = i == r.selected
	}
}

func (r *RoleTabs) paletteForRole(role string) RolePalette {
	if p, ok := r.palettes[strings.ToLower(strings.TrimSpace(role))]; ok && len(p) > 0 {
		return p
	}
	if p, ok := r.palettes["default"]; ok && len(p) > 0 {
		return p
	}
	return defaultRolePalettes()["default"]
}

func (r *RoleTabs) ensureDefaultPalette() {
	if r.palettes == nil {
		r.palettes = defaultRolePalettes()
		return
	}
	if _, ok := r.palettes["default"]; !ok {
		r.palettes["default"] = defaultRolePalettes()["default"]
	}
}

func defaultRolePalettes() map[string]RolePalette {
	return map[string]RolePalette{
		"coder": {
			"#3B82F6",
			"#60A5FA",
			"#93C5FD",
			"#BFDBFE",
			"#DBEAFE",
		},
		"planner": {
			"#8B5CF6",
			"#A78BFA",
			"#C4B5FD",
			"#DDD6FE",
			"#EDE9FE",
		},
		"reviewer": {
			"#10B981",
			"#34D399",
			"#6EE7B7",
			"#A7F3D0",
			"#D1FAE5",
		},
		"default": {
			"#6B7280",
			"#9CA3AF",
			"#D1D5DB",
			"#E5E7EB",
			"#F3F4F6",
		},
	}
}

func copyRolePalettes(src map[string]RolePalette) map[string]RolePalette {
	copied := make(map[string]RolePalette, len(src))
	for role, palette := range src {
		normRole := strings.ToLower(strings.TrimSpace(role))
		if normRole == "" {
			continue
		}
		copied[normRole] = append(RolePalette(nil), palette...)
	}
	return copied
}

func paletteColor(palette RolePalette, index int) string {
	if len(palette) == 0 {
		return ""
	}
	if index < 0 {
		index = 0
	}
	if index >= len(palette) {
		index = len(palette) - 1
	}
	return palette[index]
}
