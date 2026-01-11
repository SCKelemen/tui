package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ConfirmationBlock displays file operations with code preview and multiple choice confirmation
type ConfirmationBlock struct {
	width   int
	height  int
	focused bool

	// Operation details
	operation   string   // e.g., "Write", "Read", "Edit"
	filepath    string   // Full file path
	description string   // e.g., "Create file ../yaml-lsp/data/test-issues.yaml"
	code        []string // Code lines to preview

	// Confirmation options
	options       []string // e.g., ["Yes", "Yes, allow all edits...", "No"]
	selectedIndex int      // Currently selected option (0-indexed)
	footerHints   []string // e.g., ["Esc to cancel", "Tab to add instructions"]

	// Display settings
	startLine    int  // Starting line number (default 1)
	showPreview  int  // Number of code lines to show (0 = all)
	confirmed    bool // Whether user has confirmed
	confirmedIdx int  // Which option was selected (-1 = none)
}

// ConfirmationBlockOption configures a ConfirmationBlock
type ConfirmationBlockOption func(*ConfirmationBlock)

// WithConfirmOperation sets the operation type
func WithConfirmOperation(op string) ConfirmationBlockOption {
	return func(cb *ConfirmationBlock) {
		cb.operation = op
	}
}

// WithConfirmFilepath sets the file path
func WithConfirmFilepath(path string) ConfirmationBlockOption {
	return func(cb *ConfirmationBlock) {
		cb.filepath = path
	}
}

// WithConfirmDescription sets the description text
func WithConfirmDescription(desc string) ConfirmationBlockOption {
	return func(cb *ConfirmationBlock) {
		cb.description = desc
	}
}

// WithConfirmCode sets the code content
func WithConfirmCode(code string) ConfirmationBlockOption {
	return func(cb *ConfirmationBlock) {
		cb.code = strings.Split(code, "\n")
	}
}

// WithConfirmCodeLines sets the code content as lines
func WithConfirmCodeLines(lines []string) ConfirmationBlockOption {
	return func(cb *ConfirmationBlock) {
		cb.code = lines
	}
}

// WithConfirmOptions sets the confirmation options
func WithConfirmOptions(options []string) ConfirmationBlockOption {
	return func(cb *ConfirmationBlock) {
		cb.options = options
	}
}

// WithConfirmStartLine sets the starting line number
func WithConfirmStartLine(line int) ConfirmationBlockOption {
	return func(cb *ConfirmationBlock) {
		cb.startLine = line
	}
}

// WithConfirmPreview sets number of preview lines (0 = show all)
func WithConfirmPreview(n int) ConfirmationBlockOption {
	return func(cb *ConfirmationBlock) {
		cb.showPreview = n
	}
}

// WithConfirmFooterHints sets footer hint text
func WithConfirmFooterHints(hints []string) ConfirmationBlockOption {
	return func(cb *ConfirmationBlock) {
		cb.footerHints = hints
	}
}

// NewConfirmationBlock creates a new confirmation block
func NewConfirmationBlock(opts ...ConfirmationBlockOption) *ConfirmationBlock {
	cb := &ConfirmationBlock{
		operation:     "Write",
		startLine:     1,
		selectedIndex: 0,
		confirmedIdx:  -1,
		options: []string{
			"Yes",
			"No",
		},
		footerHints: []string{
			"Esc to cancel",
			"Tab to add additional instructions",
		},
	}

	for _, opt := range opts {
		opt(cb)
	}

	return cb
}

// Init initializes the confirmation block
func (cb *ConfirmationBlock) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (cb *ConfirmationBlock) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cb.width = msg.Width
		cb.height = msg.Height

	case tea.KeyMsg:
		if !cb.focused || cb.confirmed {
			return cb, nil
		}

		switch msg.String() {
		case "up", "k", "shift+tab":
			cb.selectedIndex--
			if cb.selectedIndex < 0 {
				cb.selectedIndex = len(cb.options) - 1
			}
		case "down", "j", "tab":
			cb.selectedIndex++
			if cb.selectedIndex >= len(cb.options) {
				cb.selectedIndex = 0
			}
		case "enter":
			cb.confirmed = true
			cb.confirmedIdx = cb.selectedIndex
			// Could return a custom message here
		case "esc":
			cb.confirmed = true
			cb.confirmedIdx = -1 // Cancelled
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Quick select by number
			idx := int(msg.Runes[0] - '1')
			if idx >= 0 && idx < len(cb.options) {
				cb.selectedIndex = idx
				cb.confirmed = true
				cb.confirmedIdx = idx
			}
		}
	}

	return cb, nil
}

