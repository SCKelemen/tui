package tui

import (
	"strings"
	"unicode/utf8"

	design "github.com/SCKelemen/design-system"
	tea "github.com/charmbracelet/bubbletea"
)

// BreadcrumbItem represents one segment in a breadcrumb path.
type BreadcrumbItem struct {
	ID    string
	Label string
}

// Breadcrumb displays a navigable path with keyboard interaction.
type Breadcrumb struct {
	items            []BreadcrumbItem
	selected         int
	separator        string
	focused          bool
	width            int
	onSelect         func(id string) tea.Cmd
	designTokens     *design.DesignTokens
	truncateFromLeft bool
}

// BreadcrumbOption configures a Breadcrumb.
type BreadcrumbOption func(*Breadcrumb)

// WithBreadcrumbDesignTokens applies design tokens to the breadcrumb.
func WithBreadcrumbDesignTokens(tokens *design.DesignTokens) BreadcrumbOption {
	return func(b *Breadcrumb) {
		if tokens != nil {
			b.designTokens = tokens
		}
	}
}

// WithBreadcrumbTheme applies a named design-system theme.
func WithBreadcrumbTheme(theme string) BreadcrumbOption {
	return func(b *Breadcrumb) {
		b.designTokens = designTokensForTheme(theme)
	}
}

// WithSeparator sets the separator used between breadcrumb segments.
func WithSeparator(sep string) BreadcrumbOption {
	return func(b *Breadcrumb) {
		if sep != "" {
			b.separator = sep
		}
	}
}

// WithBreadcrumbItems sets initial breadcrumb items.
func WithBreadcrumbItems(items ...BreadcrumbItem) BreadcrumbOption {
	return func(b *Breadcrumb) {
		b.SetItems(items)
	}
}

// WithOnBreadcrumbSelect sets the selection callback fired on Enter.
func WithOnBreadcrumbSelect(fn func(string) tea.Cmd) BreadcrumbOption {
	return func(b *Breadcrumb) {
		b.onSelect = fn
	}
}

// WithTruncateFromLeft configures overflow truncation direction.
func WithTruncateFromLeft(truncateFromLeft bool) BreadcrumbOption {
	return func(b *Breadcrumb) {
		b.truncateFromLeft = truncateFromLeft
	}
}

// NewBreadcrumb creates a new breadcrumb component.
func NewBreadcrumb(opts ...BreadcrumbOption) *Breadcrumb {
	b := &Breadcrumb{
		items:            []BreadcrumbItem{},
		selected:         -1,
		separator:        " › ",
		designTokens:     design.DefaultTheme(),
		truncateFromLeft: true,
	}

	for _, opt := range opts {
		opt(b)
	}

	if len(b.items) > 0 && (b.selected < 0 || b.selected >= len(b.items)) {
		b.selected = len(b.items) - 1
	}

	return b
}

// Init initializes the breadcrumb.
func (b *Breadcrumb) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (b *Breadcrumb) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = msg.Width
		return b, nil
	case tea.KeyMsg:
		if !b.focused || len(b.items) == 0 {
			return b, nil
		}

		switch msg.String() {
		case "h", "left":
			if b.selected > 0 {
				b.selected--
			}
		case "l", "right":
			if b.selected < len(b.items)-1 {
				b.selected++
			}
		case "enter":
			if b.onSelect != nil {
				selected := b.GetSelected()
				if selected != nil {
					return b, b.onSelect(selected.ID)
				}
			}
		}
	}

	return b, nil
}

