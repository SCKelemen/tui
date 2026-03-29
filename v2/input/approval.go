package input

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ApprovalAction describes the requested tool action for user approval.
type ApprovalAction struct {
	Tool        string
	Description string
	Risk        string
	Details     string
}

// ApprovalResultMsg is emitted when the user confirms an approval choice.
type ApprovalResultMsg struct {
	Approved bool
	Action   ApprovalAction
	Always   bool
}

// ApprovalOption configures an Approval prompt.
type ApprovalOption func(*Approval)

// WithApprovalWidth sets the preferred render width.
func WithApprovalWidth(w int) ApprovalOption {
	return func(a *Approval) {
		if w > 0 {
			a.width = w
		}
	}
}

// Approval is a permission prompt for approving or denying tool actions.
type Approval struct {
	action   ApprovalAction
	selected int
	width    int
}

// NewApproval creates a new permission prompt.
func NewApproval(action ApprovalAction, opts ...ApprovalOption) *Approval {
	a := &Approval{
		action:   action,
		selected: 0,
		width:    78,
	}

	for _, opt := range opts {
		opt(a)
	}

	if a.selected < 0 || a.selected > 2 {
		a.selected = 0
	}

	return a
}

// Init initializes the approval prompt.
func (a *Approval) Init() tea.Cmd {
	return nil
}

// Update handles keyboard navigation and emits an approval result on confirmation.
func (a *Approval) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if a.width <= 0 {
			a.width = clampApprovalWidth(msg.Width - 8)
		}
		return a, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "shift+tab":
			a.selected = (a.selected + 2) % 3
			return a, nil
		case "right", "l", "tab":
			a.selected = (a.selected + 1) % 3
			return a, nil
		case "y":
			a.selected = 0
			return a, a.emitSelection()
		case "n":
			a.selected = 1
			return a, a.emitSelection()
		case "a":
			a.selected = 2
			return a, a.emitSelection()
		case "enter":
			return a, a.emitSelection()
		}
	}

	return a, nil
}

// View renders the approval prompt with action details and choice buttons.
func (a *Approval) View() string {
	width := a.width
	if width <= 0 {
		width = 78
	}

	boxStyle := lipgloss.NewStyle().
		Width(width).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	tool := strings.TrimSpace(a.action.Tool)
	if tool == "" {
		tool = "(unknown tool)"
	}

	description := strings.TrimSpace(a.action.Description)
	if description == "" {
		description = "No description provided"
	}

	detailsText := strings.TrimSpace(a.action.Details)
	if detailsText == "" {
		detailsText = "No additional details"
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Render("Permission required")
	toolLabel := lipgloss.NewStyle().Bold(true).Render("Tool:") + " " + lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true).Render(tool)
	descLabel := lipgloss.NewStyle().Bold(true).Render("Description:") + " " + description
	riskLabel := lipgloss.NewStyle().Bold(true).Render("Risk:") + " " + a.riskBadge()
	detailsHeader := lipgloss.NewStyle().Bold(true).Render("Details:")

	innerWidth := width - 6
	if innerWidth < 20 {
		innerWidth = 20
	}
	details := lipgloss.NewStyle().Width(innerWidth).Render(detailsText)

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Left,
		a.renderButton(0, "✓ Approve", "34", "22"),
		" ",
		a.renderButton(1, "✗ Deny", "160", "52"),
		" ",
		a.renderButton(2, "⟳ Always Approve", "63", "24"),
	)

	hint := lipgloss.NewStyle().Faint(true).Render("←/→/Tab: move · Enter: confirm · y: approve · n: deny · a: always")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		toolLabel,
		descLabel,
		riskLabel,
		"",
		detailsHeader,
		details,
		"",
		buttons,
		hint,
	)

	return boxStyle.Render(content)
}

func (a *Approval) emitSelection() tea.Cmd {
	result := ApprovalResultMsg{
		Approved: a.selected != 1,
		Action:   a.action,
		Always:   a.selected == 2,
	}

	return func() tea.Msg {
		return result
	}
}

func (a *Approval) riskBadge() string {
	risk := strings.ToLower(strings.TrimSpace(a.action.Risk))
	label := strings.ToUpper(risk)
	if label == "" {
		label = "UNKNOWN"
	}

	color := lipgloss.Color("245")
	switch risk {
	case "low":
		color = lipgloss.Color("34")
	case "medium":
		color = lipgloss.Color("220")
	case "high":
		color = lipgloss.Color("160")
	}

	return lipgloss.NewStyle().
		Bold(true).
		Foreground(color).
		Render(label)
}

func (a *Approval) renderButton(index int, label, fg, bg string) string {
	buttonStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color("250")).
		Background(lipgloss.Color("238"))

	if a.selected == index {
		buttonStyle = buttonStyle.
			Bold(true).
			Foreground(lipgloss.Color(fg)).
			Background(lipgloss.Color(bg))
	}

	return buttonStyle.Render("[" + label + "]")
}

func clampApprovalWidth(w int) int {
	if w < 56 {
		return 56
	}
	if w > 100 {
		return 100
	}
	return w
}
