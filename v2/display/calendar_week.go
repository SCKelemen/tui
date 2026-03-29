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
	weekViewDefaultWidth  = 120
	weekViewDefaultHeight = 36
	weekViewGutterWidth   = 7
)

// WeekView renders a week calendar timeline.
type WeekView struct {
	events       []CalendarEvent
	date         time.Time // any day in the week to display
	width        int
	height       int
	scrollOffset int
	focused      bool
	colors       calendarColors
}

// WeekViewOption configures a WeekView component.
type WeekViewOption func(*WeekView)

// WithWeekViewEvents sets the events rendered in the week view.
func WithWeekViewEvents(events []CalendarEvent) WeekViewOption {
	return func(w *WeekView) {
		w.events = append([]CalendarEvent(nil), events...)
	}
}

// WithWeekViewDate sets the currently displayed week via any day in that week.
func WithWeekViewDate(date time.Time) WeekViewOption {
	return func(w *WeekView) {
		w.date = normalizeCalendarDate(date)
	}
}

// WithWeekViewDesignTokens applies design tokens to calendar colors.
func WithWeekViewDesignTokens(tokens *design.DesignTokens) WeekViewOption {
	return func(w *WeekView) {
		w.colors = calendarColorsFromTokens(tokens)
	}
}

// WithWeekViewWidth sets the rendered width in cells.
func WithWeekViewWidth(width int) WeekViewOption {
	return func(w *WeekView) {
		if width > 0 {
			w.width = width
		}
	}
}

// WithWeekViewHeight sets the rendered height in rows.
func WithWeekViewHeight(height int) WeekViewOption {
	return func(w *WeekView) {
		if height > 0 {
			w.height = height
		}
	}
}

// NewWeekView creates a week calendar component.
func NewWeekView(opts ...WeekViewOption) *WeekView {
	now := time.Now()
	w := &WeekView{
		events:       make([]CalendarEvent, 0),
		date:         normalizeCalendarDate(now),
		width:        weekViewDefaultWidth,
		height:       weekViewDefaultHeight,
		scrollOffset: 0,
		focused:      false,
		colors:       defaultCalendarColors(),
	}

	for _, opt := range opts {
		opt(w)
	}

	w.date = normalizeCalendarDate(w.date)
	w.scrollOffset = w.initialScrollOffset()
	return w
}

// Init initializes the component.
func (w *WeekView) Init() tea.Cmd {
	return nil
}

// Update handles keyboard navigation and scrolling.
func (w *WeekView) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			w.width = msg.Width
		}
		if msg.Height > 0 {
			w.height = msg.Height
		}
		w.clampScrollOffset()
		return w, nil
	case tea.KeyMsg:
		if !w.focused {
			return w, nil
		}

		switch msg.String() {
		case "h", "left":
			w.date = w.date.AddDate(0, 0, -7)
			w.scrollOffset = w.initialScrollOffset()
			return w, nil
		case "l", "right":
			w.date = w.date.AddDate(0, 0, 7)
			w.scrollOffset = w.initialScrollOffset()
			return w, nil
		case "j", "down":
			w.scrollOffset++
			w.clampScrollOffset()
			return w, nil
		case "k", "up":
			w.scrollOffset--
			w.clampScrollOffset()
			return w, nil
		case "t":
			w.date = normalizeCalendarDate(time.Now().In(w.date.Location()))
			w.scrollOffset = w.initialScrollOffset()
			return w, nil
		}
	}

	return w, nil
}

