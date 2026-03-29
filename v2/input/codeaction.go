package input

import (
	"fmt"
	"strings"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CodeActionKind classifies a code action similarly to LSP action kinds.
type CodeActionKind int

const (
	// KindQuickFix represents a fast issue-specific correction.
	KindQuickFix CodeActionKind = iota
	// KindRefactor represents a general code refactor action.
	KindRefactor
	// KindRefactorExtract represents extract refactors (method/variable/type).
	KindRefactorExtract
	// KindRefactorInline represents inline refactors.
	KindRefactorInline
	// KindRefactorRewrite represents rewrite-style refactors.
	KindRefactorRewrite
	// KindSource represents source-level file actions.
	KindSource
	// KindSourceOrganizeImports represents import organization actions.
	KindSourceOrganizeImports
	// KindSourceFixAll represents bulk source fixes.
	KindSourceFixAll
)

// KindLabel returns a human-readable label for an action kind.
func KindLabel(kind CodeActionKind) string {
	switch kind {
	case KindQuickFix:
		return "Quick Fix"
	case KindRefactor:
		return "Refactor"
	case KindRefactorExtract:
		return "Refactor: Extract"
	case KindRefactorInline:
		return "Refactor: Inline"
	case KindRefactorRewrite:
		return "Refactor: Rewrite"
	case KindSource:
		return "Source"
	case KindSourceOrganizeImports:
		return "Source: Organize Imports"
	case KindSourceFixAll:
		return "Source: Fix All"
	default:
		return "Action"
	}
}

// CodeActionKindIcon returns an icon for an action kind.
func CodeActionKindIcon(kind CodeActionKind) string {
	switch kind {
	case KindQuickFix:
		return "💡"
	case KindRefactor, KindRefactorExtract, KindRefactorInline, KindRefactorRewrite:
		return "🔧"
	case KindSourceOrganizeImports:
		return "📦"
	case KindSourceFixAll:
		return "🩹"
	case KindSource:
		return "🧰"
	default:
		return "•"
	}
}

// CodeAction is one executable action shown in the quick-fix menu.
type CodeAction struct {
	Title       string
	Kind        CodeActionKind
	IsPreferred bool
	Description string
	Diagnostics []string
}

// CodeActionSelectedMsg is emitted when a code action is selected.
type CodeActionSelectedMsg struct {
	Action CodeAction
}

// CodeActionMenuOption configures a CodeActionMenu.
type CodeActionMenuOption func(*CodeActionMenu)

// WithCodeActionWidth sets a fixed total menu width in terminal cells.
func WithCodeActionWidth(width int) CodeActionMenuOption {
	return func(m *CodeActionMenu) {
		if width > 0 {
			m.width = width
		}
	}
}

// WithCodeActionAnchor sets the line number where the menu should attach.
func WithCodeActionAnchor(line int) CodeActionMenuOption {
	return func(m *CodeActionMenu) {
		if line >= 0 {
			m.anchorLine = line
		}
	}
}

// CodeActionMenu is a VS Code-style quick action menu.
type CodeActionMenu struct {
	actions    []CodeAction
	cursor     int
	visible    bool
	width      int
	anchorLine int

	focused     bool
	windowWidth int

	frameStyle     lipgloss.Style
	headerStyle    lipgloss.Style
	groupStyle     lipgloss.Style
	itemStyle      lipgloss.Style
	selectedStyle  lipgloss.Style
	preferredStyle lipgloss.Style
	detailStyle    lipgloss.Style
}

// NewCodeActionMenu creates a new code action menu.
func NewCodeActionMenu(actions []CodeAction, opts ...CodeActionMenuOption) *CodeActionMenu {
	menu := &CodeActionMenu{
		actions:     append([]CodeAction(nil), actions...),
		cursor:      0,
		visible:     false,
		width:       48,
		anchorLine:  0,
		focused:     false,
		windowWidth: 0,
		frameStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4B5563")).
			Background(lipgloss.Color("#111827")).
			Padding(0, 1),
		headerStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#E5E7EB")),
		groupStyle: lipgloss.NewStyle().
			Faint(true).
			Foreground(lipgloss.Color("#9CA3AF")),
		itemStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1D5DB")),
		selectedStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#2563EB")),
		preferredStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FBBF24")).
			Bold(true),
		detailStyle: lipgloss.NewStyle().
			Faint(true).
			Foreground(lipgloss.Color("#9CA3AF")),
	}

	for _, opt := range opts {
		opt(menu)
	}

	menu.clampCursor()
	return menu
}

// Init initializes the menu.
func (m *CodeActionMenu) Init() tea.Cmd {
	return nil
}

// Update handles navigation and selection.
func (m *CodeActionMenu) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		return m, nil
	case tea.KeyMsg:
		if !m.visible || !m.focused || len(m.actions) == 0 {
			if m.visible && msg.Type == tea.KeyEsc {
				m.Hide()
			}
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.actions)-1 {
				m.cursor++
			}
			return m, nil
		case "enter":
			selected, ok := m.selectedAction()
			if !ok {
				return m, nil
			}
			m.Hide()
			return m, func() tea.Msg {
				return CodeActionSelectedMsg{Action: selected}
			}
		case "esc":
			m.Hide()
			return m, nil
		}
	}

	return m, nil
}

