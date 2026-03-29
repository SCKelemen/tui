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
	dayViewDefaultWidth  = 100
	dayViewDefaultHeight = 32
	dayViewGutterWidth   = 7 // "00:00 │"
)

// DayView renders a single-day calendar timeline.
type DayView struct {
	events       []CalendarEvent
	date         time.Time
	width        int
	height       int
	scrollOffset int // which hour row is at top
	colors       calendarColors

	focused bool
}

// DayViewOption configures a DayView component.
type DayViewOption func(*DayView)

// WithDayViewEvents sets the events rendered in the day view.
func WithDayViewEvents(events []CalendarEvent) DayViewOption {
	return func(d *DayView) {
		d.events = append([]CalendarEvent(nil), events...)
	}
}

// WithDayViewDate sets the currently displayed date.
func WithDayViewDate(date time.Time) DayViewOption {
	return func(d *DayView) {
		d.date = normalizeCalendarDate(date)
	}
}

// WithDayViewDesignTokens applies design tokens to calendar colors.
func WithDayViewDesignTokens(tokens *design.DesignTokens) DayViewOption {
	return func(d *DayView) {
		d.colors = calendarColorsFromTokens(tokens)
	}
}

// WithDayViewWidth sets the rendered width in cells.
func WithDayViewWidth(width int) DayViewOption {
	return func(d *DayView) {
		if width > 0 {
			d.width = width
		}
	}
}

// WithDayViewHeight sets the rendered height in rows.
func WithDayViewHeight(height int) DayViewOption {
	return func(d *DayView) {
		if height > 0 {
			d.height = height
		}
	}
}

// NewDayView creates a day calendar component.
func NewDayView(opts ...DayViewOption) *DayView {
	now := time.Now()
	d := &DayView{
		events:       make([]CalendarEvent, 0),
		date:         normalizeCalendarDate(now),
		width:        dayViewDefaultWidth,
		height:       dayViewDefaultHeight,
		scrollOffset: 0,
		colors:       defaultCalendarColors(),
	}

	for _, opt := range opts {
		opt(d)
	}

	d.date = normalizeCalendarDate(d.date)
	d.scrollOffset = d.initialScrollOffset()
	return d
}

// Init initializes the component.
func (d *DayView) Init() tea.Cmd {
	return nil
}

// Update handles keyboard navigation and scrolling.
func (d *DayView) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			d.width = msg.Width
		}
		if msg.Height > 0 {
			d.height = msg.Height
		}
		d.clampScrollOffset()
		return d, nil
	case tea.KeyMsg:
		if !d.focused {
			return d, nil
		}

		switch msg.String() {
		case "h", "left":
			d.date = d.date.AddDate(0, 0, -1)
			d.scrollOffset = d.initialScrollOffset()
			return d, nil
		case "l", "right":
			d.date = d.date.AddDate(0, 0, 1)
			d.scrollOffset = d.initialScrollOffset()
			return d, nil
		case "j", "down":
			d.scrollOffset++
			d.clampScrollOffset()
			return d, nil
		case "k", "up":
			d.scrollOffset--
			d.clampScrollOffset()
			return d, nil
		case "t":
			d.date = normalizeCalendarDate(time.Now().In(d.date.Location()))
			d.scrollOffset = d.initialScrollOffset()
			return d, nil
		}
	}

	return d, nil
}

// View renders the day calendar.
func (d *DayView) View() string {
	d.clampDimensions()
	d.clampScrollOffset()

	allDayEvents, timedEvents := d.eventsForDate()

	visibleTimelineRows := d.height - d.headerRows(allDayEvents)
	if visibleTimelineRows < 1 {
		visibleTimelineRows = 1
	}

	startRow := d.scrollOffset * 2
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
	b.WriteString(d.renderHeader())
	b.WriteByte('\n')
	b.WriteString(d.renderAllDayBar(allDayEvents))
	b.WriteByte('\n')
	b.WriteString(style.Fg(d.colors.border))
	b.WriteString(strings.Repeat("─", d.width))
	b.WriteString(style.ANSIReset)

	for row := startRow; row < endRow; row++ {
		b.WriteByte('\n')
		b.WriteString(d.renderTimelineRow(row, timedEvents))
	}

	return b.String()
}

