package agent

import (
	"strings"
	"testing"
	"time"

	design "github.com/SCKelemen/design-system"
	"github.com/SCKelemen/tui/v2/spinner"
	"github.com/SCKelemen/tui/v2/style"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSubagentPanelCreation(t *testing.T) {
	panel := NewSubagentPanel()
	if panel == nil {
		t.Fatal("NewSubagentPanel returned nil")
	}

	if panel.status != SubagentRunning {
		t.Errorf("expected default status SubagentRunning, got %v", panel.status)
	}

	if panel.visibleTools != 8 {
		t.Errorf("expected default visibleTools=8, got %d", panel.visibleTools)
	}

	if panel.width <= 0 {
		t.Errorf("expected default width > 0, got %d", panel.width)
	}

	if panel.spinner.FrameCount() == 0 {
		t.Error("expected default spinner to have frames")
	}
}

func TestSubagentPanelOptions(t *testing.T) {
	panel := NewSubagentPanel(
		WithSubagentTitle("Create scrollbar component"),
		WithSubagentWidth(48),
		WithSubagentVisibleTools(3),
		WithSubagentModel("gpt-5.3-codex"),
	)

	if panel.title != "Create scrollbar component" {
		t.Errorf("WithSubagentTitle was not applied, got %q", panel.title)
	}

	if panel.width != 48 {
		t.Errorf("WithSubagentWidth was not applied, got %d", panel.width)
	}

	if panel.visibleTools != 3 {
		t.Errorf("WithSubagentVisibleTools was not applied, got %d", panel.visibleTools)
	}

	if panel.modelName != "gpt-5.3-codex" {
		t.Errorf("WithSubagentModel was not applied, got %q", panel.modelName)
	}
}

func TestSubagentPanelAddAndSetTools(t *testing.T) {
	panel := NewSubagentPanel(WithSubagentVisibleTools(2))
	panel.AddTool(SubagentTool{Name: "Searched for stripANSI", Status: ToolCompleted})
	panel.AddTool(SubagentTool{Name: "Read fileexplorer.go", Status: ToolCompleted})
	panel.AddTool(SubagentTool{Name: "Read selection.go", Status: ToolCompleted})

	if len(panel.tools) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(panel.tools))
	}

	if panel.hiddenCount != 1 {
		t.Errorf("expected hiddenCount=1, got %d", panel.hiddenCount)
	}

	panel.SetTools([]SubagentTool{{Name: "Only one tool", Status: ToolRunning}})
	if len(panel.tools) != 1 {
		t.Fatalf("expected 1 tool after SetTools, got %d", len(panel.tools))
	}

	if panel.hiddenCount != 0 {
		t.Errorf("expected hiddenCount=0 after reducing tools, got %d", panel.hiddenCount)
	}
}

func TestSubagentPanelStatusChanges(t *testing.T) {
	panel := NewSubagentPanel()
	if panel.GetStatus() != SubagentRunning {
		t.Fatalf("expected initial running status")
	}

	panel.SetStatus(SubagentCompleted)
	if panel.GetStatus() != SubagentCompleted {
		t.Errorf("expected completed status, got %v", panel.GetStatus())
	}

	panel.SetStatus(SubagentFailed)
	if panel.GetStatus() != SubagentFailed {
		t.Errorf("expected failed status, got %v", panel.GetStatus())
	}

	panel.SetStatus(SubagentAborted)
	if panel.GetStatus() != SubagentAborted {
		t.Errorf("expected aborted status, got %v", panel.GetStatus())
	}
}

