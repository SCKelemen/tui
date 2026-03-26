package tui

import (
	"fmt"
	"strings"

	design "github.com/SCKelemen/design-system"
	tea "github.com/charmbracelet/bubbletea"
)

// ChecklistItemStatus represents one checklist item's execution state.
type ChecklistItemStatus int

const (
	ItemPending ChecklistItemStatus = iota
	ItemInProgress
	ItemDone
	ItemSkipped
	ItemFailed
)

// ChecklistItem is one checklist entry and optional nested children.
type ChecklistItem struct {
	ID       string
	Label    string
	Status   ChecklistItemStatus
	Children []ChecklistItem
	Expanded bool
}

// Checklist displays a navigable list of checklist items.
type Checklist struct {
	title        string
	items        []ChecklistItem
	selected     int
	focused      bool
	designTokens *design.DesignTokens
	iconSet      IconSet
	showProgress bool
	width        int

	accentColor string
}

// ChecklistOption configures a Checklist.
type ChecklistOption func(*Checklist)

// WithChecklistDesignTokens applies design-system colors.
func WithChecklistDesignTokens(tokens *design.DesignTokens) ChecklistOption {
	return func(c *Checklist) {
		c.applyDesignTokens(tokens)
	}
}

// WithChecklistTheme applies a named design-system theme.
func WithChecklistTheme(theme string) ChecklistOption {
	return func(c *Checklist) {
		c.applyDesignTokens(designTokensForTheme(theme))
	}
}

// WithChecklistTitle sets the checklist title.
func WithChecklistTitle(title string) ChecklistOption {
	return func(c *Checklist) {
		c.title = title
	}
}

// WithChecklistItems sets the initial checklist items.
func WithChecklistItems(items ...ChecklistItem) ChecklistOption {
	return func(c *Checklist) {
		c.items = append([]ChecklistItem(nil), items...)
		c.clampSelection()
	}
}

// WithChecklistIconSet sets status icons.
func WithChecklistIconSet(iconSet IconSet) ChecklistOption {
	return func(c *Checklist) {
		c.iconSet = iconSet
	}
}

// WithShowProgress toggles progress display in the title row.
func WithShowProgress(show bool) ChecklistOption {
	return func(c *Checklist) {
		c.showProgress = show
	}
}

// NewChecklist creates a new checklist component.
func NewChecklist(opts ...ChecklistOption) *Checklist {
	c := &Checklist{
		title: "Checklist",
		items: []ChecklistItem{},
		selected: 0,
		designTokens: design.DefaultTheme(),
		iconSet: IconSet{
			Running: SpinnerDots.GetFrame(0),
			Success: IconSetSymbols.Success,
			Error:   IconSetSymbols.Error,
			Warning: "↷",
			Info:    IconSetSymbols.Info,
			None:    "○",
		},
		showProgress: true,
		accentColor:  ansiCyan,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.clampSelection()
	return c
}

// Init initializes the checklist component.
func (c *Checklist) Init() tea.Cmd {
	return nil
}

// Update handles Bubble Tea messages.
func (c *Checklist) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width

	case tea.KeyMsg:
		if !c.focused {
			break
		}

		visible := c.visibleItems()
		if len(visible) == 0 {
			break
		}
		if c.selected >= len(visible) {
			c.selected = len(visible) - 1
		}
		switch msg.String() {
		case "j", "down":
			if c.selected < len(visible)-1 {
				c.selected++
			}
		case "k", "up":
			if c.selected > 0 {
				c.selected--
			}
		case " ", "x":
			item := visible[c.selected].item
			c.ToggleItem(item.ID)
		case "enter":
			item := visible[c.selected].item
			if len(item.Children) > 0 {
				item.Expanded = !item.Expanded
			}
		case "a":
			c.setAllStatuses(ItemDone)
		case "u":
			c.setAllStatuses(ItemPending)
		}
	}

	return c, nil
}

