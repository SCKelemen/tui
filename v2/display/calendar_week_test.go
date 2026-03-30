package display

import (
	"strings"
	"testing"
	"time"
)

func TestWeekViewConstructor(t *testing.T) {
	d := time.Date(2025, time.January, 15, 18, 0, 0, 0, time.UTC)
	v := NewWeekView(
		WithWeekViewDate(d),
		WithWeekViewWidth(98),
		WithWeekViewHeight(14),
	)
	if v == nil {
		t.Fatal("NewWeekView returned nil")
	}
	if v.width != 98 {
		t.Fatalf("expected width=98, got %d", v.width)
	}
	if v.height != 14 {
		t.Fatalf("expected height=14, got %d", v.height)
	}
	if !sameCalendarDay(v.date, time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected normalized date: %v", v.date)
	}
}

func TestWeekViewViewRendersWeekGrid(t *testing.T) {
	v := NewWeekView(
		WithWeekViewDate(time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC)),
		WithWeekViewWidth(98),
		WithWeekViewHeight(12),
	)

	view := dayViewStripANSI(v.View())
	if strings.TrimSpace(view) == "" {
		t.Fatal("expected non-empty week view")
	}
	if !strings.Contains(view, "Jan 13 – 19, 2025") {
		t.Fatalf("expected week range header, got:\n%s", view)
	}
	if !strings.Contains(view, "Mon 13") || !strings.Contains(view, "Sun 19") {
		t.Fatalf("expected day columns, got:\n%s", view)
	}
	if !strings.Contains(view, "Allday│") {
		t.Fatalf("expected all-day row, got:\n%s", view)
	}
	if !strings.Contains(view, "00:00 │") {
		t.Fatalf("expected timeline gutter, got:\n%s", view)
	}
}
