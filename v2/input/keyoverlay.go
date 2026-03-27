package input

import (
	"strings"

	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// KeyBinding represents a keyboard shortcut and its purpose.
type KeyBinding struct {
	Key         string
	Description string
	Category    string
}

// KeyOverlayOption configures a KeyOverlay.
type KeyOverlayOption func(*KeyOverlay)

// WithKeyOverlayTitle sets the overlay title.
func WithKeyOverlayTitle(title string) KeyOverlayOption {
	return func(ko *KeyOverlay) {
		if strings.TrimSpace(title) != "" {
			ko.title = title
		}
	}
}

// WithKeyOverlayWidth sets the preferred overlay width.
func WithKeyOverlayWidth(width int) KeyOverlayOption {
	return func(ko *KeyOverlay) {
		if width > 0 {
			ko.overlayWidth = width
		}
	}
}

// WithKeyOverlayBindings sets initial key bindings.
func WithKeyOverlayBindings(bindings []KeyBinding) KeyOverlayOption {
	return func(ko *KeyOverlay) {
		ko.bindings = append([]KeyBinding(nil), bindings...)
	}
}

// KeyOverlay is a centered help overlay that lists available shortcuts.
type KeyOverlay struct {
	title        string
	overlayWidth int
	bindings     []KeyBinding

	visible bool
	focused bool
	width   int
	height  int
}

// NewKeyOverlay creates a new KeyOverlay.
func NewKeyOverlay(opts ...KeyOverlayOption) *KeyOverlay {
	ko := &KeyOverlay{
		title:        "Keyboard Shortcuts",
		overlayWidth: 60,
		bindings:     make([]KeyBinding, 0),
	}

	for _, opt := range opts {
		opt(ko)
	}

	return ko
}

// Init initializes the overlay.
func (ko *KeyOverlay) Init() tea.Cmd {
	return nil
}

// Update handles key and window events.
func (ko *KeyOverlay) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ko.width = msg.Width
		ko.height = msg.Height
		return ko, nil
	case tea.KeyMsg:
		if !ko.focused {
			return ko, nil
		}

		if msg.Type == tea.KeyEsc {
			ko.Toggle()
			return ko, nil
		}

		if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == '?' {
			ko.Toggle()
			return ko, nil
		}
	}

	return ko, nil
}

// View renders the overlay if visible.
func (ko *KeyOverlay) View() string {
	if !ko.visible {
		return ""
	}

	grouped := ko.groupedBindings()

	contentLines := make([]string, 0, len(ko.bindings)+len(grouped)+3)
	contentLines = append(contentLines, style.ANSIBold+ko.title+style.ANSIReset, "")

	for i, group := range grouped {
		if i > 0 {
			contentLines = append(contentLines, "")
		}
		contentLines = append(contentLines, style.ANSIBold+group.category+style.ANSIReset)
		for _, binding := range group.bindings {
			line := "  " + style.ANSICyan + binding.Key + style.ANSIReset + "  " + style.ANSIDim + binding.Description + style.ANSIReset
			contentLines = append(contentLines, line)
		}
	}

	if len(grouped) == 0 {
		contentLines = append(contentLines, style.ANSIDim+"No shortcuts available"+style.ANSIReset)
	}

	boxWidth := ko.overlayWidth
	if ko.width > 0 {
		maxWidth := ko.width - 4
		if maxWidth > 0 && boxWidth > maxWidth {
			boxWidth = maxWidth
		}
	}
	if boxWidth < 24 {
		boxWidth = 24
	}

	innerWidth := boxWidth - 2
	if innerWidth < 1 {
		return ""
	}

	rendered := make([]string, 0, len(contentLines)+2)
	rendered = append(rendered, "╭"+strings.Repeat("─", innerWidth)+"╮")
	for _, line := range contentLines {
		trimmed := style.Truncate(line, innerWidth, "")
		rendered = append(rendered, "│"+style.Pad(trimmed, innerWidth)+"│")
	}
	rendered = append(rendered, "╰"+strings.Repeat("─", innerWidth)+"╯")
	leftPad := 0
	if ko.width > boxWidth {
		leftPad = (ko.width - boxWidth) / 2
	}
	if leftPad > 0 {
		pad := strings.Repeat(" ", leftPad)
		for i := range rendered {
			rendered[i] = pad + rendered[i]
		}
	}

	if ko.height <= 0 {
		return strings.Join(rendered, "\n")
	}

	topPadCount := (ko.height - len(rendered)) / 2
	if topPadCount < 0 {
		topPadCount = 0
	}
	bottomPadCount := ko.height - topPadCount - len(rendered)
	if bottomPadCount < 0 {
		bottomPadCount = 0
	}

	allLines := make([]string, 0, ko.height)
	for i := 0; i < topPadCount; i++ {
		allLines = append(allLines, "")
	}
	allLines = append(allLines, rendered...)
	for i := 0; i < bottomPadCount; i++ {
		allLines = append(allLines, "")
	}

	return strings.Join(allLines, "\n")
}

// Focus marks the overlay as focused.
func (ko *KeyOverlay) Focus() {
	ko.focused = true
}

// Blur marks the overlay as unfocused.
func (ko *KeyOverlay) Blur() {
	ko.focused = false
}

// Focused reports whether the overlay is focused.
func (ko *KeyOverlay) Focused() bool {
	return ko.focused
}

// AddBinding appends a single key binding.
func (ko *KeyOverlay) AddBinding(key, desc, category string) {
	ko.bindings = append(ko.bindings, KeyBinding{Key: key, Description: desc, Category: category})
}

// AddBindings appends multiple key bindings.
func (ko *KeyOverlay) AddBindings(bindings []KeyBinding) {
	ko.bindings = append(ko.bindings, bindings...)
}

// SetVisible sets whether the overlay is visible.
func (ko *KeyOverlay) SetVisible(visible bool) {
	ko.visible = visible
}

// Visible reports whether the overlay is visible.
func (ko *KeyOverlay) Visible() bool {
	return ko.visible
}

// Toggle flips the current visibility.
func (ko *KeyOverlay) Toggle() {
	ko.visible = !ko.visible
}

type keyOverlayGroup struct {
	category string
	bindings []KeyBinding
}

func (ko *KeyOverlay) groupedBindings() []keyOverlayGroup {
	groups := make([]keyOverlayGroup, 0)
	indexByCategory := make(map[string]int)

	for _, binding := range ko.bindings {
		category := strings.TrimSpace(binding.Category)
		if category == "" {
			category = "General"
		}

		idx, found := indexByCategory[category]
		if !found {
			idx = len(groups)
			indexByCategory[category] = idx
			groups = append(groups, keyOverlayGroup{category: category, bindings: make([]KeyBinding, 0, 4)})
		}

		groups[idx].bindings = append(groups[idx].bindings, binding)
	}

	return groups
}

func stripANSIKeyOverlay(s string) string {
	var b strings.Builder
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		b.WriteRune(r)
	}

	return b.String()
}

func truncateANSIKeyOverlay(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	var b strings.Builder
	width := 0
	inEscape := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '\x1b' {
			inEscape = true
			b.WriteByte(ch)
			continue
		}
		if inEscape {
			b.WriteByte(ch)
			if ch == 'm' {
				inEscape = false
			}
			continue
		}
		if width >= maxWidth {
			break
		}
		b.WriteByte(ch)
		width++
	}

	if inEscape {
		b.WriteString(style.ANSIReset)
	}

	return b.String()
}

var _ tui.Component = (*KeyOverlay)(nil)
