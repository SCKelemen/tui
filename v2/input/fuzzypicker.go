package input

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// FuzzyPickerItem is one selectable item in the fuzzy picker.
type FuzzyPickerItem struct {
	ID          string
	Label       string
	Description string
	Category    string
}

// FuzzyPickerSelectMsg is emitted when an item is selected or picker is cancelled.
type FuzzyPickerSelectMsg struct {
	Item      FuzzyPickerItem
	Cancelled bool
}

// FuzzyPickerOption configures a FuzzyPicker.
type FuzzyPickerOption func(*FuzzyPicker)

// WithFuzzyPickerWidth sets preferred render width.
func WithFuzzyPickerWidth(width int) FuzzyPickerOption {
	return func(f *FuzzyPicker) {
		if width > 0 {
			f.width = width
		}
	}
}

// WithFuzzyPickerHeight sets preferred render height.
func WithFuzzyPickerHeight(height int) FuzzyPickerOption {
	return func(f *FuzzyPicker) {
		if height > 0 {
			f.height = height
		}
	}
}

// WithFuzzyPickerPlaceholder sets the search box placeholder.
func WithFuzzyPickerPlaceholder(placeholder string) FuzzyPickerOption {
	return func(f *FuzzyPicker) {
		if strings.TrimSpace(placeholder) != "" {
			f.placeholder = placeholder
			f.input.Placeholder = placeholder
		}
	}
}

// WithFuzzyPickerDesignTokens applies design-system tokens.
func WithFuzzyPickerDesignTokens(tokens *design.DesignTokens) FuzzyPickerOption {
	return func(f *FuzzyPicker) {
		f.designTokens = tokens
		f.applyDesignTokens()
	}
}

type fuzzyPickerResult struct {
	item  FuzzyPickerItem
	match FuzzyMatch
}

// FuzzyPicker is a standalone fuzzy search picker.
type FuzzyPicker struct {
	items       []FuzzyPickerItem
	results     []fuzzyPickerResult
	matcher     *FuzzyMatcher
	cursor      int
	focused     bool
	width       int
	height      int
	placeholder string
	input       textinput.Model

	designTokens  *design.DesignTokens
	highlightHex  string
	selectedColor string
	mutedColor    string
}

// NewFuzzyPicker creates a new fuzzy picker that uses FuzzyMatcher for scoring.
func NewFuzzyPicker(items []FuzzyPickerItem, opts ...FuzzyPickerOption) *FuzzyPicker {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 256

	f := &FuzzyPicker{
		items:         append([]FuzzyPickerItem(nil), items...),
		results:       make([]fuzzyPickerResult, 0),
		matcher:       NewFuzzyMatcher(WithFuzzyNormalize(true)),
		cursor:        0,
		focused:       false,
		width:         72,
		height:        10,
		placeholder:   "Search...",
		input:         ti,
		designTokens:  design.DefaultTheme(),
		highlightHex:  "#61AFEF",
		selectedColor: style.ANSICyan,
		mutedColor:    style.ANSIDim,
	}

	for _, opt := range opts {
		opt(f)
	}

	f.applyDesignTokens()
	f.refreshResults()

	return f
}

// Init initializes the component.
func (f *FuzzyPicker) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles keyboard input and filtering.
func (f *FuzzyPicker) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if f.width <= 0 {
			f.width = msg.Width
		}
		f.input.Width = f.width
		return f, nil

	case tea.KeyMsg:
		if !f.focused {
			return f, nil
		}

		switch msg.String() {
		case "up", "k":
			if f.cursor > 0 {
				f.cursor--
			}
			return f, nil
		case "down", "j":
			if f.cursor < len(f.results)-1 {
				f.cursor++
			}
			return f, nil
		case "enter":
			if len(f.results) == 0 {
				return f, nil
			}
			selected := f.results[f.cursor].item
			return f, func() tea.Msg {
				return FuzzyPickerSelectMsg{Item: selected}
			}
		case "esc":
			return f, func() tea.Msg {
				return FuzzyPickerSelectMsg{Cancelled: true}
			}
		}
	}

	if !f.focused {
		return f, nil
	}

	before := f.input.Value()
	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	if f.input.Value() != before {
		f.refreshResults()
	}

	return f, cmd
}