func TestSubagentPanelTreeRenderingIconBeforeText(t *testing.T) {
	panel := NewSubagentPanel(
		WithSubagentTitle("Create scrollcontainer"),
		WithSubagentWidth(80),
	)
	panel.SetStatus(SubagentCompleted)
	panel.SetTools([]SubagentTool{
		{Name: `Searched for "stripANSI"`, Status: ToolCompleted},
		{Name: "Read fileexplorer.go", Status: ToolCompleted},
		{Name: "Read selection.go", Status: ToolCompleted},
	})

	view := panel.View()
	stripped := stripANSI(view)

	if !strings.Contains(stripped, "├ ✓ Searched for \"stripANSI\"") {
		t.Fatalf("expected tree line with icon before text, got:\n%s", stripped)
	}

	if !strings.Contains(stripped, "└ ✓ Read selection.go") {
		t.Fatalf("expected last tree connector line, got:\n%s", stripped)
	}
}

func TestSubagentPanelHiddenToolsCountRendering(t *testing.T) {
	panel := NewSubagentPanel(
		WithSubagentTitle("Task"),
		WithSubagentWidth(80),
		WithSubagentVisibleTools(2),
	)
	panel.SetTools([]SubagentTool{
		{Name: "t1", Status: ToolCompleted},
		{Name: "t2", Status: ToolCompleted},
		{Name: "t3", Status: ToolCompleted},
		{Name: "t4", Status: ToolCompleted},
	})

	view := stripANSI(panel.View())
	if !strings.Contains(view, "... +2 earlier tools") {
		t.Fatalf("expected hidden tools line, got:\n%s", view)
	}
}

func TestSubagentPanelBorders(t *testing.T) {
	panel := NewSubagentPanel(
		WithSubagentTitle("Task"),
		WithSubagentWidth(20),
	)
	lines := strings.Split(strings.TrimSuffix(panel.View(), "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least two lines in view")
	}

	view := stripANSI(panel.View())
	if !strings.Contains(view, "▄") || !strings.Contains(view, "▀") {
		t.Fatalf("expected top/bottom border characters, got:\n%s", view)
	}
}
func TestSubagentPanelFooterFormatting(t *testing.T) {
	panel := NewSubagentPanel(
		WithSubagentTitle("Task"),
		WithSubagentWidth(120),
		WithSubagentModel("gpt-5.3-codex"),
	)
	panel.SetStatus(SubagentCompleted)
	panel.SetElapsed(150 * time.Second)
	panel.SetTokenCount("235k")
	panel.SetCost("$0.41")

	view := stripANSI(panel.View())
	if !strings.Contains(view, "Done in 2m 30s") {
		t.Fatalf("expected 'Done in 2m 30s' in footer, got:\n%s", view)
	}
	if !strings.Contains(view, "gpt-5.3-codex · 235k · $0.41") {
		t.Fatalf("expected metadata on right side of footer, got:\n%s", view)
	}

	panel.SetStatus(SubagentRunning)
	view = stripANSI(panel.View())
	if !strings.Contains(view, "Working…") {
		t.Fatalf("expected 'Working…' in footer, got:\n%s", view)
	}
	if !strings.Contains(view, "gpt-5.3-codex · 235k · $0.41") {
		t.Fatalf("expected metadata on right side of running footer, got:\n%s", view)
	}
}
func TestSubagentPanelTruncatedTitle(t *testing.T) {
	panel := NewSubagentPanel(
		WithSubagentWidth(26),
		WithSubagentTitle("Create scrollcontainer and selection handling"),
	)

	view := stripANSI(panel.View())
	if !strings.Contains(view, "Subagent: ") {
		t.Fatalf("expected Subagent header, got:\n%s", view)
	}
	if !strings.Contains(view, "…") {
		t.Fatalf("expected truncated title with ellipsis, got:\n%s", view)
	}
}

func TestSubagentPanelUpdateWindowAndTick(t *testing.T) {
	panel := NewSubagentPanel(WithSubagentWidth(20))
	panel.SetStatus(SubagentRunning)

	panel.Update(tea.WindowSizeMsg{Width: 44, Height: 20})
	if panel.width != 44 {
		t.Fatalf("expected width to update to 44, got %d", panel.width)
	}

	beforeIdx := panel.spinnerIdx
	component, cmd := panel.Update(subagentPanelTickMsg{now: time.Now()})
	if cmd == nil {
		t.Fatal("expected tick command while running")
	}
	updated := component.(*SubagentPanel)
	if updated.spinnerIdx == beforeIdx {
		t.Fatal("expected spinner index to advance on tick")
	}
}

