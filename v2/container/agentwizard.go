package container

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// WizardStep represents one step in the agent setup flow.
type WizardStep struct {
	Title       string
	Description string
	Validate    func(value string) error
}

// AgentWizardCompleteMsg is emitted when all wizard steps are completed.
type AgentWizardCompleteMsg struct {
	Values map[int]string
}

// AgentWizardOption configures an AgentWizard.
type AgentWizardOption func(*AgentWizard)

// WithAgentWizardWidth sets the preferred render width.
func WithAgentWizardWidth(width int) AgentWizardOption {
	return func(a *AgentWizard) {
		if width > 0 {
			a.width = width
		}
	}
}

// WithAgentWizardDesignTokens applies design-system tokens.
func WithAgentWizardDesignTokens(tokens *design.DesignTokens) AgentWizardOption {
	return func(a *AgentWizard) {
		a.designTokens = tokens
		a.applyDesignTokens()
	}
}

// AgentWizard is a multi-step wizard for agent creation.
type AgentWizard struct {
	steps       []WizardStep
	currentStep int
	values      map[int]string
	focused     bool
	width       int

	focusIndex int // 0=input, 1=next, 2=back, 3=cancel
	input      textinput.Model
	errText    string

	designTokens *design.DesignTokens
	accentColor  string
	mutedColor   string
	errorColor   string
}

// NewAgentWizard creates a new AgentWizard.
func NewAgentWizard(steps []WizardStep, opts ...AgentWizardOption) *AgentWizard {
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 256

	a := &AgentWizard{
		steps:        append([]WizardStep(nil), steps...),
		currentStep:  0,
		values:       make(map[int]string),
		focused:      false,
		width:        72,
		focusIndex:   0,
		input:        ti,
		designTokens: design.DefaultTheme(),
		accentColor:  style.ANSICyan,
		mutedColor:   style.ANSIDim,
		errorColor:   style.ANSIRed,
	}

	for _, opt := range opts {
		opt(a)
	}

	a.applyDesignTokens()
	a.syncInputForStep()

	return a
}

// Init initializes the wizard.
func (a *AgentWizard) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles Bubble Tea messages.
func (a *AgentWizard) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if a.width <= 0 {
			a.width = msg.Width
		}
		if a.width < 20 {
			a.width = 20
		}
		a.input.Width = a.width - 2
		return a, nil

	case tea.KeyMsg:
		if !a.focused {
			return a, nil
		}

		switch msg.String() {
		case "tab":
			a.focusIndex = (a.focusIndex + 1) % 4
			return a, nil
		case "esc":
			if a.currentStep > 0 {
				a.currentStep--
				a.errText = ""
				a.syncInputForStep()
			}
			return a, nil
		case "enter":
			switch a.focusIndex {
			case 2:
				if a.currentStep > 0 {
					a.currentStep--
					a.errText = ""
					a.syncInputForStep()
				}
				return a, nil
			case 3:
				return a, tea.Quit
			default:
				return a.handleAdvance()
			}
		}

		if a.focusIndex == 0 {
			var cmd tea.Cmd
			a.input, cmd = a.input.Update(msg)
			return a, cmd
		}
	}

	return a, nil
}

// View renders the wizard.
func (a *AgentWizard) View() string {
	if len(a.steps) == 0 {
		return ""
	}

	step := a.steps[a.currentStep]
	w := a.width
	if w <= 0 {
		w = 72
	}

	indicator := fmt.Sprintf("%sStep %d/%d%s", a.mutedColor, a.currentStep+1, len(a.steps), style.ANSIReset)
	title := style.ANSIBold + strings.TrimSpace(step.Title) + style.ANSIReset
	if strings.TrimSpace(step.Title) == "" {
		title = style.ANSIBold + "Step" + style.ANSIReset
	}

	desc := strings.TrimSpace(step.Description)
	if desc == "" {
		desc = "Provide a value to continue."
	}

	inputLabel := "Value"
	if a.focusIndex == 0 {
		inputLabel = a.accentColor + style.ANSIBold + "Value" + style.ANSIReset
	}
	inputView := a.input.View()

	errLine := ""
	if strings.TrimSpace(a.errText) != "" {
		errLine = a.errorColor + "✗ " + a.errText + style.ANSIReset
	}

	footer := a.renderFooter()

	lines := []string{
		indicator,
		title,
		desc,
		"",
		inputLabel + ":",
		inputView,
	}
	if errLine != "" {
		lines = append(lines, errLine)
	}
	lines = append(lines, "", footer)

	for i := range lines {
		if w > 0 {
			lines[i] = style.Pad(style.Truncate(lines[i], w, "…"), w)
		}
	}

	return strings.Join(lines, "\n")
}

// Focus marks the component as focused.
func (a *AgentWizard) Focus() {
	a.focused = true
	a.input.Focus()
}

// Blur marks the component as unfocused.
func (a *AgentWizard) Blur() {
	a.focused = false
	a.input.Blur()
}

// Focused reports focus state.
func (a *AgentWizard) Focused() bool {
	return a.focused
}

func (a *AgentWizard) handleAdvance() (tui.Component, tea.Cmd) {
	if len(a.steps) == 0 {
		return a, nil
	}

	value := a.input.Value()
	step := a.steps[a.currentStep]
	if step.Validate != nil {
		if err := step.Validate(value); err != nil {
			a.errText = err.Error()
			return a, nil
		}
	}

	a.values[a.currentStep] = value
	a.errText = ""

	if a.currentStep >= len(a.steps)-1 {
		out := make(map[int]string, len(a.values))
		for k, v := range a.values {
			out[k] = v
		}
		return a, func() tea.Msg {
			return AgentWizardCompleteMsg{Values: out}
		}
	}

	a.currentStep++
	a.syncInputForStep()
	return a, nil
}

func (a *AgentWizard) syncInputForStep() {
	if a.currentStep < 0 {
		a.currentStep = 0
	}
	if a.currentStep >= len(a.steps) && len(a.steps) > 0 {
		a.currentStep = len(a.steps) - 1
	}

	if v, ok := a.values[a.currentStep]; ok {
		a.input.SetValue(v)
	} else {
		a.input.SetValue("")
	}
	a.input.CursorEnd()
}

func (a *AgentWizard) renderFooter() string {
	next := "[Next]"
	back := "[Back]"
	cancel := "[Cancel]"

	highlight := func(label string) string {
		return style.ANSIInverse + label + style.ANSIReset
	}

	switch a.focusIndex {
	case 1:
		next = highlight(next)
	case 2:
		back = highlight(back)
	case 3:
		cancel = highlight(cancel)
	}

	return fmt.Sprintf("%s  %s  %s", next, back, cancel)
}

func (a *AgentWizard) applyDesignTokens() {
	if a.designTokens == nil {
		return
	}
	if v := style.Fg(a.designTokens.Accent); v != "" {
		a.accentColor = v
	}
	if v := style.Fg(a.designTokens.MutedColor); v != "" {
		a.mutedColor = v
	}
	if v := style.Fg(a.designTokens.ErrorBright); v != "" {
		a.errorColor = v
	}
}

var _ tui.Component = (*AgentWizard)(nil)