// View renders the checklist.
func (c *Checklist) View() string {
	if c.width == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, c.renderHeader())

	visible := c.visibleItems()
	if len(visible) == 0 {
		lines = append(lines, "  "+ansiDim+"└ (no items)"+ansiReset)
		return strings.Join(lines, "\n") + "\n"
	}

	for i, entry := range visible {
		line := c.renderItemLine(entry)
		if c.focused && i == c.selected {
			line = ansiInverse + line + ansiReset
		}
		lines = append(lines, c.fitToWidth(line))
	}

	return strings.Join(lines, "\n") + "\n"
}

// Focus marks the component as focused.
func (c *Checklist) Focus() {
	c.focused = true
}

// Blur marks the component as unfocused.
func (c *Checklist) Blur() {
	c.focused = false
}

// Focused returns whether the component has focus.
func (c *Checklist) Focused() bool {
	return c.focused
}

// AddItem appends an item to the checklist.
func (c *Checklist) AddItem(item ChecklistItem) {
	c.items = append(c.items, item)
	c.clampSelection()
}

// SetItemStatus sets a checklist item's status by ID.
func (c *Checklist) SetItemStatus(id string, status ChecklistItemStatus) {
	item := c.GetItem(id)
	if item == nil {
		return
	}
	item.Status = status
}

// ToggleItem toggles a checklist item between done and pending.
func (c *Checklist) ToggleItem(id string) {
	item := c.GetItem(id)
	if item == nil {
		return
	}

	if item.Status == ItemDone {
		item.Status = ItemPending
		return
	}
	item.Status = ItemDone
}

// GetItem returns a pointer to the item with a matching ID.
func (c *Checklist) GetItem(id string) *ChecklistItem {
	for i := range c.items {
		if item := findChecklistItemByID(&c.items[i], id); item != nil {
			return item
		}
	}
	return nil
}

// RemoveItem removes a checklist item by ID.
func (c *Checklist) RemoveItem(id string) {
	c.items = removeChecklistItemByID(c.items, id)
	c.clampSelection()
}

// SetExpanded sets one item's expanded state by ID.
func (c *Checklist) SetExpanded(id string, expanded bool) {
	item := c.GetItem(id)
	if item == nil {
		return
	}
	item.Expanded = expanded
}

// CompletionCount returns completed and total item counts.
//
// Completed includes all non-pending statuses.
func (c *Checklist) CompletionCount() (done int, total int) {
	for i := range c.items {
		d, t := checklistCompletionCount(&c.items[i])
		done += d
		total += t
	}
	return done, total
}

// ItemCount returns total number of items including nested children.
func (c *Checklist) ItemCount() int {
	_, total := c.CompletionCount()
	return total
}

type checklistVisibleEntry struct {
	item   *ChecklistItem
	branch []bool // per depth: true means "is last"
}

func (c *Checklist) visibleItems() []checklistVisibleEntry {
	entries := make([]checklistVisibleEntry, 0)
	for i := range c.items {
		isLast := i == len(c.items)-1
		c.collectVisibleItems(&entries, &c.items[i], []bool{isLast})
	}
	return entries
}

func (c *Checklist) collectVisibleItems(entries *[]checklistVisibleEntry, item *ChecklistItem, branch []bool) {
	entry := checklistVisibleEntry{item: item, branch: append([]bool(nil), branch...)}
	*entries = append(*entries, entry)

	if !item.Expanded || len(item.Children) == 0 {
		return
	}

	for i := range item.Children {
		isLast := i == len(item.Children)-1
		nextBranch := append(append([]bool(nil), branch...), isLast)
		c.collectVisibleItems(entries, &item.Children[i], nextBranch)
	}
}

func (c *Checklist) renderHeader() string {
	title := c.title
	if strings.TrimSpace(title) == "" {
		title = "Checklist"
	}

	left := fmt.Sprintf("◼ %s", title)
	if !c.showProgress {
		return c.fitToWidth(left)
	}

	done, total := c.CompletionCount()
	progress := c.progressSummary(done, total)

	leftLen := len(stripANSI(left))
	progressLen := len(stripANSI(progress))
	spacing := c.width - leftLen - progressLen
	if spacing < 1 {
		return c.fitToWidth(left + " " + progress)
	}

	return left + strings.Repeat(" ", spacing) + progress
}