// Focus marks the view as focused.
func (d *DayView) Focus() {
	d.focused = true
}

// Blur marks the view as unfocused.
func (d *DayView) Blur() {
	d.focused = false
}

// Focused reports focus state.
func (d *DayView) Focused() bool {
	return d.focused
}

func (d *DayView) clampDimensions() {
	if d.width <= 0 {
		d.width = dayViewDefaultWidth
	}
	if d.height <= 0 {
		d.height = dayViewDefaultHeight
	}
}

func (d *DayView) headerRows(allDayEvents []CalendarEvent) int {
	_ = allDayEvents
	return 3
}

func (d *DayView) clampScrollOffset() {
	if d.scrollOffset < 0 {
		d.scrollOffset = 0
	}

	visibleTimelineRows := d.height - d.headerRows(nil)
	if visibleTimelineRows < 1 {
		visibleTimelineRows = 1
	}

	maxStart := 48 - visibleTimelineRows
	if maxStart < 0 {
		maxStart = 0
	}
	maxHourOffset := maxStart / 2
	if d.scrollOffset > maxHourOffset {
		d.scrollOffset = maxHourOffset
	}
}

func (d *DayView) initialScrollOffset() int {
	if sameCalendarDay(d.date, time.Now().In(d.date.Location())) {
		now := time.Now().In(d.date.Location())
		h := now.Hour() - 3
		if h < 0 {
			h = 0
		}
		return h
	}

	_, timed := d.eventsForDate()
	if len(timed) > 0 {
		h := timed[0].Start.In(d.date.Location()).Hour() - 1
		if h < 0 {
			h = 0
		}
		return h
	}

	return 0
}