func TestSubagentPanelFocusManagement(t *testing.T) {
	panel := NewSubagentPanel()
	if panel.Focused() {
		t.Fatal("panel should not be focused initially")
	}
	panel.Focus()
	if !panel.Focused() {
		t.Fatal("panel should be focused after Focus")
	}
	panel.Blur()
	if panel.Focused() {
		t.Fatal("panel should not be focused after Blur")
	}
}

func TestSubagentPanelOptionsInitAndTickBehavior(t *testing.T) {
	tokens := &design.DesignTokens{Accent: "#00ff00", Color: "#ff0000"}
	panel := NewSubagentPanel(
		WithSubagentVisibleTools(-3),
		WithSubagentWidth(0),
		WithSubagentDesignTokens(tokens),
		WithSubagentTheme("midnight"),
	)

	if panel.visibleTools != 0 {
		t.Fatalf("expected visibleTools clamp to 0, got %d", panel.visibleTools)
	}
	if panel.width != 36 {
		t.Fatalf("expected invalid width to keep default 36, got %d", panel.width)
	}
	if panel.runningColor == "" || panel.footerColor == "" {
		t.Fatal("expected theme/design tokens to apply colors")
	}

	if cmd := panel.Init(); cmd == nil {
		t.Fatal("expected Init command while panel is running")
	}

	panel.SetStatus(SubagentCompleted)
	panel.SetTools([]SubagentTool{{Name: "tool", Status: ToolRunning}})
	if cmd := panel.Init(); cmd == nil {
		t.Fatal("expected Init command when a tool is still running")
	}

	panel.SetTools(nil)
	if len(panel.tools) != 0 {
		t.Fatal("expected SetTools(nil) to clear tools")
	}

	panel.applyDesignTokens(nil)
}

func TestSubagentPanelRenderingHelpersAndEdgeCases(t *testing.T) {
	panel := NewSubagentPanel(WithSubagentWidth(12), WithSubagentTitle("very long subagent title"))

	panel.SetStatus(SubagentFailed)
	if got := stripANSI(panel.renderToolIcon(ToolCompleted)); !strings.Contains(got, "✓") {
		t.Fatalf("expected completed tool icon, got %q", got)
	}
	if got := stripANSI(panel.renderToolIcon(ToolFailed)); !strings.Contains(got, "✗") {
		t.Fatalf("expected failed tool icon, got %q", got)
	}
	if got := stripANSI(panel.renderToolIcon(ToolStatus(99))); got != "·" {
		t.Fatalf("expected default tool icon, got %q", got)
	}

	panel.spinner = spinner.Spinner{Frames: nil}
	panel.SetStatus(SubagentRunning)
	icon, plain := panel.renderStatusIcon()
	if plain != "◐" || !strings.Contains(stripANSI(icon), "◐") {
		t.Fatalf("expected running fallback icon, got plain=%q icon=%q", plain, icon)
	}

	panel.SetElapsed(-1 * time.Second)
	if panel.elapsed != 0 {
		t.Fatalf("expected negative elapsed to clamp to 0, got %v", panel.elapsed)
	}
	if got := formatSubagentDuration(-1 * time.Second); got != "0s" {
		t.Fatalf("expected clamped duration format 0s, got %q", got)
	}

	panel.width = 3
	if got := panel.fitToWidth("abcdef"); style.StringWidth(stripANSI(got)) != 3 {
		t.Fatalf("expected width-3 truncation, got %q", got)
	}
	panel.width = 0
	if panel.View() != "" {
		t.Fatal("expected empty view at non-positive width")
	}

	panel.width = 18
	header := stripANSI(panel.renderHeaderLine())
	if !strings.Contains(header, "Subagent:") {
		t.Fatalf("expected header content, got %q", header)
	}
}
