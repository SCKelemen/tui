package input

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CompletionKind identifies the semantic category of a completion item.
type CompletionKind int

const (
	KindFunction CompletionKind = iota
	KindVariable
	KindType
	KindInterface
	KindMethod
	KindProperty
	KindConstant
	KindKeyword
	KindSnippet
	KindField
	KindModule
	KindClass
	KindEnum
	KindValue
	KindFile
	KindFolder
)

// CompletionItem is one IntelliSense completion candidate.
type CompletionItem struct {
	Label         string
	Kind          CompletionKind
	Detail        string
	Documentation string
	InsertText    string
	SortText      string
	FilterText    string
	Deprecated    bool
}

// CompletionKindIcon returns a VS Code-style icon for a completion kind.
func CompletionKindIcon(kind CompletionKind) string {
	switch kind {
	case KindFunction:
		return "ƒ"
	case KindVariable:
		return "v"
	case KindType:
		return "T"
	case KindInterface:
		return "I"
	case KindMethod:
		return "m"
	case KindProperty:
		return "p"
	case KindConstant:
		return "C"
	case KindKeyword:
		return "K"
	case KindSnippet:
		return "S"
	case KindField:
		return "f"
	case KindModule:
		return "M"
	case KindClass:
		return "c"
	case KindEnum:
		return "E"
	case KindValue:
		return "#"
	case KindFile:
		return "🗎"
	case KindFolder:
		return "📁"
	default:
		return "?"
	}
}

// KindColor returns the visual color used for a completion kind.
func KindColor(kind CompletionKind) string {
	switch kind {
	case KindFunction, KindMethod:
		return "#C586C0"
	case KindVariable, KindField, KindProperty:
		return "#9CDCFE"
	case KindType, KindInterface, KindClass, KindEnum:
		return "#4EC9B0"
	case KindConstant, KindValue:
		return "#DCDCAA"
	case KindKeyword:
		return "#569CD6"
	case KindSnippet:
		return "#D7BA7D"
	case KindModule:
		return "#4FC1FF"
	case KindFile, KindFolder:
		return "#CE9178"
	default:
		return "#ABB2BF"
	}
}

// CompletionSelectedMsg is emitted when a completion is accepted.
type CompletionSelectedMsg struct {
	Item CompletionItem
}

// IntelliSenseOption configures IntelliSense.
type IntelliSenseOption func(*IntelliSense)

// WithIntelliSenseWidth sets the popup width.
func WithIntelliSenseWidth(width int) IntelliSenseOption {
	return func(i *IntelliSense) {
		if width > 0 {
			i.width = width
		}
	}
}

// WithIntelliSenseMaxVisible sets the maximum number of rows in the popup list.
func WithIntelliSenseMaxVisible(maxVisible int) IntelliSenseOption {
	return func(i *IntelliSense) {
		if maxVisible > 0 {
			i.maxVisible = maxVisible
		}
	}
}

// WithIntelliSenseAnchor sets the popup anchor in editor coordinates.
func WithIntelliSenseAnchor(line, col int) IntelliSenseOption {
	return func(i *IntelliSense) {
		if line >= 0 {
			i.anchorLine = line
		}
		if col >= 0 {
			i.anchorCol = col
		}
	}
}

// WithIntelliSenseShowDoc controls whether the documentation panel is shown.
func WithIntelliSenseShowDoc(show bool) IntelliSenseOption {
	return func(i *IntelliSense) {
		i.showDoc = show
	}
}

// IntelliSense is a VS Code-style completion popup.
type IntelliSense struct {
	items        []CompletionItem
	filtered     []CompletionItem
	query        string
	cursor       int
	scrollOffset int
	maxVisible   int
	visible      bool
	width        int
	showDoc      bool
	anchorLine   int
	anchorCol    int
}

// NewIntelliSense creates a new IntelliSense popup.
func NewIntelliSense(items []CompletionItem, opts ...IntelliSenseOption) *IntelliSense {
	in := &IntelliSense{
		items:      append([]CompletionItem(nil), items...),
		filtered:   append([]CompletionItem(nil), items...),
		query:      "",
		cursor:     0,
		maxVisible: 10,
		visible:    false,
		width:      48,
		showDoc:    false,
		anchorLine: 0,
		anchorCol:  0,
	}

	for _, opt := range opts {
		opt(in)
	}

	in.applyFilter()
	return in
}

// Init initializes the model.
func (i *IntelliSense) Init() tea.Cmd {
	return nil
}

// Show makes the popup visible.
func (i *IntelliSense) Show() {
	i.visible = true
	i.ensureCursorVisible()
}

// Hide conceals the popup.
func (i *IntelliSense) Hide() {
	i.visible = false
}

// Toggle toggles visibility.
func (i *IntelliSense) Toggle() {
	i.visible = !i.visible
	if i.visible {
		i.ensureCursorVisible()
	}
}

// Visible reports whether the popup is visible.
func (i *IntelliSense) Visible() bool {
	return i.visible
}

