package container

import (
	"strings"
	"unicode/utf8"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// SplitDirection controls how panes are arranged.
type SplitDirection int

const (
	// SplitHorizontal renders panes side-by-side (left/right).
	SplitHorizontal SplitDirection = iota
	// SplitVertical renders panes stacked (top/bottom).
	SplitVertical
)

// SplitPane renders two components in a resizable split layout.
type SplitPane struct {
	direction   SplitDirection
	ratio       float64
	minRatio    float64
	maxRatio    float64
	left        tui.Component
	right       tui.Component
	focusedPane int
	resizing    bool
	width       int
	height      int
	focused     bool

	designTokens *design.DesignTokens
	showDivider  bool
	dividerChar  string
}

// SplitPaneOption configures a SplitPane.
type SplitPaneOption func(*SplitPane)

// WithSplitPaneDesignTokens applies design-system tokens.
func WithSplitPaneDesignTokens(tokens *design.DesignTokens) SplitPaneOption {
	return func(s *SplitPane) {
		if tokens != nil {
			s.designTokens = tokens
		}
	}
}

// WithSplitPaneTheme applies a named theme.
func WithSplitPaneTheme(theme string) SplitPaneOption {
	return func(s *SplitPane) {
		s.designTokens = style.DesignTokensForTheme(theme)
	}
}

// WithDirection sets split direction.
func WithDirection(d SplitDirection) SplitPaneOption {
	return func(s *SplitPane) {
		oldDefault := defaultDividerChar(s.direction)
		s.direction = d
		if s.dividerChar == "" || s.dividerChar == oldDefault {
			s.dividerChar = defaultDividerChar(d)
		}
	}
}

// WithRatio sets split ratio.
func WithRatio(r float64) SplitPaneOption {
	return func(s *SplitPane) {
		s.SetRatio(r)
	}
}

// WithMinMaxRatio sets minimum and maximum ratio bounds.
func WithMinMaxRatio(min, max float64) SplitPaneOption {
	return func(s *SplitPane) {
		min = splitPaneClamp(min, 0.0, 1.0)
		max = splitPaneClamp(max, 0.0, 1.0)
		if min > max {
			min, max = max, min
		}
		s.minRatio = min
		s.maxRatio = max
		s.ratio = splitPaneClamp(s.ratio, s.minRatio, s.maxRatio)
	}
}

// WithLeftComponent sets the left pane component.
func WithLeftComponent(c tui.Component) SplitPaneOption {
	return func(s *SplitPane) {
		s.left = c
	}
}

// WithRightComponent sets the right pane component.
func WithRightComponent(c tui.Component) SplitPaneOption {
	return func(s *SplitPane) {
		s.right = c
	}
}

// WithTopComponent sets the top pane component (alias for left in vertical mode).
func WithTopComponent(c tui.Component) SplitPaneOption {
	return WithLeftComponent(c)
}

// WithBottomComponent sets the bottom pane component (alias for right in vertical mode).
func WithBottomComponent(c tui.Component) SplitPaneOption {
	return WithRightComponent(c)
}

// WithShowDivider toggles divider visibility.
func WithShowDivider(show bool) SplitPaneOption {
	return func(s *SplitPane) {
		s.showDivider = show
	}
}

// WithDividerChar sets the divider character/string.
func WithDividerChar(ch string) SplitPaneOption {
	return func(s *SplitPane) {
		s.dividerChar = ch
	}
}

// NewSplitPane creates a split pane with defaults.
func NewSplitPane(opts ...SplitPaneOption) *SplitPane {
	s := &SplitPane{
		direction:    SplitHorizontal,
		ratio:        0.5,
		minRatio:     0.1,
		maxRatio:     0.9,
		focusedPane:  0,
		designTokens: design.DefaultTheme(),
		showDivider:  true,
		dividerChar:  "│",
	}

	for _, opt := range opts {
		opt(s)
	}

	s.ratio = splitPaneClamp(s.ratio, s.minRatio, s.maxRatio)
	s.syncPaneFocus()
	return s
}

// Init initializes child components.
func (s *SplitPane) Init() tea.Cmd {
	var cmds []tea.Cmd
	if s.left != nil {
		cmds = append(cmds, s.left.Init())
	}
	if s.right != nil {
		cmds = append(cmds, s.right.Init())
	}
	return tea.Batch(cmds...)
}

// Update handles layout, keyboard controls, and delegates to child panes.
func (s *SplitPane) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = m.Width
		s.height = m.Height
		return s, s.updateChildSizes()

	case tea.KeyMsg:
		if !s.focused {
			return s, nil
		}

		switch m.String() {
		case "ctrl+h":
			s.focusPane(0)
			return s, nil
		case "ctrl+l":
			s.focusPane(1)
			return s, nil
		case "ctrl+r":
			s.resizing = !s.resizing
			return s, nil
		case "esc":
			if s.resizing {
				s.resizing = false
				return s, nil
			}
		}

		if s.resizing {
			switch m.String() {
			case "h", "-":
				s.SetRatio(s.ratio - 0.05)
				return s, s.updateChildSizes()
			case "l", "+", "=":
				s.SetRatio(s.ratio + 0.05)
				return s, s.updateChildSizes()
			}
		}
	}

	child := s.focusedChild()
	if child == nil {
		return s, nil
	}

	updated, cmd := child.Update(msg)
	if s.focusedPane == 0 {
		s.left = updated
	} else {
		s.right = updated
	}
	return s, cmd
}

