package display

import (
	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// FileStatus is a git-style file status.
type FileStatus int

const (
	FileStatusModified FileStatus = iota
	FileStatusAdded
	FileStatusDeleted
	FileStatusRenamed
	FileStatusCopied
	FileStatusUntracked
	FileStatusConflicted
)

// FileStatusBadge renders a colored git-style status badge.
type FileStatusBadge struct {
	status       FileStatus
	compact      bool
	focused      bool
	designTokens *design.DesignTokens
}

// FileStatusBadgeOption configures a FileStatusBadge.
type FileStatusBadgeOption func(*FileStatusBadge)

// WithFileStatusBadgeDesignTokens applies design-system tokens.
func WithFileStatusBadgeDesignTokens(tokens *design.DesignTokens) FileStatusBadgeOption {
	return func(b *FileStatusBadge) {
		if tokens != nil {
			b.designTokens = tokens
		}
	}
}

// WithFileStatusBadgeCompact toggles compact rendering (letter only, no brackets).
func WithFileStatusBadgeCompact(compact bool) FileStatusBadgeOption {
	return func(b *FileStatusBadge) {
		b.compact = compact
	}
}

// NewFileStatusBadge creates a new file status badge.
func NewFileStatusBadge(status FileStatus, opts ...FileStatusBadgeOption) *FileStatusBadge {
	b := &FileStatusBadge{
		status:       status,
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// Init initializes the component.
func (b *FileStatusBadge) Init() tea.Cmd {
	return nil
}

// Update handles Bubble Tea messages.
func (b *FileStatusBadge) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	return b, nil
}

// View renders the file status badge.
func (b *FileStatusBadge) View() string {
	text := b.badgeText()
	out := b.colorForStatus() + text + style.ANSIReset
	if b.focused {
		out = style.ANSIInverse + out + style.ANSIReset
	}
	return out
}

// Focus marks the component as focused.
func (b *FileStatusBadge) Focus() {
	b.focused = true
}

// Blur marks the component as unfocused.
func (b *FileStatusBadge) Blur() {
	b.focused = false
}

// Focused reports whether the component is focused.
func (b *FileStatusBadge) Focused() bool {
	return b.focused
}

func (b *FileStatusBadge) badgeText() string {
	letter := b.statusLetter()
	if b.compact {
		return letter
	}
	return "[" + letter + "]"
}

func (b *FileStatusBadge) statusLetter() string {
	switch b.status {
	case FileStatusModified:
		return "M"
	case FileStatusAdded:
		return "A"
	case FileStatusDeleted:
		return "D"
	case FileStatusRenamed:
		return "R"
	case FileStatusCopied:
		return "C"
	case FileStatusUntracked:
		return "U"
	case FileStatusConflicted:
		return "!"
	default:
		return "?"
	}
}

func (b *FileStatusBadge) colorForStatus() string {
	switch b.status {
	case FileStatusModified:
		return style.Fg("#EAB308") // AccentYellow
	case FileStatusAdded:
		return b.successColor("#22C55E") // AccentGreen
	case FileStatusDeleted:
		return b.errorColor("#EF4444") // AccentRed
	case FileStatusRenamed:
		return style.Fg("#3B82F6") // AccentBlue
	case FileStatusCopied:
		return style.Fg("#06B6D4") // AccentCyan
	case FileStatusUntracked:
		return b.mutedColor("#9CA3AF") // Gray
	case FileStatusConflicted:
		return b.warningColor("#F97316") // AccentOrange
	default:
		return style.ANSIWhite
	}
}

func (b *FileStatusBadge) successColor(fallback string) string {
	if b.designTokens != nil {
		if b.designTokens.SuccessBright != "" {
			if fg := style.Fg(b.designTokens.SuccessBright); fg != "" {
				return fg
			}
		}
	}
	return style.Fg(fallback)
}

func (b *FileStatusBadge) errorColor(fallback string) string {
	if b.designTokens != nil {
		if b.designTokens.ErrorBright != "" {
			if fg := style.Fg(b.designTokens.ErrorBright); fg != "" {
				return fg
			}
		}
	}
	return style.Fg(fallback)
}

func (b *FileStatusBadge) warningColor(fallback string) string {
	if b.designTokens != nil {
		if b.designTokens.PendingColor != "" {
			if fg := style.Fg(b.designTokens.PendingColor); fg != "" {
				return fg
			}
		}
	}
	return style.Fg(fallback)
}

func (b *FileStatusBadge) mutedColor(fallback string) string {
	if b.designTokens != nil {
		if b.designTokens.MutedColor != "" {
			if fg := style.Fg(b.designTokens.MutedColor); fg != "" {
				return fg
			}
		}
	}
	return style.Fg(fallback)
}

var _ tui.Component = (*FileStatusBadge)(nil)
