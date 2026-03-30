package input

import (
	"fmt"
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// SessionStatus represents a session lifecycle state.
type SessionStatus int

const (
	SessionActive SessionStatus = iota
	SessionDraft
	SessionCompleted
	SessionArchived
)

// SessionItem is one row in the session selector.
type SessionItem struct {
	ID           string
	Name         string
	Status       SessionStatus
	CreatedAt    time.Time
	MessageCount int
	IsChild      bool
}

// SessionSelectedMsg is emitted when Enter is pressed on a session.
type SessionSelectedMsg struct {
	ID string
}

// SessionDeletedMsg is emitted when delete is requested on a session.
type SessionDeletedMsg struct {
	ID string
}

type sessionSelectorTickMsg struct{}

// SessionSelectorOption configures a SessionSelector.
type SessionSelectorOption func(*SessionSelector)

// WithSessionSelectorDesignTokens applies design-system colors.
func WithSessionSelectorDesignTokens(tokens *design.DesignTokens) SessionSelectorOption {
	return func(s *SessionSelector) {
		s.applyDesignTokens(tokens)
	}
}

// WithSessionSelectorWidth sets a preferred render width.
func WithSessionSelectorWidth(width int) SessionSelectorOption {
	return func(s *SessionSelector) {
		if width > 0 {
			s.width = width
		}
	}
}

// WithSessionSelectorSelected sets the initial selected index.
func WithSessionSelectorSelected(index int) SessionSelectorOption {
	return func(s *SessionSelector) {
		s.cursor = index
	}
}

// SessionSelector renders a navigable session drawer list.
type SessionSelector struct {
	sessions     []SessionItem
	cursor       int
	focused      bool
	width        int
	designTokens *design.DesignTokens
	frameIdx     int

	accentColor    string
	mutedColor     string
	completeColor  string
	archivedStyle  string
	cursorBgColor  string
	metaColor      string
	activeDotStyle string
}

// NewSessionSelector creates a new SessionSelector component.
func NewSessionSelector(sessions []SessionItem, opts ...SessionSelectorOption) *SessionSelector {
	s := &SessionSelector{
		sessions:       append([]SessionItem(nil), sessions...),
		cursor:         0,
		focused:        false,
		width:          0,
		designTokens:   design.DefaultTheme(),
		frameIdx:       0,
		accentColor:    style.ANSICyan,
		mutedColor:     style.ANSIDim,
		completeColor:  style.ANSIGreen,
		archivedStyle:  style.ANSIDim,
		cursorBgColor:  style.ANSIInverse,
		metaColor:      style.ANSIDim,
		activeDotStyle: style.ANSIDim,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.clampCursor()
	s.applyDesignTokens(s.designTokens)

	return s
}

// Init initializes the component.
func (s *SessionSelector) Init() tea.Cmd {
	return s.tick()
}

// Update handles keyboard and window size messages.
func (s *SessionSelector) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if s.width <= 0 {
			s.width = msg.Width
		}
		return s, nil

	case sessionSelectorTickMsg:
		s.frameIdx = (s.frameIdx + 1) % 3
		return s, s.tick()

	case tea.KeyMsg:
		if !s.focused || len(s.sessions) == 0 {
			return s, nil
		}

		switch msg.String() {
		case "down", "j":
			if s.cursor < len(s.sessions)-1 {
				s.cursor++
			}
			return s, nil
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
			return s, nil
		case "enter":
			selected := s.sessions[s.cursor]
			return s, func() tea.Msg {
				return SessionSelectedMsg{ID: selected.ID}
			}
		case "d":
			selected := s.sessions[s.cursor]
			return s, func() tea.Msg {
				return SessionDeletedMsg{ID: selected.ID}
			}
		}
	}

	return s, nil
}

