package tui_test

import (
	"strings"
	"testing"
	"time"

	"github.com/SCKelemen/tui/v2/agent"
	"github.com/SCKelemen/tui/v2/chart"
	"github.com/SCKelemen/tui/v2/chat"
	"github.com/SCKelemen/tui/v2/container"
	"github.com/SCKelemen/tui/v2/display"
	"github.com/SCKelemen/tui/v2/input"
	"github.com/SCKelemen/tui/v2/nav"
	"github.com/SCKelemen/tui/v2/selection"
	"github.com/SCKelemen/tui/v2/style"
	tui "github.com/SCKelemen/tui/v2"
	tea "github.com/charmbracelet/bubbletea"
)

func TestApplicationWithMultipleComponentTypes(t *testing.T) {
	app := tui.NewApplication()

	thread := agent.NewThreadProgress()
	thread.AddThread("t1", "integration thread")
	thread.AppendOutput("t1", "running")
	thread.SetThreadExpanded("t1", true)

	tabbar := nav.NewTabBar(nav.WithTabs(
		nav.Tab{ID: "one", Label: "One"},
		nav.Tab{ID: "two", Label: "Two"},
	))

	checklist := input.NewChecklist(
		input.WithChecklistTitle("Tasks"),
		input.WithChecklistItems(
			input.ChecklistItem{ID: "c1", Label: "first", Status: input.ItemPending},
		),
	)

	toast := display.NewToast()
	toast.PushWithDuration("hello", display.ToastInfo, time.Minute)

	components := []tui.Component{thread, tabbar, checklist, toast}
	for _, c := range components {
		app.AddComponent(c)
	}

	if !components[0].Focused() {
		t.Fatal("expected first component to be focused")
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("app.Init panicked: %v", r)
			}
		}()
		_ = app.Init()
	}()

	_, _ = app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	for i, c := range components {
		if view := c.View(); strings.TrimSpace(view) == "" {
			t.Fatalf("component %d should render non-empty view after resize", i)
		}
	}

	for i := 1; i < len(components); i++ {
		_, _ = app.Update(tea.KeyMsg{Type: tea.KeyTab})
		if !components[i].Focused() {
			t.Fatalf("component %d should be focused", i)
		}
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !components[0].Focused() {
		t.Fatal("expected first component focused after wrap")
	}

	if out := app.View(); strings.TrimSpace(out) == "" {
		t.Fatal("expected non-empty application view")
	}
}

