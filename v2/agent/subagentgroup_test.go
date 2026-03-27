package agent

import (
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
	design "github.com/SCKelemen/design-system"
	"github.com/SCKelemen/tui/v2/style"
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

	widths := g.computeRowWidths(len(g.panels))
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

func splitRenderedColumns(lines []string, widths []int, gap int) [][]string {
	columns := make([][]string, len(widths))

	for _, line := range lines {
		plainRunes := []rune(stripANSI(line))
		start := 0

		for i, width := range widths {
			end := start + width
			if start > len(plainRunes) {
				start = end + gap
				columns[i] = append(columns[i], "")
				continue
			}
			if end > len(plainRunes) {
				end = len(plainRunes)
			}

			columns[i] = append(columns[i], string(plainRunes[start:end]))
			start = end + gap
		}
	}

	return columns
}

func assertAllColumnsSameHeight(t *testing.T, columns [][]string) int {
	t.Helper()

	if len(columns) == 0 {
		t.Fatal("expected at least one column")
	}

	height := len(columns[0])
	for i := 1; i < len(columns); i++ {
		if len(columns[i]) != height {
			t.Fatalf("expected column %d height %d, got %d", i, height, len(columns[i]))
		}
	}

	return height
}

func TestSubagentGroupThreePanelEqualHeight(t *testing.T) {
	g := NewSubagentGroup(WithSubagentGroupGap(1))
	g.width = 120

	p0 := NewSubagentPanel(WithSubagentTitle("Zero"), WithSubagentVisibleTools(10))
	p0.SetTools(nil)

	p2 := NewSubagentPanel(WithSubagentTitle("Two"), WithSubagentVisibleTools(10))
	p2.SetTools([]SubagentTool{
		{Name: "t1", Status: ToolCompleted},
		{Name: "t2", Status: ToolCompleted},
	})

	p5 := NewSubagentPanel(WithSubagentTitle("Five"), WithSubagentVisibleTools(10))
	p5.SetTools([]SubagentTool{
		{Name: "t1", Status: ToolCompleted},
		{Name: "t2", Status: ToolCompleted},
		{Name: "t3", Status: ToolCompleted},
		{Name: "t4", Status: ToolCompleted},
		{Name: "t5", Status: ToolCompleted},
	})

	panels := []*SubagentPanel{p0, p2, p5}
	g.SetPanels(panels)

	lines := g.renderPanelRowLines()
	if len(lines) == 0 {
		t.Fatal("expected rendered panel row lines")
	}

	widths := g.computeRowWidths(len(g.panels))
	columns := splitRenderedColumns(lines, widths, g.gap)
	height := assertAllColumnsSameHeight(t, columns)

	tallest := 0
	for i, panel := range panels {
		panel.width = widths[i]
		panelLines := strings.Split(strings.TrimSuffix(panel.View(), "\n"), "\n")
		if len(panelLines) > tallest {
			tallest = len(panelLines)
		}
	}

	if height != tallest {
		t.Fatalf("expected equalized height to match tallest panel (%d), got %d", tallest, height)
	}
}

func TestSubagentGroupEmptyVsFullPanelHeight(t *testing.T) {
	g := NewSubagentGroup(WithSubagentGroupGap(1))
	g.width = 80

	empty := NewSubagentPanel(WithSubagentTitle("Empty"), WithSubagentVisibleTools(10))
	empty.SetTools(nil)

	full := NewSubagentPanel(WithSubagentTitle("Full"), WithSubagentVisibleTools(10))
	full.SetTools([]SubagentTool{
		{Name: "t1", Status: ToolCompleted},
		{Name: "t2", Status: ToolCompleted},
		{Name: "t3", Status: ToolCompleted},
		{Name: "t4", Status: ToolCompleted},
	})

	g.SetPanels([]*SubagentPanel{empty, full})
	lines := g.renderPanelRowLines()
	if len(lines) == 0 {
		t.Fatal("expected rendered panel row lines")
	}

	widths := g.computeRowWidths(len(g.panels))
	columns := splitRenderedColumns(lines, widths, g.gap)
	_ = assertAllColumnsSameHeight(t, columns)
}

func TestSubagentGroupHeightAfterDynamicUpdate(t *testing.T) {
	g := NewSubagentGroup(WithSubagentGroupGap(1))
	g.width = 90

	p1 := NewSubagentPanel(WithSubagentTitle("A"), WithSubagentVisibleTools(10))
	p1.SetTools([]SubagentTool{{Name: "t1", Status: ToolCompleted}})
	p2 := NewSubagentPanel(WithSubagentTitle("B"), WithSubagentVisibleTools(10))
	p2.SetTools([]SubagentTool{{Name: "t1", Status: ToolCompleted}})
	g.SetPanels([]*SubagentPanel{p1, p2})

	before := g.renderPanelRowLines()
	beforeWidths := g.computeRowWidths(len(g.panels))
	beforeColumns := splitRenderedColumns(before, beforeWidths, g.gap)
	beforeHeight := assertAllColumnsSameHeight(t, beforeColumns)

	p2.SetTools([]SubagentTool{
		{Name: "t1", Status: ToolCompleted},
		{Name: "t2", Status: ToolCompleted},
		{Name: "t3", Status: ToolCompleted},
		{Name: "t4", Status: ToolCompleted},
		{Name: "t5", Status: ToolCompleted},
	})

	after := g.renderPanelRowLines()
	afterWidths := g.computeRowWidths(len(g.panels))
	afterColumns := splitRenderedColumns(after, afterWidths, g.gap)
	afterHeight := assertAllColumnsSameHeight(t, afterColumns)

	if afterHeight <= beforeHeight {
		t.Fatalf("expected updated height to grow after adding tools, before=%d after=%d", beforeHeight, afterHeight)
	}
}

func TestSubagentGroupEqualHeightAcrossWidths(t *testing.T) {
	g := NewSubagentGroup(WithSubagentGroupGap(1))

	p1 := NewSubagentPanel(WithSubagentTitle("A"), WithSubagentVisibleTools(10))
	p1.SetTools([]SubagentTool{{Name: "t1", Status: ToolCompleted}})
	p2 := NewSubagentPanel(WithSubagentTitle("B"), WithSubagentVisibleTools(10))
	p2.SetTools([]SubagentTool{
		{Name: "t1", Status: ToolCompleted},
		{Name: "t2", Status: ToolCompleted},
		{Name: "t3", Status: ToolCompleted},
		{Name: "t4", Status: ToolCompleted},
	})
	g.SetPanels([]*SubagentPanel{p1, p2})

	for _, width := range []int{60, 80, 120} {
		t.Run("width-"+strconv.Itoa(width), func(t *testing.T) {
			g.width = width
			lines := g.renderPanelRowLines()
			if len(lines) == 0 {
				t.Fatal("expected rendered panel row lines")
			}

			widths := g.computeRowWidths(len(g.panels))
			columns := splitRenderedColumns(lines, widths, g.gap)
			_ = assertAllColumnsSameHeight(t, columns)
		})
	}
}

func TestSubagentGroupSinglePanelHeight(t *testing.T) {
	g := NewSubagentGroup(WithSubagentGroupGap(1))
	g.width = 70

	p := NewSubagentPanel(WithSubagentTitle("Solo"), WithSubagentVisibleTools(10))
	p.SetTools([]SubagentTool{
		{Name: "t1", Status: ToolCompleted},
		{Name: "t2", Status: ToolCompleted},
		{Name: "t3", Status: ToolCompleted},
	})
	g.SetPanels([]*SubagentPanel{p})

	lines := g.renderPanelRowLines()
	if len(lines) == 0 {
		t.Fatal("expected rendered lines for single-panel group")
	}

	widths := g.computeRowWidths(len(g.panels))
	columns := splitRenderedColumns(lines, widths, g.gap)
	height := assertAllColumnsSameHeight(t, columns)
	if height == 0 {
		t.Fatal("expected non-zero height for single-panel group")
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

	widths := g.computeRowWidths(len(g.panels))
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

func TestSubagentGroupOptionsInitAndFocusPanel(t *testing.T) {
	tokens := &design.DesignTokens{Theme: "custom"}
	g := NewSubagentGroup(
		WithSubagentGroupGap(-2),
		WithSubagentGroupDesignTokens(tokens),
		WithSubagentGroupTheme("midnight"),
	)

	if g.gap != 0 {
		t.Fatalf("expected negative gap to clamp to 0, got %d", g.gap)
	}
	if g.designTokens == nil {
		t.Fatal("expected design tokens to be set")
	}

	p1 := NewSubagentPanel(WithSubagentTitle("A"))
	p2 := NewSubagentPanel(WithSubagentTitle("B"))
	g.SetPanels([]*SubagentPanel{p1, p2})

	g.FocusPanel(1)
	if g.focusedPanel != 1 {
		t.Fatalf("expected focusedPanel=1, got %d", g.focusedPanel)
	}
	g.FocusPanel(-1)
	g.FocusPanel(99)
	if g.focusedPanel != 1 {
		t.Fatalf("expected invalid FocusPanel to be ignored, got %d", g.focusedPanel)
	}

	if cmd := g.Init(); cmd == nil {
		t.Fatal("expected non-nil init command batch")
	}

	g.Blur()
	if g.Focused() {
		t.Fatal("expected group to blur")
	}
}

func TestSubagentGroupViewAndUpdateEdgeCases(t *testing.T) {
	g := NewSubagentGroup()
	g.width = 0
	if g.View() != "" {
		t.Fatal("expected empty view at non-positive width")
	}

	g.width = 6
	fitted := g.fitToWidth("123456789")
	if !strings.Contains(fitted, "…") {
		t.Fatalf("expected truncation for narrow group width, got %q", fitted)
	}
	g.width = 3
	if got := g.fitToWidth("abcdef"); style.StringWidth(stripANSI(got)) != 3 {
		t.Fatalf("expected width=3 truncation, got %q", got)
	}
	p1 := NewSubagentPanel(WithSubagentTitle("A"))
	p2 := NewSubagentPanel(WithSubagentTitle("B"))
	g.SetPanels([]*SubagentPanel{p1, p2})
	g.focusedPanel = 9
	if _, cmd := g.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}); cmd != nil {
		t.Fatal("expected nil command for unhandled key with out-of-range index fallback")
	}

	g.SetPanels(nil)
	if g.PanelCount() != 0 {
		t.Fatal("expected SetPanels(nil) to clear panels")
	}
	if _, cmd := g.Update(tea.KeyMsg{Type: tea.KeyTab}); cmd != nil {
		t.Fatal("expected nil command with no panels")
	}
}
