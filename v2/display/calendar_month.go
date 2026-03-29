package display

import (
	"fmt"
	"strings"
	"time"

	design "github.com/SCKelemen/design-system"
	tui "github.com/SCKelemen/tui/v2"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

// MonthView renders a month calendar grid with lightweight event markers.
type MonthView struct {
	events      []CalendarEvent
	focusedDate time.Time // currently selected date
	viewDate    time.Time // month being displayed (first of month)
	width       int
	height      int
	colors      calendarColors
	focused     bool
}

// MonthViewOption configures a MonthView.
type MonthViewOption func(*MonthView)

// WithMonthViewEvents sets the list of events rendered on the month view.
func WithMonthViewEvents(events []CalendarEvent) MonthViewOption {
	return func(v *MonthView) {
		v.events = append([]CalendarEvent(nil), events...)
	}
}

// WithMonthViewDate sets the initial focused date.
func WithMonthViewDate(d time.Time) MonthViewOption {
	return func(v *MonthView) {
		d = normalizeDate(d)
		v.focusedDate = d
		v.viewDate = firstOfMonth(d)
	}
}

// WithMonthViewDesignTokens applies design-system token colors.
func WithMonthViewDesignTokens(dt *design.DesignTokens) MonthViewOption {
	return func(v *MonthView) {
		v.colors = calendarColorsFromTokens(dt)
	}
}

// WithMonthViewWidth sets the rendering width.
func WithMonthViewWidth(width int) MonthViewOption {
	return func(v *MonthView) {
		if width > 0 {
			v.width = width
		}
	}
}

// NewMonthView creates a new month calendar component.
func NewMonthView(opts ...MonthViewOption) *MonthView {
	today := normalizeDate(time.Now())
	v := &MonthView{
		events:      make([]CalendarEvent, 0),
		focusedDate: today,
		viewDate:    firstOfMonth(today),
		width:       98,
		height:      0,
		colors:      defaultCalendarColors(),
	}

	for _, opt := range opts {
		opt(v)
	}

	if v.focusedDate.IsZero() {
		v.focusedDate = today
	}
	v.focusedDate = normalizeDate(v.focusedDate)
	v.viewDate = firstOfMonth(v.viewDate)
	if v.viewDate.IsZero() {
		v.viewDate = firstOfMonth(v.focusedDate)
	}
	if v.width <= 0 {
		v.width = 98
	}

	return v
}

// Init initializes the component.
func (v *MonthView) Init() tea.Cmd {
	return nil
}

// Update handles key and window-size messages.
func (v *MonthView) Update(msg tea.Msg) (tui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			v.width = msg.Width
		}
		if msg.Height > 0 {
			v.height = msg.Height
		}
		return v, nil
	case tea.KeyMsg:
		if !v.focused {
			return v, nil
		}

		switch msg.String() {
		case "h", "left":
			v.viewDate = firstOfMonth(v.viewDate.AddDate(0, -1, 0))
			v.moveFocusedWithinVisibleMonth()
		case "l", "right":
			v.viewDate = firstOfMonth(v.viewDate.AddDate(0, 1, 0))
			v.moveFocusedWithinVisibleMonth()
		case "j", "down":
			v.focusedDate = v.focusedDate.AddDate(0, 0, 7)
			v.viewDate = firstOfMonth(v.focusedDate)
		case "k", "up":
			v.focusedDate = v.focusedDate.AddDate(0, 0, -7)
			v.viewDate = firstOfMonth(v.focusedDate)
		case "t":
			today := normalizeDate(time.Now())
			v.focusedDate = today
			v.viewDate = firstOfMonth(today)
		}
		return v, nil
	}

	return v, nil
}

