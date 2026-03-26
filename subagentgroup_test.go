package tui

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSubagentGroupCreation(t *testing.T) {
	g := NewSubagentGroup(
		WithSubagentGroupGap(3),
		WithSubagentGroupFocusHint("Ctrl+O to focus"),
	)
	if g == nil {
		t.Fatal("NewSubagentGroup returned nil")
	}

	if g.gap != 3 {
		t.Fatalf("expected gap=3, got %d", g.gap)
	}
	if g.focusHint != "Ctrl+O to focus" {
		t.Fatalf("expected focus hint to be set, got %q", g.focusHint)
	}
	if g.PanelCount() != 0 {
		t.Fatalf("expected empty group, got %d panels", g.PanelCount())
	}
}

func TestSubagentGroupAddPanels(t *testing.T) {
	g := NewSubagentGroup()
	p1 := NewSubagentPanel(WithSubagentTitle("One"))
	p2 := NewSubagentPanel(WithSubagentTitle("Two"))

	g.AddPanel(p1)
	g.AddPanel(p2)

	if g.PanelCount() != 2 {
		t.Fatalf("expected 2 panels, got %d", g.PanelCount())
	}
	if len(g.GetPanels()) != 2 {
		t.Fatalf("GetPanels should return 2 panels")
	}

	g.SetPanels([]*SubagentPanel{p2})
	if g.PanelCount() != 1 {
		t.Fatalf("expected 1 panel after SetPanels, got %d", g.PanelCount())
	}
}

func TestSubagentGroupEqualHeightPadding(t *testing.T) {
	g := NewSubagentGroup(WithSubagentGroupGap(1))
	g.width = 80

	p1 := NewSubagentPanel(WithSubagentTitle("A"), WithSubagentVisibleTools(10))
	p1.SetTools([]SubagentTool{{Name: "t1", Status: ToolCompleted}})

	p2 := NewSubagentPanel(WithSubagentTitle("B"), WithSubagentVisibleTools(10))
	p2.SetTools([]SubagentTool{
		{Name: "t1", Status: ToolCompleted},
		{Name: "t2", Status: ToolCompleted},
		{Name: "t3", Status: ToolCompleted},
	})

	g.SetPanels([]*SubagentPanel{p1, p2})
	lines := g.renderPanelRowLines()
	if len(lines) == 0 {
		t.Fatal("expected rendered panel row lines")
	}

	widths := g.computePanelWidths()
	p1.width = widths[0]
	p2.width = widths[1]
	p1Lines := strings.Split(strings.TrimSuffix(p1.View(), "\n"), "\n")
	p2Lines := strings.Split(strings.TrimSuffix(p2.View(), "\n"), "\n")
	expectedHeight := len(p1Lines)
	if len(p2Lines) > expectedHeight {
		expectedHeight = len(p2Lines)
	}

	if len(lines) != expectedHeight {
		t.Fatalf("expected row height %d (max panel height), got %d", expectedHeight, len(lines))
	}

	firstWidth := utf8.RuneCountInString(stripANSI(lines[0]))
	if firstWidth != 80 {
		t.Fatalf("expected first merged row width 80, got %d: %q", firstWidth, stripANSI(lines[0]))
	}
}

func TestSubagentGroupHeaderTextGeneration(t *testing.T) {
	g := NewSubagentGroup(WithSubagentGroupFocusHint("Ctrl+O to focus"))
	p1 := NewSubagentPanel(WithSubagentTitle("A"))
	p2 := NewSubagentPanel(WithSubagentTitle("B"))
	p1.SetStatus(SubagentCompleted)
	p2.SetStatus(SubagentRunning)
	g.SetPanels([]*SubagentPanel{p1, p2})

	head := g.renderHeaderText()
	if !strings.Contains(head, "Running 2 subagents... (1/2 completed)") {
		t.Fatalf("unexpected header text: %q", head)
	}
	if !strings.Contains(head, "Ctrl+O to focus") {
		t.Fatalf("expected focus hint in header: %q", head)
	}

	p2.SetStatus(SubagentCompleted)
	head = g.renderHeaderText()
	if !strings.Contains(head, "Ran 2 subagents ✓") {
		t.Fatalf("expected completed header, got: %q", head)
	}
}

