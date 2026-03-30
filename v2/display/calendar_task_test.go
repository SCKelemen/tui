package display

import (
	"strings"
	"testing"
	"time"
)

func TestTaskViewConstructorWithTasks(t *testing.T) {
	today := normalizeCalendarDate(time.Now())
	tasks := []CalendarTask{
		{Title: "Review PR", DueDate: today, Priority: TaskPriorityHigh},
		{Title: "Write docs", DueDate: today.AddDate(0, 0, 1), Priority: TaskPriorityLow},
	}

	v := NewTaskView(WithTaskViewTasks(tasks), WithTaskViewWidth(82))
	if v == nil {
		t.Fatal("NewTaskView returned nil")
	}
	if v.width != 82 {
		t.Fatalf("expected width=82, got %d", v.width)
	}
	if len(v.tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(v.tasks))
	}
	if v.showCompleted {
		t.Fatal("expected showCompleted=false by default")
	}
}

func TestTaskViewViewRendersTaskList(t *testing.T) {
	today := normalizeCalendarDate(time.Now())
	v := NewTaskView(WithTaskViewTasks([]CalendarTask{
		{Title: "Review PR", DueDate: today, Priority: TaskPriorityHigh},
		{Title: "Write docs", DueDate: today.AddDate(0, 0, 1), Priority: TaskPriorityMedium},
	}))

	view := stripANSI(v.View())
	if strings.TrimSpace(view) == "" {
		t.Fatal("expected non-empty task view")
	}
	if !strings.Contains(view, "Tasks") {
		t.Fatalf("expected Tasks header, got:\n%s", view)
	}
	if !strings.Contains(view, "Review PR") || !strings.Contains(view, "Write docs") {
		t.Fatalf("expected task titles in rendered list, got:\n%s", view)
	}
	if !strings.Contains(view, "☐") {
		t.Fatalf("expected checkbox rendering, got:\n%s", view)
	}
}
