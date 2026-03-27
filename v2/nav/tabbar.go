package nav

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// Tab describes a single tab in a tab bar.
type Tab struct {
	ID       string
	Label    string
	Badge    string
	Closable bool
}

// TabBar displays tabs and supports keyboard navigation.
type TabBar struct {
	tabs         []Tab
	activeIdx    int
	focused      bool
	onSelect     func(id string) tea.Cmd
	onClose      func(id string) tea.Cmd
	designTokens *design.DesignTokens
	scrollOffset int
	width        int
}

// TabBarOption configures a TabBar.
type TabBarOption func(*TabBar)

// WithTabBarDesignTokens applies design tokens to the tab bar.
func WithTabBarDesignTokens(tokens *design.DesignTokens) TabBarOption {
	return func(t *TabBar) {
		if tokens != nil {
			t.designTokens = tokens
		}
	}
}

// WithTabBarTheme applies a named design-system theme.
func WithTabBarTheme(theme string) TabBarOption {
	return func(t *TabBar) {
		t.designTokens = style.DesignTokensForTheme(theme)
	}
}

// WithTabs sets the initial tabs.
func WithTabs(tabs ...Tab) TabBarOption {
	return func(t *TabBar) {
		t.tabs = append([]Tab(nil), tabs...)
		if len(t.tabs) == 0 {
			t.activeIdx = -1
		} else if t.activeIdx < 0 || t.activeIdx >= len(t.tabs) {
			t.activeIdx = 0
		}
	}
}

// WithOnSelect sets the selection callback.
func WithOnSelect(fn func(string) tea.Cmd) TabBarOption {
	return func(t *TabBar) {
		t.onSelect = fn
	}
}

// WithOnClose sets the close callback.
func WithOnClose(fn func(string) tea.Cmd) TabBarOption {
	return func(t *TabBar) {
		t.onClose = fn
	}
}

// NewTabBar creates a new tab bar.
func NewTabBar(opts ...TabBarOption) *TabBar {
	t := &TabBar{
		tabs:         []Tab{},
		activeIdx:    -1,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(t)
	}

	if len(t.tabs) > 0 && t.activeIdx == -1 {
		t.activeIdx = 0
	}

	return t
}

// Init initializes the tab bar.
func (t *TabBar) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (t *TabBar) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		return t, nil
	case tea.KeyMsg:
		if !t.focused || len(t.tabs) == 0 {
			return t, nil
		}

		switch msg.String() {
		case "tab", "l", "right":
			return t, t.selectNext()
		case "shift+tab", "h", "left":
			return t, t.selectPrev()
		case "w":
			return t, t.closeActiveTab()
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.String()[0] - '1')
			if idx >= 0 && idx < len(t.tabs) {
				t.activeIdx = idx
				t.ensureActiveVisible()
				if t.onSelect != nil {
					return t, t.onSelect(t.tabs[t.activeIdx].ID)
				}
			}
		}
	}

	return t, nil
}

// View renders the tab bar.
func (t *TabBar) View() string {
	if len(t.tabs) == 0 {
		return ""
	}

	accent := style.ANSIBold
	if t.designTokens != nil {
		if c := style.ANSIColorFromHex(t.designTokens.Accent); c != "" {
			accent = c + style.ANSIBold
		}
	}
	inactive := style.ANSIDim
	reset := style.ANSIReset

	separator := "  "
	plainParts := make([]string, 0, len(t.tabs))
	styledParts := make([]string, 0, len(t.tabs))

	for i, tab := range t.tabs {
		label := tab.Label
		if tab.Badge != "" {
			label += " (" + tab.Badge + ")"
		}

		if i == t.activeIdx {
			plain := "┌─ " + label + " ─┐"
			plainParts = append(plainParts, plain)
			styledParts = append(styledParts, accent+plain+reset)
		} else {
			plainParts = append(plainParts, label)
			styledParts = append(styledParts, inactive+label+reset)
		}
	}

	topPlain := strings.Join(plainParts, separator)
	topStyled := strings.Join(styledParts, separator)

	bottomWidth := style.StringWidth(topPlain)
	if t.width > bottomWidth {
		bottomWidth = t.width
	}
	if bottomWidth < 1 {
		bottomWidth = 1
	}

	bottom := []rune(strings.Repeat("─", bottomWidth))

	if t.activeIdx >= 0 && t.activeIdx < len(t.tabs) {
		start := 0
		for i := 0; i < t.activeIdx; i++ {
			start += style.StringWidth(plainParts[i]) + style.StringWidth(separator)
		}
		activeLen := style.StringWidth(plainParts[t.activeIdx])
		if start < len(bottom) {
			bottom[start] = '┘'
		}
		for i := start + 1; i < start+activeLen-1 && i < len(bottom); i++ {
			bottom[i] = ' '
		}
		if end := start + activeLen - 1; end >= 0 && end < len(bottom) {
			bottom[end] = '└'
		}
	}
	return topStyled + "\n" + string(bottom) + "\n"
}

