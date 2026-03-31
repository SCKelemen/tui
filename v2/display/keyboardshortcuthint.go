package display

import (
	"strings"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// ShortcutEntry is a keyboard shortcut key + label pair.
type ShortcutEntry struct {
	Key   string
	Label string
}

// KeyboardShortcutHint displays keyboard shortcuts as badges.
type KeyboardShortcutHint struct {
	shortcuts    []ShortcutEntry
	width        int
	separator    string
	focused      bool
	designTokens *design.DesignTokens
}

// KeyboardShortcutHintOption configures a KeyboardShortcutHint.
type KeyboardShortcutHintOption func(*KeyboardShortcutHint)

// WithKeyboardShortcutHintWidth sets the render width.
func WithKeyboardShortcutHintWidth(width int) KeyboardShortcutHintOption {
	return func(k *KeyboardShortcutHint) {
		if width >= 0 {
			k.width = width
		}
	}
}

// WithKeyboardShortcutHintSeparator sets the shortcut separator.
func WithKeyboardShortcutHintSeparator(separator string) KeyboardShortcutHintOption {
	return func(k *KeyboardShortcutHint) {
		if separator != "" {
			k.separator = separator
		}
	}
}

// WithKeyboardShortcutHintDesignTokens applies design tokens.
func WithKeyboardShortcutHintDesignTokens(tokens *design.DesignTokens) KeyboardShortcutHintOption {
	return func(k *KeyboardShortcutHint) {
		if tokens != nil {
			k.designTokens = tokens
		}
	}
}

// NewKeyboardShortcutHint creates a keyboard shortcut hint component.
func NewKeyboardShortcutHint(shortcuts []ShortcutEntry, opts ...KeyboardShortcutHintOption) *KeyboardShortcutHint {
	k := &KeyboardShortcutHint{
		shortcuts:    append([]ShortcutEntry(nil), shortcuts...),
		separator:    "  ",
		designTokens: design.DefaultTheme(),
	}

	for _, opt := range opts {
		opt(k)
	}

	return k
}

// Init initializes the component.
func (k *KeyboardShortcutHint) Init() tea.Cmd { return nil }

// Update handles Bubble Tea messages.
func (k *KeyboardShortcutHint) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		if k.width == 0 {
			k.width = m.Width
		}
	}
	return k, nil
}

// View renders the keyboard shortcut badges.
func (k *KeyboardShortcutHint) View() string {
	if len(k.shortcuts) == 0 {
		return ""
	}

	parts := make([]string, 0, len(k.shortcuts))
	for _, entry := range k.shortcuts {
		parts = append(parts, k.renderEntry(entry))
	}

	line := strings.Join(parts, k.separator)
	if k.width > 0 && style.StringWidth(stripANSI(line)) > k.width {
		plain := style.Truncate(stripANSI(line), k.width, "…")
		return plain
	}

	return line
}

// Focus marks the component as focused.
func (k *KeyboardShortcutHint) Focus() { k.focused = true }

// Blur marks the component as unfocused.
func (k *KeyboardShortcutHint) Blur() { k.focused = false }

// Focused reports whether the component is focused.
func (k *KeyboardShortcutHint) Focused() bool { return k.focused }

func (k *KeyboardShortcutHint) renderEntry(entry ShortcutEntry) string {
	key := strings.TrimSpace(entry.Key)
	label := strings.TrimSpace(entry.Label)

	badge := "[" + key + "]"
	if key == "" {
		badge = "[]"
	}

	keyColor := style.ANSIWhite
	labelColor := style.ANSIDim
	if k.designTokens != nil {
		if c := style.Fg(k.designTokens.Accent); c != "" {
			keyColor = c
		}
		if c := style.Fg(k.designTokens.MutedColor); c != "" {
			labelColor = c
		}
	}

	if label == "" {
		return keyColor + style.ANSIBold + badge + style.ANSIReset
	}

	return keyColor + style.ANSIBold + badge + style.ANSIReset + " " + labelColor + label + style.ANSIReset
}

var _ tui.Component = (*KeyboardShortcutHint)(nil)
