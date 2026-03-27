package tui

import (
	"fmt"
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tea "github.com/charmbracelet/bubbletea"
)

// ToastLevel is the severity level for a toast notification.
type ToastLevel int

const (
	ToastInfo ToastLevel = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// ToastItem represents a single toast notification item.
type ToastItem struct {
	ID        string
	Message   string
	Level     ToastLevel
	CreatedAt time.Time
	Duration  time.Duration
	Dismissed bool
}

// Toast displays stacked notifications with auto-dismiss behavior.
type Toast struct {
	width           int
	focused         bool
	items           []ToastItem
	iconSet         IconSet
	maxVisible      int
	defaultDuration time.Duration
	nextID          uint64

	infoColor    string
	successColor string
	warningColor string
	errorColor   string
	textColor    string
}

// ToastOption configures a Toast.
type ToastOption func(*Toast)

// WithToastDesignTokens applies design-system colors to toast rendering.
func WithToastDesignTokens(tokens *design.DesignTokens) ToastOption {
	return func(t *Toast) {
		t.applyDesignTokens(tokens)
	}
}

// WithToastTheme applies a named design-system theme.
func WithToastTheme(theme string) ToastOption {
	return func(t *Toast) {
		t.applyDesignTokens(designTokensForTheme(theme))
	}
}

// WithToastIconSet sets the icon set used by toasts.
func WithToastIconSet(iconSet IconSet) ToastOption {
	return func(t *Toast) {
		t.iconSet = iconSet
	}
}

// WithMaxVisible sets how many toasts are rendered at once.
func WithMaxVisible(n int) ToastOption {
	return func(t *Toast) {
		if n > 0 {
			t.maxVisible = n
		}
	}
}

// WithDefaultDuration sets the default auto-dismiss duration.
func WithDefaultDuration(d time.Duration) ToastOption {
	return func(t *Toast) {
		if d > 0 {
			t.defaultDuration = d
		}
	}
}

// toastTickMsg is sent periodically to auto-dismiss expired toasts.
type toastTickMsg time.Time

// NewToast creates a toast component.
func NewToast(opts ...ToastOption) *Toast {
	t := &Toast{
		items:           make([]ToastItem, 0),
		iconSet:         IconSetSymbols,
		maxVisible:      5,
		defaultDuration: 4 * time.Second,
		infoColor:       ansiCyan,
		successColor:    ansiGreen,
		warningColor:    ansiYellow,
		errorColor:      ansiRed,
		textColor:       ansiDim,
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Init initializes the toast component.
func (t *Toast) Init() tea.Cmd {
	return t.tick()
}

// Update handles toast messages.
func (t *Toast) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
	case toastTickMsg:
		t.dismissExpired(time.Time(msg))
		return t, t.tick()
	}

	return t, nil
}

// View renders stacked toasts (most recent first).
func (t *Toast) View() string {
	if len(t.items) == 0 {
		return ""
	}

	visible := t.items
	if t.maxVisible > 0 && len(visible) > t.maxVisible {
		visible = visible[:t.maxVisible]
	}

	lines := make([]string, 0, len(visible))
	for _, item := range visible {
		color := t.colorForLevel(item.Level)
		icon := t.iconForLevel(item.Level)
		line := fmt.Sprintf("%s%s%s %s%s%s", color, icon, ansiReset, t.textColor, item.Message, ansiReset)

		if t.width > 0 && len(stripANSI(line)) > t.width {
			if t.width <= 3 {
				line = strings.Repeat(".", t.width)
			} else {
				line = truncateANSI(line, t.width-3) + "..."
			}
		}

		if t.focused {
			line = ansiInverse + line + ansiReset
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n") + "\n"
}

// Focus marks this component as focused.
func (t *Toast) Focus() {
	t.focused = true
}

// Blur marks this component as unfocused.
func (t *Toast) Blur() {
	t.focused = false
}

// Focused returns whether this component is focused.
func (t *Toast) Focused() bool {
	return t.focused
}

// Push adds a toast with the default duration and returns its ID.
func (t *Toast) Push(message string, level ToastLevel) string {
	return t.PushWithDuration(message, level, t.defaultDuration)
}

// PushWithDuration adds a toast with a custom duration and returns its ID.
func (t *Toast) PushWithDuration(message string, level ToastLevel, d time.Duration) string {
	if d <= 0 {
		d = t.defaultDuration
	}

	t.nextID++
	id := fmt.Sprintf("toast-%d-%d", time.Now().UnixNano(), t.nextID)

	item := ToastItem{
		ID:        id,
		Message:   message,
		Level:     level,
		CreatedAt: time.Now(),
		Duration:  d,
	}

	// Most recent first.
	t.items = append([]ToastItem{item}, t.items...)
	return id
}

// Dismiss dismisses a toast by ID.
func (t *Toast) Dismiss(id string) {
	for i := range t.items {
		if t.items[i].ID == id {
			t.items[i].Dismissed = true
			break
		}
	}
	t.pruneDismissed()
}

// DismissAll dismisses all toasts.
func (t *Toast) DismissAll() {
	for i := range t.items {
		t.items[i].Dismissed = true
	}
	t.pruneDismissed()
}

// Count returns the number of active toasts.
func (t *Toast) Count() int {
	return len(t.items)
}

func (t *Toast) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(ts time.Time) tea.Msg {
		return toastTickMsg(ts)
	})
}

func (t *Toast) dismissExpired(now time.Time) {
	for i := range t.items {
		if t.items[i].Dismissed {
			continue
		}
		if t.items[i].Duration > 0 && now.Sub(t.items[i].CreatedAt) >= t.items[i].Duration {
			t.items[i].Dismissed = true
		}
	}
	t.pruneDismissed()
}

func (t *Toast) pruneDismissed() {
	filtered := t.items[:0]
	for _, item := range t.items {
		if !item.Dismissed {
			filtered = append(filtered, item)
		}
	}
	t.items = filtered
}

func (t *Toast) colorForLevel(level ToastLevel) string {
	switch level {
	case ToastSuccess:
		return t.successColor
	case ToastWarning:
		return t.warningColor
	case ToastError:
		return t.errorColor
	default:
		return t.infoColor
	}
}

func (t *Toast) iconForLevel(level ToastLevel) string {
	switch level {
	case ToastSuccess:
		return t.iconSet.Success
	case ToastWarning:
		return t.iconSet.Warning
	case ToastError:
		return t.iconSet.Error
	default:
		return t.iconSet.Info
	}
}

func (t *Toast) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}

	if foreground := ansiColorFromHex(tokens.Color); foreground != "" {
		t.textColor = foreground
	}
	if accent := ansiColorFromHex(tokens.Accent); accent != "" {
		t.infoColor = accent
		t.successColor = accent
		t.warningColor = accent
		t.errorColor = accent
	}
}

var _ Component = (*Toast)(nil)