// View renders the search input and filtered results.
func (f *FuzzyPicker) View() string {
	width := f.width
	if width <= 0 {
		width = 72
	}

	lines := []string{style.Pad(style.Truncate(f.input.View(), width, "…"), width)}

	maxRows := f.height - 1
	if maxRows < 1 {
		maxRows = 1
	}

	if len(f.results) == 0 {
		lines = append(lines, style.Pad(style.Truncate(f.mutedColor+"No matches"+style.ANSIReset, width, "…"), width))
		return strings.Join(lines, "\n")
	}

	for i := 0; i < len(f.results) && i < maxRows; i++ {
		row := f.results[i]
		prefix := "  "
		if i == f.cursor {
			prefix = f.selectedColor + "▸ " + style.ANSIReset
		}

		label := row.item.Label
		if row.match.Matched {
			label = HighlightMatch(FuzzyMatch{
				Matched:   true,
				Positions: row.match.Positions,
				Candidate: row.item.Label,
			}, f.highlightHex)
		}

		meta := ""
		if row.item.Category != "" {
			meta = " [" + row.item.Category + "]"
		}
		line := fmt.Sprintf("%s%s%s", prefix, label, f.mutedColor+meta+style.ANSIReset)
		if strings.TrimSpace(row.item.Description) != "" {
			line += " " + f.mutedColor + "— " + row.item.Description + style.ANSIReset
		}

		lines = append(lines, style.Pad(style.Truncate(line, width, "…"), width))
	}

	return strings.Join(lines, "\n")
}

// Focus marks the component focused.
func (f *FuzzyPicker) Focus() {
	f.focused = true
	f.input.Focus()
}

// Blur marks the component unfocused.
func (f *FuzzyPicker) Blur() {
	f.focused = false
	f.input.Blur()
}

// Focused reports focus state.
func (f *FuzzyPicker) Focused() bool { return f.focused }

func (f *FuzzyPicker) refreshResults() {
	query := strings.TrimSpace(f.input.Value())
	if query == "" {
		f.results = make([]fuzzyPickerResult, 0, len(f.items))
		for _, item := range f.items {
			f.results = append(f.results, fuzzyPickerResult{item: item, match: FuzzyMatch{Matched: true, Candidate: item.Label}})
		}
		f.cursor = 0
		return
	}

	candidates := make([]string, len(f.items))
	for i, item := range f.items {
		candidates[i] = item.Label
	}

	ranked := f.matcher.RankMatches(query, candidates)
	mapped := make([]fuzzyPickerResult, 0, len(ranked))
	for _, match := range ranked {
		for idx, candidate := range candidates {
			if candidate == match.Candidate {
				mapped = append(mapped, fuzzyPickerResult{item: f.items[idx], match: match})
				break
			}
		}
	}

	f.results = mapped
	if len(f.results) == 0 {
		f.cursor = 0
		return
	}
	if f.cursor >= len(f.results) {
		f.cursor = len(f.results) - 1
	}
	if f.cursor < 0 {
		f.cursor = 0
	}
}

func (f *FuzzyPicker) applyDesignTokens() {
	if f.designTokens == nil {
		return
	}
	if strings.TrimSpace(f.designTokens.Accent) != "" {
		f.highlightHex = strings.TrimSpace(f.designTokens.Accent)
	}
	if v := style.Fg(f.designTokens.Accent); v != "" {
		f.selectedColor = v
	}
	if v := style.Fg(f.designTokens.MutedColor); v != "" {
		f.mutedColor = v
	}
}

var _ tui.Component = (*FuzzyPicker)(nil)
