package display

import (
	"strings"
	"testing"
	"time"
)

func TestDayViewConstructor(t *testing.T) {
	d := time.Date(2025, time.January, 15, 12, 0, 0, 0, time.UTC)
	v := NewDayView(
		WithDayViewDate(d),
		WithDayViewWidth(88),
		WithDayViewHeight(14),
	)
	if v == nil {
		t.Fatal("NewDayView returned nil")
	}
	if v.width != 88 {
		t.Fatalf("expected width=88, got %d", v.width)
	}
	if v.height != 14 {
		t.Fatalf("expected height=14, got %d", v.height)
	}
	if !sameCalendarDay(v.date, time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected normalized date: %v", v.date)
	}
}

func TestDayViewViewRendersTimeSlots(t *testing.T) {
	v := NewDayView(
		WithDayViewDate(time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC)),
		WithDayViewWidth(88),
		WithDayViewHeight(12),
	)

	view := dayViewStripANSI(v.View())
	if strings.TrimSpace(view) == "" {
		t.Fatal("expected non-empty day view")
	}
	if !strings.Contains(view, "Wednesday, January 15, 2025") {
		t.Fatalf("expected date header, got:\n%s", view)
	}
	if !strings.Contains(view, "All-day:") {
		t.Fatalf("expected all-day row, got:\n%s", view)
	}
	if !strings.Contains(view, "00:00 │") || !strings.Contains(view, "01:00 │") {
		t.Fatalf("expected hourly time slots, got:\n%s", view)
	}
}