func (d *DayView) eventsForDate() ([]CalendarEvent, []CalendarEvent) {
	dayStart := d.date
	dayEnd := dayStart.Add(24 * time.Hour)

	allDay := make([]CalendarEvent, 0)
	timed := make([]CalendarEvent, 0)

	for _, ev := range d.events {
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

func (d *DayView) renderHeader() string {
	dateText := d.date.Format("Monday, January 2, 2006")
	navText := "h/← prev  l/→ next  j/↓ down  k/↑ up  t today"
	if d.focused {
		navText = style.ANSIBold + navText + style.ANSIReset
	}

	head := style.ANSIBold + style.Fg(d.colors.text) + dateText + style.ANSIReset
	nav := style.ANSIDim + style.Fg(d.colors.muted) + navText + style.ANSIReset

	line := head
	remaining := d.width - dayViewVisualWidth(dateText) - 1
	if remaining > len(dayViewStripANSI(navText)) {
		line += strings.Repeat(" ", remaining-len(dayViewStripANSI(navText))) + nav
	} else {
		line += " " + nav
	}
	return line
}

func (d *DayView) renderAllDayBar(events []CalendarEvent) string {
	label := style.ANSIDim + style.Fg(d.colors.muted) + "All-day:" + style.ANSIReset + " "

	if len(events) == 0 {
		return label + style.ANSIDim + style.Fg(d.colors.muted) + "—" + style.ANSIReset
	}

	var b strings.Builder
	b.WriteString(label)

	for i, ev := range events {
		if i > 0 {
			b.WriteString(" ")
		}
		color := ev.Color
		if strings.TrimSpace(color) == "" {
			color = d.colors.accent
		}
		chip := " " + strings.TrimSpace(ev.Title) + " "
		if chip == "  " {
			chip = " (all-day) "
		}
		b.WriteString(style.Bg(color))
		b.WriteString(style.Fg(d.colors.surface))
		b.WriteString(chip)
		b.WriteString(style.ANSIReset)
	}

	return b.String()
}

func (d *DayView) renderTimelineRow(row int, timedEvents []CalendarEvent) string {
	hour := row / 2
	half := row % 2

	gutter := "      │"
	if half == 0 {
		gutter = fmt.Sprintf("%02d:00 │", hour)
	}
	gutter = style.Fg(d.colors.muted) + gutter + style.ANSIReset

	bodyWidth := d.width - dayViewGutterWidth
	if bodyWidth < 1 {
		bodyWidth = 1
	}

	if d.isNowRow(row) {
		nowLabel := "── now ──"
		if len(nowLabel) > bodyWidth {
			nowLabel = "now"
		}
		padding := bodyWidth - len(nowLabel)
		leftPad := padding / 2
		rightPad := padding - leftPad
		line := strings.Repeat("─", leftPad) + nowLabel + strings.Repeat("─", rightPad)
		return gutter + style.Fg(d.colors.error) + style.ANSIBold + line + style.ANSIReset
	}

	event, hasEvent := d.eventForRow(row, timedEvents)
	if !hasEvent {
		if half == 0 {
			return gutter + style.Fg(d.colors.border) + strings.Repeat("─", bodyWidth) + style.ANSIReset
		}
		return gutter + style.Fg(d.colors.surfaceAlt) + strings.Repeat(" ", bodyWidth) + style.ANSIReset
	}

	fillRune := d.blockRuneForRow(event, row)
	color := event.Color
	if strings.TrimSpace(color) == "" {
		color = d.colors.accent
	}

	title := strings.TrimSpace(event.Title)
	if title == "" {
		title = "(untitled)"
	}

	lineBody := strings.Repeat(string(fillRune), bodyWidth)
	if d.isTitleRow(event, row) {
		titleText := " " + title
		if event.Location != "" {
			titleText += " @ " + strings.TrimSpace(event.Location)
		}
		if len(titleText) > bodyWidth {
			titleText = titleText[:bodyWidth]
		}
		lineBody = titleText + strings.Repeat(string(fillRune), bodyWidth-len(titleText))
	}

	return gutter + style.Bg(color) + style.Fg(d.colors.surface) + lineBody + style.ANSIReset
}

func (d *DayView) eventForRow(row int, events []CalendarEvent) (CalendarEvent, bool) {
	cellStart := d.date.Add(time.Duration(row) * 30 * time.Minute)
	cellEnd := cellStart.Add(30 * time.Minute)

	for _, ev := range events {
		start, end := clampEventToDay(ev, d.date)
		if end.After(cellStart) && start.Before(cellEnd) {
			return ev, true
		}
	}

	return CalendarEvent{}, false
}

func (d *DayView) blockRuneForRow(ev CalendarEvent, row int) rune {
	cellStart := d.date.Add(time.Duration(row) * 30 * time.Minute)
	cellEnd := cellStart.Add(30 * time.Minute)
	start, end := clampEventToDay(ev, d.date)
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

func (d *DayView) isTitleRow(ev CalendarEvent, row int) bool {
	start, _ := clampEventToDay(ev, d.date)
	titleRow := int(start.Sub(d.date) / (30 * time.Minute))
	if titleRow < 0 {
		titleRow = 0
	}
	if titleRow > 47 {
		titleRow = 47
	}
	return row == titleRow
}

func (d *DayView) isNowRow(row int) bool {
	now := time.Now().In(d.date.Location())
	if !sameCalendarDay(d.date, now) {
		return false
	}
	nowRow := now.Hour()*2 + now.Minute()/30
	return row == nowRow
}

func normalizeCalendarDate(t time.Time) time.Time {
	if t.IsZero() {
		t = time.Now()
	}
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func sameCalendarDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func eventIntersectsDay(ev CalendarEvent, dayStart, dayEnd time.Time) bool {
	start := ev.Start.In(dayStart.Location())
	end := ev.End.In(dayStart.Location())
	if !end.After(start) {
		return false
	}
	return end.After(dayStart) && start.Before(dayEnd)
}

func clampEventToDay(ev CalendarEvent, dayStart time.Time) (time.Time, time.Time) {
	dayEnd := dayStart.Add(24 * time.Hour)
	start := ev.Start.In(dayStart.Location())
	end := ev.End.In(dayStart.Location())
	if start.Before(dayStart) {
		start = dayStart
	}
	if end.After(dayEnd) {
		end = dayEnd
	}
	if end.Before(start) {
		end = start
	}
	return start, end
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func dayViewVisualWidth(s string) int {
	return len(dayViewStripANSI(s))
}

func dayViewStripANSI(s string) string {
	var b strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inEscape {
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inEscape = false
			}
			continue
		}
		if ch == '\x1b' {
			inEscape = true
			continue
		}
		b.WriteByte(ch)
	}
	return b.String()
}

var _ tui.Component = (*DayView)(nil)
