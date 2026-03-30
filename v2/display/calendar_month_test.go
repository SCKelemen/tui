package display

import (
	"strings"
	"testing"
	"time"
)

func TestMonthViewConstructor(t *testing.T) {
	d := time.Date(2025, time.January, 15, 14, 30, 0, 0, time.UTC)
	v := NewMonthView(WithMonthViewDate(d), WithMonthViewWidth(84))
	if v == nil {
		t.Fatal("NewMonthView returned nil")
	}

	if v.width != 84 {
		t.Fatalf("expected width=84, got %d", v.width)
	}
	if !sameDate(v.focusedDate, time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected focusedDate: %v", v.focusedDate)
	}
	if !sameDate(v.viewDate, time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected viewDate: %v", v.viewDate)
	}
}

func TestMonthViewViewRendersMonthGridWithDayNumbers(t *testing.T) {
	v := NewMonthView(
		WithMonthViewDate(time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC)),
		WithMonthViewWidth(84),
	)

	view := stripANSI(v.View())
	if strings.TrimSpace(view) == "" {
		t.Fatal("expected non-empty month view")
	}
	if !strings.Contains(view, "January 2025") {
		t.Fatalf("expected month header, got:\n%s", view)
	}
	if !strings.Contains(view, "Mo") || !strings.Contains(view, "Su") {
		t.Fatalf("expected weekday header, got:\n%s", view)
	}
	if !strings.Contains(view, "┌") || !strings.Contains(view, "└") {
		t.Fatalf("expected grid borders, got:\n%s", view)
	}
	if !strings.Contains(view, " 1") || !strings.Contains(view, "31") {
		t.Fatalf("expected day numbers in grid, got:\n%s", view)
	}
}
