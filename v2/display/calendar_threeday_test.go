package display

import (
	"strings"
	"testing"
	"time"

	design "github.com/SCKelemen/design-system"
	tea "github.com/charmbracelet/bubbletea"
)

func sampleThreeDayEvents() []CalendarEvent {
	loc := time.UTC
	return []CalendarEvent{
		{
			Title:  "All Hands",
			Start:  time.Date(2025, time.January, 15, 0, 0, 0, 0, loc),
			End:    time.Date(2025, time.January, 16, 0, 0, 0, 0, loc),
			AllDay: true,
		},
		{
			Title:    "Security Review",
			Start:    time.Date(2025, time.January, 15, 10, 0, 0, 0, loc),
			End:      time.Date(2025, time.January, 15, 11, 0, 0, 0, loc),
			Location: "Room A",
		},
		{
			Title: "Pairing Session",
			Start: time.Date(2025, time.January, 16, 14, 0, 0, 0, loc),
			End:   time.Date(2025, time.January, 16, 15, 30, 0, 0, loc),
		},
	}
}

func TestThreeDayViewConstructor(t *testing.T) {
	d := time.Date(2025, time.January, 15, 12, 0, 0, 0, time.UTC)
	events := sampleThreeDayEvents()

	v := NewThreeDayView(
		WithThreeDayViewDate(d),
		WithThreeDayViewEvents(events),
		WithThreeDayViewWidth(96),
		WithThreeDayViewHeight(18),
	)
	if v == nil {
		t.Fatal("NewThreeDayView returned nil")
	}
	if v.width != 96 {
		t.Fatalf("expected width=96, got %d", v.width)
	}
	if v.height != 18 {
		t.Fatalf("expected height=18, got %d", v.height)
	}
	if len(v.events) != len(events) {
		t.Fatalf("expected %d events, got %d", len(events), len(v.events))
	}
	if !sameCalendarDay(v.date, time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected normalized date: %v", v.date)
	}
}

func TestThreeDayViewRendersTimelineAndEvents(t *testing.T) {
	v := NewThreeDayView(
		WithThreeDayViewDate(time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC)),
		WithThreeDayViewEvents(sampleThreeDayEvents()),
		WithThreeDayViewWidth(100),
		WithThreeDayViewHeight(16),
	)

	view := dayViewStripANSI(v.View())
	if strings.TrimSpace(view) == "" {
		t.Fatal("expected non-empty three-day view")
	}
	if !strings.Contains(view, "Jan 15") || !strings.Contains(view, "Jan 17, 2025") {
		t.Fatalf("expected range header, got:\n%s", view)
	}
	if !strings.Contains(view, "Wed Jan 15") || !strings.Contains(view, "Thu Jan 16") || !strings.Contains(view, "Fri Jan 17") {
		t.Fatalf("expected day column headers, got:\n%s", view)
	}
	if !strings.Contains(view, "09:00") || !strings.Contains(view, "10:00") {
		t.Fatalf("expected hourly gutter labels, got:\n%s", view)
	}
	if !strings.Contains(view, "All Hands") || !strings.Contains(view, "Security Review") || !strings.Contains(view, "Pairing Session") {
		t.Fatalf("expected events in rendered output, got:\n%s", view)
	}
}

func TestThreeDayViewWindowSizeMsgUpdatesDimensions(t *testing.T) {
	v := NewThreeDayView(
		WithThreeDayViewWidth(80),
		WithThreeDayViewHeight(10),
	)

	updated, _ := v.Update(tea.WindowSizeMsg{Width: 132, Height: 27})
	next, ok := updated.(*ThreeDayView)
	if !ok {
		t.Fatalf("expected *ThreeDayView from Update, got %T", updated)
	}
	if next.width != 132 {
		t.Fatalf("expected width=132 after resize, got %d", next.width)
	}
	if next.height != 27 {
		t.Fatalf("expected height=27 after resize, got %d", next.height)
	}
}

func TestThreeDayViewDesignTokensOptionApplies(t *testing.T) {
	tokens := &design.DesignTokens{
		Accent:       "#112233",
		SurfaceBase:  "#010101",
		BorderSubtle: "#abcdef",
		Color:        "#fefefe",
	}

	v := NewThreeDayView(WithThreeDayViewDesignTokens(tokens))

	if v.colors.accent != "#112233" {
		t.Fatalf("expected accent color from tokens, got %q", v.colors.accent)
	}
	if v.colors.surface != "#010101" {
		t.Fatalf("expected surface color from tokens, got %q", v.colors.surface)
	}
	if v.colors.border != "#abcdef" {
		t.Fatalf("expected border color from tokens, got %q", v.colors.border)
	}
	if v.colors.text != "#fefefe" {
		t.Fatalf("expected text color from tokens, got %q", v.colors.text)
	}
}
