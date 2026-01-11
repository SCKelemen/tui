package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// DiffLine represents a single line in a diff
type DiffLine struct {
	Type    DiffType // Added, Removed, or Unchanged
	Content string   // Line content
	LineNum int      // Line number (for context)
}

// DiffType indicates the type of diff line
type DiffType int

const (
	DiffUnchanged DiffType = iota // No change (context)
	DiffAdded                      // Line was added (+)
	DiffRemoved                    // Line was removed (-)
)

// DiffBlock displays code changes with +/- indicators
type DiffBlock struct {
	width   int
	height  int
	focused bool

	// Content
	filename    string     // File being modified
	operation   string     // e.g., "Edit", "Refactor"
	summary     string     // Summary of changes
	lines       []DiffLine // Diff lines
	oldStart    int        // Starting line number in old file
	newStart    int        // Starting line number in new file

	// Display state
	expanded     bool // Whether diff is shown or collapsed
	showContext  int  // Number of context lines to show around changes (default 3)
	maxLines     int  // Maximum lines to show when expanded (0 = show all)
}

// DiffBlockOption configures a DiffBlock
type DiffBlockOption func(*DiffBlock)

// WithDiffFilename sets the filename
func WithDiffFilename(name string) DiffBlockOption {
	return func(db *DiffBlock) {
		db.filename = name
	}
}

// WithDiffOperation sets the operation type
func WithDiffOperation(op string) DiffBlockOption {
	return func(db *DiffBlock) {
		db.operation = op
	}
}

// WithDiffSummary sets the summary text
func WithDiffSummary(summary string) DiffBlockOption {
	return func(db *DiffBlock) {
		db.summary = summary
	}
}

// WithDiffLines sets the diff content
func WithDiffLines(lines []DiffLine) DiffBlockOption {
	return func(db *DiffBlock) {
		db.lines = lines
	}
}

// WithDiffExpanded sets whether the block starts expanded
func WithDiffExpanded(expanded bool) DiffBlockOption {
	return func(db *DiffBlock) {
		db.expanded = expanded
	}
}

// WithDiffContext sets number of context lines
func WithDiffContext(n int) DiffBlockOption {
	return func(db *DiffBlock) {
		db.showContext = n
	}
}

// WithDiffMaxLines sets maximum lines to show
func WithDiffMaxLines(max int) DiffBlockOption {
	return func(db *DiffBlock) {
		db.maxLines = max
	}
}

// NewDiffBlock creates a new diff block component
func NewDiffBlock(opts ...DiffBlockOption) *DiffBlock {
	db := &DiffBlock{
		operation:   "Edit",
		showContext: 3,
		expanded:    false,
		oldStart:    1,
		newStart:    1,
	}

	for _, opt := range opts {
		opt(db)
	}

	return db
}

// NewDiffBlockFromStrings creates a diff block from old and new content strings
func NewDiffBlockFromStrings(old, new string, opts ...DiffBlockOption) *DiffBlock {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	// Simple line-by-line diff (can be enhanced with proper diff algorithm)
	diffLines := simpleDiff(oldLines, newLines)

	db := NewDiffBlock(opts...)
	db.lines = diffLines
	return db
}

// simpleDiff creates a simple line-by-line diff
func simpleDiff(oldLines, newLines []string) []DiffLine {
	var result []DiffLine

	// Find common prefix
	commonPrefix := 0
	for commonPrefix < len(oldLines) && commonPrefix < len(newLines) && oldLines[commonPrefix] == newLines[commonPrefix] {
		result = append(result, DiffLine{
			Type:    DiffUnchanged,
			Content: oldLines[commonPrefix],
			LineNum: commonPrefix + 1,
		})
		commonPrefix++
	}

	// Find common suffix
	commonSuffix := 0
	oldRemaining := len(oldLines) - commonPrefix
	newRemaining := len(newLines) - commonPrefix
	for commonSuffix < oldRemaining && commonSuffix < newRemaining &&
		oldLines[len(oldLines)-1-commonSuffix] == newLines[len(newLines)-1-commonSuffix] {
		commonSuffix++
	}

	// Add removed lines
	for i := commonPrefix; i < len(oldLines)-commonSuffix; i++ {
		result = append(result, DiffLine{
			Type:    DiffRemoved,
			Content: oldLines[i],
			LineNum: i + 1,
		})
	}

	// Add added lines
	for i := commonPrefix; i < len(newLines)-commonSuffix; i++ {
		result = append(result, DiffLine{
			Type:    DiffAdded,
			Content: newLines[i],
			LineNum: i + 1,
		})
	}

	// Add common suffix
	for i := 0; i < commonSuffix; i++ {
		idx := len(oldLines) - commonSuffix + i
		result = append(result, DiffLine{
			Type:    DiffUnchanged,
			Content: oldLines[idx],
			LineNum: idx + 1,
		})
	}

	return result
}

