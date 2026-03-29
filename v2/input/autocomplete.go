package input

import (
	"strings"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Suggestion is one autocomplete option returned by a SuggestFunc.
type Suggestion struct {
	Text        string
	Description string
	Value       string
}

// AutocompleteMsg is emitted when a suggestion (or submitted input) is accepted.
type AutocompleteMsg struct {
	Suggestion Suggestion
}

// SuggestFunc returns suggestions for the current input value.
type SuggestFunc func(input string) []Suggestion

// AutocompleteOption configures an Autocomplete component.
type AutocompleteOption func(*Autocomplete)

// WithAutocompletePlaceholder sets the input placeholder text.
func WithAutocompletePlaceholder(placeholder string) AutocompleteOption {
	return func(a *Autocomplete) {
		a.textinput.Placeholder = placeholder
	}
}

// WithAutocompleteWidth sets the preferred render width in terminal cells.
func WithAutocompleteWidth(width int) AutocompleteOption {
	return func(a *Autocomplete) {
		if width > 0 {
			a.width = width
			a.textinput.Width = width
		}
	}
}

// WithAutocompleteMaxSuggestions sets the maximum suggestions shown in the dropdown.
func WithAutocompleteMaxSuggestions(max int) AutocompleteOption {
	return func(a *Autocomplete) {
		if max > 0 {
			a.maxSuggestions = max
		}
	}
}

// Autocomplete is a text input with keyboard-driven suggestion navigation.
type Autocomplete struct {
	textinput        textinput.Model
	suggestFn        SuggestFunc
	suggestions      []Suggestion
	cursor           int
	showSuggestions  bool
	maxSuggestions   int
	focused          bool
	width            int
}

// NewAutocomplete creates a new Autocomplete component.
func NewAutocomplete(suggestFn SuggestFunc, opts ...AutocompleteOption) *Autocomplete {
	ti := textinput.New()
	ti.Placeholder = "Type to search..."

	a := &Autocomplete{
		textinput:      ti,
		suggestFn:      suggestFn,
		suggestions:    nil,
		cursor:         0,
		showSuggestions: false,
		maxSuggestions: 5,
		focused:        false,
		width:          0,
	}

	for _, opt := range opts {
		opt(a)
	}

	a.refreshSuggestions()
	return a
}

// Init initializes the component.
func (a *Autocomplete) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles keyboard and window messages.
func (a *Autocomplete) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if a.width > 0 {
			a.textinput.Width = a.width
		} else {
			w := msg.Width
			if w < 1 {
				w = 1
			}
			a.textinput.Width = w
		}
		return a, nil

	case tea.KeyMsg:
		if !a.focused {
			return a, nil
		}

		switch msg.Type {
		case tea.KeyEsc:
			a.showSuggestions = false
			return a, nil

		case tea.KeyTab, tea.KeyDown:
			if len(a.suggestions) == 0 {
				return a, nil
			}
			if !a.showSuggestions {
				a.showSuggestions = true
				a.cursor = 0
				return a, nil
			}
			if msg.Type == tea.KeyTab {
				a.cursor = (a.cursor + 1) % len(a.suggestions)
				return a, nil
			}
			if a.cursor < len(a.suggestions)-1 {
				a.cursor++
			}
			return a, nil

		case tea.KeyUp:
			if !a.showSuggestions || len(a.suggestions) == 0 {
				return a, nil
			}
			if a.cursor > 0 {
				a.cursor--
			}
			return a, nil

		case tea.KeyEnter:
			if a.showSuggestions && len(a.suggestions) > 0 && a.cursor >= 0 && a.cursor < len(a.suggestions) {
				selected := a.suggestions[a.cursor]
				a.textinput.SetValue(selected.Text)
				a.showSuggestions = false
				a.refreshSuggestions()
				return a, func() tea.Msg {
					return AutocompleteMsg{Suggestion: selected}
				}
			}

			submitted := strings.TrimSpace(a.textinput.Value())
			if submitted == "" {
				return a, nil
			}
			return a, func() tea.Msg {
				return AutocompleteMsg{Suggestion: Suggestion{Text: submitted, Value: submitted}}
			}
		}
	}

	if !a.focused {
		return a, nil
	}

	previousValue := a.textinput.Value()
	a.textinput, cmd = a.textinput.Update(msg)
	if a.textinput.Value() != previousValue {
		a.refreshSuggestions()
	}

	return a, cmd
}

// View renders the input and, when visible, the suggestion dropdown below it.
func (a *Autocomplete) View() string {
	inputView := a.textinput.View()
	if !a.showSuggestions || len(a.suggestions) == 0 {
		return inputView
	}

	dropdownWidth := a.renderWidth()
	itemStyle := lipgloss.NewStyle()
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#61AFEF"))
	descriptionStyle := lipgloss.NewStyle().Faint(true)

	lines := make([]string, 0, len(a.suggestions))
	for i, suggestion := range a.suggestions {
		line := suggestion.Text
		if strings.TrimSpace(suggestion.Description) != "" {
			line += " " + descriptionStyle.Render("— "+suggestion.Description)
		}

		if i == a.cursor {
			line = selectedStyle.Render("▸ " + line)
		} else {
			line = itemStyle.Render("  " + line)
		}
		lines = append(lines, line)
	}

	dropdownStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 1)
	if dropdownWidth > 0 {
		dropdownStyle = dropdownStyle.Width(dropdownWidth)
	}

	return inputView + "\n" + dropdownStyle.Render(strings.Join(lines, "\n"))
}

// Focus marks the component as focused.
func (a *Autocomplete) Focus() {
	a.focused = true
	a.textinput.Focus()
	a.refreshSuggestions()
}

// Blur marks the component as unfocused.
func (a *Autocomplete) Blur() {
	a.focused = false
	a.showSuggestions = false
	a.textinput.Blur()
}

// Focused reports whether the component is focused.
func (a *Autocomplete) Focused() bool {
	return a.focused
}

// Value returns the current input value.
func (a *Autocomplete) Value() string {
	return a.textinput.Value()
}

// SetValue sets the input value and refreshes suggestions.
func (a *Autocomplete) SetValue(s string) {
	a.textinput.SetValue(s)
	a.refreshSuggestions()
}

func (a *Autocomplete) refreshSuggestions() {
	if a.suggestFn == nil {
		a.suggestions = nil
		a.cursor = 0
		a.showSuggestions = false
		return
	}

	suggestions := a.suggestFn(a.textinput.Value())
	if a.maxSuggestions > 0 && len(suggestions) > a.maxSuggestions {
		suggestions = suggestions[:a.maxSuggestions]
	}

	a.suggestions = append([]Suggestion(nil), suggestions...)
	if len(a.suggestions) == 0 {
		a.cursor = 0
		a.showSuggestions = false
		return
	}

	if a.cursor < 0 {
		a.cursor = 0
	}
	if a.cursor >= len(a.suggestions) {
		a.cursor = len(a.suggestions) - 1
	}
	if a.focused {
		a.showSuggestions = true
	}
}

func (a *Autocomplete) renderWidth() int {
	if a.width > 0 {
		return a.width
	}
	if a.textinput.Width > 0 {
		return a.textinput.Width
	}
	return 0
}

var _ tui.Component = (*Autocomplete)(nil)
