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
	threeDayViewDefaultWidth  = 120
	threeDayViewDefaultHeight = 32
	threeDayViewGutterWidth   = 7 // "00:00 │"
	threeDayViewDays          = 3
)

// ThreeDayView renders a 3-day calendar timeline.
type ThreeDayView struct {
	events       []CalendarEvent
	date         time.Time // first visible day
	width        int
	height       int
	scrollOffset int // which hour row is at top
	focused      bool
	colors       calendarColors
}

// ThreeDayViewOption configures a ThreeDayView component.
type ThreeDayViewOption func(*ThreeDayView)

// WithThreeDayViewEvents sets the events rendered in the 3-day view.
func WithThreeDayViewEvents(events []CalendarEvent) ThreeDayViewOption {
	return func(v *ThreeDayView) {
		v.events = append([]CalendarEvent(nil), events...)
	}
}

// WithThreeDayViewDate sets the first visible date in the 3-day range.
func WithThreeDayViewDate(date time.Time) ThreeDayViewOption {
	return func(v *ThreeDayView) {
		v.date = normalizeCalendarDate(date)
	}
}

// WithThreeDayViewDesignTokens applies design tokens to calendar colors.
func WithThreeDayViewDesignTokens(tokens *design.DesignTokens) ThreeDayViewOption {
	return func(v *ThreeDayView) {
		v.colors = calendarColorsFromTokens(tokens)
	}
}

// WithThreeDayViewWidth sets the rendered width in cells.
func WithThreeDayViewWidth(width int) ThreeDayViewOption {
	return func(v *ThreeDayView) {
		if width > 0 {
			v.width = width
		}
	}
}

// WithThreeDayViewHeight sets the rendered height in rows.
func WithThreeDayViewHeight(height int) ThreeDayViewOption {
	return func(v *ThreeDayView) {
		if height > 0 {
			v.height = height
		}
	}
}

// NewThreeDayView creates a 3-day calendar component.
func NewThreeDayView(opts ...ThreeDayViewOption) *ThreeDayView {
	now := time.Now()
	v := &ThreeDayView{
		events:       make([]CalendarEvent, 0),
		date:         normalizeCalendarDate(now),
		width:        threeDayViewDefaultWidth,
		height:       threeDayViewDefaultHeight,
		scrollOffset: 0,
		colors:       defaultCalendarColors(),
	}

	for _, opt := range opts {
		opt(v)
	}

	v.date = normalizeCalendarDate(v.date)
	v.scrollOffset = v.initialScrollOffset()
	return v
}

// Init initializes the component.
func (v *ThreeDayView) Init() tea.Cmd {
	return nil
}

// Update handles keyboard navigation and scrolling.
func (v *ThreeDayView) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			v.width = msg.Width
		}
		if msg.Height > 0 {
			v.height = msg.Height
		}
		v.clampScrollOffset()
		return v, nil
	case tea.KeyMsg:
		if !v.focused {
			return v, nil
		}

		switch msg.String() {
		case "h", "left":
			v.date = v.date.AddDate(0, 0, -threeDayViewDays)
			v.scrollOffset = v.initialScrollOffset()
			return v, nil
		case "l", "right":
			v.date = v.date.AddDate(0, 0, threeDayViewDays)
			v.scrollOffset = v.initialScrollOffset()
			return v, nil
		case "j", "down":
			v.scrollOffset++
			v.clampScrollOffset()
			return v, nil
		case "k", "up":
			v.scrollOffset--
			v.clampScrollOffset()
			return v, nil
		case "t":
			v.date = normalizeCalendarDate(time.Now().In(v.date.Location()))
			v.scrollOffset = v.initialScrollOffset()
			return v, nil
		}
	}

	return v, nil
}