// View renders the month calendar.
func (v *MonthView) View() string {
	width := v.width
	if width <= 0 {
		width = 98
	}

	const gutterWidth = 8 // 8 vertical separators around 7 columns.
	cellWidth := (width - gutterWidth) / 7
	if cellWidth < 4 {
		cellWidth = 4
	}

	var out strings.Builder
	out.WriteString(v.renderHeaderLine(7*cellWidth + gutterWidth))
	out.WriteString("\n")
	out.WriteString(v.renderWeekdayHeader(cellWidth))
	out.WriteString("\n")

	topBorder, midBorder, bottomBorder := gridBorders(cellWidth)
	out.WriteString(topBorder)
	out.WriteString("\n")

	gridStart := monthGridStart(v.viewDate)
	for week := 0; week < 6; week++ {
		cells := make([][3]string, 7)
		for col := 0; col < 7; col++ {
			day := gridStart.AddDate(0, 0, week*7+col)
			cells[col] = v.renderDayCell(day, cellWidth)
		}

		for line := 0; line < 3; line++ {
			var row strings.Builder
			row.WriteString("│")
			for col := 0; col < 7; col++ {
				row.WriteString(cells[col][line])
				row.WriteString("│")
			}
			out.WriteString(row.String())
			out.WriteString("\n")
		}

		if week == 5 {
			out.WriteString(bottomBorder)
		} else {
			out.WriteString(midBorder)
			out.WriteString("\n")
		}
	}

	return out.String()
}

// Focus marks this component as focused.
func (v *MonthView) Focus() {
	v.focused = true
}

// Blur marks this component as unfocused.
func (v *MonthView) Blur() {
	v.focused = false
}

// Focused reports whether this component currently has focus.
func (v *MonthView) Focused() bool {
	return v.focused
}

func (v *MonthView) renderHeaderLine(width int) string {
	monthLabel := v.viewDate.Format("January 2006")
	hints := "h/← prev  l/→ next  j/↓ +week  k/↑ -week  t today"

	plain := monthLabel
	if width > runeLen(monthLabel)+1 {
		gap := width - runeLen(monthLabel) - runeLen(hints)
		if gap >= 1 {
			plain = monthLabel + strings.Repeat(" ", gap) + hints
		} else {
			plain = truncateRunes(monthLabel+" "+hints, width)
		}
	}

	if runeLen(plain) < width {
		plain += strings.Repeat(" ", width-runeLen(plain))
	}

	monthColor := style.Fg(v.colors.text) + style.ANSIBold
	hintsColor := style.Fg(v.colors.muted) + style.ANSIDim
	reset := style.ANSIReset

	if strings.HasPrefix(plain, monthLabel) {
		left := monthColor + monthLabel + reset
		rest := strings.TrimPrefix(plain, monthLabel)
		if strings.Contains(rest, hints) {
			rest = strings.Replace(rest, hints, hintsColor+hints+reset, 1)
		}
		return left + rest
	}

	return monthColor + plain + reset
}

func (v *MonthView) renderWeekdayHeader(cellWidth int) string {
	labels := []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	var b strings.Builder
	b.WriteString("│")
	for _, label := range labels {
		padded := centerPad(label, cellWidth)
		b.WriteString(style.Fg(v.colors.muted))
		b.WriteString(style.ANSIBold)
		b.WriteString(padded)
		b.WriteString(style.ANSIReset)
		b.WriteString("│")
	}
	return b.String()
}

func (v *MonthView) renderDayCell(day time.Time, cellWidth int) [3]string {
	inMonth := day.Month() == v.viewDate.Month() && day.Year() == v.viewDate.Year()
	isToday := sameDate(day, normalizeDate(time.Now()))
	isSelected := sameDate(day, v.focusedDate)

	dayNumber := fmt.Sprintf("%2d", day.Day())
	dayNumber = leftPad(dayNumber, 2)
	dayLineText := truncateRunes(dayNumber, cellWidth)
	dayLineText = rightPad(dayLineText, cellWidth)

	events := v.eventsForDay(day)
	eventLines := make([]string, 0, 2)
	for i := 0; i < 2; i++ {
		if i >= len(events) {
			eventLines = append(eventLines, strings.Repeat(" ", cellWidth))
			continue
		}

		if i == 1 && len(events) > 2 {
			more := len(events) - 1
			line := fmt.Sprintf("● +%d", more)
			eventLines = append(eventLines, rightPad(truncateRunes(line, cellWidth), cellWidth))
			continue
		}

		e := events[i]
		line := "●"
		if cellWidth >= 5 {
			line = "● " + truncateRunes(strings.TrimSpace(e.Title), cellWidth-2)
		}
		eventLines = append(eventLines, rightPad(truncateRunes(line, cellWidth), cellWidth))
	}

	line0 := v.styleDayLine(dayLineText, inMonth, isToday, isSelected)
	line1 := v.styleEventLine(eventLines[0], inMonth, events, 0)
	line2 := v.styleEventLine(eventLines[1], inMonth, events, 1)

	return [3]string{line0, line1, line2}
}