// View renders the split pane.
func (s *SplitPane) View() string {
	if s.width <= 0 || s.height <= 0 {
		return ""
	}

	if s.direction == SplitVertical {
		return s.renderVertical()
	}
	return s.renderHorizontal()
}

// Focus marks the split pane as focused.
func (s *SplitPane) Focus() {
	s.focused = true
	s.syncPaneFocus()
}

// Blur marks the split pane as unfocused.
func (s *SplitPane) Blur() {
	s.focused = false
	if s.left != nil {
		s.left.Blur()
	}
	if s.right != nil {
		s.right.Blur()
	}
}

// Focused reports focus state.
func (s *SplitPane) Focused() bool {
	return s.focused
}

// SetRatio sets the split ratio using configured min/max bounds.
func (s *SplitPane) SetRatio(r float64) {
	s.ratio = splitPaneClamp(r, s.minRatio, s.maxRatio)
}

// GetRatio returns the active split ratio.
func (s *SplitPane) GetRatio() float64 {
	return s.ratio
}

// FocusLeft focuses the left pane.
func (s *SplitPane) FocusLeft() {
	s.focusPane(0)
}

// FocusRight focuses the right pane.
func (s *SplitPane) FocusRight() {
	s.focusPane(1)
}

// FocusTop focuses the top pane (alias of left in vertical mode).
func (s *SplitPane) FocusTop() {
	s.FocusLeft()
}

// FocusBottom focuses the bottom pane (alias of right in vertical mode).
func (s *SplitPane) FocusBottom() {
	s.FocusRight()
}

// GetFocusedPane returns the focused pane index (0 or 1).
func (s *SplitPane) GetFocusedPane() int {
	return s.focusedPane
}

// ToggleFocus switches focus between panes.
func (s *SplitPane) ToggleFocus() {
	if s.focusedPane == 0 {
		s.focusPane(1)
		return
	}
	s.focusPane(0)
}

// SetLeftComponent sets the left/top child component.
func (s *SplitPane) SetLeftComponent(c tui.Component) {
	s.left = c
	s.syncPaneFocus()
	_ = s.updateChildSizes()
}

// SetRightComponent sets the right/bottom child component.
func (s *SplitPane) SetRightComponent(c tui.Component) {
	s.right = c
	s.syncPaneFocus()
	_ = s.updateChildSizes()
}

// EnterResizeMode enables resize mode.
func (s *SplitPane) EnterResizeMode() {
	s.resizing = true
}

// ExitResizeMode disables resize mode.
func (s *SplitPane) ExitResizeMode() {
	s.resizing = false
}

// IsResizing reports whether resize mode is active.
func (s *SplitPane) IsResizing() bool {
	return s.resizing
}

func (s *SplitPane) focusPane(index int) {
	if index < 0 || index > 1 {
		return
	}
	s.focusedPane = index
	s.syncPaneFocus()
}

func (s *SplitPane) syncPaneFocus() {
	if s.left == nil && s.right == nil {
		return
	}

	if !s.focused {
		if s.left != nil {
			s.left.Blur()
		}
		if s.right != nil {
			s.right.Blur()
		}
		return
	}

	if s.focusedPane == 0 {
		if s.left != nil {
			s.left.Focus()
		}
		if s.right != nil {
			s.right.Blur()
		}
		return
	}

	if s.left != nil {
		s.left.Blur()
	}
	if s.right != nil {
		s.right.Focus()
	}
}

func (s *SplitPane) focusedChild() tui.Component {
	if s.focusedPane == 0 {
		return s.left
	}
	return s.right
}

func (s *SplitPane) renderHorizontal() string {
	leftWidth, rightWidth, dividerWidth := s.horizontalDimensions()
	leftLines := s.renderPaneLines(s.left, leftWidth, s.height)
	rightLines := s.renderPaneLines(s.right, rightWidth, s.height)
	divider := s.dividerVisual()

	lines := make([]string, s.height)
	for i := 0; i < s.height; i++ {
		line := leftLines[i]
		if dividerWidth > 0 {
			line += divider
		}
		line += rightLines[i]
		lines[i] = line
	}

	return strings.Join(lines, "\n")
}