// View renders the confirmation block
func (cb *ConfirmationBlock) View() string {
	if cb.width == 0 {
		return ""
	}

	var b strings.Builder

	// Header: ⏺ Operation(filepath)
	icon := cb.getOperationIcon()
	b.WriteString(fmt.Sprintf("%s \033[1m%s\033[0m", icon, cb.operation))
	if cb.filepath != "" {
		b.WriteString(fmt.Sprintf("(\033[36m%s\033[0m)", cb.filepath))
	}
	b.WriteString("\n\n")

	// Solid separator line (full width)
	width := cb.width
	if width == 0 {
		width = 80
	}
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")

	// Description
	if cb.description != "" {
		b.WriteString(fmt.Sprintf(" %s\n", cb.description))
	}

	// Dashed separator (╌)
	b.WriteString(strings.Repeat("╌", width))
	b.WriteString("\n")

	// Code preview with line numbers
	if len(cb.code) > 0 {
		b.WriteString(cb.renderCode())
	}

	// Dashed separator (╌)
	b.WriteString(strings.Repeat("╌", width))
	b.WriteString("\n")

	// Confirmation prompt with options
	if cb.description != "" {
		// Extract action from description (e.g., "Create file" -> "create")
		b.WriteString(fmt.Sprintf(" Do you want to %s?\n", cb.getActionVerb()))
	}

	// Render options
	for i, opt := range cb.options {
		if i == cb.selectedIndex && cb.focused && !cb.confirmed {
			// Selected option (highlighted with ❯)
			b.WriteString(fmt.Sprintf(" \033[36m❯ %d. %s\033[0m\n", i+1, opt))
		} else {
			// Unselected option
			b.WriteString(fmt.Sprintf("   %d. %s\n", i+1, opt))
		}
	}

	// Footer hints
	if len(cb.footerHints) > 0 && !cb.confirmed {
		b.WriteString("\n \033[2m")
		b.WriteString(strings.Join(cb.footerHints, " · "))
		b.WriteString("\033[0m\n")
	}

	// Confirmation result
	if cb.confirmed {
		b.WriteString("\n")
		if cb.confirmedIdx == -1 {
			b.WriteString(" \033[2mCancelled\033[0m\n")
		} else if cb.confirmedIdx >= 0 && cb.confirmedIdx < len(cb.options) {
			b.WriteString(fmt.Sprintf(" \033[32m✓ Selected: %s\033[0m\n", cb.options[cb.confirmedIdx]))
		}
	}

	return b.String()
}

// Focus is called when this component receives focus
func (cb *ConfirmationBlock) Focus() {
	cb.focused = true
}

// Blur is called when this component loses focus
func (cb *ConfirmationBlock) Blur() {
	cb.focused = false
}

// Focused returns whether this component is currently focused
func (cb *ConfirmationBlock) Focused() bool {
	return cb.focused
}

// IsConfirmed returns whether the user has made a choice
func (cb *ConfirmationBlock) IsConfirmed() bool {
	return cb.confirmed
}

// GetSelection returns the selected option index (-1 if cancelled)
func (cb *ConfirmationBlock) GetSelection() int {
	return cb.confirmedIdx
}

// Reset resets the confirmation state
func (cb *ConfirmationBlock) Reset() {
	cb.confirmed = false
	cb.confirmedIdx = -1
	cb.selectedIndex = 0
}

// getOperationIcon returns an icon for the operation type
func (cb *ConfirmationBlock) getOperationIcon() string {
	switch strings.ToLower(cb.operation) {
	case "write", "create":
		return "\033[32m⏺\033[0m" // Green circle
	case "read", "view":
		return "\033[34m⏺\033[0m" // Blue circle
	case "edit", "update":
		return "\033[33m⏺\033[0m" // Yellow circle
	case "delete", "remove":
		return "\033[31m⏺\033[0m" // Red circle
	default:
		return "\033[37m⏺\033[0m" // White circle
	}
}

// getActionVerb extracts the verb from the description
func (cb *ConfirmationBlock) getActionVerb() string {
	// Try to extract verb from description
	// e.g., "Create file test.yaml" -> "create test.yaml"
	if cb.description == "" {
		return strings.ToLower(cb.operation)
	}

	desc := strings.ToLower(cb.description)
	// Just use the description as-is, lowercased
	return desc
}

// renderCode renders the code preview with line numbers
func (cb *ConfirmationBlock) renderCode() string {
	var b strings.Builder

	linesToShow := len(cb.code)
	if cb.showPreview > 0 && linesToShow > cb.showPreview {
		linesToShow = cb.showPreview
	}

	// Calculate line number width
	maxLineNum := cb.startLine + len(cb.code) - 1
	lineNumWidth := len(fmt.Sprintf("%d", maxLineNum))

	// Render lines
	for i := 0; i < linesToShow; i++ {
		lineNum := cb.startLine + i
		b.WriteString(fmt.Sprintf(" %*d %s\n", lineNumWidth, lineNum, cb.code[i]))
	}

	// Show "... more lines" indicator if truncated
	if cb.showPreview > 0 && len(cb.code) > cb.showPreview {
		remaining := len(cb.code) - cb.showPreview
		b.WriteString(fmt.Sprintf(" \033[2m... +%d more lines\033[0m\n", remaining))
	}

	return b.String()
}
