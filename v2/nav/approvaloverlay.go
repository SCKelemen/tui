package nav

import (
	"fmt"
	"strings"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	"github.com/SCKelemen/tui/v2/style/design"
	tea "github.com/charmbracelet/bubbletea"
)

// ApprovalScope controls how long an approval decision should persist.
type ApprovalScope int

const (
	// ApprovalScopeOnce approves/denies only this request.
	ApprovalScopeOnce ApprovalScope = iota
	// ApprovalScopeSession approves for current session.
	ApprovalScopeSession
	// ApprovalScopeAlways approves permanently.
	ApprovalScopeAlways
	// ApprovalScopeDenied is an explicit deny result.
	ApprovalScopeDenied
)

// ApprovalDecisionMsg is emitted when an overlay decision is made.
type ApprovalDecisionMsg struct {
	Approved bool
	Scope    ApprovalScope
}

// ApprovalOverlay models a full-screen approval modal rendered over background content.
type ApprovalOverlay struct {
	visible      bool
	requestType  string
	summary      string
	detail       string
	background   string
	width        int
	height       int
	focused      bool
	selection    int
	designTokens *design.DesignTokens
}

// ApprovalOverlayOption configures an ApprovalOverlay.
type ApprovalOverlayOption func(*ApprovalOverlay)

// WithApprovalOverlayRequestType sets request type (command/file/network/etc.).
func WithApprovalOverlayRequestType(requestType string) ApprovalOverlayOption {
	return func(o *ApprovalOverlay) { o.requestType = strings.TrimSpace(requestType) }
}

// WithApprovalOverlaySummary sets concise summary text.
func WithApprovalOverlaySummary(summary string) ApprovalOverlayOption {
	return func(o *ApprovalOverlay) { o.summary = strings.TrimSpace(summary) }
}

// WithApprovalOverlayDetail sets full command or diff detail.
func WithApprovalOverlayDetail(detail string) ApprovalOverlayOption {
	return func(o *ApprovalOverlay) { o.detail = detail }
}

// WithApprovalOverlayVisible sets initial visibility.
func WithApprovalOverlayVisible(visible bool) ApprovalOverlayOption {
	return func(o *ApprovalOverlay) { o.visible = visible }
}

// WithApprovalOverlayDesignTokens applies design tokens.
func WithApprovalOverlayDesignTokens(tokens *design.DesignTokens) ApprovalOverlayOption {
	return func(o *ApprovalOverlay) {
		if tokens != nil {
			o.designTokens = tokens
		}
	}
}

// NewApprovalOverlay creates a full approval modal component.
func NewApprovalOverlay(opts ...ApprovalOverlayOption) *ApprovalOverlay {
	o := &ApprovalOverlay{
		visible:      true,
		requestType:  "permission",
		summary:      "Approval required",
		detail:       "",
		background:   "",
		selection:    0,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// Init satisfies the Bubble Tea model contract.
func (o *ApprovalOverlay) Init() tea.Cmd { return nil }

// Update handles keyboard shortcuts and selection.
func (o *ApprovalOverlay) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch t := msg.(type) {
	case tea.WindowSizeMsg:
		o.width = t.Width
		o.height = t.Height
	case tea.KeyMsg:
		if !o.visible || !o.focused {
			return o, nil
		}

		switch t.String() {
		case "y":
			return o, o.emitDecision(true, ApprovalScopeOnce)
		case "n":
			return o, o.emitDecision(false, ApprovalScopeDenied)
		case "a":
			return o, o.emitDecision(true, ApprovalScopeAlways)
		case "s":
			return o, o.emitDecision(true, ApprovalScopeSession)
		case "up", "k":
			if o.selection > 0 {
				o.selection--
			}
		case "down", "j":
			if o.selection < 3 {
				o.selection++
			}
		case "enter":
			switch o.selection {
			case 0:
				return o, o.emitDecision(true, ApprovalScopeOnce)
			case 1:
				return o, o.emitDecision(true, ApprovalScopeSession)
			case 2:
				return o, o.emitDecision(true, ApprovalScopeAlways)
			default:
				return o, o.emitDecision(false, ApprovalScopeDenied)
			}
		}
	}

	return o, nil
}

// View renders background content plus the modal approval panel.
func (o *ApprovalOverlay) View() string {
	if !o.visible {
		return o.background
	}

	titleColor := style.ANSIYellow
	accent := style.ANSICyan
	muted := style.ANSIDim
	if o.designTokens != nil {
		if v := style.Fg(o.designTokens.PendingColor); v != "" {
			titleColor = v
		}
		if v := style.Fg(o.designTokens.Accent); v != "" {
			accent = v
		}
		if v := style.Fg(o.designTokens.MutedColor); v != "" {
			muted = v
		}
	}

	options := []string{
		o.optionLine(0, "Approve once", "y"),
		o.optionLine(1, "Approve for session", "s"),
		o.optionLine(2, "Always approve", "a"),
		o.optionLine(3, "Deny", "n"),
	}

	panel := []string{
		titleColor + style.ANSIBold + "⚠ Approval Requested" + style.ANSIReset,
		fmt.Sprintf("Type: %s", o.requestType),
		o.summary,
		"",
		accent + "Detail:" + style.ANSIReset,
		o.detail,
		"",
		accent + "Actions:" + style.ANSIReset,
	}
	panel = append(panel, options...)
	panel = append(panel, "", muted+"shortcuts: y=yes n=no a=always s=session"+style.ANSIReset)

	return strings.TrimRight(o.background, "\n") + "\n\n" + strings.Join(panel, "\n")
}

// SetBackgroundView updates the beneath-content render snapshot.
func (o *ApprovalOverlay) SetBackgroundView(background string) { o.background = background }

// Focus marks focus state.
func (o *ApprovalOverlay) Focus() { o.focused = true }

// Blur marks blur state.
func (o *ApprovalOverlay) Blur() { o.focused = false }

// Focused reports focus state.
func (o *ApprovalOverlay) Focused() bool { return o.focused }

func (o *ApprovalOverlay) optionLine(idx int, text, key string) string {
	prefix := "  "
	if idx == o.selection {
		prefix = "> "
	}
	return fmt.Sprintf("%s%s (%s)", prefix, text, key)
}

func (o *ApprovalOverlay) emitDecision(approved bool, scope ApprovalScope) tea.Cmd {
	o.visible = false
	return func() tea.Msg {
		return ApprovalDecisionMsg{Approved: approved, Scope: scope}
	}
}

var _ tui.Component = (*ApprovalOverlay)(nil)
