package container

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ScreenSizeOverlay conditionally blocks rendering when terminal dimensions are too small.
type ScreenSizeOverlay struct {
	child tui.Component

	minWidth  int
	minHeight int
	message   string

	width   int
	height  int
	focused bool

	designTokens *design.DesignTokens
}

// ScreenSizeOverlayOption configures a ScreenSizeOverlay.
type ScreenSizeOverlayOption func(*ScreenSizeOverlay)

// WithScreenSizeMinWidth sets the minimum required terminal width.
func WithScreenSizeMinWidth(width int) ScreenSizeOverlayOption {
	return func(s *ScreenSizeOverlay) {
		if width > 0 {
			s.minWidth = width
		}
	}
}

// WithScreenSizeMinHeight sets the minimum required terminal height.
func WithScreenSizeMinHeight(height int) ScreenSizeOverlayOption {
	return func(s *ScreenSizeOverlay) {
		if height > 0 {
			s.minHeight = height
		}
	}
}

// WithScreenSizeMessage sets the overlay message displayed when the window is too small.
func WithScreenSizeMessage(message string) ScreenSizeOverlayOption {
	return func(s *ScreenSizeOverlay) {
		if strings.TrimSpace(message) != "" {
			s.message = message
		}
	}
}

// WithScreenSizeDesignTokens applies design-system colors to overlay rendering.
func WithScreenSizeDesignTokens(tokens *design.DesignTokens) ScreenSizeOverlayOption {
	return func(s *ScreenSizeOverlay) {
		if tokens != nil {
			s.designTokens = tokens
		}
	}
}

// NewScreenSizeOverlay creates a new ScreenSizeOverlay.
func NewScreenSizeOverlay(child tui.Component, opts ...ScreenSizeOverlayOption) *ScreenSizeOverlay {
	s := &ScreenSizeOverlay{
		child:        child,
		minWidth:     60,
		minHeight:    20,
		message:      "Expand window to view",
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.child == nil {
		s.child = &screenSizeOverlayEmptyComponent{}
	}

	return s
}

// Init initializes the wrapped child component.
func (s *ScreenSizeOverlay) Init() tea.Cmd {
	if s.child == nil {
		return nil
	}
	return s.child.Init()
}

// Update tracks current dimensions and forwards messages to child when minimum size is met.
func (s *ScreenSizeOverlay) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = m.Width
		s.height = m.Height
	}

	if s.isTooSmall() {
		return s, nil
	}

	if s.child == nil {
		return s, nil
	}

	updated, cmd := s.child.Update(msg)
	s.child = updated
	return s, cmd
}

// View renders either the child view or a centered screen-size warning overlay.
func (s *ScreenSizeOverlay) View() string {
	if s.isTooSmall() {
		return s.overlayView()
	}
	if s.child == nil {
		return ""
	}
	return s.child.View()
}

// Focus marks this component as focused and forwards focus to child.
func (s *ScreenSizeOverlay) Focus() {
	s.focused = true
	if s.child != nil {
		s.child.Focus()
	}
}

// Blur marks this component as unfocused and forwards blur to child.
func (s *ScreenSizeOverlay) Blur() {
	s.focused = false
	if s.child != nil {
		s.child.Blur()
	}
}

// Focused reports whether this component is focused.
func (s *ScreenSizeOverlay) Focused() bool {
	return s.focused
}

func (s *ScreenSizeOverlay) isTooSmall() bool {
	if s.width <= 0 || s.height <= 0 {
		return false
	}
	return s.width < s.minWidth || s.height < s.minHeight
}

func (s *ScreenSizeOverlay) overlayView() string {
	title := "⚠ " + strings.TrimSpace(s.message)
	details := fmt.Sprintf("Current: %dx%d  •  Minimum: %dx%d", s.width, s.height, s.minWidth, s.minHeight)

	warningColor := style.ANSIYellow
	mutedColor := style.ANSIDim
	if s.designTokens != nil {
		if s.designTokens.PendingColor != "" {
			if c := style.Fg(s.designTokens.PendingColor); c != "" {
				warningColor = c
			}
		} else if s.designTokens.Accent != "" {
			if c := style.Fg(s.designTokens.Accent); c != "" {
				warningColor = c
			}
		}

		if s.designTokens.MutedColor != "" {
			if c := style.Fg(s.designTokens.MutedColor); c != "" {
				mutedColor = c
			}
		} else if s.designTokens.Color != "" {
			if c := style.Fg(s.designTokens.Color); c != "" {
				mutedColor = c
			}
		}
	}

	maxWidth := s.width
	if maxWidth <= 0 {
		maxWidth = max(style.StringWidth(title), style.StringWidth(details))
	}

	title = style.Truncate(title, maxWidth, "…")
	details = style.Truncate(details, maxWidth, "…")

	titlePad := 0
	detailPad := 0
	if s.width > 0 {
		titlePad = (s.width - style.StringWidth(title)) / 2
		detailPad = (s.width - style.StringWidth(details)) / 2
		if titlePad < 0 {
			titlePad = 0
		}
		if detailPad < 0 {
			detailPad = 0
		}
	}

	titleLine := strings.Repeat(" ", titlePad) + warningColor + style.ANSIBold + title + style.ANSIReset
	detailLine := strings.Repeat(" ", detailPad) + mutedColor + details + style.ANSIReset

	lines := []string{titleLine, detailLine}

	if s.height <= 0 {
		return strings.Join(lines, "\n")
	}

	topPad := (s.height - len(lines)) / 2
	if topPad < 0 {
		topPad = 0
	}
	bottomPad := s.height - topPad - len(lines)
	if bottomPad < 0 {
		bottomPad = 0
	}

	out := make([]string, 0, s.height)
	for i := 0; i < topPad; i++ {
		out = append(out, "")
	}
	out = append(out, lines...)
	for i := 0; i < bottomPad; i++ {
		out = append(out, "")
	}

	return strings.Join(out, "\n")
}

type screenSizeOverlayEmptyComponent struct{}

func (e *screenSizeOverlayEmptyComponent) Init() tea.Cmd {
	return nil
}

func (e *screenSizeOverlayEmptyComponent) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	return e, nil
}

func (e *screenSizeOverlayEmptyComponent) View() string {
	return ""
}

func (e *screenSizeOverlayEmptyComponent) Focus() {}

func (e *screenSizeOverlayEmptyComponent) Blur() {}

func (e *screenSizeOverlayEmptyComponent) Focused() bool {
	return false
}

var _ tui.Component = (*ScreenSizeOverlay)(nil)
var _ tui.Component = (*screenSizeOverlayEmptyComponent)(nil)