// View renders the floating menu.
func (m *CodeActionMenu) View() string {
	if !m.visible {
		return ""
	}

	contentWidth := m.contentWidth()
	header := "💡 Quick Actions"
	if m.refactorOnly() {
		header = "🔧 Refactoring"
	}

	lines := []string{style.Pad(style.Truncate(m.headerStyle.Render(header), contentWidth, "…"), contentWidth)}

	ordered := m.orderedActionIndices()
	if len(ordered) == 0 {
		noActions := m.groupStyle.Render("No actions available")
		lines = append(lines, style.Pad(style.Truncate(noActions, contentWidth, "…"), contentWidth))
	} else {
		lastKind := CodeActionKind(-1)
		for rowIndex, actionIndex := range ordered {
			action := m.actions[actionIndex]

			if action.Kind != lastKind {
				group := m.groupStyle.Render(strings.ToUpper(KindLabel(action.Kind)))
				lines = append(lines, style.Pad(style.Truncate(group, contentWidth, "…"), contentWidth))
				lastKind = action.Kind
			}

			line := fmt.Sprintf("%s %s", CodeActionKindIcon(action.Kind), action.Title)
			if action.IsPreferred {
				line += " " + m.preferredStyle.Render("★")
			}
			if strings.TrimSpace(action.Description) != "" {
				line += " " + m.detailStyle.Render("— "+strings.TrimSpace(action.Description))
			}
			if len(action.Diagnostics) > 0 {
				line += " " + m.detailStyle.Render(fmt.Sprintf("(%d diagnostics)", len(action.Diagnostics)))
			}

			line = style.Pad(style.Truncate(line, contentWidth, "…"), contentWidth)
			if rowIndex == m.cursor {
				line = m.selectedStyle.Render(line)
			} else {
				line = m.itemStyle.Render(line)
			}
			lines = append(lines, line)
		}
	}

	rendered := m.frameStyle.Width(m.renderWidth()).Render(strings.Join(lines, "\n"))
	if m.anchorLine <= 0 {
		return rendered
	}

	return strings.Repeat("\n", m.anchorLine) + rendered
}

// Focus marks the menu as focused.
func (m *CodeActionMenu) Focus() {
	m.focused = true
}

// Blur marks the menu as unfocused.
func (m *CodeActionMenu) Blur() {
	m.focused = false
}

// Focused reports whether the menu has focus.
func (m *CodeActionMenu) Focused() bool {
	return m.focused
}

// Show makes the menu visible.
func (m *CodeActionMenu) Show() {
	m.visible = true
	m.clampCursor()
}

// Hide conceals the menu.
func (m *CodeActionMenu) Hide() {
	m.visible = false
}

// Toggle flips menu visibility.
func (m *CodeActionMenu) Toggle() {
	if m.visible {
		m.Hide()
		return
	}
	m.Show()
}

// Visible reports whether the menu is visible.
func (m *CodeActionMenu) Visible() bool {
	return m.visible
}

func (m *CodeActionMenu) selectedAction() (CodeAction, bool) {
	ordered := m.orderedActionIndices()
	if len(ordered) == 0 || m.cursor < 0 || m.cursor >= len(ordered) {
		return CodeAction{}, false
	}
	return m.actions[ordered[m.cursor]], true
}

func (m *CodeActionMenu) orderedActionIndices() []int {
	if len(m.actions) == 0 {
		return nil
	}

	kindOrder := []CodeActionKind{
		KindQuickFix,
		KindRefactor,
		KindRefactorExtract,
		KindRefactorInline,
		KindRefactorRewrite,
		KindSource,
		KindSourceOrganizeImports,
		KindSourceFixAll,
	}

	seenKinds := make(map[CodeActionKind]bool, len(kindOrder))
	ordered := make([]int, 0, len(m.actions))

	for _, kind := range kindOrder {
		for i, action := range m.actions {
			if action.Kind == kind {
				ordered = append(ordered, i)
				seenKinds[kind] = true
			}
		}
	}

	// Keep unknown kinds at the end while preserving original order.
	for i, action := range m.actions {
		if !seenKinds[action.Kind] {
			ordered = append(ordered, i)
		}
	}

	return ordered
}

func (m *CodeActionMenu) refactorOnly() bool {
	if len(m.actions) == 0 {
		return false
	}
	for _, action := range m.actions {
		switch action.Kind {
		case KindRefactor, KindRefactorExtract, KindRefactorInline, KindRefactorRewrite:
			continue
		default:
			return false
		}
	}
	return true
}

func (m *CodeActionMenu) clampCursor() {
	if len(m.actions) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.actions) {
		m.cursor = len(m.actions) - 1
	}
}

func (m *CodeActionMenu) renderWidth() int {
	if m.width > 0 {
		return m.width
	}
	if m.windowWidth <= 0 {
		return 48
	}
	w := m.windowWidth - 4
	if w < 24 {
		w = 24
	}
	if w > 72 {
		w = 72
	}
	return w
}

func (m *CodeActionMenu) contentWidth() int {
	w := m.renderWidth() - 4
	if w < 1 {
		return 1
	}
	return w
}

var _ tui.Component = (*CodeActionMenu)(nil)
