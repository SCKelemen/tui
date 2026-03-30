package input

import (
	"fmt"
	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
	"sort"
	"strings"
)

// PaletteCommand is a command shown in the floating command palette.
type PaletteCommand struct {
	Name        string // command name e.g. 'diff'
	Description string // description text
	Shortcut    string // optional keyboard shortcut hint
}

// PaletteSelectMsg is emitted when a command selection is confirmed.
type PaletteSelectMsg struct {
	Command PaletteCommand
}

// PaletteDismissMsg is emitted when the palette is dismissed.
type PaletteDismissMsg struct{}

// FloatingPaletteOption configures a FloatingPalette.
type FloatingPaletteOption func(*FloatingPalette)

// WithFloatingPaletteMaxResults sets the maximum number of visible results.
func WithFloatingPaletteMaxResults(n int) FloatingPaletteOption {
	return func(fp *FloatingPalette) {
		if n > 0 {
			fp.maxResults = n
		}
	}
}

// WithFloatingPaletteMargin sets the horizontal margin on each side.
func WithFloatingPaletteMargin(margin int) FloatingPaletteOption {
	return func(fp *FloatingPalette) {
		if margin >= 0 {
			fp.horizontalMargin = margin
		}
	}
}

// WithFloatingPalettePrompt sets the input prompt.
func WithFloatingPalettePrompt(prompt string) FloatingPaletteOption {
	return func(fp *FloatingPalette) {
		if prompt != "" {
			fp.prompt = prompt
		}
	}
}

// WithFloatingPaletteDesignTokens applies design-system colors.
func WithFloatingPaletteDesignTokens(tokens *design.DesignTokens) FloatingPaletteOption {
	return func(fp *FloatingPalette) {
		fp.applyDesignTokens(tokens)
	}
}

// WithFloatingPaletteTheme applies a named design-system theme.
func WithFloatingPaletteTheme(theme string) FloatingPaletteOption {
	return func(fp *FloatingPalette) {
		fp.applyDesignTokens(style.DesignTokensForTheme(theme))
	}
}

// FloatingPalette is a centered floating command palette overlay.
type FloatingPalette struct {
	commands         []PaletteCommand
	filtered         []PaletteCommand
	matchResults     map[int]FuzzyMatch
	fuzzy            *FuzzyMatcher
	input            string
	cursorPos        int
	selectedIndex    int
	visible          bool
	width            int
	height           int
	maxResults       int
	horizontalMargin int
	prompt           string
	focused          bool
	designTokens     *design.DesignTokens

	promptColor      string
	descriptionColor string
	selectedColor    string
}

// NewFloatingPalette creates a floating command palette component.
func NewFloatingPalette(commands []PaletteCommand, opts ...FloatingPaletteOption) *FloatingPalette {
	fp := &FloatingPalette{
		commands:         append([]PaletteCommand(nil), commands...),
		matchResults:     make(map[int]FuzzyMatch),
		fuzzy:            NewFuzzyMatcher(),
		maxResults:       10,
		horizontalMargin: 8,
		prompt:           "> ",
		promptColor:      style.ANSIReset,
		descriptionColor: style.ANSIDim,
		selectedColor:    style.ANSIInverse,
	}
	for _, opt := range opts {
		opt(fp)
	}

	fp.refilter()
	return fp
}

// Init initializes the component.
func (fp *FloatingPalette) Init() tea.Cmd {
	return nil
}

// Update handles input and window size changes.
func (fp *FloatingPalette) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		fp.width = msg.Width
		fp.height = msg.Height
		return fp, nil
	case tea.KeyMsg:
		if !fp.visible || !fp.focused {
			return fp, nil
		}

		switch msg.Type {
		case tea.KeyEsc:
			fp.Hide()
			return fp, func() tea.Msg { return PaletteDismissMsg{} }
		case tea.KeyEnter:
			selected := fp.GetSelectedCommand()
			fp.Hide()
			if selected == nil {
				return fp, nil
			}
			command := *selected
			return fp, func() tea.Msg { return PaletteSelectMsg{Command: command} }
		case tea.KeyUp, tea.KeyCtrlP:
			if fp.selectedIndex > 0 {
				fp.selectedIndex--
			}
			return fp, nil
		case tea.KeyDown, tea.KeyCtrlN:
			if fp.selectedIndex < len(fp.filtered)-1 {
				fp.selectedIndex++
			}
			return fp, nil
		case tea.KeyBackspace:
			if fp.cursorPos > 0 && len(fp.input) > 0 {
				fp.input = fp.input[:fp.cursorPos-1] + fp.input[fp.cursorPos:]
				fp.cursorPos--
				fp.refilter()
				fp.selectedIndex = 0
			}
			return fp, nil
		case tea.KeyDelete:
			if fp.cursorPos < len(fp.input) {
				fp.input = fp.input[:fp.cursorPos] + fp.input[fp.cursorPos+1:]
				fp.refilter()
				fp.selectedIndex = 0
			}
			return fp, nil
		case tea.KeyLeft:
			if fp.cursorPos > 0 {
				fp.cursorPos--
			}
			return fp, nil
		case tea.KeyRight:
			if fp.cursorPos < len(fp.input) {
				fp.cursorPos++
			}
			return fp, nil
		case tea.KeyHome:
			fp.cursorPos = 0
			return fp, nil
		case tea.KeyEnd:
			fp.cursorPos = len(fp.input)
			return fp, nil
		case tea.KeyRunes:
			if len(msg.Runes) == 0 {
				return fp, nil
			}
			insert := string(msg.Runes)
			fp.input = fp.input[:fp.cursorPos] + insert + fp.input[fp.cursorPos:]
			fp.cursorPos += len(insert)
			fp.refilter()
			fp.selectedIndex = 0
			return fp, nil
		}
	}

	return fp, nil
}