func TestSubagentGroupFooterWithProgressBar(t *testing.T) {
	g := NewSubagentGroup()
	g.startTime = time.Now().Add(-9 * time.Second)

	p1 := NewSubagentPanel(WithSubagentTitle("A"))
	p2 := NewSubagentPanel(WithSubagentTitle("B"))
	p3 := NewSubagentPanel(WithSubagentTitle("C"))
	p1.SetStatus(SubagentCompleted)
	p2.SetStatus(SubagentCompleted)
	p3.SetStatus(SubagentCompleted)
	g.SetPanels([]*SubagentPanel{p1, p2, p3})
	g.width = 120

	view := stripANSI(g.View())
	if !strings.Contains(view, "Finished running 3 subagents") {
		t.Fatalf("expected footer in view, got:\n%s", view)
	}
	if !strings.Contains(view, "▰▰▰") {
		t.Fatalf("expected block progress bar in footer, got:\n%s", view)
	}
}

func TestSubagentGroupWidthDistribution(t *testing.T) {
	g := NewSubagentGroup(WithSubagentGroupGap(2))
	g.width = 32
	g.SetPanels([]*SubagentPanel{
		NewSubagentPanel(),
		NewSubagentPanel(),
		NewSubagentPanel(),
	})

	widths := g.computePanelWidths()
	if len(widths) != 3 {
		t.Fatalf("expected 3 widths, got %d", len(widths))
	}

	sum := 0
	for _, w := range widths {
		sum += w
	}
	total := sum + g.gap*(len(widths)-1)
	if total != g.width {
		t.Fatalf("expected distributed width to match group width, got %d != %d", total, g.width)
	}
}

func TestSubagentGroupFocusCycling(t *testing.T) {
	g := NewSubagentGroup()
	p1 := NewSubagentPanel(WithSubagentTitle("A"))
	p2 := NewSubagentPanel(WithSubagentTitle("B"))
	p3 := NewSubagentPanel(WithSubagentTitle("C"))
	g.SetPanels([]*SubagentPanel{p1, p2, p3})

	g.Focus()
	if !p1.Focused() {
		t.Fatal("expected first panel focused initially")
	}

	g.Update(tea.KeyMsg{Type: tea.KeyTab})
	if g.focusedPanel != 1 || !p2.Focused() {
		t.Fatalf("expected panel 2 focused, index=%d", g.focusedPanel)
	}

	g.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if g.focusedPanel != 0 || !p1.Focused() {
		t.Fatalf("expected panel 1 focused after shift+tab, index=%d", g.focusedPanel)
	}

	g.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	if g.Focused() {
		t.Fatal("expected group focus toggled off by ctrl+o")
	}
}

func TestSubagentGroupCompletionTracking(t *testing.T) {
	g := NewSubagentGroup()
	p1 := NewSubagentPanel()
	p2 := NewSubagentPanel()
	p3 := NewSubagentPanel()
	p1.SetStatus(SubagentCompleted)
	p2.SetStatus(SubagentRunning)
	p3.SetStatus(SubagentCompleted)
	g.SetPanels([]*SubagentPanel{p1, p2, p3})

	if got := g.CompletedCount(); got != 2 {
		t.Fatalf("expected completed count=2, got %d", got)
	}
	if g.IsAllCompleted() {
		t.Fatal("expected IsAllCompleted=false with one running panel")
	}

	p2.SetStatus(SubagentCompleted)
	if !g.IsAllCompleted() {
		t.Fatal("expected IsAllCompleted=true after all completed")
	}
}

func TestSubagentGroupUpdateWindowSize(t *testing.T) {
	g := NewSubagentGroup()
	g.Update(tea.WindowSizeMsg{Width: 95, Height: 40})
	if g.width != 95 {
		t.Fatalf("expected width updated to 95, got %d", g.width)
	}
}
