package display

import (
	design "github.com/SCKelemen/design-system"
	"testing"
	"time"
)

func TestCalendarTypeCalendarEventFields(t *testing.T) {
	start := time.Date(2025, time.January, 20, 9, 0, 0, 0, time.UTC)
	end := time.Date(2025, time.January, 20, 10, 30, 0, 0, time.UTC)

	ev := CalendarEvent{
		Title:    "Team sync",
		Start:    start,
		End:      end,
		Color:    "#61AFEF",
		AllDay:   false,
		Location: "Room 1",
	}

	if ev.Title != "Team sync" {
		t.Fatalf("expected Title=%q, got %q", "Team sync", ev.Title)
	}
	if !ev.Start.Equal(start) {
		t.Fatalf("expected Start=%v, got %v", start, ev.Start)
	}
	if !ev.End.Equal(end) {
		t.Fatalf("expected End=%v, got %v", end, ev.End)
	}
	if ev.Color != "#61AFEF" {
		t.Fatalf("expected Color=%q, got %q", "#61AFEF", ev.Color)
	}
	if ev.AllDay {
		t.Fatal("expected AllDay=false")
	}
}

func TestCalendarTypeCalendarEventZeroValue(t *testing.T) {
	var ev CalendarEvent

	if ev.Title != "" {
		t.Fatalf("expected empty Title, got %q", ev.Title)
	}
	if !ev.Start.IsZero() {
		t.Fatalf("expected zero Start, got %v", ev.Start)
	}
	if !ev.End.IsZero() {
		t.Fatalf("expected zero End, got %v", ev.End)
	}
	if ev.Color != "" {
		t.Fatalf("expected empty Color, got %q", ev.Color)
	}
	if ev.AllDay {
		t.Fatal("expected AllDay=false")
	}
}

func TestCalendarTypeCalendarTaskFields(t *testing.T) {
	due := time.Date(2025, time.February, 1, 12, 0, 0, 0, time.UTC)
	task := CalendarTask{
		Title:     "Review policy",
		DueDate:   due,
		Completed: true,
		Color:     "#98C379",
		Priority:  TaskPriorityHigh,
	}

	if task.Title != "Review policy" {
		t.Fatalf("expected Title=%q, got %q", "Review policy", task.Title)
	}
	if !task.DueDate.Equal(due) {
		t.Fatalf("expected DueDate=%v, got %v", due, task.DueDate)
	}
	if !task.Completed {
		t.Fatal("expected Completed=true")
	}
	if task.Color != "#98C379" {
		t.Fatalf("expected Color=%q, got %q", "#98C379", task.Color)
	}
	if task.Priority != TaskPriorityHigh {
		t.Fatalf("expected Priority=%v, got %v", TaskPriorityHigh, task.Priority)
	}
}

func TestCalendarTypeCalendarTaskZeroValue(t *testing.T) {
	var task CalendarTask

	if task.Title != "" {
		t.Fatalf("expected empty Title, got %q", task.Title)
	}
	if !task.DueDate.IsZero() {
		t.Fatalf("expected zero DueDate, got %v", task.DueDate)
	}
	if task.Completed {
		t.Fatal("expected Completed=false")
	}
	if task.Color != "" {
		t.Fatalf("expected empty Color, got %q", task.Color)
	}
	if task.Priority != TaskPriorityNone {
		t.Fatalf("expected Priority=%v, got %v", TaskPriorityNone, task.Priority)
	}
}

func TestCalendarTypePriorityConstants(t *testing.T) {
	if TaskPriorityLow != 1 {
		t.Fatalf("expected TaskPriorityLow=1, got %d", TaskPriorityLow)
	}
	if TaskPriorityMedium != 2 {
		t.Fatalf("expected TaskPriorityMedium=2, got %d", TaskPriorityMedium)
	}
	if TaskPriorityHigh != 3 {
		t.Fatalf("expected TaskPriorityHigh=3, got %d", TaskPriorityHigh)
	}
}

func TestCalendarTypeDefaultCalendarColors(t *testing.T) {
	c := defaultCalendarColors()

	if c.surface == "" || c.surfaceAlt == "" || c.border == "" || c.accent == "" {
		t.Fatalf("expected default colors to be populated, got %+v", c)
	}
	if c.text == "" || c.today == "" || c.error == "" || c.success == "" || c.warning == "" {
		t.Fatalf("expected default colors to be populated, got %+v", c)
	}
}

func TestCalendarTypeCalendarColorsFromTokensNilUsesDefaults(t *testing.T) {
	got := calendarColorsFromTokens(nil)
	want := defaultCalendarColors()

	if got != want {
		t.Fatalf("expected default colors for nil tokens, got %+v want %+v", got, want)
	}
}

func TestCalendarTypeCalendarColorsFromTokensOverridesNonEmpty(t *testing.T) {
	tokens := &design.DesignTokens{
		SurfaceBase:   "#111111",
		SurfaceRaised: "#222222",
		BorderSubtle:  "#333333",
		Accent:        "#444444",
		MutedColor:    "#555555",
		Color:         "#666666",
		ErrorBright:   "#777777",
		SuccessBright: "#888888",
		PendingColor:  "#999999",
	}

	got := calendarColorsFromTokens(tokens)

	if got.surface != "#111111" || got.surfaceAlt != "#222222" || got.border != "#333333" {
		t.Fatalf("unexpected surface/border colors: %+v", got)
	}
	if got.accent != "#444444" || got.today != "#444444" {
		t.Fatalf("unexpected accent/today colors: %+v", got)
	}
	if got.muted != "#555555" || got.text != "#666666" {
		t.Fatalf("unexpected text colors: %+v", got)
	}
	if got.error != "#777777" || got.success != "#888888" || got.warning != "#999999" {
		t.Fatalf("unexpected status colors: %+v", got)
	}
}

func TestCalendarTypeCalendarColorsFromTokensEmptyDoesNotOverride(t *testing.T) {
	tokens := &design.DesignTokens{}
	got := calendarColorsFromTokens(tokens)
	want := defaultCalendarColors()

	if got != want {
		t.Fatalf("expected defaults with empty tokens, got %+v want %+v", got, want)
	}
}

func TestCalendarTypeDefaultEventColors(t *testing.T) {
	if len(defaultEventColors) == 0 {
		t.Fatal("expected defaultEventColors to be non-empty")
	}

	for i, c := range defaultEventColors {
		if c == "" {
			t.Fatalf("expected defaultEventColors[%d] to be non-empty", i)
		}
	}
}