// View renders the 3-day calendar.
func (v *ThreeDayView) View() string {
	v.clampDimensions()
	v.clampScrollOffset()

	days := v.visibleDays()
	allDayEvents := make([][]CalendarEvent, threeDayViewDays)
	timedEvents := make([][]CalendarEvent, threeDayViewDays)
	for i, day := range days {
		allDayEvents[i], timedEvents[i] = v.eventsForDate(day)
	}

	visibleTimelineRows := v.height - v.headerRows(allDayEvents)
	if visibleTimelineRows < 1 {
		visibleTimelineRows = 1
	}

	startRow := v.scrollOffset * 2
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

	columnWidth := (v.width - threeDayViewGutterWidth) / threeDayViewDays
	if columnWidth < 8 {
		columnWidth = 8
	}

	var b strings.Builder
	b.WriteString(v.renderHeader())
	b.WriteByte('\n')
	b.WriteString(v.renderDayHeader(days, columnWidth))
	b.WriteByte('\n')
	b.WriteString(v.renderAllDayRow(allDayEvents, columnWidth))
	b.WriteByte('\n')
	b.WriteString(style.Fg(v.colors.border))
	b.WriteString(strings.Repeat("─", v.gridWidth(columnWidth)))
	b.WriteString(style.ANSIReset)

	for row := startRow; row < endRow; row++ {
		b.WriteByte('\n')
		b.WriteString(v.renderTimelineRow(row, days, timedEvents, columnWidth))
	}

	return b.String()
}

// Focus marks the view as focused.
func (v *ThreeDayView) Focus() {
	v.focused = true
}

// Blur marks the view as unfocused.
func (v *ThreeDayView) Blur() {
	v.focused = false
}

// Focused reports focus state.
func (v *ThreeDayView) Focused() bool {
	return v.focused
}

func (v *ThreeDayView) clampDimensions() {
	if v.width <= 0 {
		v.width = threeDayViewDefaultWidth
	}
	if v.height <= 0 {
		v.height = threeDayViewDefaultHeight
	}
}

func (v *ThreeDayView) headerRows(_ [][]CalendarEvent) int {
	return 4
}

func (v *ThreeDayView) clampScrollOffset() {
	if v.scrollOffset < 0 {
		v.scrollOffset = 0
	}

	visibleTimelineRows := v.height - v.headerRows(nil)
	if visibleTimelineRows < 1 {
		visibleTimelineRows = 1
	}

	maxStart := 48 - visibleTimelineRows
	if maxStart < 0 {
		maxStart = 0
	}
	maxHourOffset := maxStart / 2
	if v.scrollOffset > maxHourOffset {
		v.scrollOffset = maxHourOffset
	}
}

func (v *ThreeDayView) initialScrollOffset() int {
	now := time.Now().In(v.date.Location())
	for _, day := range v.visibleDays() {
		if sameCalendarDay(day, now) {
			h := now.Hour() - 3
			if h < 0 {
				h = 0
			}
			return h
		}
	}

	earliestHour := 24
	found := false
	for _, day := range v.visibleDays() {
		_, timed := v.eventsForDate(day)
		if len(timed) == 0 {
			continue
		}
		h := timed[0].Start.In(day.Location()).Hour() - 1
		if h < 0 {
			h = 0
		}
		if h < earliestHour {
			earliestHour = h
			found = true
		}
	}
	if found {
		return earliestHour
	}

	return 0
}

func (v *ThreeDayView) visibleDays() []time.Time {
	days := make([]time.Time, threeDayViewDays)
	for i := 0; i < threeDayViewDays; i++ {
		days[i] = v.date.AddDate(0, 0, i)
	}
	return days
}

