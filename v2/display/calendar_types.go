package display

import (
	design "github.com/SCKelemen/design-system"
	"time"
)

// CalendarEvent represents an event on the calendar.
type CalendarEvent struct {
	Title    string
	Start    time.Time
	End      time.Time
	Color    string // hex color for the event chip
	AllDay   bool
	Location string
}

// CalendarTask represents a task with a due date.
type CalendarTask struct {
	Title     string
	DueDate   time.Time
	Completed bool
	Color     string
	Priority  TaskPriority
}

type TaskPriority int

const (
	TaskPriorityNone TaskPriority = iota
	TaskPriorityLow
	TaskPriorityMedium
	TaskPriorityHigh
)

// CalendarViewType identifies the active view.
type CalendarViewType int

const (
	CalendarViewMonth CalendarViewType = iota
	CalendarViewWeek
	CalendarViewThreeDay
	CalendarViewDay
	CalendarViewTask
)

// calendarColors holds design-token-derived colors for calendar rendering.
type calendarColors struct {
	surface    string
	surfaceAlt string
	border     string
	accent     string
	muted      string
	text       string
	today      string
	error      string
	success    string
	warning    string
}

func defaultCalendarColors() calendarColors {
	return calendarColors{
		surface:    "#282C34",
		surfaceAlt: "#31353D",
		border:     "#3C414B",
		accent:     "#61AFEF",
		muted:      "#7A818A",
		text:       "#ABB2BF",
		today:      "#61AFEF",
		error:      "#E06C75",
		success:    "#98C379",
		warning:    "#E5C07B",
	}
}

func calendarColorsFromTokens(dt *design.DesignTokens) calendarColors {
	c := defaultCalendarColors()
	if dt == nil {
		return c
	}
	if v := dt.SurfaceBase; v != "" {
		c.surface = v
	}
	if v := dt.SurfaceRaised; v != "" {
		c.surfaceAlt = v
	}
	if v := dt.BorderSubtle; v != "" {
		c.border = v
	}
	if v := dt.Accent; v != "" {
		c.accent = v
	}
	if v := dt.MutedColor; v != "" {
		c.muted = v
	}
	if v := dt.Color; v != "" {
		c.text = v
	}
	if v := dt.Accent; v != "" {
		c.today = v
	}
	if v := dt.ErrorBright; v != "" {
		c.error = v
	}
	if v := dt.SuccessBright; v != "" {
		c.success = v
	}
	if v := dt.PendingColor; v != "" {
		c.warning = v
	}
	return c
}

// eventColors for differentiating events when no custom color is set.
var defaultEventColors = []string{
	"#61AFEF", // blue
	"#98C379", // green
	"#C678DD", // purple
	"#E5C07B", // yellow
	"#E06C75", // red
	"#56B6C2", // cyan
	"#D19A66", // orange
}
