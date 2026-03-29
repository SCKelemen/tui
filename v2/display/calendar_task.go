package display

import (
	"fmt"
	"sort"
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	taskViewDefaultWidth = 78
	ansiStrikethrough    = "\033[9m"
)

// TaskView renders a task list grouped by due-date buckets.
type TaskView struct {
	tasks         []CalendarTask
	cursor        int
	width         int
	focused       bool
	colors        calendarColors
	showCompleted bool
}

// TaskViewOption configures a TaskView component.
type TaskViewOption func(*TaskView)

// WithTaskViewTasks sets the rendered task list.
func WithTaskViewTasks(tasks []CalendarTask) TaskViewOption {
	return func(v *TaskView) {
		v.tasks = append([]CalendarTask(nil), tasks...)
	}
}

// WithTaskViewDesignTokens applies design tokens to calendar colors.
func WithTaskViewDesignTokens(tokens *design.DesignTokens) TaskViewOption {
	return func(v *TaskView) {
		v.colors = calendarColorsFromTokens(tokens)
	}
}

// WithTaskViewWidth sets the rendered width in terminal cells.
func WithTaskViewWidth(width int) TaskViewOption {
	return func(v *TaskView) {
		if width > 0 {
			v.width = width
		}
	}
}

// NewTaskView creates a task calendar component.
func NewTaskView(opts ...TaskViewOption) *TaskView {
	v := &TaskView{
		tasks:         make([]CalendarTask, 0),
		cursor:        0,
		width:         taskViewDefaultWidth,
		focused:       false,
		colors:        defaultCalendarColors(),
		showCompleted: false,
	}

	for _, opt := range opts {
		opt(v)
	}

	if v.width <= 0 {
		v.width = taskViewDefaultWidth
	}

	return v
}

// Init initializes the component.
func (v *TaskView) Init() tea.Cmd {
	return nil
}

// Update handles navigation and completion toggling.
func (v *TaskView) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			v.width = msg.Width
		}
		v.clampCursor()
		return v, nil
	case tea.KeyMsg:
		if !v.focused {
			return v, nil
		}

		switch msg.String() {
		case "j", "down":
			if len(v.visibleTaskRefs()) > 0 {
				v.cursor++
				v.clampCursor()
			}
			return v, nil
		case "k", "up":
			if len(v.visibleTaskRefs()) > 0 {
				v.cursor--
				v.clampCursor()
			}
			return v, nil
		case "enter", " ":
			v.toggleCurrentTaskCompletion()
			v.clampCursor()
			return v, nil
		case "t":
			v.jumpToTodayTask()
			v.clampCursor()
			return v, nil
		}
	}

	return v, nil
}

// View renders grouped task sections.
func (v *TaskView) View() string {
	if v.width <= 0 {
		v.width = taskViewDefaultWidth
	}

	groups := v.groupedTasks()
	visible := v.visibleTaskRefsFromGroups(groups)
	if len(visible) == 0 {
		v.cursor = 0
	} else {
		v.clampCursor()
	}

	selectedKey := ""
	if len(visible) > 0 {
		selectedKey = visible[v.cursor].key
	}

	var b strings.Builder
	b.WriteString(style.ANSIBold)
	b.WriteString(style.Fg(v.colors.text))
	b.WriteString("Tasks")
	b.WriteString(style.ANSIReset)
	b.WriteString(" ")
	b.WriteString(style.ANSIDim)
	b.WriteString(style.Fg(v.colors.muted))
	b.WriteString("j/k navigate • enter/space toggle • t today")
	b.WriteString(style.ANSIReset)

	for _, section := range groups {
		b.WriteByte('\n')
		b.WriteString(v.renderSectionHeader(section))
		if len(section.tasks) == 0 {
			continue
		}

		if section.name == "Completed" && !v.showCompleted {
			continue
		}

		for _, ref := range section.tasks {
			b.WriteByte('\n')
			selected := selectedKey != "" && selectedKey == ref.key
			b.WriteString(v.renderTaskLine(ref.task, section.name, selected))
		}
	}

	return b.String()
}