// SetItems replaces all completion items.
func (i *IntelliSense) SetItems(items []CompletionItem) {
	i.items = append([]CompletionItem(nil), items...)
	i.applyFilter()
}

// SetQuery updates the filter query.
func (i *IntelliSense) SetQuery(q string) {
	i.query = q
	i.applyFilter()
}

// Update handles keyboard interaction and filtering.
func (i *IntelliSense) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !i.visible {
			return i, nil
		}

		switch msg.Type {
		case tea.KeyEsc:
			i.Hide()
			return i, nil
		case tea.KeyUp:
			if i.cursor > 0 {
				i.cursor--
				i.ensureCursorVisible()
			}
			return i, nil
		case tea.KeyDown:
			if i.cursor < len(i.filtered)-1 {
				i.cursor++
				i.ensureCursorVisible()
			}
			return i, nil
		case tea.KeyEnter, tea.KeyTab:
			if item, ok := i.selectedItem(); ok {
				i.Hide()
				return i, func() tea.Msg {
					return CompletionSelectedMsg{Item: item}
				}
			}
			return i, nil
		case tea.KeyBackspace, tea.KeyDelete:
			if i.query != "" {
				i.query = trimLastRune(i.query)
				i.applyFilter()
			}
			return i, nil
		default:
			r := msg.Runes
			if len(r) > 0 {
				if unicode.IsControl(r[0]) {
					return i, nil
				}
				i.query += string(r)
				i.applyFilter()
			}
			return i, nil
		}
	}

	return i, nil
}

// View renders the floating IntelliSense popup.
func (i *IntelliSense) View() string {
	if !i.visible {
		return ""
	}

	list := i.renderListPanel()
	if i.showDoc {
		doc := i.renderDocumentationPanel()
		list = lipgloss.JoinHorizontal(lipgloss.Top, list, " ", doc)
	}

	if i.anchorLine > 0 {
		list = strings.Repeat("\n", i.anchorLine) + list
	}
	if i.anchorCol > 0 {
		indented := make([]string, 0, strings.Count(list, "\n")+1)
		for _, line := range strings.Split(list, "\n") {
			indented = append(indented, strings.Repeat(" ", i.anchorCol)+line)
		}
		list = strings.Join(indented, "\n")
	}

	return list
}

func (i *IntelliSense) renderListPanel() string {
	borderColor := lipgloss.Color("#3C414B")
	bgColor := lipgloss.Color("#252930")
	titleColor := lipgloss.Color("#7A818A")
	detailStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7A818A"))

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(bgColor).
		Padding(0, 1)
	if i.width > 0 {
		panelStyle = panelStyle.Width(i.width)
	}

	if len(i.filtered) == 0 {
		emptyLine := lipgloss.NewStyle().Foreground(titleColor).Render("No completions")
		return panelStyle.Render(emptyLine)
	}

	start, end := i.visibleRange()
	rows := make([]string, 0, end-start+1)
	for idx := start; idx < end; idx++ {
		item := i.filtered[idx]
		icon := lipgloss.NewStyle().Foreground(lipgloss.Color(KindColor(item.Kind))).Render(CompletionKindIcon(item.Kind))
		label := item.Label
		if label == "" {
			label = item.InsertText
		}
		line := fmt.Sprintf("%s %s", icon, label)
		if strings.TrimSpace(item.Detail) != "" {
			line += " " + detailStyle.Render(item.Detail)
		}

		if item.Deprecated {
			line = lipgloss.NewStyle().Strikethrough(true).Faint(true).Render(line)
		}

		if idx == i.cursor {
			line = lipgloss.NewStyle().
				Background(lipgloss.Color("#31353D")).
				Foreground(lipgloss.Color("#E5E7EB")).
				Render("▸ " + line)
		} else {
			line = "  " + line
		}

		rows = append(rows, line)
	}

	if len(i.filtered) > i.maxVisible {
		indicator := i.scrollIndicator()
		rows = append(rows, lipgloss.NewStyle().Foreground(titleColor).Render(indicator))
	}

	return panelStyle.Render(strings.Join(rows, "\n"))
}

func (i *IntelliSense) renderDocumentationPanel() string {
	borderColor := lipgloss.Color("#3C414B")
	bgColor := lipgloss.Color("#252930")
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#7A818A"))
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E5E7EB"))

	item, ok := i.selectedItem()
	if !ok {
		panel := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Background(bgColor).
			Padding(0, 1).
			Width(i.docWidth())
		return panel.Render(muted.Render("No documentation"))
	}

	title := item.Label
	if title == "" {
		title = item.InsertText
	}
	doc := strings.TrimSpace(item.Documentation)
	if doc == "" {
		doc = "No documentation available."
	}

	content := header.Render(title) + "\n" + muted.Render(strings.TrimSpace(item.Detail))
	if strings.TrimSpace(item.Detail) == "" {
		content = header.Render(title)
	}
	content += "\n\n" + doc

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(bgColor).
		Padding(0, 1).
		Width(i.docWidth())

	return panel.Render(content)
}