// View renders the week calendar.
func (w *WeekView) View() string {
	w.clampDimensions()
	w.clampScrollOffset()

	weekStart := w.weekStart()
	columnWidth := w.columnWidth()
	allDayLayouts := w.layoutAllDayEvents(weekStart)
	timedLayouts := w.layoutTimedEvents(weekStart)

	visibleTimelineRows := w.height - w.headerRows()
	if visibleTimelineRows < 1 {
		visibleTimelineRows = 1
	}

	startRow := w.scrollOffset * 2
	maxStart := 48 - visibleTimelineRows
	if maxStart < 0 {
		maxStart = 0
	}
	if startRow > maxStart {
		startRow = maxStart
	}
	endRow := startRow + visibleTimelineRows
	if endRow > 48 {
		endRow = 48
	}

	var b strings.Builder
	b.WriteString(w.renderHeader(weekStart, columnWidth))
	b.WriteByte('\n')
	b.WriteString(w.renderDayHeaders(weekStart, columnWidth))
	b.WriteByte('\n')
	b.WriteString(w.renderAllDayBar(allDayLayouts, columnWidth))
	b.WriteByte('\n')
	b.WriteString(style.Fg(w.colors.border))
	b.WriteString(strings.Repeat("─", w.contentWidth(columnWidth)))
	b.WriteString(style.ANSIReset)

	for row := startRow; row < endRow; row++ {
		b.WriteByte('\n')
		b.WriteString(w.renderTimelineRow(weekStart, row, timedLayouts, columnWidth))
	}

	return b.String()
}

// Focus marks the view as focused.
func (w *WeekView) Focus() {
	w.focused = true
}

// Blur marks the view as unfocused.
func (w *WeekView) Blur() {
	w.focused = false
}

// Focused reports focus state.
func (w *WeekView) Focused() bool {
	return w.focused
}

func (w *WeekView) clampDimensions() {
	if w.width <= 0 {
		w.width = weekViewDefaultWidth
	}
	if w.height <= 0 {
		w.height = weekViewDefaultHeight
	}
}

func (w *WeekView) headerRows() int {
	return 4
}

func (w *WeekView) clampScrollOffset() {
	if w.scrollOffset < 0 {
		w.scrollOffset = 0
	}

	visibleTimelineRows := w.height - w.headerRows()
	if visibleTimelineRows < 1 {
		visibleTimelineRows = 1
	}

	maxStart := 48 - visibleTimelineRows
	if maxStart < 0 {
		maxStart = 0
	}
	maxHourOffset := maxStart / 2
	if w.scrollOffset > maxHourOffset {
		w.scrollOffset = maxHourOffset
	}
}

func (w *WeekView) initialScrollOffset() int {
	now := time.Now().In(w.date.Location())
	weekStart := w.weekStart()
	if !now.Before(weekStart) && now.Before(weekStart.AddDate(0, 0, 7)) {
		h := now.Hour() - 3
		if h < 0 {
			h = 0
		}
		return h
	}

	timedLayouts := w.layoutTimedEvents(weekStart)
	for day := 0; day < 7; day++ {
		if len(timedLayouts[day]) == 0 {
			continue
		}
		h := timedLayouts[day][0].startRow/2 - 1
		if h < 0 {
			h = 0
		}
		return h
	}

	return 0
}

func (w *WeekView) weekStart() time.Time {
	day := normalizeCalendarDate(w.date)
	weekday := int(day.Weekday())
	mondayOffset := (weekday + 6) % 7
	return day.AddDate(0, 0, -mondayOffset)
}

func (w *WeekView) columnWidth() int {
	cw := (w.width - weekViewGutterWidth) / 7
	if cw < 1 {
		cw = 1
	}
	return cw
}

func (w *WeekView) contentWidth(columnWidth int) int {
	return weekViewGutterWidth + 7*columnWidth + 6
}

func (w *WeekView) renderHeader(weekStart time.Time, columnWidth int) string {
	weekEnd := weekStart.AddDate(0, 0, 6)
	dateText := weekRangeLabel(weekStart, weekEnd)
	navText := "h/← prev  l/→ next  j/↓ down  k/↑ up  t today"
	if w.focused {
		navText = style.ANSIBold + navText + style.ANSIReset
	}

	head := style.ANSIBold + style.Fg(w.colors.text) + dateText + style.ANSIReset
	nav := style.ANSIDim + style.Fg(w.colors.muted) + navText + style.ANSIReset

	totalWidth := w.contentWidth(columnWidth)
	line := head
	remaining := totalWidth - len(dateText) - 1
	if remaining > len(dayViewStripANSI(navText)) {
		line += strings.Repeat(" ", remaining-len(dayViewStripANSI(navText))) + nav
	} else {
		line += " " + nav
	}
	return line
}