// Focus is called when this component receives focus.
func (t *TabBar) Focus() {
	t.focused = true
}

// Blur is called when this component loses focus.
func (t *TabBar) Blur() {
	t.focused = false
}

// Focused returns whether this component is currently focused.
func (t *TabBar) Focused() bool {
	return t.focused
}

// AddTab appends a new tab.
func (t *TabBar) AddTab(tab Tab) {
	t.tabs = append(t.tabs, tab)
	if t.activeIdx == -1 {
		t.activeIdx = 0
	}
	t.ensureActiveVisible()
}

// RemoveTab removes a tab by ID.
func (t *TabBar) RemoveTab(id string) {
	if len(t.tabs) == 0 {
		return
	}

	removeIdx := -1
	for i := range t.tabs {
		if t.tabs[i].ID == id {
			removeIdx = i
			break
		}
	}
	if removeIdx == -1 {
		return
	}

	t.tabs = append(t.tabs[:removeIdx], t.tabs[removeIdx+1:]...)

	if len(t.tabs) == 0 {
		t.activeIdx = -1
		t.scrollOffset = 0
		return
	}

	switch {
	case removeIdx < t.activeIdx:
		t.activeIdx--
	case removeIdx == t.activeIdx:
		if t.activeIdx >= len(t.tabs) {
			t.activeIdx = len(t.tabs) - 1
		}
	}

	t.ensureActiveVisible()
}

// SetActive makes the tab with ID active.
func (t *TabBar) SetActive(id string) {
	for i := range t.tabs {
		if t.tabs[i].ID == id {
			t.activeIdx = i
			t.ensureActiveVisible()
			return
		}
	}
}

// GetActiveTab returns the currently active tab.
func (t *TabBar) GetActiveTab() *Tab {
	if t.activeIdx < 0 || t.activeIdx >= len(t.tabs) {
		return nil
	}
	return &t.tabs[t.activeIdx]
}

// GetActiveID returns the active tab ID.
func (t *TabBar) GetActiveID() string {
	active := t.GetActiveTab()
	if active == nil {
		return ""
	}
	return active.ID
}

// SetBadge sets badge text for a tab ID.
func (t *TabBar) SetBadge(id string, badge string) {
	for i := range t.tabs {
		if t.tabs[i].ID == id {
			t.tabs[i].Badge = badge
			return
		}
	}
}

// TabCount returns number of tabs.
func (t *TabBar) TabCount() int {
	return len(t.tabs)
}

func (t *TabBar) selectNext() tea.Cmd {
	if len(t.tabs) == 0 {
		return nil
	}
	t.activeIdx = (t.activeIdx + 1) % len(t.tabs)
	t.ensureActiveVisible()
	if t.onSelect != nil {
		return t.onSelect(t.tabs[t.activeIdx].ID)
	}
	return nil
}

func (t *TabBar) selectPrev() tea.Cmd {
	if len(t.tabs) == 0 {
		return nil
	}
	t.activeIdx--
	if t.activeIdx < 0 {
		t.activeIdx = len(t.tabs) - 1
	}
	t.ensureActiveVisible()
	if t.onSelect != nil {
		return t.onSelect(t.tabs[t.activeIdx].ID)
	}
	return nil
}

func (t *TabBar) closeActiveTab() tea.Cmd {
	active := t.GetActiveTab()
	if active == nil || !active.Closable {
		return nil
	}

	closedID := active.ID
	t.RemoveTab(closedID)

	if t.onClose != nil {
		return t.onClose(closedID)
	}
	return nil
}

func (t *TabBar) ensureActiveVisible() {
	if t.activeIdx < 0 {
		t.scrollOffset = 0
		return
	}
	if t.activeIdx < t.scrollOffset {
		t.scrollOffset = t.activeIdx
	}
}
