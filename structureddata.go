package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ItemType represents the type of data item
type ItemType int

const (
	ItemKeyValue ItemType = iota
	ItemHeader
	ItemSeparator
	ItemValue // Value only, no key
)

// DataStatus represents the status of the data display
type DataStatus int

const (
	DataStatusNone DataStatus = iota
	DataStatusRunning
	DataStatusSuccess
	DataStatusError
	DataStatusInfo
)

// structuredDataTickMsg is sent periodically for animation
type structuredDataTickMsg time.Time

// DataItem represents a single item in structured data
type DataItem struct {
	Type   ItemType
	Key    string
	Value  string
	Indent int    // Indentation level (0 = no indent, 1 = one level, etc.)
	Color  string // Optional ANSI color code (e.g., "\033[32m" for green)
}

// StructuredData displays formatted key-value data with tree connectors
type StructuredData struct {
	width          int
	title          string
	items          []DataItem
	focused        bool
	expanded       bool
	maxLines       int        // Max lines when collapsed (0 = show all)
	icon           string
	keyWidth       int        // Width for key column (auto-calculated if 0)
	status         DataStatus // Current status (Running, Success, Error, Info)
	animationFrame int        // Frame counter for blinking animation
}

// NewStructuredData creates a new structured data component
func NewStructuredData(title string, opts ...StructuredDataOption) *StructuredData {
	sd := &StructuredData{
		title:    title,
		items:    []DataItem{},
		expanded: true, // Default to expanded
		icon:     "⏺",
		keyWidth: 0, // Auto-calculate
	}

	for _, opt := range opts {
		opt(sd)
	}

	return sd
}

// StructuredDataOption configures a StructuredData
type StructuredDataOption func(*StructuredData)

// WithMaxLines sets the maximum lines to show when collapsed
func WithStructuredDataMaxLines(n int) StructuredDataOption {
	return func(sd *StructuredData) {
		sd.maxLines = n
	}
}

// WithKeyWidth sets a fixed width for the key column
func WithKeyWidth(width int) StructuredDataOption {
	return func(sd *StructuredData) {
		sd.keyWidth = width
	}
}

// WithIcon sets a custom icon
func WithStructuredDataIcon(icon string) StructuredDataOption {
	return func(sd *StructuredData) {
		sd.icon = icon
	}
}

// Builder methods for ergonomic API

// AddRow adds a key-value row
func (sd *StructuredData) AddRow(key, value string) *StructuredData {
	sd.items = append(sd.items, DataItem{
		Type:  ItemKeyValue,
		Key:   key,
		Value: value,
	})
	return sd
}

// AddColoredRow adds a key-value row with custom color
func (sd *StructuredData) AddColoredRow(key, value, color string) *StructuredData {
	sd.items = append(sd.items, DataItem{
		Type:  ItemKeyValue,
		Key:   key,
		Value: value,
		Color: color,
	})
	return sd
}

// AddIndentedRow adds an indented key-value row
func (sd *StructuredData) AddIndentedRow(key, value string, indent int) *StructuredData {
	sd.items = append(sd.items, DataItem{
		Type:   ItemKeyValue,
		Key:    key,
		Value:  value,
		Indent: indent,
	})
	return sd
}

// AddHeader adds a section header
func (sd *StructuredData) AddHeader(text string) *StructuredData {
	sd.items = append(sd.items, DataItem{
		Type:  ItemHeader,
		Value: text,
	})
	return sd
}

// AddSeparator adds a blank line
func (sd *StructuredData) AddSeparator() *StructuredData {
	sd.items = append(sd.items, DataItem{
		Type: ItemSeparator,
	})
	return sd
}

// AddValue adds a value-only line (no key)
func (sd *StructuredData) AddValue(value string) *StructuredData {
	sd.items = append(sd.items, DataItem{
		Type:  ItemValue,
		Value: value,
	})
	return sd
}