// Focus marks this component as focused.
func (v *TaskView) Focus() {
	v.focused = true
}

// Blur marks this component as unfocused.
func (v *TaskView) Blur() {
	v.focused = false
}

// Focused reports focus state.
func (v *TaskView) Focused() bool {
	return v.focused
}

type taskRef struct {
	index int
	key   string
	task  CalendarTask
}

type taskGroup struct {
	name       string
	headerColor string
	tasks      []taskRef
}

func (v *TaskView) groupedTasks() []taskGroup {
	today := normalizeCalendarDate(time.Now())
	weekEnd := today.AddDate(0, 0, 7)

	overdue := make([]taskRef, 0)
	todayTasks := make([]taskRef, 0)
	thisWeek := make([]taskRef, 0)
	upcoming := make([]taskRef, 0)
	completed := make([]taskRef, 0)

	for i, task := range v.tasks {
		ref := taskRef{index: i, key: v.taskKey(i, task), task: task}
		due := normalizeCalendarDate(task.DueDate)
		if task.Completed {
			completed = append(completed, ref)
			continue
		}
		if due.Before(today) {
			overdue = append(overdue, ref)
			continue
		}
		if sameCalendarDay(due, today) {
			todayTasks = append(todayTasks, ref)
			continue
		}
		if !due.After(weekEnd) {
			thisWeek = append(thisWeek, ref)
			continue
		}
		upcoming = append(upcoming, ref)
	}

	sortTaskRefs(overdue)
	sortTaskRefs(todayTasks)
	sortTaskRefs(thisWeek)
	sortTaskRefs(upcoming)
	sortTaskRefs(completed)

	return []taskGroup{
		{name: "Overdue", headerColor: v.colors.error, tasks: overdue},
		{name: "Today", headerColor: v.colors.accent, tasks: todayTasks},
		{name: "This Week", headerColor: v.colors.text, tasks: thisWeek},
		{name: "Upcoming", headerColor: v.colors.muted, tasks: upcoming},
		{name: "Completed", headerColor: v.colors.muted, tasks: completed},
	}
}

func (v *TaskView) visibleTaskRefs() []taskRef {
	return v.visibleTaskRefsFromGroups(v.groupedTasks())
}

func (v *TaskView) visibleTaskRefsFromGroups(groups []taskGroup) []taskRef {
	visible := make([]taskRef, 0)
	for _, group := range groups {
		if group.name == "Completed" && !v.showCompleted {
			continue
		}
		visible = append(visible, group.tasks...)
	}
	return visible
}

func sortTaskRefs(tasks []taskRef) {
	sort.Slice(tasks, func(i, j int) bool {
		a := normalizeCalendarDate(tasks[i].task.DueDate)
		b := normalizeCalendarDate(tasks[j].task.DueDate)
		if a.Equal(b) {
			return strings.ToLower(strings.TrimSpace(tasks[i].task.Title)) < strings.ToLower(strings.TrimSpace(tasks[j].task.Title))
		}
		return a.Before(b)
	})
}

func (v *TaskView) renderSectionHeader(section taskGroup) string {
	title := section.name
	if section.name == "Completed" {
		if v.showCompleted {
			title = fmt.Sprintf("%s (%d) ▾", section.name, len(section.tasks))
		} else {
			title = fmt.Sprintf("%s (%d) ▸", section.name, len(section.tasks))
		}
	} else {
		title = fmt.Sprintf("%s (%d)", section.name, len(section.tasks))
	}

	headerStyle := style.ANSIBold + style.Fg(section.headerColor)
	if section.name == "Upcoming" || section.name == "Completed" {
		headerStyle = style.ANSIDim + style.Fg(section.headerColor)
	}
	return headerStyle + title + style.ANSIReset
}