// Init initializes the diff block
func (db *DiffBlock) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (db *DiffBlock) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		db.width = msg.Width
		db.height = msg.Height

	case tea.KeyMsg:
		if !db.focused {
			return db, nil
		}

		switch msg.String() {
		case "ctrl+o", "enter", " ":
			db.Toggle()
		}
	}

	return db, nil
}

// View renders the diff block
func (db *DiffBlock) View() string {
	if len(db.lines) == 0 {
		return ""
	}

	var b strings.Builder

	// Header: ⏺ Operation(filename)
	icon := "\033[33m⏺\033[0m" // Yellow for edit operations
	b.WriteString(fmt.Sprintf("%s \033[1m%s\033[0m", icon, db.operation))
	if db.filename != "" {
		b.WriteString(fmt.Sprintf("(\033[36m%s\033[0m)", db.filename))
	}
	b.WriteString("\n")

	// Summary line
	if db.summary != "" {
		b.WriteString(fmt.Sprintf("  \033[2m⎿  %s\033[0m\n", db.summary))
	}

	// Diff stats
	added, removed := db.countChanges()
	b.WriteString(fmt.Sprintf("  \033[2m⎿  \033[32m+%d\033[0m \033[31m-%d\033[0m\n", added, removed))

	// Diff lines
	if db.expanded {
		b.WriteString(db.renderExpanded())
	} else {
		b.WriteString(db.renderCollapsed())
	}

	return b.String()
}

// Focus is called when this component receives focus
func (db *DiffBlock) Focus() {
	db.focused = true
}

// Blur is called when this component loses focus
func (db *DiffBlock) Blur() {
	db.focused = false
}

// Focused returns whether this component is currently focused
func (db *DiffBlock) Focused() bool {
	return db.focused
}

// Toggle expands or collapses the diff block
func (db *DiffBlock) Toggle() {
	db.expanded = !db.expanded
}

// Expand shows the full diff
func (db *DiffBlock) Expand() {
	db.expanded = true
}

// Collapse hides the diff
func (db *DiffBlock) Collapse() {
	db.expanded = false
}

// IsExpanded returns whether the diff is currently expanded
func (db *DiffBlock) IsExpanded() bool {
	return db.expanded
}

// countChanges returns the number of added and removed lines
func (db *DiffBlock) countChanges() (added, removed int) {
	for _, line := range db.lines {
		switch line.Type {
		case DiffAdded:
			added++
		case DiffRemoved:
			removed++
		}
	}
	return
}

// renderCollapsed shows a summary of changes
func (db *DiffBlock) renderCollapsed() string {
	var b strings.Builder

	// Show first few changes as preview
	shownLines := 0
	maxPreview := 8

	for _, line := range db.lines {
		if line.Type == DiffUnchanged {
			continue // Skip unchanged lines in collapsed view
		}

		if shownLines >= maxPreview {
			break
		}

		b.WriteString(db.renderDiffLine(line))
		shownLines++
	}

	// Show expansion hint
	totalChanges, _ := db.countChanges()
	if totalChanges > shownLines {
		remaining := totalChanges - shownLines
		b.WriteString(fmt.Sprintf("     \033[2m… +%d more changes (\033[3mctrl+o to expand\033[0m\033[2m)\033[0m\n", remaining))
	}

	return b.String()
}

// renderExpanded shows the full diff with context
func (db *DiffBlock) renderExpanded() string {
	var b strings.Builder

	linesToShow := len(db.lines)
	if db.maxLines > 0 && linesToShow > db.maxLines {
		linesToShow = db.maxLines
	}

	for i := 0; i < linesToShow; i++ {
		b.WriteString(db.renderDiffLine(db.lines[i]))
	}

	// Show "… more lines" if truncated
	if db.maxLines > 0 && len(db.lines) > db.maxLines {
		remaining := len(db.lines) - db.maxLines
		b.WriteString(fmt.Sprintf("     \033[2m… +%d more lines (truncated)\033[0m\n", remaining))
	}

	return b.String()
}

// renderDiffLine renders a single diff line with appropriate styling
func (db *DiffBlock) renderDiffLine(line DiffLine) string {
	switch line.Type {
	case DiffAdded:
		// Green + prefix
		return fmt.Sprintf("  \033[32m+ %s\033[0m\n", line.Content)
	case DiffRemoved:
		// Red - prefix
		return fmt.Sprintf("  \033[31m- %s\033[0m\n", line.Content)
	case DiffUnchanged:
		// Dimmed, no prefix
		return fmt.Sprintf("  \033[2m  %s\033[0m\n", line.Content)
	default:
		return fmt.Sprintf("    %s\n", line.Content)
	}
}