// AddIndentedValue adds an indented value-only line
func (sd *StructuredData) AddIndentedValue(value string, indent int) *StructuredData {
	sd.items = append(sd.items, DataItem{
		Type:   ItemValue,
		Value:  value,
		Indent: indent,
	})
	return sd
}

// SetItems replaces all items (for batch operations)
func (sd *StructuredData) SetItems(items []DataItem) *StructuredData {
	sd.items = items
	return sd
}

// Clear removes all items
func (sd *StructuredData) Clear() *StructuredData {
	sd.items = []DataItem{}
	return sd
}

// Component interface implementation

// Init initializes the structured data component
func (sd *StructuredData) Init() tea.Cmd {
	if sd.status == DataStatusRunning {
		return sd.tick()
	}
	return nil
}

// tick returns a command that sends a tick message after a delay
func (sd *StructuredData) tick() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return structuredDataTickMsg(t)
	})
}

// Update handles messages
func (sd *StructuredData) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		sd.width = msg.Width

	case tea.KeyMsg:
		if sd.focused {
			switch msg.String() {
			case "ctrl+o", "enter":
				sd.ToggleExpanded()
			}
		}

	case structuredDataTickMsg:
		if sd.status == DataStatusRunning {
			sd.animationFrame++
			return sd, sd.tick()
		}
	}
	return sd, nil
}

// View renders the structured data
func (sd *StructuredData) View() string {
	if sd.width == 0 {
		return ""
	}

	var lines []string

	// Header with icon and title
	icon := sd.renderIcon()
	var header string
	if sd.title != "" {
		header = fmt.Sprintf("%s \033[1m%s\033[0m", icon, sd.title)
	} else {
		header = fmt.Sprintf("%s \033[1mData\033[0m", icon)
	}

	if sd.focused {
		header = "\033[7m" + header + "\033[0m" // Inverted when focused
	}

	lines = append(lines, header)

	// Calculate key width if not set
	keyWidth := sd.keyWidth
	if keyWidth == 0 {
		keyWidth = sd.calculateKeyWidth()
	}

	// Render items
	itemsToRender := sd.items
	hiddenCount := 0

	if !sd.expanded && sd.maxLines > 0 && len(sd.items) > sd.maxLines {
		itemsToRender = sd.items[:sd.maxLines]
		hiddenCount = len(sd.items) - sd.maxLines
	}

	if len(itemsToRender) == 0 {
		lines = append(lines, "  \033[2m⎿  (no data)\033[0m")
		return strings.Join(lines, "\n") + "\n"
	}

	for i, item := range itemsToRender {
		line := sd.renderItem(item, keyWidth, i == 0)
		lines = append(lines, line)
	}

	// Show "... +N items" if collapsed
	if hiddenCount > 0 {
		expandHint := fmt.Sprintf("     \033[2m… +%d items \033[0m\033[3m(ctrl+o to expand)\033[0m",
			hiddenCount)
		lines = append(lines, expandHint)
	}

	return strings.Join(lines, "\n") + "\n"
}

// Focus is called when this component receives focus
func (sd *StructuredData) Focus() {
	sd.focused = true
}

// Blur is called when this component loses focus
func (sd *StructuredData) Blur() {
	sd.focused = false
}

// Focused returns whether this component is currently focused
func (sd *StructuredData) Focused() bool {
	return sd.focused
}

// ToggleExpanded toggles the expanded state
func (sd *StructuredData) ToggleExpanded() {
	sd.expanded = !sd.expanded
}

// SetExpanded sets the expanded state
func (sd *StructuredData) SetExpanded(expanded bool) {
	sd.expanded = expanded
}

// Status management methods

// SetStatus sets the status and starts/stops animation
func (sd *StructuredData) SetStatus(status DataStatus) tea.Cmd {
	sd.status = status
	sd.animationFrame = 0
	if status == DataStatusRunning {
		return sd.tick()
	}
	return nil
}

// StartRunning sets status to running and begins animation
func (sd *StructuredData) StartRunning() tea.Cmd {
	return sd.SetStatus(DataStatusRunning)
}