// View renders the session list.
func (s *SessionSelector) View() string {
	if len(s.sessions) == 0 {
		return ""
	}

	lines := make([]string, 0, len(s.sessions))
	for i, session := range s.sessions {
		line := s.renderSessionLine(session)
		if s.width > 0 {
			line = style.Pad(style.Truncate(line, s.width, "…"), s.width)
		}
		if s.focused && i == s.cursor {
			line = s.cursorBgColor + line + style.ANSIReset
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// Focus marks the component as focused.
func (s *SessionSelector) Focus() {
	s.focused = true
}

// Blur marks the component as unfocused.
func (s *SessionSelector) Blur() {
	s.focused = false
}

// Focused reports whether the component currently has focus.
func (s *SessionSelector) Focused() bool {
	return s.focused
}

func (s *SessionSelector) clampCursor() {
	if len(s.sessions) == 0 {
		s.cursor = 0
		return
	}
	if s.cursor < 0 {
		s.cursor = 0
	}
	if s.cursor >= len(s.sessions) {
		s.cursor = len(s.sessions) - 1
	}
}

func (s *SessionSelector) renderSessionLine(item SessionItem) string {
	left := s.renderLeft(item)
	right := s.renderRight(item)

	if s.width <= 0 {
		if right == "" {
			return left
		}
		return left + "  " + s.metaColor + right + style.ANSIReset
	}

	leftWidth := style.StringWidth(left)
	rightWidth := style.StringWidth(right)
	if rightWidth == 0 {
		return left
	}

	if leftWidth+2+rightWidth <= s.width {
		spaces := s.width - leftWidth - rightWidth
		if spaces < 2 {
			spaces = 2
		}
		return left + strings.Repeat(" ", spaces) + s.metaColor + right + style.ANSIReset
	}

	maxLeft := s.width - 2 - rightWidth
	if maxLeft < 4 {
		maxLeft = 4
	}
	trimmedLeft := style.Truncate(left, maxLeft, "…")
	spaceCount := s.width - style.StringWidth(trimmedLeft) - rightWidth
	if spaceCount < 1 {
		spaceCount = 1
	}

	return trimmedLeft + strings.Repeat(" ", spaceCount) + s.metaColor + right + style.ANSIReset
}

func (s *SessionSelector) renderLeft(item SessionItem) string {
	name := strings.TrimSpace(item.Name)
	if name == "" {
		name = "Untitled Session"
	}

	childPrefix := ""
	if item.IsChild {
		childPrefix = "  └─"
	}

	switch item.Status {
	case SessionActive:
		activeDots := s.activeDotsFrame()
		return s.accentColor + childPrefix + "▶ " + name + " " + s.activeDotStyle + activeDots + s.accentColor + style.ANSIReset
	case SessionCompleted:
		return s.completeColor + childPrefix + "✓ " + name + style.ANSIReset
	case SessionArchived:
		return s.archivedStyle + childPrefix + "  " + name + style.ANSIReset
	case SessionDraft:
		fallthrough
	default:
		return s.mutedColor + childPrefix + "  " + name + style.ANSIReset
	}
}

func (s *SessionSelector) renderRight(item SessionItem) string {
	timePart := formatTimeAgo(item.CreatedAt)
	msgPart := formatMessageCount(item.MessageCount)

	if timePart == "" {
		return msgPart
	}
	if msgPart == "" {
		return timePart
	}
	return timePart + " · " + msgPart
}

func (s *SessionSelector) activeDotsFrame() string {
	frames := []string{"◆◇◇", "◇◆◇", "◇◇◆"}
	return frames[s.frameIdx%len(frames)]
}

func (s *SessionSelector) tick() tea.Cmd {
	return tea.Tick(400*time.Millisecond, func(time.Time) tea.Msg {
		return sessionSelectorTickMsg{}
	})
}

func (s *SessionSelector) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}
	s.designTokens = tokens

	if accent := style.ANSIColorFromHex(tokens.Accent); accent != "" {
		s.accentColor = accent
	}
	if muted := style.ANSIColorFromHex(tokens.MutedColor); muted != "" {
		s.mutedColor = muted
		s.metaColor = muted
		s.activeDotStyle = muted
	} else if muted := style.ANSIColorFromHex(tokens.Color); muted != "" {
		s.mutedColor = muted
		s.metaColor = muted
	}
	if success := style.ANSIColorFromHex(tokens.SuccessBright); success != "" {
		s.completeColor = success
	}
	if bg := style.ANSIBackgroundColorFromHex(tokens.Accent); bg != "" {
		s.cursorBgColor = bg
	}
}

func formatTimeAgo(ts time.Time) string {
	if ts.IsZero() {
		return ""
	}

	d := time.Since(ts)
	if d < 0 {
		d = 0
	}

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d/time.Minute))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d/time.Hour))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d/(24*time.Hour)))
	default:
		return ts.Format("2006-01-02")
	}
}

func formatMessageCount(count int) string {
	if count <= 0 {
		return "0 msgs"
	}
	if count == 1 {
		return "1 msg"
	}
	return fmt.Sprintf("%d msgs", count)
}

var _ tui.Component = (*SessionSelector)(nil)