func (w *WeekView) renderDayHeaders(weekStart time.Time, columnWidth int) string {
	var b strings.Builder
	b.WriteString(style.Fg(w.colors.muted))
	b.WriteString("       ")
	b.WriteString(style.ANSIReset)

	for day := 0; day < 7; day++ {
		if day > 0 {
			b.WriteString(style.Fg(w.colors.border))
			b.WriteString("│")
			b.WriteString(style.ANSIReset)
		}
		cur := weekStart.AddDate(0, 0, day)
		label := cur.Format("Mon 2")
		cell := weekCenterPad(label, columnWidth)
		if sameCalendarDay(cur, time.Now().In(cur.Location())) {
			b.WriteString(style.ANSIBold)
			b.WriteString(style.Fg(w.colors.today))
		} else {
			b.WriteString(style.ANSIBold)
			b.WriteString(style.Fg(w.colors.text))
		}
		b.WriteString(cell)
		b.WriteString(style.ANSIReset)
	}

	return b.String()
}

func (w *WeekView) renderAllDayBar(allDay [7]allDaySlot, columnWidth int) string {
	var b strings.Builder
	b.WriteString(style.ANSIDim)
	b.WriteString(style.Fg(w.colors.muted))
	b.WriteString("Allday│")
	b.WriteString(style.ANSIReset)

	for day := 0; day < 7; day++ {
		if day > 0 {
			b.WriteString(style.Fg(w.colors.border))
			b.WriteString("│")
			b.WriteString(style.ANSIReset)
		}

		slot := allDay[day]
		if !slot.active {
			b.WriteString(style.Fg(w.colors.surfaceAlt))
			b.WriteString(strings.Repeat(" ", columnWidth))
			b.WriteString(style.ANSIReset)
			continue
		}

		label := strings.Repeat(" ", columnWidth)
		if slot.startsHere {
			title := strings.TrimSpace(slot.event.Title)
			if title == "" {
				title = "(all-day)"
			}
			text := " " + title + " "
			if len(text) > columnWidth {
				text = text[:columnWidth]
			}
			label = text + strings.Repeat(" ", columnWidth-len(text))
		} else {
			marker := "────"
			if len(marker) > columnWidth {
				marker = strings.Repeat("─", columnWidth)
			}
			label = marker + strings.Repeat(" ", columnWidth-len(marker))
		}

		color := slot.event.Color
		if strings.TrimSpace(color) == "" {
			color = w.colors.accent
		}
		b.WriteString(style.Bg(color))
		b.WriteString(style.Fg(w.colors.surface))
		b.WriteString(label)
		b.WriteString(style.ANSIReset)
	}

	return b.String()
}

func (w *WeekView) renderTimelineRow(weekStart time.Time, row int, timed [7][]timedLayout, columnWidth int) string {
	hour := row / 2
	half := row % 2
	gutter := "      │"
	if half == 0 {
		gutter = fmt.Sprintf("%02d:00 │", hour)
	}
	gutter = style.Fg(w.colors.muted) + gutter + style.ANSIReset

	if w.isNowRow(weekStart, row) {
		return gutter + style.Fg(w.colors.error) + style.ANSIBold + w.nowLine(columnWidth) + style.ANSIReset
	}

	var b strings.Builder
	b.WriteString(gutter)
	for day := 0; day < 7; day++ {
		if day > 0 {
			b.WriteString(style.Fg(w.colors.border))
			b.WriteString("│")
			b.WriteString(style.ANSIReset)
		}
		b.WriteString(w.renderTimedCell(row, timed[day], columnWidth))
	}
	return b.String()
}

func (w *WeekView) nowLine(columnWidth int) string {
	var b strings.Builder
	for day := 0; day < 7; day++ {
		if day > 0 {
			b.WriteString("│")
		}
		b.WriteString(strings.Repeat("─", columnWidth))
	}
	return b.String()
}

func (w *WeekView) isNowRow(weekStart time.Time, row int) bool {
	now := time.Now().In(weekStart.Location())
	if now.Before(weekStart) || !now.Before(weekStart.AddDate(0, 0, 7)) {
		return false
	}
	nowRow := now.Hour()*2 + now.Minute()/30
	return row == nowRow
}