// View renders the floating palette overlay.
func (fp *FloatingPalette) View() string {
	if !fp.visible {
		return ""
	}

	contentWidth := fp.width - (2 * fp.horizontalMargin)
	if contentWidth < 1 {
		contentWidth = 1
	}

	topBorder := strings.Repeat("▄", contentWidth)
	bottomBorder := strings.Repeat("▀", contentWidth)

	inputLine := fp.renderInputLine(contentWidth)
	results := fp.renderResults(contentWidth)

	contentLines := make([]string, 0, 3+len(results))
	contentLines = append(contentLines, topBorder, inputLine, bottomBorder)
	contentLines = append(contentLines, results...)

	leftPad := strings.Repeat(" ", fp.horizontalMargin)
	for i := range contentLines {
		contentLines[i] = leftPad + contentLines[i]
	}

	if fp.height <= 0 {
		return strings.Join(contentLines, "\n")
	}

	totalContentLines := len(contentLines)
	topPadLines := (fp.height - totalContentLines) / 2
	if topPadLines < 0 {
		topPadLines = 0
	}
	bottomPadLines := fp.height - topPadLines - totalContentLines
	if bottomPadLines < 0 {
		bottomPadLines = 0
	}

	allLines := make([]string, 0, topPadLines+totalContentLines+bottomPadLines)
	for i := 0; i < topPadLines; i++ {
		allLines = append(allLines, "")
	}
	allLines = append(allLines, contentLines...)
	for i := 0; i < bottomPadLines; i++ {
		allLines = append(allLines, "")
	}

	return strings.Join(allLines, "\n")
}

// Focus marks the palette as focused.
func (fp *FloatingPalette) Focus() {
	fp.focused = true
}

// Blur marks the palette as unfocused.
func (fp *FloatingPalette) Blur() {
	fp.focused = false
}

// Focused returns whether the palette has focus.
func (fp *FloatingPalette) Focused() bool {
	return fp.focused
}

// Show makes the palette visible and resets filter state.
func (fp *FloatingPalette) Show() {
	fp.visible = true
	fp.input = ""
	fp.cursorPos = 0
	fp.refilter()
	fp.selectedIndex = 0
}

// Hide conceals the palette.
func (fp *FloatingPalette) Hide() {
	fp.visible = false
}

// IsVisible returns whether the palette is visible.
func (fp *FloatingPalette) IsVisible() bool {
	return fp.visible
}

// Toggle flips visibility.
func (fp *FloatingPalette) Toggle() {
	if fp.visible {
		fp.Hide()
		return
	}
	fp.Show()
}

// SetCommands updates the command list and reapplies filtering.
func (fp *FloatingPalette) SetCommands(cmds []PaletteCommand) {
	fp.commands = append([]PaletteCommand(nil), cmds...)
	fp.refilter()
	if fp.selectedIndex >= len(fp.filtered) {
		fp.selectedIndex = maxFloatingPalette(0, len(fp.filtered)-1)
	}
}

// GetSelectedCommand returns the currently selected filtered command.
func (fp *FloatingPalette) GetSelectedCommand() *PaletteCommand {
	if len(fp.filtered) == 0 || fp.selectedIndex < 0 || fp.selectedIndex >= len(fp.filtered) {
		return nil
	}
	return &fp.filtered[fp.selectedIndex]
}

