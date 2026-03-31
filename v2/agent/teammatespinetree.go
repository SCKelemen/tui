package agent

import (
	"fmt"
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TeammateInfo represents one teammate (and optional children) in the spinner tree.
type TeammateInfo struct {
	Name     string
	Status   string
	Message  string
	Children []TeammateInfo
}

// TeammateSpinnerTree renders a recursive tree of teammate spinners.
type TeammateSpinnerTree struct {
	teammates []TeammateInfo
	width     int
	focused   bool
	spinIdx   int

	designTokens *design.DesignTokens
	spinner      Spinner

	runningColor string
	doneColor    string
	errorColor   string
	statusColor  string
	nameColor    string
	mutedColor   string
}

// TeammateSpinnerTreeOption configures TeammateSpinnerTree.
type TeammateSpinnerTreeOption func(*TeammateSpinnerTree)

// WithTeammateSpinnerTreeWidth sets the rendered width.
func WithTeammateSpinnerTreeWidth(width int) TeammateSpinnerTreeOption {
	return func(t *TeammateSpinnerTree) {
		if width >= 0 {
			t.width = width
		}
	}
}

// WithTeammateSpinnerTreeDesignTokens applies design-system tokens.
func WithTeammateSpinnerTreeDesignTokens(tokens *design.DesignTokens) TeammateSpinnerTreeOption {
	return func(t *TeammateSpinnerTree) {
		if tokens == nil {
			return
		}

		t.designTokens = tokens
		t.applyDesignTokens(tokens)
	}
}

// NewTeammateSpinnerTree creates a TeammateSpinnerTree.
func NewTeammateSpinnerTree(teammates []TeammateInfo, opts ...TeammateSpinnerTreeOption) *TeammateSpinnerTree {
	t := &TeammateSpinnerTree{
		teammates:    append([]TeammateInfo(nil), teammates...),
		width:        0,
		spinner:      SpinnerDots,
		runningColor: style.ANSICyan,
		doneColor:    style.ANSIGreen,
		errorColor:   style.ANSIRed,
		statusColor:  style.ANSIDim + style.ANSIWhite,
		nameColor:    style.ANSIReset,
		mutedColor:   style.ANSIDim,
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Init initializes the component.
func (t *TeammateSpinnerTree) Init() tea.Cmd {
	if t.hasRunningTeammate(t.teammates) {
		return t.tick()
	}
	return nil
}

// Update handles animation ticks and window resizing.
func (t *TeammateSpinnerTree) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
	case teammateSpinnerTreeTickMsg:
		if t.hasRunningTeammate(t.teammates) {
			t.spinIdx++
			return t, t.tick()
		}
	}

	return t, nil
}

// View renders the recursive spinner tree.
func (t *TeammateSpinnerTree) View() string {
	if len(t.teammates) == 0 {
		return ""
	}

	lines := make([]string, 0, len(t.teammates)*2)
	for idx := range t.teammates {
		last := idx == len(t.teammates)-1
		t.renderNode(&lines, t.teammates[idx], "", last)
	}

	for i := range lines {
		lines[i] = t.fitToWidth(lines[i])
	}

	return strings.Join(lines, "\n") + "\n"
}

// Focus marks component focused.
func (t *TeammateSpinnerTree) Focus() { t.focused = true }

// Blur marks component unfocused.
func (t *TeammateSpinnerTree) Blur() { t.focused = false }

// Focused reports focus state.
func (t *TeammateSpinnerTree) Focused() bool { return t.focused }

type teammateSpinnerTreeTickMsg struct{}

func (t *TeammateSpinnerTree) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg {
		return teammateSpinnerTreeTickMsg{}
	})
}

func (t *TeammateSpinnerTree) renderNode(lines *[]string, node TeammateInfo, prefix string, isLast bool) {
	connector := "├──"
	nextPrefix := prefix + "│   "
	if isLast {
		connector = "└──"
		nextPrefix = prefix + "    "
	}

	line := prefix + t.mutedColor + connector + style.ANSIReset + " " + t.renderEntry(node)
	*lines = append(*lines, line)

	for i := range node.Children {
		childLast := i == len(node.Children)-1
		t.renderNode(lines, node.Children[i], nextPrefix, childLast)
	}
}

func (t *TeammateSpinnerTree) renderEntry(node TeammateInfo) string {
	name := strings.TrimSpace(node.Name)
	if name == "" {
		name = "(unnamed)"
	}

	status := strings.TrimSpace(node.Status)
	message := strings.TrimSpace(node.Message)

	spinnerOrIcon := t.statusGlyph(status)
	statusText := status
	if statusText == "" {
		statusText = "running"
	}

	entry := fmt.Sprintf("%s %s%s%s %s[%s]%s", spinnerOrIcon, t.nameColor, name, style.ANSIReset, t.statusColor, statusText, style.ANSIReset)
	if message != "" {
		entry += " " + t.mutedColor + "— " + message + style.ANSIReset
	}

	return entry
}

func (t *TeammateSpinnerTree) statusGlyph(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "running", "in_progress", "in-progress", "working", "active":
		return t.runningColor + t.spinner.GetFrame(t.spinIdx) + style.ANSIReset
	case "done", "success", "completed", "complete", "ok":
		return t.doneColor + "✓" + style.ANSIReset
	case "error", "failed", "failure":
		return t.errorColor + "✗" + style.ANSIReset
	default:
		return t.mutedColor + "•" + style.ANSIReset
	}
}

func (t *TeammateSpinnerTree) hasRunningTeammate(nodes []TeammateInfo) bool {
	for _, node := range nodes {
		status := strings.ToLower(strings.TrimSpace(node.Status))
		if status == "" || status == "running" || status == "in_progress" || status == "in-progress" || status == "working" || status == "active" {
			return true
		}
		if t.hasRunningTeammate(node.Children) {
			return true
		}
	}
	return false
}

func (t *TeammateSpinnerTree) fitToWidth(line string) string {
	if t.width <= 0 {
		return line
	}
	plain := stripANSI(line)
	if lipgloss.Width(plain) <= t.width {
		return line
	}
	return style.Truncate(plain, t.width, "…")
}

func (t *TeammateSpinnerTree) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}

	if accent := style.ANSIColorFromHex(tokens.Accent); accent != "" {
		t.runningColor = accent
		t.doneColor = accent
	}
	if fg := style.ANSIColorFromHex(tokens.Color); fg != "" {
		t.statusColor = fg
		t.nameColor = fg
		t.mutedColor = fg
	}
	if danger := style.ANSIColorFromHex(tokens.ErrorBright); danger != "" {
		t.errorColor = danger
	}
}
var _ tui.Component = (*TeammateSpinnerTree)(nil)