func (i *IntelliSense) selectedItem() (CompletionItem, bool) {
	if len(i.filtered) == 0 {
		return CompletionItem{}, false
	}
	i.clampCursor()
	return i.filtered[i.cursor], true
}

func (i *IntelliSense) applyFilter() {
	q := strings.ToLower(strings.TrimSpace(i.query))
	if q == "" {
		i.filtered = append([]CompletionItem(nil), i.items...)
		i.cursor = 0
		i.scrollOffset = 0
		return
	}

	type scoredItem struct {
		item  CompletionItem
		score int
	}

	scored := make([]scoredItem, 0, len(i.items))
	for _, item := range i.items {
		score, ok := completionMatchScore(q, item)
		if !ok {
			continue
		}
		scored = append(scored, scoredItem{item: item, score: score})
	}

	sort.SliceStable(scored, func(a, b int) bool {
		if scored[a].score == scored[b].score {
			left := strings.ToLower(scored[a].item.SortText)
			right := strings.ToLower(scored[b].item.SortText)
			if left == "" {
				left = strings.ToLower(scored[a].item.Label)
			}
			if right == "" {
				right = strings.ToLower(scored[b].item.Label)
			}
			return left < right
		}
		return scored[a].score > scored[b].score
	})

	i.filtered = make([]CompletionItem, 0, len(scored))
	for _, entry := range scored {
		i.filtered = append(i.filtered, entry.item)
	}

	i.cursor = 0
	i.scrollOffset = 0
	i.ensureCursorVisible()
}

func completionMatchScore(query string, item CompletionItem) (int, bool) {
	candidates := []string{
		strings.ToLower(strings.TrimSpace(item.FilterText)),
		strings.ToLower(strings.TrimSpace(item.Label)),
		strings.ToLower(strings.TrimSpace(item.InsertText)),
	}

	best := -1
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if candidate == query {
			if best < 300 {
				best = 300
			}
			continue
		}
		if strings.HasPrefix(candidate, query) {
			if best < 250 {
				best = 250
			}
			continue
		}
		if fuzzyPrefix(candidate, query) {
			if best < 120 {
				best = 120
			}
		}
	}

	if best < 0 {
		return 0, false
	}
	return best, true
}

func fuzzyPrefix(candidate, query string) bool {
	if query == "" {
		return true
	}
	if len(candidate) == 0 {
		return false
	}

	qi := 0
	for _, r := range candidate {
		if qi >= len(query) {
			break
		}
		if byte(r) == query[qi] {
			qi++
		}
	}
	return qi == len(query)
}

func trimLastRune(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	return string(runes[:len(runes)-1])
}

func (i *IntelliSense) clampCursor() {
	if len(i.filtered) == 0 {
		i.cursor = 0
		i.scrollOffset = 0
		return
	}
	if i.cursor < 0 {
		i.cursor = 0
	}
	if i.cursor >= len(i.filtered) {
		i.cursor = len(i.filtered) - 1
	}
}

func (i *IntelliSense) ensureCursorVisible() {
	if len(i.filtered) == 0 {
		i.scrollOffset = 0
		return
	}
	i.clampCursor()

	visible := i.maxVisible
	if visible <= 0 || visible > len(i.filtered) {
		visible = len(i.filtered)
	}

	if i.cursor < i.scrollOffset {
		i.scrollOffset = i.cursor
	}
	if i.cursor >= i.scrollOffset+visible {
		i.scrollOffset = i.cursor - visible + 1
	}
	if i.scrollOffset < 0 {
		i.scrollOffset = 0
	}
	maxOffset := len(i.filtered) - visible
	if maxOffset < 0 {
		maxOffset = 0
	}
	if i.scrollOffset > maxOffset {
		i.scrollOffset = maxOffset
	}
}

func (i *IntelliSense) visibleRange() (int, int) {
	if len(i.filtered) == 0 {
		return 0, 0
	}

	visible := i.maxVisible
	if visible <= 0 || visible > len(i.filtered) {
		visible = len(i.filtered)
	}

	start := i.scrollOffset
	if start < 0 {
		start = 0
	}
	if start > len(i.filtered)-1 {
		start = len(i.filtered) - 1
	}
	end := start + visible
	if end > len(i.filtered) {
		end = len(i.filtered)
	}
	return start, end
}

func (i *IntelliSense) scrollIndicator() string {
	if len(i.filtered) == 0 {
		return ""
	}
	start, end := i.visibleRange()
	top := " "
	bottom := " "
	if start > 0 {
		top = "↑"
	}
	if end < len(i.filtered) {
		bottom = "↓"
	}
	return fmt.Sprintf("%s %d-%d/%d %s", top, start+1, end, len(i.filtered), bottom)
}

func (i *IntelliSense) docWidth() int {
	if i.width <= 0 {
		return 40
	}
	if i.width < 30 {
		return 30
	}
	return i.width
}

var _ tea.Model = (*IntelliSense)(nil)
