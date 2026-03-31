package display

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// StatusIconType identifies the status icon semantic.
type StatusIconType int

const (
	StatusSuccess StatusIconType = iota
	StatusError
	StatusWarning
	StatusInfo
	StatusPending
	StatusRunning
)

// StatusIcon renders a status indicator icon with optional label.
type StatusIcon struct {
	status       StatusIconType
	label        string
	focused      bool
	designTokens *design.DesignTokens
}

// StatusIconOption configures a StatusIcon.
type StatusIconOption func(*StatusIcon)

// WithStatusIconLabel sets the optional text label.
func WithStatusIconLabel(label string) StatusIconOption {
	return func(s *StatusIcon) {
		s.label = strings.TrimSpace(label)
	}
}

// WithStatusIconDesignTokens applies design tokens.
func WithStatusIconDesignTokens(tokens *design.DesignTokens) StatusIconOption {
	return func(s *StatusIcon) {
		if tokens != nil {
			s.designTokens = tokens
		}
	}
}

// NewStatusIcon creates a StatusIcon component.
func NewStatusIcon(status StatusIconType, opts ...StatusIconOption) *StatusIcon {
	s := &StatusIcon{
		status:       status,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Init initializes the component.
func (s *StatusIcon) Init() tea.Cmd { return nil }

// Update handles Bubble Tea messages.
func (s *StatusIcon) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	return s, nil
}

// View renders the status icon and optional label.
func (s *StatusIcon) View() string {
	icon, fallback := s.iconAndColor()
	fg := fallback
	if s.designTokens != nil {
		switch s.status {
		case StatusSuccess:
			if c := style.Fg(s.designTokens.SuccessBright); c != "" {
				fg = c
			}
		case StatusError:
			if c := style.Fg(s.designTokens.ErrorBright); c != "" {
				fg = c
			}
		case StatusWarning:
			if c := style.Fg(s.designTokens.PendingColor); c != "" {
				fg = c
			}
		case StatusInfo:
			if c := style.Fg(s.designTokens.Accent); c != "" {
				fg = c
			}
		case StatusPending, StatusRunning:
			if c := style.Fg(s.designTokens.MutedColor); c != "" {
				fg = c
			}
		}
	}

	out := fg + icon + style.ANSIReset
	if s.label != "" {
		labelColor := style.ANSIDim
		if s.designTokens != nil {
			if c := style.Fg(s.designTokens.MutedColor); c != "" {
				labelColor = c
			}
		}
		out += " " + labelColor + s.label + style.ANSIReset
	}

	return out
}

// Focus marks the component as focused.
func (s *StatusIcon) Focus() { s.focused = true }

// Blur marks the component as unfocused.
func (s *StatusIcon) Blur() { s.focused = false }

// Focused reports whether the component is focused.
func (s *StatusIcon) Focused() bool { return s.focused }

func (s *StatusIcon) iconAndColor() (string, string) {
	switch s.status {
	case StatusSuccess:
		return "✓", style.ANSIGreen
	case StatusError:
		return "✗", style.ANSIRed
	case StatusWarning:
		return "⚠", style.ANSIYellow
	case StatusInfo:
		return "ℹ", style.ANSICyan
	case StatusPending:
		return "○", style.ANSIDim
	case StatusRunning:
		return "⟳", style.ANSIBlue
	default:
		return "○", style.ANSIWhite
	}
}

var _ tui.Component = (*StatusIcon)(nil)