// MarkSuccess sets status to success (green icon, no animation)
func (sd *StructuredData) MarkSuccess() {
	sd.status = DataStatusSuccess
}

// MarkError sets status to error (red icon, no animation)
func (sd *StructuredData) MarkError() {
	sd.status = DataStatusError
}

// MarkInfo sets status to info (white icon, no animation)
func (sd *StructuredData) MarkInfo() {
	sd.status = DataStatusInfo
}

// GetStatus returns the current status
func (sd *StructuredData) GetStatus() DataStatus {
	return sd.status
}

// Helper methods

// calculateKeyWidth finds the longest key for alignment
func (sd *StructuredData) calculateKeyWidth() int {
	maxWidth := 20 // Minimum width
	for _, item := range sd.items {
		if item.Type == ItemKeyValue && item.Key != "" {
			keyLen := len(item.Key) + (item.Indent * 2)
			if keyLen > maxWidth {
				maxWidth = keyLen
			}
		}
	}
	// Cap at reasonable maximum
	if maxWidth > 40 {
		maxWidth = 40
	}
	return maxWidth
}

// renderItem renders a single data item
func (sd *StructuredData) renderItem(item DataItem, keyWidth int, isFirst bool) string {
	var prefix string
	if isFirst {
		prefix = "  \033[2m⎿\033[0m  "
	} else {
		prefix = "     " // Indent for continuation lines
	}

	// Add indentation
	indent := strings.Repeat("  ", item.Indent)

	// Apply color if specified
	colorStart := ""
	colorEnd := ""
	if item.Color != "" {
		colorStart = item.Color
		colorEnd = "\033[0m"
	}

	switch item.Type {
	case ItemKeyValue:
		if item.Key == "" {
			// Value only, but in KeyValue format
			return fmt.Sprintf("%s%s%s%s%s", prefix, indent, colorStart, item.Value, colorEnd)
		}
		// Key-value pair with alignment
		key := fmt.Sprintf("%-*s", keyWidth-(item.Indent*2), item.Key+":")
		return fmt.Sprintf("%s%s%s%s %s%s", prefix, indent, colorStart, key, item.Value, colorEnd)

	case ItemHeader:
		// Section header (bold, no key)
		return fmt.Sprintf("%s%s\033[1m%s\033[0m", prefix, indent, item.Value)

	case ItemSeparator:
		// Blank line
		return prefix

	case ItemValue:
		// Value-only line
		return fmt.Sprintf("%s%s%s%s%s", prefix, indent, colorStart, item.Value, colorEnd)

	default:
		return prefix + item.Value
	}
}

// renderIcon renders the status icon with animation
func (sd *StructuredData) renderIcon() string {
	switch sd.status {
	case DataStatusRunning:
		// Blink: alternate between visible and invisible
		if sd.animationFrame%2 == 0 {
			return "\033[36m" + sd.icon + "\033[0m" // Cyan (visible)
		}
		return " " // Invisible (blank space same width as icon)

	case DataStatusSuccess:
		return "\033[32m" + sd.icon + "\033[0m" // Green

	case DataStatusError:
		return "\033[31m" + sd.icon + "\033[0m" // Red

	case DataStatusInfo:
		return "\033[37m" + sd.icon + "\033[0m" // White

	default: // DataStatusNone
		return "\033[36m" + sd.icon + "\033[0m" // Default cyan
	}
}

// Utility functions for common patterns

// FromMap creates structured data from a map
func FromMap(title string, data map[string]string) *StructuredData {
	sd := NewStructuredData(title)
	for key, value := range data {
		sd.AddRow(key, value)
	}
	return sd
}

// FromKeyValuePairs creates structured data from alternating key-value strings
func FromKeyValuePairs(title string, pairs ...string) *StructuredData {
	sd := NewStructuredData(title)
	for i := 0; i < len(pairs)-1; i += 2 {
		sd.AddRow(pairs[i], pairs[i+1])
	}
	return sd
}