func (s *SplitPane) renderVertical() string {
	topHeight, bottomHeight, dividerHeight := s.verticalDimensions()
	topLines := s.renderPaneLines(s.left, s.width, topHeight)
	bottomLines := s.renderPaneLines(s.right, s.width, bottomHeight)
	d := s.dividerVisual()

	lines := make([]string, 0, s.height)
	lines = append(lines, topLines...)
	if dividerHeight > 0 {
		dividerLine := strings.Repeat(d, s.width)
		lines = append(lines, dividerLine)
	}
	lines = append(lines, bottomLines...)

	if len(lines) < s.height {
		for i := len(lines); i < s.height; i++ {
			lines = append(lines, strings.Repeat(" ", s.width))
		}
	}
	if len(lines) > s.height {
		lines = lines[:s.height]
	}

	return strings.Join(lines, "\n")
}

func (s *SplitPane) renderPaneLines(c tui.Component, width, height int) []string {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	blank := strings.Repeat(" ", width)
	if height == 0 {
		return []string{}
	}

	if c == nil {
		lines := make([]string, height)
		for i := range lines {
			lines[i] = blank
		}
		return lines
	}

	viewLines := strings.Split(c.View(), "\n")
	lines := make([]string, height)
	for i := 0; i < height; i++ {
		if i < len(viewLines) {
			lines[i] = clipOrPad(viewLines[i], width)
		} else {
			lines[i] = blank
		}
	}
	return lines
}

func (s *SplitPane) updateChildSizes() tea.Cmd {
	var cmds []tea.Cmd

	if s.direction == SplitVertical {
		topHeight, bottomHeight, _ := s.verticalDimensions()
		if s.left != nil {
			updated, cmd := s.left.Update(tea.WindowSizeMsg{Width: s.width, Height: topHeight})
			s.left = updated
			cmds = append(cmds, cmd)
		}
		if s.right != nil {
			updated, cmd := s.right.Update(tea.WindowSizeMsg{Width: s.width, Height: bottomHeight})
			s.right = updated
			cmds = append(cmds, cmd)
		}
		return tea.Batch(cmds...)
	}

	leftWidth, rightWidth, _ := s.horizontalDimensions()
	if s.left != nil {
		updated, cmd := s.left.Update(tea.WindowSizeMsg{Width: leftWidth, Height: s.height})
		s.left = updated
		cmds = append(cmds, cmd)
	}
	if s.right != nil {
		updated, cmd := s.right.Update(tea.WindowSizeMsg{Width: rightWidth, Height: s.height})
		s.right = updated
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func (s *SplitPane) horizontalDimensions() (int, int, int) {
	dividerWidth := 0
	if s.showDivider {
		dividerWidth = 1
	}

	available := s.width - dividerWidth
	if available < 0 {
		available = 0
	}

	leftWidth := int(float64(available) * s.ratio)
	if leftWidth < 0 {
		leftWidth = 0
	}
	if leftWidth > available {
		leftWidth = available
	}
	rightWidth := available - leftWidth
	return leftWidth, rightWidth, dividerWidth
}

func (s *SplitPane) verticalDimensions() (int, int, int) {
	dividerHeight := 0
	if s.showDivider {
		dividerHeight = 1
	}

	available := s.height - dividerHeight
	if available < 0 {
		available = 0
	}

	topHeight := int(float64(available) * s.ratio)
	if topHeight < 0 {
		topHeight = 0
	}
	if topHeight > available {
		topHeight = available
	}
	bottomHeight := available - topHeight
	return topHeight, bottomHeight, dividerHeight
}

func (s *SplitPane) dividerVisual() string {
	ch := s.dividerChar
	if ch == "" {
		ch = defaultDividerChar(s.direction)
	}
	if !s.resizing {
		return ch
	}

	accent := ""
	if s.designTokens != nil {
		accent = style.ANSIColorFromHex(s.designTokens.Accent)
	}
	if accent != "" {
		return accent + ch + style.ANSIReset
	}
	return style.ANSIInverse + ch + style.ANSIReset
}

func defaultDividerChar(direction SplitDirection) string {
	if direction == SplitVertical {
		return "─"
	}
	return "│"
}

func clipOrPad(text string, width int) string {
	if width <= 0 {
		return ""
	}

	if utf8.RuneCountInString(text) > width {
		runes := []rune(text)
		return string(runes[:width])
	}

	padding := width - utf8.RuneCountInString(text)
	if padding > 0 {
		return text + strings.Repeat(" ", padding)
	}
	return text
}

func splitPaneClamp(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}