func TestSelectionManagerWithCodeBlock(t *testing.T) {
	cb := display.NewCodeBlock(
		display.WithCodeBlockMouseSelection(true),
		display.WithExpanded(true),
		display.WithCode("alpha\nbeta\ngamma"),
	)

	selMgr := cb.GetSelectionManager()
	if selMgr == nil {
		t.Fatal("expected code block selection manager")
	}

	selMgr.SetRegion(selection.SelectableRegion{
		StartRow:     1,
		EndRow:       3,
		GutterWidth:  0,
		ContentLines: []string{"alpha", "beta", "gamma"},
	})
	selMgr.SetOffset(0, 0)

	selMgr.HandleMouse(tea.MouseMsg{X: 1, Y: 1, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	selMgr.HandleMouse(tea.MouseMsg{X: 3, Y: 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	selMgr.HandleMouse(tea.MouseMsg{X: 3, Y: 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

	if !selMgr.HasSelection() {
		t.Fatal("expected finalized selection")
	}

	if !selMgr.IsSelected(0, 1) {
		t.Fatal("expected start coordinate to be selected")
	}
	if !selMgr.IsSelected(1, 2) {
		t.Fatal("expected second-line coordinate to be selected")
	}

	selected := selMgr.GetSelectedText()
	if !strings.Contains(selected, "lpha") {
		t.Fatalf("expected selected text to include first-line slice, got %q", selected)
	}
	if !strings.Contains(selected, "bet") {
		t.Fatalf("expected selected text to include second-line slice, got %q", selected)
	}
}

func TestDesignTokensPropagation(t *testing.T) {
	tokens := style.DesignTokensForTheme("midnight")

	thread := agent.NewThreadProgress(agent.WithThreadProgressDesignTokens(tokens))
	thread.AddThread("theme", "themed thread")
	thread.UpdateThread("theme", agent.ThreadCompleted)

	tabbar := nav.NewTabBar(
		nav.WithTabBarDesignTokens(tokens),
		nav.WithTabs(nav.Tab{ID: "t1", Label: "Themed"}),
	)

	checklist := input.NewChecklist(
		input.WithChecklistDesignTokens(tokens),
		input.WithChecklistItems(input.ChecklistItem{ID: "i1", Label: "done", Status: input.ItemDone}),
	)

	toast := display.NewToast(display.WithToastDesignTokens(tokens))
	toast.Push("styled toast", display.ToastSuccess)

	components := []tui.Component{thread, tabbar, checklist, toast}

	for i, c := range components {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("component %d panicked: %v", i, r)
				}
			}()
			_ = c.Init()
			_, _ = c.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
			_ = c.View()
		}()
	}

	joined := thread.View() + tabbar.View() + checklist.View() + toast.View()
	if !strings.Contains(joined, "\x1b[") {
		t.Fatal("expected themed output to contain ANSI escape sequences")
	}
}

func TestSubagentGroupWithPanels(t *testing.T) {
	p1 := agent.NewSubagentPanel(agent.WithSubagentTitle("Planner"), agent.WithSubagentWidth(30))
	p2 := agent.NewSubagentPanel(agent.WithSubagentTitle("Coder"), agent.WithSubagentWidth(30))
	p3 := agent.NewSubagentPanel(agent.WithSubagentTitle("Reviewer"), agent.WithSubagentWidth(30))

	panels := []*agent.SubagentPanel{p1, p2, p3}
	for _, p := range panels {
		p.AddTool(agent.SubagentTool{Name: "analyze", Status: agent.ToolCompleted})
		p.AddTool(agent.SubagentTool{Name: "execute", Status: agent.ToolRunning})
	}

	group := agent.NewSubagentGroup(agent.WithSubagentGroupGap(2))
	group.AddPanel(p1)
	group.AddPanel(p2)
	group.AddPanel(p3)

	if group.PanelCount() != 3 {
		t.Fatalf("expected 3 panels, got %d", group.PanelCount())
	}

	_, _ = group.Update(tea.WindowSizeMsg{Width: 120, Height: 20})

	p1.SetStatus(agent.SubagentCompleted)
	p2.SetStatus(agent.SubagentCompleted)
	p3.SetStatus(agent.SubagentCompleted)

	if group.CompletedCount() != 3 {
		t.Fatalf("expected 3 completed panels, got %d", group.CompletedCount())
	}
	if !group.IsAllCompleted() {
		t.Fatal("expected all panels completed")
	}

	view := group.View()
	for _, title := range []string{"Planner", "Coder", "Reviewer"} {
		if !strings.Contains(view, title) {
			t.Fatalf("expected group view to include panel %q", title)
		}
	}
	if !strings.Contains(view, "Finished running") {
		t.Fatal("expected completion footer in group view")
	}
}

func TestScrollContainerWithMixedComponents(t *testing.T) {
	sc := nav.NewScrollContainer(nav.WithShowScrollbar(false), nav.WithViewportHeight(4))
	sc.Focus()

	code := display.NewCodeBlock(display.WithExpanded(true), display.WithCode("one\ntwo\nthree\nfour\nfive\nsix"))
	diff := display.NewDiffBlock(display.WithDiffExpanded(true), display.WithDiffLines([]display.DiffLine{
		{Type: display.DiffUnchanged, Content: "keep", LineNum: 1},
		{Type: display.DiffAdded, Content: "add", LineNum: 2},
		{Type: display.DiffRemoved, Content: "drop", LineNum: 3},
	}))
	conv := chat.NewConversationView(chat.WithShowTimestamps(false))
	_, _ = conv.Update(tea.WindowSizeMsg{Width: 50, Height: 6})
	conv.AddMessage(chat.Message{ID: "m1", Role: chat.RoleUser, Content: "hello from user"})
	conv.AddMessage(chat.Message{ID: "m2", Role: chat.RoleAssistant, Content: "reply from assistant"})
	sc.AddChild(code)
	sc.AddChild(diff)
	sc.AddChild(conv)

	if sc.ChildCount() != 3 {
		t.Fatalf("expected 3 children, got %d", sc.ChildCount())
	}

	_, _ = sc.Update(tea.WindowSizeMsg{Width: 50, Height: 4})
	first := sc.View()
	if strings.TrimSpace(first) == "" {
		t.Fatal("expected initial non-empty scroll view")
	}
	if strings.Contains(first, "assistant") {
		t.Fatal("expected clipped top view to not include bottom conversation yet")
	}

	_, _ = sc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if sc.GetScrollOffset() == 0 {
		t.Fatal("expected non-zero scroll offset at bottom")
	}
	bottom := sc.View()
	if !strings.Contains(bottom, "assistant") {
		t.Fatal("expected bottom view to include conversation content")
	}
}

func TestSplitPaneWithFocusDelegation(t *testing.T) {
	left := nav.NewTabBar(nav.WithTabs(nav.Tab{ID: "l", Label: "Left"}))
	right := input.NewChecklist(
		input.WithChecklistTitle("Right"),
		input.WithChecklistItems(input.ChecklistItem{ID: "r1", Label: "item", Status: input.ItemPending}),
	)

	split := container.NewSplitPane(
		container.WithLeftComponent(left),
		container.WithRightComponent(right),
	)
	split.Focus()

	_, _ = split.Update(tea.WindowSizeMsg{Width: 80, Height: 8})
	if split.GetFocusedPane() != 0 {
		t.Fatalf("expected initial focused pane 0, got %d", split.GetFocusedPane())
	}
	if !left.Focused() || right.Focused() {
		t.Fatal("expected left pane focused initially")
	}

	split.ToggleFocus()
	if split.GetFocusedPane() != 1 {
		t.Fatalf("expected focus on right pane, got %d", split.GetFocusedPane())
	}
	if !right.Focused() || left.Focused() {
		t.Fatal("expected right pane focused after toggle")
	}

	view := split.View()
	if !strings.Contains(view, "Left") {
		t.Fatal("expected split view to include left pane content")
	}
	if !strings.Contains(view, "Right") {
		t.Fatal("expected split view to include right pane content")
	}
}

func TestChartComponentsRender(t *testing.T) {
	bar := chart.NewBarChart("Bar", []string{"A", "B"}, []float64{3, 5})
	line := chart.NewLineChart("Line")
	line.AddSeries("s1", []float64{1, 3, 2, 4})
	heat := chart.NewHeatMap("Heat", [][]float64{{1, 2}, {3, 4}})
	scatter := chart.NewScatterPlot("Scatter", []float64{1, 2, 3}, []float64{2, 1, 3})

	components := []tui.Component{bar, line, heat, scatter}
	for i, c := range components {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("chart component %d panicked: %v", i, r)
				}
			}()

			_ = c.Init()
			_, _ = c.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
			c.Focus()
			if !c.Focused() {
				t.Fatalf("chart component %d should be focused", i)
			}
			c.Blur()
			if c.Focused() {
				t.Fatalf("chart component %d should be blurred", i)
			}

			if out := c.View(); strings.TrimSpace(out) == "" {
				t.Fatalf("chart component %d should render non-empty output", i)
			}
		}()
	}
}

func TestHighlightedCodeBlockRendering(t *testing.T) {
	cb := display.NewCodeBlock(
		display.WithCodeBlockLanguage(display.SyntaxLanguageGo),
		display.WithExpanded(true),
		display.WithCode("package main\n\nfunc main() {\n\tif true {\n\t\treturn\n\t}\n}"),
	)
	_, _ = cb.Update(tea.WindowSizeMsg{Width: 80, Height: 20})

	out := cb.View()
	if strings.TrimSpace(out) == "" {
		t.Fatal("expected highlighted code block to render")
	}
	if !strings.Contains(out, "\x1b[") {
		t.Fatal("expected syntax highlighted output with ANSI sequences")
	}
}