func (fp *FloatingPalette) refilter() {
	if fp.fuzzy == nil {
		fp.fuzzy = NewFuzzyMatcher()
	}

	query := strings.TrimSpace(fp.input)
	fp.matchResults = make(map[int]FuzzyMatch)
	if query == "" {
		fp.filtered = append([]PaletteCommand(nil), fp.commands...)
		if fp.selectedIndex >= len(fp.filtered) {
			fp.selectedIndex = maxFloatingPalette(0, len(fp.filtered)-1)
		}
		return
	}

	type scoredCommand struct {
		command PaletteCommand
		match   FuzzyMatch
	}

	scored := make([]scoredCommand, 0, len(fp.commands))
	for _, cmd := range fp.commands {
		match := fp.fuzzy.Match(query, cmd.Name)
		if !match.Matched {
			continue
		}
		scored = append(scored, scoredCommand{command: cmd, match: match})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].match.Score != scored[j].match.Score {
			return scored[i].match.Score > scored[j].match.Score
		}
		return strings.ToLower(scored[i].command.Name) < strings.ToLower(scored[j].command.Name)
	})

	fp.filtered = make([]PaletteCommand, 0, len(scored))
	for i, candidate := range scored {
		fp.filtered = append(fp.filtered, candidate.command)
		fp.matchResults[i] = candidate.match
	}

	if fp.selectedIndex >= len(fp.filtered) {
		fp.selectedIndex = maxFloatingPalette(0, len(fp.filtered)-1)
	}
}
func (fp *FloatingPalette) renderInputLine(contentWidth int) string {
	if contentWidth <= 0 {
		return ""
	}

	prefix := fp.promptColor + fp.prompt + style.ANSIReset
	before := fp.input[:fp.cursorPos]
	after := fp.input[fp.cursorPos:]
	cursor := " "
	if len(after) > 0 {
		cursor = after[:1]
		after = after[1:]
	}
	cursor = style.ANSIInverse + cursor + style.ANSIReset

	line := prefix + before + cursor + after
	lineWidth := style.StringWidth(stripANSIFloatingPalette(line))
	if lineWidth > contentWidth {
		return style.Truncate(line, contentWidth, "…")
	}
	return style.Pad(line, contentWidth)
}
func (fp *FloatingPalette) renderResults(contentWidth int) []string {
	limit := minFloatingPalette(len(fp.filtered), fp.maxResults)
	if limit <= 0 {
		return nil
	}

	highlightColor := fp.matchHighlightColor()
	queryActive := strings.TrimSpace(fp.input) != ""

	lines := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		cmd := fp.filtered[i]
		name := cmd.Name
		if queryActive {
			if match, ok := fp.matchResults[i]; ok {
				name = HighlightMatch(match, highlightColor)
			}
		}

		line := fmt.Sprintf(" %-10s %s%s%s", name, fp.descriptionColor, cmd.Description, style.ANSIReset)
		if cmd.Shortcut != "" {
			line += fmt.Sprintf(" %s%s%s", style.ANSIDim, cmd.Shortcut, style.ANSIReset)
		}

		lineWidth := style.StringWidth(stripANSIFloatingPalette(line))
		if lineWidth > contentWidth {
			line = style.Truncate(line, contentWidth, "…")
		} else {
			line = style.Pad(line, contentWidth)
		}
		if i == fp.selectedIndex {
			line = fp.selectedColor + style.ANSIBold + line + style.ANSIReset
		}
		lines = append(lines, line)
	}

	return lines
}
func (fp *FloatingPalette) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}
	fp.designTokens = tokens
	if accent := style.ANSIColorFromHex(tokens.Accent); accent != "" {
		fp.promptColor = accent
	}
	if fg := style.ANSIColorFromHex(tokens.Color); fg != "" {
		fp.descriptionColor = fg
	}
}

func (fp *FloatingPalette) matchHighlightColor() string {
	if fp.designTokens != nil {
		if accent := strings.TrimSpace(fp.designTokens.Accent); accent != "" {
			return accent
		}
	}
	return "#4B73FF"
}

func stripANSIFloatingPalette(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
		} else if inEscape {
			if r == 'm' {
				inEscape = false
			}
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func truncateANSIFloatingPalette(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	var result strings.Builder
	visualWidth := 0
	inEscape := false

	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inEscape = true
			result.WriteByte(s[i])
		} else if inEscape {
			result.WriteByte(s[i])
			if s[i] == 'm' {
				inEscape = false
			}
		} else {
			if visualWidth >= maxWidth {
				break
			}
			result.WriteByte(s[i])
			visualWidth++
		}
	}

	return result.String()
}

func minFloatingPalette(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxFloatingPalette(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var _ tui.Component = (*FloatingPalette)(nil)