func (v *ThreeDayView) eventsForDate(day time.Time) ([]CalendarEvent, []CalendarEvent) {
	dayStart := day
	dayEnd := dayStart.Add(24 * time.Hour)

	allDay := make([]CalendarEvent, 0)
	timed := make([]CalendarEvent, 0)

	for _, ev := range v.events {
		if !eventIntersectsDay(ev, dayStart, dayEnd) {
			continue
		}
		if ev.AllDay {
			allDay = append(allDay, ev)
			continue
		}
		timed = append(timed, ev)
	}

	sort.Slice(allDay, func(i, j int) bool {
		if allDay[i].Title == allDay[j].Title {
			return allDay[i].Start.Before(allDay[j].Start)
		}
		return allDay[i].Title < allDay[j].Title
	})
	sort.Slice(timed, func(i, j int) bool {
		if timed[i].Start.Equal(timed[j].Start) {
			return timed[i].End.Before(timed[j].End)
		}
		return timed[i].Start.Before(timed[j].Start)
	})

	return allDay, timed
}

func (v *ThreeDayView) renderHeader() string {
	rangeLabel := fmt.Sprintf("%s – %s", v.date.Format("Jan 2"), v.date.AddDate(0, 0, 2).Format("Jan 2, 2006"))
	navText := "h/← prev 3d  l/→ next 3d  j/↓ down  k/↑ up  t today"
	if v.focused {
		navText = style.ANSIBold + navText + style.ANSIReset
	}

	head := style.ANSIBold + style.Fg(v.colors.text) + rangeLabel + style.ANSIReset
	nav := style.ANSIDim + style.Fg(v.colors.muted) + navText + style.ANSIReset

	line := head
	remaining := v.width - dayViewVisualWidth(rangeLabel) - 1
	if remaining > len(dayViewStripANSI(navText)) {
		line += strings.Repeat(" ", remaining-len(dayViewStripANSI(navText))) + nav
	} else {
		line += " " + nav
	}
	return line
}

func (v *ThreeDayView) renderDayHeader(days []time.Time, colWidth int) string {
	var b strings.Builder
	b.WriteString(style.Fg(v.colors.muted) + "      │" + style.ANSIReset)
	for i, day := range days {
		label := centerPad(day.Format("Mon Jan 2"), colWidth)
		isToday := sameCalendarDay(day, time.Now().In(day.Location()))
		if isToday {
			b.WriteString(style.ANSIBold + style.Fg(v.colors.today) + label + style.ANSIReset)
		} else {
			b.WriteString(style.ANSIBold + style.Fg(v.colors.text) + label + style.ANSIReset)
		}
		if i < len(days)-1 {
			b.WriteString(style.Fg(v.colors.border) + "│" + style.ANSIReset)
		}
	}
	return b.String()
}

func (v *ThreeDayView) renderAllDayRow(allDayEvents [][]CalendarEvent, colWidth int) string {
	var b strings.Builder
	b.WriteString(style.ANSIDim + style.Fg(v.colors.muted) + "Allday│" + style.ANSIReset)
	for i := 0; i < threeDayViewDays; i++ {
		text := "—"
		fg := v.colors.muted
		if len(allDayEvents[i]) > 0 {
			text = strings.TrimSpace(allDayEvents[i][0].Title)
			if text == "" {
				text = "(all-day)"
			}
			if len(allDayEvents[i]) > 1 {
				text = fmt.Sprintf("%s +%d", text, len(allDayEvents[i])-1)
			}
			if strings.TrimSpace(allDayEvents[i][0].Color) != "" {
				fg = allDayEvents[i][0].Color
			} else {
				fg = v.colors.accent
			}
		}
		cell := rightPad(truncateRunes(text, colWidth), colWidth)
		b.WriteString(style.Fg(fg) + cell + style.ANSIReset)
		if i < threeDayViewDays-1 {
			b.WriteString(style.Fg(v.colors.border) + "│" + style.ANSIReset)
		}
	}
	return b.String()
}