func (c *Checklist) progressSummary(done, total int) string {
	if total == 0 {
		return "[0/0 ░░░░░░ 0%]"
	}

	pct := int(float64(done) / float64(total) * 100)
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	barWidth := 6
	filled := int((float64(done) / float64(total)) * float64(barWidth))
	if done > 0 && filled == 0 {
		filled = 1
	}
	if filled > barWidth {
		filled = barWidth
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	return fmt.Sprintf("[%d/%d %s %d%%]", done, total, bar, pct)
}

func (c *Checklist) renderItemLine(entry checklistVisibleEntry) string {
	depth := len(entry.branch) - 1

	var prefix strings.Builder
	if depth > 0 {
		for i := 0; i < depth; i++ {
			if entry.branch[i] {
				prefix.WriteString("  ")
			} else {
				prefix.WriteString("│ ")
			}
		}
	}

	if entry.branch[depth] {
		prefix.WriteString("└ ")
	} else {
		prefix.WriteString("├ ")
	}

	statusIcon, statusColor := c.statusStyle(entry.item.Status)
	line := fmt.Sprintf("%s%s%s %s%s", prefix.String(), statusColor, statusIcon, entry.item.Label, ansiReset)
	return line
}

func (c *Checklist) statusStyle(status ChecklistItemStatus) (icon string, color string) {
	switch status {
	case ItemDone:
		icon = c.nonEmpty(c.iconSet.Success, "✓")
		color = ansiGreen
	case ItemInProgress:
		icon = c.nonEmpty(c.iconSet.Running, SpinnerDots.GetFrame(0))
		color = c.accentColor
	case ItemSkipped:
		icon = c.nonEmpty(c.iconSet.Warning, "↷")
		color = ansiYellow
	case ItemFailed:
		icon = c.nonEmpty(c.iconSet.Error, "✗")
		color = ansiRed
	default:
		icon = c.nonEmpty(c.iconSet.None, "○")
		color = ansiDim
	}
	return icon, color
}

func (c *Checklist) setAllStatuses(status ChecklistItemStatus) {
	for i := range c.items {
		setChecklistStatusRecursive(&c.items[i], status)
	}
}

func (c *Checklist) fitToWidth(line string) string {
	if c.width <= 0 {
		return line
	}
	if len(stripANSI(line)) <= c.width {
		return line
	}
	if c.width <= 3 {
		return truncateANSI(line, c.width)
	}
	return truncateANSI(line, c.width-3) + "..."
}

func (c *Checklist) clampSelection() {
	visibleCount := len(c.visibleItems())
	if visibleCount == 0 {
		c.selected = 0
		return
	}
	if c.selected < 0 {
		c.selected = 0
	}
	if c.selected >= visibleCount {
		c.selected = visibleCount - 1
	}
}

func (c *Checklist) applyDesignTokens(tokens *design.DesignTokens) {
	if tokens == nil {
		return
	}
	c.designTokens = tokens

	accent := ansiColorFromHex(tokens.Accent)
	if accent != "" {
		c.accentColor = accent
	}
}

func (c *Checklist) nonEmpty(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func findChecklistItemByID(item *ChecklistItem, id string) *ChecklistItem {
	if item == nil {
		return nil
	}
	if item.ID == id {
		return item
	}
	for i := range item.Children {
		if found := findChecklistItemByID(&item.Children[i], id); found != nil {
			return found
		}
	}
	return nil
}

func removeChecklistItemByID(items []ChecklistItem, id string) []ChecklistItem {
	filtered := items[:0]
	for _, item := range items {
		if item.ID == id {
			continue
		}
		item.Children = removeChecklistItemByID(item.Children, id)
		filtered = append(filtered, item)
	}
	return filtered
}

func checklistCompletionCount(item *ChecklistItem) (done int, total int) {
	if item == nil {
		return 0, 0
	}
	total = 1
	if item.Status != ItemPending {
		done = 1
	}
	for i := range item.Children {
		d, t := checklistCompletionCount(&item.Children[i])
		done += d
		total += t
	}
	return done, total
}

func setChecklistStatusRecursive(item *ChecklistItem, status ChecklistItemStatus) {
	if item == nil {
		return
	}
	item.Status = status
	for i := range item.Children {
		setChecklistStatusRecursive(&item.Children[i], status)
	}
}

var _ Component = (*Checklist)(nil)