// View renders the breadcrumb.
func (b *Breadcrumb) View() string {
	if len(b.items) == 0 {
		return ""
	}

	start, end, prefixEllipsis, suffixEllipsis := b.visibleRange()

	accent := ansiBold
	if b.designTokens != nil {
		if c := ansiColorFromHex(b.designTokens.Accent); c != "" {
			accent = c + ansiBold
		}
	}
	dim := ansiDim
	if b.designTokens != nil {
		if c := ansiColorFromHex(b.designTokens.Color); c != "" {
			dim = c + ansiDim
		}
	}
	underline := ansiUnderline
	reset := ansiReset

	parts := make([]string, 0, (end-start)+2)

	if prefixEllipsis {
		parts = append(parts, dim+"..."+reset)
	}

	for i := start; i < end; i++ {
		style := dim
		if i == len(b.items)-1 {
			style = accent
		}
		if b.focused && i == b.selected {
			style += underline
		}
		parts = append(parts, style+b.items[i].Label+reset)
	}

	if suffixEllipsis {
		parts = append(parts, dim+"..."+reset)
	}

	return strings.Join(parts, b.separator)
}

// Focus is called when this component receives focus.
func (b *Breadcrumb) Focus() {
	b.focused = true
}

// Blur is called when this component loses focus.
func (b *Breadcrumb) Blur() {
	b.focused = false
}

// Focused returns whether this component is currently focused.
func (b *Breadcrumb) Focused() bool {
	return b.focused
}

// SetItems replaces all breadcrumb items.
func (b *Breadcrumb) SetItems(items []BreadcrumbItem) {
	b.items = append([]BreadcrumbItem(nil), items...)
	if len(b.items) == 0 {
		b.selected = -1
		return
	}

	if b.selected < 0 || b.selected >= len(b.items) {
		b.selected = len(b.items) - 1
	}
}

// Push appends an item to the breadcrumb and selects it.
func (b *Breadcrumb) Push(item BreadcrumbItem) {
	b.items = append(b.items, item)
	b.selected = len(b.items) - 1
}

// Pop removes and returns the last item.
func (b *Breadcrumb) Pop() *BreadcrumbItem {
	if len(b.items) == 0 {
		return nil
	}

	removed := b.items[len(b.items)-1]
	b.items = b.items[:len(b.items)-1]

	if len(b.items) == 0 {
		b.selected = -1
	} else if b.selected >= len(b.items) {
		b.selected = len(b.items) - 1
	}

	return &removed
}

// GetSelected returns the currently selected item.
func (b *Breadcrumb) GetSelected() *BreadcrumbItem {
	if b.selected < 0 || b.selected >= len(b.items) {
		return nil
	}
	return &b.items[b.selected]
}

// SetSelected selects an item by ID.
func (b *Breadcrumb) SetSelected(id string) {
	for i := range b.items {
		if b.items[i].ID == id {
			b.selected = i
			return
		}
	}
}

// ItemCount returns the number of breadcrumb items.
func (b *Breadcrumb) ItemCount() int {
	return len(b.items)
}

func (b *Breadcrumb) visibleRange() (start int, end int, prefixEllipsis bool, suffixEllipsis bool) {
	if len(b.items) == 0 {
		return 0, 0, false, false
	}

	if b.width <= 0 {
		return 0, len(b.items), false, false
	}

	if b.truncateFromLeft {
		for s := 0; s < len(b.items); s++ {
			candidate := b.labelsJoined(b.items[s:])
			if s > 0 {
				candidate = "..." + b.separator + candidate
			}
			if utf8.RuneCountInString(candidate) <= b.width {
				return s, len(b.items), s > 0, false
			}
		}
		return len(b.items) - 1, len(b.items), len(b.items) > 1, false
	}

	for e := len(b.items); e > 0; e-- {
		candidate := b.labelsJoined(b.items[:e])
		if e < len(b.items) {
			candidate = candidate + b.separator + "..."
		}
		if len(candidate) <= b.width {
			return 0, e, false, e < len(b.items)
		}
	}

	return 0, 1, false, len(b.items) > 1
}

func (b *Breadcrumb) labelsJoined(items []BreadcrumbItem) string {
	labels := make([]string, 0, len(items))
	for _, item := range items {
		labels = append(labels, item.Label)
	}
	return strings.Join(labels, b.separator)
}