func (v *TaskView) renderTaskLine(task CalendarTask, section string, selected bool) string {
	checkbox := "☐"
	if task.Completed {
		checkbox = "☑"
	}
	priorityGlyph, priorityColor := taskPriorityVisual(task.Priority, v.colors)
	dueText := task.DueDate.Format("Jan 2")

	title := strings.TrimSpace(task.Title)
	if title == "" {
		title = "(untitled task)"
	}

	cursorPrefix := "  "
	if selected {
		cursorPrefix = "› "
	}

	leftPrefix := fmt.Sprintf("%s %s ", checkbox, priorityGlyph)
	titleWidth := v.width - style.StringWidth(cursorPrefix) - style.StringWidth(leftPrefix) - style.StringWidth(dueText) - 1
	if titleWidth < 4 {
		titleWidth = 4
	}

	title = style.Truncate(title, titleWidth, "…")
	title = style.Pad(title, titleWidth)

	checkboxStyle := style.Fg(v.colors.text)
	if task.Completed {
		checkboxStyle = style.Fg(v.colors.success)
	}

	taskColor := v.colors.text
	switch section {
	case "Overdue":
		taskColor = v.colors.error
	case "Today":
		taskColor = v.colors.accent
	case "Upcoming":
		taskColor = v.colors.muted
	case "Completed":
		taskColor = v.colors.muted
	}

	titleStyle := style.Fg(taskColor)
	dueStyle := style.Fg(v.colors.muted)
	if section == "Overdue" {
		dueStyle = style.Fg(v.colors.error)
	}
	if task.Completed {
		titleStyle = style.ANSIDim + ansiStrikethrough + style.Fg(v.colors.muted)
		dueStyle = style.ANSIDim + style.Fg(v.colors.muted)
	}

	line := cursorPrefix +
		checkboxStyle + checkbox + style.ANSIReset + " " +
		style.Fg(priorityColor) + priorityGlyph + style.ANSIReset + " " +
		titleStyle + title + style.ANSIReset + " " +
		dueStyle + dueText + style.ANSIReset

	if selected && v.focused {
		line = style.ANSIInverse + line + style.ANSIReset
	}

	return line
}

func taskPriorityVisual(priority TaskPriority, colors calendarColors) (string, string) {
	switch priority {
	case TaskPriorityHigh:
		return "▲", colors.error
	case TaskPriorityMedium:
		return "─", colors.warning
	case TaskPriorityLow:
		return "▼", colors.accent
	default:
		return "·", colors.muted
	}
}

func (v *TaskView) toggleCurrentTaskCompletion() {
	visible := v.visibleTaskRefs()
	if len(visible) == 0 {
		return
	}
	v.clampCursor()
	ref := visible[v.cursor]
	if ref.index < 0 || ref.index >= len(v.tasks) {
		return
	}
	v.tasks[ref.index].Completed = !v.tasks[ref.index].Completed
}

func (v *TaskView) jumpToTodayTask() {
	groups := v.groupedTasks()
	todayRefs := make([]taskRef, 0)
	visible := v.visibleTaskRefsFromGroups(groups)
	if len(visible) == 0 {
		v.cursor = 0
		return
	}

	for _, group := range groups {
		if group.name == "Today" {
			todayRefs = append(todayRefs, group.tasks...)
			break
		}
	}

	if len(todayRefs) == 0 {
		v.cursor = 0
		return
	}

	target := todayRefs[0].key
	for i, ref := range visible {
		if ref.key == target {
			v.cursor = i
			return
		}
	}
	v.cursor = 0
}

func (v *TaskView) clampCursor() {
	visible := v.visibleTaskRefs()
	if len(visible) == 0 {
		v.cursor = 0
		return
	}
	if v.cursor < 0 {
		v.cursor = 0
	}
	if v.cursor >= len(visible) {
		v.cursor = len(visible) - 1
	}
}

func (v *TaskView) taskKey(index int, task CalendarTask) string {
	return fmt.Sprintf("%d|%d|%s", index, normalizeCalendarDate(task.DueDate).Unix(), strings.TrimSpace(task.Title))
}

var _ tui.Component = (*TaskView)(nil)