func (w *WeekView) renderTimedCell(row int, events []timedLayout, columnWidth int) string {
	overlaps := make([]timedLayout, 0, 3)
	for _, ev := range events {
		if row >= ev.startRow && row < ev.endRow {
			overlaps = append(overlaps, ev)
		}
	}

	if len(overlaps) == 0 {
		if row%2 == 0 {
			return style.Fg(w.colors.border) + strings.Repeat("─", columnWidth) + style.ANSIReset
		}
		return style.Fg(w.colors.surfaceAlt) + strings.Repeat(" ", columnWidth) + style.ANSIReset
	}

	sort.Slice(overlaps, func(i, j int) bool {
		if overlaps[i].lane == overlaps[j].lane {
			if overlaps[i].startRow == overlaps[j].startRow {
				return overlaps[i].endRow < overlaps[j].endRow
			}
			return overlaps[i].startRow < overlaps[j].startRow
		}
		return overlaps[i].lane < overlaps[j].lane
	})

	laneCount := overlaps[0].laneCount
	for _, ev := range overlaps {
		if ev.laneCount > laneCount {
			laneCount = ev.laneCount
		}
	}
	if laneCount < 1 {
		laneCount = 1
	}

	if columnWidth/laneCount < 4 {
		return w.renderSingleTimedEventCell(overlaps[0], row, columnWidth)
	}

	segments := make([]string, laneCount)
	for i := 0; i < laneCount; i++ {
		segments[i] = style.Fg(w.colors.surfaceAlt) + strings.Repeat(" ", weekLaneWidth(columnWidth, laneCount, i)) + style.ANSIReset
	}

	for _, ev := range overlaps {
		lane := ev.lane
		if lane < 0 || lane >= laneCount {
			continue
		}
		segments[lane] = w.renderSingleTimedEventCell(ev, row, weekLaneWidth(columnWidth, laneCount, lane))
	}

	var b strings.Builder
	for _, seg := range segments {
		b.WriteString(seg)
	}
	return b.String()
}

func (w *WeekView) renderSingleTimedEventCell(ev timedLayout, row int, width int) string {
	if width < 1 {
		return ""
	}

	fill := strings.Repeat(string(w.blockRuneForRow(ev.event, ev.dayStart, row)), width)
	if row == ev.startRow {
		title := strings.TrimSpace(ev.event.Title)
		if title == "" {
			title = "(untitled)"
		}
		text := " " + title
		if len(text) > width {
			text = text[:width]
		}
		fill = text + strings.Repeat(" ", width-len(text))
	}

	color := ev.event.Color
	if strings.TrimSpace(color) == "" {
		color = w.colors.accent
	}

	return style.Bg(color) + style.Fg(w.colors.surface) + fill + style.ANSIReset
}

func (w *WeekView) blockRuneForRow(ev CalendarEvent, dayStart time.Time, row int) rune {
	cellStart := dayStart.Add(time.Duration(row) * 30 * time.Minute)
	cellEnd := cellStart.Add(30 * time.Minute)
	start, end := clampEventToDay(ev, dayStart)
	overlapStart := maxTime(start, cellStart)
	overlapEnd := minTime(end, cellEnd)
	overlap := overlapEnd.Sub(overlapStart)

	if overlap <= 0 {
		return ' '
	}
	if overlap >= 25*time.Minute {
		return '█'
	}
	if overlapStart.Equal(cellStart) {
		return '▀'
	}
	if overlapEnd.Equal(cellEnd) {
		return '▄'
	}
	if overlapStart.Sub(cellStart) < 15*time.Minute {
		return '▀'
	}
	return '▄'
}

type allDaySlot struct {
	event      CalendarEvent
	active     bool
	startsHere bool
}

type timedLayout struct {
	event     CalendarEvent
	dayStart  time.Time
	startRow  int
	endRow    int
	lane      int
	laneCount int
}