func (v *MonthView) styleDayLine(text string, inMonth, isToday, isSelected bool) string {
	var b strings.Builder
	if !inMonth {
		b.WriteString(style.ANSIDim)
		b.WriteString(style.Fg(v.colors.muted))
	}
	if isSelected {
		b.WriteString(style.Bg(v.colors.surfaceAlt))
		b.WriteString(style.ANSIBold)
	}
	if isToday {
		b.WriteString(style.Fg(v.colors.today))
		b.WriteString(style.ANSIBold)
	} else if inMonth {
		b.WriteString(style.Fg(v.colors.text))
	}
	b.WriteString(text)
	b.WriteString(style.ANSIReset)
	return b.String()
}

func (v *MonthView) styleEventLine(text string, inMonth bool, events []CalendarEvent, eventIndex int) string {
	var b strings.Builder
	if !inMonth {
		b.WriteString(style.ANSIDim)
		b.WriteString(style.Fg(v.colors.muted))
		b.WriteString(text)
		b.WriteString(style.ANSIReset)
		return b.String()
	}

	if strings.TrimSpace(text) == "" {
		b.WriteString(text)
		return b.String()
	}

	color := v.colors.muted
	if eventIndex < len(events) {
		if strings.TrimSpace(events[eventIndex].Color) != "" {
			color = events[eventIndex].Color
		} else {
			color = defaultEventColors[eventIndex%len(defaultEventColors)]
		}
	}

	b.WriteString(style.Fg(color))
	b.WriteString(text)
	b.WriteString(style.ANSIReset)
	return b.String()
}

func (v *MonthView) eventsForDay(day time.Time) []CalendarEvent {
	dayStart := normalizeDate(day)
	dayEnd := dayStart.Add(24*time.Hour - time.Nanosecond)

	result := make([]CalendarEvent, 0, 4)
	for _, e := range v.events {
		start := e.Start
		end := e.End
		if end.IsZero() {
			end = start
		}
		if end.Before(dayStart) || start.After(dayEnd) {
			continue
		}
		result = append(result, e)
	}
	return result
}

func (v *MonthView) moveFocusedWithinVisibleMonth() {
	y := v.viewDate.Year()
	m := v.viewDate.Month()
	d := v.focusedDate.Day()
	last := daysInMonth(y, m)
	if d > last {
		d = last
	}
	v.focusedDate = time.Date(y, m, d, 0, 0, 0, 0, v.focusedDate.Location())
}

func normalizeDate(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func firstOfMonth(t time.Time) time.Time {
	t = normalizeDate(t)
	if t.IsZero() {
		return t
	}
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func monthGridStart(monthStart time.Time) time.Time {
	monthStart = firstOfMonth(monthStart)
	weekday := int(monthStart.Weekday())
	mondayIndex := (weekday + 6) % 7 // Sunday=0 -> 6, Monday=1 -> 0
	return monthStart.AddDate(0, 0, -mondayIndex)
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
}

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func gridBorders(cellWidth int) (top string, mid string, bottom string) {
	seg := strings.Repeat("─", cellWidth)
	parts := make([]string, 7)
	for i := 0; i < 7; i++ {
		parts[i] = seg
	}
	top = "┌" + strings.Join(parts, "┬") + "┐"
	mid = "├" + strings.Join(parts, "┼") + "┤"
	bottom = "└" + strings.Join(parts, "┴") + "┘"
	return top, mid, bottom
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return string(r[:1])
	}
	return string(r[:max-1]) + "…"
}

func runeLen(s string) int {
	return len([]rune(s))
}

func rightPad(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := runeLen(s)
	if w >= width {
		return truncateRunes(s, width)
	}
	return s + strings.Repeat(" ", width-w)
}

func leftPad(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := runeLen(s)
	if w >= width {
		return truncateRunes(s, width)
	}
	return strings.Repeat(" ", width-w) + s
}

func centerPad(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := runeLen(s)
	if w >= width {
		return truncateRunes(s, width)
	}
	left := (width - w) / 2
	right := width - w - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

var _ tui.Component = (*MonthView)(nil)