func (v *ThreeDayView) renderTimelineRow(row int, days []time.Time, timedEvents [][]CalendarEvent, colWidth int) string {
	hour := row / 2
	half := row % 2

	gutter := "      │"
	if half == 0 {
		gutter = fmt.Sprintf("%02d:00 │", hour)
	}

	var b strings.Builder
	b.WriteString(style.Fg(v.colors.muted) + gutter + style.ANSIReset)

	for i, day := range days {
		if v.isNowRow(day, row) {
			nowLabel := " now "
			if colWidth > len(nowLabel) {
				pad := colWidth - len(nowLabel)
				left := pad / 2
				right := pad - left
				nowLabel = strings.Repeat("─", left) + nowLabel + strings.Repeat("─", right)
			} else {
				nowLabel = truncateRunes(nowLabel, colWidth)
			}
			b.WriteString(style.Fg(v.colors.error) + style.ANSIBold + rightPad(nowLabel, colWidth) + style.ANSIReset)
		} else {
			line := v.renderTimelineCell(day, row, half, timedEvents[i], colWidth)
			b.WriteString(line)
		}

		if i < len(days)-1 {
			b.WriteString(style.Fg(v.colors.border) + "│" + style.ANSIReset)
		}
	}

	return b.String()
}

func (v *ThreeDayView) renderTimelineCell(day time.Time, row int, half int, events []CalendarEvent, colWidth int) string {
	event, hasEvent := v.eventForDayRow(day, row, events)
	if !hasEvent {
		if half == 0 {
			return style.Fg(v.colors.border) + strings.Repeat("─", colWidth) + style.ANSIReset
		}
		return style.Fg(v.colors.surfaceAlt) + strings.Repeat(" ", colWidth) + style.ANSIReset
	}

	fillRune := v.blockRuneForRow(day, event, row)
	color := event.Color
	if strings.TrimSpace(color) == "" {
		color = v.colors.accent
	}

	title := strings.TrimSpace(event.Title)
	if title == "" {
		title = "(untitled)"
	}

	lineBody := strings.Repeat(string(fillRune), colWidth)
	if v.isTitleRow(day, event, row) {
		titleText := " " + title
		if event.Location != "" {
			titleText += " @ " + strings.TrimSpace(event.Location)
		}
		titleText = rightPad(truncateRunes(titleText, colWidth), colWidth)
		lineBody = titleText
	}

	return style.Bg(color) + style.Fg(v.colors.surface) + lineBody + style.ANSIReset
}

func (v *ThreeDayView) eventForDayRow(day time.Time, row int, events []CalendarEvent) (CalendarEvent, bool) {
	cellStart := day.Add(time.Duration(row) * 30 * time.Minute)
	cellEnd := cellStart.Add(30 * time.Minute)

	for _, ev := range events {
		start, end := clampEventToDay(ev, day)
		if end.After(cellStart) && start.Before(cellEnd) {
			return ev, true
		}
	}

	return CalendarEvent{}, false
}

func (v *ThreeDayView) blockRuneForRow(day time.Time, ev CalendarEvent, row int) rune {
	cellStart := day.Add(time.Duration(row) * 30 * time.Minute)
	cellEnd := cellStart.Add(30 * time.Minute)
	start, end := clampEventToDay(ev, day)
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

func (v *ThreeDayView) isTitleRow(day time.Time, ev CalendarEvent, row int) bool {
	start, _ := clampEventToDay(ev, day)
	titleRow := int(start.Sub(day) / (30 * time.Minute))
	if titleRow < 0 {
		titleRow = 0
	}
	if titleRow > 47 {
		titleRow = 47
	}
	return row == titleRow
}

func (v *ThreeDayView) isNowRow(day time.Time, row int) bool {
	now := time.Now().In(day.Location())
	if !sameCalendarDay(day, now) {
		return false
	}
	nowRow := now.Hour()*2 + now.Minute()/30
	return row == nowRow
}

func (v *ThreeDayView) gridWidth(columnWidth int) int {
	// 7 fixed cells in gutter area + 2 separators between 3 columns.
	return threeDayViewGutterWidth + threeDayViewDays*columnWidth + (threeDayViewDays - 1)
}

var _ tui.Component = (*ThreeDayView)(nil)