func (w *WeekView) layoutAllDayEvents(weekStart time.Time) [7]allDaySlot {
	var slots [7]allDaySlot
	weekEnd := weekStart.AddDate(0, 0, 7)

	allDay := make([]CalendarEvent, 0, len(w.events))
	for _, ev := range w.events {
		if !ev.AllDay {
			continue
		}
		if !eventIntersectsDay(ev, weekStart, weekEnd) {
			continue
		}
		allDay = append(allDay, ev)
	}

	sort.Slice(allDay, func(i, j int) bool {
		if allDay[i].Start.Equal(allDay[j].Start) {
			return allDay[i].End.Before(allDay[j].End)
		}
		return allDay[i].Start.Before(allDay[j].Start)
	})

	for _, ev := range allDay {
		for day := 0; day < 7; day++ {
			dayStart := weekStart.AddDate(0, 0, day)
			dayEnd := dayStart.Add(24 * time.Hour)
			if !eventIntersectsDay(ev, dayStart, dayEnd) {
				continue
			}
			if slots[day].active {
				continue
			}
			slots[day] = allDaySlot{
				event:      ev,
				active:     true,
				startsHere: !ev.Start.In(dayStart.Location()).Before(dayStart),
			}
		}
	}

	return slots
}

func (w *WeekView) layoutTimedEvents(weekStart time.Time) [7][]timedLayout {
	var perDay [7][]timedLayout

	for day := 0; day < 7; day++ {
		dayStart := weekStart.AddDate(0, 0, day)
		dayEnd := dayStart.Add(24 * time.Hour)
		days := make([]timedLayout, 0, 8)

		for _, ev := range w.events {
			if ev.AllDay {
				continue
			}
			if !eventIntersectsDay(ev, dayStart, dayEnd) {
				continue
			}

			start, end := clampEventToDay(ev, dayStart)
			startRow := int(start.Sub(dayStart) / (30 * time.Minute))
			endRow := int((end.Sub(dayStart) + (30*time.Minute - time.Nanosecond)) / (30 * time.Minute))
			if endRow <= startRow {
				endRow = startRow + 1
			}
			if startRow < 0 {
				startRow = 0
			}
			if endRow > 48 {
				endRow = 48
			}

			days = append(days, timedLayout{
				event:    ev,
				dayStart: dayStart,
				startRow: startRow,
				endRow:   endRow,
			})
		}

		sort.Slice(days, func(i, j int) bool {
			if days[i].startRow == days[j].startRow {
				if days[i].endRow == days[j].endRow {
					return strings.TrimSpace(days[i].event.Title) < strings.TrimSpace(days[j].event.Title)
				}
				return days[i].endRow < days[j].endRow
			}
			return days[i].startRow < days[j].startRow
		})

		lanesEnd := make([]int, 0, 4)
		for i := range days {
			assigned := -1
			for lane := 0; lane < len(lanesEnd); lane++ {
				if lanesEnd[lane] <= days[i].startRow {
					assigned = lane
					break
				}
			}
			if assigned == -1 {
				lanesEnd = append(lanesEnd, days[i].endRow)
				assigned = len(lanesEnd) - 1
			} else {
				lanesEnd[assigned] = days[i].endRow
			}
			days[i].lane = assigned
			days[i].laneCount = assigned + 1
		}

		for i := range days {
			maxLane := days[i].lane
			for j := range days {
				if i == j {
					continue
				}
				if days[j].endRow <= days[i].startRow || days[j].startRow >= days[i].endRow {
					continue
				}
				if days[j].lane > maxLane {
					maxLane = days[j].lane
				}
			}
			days[i].laneCount = maxLane + 1
		}

		perDay[day] = days
	}

	return perDay
}

func weekLaneWidth(total, lanes, lane int) int {
	if lanes <= 1 {
		return total
	}
	base := total / lanes
	extra := total % lanes
	if lane < extra {
		return base + 1
	}
	return base
}

func weekCenterPad(s string, width int) string {
	if width <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) >= width {
		return string(r[:width])
	}
	left := (width - len(r)) / 2
	right := width - len(r) - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func weekRangeLabel(start, end time.Time) string {
	if start.Year() == end.Year() {
		if start.Month() == end.Month() {
			return fmt.Sprintf("%s %d – %d, %d", start.Format("Jan"), start.Day(), end.Day(), start.Year())
		}
		return fmt.Sprintf("%s %d – %s %d, %d", start.Format("Jan"), start.Day(), end.Format("Jan"), end.Day(), start.Year())
	}
	return fmt.Sprintf("%s %d, %d – %s %d, %d", start.Format("Jan"), start.Day(), start.Year(), end.Format("Jan"), end.Day(), end.Year())
}

var _ tui.Component = (*WeekView)(nil)
